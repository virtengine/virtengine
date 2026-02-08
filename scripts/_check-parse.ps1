$tokens = $null
$errors = $null
$path = Join-Path $PSScriptRoot 'codex-monitor\ve-orchestrator.ps1'
[void][System.Management.Automation.Language.Parser]::ParseFile($path, [ref]$tokens, [ref]$errors)
Write-Host "Parse errors: $($errors.Count)"
foreach ($err in $errors) {
    Write-Host "  Line $($err.Extent.StartLineNumber): $($err.Message)"
}
if ($errors.Count -eq 0) { exit 0 } else { exit 1 }
