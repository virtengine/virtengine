//go:build e2e.integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

func TestProviderPortalE2EUsesKubernetesRuntime(t *testing.T) {
	fakeClient := newFakeKubernetesRuntimeClient()
	runtime, err := newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
		ProviderID:     "provider-1",
		ResourcePrefix: "ve",
		NewClient: func(kubeconfig string) (kubernetesRuntimeClient, error) {
			return fakeClient, nil
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workload, err := runtime.adapter.Deploy(ctx, testContainerManifest(), "alloc-1", "alloc-1", provider_daemon.DeploymentOptions{})
	require.NoError(t, err)
	fakeClient.logs[workload.Namespace+"/web-pod-0/web"] = "2026-04-10T11:00:00Z portal runtime log\n"

	listenAddr := freeListenAddr(t)
	cfg := provider_daemon.DefaultPortalAPIServerConfig()
	cfg.ListenAddr = listenAddr
	cfg.AllowInsecure = true
	cfg.RequireVEID = false
	cfg.AuditLogger = newTestAuditLogger(t)
	cfg.WorkloadLogSource = runtime
	cfg.WorkloadShellExecutor = runtime

	server, err := provider_daemon.NewPortalAPIServer(cfg)
	require.NoError(t, err)

	go func() {
		_ = server.Start(ctx)
	}()
	t.Cleanup(func() {
		_ = server.Shutdown(context.Background())
	})

	require.Eventually(t, func() bool {
		resp, err := http.Get("http://" + listenAddr + "/health")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 3*time.Second, 50*time.Millisecond)

	logResp, err := http.Get("http://" + listenAddr + "/deployments/alloc-1/logs?container=web&tail=10")
	require.NoError(t, err)
	defer logResp.Body.Close()
	require.Equal(t, http.StatusOK, logResp.StatusCode)

	logBody := &bytes.Buffer{}
	_, err = logBody.ReadFrom(logResp.Body)
	require.NoError(t, err)
	require.Contains(t, logBody.String(), "portal runtime log")

	sessionResp, err := http.Post(
		"http://"+listenAddr+"/deployments/alloc-1/shell/session",
		"application/json",
		bytes.NewBufferString(`{"container":"web"}`),
	)
	require.NoError(t, err)
	defer sessionResp.Body.Close()
	require.Equal(t, http.StatusOK, sessionResp.StatusCode)

	var sessionPayload map[string]any
	require.NoError(t, json.NewDecoder(sessionResp.Body).Decode(&sessionPayload))
	token, ok := sessionPayload["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	wsURL := "ws://" + listenAddr + "/deployments/alloc-1/shell?token=" + token + "&container=web"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	var sawStdout bool
	var exitCode int
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		require.Equal(t, websocket.BinaryMessage, messageType)
		require.NotEmpty(t, data)

		switch data[0] {
		case 100:
			sawStdout = true
		case 102:
			var result map[string]any
			require.NoError(t, json.Unmarshal(data[1:], &result))
			exitCode = int(result["exit_code"].(float64))
		}
	}

	require.True(t, sawStdout)
	require.Equal(t, 0, exitCode)
}

func freeListenAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().String()
}

func newTestAuditLogger(t *testing.T) *provider_daemon.AuditLogger {
	t.Helper()

	cfg := provider_daemon.DefaultAuditLogConfig()
	cfg.LogFile = filepath.Join(t.TempDir(), "audit.log")

	logger, err := provider_daemon.NewAuditLogger(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, logger.Close())
	})

	return logger
}
