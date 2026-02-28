package notifications

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type mockPush struct {
	deviceCount      int
	topicCount       int
	silentCount      int
	deviceErrByToken map[string]error
	silentErrByToken map[string]error
	topicErr         error
}

func (m *mockPush) SendToDevice(_ context.Context, device DeviceToken, _ Notification) error {
	m.deviceCount++
	if m.deviceErrByToken != nil {
		if err := m.deviceErrByToken[device.Token]; err != nil {
			return err
		}
	}
	return nil
}

func (m *mockPush) SendToTopic(_ context.Context, _ string, _ Notification) error {
	m.topicCount++
	return m.topicErr
}

func (m *mockPush) SendSilent(_ context.Context, device DeviceToken, _ Notification) error {
	m.silentCount++
	if m.silentErrByToken != nil {
		if err := m.silentErrByToken[device.Token]; err != nil {
			return err
		}
	}
	return nil
}

type mockEmail struct {
	count int
	err   error
}

func (m *mockEmail) Send(_ context.Context, _ Notification, _ string) error {
	m.count++
	return m.err
}

type mockInApp struct {
	count int
	err   error
}

func (m *mockInApp) Publish(_ context.Context, _ Notification) error {
	m.count++
	return m.err
}

func TestQuietHoursFilteringDoesNotMutatePreferences(t *testing.T) {
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
		Topic:       "orders",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if push.topicCount != 0 || email.count != 0 {
		t.Fatalf("expected quiet hours to suppress push/email, got push=%d email=%d", push.topicCount, email.count)
	}
	if inapp.count == 0 {
		t.Fatalf("expected in-app publish")
	}

	service.timeNow = func() time.Time {
		return time.Date(2025, 2, 8, 12, 0, 0, 0, time.UTC)
	}
	err = service.Send(context.Background(), Notification{
		UserAddress: "user1",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        "Order status changed",
		Topic:       "orders",
	})
	if err != nil {
		t.Fatalf("second send: %v", err)
	}

	if push.topicCount != 1 || email.count != 1 {
		t.Fatalf("expected stored preferences to remain intact outside quiet hours, got push=%d email=%d", push.topicCount, email.count)
	}
	stored, err := service.GetPreferences(context.Background(), "user1")
	if err != nil {
		t.Fatalf("get prefs: %v", err)
	}
	if len(stored.Channels[NotificationTypeOrderUpdate]) != 3 {
		t.Fatalf("expected order-update channels to remain intact, got %v", stored.Channels[NotificationTypeOrderUpdate])
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

func TestSendStoresAndReturnsDeliveryErrorWithoutDuplicatingRetry(t *testing.T) {
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	email := &mockEmail{err: errors.New("smtp timeout")}
	inapp := &mockInApp{}

	service := NewDefaultService(store, prefs, nil, nil, email, inapp)

	notif := Notification{
		ID:          "notif-retry-1",
		UserAddress: "user4",
		Type:        NotificationTypeVEIDStatus,
		Title:       "VEID update",
		Body:        "Approved",
		Channels:    []Channel{ChannelEmail, ChannelInApp},
	}
	err := service.Send(context.Background(), notif)
	if err == nil {
		t.Fatalf("expected delivery error")
	}
	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) {
		t.Fatalf("expected DeliveryError, got %T", err)
	}
	if len(deliveryErr.Failures) != 1 || deliveryErr.Failures[0].Channel != ChannelEmail {
		t.Fatalf("expected email failure, got %+v", deliveryErr.Failures)
	}

	stored, err := store.List(context.Background(), "user4", ListOptions{})
	if err != nil {
		t.Fatalf("list after first send: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected one stored notification after first send, got %d", len(stored))
	}
	if inapp.count != 1 {
		t.Fatalf("expected in-app publish despite email failure, got %d", inapp.count)
	}

	email.err = nil
	err = service.Send(context.Background(), notif)
	if err != nil {
		t.Fatalf("retry send: %v", err)
	}

	stored, err = store.List(context.Background(), "user4", ListOptions{})
	if err != nil {
		t.Fatalf("list after retry: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected idempotent retry to keep one record, got %d", len(stored))
	}
}

func TestSendPushDisablesInvalidTokenAndContinues(t *testing.T) {
	vault := newTestVault(t)
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	devices := NewInMemoryDeviceTokenStore(vault)
	push := &mockPush{
		deviceErrByToken: map[string]error{
			"bad-token": errors.New("registration-token-not-registered"),
		},
	}
	service := NewDefaultService(store, prefs, devices, push, nil, nil)

	for _, device := range []DeviceToken{
		{UserAddress: "user5", Token: "bad-token", Platform: PlatformIOS, AppID: "app"},
		{UserAddress: "user5", Token: "good-token", Platform: PlatformAndroid, AppID: "app"},
	} {
		if err := devices.Register(context.Background(), device); err != nil {
			t.Fatalf("register device: %v", err)
		}
	}

	err := service.Send(context.Background(), Notification{
		ID:          "notif-push-1",
		UserAddress: "user5",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        "Provisioned",
		Channels:    []Channel{ChannelPush},
	})
	if err != nil {
		t.Fatalf("send should succeed when at least one device receives the push: %v", err)
	}

	listed, err := devices.List(context.Background(), "user5")
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected two devices, got %d", len(listed))
	}

	var bad, good *DeviceToken
	for i := range listed {
		switch listed[i].Token {
		case "bad-token":
			bad = &listed[i]
		case "good-token":
			good = &listed[i]
		}
	}
	if bad == nil || good == nil {
		t.Fatalf("expected both devices in listing: %+v", listed)
	}
	if bad.Enabled || bad.DisabledAt == nil || !strings.Contains(strings.ToLower(bad.LastFailureReason), "registration-token-not-registered") {
		t.Fatalf("expected invalid token to be disabled with failure reason, got %+v", *bad)
	}
	if bad.ConsecutiveFailures != 1 {
		t.Fatalf("expected one consecutive failure on bad token, got %d", bad.ConsecutiveFailures)
	}
	if good.LastDeliveredAt == nil {
		t.Fatalf("expected successful device delivery metadata, got %+v", *good)
	}
}

