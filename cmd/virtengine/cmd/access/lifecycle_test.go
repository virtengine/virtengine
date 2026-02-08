package access

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLifecycleCommandSuccess(t *testing.T) {
	secret := "supersecret"
	principal := "principal-1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/deployments/lease1/actions", r.URL.Path)

		gotPrincipal := r.Header.Get("X-VE-Principal")
		require.Equal(t, principal, gotPrincipal)
		timestamp := r.Header.Get("X-VE-Timestamp")
		signature := r.Header.Get("X-VE-Signature")
		expected := computeHMACSignature(r.Method, r.URL.Path, r.URL.RawQuery, gotPrincipal, timestamp, secret)
		require.Equal(t, expected, signature)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req lifecycleRequest
		require.NoError(t, json.Unmarshal(body, &req))
		require.Equal(t, "start", req.Action)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"operation_id":"op-1","state":"executing"}`))
	}))
	defer server.Close()

	cmd := newLifecycleActionCmd("start")
	cmd.SetArgs([]string{"lease1"})
	require.NoError(t, cmd.Flags().Set("endpoint", server.URL))
	require.NoError(t, cmd.Flags().Set("principal", principal))
	require.NoError(t, cmd.Flags().Set("secret", secret))

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "op-1")
}

func TestLifecycleCommandPermissionDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer server.Close()

	cmd := newLifecycleActionCmd("stop")
	cmd.SetArgs([]string{"lease1"})
	require.NoError(t, cmd.Flags().Set("endpoint", server.URL))
	require.NoError(t, cmd.Flags().Set("principal", "principal"))
	require.NoError(t, cmd.Flags().Set("secret", "secret"))

	err := cmd.Execute()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "forbidden"))
}

func TestLifecycleCommandResizeRequiresParams(t *testing.T) {
	cmd := newLifecycleActionCmd("resize")
	cmd.SetArgs([]string{"lease1"})
	require.NoError(t, cmd.Flags().Set("endpoint", "http://example.com"))
	require.NoError(t, cmd.Flags().Set("principal", "principal"))
	require.NoError(t, cmd.Flags().Set("secret", "secret"))

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "resize parameter")
}
