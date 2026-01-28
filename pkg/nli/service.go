package nli

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// NLI Service Implementation
// ============================================================================

// nliService implements the Service interface
type nliService struct {
	config            Config
	classifier        Classifier
	responseGenerator *DefaultResponseGenerator
	queryExecutor     *DefaultQueryExecutor
	llmBackend        LLMBackend

	// Session management
	sessions     map[string]*sessionState
	sessionMu    sync.RWMutex
	sessionLimit int

	// Rate limiting
	rateLimiter map[string]*rateLimitEntry
	rateMu      sync.Mutex

	// Lifecycle
	closed   bool
	closedMu sync.RWMutex
}

// sessionState tracks conversation state for a session
type sessionState struct {
	history      []ChatMessage
	lastActivity time.Time
	userAddress  string
	context      *ChatContext
}

// rateLimitEntry tracks rate limiting for a session
type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

// NewService creates a new NLI service with the given configuration
func NewService(config Config) (Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create LLM backend
	backend, err := CreateLLMBackend(&config)
	if err != nil {
		return nil, err
	}

	// Create classifier based on configuration
	var classifier Classifier
	if config.LLMBackend == LLMBackendMock {
		// Use rule-based classifier for mock backend
		classifier = NewRuleBasedClassifier()
	} else {
		// Use hybrid classifier for real backends
		classifier = NewHybridClassifier(backend, config.MinConfidenceThreshold)
	}

	svc := &nliService{
		config:            config,
		classifier:        classifier,
		responseGenerator: NewDefaultResponseGenerator(config),
		queryExecutor:     NewDefaultQueryExecutor(),
		llmBackend:        backend,
		sessions:          make(map[string]*sessionState),
		sessionLimit:      1000,
		rateLimiter:       make(map[string]*rateLimitEntry),
	}

	return svc, nil
}

// Chat processes a chat request and returns a response
func (s *nliService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Check if service is closed
	s.closedMu.RLock()
	if s.closed {
		s.closedMu.RUnlock()
		return nil, ErrServiceClosed
	}
	s.closedMu.RUnlock()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Apply timeout
	if s.config.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.RequestTimeout)
		defer cancel()
	}

	// Check rate limit
	if !s.checkRateLimit(req.SessionID) {
		return nil, ErrRateLimited
	}

	// Get or create session
	session := s.getOrCreateSession(req.SessionID, req.UserAddress)

	// Update context with session history
	chatCtx := req.Context
	if chatCtx == nil {
		chatCtx = &ChatContext{}
	}
	chatCtx.History = session.history

	// Classify intent
	classification, err := s.classifier.Classify(ctx, req.Message, chatCtx)
	if err != nil {
		return nil, err
	}

	// Execute query if applicable
	var queryResult *QueryResult
	if s.config.EnableQueryExecution {
		queryResult, err = s.queryExecutor.Execute(ctx, classification.Intent, classification.Entities, req.UserAddress)
		if err != nil {
			// Log error but continue with response generation
			queryResult = nil
		}
	}

	// Generate response
	response, err := s.responseGenerator.Generate(ctx, classification.Intent, classification.Entities, queryResult, chatCtx)
	if err != nil {
		return nil, err
	}

	// Update session history
	s.updateSessionHistory(req.SessionID, req.Message, response)

	// Generate suggestions if enabled
	var suggestions []string
	if s.config.EnableSuggestions {
		suggestions = s.generateSuggestions(classification.Intent)
	}

	processingTime := time.Since(startTime)

	return &ChatResponse{
		Message:     response,
		Intent:      classification.Intent,
		Confidence:  classification.Confidence,
		SessionID:   req.SessionID,
		Data:        s.extractResponseData(queryResult),
		Suggestions: suggestions,
		Metadata: ResponseMetadata{
			ProcessingTime: processingTime,
			Timestamp:      time.Now(),
			Model:          string(s.config.LLMBackend),
		},
	}, nil
}

// ClassifyIntent classifies the intent of a message
func (s *nliService) ClassifyIntent(ctx context.Context, message string) (*ClassificationResult, error) {
	s.closedMu.RLock()
	if s.closed {
		s.closedMu.RUnlock()
		return nil, ErrServiceClosed
	}
	s.closedMu.RUnlock()

	return s.classifier.Classify(ctx, message, nil)
}

// GetSuggestions returns suggested actions or questions for a session
func (s *nliService) GetSuggestions(ctx context.Context, sessionID string) ([]string, error) {
	s.closedMu.RLock()
	if s.closed {
		s.closedMu.RUnlock()
		return nil, ErrServiceClosed
	}
	s.closedMu.RUnlock()

	// Get session
	s.sessionMu.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionMu.RUnlock()

	if !exists || len(session.history) == 0 {
		return []string{
			"What's my balance?",
			"Search marketplace offerings",
			"How do I stake tokens?",
			"Check my orders",
		}, nil
	}

	// Generate suggestions based on last interaction
	// In a real implementation, this would analyze conversation context
	return []string{
		"Tell me more",
		"Show other options",
		"Help me with something else",
	}, nil
}

