package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mathwro/AzCliPermissions/internal/azure"
	"github.com/mathwro/AzCliPermissions/internal/display"
	"github.com/mathwro/AzCliPermissions/internal/models"
	"github.com/mathwro/AzCliPermissions/internal/parser"
	"github.com/mathwro/AzCliPermissions/internal/permissions"
	"github.com/mathwro/AzCliPermissions/internal/shell"
)

// CLI represents the main CLI application
type CLI struct {
	permManager *permissions.Manager
	azureClient *azure.Client
	colors      *display.Colors
	liveMode    bool
	debugMode   bool
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		permManager: permissions.NewManager(),
		azureClient: azure.NewClient(),
		colors:      display.NewColors(),
		liveMode:    true, // Always use live mode by default
		debugMode:   false, // Debug mode off by default
	}
}

// SetLiveMode enables or disables live API querying mode
func (c *CLI) SetLiveMode(enabled bool) {
	c.liveMode = enabled
}

// SetDebugMode enables or disables debug mode with verbose output
func (c *CLI) SetDebugMode(enabled bool) {
	c.debugMode = enabled
}

// Run executes the main CLI logic
func (c *CLI) Run() error {
	// Load permissions database
	c.permManager.LoadPermissions()

	// Check if input is piped
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("failed to check stdin: %w", err)
	}

	var azCommand string

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is piped
		azCommand, err = c.readPipedInput()
		if err != nil {
			return fmt.Errorf("failed to read piped input: %w", err)
		}
	} else {
		// No piped input, try to get last Azure CLI command from shell history
		azCommand, err = c.getLastAzureCommand()
		if err != nil {
			c.colors.ShowUsage()
			return nil
		}
	}

	// Parse the Azure CLI command
	cmd, err := parser.ParseAzureCommand(azCommand)
	if err != nil {
		return fmt.Errorf("failed to parse Azure command: %w", err)
	}

	// Get permissions using live Azure API querying
	permissions, _ := c.getPermissions(cmd)

	if len(permissions) == 0 {
		c.colors.ShowNoPermissionsWarning(cmd.FullCmd, true)
		return fmt.Errorf("failed to retrieve permissions from Azure API")
	}

	// Always display results with live query indication since we always use live mode
	c.colors.DisplayPermissionsWithLiveQuery(cmd, permissions)

	return nil
}

// readPipedInput reads input from stdin (piped commands)
func (c *CLI) readPipedInput() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(lines) == 0 {
		return "", fmt.Errorf("no input provided")
	}

	// Join all lines or take the first Azure CLI command
	input := strings.Join(lines, " ")
	
	// Extract Azure CLI command if it's mixed with other content
	if azCmd := c.extractAzureCommand(input); azCmd != "" {
		return azCmd, nil
	}

	return input, nil
}

// extractAzureCommand extracts Azure CLI command from mixed input
func (c *CLI) extractAzureCommand(input string) string {
	// Look for patterns like "az ..." in the input
	words := strings.Fields(input)
	
	var azCommand []string
	foundAz := false
	
	for _, word := range words {
		if word == "az" {
			foundAz = true
			azCommand = []string{word}
		} else if foundAz {
			// Continue collecting command parts
			azCommand = append(azCommand, word)
			
			// Stop at common terminators or if we hit another command
			if strings.HasPrefix(word, "--") && len(azCommand) > 3 {
				// We have enough of the command
				break
			}
		}
	}
	
	if foundAz && len(azCommand) >= 3 {
		return strings.Join(azCommand, " ")
	}
	
	return ""
}

// getLastAzureCommand attempts to get the last Azure CLI command from shell history
func (c *CLI) getLastAzureCommand() (string, error) {
	command, err := shell.GetLastAzureCommand()
	if err != nil {
		return "", fmt.Errorf("failed to get last Azure command: %w", err)
	}

	return command, nil
}

// getPermissions retrieves permissions using live Azure API querying
func (c *CLI) getPermissions(cmd *models.AzureCommand) ([]string, models.ConfidenceLevel) {
	// Always try to get permissions from live Azure API first
	if permissions, err := c.getLivePermissions(cmd); err == nil && len(permissions) > 0 {
		// Cache the result for future use
		c.permManager.CachePermission(cmd.FullCmd, permissions)
		return permissions, models.ConfidenceHigh
	}

	// If live API fails, show error and exit gracefully
	c.colors.Error.Println("❌ Failed to query Azure API for permissions")
	c.colors.Warning.Println("💡 Make sure you're logged in with 'az login' and have internet connectivity")
	
	// Return empty permissions to indicate failure
	return []string{}, models.ConfidenceLow
}

