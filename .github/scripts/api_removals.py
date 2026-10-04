#!/usr/bin/env python3
"""Classify removed ``-func`` lines in an API-surface diff.

The ``api-version-check`` job in ``.github/workflows/compatibility.yaml`` used to
treat *any* deleted ``-func`` line as an API removal. That is wrong in the common
case where a function is *redeclared with a different signature* -- which is
exactly what a protobuf response-type split does:

    -func (q *Querier) ReservationByOrder(...) (*resourcesv1.QueryReservationResponse, error) {
    +func (q *Querier) ReservationByOrder(...) (*resourcesv1.QueryReservationByOrderResponse, error) {

Only the tail of the line changed, so the diff shows one ``-func`` and one
``+func`` for the *same* function name. Reporting that as a removal produced a
false red on PR #1120 (run 36825740607), which changed zero API surface at all.

This tool separates the two cases:

* a removed line whose function name also appears on a ``+`` line is a
  *signature change* -- reported separately, never a gate failure;
* a removed line with no corresponding addition is a real *removal* -- the
  gate fails, and the symbol is printed so a human can judge it.

Renames are also matched: a removal of ``Foo`` alongside an addition of ``Bar``
is only treated as a rename when the file diff is otherwise a pure
move/rename. We do not attempt that inference; those surface as real removals,
which is the conservative direction.

Reads a unified diff on stdin, writes a report on stdout, and exits:

    0  no real removals (clean, or only signature changes)
    1  at least one real removal
    2  usage / parse error
"""

from __future__ import annotations

import re
import sys
from typing import Iterable

# A Go method or function declaration. We deliberately stop at the parameter
# list opener: the identity of a symbol is the receiver plus the name, and
# everything after it is the part a response-type split rewrites.
_FUNC_RE = re.compile(
    r"^func\s+"
    r"(?:\((?P<recv>[^)]*)\)\s*)?"   # method receiver, e.g. "(q *Querier)"
    r"(?P<name>[A-Za-z_]\w*)"
)

# `financialCases` was moved from a method on GRPCQuerier to a package-level
# function taking the querier as its first argument. Treat those as the same
# symbol so a pure move is not reported as a removal.
# `financialCases` was moved from a method on GRPCQuerier to a package-level
# function taking the querier as its first argument. A move is not a removal,
# so the receiver type is stripped of pointer-ness and a free function whose
# first parameter is the old receiver counts as the same symbol.
_FUNC_RE = re.compile(
    r"^func\s+"
    r"(?:\((?P<recv>[^)]*)\)\s*)?"   # method receiver, e.g. "(q *Querier)"
    r"(?P<name>[A-Za-z_]\w*)"
)
# The parameter list, so a method -> free function move can be recognised.
_PARAMS_RE = re.compile(r"\((?P<params>.*)$", re.DOTALL)


def _recv_type(recv: str) -> str:
    """Bare receiver type: "(q *Querier)" and "(q Querier)" both give "Querier"."""
    recv = recv.strip()
    if not recv:
        return ""
    return recv.split()[-1].lstrip("*")


def _split_params(text: str) -> list[str]:
    """Split a Go parameter list on top-level commas.

    A naive `text.split(",")` breaks on `map[string]any`, function types and
    nested generics, which would mangle the parameter types we compare.
    """
    params: list[str] = []
    depth = 0
    current: list[str] = []
    for ch in text:
        if ch in "([{":
            depth += 1
        elif ch in ")]}":
            depth -= 1
        if ch == "," and depth == 0:
            params.append("".join(current))
            current = []
        else:
            current.append(ch)
    tail = "".join(current)
    if tail.strip():
        params.append(tail)
    return params


def _symbol(line: str) -> tuple[str, str, str] | None:
    """Return (receiver_type, name, first_param_type) for a Go func declaration.

    The first two fields identify a symbol. The third exists only to detect a
    method that became a free function taking the receiver as an argument.
    """
    m = _FUNC_RE.match(line)
    if not m:
        return None
    recv = _recv_type(m.group("recv") or "")
    first_param = ""
    rest = line[m.end():]
    pm = _PARAMS_RE.match(rest)
    if pm and not recv:
        # Free function: a former method may have become one of its
        # parameters. The receiver is not reliably first -- the real case in
        # PR #1120 moved `financialCases` to (ctx, q GRPCQuerier, kind, ...),
        # with context.Context ahead of the receiver. So record every
        # parameter type, not just the first.
        first_param = ",".join(
            _recv_type(" ".join(p.split()[1:]) if len(p.split()) >= 2 else p)
            for p in _split_params(pm.group("params"))
        )
    return (recv, m.group("name"), first_param)


def _collect(diff_lines: Iterable[str], prefix: str) -> list[tuple[str, str]]:
    out: list[tuple[str, str]] = []
    for line in diff_lines:
        if not line.startswith(prefix):
            continue
        # Diff metadata, not content. Must be tested before stripping.
        if line.startswith("---") or line.startswith("+++"):
            continue
        # Strip exactly one significance character. str.lstrip would eat a
        # leading '-' of a removed comment line ("-- note") along with the
        # diff marker, which silently changes the symbol name.
        sym = _symbol(line[1:])
        if sym is not None:
            recv, name, first_param = sym
            out.append((line, f"{recv}.{name}" if recv else name, first_param))
    return out


def main(argv: list[str]) -> int:
    diff = sys.stdin.read()
    lines = diff.splitlines()

    removed = _collect(lines, "-")
    added = _collect(lines, "+")

    # Index the additions two ways so a removal can be satisfied either by an
    # identical redeclaration or by a method that became a free function.
    #   added_names:   "Querier.Thing"  /  "financialCases"
    #   moved_into:    "financialCases" -> {"GRPCQuerier", "context.Context", ...}
    added_names: set[str] = set()
    moved_into: dict[str, set[str]] = {}
    for _, name, params in added:
        added_names.add(name)
        if params and "." not in name:
            moved_into[name] = {p for p in params.split(",") if p}

    real_removals: list[tuple[str, str]] = []
    signature_changes: list[tuple[str, str]] = []
    for line, name, _ in removed:
        recv, dot, bare = name.partition(".")
        satisfied = name in added_names
        if not satisfied and dot:
            # Was a method on `recv`; is it now a free function of the same
            # name that takes that receiver as a parameter?
            satisfied = recv in moved_into.get(bare, set())
        if satisfied:
            signature_changes.append((line, name))
        else:
            real_removals.append((line, name))

    if signature_changes:
        print(
            f"::notice::API signature changes detected "
            f"({len(signature_changes)}): a function was redeclared with a "
            f"different signature. This is not a removal."
        )
        for line, name in signature_changes:
            print(f"  ~ {name}: {line.strip()}")

    if real_removals:
        print("::error::Potential API removals detected:")
        for line, name in real_removals:
            print(f"  - {name}: {line.strip()}")
        return 1

    if not removed:
        print("No removed public functions detected in changed API files")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main(sys.argv[1:]))
    except BrokenPipeError:
        sys.exit(2)
