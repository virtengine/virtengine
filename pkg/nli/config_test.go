package nli

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty LLM backend",
			config: Config{
				LLMBackend: "",
			},
			wantErr: true,
		},
		{
			name: "mock backend",
			config: Config{
				LLMBackend: LLMBackendMock,
			},
			wantErr: false,
		},
		{
			name: "openai with config",
			config: Config{
				LLMBackend: LLMBackendOpenAI,
				OpenAIConfig: &OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-4",
				},
			},
			wantErr: false,
		},
		{
			name: "openai without config",
			config: Config{
				LLMBackend: LLMBackendOpenAI,
			},
			wantErr: true,
		},
		{
			name: "openai without api key",
			config: Config{
				LLMBackend: LLMBackendOpenAI,
				OpenAIConfig: &OpenAIConfig{
					Model: "gpt-4",
				},
			},
			wantErr: true,
		},
		{
			name: "anthropic with config",
			config: Config{
				LLMBackend: LLMBackendAnthropic,
				AnthropicConfig: &AnthropicConfig{
					APIKey: "test-key",
					Model:  "claude-3",
				},
			},
			wantErr: false,
		},
		{
			name: "anthropic without config",
			config: Config{
				LLMBackend: LLMBackendAnthropic,
			},
			wantErr: true,
		},
		{
			name: "local with config",
			config: Config{
				LLMBackend: LLMBackendLocal,
				LocalConfig: &LocalConfig{
					ModelPath:   "/path/to/model",
					ContextSize: 4096,
				},
			},
			wantErr: false,
		},
		{
			name: "local without config",
			config: Config{
				LLMBackend: LLMBackendLocal,
			},
			wantErr: true,
		},
		{
			name: "unknown backend",
			config: Config{
				LLMBackend: "unknown-backend",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_DefaultsApplied(t *testing.T) {
	config := Config{
		LLMBackend: LLMBackendMock,
		// Leave other fields at zero values
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Check defaults were applied
	if config.RequestTimeout != 30*time.Second {
		t.Errorf("RequestTimeout = %v, want 30s", config.RequestTimeout)
	}
	if config.MinConfidenceThreshold != 0.5 {
		t.Errorf("MinConfidenceThreshold = %v, want 0.5", config.MinConfidenceThreshold)
	}
	if config.MaxHistoryLength != 10 {
		t.Errorf("MaxHistoryLength = %v, want 10", config.MaxHistoryLength)
	}
	if config.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %v, want 1024", config.MaxTokens)
	}
	if config.DefaultTemperature != 0.7 {
		t.Errorf("DefaultTemperature = %v, want 0.7", config.DefaultTemperature)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.LLMBackend != LLMBackendMock {
		t.Errorf("LLMBackend = %v, want %v", config.LLMBackend, LLMBackendMock)
	}
	if config.MaxHistoryLength != 10 {
		t.Errorf("MaxHistoryLength = %v, want 10", config.MaxHistoryLength)
	}
	if config.RequestTimeout != 30*time.Second {
		t.Errorf("RequestTimeout = %v, want 30s", config.RequestTimeout)
	}
	if config.MinConfidenceThreshold != 0.5 {
		t.Errorf("MinConfidenceThreshold = %v, want 0.5", config.MinConfidenceThreshold)
	}
	if config.RateLimitRequests != 60 {
		t.Errorf("RateLimitRequests = %v, want 60", config.RateLimitRequests)
	}
	if !config.EnableQueryExecution {
		t.Error("EnableQueryExecution should be true by default")
	}
	if !config.EnableSuggestions {
		t.Error("EnableSuggestions should be true by default")
	}
	if config.SystemPrompt == "" {
		t.Error("SystemPrompt should not be empty")
	}
	if config.DistributedRateLimiter.RedisPrefix == "" {
		t.Error("DistributedRateLimiter.RedisPrefix should not be empty")
	}
}

func TestLLMBackendType_Values(t *testing.T) {
	// Verify backend type constants
	if LLMBackendOpenAI != "openai" {
		t.Errorf("LLMBackendOpenAI = %v, want openai", LLMBackendOpenAI)
	}
	if LLMBackendAnthropic != "anthropic" {
		t.Errorf("LLMBackendAnthropic = %v, want anthropic", LLMBackendAnthropic)
	}
	if LLMBackendLocal != "local" {
		t.Errorf("LLMBackendLocal = %v, want local", LLMBackendLocal)
	}
	if LLMBackendMock != "mock" {
		t.Errorf("LLMBackendMock = %v, want mock", LLMBackendMock)
	}
}
