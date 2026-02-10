package nli

import (
	"context"
	"testing"
)

func TestRuleBasedClassifier_Classify(t *testing.T) {
	classifier := NewRuleBasedClassifier()
	ctx := context.Background()

	tests := []struct {
		name          string
		message       string
		wantIntent    Intent
		minConfidence float32
		wantEntities  map[string]string
	}{
		// Balance queries
		{
			name:          "direct balance query",
			message:       "What's my balance?",
			wantIntent:    IntentQueryBalance,
			minConfidence: 0.7,
		},
		{
			name:          "show tokens",
			message:       "Show me my tokens",
			wantIntent:    IntentQueryBalance,
			minConfidence: 0.7,
		},
		{
			name:          "check funds",
			message:       "How much funds do I have?",
			wantIntent:    IntentQueryBalance,
			minConfidence: 0.7,
		},
		// Marketplace queries
		{
			name:          "find GPU providers",
			message:       "Find GPU providers",
			wantIntent:    IntentFindOfferings,
			minConfidence: 0.7,
		},
		{
			name:          "search offerings",
			message:       "Search for compute offerings",
			wantIntent:    IntentFindOfferings,
			minConfidence: 0.7,
		},
		{
			name:          "browse marketplace",
			message:       "Show me available offerings in the marketplace",
			wantIntent:    IntentFindOfferings,
			minConfidence: 0.6,
		},
		// Help queries
		{
			name:          "how to help",
			message:       "How do I use VirtEngine?",
			wantIntent:    IntentGetHelp,
			minConfidence: 0.6,
		},
		{
			name:          "explain staking",
			message:       "Explain what staking is",
			wantIntent:    IntentGetHelp,
			minConfidence: 0.6,
		},
		// Order queries
		{
			name:          "check order",
			message:       "Check my order status",
			wantIntent:    IntentCheckOrder,
			minConfidence: 0.7,
		},
		{
			name:          "order with ID",
			message:       "What's the status of order #12345?",
			wantIntent:    IntentCheckOrder,
			minConfidence: 0.7,
			wantEntities:  map[string]string{"order_id": "12345"},
		},
		// Provider queries
		{
			name:          "provider info",
			message:       "Tell me about provider ve1abc123def456",
			wantIntent:    IntentGetProviderInfo,
			minConfidence: 0.7,
		},
		// Staking queries
		{
			name:          "stake tokens",
			message:       "How do I stake my tokens?",
			wantIntent:    IntentStaking,
			minConfidence: 0.7,
		},
		{
			name:          "delegation query",
			message:       "I want to delegate to a validator",
			wantIntent:    IntentStaking,
			minConfidence: 0.7,
		},
		// Deployment queries
		{
			name:          "deploy app",
			message:       "Deploy my application",
			wantIntent:    IntentDeployment,
			minConfidence: 0.7,
		},
		{
			name:          "create deployment",
			message:       "Create a new deployment with my manifest",
			wantIntent:    IntentDeployment,
			minConfidence: 0.7,
		},
		// Identity queries
		{
			name:          "VEID score",
			message:       "What's my VEID score?",
			wantIntent:    IntentIdentity,
			minConfidence: 0.8,
		},
		{
			name:          "identity verification",
			message:       "How do I verify my identity?",
			wantIntent:    IntentIdentity,
			minConfidence: 0.7,
		},
		// General/fallback
		{
			name:          "greeting",
			message:       "Hello there!",
			wantIntent:    IntentGeneralChat,
			minConfidence: 0.3,
		},
		{
			name:          "random message",
			message:       "The weather is nice today",
			wantIntent:    IntentGeneralChat,
			minConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := classifier.Classify(ctx, tt.message, nil)
			if err != nil {
				t.Errorf("Classify() error = %v", err)
				return
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("Classify() intent = %v, want %v", result.Intent, tt.wantIntent)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("Classify() confidence = %v, want >= %v", result.Confidence, tt.minConfidence)
			}
			// Check expected entities if specified
			for k, v := range tt.wantEntities {
				if result.Entities[k] != v {
					t.Errorf("Classify() entity[%s] = %v, want %v", k, result.Entities[k], v)
				}
			}
		})
	}
}

func TestRuleBasedClassifier_ExtractEntities(t *testing.T) {
	classifier := NewRuleBasedClassifier()

	tests := []struct {
		name     string
		message  string
		expected map[string]string
	}{
		{
			name:    "extract address",
			message: "Show balance for ve1abcdefghijklmnopqrstuvwxyz1234567890ab",
			expected: map[string]string{
				"address": "ve1abcdefghijklmnopqrstuvwxyz1234567890ab",
			},
		},
		{
			name:    "extract order ID",
			message: "Check order #12345",
			expected: map[string]string{
				"order_id": "12345",
			},
		},
		{
			name:    "extract amount",
			message: "Send 100 uve to someone",
			expected: map[string]string{
				"amount": "100",
				"denom":  "uve",
			},
		},
		{
			name:    "extract resource type",
			message: "Find GPU offerings",
			expected: map[string]string{
				"resource_type": "gpu",
			},
		},
		{
			name:    "extract multiple entities",
			message: "Order #42 for 500 tokens",
			expected: map[string]string{
				"order_id": "42",
				"amount":   "500",
			},
		},
		{
			name:     "no entities",
			message:  "Hello world",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := classifier.extractEntities(tt.message)
			for k, v := range tt.expected {
				if entities[k] != v {
					t.Errorf("extractEntities() entity[%s] = %v, want %v", k, entities[k], v)
				}
			}
		})
	}
}

func TestRuleBasedClassifier_ContextCancellation(t *testing.T) {
	classifier := NewRuleBasedClassifier()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := classifier.Classify(ctx, "What's my balance?", nil)
	if err == nil {
		t.Error("Classify() should return error for cancelled context")
	}
}

func TestHybridClassifier_Classify(t *testing.T) {
	mockBackend := NewMockLLMBackend()
	classifier := NewHybridClassifier(mockBackend, 0.8)
	ctx := context.Background()

	tests := []struct {
		name          string
		message       string
		wantIntent    Intent
		minConfidence float32
	}{
		{
			name:          "high confidence rule match",
			message:       "What's my balance?",
			wantIntent:    IntentQueryBalance,
			minConfidence: 0.7,
		},
		{
			name:          "lower confidence falls back to LLM",
			message:       "tokens",
			wantIntent:    IntentQueryBalance,
			minConfidence: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := classifier.Classify(ctx, tt.message, nil)
			if err != nil {
				t.Errorf("Classify() error = %v", err)
				return
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("Classify() confidence = %v, want >= %v", result.Confidence, tt.minConfidence)
			}
		})
	}
}

func TestLLMClassifier_NoBackend(t *testing.T) {
	classifier := NewLLMClassifier(nil)
	ctx := context.Background()

	_, err := classifier.Classify(ctx, "Hello", nil)
	if err != ErrLLMBackendNotConfigured {
		t.Errorf("Classify() error = %v, want %v", err, ErrLLMBackendNotConfigured)
	}
}

func TestIntent_String(t *testing.T) {
	tests := []struct {
		intent Intent
		want   string
	}{
		{IntentQueryBalance, "query_balance"},
		{IntentFindOfferings, "find_offerings"},
		{IntentGetHelp, "get_help"},
		{IntentCheckOrder, "check_order"},
		{IntentStaking, "staking"},
		{IntentDeployment, "deployment"},
		{IntentIdentity, "identity"},
		{IntentGeneralChat, "general_chat"},
		{IntentUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.intent.String(); got != tt.want {
				t.Errorf("Intent.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
