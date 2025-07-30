# AzCliPermissions (azperm) v2.1

A cross-platform CLI tool that analyzes Azure CLI commands and shows the required Azure RBAC permissions with **REST API integration** for maximum accuracy.

## 🚀 Quick Start

```bash
# Direct usage (easiest!)
azperm az group create --name myRG --location eastus
azperm az vm start --name myVM --resource-group myRG

# Traditional piping
echo 'az storage account create --name mystorageaccount' | azperm

# Analyze your last command
az keyvault create --name myVault --resource-group myRG
azperm --last
```

## ✨ New in v2.1: REST API Integration

- 🎯 **High Confidence** - Commands mapped to actual Azure REST API endpoints
- 📊 **Confidence Indicators** - Shows how certain we are about permissions (High/Medium/Low)
- 🔍 **Definitive Mappings** - No more guessing for common commands
- 📈 **Intelligent Fallback** - Smart inference for unknown commands

## Sample Output

```
🔍 Command: group create
📋 Parameters: --name myRG --location eastus

🔐 Required RBAC Permissions (High Confidence - REST API Verified):
  • Microsoft.Resources/subscriptions/resourceGroups/write
```

## Installation

### Pre-built Binaries (Recommended)

Download from the `dist/` directory:
- `azperm-windows-amd64.exe` - Windows 64-bit
- `azperm-linux-amd64` - Linux 64-bit  
- `azperm-darwin-amd64` - macOS Intel
- `azperm-darwin-arm64` - macOS Apple Silicon

### Build from Source

```bash
git clone <repository-url>
cd AzCliPermissions
go build -o azperm      # Linux/macOS
go build -o azperm.exe  # Windows

# Or build all platforms
./build.ps1
```

## Features

- ✅ **REST API Integration** - Definitive permissions from Azure REST API specs
- ✅ **Confidence Levels** - Know how certain the permissions are
- ✅ **Direct command support** - No quotes or pipes needed
- ✅ **Dynamic discovery** - Supports ALL Azure CLI commands (`azperm --discover`)
- ✅ **Cross-platform** - Windows, Linux, macOS
- ✅ **History analysis** - Analyze your last Azure CLI command
- ✅ **Single binary** - No dependencies

## Commands

```bash
azperm --help           # Show help
azperm --last           # Analyze last Azure CLI command from history
azperm --discover       # Update permissions database with REST API integration
azperm --version        # Show version
```

## Confidence Levels

| Level | Description | Source |
|-------|-------------|--------|
| 🟢 **High** | REST API Verified | Mapped to actual Azure REST API endpoints |
| 🟡 **Medium** | Pattern Matched | Found in curated database or intelligent mapping |
| 🟠 **Low** | Intelligent Guess | Inferred from command patterns |

## Examples by Confidence Level

### High Confidence (REST API Verified) ✅
```bash
azperm az group create --name myRG --location eastus
azperm az vm start --name myVM --resource-group myRG
azperm az storage account create --name mystorageaccount
azperm az keyvault create --name myVault --resource-group myRG
azperm az webapp create --name myWebApp --resource-group myRG
```

### Medium Confidence (Pattern Matched) ⚠️
```bash
azperm az cosmosdb create --name myCosmosDB --resource-group myRG
azperm az redis create --name myRedis --resource-group myRG
```

### Low Confidence (Intelligent Guess) 🤔
```bash
# Unknown or new services fall back to intelligent inference
# Tip: Run 'azperm --discover' to improve accuracy
```

## Requirements

- Azure CLI installed (`az --version`)
- For discovery feature: logged in to Azure (`az login`)
- Internet connection for REST API integration

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
