// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Integrate Open OnDemand for web-based HPC access.
package ood_adapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Common error format strings
const (
	errFmtWrapped           = "%w: %v"
	errFmtFileBrowserDisabled = "%w: file browser is disabled"
)

// OODAdapter implements the Open OnDemand integration adapter
type OODAdapter struct {
	config       OODConfig
	client       OODClient
	authProvider VEIDAuthProvider
	signer       SessionSigner
	mu           sync.RWMutex
	sessions     map[string]*OODSession // OOD session ID -> session
	userSessions map[string][]string    // VEID address -> session IDs
	running      bool
	stopCh       chan struct{}
}

// SessionSigner signs session status updates
type SessionSigner interface {
	// Sign signs data and returns the signature
	Sign(data []byte) ([]byte, error)

	// Verify verifies a signature
	Verify(data []byte, signature []byte) bool

	// GetProviderAddress returns the provider address
	GetProviderAddress() string
}

// OnChainSessionReporter reports session status to the blockchain
type OnChainSessionReporter interface {
	// ReportSessionStatus reports session status on-chain
	ReportSessionStatus(ctx context.Context, report *SessionStatusReport) error
}

// SessionStatusReport contains session status for on-chain reporting
type SessionStatusReport struct {
	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// VirtEngineSessionID is the VirtEngine session ID
	VirtEngineSessionID string `json:"virtengine_session_id"`

	// OODSessionID is the OOD session ID
	OODSessionID string `json:"ood_session_id"`

	// AppType is the app type
	AppType InteractiveAppType `json:"app_type"`

	// State is the session state
	State SessionState `json:"state"`

	// VEIDAddress is the user's VEID address
	VEIDAddress string `json:"veid_address"`

	// StatusMessage is the status message
	StatusMessage string `json:"status_message,omitempty"`

	// UsageMetrics are the usage metrics
	UsageMetrics *SessionUsageMetrics `json:"usage_metrics,omitempty"`

	// Timestamp is when the report was created
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature
	Signature string `json:"signature"`
}

