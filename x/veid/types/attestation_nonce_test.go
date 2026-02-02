package types

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"
)

// testFingerprint is a test fingerprint for attestation nonce tests
const testFingerprint = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

// ============================================================================
// NonceRecord Tests
// ============================================================================

func TestNonceRecord_Create(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(
		nonce,
		fingerprint,
		AttestationTypeFacialVerification,
		now,
		3600, // 1 hour window
	)

	if record.Status != NonceStatusUnused {
		t.Errorf("expected unused status, got %s", record.Status)
	}

	if len(record.NonceHash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(record.NonceHash))
	}

	expectedExpiry := now.Add(1 * time.Hour)
	if !record.ExpiresAt.Equal(expectedExpiry) {
		t.Errorf("unexpected expiry time: %v", record.ExpiresAt)
	}
}

func TestNonceRecord_Validate_Valid(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	err := record.Validate()
	if err != nil {
		t.Errorf("expected valid record: %v", err)
	}
}

func TestNonceRecord_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*NonceRecord)
	}{
		{
			name:   "empty nonce hash",
			modify: func(r *NonceRecord) { r.NonceHash = "" },
		},
		{
			name:   "short nonce hash",
			modify: func(r *NonceRecord) { r.NonceHash = "abc123" },
		},
		{
			name:   "invalid hex hash",
			modify: func(r *NonceRecord) { r.NonceHash = "ghij" + r.NonceHash[4:] },
		},
		{
			name:   "empty issuer fingerprint",
			modify: func(r *NonceRecord) { r.IssuerFingerprint = "" },
		},
		{
			name:   "invalid attestation type",
			modify: func(r *NonceRecord) { r.AttestationType = AttestationType("invalid") },
		},
		{
			name:   "zero created_at",
			modify: func(r *NonceRecord) { r.CreatedAt = time.Time{} },
		},
		{
			name:   "zero expires_at",
			modify: func(r *NonceRecord) { r.ExpiresAt = time.Time{} },
		},
		{
			name:   "expires before created",
			modify: func(r *NonceRecord) { r.ExpiresAt = r.CreatedAt.Add(-1 * time.Hour) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nonce := make([]byte, 32)
			_, _ = rand.Read(nonce)
			now := time.Now().UTC()
			fingerprint := testFingerprint

			record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)
			tc.modify(record)

			err := record.Validate()
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestNonceRecord_MarkUsed(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	// Can be used initially
	if !record.CanBeUsed(now.Add(30 * time.Minute)) {
		t.Error("nonce should be usable within window")
	}

	// Mark as used
	usedAt := now.Add(10 * time.Minute)
	err := record.MarkUsed(usedAt, "attestation-001", 12345)
	if err != nil {
		t.Fatalf("failed to mark used: %v", err)
	}

	if record.Status != NonceStatusUsed {
		t.Errorf("expected used status, got %s", record.Status)
	}

	if record.UsedAt == nil || !record.UsedAt.Equal(usedAt) {
		t.Error("used_at not set correctly")
	}

	if record.AttestationID != "attestation-001" {
		t.Errorf("unexpected attestation ID: %s", record.AttestationID)
	}

	// Cannot be used again
	if record.CanBeUsed(now.Add(15 * time.Minute)) {
		t.Error("used nonce should not be usable")
	}

	err = record.MarkUsed(now.Add(15*time.Minute), "attestation-002", 12346)
	if err == nil {
		t.Error("should not be able to mark used nonce as used again")
	}
}

func TestNonceRecord_MarkUsed_Expired(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	// Try to use after expiry
	err := record.MarkUsed(now.Add(2*time.Hour), "attestation-001", 12345)
	if err == nil {
		t.Error("should not be able to use expired nonce")
	}
}

func TestNonceRecord_MarkUsed_AlreadyExpired(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)
	record.MarkExpired()

	err := record.MarkUsed(now.Add(30*time.Minute), "attestation-001", 12345)
	if err == nil {
		t.Error("should not be able to use already expired nonce")
	}
}

func TestNonceRecord_IsExpired(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	// Not expired within window
	if record.IsExpired(now.Add(30 * time.Minute)) {
		t.Error("nonce should not be expired within window")
	}

	// Expired after window
	if !record.IsExpired(now.Add(2 * time.Hour)) {
		t.Error("nonce should be expired after window")
	}
}

func TestNonceRecord_CanBeUsed(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	// Can be used within window
	if !record.CanBeUsed(now.Add(30 * time.Minute)) {
		t.Error("unused nonce within window should be usable")
	}

	// Cannot be used after expiry
	if record.CanBeUsed(now.Add(2 * time.Hour)) {
		t.Error("expired nonce should not be usable")
	}
}

// ============================================================================
// ReplayProtectionPolicy Tests
// ============================================================================

func TestDefaultReplayProtectionPolicy(t *testing.T) {
	policy := DefaultReplayProtectionPolicy()

	err := policy.Validate()
	if err != nil {
		t.Errorf("default policy should be valid: %v", err)
	}

	if policy.NonceWindowSeconds != NonceWindowDefaultSeconds {
		t.Errorf("unexpected nonce window: %d", policy.NonceWindowSeconds)
	}

	if !policy.RequireTimestampBinding {
		t.Error("timestamp binding should be required by default")
	}

	if !policy.RequireIssuerBinding {
		t.Error("issuer binding should be required by default")
	}
}

func TestReplayProtectionPolicy_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*ReplayProtectionPolicy)
	}{
		{
			name:   "window too short",
			modify: func(p *ReplayProtectionPolicy) { p.NonceWindowSeconds = 60 },
		},
		{
			name:   "window too long",
			modify: func(p *ReplayProtectionPolicy) { p.NonceWindowSeconds = 100000 },
		},
		{
			name:   "negative clock skew",
			modify: func(p *ReplayProtectionPolicy) { p.MaxClockSkewSeconds = -1 },
		},
		{
			name:   "clock skew too large",
			modify: func(p *ReplayProtectionPolicy) { p.MaxClockSkewSeconds = p.NonceWindowSeconds },
		},
		{
			name:   "zero max nonces",
			modify: func(p *ReplayProtectionPolicy) { p.MaxNoncesPerIssuer = 0 },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := DefaultReplayProtectionPolicy()
			tc.modify(&policy)

			err := policy.Validate()
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

// ============================================================================
// ValidateNonce Tests
// ============================================================================

func TestValidateNonce_Valid(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	err := ValidateNonce(nonce)
	if err != nil {
		t.Errorf("expected valid nonce: %v", err)
	}
}

func TestValidateNonce_TooShort(t *testing.T) {
	nonce := make([]byte, 8)
	_, _ = rand.Read(nonce)

	err := ValidateNonce(nonce)
	if err == nil {
		t.Error("expected error for short nonce")
	}
}

func TestValidateNonce_TooLong(t *testing.T) {
	nonce := make([]byte, 128)
	_, _ = rand.Read(nonce)

	err := ValidateNonce(nonce)
	if err == nil {
		t.Error("expected error for long nonce")
	}
}

func TestValidateNonce_AllZeros(t *testing.T) {
	nonce := make([]byte, 32) // All zeros

	err := ValidateNonce(nonce)
	if err == nil {
		t.Error("expected error for all-zeros nonce")
	}
}

func TestValidateNonce_AllOnes(t *testing.T) {
	nonce := make([]byte, 32)
	for i := range nonce {
		nonce[i] = 0xFF
	}

	err := ValidateNonce(nonce)
	if err == nil {
		t.Error("expected error for all-ones nonce")
	}
}

func TestValidateNonceHex_Valid(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	nonceHex := hex.EncodeToString(nonce)

	err := ValidateNonceHex(nonceHex)
	if err != nil {
		t.Errorf("expected valid hex nonce: %v", err)
	}
}

func TestValidateNonceHex_InvalidHex(t *testing.T) {
	err := ValidateNonceHex("not-valid-hex!!!")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

// ============================================================================
// ValidateTimestamp Tests
// ============================================================================

func TestValidateTimestamp_Valid(t *testing.T) {
	now := time.Now().UTC()
	policy := DefaultReplayProtectionPolicy()

	// Timestamp at now is valid
	err := ValidateTimestamp(now, now, policy)
	if err != nil {
		t.Errorf("current timestamp should be valid: %v", err)
	}

	// Slight past is valid
	err = ValidateTimestamp(now.Add(-30*time.Minute), now, policy)
	if err != nil {
		t.Errorf("recent past timestamp should be valid: %v", err)
	}

	// Slight future (within clock skew) is valid
	err = ValidateTimestamp(now.Add(1*time.Minute), now, policy)
	if err != nil {
		t.Errorf("near future timestamp should be valid: %v", err)
	}
}

func TestValidateTimestamp_TooFuture(t *testing.T) {
	now := time.Now().UTC()
	policy := DefaultReplayProtectionPolicy()

	err := ValidateTimestamp(now.Add(10*time.Minute), now, policy)
	if err == nil {
		t.Error("expected error for far future timestamp")
	}
}

func TestValidateTimestamp_TooPast(t *testing.T) {
	now := time.Now().UTC()
	policy := DefaultReplayProtectionPolicy()

	err := ValidateTimestamp(now.Add(-2*time.Hour), now, policy)
	if err == nil {
		t.Error("expected error for far past timestamp")
	}
}

func TestValidateTimestamp_BindingDisabled(t *testing.T) {
	now := time.Now().UTC()
	policy := DefaultReplayProtectionPolicy()
	policy.RequireTimestampBinding = false

	// Any timestamp is valid when binding is disabled
	err := ValidateTimestamp(now.Add(-100*time.Hour), now, policy)
	if err != nil {
		t.Error("timestamp should be valid when binding is disabled")
	}
}

// ============================================================================
// ComputeNonceHash Tests
// ============================================================================

func TestComputeNonceHash(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	hash := ComputeNonceHash(nonce)

	if len(hash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(hash))
	}

	// Verify it's valid hex
	_, err := hex.DecodeString(hash)
	if err != nil {
		t.Errorf("hash should be valid hex: %v", err)
	}

	// Deterministic
	hash2 := ComputeNonceHash(nonce)
	if hash != hash2 {
		t.Error("hash should be deterministic")
	}

	// Different input produces different output
	nonce2 := make([]byte, 32)
	_, _ = rand.Read(nonce2)
	hash3 := ComputeNonceHash(nonce2)
	if hash == hash3 {
		t.Error("different nonces should have different hashes")
	}
}

// ============================================================================
// NonceHistoryEntry Tests
// ============================================================================

func TestNewNonceHistoryEntry(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)
	_ = record.MarkUsed(now.Add(10*time.Minute), "attestation-001", 12345)

	entry := NewNonceHistoryEntry(record)

	if entry == nil {
		t.Fatal("expected non-nil entry for used nonce")
	}

	if entry.NonceHash != record.NonceHash {
		t.Error("nonce hash mismatch")
	}

	if entry.AttestationID != "attestation-001" {
		t.Errorf("unexpected attestation ID: %s", entry.AttestationID)
	}

	if entry.BlockHeight != 12345 {
		t.Errorf("unexpected block height: %d", entry.BlockHeight)
	}

	// Fingerprint should be truncated
	if len(entry.IssuerFingerprint) > 16 {
		t.Errorf("fingerprint should be truncated to 16 chars, got %d", len(entry.IssuerFingerprint))
	}
}

func TestNewNonceHistoryEntry_UnusedNonce(t *testing.T) {
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	now := time.Now().UTC()
	fingerprint := testFingerprint

	record := NewNonceRecord(nonce, fingerprint, AttestationTypeFacialVerification, now, 3600)

	entry := NewNonceHistoryEntry(record)

	if entry != nil {
		t.Error("expected nil entry for unused nonce")
	}
}

// ============================================================================
// NonceValidationResult Tests
// ============================================================================

func TestNonceValidationResult_Success(t *testing.T) {
	result := NewNonceValidationResultSuccess("abc123...")

	if !result.Valid {
		t.Error("success result should be valid")
	}

	if result.Error != "" {
		t.Error("success result should have no error")
	}

	if !result.TimestampValid {
		t.Error("timestamp should be valid in success result")
	}
}

func TestNonceValidationResult_Failure(t *testing.T) {
	result := NewNonceValidationResultFailure("abc123...", "nonce already used")

	if result.Valid {
		t.Error("failure result should not be valid")
	}

	if result.Error == "" {
		t.Error("failure result should have error message")
	}
}
