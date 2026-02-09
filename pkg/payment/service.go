// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/virtengine/virtengine/pkg/pricefeed"
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

	// Dispute store for persistence
	disputeStore DisputeStore

	// Rate limiting state
	rateLimiter *rateLimiter

	// Metrics
	metrics *serviceMetrics

	// Price feed aggregator for real conversion rates
	priceFeed        pricefeed.Aggregator
	priceFeedMetrics *pricefeed.PrometheusMetrics
}

// serviceMetrics tracks service-level metrics
type serviceMetrics struct {
	mu                 sync.RWMutex
	totalPayments      int64
	successfulPayments int64
	failedPayments     int64
	totalRefunds       int64
	totalDisputes      int64
	disputesWon        int64
	disputesLost       int64
	evidenceSubmitted  int64
	disputesAccepted   int64
	webhooksProcessed  int64
}

// ServiceOption is a functional option for configuring the payment service
type ServiceOption func(*paymentService)

// WithDisputeStore sets a custom dispute store
func WithDisputeStore(store DisputeStore) ServiceOption {
	return func(s *paymentService) {
		s.disputeStore = store
	}
}

// NewService creates a new payment service
func NewService(cfg Config, opts ...ServiceOption) (Service, error) {
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
	case GatewayPayPal:
		gw, err = NewPayPalAdapter(cfg.PayPalConfig)
	case GatewayACH:
		gw, err = NewACHAdapter(cfg.ACHConfig)
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
		disputeStore:    NewInMemoryDisputeStore(), // Default to in-memory store
		rateLimiter:     newRateLimiter(cfg.RateLimitConfig),
		metrics:         &serviceMetrics{},
	}

	// Initialize price feed aggregator if conversion is enabled
	if cfg.ConversionConfig.Enabled {
		priceFeedCfg := buildPriceFeedConfig(cfg)
		agg, err := pricefeed.NewAggregator(priceFeedCfg)
		if err != nil {
			// Log warning but don't fail service creation
			// Conversion will fall back to error when no price feed available
		} else {
			if priceFeedCfg.EnableMetrics {
				metrics := pricefeed.NewPrometheusMetrics()
				agg.SetMetrics(metrics)
				svc.priceFeedMetrics = metrics
			}
			svc.priceFeed = agg
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(svc)
	}

	// Register default dispute webhook handlers
	svc.registerDisputeWebhookHandlers()

	return svc, nil
}

