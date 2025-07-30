# Comprehensive Azure CLI Permissions Test Script
# Tests both Management Plane and Data Plane operations against Microsoft official documentation
# Author: Azure CLI Permissions Tool
# Date: July 31, 2025

param(
    [switch]$Verbose,
    [switch]$DebugMode,
    [string]$AzPermPath = ".\azperm.exe"
)

# ANSI color codes for output formatting
$Red = "`e[31m"
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Magenta = "`e[35m"
$Cyan = "`e[36m"
$Reset = "`e[0m"

# Test results tracking
$script:TotalTests = 0
$script:PassedTests = 0
$script:FailedTests = 0
$script:TestResults = @()

# Official expected permissions from Microsoft documentation
$ExpectedPermissions = @{
    # MANAGEMENT PLANE OPERATIONS
    
    # Resource Group Operations (Control Plane)
    "az group create --name myResourceGroup --location westeurope" = @(
        "Microsoft.Resources/subscriptions/resourceGroups/write"
    )
    
    "az group delete --name myResourceGroup --yes" = @(
        "Microsoft.Resources/subscriptions/resourceGroups/delete"
    )
    
    "az group list" = @(
        "Microsoft.Resources/subscriptions/resourceGroups/read"
    )
    
    "az group show --name myResourceGroup" = @(
        "Microsoft.Resources/subscriptions/resourceGroups/read"
    )
    
    # Virtual Machine Operations (Control Plane)
    "az vm create --resource-group myRG --name myVM --image Ubuntu2204" = @(
        "Microsoft.Compute/virtualMachines/write"
    )
    
    "az vm start --resource-group myRG --name myVM" = @(
        "Microsoft.Compute/virtualMachines/start/action"
    )
    
    "az vm stop --resource-group myRG --name myVM" = @(
        "Microsoft.Compute/virtualMachines/powerOff/action"
    )
    
    "az vm restart --resource-group myRG --name myVM" = @(
        "Microsoft.Compute/virtualMachines/restart/action"
    )
    
    "az vm show --resource-group myRG --name myVM" = @(
        "Microsoft.Compute/virtualMachines/read"
    )
    
    "az vm list --resource-group myRG" = @(
        "Microsoft.Compute/virtualMachines/read"
    )
    
    "az vm delete --resource-group myRG --name myVM --yes" = @(
        "Microsoft.Compute/virtualMachines/delete"
    )
    
    # Storage Account Operations (Control Plane)
    "az storage account create --name mystorageaccount --resource-group myRG" = @(
        "Microsoft.Storage/storageAccounts/accountLocks/write",
        "Microsoft.Storage/storageAccounts/accountMigrations/write",
        "Microsoft.Storage/storageAccounts/blobServices/write",
        "Microsoft.Storage/storageAccounts/consumerDataSharePolicies/write",
        "Microsoft.Storage/storageAccounts/dataSharePolicies/write",
        "Microsoft.Storage/storageAccounts/encryptionScopes/write",
        "Microsoft.Storage/storageAccounts/fileServices/write",
        "Microsoft.Storage/storageAccounts/fileServices/writeFileBackupSemantics/action",
        "Microsoft.Storage/storageAccounts/hoboConfigurations/write",
        "Microsoft.Storage/storageAccounts/inventoryPolicies/write",
        "Microsoft.Storage/storageAccounts/localusers/write",
        "Microsoft.Storage/storageAccounts/managementPolicies/write",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterAssociationProxies/write",
        "Microsoft.Storage/storageAccounts/objectReplicationPolicies/write",
        "Microsoft.Storage/storageAccounts/privateEndpointConnectionProxies/write",
        "Microsoft.Storage/storageAccounts/privateEndpointConnections/write",
        "Microsoft.Storage/storageAccounts/queueServices/write",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/write",
        "Microsoft.Storage/storageAccounts/tableServices/write"
    )
    
    "az storage account show --name mystorageaccount --resource-group myRG" = @(
        "Microsoft.Storage/storageAccounts/accountLocks/read",
        "Microsoft.Storage/storageAccounts/accountMigrations/read",
        "Microsoft.Storage/storageAccounts/blobServices/getInfo/action",
        "Microsoft.Storage/storageAccounts/blobServices/read",
        "Microsoft.Storage/storageAccounts/consumerDataSharePolicies/read",
        "Microsoft.Storage/storageAccounts/dataSharePolicies/read",
        "Microsoft.Storage/storageAccounts/encryptionScopes/read",
        "Microsoft.Storage/storageAccounts/fileServices/read",
        "Microsoft.Storage/storageAccounts/fileServices/readFileBackupSemantics/action",
        "Microsoft.Storage/storageAccounts/hoboConfigurations/read",
        "Microsoft.Storage/storageAccounts/inventoryPolicies/read",
        "Microsoft.Storage/storageAccounts/localusers/read",
        "Microsoft.Storage/storageAccounts/managementPolicies/read",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterAssociationProxies/read",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterConfigurations/read",
        "Microsoft.Storage/storageAccounts/objectReplicationPolicies/read",
        "Microsoft.Storage/storageAccounts/privateEndpointConnectionProxies/read",
        "Microsoft.Storage/storageAccounts/privateEndpointConnections/read",
        "Microsoft.Storage/storageAccounts/privateLinkResources/read",
        "Microsoft.Storage/storageAccounts/queueServices/read",
        "Microsoft.Storage/storageAccounts/reports/read",
        "Microsoft.Storage/storageAccounts/restorePoints/read",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/delete",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/read",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/write",
        "Microsoft.Storage/storageAccounts/tableServices/read"
    )
    
    "az storage account list --resource-group myRG" = @(
        "Microsoft.Storage/storageAccounts/accountLocks/read",
        "Microsoft.Storage/storageAccounts/accountMigrations/read",
        "Microsoft.Storage/storageAccounts/blobServices/read",
        "Microsoft.Storage/storageAccounts/consumerDataSharePolicies/read",
        "Microsoft.Storage/storageAccounts/dataSharePolicies/read",
        "Microsoft.Storage/storageAccounts/encryptionScopes/read",
        "Microsoft.Storage/storageAccounts/fileServices/read",
        "Microsoft.Storage/storageAccounts/fileServices/readFileBackupSemantics/action",
        "Microsoft.Storage/storageAccounts/hoboConfigurations/read",
        "Microsoft.Storage/storageAccounts/inventoryPolicies/read",
        "Microsoft.Storage/storageAccounts/localusers/listKeys/action",
        "Microsoft.Storage/storageAccounts/localusers/read",
        "Microsoft.Storage/storageAccounts/managementPolicies/read",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterAssociationProxies/read",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterConfigurations/read",
        "Microsoft.Storage/storageAccounts/objectReplicationPolicies/read",
        "Microsoft.Storage/storageAccounts/privateEndpointConnectionProxies/read",
        "Microsoft.Storage/storageAccounts/privateEndpointConnections/read",
        "Microsoft.Storage/storageAccounts/privateLinkResources/read",
        "Microsoft.Storage/storageAccounts/queueServices/read",
        "Microsoft.Storage/storageAccounts/reports/read",
        "Microsoft.Storage/storageAccounts/restorePoints/read",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/read",
        "Microsoft.Storage/storageAccounts/tableServices/read"
    )
    
    "az storage account delete --name mystorageaccount --resource-group myRG --yes" = @(
        "Microsoft.Storage/storageAccounts/accountLocks/delete",
        "Microsoft.Storage/storageAccounts/accountLocks/deleteLock/action",
        "Microsoft.Storage/storageAccounts/dataSharePolicies/delete",
        "Microsoft.Storage/storageAccounts/inventoryPolicies/delete",
        "Microsoft.Storage/storageAccounts/localUsers/delete",
        "Microsoft.Storage/storageAccounts/managementPolicies/delete",
        "Microsoft.Storage/storageAccounts/networkSecurityPerimeterAssociationProxies/delete",
        "Microsoft.Storage/storageAccounts/objectReplicationPolicies/delete",
        "Microsoft.Storage/storageAccounts/privateEndpointConnectionProxies/delete",
        "Microsoft.Storage/storageAccounts/privateEndpointConnections/delete",
        "Microsoft.Storage/storageAccounts/restorePoints/delete",
        "Microsoft.Storage/storageAccounts/storageTaskAssignments/delete"
    )
    
    # Key Vault Operations (Control Plane)
    "az keyvault create --name mykeyvault --resource-group myRG" = @(
        "Microsoft.KeyVault/vaults/write"
    )
    
    "az keyvault show --name mykeyvault --resource-group myRG" = @(
        "Microsoft.KeyVault/vaults/read"
    )
    
    "az keyvault list --resource-group myRG" = @(
        "Microsoft.KeyVault/vaults/read"
    )
    
    "az keyvault delete --name mykeyvault --resource-group myRG" = @(
        "Microsoft.KeyVault/vaults/delete"
    )
    
    # DATA PLANE OPERATIONS
    
    # Key Vault Data Plane Operations
    "az keyvault secret show --name mysecret --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/secrets/getSecret/action",
        "Microsoft.KeyVault/vaults/secrets/read",
        "Microsoft.KeyVault/vaults/secrets/readMetadata/action"
    )
    
    "az keyvault secret set --name mysecret --vault-name mykeyvault --value myvalue" = @(
        "Microsoft.KeyVault/vaults/secrets/setSecret/action",
        "Microsoft.KeyVault/vaults/secrets/write"
    )
    
    "az keyvault secret list --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/secrets/getSecret/action",
        "Microsoft.KeyVault/vaults/secrets/read",
        "Microsoft.KeyVault/vaults/secrets/readMetadata/action"
    )
    
    "az keyvault secret delete --name mysecret --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/secrets/delete"
    )
    
    "az keyvault key show --name mykey --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/keyrotationpolicies/read",
        "Microsoft.KeyVault/vaults/keys/read",
        "Microsoft.KeyVault/vaults/keys/versions/read"
    )
    
    "az keyvault key create --name mykey --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/keyrotationpolicies/write",
        "Microsoft.KeyVault/vaults/keys/create/action",
        "Microsoft.KeyVault/vaults/keys/write"
    )
    
    "az keyvault certificate show --name mycert --vault-name mykeyvault" = @(
        "Microsoft.KeyVault/vaults/certificates/read"
    )
    
    # Storage Data Plane Operations
    "az storage blob show --name myblob --container-name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/tags/read",
        "Microsoft.Storage/storageAccounts/blobServices/containers/getAcl/action",
        "Microsoft.Storage/storageAccounts/blobServices/containers/read",
        "Microsoft.Storage/storageAccounts/blobServices/getInfo/action",
        "Microsoft.Storage/storageAccounts/blobServices/read"
    )
    
    "az storage blob upload --file myfile.txt --name myblob --container-name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/tags/write",
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write",
        "Microsoft.Storage/storageAccounts/blobServices/containers/write",
        "Microsoft.Storage/storageAccounts/blobServices/write"
    )
    
    "az storage blob download --name myblob --container-name mycontainer --account-name mystorageaccount --file myfile.txt" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"
    )
    
    "az storage blob delete --name myblob --container-name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/delete"
    )
    
    "az storage blob list --container-name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"
    )
    
    "az storage container create --name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/write"
    )
    
    "az storage container show --name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/read"
    )
    
    "az storage container list --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/read"
    )
    
    "az storage container delete --name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/delete"
    )
    
    # Storage Blob Tags (Data Plane)
    "az storage blob tag list --name myblob --container-name mycontainer --account-name mystorageaccount" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/tags/read"
    )
    
    "az storage blob tag set --name myblob --container-name mycontainer --account-name mystorageaccount --tags key1=value1" = @(
        "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/tags/write"
    )
}

