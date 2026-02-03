// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-14B: HPC Credential Manager - secure credential management for HPC backends
package provider_daemon

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CredentialType represents the type of HPC credential
type CredentialType string

const (
	// CredentialTypeSLURM for SLURM SSH credentials
	CredentialTypeSLURM CredentialType = "slurm"

	// CredentialTypeMOAB for MOAB credentials
	CredentialTypeMOAB CredentialType = "moab"

	// CredentialTypeOOD for Open OnDemand credentials
	CredentialTypeOOD CredentialType = "ood"

	// CredentialTypeKerberos for Kerberos tickets
	CredentialTypeKerberos CredentialType = "kerberos"

	// CredentialTypeSigning for provider signing keys
	CredentialTypeSigning CredentialType = "signing"
)

// HPCCredentials contains credentials for HPC backend access
type HPCCredentials struct {
	// Type is the credential type
	Type CredentialType `json:"type"`

	// ClusterID is the cluster these credentials are for
	ClusterID string `json:"cluster_id"`

	// Username is the HPC username
	Username string `json:"username,omitempty"`

	// Password is the password (encrypted at rest)
	Password string `json:"-"` // Never serialize

	// SSHPrivateKey is the SSH private key content (encrypted at rest)
	SSHPrivateKey string `json:"-"` // Never serialize

	// SSHPrivateKeyPath is the path to SSH private key file
	SSHPrivateKeyPath string `json:"ssh_private_key_path,omitempty"`

	// SSHPassphrase is the passphrase for the SSH key
	SSHPassphrase string `json:"-"` // Never serialize

	// KerberosKeytab is the path to Kerberos keytab
	KerberosKeytab string `json:"kerberos_keytab,omitempty"`

	// KerberosPrincipal is the Kerberos principal
	KerberosPrincipal string `json:"kerberos_principal,omitempty"`

	// SigningKey is the ed25519 signing key (encrypted at rest)
	SigningKey []byte `json:"-"` // Never serialize

	// CreatedAt is when the credentials were created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the credentials expire (zero for no expiry)
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Metadata contains additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsExpired returns true if the credentials have expired
func (c *HPCCredentials) IsExpired() bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// HPCCredentialManagerConfig configures the credential manager
type HPCCredentialManagerConfig struct {
	// StorageDir is the directory for encrypted credential storage
	StorageDir string `json:"storage_dir" yaml:"storage_dir"`

	// EncryptionKeyPath is the path to the master encryption key
	EncryptionKeyPath string `json:"encryption_key_path" yaml:"encryption_key_path"`

	// AllowUnencrypted allows unencrypted storage (for testing only)
	AllowUnencrypted bool `json:"allow_unencrypted" yaml:"allow_unencrypted"`

	// RotationCheckInterval is how often to check for credential rotation
	RotationCheckInterval time.Duration `json:"rotation_check_interval" yaml:"rotation_check_interval"`

	// RotationWarningDays is days before expiry to warn about rotation
	RotationWarningDays int `json:"rotation_warning_days" yaml:"rotation_warning_days"`
}

// DefaultHPCCredentialManagerConfig returns the default configuration
func DefaultHPCCredentialManagerConfig() HPCCredentialManagerConfig {
	return HPCCredentialManagerConfig{
		StorageDir:            "/var/lib/virtengine/hpc-credentials",
		RotationCheckInterval: 24 * time.Hour,
		RotationWarningDays:   14,
	}
}

// HPCCredentialManager manages HPC credentials securely
type HPCCredentialManager struct {
	config        HPCCredentialManagerConfig
	encryptionKey []byte

	mu          sync.RWMutex
	credentials map[string]map[CredentialType]*HPCCredentials // clusterID -> type -> creds
	signingKey  ed25519.PrivateKey
	publicKey   ed25519.PublicKey
	locked      bool
}

// NewHPCCredentialManager creates a new credential manager
func NewHPCCredentialManager(config HPCCredentialManagerConfig) (*HPCCredentialManager, error) {
	cm := &HPCCredentialManager{
		config:      config,
		credentials: make(map[string]map[CredentialType]*HPCCredentials),
		locked:      true,
	}

	// Ensure storage directory exists
	if config.StorageDir != "" {
		if err := os.MkdirAll(config.StorageDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	return cm, nil
}

// Unlock unlocks the credential manager with a passphrase
func (cm *HPCCredentialManager) Unlock(passphrase string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if passphrase == "" && !cm.config.AllowUnencrypted {
		return errors.New("passphrase required")
	}

	// Derive encryption key from passphrase
	if passphrase != "" {
		hash := sha256.Sum256([]byte(passphrase))
		cm.encryptionKey = hash[:]
	}

	// Load encryption key from file if specified
	if cm.config.EncryptionKeyPath != "" {
		keyData, err := os.ReadFile(cm.config.EncryptionKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read encryption key: %w", err)
		}
		if len(keyData) < 32 {
			return errors.New("encryption key too short")
		}
		cm.encryptionKey = keyData[:32]
	}

	// Load persisted credentials
	if err := cm.loadCredentials(); err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	cm.locked = false
	return nil
}

// Lock locks the credential manager and scrubs sensitive data from memory
func (cm *HPCCredentialManager) Lock() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Scrub encryption key
	if cm.encryptionKey != nil {
		for i := range cm.encryptionKey {
			cm.encryptionKey[i] = 0
		}
		cm.encryptionKey = nil
	}

	// Scrub signing key
	if cm.signingKey != nil {
		for i := range cm.signingKey {
			cm.signingKey[i] = 0
		}
		cm.signingKey = nil
	}

	// Scrub credential secrets
	for _, clusterCreds := range cm.credentials {
		for _, creds := range clusterCreds {
			creds.Password = ""
			creds.SSHPrivateKey = ""
			creds.SSHPassphrase = ""
			if creds.SigningKey != nil {
				for i := range creds.SigningKey {
					creds.SigningKey[i] = 0
				}
				creds.SigningKey = nil
			}
		}
	}

	cm.credentials = make(map[string]map[CredentialType]*HPCCredentials)
	cm.locked = true
}

// IsLocked returns true if the credential manager is locked
func (cm *HPCCredentialManager) IsLocked() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.locked
}

// StoreCredentials stores credentials for a cluster
func (cm *HPCCredentialManager) StoreCredentials(ctx context.Context, creds *HPCCredentials) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.locked {
		return errors.New("credential manager is locked")
	}

	if creds.ClusterID == "" {
		return errors.New("cluster_id is required")
	}

	// Initialize cluster map if needed
	if cm.credentials[creds.ClusterID] == nil {
		cm.credentials[creds.ClusterID] = make(map[CredentialType]*HPCCredentials)
	}

	// Store in memory
	creds.CreatedAt = time.Now()
	cm.credentials[creds.ClusterID][creds.Type] = creds

	// Persist to disk
	if err := cm.persistCredentials(); err != nil {
		return fmt.Errorf("failed to persist credentials: %w", err)
	}

	return nil
}

