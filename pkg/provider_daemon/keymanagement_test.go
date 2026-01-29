package provider_daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HSM Tests
// =============================================================================

func TestSoftHSMProvider_Initialize(t *testing.T) {
	provider := NewSoftHSMProvider()

	assert.False(t, provider.IsInitialized())

	err := provider.Initialize(nil)
	require.NoError(t, err)
	assert.True(t, provider.IsInitialized())

	err = provider.Close()
	require.NoError(t, err)
	assert.False(t, provider.IsInitialized())
}

func TestSoftHSMProvider_GetInfo(t *testing.T) {
	provider := NewSoftHSMProvider()
	err := provider.Initialize(nil)
	require.NoError(t, err)
	defer provider.Close()

	info, err := provider.GetInfo()
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, HSMTypeSoftHSM, info.Type)
	assert.NotEmpty(t, info.ManufacturerID)
	assert.NotEmpty(t, info.SupportedMechanisms)
}

func TestSoftHSMProvider_GenerateKey(t *testing.T) {
	provider := NewSoftHSMProvider()
	err := provider.Initialize(nil)
	require.NoError(t, err)
	defer provider.Close()

	t.Run("generate ed25519 key", func(t *testing.T) {
		handle, err := provider.GenerateKey("test-key-1", HSMKeyTypeEd25519)
		require.NoError(t, err)
		require.NotNil(t, handle)

		assert.Equal(t, "test-key-1", handle.Label)
		assert.Equal(t, HSMKeyTypeEd25519, handle.KeyType)
		assert.NotEmpty(t, handle.ID)
		assert.NotEmpty(t, handle.PublicKeyFingerprint)
		assert.True(t, handle.Usage.Sign)
		assert.True(t, handle.Usage.Verify)
	})

	t.Run("generate p256 key", func(t *testing.T) {
		handle, err := provider.GenerateKey("test-key-2", HSMKeyTypeP256)
		require.NoError(t, err)
		require.NotNil(t, handle)

		assert.Equal(t, HSMKeyTypeP256, handle.KeyType)
	})

	t.Run("duplicate key label fails", func(t *testing.T) {
		_, err := provider.GenerateKey("test-key-1", HSMKeyTypeEd25519)
		require.Error(t, err)
	})
}

