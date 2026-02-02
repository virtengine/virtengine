// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Integration tests.
//
//go:build e2e.integration
// +build e2e.integration

package ood_adapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ood "github.com/virtengine/virtengine/pkg/ood_adapter"
)

// OOD_TEST_URL environment variable for OOD test server
const envOODTestURL = "OOD_TEST_URL"

// TestIntegrationOODProductionClientConnect tests connecting to a real OOD instance.
func TestIntegrationOODProductionClientConnect(t *testing.T) {
	oodURL := os.Getenv(envOODTestURL)
	if oodURL == "" {
		t.Skip("OOD_TEST_URL not set, skipping integration test")
	}

	config := ood.OODConfig{
		BaseURL:           oodURL,
		Cluster:           "test-cluster",
		ConnectionTimeout: 30 * time.Second,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	require.True(t, client.IsConnected())

	err = client.Disconnect()
	require.NoError(t, err)
	require.False(t, client.IsConnected())
}

// TestIntegrationOODMockServer tests the production client against a mock OOD server.
func TestIntegrationOODMockServer(t *testing.T) {
	// Create mock OOD server
	server := httptest.NewServer(createMockOODHandler(t))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL:           server.URL,
		Cluster:           "test-cluster",
		SLURMPartition:    "interactive",
		ConnectionTimeout: 10 * time.Second,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test connect
	err = client.Connect(ctx)
	require.NoError(t, err)
	require.True(t, client.IsConnected())

	// Test list apps
	apps, err := client.ListApps(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, apps)

	// Test launch session
	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeJupyter,
		Resources: &ood.SessionResources{
			CPUs:     4,
			MemoryGB: 16,
			Hours:    2,
		},
	}

	session, err := client.LaunchApp(ctx, spec)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotEmpty(t, session.SessionID)

	// Test get session
	retrievedSession, err := client.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, session.SessionID, retrievedSession.SessionID)

	// Test list files
	files, err := client.ListFiles(ctx, "/home/user")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Test download file
	content, err := client.DownloadFile(ctx, "/home/user/test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, content)

	// Test upload file
	err = client.UploadFile(ctx, "/home/user/upload.txt", []byte("test content"))
	require.NoError(t, err)

	// Test delete file
	err = client.DeleteFile(ctx, "/home/user/upload.txt")
	require.NoError(t, err)

	// Test list job templates
	templates, err := client.ListJobTemplates(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	// Test compose job
	params := map[string]string{
		"job_name": "test-job",
		"nodes":    "1",
		"time":     "01:00:00",
		"command":  "echo hello",
	}
	composition, err := client.ComposeJob(ctx, "basic-batch", params, &ood.SessionResources{
		CPUs:     4,
		MemoryGB: 8,
		Hours:    1,
	})
	require.NoError(t, err)
	require.NotNil(t, composition)
	require.Contains(t, composition.Script, "test-job")

	// Test submit job
	jobID, err := client.SubmitComposedJob(ctx, composition)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)

	// Test terminate session
	err = client.TerminateSession(ctx, session.SessionID)
	require.NoError(t, err)

	// Test disconnect
	err = client.Disconnect()
	require.NoError(t, err)
}

