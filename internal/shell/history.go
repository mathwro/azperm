package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	
	switch shell {
	case "powershell", "pwsh":
		// Read PowerShell history file directly without spawning new process
		return getLastAzureCommandFromPowerShellHistory()
	case "bash":
		cmd := exec.Command("bash", "-c", `history | grep -E "^\s*[0-9]+\s+az\s" | tail -1 | sed 's/^[ ]*[0-9]*[ ]*//'`)
		output, err := cmd.Output()
		if err != nil {
			return "", &ShellError{Shell: shell, Message: "could not read command history", Err: err}
		}
		commandLine := strings.TrimSpace(string(output))
		if commandLine == "" {
			return "", &ShellError{Shell: shell, Message: "no Azure CLI commands found in recent history"}
		}
		return commandLine, nil
	case "zsh":
		cmd := exec.Command("zsh", "-c", `fc -ln -1000 | grep -E "^\s*az\s" | tail -1 | sed 's/^[ ]*//'`)
		output, err := cmd.Output()
		if err != nil {
			return "", &ShellError{Shell: shell, Message: "could not read command history", Err: err}
		}
		commandLine := strings.TrimSpace(string(output))
		if commandLine == "" {
			return "", &ShellError{Shell: shell, Message: "no Azure CLI commands found in recent history"}
		}
		return commandLine, nil
	case "fish":
		cmd := exec.Command("fish", "-c", `history | grep -E "^az\s" | tail -1`)
		output, err := cmd.Output()
		if err != nil {
			return "", &ShellError{Shell: shell, Message: "could not read command history", Err: err}
		}
		commandLine := strings.TrimSpace(string(output))
		if commandLine == "" {
			return "", &ShellError{Shell: shell, Message: "no Azure CLI commands found in recent history"}
		}
		return commandLine, nil
	default:
		return "", &ShellError{Shell: shell, Message: "unsupported shell"}
	}
}

// getLastAzureCommandFromPowerShellHistory reads PowerShell history file directly
func getLastAzureCommandFromPowerShellHistory() (string, error) {
	// Get the PowerShell history file path
	historyPath, err := getPowerShellHistoryPath()
	if err != nil {
		return "", &ShellError{Shell: "powershell", Message: "could not find PowerShell history file", Err: err}
	}

	// Read the history file
	file, err := os.Open(historyPath)
	if err != nil {
		return "", &ShellError{Shell: "powershell", Message: "could not open PowerShell history file", Err: err}
	}
	defer file.Close()

	// Scan through the file and find the last Azure CLI command
	var lastAzCommand string
	azRegex := regexp.MustCompile(`^az\s`)
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if azRegex.MatchString(line) {
			lastAzCommand = line
		}
	}

	if err := scanner.Err(); err != nil {
		return "", &ShellError{Shell: "powershell", Message: "error reading PowerShell history file", Err: err}
	}

	if lastAzCommand == "" {
		return "", &ShellError{Shell: "powershell", Message: "no Azure CLI commands found in PowerShell history"}
	}

	return lastAzCommand, nil
}

// getPowerShellHistoryPath returns the path to the PowerShell history file
func getPowerShellHistoryPath() (string, error) {
	// Try to get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}

	// PowerShell history is typically stored in:
	// Windows: C:\Users\<username>\AppData\Roaming\Microsoft\Windows\PowerShell\PSReadLine\ConsoleHost_history.txt
	historyPath := filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
	
	// Check if the file exists
	if _, err := os.Stat(historyPath); err != nil {
		return "", fmt.Errorf("PowerShell history file not found at %s: %w", historyPath, err)
	}

	return historyPath, nil
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