func TestSoftHSMProvider_SignAndVerify(t *testing.T) {
	provider := NewSoftHSMProvider()
	err := provider.Initialize(nil)
	require.NoError(t, err)
	defer provider.Close()

	handle, err := provider.GenerateKey("signing-key", HSMKeyTypeEd25519)
	require.NoError(t, err)

	message := []byte("test message for signing")

	signature, err := provider.Sign(handle, message)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	valid, err := provider.Verify(handle, message, signature)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestSoftHSMProvider_ListAndDeleteKeys(t *testing.T) {
	provider := NewSoftHSMProvider()
	err := provider.Initialize(nil)
	require.NoError(t, err)
	defer provider.Close()

	// Generate multiple keys
	_, err = provider.GenerateKey("key-1", HSMKeyTypeEd25519)
	require.NoError(t, err)
	_, err = provider.GenerateKey("key-2", HSMKeyTypeP256)
	require.NoError(t, err)

	// List keys
	keys, err := provider.ListKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	// Delete a key
	err = provider.DeleteKey("key-1")
	require.NoError(t, err)

	// Verify deleted
	keys, err = provider.ListKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	// Try to get deleted key
	_, err = provider.GetKey("key-1")
	require.ErrorIs(t, err, ErrHSMKeyNotFound)
}

func TestSoftHSMProvider_ImportKey(t *testing.T) {
	provider := NewSoftHSMProvider()
	err := provider.Initialize(nil)
	require.NoError(t, err)
	defer provider.Close()

	// Create a test private key
	privateKey := make([]byte, 64)
	for i := range privateKey {
		privateKey[i] = byte(i)
	}

	handle, err := provider.ImportKey("imported-key", privateKey, HSMKeyTypeEd25519)
	require.NoError(t, err)
	require.NotNil(t, handle)

	assert.Equal(t, "imported-key", handle.Label)
	assert.Equal(t, HSMKeyTypeEd25519, handle.KeyType)
}

// =============================================================================
// Backup Tests
// =============================================================================

func TestKeyBackupManager_CreateAndRestoreBackup(t *testing.T) {
	// Create a key manager
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(config)
	require.NoError(t, err)
	err = km.Unlock("")
	require.NoError(t, err)

	// Generate some keys
	_, err = km.GenerateKey("provider1")
	require.NoError(t, err)
	_, err = km.GenerateKey("provider2")
	require.NoError(t, err)

	// Create backup manager
	backupConfig := DefaultKeyBackupConfig()
	bm := NewKeyBackupManager(backupConfig, km)

	// Create backup
	passphrase := "secure-backup-passphrase"
	backup, err := bm.CreateBackup(passphrase)
	require.NoError(t, err)
	require.NotNil(t, backup)

	assert.Equal(t, BackupVersion, backup.Version)
	assert.NotEmpty(t, backup.Ciphertext)
	assert.NotEmpty(t, backup.Salt)
	assert.NotEmpty(t, backup.Nonce)
	assert.NotEmpty(t, backup.Checksum)

	if backup.Metadata != nil {
		assert.Equal(t, 2, backup.Metadata.KeyCount)
	}
}

func TestKeyBackupManager_BackupWithWrongPassphrase(t *testing.T) {
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(config)
	require.NoError(t, err)
	err = km.Unlock("")
	require.NoError(t, err)

	_, err = km.GenerateKey("provider1")
	require.NoError(t, err)

	bm := NewKeyBackupManager(nil, km)

	// Create backup
	backup, err := bm.CreateBackup("correct-passphrase")
	require.NoError(t, err)

	// Try to restore with wrong passphrase
	_, err = bm.RestoreBackup(backup, "wrong-passphrase")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBackupDecryptionFailed)
}

func TestSecureBackupWithSecretBox(t *testing.T) {
	data := []byte("sensitive key data to encrypt")
	passphrase := "strong-passphrase"

	encrypted, err := SecureBackupWithSecretBox(data, passphrase)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := DecryptSecureBackup(encrypted, passphrase)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)

	// Wrong passphrase should fail
	_, err = DecryptSecureBackup(encrypted, "wrong-passphrase")
	require.Error(t, err)
}

// =============================================================================
// Multi-Signature Tests
// =============================================================================

func TestMultiSigManager_CreateMultiSigKey(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	msm := NewMultiSigManager(km)

	config := &MultiSigConfig{
		Threshold:       2,
		TotalSigners:    3,
		TimeoutDuration: 24 * time.Hour,
	}

	signers := []AuthorizedSigner{
		{PublicKey: "pubkey1", Label: "Signer 1", Weight: 1},
		{PublicKey: "pubkey2", Label: "Signer 2", Weight: 1},
		{PublicKey: "pubkey3", Label: "Signer 3", Weight: 1},
	}

	key, err := msm.CreateMultiSigKey(config, signers, "Test multisig")
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.NotEmpty(t, key.ID)
	assert.Equal(t, 3, len(key.AuthorizedSigners))
	assert.Equal(t, 2, key.Config.Threshold)
}

