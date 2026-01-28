package nli

import (
	"context"
	"fmt"
	"strings"
)

// ============================================================================
// Mock LLM Backend
// ============================================================================

// MockLLMBackend is a mock LLM backend for testing and development
type MockLLMBackend struct {
	responses map[string]string
}

// NewMockLLMBackend creates a new mock LLM backend
func NewMockLLMBackend() *MockLLMBackend {
	return &MockLLMBackend{
		responses: make(map[string]string),
	}
}

// Complete generates a mock completion
func (m *MockLLMBackend) Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check for pre-configured responses
	for pattern, response := range m.responses {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(pattern)) {
			return &CompletionResult{
				Text:         response,
				TokensUsed:   len(response) / 4,
				FinishReason: "stop",
			}, nil
		}
	}

	// Generate a default response
	return &CompletionResult{
		Text:         "I understand you're asking about VirtEngine. How can I help you today?",
		TokensUsed:   20,
		FinishReason: "stop",
	}, nil
}

// ClassifyIntent classifies intent using simple pattern matching
func (m *MockLLMBackend) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	lower := strings.ToLower(message)

	// Simple mock classification
	if strings.Contains(lower, "balance") || strings.Contains(lower, "tokens") {
		return &ClassificationResult{
			Intent:     IntentQueryBalance,
			Confidence: 0.9,
		}, nil
	}
	if strings.Contains(lower, "offering") || strings.Contains(lower, "provider") {
		return &ClassificationResult{
			Intent:     IntentFindOfferings,
			Confidence: 0.85,
		}, nil
	}
	if strings.Contains(lower, "help") || strings.Contains(lower, "how to") {
		return &ClassificationResult{
			Intent:     IntentGetHelp,
			Confidence: 0.8,
		}, nil
	}

	return &ClassificationResult{
		Intent:     IntentGeneralChat,
		Confidence: 0.6,
	}, nil
}

// SetResponse sets a custom response for a pattern
func (m *MockLLMBackend) SetResponse(pattern, response string) {
	m.responses[pattern] = response
}

// Close closes the mock backend (no-op)
func (m *MockLLMBackend) Close() error {
	return nil
}

// ============================================================================
// OpenAI LLM Backend (Stub)
// ============================================================================

// OpenAIBackend is a stub for the OpenAI LLM backend
// In production, this would use the OpenAI API
type OpenAIBackend struct {
	config *OpenAIConfig
}

// NewOpenAIBackend creates a new OpenAI backend
func NewOpenAIBackend(config *OpenAIConfig) (*OpenAIBackend, error) {
	if config == nil || config.APIKey == "" {
		return nil, ErrInvalidConfig
	}
	return &OpenAIBackend{config: config}, nil
}

// Complete generates a completion using OpenAI
func (o *OpenAIBackend) Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error) {
	// Stub implementation - would call OpenAI API in production
	return nil, fmt.Errorf("OpenAI backend not implemented: use mock backend for testing")
}

// ClassifyIntent classifies intent using OpenAI
func (o *OpenAIBackend) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	// Stub implementation - would call OpenAI API in production
	return nil, fmt.Errorf("OpenAI backend not implemented: use mock backend for testing")
}

// Close closes the OpenAI backend
func (o *OpenAIBackend) Close() error {
	return nil
}

// ============================================================================
// Anthropic LLM Backend (Stub)
// ============================================================================

// AnthropicBackend is a stub for the Anthropic Claude backend
type AnthropicBackend struct {
	config *AnthropicConfig
}

// NewAnthropicBackend creates a new Anthropic backend
func NewAnthropicBackend(config *AnthropicConfig) (*AnthropicBackend, error) {
	if config == nil || config.APIKey == "" {
		return nil, ErrInvalidConfig
	}
	return &AnthropicBackend{config: config}, nil
}

// Complete generates a completion using Anthropic Claude
func (a *AnthropicBackend) Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error) {
	// Stub implementation - would call Anthropic API in production
	return nil, fmt.Errorf("Anthropic backend not implemented: use mock backend for testing")
}

// ClassifyIntent classifies intent using Anthropic Claude
func (a *AnthropicBackend) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	// Stub implementation - would call Anthropic API in production
	return nil, fmt.Errorf("Anthropic backend not implemented: use mock backend for testing")
}

// Close closes the Anthropic backend
func (a *AnthropicBackend) Close() error {
	return nil
}

// ============================================================================
// Local LLM Backend (Stub)
// ============================================================================

// LocalBackend is a stub for a local LLM backend (e.g., llama.cpp)
type LocalBackend struct {
	config *LocalConfig
}

// NewLocalBackend creates a new local LLM backend
func NewLocalBackend(config *LocalConfig) (*LocalBackend, error) {
	if config == nil || config.ModelPath == "" {
		return nil, ErrInvalidConfig
	}
	return &LocalBackend{config: config}, nil
}

// Complete generates a completion using a local LLM
func (l *LocalBackend) Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error) {
	// Stub implementation - would use local LLM in production
	return nil, fmt.Errorf("Local backend not implemented: use mock backend for testing")
}

// ClassifyIntent classifies intent using a local LLM
func (l *LocalBackend) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	// Stub implementation - would use local LLM in production
	return nil, fmt.Errorf("Local backend not implemented: use mock backend for testing")
}

// Close closes the local LLM backend
func (l *LocalBackend) Close() error {
	return nil
}

// ============================================================================
// Backend Factory
// ============================================================================

// CreateLLMBackend creates an LLM backend based on configuration
func CreateLLMBackend(config *Config) (LLMBackend, error) {
	switch config.LLMBackend {
	case LLMBackendMock:
		return NewMockLLMBackend(), nil
	case LLMBackendOpenAI:
		return NewOpenAIBackend(config.OpenAIConfig)
	case LLMBackendAnthropic:
		return NewAnthropicBackend(config.AnthropicConfig)
	case LLMBackendLocal:
		return NewLocalBackend(config.LocalConfig)
	default:
		return nil, fmt.Errorf("unknown LLM backend: %s", config.LLMBackend)
	}
}
