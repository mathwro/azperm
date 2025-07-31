package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mathwro/azperm/internal/models"
)

// AzureCloudConfig represents Azure cloud configuration
type AzureCloudConfig struct {
	Name                    string `json:"name"`
	ManagementEndpointURL   string `json:"endpoints.management"`
	ResourceManagerEndpoint string `json:"endpoints.resourceManager"`
	ActiveDirectoryEndpoint string `json:"endpoints.activeDirectory"`
}

// Client represents an Azure API client
type Client struct {
	httpClient *http.Client
	apiVersion string
}

// NewClient creates a new Azure API client
func NewClient() *Client {
	// Default API version - using latest stable version for provider operations
	defaultAPIVersion := "2022-04-01"
	
	// Allow API version to be overridden via environment variable
	if envAPIVersion := os.Getenv("AZPERM_API_VERSION"); envAPIVersion != "" {
		defaultAPIVersion = envAPIVersion
	}
	
	return &Client{
		httpClient: &http.Client{},
		apiVersion: defaultAPIVersion,
	}
}

// SetAPIVersion sets a custom API version for provider operations requests
func (c *Client) SetAPIVersion(version string) {
	if version != "" {
		c.apiVersion = version
	}
}

// GetAPIVersion returns the current API version being used
func (c *Client) GetAPIVersion() string {
	return c.apiVersion
}

// GetCloudConfig returns the current Azure cloud configuration
func (c *Client) GetCloudConfig() (*AzureCloudConfig, error) {
	return c.getAzureCloudConfig()
}

// GetEffectiveEndpoint returns the actual management endpoint being used (including environment overrides)
func (c *Client) GetEffectiveEndpoint() (string, error) {
	// Check for environment variable override first
	if envEndpoint := os.Getenv("AZPERM_MANAGEMENT_ENDPOINT"); envEndpoint != "" {
		return strings.TrimSuffix(envEndpoint, "/"), nil
	}
	
	// Otherwise get from Azure CLI configuration
	cloudConfig, err := c.getAzureCloudConfig()
	if err != nil {
		// Fallback to public cloud
		return "https://management.azure.com", nil
	}
	
	return cloudConfig.ManagementEndpointURL, nil
}

// GetEffectiveCloudInfo returns comprehensive information about the effective configuration
func (c *Client) GetEffectiveCloudInfo() (string, string, string, error) {
	effectiveEndpoint, err := c.GetEffectiveEndpoint()
	if err != nil {
		return "", "", "", err
	}
	
	cloudName := "Unknown"
	isOverridden := false
	
	// Check if endpoint is overridden
	if os.Getenv("AZPERM_MANAGEMENT_ENDPOINT") != "" {
		cloudName = "Custom (Environment Override)"
		isOverridden = true
	} else {
		// Get cloud name from Azure CLI
		if cloudConfig, err := c.getAzureCloudConfig(); err == nil {
			cloudName = cloudConfig.Name
		}
	}
	
	source := "Azure CLI"
	if isOverridden {
		source = "Environment Variable"
	}
	
	return cloudName, effectiveEndpoint, source, nil
}

// FetchProviderOperations retrieves provider operations data from Azure API
func (c *Client) FetchProviderOperations(useLive bool) (map[string]models.ProviderOperationsResponse, error) {
	// Always use live API as requested by user
	return c.FetchRealProviderOperations("")
}

// getAzureCloudConfig gets the current Azure cloud configuration from Azure CLI
func (c *Client) getAzureCloudConfig() (*AzureCloudConfig, error) {
	// Get current cloud configuration from Azure CLI
	cmd := exec.Command("az", "cloud", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure cloud configuration from Azure CLI: %w", err)
	}

	var cloudConfig struct {
		Name      string `json:"name"`
		Endpoints struct {
			Management        string `json:"management"`
			ResourceManager   string `json:"resourceManager"`
			ActiveDirectory   string `json:"activeDirectory"`
		} `json:"endpoints"`
	}

	if err := json.Unmarshal(output, &cloudConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Azure cloud configuration: %w", err)
	}

	// Use Resource Manager endpoint for ARM APIs (Provider Operations API)
	// The management endpoint is for classic/legacy operations
	managementURL := cloudConfig.Endpoints.ResourceManager
	if managementURL == "" {
		// Fallback to management endpoint if Resource Manager is not available
		managementURL = cloudConfig.Endpoints.Management
	}

	// Remove trailing slash if present
	managementURL = strings.TrimSuffix(managementURL, "/")

	return &AzureCloudConfig{
		Name:                    cloudConfig.Name,
		ManagementEndpointURL:   managementURL,
		ResourceManagerEndpoint: cloudConfig.Endpoints.ResourceManager,
		ActiveDirectoryEndpoint: cloudConfig.Endpoints.ActiveDirectory,
	}, nil
}

// buildProviderOperationsURL constructs the provider operations URL for the current cloud
func (c *Client) buildProviderOperationsURL() (string, error) {
	// Allow management endpoint to be overridden via environment variable
	if envEndpoint := os.Getenv("AZPERM_MANAGEMENT_ENDPOINT"); envEndpoint != "" {
		envEndpoint = strings.TrimSuffix(envEndpoint, "/")
		return fmt.Sprintf("%s/providers/Microsoft.Authorization/providerOperations?api-version=%s&$expand=resourceTypes", 
			envEndpoint, c.apiVersion), nil
	}
	
	cloudConfig, err := c.getAzureCloudConfig()
	if err != nil {
		// Fallback to public cloud if we can't detect the environment
		return fmt.Sprintf("https://management.azure.com/providers/Microsoft.Authorization/providerOperations?api-version=%s&$expand=resourceTypes", c.apiVersion), nil
	}

	return fmt.Sprintf("%s/providers/Microsoft.Authorization/providerOperations?api-version=%s&$expand=resourceTypes", 
		cloudConfig.ManagementEndpointURL, c.apiVersion), nil
}

// FetchRealProviderOperations fetches real data from Azure Management API
func (c *Client) FetchRealProviderOperations(accessToken string) (map[string]models.ProviderOperationsResponse, error) {
	// Build URL dynamically based on current Azure cloud configuration
	url, err := c.buildProviderOperationsURL()
	if err != nil {
		return nil, fmt.Errorf("failed to build provider operations URL: %w", err)
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse struct {
		Value []models.ProviderOperationsResponse `json:"value"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to map for easier lookup
	result := make(map[string]models.ProviderOperationsResponse)
	for _, provider := range apiResponse.Value {
		// Extract namespace from the full ID (e.g., "Microsoft.Resources" from "Microsoft.Authorization/providerOperations/Microsoft.Resources")
		namespace := provider.Namespace
		if strings.Contains(namespace, "/") {
			parts := strings.Split(namespace, "/")
			namespace = parts[len(parts)-1]
		}
		result[namespace] = provider
	}

	return result, nil
}
