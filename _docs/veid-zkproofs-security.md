# VirtEngine VEID Zero-Knowledge Proof System

## Overview

The VirtEngine VEID module implements privacy-preserving identity proofs using zero-knowledge cryptography. This document details the cryptographic assumptions, security properties, and implementation choices for the ZK proof system.

## Table of Contents

1. [Cryptographic Scheme](#cryptographic-scheme)
2. [Proof Types](#proof-types)
3. [Security Assumptions](#security-assumptions)
4. [Performance Characteristics](#performance-characteristics)
5. [Determinism and Consensus Safety](#determinism-and-consensus-safety)
6. [Trusted Setup Requirements](#trusted-setup-requirements)
7. [Known Limitations](#known-limitations)
8. [Production Deployment](#production-deployment)
9. [References](#references)

## Cryptographic Scheme

### Groth16 ZK-SNARKs

The VirtEngine VEID module uses **Groth16** zero-knowledge Succinct Non-interactive Arguments of Knowledge (zk-SNARKs) for privacy-preserving identity proofs.

**Implementation:** [gnark library](https://github.com/consensys/gnark) v0.14.0

**Curve:** BN254 (also known as BN128 or alt_bn128)

**Security Level:** ~100-bit security (based on the hardness of the computational Diffie-Hellman problem over BN254)

**Proof Structure:**

- Proof size: ~200 bytes (3 group elements: 2 G1 points + 1 G2 point)
- Public inputs: Variable (embedded in verification)
- Private inputs (witness): Never revealed

### Why Groth16?

1. **Smallest proof size** - Critical for blockchain storage and verification costs
2. **Fast verification** - ~2-5ms per proof, suitable for consensus validation
3. **Mature implementation** - Well-tested gnark library with formal security analysis
4. **EVM compatibility** - BN254 curve is natively supported in Ethereum precompiles
5. **Deterministic verification** - All validators compute identical results

## Proof Types

### 1. Age Range Proofs

**Purpose:** Prove age >= threshold without revealing date of birth

**Circuit:** `AgeRangeCircuit`

**Public Inputs:**

- `ageThreshold` (uint32) - Minimum age to prove
- `currentTimestamp` (int64) - Current block time
- `commitmentHash` (bytes32) - Binding commitment to DOB

**Private Inputs (Witness):**

- `dateOfBirth` (int64) - Unix timestamp of birth date
- `salt` (bytes32) - Random salt for commitment binding

**Constraints:** ~500 R1CS (Rank-1 Constraint System) constraints

**Verification Logic:**

```
ageSeconds = currentTimestamp - dateOfBirth
ageYears = ageSeconds / 31557600  // seconds per year
ASSERT ageYears >= ageThreshold
ASSERT commitmentHash == hash(dateOfBirth, salt)
```

**Security Properties:**

- **Zero-knowledge:** Verifier learns only that age >= threshold, not actual DOB
- **Soundness:** Computationally infeasible to prove false age under CDH assumption
- **Commitment binding:** Cannot reuse proof for different DOB without knowledge of salt
- **Non-malleability:** Proof cannot be modified to prove different threshold

### 2. Residency Proofs

**Purpose:** Prove residency in country without revealing full address

**Circuit:** `ResidencyCircuit`

**Public Inputs:**

- `countryCodeHash` (bytes32) - Hash of ISO 3166-1 alpha-2 country code
- `commitmentHash` (bytes32) - Binding commitment to full address

**Private Inputs (Witness):**

- `fullAddressHash` (bytes32) - Hash of complete address
- `addressCountry` (bytes32) - Country code extracted from address
- `salt` (bytes32) - Random salt for commitment binding

**Constraints:** ~400 R1CS constraints

**Verification Logic:**

```
ASSERT countryCodeHash == addressCountry
ASSERT commitmentHash == hash(fullAddressHash, salt)
```

**Security Properties:**

- **Zero-knowledge:** Full address remains private, only country revealed
- **Soundness:** Cannot prove false residency without valid address witness
- **Commitment binding:** Links proof to specific address cryptographically
- **Privacy:** Verifier learns only country, not city/state/street

### 3. Score Range Proofs

**Purpose:** Prove trust score >= threshold without revealing exact score

**Circuit:** `ScoreRangeCircuit`

**Public Inputs:**

- `scoreThreshold` (uint32) - Minimum score to prove
- `commitmentHash` (bytes32) - Binding commitment to actual score

**Private Inputs (Witness):**

- `actualScore` (uint32) - Real trust score value
- `salt` (bytes32) - Random salt for commitment binding

**Constraints:** ~300 R1CS constraints

**Verification Logic:**

```
ASSERT actualScore >= scoreThreshold
ASSERT commitmentHash == hash(actualScore, salt)
```

**Security Properties:**

- **Zero-knowledge:** Exact score hidden, only threshold satisfaction revealed
- **Soundness:** Cannot prove higher score without valid score witness
- **Commitment binding:** Prevents score reuse attacks
- **Non-interactive:** No challenge-response protocol needed

### 4. Selective Disclosure Proofs

**Purpose:** Prove arbitrary claims about identity without revealing underlying data

**Implementation:** Composition of specialized circuits based on claim types

**Claim Types Supported:**

- `ClaimTypeAgeOver18/21/25` - Age threshold proofs
- `ClaimTypeCountryResident` - Residency proofs
- `ClaimTypeTrustScoreAbove` - Score range proofs
- `ClaimTypeEmailVerified` - Email ownership proofs
- `ClaimTypeSMSVerified` - Phone number proofs
- `ClaimTypeBiometricVerified` - Biometric verification proofs

**Security Properties:**

- **Composability:** Multiple claims can be proven in single proof
- **Selective disclosure:** Only requested claims revealed, rest hidden
- **Request binding:** Proof tied to specific disclosure request via nonce
- **Expiration:** Time-bounded validity prevents replay attacks

## Security Assumptions

### Cryptographic Assumptions

1. **Computational Diffie-Hellman (CDH) Assumption:**
   - Given g, g^a, g^b, it is hard to compute g^(ab)
   - Required for soundness of Groth16 proofs
   - BN254 curve provides ~100-bit security under this assumption

2. **Knowledge of Exponent (KEA) Assumption:**
   - Required for argument of knowledge property
   - Ensures prover actually knows witness values

3. **Random Oracle Model:**
   - Hash functions (SHA-256) modeled as random oracles
   - Used for commitment schemes and proof binding

4. **Trusted Setup Assumption:**
   - Common Reference String (CRS) generated honestly
   - Toxic waste (setup randomness) properly destroyed
   - Multi-party computation mitigates this assumption

### Threat Model

**Adversary Capabilities:**

- Full control over prover's inputs (can lie about age, address, score)
- Full access to all public blockchain data
- Cannot break underlying cryptographic assumptions
- Cannot access private witness values without key material

**Security Guarantees:**

1. **Soundness:** Adversary cannot prove false statement (probability ≤ 2^-100)
2. **Zero-knowledge:** Adversary learns nothing beyond claim truth value
3. **Non-malleability:** Cannot modify proof to prove different claim
4. **Replay resistance:** Nonces and expiration prevent proof reuse

**Out of Scope:**

- Side-channel attacks on client devices (implementation responsibility)
- Compromise of client private keys (user responsibility)
- Physical attacks on hardware (deployment environment responsibility)
- Social engineering attacks (user education responsibility)

## Performance Characteristics

### Proof Generation (Off-Chain)

| Proof Type           | Generation Time | Memory Usage | Deterministic    |
| -------------------- | --------------- | ------------ | ---------------- |
| Age Range            | 100-300ms       | ~50 MB       | ❌ (client-side) |
| Residency            | 80-250ms        | ~40 MB       | ❌ (client-side) |
| Score Range          | 50-150ms        | ~30 MB       | ❌ (client-side) |
| Selective Disclosure | 100-500ms       | ~60 MB       | ❌ (client-side) |

**Note:** Proof generation happens off-chain on user's client device. Uses randomness and is non-deterministic by design.

### Proof Verification (On-Chain)
| Selective Disclosure | 2-6ms             | ~220
| Proof Type           | Verification Time | Gas Cost (Est.) | Deterministic       |
| -------------------- | ----------------- | --------------- | ------------------- |
| Age Range            | 2-5ms             | ~200k gas       | ✅ (consensus-safe) |
| Residency            | 2-5ms             | ~180k gas       | ✅ (consensus-safe) |
| Score Range          | 2-4ms             | ~150k gas       | ✅ (consensus-safe) |
k gas       | ✅ (consensus-safe) |

**Note:** Verification happens on-chain during consensus. Fully deterministic - all validators compute identical results.

### Circuit Compilation (One-Time Setup)

| Circuit     | Compilation Time | Setup Time | Proving Key Size | Verification Key Size |
| ----------- | ---------------- | ---------- | ---------------- | --------------------- |
| Age Range   | 1-3s             | 5-10s      | ~50 MB           | ~2 KB                 |
| Residency   | 1-2s             | 4-8s       | ~40 MB           | ~2 KB                 |
| Score Range | 0.5-1.5s         | 3-6s       | ~30 MB           | ~2 KB                 |

**Note:** Compilation and trusted setup happen once during deployment, not per-proof.

### Blockchain Storage Costs

| Data Type           | Size           | Storage Location                  |
| ------------------- | -------------- | --------------------------------- |
| Proof bytes         | ~200 bytes     | On-chain (transaction data)       |
| Commitment hash     | 32 bytes       | On-chain (proof metadata)         |
| Nonce               | 32 bytes       | On-chain (proof metadata)         |
| Verification key    | ~2 KB          | On-chain (module state, one-time) |
| **Total per proof** | **~264 bytes** | **Per transaction**               |

## Determinism and Consensus Safety

### Why Determinism Matters

In blockchain consensus, all validators must compute identical state transitions. Non-deterministic operations cause chain halts.

### Deterministic Operations

✅ **Safe for consensus:**

- Proof verification (pure function)
- Commitment verification (hash-based, deterministic, claim keys sorted)
- Public input validation (pure computation)
- Nonces and salts:
  - Client-provided via `RandomnessInputs` (stored verbatim)
  - Or derived from tx context via `DeterministicRandomSource` (chain-id, block height/time, tx bytes, purpose labels)
  - Enforced as fixed 32-byte values via `resolveRandomBytes` for consistent proof IDs and commitments

### Non-Deterministic Operations

⚠️ **Unsafe for consensus (kept off-chain):**

- Full proof generation using secret witnesses
- Trusted setup (one-time, off-chain ceremony)
- Witness computation (private, client-side only)

### Implementation Strategy

1. **Client-side proof generation:**
   - User generates proof on their device
   - Uses randomness for zero-knowledge property
   - Submits proof bytes to blockchain

2. **On-chain deterministic verification:**
   - Validators verify proof using public inputs
   - No randomness in verification path
   - All validators compute same accept/reject result

3. **Fallback for testing:**
   - Hash-based commitments for test environments
   - Deterministic proof structure validation
   - Maintains same API and security properties
4. **Deterministic randomness pipeline:**
   - Clients MAY pass 32-byte nonces/salts through `RandomnessInputs`
   - If omitted, validators derive deterministic bytes from transaction context with domain separation
   - Prevents divergent randomness across validators while keeping proofs reproducible

## Trusted Setup Requirements

### What is a Trusted Setup?

Groth16 requires a **circuit-specific trusted setup** to generate proving and verification keys. The setup process generates toxic waste (random values) that must be destroyed.

### Setup Phases

1. **Powers of Tau Ceremony (Universal):**
   - Multi-party computation (MPC) with N participants
   - Only 1 honest participant needed for security
   - Can be reused across multiple circuits
   - Public ceremony with verifiable transcript

2. **Circuit-Specific Setup:**
   - Performed per circuit (age, residency, score)
   - Generates proving key (PK) and verification key (VK)
   - VK published on-chain for verification
   - PK distributed to clients for proving

### Security Properties

**Assumption:** At least one participant in MPC was honest and destroyed their secret.

**If assumption holds:**

- Proofs are sound (cannot prove false statements)
- Zero-knowledge property maintained
- System is secure

**If assumption fails:**

- Adversary can generate fake proofs
- Zero-knowledge property lost
- System security compromised

### Mitigation Strategies

1. **Large MPC with diverse participants** (100+ recommended)
2. **Verifiable setup transcript** (public audit trail)
3. **Hardware-based randomness** (HSMs, air-gapped devices)
4. **Geographically distributed participants** (reduces collusion risk)
5. **Gradual transition to transparent SNARKs** (PLONK, STARKs - no trusted setup)

### Ceremony Tooling

The repository includes a ceremony toolkit under `tools/trusted-setup/`:

- Coordinator service and CLI for managing phase1/phase2 contributions
- Participant CLI for online and air-gapped contributions
- Transcript verification command
- Documentation in `tools/trusted-setup/docs/`

Production deployments must replace the placeholder parameters in `x/veid/zk/params/` with ceremony outputs.

## Known Limitations

### 1. Trusted Setup Dependency

**Issue:** Groth16 requires trusted setup ceremony

**Impact:** Security depends on at least one honest participant

**Mitigation:**

- Multi-party ceremony with 100+ participants
- Public verifiable transcript
- Future migration to PLONK or STARKs (no trusted setup)

### 2. Circuit-Specific Keys

**Issue:** Each circuit requires separate trusted setup

**Impact:** New claim types require new ceremonies

**Mitigation:**

- Careful circuit design to minimize future changes
- Universal circuits for flexible claim types
- Upgrade path to universal SNARKs

### 3. Client-Side Computation

**Issue:** Proof generation happens on user devices

**Impact:** Mobile devices may experience slow proof generation

**Mitigation:**

- Optimize circuits for constraint count
- Offer server-side proving as optional service (privacy tradeoff)
- Progressive Web Apps for better mobile performance

### 4. BN254 Security Level

**Issue:** BN254 provides ~100-bit security, not 128-bit

**Impact:** May be insufficient for long-term secrets (50+ years)

**Mitigation:**

- Acceptable for identity claims with limited validity periods
- Future migration to BLS12-381 for 128-bit security
- Regular security reassessment

### 5. Quantum Vulnerability

**Issue:** Groth16 and BN254 are not post-quantum secure

**Impact:** Vulnerable to Shor's algorithm on quantum computers

**Mitigation:**

- Monitor quantum computing developments
- Plan migration to post-quantum ZK schemes (e.g., Ligero, Aurora)
- Identity claims have short validity periods (minimize exposure)

### 6. Proof Size vs Transparency Tradeoff

**Issue:** Groth16 has small proofs but requires trusted setup

**Impact:** Security depends on setup ceremony integrity

**Alternatives:**

- **PLONK:** Universal setup, larger proofs (~768 bytes)
- **STARKs:** Transparent (no setup), very large proofs (~100 KB)
- **Bulletproofs:** Transparent, larger proofs (~1.3 KB)

**Decision:** Groth16 chosen for smallest proof size and fastest verification

## Production Deployment

### Pre-Deployment Checklist

- [ ] Complete multi-party trusted setup ceremony (100+ participants)
- [ ] Publish and verify setup transcript
- [ ] Formal verification of circuit implementations
- [ ] Security audit by reputable cryptography firm
- [ ] Comprehensive testing on testnet (>1000 proofs)
- [ ] Performance benchmarking under load
- [ ] Documentation for client integration
- [ ] Key rotation and upgrade procedures
- [ ] Incident response plan for setup compromise
- [ ] Monitoring and alerting for proof verification failures

### Trusted Setup Ceremony Procedure

1. **Phase 1: Powers of Tau**
   - Coordinate with 100+ participants from diverse backgrounds
   - Each participant generates random contribution
   - Contributions combined using MPC protocol
   - Public transcript published for verification

2. **Phase 2: Circuit-Specific Setup**
   - For each circuit (age, residency, score):
     - Run circuit-specific ceremony
     - Generate proving and verification keys
     - Destroy toxic waste securely
   - Publish verification keys on-chain
   - Distribute proving keys to clients

3. **Verification and Auditing**
   - Independent verification of transcript
   - Reproduce setup from transcript
   - Compare generated keys with published keys
   - Security audit of setup procedure

4. **Ongoing Monitoring**
   - Monitor for anomalous proof patterns
   - Regular security audits
   - Plan for key rotation and circuit upgrades

### Client Integration

**Required Client Capabilities:**

1. Generate cryptographic nonces (32 bytes random)
2. Compute Pedersen commitments to witness values
3. Construct circuit witnesses from identity data
4. Generate Groth16 proofs using proving keys
5. Serialize proofs for blockchain submission

**Client Libraries:**

- JavaScript: `gnark-crypto-js` (WebAssembly)
- Go: `github.com/consensys/gnark`
- Rust: `ark-groth16` (alternative implementation)
- Python: `py-ecc` (for testing)

### Monitoring and Alerting

**Metrics to Track:**

- Proof verification success/failure rate
- Average proof verification time
- Distribution of proof types (age, residency, score)
- Anomalous proof patterns (potential attacks)
- Client-side proof generation errors

**Alert Thresholds:**

- Verification failure rate > 5% (investigate immediately)
- Verification time > 10ms (potential DoS)
- Unusual spike in specific proof type (potential exploit)

## References

### Academic Papers

1. **Groth16 Original Paper:**
   Jens Groth. "On the Size of Pairing-based Non-interactive Arguments." EUROCRYPT 2016.
   [https://eprint.iacr.org/2016/260](https://eprint.iacr.org/2016/260)

2. **BN Curves:**
   Paulo S. L. M. Barreto and Michael Naehrig. "Pairing-Friendly Elliptic Curves of Prime Order." SAC 2005.

3. **zk-SNARKs Overview:**
   Eli Ben-Sasson et al. "Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture." USENIX Security 2014.

4. **Trusted Setup Security:**
   Sean Bowe et al. "Scalable Multi-party Computation for zk-SNARK Parameters in the Random Beacon Model." IACR Cryptology ePrint 2017/1050.

### Implementation References

1. **gnark Library:**
   [https://github.com/consensys/gnark](https://github.com/consensys/gnark)
2. **gnark Documentation:**
   [https://docs.gnark.consensys.io/](https://docs.gnark.consensys.io/)

3. **BN254 Curve Specification:**
   [https://github.com/ethereum/EIPs/blob/master/EIPS/eip-196.md](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-196.md)

4. **Groth16 Verification in Ethereum:**
   [https://github.com/ethereum/EIPs/blob/master/EIPS/eip-197.md](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-197.md)

### Standards and Best Practices

1. **ZKProof Standards:**
   [https://zkproof.org/](https://zkproof.org/)

2. **Trusted Setup Ceremonies:**
   Zcash Powers of Tau: [https://zfnd.org/conclusion-of-the-powers-of-tau-ceremony/](https://zfnd.org/conclusion-of-the-powers-of-tau-ceremony/)

3. **Cryptographic Best Practices:**
   NIST SP 800-90A: Recommendation for Random Number Generation

---

**Document Version:** 1.0.0  
**Last Updated:** 2026-01-29  
**Maintained By:** VirtEngine Core Team  
**Security Review:** Pending (Required before mainnet)
