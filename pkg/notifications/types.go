package notifications

import "time"

// NotificationType defines the category for a notification.
type NotificationType string

const (
	NotificationTypeVEIDStatus    NotificationType = "veid_status"
	NotificationTypeOrderUpdate   NotificationType = "order_update"
	NotificationTypeEscrowDeposit NotificationType = "escrow_deposit"
	NotificationTypeSecurityAlert NotificationType = "security_alert"
	NotificationTypeProviderAlert NotificationType = "provider_alert"
)

// Channel defines delivery channels.
type Channel string

const (
	ChannelPush  Channel = "push"
	ChannelEmail Channel = "email"
	ChannelInApp Channel = "in_app"
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
	UnreadCount int
	Silent      bool
	Topics      []string

	AllowDuringQuietHours bool
}

// ListOptions controls listing notifications.
type ListOptions struct {
	Limit       int
	Offset      int
	IncludeRead bool
	Types       []NotificationType
	Since       *time.Time
	Until       *time.Time
}
