# gosecsuppressions — CI guard for suppression annotations the gate cannot read

**What this rejects.** A `//nolint:gosec` annotation that is not accompanied by a
`#nosec` the `gosec Security Scan` gate can read.

**Why that is a problem.** The gate runs standalone gosec v2.25.0, which honours only
`#nosec`. It does not read `//nolint`. So a `//nolint:gosec`-only annotation silences
`golangci-lint` and the author's editor while every finding still lands in the security
gate — a suppression that reports itself as suppressed to the tool you are looking at
and does nothing to the tool that blocks the merge. PR #1254 put 7 findings into the
gate that way, and nothing caught it before CI.

Full write-up, with the measured shape matrix, in
[`_docs/security/gosec-triage.md`](../../_docs/security/gosec-triage.md) section 9.5.1.

## Usage

```bash
go run ./scripts/gosecsuppressions -root .            # grade the tree (what CI runs)
go run ./scripts/gosecsuppressions -root . -json      # machine-readable
go run ./scripts/gosecsuppressions -root . -write-baseline   # re-record; deliberate, never in CI
```

## Exit codes

| code | meaning |
| --- | --- |
| 0 | every NEW gosec-suppressing annotation is one the gate can read |
| 1 | a new/raised annotation, a missing baseline, or an expired review window |
| 2 | could not walk the tree / a file did not parse / baseline unreadable — **deliberately not a pass** |

Exit 2 exists so an unreadable estate can never report clean.

## The ratchet

This tree already carries **366** unreadable annotations that predate the guard. A hard
fail would red every gate on the day it lands and teach everyone to ignore it, so the
guard is ratcheted against a recorded baseline
([`gosec-suppressions.baseline`](gosec-suppressions.baseline)).

The baseline is keyed on `path<TAB>comment-text`, **not** on line number: inserting a line
shifts every line number in the file, so a line-keyed baseline would detonate into
hundreds of false failures on ordinary edits and be turned off within a week.

Three rules, no fourth:

- an annotation whose `path+text` is **not in the baseline fails** — this is the
  new-file case, and the only rule that catches it;
- a count rising for a known key **fails**;
- keys or counts falling is an **improvement**: reported, never failed. Re-record with
  `-write-baseline` so improvements stick.

The budget is reviewable and expires. Raising it means editing `defaultAllowance`,
`ratchetExpires` and `budgetIssue` in [`baseline.go`](baseline.go) — which puts the
decision in a diff next to the review date, rather than in a silently widened env var.

`GOSEC_SUPPRESSION_MAX` may only **tighten** the allowance. It is refused (not clamped)
above the hard ceiling, and refused above the recorded allowance: an override that can
raise the ceiling is a laundering vector, because it buys headroom with no diff in review.

## What it does not do

- It does **not** run gosec, and does not decide whether a finding is real. It grades
  annotation *form*. A justified `#nosec` stays justified.
- It does **not** reject `#nosec`. That is the annotation the gate honours, and a
  reviewed finding with a written justification is the policy in section 5 of the triage
  doc.
- It does **not** reject a `//nolint:gosec` that is *also* paired with a readable
  `#nosec`, because that combination genuinely suppresses both tools.

## Proving it works

```bash
go test ./scripts/gosecsuppressions/   # unit + end-to-end exit-code rows
bash scripts/gosecsuppressions/probe.sh   # real binary, real trees, real exit codes
```

The probe does not import the guard's functions: the claim under test is what the CI
*step* sees — the process exit code — not what a unit test in the same package sees.