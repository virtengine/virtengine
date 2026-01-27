package rest

import (
	"context"
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil"
)

func TestNewClientWithProviderURL(t *testing.T) {
	ctx := context.Background()
	providerURL := "https://example.com:8443"
	addr, err := sdk.AccAddressFromBech32("akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63")
	require.NoError(t, err)

	t.Run("basic functionality", func(t *testing.T) {
		cl, err := NewClient(ctx, addr, WithProviderURL(providerURL))
		require.NoError(t, err)
		require.NotNil(t, cl)

		c := cl.(*client)
		require.Nil(t, c.cclient)
		require.Equal(t, ctx, c.ctx)
		require.Equal(t, addr, c.addr)
		require.NotNil(t, c.host)
		require.NotNil(t, c.tlsCfg)
	})

	t.Run("options are executed", func(t *testing.T) {
		token := "test-token"

		cl, err := NewClient(ctx, addr, WithProviderURL(providerURL), WithAuthToken(token))
		require.NoError(t, err)

		c := cl.(*client)
		require.Nil(t, c.cclient)
		require.Equal(t, token, c.opts.token)
		require.Equal(t, providerURL, c.opts.providerURL)
	})

	t.Run("invalid URL", func(t *testing.T) {
		invalidURL := "://invalid-url"
		_, err := NewClient(ctx, addr, WithProviderURL(invalidURL))
		require.Error(t, err)
	})

	t.Run("option error handling", func(t *testing.T) {
		testError := errors.New("test error")
		errorOption := func(*clientOptions) error {
			return testError
		}

		_, err := NewClient(ctx, addr, WithProviderURL(providerURL), errorOption)
		require.Error(t, err)
		require.Equal(t, testError, err)
	})

	t.Run("JWT auth without cert querier", func(t *testing.T) {
		// Test that client can be created without cert querier (for JWT auth)
		cl, err := NewClient(ctx, addr, WithProviderURL(providerURL))
		require.NoError(t, err)

		c := cl.(*client)
		require.Nil(t, c.opts.signer)  // Should be nil when no signer provided
		require.Empty(t, c.opts.token) // Should be empty when no token provided
	})

}
