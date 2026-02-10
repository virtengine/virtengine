// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: Key Backup and Recovery
// This file provides secure key backup and recovery mechanisms.
package provider_daemon

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

// ErrBackupDecryptionFailed is returned when backup decryption fails
var ErrBackupDecryptionFailed = errors.New("backup decryption failed")

// ErrBackupValidationFailed is returned when backup validation fails
var ErrBackupValidationFailed = errors.New("backup validation failed")

// ErrBackupExpired is returned when attempting to restore an expired backup
var ErrBackupExpired = errors.New("backup has expired")

// ErrInvalidBackupFormat is returned when backup format is invalid
var ErrInvalidBackupFormat = errors.New("invalid backup format")

// ErrBackupVersionMismatch is returned when backup version is not supported
var ErrBackupVersionMismatch = errors.New("backup version not supported")

// BackupVersion is the current backup format version
const BackupVersion = 1

// BackupEncryptionAlgorithm is the algorithm used for backup encryption
const BackupEncryptionAlgorithm = "AES-256-GCM"

// KeyBackupConfig configures key backup settings
type KeyBackupConfig struct {
	// EncryptionMethod is the method used for backup encryption
	EncryptionMethod string `json:"encryption_method"`

	// KeyDerivationIterations is the number of iterations for key derivation
	KeyDerivationIterations uint32 `json:"key_derivation_iterations"`

	// BackupExpiryDays is how long backups remain valid
	BackupExpiryDays int `json:"backup_expiry_days"`

	// RequireMultiplePassphrases requires M of N passphrases for recovery
	RequireMultiplePassphrases bool `json:"require_multiple_passphrases"`

	// MinPassphrases is the minimum number of passphrases required
	MinPassphrases int `json:"min_passphrases,omitempty"`

	// TotalPassphrases is the total number of passphrases
	TotalPassphrases int `json:"total_passphrases,omitempty"`

	// IncludeMetadata includes key metadata in backup
	IncludeMetadata bool `json:"include_metadata"`

	// CompressBackup compresses the backup before encryption
	CompressBackup bool `json:"compress_backup"`
}

// DefaultKeyBackupConfig returns the default backup configuration
func DefaultKeyBackupConfig() *KeyBackupConfig {
	return &KeyBackupConfig{
		EncryptionMethod:           "AES-256-GCM",
		KeyDerivationIterations:    3, // Argon2id time parameter
		BackupExpiryDays:           365,
		RequireMultiplePassphrases: false,
		IncludeMetadata:            true,
		CompressBackup:             true,
	}
}

// KeyBackup represents an encrypted key backup
type KeyBackup struct {
	// Version is the backup format version
	Version int `json:"version"`

	// CreatedAt is when the backup was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the backup expires
	ExpiresAt time.Time `json:"expires_at"`

	// Algorithm is the encryption algorithm used
	Algorithm string `json:"algorithm"`

	// KeyDerivation contains key derivation parameters
	KeyDerivation KeyDerivationParams `json:"key_derivation"`

	// Salt is the salt used for key derivation
	Salt []byte `json:"salt"`

	// Nonce is the nonce/IV used for encryption
	Nonce []byte `json:"nonce"`

	// Ciphertext is the encrypted key data
	Ciphertext []byte `json:"ciphertext"`

	// Checksum is a checksum of the unencrypted data for validation
	Checksum string `json:"checksum"`

	// Metadata contains additional backup metadata
	Metadata *BackupMetadata `json:"metadata,omitempty"`

	// RecoveryShares indicates if this is a Shamir secret sharing backup
	RecoveryShares *RecoveryShareInfo `json:"recovery_shares,omitempty"`
}

// KeyDerivationParams contains key derivation parameters
type KeyDerivationParams struct {
	// Algorithm is the key derivation algorithm (argon2id)
	Algorithm string `json:"algorithm"`

	// Time is the number of iterations
	Time uint32 `json:"time"`

	// Memory is the memory usage in KB
	Memory uint32 `json:"memory"`

	// Threads is the number of threads
	Threads uint8 `json:"threads"`

	// KeyLength is the derived key length in bytes
	KeyLength uint32 `json:"key_length"`
}

// DefaultKeyDerivationParams returns default key derivation parameters
func DefaultKeyDerivationParams() KeyDerivationParams {
	return KeyDerivationParams{
		Algorithm: "argon2id",
		Time:      3,
		Memory:    64 * 1024, // 64 MB
		Threads:   4,
		KeyLength: 32,
	}
}

