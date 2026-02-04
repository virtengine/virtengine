$workDir = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"
Set-Location $workDir

# Build command
$buildProc = Start-Process -FilePath "go" -ArgumentList "build","./x/bme/...","./x/oracle/..." -WorkingDirectory $workDir -Wait -PassThru -NoNewWindow -RedirectStandardOutput "$workDir\build_stdout.tmp" -RedirectStandardError "$workDir\build_stderr.tmp"
$buildExit = $buildProc.ExitCode
$buildStdout = Get-Content "$workDir\build_stdout.tmp" -Raw -ErrorAction SilentlyContinue
$buildStderr = Get-Content "$workDir\build_stderr.tmp" -Raw -ErrorAction SilentlyContinue
$buildOut = "$buildStdout`n$buildStderr".Trim()
if ([string]::IsNullOrWhiteSpace($buildOut)) { $buildOut = "Build passed with no output" }
$buildStatus = if ($buildExit -eq 0) { "PASSED" } else { "FAILED" }

# Test command
$testProc = Start-Process -FilePath "go" -ArgumentList "test","./x/bme/keeper/...","./x/oracle/keeper/..." -WorkingDirectory $workDir -Wait -PassThru -NoNewWindow -RedirectStandardOutput "$workDir\test_stdout.tmp" -RedirectStandardError "$workDir\test_stderr.tmp"
$testExit = $testProc.ExitCode
$testStdout = Get-Content "$workDir\test_stdout.tmp" -Raw -ErrorAction SilentlyContinue
$testStderr = Get-Content "$workDir\test_stderr.tmp" -Raw -ErrorAction SilentlyContinue
$testOut = "$testStdout`n$testStderr".Trim()
$testStatus = if ($testExit -eq 0) { "PASSED" } else { "FAILED" }

# Create results file
$content = @"
=== BUILD RESULTS ===
$buildOut
BUILD EXIT CODE: $buildExit
BUILD STATUS: $buildStatus

=== TEST RESULTS ===
$testOut
TEST EXIT CODE: $testExit
TEST STATUS: $testStatus
"@

$content | Out-File -FilePath "$workDir\results.txt" -Encoding utf8

# Cleanup temp files
Remove-Item "$workDir\build_stdout.tmp","$workDir\build_stderr.tmp","$workDir\test_stdout.tmp","$workDir\test_stderr.tmp" -ErrorAction SilentlyContinue

Write-Host "Done: Build=$buildStatus Tests=$testStatus"
