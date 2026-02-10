package pkcs11

import (
	"log/slog"
	"sync"
)

// SessionState tracks the state of a PKCS#11 session.
type SessionState int

const (
	// SessionClosed indicates no active session.
	SessionClosed SessionState = iota

	// SessionOpen indicates an active, authenticated session.
	SessionOpen

	// SessionError indicates the session encountered a fatal error.
	SessionError
)

// SessionPool manages a pool of PKCS#11 sessions for concurrent access.
type SessionPool struct {
	provider *Provider
	logger   *slog.Logger
	mu       sync.Mutex
	state    SessionState
}

// NewSessionPool creates a session pool for the given provider.
func NewSessionPool(provider *Provider, logger *slog.Logger) *SessionPool {
	return &SessionPool{
		provider: provider,
		logger:   logger,
		state:    SessionClosed,
	}
}

// Open opens a new session (or reuses the provider's existing connection).
func (sp *SessionPool) Open() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.state == SessionOpen {
		return nil
	}

	sp.state = SessionOpen
	sp.logger.Debug("PKCS#11 session opened")
	return nil
}

// Close closes the active session.
func (sp *SessionPool) Close() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.state = SessionClosed
	sp.logger.Debug("PKCS#11 session closed")
	return nil
}

// State returns the current session state.
func (sp *SessionPool) State() SessionState {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.state
}
