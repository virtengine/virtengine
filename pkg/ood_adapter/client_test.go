package ood_adapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ood "github.com/virtengine/virtengine/pkg/ood_adapter"
)

const (
	// pingEndpoint is the OOD dashboard ping endpoint path.
	pingEndpoint = "/pun/sys/dashboard/ping"
)

// TestOODProductionClientCreation tests client creation.
func TestOODProductionClientCreation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := ood.OODConfig{
			BaseURL:           "https://ondemand.example.com",
			Cluster:           "test-cluster",
			ConnectionTimeout: 30 * time.Second,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("invalid URL", func(t *testing.T) {
		config := ood.OODConfig{
			BaseURL: "://invalid-url",
		}

		_, err := ood.NewOODProductionClient(config)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid base URL")
	})
}

// TestOODProductionClientConnect tests connection behavior.
func TestOODProductionClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = client.Connect(ctx)
		require.NoError(t, err)
		require.True(t, client.IsConnected())
	})

	t.Run("server requires auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = client.Connect(ctx)
		require.NoError(t, err) // 401 is acceptable
		require.True(t, client.IsConnected())
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = client.Connect(ctx)
		require.Error(t, err)
		require.False(t, client.IsConnected())
	})
}

// TestOODProductionClientDisconnect tests disconnection.
func TestOODProductionClientDisconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL: server.URL,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_ = client.Connect(ctx)
	require.True(t, client.IsConnected())

	err = client.Disconnect()
	require.NoError(t, err)
	require.False(t, client.IsConnected())
}

// TestOODProductionClientAuthenticate tests authentication.
func TestOODProductionClientAuthenticate(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		token := &ood.VEIDToken{
			AccessToken:   "valid-token",
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			VEIDAddress:   "veid1user",
			IdentityScore: 0.95,
		}

		err = client.Authenticate(ctx, "veid1user", token)
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := ood.OODConfig{
			BaseURL: "https://example.com",
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		token := &ood.VEIDToken{
			AccessToken: "token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		err = client.Authenticate(ctx, "veid1user", token)
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		err = client.Authenticate(ctx, "veid1user", nil)
		require.ErrorIs(t, err, ood.ErrInvalidToken)

		expiredToken := &ood.VEIDToken{
			AccessToken: "expired",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}
		err = client.Authenticate(ctx, "veid1user", expiredToken)
		require.ErrorIs(t, err, ood.ErrInvalidToken)
	})
}

// TestOODProductionClientListApps tests listing apps.
func TestOODProductionClientListApps(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/pun/sys/dashboard/batch_connect/sessions/apps" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"token":"sys/bc_jupyter","title":"Jupyter","description":"Notebook"}]`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		apps, err := client.ListApps(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, apps)
	})

	t.Run("fallback to defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		apps, err := client.ListApps(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, apps) // Should return defaults
	})
}

// TestOODProductionClientLaunchApp tests launching apps.
func TestOODProductionClientLaunchApp(t *testing.T) {
	t.Run("successful launch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.Method == http.MethodPost {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"session-123","status":"queued"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
			Cluster: "test",
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     2,
				MemoryGB: 4,
				Hours:    1,
			},
		}

		session, err := client.LaunchApp(ctx, spec)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, "session-123", session.SessionID)
	})

	t.Run("launch failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"internal error"}`))
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		spec := &ood.InteractiveAppSpec{
			AppType: ood.AppTypeJupyter,
			Resources: &ood.SessionResources{
				CPUs:     2,
				MemoryGB: 4,
				Hours:    1,
			},
		}

		_, err = client.LaunchApp(ctx, spec)
		require.Error(t, err)
		require.ErrorIs(t, err, ood.ErrSessionCreationFailed)
	})
}

// TestOODProductionClientGetSession tests getting session status.
func TestOODProductionClientGetSession(t *testing.T) {
	t.Run("session found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"session-123","status":"running","host":"compute-1","port":8080}`))
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		session, err := client.GetSession(ctx, "session-123")
		require.NoError(t, err)
		require.Equal(t, "session-123", session.SessionID)
		require.Equal(t, ood.SessionStateRunning, session.State)
	})

	t.Run("session not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		_, err = client.GetSession(ctx, "nonexistent")
		require.ErrorIs(t, err, ood.ErrSessionNotFound)
	})
}

// TestOODProductionClientTerminateSession tests session termination.
func TestOODProductionClientTerminateSession(t *testing.T) {
	t.Run("successful termination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		err = client.TerminateSession(ctx, "session-123")
		require.NoError(t, err)
	})

	t.Run("session not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pingEndpoint {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := ood.OODConfig{
			BaseURL: server.URL,
		}

		client, err := ood.NewOODProductionClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		_ = client.Connect(ctx)

		err = client.TerminateSession(ctx, "nonexistent")
		require.ErrorIs(t, err, ood.ErrSessionNotFound)
	})
}

