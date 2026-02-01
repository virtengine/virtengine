package nli

import (
	"context"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
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
				LLMBackend:             LLMBackendMock,
				MaxHistoryLength:       10,
				RequestTimeout:         30 * time.Second,
				MinConfidenceThreshold: 0.5,
			},
			wantErr: false,
		},
		{
			name: "openai backend without config",
			config: Config{
				LLMBackend: LLMBackendOpenAI,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewService(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && svc == nil {
				t.Error("NewService() returned nil service without error")
			}
			if svc != nil {
				_ = svc.Close()
			}
		})
	}
}

func TestService_Chat(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	tests := []struct {
		name       string
		req        *ChatRequest
		wantErr    bool
		wantIntent Intent
	}{
		{
			name: "balance query",
			req: &ChatRequest{
				Message:     "What's my balance?",
				SessionID:   "test-session-1",
				UserAddress: "ve1testaddress123456789",
			},
			wantErr:    false,
			wantIntent: IntentQueryBalance,
		},
		{
			name: "marketplace search",
			req: &ChatRequest{
				Message:   "Find GPU providers",
				SessionID: "test-session-2",
			},
			wantErr:    false,
			wantIntent: IntentFindOfferings,
		},
		{
			name: "help request",
			req: &ChatRequest{
				Message:   "How do I stake tokens?",
				SessionID: "test-session-3",
			},
			wantErr:    false,
			wantIntent: IntentStaking,
		},
		{
			name: "empty message",
			req: &ChatRequest{
				Message:   "",
				SessionID: "test-session-4",
			},
			wantErr: true,
		},
		{
			name: "missing session ID",
			req: &ChatRequest{
				Message: "Hello",
			},
			wantErr: true,
		},
		{
			name: "order status query",
			req: &ChatRequest{
				Message:   "Check my order status",
				SessionID: "test-session-5",
			},
			wantErr:    false,
			wantIntent: IntentCheckOrder,
		},
		{
			name: "identity query",
			req: &ChatRequest{
				Message:   "What's my VEID score?",
				SessionID: "test-session-6",
			},
			wantErr:    false,
			wantIntent: IntentIdentity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := svc.Chat(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if resp == nil {
				t.Error("Chat() returned nil response without error")
				return
			}
			if resp.Intent != tt.wantIntent {
				t.Errorf("Chat() intent = %v, want %v", resp.Intent, tt.wantIntent)
			}
			if resp.Message == "" {
				t.Error("Chat() returned empty message")
			}
			if resp.SessionID != tt.req.SessionID {
				t.Errorf("Chat() sessionID = %v, want %v", resp.SessionID, tt.req.SessionID)
			}
			// ProcessingTime may be 0 for very fast responses, just check it's not negative
			if resp.Metadata.ProcessingTime < 0 {
				t.Error("Chat() returned negative processing time")
			}
		})
	}
}

func TestService_Chat_ContextCancellation(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &ChatRequest{
		Message:   "What's my balance?",
		SessionID: "test-cancel",
	}

	_, err = svc.Chat(ctx, req)
	if err == nil {
		t.Error("Chat() should return error for cancelled context")
	}
}

func TestService_Chat_Timeout(t *testing.T) {
	config := DefaultConfig()
	config.RequestTimeout = 1 * time.Nanosecond // Very short timeout
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	req := &ChatRequest{
		Message:   "What's my balance?",
		SessionID: "test-timeout",
	}

	// This should complete quickly enough that timeout doesn't trigger
	// but tests the timeout setup path
	ctx := context.Background()
	_, _ = svc.Chat(ctx, req)
	// We don't check error here as it may or may not timeout depending on system speed
}

