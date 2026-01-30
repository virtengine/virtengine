// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Session Manager Implementation
// ============================================================================

// sessionManager implements the SessionManager interface
type sessionManager struct {
	config       Config
	sessions     map[string]*Session  // sessionID -> Session
	walletIndex  map[string][]string  // walletAddress -> sessionIDs
	assertionIDs map[string]time.Time // assertionID -> expiry (for replay detection)
	signingKey   []byte
	mu           sync.RWMutex
}

// newSessionManager creates a new session manager
func newSessionManager(config Config) SessionManager {
	// Generate signing key if not provided
	signingKey := []byte(config.SessionStorage.EncryptionKey)
	if len(signingKey) == 0 {
		signingKey = make([]byte, 32)
		_, _ = rand.Read(signingKey)
	}

	return &sessionManager{
		config:       config,
		sessions:     make(map[string]*Session),
		walletIndex:  make(map[string][]string),
		assertionIDs: make(map[string]time.Time),
		signingKey:   signingKey,
	}
}

// Create creates a new session
func (m *sessionManager) Create(ctx context.Context, assertion *SAMLAssertion, walletAddress string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check max sessions per wallet
	existing := m.walletIndex[walletAddress]
	if len(existing) >= m.config.SessionStorage.MaxSessions {
		// Remove oldest session
		if len(existing) > 0 {
			oldestID := existing[0]
			delete(m.sessions, oldestID)
			m.walletIndex[walletAddress] = existing[1:]
		}
	}

	// Generate session ID
	sessionID, err := m.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Get institution name from assertion
	institutionName := assertion.IssuerEntityID // Could look up display name

	now := time.Now()
	session := &Session{
		ID:              sessionID,
		WalletAddress:   walletAddress,
		InstitutionID:   assertion.IssuerEntityID,
		InstitutionName: institutionName,
		Attributes:      assertion.Attributes,
		AuthnInstant:    assertion.AuthnInstant,
		CreatedAt:       now,
		ExpiresAt:       now.Add(m.config.SessionDuration),
		Status:          SessionStatusActive,
		IsMFA:           assertion.IsMFA,
		SessionIndex:    assertion.SessionIndex,
		AssertionID:     assertion.ID,
	}

	// Store session
	m.sessions[sessionID] = session

	// Update wallet index
	m.walletIndex[walletAddress] = append(m.walletIndex[walletAddress], sessionID)

	return session, nil
}

