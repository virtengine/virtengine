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
$script:VK_BASE_URL = $env:VK_ENDPOINT_URL ?? $env:VK_BASE_URL ?? "http://127.0.0.1:54089"
$script:VK_PROJECT_NAME = $env:VK_PROJECT_NAME ?? "virtengine"
$script:VK_PROJECT_ID = $env:VK_PROJECT_ID ?? ""   # Auto-detected if empty
$script:VK_REPO_ID = $env:VK_REPO_ID ?? ""   # Auto-detected if empty
$script:GH_OWNER = $env:GH_OWNER ?? "virtengine"
$script:GH_REPO = $env:GH_REPO ?? "virtengine"
$script:VK_TARGET_BRANCH = $env:VK_TARGET_BRANCH ?? "origin/main"
$script:VK_INITIALIZED = $false
$script:VK_LAST_GH_ERROR = $null
$script:VK_LAST_GH_ERROR_AT = $null
$script:VK_CLI_RAW_LINE = $null

# Executor profiles — expanded pool with health tracking and model metadata
# Each profile: executor (VK executor type), variant (VK variant), tier (primary/backup/fallback),
#   provider (for grouping rate limits), max_concurrent (per-executor limit),
#   model (actual model name), region (Azure region), suitability (task fit)
$script:VK_EXECUTORS = @(
    # ── Primary tier: Main workhorses ──
    @{ executor = "CODEX"; variant = "DEFAULT"; tier = "primary"; provider = "azure_codex_52"; max_concurrent = 3;
        model = "gpt-5.2-codex"; region = "us"; suitability = @("large", "medium", "small") 
    }
    @{ executor = "COPILOT"; variant = "CLAUDE_OPUS_4_6"; tier = "primary"; provider = "github_copilot"; max_concurrent = 2;
        model = "claude-opus-4.6"; region = "global"; suitability = @("large", "medium") 
    }
    # ── Backup tier: Azure alternate deployments ──
    @{ executor = "CODEX"; variant = "GPT51_CODEX_MAX"; tier = "backup"; provider = "azure_codex_51_max"; max_concurrent = 2;
        model = "gpt-5.1-codex-max"; region = "us"; suitability = @("large", "medium") 
    }
    @{ executor = "CODEX"; variant = "GPT51_CODEX_MINI"; tier = "backup"; provider = "azure_codex_51_mini"; max_concurrent = 2;
        model = "gpt-5.1-codex-mini"; region = "us"; suitability = @("small", "medium") 
    }
    # ── Fallback tier: Session-limited CLIs ──
    @{ executor = "CODEX"; variant = "CHATGPT_AUTH"; tier = "fallback"; provider = "chatgpt_codex"; max_concurrent = 1;
        model = "gpt-5.2-codex"; region = "global"; suitability = @("medium", "small") 
    }
    @{ executor = "CODEX"; variant = "CLAUDE_CODE_CLI"; tier = "fallback"; provider = "claude_code_cli"; max_concurrent = 1;
        model = "claude-code"; region = "global"; suitability = @("large", "medium") 
    }
)
$script:VK_EXECUTOR_INDEX = 0   # Tracks cycling state

# ── Executor Health State ──────────────────────────────────────────────────────
# Maps provider → { status, degraded_at, cooldown_until, failure_count, total_timeouts,
#                    total_rate_limits, last_success_at, consecutive_failures }
$script:ExecutorHealth = @{}

function Initialize-ExecutorHealth {
    <#
    .SYNOPSIS Initialize health tracking for all executor providers.
    #>
    foreach ($exec in $script:VK_EXECUTORS) {
        $provider = $exec.provider
        if (-not $script:ExecutorHealth.ContainsKey($provider)) {
            $script:ExecutorHealth[$provider] = @{
                status               = "healthy"      # healthy | degraded | cooldown | disabled
                degraded_at          = $null
                cooldown_until       = $null
                failure_count        = 0               # failures in current window
                consecutive_failures = 0
                total_timeouts       = 0
                total_rate_limits    = 0
                total_successes      = 0
                last_success_at      = $null
                last_failure_at      = $null
                last_failure_reason  = $null
                active_tasks         = 0               # currently running tasks on this provider
            }
        }
    }
}

function Get-ExecutorHealthStatus {
    <#
    .SYNOPSIS Get health status for a provider. Returns: healthy, degraded, cooldown, disabled.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Provider)
    if (-not $script:ExecutorHealth.ContainsKey($Provider)) { return "healthy" }
    $h = $script:ExecutorHealth[$Provider]

    # Check if cooldown expired
    if ($h.status -eq "cooldown" -and $h.cooldown_until -and (Get-Date) -gt $h.cooldown_until) {
        $h.status = "degraded"  # Tentatively lift to degraded (not healthy until success)
        $h.cooldown_until = $null
        $h.consecutive_failures = [Math]::Max(0, $h.consecutive_failures - 1)
    }
    return $h.status
}

function Report-ExecutorFailure {
    <#
    .SYNOPSIS Report a failure (timeout, rate limit, error) for an executor provider.
              Automatically transitions: healthy → degraded → cooldown based on severity.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Provider,
        [ValidateSet("timeout", "rate_limit", "error", "reconnect_loop")]
        [Parameter(Mandatory)][string]$FailureType,
        [string]$Detail
    )
    Initialize-ExecutorHealth
    if (-not $script:ExecutorHealth.ContainsKey($Provider)) { return }
    $h = $script:ExecutorHealth[$Provider]

    $h.failure_count++
    $h.consecutive_failures++
    $h.last_failure_at = Get-Date
    $h.last_failure_reason = "$FailureType : $Detail"

    switch ($FailureType) {
        "timeout" { $h.total_timeouts++ }
        "rate_limit" { $h.total_rate_limits++ }
        "reconnect_loop" { $h.total_timeouts += 3 }  # Reconnect loop counts as severe
    }

    # Escalation logic
    $cooldownMinutes = 0
    if ($h.consecutive_failures -ge 5 -or $FailureType -eq "rate_limit") {
        # Hard cooldown — too many failures or explicit rate limit
        $cooldownMinutes = switch ($FailureType) {
            "rate_limit" { 45 }   # Copilot says "45 minutes" — respect it
            "reconnect_loop" { 30 }   # Severe connectivity issue
            default { 15 }   # General failure cooldown
        }
        $h.status = "cooldown"
        $h.cooldown_until = (Get-Date).AddMinutes($cooldownMinutes)
    }
    elseif ($h.consecutive_failures -ge 2) {
        # Degraded — prefer other executors but don't fully block
        $h.status = "degraded"
        $h.degraded_at = Get-Date
    }
    # Single failure stays "healthy" (transient)

    $statusMsg = "$($h.status)"
    if ($cooldownMinutes -gt 0) { $statusMsg += " (${cooldownMinutes}m)" }
    return @{ provider = $Provider; status = $h.status; cooldown_min = $cooldownMinutes; consecutive = $h.consecutive_failures }
}

