#!/usr/bin/env pwsh
<#
.SYNOPSIS
    VirtEngine Task Orchestrator — automated task chaining with parallel execution.

.DESCRIPTION
    Long-running orchestration loop that:
    1. Maintains a target number of parallel task attempts
    2. Cycles agents 50/50 between Codex and Copilot to avoid rate-limiting
    3. Monitors agent completion (PR creation) and CI status
    4. Auto-merges PRs when CI passes
    5. Ensures previous tasks are merged before starting new ones
    6. Marks completed tasks as done
    7. Submits the next todo task to fill the slot
    8. Repeats until the backlog is empty

.PARAMETER MaxParallel
    Maximum number of concurrent task attempts (default: 2).

.PARAMETER PollIntervalSec
    Seconds between orchestration cycles (default: 300 = 5 minutes).

.PARAMETER MutexWaitMaxMin
    Maximum minutes to wait for an existing orchestrator mutex before exiting.
    Use 0 to wait indefinitely.

.PARAMETER DryRun
    If set, logs what would happen without making changes.

.PARAMETER WaitForMutex
    Force waiting for an existing orchestrator instance to release the mutex.

.PARAMETER NoWaitForMutex
    Exit immediately if the mutex is held (overrides default wait).

.PARAMETER OneShot
    Run a single orchestration cycle and exit (useful for testing).

.EXAMPLE
    # Run with 2 parallel agents, polling every 5 minutes
    ./ve-orchestrator.ps1

    # Run with 3 parallel agents, polling every 3 minutes
    ./ve-orchestrator.ps1 -MaxParallel 3 -PollIntervalSec 180

    # Dry-run to see what would happen
    ./ve-orchestrator.ps1 -DryRun

    # Single cycle (no loop)
    ./ve-orchestrator.ps1 -OneShot
#>
[CmdletBinding()]
param(
    [int]$MaxParallel = 2,
    [int]$PollIntervalSec = 90,
    [int]$MutexWaitMaxMin = 15,
    [int]$GitHubCooldownSec = 120,
    [int]$VKApiTimeoutSec = 20,
    [int]$VKApiRetryCount = 3,
    [int]$VKApiRetryDelaySec = 5,
    [int]$GitHubCommandTimeoutSec = 120,
    [int]$IdleTimeoutMin = 60,
    [int]$IdleConfirmMin = 15,
    [int]$StaleRunningTimeoutMin = 45,
    [int]$CiWaitMin = 15,
    [int]$MaxRetries = 5,
    [int]$KillBurstThreshold = 2,
    [int]$KillBurstWindowMin = 10,
    [int]$KillCooldownMin = 15,
    [int]$KillThrottleMaxParallel = 2,
    [switch]$UseAutoMerge,
    [switch]$UseAdminMerge,
    [switch]$SkipSecurityChecks,
    [switch]$WaitForMutex,
    [switch]$NoWaitForMutex,
    [switch]$DryRun,
    [switch]$OneShot,
    [switch]$RunMergeStrategy,
    [switch]$SyncCopilotState
)

# Default to waiting on the mutex for long-running runs unless explicitly overridden.
if ($NoWaitForMutex) {
    $script:WaitForMutexEffective = $false
}
elseif ($WaitForMutex) {
    $script:WaitForMutexEffective = $true
}
elseif (-not $OneShot) {
    $script:WaitForMutexEffective = $true
}
else {
    $script:WaitForMutexEffective = $false
}

# ─── Load ve-kanban library ──────────────────────────────────────────────────
function Get-VeKanbanLibraryCandidates {
    $candidates = @()
    if ($PSScriptRoot) {
        $candidates += (Join-Path $PSScriptRoot "ve-kanban.ps1")
        $candidates += (Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) "ve-kanban.ps1")
        $candidates += (Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) "scripts\\ve-kanban.ps1")
    }
    if ($PSCommandPath) {
        $candidates += (Join-Path (Split-Path -Parent $PSCommandPath) "ve-kanban.ps1")
    }
    $candidates += (Join-Path (Get-Location) "ve-kanban.ps1")
    $candidates = $candidates | Select-Object -Unique

    return $candidates
}

function Resolve-VeKanbanLibraryPath {
    <#
    .SYNOPSIS Resolve the ve-kanban library path from known candidates.
    #>
    try {
        if ($script:VeKanbanLibraryPath -and (Test-Path -LiteralPath $script:VeKanbanLibraryPath)) {
            return $script:VeKanbanLibraryPath
        }
    }
    catch { }

    $candidates = Get-VeKanbanLibraryCandidates
    foreach ($path in $candidates) {
        if (-not (Test-Path -LiteralPath $path)) { continue }
        return $path
    }
    return $null
}

$script:VeKanbanLibraryCandidates = Get-VeKanbanLibraryCandidates
$script:VeKanbanLibraryPath = $null
foreach ($path in $script:VeKanbanLibraryCandidates) {
    if (-not (Test-Path -LiteralPath $path)) { continue }
    $script:VeKanbanLibraryPath = $path
    break
}
if (-not $script:VeKanbanLibraryPath) {
    throw "ve-kanban library not found. Tried: $($script:VeKanbanLibraryCandidates -join ', ')"
}
try {
    . $script:VeKanbanLibraryPath
}
catch {
    throw "Failed to load ve-kanban library from '$script:VeKanbanLibraryPath': $($_.Exception.Message)"
}

$requiredFunctions = @(
    "Initialize-VKConfig",
    "Get-VKTasks",
    "Get-VKAttempts",
    "Get-VKAttemptSummaries",
    "Get-OpenPullRequests",
    "Get-VKLastGithubError",
    "Get-CurrentExecutorProfile",
    "Get-NextExecutorProfile",
    "Initialize-ExecutorHealth",
    "Report-ExecutorFailure",
    "Report-ExecutorSuccess",
    "Get-HealthyExecutorProfile",
    "Get-ExecutorProviderForAttempt",
    "Increment-ExecutorActiveCount",
    "Decrement-ExecutorActiveCount",
    "Get-ExecutorHealthSummary",
    "Get-TaskComplexity",
    "Get-BestExecutorForTask",
    "Switch-CodexRegion",
    "Get-RegionStatus",
    "Initialize-CodexRegionTracking",
    "Set-RegionOverride",
    "Test-RegionCooldownExpired",
    "Get-ActiveCodexRegion"
)
function Ensure-VeKanbanLibraryLoaded {
    param([string[]]$RequiredFunctions)
    $missing = $RequiredFunctions | Where-Object { -not (Get-Command $_ -ErrorAction SilentlyContinue) }
    if ($missing.Count -eq 0) { return $true }

    # Retry dot-sourcing in case the initial load occurred in a narrower scope.
    try {
        . $script:VeKanbanLibraryPath
    }
    catch {
        Write-Error "Failed to reload ve-kanban library from '$script:VeKanbanLibraryPath': $($_.Exception.Message)"
    }

    $missing = $RequiredFunctions | Where-Object { -not (Get-Command $_ -ErrorAction SilentlyContinue) }
    if ($missing.Count -eq 0) { return $true }

    # Final fallback: import as a module to force global visibility.
    try {
        Import-Module -Name $script:VeKanbanLibraryPath -Force -Global
    }
    catch {
        Write-Error "Failed to import ve-kanban library module from '$script:VeKanbanLibraryPath': $($_.Exception.Message)"
    }

    $missing = $RequiredFunctions | Where-Object { -not (Get-Command $_ -ErrorAction SilentlyContinue) }
    if ($missing.Count -gt 0) {
        throw "ve-kanban library loaded but missing required functions: $($missing -join ', ')"
    }
    return $true
}

Ensure-VeKanbanLibraryLoaded -RequiredFunctions $requiredFunctions | Out-Null

# ─── State tracking ──────────────────────────────────────────────────────────
$script:CycleCount = 0
$script:TasksCompleted = 0
$script:TasksSubmitted = 0
$script:StartTime = Get-Date
$script:GitHubCooldownUntil = $null
$script:KillCooldownUntil = $null
$script:KillEvents = @()
$script:TaskRetryCounts = @{}
$script:AttemptSummaries = @{}
$script:StatePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\ve-orchestrator-state.json"
$script:CopilotStatePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\ve-orchestrator-copilot.json"
$script:StatusStatePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\ve-orchestrator-status.json"
$script:StopFilePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\ve-orchestrator-stop"
$script:CompletedTasks = @()
$script:SubmittedTasks = @()
$script:FollowUpEvents = @()
$script:CopilotRequests = @()
$script:ConflictRetries = @()
$script:TodoBacklogCount = $null
$script:TodoBacklogFetchedAt = $null

# ─── Success rate tracking ────────────────────────────────────────────────────
$script:FirstShotSuccess = 0          # Merged on first attempt, no copilot fix
$script:TasksNeededFix = 0            # Required copilot sub-PR or manual intervention
$script:TasksFailed = 0               # Abandoned / rejected / manual_review
$script:SuccessMetricsPath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\ve-success-metrics.json"
$script:OrchestratorMutex = $null
$script:StopRequested = $false
$script:StopReason = $null

# Track attempts we're monitoring: attempt_id → { task_id, branch, pr_number, status, executor }
$script:TrackedAttempts = @{}

# Merge gate: track tasks that completed their PR but haven't been merged yet
# This ensures we don't start new tasks until previous ones are merged & confirmed
$script:PendingMerges = @{}

# ─── Logging ──────────────────────────────────────────────────────────────────
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $ts = Get-Date -Format "HH:mm:ss"
    $color = switch ($Level) {
        "INFO" { "White" }
        "OK" { "Green" }
        "WARN" { "Yellow" }
        "ERROR" { "Red" }
        "ACTION" { "Cyan" }
        default { "Gray" }
    }
    Write-Host "  [$ts] $Message" -ForegroundColor $color
}

function Reset-TerminalIfRequested {
    <#
    .SYNOPSIS Clear the terminal when VE_RESET_TERMINAL is set (temporary operator override).
    #>
    $reset = $env:VE_RESET_TERMINAL
    if (-not $reset) { return }
    if ($reset -in @("1", "true", "True", "TRUE", "yes", "YES")) {
        try {
            # ANSI clear + cursor home for terminals that respect VT sequences
            [Console]::Write("`e[2J`e[H")
        }
        catch { }
        try { Clear-Host } catch { }
        Write-Log "Terminal reset requested (VE_RESET_TERMINAL=$reset)" -Level "INFO"
    }
}

function Request-OrchestratorStop {
    param([string]$Reason = "shutdown")
    if ($script:StopRequested) { return }
    $script:StopRequested = $true
    $script:StopReason = $Reason
    Write-Log "Stop requested ($Reason). Exiting after current step." -Level "WARN"
}

function Test-OrchestratorStop {
    if ($script:StopRequested) { return $true }
    if ($script:StopFilePath -and (Test-Path -LiteralPath $script:StopFilePath)) {
        Request-OrchestratorStop -Reason "stop-file"
        return $true
    }
    return $false
}

function Start-InterruptibleSleep {
    param(
        [Parameter(Mandatory)][int]$Seconds,
        [string]$Reason = "sleep",
        [int]$HeartbeatSec = 0
    )
    $remaining = [Math]::Max(0, [int]$Seconds)
    $nextHeartbeat = if ($HeartbeatSec -gt 0) { (Get-Date).AddSeconds($HeartbeatSec) } else { $null }
    while ($remaining -gt 0) {
        if (Test-OrchestratorStop) { return $false }
        $chunk = [Math]::Min(5, $remaining)
        Start-Sleep -Seconds $chunk
        $remaining -= $chunk
        if ($nextHeartbeat -and (Get-Date) -ge $nextHeartbeat -and $remaining -gt 0) {
            Write-Log "Sleep heartbeat ($Reason): ${remaining}s remaining" -Level "INFO"
            $nextHeartbeat = $nextHeartbeat.AddSeconds($HeartbeatSec)
        }
    }
    return $true
}

function Start-MutexWaitWithHeartbeat {
    param(
        [Parameter(Mandatory)][int]$Seconds,
        [int]$HeartbeatSec = 15
    )
    $remaining = [Math]::Max(0, [int]$Seconds)
    if ($HeartbeatSec -le 0) {
        return (Start-InterruptibleSleep -Seconds $remaining -Reason "mutex-wait")
    }
    while ($remaining -gt 0) {
        $chunk = [Math]::Min($HeartbeatSec, $remaining)
        if (-not (Start-InterruptibleSleep -Seconds $chunk -Reason "mutex-wait")) { return $false }
        $remaining -= $chunk
        if ($remaining -gt 0) {
            Write-Log "Still waiting for active orchestrator. Retrying in ${remaining}s." -Level "WARN"
        }
    }
    return $true
}

function Register-OrchestratorShutdownHandlers {
    try {
        Register-EngineEvent -SourceIdentifier "PowerShell.Exiting" -Action {
            try { Request-OrchestratorStop -Reason "PowerShell.Exiting" } catch { }
        } | Out-Null
    }
    catch { }
    try {
        Register-ObjectEvent -InputObject ([Console]) -EventName CancelKeyPress -Action {
            try { Request-OrchestratorStop -Reason "console-cancel" } catch { }
            $EventArgs.Cancel = $true
        } | Out-Null
    }
    catch { }
}

function Get-OrchestratorMutexName {
    param([string]$BaseName = "VirtEngineOrchestrator")
    $rootPath = $null
    try {
        $rootPath = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
    }
    catch {
        $rootPath = (Get-Location).Path
    }
    if (-not $rootPath) {
        return $BaseName
    }
    $hash = [System.Security.Cryptography.SHA256]::Create()
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($rootPath.ToLowerInvariant())
    $digest = $hash.ComputeHash($bytes)
    $hash.Dispose()
    $hashHex = ([System.BitConverter]::ToString($digest)).Replace("-", "")
    $suffix = $hashHex.Substring(0, 12)
    return "$BaseName.$suffix"
}

function Enter-OrchestratorMutex {
    param([string]$Name = (Get-OrchestratorMutexName))
    $createdNew = $false
    # Use a plain name — skip Global\ prefix to avoid Windows session/permission errors.
    $mutexName = $Name
    $mutex = $null
    try {
        $mutex = New-Object System.Threading.Mutex($false, $mutexName, [ref]$createdNew)
    }
    catch {
        Write-Warning "Mutex creation failed: $($_.Exception.Message). Proceeding without lock."
        # Return a sentinel so the caller knows to proceed (not block).
        return [PSCustomObject]@{ NoOp = $true }
    }
    $acquired = $false
    try {
        $acquired = $mutex.WaitOne(0)
    }
    catch [System.Threading.AbandonedMutexException] {
        # Previous orchestrator was killed; take ownership to avoid a deadlock loop.
        $acquired = $true
    }
    catch {
        $acquired = $false
    }
    if (-not $acquired) {
        if ($mutex) { $mutex.Dispose() }
        return $null
    }
    return $mutex
}

function Format-ShortId {
    param([string]$Value, [int]$Length = 8)
    if ([string]::IsNullOrWhiteSpace($Value)) { return "" }
    if ($Value.Length -le $Length) { return $Value }
    return $Value.Substring(0, $Length)
}

function Write-Banner {
    $nextStr = "/"
    $healthStr = "N/A"
    if (Get-Command Get-CurrentExecutorProfile -ErrorAction SilentlyContinue) {
        try {
            $nextExec = Get-CurrentExecutorProfile -ErrorAction Stop
            $nextStr = "$($nextExec.executor)/$($nextExec.variant)"
        }
        catch {
            $nextStr = "/"
        }
    }
    if (Get-Command Get-ExecutorHealthSummary -ErrorAction SilentlyContinue) {
        try {
            $healthStr = Get-ExecutorHealthSummary -ErrorAction Stop
        }
        catch {
            $healthStr = "N/A"
        }
    }
    Write-Host ""
    Write-Host "  ╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║          VirtEngine Task Orchestrator                    ║" -ForegroundColor Cyan
    Write-Host "  ║                                                         ║" -ForegroundColor Cyan
    Write-Host "  ║   Parallel: $($MaxParallel.ToString().PadRight(4))  Poll: ${PollIntervalSec}s  $(if($DryRun){'DRY-RUN'}else{'LIVE'})                ║" -ForegroundColor Cyan
    Write-Host "  ║   Stale:    ${StaleRunningTimeoutMin}m timeout                                  ║" -ForegroundColor Cyan
    Write-Host "  ║   Routing:  Health-Aware Multi-Executor                 ║" -ForegroundColor Cyan
    Write-Host "  ║   Next:     $($nextStr.PadRight(44))║" -ForegroundColor Cyan
    Write-Host "  ║   Health:   $($healthStr.PadRight(44))║" -ForegroundColor Cyan
    Write-Host "  ╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Ensure-GitIdentity {
    <#
    .SYNOPSIS Ensure git author/committer identity is set when env overrides are provided.
    #>
    $name = $env:VE_GIT_AUTHOR_NAME
    if (-not $name) { $name = $env:GIT_AUTHOR_NAME }
    $email = $env:VE_GIT_AUTHOR_EMAIL
    if (-not $email) { $email = $env:GIT_AUTHOR_EMAIL }

    if ($name) {
        try { git config user.name $name | Out-Null } catch { }
        $env:GIT_AUTHOR_NAME = $name
        $env:GIT_COMMITTER_NAME = $name
    }
    if ($email) {
        try { git config user.email $email | Out-Null } catch { }
        $env:GIT_AUTHOR_EMAIL = $email
        $env:GIT_COMMITTER_EMAIL = $email
    }
    if ($name -or $email) {
        Write-Log "Git identity configured from VE_GIT_AUTHOR_*" -Level "INFO"
    }
}

