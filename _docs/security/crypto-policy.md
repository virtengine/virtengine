# Cryptography Policy

## Purpose
This policy defines approved cryptographic algorithms, key sizes, protocols, and implementation requirements for VirtEngine. It applies to all code in this repository, including off-chain services and tooling.

## Scope
- On-chain module code (`x/`)
- Off-chain services (`pkg/`)
- CLI tools (`cmd/`, `sdk/`)
- Infrastructure and test tooling when handling sensitive data

## Approved Algorithms

### Symmetric Encryption
- AES-256-GCM
- ChaCha20-Poly1305

### Asymmetric Encryption / Key Agreement
- X25519 for key agreement
- P-256 (ECDH) only when interoperability requires it

### Digital Signatures
- Ed25519 (preferred)
- secp256k1 (Cosmos SDK compatibility)
- P-256 only when interoperability requires it

### Hashing
- SHA-256 / SHA-384 / SHA-512
- BLAKE2b or BLAKE2s when performance or compatibility requires it

### Message Authentication
- HMAC-SHA-256 / HMAC-SHA-512

### Password Hashing
- Argon2id (preferred)
- scrypt only for legacy compatibility

### Randomness
- Use `crypto/rand` for all security-sensitive randomness
- `math/rand` is permitted only for tests, simulations, or non-security uses

## Disallowed Algorithms
- MD5
- SHA-1
- RSA with keys < 2048 bits
- RC4, DES, 3DES
- ECB mode

## TLS / HTTP Security
- Minimum TLS version: 1.2 (1.3 preferred)
- Disable insecure renegotiation
- Use strong cipher suites aligned with the Go standard library defaults
- Always verify TLS certificates; `InsecureSkipVerify` is prohibited
- Require explicit timeouts for all outbound HTTP clients

## Key Management
- Store secrets only in secure stores (KMS, HSM, secrets manager, or OS keychain)
- Never hardcode credentials, API keys, or private keys
- Rotate long-lived keys at least every 12 months or on suspected compromise

## Implementation Requirements
- Use the helpers in `pkg/security/` for random generation, safe command execution, path validation, and HTTP defaults
- All `//nolint` or `//nosec` annotations must include a justification comment
- Any crypto-related change must include a security review checklist in the PR description

## Exceptions
Exceptions require approval from security maintainers and must include:
- Reason for the exception
- Compensating controls
- Expiry date and owner for remediation

## References
- Go `crypto` package documentation
- VirtEngine Security Architecture (`docs/security/SECURITY_ARCHITECTURE.md`)
