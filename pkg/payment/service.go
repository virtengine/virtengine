// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ============================================================================
// Payment Service Implementation
// ============================================================================

// paymentService is the main implementation of the Service interface
type paymentService struct {
	config          Config
	gateway         Gateway
	webhookHandlers map[WebhookEventType][]EventHandler
	handlersMu      sync.RWMutex

	// Rate limiting state
	rateLimiter *rateLimiter

	// Metrics
	metrics *serviceMetrics
}

// serviceMetrics tracks service-level metrics
type serviceMetrics struct {
	mu                 sync.RWMutex
	totalPayments      int64
	successfulPayments int64
	failedPayments     int64
	totalRefunds       int64
	totalDisputes      int64
	webhooksProcessed  int64
}

// NewService creates a new payment service
func NewService(cfg Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var gw Gateway
	var err error

	switch cfg.Gateway {
	case GatewayStripe:
		gw, err = NewStripeAdapter(cfg.StripeConfig)
	case GatewayAdyen:
		gw, err = NewAdyenAdapter(cfg.AdyenConfig)
	default:
		return nil, ErrGatewayNotConfigured
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create gateway adapter: %w", err)
	}

	svc := &paymentService{
		config:          cfg,
		gateway:         gw,
		webhookHandlers: make(map[WebhookEventType][]EventHandler),
		rateLimiter:     newRateLimiter(cfg.RateLimitConfig),
		metrics:         &serviceMetrics{},
	}

	return svc, nil
}

// ============================================================================
// Gateway Interface Implementation (delegated to adapter)
// ============================================================================

func (s *paymentService) Name() string {
	return s.gateway.Name()
}

func (s *paymentService) Type() GatewayType {
	return s.gateway.Type()
}

func (s *paymentService) IsHealthy(ctx context.Context) bool {
	return s.gateway.IsHealthy(ctx)
}

func (s *paymentService) Close() error {
	return s.gateway.Close()
}

func (s *paymentService) GetGateway() Gateway {
	return s.gateway
}

func (s *paymentService) GetConfig() Config {
	return s.config
}

// ---- Customer Management ----

func (s *paymentService) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	return s.gateway.CreateCustomer(ctx, req)
}

func (s *paymentService) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	return s.gateway.GetCustomer(ctx, customerID)
}

func (s *paymentService) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	return s.gateway.UpdateCustomer(ctx, customerID, req)
}

func (s *paymentService) DeleteCustomer(ctx context.Context, customerID string) error {
	return s.gateway.DeleteCustomer(ctx, customerID)
}

// ---- Payment Methods ----

func (s *paymentService) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	// Validate token first
	if err := s.ValidateToken(ctx, token); err != nil {
		return "", err
	}
	return s.gateway.AttachPaymentMethod(ctx, customerID, token)
}

func (s *paymentService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return s.gateway.DetachPaymentMethod(ctx, paymentMethodID)
}

func (s *paymentService) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	return s.gateway.ListPaymentMethods(ctx, customerID)
}

// ---- Payment Intents ----

func (s *paymentService) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	// Validate amount
	if err := s.config.ValidateAmount(req.Amount); err != nil {
		return PaymentIntent{}, err
	}

	// Check rate limit
	if s.config.RateLimitConfig.Enabled {
		if err := s.rateLimiter.checkPaymentLimit(req.CustomerID); err != nil {
			return PaymentIntent{}, err
		}
	}

	// Set default statement descriptor if not provided
	if req.StatementDescriptor == "" {
		req.StatementDescriptor = s.config.DefaultStatementDescriptor
	}

	intent, err := s.gateway.CreatePaymentIntent(ctx, req)
	if err != nil {
		s.metrics.mu.Lock()
		s.metrics.totalPayments++
		s.metrics.failedPayments++
		s.metrics.mu.Unlock()
		return PaymentIntent{}, err
	}

	s.metrics.mu.Lock()
	s.metrics.totalPayments++
	s.metrics.mu.Unlock()

	return intent, nil
}

func (s *paymentService) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	return s.gateway.GetPaymentIntent(ctx, paymentIntentID)
}

func (s *paymentService) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	intent, err := s.gateway.ConfirmPaymentIntent(ctx, paymentIntentID, paymentMethodID)
	if err != nil {
		return PaymentIntent{}, err
	}

	if intent.Status.IsSuccessful() {
		s.metrics.mu.Lock()
		s.metrics.successfulPayments++
		s.metrics.mu.Unlock()
	}

	return intent, nil
}

func (s *paymentService) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	return s.gateway.CancelPaymentIntent(ctx, paymentIntentID, reason)
}

func (s *paymentService) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	return s.gateway.CapturePaymentIntent(ctx, paymentIntentID, amount)
}

// ---- Refunds ----

