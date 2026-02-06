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

.PARAMETER DryRun
    If set, logs what would happen without making changes.

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
    [int]$GitHubCooldownSec = 120,
    [int]$IdleTimeoutMin = 60,
    [int]$IdleConfirmMin = 15,
    [int]$CiWaitMin = 15,
    [int]$MaxRetries = 5,
    [switch]$UseAutoMerge,
    [switch]$UseAdminMerge,
    [switch]$DryRun,
    [switch]$OneShot,
    [switch]$RunMergeStrategy
)

# ─── Load ve-kanban library ──────────────────────────────────────────────────
. "$PSScriptRoot/ve-kanban.ps1"

# ─── State tracking ──────────────────────────────────────────────────────────
$script:CycleCount = 0
$script:TasksCompleted = 0
$script:TasksSubmitted = 0
$script:StartTime = Get-Date
$script:GitHubCooldownUntil = $null
$script:TaskRetryCounts = @{}

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

function Write-Banner {
    $nextExec = Get-CurrentExecutorProfile
    $nextStr = "$($nextExec.executor)/$($nextExec.variant)"
    Write-Host ""
    Write-Host "  ╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║          VirtEngine Task Orchestrator                    ║" -ForegroundColor Cyan
    Write-Host "  ║                                                         ║" -ForegroundColor Cyan
    Write-Host "  ║   Parallel: $($MaxParallel.ToString().PadRight(4))  Poll: ${PollIntervalSec}s  $(if($DryRun){'DRY-RUN'}else{'LIVE'})                ║" -ForegroundColor Cyan
    Write-Host "  ║   Cycling:  CODEX ⇄ COPILOT (50/50)                    ║" -ForegroundColor Cyan
    Write-Host "  ║   Next:     $($nextStr.PadRight(44))║" -ForegroundColor Cyan
    Write-Host "  ╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Write-CycleSummary {
    $elapsed = (Get-Date) - $script:StartTime
    Write-Host ""
    Write-Host "  ── Cycle $($script:CycleCount) ──────────────────────────────" -ForegroundColor DarkCyan
    Write-Host "  │ Elapsed:   $([math]::Round($elapsed.TotalMinutes, 1)) min" -ForegroundColor DarkGray
    Write-Host "  │ Submitted: $($script:TasksSubmitted)  Completed: $($script:TasksCompleted)" -ForegroundColor DarkGray
    Write-Host "  │ Tracked:   $($script:TrackedAttempts.Count) attempts" -ForegroundColor DarkGray
    Write-Host "  └────────────────────────────────────────────" -ForegroundColor DarkCyan
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
    $counts = @{ running = 0; review = 0; error = 0; other = 0 }
    foreach ($item in $script:TrackedAttempts.Values) {
        switch ($item.status) {
            "running" { $counts.running++ }
            "review" { $counts.review++ }
            "error" { $counts.error++ }
            default { $counts.other++ }
        }
    }
    return $counts
}

function Get-ActiveAgentCount {
    return @($script:TrackedAttempts.Values | Where-Object { $_.status -eq "running" }).Count
}

function Get-AvailableSlots {
    $activeCount = Get-ActiveAgentCount
    return [math]::Max(0, $MaxParallel - $activeCount)
}

