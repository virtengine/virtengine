// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-920: Ansible Vault integration for encrypted variables
package provider_daemon

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

// Vault-specific errors
var (
	// ErrVaultDecryptionFailed is returned when vault decryption fails
	ErrVaultDecryptionFailed = errors.New("vault decryption failed")

	// ErrVaultEncryptionFailed is returned when vault encryption fails
	ErrVaultEncryptionFailed = errors.New("vault encryption failed")

	// ErrVaultInvalidFormat is returned when vault format is invalid
	ErrVaultInvalidFormat = errors.New("invalid vault format")

	// ErrVaultHMACMismatch is returned when HMAC verification fails
	ErrVaultHMACMismatch = errors.New("vault HMAC verification failed")

	// ErrVaultPasswordEmpty is returned when vault password is empty
	ErrVaultPasswordEmpty = errors.New("vault password cannot be empty")

	// ErrVaultUnsupportedVersion is returned for unsupported vault versions
	ErrVaultUnsupportedVersion = errors.New("unsupported vault version")
)

// Vault constants
const (
	// VaultHeader is the Ansible Vault header
	VaultHeader = "$ANSIBLE_VAULT"

	// VaultVersion11 is Ansible Vault version 1.1
	VaultVersion11 = "1.1"

	// VaultVersion12 is Ansible Vault version 1.2
	VaultVersion12 = "1.2"

	// VaultCipherAES256 is the AES256 cipher
	VaultCipherAES256 = "AES256"

	// vaultSaltSize is the salt size in bytes
	vaultSaltSize = 32

	// vaultKeyLength is the derived key length
	vaultKeyLength = 80

	// vaultPBKDF2Iterations is the number of PBKDF2 iterations
	vaultPBKDF2Iterations = 10000

	// vaultLineLength is the maximum line length for vault output
	vaultLineLength = 80
)

// VaultSecretID represents a vault secret identifier
type VaultSecretID string

// AnsibleVault provides Ansible Vault encryption/decryption capabilities
type AnsibleVault struct {
	// mu protects the password cache
	mu sync.RWMutex

	// passwordCache caches vault passwords by ID
	passwordCache map[VaultSecretID][]byte

	// defaultSecretID is the default secret ID
	defaultSecretID VaultSecretID
}

// NewAnsibleVault creates a new AnsibleVault instance
func NewAnsibleVault() *AnsibleVault {
	return &AnsibleVault{
		passwordCache:   make(map[VaultSecretID][]byte),
		defaultSecretID: "default",
	}
}

// SetPassword sets a vault password for a given secret ID
// WARNING: Password is stored in memory - ensure proper memory handling
func (v *AnsibleVault) SetPassword(secretID VaultSecretID, password string) error {
	if password == "" {
		return ErrVaultPasswordEmpty
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Store password securely (make a copy)
	v.passwordCache[secretID] = []byte(password)
	return nil
}

// SetDefaultPassword sets the password for the default secret ID
func (v *AnsibleVault) SetDefaultPassword(password string) error {
	return v.SetPassword(v.defaultSecretID, password)
}

// ClearPassword clears a vault password from memory
func (v *AnsibleVault) ClearPassword(secretID VaultSecretID) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if pw, ok := v.passwordCache[secretID]; ok {
		// Zero out the password bytes
		for i := range pw {
			pw[i] = 0
		}
		delete(v.passwordCache, secretID)
	}
}

// ClearAllPasswords clears all vault passwords from memory
func (v *AnsibleVault) ClearAllPasswords() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for id, pw := range v.passwordCache {
		for i := range pw {
			pw[i] = 0
		}
		delete(v.passwordCache, id)
	}
}

// getPassword retrieves a vault password
func (v *AnsibleVault) getPassword(secretID VaultSecretID) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if secretID == "" {
		secretID = v.defaultSecretID
	}

	pw, ok := v.passwordCache[secretID]
	if !ok {
		return nil, ErrVaultPasswordRequired
	}

	// Return a copy
	pwCopy := make([]byte, len(pw))
	copy(pwCopy, pw)
	return pwCopy, nil
}

// Encrypt encrypts data using Ansible Vault format
func (v *AnsibleVault) Encrypt(plaintext []byte, secretID VaultSecretID) (string, error) {
	password, err := v.getPassword(secretID)
	if err != nil {
		return "", err
	}
	defer clearBytes(password)

	return v.EncryptWithPassword(plaintext, password, secretID)
}

