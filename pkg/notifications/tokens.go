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
	"strings"
	"sync"
	"time"
)

var ErrEncryptionKeyRequired = errors.New("notification token encryption key required")
var ErrInvalidDeviceToken = errors.New("invalid notification device token")

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
	mu      sync.RWMutex
	vault   *TokenVault
	tokens  map[string][]tokenRecord
	timeNow func() time.Time
}

// NewInMemoryDeviceTokenStore creates a new device token store.
func NewInMemoryDeviceTokenStore(vault *TokenVault) *InMemoryDeviceTokenStore {
	return &InMemoryDeviceTokenStore{
		vault:  vault,
		tokens: make(map[string][]tokenRecord),
		timeNow: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Register adds or updates a device token.
func (s *InMemoryDeviceTokenStore) Register(_ context.Context, device DeviceToken) error {
	if s.vault == nil {
		return ErrEncryptionKeyRequired
	}
	now := s.timeNow()
	device, err := normalizeDeviceToken(device, now)
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(device.Token))
	tokenHash := base64.StdEncoding.EncodeToString(hash[:])
	ciphertext, err := s.vault.Encrypt(device.Token)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.tokens[device.UserAddress]
	for i, record := range records {
		if record.tokenHash == tokenHash {
			device.ID = firstNonEmpty(device.ID, record.device.ID)
			device.CreatedAt = pickTime(device.CreatedAt, record.device.CreatedAt)
			device.LastDeliveredAt = cloneTimePtr(record.device.LastDeliveredAt)
			device.ConsecutiveFailures = 0
			device.DisabledAt = nil
			device.Enabled = true
			device.LastFailureAt = cloneTimePtr(record.device.LastFailureAt)
			device.LastFailureReason = record.device.LastFailureReason
			record.device = cloneDeviceToken(device)
			record.tokenCipher = ciphertext
			record.tokenHash = tokenHash
			records[i] = record
			s.tokens[device.UserAddress] = records
			return nil
		}
	}

	s.tokens[device.UserAddress] = append(records, tokenRecord{
		device:      cloneDeviceToken(device),
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
	if s.vault == nil {
		return nil, ErrEncryptionKeyRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.tokens[userAddr]
	devices := make([]DeviceToken, 0, len(records))
	for i, record := range records {
		token, err := s.vault.Decrypt(record.tokenCipher)
		if err != nil {
			failedAt := s.timeNow()
			record.device.Enabled = false
			record.device.DisabledAt = cloneTimePtr(&failedAt)
			record.device.LastFailureAt = cloneTimePtr(&failedAt)
			record.device.LastFailureReason = "token decrypt failed"
			record.device.ConsecutiveFailures++
			records[i] = record
			continue
		}
		device := cloneDeviceToken(record.device)
		device.Token = token
		devices = append(devices, device)
	}
	return devices, nil
}

// RecordDeliverySuccess updates delivery state for a device token.
func (s *InMemoryDeviceTokenStore) RecordDeliverySuccess(_ context.Context, userAddr, token string, deliveredAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.updateRecordByToken(userAddr, token, func(device *DeviceToken) {
		device.Enabled = true
		device.DisabledAt = nil
		device.LastDeliveredAt = cloneTimePtr(&deliveredAt)
		device.LastSeenAt = deliveredAt.UTC()
		device.ConsecutiveFailures = 0
	})
	return nil
}

// RecordDeliveryFailure updates failure state for a device token.
func (s *InMemoryDeviceTokenStore) RecordDeliveryFailure(_ context.Context, userAddr, token, reason string, failedAt time.Time, disable bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.updateRecordByToken(userAddr, token, func(device *DeviceToken) {
		device.LastFailureAt = cloneTimePtr(&failedAt)
		device.LastFailureReason = strings.TrimSpace(reason)
		device.ConsecutiveFailures++
		if disable {
			device.Enabled = false
			device.DisabledAt = cloneTimePtr(&failedAt)
		}
	})
	return nil
}

func (s *InMemoryDeviceTokenStore) updateRecordByToken(userAddr, token string, apply func(device *DeviceToken)) {
	if userAddr == "" || token == "" {
		return
	}
	hash := sha256.Sum256([]byte(token))
	target := base64.StdEncoding.EncodeToString(hash[:])
	records := s.tokens[userAddr]
	for i, record := range records {
		if record.tokenHash != target {
			continue
		}
		device := cloneDeviceToken(record.device)
		apply(&device)
		record.device = device
		records[i] = record
		s.tokens[userAddr] = records
		return
	}
}

func normalizeDeviceToken(device DeviceToken, now time.Time) (DeviceToken, error) {
	device.UserAddress = strings.TrimSpace(device.UserAddress)
	device.Token = strings.TrimSpace(device.Token)
	device.AppID = strings.TrimSpace(device.AppID)
	device.Topic = strings.TrimSpace(device.Topic)
	if device.UserAddress == "" || device.Token == "" {
		return DeviceToken{}, fmt.Errorf("%w: missing user address or token", ErrInvalidDeviceToken)
	}
	switch device.Platform {
	case PlatformIOS, PlatformAndroid:
	default:
		return DeviceToken{}, fmt.Errorf("%w: unsupported platform %q", ErrInvalidDeviceToken, device.Platform)
	}
	if device.CreatedAt.IsZero() {
		device.CreatedAt = now.UTC()
	} else {
		device.CreatedAt = device.CreatedAt.UTC()
	}
	device.LastSeenAt = now.UTC()
	device.Enabled = true
	device.DisabledAt = nil
	return device, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func pickTime(preferred, fallback time.Time) time.Time {
	if !preferred.IsZero() {
		return preferred.UTC()
	}
	return fallback.UTC()
}