// TestIntegrationOAuth2TokenManager tests OAuth2 token management.
func TestIntegrationOAuth2TokenManager(t *testing.T) {
	// Create mock OAuth2 server
	server := httptest.NewServer(createMockOAuth2Handler(t))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:             server.URL,
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		RedirectURL:           "http://localhost:8080/callback",
		Scopes:                []string{"openid", "profile", "email", "veid"},
		UsePKCE:               true,
		TokenRefreshThreshold: 5 * time.Minute,
	}

	manager := ood.NewOAuth2TokenManager(config)

	ctx := context.Background()

	// Initialize
	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Generate authorization URL
	authURL, verifier, err := manager.GenerateAuthorizationURL("test-state")
	require.NoError(t, err)
	require.NotEmpty(t, authURL)
	require.NotEmpty(t, verifier)
	require.Contains(t, authURL, "code_challenge")

	// Exchange code for token
	token, err := manager.ExchangeCode(ctx, "test-code", verifier)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.VEIDAddress)

	// Get token
	retrieved := manager.GetToken(token.VEIDAddress)
	require.NotNil(t, retrieved)
	require.Equal(t, token.AccessToken, retrieved.AccessToken)

	// Get valid token
	valid, err := manager.GetValidToken(ctx, token.VEIDAddress)
	require.NoError(t, err)
	require.NotNil(t, valid)

	// Refresh token
	err = manager.RefreshToken(ctx, token.VEIDAddress)
	require.NoError(t, err)

	refreshed := manager.GetToken(token.VEIDAddress)
	require.NotNil(t, refreshed)
	require.Equal(t, 1, refreshed.RefreshCount)

	// Revoke token
	err = manager.RevokeToken(ctx, token.VEIDAddress)
	require.NoError(t, err)

	revoked := manager.GetToken(token.VEIDAddress)
	require.Nil(t, revoked)
}

// TestIntegrationOAuth2ProviderAdapter tests the OAuth2 provider adapter.
func TestIntegrationOAuth2ProviderAdapter(t *testing.T) {
	server := httptest.NewServer(createMockOAuth2Handler(t))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:             server.URL,
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		RedirectURL:           "http://localhost:8080/callback",
		Scopes:                []string{"openid", "profile", "email", "veid"},
		TokenRefreshThreshold: 5 * time.Minute,
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	adapter := ood.NewOAuth2ProviderAdapter(manager)

	// Test GetAuthorizationURL
	authURL := adapter.GetAuthorizationURL("test-state", "http://localhost/callback")
	require.NotEmpty(t, authURL)

	// Test ExchangeCodeForToken
	token, err := adapter.ExchangeCodeForToken(ctx, "test-code", "http://localhost/callback")
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)

	// Test ValidateToken
	validatedToken, err := adapter.ValidateToken(ctx, token.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, validatedToken)
	require.Equal(t, "veid1testuser", validatedToken.VEIDAddress)

	// Test RevokeToken
	err = adapter.RevokeToken(ctx, token.AccessToken)
	require.NoError(t, err)
}

// TestIntegrationFullWorkflow tests the complete OOD workflow.
func TestIntegrationFullWorkflow(t *testing.T) {
	// Create mock servers
	oodServer := httptest.NewServer(createMockOODHandler(t))
	defer oodServer.Close()

	authServer := httptest.NewServer(createMockOAuth2Handler(t))
	defer authServer.Close()

	// Setup OAuth2
	oauth2Config := &ood.OAuth2Config{
		IssuerURL:             authServer.URL,
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		RedirectURL:           "http://localhost:8080/callback",
		Scopes:                []string{"openid", "profile", "email", "veid"},
		TokenRefreshThreshold: 5 * time.Minute,
	}

	tokenManager := ood.NewOAuth2TokenManager(oauth2Config)
	ctx := context.Background()

	err := tokenManager.Initialize(ctx)
	require.NoError(t, err)

	authProvider := ood.NewOAuth2ProviderAdapter(tokenManager)

	// Setup OOD client
	oodConfig := ood.OODConfig{
		BaseURL:             oodServer.URL,
		Cluster:             "test-cluster",
		SLURMPartition:      "interactive",
		SessionPollInterval: 100 * time.Millisecond,
		ConnectionTimeout:   10 * time.Second,
		EnableFileBrowser:   true,
	}

	oodClient, err := ood.NewOODProductionClient(oodConfig)
	require.NoError(t, err)

	mockSigner := ood.NewMockSessionSigner("provider1")

	// Create adapter
	adapter := ood.NewOODAdapter(oodConfig, oodClient, authProvider, mockSigner)

	// Start adapter
	err = adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Authenticate user
	token, err := tokenManager.ExchangeCode(ctx, "test-code", "test-verifier")
	require.NoError(t, err)

	err = adapter.AuthenticateUser(ctx, token.VEIDAddress, token.VEIDToken)
	require.NoError(t, err)

	// List available apps
	apps, err := adapter.ListAvailableApps(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, apps)

	// Launch interactive app
	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeJupyter,
		Resources: &ood.SessionResources{
			CPUs:     4,
			MemoryGB: 16,
			Hours:    4,
		},
		JupyterOptions: &ood.JupyterOptions{
			KernelType:          "python3",
			EnableLabExtensions: true,
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-session-123", token.VEIDAddress, spec)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotEmpty(t, session.SessionID)

	// Get session status
	retrievedSession, err := adapter.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, session.SessionID, retrievedSession.SessionID)

	// Get user sessions
	userSessions, err := adapter.GetUserSessions(ctx, token.VEIDAddress)
	require.NoError(t, err)
	require.NotEmpty(t, userSessions)

	// List files
	files, err := adapter.ListFiles(ctx, "/home/user")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Create status report
	report, err := adapter.CreateStatusReport(session)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotEmpty(t, report.Signature)

	// Terminate session
	err = adapter.TerminateSession(ctx, session.SessionID)
	require.NoError(t, err)
}

