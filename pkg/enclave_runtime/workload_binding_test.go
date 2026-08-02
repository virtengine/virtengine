package enclave_runtime

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

const workloadBindingMeasurementHex = "616263"

func validWorkloadBinding(t *testing.T, platform AttestationType) WorkloadBindingV1 {
	t.Helper()
	measurement, err := hex.DecodeString(workloadBindingMeasurementHex)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(measurement)
	return WorkloadBindingV1{
		Domain:                   WorkloadBindingDomainV1,
		Version:                  WorkloadBindingVersionV1,
		ChainID:                  "chain-alpha-7",
		WorkloadID:               "workload/model-a:rev3",
		Platform:                 platform,
		MeasurementSHA256:        hex.EncodeToString(digest[:]),
		MeasurementValueHex:      workloadBindingMeasurementHex,
		ReceiptSignerKeyID:       "kms:key/receipt-17",
		ReceiptSignerFingerprint: strings.Repeat("1", 64),
		ModelDigest:              strings.Repeat("2", 64),
		RuntimeDigest:            strings.Repeat("3", 64),
		Nonce:                    strings.Repeat("4", 64),
		ProfileID:                "profile/secure-v2",
		ProfileDigest:            strings.Repeat("5", 64),
		ActivationHeight:         120,
		ExpiryHeight:             180,
		Collateral: WorkloadBindingCollateralReferenceV1{
			Kind:         "opaque-bundle",
			OpaqueID:     "provider/ref-009",
			SHA256Digest: strings.Repeat("6", 64),
		},
		DebugMode: false,
	}
}

func TestWorkloadBindingGoldenCanonicalAndChallenge(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	const wantCanonical = `{"domain":"virtengine.workload-binding","version":"1","chain_id":"chain-alpha-7","workload_id":"workload/model-a:rev3","platform":1,"measurement_sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","measurement_value_hex":"616263","receipt_signer_key_id":"kms:key/receipt-17","receipt_signer_sha256":"1111111111111111111111111111111111111111111111111111111111111111","model_sha256":"2222222222222222222222222222222222222222222222222222222222222222","runtime_sha256":"3333333333333333333333333333333333333333333333333333333333333333","nonce_hex":"4444444444444444444444444444444444444444444444444444444444444444","profile_id":"profile/secure-v2","profile_sha256":"5555555555555555555555555555555555555555555555555555555555555555","activation_height":120,"expiry_height":180,"collateral":{"kind":"opaque-bundle","opaque_id":"provider/ref-009","sha256_digest":"6666666666666666666666666666666666666666666666666666666666666666"},"debug_mode":false}`
	const wantChallenge = "932708fe86626e775cd71aa75ad12c82e131e997303b9482b10484a55534d257"

	canonical, err := binding.CanonicalBytes()
	if err != nil {
		t.Fatalf("CanonicalBytes() error = %v", err)
	}
	if string(canonical) != wantCanonical {
		t.Fatalf("canonical mismatch\ngot:  %s\nwant: %s", canonical, wantCanonical)
	}
	digest, err := binding.ChallengeDigestHex()
	if err != nil {
		t.Fatalf("ChallengeDigestHex() error = %v", err)
	}
	if digest != wantChallenge {
		t.Fatalf("challenge mismatch: got %s want %s", digest, wantChallenge)
	}
}

func TestWorkloadBindingDeterministicFreshSlices(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	firstCanonical, err := binding.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	secondCanonical, err := binding.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstCanonical, secondCanonical) {
		t.Fatal("canonical encodings are not deterministic")
	}
	firstCanonical[0] ^= 0xff
	if bytes.Equal(firstCanonical, secondCanonical) {
		t.Fatal("CanonicalBytes returned aliased slices")
	}
	firstDigest, err := binding.ChallengeDigest()
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := binding.ChallengeDigest()
	if err != nil {
		t.Fatal(err)
	}
	firstDigest[0] ^= 0xff
	if bytes.Equal(firstDigest, secondDigest) {
		t.Fatal("ChallengeDigest returned aliased slices")
	}
}

