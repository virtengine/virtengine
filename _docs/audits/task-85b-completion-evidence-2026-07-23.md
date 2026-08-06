# Task 85B Completion and Evidence Report

**Date:** 2026-07-23
**Status:** `engineering_complete_external_blocked`

Task 85B is complete at the deterministic local engineering/fixture boundary allowed by the plan's `engineering_complete_external_blocked` state. It is not testnet/sandbox-certified, production-certified or production-enabled. No real Osmosis testnet execution or payout-provider sandbox execution is claimed. The mandatory Claim 4 floor remains externally blocked because no external network/pool/liquidity/signing access, production custody backend, payout-provider contract or credentials, certified oracle/governance bundle, approved payout corridor, or independently reconciled real execution evidence has been supplied.

## Delivered engineering scope

- The production DEX factory now selects `RealOsmosisAdapter`, not the one-to-one placeholder. Its only implemented production math profile is Osmosis `gamm-v1beta1-cp-equal-weight-v1` for two-asset, equal-weight constant-product `osmosis.gamm.v1beta1.Pool` pools. Provider-daemon construction requires an authenticated pool response bound to one height, block hash and node identity; an unheighted LCD response is not executable.
- The Osmosis path pins chain/profile identity, pool allowlists, exact denoms and decimals, observation age/height, liquidity/reserve/amount limits, hop bounds, price impact, oracle deviation, quote expiry, minimum output, canonical execution payloads, target-chain `TxRaw` verification, confirmation height/block identity, finality and actual output.
- Uniswap V2 and Curve implementations remain explicit test-only placeholders; production factory selection rejects them.
- The off-ramp package now provides a strict, provider-neutral JSON HTTP adapter with exact API/profile versions, fixed-origin/path allowlisting, bounded strict JSON, HTTPS/transport checks, secret references, idempotency and correlation binding, independent compliance authorization, tokenized beneficiary references, durable production daily limits, status reconciliation and ambiguous-result classification.
- HMAC-SHA256 webhook verification pins API/key versions, authenticates timestamp plus raw body before parsing, binds callbacks to an initiated payout, detects exact and conflicting replays, and requires durable replay/binding/event stores in production.
- `x/settlement` no longer retains DEX or off-ramp clients. It commits the intent/value hold and accepts the generated, provider-signed `MsgRecordFiatConversionObservation` through six ordered happy-path stages, plus bounded failed/cancelled outcomes.
- Observation sequencing, idempotency, immutable accepted profile/compliance digests, quote/finality evidence, lineage indexes, daily quota accounting, linked payout state, financial-case holds and terminal invariants are consensus-validated. Current governance/profile/compliance/hold checks stop each new external side-effect boundary, while submission/finality observations safely reconcile an already-crossed irreversible boundary against immutable commitments.
- Authenticated payout completion atomically moves the native net payout from the settlement module account to the internal-only fiat-custody module account. It does not synthesize a provider-account or external bank/chain transfer. Platform fee, validator fee and holdback treasury entries use payout/type-stable IDs and are recorded exactly once in the same cached transaction.
- A financial case opened after swap submission cannot allocate the already-committed payout exposure, cancel it, or trigger a second provider/customer/platform allocation. It remains an incident/evidence hold requiring governed external reconciliation.
- The Task 85A mutation registry durably registers `settlement.record_fiat_conversion_observation`; observations are SDK-signed, persisted, broadcast and reconciled through the durable provider mutation queue.
- The provider-daemon orchestrator persists privacy-safe work state and payout/webhook repositories, uses a fenced local lease, recovers ambiguous swap/payout boundaries after restart, restores a volatile payout-provider binding only after an exact durable metadata/economic match, and never stores credentials, raw beneficiary data, raw execution payloads or signed transaction bytes.
- Expired accepted DEX quotes and expired, unsubmitted payout quotes can be replaced only before their respective irreversible submissions. Replacements append contiguous observations, retain profile/compliance/settlement lineage and quota/value holds, and cannot reuse the expired quote identity/digest.
- Upgrade `v1.8.0` disables fiat conversion by default, resets profile commitments to `engineering_complete_external_blocked`, deterministically preserves/quarantines legacy records, rebuilds indexes/accounting and checks invariants.
- Provider-daemon startup verifies versioned profile files but deliberately fails closed for production: reviewed external custody, payout partner, secret, destination, compliance and webhook implementations must be injected. Backend identifiers alone are not factories.

## Validation evidence

Recorded local evidence for the current Task 85B implementation:

| Gate | Result |
| --- | --- |
| Aggregate targeted Go tests covering DEX, off-ramp, provider-daemon/orchestrator, provider-daemon command startup, settlement protocol/consensus, `v1.8.0`, wire compatibility and upgrade registry | PASS |
| Relevant `go vet` scope | PASS |
| Relevant `golangci-lint` scope | PASS, 0 issues |
| Provider-daemon and virtengine command builds | PASS |
| Focused WSL race tests for DEX/off-ramp, provider fiat/orchestrator and settlement fiat/consensus paths | PASS |
| Pinned generation, followed by a second generation pass | PASS, zero second-pass drift |
| TypeScript SDK lint, build/export validation and unit/functional tests | PASS |
| Tagged settlement integration and `e2e.upgrade` compile/worker-registration checks | PASS |

Current generated artifact hashes were rechecked on 2026-07-23:

- Descriptor SHA-256: `1ebde455e84065fa2f53cfc90536b9aece3423c55aed1abdb323ce42a7b9fb9e`
- Inventory SHA-256: `654f503d6e84c881f62882f2a57cca3bf8077b2772bac3176c452dc6ebac50c4`
- OpenAPI SHA-256: `9e07efd4b158589d71907f350422fbba4ab9da36a3870a3065547ef8935d4d2d`

The Task 85B full preflight recomputes these three files and requires both this report and `_docs/ralph/progress.md` to contain the current values, so generated-artifact drift cannot leave Task 85B evidence silently stale.

## External blockers and release consequence

No checked-in configuration or evidence identifies a real executable testnet or production pool, funded liquidity, token route, oracle, custody signer, provider contract, sandbox/production credential, bank/custody account, approved jurisdiction/currency/rail corridor, webhook registration, or named evidence owner. Deterministic local fixtures exercise quote math, protocol transitions, custody accounting and provider contracts, but no real testnet swap, provider sandbox payout, low-value production swap or production payout has been executed and independently reconciled.

Therefore:

- every production support row remains `engineering_complete_external_blocked` or `unsupported`;
- the daemon must remain disabled or in non-executing engineering external-blocked mode;
- on-chain fiat conversion remains disabled after `v1.8.0` until governance commits independently approved exact profile IDs/digests; and
- real testnet/sandbox conformance remains an external evidence gate before certification, and production execution evidence remains mandatory for the production floor; and
- overall `planned_functionality_complete` remains blocked.

## Completion artifacts

- `_docs/protocols/task-85b-dex-payout-support-matrices.md`
- `_docs/protocols/fiat-conversion-orchestrator-protocol.md`
- `_docs/runbooks/fiat-conversion-incident-recovery.md`
- `_docs/task-85b-external-prerequisite-certification-ledger.md`
