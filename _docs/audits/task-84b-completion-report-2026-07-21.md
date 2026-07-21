# Task 84B Completion Report — 2026-07-21

## Status

**Implemented and validated to the best achievable local standard.** Full acceptance is not claimed because no separately running multi-validator chain with a live provider-daemon process, governed key registration transactions, real customer wallet acknowledgment, and restart/retry accounting observation was available during this work session.

No commit, push, stage, unstage, reset, checkout, or revert operation was performed. The pre-existing staged Task 84A/86A and unrelated change set remained staged; Task 84B changes remain unstaged/untracked.

## Implemented

- Version-1 fixed-width/length-prefixed canonical provider usage and customer acknowledgment payloads.
- Full chain/provider/customer/order/lease/allocation/period/raw-metric/price/formula/model/sequence/replay/key-epoch/expiry binding.
- Exact integer metric-to-unit formulas, canonical decimal validation, and overflow/skew/lifetime/period bounds.
- Ed25519 and canonical low-S Cosmos secp256k1 verification; X25519, multisig, unsupported keys, malformed, and high-S forms fail closed.
- Governed `x/provider` key epoch history, proof-based one-step rotation, bounded overlap, revocation, genesis, migration, queries, and transactions.
- `x/auth` account-public-key customer acknowledgment verification over the exact stored usage digest.
- Collision-safe sequence, nonce, idempotency, period, and acknowledgment replay indexes plus invariant checks.
- Exact retry returns the original usage ID, emits no duplicate event, and performs no duplicate billing/reward/escrow/settlement transition.
- Cached-context atomicity for usage, acknowledgments, settlement, and HPC settlement orchestration.
- Legacy migration to `legacy_unverified`; legacy usage remains queryable but cannot newly trigger financial effects.
- Provider daemon governed key-state resolution, KeyManager identity matching, durable proof/sequence allocation, canonical signing, and durable signed transaction submission.
- HPC keepers no longer derive signatures from calculation hashes/record IDs. Unsigned accounting is pending/unbillable until off-chain authenticated usage IDs exist.
- Additive provider/settlement protobufs, generated Go/gateway/descriptor/OpenAPI/TypeScript outputs, and binary/field-number compatibility fixtures.

## Versions

- Software upgrade: `v1.5.0`
- Settlement module: consensus version `2` (`1 -> 2` marks legacy records and activates authenticated metering)
- Provider module: consensus version `4` (`3 -> 4` backfills immutable signing-key epochs)
- Canonical usage signature version: `1`
- Customer acknowledgment signature version: `1`
- Provider key rotation proof version: `1`
- Pricing/formula/model versions available: `1/1/1`

Rollback below `v1.5.0` after activation is unsupported because old binaries do not enforce the replay and legacy financial gates.

## Golden and compatibility evidence

- Canonical binary golden: `x/settlement/types/metering_auth_test.go`
- Legacy protobuf binaries and additive field-number guards: `tests/compatibility/task84b_wire_compatibility_test.go`
- Provider key migration and lifecycle: `x/provider/keeper/signing_key_epoch_test.go`
- Replay, tamper, acknowledgment, migration, and exact-once behavior: `x/settlement/keeper/usage_authentication_test.go`

## Commands and results

Passed:

