// Package sms provides SMS provider abstraction for verification SMS.
package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
)

// Provider name constants
const (
	providerTwilio = "twilio"
	providerSNS    = "sns"
)

// ============================================================================
// SMS Provider Interface
// ============================================================================

// Provider defines the interface for SMS sending providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Send sends an SMS using the provider
	Send(ctx context.Context, msg *SMSMessage) (*SendResult, error)

	// LookupCarrier performs carrier lookup on a phone number
	LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error)

	// GetDeliveryStatus queries the delivery status of a message
	GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error)

	// ParseWebhook parses a webhook payload into events
	ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error)

	// HealthCheck checks if the provider is healthy
	HealthCheck(ctx context.Context) error

	// SupportedRegions returns the list of supported country codes
	SupportedRegions() []string

	// Close closes the provider and releases resources
	Close() error
}

// ============================================================================
// SMS Types
// ============================================================================

// SMSMessage represents an SMS to be sent
type SMSMessage struct {
	// To is the recipient phone number in E.164 format
	To string `json:"to"`

	// From is the sender phone number or sender ID
	From string `json:"from,omitempty"`

	// Body is the message body
	Body string `json:"body"`

	// MediaURLs is a list of media URLs (for MMS)
	MediaURLs []string `json:"media_urls,omitempty"`

	// Tags are custom tags for analytics
	Tags []string `json:"tags,omitempty"`

	// Metadata contains custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// StatusCallback is the webhook URL for delivery status
	StatusCallback string `json:"status_callback,omitempty"`

	// ValidityPeriod is how long the SMS is valid for delivery (seconds)
	ValidityPeriod int `json:"validity_period,omitempty"`

	// MaxPrice is the maximum price to pay for delivery
	MaxPrice string `json:"max_price,omitempty"`
}

// SendResult contains the result of sending an SMS
type SendResult struct {
	// Success indicates if the send was successful
	Success bool `json:"success"`

	// MessageID is the provider's message ID
	MessageID string `json:"message_id"`

	// Timestamp is when the SMS was sent
	Timestamp time.Time `json:"timestamp"`

	// Error is the error message if sending failed
	Error string `json:"error,omitempty"`

	// ErrorCode is the error code if sending failed
	ErrorCode string `json:"error_code,omitempty"`

	// Provider is the provider name
	Provider string `json:"provider"`

	// Price is the price of the SMS
	Price string `json:"price,omitempty"`

	// PriceUnit is the currency unit
	PriceUnit string `json:"price_unit,omitempty"`

	// SegmentCount is the number of SMS segments
	SegmentCount int `json:"segment_count,omitempty"`
}

// CarrierLookupResult contains carrier information for a phone number
type CarrierLookupResult struct {
	// PhoneNumber is the phone number in E.164 format
	PhoneNumber string `json:"phone_number"`

	// CountryCode is the ISO country code
	CountryCode string `json:"country_code"`

	// CarrierName is the name of the carrier
	CarrierName string `json:"carrier_name"`

	// CarrierType is the type of carrier
	CarrierType CarrierType `json:"carrier_type"`

	// NetworkCode is the mobile network code
	NetworkCode string `json:"network_code,omitempty"`

	// IsVoIP indicates if this is a VoIP number
	IsVoIP bool `json:"is_voip"`

	// IsMobile indicates if this is a mobile number
	IsMobile bool `json:"is_mobile"`

	// IsPrepaid indicates if this is a prepaid number
	IsPrepaid bool `json:"is_prepaid"`

	// IsValid indicates if the number is valid
	IsValid bool `json:"is_valid"`

	// IsPorted indicates if the number was ported
	IsPorted bool `json:"is_ported"`

	// RiskScore is a risk score (0-100)
	RiskScore uint32 `json:"risk_score"`

	// RiskFactors lists detected risk factors
	RiskFactors []string `json:"risk_factors,omitempty"`

	// LookupTimestamp is when the lookup was performed
	LookupTimestamp time.Time `json:"lookup_timestamp"`
}