function Report-ExecutorSuccess {
    <#
    .SYNOPSIS Report a successful task completion for an executor provider.
              Resets failure counters and promotes back to healthy.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Provider)
    Initialize-ExecutorHealth
    if (-not $script:ExecutorHealth.ContainsKey($Provider)) { return }
    $h = $script:ExecutorHealth[$Provider]

    $h.consecutive_failures = 0
    $h.failure_count = [Math]::Max(0, $h.failure_count - 1)
    $h.total_successes++
    $h.last_success_at = Get-Date
    if ($h.status -in @("degraded", "cooldown")) {
        $h.status = "healthy"
        $h.cooldown_until = $null
    }
}

function Get-HealthyExecutorProfile {
    <#
    .SYNOPSIS Get the best available executor, preferring healthy primary > healthy backup > degraded > fallback.
              Skips executors in cooldown or over their max_concurrent limit.
    #>
    [CmdletBinding()]
    param([int]$CurrentActiveTasks = 0)
    Initialize-ExecutorHealth

    # Build scored list: tier priority + health bonus
    $tierOrder = @{ "primary" = 0; "backup" = 1; "fallback" = 2 }
    $healthOrder = @{ "healthy" = 0; "degraded" = 1; "cooldown" = 2; "disabled" = 3 }

    $candidates = @()
    foreach ($exec in $script:VK_EXECUTORS) {
        $provider = $exec.provider
        $health = Get-ExecutorHealthStatus -Provider $provider
        if ($health -eq "disabled") { continue }
        if ($health -eq "cooldown") { continue }  # Skip cooldown executors entirely

        $h = $script:ExecutorHealth[$provider]
        $active = if ($h) { $h.active_tasks } else { 0 }
        if ($active -ge $exec.max_concurrent) { continue }  # Over capacity

        $tierScore = $tierOrder[$exec.tier] * 10
        $healthScore = $healthOrder[$health] * 5
        $loadScore = $active  # Prefer less loaded

        $candidates += @{
            executor = $exec
            score    = $tierScore + $healthScore + $loadScore
            health   = $health
            active   = $active
        }
    }

    if ($candidates.Count -eq 0) {
        # All executors exhausted — return least-bad option (lowest consecutive failures)
        $leastBad = $script:VK_EXECUTORS | Sort-Object {
            $p = $_.provider
            if ($script:ExecutorHealth.ContainsKey($p)) { $script:ExecutorHealth[$p].consecutive_failures } else { 0 }
        } | Select-Object -First 1
        return $leastBad
    }

    $best = ($candidates | Sort-Object { $_.score } | Select-Object -First 1).executor
    return $best
}

function Get-ExecutorProviderForAttempt {
    <#
    .SYNOPSIS Look up the provider string for an executor/variant pair.
    #>
    [CmdletBinding()]
    param(
        [string]$Executor,
        [string]$Variant
    )
    $match = $script:VK_EXECUTORS | Where-Object {
        $_.executor -eq $Executor -and $_.variant -eq $Variant
    } | Select-Object -First 1
    if ($match) { return $match.provider }
    # Fallback: construct from executor name
    return "$($Executor)_$($Variant)".ToLower()
}

function Increment-ExecutorActiveCount {
    <#
    .SYNOPSIS Increment active task count for a provider.
    #>
    param([Parameter(Mandatory)][string]$Provider)
    Initialize-ExecutorHealth
    if ($script:ExecutorHealth.ContainsKey($Provider)) {
        $script:ExecutorHealth[$Provider].active_tasks++
    }
}

function Decrement-ExecutorActiveCount {
    <#
    .SYNOPSIS Decrement active task count for a provider.
    #>
    param([Parameter(Mandatory)][string]$Provider)
    Initialize-ExecutorHealth
    if ($script:ExecutorHealth.ContainsKey($Provider)) {
        $script:ExecutorHealth[$Provider].active_tasks = [Math]::Max(0, $script:ExecutorHealth[$Provider].active_tasks - 1)
    }
}

function Get-ExecutorHealthSummary {
    <#
    .SYNOPSIS Get a compact summary of all executor health for logging/status.
    #>
    Initialize-ExecutorHealth
    $parts = @()
    foreach ($exec in $script:VK_EXECUTORS) {
        $provider = $exec.provider
        if ($parts -contains $provider) { continue }  # Dedup
        $h = $script:ExecutorHealth[$provider]
        $status = Get-ExecutorHealthStatus -Provider $provider
        $icon = switch ($status) {
            "healthy" { "✓" }
            "degraded" { "⚠" }
            "cooldown" { "⏸" }
            "disabled" { "✗" }
            default { "?" }
        }
        $active = if ($h) { $h.active_tasks } else { 0 }
        $parts += "$icon $($exec.executor)/$($exec.variant)($active)"
    }
    return $parts -join " | "
}

# ─── Executor Cycling ─────────────────────────────────────────────────────────

function Get-NextExecutorProfile {
    <#
    .SYNOPSIS Get the next executor profile, preferring healthy executors.
              Falls back to simple round-robin only if health system has no preference.
    #>
    $healthy = Get-HealthyExecutorProfile
    if ($healthy) {
        # Advance index past any matching entry to avoid re-selecting
        for ($i = 0; $i -lt $script:VK_EXECUTORS.Count; $i++) {
            if ($script:VK_EXECUTORS[$script:VK_EXECUTOR_INDEX].provider -eq $healthy.provider) {
                $script:VK_EXECUTOR_INDEX = ($script:VK_EXECUTOR_INDEX + 1) % $script:VK_EXECUTORS.Count
                break
            }
            $script:VK_EXECUTOR_INDEX = ($script:VK_EXECUTOR_INDEX + 1) % $script:VK_EXECUTORS.Count
        }
        return $healthy
    }
    # Fallback: simple round-robin (legacy behavior)
    $profile = $script:VK_EXECUTORS[$script:VK_EXECUTOR_INDEX]
    $script:VK_EXECUTOR_INDEX = ($script:VK_EXECUTOR_INDEX + 1) % $script:VK_EXECUTORS.Count
    return $profile
}

function Get-CurrentExecutorProfile {
    <#
    .SYNOPSIS Peek at the next executor profile without advancing the cycle.
              Returns the healthiest available executor.
    #>
    $healthy = Get-HealthyExecutorProfile
    if ($healthy) { return $healthy }
    return $script:VK_EXECUTORS[$script:VK_EXECUTOR_INDEX]
}

# ─── Azure Region Management ─────────────────────────────────────────────────
# Manages Codex config.toml switching between US and Sweden Azure regions.
# Sweden env vars: AZURE_SWEDEN_API_KEY, AZURE_SWEDEN_ENDPOINT

