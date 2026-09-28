#!/usr/bin/env python3
# Filter govulncheck findings against the vulnerability allowlist.
# Reads JSON report from file, allowlist from file, writes filtered findings to stdout.
# Exits 0 on success (no un-allowlisted findings), 3 if vulnerabilities remain, non-zero on error.

import json
import sys
import yaml
from pathlib import Path

def load_allowlist(path):
    with open(path, 'r') as f:
        data = yaml.safe_load(f)
    allowed_ids = set()
    for ecosystem, entries in data.get('exceptions', {}).items():
        if isinstance(entries, list):
            for entry in entries:
                if isinstance(entry, dict) and 'id' in entry:
                    allowed_ids.add(entry['id'])
    return allowed_ids

def extract_findings(report_path):
    """Extract finding objects from the JSON report stream (concatenated JSON objects)."""
    findings = []
    with open(report_path, 'r') as f:
        content = f.read()
    
    decoder = json.JSONDecoder()
    idx = 0
    while idx < len(content):
        # Skip whitespace
        while idx < len(content) and content[idx].isspace():
            idx += 1
        if idx >= len(content):
            break
        try:
            obj, idx = decoder.raw_decode(content, idx)
            if isinstance(obj, dict) and 'finding' in obj:
                findings.append(obj['finding'])
        except json.JSONDecodeError:
            # Try to skip to next potential object start
            idx = content.find('{', idx + 1)
            if idx == -1:
                break
    return findings

def main():
    if len(sys.argv) < 3:
        print("usage: filter_allowlist.py <report.json> <allowlist.yaml>", file=sys.stderr)
        return 2

    report_path = sys.argv[1]
    allowlist_path = sys.argv[2]

    allowed_ids = load_allowlist(Path(allowlist_path))
    findings = extract_findings(report_path)

    # Filter out allowlisted findings
    filtered = [f for f in findings if f.get('osv') not in allowed_ids]

    # Output filtered findings as JSON lines (one per line)
    for f in filtered:
        print(json.dumps(f))

    # Also print summary to stderr
    if findings:
        print(f"Total findings: {len(findings)}, After allowlist: {len(filtered)}", file=sys.stderr)
        if filtered:
            osv_ids = sorted(set(f.get('osv') for f in filtered if f.get('osv')))
            print(f"Un-allowlisted advisories: {', '.join(osv_ids)}", file=sys.stderr)
    else:
        print("No findings in report", file=sys.stderr)

    # Exit code: 0 if clean (no un-allowlisted findings), 3 if vulnerabilities remain
    return 0 if not filtered else 3

if __name__ == '__main__':
    sys.exit(main())
