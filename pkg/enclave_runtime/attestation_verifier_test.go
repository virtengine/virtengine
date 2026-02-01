package enclave_runtime

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSGXVerification(t *testing.T) {
	allowlist := NewMeasurementAllowlist()

	// Add trusted measurement
	trustedMRENCLAVE := make([]byte, 32)
	for i := range trustedMRENCLAVE {
		trustedMRENCLAVE[i] = byte(i)
	}
	if err := allowlist.AddMeasurement(AttestationTypeSGX, trustedMRENCLAVE, "test enclave"); err != nil {
		t.Fatalf("failed to add measurement: %v", err)
	}

	verifier := NewSGXDCAPVerifier(allowlist)

	t.Run("valid production quote", func(t *testing.T) {
		mrsigner := make([]byte, 32)
		nonce := []byte("test-nonce-12345")
		quote := CreateTestSGXQuote(trustedMRENCLAVE, mrsigner, false, nonce)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(quote, nonce, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}

		if result.AttestationType != AttestationTypeSGX {
			t.Errorf("expected SGX type, got %v", result.AttestationType)
		}

		if result.DebugMode {
			t.Error("expected non-debug mode")
		}

		if !bytes.Equal(result.Measurement, trustedMRENCLAVE) {
			t.Errorf("measurement mismatch: got %x, want %x", result.Measurement, trustedMRENCLAVE)
		}
	})

	t.Run("reject debug mode with strict policy", func(t *testing.T) {
		mrsigner := make([]byte, 32)
		quote := CreateTestSGXQuote(trustedMRENCLAVE, mrsigner, true, nil)

		policy := DefaultVerificationPolicy()
		policy.RequireNonce = false
		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for debug mode with strict policy")
		}

		if !result.DebugMode {
			t.Error("expected debug mode to be detected")
		}

		hasDebugError := false
		for _, e := range result.Errors {
			if containsSubstring(e, "debug mode") {
				hasDebugError = true
				break
			}
		}
		if !hasDebugError {
			t.Errorf("expected debug mode error, got: %v", result.Errors)
		}
	})

	t.Run("allow debug mode with permissive policy", func(t *testing.T) {
		mrsigner := make([]byte, 32)
		quote := CreateTestSGXQuote(trustedMRENCLAVE, mrsigner, true, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result with permissive policy, got errors: %v", result.Errors)
		}
	})

	t.Run("reject untrusted measurement", func(t *testing.T) {
		untrustedMRENCLAVE := make([]byte, 32)
		for i := range untrustedMRENCLAVE {
			untrustedMRENCLAVE[i] = byte(i + 100)
		}
		mrsigner := make([]byte, 32)
		quote := CreateTestSGXQuote(untrustedMRENCLAVE, mrsigner, false, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for untrusted measurement")
		}

		hasMeasurementError := false
		for _, e := range result.Errors {
			if containsSubstring(e, "not in allowlist") {
				hasMeasurementError = true
				break
			}
		}
		if !hasMeasurementError {
			t.Errorf("expected measurement error, got: %v", result.Errors)
		}
	})

	t.Run("reject too small quote", func(t *testing.T) {
		smallQuote := []byte{0x03, 0x00, 0x01, 0x02}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(smallQuote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for small quote")
		}
	})

	t.Run("nonce verification", func(t *testing.T) {
		mrsigner := make([]byte, 32)
		expectedNonce := []byte("expected-nonce-1234567890")
		wrongNonce := []byte("wrong-nonce-0987654321")
		quote := CreateTestSGXQuote(trustedMRENCLAVE, mrsigner, false, wrongNonce)

		policy := VerificationPolicy{
			AllowDebugMode:   true,
			AllowedPlatforms: []AttestationType{AttestationTypeSGX},
			RequireNonce:     true,
		}
		result, err := verifier.Verify(quote, expectedNonce, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for nonce mismatch")
		}

		hasNonceError := false
		for _, e := range result.Errors {
			if containsSubstring(e, "nonce") {
				hasNonceError = true
				break
			}
		}
		if !hasNonceError {
			t.Errorf("expected nonce error, got: %v", result.Errors)
		}
	})
}