// Get returns a session by ID
func (m *sessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// ValidateToken validates a session token and returns the session
func (m *sessionManager) ValidateToken(ctx context.Context, token string) (*Session, error) {
	// Decode token
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, ErrInvalidSessionToken
	}

	// Token format: sessionID + "." + timestamp + "." + signature
	var tokenData struct {
		SessionID string    `json:"sid"`
		ExpiresAt time.Time `json:"exp"`
		Signature string    `json:"sig"`
	}

	if err := json.Unmarshal(decoded, &tokenData); err != nil {
		return nil, ErrInvalidSessionToken
	}

	// Verify signature
	expectedSig := m.signToken(tokenData.SessionID, tokenData.ExpiresAt)
	if tokenData.Signature != expectedSig {
		return nil, ErrInvalidSessionToken
	}

	// Check token expiry
	if time.Now().After(tokenData.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Get session
	return m.Get(ctx, tokenData.SessionID)
}

// GenerateToken generates a session token
func (m *sessionManager) GenerateToken(session *Session) (string, error) {
	tokenData := struct {
		SessionID string    `json:"sid"`
		ExpiresAt time.Time `json:"exp"`
		Signature string    `json:"sig"`
	}{
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	}

	tokenData.Signature = m.signToken(session.ID, session.ExpiresAt)

	data, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

// Revoke revokes a session
func (m *sessionManager) Revoke(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	session.Status = SessionStatusRevoked

	// Remove from wallet index
	walletSessions := m.walletIndex[session.WalletAddress]
	for i, id := range walletSessions {
		if id == sessionID {
			m.walletIndex[session.WalletAddress] = append(walletSessions[:i], walletSessions[i+1:]...)
			break
		}
	}

	delete(m.sessions, sessionID)

	return nil
}

// RevokeAll revokes all sessions for a wallet
func (m *sessionManager) RevokeAll(ctx context.Context, walletAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionIDs, ok := m.walletIndex[walletAddress]
	if !ok {
		return nil
	}

	for _, sessionID := range sessionIDs {
		if session, ok := m.sessions[sessionID]; ok {
			session.Status = SessionStatusRevoked
			delete(m.sessions, sessionID)
		}
	}

	delete(m.walletIndex, walletAddress)

	return nil
}

// List lists sessions for a wallet
func (m *sessionManager) List(ctx context.Context, walletAddress string) ([]Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionIDs, ok := m.walletIndex[walletAddress]
	if !ok {
		return nil, nil
	}

	var sessions []Session
	for _, sessionID := range sessionIDs {
		if session, ok := m.sessions[sessionID]; ok {
			sessions = append(sessions, *session)
		}
	}

	return sessions, nil
}

// Cleanup removes expired sessions
func (m *sessionManager) Cleanup(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0

	// Clean up expired sessions
	for sessionID, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			session.Status = SessionStatusExpired
			delete(m.sessions, sessionID)

			// Remove from wallet index
			walletSessions := m.walletIndex[session.WalletAddress]
			for i, id := range walletSessions {
				if id == sessionID {
					m.walletIndex[session.WalletAddress] = append(walletSessions[:i], walletSessions[i+1:]...)
					break
				}
			}

			removed++
		}
	}

	// Clean up expired assertion IDs
	for assertionID, expiry := range m.assertionIDs {
		if now.After(expiry) {
			delete(m.assertionIDs, assertionID)
		}
	}

	return removed, nil
}

// GetStats returns session statistics
func (m *sessionManager) GetStats(ctx context.Context) (*SessionStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &SessionStats{
		TotalSessions: len(m.sessions),
		UniqueWallets: len(m.walletIndex),
		ByInstitution: make(map[string]int),
	}

	now := time.Now()
	for _, session := range m.sessions {
		if session.Status == SessionStatusActive && now.Before(session.ExpiresAt) {
			stats.ActiveSessions++
		} else {
			stats.ExpiredSessions++
		}

		stats.ByInstitution[session.InstitutionID]++
	}

	return stats, nil
}

// TrackAssertionID tracks an assertion ID for replay detection
func (m *sessionManager) TrackAssertionID(ctx context.Context, assertionID string, expiry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.assertionIDs[assertionID] = expiry
	return nil
}

// IsAssertionReplayed checks if an assertion ID has been seen
func (m *sessionManager) IsAssertionReplayed(ctx context.Context, assertionID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	expiry, exists := m.assertionIDs[assertionID]
	if !exists {
		return false, nil
	}

	// If the entry has expired, it's not a replay
	if time.Now().After(expiry) {
		return false, nil
	}

	return true, nil
}

// generateSessionID generates a random session ID
func (m *sessionManager) generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// signToken creates an HMAC signature for a token
func (m *sessionManager) signToken(sessionID string, expiresAt time.Time) string {
	data := fmt.Sprintf("%s:%d", sessionID, expiresAt.Unix())
	h := hmac.New(sha256.New, m.signingKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Session Token Types
// ============================================================================

// SessionToken represents a session token for API authentication
type SessionToken struct {
	// Token is the encoded token string
	Token string `json:"token"`

	// ExpiresAt is when the token expires
	ExpiresAt time.Time `json:"expires_at"`

	// SessionID is the session ID
	SessionID string `json:"session_id"`

	// WalletAddress is the associated wallet
	WalletAddress string `json:"wallet_address"`
}

// CreateSessionToken creates a session token from a session
func CreateSessionToken(session *Session, manager SessionManager) (*SessionToken, error) {
	token, err := manager.GenerateToken(session)
	if err != nil {
		return nil, err
	}

	return &SessionToken{
		Token:         token,
		ExpiresAt:     session.ExpiresAt,
		SessionID:     session.ID,
		WalletAddress: session.WalletAddress,
	}, nil
}
