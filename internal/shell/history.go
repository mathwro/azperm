package shell

import (
	"os"
	"os/exec"
	"strings"
)

// DetectShell detects the current shell environment
func DetectShell() string {
	// Try to detect current shell
	if shell := os.Getenv("SHELL"); shell != "" {
		// Unix-like systems
		if strings.Contains(shell, "bash") {
			return "bash"
		}
		if strings.Contains(shell, "zsh") {
			return "zsh"
		}
		if strings.Contains(shell, "fish") {
			return "fish"
		}
	}
	
	// Check for PowerShell on any platform
	if _, err := exec.LookPath("pwsh"); err == nil {
		return "pwsh"
	}
	if _, err := exec.LookPath("powershell"); err == nil {
		return "powershell"
	}
	
	// Default fallbacks
	if os.Getenv("PSModulePath") != "" {
		// We're likely in PowerShell
		return "pwsh"
	}
	
	// Default to bash for Unix-like systems
	return "bash"
}

// GetLastAzureCommand retrieves the last Azure CLI command from shell history
func GetLastAzureCommand() (string, error) {
	shell := DetectShell()
	
	var cmd *exec.Cmd
	
	switch shell {
	case "powershell", "pwsh":
		cmd = exec.Command(shell, "-Command", `Get-History | Where-Object {$_.CommandLine -match '^az\s'} | Select-Object -Last 1 -ExpandProperty CommandLine`)
	case "bash":
		cmd = exec.Command("bash", "-c", `history | grep -E "^\s*[0-9]+\s+az\s" | tail -1 | sed 's/^[ ]*[0-9]*[ ]*//'`)
	case "zsh":
		cmd = exec.Command("zsh", "-c", `fc -ln -1000 | grep -E "^\s*az\s" | tail -1 | sed 's/^[ ]*//'`)
	case "fish":
		cmd = exec.Command("fish", "-c", `history | grep -E "^az\s" | tail -1`)
	default:
		return "", &ShellError{Shell: shell, Message: "unsupported shell"}
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", &ShellError{Shell: shell, Message: "could not read command history", Err: err}
	}

	commandLine := strings.TrimSpace(string(output))
	if commandLine == "" {
		return "", &ShellError{Shell: shell, Message: "no Azure CLI commands found in recent history"}
	}

	return commandLine, nil
}

// ShellError represents an error related to shell operations
type ShellError struct {
	Shell   string
	Message string
	Err     error
}

func (e *ShellError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ShellError) Unwrap() error {
	return e.Err
}
