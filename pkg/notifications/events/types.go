package events

import "github.com/virtengine/virtengine/pkg/notifications"

// Event represents a notification trigger event.
type Event struct {
	UserAddress           string
	Type                  notifications.NotificationType
	Title                 string
	Body                  string
	Data                  map[string]string
	Channels              []notifications.Channel
	Topics                []string
	Silent                bool
	AllowDuringQuietHours bool
}
