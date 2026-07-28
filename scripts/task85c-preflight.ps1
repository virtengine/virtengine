#!/usr/bin/env pwsh
# Task 85C deterministic local engineering gate.
[CmdletBinding()]
param(
    [switch]$SkipRace
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    function Invoke-Gate {
        param([string]$Name, [scriptblock]$Command)
        Write-Host "--- $Name ---" -ForegroundColor Yellow
        & $Command
        if ($LASTEXITCODE -ne 0) {
            throw "$Name failed with exit code $LASTEXITCODE"
        }
    }

    $goFiles = @(
        'pkg/provider_daemon/key_manager.go',
        'pkg/provider_daemon/key_manager_test.go',
        'pkg/provider_daemon/provider_mutation.go',
        'pkg/provider_daemon/provider_mutation_store.go',
        'pkg/provider_daemon/provider_mutation_test.go',
        'pkg/provider_daemon/chain_submitter_queue.go',
        'pkg/provider_daemon/chain_client.go',
        'pkg/provider_daemon/chain_client_test.go',
        'pkg/provider_daemon/submitter_lease_file.go',
        'pkg/provider_daemon/submitter_lease_file_test.go',
        'pkg/provider_daemon/submitter_lease_kubernetes.go',
        'pkg/provider_daemon/submitter_lease_kubernetes_test.go',
        'pkg/provider_daemon/portal_api.go',
        'pkg/provider_daemon/portal_readiness_test.go',
        'cmd/provider-daemon/main.go',
        'cmd/provider-daemon/key_backup_test.go',
        'x/settlement/keeper/fiat_conversion_cross_process_test.go'
    ) | Where-Object { Test-Path $_ }

    Invoke-Gate 'gofmt' {
        gofmt -w $goFiles
        $unformatted = @(gofmt -l $goFiles)
        if ($unformatted.Count -gt 0) {
            $unformatted | Write-Error
            exit 1
        }
    }
    Invoke-Gate 'provider tests' { go test ./pkg/provider_daemon ./cmd/provider-daemon -count=1 }
    Invoke-Gate 'settlement continuity test' { go test ./x/settlement/keeper -run 'FiatObservationDurableMutationToAuthenticatedMsgServerProgression' -count=1 }
    Invoke-Gate 'go vet' { go vet ./pkg/provider_daemon ./cmd/provider-daemon }

    $lint = Join-Path $HOME 'go/bin/golangci-lint.exe'
    if (-not (Test-Path $lint)) {
        $lintCommand = Get-Command golangci-lint -ErrorAction SilentlyContinue
        if ($null -eq $lintCommand) { throw 'golangci-lint is required for Task 85C preflight' }
        $lint = $lintCommand.Source
    }
    Invoke-Gate 'golangci-lint' { & $lint run ./pkg/provider_daemon/... ./cmd/provider-daemon/... }
    Invoke-Gate 'provider build' { go build ./cmd/provider-daemon }
    Invoke-Gate 'canonical Kubernetes render and policy' { node scripts/task85c-validate-kubernetes.mjs }
    Invoke-Gate 'AGENTS documentation validation' { node scripts/validate-agents-docs.mjs }
    Invoke-Gate 'provider backup/restore continuity drill' {
        wsl.exe --exec bash -lc 'cd /mnt/d/source/repos/virtengine-gh/virtengine && bash -n scripts/dr/backup-provider-state.sh scripts/ci/backup-restore-smoke-test.sh && bash scripts/ci/backup-restore-smoke-test.sh'
    }
    Invoke-Gate 'PowerShell syntax' {
        $tokens = $null
        $errors = $null
        [void][System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path 'scripts/task85c-preflight.ps1'), [ref]$tokens, [ref]$errors)
        if ($errors.Count -gt 0) {
            $errors | ForEach-Object { Write-Error $_.Message }
            exit 1
        }
    }
    Invoke-Gate 'diff whitespace' { git diff --check -- pkg/provider_daemon cmd/provider-daemon deploy/kubernetes infra/kubernetes scripts _docs }

    if ($SkipRace) {
        Write-Warning 'Task 85C race test explicitly skipped; this run is not full local acceptance evidence.'
    } else {
        $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
        if ($null -eq $wsl) { throw 'WSL is required for the Task 85C race gate' }
        Invoke-Gate 'WSL race tests' {
            wsl.exe --exec bash -lc 'cd /mnt/d/source/repos/virtengine-gh/virtengine && CGO_ENABLED=1 GOMAXPROCS=2 go test -race -p=1 ./pkg/provider_daemon ./cmd/provider-daemon -run "KeyManager|SubmitterLease|StandbyTakesOver|MutationGuardFencesStandby|FileProviderMutationStore|PortalReadiness|ProviderKeyPassphrase|ProviderMutationSubmitterRestart" -count=1'
        }
    }

    Write-Host 'Task 85C local preflight passed.' -ForegroundColor Green
}
finally {
    Pop-Location
}
