package types

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"
)

// ============================================================================
// Test Helpers
// ============================================================================

//nolint:unparam // length kept for future variable-length nonce tests
func generateTestNonce(t *testing.T, _ int) []byte {
	t.Helper()
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
	if err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}
	return nonce
}

func generateTestFingerprint(t *testing.T) string {
	t.Helper()
	fp := make([]byte, 32)
	_, err := rand.Read(fp)
	if err != nil {
		t.Fatalf("failed to generate fingerprint: %v", err)
	}
	return hex.EncodeToString(fp)
}

// ============================================================================
// VerificationAttestation Tests
// ============================================================================

func TestVerificationAttestation_Validate_Valid(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	// Add a proof
	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		issuer.ID+"#keys-1",
		[]byte("test-signature-bytes"),
		hex.EncodeToString(nonce),
	)
	attestation.SetProof(proof)

	err := attestation.Validate()
	if err != nil {
		t.Errorf("expected valid attestation, got error: %v", err)
	}
}

func TestVerificationAttestation_Validate_InvalidType(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationType("invalid_type"),
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		issuer.ID+"#keys-1",
		[]byte("test-signature-bytes"),
		hex.EncodeToString(nonce),
	)
	attestation.SetProof(proof)

	err := attestation.Validate()
	if err == nil {
		t.Error("expected error for invalid attestation type")
	}
}

func TestVerificationAttestation_Validate_ExpiredBeforeIssued(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	// Set expires before issued
	attestation.ExpiresAt = now.Add(-1 * time.Hour)

	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		issuer.ID+"#keys-1",
		[]byte("test-signature-bytes"),
		hex.EncodeToString(nonce),
	)
	attestation.SetProof(proof)

	err := attestation.Validate()
	if err == nil {
		t.Error("expected error for expires_at before issued_at")
	}
}

func TestVerificationAttestation_Validate_ScoreExceeds100(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		150, // Invalid score
		90,
	)

	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		issuer.ID+"#keys-1",
		[]byte("test-signature-bytes"),
		hex.EncodeToString(nonce),
	)
	attestation.SetProof(proof)

	err := attestation.Validate()
	if err == nil {
		t.Error("expected error for score exceeding 100")
	}
}

func TestVerificationAttestation_CanonicalBytes(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	// Canonical bytes should be deterministic
	bytes1, err := attestation.CanonicalBytes()
	if err != nil {
		t.Fatalf("failed to get canonical bytes: %v", err)
	}

	bytes2, err := attestation.CanonicalBytes()
	if err != nil {
		t.Fatalf("failed to get canonical bytes: %v", err)
	}

	if string(bytes1) != string(bytes2) {
		t.Error("canonical bytes should be deterministic")
	}
}

func TestVerificationAttestation_Hash(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	hash, err := attestation.Hash()
	if err != nil {
		t.Fatalf("failed to compute hash: %v", err)
	}

	if len(hash) != 32 {
		t.Errorf("expected SHA256 hash length 32, got %d", len(hash))
	}

	// Hash should be deterministic
	hash2, err := attestation.Hash()
	if err != nil {
		t.Fatalf("failed to compute hash: %v", err)
	}

	if string(hash) != string(hash2) {
		t.Error("hash should be deterministic")
	}
}

func TestVerificationAttestation_IsExpired(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)
	fingerprint := generateTestFingerprint(t)

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")
	subject := NewAttestationSubject("virtengine1account...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		1*time.Hour,
		85,
		90,
	)

	// Should not be expired immediately
	if attestation.IsExpired(now) {
		t.Error("attestation should not be expired at issuance time")
	}

	// Should be expired after expiry
	if !attestation.IsExpired(now.Add(2 * time.Hour)) {
		t.Error("attestation should be expired after expiry time")
	}
}

func TestVerificationAttestation_ToScopeType(t *testing.T) {
	tests := []struct {
		attestationType AttestationType
		expectedScope   ScopeType
	}{
		{AttestationTypeFacialVerification, ScopeTypeSelfie},
		{AttestationTypeLivenessCheck, ScopeTypeFaceVideo},
		{AttestationTypeDocumentVerification, ScopeTypeIDDocument},
		{AttestationTypeEmailVerification, ScopeTypeEmailProof},
		{AttestationTypeSMSVerification, ScopeTypeSMSProof},
		{AttestationTypeDomainVerification, ScopeTypeDomainVerify},
		{AttestationTypeSSOVerification, ScopeTypeSSOMetadata},
		{AttestationTypeBiometricVerification, ScopeTypeBiometric},
	}

	for _, tc := range tests {
		t.Run(string(tc.attestationType), func(t *testing.T) {
			now := time.Now().UTC()
			nonce := generateTestNonce(t, 32)
			fingerprint := generateTestFingerprint(t)

			issuer := NewAttestationIssuer(fingerprint, "")
			subject := NewAttestationSubject("virtengine1account...")

			attestation := NewVerificationAttestation(
				issuer,
				subject,
				tc.attestationType,
				nonce,
				now,
				24*time.Hour,
				85,
				90,
			)

			if attestation.ToScopeType() != tc.expectedScope {
				t.Errorf("expected scope type %s, got %s", tc.expectedScope, attestation.ToScopeType())
			}
		})
	}
}