// TestIntegrationSessionLifecycleWithPolling tests session state transitions with polling.
func TestIntegrationSessionLifecycleWithPolling(t *testing.T) {
	server := httptest.NewServer(createMockOODHandlerWithStateTransitions(t))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL:             server.URL,
		Cluster:             "test-cluster",
		SessionPollInterval: 50 * time.Millisecond,
		ConnectionTimeout:   10 * time.Second,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, client, mockAuth, mockSigner)

	ctx := context.Background()

	err = adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeJupyter,
		Resources: &ood.SessionResources{
			CPUs:     2,
			MemoryGB: 4,
			Hours:    1,
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-lifecycle", "veid1user", spec)
	require.NoError(t, err)
	require.Equal(t, ood.SessionStatePending, session.State)

	// Wait for session to transition to running (mock server simulates this)
	time.Sleep(200 * time.Millisecond)

	updatedSession, err := adapter.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, ood.SessionStateRunning, updatedSession.State)
	require.NotEmpty(t, updatedSession.ConnectURL)
}

// createMockOODHandler creates a mock OOD HTTP handler for testing.
func createMockOODHandler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	sessions := make(map[string]*mockSession)
	sessionCounter := 0

	// Ping endpoint
	mux.HandleFunc("/pun/sys/dashboard/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	// Active jobs
	mux.HandleFunc("/pun/sys/dashboard/activejobs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jobs":[]}`))
	})

	// List apps
	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/apps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"token":"sys/bc_jupyter","title":"Jupyter Notebook","description":"Interactive Jupyter"},
			{"token":"sys/bc_rstudio","title":"RStudio Server","description":"RStudio"}
		]`))
	})

	// Launch session
	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/contexts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionCounter++
		sessionID := "ood-session-" + string(rune('0'+sessionCounter))
		sessions[sessionID] = &mockSession{
			id:        sessionID,
			status:    "queued",
			createdAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id":"` + sessionID + `",
			"status":"queued",
			"created_at":"` + time.Now().Format(time.RFC3339) + `"
		}`))
	})

	// Get/Delete session
	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Path[len("/pun/sys/dashboard/batch_connect/sessions/"):]

		if r.Method == http.MethodDelete {
			delete(sessions, sessionID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		session, ok := sessions[sessionID]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Simulate state transition
		if time.Since(session.createdAt) > 100*time.Millisecond {
			session.status = "running"
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"` + session.id + `",
			"status":"` + session.status + `",
			"host":"compute-1",
			"port":8080,
			"connect":"https://ondemand.example.com/node/compute-1/8080"
		}`))
	})

	// Files API
	mux.HandleFunc("/pun/sys/files/api/v1/fs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("download") == "true" {
				_, _ = w.Write([]byte("file content"))
			} else {
				_, _ = w.Write([]byte(`{"files":[
					{"name":"test.txt","size":100,"directory":false,"modified_at":1704067200},
					{"name":"data","size":0,"directory":true,"modified_at":1704067200}
				]}`))
			}
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	// Job templates
	mux.HandleFunc("/pun/sys/myjobs/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`)) // Return empty to use defaults
	})

	// Job submission
	mux.HandleFunc("/pun/sys/myjobs/workflows", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"job_id":"slurm-12345","status":"submitted"}`))
	})

	return mux
}

// createMockOAuth2Handler creates a mock OAuth2 HTTP handler for testing.
func createMockOAuth2Handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()

	// Discovery endpoint
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		baseURL := "http://" + r.Host
		_, _ = w.Write([]byte(`{
			"issuer":"` + baseURL + `",
			"authorization_endpoint":"` + baseURL + `/authorize",
			"token_endpoint":"` + baseURL + `/token",
			"userinfo_endpoint":"` + baseURL + `/userinfo",
			"revocation_endpoint":"` + baseURL + `/revoke",
			"jwks_uri":"` + baseURL + `/jwks"
		}`))
	})

	// Token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Return mock token with mock ID token
		idToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJzdWIiOiJ0ZXN0dXNlciIsInZlaWRfYWRkcmVzcyI6InZlaWQxdGVzdHVzZXIiLCJpZGVudGl0eV9zY29yZSI6MC45NX0." +
			"signature"

		_, _ = w.Write([]byte(`{
			"access_token":"mock-access-token-` + time.Now().Format("150405") + `",
			"token_type":"Bearer",
			"expires_in":3600,
			"refresh_token":"mock-refresh-token",
			"scope":"openid profile email veid",
			"id_token":"` + idToken + `"
		}`))
	})

	// Userinfo endpoint
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sub":"testuser",
			"veid_address":"veid1testuser",
			"identity_score":0.95,
			"email":"test@example.com",
			"name":"Test User"
		}`))
	})

	// Revocation endpoint
	mux.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