// GetCredentials retrieves credentials for a cluster
func (cm *HPCCredentialManager) GetCredentials(ctx context.Context, clusterID string, credType CredentialType) (*HPCCredentials, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.locked {
		return nil, errors.New("credential manager is locked")
	}

	clusterCreds, exists := cm.credentials[clusterID]
	if !exists {
		return nil, fmt.Errorf("no credentials found for cluster %s", clusterID)
	}

	creds, exists := clusterCreds[credType]
	if !exists {
		return nil, fmt.Errorf("no %s credentials found for cluster %s", credType, clusterID)
	}

	if creds.IsExpired() {
		return nil, fmt.Errorf("credentials for cluster %s have expired", clusterID)
	}

	return creds, nil
}

// DeleteCredentials deletes credentials for a cluster
func (cm *HPCCredentialManager) DeleteCredentials(ctx context.Context, clusterID string, credType CredentialType) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.locked {
		return errors.New("credential manager is locked")
	}

	clusterCreds, exists := cm.credentials[clusterID]
	if !exists {
		return nil
	}

	// Scrub the credential
	if creds, exists := clusterCreds[credType]; exists {
		creds.Password = ""
		creds.SSHPrivateKey = ""
		creds.SSHPassphrase = ""
		if creds.SigningKey != nil {
			for i := range creds.SigningKey {
				creds.SigningKey[i] = 0
			}
		}
	}

	delete(clusterCreds, credType)

	if len(clusterCreds) == 0 {
		delete(cm.credentials, clusterID)
	}

	return cm.persistCredentials()
}

