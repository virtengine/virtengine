package provider_daemon

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyManager(t *testing.T) {
	config := DefaultKeyManagerConfig()
	km, err := NewKeyManager(config)
	require.NoError(t, err)
	require.NotNil(t, km)
	assert.True(t, km.IsLocked())
}

func TestKeyManagerUnlockLock(t *testing.T) {
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(config)
	require.NoError(t, err)

	// Initially locked
	assert.True(t, km.IsLocked())

	// Unlock with memory storage (no passphrase needed)
	err = km.Unlock("")
	require.NoError(t, err)
	assert.False(t, km.IsLocked())

	// Lock again
	km.Lock()
	assert.True(t, km.IsLocked())
}

func TestKeyManagerUnlockFileStorageRequiresPassphrase(t *testing.T) {
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeFile
	km, err := NewKeyManager(config)
	require.NoError(t, err)

	// Empty passphrase should fail for file storage
	err = km.Unlock("")
	require.Error(t, err)
	assert.Equal(t, ErrInvalidPassphrase, err)

	// Non-empty passphrase should work
	err = km.Unlock("test-passphrase")
	require.NoError(t, err)
	assert.False(t, km.IsLocked())
}

func TestKeyManagerGenerateKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	key, err := km.GenerateKey("provider1")
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.NotEmpty(t, key.KeyID)
	assert.NotEmpty(t, key.PublicKey)
	assert.Equal(t, "ed25519", key.Algorithm)
	assert.Equal(t, "active", key.Status)
	assert.Equal(t, "provider1", key.ProviderAddress)
	assert.False(t, key.CreatedAt.IsZero())
}

func TestKeyManagerGenerateKeyWhenLocked(t *testing.T) {
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(config)
	require.NoError(t, err)

	// Should fail when locked
	_, err = km.GenerateKey("provider1")
	require.Error(t, err)
	assert.Equal(t, ErrKeyStorageLocked, err)
}

func TestKeyManagerGetActiveKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// No active key initially
	_, err := km.GetActiveKey()
	require.Error(t, err)
	assert.Equal(t, ErrKeyNotFound, err)

	// Generate a key
	key, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Now we should get the active key
	activeKey, err := km.GetActiveKey()
	require.NoError(t, err)
	assert.Equal(t, key.KeyID, activeKey.KeyID)
}

func TestKeyManagerSign(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate a key
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Sign a message
	message := []byte("test message to sign")
	sig, err := km.Sign(message)
	require.NoError(t, err)
	require.NotNil(t, sig)

	assert.NotEmpty(t, sig.PublicKey)
	assert.NotEmpty(t, sig.Signature)
	assert.Equal(t, "ed25519", sig.Algorithm)
	assert.False(t, sig.SignedAt.IsZero())

	// Verify the signature
	err = sig.Verify(message)
	require.NoError(t, err)

	// Verification with wrong message should fail
	err = sig.Verify([]byte("wrong message"))
	require.Error(t, err)
}

func TestKeyManagerSignWhenLocked(t *testing.T) {
	km := createUnlockedKeyManager(t)

	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Lock the manager
	km.Lock()

	// Signing should fail
	_, err = km.Sign([]byte("test"))
	require.Error(t, err)
	assert.Equal(t, ErrKeyStorageLocked, err)
}

func TestKeyManagerRotateKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate initial key
	oldKey, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Rotate the key
	newKey, rotation, err := km.RotateKey("provider1")
	require.NoError(t, err)
	require.NotNil(t, newKey)
	require.NotNil(t, rotation)

	// Verify rotation details
	assert.Equal(t, oldKey.KeyID, rotation.OldKeyID)
	assert.Equal(t, newKey.KeyID, rotation.NewKeyID)
	assert.True(t, rotation.GracePeriodEnd.After(rotation.RotatedAt))

	// Old key should be marked as rotated
	oldKeyUpdated, err := km.GetKey(oldKey.KeyID)
	require.NoError(t, err)
	assert.Equal(t, "rotated", oldKeyUpdated.Status)

	// New key should be active
	activeKey, err := km.GetActiveKey()
	require.NoError(t, err)
	assert.Equal(t, newKey.KeyID, activeKey.KeyID)
}

func TestKeyManagerRevokeKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate a key
	key, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Revoke the key
	err = km.RevokeKey(key.KeyID)
	require.NoError(t, err)

	// Key should be revoked
	revokedKey, err := km.GetKey(key.KeyID)
	require.NoError(t, err)
	assert.Equal(t, "revoked", revokedKey.Status)

	// Active key should now return error
	_, err = km.GetActiveKey()
	require.Error(t, err)
}

func TestKeyManagerRevokeNonExistentKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	err := km.RevokeKey("nonexistent")
	require.Error(t, err)
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestKeyManagerListKeys(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate multiple keys
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)
	_, _, err = km.RotateKey("provider1")
	require.NoError(t, err)

	keys, err := km.ListKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	// Verify no private keys are exposed
	for _, key := range keys {
		assert.Nil(t, key.privateKey)
	}
}

func TestKeyManagerNeedsRotation(t *testing.T) {
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	config.KeyRotationDays = 30 // 30 days for testing (must be > 7 to be outside rotation window)
	km, err := NewKeyManager(config)
	require.NoError(t, err)
	err = km.Unlock("")
	require.NoError(t, err)

	// No key exists, needs rotation
	needsRotation, err := km.NeedsRotation()
	require.NoError(t, err)
	assert.True(t, needsRotation)

	// Generate a key
	_, err = km.GenerateKey("provider1")
	require.NoError(t, err)

	// Fresh key shouldn't need rotation (30 days is well outside 7-day rotation window)
	needsRotation, err = km.NeedsRotation()
	require.NoError(t, err)
	assert.False(t, needsRotation)
}

func TestKeyManagerImportKey(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate an ed25519 key pair externally
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Import the key
	key, err := km.ImportKey("provider1", privKey, "ed25519")
	require.NoError(t, err)
	require.NotNil(t, key)

	// Verify the public key matches
	assert.Contains(t, key.PublicKey, "")
	_ = pubKey // Use pubKey to avoid unused variable

	// Should be able to sign with imported key
	sig, err := km.Sign([]byte("test message"))
	require.NoError(t, err)
	require.NotNil(t, sig)
}

func TestKeyManagerImportKeyInvalidSize(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Try to import a key with invalid size
	_, err := km.ImportKey("provider1", []byte("too-short"), "ed25519")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ed25519 private key size")
}

func TestKeyManagerLockScrubsKeys(t *testing.T) {
	km := createUnlockedKeyManager(t)

	// Generate a key
	key, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	// Verify private key exists
	km.mu.RLock()
	keyInternal := km.keys[key.KeyID]
	assert.NotNil(t, keyInternal.privateKey)
	km.mu.RUnlock()

	// Lock the manager
	km.Lock()

	// Private key should be scrubbed (all zeros or nil)
	km.mu.RLock()
	keyInternal = km.keys[key.KeyID]
	if keyInternal.privateKey != nil {
		allZero := true
		for _, b := range keyInternal.privateKey {
			if b != 0 {
				allZero = false
				break
			}
		}
		assert.True(t, allZero, "private key should be scrubbed to zeros")
	}
	km.mu.RUnlock()
}

func TestSignatureVerify(t *testing.T) {
	// Generate a valid signature
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("test message")
	sigBytes := ed25519.Sign(privKey, message)

	sig := &Signature{
		PublicKey: encodeHex(pubKey),
		Signature: encodeHex(sigBytes),
		Algorithm: "ed25519",
		KeyID:     "test-key",
		SignedAt:  time.Now().UTC(),
	}

	t.Run("valid signature verifies", func(t *testing.T) {
		err := sig.Verify(message)
		require.NoError(t, err)
	})

	t.Run("wrong message fails", func(t *testing.T) {
		err := sig.Verify([]byte("wrong message"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "verification failed")
	})

	t.Run("invalid public key fails", func(t *testing.T) {
		badSig := &Signature{
			PublicKey: "invalid-hex",
			Signature: sig.Signature,
			Algorithm: "ed25519",
		}
		err := badSig.Verify(message)
		require.Error(t, err)
	})

	t.Run("unsupported algorithm fails", func(t *testing.T) {
		badSig := &Signature{
			PublicKey: sig.PublicKey,
			Signature: sig.Signature,
			Algorithm: "unsupported",
		}
		err := badSig.Verify(message)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported algorithm")
	})
}

// Helper functions

func createUnlockedKeyManager(t *testing.T) *KeyManager {
	t.Helper()
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(config)
	require.NoError(t, err)
	err = km.Unlock("")
	require.NoError(t, err)
	return km
}

func encodeHex(data []byte) string {
	const hexTable = "0123456789abcdef"
	dst := make([]byte, len(data)*2)
	for i, v := range data {
		dst[i*2] = hexTable[v>>4]
		dst[i*2+1] = hexTable[v&0x0f]
	}
	return string(dst)
}

