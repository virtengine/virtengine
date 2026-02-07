package notifications

import (
	"context"
	"testing"
	"time"
)

type mockPush struct {
	deviceCount int
	topicCount  int
	silentCount int
}

func (m *mockPush) SendToDevice(_ context.Context, _ DeviceToken, _ Notification) error {
	m.deviceCount++
	return nil
}

func (m *mockPush) SendToTopic(_ context.Context, _ string, _ Notification) error {
	m.topicCount++
	return nil
}

func (m *mockPush) SendSilent(_ context.Context, _ DeviceToken, _ Notification) error {
	m.silentCount++
	return nil
}

type mockEmail struct {
	count int
}

func (m *mockEmail) Send(_ context.Context, _ Notification, _ string) error {
	m.count++
	return nil
}

type mockInApp struct {
	count int
}

func (m *mockInApp) Publish(_ context.Context, _ Notification) error {
	m.count++
	return nil
}

func TestQuietHoursFiltering(t *testing.T) {
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	push := &mockPush{}
	email := &mockEmail{}
	inapp := &mockInApp{}

	service := NewDefaultService(store, prefs, nil, push, email, inapp)
	service.timeNow = func() time.Time {
		return time.Date(2025, 2, 7, 23, 0, 0, 0, time.UTC)
	}

	err := prefs.Put(context.Background(), Preferences{
		UserAddress: "user1",
		Channels: map[NotificationType][]Channel{
			NotificationTypeOrderUpdate: {ChannelPush, ChannelEmail, ChannelInApp},
		},
		QuietHours: &QuietHours{
			Enabled:   true,
			StartHour: 22,
			EndHour:   6,
			Timezone:  "UTC",
		},
	})
	if err != nil {
		t.Fatalf("prefs put: %v", err)
	}

	err = service.Send(context.Background(), Notification{
		UserAddress: "user1",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        "Order status changed",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if push.deviceCount != 0 || email.count != 0 {
		t.Fatalf("expected quiet hours to suppress push/email, got push=%d email=%d", push.deviceCount, email.count)
	}
	if inapp.count == 0 {
		t.Fatalf("expected in-app publish")
	}
}

func TestPreferenceRouting(t *testing.T) {
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	push := &mockPush{}
	email := &mockEmail{}

	service := NewDefaultService(store, prefs, nil, push, email, nil)

	err := prefs.Put(context.Background(), Preferences{
		UserAddress: "user2",
		Channels: map[NotificationType][]Channel{
			NotificationTypeVEIDStatus: {ChannelEmail},
		},
	})
	if err != nil {
		t.Fatalf("prefs put: %v", err)
	}

	err = service.Send(context.Background(), Notification{
		UserAddress: "user2",
		Type:        NotificationTypeVEIDStatus,
		Title:       "VEID update",
		Body:        "Approved",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if email.count != 1 {
		t.Fatalf("expected email sent, got %d", email.count)
	}
	if push.deviceCount != 0 {
		t.Fatalf("expected no push sent, got %d", push.deviceCount)
	}
}

func TestDigestFrequencySuppressesImmediateDelivery(t *testing.T) {
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	push := &mockPush{}
	email := &mockEmail{}

	service := NewDefaultService(store, prefs, nil, push, email, nil)

	err := prefs.Put(context.Background(), Preferences{
		UserAddress:   "user3",
		DigestEnabled: true,
		Channels: map[NotificationType][]Channel{
			NotificationTypeOrderUpdate: {ChannelEmail, ChannelPush, ChannelInApp},
		},
		Frequencies: map[NotificationType]DeliveryFrequency{
			NotificationTypeOrderUpdate: FrequencyDigest,
		},
	})
	if err != nil {
		t.Fatalf("prefs put: %v", err)
	}

	err = service.Send(context.Background(), Notification{
		UserAddress: "user3",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        "Delayed",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if email.count != 0 || push.deviceCount != 0 {
		t.Fatalf("expected digest to suppress immediate delivery")
	}
}