// createMockOODHandlerWithStateTransitions creates a mock handler with realistic state transitions.
func createMockOODHandlerWithStateTransitions(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	sessions := make(map[string]*mockSession)
	sessionCounter := 0

	mux.HandleFunc("/pun/sys/dashboard/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/apps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"token":"sys/bc_jupyter","title":"Jupyter","description":"Jupyter Notebook"}]`))
	})

	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/contexts/", func(w http.ResponseWriter, r *http.Request) {
		sessionCounter++
		sessionID := "ood-session-" + string(rune('A'+sessionCounter))
		sessions[sessionID] = &mockSession{
			id:        sessionID,
			status:    "queued",
			createdAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"` + sessionID + `","status":"queued"}`))
	})

	mux.HandleFunc("/pun/sys/dashboard/batch_connect/sessions/", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Path[len("/pun/sys/dashboard/batch_connect/sessions/"):]

		if r.Method == http.MethodDelete {
			if s, ok := sessions[sessionID]; ok {
				s.status = "cancelled"
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		session, ok := sessions[sessionID]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Simulate state transitions over time
		elapsed := time.Since(session.createdAt)
		if elapsed > 150*time.Millisecond {
			session.status = "running"
		} else if elapsed > 50*time.Millisecond {
			session.status = "starting"
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"` + session.id + `",
			"status":"` + session.status + `",
			"host":"compute-1",
			"port":8080,
			"connect":"https://ondemand.example.com/node/compute-1/jupyter"
		}`))
	})

	return mux
}

type mockSession struct {
	id        string
	status    string
	createdAt time.Time
}