function Write-Header {
    param([string]$Title)
    Write-Host "`n$Cyan========================================$Reset" -ForegroundColor Cyan
    Write-Host "$Cyan $Title $Reset" -ForegroundColor Cyan
    Write-Host "$Cyan========================================$Reset`n" -ForegroundColor Cyan
}

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message,
        [array]$Expected = @(),
        [array]$Actual = @()
    )
    
    $script:TotalTests++
    
    if ($Passed) {
        $script:PassedTests++
        Write-Host "  $Green✓ PASS$Reset - $TestName" -ForegroundColor Green
        if ($Verbose -and $Message) {
            Write-Host "    $Message" -ForegroundColor Gray
        }
    } else {
        $script:FailedTests++
        Write-Host "  $Red✗ FAIL$Reset - $TestName" -ForegroundColor Red
        Write-Host "    $Red$Message$Reset" -ForegroundColor Red
        
        if ($Expected.Count -gt 0) {
            Write-Host "    ${Yellow}Expected:$Reset" -ForegroundColor Yellow
            foreach ($perm in $Expected) {
                Write-Host "      • $perm" -ForegroundColor Gray
            }
        }
        
        if ($Actual.Count -gt 0) {
            Write-Host "    ${Yellow}Actual:$Reset" -ForegroundColor Yellow
            foreach ($perm in $Actual) {
                Write-Host "      • $perm" -ForegroundColor Gray
            }
        }
    }
    
    $script:TestResults += [PSCustomObject]@{
        TestName = $TestName
        Command = ""
        Passed = $Passed
        Message = $Message
        Expected = $Expected
        Actual = $Actual
    }
}

