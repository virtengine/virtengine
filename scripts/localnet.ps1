[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('start', 'stop', 'restart', 'update', 'reset', 'status', 'logs', 'test', 'shell', 'create-admin', 'init-categories')]
    [string]$Command = 'start',

    [Parameter(Position = 1)]
    [string]$Service,
    [string]$Username,
    [string]$Password,
    [string]$Email,
    [switch]$Foreground,
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$ComposeFile = Join-Path $ProjectRoot 'docker-compose.yaml'
$EnvironmentFile = Join-Path $ProjectRoot '.env.localnet'
$LocalEnvironmentFile = "$EnvironmentFile.local"
$LastUpdateFile = Join-Path $ProjectRoot '.localnet-last-update'
$ChainId = if ($env:CHAIN_ID) { $env:CHAIN_ID } else { 'virtengine-localnet-1' }
$LogLevel = if ($env:LOG_LEVEL) { $env:LOG_LEVEL } else { 'info' }

function Write-Info([string]$Message) {
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-Success([string]$Message) {
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-WarningMessage([string]$Message) {
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Test-Command([string]$Name) {
    return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Invoke-Checked([scriptblock]$Action, [string]$FailureMessage) {
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw $FailureMessage
    }
}

function Test-Docker {
    if (-not (Test-Command 'docker')) {
        throw 'Docker Desktop is required. Install and start Docker Desktop, then retry.'
    }

    Invoke-Checked { docker info | Out-Null } 'Docker Desktop is not running or is not available to this user.'
    Invoke-Checked { docker compose version | Out-Null } 'Docker Compose V2 is required. Update Docker Desktop and retry.'
}

function Invoke-Compose {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)

    $composeArguments = @('compose', '-f', $ComposeFile)
    if (Test-Path $EnvironmentFile) {
        $composeArguments += @('--env-file', $EnvironmentFile)
    }
    if (Test-Path $LocalEnvironmentFile) {
        $composeArguments += @('--env-file', $LocalEnvironmentFile)
    }
    $composeArguments += $Arguments
    & docker @composeArguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose $($Arguments -join ' ') failed."
    }
}

function Get-Endpoint([string]$Uri, [switch]$SkipCertificateCheck) {
    try {
        $parameters = @{ Uri = $Uri; TimeoutSec = 5; ErrorAction = 'Stop' }
        if ($SkipCertificateCheck) {
            $parameters.SkipCertificateCheck = $true
        }
        return Invoke-RestMethod @parameters
    }
    catch {
        return $null
    }
}

function Wait-ForEndpoint([string]$Uri, [string]$Name, [int]$Attempts = 60, [int]$DelaySeconds = 2, [switch]$SkipCertificateCheck) {
    Write-Info "Waiting for $Name to be ready..."
    for ($attempt = 1; $attempt -le $Attempts; $attempt++) {
        if ($null -ne (Get-Endpoint -Uri $Uri -SkipCertificateCheck:$SkipCertificateCheck)) {
            Write-Success "$Name is ready."
            return $true
        }
        Write-Host -NoNewline '.'
        Start-Sleep -Seconds $DelaySeconds
    }
    Write-Host ''
    return $false
}

function Get-ServiceSourceHash([string]$Name) {
    $paths = switch ($Name) {
        'virtengine-node' { @('cmd/virtengine', 'x', 'app', '_build/Dockerfile.virtengine', 'scripts/init-chain.sh') }
        'provider-daemon' { @('cmd/provider-daemon', 'pkg/provider_daemon', '_build/Dockerfile.provider-daemon') }
        'portal' { @('portal', 'lib/portal', 'lib/capture', 'lib/admin', '_build/Dockerfile.portal', 'pnpm-workspace.yaml', 'pnpm-lock.yaml') }
        default { return 'upstream' }
    }

    $files = foreach ($path in $paths) {
        $item = Join-Path $ProjectRoot $path
        if (Test-Path $item -PathType Container) {
            Get-ChildItem -Path $item -File -Recurse | Where-Object {
                $_.Extension -in '.go', '.ts', '.tsx', '.js', '.jsx', '.json', '.css', '.md', '.sh' -or $_.Name -like 'Dockerfile*'
            }
        }
        elseif (Test-Path $item -PathType Leaf) {
            Get-Item $item
        }
    }

    $hashInput = $files | Sort-Object FullName | ForEach-Object { (Get-FileHash -Algorithm SHA256 $_.FullName).Hash }
    if (-not $hashInput) {
        return 'empty'
    }
    return ($hashInput -join "`n" | ForEach-Object { [System.BitConverter]::ToString([System.Security.Cryptography.SHA256]::Create().ComputeHash([System.Text.Encoding]::UTF8.GetBytes($_))).Replace('-', '').ToLowerInvariant() })
}

function Start-Localnet {
    Test-Docker
    Write-Info 'Building Docker images (virtengine-node, provider-daemon, portal)...'
    Invoke-Compose build virtengine-node provider-daemon portal

    if ($Foreground) {
        Write-Info 'Starting services in the foreground...'
        Invoke-Compose up
        return
    }

    Write-Info 'Starting services in the background...'
    Invoke-Compose up -d
    if (-not (Wait-ForEndpoint -Uri 'http://localhost:26657/status' -Name 'VirtEngine chain')) {
        throw 'Timed out waiting for the VirtEngine chain. Run .\scripts\localnet.ps1 logs virtengine-node for details.'
    }
    if (-not (Wait-ForEndpoint -Uri 'https://localhost/health-check/' -Name 'Waldur API' -Attempts 24 -DelaySeconds 5 -SkipCertificateCheck)) {
        Write-WarningMessage 'Waldur API is still starting. The localnet remains available; check status again shortly.'
    }
    Show-LocalnetStatus
}

function Stop-Localnet {
    Test-Docker
    Invoke-Compose down
    Write-Success 'VirtEngine localnet stopped.'
}

function Show-LocalnetStatus {
    Test-Docker
    Invoke-Compose ps
    $chainStatus = Get-Endpoint -Uri 'http://localhost:26657/status'
    if ($null -ne $chainStatus) {
        Write-Success "Chain RPC: Healthy (height $($chainStatus.result.sync_info.latest_block_height))"
    }
    else {
        Write-WarningMessage 'Chain RPC: Not responding'
    }
    if ($null -ne (Get-Endpoint -Uri 'https://localhost/health-check/' -SkipCertificateCheck)) {
        Write-Success 'Waldur API: Healthy'
    }
    else {
        Write-WarningMessage 'Waldur API: Not responding'
    }
}

function Update-Localnet {
    Test-Docker
    $running = & docker ps --format '{{.Names}}'
    if ($LASTEXITCODE -ne 0 -or $running -notcontains 'virtengine-node') {
        Write-WarningMessage 'Localnet is not running. Starting it now.'
        Start-Localnet
        return
    }

    $changedServices = @()
    foreach ($serviceName in 'virtengine-node', 'provider-daemon', 'portal') {
        $currentHash = Get-ServiceSourceHash $serviceName
        $marker = "$LastUpdateFile.$serviceName"
        $previousHash = if (Test-Path $marker) { Get-Content -Raw $marker } else { '' }
        if ($currentHash -ne $previousHash.Trim()) {
            $changedServices += $serviceName
        }
    }
    if ($changedServices.Count -eq 0) {
        Write-Success 'No source changes detected.'
        return
    }

    foreach ($serviceName in $changedServices) {
        Write-Info "Rebuilding $serviceName..."
        Invoke-Compose build --no-cache $serviceName
        Set-Content -NoNewline -Path "$LastUpdateFile.$serviceName" -Value (Get-ServiceSourceHash $serviceName)
        Invoke-Compose up -d --no-deps $serviceName
    }
    if ($changedServices -contains 'virtengine-node') {
        if (-not (Wait-ForEndpoint -Uri 'http://localhost:26657/status' -Name 'VirtEngine chain')) {
            throw 'Timed out waiting for rebuilt VirtEngine chain.'
        }
    }
    Write-Success "Update complete. Rebuilt: $($changedServices -join ', ')."
}

function Reset-Localnet {
    if (-not $Force) {
        $confirmation = Read-Host 'This deletes all localnet data. Continue? (y/N)'
        if ($confirmation -notmatch '^[Yy]$') {
            Write-Info 'Aborted.'
            return
        }
    }
    Test-Docker
    Invoke-Compose down -v --remove-orphans
    Remove-Item -Force -Recurse (Join-Path $ProjectRoot '.localnet') -ErrorAction SilentlyContinue
    Start-Localnet
}

function Show-Logs {
    Test-Docker
    if ($Service) {
        Invoke-Compose logs -f $Service
    }
    else {
        Invoke-Compose logs -f
    }
}

function Invoke-IntegrationTests {
    Test-Docker
    if ($null -eq (Get-Endpoint -Uri 'http://localhost:26657/status')) {
        Write-WarningMessage 'Chain is not running. Starting localnet first.'
        Start-Localnet
    }
    Invoke-Compose run --rm test-runner go test -v ./tests/integration/...
}

function Open-LocalnetShell {
    Test-Docker
    Invoke-Compose run --rm test-runner /bin/sh
}

function New-WaldurAdmin {
    Test-Docker
    if (-not $Username) { $Username = Read-Host 'Admin username (admin)' }
    if (-not $Username) { $Username = 'admin' }
    if (-not $Email) { $Email = Read-Host "Admin email ($Username@localhost)" }
    if (-not $Email) { $Email = "$Username@localhost" }
    if (-not $Password) { $Password = Read-Host 'Admin password' -AsSecureString | ConvertFrom-SecureString -AsPlainText }
    if (-not $Password) { throw 'An admin password is required.' }

    Invoke-Checked { docker exec waldur-mastermind-api waldur createsuperuser --username $Username --email $Email --noinput } 'Failed to create the Waldur administrator.'
    $python = "from django.contrib.auth import get_user_model; user = get_user_model().objects.get(username='$Username'); user.set_password('$Password'); user.save()"
    Invoke-Checked { docker exec waldur-mastermind-api waldur shell -c $python } 'Failed to set the Waldur administrator password.'
    Write-Success "Waldur administrator '$Username' is ready."
}

switch ($Command) {
    'start' { Start-Localnet }
    'stop' { Stop-Localnet }
    'restart' { Stop-Localnet; Start-Localnet }
    'update' { Update-Localnet }
    'reset' { Reset-Localnet }
    'status' { Show-LocalnetStatus }
    'logs' { Show-Logs }
    'test' { Invoke-IntegrationTests }
    'shell' { Open-LocalnetShell }
    'create-admin' { New-WaldurAdmin }
    'init-categories' { throw 'Category initialization is not yet exposed by the PowerShell launcher. Use create-admin, then rerun start after Waldur is healthy.' }
}