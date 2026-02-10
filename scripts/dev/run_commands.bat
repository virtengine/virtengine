@echo off
cd /d C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine

set OUTFILE=results.txt
set BUILD_TEMP=build_temp.txt
set TEST_TEMP=test_temp.txt

REM Run build and capture output
go build ./x/bme/... ./x/oracle/... > %BUILD_TEMP% 2>&1
set BUILD_EXIT=%ERRORLEVEL%

REM Run tests and capture output  
go test ./x/bme/keeper/... ./x/oracle/keeper/... > %TEST_TEMP% 2>&1
set TEST_EXIT=%ERRORLEVEL%

REM Determine statuses
if %BUILD_EXIT%==0 (
    set BUILD_STATUS=PASSED
) else (
    set BUILD_STATUS=FAILED
)

if %TEST_EXIT%==0 (
    set TEST_STATUS=PASSED
) else (
    set TEST_STATUS=FAILED
)

REM Write final results file
echo === BUILD RESULTS === > %OUTFILE%
for %%I in (%BUILD_TEMP%) do if %%~zI==0 (
    echo Build passed with no output >> %OUTFILE%
) else (
    type %BUILD_TEMP% >> %OUTFILE%
)
echo BUILD EXIT CODE: %BUILD_EXIT% >> %OUTFILE%
echo BUILD STATUS: %BUILD_STATUS% >> %OUTFILE%
echo. >> %OUTFILE%
echo === TEST RESULTS === >> %OUTFILE%
type %TEST_TEMP% >> %OUTFILE%
echo TEST EXIT CODE: %TEST_EXIT% >> %OUTFILE%
echo TEST STATUS: %TEST_STATUS% >> %OUTFILE%

REM Cleanup
del %BUILD_TEMP% 2>nul
del %TEST_TEMP% 2>nul

echo Done - results written to %OUTFILE%
