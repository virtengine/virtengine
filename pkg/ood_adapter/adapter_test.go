package ood_adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ood "github.com/virtengine/virtengine/pkg/ood_adapter"
)

// Common test paths
const testHomePath = "/home/test.txt"

// VE-918: Open OnDemand integration tests

// TestOODAdapterCreation tests adapter initialization
func TestOODAdapterCreation(t *testing.T) {
	config := ood.OODConfig{
		BaseURL:             "https://ondemand.example.com",
		Cluster:             "test-cluster",
		SessionPollInterval: time.Second * 10,
	}

	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	require.NotNil(t, adapter)
	require.Equal(t, "test-cluster", adapter.GetConfig().Cluster)
}

// TestOODAdapterStartStop tests starting and stopping the adapter
func TestOODAdapterStartStop(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()

	// Start adapter
	err := adapter.Start(ctx)
	require.NoError(t, err)
	require.True(t, adapter.IsRunning())

	// Stop adapter
	err = adapter.Stop()
	require.NoError(t, err)
	require.False(t, adapter.IsRunning())
}

// TestOODAdapterConnectionFailure tests connection failure handling
func TestOODAdapterConnectionFailure(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockClient.SetFailConnect(true)
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to connect")
	require.False(t, adapter.IsRunning())
}

// TestVEIDSSOAuthentication tests VEID SSO authentication
func TestVEIDSSOAuthentication(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Create valid token
	token := &ood.VEIDToken{
		AccessToken:   "valid-token",
		TokenType:     "Bearer",
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		VEIDAddress:   "veid1mockaddress",
		IdentityScore: 0.95,
	}

	// Authenticate user
	err = adapter.AuthenticateUser(ctx, "veid1mockaddress", token)
	require.NoError(t, err)
}

