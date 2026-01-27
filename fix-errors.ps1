# VirtEngine Error Code Allocation
# Range: 100-999 (avoiding Cosmos SDK 1-50, IBC-Go 1-50, CosmWasm 1-50)

$modules = @{
    'x/veid/types' = 100
    'x/mfa/types' = 200
    'x/encryption/types' = 300
    'x/roles/types' = 400
    'x/settlement/types' = 500
    'x/config/types' = 600
    'x/benchmark/types' = 700
    'x/delegation/types' = 800
    'x/enclave/types' = 900
    'x/fraud/types' = 1000
    'x/hpc/types' = 1100
    'x/market/types/marketplace' = 1200
    'x/review/types' = 1300
    'x/staking/types' = 1400
    'pkg/artifact_store' = 1500
    'sdk/go/node/bme/v1' = 1600
}

foreach ($module in $modules.Keys) {
    $startCode = $modules[$module]
    $file = Join-Path $PSScriptRoot $module 'errors.go'
    if ($module -eq 'pkg/artifact_store' -or $module -eq 'sdk/go/node/bme/v1') {
        $file = Join-Path $PSScriptRoot $module 'errors.go'
    }
    
    if (Test-Path $file) {
        $content = Get-Content $file -Raw
        $newCode = $startCode
        $content = $content -replace 'Register\(([^,]+),\s*\d+,', {
            param($match)
            $result = "Register($($match.Groups[1].Value), $newCode,"
            $script:newCode++
            return $result
        }
        Set-Content $file $content -NoNewline
        Write-Host "Updated $module to start at $startCode"
    }
}
