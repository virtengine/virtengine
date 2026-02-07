package apns

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/virtengine/virtengine/pkg/notifications"
)

// Config defines APNs configuration.
type Config struct {
	// Topic is the APNs topic (bundle ID).
	Topic string
	// UseSandbox indicates APNs sandbox endpoint.
	UseSandbox bool

	// Token-based auth.
	KeyID   string
	TeamID  string
	AuthKey []byte

	// Certificate-based auth.
	Certificate tls.Certificate
}

// Client sends notifications via APNs.
type Client struct {
	client *apns2.Client
	topic  string
}

// NewClient creates a new APNs client.
func NewClient(config Config) (*Client, error) {
	if config.Topic == "" {
		return nil, fmt.Errorf("apns topic is required")
	}

	var client *apns2.Client
	if len(config.AuthKey) > 0 {
		authKey, err := token.AuthKeyFromBytes(config.AuthKey)
		if err != nil {
			return nil, fmt.Errorf("apns auth key: %w", err)
		}
		jwtToken := &token.Token{
			AuthKey: authKey,
			KeyID:   config.KeyID,
			TeamID:  config.TeamID,
		}
		client = apns2.NewTokenClient(jwtToken)
	} else if len(config.Certificate.Certificate) > 0 {
		client = apns2.NewClient(config.Certificate)
	} else {
		return nil, fmt.Errorf("apns auth not configured")
	}

	if config.UseSandbox {
		client = client.Development()
	} else {
		client = client.Production()
	}

	return &Client{client: client, topic: config.Topic}, nil
}

// SendToDevice sends to a device token.
func (c *Client) SendToDevice(ctx context.Context, token string, notif notifications.Notification) error {
	n := &apns2.Notification{
		DeviceToken: token,
		Topic:       c.topic,
		Payload:     buildPayload(notif),
	}
	res, err := c.client.PushWithContext(ctx, n)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("apns delivery failed: %s", res.Reason)
	}
	return nil
}

// SendToTopic sends a broadcast by overriding the topic if provided.
func (c *Client) SendToTopic(ctx context.Context, topic string, notif notifications.Notification) error {
	_ = ctx
	_ = notif
	return fmt.Errorf("apns topic broadcasts are not supported (topic=%s)", topic)
}

func buildPayload(notif notifications.Notification) *payload.Payload {
	builder := payload.NewPayload()
	if notif.Silent {
		builder.ContentAvailable()
	} else {
		builder.AlertTitle(notif.Title)
		builder.AlertBody(notif.Body)
	}
	if notif.UnreadCount > 0 {
		builder.Badge(notif.UnreadCount)
	}
	for key, value := range notif.Data {
		builder.Custom(key, value)
	}
	return builder
}
