package nli

import (
	"context"
	"time"
)

// ============================================================================
// Intent Types
// ============================================================================

// Intent represents a classified user intent
type Intent string

const (
	// IntentQueryBalance indicates the user wants to check their token balance
	IntentQueryBalance Intent = "query_balance"

	// IntentFindOfferings indicates the user wants to search marketplace offerings
	IntentFindOfferings Intent = "find_offerings"

	// IntentGetHelp indicates the user needs help with a topic
	IntentGetHelp Intent = "get_help"

	// IntentCheckOrder indicates the user wants to check order status
	IntentCheckOrder Intent = "check_order"

	// IntentGetProviderInfo indicates the user wants information about a provider
	IntentGetProviderInfo Intent = "get_provider_info"

	// IntentStaking indicates the user wants to stake or delegate tokens
	IntentStaking Intent = "staking"

	// IntentDeployment indicates the user wants help with deployments
	IntentDeployment Intent = "deployment"

	// IntentIdentity indicates the user has identity/VEID related questions
	IntentIdentity Intent = "identity"

	// IntentGeneralChat is a fallback for general conversation
	IntentGeneralChat Intent = "general_chat"

	// IntentUnknown indicates the intent could not be classified
	IntentUnknown Intent = "unknown"
)

// String returns the string representation of an Intent
func (i Intent) String() string {
	return string(i)
}

// ============================================================================
// LLM Backend Types
// ============================================================================

// LLMBackendType specifies the LLM backend to use
type LLMBackendType string

const (
	// LLMBackendOpenAI uses OpenAI's API
	LLMBackendOpenAI LLMBackendType = "openai"

	// LLMBackendAnthropic uses Anthropic's Claude API
	LLMBackendAnthropic LLMBackendType = "anthropic"

	// LLMBackendLocal uses a local LLM (e.g., llama.cpp)
	LLMBackendLocal LLMBackendType = "local"

	// LLMBackendMock uses a mock backend for testing
	LLMBackendMock LLMBackendType = "mock"
)

// ============================================================================
// Request/Response Types
// ============================================================================

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	// Message is the user's input message
	Message string `json:"message"`

	// SessionID identifies the conversation session
	SessionID string `json:"session_id"`

	// UserAddress is the blockchain address of the user (optional)
	UserAddress string `json:"user_address,omitempty"`

	// Context provides additional context for the conversation
	Context *ChatContext `json:"context,omitempty"`
}

// Validate validates the chat request
func (r *ChatRequest) Validate() error {
	if r.Message == "" {
		return ErrEmptyMessage
	}
	if r.SessionID == "" {
		return ErrMissingSessionID
	}
	return nil
}

// ChatResponse represents the response to a chat request
type ChatResponse struct {
	// Message is the AI-generated response
	Message string `json:"message"`

	// Intent is the classified intent of the user's message
	Intent Intent `json:"intent"`

	// Confidence is the confidence score of intent classification (0.0-1.0)
	Confidence float32 `json:"confidence"`

	// SessionID identifies the conversation session
	SessionID string `json:"session_id"`

	// Data contains structured data relevant to the response
	Data map[string]interface{} `json:"data,omitempty"`

	// Suggestions are follow-up actions or questions
	Suggestions []string `json:"suggestions,omitempty"`

	// Metadata contains response metadata
	Metadata ResponseMetadata `json:"metadata"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	// ProcessingTime is how long the request took to process
	ProcessingTime time.Duration `json:"processing_time"`

	// TokensUsed is the number of tokens consumed (for LLM backends)
	TokensUsed int `json:"tokens_used,omitempty"`

	// Model is the model used for generation
	Model string `json:"model,omitempty"`

	// Timestamp is when the response was generated
	Timestamp time.Time `json:"timestamp"`
}

// ChatContext provides context for the conversation
type ChatContext struct {
	// History is the conversation history
	History []ChatMessage `json:"history,omitempty"`

	// UserProfile contains user-specific context
	UserProfile *UserProfile `json:"user_profile,omitempty"`

	// BlockHeight is the current blockchain height for context
	BlockHeight int64 `json:"block_height,omitempty"`

	// NetworkID is the blockchain network ID
	NetworkID string `json:"network_id,omitempty"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	// Role is either "user" or "assistant"
	Role string `json:"role"`

	// Content is the message content
	Content string `json:"content"`

	// Timestamp is when the message was sent
	Timestamp time.Time `json:"timestamp"`
}

