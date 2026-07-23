# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Genesis,
    [string]$Output
)

$ErrorActionPreference = 'Stop'
$document = Get-Content -LiteralPath $Genesis -Raw | ConvertFrom-Json -Depth 100
$state = $document.app_state
if ($null -eq $state) { throw 'genesis has no app_state' }

$market = $state.market
$catalog = $state.mktplace
$resources = $state.resources
$hpc = $state.hpc

$repositoryRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$resolvedGenesis = (Resolve-Path -LiteralPath $Genesis).Path
$sourcePath = [System.IO.Path]::GetRelativePath($repositoryRoot, $resolvedGenesis).Replace('\', '/')

function Count-Items($value) {
    if ($null -eq $value) { return 0 }
    return @($value).Count
}

$report = [ordered]@{
    schema_version = 1
    source = $sourcePath
    source_sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $resolvedGenesis).Hash.ToLowerInvariant()
    market = [ordered]@{
        orders = Count-Items $market.orders
        bids = Count-Items $market.bids
        leases = Count-Items $market.leases
    }
    non_owner = [ordered]@{
        offerings = Count-Items $catalog.offerings
        orders = Count-Items $catalog.orders
        bids = Count-Items $catalog.bids
        allocations = Count-Items $catalog.allocations
    }
    resources = [ordered]@{
        inventories = Count-Items $resources.inventories
        allocations = Count-Items $resources.allocations
        reservations = Count-Items $resources.reservations
    }
    hpc = [ordered]@{
        jobs = Count-Items $hpc.jobs
    }
    policy = [ordered]@{
        ambiguous_records = 'quarantine'
        synthetic_capacity = 'forbidden'
        non_owner_mutation_after_activation = 'rejected'
    }
}

$json = $report | ConvertTo-Json -Depth 8
if ($Output) {
    $json | Set-Content -LiteralPath $Output -Encoding utf8NoBOM
}
$json