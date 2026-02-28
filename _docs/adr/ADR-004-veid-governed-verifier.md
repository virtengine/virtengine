# ADR-004: VEID Governed Verifier Architecture

## Status

Proposed

## Date

2026-04-10

## Context

VirtEngine's existing VEID implementation already contains deterministic pipeline-version and model-version concepts inside [`x/veid/types/pipeline_version.go`](../../x/veid/types/pipeline_version.go) and [`x/veid/types/model_version.go`](../../x/veid/types/model_version.go). Those types are useful, but they currently keep three concerns tightly coupled inside `x/veid`:

1. identity evidence capture and attestation
2. verifier version selection and activation
3. issuance economics tied to successful verification

That coupling creates two strategic problems:

- It makes the canonical verifier look founder- or implementation-controlled even if validators can already compare versions deterministically.
- It makes economic changes look like verifier changes, which increases governance risk and makes future upgrades harder to reason about.

VirtEngine needs a design in which validators still compute the same score at a given height, but the active verifier is selected by network governance rather than by a foundation key, a reference binary, or an operator allowlist.

## Decision

VirtEngine will separate VEID verifier governance from VEID verification execution by introducing two Cosmos modules:

- `x/veidregistry`
- `x/issuancepolicy`

`x/veid` remains the execution and evidence module. It continues to manage identity scopes, attestations, verification requests, score persistence, and downstream integration with MFA and marketplace controls.

`x/veidregistry` becomes the source of truth for which verifier version is canonical at any block height.

`x/issuancepolicy` becomes the source of truth for how successful VEID proofs affect token issuance, rate limits, and pause controls.

## Decision Details

### 1. Consensus is versioned by height

The network consensus rule is:

`active_verifier(height) = {spec_version, weights_sha256, test_vectors_sha256, activation_height}`

At any given height, all validators must execute the same active verifier using deterministic execution rules and produce the same integer output.

The protocol does **not** require verifier version `N+1` to produce the same output as verifier version `N`. It requires all validators to agree on the verifier that is active at the current height.

### 2. The canonical verifier is artifact-based, not binary-based

Consensus attaches to public artifacts and deterministic rules, not to a single implementation:

- formal verifier spec
- model and ensemble weights
- preprocessing rules
- thresholds and rounding rules
- conformance test vectors

Different implementations in Go, Rust, Python, C++, ONNX Runtime, TensorFlow Lite, or future proving systems are allowed if they pass the same conformance vectors and implement the same fixed-point semantics.

### 3. Governance controls activation, not a multisig

After genesis, the active verifier can only change through on-chain governance.

The registry must have:

- no admin bypass key
- no foundation multisig override
- no hidden emergency signer path

Emergency response is handled by governance-visible pause and upgrade mechanisms, not by silent off-chain intervention.

### 4. Issuance policy is independent from verifier selection

Verifier approval and token issuance are distinct domains:

- `x/veidregistry` answers: "Which verifier version is valid right now?"
- `x/issuancepolicy` answers: "What economic effect does a successful proof have right now?"

This split allows governance to change mint amounts, caps, or regional throttles without redefining the canonical verifier.

## Module Boundaries

### `x/veid`

Retains ownership of:

- encrypted identity scope submission
- evidence validation and attestation ingestion
- verification request lifecycle
- deterministic score computation against the currently active verifier
- account score state and tier transitions
- MFA and marketplace integration

`x/veid` must no longer be the authority for pipeline activation policy or issuance economics.

### `x/veidregistry`

Owns:

- verifier artifact registry
- proposal metadata for new verifier versions
- activation height scheduling
- validator readiness reporting
- active verifier pointer by height
- deprecation and retirement lifecycle
- governance-enforced pause of verifier activation if preconditions are not met

### `x/issuancepolicy`

Owns:

- mint-per-proof policy
- epoch and daily issuance caps
- issuance pause controls
- optional jurisdictional or class-based throttles
- treasury/founder/reward distribution policy for verification-triggered minting

## Governance Rules

### Standard verifier upgrade

Standard verifier upgrades require:

- on-chain proposal submission
- public artifact hashes and IPFS content references
- public changelog
- conformance vectors
- reproducible build metadata
- at least two independent passing implementations from different organizations
- voting period
- timelock
- deterministic activation height

Recommended baseline:

- 14-day voting period
- 21-day timelock
- activation only after readiness conditions are met

### Fast-track security upgrade

Fast-track upgrades are allowed only for public, documented verifier defects and require:

- public security report hash on-chain
- higher supermajority threshold
- short but non-zero timelock

Recommended baseline:

- 80% supermajority
- 48-hour timelock

### Fail-safe response

If the active verifier is found to be unsafe, the network should prefer:

- pausing VEID-triggered issuance in `x/issuancepolicy`

instead of:

- halting the chain
- silently replacing the verifier
- forcing validators to choose between known-bad behavior and slashing

## Required Invariants

The architecture adopts the following invariants:

1. There is exactly one active verifier at a given height.
2. The active verifier is identified by immutable public artifacts, not by a specific binary.
3. `x/veid` must reject verification results that do not match the active verifier version.
4. `x/issuancepolicy` must never approve issuance for a proof produced under a non-active verifier.
5. Governance can pause issuance without redefining verifier validity.
6. No privileged key can activate, replace, or unpause a verifier outside governance.

## Migration Strategy

VirtEngine should migrate incrementally rather than rewriting VEID in one pass.

### Phase 1

Keep the current execution flow in `x/veid`, but reclassify existing pipeline-version and model-version types as legacy-internal structures.

### Phase 2

Move activation authority and proposal state out of `x/veid` into `x/veidregistry`.

This includes the concepts currently represented in:

- [`x/veid/types/pipeline_version.go`](../../x/veid/types/pipeline_version.go)
- [`x/veid/types/model_version.go`](../../x/veid/types/model_version.go)
- [`x/veid/keeper/model_version.go`](../../x/veid/keeper/model_version.go)

### Phase 3

Add `x/issuancepolicy` and route all VEID-triggered mint decisions through it.

### Phase 4

Reduce `x/veid` to a consumer of:

- `veidregistry.GetActiveVerifier(height)`
- `issuancepolicy.AuthorizeMint(proofContext)`

## Consequences

### Positive

- Stronger decentralization story: the network, not a founder-controlled key, decides the canonical verifier.
- Cleaner protocol boundaries between identity verification and token economics.
- Better implementation freedom for independent validator clients.
- Safer emergency response through issuance pause instead of chain halt.

### Negative

- More module complexity and cross-module wiring.
- More governance surface area.
- Additional implementation work to migrate current versioning logic out of `x/veid`.

### Accepted Trade-off

VirtEngine accepts added module complexity in exchange for a verifier architecture that is deterministic for consensus, publicly auditable, and governance-controlled without a backdoor.