function Write-CycleSummary {
    $elapsed = (Get-Date) - $script:StartTime
    Write-Host ""
    Write-Host "  ── Cycle $($script:CycleCount) ──────────────────────────────" -ForegroundColor DarkCyan
    Write-Host "  │ Elapsed:   $([math]::Round($elapsed.TotalMinutes, 1)) min" -ForegroundColor DarkGray
    Write-Host "  │ Submitted: $($script:TasksSubmitted)  Completed: $($script:TasksCompleted)" -ForegroundColor DarkGray
    Write-Host "  │ Tracked:   $($script:TrackedAttempts.Count) attempts" -ForegroundColor DarkGray
    $successSummary = Get-SuccessRateSummary
    Write-Host "  │ $successSummary" -ForegroundColor DarkGray
    Write-Host "  └────────────────────────────────────────────" -ForegroundColor DarkCyan
}

function Format-ShortId {
    param([string]$Value, [int]$Length = 8)
    if ([string]::IsNullOrWhiteSpace($Value)) { return "" }
    if ($Value.Length -le $Length) { return $Value }
    return $Value.Substring(0, $Length)
}

function Invoke-GhWithTimeout {
    <#
    .SYNOPSIS Run gh CLI with a hard timeout to avoid hung PR creation.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string[]]$Args,
        [int]$TimeoutSec = 120
    )

    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "gh"
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true
    $psi.Environment["GH_PROMPT_DISABLED"] = "1"
    $psi.Environment["GIT_TERMINAL_PROMPT"] = "0"

    if ($psi.PSObject.Properties.Name -contains "ArgumentList") {
        foreach ($arg in $Args) {
            $null = $psi.ArgumentList.Add($arg)
        }
    }
    else {
        $escaped = $Args | ForEach-Object {
            if ($_ -match '[\s"]') { '"' + ($_ -replace '"', '\"') + '"' } else { $_ }
        }
        $psi.Arguments = ($escaped -join ' ')
    }

    $proc = New-Object System.Diagnostics.Process
    $proc.StartInfo = $psi
    $null = $proc.Start()

    $timedOut = -not $proc.WaitForExit([Math]::Max(1, $TimeoutSec) * 1000)
    if ($timedOut) {
        try { $proc.Kill($true) } catch { try { $proc.Kill() } catch { } }
        return @{
            timed_out = $true
            exit_code = $null
            output    = ""
            error     = "timeout"
        }
    }

    $output = $proc.StandardOutput.ReadToEnd()
    $error = $proc.StandardError.ReadToEnd()
    return @{
        timed_out = $false
        exit_code = $proc.ExitCode
        output    = $output
        error     = $error
    }
}

function Create-PRForBranchSafe {
    <#
    .SYNOPSIS Create a PR for a branch with timeout + non-interactive guardrails.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Branch,
        [Parameter(Mandatory)][string]$Title,
        [string]$Body = "Automated PR created by ve-orchestrator"
    )
    $baseBranch = $script:VK_TARGET_BRANCH
    if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }

    $result = Invoke-GhWithTimeout -Args @(
        "pr", "create", "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--head", $Branch,
        "--base", $baseBranch,
        "--title", $Title,
        "--body", $Body
    ) -TimeoutSec $GitHubCommandTimeoutSec

    if ($result.timed_out) {
        if (Get-Command Set-VKLastGithubError -ErrorAction SilentlyContinue) {
            Set-VKLastGithubError -Type "timeout" -Message "gh pr create timed out after ${GitHubCommandTimeoutSec}s"
        }
        Write-Log "gh pr create timed out after ${GitHubCommandTimeoutSec}s for $Branch" -Level "WARN"
        return $false
    }

    $combined = ($result.output + "`n" + $result.error).Trim()
    if ($result.exit_code -ne 0) {
        if ($combined -match "rate limit|API rate limit exceeded|secondary rate limit|abuse detection") {
            if (Get-Command Set-VKLastGithubError -ErrorAction SilentlyContinue) {
                Set-VKLastGithubError -Type "rate_limit" -Message $combined
            }
        }
        else {
            if (Get-Command Set-VKLastGithubError -ErrorAction SilentlyContinue) {
                Set-VKLastGithubError -Type "error" -Message $combined
            }
        }

        if ($combined -match "pull request already exists|already exists") {
            if (Get-Command Clear-VKLastGithubError -ErrorAction SilentlyContinue) {
                Clear-VKLastGithubError
            }
            return $true
        }

        # Non-retryable: branch has no commits vs base — stop looping
        if ($combined -match "No commits between|no difference|nothing to compare") {
            Write-Log "gh pr create failed for ${Branch}: branch has no commits vs base — marking no_commits" -Level "WARN"
            return "no_commits"
        }

        Write-Log "gh pr create failed for ${Branch}: $combined" -Level "WARN"
        return $false
    }

    if (Get-Command Clear-VKLastGithubError -ErrorAction SilentlyContinue) {
        Clear-VKLastGithubError
    }
    return $true
}

function Test-VKApiReachable {
    <#
    .SYNOPSIS Fast health check for the vibe-kanban API to avoid blocking init.
    #>
    [CmdletBinding()]
    param([int]$TimeoutSec = 5)

    $baseUrl = Get-VKBaseUrl
    $timeout = if ($TimeoutSec -gt 0) { $TimeoutSec } else { 5 }
    $client = $null
    try {
        $handler = New-Object System.Net.Http.HttpClientHandler
        $client = New-Object System.Net.Http.HttpClient($handler)
        $client.Timeout = [TimeSpan]::FromSeconds($timeout)
        $resp = $client.GetAsync("$baseUrl/api/projects").GetAwaiter().GetResult()
        return $resp.IsSuccessStatusCode
    }
    catch {
        return $false
    }
    finally {
        if ($client) { $client.Dispose() }
    }
}

function Get-VKBaseUrl {
    if ($script:VK_BASE_URL) { return $script:VK_BASE_URL }
    if ($env:VK_BASE_URL) { return $env:VK_BASE_URL }
    return "http://127.0.0.1:54089"
}

function Get-OrchestratorState {
    if (-not (Test-Path $script:StatePath)) {
        return @{
            last_sequence_value     = $null
            last_task_id            = $null
            last_submitted_at       = $null
            last_ci_sweep_completed = 0
            vk_project_id           = $null
            vk_repo_id              = $null
        }
    }
    try {
        $raw = Get-Content -Path $script:StatePath -Raw
        if (-not $raw) {
            return @{
                last_sequence_value     = $null
                last_task_id            = $null
                last_submitted_at       = $null
                last_ci_sweep_completed = 0
                vk_project_id           = $null
                vk_repo_id              = $null
            }
        }
        $state = $raw | ConvertFrom-Json -Depth 5
        $lastSweep = if ($state.last_ci_sweep_completed) { [int]$state.last_ci_sweep_completed } else { 0 }
        return @{
            last_sequence_value     = $state.last_sequence_value
            last_task_id            = $state.last_task_id
            last_submitted_at       = $state.last_submitted_at
            last_ci_sweep_completed = $lastSweep
            vk_project_id           = $state.vk_project_id
            vk_repo_id              = $state.vk_repo_id
        }
    }
    catch {
        return @{
            last_sequence_value     = $null
            last_task_id            = $null
            last_submitted_at       = $null
            last_ci_sweep_completed = 0
            vk_project_id           = $null
            vk_repo_id              = $null
        }
    }
}

function Save-OrchestratorState {
    param(
        [Parameter(Mandatory)][hashtable]$State
    )
    $dir = Split-Path -Parent $script:StatePath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
    $State | ConvertTo-Json -Depth 5 | Set-Content -Path $script:StatePath -Encoding UTF8
}

function Apply-CachedVKConfig {
    $state = Get-OrchestratorState
    $cachedProjectId = $state.vk_project_id
    $cachedRepoId = $state.vk_repo_id
    $applied = $false

    if (-not $script:VK_PROJECT_ID -and $cachedProjectId) {
        $script:VK_PROJECT_ID = $cachedProjectId
        $env:VK_PROJECT_ID = $cachedProjectId
        $applied = $true
        Write-Log "Using cached VK project ID $($cachedProjectId.Substring(0,8))..." -Level "INFO"
    }

    if (-not $script:VK_REPO_ID -and $cachedRepoId) {
        $script:VK_REPO_ID = $cachedRepoId
        $env:VK_REPO_ID = $cachedRepoId
        $applied = $true
        Write-Log "Using cached VK repo ID $($cachedRepoId.Substring(0,8))..." -Level "INFO"
    }

    return $applied
}

function Save-VKConfigCache {
    $projectId = $script:VK_PROJECT_ID
    $repoId = $script:VK_REPO_ID
    if (-not $projectId -and -not $repoId) { return }

    $state = Get-OrchestratorState
    $updated = $false
    if ($projectId -and $state.vk_project_id -ne $projectId) {
        $state.vk_project_id = $projectId
        $updated = $true
    }
    if ($repoId -and $state.vk_repo_id -ne $repoId) {
        $state.vk_repo_id = $repoId
        $updated = $true
    }
    if ($updated) {
        Save-OrchestratorState -State $state
    }
}

function Get-CopilotState {
    if (-not (Test-Path $script:CopilotStatePath)) {
        return @{ prs = @{} }
    }
    try {
        $raw = Get-Content -Path $script:CopilotStatePath -Raw
        if (-not $raw) { return @{ prs = @{} } }
        $state = $raw | ConvertFrom-Json -Depth 6
        if (-not $state.prs) { return @{ prs = @{} } }
        if ($state.prs -is [hashtable]) {
            return @{ prs = $state.prs }
        }
        $prs = @{}
        foreach ($prop in $state.prs.PSObject.Properties) {
            $prs[$prop.Name] = $prop.Value
        }
        return @{ prs = $prs }
    }
    catch {
        return @{ prs = @{} }
    }
}

function Save-CopilotState {
    param(
        [Parameter(Mandatory)][hashtable]$State
    )
    $dir = Split-Path -Parent $script:CopilotStatePath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
    $State | ConvertTo-Json -Depth 6 | Set-Content -Path $script:CopilotStatePath -Encoding UTF8
}

function Get-CopilotPRState {
    param([Parameter(Mandatory)][int]$PRNumber)
    $state = Get-CopilotState
    $key = $PRNumber.ToString()
    if ($state.prs.ContainsKey($key)) { return $state.prs[$key] }
    return $null
}

function Set-CopilotPRState {
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [Parameter(Mandatory)][hashtable]$Value
    )
    $state = Get-CopilotState
    if (-not ($state.prs -is [hashtable])) {
        $state.prs = @{}
    }
    $key = $PRNumber.ToString()
    $state.prs[$key] = $Value
    Save-CopilotState -State $state
}

function Remove-CopilotPRState {
    param([Parameter(Mandatory)][int]$PRNumber)
    $state = Get-CopilotState
    if (-not ($state.prs -is [hashtable])) {
        $state.prs = @{}
    }
    $key = $PRNumber.ToString()
    if ($state.prs.ContainsKey($key)) {
        $state.prs.Remove($key)
        Save-CopilotState -State $state
    }
}

function Upsert-CopilotPRState {
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [Parameter(Mandatory)][hashtable]$Update
    )
    $state = Get-CopilotState
    if (-not ($state.prs -is [hashtable])) {
        $state.prs = @{}
    }
    $key = $PRNumber.ToString()
    $existing = $null
    if ($state.prs.ContainsKey($key)) {
        $existing = $state.prs[$key]
    }
    $requestedAt = if ($existing -and $existing.requested_at) { $existing.requested_at } else { $Update.requested_at }
    $mergedAt = if ($Update.merged_at) { $Update.merged_at } else { $existing.merged_at }
    $state.prs[$key] = @{
        requested_at = $requestedAt
        completed    = $Update.completed
        copilot_pr   = $Update.copilot_pr
        merged_at    = $mergedAt
    }
    Save-CopilotState -State $state
}

function Apply-CopilotStateToInfo {
    param(
        [Parameter(Mandatory)][hashtable]$Info,
        [Parameter(Mandatory)][int]$PRNumber
    )
    $copilotState = Get-CopilotPRState -PRNumber $PRNumber
    if (-not $copilotState) { return $null }
    $Info.copilot_fix_requested = $true
    $Info.copilot_fix_pr_number = $copilotState.copilot_pr
    $Info.copilot_fix_merged = [bool]$copilotState.completed
    if ($copilotState.requested_at) {
        try { $Info.copilot_fix_requested_at = [datetime]::Parse($copilotState.requested_at) } catch { }
    }
    if ($copilotState.merged_at) {
        try { $Info.copilot_fix_merged_at = [datetime]::Parse($copilotState.merged_at) } catch { }
    }
    return $copilotState
}

function Get-ReferencedPRNumbers {
    param([string[]]$Texts)
    $refs = New-Object System.Collections.Generic.HashSet[int]
    if (-not $Texts) { return @() }
    $pattern = '(?i)(?:PR\s*#?\s*|#)(\d{1,6})'
    foreach ($text in $Texts) {
        if (-not $text) { continue }
        foreach ($match in [regex]::Matches($text, $pattern)) {
            $value = $match.Groups[1].Value
            if ($value) { [void]$refs.Add([int]$value) }
        }
    }
    return @($refs | ForEach-Object { $_ })
}

function Sync-CopilotPRState {
    <#
    .SYNOPSIS Seed Copilot state from existing open Copilot PRs.
    #>
    $openPrs = Get-OpenPullRequests -Limit 200
    if (Test-GithubRateLimit) { return }
    if (-not $openPrs -or @($openPrs).Count -eq 0) { return }

    foreach ($pr in $openPrs) {
        if (-not $pr.author -or -not (Test-IsCopilotAuthor -Author $pr.author)) { continue }
        $details = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        $refs = Get-ReferencedPRNumbers -Texts @(
            $pr.title,
            $details.body,
            $details.headRefName,
            $pr.headRefName
        )
        $isComplete = if ($pr.title -match '^\[WIP\]') { $false } else { $true }
        $requestedAt = if ($pr.createdAt) { $pr.createdAt } else { (Get-Date).ToString("o") }
        if (-not $refs -or @($refs).Count -eq 0) {
            if (-not (Get-CopilotPRState -PRNumber $pr.number)) {
                Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                    requested_at = $requestedAt
                    completed    = $isComplete
                    copilot_pr   = $pr.number
                    merged_at    = $null
                }
                Write-Log "Synced Copilot PR #$($pr.number) (no reference)" -Level "INFO"
            }
            continue
        }

        foreach ($ref in $refs) {
            if (Get-CopilotPRState -PRNumber $ref) { continue }
            Upsert-CopilotPRState -PRNumber $ref -Update @{
                requested_at = $requestedAt
                completed    = $isComplete
                copilot_pr   = $pr.number
                merged_at    = $null
            }
            Write-Log "Synced Copilot PR #$($pr.number) for PR #$ref" -Level "INFO"
        }
    }
}

function Get-TaskUrl {
    param([Parameter(Mandatory)][string]$TaskId)
    $template = $env:VK_TASK_URL_TEMPLATE
    if ($template) {
        return $template.Replace("{taskId}", $TaskId).Replace("{projectId}", $script:VK_PROJECT_ID)
    }
    $base = $env:VK_BOARD_URL
    if (-not $base) { $base = $env:VK_WEB_URL }
    if (-not $base) { return $null }
    return "$base/tasks/$TaskId"
}

function Add-RecentItem {
    param(
        [Parameter(Mandatory)][hashtable]$Item,
        [Parameter(Mandatory)][string]$ListName,
        [int]$Limit = 200
    )
    try {
        $current = Get-Variable -Name $ListName -Scope Script -ValueOnly -ErrorAction Stop
    }
    catch {
        $current = @()
        Set-Variable -Name $ListName -Scope Script -Value $current
    }
    $updated = @($current + $Item)
    if ($updated.Count -gt $Limit) {
        $updated = $updated | Select-Object -Last $Limit
    }
    Set-Variable -Name $ListName -Scope Script -Value $updated
}

function Set-TodoBacklogCount {
    param([int]$Count)
    $script:TodoBacklogCount = [math]::Max(0, $Count)
    $script:TodoBacklogFetchedAt = Get-Date
}

function Get-TodoBacklogCount {
    param([int]$MaxAgeSec = 0)
    if (-not $script:TodoBacklogFetchedAt) { return $null }
    if ($MaxAgeSec -gt 0) {
        $ageSec = ((Get-Date) - $script:TodoBacklogFetchedAt).TotalSeconds
        if ($ageSec -gt $MaxAgeSec) { return $null }
    }
    return $script:TodoBacklogCount
}

function Save-StatusSnapshot {
    $dir = Split-Path -Parent $script:StatusStatePath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
    $attempts = @{}
    foreach ($entry in $script:TrackedAttempts.GetEnumerator()) {
        $attemptId = $entry.Key
        $info = $entry.Value
        $attempts.$attemptId = @{
            task_id               = $info.task_id
            branch                = $info.branch
            pr_number             = $info.pr_number
            status                = $info.status
            updated_at            = (Get-Date).ToString("o")
            last_process_status   = $info.last_process_status
            copilot_fix_requested = $info.copilot_fix_requested
            copilot_fix_pr_number = $info.copilot_fix_pr_number
            copilot_fix_merged    = $info.copilot_fix_merged
        }
    }
    $counts = Get-TrackedStatusCounts
    $reviewTasks = @($script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.status -eq "review" } | ForEach-Object { $_.Value.task_id })
    $errorTasks = @($script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.status -eq "error" } | ForEach-Object { $_.Value.task_id })
    $manualReviewTasks = @($script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.status -eq "manual_review" } | ForEach-Object { $_.Value.task_id })
    $backlogRemaining = Get-TodoBacklogCount -MaxAgeSec ([math]::Max(120, ($PollIntervalSec * 2)))
    if ($null -eq $backlogRemaining) { $backlogRemaining = 0 }
    $snapshot = @{
        updated_at          = (Get-Date).ToString("o")
        counts              = $counts
        tasks_submitted     = $script:TasksSubmitted
        tasks_completed     = $script:TasksCompleted
        backlog_remaining   = $backlogRemaining
        completed_tasks     = $script:CompletedTasks
        submitted_tasks     = $script:SubmittedTasks
        followup_events     = $script:FollowUpEvents
        copilot_requests    = $script:CopilotRequests
        review_tasks        = $reviewTasks
        error_tasks         = $errorTasks
        manual_review_tasks = $manualReviewTasks
        attempts            = $attempts
        success_metrics     = @{
            first_shot_success = $script:FirstShotSuccess
            needed_fix         = $script:TasksNeededFix
            failed             = $script:TasksFailed
            first_shot_rate    = $(
                $t = $script:FirstShotSuccess + $script:TasksNeededFix + $script:TasksFailed
                if ($t -gt 0) { [math]::Round(($script:FirstShotSuccess / $t) * 100, 1) } else { 0 }
            )
        }
    }
    $snapshot | ConvertTo-Json -Depth 6 | Set-Content -Path $script:StatusStatePath -Encoding UTF8
}

