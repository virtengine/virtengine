/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrMissingUser = errors.New("notification user address missing")

// DeliveryFailure captures a single channel failure during notification delivery.
type DeliveryFailure struct {
	Channel   Channel
	Target    string
	Err       error
	Retryable bool
}

func (f DeliveryFailure) Error() string {
	target := strings.TrimSpace(f.Target)
	if target == "" {
		return fmt.Sprintf("%s delivery failed: %v", f.Channel, f.Err)
	}
	return fmt.Sprintf("%s delivery failed for %s: %v", f.Channel, target, f.Err)
}

// DeliveryError aggregates channel failures for a notification.
type DeliveryError struct {
	NotificationID string
	Failures       []DeliveryFailure
}

func (e *DeliveryError) Error() string {
	if e == nil || len(e.Failures) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e.Failures))
	for _, failure := range e.Failures {
		parts = append(parts, failure.Error())
	}
	return fmt.Sprintf("notification %s delivery failed: %s", e.NotificationID, strings.Join(parts, "; "))
}

func (e *DeliveryError) Unwrap() error {
	if e == nil || len(e.Failures) == 0 {
		return nil
	}
	errs := make([]error, 0, len(e.Failures))
	for _, failure := range e.Failures {
		errs = append(errs, failure.Err)
	}
	return errors.Join(errs...)
}

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
	notif = cloneNotification(notif)
	notif.UserAddress = strings.TrimSpace(notif.UserAddress)
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
			prefs = normalizePreferences(stored)
		}
	}
	prefs = normalizePreferences(prefs)

	channels := append([]Channel(nil), notif.Channels...)
	if len(channels) == 0 {
		channels = append([]Channel(nil), prefs.Channels[notif.Type]...)
	}
	if len(channels) == 0 {
		channels = []Channel{ChannelInApp}
	}
	channels = normalizeChannels(channels)

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

	failures := make([]DeliveryFailure, 0, len(channels))
	if s.inApp != nil {
		if err := s.inApp.Publish(ctx, notif); err != nil {
			failures = append(failures, DeliveryFailure{
				Channel:   ChannelInApp,
				Target:    notif.UserAddress,
				Err:       err,
				Retryable: true,
			})
		}
	}

	if len(channels) == 0 {
		return buildDeliveryError(notif.ID, failures)
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
			failures = append(failures, s.sendPush(ctx, notif)...)
		case ChannelEmail:
			if s.email == nil {
				continue
			}
			if err := s.email.Send(ctx, notif, notif.UserAddress); err != nil {
				failures = append(failures, DeliveryFailure{
					Channel:   ChannelEmail,
					Target:    notif.UserAddress,
					Err:       err,
					Retryable: true,
				})
			}
		case ChannelInApp:
			// already stored/published
		}
	}

	return buildDeliveryError(notif.ID, failures)
}

// SendBatch sends multiple notifications.
func (s *DefaultService) SendBatch(ctx context.Context, notifs []Notification) error {
	var errs []error
	for _, notif := range notifs {
		if err := s.Send(ctx, notif); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
	return s.prefs.Put(ctx, normalizePreferences(prefs))
}

// GetPreferences returns preferences for a user.
func (s *DefaultService) GetPreferences(ctx context.Context, userAddr string) (Preferences, error) {
	if s.prefs == nil {
		return clonePreferences(s.defaultPrefs), nil
	}
	prefs, err := s.prefs.Get(ctx, userAddr)
	if err != nil {
		return clonePreferences(s.defaultPrefs), nil
	}
	return normalizePreferences(prefs), nil
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

func (s *DefaultService) sendPush(ctx context.Context, notif Notification) []DeliveryFailure {
	if notif.Topic != "" {
		if err := s.push.SendToTopic(ctx, notif.Topic, notif); err != nil {
			disposition := classifyPushError(err)
			return []DeliveryFailure{{
				Channel:   ChannelPush,
				Target:    notif.Topic,
				Err:       err,
				Retryable: disposition.retryable,
			}}
		}
		return nil
	}

	if s.devices == nil {
		return nil
	}
	devices, err := s.devices.List(ctx, notif.UserAddress)
	if err != nil {
		return []DeliveryFailure{{
			Channel:   ChannelPush,
			Target:    notif.UserAddress,
			Err:       err,
			Retryable: true,
		}}
	}
	recorder, _ := s.devices.(DeviceDeliveryRecorder)
	failures := make([]DeliveryFailure, 0, len(devices))
	successes := 0
	for _, device := range devices {
		if !device.Enabled {
			continue
		}
		var sendErr error
		if notif.Silent {
			sendErr = s.push.SendSilent(ctx, device, notif)
		} else {
			sendErr = s.push.SendToDevice(ctx, device, notif)
		}
		if sendErr != nil {
			disposition := classifyPushError(sendErr)
			if recorder != nil {
				_ = recorder.RecordDeliveryFailure(ctx, device.UserAddress, device.Token, sendErr.Error(), s.timeNow(), disposition.disableToken)
			}
			failures = append(failures, DeliveryFailure{
				Channel:   ChannelPush,
				Target:    describeDevice(device),
				Err:       sendErr,
				Retryable: disposition.retryable,
			})
			continue
		}
		successes++
		if recorder != nil {
			_ = recorder.RecordDeliverySuccess(ctx, device.UserAddress, device.Token, s.timeNow())
		}
	}

	if successes > 0 {
		return nil
	}
	return failures
}

func filterChannels(channels []Channel, allow func(Channel) bool) []Channel {
	filtered := make([]Channel, 0, len(channels))
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
	return normalizePreferences(defaultPreferencesSeed())
}

func defaultPreferencesSeed() Preferences {
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

func buildDeliveryError(notificationID string, failures []DeliveryFailure) error {
	if len(failures) == 0 {
		return nil
	}
	return &DeliveryError{
		NotificationID: notificationID,
		Failures:       failures,
	}
}

type pushFailureDisposition struct {
	retryable    bool
	disableToken bool
}

func classifyPushError(err error) pushFailureDisposition {
	if err == nil {
		return pushFailureDisposition{}
	}
	lower := strings.ToLower(err.Error())
	invalidMarkers := []string{
		"unregistered",
		"registration-token-not-registered",
		"notregistered",
		"baddevicetoken",
		"bad device token",
		"invalid token",
		"invalid registration token",
		"device token not for topic",
	}
	for _, marker := range invalidMarkers {
		if strings.Contains(lower, marker) {
			return pushFailureDisposition{disableToken: true}
		}
	}

	retryableMarkers := []string{
		"timeout",
		"temporarily unavailable",
		"unavailable",
		"deadline exceeded",
		"connection reset",
		"too many requests",
		"429",
		"500",
		"502",
		"503",
		"504",
	}
	for _, marker := range retryableMarkers {
		if strings.Contains(lower, marker) {
			return pushFailureDisposition{retryable: true}
		}
	}

	return pushFailureDisposition{}
}

func describeDevice(device DeviceToken) string {
	if device.ID != "" {
		return device.ID
	}
	parts := []string{string(device.Platform), device.AppID}
	description := strings.Join(parts, "/")
	description = strings.Trim(description, "/")
	if description == "" {
		return device.UserAddress
	}
	return description
}
