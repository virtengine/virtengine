#!/usr/bin/env pwsh
# Task 85D deterministic local engineering gate.
[CmdletBinding()]
param()

$nestedPreflightSkipMessage = 'Task 85D preflight recursion guard active; skipping nested dispatch.'
if ($env:VE_TASK85D_PREFLIGHT_ACTIVE -eq '1') {
    Write-Host $nestedPreflightSkipMessage
    exit 0
}

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Assert-Command {
    param([Parameter(Mandatory)][string]$Name)

    if ($null -eq (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name is required for the Task 85D preflight"
    }
}

function Assert-Path {
    param([Parameter(Mandatory)][string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        throw "Required Task 85D preflight path is missing: $Path"
    }
}

function Invoke-Gate {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][scriptblock]$Command
    )

    Write-Host "--- $Name ---" -ForegroundColor Yellow
    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
Push-Location $repoRoot
try {
    @('git', 'go', 'gofmt', 'python', 'node', 'pwsh') | ForEach-Object {
        Assert-Command $_
    }

    $declaredFiles = @(
        'x/veid/types/evidence_trust_inventory.go',
        'x/veid/types/evidence_trust_inventory_test.go',
        'scripts/task85d-preflight.ps1',
        '_docs/ralph/handoffs/prototype-t1/HANDOFF.yaml'
    )
    $requiredPaths = @(
        'x/veid/types',
        'x/veid/keeper',
        'x/veidregistry',
        'pkg/inference',
        'cmd/inference-sidecar',
        'app',
        'tests/integration',
        '.github/tests/test_inference_deployment_policy.py',
        '.github/scripts/validate_inference_deployment_policy.py',
        'scripts/consensusdeterminism',
        'scripts/validate-agents-docs.mjs',
        'scripts/agent-preflight.ps1',
        'scripts/task85d-preflight.ps1'
    ) + $declaredFiles
    $requiredPaths | ForEach-Object { Assert-Path $_ }

    $goFiles = @(
        'x/veid/types/evidence_trust_inventory.go',
        'x/veid/types/evidence_trust_inventory_test.go'
    )

    Invoke-Gate 'gofmt check' {
        $unformatted = @(& gofmt -l -- $goFiles)
        if ($LASTEXITCODE -ne 0) {
            throw "gofmt check failed with exit code $LASTEXITCODE"
        }
        if ($unformatted.Count -gt 0) {
            throw "gofmt is required for:`n$($unformatted -join [Environment]::NewLine)"
        }
    }
    Invoke-Gate 'evidence trust inventory' { go test -count=1 ./x/veid/types -run EvidenceTrust }
    Invoke-Gate 'VEID types vet' { go vet ./x/veid/types }

    Invoke-Gate 'Task 85D package baseline' {
        go test -timeout=30m -count=1 ./x/veid/types ./x/veid/keeper ./x/veidregistry/... ./pkg/inference ./cmd/inference-sidecar
    }
    Invoke-Gate 'Task 85D race baseline' {
        go test -race -timeout=45m -count=1 ./x/veid/keeper ./pkg/inference ./cmd/inference-sidecar
    }
    Invoke-Gate 'application consensus and inference baseline' {
        go test -count=1 ./app -run 'VoteExtension|Proposal|Inference'
    }
    Invoke-Gate 'VEID and inference integration baseline' {
        go test -tags=e2e.integration -count=1 ./tests/integration/... -run 'VEID|Inference'
    }
    Invoke-Gate 'inference deployment policy tests' {
        python .github/tests/test_inference_deployment_policy.py
    }
    Invoke-Gate 'inference deployment policy validation' {
        python .github/scripts/validate_inference_deployment_policy.py
    }
    Invoke-Gate 'consensus determinism' { go run ./scripts/consensusdeterminism -root . }
    Invoke-Gate 'AGENTS documentation validation' { node scripts/validate-agents-docs.mjs }

    Invoke-Gate 'PowerShell syntax' {
        $tokens = $null
        $parseErrors = $null
        [void][System.Management.Automation.Language.Parser]::ParseFile(
            (Resolve-Path 'scripts/task85d-preflight.ps1'),
            [ref]$tokens,
            [ref]$parseErrors
        )
        if ($parseErrors.Count -gt 0) {
            throw "PowerShell syntax validation failed:`n$($parseErrors.Message -join [Environment]::NewLine)"
        }
    }
    Invoke-Gate 'declared file hygiene' {
        foreach ($path in $declaredFiles) {
            Assert-Path $path
            $resolvedPath = (Resolve-Path -LiteralPath $path).Path
            $content = [System.IO.File]::ReadAllText($resolvedPath)
            $unexpectedCarriageReturns = if ($path.EndsWith('.go')) {
                $content.Contains("`r")
            }
            else {
                $content.Replace("`r`n", '').Contains("`r")
            }
            if ($unexpectedCarriageReturns) {
                throw "Unexpected carriage returns in declared file: $path"
            }
            if ($content -match '(?m)[ \t]+$') {
                throw "Trailing whitespace is not allowed in declared file: $path"
            }
        }

        $unformatted = @(& gofmt -l -- $goFiles)
        if ($LASTEXITCODE -ne 0) {
            throw "declared file gofmt check failed with exit code $LASTEXITCODE"
        }
        if ($unformatted.Count -gt 0) {
            throw "gofmt is required for declared Go files:`n$($unformatted -join [Environment]::NewLine)"
        }
    }
    Invoke-Gate 'diff whitespace' { git diff --check }

    Invoke-Gate 'nested preflight recursion guard' {
        $currentPowerShellPath = (Get-Process -Id $PID).Path
        $previousSelfTestGuard = $env:VE_TASK85D_PREFLIGHT_ACTIVE
        try {
            $env:VE_TASK85D_PREFLIGHT_ACTIVE = '1'
            $selfTestOutput = @(
                & $currentPowerShellPath -NoProfile -File (Resolve-Path 'scripts/task85d-preflight.ps1').Path 2>&1
            )
            $selfTestExitCode = $LASTEXITCODE
        }
        finally {
            $env:VE_TASK85D_PREFLIGHT_ACTIVE = $previousSelfTestGuard
        }

        $selfTestText = $selfTestOutput -join [Environment]::NewLine
        if ($selfTestExitCode -ne 0) {
            throw "nested preflight recursion guard exited with code $selfTestExitCode"
        }
        if (-not $selfTestText.Contains($nestedPreflightSkipMessage)) {
            throw "nested preflight recursion guard did not emit expected text: $nestedPreflightSkipMessage"
        }
    }

    $previousGuard = $env:VE_TASK85D_PREFLIGHT_ACTIVE
    try {
        # Future agent-preflight implementations can inspect this process-scoped
        # guard and avoid invoking task85d-preflight recursively.
        $env:VE_TASK85D_PREFLIGHT_ACTIVE = '1'
        Invoke-Gate 'agent preflight' { pwsh scripts/agent-preflight.ps1 }
    }
    finally {
        $env:VE_TASK85D_PREFLIGHT_ACTIVE = $previousGuard
    }

    Write-Host 'Task 85D local preflight passed.' -ForegroundColor Green
}
finally {
    Pop-Location
}