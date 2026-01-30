# VirtEngine Security Test Suite

**Version:** 2.0.0  
**Date:** 2026-01-30  
**Task Reference:** SECURITY-005

---

## Overview

This directory contains the comprehensive security testing suite for the VirtEngine blockchain platform, implementing the penetration testing program defined in `PENETRATION_TESTING_PROGRAM.md`.

## Directory Structure

```
tests/security/
├── README.md                    # This file
├── VULNERABILITY_TEMPLATE.md    # Vulnerability tracking template
├── blockchain/                  # Blockchain-specific security tests
│   ├── consensus_test.go        # BC-001, BC-002: Consensus attacks
│   ├── replay_test.go           # BC-003, BC-004: Replay/malleability
│   ├── state_machine_test.go    # BC-005, BC-006: State/authority
│   └── crypto_test.go           # BC-007, BC-008: Cryptographic tests
├── api/                         # API security tests
│   ├── auth_test.go             # API-AUTH-*: Authentication tests
│   ├── authz_test.go            # API-AUTHZ-*: Authorization tests
│   └── injection_test.go        # WIN-*: Injection tests
├── fuzz/                        # Fuzzing harnesses
│   └── fuzz_test.go             # Envelope, signature, message fuzzers
├── configs/                     # Security tool configurations
│   └── nuclei/                  # Nuclei scanning templates
│       ├── virtengine-auth-bypass.yaml
│       ├── virtengine-security-headers.yaml
│       ├── virtengine-info-disclosure.yaml
│       ├── virtengine-grpc-reflection.yaml
│       └── virtengine-rate-limit.yaml
└── scripts/                     # Security testing scripts
    ├── scan_dependencies.sh     # Dependency vulnerability scan
    ├── static_analysis.sh       # Static security analysis
    └── secret_scan.sh           # Secret detection
```

## Test Categories

### 1. Blockchain Security Tests (`blockchain/`)

| Test File | Attack IDs | Description |
|-----------|------------|-------------|
| `consensus_test.go` | BC-001, BC-002 | Byzantine fault tolerance, consensus stall |
| `replay_test.go` | BC-003, BC-004 | Transaction replay, malleability |
| `state_machine_test.go` | BC-005, BC-006 | State transitions, authority bypass |
| `crypto_test.go` | BC-007, BC-008 | Encryption, signature security |

### 2. API Security Tests (`api/`)

| Test File | Attack IDs | Description |
|-----------|------------|-------------|
| `auth_test.go` | API-AUTH-*, API-RATE-* | Authentication, rate limiting |
| `authz_test.go` | API-AUTHZ-* | Authorization, IDOR, parameter tampering |
| `injection_test.go` | WIN-*, XSS | SQL, NoSQL, command, template injection |

### 3. Fuzzing (`fuzz/`)

| Fuzzer | Target | Description |
|--------|--------|-------------|
| `FuzzEnvelopeEncryption` | Envelope parsing | Encryption envelope validation |
| `FuzzSignatureVerification` | Signature logic | Signature verification bypass |
| `FuzzMessageValidation` | Message handling | Message validation logic |
| `FuzzProtoDecoding` | Protobuf | Proto wire format parsing |
| `FuzzAddressParsing` | Address validation | Bech32 address parsing |
| `FuzzSaltValidation` | Salt binding | Capture protocol salt validation |

## Running Tests

### Security Unit Tests

```bash
# Run all security tests (requires 'security' build tag)
go test -v -tags="security" ./tests/security/...

# Run specific category
go test -v -tags="security" ./tests/security/blockchain/...
go test -v -tags="security" ./tests/security/api/...

# Run with race detection
go test -race -v -tags="security" ./tests/security/...
```

### Fuzzing

```bash
# Run envelope fuzzer for 60 seconds
go test -fuzz=FuzzEnvelopeEncryption -fuzztime=60s ./tests/security/fuzz/

# Run signature fuzzer for 60 seconds
go test -fuzz=FuzzSignatureVerification -fuzztime=60s ./tests/security/fuzz/

# Run all fuzzers for extended period (CI)
go test -fuzz=Fuzz -fuzztime=10m ./tests/security/fuzz/
```

### Automated Scanning

```bash
# Dependency vulnerability scan
./tests/security/scripts/scan_dependencies.sh

# Static security analysis
./tests/security/scripts/static_analysis.sh

# Secret detection
./tests/security/scripts/secret_scan.sh

# Nuclei scanning (external targets)
nuclei -t ./tests/security/configs/nuclei/ -u https://api.virtengine.io
```

## Test Coverage by Attack Scenario

| Attack ID | Test Location | Status |
|-----------|---------------|--------|
| BC-001 | blockchain/consensus_test.go | ✅ Implemented |
| BC-002 | blockchain/consensus_test.go | ✅ Implemented |
| BC-003 | blockchain/replay_test.go | ✅ Implemented |
| BC-004 | blockchain/replay_test.go | ✅ Implemented |
| BC-005 | blockchain/state_machine_test.go | ✅ Implemented |
| BC-006 | blockchain/state_machine_test.go | ✅ Implemented |
| BC-007 | blockchain/crypto_test.go | ✅ Implemented |
| BC-008 | blockchain/crypto_test.go | ✅ Implemented |
| API-AUTH-* | api/auth_test.go | ✅ Implemented |
| API-AUTHZ-* | api/authz_test.go | ✅ Implemented |
| API-RATE-* | api/auth_test.go | ✅ Implemented |
| WIN-* | api/injection_test.go | ✅ Implemented |

## Security Warnings

> ⚠️ **NEVER** include actual private keys, secrets, or credentials in tests  
> ⚠️ All test keys must be generated fresh for each test run  
> ⚠️ Do not commit any files containing real cryptographic material  
> ⚠️ Use deterministic seeds only for reproducible test vectors, never for production

## CI/CD Integration

Security tests are integrated into the CI/CD pipeline:

1. **PR Checks** (runs on every PR):
   - Static analysis (gosec, staticcheck)
   - Dependency scanning (govulncheck)
   - Fast security unit tests
   - Secret scanning (gitleaks)

2. **Nightly** (runs daily):
   - Full security test suite
   - Extended fuzzing (10+ minutes)
   - Nuclei scanning on staging

3. **Release** (before each release):
   - Complete penetration test
   - Manual security review
   - Third-party audit coordination

## Vulnerability Tracking

Use `VULNERABILITY_TEMPLATE.md` for tracking discovered vulnerabilities. Key fields:
- Finding ID: VE-YYYY-XXXX
- Severity: CRITICAL/HIGH/MEDIUM/LOW
- CVSS Score and Vector
- Status tracking through remediation lifecycle

## Related Documentation

- `PENETRATION_TESTING_PROGRAM.md` - Full penetration testing program
- `SECURITY_SCOPE.md` - Security audit scope definition
- `SECURITY_AUDIT_GAP_ANALYSIS.md` - Module security analysis
- `_docs/threat-model.md` - Threat modeling documentation