$script:CodexConfigPath = Join-Path $env:USERPROFILE ".codex" "config.toml"
$script:ActiveRegion = "us"                    # Current active region: us | sweden
$script:RegionSwitchedAt = $null               # When last region switch happened
$script:RegionCooldownMinutes = 120            # How long before auto-switching back to US
$script:RegionOverride = $null                 # Manual override: $null (auto) | "us" | "sweden"

# Original US config values (captured on first load)
$script:USEndpoint = $null
$script:USEnvKey = $null

function Get-ActiveCodexRegion {
    <#
    .SYNOPSIS Get the currently active Codex region.
    #>
    return $script:ActiveRegion
}

function Initialize-CodexRegionTracking {
    <#
    .SYNOPSIS Capture the current US config values for later restoration.
    #>
    if ($script:USEndpoint) { return }  # Already initialized
    if (-not (Test-Path $script:CodexConfigPath)) {
        Write-Warning "Codex config.toml not found at $($script:CodexConfigPath)"
        return
    }
    $content = Get-Content $script:CodexConfigPath -Raw
    # Parse base_url from [model_providers.azure]
    if ($content -match 'base_url\s*=\s*"([^"]+)"') {
        $script:USEndpoint = $Matches[1]
    }
    if ($content -match 'env_key\s*=\s*"([^"]+)"') {
        $script:USEnvKey = $Matches[1]
    }
    $script:ActiveRegion = "us"
}

function Switch-CodexRegion {
    <#
    .SYNOPSIS Switch Codex config.toml between US and Sweden Azure regions.
    .PARAMETER Region Target region: "us" or "sweden"
    .PARAMETER Force  Bypass cooldown checks
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][ValidateSet("us", "sweden")][string]$Region,
        [switch]$Force
    )
    Initialize-CodexRegionTracking
    if ($Region -eq $script:ActiveRegion -and -not $Force) {
        return @{ changed = $false; region = $Region; reason = "already active" }
    }

    if (-not (Test-Path $script:CodexConfigPath)) {
        return @{ changed = $false; region = $script:ActiveRegion; reason = "config.toml not found" }
    }

    $content = Get-Content $script:CodexConfigPath -Raw
    $backupPath = "$($script:CodexConfigPath).bak-$($script:ActiveRegion)"

    # Backup current config before switching
    Copy-Item $script:CodexConfigPath $backupPath -Force -ErrorAction SilentlyContinue

    if ($Region -eq "sweden") {
        $swedenEndpoint = $env:AZURE_SWEDEN_ENDPOINT
        $swedenApiKey = $env:AZURE_SWEDEN_API_KEY
        if (-not $swedenEndpoint -or -not $swedenApiKey) {
            return @{ changed = $false; region = $script:ActiveRegion; reason = "AZURE_SWEDEN_ENDPOINT or AZURE_SWEDEN_API_KEY not set" }
        }

        # Ensure endpoint ends with /openai/v1
        if ($swedenEndpoint -notmatch '/openai/v1$') {
            $swedenEndpoint = $swedenEndpoint.TrimEnd('/') + '/openai/v1'
        }

        # Replace base_url  
        $content = $content -replace '(base_url\s*=\s*")[^"]+(")', "`${1}$swedenEndpoint`${2}"
        # Replace env_key to use Sweden key  
        $content = $content -replace '(env_key\s*=\s*")[^"]+(")', '${1}AZURE_SWEDEN_API_KEY${2}'

        Set-Content -Path $script:CodexConfigPath -Value $content -Encoding UTF8 -NoNewline
        $script:ActiveRegion = "sweden"
        $script:RegionSwitchedAt = Get-Date
    }
    elseif ($Region -eq "us") {
        if (-not $script:USEndpoint -or -not $script:USEnvKey) {
            # Try to restore from backup
            $usBak = "$($script:CodexConfigPath).bak-us"
            if (Test-Path $usBak) {
                Copy-Item $usBak $script:CodexConfigPath -Force
                $script:ActiveRegion = "us"
                $script:RegionSwitchedAt = Get-Date
                return @{ changed = $true; region = "us"; reason = "restored from backup" }
            }
            return @{ changed = $false; region = $script:ActiveRegion; reason = "US endpoint not captured" }
        }

        $content = $content -replace '(base_url\s*=\s*")[^"]+(")', "`${1}$($script:USEndpoint)`${2}"
        $content = $content -replace '(env_key\s*=\s*")[^"]+(")', "`${1}$($script:USEnvKey)`${2}"

        Set-Content -Path $script:CodexConfigPath -Value $content -Encoding UTF8 -NoNewline
        $script:ActiveRegion = "us"
        $script:RegionSwitchedAt = Get-Date
    }

    return @{ changed = $true; region = $Region; reason = "switched" }
}

function Test-RegionCooldownExpired {
    <#
    .SYNOPSIS Check if the Sweden → US cooldown has expired.
    .DESCRIPTION After switching to Sweden, waits RegionCooldownMinutes before auto-switching back.
    #>
    if ($script:ActiveRegion -ne "sweden") { return $false }
    if (-not $script:RegionSwitchedAt) { return $true }
    return ((Get-Date) - $script:RegionSwitchedAt).TotalMinutes -ge $script:RegionCooldownMinutes
}

function Set-RegionOverride {
    <#
    .SYNOPSIS Manually override the active region. Set to $null for auto mode.
    #>
    param([AllowNull()][string]$Region)
    $script:RegionOverride = $Region
    if ($Region) {
        $result = Switch-CodexRegion -Region $Region -Force
        return $result
    }
    return @{ changed = $false; region = $script:ActiveRegion; reason = "override cleared — auto mode" }
}

function Get-RegionStatus {
    <#
    .SYNOPSIS Get a summary of the current region state.
    #>
    $switchedAgo = if ($script:RegionSwitchedAt) {
        [math]::Round(((Get-Date) - $script:RegionSwitchedAt).TotalMinutes, 1)
    }
    else { $null }

    return @{
        active_region    = $script:ActiveRegion
        override         = $script:RegionOverride
        switched_at      = $script:RegionSwitchedAt
        switched_ago_min = $switchedAgo
        cooldown_min     = $script:RegionCooldownMinutes
        cooldown_expired = (Test-RegionCooldownExpired)
        us_endpoint      = $script:USEndpoint
        sweden_available = [bool]($env:AZURE_SWEDEN_API_KEY -and $env:AZURE_SWEDEN_ENDPOINT)
    }
}

# ─── Task Complexity Classification ──────────────────────────────────────────
# Classifies tasks as small/medium/large based on title keywords and scope.
# Used by the orchestrator to route tasks to the best executor.

