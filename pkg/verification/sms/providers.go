// Package sms provides SMS provider abstraction for verification SMS.
package sms

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
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
	primary      Provider
	secondary    Provider
	logger       zerolog.Logger
	metrics      *ProviderMetrics
	mu           sync.RWMutex
	primaryFails int64
	lastPrimaryFail time.Time
}

// ProviderMetrics tracks provider metrics
type ProviderMetrics struct {
	mu               sync.RWMutex
	totalSent        int64
	primarySent      int64
	secondarySent    int64
	primaryFailures  int64
	secondaryFailures int64
	totalFailures    int64
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

// TwilioProvider implements Provider using Twilio
type TwilioProvider struct {
	config ProviderConfig
	logger zerolog.Logger
}

// NewTwilioProvider creates a new Twilio provider
func NewTwilioProvider(config ProviderConfig, logger zerolog.Logger) (*TwilioProvider, error) {
	if config.AccountSID == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "account_sid is required for Twilio")
	}
	if config.AuthToken == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "auth_token is required for Twilio")
	}
	if config.FromNumber == "" && config.MessagingServiceSID == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "from_number or messaging_service_sid is required for Twilio")
	}

	return &TwilioProvider{
		config: config,
		logger: logger.With().Str("provider", "twilio").Logger(),
	}, nil
}

// Name returns the provider name
func (p *TwilioProvider) Name() string {
	return "twilio"
}

// Send sends an SMS using Twilio
func (p *TwilioProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	// TODO: Implement actual Twilio API call using Twilio Go SDK
	p.logger.Info().
		Str("to", MaskPhoneNumber(msg.To)).
		Int("body_length", len(msg.Body)).
		Msg("Twilio send called (placeholder)")

	messageID := fmt.Sprintf("twilio-%d", time.Now().UnixNano())
	return &SendResult{
		Success:      true,
		MessageID:    messageID,
		Timestamp:    time.Now(),
		Provider:     "twilio",
		SegmentCount: (len(msg.Body) / 160) + 1,
	}, nil
}

// LookupCarrier performs carrier lookup using Twilio Lookup API
func (p *TwilioProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	// TODO: Implement actual Twilio Lookup API call
	p.logger.Info().
		Str("phone", MaskPhoneNumber(phoneNumber)).
		Msg("Twilio carrier lookup called (placeholder)")

	return &CarrierLookupResult{
		PhoneNumber:     phoneNumber,
		CountryCode:     "US",
		CarrierName:     "Unknown",
		CarrierType:     CarrierTypeMobile,
		IsVoIP:          false,
		IsMobile:        true,
		IsValid:         true,
		RiskScore:       20,
		LookupTimestamp: time.Now(),
	}, nil
}

// GetDeliveryStatus gets delivery status from Twilio
func (p *TwilioProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	// TODO: Implement actual Twilio API call
	return &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}, nil
}

// ParseWebhook parses a Twilio webhook
func (p *TwilioProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	// TODO: Implement Twilio webhook parsing and validation
	return []WebhookEvent{}, nil
}

// HealthCheck checks if Twilio is healthy
func (p *TwilioProvider) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual health check
	return nil
}

// SupportedRegions returns Twilio's supported regions
func (p *TwilioProvider) SupportedRegions() []string {
	if len(p.config.SupportedRegions) > 0 {
		return p.config.SupportedRegions
	}
	// Twilio supports most regions
	return []string{"US", "CA", "GB", "AU", "DE", "FR", "IN", "JP", "BR", "MX"}
}

// Close closes the provider
func (p *TwilioProvider) Close() error {
	return nil
}

// Ensure TwilioProvider implements Provider
var _ Provider = (*TwilioProvider)(nil)

// ============================================================================
// AWS SNS Provider (Placeholder)
// ============================================================================

// SNSProvider implements Provider using AWS SNS
type SNSProvider struct {
	config ProviderConfig
	logger zerolog.Logger
}

// NewSNSProvider creates a new AWS SNS provider
func NewSNSProvider(config ProviderConfig, logger zerolog.Logger) (*SNSProvider, error) {
	if config.Region == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "region is required for SNS")
	}

	return &SNSProvider{
		config: config,
		logger: logger.With().Str("provider", "sns").Logger(),
	}, nil
}

// Name returns the provider name
func (p *SNSProvider) Name() string {
	return "sns"
}

// Send sends an SMS using AWS SNS
func (p *SNSProvider) Send(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
	// TODO: Implement actual AWS SNS API call
	p.logger.Info().
		Str("to", MaskPhoneNumber(msg.To)).
		Int("body_length", len(msg.Body)).
		Msg("SNS send called (placeholder)")

	messageID := fmt.Sprintf("sns-%d", time.Now().UnixNano())
	return &SendResult{
		Success:      true,
		MessageID:    messageID,
		Timestamp:    time.Now(),
		Provider:     "sns",
		SegmentCount: 1,
	}, nil
}

// LookupCarrier is not supported by SNS
func (p *SNSProvider) LookupCarrier(ctx context.Context, phoneNumber string) (*CarrierLookupResult, error) {
	return nil, errors.Wrap(ErrCarrierLookupFailed, "carrier lookup not supported by SNS")
}

// GetDeliveryStatus gets delivery status from SNS
func (p *SNSProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*DeliveryStatusResult, error) {
	// TODO: Implement via CloudWatch logs
	return &DeliveryStatusResult{
		MessageID: messageID,
		Status:    DeliveryDelivered,
		Timestamp: time.Now(),
	}, nil
}

// ParseWebhook parses an SNS webhook (via Lambda/SQS)
func (p *SNSProvider) ParseWebhook(ctx context.Context, payload []byte, signature string) ([]WebhookEvent, error) {
	return []WebhookEvent{}, nil
}

// HealthCheck checks if SNS is healthy
func (p *SNSProvider) HealthCheck(ctx context.Context) error {
	return nil
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

// ============================================================================
// Provider Factory
// ============================================================================

// NewProvider creates a new SMS provider based on configuration
func NewProvider(providerType string, config ProviderConfig, logger zerolog.Logger) (Provider, error) {
	switch providerType {
	case "twilio":
		return NewTwilioProvider(config, logger)
	case "sns", "aws_sns":
		return NewSNSProvider(config, logger)
	case "mock":
		return NewMockProvider(logger), nil
	default:
		return nil, errors.Wrapf(ErrInvalidConfig, "unsupported SMS provider: %s", providerType)
	}
}

