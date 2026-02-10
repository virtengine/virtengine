// Package keystorage provides secure key storage backends for the signer service.
package keystorage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// FileStorage implements KeyStorage using encrypted files.
type FileStorage struct {
	mu            sync.RWMutex
	config        FileStorageConfig
	encryptionKey []byte
	closed        bool
}

// encryptedKeyFile represents the structure of an encrypted key file.
type encryptedKeyFile struct {
	Version    string            `json:"version"`
	KeyID      string            `json:"key_id"`
	SignerID   string            `json:"signer_id"`
	Metadata   StoredKeyMetadata `json:"metadata"`
	PublicKey  string            `json:"public_key"` // base64
	Ciphertext string            `json:"ciphertext"` // base64 encrypted private key
	Nonce      string            `json:"nonce"`      // base64 AES-GCM nonce
	CreatedAt  time.Time         `json:"created_at"`
}

const (
	fileStorageVersion = "1.0.0"
	keyFileExtension   = ".key.json"
)

// NewFileStorage creates a new file-based key storage.
func NewFileStorage(config FileStorageConfig) (*FileStorage, error) {
	if config.Directory == "" {
		return nil, ErrInvalidConfig.Wrap("directory is required")
	}

	// Validate directory path to prevent traversal
	if strings.Contains(config.Directory, "\x00") {
		return nil, ErrInvalidConfig.Wrap("directory path contains null byte")
	}

	// Clean and resolve to absolute path
	absDir, err := filepath.Abs(filepath.Clean(config.Directory))
	if err != nil {
		return nil, ErrInvalidConfig.Wrapf("invalid directory path: %v", err)
	}
	config.Directory = absDir

	if config.EncryptionKey == "" {
		return nil, ErrInvalidConfig.Wrap("encryption_key is required")
	}

	// Decode encryption key
	encKey, err := base64.StdEncoding.DecodeString(config.EncryptionKey)
	if err != nil {
		return nil, ErrInvalidConfig.Wrap("invalid encryption_key: must be base64 encoded")
	}

	// Validate key length (must be 16, 24, or 32 bytes for AES)
	switch len(encKey) {
	case 16, 24, 32:
		// Valid
	default:
		return nil, ErrInvalidConfig.Wrap("encryption_key must be 16, 24, or 32 bytes")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Directory, 0700); err != nil {
		return nil, ErrStorageError.Wrapf("failed to create directory: %v", err)
	}

	if config.FilePermissions == 0 {
		config.FilePermissions = 0600
	}

	return &FileStorage{
		config:        config,
		encryptionKey: encKey,
	}, nil
}

// StoreKey stores a key pair in an encrypted file.
func (f *FileStorage) StoreKey(ctx context.Context, keyInfo *veidtypes.SignerKeyInfo, privateKey []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	filename := f.keyFilePath(keyInfo.KeyID)

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return ErrKeyExists.Wrapf("key ID: %s", keyInfo.KeyID)
	}

	// Encrypt the private key
	ciphertext, nonce, err := f.encrypt(privateKey)
	if err != nil {
		return ErrEncryptionError.Wrapf("failed to encrypt private key: %v", err)
	}

	// Build file content
	keyFile := encryptedKeyFile{
		Version:  fileStorageVersion,
		KeyID:    keyInfo.KeyID,
		SignerID: keyInfo.SignerID,
		Metadata: StoredKeyMetadata{
			KeyID:       keyInfo.KeyID,
			SignerID:    keyInfo.SignerID,
			Fingerprint: keyInfo.Fingerprint,
			Algorithm:   keyInfo.Algorithm,
			State:       keyInfo.State,
			CreatedAt:   keyInfo.CreatedAt,
			StoredAt:    time.Now(),
			ExpiresAt:   keyInfo.ExpiresAt,
			Version:     fileStorageVersion,
		},
		PublicKey:  base64.StdEncoding.EncodeToString(keyInfo.PublicKey),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		CreatedAt:  time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return ErrStorageError.Wrapf("failed to marshal key file: %v", err)
	}

	// Write to file
	// #nosec G304 -- filename is constructed from validated keyID via keyFilePath()
	if err := os.WriteFile(filename, data, os.FileMode(f.config.FilePermissions)); err != nil {
		return ErrStorageError.Wrapf("failed to write key file: %v", err)
	}

	return nil
}

// GetKeyInfo retrieves key metadata.
func (f *FileStorage) GetKeyInfo(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	keyFile, err := f.readKeyFile(keyID)
	if err != nil {
		return nil, err
	}

	publicKey, err := base64.StdEncoding.DecodeString(keyFile.PublicKey)
	if err != nil {
		return nil, ErrStorageError.Wrap("failed to decode public key")
	}

	keyInfo := &veidtypes.SignerKeyInfo{
		KeyID:       keyFile.KeyID,
		Fingerprint: keyFile.Metadata.Fingerprint,
		PublicKey:   publicKey,
		Algorithm:   keyFile.Metadata.Algorithm,
		State:       keyFile.Metadata.State,
		SignerID:    keyFile.SignerID,
		CreatedAt:   keyFile.Metadata.CreatedAt,
		ExpiresAt:   keyFile.Metadata.ExpiresAt,
	}

	return keyInfo, nil
}

