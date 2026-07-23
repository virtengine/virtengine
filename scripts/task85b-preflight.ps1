#!/usr/bin/env pwsh
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
# Task 85B focused preflight for authenticated DEX routing and fiat off-ramp work.
# Usage: pwsh scripts/task85b-preflight.ps1 [-Quick] [-SkipRace]
[CmdletBinding()]
param(
    [switch]$Quick,
    [switch]$SkipRace
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$PSNativeCommandUseErrorActionPreference = $false

$repo = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$taskPaths = @(
    'pkg/dex',
    'pkg/payments/offramp',
    'pkg/provider_daemon',
    'cmd/provider-daemon/main.go',
    'cmd/provider-daemon/fiat_conversion_startup_test.go',
    'app/app.go',
    'app/mac.go',
    'app/mac_task85b_test.go',
    'x/settlement',
    'upgrades/software/v1.8.0',
    'upgrades/types/types.go',
    'upgrades/upgrades.go',
    'upgrades/upgrades_test.go',
    'tests/compatibility/task85b_wire_compatibility_test.go',
    'tests/integration/settlement',
    'tests/upgrade',
    'sdk/proto/node/virtengine/settlement/v1',
    'sdk/go/node/settlement/v1',
    'sdk/ts/src/generated/protos/virtengine/settlement/v1',
    'sdk/artifacts/proto',
    'api/openapi/virtengine-proto.swagger.json',
    '_docs/audits/task-85b-completion-evidence-2026-07-23.md',
    '_docs/protocols/fiat-conversion-orchestrator-protocol.md',
    '_docs/protocols/task-85b-dex-payout-support-matrices.md',
    '_docs/runbooks/fiat-conversion-incident-recovery.md',
    '_docs/task-85b-external-prerequisite-certification-ledger.md',
    'scripts/task85b-preflight.ps1'
)
$goPackages = @(
    './pkg/dex',
    './pkg/payments/offramp',
    './pkg/provider_daemon',
    './cmd/provider-daemon',
    './x/settlement',
    './x/settlement/keeper',
    './x/settlement/types',
    './x/settlement/ibc',
    './upgrades/software/v1.8.0',
    './tests/compatibility',
    './tests/upgrade'
)
$vetPackages = @(
    './pkg/dex',
    './pkg/payments/offramp',
    './pkg/provider_daemon',
    './cmd/provider-daemon',
    './x/settlement',
    './x/settlement/keeper',
    './x/settlement/types',
    './x/settlement/ibc',
    './upgrades/software/v1.8.0',
    './tests/compatibility',
    './tests/upgrade'
)
$lintPackages = @(
    './pkg/dex/...',
    './pkg/payments/offramp/...',
    './pkg/provider_daemon',
    './cmd/provider-daemon',
    './x/settlement',
    './x/settlement/keeper',
    './x/settlement/types',
    './x/settlement/ibc',
    './upgrades/software/v1.8.0',
    './tests/compatibility',
    './tests/upgrade'
)

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Action
    )

    Write-Host "[85B] $Name" -ForegroundColor Cyan
    $global:LASTEXITCODE = 0
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

function Get-RequiredContent {
    param([Parameter(Mandatory = $true)][string]$Path)

    $fullPath = Join-Path $repo $Path
    if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) {
        throw "required Task 85B file is missing: $Path"
    }
    return Get-Content -LiteralPath $fullPath -Raw
}

function Assert-Contains {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Pattern,
        [Parameter(Mandatory = $true)][string]$Claim
    )

    $content = Get-RequiredContent -Path $Path
    if ($content -notmatch $Pattern) {
        throw "$Claim is not asserted by $Path"
    }
}

