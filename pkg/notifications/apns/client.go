/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package apns

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"

	"github.com/virtengine/virtengine/pkg/notifications"
)

// Config configures the APNs client.
type Config struct {
	TeamID     string
	KeyID      string
	KeyPath    string
	Topic      string
	CertPath   string
	CertKey    string
	UseSandbox bool
}

// Client wraps APNs client.
type Client struct {
	client *apns2.Client
	topic  string
}

// NewClient initializes APNs client.
func NewClient(config Config) (*Client, error) {
	if config.Topic == "" {
		return nil, fmt.Errorf("apns topic required")
	}

	var client *apns2.Client
	if config.KeyPath != "" {
		authKey, err := token.AuthKeyFromFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("load apns key: %w", err)
		}
		authToken := &token.Token{
			AuthKey: authKey,
			KeyID:   config.KeyID,
			TeamID:  config.TeamID,
		}
		client = apns2.NewTokenClient(authToken)
	} else if config.CertPath != "" && config.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(config.CertPath, config.CertKey)
		if err != nil {
			return nil, fmt.Errorf("load apns certificate: %w", err)
		}
		client = apns2.NewClient(cert)
	} else {
		return nil, fmt.Errorf("apns auth key or certificate required")
	}

	if config.UseSandbox {
		client = client.Development()
	}

	return &Client{client: client, topic: config.Topic}, nil
}

// SendToDevice sends a notification to a device token.
func (c *Client) SendToDevice(ctx context.Context, device notifications.DeviceToken, notif notifications.Notification) error {
	notification := buildNotification(device.Token, c.topic, notif, false)
	resp, err := c.client.PushWithContext(ctx, notification)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("apns status: %d reason: %s", resp.StatusCode, resp.Reason)
	}
	return nil
}

// SendToTopic is not supported for APNs.
func (c *Client) SendToTopic(_ context.Context, _ string, _ notifications.Notification) error {
	return fmt.Errorf("apns does not support topic broadcast")
}

// SendSilent sends a silent notification to a device.
func (c *Client) SendSilent(ctx context.Context, device notifications.DeviceToken, notif notifications.Notification) error {
	notification := buildNotification(device.Token, c.topic, notif, true)
	resp, err := c.client.PushWithContext(ctx, notification)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("apns status: %d reason: %s", resp.StatusCode, resp.Reason)
	}
	return nil
}

func buildNotification(token string, topic string, notif notifications.Notification, silent bool) *apns2.Notification {
	apnsPayload := payload.NewPayload().
		AlertTitle(notif.Title).
		AlertBody(notif.Body).
		Sound("default").
		Badge(1).
		Custom("type", string(notif.Type))

	for key, value := range notif.Data {
		apnsPayload.Custom(key, value)
	}

	if silent {
		apnsPayload = payload.NewPayload().ContentAvailable()
	}

	pushType := apns2.PushTypeAlert
	if silent {
		pushType = apns2.PushTypeBackground
	}

	return &apns2.Notification{
		DeviceToken: token,
		Topic:       topic,
		Payload:     apnsPayload,
		PushType:    pushType,
		Expiration:  time.Now().Add(24 * time.Hour),
	}
}
