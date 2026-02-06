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
    [int]$PollIntervalSec = 300,
    [int]$GitHubCooldownSec = 300,
    [switch]$DryRun,
    [switch]$OneShot
)

# ─── Load ve-kanban library ──────────────────────────────────────────────────
. "$PSScriptRoot/ve-kanban.ps1"

# ─── State tracking ──────────────────────────────────────────────────────────
$script:CycleCount = 0
$script:TasksCompleted = 0
$script:TasksSubmitted = 0
$script:StartTime = Get-Date
$script:GitHubCooldownUntil = $null

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
        "INFO"  { "White" }
        "OK"    { "Green" }
        "WARN"  { "Yellow" }
        "ERROR" { "Red" }
        "ACTION"{ "Cyan" }
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

# ─── Core Orchestration ──────────────────────────────────────────────────────

function Sync-TrackedAttempts {
    <#
    .SYNOPSIS Refresh tracked attempts from vibe-kanban API + discover new active ones.
    #>
    $apiAttempts = Get-VKAttempts -ActiveOnly
    if (-not $apiAttempts) { return }

    foreach ($a in $apiAttempts) {
        if (-not $a.branch) { continue }
        if (-not $script:TrackedAttempts.ContainsKey($a.id)) {
            # Newly discovered active attempt
            $script:TrackedAttempts[$a.id] = @{
                task_id   = $a.task_id
                branch    = $a.branch
                pr_number = $null
                status    = "running"
                name      = $a.name
            }
            Write-Log "Tracking new attempt: $($a.id.Substring(0,8)) → $($a.branch)" -Level "INFO"
        }
    }

    # Also mark attempts that disappeared from active list as potentially done
    $activeIds = @($apiAttempts | ForEach-Object { $_.id })
    $toCheck = @($script:TrackedAttempts.Keys | Where-Object { $_ -notin $activeIds })
    foreach ($id in $toCheck) {
        if ($script:TrackedAttempts[$id].status -eq "running") {
            $script:TrackedAttempts[$id].status = "agent_done"
            Write-Log "Attempt $($id.Substring(0,8)) agent finished (no longer active)" -Level "INFO"
        }
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
        if ($info.status -in @("merged","done")) { continue }

        $branch = $info.branch
        if (-not $branch) { continue }

        # Check for PR
        $pr = Get-PRForBranch -Branch $branch
        if (Test-GithubRateLimit) { return }
        if (-not $pr) {
            # No PR yet — agent might still be working, or cleanup script hasn't run
            continue
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

        # PR exists but not merged — check CI
        $checkStatus = Get-PRCheckStatus -PRNumber $pr.number
        if (Test-GithubRateLimit) { return }
        Write-Log "PR #$($pr.number) ($branch): CI=$checkStatus" -Level "INFO"

        switch ($checkStatus) {
            "passing" {
                Write-Log "CI passing for PR #$($pr.number) — merging..." -Level "ACTION"
                if (-not $DryRun) {
                    $merged = Merge-PR -PRNumber $pr.number
                    if (Test-GithubRateLimit) { return }
                    if ($merged) {
                        $info.status = "merged"
                        Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
                        $processed += $attemptId
                    } else {
                        Write-Log "Merge failed for PR #$($pr.number), will retry" -Level "WARN"
                        # Try enabling auto-merge as fallback
                        Enable-AutoMerge -PRNumber $pr.number
                        if (Test-GithubRateLimit) { return }
                    }
                } else {
                    Write-Log "[DRY-RUN] Would merge PR #$($pr.number)" -Level "ACTION"
                }
            }
            "pending" {
                # CI still running — enable auto-merge if not already
                Write-Log "CI pending for PR #$($pr.number) — enabling auto-merge" -Level "INFO"
                if (-not $DryRun) {
                    Enable-AutoMerge -PRNumber $pr.number
                    if (Test-GithubRateLimit) { return }
                }
            }
            "failing" {
                Write-Log "CI FAILING for PR #$($pr.number) — needs attention" -Level "WARN"
                # Don't block the slot — the agent or a human needs to fix this
                # We mark it so we don't keep retrying every cycle
                $info.status = "ci_failing"
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
    # Count truly active slots: running + pending CI + agent_done (waiting for PR eval)
    $activeStatuses = @("running", "agent_done", "ci_failing")
    $activeCount = @($script:TrackedAttempts.Values | Where-Object { $_.status -in $activeStatuses }).Count

    $slotsAvailable = $MaxParallel - $activeCount
    if ($slotsAvailable -le 0) {
        Write-Log "All $MaxParallel slots occupied ($activeCount active)" -Level "INFO"
        return
    }

    # ─── MERGE GATE ───────────────────────────────────────────────────────
    # Don't start new tasks if there are unmerged PRs from previous tasks.
    # This ensures sequential task confirmation before moving on.
    $unmatchedPRs = @($script:TrackedAttempts.Values | Where-Object {
        $_.status -in @("agent_done", "ci_failing") -and $_.pr_number
    })
    if ($unmatchedPRs.Count -gt 0) {
        $prNums = ($unmatchedPRs | ForEach-Object { "#$($_.pr_number)" }) -join ", "
        Write-Log "MERGE GATE: $($unmatchedPRs.Count) unmerged PR(s) ($prNums) — waiting before new submissions" -Level "WARN"
        return
    }

    # Also check for agent_done attempts without PRs yet (cleanup script may not have run)
    $waitingForPR = @($script:TrackedAttempts.Values | Where-Object {
        $_.status -eq "agent_done" -and -not $_.pr_number
    })
    if ($waitingForPR.Count -gt 0) {
        Write-Log "MERGE GATE: $($waitingForPR.Count) attempt(s) finished but no PR yet — waiting" -Level "WARN"
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
        } else {
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
    Write-Log "Initial sync: $($script:TrackedAttempts.Count) active attempts" -Level "INFO"

    do {
        $script:CycleCount++
        Write-CycleSummary

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
