$lines = Get-Content (Join-Path $PSScriptRoot 'codex-monitor\ve-orchestrator.ps1')
Write-Host "Total lines: $($lines.Count)"
Write-Host "L3417: $($lines[3416])"
Write-Host "L3418: $($lines[3417])"
Write-Host "L3419: $($lines[3418])"