function Get-TaskComplexity {
    <#
    .SYNOPSIS Classify task complexity from its title and description.
    .DESCRIPTION Returns: "small", "medium", or "large"

    Heuristics:
      large  → multi-module, architecture, refactor, security audit, full feature
      medium → single module feature, integration, test suite, API endpoint
      small  → fix, typo, docs, config, lint, rename, bump, simple test
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Title,
        [string]$Description = ""
    )
    $text = "$Title $Description".ToLower()

    # ── Large patterns ──
    $largePatterns = @(
        'refactor',
        'architect',
        'redesign',
        'migration',
        'multi.?module',
        'full.?feature',
        'security.?audit',
        'e2e.?test',
        'integration.?test',
        'cross.?module',
        'complete.?implement',
        'overhaul',
        'rewrite',
        'major',
        'provider.?daemon',
        'marketplace.*escrow',
        'veid.*encryption',
        'tee.?integration'
    )
    foreach ($p in $largePatterns) {
        if ($text -match $p) { return "large" }
    }

    # ── Small patterns ──
    $smallPatterns = @(
        'fix\s',
        'typo',
        'docs?\s',
        'documentation',
        'readme',
        'changelog',
        'config',
        'lint',
        'format',
        'rename',
        'bump',
        'version',
        'cleanup',
        'comment',
        'annotation',
        'todo',
        'nit',
        'spelling',
        'label',
        'badge',
        'copyright',
        'license\s',
        'gitignore',
        'simple\s',
        'minor\s',
        'trivial',
        'add\s+test\sfor',
        'update\s+dep'
    )
    foreach ($p in $smallPatterns) {
        if ($text -match $p) { return "small" }
    }

    # Default: medium
    return "medium"
}

function Get-BestExecutorForTask {
    <#
    .SYNOPSIS Select the best executor for a task based on complexity and health.
    .DESCRIPTION
        Routes tasks to executors based on:
        1. Task complexity → suitable models
        2. Executor health → skip degraded/cooldown
        3. Load balancing → prefer less loaded

    Task routing:
        large  → Claude 4.6 (supreme) > GPT-5.2-codex (primary) > GPT-5.1-codex-max
        medium → GPT-5.2-codex (primary) > GPT-5.1-codex-max > Claude 4.6
        small  → GPT-5.1-codex-mini > GPT-5.2-codex > GPT-5.1-codex-max
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Complexity  # small | medium | large
    )
    Initialize-ExecutorHealth

    $tierOrder = @{ "primary" = 0; "backup" = 1; "fallback" = 2 }
    $healthOrder = @{ "healthy" = 0; "degraded" = 1; "cooldown" = 2; "disabled" = 3 }

    # Model preference by complexity (ordered preference)
    $preferredModels = switch ($Complexity) {
        "large" { @("claude-opus-4.6", "gpt-5.2-codex", "gpt-5.1-codex-max", "claude-code") }
        "medium" { @("gpt-5.2-codex", "gpt-5.1-codex-max", "claude-opus-4.6", "gpt-5.1-codex-mini") }
        "small" { @("gpt-5.1-codex-mini", "gpt-5.2-codex", "gpt-5.1-codex-max") }
        default { @("gpt-5.2-codex", "claude-opus-4.6", "gpt-5.1-codex-max", "gpt-5.1-codex-mini") }
    }

    $candidates = @()
    foreach ($exec in $script:VK_EXECUTORS) {
        $provider = $exec.provider
        $health = Get-ExecutorHealthStatus -Provider $provider
        if ($health -in @("disabled", "cooldown")) { continue }

        # Check suitability
        if ($exec.suitability -and $Complexity -notin $exec.suitability) { continue }

        $h = $script:ExecutorHealth[$provider]
        $active = if ($h) { $h.active_tasks } else { 0 }
        if ($active -ge $exec.max_concurrent) { continue }

        # Preference score: lower index in preferredModels = better fit
        $modelPref = $preferredModels.IndexOf($exec.model)
        if ($modelPref -lt 0) { $modelPref = 99 }  # Model not in preference list

        $healthScore = $healthOrder[$health] * 5
        $loadScore = $active

        $candidates += @{
            executor = $exec
            score    = ($modelPref * 10) + $healthScore + $loadScore
            health   = $health
            active   = $active
            model    = $exec.model
        }
    }

    if ($candidates.Count -eq 0) {
        # Fallback: return whatever Get-HealthyExecutorProfile gives
        return Get-HealthyExecutorProfile
    }

    $best = ($candidates | Sort-Object { $_.score } | Select-Object -First 1).executor
    return $best
}

# ─── HTTP Helpers ─────────────────────────────────────────────────────────────

function Invoke-VKApi {
    <#
    .SYNOPSIS Invoke the vibe-kanban REST API.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Path,
        [string]$Method = "GET",
        [object]$Body,
        [int]$TimeoutSec
    )
    $uri = "$script:VK_BASE_URL$Path"
    $params = @{ Uri = $uri; Method = $Method; ContentType = "application/json"; UseBasicParsing = $true }
    if ($TimeoutSec -gt 0) { $params.TimeoutSec = $TimeoutSec }
    if ($Body) { $params.Body = ($Body | ConvertTo-Json -Depth 10 -Compress) }
    try {
        $raw = Invoke-WebRequest @params -ErrorAction Stop
        $resp = $raw.Content | ConvertFrom-Json -Depth 20
        if ($resp.success -eq $false) {
            Write-Warning "API error on $Path : $($resp.message)"
            return $null
        }
        return $resp.data ?? $resp
    }
    catch {
        Write-Error -Message "HTTP $Method $uri failed: $_" -ErrorAction Continue
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
    }
    else {
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
        }
        else {
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
        [ValidateSet("todo", "inprogress", "inreview", "done", "cancelled")]
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
        [ValidateSet("todo", "inprogress", "inreview", "done", "cancelled")]
        [string]$Status
    )
    return Invoke-VKApi -Path "/api/tasks/$TaskId" -Method "PUT" -Body @{ status = $Status }
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

function Get-VKAttemptSummaries {
    <#
    .SYNOPSIS Get attempt summaries for workspace status (idle/running/failed).
    #>
    [CmdletBinding()]
    param([bool]$Archived = $false)
    $body = @{ archived = $Archived }
    $result = Invoke-VKApi -Path "/api/task-attempts/summary" -Method "POST" -Body $body
    if (-not $result) { return @() }
    $summaries = if ($result.summaries) { $result.summaries } else { $result }
    return @($summaries)
}

function Get-VKArchivedAttempts {
    <#
    .SYNOPSIS List archived task attempts.
    #>
    [CmdletBinding()]
    $attempts = Get-VKAttempts
    if (-not $attempts) { return @() }
    return @($attempts | Where-Object { $_.archived })
}

