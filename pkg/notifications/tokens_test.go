package notifications

import (
	"context"
	"crypto/rand"
	"testing"
	"time"
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
	vault := newTestVault(t)
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

func TestDeviceTokenStoreReRegisterClearsDisabledState(t *testing.T) {
	vault := newTestVault(t)
	store := NewInMemoryDeviceTokenStore(vault)

	device := DeviceToken{
		UserAddress: "user2",
		Token:       "token-rotate",
		Platform:    PlatformIOS,
		AppID:       "com.virtengine.portal",
	}
	if err := store.Register(context.Background(), device); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := store.RecordDeliveryFailure(context.Background(), "user2", "token-rotate", "invalid token", timeNowForTests(), true); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	if err := store.Register(context.Background(), device); err != nil {
		t.Fatalf("re-register: %v", err)
	}

	devices, err := store.List(context.Background(), "user2")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected one device, got %d", len(devices))
	}
	if !devices[0].Enabled || devices[0].DisabledAt != nil || devices[0].ConsecutiveFailures != 0 {
		t.Fatalf("expected re-registration to reactivate token, got %+v", devices[0])
	}
}

func TestDeviceTokenStoreListSkipsCorruptToken(t *testing.T) {
	vault := newTestVault(t)
	store := NewInMemoryDeviceTokenStore(vault)

	if err := store.Register(context.Background(), DeviceToken{
		UserAddress: "user3",
		Token:       "token-good",
		Platform:    PlatformAndroid,
		AppID:       "com.virtengine.portal",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	store.mu.Lock()
	store.tokens["user3"] = append(store.tokens["user3"], tokenRecord{
		device: DeviceToken{
			UserAddress: "user3",
			Platform:    PlatformIOS,
			AppID:       "com.virtengine.portal",
			Enabled:     true,
		},
		tokenCipher: "corrupt",
		tokenHash:   "corrupt",
	})
	store.mu.Unlock()

	devices, err := store.List(context.Background(), "user3")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(devices) != 1 || devices[0].Token != "token-good" {
		t.Fatalf("expected corrupt token to be skipped, got %+v", devices)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	if len(store.tokens["user3"]) != 2 {
		t.Fatalf("expected corrupt record to remain for audit")
	}
	if store.tokens["user3"][1].device.Enabled {
		t.Fatalf("expected corrupt record to be disabled after failed decrypt")
	}
	if store.tokens["user3"][1].device.LastFailureReason != "token decrypt failed" {
		t.Fatalf("expected decrypt failure reason, got %q", store.tokens["user3"][1].device.LastFailureReason)
	}
}

func newTestVault(t *testing.T) *TokenVault {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand key: %v", err)
	}
	vault, err := NewTokenVault(key)
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	return vault
}

func timeNowForTests() time.Time {
	return time.Date(2026, 4, 11, 8, 0, 0, 0, time.UTC)
}
