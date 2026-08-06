param(
    [switch]$SkipRace,
    [switch]$SkipLint,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"

function Run-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Command
    )

    Write-Host "==> $Name"
    $global:LASTEXITCODE = 0
    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

Run-Step "provider mutation focused tests" {
    go test ./pkg/provider_daemon -run ProviderMutation -count=1
}

Run-Step "provider daemon package tests" {
    go test ./pkg/provider_daemon -count=1
}

if (-not $SkipRace) {
    Run-Step "provider mutation race tests" {
        $previousCGO = $env:CGO_ENABLED
        try {
            $env:CGO_ENABLED = "1"
            go test -race ./pkg/provider_daemon -run ProviderMutation -count=1
            if ($LASTEXITCODE -eq 0) {
                return
            }

            Write-Host "Native race test unavailable; trying WSL race test."
            if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
                throw "native race failed and WSL is unavailable"
            }
            wsl bash -lc 'cd /mnt/d/source/repos/virtengine-gh/virtengine && go test -race ./pkg/provider_daemon -run ProviderMutation -count=1'
        } finally {
            $env:CGO_ENABLED = $previousCGO
        }
    }
}

Run-Step "provider daemon vet" {
    go vet ./pkg/provider_daemon
}

if (-not $SkipLint) {
    Run-Step "provider daemon lint" {
        golangci-lint run ./pkg/provider_daemon --timeout 10m
    }
}

if (-not $SkipBuild) {
    Run-Step "provider daemon command build" {
        go build ./cmd/provider-daemon
    }

    Run-Step "virtengine command build" {
        go build ./cmd/virtengine
    }
}

Run-Step "AGENTS docs validator" {
    node scripts/validate-agents-docs.mjs
}

Run-Step "direct generated MsgClient scan" {
    $matches = rg -n 'NewMsgClient|\.MsgClient\(' pkg/provider_daemon cmd/provider-daemon -g "*.go"
    if ($LASTEXITCODE -eq 0) {
        throw "direct generated MsgClient mutation client references remain:`n$matches"
    }
    if ($LASTEXITCODE -ne 1) {
        throw "MsgClient scan failed with exit code $LASTEXITCODE"
    }
    $global:LASTEXITCODE = 0
}

Write-Host "Task 85A preflight passed."
