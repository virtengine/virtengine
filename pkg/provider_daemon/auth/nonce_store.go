package auth

import (
	"sync"
	"time"
)

// NonceStore tracks recently seen nonces.
type NonceStore interface {
	HasSeen(nonce string) bool
	MarkSeen(nonce string, expiry time.Time)
}

// InMemoryNonceStore tracks nonces in memory with expiry.
type InMemoryNonceStore struct {
	mu     sync.RWMutex
	nonces map[string]time.Time
	stopCh chan struct{}
}

// NewInMemoryNonceStore creates a nonce store with cleanup goroutine.
func NewInMemoryNonceStore() *InMemoryNonceStore {
	store := &InMemoryNonceStore{
		nonces: make(map[string]time.Time),
		stopCh: make(chan struct{}),
	}
	go store.cleanupLoop()
	return store
}

// HasSeen returns true if the nonce exists and has not expired.
func (s *InMemoryNonceStore) HasSeen(nonce string) bool {
	s.mu.RLock()
	expiry, exists := s.nonces[nonce]
	s.mu.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		s.mu.Lock()
		delete(s.nonces, nonce)
		s.mu.Unlock()
		return false
	}
	return true
}

// MarkSeen records the nonce with an expiry time.
func (s *InMemoryNonceStore) MarkSeen(nonce string, expiry time.Time) {
	s.mu.Lock()
	s.nonces[nonce] = expiry
	s.mu.Unlock()
}

// Stop stops the cleanup goroutine.
func (s *InMemoryNonceStore) Stop() {
	close(s.stopCh)
}

func (s *InMemoryNonceStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.stopCh:
			return
		}
	}
}

func (s *InMemoryNonceStore) cleanupExpired() {
	now := time.Now()
	s.mu.Lock()
	for nonce, expiry := range s.nonces {
		if now.After(expiry) {
			delete(s.nonces, nonce)
		}
	}
	s.mu.Unlock()
}
