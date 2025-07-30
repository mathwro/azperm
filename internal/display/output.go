package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mathwro/AzCliPermissions/internal/models"
)

// Colors holds the color configurations for different output types
type Colors struct {
	Success *color.Color
	Error   *color.Color
	Warning *color.Color
	Info    *color.Color
	Header  *color.Color
}

// NewColors creates a new Colors instance with default color settings
func NewColors() *Colors {
	return &Colors{
		Success: color.New(color.FgGreen, color.Bold),
		Error:   color.New(color.FgRed, color.Bold),
		Warning: color.New(color.FgYellow, color.Bold),
		Info:    color.New(color.FgBlue, color.Bold),
		Header:  color.New(color.FgCyan, color.Bold),
	}
}

// DisplayPermissions shows permissions with confidence level indicators
func (c *Colors) DisplayPermissions(cmd *models.AzureCommand, permissions []string, confidence models.ConfidenceLevel) {
	// Header
	c.Header.Printf("🔍 Command: %s\n", cmd.FullCmd)

	if len(cmd.Parameters) > 0 {
		fmt.Printf("📋 Parameters: ")
		var paramList []string
		for key, value := range cmd.Parameters {
			if value != "" {
				paramList = append(paramList, fmt.Sprintf("--%s %s", key, value))
			} else {
				paramList = append(paramList, fmt.Sprintf("--%s", key))
			}
		}
		fmt.Printf("%s\n", strings.Join(paramList, " "))
	}

	fmt.Println()
	
	// Show confidence level
	switch confidence {
	case models.ConfidenceHigh:
		c.Success.Println("🔐 Required RBAC Permissions (High Confidence - REST API Verified):")
	case models.ConfidenceMedium:
		c.Info.Println("🔐 Required RBAC Permissions (Medium Confidence - Pattern Matched):")
	case models.ConfidenceLow:
		c.Warning.Println("🔐 Required RBAC Permissions (Low Confidence - Intelligent Guess):")
	default:
		c.Success.Println("🔐 Required RBAC Permissions:")
	}

	// Sort permissions for consistent output
	sort.Strings(permissions)

	for _, permission := range permissions {
		fmt.Printf("  • %s\n", permission)
	}

	// Add confidence explanation for lower confidence levels
	if confidence == models.ConfidenceLow {
		fmt.Println()
		c.Warning.Println("💡 Tip: Run 'azperm --discover' to improve accuracy with REST API integration")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()
}

// DisplayPermissionsWithLiveQuery shows permissions with live query indication
func (c *Colors) DisplayPermissionsWithLiveQuery(cmd *models.AzureCommand, permissions []string) {
	// Header  
	c.Header.Printf("🔍 Command: %s\n", cmd.FullCmd)

	if len(cmd.Parameters) > 0 {
		fmt.Printf("📋 Parameters: ")
		var paramList []string
		for key, value := range cmd.Parameters {
			if value != "" {
				paramList = append(paramList, fmt.Sprintf("--%s %s", key, value))
			} else {
				paramList = append(paramList, fmt.Sprintf("--%s", key))
			}
		}
		fmt.Printf("%s\n", strings.Join(paramList, " "))
	}

	fmt.Println()
	
	// Always show as live queried
	c.Success.Println("🔐 Required RBAC Permissions:")

	// Sort permissions for consistent output
	sort.Strings(permissions)

	for _, permission := range permissions {
		fmt.Printf("  • %s\n", permission)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()
}

// ShowUsage displays the usage information
func (c *Colors) ShowUsage() {
	c.Header.Println("Azure CLI Permissions Analyzer (azperm) v2.2")
	fmt.Println()
	c.Info.Println("USAGE:")
	fmt.Println("  # Method 1: Direct command arguments")
	c.Header.Println("  azperm az group create --name myRG --location eastus")
	c.Header.Println("  azperm az vm start --name myVM --resource-group myRG")
	fmt.Println()
	fmt.Println("  # Method 2: Pipe command")
	fmt.Println("  echo 'az group create --name myRG --location eastus' | azperm")
	fmt.Println("  echo 'az vm start --name myVM --resource-group myRG' | azperm")
	fmt.Println()
	c.Info.Println("FLAGS:")
	fmt.Println("  --version, -v           Show version information")
	fmt.Println("  --help, -h              Show this help message")
	fmt.Println("  --debug, -d             Enable debug mode with verbose output")
	fmt.Println()
	c.Info.Println("DESCRIPTION:")
	fmt.Println("  This tool analyzes Azure CLI commands and shows the required RBAC permissions.")
	fmt.Println("  🌐 ALWAYS queries live from Azure Management API for real-time accuracy!")
	fmt.Println()
	c.Info.Println("EXAMPLES:")
	fmt.Println("  # Direct usage (always live from Azure API)")
	c.Header.Println("  azperm az vm create --name myVM --resource-group myRG")
	c.Header.Println("  azperm az storage account create --name mystorageaccount")
	fmt.Println()
	fmt.Println("  # Piped usage (always live from Azure API)")
	fmt.Println("  echo 'az vm create --name myVM --resource-group myRG' | azperm")
	fmt.Println("  echo 'az storage account create --name mystorageaccount' | azperm")
	fmt.Println("  echo 'az keyvault secret set --vault-name myVault --name mySecret' | azperm")
	fmt.Println()
	c.Info.Println("FEATURES:")
	fmt.Println("  🌐 ALWAYS uses live Azure REST API for definitive permissions")
	fmt.Println("  ✅ Real-time accuracy - no cached or outdated data")
	fmt.Println("  ✅ Dynamic discovery of ALL Azure CLI commands")
	fmt.Println("  ✅ Cross-platform support (Windows, Linux, macOS)")
	fmt.Println()
	c.Warning.Println("REQUIREMENTS:")
	fmt.Println("  • Azure CLI installed and logged in (az login)")
	fmt.Println("  • Internet connection for live Azure API integration")
}

// ShowNoPermissionsWarning displays a warning when no permissions are found
func (c *Colors) ShowNoPermissionsWarning(command string, isLive bool) {
	c.Warning.Printf("⚠️  No permissions found for command: %s\n", command)
	c.Info.Println("   Command may not be supported yet or may not require specific RBAC permissions")
	if isLive {
		c.Info.Println("   💡 This was queried live from Azure API")
	} else {
		c.Info.Println("   💡 Try running 'azperm --discover' to update the permission database")
	}
}
