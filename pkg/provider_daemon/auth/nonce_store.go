package auth

import (
	"sync"
	"time"
)

// NonceStore tracks used nonces to prevent replay attacks.
type NonceStore interface {
	HasSeen(nonce string) bool
	MarkSeen(nonce string, expiry time.Time)
}

// InMemoryNonceStore keeps nonces in memory with expirations.
type InMemoryNonceStore struct {
	mu     sync.RWMutex
	nonces map[string]time.Time
}

func NewInMemoryNonceStore() *InMemoryNonceStore {
	store := &InMemoryNonceStore{
		nonces: make(map[string]time.Time),
	}
	go store.cleanup()
	return store
}

func (s *InMemoryNonceStore) HasSeen(nonce string) bool {
	if nonce == "" {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, exists := s.nonces[nonce]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(s.nonces, nonce)
		return false
	}
	return true
}

func (s *InMemoryNonceStore) MarkSeen(nonce string, expiry time.Time) {
	if nonce == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nonces[nonce] = expiry
}

func (s *InMemoryNonceStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for nonce, expiry := range s.nonces {
			if now.After(expiry) {
				delete(s.nonces, nonce)
			}
		}
		s.mu.Unlock()
	}
}