func TestMultiSigManager_InitiateAndCompleteOperation(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	msm := NewMultiSigManager(km)

	config := &MultiSigConfig{
		Threshold:       2,
		TotalSigners:    3,
		TimeoutDuration: 1 * time.Hour,
	}

	signers := []AuthorizedSigner{
		{PublicKey: "pubkey1", Label: "Signer 1", Weight: 1},
		{PublicKey: "pubkey2", Label: "Signer 2", Weight: 1},
		{PublicKey: "pubkey3", Label: "Signer 3", Weight: 1},
	}

	key, err := msm.CreateMultiSigKey(config, signers, "Test")
	require.NoError(t, err)

	// Initiate operation
	message := []byte("transaction to sign")
	op, err := msm.InitiateOperation(key.ID, message, "initiator", "Test transaction")
	require.NoError(t, err)
	require.NotNil(t, op)

	assert.Equal(t, MultiSigStatusPending, op.Status)

	// Add first signature
	err = msm.AddSignature(op.ID, "pubkey1", []byte("signature1"))
	require.NoError(t, err)

	// Should still be pending
	op, _ = msm.GetOperation(op.ID)
	assert.Equal(t, MultiSigStatusPending, op.Status)

	// Add second signature
	err = msm.AddSignature(op.ID, "pubkey2", []byte("signature2"))
	require.NoError(t, err)

	// Should now meet threshold
	op, _ = msm.GetOperation(op.ID)
	assert.Equal(t, MultiSigStatusThresholdMet, op.Status)

	// Complete operation
	completedOp, err := msm.CompleteOperation(op.ID)
	require.NoError(t, err)
	assert.Equal(t, MultiSigStatusComplete, completedOp.Status)
	assert.NotEmpty(t, completedOp.FinalSignature)
}

func TestMultiSigManager_DuplicateSignature(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	msm := NewMultiSigManager(km)

	config := DefaultMultiSigConfig()
	signers := []AuthorizedSigner{
		{PublicKey: "pubkey1", Label: "Signer 1", Weight: 1},
		{PublicKey: "pubkey2", Label: "Signer 2", Weight: 1},
		{PublicKey: "pubkey3", Label: "Signer 3", Weight: 1},
	}

	key, _ := msm.CreateMultiSigKey(config, signers, "Test")
	op, _ := msm.InitiateOperation(key.ID, []byte("msg"), "init", "desc")

	// First signature succeeds
	err := msm.AddSignature(op.ID, "pubkey1", []byte("sig"))
	require.NoError(t, err)

	// Duplicate signature fails
	err = msm.AddSignature(op.ID, "pubkey1", []byte("sig2"))
	require.ErrorIs(t, err, ErrDuplicateSignature)
}

func TestMultiSigManager_UnauthorizedSigner(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	msm := NewMultiSigManager(km)

	config := DefaultMultiSigConfig()
	signers := []AuthorizedSigner{
		{PublicKey: "pubkey1", Label: "Signer 1", Weight: 1},
		{PublicKey: "pubkey2", Label: "Signer 2", Weight: 1},
		{PublicKey: "pubkey3", Label: "Signer 3", Weight: 1},
	}

	key, _ := msm.CreateMultiSigKey(config, signers, "Test")
	op, _ := msm.InitiateOperation(key.ID, []byte("msg"), "init", "desc")

	// Unauthorized signer fails
	err := msm.AddSignature(op.ID, "unauthorized-key", []byte("sig"))
	require.ErrorIs(t, err, ErrSignerNotAuthorized)
}

// =============================================================================
// Compromise Detection Tests
// =============================================================================

func TestCompromiseDetector_RecordKeyUsage(t *testing.T) {
	config := DefaultCompromiseDetectorConfig()
	config.UsageThresholdPerMinute = 3
	detector := NewCompromiseDetector(config, nil)

	keyID := "test-key"
	now := time.Now()

	// First few uses should be fine
	indicators := detector.RecordKeyUsage(keyID, "192.168.1.1", now)
	assert.Empty(t, indicators)

	indicators = detector.RecordKeyUsage(keyID, "192.168.1.1", now)
	assert.Empty(t, indicators)

	indicators = detector.RecordKeyUsage(keyID, "192.168.1.1", now)
	assert.Empty(t, indicators)

	// Fourth use should trigger rapid usage
	indicators = detector.RecordKeyUsage(keyID, "192.168.1.1", now)
	assert.Contains(t, indicators, IndicatorRapidUsage)
}

func TestCompromiseDetector_AnomalousTime(t *testing.T) {
	config := DefaultCompromiseDetectorConfig()
	config.AnomalousTimeWindowStart = 9  // 9 AM
	config.AnomalousTimeWindowEnd = 17   // 5 PM
	detector := NewCompromiseDetector(config, nil)

	keyID := "test-key"

	// Use at 3 AM (outside window)
	earlyMorning := time.Date(2024, 1, 15, 3, 0, 0, 0, time.UTC)
	indicators := detector.RecordKeyUsage(keyID, "192.168.1.1", earlyMorning)
	assert.Contains(t, indicators, IndicatorAnomalousTime)
}

