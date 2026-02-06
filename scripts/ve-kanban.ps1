#!/usr/bin/env pwsh
<#
.SYNOPSIS
    VirtEngine Kanban CLI — wraps the vibe-kanban HTTP API for task + attempt management.

.DESCRIPTION
    Provides commands to list tasks, submit task attempts, check attempt status,
    rebase attempts, merge PRs, and mark tasks as complete. Designed to be both
    a standalone CLI and a dot-sourceable library for ve-orchestrator.ps1.

.EXAMPLE
    # List todo tasks
    ./ve-kanban.ps1 list --status todo

    # Submit the next N tasks as attempts
    ./ve-kanban.ps1 submit-next --count 2

    # Show active attempts
    ./ve-kanban.ps1 status

    # Merge a completed PR
    ./ve-kanban.ps1 merge --branch ve/abc1-feat-portal

    # Run orchestration loop
    ./ve-kanban.ps1 orchestrate --parallel 2
#>

# ─── Configuration ────────────────────────────────────────────────────────────
$script:VK_BASE_URL        = $env:VK_BASE_URL        ?? "http://127.0.0.1:54089"
$script:VK_PROJECT_NAME    = $env:VK_PROJECT_NAME    ?? "virtengine"
$script:VK_PROJECT_ID      = $env:VK_PROJECT_ID      ?? ""   # Auto-detected if empty
$script:VK_REPO_ID         = $env:VK_REPO_ID         ?? ""   # Auto-detected if empty
$script:GH_OWNER           = $env:GH_OWNER           ?? "virtengine-gh"
$script:GH_REPO            = $env:GH_REPO            ?? "virtengine"
$script:VK_TARGET_BRANCH   = $env:VK_TARGET_BRANCH   ?? "origin/main"
$script:VK_INITIALIZED     = $false

# Executor profiles (used for 50/50 cycling between Codex and Copilot)
$script:VK_EXECUTORS = @(
    @{ executor = "CODEX";   variant = "DEFAULT" }
    @{ executor = "COPILOT"; variant = "CLAUDE_OPUS_4_6" }
)
$script:VK_EXECUTOR_INDEX = 0   # Tracks cycling state

# ─── HTTP Helpers ─────────────────────────────────────────────────────────────

function Invoke-VKApi {
    <#
    .SYNOPSIS Invoke the vibe-kanban REST API.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Path,
        [string]$Method = "GET",
        [object]$Body
    )
    $uri = "$script:VK_BASE_URL$Path"
    $params = @{ Uri = $uri; Method = $Method; ContentType = "application/json"; UseBasicParsing = $true }
    if ($Body) { $params.Body = ($Body | ConvertTo-Json -Depth 10 -Compress) }
    try {
        $raw = Invoke-WebRequest @params -ErrorAction Stop
        $resp = $raw.Content | ConvertFrom-Json -Depth 20
        if ($resp.success -eq $false) {
            Write-Warning "API error on $Path : $($resp.message)"
            return $null
        }
        return $resp.data ?? $resp
    } catch {
        Write-Error "HTTP $Method $uri failed: $_"
        return $null
    }
}

# ─── Auto-Detection & Initialization ─────────────────────────────────────────