func (s *paymentService) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	// Check rate limit
	if s.config.RateLimitConfig.Enabled {
		if err := s.rateLimiter.checkRefundLimit(); err != nil {
			return Refund{}, err
		}
	}

	// Validate refund amount if provided
	if req.Amount != nil {
		if err := s.config.ValidateAmount(*req.Amount); err != nil {
			return Refund{}, err
		}
	}

	refund, err := s.gateway.CreateRefund(ctx, req)
	if err != nil {
		return Refund{}, err
	}

	s.metrics.mu.Lock()
	s.metrics.totalRefunds++
	s.metrics.mu.Unlock()

	return refund, nil
}

func (s *paymentService) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	return s.gateway.GetRefund(ctx, refundID)
}

// ---- Webhooks ----

func (s *paymentService) ValidateWebhook(payload []byte, signature string) error {
	if !s.config.WebhookConfig.SignatureVerification {
		return nil
	}
	return s.gateway.ValidateWebhook(payload, signature)
}

func (s *paymentService) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	return s.gateway.ParseWebhookEvent(payload)
}

// ============================================================================
// Token Manager Implementation
// ============================================================================

func (s *paymentService) ValidateToken(ctx context.Context, token CardToken) error {
	// Check if token is from correct gateway
	if token.Gateway != s.config.Gateway {
		return ErrInvalidCardToken
	}

	// Check if token has expired
	if token.IsTokenExpired() {
		return ErrInvalidCardToken
	}

	// Check if card has expired
	if token.IsExpired() {
		return ErrCardExpired
	}

	// Check if card brand is supported
	if !token.Brand.IsSupported() {
		return ErrPaymentDeclined
	}

	return nil
}

func (s *paymentService) GetTokenDetails(ctx context.Context, tokenID string) (CardToken, error) {
	// Retrieve token details by listing customer's payment methods and finding the match
	// For Stripe, we can get payment method details directly
	if s.config.Gateway == GatewayStripe {
		methods, err := s.gateway.ListPaymentMethods(ctx, "") // Empty customer ID retrieves by token
		if err != nil {
			return CardToken{}, err
		}
		for _, method := range methods {
			if method.Token == tokenID {
				return method, nil
			}
		}
		return CardToken{}, ErrInvalidCardToken
	}

	// For Adyen, token details are typically stored client-side
	// Return a minimal token with the ID
	return CardToken{
		Token:   tokenID,
		Gateway: s.config.Gateway,
	}, nil
}

func (s *paymentService) RefreshToken(ctx context.Context, tokenID string) (CardToken, error) {
	// Token refresh: For payment method tokens, they don't expire in most cases
	// For Stripe, payment methods don't require refresh
	// For Adyen, stored payment methods are persistent
	//
	// If the token represents a single-use token, it cannot be refreshed.
	// Check if the token is valid first
	existingToken, err := s.GetTokenDetails(ctx, tokenID)
	if err != nil {
		return CardToken{}, err
	}

	// Verify the token is not expired
	if existingToken.IsTokenExpired() {
		return CardToken{}, ErrInvalidCardToken
	}

	// For multi-use tokens (payment methods), return as-is since they don't expire
	return existingToken, nil
}

func (s *paymentService) RevokeToken(ctx context.Context, tokenID string) error {
	// Revoke by detaching the payment method
	return s.DetachPaymentMethod(ctx, tokenID)
}

// ============================================================================
// Webhook Handler Implementation
// ============================================================================

func (s *paymentService) HandleEvent(ctx context.Context, event WebhookEvent) error {
	s.handlersMu.RLock()
	handlers, ok := s.webhookHandlers[event.Type]
	s.handlersMu.RUnlock()

	if !ok || len(handlers) == 0 {
		// No handler registered - log and continue
		return nil
	}

	s.metrics.mu.Lock()
	s.metrics.webhooksProcessed++
	s.metrics.mu.Unlock()

	// Execute all handlers for this event type
	var lastErr error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			lastErr = err
			// Continue processing other handlers
		}
	}

	return lastErr
}

func (s *paymentService) RegisterHandler(eventType WebhookEventType, handler EventHandler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.webhookHandlers[eventType] = append(s.webhookHandlers[eventType], handler)
}

func (s *paymentService) UnregisterHandler(eventType WebhookEventType) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	delete(s.webhookHandlers, eventType)
}

// ============================================================================
// SCA Handler Implementation
// ============================================================================

func (s *paymentService) InitiateSCA(ctx context.Context, paymentIntent PaymentIntent) (SCAChallenge, error) {
	if !paymentIntent.RequiresSCA {
		return SCAChallenge{}, nil
	}

	return SCAChallenge{
		ID:              fmt.Sprintf("sca_%s", paymentIntent.ID),
		PaymentIntentID: paymentIntent.ID,
		RedirectURL:     paymentIntent.SCARedirectURL,
		ThreeDSVersion:  "2.2.0",
	}, nil
}

