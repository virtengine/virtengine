#!/usr/bin/env pwsh
<#
.SYNOPSIS
    VirtEngine Task Orchestrator — automated task chaining with parallel execution.

.DESCRIPTION
    Long-running orchestration loop that:
    1. Maintains a target number of parallel task attempts
    2. Monitors agent completion (PR creation) and CI status
    3. Auto-merges PRs when CI passes
    4. Marks completed tasks as done
    5. Submits the next todo task to fill the slot
    6. Repeats until the backlog is empty

.PARAMETER MaxParallel
    Maximum number of concurrent task attempts (default: 2).

.PARAMETER PollIntervalSec
    Seconds between orchestration cycles (default: 90).

.PARAMETER DryRun
    If set, logs what would happen without making changes.

.PARAMETER OneShot
    Run a single orchestration cycle and exit (useful for testing).

.EXAMPLE
    # Run with 3 parallel agents, polling every 2 minutes
    ./ve-orchestrator.ps1 -MaxParallel 3 -PollIntervalSec 120

    # Dry-run to see what would happen
    ./ve-orchestrator.ps1 -DryRun

    # Single cycle (no loop)
    ./ve-orchestrator.ps1 -OneShot
#>
[CmdletBinding()]
param(
    [int]$MaxParallel = 2,
    [int]$PollIntervalSec = 90,
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

# Track attempts we're monitoring: attempt_id → { task_id, branch, pr_number, status }
$script:TrackedAttempts = @{}

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
    Write-Host ""
    Write-Host "  ╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║          VirtEngine Task Orchestrator                    ║" -ForegroundColor Cyan
    Write-Host "  ║                                                         ║" -ForegroundColor Cyan
    Write-Host "  ║   Parallel: $($MaxParallel.ToString().PadRight(4))  Poll: ${PollIntervalSec}s  $(if($DryRun){'DRY-RUN'}else{'LIVE'})                  ║" -ForegroundColor Cyan
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
        Write-Log "PR #$($pr.number) ($branch): CI=$checkStatus" -Level "INFO"

        switch ($checkStatus) {
            "passing" {
                Write-Log "CI passing for PR #$($pr.number) — merging..." -Level "ACTION"
                if (-not $DryRun) {
                    $merged = Merge-PR -PRNumber $pr.number
                    if ($merged) {
                        $info.status = "merged"
                        Complete-Task -AttemptId $attemptId -TaskId $info.task_id -PRNumber $pr.number
                        $processed += $attemptId
                    } else {
                        Write-Log "Merge failed for PR #$($pr.number), will retry" -Level "WARN"
                        # Try enabling auto-merge as fallback
                        Enable-AutoMerge -PRNumber $pr.number
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
    #>
    # Count truly active slots: running + pending CI + agent_done (waiting for PR eval)
    $activeStatuses = @("running", "agent_done", "ci_failing")
    $activeCount = @($script:TrackedAttempts.Values | Where-Object { $_.status -in $activeStatuses }).Count

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
        Write-Log "Submitting: $shortTitle" -Level "ACTION"

        if (-not $DryRun) {
            $attempt = Submit-VKTaskAttempt -TaskId $task.id
            if ($attempt) {
                $script:TrackedAttempts[$attempt.id] = @{
                    task_id   = $task.id
                    branch    = $attempt.branch
                    pr_number = $null
                    status    = "running"
                    name      = $task.title
                }
                $script:TasksSubmitted++
            }
        } else {
            Write-Log "[DRY-RUN] Would submit task $($task.id.Substring(0,8))" -Level "ACTION"
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

    # Test vibe-kanban API connectivity
    $projects = Invoke-VKApi -Path "/api/projects"
    if (-not $projects) {
        Write-Log "Cannot connect to vibe-kanban at $script:VK_BASE_URL" -Level "ERROR"
        return
    }
    Write-Log "Connected to vibe-kanban ($(@($projects).Count) projects)" -Level "OK"

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
