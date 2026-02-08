package keys

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKeyManager_Initialize(t *testing.T) {
	km := NewKeyManager()
	err := km.Initialize()
	require.NoError(t, err)

	// Verify keys were generated for all scopes
	scopes := []Scope{ScopeVEID, ScopeSupport, ScopeMarket, ScopeAudit}
	for _, scope := range scopes {
		keyInfo, err := km.GetActiveKey(scope)
		require.NoError(t, err)
		require.NotNil(t, keyInfo)
		require.Equal(t, scope, keyInfo.Scope)
		require.Equal(t, uint32(1), keyInfo.Version)
		require.Equal(t, KeyStatusActive, keyInfo.Status)
	}
}

func TestKeyManager_GenerateKey(t *testing.T) {
	km := NewKeyManager()

	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	// Get active key
	keyInfo, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	require.NotNil(t, keyInfo)
	require.Equal(t, ScopeVEID, keyInfo.Scope)
	require.NotEmpty(t, keyInfo.ID)
	require.NotEmpty(t, keyInfo.PublicKey)
	require.NotEmpty(t, keyInfo.PrivateKey)
}

func TestKeyManager_RotateKey(t *testing.T) {
	km := NewKeyManager()

	// Generate initial key
	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	// Get initial key
	oldKey, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	oldKeyID := oldKey.ID

	// Rotate key
	overlapDuration := 1 * time.Hour
	err = km.RotateKey(ScopeVEID, overlapDuration)
	require.NoError(t, err)

	// Get new active key
	newKey, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	require.NotEqual(t, oldKeyID, newKey.ID)
	require.Equal(t, uint32(2), newKey.Version)
	require.Equal(t, KeyStatusActive, newKey.Status)

	// Verify old key is deprecated
	oldKeyUpdated, err := km.GetKey(ScopeVEID, oldKeyID)
	require.NoError(t, err)
	require.Equal(t, KeyStatusDeprecated, oldKeyUpdated.Status)
	require.NotNil(t, oldKeyUpdated.DeprecatedAt)

	// Verify rotation state
	rotation, err := km.GetRotationStatus(ScopeVEID)
	require.NoError(t, err)
	require.Equal(t, oldKeyID, rotation.OldKeyID)
	require.Equal(t, newKey.ID, rotation.NewKeyID)
	require.Equal(t, RotationStatusInProgress, rotation.Status)
}

func TestKeyManager_CompleteRotation(t *testing.T) {
	km := NewKeyManager()

	// Generate initial key and rotate
	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)
	oldKey, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	oldKeyID := oldKey.ID

	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	// Complete rotation
	err = km.CompleteRotation(ScopeVEID)
	require.NoError(t, err)

	// Verify old key is retired
	oldKeyRetired, err := km.GetKey(ScopeVEID, oldKeyID)
	require.NoError(t, err)
	require.Equal(t, KeyStatusRetired, oldKeyRetired.Status)

	// Verify rotation state
	rotation, err := km.GetRotationStatus(ScopeVEID)
	require.NoError(t, err)
	require.Equal(t, RotationStatusCompleted, rotation.Status)
}

func TestKeyManager_RotateKey_PreventsDuplicateRotation(t *testing.T) {
	km := NewKeyManager()

	// Generate initial key
	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	// Start rotation
	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	// Try to rotate again - should fail
	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.ErrorIs(t, err, ErrKeyRotationInProgress)
}

func TestKeyManager_BackwardCompatibleDecrypt(t *testing.T) {
	km := NewKeyManager()

	// Generate initial key
	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	key1, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	key1ID := key1.ID

	// Rotate to key 2
	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	// Rotate to key 3
	err = km.CompleteRotation(ScopeVEID)
	require.NoError(t, err)
	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	// Verify we can still get key1 for decryption
	key1Retrieved, err := km.GetKey(ScopeVEID, key1ID)
	require.NoError(t, err)
	require.Equal(t, key1ID, key1Retrieved.ID)
	require.Equal(t, KeyStatusRetired, key1Retrieved.Status)

	// Verify private key is still available for decryption
	require.NotEmpty(t, key1Retrieved.PrivateKey)
}

func TestKeyManager_ListKeys(t *testing.T) {
	km := NewKeyManager()

	// Generate multiple keys through rotation
	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	err = km.CompleteRotation(ScopeVEID)
	require.NoError(t, err)

	err = km.RotateKey(ScopeVEID, 1*time.Hour)
	require.NoError(t, err)

	// List all keys
	keys, err := km.ListKeys(ScopeVEID)
	require.NoError(t, err)
	require.Len(t, keys, 3)

	// Verify statuses
	var activeCount, deprecatedCount, retiredCount int
	for _, key := range keys {
		switch key.Status {
		case KeyStatusActive:
			activeCount++
		case KeyStatusDeprecated:
			deprecatedCount++
		case KeyStatusRetired:
			retiredCount++
		}
	}

	require.Equal(t, 1, activeCount, "should have 1 active key")
	require.Equal(t, 1, deprecatedCount, "should have 1 deprecated key")
	require.Equal(t, 1, retiredCount, "should have 1 retired key")
}

func TestKeyManager_RotationPolicy(t *testing.T) {
	km := NewKeyManager()

	policy := &RotationPolicy{
		MaxAge:           90 * 24 * time.Hour, // 90 days
		MaxVersions:      5,
		AutoRotate:       true,
		RotationSchedule: "0 0 * * 0", // Weekly on Sunday
	}

	km.SetRotationPolicy(ScopeVEID, policy)

	retrieved := km.GetRotationPolicy(ScopeVEID)
	require.NotNil(t, retrieved)
	require.Equal(t, policy.MaxAge, retrieved.MaxAge)
	require.Equal(t, policy.MaxVersions, retrieved.MaxVersions)
	require.Equal(t, policy.AutoRotate, retrieved.AutoRotate)
	require.Equal(t, policy.RotationSchedule, retrieved.RotationSchedule)
}

func TestKeyManager_RevokeKey(t *testing.T) {
	km := NewKeyManager()

	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	key, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)

	err = km.RevokeKey(ScopeVEID, key.ID)
	require.NoError(t, err)

	revoked, err := km.GetKey(ScopeVEID, key.ID)
	require.NoError(t, err)
	require.Equal(t, KeyStatusRevoked, revoked.Status)
	require.NotNil(t, revoked.RevokedAt)
}

func TestKeyManager_EmergencyRotateKey(t *testing.T) {
	km := NewKeyManager()

	err := km.GenerateKey(ScopeVEID)
	require.NoError(t, err)

	oldKey, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)

	newKey, err := km.EmergencyRotateKey(ScopeVEID)
	require.NoError(t, err)
	require.NotNil(t, newKey)
	require.NotEqual(t, oldKey.ID, newKey.ID)

	oldKeyUpdated, err := km.GetKey(ScopeVEID, oldKey.ID)
	require.NoError(t, err)
	require.Equal(t, KeyStatusRevoked, oldKeyUpdated.Status)

	active, err := km.GetActiveKey(ScopeVEID)
	require.NoError(t, err)
	require.Equal(t, newKey.ID, active.ID)
}
