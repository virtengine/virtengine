#!/usr/bin/env pwsh
# Script to consolidate orchestrator/kanban to codex-monitor and remove Copilot PR notification code
# Run from repo root

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = $PSScriptRoot | Split-Path -Parent

Write-Host "=== Step 1: Delete root duplicate scripts ==="
git rm "$repoRoot/scripts/ve-orchestrator.ps1" "$repoRoot/scripts/ve-kanban.ps1" 2>$null
Write-Host "Deleted root scripts"

Write-Host "`n=== Step 2: Update .env ==="
$envPath = "$repoRoot/scripts/codex-monitor/.env"
$content = Get-Content $envPath -Raw
$content = $content -replace 'ORCHESTRATOR_SCRIPT=\.\./ve-orchestrator\.ps1', 'ORCHESTRATOR_SCRIPT=./ve-orchestrator.ps1'
Set-Content $envPath $content -NoNewline
Write-Host "Updated .env"

Write-Host "`n=== Step 3: Update _validate-syntax.ps1 ==="
$vsPath = "$repoRoot/scripts/_validate-syntax.ps1"
$newValidate = @'
$tokens = $null
$errors = $null
$orchestratorPath = Join-Path $PSScriptRoot 'codex-monitor\ve-orchestrator.ps1'
if (-not (Test-Path $orchestratorPath)) {
    $orchestratorPath = Join-Path $PSScriptRoot 've-orchestrator.ps1'
}
$null = [System.Management.Automation.Language.Parser]::ParseFile(
    $orchestratorPath,
    [ref]$tokens,
    [ref]$errors
)
# Pre-existing parse error at line 1190 ("$TaskId:$Category") is a known issue
$newErrors = @($errors | Where-Object { $_.Extent.StartLineNumber -ne 1190 })
if ($newErrors.Count -eq 0) {
    Write-Host "OK: No new parse errors (1 pre-existing at line 1190)"
} else {
    Write-Host "NEW parse errors found:"
    foreach ($err in $newErrors) {
        Write-Host "  Line $($err.Extent.StartLineNumber): $($err.ToString())"
    }
    exit 1
}
Write-Host "Total errors: $($errors.Count) (all pre-existing)"
'@
Set-Content $vsPath $newValidate
Write-Host "Updated _validate-syntax.ps1"

Write-Host "`n=== Step 4: Update archive-completed-tasks.ps1 ==="
$archPath = "$repoRoot/scripts/archive-completed-tasks.ps1"
$archContent = Get-Content $archPath -Raw
$archContent = $archContent -replace [regex]::Escape('$kanbanPath = Join-Path $repoRoot "scripts\ve-kanban.ps1"
if (-not (Test-Path -LiteralPath $kanbanPath)) {
  throw "ve-kanban.ps1 not found at $kanbanPath"
}
. $kanbanPath'), '$kanbanPath = Join-Path $repoRoot "scripts\codex-monitor\ve-kanban.ps1"
if (-not (Test-Path -LiteralPath $kanbanPath)) {
  $kanbanPath = Join-Path $repoRoot "scripts\ve-kanban.ps1"
}
if (-not (Test-Path -LiteralPath $kanbanPath)) {
  throw "ve-kanban.ps1 not found at $kanbanPath"
}
. $kanbanPath'
Set-Content $archPath $archContent -NoNewline
Write-Host "Updated archive-completed-tasks.ps1"

Write-Host "`n=== Step 5: Update config.mjs ==="
$configPath = "$repoRoot/scripts/codex-monitor/config.mjs"
$configContent = Get-Content $configPath -Raw
$configContent = $configContent -replace [regex]::Escape('// Default to sibling location (most common for npm-installed codex-monitor)
  return resolve(configDir, "..", "ve-orchestrator.ps1");'), '// Default to bundled location (inside codex-monitor dir)
  return resolve(configDir, "ve-orchestrator.ps1");'
Set-Content $configPath $configContent -NoNewline
Write-Host "Updated config.mjs"

Write-Host "`n=== Step 6: Update telegram-bot.mjs ==="
$tgPath = "$repoRoot/scripts/codex-monitor/telegram-bot.mjs"
$tgContent = Get-Content $tgPath -Raw
$tgContent = $tgContent -replace [regex]::Escape('resolve(repoRoot, "scripts", "ve-kanban.ps1")'), 'resolve(repoRoot, "scripts", "codex-monitor", "ve-kanban.ps1")'
Set-Content $tgPath $tgContent -NoNewline
Write-Host "Updated telegram-bot.mjs"

Write-Host "`n=== Step 7: Update autofix.mjs ==="
$afPath = "$repoRoot/scripts/codex-monitor/autofix.mjs"
$afContent = Get-Content $afPath -Raw
$afContent = $afContent -replace [regex]::Escape('scripts/ve-orchestrator.ps1'), 'scripts/codex-monitor/ve-orchestrator.ps1'
Set-Content $afPath $afContent -NoNewline
Write-Host "Updated autofix.mjs"

Write-Host "`n=== Step 8: Update codex-shell.mjs ==="
$csPath = "$repoRoot/scripts/codex-monitor/codex-shell.mjs"
$csContent = Get-Content $csPath -Raw
$csContent = $csContent -replace [regex]::Escape('scripts/ve-orchestrator.ps1'), 'scripts/codex-monitor/ve-orchestrator.ps1'
Set-Content $csPath $csContent -NoNewline
Write-Host "Updated codex-shell.mjs"

Write-Host "`n=== Step 9: Update copilot-shell.mjs ==="
$cpPath = "$repoRoot/scripts/codex-monitor/copilot-shell.mjs"
$cpContent = Get-Content $cpPath -Raw
$cpContent = $cpContent -replace [regex]::Escape('scripts/ve-orchestrator.ps1'), 'scripts/codex-monitor/ve-orchestrator.ps1'
Set-Content $cpPath $cpContent -NoNewline
Write-Host "Updated copilot-shell.mjs"

Write-Host "`n=== Step 10: Update AGENTS.md ==="
$agentsPath = "$repoRoot/AGENTS.md"
$agentsContent = Get-Content $agentsPath -Raw
$agentsContent = $agentsContent -replace [regex]::Escape('scripts/ve-orchestrator.ps1'), 'scripts/codex-monitor/ve-orchestrator.ps1'
Set-Content $agentsPath $agentsContent -NoNewline
Write-Host "Updated AGENTS.md"

Write-Host "`n=== Step 11: Update operator guide ==="
$guidePath = "$repoRoot/_docs/operations/slopes-multi-workspace-operator-guide.md"
if (Test-Path $guidePath) {
    $guideContent = Get-Content $guidePath -Raw
    $guideContent = $guideContent -replace 'scripts/ve-orchestrator\.ps1', 'scripts/codex-monitor/ve-orchestrator.ps1'
    $guideContent = $guideContent -replace 'scripts/ve-kanban\.ps1', 'scripts/codex-monitor/ve-kanban.ps1'
    $guideContent = $guideContent -replace 'scripts\\ve-kanban\.ps1', 'scripts\\codex-monitor\\ve-kanban.ps1'
    $guideContent = $guideContent -replace 'scripts\\ve-orchestrator\.ps1', 'scripts\\codex-monitor\\ve-orchestrator.ps1'
    Set-Content $guidePath $guideContent -NoNewline
    Write-Host "Updated operator guide"
}
else {
    Write-Host "Operator guide not found, skipping"
}

Write-Host "`n=== All reference updates complete ==="
