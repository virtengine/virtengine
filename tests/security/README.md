# VirtEngine Security Test Suite

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-800

---

## Overview

This directory contains security tests for the VirtEngine blockchain platform. These tests are designed to verify the security of cryptographic operations, signature validation, MFA enforcement, and input validation.

## Test Categories

### 1. Cryptography Tests (`crypto_test.go`)
- Encryption envelope validation
- Key fingerprint computation
- Nonce uniqueness
- Algorithm compliance
- Multi-recipient encryption
- Key rotation procedures

### 2. Signature Tests (`signature_test.go`)
- Identity upload signature verification
- Approved client signature checks
- Signature forgery detection
- Replay attack prevention
- Malformed signature handling

### 3. MFA Enforcement Tests (`mfa_enforcement_test.go`)
- Sensitive transaction gating
- Challenge verification
- Session management
- Device trust validation
- Recovery flow enforcement

### 4. Input Validation Tests (`input_validation_test.go`)
- Malformed payload handling
- Overflow detection
- Injection prevention
- Size limit enforcement
- Encoding validation

### 5. Key Rotation Tests (`key_rotation_test.go`)
- Provider daemon key rotation
- Approved client key rotation
- Validator key rotation
- User account key rotation

## Running Tests

```bash
# Run all security tests
go test -v ./tests/security/...

# Run with race detection
go test -race -v ./tests/security/...

# Run specific test category
go test -v ./tests/security/... -run TestCrypto
go test -v ./tests/security/... -run TestSignature
go test -v ./tests/security/... -run TestMFA
go test -v ./tests/security/... -run TestInputValidation
go test -v ./tests/security/... -run TestKeyRotation

# Run with coverage
go test -cover -coverprofile=coverage.out ./tests/security/...
```

## Security Warnings

> ⚠️ **NEVER** include actual private keys, secrets, or credentials in tests  
> ⚠️ All test keys must be generated fresh for each test run  
> ⚠️ Do not commit any files containing real cryptographic material  
> ⚠️ Use deterministic seeds only for reproducible test vectors, never for production

## Test Data

Test vectors and fixtures are stored in `testdata/` subdirectory. These include:
- Sample encrypted envelopes
- Invalid envelope formats
- Malformed signatures
- Boundary condition payloads

## Continuous Integration

Security tests are run as part of CI on every PR:
1. Static analysis (gosec, staticcheck)
2. Dependency scanning (govulncheck)
3. Unit security tests
4. Fuzz testing (for input validation)

## Audit Preparation

These tests support security audit readiness by demonstrating:
- Consistent cryptographic envelope handling
- Key rotation procedures for all key types
- MFA enforcement for all sensitive transactions
- Input validation and error handling
- Replay protection mechanisms