- `go test ./x/provider/... ./x/settlement/... ./x/hpc/... ./pkg/provider_daemon ./tests/compatibility ./upgrades/software/v1.5.0 ./app/types -count=1`
- `go test ./x/settlement/types -run '^$' -fuzz '^FuzzCanonicalUsageSignBytes$' -fuzztime=5s` — 108,068 executions in the final recorded run.
- WSL race: all `x/settlement`, `x/provider`, and `x/hpc` packages; targeted Task 84B provider daemon tests.
- `go vet` on all affected packages.
- `golangci-lint run` on all affected packages — `0 issues`.
- `go build ./cmd/virtengine ./cmd/provider-daemon ./cmd/hpc-node-agent`.
- `go run ./scripts/consensusdeterminism -root .` — `0` unapproved, `25` narrowly allowlisted findings.
- Pinned `./scripts/proto-generate-wsl.sh all` twice — second-pass direct-output hash delta `0`.
- Pinned `./scripts/proto-generate-wsl.sh verify` — descriptor/inventory tests `3/3` passed.
- `npm --prefix sdk/ts run build` in pinned WSL environment.
- `NODE_OPTIONS=--max-old-space-size=4096 npm --prefix sdk/ts test -- --runInBand` with pinned Buf on `PATH` — `36/36` suites, `1101/1101` tests, `16/16` snapshots.
- Git for Windows Bash `./scripts/verify-modules.sh` — passed all seven modules and policy checks.
- `node scripts/validate-agents-docs.mjs` — passed.
- `git diff --check` — passed.
- Real local-node settlement CLI integration — arbitrary noncanonical proof rejected after fresh-genesis activation.

Environment-qualified failures:

- Full WSL `go test -race ./pkg/provider_daemon` exposed pre-existing races in compromise-alert tests (`compromise.go` versus `keymanagement_test.go`), outside Task 84B files. Targeted Task 84B provider tests passed under race.
- Repository preflight failed because it consumes the protected staged Task 86A deletion/module/tag set: it passes deleted `gateway_stub.go` paths to `gofmt` and treats `sdk/generation` and opt-in gateway/consensus packages as root packages. Targeted vet/tests/build/lint/module/proto/docs/diff gates passed.
- Windows TypeScript execution used Linux-installed native `esbuild`; the same build passed in pinned WSL. The first WSL test invocation lacked Buf on `PATH` and exhausted the default heap; rerun with pinned Buf and a 4 GiB heap passed all tests.
- Docker-backed `scripts/verify-proto-generation.sh` was not used; the pinned WSL equivalent was executed twice and verified.

## Process-boundary evidence

Available and executed:

- Provider daemon KeyManager signs canonical usage only after governed signing-state resolution.
- Durable transaction queue stores proof allocation, survives restart, and rebuilds identical sequence/nonce/idempotency/signature bytes.
- Generated real Cosmos transaction construction and broadcast-client path tests execute.
- A real local node rejects arbitrary legacy signature bytes after Task 84B activation.

Not available; therefore not claimed:

- A separately running real provider daemon registering its key via chain transactions, signing and broadcasting to a live multi-validator network.
- A real customer wallet process creating the detached acknowledgment against the persisted digest.
- A live restart/retry observation proving exactly one billing/reward/event transition across multiple validators.
- Production command composition still constructs the raw usage meter without a metric collector/chain recorder and does not instantiate `ChainUsageSubmitter`; the canonical durable producer component is implemented/tested but must be wired when the Task 85A broadcaster composition is completed.
- Production file/HSM/non-custodial KeyManager persistence remains Task 85C scope; local positive signing evidence used memory custody.

## Acceptance assessment

Achieved locally:

- Arbitrary signatures no longer satisfy authenticated usage.
- Provider and customer proofs are independently verified and domain-separated.
- Exact retries are idempotent; conflicting replay keys fail.
- Rotation overlap/revocation, legacy behavior, migration, wire compatibility, and negative/tamper cases are tested.
- HPC pseudo-signatures are removed and unsigned output fails closed.
- Generated contracts and SDK compile/test reproducibly.

Not achieved due to external/environment limits:

- Required live multi-validator, real-process provider/customer/restart/billing evidence.
- Full provider-daemon package race cleanliness because of unrelated pre-existing compromise-alert races.
- Clean aggregate repository preflight in the protected mixed staged worktree.
- Positive production provider-daemon composition until the existing raw meter is connected to the durable canonical submitter with complete customer/allocation lineage.

Task 84B should not be represented as process-boundary complete until those remaining gates are run on an isolated descendant revision with the staged prerequisite work committed.
