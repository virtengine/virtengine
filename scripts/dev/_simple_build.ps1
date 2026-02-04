$outFile = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine\_build_test_results.txt"
$workDir = "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

"Script started at $(Get-Date)" | Out-File $outFile -Force
"Working dir: $workDir" | Out-File $outFile -Append
"" | Out-File $outFile -Append

try {
    Set-Location $workDir
    
    "=== COMMAND 1: go build ./x/bme/... ./x/oracle/... ===" | Out-File $outFile -Append
    "Start: $(Get-Date)" | Out-File $outFile -Append
    
    $buildOutput = & go build ./x/bme/... ./x/oracle/... 2>&1 | Out-String
    $buildExit = $LASTEXITCODE
    
    "STDOUT/STDERR:" | Out-File $outFile -Append
    $buildOutput | Out-File $outFile -Append
    "Exit Code: $buildExit" | Out-File $outFile -Append
    "End: $(Get-Date)" | Out-File $outFile -Append
    "" | Out-File $outFile -Append
    
    "=== COMMAND 2: go test ./x/bme/keeper/... ./x/oracle/keeper/... ===" | Out-File $outFile -Append
    "Start: $(Get-Date)" | Out-File $outFile -Append
    
    $testOutput = & go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1 | Out-String
    $testExit = $LASTEXITCODE
    
    "STDOUT/STDERR:" | Out-File $outFile -Append
    $testOutput | Out-File $outFile -Append
    "Exit Code: $testExit" | Out-File $outFile -Append
    "End: $(Get-Date)" | Out-File $outFile -Append
    
} catch {
    "ERROR: $_" | Out-File $outFile -Append
}

"Script completed at $(Get-Date)" | Out-File $outFile -Append
