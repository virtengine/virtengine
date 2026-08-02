# T4-13A Generated Contract Inventory Evidence

Status: diagnostic-green, generation-blocked

The inventory fixes five required contract families: stage decisions,
uniqueness receipts, eligibility decisions, claim presentations, and fund
authorization. Current producer trees contain handwritten or consumer-side
precursors for four families, but no canonical protobuf source exists for any
family and stage decisions are absent as a standalone payload.

The validator reserves one canonical `all` generation window, restricts source
and target roots, requires an accepted producer tag/payload for every generated
family, verifies producer source blobs through the accepted T4 ledger, and
requires compatibility fixtures. `--require-ready` rejects the current state.

This checkpoint does not open the generation window or claim generated output.
Generation remains blocked until immutable producer checkpoints publish all
canonical source contracts and T4 accepts them after the epoch cutoff.