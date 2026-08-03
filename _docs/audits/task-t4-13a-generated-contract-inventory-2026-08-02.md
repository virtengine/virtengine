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
requires compatibility fixtures. A completed window must retain first/second
run exit codes, zero-drift status, and an evidence path/hash verified at the
exact generation source commit. Accepted producer tags are also re-read and
every proto source must be declared in that checkpoint's committed handoff
`files_changed`; unrelated historical proto files cannot satisfy readiness.
Every declared proto source, generated target, and compatibility fixture must
also resolve to an existing regular file under the checked repository root;
canonical-looking invented paths cannot satisfy generated readiness.
Generated targets must also match their root's generated artifact type: OpenAPI
JSON/YAML, binary/hash inventory artifacts, Go protobuf/gateway output, or
generated TypeScript. Existing handwritten files under a canonical target root
cannot satisfy readiness.
Per-contract blocker lists are sorted and unique, and the root blocker list
must exactly match first-use contract references. Duplicate, undeclared, or
stale generated-contract blockers fail validation.
`--require-ready` rejects the current state.

This checkpoint does not open the generation window or claim generated output.
Generation remains blocked until immutable producer checkpoints publish all
canonical source contracts and T4 accepts them after the epoch cutoff.