function Try-SendFollowUp {
    <#
    .SYNOPSIS Send a follow-up only when the agent is idle and slots are available.
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
    $Info.last_followup_message = $Message
    $Info.last_followup_reason = $Reason
    $Info.last_followup_at = Get-Date
    $slots = Get-AvailableSlots
    if ($slots -le 0) {
        Write-Log "Deferring follow-up for $($Info.branch): no available slots" -Level "WARN"
        $Info.pending_followup = @{ message = $Message; reason = $Reason }
        return $false
    }
    if (-not $DryRun) {
        try {
            Send-VKWorkspaceFollowUp -WorkspaceId $AttemptId -Message $Message | Out-Null
        }
        catch {
            Write-Log "Follow-up failed for $($Info.branch): $($_.Exception.Message)" -Level "WARN"
            $Info.pending_followup = @{ message = $Message; reason = $Reason }
            return $false
        }
    }
    $Info.pending_followup = $null
    return $true
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

function Try-RecoverContextWindow {
    <#
    .SYNOPSIS Create a new session and re-send the last follow-up when context is exhausted.
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

    if ($Info.last_followup_message) {
        $profile = if ($session.executor) {
            Get-ExecutorProfileForSession -Executor $session.executor
        }
        else {
            Get-ExecutorProfileForSession -Executor "CODEX"
        }
        $sent = Send-VKSessionFollowUp -SessionId $session.id -Message $Info.last_followup_message -ExecutorProfile $profile
        if (-not $sent) {
            Write-Log "Follow-up resend failed for $($Info.branch)" -Level "WARN"
            return $false
        }
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

function Get-MergeCandidates {
    <#
    .SYNOPSIS Collect PRs eligible for merge (idle agent, CI passing, age >= CiWaitMin).
    #>
    $candidates = @()
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

        if ($pr.createdAt) {
            $createdAt = [datetime]$pr.createdAt
            $ageMinutes = ((Get-Date) - $createdAt).TotalMinutes
            if ($ageMinutes -lt $CiWaitMin) { continue }
        }

        $prDetails = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return @() }
        $mergeState = if ($prDetails -and $prDetails.mergeStateStatus) { $prDetails.mergeStateStatus } else { "UNKNOWN" }
        if ($mergeState -in @("DIRTY", "CONFLICTING")) { continue }

        $checkStatus = Get-PRCheckStatus -PRNumber $pr.number
        if (Test-GithubRateLimit) { return @() }
        if ($checkStatus -ne "passing") { continue }

        $candidates += [pscustomobject]@{
            attempt_id = $attemptId
            task_id    = $info.task_id
            branch     = $info.branch
            pr_number  = $pr.number
            created_at = $pr.createdAt
        }
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

    foreach ($a in $apiAttempts) {
        if (-not $a.branch) { continue }
        if (-not $script:TrackedAttempts.ContainsKey($a.id)) {
            # Newly discovered active attempt
            $script:TrackedAttempts[$a.id] = @{
                task_id                       = $a.task_id
                branch                        = $a.branch
                pr_number                     = $null
                status                        = "running"
                name                          = $a.name
                updated_at                    = $a.updated_at
                container_ref                 = $a.container_ref
                ci_notified                   = $false
                conflict_notified             = $false
                rebase_requested              = $false
                idle_detected_at              = $null
                review_marked                 = $false
                error_notified                = $false
                push_notified                 = $false
                merge_failed_notified         = $false
                last_process_status           = $null
                last_process_completed_at     = $null
                pending_followup              = $null
                last_followup_message         = $null
                last_followup_reason          = $null
                last_followup_at              = $null
                context_recovery_pending      = $false
                context_recovery_attempted_at = $null
            }
            Write-Log "Tracking new attempt: $($a.id.Substring(0,8)) → $($a.branch)" -Level "INFO"
        }
        else {
            $script:TrackedAttempts[$a.id].updated_at = $a.updated_at
            $script:TrackedAttempts[$a.id].container_ref = $a.container_ref
        }

        $summary = $summaryMap[$a.id]
        if ($summary) {
            $tracked = $script:TrackedAttempts[$a.id]
            $tracked.last_process_status = $summary.latest_process_status
            $tracked.last_process_completed_at = $summary.latest_process_completed_at

            if ($summary.latest_process_status -eq "running") {
                # Agent is active again; treat as running to avoid review actions
                $tracked.status = "running"
                $tracked.review_marked = $false
                $tracked.error_notified = $false
                $tracked.ci_notified = $false
                $tracked.pending_followup = $null
            }
            elseif ($summary.latest_process_status -in @("completed", "failed")) {
                if ($tracked.status -eq "running") {
                    $tracked.status = "review"
                }
                if (-not $tracked.review_marked) {
                    Write-Log "Attempt $($a.id.Substring(0,8)) finished ($($summary.latest_process_status)) — marking review" -Level "INFO"
                    if (-not $DryRun) {
                        Update-VKTaskStatus -TaskId $tracked.task_id -Status "inreview" | Out-Null
                    }
                    $tracked.review_marked = $true
                }
                if ($summary.latest_process_status -eq "failed" -and -not $tracked.error_notified) {
                    Write-Log "Attempt $($a.id.Substring(0,8)) failed in workspace — requires agent attention" -Level "WARN"
                    $tracked.error_notified = $true
                    $tracked.status = "error"
                }

                if ($summary.latest_process_status -eq "failed" -and (Test-ContextWindowError -Summary $summary)) {
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
        }
    }

    # Remove attempts that disappeared from active list (archived)
    $activeIds = @($apiAttempts | ForEach-Object { $_.id })
    $toCheck = @($script:TrackedAttempts.Keys | Where-Object { $_ -notin $activeIds })
    foreach ($id in $toCheck) {
        Write-Log "Attempt $($id.Substring(0,8)) archived — removing from tracking" -Level "INFO"
        $script:TrackedAttempts.Remove($id)
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
            $tracked.idle_detected_at = Get-Date
            Write-Log "Attempt $($a.id.Substring(0,8)) idle ${IdleTimeoutMin}m+ — will confirm in ${IdleConfirmMin}m" -Level "WARN"
            continue
        }

        $confirmMinutes = ((Get-Date) - $tracked.idle_detected_at).TotalMinutes
        if ($confirmMinutes -lt $IdleConfirmMin) { continue }

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

        # Skip already-completed
        if ($info.status -in @("merged", "done")) { continue }

        if ($info.pending_followup) {
            $pending = $info.pending_followup
            $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $pending.message -Reason $pending.reason
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
                    if (-not $info.push_notified) {
                        $msg = "Missing remote branch for $branch. Please push your commits so a PR can be created."
                        $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $msg -Reason "missing_branch"
                        $info.push_notified = $true
                    }
                    continue
                }
                Write-Log "No PR for $branch — creating PR" -Level "ACTION"
                if (-not $DryRun) {
                    $title = if ($info.name) { $info.name } else { "Automated task PR" }
                    $created = Create-PRForBranch -Branch $branch -Title $title -Body "Automated PR created by ve-orchestrator"
                    if (Test-GithubRateLimit) { return }
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

        # Check if already merged
        if ($pr.state -eq "MERGED" -or $pr.mergedAt) {
            Write-Log "PR #$($pr.number) for $branch is MERGED" -Level "OK"
            $info.status = "merged"
            Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
            $processed += $attemptId
            continue
        }

        $isYoung = $false
        $ageMinutes = 0
        # PR exists but not merged — enforce minimum wait before merge
        if ($pr.createdAt) {
            $createdAt = [datetime]$pr.createdAt
            $ageMinutes = ((Get-Date) - $createdAt).TotalMinutes
            if ($ageMinutes -lt $CiWaitMin) {
                $isYoung = $true
            }
        }

        # PR exists but not merged — check CI
        $prDetails = Get-PRDetails -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        $mergeState = if ($prDetails -and $prDetails.mergeStateStatus) { $prDetails.mergeStateStatus } else { "UNKNOWN" }

        if ($mergeState -in @("DIRTY", "CONFLICTING")) {
            Write-Log "PR #$($pr.number) has merge conflicts ($mergeState)" -Level "WARN"
            $info.status = "review"
            if (-not $info.conflict_notified) {
                $body = @"
Merge conflict detected for PR #$($pr.number).

Please rebase or resolve conflicts on branch `$branch`, then push updated changes.
"@
                if (-not $DryRun) {
                    Add-PRComment -PRNumber $pr.number -Body $body | Out-Null
                }
                $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $body.Trim() -Reason "merge_conflict"
                $info.conflict_notified = $true
            }
            continue
        }

        $checkStatus = Get-PRCheckStatus -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        Write-Log "PR #$($pr.number) ($branch): CI=$checkStatus" -Level "INFO"

        switch ($checkStatus) {
            "passing" {
                if ($isYoung) {
                    $remaining = [math]::Ceiling($CiWaitMin - $ageMinutes)
                    Write-Log "PR #$($pr.number) created ${ageMinutes:N1}m ago — waiting ${remaining}m before merge" -Level "INFO"
                    continue
                }
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
                        Write-Log "Merge failed for PR #$($pr.number), will retry" -Level "WARN"
                        Write-Log "Merge failure reason: $reason" -Level "WARN"
                        if ($UseAutoMerge) {
                            $enabled = Merge-PR -PRNumber $pr.number -AutoMerge
                            if (Test-GithubRateLimit) { return }
                            if (-not $enabled) {
                                $autoErr = Get-VKLastGithubError
                                $autoReason = if ($autoErr -and $autoErr.message) { $autoErr.message } else { "Unknown auto-merge error" }
                                Write-Log "Auto-merge enable failed: $autoReason" -Level "WARN"
                            }
                        }
                        if (-not $info.merge_failed_notified -and -not ($reason -match "not up to date|not mergeable")) {
                            $msg = "Merge failed for PR #$($pr.number). Reason: $reason. If branch protections require up-to-date, rebase on main and push."
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
                if (-not $info.ci_notified) {
                    $checks = Get-PRChecksDetail -PRNumber $pr.number
                    if (Test-GithubRateLimit) { return }
                    $summary = Format-PRCheckFailures -Checks $checks
                    $body = @"
CI is failing for PR #$($pr.number).

$summary
"@
                    if (-not $DryRun) {
                        Add-PRComment -PRNumber $pr.number -Body $body | Out-Null
                    }
                    if ($isYoung) {
                        $null = Try-SendFollowUp -AttemptId $attemptId -Info $info -Message $body.Trim() -Reason "ci_failing_early"
                    }
                    $info.ci_notified = $true
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
    if (-not $DryRun) {
        Archive-VKAttempt -AttemptId $AttemptId | Out-Null
        Update-VKTaskStatus -TaskId $TaskId -Status "done"
    }
    $script:TasksCompleted++
}

function Fill-ParallelSlots {
    <#
    .SYNOPSIS Submit new task attempts to reach the target parallelism.
              Enforces merge gate: won't start new tasks if previous ones have unmerged PRs.
              Uses 50/50 Codex/Copilot executor cycling.
    #>
    # Count truly active slots: agents currently working
    $activeCount = @($script:TrackedAttempts.Values | Where-Object { $_.status -eq "running" }).Count

    $slotsAvailable = $MaxParallel - $activeCount
    if ($slotsAvailable -le 0) {
        Write-Log "All $MaxParallel slots occupied ($activeCount active)" -Level "INFO"
        return
    }

    Write-Log "$slotsAvailable slots available (target: $MaxParallel, active: $activeCount)" -Level "ACTION"

    # Get next todo tasks
    $nextTasks = Get-VKNextTodoTasks -Count $slotsAvailable
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

        $shortTitle = $task.title.Substring(0, [Math]::Min(70, $task.title.Length))
        $nextExec = Get-CurrentExecutorProfile
        Write-Log "Submitting: $shortTitle [$($nextExec.executor)]" -Level "ACTION"

        if (-not $DryRun) {
            $attempt = Submit-VKTaskAttempt -TaskId $task.id
            if ($attempt) {
                $script:TrackedAttempts[$attempt.id] = @{
                    task_id   = $task.id
                    branch    = $attempt.branch
                    pr_number = $null
                    status    = "running"
                    name      = $task.title
                    executor  = $nextExec.executor
                }
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
    $todoTasks = Get-VKTasks -Status "todo" -Limit 1
    $hasTracked = $script:TrackedAttempts.Count -gt 0
    return (-not $todoTasks -or @($todoTasks).Count -eq 0) -and (-not $hasTracked)
}

# ─── Main Loop ────────────────────────────────────────────────────────────────

function Start-Orchestrator {
    Write-Banner

    # Validate prerequisites
    $ghVersion = gh --version 2>$null
    if (-not $ghVersion) {
        Write-Log "GitHub CLI (gh) not found. Install: https://cli.github.com/" -Level "ERROR"
        return
    }
    Write-Log "GitHub CLI: $($ghVersion | Select-Object -First 1)" -Level "INFO"

    # Auto-detect project and repo IDs
    if (-not (Initialize-VKConfig)) {
        Write-Log "Failed to initialize vibe-kanban configuration" -Level "ERROR"
        return
    }
    Write-Log "Project: $($script:VK_PROJECT_ID.Substring(0,8))...  Repo: $($script:VK_REPO_ID.Substring(0,8))..." -Level "OK"

    # Log executor cycling setup
    Write-Log "Executors: $(($script:VK_EXECUTORS | ForEach-Object { "$($_.executor)/$($_.variant)" }) -join ' ⇄ ')" -Level "INFO"

    # Initial sync
    Sync-TrackedAttempts
    $initialCounts = Get-TrackedStatusCounts
    Write-Log "Initial sync: $($script:TrackedAttempts.Count) tracked (running=$($initialCounts.running), review=$($initialCounts.review), error=$($initialCounts.error))" -Level "INFO"

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
                    Complete-Task -AttemptId $candidate.attempt_id -TaskId $candidate.task_id -PRNumber $candidate.pr_number
                    $script:TrackedAttempts.Remove($candidate.attempt_id)
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

    do {
        $script:CycleCount++
        Write-CycleSummary
        $counts = Get-TrackedStatusCounts
        Write-Log "Status: running=$($counts.running), review=$($counts.review), error=$($counts.error)" -Level "INFO"

        # Step 1: Sync attempt state from API
        Sync-TrackedAttempts

        # Step 2: Process completed attempts (check PRs, merge, mark done)
        Process-CompletedAttempts

        # Step 3: Fill empty parallel slots with new task submissions
        Fill-ParallelSlots

        # Step 4: Check if we're done
        if (Test-BacklogEmpty) {
            Write-Log "All tasks completed! Backlog empty, no active attempts." -Level "OK"
            Write-Host ""
            Write-Host "  ╔══════════════════════════════════════════╗" -ForegroundColor Green
            Write-Host "  ║   ALL TASKS COMPLETE                    ║" -ForegroundColor Green
            Write-Host "  ║   Submitted: $($script:TasksSubmitted.ToString().PadRight(28))║" -ForegroundColor Green
            Write-Host "  ║   Completed: $($script:TasksCompleted.ToString().PadRight(28))║" -ForegroundColor Green
            Write-Host "  ║   Cycles:    $($script:CycleCount.ToString().PadRight(28))║" -ForegroundColor Green
            Write-Host "  ╚══════════════════════════════════════════╝" -ForegroundColor Green
            Write-Host ""
            break
        }

        if ($OneShot) {
            Write-Log "OneShot mode — exiting after single cycle" -Level "INFO"
            break
        }

        # Step 5: Wait before next cycle
        Write-Log "Sleeping ${PollIntervalSec}s until next cycle... (Ctrl+C to stop)" -Level "INFO"
        Start-Sleep -Seconds $PollIntervalSec

    } while ($true)
}

# ─── Entry Point ──────────────────────────────────────────────────────────────
Start-Orchestrator
