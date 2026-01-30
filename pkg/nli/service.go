package nli

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/virtengine/virtengine/pkg/ratelimit"
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
	logger            zerolog.Logger

	// Session management (distributed via SessionStore)
	sessionStore SessionStore

	// Legacy session management (fallback when SessionStore is nil)
	sessions     map[string]*sessionState
	sessionMu    sync.RWMutex
	sessionLimit int

	// Rate limiting (distributed via pkg/ratelimit)
	distributedRateLimiter ratelimit.RateLimiter

	// Legacy rate limiting (fallback when distributed is disabled)
	rateLimiter map[string]*rateLimitEntry
	rateMu      sync.Mutex

	// Metrics
	metrics          *NLIMetrics
	metricsCollector *MetricsCollector

	// Lifecycle
	closed   bool
	closedMu sync.RWMutex
}

// sessionState tracks conversation state for a session (legacy in-memory)
type sessionState struct {
	history      []ChatMessage
	lastActivity time.Time
	userAddress  string
	context      *ChatContext
}

// rateLimitEntry tracks rate limiting for a session (legacy in-memory)
type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

// ServiceOption is a functional option for configuring the service
type ServiceOption func(*nliService) error

// WithLogger sets a custom logger
func WithLogger(logger zerolog.Logger) ServiceOption {
	return func(s *nliService) error {
		s.logger = logger
		return nil
	}
}

// WithSessionStore sets a custom session store
func WithSessionStore(store SessionStore) ServiceOption {
	return func(s *nliService) error {
		s.sessionStore = store
		return nil
	}
}

// WithDistributedRateLimiter sets a distributed rate limiter
func WithDistributedRateLimiter(limiter ratelimit.RateLimiter) ServiceOption {
	return func(s *nliService) error {
		s.distributedRateLimiter = limiter
		return nil
	}
}

// NewService creates a new NLI service with the given configuration
func NewService(config Config, opts ...ServiceOption) (Service, error) {
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
		logger:            zerolog.Nop(),
		sessions:          make(map[string]*sessionState),
		sessionLimit:      1000,
		rateLimiter:       make(map[string]*rateLimitEntry),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}

	// Initialize metrics
	svc.metrics = NewNLIMetrics(config.MetricsNamespace, svc.logger)

	// Start metrics collector if we have a session store
	if svc.sessionStore != nil {
		svc.metricsCollector = NewMetricsCollector(svc.metrics, svc.sessionStore, svc.logger)
		go svc.metricsCollector.Start(context.Background(), time.Minute)
	}

	return svc, nil
}

