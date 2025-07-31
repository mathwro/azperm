## Project Overview

This project is a cross-platform CLI tool written in Go. It analyzes Azure CLI commands and determines which Azure RBAC (Role-Based Access Control) permissions are required to execute those commands. The tool should work in both PowerShell and Linux-based shells.

Copilot Context & Instructions

    ✅ Copilot has access to the internet, including official Microsoft documentation (MCPs), through the GitHub Copilot extension in Visual Studio Code.

Copilot should:

    Use Microsoft Learn / Azure Docs and Azure CLI reference to understand command structures.

    Leverage the RBAC permission list to determine required permissions.

    Reference the Azure REST API documentation to infer which operations the CLI command maps to, and from that deduce required RBAC actions.

## Goal

Enable users to pipe Azure CLI commands into this tool and receive a list of required Azure permissions for that command. For example:

```bash
az group create --name myResourceGroup --location westeurope | azperm
```

The tool (`azperm`) should output a list of Azure RBAC actions such as:

```
Microsoft.Resources/subscriptions/resourceGroups/write
```

## Key Features

* Accept input via stdin (piped commands).
* Parse the Azure CLI command structure and extract:

  * Primary command and subcommand (e.g., `group create`)
  * Any relevant flags or arguments (e.g., `--name`, `--location`)
* Map Azure CLI commands to required RBAC actions.

* It should be dynamic so it can handle various Azure CLI commands without hardcoding every possible command.
* Support Windows (PowerShell) and Linux/macOS (Bash) terminals.
* Provide clear, colorized output if possible.
* Return non-zero exit codes for invalid or unsupported commands.

## Technical Constraints

* Language: Go (Golang)
* Must compile as a standalone binary for both Windows and Linux
* Avoid using OS-specific features or libraries
* Must always query the Azure CLI command structure and RBAC permissions dynamically, rather than hardcoding them.
* it should never generate any external files or require additional configuration files.

## Suggestions for Copilot

* Use the Cobra library (`spf13/cobra`) for CLI parsing if a CLI interface is needed later.
* Use the `os.Stdin` or `bufio.NewScanner(os.Stdin)` to read piped input.
* Use regex or tokenization to parse the Azure CLI input.
* Create a struct or map to model known Azure CLI commands and their required permissions.
* Focus on readable, idiomatic Go code.

## Example Permission Mapping (for testing)

```json
{
  "az group create": [
    "Microsoft.Resources/subscriptions/resourceGroups/write"
  ],
  "az vm start": [
    "Microsoft.Compute/virtualMachines/start/action"
  ]
}
```

## Future Features

* Option to fetch or sync permissions dynamically from Microsoft documentation or API (if available).
* Provide suggestions when the command is partially matched.
* Generate output in JSON or markdown.
