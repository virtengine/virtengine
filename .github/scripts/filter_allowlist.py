#!/usr/bin/env python3
# Filter govulncheck findings against the vulnerability allowlist.
# Reads JSON report from file, allowlist from file, writes filtered findings to stdout.
#
# Exit codes (the workflow branches on these; see security.yaml go-vuln-scan):
#   0  clean          - scan completed and every finding is allowlisted
#   2  usage/IO error
#   3  vulnerabilities - un-allowlisted findings remain
#   4  incomplete     - the report did not parse to the end, so NO verdict can
#                       be claimed. This is deliberately NOT 0: a scan that was
#                       killed mid-write must never be reported as clean.
#
# govulncheck -format json emits a STREAM of concatenated JSON objects rather
# than one document, so a naive json.load fails with 'Extra data'. We walk the
# stream with raw_decode and require that it be consumed exactly: any
# unparseable remainder means the report is incomplete.

import json
import sys
from pathlib import Path

import yaml

# A real govulncheck report for this repo is ~940KB across ~390 top-level
# objects. Allow generous headroom so a genuinely larger scan is not called
# broken, while still bounding the work on a corrupt/huge input.
MAX_REPORT_BYTES = 64 * 1024 * 1024


def load_allowlist(path):
    with open(path, 'r') as f:
        data = yaml.safe_load(f)
    allowed_ids = set()
    for ecosystem, entries in (data or {}).get('exceptions', {}).items():
        if isinstance(entries, list):
            for entry in entries:
                if isinstance(entry, dict) and 'id' in entry:
                    allowed_ids.add(entry['id'])
    return allowed_ids


def extract_findings(content):
    """Walk the concatenated-JSON stream, returning (findings, complete).

    `complete` is False when the stream could not be parsed to its end. The
    caller must treat that as "no verdict" rather than "no findings".
    """
    findings = []
    decoder = json.JSONDecoder()
    idx = 0
    length = len(content)
    while idx < length:
        while idx < length and content[idx].isspace():
            idx += 1
        if idx >= length:
            break
        try:
            obj, idx = decoder.raw_decode(content, idx)
        except json.JSONDecodeError:
            # An object we cannot decode is a truncated or corrupt report. Stop
            # here and report incompleteness; do NOT skip ahead and claim the
            # remainder was clean.
            return findings, False
        if isinstance(obj, dict) and 'finding' in obj:
            findings.append(obj['finding'])
    return findings, True


def unique_ids(findings):
    """Sorted unique advisory IDs, skipping findings with no OSV id."""
    return sorted({f.get('osv') for f in findings if f.get('osv')})


def main():
    if len(sys.argv) < 3:
        print("usage: filter_allowlist.py <report.json> <allowlist.yaml>", file=sys.stderr)
        return 2

    report_path = sys.argv[1]
    allowlist_path = sys.argv[2]

    report = Path(report_path)
    if not report.is_file():
        print(f"error: govulncheck report not found: {report}", file=sys.stderr)
        return 2
    if report.stat().st_size > MAX_REPORT_BYTES:
        print(
            f"error: govulncheck report is {report.stat().st_size} bytes, "
            f"above the {MAX_REPORT_BYTES} byte ceiling; refusing to parse",
            file=sys.stderr,
        )
        return 2

    try:
        content = report.read_text(encoding='utf-8', errors='replace')
    except OSError as exc:
        print(f"error: cannot read {report}: {exc}", file=sys.stderr)
        return 2

    try:
        allowed_ids = load_allowlist(Path(allowlist_path))
    except (OSError, yaml.YAMLError) as exc:
        print(f"error: cannot read allowlist {allowlist_path}: {exc}", file=sys.stderr)
        return 2

    findings, complete = extract_findings(content)
    filtered = [f for f in findings if f.get('osv') not in allowed_ids]

    # stdout: the un-allowlisted findings, one JSON object per line.
    for f in filtered:
        print(json.dumps(f))

    # stderr: the summary. Every advisory the scan saw is named, so a reviewer
    # can see what the allowlist suppressed and what is still outstanding.
    detected = unique_ids(findings)
    remaining = unique_ids(filtered)
    if complete:
        print(
            f"Total findings: {len(findings)}, After allowlist: {len(filtered)}",
            file=sys.stderr,
        )
        print(f"Advisories detected ({len(detected)}): {', '.join(detected) or 'none'}",
              file=sys.stderr)
        if remaining:
            print(f"Un-allowlisted advisories ({len(remaining)}): {', '.join(remaining)}",
                  file=sys.stderr)
        else:
            print("Un-allowlisted advisories (0): none", file=sys.stderr)
    else:
        print(
            "govulncheck report is INCOMPLETE - the JSON stream did not parse to "
            "the end, so no clean verdict can be claimed. Findings decoded before "
            f"the break: {len(findings)}",
            file=sys.stderr,
        )
        if detected:
            print(f"Advisories decoded before the break: {', '.join(detected)}",
                  file=sys.stderr)
        if remaining:
            print(f"Un-allowlisted advisories before the break ({len(remaining)}): "
                  f"{', '.join(remaining)}", file=sys.stderr)

    if not complete:
        return 4
    return 0 if not filtered else 3


if __name__ == '__main__':
    sys.exit(main())
