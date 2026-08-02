# Prototype Thread Intake Runbook

## Purpose

This runbook is authoritative for the five-thread prototype campaign. It
overrides generic branch and handoff guidance when integrating T1, T2, T3, and
T5 into T4.

## Frozen Intake Epochs

T4 opens a batch by publishing an immutable annotated tag:

```text
checkpoint/prototype-integration/epoch-N-base
```

The tag target is the frozen `intake_base_sha`. Producers synchronize with that
SHA once for the batch. Later T4 commits do not invalidate producer checkpoints
within the same epoch.

T4 also commits an epoch manifest with epoch number, base tag/SHA, planning SHA,
`opens_at`, `announcement_cutoff`, and producer states. A checkpoint is announced
only when its remote annotated tag exists before the cutoff. At cutoff, T4 freezes
the announced tag list. Unannounced work automatically moves to the next epoch;
T4 records it as `status: unannounced`, `tag: null`, and `decision: frozen-out`
before treating it as terminal for closure. T4 accepts or rejects every frozen
tag, records the decision, closes the epoch, then opens the next. A rejected payload is immutable;
a correction receives a new checkpoint ID in a later epoch.

A producer merges the epoch base into its branch before validation. It must not
rebase or rewrite shared history.

## Producer Publication

Each producer checkpoint has two commits:

1. `payload_head`: the last implementation/test commit;
2. handoff commit: a direct descendant that changes only the handoff and retained
   evidence paths.

The immutable annotated checkpoint tag targets the handoff commit, not
`payload_head`. The handoff does not attempt to record its own SHA. T4 resolves
that SHA from the remote tag.

Required handoff fields:

```yaml
campaign: three-day-prototype
thread: T2
checkpoint: T2-03B
branch: ve/prototype-t2-product
frozen_baseline: 79391a3df86d85522b92e0400c6904971ecbe65d
planning_sha: <stable planning commit consumed by the thread>
intake_epoch: 1
intake_base_sha: <epoch-N-base tag target>
payload_head: <last implementation/test commit>
prior_accepted_payload: <sha-or-null>
tree_clean: true
commits_since_prior_acceptance: []
owned_paths: []
files_changed: []
tests: []
generated_hashes: {}
migrations: []
external_evidence: []
known_failures: []
blockers: []
next_checkpoint: T2-04
```

Every test entry contains the literal command, exit code, `result: passed`, test
count when available, exact tool version output, and a retained output artifact
under `_docs/ralph/evidence/prototype-tN/<checkpoint>/`. A summary is not an
artifact.

Publish branch and tag:

```powershell
git push -u origin HEAD:ve/prototype-t2-product
git tag -a checkpoint/prototype-t2/t2-03b -m "Prototype T2 checkpoint 03B"
git push origin refs/tags/checkpoint/prototype-t2/t2-03b
```

## T4 Intake

T4 performs intake only from a clean worktree. Its own dirty checkpoint must be
completed or preserved before intake begins.

For each producer tag, T4 must:

1. run `git fetch --prune origin` and resolve the remote annotated tag;
2. load the committed handoff with `git show <tag>:<handoff-path>`;
3. validate the handoff against the actual schema;
4. prove `payload_head` descends from `intake_base_sha`;
5. prove the tag target descends directly from `payload_head` through only
   declared handoff/evidence commits;
6. compare the complete changed-file range with ownership and lease rules;
7. verify branch/tag publication, exact commit lists, generated hashes,
   migrations, retained output artifacts and pinned tool versions;
8. reject duplicate acceptance and stale or unknown epochs;
9. merge the declared range with a normal merge commit;
10. run focused producer gates and cross-thread integration gates before updating
    the accepted ledger.

The validator must include negative tests for missing refs, invalid schema,
wrong baseline/planning SHA, stale epoch, absent payload commits, undeclared
paths, missing evidence, dirty T4, and already accepted checkpoints. A validator
that checks only constants or array presence is not an intake gate.

T4-01 must provide these authoritative files and commands before any intake:

```text
_docs/ralph/prototype-integration/producer-handoff.schema.json
_docs/ralph/prototype-integration/epochs/epoch-N.json
scripts/validate-prototype-intake.cjs
scripts/validate-prototype-intake.test.cjs
```

