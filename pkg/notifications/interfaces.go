/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import "context"

// Service defines the notification service contract.
type Service interface {
	Send(ctx context.Context, notif Notification) error
	SendBatch(ctx context.Context, notifs []Notification) error
	GetUserNotifications(ctx context.Context, userAddr string, opts ListOptions) ([]Notification, error)
	MarkAsRead(ctx context.Context, userAddr string, notifIDs []string) error
	UpdatePreferences(ctx context.Context, userAddr string, prefs Preferences) error
	GetPreferences(ctx context.Context, userAddr string) (Preferences, error)
	RegisterDevice(ctx context.Context, device DeviceToken) error
	UnregisterDevice(ctx context.Context, userAddr, token string) error
	ListDevices(ctx context.Context, userAddr string) ([]DeviceToken, error)
}

// Store persists notifications.
type Store interface {
	Add(ctx context.Context, notif Notification) error
	List(ctx context.Context, userAddr string, opts ListOptions) ([]Notification, error)
	MarkRead(ctx context.Context, userAddr string, ids []string) error
}

// PreferencesStore persists notification preferences.
type PreferencesStore interface {
	Get(ctx context.Context, userAddr string) (Preferences, error)
	Put(ctx context.Context, prefs Preferences) error
}

// DeviceTokenStore persists encrypted device tokens.
type DeviceTokenStore interface {
	Register(ctx context.Context, device DeviceToken) error
	Unregister(ctx context.Context, userAddr, token string) error
	List(ctx context.Context, userAddr string) ([]DeviceToken, error)
}

// PushClient delivers push notifications.
type PushClient interface {
	SendToDevice(ctx context.Context, device DeviceToken, notif Notification) error
	SendToTopic(ctx context.Context, topic string, notif Notification) error
	SendSilent(ctx context.Context, device DeviceToken, notif Notification) error
}

// EmailSender delivers email notifications.
type EmailSender interface {
	Send(ctx context.Context, notif Notification, recipient string) error
}

// InAppPublisher streams in-app notifications.
type InAppPublisher interface {
	Publish(ctx context.Context, notif Notification) error
}
