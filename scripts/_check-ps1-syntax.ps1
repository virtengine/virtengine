<#
.SYNOPSIS Validate PowerShell syntax for .ps1 files.
.DESCRIPTION Parses PS1 files and reports any syntax errors. Returns exit code 1 if errors found.
.PARAMETER Path One or more PS1 file paths to check. If omitted, checks ve-orchestrator.ps1 and ve-kanban.ps1.
#>
param(
    [string[]]$Path
)

if (-not $Path) {
    $scriptDir = Join-Path $PSScriptRoot "codex-monitor"
    $Path = @(
        (Join-Path $scriptDir "ve-orchestrator.ps1"),
        (Join-Path $scriptDir "ve-kanban.ps1")
    )
}

$totalErrors = 0
foreach ($file in $Path) {
    if (-not (Test-Path $file)) {
        Write-Host "[ps1-syntax] SKIP: $file (not found)" -ForegroundColor Yellow
        continue
    }
    $tokens = $null
    $errors = $null
    [void][System.Management.Automation.Language.Parser]::ParseFile($file, [ref]$tokens, [ref]$errors)
    $basename = Split-Path $file -Leaf
    if ($errors.Count -gt 0) {
        foreach ($err in $errors) {
            $line = $err.Extent.StartLineNumber
            $col  = $err.Extent.StartColumnNumber
            Write-Host "[ps1-syntax] ERROR: ${basename}:${line}:${col} — $($err.Message)" -ForegroundColor Red
        }
        $totalErrors += $errors.Count
    } else {
        Write-Host "[ps1-syntax] OK: $basename" -ForegroundColor Green
    }
}

if ($totalErrors -gt 0) {
    Write-Host "[ps1-syntax] FAILED: $totalErrors error(s) found" -ForegroundColor Red
    exit 1
} else {
    Write-Host "[ps1-syntax] All files OK" -ForegroundColor Green
    exit 0
}
