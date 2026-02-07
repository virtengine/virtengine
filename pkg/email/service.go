/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package email

import (
	"context"
	"fmt"
	"net/url"
)

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
	rendered, err := s.Render(name, data)
	if err != nil {
		return err
	}
	message := EmailMessage{
		To:      to,
		From:    s.defaultFrom,
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
		Headers: map[string]string{
			"List-Unsubscribe": fmt.Sprintf("<%s>", s.BuildUnsubscribeURL(unsubscribeToken)),
		},
	}
	if s.sender == nil {
		return nil
	}
	return s.sender.Send(ctx, message)
}
