@echo off
cd /d "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

echo === COMMAND 1: go build ./x/bme/... ./x/oracle/... ===
go build ./x/bme/... ./x/oracle/... 2>&1
echo BUILD_EXIT_CODE: %ERRORLEVEL%

echo.
echo === COMMAND 2: go test ./x/bme/keeper/... ./x/oracle/keeper/... ===
go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1
echo TEST_EXIT_CODE: %ERRORLEVEL%
