// Package email provides email provider abstraction for verification emails.
package email

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
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
	mu         sync.RWMutex
	name       string
	sentEmails []*Email
	results    map[string]*SendResult
	deliveries map[string]*DeliveryStatusResult
	sendFunc   func(ctx context.Context, email *Email) (*SendResult, error)
	healthErr  error
	logger     zerolog.Logger
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

const (
	awsSESAPIVersion       = "2010-12-01"
	defaultProviderTimeout = 30 * time.Second
	sendGridAPIBaseURL     = "https://api.sendgrid.com"
)

// SESProvider implements Provider using AWS SES
type SESProvider struct {
	config      SESConfig
	logger      zerolog.Logger
	httpClient  *http.Client
	endpoint    string
	credentials *credentials.Credentials
	signer      *v4.Signer
}

// NewSESProvider creates a new AWS SES provider
func NewSESProvider(config SESConfig, logger zerolog.Logger) (*SESProvider, error) {
	if config.Region == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "region is required for SES")
	}

	httpClient := security.NewSecureHTTPClient(security.WithTimeout(defaultProviderTimeout))
	creds, signer, err := newAWSSigner(config.Region, config.AccessKeyID, config.SecretAccessKey, httpClient)
	if err != nil {
		return nil, err
	}

	return &SESProvider{
		config:      config,
		logger:      logger.With().Str("provider", "ses").Logger(),
		httpClient:  httpClient,
		endpoint:    fmt.Sprintf("https://email.%s.amazonaws.com/", config.Region),
		credentials: creds,
		signer:      signer,
	}, nil
}

// Name returns the provider name
func (p *SESProvider) Name() string {
	return "ses"
}

// Send sends an email using SES
func (p *SESProvider) Send(ctx context.Context, email *Email) (*SendResult, error) {
	fromAddress, err := resolveEmailFromAddress(email, p.config.FromDomain)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(email.TextBody) == "" && strings.TrimSpace(email.HTMLBody) == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "email body is required for SES")
	}

	form := url.Values{}
	form.Set("Action", "SendEmail")
	form.Set("Version", awsSESAPIVersion)
	form.Set("Source", fromAddress)
	form.Set("Destination.ToAddresses.member.1", email.To)
	form.Set("Message.Subject.Data", email.Subject)
	form.Set("Message.Subject.Charset", "UTF-8")

	if email.TextBody != "" {
		form.Set("Message.Body.Text.Data", email.TextBody)
		form.Set("Message.Body.Text.Charset", "UTF-8")
	}
	if email.HTMLBody != "" {
		form.Set("Message.Body.Html.Data", email.HTMLBody)
		form.Set("Message.Body.Html.Charset", "UTF-8")
	}
	if email.ReplyTo != "" {
		form.Set("ReplyToAddresses.member.1", email.ReplyTo)
	}
	if p.config.ConfigurationSetName != "" {
		form.Set("ConfigurationSetName", p.config.ConfigurationSetName)
	}

	tagIndex := 1
	for key, value := range email.Metadata {
		form.Set(fmt.Sprintf("Tags.member.%d.Name", tagIndex), key)
		form.Set(fmt.Sprintf("Tags.member.%d.Value", tagIndex), value)
		tagIndex++
	}
	for _, tag := range email.Tags {
		form.Set(fmt.Sprintf("Tags.member.%d.Name", tagIndex), "category")
		form.Set(fmt.Sprintf("Tags.member.%d.Value", tagIndex), tag)
		tagIndex++
	}

	body, err := p.doSignedRequest(ctx, form)
	if err != nil {
		return nil, err
	}

	var response struct {
		MessageID string `xml:"SendEmailResult>MessageId"`
	}
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to parse SES response: %v", err)
	}
	if response.MessageID == "" {
		return nil, errors.Wrap(ErrProviderError, "SES response missing message id")
	}

	p.logger.Info().
		Str("to", email.To).
		Str("subject", email.Subject).
		Str("message_id", response.MessageID).
		Msg("email sent via SES")

	return &SendResult{
		Success:   true,
		MessageID: response.MessageID,
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
	return nil, errors.Wrap(ErrProviderError, "SES delivery status is event-driven; use webhooks or cached status")
}

