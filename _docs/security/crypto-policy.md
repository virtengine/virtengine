# VirtEngine Cryptography Policy

This document defines the cryptographic standards and policies for VirtEngine codebase.

## Approved Algorithms

### Hashing

| Algorithm | Status | Use Case |
|-----------|--------|----------|
| SHA-256 | ✅ Approved | General-purpose hashing, content addressing |
| SHA-384 | ✅ Approved | High-security hashing |
| SHA-512 | ✅ Approved | High-security hashing |
| SHA3-256 | ✅ Approved | Modern applications requiring SHA-3 |
| BLAKE2b | ✅ Approved | High-performance hashing |
| SHA-1 | ⚠️ Legacy Only | SAML 2.0 XML Encryption (see exceptions) |
| MD5 | ⛔ Prohibited | Not to be used for any security purpose |

### Symmetric Encryption

| Algorithm | Status | Use Case |
|-----------|--------|----------|
| AES-256-GCM | ✅ Approved | Symmetric encryption with authentication |
| AES-128-GCM | ✅ Approved | Symmetric encryption with authentication |
| ChaCha20-Poly1305 | ✅ Approved | Alternative to AES-GCM |
| XSalsa20-Poly1305 | ✅ Approved | Used in VEID envelope encryption |
| AES-CBC | ⚠️ Legacy Only | Only with HMAC authentication |

### Asymmetric Encryption

| Algorithm | Status | Use Case |
|-----------|--------|----------|
| X25519 | ✅ Approved | Key exchange (ECDH) |
| Ed25519 | ✅ Approved | Digital signatures |
| RSA-4096 | ✅ Approved | Legacy systems requiring RSA |
| RSA-2048 | ⚠️ Transitional | Migrate to 4096-bit or Ed25519 |
| ECDSA P-256/P-384 | ✅ Approved | Digital signatures |

### Key Derivation

| Algorithm | Status | Use Case |
|-----------|--------|----------|
| HKDF-SHA256 | ✅ Approved | Key derivation |
| Argon2id | ✅ Approved | Password hashing |
| scrypt | ✅ Approved | Password hashing |
| PBKDF2-SHA256 | ⚠️ Legacy Only | With iteration count ≥ 100,000 |

### TLS Configuration

| Version | Status |
|---------|--------|
| TLS 1.3 | ✅ Preferred |
| TLS 1.2 | ✅ Minimum Required |
| TLS 1.1 | ⛔ Prohibited |
| TLS 1.0 | ⛔ Prohibited |
| SSL 3.0 | ⛔ Prohibited |

#### Approved TLS Cipher Suites (TLS 1.2)

```
TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
```

## Random Number Generation

### Security-Sensitive Operations

**MUST use `crypto/rand`:**
- Key generation
- Nonce/IV generation
- Token generation (API keys, session tokens, CSRF tokens)
- Salt generation for password hashing
- Any value that must be unpredictable

**See:** `pkg/security/random.go` for approved utilities.

### Non-Security Operations

**MAY use `math/rand`:**
- Simulation code (with documented seed for reproducibility)
- Load balancing (when unpredictability is not required)
- Test data generation
- Jitter where prediction has no security impact

**Requirements:**
- Add `//nolint:gosec` with justification comment
- Document why `crypto/rand` is not required

### Consensus-Critical Code

For blockchain consensus:
- Use deterministic sources only
- Document seed sources explicitly
- See `x/enclave/keeper/committee.go` for patterns

## Legacy Algorithm Exceptions

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

## Security Utilities

### HTTP Clients

Always use `pkg/security/httpclient.go`:

```go
import "github.com/virtengine/virtengine/pkg/security"

client := security.NewSecureHTTPClient()
// or for TLS 1.3 only:
client := security.NewSecureHTTPClientTLS13()
```

### Random Generation

Always use `pkg/security/random.go`:

```go
import "github.com/virtengine/virtengine/pkg/security"

token, err := security.SecureRandomToken(32)
id, err := security.SecureRandomID(16)
n, err := security.SecureRandomInt(100)
```

### Integer Overflow Protection

Use `pkg/security/overflow.go` for safe conversions:

```go
import "github.com/virtengine/virtengine/pkg/security"

v, err := security.SafeInt64(sdkMathInt)
v, err := security.SafeUint64ToInt(uint64Value)
```

### Path Validation

Use `pkg/security/path.go` for file operations:

```go
import "github.com/virtengine/virtengine/pkg/security"

validator := security.NewPathValidator(baseDir)
err := validator.ValidatePath(userInput)

// Or for simple CLI paths:
data, err := security.SafeReadFile(path)
```

### Command Execution

Use `pkg/security/validation.go` for external commands:

```go
import "github.com/virtengine/virtengine/pkg/security"

err := security.ValidateExecutable("ansible", path)
args, err := security.PingArgs(target, count)
```

## Code Review Checklist

When reviewing code for cryptographic security:

- [ ] No hardcoded secrets or keys
- [ ] TLS configuration uses minimum TLS 1.2
- [ ] HTTP clients have proper timeouts
- [ ] Random generation uses `crypto/rand` for security purposes
- [ ] Integer conversions check for overflow
- [ ] File paths are validated before use
- [ ] External commands use allowlisted executables
- [ ] Sensitive data is cleared from memory after use

## Gosec Alert Handling

### Legitimate Security Issues

Fix immediately:
- G101: Hardcoded credentials (if actual secrets)
- G201: SQL query construction using format strings
- G301: Poor file permissions
- G304: File path from variable (path traversal)
- G401: Use of weak crypto (MD5, SHA1) for security
- G402: TLS InsecureSkipVerify
- G501: Import of deprecated crypto package

### Documented Exceptions

Add `//nolint:gosec` with justification:
- G103: Unsafe pointer (if reviewed and necessary)
- G104: Unaudited defer Close (if error handling not critical)
- G110: Decompression bomb (if input is trusted)
- G306: Write file with broader permissions (if intentional)
- G404: Weak random (if not security-sensitive)

### False Positives

Update `.golangci.yaml` exclusions:
- G101 for config struct field names (not actual secrets)
- Generated protobuf code
- Test fixtures with fake credentials

## Contact

For questions about cryptographic policy, contact the security team or open an issue with the `security` label.

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2026-02-10 | 1.1 | Updated legacy exceptions with file paths; documented TOTP SHA-1 → SHA-256 migration |
| 2026-02-06 | 1.0 | Initial policy document |