function Test-AzPermCommand {
    param(
        [string]$Command,
        [array]$ExpectedPermissions,
        [string]$TestCategory = "General"
    )
    
    Write-Host "  ${Blue}Testing:$Reset $Command" -ForegroundColor Blue
    
    try {
        # Build the command arguments
        $commandArgs = @()
        if ($DebugMode) {
            $commandArgs += "--debug"
        }
        
        # Run the azperm command
        $result = Write-Output $Command | & $AzPermPath @commandArgs 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            Write-TestResult -TestName "$TestCategory - $Command" -Passed $false -Message "Command failed with exit code $LASTEXITCODE. Output: $($result -join ' ')" -Expected $ExpectedPermissions -Actual @()
            return
        }
        
        # Parse the output to extract permissions
        $actualPermissions = @()
        $inPermissionsSection = $false
        
        foreach ($line in $result) {
            $line = $line.ToString().Trim()
            
            # Look for the permissions section header
            if ($line -match "Required RBAC Permissions" -or $line -match "≡ƒöÉ") {
                $inPermissionsSection = $true
                continue
            }
            
            if ($inPermissionsSection) {
                # Stop at separator line (long line of dashes/Unicode box chars)
                if ($line -match "^[─Γö]{10,}$") {
                    break
                }
                
                # Skip empty lines
                if ($line -eq "") {
                    continue
                }
                
                # Look for Microsoft.* permissions anywhere in the line
                if ($line -match "(Microsoft\.[^\s]+)") {
                    $permission = $matches[1].Trim()
                    if ($permission -and $permission -notmatch "^(Command:|Parameters:)" -and $actualPermissions -notcontains $permission) {
                        $actualPermissions += $permission
                    }
                }
            }
        }
        
        # Compare expected vs actual permissions
        $success = $true
        $missingPermissions = @()
        $extraPermissions = @()
        
        # Check for missing expected permissions
        foreach ($expected in $ExpectedPermissions) {
            if ($expected -notin $actualPermissions) {
                $missingPermissions += $expected
                $success = $false
            }
        }
        
        # Check for extra permissions (not necessarily wrong, but worth noting)
        foreach ($actual in $actualPermissions) {
            if ($actual -notin $ExpectedPermissions) {
                $extraPermissions += $actual
            }
        }
        
        # Determine pass/fail and create message
        if ($success) {
            $message = "All expected permissions found"
            if ($extraPermissions.Count -gt 0) {
                $message += " (plus $($extraPermissions.Count) additional permissions)"
            }
            Write-TestResult -TestName "$TestCategory - $(($Command -split ' ')[0..2] -join ' ')" -Passed $true -Message $message -Expected $ExpectedPermissions -Actual $actualPermissions
        } else {
            $message = "Missing permissions: $($missingPermissions -join ', ')"
            if ($extraPermissions.Count -gt 0) {
                $message += ". Extra permissions: $($extraPermissions -join ', ')"
            }
            Write-TestResult -TestName "$TestCategory - $(($Command -split ' ')[0..2] -join ' ')" -Passed $false -Message $message -Expected $ExpectedPermissions -Actual $actualPermissions
        }
        
    } catch {
        Write-TestResult -TestName "$TestCategory - $Command" -Passed $false -Message "Exception occurred: $($_.Exception.Message)" -Expected $ExpectedPermissions -Actual @()
    }
}

