package nli

import (
	"time"
)

// ============================================================================
// Configuration
// ============================================================================

// Config contains the NLI service configuration
type Config struct {
	// LLMBackend specifies which LLM backend to use
	LLMBackend LLMBackendType `json:"llm_backend"`

	// OpenAIConfig contains OpenAI-specific configuration
	OpenAIConfig *OpenAIConfig `json:"openai_config,omitempty"`

	// AnthropicConfig contains Anthropic-specific configuration
	AnthropicConfig *AnthropicConfig `json:"anthropic_config,omitempty"`

	// LocalConfig contains local LLM configuration
	LocalConfig *LocalConfig `json:"local_config,omitempty"`

	// MaxHistoryLength is the maximum conversation history to retain
	MaxHistoryLength int `json:"max_history_length"`

	// RequestTimeout is the timeout for processing requests
	RequestTimeout time.Duration `json:"request_timeout"`

	// MinConfidenceThreshold is the minimum confidence for intent classification
	MinConfidenceThreshold float32 `json:"min_confidence_threshold"`

	// RateLimitRequests is the max requests per minute per session
	RateLimitRequests int `json:"rate_limit_requests"`

	// EnableQueryExecution enables blockchain query execution
	EnableQueryExecution bool `json:"enable_query_execution"`

	// SystemPrompt is the base system prompt for the LLM
	SystemPrompt string `json:"system_prompt"`

	// DefaultTemperature is the default temperature for LLM generation
	DefaultTemperature float32 `json:"default_temperature"`

	// MaxTokens is the maximum tokens for response generation
	MaxTokens int `json:"max_tokens"`

	// EnableSuggestions enables follow-up suggestions
	EnableSuggestions bool `json:"enable_suggestions"`

	// LogLevel controls logging verbosity
	LogLevel string `json:"log_level"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.LLMBackend == "" {
		return ErrInvalidConfig
	}

	switch c.LLMBackend {
	case LLMBackendOpenAI:
		if c.OpenAIConfig == nil || c.OpenAIConfig.APIKey == "" {
			return ErrInvalidConfig
		}
	case LLMBackendAnthropic:
		if c.AnthropicConfig == nil || c.AnthropicConfig.APIKey == "" {
			return ErrInvalidConfig
		}
	case LLMBackendLocal:
		if c.LocalConfig == nil || c.LocalConfig.ModelPath == "" {
			return ErrInvalidConfig
		}
	case LLMBackendMock:
		// Mock backend doesn't require additional config
	default:
		return ErrInvalidConfig
	}

	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}

	if c.MinConfidenceThreshold <= 0 || c.MinConfidenceThreshold > 1 {
		c.MinConfidenceThreshold = 0.5
	}

	if c.MaxHistoryLength <= 0 {
		c.MaxHistoryLength = 10
	}

	if c.MaxTokens <= 0 {
		c.MaxTokens = 1024
	}

	if c.DefaultTemperature <= 0 {
		c.DefaultTemperature = 0.7
	}

	return nil
}

// DefaultConfig returns a default configuration using the mock backend
func DefaultConfig() Config {
	return Config{
		LLMBackend:             LLMBackendMock,
		MaxHistoryLength:       10,
		RequestTimeout:         30 * time.Second,
		MinConfidenceThreshold: 0.5,
		RateLimitRequests:      60,
		EnableQueryExecution:   true,
		SystemPrompt:           DefaultSystemPrompt,
		DefaultTemperature:     0.7,
		MaxTokens:              1024,
		EnableSuggestions:      true,
		LogLevel:               "info",
	}
}

// OpenAIConfig contains OpenAI-specific configuration
type OpenAIConfig struct {
	// APIKey is the OpenAI API key
	APIKey string `json:"api_key"`

	// Model is the model to use (e.g., "gpt-4", "gpt-3.5-turbo")
	Model string `json:"model"`

	// OrganizationID is the optional organization ID
	OrganizationID string `json:"organization_id,omitempty"`

	// BaseURL is an optional custom base URL
	BaseURL string `json:"base_url,omitempty"`
}

// AnthropicConfig contains Anthropic-specific configuration
type AnthropicConfig struct {
	// APIKey is the Anthropic API key
	APIKey string `json:"api_key"`

	// Model is the model to use (e.g., "claude-3-opus", "claude-3-sonnet")
	Model string `json:"model"`

	// BaseURL is an optional custom base URL
	BaseURL string `json:"base_url,omitempty"`
}

// LocalConfig contains local LLM configuration
type LocalConfig struct {
	// ModelPath is the path to the local model
	ModelPath string `json:"model_path"`

	// ContextSize is the context window size
	ContextSize int `json:"context_size"`

	// Threads is the number of threads to use
	Threads int `json:"threads"`

	// GPULayers is the number of layers to offload to GPU
	GPULayers int `json:"gpu_layers"`
}

// DefaultSystemPrompt is the default system prompt for the NLI
const DefaultSystemPrompt = `You are VirtEngine Assistant, an AI helper for the VirtEngine decentralized cloud computing platform.

VirtEngine is a blockchain-based marketplace for cloud computing resources with:
- VEID: Decentralized identity verification using ML-powered scoring
- Marketplace: Buy and sell compute, storage, and GPU resources
- HPC: High-performance computing with SLURM cluster integration
- Staking: Token staking for validators and delegators

Your role is to help users:
1. Understand their account status (balances, identity verification)
2. Find and compare marketplace offerings
3. Manage deployments and orders
4. Learn about VirtEngine features and capabilities
5. Troubleshoot common issues

Be helpful, concise, and accurate. When you don't know something, say so.
If the user asks about their balance or orders, and you have that data, include it in your response.
Format responses in a friendly but professional manner.`