function Get-SequenceValue {
    <#
    .SYNOPSIS Extract sequence value from a task title like 29A/30B (numeric * 100 + letter index).
    #>
    param([Parameter(Mandatory)][string]$Title)
    $match = [regex]::Match($Title, "\b(\d{2,3})([A-Z])\b")
    if (-not $match.Success) { return $null }
    $num = [int]$match.Groups[1].Value
    $letter = $match.Groups[2].Value
    $letterIndex = ([int][char]$letter) - ([int][char]'A') + 1
    return ($num * 100) + $letterIndex
}

function Get-OrderedTodoTasks {
    <#
    .SYNOPSIS Return todo tasks ordered by sequence (29A..29Z, 30A..), then by created_at.
    #>
    [CmdletBinding()]
    param([int]$Count = 1)

    $tasks = Get-VKTasks -Status "todo"
    if (-not $tasks) {
        Set-TodoBacklogCount -Count 0
        return @()
    }
    $tasks = @($tasks)
    Set-TodoBacklogCount -Count $tasks.Count

    $slopSeqTasks = [System.Collections.Generic.List[object]]::new()
    $slopOtherTasks = [System.Collections.Generic.List[object]]::new()
    $seqTasks = [System.Collections.Generic.List[object]]::new()
    $otherTasks = [System.Collections.Generic.List[object]]::new()
    $slopPattern = [regex]::new("(?i)\bslop(dev)?\b|slopes")
    foreach ($task in $tasks) {
        $seqValue = Get-SequenceValue -Title $task.title
        $title = if ($task.title) { $task.title } else { "" }
        $desc = if ($task.description) { $task.description } else { "" }
        $isSlop = $slopPattern.IsMatch($title) -or $slopPattern.IsMatch($desc)
        if ($isSlop) {
            if ($seqValue) {
                $slopSeqTasks.Add([pscustomobject]@{ task = $task; seq = $seqValue })
            }
            else {
                $slopOtherTasks.Add($task)
            }
        }
        else {
            if ($seqValue) {
                $seqTasks.Add([pscustomobject]@{ task = $task; seq = $seqValue })
            }
            else {
                $otherTasks.Add($task)
            }
        }
    }

    $slopSeqTasks = @($slopSeqTasks | Sort-Object -Property seq)
    $slopOtherTasks = @($slopOtherTasks | Sort-Object -Property created_at)
    $seqTasks = @($seqTasks | Sort-Object -Property seq)
    $otherTasks = @($otherTasks | Sort-Object -Property created_at)

    if ($slopSeqTasks.Count -eq 0 -and $slopOtherTasks.Count -eq 0 -and $seqTasks.Count -eq 0) {
        return @($otherTasks | Select-Object -First $Count)
    }

    $state = Get-OrchestratorState
    $lastSeq = $state.last_sequence_value

    $ordered = [System.Collections.Generic.List[object]]::new()
    if ($slopSeqTasks.Count -gt 0) {
        foreach ($item in $slopSeqTasks) { $ordered.Add($item.task) }
    }
    if ($slopOtherTasks.Count -gt 0) {
        foreach ($item in $slopOtherTasks) { $ordered.Add($item) }
    }
    if ($lastSeq) {
        $after = @($seqTasks | Where-Object { $_.seq -gt $lastSeq })
        $before = @($seqTasks | Where-Object { $_.seq -le $lastSeq })
        foreach ($item in $after) { $ordered.Add($item.task) }
        foreach ($item in $before) { $ordered.Add($item.task) }
    }
    else {
        foreach ($item in $seqTasks) { $ordered.Add($item.task) }
    }

    if ($otherTasks.Count -gt 0) {
        foreach ($item in $otherTasks) { $ordered.Add($item) }
    }

    return @($ordered | Select-Object -First $Count)
}

function Test-GithubCooldown {
    if ($script:GitHubCooldownUntil -and (Get-Date) -lt $script:GitHubCooldownUntil) {
        $remaining = [math]::Ceiling(($script:GitHubCooldownUntil - (Get-Date)).TotalSeconds)
        Write-Log "GitHub cooldown active ($remaining s remaining) — skipping GitHub calls" -Level "WARN"
        return $true
    }
    return $false
}

function Set-GithubCooldown {
    param([string]$Reason)
    $script:GitHubCooldownUntil = (Get-Date).AddSeconds($GitHubCooldownSec)
    Write-Log "GitHub rate limit detected — cooling down for ${GitHubCooldownSec}s" -Level "WARN"
    if ($Reason) {
        Write-Log "Reason: $Reason" -Level "WARN"
    }
}

function Test-GithubRateLimit {
    $err = Get-VKLastGithubError
    if ($err -and $err.type -eq "rate_limit") {
        Set-GithubCooldown -Reason $err.message
        return $true
    }
    return $false
}

function Test-KillCooldown {
    if ($script:KillCooldownUntil -and (Get-Date) -lt $script:KillCooldownUntil) {
        $remaining = [math]::Ceiling(($script:KillCooldownUntil - (Get-Date)).TotalMinutes)
        Write-Log "Kill cooldown active ($remaining min remaining) — throttling new attempts" -Level "WARN"
        return $true
    }
    return $false
}

function Get-EffectiveMaxParallel {
    $effective = $MaxParallel
    if (Test-KillCooldown) {
        if ($KillThrottleMaxParallel -gt 0) {
            $effective = [math]::Min($effective, $KillThrottleMaxParallel)
        }
    }
    return $effective
}

function Register-KilledAttempt {
    param(
        [Parameter(Mandatory)][string]$AttemptId,
        [Parameter(Mandatory)][hashtable]$Info
    )
    if ($KillBurstThreshold -le 0) { return }
    $now = Get-Date
    $window = [math]::Max(1, $KillBurstWindowMin)
    $recent = @($script:KillEvents | Where-Object { ($_ -is [datetime]) -and (($now - $_).TotalMinutes -le $window) })
    $recent += $now
    $script:KillEvents = $recent
    $count = $recent.Count
    Write-Log "Killed attempt detected ($($AttemptId.Substring(0,8))). Recent killed: $count in last ${window}m." -Level "WARN"
    if ($count -ge $KillBurstThreshold) {
        $cooldownMin = [math]::Max(1, $KillCooldownMin)
        $script:KillCooldownUntil = $now.AddMinutes($cooldownMin)
        $effective = if ($KillThrottleMaxParallel -gt 0) {
            [math]::Min($MaxParallel, $KillThrottleMaxParallel)
        }
        else {
            $MaxParallel
        }
        Write-Log "Kill burst detected — throttling parallelism to $effective for ${cooldownMin}m" -Level "WARN"
    }
}

function Get-AttemptLastActivity {
    <#
    .SYNOPSIS Estimate last activity time for an attempt using summary status.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][object]$Summary)

    if (-not $Summary) { return $null }
    if (-not $Summary.latest_process_completed_at) { return $null }
    try {
        return ([datetimeoffset]::Parse($Summary.latest_process_completed_at)).ToLocalTime().DateTime
    }
    catch {
        return $null
    }
}

function Get-TrackedStatusCounts {
    $counts = @{ running = 0; review = 0; error = 0; manual_review = 0; other = 0 }
    foreach ($item in $script:TrackedAttempts.Values) {
        switch ($item.status) {
            "running" { $counts.running++ }
            "review" { $counts.review++ }
            "error" { $counts.error++ }
            "manual_review" { $counts.manual_review++ }
            default { $counts.other++ }
        }
    }
    return $counts
}

function Clear-PendingFollowUp {
    param(
        [Parameter(Mandatory)][hashtable]$Info,
        [string]$Reason
    )
    if ($Info.pending_followup) {
        $Info.pending_followup = $null
        $Info.last_followup_reason = $Reason
    }
}

function Get-ActiveAgentCount {
    return @($script:TrackedAttempts.Values | Where-Object { $_.status -eq "running" }).Count
}

function Get-AvailableSlots {
    $activeCount = Get-ActiveAgentCount
    $effectiveMax = Get-EffectiveMaxParallel
    return [math]::Max(0, $effectiveMax - $activeCount)
}

function Try-SendFollowUp {
    <#
    .SYNOPSIS Send a follow-up only when the agent is idle and slots are available.
              Always uses the task continuation prompt format for agent clarity.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$AttemptId,
        [Parameter(Mandatory)][hashtable]$Info,
        [Parameter(Mandatory)][string]$Message,
        [string]$Reason
    )
    if ($Info.status -eq "running" -or $Info.last_process_status -eq "running") {
        Write-Log "Skipping follow-up for $($Info.branch): agent active" -Level "INFO"
        return $false
    }

    # Build continuation prompt with original task context instead of sending raw message
    $continuationPrompt = Build-TaskContinuationPrompt -Info $Info -ErrorMessage $Message -FailureCategory $Reason
    Write-Log "Built continuation prompt for $($Info.branch) ($($continuationPrompt.Length) chars)" -Level "INFO"

    $Info.last_followup_message = $continuationPrompt
    $Info.last_followup_reason = $Reason
    $Info.last_followup_at = Get-Date
    $slots = Get-AvailableSlots
    if ($slots -le 0) {
        Write-Log "Deferring follow-up for $($Info.branch): no available slots" -Level "WARN"
        $Info.pending_followup = @{ message = $continuationPrompt; reason = $Reason }
        return $false
    }
    if (-not $DryRun) {
        try {
            Send-VKWorkspaceFollowUp -WorkspaceId $AttemptId -Message $continuationPrompt | Out-Null
        }
        catch {
            Write-Log "Follow-up failed for $($Info.branch): $($_.Exception.Message)" -Level "WARN"
            $Info.pending_followup = @{ message = $continuationPrompt; reason = $Reason }
            return $false
        }
    }
    $Info.pending_followup = $null
    $Info.status = "running"
    $taskUrl = if ($Info.task_id) { Get-TaskUrl -TaskId $Info.task_id } else { $null }
    Add-RecentItem -ListName "FollowUpEvents" -Item @{
        task_id     = $Info.task_id
        task_title  = $Info.name
        attempt_id  = $AttemptId
        branch      = $Info.branch
        reason      = $Reason
        task_url    = $taskUrl
        occurred_at = (Get-Date).ToString("o")
    }
    return $true
}

function Get-PendingFollowUpAttempts {
    return @($script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.pending_followup })
}

function Process-PendingFollowUps {
    <#
    .SYNOPSIS Send queued follow-ups before starting new tasks.
    #>
    $slots = Get-AvailableSlots
    if ($slots -le 0) { return }

    $pending = Get-PendingFollowUpAttempts
    if (-not $pending -or @($pending).Count -eq 0) { return }

    foreach ($entry in $pending) {
        if ($slots -le 0) { break }
        $attemptId = $entry.Key
        $info = $entry.Value
        $message = $info.pending_followup.message
        $reason = $info.pending_followup.reason
        $useNewSession = $info.pending_followup.new_session
        $sent = if ($useNewSession) {
            Try-SendFollowUpNewSession -AttemptId $attemptId -Info $info -Message $message -Reason $reason
        }
        else {
            Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $message -Reason $reason
        }
        if ($sent) {
            $slots -= 1
        }
    }
}

function Test-ContextWindowError {
    <#
    .SYNOPSIS Detect context window errors from a summary payload.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][object]$Summary)
    $patterns = @(
        "ContextWindowExceeded",
        "context window",
        "ran out of room"
    )
    foreach ($prop in $Summary.PSObject.Properties) {
        $value = $prop.Value
        if ($value -is [string]) {
            foreach ($pattern in $patterns) {
                if ($value -match [regex]::Escape($pattern)) { return $true }
            }
        }
    }
    return $false
}

function Get-SummaryTextValues {
    <#
    .SYNOPSIS Collect string values from a summary payload (recursive, shallow).
             Guards against huge payloads (logs/stdout) that can exhaust memory.
    #>
    [CmdletBinding()]
    param(
        [AllowNull()][object]$Value,
        [int]$Depth = 0,
        [int]$MaxDepth = 4,
        [int]$MaxItems = 400,
        [int]$MaxChars = 50000,
        [int]$MaxItemChars = 8000
    )

    $state = @{
        results = [System.Collections.Generic.List[string]]::new()
        items   = 0
        chars   = 0
    }

    $skipKeys = @(
        "log", "logs", "stdout", "stderr", "output", "trace", "stacktrace", "messages"
    )

    $collect = {
        param([AllowNull()][object]$Val, [int]$Level)
        if ($null -eq $Val) { return }
        if ($state.items -ge $MaxItems -or $state.chars -ge $MaxChars) { return }
        if ($Level -gt $MaxDepth) { return }

        if ($Val -is [string]) {
            $text = $Val
            if ($text.Length -gt $MaxItemChars) {
                $text = $text.Substring(0, $MaxItemChars)
            }
            if ($text) {
                $state.results.Add($text)
                $state.items += 1
                $state.chars += $text.Length
            }
            return
        }

        if ($Val -is [System.Collections.IDictionary]) {
            foreach ($entry in $Val.GetEnumerator()) {
                if ($state.items -ge $MaxItems -or $state.chars -ge $MaxChars) { return }
                $keyName = if ($entry.Key) { $entry.Key.ToString().ToLowerInvariant() } else { "" }
                if ($keyName -and $skipKeys -contains $keyName) { continue }
                & $collect $entry.Value ($Level + 1)
            }
            return
        }

        if ($Val -is [System.Collections.IEnumerable] -and -not ($Val -is [string])) {
            foreach ($v in $Val) {
                if ($state.items -ge $MaxItems -or $state.chars -ge $MaxChars) { return }
                & $collect $v ($Level + 1)
            }
            return
        }

        if ($Val.PSObject -and $Val.PSObject.Properties) {
            foreach ($prop in $Val.PSObject.Properties) {
                if ($state.items -ge $MaxItems -or $state.chars -ge $MaxChars) { return }
                $propName = if ($prop.Name) { $prop.Name.ToLowerInvariant() } else { "" }
                if ($propName -and $skipKeys -contains $propName) { continue }
                & $collect $prop.Value ($Level + 1)
            }
        }
    }

    & $collect $Value $Depth
    return @($state.results)
}

function Get-AttemptFailureCategory {
    <#
    .SYNOPSIS Classify failure causes from attempt summary payloads.
    #>
    [CmdletBinding()]
    param(
        [object]$Summary,
        [hashtable]$Info
    )
    $latestStatus = $null
    if ($Summary -and $Summary.latest_process_status) {
        $latestStatus = $Summary.latest_process_status
    }
    elseif ($Info -and $Info.last_process_status) {
        $latestStatus = $Info.last_process_status
    }

    if ($Summary -is [System.DBNull]) {
        $Summary = $null
    }

    $textValues = @()
    try {
        $summaryTextSource = $Summary
        if ($summaryTextSource -is [System.DBNull]) { $summaryTextSource = $null }
        if ($null -ne $summaryTextSource) {
            $textValues = Get-SummaryTextValues -Value $summaryTextSource -ErrorAction Stop
        }
    }
    catch {
        $textValues = @()
    }
    $combined = ($textValues | Where-Object { $_ -is [string] -and $_.Trim() }) -join "`n"
    $lower = $combined.ToLowerInvariant()

    $apiKeyPatterns = @(
        "api key",
        "api_key",
        "openai_api_key",
        "missing api key",
        "invalid_api_key",
        "unauthorized",
        "authentication",
        "401",
        "403"
    )
    foreach ($pattern in $apiKeyPatterns) {
        if ($lower -match [regex]::Escape($pattern)) {
            return @{ category = "api_key"; status = $latestStatus; detail = $pattern }
        }
    }

    if ($lower -match "context window" -or $lower -match "contextwindowexceeded" -or $lower -match "ran out of room") {
        return @{ category = "context_window"; status = $latestStatus; detail = "context window exceeded" }
    }

    # ─── Timeout / stale detection ───────────────────────────────────────────
    $timeoutPatterns = @(
        "reconnecting",
        "reconnecting\.\.\.",
        "timed? ?out",
        "deadline exceeded",
        "connection reset",
        "hard timeout",
        "idle timeout",
        "operation timed out"
    )
    foreach ($tp in $timeoutPatterns) {
        if ($lower -match $tp) {
            return @{ category = "timeout"; status = $latestStatus; detail = $tp }
        }
    }
    if ($latestStatus -eq "timeout") {
        return @{ category = "timeout"; status = $latestStatus; detail = "process timeout" }
    }

    # ─── Rate limit detection ────────────────────────────────────────────────
    $rateLimitPatterns = @(
        "rate.?limit",
        "too many requests",
        "429",
        "oops.?you can.?t create more requests",
        "please wait",
        "quota exceeded",
        "throttl",
        "capacity"
    )
    foreach ($rp in $rateLimitPatterns) {
        if ($lower -match $rp) {
            return @{ category = "rate_limit"; status = $latestStatus; detail = $rp }
        }
    }

    # ─── Reconnect loop detection ────────────────────────────────────────────
    if ($lower -match "reconnecting\.\.\.\s*\d+/\d+") {
        return @{ category = "reconnect_loop"; status = $latestStatus; detail = "reconnecting loop" }
    }

    if ($latestStatus -in @("failed", "killed", "crashed", "error", "aborted")) {
        return @{ category = "agent_failed"; status = $latestStatus; detail = $latestStatus }
    }

    return @{ category = "unknown"; status = $latestStatus; detail = $null }
}

