# Canonical Financial Case State Machine v1

**Version:** 1  
**Owner:** `x/settlement`  
**Activation:** software upgrade `v1.7.0`

## Canonical identity

`case_id = "financial-case/" || hex(SHA-256(LP("virtengine/settlement/financial-case/v1"), LP(subject_key)))`.

`claim_id = "financial-claim/" || hex(SHA-256(LP("virtengine/settlement/financial-claim/v1"), LP(case_id), LP(claim_type), LP(claimant), LP(source_module), LP(source_reference), LP(evidence_hash), LP(encrypted_reference), LP(idempotency_key), LP(recommendation)))`.

`LP` is a four-byte big-endian length followed by bytes. No block time, random value or sequence participates in these IDs. Exact idempotency retries return the original case/claim. The same key with a different payload fails.

The canonical subject root has a declared type and primary ID but indexes every supplied alias: order, invoice, usage, HPC job, settlement, escrow, lease and reservation. An active alias can map to exactly one case. A filing that shares any active alias adds a claim to that root.

Each case separately records `provider` and `customer`. `claimant` and `respondent` are filing roles and may be either financial party; terminal `provider` and `customer` allocations always route to the explicit canonical financial roles. Public settlement messages must prove the signer is one of those committed parties. Only wired in-process fraud/HPC/billing adapters may use the trusted-adapter boundary, and they still must provide two distinct valid parties.

## Privacy and bounds

Only SHA-256 evidence hashes, bounded encrypted/content-addressed references and bounded recommendation codes are public. Raw evidence, descriptions, manifests, PII, biometrics, private narratives, ciphertext payloads and keys are forbidden in the case, event attributes and logs.

Defaults (governance configurable within hard protocol maxima):

- claims: 32, maximum 256;
- appeals: 1, maximum 8;
- encrypted reference: 512 bytes, maximum 4096;
- timeout processing: 100 cases/block, maximum 1000;
- idempotency key: 1-128 bytes;
- source module: 1-32 bytes;
- source reference/recommendation: at most 256 bytes.

## States

| From | To | Authorized actor | Required conditions |
| --- | --- | --- | --- |
| `OPEN` | `EVIDENCE` | party or adapter | valid unique claim |
| `OPEN`/`EVIDENCE` | `REVIEW` | claimant, respondent or governance | filing/evidence deadline valid |
| `OPEN`/`EVIDENCE`/`REVIEW` | `ESCALATED` | party, adapter moderator or governance | SHA-256 reason; holds remain |
| `REVIEW`/`ESCALATED`/`QUARANTINED` | `RESOLVED_PENDING_APPEAL` | governance resolver | resolver is not a party; allocation conserves every denom |
| `RESOLVED_PENDING_APPEAL` | `REVIEW` | claimant/respondent | appeal window open; appeal count below limit; allocation cleared; holds remain |
| `RESOLVED_PENDING_APPEAL` | `FINAL` | governance resolver | appeal window passed; all effects apply exactly once |
| `OPEN`/`EVIDENCE` | `CANCELLED` | claimant or governance | no review; holds restored |
| active nonterminal | `QUARANTINED` | migration/governance | ambiguity or malformed lineage; no release |
| timed-out open/evidence/review | `ESCALATED` | deterministic blocker | deadline passed; never automatic release |
| active | `REJECTED`/`EXPIRED` | governed recovery only | explicit policy allocation/hold restoration required |

`FINAL`, `REJECTED`, `CANCELLED` and `EXPIRED` are terminal. `QUARANTINED` is visible and nonterminal so value cannot be stranded invisibly.

## Deadlines

Each phase records both a block-height and block-time deadline from committed consensus inputs. A deadline is passed if either configured boundary is exceeded. Filing, evidence, review, appeal and escalation windows are governance parameters. Overflow saturates rather than wrapping.

The bounded EndBlock timeout worker escalates expired evidence/review cases. It does not release payout, rewards, escrow or capacity. Thus every active hold has a visible escalation/recovery route and no local wall clock participates.

## Holds

Opening an admissible value-bearing case occurs in a Cosmos cached context and establishes all available holds before the case commits:

- payout state `HELD`, bound to `case_id`;
- settlement escrow state `DISPUTED`;
- Task 84C reservation state `DISPUTED`, retaining capacity and recording its pre-dispute state;
- party claimable rewards are blocked while an active case indexes that party;
- HPC accounting/reward generation and settlement are blocked by job subject.

An exact hold retry for the same case is a no-op. A different case attempting to hold the same payout/reservation fails. Missing required keepers or malformed references fail the whole cached operation.

After activation, direct settlement `MsgDisputeEscrow` opens/merges the canonical case. Legacy settlement payout hold-release/refund callbacks and the independent escrow payout create/execute/cancel/refund paths return the stable financial-mutation-fenced error for canonical value; they cannot bypass appeal or effect markers.

## Allocation and effects

`original_exposure`, `provider`, `customer`, `platform` and `slash_witness` are sorted canonical `sdk.Coins`. For each denomination, allocations sum exactly to original exposure. Empty/extra denominations and negative/invalid coins fail before writes.

`original_exposure` includes every independently held value source exactly once. A payout-backed amount is not counted again as escrow balance; separately claimable rewards are added and their reward entries are consumed by the terminal reward effect before the party allocation becomes final.

Persisted effects are deterministic IDs under the case:

- `/provider`;
- `/customer`;
- `/platform`;
- `/slash-witness`;
- `/reservation` when applicable;
- `/projection`.

Every marker records pending/applied/failed, attempts, applied height/time and stable error code. All bank/store effects execute inside one cached context. A failed attempt commits neither marker nor transfer. Restart/retry converges; an applied marker cannot be re-applied.

Provider win pays provider; customer win refunds customer; partial/mutual split use the exact allocation; fraud-confirmed may allocate slash/witness value and terminally slash the reservation; inconclusive timeout requires a governed conserved allocation. Reservation cancellation restores its exact prior state; final resolution releases or slashes it exactly once.

## Adapter rules

- **Fraud:** financial reports open/add `FRAUD` claims. Resolution labels are reputation recommendations; financially linked resolution/escalation requests canonical escalation.
- **HPC:** `FlagDispute` opens/adds `HPC` claim and stores canonical ID/status. `ResolveDispute` after activation is an escalation request only. Rewards/accounting and reservation remain held.
- **Review:** if an active case exists for the reviewed order, only content hash and `rating:N` recommendation are added. Reviews never move value.
- **Escrow billing:** active legacy initiation projects a canonical invoice claim. Legacy resolution is fenced and requests canonical escalation; historical records remain queryable.

## Events and audit

Generated events: opened, claim-added, held, reviewed, escalated, resolved, appealed, finalized, effect-applied, quarantined and expired. Event fields are IDs, enums, counts and hashes only.

`FinancialCaseTransition` is append-only with contiguous sequence, from/to, actor, action, reason hash, block height and block time. Queries paginate cases and transition lineage. A malformed transition sequence, claim root, index, case record or effect fails invariants closed.