// ============================================================================
// AttestationIssuer Tests
// ============================================================================

func TestAttestationIssuer_Validate_Valid(t *testing.T) {
	fingerprint := generateTestFingerprint(t)
	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator...")

	err := issuer.Validate()
	if err != nil {
		t.Errorf("expected valid issuer, got error: %v", err)
	}
}

func TestAttestationIssuer_Validate_EmptyFingerprint(t *testing.T) {
	issuer := AttestationIssuer{
		ID:             "did:virtengine:validator:test",
		KeyFingerprint: "",
	}

	err := issuer.Validate()
	if err == nil {
		t.Error("expected error for empty fingerprint")
	}
}

func TestAttestationIssuer_Validate_ShortFingerprint(t *testing.T) {
	issuer := AttestationIssuer{
		ID:             "did:virtengine:validator:test",
		KeyFingerprint: "abcd1234", // Too short
	}

	err := issuer.Validate()
	if err == nil {
		t.Error("expected error for short fingerprint")
	}
}

func TestAttestationIssuer_Validate_InvalidHex(t *testing.T) {
	issuer := AttestationIssuer{
		ID:             "did:virtengine:validator:test",
		KeyFingerprint: "ghij" + generateTestFingerprint(t)[4:], // Invalid hex chars
	}

	err := issuer.Validate()
	if err == nil {
		t.Error("expected error for invalid hex fingerprint")
	}
}

// ============================================================================
// AttestationSubject Tests
// ============================================================================

func TestAttestationSubject_Validate_Valid(t *testing.T) {
	subject := NewAttestationSubject("virtengine1account...")

	err := subject.Validate()
	if err != nil {
		t.Errorf("expected valid subject, got error: %v", err)
	}
}

func TestAttestationSubject_Validate_EmptyAddress(t *testing.T) {
	subject := AttestationSubject{
		ID:             "did:virtengine:test",
		AccountAddress: "",
	}

	err := subject.Validate()
	if err == nil {
		t.Error("expected error for empty account address")
	}
}

// ============================================================================
// VerificationProofDetail Tests
// ============================================================================

func TestVerificationProofDetail_Validate_Valid(t *testing.T) {
	proof := NewVerificationProofDetail(
		"facial_match",
		"abc123...",
		85,
		70,
		time.Now().UTC(),
	)

	err := proof.Validate()
	if err != nil {
		t.Errorf("expected valid proof detail, got error: %v", err)
	}

	if !proof.Passed {
		t.Error("proof should pass when score >= threshold")
	}
}

func TestVerificationProofDetail_Validate_ScoreExceeds100(t *testing.T) {
	proof := NewVerificationProofDetail(
		"facial_match",
		"abc123...",
		150, // Invalid
		70,
		time.Now().UTC(),
	)

	err := proof.Validate()
	if err == nil {
		t.Error("expected error for score exceeding 100")
	}
}

func TestVerificationProofDetail_Passed(t *testing.T) {
	// Score below threshold
	proofFail := NewVerificationProofDetail("test", "hash", 50, 70, time.Now())
	if proofFail.Passed {
		t.Error("proof should not pass when score < threshold")
	}

	// Score equals threshold
	proofEqual := NewVerificationProofDetail("test", "hash", 70, 70, time.Now())
	if !proofEqual.Passed {
		t.Error("proof should pass when score == threshold")
	}

	// Score above threshold
	proofPass := NewVerificationProofDetail("test", "hash", 90, 70, time.Now())
	if !proofPass.Passed {
		t.Error("proof should pass when score > threshold")
	}
}

// ============================================================================
// AttestationProof Tests
// ============================================================================

func TestAttestationProof_Validate_Valid(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)

	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		"did:virtengine:validator:test#keys-1",
		[]byte("signature-bytes"),
		hex.EncodeToString(nonce),
	)

	err := proof.Validate()
	if err != nil {
		t.Errorf("expected valid proof, got error: %v", err)
	}
}

func TestAttestationProof_Validate_InvalidType(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)

	proof := AttestationProof{
		Type:               AttestationProofType("invalid"),
		Created:            now,
		VerificationMethod: "did:virtengine:test#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         "dGVzdA==", // base64 "test"
		Nonce:              hex.EncodeToString(nonce),
	}

	err := proof.Validate()
	if err == nil {
		t.Error("expected error for invalid proof type")
	}
}

func TestAttestationProof_Validate_EmptyProofValue(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)

	proof := AttestationProof{
		Type:               ProofTypeEd25519,
		Created:            now,
		VerificationMethod: "did:virtengine:test#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         "",
		Nonce:              hex.EncodeToString(nonce),
	}

	err := proof.Validate()
	if err == nil {
		t.Error("expected error for empty proof value")
	}
}

