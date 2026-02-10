Set-Location "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine"
$output = go build -tags=e2e.integration ./tests/e2e/... 2>&1
$output | Out-File -FilePath "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine\build_result.txt"
Write-Output "BUILD COMPLETE"