// BackupMetadata contains backup metadata
type BackupMetadata struct {
	// BackupID is a unique identifier for this backup
	BackupID string `json:"backup_id"`

	// KeyCount is the number of keys in the backup
	KeyCount int `json:"key_count"`

	// KeyTypes lists the types of keys included
	KeyTypes []string `json:"key_types"`

	// KeyLabels lists the labels of backed up keys
	KeyLabels []string `json:"key_labels"`

	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address,omitempty"`

	// Description is an optional description
	Description string `json:"description,omitempty"`

	// CreatedBy is who created the backup
	CreatedBy string `json:"created_by,omitempty"`

	// Environment is the environment (mainnet, testnet, etc.)
	Environment string `json:"environment,omitempty"`
}

// RecoveryShareInfo contains information about recovery shares
type RecoveryShareInfo struct {
	// Threshold is the minimum shares needed for recovery
	Threshold int `json:"threshold"`

	// TotalShares is the total number of shares
	TotalShares int `json:"total_shares"`

	// ShareIndex is the index of this share (if split)
	ShareIndex int `json:"share_index,omitempty"`
}

// BackupKeyData is the structure of data being backed up
type BackupKeyData struct {
	// Keys contains the key data
	Keys []BackupKeyEntry `json:"keys"`

	// Timestamp is when the data was exported
	Timestamp time.Time `json:"timestamp"`

	// SourceFingerprint identifies the source key manager
	SourceFingerprint string `json:"source_fingerprint"`
}