// DeliveryStatusResult contains the delivery status of a message
type DeliveryStatusResult struct {
	// MessageID is the message ID
	MessageID string `json:"message_id"`

	// Status is the delivery status
	Status DeliveryStatus `json:"status"`

	// Timestamp is when the status was recorded
	Timestamp time.Time `json:"timestamp"`

	// ErrorCode is the error code (if applicable)
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message
	ErrorMessage string `json:"error_message,omitempty"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Failover Provider
// ============================================================================

// FailoverProvider wraps multiple providers with automatic failover
type FailoverProvider struct {
	primary         Provider
	secondary       Provider
	logger          zerolog.Logger
	metrics         *ProviderMetrics
	mu              sync.RWMutex
	primaryFails    int64
	lastPrimaryFail time.Time
}

// ProviderMetrics tracks provider metrics
type ProviderMetrics struct {
	totalSent         int64
	primarySent       int64
	secondarySent     int64
	primaryFailures   int64
	secondaryFailures int64
	totalFailures     int64
}

// NewFailoverProvider creates a new failover provider
func NewFailoverProvider(primary Provider, secondary Provider, logger zerolog.Logger) *FailoverProvider {
	return &FailoverProvider{
		primary:   primary,
		secondary: secondary,
		logger:    logger.With().Str("component", "failover_provider").Logger(),
		metrics:   &ProviderMetrics{},
	}
}

// Name returns the provider name
func (p *FailoverProvider) Name() string {
	return fmt.Sprintf("failover(%s,%s)", p.primary.Name(), p.getSecondaryName())
}

func (p *FailoverProvider) getSecondaryName() string {
	if p.secondary != nil {
		return p.secondary.Name()
	}
	return "none"
}

// Send sends an SMS with failover support
func (p *FailoverProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	// Try primary first
	result, err := p.primary.Send(ctx, msg)
	if err == nil && result.Success {
		atomic.AddInt64(&p.metrics.primarySent, 1)
		atomic.AddInt64(&p.metrics.totalSent, 1)
		return result, nil
	}

	// Primary failed
	atomic.AddInt64(&p.metrics.primaryFailures, 1)
	p.mu.Lock()
	p.primaryFails++
	p.lastPrimaryFail = time.Now()
	p.mu.Unlock()

	p.logger.Warn().
		Err(err).
		Str("provider", p.primary.Name()).
		Str("to", MaskPhoneNumber(msg.To)).
		Msg("primary provider failed, attempting failover")

	// Try secondary if available
	if p.secondary == nil {
		atomic.AddInt64(&p.metrics.totalFailures, 1)
		return nil, errors.Wrapf(ErrPrimaryProviderFailed, "no failover provider: %v", err)
	}

	result, err = p.secondary.Send(ctx, msg)
	if err == nil && result.Success {
		atomic.AddInt64(&p.metrics.secondarySent, 1)
		atomic.AddInt64(&p.metrics.totalSent, 1)
		result.Provider = p.secondary.Name()
		p.logger.Info().
			Str("provider", p.secondary.Name()).
			Str("message_id", result.MessageID).
			Msg("failover provider succeeded")
		return result, nil
	}

	// Both failed
	atomic.AddInt64(&p.metrics.secondaryFailures, 1)
	atomic.AddInt64(&p.metrics.totalFailures, 1)

	p.logger.Error().
		Err(err).
		Str("primary", p.primary.Name()).
		Str("secondary", p.secondary.Name()).
		Msg("all providers failed")

	return nil, ErrAllProvidersFailed
}

// LookupCarrier performs carrier lookup using primary provider
func (p *FailoverProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	result, err := p.primary.LookupCarrier(ctx, phoneNumber)
	if err == nil {
		return result, nil
	}

	// Try secondary
	if p.secondary != nil {
		return p.secondary.LookupCarrier(ctx, phoneNumber)
	}

	return nil, err
}

// GetDeliveryStatus gets delivery status from the appropriate provider
func (p *FailoverProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	// Try primary first
	result, err := p.primary.GetDeliveryStatus(ctx, messageID)
	if err == nil {
		return result, nil
	}

	// Try secondary
	if p.secondary != nil {
		return p.secondary.GetDeliveryStatus(ctx, messageID)
	}

	return nil, err
}

// ParseWebhook parses webhooks from either provider
func (p *FailoverProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	// Try primary
	events, err := p.primary.ParseWebhook(ctx, payload, signature)
	if err == nil && len(events) > 0 {
		return events, nil
	}

	// Try secondary
	if p.secondary != nil {
		return p.secondary.ParseWebhook(ctx, payload, signature)
	}

	return events, err
}

// HealthCheck checks both providers
func (p *FailoverProvider) HealthCheck(ctx context.Context) error {
	primaryErr := p.primary.HealthCheck(ctx)
	if primaryErr == nil {
		return nil
	}

	if p.secondary != nil {
		return p.secondary.HealthCheck(ctx)
	}

	return primaryErr
}

// SupportedRegions returns the union of supported regions
func (p *FailoverProvider) SupportedRegions() []string {
	regions := make(map[string]bool)
	for _, r := range p.primary.SupportedRegions() {
		regions[r] = true
	}
	if p.secondary != nil {
		for _, r := range p.secondary.SupportedRegions() {
			regions[r] = true
		}
	}

	result := make([]string, 0, len(regions))
	for r := range regions {
		result = append(result, r)
	}
	return result
}

// Close closes both providers
func (p *FailoverProvider) Close() error {
	var errs []error
	if err := p.primary.Close(); err != nil {
		errs = append(errs, err)
	}
	if p.secondary != nil {
		if err := p.secondary.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// GetMetrics returns provider metrics
func (p *FailoverProvider) GetMetrics() map[string]int64 {
	return map[string]int64{
		"total_sent":         atomic.LoadInt64(&p.metrics.totalSent),
		"primary_sent":       atomic.LoadInt64(&p.metrics.primarySent),
		"secondary_sent":     atomic.LoadInt64(&p.metrics.secondarySent),
		"primary_failures":   atomic.LoadInt64(&p.metrics.primaryFailures),
		"secondary_failures": atomic.LoadInt64(&p.metrics.secondaryFailures),
		"total_failures":     atomic.LoadInt64(&p.metrics.totalFailures),
	}
}

// Ensure FailoverProvider implements Provider
var _ Provider = (*FailoverProvider)(nil)

// ============================================================================
// Mock Provider
// ============================================================================

// MockProvider implements Provider for testing
type MockProvider struct {
	mu             sync.RWMutex
	name           string
	sentMessages   []*SMSMessage
	results        map[string]*SendResult
	deliveries     map[string]*DeliveryStatusResult
	carrierResults map[string]*CarrierLookupResult
	sendFunc       func(ctx context.Context, msg *SMSMessage) (*SendResult, error)
	lookupFunc     func(ctx context.Context, phone string) (*CarrierLookupResult, error)
	healthErr      error
	regions        []string
	logger         zerolog.Logger
}

// MockProviderOption is a functional option for configuring MockProvider
type MockProviderOption func(*MockProvider)

// WithMockSendFunc sets a custom send function
func WithMockSendFunc(fn func(ctx context.Context, msg *SMSMessage) (*SendResult, error)) MockProviderOption {
	return func(p *MockProvider) {
		p.sendFunc = fn
	}
}

// WithMockLookupFunc sets a custom carrier lookup function
func WithMockLookupFunc(fn func(ctx context.Context, phone string) (*CarrierLookupResult, error)) MockProviderOption {
	return func(p *MockProvider) {
		p.lookupFunc = fn
	}
}

// WithMockHealthError sets the health check error
func WithMockHealthError(err error) MockProviderOption {
	return func(p *MockProvider) {
		p.healthErr = err
	}
}

// WithMockRegions sets the supported regions
func WithMockRegions(regions []string) MockProviderOption {
	return func(p *MockProvider) {
		p.regions = regions
	}
}

// NewMockProvider creates a new mock SMS provider
func NewMockProvider(logger zerolog.Logger, opts ...MockProviderOption) *MockProvider {
	p := &MockProvider{
		name:           "mock",
		sentMessages:   make([]*SMSMessage, 0),
		results:        make(map[string]*SendResult),
		deliveries:     make(map[string]*DeliveryStatusResult),
		carrierResults: make(map[string]*CarrierLookupResult),
		regions:        []string{"US", "CA", "GB", "AU"},
		logger:         logger.With().Str("provider", "mock").Logger(),
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

// Send sends an SMS using the mock provider
func (p *MockProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Use custom send function if provided
	if p.sendFunc != nil {
		return p.sendFunc(ctx, msg)
	}

	// Default: always succeed
	messageID := fmt.Sprintf("mock-sms-%d", len(p.sentMessages)+1)
	result := &SendResult{
		Success:      true,
		MessageID:    messageID,
		Timestamp:    time.Now(),
		Provider:     p.name,
		SegmentCount: 1,
	}

	p.sentMessages = append(p.sentMessages, msg)
	p.results[messageID] = result
	p.deliveries[messageID] = &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}

	p.logger.Debug().
		Str("to", MaskPhoneNumber(msg.To)).
		Str("message_id", messageID).
		Int("body_length", len(msg.Body)).
		Msg("mock SMS sent")

	return result, nil
}

// LookupCarrier performs carrier lookup
func (p *MockProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Use custom lookup function if provided
	if p.lookupFunc != nil {
		return p.lookupFunc(ctx, phoneNumber)
	}

	// Check pre-configured results
	if result, ok := p.carrierResults[phoneNumber]; ok {
		return result, nil
	}

	// Default: return mobile carrier
	return &CarrierLookupResult{
		PhoneNumber:     phoneNumber,
		CountryCode:     "US",
		CarrierName:     "Mock Carrier",
		CarrierType:     CarrierTypeMobile,
		IsVoIP:          false,
		IsMobile:        true,
		IsValid:         true,
		RiskScore:       10,
		LookupTimestamp: time.Now(),
	}, nil
}

// GetDeliveryStatus returns the delivery status
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
	return []WebhookEvent{}, nil
}

// HealthCheck checks if the provider is healthy
func (p *MockProvider) HealthCheck(ctx context.Context) error {
	return p.healthErr
}

// SupportedRegions returns supported regions
func (p *MockProvider) SupportedRegions() []string {
	return p.regions
}

// Close closes the provider
func (p *MockProvider) Close() error {
	return nil
}

// SetCarrierResult sets the carrier lookup result for a phone number
func (p *MockProvider) SetCarrierResult(phoneNumber string, result *CarrierLookupResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.carrierResults[phoneNumber] = result
}

// SetDeliveryStatus sets the delivery status for a message
func (p *MockProvider) SetDeliveryStatus(messageID string, status DeliveryStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.deliveries[messageID] = &DeliveryStatusResult{
		MessageID: messageID,
		Status:    status,
		Timestamp: time.Now(),
	}
}

// GetSentMessages returns all sent messages (for testing)
func (p *MockProvider) GetSentMessages() []*SMSMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*SMSMessage, len(p.sentMessages))
	copy(result, p.sentMessages)
	return result
}

// ClearSentMessages clears all sent messages (for testing)
func (p *MockProvider) ClearSentMessages() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sentMessages = make([]*SMSMessage, 0)
}

// Ensure MockProvider implements Provider
var _ Provider = (*MockProvider)(nil)

// ============================================================================
// Twilio Provider (Placeholder)
// ============================================================================

const (
	awsSNSAPIVersion      = "2010-03-31"
	defaultSMSTimeout     = 30 * time.Second
	transactionalSMSType  = "Transactional"
	awsSNSOriginationAttr = "AWS.MM.SMS.OriginationNumber"
)

// TwilioProvider implements Provider using Twilio
type TwilioProvider struct {
	gateway SMSGateway
}

// NewTwilioProvider creates a new Twilio provider
func NewTwilioProvider(config ProviderConfig, logger zerolog.Logger) (*TwilioProvider, error) {
	gateway, err := NewTwilioGateway(config, logger)
	if err != nil {
		return nil, err
	}

	return &TwilioProvider{gateway: gateway}, nil
}

// Name returns the provider name
func (p *TwilioProvider) Name() string {
	return p.gateway.Name()
}

// Send sends an SMS using Twilio
func (p *TwilioProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	return p.gateway.Send(ctx, msg)
}

// LookupCarrier performs carrier lookup using Twilio Lookup API
func (p *TwilioProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	return p.gateway.LookupCarrier(ctx, phoneNumber)
}

// GetDeliveryStatus gets delivery status from Twilio
func (p *TwilioProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	return p.gateway.GetDeliveryStatus(ctx, messageID)
}

// ParseWebhook parses a Twilio webhook
func (p *TwilioProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	return p.gateway.ParseWebhook(ctx, payload, signature)
}

// HealthCheck checks if Twilio is healthy
func (p *TwilioProvider) HealthCheck(ctx context.Context) error {
	return p.gateway.HealthCheck(ctx)
}

// SupportedRegions returns Twilio's supported regions
func (p *TwilioProvider) SupportedRegions() []string {
	return p.gateway.SupportedRegions()
}

// Close closes the provider
func (p *TwilioProvider) Close() error {
	return p.gateway.Close()
}

// Ensure TwilioProvider implements Provider
var _ Provider = (*TwilioProvider)(nil)

// ============================================================================
// AWS SNS Provider (Placeholder)
// ============================================================================

// SNSProvider implements Provider using AWS SNS
type SNSProvider struct {
	config      ProviderConfig
	logger      zerolog.Logger
	httpClient  *http.Client
	endpoint    string
	credentials *credentials.Credentials
	signer      *v4.Signer
}

// NewSNSProvider creates a new AWS SNS provider
func NewSNSProvider(config ProviderConfig, logger zerolog.Logger) (*SNSProvider, error) {
	if config.Region == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "region is required for SNS")
	}

	httpClient := security.NewSecureHTTPClient(security.WithTimeout(defaultSMSTimeout))
	accessKeyID, secretAccessKey := resolveSNSCredentials(config)
	creds, signer, err := newAWSSMSSigner(config.Region, accessKeyID, secretAccessKey, httpClient)
	if err != nil {
		return nil, err
	}

	return &SNSProvider{
		config:      config,
		logger:      logger.With().Str("provider", providerSNS).Logger(),
		httpClient:  httpClient,
		endpoint:    fmt.Sprintf("https://sns.%s.amazonaws.com/", config.Region),
		credentials: creds,
		signer:      signer,
	}, nil
}

// Name returns the provider name
func (p *SNSProvider) Name() string {
	return providerSNS
}

// Send sends an SMS using AWS SNS
func (p *SNSProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	form := url.Values{}
	form.Set("Action", "Publish")
	form.Set("Version", awsSNSAPIVersion)
	form.Set("PhoneNumber", msg.To)
	form.Set("Message", msg.Body)

	attributeIndex := 1
	addAttribute := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		form.Set(fmt.Sprintf("MessageAttributes.entry.%d.Name", attributeIndex), name)
		form.Set(fmt.Sprintf("MessageAttributes.entry.%d.Value.DataType", attributeIndex), "String")
		form.Set(fmt.Sprintf("MessageAttributes.entry.%d.Value.StringValue", attributeIndex), value)
		attributeIndex++
	}

	addAttribute("AWS.SNS.SMS.SMSType", transactionalSMSType)
	addAttribute("AWS.SNS.SMS.SenderID", p.config.SenderID)
	addAttribute(awsSNSOriginationAttr, p.config.FromNumber)
	addAttribute("AWS.SNS.SMS.MaxPrice", firstNonEmptySMSString(msg.MaxPrice))

	body, err := p.doSignedRequest(ctx, form)
	if err != nil {
		return nil, err
	}

	var response struct {
		MessageID string `xml:"PublishResult>MessageId"`
	}
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to parse SNS response: %v", err)
	}
	if response.MessageID == "" {
		return nil, errors.Wrap(ErrProviderError, "SNS response missing message id")
	}

	p.logger.Info().
		Str("to", MaskPhoneNumber(msg.To)).
		Str("message_id", response.MessageID).
		Msg("SMS sent via SNS")

	return &SendResult{
		Success:      true,
		MessageID:    response.MessageID,
		Timestamp:    time.Now(),
		Provider:     providerSNS,
		SegmentCount: 1,
	}, nil
}

// LookupCarrier is not supported by SNS
func (p *SNSProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	return nil, errors.Wrap(ErrCarrierLookupFailed, "carrier lookup not supported by SNS")
}

// GetDeliveryStatus gets delivery status from SNS
func (p *SNSProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	return nil, errors.Wrap(ErrProviderError, "SNS delivery status is event-driven; use webhooks or cached status")
}

// ParseWebhook parses an SNS webhook (via Lambda/SQS)
func (p *SNSProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	type snsEnvelope struct {
		Type    string `json:"Type"`
		Message string `json:"Message"`
	}
	type deliveryStatus struct {
		MessageID string `json:"messageId"`
		Status    string `json:"status"`
		Provider  string `json:"providerResponse"`
	}

	messageBytes := payload
	var envelope snsEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil && envelope.Message != "" {
		messageBytes = []byte(envelope.Message)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(messageBytes, &raw); err != nil {
		return nil, errors.Wrapf(ErrWebhookInvalid, "failed to parse SNS webhook: %v", err)
	}

	status := strings.ToLower(firstNonEmptySMSString(asSMSString(raw["status"]), asSMSString(raw["deliveryStatus"]), asSMSString(raw["eventType"])))
	messageID := firstNonEmptySMSString(
		asSMSString(raw["messageId"]),
		asSMSString(raw["messageID"]),
	)
	if delivery, ok := raw["delivery"].(map[string]interface{}); ok {
		messageID = firstNonEmptySMSString(messageID, asSMSString(delivery["messageId"]))
		status = firstNonEmptySMSString(status, strings.ToLower(asSMSString(delivery["status"])))
	}

	event := WebhookEvent{
		MessageID: messageID,
		Timestamp: time.Now().UTC(),
		Provider:  providerSNS,
		Raw:       raw,
	}

	switch status {
	case "delivered", "success":
		event.EventType = WebhookEventDelivered
	case "sent", "published":
		event.EventType = WebhookEventSent
	case "failed":
		event.EventType = WebhookEventFailed
	case "undelivered":
		event.EventType = WebhookEventUndelivered
	default:
		return nil, errors.Wrapf(ErrWebhookInvalid, "unsupported SNS delivery status: %s", status)
	}

	if errCode, ok := raw["errorCode"]; ok {
		event.ErrorCode = asSMSString(errCode)
	}
	if errMessage, ok := raw["errorMessage"]; ok {
		event.ErrorMessage = asSMSString(errMessage)
	}

	return []WebhookEvent{event}, nil
}

// HealthCheck checks if SNS is healthy
func (p *SNSProvider) HealthCheck(ctx context.Context) error {
	form := url.Values{}
	form.Set("Action", "GetSMSAttributes")
	form.Set("Version", awsSNSAPIVersion)
	form.Set("attributes.member.1", "MonthlySpendLimit")
	_, err := p.doSignedRequest(ctx, form)
	return err
}

// SupportedRegions returns SNS's supported regions
func (p *SNSProvider) SupportedRegions() []string {
	if len(p.config.SupportedRegions) > 0 {
		return p.config.SupportedRegions
	}
	return []string{"US", "CA", "GB", "AU", "DE", "FR", "IN", "JP"}
}

// Close closes the provider
func (p *SNSProvider) Close() error {
	return nil
}

// Ensure SNSProvider implements Provider
var _ Provider = (*SNSProvider)(nil)

func resolveSNSCredentials(config ProviderConfig) (string, string) {
	accessKeyID := strings.TrimSpace(config.APIKey)
	secretAccessKey := strings.TrimSpace(config.APISecret)
	if secretAccessKey == "" {
		secretAccessKey = strings.TrimSpace(config.AuthToken)
	}
	return accessKeyID, secretAccessKey
}

func newAWSSMSSigner(region, accessKeyID, secretAccessKey string, httpClient *http.Client) (*credentials.Credentials, *v4.Signer, error) {
	cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient)
	switch {
	case accessKeyID != "" && secretAccessKey != "":
		cfg = cfg.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
	case accessKeyID != "" || secretAccessKey != "":
		return nil, nil, errors.Wrap(ErrInvalidConfig, "both access credentials are required for SNS")
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

func (p *SNSProvider) doSignedRequest(ctx context.Context, form url.Values) ([]byte, error) {
	payload := []byte(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to create SNS request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	if _, err := p.signer.Sign(req, bytes.NewReader(payload), "sns", p.config.Region, time.Now().UTC()); err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to sign SNS request: %v", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "SNS request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(ErrProviderError, "failed to read SNS response: %v", err)
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseAWSProviderError("SNS", resp.StatusCode, body)
	}

	return body, nil
}

func parseAWSProviderError(provider string, statusCode int, body []byte) error {
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

func asSMSString(value interface{}) string {
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

func firstNonEmptySMSString(values ...string) string {
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

// NewProvider creates a new SMS provider based on configuration
func NewProvider(providerType string, config ProviderConfig, logger zerolog.Logger) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case providerTwilio:
		return NewTwilioProvider(config, logger)
	case providerVonage, "nexmo":
		return NewGateway(providerType, config, logger)
	case "sns", "aws_sns":
		return NewSNSProvider(config, logger)
	case "mock":
		return NewMockProvider(logger), nil
	default:
		return nil, errors.Wrapf(ErrInvalidConfig, "unsupported SMS provider: %s", providerType)
	}
}
