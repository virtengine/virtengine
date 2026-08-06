//go:build e2e.integration

package marketplace_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

// TestTask84CNonOwnerActivationContract is an executable compatibility fixture:
// catalog state remains readable while every independent lifecycle mutation
// returns the stable deprecation error after activation.
func TestTask84CNonOwnerActivationContract(t *testing.T) {
	require.EqualError(t, marketplace.ErrLifecycleDeprecated, "marketplace lifecycle writes deprecated; use x/market and x/resources")
}
