# Task 85D External Prerequisite Certification Ledger

Status: `engineering_complete_external_blocked`

Local Task 85D helper-process tests are engineering conformance evidence only. The rows below remain blocked until retained external evidence exists.

| Area | Required external evidence | Current status |
| --- | --- | --- |
| Real OIDC/SAML providers | Signed attestations from real OIDC/SAML sandbox providers, retained provider metadata, request/nonce logs, and negative replay/tamper evidence. | Blocked; local deterministic SSO helper is not a real IdP. |
| Real email connector sandbox | Retained sandbox delivery/verification logs, provider key custody evidence, bounce/replay/freshness failures, and signed chain submission evidence. | Blocked; local deterministic email helper is not a provider sandbox. |
| Real SMS connector sandbox | Retained sandbox SMS delivery/verification logs, carrier/VoIP metadata, replay/freshness failures, and signed chain submission evidence. | Blocked; local deterministic SMS helper is not a provider sandbox. |
| Real social connector sandbox | Retained OAuth/social graph sandbox evidence, privacy-safe field hashes, encrypted payload custody evidence, and negative tamper cases. | Blocked; local deterministic social helper is not a provider sandbox. |
| Production HSM/KMS issuer custody | Evidence that issuer and receipt signer private keys are generated, stored, rotated, revoked, and audited in approved production HSM/KMS custody. | Blocked; local helper-process Ed25519 keys are deterministic test keys only. |
| Production mTLS CA/workload identity | Retained CA, certificate rotation, SPIFFE/workload identity or equivalent, revocation, and mutual-auth enforcement evidence. | Blocked; local mTLS tests are engineering checks only. |
| Production approved model/runtime/dataset | Approved model manifest, runtime image digest, dataset lineage, reproducibility record, deterministic CPU/seed config, and independent review. | Blocked; local deterministic sandbox runtime is not production model certification. |
| Live four-validator network | Retained live validator vote-extension, quorum, dissent exclusion, consensus finalization, and state/app-hash evidence from four real validators. | Blocked; package-level four-keeper test is in-process engineering evidence only. |
| Signed chain transaction/app-hash retention | Retained signed transactions, block heights, app hashes, receipts, and audit bundle proving state on a live chain. | Blocked; local keeper/app tests do not claim live-chain evidence. |

Certification cannot move beyond `engineering_complete_external_blocked` until these rows have external evidence and audit retention.
