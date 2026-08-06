package inference

import (
	"strings"
	"testing"
	"time"
)

func TestProductionPipelineValidatesInputEnvelope(t *testing.T) {
	pipeline := NewFeaturePipeline(ProductionPipelineConfig())
	valid := completePipelineInput()
	if err := pipeline.validatePipelineInput(valid); err != nil {
		t.Fatalf("expected complete input to validate: %v", err)
	}

	for name, mutate := range map[string]func(*PipelineInput){
		"missing account":        func(input *PipelineInput) { input.AccountAddress = "" },
		"duplicate scope ID":     func(input *PipelineInput) { input.Scopes[1].ScopeID = input.Scopes[0].ScopeID },
		"unsupported scope type": func(input *PipelineInput) { input.Scopes[0].ScopeType = "unknown" },
		"empty scope data":       func(input *PipelineInput) { input.Scopes[0].Data = nil },
		"invalid metadata":       func(input *PipelineInput) { input.Scopes[1].Metadata = map[string]string{" ": "passport"} },
		"too many scopes": func(input *PipelineInput) {
			for i := len(input.Scopes); i <= maxPipelineScopes; i++ {
				input.Scopes = append(input.Scopes, ScopeData{ScopeID: "extra-" + string(rune('a'+i)), ScopeType: "selfie", Data: []byte("data")})
			}
		},
		"missing liveness scope": func(input *PipelineInput) { input.Scopes = input.Scopes[:2] },
		"negative block height":  func(input *PipelineInput) { input.BlockHeight = -1 },
		"missing request ID":     func(input *PipelineInput) { input.RequestID = "" },
		"missing block time":     func(input *PipelineInput) { input.BlockTime = time.Time{} },
	} {
		t.Run(name, func(t *testing.T) {
			input := clonePipelineInput(valid)
			mutate(input)
			if err := pipeline.validatePipelineInput(input); err == nil {
				t.Fatal("expected pipeline input validation error")
			}
		})
	}
}

func TestFeaturePipelineInputHashBindsAllMetadata(t *testing.T) {
	pipeline := NewFeaturePipeline(DefaultPipelineConfig())
	input := completePipelineInput()
	baseline := pipeline.computeInputHash(input)

	changes := []func(*PipelineInput){
		func(value *PipelineInput) { value.BlockTime = value.BlockTime.Add(time.Nanosecond) },
		func(value *PipelineInput) { value.ValidatorAddress = "virtvaloper1other" },
		func(value *PipelineInput) { value.Scopes[1].Metadata["document_type"] = "id_card" },
		func(value *PipelineInput) { value.Scopes[1].Metadata["issuer"] = "AUS" },
	}
	for _, change := range changes {
		candidate := clonePipelineInput(input)
		change(candidate)
		if got := pipeline.computeInputHash(candidate); got == baseline {
			t.Fatal("pipeline input hash must change when bound metadata changes")
		}
	}
}

func TestFeaturePipelineInputSnapshotOwnsMutableScopeData(t *testing.T) {
	input := completePipelineInput()
	cloned := clonePipelineInput(input)
	input.Scopes[0].Data[0] = 'X'
	input.Scopes[1].Metadata["document_type"] = "tampered"

	if strings.HasPrefix(string(cloned.Scopes[0].Data), "X") || cloned.Scopes[1].Metadata["document_type"] != "passport" {
		t.Fatal("pipeline input snapshot must not share caller-owned mutable data")
	}
}

func completePipelineInput() *PipelineInput {
	return &PipelineInput{
		Scopes: []ScopeData{
			{ScopeID: "selfie-1", ScopeType: "selfie", Data: []byte("selfie")},
			{ScopeID: "document-1", ScopeType: "id_document", Data: []byte("document"), Metadata: map[string]string{"document_type": "passport"}},
			{ScopeID: "video-1", ScopeType: "face_video", Data: []byte("video")},
		},
		AccountAddress:   "virt1account",
		BlockHeight:      42,
		BlockTime:        time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC),
		RequestID:        "request-1",
		ValidatorAddress: "virtvaloper1validator",
	}
}