function Ensure-FreshOriginMain {
    <#
    .SYNOPSIS Fetch latest origin/main so new branches are always up-to-date.
              Prevents stale workspace creation that leads to merge conflicts.
    #>
    [CmdletBinding()]
    param([string]$BaseBranch = "main")
    try {
        $fetchOut = git fetch origin $BaseBranch 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  ⚠ git fetch origin $BaseBranch failed: $fetchOut" -ForegroundColor Yellow
            return $false
        }
        # Get the latest commit SHA for reference
        $latestSha = git rev-parse "origin/$BaseBranch" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✓ origin/$BaseBranch up-to-date at $($latestSha.Substring(0,8))" -ForegroundColor DarkGray
        }
        return $true
    }
    catch {
        Write-Host "  ⚠ Ensure-FreshOriginMain failed: $($_.Exception.Message)" -ForegroundColor Yellow
        return $false
    }
}

function Submit-VKTaskAttempt {
    <#
    .SYNOPSIS Submit a task as a new attempt (creates worktree + starts agent).
              Uses the next executor in the Codex/Copilot rotation cycle.
              Always fetches latest origin/main first to prevent stale workspaces.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$TaskId,
        [string]$TargetBranch = $script:VK_TARGET_BRANCH,
        [hashtable]$ExecutorOverride
    )
    if (-not (Initialize-VKConfig)) { return $null }

    # CRITICAL: Fetch latest origin/main BEFORE creating workspace
    # This prevents workspaces from being created 100s of commits behind
    $baseBranchClean = $TargetBranch
    if ($baseBranchClean -like "origin/*") { $baseBranchClean = $baseBranchClean.Substring(7) }
    Ensure-FreshOriginMain -BaseBranch $baseBranchClean | Out-Null

    # Use override if provided, otherwise cycle to next executor
    $execProfile = if ($ExecutorOverride) { $ExecutorOverride } else { Get-NextExecutorProfile }

    $body = @{
        task_id             = $TaskId
        repos               = @(
            @{
                repo_id       = $script:VK_REPO_ID
                target_branch = $TargetBranch
                base_branch   = $baseBranchClean
            }
        )
        executor_profile_id = @{
            executor = $execProfile.executor
            variant  = $execProfile.variant
        }
    }
    Write-Host "  Submitting attempt for task $TaskId ($($execProfile.executor)/$($execProfile.variant)) ..." -ForegroundColor Cyan
    $result = Invoke-VKApi -Path "/api/task-attempts" -Method "POST" -Body $body -TimeoutSec 90
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

function Archive-VKAttempt {
    <#
    .SYNOPSIS Archive a task attempt so it no longer counts as active.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$AttemptId)
    $body = @{ archived = $true }
    return Invoke-VKApi -Path "/api/task-attempts/$AttemptId" -Method "PUT" -Body $body
}

function Unarchive-VKAttempt {
    <#
    .SYNOPSIS Unarchive a task attempt so it counts as active again.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$AttemptId)
    $body = @{ archived = $false }
    return Invoke-VKApi -Path "/api/task-attempts/$AttemptId" -Method "PUT" -Body $body
}

# ─── GitHub PR Functions ─────────────────────────────────────────────────────

function Set-VKLastGithubError {
    param(
        [string]$Type,
        [string]$Message
    )
    $script:VK_LAST_GH_ERROR = @{ type = $Type; message = $Message }
    $script:VK_LAST_GH_ERROR_AT = Get-Date
}

function Clear-VKLastGithubError {
    $script:VK_LAST_GH_ERROR = $null
    $script:VK_LAST_GH_ERROR_AT = $null
}

function Get-VKLastGithubError {
    return $script:VK_LAST_GH_ERROR
}

function Invoke-VKGithub {
    <#
    .SYNOPSIS Invoke gh CLI with rate-limit detection.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string[]]$Args
    )
    $output = & gh @Args 2>&1
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        if ($output -match "rate limit|API rate limit exceeded|secondary rate limit|abuse detection") {
            Set-VKLastGithubError -Type "rate_limit" -Message $output
        }
        else {
            Set-VKLastGithubError -Type "error" -Message $output
        }
        return $null
    }
    Clear-VKLastGithubError
    return $output
}

function Get-PRForBranch {
    <#
    .SYNOPSIS Find an open or merged PR for a given branch.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Branch)
    $prJson = Invoke-VKGithub -Args @(
        "pr", "list", "--head", $Branch, "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--json", "number,state,title,mergeable,statusCheckRollup,mergedAt,createdAt,url",
        "--limit", "1"
    )
    if (-not $prJson -or $prJson -eq "[]") {
        # Also check merged/closed
        $prJson = Invoke-VKGithub -Args @(
            "pr", "list", "--head", $Branch, "--repo", "$script:GH_OWNER/$script:GH_REPO",
            "--state", "merged", "--json", "number,state,title,mergedAt,createdAt,url",
            "--limit", "1"
        )
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
    $checksJson = Invoke-VKGithub -Args @(
        "pr", "checks", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--json", "name,state"
    )
    if (-not $checksJson) { return "unknown" }
    $checks = $checksJson | ConvertFrom-Json
    if (-not $checks -or $checks.Count -eq 0) { return "unknown" }

    $failing = $checks | Where-Object { $_.state -in @("FAILURE", "ERROR") }
    $pending = $checks | Where-Object { $_.state -in @("PENDING", "IN_PROGRESS", "QUEUED") }

    if ($failing.Count -gt 0) { return "failing" }
    if ($pending.Count -gt 0) { return "pending" }
    return "passing"
}

function Get-PRLatestCheckTimestamp {
    <#
    .SYNOPSIS Get the most recent check started/completed time for a PR.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $checksJson = Invoke-VKGithub -Args @(
        "pr", "checks", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--json", "name,state,startedAt,completedAt"
    )
    if (-not $checksJson) { return $null }
    $checks = $checksJson | ConvertFrom-Json
    if (-not $checks -or $checks.Count -eq 0) { return $null }

    $timestamps = @()
    foreach ($check in $checks) {
        if ($check.startedAt) { $timestamps += [datetime]$check.startedAt }
        if ($check.completedAt) { $timestamps += [datetime]$check.completedAt }
    }
    if ($timestamps.Count -eq 0) { return $null }
    return ($timestamps | Sort-Object -Descending | Select-Object -First 1)
}

function Get-OpenPullRequests {
    <#
    .SYNOPSIS List open PRs with metadata.
    #>
    [CmdletBinding()]
    param([int]$Limit = 100)
    $prJson = Invoke-VKGithub -Args @(
        "pr", "list", "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--state", "open", "--limit", $Limit.ToString(),
        "--json", "number,title,author,isDraft,createdAt,headRefName,baseRefName,mergeStateStatus,url,body"
    )
    if (-not $prJson -or $prJson -eq "[]") { return @() }
    return $prJson | ConvertFrom-Json
}

function Test-IsCopilotAuthor {
    <#
    .SYNOPSIS Determine whether a PR author is Copilot.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][object]$Author)
    $login = $Author.login
    if (-not $login) { return $false }
    return ($login -match "copilot")
}