func TestCompromiseDetector_ReportCompromise(t *testing.T) {
	detector := NewCompromiseDetector(nil, nil)

	err := detector.ReportCompromise(
		"compromised-key",
		IndicatorExternalReport,
		SeverityHigh,
		"Key leaked on public forum",
		"security-team",
	)
	require.NoError(t, err)

	events := detector.GetEventsByKey("compromised-key")
	require.Len(t, events, 1)
	assert.Equal(t, IndicatorExternalReport, events[0].Indicator)
	assert.Equal(t, SeverityHigh, events[0].Severity)
}

func TestCompromiseDetector_IsKeyCompromised(t *testing.T) {
	detector := NewCompromiseDetector(nil, nil)

	// Initially not compromised
	assert.False(t, detector.IsKeyCompromised("test-key"))

	// Report a high severity compromise
	detector.ReportCompromise("test-key", IndicatorKeyLeakage, SeverityHigh, "Leaked", "admin")
	assert.True(t, detector.IsKeyCompromised("test-key"))

	// Acknowledge the event
	events := detector.GetEventsByKey("test-key")
	require.Len(t, events, 1)
	err := detector.AcknowledgeEvent(events[0].ID, "admin")
	require.NoError(t, err)

	// No longer considered compromised after acknowledgement
	assert.False(t, detector.IsKeyCompromised("test-key"))
}

// =============================================================================
// Lifecycle Management Tests
// =============================================================================

func TestKeyLifecycleManager_RegisterAndActivateKey(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	lm := NewKeyLifecycleManager(km)

	record, err := lm.RegisterKey("key-1", "ed25519", "fingerprint-1", "default")
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, KeyStateCreated, record.CurrentState)
	assert.Len(t, record.StateHistory, 1)

	// Activate the key
	err = lm.ActivateKey("key-1", "admin")
	require.NoError(t, err)

	record, _ = lm.GetRecord("key-1")
	assert.Equal(t, KeyStateActive, record.CurrentState)
	assert.NotNil(t, record.ActivatedAt)
}

func TestKeyLifecycleManager_InvalidTransition(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	lm := NewKeyLifecycleManager(km)

	lm.RegisterKey("key-1", "ed25519", "fp-1", "default")

	// Cannot go directly from Created to Expired
	err := lm.TransitionState("key-1", KeyStateExpired, "admin", "test")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyLifecycleInvalidTransition)
}

func TestKeyLifecycleManager_RotateKey(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	lm := NewKeyLifecycleManager(km)

	// Create and activate first key
	lm.RegisterKey("key-1", "ed25519", "fp-1", "default")
	lm.ActivateKey("key-1", "admin")

	// Create second key
	lm.RegisterKey("key-2", "ed25519", "fp-2", "default")
	lm.ActivateKey("key-2", "admin")

	// Rotate first key to second
	err := lm.RotateKey("key-1", "key-2", "admin")
	require.NoError(t, err)

	record, _ := lm.GetRecord("key-1")
	assert.Equal(t, KeyStateRotating, record.CurrentState)
	assert.Equal(t, "key-2", record.SuccessorKeyID)

	// Complete rotation
	err = lm.CompleteRotation("key-1", "admin")
	require.NoError(t, err)

	record, _ = lm.GetRecord("key-1")
	assert.Equal(t, KeyStateDeactivated, record.CurrentState)
}

