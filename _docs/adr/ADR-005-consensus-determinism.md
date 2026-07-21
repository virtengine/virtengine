# ADR-005: Deterministic Consensus Input and ABCI++ Admission v1

- **Status:** Accepted; activation pending the registered `v1.4.0` software upgrade
- **Decision date:** 2026-07-20
- **Decision version:** 1
- **Owners:** Protocol and validator engineering

## Context

VirtEngine previously installed unconditional no-op `PrepareProposal` and `ProcessProposal` handlers. Consensus-reachable VEID and settlement code also admitted validator-local time and externally backed callbacks. CometBFT 0.38 supplies signed `ExtendedCommitInfo` only to `PrepareProposal`; `ProcessProposal` receives `CommitInfo` without extension payloads or extension signatures. The application therefore carries the complete signed extended commit in a registered Cosmos SDK transaction instead of inventing a raw-byte envelope.

## Decision

### Consensus input boundary

Consensus decisions may depend only on:

1. committed KV state;
2. ordinary transactions decoded and canonically re-encoded by the configured Cosmos SDK transaction codec;
3. ABCI request fields supplied by CometBFT;
4. `ctx.BlockTime()`, `ctx.BlockHeight()`, consensus parameters, and deterministic integer/fixed-point arithmetic; and
5. cryptographic evidence whose signer, chain ID, height, round, validator address, power, and signature are verified through supported CometBFT/Cosmos SDK APIs.

Consensus code must not depend on local wall time, host locale/time zone, process speed, map iteration order for hashes or state ordering, floating-point state decisions, random bytes, host files/environment, DNS, HTTP, gRPC, sidecars, DEXs, payout providers, or endpoint availability.

### Activation and compatibility

The handlers are installed for every node binary. Strict admission is activated from block `H+1`, where `H` is the committed done height of the registered `v1.4.0` software upgrade.

- Before and at `H`, prepare preserves the historical transaction list and process accepts it. This allows the upgrade block to execute under the old policy.
- From `H+1`, strict admission is mandatory.
- An activation-store read failure rejects processing and makes a proposer return an empty proposal. The installed wrapper also converts proposal-handler errors, nil responses, and panics to an empty proposal so the current BaseApp fork cannot restore unvalidated request candidates.
- New networks must schedule and execute `v1.4.0`; absence of its persisted done marker intentionally leaves compatibility mode active rather than inferring activation from local configuration.
- Rollback to a pre-`v1.4.0` binary after `H` is unsupported. Operators may roll back before finalizing `H`; after finalization they must use a binary that implements this ADR or coordinate a governance-approved successor upgrade.

The `v1.4.0` handler sets `VoteExtensionsEnableHeight` to `H+1`, runs the VEID v1-to-v2 module migration, and leaves existing records untouched. Its persisted x/upgrade done marker remains the strict-admission activation record.

### Proposal invariants and limits

Strict `PrepareProposal` first bounds and canonicalizes the CometBFT candidate window, then delegates ordering and byte selection to the SDK default proposal handler. It runs the SDK prepare ante verification before returning selected transactions. Strict `ProcessProposal` applies the same canonical byte validator and the SDK process ante verification.

The v1 limits are:

| Limit | Value |
| --- | ---: |
| Candidate transactions inspected by prepare | 10,000 |
| Transactions admitted per block | 5,000 |
| Bytes per ordinary transaction | 1 MiB |
| Policy proposal-byte ceiling | 16 MiB, further reduced by the ABCI/consensus limit |
| Gas | Committed consensus `Block.MaxGas` when positive |

Transactions are identified inside a proposal by SHA-256 of their canonical SDK bytes. Duplicate canonical transaction identifiers are forbidden. Non-canonical protobuf encodings are forbidden even when a permissive decoder could parse them. Proposal-byte limits use CometBFT's protobuf transaction-list wire size, including per-item framing, rather than only raw payload lengths. Prepare omits malformed, duplicate, non-canonical, over-limit, and ante-invalid candidates. Process rejects the whole proposal for any such item.

Candidate order is preserved because Cosmos account sequence semantics are order-sensitive. The handler does not sort ordinary transactions or decode them as an ad-hoc envelope.

### Stable rejection codes

| Code | Meaning |
| --- | --- |
| `VE-CONS-0001` | Empty transaction bytes |
| `VE-CONS-0002` | SDK decoder/encoder rejected transaction |
| `VE-CONS-0003` | Bytes differ from canonical SDK re-encoding |
| `VE-CONS-0004` | Duplicate canonical transaction identifier |
| `VE-CONS-0005` | Per-transaction byte limit exceeded |
| `VE-CONS-0006` | Proposal byte limit exceeded |
| `VE-CONS-0007` | Transaction count/work limit exceeded |
| `VE-CONS-0008` | Consensus gas limit exceeded or overflowed |
| `VE-CONS-0009` | SDK ante verification rejected transaction |
| `VE-CONS-0010` | Persisted activation state unavailable |
| `VE-CONS-0011` | Missing, misplaced, duplicated, malformed, or tampered system transaction |

ABCI `ProcessProposal` exposes only accept/reject status. Codes are therefore used by the shared validator, tests, and observational logs; free-form diagnostic text is not consensus input.

