/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/virtengine/virtengine/pkg/notifications"
)

// Client wraps Firebase Cloud Messaging client.
type Client struct {
	app       *firebase.App
	messaging *messaging.Client
}

// Config for Firebase client initialization.
type Config struct {
	CredentialsFile string
}

// NewClient initializes Firebase app and messaging client.
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if config.CredentialsFile == "" {
		return nil, fmt.Errorf("firebase credentials file required")
	}
	opt := option.WithCredentialsFile(config.CredentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app init: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging: %w", err)
	}

	return &Client{app: app, messaging: client}, nil
}

// SendToDevice sends a notification to a device token.
func (c *Client) SendToDevice(ctx context.Context, device notifications.DeviceToken, notif notifications.Notification) error {
	message := buildMessage(&notif)
	message.Token = device.Token
	_, err := c.messaging.Send(ctx, message)
	return err
}

// SendToTopic sends a notification to a topic.
func (c *Client) SendToTopic(ctx context.Context, topic string, notif notifications.Notification) error {
	message := buildMessage(&notif)
	message.Topic = topic
	_, err := c.messaging.Send(ctx, message)
	return err
}

// SendSilent sends a silent (data-only) notification.
func (c *Client) SendSilent(ctx context.Context, device notifications.DeviceToken, notif notifications.Notification) error {
	message := buildMessage(&notif)
	message.Token = device.Token
	message.Notification = nil
	if message.Android != nil {
		message.Android.Priority = "normal"
	}
	if message.APNS != nil {
		message.APNS.Payload.Aps.ContentAvailable = true
		message.APNS.Payload.Aps.Sound = ""
	}
	_, err := c.messaging.Send(ctx, message)
	return err
}

func buildMessage(notif *notifications.Notification) *messaging.Message {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: notif.Title,
			Body:  notif.Body,
		},
		Data: notif.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: new(int),
				},
			},
		},
	}

	return message
}