// TestVEIDSSOAuthenticationFailure tests authentication failure
func TestVEIDSSOAuthenticationFailure(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Test with nil token
	err = adapter.AuthenticateUser(ctx, "veid1test", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ood.ErrInvalidToken)

	// Test with expired token
	expiredToken := &ood.VEIDToken{
		AccessToken: "expired-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	err = adapter.AuthenticateUser(ctx, "veid1test", expiredToken)
	require.Error(t, err)
	require.ErrorIs(t, err, ood.ErrInvalidToken)
}

// TestListInteractiveApps tests listing available interactive apps
func TestListInteractiveApps(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	apps, err := adapter.ListAvailableApps(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, apps)

	// Check for expected apps
	appTypes := make(map[ood.InteractiveAppType]bool)
	for _, app := range apps {
		appTypes[app.Type] = true
	}

	require.True(t, appTypes[ood.AppTypeJupyter], "Jupyter should be available")
	require.True(t, appTypes[ood.AppTypeRStudio], "RStudio should be available")
	require.True(t, appTypes[ood.AppTypeVNCDesktop], "VNC Desktop should be available")
}

// TestLaunchJupyterSession tests launching a Jupyter session
func TestLaunchJupyterSession(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

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

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-session-123", "veid1user", spec)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotEmpty(t, session.SessionID)
	require.Equal(t, ood.AppTypeJupyter, session.AppType)
	require.Equal(t, "ve-session-123", session.VirtEngineSessionID)
	require.Equal(t, "veid1user", session.VEIDAddress)
}

// TestLaunchRStudioSession tests launching an RStudio session
func TestLaunchRStudioSession(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeRStudio,
		Resources: &ood.SessionResources{
			CPUs:     2,
			MemoryGB: 8,
			Hours:    2,
		},
		RStudioOptions: &ood.RStudioOptions{
			RVersion:    "4.3.0",
			EnableShiny: true,
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-session-456", "veid1user", spec)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, ood.AppTypeRStudio, session.AppType)
}

// TestLaunchVirtualDesktop tests launching a virtual desktop session
func TestLaunchVirtualDesktop(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeVNCDesktop,
		Resources: &ood.SessionResources{
			CPUs:     4,
			MemoryGB: 8,
			Hours:    2,
		},
		DesktopOptions: &ood.DesktopOptions{
			Resolution:         "1920x1080",
			DesktopEnvironment: "XFCE",
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-session-789", "veid1user", spec)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, ood.AppTypeVNCDesktop, session.AppType)
}

// TestSessionLifecycle tests the complete session lifecycle
func TestSessionLifecycle(t *testing.T) {
	config := ood.DefaultOODConfig()
	config.SessionPollInterval = 50 * time.Millisecond
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Launch session
	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeJupyter,
		Resources: &ood.SessionResources{
			CPUs:     2,
			MemoryGB: 4,
			Hours:    1,
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-lifecycle-test", "veid1user", spec)
	require.NoError(t, err)
	require.Equal(t, ood.SessionStatePending, session.State)

	// Wait for session to start
	time.Sleep(200 * time.Millisecond)

	// Get updated session status
	updatedSession, err := adapter.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, ood.SessionStateRunning, updatedSession.State)
	require.NotEmpty(t, updatedSession.ConnectURL)

	// Terminate session
	err = adapter.TerminateSession(ctx, session.SessionID)
	require.NoError(t, err)

	// Verify session is cancelled
	finalSession, err := adapter.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	require.Equal(t, ood.SessionStateCancelled, finalSession.State)
}

// TestGetUserSessions tests getting all sessions for a user
func TestGetUserSessions(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	veidAddress := "veid1testuser"

	// Launch multiple sessions
	for i := 0; i < 3; i++ {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
		}
		_, err := adapter.LaunchInteractiveApp(ctx, "", veidAddress, spec)
		require.NoError(t, err)
	}

	// Get user sessions
	sessions, err := adapter.GetUserSessions(ctx, veidAddress)
	require.NoError(t, err)
	require.Len(t, sessions, 3)
}

// TestFileBrowsing tests file browsing functionality
func TestFileBrowsing(t *testing.T) {
	config := ood.DefaultOODConfig()
	config.EnableFileBrowser = true
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// List files
	files, err := adapter.ListFiles(ctx, "/home")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Download file
	content, err := adapter.DownloadFile(ctx, "/home/user/data.csv")
	require.NoError(t, err)
	require.NotEmpty(t, content)

	// Upload file
	err = adapter.UploadFile(ctx, "/home/user/test.txt", []byte("test content"))
	require.NoError(t, err)

	// Delete file
	err = adapter.DeleteFile(ctx, "/home/user/test.txt")
	require.NoError(t, err)
}

// TestFileBrowsingDisabled tests file browsing when disabled
func TestFileBrowsingDisabled(t *testing.T) {
	config := ood.DefaultOODConfig()
	config.EnableFileBrowser = false
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// All file operations should fail
	_, err = adapter.ListFiles(ctx, "/home")
	require.Error(t, err)
	require.ErrorIs(t, err, ood.ErrFileBrowsingFailed)

	_, err = adapter.DownloadFile(ctx, testHomePath)
	require.Error(t, err)

	err = adapter.UploadFile(ctx, testHomePath, []byte("content"))
	require.Error(t, err)

	err = adapter.DeleteFile(ctx, testHomePath)
	require.Error(t, err)
}

// TestJobComposition tests job composition functionality
func TestJobComposition(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// List templates
	templates, err := adapter.ListJobTemplates(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	// Compose job
	params := map[string]string{
		"job_name": "test-job",
		"nodes":    "2",
		"time":     "01:00:00",
		"command":  "echo hello",
	}
	resources := &ood.SessionResources{
		CPUs:     4,
		MemoryGB: 8,
		Hours:    1,
	}

	composition, err := adapter.ComposeJob(ctx, "basic-batch", params, resources)
	require.NoError(t, err)
	require.NotNil(t, composition)
	require.Contains(t, composition.Script, "test-job")
	require.Contains(t, composition.Script, "echo hello")

	// Submit composed job
	jobID, err := adapter.SubmitComposedJob(ctx, composition)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)
}

// TestSessionStatusReport tests creating signed status reports
func TestSessionStatusReport(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Launch session
	spec := &ood.InteractiveAppSpec{
		AppType: ood.AppTypeJupyter,
		Resources: &ood.SessionResources{
			CPUs:     2,
			MemoryGB: 4,
			Hours:    1,
		},
	}

	session, err := adapter.LaunchInteractiveApp(ctx, "ve-report-test", "veid1user", spec)
	require.NoError(t, err)

	// Create status report
	report, err := adapter.CreateStatusReport(session)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, "provider1", report.ProviderAddress)
	require.Equal(t, "ve-report-test", report.VirtEngineSessionID)
	require.Equal(t, ood.AppTypeJupyter, report.AppType)
	require.NotEmpty(t, report.Signature)
}

// TestAppSpecValidation tests interactive app spec validation
func TestAppSpecValidation(t *testing.T) {
	testCases := []struct {
		name        string
		spec        ood.InteractiveAppSpec
		expectError bool
	}{
		{
			name: "valid spec",
			spec: ood.InteractiveAppSpec{
				AppType: ood.AppTypeJupyter,
				Resources: &ood.SessionResources{
					CPUs:     4,
					MemoryGB: 8,
					Hours:    2,
				},
			},
			expectError: false,
		},
		{
			name: "missing app type",
			spec: ood.InteractiveAppSpec{
				AppType: "",
				Resources: &ood.SessionResources{
					CPUs:     4,
					MemoryGB: 8,
					Hours:    2,
				},
			},
			expectError: true,
		},
		{
			name: "missing resources",
			spec: ood.InteractiveAppSpec{
				AppType:   ood.AppTypeJupyter,
				Resources: nil,
			},
			expectError: true,
		},
		{
			name: "zero CPUs",
			spec: ood.InteractiveAppSpec{
				AppType: ood.AppTypeJupyter,
				Resources: &ood.SessionResources{
					CPUs:     0,
					MemoryGB: 8,
					Hours:    2,
				},
			},
			expectError: true,
		},
		{
			name: "zero memory",
			spec: ood.InteractiveAppSpec{
				AppType: ood.AppTypeJupyter,
				Resources: &ood.SessionResources{
					CPUs:     4,
					MemoryGB: 0,
					Hours:    2,
				},
			},
			expectError: true,
		},
		{
			name: "zero hours",
			spec: ood.InteractiveAppSpec{
				AppType: ood.AppTypeJupyter,
				Resources: &ood.SessionResources{
					CPUs:     4,
					MemoryGB: 8,
					Hours:    0,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestVEIDTokenValidation tests VEID token validation
func TestVEIDTokenValidation(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		token := &ood.VEIDToken{
			AccessToken: "valid-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}
		require.True(t, token.IsValid())
		require.False(t, token.IsExpired())
	})

	t.Run("expired token", func(t *testing.T) {
		token := &ood.VEIDToken{
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}
		require.False(t, token.IsValid())
		require.True(t, token.IsExpired())
	})

	t.Run("empty token", func(t *testing.T) {
		token := &ood.VEIDToken{
			AccessToken: "",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}
		require.False(t, token.IsValid())
	})
}

// TestSessionStateMapping tests session state mapping
func TestSessionStateMapping(t *testing.T) {
	testCases := []struct {
		state    ood.SessionState
		expected string
	}{
		{ood.SessionStatePending, "pending"},
		{ood.SessionStateStarting, "starting"},
		{ood.SessionStateRunning, "running"},
		{ood.SessionStateSuspended, "suspended"},
		{ood.SessionStateCompleted, "completed"},
		{ood.SessionStateFailed, "failed"},
		{ood.SessionStateCancelled, "cancelled"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.state), func(t *testing.T) {
			result := ood.MapToVirtEngineState(tc.state)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestSessionIsActiveAndTerminal tests session state helpers
func TestSessionIsActiveAndTerminal(t *testing.T) {
	t.Run("pending is active", func(t *testing.T) {
		session := &ood.OODSession{State: ood.SessionStatePending}
		require.True(t, session.IsActive())
		require.False(t, session.IsTerminal())
	})

	t.Run("running is active", func(t *testing.T) {
		session := &ood.OODSession{State: ood.SessionStateRunning}
		require.True(t, session.IsActive())
		require.False(t, session.IsTerminal())
	})

	t.Run("completed is terminal", func(t *testing.T) {
		session := &ood.OODSession{State: ood.SessionStateCompleted}
		require.False(t, session.IsActive())
		require.True(t, session.IsTerminal())
	})

	t.Run("failed is terminal", func(t *testing.T) {
		session := &ood.OODSession{State: ood.SessionStateFailed}
		require.False(t, session.IsActive())
		require.True(t, session.IsTerminal())
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		session := &ood.OODSession{State: ood.SessionStateCancelled}
		require.False(t, session.IsActive())
		require.True(t, session.IsTerminal())
	})
}

// TestInteractiveAppsManager tests the interactive apps manager
func TestInteractiveAppsManager(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	manager := ood.NewInteractiveAppsManager(adapter)

	t.Run("launch jupyter", func(t *testing.T) {
		resources := &ood.SessionResources{CPUs: 2, MemoryGB: 4, Hours: 1}
		options := &ood.JupyterOptions{KernelType: "python3"}

		session, err := manager.LaunchJupyter(ctx, "ve-jupyter-1", "veid1user", resources, options)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, ood.AppTypeJupyter, session.AppType)
	})

	t.Run("launch rstudio", func(t *testing.T) {
		resources := &ood.SessionResources{CPUs: 2, MemoryGB: 4, Hours: 1}
		options := &ood.RStudioOptions{RVersion: "4.3.0"}

		session, err := manager.LaunchRStudio(ctx, "ve-rstudio-1", "veid1user", resources, options)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, ood.AppTypeRStudio, session.AppType)
	})

	t.Run("launch virtual desktop", func(t *testing.T) {
		resources := &ood.SessionResources{CPUs: 4, MemoryGB: 8, Hours: 2}
		options := &ood.DesktopOptions{Resolution: "1920x1080"}

		session, err := manager.LaunchVirtualDesktop(ctx, "ve-desktop-1", "veid1user", resources, options)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, ood.AppTypeVNCDesktop, session.AppType)
	})

	t.Run("launch vscode", func(t *testing.T) {
		resources := &ood.SessionResources{CPUs: 2, MemoryGB: 4, Hours: 4}

		session, err := manager.LaunchVSCode(ctx, "ve-vscode-1", "veid1user", resources, "/home/user/project")
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, ood.AppTypeVSCode, session.AppType)
	})
}

// TestSessionTokenManager tests the session token manager
func TestSessionTokenManager(t *testing.T) {
	manager := ood.NewSessionTokenManager()

	token := &ood.VEIDToken{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		VEIDAddress: "veid1user",
	}

	// Store token
	manager.StoreToken("veid1user", token)

	// Get token
	retrieved := manager.GetToken("veid1user")
	require.NotNil(t, retrieved)
	require.Equal(t, "veid1user", retrieved.VEIDAddress)

	// Check validity
	require.True(t, manager.IsTokenValid("veid1user"))

	// Remove token
	manager.RemoveToken("veid1user")
	require.Nil(t, manager.GetToken("veid1user"))
	require.False(t, manager.IsTokenValid("veid1user"))
}

// TestQuotaValidation tests quota validation
func TestQuotaValidation(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	manager := ood.NewInteractiveAppsManager(adapter)
	quota := ood.DefaultAppQuota()

	t.Run("valid request", func(t *testing.T) {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     4,
				MemoryGB: 8,
				Hours:    2,
			},
		}
		err := manager.ValidateAgainstQuota(ctx, "veid1user", spec, quota)
		require.NoError(t, err)
	})

	t.Run("exceeds CPU quota", func(t *testing.T) {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     100, // Exceeds quota
				MemoryGB: 8,
				Hours:    2,
			},
		}
		err := manager.ValidateAgainstQuota(ctx, "veid1user", spec, quota)
		require.Error(t, err)
		require.Contains(t, err.Error(), "CPUs")
	})

	t.Run("exceeds memory quota", func(t *testing.T) {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     4,
				MemoryGB: 1000, // Exceeds quota
				Hours:    2,
			},
		}
		err := manager.ValidateAgainstQuota(ctx, "veid1user", spec, quota)
		require.Error(t, err)
		require.Contains(t, err.Error(), "memory")
	})

	t.Run("exceeds hours quota", func(t *testing.T) {
		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     4,
				MemoryGB: 8,
				Hours:    100, // Exceeds quota
			},
		}
		err := manager.ValidateAgainstQuota(ctx, "veid1user", spec, quota)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hours")
	})
}

// TestOperationsWhenNotRunning tests that operations fail when adapter is not running
func TestOperationsWhenNotRunning(t *testing.T) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner("provider1")

	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)
	ctx := context.Background()

	// All operations should fail when not running
	_, err := adapter.ListAvailableApps(ctx)
	require.ErrorIs(t, err, ood.ErrOODNotConnected)

	spec := &ood.InteractiveAppSpec{
		AppType:   ood.AppTypeJupyter,
		Resources: &ood.SessionResources{CPUs: 1, MemoryGB: 2, Hours: 1},
	}
	_, err = adapter.LaunchInteractiveApp(ctx, "test", "veid1user", spec)
	require.ErrorIs(t, err, ood.ErrOODNotConnected)

	_, err = adapter.GetSession(ctx, "session-1")
	require.ErrorIs(t, err, ood.ErrOODNotConnected)

	err = adapter.TerminateSession(ctx, "session-1")
	require.ErrorIs(t, err, ood.ErrOODNotConnected)

	_, err = adapter.ListFiles(ctx, "/home")
	require.ErrorIs(t, err, ood.ErrOODNotConnected)

	_, err = adapter.ListJobTemplates(ctx)
	require.ErrorIs(t, err, ood.ErrOODNotConnected)
}