function Get-TaskRetryKey {
    param(
        [string]$TaskId,
        [string]$Category
    )
    if (-not $TaskId) { return $Category }
    return "${TaskId}:$Category"
}

function Increment-TaskRetryCount {
    param(
        [string]$TaskId,
        [string]$Category
    )
    $key = Get-TaskRetryKey -TaskId $TaskId -Category $Category
    $current = 0
    if ($script:TaskRetryCounts.ContainsKey($key)) {
        $current = [int]$script:TaskRetryCounts[$key]
    }
    $current += 1
    $script:TaskRetryCounts[$key] = $current
    return $current
}

function Build-TaskContinuationPrompt {
    <#
    .SYNOPSIS Build a continuation prompt that includes the original task context.
              Used when restarting agents so they don't lose track of what they were doing.
              Format: {{ORIGINAL TASK SUMMARY}} + "PLEASE CONTINUE" directive.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][hashtable]$Info,
        [string]$ErrorMessage,
        [string]$FailureCategory
    )

    # Fetch full task details from VK API
    $taskDescription = $null
    if ($Info.task_id) {
        try {
            $task = Get-VKTask -TaskId $Info.task_id
            if ($task -and $task.description) {
                $taskDescription = $task.description
            }
        }
        catch {
            Write-Log "Could not fetch task details for continuation prompt: $($_.Exception.Message)" -Level "WARN"
        }
    }

    $taskTitle = if ($Info.name) { $Info.name } else { "Unknown task" }
    $branch = if ($Info.branch) { $Info.branch } else { "unknown" }

    # Build the original task summary section
    $prompt = @"
## Original Task

**Title:** $taskTitle
**Branch:** ``$branch``
"@

    if ($taskDescription) {
        $prompt += "`n`n**Task Description:**`n$taskDescription"
    }

    $prompt += @"


---

PLEASE CONTINUE ON THE ABOVE TASK. THE AGENT SESSION ENDED UNEXPECTEDLY.

IF YOU HAVE ALREADY COMPLETED THE TASK, ENSURE IT IS AVAILABLE IN THE UPSTREAM BRANCH (committed, pushed, and PR created).

### Instructions
1. Check ``git status`` and ``git log --oneline -5`` to see what was already done
2. Review the branch for any partial work from the previous session
3. **Continue from where the previous agent left off** — do NOT start over
4. If the task is already complete (all changes committed and pushed), verify and confirm
5. Ensure all tests pass and the build succeeds before pushing
6. Create a PR if one doesn't exist yet (``gh pr list --head $branch``)

### Critical Rules
- Do NOT redo work that is already committed on this branch
- Focus ONLY on completing the remaining work for this task
- Follow AGENTS.md coding standards
- All tests must pass with 0 warnings before committing
"@

    return $prompt
}

function Try-SendFollowUpNewSession {
    <#
    .SYNOPSIS Create a new session and send a follow-up message.
              Automatically includes original task context for agent continuity.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$AttemptId,
        [Parameter(Mandatory)][hashtable]$Info,
        [Parameter(Mandatory)][string]$Message,
        [string]$Reason
    )
    if ($Info.status -eq "running" -or $Info.last_process_status -eq "running") {
        Write-Log "Skipping new-session follow-up for $($Info.branch): agent active" -Level "INFO"
        return $false
    }

    # Build a rich continuation prompt with original task context
    $failureCategory = $null
    if ($Reason -match "retry|failed|error|context|missing|killed") {
        $failureCategory = $Reason -replace "_", " "
    }
    $continuationPrompt = Build-TaskContinuationPrompt -Info $Info -ErrorMessage $Message -FailureCategory $failureCategory
    Write-Log "Built continuation prompt for $($Info.branch) ($($continuationPrompt.Length) chars)" -Level "INFO"

    $Info.last_followup_message = $continuationPrompt
    $Info.last_followup_reason = $Reason
    $Info.last_followup_at = Get-Date
    $slots = Get-AvailableSlots
    if ($slots -le 0) {
        Write-Log "Deferring new-session follow-up for $($Info.branch): no available slots" -Level "WARN"
        $Info.pending_followup = @{ message = $continuationPrompt; reason = $Reason; new_session = $true }
        return $false
    }
    if (-not $DryRun) {
        $session = New-VKSession -WorkspaceId $AttemptId
        if (-not $session) {
            Write-Log "Failed to start new session for $($Info.branch)" -Level "WARN"
            $Info.pending_followup = @{ message = $continuationPrompt; reason = $Reason; new_session = $true }
            return $false
        }
        $profile = if ($session.executor) { Get-ExecutorProfileForSession -Executor $session.executor } else { Get-CurrentExecutorProfile }
        try {
            Send-VKSessionFollowUp -SessionId $session.id -Message $continuationPrompt -ExecutorProfile $profile | Out-Null
        }
        catch {
            Write-Log "New-session follow-up failed for $($Info.branch): $($_.Exception.Message)" -Level "WARN"
            $Info.pending_followup = @{ message = $continuationPrompt; reason = $Reason; new_session = $true }
            return $false
        }
    }
    $Info.pending_followup = $null
    $Info.status = "running"
    $taskUrl = if ($Info.task_id) { Get-TaskUrl -TaskId $Info.task_id } else { $null }
    Add-RecentItem -ListName "FollowUpEvents" -Item @{
        task_id     = $Info.task_id
        task_title  = $Info.name
        attempt_id  = $AttemptId
        branch      = $Info.branch
        reason      = $Reason
        task_url    = $taskUrl
        occurred_at = (Get-Date).ToString("o")
    }
    return $true
}

function Try-RecoverContextWindow {
    <#
    .SYNOPSIS Create a new session and re-send with original task context when context is exhausted.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$AttemptId,
        [Parameter(Mandatory)][hashtable]$Info
    )
    if ($Info.status -eq "running" -or $Info.last_process_status -eq "running") {
        return $false
    }

    $slots = Get-AvailableSlots
    if ($slots -le 0) {
        Write-Log "Deferring context recovery for $($Info.branch): no available slots" -Level "WARN"
        $Info.context_recovery_pending = $true
        return $false
    }

    $session = New-VKSession -WorkspaceId $AttemptId
    if (-not $session) {
        Write-Log "Failed to create new session for $($Info.branch)" -Level "WARN"
        return $false
    }

    $Info.context_recovery_pending = $false
    $Info.context_recovery_attempted_at = Get-Date
    $Info.status = "running"

    # Build continuation prompt with original task context (not just the old follow-up)
    $continuationMsg = Build-TaskContinuationPrompt -Info $Info `
        -ErrorMessage "Context window was exhausted in the previous session. A fresh session has been started." `
        -FailureCategory "context window exceeded"

    $profile = if ($session.executor) {
        Get-ExecutorProfileForSession -Executor $session.executor
    }
    else {
        Get-ExecutorProfileForSession -Executor "CODEX"
    }
    $sent = Send-VKSessionFollowUp -SessionId $session.id -Message $continuationMsg -ExecutorProfile $profile
    if (-not $sent) {
        Write-Log "Follow-up resend failed for $($Info.branch)" -Level "WARN"
        return $false
    }

    Write-Log "Context recovery started for $($Info.branch) via new session" -Level "INFO"
    return $true
}

function Merge-PRWithFallback {
    <#
    .SYNOPSIS Merge a PR, optionally using admin override and retrying when out of date.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [switch]$ForceAdmin
    )
    $merged = Merge-PR -PRNumber $PRNumber -Admin:$ForceAdmin
    if (Test-GithubRateLimit) { return @{ merged = $false; reason = "rate_limit"; used_admin = $ForceAdmin } }
    if ($merged) { return @{ merged = $true; reason = $null; used_admin = $ForceAdmin } }

    $mergeErr = Get-VKLastGithubError
    $reason = if ($mergeErr -and $mergeErr.message) { $mergeErr.message } else { "Unknown merge error" }
    if ($reason -match "not up to date|not mergeable") {
        Write-Log "Retrying merge with admin override for PR #$PRNumber" -Level "WARN"
        $mergedAdmin = Merge-PR -PRNumber $PRNumber -Admin
        if (Test-GithubRateLimit) { return @{ merged = $false; reason = "rate_limit"; used_admin = $true } }
        if ($mergedAdmin) { return @{ merged = $true; reason = $null; used_admin = $true } }
        return @{ merged = $false; reason = $reason; used_admin = $true }
    }

    return @{ merged = $false; reason = $reason; used_admin = $ForceAdmin }
}

function Invoke-DirectRebase {
    <#
    .SYNOPSIS Attempt a direct git rebase of a PR branch onto main.
              Falls back to VK rebase if direct fails.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Branch,
        [string]$BaseBranch = "main",
        [string]$AttemptId
    )

    Write-Log "Attempting direct rebase of $Branch onto $BaseBranch" -Level "ACTION"

    # Fetch latest
    $fetchOut = git fetch origin $BaseBranch 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Log "git fetch failed: $fetchOut" -Level "WARN"
        if ($AttemptId) {
            return Rebase-VKAttempt -AttemptId $AttemptId
        }
        return $false
    }

    $fetchBranch = git fetch origin "${Branch}:refs/remotes/origin/${Branch}" 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Log "git fetch branch failed: $fetchBranch" -Level "WARN"
        if ($AttemptId) {
            return Rebase-VKAttempt -AttemptId $AttemptId
        }
        return $false
    }

    # Create a temp local branch, rebase, force-push
    $tempBranch = "rebase-temp-$(Get-Random)"
    try {
        git checkout -b $tempBranch "origin/$Branch" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "checkout failed" }

        $rebaseOut = git rebase "origin/$BaseBranch" 2>&1
        if ($LASTEXITCODE -ne 0) {
            git rebase --abort 2>&1 | Out-Null
            throw "rebase conflict: $rebaseOut"
        }

        $pushOut = git push origin "${tempBranch}:${Branch}" --force-with-lease 2>&1
        if ($LASTEXITCODE -ne 0) { throw "push failed: $pushOut" }

        Write-Log "Direct rebase succeeded for $Branch" -Level "OK"
        return $true
    }
    catch {
        Write-Log "Direct rebase failed: $($_.Exception.Message) — falling back to VK rebase" -Level "WARN"
        if ($AttemptId) {
            return Rebase-VKAttempt -AttemptId $AttemptId
        }
        return $false
    }
    finally {
        # Clean up temp branch
        git checkout - 2>&1 | Out-Null
        git branch -D $tempBranch 2>&1 | Out-Null
    }
}

function Get-MergeFailureInfo {
    <#
    .SYNOPSIS Classify merge failures with PR metadata for clearer diagnostics.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [string]$Reason
    )
    $details = Get-PRDetails -PRNumber $PRNumber
    if (Test-GithubRateLimit) {
        return @{ category = "rate_limit"; action = "wait"; summary = "rate limit"; reason = $Reason }
    }

    $mergeState = if ($details -and $details.mergeStateStatus) { $details.mergeStateStatus } else { "UNKNOWN" }
    $reviewDecision = if ($details -and $details.reviewDecision) { $details.reviewDecision } else { "UNKNOWN" }
    $isDraft = if ($details) { [bool]$details.isDraft } else { $false }

    $category = "unknown"
    $action = "review"

    if ($mergeState -in @("DIRTY", "CONFLICTING")) {
        $category = "conflict"
        $action = "resolve_conflicts"
    }
    elseif ($mergeState -eq "BEHIND") {
        $category = "behind"
        $action = "rebase"
    }
    elseif ($mergeState -eq "BLOCKED") {
        $category = "blocked"
        $action = "checks_or_reviews"
    }
    elseif ($isDraft) {
        $category = "draft"
        $action = "mark_ready"
    }

    if ($reviewDecision -in @("REVIEW_REQUIRED", "CHANGES_REQUESTED")) {
        $category = "review_required"
        $action = "request_review"
    }

    if ($Reason -match "policy prohibits|base branch policy") {
        $category = "policy"
        $action = "admin_or_reviews"
    }
    elseif ($Reason -match "Base branch was modified") {
        $category = "base_changed"
        $action = "retry"
    }
    elseif ($Reason -match "not up to date|not mergeable") {
        $category = "behind"
        $action = "rebase"
    }
    elseif ($Reason -match "rate limit|abuse detection|secondary rate limit") {
        $category = "rate_limit"
        $action = "wait"
    }

    $summary = "state=$mergeState review=$reviewDecision category=$category action=$action"
    return @{ category = $category; action = $action; summary = $summary; reason = $Reason; merge_state = $mergeState; review_decision = $reviewDecision }
}

function Get-MergeCandidates {
    <#
    .SYNOPSIS Collect PRs eligible for merge (idle agent, CI passing, age >= CiWaitMin).
    #>
    $candidates = @()
    $candidateNumbers = @{}
    foreach ($entry in $script:TrackedAttempts.GetEnumerator()) {
        $attemptId = $entry.Key
        $info = $entry.Value

        if ($info.status -in @("merged", "done", "running")) { continue }
        if ($info.last_process_status -eq "running") { continue }
        if (-not $info.branch) { continue }

        $pr = Get-PRForBranch -Branch $info.branch
        if (Test-GithubRateLimit) { return @() }
        if (-not $pr) { continue }

        if ($pr.state -eq "MERGED" -or $pr.mergedAt) {
            Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
            continue
        }

        $prDetails = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return @() }
        $mergeState = if ($prDetails -and $prDetails.mergeStateStatus) { $prDetails.mergeStateStatus } else { "UNKNOWN" }
        if ($mergeState -in @("DIRTY", "CONFLICTING")) { continue }

        $baseBranch = if ($prDetails -and $prDetails.baseRefName) { $prDetails.baseRefName } else { $script:VK_TARGET_BRANCH }
        if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }
        $checkStatus = Get-PRRequiredCheckStatus -PRNumber $pr.number -BaseBranch $baseBranch
        if (Test-GithubRateLimit) { return @() }
        if ($checkStatus -ne "passing") { continue }

        $candidates += [pscustomobject]@{
            attempt_id = $attemptId
            task_id    = $info.task_id
            branch     = $info.branch
            pr_number  = $pr.number
            created_at = $pr.createdAt
        }
        $candidateNumbers[$pr.number] = $true
    }

    $openPrs = Get-OpenPullRequests
    if (Test-GithubRateLimit) { return @() }
    foreach ($pr in $openPrs) {
        if (-not $pr.author -or -not (Test-IsCopilotAuthor -Author $pr.author)) { continue }
        if ($candidateNumbers.ContainsKey($pr.number)) { continue }

        $prDetails = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return @() }
        $mergeState = if ($prDetails -and $prDetails.mergeStateStatus) { $prDetails.mergeStateStatus } else { "UNKNOWN" }
        if ($mergeState -in @("DIRTY", "CONFLICTING")) { continue }

        $baseBranch = if ($pr.baseRefName) { $pr.baseRefName } else { $script:VK_TARGET_BRANCH }
        if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }
        $checkStatus = Get-PRRequiredCheckStatus -PRNumber $pr.number -BaseBranch $baseBranch
        if (Test-GithubRateLimit) { return @() }
        if ($checkStatus -ne "passing") { continue }

        if ($pr.isDraft) {
            $null = Mark-PRReady -PRNumber $pr.number
        }

        $candidates += [pscustomobject]@{
            attempt_id = $null
            task_id    = $null
            branch     = $pr.headRefName
            pr_number  = $pr.number
            created_at = $pr.createdAt
        }
        $candidateNumbers[$pr.number] = $true
    }
    $candidates = $candidates | Sort-Object -Property {
        if ($_.created_at) {
            try { [datetime]::Parse($_.created_at) } catch { [datetime]::MaxValue }
        }
        else { [datetime]::MaxValue }
    }
    return $candidates
}

# ─── Core Orchestration ──────────────────────────────────────────────────────