// registerDisputeWebhookHandlers sets up handlers for dispute-related webhooks
func (s *paymentService) registerDisputeWebhookHandlers() {
	// Handle dispute created
	s.RegisterHandler(WebhookEventChargeDisputeCreated, func(ctx context.Context, event WebhookEvent) error {
		return s.handleDisputeCreated(ctx, event)
	})

	// Handle dispute updated
	s.RegisterHandler(WebhookEventChargeDisputeUpdated, func(ctx context.Context, event WebhookEvent) error {
		return s.handleDisputeUpdated(ctx, event)
	})

	// Handle dispute closed
	s.RegisterHandler(WebhookEventChargeDisputeClosed, func(ctx context.Context, event WebhookEvent) error {
		return s.handleDisputeClosed(ctx, event)
	})

	// Handle funds withdrawn
	s.RegisterHandler(WebhookEventChargeDisputeFundsWithdrawn, func(ctx context.Context, event WebhookEvent) error {
		return s.handleDisputeFundsWithdrawn(ctx, event)
	})

	// Handle funds reinstated
	s.RegisterHandler(WebhookEventChargeDisputeFundsReinstated, func(ctx context.Context, event WebhookEvent) error {
		return s.handleDisputeFundsReinstated(ctx, event)
	})
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
	// Close price feed aggregator if initialized
	if s.priceFeed != nil {
		s.priceFeed.Close()
	}
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
// Dispute Handler Implementation - PAY-003
// ============================================================================

// GetDispute retrieves a dispute by ID, first checking local store then gateway
func (s *paymentService) GetDispute(ctx context.Context, disputeID string) (Dispute, error) {
	// First, check local store
	if record, found := s.disputeStore.Get(disputeID); found {
		return record.Dispute, nil
	}

	// If not in store, fetch from gateway
	adapter, err := resolveDisputeGateway(s.gateway)
	if err != nil {
		return Dispute{
			ID:      disputeID,
			Gateway: s.config.Gateway,
			Status:  DisputeStatusNeedsResponse,
		}, nil
	}

	disp, err := adapter.GetDispute(ctx, disputeID)

	if err != nil {
		return Dispute{}, err
	}

	// Store the fetched dispute
	record := NewDisputeRecord(disp, s.config.Gateway, "system:gateway_fetch")
	_ = s.disputeStore.Save(record)

	return disp, nil
}

// ListDisputes retrieves disputes for a payment intent
func (s *paymentService) ListDisputes(ctx context.Context, paymentIntentID string) ([]Dispute, error) {
	// First check local store
	records := s.disputeStore.GetByPaymentIntent(paymentIntentID)
	if len(records) > 0 {
		disputes := make([]Dispute, len(records))
		for i, record := range records {
			disputes[i] = record.Dispute
		}
		return disputes, nil
	}

	// If not in store, fetch from gateway
	adapter, err := resolveDisputeGateway(s.gateway)
	if err != nil {
		return []Dispute{}, nil
	}

	disputes, err := adapter.ListDisputes(ctx, paymentIntentID)

	if err != nil {
		return nil, err
	}

	// Store fetched disputes
	for _, disp := range disputes {
		record := NewDisputeRecord(disp, s.config.Gateway, "system:gateway_fetch")
		_ = s.disputeStore.Save(record)
	}

	return disputes, nil
}

// SubmitEvidence submits evidence for a dispute and updates the local store
func (s *paymentService) SubmitEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error {
	// Validate evidence has at least one field
	if evidence.ProductDescription == "" && evidence.CustomerEmail == "" &&
		evidence.Receipt == nil && evidence.UncategorizedText == "" {
		return fmt.Errorf("evidence must contain at least one field")
	}

	// Submit to gateway
	adapter, err := resolveDisputeGateway(s.gateway)
	if err != nil {
		return ErrGatewayNotConfigured
	}

	err = adapter.SubmitDisputeEvidence(ctx, disputeID, evidence)

	if err != nil {
		return err
	}

	// Update local store
	record, found := s.disputeStore.Get(disputeID)
	if found {
		evidenceType := "mixed"
		if evidence.ProductDescription != "" {
			evidenceType = "product_description"
		} else if evidence.Receipt != nil {
			evidenceType = "receipt"
		}
		record.MarkEvidenceSubmitted("api:submit_evidence", evidenceType)
		_ = s.disputeStore.Save(record)
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.evidenceSubmitted++
	s.metrics.mu.Unlock()

	return nil
}

// AcceptDispute accepts (concedes) a dispute
func (s *paymentService) AcceptDispute(ctx context.Context, disputeID string) error {
	// Submit to gateway
	adapter, err := resolveDisputeGateway(s.gateway)
	if err != nil {
		return ErrGatewayNotConfigured
	}

	err = adapter.AcceptDispute(ctx, disputeID)

	if err != nil {
		return err
	}

	// Update local store
	record, found := s.disputeStore.Get(disputeID)
	if found {
		record.MarkAccepted("api:accept_dispute", "merchant_accepted")
		_ = s.disputeStore.Save(record)
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.disputesAccepted++
	s.metrics.mu.Unlock()

	return nil
}

// GetDisputeStore returns the dispute store for direct access
func (s *paymentService) GetDisputeStore() DisputeStore {
	return s.disputeStore
}

// ResolveDispute executes a resolution workflow and persists the record.
func (s *paymentService) ResolveDispute(ctx context.Context, req DisputeResolutionRequest) (*DisputeResolutionResult, error) {
	return ResolveDispute(ctx, s, s.disputeStore, DefaultDisputeLifecycle(), req)
}

// GenerateDisputeReport generates a finance report from stored dispute data.
func (s *paymentService) GenerateDisputeReport(opts DisputeReportOptions) (*DisputeReport, error) {
	return GenerateDisputeReport(s.disputeStore, opts)
}

// ============================================================================
// Dispute Webhook Handlers - PAY-003
// ============================================================================

// handleDisputeCreated processes dispute created webhook events
//
//nolint:unparam // ctx kept for future async dispute processing
func (s *paymentService) handleDisputeCreated(_ context.Context, event WebhookEvent) error {
	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		return nil
	}

	// Create and store the dispute record
	record := NewDisputeRecord(*dispute, event.Gateway, "webhook:"+string(event.Type))
	if err := s.disputeStore.Save(record); err != nil {
		return err
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.totalDisputes++
	s.metrics.mu.Unlock()

	return nil
}

// handleDisputeUpdated processes dispute updated webhook events
//
//nolint:unparam // ctx kept for future async dispute processing
func (s *paymentService) handleDisputeUpdated(_ context.Context, event WebhookEvent) error {
	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		return nil
	}

	record, found := s.disputeStore.Get(dispute.ID)
	lifecycle := DefaultDisputeLifecycle()
	if !found {
		// Create new record if not found
		record = NewDisputeRecord(*dispute, event.Gateway, "webhook:"+string(event.Type))
	} else {
		// Update existing record
		transitionErr := lifecycle.Transition(record, dispute.Status, "webhook:"+string(event.Type), map[string]string{
			"event_id": event.ID,
		})
		if transitionErr != nil {
			record.AddAuditEntry(DisputeAuditEntry{
				Timestamp: time.Now(),
				Action:    "status_update_rejected",
				Actor:     "webhook:" + string(event.Type),
				Details: map[string]string{
					"event_id": event.ID,
					"error":    transitionErr.Error(),
				},
			})
			dispute.Status = record.Status
		}
		record.Dispute = *dispute
	}

	return s.disputeStore.Save(record)
}

// handleDisputeClosed processes dispute closed webhook events
//
//nolint:unparam // ctx kept for future async dispute processing
func (s *paymentService) handleDisputeClosed(_ context.Context, event WebhookEvent) error {
	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		return nil
	}

	record, found := s.disputeStore.Get(dispute.ID)
	lifecycle := DefaultDisputeLifecycle()
	if !found {
		record = NewDisputeRecord(*dispute, event.Gateway, "webhook:"+string(event.Type))
	} else {
		transitionErr := lifecycle.Transition(record, dispute.Status, "webhook:"+string(event.Type), map[string]string{
			"event_id": event.ID,
			"outcome":  string(dispute.Status),
		})
		if transitionErr != nil {
			record.AddAuditEntry(DisputeAuditEntry{
				Timestamp: time.Now(),
				Action:    "status_update_rejected",
				Actor:     "webhook:" + string(event.Type),
				Details: map[string]string{
					"event_id": event.ID,
					"error":    transitionErr.Error(),
				},
			})
			dispute.Status = record.Status
		}
		record.Dispute = *dispute
	}

	// Update metrics based on outcome
	s.metrics.mu.Lock()
	switch dispute.Status {
	case DisputeStatusWon:
		s.metrics.disputesWon++
	case DisputeStatusLost:
		s.metrics.disputesLost++
	}
	s.metrics.mu.Unlock()

	return s.disputeStore.Save(record)
}

// handleDisputeFundsWithdrawn processes funds withdrawn webhook events
//
//nolint:unparam // ctx kept for future async funds tracking
func (s *paymentService) handleDisputeFundsWithdrawn(_ context.Context, event WebhookEvent) error {
	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		return nil
	}

	record, found := s.disputeStore.Get(dispute.ID)
	if found {
		record.AddAuditEntry(DisputeAuditEntry{
			Timestamp: time.Now(),
			Action:    "funds_withdrawn",
			Actor:     "webhook:" + string(event.Type),
			Details: map[string]string{
				"event_id": event.ID,
				"amount":   fmt.Sprintf("%d", dispute.Amount.Value),
			},
		})
		return s.disputeStore.Save(record)
	}

	return nil
}

