package types

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInferenceProfileSnapshotValidationAndDeepCopy(t *testing.T) {
	snapshot := testInferenceProfileSnapshot()
	require.NoError(t, snapshot.Validate())

	request := NewVerificationRequest("request-1", "addr1", []string{"scope-a"}, time.Unix(1, 0).UTC(), 10)
	require.NoError(t, request.SetInferenceProfileSnapshot(snapshot))
	snapshot.RuntimeDigest[0] ^= 0xff
	require.NotEqual(t, snapshot.RuntimeDigest, request.InferenceProfileSnapshot.RuntimeDigest)

	copySnapshot, err := request.RequireInferenceProfileSnapshot()
	require.NoError(t, err)
	copySnapshot.ModelDigest[0] ^= 0xff
	require.NotEqual(t, copySnapshot.ModelDigest, request.InferenceProfileSnapshot.ModelDigest)
}

func TestInferenceProfileSnapshotRejectsMalformedMissingAndAliasedDigests(t *testing.T) {
	base := testInferenceProfileSnapshot()
	tests := []struct {
		name   string
		mutate func(*InferenceProfileSnapshot)
	}{
		{"missing pipeline", func(s *InferenceProfileSnapshot) { s.PipelineVersion = "" }},
		{"invalid activation", func(s *InferenceProfileSnapshot) { s.ActivationHeight = 0 }},
		{"missing digest", func(s *InferenceProfileSnapshot) { s.ModelDigest = nil }},
		{"short digest", func(s *InferenceProfileSnapshot) { s.FeatureSchemaDigest = []byte{1, 2, 3} }},
		{"aliased digest", func(s *InferenceProfileSnapshot) { s.ModelDigest = s.ModelManifestDigest }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := base.DeepCopy()
			tc.mutate(snapshot)
			require.Error(t, snapshot.Validate())
		})
	}
}

func testInferenceProfileSnapshot() *InferenceProfileSnapshot {
	return &InferenceProfileSnapshot{
		PipelineVersion:         "v1.0.0",
		RuntimeImageDigest:      testSnapshotDigest(0x01),
		RuntimeDigest:           testSnapshotDigest(0x01),
		ModelManifestDigest:     testSnapshotDigest(0x02),
		ModelDigest:             testSnapshotDigest(0x03),
		DeterminismConfigDigest: testSnapshotDigest(0x04),
		FeatureSchemaDigest:     testSnapshotDigest(0x05),
		ActivationHeight:        10,
	}
}

func testSnapshotDigest(value byte) []byte {
	out := make([]byte, sha256.Size)
	for i := range out {
		out[i] = value
	}
	return out
}
