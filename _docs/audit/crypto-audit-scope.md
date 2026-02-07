# VirtEngine Cryptography Audit Scope

## Overview
Review of cryptographic implementations in VirtEngine for correctness, security,
protocol soundness, and compliance with best practices. The audit should focus
on confidentiality, integrity, and authentication guarantees across all
cryptographic boundaries.

## Objectives
- Validate cryptographic protocol design and message flows.
- Verify correct library usage and parameter choices.
- Confirm key lifecycle, storage, and zeroization controls.
- Assess side-channel exposure and timing risks.
- Identify implementation bugs or unsafe assumptions.

## In-Scope Components

### VEID Encryption (pkg/veid/crypto/)
- X25519 key exchange and peer identity binding.
- XSalsa20-Poly1305 envelope format and metadata validation.
- HKDF-based key derivation and context separation.
- Nonce generation and reuse safeguards.

### HSM Integration (pkg/keymanagement/hsm/)
- PKCS#11 session management and key handles.
- Key generation, import, and rotation flows.
- Signing operations and audit logging.
- Error handling and fallback logic.

### TLS / mTLS (pkg/mtls/)
- Certificate generation, renewal, and chain validation.
- Revocation handling and certificate pinning where applicable.
- Client authentication flows.

### Signature Schemes
- Ed25519 and Secp256k1 usage in signatures.
- BLS (if enabled) threshold signing semantics.
- Domain separation tags and context binding.

## Specific Reviews
1. RNG sources and entropy assumptions.
2. Key storage, rotation, and secure deletion.
3. Protocol handshake transcripts and replay protection.
4. Timing attack resistance in comparisons and validation.
5. Envelope malleability checks and ciphertext validation.

## Standards Alignment
- NIST SP 800-57 (Key Management Guidelines).
- NIST SP 800-131A (Cryptographic Standards and Key Sizes).
- FIPS 140-2 considerations for HSM boundaries (if applicable).
- OWASP ASVS cryptography requirements (relevant sections).

## Out of Scope
- Pure dependency analysis of upstream crypto libraries.
- Hardware physical security testing.
- General infrastructure security (covered elsewhere).

## Deliverables
- Cryptographic design review report.
- Implementation audit findings with severity ratings.
- Protocol analysis and exploitability assessment.
- Compliance gap analysis (if any).
- Retest verification and closure statement.

## Timeline
- Estimated duration: 4 weeks.
- Draft report by week 3.
- Final report and retest by week 4.

## Points of Contact
- Crypto Lead: security@virtengine.io
- Key Management Owner: keymanagement@virtengine.io
- Audit Coordinator: audit@virtengine.io
