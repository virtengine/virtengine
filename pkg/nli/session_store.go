package nli

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// SessionStore defines the interface for session storage
type SessionStore interface {
	// Get retrieves a session by ID
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Set stores or updates a session
	Set(ctx context.Context, session *Session) error

	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error

	// Touch updates the session's last activity time
	Touch(ctx context.Context, sessionID string) error

	// Count returns the number of active sessions
	Count(ctx context.Context) (int64, error)

	// Close closes the session store
	Close() error
}

// Session represents a stored NLI session
type Session struct {
	ID           string        `json:"id"`
	History      []ChatMessage `json:"history"`
	LastActivity time.Time     `json:"last_activity"`
	UserAddress  string        `json:"user_address,omitempty"`
	Context      *ChatContext  `json:"context,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// SessionStoreConfig holds configuration for session storage
type SessionStoreConfig struct {
	// Backend specifies the session store backend ("memory" or "redis")
	Backend string `json:"backend"`

	// RedisURL is the Redis connection URL
	RedisURL string `json:"redis_url"`

	// RedisPrefix is the key prefix for session keys
	RedisPrefix string `json:"redis_prefix"`

	// SessionTTL is the TTL for session data
	SessionTTL time.Duration `json:"session_ttl"`

	// MaxSessions is the maximum number of sessions (for memory store)
	MaxSessions int `json:"max_sessions"`

	// MaxHistoryLength limits conversation history per session
	MaxHistoryLength int `json:"max_history_length"`
}

// DefaultSessionStoreConfig returns the default session store configuration
func DefaultSessionStoreConfig() SessionStoreConfig {
	return SessionStoreConfig{
		Backend:          "memory",
		RedisURL:         "redis://localhost:6379/0",
		RedisPrefix:      "virtengine:nli:session:",
		SessionTTL:       30 * time.Minute,
		MaxSessions:      10000,
		MaxHistoryLength: 20,
	}
}

// ============================================================================
// Redis Session Store
// ============================================================================

// RedisSessionStore implements SessionStore using Redis
type RedisSessionStore struct {
	client  *redis.Client
	config  SessionStoreConfig
	logger  zerolog.Logger
	metrics *sessionMetrics
}

// sessionMetrics tracks session store metrics
type sessionMetrics struct {
	mu         sync.RWMutex
	gets       uint64
	sets       uint64
	deletes    uint64
	hits       uint64
	misses     uint64
	expiredTTL uint64
}

// NewRedisSessionStore creates a new Redis-based session store
func NewRedisSessionStore(ctx context.Context, config SessionStoreConfig, logger zerolog.Logger) (*RedisSessionStore, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("nli: invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("nli: failed to connect to redis: %w", err)
	}

	store := &RedisSessionStore{
		client:  client,
		config:  config,
		logger:  logger.With().Str("component", "nli-session-store").Logger(),
		metrics: &sessionMetrics{},
	}

	logger.Info().
		Str("redis_url", config.RedisURL).
		Str("prefix", config.RedisPrefix).
		Dur("ttl", config.SessionTTL).
		Msg("redis session store initialized")

	return store, nil
}

// Get retrieves a session by ID
func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	s.metrics.mu.Lock()
	s.metrics.gets++
	s.metrics.mu.Unlock()

	key := s.formatKey(sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			s.metrics.mu.Lock()
			s.metrics.misses++
			s.metrics.mu.Unlock()
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("nli: redis get failed: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("nli: session unmarshal failed: %w", err)
	}

	s.metrics.mu.Lock()
	s.metrics.hits++
	s.metrics.mu.Unlock()

	return &session, nil
}

// Set stores or updates a session
func (s *RedisSessionStore) Set(ctx context.Context, session *Session) error {
	s.metrics.mu.Lock()
	s.metrics.sets++
	s.metrics.mu.Unlock()

	// Ensure session has required fields
	if session.ID == "" {
		return ErrMissingSessionID
	}

	// Trim history if needed
	if len(session.History) > s.config.MaxHistoryLength*2 {
		session.History = session.History[len(session.History)-s.config.MaxHistoryLength*2:]
	}

	// Update last activity
	session.LastActivity = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("nli: session marshal failed: %w", err)
	}

	key := s.formatKey(session.ID)
	if err := s.client.Set(ctx, key, data, s.config.SessionTTL).Err(); err != nil {
		return fmt.Errorf("nli: redis set failed: %w", err)
	}

	return nil
}

// Delete removes a session
func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	s.metrics.mu.Lock()
	s.metrics.deletes++
	s.metrics.mu.Unlock()

	key := s.formatKey(sessionID)
	return s.client.Del(ctx, key).Err()
}

// Touch updates the session's last activity time and refreshes TTL
func (s *RedisSessionStore) Touch(ctx context.Context, sessionID string) error {
	key := s.formatKey(sessionID)

	// Get, update, and set back
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("nli: session marshal failed: %w", err)
	}

	return s.client.Set(ctx, key, data, s.config.SessionTTL).Err()
}

// Count returns the number of active sessions
func (s *RedisSessionStore) Count(ctx context.Context) (int64, error) {
	pattern := s.config.RedisPrefix + "*"

	var count int64
	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return 0, fmt.Errorf("nli: redis scan failed: %w", err)
		}
		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// Close closes the Redis connection
func (s *RedisSessionStore) Close() error {
	return s.client.Close()
}

// GetMetrics returns session store metrics
func (s *RedisSessionStore) GetMetrics() SessionStoreMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return SessionStoreMetrics{
		Gets:    s.metrics.gets,
		Sets:    s.metrics.sets,
		Deletes: s.metrics.deletes,
		Hits:    s.metrics.hits,
		Misses:  s.metrics.misses,
	}
}

// formatKey formats a session key with the prefix
func (s *RedisSessionStore) formatKey(sessionID string) string {
	return s.config.RedisPrefix + sessionID
}

// ============================================================================
// In-Memory Session Store (fallback)
// ============================================================================

// InMemorySessionStore implements SessionStore using in-memory storage
type InMemorySessionStore struct {
	sessions   map[string]*Session
	mu         sync.RWMutex
	config     SessionStoreConfig
	logger     zerolog.Logger
	metrics    *sessionMetrics
	cleanupTTL time.Duration
	stopCh     chan struct{}
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore(config SessionStoreConfig, logger zerolog.Logger) *InMemorySessionStore {
	store := &InMemorySessionStore{
		sessions: make(map[string]*Session),
		config:   config,
		logger:   logger.With().Str("component", "nli-memory-session-store").Logger(),
		metrics:  &sessionMetrics{},
		stopCh:   make(chan struct{}),
	}

	// Start background cleanup
	go store.cleanupLoop()

	logger.Info().
		Int("max_sessions", config.MaxSessions).
		Dur("ttl", config.SessionTTL).
		Msg("in-memory session store initialized")

	return store
}

// Get retrieves a session by ID
func (s *InMemorySessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	s.metrics.mu.Lock()
	s.metrics.gets++
	s.metrics.mu.Unlock()

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		s.metrics.mu.Lock()
		s.metrics.misses++
		s.metrics.mu.Unlock()
		return nil, ErrSessionNotFound
	}

	// Check if expired
	if time.Since(session.LastActivity) > s.config.SessionTTL {
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()

		s.metrics.mu.Lock()
		s.metrics.expiredTTL++
		s.metrics.misses++
		s.metrics.mu.Unlock()
		return nil, ErrSessionNotFound
	}

	s.metrics.mu.Lock()
	s.metrics.hits++
	s.metrics.mu.Unlock()

	// Return a copy to prevent mutation
	sessionCopy := *session
	historyCopy := make([]ChatMessage, len(session.History))
	copy(historyCopy, session.History)
	sessionCopy.History = historyCopy

	return &sessionCopy, nil
}

// Set stores or updates a session
func (s *InMemorySessionStore) Set(ctx context.Context, session *Session) error {
	s.metrics.mu.Lock()
	s.metrics.sets++
	s.metrics.mu.Unlock()

	if session.ID == "" {
		return ErrMissingSessionID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Enforce max sessions limit
	if _, exists := s.sessions[session.ID]; !exists && len(s.sessions) >= s.config.MaxSessions {
		s.evictOldest()
	}

	// Trim history if needed
	if len(session.History) > s.config.MaxHistoryLength*2 {
		session.History = session.History[len(session.History)-s.config.MaxHistoryLength*2:]
	}

	// Make a copy and update timestamp
	sessionCopy := *session
	sessionCopy.LastActivity = time.Now()
	historyCopy := make([]ChatMessage, len(session.History))
	copy(historyCopy, session.History)
	sessionCopy.History = historyCopy

	s.sessions[session.ID] = &sessionCopy

	return nil
}

// Delete removes a session
func (s *InMemorySessionStore) Delete(ctx context.Context, sessionID string) error {
	s.metrics.mu.Lock()
	s.metrics.deletes++
	s.metrics.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// Touch updates the session's last activity time
func (s *InMemorySessionStore) Touch(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	session.LastActivity = time.Now()
	return nil
}

// Count returns the number of active sessions
func (s *InMemorySessionStore) Count(ctx context.Context) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.sessions)), nil
}

// Close stops the cleanup loop
func (s *InMemorySessionStore) Close() error {
	close(s.stopCh)
	return nil
}

// GetMetrics returns session store metrics
func (s *InMemorySessionStore) GetMetrics() SessionStoreMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return SessionStoreMetrics{
		Gets:    s.metrics.gets,
		Sets:    s.metrics.sets,
		Deletes: s.metrics.deletes,
		Hits:    s.metrics.hits,
		Misses:  s.metrics.misses,
	}
}

// evictOldest removes the oldest 10% of sessions
func (s *InMemorySessionStore) evictOldest() {
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
		ages = append(ages, sessionAge{id: id, time: session.LastActivity})
	}

	// Sort by age (oldest first)
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

// cleanupLoop periodically removes expired sessions
func (s *InMemorySessionStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes expired sessions
func (s *InMemorySessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.LastActivity) > s.config.SessionTTL {
			delete(s.sessions, id)
			s.metrics.mu.Lock()
			s.metrics.expiredTTL++
			s.metrics.mu.Unlock()
		}
	}
}

// ============================================================================
// Session Store Metrics
// ============================================================================

// SessionStoreMetrics contains session store metrics
type SessionStoreMetrics struct {
	Gets    uint64 `json:"gets"`
	Sets    uint64 `json:"sets"`
	Deletes uint64 `json:"deletes"`
	Hits    uint64 `json:"hits"`
	Misses  uint64 `json:"misses"`
}

// NewSessionStore creates a session store based on configuration
func NewSessionStore(ctx context.Context, config SessionStoreConfig, logger zerolog.Logger) (SessionStore, error) {
	switch config.Backend {
	case "redis":
		return NewRedisSessionStore(ctx, config, logger)
	case "memory", "":
		return NewInMemorySessionStore(config, logger), nil
	default:
		return nil, fmt.Errorf("nli: unknown session store backend: %s", config.Backend)
	}
}
