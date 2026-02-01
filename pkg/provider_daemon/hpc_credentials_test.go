// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-14B: HPC Credential Manager tests
package provider_daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHPCCredentials_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "zero expiry",
			expiresAt: time.Time{},
			want:      false,
		},
		{
			name:      "future expiry",
			expiresAt: time.Now().Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "past expiry",
			expiresAt: time.Now().Add(-24 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &HPCCredentials{
				ExpiresAt: tt.expiresAt,
			}
			if got := creds.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHPCCredentialManager_StoreAndGet(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = "" // No persistence for test

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	// Unlock without passphrase (allowed because AllowUnencrypted)
	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()
	creds := &HPCCredentials{
		Type:              CredentialTypeSLURM,
		ClusterID:         "test-cluster",
		Username:          "testuser",
		Password:          "secret-password",
		SSHPrivateKeyPath: "/path/to/key",
	}

	// Store credentials
	if err := cm.StoreCredentials(ctx, creds); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Retrieve credentials
	retrieved, err := cm.GetCredentials(ctx, "test-cluster", CredentialTypeSLURM)
	if err != nil {
		t.Fatalf("GetCredentials() error = %v", err)
	}

	if retrieved.Username != creds.Username {
		t.Errorf("Username = %v, want %v", retrieved.Username, creds.Username)
	}
	if retrieved.Password != creds.Password {
		t.Errorf("Password not preserved correctly")
	}
	if retrieved.SSHPrivateKeyPath != creds.SSHPrivateKeyPath {
		t.Errorf("SSHPrivateKeyPath = %v, want %v", retrieved.SSHPrivateKeyPath, creds.SSHPrivateKeyPath)
	}
}

func TestHPCCredentialManager_Locking(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	// Should be locked by default
	if !cm.IsLocked() {
		t.Error("Should be locked by default")
	}

	// Store should fail when locked
	ctx := context.Background()
	creds := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "test-cluster",
	}
	if err := cm.StoreCredentials(ctx, creds); err == nil {
		t.Error("StoreCredentials() should fail when locked")
	}

	// Unlock
	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	if cm.IsLocked() {
		t.Error("Should be unlocked after Unlock()")
	}

	// Store should work now
	if err := cm.StoreCredentials(ctx, creds); err != nil {
		t.Errorf("StoreCredentials() error = %v", err)
	}

	// Lock
	cm.Lock()

	if !cm.IsLocked() {
		t.Error("Should be locked after Lock()")
	}

	// Get should fail when locked
	if _, err := cm.GetCredentials(ctx, "test-cluster", CredentialTypeSLURM); err == nil {
		t.Error("GetCredentials() should fail when locked")
	}
}

func TestHPCCredentialManager_DeleteCredentials(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()
	creds := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "test-cluster",
		Username:  "testuser",
	}

	// Store
	if err := cm.StoreCredentials(ctx, creds); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Verify it exists
	if _, err := cm.GetCredentials(ctx, "test-cluster", CredentialTypeSLURM); err != nil {
		t.Fatalf("GetCredentials() error = %v", err)
	}

	// Delete
	if err := cm.DeleteCredentials(ctx, "test-cluster", CredentialTypeSLURM); err != nil {
		t.Fatalf("DeleteCredentials() error = %v", err)
	}

	// Verify it's gone
	if _, err := cm.GetCredentials(ctx, "test-cluster", CredentialTypeSLURM); err == nil {
		t.Error("GetCredentials() should fail after deletion")
	}
}

func TestHPCCredentialManager_ListClusters(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()

	// Store credentials for multiple clusters
	clusters := []string{"cluster-1", "cluster-2", "cluster-3"}
	for _, clusterID := range clusters {
		creds := &HPCCredentials{
			Type:      CredentialTypeSLURM,
			ClusterID: clusterID,
			Username:  "user",
		}
		if err := cm.StoreCredentials(ctx, creds); err != nil {
			t.Fatalf("StoreCredentials() error = %v", err)
		}
	}

	// List clusters
	listed := cm.ListClusters()
	if len(listed) != len(clusters) {
		t.Errorf("ListClusters() returned %d clusters, want %d", len(listed), len(clusters))
	}

	// Verify all clusters are present
	for _, cluster := range clusters {
		found := false
		for _, c := range listed {
			if c == cluster {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Cluster %s not found in listed clusters", cluster)
		}
	}
}

func TestHPCCredentialManager_SigningKey(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	// Generate signing key
	if err := cm.GenerateSigningKey(); err != nil {
		t.Fatalf("GenerateSigningKey() error = %v", err)
	}

	// Get public key
	pubKey, err := cm.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}
	if len(pubKey) == 0 {
		t.Error("Public key should not be empty")
	}

	// Sign data
	message := []byte("test message to sign")
	signature, err := cm.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if len(signature) == 0 {
		t.Error("Signature should not be empty")
	}

	// Verify signature
	if !cm.Verify(message, signature) {
		t.Error("Verify() should return true for valid signature")
	}

	// Verify with wrong message should fail
	if cm.Verify([]byte("wrong message"), signature) {
		t.Error("Verify() should return false for wrong message")
	}
}