function Initialize-VKConfig {
    <#
    .SYNOPSIS Auto-detect project ID and repo ID from vibe-kanban by project name.
              Works for any user with their own vibe-kanban setup.
    #>
    if ($script:VK_INITIALIZED) { return $true }

    # If project ID already set (via env), skip auto-detection for project
    if ($script:VK_PROJECT_ID) {
        Write-Verbose "Using configured project ID: $script:VK_PROJECT_ID"
    } else {
        Write-Host "  Auto-detecting project '$script:VK_PROJECT_NAME'..." -ForegroundColor DarkGray
        $projects = Invoke-VKApi -Path "/api/projects"
        if (-not $projects) {
            Write-Error "Cannot connect to vibe-kanban at $script:VK_BASE_URL"
            return $false
        }
        $projectList = if ($projects -is [System.Array]) { $projects } elseif ($projects.projects) { $projects.projects } else { @($projects) }

        # Find project by name (case-insensitive)
        $match = $projectList | Where-Object {
            $_.name -ieq $script:VK_PROJECT_NAME -or
            $_.display_name -ieq $script:VK_PROJECT_NAME -or
            $_.title -ieq $script:VK_PROJECT_NAME
        } | Select-Object -First 1

        if (-not $match) {
            Write-Error "No project named '$script:VK_PROJECT_NAME' found. Available: $($projectList | ForEach-Object { $_.name ?? $_.display_name ?? $_.title } | Join-String -Separator ', ')"
            return $false
        }
        $script:VK_PROJECT_ID = $match.id
        Write-Host "  ✓ Project: $($match.name ?? $match.display_name ?? $match.title) ($($script:VK_PROJECT_ID.Substring(0,8))...)" -ForegroundColor Green
    }

    # Auto-detect repo ID if not set
    if (-not $script:VK_REPO_ID) {
        Write-Host "  Auto-detecting repository..." -ForegroundColor DarkGray
        $repos = Invoke-VKApi -Path "/api/repos?project_id=$script:VK_PROJECT_ID"
        if (-not $repos) {
            # Try alternate endpoint
            $repos = Invoke-VKApi -Path "/api/projects/$script:VK_PROJECT_ID/repos"
        }
        $repoList = if ($repos -is [System.Array]) { $repos } elseif ($repos.repos) { $repos.repos } else { @($repos) }

        if ($repoList -and @($repoList).Count -gt 0) {
            # Prefer repo matching GH_REPO name, otherwise take first
            $repoMatch = $repoList | Where-Object { $_.name -ieq $script:GH_REPO } | Select-Object -First 1
            if (-not $repoMatch) { $repoMatch = $repoList | Select-Object -First 1 }
            $script:VK_REPO_ID = $repoMatch.id
            Write-Host "  ✓ Repo: $($repoMatch.name ?? 'default') ($($script:VK_REPO_ID.Substring(0,8))...)" -ForegroundColor Green
        } else {
            Write-Warning "No repos found for project. Repo ID must be set via VK_REPO_ID env var."
            return $false
        }
    }

    $script:VK_INITIALIZED = $true
    return $true
}

# ─── Executor Cycling ─────────────────────────────────────────────────────────

function Get-NextExecutorProfile {
    <#
    .SYNOPSIS Get the next executor profile in the 50/50 Codex/Copilot rotation.
    #>
    $profile = $script:VK_EXECUTORS[$script:VK_EXECUTOR_INDEX]
    $script:VK_EXECUTOR_INDEX = ($script:VK_EXECUTOR_INDEX + 1) % $script:VK_EXECUTORS.Count
    return $profile
}

function Get-CurrentExecutorProfile {
    <#
    .SYNOPSIS Peek at the next executor profile without advancing the cycle.
    #>
    return $script:VK_EXECUTORS[$script:VK_EXECUTOR_INDEX]
}

# ─── Task Functions ───────────────────────────────────────────────────────────

function Get-VKTasks {
    <#
    .SYNOPSIS List tasks with optional status filter (client-side filtering).
    #>
    [CmdletBinding()]
    param(
        [ValidateSet("todo","inprogress","inreview","done","cancelled")]
        [string]$Status,
        [int]$Limit = 500
    )
    if (-not (Initialize-VKConfig)) { return @() }
    $result = Invoke-VKApi -Path "/api/tasks?project_id=$script:VK_PROJECT_ID"
    if (-not $result) { return @() }
    # Result is either the tasks array directly, or an object with .tasks
    $tasks = if ($result -is [System.Array]) { $result } elseif ($result.tasks) { $result.tasks } else { @($result) }
    # API doesn't filter server-side, so filter client-side
    if ($Status -and $tasks) {
        $tasks = @($tasks | Where-Object { $_.status -eq $Status })
    }
    return $tasks
}

function Get-VKTask {
    <#
    .SYNOPSIS Get a single task by ID.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$TaskId)
    return Invoke-VKApi -Path "/api/tasks/$TaskId"
}

function Update-VKTaskStatus {
    <#
    .SYNOPSIS Update a task's status.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$TaskId,
        [Parameter(Mandatory)]
        [ValidateSet("todo","inprogress","inreview","done","cancelled")]
        [string]$Status
    )
    return Invoke-VKApi -Path "/api/tasks/$TaskId" -Method "PATCH" -Body @{ status = $Status }
}

function Get-VKNextTodoTasks {
    <#
    .SYNOPSIS Get the next N tasks in 'todo' status, ordered by creation date (earliest first = highest priority).
    #>
    [CmdletBinding()]
    param([int]$Count = 1)
    $tasks = Get-VKTasks -Status "todo" -Limit $Count
    if (-not $tasks) { return @() }
    # Return as array
    $arr = @($tasks)
    return $arr | Select-Object -First $Count
}

# ─── Attempt Functions ────────────────────────────────────────────────────────

