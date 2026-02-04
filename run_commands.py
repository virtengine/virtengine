import subprocess
import os

os.chdir(r"C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine")

# Run build
build_result = subprocess.run(
    ["go", "build", "./x/bme/...", "./x/oracle/..."],
    capture_output=True, text=True, timeout=600
)
build_output = build_result.stdout + build_result.stderr
if not build_output.strip():
    build_output = "Build passed with no output"
build_exit = build_result.returncode
build_status = "PASSED" if build_exit == 0 else "FAILED"

# Run tests
test_result = subprocess.run(
    ["go", "test", "./x/bme/keeper/...", "./x/oracle/keeper/..."],
    capture_output=True, text=True, timeout=600
)
test_output = test_result.stdout + test_result.stderr
test_exit = test_result.returncode
test_status = "PASSED" if test_exit == 0 else "FAILED"

# Write results
with open("results.txt", "w", encoding="utf-8") as f:
    f.write(f"""=== BUILD RESULTS ===
{build_output}
BUILD EXIT CODE: {build_exit}
BUILD STATUS: {build_status}

=== TEST RESULTS ===
{test_output}
TEST EXIT CODE: {test_exit}
TEST STATUS: {test_status}
""")

print(f"Build: {build_status} (exit {build_exit})")
print(f"Tests: {test_status} (exit {test_exit})")
print("Results written to results.txt")
