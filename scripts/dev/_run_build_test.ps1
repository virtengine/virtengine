$ErrorActionPreference = "Continue"
Set-Location "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

$outFile = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine\_output.log"

"=== STARTING BUILD ===" | Out-File $outFile
"Command: go build ./x/bme/... ./x/oracle/..." | Add-Content $outFile
"Started: $(Get-Date)" | Add-Content $outFile
"" | Add-Content $outFile

$buildOutput = go build -v ./x/bme/... ./x/oracle/... 2>&1 | Out-String
$buildCode = $LASTEXITCODE
$buildOutput | Add-Content $outFile

"" | Add-Content $outFile
"BUILD_EXIT_CODE: $buildCode" | Add-Content $outFile
"Completed: $(Get-Date)" | Add-Content $outFile
"" | Add-Content $outFile

"=== STARTING TESTS ===" | Add-Content $outFile
"Command: go test ./x/bme/keeper/... ./x/oracle/keeper/..." | Add-Content $outFile
"Started: $(Get-Date)" | Add-Content $outFile
"" | Add-Content $outFile

$testOutput = go test -v ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1 | Out-String
$testCode = $LASTEXITCODE
$testOutput | Add-Content $outFile

"" | Add-Content $outFile
"TEST_EXIT_CODE: $testCode" | Add-Content $outFile
"Completed: $(Get-Date)" | Add-Content $outFile
"" | Add-Content $outFile
"=== ALL DONE ===" | Add-Content $outFile
