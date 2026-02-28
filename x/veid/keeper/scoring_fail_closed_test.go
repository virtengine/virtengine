package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRealInferenceReadyRejectsPlaceholderHash(t *testing.T) {
	err := isRealInferenceReady(&TensorFlowScoringConfig{
		ModelPath:    "models/trust_score",
		ExpectedHash: "placeholder-hash",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "placeholder")
}

func TestGetMLScorerFailsClosedWhenTensorFlowBundleMissing(t *testing.T) {
	t.Setenv("VEID_USE_TENSORFLOW", "true")
	t.Setenv("VEID_DISABLE_TENSORFLOW", "")
	t.Setenv("VEID_INFERENCE_ENABLED", "")
	t.Setenv("VEID_INFERENCE_MODEL_PATH", "models/does-not-exist")
	t.Setenv("VEID_INFERENCE_MODEL_HASH", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	k, _ := setupPipelineTestKeeper(t)
	scorer := k.getMLScorer()
	t.Cleanup(func() {
		_ = scorer.Close()
	})

	require.False(t, scorer.IsHealthy())

	_, err := scorer.Score(&ScoringInput{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not ready")
}
