# Task 84B Completion Report — 2026-07-21

## Status

**COMPLETED.** Task 84B is implemented and executable local acceptance evidence passes. This completion pass began at the requested base commit `e2d78c01a0b653baabc9d27d0ca26d8305dd54c6`; the concurrently amended current checkout is `99973c3af6c084a915e2913016d5d08899dcdaa6`. No commit, push, stage, unstage, reset, checkout, revert, or amend operation was performed by this agent.

The opt-in harness runs four independently constructed VirtEngine applications behind four CometBFT validators. Provider submission/restart and customer acknowledgment run in separate test-binary OS subprocesses and communicate with the parent only through `0600` temporary JSON files plus public Comet RPC/gRPC. No keeper pointer crosses into a helper process.

## Protocol and activation

- Software upgrade: `v1.5.0`; settlement consensus version `2`; provider consensus version `4`.
- Provider usage, customer acknowledgment, and provider-key rotation signature versions are `1`.
- Fresh-chain activation: settlement default genesis sets `usage_authentication_active=true`; the process test reads the activation marker from every validator's committed state before submitting usage.
- Existing-chain activation: registered `v1.5.0` provider `3 -> 4` and settlement `1 -> 2` migrations.
- Rollback below `v1.5.0` after activation is unsupported.

## Executable process and consensus evidence

Command:

`$env:VE_RUN_TASK84B_PROCESS='1'; go test -tags='e2e.integration' ./tests/integration/settlement -run '^TestTask84BFourValidatorProviderRestartCustomerAckExactlyOnce$' -count=1 -v; Remove-Item Env:VE_RUN_TASK84B_PROCESS`

Final result: **PASS** in `10.639s`.

`TASK84B_RESULT validators=4 activation=fresh_genesis height=20 app_hash=4C39E284A499E4EDC5BC363F541D8E2085E228E9391C75D9CEDEB951EC972E82 validator_heights=[20 20 20 20] arbitrary_signature=rejected queue_restart=verified local_duplicate=ErrDuplicateReport chain_exact_retry=success conflict_replay=rejected usage=usage-1784623092-1 usage_events=1 acknowledgment=1 settlements=1 usage_rewards=1 provider_balance=1000000000000000094uve escrow_balance=9900uve`

The app hash comes from `BlockResults(height=20)` through every validator's local Comet RPC client. It is not derived from a historical-height SDK context over latest state.

### Topology and isolation

- Four validator nodes and four application instances on one Comet consensus network.
- Provider process 1 owns an Ed25519 `KeyManager`, canonical proof allocation, transaction signing, and HTTP RPC broadcast.
- Provider process 2 recreates `ChainUsageSubmitter` from the same durable queue and proves local suppression returns `ErrDuplicateReport` without broadcast.
- Provider process 3 retains durable proof allocation but clears only local queue items, then broadcasts the exact original proof. Chain replay indexes return idempotent success without another `usage_recorded` event.
- Customer process owns the secp256k1 wallet key, signs the exact stored digest under the customer domain, builds a Cosmos transaction, and broadcasts over HTTP RPC.
- Temporary helper key files are test-only `0600` files; on Windows their effective protection also depends on the host temporary-directory ACL. Production file/HSM custody is not inferred.

### Negative and exactly-once proof

- A non-empty arbitrary 64-byte provider signature is rejected before the positive flow; usage/event counts remain zero.
- A correctly signed different payload reusing the accepted replay tuple is rejected; usage remains one and no event is emitted.
- Exact retry semantics are distinguished: same durable queue returns producer `ErrDuplicateReport`; a fresh local item with retained proof allocation reaches the chain and succeeds idempotently.
- Exactly one usage record and one committed `usage_recorded` event exist after all retries.
- Customer proof is independently signed and verified against the x/auth key and exact stored digest.
- Exactly one settlement and one usage reward exist. Reward metadata references the settlement and its recipient references the usage.
- A second settlement fails; settlement/reward/usage/event counts and provider/escrow balances remain unchanged.
- Stable named ABCI events are now emitted for usage, order settlement, and usage rewards. This fixes a concrete bug where handwritten unregistered typed-event structs were silently dropped.

## Implementation and compatibility

