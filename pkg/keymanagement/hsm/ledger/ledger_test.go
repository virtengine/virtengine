package ledger

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

func newTestSigner(t *testing.T) *Signer {
	t.Helper()
	s, err := NewSigner(hsm.LedgerConfig{}, nil)
	require.NoError(t, err)
	require.NoError(t, s.Connect(context.Background()))
	return s
}

func TestSignerDefaults(t *testing.T) {
	s, err := NewSigner(hsm.LedgerConfig{}, nil)
	require.NoError(t, err)
	assert.Equal(t, DefaultDerivationPath, s.config.DerivationPath)
	assert.Equal(t, DefaultHRP, s.config.HRP)
}

func TestSignerGenerateKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	info, err := s.GenerateKey(ctx, hsm.KeyTypeEd25519, "ledger-key")
	require.NoError(t, err)
	assert.Equal(t, "ledger-key", info.Label)
	assert.Equal(t, hsm.KeyTypeEd25519, info.Type)
}

func TestSignerGenerateKeySecp256k1(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	info, err := s.GenerateKey(context.Background(), hsm.KeyTypeSecp256k1, "secp-key")
	require.NoError(t, err)
	assert.Equal(t, hsm.KeyTypeSecp256k1, info.Type)
}

func TestSignerGenerateKeyUnsupported(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	_, err := s.GenerateKey(context.Background(), "rsa2048", "bad")
	require.ErrorIs(t, err, hsm.ErrUnsupportedKeyType)
}

func TestSignerImportKeyNotSupported(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	_, err := s.ImportKey(context.Background(), hsm.KeyTypeEd25519, "test", []byte("key"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestSignerSignAndVerify(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeEd25519, "sign-key")
	require.NoError(t, err)

	msg := []byte("ledger sign test")
	sig, err := s.Sign(ctx, "sign-key", msg)
	require.NoError(t, err)

	pubKey, err := s.GetPublicKey(ctx, "sign-key")
	require.NoError(t, err)
	edPub := pubKey.(ed25519.PublicKey)
	assert.True(t, ed25519.Verify(edPub, msg, sig))
}

func TestSignerNotConnected(t *testing.T) {
	s, err := NewSigner(hsm.LedgerConfig{}, nil)
	require.NoError(t, err)

	_, err = s.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "fail")
	require.ErrorIs(t, err, hsm.ErrNotConnected)
}

func TestSignerDeleteKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeEd25519, "del-key")
	require.NoError(t, err)

	require.NoError(t, s.DeleteKey(ctx, "del-key"))
	_, err = s.GetKey(ctx, "del-key")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestSignerListKeys(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeEd25519, "k1")
	require.NoError(t, err)
	_, err = s.GenerateKey(ctx, hsm.KeyTypeEd25519, "k2")
	require.NoError(t, err)

	keys, err := s.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestSignerDuplicateKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeEd25519, "dup")
	require.NoError(t, err)

	_, err = s.GenerateKey(ctx, hsm.KeyTypeEd25519, "dup")
	require.ErrorIs(t, err, hsm.ErrKeyExists)
}
