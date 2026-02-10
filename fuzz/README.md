# VirtEngine Fuzz Testing

This directory contains fuzz testing infrastructure for VirtEngine, implementing comprehensive
fuzzing coverage for input validation, parsing, and cryptographic operations.

**Task Reference:** QUALITY-001 - Fuzz Testing Implementation

## Overview

Fuzz testing (fuzzing) is an automated software testing technique that involves providing
invalid, unexpected, or random data as inputs to discover bugs, security vulnerabilities,
and edge cases that traditional testing might miss.

VirtEngine implements fuzz testing using:
- **Go's native fuzzing** (Go 1.18+) for native fuzz tests
- **OSS-Fuzz integration** for continuous fuzzing via Google's infrastructure
- **libFuzzer** for coverage-guided fuzzing with sanitizers

## Fuzz Test Coverage

### Cryptographic Operations (`x/encryption/crypto/`)
| Fuzz Target | Description |
|-------------|-------------|
| `FuzzCreateEnvelope` | Envelope creation with arbitrary plaintext |
| `FuzzOpenEnvelope` | Decryption with corrupted ciphertext |
| `FuzzNonceUniqueness` | Nonce uniqueness across encryptions |
| `FuzzMultiRecipientEnvelope` | Multi-recipient envelope handling |
| `FuzzInvalidKeySize` | Invalid key size handling |
| `FuzzKeyPairGeneration` | Key pair generation validation |
| `FuzzAlgorithmEncryption` | Algorithm interface implementation |

### Encryption Types (`x/encryption/types/`)
| Fuzz Target | Description |
|-------------|-------------|
| `FuzzEnvelopeValidate` | EncryptedPayloadEnvelope validation |
| `FuzzEnvelopeSigningPayload` | Signing payload generation |
| `FuzzEnvelopeHash` | Hash computation |
| `FuzzEnvelopeDeterministicBytes` | Deterministic serialization |
| `FuzzRecipientKeyRecordValidate` | Recipient key record validation |
| `FuzzComputeKeyFingerprint` | Key fingerprint computation |
| `FuzzAlgorithmValidation` | Algorithm parameter validation |

### VEID Scope Validation (`x/veid/types/`)
| Fuzz Target | Description |
|-------------|-------------|
| `FuzzIdentityScopeValidate` | Identity scope validation |
| `FuzzScopeTypeValidation` | Scope type validation |
| `FuzzVerificationStatusValidation` | Verification status validation |
| `FuzzVerificationStatusTransitions` | Status state machine |
| `FuzzUploadMetadataValidate` | Upload metadata validation |
| `FuzzVerificationEventValidate` | Verification event validation |
| `FuzzSimpleVerificationResultValidate` | Verification result validation |
| `FuzzApprovedClientValidate` | Approved client validation |
| `FuzzComputeSaltHash` | Salt hash computation |

### Marketplace Validation (`x/market/types/marketplace/`)
| Fuzz Target | Description |
|-------------|-------------|
| `FuzzOrderValidate` | Order validation |
| `FuzzOrderIDValidate` | Order ID parsing and validation |
| `FuzzOrderStateTransitions` | Order state machine |
| `FuzzOfferingValidate` | Offering validation |
| `FuzzOfferingIDValidate` | Offering ID validation |
| `FuzzAllocationIDValidate` | Allocation ID validation |
| `FuzzIdentityRequirementValidate` | Identity requirement validation |
| `FuzzPricingInfoValidate` | Pricing info validation |
| `FuzzOrderSetState` | Order state transition logic |
| `FuzzOrderCanAcceptBid` | Bid acceptance logic |

### Provider Daemon (`pkg/provider_daemon/`)

> **Note:** These fuzz targets are currently disabled due to pre-existing compilation issues
> in the provider_daemon and related packages. They will be enabled once those issues are resolved.

| Fuzz Target | Description | Status |
|-------------|-------------|--------|
| `FuzzManifestParse` | Manifest JSON parsing | Disabled |
| `FuzzManifestValidate` | Manifest validation | Disabled |
| `FuzzServiceSpecValidation` | Service specification validation | Disabled |
| `FuzzPortSpecValidation` | Port specification validation | Disabled |
| `FuzzVolumeSpecValidation` | Volume specification validation | Disabled |
| `FuzzNetworkSpecValidation` | Network specification validation | Disabled |
| `FuzzHealthCheckSpecValidation` | Health check validation | Disabled |
| `FuzzConstraintsValidation` | Deployment constraints validation | Disabled |
| `FuzzManifestTotalResources` | Resource calculation | Disabled |

