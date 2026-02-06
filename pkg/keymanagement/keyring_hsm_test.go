package keymanagement

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	"github.com/virtengine/virtengine/pkg/keymanagement/hsm/pkcs11"
)

func newTestKeyring(t *testing.T) (*HSMKeyring, *pkcs11.Provider) {
	t.Helper()
	p, err := pkcs11.New(hsm.PKCS11Config{LibraryPath: "/tmp/test.so"}, nil)
	require.NoError(t, err)
	require.NoError(t, p.Connect(context.Background()))

	kr, err := NewHSMKeyring(p)
	require.NoError(t, err)
	return kr, p
}

func TestNewHSMKeyringNilProvider(t *testing.T) {
	_, err := NewHSMKeyring(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be nil")
}

func TestHSMKeyringSign(t *testing.T) {
	kr, p := newTestKeyring(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "sign-key")
	require.NoError(t, err)

	msg := []byte("keyring sign test")
	sig, pubKey, err := kr.Sign("sign-key", msg)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)

	edPub, ok := pubKey.(ed25519.PublicKey)
	require.True(t, ok)
	assert.True(t, ed25519.Verify(edPub, msg, sig))
}

func TestHSMKeyringPublicKey(t *testing.T) {
	kr, p := newTestKeyring(t)
	defer p.Close()

	_, err := p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "pub-key")
	require.NoError(t, err)

	pubKey, err := kr.PublicKey("pub-key")
	require.NoError(t, err)
	assert.NotNil(t, pubKey)
}

func TestHSMKeyringHasKey(t *testing.T) {
	kr, p := newTestKeyring(t)
	defer p.Close()

	assert.False(t, kr.HasKey("nope"))

	_, err := p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "exists")
	require.NoError(t, err)

	assert.True(t, kr.HasKey("exists"))
}

func TestHSMKeyringListKeys(t *testing.T) {
	kr, p := newTestKeyring(t)
	defer p.Close()

	ctx := context.Background()
	_, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "k1")
	require.NoError(t, err)
	_, err = p.GenerateKey(ctx, hsm.KeyTypeEd25519, "k2")
	require.NoError(t, err)

	labels, err := kr.ListKeys()
	require.NoError(t, err)
	assert.Len(t, labels, 2)
}
