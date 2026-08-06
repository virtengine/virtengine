//go:build security

// Package blockchain contains security tests for consensus-critical verification logic.
package blockchain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"testing"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type consensusTestScorer struct {
	version string
	healthy bool
	err     error
}

func (s consensusTestScorer) Score(_ *veidkeeper.ScoringInput) (*veidkeeper.ScoringOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &veidkeeper.ScoringOutput{
		Score:        85,
		ModelVersion: s.version,
		ReasonCodes:  []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess},
		InputHash:    []byte("computed-input-hash"),
	}, nil
}

func (s consensusTestScorer) GetModelVersion() string { return s.version }
func (s consensusTestScorer) IsHealthy() bool         { return s.healthy }
func (s consensusTestScorer) Close() error            { return nil }

type consensusTestKeyProvider struct{}

func (consensusTestKeyProvider) GetPrivateKey() ([]byte, error) { return []byte("validator-private-key"), nil }
func (consensusTestKeyProvider) GetKeyFingerprint() string      { return "validator-fingerprint" }
func (consensusTestKeyProvider) Close() error                   { return nil }

func TestBC001_ConsensusVerifierRejectsDivergentResults(t *testing.T) {
	params := veidkeeper.DefaultConsensusParams()
	verifier := veidkeeper.NewConsensusVerifier(
		nil,
		consensusTestScorer{version: "model-v1", healthy: true},
		consensusTestKeyProvider{},
		params,
		log.NewNopLogger(),
	)

	baseHash := sha256.Sum256([]byte("same-inputs"))
	proposed := veidtypes.VerificationResult{
		RequestID:    "req-100",
		Score:        85,
		Status:       veidtypes.VerificationResultStatusSuccess,
		ModelVersion: "model-v1",
		InputHash:    baseHash[:],
	}

	t.Run("exact_match_is_accepted", func(t *testing.T) {
		computed := proposed
		comparison := verifier.CompareResults(proposed, computed)

		require.True(t, comparison.Match)
		require.Empty(t, comparison.Differences)
		require.Equal(t, int32(0), comparison.ScoreDifference)
		require.True(t, comparison.ModelVersionMatch)
		require.True(t, comparison.InputHashMatch)
		require.True(t, comparison.StatusMatch)
	})

	t.Run("score_difference_breaks_consensus", func(t *testing.T) {
		computed := proposed
		computed.Score = 86

		comparison := verifier.CompareResults(proposed, computed)

		require.False(t, comparison.Match)
		require.Equal(t, int32(1), comparison.ScoreDifference)
		require.Contains(t, comparison.Differences[0], "score difference 1 exceeds tolerance 0")
	})

	t.Run("model_version_mismatch_breaks_consensus", func(t *testing.T) {
		computed := proposed
		computed.ModelVersion = "model-v2"

		comparison := verifier.CompareResults(proposed, computed)

		require.False(t, comparison.Match)
		require.False(t, comparison.ModelVersionMatch)
		require.Contains(t, comparison.Differences, "model version mismatch: proposed=model-v1, computed=model-v2")
	})

	t.Run("input_hash_mismatch_breaks_consensus", func(t *testing.T) {
		computed := proposed
		otherHash := sha256.Sum256([]byte("different-inputs"))
		computed.InputHash = otherHash[:]

		comparison := verifier.CompareResults(proposed, computed)

		require.False(t, comparison.Match)
		require.False(t, comparison.InputHashMatch)
		require.Contains(t, comparison.Differences[0], "input hash mismatch")
	})

	t.Run("status_mismatch_breaks_consensus", func(t *testing.T) {
		computed := proposed
		computed.Status = veidtypes.VerificationResultStatusFailed

		comparison := verifier.CompareResults(proposed, computed)

		require.False(t, comparison.Match)
		require.False(t, comparison.StatusMatch)
		require.Contains(t, comparison.Differences[0], "status mismatch")
	})
}

func TestBC002_ConsensusVerifierValidatesLocalModelState(t *testing.T) {
	verifier := veidkeeper.NewConsensusVerifier(
		nil,
		consensusTestScorer{version: "model-v1", healthy: true},
		consensusTestKeyProvider{},
		veidkeeper.DefaultConsensusParams(),
		log.NewNopLogger(),
	)

	t.Run("matching_version_and_healthy_scorer_pass", func(t *testing.T) {
		require.NoError(t, verifier.ValidateModelVersion(sdk.Context{}, "model-v1"))
	})

	t.Run("mismatched_model_version_is_rejected", func(t *testing.T) {
		err := verifier.ValidateModelVersion(sdk.Context{}, "model-v2")
		require.Error(t, err)
		require.Contains(t, err.Error(), "model version mismatch")
	})

	t.Run("unhealthy_scorer_is_rejected", func(t *testing.T) {
		unhealthy := veidkeeper.NewConsensusVerifier(
			nil,
			consensusTestScorer{version: "model-v1", healthy: false},
			consensusTestKeyProvider{},
			veidkeeper.DefaultConsensusParams(),
			log.NewNopLogger(),
		)

		err := unhealthy.ValidateModelVersion(sdk.Context{}, "model-v1")
		require.Error(t, err)
		require.Contains(t, err.Error(), "ML scorer is not healthy")
	})

	t.Run("scorer_errors_do_not_bypass_model_checks", func(t *testing.T) {
		broken := veidkeeper.NewConsensusVerifier(
			nil,
			consensusTestScorer{version: "model-v1", healthy: true, err: errors.New("backend down")},
			consensusTestKeyProvider{},
			veidkeeper.DefaultConsensusParams(),
			log.NewNopLogger(),
		)

		require.NoError(t, broken.ValidateModelVersion(sdk.Context{}, "model-v1"))
	})
}

func TestBC002_ResultHashIsDeterministicAndFieldSensitive(t *testing.T) {
	baseHash := sha256.Sum256([]byte("stable-input-hash"))
	result := veidtypes.VerificationResult{
		RequestID:      "req-200",
		AccountAddress: "virtengine1auditsecurity0000000000000000000000",
		Score:          91,
		Status:         veidtypes.VerificationResultStatusSuccess,
		ModelVersion:   "model-v1",
		InputHash:      baseHash[:],
		BlockHeight:    2048,
	}

	first := veidkeeper.ComputeResultHash(result)
	second := veidkeeper.ComputeResultHash(result)

	require.True(t, bytes.Equal(first, second), "same result must hash identically")

	mutatedScore := result
	mutatedScore.Score++
	require.False(t, bytes.Equal(first, veidkeeper.ComputeResultHash(mutatedScore)))

	mutatedHash := result
	otherHash := sha256.Sum256([]byte("mutated-input-hash"))
	mutatedHash.InputHash = otherHash[:]
	require.False(t, bytes.Equal(first, veidkeeper.ComputeResultHash(mutatedHash)))

	mutatedStatus := result
	mutatedStatus.Status = veidtypes.VerificationResultStatusFailed
	require.False(t, bytes.Equal(first, veidkeeper.ComputeResultHash(mutatedStatus)))
}