// SessionUsageMetrics contains session usage metrics
type SessionUsageMetrics struct {
	// WallClockSeconds is wall clock time
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// CPUTimeSeconds is CPU time used
	CPUTimeSeconds int64 `json:"cpu_time_seconds"`

	// MemoryBytesAvg is average memory usage
	MemoryBytesAvg int64 `json:"memory_bytes_avg"`

	// GPUSeconds is GPU time used
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`
}

// Hash generates a hash for signing
func (r *SessionStatusReport) Hash() []byte {
	data := struct {
		ProviderAddress     string `json:"provider_address"`
		VirtEngineSessionID string `json:"virtengine_session_id"`
		OODSessionID        string `json:"ood_session_id"`
		AppType             string `json:"app_type"`
		State               string `json:"state"`
		VEIDAddress         string `json:"veid_address"`
		Timestamp           int64  `json:"timestamp"`
	}{
		ProviderAddress:     r.ProviderAddress,
		VirtEngineSessionID: r.VirtEngineSessionID,
		OODSessionID:        r.OODSessionID,
		AppType:             string(r.AppType),
		State:               string(r.State),
		VEIDAddress:         r.VEIDAddress,
		Timestamp:           r.Timestamp.Unix(),
	}
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// NewOODAdapter creates a new Open OnDemand adapter
func NewOODAdapter(config OODConfig, client OODClient, authProvider VEIDAuthProvider, signer SessionSigner) *OODAdapter {
	return &OODAdapter{
		config:       config,
		client:       client,
		authProvider: authProvider,
		signer:       signer,
		sessions:     make(map[string]*OODSession),
		userSessions: make(map[string][]string),
		stopCh:       make(chan struct{}),
	}
}

// Start starts the OOD adapter
func (a *OODAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.mu.Unlock()

	// Connect to Open OnDemand
	if err := a.client.Connect(ctx); err != nil {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
		return fmt.Errorf("failed to connect to Open OnDemand: %w", err)
	}

	// Start session polling
	go a.pollSessions()

	return nil
}

// Stop stops the OOD adapter
func (a *OODAdapter) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = false
	close(a.stopCh)
	a.mu.Unlock()

	return a.client.Disconnect()
}

// IsRunning checks if the adapter is running
func (a *OODAdapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetConfig returns the adapter configuration
func (a *OODAdapter) GetConfig() OODConfig {
	return a.config
}

// AuthenticateUser authenticates a user via VEID SSO
func (a *OODAdapter) AuthenticateUser(ctx context.Context, veidAddress string, token *VEIDToken) error {
	if !a.IsRunning() {
		return ErrOODNotConnected
	}

	if token == nil || !token.IsValid() {
		return ErrInvalidToken
	}

	// Validate token with VEID
	validatedToken, err := a.authProvider.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return fmt.Errorf(errFmtWrapped, ErrAuthenticationFailed, err)
	}

	// Ensure token belongs to the claimed VEID address
	if validatedToken.VEIDAddress != veidAddress {
		return fmt.Errorf("%w: token does not match VEID address", ErrAuthenticationFailed)
	}

	// Authenticate with Open OnDemand
	if err := a.client.Authenticate(ctx, veidAddress, validatedToken); err != nil {
		return fmt.Errorf(errFmtWrapped, ErrAuthenticationFailed, err)
	}

	return nil
}

// RefreshUserToken refreshes a user's VEID token
func (a *OODAdapter) RefreshUserToken(ctx context.Context, refreshToken string) (*VEIDToken, error) {
	// Note: refreshToken is sensitive, never log it
	return a.authProvider.RefreshToken(ctx, refreshToken)
}

// GetAuthorizationURL gets the OIDC authorization URL for VEID SSO
func (a *OODAdapter) GetAuthorizationURL(state string, redirectURI string) string {
	return a.authProvider.GetAuthorizationURL(state, redirectURI)
}

// ExchangeAuthCode exchanges an authorization code for tokens
func (a *OODAdapter) ExchangeAuthCode(ctx context.Context, code string, redirectURI string) (*VEIDToken, error) {
	// Note: code is sensitive, never log it
	return a.authProvider.ExchangeCodeForToken(ctx, code, redirectURI)
}

// ListAvailableApps lists available interactive apps
func (a *OODAdapter) ListAvailableApps(ctx context.Context) ([]InteractiveApp, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	return a.client.ListApps(ctx)
}

// LaunchInteractiveApp launches an interactive app session
func (a *OODAdapter) LaunchInteractiveApp(ctx context.Context, virtEngineSessionID string, veidAddress string, spec *InteractiveAppSpec) (*OODSession, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid app spec: %w", err)
	}

	// Check if app is available
	apps, err := a.client.ListApps(ctx)
	if err != nil {
		return nil, err
	}

	appAvailable := false
	for _, app := range apps {
		if app.Type == spec.AppType && app.Available {
			appAvailable = true
			break
		}
	}
	if !appAvailable {
		return nil, ErrAppNotAvailable
	}

	// Launch the app
	session, err := a.client.LaunchApp(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf(errFmtWrapped, ErrSessionCreationFailed, err)
	}

	// Set VirtEngine-specific fields
	session.VirtEngineSessionID = virtEngineSessionID
	session.VEIDAddress = veidAddress
	session.CreatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(time.Duration(spec.Resources.Hours) * time.Hour)

	// Store session
	a.mu.Lock()
	a.sessions[session.SessionID] = session
	a.userSessions[veidAddress] = append(a.userSessions[veidAddress], session.SessionID)
	a.mu.Unlock()

	return session, nil
}

// GetSession gets a session by ID
func (a *OODAdapter) GetSession(ctx context.Context, sessionID string) (*OODSession, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	a.mu.RLock()
	session, exists := a.sessions[sessionID]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	// Get latest status from OOD
	updatedSession, err := a.client.GetSession(ctx, sessionID)
	if err != nil {
		// Return cached session if query fails
		return session, nil
	}

	// Preserve VirtEngine-specific fields
	updatedSession.VirtEngineSessionID = session.VirtEngineSessionID
	updatedSession.VEIDAddress = session.VEIDAddress
	updatedSession.ExpiresAt = session.ExpiresAt

	// Update cache
	a.mu.Lock()
	a.sessions[sessionID] = updatedSession
	a.mu.Unlock()

	return updatedSession, nil
}

// TerminateSession terminates a session
func (a *OODAdapter) TerminateSession(ctx context.Context, sessionID string) error {
	if !a.IsRunning() {
		return ErrOODNotConnected
	}

	a.mu.RLock()
	session, exists := a.sessions[sessionID]
	a.mu.RUnlock()

	if !exists {
		return ErrSessionNotFound
	}

	if err := a.client.TerminateSession(ctx, sessionID); err != nil {
		return err
	}

	// Update local state
	a.mu.Lock()
	session.State = SessionStateCancelled
	session.StatusMessage = "Terminated by user"
	now := time.Now()
	session.EndedAt = &now
	a.sessions[sessionID] = session
	a.mu.Unlock()

	return nil
}

// GetUserSessions gets all sessions for a user
func (a *OODAdapter) GetUserSessions(ctx context.Context, veidAddress string) ([]*OODSession, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	a.mu.RLock()
	sessionIDs := a.userSessions[veidAddress]
	sessions := make([]*OODSession, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		if session, exists := a.sessions[id]; exists {
			sessions = append(sessions, session)
		}
	}
	a.mu.RUnlock()

	return sessions, nil
}

// GetActiveSessions gets all active sessions
func (a *OODAdapter) GetActiveSessions() []*OODSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sessions := make([]*OODSession, 0)
	for _, session := range a.sessions {
		if session.IsActive() {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// ListFiles lists files in a directory
func (a *OODAdapter) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	if !a.config.EnableFileBrowser {
		return nil, fmt.Errorf(errFmtFileBrowserDisabled, ErrFileBrowsingFailed)
	}

	return a.client.ListFiles(ctx, path)
}

// DownloadFile downloads a file
func (a *OODAdapter) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	if !a.config.EnableFileBrowser {
		return nil, fmt.Errorf(errFmtFileBrowserDisabled, ErrFileBrowsingFailed)
	}

	return a.client.DownloadFile(ctx, path)
}

// UploadFile uploads a file
func (a *OODAdapter) UploadFile(ctx context.Context, path string, content []byte) error {
	if !a.IsRunning() {
		return ErrOODNotConnected
	}

	if !a.config.EnableFileBrowser {
		return fmt.Errorf(errFmtFileBrowserDisabled, ErrFileBrowsingFailed)
	}

	return a.client.UploadFile(ctx, path, content)
}

// DeleteFile deletes a file
func (a *OODAdapter) DeleteFile(ctx context.Context, path string) error {
	if !a.IsRunning() {
		return ErrOODNotConnected
	}

	if !a.config.EnableFileBrowser {
		return fmt.Errorf(errFmtFileBrowserDisabled, ErrFileBrowsingFailed)
	}

	return a.client.DeleteFile(ctx, path)
}

// ListJobTemplates lists available job templates
func (a *OODAdapter) ListJobTemplates(ctx context.Context) ([]JobTemplate, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	return a.client.ListJobTemplates(ctx)
}

// ComposeJob composes a job from a template
func (a *OODAdapter) ComposeJob(ctx context.Context, templateID string, params map[string]string, resources *SessionResources) (*JobComposition, error) {
	if !a.IsRunning() {
		return nil, ErrOODNotConnected
	}

	return a.client.ComposeJob(ctx, templateID, params, resources)
}

// SubmitComposedJob submits a composed job
func (a *OODAdapter) SubmitComposedJob(ctx context.Context, composition *JobComposition) (string, error) {
	if !a.IsRunning() {
		return "", ErrOODNotConnected
	}

	return a.client.SubmitComposedJob(ctx, composition)
}

// CreateStatusReport creates a signed status report for on-chain submission
func (a *OODAdapter) CreateStatusReport(session *OODSession) (*SessionStatusReport, error) {
	report := &SessionStatusReport{
		ProviderAddress:     a.signer.GetProviderAddress(),
		VirtEngineSessionID: session.VirtEngineSessionID,
		OODSessionID:        session.SessionID,
		AppType:             session.AppType,
		State:               session.State,
		VEIDAddress:         session.VEIDAddress,
		StatusMessage:       session.StatusMessage,
		Timestamp:           time.Now(),
	}

	// Calculate usage metrics if session has ended
	if session.StartedAt != nil && session.EndedAt != nil {
		duration := session.EndedAt.Sub(*session.StartedAt)
		report.UsageMetrics = &SessionUsageMetrics{
			WallClockSeconds: int64(duration.Seconds()),
		}
	}

	// Sign the report
	hash := report.Hash()
	sig, err := a.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}
	report.Signature = hex.EncodeToString(sig)

	return report, nil
}

// pollSessions polls for session status updates
func (a *OODAdapter) pollSessions() {
	ticker := time.NewTicker(a.config.SessionPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.updateSessionStatuses()
		}
	}
}

// updateSessionStatuses updates all session statuses
func (a *OODAdapter) updateSessionStatuses() {
	a.mu.RLock()
	sessionIDs := make([]string, 0)
	for id, session := range a.sessions {
		// Only poll active sessions
		if session.IsActive() {
			sessionIDs = append(sessionIDs, id)
		}
	}
	a.mu.RUnlock()

	ctx := context.Background()
	for _, sessionID := range sessionIDs {
		a.mu.RLock()
		session := a.sessions[sessionID]
		a.mu.RUnlock()

		updatedSession, err := a.client.GetSession(ctx, sessionID)
		if err != nil {
			continue
		}

		// Preserve VirtEngine-specific fields
		updatedSession.VirtEngineSessionID = session.VirtEngineSessionID
		updatedSession.VEIDAddress = session.VEIDAddress
		updatedSession.ExpiresAt = session.ExpiresAt

		a.mu.Lock()
		a.sessions[sessionID] = updatedSession
		a.mu.Unlock()

		// Check for session expiration
		if time.Now().After(updatedSession.ExpiresAt) && updatedSession.IsActive() {
			_ = a.TerminateSession(ctx, sessionID)
		}
	}
}

// MapToVirtEngineState maps OOD session state to VirtEngine state
func MapToVirtEngineState(state SessionState) string {
	switch state {
	case SessionStatePending:
		return "pending"
	case SessionStateStarting:
		return "starting"
	case SessionStateRunning:
		return "running"
	case SessionStateSuspended:
		return "suspended"
	case SessionStateCompleted:
		return "completed"
	case SessionStateFailed:
		return "failed"
	case SessionStateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}