function Get-VKAttempts {
    <#
    .SYNOPSIS List all task attempts.
    #>
    [CmdletBinding()]
    param([switch]$ActiveOnly)
    $result = Invoke-VKApi -Path "/api/task-attempts"
    if (-not $result) { return @() }
    $attempts = if ($result -is [System.Array]) { @($result) } else { @($result) }
    if ($ActiveOnly) {
        $attempts = @($attempts | Where-Object { -not $_.archived })
    }
    return $attempts
}

function Submit-VKTaskAttempt {
    <#
    .SYNOPSIS Submit a task as a new attempt (creates worktree + starts agent).
              Uses the next executor in the Codex/Copilot rotation cycle.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$TaskId,
        [string]$TargetBranch = $script:VK_TARGET_BRANCH,
        [hashtable]$ExecutorOverride
    )
    if (-not (Initialize-VKConfig)) { return $null }

    # Use override if provided, otherwise cycle to next executor
    $execProfile = if ($ExecutorOverride) { $ExecutorOverride } else { Get-NextExecutorProfile }

    $body = @{
        task_id = $TaskId
        repos = @(
            @{
                repo_id       = $script:VK_REPO_ID
                target_branch = $TargetBranch
            }
        )
        executor_profile_id = @{
            executor = $execProfile.executor
            variant  = $execProfile.variant
        }
    }
    Write-Host "  Submitting attempt for task $TaskId ($($execProfile.executor)/$($execProfile.variant)) ..." -ForegroundColor Cyan
    $result = Invoke-VKApi -Path "/api/task-attempts" -Method "POST" -Body $body
    if ($result) {
        Write-Host "  ✓ Attempt created: $($result.id) → branch $($result.branch)" -ForegroundColor Green
    }
    return $result
}

function Invoke-VKRebase {
    <#
    .SYNOPSIS Rebase an attempt's branch onto the latest target.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$AttemptId,
        [string]$BaseBranch = $script:VK_TARGET_BRANCH
    )
    $body = @{
        repo_id         = $script:VK_REPO_ID
        old_base_branch = $BaseBranch
        new_base_branch = $BaseBranch
    }
    return Invoke-VKApi -Path "/api/task-attempts/$AttemptId/rebase" -Method "POST" -Body $body
}

# ─── GitHub PR Functions ─────────────────────────────────────────────────────

function Get-PRForBranch {
    <#
    .SYNOPSIS Find an open or merged PR for a given branch.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Branch)
    $prJson = gh pr list --head $Branch --repo "$script:GH_OWNER/$script:GH_REPO" --json number,state,title,mergeable,statusCheckRollup,mergedAt,url --limit 1 2>$null
    if (-not $prJson -or $prJson -eq "[]") {
        # Also check merged/closed
        $prJson = gh pr list --head $Branch --repo "$script:GH_OWNER/$script:GH_REPO" --state merged --json number,state,title,mergedAt,url --limit 1 2>$null
    }
    if (-not $prJson -or $prJson -eq "[]") { return $null }
    return ($prJson | ConvertFrom-Json) | Select-Object -First 1
}

function Get-PRCheckStatus {
    <#
    .SYNOPSIS Get CI check status for a PR.
    .OUTPUTS "passing", "failing", "pending", or "unknown"
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $checksJson = gh pr checks $PRNumber --repo "$script:GH_OWNER/$script:GH_REPO" --json name,state,status 2>$null
    if (-not $checksJson) { return "unknown" }
    $checks = $checksJson | ConvertFrom-Json
    if (-not $checks -or $checks.Count -eq 0) { return "unknown" }

    $failing  = $checks | Where-Object { $_.state -eq "FAILURE" -or $_.state -eq "ERROR" }
    $pending  = $checks | Where-Object { $_.state -eq "PENDING" -or $_.status -eq "IN_PROGRESS" -or $_.status -eq "QUEUED" }

    if ($failing.Count -gt 0) { return "failing" }
    if ($pending.Count -gt 0) { return "pending" }
    return "passing"
}

function Merge-PR {
    <#
    .SYNOPSIS Merge a PR after rebase onto latest main. Returns $true on success.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [switch]$AutoMerge
    )
    if ($AutoMerge) {
        Write-Host "  Enabling auto-merge for PR #$PRNumber ..." -ForegroundColor Yellow
        gh pr merge $PRNumber --repo "$script:GH_OWNER/$script:GH_REPO" --auto --merge --delete-branch 2>&1
        return $?
    } else {
        Write-Host "  Merging PR #$PRNumber ..." -ForegroundColor Yellow
        gh pr merge $PRNumber --repo "$script:GH_OWNER/$script:GH_REPO" --merge --delete-branch 2>&1
        return $?
    }
}

