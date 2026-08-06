// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	coordinationclient "k8s.io/client-go/kubernetes/typed/coordination/v1"
	"k8s.io/client-go/rest"
)

func TestKubernetesSubmitterLeaseFencesSplitBrainAndFailover(t *testing.T) {
	client := newLeaseTestClient(t)
	base := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	first, err := NewKubernetesSubmitterLease(client, "virtengine", "pod-a")
	require.NoError(t, err)
	second, err := NewKubernetesSubmitterLease(client, "virtengine", "pod-b")
	require.NoError(t, err)
	first.now = func() time.Time { return base }
	second.now = func() time.Time { return base }

	tokenA, err := first.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.NoError(t, err)
	require.Equal(t, uint64(1), tokenA)
	_, err = second.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.Error(t, err)
	require.True(t, first.Held(context.Background(), "provider-mutation:account", tokenA))

	second.now = func() time.Time { return base.Add(2 * time.Minute) }
	tokenB, err := second.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.NoError(t, err)
	require.Equal(t, uint64(2), tokenB)
	require.False(t, first.Held(context.Background(), "provider-mutation:account", tokenA))
	require.ErrorIs(t, first.Renew(context.Background(), "provider-mutation:account", tokenA, time.Minute), ErrSubmitterLeaseNotHeld)
	require.True(t, second.Held(context.Background(), "provider-mutation:account", tokenB))
}

func TestKubernetesSubmitterLeaseReleaseCannotClearNewOwner(t *testing.T) {
	client := newLeaseTestClient(t)
	base := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	first, err := NewKubernetesSubmitterLease(client, "virtengine", "pod-a")
	require.NoError(t, err)
	second, err := NewKubernetesSubmitterLease(client, "virtengine", "pod-b")
	require.NoError(t, err)
	first.now = func() time.Time { return base }
	second.now = func() time.Time { return base.Add(2 * time.Minute) }
	tokenA, err := first.Acquire(context.Background(), "shared", time.Minute)
	require.NoError(t, err)
	tokenB, err := second.Acquire(context.Background(), "shared", time.Minute)
	require.NoError(t, err)
	require.NoError(t, first.Release(context.Background(), "shared", tokenA))
	require.True(t, second.Held(context.Background(), "shared", tokenB))
}

func newLeaseTestClient(t *testing.T) coordinationclient.LeasesGetter {
	t.Helper()
	server := newLeaseAPIServer()
	t.Cleanup(server.Close)
	client, err := coordinationclient.NewForConfig(&rest.Config{Host: server.URL, ContentConfig: rest.ContentConfig{ContentType: "application/json", AcceptContentTypes: "application/json"}})
	require.NoError(t, err)
	return client
}

type leaseAPIState struct {
	mu      sync.Mutex
	objects map[string]json.RawMessage
	nextRV  uint64
}

func newLeaseAPIServer() *httptest.Server {
	state := &leaseAPIState{objects: make(map[string]json.RawMessage)}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/apis/coordination.k8s.io/v1/namespaces/virtengine/leases/")
		collection := strings.HasSuffix(r.URL.Path, "/leases")
		state.mu.Lock()
		defer state.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			value, ok := state.objects[name]
			if !ok || collection {
				writeLeaseAPIError(w, http.StatusNotFound, name)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(value)
		case http.MethodPost:
			if !collection {
				writeLeaseAPIError(w, http.StatusNotFound, name)
				return
			}
			var object map[string]any
			if json.NewDecoder(r.Body).Decode(&object) != nil {
				writeLeaseAPIError(w, http.StatusBadRequest, "decode")
				return
			}
			metadata := object["metadata"].(map[string]any)
			name = metadata["name"].(string)
			if _, exists := state.objects[name]; exists {
				writeLeaseAPIError(w, http.StatusConflict, name)
				return
			}
			state.nextRV++
			metadata["resourceVersion"] = fmt.Sprint(state.nextRV)
			value, _ := json.Marshal(object)
			state.objects[name] = value
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(value)
		case http.MethodPut:
			var object map[string]any
			if json.NewDecoder(r.Body).Decode(&object) != nil {
				writeLeaseAPIError(w, http.StatusBadRequest, "decode")
				return
			}
			metadata := object["metadata"].(map[string]any)
			var current map[string]any
			if json.Unmarshal(state.objects[name], &current) != nil || metadata["resourceVersion"] != current["metadata"].(map[string]any)["resourceVersion"] {
				writeLeaseAPIError(w, http.StatusConflict, name)
				return
			}
			state.nextRV++
			metadata["resourceVersion"] = fmt.Sprint(state.nextRV)
			value, _ := json.Marshal(object)
			state.objects[name] = value
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(value)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func writeLeaseAPIError(w http.ResponseWriter, status int, name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Failure","message":"lease %s","reason":"NotFound","details":{"name":"%s","group":"coordination.k8s.io","kind":"leases"},"code":%d}`, name, name, status)
}