// GetPrivateKey retrieves the private key.
func (f *FileStorage) GetPrivateKey(ctx context.Context, keyID string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	keyFile, err := f.readKeyFile(keyID)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(keyFile.Ciphertext)
	if err != nil {
		return nil, ErrStorageError.Wrap("failed to decode ciphertext")
	}

	nonce, err := base64.StdEncoding.DecodeString(keyFile.Nonce)
	if err != nil {
		return nil, ErrStorageError.Wrap("failed to decode nonce")
	}

	privateKey, err := f.decrypt(ciphertext, nonce)
	if err != nil {
		return nil, ErrDecryptionError.Wrapf("failed to decrypt private key: %v", err)
	}

	return privateKey, nil
}

// ListKeys returns all keys for a signer.
func (f *FileStorage) ListKeys(ctx context.Context, signerID string) ([]*veidtypes.SignerKeyInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, ErrStorageError.Wrap("storage is closed")
	}

	pattern := filepath.Join(f.config.Directory, "*"+keyFileExtension)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, ErrStorageError.Wrapf("failed to list key files: %v", err)
	}

	result := make([]*veidtypes.SignerKeyInfo, 0, len(files))
	for _, file := range files {
		// #nosec G304 -- files are from trusted storage directory (filepath.Glob on config.Directory)
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var keyFile encryptedKeyFile
		if err := json.Unmarshal(data, &keyFile); err != nil {
			continue
		}

		if keyFile.SignerID != signerID {
			continue
		}

		publicKey, err := base64.StdEncoding.DecodeString(keyFile.PublicKey)
		if err != nil {
			continue
		}

		keyInfo := &veidtypes.SignerKeyInfo{
			KeyID:       keyFile.KeyID,
			Fingerprint: keyFile.Metadata.Fingerprint,
			PublicKey:   publicKey,
			Algorithm:   keyFile.Metadata.Algorithm,
			State:       keyFile.Metadata.State,
			SignerID:    keyFile.SignerID,
			CreatedAt:   keyFile.Metadata.CreatedAt,
			ExpiresAt:   keyFile.Metadata.ExpiresAt,
		}

		result = append(result, keyInfo)
	}

	return result, nil
}

// UpdateKeyState updates the state of a key.
func (f *FileStorage) UpdateKeyState(ctx context.Context, keyID string, state veidtypes.SignerKeyState) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	filename := f.keyFilePath(keyID)

	// #nosec G304 -- filename is constructed from validated keyID via keyFilePath()
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrKeyNotFound.Wrapf("key ID: %s", keyID)
		}
		return ErrStorageError.Wrapf("failed to read key file: %v", err)
	}

	var keyFile encryptedKeyFile
	if err := json.Unmarshal(data, &keyFile); err != nil {
		return ErrStorageError.Wrapf("failed to parse key file: %v", err)
	}

	keyFile.Metadata.State = state

	updatedData, err := json.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return ErrStorageError.Wrapf("failed to marshal key file: %v", err)
	}

	if err := os.WriteFile(filename, updatedData, os.FileMode(f.config.FilePermissions)); err != nil {
		return ErrStorageError.Wrapf("failed to write key file: %v", err)
	}

	return nil
}

// DeleteKey deletes a key.
func (f *FileStorage) DeleteKey(ctx context.Context, keyID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	filename := f.keyFilePath(keyID)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	if err := os.Remove(filename); err != nil {
		return ErrStorageError.Wrapf("failed to delete key file: %v", err)
	}

	return nil
}

// HealthCheck verifies the storage is accessible.
func (f *FileStorage) HealthCheck(ctx context.Context) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return ErrStorageError.Wrap("storage is closed")
	}

	// Check directory is accessible
	info, err := os.Stat(f.config.Directory)
	if err != nil {
		return ErrStorageError.Wrapf("directory not accessible: %v", err)
	}

	if !info.IsDir() {
		return ErrStorageError.Wrap("path is not a directory")
	}

	return nil
}

// Close closes the storage.
func (f *FileStorage) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clear encryption key
	for i := range f.encryptionKey {
		f.encryptionKey[i] = 0
	}

	f.closed = true
	return nil
}

// Helper methods

func (f *FileStorage) keyFilePath(keyID string) string {
	// Sanitize key ID for filename - prevent path traversal
	safeKeyID := filepath.Base(keyID)
	// Double-check for any remaining traversal sequences
	if strings.Contains(safeKeyID, "..") || safeKeyID == "." || safeKeyID == "" {
		safeKeyID = "invalid_key"
	}
	return filepath.Join(f.config.Directory, safeKeyID+keyFileExtension)
}

// validateKeyPath ensures a path is within the storage directory
func (f *FileStorage) validateKeyPath(path string) error {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	absDir, err := filepath.Abs(f.config.Directory)
	if err != nil {
		return fmt.Errorf("cannot resolve directory: %w", err)
	}

	// Ensure path is within storage directory
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) && absPath != absDir {
		return fmt.Errorf("path %s is outside storage directory", path)
	}

	return nil
}

func (f *FileStorage) readKeyFile(keyID string) (*encryptedKeyFile, error) {
	filename := f.keyFilePath(keyID)

	// Validate path before reading
	if err := f.validateKeyPath(filename); err != nil {
		return nil, ErrStorageError.Wrapf("invalid key path: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(filename)) // #nosec G304 -- path validated above
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrKeyNotFound.Wrapf("key ID: %s", keyID)
		}
		return nil, ErrStorageError.Wrapf("failed to read key file: %v", err)
	}

	var keyFile encryptedKeyFile
	if err := json.Unmarshal(data, &keyFile); err != nil {
		return nil, ErrStorageError.Wrapf("failed to parse key file: %v", err)
	}

	return &keyFile, nil
}

func (f *FileStorage) encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(f.encryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (f *FileStorage) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// Ensure FileStorage implements KeyStorage
var _ KeyStorage = (*FileStorage)(nil)
