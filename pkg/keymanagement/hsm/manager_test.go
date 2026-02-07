package hsm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerNewManagerValidation(t *testing.T) {
	_, err := NewManager(Config{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backend is required")
}

func TestManagerConnectWithoutProvider(t *testing.T) {
	cfg := DefaultConfig()
	mgr, err := NewManager(cfg, nil)
	require.NoError(t, err)

	err = mgr.Connect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no provider configured")
}

func TestManagerCloseIdempotent(t *testing.T) {
	cfg := DefaultConfig()
	mgr, err := NewManager(cfg, nil)
	require.NoError(t, err)

	// Close twice should not error
	require.NoError(t, mgr.Close())
	require.NoError(t, mgr.Close())
}

func TestManagerOperationsWithoutProvider(t *testing.T) {
	cfg := DefaultConfig()
	mgr, err := NewManager(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = mgr.GenerateKey(ctx, KeyTypeEd25519, "test")
	require.Error(t, err)

	_, err = mgr.ImportKey(ctx, KeyTypeEd25519, "test", nil)
	require.Error(t, err)

	_, err = mgr.GetKey(ctx, "test")
	require.Error(t, err)

	_, err = mgr.ListKeys(ctx)
	require.Error(t, err)

	err = mgr.DeleteKey(ctx, "test")
	require.Error(t, err)

	_, err = mgr.Sign(ctx, "test", nil)
	require.Error(t, err)

	_, err = mgr.GetPublicKey(ctx, "test")
	require.Error(t, err)
}
