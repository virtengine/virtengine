package types

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"
)

// ============================================================================
// SignerKeyPolicy Tests
// ============================================================================

func TestDefaultSignerKeyPolicy(t *testing.T) {
	policy := DefaultSignerKeyPolicy()

	err := policy.Validate()
	if err != nil {
		t.Errorf("default policy should be valid: %v", err)
	}

	if policy.MaxKeyAgeSeconds != 7776000 {
		t.Errorf("expected max key age 7776000, got %d", policy.MaxKeyAgeSeconds)
	}

	if policy.RotationOverlapSeconds != 604800 {
		t.Errorf("expected rotation overlap 604800, got %d", policy.RotationOverlapSeconds)
	}

	if !policy.RequireSuccessorKey {
		t.Error("expected require_successor_key to be true by default")
	}
}

func TestSignerKeyPolicy_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*SignerKeyPolicy)
	}{
		{
			name:   "negative max key age",
			modify: func(p *SignerKeyPolicy) { p.MaxKeyAgeSeconds = -1 },
		},
		{
			name:   "negative overlap",
			modify: func(p *SignerKeyPolicy) { p.RotationOverlapSeconds = -1 },
		},
		{
			name:   "overlap exceeds max age",
			modify: func(p *SignerKeyPolicy) { p.RotationOverlapSeconds = p.MaxKeyAgeSeconds + 1 },
		},
		{
			name:   "no key algorithms",
			modify: func(p *SignerKeyPolicy) { p.KeyAlgorithms = nil },
		},
		{
			name:   "invalid algorithm",
			modify: func(p *SignerKeyPolicy) { p.KeyAlgorithms = []AttestationProofType{"invalid"} },
		},
		{
			name:   "weak key strength",
			modify: func(p *SignerKeyPolicy) { p.MinKeyStrength = 64 },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := DefaultSignerKeyPolicy()
			tc.modify(&policy)

			err := policy.Validate()
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestSignerKeyPolicy_IsAlgorithmAllowed(t *testing.T) {
	policy := DefaultSignerKeyPolicy()

	if !policy.IsAlgorithmAllowed(ProofTypeEd25519) {
		t.Error("Ed25519 should be allowed by default")
	}

	if !policy.IsAlgorithmAllowed(ProofTypeSecp256k1) {
		t.Error("Secp256k1 should be allowed by default")
	}

	if policy.IsAlgorithmAllowed(ProofTypeSr25519) {
		t.Error("Sr25519 should not be allowed by default")
	}
}

// ============================================================================
// SignerKeyInfo Tests
// ============================================================================

func TestSignerKeyInfo_Create(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	if keyInfo.State != SignerKeyStatePending {
		t.Errorf("new key should be in pending state, got %s", keyInfo.State)
	}

	if keyInfo.KeyID != "validator-001:1" {
		t.Errorf("unexpected key ID: %s", keyInfo.KeyID)
	}

	if len(keyInfo.Fingerprint) != 64 {
		t.Errorf("expected 64-char fingerprint, got %d", len(keyInfo.Fingerprint))
	}
}

func TestSignerKeyInfo_Validate_Valid(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	err := keyInfo.Validate()
	if err != nil {
		t.Errorf("expected valid key info: %v", err)
	}
}

func TestSignerKeyInfo_Validate_FingerprintMismatch(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	// Corrupt the fingerprint
	keyInfo.Fingerprint = "0000000000000000000000000000000000000000000000000000000000000000"

	err := keyInfo.Validate()
	if err == nil {
		t.Error("expected error for fingerprint mismatch")
	}
}

func TestSignerKeyInfo_Lifecycle(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	// Initial state
	if keyInfo.IsActive() {
		t.Error("pending key should not be active")
	}

	// Activate
	expiresAt := now.Add(90 * 24 * time.Hour)
	err := keyInfo.Activate(now, expiresAt)
	if err != nil {
		t.Fatalf("failed to activate: %v", err)
	}

	if keyInfo.State != SignerKeyStateActive {
		t.Errorf("expected active state, got %s", keyInfo.State)
	}

	if !keyInfo.IsActive() {
		t.Error("active key should be active")
	}

	// Cannot activate again
	err = keyInfo.Activate(now, expiresAt)
	if err == nil {
		t.Error("should not be able to activate non-pending key")
	}

	// Start rotation
	err = keyInfo.StartRotation("validator-001:2")
	if err != nil {
		t.Fatalf("failed to start rotation: %v", err)
	}

	if keyInfo.State != SignerKeyStateRotating {
		t.Errorf("expected rotating state, got %s", keyInfo.State)
	}

	// Still active during rotation
	if !keyInfo.IsActive() {
		t.Error("rotating key should still be active")
	}

	// Revoke
	err = keyInfo.Revoke(now, RevocationReasonRotation)
	if err != nil {
		t.Fatalf("failed to revoke: %v", err)
	}

	if keyInfo.State != SignerKeyStateRevoked {
		t.Errorf("expected revoked state, got %s", keyInfo.State)
	}

	if keyInfo.IsActive() {
		t.Error("revoked key should not be active")
	}

	// Cannot revoke again
	err = keyInfo.Revoke(now, RevocationReasonCompromised)
	if err == nil {
		t.Error("should not be able to revoke already revoked key")
	}
}

func TestSignerKeyInfo_ShouldRotate(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	policy := DefaultSignerKeyPolicy()

	// Pending key should not rotate
	if keyInfo.ShouldRotate(now, policy) {
		t.Error("pending key should not need rotation")
	}

	// Activate key
	expiresAt := now.Add(90 * 24 * time.Hour)
	keyInfo.Activate(now, expiresAt)

	// Recently activated key should not need rotation
	if keyInfo.ShouldRotate(now.Add(1*time.Hour), policy) {
		t.Error("recently activated key should not need rotation")
	}

	// Key approaching expiry should need rotation
	// policy.MaxKeyAgeSeconds - policy.MinRotationNoticeSeconds = 87 days
	nearExpiry := now.Add(88 * 24 * time.Hour)
	if !keyInfo.ShouldRotate(nearExpiry, policy) {
		t.Error("key approaching expiry should need rotation")
	}
}

func TestSignerKeyInfo_IsExpired(t *testing.T) {
	now := time.Now().UTC()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	keyInfo := NewSignerKeyInfo(
		"validator-001",
		publicKey,
		ProofTypeEd25519,
		1,
		now,
	)

	// No expiry set
	if keyInfo.IsExpired(now.Add(100 * 365 * 24 * time.Hour)) {
		t.Error("key without expiry should not be expired")
	}

	// Set expiry
	expiresAt := now.Add(24 * time.Hour)
	keyInfo.Activate(now, expiresAt)

	if keyInfo.IsExpired(now) {
		t.Error("key should not be expired before expiry")
	}

	if !keyInfo.IsExpired(now.Add(25 * time.Hour)) {
		t.Error("key should be expired after expiry")
	}
}

// ============================================================================
// SignerKeyState Tests
// ============================================================================

func TestSignerKeyState_CanSign(t *testing.T) {
	tests := []struct {
		state   SignerKeyState
		canSign bool
	}{
		{SignerKeyStateActive, true},
		{SignerKeyStateRotating, true},
		{SignerKeyStatePending, false},
		{SignerKeyStateRevoked, false},
		{SignerKeyStateExpired, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.state), func(t *testing.T) {
			if tc.state.CanSign() != tc.canSign {
				t.Errorf("expected CanSign=%v for state %s", tc.canSign, tc.state)
			}
		})
	}
}

func TestSignerKeyState_CanVerify(t *testing.T) {
	tests := []struct {
		state     SignerKeyState
		canVerify bool
	}{
		{SignerKeyStateActive, true},
		{SignerKeyStateRotating, true},
		{SignerKeyStatePending, false},
		{SignerKeyStateRevoked, false},
		{SignerKeyStateExpired, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.state), func(t *testing.T) {
			if tc.state.CanVerify() != tc.canVerify {
				t.Errorf("expected CanVerify=%v for state %s", tc.canVerify, tc.state)
			}
		})
	}
}

