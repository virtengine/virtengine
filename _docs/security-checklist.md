# VirtEngine Security Checklist

**Version:** 1.0.0  
**Last Updated:** 2025-01-20  
**Maintainer:** VirtEngine Security Team

---

## Overview

This security checklist must be completed before merging any code changes that involve:

- Authentication, authorization, or identity verification logic
- Cryptographic operations or key management
- Sensitive data handling (PII, credentials, private keys)
- External API integrations or network operations
- Database queries or state mutations
- New dependencies or dependency updates
- Provider daemon infrastructure adapters
- ML inference pipelines affecting consensus

**Usage:** Copy the relevant sections to your PR description and check off items as you verify them. All applicable items must be checked before approval.

---

## Pre-Merge Security Checklist

### Input Validation

- [ ] All user inputs are validated at the entry point before processing
- [ ] Input length limits are enforced to prevent buffer overflow/DoS attacks
- [ ] Numeric inputs are checked for overflow/underflow conditions
- [ ] String inputs are sanitized to prevent injection attacks (SQL, command, path traversal)
- [ ] Protobuf message fields are validated using `ValidateBasic()` methods
- [ ] File paths are validated and canonicalized before use
- [ ] Array/slice indices are bounds-checked before access
- [ ] Regex patterns have timeout limits to prevent ReDoS attacks

### Authentication & Authorization

- [ ] All endpoints require appropriate authentication
- [ ] Authorization checks occur before any privileged operation
- [ ] Message signers are verified using `sdk.AccAddress` validation
- [ ] Module authority is checked via `x/gov` account (not hardcoded addresses)
- [ ] MFA requirements are enforced for sensitive operations (x/mfa integration)
- [ ] Session tokens have appropriate expiration times
- [ ] Failed authentication attempts are rate-limited
- [ ] Privilege escalation paths are blocked (no user-to-admin bypasses)

### Cryptography

- [ ] No custom cryptographic implementations (use established libraries)
- [ ] Encryption uses approved algorithms (X25519-XSalsa20-Poly1305 for envelopes)
- [ ] Random number generation uses `crypto/rand`, not `math/rand`
- [ ] Key sizes meet minimum requirements (256-bit for symmetric, 2048-bit for RSA)
- [ ] Nonces are unique and never reused (use counters or random generation)
- [ ] Cryptographic keys are zeroized after use (`defer memzero(key)`)
- [ ] Signature verification occurs before processing signed data
- [ ] Timing-safe comparison is used for secret comparison (`subtle.ConstantTimeCompare`)

### Data Protection

- [ ] Sensitive data is encrypted at rest using `EncryptionEnvelope` format
- [ ] PII is never logged or included in error messages
- [ ] Memory containing secrets is cleared after use
- [ ] Database queries use parameterized statements (no string concatenation)
- [ ] Sensitive fields are marked with appropriate protobuf options
- [ ] Data retention policies are enforced (automatic cleanup)
- [ ] Cross-tenant data isolation is maintained
- [ ] Backup data is encrypted with separate keys

### Error Handling

- [ ] Errors do not leak sensitive information (stack traces, internal paths)
- [ ] Error messages are generic for authentication failures (prevent enumeration)
- [ ] Panic recovery is implemented for all goroutines
- [ ] Failed operations clean up partial state (atomic transactions)

### Dependencies

- [ ] New dependencies are from trusted sources with active maintenance
- [ ] Dependencies are pinned to specific versions (not `latest` or ranges)
- [ ] No known CVEs exist for added/updated dependencies (`go mod audit`)
- [ ] Dependency licenses are compatible with project license
- [ ] Minimal dependencies are added (prefer stdlib when possible)

### Secrets Management

- [ ] Secrets are loaded from environment variables or secure vaults (not config files)
- [ ] API keys and tokens are never committed to version control
- [ ] Secrets are not passed via command-line arguments (visible in process list)
- [ ] Secret rotation is supported without service restart
- [ ] Secrets are cleared from memory immediately after use
- [ ] Default/example secrets are clearly marked as non-production

### Logging & Monitoring

- [ ] Security-relevant events are logged (auth attempts, permission changes)
- [ ] Logs do not contain sensitive data (passwords, tokens, PII)
- [ ] Log injection is prevented (user input is escaped in logs)
- [ ] Anomaly detection alerts are configured for security events
- [ ] Audit trails are immutable and tamper-evident

---

## Module-Specific Checklists

### x/veid Module (Identity Verification)

- [ ] All identity scopes use `EncryptionEnvelope` with validator key encryption
- [ ] Three-signature validation is enforced (client, user, salt binding)
- [ ] ML scoring uses deterministic configuration (`ForceCPU: true`, `RandomSeed: 42`)
- [ ] Biometric templates are never stored in plaintext
- [ ] Scope upload rate limits prevent enumeration attacks
- [ ] Identity verification failures are logged without PII
- [ ] Expired identity proofs are automatically invalidated
- [ ] Cross-scope correlation is prevented (different salts per scope)