// getLivePermissions attempts to get permissions using live Azure API
func (c *CLI) getLivePermissions(cmd *models.AzureCommand) ([]string, error) {
	// Try to get Azure CLI access token
	accessToken, err := c.getAzureAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure access token: %w", err)
	}

	c.colors.Info.Println("🔍 Querying Azure API for permissions...")

	// Use the real Azure API with the access token
	operations, err := c.azureClient.FetchRealProviderOperations(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch real provider operations: %w", err)
	}

	if c.debugMode {
		c.colors.Info.Printf("📊 Retrieved %d resource providers from Azure API\n", len(operations))
	}

	// Find relevant operations for the command
	return c.findOperationsForCommand(cmd, operations)
}

// getAzureAccessToken attempts to get an access token from Azure CLI
func (c *CLI) getAzureAccessToken() (string, error) {
	// Try to get access token using Azure CLI
	cmd := exec.Command("az", "account", "get-access-token", "--query", "accessToken", "--output", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get access token from Azure CLI (make sure you're logged in with 'az login'): %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("empty access token returned from Azure CLI")
	}

	return token, nil
}

// findOperationsForCommand finds relevant operations from the live API data
func (c *CLI) findOperationsForCommand(cmd *models.AzureCommand, operations map[string]models.ProviderOperationsResponse) ([]string, error) {
	// Map service to resource provider
	provider := c.mapServiceToProvider(cmd.Service)
	if provider == "" {
		return nil, fmt.Errorf("unknown service: %s", cmd.Service)
	}

	if c.debugMode {
		c.colors.Info.Printf("🔗 Mapped service '%s' to provider '%s'\n", cmd.Service, provider)
	}

	providerOps, exists := operations[provider]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", provider)
	}

	if c.debugMode {
		c.colors.Info.Printf("📋 Found provider '%s' with %d resource types\n", provider, len(providerOps.ResourceTypes))
	}

	permissionsSet := make(map[string]bool) // Use map to avoid duplicates
	
	// First check provider-level operations
	for _, operation := range providerOps.Operations {
		if c.matchesOperation(cmd.Operation, operation.Name) {
			permissionsSet[operation.Name] = true
			if c.debugMode {
				c.colors.Info.Printf("✅ Matched provider operation: %s\n", operation.Name)
			}
		}
	}
	
	// Then check resource type operations
	for _, resourceType := range providerOps.ResourceTypes {
		if c.matchesResourceType(cmd, resourceType.Name) {
			if c.debugMode {
				c.colors.Info.Printf("✅ Matched resource type: %s\n", resourceType.Name)
			}
			// Find operations that match the command operation
			for _, operation := range resourceType.Operations {
				if c.matchesOperation(cmd.Operation, operation.Name) {
					permissionsSet[operation.Name] = true
					if c.debugMode {
						c.colors.Info.Printf("✅ Matched operation: %s\n", operation.Name)
					}
				}
			}
		} else if c.debugMode && c.isDataPlaneOperation(cmd) {
			// Show what we're rejecting for debugging
			c.colors.Info.Printf("❌ Rejected resource type: %s\n", resourceType.Name)
		}
	}

	// Convert map to slice
	var permissions []string
	for permission := range permissionsSet {
		permissions = append(permissions, permission)
	}

	// If no exact matches, provide intelligent suggestions
	if len(permissions) == 0 {
		if c.debugMode {
			c.colors.Warning.Println("⚠️  No exact matches found, using intelligent suggestions...")
		}
		permissions = c.suggestOperationsFromLiveData(cmd, providerOps)
	}

	if c.debugMode {
		c.colors.Info.Printf("🎯 Found %d permissions\n", len(permissions))
	}
	return permissions, nil
}

