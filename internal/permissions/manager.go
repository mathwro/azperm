package permissions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mathwro/AzCliPermissions/internal/models"
)

// Manager handles permission mappings and caching
type Manager struct {
	mappings models.PermissionMapping
}

// NewManager creates a new permission manager
func NewManager() *Manager {
	return &Manager{
		mappings: models.PermissionMapping{
			Commands: make(map[string][]string),
		},
	}
}

// LoadPermissions loads permissions from file or uses defaults
func (m *Manager) LoadPermissions() {
	// Try to load from permissions.json file
	if data, err := os.ReadFile("permissions.json"); err == nil {
		if err := json.Unmarshal(data, &m.mappings); err != nil {
			fmt.Printf("Warning: Failed to parse permissions.json: %v\n", err)
			m.loadDefaultPermissions()
		}
	} else {
		// File doesn't exist, use defaults
		m.loadDefaultPermissions()
	}
}

// loadDefaultPermissions sets up default permission mappings
func (m *Manager) loadDefaultPermissions() {
	m.mappings = models.PermissionMapping{
		Commands: map[string][]string{
			"group create": {
				"Microsoft.Resources/subscriptions/resourceGroups/write",
			},
			"group delete": {
				"Microsoft.Resources/subscriptions/resourceGroups/delete",
			},
			"vm start": {
				"Microsoft.Compute/virtualMachines/start/action",
			},
			"vm stop": {
				"Microsoft.Compute/virtualMachines/powerOff/action",
			},
			"storage account create": {
				"Microsoft.Storage/storageAccounts/write",
			},
		},
		LastUpdated: "built-in",
		Source:      "default-minimal",
	}
}

// GetPermissions retrieves permissions for a command with fallback logic
func (m *Manager) GetPermissions(cmd *models.AzureCommand) ([]string, models.ConfidenceLevel) {
	// Check for exact match in our database
	if permissions, exists := m.mappings.Commands[cmd.FullCmd]; exists {
		return permissions, models.ConfidenceMedium
	}

	// Try to find partial matches
	permissions := m.findPartialMatches(cmd)
	if len(permissions) > 0 {
		return permissions, models.ConfidenceLow
	}

	return nil, models.ConfidenceLow
}

// findPartialMatches tries to find similar commands
func (m *Manager) findPartialMatches(cmd *models.AzureCommand) []string {
	var permissions []string

	// Try to find similar commands
	for key, perms := range m.mappings.Commands {
		// Check if the command starts with the same service
		if strings.HasPrefix(key, cmd.Service) && strings.Contains(key, cmd.Operation) {
			permissions = append(permissions, perms...)
			break
		}
	}

	// If still no match, provide intelligent suggestions based on operation
	if len(permissions) == 0 {
		permissions = m.suggestPermissionsByOperation(cmd)
	}

	return permissions
}

// suggestPermissionsByOperation provides intelligent permission suggestions
func (m *Manager) suggestPermissionsByOperation(cmd *models.AzureCommand) []string {
	// Map common operations to likely permission patterns
	operationMap := map[string]string{
		"create":  "write",
		"update":  "write",
		"set":     "write",
		"delete":  "delete",
		"remove":  "delete",
		"list":    "read",
		"show":    "read",
		"get":     "read",
		"start":   "start/action",
		"stop":    "powerOff/action",
		"restart": "restart/action",
		"scale":   "scale/action",
	}

	if action, exists := operationMap[cmd.Operation]; exists {
		// Try to construct a reasonable permission based on service
		resourceProvider := getResourceProviderForService(cmd.Service)
		if resourceProvider != "" {
			resourceType := getResourceTypeForService(cmd.Service)
			permission := fmt.Sprintf("%s/%s/%s", resourceProvider, resourceType, action)
			return []string{permission}
		}
	}

	return []string{}
}