function Sync-TrackedAttempts {
    <#
    .SYNOPSIS Refresh tracked attempts from vibe-kanban API + discover new active ones.
    #>
    $apiAttempts = Get-VKAttempts -ActiveOnly
    $summaries = Get-VKAttemptSummaries -Archived:$false
    if (-not $apiAttempts) { return }

    $summaryMap = @{}
    foreach ($s in $summaries) {
        if ($s.workspace_id) { $summaryMap[$s.workspace_id] = $s }
    }
    $script:AttemptSummaries = $summaryMap

    foreach ($a in $apiAttempts) {
        if (-not $a.branch) { continue }
        if (-not $script:TrackedAttempts.ContainsKey($a.id)) {
            # Determine initial status from summary instead of blindly assuming "running"
            $initialSummary = $summaryMap[$a.id]
            $initialStatus = "running"
            if ($initialSummary -and $initialSummary.latest_process_status) {
                $vkProcStatus = $initialSummary.latest_process_status
                if ($vkProcStatus -in @("completed", "failed", "killed", "stopped", "timeout")) {
                    $initialStatus = if ($vkProcStatus -in @("failed", "killed", "timeout")) { "error" } else { "review" }
                }
            }

            # Newly discovered active attempt
            $script:TrackedAttempts[$a.id] = @{
                task_id                       = $a.task_id
                branch                        = $a.branch
                pr_number                     = $null
                status                        = $initialStatus
                name                          = $a.name
                executor                      = $null
                executor_provider             = $null
                health_reported               = $false
                updated_at                    = $a.updated_at
                container_ref                 = $a.container_ref
                ci_notified                   = $false
                conflict_notified             = $false
                rebase_requested              = $false
                idle_detected_at              = $null
                review_marked                 = ($initialStatus -ne "running")
                error_notified                = ($initialStatus -eq "error")
                push_notified                 = $false
                merge_failed_notified         = $false
                merge_failures_total          = 0
                merge_failure_cycles          = 0
                last_merge_notify_at          = $null
                last_merge_failure_reason     = $null
                last_merge_failure_category   = $null
                last_merge_failure_at         = $null
                manual_review_notified        = $false
                last_process_status           = if ($initialSummary) { $initialSummary.latest_process_status } else { $null }
                last_process_completed_at     = if ($initialSummary) { $initialSummary.latest_process_completed_at } else { $null }
                pending_followup              = $null
                last_followup_message         = $null
                last_followup_reason          = $null
                last_followup_at              = $null
                context_recovery_pending      = $false
                context_recovery_attempted_at = $null
                copilot_fix_requested         = $false
                copilot_fix_requested_at      = $null
                copilot_fix_pr_number         = $null
                copilot_fix_merged            = $false
                copilot_fix_merged_at         = $null
                kill_recorded                 = $false
            }
            $statusTag = if ($initialStatus -ne "running") { " [$initialStatus]" } else { "" }
            Write-Log "Tracking new attempt: $($a.id.Substring(0,8)) → $($a.branch)$statusTag" -Level "INFO"
        }
        else {
            $script:TrackedAttempts[$a.id].updated_at = $a.updated_at
            $script:TrackedAttempts[$a.id].container_ref = $a.container_ref
        }

        $summary = $summaryMap[$a.id]
        if ($summary) {
            $tracked = $script:TrackedAttempts[$a.id]
            if ($tracked -is [hashtable] -and -not $tracked.ContainsKey("kill_recorded")) {
                $tracked.kill_recorded = $false
            }
            $summaryStatus = $summary.latest_process_status
            $tracked.last_process_status = $summaryStatus
            $tracked.last_process_completed_at = $summary.latest_process_completed_at

            if ([string]::IsNullOrWhiteSpace($summaryStatus)) {
                # VK has not reported a process status yet; keep current state.
                continue
            }

            if ($summaryStatus -eq "running") {
                # Agent is active again; treat as running to avoid review actions
                $tracked.status = "running"
                $tracked.review_marked = $false
                $tracked.error_notified = $false
                $tracked.ci_notified = $false
                $tracked.pending_followup = $null
            }
            elseif ($summaryStatus -in @("completed", "failed", "killed", "stopped", "timeout")) {
                # Any terminal status — workspace process is no longer running
                $effectiveStatus = $summaryStatus
                if ($tracked.status -eq "running") {
                    $tracked.status = "review"
                }
                if (-not $tracked.review_marked) {
                    Write-Log "Attempt $($a.id.Substring(0,8)) finished ($effectiveStatus) — marking review" -Level "INFO"
                    if (-not $DryRun) {
                        Update-VKTaskStatus -TaskId $tracked.task_id -Status "inreview" | Out-Null
                    }
                    $tracked.review_marked = $true
                }
                if ($effectiveStatus -eq "killed" -and -not $tracked.kill_recorded) {
                    Register-KilledAttempt -AttemptId $a.id -Info $tracked
                    $tracked.kill_recorded = $true
                }
                if ($effectiveStatus -in @("failed", "killed", "timeout") -and -not $tracked.error_notified) {
                    Write-Log "Attempt $($a.id.Substring(0,8)) $effectiveStatus in workspace — requires agent attention" -Level "WARN"
                    $tracked.error_notified = $true
                    $tracked.status = "error"
                }

                # ─── Report executor failure to health system ────────────────
                $execProvider = $tracked.executor_provider
                if (-not $execProvider -and $tracked.executor) {
                    # Legacy attempts without executor_provider — derive from executor name
                    $execProvider = Get-ExecutorProviderForAttempt -Executor $tracked.executor -Variant $null
                }
                if ($execProvider -and -not $tracked.health_reported) {
                    if ($effectiveStatus -in @("failed", "killed", "timeout")) {
                        $failCategory = Get-AttemptFailureCategory -Summary $summary -Info $tracked
                        $failType = switch ($failCategory.category) {
                            "timeout" { "timeout" }
                            "rate_limit" { "rate_limit" }
                            "reconnect_loop" { "reconnect_loop" }
                            default { "error" }
                        }
                        $null = Report-ExecutorFailure -Provider $execProvider -FailureType $failType
                        Write-Log "Health: reported $failType for $execProvider (attempt $($a.id.Substring(0,8)))" -Level "INFO"
                    }
                    else {
                        # completed/stopped — success
                        Report-ExecutorSuccess -Provider $execProvider
                    }
                    Decrement-ExecutorActiveCount -Provider $execProvider
                    $tracked.health_reported = $true
                }

                if ($effectiveStatus -in @("failed", "killed") -and (Test-ContextWindowError -Summary $summary)) {
                    $shouldRecover = $true
                    if ($tracked.context_recovery_attempted_at) {
                        $sinceAttempt = ((Get-Date) - $tracked.context_recovery_attempted_at).TotalMinutes
                        if ($sinceAttempt -lt 10) { $shouldRecover = $false }
                    }
                    if ($shouldRecover) {
                        $null = Try-RecoverContextWindow -AttemptId $a.id -Info $tracked
                    }
                }
            }
            else {
                # Unknown status (new VK status) — log for visibility
                Write-Log "Attempt $($a.id.Substring(0,8)) has unknown process status '$summaryStatus' — treating as terminal" -Level "WARN"
                if ($tracked.status -eq "running") {
                    $tracked.status = "review"
                    $tracked.review_marked = $true
                    if (-not $DryRun) {
                        Update-VKTaskStatus -TaskId $tracked.task_id -Status "inreview" | Out-Null
                    }
                }
            }
        }
    }

    # Stale-running detection: catch workspaces stuck as "running" when they're actually dead.
    # Two scenarios:
    #   1. VK says "running" but updated_at hasn't changed in StaleRunningTimeoutMin
    #   2. No summary data at all and updated_at is stale
    foreach ($a in $apiAttempts) {
        $tracked = $script:TrackedAttempts[$a.id]
        if (-not $tracked) { continue }
        if ($tracked.status -ne "running") { continue }

        # Use updated_at from VK API as heartbeat
        $lastHeartbeat = $null
        if ($tracked.updated_at) {
            try { $lastHeartbeat = [datetimeoffset]::Parse($tracked.updated_at).ToLocalTime().DateTime }
            catch { }
        }
        # Also consider latest_process_completed_at as secondary signal
        if (-not $lastHeartbeat -and $tracked.last_process_completed_at) {
            try { $lastHeartbeat = [datetimeoffset]::Parse($tracked.last_process_completed_at).ToLocalTime().DateTime }
            catch { }
        }

        if (-not $lastHeartbeat) { continue }

        $staleMins = ((Get-Date) - $lastHeartbeat).TotalMinutes
        if ($staleMins -ge $StaleRunningTimeoutMin) {
            $vkStatus = $tracked.last_process_status
            Write-Log "Attempt $($a.id.Substring(0,8)) stale-running for $([math]::Round($staleMins))m (VK status: $vkStatus) — marking as error" -Level "WARN"
            $tracked.status = "error"
            $tracked.error_notified = $true
            if (-not $tracked.review_marked) {
                $tracked.review_marked = $true
                if (-not $DryRun) {
                    if ($tracked.task_id) {
                        Update-VKTaskStatus -TaskId $tracked.task_id -Status "inreview" | Out-Null
                    }
                    else {
                        Write-Log "Attempt $($a.id.Substring(0,8)) missing task_id — skipping inreview update" -Level "WARN"
                    }
                }
            }
        }
    }

    # De-duplicate active attempts for the same task (keep most recent)
    $attemptsByTask = @{}
    foreach ($a in $apiAttempts) {
        if (-not $a.task_id) { continue }
        if (-not $attemptsByTask.ContainsKey($a.task_id)) {
            $attemptsByTask[$a.task_id] = @()
        }
        $attemptsByTask[$a.task_id] += $a
    }

    foreach ($taskId in $attemptsByTask.Keys) {
        $group = @($attemptsByTask[$taskId])
        if ($group.Count -le 1) { continue }

        $ordered = $group | Sort-Object -Property @{
            Expression = {
                try { [datetime]::Parse($_.updated_at) } catch { Get-Date 0 }
            }
            Descending = $true
        }
        $keeper = $ordered[0]
        $duplicates = $ordered | Select-Object -Skip 1

        foreach ($dup in $duplicates) {
            Write-Log "Duplicate attempt for task $($taskId.Substring(0,8)) — archiving $($dup.id.Substring(0,8)) (keeping $($keeper.id.Substring(0,8)))" -Level "WARN"
            if (-not $DryRun) {
                Archive-VKAttempt -AttemptId $dup.id | Out-Null
            }
            if ($script:TrackedAttempts.ContainsKey($dup.id)) {
                $tracked = $script:TrackedAttempts[$dup.id]
                if ($tracked) {
                    Clear-PendingFollowUp -Info $tracked -Reason "duplicate_attempt"
                }
                $script:TrackedAttempts.Remove($dup.id)
            }
        }
    }

    # Remove attempts that disappeared from active list (archived)
    $activeIds = @($apiAttempts | ForEach-Object { $_.id })
    $toCheck = @($script:TrackedAttempts.Keys | Where-Object { $_ -notin $activeIds })
    foreach ($id in $toCheck) {
        $tracked = $script:TrackedAttempts[$id]
        if ($tracked) {
            Clear-PendingFollowUp -Info $tracked -Reason "attempt_archived"
        }
        Write-Log "Attempt $($id.Substring(0,8)) archived — removing from tracking" -Level "INFO"
        $script:TrackedAttempts.Remove($id)
    }

    $summaryKeys = @($script:AttemptSummaries.Keys)
    foreach ($id in $summaryKeys) {
        if ($id -notin $activeIds) {
            $script:AttemptSummaries.Remove($id)
        }
    }

    # Handle stale/idle attempts (no PR, idle beyond threshold) based on summaries
    foreach ($a in $apiAttempts) {
        $tracked = $script:TrackedAttempts[$a.id]
        if (-not $tracked) { continue }
        if ($tracked.status -ne "review") { continue }

        $summary = $summaryMap[$a.id]
        if (-not $summary) { continue }
        if ($summary.latest_process_status -ne "completed") { continue }

        $lastUpdate = Get-AttemptLastActivity -Summary $summary
        if (-not $lastUpdate) { continue }

        $idleMinutes = ((Get-Date) - $lastUpdate).TotalMinutes
        if ($idleMinutes -lt $IdleTimeoutMin) {
            $tracked.idle_detected_at = $null
            continue
        }

        if (-not $tracked.idle_detected_at) {
            # Anchor confirm timing to when the attempt crossed IdleTimeoutMin.
            $tracked.idle_detected_at = $lastUpdate.AddMinutes($IdleTimeoutMin)
        }

        $confirmMinutes = ((Get-Date) - $tracked.idle_detected_at).TotalMinutes
        if ($confirmMinutes -lt $IdleConfirmMin) {
            $remaining = [math]::Ceiling($IdleConfirmMin - $confirmMinutes)
            Write-Log "Attempt $($a.id.Substring(0,8)) idle ${IdleTimeoutMin}m+ — will confirm in ${remaining}m" -Level "WARN"
            continue
        }

        # Do not archive idle attempts; they are already in review
        Write-Log "Attempt $($a.id.Substring(0,8)) idle ${IdleTimeoutMin}m+ — awaiting PR" -Level "WARN"
    }

    foreach ($a in $apiAttempts) {
        $tracked = $script:TrackedAttempts[$a.id]
        if (-not $tracked) { continue }
        if (-not $tracked.context_recovery_pending) { continue }
        $summary = $summaryMap[$a.id]
        if (-not $summary) { continue }
        if ($summary.latest_process_status -eq "running") { continue }
        $null = Try-RecoverContextWindow -AttemptId $a.id -Info $tracked
    }
}

