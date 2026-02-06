package cloud

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

func TestAWSCloudHSMProvider(t *testing.T) {
	p, err := NewAWSCloudHSMProvider(hsm.CloudConfig{Region: "us-east-1"}, nil)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, p.Connect(ctx))
	defer p.Close()

	// Generate
	info, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "aws-key")
	require.NoError(t, err)
	assert.Equal(t, "aws-key", info.Label)

	// Sign and verify
	msg := []byte("aws test")
	sig, err := p.Sign(ctx, "aws-key", msg)
	require.NoError(t, err)

	pubKey, err := p.GetPublicKey(ctx, "aws-key")
	require.NoError(t, err)
	edPub := pubKey.(ed25519.PublicKey)
	assert.True(t, ed25519.Verify(edPub, msg, sig))

	// List
	keys, err := p.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	// Delete
	require.NoError(t, p.DeleteKey(ctx, "aws-key"))
	_, err = p.GetKey(ctx, "aws-key")
	require.ErrorIs(t, err, hsm.ErrKeyNotFound)
}

func TestAWSCloudHSMImport(t *testing.T) {
	p, err := NewAWSCloudHSMProvider(hsm.CloudConfig{Region: "eu-west-1"}, nil)
	require.NoError(t, err)
	require.NoError(t, p.Connect(context.Background()))
	defer p.Close()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	info, err := p.ImportKey(context.Background(), hsm.KeyTypeEd25519, "imported", priv)
	require.NoError(t, err)
	assert.Equal(t, "imported", info.Label)
}

func TestAWSCloudHSMNotConnected(t *testing.T) {
	p, err := NewAWSCloudHSMProvider(hsm.CloudConfig{Region: "us-east-1"}, nil)
	require.NoError(t, err)

	_, err = p.GenerateKey(context.Background(), hsm.KeyTypeEd25519, "fail")
	require.ErrorIs(t, err, hsm.ErrNotConnected)
}

func TestAWSCloudHSMValidation(t *testing.T) {
	_, err := NewAWSCloudHSMProvider(hsm.CloudConfig{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "region is required")
}

func TestGCPCloudHSMProvider(t *testing.T) {
	p, err := NewGCPCloudHSMProvider(hsm.CloudConfig{
		ProjectID:   "test-project",
		KeyRingName: "test-ring",
	}, nil)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, p.Connect(ctx))
	defer p.Close()

	info, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "gcp-key")
	require.NoError(t, err)
	assert.Equal(t, "gcp-key", info.Label)

	msg := []byte("gcp test")
	sig, err := p.Sign(ctx, "gcp-key", msg)
	require.NoError(t, err)

	pubKey, err := p.GetPublicKey(ctx, "gcp-key")
	require.NoError(t, err)
	edPub := pubKey.(ed25519.PublicKey)
	assert.True(t, ed25519.Verify(edPub, msg, sig))
}

func TestGCPCloudHSMValidation(t *testing.T) {
	_, err := NewGCPCloudHSMProvider(hsm.CloudConfig{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required")

	_, err = NewGCPCloudHSMProvider(hsm.CloudConfig{ProjectID: "p"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key_ring_name is required")
}

func TestAzureHSMProvider(t *testing.T) {
	p, err := NewAzureHSMProvider(hsm.CloudConfig{KeyVaultName: "test-vault"}, nil)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, p.Connect(ctx))
	defer p.Close()

	info, err := p.GenerateKey(ctx, hsm.KeyTypeEd25519, "az-key")
	require.NoError(t, err)
	assert.Equal(t, "az-key", info.Label)

	msg := []byte("azure test")
	sig, err := p.Sign(ctx, "az-key", msg)
	require.NoError(t, err)

	pubKey, err := p.GetPublicKey(ctx, "az-key")
	require.NoError(t, err)
	edPub := pubKey.(ed25519.PublicKey)
	assert.True(t, ed25519.Verify(edPub, msg, sig))
}

func TestAzureHSMValidation(t *testing.T) {
	_, err := NewAzureHSMProvider(hsm.CloudConfig{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key_vault_name is required")
}
