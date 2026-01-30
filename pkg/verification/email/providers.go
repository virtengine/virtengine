// Package email provides email provider abstraction for verification emails.
package email

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
)

// ============================================================================
// Email Provider Interface
// ============================================================================

// Provider defines the interface for email sending providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Send sends an email using the provider
	Send(ctx context.Context, email *Email) (*SendResult, error)

	// BatchSend sends multiple emails (optional optimization)
	BatchSend(ctx context.Context, emails []*Email) ([]*SendResult, error)

	// GetDeliveryStatus queries the delivery status of a message
	GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error)

	// ParseWebhook parses a webhook payload into events
	ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error)

	// HealthCheck checks if the provider is healthy
	HealthCheck(ctx context.Context) error

	// Close closes the provider and releases resources
	Close() error
}

// ============================================================================
// Email Types
// ============================================================================

// Email represents an email to be sent
type Email struct {
	// To is the recipient email address
	To string `json:"to"`

	// From is the sender email address (optional, uses config default)
	From string `json:"from,omitempty"`

	// FromName is the sender display name
	FromName string `json:"from_name,omitempty"`

	// Subject is the email subject
	Subject string `json:"subject"`

	// TextBody is the plain text body
	TextBody string `json:"text_body,omitempty"`

	// HTMLBody is the HTML body
	HTMLBody string `json:"html_body,omitempty"`

	// TemplateID is the ID of a pre-defined template
	TemplateID string `json:"template_id,omitempty"`

	// TemplateData is the data to populate the template
	TemplateData map[string]interface{} `json:"template_data,omitempty"`

	// Tags are custom tags for analytics
	Tags []string `json:"tags,omitempty"`

	// Metadata contains custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// ReplyTo is the reply-to address
	ReplyTo string `json:"reply_to,omitempty"`

	// Headers contains custom headers
	Headers map[string]string `json:"headers,omitempty"`
}

// SendResult contains the result of sending an email
type SendResult struct {
	// Success indicates if the send was successful
	Success bool `json:"success"`

	// MessageID is the provider's message ID
	MessageID string `json:"message_id"`

	// Timestamp is when the email was sent
	Timestamp time.Time `json:"timestamp"`

	// Error is the error message if sending failed
	Error string `json:"error,omitempty"`

	// ErrorCode is the error code if sending failed
	ErrorCode string `json:"error_code,omitempty"`

	// Provider is the provider name
	Provider string `json:"provider"`
}

