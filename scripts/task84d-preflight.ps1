# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repo = Split-Path -Parent $PSScriptRoot
Push-Location $repo
try {
    Write-Host '[84D] checking conflict markers'
    $markers = & rg -l '^<<<<<<< |^>>>>>>> |^\|\|\|\|\|\|\| ' `
        -g '!vendor/**' -g '!node_modules/**' -g '!.cache/**' -g '!artifacts/**'
    if ($LASTEXITCODE -notin @(0, 1)) { throw 'ripgrep conflict-marker scan failed' }
    if ($markers) { throw "conflict markers remain:`n$($markers -join "`n")" }

    Write-Host '[84D] validating checked-in reconciliation artifact'
    $artifactPath = 'artifacts/mainnet/task84d-reconciliation.json'
    $artifactRaw = Get-Content $artifactPath -Raw
    $artifact = $artifactRaw | ConvertFrom-Json
    if ($artifact.schema_version -ne 1 -or $artifact.upgrade -ne 'v1.7.0') {
        throw 'unexpected Task 84D reconciliation schema or upgrade'
    }
    $counts = @(
        $artifact.settlement.payouts_scanned,
        $artifact.settlement.escrows_scanned,
        $artifact.settlement.cases_created,
        $artifact.settlement.claims_merged,
        $artifact.settlement.quarantined,
        $artifact.settlement.terminal_preserved,
        $artifact.settlement.already_migrated,
        $artifact.settlement.malformed_orphans
    )
    $bytes = [System.Collections.Generic.List[byte]]::new()
    foreach ($count in $counts) {
        $encoded = [BitConverter]::GetBytes([uint64]$count)
        if ([BitConverter]::IsLittleEndian) { [Array]::Reverse($encoded) }
        $bytes.AddRange($encoded)
    }
    $actualDigest = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes.ToArray())).ToLowerInvariant()
    if ($actualDigest -ne $artifact.settlement.digest_sha256) {
        throw "reconciliation digest mismatch: expected $($artifact.settlement.digest_sha256), got $actualDigest"
    }
    if ($artifact.settlement.digest_sha256 -notmatch '^[0-9a-f]{64}$') {
        throw 'reconciliation digest must be a lowercase SHA-256 value'
    }
    foreach ($name in @('fraud_claims', 'hpc_claims', 'billing_claims', 'review_claims', 'quarantined')) {
        $value = $artifact.adapters.$name
        if ($null -eq $value -or [int64]$value -lt 0) {
            throw "invalid adapter counter: $name"
        }
    }
    $canonicalArtifact = $artifact | ConvertTo-Json -Depth 10
    if ($canonicalArtifact.Trim() -ne $artifactRaw.Trim()) {
        throw "$artifactPath is not canonical PowerShell JSON"
    }
    $reportPath = '_docs/audits/task-84d-completion-report-2026-07-22.md'
    $report = Get-Content $reportPath -Raw
    if (-not $report.Contains($artifactPath) -or -not $report.Contains($artifact.settlement.digest_sha256)) {
        throw 'Task 84D report does not reference the exact reconciliation path and digest'
    }

    Write-Host '[84D] running focused Go suites'
    & go test ./x/settlement/... ./x/escrow/... ./x/fraud/... ./x/hpc/... ./x/review/... ./x/resources/... ./upgrades/software/v1.7.0 ./tests/upgrade -count=1
    if ($LASTEXITCODE -ne 0) { throw 'Task 84D focused Go tests failed' }

    Write-Host '[84D] checking formatting and whitespace'
    $goFiles = & git diff --name-only --diff-filter=ACMRT -- '*.go'
    if ($goFiles) {
        $unformatted = & gofmt -l @goFiles
        if ($unformatted) { throw "unformatted Go files:`n$($unformatted -join "`n")" }
    }
    & git diff --check
    if ($LASTEXITCODE -ne 0) { throw 'git diff --check failed' }

    Write-Host '[84D] preflight passed'
}
finally {
    Pop-Location
}
