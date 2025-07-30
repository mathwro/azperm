package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mathwro/AzCliPermissions/internal/models"
)

// Client represents an Azure API client
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Azure API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// FetchProviderOperations retrieves provider operations data from Azure API
func (c *Client) FetchProviderOperations(useLive bool) (map[string]models.ProviderOperationsResponse, error) {
	// Always use live API as requested by user
	return c.FetchRealProviderOperations("")
}

// FetchRealProviderOperations fetches real data from Azure Management API
func (c *Client) FetchRealProviderOperations(accessToken string) (map[string]models.ProviderOperationsResponse, error) {
	url := "https://management.azure.com/providers/Microsoft.Authorization/providerOperations?api-version=2018-01-01-preview&$expand=resourceTypes"
	
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