function Process-CompletedAttempts {
    <#
    .SYNOPSIS Check tracked attempts for PR status and handle merging.
    #>
    if (Test-GithubCooldown) { return }
    $processed = @()

    foreach ($entry in $script:TrackedAttempts.GetEnumerator()) {
        $attemptId = $entry.Key
        $info = $entry.Value

        # Skip already-completed or manual review
        if ($info.status -in @("merged", "done", "manual_review")) { continue }

        if ($info.pending_followup) {
            $pending = $info.pending_followup
            if ($pending.new_session) {
                $null = Try-SendFollowUpNewSession -AttemptId $attemptId -Info $info -Message $pending.message -Reason $pending.reason
            }
            else {
                $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $pending.message -Reason $pending.reason
            }
        }

        $branch = $info.branch
        if (-not $branch) { continue }

        if ($info.status -eq "running" -or $info.last_process_status -eq "running") {
            continue
        }

        # Check for PR
        $pr = Get-PRForBranch -Branch $branch
        if (Test-GithubRateLimit) { return }
        if (-not $pr) {
            if ($info.status -in @("review", "error")) {
                if (-not (Test-RemoteBranchExists -Branch $branch)) {
                    Write-Log "No remote branch for $branch — agent must push before PR" -Level "WARN"
                    $summary = $script:AttemptSummaries[$attemptId]
                    $failure = Get-AttemptFailureCategory -Summary $summary -Info $info
                    $recentFollowup = $false
                    if ($info.last_followup_at) {
                        $recentFollowup = (((Get-Date) - $info.last_followup_at).TotalMinutes -lt 10)
                    }

                    if (-not $recentFollowup) {
                        if ($failure.category -in @("api_key", "agent_failed")) {
                            $count = Increment-TaskRetryCount -TaskId $info.task_id -Category $failure.category
                            if ($count -ge 2) {
                                $msg = "Agent failure detected ($($failure.category)), retry $count — starting fresh session."
                                $null = Try-SendFollowUpNewSession -AttemptId $attemptId -Info $info -Message $msg -Reason "retry_new_session"
                            }
                            else {
                                $msg = "Agent failure detected ($($failure.category)), retry $count."
                                $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "retry_task"
                            }
                            $info.push_notified = $true
                            continue
                        }

                        $count = Increment-TaskRetryCount -TaskId $info.task_id -Category "missing_branch"
                        if ($count -ge 2) {
                            $msg = "Remote branch $branch not found, retry $count — starting fresh session."
                            $null = Try-SendFollowUpNewSession -AttemptId $attemptId -Info $info -Message $msg -Reason "missing_branch_new_session"
                        }
                        else {
                            $msg = "Remote branch $branch not found, retry $count."
                            $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "missing_branch"
                        }
                        $info.push_notified = $true
                    }
                    continue
                }
                Write-Log "No PR for $branch — creating PR" -Level "ACTION"
                if (-not $DryRun) {
                    $title = if ($info.name) { $info.name } else { "Automated task PR" }
                    $created = Create-PRForBranchSafe -Branch $branch -Title $title -Body "Automated PR created by ve-orchestrator"
                    if (Test-GithubRateLimit) { return }
                    if ($created -eq "no_commits") {
                        Write-Log "Branch $branch has no commits vs base — marking for manual review" -Level "WARN"
                        $info.status = "manual_review"
                        $info.no_commits = $true
                        $script:TasksFailed++
                        Save-SuccessMetrics
                        $msg = "Branch $branch has no commits compared to base."
                        $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "no_commits"
                        continue
                    }
                    if (-not $created) { continue }
                    $pr = Get-PRForBranch -Branch $branch
                    if (Test-GithubRateLimit) { return }
                    if (-not $pr) { continue }
                }
                else {
                    Write-Log "[DRY-RUN] Would create PR for $branch" -Level "ACTION"
                    continue
                }
            }
            else {
                # No PR yet — agent might still be working or PR not created
                continue
            }
        }

        $info.pr_number = $pr.number

        $copilotState = Apply-CopilotStateToInfo -Info $info -PRNumber $pr.number

        # Check if already merged
        if ($pr.state -eq "MERGED" -or $pr.mergedAt) {
            Write-Log "PR #$($pr.number) for $branch is MERGED" -Level "OK"
            Clear-PendingFollowUp -Info $info -Reason "pr_merged"
            Remove-CopilotPRState -PRNumber $pr.number
            $info.status = "merged"
            Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
            $processed += $attemptId
            continue
        }

        if ($pr.state -eq "CLOSED") {
            Write-Log "PR #$($pr.number) for $branch is CLOSED — clearing queued follow-ups" -Level "WARN"
            Clear-PendingFollowUp -Info $info -Reason "pr_closed"
            Remove-CopilotPRState -PRNumber $pr.number
            $info.status = "review"
            continue
        }

        $copilotState = Get-CopilotPRState -PRNumber $pr.number
        if ($copilotState -and -not $copilotState.completed) {
            Write-Log "PR #$($pr.number) has Copilot fix pending — deferring agent follow-ups" -Level "INFO"
        }

        # PR exists but not merged — check CI
        $prDetails = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        $mergeState = if ($prDetails -and $prDetails.mergeStateStatus) { $prDetails.mergeStateStatus } else { "UNKNOWN" }

        if ($mergeState -in @("DIRTY", "CONFLICTING")) {
            Write-Log "PR #$($pr.number) has merge conflicts ($mergeState) — archiving attempt, will retry fresh" -Level "WARN"
            if (-not $DryRun) {
                # Archive-and-retry: close conflicting PR, archive attempt, resubmit fresh
                Add-RecentItem -ListName "ConflictRetries" -Item @{
                    pr_number   = $pr.number
                    pr_title    = $pr.title
                    branch      = $branch
                    attempt_id  = $attemptId
                    task_id     = $info.task_id
                    reason      = "merge_conflict_archive_retry"
                    occurred_at = (Get-Date).ToString("o")
                }
                # Close the conflicting PR and delete its branch
                $null = Close-PRDeleteBranch -PRNumber $pr.number
                # Archive the stale attempt
                Archive-VKAttempt -AttemptId $attemptId | Out-Null
                # Reset task to todo so it re-enters the backlog for fresh attempt
                Update-VKTaskStatus -TaskId $info.task_id -Status "todo" | Out-Null
                # Remove from tracked attempts so a fresh slot opens
                $processed += $attemptId
                Write-Log "Archived attempt $($attemptId.Substring(0,8)) for task $($info.task_id.Substring(0,8)) — reset to 'todo' for fresh retry" -Level "ACTION"
            }
            continue
        }

        $baseBranch = if ($prDetails -and $prDetails.baseRefName) { $prDetails.baseRefName } else { $script:VK_TARGET_BRANCH }
        if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }
        $checkStatus = Get-PRRequiredCheckStatus -PRNumber $pr.number -BaseBranch $baseBranch
        $securityStatus = if ($SkipSecurityChecks) { "skipped" } else { Get-PRSecurityCheckStatus -PRNumber $pr.number }
        if (Test-GithubRateLimit) { return }
        Write-Log "PR #$($pr.number) ($branch): CI=$checkStatus Security=$securityStatus" -Level "INFO"

        # Security checks are advisory; only required CI gates control merge.
        $effectiveStatus = $checkStatus

        switch ($effectiveStatus) {
            "passing" {
                Write-Log "CI passing for PR #$($pr.number) — merging..." -Level "ACTION"
                if (-not $DryRun) {
                    $mergeResult = Merge-PRWithFallback -PRNumber $pr.number -ForceAdmin:$UseAdminMerge
                    if (Test-GithubRateLimit) { return }
                    if ($mergeResult.merged) {
                        $info.status = "merged"
                        Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
                        $processed += $attemptId
                    }
                    else {
                        $reason = if ($mergeResult.reason) { $mergeResult.reason } else { "Unknown merge error" }
                        $failure = Get-MergeFailureInfo -PRNumber $pr.number -Reason $reason
                        if (Test-GithubRateLimit) { return }
                        Write-Log "Merge failed for PR #$($pr.number) ($($failure.summary))" -Level "WARN"

                        $retryResult = $null
                        if ($failure.category -notin @("conflict", "review_required", "draft", "blocked", "rate_limit")) {
                            if (-not (Start-InterruptibleSleep -Seconds 2 -Reason "merge-retry")) { break }
                            $retryResult = Merge-PRWithFallback -PRNumber $pr.number -ForceAdmin:$UseAdminMerge
                            if (Test-GithubRateLimit) { return }
                            if ($retryResult.merged) {
                                $info.status = "merged"
                                Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
                                $processed += $attemptId
                                break
                            }
                            $reason = if ($retryResult.reason) { $retryResult.reason } else { $reason }
                            $failure = Get-MergeFailureInfo -PRNumber $pr.number -Reason $reason
                            if (Test-GithubRateLimit) { return }
                            Write-Log "Retry merge failed for PR #$($pr.number) ($($failure.summary))" -Level "WARN"
                        }

                        if ($failure.category -eq "draft") {
                            Write-Log "Marking PR #$($pr.number) ready (draft)" -Level "ACTION"
                            if (-not $DryRun) {
                                $null = Mark-PRReady -PRNumber $pr.number
                            }
                        }

                        if ($failure.category -eq "behind" -and $attemptId -and -not $info.rebase_requested) {
                            Write-Log "Requesting direct rebase for PR #$($pr.number) (attempt $($attemptId.Substring(0,8)))" -Level "ACTION"
                            $info.rebase_requested = $true
                            $rebaseOk = Invoke-DirectRebase -Branch $branch -AttemptId $attemptId
                            if ($rebaseOk) {
                                $info.rebase_requested = $false  # Reset so rebase can be retried if still behind
                            }
                        }

                        $info.merge_failures_total = [int]($info.merge_failures_total ?? 0) + 1
                        $info.merge_failure_cycles = [int]($info.merge_failure_cycles ?? 0) + 1
                        $info.last_merge_failure_reason = $reason
                        $info.last_merge_failure_category = $failure.category
                        $info.last_merge_failure_at = Get-Date

                        # Safety valve: if merge keeps failing (3+ cycles for behind/conflict),
                        # archive and retry fresh instead of getting stuck
                        if ($info.merge_failure_cycles -ge 3 -and $failure.category -in @("behind", "base_changed") -and -not $DryRun) {
                            Write-Log "PR #$($pr.number) stuck behind after $($info.merge_failure_cycles) rebase attempts — archiving for fresh retry" -Level "WARN"
                            $null = Close-PRDeleteBranch -PRNumber $pr.number
                            Archive-VKAttempt -AttemptId $attemptId | Out-Null
                            Update-VKTaskStatus -TaskId $info.task_id -Status "todo" | Out-Null
                            Add-RecentItem -ListName "ConflictRetries" -Item @{
                                pr_number   = $pr.number
                                branch      = $branch
                                attempt_id  = $attemptId
                                task_id     = $info.task_id
                                reason      = "merge_stuck_behind_archive"
                                cycles      = $info.merge_failure_cycles
                                occurred_at = (Get-Date).ToString("o")
                            }
                            $processed += $attemptId
                            continue
                        }

                        if (-not $info.manual_review_notified -and $info.merge_failure_cycles -ge 2) {
                            $info.last_merge_notify_at = Get-Date
                            $info.manual_review_notified = $true
                            $info.status = "manual_review"
                            if (-not $DryRun) {
                                Update-VKTaskStatus -TaskId $info.task_id -Status "inreview" | Out-Null
                            }
                            Write-Log "Merge notify: PR #$($pr.number) stage=manual_review category=$($failure.category) action=$($failure.action) reason=$reason" -Level "WARN"
                        }

                        if ($UseAutoMerge) {
                            $enabled = Merge-PR -PRNumber $pr.number -AutoMerge
                            if (Test-GithubRateLimit) { return }
                            if (-not $enabled) {
                                $autoErr = Get-VKLastGithubError
                                $autoReason = if ($autoErr -and $autoErr.message) { $autoErr.message } else { "Unknown auto-merge error" }
                                Write-Log "Auto-merge enable failed: $autoReason" -Level "WARN"
                            }
                        }
                        if (-not $info.merge_failed_notified -and $failure.category -notin @("behind", "base_changed", "rate_limit")) {
                            $msg = "Merge failed for PR #$($pr.number). Reason: $reason. ($($failure.summary))"
                            $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "merge_failed"
                            $info.merge_failed_notified = $true
                        }
                    }
                }
                else {
                    Write-Log "[DRY-RUN] Would merge PR #$($pr.number)" -Level "ACTION"
                }
                $info.ci_notified = $false
            }
            "pending" {
                # CI still running — enable auto-merge if not already
                Write-Log "CI pending for PR #$($pr.number) — enabling auto-merge" -Level "INFO"
                if (-not $DryRun) {
                    if ($UseAutoMerge) {
                        $enabled = Enable-AutoMerge -PRNumber $pr.number
                        if (Test-GithubRateLimit) { return }
                        if (-not $enabled) {
                            $mergeErr = Get-VKLastGithubError
                            $reason = if ($mergeErr -and $mergeErr.message) { $mergeErr.message } else { "Unknown auto-merge error" }
                            Write-Log "Auto-merge enable failed: $reason" -Level "WARN"
                        }
                    }
                }
            }
            "failing" {
                Write-Log "CI FAILING for PR #$($pr.number) — needs attention" -Level "WARN"
                # Don't block the slot — the agent or a human needs to fix this
                # We mark it so we don't keep retrying every cycle
                $info.status = "review"
                $checks = Get-PRChecksDetail -PRNumber $pr.number
                if (Test-GithubRateLimit) { return }
                if ($null -eq $checks -or $checks -is [System.DBNull] -or -not $checks) {
                    $checks = @()
                }
                $summary = Format-PRCheckFailures -Checks $checks
                $body = @"
CI is failing for PR #$($pr.number).

$summary
"@

                if ($copilotState) {
                    $info.copilot_fix_requested = $true
                    break
                }

                if (-not $info.copilot_fix_pr_number) {
                    $existingCopilot = Find-CopilotFixPR -OriginalPRNumber $pr.number
                    if ($existingCopilot) {
                        $info.copilot_fix_requested = $true
                        $info.copilot_fix_pr_number = $existingCopilot.number
                        Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                            requested_at = (Get-Date).ToString("o")
                            completed    = $false
                            copilot_pr   = $existingCopilot.number
                            merged_at    = $null
                        }
                        break
                    }
                }

                if (-not $info.copilot_fix_requested) {
                    $copilotBody = @"
@copilot CI is failing for PR #$($pr.number).

$summary
"@
                    Write-Log "Requesting Copilot fix for PR #$($pr.number)" -Level "ACTION"
                    if (-not $DryRun) {
                        Add-PRComment -PRNumber $pr.number -Body $copilotBody.Trim() | Out-Null
                        Add-RecentItem -ListName "CopilotRequests" -Item @{
                            pr_number   = $pr.number
                            pr_title    = $pr.title
                            pr_url      = $pr.url
                            reason      = "ci_failing"
                            occurred_at = (Get-Date).ToString("o")
                        }
                    }
                    $info.copilot_fix_requested = $true
                    $info.copilot_fix_requested_at = Get-Date
                    $info.ci_notified = $false
                    Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                        requested_at = (Get-Date).ToString("o")
                        completed    = $false
                        copilot_pr   = $null
                        merged_at    = $null
                    }
                    break
                }
                elseif ($info.copilot_fix_requested -and -not $info.copilot_fix_merged) {
                    # Check if copilot fix is stale (requested > 60 min ago, not merged)
                    $sinceRequested = if ($info.copilot_fix_requested_at) { ((Get-Date) - $info.copilot_fix_requested_at).TotalMinutes } else { 0 }
                    if ($sinceRequested -ge 60) {
                        Write-Log "Copilot fix for PR #$($pr.number) is stale ($([int]$sinceRequested) min) — marking manual_review" -Level "WARN"
                        if ($info.copilot_fix_pr_number -and -not $DryRun) {
                            Write-Log "Closing stale Copilot sub-PR #$($info.copilot_fix_pr_number)" -Level "ACTION"
                            $null = Close-PRDeleteBranch -PRNumber $info.copilot_fix_pr_number
                        }
                        $info.status = "manual_review"
                        $info.copilot_fix_stale = $true
                        if (-not $DryRun) {
                            Update-VKTaskStatus -TaskId $info.task_id -Status "inreview" | Out-Null
                        }
                        $msg = "Copilot fix stale after $([int]$sinceRequested) min for PR #$($pr.number). Manual review required."
                        $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "copilot_fix_stale"
                        break
                    }
                }

                if (-not $info.copilot_fix_merged) {
                    if (-not $info.copilot_fix_pr_number) {
                        $copilotPr = Find-CopilotFixPR -OriginalPRNumber $pr.number
                        if ($copilotPr) {
                            $info.copilot_fix_pr_number = $copilotPr.number
                            Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                                requested_at = (Get-Date).ToString("o")
                                completed    = $false
                                copilot_pr   = $copilotPr.number
                                merged_at    = $null
                            }
                        }
                    }

                    if ($info.copilot_fix_pr_number) {
                        $copilotDetails = Get-PRDetails -PRNumber $info.copilot_fix_pr_number
                        if (Test-GithubRateLimit) { return }
                        if ($copilotDetails -and $copilotDetails.isDraft) {
                            Write-Log "Marking Copilot PR #$($info.copilot_fix_pr_number) ready" -Level "ACTION"
                            if (-not $DryRun) {
                                $null = Mark-PRReady -PRNumber $info.copilot_fix_pr_number
                            }
                        }

                        $isCopilotComplete = if ($copilotDetails) { Test-CopilotPRComplete -PRDetails $copilotDetails } else { $false }
                        if ($isCopilotComplete) {
                            Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                                requested_at = (Get-Date).ToString("o")
                                completed    = $false
                                copilot_pr   = $info.copilot_fix_pr_number
                                merged_at    = $null
                            }
                        }

                        if ($copilotDetails -and $isCopilotComplete) {
                            Write-Log "Merging Copilot PR #$($info.copilot_fix_pr_number) into $branch" -Level "ACTION"
                            if (-not $DryRun) {
                                $mergedFix = Merge-BranchFromPR -BaseBranch $branch -HeadBranch $copilotDetails.headRefName
                                if ($mergedFix) {
                                    $info.copilot_fix_merged = $true
                                    $info.copilot_fix_merged_at = Get-Date
                                    Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                                        requested_at = (Get-Date).ToString("o")
                                        completed    = $true
                                        copilot_pr   = $info.copilot_fix_pr_number
                                        merged_at    = (Get-Date).ToString("o")
                                    }
                                    $null = Close-PRDeleteBranch -PRNumber $info.copilot_fix_pr_number
                                }
                            }
                        }
                    }

                    break
                }

                if ($info.copilot_fix_merged -and -not $info.ci_notified) {
                    $sinceMerge = if ($info.copilot_fix_merged_at) { ((Get-Date) - $info.copilot_fix_merged_at).TotalMinutes } else { 0 }
                    if ($sinceMerge -ge $CiWaitMin) {
                        $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $body.Trim() -Reason "ci_failing_after_copilot_merge"
                        $info.ci_notified = $true
                    }
                }
            }
            default {
                Write-Log "CI status unknown for PR #$($pr.number)" -Level "WARN"
            }
        }
    }

    # Clean up completed attempts from tracking
    foreach ($id in $processed) {
        $script:TrackedAttempts.Remove($id)
    }
    Save-StatusSnapshot
}

function Close-StaleCopilotPRs {
    <#
    .SYNOPSIS Close copilot fix PRs that have been open > 90 minutes without merging.
              Runs once per cycle to prevent endless sub-PR variants.
    #>
    foreach ($entry in $script:TrackedAttempts.GetEnumerator()) {
        $info = $entry.Value
        if (-not $info.copilot_fix_requested) { continue }
        if ($info.copilot_fix_merged) { continue }
        if ($info.copilot_fix_stale) { continue }

        $sinceRequested = if ($info.copilot_fix_requested_at) { ((Get-Date) - $info.copilot_fix_requested_at).TotalMinutes } else { 0 }
        if ($sinceRequested -lt 90) { continue }

        $attemptId = $entry.Key
        Write-Log "Stale copilot fix detected for attempt $($attemptId.Substring(0,8)) ($([int]$sinceRequested) min)" -Level "WARN"

        if ($info.copilot_fix_pr_number -and -not $DryRun) {
            Write-Log "Closing stale Copilot sub-PR #$($info.copilot_fix_pr_number)" -Level "ACTION"
            $null = Close-PRDeleteBranch -PRNumber $info.copilot_fix_pr_number
        }

        # Reset copilot fix state so task can be retried or manually reviewed
        $info.copilot_fix_requested = $false
        $info.copilot_fix_requested_at = $null
        $info.copilot_fix_pr_number = $null
        $info.copilot_fix_stale = $true
        $info.status = "manual_review"
        if (-not $DryRun) {
            Update-VKTaskStatus -TaskId $info.task_id -Status "inreview" | Out-Null
        }
    }
}

