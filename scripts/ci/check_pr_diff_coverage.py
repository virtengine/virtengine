#!/usr/bin/env python3
"""Fail PR coverage when changed executable Go lines are insufficiently covered."""

from __future__ import annotations

import os
import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path


MIN_COVERAGE = 80.0


def load_changed_lines(base_ref: str) -> dict[str, set[int]]:
    diff = subprocess.check_output(
        ["git", "diff", "--unified=0", "--no-color", f"origin/{base_ref}...HEAD", "--", "*.go"],
        text=True,
    )

    changed_lines: dict[str, set[int]] = defaultdict(set)
    current_file: str | None = None
    new_line: int | None = None

    for raw_line in diff.splitlines():
        if raw_line.startswith("+++ b/"):
            current_file = raw_line[6:]
            continue
        if raw_line.startswith("@@"):
            match = re.search(r"\+(\d+)(?:,(\d+))?", raw_line)
            if not match:
                continue
            new_line = int(match.group(1))
            continue
        if current_file is None or new_line is None:
            continue
        if raw_line.startswith("+") and not raw_line.startswith("+++"):
            changed_lines[current_file].add(new_line)
            new_line += 1
            continue
        if raw_line.startswith("-") and not raw_line.startswith("---"):
            continue
        if raw_line.startswith("\\"):
            continue
        new_line += 1

    return changed_lines


def load_coverage_ranges(coverage_file: Path) -> dict[str, list[tuple[int, int, int]]]:
    coverage_ranges: dict[str, list[tuple[int, int, int]]] = defaultdict(list)
    with coverage_file.open("r", encoding="utf-8") as handle:
        next(handle, None)
        for raw_line in handle:
            line = raw_line.strip()
            if not line:
                continue
            location, _num_statements, count = line.rsplit(" ", 2)
            filename, span = location.split(":", 1)
            start, end = span.split(",", 1)
            start_line = int(start.split(".", 1)[0])
            end_line = int(end.split(".", 1)[0])
            coverage_ranges[filename].append((start_line, end_line, int(count)))
    return coverage_ranges


def matching_ranges(
    coverage_ranges: dict[str, list[tuple[int, int, int]]],
    path: str,
) -> list[tuple[int, int, int]]:
    suffix = f"/{path.replace(os.sep, '/')}"
    for filename, ranges in coverage_ranges.items():
        normalized = filename.replace("\\", "/")
        if normalized == path or normalized.endswith(suffix):
            return ranges
    return []


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print("Usage: check_pr_diff_coverage.py <coverage-file>", file=sys.stderr)
        return 2

    coverage_file = Path(argv[1])
    if not coverage_file.is_file():
        print(f"Coverage file not found: {coverage_file}", file=sys.stderr)
        return 1

    base_ref = os.environ.get("BASE_REF", "").strip()
    if not base_ref:
        print("Missing BASE_REF for PR coverage check", file=sys.stderr)
        return 1

    changed_lines = load_changed_lines(base_ref)
    if not changed_lines:
        print("No changed Go lines detected; skipping PR diff coverage gate.")
        return 0

    coverage_ranges = load_coverage_ranges(coverage_file)
    executable = 0
    covered = 0
    uncovered: list[str] = []

    for path, lines in changed_lines.items():
        ranges = matching_ranges(coverage_ranges, path)
        for line_no in sorted(lines):
            line_ranges = [entry for entry in ranges if entry[0] <= line_no <= entry[1]]
            if not line_ranges:
                continue
            executable += 1
            if any(count > 0 for _, _, count in line_ranges):
                covered += 1
            else:
                uncovered.append(f"{path}:{line_no}")

    if executable == 0:
        print("No executable changed Go lines detected; skipping PR diff coverage gate.")
        return 0

    coverage_pct = (covered * 100.0) / executable
    print(f"Changed Go line coverage: {covered}/{executable} = {coverage_pct:.1f}%")
    if coverage_pct < MIN_COVERAGE:
        preview = ", ".join(uncovered[:20])
        print(
            f"::error::Changed Go line coverage is {coverage_pct:.1f}%, "
            f"below minimum {MIN_COVERAGE:.0f}%"
        )
        if preview:
            print(f"Uncovered changed lines: {preview}")
        return 1

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
