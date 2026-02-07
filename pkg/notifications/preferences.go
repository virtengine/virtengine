package notifications

import (
	"context"
	"fmt"
	"time"
)

// Preferences stores per-user notification preferences.
type Preferences struct {
	UserAddress   string
	Channels      map[NotificationType][]Channel
	QuietHours    *QuietHours
	DigestEnabled bool
	DigestTime    string
}

// QuietHours controls quiet hours for notifications.
type QuietHours struct {
	Enabled   bool
	StartHour int
	EndHour   int
	Timezone  string
}

// PreferencesStore persists notification preferences.
type PreferencesStore interface {
	UpdatePreferences(ctx context.Context, userAddress string, prefs Preferences) error
	GetPreferences(ctx context.Context, userAddress string) (Preferences, error)
}

// DefaultPreferences returns a conservative default preference set.
func DefaultPreferences(userAddress string) Preferences {
	return Preferences{
		UserAddress: userAddress,
		Channels: map[NotificationType][]Channel{
			NotificationTypeVEIDStatus:    {ChannelInApp, ChannelEmail},
			NotificationTypeOrderUpdate:   {ChannelInApp, ChannelPush, ChannelEmail},
			NotificationTypeEscrowDeposit: {ChannelInApp, ChannelEmail},
			NotificationTypeSecurityAlert: {ChannelInApp, ChannelPush, ChannelEmail},
			NotificationTypeProviderAlert: {ChannelInApp, ChannelPush},
		},
		DigestEnabled: false,
		DigestTime:    "09:00",
	}
}

// Validate validates preference values.
func (p Preferences) Validate() error {
	if p.UserAddress == "" {
		return fmt.Errorf("user address is required")
	}
	if p.DigestTime != "" {
		if _, err := time.Parse("15:04", p.DigestTime); err != nil {
			return fmt.Errorf("invalid digest time: %w", err)
		}
	}
	if p.QuietHours != nil {
		if p.QuietHours.StartHour < 0 || p.QuietHours.StartHour > 23 {
			return fmt.Errorf("quiet hours start hour out of range")
		}
		if p.QuietHours.EndHour < 0 || p.QuietHours.EndHour > 23 {
			return fmt.Errorf("quiet hours end hour out of range")
		}
	}
	return nil
}

func (p Preferences) channelsFor(notificationType NotificationType) []Channel {
	if p.Channels == nil {
		return nil
	}
	return p.Channels[notificationType]
}
