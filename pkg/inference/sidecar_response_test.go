package inference

import (
	"math"
	"strings"
	"testing"

	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

func TestSidecarClientValidatesScoreResponseBinding(t *testing.T) {
	config := DefaultInferenceConfig()
	config.UseSidecar = true
	config.SidecarAddress = testSidecarAddress
	config.ExpectedHash = strings.Repeat("a", 64)
	config.RequireHashVerification = true
	config.UseFallbackOnError = false
	config.StrictDeterminism = true

	client := &SidecarClient{
		config:       config,
		determinism:  NewDeterminismController(config.RandomSeed, config.ForceCPU),
		modelVersion: config.ModelVersion,
		modelHash:    config.ExpectedHash,
	}
	features := make([]float32, TotalFeatureDim)
	features[0] = 0.25
	featureHash := client.determinism.ComputeFeatureHash(features)
	rawScore := float32(42.5)
	outputHash := client.determinism.ComputeOutputHash([]float32{rawScore})

	valid := &inferencepb.ComputeScoreResponse{
		Score:        42,
		RawScore:     rawScore,
		Confidence:   0.9,
		InputHash:    featureHash,
		OutputHash:   outputHash,
		ModelVersion: config.ModelVersion,
		ModelHash:    config.ExpectedHash,
	}
	if err := client.validateScoreResponse(valid, featureHash); err != nil {
		t.Fatalf("expected bound response to validate: %v", err)
	}

	for name, mutate := range map[string]func(*inferencepb.ComputeScoreResponse){
		"missing input hash":    func(response *inferencepb.ComputeScoreResponse) { response.InputHash = "" },
		"mismatched input hash": func(response *inferencepb.ComputeScoreResponse) { response.InputHash = strings.Repeat("b", 64) },
		"missing output hash":   func(response *inferencepb.ComputeScoreResponse) { response.OutputHash = "" },
		"mismatched output hash": func(response *inferencepb.ComputeScoreResponse) {
			response.OutputHash = strings.Repeat("c", 64)
		},
		"missing model identity": func(response *inferencepb.ComputeScoreResponse) { response.ModelHash = "" },
		"mismatched model version": func(response *inferencepb.ComputeScoreResponse) {
			response.ModelVersion = "unexpected"
		},
		"mismatched model hash": func(response *inferencepb.ComputeScoreResponse) {
			response.ModelHash = strings.Repeat("d", 64)
		},
		"score out of range":      func(response *inferencepb.ComputeScoreResponse) { response.Score = 101 },
		"raw score is not finite": func(response *inferencepb.ComputeScoreResponse) { response.RawScore = float32(math.NaN()) },
		"confidence out of range": func(response *inferencepb.ComputeScoreResponse) { response.Confidence = 1.1 },
		"invalid input hash syntax": func(response *inferencepb.ComputeScoreResponse) {
			response.InputHash = "invalid"
		},
		"invalid output hash syntax": func(response *inferencepb.ComputeScoreResponse) {
			response.OutputHash = "invalid"
		},
		"invalid model hash syntax": func(response *inferencepb.ComputeScoreResponse) {
			response.ModelHash = "invalid"
		},
		"negative compute time": func(response *inferencepb.ComputeScoreResponse) { response.ComputeTimeMs = -1 },
		"invalid contribution": func(response *inferencepb.ComputeScoreResponse) {
			response.FeatureContributions = map[string]float32{"face": float32(math.Inf(1))}
		},
		"blank reason code": func(response *inferencepb.ComputeScoreResponse) { response.ReasonCodes = []string{" "} },
		"too many reason codes": func(response *inferencepb.ComputeScoreResponse) {
			response.ReasonCodes = make([]string, 33)
			for i := range response.ReasonCodes {
				response.ReasonCodes[i] = "SUCCESS"
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			response := *valid
			mutate(&response)
			if err := client.validateScoreResponse(&response, featureHash); err == nil {
				t.Fatal("expected score response binding validation error")
			}
		})
	}
}