```powershell
node --test scripts/validate-prototype-intake.test.cjs
$expectedT4 = (git rev-parse HEAD).Trim()
$validationClone = Join-Path $env:TEMP ("virtengine-prototype-intake-" + [guid]::NewGuid())
git clone --no-local --branch ve/prototype-integration https://github.com/virtengine/virtengine.git $validationClone
if ((git -C $validationClone rev-parse HEAD).Trim() -ne $expectedT4) { throw "validation clone is not the exact T4 SHA" }
node scripts/validate-prototype-intake.cjs --repo $validationClone --epoch 1 --tag checkpoint/prototype-t2/t2-03b
```

Both commands are mandatory. T4 performs no producer merge until the validator
and its negative tests pass on the exact T4 SHA. The validation clone must be
new or independently verified clean; do not relax the validator's worktree
cleanliness check to accommodate unrelated local files.

After `announcement_cutoff`, plan the frozen roster without mutating the epoch.
Select at most one intake-format annotated tag per producer explicitly; the
planner rejects early, late, lightweight, wrong-thread, duplicate, or unknown
selections and marks unselected producers frozen out:

```powershell
node scripts/plan-prototype-intake-freeze.test.cjs
node scripts/plan-prototype-intake-freeze.cjs --epoch 1 `
  --observation _docs/ralph/prototype-integration/epochs/epoch-1-tag-observation.json `
  --tag T1=checkpoint/prototype-t1/t1-09 `
  --tag T3=checkpoint/prototype-t3/t3-13a `
  > $env:TEMP\epoch-1-frozen-plan.json
```

Review the proposed epoch separately. The planner never writes the epoch file
and does not accept or merge any producer checkpoint. From a clean exact-SHA T4
clone, apply only the reviewed open-to-frozen transition, then inspect and
commit the epoch file:

```powershell
node scripts/apply-prototype-intake-freeze.test.cjs
node scripts/apply-prototype-intake-freeze.cjs --epoch 1 `
  --plan $env:TEMP\epoch-1-frozen-plan.json
git diff -- _docs/ralph/prototype-integration/epochs/epoch-1.json
```

The application command rejects dirty worktrees, pre-cutoff execution, changed
epoch metadata or roster order, and producer decisions other than announced or
frozen out. Acceptance remains a later, separately validated transition.

## Core RC Publication Preflight

T4-09A is diagnostic-only. It never creates or pushes a tag. Run it with the
exact candidate, intake epoch, and reserved checkpoint tag:

```powershell
node scripts/preflight-core-rc-publication.cjs --candidate <full-sha> --epoch 1 --tag checkpoint/prototype-integration/t4-09a --json
```

The command exits nonzero while any blocker remains and emits a deterministic
report conforming to `core-rc-publication-preflight.schema.json`. `--publish` is
intentionally unavailable. Publication requires a clean local/remote exact-SHA
boundary, terminal producer decisions, accepted ledger/tag/payload
correspondence, and revalidated accepted tags. The current strict v0 manifest
must retain its schema-required false authority flags; publication readiness is
evaluated separately from that contract and remains blocked until a future
valid manifest reports a ready status with no prototype-success blockers.

Gate results must validate against the exact execution plan computed from the
declared base to the candidate. Their committed bytes must be SHA-256 bound in
both the manifest and integration ledger. Required CI evidence is accepted only
from the immutable workflow paths and job names, exact VirtEngine repository,
`ve/prototype-integration` branch, allowed event, successful run attempt and
job, exact candidate SHA, and matching required-gate artifact digest. The local
and remote publication tag must remain absent.

## Toolchains

Go checkpoints use the repository-pinned Go `1.25.8` executable and record
literal `go version` output. Node checkpoints use Node 20 or the repository
declared version and record `node --version`. When the shared terminal corrupts
input or PATH resolution is unreliable, use isolated tasks and absolute
executable paths.

## Claim Levels

- `partial`: useful implementation exists, but one or more stated exit criteria
  or required gates remain open;
- `green producer checkpoint`: all checkpoint acceptance criteria pass and the
  branch/tag/handoff/evidence are remotely discoverable;
- `integrated`: T4 accepted and validated the producer payload;
- `task complete`: all roadmap acceptance criteria and external/non-claim rules
  pass.

Local commits without remote publication are progress, but never intake-ready.