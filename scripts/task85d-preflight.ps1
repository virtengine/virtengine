#!/usr/bin/env pwsh
[CmdletBinding()]
param(
    [switch]$SkipRace,
    [switch]$SkipLint,
    [switch]$SkipGeneration,
    [switch]$SkipExpensive
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-DiagnosticSkip {
    param([string]$Message)
    Write-Host "DIAGNOSTIC SKIP: $Message" -ForegroundColor Yellow
    Write-Host "This output is NOT release evidence for Task 85D." -ForegroundColor Yellow
}

function Invoke-Task85DStep {
    param(
        [string]$Name,
        [scriptblock]$Script
    )
    Write-Host "--- $Name ---" -ForegroundColor Cyan
    $global:LASTEXITCODE = 0
    try {
        & $Script
    }
    catch {
        throw "$Name failed: $($_.Exception.Message)"
    }
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

function Assert-Command {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name is required for full Task 85D preflight"
    }
}

function ConvertTo-RepoPath {
    param([string]$Path)
    return ($Path -replace '\\', '/')
}

function Test-GeneratedContractPath {
    param([string]$Path)
    $p = ConvertTo-RepoPath $Path
    $patterns = @(
        '^proto/virtengine/veid/v1/.*\.proto$',
        '^sdk/proto/node/virtengine/veid/v1/.*\.proto$',
        '^sdk/go/node/veid/v1/.*\.pb\.go$',
        '^sdk/go/node/veid/v1/.*\.pb\.gw\.go$',
        '^sdk/ts/src/generated/protos/virtengine/veid/v1/',
        '^sdk/artifacts/proto/.*veid',
        '^artifacts/proto/.*veid',
        '^api/openapi/virtengine-proto\.swagger\.json$',
        '^api/openapi/.*veid.*\.json$',
        '^buf\.(yaml|lock)$',
        '^proto/buf\.(yaml|lock)$'
    )
    foreach ($pattern in $patterns) {
        if ($p -match $pattern) { return $true }
    }
    return $false
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $repoRoot

$env:GOCACHE = Join-Path $repoRoot ".cache\go-build"
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
$env:GOLANGCI_LINT_CACHE = Join-Path $repoRoot ".cache\golangci-lint"
New-Item -ItemType Directory -Force -Path $env:GOLANGCI_LINT_CACHE | Out-Null

if ($SkipRace) { Write-DiagnosticSkip "Task 85D WSL race gate was skipped." }
if ($SkipLint) { Write-DiagnosticSkip "Task 85D golangci-lint gate was skipped." }
if ($SkipGeneration) { Write-DiagnosticSkip "Task 85D generated contract drift gate was skipped." }
if ($SkipExpensive) { Write-DiagnosticSkip "Task 85D expensive gates were skipped: full VEID integration package, full VEID/inference package suite, and WSL race." }

Assert-Command go
Assert-Command python
Assert-Command node

$taskGoFiles = @(
    "tests/integration/veid/task85d_process_boundary_test.go",
    "tests/integration/veid/veid_flow_test.go",
    "x/veid/keeper/task85d_consensus_conformance_test.go",
    "x/veid/keeper/web_scope_scoring.go"
) | Where-Object { Test-Path -LiteralPath $_ }

$taskScopePaths = @(
    "tests/integration/veid/task85d_process_boundary_test.go",
    "tests/integration/veid/veid_flow_test.go",
    "x/veid/keeper/task85d_consensus_conformance_test.go",
    "x/veid/keeper/web_scope_scoring.go",
    "scripts/task85d-preflight.ps1",
    "scripts/agent-preflight.ps1",
    "scripts/AGENTS.md",
    ".github/AGENT_PREFLIGHT.md",
    "_docs/INDEX.md",
    "_docs/ralph/progress.md",
    "_docs/audits/task-85d-process-boundary-conformance-report-2026-07-28.md",
    "_docs/task-85d-external-prerequisite-certification-ledger.md"
) | Where-Object { Test-Path -LiteralPath $_ }

Invoke-Task85DStep "PowerShell parse check" {
    foreach ($script in @("scripts/task85d-preflight.ps1", "scripts/agent-preflight.ps1")) {
        $null = [scriptblock]::Create((Get-Content -Raw -LiteralPath $script))
    }
}

Invoke-Task85DStep "gofmt check" {
    if ($taskGoFiles.Count -eq 0) {
        throw "no Task 85D Go files were found for gofmt check"
    }
    $unformatted = @(& gofmt -l $taskGoFiles)
    if ($unformatted.Count -gt 0) {
        $unformatted | ForEach-Object { Write-Host "unformatted: $_" -ForegroundColor Red }
        exit 1
    }
}

Invoke-Task85DStep "Task 85D integration process-boundary tests" {
    & go test ./tests/integration/veid -run "TestTask85D" -count=1
}

Invoke-Task85DStep "receipt canonical tests" {
    & go test ./x/veid/types -run "Test(InferenceReceipt|CanonicalInference|DecodeCanonicalSignedInferenceReceipt)" -count=1
}

Invoke-Task85DStep "receipt finalization, consensus, and production stub rejection tests" {
    & go test ./x/veid/keeper -run "Test(Task85D|EnvironmentCannotSelectStubButExplicitInjectionWorks|SubmitConsensusVerificationFinalizesOnlyQuorumReceiptBytes|FinalizeCarriedConsensusReceiptValidationMatrix|SubmitConsensusVerificationPrevalidatesReceiptsAtomically)" -count=1
}

Invoke-Task85DStep "inference sidecar mTLS and fallback policy tests" {
    & go test ./cmd/inference-sidecar ./pkg/inference -run "Test.*(StartupOptionsFailClosedDefaultsAndFallbackPolicy|NewTensorFlowScorerRequiresExplicitStubOptIn|ModelRunFailsClosedWithoutStubOptIn|SidecarClientRequiresExplicitStubOptIn|TensorFlowRuntimeStubRequiresExplicitOptIn|MTLS|Mtls|mTLS|Stub|Fallback|Policy|Determin|Receipt)" -count=1
}

if ($SkipExpensive) {
    Write-DiagnosticSkip "full VEID integration package and full VEID/inference package suite were not run."
}
else {
    Invoke-Task85DStep "full VEID integration package" {
        & go test ./tests/integration/veid -count=1
    }
    Invoke-Task85DStep "full VEID and inference package suite" {
        & go test ./x/veid/types ./x/veid/keeper ./cmd/inference-sidecar ./pkg/inference -count=1
    }
}

Invoke-Task85DStep "deployment policy validator tests" {
    & python .github/tests/test_inference_deployment_policy.py
}

Invoke-Task85DStep "deployment policy validator" {
    & python .github/scripts/validate_inference_deployment_policy.py
}

Invoke-Task85DStep "AGENTS docs validator" {
    & node scripts/validate-agents-docs.mjs
}

Invoke-Task85DStep "focused go vet" {
    & go vet ./tests/integration/veid ./x/veid/types ./x/veid/keeper ./cmd/inference-sidecar ./pkg/inference
}

if ($SkipLint) {
    Write-DiagnosticSkip "golangci-lint was not run."
}
else {
    Invoke-Task85DStep "task-scoped golangci-lint" {
        Assert-Command golangci-lint
        & golangci-lint run --new-from-rev=HEAD --timeout=10m ./tests/integration/veid ./x/veid/types ./x/veid/keeper ./cmd/inference-sidecar ./pkg/inference
    }
}

Invoke-Task85DStep "task-scoped diff whitespace check" {
    & git diff --check -- $taskScopePaths
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    & git diff --cached --check -- $taskScopePaths
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    foreach ($path in $taskScopePaths) {
        $lineNo = 0
        foreach ($line in Get-Content -LiteralPath $path) {
            $lineNo++
            if ($line -match '\s+$') {
                Write-Host "${path}:${lineNo}: trailing whitespace" -ForegroundColor Red
                exit 1
            }
            if ($line -match '^(<<<<<<<|=======|>>>>>>>)') {
                Write-Host "${path}:${lineNo}: conflict marker" -ForegroundColor Red
                exit 1
            }
        }
    }
}

$changedFiles = @(
    & git diff --name-only 2>$null
    & git diff --cached --name-only 2>$null
    & git ls-files --others --exclude-standard 2>$null
) | Where-Object { $_ } | Sort-Object -Unique

$generatedContractTouched = @($changedFiles | Where-Object { Test-GeneratedContractPath $_ })
if ($generatedContractTouched.Count -gt 0) {
    if ($SkipGeneration) {
        Write-DiagnosticSkip "generated VEID contract drift was applicable but skipped for: $($generatedContractTouched -join ', ')"
    }
    else {
        throw "Generated VEID contract paths changed: $($generatedContractTouched -join ', '). This layered checkout is dirty; do not run broad generation in place. Verify generation in a clean isolated worktree/container and compare artifacts. -SkipGeneration is diagnostic only."
    }
}
else {
    Write-Host "generated contract drift check: not applicable to current Task 85D files" -ForegroundColor DarkGray
}

if ($SkipRace -or $SkipExpensive) {
    Write-DiagnosticSkip "WSL race test was not run."
}
else {
    Invoke-Task85DStep "WSL race focused Task 85D tests" {
        $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
        if (-not $wsl) {
            throw "wsl.exe is required for full Task 85D race preflight"
        }
        $wslStatus = @(& wsl.exe --status 2>&1)
        if ($LASTEXITCODE -ne 0) {
            $wslStatus | ForEach-Object { Write-Host $_ }
            exit $LASTEXITCODE
        }
        $wslList = @(& wsl.exe --list --verbose 2>&1)
        if ($LASTEXITCODE -ne 0 -or ($wslList -join "`n") -match 'no installed distributions') {
            $wslList | ForEach-Object { Write-Host $_ }
            throw "WSL is installed but no Linux distribution is available; full Task 85D race preflight fails closed. Use -SkipRace only for diagnostic non-release evidence."
        }
        if ($repoRoot -notmatch '^([A-Za-z]):(\\.*)$') {
            throw "Task 85D WSL race preflight requires a local Windows drive path, got $repoRoot"
        }
        $drive = $Matches[1].ToLowerInvariant()
        $relativePath = $Matches[2].Replace('\', '/')
        $linuxPath = "/mnt/$drive$relativePath"
        & wsl.exe --cd $linuxPath bash -lc 'mkdir -p .cache/go-build-wsl && GOCACHE="$(pwd)/.cache/go-build-wsl" GOWORK=off GOFLAGS=-mod=mod CGO_ENABLED=1 go test -race ./tests/integration/veid ./x/veid/keeper -run "TestTask85D" -count=1'
    }
}

Write-Host "Task 85D preflight passed." -ForegroundColor Green