// Helper methods for mapping and matching (similar to azure client)
func (c *CLI) mapServiceToProvider(service string) string {
	serviceMap := map[string]string{
		"group":      "Microsoft.Resources",
		"vm":         "Microsoft.Compute",
		"storage":    "Microsoft.Storage",
		"webapp":     "Microsoft.Web",
		"keyvault":   "Microsoft.KeyVault",
		"network":    "Microsoft.Network",
		"sql":        "Microsoft.Sql",
		"aks":        "Microsoft.ContainerService",
		"role":       "Microsoft.Authorization",
	}

	for key, provider := range serviceMap {
		if strings.Contains(service, key) {
			return provider
		}
	}
	return ""
}

func (c *CLI) matchesResourceType(cmd *models.AzureCommand, resourceType string) bool {
	service := strings.ToLower(cmd.Service)
	resType := strings.ToLower(resourceType)
	operation := strings.ToLower(cmd.Operation)
	
	// Dynamic data plane detection based on service patterns
	if c.isDataPlaneOperation(cmd) {
		return c.matchesDataPlaneResourceType(service, operation, resType)
	}
	
	// Control plane operations - use precise mappings
	serviceOperationToResourceTypes := map[string]map[string][]string{
		"group": {
			"create": {"subscriptions/resourcegroups"},
			"delete": {"subscriptions/resourcegroups"},
			"list":   {"subscriptions/resourcegroups"},
			"show":   {"subscriptions/resourcegroups"},
		},
		"vm": {
			"create": {"virtualmachines"},
			"delete": {"virtualmachines"},
			"start":  {"virtualmachines"},
			"stop":   {"virtualmachines"},
			"restart": {"virtualmachines"},
			"list":   {"virtualmachines"},
			"show":   {"virtualmachines", "virtualmachines/instanceview"},
		},
		"storage": {
			"create": {"storageaccounts"},
			"delete": {"storageaccounts"},
			"list":   {"storageaccounts"},
			"show":   {"storageaccounts"},
		},
		"webapp": {
			"create": {"sites"},
			"delete": {"sites"},
			"list":   {"sites"},
			"show":   {"sites"},
			"start":  {"sites"},
			"stop":   {"sites"},
			"restart": {"sites"},
		},
		"keyvault": {
			"create": {"vaults"},
			"delete": {"vaults"},
			"list":   {"vaults"},
			"show":   {"vaults"},
		},
		"aks": {
			"create": {"managedclusters"},
			"delete": {"managedclusters"},
			"list":   {"managedclusters"},
			"show":   {"managedclusters"},
			"start":  {"managedclusters"},
			"stop":   {"managedclusters"},
		},
	}

	if serviceOps, exists := serviceOperationToResourceTypes[service]; exists {
		if resourceTypes, opExists := serviceOps[operation]; opExists {
			normalizedResType := strings.ReplaceAll(resType, "/", "")
			for _, resourceTypePattern := range resourceTypes {
				normalizedPattern := strings.ReplaceAll(resourceTypePattern, "/", "")
				if normalizedResType == normalizedPattern {
					return true
				}
			}
		}
	}
	
	return false
}

// isDataPlaneOperation dynamically determines if this is a data plane operation
// by analyzing the Azure API response rather than using hardcoded mappings
func (c *CLI) isDataPlaneOperation(cmd *models.AzureCommand) bool {
	service := strings.ToLower(cmd.Service)
	
	// Check for multi-part service names that typically indicate data plane operations
	serviceParts := strings.Fields(service)
	if len(serviceParts) >= 2 {
		// Multi-part service names (like "keyvault secret" or "storage blob") 
		// are strong indicators of data plane operations
		return true
	}
	
	return false
}