// ParseWebhook parses an SES webhook (SNS notification)
func (p *SESProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	type sesEnvelope struct {
		Type    string `json:"Type"`
		Message string `json:"Message"`
	}
	type recipient struct {
		EmailAddress string `json:"emailAddress"`
	}
	type notification struct {
		NotificationType string `json:"notificationType"`
		Mail             struct {
			MessageID   string    `json:"messageId"`
			Timestamp   time.Time `json:"timestamp"`
			Destination []string  `json:"destination"`
		} `json:"mail"`
		Delivery struct {
			Timestamp time.Time `json:"timestamp"`
		} `json:"delivery"`
		Bounce struct {
			BounceType        string      `json:"bounceType"`
			BounceSubType     string      `json:"bounceSubType"`
			Timestamp         time.Time   `json:"timestamp"`
			BouncedRecipients []recipient `json:"bouncedRecipients"`
		} `json:"bounce"`
		Complaint struct {
			ComplaintFeedbackType string      `json:"complaintFeedbackType"`
			Timestamp             time.Time   `json:"timestamp"`
			ComplainedRecipients  []recipient `json:"complainedRecipients"`
		} `json:"complaint"`
	}

	messageBytes := payload
	var envelope sesEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil && envelope.Message != "" {
		messageBytes = []byte(envelope.Message)
	}

	var event notification
	if err := json.Unmarshal(messageBytes, &event); err != nil {
		return nil, errors.Wrapf(ErrWebhookInvalid, "failed to parse SES webhook: %v", err)
	}

	buildEvent := func(eventType WebhookEventType, ts time.Time, recipientEmail, bounceType, bounceSubtype, complaintType string) WebhookEvent {
		raw := make(map[string]interface{})
		_ = json.Unmarshal(messageBytes, &raw)
		return WebhookEvent{
			EventType:     eventType,
			MessageID:     event.Mail.MessageID,
			Timestamp:     ts,
			RecipientHash: hashEmailRecipient(recipientEmail),
			BounceType:    bounceType,
			BounceSubtype: bounceSubtype,
			ComplaintType: complaintType,
			Raw:           raw,
		}
	}

	switch strings.ToLower(event.NotificationType) {
	case "delivery":
		recipientEmail := firstString(event.Mail.Destination)
		return []WebhookEvent{buildEvent(WebhookEventDelivered, event.Delivery.Timestamp, recipientEmail, "", "", "")}, nil
	case "bounce":
		events := make([]WebhookEvent, 0, len(event.Bounce.BouncedRecipients))
		if len(event.Bounce.BouncedRecipients) == 0 {
			return []WebhookEvent{buildEvent(WebhookEventBounced, event.Bounce.Timestamp, firstString(event.Mail.Destination), event.Bounce.BounceType, event.Bounce.BounceSubType, "")}, nil
		}
		for _, recipient := range event.Bounce.BouncedRecipients {
			events = append(events, buildEvent(WebhookEventBounced, event.Bounce.Timestamp, recipient.EmailAddress, event.Bounce.BounceType, event.Bounce.BounceSubType, ""))
		}
		return events, nil
	case "complaint":
		events := make([]WebhookEvent, 0, len(event.Complaint.ComplainedRecipients))
		if len(event.Complaint.ComplainedRecipients) == 0 {
			return []WebhookEvent{buildEvent(WebhookEventComplaint, event.Complaint.Timestamp, firstString(event.Mail.Destination), "", "", event.Complaint.ComplaintFeedbackType)}, nil
		}
		for _, recipient := range event.Complaint.ComplainedRecipients {
			events = append(events, buildEvent(WebhookEventComplaint, event.Complaint.Timestamp, recipient.EmailAddress, "", "", event.Complaint.ComplaintFeedbackType))
		}
		return events, nil
	default:
		return nil, errors.Wrapf(ErrWebhookInvalid, "unsupported SES notification type: %s", event.NotificationType)
	}
}

// HealthCheck checks if SES is healthy
func (p *SESProvider) HealthCheck(ctx context.Context) error {
	form := url.Values{}
	form.Set("Action", "GetSendQuota")
	form.Set("Version", awsSESAPIVersion)
	_, err := p.doSignedRequest(ctx, form)
	return err
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
type SendGridProvider struct {
	config     SendGridConfig
	logger     zerolog.Logger
	httpClient *http.Client
	baseURL    string
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
		config:     config,
		logger:     logger.With().Str("provider", "sendgrid").Logger(),
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(defaultProviderTimeout)),
		baseURL:    sendGridAPIBaseURL,
	}, nil
}

