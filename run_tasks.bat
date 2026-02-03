@echo off
cd /d "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine"

echo === Git Status ===
git --no-pager status

echo.
echo === Current Branch ===
git branch --show-current

echo.
echo === Building E2E tests ===
go build -tags=e2e.integration ./tests/e2e/... 2>&1

echo.
echo === Build Exit Code: %ERRORLEVEL% ===

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b 1
)

echo Build succeeded!

echo.
echo === Staging changes ===
git add -A

echo.
echo === Committing ===
git commit -m "test(hpc): implement comprehensive HPC E2E test suite"

echo.
echo === Pushing ===
git push --set-upstream origin vk/42bf-test-hpc-impleme

echo.
echo Done!
