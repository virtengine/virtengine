@echo off
setlocal EnableDelayedExpansion

REM ============================================================
REM HPC Proto Setup Script
REM Copies proto files to correct location and runs buf generate
REM ============================================================

echo =====================================================
echo HPC Proto Generation Setup
echo =====================================================
echo.

REM Step 1: Create HPC proto directories
echo Step 1: Creating HPC proto directories...
if not exist "sdk\proto\node\virtengine\hpc\v1" (
    mkdir "sdk\proto\node\virtengine\hpc\v1" 2>nul
    if errorlevel 1 (
        echo ERROR: Failed to create proto directory
        exit /b 1
    )
)
if not exist "sdk\go\node\hpc\v1" (
    mkdir "sdk\go\node\hpc\v1" 2>nul
)
echo   Created: sdk\proto\node\virtengine\hpc\v1
echo   Created: sdk\go\node\hpc\v1
echo.

REM Step 2: Copy proto files
echo Step 2: Copying proto files...
copy /Y "hpc_types.proto.txt" "sdk\proto\node\virtengine\hpc\v1\types.proto" >nul
if errorlevel 1 (
    echo ERROR: Failed to copy types.proto
    exit /b 1
)
echo   Copied: types.proto

copy /Y "hpc_tx.proto.txt" "sdk\proto\node\virtengine\hpc\v1\tx.proto" >nul
if errorlevel 1 (
    echo ERROR: Failed to copy tx.proto
    exit /b 1
)
echo   Copied: tx.proto

copy /Y "hpc_query.proto.txt" "sdk\proto\node\virtengine\hpc\v1\query.proto" >nul
if errorlevel 1 (
    echo ERROR: Failed to copy query.proto
    exit /b 1
)
echo   Copied: query.proto

copy /Y "hpc_genesis.proto.txt" "sdk\proto\node\virtengine\hpc\v1\genesis.proto" >nul
if errorlevel 1 (
    echo ERROR: Failed to copy genesis.proto
    exit /b 1
)
echo   Copied: genesis.proto
echo.

REM Step 3: Verify files
echo Step 3: Verifying proto files...
dir /b "sdk\proto\node\virtengine\hpc\v1\*.proto"
echo.

REM Step 4: Instructions for proto generation
echo =====================================================
echo SUCCESS: Proto files copied to correct location
echo =====================================================
echo.
echo Proto files are now in:
echo   sdk\proto\node\virtengine\hpc\v1\types.proto
echo   sdk\proto\node\virtengine\hpc\v1\tx.proto
echo   sdk\proto\node\virtengine\hpc\v1\query.proto
echo   sdk\proto\node\virtengine\hpc\v1\genesis.proto
echo.
echo Next steps:
echo   1. cd sdk
echo   2. buf generate (or run: ./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go)
echo   3. Generated Go files will be in: sdk/go/node/hpc/v1/
echo.
echo After generation, clean up temporary files:
echo   del hpc_*.proto.txt setup_hpc_proto.bat setup_hpc_dirs.js HPC_PROTO_README.md
echo.
echo =====================================================

endlocal
