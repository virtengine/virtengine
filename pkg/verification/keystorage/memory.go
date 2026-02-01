// Package keystorage provides secure key storage backends for the signer service.
package keystorage

import (
	"context"
	"sync"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// MemoryStorage implements KeyStorage using in-memory storage.
// This is primarily for testing and development; not suitable for production.
type MemoryStorage struct {
	mu         sync.RWMutex
	keys       map[string]*storedKey
	maxKeys    int
	closed     bool
}

// storedKey holds both key info and private key in memory.
type storedKey struct {
	info       *veidtypes.SignerKeyInfo
	privateKey []byte
	storedAt   time.Time
}

// NewMemoryStorage creates a new in-memory key storage.
func NewMemoryStorage(config *MemoryStorageConfig) (*MemoryStorage, error) {
	maxKeys := 100
	if config != nil && config.MaxKeys > 0 {
		maxKeys = config.MaxKeys
	}

	return &MemoryStorage{
		keys:    make(map[string]*storedKey),
		maxKeys: maxKeys,
	}, nil
}

// StoreKey stores a key pair in memory.
func (m *MemoryStorage) StoreKey(ctx context.Context, keyInfo *veidtypes.SignerKeyInfo, privateKey []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	if len(m.keys) >= m.maxKeys {
		return ErrStorageFull.Wrapf("max keys: %d", m.maxKeys)
	}

	if _, exists := m.keys[keyInfo.KeyID]; exists {
		return ErrKeyExists.Wrapf("key ID: %s", keyInfo.KeyID)
	}

	// Copy the private key to avoid external modification
	privateKeyCopy := make([]byte, len(privateKey))
	copy(privateKeyCopy, privateKey)

	// Copy the key info
	keyInfoCopy := *keyInfo

	m.keys[keyInfo.KeyID] = &storedKey{
		info:       &keyInfoCopy,
		privateKey: privateKeyCopy,
		storedAt:   time.Now(),
	}

	return nil
}

// GetKeyInfo retrieves key metadata.
func (m *MemoryStorage) GetKeyInfo(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	stored, ok := m.keys[keyID]
	if !ok {
		return nil, ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	// Return a copy
	infoCopy := *stored.info
	return &infoCopy, nil
}

// GetPrivateKey retrieves the private key.
func (m *MemoryStorage) GetPrivateKey(ctx context.Context, keyID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	stored, ok := m.keys[keyID]
	if !ok {
		return nil, ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	// Return a copy of the private key
	privateKeyCopy := make([]byte, len(stored.privateKey))
	copy(privateKeyCopy, stored.privateKey)

	return privateKeyCopy, nil
}

// ListKeys returns all keys for a signer.
func (m *MemoryStorage) ListKeys(ctx context.Context, signerID string) ([]*veidtypes.SignerKeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	var result []*veidtypes.SignerKeyInfo
	for _, stored := range m.keys {
		if stored.info.SignerID == signerID {
			infoCopy := *stored.info
			result = append(result, &infoCopy)
		}
	}

	return result, nil
}

// UpdateKeyState updates the state of a key.
func (m *MemoryStorage) UpdateKeyState(ctx context.Context, keyID string, state veidtypes.SignerKeyState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	stored, ok := m.keys[keyID]
	if !ok {
		return ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	stored.info.State = state
	return nil
}

// DeleteKey deletes a key.
func (m *MemoryStorage) DeleteKey(ctx context.Context, keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	stored, ok := m.keys[keyID]
	if !ok {
		return ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	// Clear private key before deletion
	for i := range stored.privateKey {
		stored.privateKey[i] = 0
	}

	delete(m.keys, keyID)
	return nil
}

// HealthCheck verifies the storage is accessible.
func (m *MemoryStorage) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	return nil
}

// Close closes the storage and clears all keys.
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	// Clear all private keys
	for _, stored := range m.keys {
		for i := range stored.privateKey {
			stored.privateKey[i] = 0
		}
	}

	m.keys = nil
	m.closed = true
	return nil
}

// Ensure MemoryStorage implements KeyStorage
var _ KeyStorage = (*MemoryStorage)(nil)

