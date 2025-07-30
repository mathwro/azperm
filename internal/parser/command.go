package parser

import (
	"fmt"
	"strings"

	"github.com/mathwro/AzCliPermissions/internal/models"
)

// ParseAzureCommand parses an Azure CLI command string into a structured command
func ParseAzureCommand(input string) (*models.AzureCommand, error) {
	// Remove 'az' prefix if present and normalize
	input = strings.TrimSpace(input)

	// Handle different input formats
	if strings.HasPrefix(input, "az ") {
		input = strings.TrimPrefix(input, "az ")
	}

	// Split command into parts
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid Azure CLI command format")
	}

	// Extract service and operation
	service := parts[0]
	operation := parts[1]

	// Handle multi-part services (e.g., "network vnet", "storage account")
	if len(parts) > 2 && !strings.HasPrefix(parts[2], "--") {
		service = parts[0] + " " + parts[1]
		operation = parts[2]
		parts = parts[1:] // Adjust parts for parameter parsing
	}

	// Parse parameters
	parameters := make(map[string]string)
	for i := 2; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--") {
			paramName := strings.TrimPrefix(parts[i], "--")
			paramValue := ""

			// Check if next part is the value (not another parameter)
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
				paramValue = parts[i+1]
				i++ // Skip the value in next iteration
			}

			parameters[paramName] = paramValue
		}
	}

	return &models.AzureCommand{
		Service:    service,
		Operation:  operation,
		Parameters: parameters,
		FullCmd:    fmt.Sprintf("%s %s", service, operation),
	}, nil
}
