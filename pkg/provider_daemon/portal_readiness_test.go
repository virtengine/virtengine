// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestPortalReadinessFailsClosedWithoutDependencies(t *testing.T) {
	server, err := NewPortalAPIServer(DefaultPortalAPIServerConfig())
	require.NoError(t, err)
	router := mux.NewRouter()
	server.setupRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestPortalReadinessReflectsFencedMutationSubmitter(t *testing.T) {
	config := DefaultPortalAPIServerConfig()
	config.Readiness = func(context.Context) ProviderMutationReadiness {
		return ProviderMutationReadiness{Ready: true, Started: true, StoreReady: true, LeaseHeld: true, KeyReady: true}
	}
	server, err := NewPortalAPIServer(config)
	require.NoError(t, err)
	router := mux.NewRouter()
	server.setupRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ready", response.Body.String())
}
