package pkcs11

import (
	"context"
	"fmt"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// MigrateKey copies a key from a source provider into this PKCS#11 provider.
// This is used for file-to-HSM key migration.
func (p *Provider) MigrateKey(ctx context.Context, label string, keyType hsm.KeyType, privateKey []byte) (*hsm.KeyInfo, error) {
	info, err := p.ImportKey(ctx, keyType, label, privateKey)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: migrate key: %w", err)
	}
	return info, nil
}

// KeyExists checks whether a key with the given label exists.
func (p *Provider) KeyExists(label string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.keys[label]
	return ok
}

// KeyCount returns the number of keys stored.
func (p *Provider) KeyCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.keys)
}
