package keymanagement

import (
	"context"
	"crypto"
	"fmt"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// HSMKeyring wraps an hsm.HSMProvider to provide a high-level keyring
// interface for signing and public-key retrieval.
type HSMKeyring struct {
	provider hsm.HSMProvider
}

// NewHSMKeyring creates a keyring backed by the given HSM provider.
func NewHSMKeyring(provider hsm.HSMProvider) (*HSMKeyring, error) {
	if provider == nil {
		return nil, fmt.Errorf("keymanagement: provider must not be nil")
	}
	return &HSMKeyring{provider: provider}, nil
}

// Sign signs msg with the key identified by uid.
func (k *HSMKeyring) Sign(uid string, msg []byte) ([]byte, crypto.PublicKey, error) {
	ctx := context.Background()

	sig, err := k.provider.Sign(ctx, uid, msg)
	if err != nil {
		return nil, nil, fmt.Errorf("hsm keyring: sign: %w", err)
	}

	pubKey, err := k.provider.GetPublicKey(ctx, uid)
	if err != nil {
		return nil, nil, fmt.Errorf("hsm keyring: get public key: %w", err)
	}

	return sig, pubKey, nil
}

// PublicKey returns the public key for the given uid.
func (k *HSMKeyring) PublicKey(uid string) (crypto.PublicKey, error) {
	return k.provider.GetPublicKey(context.Background(), uid)
}

// HasKey returns true if a key with the given uid exists.
func (k *HSMKeyring) HasKey(uid string) bool {
	_, err := k.provider.GetKey(context.Background(), uid)
	return err == nil
}

// ListKeys returns all key labels in the keyring.
func (k *HSMKeyring) ListKeys() ([]string, error) {
	keys, err := k.provider.ListKeys(context.Background())
	if err != nil {
		return nil, err
	}
	labels := make([]string, len(keys))
	for i, ki := range keys {
		labels[i] = ki.Label
	}
	return labels, nil
}