function Get-PRDetails {
    <#
    .SYNOPSIS Get mergeability details for a PR.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $prJson = Invoke-VKGithub -Args @(
        "pr", "view", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--json", "number,state,mergeable,mergeStateStatus,isDraft,reviewDecision,url,headRefName,baseRefName,title,body"
    )
    if (-not $prJson) { return $null }
    return $prJson | ConvertFrom-Json
}

function Get-PRChecksDetail {
    <#
    .SYNOPSIS Get detailed check info for a PR.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $checksJson = Invoke-VKGithub -Args @(
        "pr", "checks", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--json", "name,state,link,startedAt,completedAt"
    )
    if (-not $checksJson) { return @() }
    return $checksJson | ConvertFrom-Json
}

function Get-RequiredChecksForBranch {
    <#
    .SYNOPSIS Get required status check names for a branch.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Branch)
    $encodedBranch = [System.Uri]::EscapeDataString($Branch)
    $result = Invoke-VKGithub -Args @(
        "api",
        "repos/$script:GH_OWNER/$script:GH_REPO/branches/$encodedBranch/protection/required_status_checks"
    )
    if (-not $result) { return @() }
    $payload = $result | ConvertFrom-Json
    if (-not $payload) { return @() }

    $names = @()
    if ($payload.contexts) {
        $names += @($payload.contexts | Where-Object { $_ })
    }
    if ($payload.checks) {
        $names += @($payload.checks | ForEach-Object {
                if ($_.context) { $_.context } else { $_.name }
            } | Where-Object { $_ })
    }

    return @($names | Sort-Object -Unique)
}

function Get-PRRequiredCheckStatus {
    <#
    .SYNOPSIS Evaluate only required checks for a PR.
    .OUTPUTS "passing", "failing", "pending", or "unknown"
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [Parameter(Mandatory)][string]$BaseBranch
    )
    $required = Get-RequiredChecksForBranch -Branch $BaseBranch
    if (-not $required -or @($required).Count -eq 0) {
        return "passing"
    }

    $checks = Get-PRChecksDetail -PRNumber $PRNumber
    if (-not $checks) { return "unknown" }

    $requiredLower = $required | ForEach-Object { $_.ToLowerInvariant() }
    $checksByName = @{}
    foreach ($check in $checks) {
        if (-not $check.name) { continue }
        $checksByName[$check.name.ToLowerInvariant()] = $check.state
    }

    $hasPending = $false
    foreach ($name in $requiredLower) {
        if (-not $checksByName.ContainsKey($name)) {
            $hasPending = $true
            continue
        }
        $state = $checksByName[$name]
        if ($state -in @("FAILURE", "ERROR")) { return "failing" }
        if ($state -in @("PENDING", "IN_PROGRESS", "QUEUED")) { $hasPending = $true }
    }

    if ($hasPending) { return "pending" }
    return "passing"
}

function Get-PRSecurityCheckStatus {
    <#
    .SYNOPSIS Evaluate security-related checks for a PR.
    .OUTPUTS "passing", "failing", "pending", or "missing"
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)

    $checks = Get-PRChecksDetail -PRNumber $PRNumber
    if (-not $checks) { return "missing" }

    $patterns = @(
        "codeql",
        "security",
        "snyk",
        "semgrep",
        "trivy",
        "gitleaks",
        "secret",
        "scorecard",
        "dependabot",
        "ossf"
    )

    $securityChecks = @(
        $checks | Where-Object {
            $name = if ($_.name) { $_.name.ToString().ToLowerInvariant() } else { "" }
            foreach ($pattern in $patterns) {
                if ($name -match $pattern) { return $true }
            }
            return $false
        }
    )

    if (-not $securityChecks -or $securityChecks.Count -eq 0) {
        return "missing"
    }

    $hasPending = $false
    foreach ($check in $securityChecks) {
        $state = $check.state
        if ($state -in @("FAILURE", "ERROR", "CANCELLED")) { return "failing" }
        if ($state -in @("PENDING", "IN_PROGRESS", "QUEUED")) { $hasPending = $true }
        if ($state -in @("SKIPPED")) { return "failing" }
    }

    if ($hasPending) { return "pending" }
    return "passing"
}

function Create-GithubIssue {
    <#
    .SYNOPSIS Create a GitHub issue.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Title,
        [Parameter(Mandatory)][string]$Body
    )
    $result = Invoke-VKGithub -Args @(
        "api",
        "repos/$script:GH_OWNER/$script:GH_REPO/issues",
        "-f", "title=$Title",
        "-f", "body=$Body"
    )
    if (-not $result) { return $null }
    return $result | ConvertFrom-Json
}

function Assign-IssueToCopilot {
    <#
    .SYNOPSIS Assign a GitHub issue to Copilot.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$IssueNumber)
    $result = Invoke-VKGithub -Args @(
        "api",
        "repos/$script:GH_OWNER/$script:GH_REPO/issues/$IssueNumber/assignees",
        "-f", "assignees[]=copilot"
    )
    return ($null -ne $result)
}

function Get-RecentMergedPRs {
    <#
    .SYNOPSIS List recent merged PRs.
    #>
    [CmdletBinding()]
    param([int]$Limit = 15)
    $prJson = Invoke-VKGithub -Args @(
        "pr", "list", "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--state", "merged", "--limit", $Limit.ToString(),
        "--json", "number,title,url,mergedAt"
    )
    if (-not $prJson -or $prJson -eq "[]") { return @() }
    return $prJson | ConvertFrom-Json
}

function Get-FailingWorkflowRuns {
    <#
    .SYNOPSIS Get recent failing workflow runs on main.
    #>
    [CmdletBinding()]
    param([int]$Limit = 8)
    $result = Invoke-VKGithub -Args @(
        "api",
        "repos/$script:GH_OWNER/$script:GH_REPO/actions/runs",
        "-f", "branch=main",
        "-f", "status=completed",
        "-f", "per_page=$Limit"
    )
    if (-not $result) { return @() }
    $payload = $result | ConvertFrom-Json
    $runs = $payload.workflow_runs
    if (-not $runs) { return @() }
    return $runs | Where-Object {
        $_.conclusion -in @("failure", "cancelled", "timed_out", "action_required")
    }
}

function Format-PRCheckFailures {
    <#
    .SYNOPSIS Format failing checks into a short markdown list.
    #>
    [CmdletBinding()]
    param([AllowNull()][object[]]$Checks = @())
    if ($null -eq $Checks) { $Checks = @() }
    if (-not $Checks -or $Checks.Count -eq 0) { return "- No failing checks found" }
    $failed = $Checks | Where-Object { $_.state -eq "FAILURE" -or $_.state -eq "ERROR" }
    if (-not $failed -or $failed.Count -eq 0) { return "- No failing checks found" }
    $lines = $failed | ForEach-Object {
        $link = if ($_.link) { " ($($_.link))" } else { "" }
        "- $($_.name): $($_.state)$link"
    }
    return ($lines -join "`n")
}