function Enable-AutoMerge {
    <#
    .SYNOPSIS Enable auto-merge on a PR so it merges when CI passes.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    gh pr merge $PRNumber --repo "$script:GH_OWNER/$script:GH_REPO" --auto --merge --delete-branch 2>&1
    return $?
}

# ─── Orchestration Helpers ────────────────────────────────────────────────────

function Get-ActiveAttemptCount {
    <#
    .SYNOPSIS Count non-archived attempts that have running agents.
    #>
    $attempts = Get-VKAttempts -ActiveOnly
    return @($attempts).Count
}

function Get-AttemptTaskMap {
    <#
    .SYNOPSIS Build a hashtable mapping task_id → latest attempt for active (non-archived) attempts.
    #>
    $attempts = Get-VKAttempts -ActiveOnly
    $map = @{}
    foreach ($a in $attempts) {
        $map[$a.task_id] = $a
    }
    return $map
}

function Test-AttemptComplete {
    <#
    .SYNOPSIS Check if an attempt has completed (branch pushed + PR exists or is merged).
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][object]$Attempt)
    if (-not $Attempt.branch) { return $false }
    $pr = Get-PRForBranch -Branch $Attempt.branch
    return ($null -ne $pr)
}

# ─── Display Helpers ──────────────────────────────────────────────────────────

function Show-Tasks {
    [CmdletBinding()]
    param(
        [ValidateSet("todo","inprogress","inreview","done","cancelled")]
        [string]$Status = "todo"
    )
    $tasks = Get-VKTasks -Status $Status
    if (-not $tasks) {
        Write-Host "  No tasks with status '$Status'" -ForegroundColor Yellow
        return
    }
    Write-Host ""
    Write-Host "  ┌─ $($Status.ToUpper()) Tasks ($(@($tasks).Count)) ─────────────────" -ForegroundColor Cyan
    foreach ($t in $tasks) {
        $age = if ($t.created_at) { [math]::Round(((Get-Date) - [datetime]$t.created_at).TotalDays, 0) } else { "?" }
        $inProgress = if ($t.has_in_progress_attempt) { " [RUNNING]" } else { "" }
        Write-Host "  │ $($t.id.Substring(0,8))  $($t.title)$inProgress" -ForegroundColor $(if ($t.has_in_progress_attempt) { "Yellow" } else { "White" })
    }
    Write-Host "  └───────────────────────────────────────────" -ForegroundColor Cyan
    Write-Host ""
}

function Show-Status {
    if (-not (Initialize-VKConfig)) { return }
    $active = Get-VKAttempts -ActiveOnly
    $todoTasks = Get-VKTasks -Status "todo"
    $inProgressTasks = Get-VKTasks -Status "inprogress"

    Write-Host ""
    Write-Host "  ╔═══════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║     VirtEngine Kanban Status              ║" -ForegroundColor Cyan
    Write-Host "  ╚═══════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  Todo:        $(@($todoTasks).Count) tasks" -ForegroundColor White
    Write-Host "  In-Progress: $(@($inProgressTasks).Count) tasks" -ForegroundColor Yellow
    Write-Host "  Active Attempts: $(@($active).Count)" -ForegroundColor Green
    Write-Host ""

    if ($active -and @($active).Count -gt 0) {
        Write-Host "  ┌─ Active Attempts ──────────────────────────" -ForegroundColor Green
        foreach ($a in $active) {
            $name = if ($a.name) { $a.name.Substring(0, [Math]::Min(60, $a.name.Length)) } else { "(unnamed)" }
            $pr = Get-PRForBranch -Branch $a.branch 2>$null
            $prStatus = if ($pr) { "PR #$($pr.number) ($($pr.state))" } else { "No PR" }
            Write-Host "  │ $($a.id.Substring(0,8))  $($a.branch)" -ForegroundColor White
            Write-Host "  │   $name" -ForegroundColor DarkGray
            Write-Host "  │   $prStatus" -ForegroundColor $(if ($pr -and $pr.state -eq "MERGED") { "Green" } elseif ($pr) { "Yellow" } else { "DarkGray" })
        }
        Write-Host "  └────────────────────────────────────────────" -ForegroundColor Green
    }
    Write-Host ""
}

# ─── CLI Dispatch ─────────────────────────────────────────────────────────────

