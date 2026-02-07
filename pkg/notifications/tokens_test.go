package notifications

import (
	"context"
	"crypto/rand"
	"testing"
)

func TestTokenVaultRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand key: %v", err)
	}
	vault, err := NewTokenVault(key)
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	ciphertext, err := vault.Encrypt("device-token-123")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plain, err := vault.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != "device-token-123" {
		t.Fatalf("unexpected plaintext: %s", plain)
	}
}

func TestDeviceTokenStoreRegisterList(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand key: %v", err)
	}
	vault, err := NewTokenVault(key)
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	store := NewInMemoryDeviceTokenStore(vault)

	device := DeviceToken{
		UserAddress: "user1",
		Token:       "token-abc",
		Platform:    PlatformIOS,
		AppID:       "com.virtengine.portal",
	}
	if err := store.Register(context.Background(), device); err != nil {
		t.Fatalf("register: %v", err)
	}

	devices, err := store.List(context.Background(), "user1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Token != "token-abc" {
		t.Fatalf("token mismatch")
	}
}