// Close closes the service and releases resources
func (s *nliService) Close() error {
	s.closedMu.Lock()
	s.closed = true
	s.closedMu.Unlock()

	// Close LLM backend
	if s.llmBackend != nil {
		return s.llmBackend.Close()
	}

	return nil
}

// ============================================================================
// Session Management
// ============================================================================

// getOrCreateSession gets an existing session or creates a new one
func (s *nliService) getOrCreateSession(sessionID, userAddress string) *sessionState {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		session.lastActivity = time.Now()
		if userAddress != "" {
			session.userAddress = userAddress
		}
		return session
	}

	// Create new session
	session := &sessionState{
		history:      make([]ChatMessage, 0, s.config.MaxHistoryLength),
		lastActivity: time.Now(),
		userAddress:  userAddress,
	}

	// Enforce session limit by removing oldest sessions
	if len(s.sessions) >= s.sessionLimit {
		s.cleanOldSessions()
	}

	s.sessions[sessionID] = session
	return session
}

// updateSessionHistory adds messages to the session history
func (s *nliService) updateSessionHistory(sessionID, userMessage, assistantResponse string) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return
	}

	// Add user message
	session.history = append(session.history, ChatMessage{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	})

	// Add assistant response
	session.history = append(session.history, ChatMessage{
		Role:      "assistant",
		Content:   assistantResponse,
		Timestamp: time.Now(),
	})

	// Trim history if it exceeds max length
	if len(session.history) > s.config.MaxHistoryLength*2 {
		session.history = session.history[2:]
	}
}

// cleanOldSessions removes the oldest sessions when limit is reached
func (s *nliService) cleanOldSessions() {
	// Find and remove oldest 10% of sessions
	toRemove := len(s.sessions) / 10
	if toRemove < 1 {
		toRemove = 1
	}

	type sessionAge struct {
		id   string
		time time.Time
	}

	ages := make([]sessionAge, 0, len(s.sessions))
	for id, session := range s.sessions {
		ages = append(ages, sessionAge{id: id, time: session.lastActivity})
	}

	// Simple sort by age
	for i := range ages {
		for j := i + 1; j < len(ages); j++ {
			if ages[j].time.Before(ages[i].time) {
				ages[i], ages[j] = ages[j], ages[i]
			}
		}
	}

	// Remove oldest sessions
	for i := 0; i < toRemove && i < len(ages); i++ {
		delete(s.sessions, ages[i].id)
	}
}

// ============================================================================
// Rate Limiting
// ============================================================================

// checkRateLimit checks if the request should be rate limited
func (s *nliService) checkRateLimit(sessionID string) bool {
	if s.config.RateLimitRequests <= 0 {
		return true
	}

	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := time.Now()
	entry, exists := s.rateLimiter[sessionID]

	if !exists || now.After(entry.resetTime) {
		// Create or reset entry
		s.rateLimiter[sessionID] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(time.Minute),
		}
		return true
	}

	if entry.count >= s.config.RateLimitRequests {
		return false
	}

	entry.count++
	return true
}

// ============================================================================
// Helper Methods
// ============================================================================

// generateSuggestions generates follow-up suggestions based on intent
func (s *nliService) generateSuggestions(intent Intent) []string {
	switch intent {
	case IntentQueryBalance:
		return []string{
			"Show my transaction history",
			"Send tokens",
			"Stake my tokens",
		}
	case IntentFindOfferings:
		return []string{
			"Compare providers",
			"Show GPU offerings",
			"Filter by price",
		}
	case IntentCheckOrder:
		return []string{
			"Cancel order",
			"View deployment logs",
			"Extend lease",
		}
	case IntentStaking:
		return []string{
			"Show available validators",
			"View my delegations",
			"Claim rewards",
		}
	case IntentDeployment:
		return []string{
			"Show deployment templates",
			"View my deployments",
			"Check deployment logs",
		}
	case IntentIdentity:
		return []string{
			"Start verification",
			"Check verification status",
			"What documents do I need?",
		}
	default:
		return []string{
			"Check my balance",
			"Search marketplace",
			"View my orders",
		}
	}
}

// extractResponseData extracts structured data from query results
func (s *nliService) extractResponseData(result *QueryResult) map[string]interface{} {
	if result == nil {
		return nil
	}
	return map[string]interface{}{
		"query_type": result.QueryType,
		"success":    result.Success,
		"data":       result.Data,
	}
}

// SetQueryExecutor sets a custom query executor (for integration)
func (s *nliService) SetQueryExecutor(executor *DefaultQueryExecutor) {
	s.queryExecutor = executor
}

// GetQueryExecutor returns the query executor (for integration)
func (s *nliService) GetQueryExecutor() *DefaultQueryExecutor {
	return s.queryExecutor
}
