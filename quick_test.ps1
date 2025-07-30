#!/usr/bin/env pwsh
# Quick Azure CLI Permissions Test
# Tests core functionality with key scenarios

param(
    [switch]$Debug
)

# Colors for output
$Green = "`e[32m"
$Red = "`e[31m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

$testResults = @()

function Test-Command {
    param(
        [string]$Command,
        [string[]]$MustContain,
        [string]$TestName
    )
    
    Write-Host "${Blue}Testing: $TestName$Reset" -ForegroundColor Blue
    Write-Host "  Command: $Command" -ForegroundColor Gray
    
    try {
        $commandArgs = @()
        if ($Debug) { $commandArgs += "--debug" }
        
        $result = Write-Output $Command | .\azperm.exe @commandArgs 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  ${Red}✗ FAILED$Reset - Command failed with exit code $LASTEXITCODE" -ForegroundColor Red
            $script:testResults += @{ Name = $TestName; Status = "FAILED"; Reason = "Exit code $LASTEXITCODE" }
            return
        }
        
        $output = $result -join "`n"
        $allFound = $true
        $missing = @()
        
        foreach ($required in $MustContain) {
            if ($output -notmatch [regex]::Escape($required)) {
                $allFound = $false
                $missing += $required
            }
        }
        
        if ($allFound) {
            Write-Host "  ${Green}✓ PASSED$Reset - All required permissions found" -ForegroundColor Green
            $script:testResults += @{ Name = $TestName; Status = "PASSED"; Reason = "All permissions found" }
        } else {
            Write-Host "  ${Red}✗ FAILED$Reset - Missing: $($missing -join ', ')" -ForegroundColor Red
            $script:testResults += @{ Name = $TestName; Status = "FAILED"; Reason = "Missing: $($missing -join ', ')" }
        }
        
        if ($Debug) {
            Write-Host "  Output:" -ForegroundColor Gray
            foreach ($line in $result) {
                Write-Host "    $line" -ForegroundColor DarkGray
            }
        }
        
    } catch {
        Write-Host "  ${Red}✗ FAILED$Reset - Exception: $($_.Exception.Message)" -ForegroundColor Red
        $script:testResults += @{ Name = $TestName; Status = "FAILED"; Reason = $_.Exception.Message }
    }
    Write-Host ""
}

Write-Host "${Yellow}=== Quick Azure CLI Permissions Test ===$Reset" -ForegroundColor Yellow
Write-Host ""

# Build if needed
if (-not (Test-Path ".\azperm.exe")) {
    Write-Host "Building azperm.exe..." -ForegroundColor Yellow
    go build -o azperm.exe .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "${Red}Build failed!$Reset" -ForegroundColor Red
        exit 1
    }
}

# Core tests covering both management and data plane
$tests = @(
    @{
        Command = "az group create --name myResourceGroup --location westeurope"
        MustContain = @("Microsoft.Resources/subscriptions/resourceGroups/write")
        TestName = "Resource Group Create (Management Plane)"
    },
    @{
        Command = "az vm start --resource-group myRG --name myVM"
        MustContain = @("Microsoft.Compute/virtualMachines/start/action")
        TestName = "VM Start (Management Plane)"
    },
    @{
        Command = "az vm stop --resource-group myRG --name myVM"
        MustContain = @("Microsoft.Compute/virtualMachines/powerOff/action")
        TestName = "VM Stop (Management Plane)"
    },
    @{
        Command = "az keyvault secret show --name mysecret --vault-name mykeyvault"
        MustContain = @("Microsoft.KeyVault/vaults/secrets/getSecret/action", "Microsoft.KeyVault/vaults/secrets/readMetadata/action")
        TestName = "KeyVault Secret Show (Data Plane)"
    },
    @{
        Command = "az storage blob show --name myblob --container-name mycontainer --account-name mystorageaccount"
        MustContain = @("Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read")
        TestName = "Storage Blob Show (Data Plane)"
    },
    @{
        Command = "az storage blob upload --file myfile.txt --name myblob --container-name mycontainer --account-name mystorageaccount"
        MustContain = @("Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write")
        TestName = "Storage Blob Upload (Data Plane)"
    }
)

foreach ($test in $tests) {
    Test-Command -Command $test.Command -MustContain $test.MustContain -TestName $test.TestName
}

# Summary
Write-Host "${Yellow}=== Test Summary ===$Reset" -ForegroundColor Yellow
$passed = ($script:testResults | Where-Object { $_.Status -eq "PASSED" }).Count
$failed = ($script:testResults | Where-Object { $_.Status -eq "FAILED" }).Count
$total = $script:testResults.Count

Write-Host "Passed: ${Green}$passed$Reset" -ForegroundColor Green
Write-Host "Failed: ${Red}$failed$Reset" -ForegroundColor Red
Write-Host "Total:  ${Blue}$total$Reset" -ForegroundColor Blue

if ($failed -gt 0) {
    Write-Host "`nFailed tests:" -ForegroundColor Red
    $script:testResults | Where-Object { $_.Status -eq "FAILED" } | ForEach-Object {
        Write-Host "  • $($_.Name): $($_.Reason)" -ForegroundColor Red
    }
    exit 1
} else {
    Write-Host "`n${Green}All tests passed!$Reset" -ForegroundColor Green
    exit 0
}
