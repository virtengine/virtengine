// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type authoritativePortalChainQuery struct {
	NoopChainQuery
}

func (authoritativePortalChainQuery) PortalCapability(PortalRouteCapability) error {
	return nil
}

type legacyPortalChainQuery struct {
	ChainQuery
}

func TestPortalReadinessFailsClosedWithoutDependencies(t *testing.T) {
	server, err := NewPortalAPIServer(DefaultPortalAPIServerConfig())
	require.NoError(t, err)
	require.ErrorIs(t, server.RouteCapability(PortalCapabilityTickets), ErrPortalFeatureUnavailable)
	router := mux.NewRouter()
	server.setupRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestPortalReadinessReflectsFencedMutationSubmitter(t *testing.T) {
	config := DefaultPortalAPIServerConfig()
	config.ChainQuery = authoritativePortalChainQuery{}
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

func TestPortalStartupRejectsUnavailableEnabledRoutes(t *testing.T) {
	server, err := NewPortalAPIServer(DefaultPortalAPIServerConfig())
	require.NoError(t, err)

	err = server.Start(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPortalFeatureUnavailable))
	require.Contains(t, err.Error(), "tickets")
}

func TestLegacyPortalChainQueryFailsClosed(t *testing.T) {
	config := DefaultPortalAPIServerConfig()
	config.ChainQuery = legacyPortalChainQuery{ChainQuery: authoritativePortalChainQuery{}}
	config.Readiness = func(context.Context) ProviderMutationReadiness {
		return ProviderMutationReadiness{Ready: true, Started: true, StoreReady: true, LeaseHeld: true, KeyReady: true}
	}
	server, err := NewPortalAPIServer(config)
	require.NoError(t, err)

	for _, capability := range requiredPortalRouteCapabilities {
		capabilityErr := server.RouteCapability(capability)
		require.ErrorIs(t, capabilityErr, ErrPortalFeatureUnavailable)
		var unavailable *PortalFeatureUnavailableError
		require.ErrorAs(t, capabilityErr, &unavailable)
		require.Equal(t, capability, unavailable.Capability)
		require.Equal(t, "86C", unavailable.Owner)
	}

	router := mux.NewRouter()
	server.setupRoutes(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)

	startupErr := server.Start(context.Background())
	require.ErrorIs(t, startupErr, ErrPortalFeatureUnavailable)
	var unavailable *PortalFeatureUnavailableError
	require.ErrorAs(t, startupErr, &unavailable)
}
