package enclave_runtime

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

type workloadBindingFakeVerifier struct {
	result       *VerificationResult
	err          error
	nonce        []byte
	policy       VerificationPolicy
	attestation  []byte
	mutateInputs bool
}

func (f *workloadBindingFakeVerifier) Verify(attestation []byte, nonce []byte, policy VerificationPolicy) (*VerificationResult, error) {
	f.attestation = append([]byte(nil), attestation...)
	f.nonce = append([]byte(nil), nonce...)
	f.policy = cloneVerificationPolicy(policy)
	if f.mutateInputs {
		attestation[0] ^= 0xff
		nonce[0] ^= 0xff
		policy.AllowedPlatforms[0] = AttestationTypeSimulated
		if len(policy.TrustedSignerKeys) > 0 && len(policy.TrustedSignerKeys[0]) > 0 {
			policy.TrustedSignerKeys[0][0] ^= 0xff
		}
	}
	return f.result, f.err
}

func successfulWorkloadBindingFake(t *testing.T, binding WorkloadBindingV1) *workloadBindingFakeVerifier {
	t.Helper()
	measurement, err := decodeMeasurementValue(binding.MeasurementValueHex)
	if err != nil {
		t.Fatal(err)
	}
	challenge, err := binding.ChallengeDigest()
	if err != nil {
		t.Fatal(err)
	}
	return &workloadBindingFakeVerifier{result: &VerificationResult{
		Valid:           true,
		AttestationType: binding.Platform,
		Measurement:     measurement,
		SignerKey:       []byte("signer-output"),
		TCBVersion:      "fixture-tcb",
		RawAttestation:  []byte("raw-output"),
		Nonce:           challenge,
		Warnings:        []string{"fixture-warning"},
		SecurityLevel:   3,
	}}
}

func TestWorkloadBindingVerifierAcceptsHardwarePlatforms(t *testing.T) {
	for _, platform := range []AttestationType{AttestationTypeSGX, AttestationTypeSEVSNP, AttestationTypeNitro} {
		t.Run(platform.String(), func(t *testing.T) {
			binding := validWorkloadBinding(t, platform)
			fake := successfulWorkloadBindingFake(t, binding)
			adapter := NewWorkloadBindingVerifier(fake, PermissiveVerificationPolicy())
			result, err := adapter.Verify(&binding, binding.ChainID, binding.ActivationHeight, []byte("fixture-attestation"))
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if result == fake.result {
				t.Fatal("Verify returned verifier-owned result pointer")
			}
			challenge, _ := binding.ChallengeDigest()
			if !bytes.Equal(fake.nonce, challenge) {
				t.Fatal("verifier did not receive binding challenge")
			}
		})
	}
}

func TestWorkloadBindingVerifierForcesStrictPolicy(t *testing.T) {
	strictMaxAge := DefaultVerificationPolicy().MaxAttestationAge
	tests := []struct {
		name       string
		base       VerificationPolicy
		wantMaxAge time.Duration
	}{
		{"caps permissive age", PermissiveVerificationPolicy(), strictMaxAge},
		{"preserves stricter positive age", func() VerificationPolicy {
			policy := PermissiveVerificationPolicy()
			policy.MaxAttestationAge = time.Hour
			return policy
		}(), time.Hour},
		{"replaces non-positive age", VerificationPolicy{}, strictMaxAge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binding := validWorkloadBinding(t, AttestationTypeNitro)
			fake := successfulWorkloadBindingFake(t, binding)
			test.base.AllowedPlatforms = []AttestationType{AttestationTypeSimulated}
			adapter := NewWorkloadBindingVerifier(fake, test.base)
			if _, err := adapter.Verify(&binding, binding.ChainID, 150, []byte("attestation")); err != nil {
				t.Fatal(err)
			}
			if fake.policy.AllowDebugMode || !fake.policy.RequireNonce || fake.policy.AllowIncompleteVerification {
				t.Fatalf("policy was not hardened: %+v", fake.policy)
			}
			if len(fake.policy.AllowedPlatforms) != 1 || fake.policy.AllowedPlatforms[0] != AttestationTypeNitro {
				t.Fatalf("allowed platforms = %v, want only Nitro", fake.policy.AllowedPlatforms)
			}
			if fake.policy.MaxAttestationAge != test.wantMaxAge {
				t.Fatalf("MaxAttestationAge = %v, want %v", fake.policy.MaxAttestationAge, test.wantMaxAge)
			}
		})
	}
}