func TestSEVSNPVerification(t *testing.T) {
	allowlist := NewMeasurementAllowlist()

	// Add trusted measurement (48 bytes for SEV-SNP)
	trustedMeasurement := make([]byte, 48)
	for i := range trustedMeasurement {
		trustedMeasurement[i] = byte(i)
	}
	if err := allowlist.AddMeasurement(AttestationTypeSEVSNP, trustedMeasurement, "test launch digest"); err != nil {
		t.Fatalf("failed to add measurement: %v", err)
	}

	verifier := NewSEVSNPVerifier(allowlist)

	t.Run("valid production report", func(t *testing.T) {
		nonce := []byte("test-nonce-for-sev")
		report := CreateTestSEVSNPReport(trustedMeasurement, false, nonce)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(report, nonce, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}

		if result.AttestationType != AttestationTypeSEVSNP {
			t.Errorf("expected SEV-SNP type, got %v", result.AttestationType)
		}

		if result.DebugMode {
			t.Error("expected non-debug mode")
		}
	})

	t.Run("reject debug policy with strict settings", func(t *testing.T) {
		report := CreateTestSEVSNPReport(trustedMeasurement, true, nil)

		policy := DefaultVerificationPolicy()
		policy.RequireNonce = false
		result, err := verifier.Verify(report, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for debug policy")
		}

		if !result.DebugMode {
			t.Error("expected debug mode to be detected")
		}
	})

	t.Run("reject untrusted launch digest", func(t *testing.T) {
		untrustedMeasurement := make([]byte, 48)
		for i := range untrustedMeasurement {
			untrustedMeasurement[i] = byte(i + 50)
		}
		report := CreateTestSEVSNPReport(untrustedMeasurement, false, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(report, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for untrusted measurement")
		}
	})

	t.Run("reject too small report", func(t *testing.T) {
		smallReport := []byte{0x01, 0x00, 0x00, 0x00}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(smallReport, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for small report")
		}
	})

	t.Run("TCB version extraction", func(t *testing.T) {
		report := CreateTestSEVSNPReport(trustedMeasurement, false, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(report, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.TCBVersion == "" {
			t.Error("expected TCB version to be extracted")
		}
	})
}

func TestNitroVerification(t *testing.T) {
	allowlist := NewMeasurementAllowlist()

	// Add trusted PCR values - must match what CreateTestNitroDocument produces
	// The measurement extraction includes the nonce area in our test document
	trustedPCRs := make([]byte, 48)
	for i := range trustedPCRs {
		trustedPCRs[i] = byte(i)
	}

	verifier := NewNitroVerifier(allowlist)

	t.Run("valid document", func(t *testing.T) {
		// Create document without nonce first to get clean measurement
		doc := CreateTestNitroDocument(trustedPCRs, nil)

		// Add the actual extracted measurement to the allowlist
		// The measurement is extracted from offset 32 with length 48
		actualMeasurement := doc[nitroPCROffset : nitroPCROffset+nitroPCRLength]
		if err := allowlist.AddMeasurement(AttestationTypeNitro, actualMeasurement, "test PCRs"); err != nil {
			t.Fatalf("failed to add measurement: %v", err)
		}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(doc, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}

		if result.AttestationType != AttestationTypeNitro {
			t.Errorf("expected Nitro type, got %v", result.AttestationType)
		}

		if result.DebugMode {
			t.Error("Nitro should not report debug mode in production")
		}
	})

	t.Run("reject untrusted PCRs", func(t *testing.T) {
		untrustedPCRs := make([]byte, 48)
		for i := range untrustedPCRs {
			untrustedPCRs[i] = byte(i + 100)
		}
		doc := CreateTestNitroDocument(untrustedPCRs, nil)

		// Create a fresh allowlist without the untrusted measurement
		freshAllowlist := NewMeasurementAllowlist()
		freshVerifier := NewNitroVerifier(freshAllowlist)

		policy := PermissiveVerificationPolicy()
		result, err := freshVerifier.Verify(doc, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for untrusted PCRs")
		}
	})

	t.Run("reject too small document", func(t *testing.T) {
		smallDoc := []byte{0xD2, 0x84}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(smallDoc, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for small document")
		}
	})

	t.Run("enclave ID extraction", func(t *testing.T) {
		doc := CreateTestNitroDocument(trustedPCRs, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(doc, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.NitroEnclaveID == "" {
			t.Error("expected Nitro enclave ID to be extracted")
		}
	})
}

func TestMeasurementAllowlist(t *testing.T) {
	t.Run("add and check measurement", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()

		measurement := []byte{0x01, 0x02, 0x03, 0x04}
		if err := allowlist.AddMeasurement(AttestationTypeSGX, measurement, "test"); err != nil {
			t.Fatalf("failed to add measurement: %v", err)
		}

		if !allowlist.IsTrusted(AttestationTypeSGX, measurement) {
			t.Error("expected measurement to be trusted")
		}

		if allowlist.IsTrusted(AttestationTypeSEVSNP, measurement) {
			t.Error("measurement should not be trusted for different platform")
		}
	})

	t.Run("remove measurement", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()

		measurement := []byte{0x01, 0x02, 0x03, 0x04}
		if err := allowlist.AddMeasurement(AttestationTypeSGX, measurement, "test"); err != nil {
			t.Fatalf("failed to add measurement: %v", err)
		}

		if err := allowlist.RemoveMeasurement(AttestationTypeSGX, measurement); err != nil {
			t.Fatalf("failed to remove measurement: %v", err)
		}

		if allowlist.IsTrusted(AttestationTypeSGX, measurement) {
			t.Error("measurement should not be trusted after removal")
		}
	})

	t.Run("list measurements", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()

		m1 := []byte{0x01, 0x02, 0x03}
		m2 := []byte{0x04, 0x05, 0x06}
		_ = allowlist.AddMeasurement(AttestationTypeSGX, m1, "first")
		_ = allowlist.AddMeasurement(AttestationTypeSGX, m2, "second")

		list := allowlist.ListMeasurements(AttestationTypeSGX)
		if len(list) != 2 {
			t.Errorf("expected 2 measurements, got %d", len(list))
		}
	})

	t.Run("empty measurement rejected", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()

		err := allowlist.AddMeasurement(AttestationTypeSGX, []byte{}, "empty")
		if err == nil {
			t.Error("expected error for empty measurement")
		}
	})

	t.Run("remove nonexistent measurement", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()

		err := allowlist.RemoveMeasurement(AttestationTypeSGX, []byte{0x01})
		if err == nil {
			t.Error("expected error for removing nonexistent measurement")
		}
	})

	t.Run("JSON persistence", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "allowlist.json")

		// Create and save
		allowlist := NewMeasurementAllowlist()
		m1 := []byte{0x01, 0x02, 0x03, 0x04}
		m2 := []byte{0x05, 0x06, 0x07, 0x08}
		_ = allowlist.AddMeasurement(AttestationTypeSGX, m1, "sgx test")
		_ = allowlist.AddMeasurement(AttestationTypeSEVSNP, m2, "sevsnp test")

		if err := allowlist.SaveToJSON(filePath); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatal("allowlist file was not created")
		}

		// Load into new allowlist
		newAllowlist := NewMeasurementAllowlist()
		if err := newAllowlist.LoadFromJSON(filePath); err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if !newAllowlist.IsTrusted(AttestationTypeSGX, m1) {
			t.Error("SGX measurement not found after load")
		}
		if !newAllowlist.IsTrusted(AttestationTypeSEVSNP, m2) {
			t.Error("SEV-SNP measurement not found after load")
		}
	})
}

