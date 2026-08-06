//go:build e2e.integration

package notifications

import (
	"context"
	"testing"
	"time"
)

func TestNotificationServiceE2EReadLifecycle(t *testing.T) {
	vault := newTestVault(t)
	store := NewInMemoryStore()
	prefs := NewInMemoryPreferencesStore()
	devices := NewInMemoryDeviceTokenStore(vault)
	hub := NewInAppHub()
	push := &mockPush{}
	email := &mockEmail{}
	service := NewDefaultService(store, prefs, devices, push, email, hub)
	service.timeNow = func() time.Time {
		return time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)
	}

	if err := devices.Register(context.Background(), DeviceToken{
		UserAddress: "e2e-user",
		Token:       "e2e-token",
		Platform:    PlatformIOS,
		AppID:       "com.virtengine.portal",
	}); err != nil {
		t.Fatalf("register device: %v", err)
	}
	if err := prefs.Put(context.Background(), Preferences{
		UserAddress: "e2e-user",
		Channels: map[NotificationType][]Channel{
			NotificationTypeVEIDStatus: {ChannelPush, ChannelEmail, ChannelInApp},
		},
	}); err != nil {
		t.Fatalf("put prefs: %v", err)
	}

	stream, cancel := hub.Subscribe("e2e-user", 1)
	defer cancel()

	notif := Notification{
		ID:          "notif-e2e-1",
		UserAddress: "e2e-user",
		Type:        NotificationTypeVEIDStatus,
		Title:       "VEID verification update",
		Body:        "Approved",
	}
	if err := service.Send(context.Background(), notif); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case received := <-stream:
		if received.ID != notif.ID {
			t.Fatalf("expected stream id %q, got %q", notif.ID, received.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected in-app stream delivery")
	}

	unread, err := service.GetUserNotifications(context.Background(), "e2e-user", ListOptions{UnreadOnly: true})
	if err != nil {
		t.Fatalf("list unread: %v", err)
	}
	if len(unread) != 1 {
		t.Fatalf("expected one unread notification, got %d", len(unread))
	}

	if err := service.MarkAsRead(context.Background(), "e2e-user", []string{notif.ID}); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	unread, err = service.GetUserNotifications(context.Background(), "e2e-user", ListOptions{UnreadOnly: true})
	if err != nil {
		t.Fatalf("list unread after mark read: %v", err)
	}
	if len(unread) != 0 {
		t.Fatalf("expected zero unread notifications after mark read, got %d", len(unread))
	}

	all, err := service.GetUserNotifications(context.Background(), "e2e-user", ListOptions{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 1 || all[0].ReadAt == nil {
		t.Fatalf("expected stored read notification, got %+v", all)
	}
}
