package nli

import (
	"context"
	"testing"
)

func TestDefaultResponseGenerator_Generate(t *testing.T) {
	config := DefaultConfig()
	generator := NewDefaultResponseGenerator(config)
	ctx := context.Background()

	tests := []struct {
		name        string
		intent      Intent
		entities    map[string]string
		queryResult *QueryResult
		chatCtx     *ChatContext
		wantEmpty   bool
	}{
		{
			name:     "balance response with data",
			intent:   IntentQueryBalance,
			entities: map[string]string{},
			queryResult: &QueryResult{
				Success:   true,
				QueryType: "balance",
				Data: []BalanceInfo{
					{Denom: "uve", Amount: "1000", USD: "100.00"},
				},
			},
			wantEmpty: false,
		},
		{
			name:        "balance response without data",
			intent:      IntentQueryBalance,
			entities:    map[string]string{},
			queryResult: nil,
			wantEmpty:   false,
		},
		{
			name:     "offerings response with data",
			intent:   IntentFindOfferings,
			entities: map[string]string{"resource_type": "gpu"},
			queryResult: &QueryResult{
				Success:   true,
				QueryType: "offerings",
				Data: []OfferingInfo{
					{ID: "1", Provider: "ve1test", Type: "GPU", Price: "10 UVE", Available: true},
				},
			},
			wantEmpty: false,
		},
		{
			name:     "offerings response empty",
			intent:   IntentFindOfferings,
			entities: map[string]string{},
			queryResult: &QueryResult{
				Success:   true,
				QueryType: "offerings",
				Data:      []OfferingInfo{},
			},
			wantEmpty: false,
		},
		{
			name:      "help response",
			intent:    IntentGetHelp,
			entities:  map[string]string{},
			wantEmpty: false,
		},
		{
			name:     "order response with data",
			intent:   IntentCheckOrder,
			entities: map[string]string{},
			queryResult: &QueryResult{
				Success:   true,
				QueryType: "order",
				Data: OrderInfo{
					ID:       "123",
					Status:   "active",
					Provider: "ve1test",
				},
			},
			wantEmpty: false,
		},
		{
			name:      "staking response",
			intent:    IntentStaking,
			entities:  map[string]string{},
			wantEmpty: false,
		},
		{
			name:      "deployment response",
			intent:    IntentDeployment,
			entities:  map[string]string{},
			wantEmpty: false,
		},
		{
			name:     "identity response verified",
			intent:   IntentIdentity,
			entities: map[string]string{},
			chatCtx: &ChatContext{
				UserProfile: &UserProfile{
					Address:       "ve1test",
					IsVerified:    true,
					IdentityScore: 0.85,
				},
			},
			wantEmpty: false,
		},
		{
			name:     "identity response not verified",
			intent:   IntentIdentity,
			entities: map[string]string{},
			chatCtx: &ChatContext{
				UserProfile: &UserProfile{
					Address:    "ve1test",
					IsVerified: false,
				},
			},
			wantEmpty: false,
		},
		{
			name:      "general chat response",
			intent:    IntentGeneralChat,
			entities:  map[string]string{},
			wantEmpty: false,
		},
		{
			name:     "provider info response",
			intent:   IntentGetProviderInfo,
			entities: map[string]string{"address": "ve1provider"},
			queryResult: &QueryResult{
				Success: true,
				Data:    map[string]interface{}{"name": "Test Provider"},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := generator.Generate(ctx, tt.intent, tt.entities, tt.queryResult, tt.chatCtx)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}
			if (response == "") != tt.wantEmpty {
				t.Errorf("Generate() response empty = %v, want empty = %v", response == "", tt.wantEmpty)
			}
		})
	}
}