func TestWorkloadBindingVerifierAcceptsZeroPaddedReportData(t *testing.T) {
	for _, platform := range []AttestationType{AttestationTypeSGX, AttestationTypeSEVSNP} {
		t.Run(platform.String(), func(t *testing.T) {
			binding := validWorkloadBinding(t, platform)
			fake := successfulWorkloadBindingFake(t, binding)
			fake.result.Nonce = append(fake.result.Nonce, make([]byte, 64-len(fake.result.Nonce))...)
			adapter := NewWorkloadBindingVerifier(fake, DefaultVerificationPolicy())
			if _, err := adapter.Verify(&binding, binding.ChainID, 150, []byte("attestation")); err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
		})
	}
}

func TestWorkloadBindingVerifierRejectsUnsafeResults(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	var typedNilVerifier *workloadBindingFakeVerifier
	tests := []struct {
		name   string
		build  func(*testing.T, WorkloadBindingV1) *WorkloadBindingVerifier
		mutate func(*WorkloadBindingV1)
	}{
		{"nil verifier", func(*testing.T, WorkloadBindingV1) *WorkloadBindingVerifier {
			return NewWorkloadBindingVerifier(nil, VerificationPolicy{})
		}, nil},
		{"nil result", func(*testing.T, WorkloadBindingV1) *WorkloadBindingVerifier {
			return NewWorkloadBindingVerifier(&workloadBindingFakeVerifier{}, VerificationPolicy{})
		}, nil},
		{"verifier error", func(*testing.T, WorkloadBindingV1) *WorkloadBindingVerifier {
			return NewWorkloadBindingVerifier(&workloadBindingFakeVerifier{err: errors.New("fixture failure")}, VerificationPolicy{})
		}, nil},
		{"invalid result", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Valid = false
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"unknown result platform", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.AttestationType = AttestationTypeUnknown
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"simulated result platform", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.AttestationType = AttestationTypeSimulated
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"platform mismatch", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.AttestationType = AttestationTypeNitro
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"empty measurement", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Measurement = nil
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"measurement mismatch", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Measurement[0] ^= 0xff
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"debug result", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.DebugMode = true
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"empty nonce", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Nonce = nil
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"wrong nonce prefix", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Nonce[0] ^= 0xff
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"nonzero nonce suffix", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Nonce = append(f.result.Nonce, 0, 1)
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"short nonce", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			f := successfulWorkloadBindingFake(t, b)
			f.result.Nonce = f.result.Nonce[:len(f.result.Nonce)-1]
			return NewWorkloadBindingVerifier(f, VerificationPolicy{})
		}, nil},
		{"simulated binding", func(t *testing.T, b WorkloadBindingV1) *WorkloadBindingVerifier {
			return NewWorkloadBindingVerifier(successfulWorkloadBindingFake(t, b), VerificationPolicy{})
		}, func(b *WorkloadBindingV1) { b.Platform = AttestationTypeSimulated }},
		{"typed nil verifier", func(*testing.T, WorkloadBindingV1) *WorkloadBindingVerifier {
			return NewWorkloadBindingVerifier(typedNilVerifier, VerificationPolicy{})
		}, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := binding
			adapter := test.build(t, candidate)
			if test.mutate != nil {
				test.mutate(&candidate)
			}
			if _, err := adapter.Verify(&candidate, candidate.ChainID, 150, []byte("attestation")); err == nil {
				t.Fatal("Verify() error = nil, want rejection")
			}
		})
	}
	var nilAdapter *WorkloadBindingVerifier
	if _, err := nilAdapter.Verify(&binding, binding.ChainID, 150, nil); err == nil {
		t.Fatal("nil adapter accepted")
	}
	adapter := NewWorkloadBindingVerifier(successfulWorkloadBindingFake(t, binding), VerificationPolicy{})
	if _, err := adapter.Verify(nil, binding.ChainID, 150, nil); err == nil {
		t.Fatal("nil binding accepted")
	}
}

