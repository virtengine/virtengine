/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"fmt"
)

// PushRouter routes push notifications to platform-specific clients.
type PushRouter struct {
	android PushClient
	ios     PushClient
}

// NewPushRouter creates a router for Android (FCM) and iOS (APNs) clients.
func NewPushRouter(android PushClient, ios PushClient) *PushRouter {
	return &PushRouter{android: android, ios: ios}
}

// SendToDevice routes by device platform.
func (r *PushRouter) SendToDevice(ctx context.Context, device DeviceToken, notif Notification) error {
	client := r.clientFor(device.Platform)
	if client == nil {
		return fmt.Errorf("no push client for platform %s", device.Platform)
	}
	return client.SendToDevice(ctx, device, notif)
}

// SendToTopic routes topic broadcasts to Android/FCM client.
func (r *PushRouter) SendToTopic(ctx context.Context, topic string, notif Notification) error {
	if r.android == nil {
		return fmt.Errorf("no push client for topic %s", topic)
	}
	return r.android.SendToTopic(ctx, topic, notif)
}

// SendSilent routes silent notifications.
func (r *PushRouter) SendSilent(ctx context.Context, device DeviceToken, notif Notification) error {
	client := r.clientFor(device.Platform)
	if client == nil {
		return fmt.Errorf("no push client for platform %s", device.Platform)
	}
	return client.SendSilent(ctx, device, notif)
}

func (r *PushRouter) clientFor(platform DevicePlatform) PushClient {
	switch platform {
	case PlatformAndroid:
		return r.android
	case PlatformIOS:
		return r.ios
	default:
		return nil
	}
}