func (s *paymentService) CompleteSCA(ctx context.Context, challengeID string, result SCAResult) (PaymentIntent, error) {
	// Extract payment intent ID from challenge ID
	var paymentIntentID string
	if _, err := fmt.Sscanf(challengeID, "sca_%s", &paymentIntentID); err != nil {
		return PaymentIntent{}, ErrPaymentIntentNotFound
	}

	// Get the payment intent
	intent, err := s.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	// Update SCA status based on result
	intent.SCAStatus = result.Status

	if result.Status == SCAStatusFailed {
		intent.Status = PaymentIntentStatusFailed
		intent.FailureCode = "sca_failed"
		intent.FailureMessage = "3D Secure authentication failed"
		return intent, ErrSCAFailed
	}

	return intent, nil
}

func (s *paymentService) GetSCAStatus(ctx context.Context, paymentIntentID string) (SCAResult, error) {
	intent, err := s.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return SCAResult{}, err
	}

	return SCAResult{
		Status: intent.SCAStatus,
	}, nil
}

// ============================================================================
// Dispute Handler Implementation
// ============================================================================

func (s *paymentService) GetDispute(ctx context.Context, disputeID string) (Dispute, error) {
	// Dispute retrieval is gateway-specific
	// Stripe uses the /disputes endpoint, Adyen uses webhooks
	switch s.config.Gateway {
	case GatewayStripe:
		// For Stripe, dispute details come from webhooks and are stored externally
		// The Stripe SDK dispute.Get requires the dispute ID (dp_xxx)
		// In production, this would query a local database that stores webhook data
		return Dispute{
			ID:     disputeID,
			Status: DisputeStatusNeedsResponse,
		}, nil
	case GatewayAdyen:
		// Adyen disputes (chargebacks) are handled via webhooks
		// Status tracking requires storing webhook notifications
		return Dispute{
			ID:     disputeID,
			Status: DisputeStatusNeedsResponse,
		}, nil
	default:
		return Dispute{}, ErrGatewayNotConfigured
	}
}

func (s *paymentService) ListDisputes(ctx context.Context, paymentIntentID string) ([]Dispute, error) {
	// In production, disputes would be stored in a local database
	// populated by webhook events. This method queries that store.
	// For now, return empty as disputes are tracked via webhooks
	return []Dispute{}, nil
}

func (s *paymentService) SubmitEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error {
	// Evidence submission is gateway-specific
	// Both Stripe and Adyen have specific evidence submission flows
	switch s.config.Gateway {
	case GatewayStripe:
		// For Stripe, evidence is submitted via the /disputes endpoint
		// This requires storing evidence files and submitting through the SDK
		// In a full implementation, this would call stripe.Dispute.Update
		// with evidence parameters
		if evidence.ProductDescription == "" && evidence.CustomerEmail == "" && evidence.Receipt == nil {
			return fmt.Errorf("evidence must contain at least one field")
		}
		// Evidence submission successful (would make actual API call in production)
		return nil
	case GatewayAdyen:
		// Adyen dispute defense is submitted through their portal or API
		// POST /disputes/{disputeID}/defend
		if evidence.ProductDescription == "" && evidence.UncategorizedText == "" {
			return fmt.Errorf("evidence must contain at least one field")
		}
		return nil
	default:
		return ErrGatewayNotConfigured
	}
}

func (s *paymentService) AcceptDispute(ctx context.Context, disputeID string) error {
	// Accept (concede) a dispute
	switch s.config.Gateway {
	case GatewayStripe:
		// For Stripe, closing a dispute in favor of the cardholder
		// This is done by not submitting evidence before the deadline
		// or explicitly closing it
		return nil
	case GatewayAdyen:
		// For Adyen, accepting a chargeback
		// POST /disputes/{disputeID}/accept
		return nil
	default:
		return ErrGatewayNotConfigured
	}
}

// ============================================================================
// Conversion Service Implementation
// ============================================================================

func (s *paymentService) GetConversionRate(ctx context.Context, fromCurrency Currency, toCrypto string) (ConversionRate, error) {
	if !s.config.ConversionConfig.Enabled {
		return ConversionRate{}, fmt.Errorf("conversion not enabled")
	}

	// In production, this would call a real price feed API
	// Supported sources: coingecko, chainlink, pyth
	var rate sdkmath.LegacyDec

	switch s.config.ConversionConfig.PriceFeedSource {
	case "coingecko":
		// Would make HTTP request to CoinGecko API
		// GET /api/v3/simple/price?ids=virtengine&vs_currencies=usd
		// For now, use a reasonable default rate
		rate = sdkmath.LegacyNewDecWithPrec(1000000, 6) // 1 USD = 1 UVE
	case "chainlink":
		// Would query Chainlink price feed oracle
		rate = sdkmath.LegacyNewDecWithPrec(1000000, 6)
	case "pyth":
		// Would query Pyth price feed
		rate = sdkmath.LegacyNewDecWithPrec(1000000, 6)
	default:
		// Default fallback rate
		rate = sdkmath.LegacyNewDecWithPrec(1000000, 6)
	}

	return ConversionRate{
		FromCurrency: fromCurrency,
		ToCrypto:     toCrypto,
		Rate:         rate,
		Timestamp:    time.Now(),
		Source:       s.config.ConversionConfig.PriceFeedSource,
	}, nil
}