function Test-Prerequisites {
    Write-Header "Prerequisites Check"
    
    # Check if azperm.exe exists
    if (-not (Test-Path $AzPermPath)) {
        Write-TestResult -TestName "azperm.exe availability" -Passed $false -Message "azperm.exe not found at path: $AzPermPath"
        return $false
    }
    Write-TestResult -TestName "azperm.exe availability" -Passed $true -Message "Found at $AzPermPath"
    
    # Check if user is logged into Azure CLI
    try {
        $account = az account show --output json 2>$null | ConvertFrom-Json
        if ($account -and $account.user) {
            Write-TestResult -TestName "Azure CLI authentication" -Passed $true -Message "Logged in as $($account.user.name)"
        } else {
            Write-TestResult -TestName "Azure CLI authentication" -Passed $false -Message "Not logged into Azure CLI. Run 'az login' first."
            return $false
        }
    } catch {
        Write-TestResult -TestName "Azure CLI authentication" -Passed $false -Message "Azure CLI not available or not logged in"
        return $false
    }
    
    return $true
}

function Invoke-ManagementPlaneTests {
    Write-Header "Management Plane Operations Tests"
    
    # Resource Group Tests
    Write-Host "  ${Magenta}Resource Group Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az group create --name myResourceGroup --location westeurope" $ExpectedPermissions["az group create --name myResourceGroup --location westeurope"] "Management Plane"
    Test-AzPermCommand "az group delete --name myResourceGroup --yes" $ExpectedPermissions["az group delete --name myResourceGroup --yes"] "Management Plane"
    Test-AzPermCommand "az group list" $ExpectedPermissions["az group list"] "Management Plane"
    Test-AzPermCommand "az group show --name myResourceGroup" $ExpectedPermissions["az group show --name myResourceGroup"] "Management Plane"
    
    # Virtual Machine Tests
    Write-Host "`n  ${Magenta}Virtual Machine Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az vm create --resource-group myRG --name myVM --image Ubuntu2204" $ExpectedPermissions["az vm create --resource-group myRG --name myVM --image Ubuntu2204"] "Management Plane"
    Test-AzPermCommand "az vm start --resource-group myRG --name myVM" $ExpectedPermissions["az vm start --resource-group myRG --name myVM"] "Management Plane"
    Test-AzPermCommand "az vm stop --resource-group myRG --name myVM" $ExpectedPermissions["az vm stop --resource-group myRG --name myVM"] "Management Plane"
    Test-AzPermCommand "az vm restart --resource-group myRG --name myVM" $ExpectedPermissions["az vm restart --resource-group myRG --name myVM"] "Management Plane"
    Test-AzPermCommand "az vm show --resource-group myRG --name myVM" $ExpectedPermissions["az vm show --resource-group myRG --name myVM"] "Management Plane"
    Test-AzPermCommand "az vm list --resource-group myRG" $ExpectedPermissions["az vm list --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az vm delete --resource-group myRG --name myVM --yes" $ExpectedPermissions["az vm delete --resource-group myRG --name myVM --yes"] "Management Plane"
    
    # Storage Account Tests
    Write-Host "`n  ${Magenta}Storage Account Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az storage account create --name mystorageaccount --resource-group myRG" $ExpectedPermissions["az storage account create --name mystorageaccount --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az storage account show --name mystorageaccount --resource-group myRG" $ExpectedPermissions["az storage account show --name mystorageaccount --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az storage account list --resource-group myRG" $ExpectedPermissions["az storage account list --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az storage account delete --name mystorageaccount --resource-group myRG --yes" $ExpectedPermissions["az storage account delete --name mystorageaccount --resource-group myRG --yes"] "Management Plane"
    
    # Key Vault Tests
    Write-Host "`n  ${Magenta}Key Vault Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az keyvault create --name mykeyvault --resource-group myRG" $ExpectedPermissions["az keyvault create --name mykeyvault --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az keyvault show --name mykeyvault --resource-group myRG" $ExpectedPermissions["az keyvault show --name mykeyvault --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az keyvault list --resource-group myRG" $ExpectedPermissions["az keyvault list --resource-group myRG"] "Management Plane"
    Test-AzPermCommand "az keyvault delete --name mykeyvault --resource-group myRG" $ExpectedPermissions["az keyvault delete --name mykeyvault --resource-group myRG"] "Management Plane"
}

function Invoke-DataPlaneTests {
    Write-Header "Data Plane Operations Tests"
    
    # Key Vault Data Plane Tests
    Write-Host "  ${Magenta}Key Vault Data Plane Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az keyvault secret show --name mysecret --vault-name mykeyvault" $ExpectedPermissions["az keyvault secret show --name mysecret --vault-name mykeyvault"] "Data Plane"
    Test-AzPermCommand "az keyvault secret set --name mysecret --vault-name mykeyvault --value myvalue" $ExpectedPermissions["az keyvault secret set --name mysecret --vault-name mykeyvault --value myvalue"] "Data Plane"
    Test-AzPermCommand "az keyvault secret list --vault-name mykeyvault" $ExpectedPermissions["az keyvault secret list --vault-name mykeyvault"] "Data Plane"
    Test-AzPermCommand "az keyvault secret delete --name mysecret --vault-name mykeyvault" $ExpectedPermissions["az keyvault secret delete --name mysecret --vault-name mykeyvault"] "Data Plane"
    Test-AzPermCommand "az keyvault key show --name mykey --vault-name mykeyvault" $ExpectedPermissions["az keyvault key show --name mykey --vault-name mykeyvault"] "Data Plane"
    Test-AzPermCommand "az keyvault key create --name mykey --vault-name mykeyvault" $ExpectedPermissions["az keyvault key create --name mykey --vault-name mykeyvault"] "Data Plane"
    Test-AzPermCommand "az keyvault certificate show --name mycert --vault-name mykeyvault" $ExpectedPermissions["az keyvault certificate show --name mycert --vault-name mykeyvault"] "Data Plane"
    
    # Storage Data Plane Tests
    Write-Host "`n  ${Magenta}Storage Data Plane Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az storage blob show --name myblob --container-name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage blob show --name myblob --container-name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage blob upload --file myfile.txt --name myblob --container-name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage blob upload --file myfile.txt --name myblob --container-name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage blob download --name myblob --container-name mycontainer --account-name mystorageaccount --file myfile.txt" $ExpectedPermissions["az storage blob download --name myblob --container-name mycontainer --account-name mystorageaccount --file myfile.txt"] "Data Plane"
    Test-AzPermCommand "az storage blob delete --name myblob --container-name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage blob delete --name myblob --container-name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage blob list --container-name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage blob list --container-name mycontainer --account-name mystorageaccount"] "Data Plane"
    
    # Storage Container Tests
    Write-Host "`n  ${Magenta}Storage Container Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az storage container create --name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage container create --name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage container show --name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage container show --name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage container list --account-name mystorageaccount" $ExpectedPermissions["az storage container list --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage container delete --name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage container delete --name mycontainer --account-name mystorageaccount"] "Data Plane"
    
    # Storage Blob Tags Tests
    Write-Host "`n  ${Magenta}Storage Blob Tags Operations$Reset" -ForegroundColor Magenta
    Test-AzPermCommand "az storage blob tag list --name myblob --container-name mycontainer --account-name mystorageaccount" $ExpectedPermissions["az storage blob tag list --name myblob --container-name mycontainer --account-name mystorageaccount"] "Data Plane"
    Test-AzPermCommand "az storage blob tag set --name myblob --container-name mycontainer --account-name mystorageaccount --tags key1=value1" $ExpectedPermissions["az storage blob tag set --name myblob --container-name mycontainer --account-name mystorageaccount --tags key1=value1"] "Data Plane"
}

function Show-Summary {
    Write-Header "Test Summary"
    
    $passRate = if ($script:TotalTests -gt 0) { [math]::Round(($script:PassedTests / $script:TotalTests) * 100, 2) } else { 0 }
    
    Write-Host "  ${Green}Passed:$Reset $script:PassedTests" -ForegroundColor Green
    Write-Host "  ${Red}Failed:$Reset $script:FailedTests" -ForegroundColor Red
    Write-Host "  ${Blue}Total:$Reset  $script:TotalTests" -ForegroundColor Blue
    Write-Host "  ${Yellow}Pass Rate:$Reset $passRate%" -ForegroundColor Yellow
    
    if ($script:FailedTests -gt 0) {
        Write-Host "`n  ${Yellow}Failed Tests Summary:$Reset" -ForegroundColor Yellow
        $failedTests = $script:TestResults | Where-Object { -not $_.Passed }
        foreach ($test in $failedTests) {
            Write-Host "    $Red• $($test.TestName)$Reset" -ForegroundColor Red
            Write-Host "      $($test.Message)" -ForegroundColor Gray
        }
    }
    
    Write-Host "`n  ${Cyan}Test completed at $(Get-Date)$Reset" -ForegroundColor Cyan
    
    # Return exit code based on test results
    if ($script:FailedTests -gt 0) {
        exit 1
    } else {
        exit 0
    }
}

# Main execution
try {
    Write-Header "Azure CLI Permissions Comprehensive Test Suite"
    Write-Host "Testing both Management Plane and Data Plane operations against official Microsoft documentation"
    Write-Host "Started at $(Get-Date)"
    
    if ($DebugMode) {
        Write-Host "${Yellow}Debug mode enabled - will show detailed output from azperm$Reset" -ForegroundColor Yellow
    }
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        Write-Host "${Red}Prerequisites check failed. Cannot continue with tests.$Reset" -ForegroundColor Red
        exit 1
    }
    
    # Build the application if it doesn't exist
    if (-not (Test-Path $AzPermPath)) {
        Write-Host "${Yellow}Building azperm.exe...$Reset" -ForegroundColor Yellow
        $buildResult = go build -o azperm.exe . 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "${Red}Failed to build azperm.exe: $buildResult$Reset" -ForegroundColor Red
            exit 1
        }
        Write-Host "${Green}Build successful$Reset" -ForegroundColor Green
    }
    
    # Run the test suites
    Invoke-ManagementPlaneTests
    Invoke-DataPlaneTests
    
    # Show final summary
    Show-Summary
    
} catch {
    Write-Host "${Red}Fatal error during test execution: $($_.Exception.Message)$Reset" -ForegroundColor Red
    Write-Host "Stack trace: $($_.ScriptStackTrace)" -ForegroundColor Red
    exit 1
}