// TestIntegrationTokenRefreshBackground tests automatic token refresh.
func TestIntegrationTokenRefreshBackground(t *testing.T) {
	server := httptest.NewServer(createMockOAuth2Handler(t))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:             server.URL,
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		RedirectURL:           "http://localhost:8080/callback",
		Scopes:                []string{"openid", "profile", "email", "veid"},
		TokenRefreshThreshold: 50 * time.Millisecond, // Short for testing
	}

	manager := ood.NewOAuth2TokenManager(config)

	ctx := context.Background()
	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Exchange code for token
	token, err := manager.ExchangeCode(ctx, "test-code", "test-verifier")
	require.NoError(t, err)

	// Start manager
	manager.Start()
	defer manager.Stop()

	// Force token to be near expiry
	token.ExpiresAt = time.Now().Add(30 * time.Millisecond)

	// Wait for refresh to occur
	time.Sleep(200 * time.Millisecond)

	// Token should have been refreshed
	refreshed := manager.GetToken(token.VEIDAddress)
	require.NotNil(t, refreshed)
}

// TestIntegrationConcurrentSessions tests handling multiple concurrent sessions.
func TestIntegrationConcurrentSessions(t *testing.T) {
	server := httptest.NewServer(createMockOODHandler(t))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL:             server.URL,
		Cluster:             "test-cluster",
		SessionPollInterval: 100 * time.Millisecond,
		ConnectionTimeout:   10 * time.Second,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, client, mockAuth, mockSigner)

	ctx := context.Background()
	err = adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Launch multiple sessions
	veidAddress := "veid1testuser"
	sessionIDs := make([]string, 5)

	for i := 0; i < 5; i++ {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     2,
				MemoryGB: 4,
				Hours:    1,
			},
		}

		session, err := adapter.LaunchInteractiveApp(ctx, "ve-concurrent-"+string(rune('A'+i)), veidAddress, spec)
		require.NoError(t, err)
		sessionIDs[i] = session.SessionID
	}

	// Verify all sessions are tracked
	sessions, err := adapter.GetUserSessions(ctx, veidAddress)
	require.NoError(t, err)
	require.Len(t, sessions, 5)

	// Verify active sessions
	activeSessions := adapter.GetActiveSessions()
	require.Len(t, activeSessions, 5)

	// Terminate some sessions
	for i := 0; i < 3; i++ {
		err = adapter.TerminateSession(ctx, sessionIDs[i])
		require.NoError(t, err)
	}

	// Verify remaining active sessions
	activeSessions = adapter.GetActiveSessions()
	require.Len(t, activeSessions, 2)
}

// TestIntegrationJobSubmissionWorkflow tests the complete job submission workflow.
func TestIntegrationJobSubmissionWorkflow(t *testing.T) {
	server := httptest.NewServer(createMockOODHandler(t))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL:           server.URL,
		Cluster:           "test-cluster",
		ConnectionTimeout: 10 * time.Second,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, client, mockAuth, mockSigner)

	ctx := context.Background()
	err = adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// List templates
	templates, err := adapter.ListJobTemplates(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	// Find a template
	var templateID string
	for _, t := range templates {
		if t.TemplateID == "basic-batch" {
			templateID = t.TemplateID
			break
		}
	}
	require.NotEmpty(t, templateID)

	// Compose job
	params := map[string]string{
		"job_name": "test-batch-job",
		"nodes":    "4",
		"time":     "02:00:00",
		"command":  "python train.py",
	}
	resources := &ood.SessionResources{
		CPUs:      8,
		MemoryGB:  32,
		Hours:     2,
		Partition: "gpu",
	}

	composition, err := adapter.ComposeJob(ctx, templateID, params, resources)
	require.NoError(t, err)
	require.NotNil(t, composition)
	require.Contains(t, composition.Script, "test-batch-job")
	require.Contains(t, composition.Script, "python train.py")
	require.Contains(t, composition.Script, "--cpus-per-task=8")
	require.Contains(t, composition.Script, "--mem=32G")

	// Submit job
	jobID, err := adapter.SubmitComposedJob(ctx, composition)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)
}
