#!/usr/bin/env python3
import subprocess
import os

os.chdir(r'C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine')
output_file = r'C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\42bf-test-hpc-impleme\virtengine\_check_output.txt'

with open(output_file, 'w') as f:
    f.write("=" * 60 + "\n")
    f.write("1. CURRENT BRANCH:\n")
    f.write("=" * 60 + "\n")
    result = subprocess.run(['git', 'branch', '--show-current'], capture_output=True, text=True)
    f.write((result.stdout.strip() or result.stderr.strip() or "(empty output)") + "\n")

    f.write("\n" + "=" * 60 + "\n")
    f.write("2. GIT STATUS (modified/new files):\n")
    f.write("=" * 60 + "\n")
    result = subprocess.run(['git', 'status', '--short'], capture_output=True, text=True)
    f.write((result.stdout.strip() or result.stderr.strip() or "(no changes)") + "\n")

    f.write("\n" + "=" * 60 + "\n")
    f.write("3. GO BUILD ERRORS:\n")
    f.write("=" * 60 + "\n")
    result = subprocess.run(
        ['go', 'build', '-tags=e2e.integration', './tests/e2e/...'],
        capture_output=True,
        text=True
    )
    if result.returncode == 0:
        f.write("BUILD SUCCESSFUL - No compilation errors\n")
    else:
        f.write("BUILD FAILED:\n")
        f.write(result.stderr + "\n")
        if result.stdout:
            f.write(result.stdout + "\n")

print("Done - output written to", output_file)
