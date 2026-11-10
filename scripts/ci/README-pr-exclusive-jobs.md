# PR-exclusive job census

`node scripts/ci/check-pr-exclusive-jobs.mjs` enumerates every job in every
workflow in this repo and reports which ones **cannot fail before a change is
merged**.

## Why this exists

Measured 2026-10-05: `Build (macOS)` was red on 3 of 3 runs since PR #967 and had
never once been green, while every `develop` PR was green. The cause was a
composite-action defect — `setup-macos` exported `GO_LINKMODE=internal`
unconditionally, and the Makefile declares `GO_LINKMODE ?= external`, which
*yields to the environment*. With `BUILD_TAGS` including `hidraw` (a cgo-only
package), `internal` is illegal on that runner.

The fix landed in PR #1269. But the reason it *could* land undetected is the
structural point: that job is gated

```yaml
if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/v')
```

so the defect could only ever appear **after** the merge that carried it. That is
the whole class this census covers.

## What it reports

The live census on `develop` @ `2617aa365`:

```
37 workflows, 187 jobs — 96 cannot fail before merge, 91 are observable pre-merge
```

The 96 break down as:

| kind | count | example |
|---|---|---|
| deploy path (`workflow_dispatch`-only workflow) | ~78 | `multi-region-deploy.yaml :: Deploy: Global resources` |
| release artifact (tags/main-only) | 14 | `ci.yaml :: Build (macOS)`, `supply-chain.yaml :: Sign Release Artifacts` |
| schedule-only | ~4 | `stale.yaml :: stale-pr`, `pr-shepherd.yaml :: shepherd` |

**A post-merge-only job is not a defect.** `if: github.ref == 'refs/heads/main'`
on a production deploy is *correct*. The census is report-only for that reason:
failing the build for it would train everyone to delete the gate instead of
reasoning about it. What it buys is that every such job is now **enumerated in
review**, so a newly added one — or a changed `if:` — shows up as a visible diff.

`JUSTIFIED` in the script records the jobs whose exclusion is deliberate; `--strict`
turns an unrecorded one into a failure for callers that want the ratchet.

## Why the gate is fail-closed

Exit 2 means the census could not build its evidence. That is deliberately **not**
a pass: a census that silently measured nothing is exactly the failure mode this
repo keeps guarding against. It exits 2 when there are no workflow files, or any
workflow yields zero parsed jobs.

## How it classifies

A job is PR-EXCLUSIVE when no `pull_request` event can satisfy it. Two
independent causes are tracked separately:

- **workflow** — the workflow's `on:` never names a PR event
- **job-if** — the job's own `if:` demands a ref a PR does not have. GitHub gives
  a PR run the synthetic ref `refs/pull/<n>/merge`, so a comparison against
  `github.ref` naming a branch excludes it.

The OR case is the one a naive `contains github.ref` check gets wrong, and case 9
pins it: `github.ref == 'refs/heads/main' || github.event_name == 'pull_request'`
*is* satisfiable on a PR.

## Self-test

`--selftest` runs 23 cases and 6 mutations, and the job's first step runs it. A
census that cannot detect a known PR-exclusive job is worse than no census, so
every classifier branch is pinned, including against mutation.

The **control row must report INERT**. It is the only row proving the harness
measures instead of always answering CAUGHT. Two harness defects were found by
that row during development and are worth not repeating:

1. The control was anchored to a string that *also* occurred in the mutation
   table's own `src` literal, so `replace` mutated the table instead of the
   module and the control came back CAUGHT. Anchors are now asserted unique
   (`ERROR:anchor-found-Nx`).
2. The mutated file was rebuilt as `src.slice(0, region.length) + … +
   src.slice(region.length)`. Because `replace` changes the region's length, the
   tail was sliced at the wrong offset and produced a 980-line file from a
   594-line source, with a second shebang at line 387. Node rejects that as
   `Invalid or unexpected token`, which made **every** mutant report ERROR — a
   harness reporting a uniform failure, which is what a broken harness looks like.
   The tail is now sliced from the original by offset and re-joined.