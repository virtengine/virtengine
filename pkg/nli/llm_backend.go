package nli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
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
// OpenAI LLM Backend
// ============================================================================

const (
	defaultOpenAIBaseURL     = "https://api.openai.com/v1"
	defaultOpenAIModel       = "gpt-4o-mini"
	defaultOpenAIMaxTokens   = 1024
	defaultOpenAITemperature = 0.7
	defaultOpenAITimeout     = 30 * time.Second
)

// OpenAIBackend implements LLMBackend using the OpenAI API
type OpenAIBackend struct {
	config     *OpenAIConfig
	httpClient *http.Client
	baseURL    string
}

// openAIChatRequest represents the request body for OpenAI Chat Completions API
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float32             `json:"temperature,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
}

// openAIChatMessage represents a message in the OpenAI chat format
type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse represents the response from OpenAI Chat Completions API
type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int               `json:"index"`
		Message      openAIChatMessage `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error,omitempty"`
}

// openAIError represents an error response from OpenAI
type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIBackend creates a new OpenAI backend
func NewOpenAIBackend(config *OpenAIConfig) (*OpenAIBackend, error) {
	if config == nil || config.APIKey == "" {
		return nil, ErrInvalidConfig
	}

	// Set defaults
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}

	timeout := defaultOpenAITimeout

	return &OpenAIBackend{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}, nil
}

// Complete generates a completion using OpenAI Chat Completions API
func (o *OpenAIBackend) Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error) {
	// Check context cancellation early
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build request parameters with defaults
	model := o.config.Model
	if model == "" {
		model = defaultOpenAIModel
	}

	maxTokens := defaultOpenAIMaxTokens
	temperature := float32(defaultOpenAITemperature)

	if options != nil {
		if options.MaxTokens > 0 {
			maxTokens = options.MaxTokens
		}
		if options.Temperature > 0 {
			temperature = options.Temperature
		}
	}

	// Build messages
	messages := make([]openAIChatMessage, 0, 2)

	// Add system prompt if provided
	if options != nil && options.SystemPrompt != "" {
		messages = append(messages, openAIChatMessage{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	// Add user message
	messages = append(messages, openAIChatMessage{
		Role:    "user",
		Content: prompt,
	})

	// Build request body
	reqBody := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	if options != nil && len(options.StopSequences) > 0 {
		reqBody.Stop = options.StopSequences
	}

	// Serialize request
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", ErrLLMCompletionFailed, err)
	}

	// Create HTTP request
	url := o.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrLLMCompletionFailed, err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)
	if o.config.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", o.config.OrganizationID)
	}

	// Execute request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: request failed: %v", ErrLLMBackendUnavailable, err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrLLMCompletionFailed, err)
	}

	// Handle error status codes
	if resp.StatusCode != http.StatusOK {
		return nil, o.handleErrorResponse(resp.StatusCode, respBody)
	}

	// Parse response
	var chatResp openAIChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrInvalidLLMResponse, err)
	}

	// Check for API error in response body
	if chatResp.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrLLMCompletionFailed, chatResp.Error.Message)
	}

	// Validate response has choices
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrInvalidLLMResponse)
	}

	return &CompletionResult{
		Text:         chatResp.Choices[0].Message.Content,
		TokensUsed:   chatResp.Usage.TotalTokens,
		FinishReason: chatResp.Choices[0].FinishReason,
	}, nil
}

// handleErrorResponse handles non-200 responses from OpenAI
func (o *OpenAIBackend) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse error response
	var errResp struct {
		Error openAIError `json:"error"`
	}
	_ = json.Unmarshal(body, &errResp)

	errMsg := errResp.Error.Message
	if errMsg == "" {
		errMsg = string(body)
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: authentication failed: %s", ErrLLMBackendUnavailable, errMsg)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: rate limit exceeded: %s", ErrRateLimited, errMsg)
	case http.StatusBadRequest:
		return fmt.Errorf("%w: bad request: %s", ErrLLMCompletionFailed, errMsg)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("%w: service unavailable (status %d): %s", ErrLLMBackendUnavailable, statusCode, errMsg)
	default:
		return fmt.Errorf("%w: unexpected status %d: %s", ErrLLMCompletionFailed, statusCode, errMsg)
	}
}