// Name returns the provider name
func (p *SendGridProvider) Name() string {
	return "sendgrid"
}

// Send sends an email using SendGrid
func (p *SendGridProvider) Send(ctx context.Context, email *Email) (*SendResult, error) {
	fromAddress, err := resolveEmailFromAddress(email, p.config.FromDomain)
	if err != nil {
		return nil, err
	}

	personalization := map[string]interface{}{
		"to": []map[string]string{
			{"email": email.To},
		},
	}
	if len(email.Metadata) > 0 {
		personalization["custom_args"] = email.Metadata
	}

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{personalization},
		"from": map[string]string{
			"email": fromAddress,
		},
		"subject": email.Subject,
	}
	if email.FromName != "" {
		payload["from"].(map[string]string)["name"] = email.FromName
	}
	if email.ReplyTo != "" {
		payload["reply_to"] = map[string]string{"email": email.ReplyTo}
	}
	if len(email.Headers) > 0 {
		payload["headers"] = email.Headers
	}
	if len(email.Tags) > 0 {
		payload["categories"] = email.Tags
	}
	if email.TemplateID != "" {
		payload["template_id"] = email.TemplateID
		if len(email.TemplateData) > 0 {
			payload["personalizations"] = []map[string]interface{}{
				{
					"to":                    personalization["to"],
					"custom_args":           personalization["custom_args"],
					"dynamic_template_data": email.TemplateData,
				},
			}
		}
	} else {
		content := make([]map[string]string, 0, 2)
		if email.TextBody != "" {
			content = append(content, map[string]string{
				"type":  "text/plain",
				"value": email.TextBody,
			})
		}
		if email.HTMLBody != "" {
			content = append(content, map[string]string{
				"type":  "text/html",
				"value": email.HTMLBody,
			})
		}
		if len(content) == 0 {
			return nil, errors.Wrap(ErrInvalidConfig, "email body or template is required for SendGrid")
		}
		payload["content"] = content
	}
	if p.config.SandboxMode {
		payload["mail_settings"] = map[string]interface{}{
			"sandbox_mode": map[string]bool{
				"enable": true,
			},
		}
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to encode SendGrid payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v3/mail/send", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to create SendGrid request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "SendGrid request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to read SendGrid response: %v", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseSendGridError(resp.StatusCode, respBody)
	}

	messageID := resp.Header.Get("X-Message-Id")
	if messageID == "" {
		messageID = resp.Header.Get("X-Message-ID")
	}

	p.logger.Info().
		Str("to", email.To).
		Str("subject", email.Subject).
		Str("message_id", messageID).
		Bool("sandbox", p.config.SandboxMode).
		Msg("email sent via SendGrid")

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
	return nil, errors.Wrap(ErrProviderError, "SendGrid delivery status is event-driven; use webhooks or cached status")
}

// ParseWebhook parses a SendGrid webhook
func (p *SendGridProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	var rawEvents []map[string]interface{}
	if err := json.Unmarshal(payload, &rawEvents); err != nil {
		var envelope struct {
			Events []map[string]interface{} `json:"events"`
		}
		if errEnvelope := json.Unmarshal(payload, &envelope); errEnvelope != nil {
			return nil, errors.Wrapf(ErrWebhookInvalid, "failed to parse SendGrid webhook: %v", err)
		}
		rawEvents = envelope.Events
	}

	events := make([]WebhookEvent, 0, len(rawEvents))
	for _, rawEvent := range rawEvents {
		eventName, _ := rawEvent["event"].(string)
		emailAddress, _ := rawEvent["email"].(string)
		messageID := firstNonEmptyString(
			asString(rawEvent["sg_message_id"]),
			asString(rawEvent["smtp-id"]),
			asString(rawEvent["sg_event_id"]),
		)
		timestamp := parseUnixEventTimestamp(rawEvent["timestamp"])

		event := WebhookEvent{
			MessageID:     strings.Trim(messageID, "<>"),
			Timestamp:     timestamp,
			RecipientHash: hashEmailRecipient(emailAddress),
			Raw:           rawEvent,
		}

		switch strings.ToLower(eventName) {
		case "delivered":
			event.EventType = WebhookEventDelivered
		case "bounce", "dropped", "deferred":
			event.EventType = WebhookEventBounced
			event.BounceType = asString(rawEvent["type"])
			event.BounceSubtype = asString(rawEvent["reason"])
		case "spamreport":
			event.EventType = WebhookEventComplaint
			event.ComplaintType = asString(rawEvent["asm_group_id"])
		case "open":
			event.EventType = WebhookEventOpened
		case "click":
			event.EventType = WebhookEventClicked
		default:
			continue
		}

		events = append(events, event)
	}

	return events, nil
}

