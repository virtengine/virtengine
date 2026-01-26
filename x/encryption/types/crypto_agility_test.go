package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/encryption/types"
)

// ============================================================================
// Cryptography Agility Tests (VE-227: Cryptography Agility)
// ============================================================================

func TestAlgorithmSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    *types.AlgorithmSpec
		wantErr bool
	}{
		{
			name: "valid classical algorithm",
			spec: types.NewAlgorithmSpec(
				types.AlgorithmIDNaClBox,
				"NaCl Box",
				types.AlgorithmFamilyClassical,
				types.NISTSecurityLevel128,
				true,
			),
			wantErr: false,
		},
		{
			name: "valid post-quantum algorithm",
			spec: types.NewAlgorithmSpec(
				types.AlgorithmIDMLKEM768,
				"ML-KEM-768",
				types.AlgorithmFamilyPostQuantum,
				types.NISTSecurityLevel192,
				false,
			),
			wantErr: false,
		},
		{
			name: "valid hybrid algorithm",
			spec: types.NewAlgorithmSpec(
				types.AlgorithmIDHybridNaClMLKEM,
				"NaCl + ML-KEM Hybrid",
				types.AlgorithmFamilyHybrid,
				types.NISTSecurityLevel192,
				false,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty algorithm ID",
			spec: &types.AlgorithmSpec{
				Version:       types.AlgorithmSpecVersion,
				AlgorithmID:   "",
				Name:          "Test Algorithm",
				Family:        types.AlgorithmFamilyClassical,
				SecurityLevel: types.NISTSecurityLevel128,
				IsDefault:     false,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty name",
			spec: &types.AlgorithmSpec{
				Version:       types.AlgorithmSpecVersion,
				AlgorithmID:   "test-algo",
				Name:          "",
				Family:        types.AlgorithmFamilyClassical,
				SecurityLevel: types.NISTSecurityLevel128,
				IsDefault:     false,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid family",
			spec: &types.AlgorithmSpec{
				Version:       types.AlgorithmSpecVersion,
				AlgorithmID:   "test-algo",
				Name:          "Test Algorithm",
				Family:        "invalid",
				SecurityLevel: types.NISTSecurityLevel128,
				IsDefault:     false,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid security level",
			spec: &types.AlgorithmSpec{
				Version:       types.AlgorithmSpecVersion,
				AlgorithmID:   "test-algo",
				Name:          "Test Algorithm",
				Family:        types.AlgorithmFamilyClassical,
				SecurityLevel: 999,
				IsDefault:     false,
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

func TestSecurityLevels(t *testing.T) {
	// Test all levels are valid
	for _, level := range types.AllSecurityLevels() {
		assert.True(t, types.IsValidSecurityLevel(level), "AllSecurityLevels returned invalid level: %d", level)
	}

	// Test invalid level
	assert.False(t, types.IsValidSecurityLevel(999), "IsValidSecurityLevel should return false for invalid level")
}

func TestAgilityMetadata_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		metadata *types.AgilityMetadata
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: types.NewAgilityMetadata(
				"envelope-123",
				types.AlgorithmIDNaClBox,
				1,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid metadata with fallback",
			metadata: func() *types.AgilityMetadata {
				m := types.NewAgilityMetadata(
					"envelope-123",
					types.AlgorithmIDMLKEM768,
					2,
					now,
				)
				m.FallbackAlgorithm = types.AlgorithmIDNaClBox
				return m
			}(),
			wantErr: false,
		},
		{
			name: "invalid - empty envelope ID",
			metadata: &types.AgilityMetadata{
				Version:          types.AgilityMetadataVersion,
				EnvelopeID:       "",
				AlgorithmID:      types.AlgorithmIDNaClBox,
				AlgorithmVersion: 1,
				CreatedAt:        now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty algorithm ID",
			metadata: &types.AgilityMetadata{
				Version:          types.AgilityMetadataVersion,
				EnvelopeID:       "envelope-123",
				AlgorithmID:      "",
				AlgorithmVersion: 1,
				CreatedAt:        now,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero algorithm version",
			metadata: &types.AgilityMetadata{
				Version:          types.AgilityMetadataVersion,
				EnvelopeID:       "envelope-123",
				AlgorithmID:      types.AlgorithmIDNaClBox,
				AlgorithmVersion: 0,
				CreatedAt:        now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
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
				"envelope-123",
				types.AlgorithmIDNaClBox,
				types.AlgorithmIDMLKEM768,
				types.RotationReasonUpgrade,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid rotation for expiry",
			record: types.NewKeyRotationRecord(
				"rotation-456",
				"envelope-456",
				types.AlgorithmIDNaClBox,
				types.AlgorithmIDNaClBox,
				types.RotationReasonExpiry,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty rotation ID",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "",
				EnvelopeID:       "envelope-123",
				OldAlgorithm:     types.AlgorithmIDNaClBox,
				NewAlgorithm:     types.AlgorithmIDMLKEM768,
				Reason:           types.RotationReasonUpgrade,
				RotatedAt:        now,
				Status:           types.RotationStatusPending,
				ReEncryptionDone: false,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty envelope ID",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "rotation-123",
				EnvelopeID:       "",
				OldAlgorithm:     types.AlgorithmIDNaClBox,
				NewAlgorithm:     types.AlgorithmIDMLKEM768,
				Reason:           types.RotationReasonUpgrade,
				RotatedAt:        now,
				Status:           types.RotationStatusPending,
				ReEncryptionDone: false,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty old algorithm",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "rotation-123",
				EnvelopeID:       "envelope-123",
				OldAlgorithm:     "",
				NewAlgorithm:     types.AlgorithmIDMLKEM768,
				Reason:           types.RotationReasonUpgrade,
				RotatedAt:        now,
				Status:           types.RotationStatusPending,
				ReEncryptionDone: false,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty new algorithm",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "rotation-123",
				EnvelopeID:       "envelope-123",
				OldAlgorithm:     types.AlgorithmIDNaClBox,
				NewAlgorithm:     "",
				Reason:           types.RotationReasonUpgrade,
				RotatedAt:        now,
				Status:           types.RotationStatusPending,
				ReEncryptionDone: false,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid reason",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "rotation-123",
				EnvelopeID:       "envelope-123",
				OldAlgorithm:     types.AlgorithmIDNaClBox,
				NewAlgorithm:     types.AlgorithmIDMLKEM768,
				Reason:           "invalid",
				RotatedAt:        now,
				Status:           types.RotationStatusPending,
				ReEncryptionDone: false,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid status",
			record: &types.KeyRotationRecord{
				Version:          types.KeyRotationVersion,
				RotationID:       "rotation-123",
				EnvelopeID:       "envelope-123",
				OldAlgorithm:     types.AlgorithmIDNaClBox,
				NewAlgorithm:     types.AlgorithmIDMLKEM768,
				Reason:           types.RotationReasonUpgrade,
				RotatedAt:        now,
				Status:           "invalid",
				ReEncryptionDone: false,
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

func TestRotationReasons(t *testing.T) {
	// Test all reasons are valid
	for _, reason := range types.AllRotationReasons() {
		assert.True(t, types.IsValidRotationReason(reason), "AllRotationReasons returned invalid reason: %s", reason)
	}

	// Test invalid reason
	assert.False(t, types.IsValidRotationReason("invalid"), "IsValidRotationReason should return false for invalid reason")
}

func TestRotationStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllRotationStatuses() {
		assert.True(t, types.IsValidRotationStatus(status), "AllRotationStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidRotationStatus("invalid"), "IsValidRotationStatus should return false for invalid status")
}

func TestPostQuantumRoadmap_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		roadmap *types.PostQuantumRoadmap
		wantErr bool
	}{
		{
			name: "valid roadmap",
			roadmap: types.NewPostQuantumRoadmap(
				"roadmap-1",
				"VEID Post-Quantum Migration",
				1,
			),
			wantErr: false,
		},
		{
			name: "valid roadmap with phases",
			roadmap: func() *types.PostQuantumRoadmap {
				r := types.NewPostQuantumRoadmap(
					"roadmap-1",
					"VEID Post-Quantum Migration",
					1,
				)
				r.Phases = []types.RoadmapPhase{
					{
						PhaseID:     "phase-1",
						Name:        "Assessment",
						Description: "Assess current cryptographic usage",
						Status:      types.PhaseStatusActive,
						StartDate:   now,
						EndDate:     now.Add(90 * 24 * time.Hour),
					},
				}
				return r
			}(),
			wantErr: false,
		},
		{
			name: "invalid - empty roadmap ID",
			roadmap: &types.PostQuantumRoadmap{
				Version:        types.RoadmapVersion,
				RoadmapID:      "",
				Name:           "Test Roadmap",
				CurrentVersion: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty name",
			roadmap: &types.PostQuantumRoadmap{
				Version:        types.RoadmapVersion,
				RoadmapID:      "roadmap-1",
				Name:           "",
				CurrentVersion: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero version",
			roadmap: &types.PostQuantumRoadmap{
				Version:        types.RoadmapVersion,
				RoadmapID:      "roadmap-1",
				Name:           "Test Roadmap",
				CurrentVersion: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.roadmap.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoadmapPhase_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		phase   *types.RoadmapPhase
		wantErr bool
	}{
		{
			name: "valid phase",
			phase: &types.RoadmapPhase{
				PhaseID:     "phase-1",
				Name:        "Assessment",
				Description: "Assess cryptographic usage",
				Status:      types.PhaseStatusPlanned,
				StartDate:   now,
				EndDate:     now.Add(90 * 24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "invalid - empty phase ID",
			phase: &types.RoadmapPhase{
				PhaseID:     "",
				Name:        "Assessment",
				Description: "description",
				Status:      types.PhaseStatusPlanned,
				StartDate:   now,
				EndDate:     now.Add(90 * 24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty name",
			phase: &types.RoadmapPhase{
				PhaseID:     "phase-1",
				Name:        "",
				Description: "description",
				Status:      types.PhaseStatusPlanned,
				StartDate:   now,
				EndDate:     now.Add(90 * 24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid status",
			phase: &types.RoadmapPhase{
				PhaseID:     "phase-1",
				Name:        "Assessment",
				Description: "description",
				Status:      "invalid",
				StartDate:   now,
				EndDate:     now.Add(90 * 24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - end before start",
			phase: &types.RoadmapPhase{
				PhaseID:     "phase-1",
				Name:        "Assessment",
				Description: "description",
				Status:      types.PhaseStatusPlanned,
				StartDate:   now,
				EndDate:     now.Add(-24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.phase.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPhaseStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllPhaseStatuses() {
		assert.True(t, types.IsValidPhaseStatus(status), "AllPhaseStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidPhaseStatus("invalid"), "IsValidPhaseStatus should return false for invalid status")
}

func TestDefaultAlgorithmRegistry(t *testing.T) {
	registry := types.DefaultAlgorithmRegistry()

	require.NotEmpty(t, registry, "DefaultAlgorithmRegistry should return at least one algorithm")

	// Validate all default algorithms
	for _, spec := range registry {
		err := spec.Validate()
		require.NoError(t, err, "Default algorithm %s should be valid", spec.AlgorithmID)
	}

	// Check specific algorithms exist
	algoMap := make(map[string]*types.AlgorithmSpec)
	for _, spec := range registry {
		algoMap[spec.AlgorithmID] = spec
	}

	// NaCl Box should exist and be default
	naclBox, exists := algoMap[types.AlgorithmIDNaClBox]
	require.True(t, exists, "NaCl Box algorithm should exist")
	assert.True(t, naclBox.IsDefault, "NaCl Box should be the default algorithm")
	assert.Equal(t, types.AlgorithmFamilyClassical, naclBox.Family, "NaCl Box should be classical family")

	// ML-KEM-768 should exist
	mlkem, exists := algoMap[types.AlgorithmIDMLKEM768]
	require.True(t, exists, "ML-KEM-768 algorithm should exist")
	assert.Equal(t, types.AlgorithmFamilyPostQuantum, mlkem.Family, "ML-KEM-768 should be post-quantum family")
	assert.False(t, mlkem.IsDefault, "ML-KEM-768 should not be default yet")
}

func TestDefaultPostQuantumRoadmap(t *testing.T) {
	roadmap := types.DefaultPostQuantumRoadmap()

	err := roadmap.Validate()
	require.NoError(t, err, "Default roadmap should be valid")

	assert.NotEmpty(t, roadmap.Phases, "Roadmap should have phases")

	// Check phases are valid
	for _, phase := range roadmap.Phases {
		err := phase.Validate()
		require.NoError(t, err, "Phase %s should be valid", phase.PhaseID)
	}

	// Check expected phases exist
	phaseMap := make(map[string]*types.RoadmapPhase)
	for i := range roadmap.Phases {
		phaseMap[roadmap.Phases[i].PhaseID] = &roadmap.Phases[i]
	}

	assert.Contains(t, phaseMap, "assessment", "Roadmap should have assessment phase")
	assert.Contains(t, phaseMap, "hybrid-testing", "Roadmap should have hybrid-testing phase")
	assert.Contains(t, phaseMap, "migration", "Roadmap should have migration phase")
}

func TestKeyRotationRecord_Complete(t *testing.T) {
	now := time.Now()

	record := types.NewKeyRotationRecord(
		"rotation-123",
		"envelope-123",
		types.AlgorithmIDNaClBox,
		types.AlgorithmIDMLKEM768,
		types.RotationReasonUpgrade,
		now,
	)

	assert.Equal(t, types.RotationStatusPending, record.Status, "New record should have pending status")
	assert.False(t, record.ReEncryptionDone, "ReEncryptionDone should be false initially")

	// Complete the rotation
	completeTime := now.Add(1 * time.Hour)
	record.Complete(completeTime)

	assert.Equal(t, types.RotationStatusCompleted, record.Status, "Status should be completed")
	assert.True(t, record.ReEncryptionDone, "ReEncryptionDone should be true")
	require.NotNil(t, record.CompletedAt, "CompletedAt should be set")
	assert.True(t, record.CompletedAt.Equal(completeTime), "CompletedAt should be set to complete time")
}

func TestKeyRotationRecord_Fail(t *testing.T) {
	now := time.Now()

	record := types.NewKeyRotationRecord(
		"rotation-123",
		"envelope-123",
		types.AlgorithmIDNaClBox,
		types.AlgorithmIDMLKEM768,
		types.RotationReasonUpgrade,
		now,
	)

	// Fail the rotation
	failTime := now.Add(1 * time.Hour)
	record.Fail("Algorithm not supported", failTime)

	assert.Equal(t, types.RotationStatusFailed, record.Status, "Status should be failed")
	assert.Equal(t, "Algorithm not supported", record.ErrorMessage, "Error message should be set")
	assert.False(t, record.ReEncryptionDone, "ReEncryptionDone should still be false")
}

func TestAgilityMetadata_NeedsRotation(t *testing.T) {
	now := time.Now()

	// Not deprecated, no rotation needed
	metadata := types.NewAgilityMetadata(
		"envelope-123",
		types.AlgorithmIDNaClBox,
		1,
		now,
	)

	assert.False(t, metadata.NeedsRotation(), "Non-deprecated algorithm should not need rotation")

	// Deprecated algorithm needs rotation
	metadata.IsDeprecated = true

	assert.True(t, metadata.NeedsRotation(), "Deprecated algorithm should need rotation")
}

func TestAlgorithmSpec_IsQuantumSafe(t *testing.T) {
	// Classical algorithm is not quantum-safe
	classical := types.NewAlgorithmSpec(
		types.AlgorithmIDNaClBox,
		"NaCl Box",
		types.AlgorithmFamilyClassical,
		types.NISTSecurityLevel128,
		true,
	)

	assert.False(t, classical.IsQuantumSafe(), "Classical algorithm should not be quantum-safe")

	// Hybrid algorithm is quantum-safe
	hybrid := types.NewAlgorithmSpec(
		types.AlgorithmIDHybridNaClMLKEM,
		"NaCl + ML-KEM Hybrid",
		types.AlgorithmFamilyHybrid,
		types.NISTSecurityLevel192,
		false,
	)

	assert.True(t, hybrid.IsQuantumSafe(), "Hybrid algorithm should be quantum-safe")

	// Post-quantum algorithm is quantum-safe
	pq := types.NewAlgorithmSpec(
		types.AlgorithmIDMLKEM768,
		"ML-KEM-768",
		types.AlgorithmFamilyPostQuantum,
		types.NISTSecurityLevel192,
		false,
	)

	assert.True(t, pq.IsQuantumSafe(), "Post-quantum algorithm should be quantum-safe")
}
