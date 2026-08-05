package inference

import (
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
		RawScore:     rawScore,
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
