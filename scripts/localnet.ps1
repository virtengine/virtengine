[CmdletBinding()]
param(
    [ValidateSet('start', 'stop', 'restart', 'update', 'reset', 'status', 'logs', 'test', 'shell', 'help')]
    [string]$Command = 'start',
    [string]$Service
)

$ErrorActionPreference = 'Stop'
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$ComposeFile = Join-Path $ProjectRoot 'docker-compose.yaml'
$EnvFile = Join-Path $ProjectRoot '.env.localnet'

function Assert-Docker {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw 'Docker is not installed or is not available on PATH.'
    }
    & docker info *> $null
    if ($LASTEXITCODE -ne 0) { throw 'Docker daemon is not running.' }
    & docker compose version *> $null
    if ($LASTEXITCODE -ne 0) { throw 'Docker Compose V2 is required.' }
}

function Invoke-Compose {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    $composeArguments = @('compose', '-f', $ComposeFile)
    if (Test-Path $EnvFile) { $composeArguments += @('--env-file', $EnvFile) }
    if (Test-Path "$EnvFile.local") { $composeArguments += @('--env-file', "$EnvFile.local") }
    $composeArguments += $Arguments
    & docker @composeArguments
    if ($LASTEXITCODE -ne 0) { throw "docker compose failed with exit code $LASTEXITCODE" }
}

if ($Command -eq 'help') {
    Write-Host 'Usage: ./scripts/localnet.ps1 [start|stop|restart|update|reset|status|logs|test|shell|help] [-Service name]'
    exit 0
}

Assert-Docker
switch ($Command) {
    'start' { Invoke-Compose up -d --build }
    'stop' { Invoke-Compose down }
    'restart' { Invoke-Compose down; Invoke-Compose up -d --build }
    'update' { Invoke-Compose up -d --build }
    'reset' { Invoke-Compose down -v --remove-orphans; Invoke-Compose up -d --build }
    'status' { Invoke-Compose ps }
    'logs' {
        if ($Service) { Invoke-Compose logs -f $Service } else { Invoke-Compose logs -f }
    }
    'test' { Invoke-Compose run --rm test-runner go test -count=1 -tags=e2e.integration -v ./tests/integration/... }
    'shell' { Invoke-Compose run --rm test-runner /bin/sh }
}