// ============================================================================
// KeyRotationRecord Tests
// ============================================================================

func TestKeyRotationRecord_Create(t *testing.T) {
	now := time.Now().UTC()
	publicKey1 := make([]byte, 32)
	publicKey2 := make([]byte, 32)
	rand.Read(publicKey1)
	rand.Read(publicKey2)

	oldKey := NewSignerKeyInfo("validator-001", publicKey1, ProofTypeEd25519, 1, now.Add(-90*24*time.Hour))
	newKey := NewSignerKeyInfo("validator-001", publicKey2, ProofTypeEd25519, 2, now)

	record := NewKeyRotationRecord(
		"rotation-001",
		"validator-001",
		oldKey,
		newKey,
		now,
		604800, // 7 days overlap
		RevocationReasonRotation,
		"admin",
		12345,
	)

	if record.Status != RotationStatusInProgress {
		t.Errorf("expected in_progress status, got %s", record.Status)
	}

	expectedOverlapEnd := now.Add(7 * 24 * time.Hour)
	if !record.OverlapEndsAt.Equal(expectedOverlapEnd) {
		t.Errorf("unexpected overlap end time: %v", record.OverlapEndsAt)
	}
}

func TestKeyRotationRecord_IsOverlapPeriodActive(t *testing.T) {
	now := time.Now().UTC()
	publicKey1 := make([]byte, 32)
	publicKey2 := make([]byte, 32)
	rand.Read(publicKey1)
	rand.Read(publicKey2)

	oldKey := NewSignerKeyInfo("validator-001", publicKey1, ProofTypeEd25519, 1, now)
	newKey := NewSignerKeyInfo("validator-001", publicKey2, ProofTypeEd25519, 2, now)

	record := NewKeyRotationRecord(
		"rotation-001",
		"validator-001",
		oldKey,
		newKey,
		now,
		3600, // 1 hour overlap
		RevocationReasonRotation,
		"admin",
		12345,
	)

	// During overlap
	if !record.IsOverlapPeriodActive(now.Add(30 * time.Minute)) {
		t.Error("overlap should be active during overlap period")
	}

	// After overlap
	if record.IsOverlapPeriodActive(now.Add(2 * time.Hour)) {
		t.Error("overlap should not be active after overlap period")
	}

	// After completion
	record.Complete(now.Add(30 * time.Minute))
	if record.IsOverlapPeriodActive(now.Add(30 * time.Minute)) {
		t.Error("overlap should not be active after completion")
	}
}