// ClassifyIntent classifies intent using OpenAI
func (o *OpenAIBackend) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	// Check context cancellation early
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build classification prompt
	systemPrompt := `You are an intent classifier for VirtEngine, a decentralized cloud computing platform.
Classify the user's message into exactly one of these intents:
- query_balance: User wants to check their token balance
- find_offerings: User wants to search marketplace offerings or find providers
- get_help: User needs help or has questions about how to use something
- check_order: User wants to check their order or deployment status
- get_provider_info: User wants information about a specific provider
- staking: User wants to stake, delegate, or manage their staked tokens
- deployment: User wants help with deployments or workloads
- identity: User has questions about identity verification or VEID
- general_chat: General conversation not fitting other categories
- unknown: Cannot determine the intent

Respond with ONLY a JSON object in this exact format:
{"intent": "<intent_name>", "confidence": <0.0-1.0>}

Do not include any other text.`

	options := &CompletionOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    100, // Classification needs few tokens
		Temperature:  0.1, // Low temperature for consistent classification
	}

	result, err := o.Complete(ctx, message, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClassificationFailed, err)
	}

	// Parse the classification response
	return o.parseClassificationResponse(result.Text)
}

// parseClassificationResponse parses the JSON classification response from the LLM
func (o *OpenAIBackend) parseClassificationResponse(response string) (*ClassificationResult, error) {
	// Clean up the response (remove potential markdown code blocks)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Parse JSON
	var parsed struct {
		Intent     string  `json:"intent"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to extract intent from raw text as fallback
		return o.fallbackParseIntent(response)
	}

	// Map string to Intent type
	intent := Intent(parsed.Intent)
	if !isValidIntent(intent) {
		intent = IntentUnknown
	}

	// Clamp confidence to valid range
	confidence := float32(parsed.Confidence)
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return &ClassificationResult{
		Intent:     intent,
		Confidence: confidence,
	}, nil
}

// fallbackParseIntent attempts to extract intent from non-JSON response
func (o *OpenAIBackend) fallbackParseIntent(response string) (*ClassificationResult, error) {
	lower := strings.ToLower(response)

	// Check for known intent strings
	intents := []Intent{
		IntentQueryBalance,
		IntentFindOfferings,
		IntentGetHelp,
		IntentCheckOrder,
		IntentGetProviderInfo,
		IntentStaking,
		IntentDeployment,
		IntentIdentity,
		IntentGeneralChat,
	}

	for _, intent := range intents {
		if strings.Contains(lower, string(intent)) {
			return &ClassificationResult{
				Intent:     intent,
				Confidence: 0.6, // Lower confidence for fallback parsing
			}, nil
		}
	}

	return &ClassificationResult{
		Intent:     IntentUnknown,
		Confidence: 0.3,
	}, nil
}

// isValidIntent checks if an intent is a known valid intent
func isValidIntent(intent Intent) bool {
	switch intent {
	case IntentQueryBalance, IntentFindOfferings, IntentGetHelp, IntentCheckOrder,
		IntentGetProviderInfo, IntentStaking, IntentDeployment, IntentIdentity,
		IntentGeneralChat, IntentUnknown:
		return true
	default:
		return false
	}
}

// Close closes the OpenAI backend
func (o *OpenAIBackend) Close() error {
	// Close idle connections
	o.httpClient.CloseIdleConnections()
	return nil
}

// SetHTTPClient allows setting a custom HTTP client (useful for testing)
func (o *OpenAIBackend) SetHTTPClient(client *http.Client) {
	o.httpClient = client
}

// extractConfidenceFromText attempts to extract confidence score from text
//
//nolint:unused // Reserved for future LLM response parsing
func extractConfidenceFromText(text string) float32 {
	// Try to find patterns like "0.9", "0.85", "confidence: 0.7"
	text = strings.ToLower(text)

	// Look for decimal numbers that could be confidence scores
	for i := 0; i < len(text)-2; i++ {
		if text[i] == '0' && text[i+1] == '.' {
			// Found potential confidence score
			end := i + 2
			for end < len(text) && text[end] >= '0' && text[end] <= '9' {
				end++
			}
			if end > i+2 {
				if val, err := strconv.ParseFloat(text[i:end], 32); err == nil && val >= 0 && val <= 1 {
					return float32(val)
				}
			}
		}
	}

	return 0.5 // Default confidence
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