func TestWorkloadBindingValidationFailures(t *testing.T) {
	tooLong := strings.Repeat("a", maxWorkloadBindingIdentifierLength+1)
	uppercaseDigest := "A" + strings.Repeat("a", 63)
	tests := []struct {
		name   string
		mutate func(*WorkloadBindingV1)
		chain  string
		height int64
	}{
		{"domain", func(b *WorkloadBindingV1) { b.Domain = "other" }, "chain-alpha-7", 120},
		{"version", func(b *WorkloadBindingV1) { b.Version = "2" }, "chain-alpha-7", 120},
		{"empty chain", func(b *WorkloadBindingV1) { b.ChainID = "" }, "", 120},
		{"chain mismatch", func(*WorkloadBindingV1) {}, "chain-beta", 120},
		{"untrimmed workload", func(b *WorkloadBindingV1) { b.WorkloadID = " workload" }, "chain-alpha-7", 120},
		{"non printable workload", func(b *WorkloadBindingV1) { b.WorkloadID = "work\nload" }, "chain-alpha-7", 120},
		{"non ASCII workload", func(b *WorkloadBindingV1) { b.WorkloadID = "workload-\u00e9" }, "chain-alpha-7", 120},
		{"long workload", func(b *WorkloadBindingV1) { b.WorkloadID = tooLong }, "chain-alpha-7", 120},
		{"empty signer key ID", func(b *WorkloadBindingV1) { b.ReceiptSignerKeyID = "" }, "chain-alpha-7", 120},
		{"empty profile ID", func(b *WorkloadBindingV1) { b.ProfileID = "" }, "chain-alpha-7", 120},
		{"empty collateral kind", func(b *WorkloadBindingV1) { b.Collateral.Kind = "" }, "chain-alpha-7", 120},
		{"empty collateral opaque ID", func(b *WorkloadBindingV1) { b.Collateral.OpaqueID = "" }, "chain-alpha-7", 120},
		{"unknown platform", func(b *WorkloadBindingV1) { b.Platform = AttestationTypeUnknown }, "chain-alpha-7", 120},
		{"simulated platform", func(b *WorkloadBindingV1) { b.Platform = AttestationTypeSimulated }, "chain-alpha-7", 120},
		{"measurement digest", func(b *WorkloadBindingV1) { b.MeasurementSHA256 = uppercaseDigest }, "chain-alpha-7", 120},
		{"measurement value empty", func(b *WorkloadBindingV1) { b.MeasurementValueHex = "" }, "chain-alpha-7", 120},
		{"measurement value uppercase", func(b *WorkloadBindingV1) { b.MeasurementValueHex = "AA" }, "chain-alpha-7", 120},
		{"measurement value odd", func(b *WorkloadBindingV1) { b.MeasurementValueHex = "abc" }, "chain-alpha-7", 120},
		{"measurement commitment mismatch", func(b *WorkloadBindingV1) { b.MeasurementValueHex = "00" }, "chain-alpha-7", 120},
		{"signer fingerprint", func(b *WorkloadBindingV1) { b.ReceiptSignerFingerprint = "0" }, "chain-alpha-7", 120},
		{"model digest", func(b *WorkloadBindingV1) { b.ModelDigest = uppercaseDigest }, "chain-alpha-7", 120},
		{"runtime digest", func(b *WorkloadBindingV1) { b.RuntimeDigest = "g" + strings.Repeat("0", 63) }, "chain-alpha-7", 120},
		{"nonce", func(b *WorkloadBindingV1) { b.Nonce = strings.Repeat("0", 62) }, "chain-alpha-7", 120},
		{"profile digest", func(b *WorkloadBindingV1) { b.ProfileDigest = "" }, "chain-alpha-7", 120},
		{"collateral digest", func(b *WorkloadBindingV1) { b.Collateral.SHA256Digest = uppercaseDigest }, "chain-alpha-7", 120},
		{"activation zero", func(b *WorkloadBindingV1) { b.ActivationHeight = 0 }, "chain-alpha-7", 120},
		{"expiry equal activation", func(b *WorkloadBindingV1) { b.ExpiryHeight = b.ActivationHeight }, "chain-alpha-7", 120},
		{"debug", func(b *WorkloadBindingV1) { b.DebugMode = true }, "chain-alpha-7", 120},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binding := validWorkloadBinding(t, AttestationTypeSGX)
			test.mutate(&binding)
			if err := binding.Validate(test.chain, test.height); err == nil {
				t.Fatal("Validate() error = nil, want failure")
			}
		})
	}
}

