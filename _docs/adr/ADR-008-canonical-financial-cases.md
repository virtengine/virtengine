# ADR-008: Settlement Owns Canonical Financial Cases

- **Status:** Accepted
- **Date:** 2026-07-22
- **Decision authority:** protocol engineering with governance-controlled activation at `v1.7.0`
- **Depends on:** ADR-006 authenticated metering and ADR-007 canonical reservations

## Context and evidence

Five existing surfaces overlap but do not have equivalent authority.

| Surface | Store/API ownership | Existing value authority | Evidence/compatibility value | Why it is not selected |
| --- | --- | --- | --- | --- |
| `x/settlement` | authenticated usage, escrow mirror, settlements, payouts, claimable rewards, payout ledgers and queries | directly moves/refunds settlement-module value and already holds payouts | complete order/invoice/usage/settlement lineage | Selected: it is the only store that can enforce one allocation before payout/reward execution. |
| `x/escrow` billing disputes | invoice workflow, corrections and historical billing queries | could change invoice status/corrections; lower-level escrow keeper owns deposits/payments | deployed billing compatibility surface | It has no authenticated usage/reward/reservation aggregate and would duplicate settlement payout authority. |
| `x/fraud` | encrypted fraud reports, moderator queue and audit records | recommendations included a legacy `refund` label but no atomic payout owner | fraud evidence and reputation recommendation | Moderators must not race a financial resolver or independently transfer/refund value. |
| `x/hpc` | jobs, accounting, rewards and HPC dispute projection | legacy reward distribution and dispute resolution can advance accounting | job/reward/capacity-specific evidence | HPC depends on settlement/resources and cannot become the financial owner without a cycle and duplicate billing authority. |
| `x/review` | reviews and provider rating aggregate | reputation only | content hash, verified order and moderation recommendation | Correctly remains evidence/reputation. Public review text is not financial evidence and cannot release value. |

Current code evidence also shows settlement already provides payout hold/refund callbacks, authenticated usage replay indexes, exact payout idempotency, multi-denom `sdk.Coins`, and Task 84C reservation lineage. Selecting any other keeper would require a second money ledger or a cycle back into settlement.

## Decision

`x/settlement` is the sole canonical mutable owner of `FinancialCase`. A case is rooted in a canonical subject group and indexes all known order, invoice, usage, HPC job, settlement, escrow, lease and reservation aliases. All filings merge as typed claims into that root.

The aggregate records canonical `provider` and `customer` separately from claimant/respondent filing roles. This is required because either financial party may file; terminal provider/customer allocation fields must never be interpreted as claimant/respondent shares.

Fraud, HPC, review/moderation and escrow billing are adapters/projections:

- they may open or add a privacy-safe claim;
- they may request review/escalation or add a recommendation;
- they may display the canonical case ID/status;
- they cannot release, refund, settle, reward, slash or transfer money independently after activation.

Only application-wired adapters may cross the trusted-adapter boundary. Public callers cannot self-declare a source module to bypass subject-party authorization.

The generated aggregate, messages, gRPC/REST queries and events are defined in the settlement v1 contracts. Existing wire fields and type URLs remain unchanged; additions use new field numbers.

## Financial authority and conservation

Exposure and terminal allocations use repeated Cosmos `Coin`/`sdk.Coins`. For every denomination independently:

$$
provider + customer + platform + slashWitness = originalHeld
$$

No float or cross-denom netting is permitted. Resolution first enters `RESOLVED_PENDING_APPEAL`; value remains held until the appeal deadline passes. Finalization applies persisted effects exactly once and then marks the case final. A retry observes completed markers and cannot repeat transfers.

## State and authorization summary

The normative transition and authorization tables are in `_docs/protocols/financial-case-state-machine.md`.

- parties file/add claims and may submit for review;
- governance authority resolves/escalated cases and cannot be claimant/respondent;
- appeals are party-only and bounded by parameter;
- cancellation is claimant/governance-only before review;
- timeouts escalate rather than silently release value;
- malformed state, missing dependencies or orphan indexes fail closed.

## Compatibility and activation

Before `v1.7.0`, historical APIs retain their existing replay behavior. At `v1.7.0`:

1. legacy settlement holds/disputed escrows are imported first;
2. financial fraud/HPC/billing records are added as claims;
3. duplicate aliases merge into one root;
4. ambiguous active value is quarantined with holds retained;
5. terminal completed finances are never reopened or rewritten;
6. module migrations rebuild indexes/effect defaults;
7. non-owner financial mutation paths become compatibility requests or reject with a stable fencing error.

Historical terminal records remain queryable. Read/projection APIs are not removed during this compatibility window.

## Consequences

- One query provides authoritative status and allocation for every financial lineage alias.
- Payout, settlement, reward and reservation paths have a single active-case guard.
- Adapter failures cannot partially commit because cross-module operations use cached contexts.
- Evidence on public state is limited to hashes and bounded encrypted references; narratives, manifests, biometrics, keys and raw evidence are forbidden.
- Operators must remediate ambiguous migration roots through governance; automatic migration never guesses an allocation.

## Rejected alternatives

1. **Escrow billing as owner:** rejected because invoice correction is not a complete payout/reward/reservation authority.
2. **Fraud as owner:** rejected because evidence moderation must not own customer/provider balances.
3. **HPC as owner:** rejected because it would privilege one product path and create keeper cycles.
4. **New dispute module:** rejected because it would duplicate settlement's value ledger and increase atomicity boundaries.
5. **Independent workflows with callbacks:** rejected because callback ordering permits conflicting terminal decisions and releases.