function Assert-SHA256Sidecar {
    param(
        [Parameter(Mandatory = $true)][string]$Artifact,
        [Parameter(Mandatory = $true)][string]$Sidecar
    )

    $artifactPath = Join-Path $repo $Artifact
    $sidecarContent = (Get-RequiredContent -Path $Sidecar).Trim()
    $expected = ($sidecarContent -split '\s+', 2)[0]
    if ($expected -notmatch '^[0-9a-fA-F]{64}$') {
        throw "invalid SHA-256 sidecar format: $Sidecar"
    }
    $actual = (Get-FileHash -LiteralPath $artifactPath -Algorithm SHA256).Hash
    if (-not $actual.Equals($expected, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "SHA-256 mismatch for $Artifact"
    }
}

function Get-Task85BDocuments {
    $documents = Get-ChildItem -LiteralPath (Join-Path $repo '_docs') -Recurse -File -Filter '*.md'
    return @($documents | Where-Object {
        $relative = [System.IO.Path]::GetRelativePath($repo, $_.FullName).Replace('\', '/')
        $content = Get-Content -LiteralPath $_.FullName -Raw
        $relative -match '(?i)(task[-_]?85b|dex|fiat|off.?ramp|payout|corridor)' -or
        $content -match '(?i)Task\s*85B|85B.*(?:DEX|fiat|off.?ramp)|(?:DEX|fiat|off.?ramp).*85B'
    })
}

function Find-Document {
    param(
        [Parameter(Mandatory = $true)][System.IO.FileInfo[]]$Documents,
        [Parameter(Mandatory = $true)][string[]]$Patterns,
        [Parameter(Mandatory = $true)][string]$Description
    )

    foreach ($document in $Documents) {
        $content = Get-Content -LiteralPath $document.FullName -Raw
        $matches = $true
        foreach ($pattern in $Patterns) {
            if ($content -notmatch $pattern) {
                $matches = $false
                break
            }
        }
        if ($matches) {
            return $document
        }
    }
    throw "Task 85B $Description document was not found under _docs"
}

function Convert-ToWSLPath {
    param([Parameter(Mandatory = $true)][string]$WindowsPath)

    $translated = & wsl.exe -e wslpath -a -u $WindowsPath 2>&1
    if ($LASTEXITCODE -ne 0 -or -not $translated) {
        throw "unable to translate repository path for WSL: $translated"
    }
    return ($translated | Select-Object -First 1).Trim()
}

Push-Location $repo
try {
    Invoke-Step 'checking required tools' {
        foreach ($command in @('git', 'go')) {
            if (-not (Get-Command $command -ErrorAction SilentlyContinue)) {
                throw "required command is unavailable: $command"
            }
        }
        $global:LASTEXITCODE = 0
    }

    Invoke-Step 'checking Task 85B conflict markers and whitespace' {
        $taskTextFiles = foreach ($path in $taskPaths) {
            if (Test-Path -LiteralPath $path -PathType Container) {
                Get-ChildItem -LiteralPath $path -Recurse -File
            }
            elseif (Test-Path -LiteralPath $path -PathType Leaf) {
                Get-Item -LiteralPath $path
            }
        }
        $taskTextFiles = @($taskTextFiles | Where-Object {
            $_.Extension -in @('.go', '.proto', '.md', '.json', '.yaml', '.yml', '.ps1', '.toml', '.txt')
        } | Sort-Object FullName -Unique)
        $markers = $taskTextFiles | Select-String -Pattern '^(<<<<<<< |=======$|>>>>>>> |\|\|\|\|\|\|\| )'
        if ($markers) {
            throw "Task 85B conflict markers remain:`n$($markers -join "`n")"
        }

        & git diff --check -- @taskPaths
        if ($LASTEXITCODE -ne 0) {
            throw 'git diff --check failed for Task 85B-owned paths'
        }
    }

    Invoke-Step 'checking lifecycle, matrices, runbook, and prerequisite ledger' {
        $documents = Get-Task85BDocuments
        if ($documents.Count -eq 0) {
            throw 'no checked-in Task 85B documentation was discovered under _docs'
        }

        $lifecycle = Find-Document -Documents $documents -Description 'lifecycle/status' -Patterns @(
            '(?i)Task\s*85B',
            '(?i)(completion\s+and\s+evidence\s+report|Task\s+85B\s+completion)',
            '(?i)(lifecycle|status)',
            'engineering[_-]complete[_-]external[_-]blocked'
        )
        $dexMatrix = Find-Document -Documents $documents -Description 'DEX support matrix' -Patterns @(
            '(?i)(DEX\s+route\s+support\s+matrix|DEX\s+support\s+matrix)',
            '(?i)DEX',
            '(?i)(support\s+matrix|route\s+matrix|matrix)',
            'engineering[_-]complete[_-]external[_-]blocked'
        )
        $payoutMatrix = Find-Document -Documents $documents -Description 'payout support matrix' -Patterns @(
            '(?i)(payout\s+corridor\s+support\s+matrix|payout\s+support\s+matrix)',
            '(?i)(payout|off.?ramp|corridor)',
            '(?i)(support\s+matrix|corridor\s+matrix|matrix)',
            'engineering[_-]complete[_-]external[_-]blocked'
        )
        $runbook = Find-Document -Documents $documents -Description 'operator runbook' -Patterns @(
            '(?i)(runbook|incident\s+response|operator\s+procedure)',
            '(?i)(DEX|fiat|off.?ramp|payout|corridor)'
        )
        $prerequisites = Find-Document -Documents $documents -Description 'external-prerequisite ledger' -Patterns @(
            '(?i)(external\s+prerequisite|prerequisite\s+and\s+certification\s+ledger)',
            '(?i)prerequisite',
            '(?i)(ledger|owner|evidence|blocker)',
            'engineering[_-]complete[_-]external[_-]blocked'
        )

        if ($dexMatrix.FullName -eq $payoutMatrix.FullName) {
            $matrixContent = Get-Content -LiteralPath $dexMatrix.FullName -Raw
            if ($matrixContent -notmatch '(?i)##\s+DEX\s+route\s+support\s+matrix' -or
                $matrixContent -notmatch '(?i)##\s+Payout\s+corridor\s+support\s+matrix') {
                throw 'combined Task 85B matrix artifact must contain separate DEX route and payout corridor support matrix sections'
            }
        }

        $claimDocuments = @($lifecycle, $dexMatrix, $payoutMatrix, $runbook, $prerequisites) |
            Sort-Object FullName -Unique
        foreach ($document in $claimDocuments) {
            $content = Get-Content -LiteralPath $document.FullName -Raw
            $relative = [System.IO.Path]::GetRelativePath($repo, $document.FullName).Replace('\', '/')
            if ($relative -match '(?i)(task[-_]?85b|matrix|lifecycle|status|completion|evidence|ledger)' -and
                $relative -ne '_docs/protocol-completion-continuation-plan.md') {
                $prohibitedCertifiedLines = @($content -split "`r?`n" | Where-Object {
                    $_ -match '(?i)certified_enabled' -and
                    $_ -notmatch '(?i)(no\s+(?:row|route|corridor|profile|state|status)|not\s+(?:claim|production|enabled|certified)|must\s+not|cannot|never|none|without|only\s+after|remains?\s+(?:blocked|unavailable)|accepted\s+.*states?.*only)'
                })
                if ($prohibitedCertifiedLines.Count -gt 0) {
                    throw "checked-in Task 85B documentation claims certified_enabled: $relative`n$($prohibitedCertifiedLines -join "`n")"
                }
            }
        }

        $global:LASTEXITCODE = 0
    }

    Invoke-Step 'checking production fail-closed architecture' {
        Assert-Contains -Path 'pkg/dex/service.go' -Pattern 'test-only DEX adapter cannot be registered in a runtime service' -Claim 'placeholder DEX production rejection'
        Assert-Contains -Path 'pkg/dex/service.go' -Pattern 'DEX adapter is not production-authorized' -Claim 'production DEX authorization boundary'
        Assert-Contains -Path 'pkg/dex/security_test.go' -Pattern 'func TestRuntimeServiceRejectsManualPlaceholderRegistration' -Claim 'placeholder DEX rejection test'
        Assert-Contains -Path 'pkg/dex/osmosis_adapter_test.go' -Pattern 'RejectsSyntheticEstimate' -Claim 'synthetic DEX estimate rejection test'
        Assert-Contains -Path 'pkg/payments/offramp/mock_provider.go' -Pattern 'func \(p \*MockProvider\) IsTestOnly\(\) bool \{ return true \}' -Claim 'test-only payout provider marker'
        Assert-Contains -Path 'pkg/payments/offramp/bridge.go' -Pattern 'if testAdapter, ok := adapter\.\(TestOnlyAdapter\); ok && testAdapter\.IsTestOnly\(\)' -Claim 'test-only payout production rejection'
        Assert-Contains -Path 'pkg/payments/offramp/profile_bridge_test.go' -Pattern 'func TestProductionBridgeRejectsMockAndExternalBlockedProfiles' -Claim 'production payout rejection test'
        Assert-Contains -Path 'cmd/provider-daemon/main.go' -Pattern 'FlagFiatConversionEnabled\s*=\s*"fiat-conversion-enabled"' -Claim 'fiat conversion startup flag'
        Assert-Contains -Path 'cmd/provider-daemon/main.go' -Pattern 'PersistentFlags\(\)\.Bool\(FlagFiatConversionEnabled, false' -Claim 'disabled-by-default production orchestrator'
        Assert-Contains -Path 'cmd/provider-daemon/main.go' -Pattern 'func validateFiatConversionStartup\(' -Claim 'fail-closed production startup validator'
        Assert-Contains -Path 'cmd/provider-daemon/fiat_conversion_startup_test.go' -Pattern 'func TestValidateFiatConversionStartupProductionRejectsBackendIdentifiers' -Claim 'production orchestrator startup rejection test'

        $settlementGoFiles = Get-ChildItem -LiteralPath (Join-Path $repo 'x/settlement/keeper') -File -Filter '*.go' |
            Where-Object { $_.Name -notlike '*_test.go' }
        foreach ($file in $settlementGoFiles) {
            $content = Get-Content -LiteralPath $file.FullName -Raw
            if ($content -match '(?m)^\s*"(?:net|net/http|net/url|github\.com/gorilla/websocket|google\.golang\.org/grpc|google\.golang\.org/grpc/(?:clientconn|credentials|resolver)(?:/[^\"]*)?)"\s*$') {
                throw "settlement consensus keeper imports a network transport: $($file.Name)"
            }
        }
        Assert-Contains -Path 'x/settlement/keeper/keeper.go' -Pattern 'func \(k \*Keeper\) SetDexSwapExecutor\(' -Claim 'settlement DEX adapter discard boundary'
        Assert-Contains -Path 'x/settlement/keeper/keeper.go' -Pattern 'func \(k \*Keeper\) SetOffRampBridge\(' -Claim 'settlement payout adapter discard boundary'
        Assert-Contains -Path 'x/settlement/keeper/dex.go' -Pattern 'func ensureNoConsensusExternalIO\(' -Claim 'settlement external I/O rejection'
        Assert-Contains -Path 'x/settlement/keeper/consensus_boundary_test.go' -Pattern 'func TestKeeperDiscardsExternalConsensusAdapters' -Claim 'settlement external I/O boundary test'
        Assert-Contains -Path 'tests/integration/settlement/dex_offramp_test.go' -Pattern 'func TestAuthenticatedFiatConversionObservationPipeline' -Claim 'tagged authenticated fiat protocol integration test'
        Assert-Contains -Path 'tests/integration/settlement/dex_offramp_test.go' -Pattern 'RecordFiatConversionObservation' -Claim 'six-stage observation integration coverage'
        Assert-Contains -Path 'tests/integration/settlement/dex_offramp_test.go' -Pattern 'ErrExternalIOForbidden' -Claim 'tagged consensus external-I/O rejection coverage'
        $global:LASTEXITCODE = 0
    }

    Invoke-Step 'checking protocol and upgrade registration' {
        Assert-Contains -Path 'sdk/proto/node/virtengine/settlement/v1/tx.proto' -Pattern 'rpc RecordFiatConversionObservation\(MsgRecordFiatConversionObservation\)' -Claim 'MsgRecordFiatConversionObservation Msg service registration'
        Assert-Contains -Path 'sdk/proto/node/virtengine/settlement/v1/tx.proto' -Pattern 'message MsgRecordFiatConversionObservation\s*\{' -Claim 'MsgRecordFiatConversionObservation declaration'
        Assert-Contains -Path 'x/settlement/types/codec.go' -Pattern '&MsgRecordFiatConversionObservation\{\}' -Claim 'MsgRecordFiatConversionObservation interface registration'
        Assert-Contains -Path 'pkg/provider_daemon/provider_mutation.go' -Pattern 'MutationSettlementFiatObservation ProviderMutationKind = "settlement\.record_fiat_conversion_observation"' -Claim 'fiat observation mutation kind'
        Assert-Contains -Path 'pkg/provider_daemon/provider_mutation.go' -Pattern 'MutationSettlementFiatObservation, &settlementv1\.MsgRecordFiatConversionObservation' -Claim 'fiat observation durable mutation registry'
        Assert-Contains -Path 'pkg/provider_daemon/provider_mutation.go' -Pattern 'func \(s \*ProviderMutationSubmitter\) SubmitFiatConversionObservation\(' -Claim 'fiat observation durable submitter'
        Assert-Contains -Path 'pkg/provider_daemon/provider_mutation_chain.go' -Pattern 'case \*settlementv1\.MsgRecordFiatConversionObservation:' -Claim 'fiat observation logical reconciler routing'
        Assert-Contains -Path 'pkg/provider_daemon/provider_mutation_fiat_test.go' -Pattern 'func TestProviderMutationFiatObservationLogicalReconciliation' -Claim 'fiat observation logical reconciliation test'
        Assert-Contains -Path 'app/mac.go' -Pattern 'settlementtypes\.FiatConversionCustodyAccountName' -Claim 'fiat custody module account registration'
        Assert-Contains -Path 'app/mac_task85b_test.go' -Pattern 'func TestFiatConversionCustodyModuleAccountIsRegisteredAsInternalOnlySink' -Claim 'fiat custody blocked-account test'
        Assert-Contains -Path 'x/settlement/module.go' -Pattern 'func \(AppModule\) ConsensusVersion\(\) uint64 \{ return 4 \}' -Claim 'settlement consensus version 4'
        Assert-Contains -Path 'upgrades/types/types.go' -Pattern 'const AuthenticatedFiatConversionsUpgradeName = "v1\.8\.0"' -Claim 'v1.8.0 upgrade name'
        Assert-Contains -Path 'upgrades/software/v1.8.0/init.go' -Pattern 'RegisterUpgrade\(UpgradeName, initUpgrade\)' -Claim 'v1.8.0 upgrade registration'
        Assert-Contains -Path 'upgrades/upgrades.go' -Pattern 'upgrades/software/v1\.8\.0' -Claim 'v1.8.0 application registration import'
        Assert-Contains -Path 'tests/upgrade/registry_test.go' -Pattern 'v180\s+"github\.com/virtengine/virtengine/upgrades/software/v1\.8\.0"' -Claim 'v1.8.0 upgrade registry coverage'
        $global:LASTEXITCODE = 0
    }

    Invoke-Step 'validating protobuf inventory and artifact hashes' {
        if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
            throw 'node is required for the protobuf inventory gate'
        }
        & node --test scripts/protoinventory/inventory.test.mjs
        if ($LASTEXITCODE -ne 0) { return }

        $tempInventory = Join-Path ([System.IO.Path]::GetTempPath()) "virtengine-task85b-inventory-$PID.json"
        try {
            & node scripts/protoinventory/inventory.mjs $tempInventory
            if ($LASTEXITCODE -ne 0) { return }
            $checkedInventory = Join-Path $repo 'sdk/artifacts/proto/inventory.json'
            if (-not (Compare-Object (Get-Content -LiteralPath $checkedInventory) (Get-Content -LiteralPath $tempInventory) -SyncWindow 0)) {
                $global:LASTEXITCODE = 0
            }
            else {
                throw 'sdk/artifacts/proto/inventory.json is stale'
            }
            Assert-SHA256Sidecar -Artifact 'sdk/artifacts/proto/inventory.json' -Sidecar 'sdk/artifacts/proto/inventory.json.sha256'
            Assert-SHA256Sidecar -Artifact 'sdk/artifacts/proto/virtengine.binpb' -Sidecar 'sdk/artifacts/proto/virtengine.binpb.sha256'
        }
        finally {
            Remove-Item -LiteralPath $tempInventory, "$tempInventory.sha256" -Force -ErrorAction SilentlyContinue
        }
        $global:LASTEXITCODE = 0
    }

    Invoke-Step 'running tagged settlement integration tests' {
        & go test -tags='e2e.integration' ./tests/integration/settlement/... -count=1
    }

    Invoke-Step 'running targeted Go tests' {
        & go test @goPackages -count=1
    }

    Invoke-Step 'running focused settlement package-tree tests' {
        & go test ./x/settlement/... -run 'FiatConversion|ConsensusExternal' -count=1
    }

    Invoke-Step 'running relevant go vet' {
        & go vet @vetPackages
    }

    if (-not $Quick) {
        Invoke-Step 'verifying generated contract drift' {
            $dockerReady = $false
            if (Get-Command docker -ErrorAction SilentlyContinue) {
                & docker info --format '{{.ServerVersion}}' *> $null
                $dockerReady = $LASTEXITCODE -eq 0
                $global:LASTEXITCODE = 0
            }
            if ($dockerReady) {
                $gitBash = 'C:\Program Files\Git\bin\bash.exe'
                if (-not (Test-Path -LiteralPath $gitBash -PathType Leaf)) {
                    throw 'Git for Windows bash is required to run verify-proto-generation.sh'
                }
                & $gitBash ./scripts/verify-proto-generation.sh
            }
            elseif (Get-Command wsl.exe -ErrorAction SilentlyContinue) {
                $wslRepo = Convert-ToWSLPath -WindowsPath $repo
                $quotedRepo = $wslRepo.Replace("'", "'`"'`"'")
                $command = "set -e; cd '$quotedRepo'; before=`$(mktemp); after=`$(mktemp); trap 'rm -f `"`$before`" `"`$after`"' EXIT; find sdk/go/node sdk/ts/src/generated sdk/artifacts/proto api/openapi -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*.ts' -o -name '*.binpb' -o -name '*.sha256' -o -name 'inventory.json' -o -name 'virtengine-proto.swagger.json' \) -print0 | sort -z | xargs -0 sha256sum > `"`$before`"; ./scripts/proto-generate-wsl.sh all; find sdk/go/node sdk/ts/src/generated sdk/artifacts/proto api/openapi -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*.ts' -o -name '*.binpb' -o -name '*.sha256' -o -name 'inventory.json' -o -name 'virtengine-proto.swagger.json' \) -print0 | sort -z | xargs -0 sha256sum > `"`$after`"; diff -u `"`$before`" `"`$after`"; node --test scripts/protoinventory/inventory.test.mjs"
                & wsl.exe -e bash -lc $command
            }
            else {
                throw 'generated drift verification requires Docker or WSL'
            }
        }

        Invoke-Step 'running focused golangci-lint' {
            $golangci = Get-Command golangci-lint -ErrorAction SilentlyContinue
            if (-not $golangci) {
                $fallback = Join-Path (go env GOPATH) 'bin/golangci-lint.exe'
                if (-not (Test-Path -LiteralPath $fallback -PathType Leaf)) {
                    throw 'golangci-lint is required but was not found on PATH or in GOPATH/bin'
                }
                & $fallback run @lintPackages
                return
            }
            & $golangci.Source run @lintPackages
        }

        Invoke-Step 'building provider-daemon' {
            $output = Join-Path ([System.IO.Path]::GetTempPath()) "virtengine-task85b-provider-daemon-$PID.exe"
            try {
                & go build -o $output ./cmd/provider-daemon
            }
            finally {
                Remove-Item -LiteralPath $output -Force -ErrorAction SilentlyContinue
            }
        }

        Invoke-Step 'building virtengine' {
            $output = Join-Path ([System.IO.Path]::GetTempPath()) "virtengine-task85b-$PID.exe"
            try {
                & go build -o $output ./cmd/virtengine
            }
            finally {
                Remove-Item -LiteralPath $output -Force -ErrorAction SilentlyContinue
            }
        }
    }
    else {
        Write-Host '[85B] Quick mode: skipped golangci-lint and binary builds' -ForegroundColor DarkGray
    }

    if (-not $SkipRace) {
        Invoke-Step 'running focused WSL race tests' {
            if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
                throw 'WSL race gate requested but wsl.exe is unavailable; use -SkipRace only when explicitly approved'
            }
            $wslRepo = Convert-ToWSLPath -WindowsPath $repo
            $quotedRepo = $wslRepo.Replace("'", "'`"'`"'")
            $command = "set -e; cd '$quotedRepo'; CGO_ENABLED=1 go test -p 1 -race ./pkg/dex ./pkg/payments/offramp -count=1; CGO_ENABLED=1 go test -p 1 -race ./pkg/provider_daemon ./cmd/provider-daemon -run 'FiatConversion|FiatWebhook|ProviderMutationFiat|SubmitFiatConversion' -count=1; CGO_ENABLED=1 go test -p 1 -race ./x/settlement/keeper -run 'FiatConversion|ConsensusExternal' -count=1"
            & wsl.exe -e bash -lc $command
        }
    }
    else {
        Write-Host '[85B] Race gate skipped by -SkipRace' -ForegroundColor DarkGray
    }

    Write-Host '[85B] preflight passed' -ForegroundColor Green
}
finally {
    Pop-Location
}