func TestKeyLifecycleManager_GetKeysNeedingRotation(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	lm := NewKeyLifecycleManager(km)

	// Create a policy with very short rotation period
	policy := &KeyLifecyclePolicy{
		Name:             "short-rotation",
		MaxActiveAgeDays: 0, // Immediate rotation needed
		ExpirationDays:   365,
	}
	lm.RegisterPolicy(policy)

	// Create and activate a key with this policy
	lm.RegisterKey("old-key", "ed25519", "fp-1", "short-rotation")
	lm.ActivateKey("old-key", "admin")

	// Should need rotation immediately
	keys := lm.GetKeysNeedingRotation()
	assert.Len(t, keys, 1)
	assert.Equal(t, "old-key", keys[0].KeyID)
}

func TestKeyLifecycleManager_GenerateReport(t *testing.T) {
	km, _ := NewKeyManager(DefaultKeyManagerConfig())
	lm := NewKeyLifecycleManager(km)

	// Create various keys in different states
	lm.RegisterKey("key-1", "ed25519", "fp-1", "default")
	lm.ActivateKey("key-1", "admin")

	lm.RegisterKey("key-2", "p256", "fp-2", "default")
	lm.ActivateKey("key-2", "admin")
	lm.SuspendKey("key-2", "admin", "maintenance")

	report := lm.GenerateLifecycleReport()
	require.NotNil(t, report)

	assert.Equal(t, 2, report.TotalKeys)
	assert.Equal(t, 1, report.ActiveKeys)
	assert.Equal(t, 1, report.ByState[string(KeyStateSuspended)])
}

// =============================================================================
// Access Control and Audit Tests
// =============================================================================

func TestAccessController_CreateSessionAndCheckPermission(t *testing.T) {
	ac := NewAccessController(nil)

	// Create a principal
	principal := &Principal{
		ID:    "user-1",
		Type:  "user",
		Name:  "Test User",
		Roles: []string{"operator"},
	}
	err := ac.CreatePrincipal(principal)
	require.NoError(t, err)

	// Create session
	session, err := ac.CreateSession("user-1", "192.168.1.1", "test-agent")
	require.NoError(t, err)
	require.NotNil(t, session)

	// Check allowed permission
	err = ac.CheckPermission(session.ID, PermissionKeySign)
	require.NoError(t, err)

	// Check denied permission
	err = ac.CheckPermission(session.ID, PermissionPolicyManage)
	require.ErrorIs(t, err, ErrInsufficientPermissions)
}

func TestAccessController_AdminHasAllPermissions(t *testing.T) {
	ac := NewAccessController(nil)

	principal := &Principal{
		ID:    "admin-1",
		Type:  "user",
		Name:  "Admin User",
		Roles: []string{"admin"},
	}
	err := ac.CreatePrincipal(principal)
	require.NoError(t, err)

	session, _ := ac.CreateSession("admin-1", "192.168.1.1", "")

	// Admin should have all permissions
	for _, perm := range AllPermissions() {
		err = ac.CheckPermission(session.ID, perm)
		assert.NoError(t, err, "Admin should have permission: %s", perm)
	}
}

func TestAccessController_SessionExpiration(t *testing.T) {
	config := &AccessControlConfig{
		SessionTimeoutMinutes: 0, // Expires immediately
	}
	ac := NewAccessController(config)

	principal := &Principal{
		ID:    "user-1",
		Type:  "user",
		Name:  "Test User",
		Roles: []string{"operator"},
	}
	ac.CreatePrincipal(principal)

	session, _ := ac.CreateSession("user-1", "", "")

	// Wait a moment for expiration
	time.Sleep(10 * time.Millisecond)

	_, err := ac.ValidateSession(session.ID)
	require.ErrorIs(t, err, ErrSessionExpired)
}

func TestAuditLogger_LogAndRetrieveEvents(t *testing.T) {
	config := DefaultAuditLogConfig()
	config.LogFile = "" // In-memory only
	logger, err := NewAuditLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Log some events
	err = logger.LogKeyOperation(
		AuditEventKeyCreated,
		"session-1",
		"user-1",
		"Test User",
		"key-1",
		"generate_key",
		true,
		"",
		map[string]interface{}{"algorithm": "ed25519"},
	)
	require.NoError(t, err)

	err = logger.LogKeyOperation(
		AuditEventKeySigned,
		"session-1",
		"user-1",
		"Test User",
		"key-1",
		"sign_message",
		true,
		"",
		nil,
	)
	require.NoError(t, err)

	// Retrieve events
	since := time.Now().Add(-1 * time.Hour)
	events := logger.GetEvents(since, nil)
	assert.Len(t, events, 2)

	// Filter by type
	eventType := AuditEventKeyCreated
	events = logger.GetEvents(since, &eventType)
	assert.Len(t, events, 1)
}