func TestPolicyEnforcement(t *testing.T) {
	allowlist := NewMeasurementAllowlist()
	trustedMeasurement := make([]byte, 32)
	for i := range trustedMeasurement {
		trustedMeasurement[i] = byte(i)
	}
	_ = allowlist.AddMeasurement(AttestationTypeSGX, trustedMeasurement, "trusted")

	verifier := NewSGXDCAPVerifier(allowlist)
	mrsigner := make([]byte, 32)

	t.Run("platform restriction", func(t *testing.T) {
		quote := CreateTestSGXQuote(trustedMeasurement, mrsigner, false, nil)

		policy := VerificationPolicy{
			AllowDebugMode:   true,
			AllowedPlatforms: []AttestationType{AttestationTypeSEVSNP}, // SGX not allowed
		}

		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result when platform not allowed")
		}

		hasPlatformError := false
		for _, e := range result.Errors {
			if containsSubstring(e, "not allowed") {
				hasPlatformError = true
				break
			}
		}
		if !hasPlatformError {
			t.Errorf("expected platform error, got: %v", result.Errors)
		}
	})

	t.Run("empty platform list allows all", func(t *testing.T) {
		quote := CreateTestSGXQuote(trustedMeasurement, mrsigner, false, nil)

		policy := VerificationPolicy{
			AllowDebugMode:   true,
			AllowedPlatforms: []AttestationType{}, // Empty = all allowed
		}

		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result with empty platform list, got errors: %v", result.Errors)
		}
	})

	t.Run("default policy", func(t *testing.T) {
		policy := DefaultVerificationPolicy()

		if policy.AllowDebugMode {
			t.Error("default policy should not allow debug mode")
		}

		if len(policy.AllowedPlatforms) != 3 {
			t.Errorf("expected 3 allowed platforms, got %d", len(policy.AllowedPlatforms))
		}
	})

	t.Run("permissive policy", func(t *testing.T) {
		policy := PermissiveVerificationPolicy()

		if !policy.AllowDebugMode {
			t.Error("permissive policy should allow debug mode")
		}

		// Should include simulated
		hasSimulated := false
		for _, p := range policy.AllowedPlatforms {
			if p == AttestationTypeSimulated {
				hasSimulated = true
				break
			}
		}
		if !hasSimulated {
			t.Error("permissive policy should allow simulated platform")
		}
	})
}

func TestAttestationTypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected AttestationType
	}{
		{
			name:     "SGX quote",
			data:     CreateTestSGXQuote(make([]byte, 32), make([]byte, 32), false, nil),
			expected: AttestationTypeSGX,
		},
		{
			name:     "SEV-SNP report",
			data:     CreateTestSEVSNPReport(make([]byte, 48), false, nil),
			expected: AttestationTypeSEVSNP,
		},
		{
			name:     "Nitro document",
			data:     CreateTestNitroDocument(make([]byte, 48), nil),
			expected: AttestationTypeNitro,
		},
		{
			name:     "Simulated attestation",
			data:     CreateTestSimulatedAttestation(make([]byte, 32), nil),
			expected: AttestationTypeSimulated,
		},
		{
			name:     "Unknown data",
			data:     []byte{0xFF, 0xFE, 0xFD},
			expected: AttestationTypeUnknown,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: AttestationTypeUnknown,
		},
		{
			name:     "Too short",
			data:     []byte{0x03},
			expected: AttestationTypeUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			detected := DetectAttestationType(tc.data)
			if detected != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, detected)
			}
		})
	}
}

func TestUniversalVerifier(t *testing.T) {
	allowlist := NewMeasurementAllowlist()

	// Add trusted measurements for each platform
	sgxMeasurement := make([]byte, 32)
	sevMeasurement := make([]byte, 48)
	nitroMeasurement := make([]byte, 48)
	simMeasurement := make([]byte, 32)

	for i := range sgxMeasurement {
		sgxMeasurement[i] = byte(i)
	}
	for i := range sevMeasurement {
		sevMeasurement[i] = byte(i + 10)
	}
	for i := range nitroMeasurement {
		nitroMeasurement[i] = byte(i + 20)
	}
	for i := range simMeasurement {
		simMeasurement[i] = byte(i + 30)
	}

	_ = allowlist.AddMeasurement(AttestationTypeSGX, sgxMeasurement, "sgx")
	_ = allowlist.AddMeasurement(AttestationTypeSEVSNP, sevMeasurement, "sev")
	_ = allowlist.AddMeasurement(AttestationTypeNitro, nitroMeasurement, "nitro")
	_ = allowlist.AddMeasurement(AttestationTypeSimulated, simMeasurement, "sim")

	verifier := NewUniversalAttestationVerifier(allowlist)

	t.Run("auto-detect SGX", func(t *testing.T) {
		quote := CreateTestSGXQuote(sgxMeasurement, make([]byte, 32), false, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(quote, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.AttestationType != AttestationTypeSGX {
			t.Errorf("expected SGX, got %v", result.AttestationType)
		}
	})

	t.Run("auto-detect SEV-SNP", func(t *testing.T) {
		report := CreateTestSEVSNPReport(sevMeasurement, false, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(report, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.AttestationType != AttestationTypeSEVSNP {
			t.Errorf("expected SEV-SNP, got %v", result.AttestationType)
		}
	})

	t.Run("auto-detect Nitro", func(t *testing.T) {
		doc := CreateTestNitroDocument(nitroMeasurement, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(doc, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.AttestationType != AttestationTypeNitro {
			t.Errorf("expected Nitro, got %v", result.AttestationType)
		}
	})

	t.Run("auto-detect Simulated", func(t *testing.T) {
		att := CreateTestSimulatedAttestation(simMeasurement, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(att, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.AttestationType != AttestationTypeSimulated {
			t.Errorf("expected Simulated, got %v", result.AttestationType)
		}
	})

	t.Run("unknown attestation type", func(t *testing.T) {
		unknown := []byte{0xFF, 0xFE, 0xFD, 0xFC}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(unknown, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for unknown type")
		}

		if result.AttestationType != AttestationTypeUnknown {
			t.Errorf("expected Unknown, got %v", result.AttestationType)
		}
	})

	t.Run("verify multiple attestations", func(t *testing.T) {
		attestations := [][]byte{
			CreateTestSGXQuote(sgxMeasurement, make([]byte, 32), false, nil),
			CreateTestSEVSNPReport(sevMeasurement, false, nil),
			CreateTestNitroDocument(nitroMeasurement, nil),
		}

		policy := PermissiveVerificationPolicy()
		results, err := verifier.VerifyMultiple(attestations, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		expectedTypes := []AttestationType{AttestationTypeSGX, AttestationTypeSEVSNP, AttestationTypeNitro}
		for i, result := range results {
			if result.AttestationType != expectedTypes[i] {
				t.Errorf("result %d: expected %v, got %v", i, expectedTypes[i], result.AttestationType)
			}
		}
	})

	t.Run("get specific verifier", func(t *testing.T) {
		sgxVerifier := verifier.GetVerifier(AttestationTypeSGX)
		if sgxVerifier == nil {
			t.Error("expected SGX verifier")
		}
		if sgxVerifier.Type() != AttestationTypeSGX {
			t.Errorf("expected SGX type, got %v", sgxVerifier.Type())
		}

		unknownVerifier := verifier.GetVerifier(AttestationTypeUnknown)
		if unknownVerifier != nil {
			t.Error("expected nil for unknown verifier")
		}
	})
}

func TestMeasurementAllowlistManager(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()
		manager := NewMeasurementAllowlistManager(allowlist)

		measurement := []byte{0x01, 0x02, 0x03}
		if err := manager.AddMeasurement(AttestationTypeSGX, measurement, "test"); err != nil {
			t.Fatalf("failed to add: %v", err)
		}

		if !manager.IsTrusted(AttestationTypeSGX, measurement) {
			t.Error("measurement should be trusted")
		}

		list := manager.ListMeasurements(AttestationTypeSGX)
		if len(list) != 1 {
			t.Errorf("expected 1 measurement, got %d", len(list))
		}

		if err := manager.RemoveMeasurement(AttestationTypeSGX, measurement); err != nil {
			t.Fatalf("failed to remove: %v", err)
		}

		if manager.IsTrusted(AttestationTypeSGX, measurement) {
			t.Error("measurement should not be trusted after removal")
		}
	})

	t.Run("auto-save", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "manager-allowlist.json")

		allowlist := NewMeasurementAllowlist()
		manager := NewMeasurementAllowlistManager(allowlist)
		manager.SetFilePath(filePath)
		manager.SetAutoSave(true)

		measurement := []byte{0x01, 0x02, 0x03}
		if err := manager.AddMeasurement(AttestationTypeSGX, measurement, "auto-save test"); err != nil {
			t.Fatalf("failed to add: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("file should have been auto-saved")
		}

		// Load in new manager
		newAllowlist := NewMeasurementAllowlist()
		newManager := NewMeasurementAllowlistManager(newAllowlist)
		if err := newManager.LoadFromJSON(filePath); err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if !newManager.IsTrusted(AttestationTypeSGX, measurement) {
			t.Error("measurement should be trusted after load")
		}
	})

	t.Run("get underlying allowlist", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()
		manager := NewMeasurementAllowlistManager(allowlist)

		if manager.GetAllowlist() != allowlist {
			t.Error("expected same allowlist reference")
		}
	})

	t.Run("save/load without path fails", func(t *testing.T) {
		allowlist := NewMeasurementAllowlist()
		manager := NewMeasurementAllowlistManager(allowlist)

		if err := manager.SaveToJSON(""); err == nil {
			t.Error("expected error when saving without path")
		}

		if err := manager.LoadFromJSON(""); err == nil {
			t.Error("expected error when loading without path")
		}
	})
}

func TestAttestationTypeString(t *testing.T) {
	tests := []struct {
		attType  AttestationType
		expected string
	}{
		{AttestationTypeSGX, "SGX"},
		{AttestationTypeSEVSNP, "SEV-SNP"},
		{AttestationTypeNitro, "NITRO"},
		{AttestationTypeSimulated, "SIMULATED"},
		{AttestationTypeUnknown, "UNKNOWN"},
		{AttestationType(99), "UNKNOWN"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.attType.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, tc.attType.String())
			}
		})
	}
}

func TestVerificationResultHelpers(t *testing.T) {
	t.Run("add error invalidates result", func(t *testing.T) {
		result := &VerificationResult{Valid: true}
		result.AddError("test error %d", 42)

		if result.Valid {
			t.Error("result should be invalid after adding error")
		}

		if len(result.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(result.Errors))
		}

		if result.Errors[0] != "test error 42" {
			t.Errorf("unexpected error message: %s", result.Errors[0])
		}
	})

	t.Run("add warning keeps result valid", func(t *testing.T) {
		result := &VerificationResult{Valid: true}
		result.AddWarning("test warning %s", "info")

		if !result.Valid {
			t.Error("result should still be valid after adding warning")
		}

		if len(result.Warnings) != 1 {
			t.Errorf("expected 1 warning, got %d", len(result.Warnings))
		}
	})
}

func TestSimulatedVerifier(t *testing.T) {
	allowlist := NewMeasurementAllowlist()
	measurement := make([]byte, 32)
	for i := range measurement {
		measurement[i] = byte(i)
	}
	_ = allowlist.AddMeasurement(AttestationTypeSimulated, measurement, "test sim")

	verifier := NewSimulatedVerifier(allowlist)

	t.Run("valid simulated attestation", func(t *testing.T) {
		att := CreateTestSimulatedAttestation(measurement, nil)

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(att, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}

		if result.AttestationType != AttestationTypeSimulated {
			t.Errorf("expected Simulated, got %v", result.AttestationType)
		}

		if !result.DebugMode {
			t.Error("simulated should always be in debug mode")
		}

		if result.SecurityLevel != 1 {
			t.Errorf("expected security level 1, got %d", result.SecurityLevel)
		}
	})

	t.Run("reject with strict policy", func(t *testing.T) {
		att := CreateTestSimulatedAttestation(measurement, nil)

		policy := DefaultVerificationPolicy()
		result, err := verifier.Verify(att, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result with strict policy (debug mode)")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		invalidAtt := []byte{0x00, 0x00, 0x00}

		policy := PermissiveVerificationPolicy()
		result, err := verifier.Verify(invalidAtt, nil, policy)
		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}

		if result.Valid {
			t.Error("expected invalid result for wrong format")
		}
	})

	t.Run("verifier type", func(t *testing.T) {
		if verifier.Type() != AttestationTypeSimulated {
			t.Errorf("expected Simulated, got %v", verifier.Type())
		}
	})
}

func TestConcurrentAllowlistAccess(t *testing.T) {
	allowlist := NewMeasurementAllowlist()
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			measurement := []byte{byte(id), byte(id + 1), byte(id + 2)}
			_ = allowlist.AddMeasurement(AttestationType(id%4+1), measurement, "concurrent test")
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			measurement := []byte{byte(id), byte(id + 1), byte(id + 2)}
			_ = allowlist.IsTrusted(AttestationType(id%4+1), measurement)
			_ = allowlist.ListMeasurements(AttestationType(id%4 + 1))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestMeasurementExpiration(t *testing.T) {
	allowlist := NewMeasurementAllowlist()

	measurement := []byte{0x01, 0x02, 0x03}
	if err := allowlist.AddMeasurement(AttestationTypeSGX, measurement, "expiring"); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	// Manually set expiration in the past
	allowlist.mu.Lock()
	key := hex.EncodeToString(measurement)
	meas := allowlist.measurements[AttestationTypeSGX][key]
	past := time.Now().Add(-1 * time.Hour)
	meas.ExpiresAt = &past
	allowlist.measurements[AttestationTypeSGX][key] = meas
	allowlist.mu.Unlock()

	// Should not be trusted anymore
	if allowlist.IsTrusted(AttestationTypeSGX, measurement) {
		t.Error("expired measurement should not be trusted")
	}
}

// containsSubstring is a helper to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

