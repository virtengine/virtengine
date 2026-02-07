package notifications

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PushProvider handles sending push notifications.
type PushProvider interface {
	SendToDevice(ctx context.Context, token string, notification Notification) error
	SendToTopic(ctx context.Context, topic string, notification Notification) error
}

// EmailProvider handles sending email notifications.
type EmailProvider interface {
	Send(ctx context.Context, notification Notification) error
}

// InAppPublisher handles live in-app notification events (e.g. WebSocket).
type InAppPublisher interface {
	Publish(ctx context.Context, notification Notification) error
}

type clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// Service provides notification delivery and storage.
type Service struct {
	store      NotificationStore
	prefs      PreferencesStore
	devices    DeviceStore
	push       PushProvider
	email      EmailProvider
	inApp      InAppPublisher
	now        clock
	quietGuard QuietHoursGuard
}

// QuietHoursGuard checks if a notification should be delayed.
type QuietHoursGuard interface {
	InQuietHours(now time.Time, prefs *QuietHours) bool
}

type defaultQuietHoursGuard struct{}

func (defaultQuietHoursGuard) InQuietHours(now time.Time, prefs *QuietHours) bool {
	if prefs == nil || !prefs.Enabled {
		return false
	}

	location := time.UTC
	if prefs.Timezone != "" {
		if loc, err := time.LoadLocation(prefs.Timezone); err == nil {
			location = loc
		}
	}

	local := now.In(location)
	hour := local.Hour()

	if prefs.StartHour == prefs.EndHour {
		return false
	}

	if prefs.StartHour < prefs.EndHour {
		return hour >= prefs.StartHour && hour < prefs.EndHour
	}

	return hour >= prefs.StartHour || hour < prefs.EndHour
}

// NewService builds a notification service.
func NewService(store NotificationStore, prefs PreferencesStore, devices DeviceStore, push PushProvider, email EmailProvider, inApp InAppPublisher) *Service {
	return &Service{
		store:      store,
		prefs:      prefs,
		devices:    devices,
		push:       push,
		email:      email,
		inApp:      inApp,
		now:        systemClock{},
		quietGuard: defaultQuietHoursGuard{},
	}
}

// Send sends a single notification.
func (s *Service) Send(ctx context.Context, notification Notification) error {
	if s.store == nil || s.prefs == nil {
		return fmt.Errorf("notifications service not configured")
	}
	if notification.UserAddress == "" {
		return fmt.Errorf("user address is required")
	}
	if notification.Type == "" {
		return fmt.Errorf("notification type is required")
	}

	if notification.ID == "" {
		notification.ID = uuid.NewString()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = s.now.Now()
	}

	prefs, err := s.getOrDefaultPreferences(ctx, notification.UserAddress)
	if err != nil {
		return err
	}

	channels := notification.Channels
	if len(channels) == 0 {
		channels = prefs.channelsFor(notification.Type)
	}
	if len(channels) == 0 {
		channels = []Channel{ChannelInApp}
	}

	inQuiet := s.quietGuard.InQuietHours(notification.CreatedAt, prefs.QuietHours)

	if err := s.store.Save(ctx, notification); err != nil {
		return err
	}

	var errs multiError
	for _, channel := range channels {
		if inQuiet && !notification.AllowDuringQuietHours {
			if channel == ChannelPush || channel == ChannelEmail {
				continue
			}
		}

		switch channel {
		case ChannelInApp:
			if s.inApp != nil {
				if err := s.inApp.Publish(ctx, notification); err != nil {
					errs.append(err)
				}
			}
		case ChannelPush:
			if err := s.sendPush(ctx, notification); err != nil {
				errs.append(err)
			}
		case ChannelEmail:
			if err := s.sendEmail(ctx, notification, prefs); err != nil {
				errs.append(err)
			}
		default:
			errs.append(fmt.Errorf("unknown channel %s", channel))
		}
	}

	return errs.errorOrNil()
}

// SendBatch sends multiple notifications, returning the first error if any.
func (s *Service) SendBatch(ctx context.Context, notifications []Notification) error {
	var errs multiError
	for _, notification := range notifications {
		if err := s.Send(ctx, notification); err != nil {
			errs.append(err)
		}
	}
	return errs.errorOrNil()
}

// GetUserNotifications lists notifications for a user.
func (s *Service) GetUserNotifications(ctx context.Context, userAddr string, opts ListOptions) ([]Notification, int, error) {
	if s.store == nil {
		return nil, 0, fmt.Errorf("notifications store not configured")
	}
	return s.store.List(ctx, userAddr, opts)
}

// MarkAsRead marks notifications as read.
func (s *Service) MarkAsRead(ctx context.Context, userAddr string, notifIDs []string) error {
	if s.store == nil {
		return fmt.Errorf("notifications store not configured")
	}
	return s.store.MarkAsRead(ctx, userAddr, notifIDs)
}

// UpdatePreferences updates a user notification preference.
func (s *Service) UpdatePreferences(ctx context.Context, userAddr string, prefs Preferences) error {
	if s.prefs == nil {
		return fmt.Errorf("preferences store not configured")
	}
	prefs.UserAddress = userAddr
	if err := prefs.Validate(); err != nil {
		return err
	}
	return s.prefs.UpdatePreferences(ctx, userAddr, prefs)
}

// GetPreferences returns preferences for a user.
func (s *Service) GetPreferences(ctx context.Context, userAddr string) (Preferences, error) {
	if s.prefs == nil {
		return Preferences{}, fmt.Errorf("preferences store not configured")
	}
	return s.getOrDefaultPreferences(ctx, userAddr)
}

func (s *Service) getOrDefaultPreferences(ctx context.Context, userAddr string) (Preferences, error) {
	prefs, err := s.prefs.GetPreferences(ctx, userAddr)
	if err == nil {
		return prefs, nil
	}
	if !errors.Is(err, ErrPreferencesNotFound) {
		return Preferences{}, err
	}
	return DefaultPreferences(userAddr), nil
}

func (s *Service) sendPush(ctx context.Context, notification Notification) error {
	if s.push == nil {
		return fmt.Errorf("push provider not configured")
	}

	if len(notification.Topics) > 0 {
		for _, topic := range notification.Topics {
			if err := s.push.SendToTopic(ctx, topic, notification); err != nil {
				return err
			}
		}
	}

	if s.devices == nil {
		return nil
	}

	devices, err := s.devices.ListDevices(ctx, notification.UserAddress)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.DisabledAt != nil {
			continue
		}
		if err := s.push.SendToDevice(ctx, device.Token, notification); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) sendEmail(ctx context.Context, notification Notification, prefs Preferences) error {
	if prefs.DigestEnabled {
		return nil
	}
	if s.email == nil {
		return fmt.Errorf("email provider not configured")
	}
	return s.email.Send(ctx, notification)
}

type multiError struct {
	mu   sync.Mutex
	errs []error
}

func (m *multiError) append(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errs = append(m.errs, err)
}

func (m *multiError) errorOrNil() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.errs) == 0 {
		return nil
	}
	if len(m.errs) == 1 {
		return m.errs[0]
	}
	return fmt.Errorf("multiple notification errors: %w", errors.Join(m.errs...))
}
