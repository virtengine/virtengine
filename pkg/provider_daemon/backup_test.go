package provider_daemon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyBackupManager_RestorePreservesSigningKeys(t *testing.T) {
	t.Helper()

	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory

	sourceKM, err := NewKeyManager(config)
	require.NoError(t, err)
	require.NoError(t, sourceKM.Unlock(""))

	firstKey, err := sourceKM.GenerateKey("provider-1")
	require.NoError(t, err)
	secondKey, err := sourceKM.GenerateKey("provider-2")
	require.NoError(t, err)

	backupManager := NewKeyBackupManager(DefaultKeyBackupConfig(), sourceKM)
	backup, err := backupManager.CreateBackup("backup-passphrase")
	require.NoError(t, err)

	restoreKM, err := NewKeyManager(config)
	require.NoError(t, err)
	require.NoError(t, restoreKM.Unlock(""))

	restoreManager := NewKeyBackupManager(DefaultKeyBackupConfig(), restoreKM)
	result, err := restoreManager.RestoreBackup(backup, "backup-passphrase")
	require.NoError(t, err)
	require.Len(t, result.RestoredKeys, 2)
	require.Empty(t, result.Errors)

	message := []byte("provider-backup-roundtrip")
	for _, original := range []*ManagedKey{firstKey, secondKey} {
		signature, signErr := restoreKM.SignWithKey(original.KeyID, message)
		require.NoError(t, signErr)
		require.Equal(t, original.PublicKey, signature.PublicKey)
		require.NoError(t, signature.Verify(message))
	}
}