### System transaction and vote-extension carrier rule

Carrier version 1 is active from `H+1`:

- `ExtendVote` emits a deterministic protobuf bundle bound to carrier version, chain ID, height, block hash, active pipeline version, OCI runtime digest, model-manifest digest, and at most ten precomputed results. Results are strictly ordered and bind request, account, score, status, model, full SHA-256 input hash, canonical reason codes, and result hash.
- `VerifyVoteExtension` rejects absent, malformed, non-canonical, oversized, wrong-chain, wrong-height, wrong-block, wrong-runtime, wrong-model, duplicate, unordered, or tampered bundles. It performs no inference, decryption, filesystem access, or network call.
- CometBFT signs each accepted extension. `PrepareProposal` validates `LocalLastCommit` with `baseapp.ValidateVoteExtensions` and the staking keeper, recomputes a strict `(2/3)+1` voting-power aggregate, and encodes the complete signed `ExtendedCommitInfo` plus aggregate in `MsgSubmitConsensusVerification`.
- The message is encoded as an ordinary canonical Cosmos SDK `TxRaw`, has no ordinary account signer, fixed zero fee and bounded gas, and is authenticated by the carried Comet signatures. The SDK decoder rejects malformed envelopes and the first ante boundary rejects decoded user/mempool injection, recheck, and simulation while preserving the SDK's ordinary CheckTx response/indexing behavior. This is the supported SDK mechanism used instead of a synthetic raw prefix.
- `ProcessProposal` requires exactly one system transaction at index zero, compares the carried addresses/power/flags to `ProposedLastCommit`, checks each declared power against the committed staking last-validator-power index, revalidates all extension signatures against committed staking keys, recomputes the aggregate, and rejects any mismatch. Ordinary SDK transactions begin at index one and retain all bounded canonical/ante checks.
- Final execution routes the same registered message through BaseApp. PreBlock independently revalidates the exact index-zero bytes, decided commit, signatures, staking powers, and aggregate (including block-sync/replay paths) and places an unexported authorization marker in the immutable SDK context. Ante and the VEID message server require that marker to match `ctx.TxBytes()`. The message server then recomputes the carried signed evidence, applies all results atomically from a cache context, persists the canonical aggregate/result state, and records a consumed-height digest. Duplicate, direct, or replayed consumption is rejected.
- At the initial vote-extension enable height no prior extension signatures exist. The index-zero transaction still carries and validates the previous commit but has an empty aggregate. Later heights require valid non-empty extension payloads from every commit voter, even when the result list itself is empty.

The old JSON helper remains source-compatible test/legacy code and is never used by consensus handlers. The active result list remains empty unless validator-controlled pre-consensus work has committed an active Task 85D pipeline/runtime/model profile and stored bounded result records. Task 84A securely transports and executes those records; Task 85D still owns authenticated sidecar receipt production and must not reintroduce inference or network I/O into callbacks.

### Time, work, and telemetry

VEID score and identity-scope activity decisions use explicit `At` variants with block time. Host-speed duration fields written by verification/scoring code are fixed to zero; off-chain metrics may measure runtime but may not affect stored values, quotas, or selection. Work is bounded by counts, bytes, and consensus gas, never elapsed wall time.

### External pricing, DEX, and payout behavior

The application no longer installs external Band/Chainlink price adapters in the settlement keeper. The compatibility `SetPriceFeed` method discards non-Cosmos sources. DEX and off-ramp setters discard adapters. Consensus may commit a fiat-conversion request, but payout remains pending and no quote, swap, payout, status, recovery, or cancellation callback is invoked. Reconciliation fails with the stable settlement error `ErrExternalIOForbidden`.

Committed x/oracle values and governance manual overrides remain valid deterministic price inputs. A future payout task must add authenticated on-chain observation/result messages before external execution can advance consensus state.

## Static enforcement

`scripts/consensusdeterminism` scans proposal, ante, module begin/end, message-server, and all production keeper files in the settlement, VEID, HPC, and audit modules for wall clocks, randomness, filesystem/environment reads, network calls, external callback methods, map iteration, duration-to-float conversion, and floating-point branch decisions. The allowlist is exact by rule, path, and function and carries a rationale. CI executes the checker in the Go lint job.

The settlement keeper contains no production invocation of DEX/off-ramp callback methods. Source-compatible adapter interfaces and setters remain, but setters discard inputs and reconciliation/execution return `ErrExternalIOForbidden`. Remaining allowlist entries cover audited deterministic key sorting/copies and inactive off-chain ML compatibility helpers; carrier version 1 does not execute those helpers.

## Consequences

- Malformed or ante-invalid proposals are rejected after activation instead of being accepted unconditionally.
- Existing networks require a coordinated `v1.4.0` upgrade and cannot activate from validator-local configuration.
- VEID vote extensions and their signed commit are carried through one canonical SDK-native index-zero transaction; actual inference receipt production remains a Task 85D predecessor boundary.
- Fiat conversion requests can be observed off-chain, but no conversion can complete on-chain until a separately authenticated observation schema is implemented.
- Proposal work and allocations are bounded and repeatable across validators.
