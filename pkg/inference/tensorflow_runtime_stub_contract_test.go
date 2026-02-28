//go:build !mlruntime

package inference

import (
	"errors"
	"testing"
)

func TestTensorFlowRuntimeStubRequiresExplicitOptIn(t *testing.T) {
	modelDir := createTestModelDir(t)

	config := DefaultTFRuntimeConfig()
	config.ModelPath = modelDir

	runtime := NewTensorFlowRuntime(config)
	err := runtime.Initialize()
	if !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error, got %v", err)
	}
}

func TestTensorFlowRuntimeStubAllowsExplicitOptIn(t *testing.T) {
	modelDir := createTestModelDir(t)

	config := DefaultTFRuntimeConfig()
	config.ModelPath = modelDir
	config.AllowFallbackToStub = true

	runtime := NewTensorFlowRuntime(config)
	if err := runtime.Initialize(); err != nil {
		t.Fatalf("initialize stub runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	output, err := runtime.Run(make([]float32, TotalFeatureDim))
	if err != nil {
		t.Fatalf("run stub runtime: %v", err)
	}
	if len(output) != 1 {
		t.Fatalf("expected single score output, got %#v", output)
	}
}