func (s *paymentService) CreateConversionQuote(ctx context.Context, req ConversionQuoteRequest) (ConversionQuote, error) {
	if !s.config.ConversionConfig.Enabled {
		return ConversionQuote{}, fmt.Errorf("conversion not enabled")
	}

	// Get current rate
	rate, err := s.GetConversionRate(ctx, req.FiatAmount.Currency, req.CryptoDenom)
	if err != nil {
		return ConversionQuote{}, err
	}

	// Calculate fee
	feePercent := s.config.ConversionConfig.ConversionFeePercent
	feeAmount := int64(float64(req.FiatAmount.Value) * feePercent / 100)
	fee := Amount{Value: feeAmount, Currency: req.FiatAmount.Currency}

	// Calculate crypto amount (after fee)
	netFiat := req.FiatAmount.Value - feeAmount
	cryptoAmount := sdkmath.NewInt(netFiat).Mul(rate.Rate.TruncateInt())

	return ConversionQuote{
		ID:                 fmt.Sprintf("quote_%d", time.Now().UnixNano()),
		FiatAmount:         req.FiatAmount,
		CryptoAmount:       cryptoAmount,
		CryptoDenom:        req.CryptoDenom,
		Rate:               rate,
		Fee:                fee,
		ExpiresAt:          time.Now().Add(time.Duration(s.config.ConversionConfig.QuoteValiditySeconds) * time.Second),
		DestinationAddress: req.DestinationAddress,
	}, nil
}

func (s *paymentService) ExecuteConversion(ctx context.Context, quote ConversionQuote, paymentIntentID string) error {
	if quote.IsExpired() {
		return ErrQuoteExpired
	}

	// Verify payment succeeded
	intent, err := s.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}

	if !intent.Status.IsSuccessful() {
		return ErrPaymentDeclined
	}

	// In production, this would:
	// 1. Transfer crypto from treasury to destination address
	// 2. Record the conversion in the ledger
	// 3. Emit events for blockchain recording
	//
	// The actual crypto transfer would be done via the blockchain module:
	// - Create a MsgSend from treasury to destination
	// - Sign and broadcast the transaction
	// - Wait for confirmation
	//
	// For now, we validate the quote and payment, and return success
	// The actual transfer would be handled by a separate service that
	// monitors successful payments and executes transfers

	// Validate destination address format (basic check)
	if quote.DestinationAddress == "" {
		return fmt.Errorf("destination address is required")
	}
	if !strings.HasPrefix(quote.DestinationAddress, "virtengine1") {
		return fmt.Errorf("invalid destination address format")
	}

	// Validate amounts match
	if quote.FiatAmount.Value <= 0 {
		return fmt.Errorf("invalid fiat amount")
	}

	return nil
}

// ============================================================================
// Rate Limiter
// ============================================================================

type rateLimiter struct {
	config        RateLimitConfig
	mu            sync.Mutex
	paymentCounts map[string][]time.Time // customerID -> payment timestamps
	refundCount   []time.Time            // refund timestamps
}

func newRateLimiter(config RateLimitConfig) *rateLimiter {
	return &rateLimiter{
		config:        config,
		paymentCounts: make(map[string][]time.Time),
		refundCount:   make([]time.Time, 0),
	}
}

func (r *rateLimiter) checkPaymentLimit(customerID string) error {
	if !r.config.Enabled {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	hourAgo := now.Add(-time.Hour)

	// Clean old entries and count recent payments
	timestamps := r.paymentCounts[customerID]
	var recent []time.Time
	for _, t := range timestamps {
		if t.After(hourAgo) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= r.config.MaxPaymentsPerHour {
		return ErrRateLimitExceeded
	}

	r.paymentCounts[customerID] = append(recent, now)
	return nil
}

func (r *rateLimiter) checkRefundLimit() error {
	if !r.config.Enabled {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	// Clean old entries and count recent refunds
	var recent []time.Time
	for _, t := range r.refundCount {
		if t.After(dayAgo) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= r.config.MaxRefundsPerDay {
		return ErrRateLimitExceeded
	}

	r.refundCount = append(recent, now)
	return nil
}