// ListClusters returns all cluster IDs with stored credentials
func (cm *HPCCredentialManager) ListClusters() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusters := make([]string, 0, len(cm.credentials))
	for clusterID := range cm.credentials {
		clusters = append(clusters, clusterID)
	}
	return clusters
}

// RotateCredentials rotates credentials for a cluster
func (cm *HPCCredentialManager) RotateCredentials(ctx context.Context, clusterID string, credType CredentialType, newCreds *HPCCredentials) error {
	// Get old credentials for audit
	oldCreds, err := cm.GetCredentials(ctx, clusterID, credType)
	if err != nil {
		// No existing credentials, just store new ones
		return cm.StoreCredentials(ctx, newCreds)
	}

	// Store new credentials
	newCreds.Metadata = make(map[string]string)
	newCreds.Metadata["rotated_from"] = oldCreds.CreatedAt.Format(time.RFC3339)
	if err := cm.StoreCredentials(ctx, newCreds); err != nil {
		return err
	}

	return nil
}

// Sign signs data with the provider signing key
func (cm *HPCCredentialManager) Sign(data []byte) ([]byte, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.locked {
		return nil, errors.New("credential manager is locked")
	}

	if cm.signingKey == nil {
		return nil, errors.New("signing key not initialized")
	}

	return ed25519.Sign(cm.signingKey, data), nil
}

// Verify verifies a signature
func (cm *HPCCredentialManager) Verify(data []byte, signature []byte) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.publicKey == nil {
		return false
	}

	return ed25519.Verify(cm.publicKey, data, signature)
}

// GetPublicKey returns the public signing key
func (cm *HPCCredentialManager) GetPublicKey() ([]byte, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.publicKey == nil {
		return nil, errors.New("signing key not initialized")
	}

	return []byte(cm.publicKey), nil
}

// GenerateSigningKey generates a new signing key pair
func (cm *HPCCredentialManager) GenerateSigningKey() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.locked {
		return errors.New("credential manager is locked")
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate signing key: %w", err)
	}

	cm.signingKey = privKey
	cm.publicKey = pubKey

	return cm.persistCredentials()
}

// ImportSigningKey imports an existing signing key
func (cm *HPCCredentialManager) ImportSigningKey(privateKey []byte) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.locked {
		return errors.New("credential manager is locked")
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size: %d", len(privateKey))
	}

	cm.signingKey = ed25519.PrivateKey(privateKey)
	cm.publicKey = cm.signingKey.Public().(ed25519.PublicKey)

	return cm.persistCredentials()
}

// Persistence methods

type persistedData struct {
	Credentials map[string]map[string]*persistedCredential `json:"credentials"`
	SigningKey  string                                     `json:"signing_key,omitempty"`
	Version     int                                        `json:"version"`
}

