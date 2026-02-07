package auth

import (
	"sync"
	"time"
)

// NonceStore tracks seen nonces to prevent replay attacks.
type NonceStore interface {
	HasSeen(key string) bool
	MarkSeen(key string, expiry time.Time)
}

// InMemoryNonceStore stores nonces in memory with periodic cleanup.
type InMemoryNonceStore struct {
	mu     sync.RWMutex
	nonces map[string]time.Time
	stopCh chan struct{}
}

// NewInMemoryNonceStore creates a nonce store with background cleanup.
func NewInMemoryNonceStore() *InMemoryNonceStore {
	store := &InMemoryNonceStore{
		nonces: make(map[string]time.Time),
		stopCh: make(chan struct{}),
	}
	go store.cleanup()
	return store
}

// HasSeen returns true if the nonce key is already recorded.
func (s *InMemoryNonceStore) HasSeen(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, ok := s.nonces[key]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		return false
	}
	return true
}

// MarkSeen records a nonce key until expiry.
func (s *InMemoryNonceStore) MarkSeen(key string, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nonces[key] = expiry
}

// Stop stops background cleanup.
func (s *InMemoryNonceStore) Stop() {
	close(s.stopCh)
}

func (s *InMemoryNonceStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for key, expiry := range s.nonces {
				if now.After(expiry) {
					delete(s.nonces, key)
				}
			}
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}