func TestAuditLogger_HashChaining(t *testing.T) {
	config := DefaultAuditLogConfig()
	config.LogFile = ""
	config.EnableChaining = true
	logger, _ := NewAuditLogger(config)
	defer logger.Close()

	// Log multiple events
	for i := 0; i < 5; i++ {
		logger.Log(&AuditEvent{
			Type:      AuditEventKeyRead,
			KeyID:     "key-1",
			Operation: "read",
			Success:   true,
		})
	}

	// Verify integrity
	valid, errors := logger.VerifyIntegrity()
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestAuditLogger_GenerateReport(t *testing.T) {
	config := DefaultAuditLogConfig()
	config.LogFile = ""
	logger, _ := NewAuditLogger(config)
	defer logger.Close()

	// Log various events
	logger.LogKeyOperation(AuditEventKeyCreated, "s1", "u1", "User 1", "k1", "create", true, "", nil)
	logger.LogKeyOperation(AuditEventKeySigned, "s1", "u1", "User 1", "k1", "sign", true, "", nil)
	logger.LogKeyOperation(AuditEventKeySigned, "s2", "u2", "User 2", "k2", "sign", false, "key locked", nil)

	since := time.Now().Add(-1 * time.Hour)
	report := logger.GenerateAuditReport(since)

	assert.Equal(t, 3, report.TotalEvents)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 1, report.FailureCount)
	assert.Equal(t, 2, report.ByType[string(AuditEventKeySigned)])
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestKeyManagementIntegration(t *testing.T) {
	// Create all components
	config := DefaultKeyManagerConfig()
	config.StorageType = KeyStorageTypeMemory
	km, _ := NewKeyManager(config)
	km.Unlock("")

	lm := NewKeyLifecycleManager(km)
	detector := NewCompromiseDetector(nil, km)
	ac := NewAccessController(nil)
	auditConfig := DefaultAuditLogConfig()
	auditConfig.LogFile = ""
	logger, _ := NewAuditLogger(auditConfig)
	defer logger.Close()

	// Create admin user
	ac.CreatePrincipal(&Principal{
		ID:    "admin",
		Type:  "user",
		Name:  "Admin",
		Roles: []string{"admin"},
	})

	session, _ := ac.CreateSession("admin", "127.0.0.1", "test")

	// Generate a key
	key, err := km.GenerateKey("provider-1")
	require.NoError(t, err)

	// Register with lifecycle manager
	_, err = lm.RegisterKey(key.KeyID, key.Algorithm, key.PublicKey, "default")
	require.NoError(t, err)

	// Activate the key
	err = lm.ActivateKey(key.KeyID, "admin")
	require.NoError(t, err)

	// Log the operation
	err = logger.LogKeyOperation(
		AuditEventKeyCreated,
		session.ID,
		"admin",
		"Admin",
		key.KeyID,
		"generate_and_activate",
		true,
		"",
		nil,
	)
	require.NoError(t, err)

	// Sign a message and record usage
	message := []byte("transaction data")
	sig, err := km.Sign(message)
	require.NoError(t, err)
	require.NotNil(t, sig)

	// Record the key usage for compromise detection
	indicators := detector.RecordKeyUsage(key.KeyID, "127.0.0.1", time.Now())
	assert.Empty(t, indicators)

	// Verify the lifecycle state
	record, _ := lm.GetRecord(key.KeyID)
	assert.Equal(t, KeyStateActive, record.CurrentState)

	// Verify audit log
	events := logger.GetEventsByKey(key.KeyID, time.Now().Add(-1*time.Hour))
	assert.Len(t, events, 1)
}
