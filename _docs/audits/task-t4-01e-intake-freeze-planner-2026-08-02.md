# T4-01E Intake Freeze Planner Evidence

Status: diagnostic-green, cutoff-blocked

The diagnostic planner accepts explicit producer tag selections and emits a
proposed frozen epoch without writing repository files. It verifies strict tag
grammar, producer ownership, annotated tag type, commit target shape, and tagger
time at or before the epoch cutoff. Unselected producers become frozen out.

Remote publication timing is not inferred from the producer-controlled tagger
date alone. The planner requires a committed pre-cutoff observation containing
each annotated tag object and peeled target; unobserved tags cannot enter the
frozen roster.

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
are strict option/value pairs; unknown, duplicate, or incomplete inputs fail. The
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