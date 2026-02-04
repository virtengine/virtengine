$ErrorActionPreference = "Continue"
$resultsFile = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine\_command_results.txt"
Set-Location "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

"=== GO BUILD ./x/bme/... ./x/oracle/... ===" | Out-File $resultsFile -Encoding utf8
"Started at: $(Get-Date)" | Out-File $resultsFile -Append

$buildOutput = & go build ./x/bme/... ./x/oracle/... 2>&1 | Out-String
$buildExit = $LASTEXITCODE

"$buildOutput" | Out-File $resultsFile -Append
"BUILD_EXIT_CODE: $buildExit" | Out-File $resultsFile -Append
"" | Out-File $resultsFile -Append

"=== GO TEST ./x/bme/keeper/... ./x/oracle/keeper/... ===" | Out-File $resultsFile -Append
"Started at: $(Get-Date)" | Out-File $resultsFile -Append

$testOutput = & go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1 | Out-String
$testExit = $LASTEXITCODE

"$testOutput" | Out-File $resultsFile -Append
"TEST_EXIT_CODE: $testExit" | Out-File $resultsFile -Append
"" | Out-File $resultsFile -Append
"=== COMPLETED at: $(Get-Date) ===" | Out-File $resultsFile -Append
