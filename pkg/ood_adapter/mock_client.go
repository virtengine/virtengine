// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Mock client for testing.
package ood_adapter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Common template parameter labels
const (
	labelJobName   = "Job Name"
	labelTimeLimit = "Time Limit"
)

// MockOODClient is a mock implementation of OODClient for testing
type MockOODClient struct {
	mu             sync.RWMutex
	connected      bool
	authenticated  map[string]bool
	sessions       map[string]*OODSession
	apps           []InteractiveApp
	files          map[string][]FileInfo
	templates      []JobTemplate
	nextSessionID  int
	failConnect    bool
	failAuth       bool
	failLaunch     bool
	failTerminate  bool
}

// NewMockOODClient creates a new mock OOD client
func NewMockOODClient() *MockOODClient {
	return &MockOODClient{
		authenticated: make(map[string]bool),
		sessions:      make(map[string]*OODSession),
		files:         make(map[string][]FileInfo),
		apps:          defaultInteractiveApps(),
		templates:     defaultJobTemplates(),
		nextSessionID: 1000,
	}
}

// defaultInteractiveApps returns default available apps
func defaultInteractiveApps() []InteractiveApp {
	return []InteractiveApp{
		{
			Type:        AppTypeJupyter,
			Name:        "Jupyter Notebook",
			Description: "Interactive Jupyter notebook with Python 3",
			Version:     "3.1.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     32,
				MemoryGB: 256,
				GPUs:     4,
				Hours:    48,
			},
		},
		{
			Type:        AppTypeRStudio,
			Name:        "RStudio Server",
			Description: "RStudio for statistical computing",
			Version:     "2023.12.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 128,
				Hours:    24,
			},
		},
		{
			Type:        AppTypeVNCDesktop,
			Name:        "Virtual Desktop",
			Description: "XFCE virtual desktop via VNC",
			Version:     "1.0.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     2,
				MemoryGB: 4,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 64,
				Hours:    12,
			},
		},
		{
			Type:        AppTypeVSCode,
			Name:        "VS Code Server",
			Description: "Visual Studio Code in the browser",
			Version:     "1.85.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 64,
				Hours:    24,
			},
		},
	}
}