func TestWorkloadBindingVerifierReturnsDeepCopy(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	fake := successfulWorkloadBindingFake(t, binding)
	fake.result.Errors = []string{"fixture-error"}
	adapter := NewWorkloadBindingVerifier(fake, VerificationPolicy{})
	result, err := adapter.Verify(&binding, binding.ChainID, 150, []byte("attestation"))
	if err != nil {
		t.Fatal(err)
	}
	fake.result.Measurement[0] ^= 0xff
	fake.result.SignerKey[0] ^= 0xff
	fake.result.RawAttestation[0] ^= 0xff
	fake.result.Nonce[0] ^= 0xff
	fake.result.Errors[0] = "changed"
	fake.result.Warnings[0] = "changed"
	if bytes.Equal(result.Measurement, fake.result.Measurement) || bytes.Equal(result.SignerKey, fake.result.SignerKey) ||
		bytes.Equal(result.RawAttestation, fake.result.RawAttestation) || bytes.Equal(result.Nonce, fake.result.Nonce) ||
		result.Errors[0] == fake.result.Errors[0] || result.Warnings[0] == fake.result.Warnings[0] {
		t.Fatal("returned VerificationResult contains verifier-owned slices")
	}
}

func TestNewWorkloadBindingVerifierCopiesBasePolicy(t *testing.T) {
	base := DefaultVerificationPolicy()
	base.AllowedPlatforms = []AttestationType{AttestationTypeSGX}
	base.TrustedSignerKeys = [][]byte{{1, 2, 3}}
	adapter := NewWorkloadBindingVerifier(&workloadBindingFakeVerifier{}, base)

	base.AllowedPlatforms[0] = AttestationTypeSimulated
	base.TrustedSignerKeys[0][0] = 9
	if adapter.basePolicy.AllowedPlatforms[0] != AttestationTypeSGX {
		t.Fatal("constructor retained caller-owned AllowedPlatforms")
	}
	if adapter.basePolicy.TrustedSignerKeys[0][0] != 1 {
		t.Fatal("constructor retained caller-owned TrustedSignerKeys")
	}
}

func TestWorkloadBindingVerifierCopiesVerifierInputs(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	originalBinding := binding
	originalChallenge, err := binding.ChallengeDigest()
	if err != nil {
		t.Fatal(err)
	}
	attestation := []byte("fixture-attestation")
	originalAttestation := append([]byte(nil), attestation...)
	base := DefaultVerificationPolicy()
	base.TrustedSignerKeys = [][]byte{{1, 2, 3}}
	fake := successfulWorkloadBindingFake(t, binding)
	fake.mutateInputs = true
	adapter := NewWorkloadBindingVerifier(fake, base)
	if _, err := adapter.Verify(&binding, binding.ChainID, 150, attestation); err != nil {
		t.Fatal(err)
	}
	challengeAfter, err := binding.ChallengeDigest()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(attestation, originalAttestation) {
		t.Fatal("verifier mutated caller-owned attestation")
	}
	if binding != originalBinding || !bytes.Equal(challengeAfter, originalChallenge) {
		t.Fatal("verifier mutation changed caller-owned binding or challenge")
	}
	if adapter.basePolicy.TrustedSignerKeys[0][0] != 1 {
		t.Fatal("verifier mutated adapter-owned policy")
	}
}