// EncryptWithPassword encrypts data using the provided password
func (v *AnsibleVault) EncryptWithPassword(plaintext, password []byte, secretID VaultSecretID) (string, error) {
	if len(password) == 0 {
		return "", ErrVaultPasswordEmpty
	}

	// Generate salt
	salt := make([]byte, vaultSaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("%w: failed to generate salt: %v", ErrVaultEncryptionFailed, err)
	}

	// Derive keys using PBKDF2
	derivedKey := pbkdf2.Key(password, salt, vaultPBKDF2Iterations, vaultKeyLength, sha256.New)
	defer clearBytes(derivedKey)

	aesKey := derivedKey[:32]
	hmacKey := derivedKey[32:64]
	iv := derivedKey[64:80]

	// Pad plaintext to AES block size
	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)
	defer clearBytes(paddedPlaintext)

	// Encrypt with AES-256-CTR
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("%w: failed to create cipher: %v", ErrVaultEncryptionFailed, err)
	}

	ciphertext := make([]byte, len(paddedPlaintext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, paddedPlaintext)

	// Calculate HMAC
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(ciphertext)
	hmacValue := mac.Sum(nil)

	// Build vault payload: salt + hmac + ciphertext (all hex encoded)
	payload := hex.EncodeToString(salt) + "\n" + hex.EncodeToString(hmacValue) + "\n" + hex.EncodeToString(ciphertext)
	payloadHex := hex.EncodeToString([]byte(payload))

	// Build vault header
	var header string
	if secretID != "" && secretID != v.defaultSecretID {
		header = fmt.Sprintf("%s;%s;%s;%s", VaultHeader, VaultVersion12, VaultCipherAES256, secretID)
	} else {
		header = fmt.Sprintf("%s;%s;%s", VaultHeader, VaultVersion11, VaultCipherAES256)
	}

	// Format output with line wrapping
	return formatVaultOutput(header, payloadHex), nil
}

// Decrypt decrypts Ansible Vault encrypted data
func (v *AnsibleVault) Decrypt(vaultText string) ([]byte, error) {
	header, payload, err := parseVaultText(vaultText)
	if err != nil {
		return nil, err
	}

	secretID := extractSecretID(header)
	password, err := v.getPassword(secretID)
	if err != nil {
		return nil, err
	}
	defer clearBytes(password)

	return v.DecryptWithPassword(payload, password)
}

// DecryptWithPassword decrypts Ansible Vault data using the provided password
func (v *AnsibleVault) DecryptWithPassword(payload string, password []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrVaultPasswordEmpty
	}

	// Decode hex payload
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode payload: %v", ErrVaultInvalidFormat, err)
	}

	// Parse payload lines
	parts := strings.Split(string(payloadBytes), "\n")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: expected 3 parts, got %d", ErrVaultInvalidFormat, len(parts))
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode salt: %v", ErrVaultInvalidFormat, err)
	}

	storedHMAC, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode HMAC: %v", ErrVaultInvalidFormat, err)
	}

	ciphertext, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode ciphertext: %v", ErrVaultInvalidFormat, err)
	}

	// Derive keys using PBKDF2
	derivedKey := pbkdf2.Key(password, salt, vaultPBKDF2Iterations, vaultKeyLength, sha256.New)
	defer clearBytes(derivedKey)

	aesKey := derivedKey[:32]
	hmacKey := derivedKey[32:64]
	iv := derivedKey[64:80]

	// Verify HMAC
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(ciphertext)
	calculatedHMAC := mac.Sum(nil)

	if !hmac.Equal(storedHMAC, calculatedHMAC) {
		return nil, ErrVaultHMACMismatch
	}

	// Decrypt with AES-256-CTR
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create cipher: %v", ErrVaultDecryptionFailed, err)
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	// Remove PKCS7 padding
	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unpad: %v", ErrVaultDecryptionFailed, err)
	}

	return unpadded, nil
}

// EncryptString encrypts a string value
func (v *AnsibleVault) EncryptString(plaintext string) (string, error) {
	return v.Encrypt([]byte(plaintext), "")
}

