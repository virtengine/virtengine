/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrMissingUser = errors.New("notification user address missing")

// DefaultService implements the notification service.
type DefaultService struct {
	store        Store
	prefs        PreferencesStore
	devices      DeviceTokenStore
	push         PushClient
	email        EmailSender
	inApp        InAppPublisher
	timeNow      func() time.Time
	defaultPrefs Preferences
}

// NewDefaultService creates a notification service.
func NewDefaultService(store Store, prefs PreferencesStore, devices DeviceTokenStore, push PushClient, email EmailSender, inApp InAppPublisher) *DefaultService {
	return &DefaultService{
		store:   store,
		prefs:   prefs,
		devices: devices,
		push:    push,
		email:   email,
		inApp:   inApp,
		timeNow: func() time.Time {
			return time.Now().UTC()
		},
		defaultPrefs: DefaultPreferences(),
	}
}

// Send delivers a notification based on user preferences.
func (s *DefaultService) Send(ctx context.Context, notif Notification) error {
	if notif.UserAddress == "" {
		return ErrMissingUser
	}
	if notif.ID == "" {
		notif.ID = fmt.Sprintf("notif_%d", s.timeNow().UnixNano())
	}
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = s.timeNow()
	}

	prefs := s.defaultPrefs
	if s.prefs != nil {
		if stored, err := s.prefs.Get(ctx, notif.UserAddress); err == nil {
			prefs = stored
		}
	}

	channels := notif.Channels
	if len(channels) == 0 {
		channels = prefs.Channels[notif.Type]
	}
	if len(channels) == 0 {
		channels = []Channel{ChannelInApp}
	}

	frequency := prefs.Frequencies[notif.Type]
	if frequency == "" {
		frequency = FrequencyImmediate
	}
	if prefs.DigestEnabled && frequency == FrequencyDigest && notif.Type != NotificationTypeSecurityAlert {
		channels = filterChannels(channels, func(ch Channel) bool { return ch == ChannelInApp })
	}

	// Store in-app notification regardless of delivery channel.
	if s.store != nil {
		if err := s.store.Add(ctx, notif); err != nil {
			return err
		}
	}

	if s.inApp != nil {
		_ = s.inApp.Publish(ctx, notif)
	}

	if len(channels) == 0 {
		return nil
	}

	if isQuietHours(prefs.QuietHours, s.timeNow(), notif.Type) {
		channels = filterChannels(channels, func(ch Channel) bool { return ch == ChannelInApp })
	}

	for _, channel := range channels {
		switch channel {
		case ChannelPush:
			if s.push == nil {
				continue
			}
			if err := s.sendPush(ctx, notif); err != nil {
				return err
			}
		case ChannelEmail:
			if s.email == nil {
				continue
			}
			if err := s.email.Send(ctx, notif, notif.UserAddress); err != nil {
				return err
			}
		case ChannelInApp:
			// already stored/published
		}
	}

	return nil
}

// SendBatch sends multiple notifications.
func (s *DefaultService) SendBatch(ctx context.Context, notifs []Notification) error {
	for _, notif := range notifs {
		if err := s.Send(ctx, notif); err != nil {
			return err
		}
	}
	return nil
}

// GetUserNotifications returns notifications for a user.
func (s *DefaultService) GetUserNotifications(ctx context.Context, userAddr string, opts ListOptions) ([]Notification, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.List(ctx, userAddr, opts)
}

// MarkAsRead marks notifications as read.
func (s *DefaultService) MarkAsRead(ctx context.Context, userAddr string, notifIDs []string) error {
	if s.store == nil {
		return nil
	}
	return s.store.MarkRead(ctx, userAddr, notifIDs)
}

// UpdatePreferences updates preferences for a user.
func (s *DefaultService) UpdatePreferences(ctx context.Context, userAddr string, prefs Preferences) error {
	if s.prefs == nil {
		return nil
	}
	prefs.UserAddress = userAddr
	return s.prefs.Put(ctx, prefs)
}

