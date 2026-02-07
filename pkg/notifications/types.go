/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import "time"

// NotificationType defines supported notification categories.
type NotificationType string

const (
	NotificationTypeVEIDStatus    NotificationType = "veid_status"
	NotificationTypeOrderUpdate   NotificationType = "order_update"
	NotificationTypeEscrowDeposit NotificationType = "escrow_deposit"
	NotificationTypeSecurityAlert NotificationType = "security_alert"
	NotificationTypeProviderAlert NotificationType = "provider_alert"
	NotificationTypeSLABreach     NotificationType = "sla_breach"
	NotificationTypeSLAWarning    NotificationType = "sla_warning"
)

// Channel defines delivery channels.
type Channel string

const (
	ChannelPush  Channel = "push"
	ChannelEmail Channel = "email"
	ChannelInApp Channel = "in_app"
)

// DeliveryFrequency controls immediate vs digest delivery.
type DeliveryFrequency string

const (
	FrequencyImmediate DeliveryFrequency = "immediate"
	FrequencyDigest    DeliveryFrequency = "digest"
)

// Notification represents a user-facing notification.
type Notification struct {
	ID          string
	UserAddress string
	Type        NotificationType
	Title       string
	Body        string
	Data        map[string]string
	CreatedAt   time.Time
	ReadAt      *time.Time
	Channels    []Channel
	Topic       string
	Silent      bool
}

// ListOptions configures notification list queries.
type ListOptions struct {
	Limit      int
	Cursor     string
	UnreadOnly bool
}

// Preferences stores user notification preferences.
type Preferences struct {
	UserAddress   string
	Channels      map[NotificationType][]Channel
	Frequencies   map[NotificationType]DeliveryFrequency
	QuietHours    *QuietHours
	DigestEnabled bool
	DigestTime    string // "09:00" UTC by default
}

// QuietHours defines a user quiet hours window.
type QuietHours struct {
	Enabled   bool
	StartHour int // 0-23
	EndHour   int // 0-23
	Timezone  string
}

// DevicePlatform indicates device platform for push delivery.
type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
)

// DeviceToken is a registered device for push notifications.
type DeviceToken struct {
	ID          string
	UserAddress string
	Platform    DevicePlatform
	Token       string
	AppID       string
	Topic       string
	CreatedAt   time.Time
	LastSeenAt  time.Time
	Enabled     bool
}