// UserProfile contains user-specific context information
type UserProfile struct {
	// Address is the user's blockchain address
	Address string `json:"address,omitempty"`

	// IsVerified indicates if the user has completed VEID verification
	IsVerified bool `json:"is_verified"`

	// IdentityScore is the user's VEID identity score (0.0-1.0)
	IdentityScore float32 `json:"identity_score,omitempty"`

	// Roles are the user's assigned roles
	Roles []string `json:"roles,omitempty"`

	// Preferences are user preferences for the chat interface
	Preferences map[string]string `json:"preferences,omitempty"`
}

// ============================================================================
// Intent Classification Types
// ============================================================================

// ClassificationResult represents the result of intent classification
type ClassificationResult struct {
	// Intent is the classified intent
	Intent Intent `json:"intent"`

	// Confidence is the confidence score (0.0-1.0)
	Confidence float32 `json:"confidence"`

	// Entities are extracted entities from the message
	Entities map[string]string `json:"entities,omitempty"`

	// AlternativeIntents are other possible intents with lower confidence
	AlternativeIntents []IntentScore `json:"alternative_intents,omitempty"`
}

// IntentScore pairs an intent with its confidence score
type IntentScore struct {
	Intent     Intent  `json:"intent"`
	Confidence float32 `json:"confidence"`
}

// ============================================================================
// Query Executor Types
// ============================================================================

// QueryResult represents the result of a blockchain query
type QueryResult struct {
	// Success indicates if the query was successful
	Success bool `json:"success"`

	// Data contains the query result data
	Data interface{} `json:"data,omitempty"`

	// Error contains error information if the query failed
	Error string `json:"error,omitempty"`

	// QueryType identifies the type of query executed
	QueryType string `json:"query_type"`
}

// BalanceInfo represents token balance information
type BalanceInfo struct {
	// Denom is the token denomination
	Denom string `json:"denom"`

	// Amount is the token amount
	Amount string `json:"amount"`

	// USD is the USD equivalent (if available)
	USD string `json:"usd,omitempty"`
}

// OfferingInfo represents a marketplace offering
type OfferingInfo struct {
	// ID is the offering identifier
	ID string `json:"id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// Type is the offering type (compute, storage, etc.)
	Type string `json:"type"`

	// Specs contains the offering specifications
	Specs map[string]string `json:"specs"`

	// Price is the offering price
	Price string `json:"price"`

	// Available indicates if the offering is available
	Available bool `json:"available"`
}

// OrderInfo represents order status information
type OrderInfo struct {
	// ID is the order identifier
	ID string `json:"id"`

	// Status is the current order status
	Status string `json:"status"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// CreatedAt is when the order was created
	CreatedAt time.Time `json:"created_at"`

	// Details contains additional order details
	Details map[string]string `json:"details,omitempty"`
}

// ============================================================================
// Interfaces
// ============================================================================

// Service is the main NLI service interface
type Service interface {
	// Chat processes a chat request and returns a response
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ClassifyIntent classifies the intent of a message
	ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error)

	// GetSuggestions returns suggested actions or questions
	GetSuggestions(ctx context.Context, sessionID string) ([]string, error)

	// Close closes the service and releases resources
	Close() error
}

// Classifier classifies user intents
type Classifier interface {
	// Classify classifies the intent of a message
	Classify(ctx context.Context, message string, context *ChatContext) (*ClassificationResult, error)
}

// ResponseGenerator generates responses based on intent and context
type ResponseGenerator interface {
	// Generate generates a response for the given intent and context
	Generate(ctx context.Context, intent Intent, entities map[string]string, queryResult *QueryResult, context *ChatContext) (string, error)
}

// QueryExecutor executes blockchain queries
type QueryExecutor interface {
	// Execute executes a query based on intent and entities
	Execute(ctx context.Context, intent Intent, entities map[string]string, userAddress string) (*QueryResult, error)
}

// LLMBackend is the interface for LLM providers
type LLMBackend interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string, options *CompletionOptions) (*CompletionResult, error)

	// ClassifyIntent uses the LLM to classify intent
	ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error)

	// Close closes the backend connection
	Close() error
}

// CompletionOptions contains options for LLM completion
type CompletionOptions struct {
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0.0-2.0)
	Temperature float32 `json:"temperature,omitempty"`

	// SystemPrompt is the system prompt for the conversation
	SystemPrompt string `json:"system_prompt,omitempty"`

	// StopSequences are sequences that stop generation
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// CompletionResult is the result of an LLM completion
type CompletionResult struct {
	// Text is the generated text
	Text string `json:"text"`

	// TokensUsed is the number of tokens consumed
	TokensUsed int `json:"tokens_used"`

	// FinishReason indicates why generation stopped
	FinishReason string `json:"finish_reason"`
}