### MFA Types (`x/mfa/types/`)
| Fuzz Target | Description |
|-------------|-------------|
| `FuzzFactorTypeFromString` | Factor type string parsing |
| `FuzzFactorTypeProperties` | Factor type properties |
| `FuzzFactorEnrollmentValidate` | Factor enrollment validation |
| `FuzzMFAPolicyValidate` | MFA policy validation |
| `FuzzFactorCombinationIsSatisfiedBy` | Factor combination satisfaction |
| `FuzzChallengeValidate` | Challenge validation |
| `FuzzAuthorizationSessionValidate` | Authorization session validation |
| `FuzzTrustedDeviceValidate` | Trusted device validation |
| `FuzzComputeFactorFingerprint` | Factor fingerprint computation |
| `FuzzComputeDeviceFingerprint` | Device fingerprint computation |
| `FuzzComputeChallengeID` | Challenge ID computation |
| `FuzzMsgEnrollFactorValidateBasic` | Message validation |
| `FuzzGenesisStateValidate` | Genesis state validation |

### Enclave Types (`x/enclave/types/`)

> **Note:** These fuzz targets are currently disabled due to pre-existing compilation issues
> in the enclave package (merge conflicts and duplicate type declarations). They will be 
> enabled once those issues are resolved.

| Fuzz Target | Description | Status |
|-------------|-------------|--------|
| `FuzzParseSGXDCAPQuoteV3` | SGX DCAP quote parsing | Disabled |
| `FuzzParseSEVSNPReport` | SEV-SNP report parsing | Disabled |
| `FuzzAddMeasurementProposalValidateBasic` | Add measurement proposal validation | Disabled |
| `FuzzRevokeMeasurementProposalValidateBasic` | Revoke measurement proposal validation | Disabled |
| `FuzzTEETypeValidation` | TEE type validation | Disabled |
| `FuzzEnclaveIdentityValidate` | Enclave identity validation | Disabled |
| `FuzzAttestationValidate` | Attestation validation | Disabled |
| `FuzzMeasurementAllowlistEntry` | Measurement allowlist validation | Disabled |
| `FuzzMsgRegisterEnclaveValidateBasic` | Register enclave message validation | Disabled |
| `FuzzMsgSubmitAttestationValidateBasic` | Submit attestation message validation | Disabled |

## Running Fuzz Tests

### Local Development

Run fuzz tests using Go's native fuzzing:

```bash
# Run all fuzz tests for a package (30 seconds each)
go test -fuzz=. -fuzztime=30s ./x/encryption/types/...
go test -fuzz=. -fuzztime=30s ./x/encryption/crypto/...
go test -fuzz=. -fuzztime=30s ./x/veid/types/...
go test -run='^$' -fuzz=. -fuzztime=30s ./x/market/types/marketplace
go test -fuzz=. -fuzztime=30s ./x/mfa/types/...
# Note: enclave and provider_daemon packages currently have compilation issues

# Run a specific fuzz test
go test -fuzz=FuzzEnvelopeValidate -fuzztime=60s ./x/encryption/types/...

# Run fuzz tests for longer (recommended for thorough testing)
go test -fuzz=. -fuzztime=10m ./x/encryption/types/...

# Run with verbose output
go test -v -fuzz=FuzzManifestParse -fuzztime=30s ./pkg/provider_daemon/...
```

### Continuous Fuzzing via OSS-Fuzz

