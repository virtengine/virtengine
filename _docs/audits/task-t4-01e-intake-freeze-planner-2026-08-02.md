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
T4 and a committed strict acceptance artifact whose base/candidate SHAs match,
whose tags peel to the declared payloads, and whose accepted payloads cover all
contained producer history. The current `ve/prototype-integration-live` branch
fails because its acceptance summary is uncommitted and its base predates
canonical T4; no promotion or merge is claimed.

The negative suite rejects planning before cutoff, late tags, wrong-thread tags,
invalid targets, unknown producers, duplicate selections, unobserved tags, and
post-cutoff observations. The real epoch
planner currently exits nonzero because the authoritative UTC cutoff has not
elapsed. Planning does not accept or merge producer payloads.