package email

import "context"

// Message represents an email payload.
type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Sender sends email messages.
type Sender interface {
	Send(ctx context.Context, message Message) error
}

// NoopSender performs no operation.
type NoopSender struct{}

func (NoopSender) Send(ctx context.Context, message Message) error {
	return nil
}