// DecryptString decrypts a vault string to plaintext
func (v *AnsibleVault) DecryptString(vaultText string) (string, error) {
	plaintext, err := v.Decrypt(vaultText)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// IsVaultEncrypted checks if text is Ansible Vault encrypted
func IsVaultEncrypted(text string) bool {
	text = strings.TrimSpace(text)
	return strings.HasPrefix(text, VaultHeader)
}

// parseVaultText parses vault text into header and payload
func parseVaultText(vaultText string) (header, payload string, err error) {
	vaultText = strings.TrimSpace(vaultText)
	lines := strings.Split(vaultText, "\n")
	if len(lines) < 2 {
		return "", "", ErrVaultInvalidFormat
	}

	header = strings.TrimSpace(lines[0])
	if !strings.HasPrefix(header, VaultHeader) {
		return "", "", ErrVaultInvalidFormat
	}

	// Validate header format
	parts := strings.Split(header, ";")
	if len(parts) < 3 {
		return "", "", ErrVaultInvalidFormat
	}

	version := parts[1]
	if version != VaultVersion11 && version != VaultVersion12 {
		return "", "", fmt.Errorf("%w: %s", ErrVaultUnsupportedVersion, version)
	}

	cipherName := parts[2]
	if cipherName != VaultCipherAES256 {
		return "", "", fmt.Errorf("%w: unsupported cipher: %s", ErrVaultInvalidFormat, cipherName)
	}

	// Concatenate payload lines
	var payloadBuilder strings.Builder
	for i := 1; i < len(lines); i++ {
		payloadBuilder.WriteString(strings.TrimSpace(lines[i]))
	}

	return header, payloadBuilder.String(), nil
}

// extractSecretID extracts the secret ID from the vault header
func extractSecretID(header string) VaultSecretID {
	parts := strings.Split(header, ";")
	if len(parts) >= 4 {
		return VaultSecretID(parts[3])
	}
	return ""
}

// formatVaultOutput formats the vault output with proper line wrapping
func formatVaultOutput(header, payload string) string {
	var buf bytes.Buffer
	buf.WriteString(header)
	buf.WriteString("\n")

	for i := 0; i < len(payload); i += vaultLineLength {
		end := i + vaultLineLength
		if end > len(payload) {
			end = len(payload)
		}
		buf.WriteString(payload[i:end])
		buf.WriteString("\n")
	}

	return strings.TrimRight(buf.String(), "\n")
}

// pkcs7Pad pads data to the specified block size using PKCS7
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, errors.New("invalid padding")
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}

// clearBytes zeroes out a byte slice
func clearBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// VaultVariables provides a map-like interface for vault-encrypted variables
type VaultVariables struct {
	vault     *AnsibleVault
	variables map[string]interface{}
	encrypted map[string]bool
	mu        sync.RWMutex
}

// NewVaultVariables creates a new VaultVariables instance
func NewVaultVariables(vault *AnsibleVault) *VaultVariables {
	return &VaultVariables{
		vault:     vault,
		variables: make(map[string]interface{}),
		encrypted: make(map[string]bool),
	}
}

// Set sets a variable value
func (vv *VaultVariables) Set(key string, value interface{}) {
	vv.mu.Lock()
	defer vv.mu.Unlock()
	vv.variables[key] = value
}

// SetEncrypted sets an encrypted variable value
func (vv *VaultVariables) SetEncrypted(key string, value string) error {
	encrypted, err := vv.vault.EncryptString(value)
	if err != nil {
		return err
	}

	vv.mu.Lock()
	defer vv.mu.Unlock()
	vv.variables[key] = encrypted
	vv.encrypted[key] = true
	return nil
}

// Get retrieves a variable value, decrypting if necessary
func (vv *VaultVariables) Get(key string) (interface{}, error) {
	vv.mu.RLock()
	defer vv.mu.RUnlock()

	value, ok := vv.variables[key]
	if !ok {
		return nil, nil
	}

	if vv.encrypted[key] {
		strVal, ok := value.(string)
		if !ok {
			return nil, errors.New("encrypted value is not a string")
		}
		decrypted, err := vv.vault.DecryptString(strVal)
		if err != nil {
			return nil, err
		}
		return decrypted, nil
	}

	return value, nil
}

// GetEncrypted retrieves the encrypted form of a variable
func (vv *VaultVariables) GetEncrypted(key string) (string, bool) {
	vv.mu.RLock()
	defer vv.mu.RUnlock()

	if !vv.encrypted[key] {
		return "", false
	}

	strVal, ok := vv.variables[key].(string)
	return strVal, ok
}

// ToMap returns all variables as a map (encrypted values remain encrypted)
func (vv *VaultVariables) ToMap() map[string]interface{} {
	vv.mu.RLock()
	defer vv.mu.RUnlock()

	result := make(map[string]interface{}, len(vv.variables))
	for k, v := range vv.variables {
		result[k] = v
	}
	return result
}

// Keys returns all variable keys
func (vv *VaultVariables) Keys() []string {
	vv.mu.RLock()
	defer vv.mu.RUnlock()

	keys := make([]string, 0, len(vv.variables))
	for k := range vv.variables {
		keys = append(keys, k)
	}
	return keys
}

// IsEncrypted checks if a variable is encrypted
func (vv *VaultVariables) IsEncrypted(key string) bool {
	vv.mu.RLock()
	defer vv.mu.RUnlock()
	return vv.encrypted[key]
}

// DecryptAll returns all variables with encrypted values decrypted
func (vv *VaultVariables) DecryptAll() (map[string]interface{}, error) {
	vv.mu.RLock()
	defer vv.mu.RUnlock()

	result := make(map[string]interface{}, len(vv.variables))
	for k, v := range vv.variables {
		if vv.encrypted[k] {
			strVal, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("encrypted value for key %s is not a string", k)
			}
			decrypted, err := vv.vault.DecryptString(strVal)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt key %s: %w", k, err)
			}
			result[k] = decrypted
		} else {
			result[k] = v
		}
	}
	return result, nil
}