// HealthCheck checks if SendGrid is healthy
func (p *SendGridProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/v3/scopes", nil)
	if err != nil {
		return errors.Wrapf(ErrProviderError, "failed to create SendGrid health check: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(ErrProviderError, "SendGrid health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return parseSendGridError(resp.StatusCode, body)
	}

	return nil
}

// Close closes the provider
func (p *SendGridProvider) Close() error {
	return nil
}

// Ensure SendGridProvider implements Provider
var _ Provider = (*SendGridProvider)(nil)

func newAWSSigner(region, accessKeyID, secretAccessKey string, httpClient *http.Client) (*credentials.Credentials, *v4.Signer, error) {
	cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient)
	switch {
	case accessKeyID != "" && secretAccessKey != "":
		cfg = cfg.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
	case accessKeyID != "" || secretAccessKey != "":
		return nil, nil, errors.Wrap(ErrInvalidConfig, "both access_key_id and secret_access_key are required")
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, nil, errors.Wrapf(ErrInvalidConfig, "failed to initialize AWS session: %v", err)
	}

	if sess.Config.Credentials == nil {
		return nil, nil, errors.Wrap(ErrInvalidConfig, "AWS credentials are not configured")
	}

	return sess.Config.Credentials, v4.NewSigner(sess.Config.Credentials), nil
}

func (p *SESProvider) doSignedRequest(ctx context.Context, form url.Values) ([]byte, error) {
	payload := []byte(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to create SES request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	if _, err := p.signer.Sign(req, bytes.NewReader(payload), "ses", p.config.Region, time.Now().UTC()); err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to sign SES request: %v", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "SES request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to read SES response: %v", err)
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseAWSQueryError("SES", resp.StatusCode, body)
	}

	return body, nil
}

func resolveEmailFromAddress(email *Email, fallbackDomain string) (string, error) {
	if strings.TrimSpace(email.From) != "" {
		return strings.TrimSpace(email.From), nil
	}
	if strings.TrimSpace(fallbackDomain) != "" {
		return "no-reply@" + strings.TrimSpace(fallbackDomain), nil
	}
	return "", errors.Wrap(ErrInvalidConfig, "from address is required")
}

func parseSendGridError(statusCode int, body []byte) error {
	var response struct {
		Errors []struct {
			Message string `json:"message"`
			Field   string `json:"field"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &response); err == nil && len(response.Errors) > 0 {
		first := response.Errors[0]
		if first.Field != "" {
			return errors.Wrapf(ErrDeliveryFailed, "SendGrid API error (%d): %s (%s)", statusCode, first.Message, first.Field)
		}
		return errors.Wrapf(ErrDeliveryFailed, "SendGrid API error (%d): %s", statusCode, first.Message)
	}

	return errors.Wrapf(ErrDeliveryFailed, "SendGrid API error (%d): %s", statusCode, strings.TrimSpace(string(body)))
}

func parseAWSQueryError(provider string, statusCode int, body []byte) error {
	var response struct {
		Error struct {
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		} `xml:"Error"`
	}

	if err := xml.Unmarshal(body, &response); err == nil && response.Error.Code != "" {
		return errors.Wrapf(ErrDeliveryFailed, "%s API error (%d): %s: %s", provider, statusCode, response.Error.Code, response.Error.Message)
	}

	return errors.Wrapf(ErrDeliveryFailed, "%s API error (%d): %s", provider, statusCode, strings.TrimSpace(string(body)))
}

func hashEmailRecipient(recipient string) string {
	normalized := strings.TrimSpace(strings.ToLower(recipient))
	if normalized == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func parseUnixEventTimestamp(value interface{}) time.Time {
	switch ts := value.(type) {
	case float64:
		return time.Unix(int64(ts), 0).UTC()
	case int64:
		return time.Unix(ts, 0).UTC()
	case json.Number:
		parsed, err := ts.Int64()
		if err == nil {
			return time.Unix(parsed, 0).UTC()
		}
	}

	return time.Now().UTC()
}

func asString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		return ""
	}
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

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
