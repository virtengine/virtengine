/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package email

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

var ErrSenderNotConfigured = errors.New("email sender not configured")
var ErrRecipientRequired = errors.New("email recipient required")
var ErrFromAddressRequired = errors.New("email sender address required")

// Sender delivers email messages.
type Sender interface {
	Send(ctx context.Context, message EmailMessage) error
}

// Service sends notifications with templated email.
type Service struct {
	renderer        Renderer
	sender          Sender
	defaultFrom     string
	unsubscribeBase string
}

// NewService creates a new email service.
func NewService(renderer Renderer, sender Sender, defaultFrom string, unsubscribeBase string) *Service {
	return &Service{
		renderer:        renderer,
		sender:          sender,
		defaultFrom:     defaultFrom,
		unsubscribeBase: unsubscribeBase,
	}
}

// BuildUnsubscribeURL builds a one-click unsubscribe link.
func (s *Service) BuildUnsubscribeURL(token string) string {
	if s.unsubscribeBase == "" {
		return ""
	}
	u, err := url.Parse(s.unsubscribeBase)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}

// Render renders a template and returns a message payload.
func (s *Service) Render(name TemplateName, data any) (RenderedEmail, error) {
	if s.renderer == nil {
		return RenderedEmail{}, fmt.Errorf("renderer not configured")
	}
	return s.renderer.Render(name, data)
}

// SendTemplate renders and sends a templated email.
func (s *Service) SendTemplate(ctx context.Context, to string, unsubscribeToken string, name TemplateName, data any) error {
	if s.sender == nil {
		return ErrSenderNotConfigured
	}
	recipient, err := validateMailbox(to, ErrRecipientRequired)
	if err != nil {
		return err
	}
	from, err := validateMailbox(s.defaultFrom, ErrFromAddressRequired)
	if err != nil {
		return err
	}
	rendered, err := s.Render(name, data)
	if err != nil {
		return err
	}
	rendered.Subject = sanitizeHeaderValue(rendered.Subject)
	if rendered.Subject == "" {
		return fmt.Errorf("email subject is empty")
	}
	message := EmailMessage{
		To:      recipient,
		From:    from,
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
	}
	if unsubscribeURL := s.BuildUnsubscribeURL(unsubscribeToken); unsubscribeURL != "" {
		message.Headers = map[string]string{
			"List-Unsubscribe": fmt.Sprintf("<%s>", unsubscribeURL),
		}
	}
	return s.sender.Send(ctx, message)
}

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.Join(strings.Fields(value), " ")
}

func validateMailbox(value string, emptyErr error) (string, error) {
	value = sanitizeHeaderValue(value)
	if value == "" {
		return "", emptyErr
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return "", fmt.Errorf("invalid email address %q: %w", value, err)
	}
	return value, nil
}
