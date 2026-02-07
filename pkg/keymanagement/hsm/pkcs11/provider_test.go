package pkcs11

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	p, err := New(hsm.PKCS11Config{LibraryPath: "/usr/lib/softhsm/libsofthsm2.so"}, logger)
	require.NoError(t, err)
	require.NoError(t, p.Connect(context.Background()))
	return p
}

func TestProviderNew(t *testing.T) {
	_, err := New(hsm.PKCS11Config{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "library path required")
}

func TestProviderConnectIdempotent(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	// Second connect should be a no-op
	require.NoError(t, p.Connect(context.Background()))
}

func TestProviderGenerateKey(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()

	info, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-key", info.Label)
	assert.Equal(t, hsm.KeyTypeEd25519, info.Type)
	assert.False(t, info.Extractable)
	assert.NotEmpty(t, info.Fingerprint)
}

func TestProviderGenerateKeyDuplicate(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "dup-key")
	require.NoError(t, err)

	_, err = p.GenerateKey(ctx, hsm.KeyTypeEd25519, "dup-key")
	require.ErrorIs(t, err, hsm.ErrKeyExists)
}

func TestProviderGenerateKeyUnsupported(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, err := p.GenerateKey(context.Background(), "rsa4096", "rsa-key")
	require.Error(t, err)
	assert.ErrorIs(t, err, hsm.ErrUnsupportedKeyType)
}

func TestProviderNotConnected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p, err := New(hsm.PKCS11Config{LibraryPath: "/tmp/test.so"}, logger)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = p.GenerateKey(ctx, hsm.KeyTypeEd25519, "test")
	require.ErrorIs(t, err, hsm.ErrNotConnected)
}

func TestProviderImportKey(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	ctx := context.Background()
	info, err := p.ImportKey(ctx, hsm.KeyTypeEd25519, "imported", priv)
	require.NoError(t, err)
	assert.Equal(t, "imported", info.Label)
}

func TestProviderImportKeyInvalidSize(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, err := p.ImportKey(context.Background(), hsm.KeyTypeEd25519, "bad", []byte("short"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ed25519 private key size")
}

func TestProviderGetKey(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "get-test")
	require.NoError(t, err)

	info, err := p.GetKey(ctx, "get-test")
	require.NoError(t, err)
	assert.Equal(t, "get-test", info.Label)

	_, err = p.GetKey(ctx, "nonexistent")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestProviderListKeys(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "list-1")
	require.NoError(t, err)
	_, err = p.GenerateKey(ctx, hsm.KeyTypeEd25519, "list-2")
	require.NoError(t, err)

	keys, err := p.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestProviderDeleteKey(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "del-key")
	require.NoError(t, err)

	require.NoError(t, p.DeleteKey(ctx, "del-key"))

	_, err = p.GetKey(ctx, "del-key")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)

	// Delete non-existent
	require.ErrorIs(t, p.DeleteKey(ctx, "del-key"), hsm.ErrKeyNotFound)
}

func TestProviderSignAndVerify(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "sign-key")
	require.NoError(t, err)

	msg := []byte("hello world")
	sig, err := p.Sign(ctx, "sign-key", msg)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)

	pubKey, err := p.GetPublicKey(ctx, "sign-key")
	require.NoError(t, err)

	edPub, ok := pubKey.(ed25519.PublicKey)
	require.True(t, ok)
	assert.True(t, ed25519.Verify(edPub, msg, sig))
}

func TestProviderSignNonexistent(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, err := p.Sign(context.Background(), "ghost", []byte("test"))
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestProviderGetPublicKeyNonexistent(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, err := p.GetPublicKey(context.Background(), "ghost")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestProviderClose(t *testing.T) {
	p := newTestProvider(t)

	_, err := p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "close-key")
	require.NoError(t, err)

	require.NoError(t, p.Close())

	// After close, operations should fail
	_, err = p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "after-close")
	require.ErrorIs(t, err, hsm.ErrNotConnected)
}

func TestKeyExistsAndCount(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	assert.False(t, p.KeyExists("test"))
	assert.Equal(t, 0, p.KeyCount())

	_, err := p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "test")
	require.NoError(t, err)

	assert.True(t, p.KeyExists("test"))
	assert.Equal(t, 1, p.KeyCount())
}

func TestMigrateKey(t *testing.T) {
	p := newTestProvider(t)
	defer p.Close()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	info, err := p.MigrateKey(context.Background(), "migrated", hsm.KeyTypeEd25519, priv)
	require.NoError(t, err)
	assert.Equal(t, "migrated", info.Label)

	// Should be able to sign with migrated key
	sig, err := p.Sign(context.Background(), "migrated", []byte("test"))
	require.NoError(t, err)

	pubKey, err := p.GetPublicKey(context.Background(), "migrated")
	require.NoError(t, err)
	edPub := pubKey.(ed25519.PublicKey)
	assert.True(t, ed25519.Verify(edPub, []byte("test"), sig))
}
