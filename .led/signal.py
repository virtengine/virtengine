"""Extract the real failure signal from each recovered job log (drop post-job noise)."""
import glob
import os
import re

D = os.path.join(os.path.dirname(os.path.abspath(__file__)), "logs")

# Real-signal patterns: errors, failures, the command that failed, non-zero exits.
SIG = re.compile(
    r"(##\[error\]|::error|::warning::[A-Z]|exit status [1-9]|exit code [1-9]|"
    r"\bFAIL\b|\bERROR\b|\berror:|\bfailed\b|\bFailure\b|\bdenied\b|\bviolations?\b|"
    r"\bcannot\b|\bfind: |command not found|Traceback \(most recent|"
    r"Error:|npm ERR!|pnpm.*ERR|deprecated|out of date|"
    r"Breaking change|BREAKING|is not compatible|not found)",
    re.I,
)
# Lines that are pure GH Actions bookkeeping, never the cause.
JUNK = re.compile(
    r"(Post job cleanup|Cleaning up orphan|^\s*\[command\]|Temporarily overriding HOME|"
    r"Adding repository directory|safe\.directory|core\.sshCommand|extraheader|"
    r"http\.https://github\.com|includeIf\.gitdir|remote\.origin\.url|"
    r"Node\.js 20 is deprecated|Artifact download URL|Artifact ID|"
    r"Successfully uploaded|final size is|##\[group\]|##\[endgroup\]|"
    r"##\[(start|end)-action|Cleaning up certificates|Post cache|State not set|"
    r"Download action repository|Prepare workflow directory|Complete job|"
    r"Restore cache|Save cache|runner name|Runner Image|Current runner version)",
    re.I,
)

for path in sorted(glob.glob(os.path.join(D, "*.log"))):
    name = os.path.basename(path)
    hits, seen = [], set()
    for line in open(path, encoding="utf-8", errors="replace"):
        if JUNK.search(line) or not SIG.search(line):
            continue
        # strip the "<job>\t<STEP>\t<ts> " prefix
        body = re.sub(r"^.*?\t[^\t]*\t\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?", "", line.rstrip())
        body = re.sub(r"\x1b\[[0-9;]*m", "", body).strip()
        if not body or body in seen:
            continue
        seen.add(body)
        hits.append(body)
    print("=" * 100)
    print(name)
    print("-" * 100)
    for h in hits[:40]:
        print("  *", h[:400])
    if not hits:
        print("  <no signal lines matched>")
    print()