### x/mfa Module (Multi-Factor Authentication)

- [ ] TOTP secrets are encrypted in storage
- [ ] TOTP verification uses time-constant comparison
- [ ] Backup codes are hashed (not stored in plaintext)
- [ ] MFA bypass is not possible via API manipulation
- [ ] Recovery flows require additional verification
- [ ] Failed MFA attempts trigger account lockout
- [ ] MFA enrollment requires initial authentication
- [ ] Device registration includes attestation verification

### x/encryption Module

- [ ] Key derivation uses approved KDF (Argon2id, HKDF)
- [ ] Envelope format includes algorithm identifier for future migration
- [ ] Recipient fingerprint is verified before decryption
- [ ] Ciphertext integrity is verified (Poly1305 tag)
- [ ] Key rotation is supported without data loss
- [ ] Decryption failures are logged (potential tampering)
- [ ] Public keys are validated (on curve, not identity)
- [ ] Key export requires additional authorization

### x/market Module (Marketplace)

- [ ] Bid amounts are validated against escrow balance
- [ ] Order cancellation requires owner signature
- [ ] Lease creation atomically locks escrow funds
- [ ] Price calculations prevent overflow/underflow
- [ ] Rate limiting prevents bid spam attacks
- [ ] Expired orders are cleaned up automatically
- [ ] Cross-order manipulation is prevented (isolation)
- [ ] Provider reputation cannot be artificially inflated

### Provider Daemon (pkg/provider_daemon)

- [ ] Provider keys support hardware/ledger storage
- [ ] Kubernetes adapter uses namespace isolation
- [ ] Ansible vault passwords are cleared after use
- [ ] Usage metering is tamper-evident
- [ ] Manifest validation prevents resource exhaustion
- [ ] Workload state transitions follow defined FSM
- [ ] Infrastructure credentials use short-lived tokens
- [ ] Pod security policies are enforced (no privileged containers)
- [ ] Network policies restrict cross-tenant communication
- [ ] Secrets are mounted as tmpfs (not persistent volumes)

---

## Release Security Checklist

Complete before any mainnet release:

### Pre-Release Verification

- [ ] All security-related PRs have been reviewed by security team
- [ ] Dependency audit shows no high/critical CVEs
- [ ] Static analysis (gosec, semgrep) shows no new findings
- [ ] Dynamic analysis/fuzzing completed for critical paths
- [ ] Upgrade migration path tested on testnet
- [ ] Rollback procedure documented and tested

### Release Artifacts

- [ ] Release binaries are reproducibly built
- [ ] Binaries are signed with release keys
- [ ] Checksums are published via secure channel
- [ ] Docker images use minimal base (distroless/scratch)
- [ ] No development/debug tools in production images

### Post-Release

- [ ] Security monitoring is active for new version
- [ ] Incident response team is on standby
- [ ] Rollback procedure is ready to execute
- [ ] Communication channels are monitored for reports

---

## Security Testing Requirements

### Before Feature Merge

| Test Type | Requirement | Tool |
|-----------|-------------|------|
| Unit Tests | 80%+ coverage for security-critical code | `go test -cover` |
| Input Fuzzing | All parsers and validators | `go-fuzz`, `tests/fuzz/` |
| Static Analysis | No high/critical findings | `golangci-lint`, `gosec` |
| Dependency Scan | No known CVEs | `go mod audit`, Dependabot |

### Before Release

| Test Type | Requirement | Tool |
|-----------|-------------|------|
| Penetration Testing | Critical/High findings remediated | External auditor |
| Consensus Testing | Determinism verified across nodes | `tests/consensus/` |
| Upgrade Testing | Clean migration on testnet | `scripts/upgrade-test.sh` |
| Load Testing | No degradation under stress | `tests/load/` |
| Chaos Testing | Recovery from failures | `tests/chaos/` |

### Continuous

| Test Type | Frequency | Tool |
|-----------|-----------|------|
| Dependency Monitoring | Daily | Dependabot, Snyk |
| Container Scanning | Per build | Trivy, Grype |
| Secret Scanning | Per commit | git-secrets, TruffleHog |
| SAST | Per PR | CodeQL, Semgrep |

---

## Quick Reference

### Approved Cryptographic Algorithms

| Purpose | Algorithm | Notes |
|---------|-----------|-------|
| Symmetric Encryption | XSalsa20-Poly1305 | For envelope payloads |
| Key Exchange | X25519 | Ephemeral keys only |
| Hashing | SHA-256, BLAKE2b | SHA-256 for Cosmos compatibility |
| Signatures | Ed25519, secp256k1 | secp256k1 for Cosmos accounts |
| KDF | Argon2id, HKDF | Argon2id for passwords, HKDF for key derivation |

### Security Contacts

- **Security Issues:** security@virtengine.io
- **Bug Bounty:** https://virtengine.io/security/bounty
- **Responsible Disclosure:** See `SECURITY.md`

---

*This checklist is a living document. Submit improvements via PR to `_docs/security-checklist.md`.*