type persistedCredential struct {
	Type                string            `json:"type"`
	ClusterID           string            `json:"cluster_id"`
	Username            string            `json:"username,omitempty"`
	EncryptedPassword   string            `json:"encrypted_password,omitempty"`
	EncryptedSSHKey     string            `json:"encrypted_ssh_key,omitempty"`
	SSHKeyPath          string            `json:"ssh_key_path,omitempty"`
	EncryptedPassphrase string            `json:"encrypted_passphrase,omitempty"`
	KerberosKeytab      string            `json:"kerberos_keytab,omitempty"`
	KerberosPrincipal   string            `json:"kerberos_principal,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	ExpiresAt           time.Time         `json:"expires_at,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

func (cm *HPCCredentialManager) persistCredentials() error {
	if cm.config.StorageDir == "" {
		return nil // No persistence configured
	}

	data := persistedData{
		Credentials: make(map[string]map[string]*persistedCredential),
		Version:     1,
	}

	// Convert credentials
	for clusterID, clusterCreds := range cm.credentials {
		data.Credentials[clusterID] = make(map[string]*persistedCredential)
		for credType, creds := range clusterCreds {
			pc := &persistedCredential{
				Type:              string(creds.Type),
				ClusterID:         creds.ClusterID,
				Username:          creds.Username,
				SSHKeyPath:        creds.SSHPrivateKeyPath,
				KerberosKeytab:    creds.KerberosKeytab,
				KerberosPrincipal: creds.KerberosPrincipal,
				CreatedAt:         creds.CreatedAt,
				ExpiresAt:         creds.ExpiresAt,
				Metadata:          creds.Metadata,
			}

			// Encrypt sensitive fields
			if creds.Password != "" {
				encrypted, err := cm.encrypt([]byte(creds.Password))
				if err != nil {
					return fmt.Errorf("failed to encrypt password: %w", err)
				}
				pc.EncryptedPassword = encrypted
			}

			if creds.SSHPrivateKey != "" {
				encrypted, err := cm.encrypt([]byte(creds.SSHPrivateKey))
				if err != nil {
					return fmt.Errorf("failed to encrypt SSH key: %w", err)
				}
				pc.EncryptedSSHKey = encrypted
			}

			if creds.SSHPassphrase != "" {
				encrypted, err := cm.encrypt([]byte(creds.SSHPassphrase))
				if err != nil {
					return fmt.Errorf("failed to encrypt passphrase: %w", err)
				}
				pc.EncryptedPassphrase = encrypted
			}

			data.Credentials[clusterID][string(credType)] = pc
		}
	}

	// Encrypt signing key
	if cm.signingKey != nil {
		encrypted, err := cm.encrypt([]byte(cm.signingKey))
		if err != nil {
			return fmt.Errorf("failed to encrypt signing key: %w", err)
		}
		data.SigningKey = encrypted
	}

	// Write to file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	filePath := filepath.Join(cm.config.StorageDir, "credentials.json")
	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

func (cm *HPCCredentialManager) loadCredentials() error {
	if cm.config.StorageDir == "" {
		return nil // No persistence configured
	}

	filePath := filepath.Join(cm.config.StorageDir, "credentials.json")
	fileData, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil // No credentials file yet
	}
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	var data persistedData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	// Convert credentials
	cm.credentials = make(map[string]map[CredentialType]*HPCCredentials)
	for clusterID, clusterCreds := range data.Credentials {
		cm.credentials[clusterID] = make(map[CredentialType]*HPCCredentials)
		for credTypeStr, pc := range clusterCreds {
			creds := &HPCCredentials{
				Type:              CredentialType(pc.Type),
				ClusterID:         pc.ClusterID,
				Username:          pc.Username,
				SSHPrivateKeyPath: pc.SSHKeyPath,
				KerberosKeytab:    pc.KerberosKeytab,
				KerberosPrincipal: pc.KerberosPrincipal,
				CreatedAt:         pc.CreatedAt,
				ExpiresAt:         pc.ExpiresAt,
				Metadata:          pc.Metadata,
			}

			// Decrypt sensitive fields
			if pc.EncryptedPassword != "" {
				decrypted, err := cm.decrypt(pc.EncryptedPassword)
				if err != nil {
					return fmt.Errorf("failed to decrypt password: %w", err)
				}
				creds.Password = string(decrypted)
			}

			if pc.EncryptedSSHKey != "" {
				decrypted, err := cm.decrypt(pc.EncryptedSSHKey)
				if err != nil {
					return fmt.Errorf("failed to decrypt SSH key: %w", err)
				}
				creds.SSHPrivateKey = string(decrypted)
			}

			if pc.EncryptedPassphrase != "" {
				decrypted, err := cm.decrypt(pc.EncryptedPassphrase)
				if err != nil {
					return fmt.Errorf("failed to decrypt passphrase: %w", err)
				}
				creds.SSHPassphrase = string(decrypted)
			}

			cm.credentials[clusterID][CredentialType(credTypeStr)] = creds
		}
	}

	// Decrypt signing key
	if data.SigningKey != "" {
		decrypted, err := cm.decrypt(data.SigningKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt signing key: %w", err)
		}
		if len(decrypted) == ed25519.PrivateKeySize {
			cm.signingKey = ed25519.PrivateKey(decrypted)
			cm.publicKey = cm.signingKey.Public().(ed25519.PublicKey)
		}
	}

	return nil
}

// Encryption helpers

func (cm *HPCCredentialManager) encrypt(plaintext []byte) (string, error) {
	if cm.encryptionKey == nil {
		if cm.config.AllowUnencrypted {
			return base64.StdEncoding.EncodeToString(plaintext), nil
		}
		return "", errors.New("encryption key not set")
	}

	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

func (cm *HPCCredentialManager) decrypt(ciphertext string) ([]byte, error) {
	if cm.encryptionKey == nil {
		if cm.config.AllowUnencrypted {
			return base64.StdEncoding.DecodeString(ciphertext)
		}
		return nil, errors.New("encryption key not set")
	}

	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// CredentialRotationWarning represents a warning about credential rotation
type CredentialRotationWarning struct {
	ClusterID string         `json:"cluster_id"`
	Type      CredentialType `json:"type"`
	ExpiresAt time.Time      `json:"expires_at"`
	DaysLeft  int            `json:"days_left"`
	Message   string         `json:"message"`
}

// CheckRotationWarnings checks for credentials that need rotation
func (cm *HPCCredentialManager) CheckRotationWarnings() []CredentialRotationWarning {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var warnings []CredentialRotationWarning
	warningThreshold := time.Now().Add(time.Duration(cm.config.RotationWarningDays) * 24 * time.Hour)

	for clusterID, clusterCreds := range cm.credentials {
		for _, creds := range clusterCreds {
			if creds.ExpiresAt.IsZero() {
				continue
			}

			if creds.ExpiresAt.Before(warningThreshold) {
				daysLeft := int(time.Until(creds.ExpiresAt).Hours() / 24)
				message := fmt.Sprintf("Credentials for cluster %s (%s) expire in %d days",
					clusterID, creds.Type, daysLeft)
				if daysLeft <= 0 {
					message = fmt.Sprintf("Credentials for cluster %s (%s) have expired",
						clusterID, creds.Type)
				}

				warnings = append(warnings, CredentialRotationWarning{
					ClusterID: clusterID,
					Type:      creds.Type,
					ExpiresAt: creds.ExpiresAt,
					DaysLeft:  daysLeft,
					Message:   message,
				})
			}
		}
	}

	return warnings
}

// CredentialHealth represents the health status of credentials
type CredentialHealth struct {
	ClusterID string         `json:"cluster_id"`
	Type      CredentialType `json:"type"`
	Valid     bool           `json:"valid"`
	Message   string         `json:"message"`
}

// CheckHealth checks the health of all stored credentials
func (cm *HPCCredentialManager) CheckHealth() []CredentialHealth {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Pre-calculate capacity for health slice
	totalCreds := 0
	for _, clusterCreds := range cm.credentials {
		totalCreds += len(clusterCreds)
	}
	health := make([]CredentialHealth, 0, totalCreds)

	for clusterID, clusterCreds := range cm.credentials {
		for _, creds := range clusterCreds {
			h := CredentialHealth{
				ClusterID: clusterID,
				Type:      creds.Type,
				Valid:     true,
				Message:   "OK",
			}

			if creds.IsExpired() {
				h.Valid = false
				h.Message = "Credentials expired"
			} else if creds.Type == CredentialTypeSLURM {
				// Check SSH credentials completeness
				hasAuth := creds.SSHPrivateKey != "" || creds.SSHPrivateKeyPath != "" || creds.Password != ""
				if !hasAuth {
					h.Valid = false
					h.Message = "No SSH authentication configured"
				}
			}

			health = append(health, h)
		}
	}

	return health
}
