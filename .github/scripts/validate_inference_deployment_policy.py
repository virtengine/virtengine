#!/usr/bin/env python3

from __future__ import annotations

import argparse
import sys
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


FORBIDDEN_TOKENS = (
    "--allow-fallback-to-stub",
    "local_stub",
)


@dataclass(frozen=True)
class PolicySurface:
    path: Path
    recursive: bool = False


DEFAULT_SURFACES = (
    PolicySurface(ROOT / "deploy", recursive=True),
    PolicySurface(ROOT / "_build", recursive=True),
    PolicySurface(ROOT / ".github" / "workflows", recursive=True),
    PolicySurface(ROOT / "docs", recursive=True),
)


def repo_path(path: Path) -> str:
    try:
        return str(path.relative_to(ROOT)).replace("\\", "/")
    except ValueError:
        return str(path)


def iter_surface_files(surfaces: list[PolicySurface]) -> list[Path]:
    files: list[Path] = []
    for surface in surfaces:
        if not surface.path.exists():
            continue
        if surface.path.is_file():
            files.append(surface.path)
            continue

        pattern = "**/*" if surface.recursive else "*"
        for candidate in sorted(surface.path.glob(pattern)):
            if candidate.is_file():
                files.append(candidate)
    return files


def find_violations(path: Path) -> list[str]:
    try:
        raw = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []

    errors: list[str] = []
    for line_number, line in enumerate(raw.splitlines(), start=1):
        for token in FORBIDDEN_TOKENS:
            if token in line:
                errors.append(
                    f"{repo_path(path)}:{line_number} contains forbidden production inference fallback token {token}"
                )
    return errors


def validate_surfaces(surfaces: list[PolicySurface]) -> list[str]:
    errors: list[str] = []
    for file_path in iter_surface_files(surfaces):
        errors.extend(find_violations(file_path))
    return errors


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate that production deployment assets do not enable inference stub fallback."
    )
    parser.add_argument(
        "--paths",
        nargs="*",
        help="Optional explicit files or directories to validate. Defaults to production deployment surfaces.",
    )
    return parser.parse_args()


def resolve_surfaces(args: argparse.Namespace) -> list[PolicySurface]:
    if not args.paths:
        return list(DEFAULT_SURFACES)

    surfaces: list[PolicySurface] = []
    for item in args.paths:
        path = Path(item)
        resolved = path if path.is_absolute() else ROOT / path
        surfaces.append(PolicySurface(resolved, recursive=resolved.is_dir()))
    return surfaces


def main() -> int:
    args = parse_args()
    errors = validate_surfaces(resolve_surfaces(args))

    if errors:
        for error in errors:
            print(f"FAIL {error}")
        return 1

    print("PASS inference deployment policy")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
