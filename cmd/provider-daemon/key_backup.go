package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

func restoreProviderKeysFromBackup(keyManager *provider_daemon.KeyManager, backupPath, passphrase string) (*provider_daemon.RestoreResult, error) {
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if backupPath == "" {
		return nil, fmt.Errorf("backup path is required")
	}
	if passphrase == "" {
		return nil, fmt.Errorf("backup passphrase is required")
	}

	backup, err := readKeyBackup(backupPath)
	if err != nil {
		return nil, err
	}

	manager := provider_daemon.NewKeyBackupManager(nil, keyManager)
	result, err := manager.RestoreBackup(backup, passphrase)
	if err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("restore completed with errors: %v", result.Errors)
	}
	if len(result.RestoredKeys) == 0 && len(result.SkippedKeys) == 0 {
		return nil, fmt.Errorf("backup %s did not restore any keys", backupPath)
	}

	return result, nil
}

func writeProviderKeyBackup(keyManager *provider_daemon.KeyManager, backupPath, passphrase string) (*provider_daemon.KeyBackup, error) {
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if backupPath == "" {
		return nil, fmt.Errorf("backup path is required")
	}
	if passphrase == "" {
		return nil, fmt.Errorf("backup passphrase is required")
	}

	manager := provider_daemon.NewKeyBackupManager(nil, keyManager)
	backup, err := manager.CreateBackup(passphrase)
	if err != nil {
		return nil, err
	}
	if err := writeKeyBackupFile(backupPath, backup); err != nil {
		return nil, err
	}

	return backup, nil
}

func ensureProviderKey(keyManager *provider_daemon.KeyManager, providerKeyName string) (*provider_daemon.ManagedKey, bool, error) {
	if keyManager == nil {
		return nil, false, fmt.Errorf("key manager is required")
	}

	key, err := keyManager.GetActiveKey()
	switch {
	case err == nil:
		return key, false, nil
	case errors.Is(err, provider_daemon.ErrKeyNotFound):
		if providerKeyName == "" {
			return nil, false, fmt.Errorf("provider key name is required")
		}
		key, err = keyManager.GenerateKey(providerKeyName)
		if err != nil {
			return nil, false, fmt.Errorf("generate provider key: %w", err)
		}
		return key, true, nil
	default:
		return nil, false, fmt.Errorf("load active key: %w", err)
	}
}

func readKeyBackup(path string) (*provider_daemon.KeyBackup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key backup: %w", err)
	}

	var backup provider_daemon.KeyBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("decode key backup: %w", err)
	}

	return &backup, nil
}

func writeKeyBackupFile(path string, backup *provider_daemon.KeyBackup) error {
	if backup == nil {
		return fmt.Errorf("backup is required")
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("encode key backup: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write key backup: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace key backup: %w", err)
	}

	return nil
}