function Process-StandaloneCopilotPRs {
    <#
    .SYNOPSIS Merge completed Copilot-authored PRs (non-[WIP]) after required + security checks pass.
    #>
    if (Test-GithubCooldown) { return }
    $openPrs = Get-OpenPullRequests -Limit 200
    if (Test-GithubRateLimit) { return }
    if (-not $openPrs -or @($openPrs).Count -eq 0) { return }

    $copilotCandidates = @($openPrs | Where-Object {
            $_.author -and (Test-IsCopilotAuthor -Author $_.author)
        })
    if ($copilotCandidates.Count -gt 0) {
        Write-Log "Copilot PRs found: $($copilotCandidates.Count)" -Level "INFO"
    }

    foreach ($pr in $copilotCandidates) {
        if (-not $pr.author -or -not (Test-IsCopilotAuthor -Author $pr.author)) { continue }
        if ($pr.title -match '^\[WIP\]') { continue }

        $details = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        if (-not $details) { continue }
        $mergeState = if ($details.mergeStateStatus) { $details.mergeStateStatus } else { "UNKNOWN" }
        $mergeableState = if ($details.mergeable) { $details.mergeable } else { "UNKNOWN" }
        if ($mergeState -eq "CONFLICTING" -or $mergeableState -eq "CONFLICTING") {
            Write-Log "Closing Copilot PR #$($pr.number) due to conflicts ($mergeState) — archive and retry" -Level "WARN"
            if (-not $DryRun) {
                $null = Close-PRDeleteBranch -PRNumber $pr.number
                Add-RecentItem -ListName "ConflictRetries" -Item @{
                    pr_number   = $pr.number
                    pr_title    = $pr.title
                    branch      = $pr.headRefName
                    reason      = "copilot_pr_conflict_archive"
                    occurred_at = (Get-Date).ToString("o")
                }
                # Find and archive the tracked attempt for this branch so the task can be retried
                $branchName = if ($details.headRefName) { $details.headRefName } else { $pr.headRefName }
                $matchingAttempt = $script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.branch -eq $branchName } | Select-Object -First 1
                if ($matchingAttempt) {
                    Archive-VKAttempt -AttemptId $matchingAttempt.Key | Out-Null
                    # Reset task to todo so it re-enters backlog
                    if ($matchingAttempt.Value.task_id) {
                        Update-VKTaskStatus -TaskId $matchingAttempt.Value.task_id -Status "todo" | Out-Null
                    }
                    $script:TrackedAttempts.Remove($matchingAttempt.Key)
                    Write-Log "Archived attempt $($matchingAttempt.Key.Substring(0,8)) — task will be resubmitted fresh" -Level "ACTION"
                }
                $requestedAt = if ($pr.createdAt) { $pr.createdAt } else { (Get-Date).ToString("o") }
                $refs = Get-ReferencedPRNumbers -Texts @(
                    $pr.title,
                    $details.body,
                    $details.headRefName,
                    $pr.headRefName
                )
                if (-not $refs -or @($refs).Count -eq 0) {
                    Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                        requested_at = $requestedAt
                        completed    = $true
                        copilot_pr   = $pr.number
                        merged_at    = $null
                    }
                }
                foreach ($ref in $refs) {
                    Upsert-CopilotPRState -PRNumber $ref -Update @{
                        requested_at = $requestedAt
                        completed    = $true
                        copilot_pr   = $pr.number
                        merged_at    = $null
                    }
                }
            }
            continue
        }
        if ($mergeState -eq "DIRTY" -and $mergeableState -ne "CONFLICTING") {
            Write-Log "Copilot PR #$($pr.number) is DIRTY — closing and archiving for fresh retry" -Level "WARN"
            if (-not $DryRun) {
                $null = Close-PRDeleteBranch -PRNumber $pr.number
                Add-RecentItem -ListName "ConflictRetries" -Item @{
                    pr_number   = $pr.number
                    pr_title    = $pr.title
                    branch      = $pr.headRefName
                    reason      = "copilot_pr_dirty_archive"
                    occurred_at = (Get-Date).ToString("o")
                }
                $branchName = if ($details.headRefName) { $details.headRefName } else { $pr.headRefName }
                $matchingAttempt = $script:TrackedAttempts.GetEnumerator() | Where-Object { $_.Value.branch -eq $branchName } | Select-Object -First 1
                if ($matchingAttempt) {
                    Archive-VKAttempt -AttemptId $matchingAttempt.Key | Out-Null
                    # Reset task to todo so it re-enters backlog
                    if ($matchingAttempt.Value.task_id) {
                        Update-VKTaskStatus -TaskId $matchingAttempt.Value.task_id -Status "todo" | Out-Null
                    }
                    $script:TrackedAttempts.Remove($matchingAttempt.Key)
                    Write-Log "Archived attempt $($matchingAttempt.Key.Substring(0,8)) — task will be resubmitted fresh" -Level "ACTION"
                }
            }
            continue
        }

        if ($details.isDraft) {
            Write-Log "Marking Copilot PR #$($pr.number) ready" -Level "ACTION"
            if (-not $DryRun) {
                $null = Mark-PRReady -PRNumber $pr.number
            }
            # Re-fetch once after marking ready to attempt same-cycle merge.
            $details = Get-PRDetails -PRNumber $pr.number
            if (Test-GithubRateLimit) { return }
            if (-not $details -or $details.isDraft) { continue }
        }

        $baseBranch = if ($details.baseRefName) { $details.baseRefName } else { $script:VK_TARGET_BRANCH }
        if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }
        $checkStatus = Get-PRRequiredCheckStatus -PRNumber $pr.number -BaseBranch $baseBranch
        $securityStatus = if ($SkipSecurityChecks) { "skipped" } else { Get-PRSecurityCheckStatus -PRNumber $pr.number }
        if (Test-GithubRateLimit) { return }
        if ($checkStatus -ne "passing") {
            Write-Log "Copilot PR #$($pr.number) not mergeable — CI=$checkStatus" -Level "INFO"
            continue
        }
        if ($securityStatus -ne "passing") {
            Write-Log "Copilot PR #$($pr.number) security checks are $securityStatus — continuing based on CI" -Level "WARN"
        }

        Write-Log "Merging Copilot PR #$($pr.number)" -Level "ACTION"
        if (-not $DryRun) {
            $mergeResult = Merge-PRWithFallback -PRNumber $pr.number -ForceAdmin:$UseAdminMerge
            if ($mergeResult.merged) {
                $requestedAt = if ($pr.createdAt) { $pr.createdAt } else { (Get-Date).ToString("o") }
                $refs = Get-ReferencedPRNumbers -Texts @(
                    $pr.title,
                    $details.body,
                    $details.headRefName,
                    $pr.headRefName
                )
                if (-not $refs -or @($refs).Count -eq 0) {
                    Upsert-CopilotPRState -PRNumber $pr.number -Update @{
                        requested_at = $requestedAt
                        completed    = $true
                        copilot_pr   = $pr.number
                        merged_at    = (Get-Date).ToString("o")
                    }
                }
                foreach ($ref in $refs) {
                    Upsert-CopilotPRState -PRNumber $ref -Update @{
                        requested_at = $requestedAt
                        completed    = $true
                        copilot_pr   = $pr.number
                        merged_at    = (Get-Date).ToString("o")
                    }
                }
            }
            else {
                $reason = if ($mergeResult.reason) { $mergeResult.reason } else { "Unknown merge error" }
                Write-Log "Merge failed for Copilot PR #$($pr.number): $reason" -Level "WARN"
            }
        }
    }
}

function Complete-Task {
    <#
    .SYNOPSIS Mark a task as done after its PR is merged.
    #>
    [CmdletBinding()]
    param(
        [string]$AttemptId,
        [string]$TaskId,
        [int]$PRNumber
    )
    Write-Log "Marking task $($TaskId.Substring(0,8)) as done (PR #$PRNumber merged)" -Level "OK"
    $taskTitle = $null
    if ($AttemptId -and $script:TrackedAttempts.ContainsKey($AttemptId)) {
        $taskTitle = $script:TrackedAttempts[$AttemptId].name
    }
    $prDetails = Get-PRDetails -PRNumber $PRNumber
    if (Test-GithubRateLimit) {
        $prDetails = $null
    }
    $prTitle = if ($prDetails -and $prDetails.title) { $prDetails.title } else { $null }
    $prUrl = if ($prDetails -and $prDetails.url) { $prDetails.url } else { $null }
    if (-not $DryRun) {
        Archive-VKAttempt -AttemptId $AttemptId | Out-Null
        Update-VKTaskStatus -TaskId $TaskId -Status "done"
    }

    # ─── Executor health: report success & release slot ──────────────────────
    if ($AttemptId -and $script:TrackedAttempts.ContainsKey($AttemptId)) {
        $execProvider = $script:TrackedAttempts[$AttemptId].executor_provider
        if ($execProvider) {
            Report-ExecutorSuccess -Provider $execProvider
            Decrement-ExecutorActiveCount -Provider $execProvider
        }
    }

    Add-RecentItem -ListName "CompletedTasks" -Item @{
        task_id      = $TaskId
        task_title   = $taskTitle
        pr_number    = $PRNumber
        pr_title     = $prTitle
        pr_url       = $prUrl
        completed_at = (Get-Date).ToString("o")
    }
    $script:TasksCompleted++

    # ─── Success rate classification ─────────────────────────────────────────
    $neededFix = $false
    if ($AttemptId -and $script:TrackedAttempts.ContainsKey($AttemptId)) {
        $info = $script:TrackedAttempts[$AttemptId]
        if ($info.copilot_fix_requested -or $info.copilot_fix_merged) {
            $neededFix = $true
        }
        if ($info.merge_failures_total -gt 0 -and $info.last_merge_failure_category -eq "conflict") {
            $neededFix = $true
        }
        if ($info.status -in @("error", "manual_review")) {
            $neededFix = $true
        }
    }
    if ($neededFix) {
        $script:TasksNeededFix++
        Write-Log "Task $($TaskId.Substring(0,8)) merged after fix (needed intervention)" -Level "INFO"
    }
    else {
        $script:FirstShotSuccess++
        Write-Log "Task $($TaskId.Substring(0,8)) merged first-shot" -Level "OK"
    }
    Save-SuccessMetrics
    Maybe-TriggerCISweep
}

function Save-SuccessMetrics {
    <#
    .SYNOPSIS Persist success rate metrics to disk for monitoring.
    #>
    $total = $script:FirstShotSuccess + $script:TasksNeededFix + $script:TasksFailed
    $rate = if ($total -gt 0) { [math]::Round(($script:FirstShotSuccess / $total) * 100, 1) } else { 0 }
    $metrics = @{
        updated_at         = (Get-Date).ToString("o")
        first_shot_success = $script:FirstShotSuccess
        needed_fix         = $script:TasksNeededFix
        failed             = $script:TasksFailed
        total_decided      = $total
        first_shot_rate    = $rate
        session_start      = $script:StartTime.ToString("o")
    }
    $dir = Split-Path -Parent $script:SuccessMetricsPath
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }
    $metrics | ConvertTo-Json -Depth 3 | Set-Content -Path $script:SuccessMetricsPath -Encoding UTF8
}

function Get-SuccessRateSummary {
    <#
    .SYNOPSIS Return a formatted string showing current success rate.
    #>
    $total = $script:FirstShotSuccess + $script:TasksNeededFix + $script:TasksFailed
    $rate = if ($total -gt 0) { [math]::Round(($script:FirstShotSuccess / $total) * 100, 1) } else { 0 }
    return "First-shot: ${rate}% ($($script:FirstShotSuccess)/$total) | Fix: $($script:TasksNeededFix) | Failed: $($script:TasksFailed)"
}

function Maybe-TriggerCISweep {
    <#
    .SYNOPSIS Trigger a Copilot CI sweep every 15 completed tasks.
    #>
    if ($script:TasksCompleted -lt 15) { return }
    if (($script:TasksCompleted % 15) -ne 0) { return }

    $state = Get-OrchestratorState
    if ($state.last_ci_sweep_completed -eq $script:TasksCompleted) { return }

    Trigger-CISweep
    $state.last_ci_sweep_completed = $script:TasksCompleted
    Save-OrchestratorState -State $state
}

function Trigger-CISweep {
    <#
    .SYNOPSIS Create a Copilot-driven CI sweep issue.
    #>
    $failedRuns = @()
    $recentMerged = @()
    if (-not (Test-GithubRateLimit)) {
        $failedRuns = Get-FailingWorkflowRuns -Limit 10
        $recentMerged = Get-RecentMergedPRs -Limit 15
    }

    $failedRunLines = if ($failedRuns -and @($failedRuns).Count -gt 0) {
        $failedRuns | ForEach-Object {
            $name = if ($_.name) { $_.name } elseif ($_.workflow_name) { $_.workflow_name } else { "workflow" }
            $url = if ($_.html_url) { $_.html_url } else { "" }
            $conclusion = if ($_.conclusion) { $_.conclusion } else { "unknown" }
            if ($url) { "- $name ($conclusion): $url" } else { "- $name ($conclusion)" }
        }
    }
    else { @("- none") }

    $mergedLines = if ($recentMerged -and @($recentMerged).Count -gt 0) {
        $recentMerged | ForEach-Object {
            $title = if ($_.title) { $_.title } else { "PR #$($_.number)" }
            $url = if ($_.url) { $_.url } else { "" }
            if ($url) { "- #$($_.number) ${title}: $url" } else { "- #$($_.number) $title" }
        }
    }
    else { @("- none") }

    $title = "ci sweep: resolve failing workflows"
    $body = @"
Copilot assignment: this issue will be assigned via API. If it is unassigned, use the "Assign to Copilot" button.

@copilot Please review GitHub Actions failures across the repository and resolve them.

Scope:
- Identify failing workflows on main.
- Prioritize required checks and security scans.
- Apply minimal fixes and open PRs as needed.

Recent failing workflow runs (main):
$($failedRunLines -join "`n")

Recent merged PRs (last 15):
$($mergedLines -join "`n")
"@

    Write-Log "Triggering CI sweep (every 15 tasks)" -Level "ACTION"
    if (-not $DryRun) {
        $issue = Create-GithubIssue -Title $title -Body $body
        if ($issue -and $issue.number) {
            $assigned = Assign-IssueToCopilot -IssueNumber $issue.number
            if (-not $assigned) {
                Write-Log "Copilot assignment failed for issue #$($issue.number)" -Level "WARN"
            }
        }
    }
}

function Test-TaskFileOverlap {
    <#
    .SYNOPSIS Check if a task's likely target files overlap with active agent branches.
              Returns $true if overlap detected (should skip), $false if safe.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$TaskTitle,
        [Parameter(Mandatory)][string]$TaskId
    )

    # Map task title keywords to likely file paths/modules
    $modulePatterns = @{
        'portal'     = @('portal/', 'lib/portal/', 'lib/capture/', 'lib/admin/', 'pnpm-lock.yaml')
        'veid'       = @('x/veid/', 'pkg/inference/')
        'market'     = @('x/market/')
        'escrow'     = @('x/escrow/')
        'mfa'        = @('x/mfa/')
        'encryption' = @('x/encryption/')
        'provider'   = @('pkg/provider_daemon/', 'cmd/provider-daemon/')
        'hpc'        = @('x/hpc/')
        'roles'      = @('x/roles/')
        'sdk'        = @('sdk/')
        'app'        = @('app/')
        'ci'         = @('.github/', 'Makefile', 'make/')
        'ml'         = @('ml/')
        'deps'       = @('go.mod', 'go.sum', 'vendor/')
    }

    # Determine this task's likely modules from title
    $titleLower = $TaskTitle.ToLower()
    $taskModules = [System.Collections.Generic.List[string]]::new()
    foreach ($kv in $modulePatterns.GetEnumerator()) {
        if ($titleLower -match $kv.Key) {
            foreach ($p in $kv.Value) { $taskModules.Add($p) }
        }
    }

    # If we can't determine the module, allow it (don't block unknowns)
    if ($taskModules.Count -eq 0) { return $false }

    # Check active agents' branches for file overlap
    $activeAttempts = $script:TrackedAttempts.Values | Where-Object { $_.status -eq "running" -and $_.branch }
    foreach ($active in $activeAttempts) {
        if ($active.task_id -eq $TaskId) { continue }

        $activeName = if ($active.name) { $active.name.ToLower() } else { "" }
        foreach ($kv in $modulePatterns.GetEnumerator()) {
            if ($activeName -match $kv.Key) {
                # Check if any of the active's modules overlap with this task's modules
                foreach ($activeP in $kv.Value) {
                    foreach ($taskP in $taskModules) {
                        if ($activeP -eq $taskP -or $activeP.StartsWith($taskP) -or $taskP.StartsWith($activeP)) {
                            Write-Log "Task overlap: '$($TaskTitle.Substring(0, [Math]::Min(50, $TaskTitle.Length)))' conflicts with active '$($active.name)' on $activeP" -Level "WARN"
                            return $true
                        }
                    }
                }
            }
        }
    }

    # Global conflict files — Go module changes always risk go.mod/vendor conflicts
    $goModules = @('veid', 'market', 'escrow', 'mfa', 'encryption', 'hpc', 'roles', 'provider', 'app', 'sdk', 'ml')
    $taskIsGo = $goModules | Where-Object { $titleLower -match $_ }
    if ($taskIsGo) {
        $activeGoCount = @($activeAttempts | Where-Object {
                $n = if ($_.name) { $_.name.ToLower() } else { "" }
                $goModules | Where-Object { $n -match $_ }
            }).Count
        # Allow up to 2 parallel Go module tasks, defer beyond that
        if ($activeGoCount -ge 2) {
            Write-Log "Task overlap: deferring Go task '$($TaskTitle.Substring(0, [Math]::Min(50, $TaskTitle.Length)))' — $activeGoCount Go tasks already active" -Level "WARN"
            return $true
        }
    }

    # Portal tasks — only 1 at a time (pnpm-lock.yaml conflicts)
    $taskIsPortal = $titleLower -match 'portal|frontend|capture|admin|ui'
    if ($taskIsPortal) {
        $activePortalCount = @($activeAttempts | Where-Object {
                $n = if ($_.name) { $_.name.ToLower() } else { "" }
                $n -match 'portal|frontend|capture|admin|ui'
            }).Count
        if ($activePortalCount -ge 1) {
            Write-Log "Task overlap: deferring portal task — $activePortalCount portal tasks already active" -Level "WARN"
            return $true
        }
    }

    return $false
}

