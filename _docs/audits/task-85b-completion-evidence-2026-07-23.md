# Task 85B Completion and Evidence Report

**Date:** 2026-07-23
**Status:** `engineering_complete_external_blocked`

Task 85B is locally engineering-complete. It is not production-certified or production-enabled. The mandatory Claim 4 floor remains externally blocked because no production custody backend, payout-provider contract or credentials, certified pool/liquidity/oracle/governance bundle, approved payout corridor, or independently reconciled production execution evidence has been supplied.

## Delivered engineering scope

- The production DEX factory now selects `RealOsmosisAdapter`, not the one-to-one placeholder. Its only implemented production math profile is Osmosis `gamm-v1beta1-cp-equal-weight-v1` for two-asset, equal-weight constant-product `osmosis.gamm.v1beta1.Pool` pools.
- The Osmosis path pins chain/profile identity, pool allowlists, exact denoms and decimals, observation age/height, liquidity/reserve/amount limits, hop bounds, price impact, oracle deviation, quote expiry, minimum output, canonical execution payloads, target-chain `TxRaw` verification, confirmation height/block identity, finality and actual output.
- Uniswap V2 and Curve implementations remain explicit test-only placeholders; production factory selection rejects them.
- The off-ramp package now provides a strict, provider-neutral JSON HTTP adapter with exact API/profile versions, fixed-origin/path allowlisting, bounded strict JSON, HTTPS/transport checks, secret references, idempotency and correlation binding, independent compliance authorization, tokenized beneficiary references, durable production daily limits, status reconciliation and ambiguous-result classification.
- HMAC-SHA256 webhook verification pins API/key versions, authenticates timestamp plus raw body before parsing, binds callbacks to an initiated payout, detects exact and conflicting replays, and requires durable replay/binding/event stores in production.
- `x/settlement` no longer retains DEX or off-ramp clients. It commits the intent/value hold and accepts the generated, provider-signed `MsgRecordFiatConversionObservation` through six ordered happy-path stages, plus bounded failed/cancelled outcomes.
- Observation sequencing, idempotency, profile digests, compliance digests, quote/finality evidence, lineage indexes, daily quota accounting, linked payout state, financial-case holds and terminal invariants are consensus-validated.
- The Task 85A mutation registry durably registers `settlement.record_fiat_conversion_observation`; observations are SDK-signed, persisted, broadcast and reconciled through the durable provider mutation queue.
- The provider-daemon orchestrator persists privacy-safe work state and payout/webhook repositories, uses a fenced local lease, recovers ambiguous swap/payout boundaries after restart, and never stores credentials, raw beneficiary data, raw execution payloads or signed transaction bytes.
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

Current generated artifact hashes were rechecked on 2026-07-23:

- Descriptor SHA-256: `d2b18c893b4b853fb783d77aa30b95a4879bb46f7ced0bf4be524ad48192fab6`
- Inventory SHA-256: `654f503d6e84c881f62882f2a57cca3bf8077b2772bac3176c452dc6ebac50c4`
- OpenAPI SHA-256: `e8b617f0722ba4cc26df5b66bb0a5d84328b5a08d226494dab1c75723ed88037`

### Race caveat outside Task 85B

A broad, unscoped provider-daemon race run exposed an unrelated pre-existing race around `CompromiseDetector.recordAlertDispatch()`. The focused Task 85B WSL race scope passed. This caveat is not Task 85B execution evidence and remains a separate provider-daemon defect to remediate before treating a broad package race gate as clean.

## External blockers and release consequence

No checked-in configuration or evidence identifies a real production pool, funded liquidity, token route, oracle, custody signer, provider contract, production credential, bank/custody account, approved jurisdiction/currency/rail corridor, webhook registration, or named evidence owner. No low-value production swap or payout has been executed and independently reconciled.

Therefore:

- every production support row remains `engineering_complete_external_blocked` or `unsupported`;
- the daemon must remain disabled or in non-executing engineering external-blocked mode;
- on-chain fiat conversion remains disabled after `v1.8.0` until governance commits independently approved exact profile IDs/digests; and
- overall `planned_functionality_complete` remains blocked.

## Completion artifacts

- `_docs/protocols/task-85b-dex-payout-support-matrices.md`
- `_docs/protocols/fiat-conversion-orchestrator-protocol.md`
- `_docs/runbooks/fiat-conversion-incident-recovery.md`
- `_docs/task-85b-external-prerequisite-certification-ledger.md`