// NewServiceWithRedis creates a new NLI service with Redis-backed session store and rate limiter
func NewServiceWithRedis(ctx context.Context, config Config, logger zerolog.Logger) (Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create session store
	config.SessionStore.MaxHistoryLength = config.MaxHistoryLength
	sessionStore, err := NewSessionStore(ctx, config.SessionStore, logger)
	if err != nil {
		return nil, err
	}

	// Create rate limiter if enabled
	var rateLimiter ratelimit.RateLimiter
	if config.DistributedRateLimiter.Enabled {
		rateLimitConfig := ratelimit.RateLimitConfig{
			RedisURL:    config.DistributedRateLimiter.RedisURL,
			RedisPrefix: config.DistributedRateLimiter.RedisPrefix,
			Enabled:     true,
			UserLimits: ratelimit.LimitRules{
				RequestsPerSecond: config.DistributedRateLimiter.RequestsPerSecond,
				RequestsPerMinute: config.DistributedRateLimiter.RequestsPerMinute,
				BurstSize:         config.DistributedRateLimiter.BurstSize,
			},
		}
		rateLimiter, err = ratelimit.NewRedisRateLimiter(ctx, rateLimitConfig, logger)
		if err != nil {
			sessionStore.Close()
			return nil, err
		}
	}

	return NewService(config,
		WithLogger(logger),
		WithSessionStore(sessionStore),
		WithDistributedRateLimiter(rateLimiter),
	)
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
		s.metrics.RecordError("validation")
		return nil, err
	}

	// Apply timeout
	if s.config.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.RequestTimeout)
		defer cancel()
	}

	// Check rate limit
	allowed, err := s.checkRateLimitCtx(ctx, req.SessionID)
	if err != nil {
		s.logger.Warn().Err(err).Str("session_id", req.SessionID).Msg("rate limit check failed")
	}
	if !allowed {
		s.metrics.RecordRateLimitHit()
		s.metrics.RecordError("ratelimit")
		return nil, ErrRateLimited
	}
	s.metrics.RecordRateLimitPass()

	// Get or create session
	session, err := s.getOrCreateSessionCtx(ctx, req.SessionID, req.UserAddress)
	if err != nil {
		s.logger.Warn().Err(err).Str("session_id", req.SessionID).Msg("session retrieval failed, using empty history")
		session = &Session{
			ID:           req.SessionID,
			UserAddress:  req.UserAddress,
			History:      []ChatMessage{},
			LastActivity: time.Now(),
			CreatedAt:    time.Now(),
		}
	}

	// Update context with session history
	chatCtx := req.Context
	if chatCtx == nil {
		chatCtx = &ChatContext{}
	}
	chatCtx.History = session.History

	// Classify intent
	classification, err := s.classifier.Classify(ctx, req.Message, chatCtx)
	if err != nil {
		s.metrics.RecordError("classification")
		s.metrics.RecordRequest("error", "unknown", time.Since(startTime))
		return nil, err
	}
	s.metrics.RecordIntentClassified(string(classification.Intent))

	// Execute query if applicable
	var queryResult *QueryResult
	if s.config.EnableQueryExecution {
		queryResult, err = s.queryExecutor.Execute(ctx, classification.Intent, classification.Entities, req.UserAddress)
		if err != nil {
			// Log error but continue with response generation
			s.logger.Debug().Err(err).Msg("query execution failed")
			queryResult = nil
		}
	}

	// Generate response
	response, err := s.responseGenerator.Generate(ctx, classification.Intent, classification.Entities, queryResult, chatCtx)
	if err != nil {
		s.metrics.RecordError("generation")
		s.metrics.RecordRequest("error", string(classification.Intent), time.Since(startTime))
		return nil, err
	}

	// Update session history
	if err := s.updateSessionHistoryCtx(ctx, req.SessionID, req.Message, response); err != nil {
		s.logger.Warn().Err(err).Str("session_id", req.SessionID).Msg("failed to update session history")
	}

	// Generate suggestions if enabled
	var suggestions []string
	if s.config.EnableSuggestions {
		suggestions = s.generateSuggestions(classification.Intent)
	}

	processingTime := time.Since(startTime)
	s.metrics.RecordRequest("success", string(classification.Intent), processingTime)

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

	defaultSuggestions := []string{
		"What's my balance?",
		"Search marketplace offerings",
		"How do I stake tokens?",
		"Check my orders",
	}

	// Try session store first
	if s.sessionStore != nil {
		session, err := s.sessionStore.Get(ctx, sessionID)
		if err != nil || len(session.History) == 0 {
			return defaultSuggestions, nil
		}

		return []string{
			"Tell me more",
			"Show other options",
			"Help me with something else",
		}, nil
	}

	// Fall back to legacy session management
	s.sessionMu.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionMu.RUnlock()

	if !exists || len(session.history) == 0 {
		return defaultSuggestions, nil
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

	var errs []error

	// Stop metrics collector
	if s.metricsCollector != nil {
		s.metricsCollector.Stop()
	}

	// Close session store
	if s.sessionStore != nil {
		if err := s.sessionStore.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close distributed rate limiter
	if s.distributedRateLimiter != nil {
		if err := s.distributedRateLimiter.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close LLM backend
	if s.llmBackend != nil {
		if err := s.llmBackend.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
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

// checkRateLimitCtx checks rate limit using distributed limiter if available
func (s *nliService) checkRateLimitCtx(ctx context.Context, sessionID string) (bool, error) {
	// Use distributed rate limiter if available
	if s.distributedRateLimiter != nil {
		allowed, result, err := s.distributedRateLimiter.Allow(ctx, sessionID, ratelimit.LimitTypeUser)
		if err != nil {
			// Fall back to legacy rate limiting
			s.logger.Warn().Err(err).Msg("distributed rate limiter failed, using legacy")
			return s.checkRateLimit(sessionID), nil
		}
		if !allowed && result != nil {
			s.logger.Debug().
				Str("session_id", sessionID).
				Int("retry_after", result.RetryAfter).
				Msg("rate limited by distributed limiter")
		}
		return allowed, nil
	}

	// Fall back to legacy rate limiting
	return s.checkRateLimit(sessionID), nil
}

// getOrCreateSessionCtx gets or creates a session using the session store
func (s *nliService) getOrCreateSessionCtx(ctx context.Context, sessionID, userAddress string) (*Session, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordSessionOperation("get", time.Since(start))
	}()

	// Use session store if available
	if s.sessionStore != nil {
		session, err := s.sessionStore.Get(ctx, sessionID)
		if err == nil {
			// Session found, update user address if provided
			if userAddress != "" && session.UserAddress != userAddress {
				session.UserAddress = userAddress
				_ = s.sessionStore.Set(ctx, session) // Best effort update
			}
			return session, nil
		}

		// Create new session
		session = &Session{
			ID:           sessionID,
			History:      make([]ChatMessage, 0, s.config.MaxHistoryLength),
			LastActivity: time.Now(),
			UserAddress:  userAddress,
			CreatedAt:    time.Now(),
		}

		if err := s.sessionStore.Set(ctx, session); err != nil {
			return nil, err
		}

		return session, nil
	}

	// Fall back to legacy session management
	legacySession := s.getOrCreateSession(sessionID, userAddress)
	return &Session{
		ID:           sessionID,
		History:      legacySession.history,
		LastActivity: legacySession.lastActivity,
		UserAddress:  legacySession.userAddress,
		Context:      legacySession.context,
	}, nil
}

// updateSessionHistoryCtx updates session history using the session store
func (s *nliService) updateSessionHistoryCtx(ctx context.Context, sessionID, userMessage, assistantResponse string) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordSessionOperation("set", time.Since(start))
	}()

	// Use session store if available
	if s.sessionStore != nil {
		session, err := s.sessionStore.Get(ctx, sessionID)
		if err != nil {
			// Session not found, create a new one
			session = &Session{
				ID:           sessionID,
				History:      make([]ChatMessage, 0, s.config.MaxHistoryLength),
				LastActivity: time.Now(),
				CreatedAt:    time.Now(),
			}
		}

		// Add user message
		session.History = append(session.History, ChatMessage{
			Role:      "user",
			Content:   userMessage,
			Timestamp: time.Now(),
		})

		// Add assistant response
		session.History = append(session.History, ChatMessage{
			Role:      "assistant",
			Content:   assistantResponse,
			Timestamp: time.Now(),
		})

		return s.sessionStore.Set(ctx, session)
	}

	// Fall back to legacy session management
	s.updateSessionHistory(sessionID, userMessage, assistantResponse)
	return nil
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
