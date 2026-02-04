@echo off
cd /d C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine
echo Starting build at %TIME%
go build -tags=e2e.integration ./tests/e2e/... > e2e_build_result.txt 2>&1
echo Exit code: %ERRORLEVEL% >> e2e_build_result.txt
echo Build completed at %TIME%
