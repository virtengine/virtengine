package firebase

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/virtengine/virtengine/pkg/notifications"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Config controls Firebase client configuration.
type Config struct {
	CredentialsFile string
}

// Client delivers notifications via Firebase Cloud Messaging.
type Client struct {
	app       *firebase.App
	messaging *messaging.Client
}

// NewClient creates a Firebase client.
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if config.CredentialsFile == "" {
		return nil, fmt.Errorf("firebase credentials file required")
	}
	credentials, err := os.ReadFile(config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read firebase credentials: %w", err)
	}
	creds, err := google.CredentialsFromJSON(ctx, credentials, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("parse firebase credentials: %w", err)
	}
	opt := option.WithCredentials(creds)
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
func (c *Client) SendToDevice(ctx context.Context, token string, notif notifications.Notification) error {
	message := buildMessage(notif)
	message.Token = token
	_, err := c.messaging.Send(ctx, message)
	return err
}

// SendToTopic sends a notification to a topic.
func (c *Client) SendToTopic(ctx context.Context, topic string, notif notifications.Notification) error {
	message := buildMessage(notif)
	message.Topic = topic
	_, err := c.messaging.Send(ctx, message)
	return err
}

func buildMessage(notif notifications.Notification) *messaging.Message {
	message := &messaging.Message{
		Data: notif.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: &notif.UnreadCount,
				},
			},
		},
	}

	if notif.Title != "" || notif.Body != "" {
		message.Notification = &messaging.Notification{
			Title: notif.Title,
			Body:  notif.Body,
		}
	}

	if notif.Silent {
		message.APNS.Payload.Aps.ContentAvailable = true
		message.Android.Priority = "normal"
	}

	return message
}
