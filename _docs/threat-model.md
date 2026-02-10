# VirtEngine Threat Model

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Status:** Authoritative Baseline  
**Task Reference:** VE-000

---

## Table of Contents

1. [Overview](#overview)
2. [Threat Actors](#threat-actors)
3. [Attack Surface Analysis](#attack-surface-analysis)
4. [Threat Categories](#threat-categories)
   - [T1: Key Theft](#t1-key-theft)
   - [T2: Malicious Provider](#t2-malicious-provider)
   - [T3: Fraudulent Identity Uploads](#t3-fraudulent-identity-uploads)
   - [T4: Replay Attacks](#t4-replay-attacks)
   - [T5: Validator Collusion](#t5-validator-collusion)
   - [T6: Rogue Approved Client](#t6-rogue-approved-client)
   - [T9: ML Inference and Verification Services](#t9-ml-inference-and-verification-services)
5. [Additional Threats](#additional-threats)
6. [Risk Matrix](#risk-matrix)
7. [Mitigation Summary](#mitigation-summary)

---

## Overview

This threat model identifies and analyzes security threats to the VirtEngine platform, covering:

- **VEID (Identity Verification)**: ML-scored identity with validator consensus
- **Marketplace**: Encrypted orders/offerings with escrow
- **Provider Daemon**: Off-chain provisioning with on-chain usage records
- **Supercomputer/HPC**: SLURM-based distributed computing

### Scope

| In Scope | Out of Scope |
|----------|--------------|
| On-chain modules and state | Physical datacenter security |
| Provider daemon operations | Third-party orchestrator vulnerabilities |
| Client applications (approved) | End-user device security |
| Validator node operations | Network-level DDoS (ISP-level) |
| Cryptographic protocols | Quantum computing attacks |

---

## Threat Actors

| Actor | Motivation | Capability | Access Level |
|-------|------------|------------|--------------|
| **External Attacker** | Financial gain, chaos | Medium-High | None (network access only) |
| **Malicious User** | Fraud, free resources | Low-Medium | Registered account |
| **Malicious Provider** | Theft, data exfiltration | Medium-High | Provider daemon, infrastructure |
| **Rogue Validator** | Manipulation, theft | High | Validator node, consensus |
| **Colluding Validators** | Consensus manipulation | Very High | Multiple validator nodes |
| **Compromised Client Dev** | Backdoor, data theft | High | Approved client signing keys |
| **Insider (Admin)** | Sabotage, theft | Very High | Administrative access |
| **Nation-State** | Surveillance, disruption | Very High | Advanced persistent threat |

---

## Attack Surface Analysis

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ATTACK SURFACE MAP                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  CLIENT LAYER (Untrusted)                                            │   │
│  │  • Web Portal: XSS, CSRF, session hijacking                         │   │
│  │  • Mobile App: Reverse engineering, tampering, keylogging           │   │
│  │  • CLI/SDK: Credential exposure, insecure storage                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  API LAYER (Semi-Trusted)                                            │   │
│  │  • REST/gRPC: Injection, auth bypass, rate limiting                 │   │
│  │  • WebSocket: Connection hijacking, message tampering               │   │
│  │  • TLS: Downgrade attacks, certificate issues                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  BLOCKCHAIN LAYER (Trusted Compute)                                  │   │
│  │  • Transaction validation: Signature forgery, malformed tx          │   │
│  │  • State machine: Logic bugs, overflow, access control bypass       │   │
│  │  • Consensus: Byzantine faults, validator misbehavior               │   │
│  │  • Cryptography: Weak randomness, key exposure                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  OFF-CHAIN SERVICES (Provider-Controlled)                            │   │
│  │  • Provider Daemon: Impersonation, false usage, resource theft      │   │
│  │  • Orchestrators (K8s/SLURM): Container escape, privilege escalation│   │
│  │  • Waldur: SQL injection, auth bypass, data exposure                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Threat Categories

### T1: Key Theft

#### T1.1: User Account Key Compromise

| Attribute | Value |
|-----------|-------|
| **ID** | T1.1 |
| **Name** | User Account Key Compromise |
| **Description** | Attacker obtains user's private key through phishing, malware, or insecure storage |
| **Threat Actor** | External Attacker, Malicious User |
| **Impact** | HIGH - Full account takeover, fund theft, identity fraud |
| **Likelihood** | MEDIUM - Common attack vector |

**Attack Scenarios:**
1. Phishing site mimics VE Portal, captures mnemonic/private key
2. Malware on user device exports keystore file
3. User stores private key in plaintext (notes app, email)
4. Social engineering via fake support agent

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| MFA for sensitive transactions | Preventive | HIGH |
| Hardware wallet support (Ledger/Trezor) | Preventive | HIGH |
| Account recovery requires MFA + VEID | Preventive | HIGH |
| Session timeout + re-authentication | Detective | MEDIUM |
| Anomaly detection (unusual tx patterns) | Detective | MEDIUM |
| User education on phishing | Preventive | LOW |

**Residual Risk:** MEDIUM - MFA significantly reduces impact even if key is compromised.

---

#### T1.2: Validator Key Compromise

| Attribute | Value |
|-----------|-------|
| **ID** | T1.2 |
| **Name** | Validator Key Compromise |
| **Description** | Attacker obtains validator's consensus key or identity decryption key |
| **Threat Actor** | External Attacker, Insider |
| **Impact** | CRITICAL - Can decrypt identity data, manipulate consensus |
| **Likelihood** | LOW - Validators should have strong security |

**Attack Scenarios:**
1. Validator node compromised via unpatched vulnerability
2. Insider with access to validator HSM
3. Supply chain attack on validator software
4. Coercion/bribery of validator operator

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| HSM for validator keys | Preventive | HIGH |
| Threshold signatures (multi-party) | Preventive | HIGH |
| Key rotation procedures | Preventive | MEDIUM |
| Validator slashing for misbehavior | Deterrent | HIGH |
| Security audits of validator nodes | Detective | MEDIUM |
| Distributed identity decryption (threshold) | Preventive | HIGH |

**Residual Risk:** LOW - HSM + threshold signatures make key extraction extremely difficult.

---

#### T1.3: Approved Client Signing Key Compromise

| Attribute | Value |
|-----------|-------|
| **ID** | T1.3 |
| **Name** | Approved Client Signing Key Compromise |
| **Description** | Attacker obtains the private key used to sign identity uploads from approved clients |
| **Threat Actor** | External Attacker, Compromised Client Dev |
| **Impact** | HIGH - Can forge identity uploads, bypass gallery detection |
| **Likelihood** | LOW - Keys should be in secure enclaves |

**Attack Scenarios:**
1. Client app reverse engineered, key extracted
2. Build pipeline compromised, malicious key inserted
3. Developer laptop compromised with signing key access
4. Key stored in insecure cloud storage

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Hardware-backed keystore (TEE/Secure Enclave) | Preventive | HIGH |
| Key rotation with on-chain allowlist update | Preventive | MEDIUM |
| Code signing + integrity checks | Detective | HIGH |
| Build pipeline security (SLSA Level 3+) | Preventive | HIGH |
| Multiple approved clients (reduces single point) | Preventive | MEDIUM |
| Governance process for client approval | Preventive | MEDIUM |

**Residual Risk:** MEDIUM - Even with compromise, identity ML scoring provides secondary check.

---

### T2: Malicious Provider

#### T2.1: Resource Theft / Non-Delivery

| Attribute | Value |
|-----------|-------|
| **ID** | T2.1 |
| **Name** | Resource Theft / Non-Delivery |
| **Description** | Provider accepts order, takes escrow payment, but doesn't deliver resources |
| **Threat Actor** | Malicious Provider |
| **Impact** | MEDIUM - Customer loses funds, platform reputation damage |
| **Likelihood** | MEDIUM - Financial incentive exists |

**Attack Scenarios:**
1. Provider bids on order, wins, never provisions workload
2. Provider provisions but immediately terminates, reports minimal usage
3. Provider oversells capacity, delivers degraded performance
4. Provider goes offline during active lease

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Escrow with delayed settlement | Preventive | HIGH |
| Provision verification (customer confirms) | Detective | HIGH |
| Heartbeat/liveness checks | Detective | HIGH |
| Reputation scoring based on history | Deterrent | MEDIUM |
| Staking requirement for providers | Deterrent | HIGH |
| Slashing for non-delivery | Deterrent | HIGH |
| Dispute resolution process | Corrective | MEDIUM |

**Residual Risk:** LOW - Escrow + staking + slashing creates strong disincentives.

---

#### T2.2: Data Exfiltration from Customer Workloads

| Attribute | Value |
|-----------|-------|
| **ID** | T2.2 |
| **Name** | Data Exfiltration from Customer Workloads |
| **Description** | Provider accesses and steals data from customer containers/VMs |
| **Threat Actor** | Malicious Provider |
| **Impact** | HIGH - Customer data breach, confidentiality violation |
| **Likelihood** | MEDIUM - Provider has infrastructure access |

**Attack Scenarios:**
1. Provider mounts customer storage volumes and copies data
2. Provider inspects network traffic from customer containers
3. Provider accesses customer container memory
4. Provider modifies orchestrator to inject data collection

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Confidential computing (SGX/SEV) | Preventive | HIGH |
| Customer-managed encryption keys | Preventive | HIGH |
| Attestation of execution environment | Detective | HIGH |
| Network encryption (mTLS in-cluster) | Preventive | MEDIUM |
| Provider reputation + legal agreements | Deterrent | LOW |
| Audit logging at orchestrator level | Detective | MEDIUM |

**Residual Risk:** MEDIUM - Confidential computing not universally available; trust in provider remains.

---

#### T2.3: False Usage Reporting

| Attribute | Value |
|-----------|-------|
| **ID** | T2.3 |
| **Name** | False Usage Reporting |
| **Description** | Provider submits inflated usage records to receive higher payments |
| **Threat Actor** | Malicious Provider |
| **Impact** | MEDIUM - Customer overcharged, token inflation |
| **Likelihood** | MEDIUM - Direct financial incentive |

**Attack Scenarios:**
1. Provider reports 100% CPU usage when actual is 20%
2. Provider reports network egress that didn't occur
3. Provider modifies daemon to inflate metrics
4. Collusion between multiple providers to normalize fraud

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Customer-side usage verification | Detective | HIGH |
| Third-party attestation nodes | Detective | HIGH |
| Statistical anomaly detection | Detective | MEDIUM |
| Signed usage reports (provider reputation at stake) | Deterrent | MEDIUM |
| Benchmarking daemon cross-checks | Detective | MEDIUM |
| Slashing for proven fraud | Deterrent | HIGH |

**Residual Risk:** MEDIUM - Detection is probabilistic; small inflation may go unnoticed.

---

### T3: Fraudulent Identity Uploads

#### T3.1: Synthetic/Fake Identity Documents

| Attribute | Value |
|-----------|-------|
| **ID** | T3.1 |
| **Name** | Synthetic/Fake Identity Documents |
| **Description** | User uploads AI-generated or forged identity documents to gain verified status |
| **Threat Actor** | Malicious User, Organized Fraud |
| **Impact** | HIGH - Undermines identity verification, enables fraud |
| **Likelihood** | HIGH - Tools for document forgery widely available |

**Attack Scenarios:**
1. AI-generated passport/ID card images
2. Photoshopped documents with real template
3. Purchased stolen identity documents
4. Deep-fake selfie video to match forged document

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| ML-based document authenticity detection | Detective | HIGH |
| Liveness detection (blink, head turn, 3D) | Detective | HIGH |
| Cross-reference with external databases | Detective | MEDIUM |
| Device fingerprinting + behavioral analysis | Detective | MEDIUM |
| Salt-binding to prevent reuse | Preventive | MEDIUM |
| Validator consensus (multiple ML models) | Detective | HIGH |
| Progressive trust (start with low score) | Preventive | MEDIUM |

**Residual Risk:** MEDIUM - Arms race with forgery technology; continuous model updates required.

---

#### T3.2: Gallery Upload / Replay from Photo

| Attribute | Value |
|-----------|-------|
| **ID** | T3.2 |
| **Name** | Gallery Upload / Replay from Photo |
| **Description** | User uploads pre-existing photo instead of live capture |
| **Threat Actor** | Malicious User |
| **Impact** | MEDIUM - Identity verification bypassed |
| **Likelihood** | HIGH - Trivial attack without mitigations |

**Attack Scenarios:**
1. User screenshots victim's social media photo for selfie
2. User photographs a printed photo of victim
3. User plays video of victim on second device
4. User uploads old capture from device gallery

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Salt-binding in capture metadata | Preventive | HIGH |
| Approved-client signature (no gallery access) | Preventive | HIGH |
| EXIF/metadata analysis | Detective | MEDIUM |
| Liveness detection (motion, depth) | Detective | HIGH |
| Challenge-response (random gestures) | Detective | HIGH |
| Device attestation (camera API integrity) | Preventive | MEDIUM |

**Residual Risk:** LOW - Approved-client signature + liveness makes gallery attacks very difficult.

---

#### T3.3: Identity Document Reuse Across Accounts

| Attribute | Value |
|-----------|-------|
| **ID** | T3.3 |
| **Name** | Identity Document Reuse Across Accounts |
| **Description** | Same identity documents used to verify multiple accounts |
| **Threat Actor** | Malicious User, Sybil Attacker |
| **Impact** | MEDIUM - Sybil resistance undermined |
| **Likelihood** | MEDIUM - Motivated attackers will try |

**Attack Scenarios:**
1. User creates multiple accounts with same documents
2. User sells verified account credentials
3. Identity theft - attacker uses victim's documents

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Face embedding deduplication | Detective | HIGH |
| Document hash/feature deduplication | Detective | HIGH |
| One-verified-identity-per-account enforcement | Preventive | HIGH |
| Cross-account anomaly detection | Detective | MEDIUM |
| Identity revocation on fraud detection | Corrective | HIGH |

**Residual Risk:** LOW - Deduplication catches most reuse attempts.

---

### T4: Replay Attacks

#### T4.1: Transaction Replay

| Attribute | Value |
|-----------|-------|
| **ID** | T4.1 |
| **Name** | Transaction Replay |
| **Description** | Previously valid transaction replayed to cause duplicate state changes |
| **Threat Actor** | External Attacker |
| **Impact** | MEDIUM - Duplicate payments, order confusion |
| **Likelihood** | LOW - Standard protections in Cosmos SDK |

**Attack Scenarios:**
1. Attacker captures valid signed transaction, rebroadcasts
2. Cross-chain replay if same keys used on multiple chains
3. Fork replay if chain undergoes hard fork

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Cosmos SDK sequence numbers | Preventive | HIGH |
| Chain ID in transaction signature | Preventive | HIGH |
| Nonce tracking per account | Preventive | HIGH |
| Mempool duplicate detection | Preventive | HIGH |

**Residual Risk:** VERY LOW - Standard Cosmos SDK protections are robust.

---

#### T4.2: Identity Verification Replay

| Attribute | Value |
|-----------|-------|
| **ID** | T4.2 |
| **Name** | Identity Verification Replay |
| **Description** | Attacker replays old identity verification data to gain status |
| **Threat Actor** | Malicious User |
| **Impact** | MEDIUM - Stale identity used for verification |
| **Likelihood** | LOW - Salt-binding prevents |

**Attack Scenarios:**
1. User captures identity data, uses for future verification
2. Attacker intercepts identity upload, replays for different account
3. Compromised client replays old captures

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Per-upload salt in metadata | Preventive | HIGH |
| Timestamp validation (freshness) | Preventive | HIGH |
| Salt bound to account + session | Preventive | HIGH |
| Approved-client signature includes salt | Preventive | HIGH |
| Nonce in verification request | Preventive | HIGH |

**Residual Risk:** VERY LOW - Multiple binding factors prevent replay.

---

#### T4.3: MFA Token Replay

| Attribute | Value |
|-----------|-------|
| **ID** | T4.3 |
| **Name** | MFA Token Replay |
| **Description** | Attacker captures and replays MFA verification token |
| **Threat Actor** | External Attacker |
| **Impact** | HIGH - MFA bypass |
| **Likelihood** | LOW - Short validity windows |

**Attack Scenarios:**
1. Man-in-the-middle captures TOTP code, uses before expiry
2. Session token stolen, reused for sensitive transaction
3. WebAuthn assertion replayed

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| TOTP codes valid for single use + 30s window | Preventive | HIGH |
| WebAuthn challenge-response (nonce-bound) | Preventive | HIGH |
| Session binding to IP/device | Preventive | MEDIUM |
| Short session timeouts for elevated privileges | Preventive | HIGH |
| Rate limiting on MFA attempts | Preventive | HIGH |

**Residual Risk:** LOW - Standard MFA protocols have replay protection.

---

### T5: Validator Collusion

#### T5.1: Consensus Manipulation (>1/3 Byzantine)

| Attribute | Value |
|-----------|-------|
| **ID** | T5.1 |
| **Name** | Consensus Manipulation |
| **Description** | >1/3 of validators collude to halt chain or censor transactions |
| **Threat Actor** | Colluding Validators |
| **Impact** | CRITICAL - Chain halts, censorship |
| **Likelihood** | VERY LOW - Requires significant stake control |

**Attack Scenarios:**
1. 34%+ of stake refuses to vote, halting finality
2. Validators censor specific user's transactions
3. Validators produce empty blocks to degrade performance

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Decentralized validator set (min 21+ validators) | Preventive | HIGH |
| Stake distribution monitoring | Detective | HIGH |
| Slashing for liveness failures | Deterrent | HIGH |
| Governance intervention capability | Corrective | MEDIUM |
| Geographic/jurisdictional distribution | Preventive | MEDIUM |

**Residual Risk:** LOW - Economic incentives align validators with honest behavior.

---

#### T5.2: Identity Score Manipulation (>2/3 Collude)

| Attribute | Value |
|-----------|-------|
| **ID** | T5.2 |
| **Name** | Identity Score Manipulation |
| **Description** | >2/3 of validators collude to approve fraudulent identity or reject legitimate one |
| **Threat Actor** | Colluding Validators, Bribed Validators |
| **Impact** | HIGH - Identity system integrity compromised |
| **Likelihood** | VERY LOW - Requires massive coordination |

**Attack Scenarios:**
1. Validators agree to score known fraudulent identity as 100
2. Validators agree to reject competitor's identity
3. Validators run modified ML model to bias outcomes

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Deterministic ML inference (same result required) | Detective | HIGH |
| Model version pinning in chain config | Preventive | HIGH |
| Third-party audit of validator behavior | Detective | MEDIUM |
| Whistleblower rewards for reporting collusion | Deterrent | MEDIUM |
| Governance oversight of identity disputes | Corrective | MEDIUM |
| Threshold decryption (prevents single-validator access) | Preventive | HIGH |

**Residual Risk:** LOW - Deterministic verification makes collusion detectable.

---

#### T5.3: Validator Data Theft (Identity Decryption Abuse)

| Attribute | Value |
|-----------|-------|
| **ID** | T5.3 |
| **Name** | Validator Data Theft |
| **Description** | Validator decrypts identity data and exfiltrates for malicious use |
| **Threat Actor** | Rogue Validator |
| **Impact** | CRITICAL - Mass identity data breach |
| **Likelihood** | LOW - Strong disincentives, technical controls |

**Attack Scenarios:**
1. Validator stores decrypted identity data before/after scoring
2. Validator sells identity data on dark web
3. Validator shares decryption keys with external party

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Threshold decryption (t-of-n validators required) | Preventive | HIGH |
| Secure enclave for decryption (SGX/TDX) | Preventive | HIGH |
| Audit logging of decryption operations | Detective | HIGH |
| Data minimization (decrypt only needed fields) | Preventive | MEDIUM |
| Legal/contractual obligations on validators | Deterrent | LOW |
| Slashing for proven data breach | Deterrent | HIGH |

**Residual Risk:** MEDIUM - Threshold decryption + enclaves significantly reduce risk.

---

### T6: Rogue Approved Client

#### T6.1: Backdoored Approved Client

| Attribute | Value |
|-----------|-------|
| **ID** | T6.1 |
| **Name** | Backdoored Approved Client |
| **Description** | Approved client software contains hidden functionality to steal data or forge uploads |
| **Threat Actor** | Compromised Client Dev, Supply Chain Attacker |
| **Impact** | CRITICAL - All users of client compromised |
| **Likelihood** | LOW - Requires sophisticated attack |

**Attack Scenarios:**
1. Malicious code in dependency captures identity data
2. Build server compromised, injects backdoor
3. Developer intentionally adds exfiltration code
4. Update mechanism hijacked to push malicious version

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Code signing with hardware keys | Preventive | HIGH |
| Reproducible builds | Detective | HIGH |
| Third-party security audits | Detective | HIGH |
| App store review (iOS/Android) | Detective | MEDIUM |
| Open-source client code | Detective | MEDIUM |
| SLSA Level 3+ build provenance | Preventive | HIGH |
| Multiple approved clients (diversity) | Preventive | MEDIUM |

**Residual Risk:** MEDIUM - Supply chain security is challenging; defense in depth required.

---

#### T6.2: Client Impersonation

| Attribute | Value |
|-----------|-------|
| **ID** | T6.2 |
| **Name** | Client Impersonation |
| **Description** | Attacker creates fake client that appears legitimate to users |
| **Threat Actor** | External Attacker |
| **Impact** | HIGH - Credential theft, fraudulent uploads |
| **Likelihood** | MEDIUM - Phishing is common |

**Attack Scenarios:**
1. Fake mobile app in third-party app store
2. Phishing site with fake web capture interface
3. Malware masquerading as VE client

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Approved-client signature (not in fake apps) | Preventive | HIGH |
| Official app store distribution only | Preventive | MEDIUM |
| User education on official sources | Preventive | LOW |
| Domain verification for web clients | Preventive | MEDIUM |
| Client attestation (platform-verified) | Preventive | HIGH |

**Residual Risk:** MEDIUM - Users may still be tricked; approved-client signature is key defense.

---

## Additional Threats

### T7: Denial of Service

| ID | Name | Impact | Likelihood | Key Mitigations |
|----|------|--------|------------|-----------------|
| T7.1 | Blockchain spam (tx flood) | MEDIUM | MEDIUM | Gas fees, mempool limits |
| T7.2 | API layer DDoS | MEDIUM | HIGH | Rate limiting, CDN, WAF |
| T7.3 | Provider daemon overload | LOW | MEDIUM | Backpressure, circuit breakers |
| T7.4 | Validator node DDoS | HIGH | MEDIUM | Sentry nodes, IP hiding |

### T8: Smart Contract / Module Bugs

| ID | Name | Impact | Likelihood | Key Mitigations |
|----|------|--------|------------|-----------------|
| T8.1 | State machine logic error | CRITICAL | LOW | Audits, formal verification |
| T8.2 | Integer overflow/underflow | HIGH | LOW | Safe math, audits |
| T8.3 | Access control bypass | CRITICAL | LOW | RBAC testing, audits |
| T8.4 | Reentrancy (if applicable) | HIGH | VERY LOW | Cosmos SDK patterns |

### T9: Insider Threats

| ID | Name | Impact | Likelihood | Key Mitigations |
|----|------|--------|------------|-----------------|
| T9.1 | Admin key abuse | CRITICAL | LOW | Multi-sig, timelocks |
| T9.2 | Genesis account compromise | CRITICAL | VERY LOW | Hardware custody, multi-sig |
| T9.3 | Support agent data access | MEDIUM | MEDIUM | Least privilege, audit logs |

---

## Risk Matrix

```
                           IMPACT
           │ NEGLIGIBLE │   LOW    │  MEDIUM  │   HIGH   │ CRITICAL │
     ──────┼────────────┼──────────┼──────────┼──────────┼──────────┤
     VERY  │            │          │          │          │          │
     HIGH  │            │          │   T3.1   │   T1.1   │          │
           │            │          │   T3.2   │          │          │
LIKELIHOOD ├────────────┼──────────┼──────────┼──────────┼──────────┤
     HIGH  │            │          │   T7.2   │          │          │
           │            │          │          │          │          │
     ──────┼────────────┼──────────┼──────────┼──────────┼──────────┤
     MEDIUM│            │   T7.3   │   T2.1   │   T2.2   │          │
           │            │          │   T2.3   │   T6.2   │          │
           │            │          │   T3.3   │          │          │
     ──────┼────────────┼──────────┼──────────┼──────────┼──────────┤
     LOW   │            │          │   T4.3   │   T1.3   │   T1.2   │
           │            │          │          │   T5.3   │   T5.1   │
           │            │          │          │          │   T6.1   │
           │            │          │          │          │   T8.1   │
     ──────┼────────────┼──────────┼──────────┼──────────┼──────────┤
     VERY  │            │   T4.1   │          │          │   T5.2   │
     LOW   │            │   T4.2   │          │          │   T9.2   │
           │            │          │          │          │          │
```

---

## Mitigation Summary

### High-Priority Controls

| Control | Threats Addressed | Implementation Status |
|---------|-------------------|----------------------|
| MFA for sensitive transactions | T1.1, T4.3 | Planned (VE-102) |
| Threshold decryption for identity | T1.2, T5.3 | Planned |
| Approved-client signatures | T3.2, T6.1, T6.2 | Planned (VE-201) |
| Salt-binding for identity uploads | T3.2, T4.2 | Planned (VE-207) |
| Deterministic ML inference | T5.2 | Planned (VE-203, VE-205) |
| Escrow + staking for providers | T2.1, T2.3 | Planned |
| Liveness detection | T3.1, T3.2 | Planned (VE-210) |
| HSM for validator keys | T1.2 | Recommended |

### Defense-in-Depth Layers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         DEFENSE-IN-DEPTH STACK                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 7: USER AWARENESS                                                    │
│  └─ Education, phishing training, official source guidance                 │
│                                                                              │
│  Layer 6: APPLICATION SECURITY                                              │
│  └─ Input validation, secure coding, dependency scanning                   │
│                                                                              │
│  Layer 5: IDENTITY & ACCESS                                                 │
│  └─ VEID scoring, MFA gating, RBAC, least privilege                       │
│                                                                              │
│  Layer 4: DATA PROTECTION                                                   │
│  └─ Encryption at rest/transit, threshold decryption, key management       │
│                                                                              │
│  Layer 3: CONSENSUS SECURITY                                                │
│  └─ PoS slashing, deterministic verification, validator distribution       │
│                                                                              │
│  Layer 2: NETWORK SECURITY                                                  │
│  └─ TLS, rate limiting, DDoS protection, sentry nodes                      │
│                                                                              │
│  Layer 1: INFRASTRUCTURE                                                    │
│  └─ HSMs, secure enclaves, hardened OS, monitoring                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## T9: ML Inference and Verification Services

### T9.1: ML Model Tampering

| Attribute | Value |
|-----------|-------|
| **ID** | T9.1 |
| **Name** | ML Model Tampering |
| **Description** | Attacker modifies ML model weights or inference code to produce biased or incorrect verification scores |
| **Threat Actor** | Rogue Validator, Compromised Server |
| **Impact** | CRITICAL - Identity verification integrity compromised |
| **Likelihood** | LOW - Model integrity checks in place |

**Attack Scenarios:**
1. Validator replaces model weights with malicious version
2. Inference code modified to always approve specific identities
3. Model poisoning through adversarial updates

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Model version pinning in chain config | Preventive | HIGH |
| Model hash verification before inference | Detective | HIGH |
| Deterministic inference with hash comparison | Detective | HIGH |
| Third-party model audits | Detective | MEDIUM |

**Residual Risk:** LOW - Multiple integrity checks prevent tampering.

---

### T9.2: Non-Deterministic Inference

| Attribute | Value |
|-----------|-------|
| **ID** | T9.2 |
| **Name** | Non-Deterministic Inference |
| **Description** | ML inference produces different results across validators, breaking consensus |
| **Threat Actor** | Accidental, Environment Variance |
| **Impact** | HIGH - Consensus failures, chain halts |
| **Likelihood** | MEDIUM - Requires careful configuration |

**Attack Scenarios:**
1. GPU floating-point variance causes score differences
2. Random seed not properly fixed across validators
3. Library version differences cause numerical variance

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| CPU-only inference enforcement | Preventive | HIGH |
| Fixed random seed (42) | Preventive | HIGH |
| Deterministic TensorFlow ops | Preventive | HIGH |
| Hash precision normalization (6 decimals) | Preventive | HIGH |
| Cross-validator conformance tests | Detective | HIGH |

**Residual Risk:** LOW - Comprehensive determinism controls in place.

---

### T9.3: Verification Service Bypass

| Attribute | Value |
|-----------|-------|
| **ID** | T9.3 |
| **Name** | Verification Service Bypass |
| **Description** | Attacker bypasses email/SMS/OIDC verification to claim identity |
| **Threat Actor** | External Attacker, Malicious User |
| **Impact** | HIGH - Fraudulent identity verification |
| **Likelihood** | LOW - Multiple verification layers |

**Attack Scenarios:**
1. OTP brute-force attack
2. SMS interception via SIM swap
3. OIDC token manipulation
4. Email verification link guessing

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| OTP rate limiting (5/minute) | Preventive | HIGH |
| Cryptographically random OTPs (6+ digits) | Preventive | HIGH |
| Short OTP expiry (5 minutes) | Preventive | HIGH |
| SMS anti-fraud (VoIP detection) | Preventive | HIGH |
| OIDC signature verification | Preventive | HIGH |
| Multi-factor attestation requirement | Preventive | HIGH |

**Residual Risk:** LOW - Defense-in-depth verification.

---

### T9.4: Attestation Replay Attack

| Attribute | Value |
|-----------|-------|
| **ID** | T9.4 |
| **Name** | Attestation Replay Attack |
| **Description** | Attacker replays valid attestation to gain unauthorized verification |
| **Threat Actor** | External Attacker |
| **Impact** | MEDIUM - Unauthorized verification claims |
| **Likelihood** | VERY LOW - Nonce binding prevents replay |

**Attack Scenarios:**
1. Capture and replay signed attestation
2. Reuse nonce across different subjects
3. Cross-issuer nonce manipulation

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| Cryptographic nonce binding | Preventive | HIGH |
| Issuer fingerprint binding | Preventive | HIGH |
| Subject address binding | Preventive | HIGH |
| Configurable expiry window | Preventive | HIGH |
| Atomic validate-and-use operation | Preventive | HIGH |

**Residual Risk:** VERY LOW - Comprehensive replay protection.

---

### T9.5: Key Rotation Failures

| Attribute | Value |
|-----------|-------|
| **ID** | T9.5 |
| **Name** | Key Rotation Failures |
| **Description** | Signer key rotation fails, leaving old keys active or new keys unusable |
| **Threat Actor** | Accidental, Rogue Administrator |
| **Impact** | MEDIUM - Signing service degradation |
| **Likelihood** | LOW - Overlapping key rotation |

**Attack Scenarios:**
1. Old key not revoked after rotation
2. New key activation fails, leaving no active keys
3. Key state inconsistency across services

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| State machine key transitions | Preventive | HIGH |
| Overlapping key validity period | Preventive | HIGH |
| Audit logging of key operations | Detective | HIGH |
| Private key memory clearing | Preventive | HIGH |
| Key storage encryption (AES-GCM) | Preventive | HIGH |

**Residual Risk:** LOW - Robust key lifecycle management.

---

### T9.6: SMS Fraud Abuse

| Attribute | Value |
|-----------|-------|
| **ID** | T9.6 |
| **Name** | SMS Fraud Abuse |
| **Description** | Attacker exploits SMS verification for toll fraud or account enumeration |
| **Threat Actor** | Malicious User, Fraud Ring |
| **Impact** | MEDIUM - Financial loss, service abuse |
| **Likelihood** | MEDIUM - SMS is inherently vulnerable |

**Attack Scenarios:**
1. VoIP number farming for mass verification
2. Velocity abuse from single IP
3. Device fingerprint spoofing
4. Premium rate number exploitation

**Mitigations:**
| Control | Type | Effectiveness |
|---------|------|---------------|
| VoIP carrier detection | Preventive | HIGH |
| Per-phone velocity limits | Preventive | HIGH |
| Per-IP velocity limits | Preventive | HIGH |
| Device fingerprint tracking | Detective | MEDIUM |
| Risk scoring with thresholds | Preventive | HIGH |
| Toll-free number blocking | Preventive | HIGH |

**Residual Risk:** MEDIUM - SMS inherently less secure than other methods.

---

## Appendix: STRIDE Mapping

| Threat Category | STRIDE | Primary Threats |
|-----------------|--------|-----------------|
| Spoofing | S | T1.1, T6.2, T3.1, T3.2, T9.3 |
| Tampering | T | T2.3, T5.2, T6.1, T9.1 |
| Repudiation | R | (Mitigated by on-chain audit logs) |
| Information Disclosure | I | T1.2, T2.2, T5.3 |
| Denial of Service | D | T7.1, T7.2, T5.1, T9.2 |
| Elevation of Privilege | E | T1.1, T1.3, T8.3, T9.3 |

---

*Document maintained by VirtEngine Security Team*  
*Last updated: 2026-01-25*  
*Security Review: VE-8D - ML and Verification Services*
