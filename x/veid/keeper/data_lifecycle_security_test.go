package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateNoRawBiometricsOnChain(t *testing.T) {
	k, ctx, stateStore := setupTestKeeper(t)
	defer CloseStoreIfNeeded(stateStore)

	t.Run("accepts compact hash payloads", func(t *testing.T) {
		hash := sha256.Sum256([]byte("derived-feature"))
		require.NoError(t, k.ValidateNoRawBiometricsOnChain(ctx, hash[:]))
	})

	t.Run("accepts compact metadata payloads", func(t *testing.T) {
		payload := []byte(`{"face_hash":"6f3c1d2f4b5a6e7d8c9b0a11223344556677889900aabbccddeeff0011223344","model_version":"v1.2.0"}`)
		require.NoError(t, k.ValidateNoRawBiometricsOnChain(ctx, payload))
	})

	t.Run("rejects image magic bytes", func(t *testing.T) {
		payload := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x00}, 128)...)
		err := k.ValidateNoRawBiometricsOnChain(ctx, payload)
		require.Error(t, err)
		require.Contains(t, err.Error(), "raw biometric media payloads")
	})

	t.Run("rejects biometric json fields", func(t *testing.T) {
		payload := []byte(`{"document_scan":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="}`)
		err := k.ValidateNoRawBiometricsOnChain(ctx, payload)
		require.Error(t, err)
		require.Contains(t, err.Error(), "raw biometric")
	})

	t.Run("rejects oversized opaque binary payloads", func(t *testing.T) {
		payload := bytes.Repeat([]byte{0x01, 0x02, 0x03, 0x04}, 300)
		err := k.ValidateNoRawBiometricsOnChain(ctx, payload)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must remain off-chain")
	})
}
