// Package nli provides a Natural Language Interface for AI-powered chat assistance.
//
// This package implements VE-904 from the VirtEngine roadmap, providing a modular
// NLI service that supports:
//   - Natural language query processing
//   - Intent classification (balance queries, marketplace search, help requests, etc.)
//   - Context-aware response generation
//   - Integration with blockchain state queries
//   - Multiple LLM backend support
//
// # Architecture
//
// The NLI package is designed with the following components:
//
//   - Service: Main entry point implementing the NLI interface
//   - Classifier: Intent classification system for routing queries
//   - ResponseGenerator: Context-aware response generation
//   - QueryExecutor: Blockchain state query integration
//   - LLMBackend: Abstraction for LLM providers (OpenAI, local models, etc.)
//
// # Usage
//
// Basic usage:
//
//	config := nli.DefaultConfig()
//	service, err := nli.NewService(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	req := &nli.ChatRequest{
//	    Message:   "What is my balance?",
//	    SessionID: "session-123",
//	    UserAddress: "ve1...",
//	}
//
//	resp, err := service.Chat(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Message)
//
// # LLM Backend Configuration
//
// The package supports multiple LLM backends:
//
//	// OpenAI backend
//	config.LLMBackend = nli.LLMBackendOpenAI
//	config.OpenAIConfig = &nli.OpenAIConfig{
//	    APIKey: os.Getenv("OPENAI_API_KEY"),
//	    Model:  "gpt-4",
//	}
//
//	// Local/mock backend for testing
//	config.LLMBackend = nli.LLMBackendLocal
//
// # Intent Classification
//
// The classifier recognizes the following intents:
//   - QueryBalance: "What is my balance?", "Show me my tokens"
//   - FindOfferings: "Find GPU providers", "Search marketplace"
//   - GetHelp: "How do I stake?", "Help with deployment"
//   - CheckOrder: "Order status", "Track my order"
//   - GetProviderInfo: "Tell me about provider X"
//   - GeneralChat: Fallback for general conversation
//
// VE-904: Natural Language Interface: AI chat
package nli
