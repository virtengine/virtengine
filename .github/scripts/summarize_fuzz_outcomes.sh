#!/usr/bin/env bash
# Summarize the Fuzz Testing workflow's per-step outcomes, loudly.
#
# WHY THIS EXISTS
# The fuzz steps carry `continue-on-error: true` on purpose: a crash in a
# pre-existing fuzz target must not red the nightly workflow. The consequence
# was that the masking was invisible - the job reported success while every step
# exited 1, and the only trace was in the raw step logs. This prints the outcome
# of each fuzz step into the job summary and raises a warning annotation when one
# did not succeed.
#
# It is a script, not an inline `run:` block, because the inline version shipped a
# false positive: a YAML block scalar's value ends with a newline, so a bare
# `grep -v '=(success|skipped)$'` matched the trailing empty line and warned on a
# run where every step succeeded. Blank lines are ignored here.
#
# Usage: summarize_fuzz_outcomes.sh [summary-file]
#   Reads `name=outcome` lines on stdin (one per fuzz step).
#   Appends the markdown summary to the file (default: $GITHUB_STEP_SUMMARY).
# Always exits 0 - it reports, it does not gate.
set -uo pipefail

summary_file="${1:-${GITHUB_STEP_SUMMARY:-}}"

lines="$(grep -vE '^[[:space:]]*$' || true)"

if [ -z "${lines}" ]; then
  echo "no fuzz step outcomes reported"
  exit 0
fi

if [ -n "${summary_file}" ]; then
  {
    echo "# Fuzz step outcomes"
    echo ""
    echo '```'
    printf '%s\n' "${lines}"
    echo '```'
    echo ""
    echo "These steps are non-gating (\`continue-on-error\`); anything other than"
    echo "\`success\` or \`skipped\` above means that fuzz target did not run cleanly -"
    echo "read the step log for the target that failed."
  } >> "${summary_file}"
fi

printf '%s\n' "${lines}"

unexpected="$(printf '%s\n' "${lines}" | grep -vE '=(success|skipped)$' || true)"
if [ -n "${unexpected}" ]; then
  echo "::warning::fuzz steps that did not succeed: $(printf '%s' "${unexpected}" | tr '\n' ' ')"
fi

exit 0