// defaultJobTemplates returns default job templates
func defaultJobTemplates() []JobTemplate {
	return []JobTemplate{
		{
			TemplateID:  "basic-batch",
			Name:        "Basic Batch Job",
			Description: "Simple batch job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --time={{time}}\n\n{{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: labelJobName, Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "1"},
				{Name: "time", Label: labelTimeLimit, Type: "string", Required: true, Default: "01:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 4, MemoryGB: 8, Hours: 1},
		},
		{
			TemplateID:  "mpi-job",
			Name:        "MPI Parallel Job",
			Description: "MPI parallel job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --ntasks-per-node={{tasks}}\n#SBATCH --time={{time}}\n\nmodule load mpi\nmpirun {{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: labelJobName, Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "2"},
				{Name: "tasks", Label: "Tasks per Node", Type: "number", Required: true, Default: "4"},
				{Name: "time", Label: labelTimeLimit, Type: "string", Required: true, Default: "04:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 8, MemoryGB: 16, Hours: 4},
		},
		{
			TemplateID:  "gpu-job",
			Name:        "GPU Job",
			Description: "GPU-enabled job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --gres=gpu:{{gpus}}\n#SBATCH --time={{time}}\n\nmodule load cuda\n{{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: labelJobName, Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "1"},
				{Name: "gpus", Label: "GPUs", Type: "number", Required: true, Default: "1"},
				{Name: "time", Label: labelTimeLimit, Type: "string", Required: true, Default: "02:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 4, MemoryGB: 32, GPUs: 1, Hours: 2},
		},
	}
}

// SetFailConnect sets whether Connect should fail
func (c *MockOODClient) SetFailConnect(fail bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failConnect = fail
}

// SetFailAuth sets whether Authenticate should fail
func (c *MockOODClient) SetFailAuth(fail bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failAuth = fail
}

// SetFailLaunch sets whether LaunchApp should fail
func (c *MockOODClient) SetFailLaunch(fail bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failLaunch = fail
}

// Connect connects to Open OnDemand
func (c *MockOODClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.failConnect {
		return fmt.Errorf("mock connection failure")
	}

	c.connected = true
	return nil
}

// Disconnect disconnects from Open OnDemand
func (c *MockOODClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected checks if connected
func (c *MockOODClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Authenticate authenticates a user via VEID SSO
func (c *MockOODClient) Authenticate(ctx context.Context, veidAddress string, token *VEIDToken) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.failAuth {
		return ErrAuthenticationFailed
	}

	if token == nil || !token.IsValid() {
		return ErrInvalidToken
	}

	c.authenticated[veidAddress] = true
	return nil
}

// ListApps lists available interactive apps
func (c *MockOODClient) ListApps(ctx context.Context) ([]InteractiveApp, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	return c.apps, nil
}

// LaunchApp launches an interactive app session
func (c *MockOODClient) LaunchApp(ctx context.Context, spec *InteractiveAppSpec) (*OODSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	if c.failLaunch {
		return nil, ErrSessionCreationFailed
	}

	c.nextSessionID++
	sessionID := fmt.Sprintf("ood-session-%d", c.nextSessionID)

	session := &OODSession{
		SessionID: sessionID,
		AppType:   spec.AppType,
		State:     SessionStatePending,
		Resources: spec.Resources,
		CreatedAt: time.Now(),
	}

	// Simulate session starting after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		c.mu.Lock()
		if s, ok := c.sessions[sessionID]; ok {
			s.State = SessionStateRunning
			now := time.Now()
			s.StartedAt = &now
			s.ConnectURL = fmt.Sprintf("https://ondemand.example.com/node/compute-%d/%s", c.nextSessionID, spec.AppType)
			s.Host = fmt.Sprintf("compute-%d", c.nextSessionID)
			s.Port = 8080
		}
		c.mu.Unlock()
	}()

	c.sessions[sessionID] = session
	return session, nil
}

// GetSession gets session status
func (c *MockOODClient) GetSession(ctx context.Context, sessionID string) (*OODSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	session, ok := c.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// TerminateSession terminates a session
func (c *MockOODClient) TerminateSession(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrOODNotConnected
	}

	if c.failTerminate {
		return fmt.Errorf("mock termination failure")
	}

	session, ok := c.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	session.State = SessionStateCancelled
	now := time.Now()
	session.EndedAt = &now

	return nil
}

// ListSessions lists active sessions for a user
func (c *MockOODClient) ListSessions(ctx context.Context, veidAddress string) ([]*OODSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	sessions := make([]*OODSession, 0)
	for _, session := range c.sessions {
		if session.VEIDAddress == veidAddress {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// ListFiles lists files in a directory
func (c *MockOODClient) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	// Return mock files for home directory
	if path == "" || path == "/" || path == "/home" {
		return []FileInfo{
			{Name: "Documents", Path: "/home/user/Documents", IsDirectory: true, ModTime: time.Now()},
			{Name: "Downloads", Path: "/home/user/Downloads", IsDirectory: true, ModTime: time.Now()},
			{Name: "data.csv", Path: "/home/user/data.csv", Size: 1024, IsDirectory: false, ModTime: time.Now()},
			{Name: "script.py", Path: "/home/user/script.py", Size: 512, IsDirectory: false, ModTime: time.Now()},
		}, nil
	}

	files, ok := c.files[path]
	if !ok {
		return []FileInfo{}, nil
	}

	return files, nil
}

// DownloadFile downloads a file
func (c *MockOODClient) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	// Return mock content
	return []byte("mock file content for: " + path), nil
}

// UploadFile uploads a file
func (c *MockOODClient) UploadFile(ctx context.Context, path string, content []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrOODNotConnected
	}

	// Mock successful upload
	return nil
}

// DeleteFile deletes a file
func (c *MockOODClient) DeleteFile(ctx context.Context, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrOODNotConnected
	}

	// Mock successful deletion
	return nil
}

// ListJobTemplates lists available job templates
func (c *MockOODClient) ListJobTemplates(ctx context.Context) ([]JobTemplate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	return c.templates, nil
}

// ComposeJob composes a job from a template
func (c *MockOODClient) ComposeJob(ctx context.Context, templateID string, params map[string]string, resources *SessionResources) (*JobComposition, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrOODNotConnected
	}

	// Find template
	var template *JobTemplate
	for i := range c.templates {
		if c.templates[i].TemplateID == templateID {
			template = &c.templates[i]
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Compose script by replacing placeholders
	script := template.Script
	for key, value := range params {
		script = replaceParam(script, key, value)
	}

	return &JobComposition{
		TemplateID:      templateID,
		Parameters:      params,
		Resources:       resources,
		Script:          script,
		OutputDirectory: "/home/user/jobs",
	}, nil
}

// replaceParam replaces a parameter placeholder in a script
func replaceParam(script, key, value string) string {
	placeholder := "{{" + key + "}}"
	result := script
	for {
		idx := findPlaceholder(result, placeholder)
		if idx == -1 {
			break
		}
		result = result[:idx] + value + result[idx+len(placeholder):]
	}
	return result
}

// findPlaceholder finds a placeholder in a string
func findPlaceholder(s, placeholder string) int {
	for i := 0; i <= len(s)-len(placeholder); i++ {
		if s[i:i+len(placeholder)] == placeholder {
			return i
		}
	}
	return -1
}

// SubmitComposedJob submits a composed job
func (c *MockOODClient) SubmitComposedJob(ctx context.Context, composition *JobComposition) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", ErrOODNotConnected
	}

	// Return mock job ID
	return fmt.Sprintf("slurm-job-%d", time.Now().UnixNano()%10000), nil
}

// AddMockSession adds a mock session for testing
func (c *MockOODClient) AddMockSession(session *OODSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[session.SessionID] = session
}

// SetSessionState sets the state of a session for testing
func (c *MockOODClient) SetSessionState(sessionID string, state SessionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if session, ok := c.sessions[sessionID]; ok {
		session.State = state
		if state == SessionStateRunning {
			now := time.Now()
			session.StartedAt = &now
		}
		if state == SessionStateCompleted || state == SessionStateFailed || state == SessionStateCancelled {
			now := time.Now()
			session.EndedAt = &now
		}
	}
}

// MockVEIDAuthProvider is a mock VEID auth provider for testing
type MockVEIDAuthProvider struct {
	failExchange bool
	failRefresh  bool
	failValidate bool
}

// NewMockVEIDAuthProvider creates a new mock auth provider
func NewMockVEIDAuthProvider() *MockVEIDAuthProvider {
	return &MockVEIDAuthProvider{}
}

// SetFailExchange sets whether ExchangeCodeForToken should fail
func (p *MockVEIDAuthProvider) SetFailExchange(fail bool) {
	p.failExchange = fail
}

// SetFailRefresh sets whether RefreshToken should fail
func (p *MockVEIDAuthProvider) SetFailRefresh(fail bool) {
	p.failRefresh = fail
}

// SetFailValidate sets whether ValidateToken should fail
func (p *MockVEIDAuthProvider) SetFailValidate(fail bool) {
	p.failValidate = fail
}

// ExchangeCodeForToken exchanges an authorization code for tokens
func (p *MockVEIDAuthProvider) ExchangeCodeForToken(ctx context.Context, code string, redirectURI string) (*VEIDToken, error) {
	if p.failExchange {
		return nil, ErrAuthenticationFailed
	}

	return &VEIDToken{
		AccessToken:   "mock-access-token",
		TokenType:     "Bearer",
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		RefreshToken:  "mock-refresh-token",
		Scope:         "openid profile email veid",
		VEIDAddress:   "veid1mockaddress",
		IdentityScore: 0.95,
	}, nil
}

// RefreshToken refreshes an expired token
func (p *MockVEIDAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*VEIDToken, error) {
	if p.failRefresh {
		return nil, ErrInvalidToken
	}

	return &VEIDToken{
		AccessToken:   "mock-refreshed-access-token",
		TokenType:     "Bearer",
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		RefreshToken:  "mock-new-refresh-token",
		Scope:         "openid profile email veid",
		VEIDAddress:   "veid1mockaddress",
		IdentityScore: 0.95,
	}, nil
}

// ValidateToken validates a token and returns user info
func (p *MockVEIDAuthProvider) ValidateToken(ctx context.Context, accessToken string) (*VEIDToken, error) {
	if p.failValidate {
		return nil, ErrInvalidToken
	}

	return &VEIDToken{
		AccessToken:   accessToken,
		TokenType:     "Bearer",
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		VEIDAddress:   "veid1mockaddress",
		IdentityScore: 0.95,
	}, nil
}

// GetAuthorizationURL gets the authorization URL for OIDC flow
func (p *MockVEIDAuthProvider) GetAuthorizationURL(state string, redirectURI string) string {
	return fmt.Sprintf("https://auth.veid.example.com/authorize?state=%s&redirect_uri=%s", state, redirectURI)
}

// RevokeToken revokes a token
func (p *MockVEIDAuthProvider) RevokeToken(ctx context.Context, token string) error {
	return nil
}

// MockSessionSigner is a mock session signer for testing
type MockSessionSigner struct {
	providerAddress string
}

// NewMockSessionSigner creates a new mock session signer
func NewMockSessionSigner(providerAddress string) *MockSessionSigner {
	return &MockSessionSigner{
		providerAddress: providerAddress,
	}
}

// Sign signs data and returns the signature
func (s *MockSessionSigner) Sign(data []byte) ([]byte, error) {
	// Return mock signature (hash of data)
	return data[:min(32, len(data))], nil
}

// Verify verifies a signature
func (s *MockSessionSigner) Verify(data []byte, signature []byte) bool {
	return true
}

// GetProviderAddress returns the provider address
func (s *MockSessionSigner) GetProviderAddress() string {
	return s.providerAddress
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
