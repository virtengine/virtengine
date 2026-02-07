package notifications

import (
	"context"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time { return f.now }

type fakePush struct {
	deviceCalls int
	topicCalls  int
}

func (f *fakePush) SendToDevice(ctx context.Context, token string, notification Notification) error {
	_ = ctx
	_ = token
	_ = notification
	f.deviceCalls++
	return nil
}

func (f *fakePush) SendToTopic(ctx context.Context, topic string, notification Notification) error {
	_ = ctx
	_ = topic
	_ = notification
	f.topicCalls++
	return nil
}

type fakeEmail struct {
	calls int
}

func (f *fakeEmail) Send(ctx context.Context, notification Notification) error {
	_ = ctx
	_ = notification
	f.calls++
	return nil
}

type fakeInApp struct {
	calls int
}

func (f *fakeInApp) Publish(ctx context.Context, notification Notification) error {
	_ = ctx
	_ = notification
	f.calls++
	return nil
}

func TestQuietHoursSuppressesPushAndEmail(t *testing.T) {
	store := NewMemoryStore(NoopCipher{})
	push := &fakePush{}
	email := &fakeEmail{}
	inApp := &fakeInApp{}

	service := NewService(store, store, store, push, email, inApp)
	service.now = fakeClock{now: time.Date(2026, time.February, 7, 2, 0, 0, 0, time.UTC)}

	prefs := DefaultPreferences("user")
	prefs.QuietHours = &QuietHours{Enabled: true, StartHour: 0, EndHour: 6, Timezone: "UTC"}
	if err := store.UpdatePreferences(context.Background(), "user", prefs); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	_, _ = store.RegisterDevice(context.Background(), "user", DeviceRegistration{
		Token:    "token",
		Platform: DevicePlatformIOS,
	})

	err := service.Send(context.Background(), Notification{
		UserAddress: "user",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order Update",
		Body:        "Order ready",
		Channels:    []Channel{ChannelPush, ChannelEmail, ChannelInApp},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if push.deviceCalls != 0 {
		t.Fatalf("expected push suppressed in quiet hours")
	}
	if email.calls != 0 {
		t.Fatalf("expected email suppressed in quiet hours")
	}
	if inApp.calls != 1 {
		t.Fatalf("expected in-app delivery")
	}
}

func TestAllowDuringQuietHours(t *testing.T) {
	store := NewMemoryStore(NoopCipher{})
	push := &fakePush{}
	email := &fakeEmail{}

	service := NewService(store, store, store, push, email, nil)
	service.now = fakeClock{now: time.Date(2026, time.February, 7, 2, 0, 0, 0, time.UTC)}

	prefs := DefaultPreferences("user")
	prefs.QuietHours = &QuietHours{Enabled: true, StartHour: 0, EndHour: 6, Timezone: "UTC"}
	if err := store.UpdatePreferences(context.Background(), "user", prefs); err != nil {
		t.Fatalf("update prefs: %v", err)
	}

	_, _ = store.RegisterDevice(context.Background(), "user", DeviceRegistration{
		Token:    "token",
		Platform: DevicePlatformIOS,
	})

	err := service.Send(context.Background(), Notification{
		UserAddress:           "user",
		Type:                  NotificationTypeSecurityAlert,
		Title:                 "Alert",
		Body:                  "MFA change",
		Channels:              []Channel{ChannelPush, ChannelEmail},
		AllowDuringQuietHours: true,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if push.deviceCalls != 1 {
		t.Fatalf("expected push delivered")
	}
	if email.calls != 1 {
		t.Fatalf("expected email delivered")
	}
}