// ============================================================================
// SignerRegistryEntry Tests
// ============================================================================

func TestSignerRegistryEntry_Create(t *testing.T) {
	now := time.Now().UTC()

	entry := NewSignerRegistryEntry(
		"signer-001",
		"Test Validator",
		"virtengine1validator...",
		"signer-001:1",
		now,
	)

	if !entry.Active {
		t.Error("new signer should be active")
	}

	if len(entry.KeyHistory) != 1 {
		t.Errorf("expected 1 key in history, got %d", len(entry.KeyHistory))
	}

	if entry.ActiveKeyID != "signer-001:1" {
		t.Errorf("unexpected active key ID: %s", entry.ActiveKeyID)
	}
}

func TestSignerRegistryEntry_RotateKey(t *testing.T) {
	now := time.Now().UTC()

	entry := NewSignerRegistryEntry(
		"signer-001",
		"Test Validator",
		"virtengine1validator...",
		"signer-001:1",
		now,
	)

	rotatedAt := now.Add(30 * 24 * time.Hour)
	entry.RotateKey("signer-001:2", rotatedAt)

	if entry.ActiveKeyID != "signer-001:2" {
		t.Errorf("expected new active key, got %s", entry.ActiveKeyID)
	}

	if len(entry.KeyHistory) != 2 {
		t.Errorf("expected 2 keys in history, got %d", len(entry.KeyHistory))
	}

	if entry.LastRotationAt == nil || !entry.LastRotationAt.Equal(rotatedAt) {
		t.Error("last rotation time not updated")
	}
}

func TestSignerRegistryEntry_Deactivate(t *testing.T) {
	now := time.Now().UTC()

	entry := NewSignerRegistryEntry(
		"signer-001",
		"Test Validator",
		"virtengine1validator...",
		"signer-001:1",
		now,
	)

	deactivatedAt := now.Add(24 * time.Hour)
	entry.Deactivate(deactivatedAt)

	if entry.Active {
		t.Error("signer should be inactive after deactivation")
	}

	if entry.DeactivatedAt == nil || !entry.DeactivatedAt.Equal(deactivatedAt) {
		t.Error("deactivated time not set")
	}
}

// ============================================================================
// Revocation Reason Tests
// ============================================================================

func TestIsValidRevocationReason(t *testing.T) {
	validReasons := AllRevocationReasons()
	for _, r := range validReasons {
		if !IsValidRevocationReason(r) {
			t.Errorf("expected %s to be valid", r)
		}
	}

	if IsValidRevocationReason(KeyRevocationReason("invalid")) {
		t.Error("expected invalid reason to be invalid")
	}
}

// ============================================================================
// ComputeKeyFingerprint Tests
// ============================================================================

func TestComputeKeyFingerprint(t *testing.T) {
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	fp := ComputeKeyFingerprint(publicKey)

	if len(fp) != 64 {
		t.Errorf("expected 64-char fingerprint, got %d", len(fp))
	}

	// Verify it's valid hex
	_, err := hex.DecodeString(fp)
	if err != nil {
		t.Errorf("fingerprint should be valid hex: %v", err)
	}

	// Same input should produce same output
	fp2 := ComputeKeyFingerprint(publicKey)
	if fp != fp2 {
		t.Error("fingerprint should be deterministic")
	}

	// Different input should produce different output
	publicKey2 := make([]byte, 32)
	rand.Read(publicKey2)
	fp3 := ComputeKeyFingerprint(publicKey2)
	if fp == fp3 {
		t.Error("different keys should have different fingerprints")
	}
}