// SavePermissions saves the current permission mappings to file
func (m *Manager) SavePermissions() error {
	data, err := json.MarshalIndent(m.mappings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("permissions.json", data, 0644)
}

// CachePermission caches a permission mapping for future use
func (m *Manager) CachePermission(command string, permissions []string) {
	if m.mappings.Commands == nil {
		m.mappings.Commands = make(map[string][]string)
	}
	
	m.mappings.Commands[command] = permissions
	
	// Optionally save to file for persistence
	m.SavePermissions()
}

// UpdateMappings updates the internal mappings with new data
func (m *Manager) UpdateMappings(newMappings models.PermissionMapping) {
	m.mappings = newMappings
}

// GetMappings returns the current permission mappings
func (m *Manager) GetMappings() models.PermissionMapping {
	return m.mappings
}

// SetMappingsWithAPIIntegration updates mappings from API discovery
func (m *Manager) SetMappingsWithAPIIntegration(commands []string, operations map[string][]string) models.PermissionMapping {
	mappings := make(map[string][]string)
	
	for _, cmdStr := range commands {
		permissions := inferPermissions(cmdStr, operations)
		if len(permissions) > 0 {
			mappings[cmdStr] = permissions
		}
	}

	newMappings := models.PermissionMapping{
		Commands:    mappings,
		LastUpdated: time.Now().Format("2006-01-02 15:04:05"),
		Source:      "azure-rest-api-integration",
	}

	m.mappings = newMappings
	return newMappings
}

// getResourceProviderForService maps service names to Azure resource providers
func getResourceProviderForService(service string) string {
	serviceMap := map[string]string{
		"group":             "Microsoft.Resources",
		"vm":                "Microsoft.Compute",
		"storage":           "Microsoft.Storage",
		"webapp":            "Microsoft.Web",
		"functionapp":       "Microsoft.Web",
		"keyvault":          "Microsoft.KeyVault",
		"network":           "Microsoft.Network",
		"sql":               "Microsoft.Sql",
		"aks":               "Microsoft.ContainerService",
		"cosmosdb":          "Microsoft.DocumentDB",
		"role":              "Microsoft.Authorization",
		"ad":                "Microsoft.Graph",
		"monitor":           "Microsoft.Insights",
		"backup":            "Microsoft.RecoveryServices",
		"cdn":               "Microsoft.Cdn",
		"redis":             "Microsoft.Cache",
		"servicebus":        "Microsoft.ServiceBus",
		"eventhub":          "Microsoft.EventHub",
		"iot":               "Microsoft.Devices",
		"batch":             "Microsoft.Batch",
		"hdinsight":         "Microsoft.HDInsight",
		"search":            "Microsoft.Search",
		"cognitiveservices": "Microsoft.CognitiveServices",
	}

	// Handle compound services
	for key, provider := range serviceMap {
		if strings.Contains(service, key) {
			return provider
		}
	}

	return ""
}

// getResourceTypeForService maps service names to Azure resource types
func getResourceTypeForService(service string) string {
	serviceMap := map[string]string{
		"group":              "subscriptions/resourceGroups",
		"vm":                 "virtualMachines",
		"storage account":    "storageAccounts",
		"storage blob":       "storageAccounts/blobServices",
		"webapp":             "sites",
		"functionapp":        "sites",
		"keyvault":           "vaults",
		"network vnet":       "virtualNetworks",
		"network nsg":        "networkSecurityGroups",
		"sql server":         "servers",
		"sql db":             "servers/databases",
		"aks":                "managedClusters",
		"cosmosdb":           "databaseAccounts",
		"role assignment":    "roleAssignments",
		"role definition":    "roleDefinitions",
		"ad user":            "users",
	}

	if resourceType, exists := serviceMap[service]; exists {
		return resourceType
	}

	// Default fallback - try to construct from service name
	parts := strings.Fields(service)
	if len(parts) > 1 {
		return strings.Join(parts, "")
	}

	return service
}

// inferPermissions infers permissions from command and operations
func inferPermissions(command string, operations map[string][]string) []string {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return nil
	}

	service := parts[0]
	operation := parts[len(parts)-1]

	// Handle compound services
	if len(parts) > 2 {
		service = strings.Join(parts[:len(parts)-1], " ")
	}

	// Get resource provider for service
	provider := getResourceProviderForService(service)
	if provider == "" {
		return nil
	}

	// Get resource type
	resourceType := getResourceTypeForService(service)

	// Map operation to action
	actionMap := map[string][]string{
		"create":  {"write"},
		"update":  {"write"},
		"set":     {"write"},
		"delete":  {"delete"},
		"remove":  {"delete"},
		"list":    {"read"},
		"show":    {"read"},
		"get":     {"read"},
		"start":   {"start/action"},
		"stop":    {"powerOff/action", "stop/action"},
		"restart": {"restart/action"},
		"scale":   {"scale/action"},
	}

	var permissions []string
	if actions, exists := actionMap[operation]; exists {
		for _, action := range actions {
			permission := fmt.Sprintf("%s/%s/%s", provider, resourceType, action)
			permissions = append(permissions, permission)
		}
	} else {
		// Default to read permission for unknown operations
		permission := fmt.Sprintf("%s/%s/read", provider, resourceType)
		permissions = append(permissions, permission)
	}

	return permissions
}