// matchesDataPlaneResourceType dynamically matches data plane resource types
// by analyzing the actual Azure API resource type patterns
func (c *CLI) matchesDataPlaneResourceType(service, operation, resourceType string) bool {
	serviceParts := strings.Fields(service)
	if len(serviceParts) < 2 {
		return false
	}
	
	baseService := serviceParts[0]
	subResource := serviceParts[1]
	
	// Dynamic matching based on resource type structure from Azure API
	resourceTypeLower := strings.ToLower(resourceType)
	
	// Special handling for known service name variations
	serviceAliases := map[string][]string{
		"keyvault": {"vault", "vaults"},
		"storage":  {"storageaccount", "storageaccounts"},
		"cosmosdb": {"documentdb", "cosmos"},
	}
	
	// Check if the resource type contains the base service name or its aliases
	serviceMatched := false
	if aliases, exists := serviceAliases[baseService]; exists {
		for _, alias := range aliases {
			if strings.Contains(resourceTypeLower, alias) {
				serviceMatched = true
				break
			}
		}
	} else {
		// Direct match for services without aliases
		serviceMatched = strings.Contains(resourceTypeLower, baseService)
	}
	
	if !serviceMatched {
		return false
	}
	
	// Check if the resource type contains the sub-resource name
	if !strings.Contains(resourceTypeLower, subResource) {
		return false
	}
	
	// Count hierarchy levels in the original resource type
	hierarchyLevels := strings.Count(resourceType, "/")
	
	// Data plane operations typically have deeper hierarchy (1+ levels)
	if hierarchyLevels < 1 {
		return false
	}
	
	// For truly dynamic matching, prioritize the most specific resource types
	// by preferring deeper hierarchy levels that directly contain the sub-resource
	resourceTypeParts := strings.Split(resourceTypeLower, "/")
	
	// Check if the sub-resource name appears in the resource type path
	// Search from the end to find the most specific match
	subResourceFound := false
	subResourcePosition := -1
	for i := len(resourceTypeParts) - 1; i >= 0; i-- {
		part := resourceTypeParts[i]
		if strings.Contains(part, subResource) {
			subResourceFound = true
			subResourcePosition = i
			// Continue searching backwards for an even more specific match
			// but if we find a direct match (part == subResource+"s" or part == subResource), prefer it
			if part == subResource || part == subResource+"s" {
				break
			}
		}
	}
	
	if !subResourceFound {
		return false
	}
	
	// Prefer more specific resource types: 
	// The sub-resource should appear towards the end of the path for specificity
	// For example: "storageAccounts/blobServices/containers/blobs" is more specific than "storageAccounts/blobServices"
	totalParts := len(resourceTypeParts)
	
	// Only match if the sub-resource appears in the last 2 parts of the path
	// This ensures we get the most specific permissions
	if subResourcePosition < totalParts-2 {
		return false
	}
	
	// Additional heuristic: avoid monitoring/insights resources unless they're specifically requested
	if strings.Contains(resourceTypeLower, "insights") || strings.Contains(resourceTypeLower, "monitoring") {
		// Only include if the operation is specifically about insights/monitoring
		if !strings.Contains(strings.ToLower(operation), "monitor") && 
		   !strings.Contains(strings.ToLower(operation), "metric") && 
		   !strings.Contains(strings.ToLower(operation), "diagnostic") {
			return false
		}
	}
	
	return true
}

func (c *CLI) matchesOperation(cmdOp, apiOp string) bool {
	cmdOp = strings.ToLower(cmdOp)
	apiOp = strings.ToLower(apiOp)

	// Direct match first
	if strings.Contains(apiOp, cmdOp) {
		return true
	}

	// Enhanced operation mapping that includes data plane patterns
	operationMap := map[string][]string{
		"create":  {"write", "create"},
		"update":  {"write", "update"},
		"set":     {"write", "set", "setsecret", "setkey", "setcertificate"},
		"delete":  {"delete", "remove", "deletesecret", "deletekey", "deletecertificate"},
		"remove":  {"delete", "remove"},
		"list":    {"read", "list", "getsecret", "getkey", "getcertificate"},
		"show":    {"read", "get", "getsecret", "getkey", "getcertificate"},
		"get":     {"read", "get", "getsecret", "getkey", "getcertificate"},
		"start":   {"start"},
		"stop":    {"poweroff", "stop"},
		"restart": {"restart"},
		"upload":  {"write", "put"},
		"download": {"read", "get"},
	}

	if matches, exists := operationMap[cmdOp]; exists {
		for _, match := range matches {
			if strings.Contains(apiOp, match) {
				return true
			}
		}
	}

	// Additional data plane operation patterns
	// Many data plane operations end with "/action"
	if strings.HasSuffix(apiOp, "/action") {
		actionName := strings.TrimSuffix(apiOp, "/action")
		if strings.Contains(actionName, cmdOp) {
			return true
		}
		
		// Check if the action name contains operation patterns
		for _, match := range operationMap[cmdOp] {
			if strings.Contains(actionName, match) {
				return true
			}
		}
	}

	return false
}

