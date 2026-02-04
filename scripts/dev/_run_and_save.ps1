$ErrorActionPreference = "Continue"
$outputFile = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine\_command_results.txt"

Set-Location "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

"=== COMMAND 1: go build ./x/bme/... ./x/oracle/... ===" | Out-File -FilePath $outputFile
"Started at: $(Get-Date)" | Out-File -FilePath $outputFile -Append

$buildOutput = go build ./x/bme/... ./x/oracle/... 2>&1 | Out-String
$buildExitCode = $LASTEXITCODE

$buildOutput | Out-File -FilePath $outputFile -Append
"BUILD_EXIT_CODE: $buildExitCode" | Out-File -FilePath $outputFile -Append
"Completed at: $(Get-Date)" | Out-File -FilePath $outputFile -Append

"" | Out-File -FilePath $outputFile -Append
"=== COMMAND 2: go test ./x/bme/keeper/... ./x/oracle/keeper/... ===" | Out-File -FilePath $outputFile -Append
"Started at: $(Get-Date)" | Out-File -FilePath $outputFile -Append

$testOutput = go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1 | Out-String
$testExitCode = $LASTEXITCODE

$testOutput | Out-File -FilePath $outputFile -Append
"TEST_EXIT_CODE: $testExitCode" | Out-File -FilePath $outputFile -Append
"Completed at: $(Get-Date)" | Out-File -FilePath $outputFile -Append

"=== DONE ===" | Out-File -FilePath $outputFile -Append