function Add-PRComment {
    <#
    .SYNOPSIS Add a comment to a PR.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [Parameter(Mandatory)][string]$Body
    )
    $result = Invoke-VKGithub -Args @(
        "pr", "comment", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--body", $Body
    )
    return ($null -ne $result)
}

function Mark-PRReady {
    <#
    .SYNOPSIS Mark a PR as ready for review (exit draft).
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $result = Invoke-VKGithub -Args @(
        "pr", "ready", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO"
    )
    return ($null -ne $result)
}

function Find-CopilotFixPR {
    <#
    .SYNOPSIS Find a Copilot fix PR referencing an original PR number.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$OriginalPRNumber)
    $search = "$OriginalPRNumber in:title,body"
    $prJson = Invoke-VKGithub -Args @(
        "pr", "list", "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--state", "open", "--search", $search,
        "--json", "number,title,body,headRefName,baseRefName,isDraft,url,createdAt",
        "--limit", "20"
    )
    if (-not $prJson) { return $null }
    $prs = $prJson | ConvertFrom-Json
    if (-not $prs) { return $null }
    $pattern = "(?i)(PR\s*#?$OriginalPRNumber|#$OriginalPRNumber)"
    $match = $prs | Where-Object {
        ($_.title -match $pattern) -or ($_.body -match $pattern)
    } | Sort-Object -Property createdAt -Descending | Select-Object -First 1
    return $match
}

function Test-CopilotPRComplete {
    <#
    .SYNOPSIS Determine whether a Copilot PR is complete (not WIP and not draft).
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][object]$PRDetails)
    if (-not $PRDetails) { return $false }
    if ($PRDetails.isDraft) { return $false }
    if ($PRDetails.title -match "^\[WIP\]") { return $false }
    return $true
}

function Close-PRDeleteBranch {
    <#
    .SYNOPSIS Close a PR and delete its branch.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $result = Invoke-VKGithub -Args @(
        "pr", "close", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--delete-branch"
    )
    return ($null -ne $result)
}

function Merge-BranchFromPR {
    <#
    .SYNOPSIS Merge a PR head branch into a target base branch.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$BaseBranch,
        [Parameter(Mandatory)][string]$HeadBranch
    )
    $result = Invoke-VKGithub -Args @(
        "api", "repos/$script:GH_OWNER/$script:GH_REPO/merges",
        "-f", "base=$BaseBranch",
        "-f", "head=$HeadBranch"
    )
    return ($null -ne $result)
}

function Create-PRForBranch {
    <#
    .SYNOPSIS Create a PR for a branch when the agent did not open one.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Branch,
        [Parameter(Mandatory)][string]$Title,
        [string]$Body = "Automated PR created by ve-orchestrator"
    )
    $baseBranch = $script:VK_TARGET_BRANCH
    if ($baseBranch -like "origin/*") { $baseBranch = $baseBranch.Substring(7) }
    $result = Invoke-VKGithub -Args @(
        "pr", "create", "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--head", $Branch,
        "--base", $baseBranch,
        "--title", $Title,
        "--body", $Body
    )
    return ($null -ne $result)
}

function Test-RemoteBranchExists {
    <#
    .SYNOPSIS Check if a branch exists on the remote GitHub repo.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Branch)
    $encodedBranch = [System.Uri]::EscapeDataString($Branch)
    $result = Invoke-VKGithub -Args @(
        "api",
        "repos/$script:GH_OWNER/$script:GH_REPO/branches/$encodedBranch"
    )
    return ($null -ne $result)
}

function Get-VKSessions {
    <#
    .SYNOPSIS List sessions for a workspace.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$WorkspaceId)
    $result = Invoke-VKApi -Path "/api/sessions?workspace_id=$WorkspaceId"
    if (-not $result) { return @() }
    return @($result)
}

function New-VKSession {
    <#
    .SYNOPSIS Create a new session for a workspace.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$WorkspaceId)
    $body = @{ workspace_id = $WorkspaceId }
    return Invoke-VKApi -Path "/api/sessions" -Method "POST" -Body $body
}

function Get-ExecutorProfileForSession {
    <#
    .SYNOPSIS Map a session executor to an executor profile.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Executor)
    switch ($Executor) {
        "COPILOT" { return @{ executor = "COPILOT"; variant = "CLAUDE_OPUS_4_6" } }
        "CODEX" { return @{ executor = "CODEX"; variant = "DEFAULT" } }
        default { return @{ executor = "CODEX"; variant = "DEFAULT" } }
    }
}

function Send-VKWorkspaceFollowUp {
    <#
    .SYNOPSIS Send a direct follow-up message to the latest session for a workspace.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$WorkspaceId,
        [Parameter(Mandatory)][string]$Message,
        [hashtable]$ExecutorOverride
    )
    $sessions = Get-VKSessions -WorkspaceId $WorkspaceId
    if (-not $sessions -or @($sessions).Count -eq 0) { return $false }
    $session = $sessions | Sort-Object -Property updated_at -Descending | Select-Object -First 1
    $profile = if ($ExecutorOverride) { $ExecutorOverride } else { Get-ExecutorProfileForSession -Executor $session.executor }
    return Send-VKSessionFollowUp -SessionId $session.id -Message $Message -ExecutorProfile $profile
}

function Send-VKSessionFollowUp {
    <#
    .SYNOPSIS Send a follow-up message to a specific session.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$SessionId,
        [Parameter(Mandatory)][string]$Message,
        [Parameter(Mandatory)][hashtable]$ExecutorProfile
    )
    $body = @{ prompt = $Message; executor_profile_id = $ExecutorProfile }
    $result = Invoke-VKApi -Path "/api/sessions/$SessionId/follow-up" -Method "POST" -Body $body
    return ($null -ne $result)
}

function Queue-VKWorkspaceMessage {
    <#
    .SYNOPSIS Enqueue a message to the latest session for a workspace.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$WorkspaceId,
        [Parameter(Mandatory)][string]$Message
    )
    $sessions = Get-VKSessions -WorkspaceId $WorkspaceId
    if (-not $sessions -or @($sessions).Count -eq 0) { return $false }
    $session = $sessions | Sort-Object -Property updated_at -Descending | Select-Object -First 1
    $profile = Get-ExecutorProfileForSession -Executor $session.executor
    $body = @{ executor_profile_id = $profile; message = $Message }
    $result = Invoke-VKApi -Path "/api/sessions/$($session.id)/queue" -Method "POST" -Body $body
    return ($null -ne $result)
}