VirtEngine is integrated with [OSS-Fuzz](https://google.github.io/oss-fuzz/) for continuous
fuzzing. The integration runs 24/7 on Google's infrastructure with multiple sanitizers.

**Local OSS-Fuzz testing:**

```bash
# Clone OSS-Fuzz
git clone https://github.com/google/oss-fuzz.git
cd oss-fuzz

# Build fuzz targets locally
python infra/helper.py build_fuzzers virtengine

# Run a specific fuzzer
python infra/helper.py run_fuzzer virtengine fuzz_envelope_validate

# Run with corpus
python infra/helper.py run_fuzzer virtengine fuzz_manifest_parse \
    -d /path/to/virtengine/fuzz/corpus/manifest
```

## Corpus Management

### Seed Corpus

Initial seed corpus files are located in `fuzz/corpus/`:

```
fuzz/corpus/
├── envelope/      # Encrypted envelope test cases
├── manifest/      # Deployment manifest test cases
├── order/         # Marketplace order test cases
└── scope/         # VEID scope test cases
```

### Adding to Corpus

When a fuzz test finds an interesting case, it automatically saves it to `testdata/fuzz/<TestName>/`:

```bash
# View discovered test cases
ls -la ./x/encryption/types/testdata/fuzz/FuzzEnvelopeValidate/

# Copy interesting cases to seed corpus
cp ./x/encryption/types/testdata/fuzz/FuzzEnvelopeValidate/* fuzz/corpus/envelope/
```

### Corpus Minimization

Periodically minimize the corpus to remove redundant test cases:

```bash
# Using Go's native corpus management
go test -fuzz=FuzzEnvelopeValidate -fuzztime=1s -fuzzminimizetime=30s ./x/encryption/types/...

# Using OSS-Fuzz's minimize command
python infra/helper.py minimize_corpus virtengine fuzz_envelope_validate
```

## Crash Triage Process

### When a Crash is Found

1. **Reproduce the crash:**
   ```bash
   go test -run=FuzzEnvelopeValidate/crashfile -v ./x/encryption/types/...
   ```

2. **Minimize the test case:**
   ```bash
   go test -fuzz=FuzzEnvelopeValidate -run=FuzzEnvelopeValidate/crashfile \
       -fuzzminimizetime=30s ./x/encryption/types/...
   ```

3. **Analyze the crash:**
   - Check for panics, nil pointer dereferences, out-of-bounds access
   - Review stack trace for affected code paths
   - Determine security impact (DoS, memory corruption, information leak)

4. **File an issue:**
   - For security issues: Follow responsible disclosure via security@virtengine.io
   - For non-security bugs: Create GitHub issue with reproduction steps

### OSS-Fuzz Crash Reports

OSS-Fuzz automatically files bug reports for discovered crashes:
- Security issues: Filed privately, 90-day disclosure deadline
- Non-security issues: Filed publicly after 30 days
- Monitor: https://bugs.chromium.org/p/oss-fuzz/issues/list?q=virtengine

### Reproducing OSS-Fuzz Crashes

```bash
# Download crash testcase from OSS-Fuzz
python infra/helper.py reproduce virtengine fuzz_envelope_validate \
    /path/to/testcase

# Debug with sanitizers
python infra/helper.py build_fuzzers --sanitizer address virtengine
python infra/helper.py reproduce virtengine fuzz_envelope_validate \
    /path/to/testcase
```

## Dictionaries

Fuzzing dictionaries improve mutation effectiveness by providing domain-specific tokens:

- `dictionaries/json.dict` - JSON structural tokens and VirtEngine type keys
- `dictionaries/crypto.dict` - Cryptographic algorithm names and key patterns

## Security Considerations

1. **Never commit crash-inducing inputs** that could be exploited before fix
2. **Security bugs** should be reported privately to security@virtengine.io
3. **Sensitive test data** (actual private keys, credentials) must not be in corpus
4. **Sanitizers** (ASan, MSan, UBSan) should be used to catch memory issues

## Integration with CI/CD

The CI pipeline runs fuzz tests on every PR:

```yaml
# .github/workflows/fuzz.yml
fuzz:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - name: Run fuzz tests
      run: |
        go test -fuzz=. -fuzztime=30s ./x/encryption/types/...
        go test -fuzz=. -fuzztime=30s ./x/veid/types/...
        go test -fuzz=. -fuzztime=30s ./x/market/types/marketplace/...
        go test -fuzz=. -fuzztime=30s ./pkg/provider_daemon/...
```

## Best Practices

1. **Fuzz tests should be fast** - Avoid expensive operations in fuzz targets
2. **Seed corpus should be diverse** - Include valid, invalid, and edge cases
3. **Check for panics** - All fuzz targets should handle any input without panicking
4. **Validate consistency** - Ensure deterministic behavior for same input
5. **Cover error paths** - Fuzz both success and failure code paths

## References

- [Go Fuzzing Guide](https://go.dev/doc/fuzz/)
- [OSS-Fuzz Documentation](https://google.github.io/oss-fuzz/)
- [libFuzzer Tutorial](https://llvm.org/docs/LibFuzzer.html)
- [ClusterFuzz Documentation](https://google.github.io/clusterfuzz/)