// DeliveryStatusResult contains the delivery status of a message
type DeliveryStatusResult struct {
	// MessageID is the message ID
	MessageID string `json:"message_id"`

	// Status is the delivery status
	Status DeliveryStatus `json:"status"`

	// Timestamp is when the status was recorded
	Timestamp time.Time `json:"timestamp"`

	// BounceType is the type of bounce (if applicable)
	BounceType string `json:"bounce_type,omitempty"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Mock Provider (for testing)
// ============================================================================

// MockProvider implements Provider for testing
type MockProvider struct {
	mu           sync.RWMutex
	name         string
	sentEmails   []*Email
	results      map[string]*SendResult
	deliveries   map[string]*DeliveryStatusResult
	sendFunc     func(ctx context.Context, email *Email) (*SendResult, error)
	healthErr    error
	logger       zerolog.Logger
}

// MockProviderOption is a functional option for configuring MockProvider
type MockProviderOption func(*MockProvider)

// WithMockSendFunc sets a custom send function
func WithMockSendFunc(fn func(ctx context.Context, email *Email) (*SendResult, error)) MockProviderOption {
	return func(p *MockProvider) {
		p.sendFunc = fn
	}
}

// WithMockHealthError sets the health check error
func WithMockHealthError(err error) MockProviderOption {
	return func(p *MockProvider) {
		p.healthErr = err
	}
}

// NewMockProvider creates a new mock email provider
func NewMockProvider(logger zerolog.Logger, opts ...MockProviderOption) *MockProvider {
	p := &MockProvider{
		name:       "mock",
		sentEmails: make([]*Email, 0),
		results:    make(map[string]*SendResult),
		deliveries: make(map[string]*DeliveryStatusResult),
		logger:     logger.With().Str("provider", "mock").Logger(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return p.name
}

// Send sends an email using the mock provider
func (p *MockProvider) Send(ctx context.Context, email *Email) (*SendResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Use custom send function if provided
	if p.sendFunc != nil {
		return p.sendFunc(ctx, email)
	}

	// Default: always succeed
	messageID := fmt.Sprintf("mock-%d", len(p.sentEmails)+1)
	result := &SendResult{
		Success:   true,
		MessageID: messageID,
		Timestamp: time.Now(),
		Provider:  p.name,
	}

	p.sentEmails = append(p.sentEmails, email)
	p.results[messageID] = result
	p.deliveries[messageID] = &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}

	p.logger.Debug().
		Str("to", email.To).
		Str("subject", email.Subject).
		Str("message_id", messageID).
		Msg("mock email sent")

	return result, nil
}

// BatchSend sends multiple emails
func (p *MockProvider) BatchSend(ctx context.Context, emails []*Email) ([]*SendResult, error) {
	results := make([]*SendResult, len(emails))
	for i, email := range emails {
		result, err := p.Send(ctx, email)
		if err != nil {
			results[i] = &SendResult{
				Success:  false,
				Error:    err.Error(),
				Provider: p.name,
			}
		} else {
			results[i] = result
		}
	}
	return results, nil
}

// GetDeliveryStatus returns the delivery status of a message
func (p *MockProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status, ok := p.deliveries[messageID]
	if !ok {
		return nil, errors.Wrapf(ErrChallengeNotFound, "message not found: %s", messageID)
	}

	return status, nil
}

// ParseWebhook parses a webhook payload
func (p *MockProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	// Mock implementation - return empty events
	return []WebhookEvent{}, nil
}

// HealthCheck checks if the provider is healthy
func (p *MockProvider) HealthCheck(ctx context.Context) error {
	return p.healthErr
}

// Close closes the provider
func (p *MockProvider) Close() error {
	return nil
}

// GetSentEmails returns all sent emails (for testing)
func (p *MockProvider) GetSentEmails() []*Email {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Email, len(p.sentEmails))
	copy(result, p.sentEmails)
	return result
}

// ClearSentEmails clears all sent emails (for testing)
func (p *MockProvider) ClearSentEmails() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sentEmails = make([]*Email, 0)
}

// SetDeliveryStatus sets the delivery status for a message (for testing)
func (p *MockProvider) SetDeliveryStatus(messageID string, status DeliveryStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.deliveries[messageID] = &DeliveryStatusResult{
		MessageID: messageID,
		Status:    status,
		Timestamp: time.Now(),
	}
}

// Ensure MockProvider implements Provider
var _ Provider = (*MockProvider)(nil)

// ============================================================================
// SES Provider
// ============================================================================

// SESConfig contains configuration for AWS SES
type SESConfig struct {
	// Region is the AWS region
	Region string `json:"region"`

	// AccessKeyID is the AWS access key ID
	AccessKeyID string `json:"access_key_id"`

	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secret_access_key"`

	// ConfigurationSetName is the SES configuration set name
	ConfigurationSetName string `json:"configuration_set_name,omitempty"`

	// FromDomain is the verified sending domain
	FromDomain string `json:"from_domain"`
}

// SESProvider implements Provider using AWS SES
// Note: This is a placeholder - actual AWS SDK integration would be added
type SESProvider struct {
	config SESConfig
	logger zerolog.Logger
}

// NewSESProvider creates a new AWS SES provider
func NewSESProvider(config SESConfig, logger zerolog.Logger) (*SESProvider, error) {
	if config.Region == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "region is required for SES")
	}
	if config.FromDomain == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "from_domain is required for SES")
	}

	return &SESProvider{
		config: config,
		logger: logger.With().Str("provider", "ses").Logger(),
	}, nil
}

// Name returns the provider name
func (p *SESProvider) Name() string {
	return "ses"
}

// Send sends an email using SES
func (p *SESProvider) Send(ctx context.Context, email *Email) (*SendResult, error) {
	// TODO: Implement actual SES API call using AWS SDK
	// This is a placeholder implementation
	p.logger.Info().
		Str("to", email.To).
		Str("subject", email.Subject).
		Msg("SES send called (placeholder)")

	messageID := fmt.Sprintf("ses-%d", time.Now().UnixNano())
	return &SendResult{
		Success:   true,
		MessageID: messageID,
		Timestamp: time.Now(),
		Provider:  "ses",
	}, nil
}

// BatchSend sends multiple emails
func (p *SESProvider) BatchSend(ctx context.Context, emails []*Email) ([]*SendResult, error) {
	results := make([]*SendResult, len(emails))
	for i, email := range emails {
		result, err := p.Send(ctx, email)
		if err != nil {
			results[i] = &SendResult{
				Success:  false,
				Error:    err.Error(),
				Provider: "ses",
			}
		} else {
			results[i] = result
		}
	}
	return results, nil
}

// GetDeliveryStatus returns the delivery status
func (p *SESProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	// TODO: Implement using SES events or CloudWatch
	return &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}, nil
}

