package provider_daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type mockShellExecutor struct {
	calls int
	err   error
}

func (m *mockShellExecutor) OpenShell(ctx context.Context, req *ShellExecutionRequest) error {
	m.calls++
	if m.err != nil {
		return m.err
	}
	if req != nil && req.Stdout != nil {
		_, _ = req.Stdout.Write([]byte("shell-ready\n"))
	}
	return nil
}

type mockPortalLogSource struct {
	entries []LogEntry
	calls   int
}

func (m *mockPortalLogSource) TailLogs(ctx context.Context, deploymentID, container string, tail int) ([]LogEntry, error) {
	m.calls++
	return append([]LogEntry(nil), m.entries...), nil
}

func (m *mockPortalLogSource) StreamLogs(ctx context.Context, deploymentID, container string, tail int) (<-chan LogEntry, func(), error) {
	ch := make(chan LogEntry, len(m.entries))
	for _, entry := range m.entries {
		ch <- entry
	}
	close(ch)
	return ch, func() {}, nil
}

func TestPortalShellSessionRequiresRuntimeBackend(t *testing.T) {
	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = true
	cfg.RequireVEID = false
	cfg.AuditLogger = newTestPortalAuditLogger(t)

	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/deployments/lease1/shell/session", strings.NewReader(`{"container":"web"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "lease1"})
	rr := httptest.NewRecorder()

	server.handleShellSession(rr, req)

	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
	require.Contains(t, rr.Body.String(), "shell backend unavailable")
}

func TestPortalShellSessionIssuesTokenWhenBackendConfigured(t *testing.T) {
	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = true
	cfg.RequireVEID = false
	cfg.WorkloadShellExecutor = &mockShellExecutor{}
	cfg.AuditLogger = newTestPortalAuditLogger(t)

	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/deployments/lease1/shell/session", strings.NewReader(`{"container":"web"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "lease1"})
	rr := httptest.NewRecorder()

	server.handleShellSession(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &payload))
	require.NotEmpty(t, payload["token"])
	require.Equal(t, "lease1", payload["deployment"])
	require.Equal(t, "web", payload["container"])
}

func TestPortalLogsUseWorkloadLogSourceWhenConfigured(t *testing.T) {
	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = true
	cfg.RequireVEID = false
	cfg.AuditLogger = newTestPortalAuditLogger(t)
	cfg.WorkloadLogSource = &mockPortalLogSource{
		entries: []LogEntry{{
			Timestamp: time.Date(2026, time.April, 10, 11, 0, 0, 0, time.UTC),
			Level:     "info",
			Message:   "runtime log entry",
		}},
	}

	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/deployments/lease1/logs?tail=10", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "lease1"})
	rr := httptest.NewRecorder()

	server.handleLogs(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "runtime log entry")
}

func newTestPortalAuditLogger(t *testing.T) *AuditLogger {
	t.Helper()

	cfg := DefaultAuditLogConfig()
	cfg.LogFile = filepath.Join(t.TempDir(), "audit.log")

	logger, err := NewAuditLogger(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, logger.Close())
	})

	return logger
}

var _ WorkloadShellExecutor = (*mockShellExecutor)(nil)
var _ WorkloadLogSource = (*mockPortalLogSource)(nil)
