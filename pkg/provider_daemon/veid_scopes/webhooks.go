package veid_scopes

import (
	"context"
	"fmt"
	"time"
)

// WebhookProcessor defines the interface for processing provider webhooks.
type WebhookProcessor interface {
	ProcessWebhook(ctx context.Context, provider string, payload []byte) error
}

// WebhookConfig controls retry behavior for web-scope webhooks.
type WebhookConfig struct {
	MaxRetries int
	RetryDelay time.Duration
	Async      bool
}

// DefaultWebhookConfig returns default webhook settings.
func DefaultWebhookConfig() WebhookConfig {
	return WebhookConfig{
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
		Async:      true,
	}
}

// WebScopeWebhookHandler processes email/SMS webhooks with retry support.
type WebScopeWebhookHandler struct {
	emailProcessor WebhookProcessor
	smsProcessor   WebhookProcessor
	cfg            WebhookConfig
}

// NewWebScopeWebhookHandler creates a new webhook handler.
func NewWebScopeWebhookHandler(emailProcessor, smsProcessor WebhookProcessor, cfg WebhookConfig) *WebScopeWebhookHandler {
	return &WebScopeWebhookHandler{
		emailProcessor: emailProcessor,
		smsProcessor:   smsProcessor,
		cfg:            cfg,
	}
}

// HandleEmailWebhook processes an email webhook with retries.
func (h *WebScopeWebhookHandler) HandleEmailWebhook(ctx context.Context, provider string, payload []byte) error {
	if h.emailProcessor == nil {
		return fmt.Errorf("email webhook processor not configured")
	}
	return h.process(ctx, h.emailProcessor, provider, payload)
}

// HandleSMSWebhook processes an SMS webhook with retries.
func (h *WebScopeWebhookHandler) HandleSMSWebhook(ctx context.Context, provider string, payload []byte) error {
	if h.smsProcessor == nil {
		return fmt.Errorf("sms webhook processor not configured")
	}
	return h.process(ctx, h.smsProcessor, provider, payload)
}

// HandleEmailWebhookAsync processes an email webhook asynchronously.
func (h *WebScopeWebhookHandler) HandleEmailWebhookAsync(ctx context.Context, provider string, payload []byte) {
	if !h.cfg.Async {
		_ = h.HandleEmailWebhook(ctx, provider, payload)
		return
	}
	go func() {
		_ = h.HandleEmailWebhook(ctx, provider, payload)
	}()
}

// HandleSMSWebhookAsync processes an SMS webhook asynchronously.
func (h *WebScopeWebhookHandler) HandleSMSWebhookAsync(ctx context.Context, provider string, payload []byte) {
	if !h.cfg.Async {
		_ = h.HandleSMSWebhook(ctx, provider, payload)
		return
	}
	go func() {
		_ = h.HandleSMSWebhook(ctx, provider, payload)
	}()
}

func (h *WebScopeWebhookHandler) process(ctx context.Context, processor WebhookProcessor, provider string, payload []byte) error {
	var lastErr error
	attempts := h.cfg.MaxRetries
	if attempts <= 0 {
		attempts = 1
	}

	for i := 0; i < attempts; i++ {
		if err := processor.ProcessWebhook(ctx, provider, payload); err == nil {
			return nil
		} else {
			lastErr = err
			if i < attempts-1 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(h.cfg.RetryDelay):
				}
			}
		}
	}

	return lastErr
}