func TestAttestationProof_Validate_InvalidBase64(t *testing.T) {
	now := time.Now().UTC()
	nonce := generateTestNonce(t, 32)

	proof := AttestationProof{
		Type:               ProofTypeEd25519,
		Created:            now,
		VerificationMethod: "did:virtengine:test#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         "not-valid-base64!!!",
		Nonce:              hex.EncodeToString(nonce),
	}

	err := proof.Validate()
	if err == nil {
		t.Error("expected error for invalid base64 proof value")
	}
}

// ============================================================================
// Attestation Type Tests
// ============================================================================

func TestIsValidAttestationType(t *testing.T) {
	validTypes := AllAttestationTypes()
	for _, at := range validTypes {
		if !IsValidAttestationType(at) {
			t.Errorf("expected %s to be valid", at)
		}
	}

	if IsValidAttestationType(AttestationType("invalid")) {
		t.Error("expected invalid type to be invalid")
	}
}

func TestAllProofTypes(t *testing.T) {
	proofTypes := AllProofTypes()
	if len(proofTypes) == 0 {
		t.Error("expected at least one proof type")
	}

	for _, pt := range proofTypes {
		if !IsValidProofType(pt) {
			t.Errorf("expected %s to be valid", pt)
		}
	}
}

// ============================================================================
// Example Attestation Fixtures
// ============================================================================

func TestExampleFacialVerificationAttestation(t *testing.T) {
	// This test demonstrates a complete facial verification attestation
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	nonce, _ := hex.DecodeString("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2")
	fingerprint := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	issuer := NewAttestationIssuer(fingerprint, "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu")
	issuer.KeyID = "validator-key-001"
	issuer.ServiceEndpoint = "https://veid-signer.virtengine.io/v1"

	subject := NewAttestationSubject("virtengine1abc123def456ghi789jkl012mno345pqr678stu")
	subject.ScopeID = "scope-facial-001"
	subject.RequestID = "req-20240115-001"

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeFacialVerification,
		nonce,
		now,
		30*24*time.Hour, // 30 days validity
		92,              // High score
		95,              // High confidence
	)

	attestation.ModelVersion = "facial-v2.3.0"
	attestation.SetMetadata("pipeline_version", "1.0.0")
	attestation.SetMetadata("validator_consensus", "5/5")

	// Add verification proofs
	attestation.AddVerificationProof(NewVerificationProofDetail(
		"facial_embedding_match",
		"sha256:abc123...",
		94,
		80,
		now,
	))
	attestation.AddVerificationProof(NewVerificationProofDetail(
		"liveness_score",
		"sha256:def456...",
		91,
		85,
		now,
	))

	// Add proof signature
	proof := NewAttestationProof(
		ProofTypeEd25519,
		now,
		issuer.ID+"#"+issuer.KeyID,
		[]byte("example-ed25519-signature-64-bytes-padded-for-test-purposes!!"),
		hex.EncodeToString(nonce),
	)
	proof.Domain = "veid.virtengine.io"
	attestation.SetProof(proof)

	// Validate
	err := attestation.Validate()
	if err != nil {
		t.Errorf("example attestation should be valid: %v", err)
	}

	// Test JSON serialization
	jsonBytes, err := attestation.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize to JSON: %v", err)
	}

	// Deserialize and validate
	parsed, err := AttestationFromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("failed to deserialize from JSON: %v", err)
	}

	if parsed.ID != attestation.ID {
		t.Error("parsed attestation ID mismatch")
	}

	if parsed.Score != attestation.Score {
		t.Error("parsed attestation score mismatch")
	}

	if len(parsed.VerificationProofs) != 2 {
		t.Errorf("expected 2 verification proofs, got %d", len(parsed.VerificationProofs))
	}

	t.Logf("Example attestation JSON:\n%s", string(jsonBytes))
}

func TestExampleDocumentVerificationAttestation(t *testing.T) {
	// Document verification attestation example
	now := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	nonce, _ := hex.DecodeString("b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3")
	fingerprint := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	issuer := NewAttestationIssuer(fingerprint, "virtengine1validator123...")
	subject := NewAttestationSubject("virtengine1user456...")

	attestation := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeDocumentVerification,
		nonce,
		now,
		365*24*time.Hour, // 1 year validity for document verification
		88,
		85,
	)

	attestation.AddVerificationProof(NewVerificationProofDetail(
		"document_authenticity",
		"sha256:doc123...",
		90,
		80,
		now,
	))
	attestation.AddVerificationProof(NewVerificationProofDetail(
		"ocr_extraction",
		"sha256:ocr456...",
		95,
		90,
		now,
	))
	attestation.AddVerificationProof(NewVerificationProofDetail(
		"document_face_match",
		"sha256:face789...",
		82,
		75,
		now,
	))

	proof := NewAttestationProof(
		ProofTypeSecp256k1,
		now,
		issuer.ID+"#keys-1",
		[]byte("example-secp256k1-signature"),
		hex.EncodeToString(nonce),
	)
	attestation.SetProof(proof)

	err := attestation.Validate()
	if err != nil {
		t.Errorf("example document attestation should be valid: %v", err)
	}

	// Verify scope type mapping
	if attestation.ToScopeType() != ScopeTypeIDDocument {
		t.Error("document attestation should map to id_document scope")
	}
}
