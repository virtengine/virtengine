@echo off
cd /d "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine"
go build -tags=e2e.integration ./tests/e2e/... 2>&1
echo EXIT_CODE=%ERRORLEVEL%
