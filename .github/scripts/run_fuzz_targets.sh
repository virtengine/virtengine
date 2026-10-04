#!/usr/bin/env bash
# Run every Go fuzz target in the given packages, one target at a time.
#
# WHY THIS EXISTS
# `go test -fuzz=.` requires the fuzz pattern to match EXACTLY ONE fuzz target in
# a package. This repo's fuzzable packages declare 7-14 targets each, so the old
# Fuzz Testing steps aborted before fuzzing anything with:
#
#   testing: will not fuzz, -fuzz matches more than one fuzz test: [...]
#
# and exited 1. Every step was guarded by `continue-on-error: true`, so the job
# still reported success while four of the five fuzz targets never executed -
# a decorative gate. This script fixes the call site: it enumerates each
# package's fuzz targets and runs `-fuzz=^<Target>$` once per target, so every
# target is actually fuzzed and a crashing target is reported by name.
#
# THE TIME BUDGET IS PER STEP, NOT PER TARGET
# `-fuzztime` used to be handed to a single `go test` call, so "5m" meant "this
# step takes about 5m". With one call per target, the same 5m multiplied by the
# number of targets (about 50 across these packages) would blow the 60-minute job
# timeout on the nightly schedule. So the duration passed here stays a per-step
# budget: it is divided across the targets this run discovered, with a floor
# (MIN_TARGET_SECONDS) so an individual target still fuzzes properly.
#
# Usage: run_fuzz_targets.sh <budget> <go-package-pattern> [...]
#   e.g. run_fuzz_targets.sh 30s ./x/veid/types/...
# Budget accepts plain seconds or Go duration style: 45s, 5m, 1m30s, 2h.
#
# Exit codes: 0 every target fuzzed clean, 1 a target failed (or nothing could be
#              fuzzed at all), 2 usage error.
set -uo pipefail

# A target given less than this barely gets past baseline coverage, so the floor
# wins over the division. Keeps PR runs (30s budget) meaningful.
MIN_TARGET_SECONDS=5

usage() {
  echo "usage: $0 <budget> <go-package-pattern> [...]" >&2
}

duration_to_seconds() {
  local d="$1"
  local total=0 num unit rest
  if [ -z "$d" ]; then
    return 1
  fi
  while [ -n "$d" ]; do
    if [[ "$d" =~ ^([0-9]+)([smh])(.*)$ ]]; then
      num="${BASH_REMATCH[1]}"
      unit="${BASH_REMATCH[2]}"
      rest="${BASH_REMATCH[3]}"
      case "$unit" in
        s) total=$((total + num)) ;;
        m) total=$((total + num * 60)) ;;
        h) total=$((total + num * 3600)) ;;
      esac
      d="$rest"
    elif [[ "$d" =~ ^([0-9]+)$ ]]; then
      total=$((total + d))
      d=""
    else
      return 1
    fi
  done
  printf '%s' "$total"
}

budget="${1:-}"
if [ -n "${budget}" ]; then
  shift
fi
if [ -z "${budget}" ] || [ "$#" -eq 0 ]; then
  usage
  exit 2
fi

if ! budget_seconds="$(duration_to_seconds "${budget}")"; then
  echo "::error::unrecognised time budget '${budget}' (want e.g. 30s, 5m, 1m30s)" >&2
  exit 2
fi
if [ "${budget_seconds}" -le 0 ]; then
  echo "::error::time budget must be positive, got '${budget}'" >&2
  exit 2
fi

# Resolve the caller's patterns to concrete packages that actually carry test
# files. `go list` (not `ls`) is what turns `./x/veid/types/...` into the import
# paths `go test` needs, and it is also what makes a pattern matching nothing a
# hard error instead of a silent pass.
if ! packages=$(go list \
  -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' "$@" 2>&1); then
  echo "::error::go list failed for: $*"
  printf '%s\n' "${packages}" | tail -20
  exit 1
fi

# Pass 1: enumerate every fuzz target before running any, so the budget can be
# divided up front. Names must be listed PER PACKAGE - listing a whole tree at
# once prints every package's targets together, and `-fuzz=^Name$` would then
# match nothing in the package that does not declare it.
packages_without_targets=0
target_packages=()
target_names=()

for pkg in ${packages}; do
  if ! listing=$(go test -list '^Fuzz' "${pkg}" 2>&1); then
    echo "::error::could not list fuzz targets in ${pkg}"
    printf '%s\n' "${listing}" | tail -10
    exit 1
  fi
  names=$(printf '%s\n' "${listing}" | grep -E '^Fuzz' || true)
  if [ -z "${names}" ]; then
    packages_without_targets=$((packages_without_targets + 1))
    continue
  fi
  while IFS= read -r name; do
    [ -n "${name}" ] || continue
    target_packages+=("${pkg}")
    target_names+=("${name}")
  done <<< "${names}"
done

targets_seen="${#target_names[@]}"
if [ "${targets_seen}" -eq 0 ]; then
  echo "::error::no fuzz target was found in: $* - this fuzz step is decorative"
  exit 1
fi

per_target_seconds=$((budget_seconds / targets_seen))
if [ "${per_target_seconds}" -lt "${MIN_TARGET_SECONDS}" ]; then
  per_target_seconds="${MIN_TARGET_SECONDS}"
fi
fuzztime="${per_target_seconds}s"

echo "budget ${budget} (${budget_seconds}s) across ${targets_seen} fuzz targets -> ${fuzztime} each"
echo "packages: ${packages} (${packages_without_targets} with no fuzz target)"

# Pass 2: fuzz each target. A crash fails the step, but the remaining targets
# still run so one bad target cannot hide the rest.
overall=0
targets_ok=0

for i in "${!target_names[@]}"; do
  pkg="${target_packages[$i]}"
  name="${target_names[$i]}"
  echo "::group::fuzz ${name} - ${pkg} (${fuzztime})"
  # -run=^$ keeps the package's ordinary tests out of the way so the target is
  # what gets fuzzed; the fuzz target itself is selected by -fuzz.
  if go test -run '^$' -fuzz "^${name}$" -fuzztime="${fuzztime}" "${pkg}"; then
    targets_ok=$((targets_ok + 1))
  else
    status=$?
    echo "::error::fuzz target ${name} failed in ${pkg} (exit ${status})"
    overall=1
  fi
  echo "::endgroup::"
done

echo "fuzz targets: ${targets_ok}/${targets_seen} fuzzed clean (${packages_without_targets} packages had no fuzz target)"

exit "${overall}"
