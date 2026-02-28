/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var ErrPreferencesNotFound = errors.New("notification preferences not found")

// InMemoryPreferencesStore stores preferences in memory.
type InMemoryPreferencesStore struct {
	mu    sync.RWMutex
	prefs map[string]Preferences
}

// NewInMemoryPreferencesStore creates a new preference store.
func NewInMemoryPreferencesStore() *InMemoryPreferencesStore {
	return &InMemoryPreferencesStore{
		prefs: make(map[string]Preferences),
	}
}

// Get returns preferences for a user.
func (s *InMemoryPreferencesStore) Get(_ context.Context, userAddr string) (Preferences, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefs, ok := s.prefs[userAddr]
	if !ok {
		return Preferences{}, ErrPreferencesNotFound
	}

	return clonePreferences(prefs), nil
}

// Put stores preferences for a user.
func (s *InMemoryPreferencesStore) Put(_ context.Context, prefs Preferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prefs[prefs.UserAddress] = normalizePreferences(prefs)
	return nil
}

func normalizePreferences(prefs Preferences) Preferences {
	defaults := defaultPreferencesSeed()
	normalized := Preferences{
		UserAddress:   strings.TrimSpace(prefs.UserAddress),
		Channels:      make(map[NotificationType][]Channel, len(defaults.Channels)),
		Frequencies:   make(map[NotificationType]DeliveryFrequency, len(defaults.Frequencies)),
		QuietHours:    normalizeQuietHours(prefs.QuietHours),
		DigestEnabled: prefs.DigestEnabled,
		DigestTime:    strings.TrimSpace(prefs.DigestTime),
	}
	if normalized.DigestTime == "" {
		normalized.DigestTime = defaults.DigestTime
	}

	for notifType, defaultChannels := range defaults.Channels {
		channels, ok := prefs.Channels[notifType]
		if !ok {
			channels = defaultChannels
		}
		normalized.Channels[notifType] = normalizeChannels(channels)
	}
	for notifType, channels := range prefs.Channels {
		if _, exists := normalized.Channels[notifType]; !exists {
			normalized.Channels[notifType] = normalizeChannels(channels)
		}
	}

	for notifType, defaultFrequency := range defaults.Frequencies {
		frequency, ok := prefs.Frequencies[notifType]
		if !ok {
			frequency = defaultFrequency
		}
		normalized.Frequencies[notifType] = normalizeFrequency(frequency)
	}
	for notifType, frequency := range prefs.Frequencies {
		if _, exists := normalized.Frequencies[notifType]; !exists {
			normalized.Frequencies[notifType] = normalizeFrequency(frequency)
		}
	}

	normalized.Channels[NotificationTypeSecurityAlert] = ensureChannels(normalized.Channels[NotificationTypeSecurityAlert], ChannelEmail, ChannelInApp)
	normalized.Frequencies[NotificationTypeSecurityAlert] = FrequencyImmediate

	return clonePreferences(normalized)
}

func normalizeChannels(channels []Channel) []Channel {
	if len(channels) == 0 {
		return nil
	}
	seen := make(map[Channel]struct{}, len(channels))
	normalized := make([]Channel, 0, len(channels))
	for _, channel := range channels {
		switch channel {
		case ChannelPush, ChannelEmail, ChannelInApp:
		default:
			continue
		}
		if _, exists := seen[channel]; exists {
			continue
		}
		seen[channel] = struct{}{}
		normalized = append(normalized, channel)
	}
	return normalized
}

func ensureChannels(channels []Channel, required ...Channel) []Channel {
	normalized := normalizeChannels(channels)
	for _, channel := range required {
		found := false
		for _, existing := range normalized {
			if existing == channel {
				found = true
				break
			}
		}
		if !found {
			normalized = append(normalized, channel)
		}
	}
	return normalized
}

func normalizeFrequency(frequency DeliveryFrequency) DeliveryFrequency {
	switch frequency {
	case FrequencyDigest:
		return FrequencyDigest
	case FrequencyImmediate:
		fallthrough
	default:
		return FrequencyImmediate
	}
}

func normalizeQuietHours(qh *QuietHours) *QuietHours {
	if qh == nil {
		return nil
	}
	normalized := *qh
	normalized.StartHour = normalizeHour(normalized.StartHour)
	normalized.EndHour = normalizeHour(normalized.EndHour)
	normalized.Timezone = strings.TrimSpace(normalized.Timezone)
	return &normalized
}

func normalizeHour(hour int) int {
	if hour < 0 {
		return 0
	}
	if hour > 23 {
		return 23
	}
	return hour
}
