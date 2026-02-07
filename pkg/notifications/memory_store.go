package notifications

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrPreferencesNotFound indicates missing preferences.
var ErrPreferencesNotFound = fmt.Errorf("preferences not found")

// MemoryStore is an in-memory implementation of the notification stores.
type MemoryStore struct {
	mu            sync.RWMutex
	notifications map[string][]Notification
	prefs         map[string]Preferences
	devices       map[string][]DeviceRegistration
	cipher        TokenCipher
}

// NewMemoryStore creates a memory store with optional token cipher.
func NewMemoryStore(cipher TokenCipher) *MemoryStore {
	if cipher == nil {
		cipher = NoopCipher{}
	}
	return &MemoryStore{
		notifications: make(map[string][]Notification),
		prefs:         make(map[string]Preferences),
		devices:       make(map[string][]DeviceRegistration),
		cipher:        cipher,
	}
}

// Save persists a notification.
func (m *MemoryStore) Save(ctx context.Context, notification Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications[notification.UserAddress] = append(m.notifications[notification.UserAddress], notification)
	return nil
}

// SaveBatch persists multiple notifications.
func (m *MemoryStore) SaveBatch(ctx context.Context, notifications []Notification) error {
	for _, notification := range notifications {
		if err := m.Save(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// List returns notifications and unread count.
func (m *MemoryStore) List(ctx context.Context, userAddress string, opts ListOptions) ([]Notification, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := m.notifications[userAddress]
	filtered := make([]Notification, 0, len(items))
	unreadCount := 0

	for _, item := range items {
		if item.ReadAt == nil {
			unreadCount++
		}
		if !opts.IncludeRead && item.ReadAt != nil {
			continue
		}
		if len(opts.Types) > 0 && !containsNotificationType(opts.Types, item.Type) {
			continue
		}
		if opts.Since != nil && item.CreatedAt.Before(*opts.Since) {
			continue
		}
		if opts.Until != nil && item.CreatedAt.After(*opts.Until) {
			continue
		}
		filtered = append(filtered, item)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	start := opts.Offset
	if start < 0 {
		start = 0
	}
	end := len(filtered)
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}
	if start > len(filtered) {
		return []Notification{}, unreadCount, nil
	}

	return filtered[start:end], unreadCount, nil
}

// MarkAsRead marks notifications as read.
func (m *MemoryStore) MarkAsRead(ctx context.Context, userAddress string, notificationIDs []string) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	items := m.notifications[userAddress]
	for idx, item := range items {
		if item.ReadAt != nil {
			continue
		}
		if containsString(notificationIDs, item.ID) {
			item.ReadAt = &now
			items[idx] = item
		}
	}
	m.notifications[userAddress] = items
	return nil
}

// UpdatePreferences updates preferences.
func (m *MemoryStore) UpdatePreferences(ctx context.Context, userAddress string, prefs Preferences) error {
	if err := prefs.Validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prefs[userAddress] = prefs
	return nil
}

// GetPreferences returns stored preferences.
func (m *MemoryStore) GetPreferences(ctx context.Context, userAddress string) (Preferences, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefs, ok := m.prefs[userAddress]
	if !ok {
		return Preferences{}, ErrPreferencesNotFound
	}
	return prefs, nil
}

// RegisterDevice registers a device token.
func (m *MemoryStore) RegisterDevice(ctx context.Context, userAddress string, registration DeviceRegistration) (DeviceRegistration, error) {
	if registration.ID == "" {
		registration.ID = uuid.NewString()
	}
	if registration.CreatedAt == 0 {
		registration.CreatedAt = time.Now().UTC().Unix()
	}

	encrypted, err := m.cipher.Encrypt(registration.Token)
	if err != nil {
		return DeviceRegistration{}, err
	}
	registration.Token = encrypted

	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[userAddress] = append(m.devices[userAddress], registration)
	return registration, nil
}

// ListDevices returns device registrations.
func (m *MemoryStore) ListDevices(ctx context.Context, userAddress string) ([]DeviceRegistration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	devices := m.devices[userAddress]
	results := make([]DeviceRegistration, 0, len(devices))
	for _, device := range devices {
		token, err := m.cipher.Decrypt(device.Token)
		if err != nil {
			return nil, err
		}
		device.Token = token
		results = append(results, device)
	}
	return results, nil
}

// UpdateDevice updates a device registration.
func (m *MemoryStore) UpdateDevice(ctx context.Context, userAddress string, registration DeviceRegistration) error {
	if registration.ID == "" {
		return fmt.Errorf("device id required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	devices := m.devices[userAddress]
	for idx, device := range devices {
		if device.ID != registration.ID {
			continue
		}
		encrypted, err := m.cipher.Encrypt(registration.Token)
		if err != nil {
			return err
		}
		registration.Token = encrypted
		devices[idx] = registration
		m.devices[userAddress] = devices
		return nil
	}
	return fmt.Errorf("device not found")
}

// RemoveDevice removes a device registration.
func (m *MemoryStore) RemoveDevice(ctx context.Context, userAddress string, deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	devices := m.devices[userAddress]
	filtered := devices[:0]
	for _, device := range devices {
		if device.ID == deviceID {
			continue
		}
		filtered = append(filtered, device)
	}
	m.devices[userAddress] = filtered
	return nil
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func containsNotificationType(values []NotificationType, value NotificationType) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