- Versioned length-prefixed canonical payloads bind full chain, signer, economic lineage, raw metrics, pricing, replay, key epoch, and expiry state.
- Ed25519 and canonical low-S secp256k1 verify; malformed, high-S, X25519, multisig, and unsupported forms fail closed.
- Provider key history, proof-based rotation, overlap, revocation, genesis, migration, query, and transaction behavior are enforced.
- Replay indexes and cached-context mutation make usage, acknowledgment, settlement, reward, and event behavior atomic and exact-once.
- Authenticated usage proof/derived-state immutability uses deterministic sorted metadata-key comparison.
- Legacy usage stays queryable as `legacy_unverified` but cannot newly trigger value-bearing effects.
- HPC pseudo-signatures are removed; unsigned accounting stays pending/unbillable.
- Provider genesis rejects nil state, duplicate/orphan records, duplicate/gapped epochs or IDs, broken lineage, activation regression, absent current epoch, and a non-highest current epoch.
- No schema or wire-number change was made in this pass; golden/additive compatibility fixtures remain valid.

## Commands and results

Passed:

- `go test ./x/provider/... ./x/settlement/... ./x/hpc/... ./pkg/provider_daemon ./tests/compatibility ./upgrades/software/v1.5.0 ./app/types -count=1`
- Opt-in four-validator process test above.
- `go test ./x/settlement/types -run '^$' -fuzz '^FuzzCanonicalUsageSignBytes$' -fuzztime=5s` — `123,173` executions.
- `go vet` across all affected packages and the build-tagged integration package.
- Delta `golangci-lint` for affected normal and `e2e.integration` packages — `0 issues`.
- `go build ./cmd/virtengine ./cmd/provider-daemon ./cmd/hpc-node-agent`.
- Focused WSL race suites for provider, settlement, HPC, and Task 84B provider-daemon paths.
- `go run ./scripts/consensusdeterminism -root .` — `0` unapproved, `25` narrowly allowlisted.
- Git for Windows Bash `./scripts/verify-modules.sh` — seven modules and policy passed.
- `node scripts/validate-agents-docs.mjs` — 9 files passed.
- `git diff --check`.
- `pwsh -NoProfile -File scripts/agent-preflight.ps1` after fixing its Windows test-package exclusion/argument handling.

Environment-qualified observations:

- Broad WSL race passed core packages but the settlement CLI integration subpackage timed out querying a transaction under race instrumentation. That CLI package passes normally; the focused race command excluding the timing-sensitive CLI harness passed every Task 84B core package.
- The Windows harness retains its isolated CometBFT network directory after shutdown. Removing it immediately exposed an upstream peer-routine/closed-LevelDB teardown race after all protocol assertions had passed. Provider/customer helper directories still use test-owned temporary paths; release environments must manage node data through normal process shutdown and lifecycle tooling.

Previously recorded prerequisite evidence in the `e2d78c01`/`99973c3a` amended Task 84A/86A/84B base remains: pinned protobuf generation twice with zero second-pass delta; generation verify `3/3`; TypeScript SDK `36/36` suites, `1101/1101` tests, and `16/16` snapshots.

## Acceptance matrix

| Requirement | Result | Evidence |
| --- | --- | --- |
| Arbitrary bytes rejected | PASS | Live signed transaction rejected; zero usage/events |
| Provider proof independently verified | PASS | Governed epoch, canonical Ed25519 proof, provider subprocess |
| Customer proof independently verified | PASS | Stored digest, x/auth secp256k1 key, customer subprocess |
| Exact retry semantics | PASS | Local `ErrDuplicateReport`; chain exact retry succeeds idempotently |
| Conflicting replay rejected | PASS | Different correctly signed payload under accepted tuple fails |
| No duplicate financial/event state | PASS | Counts stay `1/1/1/1`; balances unchanged |
| Restart durability | PASS | Recreated submitter loads queue and proof allocation |
| Four-validator convergence | PASS | Height `20`, exact `4C39...2E82` hash from all RPC clients |
| Fresh genesis active | PASS | Marker read on every validator before negative flow |
| Upgrade/rotation/legacy compatibility | PASS | Targeted upgrade/provider/settlement/compatibility suites |

## Remaining release-only evidence

Task 84B has no remaining local engineering acceptance blocker. Later tasks own:

- production raw-collector composition into `ChainUsageSubmitter` (85A);
- production file/HSM/non-custodial custody, restore, rotation, and HA fencing (85C);
- independently deployed validator hosts rather than four in-process nodes;
- long-duration testnet soak, crash-at-every-write fault injection, and exact-image evidence (88B/88D);
- production billing, payout, DEX/off-ramp, and external reconciliation certification.

These release-only items do not reopen Task 84B's authenticated metering, replay safety, exact-once state/event/financial behavior, or local process/multi-validator completion.
