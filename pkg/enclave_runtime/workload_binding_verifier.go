package enclave_runtime

import (
	"bytes"
	"fmt"
	"reflect"
)

// WorkloadBindingAttestationVerifier is the network-free verifier surface used
// by the Workload Binding verifier contract and fixture boundary. Implementing
// it does not certify hardware, collateral, signer ownership, replay consumption,
// or production readiness.
type WorkloadBindingAttestationVerifier interface {
	Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error)
}

// WorkloadBindingVerifier adapts existing attestation verifier semantics to a
// Workload Binding fixture boundary. It makes no hardware, collateral, signer
// ownership, replay consumption, or production-readiness certification.
type WorkloadBindingVerifier struct {
	verifier   WorkloadBindingAttestationVerifier
	basePolicy VerificationPolicy
}

// NewWorkloadBindingVerifier constructs a verifier contract/fixture adapter. The
// base policy is hardened. It does not certify hardware, collateral, signer
// ownership, replay consumption, or production readiness.
func NewWorkloadBindingVerifier(verifier WorkloadBindingAttestationVerifier, basePolicy VerificationPolicy) *WorkloadBindingVerifier {
	return &WorkloadBindingVerifier{
		verifier:   verifier,
		basePolicy: cloneVerificationPolicy(basePolicy),
	}
}

// Verify validates a binding, verifies attestation against its challenge, and
// returns a deep-copied result. This contract boundary does not certify hardware,
// collateral, signer ownership, replay consumption, or production readiness.
func (v *WorkloadBindingVerifier) Verify(binding *WorkloadBindingV1, chainID string, height int64, attestation []byte) (*VerificationResult, error) {
	if v == nil || isNilWorkloadBindingVerifier(v.verifier) {
		return nil, fmt.Errorf("workload binding verifier is nil")
	}
	if binding == nil {
		return nil, fmt.Errorf("workload binding is nil")
	}
	if err := binding.Validate(chainID, height); err != nil {
		return nil, fmt.Errorf("validate workload binding: %w", err)
	}
	if binding.Platform == AttestationTypeSimulated {
		return nil, fmt.Errorf("simulated attestation is forbidden")
	}
	challenge, err := binding.ChallengeDigest()
	if err != nil {
		return nil, fmt.Errorf("derive workload binding challenge: %w", err)
	}
	policy := cloneVerificationPolicy(v.basePolicy)
	policy.AllowDebugMode = false
	policy.RequireNonce = true
	policy.AllowIncompleteVerification = false
	policy.AllowedPlatforms = []AttestationType{binding.Platform}
	strictMaxAttestationAge := DefaultVerificationPolicy().MaxAttestationAge
	if policy.MaxAttestationAge <= 0 || policy.MaxAttestationAge > strictMaxAttestationAge {
		policy.MaxAttestationAge = strictMaxAttestationAge
	}

	result, err := v.verifier.Verify(append([]byte(nil), attestation...), append([]byte(nil), challenge...), policy)
	if err != nil {
		return nil, fmt.Errorf("verify workload binding attestation: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("attestation verifier returned nil result")
	}
	if !result.Valid {
		return nil, fmt.Errorf("attestation verification result is invalid")
	}
	if result.AttestationType == AttestationTypeUnknown || result.AttestationType == AttestationTypeSimulated {
		return nil, fmt.Errorf("attestation verifier returned unsupported platform")
	}
	if result.AttestationType != binding.Platform {
		return nil, fmt.Errorf("attestation platform mismatch")
	}
	measurement, err := decodeMeasurementValue(binding.MeasurementValueHex)
	if err != nil {
		return nil, err
	}
	if len(result.Measurement) == 0 || !bytes.Equal(result.Measurement, measurement) {
		return nil, fmt.Errorf("attestation measurement mismatch")
	}
	if result.DebugMode {
		return nil, fmt.Errorf("debug attestation result is forbidden")
	}
	if !workloadBindingNonceMatchesChallenge(result.Nonce, challenge) {
		return nil, fmt.Errorf("attestation nonce mismatch")
	}
	return cloneVerificationResult(result), nil
}

// workloadBindingNonceMatchesChallenge normalizes SGX and SEV-SNP 64-byte
// report data without weakening challenge binding: only zero padding after the
// complete, exact challenge is accepted.
func workloadBindingNonceMatchesChallenge(reportData, challenge []byte) bool {
	if len(reportData) < len(challenge) || !bytes.Equal(reportData[:len(challenge)], challenge) {
		return false
	}
	for _, suffixByte := range reportData[len(challenge):] {
		if suffixByte != 0 {
			return false
		}
	}
	return true
}

func isNilWorkloadBindingVerifier(verifier WorkloadBindingAttestationVerifier) bool {
	if verifier == nil {
		return true
	}
	value := reflect.ValueOf(verifier)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func cloneVerificationPolicy(policy VerificationPolicy) VerificationPolicy {
	cloned := policy
	cloned.AllowedPlatforms = append([]AttestationType(nil), policy.AllowedPlatforms...)
	cloned.TrustedSignerKeys = make([][]byte, len(policy.TrustedSignerKeys))
	for index := range policy.TrustedSignerKeys {
		cloned.TrustedSignerKeys[index] = append([]byte(nil), policy.TrustedSignerKeys[index]...)
	}
	return cloned
}

func cloneVerificationResult(result *VerificationResult) *VerificationResult {
	cloned := *result
	cloned.Measurement = append([]byte(nil), result.Measurement...)
	cloned.SignerKey = append([]byte(nil), result.SignerKey...)
	cloned.Errors = append([]string(nil), result.Errors...)
	cloned.Warnings = append([]string(nil), result.Warnings...)
	cloned.RawAttestation = append([]byte(nil), result.RawAttestation...)
	cloned.Nonce = append([]byte(nil), result.Nonce...)
	return &cloned
}