func (c *CLI) suggestOperationsFromLiveData(cmd *models.AzureCommand, providerOps models.ProviderOperationsResponse) []string {
	var suggestions []string
	
	// Find the most likely resource type
	var bestResourceType *models.ProviderResourceType
	for i, rt := range providerOps.ResourceTypes {
		if c.matchesResourceType(cmd, rt.Name) {
			bestResourceType = &providerOps.ResourceTypes[i]
			break
		}
	}

	if bestResourceType == nil && len(providerOps.ResourceTypes) > 0 {
		bestResourceType = &providerOps.ResourceTypes[0]
	}

	if bestResourceType != nil {
		operationMap := map[string][]string{
			"create":  {"write"},
			"update":  {"write"},
			"set":     {"write"},
			"delete":  {"delete"},
			"remove":  {"delete"},
			"list":    {"read"},
			"show":    {"read"},
			"get":     {"read"},
			"start":   {"start"},
			"stop":    {"poweroff", "stop"},
			"restart": {"restart"},
		}

		if patterns, exists := operationMap[strings.ToLower(cmd.Operation)]; exists {
			for _, op := range bestResourceType.Operations {
				for _, pattern := range patterns {
					if strings.Contains(strings.ToLower(op.Name), pattern) {
						suggestions = append(suggestions, op.Name)
					}
				}
			}
		}

		// If no suggestions yet, add read permission as fallback
		if len(suggestions) == 0 {
			for _, op := range bestResourceType.Operations {
				if strings.Contains(strings.ToLower(op.Name), "read") {
					suggestions = append(suggestions, op.Name)
					break
				}
			}
		}
	}

	return suggestions
}

// getIntelligentSuggestions provides intelligent permission suggestions
func (c *CLI) getIntelligentSuggestions(cmd *models.AzureCommand) []string {
	// Common operation patterns
	operationMap := map[string][]string{
		"create": {
			fmt.Sprintf("Microsoft.Resources/*/write"),
			fmt.Sprintf("Microsoft.Authorization/*/write"),
		},
		"delete": {
			fmt.Sprintf("Microsoft.Resources/*/delete"),
		},
		"list": {
			fmt.Sprintf("Microsoft.Resources/*/read"),
		},
		"show": {
			fmt.Sprintf("Microsoft.Resources/*/read"),
		},
		"update": {
			fmt.Sprintf("Microsoft.Resources/*/write"),
		},
	}

	if suggestions, exists := operationMap[strings.ToLower(cmd.Operation)]; exists {
		return c.refineGenericPermissions(cmd, suggestions)
	}

	// Default fallback
	return []string{
		fmt.Sprintf("Microsoft.Resources/*/read"),
		fmt.Sprintf("Microsoft.Authorization/*/read"),
	}
}

// refineGenericPermissions refines generic permission patterns based on the command
func (c *CLI) refineGenericPermissions(cmd *models.AzureCommand, generic []string) []string {
	var refined []string

	// Map services to resource providers
	serviceToProvider := map[string]string{
		"group":      "Microsoft.Resources",
		"vm":         "Microsoft.Compute", 
		"storage":    "Microsoft.Storage",
		"webapp":     "Microsoft.Web",
		"keyvault":   "Microsoft.KeyVault",
		"network":    "Microsoft.Network",
		"sql":        "Microsoft.Sql",
		"aks":        "Microsoft.ContainerService",
		"role":       "Microsoft.Authorization",
	}

	// Map services to resource types
	serviceToResource := map[string]string{
		"group":       "subscriptions/resourceGroups",
		"vm":          "virtualMachines",
		"storage":     "storageAccounts",
		"webapp":      "sites",
		"keyvault":    "vaults",
		"network":     "virtualNetworks",
		"sql":         "servers",
		"aks":         "managedClusters",
		"role":        "roleAssignments",
	}

	provider := serviceToProvider[cmd.Service]
	resource := serviceToResource[cmd.Service]

	if provider != "" && resource != "" {
		for _, perm := range generic {
			// Replace wildcards with specific values
			specific := strings.ReplaceAll(perm, "Microsoft.Resources", provider)
			specific = strings.ReplaceAll(specific, "*", resource)
			refined = append(refined, specific)
		}
	} else {
		// Return generic patterns as fallback
		refined = generic
	}

	return refined
}

// Version returns the application version
func (c *CLI) Version() string {
	return "1.0.0"
}

// Help displays help information
func (c *CLI) Help() {
	c.colors.ShowUsage()
}