function Invoke-CLI {
    param([string[]]$Arguments)

    if ($Arguments.Count -eq 0) {
        Show-Usage
        return
    }

    $command = $Arguments[0]
    $rest = if ($Arguments.Count -gt 1) { $Arguments[1..($Arguments.Count-1)] } else { @() }

    switch ($command) {
        "list" {
            $status = "todo"
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--status","-s") -and ($i+1) -lt $rest.Count) { $status = $rest[$i+1] }
            }
            Show-Tasks -Status $status
        }
        "status" {
            Show-Status
        }
        "submit" {
            if ($rest.Count -eq 0) { Write-Error "Usage: ve-kanban submit <task-id>"; return }
            $taskId = $rest[0]
            Submit-VKTaskAttempt -TaskId $taskId
        }
        "submit-next" {
            $count = 1
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--count","-n") -and ($i+1) -lt $rest.Count) { $count = [int]$rest[$i+1] }
            }
            $tasks = Get-VKNextTodoTasks -Count $count
            if (-not $tasks -or @($tasks).Count -eq 0) {
                Write-Host "  No todo tasks available." -ForegroundColor Yellow
                return
            }
            foreach ($t in $tasks) {
                Write-Host "  → $($t.title)" -ForegroundColor White
                Submit-VKTaskAttempt -TaskId $t.id
            }
        }
        "rebase" {
            if ($rest.Count -eq 0) { Write-Error "Usage: ve-kanban rebase <attempt-id>"; return }
            Invoke-VKRebase -AttemptId $rest[0]
        }
        "merge" {
            $branch = $null
            $auto = $false
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--branch","-b") -and ($i+1) -lt $rest.Count) { $branch = $rest[$i+1] }
                if ($rest[$i] -eq "--auto") { $auto = $true }
            }
            if (-not $branch -and $rest.Count -gt 0 -and $rest[0] -notlike "--*") { $branch = $rest[0] }
            if (-not $branch) { Write-Error "Usage: ve-kanban merge <branch> [--auto]"; return }
            $pr = Get-PRForBranch -Branch $branch
            if (-not $pr) { Write-Error "No PR found for branch $branch"; return }
            if ($auto) {
                Enable-AutoMerge -PRNumber $pr.number
            } else {
                Merge-PR -PRNumber $pr.number
            }
        }
        "complete" {
            if ($rest.Count -eq 0) { Write-Error "Usage: ve-kanban complete <task-id>"; return }
            Update-VKTaskStatus -TaskId $rest[0] -Status "done"
            Write-Host "  ✓ Task $($rest[0]) marked as done" -ForegroundColor Green
        }
        "orchestrate" {
            $parallel = 2
            $interval = 60
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--parallel","-p") -and ($i+1) -lt $rest.Count) { $parallel = [int]$rest[$i+1] }
                if ($rest[$i] -in @("--interval","-i") -and ($i+1) -lt $rest.Count) { $interval = [int]$rest[$i+1] }
            }
            Write-Host "  Delegating to ve-orchestrator.ps1 with parallel=$parallel, interval=$interval" -ForegroundColor Cyan
            & "$PSScriptRoot/ve-orchestrator.ps1" -MaxParallel $parallel -PollIntervalSec $interval
        }
        "help" { Show-Usage }
        default {
            Write-Error "Unknown command: $command"
            Show-Usage
        }
    }
}

function Show-Usage {
    Write-Host @"

  VirtEngine Kanban CLI (ve-kanban)
  ═════════════════════════════════

  COMMANDS:
    list [--status <status>]        List tasks (default: todo)
    status                          Show dashboard (active attempts + queues)
    submit <task-id>                Submit a single task as an attempt
    submit-next [--count N]         Submit next N todo tasks as attempts
    rebase <attempt-id>             Rebase an attempt branch onto latest main
    merge <branch> [--auto]         Merge PR for a branch (--auto enables auto-merge)
    complete <task-id>              Mark a task as done
    orchestrate [--parallel N]      Run orchestration loop (default: 2 parallel)
    help                            Show this help

  ENVIRONMENT:
    VK_BASE_URL       Vibe-kanban API (default: http://127.0.0.1:54089)
    VK_PROJECT_NAME   Project name to auto-detect (default: virtengine)
    VK_PROJECT_ID     Project UUID (auto-detected if empty)
    VK_REPO_ID        Repository UUID (auto-detected if empty)
    GH_OWNER          GitHub owner (default: virtengine-gh)
    GH_REPO           GitHub repo (default: virtengine)

  EXECUTOR CYCLING:
    Alternates between CODEX/DEFAULT and COPILOT/CLAUDE_OPUS_4_6
    at 50/50 rate to avoid rate-limiting on either agent.
"@ -ForegroundColor White
}

# Run CLI if invoked directly (not dot-sourced)
if ($MyInvocation.InvocationName -ne ".") {
    Invoke-CLI -Arguments $args
}
