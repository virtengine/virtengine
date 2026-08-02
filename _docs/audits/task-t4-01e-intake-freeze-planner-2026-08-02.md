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

The negative suite rejects planning before cutoff, late tags, wrong-thread tags,
invalid targets, unknown producers, duplicate selections, unobserved tags, and
post-cutoff observations. The real epoch
planner currently exits nonzero because the authoritative UTC cutoff has not
elapsed. Planning does not accept or merge producer payloads.