function Fill-ParallelSlots {
    <#
    .SYNOPSIS Submit new task attempts to reach the target parallelism.
              Enforces merge gate: won't start new tasks if previous ones have unmerged PRs.
              Uses 50/50 Codex/Copilot executor cycling.
    #>
    # Count truly active slots: agents currently working
    $activeCount = @($script:TrackedAttempts.Values | Where-Object { $_.status -eq "running" }).Count

    $effectiveMax = Get-EffectiveMaxParallel
    $slotsAvailable = $effectiveMax - $activeCount
    if ($slotsAvailable -le 0) {
        Write-Log "All $effectiveMax slots occupied ($activeCount active)" -Level "INFO"
        return
    }

    Write-Log "$slotsAvailable slots available (target: $effectiveMax, active: $activeCount)" -Level "ACTION"

    # Get next todo tasks
    $nextTasks = Get-OrderedTodoTasks -Count $slotsAvailable
    if (-not $nextTasks -or @($nextTasks).Count -eq 0) {
        Write-Log "No more todo tasks in backlog" -Level "WARN"
        return
    }

    foreach ($task in $nextTasks) {
        # Skip tasks that already have a tracked (non-done) attempt
        $existingAttempt = $script:TrackedAttempts.Values | Where-Object { $_.task_id -eq $task.id -and $_.status -ne "done" }
        if ($existingAttempt) {
            Write-Log "Task $($task.id.Substring(0,8)) already has active attempt, skipping" -Level "INFO"
            continue
        }

        # Pre-flight: check file overlap with active agents
        $taskTitle = if ([string]::IsNullOrWhiteSpace($task.title)) { "" } else { $task.title }
        if (Test-TaskFileOverlap -TaskTitle $taskTitle -TaskId $task.id) {
            Write-Log "Deferring task $($task.id.Substring(0,8)) — file overlap with active agent" -Level "INFO"
            continue
        }

        $title = if ([string]::IsNullOrWhiteSpace($task.title)) {
            if ($task.id) { "Task $($task.id.Substring(0,8))" } else { "Task (untitled)" }
        }
        else {
            $task.title
        }
        $shortTitle = $title.Substring(0, [Math]::Min(70, $title.Length))

        # ─── Smart executor routing: classify task, pick best model ──────
        $taskDesc = if ($task.description) { $task.description } else { "" }
        $complexity = Get-TaskComplexity -Title $title -Description $taskDesc

        # Check for Telegram /model override
        $overridePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot "..")) ".cache\executor-override.json"
        $nextExec = $null
        if (Test-Path $overridePath) {
            try {
                $override = Get-Content $overridePath -Raw | ConvertFrom-Json
                $expired = $false
                if ($override.expires_at -and (Get-Date) -gt [DateTime]::Parse($override.expires_at)) { $expired = $true }
                if ($override.remaining_tasks -le 0) { $expired = $true }

                if (-not $expired -and $override.model) {
                    # Find executor matching the override model
                    $match = $script:VK_EXECUTORS | Where-Object { $_.model -eq $override.model } | Select-Object -First 1
                    if ($match) {
                        $nextExec = $match
                        # Decrement remaining tasks
                        $override.remaining_tasks = [Math]::Max(0, $override.remaining_tasks - 1)
                        $override | ConvertTo-Json -Depth 3 | Set-Content $overridePath -Encoding UTF8
                        Write-Log "Model override active: $($override.model) ($($override.remaining_tasks) remaining)" -Level "INFO"
                        if ($override.remaining_tasks -le 0) {
                            Remove-Item $overridePath -ErrorAction SilentlyContinue
                            Write-Log "Model override exhausted — returning to smart routing" -Level "INFO"
                        }
                    }
                }
                else {
                    # Override expired — clean up
                    Remove-Item $overridePath -ErrorAction SilentlyContinue
                }
            }
            catch {
                Write-Log "Failed to read model override: $($_.Exception.Message)" -Level "WARN"
            }
        }

        if (-not $nextExec) {
            $nextExec = Get-BestExecutorForTask -Complexity $complexity
        }
        if (-not $nextExec) { $nextExec = Get-CurrentExecutorProfile }
        $modelTag = if ($nextExec.model) { $nextExec.model } else { $nextExec.variant }
        Write-Log "Submitting: $shortTitle [$($nextExec.executor)/$($nextExec.variant)] model=$modelTag complexity=$complexity" -Level "ACTION"

        if (-not $DryRun) {
            $attempt = Submit-VKTaskAttempt -TaskId $task.id -ExecutorOverride $nextExec
            if ($attempt) {
                $executorProvider = if ($nextExec.provider) { $nextExec.provider } else { "unknown" }
                Increment-ExecutorActiveCount -Provider $executorProvider
                $script:TrackedAttempts[$attempt.id] = @{
                    task_id           = $task.id
                    branch            = $attempt.branch
                    pr_number         = $null
                    status            = "running"
                    name              = $title
                    executor          = $nextExec.executor
                    executor_provider = $executorProvider
                }
                $taskUrl = Get-TaskUrl -TaskId $task.id
                Add-RecentItem -ListName "SubmittedTasks" -Item @{
                    task_id      = $task.id
                    task_title   = $title
                    attempt_id   = $attempt.id
                    submitted_at = (Get-Date).ToString("o")
                    task_url     = $taskUrl
                }
                $seqValue = Get-SequenceValue -Title $task.title
                $state = Get-OrchestratorState
                if ($seqValue) {
                    $state.last_sequence_value = $seqValue
                }
                $state.last_task_id = $task.id
                $state.last_submitted_at = (Get-Date).ToString("o")
                Save-OrchestratorState -State $state
                $script:TasksSubmitted++
            }
        }
        else {
            Write-Log "[DRY-RUN] Would submit task $($task.id.Substring(0,8)) via $($nextExec.executor)" -Level "ACTION"
            # Still advance the cycling index in dry-run for accurate preview
            $null = Get-NextExecutorProfile
        }
    }
}

function Test-BacklogEmpty {
    <#
    .SYNOPSIS Check if there are no more tasks to process.
    #>
    $hasTracked = $script:TrackedAttempts.Count -gt 0
    if ($hasTracked) { return $false }

    $backlogRemaining = Get-TodoBacklogCount -MaxAgeSec ([math]::Max(120, ($PollIntervalSec * 2)))
    if ($null -eq $backlogRemaining) {
        $todoTasks = Get-VKTasks -Status "todo"
        $backlogRemaining = if ($todoTasks) { @($todoTasks).Count } else { 0 }
        Set-TodoBacklogCount -Count $backlogRemaining
    }

    return $backlogRemaining -eq 0
}

# ─── Main Loop ────────────────────────────────────────────────────────────────

function Start-Orchestrator {
    $mutexWaitStart = $null
    do {
        $script:OrchestratorMutex = Enter-OrchestratorMutex
        if ($script:OrchestratorMutex) { break }

        if ($OneShot) {
            Write-Log "Another orchestrator instance is already running. Exiting (OneShot mode)." -Level "WARN"
            return
        }

        if (-not $script:WaitForMutexEffective) {
            Write-Log "Another orchestrator instance is already running. Exiting (mutex held). Remove -NoWaitForMutex to wait." -Level "WARN"
            return
        }

        $sleepSec = $PollIntervalSec
        if ($MutexWaitMaxMin -gt 0) {
            if (-not $mutexWaitStart) { $mutexWaitStart = Get-Date }
            $elapsed = (Get-Date) - $mutexWaitStart
            $remainingSec = [math]::Max(0, [int](($MutexWaitMaxMin * 60) - $elapsed.TotalSeconds))
            if ($remainingSec -le 0) {
                Write-Log "Mutex wait exceeded ${MutexWaitMaxMin}m. Exiting to avoid stale wait (use -MutexWaitMaxMin 0 to wait indefinitely)." -Level "WARN"
                return
            }
            $sleepSec = [math]::Min($sleepSec, $remainingSec)
        }

        Write-Log "Another orchestrator instance is already running. Waiting ${sleepSec}s before retry." -Level "WARN"
        if (-not (Start-MutexWaitWithHeartbeat -Seconds $sleepSec -HeartbeatSec 15)) { return }
    } while ($true)
    Register-OrchestratorShutdownHandlers
    if (Test-OrchestratorStop) { return }
    try {
        Ensure-VeKanbanLibraryLoaded -RequiredFunctions $requiredFunctions | Out-Null
    }
    catch {
        Write-Log $_.Exception.Message -Level "ERROR"
        return
    }

    Reset-TerminalIfRequested
    Write-Banner
    Ensure-GitIdentity

    # Validate prerequisites
    $ghVersion = gh --version 2>$null
    if (-not $ghVersion) {
        Write-Log "GitHub CLI (gh) not found. Install: https://cli.github.com/" -Level "ERROR"
        return
    }
    Write-Log "GitHub CLI: $($ghVersion | Select-Object -First 1)" -Level "INFO"

    $healthTimeout = [math]::Max(2, [math]::Min($VKApiTimeoutSec, 10))
    $healthBaseUrl = Get-VKBaseUrl
    $vkRetryDelay = 15  # First retry fast, then exponential backoff up to 300s
    while (-not (Test-VKApiReachable -TimeoutSec $healthTimeout)) {
        Write-Log "Vibe-kanban API not reachable at $healthBaseUrl. Waiting ${vkRetryDelay}s before retry." -Level "WARN"
        if ($OneShot) { return }
        $vkHeartbeat = if ($vkRetryDelay -ge 60) { 30 } else { 0 }
        if (-not (Start-InterruptibleSleep -Seconds $vkRetryDelay -Reason "vk-health" -HeartbeatSec $vkHeartbeat)) { return }
        $vkRetryDelay = [math]::Min($vkRetryDelay * 2, 300)
    }

    if (-not (Get-Command Initialize-VKConfig -ErrorAction SilentlyContinue)) {
        $reloadPath = Resolve-VeKanbanLibraryPath
        if ($reloadPath) {
            Write-Log "Reloading ve-kanban library from $reloadPath" -Level "WARN"
            try {
                . $reloadPath
            }
            catch {
                Write-Log "Failed to reload ve-kanban library: $($_.Exception.Message)" -Level "ERROR"
                return
            }
        }
    }
    if (-not (Get-Command Initialize-VKConfig -ErrorAction SilentlyContinue)) {
        Write-Log "ve-kanban library not loaded (Initialize-VKConfig missing). Check scripts/ve-kanban.ps1 path." -Level "ERROR"
        return
    }

    if ($VKApiTimeoutSec -gt 0) {
        $global:PSDefaultParameterValues["Invoke-WebRequest:TimeoutSec"] = $VKApiTimeoutSec
        $global:PSDefaultParameterValues["Invoke-RestMethod:TimeoutSec"] = $VKApiTimeoutSec
        Write-Log "VK API timeout set to ${VKApiTimeoutSec}s" -Level "INFO"
    }

    # Auto-detect project and repo IDs
    $initOk = $false
    for ($attempt = 1; $attempt -le $VKApiRetryCount; $attempt++) {
        try {
            $initOk = Initialize-VKConfig
        }
        catch {
            $initOk = $false
            Write-Log "Initialize-VKConfig error: $($_.Exception.Message)" -Level "WARN"
        }
        if ($initOk) { break }
        if ($attempt -lt $VKApiRetryCount) {
            Write-Log "Retrying vibe-kanban init in ${VKApiRetryDelaySec}s (attempt $attempt/$VKApiRetryCount)" -Level "WARN"
            if (-not (Start-InterruptibleSleep -Seconds $VKApiRetryDelaySec -Reason "vk-init-retry")) { return }
        }
    }
    if (-not $initOk) {
        Write-Log "Failed to initialize vibe-kanban configuration after $VKApiRetryCount attempts" -Level "ERROR"
        return
    }
    $projectShort = Format-ShortId -Value $script:VK_PROJECT_ID
    $repoShort = Format-ShortId -Value $script:VK_REPO_ID
    if (-not $projectShort) { $projectShort = "unknown" }
    if (-not $repoShort) { $repoShort = "unknown" }
    Write-Log "Project: $projectShort...  Repo: $repoShort..." -Level "OK"

    # Initialize executor health tracking
    Initialize-ExecutorHealth

    # Initialize Codex region tracking (capture US config for later restore)
    Initialize-CodexRegionTracking
    $regionInfo = Get-RegionStatus
    $regionTag = "$($regionInfo.active_region)"
    if ($regionInfo.sweden_available) { $regionTag += " (sweden backup available)" }
    else { $regionTag += " (sweden NOT configured — set AZURE_SWEDEN_API_KEY + AZURE_SWEDEN_ENDPOINT)" }
    Write-Log "Codex region: $regionTag" -Level "INFO"

    # Log executor health status
    $healthSummary = Get-ExecutorHealthSummary
    Write-Log "Executors: $healthSummary" -Level "INFO"

    # Initial sync
    Sync-TrackedAttempts
    $initialCounts = Get-TrackedStatusCounts
    Write-Log "Initial sync: $($script:TrackedAttempts.Count) tracked (running=$($initialCounts.running), review=$($initialCounts.review), error=$($initialCounts.error))" -Level "INFO"

    Write-Log "SyncCopilotState: seeding Copilot PR state" -Level "INFO"
    Sync-CopilotPRState

    if ($RunMergeStrategy) {
        Write-Log "RunMergeStrategy: merge-only mode" -Level "INFO"
        $candidates = Get-MergeCandidates
        if (-not $candidates -or @($candidates).Count -eq 0) {
            Write-Log "No merge candidates found" -Level "INFO"
            Write-Log "RunMergeStrategy complete" -Level "OK"
            return
        }

        $candidateList = $candidates | ForEach-Object { "#$($_.pr_number) ($($_.branch))" }
        Write-Log "Merge candidates: $($candidateList -join ', ')" -Level "INFO"

        $index = 0
        foreach ($candidate in $candidates) {
            $forceAdmin = $index -gt 0
            $adminLabel = if ($forceAdmin) { "admin" } else { "standard" }
            Write-Log "Merging PR #$($candidate.pr_number) [$adminLabel]" -Level "ACTION"
            if (-not $DryRun) {
                $mergeResult = Merge-PRWithFallback -PRNumber $candidate.pr_number -ForceAdmin:$forceAdmin
                if ($mergeResult.merged) {
                    if ($candidate.attempt_id -and $candidate.task_id) {
                        Complete-Task -AttemptId $candidate.attempt_id -TaskId $candidate.task_id -PRNumber $candidate.pr_number
                        $script:TrackedAttempts.Remove($candidate.attempt_id)
                    }
                    else {
                        Write-Log "Merged PR #$($candidate.pr_number) (no tracked attempt)" -Level "OK"
                    }
                }
                else {
                    Write-Log "Merge failed for PR #$($candidate.pr_number)" -Level "WARN"
                }
            }
            else {
                Write-Log "[DRY-RUN] Would merge PR #$($candidate.pr_number)" -Level "ACTION"
            }
            $index++
        }
        Write-Log "RunMergeStrategy complete" -Level "OK"
        return
    }

    try {
        do {
            if (Test-OrchestratorStop) { break }
            $script:CycleCount++
            Write-CycleSummary
            $counts = Get-TrackedStatusCounts
            Write-Log "Status: running=$($counts.running), review=$($counts.review), error=$($counts.error)" -Level "INFO"

            # Step 1: Sync attempt state from API
            Sync-TrackedAttempts
            if (Test-OrchestratorStop) { break }

            # Step 1b: Region auto-recovery — switch back to US after cooldown
            if ((Get-ActiveCodexRegion) -eq "sweden" -and -not $script:RegionOverride) {
                if (Test-RegionCooldownExpired) {
                    Write-Log "Region cooldown expired — switching Codex back to US" -Level "ACTION"
                    $result = Switch-CodexRegion -Region "us"
                    if ($result.changed) {
                        Write-Log "Region switched: sweden → us" -Level "OK"
                        # Reset the primary codex health to allow it to be tried again
                        Report-ExecutorSuccess -Provider "azure_codex_52"
                    }
                }
            }

            # Step 1c: Auto-failover to Sweden on sustained US degradation
            if ((Get-ActiveCodexRegion) -eq "us" -and -not $script:RegionOverride) {
                $usHealth = Get-ExecutorHealthStatus -Provider "azure_codex_52"
                if ($usHealth -eq "cooldown") {
                    $regionStatus = Get-RegionStatus
                    if ($regionStatus.sweden_available) {
                        Write-Log "Primary US codex in cooldown — failover to Sweden" -Level "ACTION"
                        $result = Switch-CodexRegion -Region "sweden"
                        if ($result.changed) {
                            Write-Log "Region switched: us → sweden (auto-failover)" -Level "OK"
                        }
                    }
                }
            }

            # Step 2: Merge standalone Copilot PRs
            Process-StandaloneCopilotPRs
            if (Test-OrchestratorStop) { break }

            # Step 2b: Close stale Copilot fix PRs
            Close-StaleCopilotPRs
            if (Test-OrchestratorStop) { break }

            # Step 3: Process completed attempts (check PRs, merge, mark done)
            Process-CompletedAttempts
            if (Test-OrchestratorStop) { break }

            # Step 3b: Send queued follow-ups before starting new tasks
            Process-PendingFollowUps
            if (Test-OrchestratorStop) { break }

            # Step 4: Fill empty parallel slots with new task submissions
            if (@(Get-PendingFollowUpAttempts).Count -eq 0) {
                Fill-ParallelSlots
            }

            # Step 5: Check if we're done
            if (Test-BacklogEmpty) {
                Write-Log "All tasks completed! Backlog empty, no active attempts." -Level "OK"
                Write-Host ""
                Write-Host "  ╔══════════════════════════════════════════╗" -ForegroundColor Green
                Write-Host "  ║   ALL TASKS COMPLETE                    ║" -ForegroundColor Green
                Write-Host "  ║   Submitted: $($script:TasksSubmitted.ToString().PadRight(28))║" -ForegroundColor Green
                Write-Host "  ║   Completed: $($script:TasksCompleted.ToString().PadRight(28))║" -ForegroundColor Green
                Write-Host "  ║   Cycles:    $($script:CycleCount.ToString().PadRight(28))║" -ForegroundColor Green
                $successLine = Get-SuccessRateSummary
                Write-Host "  ║   $($successLine.PadRight(39))║" -ForegroundColor Green
                Write-Host "  ╚══════════════════════════════════════════╝" -ForegroundColor Green
                Write-Host ""
                break
            }

            if ($OneShot) {
                Write-Log "OneShot mode — exiting after single cycle" -Level "INFO"
                break
            }

            # Step 6: Wait before next cycle
            Write-Log "Sleeping ${PollIntervalSec}s until next cycle... (Ctrl+C to stop)" -Level "INFO"
            $sleepHeartbeat = if ($PollIntervalSec -ge 60) { 30 } else { 0 }
            if (-not (Start-InterruptibleSleep -Seconds $PollIntervalSec -Reason "cycle-wait" -HeartbeatSec $sleepHeartbeat)) { break }

            Save-StatusSnapshot

        } while ($true)
    }
    finally {
        if ($script:OrchestratorMutex -and
            $script:OrchestratorMutex -is [System.Threading.Mutex]) {
            try { $script:OrchestratorMutex.ReleaseMutex() } catch { }
            try { $script:OrchestratorMutex.Dispose() } catch { }
            $script:OrchestratorMutex = $null
        }
    }
}

# ─── Entry Point ──────────────────────────────────────────────────────────────
Start-Orchestrator
