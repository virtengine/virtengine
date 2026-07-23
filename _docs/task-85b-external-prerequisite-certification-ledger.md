# Task 85B External Prerequisite and Certification Ledger

**Ledger version:** 1.0
**Date:** 2026-07-23
**Current status:** `engineering_complete_external_blocked`

This ledger records the evidence required outside the checked-in engineering implementation. `TBD_EXTERNAL` means not supplied. Evidence references must be non-secret, immutable or access-controlled record IDs; credentials, contracts containing confidential terms, beneficiary data and KYC evidence must not be copied into this document.

## Blocking ledger

| Area | Required production evidence | Current reference | Evidence owner | Gate |
| --- | --- | --- | --- | --- |
| DEX pool/module and token route | Real allowlisted pool/module IDs; exact denoms, decimals and IBC traces; adapter-version applicability review | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Pools and liquidity | Retained pool state, reserves, fees, depth, route-size limits, price impact and funded liquidity for the exact row | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Oracle | Authenticated source, height/freshness/deviation policy, outage behavior, reconciliation and incident owner | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| DEX network access and finality | Approved RPC/archive endpoints, chain identity, finality policy, reorg drill and independently verified transaction evidence | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Custody/HSM | Production target-chain custody backend, funded account, non-test readiness, key ceremony, access policy, signing/withdrawal limits, recovery and rotation | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Custody governance | Approved signer role separation, provider authorization policy, emergency pause and audited key fingerprints/epochs | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Payout provider contract | Selected licensed provider, executed production contract, exact API/webhook versions, service limits and support/escalation terms | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Production credentials and secret manager | API/webhook credential references in an approved secret manager, least privilege, rotation/revocation test and access audit | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Banking/custody accounts | Approved settlement/funding/bank accounts, ownership evidence, funding limits and reconciliation access | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Corridor approval | Exact jurisdiction, currency, rail, beneficiary requirements, minimum/maximum/daily limits, finality definition and destination receipt policy | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Legal and regulatory | Legal opinion, licensing/provider reliance, terms/disclosures and jurisdiction-specific approval | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Compliance | KYC and sanctions decision authority, risk policy, revocation/expiry handling, manual review/SAR procedures and audit retention | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| DPA/privacy | Executed DPA, data-flow/retention/access review, beneficiary-tokenization service approval and cross-border assessment | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Webhook registration | Registered private endpoint, HMAC-SHA256 key IDs/versions, key ceremony/rotation overlap, ingress controls and replay/failure drill | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Durable operational stores | Approved persistence, backup/restore, encryption, ownership and monitoring for orchestrator, payout, webhook and Task 85A mutation state | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| External backend composition | Reviewed injected factories for DEX custody/finality, payout partner, secrets, destination, compliance and webhook ingress | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Governance/profile authority | Named profile authority, Ed25519 trust root ceremony, exact signed profile files/digests, on-chain proposal/height and pause/rollback authority | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Operations and security ownership | Named engineering, operations, security and compliance owners; on-call/escalation and incident exercise | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Low-value production DEX evidence | Signed production swap under approved limit with pool state, oracle, custody, inclusion/finality, fees and independently reconciled balances | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| Low-value production payout evidence | Production payout under approved limit with compliance decision, provider settlement, authenticated webhook, destination receipt and independent reconciliation | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |
| End-to-end conservation | One correlation lineage joining settlement, held value, quote, swap, custody movement, stable output, provider debit/fee, fiat receipt and terminal observation | `TBD_EXTERNAL` | `TBD_EXTERNAL` | Blocking |

## Certification sequence

The following sequence is mandatory; no individual item is sufficient by itself:

1. Select real DEX and payout rows and complete every applicable ledger entry with named owners and reviewable evidence.
2. Produce exact, non-placeholder profile files. Real pool IDs, denoms/decimals, limits, oracle, custody mode, provider/API/webhook versions and corridor data must match the evidence.
3. Obtain independent engineering, operations, security, legal and compliance sign-off. The profile authority signs the canonical profile digests from its separately controlled trust root.
4. Inject and review the external backend implementations. Production readiness must pass with non-test custody and durable stores; backend name strings are not acceptable composition.
5. Submit governance that commits the exact profile IDs/digests and activation/rollback plan. Default `v1.8.0` external-blocked state remains until this approval is final.
6. Execute approved low-value production DEX and payout operations. Independently reconcile chain/custody/provider/bank/destination and on-chain observation evidence with zero unexplained delta or duplicate movement.
7. Bind the complete evidence bundle to the release manifest and record owner approvals, incident controls and rollback readiness.

Only after all steps pass may a later matrix revision propose a production-enabled state. Until then, the daemon remains disabled or non-executing, the matrices remain external-blocked/unsupported, and `planned_functionality_complete` remains blocked.

## Current attestation

As of 2026-07-23, no production custody, provider contract/credentials, pool/liquidity/oracle/governance certification, approved corridor, or reconciled production execution evidence exists in the workspace. No certification transition is authorized by this ledger version.