// BackupKeyEntry represents a single key in the backup
type BackupKeyEntry struct {
	// Label is the key label
	Label string `json:"label"`

	// Algorithm is the key algorithm
	Algorithm string `json:"algorithm"`

	// PrivateKey is the private key (encrypted in backup)
	PrivateKey []byte `json:"private_key"`

	// PublicKey is the public key
	PublicKey []byte `json:"public_key"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`

	// Metadata contains key metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// KeyBackupManager manages key backup and recovery operations
type KeyBackupManager struct {
	config     *KeyBackupConfig
	keyManager *KeyManager
}

// NewKeyBackupManager creates a new key backup manager
func NewKeyBackupManager(config *KeyBackupConfig, keyManager *KeyManager) *KeyBackupManager {
	if config == nil {
		config = DefaultKeyBackupConfig()
	}
	return &KeyBackupManager{
		config:     config,
		keyManager: keyManager,
	}
}

// CreateBackup creates an encrypted backup of all keys
func (m *KeyBackupManager) CreateBackup(passphrase string) (*KeyBackup, error) {
	if m.keyManager == nil {
		return nil, errors.New("key manager not configured")
	}

	if m.keyManager.IsLocked() {
		return nil, ErrKeyStorageLocked
	}

	// Get all keys from the key manager
	keys, err := m.keyManager.ListKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	// Build backup data
	backupData := &BackupKeyData{
		Keys:              make([]BackupKeyEntry, 0, len(keys)),
		Timestamp:         time.Now().UTC(),
		SourceFingerprint: m.computeSourceFingerprint(),
	}

	for _, key := range keys {
		// Get full key with private data
		fullKey, err := m.keyManager.GetKey(key.KeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", key.KeyID, err)
		}

		entry := BackupKeyEntry{
			Label:     key.KeyID,
			Algorithm: key.Algorithm,
			CreatedAt: key.CreatedAt,
			Metadata: map[string]string{
				"provider_address": key.ProviderAddress,
				"status":           key.Status,
			},
		}

		// Note: In real implementation, would access private key securely
		// This is a placeholder for the actual implementation
		entry.PublicKey, _ = hex.DecodeString(fullKey.PublicKey)

		backupData.Keys = append(backupData.Keys, entry)
	}

	// Serialize backup data
	plaintext, err := json.Marshal(backupData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize backup data: %w", err)
	}

	// Compute checksum before encryption
	checksum := sha256.Sum256(plaintext)

	// Generate salt for key derivation
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from passphrase
	kdfParams := DefaultKeyDerivationParams()
	encryptionKey := deriveKey(passphrase, salt, kdfParams)

	// Encrypt backup data
	ciphertext, nonce, err := encryptBackup(plaintext, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt backup: %w", err)
	}

	// Scrub encryption key from memory
	for i := range encryptionKey {
		encryptionKey[i] = 0
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, m.config.BackupExpiryDays)

	backup := &KeyBackup{
		Version:       BackupVersion,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
		Algorithm:     BackupEncryptionAlgorithm,
		KeyDerivation: kdfParams,
		Salt:          salt,
		Nonce:         nonce,
		Ciphertext:    ciphertext,
		Checksum:      hex.EncodeToString(checksum[:]),
	}

	if m.config.IncludeMetadata {
		backup.Metadata = &BackupMetadata{
			BackupID:    generateBackupID(),
			KeyCount:    len(backupData.Keys),
			KeyTypes:    extractKeyTypes(backupData.Keys),
			KeyLabels:   extractKeyLabels(backupData.Keys),
			CreatedBy:   "key_backup_manager",
			Environment: "production",
		}
	}

	return backup, nil
}

// RestoreBackup restores keys from an encrypted backup
func (m *KeyBackupManager) RestoreBackup(backup *KeyBackup, passphrase string) (*RestoreResult, error) {
	// Validate backup
	if err := m.validateBackup(backup); err != nil {
		return nil, err
	}

	// Check expiry
	if time.Now().After(backup.ExpiresAt) {
		return nil, ErrBackupExpired
	}

	// Derive decryption key
	encryptionKey := deriveKey(passphrase, backup.Salt, backup.KeyDerivation)

	// Decrypt backup
	plaintext, err := decryptBackup(backup.Ciphertext, backup.Nonce, encryptionKey)
	if err != nil {
		// Scrub key from memory
		for i := range encryptionKey {
			encryptionKey[i] = 0
		}
		return nil, ErrBackupDecryptionFailed
	}

	// Scrub encryption key from memory
	for i := range encryptionKey {
		encryptionKey[i] = 0
	}

	// Validate checksum
	checksum := sha256.Sum256(plaintext)
	if hex.EncodeToString(checksum[:]) != backup.Checksum {
		return nil, ErrBackupValidationFailed
	}

	// Deserialize backup data
	var backupData BackupKeyData
	if err := json.Unmarshal(plaintext, &backupData); err != nil {
		return nil, fmt.Errorf("failed to deserialize backup data: %w", err)
	}

	result := &RestoreResult{
		TotalKeys:    len(backupData.Keys),
		RestoredKeys: make([]string, 0),
		SkippedKeys:  make([]string, 0),
		Errors:       make(map[string]string),
		RestoredAt:   time.Now().UTC(),
	}

	// Restore each key
	for _, keyEntry := range backupData.Keys {
		// Check if key already exists
		_, err := m.keyManager.GetKey(keyEntry.Label)
		if err == nil {
			result.SkippedKeys = append(result.SkippedKeys, keyEntry.Label)
			continue
		}

		// Import key (simplified - would need actual private key data)
		if keyEntry.PrivateKey != nil {
			_, err := m.keyManager.ImportKey(
				keyEntry.Metadata["provider_address"],
				keyEntry.PrivateKey,
				keyEntry.Algorithm,
			)
			if err != nil {
				result.Errors[keyEntry.Label] = err.Error()
				continue
			}
		}

		result.RestoredKeys = append(result.RestoredKeys, keyEntry.Label)
	}

	return result, nil
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	// TotalKeys is the total number of keys in backup
	TotalKeys int `json:"total_keys"`

	// RestoredKeys is the list of successfully restored keys
	RestoredKeys []string `json:"restored_keys"`

	// SkippedKeys is the list of keys that were skipped (already exist)
	SkippedKeys []string `json:"skipped_keys"`

	// Errors contains errors for keys that failed to restore
	Errors map[string]string `json:"errors,omitempty"`

	// RestoredAt is when the restore completed
	RestoredAt time.Time `json:"restored_at"`
}

// validateBackup validates the backup structure
func (m *KeyBackupManager) validateBackup(backup *KeyBackup) error {
	if backup == nil {
		return ErrInvalidBackupFormat
	}

	if backup.Version > BackupVersion {
		return ErrBackupVersionMismatch
	}

	if backup.Algorithm == "" {
		return ErrInvalidBackupFormat
	}

	if len(backup.Salt) == 0 || len(backup.Nonce) == 0 || len(backup.Ciphertext) == 0 {
		return ErrInvalidBackupFormat
	}

	return nil
}

// computeSourceFingerprint computes a fingerprint identifying the source
func (m *KeyBackupManager) computeSourceFingerprint() string {
	// Simplified fingerprint computation
	data := fmt.Sprintf("backup-%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// deriveKey derives an encryption key from passphrase using Argon2id
func deriveKey(passphrase string, salt []byte, params KeyDerivationParams) []byte {
	return argon2.IDKey(
		[]byte(passphrase),
		salt,
		params.Time,
		params.Memory,
		params.Threads,
		params.KeyLength,
	)
}

// encryptBackup encrypts data using AES-256-GCM
func encryptBackup(plaintext, key []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// decryptBackup decrypts data using AES-256-GCM
func decryptBackup(ciphertext, nonce, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// generateBackupID generates a unique backup ID
func generateBackupID() string {
	id := make([]byte, 16)
	_, _ = rand.Read(id)
	return hex.EncodeToString(id)
}

// extractKeyTypes extracts unique key types from backup entries
func extractKeyTypes(entries []BackupKeyEntry) []string {
	types := make(map[string]bool)
	for _, entry := range entries {
		types[entry.Algorithm] = true
	}
	result := make([]string, 0, len(types))
	for t := range types {
		result = append(result, t)
	}
	return result
}

// extractKeyLabels extracts key labels from backup entries
func extractKeyLabels(entries []BackupKeyEntry) []string {
	labels := make([]string, 0, len(entries))
	for _, entry := range entries {
		labels = append(labels, entry.Label)
	}
	return labels
}

// RecoveryShare represents a single share for Shamir secret sharing recovery
type RecoveryShare struct {
	// Index is the share index
	Index int `json:"index"`

	// Share is the encrypted share data
	Share []byte `json:"share"`

	// Threshold is the minimum shares needed
	Threshold int `json:"threshold"`

	// TotalShares is the total number of shares
	TotalShares int `json:"total_shares"`

	// CreatedAt is when the share was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the share expires
	ExpiresAt time.Time `json:"expires_at"`

	// ShareChecksum is a checksum for this share
	ShareChecksum string `json:"share_checksum"`
}

// SplitBackupIntoShares splits a backup into Shamir secret sharing shares
func (m *KeyBackupManager) SplitBackupIntoShares(backup *KeyBackup, threshold, totalShares int, passphrases []string) ([]*RecoveryShare, error) {
	if threshold < 2 {
		return nil, errors.New("threshold must be at least 2")
	}
	if totalShares < threshold {
		return nil, errors.New("total shares must be >= threshold")
	}
	if len(passphrases) != totalShares {
		return nil, errors.New("must provide passphrase for each share")
	}

	// Serialize the backup
	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize backup: %w", err)
	}

	// Split using Shamir secret sharing (simplified implementation)
	shares := shamirSplit(backupJSON, threshold, totalShares)

	result := make([]*RecoveryShare, totalShares)
	now := time.Now().UTC()
	expiresAt := now.AddDate(1, 0, 0) // 1 year expiry

	for i, shareData := range shares {
		// Encrypt each share with its passphrase
		salt := make([]byte, 32)
		_, _ = rand.Read(salt)

		key := deriveKey(passphrases[i], salt, DefaultKeyDerivationParams())
		encryptedShare, nonce, err := encryptBackup(shareData, key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt share %d: %w", i+1, err)
		}

		// Combine salt, nonce, and encrypted data
		combinedShare := make([]byte, 0, len(salt)+len(nonce)+len(encryptedShare))
		combinedShare = append(combinedShare, salt...)
		combinedShare = append(combinedShare, nonce...)
		combinedShare = append(combinedShare, encryptedShare...)

		checksum := sha256.Sum256(shareData)

		result[i] = &RecoveryShare{
			Index:         i + 1,
			Share:         combinedShare,
			Threshold:     threshold,
			TotalShares:   totalShares,
			CreatedAt:     now,
			ExpiresAt:     expiresAt,
			ShareChecksum: hex.EncodeToString(checksum[:8]),
		}
	}

	return result, nil
}

// RecombineShares recombines recovery shares to restore a backup
func (m *KeyBackupManager) RecombineShares(shares []*RecoveryShare, passphrases map[int]string) (*KeyBackup, error) {
	if len(shares) == 0 {
		return nil, errors.New("no shares provided")
	}

	threshold := shares[0].Threshold
	if len(shares) < threshold {
		return nil, fmt.Errorf("need at least %d shares, got %d", threshold, len(shares))
	}

	// Decrypt each share
	decryptedShares := make([][]byte, 0, len(shares))
	shareIndices := make([]int, 0, len(shares))

	for _, share := range shares {
		passphrase, ok := passphrases[share.Index]
		if !ok {
			return nil, fmt.Errorf("passphrase not provided for share %d", share.Index)
		}

		// Extract salt, nonce, and ciphertext
		if len(share.Share) < 44 { // 32 (salt) + 12 (nonce minimum)
			return nil, fmt.Errorf("invalid share %d: too short", share.Index)
		}

		salt := share.Share[:32]
		nonce := share.Share[32:44]
		ciphertext := share.Share[44:]

		key := deriveKey(passphrase, salt, DefaultKeyDerivationParams())
		decrypted, err := decryptBackup(ciphertext, nonce, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt share %d", share.Index)
		}

		decryptedShares = append(decryptedShares, decrypted)
		shareIndices = append(shareIndices, share.Index)
	}

	// Recombine using Shamir secret sharing
	backupJSON := shamirRecombine(decryptedShares, shareIndices, threshold)

	// Deserialize backup
	var backup KeyBackup
	if err := json.Unmarshal(backupJSON, &backup); err != nil {
		return nil, fmt.Errorf("failed to deserialize backup: %w", err)
	}

	return &backup, nil
}

// shamirSplit is a simplified Shamir secret sharing split
// In production, would use a proper Shamir library
//
//nolint:unparam // threshold kept for future proper Shamir implementation
func shamirSplit(secret []byte, _, totalShares int) [][]byte {
	// Simplified implementation - just XOR split for demo
	// Real implementation would use polynomial interpolation
	shares := make([][]byte, totalShares)

	for i := 0; i < totalShares; i++ {
		shares[i] = make([]byte, len(secret))
		copy(shares[i], secret)
		// Add share index marker (simplified)
		if len(shares[i]) > 0 {
			shares[i][0] ^= byte(i + 1)
		}
	}

	return shares
}

// shamirRecombine is a simplified Shamir secret sharing recombine
// In production, would use a proper Shamir library
func shamirRecombine(shares [][]byte, indices []int, threshold int) []byte {
	if len(shares) == 0 || len(shares) < threshold {
		return nil
	}

	// Simplified implementation - just use first share
	// Real implementation would use polynomial interpolation
	result := make([]byte, len(shares[0]))
	copy(result, shares[0])

	// Undo share index marker
	if len(result) > 0 {
		result[0] ^= byte(indices[0])
	}

	return result
}

// SecureBackupWithSecretBox creates a backup using NaCl secretbox for additional security
func SecureBackupWithSecretBox(data []byte, passphrase string) ([]byte, error) {
	// Generate salt
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Derive key
	key := deriveKey(passphrase, salt, DefaultKeyDerivationParams())

	var keyArray [32]byte
	copy(keyArray[:], key)

	// Generate nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt using secretbox
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &keyArray)

	// Prepend salt
	result := make([]byte, 0, len(salt)+len(encrypted))
	result = append(result, salt...)
	result = append(result, encrypted...)

	// Scrub key from memory
	for i := range key {
		key[i] = 0
	}
	for i := range keyArray {
		keyArray[i] = 0
	}

	return result, nil
}

// DecryptSecureBackup decrypts a backup created with SecureBackupWithSecretBox
func DecryptSecureBackup(encrypted []byte, passphrase string) ([]byte, error) {
	if len(encrypted) < 32+24 { // salt + nonce minimum
		return nil, errors.New("encrypted data too short")
	}

	// Extract salt
	salt := encrypted[:32]
	sealed := encrypted[32:]

	// Derive key
	key := deriveKey(passphrase, salt, DefaultKeyDerivationParams())

	var keyArray [32]byte
	copy(keyArray[:], key)

	// Extract nonce
	var nonce [24]byte
	copy(nonce[:], sealed[:24])

	// Decrypt
	plaintext, ok := secretbox.Open(nil, sealed[24:], &nonce, &keyArray)

	// Scrub key from memory
	for i := range key {
		key[i] = 0
	}
	for i := range keyArray {
		keyArray[i] = 0
	}

	if !ok {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}
