// Package keystorage provides secure key storage backends for the signer service.
//
// This package implements multiple key storage backends including in-memory,
// file-based, HashiCorp Vault, and HSM integrations.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package keystorage

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Key Storage Interface
// ============================================================================

// KeyStorage defines the interface for secure key storage backends.
type KeyStorage interface {
	// StoreKey stores a key pair securely.
	StoreKey(ctx context.Context, keyInfo *veidtypes.SignerKeyInfo, privateKey []byte) error

	// GetKeyInfo retrieves key metadata (public info only).
	GetKeyInfo(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error)

	// GetPrivateKey retrieves the private key for signing.
	// The caller is responsible for clearing the returned bytes after use.
	GetPrivateKey(ctx context.Context, keyID string) ([]byte, error)

	// ListKeys returns all keys for a signer.
	ListKeys(ctx context.Context, signerID string) ([]*veidtypes.SignerKeyInfo, error)

	// UpdateKeyState updates the state of a key.
	UpdateKeyState(ctx context.Context, keyID string, state veidtypes.SignerKeyState) error

	// DeleteKey deletes a key (for cleanup of expired keys).
	DeleteKey(ctx context.Context, keyID string) error

	// HealthCheck verifies the storage backend is accessible.
	HealthCheck(ctx context.Context) error

	// Close closes the storage backend.
	Close() error
}

// ============================================================================
// Storage Configuration
// ============================================================================

// StorageConfig contains configuration for key storage backends.
type StorageConfig struct {
	// Type specifies the storage backend type
	Type StorageType `json:"type"`

	// Memory-specific config
	Memory *MemoryStorageConfig `json:"memory,omitempty"`

	// File-specific config
	File *FileStorageConfig `json:"file,omitempty"`

	// Vault-specific config
	Vault *VaultStorageConfig `json:"vault,omitempty"`

	// HSM-specific config
	HSM *HSMStorageConfig `json:"hsm,omitempty"`

	// KMS-specific config
	KMS *KMSStorageConfig `json:"kms,omitempty"`
}

// StorageType identifies the storage backend type.
type StorageType string

const (
	StorageTypeMemory StorageType = "memory"
	StorageTypeFile   StorageType = "file"
	StorageTypeVault  StorageType = "vault"
	StorageTypeHSM    StorageType = "hsm"
	StorageTypeKMS    StorageType = "kms"
)

// MemoryStorageConfig contains config for in-memory storage.
type MemoryStorageConfig struct {
	// MaxKeys is the maximum number of keys to store
	MaxKeys int `json:"max_keys"`
}

// FileStorageConfig contains config for file-based storage.
type FileStorageConfig struct {
	// Directory is the directory to store key files
	Directory string `json:"directory"`

	// EncryptionKey is the key for encrypting stored keys (base64 encoded)
	EncryptionKey string `json:"encryption_key"`

	// FilePermissions is the Unix file permissions for key files
	FilePermissions uint32 `json:"file_permissions"`
}

// VaultStorageConfig contains config for HashiCorp Vault storage.
type VaultStorageConfig struct {
	// Address is the Vault server address
	Address string `json:"address"`

	// Token is the Vault authentication token
	Token string `json:"token"`

	// MountPath is the secrets engine mount path
	MountPath string `json:"mount_path"`

	// KeyPath is the path prefix for keys
	KeyPath string `json:"key_path"`

	// TLSConfig contains TLS configuration
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`

	// Namespace is the Vault namespace (Enterprise only)
	Namespace string `json:"namespace,omitempty"`
}

// HSMStorageConfig contains config for HSM storage.
type HSMStorageConfig struct {
	// Library is the path to the PKCS#11 library
	Library string `json:"library"`

	// SlotID is the HSM slot ID
	SlotID uint `json:"slot_id"`

	// PIN is the HSM PIN
	PIN string `json:"pin"`

	// TokenLabel is the HSM token label
	TokenLabel string `json:"token_label"`
}

// KMSStorageConfig contains config for cloud KMS storage.
type KMSStorageConfig struct {
	// Provider is the KMS provider (aws, gcp, azure)
	Provider string `json:"provider"`

	// KeyRingID is the key ring identifier
	KeyRingID string `json:"key_ring_id"`

	// Region is the cloud region
	Region string `json:"region"`

	// CredentialsPath is the path to credentials file
	CredentialsPath string `json:"credentials_path"`
}

// TLSConfig contains TLS configuration.
type TLSConfig struct {
	// CACert is the path to the CA certificate
	CACert string `json:"ca_cert"`

	// ClientCert is the path to the client certificate
	ClientCert string `json:"client_cert"`

	// ClientKey is the path to the client key
	ClientKey string `json:"client_key"`

	// ServerName is the expected server name
	ServerName string `json:"server_name"`

	// InsecureSkipVerify skips TLS verification (DANGEROUS)
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

// ============================================================================
// Key Metadata
// ============================================================================

// StoredKeyMetadata contains metadata about a stored key.
type StoredKeyMetadata struct {
	// KeyID is the key identifier
	KeyID string `json:"key_id"`

	// SignerID is the signer that owns this key
	SignerID string `json:"signer_id"`

	// Fingerprint is the key fingerprint
	Fingerprint string `json:"fingerprint"`

	// Algorithm is the key algorithm
	Algorithm veidtypes.AttestationProofType `json:"algorithm"`

	// State is the key state
	State veidtypes.SignerKeyState `json:"state"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`

	// StoredAt is when the key was stored
	StoredAt time.Time `json:"stored_at"`

	// ExpiresAt is when the key expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Version is the storage format version
	Version string `json:"version"`
}

// DefaultStorageConfig returns a default storage configuration.
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Type: StorageTypeMemory,
		Memory: &MemoryStorageConfig{
			MaxKeys: 100,
		},
	}
}

// DefaultFileStorageConfig returns a default file storage configuration.
func DefaultFileStorageConfig() FileStorageConfig {
	return FileStorageConfig{
		Directory:       "/var/lib/virtengine/keys",
		FilePermissions: 0600,
	}
}

// DefaultVaultStorageConfig returns a default Vault storage configuration.
func DefaultVaultStorageConfig() VaultStorageConfig {
	return VaultStorageConfig{
		Address:   "https://vault.virtengine.local:8200",
		MountPath: "secret",
		KeyPath:   "veid/signer/keys",
	}
}