func TestWorkloadBindingHeightBoundaries(t *testing.T) {
	binding := validWorkloadBinding(t, AttestationTypeSGX)
	tests := []struct {
		height int64
		valid  bool
	}{{119, false}, {120, true}, {179, true}, {180, false}}
	for _, test := range tests {
		err := binding.Validate(binding.ChainID, test.height)
		if (err == nil) != test.valid {
			t.Errorf("Validate(height=%d) error = %v, valid want %v", test.height, err, test.valid)
		}
	}
}

func TestWorkloadBindingAllFieldTamperChangesChallenge(t *testing.T) {
	original := validWorkloadBinding(t, AttestationTypeSGX)
	want, err := original.ChallengeDigestHex()
	if err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name   string
		mutate func(*WorkloadBindingV1)
	}{
		{"chain", func(b *WorkloadBindingV1) { b.ChainID = "chain-beta-8" }},
		{"workload", func(b *WorkloadBindingV1) { b.WorkloadID += "-next" }},
		{"platform", func(b *WorkloadBindingV1) { b.Platform = AttestationTypeSEVSNP }},
		{"measurement value and digest", func(b *WorkloadBindingV1) {
			b.MeasurementValueHex = "10" + b.MeasurementValueHex[2:]
			value, _ := hex.DecodeString(b.MeasurementValueHex)
			digest := sha256.Sum256(value)
			b.MeasurementSHA256 = hex.EncodeToString(digest[:])
		}},
		{"signer key ID", func(b *WorkloadBindingV1) { b.ReceiptSignerKeyID += "-next" }},
		{"signer fingerprint", func(b *WorkloadBindingV1) { b.ReceiptSignerFingerprint = strings.Repeat("7", 64) }},
		{"model digest", func(b *WorkloadBindingV1) { b.ModelDigest = strings.Repeat("7", 64) }},
		{"runtime digest", func(b *WorkloadBindingV1) { b.RuntimeDigest = strings.Repeat("7", 64) }},
		{"nonce", func(b *WorkloadBindingV1) { b.Nonce = strings.Repeat("7", 64) }},
		{"profile ID", func(b *WorkloadBindingV1) { b.ProfileID += "-next" }},
		{"profile digest", func(b *WorkloadBindingV1) { b.ProfileDigest = strings.Repeat("7", 64) }},
		{"activation", func(b *WorkloadBindingV1) { b.ActivationHeight++ }},
		{"expiry", func(b *WorkloadBindingV1) { b.ExpiryHeight++ }},
		{"collateral kind", func(b *WorkloadBindingV1) { b.Collateral.Kind += "-next" }},
		{"collateral opaque ID", func(b *WorkloadBindingV1) { b.Collateral.OpaqueID += "-next" }},
		{"collateral digest", func(b *WorkloadBindingV1) { b.Collateral.SHA256Digest = strings.Repeat("7", 64) }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			binding := original
			mutation.mutate(&binding)
			got, err := binding.ChallengeDigestHex()
			if err != nil {
				t.Fatalf("valid mutation rejected: %v", err)
			}
			if got == want {
				t.Fatal("mutation did not change challenge")
			}
		})
	}
	invalidMutations := []struct {
		name   string
		mutate func(*WorkloadBindingV1)
	}{
		{"domain", func(b *WorkloadBindingV1) { b.Domain += ".other" }},
		{"version", func(b *WorkloadBindingV1) { b.Version = "2" }},
		{"debug", func(b *WorkloadBindingV1) { b.DebugMode = true }},
	}
	for _, mutation := range invalidMutations {
		t.Run("invalid "+mutation.name, func(t *testing.T) {
			binding := original
			mutation.mutate(&binding)
			if _, err := binding.ChallengeDigest(); err == nil {
				t.Fatal("invalid mutation accepted")
			}
		})
	}
}
