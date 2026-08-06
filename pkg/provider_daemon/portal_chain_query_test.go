package provider_daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopChainQueryFailsClosed(t *testing.T) {
	query := NoopChainQuery{}

	organizations, _, err := query.ListOrganizations(context.Background(), "customer", 20, "")
	require.ErrorIs(t, err, ErrPortalFeatureUnavailable)
	require.Nil(t, organizations)

	tickets, err := query.ListTickets(context.Background(), "customer", "", "")
	require.True(t, errors.Is(err, ErrPortalFeatureUnavailable))
	require.Nil(t, tickets)
}

func TestNoopChainQueryDeclaresRouteCapabilitiesUnavailable(t *testing.T) {
	query := NoopChainQuery{}

	for _, capability := range append(requiredPortalRouteCapabilities, PortalCapabilityOrganizations) {
		err := query.PortalCapability(capability)
		require.ErrorIs(t, err, ErrPortalFeatureUnavailable)

		var unavailable *PortalFeatureUnavailableError
		require.ErrorAs(t, err, &unavailable)
		require.Equal(t, capability, unavailable.Capability)
		if capability == PortalCapabilityOrganizations {
			require.Equal(t, "89C", unavailable.Owner)
		} else {
			require.Equal(t, "86C", unavailable.Owner)
		}
	}
}