func TestDefaultResponseGenerator_ContextCancellation(t *testing.T) {
	config := DefaultConfig()
	generator := NewDefaultResponseGenerator(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := generator.Generate(ctx, IntentGetHelp, nil, nil, nil)
	if err == nil {
		t.Error("Generate() should return error for cancelled context")
	}
}

func TestTruncateAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "ve1abc",
			expected: "ve1abc",
		},
		{
			input:    "ve1abcdefghijklmnopqrstuvwxyz1234567890abcdef",
			expected: "ve1abcde...abcdef",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "short",
			expected: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateAddress(tt.input)
			if result != tt.expected {
				t.Errorf("truncateAddress(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"open", "Open"},
		{"active", "Active"},
		{"closed", "Closed"},
		{"matched", "Matched"},
		{"failed", "Failed"},
		{"unknown", "unknown"},
		{"OPEN", "Open"}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatStatus(tt.input)
			if result == "" {
				t.Errorf("formatStatus(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestDefaultQueryExecutor_Execute(t *testing.T) {
	executor := NewDefaultQueryExecutor()
	ctx := context.Background()

	tests := []struct {
		name        string
		intent      Intent
		entities    map[string]string
		userAddress string
		wantSuccess bool
	}{
		{
			name:        "balance query with address",
			intent:      IntentQueryBalance,
			entities:    map[string]string{},
			userAddress: "ve1test123",
			wantSuccess: true,
		},
		{
			name:        "balance query without address",
			intent:      IntentQueryBalance,
			entities:    map[string]string{},
			userAddress: "",
			wantSuccess: false,
		},
		{
			name:        "offerings query",
			intent:      IntentFindOfferings,
			entities:    map[string]string{"resource_type": "gpu"},
			userAddress: "",
			wantSuccess: true,
		},
		{
			name:        "order query with order ID",
			intent:      IntentCheckOrder,
			entities:    map[string]string{"order_id": "123"},
			userAddress: "",
			wantSuccess: true,
		},
		{
			name:        "order query without ID or address",
			intent:      IntentCheckOrder,
			entities:    map[string]string{},
			userAddress: "",
			wantSuccess: true, // Returns mock data when no querier is configured
		},
		{
			name:        "provider query with address",
			intent:      IntentGetProviderInfo,
			entities:    map[string]string{"address": "ve1provider"},
			userAddress: "",
			wantSuccess: true,
		},
		{
			name:        "provider query without address",
			intent:      IntentGetProviderInfo,
			entities:    map[string]string{},
			userAddress: "",
			wantSuccess: false,
		},
		{
			name:        "identity query with address",
			intent:      IntentIdentity,
			entities:    map[string]string{},
			userAddress: "ve1test",
			wantSuccess: true,
		},
		{
			name:        "general chat (no query needed)",
			intent:      IntentGeneralChat,
			entities:    map[string]string{},
			userAddress: "",
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.intent, tt.entities, tt.userAddress)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Execute() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}

func TestDefaultQueryExecutor_ContextCancellation(t *testing.T) {
	executor := NewDefaultQueryExecutor()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executor.Execute(ctx, IntentQueryBalance, nil, "ve1test")
	if err == nil {
		t.Error("Execute() should return error for cancelled context")
	}
}

func TestMockLLMBackend_Complete(t *testing.T) {
	backend := NewMockLLMBackend()
	ctx := context.Background()

	// Test default response
	result, err := backend.Complete(ctx, "Hello", nil)
	if err != nil {
		t.Errorf("Complete() error = %v", err)
		return
	}
	if result.Text == "" {
		t.Error("Complete() returned empty text")
	}

	// Test custom response
	backend.SetResponse("custom", "This is a custom response")
	result, err = backend.Complete(ctx, "Say something custom", nil)
	if err != nil {
		t.Errorf("Complete() error = %v", err)
		return
	}
	if result.Text != "This is a custom response" {
		t.Errorf("Complete() = %q, want custom response", result.Text)
	}
}

func TestMockLLMBackend_ClassifyIntent(t *testing.T) {
	backend := NewMockLLMBackend()
	ctx := context.Background()

	tests := []struct {
		message    string
		wantIntent Intent
	}{
		{"What's my balance?", IntentQueryBalance},
		{"Show offerings", IntentFindOfferings},
		{"Help me", IntentGetHelp},
		{"Hello", IntentGeneralChat},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result, err := backend.ClassifyIntent(ctx, tt.message)
			if err != nil {
				t.Errorf("ClassifyIntent() error = %v", err)
				return
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("ClassifyIntent() = %v, want %v", result.Intent, tt.wantIntent)
			}
		})
	}
}

func TestMockLLMBackend_ContextCancellation(t *testing.T) {
	backend := NewMockLLMBackend()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := backend.Complete(ctx, "Hello", nil)
	if err == nil {
		t.Error("Complete() should return error for cancelled context")
	}

	_, err = backend.ClassifyIntent(ctx, "Hello")
	if err == nil {
		t.Error("ClassifyIntent() should return error for cancelled context")
	}
}

func TestCreateLLMBackend(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "mock backend",
			config:  Config{LLMBackend: LLMBackendMock},
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
			name: "anthropic without config",
			config: Config{
				LLMBackend: LLMBackendAnthropic,
			},
			wantErr: true,
		},
		{
			name: "local without config",
			config: Config{
				LLMBackend: LLMBackendLocal,
			},
			wantErr: true,
		},
		{
			name:    "unknown backend",
			config:  Config{LLMBackend: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := CreateLLMBackend(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLLMBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && backend == nil {
				t.Error("CreateLLMBackend() returned nil without error")
			}
		})
	}
}
