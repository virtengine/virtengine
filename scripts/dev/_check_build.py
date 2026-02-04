#!/usr/bin/env python3
"""Check go build status for e2e tests."""
import subprocess
import sys
import os

os.chdir(r"C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine")

print("Starting build check...")
print("Working directory:", os.getcwd())

try:
    result = subprocess.run(
        ["go", "build", "-tags=e2e.integration", "./tests/e2e/..."],
        capture_output=True,
        text=True,
        timeout=1200  # 20 minute timeout
    )
    
    print("\n=== STDOUT ===")
    print(result.stdout if result.stdout else "(empty)")
    
    print("\n=== STDERR ===")
    print(result.stderr if result.stderr else "(empty)")
    
    print(f"\n=== EXIT CODE: {result.returncode} ===")
    
    if result.returncode == 0:
        print("\n*** Build succeeded with no errors ***")
    else:
        print("\n*** Build FAILED ***")
        
except subprocess.TimeoutExpired:
    print("ERROR: Build timed out after 20 minutes")
except Exception as e:
    print(f"ERROR: {e}")
