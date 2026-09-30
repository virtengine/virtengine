"""Per-job deep dive: show the failing step region for specific jobs, unfiltered-ish."""
import re
import sys

JOBS = {
    "Proto Breaking Checks": "36073455131-107879358144-Proto-Breaking-Checks.log",
    "Compatibility Tests": "36073455131-107879358471-Compatibility-Tests.log",
    "Supply Chain Attack Detection": "36073455817-107879360486-Supply-Chain-Attack-Detection.log",
    "Dependency Risk Assessment": "36073455817-107879360405-Dependency-Risk-Assessment.log",
    "Go Tests": "36073455679-107879836125-Go-Tests.log",
    "Windows Native": "36073455679-107879835955-Windows-Native-Build-and-Unit-Tests.log",
}

D = "C:/Users/jON/virtengine-ops/virtengine/.worktrees/t_baf629c7/.led/logs/"

for label, fn in JOBS.items():
    path = D + fn
    lines = open(path, encoding="utf-8", errors="replace").read().splitlines()
    # Index of the first ##[error]Process completed line = end of the run
    end = len(lines)
    for i, l in enumerate(lines):
        if "Process completed with exit code" in l:
            end = i
            break
    body = []
    for l in lines[: end + 2]:
        b = re.sub(r"^.*?\t[^\t]*\t\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?", "", l)
        b = re.sub(r"\x1b\[[0-9;]*m", "", b).rstrip()
        body.append(b)
    print("#" * 100)
    print(f"### {label}  ({fn})  total={len(lines)} error_at={end}")
    print("#" * 100)
    if label == "Go Tests":
        for b in body:
            if re.search(r"(^\s*--- FAIL|^FAIL|^ok .*FAIL|panic:|DATA RACE|coverage:.*fail)", b) or re.match(r"^--- FAIL", b.strip()):
                print("  *", b.strip()[:300])
        continue
    if label == "Windows Native":
        for b in body:
            if re.search(r"(--- FAIL|^FAIL|Error:|persisted_schema_test|not found|Received unexpected)", b):
                print("  *", b.strip()[:300])
        continue
    # the others: last 90 meaningful lines before the exit
    keep = [b for b in body if b.strip() and "Temporarily overriding HOME" not in b
            and "safe.directory" not in b and "sshCommand" not in b
            and "extraheader" not in b and "includeIf" not in b
            and "remote.origin.url" not in b and "DeprecationWarning" not in b]
    for b in keep[-90:]:
        print("  ", b[:280])
    print()