function Merge-PR {
    <#
    .SYNOPSIS Merge a PR after rebase onto latest main. Returns $true on success.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$PRNumber,
        [switch]$AutoMerge,
        [switch]$Admin
    )
    $args = @(
        "pr", "merge", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--merge", "--delete-branch"
    )
    if ($AutoMerge) { $args += "--auto" }
    if ($Admin) { $args += "--admin" }
    if ($AutoMerge) {
        Write-Host "  Enabling auto-merge for PR #$PRNumber ..." -ForegroundColor Yellow
    }
    elseif ($Admin) {
        Write-Host "  Admin merging PR #$PRNumber ..." -ForegroundColor Yellow
    }
    else {
        Write-Host "  Merging PR #$PRNumber ..." -ForegroundColor Yellow
    }
    $result = Invoke-VKGithub -Args $args
    return ($null -ne $result)
}

function Enable-AutoMerge {
    <#
    .SYNOPSIS Enable auto-merge on a PR so it merges when CI passes.
    #>
    [CmdletBinding()]
    param([Parameter(Mandatory)][int]$PRNumber)
    $result = Invoke-VKGithub -Args @(
        "pr", "merge", $PRNumber.ToString(), "--repo", "$script:GH_OWNER/$script:GH_REPO",
        "--auto", "--merge", "--delete-branch"
    )
    return ($null -ne $result)
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
        [ValidateSet("todo", "inprogress", "inreview", "done", "cancelled")]
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
    [CmdletBinding()]
    param(
        [switch]$ShowIdle
    )
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
            if ($ShowIdle) {
                $lastUpdate = if ($a.updated_at) { [datetime]$a.updated_at } elseif ($a.created_at) { [datetime]$a.created_at } else { $null }
                $idleMinutes = if ($lastUpdate) { [math]::Round(((Get-Date) - $lastUpdate).TotalMinutes, 1) } else { "?" }
                $updatedAtText = if ($lastUpdate) { $lastUpdate.ToString("u") } else { "unknown" }
                Write-Host "  │   Idle: ${idleMinutes} min (last update: $updatedAtText)" -ForegroundColor DarkGray
            }
        }
        Write-Host "  └────────────────────────────────────────────" -ForegroundColor Green
    }
    Write-Host ""
}

# ─── CLI Dispatch ─────────────────────────────────────────────────────────────

function Invoke-CLI {
    param([string[]]$Arguments)

    if ($null -eq $Arguments) {
        $Arguments = @()
    }
    else {
        $Arguments = @($Arguments)
    }

    if ($Arguments.Count -eq 0) {
        Show-Usage
        return
    }

    $command = $Arguments[0]
    $rest = if ($Arguments.Count -gt 1) { $Arguments[1..($Arguments.Count - 1)] } else { @() }

    switch ($command) {
        "list" {
            $status = "todo"
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--status", "-s") -and ($i + 1) -lt $rest.Count) { $status = $rest[$i + 1] }
            }
            Show-Tasks -Status $status
        }
        "status" {
            $showIdle = $false
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--verbose", "-v")) { $showIdle = $true }
            }
            if ($VerbosePreference -eq "Continue") { $showIdle = $true }
            if (Test-CLIFlagPresent -Flags @("--verbose", "-v") -Arguments $rest) { $showIdle = $true }
            Show-Status -ShowIdle:$showIdle
        }
        "archived" {
            $archived = Get-VKArchivedAttempts
            if (-not $archived -or @($archived).Count -eq 0) {
                Write-Host "  No archived attempts." -ForegroundColor Yellow
                return
            }
            Write-Host "";
            Write-Host "  ┌─ Archived Attempts ─────────────────────────" -ForegroundColor DarkGray
            foreach ($a in $archived) {
                $name = if ($a.name) { $a.name.Substring(0, [Math]::Min(60, $a.name.Length)) } else { "(unnamed)" }
                Write-Host "  │ $($a.id.Substring(0,8))  $($a.branch)" -ForegroundColor White
                Write-Host "  │   $name" -ForegroundColor DarkGray
            }
            Write-Host "  └────────────────────────────────────────────" -ForegroundColor DarkGray
            Write-Host ""
        }
        "unarchive" {
            if ($rest.Count -eq 0) { Write-Error "Usage: ve-kanban unarchive <attempt-id>"; return }
            Unarchive-VKAttempt -AttemptId $rest[0] | Out-Null
            Write-Host "  ✓ Attempt $($rest[0]) unarchived" -ForegroundColor Green
        }
        "submit" {
            if ($rest.Count -eq 0) { Write-Error "Usage: ve-kanban submit <task-id>"; return }
            $taskId = $rest[0]
            Submit-VKTaskAttempt -TaskId $taskId
        }
        "submit-next" {
            $count = 1
            for ($i = 0; $i -lt $rest.Count; $i++) {
                if ($rest[$i] -in @("--count", "-n") -and ($i + 1) -lt $rest.Count) { $count = [int]$rest[$i + 1] }
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
                if ($rest[$i] -in @("--branch", "-b") -and ($i + 1) -lt $rest.Count) { $branch = $rest[$i + 1] }
                if ($rest[$i] -eq "--auto") { $auto = $true }
            }
            if (-not $branch -and $rest.Count -gt 0 -and $rest[0] -notlike "--*") { $branch = $rest[0] }
            if (-not $branch) { Write-Error "Usage: ve-kanban merge <branch> [--auto]"; return }
            $pr = Get-PRForBranch -Branch $branch
            if (-not $pr) { Write-Error "No PR found for branch $branch"; return }
            if ($auto) {
                Enable-AutoMerge -PRNumber $pr.number
            }
            else {
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
                if ($rest[$i] -in @("--parallel", "-p") -and ($i + 1) -lt $rest.Count) { $parallel = [int]$rest[$i + 1] }
                if ($rest[$i] -in @("--interval", "-i") -and ($i + 1) -lt $rest.Count) { $interval = [int]$rest[$i + 1] }
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
    status --verbose | -v            Show dashboard with idle minutes
    archived                        List archived attempts
    unarchive <attempt-id>          Unarchive an attempt
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

function Test-CLIFlagPresent {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string[]]$Flags,
        [string[]]$Arguments
    )
    if ($Arguments) {
        foreach ($flag in $Flags) {
            if ($Arguments -contains $flag) { return $true }
        }
    }
    if ($script:VK_CLI_RAW_LINE) {
        foreach ($flag in $Flags) {
            $pattern = "(^|\s)" + [regex]::Escape($flag) + "(\s|$)"
            if ($script:VK_CLI_RAW_LINE -match $pattern) { return $true }
        }
    }
    return $false
}

# Run CLI if invoked directly (not dot-sourced)
if ($MyInvocation.InvocationName -ne ".") {
    $script:VK_CLI_RAW_LINE = $MyInvocation.Line
    Invoke-CLI -Arguments $args
}
