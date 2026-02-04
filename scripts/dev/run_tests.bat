@echo off
cd /d "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

echo === BUILD RESULTS === > results.txt
go build ./x/bme/... ./x/oracle/... >> results.txt 2>&1
if %errorlevel%==0 (
    echo Build passed with no output >> results.txt
    echo BUILD EXIT CODE: 0 >> results.txt
    echo BUILD STATUS: PASSED >> results.txt
) else (
    echo BUILD EXIT CODE: %errorlevel% >> results.txt
    echo BUILD STATUS: FAILED >> results.txt
)

echo. >> results.txt
echo === TEST RESULTS === >> results.txt
go test ./x/bme/keeper/... ./x/oracle/keeper/... >> results.txt 2>&1
set TESTEXIT=%errorlevel%
echo TEST EXIT CODE: %TESTEXIT% >> results.txt
if %TESTEXIT%==0 (
    echo TEST STATUS: PASSED >> results.txt
) else (
    echo TEST STATUS: FAILED >> results.txt
)
