package ledger

import (
	"context"
	"testing"

	sdksecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	vecrypto "github.com/virtengine/virtengine/x/encryption/crypto"
)

type fakeWalletClient struct {
	publicKey []byte
	signature []byte
}

func (f *fakeWalletClient) Connect(context.Context) error { return nil }
func (f *fakeWalletClient) Disconnect() error             { return nil }
func (f *fakeWalletClient) GetPublicKey(context.Context, string) ([]byte, error) {
	return append([]byte(nil), f.publicKey...), nil
}
func (f *fakeWalletClient) SignTransaction(context.Context, string, []byte) (*vecrypto.LedgerSignature, error) {
	return &vecrypto.LedgerSignature{
		Signature: append([]byte(nil), f.signature...),
		PublicKey: append([]byte(nil), f.publicKey...),
	}, nil
}

func newTestSigner(t *testing.T) *Signer {
	t.Helper()

	originalFactory := newLedgerWalletClient
	t.Cleanup(func() {
		newLedgerWalletClient = originalFactory
	})

	newLedgerWalletClient = func(config *vecrypto.LedgerWalletConfig) ledgerWalletClient {
		return &fakeWalletClient{
			publicKey: []byte{
				0x02, 0x7d, 0x59, 0x4f, 0x57, 0x2a, 0x30, 0xa2, 0x58, 0xd8, 0xf4,
				0x6d, 0x9c, 0x8f, 0x22, 0x8f, 0xae, 0x91, 0x7a, 0x34, 0x4d, 0x97,
				0xf9, 0x72, 0x86, 0x34, 0x4a, 0xe1, 0xea, 0x56, 0x08, 0x3a, 0x44,
			},
			signature: []byte{0x30, 0x44, 0x02, 0x20, 0x11, 0x22},
		}
	}

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

func TestSignerGenerateKeySecp256k1(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	info, err := s.GenerateKey(context.Background(), hsm.KeyTypeSecp256k1, "secp-key")
	require.NoError(t, err)
	assert.Equal(t, "secp-key", info.Label)
	assert.Equal(t, hsm.KeyTypeSecp256k1, info.Type)
	assert.False(t, info.Extractable)
	assert.NotEmpty(t, info.Fingerprint)
}

func TestSignerGenerateKeyUnsupported(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	_, err := s.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "bad")
	require.ErrorIs(t, err, hsm.ErrUnsupportedKeyType)
}

func TestSignerImportKeyNotSupported(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	_, err := s.ImportKey(context.Background(), hsm.KeyTypeSecp256k1, "test", []byte("key"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestSignerSignAndExposePublicKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "sign-key")
	require.NoError(t, err)

	msg := []byte("ledger sign test")
	sig, err := s.Sign(ctx, "sign-key", msg)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x30, 0x44, 0x02, 0x20, 0x11, 0x22}, sig)

	pubKey, err := s.GetPublicKey(ctx, "sign-key")
	require.NoError(t, err)
	secpPub := pubKey.(*sdksecp256k1.PubKey)
	assert.Len(t, secpPub.Key, 33)
}

func TestSignerNotConnected(t *testing.T) {
	s, err := NewSigner(hsm.LedgerConfig{}, nil)
	require.NoError(t, err)

	_, err = s.GenerateKey(context.Background(), hsm.KeyTypeSecp256k1, "fail")
	require.ErrorIs(t, err, hsm.ErrNotConnected)
}

func TestSignerDeleteKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "del-key")
	require.NoError(t, err)

	require.NoError(t, s.DeleteKey(ctx, "del-key"))
	_, err = s.GetKey(ctx, "del-key")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestSignerListKeys(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "k1")
	require.NoError(t, err)
	_, err = s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "k2")
	require.NoError(t, err)

	keys, err := s.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestSignerDuplicateKey(t *testing.T) {
	s := newTestSigner(t)
	defer s.Close()

	ctx := context.Background()
	_, err := s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "dup")
	require.NoError(t, err)

	_, err = s.GenerateKey(ctx, hsm.KeyTypeSecp256k1, "dup")
	require.ErrorIs(t, err, hsm.ErrKeyExists)
}