// GetPreferences returns preferences for a user.
func (s *DefaultService) GetPreferences(ctx context.Context, userAddr string) (Preferences, error) {
	if s.prefs == nil {
		return s.defaultPrefs, nil
	}
	prefs, err := s.prefs.Get(ctx, userAddr)
	if err != nil {
		return s.defaultPrefs, nil
	}
	return prefs, nil
}

// RegisterDevice stores a device token.
func (s *DefaultService) RegisterDevice(ctx context.Context, device DeviceToken) error {
	if s.devices == nil {
		return nil
	}
	return s.devices.Register(ctx, device)
}

// UnregisterDevice removes a device token.
func (s *DefaultService) UnregisterDevice(ctx context.Context, userAddr, token string) error {
	if s.devices == nil {
		return nil
	}
	return s.devices.Unregister(ctx, userAddr, token)
}

// ListDevices lists user device tokens.
func (s *DefaultService) ListDevices(ctx context.Context, userAddr string) ([]DeviceToken, error) {
	if s.devices == nil {
		return nil, nil
	}
	return s.devices.List(ctx, userAddr)
}

func (s *DefaultService) sendPush(ctx context.Context, notif Notification) error {
	if notif.Topic != "" {
		return s.push.SendToTopic(ctx, notif.Topic, notif)
	}

	if s.devices == nil {
		return nil
	}
	devices, err := s.devices.List(ctx, notif.UserAddress)
	if err != nil {
		return err
	}
	for _, device := range devices {
		if !device.Enabled {
			continue
		}
		if notif.Silent {
			if err := s.push.SendSilent(ctx, device, notif); err != nil {
				return err
			}
			continue
		}
		if err := s.push.SendToDevice(ctx, device, notif); err != nil {
			return err
		}
	}

	return nil
}

func filterChannels(channels []Channel, allow func(Channel) bool) []Channel {
	filtered := channels[:0]
	for _, ch := range channels {
		if allow(ch) {
			filtered = append(filtered, ch)
		}
	}
	return filtered
}

func isQuietHours(qh *QuietHours, now time.Time, notifType NotificationType) bool {
	if qh == nil || !qh.Enabled || notifType == NotificationTypeSecurityAlert {
		return false
	}
	location := time.UTC
	if qh.Timezone != "" {
		if tz, err := time.LoadLocation(qh.Timezone); err == nil {
			location = tz
		}
	}
	local := now.In(location)
	hour := local.Hour()

	if qh.StartHour == qh.EndHour {
		return false
	}
	if qh.StartHour < qh.EndHour {
		return hour >= qh.StartHour && hour < qh.EndHour
	}
	return hour >= qh.StartHour || hour < qh.EndHour
}

// DefaultPreferences provides sensible defaults for new users.
func DefaultPreferences() Preferences {
	return Preferences{
		Channels: map[NotificationType][]Channel{
			NotificationTypeVEIDStatus:    {ChannelEmail, ChannelPush, ChannelInApp},
			NotificationTypeOrderUpdate:   {ChannelPush, ChannelInApp},
			NotificationTypeEscrowDeposit: {ChannelEmail, ChannelInApp},
			NotificationTypeSecurityAlert: {ChannelEmail, ChannelPush, ChannelInApp},
			NotificationTypeProviderAlert: {ChannelPush, ChannelInApp},
		},
		Frequencies: map[NotificationType]DeliveryFrequency{
			NotificationTypeVEIDStatus:    FrequencyImmediate,
			NotificationTypeOrderUpdate:   FrequencyImmediate,
			NotificationTypeEscrowDeposit: FrequencyImmediate,
			NotificationTypeSecurityAlert: FrequencyImmediate,
			NotificationTypeProviderAlert: FrequencyImmediate,
		},
		DigestEnabled: false,
		DigestTime:    "09:00",
	}
}
