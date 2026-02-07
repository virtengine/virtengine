/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var ErrEncryptionKeyRequired = errors.New("notification token encryption key required")

// TokenVault encrypts and decrypts device tokens at rest.
type TokenVault struct {
	gcm cipher.AEAD
}

// NewTokenVault creates a vault using a raw 32-byte key.
func NewTokenVault(key []byte) (*TokenVault, error) {
	if len(key) == 0 {
		return nil, ErrEncryptionKeyRequired
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cipher init: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm init: %w", err)
	}
	return &TokenVault{gcm: gcm}, nil
}

// NewTokenVaultFromBase64 creates a vault from a base64-encoded key.
func NewTokenVaultFromBase64(key string) (*TokenVault, error) {
	raw, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	return NewTokenVault(raw)
}

// Encrypt encrypts a device token.
func (v *TokenVault) Encrypt(token string) (string, error) {
	nonce := make([]byte, v.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	ciphertext := v.gcm.Seal(nil, nonce, []byte(token), nil)
	payload := make([]byte, 0, len(nonce)+len(ciphertext))
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

// Decrypt decrypts a device token.
func (v *TokenVault) Decrypt(payload string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("decode payload: %w", err)
	}
	if len(raw) < v.gcm.NonceSize() {
		return "", errors.New("payload too short")
	}
	nonce := raw[:v.gcm.NonceSize()]
	ciphertext := raw[v.gcm.NonceSize():]
	plain, err := v.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}

type tokenRecord struct {
	device      DeviceToken
	tokenCipher string
	tokenHash   string
}

// InMemoryDeviceTokenStore stores encrypted tokens in memory.
type InMemoryDeviceTokenStore struct {
	mu     sync.RWMutex
	vault  *TokenVault
	tokens map[string][]tokenRecord
}

// NewInMemoryDeviceTokenStore creates a new device token store.
func NewInMemoryDeviceTokenStore(vault *TokenVault) *InMemoryDeviceTokenStore {
	return &InMemoryDeviceTokenStore{
		vault:  vault,
		tokens: make(map[string][]tokenRecord),
	}
}

// Register adds or updates a device token.
func (s *InMemoryDeviceTokenStore) Register(_ context.Context, device DeviceToken) error {
	if s.vault == nil {
		return ErrEncryptionKeyRequired
	}
	if device.UserAddress == "" || device.Token == "" {
		return errors.New("device token missing user address or token")
	}

	hash := sha256.Sum256([]byte(device.Token))
	tokenHash := base64.StdEncoding.EncodeToString(hash[:])
	ciphertext, err := s.vault.Encrypt(device.Token)
	if err != nil {
		return err
	}

	device.CreatedAt = device.CreatedAt.UTC()
	if device.CreatedAt.IsZero() {
		device.CreatedAt = time.Now().UTC()
	}
	device.LastSeenAt = time.Now().UTC()
	device.Enabled = true

	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.tokens[device.UserAddress]
	for i, record := range records {
		if record.tokenHash == tokenHash {
			record.device = device
			record.tokenCipher = ciphertext
			record.tokenHash = tokenHash
			records[i] = record
			s.tokens[device.UserAddress] = records
			return nil
		}
	}

	s.tokens[device.UserAddress] = append(records, tokenRecord{
		device:      device,
		tokenCipher: ciphertext,
		tokenHash:   tokenHash,
	})

	return nil
}

// Unregister removes a device token.
func (s *InMemoryDeviceTokenStore) Unregister(_ context.Context, userAddr, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.tokens[userAddr]
	if len(records) == 0 {
		return nil
	}

	hash := sha256.Sum256([]byte(token))
	target := base64.StdEncoding.EncodeToString(hash[:])
	updated := records[:0]
	for _, record := range records {
		if record.tokenHash != target {
			updated = append(updated, record)
		}
	}
	s.tokens[userAddr] = updated
	return nil
}

// List returns device tokens for a user with decrypted values.
func (s *InMemoryDeviceTokenStore) List(_ context.Context, userAddr string) ([]DeviceToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.tokens[userAddr]
	devices := make([]DeviceToken, 0, len(records))
	for _, record := range records {
		token, err := s.vault.Decrypt(record.tokenCipher)
		if err != nil {
			return nil, err
		}
		device := record.device
		device.Token = token
		devices = append(devices, device)
	}
	return devices, nil
}
