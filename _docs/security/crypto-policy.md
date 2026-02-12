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

## Legacy Algorithm Exceptions
Legacy exceptions are approved for compatibility with third-party standards or services. Any new exception must follow the approval process in the Exceptions section below.

### SHA-1 in SAML/XML Signatures (eduGAIN Federation)

**Files affected:**
- `pkg/edugain/verification.go` (lines 633-639)

**Justification:** SAML 2.0 XML signature specification defines SHA-1 algorithm URIs. These constants are required to **detect and reject** weak algorithms in incoming SAML metadata. We explicitly validate against these URIs to ensure SHA-256 or stronger is used.

**Mitigation:**
- Constants defined for detection/rejection only
- Not used for signature generation
- Validation explicitly rejects SHA-1 signatures
- Marked with `//nolint:gosec // G401: SHA-1 constants required by SAML 2.0 spec - rejected during validation`

**Migration plan:** SAML 2.0 successor (if any) that removes SHA-1 URIs from spec → remove these constants.

### HMAC-SHA1 in Twilio Webhooks

**Files affected:**
- `pkg/verification/sms/gateway.go` (line 497)

**Justification:** Twilio webhook signature validation uses HMAC-SHA1 as defined by their API. This is a third-party service requirement, not a design choice.

**Mitigation:**
- Isolated to webhook signature validation only
- Not used for any internal cryptographic operations
- Marked with `//nolint:gosec // G401: HMAC-SHA1 required by Twilio webhook API`

**Migration plan:** When Twilio provides HMAC-SHA256 webhook signatures → update implementation.

### TOTP SHA-1 (Deprecated as of v0.9.x)

**Previous default:**
- `x/mfa/keeper/verification.go` used SHA-1 as default TOTP algorithm

**Change (v0.9.x):**
- Default changed to SHA-256
- Existing SHA-1 enrollments continue to work (algorithm stored per enrollment)
- New enrollments use SHA-256 by default

**Migration for users:**
```go
// Old default (< v0.9.x)
DefaultTOTPConfig() // Returned SHA-1

// New default (≥ v0.9.x)
DefaultTOTPConfig() // Returns SHA-256

// Explicit SHA-1 for backward compatibility (not recommended)
config := TOTPConfig{
    Algorithm: "SHA1", // Only for legacy systems
}
```

**Keycloak realm configuration:**
- `config/waldur/keycloak/realm.json` uses `HmacSHA256` as of February 12, 2026.

### MD5 for Non-Security Checksums

MD5 is **prohibited** for security purposes. For legacy file checksums where collision resistance is not a security requirement, use SHA-256 instead.

**Updated tooling (February 12, 2026):**
- `scripts/localnet.sh` uses `sha256sum` for change detection.
- Runbooks now use `sha256sum` for deterministic hash checks.
- Synthetic dataset generation uses SHA-256 for deterministic image seeds.

## Migration Timeline

**February 12, 2026**
- Replace MD5-based checksums in localnet tooling with SHA-256.
- Replace MD5-based deterministic seeds in ML synthetic dataset generation with SHA-256.
- Replace SHA-1 based deterministic UUIDs in E2E tests with SHA-256 based derivation.
- Update Keycloak TOTP configuration to HMAC-SHA256.

**June 30, 2026**
- Remove any remaining MD5-based checksum usage in scripts or runbooks.

**December 31, 2026**
- Re-evaluate third-party integrations that require SHA-1 (e.g., Twilio webhooks).
- Migrate to HMAC-SHA256 when the provider supports it.

**Ongoing**
- SAML SHA-1 algorithm URIs remain defined only to detect and reject weak signatures until the SAML spec removes them.

## Exceptions
Exceptions require approval from security maintainers and must include:
- Reason for the exception
- Compensating controls
- Expiry date and owner for remediation

## References
- Go `crypto` package documentation
- VirtEngine Security Architecture (`docs/security/SECURITY_ARCHITECTURE.md`)

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2026-02-12 | 1.2 | Documented legacy exceptions and deprecation timeline for weak crypto |
| 2026-02-10 | 1.1 | Updated legacy exceptions with file paths; documented TOTP SHA-1 → SHA-256 migration |
| 2026-02-06 | 1.0 | Initial policy document |
