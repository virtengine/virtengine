# T4-01E Intake Freeze Planner Evidence

Status: diagnostic-green, cutoff-blocked

The diagnostic planner accepts explicit producer tag selections and emits a
proposed frozen epoch without writing repository files. It verifies strict tag
grammar, producer ownership, annotated tag type, commit target shape, and tagger
time at or before the epoch cutoff. Unselected producers become frozen out.

Remote publication timing is not inferred from the producer-controlled tagger
date alone. The planner requires a committed pre-cutoff observation containing
each annotated tag object and peeled target; unobserved tags cannot enter the
frozen roster. Observer CLI options are single-assignment and reject missing or
option-looking values, so an ambiguous epoch, remote, or repository selection
cannot produce observation evidence. Freeze-planner scalar options enforce the
same rule while retaining repeated `--tag` selections for distinct producers.
Freeze application also requires every reviewed boundary option exactly once
and rejects missing or option-looking values before acquiring its lease. Lease
inspection and compare-and-claim recovery enforce the same single-assignment
rule for the reviewed epoch, HEAD, plan digest, and retained lease digest. New
epoch opening likewise rejects duplicate, missing, or option-looking reviewed
HEAD and UTC window arguments before checking the published base boundary. The
producer intake validator applies the same rule to epoch, tag, repository, and
remote arguments before fetching or evaluating a checkpoint.

Exact-SHA manifest generation also rejects duplicate, missing, or option-looking
source, tooling-source, output, and check arguments. Checked-manifest path
identity resolves filesystem aliases and compares case-insensitively on Windows,
so alternate spelling cannot bypass the clean-worktree guard.

The diagnostic core-RC publication preflight makes candidate, epoch, tag,
repository, remote, and JSON mode single-assignment. Duplicate, missing, or
option-looking values fail before publication evidence is evaluated.

Integration candidate and canonical inputs must be exact lowercase commit SHAs
or `origin/*` branches whose tracking commits equal the live remote head.
Mutable local names such as `HEAD`, local branches, and abbreviated SHAs fail
before acceptance evidence is read.

Manifest handoff test records may omit counts and remain explicitly partial,
but any declared count must be a positive integer. Zero-count records cannot
produce complete test evidence. The manifest schema mirrors runtime semantics:
complete evidence has zero uncounted records, while partial evidence retains at
least one uncounted record.

Manifest test evidence must contain at least one unique literal command.
Duplicate records and whitespace-padded command identities are rejected before
record and declared-test counts are projected.

Tool-version evidence also requires nonempty literal tool names and version
strings with no surrounding whitespace in both runtime and schema validation.

Complete manifest test evidence carries no partial-evidence blocker. The
`test-evidence-partial` blocker is emitted only when one or more passing records
omit test counts, and schema validation enforces that status/blocker coupling.

Root manifest blockers exactly equal the blocker IDs referenced by artifact
groups, blocked sections, external dependencies, and AI non-certification.
Cleared artifact coverage cannot leave a stale blocker declaration.

Rejected producer evidence has exact thread, checkpoint, tip, and literal
reason fields. Runtime and schema validation reject malformed records and exact
duplicates, while the manifest remains byte-bound to the integration ledger.

Control artifacts reject duplicate records in schema validation and duplicate
IDs or paths at runtime before source-byte hashes and statuses are accepted.

Schema-only validation also rejects exact duplicates in tooling artifacts,
toolchains, artifact groups, external dependencies, and root blockers, matching
their stronger runtime identity and inventory checks.

Root blocker IDs are canonical lowercase kebab-case and descriptions are
nonempty literal strings without surrounding whitespace. Runtime and schema
validation reject malformed or duplicate blocker declarations.

External dependencies use canonical lowercase IDs, remain explicitly
unavailable, and reference canonical existing blockers. Runtime and schema
validation reject malformed or duplicate dependency declarations.

Tooling artifacts use unique repository-relative literal paths. Absolute,
parent-traversing, backslash, padded, or duplicate paths fail before Git blob
and SHA-256 provenance is accepted.

The planner hashes the selected observation, requires the checked core-RC
manifest to bind that exact path/digest, verifies identical bytes at the
manifest source commit, and requires that source to be an ancestor of current
T4. Local post-manifest observation edits therefore cannot affect the roster.

Aggregate integration validation also inspects every registered remote producer
branch from the epoch base. Any producer commit reachable from T4 must be
covered by an accepted same-thread payload that is itself reachable from T4;
missing producer refs or unaccepted out-of-band merges fail closed.

Candidate-branch promotion additionally requires descent from current canonical
T4 and a committed strict acceptance artifact whose `base_sha` matches canonical
T4 and whose `candidate_sha` names the validated implementation parent. Candidate
evidence is read from the exact candidate head resolved at preflight start, not
from the mutable candidate ref. Canonical or candidate inputs under `origin/*`
must equal the current exact remote branch head before resolution. CLI options
are strict option/value pairs; unknown, duplicate, or incomplete inputs fail.
Required gates run unit coverage and the published live-candidate policy
separately; the latter currently fails on its informal acceptance schema and is
projected as `integration-candidate-preflight-blocked`. Test results must execute
every discovered case, while policy results must report zero test counters.
Result evidence includes schema-validated command kind and must match the plan. The
reported tool set must exactly match pinned tools by count, unique name, version,
availability, and field shape. Matrix commands, tools, dependencies, blockers,
categories, and path allowances reject duplicate declarations. The
dependency schema is exact, and each unavailable dependency requires its
canonical `dependency-<id>` blocker with no stale dependency blockers retained.
Categories marked `ready` or `complete` cannot retain any blocker. The
matrix status is derived from category states, completion claims are complete-only,
and root blockers exactly equal category blocker usage. Execution-plan categories,
allowances, commands, and pinned tools reject duplicate records. Plans retain
category dependencies and blockers, and execution rechecks both independently
of projected status. Result envelopes reject duplicate records in schema and
runtime validation and bind a deterministic digest of the complete execution
plan using recursively key-sorted JSON while preserving array order. Passing
result evidence is rejected unless that exact plan is execution-ready. Canonical
serialization rejects sparse arrays, non-plain objects, unsupported values,
non-finite or unsafe numbers, and negative zero. Runner CLI options are
single-assignment and reject missing or option-looking values. Required-gate
planning rejects identical base/head revisions and ranges with no changed paths,
so an empty result envelope cannot represent a no-op checkpoint. The
acceptance commit must be the implementation's direct, single-parent child and
change only the acceptance artifact. Accepted annotated tags must be exact tag
objects published by `origin`; local-only or stale same-name tags are rejected.
Those tags target committed handoffs that bind the declared payloads, and those
payloads cover all contained producer history. Containment scans require every
local producer tracking ref to equal the current exact `origin` branch head, so
stale refs cannot omit unaccepted commits. The current `ve/prototype-integration-live` branch
fails because its acceptance summary is uncommitted and its base predates
canonical T4; no promotion or merge is claimed.

The negative suite rejects planning before cutoff, late tags, wrong-thread tags,
invalid targets, unknown producers, duplicate selections, unobserved tags, and
post-cutoff observations. The real epoch
planner currently exits nonzero because the authoritative UTC cutoff has not
elapsed. Planning does not accept or merge producer payloads.