// handleDisputeFundsReinstated processes funds reinstated webhook events
//
//nolint:unparam // ctx kept for future async funds tracking
func (s *paymentService) handleDisputeFundsReinstated(_ context.Context, event WebhookEvent) error {
	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		return nil
	}

	record, found := s.disputeStore.Get(dispute.ID)
	if found {
		record.AddAuditEntry(DisputeAuditEntry{
			Timestamp: time.Now(),
			Action:    "funds_reinstated",
			Actor:     "webhook:" + string(event.Type),
			Details: map[string]string{
				"event_id": event.ID,
				"amount":   fmt.Sprintf("%d", dispute.Amount.Value),
			},
		})
		return s.disputeStore.Save(record)
	}

	return nil
}

// ============================================================================
// Conversion Service Implementation
// ============================================================================

func (s *paymentService) GetConversionRate(ctx context.Context, fromCurrency Currency, toCrypto string) (ConversionRate, error) {
	if !s.config.ConversionConfig.Enabled {
		return ConversionRate{}, fmt.Errorf("conversion not enabled")
	}

	// Use real price feed aggregator
	if s.priceFeed != nil {
		// Map currency to asset pair format
		// For fiat-to-crypto, we get crypto price in fiat then invert
		baseAsset := strings.ToLower(toCrypto)
		quoteAsset := strings.ToLower(string(fromCurrency))

		aggPrice, err := s.priceFeed.GetPrice(ctx, baseAsset, quoteAsset)
		if err == nil && aggPrice.IsValid() {
			// Price is in "crypto per fiat unit", so we need to invert
			// e.g., if 1 UVE = $0.50, then rate is 2 UVE per $1
			var rate sdkmath.LegacyDec
			if !aggPrice.Price.IsZero() {
				rate = sdkmath.LegacyOneDec().Quo(aggPrice.Price)
			} else {
				rate = sdkmath.LegacyZeroDec()
			}

			return ConversionRate{
				FromCurrency:      fromCurrency,
				ToCrypto:          toCrypto,
				Rate:              rate,
				Timestamp:         aggPrice.Timestamp,
				Source:            aggPrice.Source,
				Strategy:          string(aggPrice.Strategy),
				SourceAttribution: buildRateAttribution(aggPrice),
			}, nil
		}

		// Log the error but try fallback
		// In production, this would be logged with proper observability
	}

	// Fallback: Return error if no price feed available
	// This ensures we don't use stale/incorrect rates for financial transactions
	return ConversionRate{}, fmt.Errorf("price feed unavailable: unable to get real-time rate for %s/%s", toCrypto, fromCurrency)
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
	cryptoAmount := rate.Rate.MulInt(sdkmath.NewInt(netFiat)).TruncateInt()
	expiresAt := conversionQuoteExpiry(rate.Timestamp, s.config.ConversionConfig.QuoteValiditySeconds)
	quoteID := deterministicQuoteID(req, rate, fee, cryptoAmount, expiresAt)

	return ConversionQuote{
		ID:                 quoteID,
		FiatAmount:         req.FiatAmount,
		CryptoAmount:       cryptoAmount,
		CryptoDenom:        req.CryptoDenom,
		Rate:               rate,
		Fee:                fee,
		ExpiresAt:          expiresAt,
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

	recent = append(recent, now)
	r.refundCount = recent
	return nil
}
