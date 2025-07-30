package models

// PermissionMapping represents the structure for command-to-permission mappings
type PermissionMapping struct {
	Commands    map[string][]string `json:"commands"`
	LastUpdated string              `json:"last_updated,omitempty"`
	Source      string              `json:"source,omitempty"`
}

// AzureCommand represents a parsed Azure CLI command
type AzureCommand struct {
	Service    string            `json:"service"`
	Operation  string            `json:"operation"`
	Parameters map[string]string `json:"parameters"`
	FullCmd    string            `json:"full_command"`
}

// ProviderOperation represents an Azure Resource Provider operation
type ProviderOperation struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	Origin       string `json:"origin"`
	IsDataAction bool   `json:"isDataAction"`
}

// ResourceType represents an Azure resource type with operations
type ResourceType struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"displayName"`
	ResourceType string   `json:"resourceType"`
	ApiVersions  []string `json:"apiVersions"`
	Operations   []string `json:"operations"`
}

// ProviderResourceType represents a resource type from the Provider Operations API
type ProviderResourceType struct {
	Name         string              `json:"name"`
	DisplayName  string              `json:"displayName"`
	Operations   []ProviderOperation `json:"operations"`
}

// ProviderOperationsResponse represents the complete Azure provider operations response
type ProviderOperationsResponse struct {
	Namespace     string                  `json:"id"`
	DisplayName   string                  `json:"displayName"`
	Operations    []ProviderOperation     `json:"operations"`
	ResourceTypes []ProviderResourceType  `json:"resourceTypes"`
}

// ResourceProvider represents an Azure Resource Provider
type ResourceProvider struct {
	Namespace  string              `json:"namespace"`
	Operations []ProviderOperation `json:"operations"`
}

// CommandToAPIMapping represents mapping from CLI command to REST API
type CommandToAPIMapping struct {
	Command     string   `json:"command"`
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Operation   string   `json:"operation"`
	Permissions []string `json:"permissions"`
}

// ConfidenceLevel represents the confidence level of permission detection
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)