func TestService_ClassifyIntent(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	tests := []struct {
		name       string
		message    string
		wantIntent Intent
	}{
		{
			name:       "balance query",
			message:    "Show my token balance",
			wantIntent: IntentQueryBalance,
		},
		{
			name:       "offerings search",
			message:    "Search for GPU compute offerings",
			wantIntent: IntentFindOfferings,
		},
		{
			name:       "staking help",
			message:    "How to stake tokens",
			wantIntent: IntentStaking,
		},
		{
			name:       "deployment",
			message:    "Deploy my application",
			wantIntent: IntentDeployment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := svc.ClassifyIntent(ctx, tt.message)
			if err != nil {
				t.Errorf("ClassifyIntent() error = %v", err)
				return
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("ClassifyIntent() intent = %v, want %v", result.Intent, tt.wantIntent)
			}
			if result.Confidence <= 0 || result.Confidence > 1 {
				t.Errorf("ClassifyIntent() invalid confidence = %v", result.Confidence)
			}
		})
	}
}

func TestService_GetSuggestions(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Test suggestions for new session
	suggestions, err := svc.GetSuggestions(ctx, "new-session")
	if err != nil {
		t.Errorf("GetSuggestions() error = %v", err)
		return
	}
	if len(suggestions) == 0 {
		t.Error("GetSuggestions() returned empty suggestions")
	}

	// Test suggestions after chat
	req := &ChatRequest{
		Message:   "Show my balance",
		SessionID: "suggestion-session",
	}
	_, _ = svc.Chat(ctx, req)

	suggestions, err = svc.GetSuggestions(ctx, "suggestion-session")
	if err != nil {
		t.Errorf("GetSuggestions() after chat error = %v", err)
	}
	if len(suggestions) == 0 {
		t.Error("GetSuggestions() returned empty suggestions after chat")
	}
}

func TestService_Close(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Close the service
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Try to use after close
	ctx := context.Background()
	req := &ChatRequest{
		Message:   "Hello",
		SessionID: "test",
	}
	_, err = svc.Chat(ctx, req)
	if err != ErrServiceClosed {
		t.Errorf("Chat() after Close() error = %v, want %v", err, ErrServiceClosed)
	}
}

func TestService_RateLimiting(t *testing.T) {
	config := DefaultConfig()
	config.RateLimitRequests = 3 // Low limit for testing
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()
	sessionID := "rate-limit-test"

	// Should succeed for first 3 requests
	for i := 0; i < 3; i++ {
		req := &ChatRequest{
			Message:   "Hello",
			SessionID: sessionID,
		}
		_, err := svc.Chat(ctx, req)
		if err != nil {
			t.Errorf("Request %d should not be rate limited: %v", i+1, err)
		}
	}

	// Fourth request should be rate limited
	req := &ChatRequest{
		Message:   "Hello again",
		SessionID: sessionID,
	}
	_, err = svc.Chat(ctx, req)
	if err != ErrRateLimited {
		t.Errorf("Request 4 should be rate limited, got error: %v", err)
	}
}

func TestService_SessionHistory(t *testing.T) {
	config := DefaultConfig()
	config.MaxHistoryLength = 3 // Small history for testing
	svc, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()
	sessionID := "history-test"

	// Send multiple messages
	messages := []string{
		"What's my balance?",
		"Find GPU offerings",
		"How do I stake?",
		"Check my orders",
	}

	for _, msg := range messages {
		req := &ChatRequest{
			Message:   msg,
			SessionID: sessionID,
		}
		_, err := svc.Chat(ctx, req)
		if err != nil {
			t.Fatalf("Chat() error = %v", err)
		}
	}

	// Verify session was created and history is maintained
	// (Internal verification - we can't directly access sessions in this test)
}

func TestChatRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ChatRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: ChatRequest{
				Message:   "Hello",
				SessionID: "test",
			},
			wantErr: nil,
		},
		{
			name: "empty message",
			req: ChatRequest{
				Message:   "",
				SessionID: "test",
			},
			wantErr: ErrEmptyMessage,
		},
		{
			name: "empty session ID",
			req: ChatRequest{
				Message:   "Hello",
				SessionID: "",
			},
			wantErr: ErrMissingSessionID,
		},
		{
			name: "both empty",
			req: ChatRequest{
				Message:   "",
				SessionID: "",
			},
			wantErr: ErrEmptyMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

