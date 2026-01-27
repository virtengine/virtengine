package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// ============================================================================
// Cryptography Agility Tests (VE-227: Cryptography Agility)
// ============================================================================

func TestAlgorithmSpec_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		spec    *types.AlgorithmSpec
		wantErr bool
	}{
		{
			name: "valid classical algorithm",
			spec: types.NewAlgorithmSpec(
				"X25519-XSALSA20-POLY1305",
				1,
				types.AlgorithmFamilyClassical,
				"NaCl Box",
				32,
				24,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid post-quantum algorithm",
			spec: types.NewAlgorithmSpec(
				"ML-KEM-768",
				1,
				types.AlgorithmFamilyPostQuantum,
				"ML-KEM-768",
				32,
				24,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid hybrid algorithm",
			spec: types.NewAlgorithmSpec(
				"HYBRID-NACL-MLKEM",
				1,
				types.AlgorithmFamilyHybrid,
				"NaCl + ML-KEM Hybrid",
				32,
				24,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty algorithm ID",
			spec: &types.AlgorithmSpec{
				Version:        1,
				ID:             "",
				Description:    "Test Algorithm",
				Family:         types.AlgorithmFamilyClassical,
				KeySizeBytes:   32,
				NonceSizeBytes: 24,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero version",
			spec: &types.AlgorithmSpec{
				Version:        0,
				ID:             "test-algo",
				Description:    "Test Algorithm",
				Family:         types.AlgorithmFamilyClassical,
				KeySizeBytes:   32,
				NonceSizeBytes: 24,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid family",
			spec: &types.AlgorithmSpec{
				Version:        1,
				ID:             "test-algo",
				Description:    "Test Algorithm",
				Family:         "invalid",
				KeySizeBytes:   32,
				NonceSizeBytes: 24,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero key size",
			spec: &types.AlgorithmSpec{
				Version:        1,
				ID:             "test-algo",
				Description:    "Test Algorithm",
				Family:         types.AlgorithmFamilyClassical,
				KeySizeBytes:   0,
				NonceSizeBytes: 24,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero nonce size",
			spec: &types.AlgorithmSpec{
				Version:        1,
				ID:             "test-algo",
				Description:    "Test Algorithm",
				Family:         types.AlgorithmFamilyClassical,
				KeySizeBytes:   32,
				NonceSizeBytes: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAlgorithmFamilies(t *testing.T) {
	// Test all families are valid
	for _, family := range types.AllAlgorithmFamilies() {
		assert.True(t, types.IsValidAlgorithmFamily(family), "AllAlgorithmFamilies returned invalid family: %s", family)
	}

	// Test invalid family
	assert.False(t, types.IsValidAlgorithmFamily("invalid"), "IsValidAlgorithmFamily should return false for invalid family")
}

func TestAlgorithmStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllAlgorithmStatuses() {
		assert.True(t, types.IsValidAlgorithmStatus(status), "AllAlgorithmStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidAlgorithmStatus("invalid"), "IsValidAlgorithmStatus should return false for invalid status")
}

func TestAlgorithmSpec_IsUsable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		status   types.AlgorithmStatus
		expected bool
	}{
		{"approved is usable", types.AlgorithmStatusApproved, true},
		{"recommended is usable", types.AlgorithmStatusRecommended, true},
		{"experimental is not usable", types.AlgorithmStatusExperimental, false},
		{"deprecated is not usable", types.AlgorithmStatusDeprecated, false},
		{"disabled is not usable", types.AlgorithmStatusDisabled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := types.NewAlgorithmSpec("test", 1, types.AlgorithmFamilyClassical, "test", 32, 24, now)
			spec.Status = tt.status
			assert.Equal(t, tt.expected, spec.IsUsable())
		})
	}
}

func TestAlgorithmSpec_IsDecryptable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		status   types.AlgorithmStatus
		expected bool
	}{
		{"approved is decryptable", types.AlgorithmStatusApproved, true},
		{"recommended is decryptable", types.AlgorithmStatusRecommended, true},
		{"experimental is decryptable", types.AlgorithmStatusExperimental, true},
		{"deprecated is decryptable", types.AlgorithmStatusDeprecated, true},
		{"disabled is not decryptable", types.AlgorithmStatusDisabled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := types.NewAlgorithmSpec("test", 1, types.AlgorithmFamilyClassical, "test", 32, 24, now)
			spec.Status = tt.status
			assert.Equal(t, tt.expected, spec.IsDecryptable())
		})
	}
}

func TestAgilityMetadata(t *testing.T) {
	now := time.Now()

	// Test creating new metadata
	metadata := types.NewAgilityMetadata(
		types.AlgorithmX25519XSalsa20Poly1305,
		1,
		types.AlgorithmFamilyClassical,
		now,
	)

	require.NotNil(t, metadata)
	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, metadata.AlgorithmID)
	assert.Equal(t, uint32(1), metadata.AlgorithmVersion)
	assert.Equal(t, types.AlgorithmFamilyClassical, metadata.AlgorithmFamily)
	assert.True(t, metadata.MigrationEligible)
}

func TestAgilityMetadata_SetHybridAlgorithm(t *testing.T) {
	now := time.Now()

	metadata := types.NewAgilityMetadata(
		types.AlgorithmX25519XSalsa20Poly1305,
		1,
		types.AlgorithmFamilyClassical,
		now,
	)

	// Set hybrid algorithm
	metadata.SetHybridAlgorithm("ML-KEM-768", 1)

	assert.Equal(t, "ML-KEM-768", metadata.HybridAlgorithmID)
	assert.Equal(t, uint32(1), metadata.HybridAlgorithmVersion)
	assert.Equal(t, types.AlgorithmFamilyHybrid, metadata.AlgorithmFamily)
}

func TestKeyRotationReasons(t *testing.T) {
	// Test all reasons are valid
	reasons := types.AllKeyRotationReasons()
	require.NotEmpty(t, reasons)

	// Should include common rotation reasons
	assert.Contains(t, reasons, types.KeyRotationScheduled)
	assert.Contains(t, reasons, types.KeyRotationAlgorithmMigration)
	assert.Contains(t, reasons, types.KeyRotationCompromise)
}

func TestKeyRotationRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *types.KeyRotationRecord
		wantErr bool
	}{
		{
			name: "valid rotation record",
			record: types.NewKeyRotationRecord(
				"rotation-123",
				"virtengine1address123",
				types.KeyRotationScheduled,
				types.AlgorithmX25519XSalsa20Poly1305,
				1,
				"ML-KEM-768",
				1,
				"old-fingerprint",
				"new-fingerprint",
				now,
				30,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty rotation ID",
			record: &types.KeyRotationRecord{
				RotationID:     "",
				AccountAddress: "virtengine1address123",
				OldAlgorithmID: types.AlgorithmX25519XSalsa20Poly1305,
				NewAlgorithmID: "ML-KEM-768",
				InitiatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			record: &types.KeyRotationRecord{
				RotationID:     "rotation-123",
				AccountAddress: "",
				OldAlgorithmID: types.AlgorithmX25519XSalsa20Poly1305,
				NewAlgorithmID: "ML-KEM-768",
				InitiatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty old algorithm",
			record: &types.KeyRotationRecord{
				RotationID:     "rotation-123",
				AccountAddress: "virtengine1address123",
				OldAlgorithmID: "",
				NewAlgorithmID: "ML-KEM-768",
				InitiatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty new algorithm",
			record: &types.KeyRotationRecord{
				RotationID:     "rotation-123",
				AccountAddress: "virtengine1address123",
				OldAlgorithmID: types.AlgorithmX25519XSalsa20Poly1305,
				NewAlgorithmID: "",
				InitiatedAt:    now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyRotationRecord_IsInTransition(t *testing.T) {
	now := time.Now()

	record := types.NewKeyRotationRecord(
		"rotation-123",
		"virtengine1address123",
		types.KeyRotationScheduled,
		types.AlgorithmX25519XSalsa20Poly1305,
		1,
		"ML-KEM-768",
		1,
		"old-fp",
		"new-fp",
		now,
		30,
	)

	// Set to in-transition status
	record.Status = types.KeyRotationStatusInTransition

	// Should be in transition now
	assert.True(t, record.IsInTransition(now))

	// Should not be in transition after window ends
	assert.False(t, record.IsInTransition(now.AddDate(0, 2, 0)))
}

func TestKeyRotationRecord_MarkCompleted(t *testing.T) {
	now := time.Now()

	record := types.NewKeyRotationRecord(
		"rotation-123",
		"virtengine1address123",
		types.KeyRotationScheduled,
		types.AlgorithmX25519XSalsa20Poly1305,
		1,
		"ML-KEM-768",
		1,
		"old-fp",
		"new-fp",
		now,
		30,
	)

	completedAt := now.Add(time.Hour)
	record.MarkCompleted(completedAt)

	assert.Equal(t, types.KeyRotationStatusCompleted, record.Status)
	require.NotNil(t, record.CompletedAt)
	assert.Equal(t, completedAt, *record.CompletedAt)
}

func TestPostQuantumReadinessLevels(t *testing.T) {
	// Test all readiness levels exist
	levels := []types.PostQuantumReadinessLevel{
		types.PQReadinessNone,
		types.PQReadinessPlanned,
		types.PQReadinessHybrid,
		types.PQReadinessFull,
	}

	for _, level := range levels {
		assert.NotEmpty(t, string(level))
	}
}

func TestDefaultPostQuantumRoadmap(t *testing.T) {
	now := time.Now()

	roadmap := types.DefaultPostQuantumRoadmap(now)

	require.NotNil(t, roadmap)
	assert.Equal(t, types.CryptoAgilityVersion, roadmap.Version)
	assert.Equal(t, types.PQReadinessPlanned, roadmap.CurrentLevel)
	assert.Equal(t, types.PQReadinessFull, roadmap.TargetLevel)
	assert.NotEmpty(t, roadmap.PlannedMilestones)
	assert.NotEmpty(t, roadmap.RecommendedAlgorithms)
	assert.Contains(t, roadmap.RecommendedAlgorithms, "ML-KEM-768")
}

func TestDefaultAlgorithmRegistry(t *testing.T) {
	now := time.Now()

	registry := types.DefaultAlgorithmRegistry(now)

	require.NotEmpty(t, registry)

	// Should contain the primary algorithm
	var foundPrimary bool
	for _, spec := range registry {
		if spec.ID == types.AlgorithmX25519XSalsa20Poly1305 {
			foundPrimary = true
			assert.Equal(t, types.AlgorithmFamilyClassical, spec.Family)
			assert.Equal(t, types.AlgorithmStatusRecommended, spec.Status)
		}
	}
	assert.True(t, foundPrimary, "Registry should contain primary algorithm")
}

func TestAlgorithmConstants(t *testing.T) {
	// Test algorithm ID constants are properly defined
	assert.Equal(t, "X25519-XSALSA20-POLY1305", types.AlgorithmX25519XSalsa20Poly1305)
	assert.Equal(t, "AGE-X25519", types.AlgorithmAgeX25519)

	// Test compatibility aliases
	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, types.AlgorithmIDNaClBox)
	assert.Equal(t, "ML-KEM-768", types.AlgorithmIDMLKEM768)
	assert.Equal(t, "HYBRID-NACL-MLKEM", types.AlgorithmIDHybridNaClMLKEM)

	// Test NIST security levels
	assert.Equal(t, 128, types.NISTSecurityLevel128)
	assert.Equal(t, 192, types.NISTSecurityLevel192)
	assert.Equal(t, 256, types.NISTSecurityLevel256)
}