func TestHPCCredentialManager_Persistence(t *testing.T) {
	// Create temp directory for persistence
	tempDir, err := os.MkdirTemp("", "hpc-cred-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	passphrase := "test-passphrase-123"

	// Create and populate credential manager
	config := DefaultHPCCredentialManagerConfig()
	config.StorageDir = tempDir

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(passphrase); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()
	creds := &HPCCredentials{
		Type:              CredentialTypeSLURM,
		ClusterID:         "persist-cluster",
		Username:          "testuser",
		Password:          "secret-password",
		SSHPrivateKey:     "-----BEGIN OPENSSH PRIVATE KEY-----\nfake-key-content\n-----END OPENSSH PRIVATE KEY-----",
		SSHPrivateKeyPath: "/path/to/key",
	}

	// Save expected values before storing (Lock() scrubs the stored pointer)
	expectedUsername := creds.Username
	expectedPassword := creds.Password
	expectedSSHPrivateKey := creds.SSHPrivateKey

	if err := cm.StoreCredentials(ctx, creds); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Generate signing key
	if err := cm.GenerateSigningKey(); err != nil {
		t.Fatalf("GenerateSigningKey() error = %v", err)
	}

	originalPubKey, _ := cm.GetPublicKey()

	// Lock and close
	cm.Lock()

	// Verify file was created
	credFile := filepath.Join(tempDir, "credentials.json")
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		t.Error("Credentials file should exist after persistence")
	}

	// Create new manager and load
	cm2, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm2.Unlock(passphrase); err != nil {
		t.Fatalf("Unlock() on new manager error = %v", err)
	}

	// Verify credentials were loaded
	loaded, err := cm2.GetCredentials(ctx, "persist-cluster", CredentialTypeSLURM)
	if err != nil {
		t.Fatalf("GetCredentials() on loaded manager error = %v", err)
	}

	if loaded.Username != expectedUsername {
		t.Errorf("Loaded Username = %v, want %v", loaded.Username, expectedUsername)
	}
	if loaded.Password != expectedPassword {
		t.Errorf("Loaded Password = %q, want %q", loaded.Password, expectedPassword)
	}
	if loaded.SSHPrivateKey != expectedSSHPrivateKey {
		t.Errorf("Loaded SSHPrivateKey = %q, want %q", loaded.SSHPrivateKey, expectedSSHPrivateKey)
	}

	// Verify signing key was loaded
	loadedPubKey, err := cm2.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() on loaded manager error = %v", err)
	}
	if string(loadedPubKey) != string(originalPubKey) {
		t.Error("Loaded public key doesn't match original")
	}
}

func TestHPCCredentialManager_RotationWarnings(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""
	config.RotationWarningDays = 30

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()

	// Credential that expires soon
	soonExpiring := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "expiring-cluster",
		Username:  "user",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}
	if err := cm.StoreCredentials(ctx, soonExpiring); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Credential that doesn't expire soon
	notExpiring := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "valid-cluster",
		Username:  "user",
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour), // 90 days
	}
	if err := cm.StoreCredentials(ctx, notExpiring); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	warnings := cm.CheckRotationWarnings()

	// Should have warning for expiring-cluster only
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 && warnings[0].ClusterID != "expiring-cluster" {
		t.Errorf("Warning ClusterID = %v, want expiring-cluster", warnings[0].ClusterID)
	}
}

func TestHPCCredentialManager_HealthCheck(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()
	config.AllowUnencrypted = true
	config.StorageDir = ""

	cm, err := NewHPCCredentialManager(config)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	if err := cm.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	ctx := context.Background()

	// Valid credential
	valid := &HPCCredentials{
		Type:              CredentialTypeSLURM,
		ClusterID:         "valid-cluster",
		Username:          "user",
		SSHPrivateKeyPath: "/path/to/key",
	}
	if err := cm.StoreCredentials(ctx, valid); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Expired credential
	expired := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "expired-cluster",
		Username:  "user",
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	if err := cm.StoreCredentials(ctx, expired); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	// Invalid credential (no auth)
	invalid := &HPCCredentials{
		Type:      CredentialTypeSLURM,
		ClusterID: "invalid-cluster",
		Username:  "user",
		// No password or SSH key
	}
	if err := cm.StoreCredentials(ctx, invalid); err != nil {
		t.Fatalf("StoreCredentials() error = %v", err)
	}

	health := cm.CheckHealth()

	// Should have 3 entries
	if len(health) != 3 {
		t.Errorf("Expected 3 health entries, got %d", len(health))
	}

	// Count valid and invalid
	validCount := 0
	invalidCount := 0
	for _, h := range health {
		if h.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	if validCount != 1 {
		t.Errorf("Expected 1 valid credential, got %d", validCount)
	}
	if invalidCount != 2 {
		t.Errorf("Expected 2 invalid credentials, got %d", invalidCount)
	}
}

func TestCredentialTypes(t *testing.T) {
	types := []CredentialType{
		CredentialTypeSLURM,
		CredentialTypeMOAB,
		CredentialTypeOOD,
		CredentialTypeKerberos,
		CredentialTypeSigning,
	}

	for _, ct := range types {
		if ct == "" {
			t.Error("CredentialType should not be empty")
		}
	}
}

func TestDefaultHPCCredentialManagerConfig(t *testing.T) {
	config := DefaultHPCCredentialManagerConfig()

	if config.StorageDir == "" {
		t.Error("Default StorageDir should not be empty")
	}
	if config.RotationCheckInterval <= 0 {
		t.Error("Default RotationCheckInterval should be positive")
	}
	if config.RotationWarningDays <= 0 {
		t.Error("Default RotationWarningDays should be positive")
	}
}