// TestOODProductionClientFiles tests file operations.
func TestOODProductionClientFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pingEndpoint {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("download") == "true" {
				_, _ = w.Write([]byte("file content"))
			} else {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"files":[{"name":"test.txt","size":100,"directory":false}]}`))
			}
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL: server.URL,
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_ = client.Connect(ctx)

	t.Run("list files", func(t *testing.T) {
		files, err := client.ListFiles(ctx, "/home/user")
		require.NoError(t, err)
		require.NotEmpty(t, files)
	})

	t.Run("download file", func(t *testing.T) {
		content, err := client.DownloadFile(ctx, "/home/user/test.txt")
		require.NoError(t, err)
		// Server returns file content when download=true query param is present
		require.NotEmpty(t, content)
	})

	t.Run("upload file", func(t *testing.T) {
		err := client.UploadFile(ctx, "/home/user/new.txt", []byte("new content"))
		require.NoError(t, err)
	})

	t.Run("delete file", func(t *testing.T) {
		err := client.DeleteFile(ctx, "/home/user/test.txt")
		require.NoError(t, err)
	})
}

// TestOODProductionClientJobs tests job operations.
func TestOODProductionClientJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pingEndpoint {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/pun/sys/myjobs/templates" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`)) // Return empty for defaults
			return
		}
		if r.URL.Path == "/pun/sys/myjobs/workflows" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"job_id":"slurm-12345"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := ood.OODConfig{
		BaseURL: server.URL,
		Cluster: "test",
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_ = client.Connect(ctx)

	t.Run("list templates", func(t *testing.T) {
		templates, err := client.ListJobTemplates(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, templates) // Should have defaults
	})

	t.Run("compose job", func(t *testing.T) {
		params := map[string]string{
			"job_name": "test",
			"nodes":    "1",
			"time":     "01:00:00",
			"command":  "echo hello",
		}
		composition, err := client.ComposeJob(ctx, "basic-batch", params, nil)
		require.NoError(t, err)
		require.Contains(t, composition.Script, "test")
	})

	t.Run("submit job", func(t *testing.T) {
		composition := &ood.JobComposition{
			TemplateID: "basic-batch",
			Parameters: map[string]string{"job_name": "test"},
			Script:     "#!/bin/bash\necho hello",
		}
		jobID, err := client.SubmitComposedJob(ctx, composition)
		require.NoError(t, err)
		require.Equal(t, "slurm-12345", jobID)
	})
}

// TestOODProductionClientSessionStateMapping tests session state mapping.
func TestOODProductionClientSessionStateMapping(t *testing.T) {
	testCases := []struct {
		status   string
		expected ood.SessionState
	}{
		{"queued", ood.SessionStatePending},
		{"pending", ood.SessionStatePending},
		{"starting", ood.SessionStateStarting},
		{"running", ood.SessionStateRunning},
		{"suspended", ood.SessionStateSuspended},
		{"held", ood.SessionStateSuspended},
		{"completed", ood.SessionStateCompleted},
		{"done", ood.SessionStateCompleted},
		{"failed", ood.SessionStateFailed},
		{"error", ood.SessionStateFailed},
		{"cancelled", ood.SessionStateCancelled},
		{"deleted", ood.SessionStateCancelled},
		{"unknown", ood.SessionStatePending},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == pingEndpoint {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"test","status":"` + tc.status + `"}`))
			}))
			defer server.Close()

			config := ood.OODConfig{BaseURL: server.URL}
			client, _ := ood.NewOODProductionClient(config)
			ctx := context.Background()
			_ = client.Connect(ctx)

			session, err := client.GetSession(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, tc.expected, session.State)
		})
	}
}

// TestOODProductionClientAppTypeMapping tests app type mapping.
func TestOODProductionClientAppTypeMapping(t *testing.T) {
	testCases := []struct {
		token    string
		expected ood.InteractiveAppType
	}{
		{"sys/bc_jupyter", ood.AppTypeJupyter},
		{"jupyter_notebook", ood.AppTypeJupyter},
		{"sys/bc_rstudio", ood.AppTypeRStudio},
		{"rstudio_server", ood.AppTypeRStudio},
		{"sys/bc_desktop", ood.AppTypeVNCDesktop},
		{"vnc_desktop", ood.AppTypeVNCDesktop},
		{"sys/bc_codeserver", ood.AppTypeVSCode},
		{"vscode_server", ood.AppTypeVSCode},
		{"sys/bc_matlab", ood.AppTypeMatlab},
		{"sys/bc_paraview", ood.AppTypeParaView},
		{"custom_app", ood.AppTypeCustom},
	}

	for _, tc := range testCases {
		t.Run(tc.token, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == pingEndpoint {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"token":"` + tc.token + `","title":"App"}]`))
			}))
			defer server.Close()

			config := ood.OODConfig{BaseURL: server.URL}
			client, _ := ood.NewOODProductionClient(config)
			ctx := context.Background()
			_ = client.Connect(ctx)

			apps, err := client.ListApps(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, apps)
			require.Equal(t, tc.expected, apps[0].Type)
		})
	}
}

// TestOODProductionClientNotConnected tests operations when not connected.
func TestOODProductionClientNotConnected(t *testing.T) {
	config := ood.OODConfig{
		BaseURL: "https://example.com",
	}

	client, err := ood.NewOODProductionClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("list apps", func(t *testing.T) {
		_, err := client.ListApps(ctx)
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("launch app", func(t *testing.T) {
		_, err := client.LaunchApp(ctx, &ood.InteractiveAppSpec{})
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("get session", func(t *testing.T) {
		_, err := client.GetSession(ctx, "test")
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("terminate session", func(t *testing.T) {
		err := client.TerminateSession(ctx, "test")
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("list files", func(t *testing.T) {
		_, err := client.ListFiles(ctx, "/home")
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("list templates", func(t *testing.T) {
		_, err := client.ListJobTemplates(ctx)
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})

	t.Run("submit job", func(t *testing.T) {
		_, err := client.SubmitComposedJob(ctx, &ood.JobComposition{})
		require.ErrorIs(t, err, ood.ErrOODNotConnected)
	})
}

