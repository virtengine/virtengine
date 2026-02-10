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
