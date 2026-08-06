//go:build integration

package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNotificationServiceIntegrationDurableRetryFlow(t *testing.T) {
	vault := newTestVault(t)
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	devices := NewInMemoryDeviceTokenStore(vault)
	hub := NewInAppHub()
	push := &mockPush{
		deviceErrByToken: map[string]error{
			"bad-token": errors.New("registration-token-not-registered"),
		},
	}
	email := &mockEmail{err: errors.New("smtp temporary failure")}
	service := NewDefaultService(store, prefs, devices, push, email, hub)
	service.timeNow = func() time.Time {
		return time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC)
	}

	for _, device := range []DeviceToken{
		{UserAddress: "integration-user", Token: "bad-token", Platform: PlatformIOS, AppID: "com.virtengine.portal"},
		{UserAddress: "integration-user", Token: "good-token", Platform: PlatformAndroid, AppID: "com.virtengine.portal"},
	} {
		if err := devices.Register(context.Background(), device); err != nil {
			t.Fatalf("register device: %v", err)
		}
	}
	if err := prefs.Put(context.Background(), Preferences{
		UserAddress: "integration-user",
		Channels: map[NotificationType][]Channel{
			NotificationTypeOrderUpdate: {ChannelPush, ChannelEmail, ChannelInApp},
		},
	}); err != nil {
		t.Fatalf("put prefs: %v", err)
	}

	stream, cancel := hub.Subscribe("integration-user", 2)
	defer cancel()

	notif := Notification{
		ID:          "notif-integration-1",
		UserAddress: "integration-user",
		Type:        NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        "Provisioned",
	}
	err := service.Send(context.Background(), notif)
	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) {
		t.Fatalf("expected delivery error from email failure, got %v", err)
	}
	if len(deliveryErr.Failures) != 1 || deliveryErr.Failures[0].Channel != ChannelEmail {
		t.Fatalf("expected only email to fail because one push device still succeeds, got %+v", deliveryErr.Failures)
	}

	select {
	case received := <-stream:
		if received.ID != notif.ID {
			t.Fatalf("expected streamed notification id %q, got %q", notif.ID, received.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected in-app stream delivery")
	}

	listed, err := service.GetUserNotifications(context.Background(), "integration-user", ListOptions{})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected exactly one stored notification, got %d", len(listed))
	}

	devicesAfter, err := devices.List(context.Background(), "integration-user")
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	var disabledCount, deliveredCount int
	for _, device := range devicesAfter {
		if !device.Enabled && device.DisabledAt != nil {
			disabledCount++
		}
		if device.LastDeliveredAt != nil {
			deliveredCount++
		}
	}
	if disabledCount != 1 || deliveredCount != 1 {
		t.Fatalf("expected one disabled token and one delivered token, got disabled=%d delivered=%d", disabledCount, deliveredCount)
	}

	email.err = nil
	if err := service.Send(context.Background(), notif); err != nil {
		t.Fatalf("retry send: %v", err)
	}
	listed, err = service.GetUserNotifications(context.Background(), "integration-user", ListOptions{})
	if err != nil {
		t.Fatalf("list notifications after retry: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected retry to stay idempotent, got %d notifications", len(listed))
	}
}
