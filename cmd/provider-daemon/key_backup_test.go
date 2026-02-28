package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

func TestWriteAndRestoreProviderKeyBackupRoundTrip(t *testing.T) {
	t.Helper()

	config := provider_daemon.DefaultKeyManagerConfig()
	config.StorageType = provider_daemon.KeyStorageTypeMemory

	sourceKM, err := provider_daemon.NewKeyManager(config)
	require.NoError(t, err)
	require.NoError(t, sourceKM.Unlock(""))

	originalKey, err := sourceKM.GenerateKey("provider-primary")
	require.NoError(t, err)

	backupPath := t.TempDir() + "/provider-backup.json"
	_, err = writeProviderKeyBackup(sourceKM, backupPath, "backup-passphrase")
	require.NoError(t, err)

	restoreKM, err := provider_daemon.NewKeyManager(config)
	require.NoError(t, err)
	require.NoError(t, restoreKM.Unlock(""))

	result, err := restoreProviderKeysFromBackup(restoreKM, backupPath, "backup-passphrase")
	require.NoError(t, err)
	require.Len(t, result.RestoredKeys, 1)

	restoredKey, generated, err := ensureProviderKey(restoreKM, "provider-primary")
	require.NoError(t, err)
	require.False(t, generated)
	require.Equal(t, originalKey.KeyID, restoredKey.KeyID)

	message := []byte("provider-backup-cli-roundtrip")
	signature, err := restoreKM.SignWithKey(restoredKey.KeyID, message)
	require.NoError(t, err)
	require.Equal(t, originalKey.PublicKey, signature.PublicKey)
	require.NoError(t, signature.Verify(message))
}