func TestGetPreferencesReturnsSafeCopyAndSecurityDefaults(t *testing.T) {
	store := NewInMemoryPreferencesStore()
	input := Preferences{
		UserAddress: "user6",
		Channels: map[NotificationType][]Channel{
			NotificationTypeSecurityAlert: {ChannelPush},
		},
		Frequencies: map[NotificationType]DeliveryFrequency{
			NotificationTypeSecurityAlert: FrequencyDigest,
		},
	}
	if err := store.Put(context.Background(), input); err != nil {
		t.Fatalf("put prefs: %v", err)
	}

	input.Channels[NotificationTypeSecurityAlert][0] = ChannelEmail
	got, err := store.Get(context.Background(), "user6")
	if err != nil {
		t.Fatalf("get prefs: %v", err)
	}
	if got.Frequencies[NotificationTypeSecurityAlert] != FrequencyImmediate {
		t.Fatalf("expected security alerts to remain immediate, got %s", got.Frequencies[NotificationTypeSecurityAlert])
	}
	if !hasChannel(got.Channels[NotificationTypeSecurityAlert], ChannelPush) ||
		!hasChannel(got.Channels[NotificationTypeSecurityAlert], ChannelEmail) ||
		!hasChannel(got.Channels[NotificationTypeSecurityAlert], ChannelInApp) {
		t.Fatalf("expected security alerts to keep required channels, got %v", got.Channels[NotificationTypeSecurityAlert])
	}

	got.Channels[NotificationTypeSecurityAlert][0] = ChannelInApp
	again, err := store.Get(context.Background(), "user6")
	if err != nil {
		t.Fatalf("get prefs second time: %v", err)
	}
	if !hasChannel(again.Channels[NotificationTypeSecurityAlert], ChannelEmail) {
		t.Fatalf("expected stored prefs to be isolated from returned mutations, got %v", again.Channels[NotificationTypeSecurityAlert])
	}
}

func hasChannel(channels []Channel, target Channel) bool {
	for _, channel := range channels {
		if channel == target {
			return true
		}
	}
	return false
}