// ParseWebhook parses an SES webhook (SNS notification)
func (p *SESProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	// TODO: Implement SNS notification parsing for SES events
	return []WebhookEvent{}, nil
}

// HealthCheck checks if SES is healthy
func (p *SESProvider) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual health check
	return nil
}

// Close closes the provider
func (p *SESProvider) Close() error {
	return nil
}

// Ensure SESProvider implements Provider
var _ Provider = (*SESProvider)(nil)

// ============================================================================
// SendGrid Provider
// ============================================================================

// SendGridConfig contains configuration for SendGrid
type SendGridConfig struct {
	// APIKey is the SendGrid API key
	APIKey string `json:"api_key"`

	// APIKeyID is the API key ID (optional, for webhook validation)
	APIKeyID string `json:"api_key_id,omitempty"`

	// FromDomain is the verified sending domain
	FromDomain string `json:"from_domain"`

	// SandboxMode enables sandbox mode
	SandboxMode bool `json:"sandbox_mode"`
}

// SendGridProvider implements Provider using SendGrid
// Note: This is a placeholder - actual SendGrid SDK integration would be added
type SendGridProvider struct {
	config SendGridConfig
	logger zerolog.Logger
}

// NewSendGridProvider creates a new SendGrid provider
func NewSendGridProvider(config SendGridConfig, logger zerolog.Logger) (*SendGridProvider, error) {
	if config.APIKey == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "api_key is required for SendGrid")
	}
	if config.FromDomain == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "from_domain is required for SendGrid")
	}

	return &SendGridProvider{
		config: config,
		logger: logger.With().Str("provider", "sendgrid").Logger(),
	}, nil
}

// Name returns the provider name
func (p *SendGridProvider) Name() string {
	return "sendgrid"
}

// Send sends an email using SendGrid
func (p *SendGridProvider) Send(ctx context.Context, email *Email) (*SendResult, error) {
	// TODO: Implement actual SendGrid API call
	p.logger.Info().
		Str("to", email.To).
		Str("subject", email.Subject).
		Bool("sandbox", p.config.SandboxMode).
		Msg("SendGrid send called (placeholder)")

	messageID := fmt.Sprintf("sg-%d", time.Now().UnixNano())
	return &SendResult{
		Success:   true,
		MessageID: messageID,
		Timestamp: time.Now(),
		Provider:  "sendgrid",
	}, nil
}

// BatchSend sends multiple emails
func (p *SendGridProvider) BatchSend(ctx context.Context, emails []*Email) ([]*SendResult, error) {
	results := make([]*SendResult, len(emails))
	for i, email := range emails {
		result, err := p.Send(ctx, email)
		if err != nil {
			results[i] = &SendResult{
				Success:  false,
				Error:    err.Error(),
				Provider: "sendgrid",
			}
		} else {
			results[i] = result
		}
	}
	return results, nil
}

// GetDeliveryStatus returns the delivery status
func (p *SendGridProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	// TODO: Implement using SendGrid events API
	return &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}, nil
}

// ParseWebhook parses a SendGrid webhook
func (p *SendGridProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	// TODO: Implement SendGrid webhook parsing
	return []WebhookEvent{}, nil
}

// HealthCheck checks if SendGrid is healthy
func (p *SendGridProvider) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual health check
	return nil
}

// Close closes the provider
func (p *SendGridProvider) Close() error {
	return nil
}

// Ensure SendGridProvider implements Provider
var _ Provider = (*SendGridProvider)(nil)

// ============================================================================
// Provider Factory
// ============================================================================

// NewProvider creates a new email provider based on configuration
func NewProvider(config Config, logger zerolog.Logger) (Provider, error) {
	switch config.Provider {
	case "ses":
		sesConfig := SESConfig{
			Region:               config.ProviderConfig["region"],
			AccessKeyID:          config.ProviderConfig["access_key_id"],
			SecretAccessKey:      config.ProviderConfig["secret_access_key"],
			ConfigurationSetName: config.ProviderConfig["configuration_set_name"],
			FromDomain:           config.ProviderConfig["from_domain"],
		}
		return NewSESProvider(sesConfig, logger)

	case "sendgrid":
		sgConfig := SendGridConfig{
			APIKey:      config.ProviderConfig["api_key"],
			APIKeyID:    config.ProviderConfig["api_key_id"],
			FromDomain:  config.ProviderConfig["from_domain"],
			SandboxMode: config.ProviderConfig["sandbox_mode"] == "true",
		}
		return NewSendGridProvider(sgConfig, logger)

	case "mock":
		return NewMockProvider(logger), nil

	default:
		return nil, errors.Wrapf(ErrInvalidConfig, "unsupported provider: %s", config.Provider)
	}
}
