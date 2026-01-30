// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/virtengine/virtengine/pkg/payment"
)

// ============================================================================
// Off-Ramp Service Implementation
// ============================================================================

// offRampService is the main implementation of the Service interface.
type offRampService struct {
	config        Config
	providers     map[ProviderType]Provider
	defaultProvider Provider

	// Stores
	payoutStore       PayoutStore
	quoteStore        *QuoteStore
	reconcileStore    ReconciliationStore
	limitsStore       LimitsStore

	// KYC/AML
	kycGate     KYCGate
	amlScreener AMLScreener

	// Webhook handling
	webhookHandler WebhookHandler

	// Reconciliation
	reconciliation *ReconciliationService
	reconcileJob   *ReconciliationJob

	// Retry
	retryConfig RetryConfig

	// Metrics
	metrics *serviceMetrics

	// Close handling
	closeMu  sync.Mutex
	closed   bool
}

// serviceMetrics tracks service-level metrics.
type serviceMetrics struct {
	mu                  sync.RWMutex
	totalPayouts        int64
	successfulPayouts   int64
	failedPayouts       int64
	totalAmount         int64
	kycRejections       int64
	amlRejections       int64
	webhooksProcessed   int64
	reconciliationsRun  int64
}

// ServiceOption is a functional option for configuring the service.
type ServiceOption func(*offRampService)

// WithPayoutStore sets a custom payout store.
func WithPayoutStore(store PayoutStore) ServiceOption {
	return func(s *offRampService) {
		s.payoutStore = store
	}
}

// WithReconciliationStore sets a custom reconciliation store.
func WithReconciliationStore(store ReconciliationStore) ServiceOption {
	return func(s *offRampService) {
		s.reconcileStore = store
	}
}

// WithLimitsStore sets a custom limits store.
func WithLimitsStore(store LimitsStore) ServiceOption {
	return func(s *offRampService) {
		s.limitsStore = store
	}
}

// WithKYCGate sets a custom KYC gate.
func WithKYCGate(gate KYCGate) ServiceOption {
	return func(s *offRampService) {
		s.kycGate = gate
	}
}

// WithAMLScreener sets a custom AML screener.
func WithAMLScreener(screener AMLScreener) ServiceOption {
	return func(s *offRampService) {
		s.amlScreener = screener
	}
}

// WithProvider adds a provider to the service.
func WithProvider(provider Provider) ServiceOption {
	return func(s *offRampService) {
		s.providers[provider.Type()] = provider
		if s.defaultProvider == nil {
			s.defaultProvider = provider
		}
	}
}

// NewService creates a new off-ramp service.
func NewService(cfg Config, opts ...ServiceOption) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	svc := &offRampService{
		config:     cfg,
		providers:  make(map[ProviderType]Provider),
		quoteStore: NewQuoteStore(),
		retryConfig: cfg.RetryConfig,
		metrics:    &serviceMetrics{},
	}

	// Initialize providers
	if cfg.PayPalConfig.ClientID != "" {
		paypal, err := NewPayPalAdapter(cfg.PayPalConfig)
		if err == nil {
			svc.providers[ProviderPayPal] = paypal
		}
	}

	if cfg.ACHConfig.SecretKey != "" {
		ach, err := NewACHAdapter(cfg.ACHConfig)
		if err == nil {
			svc.providers[ProviderACH] = ach
		}
	}

	// Set default provider
	if provider, ok := svc.providers[cfg.DefaultProvider]; ok {
		svc.defaultProvider = provider
	} else if len(svc.providers) > 0 {
		for _, p := range svc.providers {
			svc.defaultProvider = p
			break
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(svc)
	}

	// Initialize default stores if not provided
	if svc.payoutStore == nil {
		svc.payoutStore = NewInMemoryPayoutStore()
	}
	if svc.reconcileStore == nil {
		svc.reconcileStore = NewInMemoryReconciliationStore()
	}
	if svc.limitsStore == nil {
		svc.limitsStore = NewInMemoryLimitsStore(cfg.LimitsConfig)
	}

	// Initialize default KYC gate if not provided
	if svc.kycGate == nil {
		svc.kycGate = NewVEIDKYCGate(cfg.KYCConfig, NewMockVEIDChecker())
	}

	// Initialize default AML screener if not provided
	if svc.amlScreener == nil {
		svc.amlScreener = NewDefaultAMLScreener(cfg.AMLConfig, NewMockAMLClient())
	}

	// Initialize webhook handler
	svc.webhookHandler = NewDefaultWebhookHandler(svc.payoutStore)

	// Initialize reconciliation service
	svc.reconciliation = NewReconciliationService(
		cfg.ReconciliationConfig,
		svc.payoutStore,
		svc.reconcileStore,
		svc.providers,
	)

	// Start reconciliation job if enabled
	if cfg.ReconciliationConfig.Enabled && cfg.ReconciliationConfig.IntervalMinutes > 0 {
		interval := time.Duration(cfg.ReconciliationConfig.IntervalMinutes) * time.Minute
		svc.reconcileJob = NewReconciliationJob(svc.reconciliation, interval)
		svc.reconcileJob.Start(context.Background())
	}

	return svc, nil
}

// ============================================================================
// Payout Operations
// ============================================================================

// CreatePayoutQuote creates a quote for a payout.
func (s *offRampService) CreatePayoutQuote(ctx context.Context, req CreatePayoutRequest) (*PayoutQuote, error) {
	// Validate request
	if req.AccountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if req.CryptoAmount <= 0 {
		return nil, fmt.Errorf("crypto amount must be positive")
	}
	if !req.FiatCurrency.IsValid() || !s.config.IsCurrencySupported(req.FiatCurrency) {
		return nil, payment.ErrInvalidCurrency
	}

	// Check KYC eligibility
	kycResult, err := s.kycGate.CheckKYCStatus(ctx, req.AccountAddress, req.VEIDID)
	if err != nil {
		return nil, fmt.Errorf("KYC check failed: %w", err)
	}

	if !kycResult.Status.IsVerified() {
		s.metrics.mu.Lock()
		s.metrics.kycRejections++
		s.metrics.mu.Unlock()
		return nil, ErrKYCNotVerified
	}

	// Get conversion rate (simplified - would use price feed in production)
	conversionRate := "0.10" // Example: 1 crypto = $0.10

	// Calculate fiat amount
	fiatValue := req.CryptoAmount * 10 // Simplified calculation

	// Calculate fee
	feeAmount := int64(float64(fiatValue) * s.config.ConversionConfig.FeePercent / 100)
	if minFee, ok := s.config.ConversionConfig.FeeMinimum[req.FiatCurrency]; ok && feeAmount < minFee {
		feeAmount = minFee
	}
	if maxFee, ok := s.config.ConversionConfig.FeeMaximum[req.FiatCurrency]; ok && feeAmount > maxFee {
		feeAmount = maxFee
	}

	netAmount := fiatValue - feeAmount

	// Validate amount
	if err := s.config.ValidatePayoutAmount(payment.Amount{Value: netAmount, Currency: req.FiatCurrency}); err != nil {
		return nil, err
	}

	// Check limits
	limits, err := s.limitsStore.GetLimits(ctx, req.AccountAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get limits: %w", err)
	}

	if ok, err := limits.CanPayout(netAmount); !ok {
		return nil, err
	}

	// Determine provider
	provider := req.Provider
	if provider == "" {
		provider = s.config.DefaultProvider
	}

	// Calculate estimated arrival
	var estimatedArrival time.Time
	switch provider {
	case ProviderPayPal:
		estimatedArrival = time.Now().Add(15 * time.Minute) // PayPal is fast
	case ProviderACH:
		estimatedArrival = time.Now().Add(time.Duration(s.config.ACHConfig.ProcessingDays) * 24 * time.Hour)
	default:
		estimatedArrival = time.Now().Add(24 * time.Hour)
	}

	// Create quote
	now := time.Now()
	quote := &PayoutQuote{
		QuoteID:          uuid.New().String(),
		CryptoAmount:     req.CryptoAmount,
		CryptoDenom:      req.CryptoDenom,
		FiatAmount:       payment.Amount{Value: fiatValue, Currency: req.FiatCurrency},
		ConversionRate:   conversionRate,
		Fee:              payment.Amount{Value: feeAmount, Currency: req.FiatCurrency},
		NetAmount:        payment.Amount{Value: netAmount, Currency: req.FiatCurrency},
		Provider:         provider,
		EstimatedArrival: estimatedArrival,
		ExpiresAt:        now.Add(time.Duration(s.config.QuoteValiditySeconds) * time.Second),
		CreatedAt:        now,
	}

	// Store quote
	if err := s.quoteStore.Save(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to save quote: %w", err)
	}

	return quote, nil
}

// ExecutePayout executes a payout from an approved quote.
func (s *offRampService) ExecutePayout(ctx context.Context, quoteID string) (*PayoutIntent, error) {
	// Get quote
	quote, err := s.quoteStore.Get(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("quote not found: %w", err)
	}

	if quote.IsExpired() {
		return nil, ErrPayoutFailed
	}

	// This is a simplified implementation
	// In production, you would:
	// 1. Lock the crypto amount on-chain
	// 2. Perform AML screening
	// 3. Submit to provider
	// 4. Update on-chain state

	return nil, fmt.Errorf("execute payout requires full request context - use CreatePayoutRequest")
}

// CreateAndExecutePayout creates and executes a payout in one operation.
func (s *offRampService) CreateAndExecutePayout(ctx context.Context, req CreatePayoutRequest) (*PayoutIntent, error) {
	// Create quote first
	quote, err := s.CreatePayoutQuote(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform AML screening
	amlResult, err := s.amlScreener.Screen(ctx, AMLScreenRequest{
		AccountAddress: req.AccountAddress,
		VEIDID:         req.VEIDID,
		FullName:       req.Metadata["full_name"],
		Country:        req.Metadata["country"],
		PayoutAmount:   quote.NetAmount.Value,
		PayoutCurrency: string(quote.FiatAmount.Currency),
	})
	if err != nil {
		return nil, fmt.Errorf("AML screening failed: %w", err)
	}

	if amlResult.Status == AMLStatusRejected {
		s.metrics.mu.Lock()
		s.metrics.amlRejections++
		s.metrics.mu.Unlock()
		return nil, ErrAMLCheckFailed
	}

	// Create payout intent
	now := time.Now()
	intent := &PayoutIntent{
		ID:             uuid.New().String(),
		Provider:       quote.Provider,
		Status:         PayoutStatusApproved,
		AccountAddress: req.AccountAddress,
		VEIDID:         req.VEIDID,
		CryptoAmount:   quote.CryptoAmount,
		CryptoDenom:    quote.CryptoDenom,
		FiatAmount:     quote.NetAmount,
		ConversionRate: quote.ConversionRate,
		Fee:            quote.Fee,
		Destination:    req.Destination,
		Description:    req.Description,
		Metadata:       req.Metadata,
		KYCStatus:      KYCStatusVerified,
		AMLStatus:      amlResult.Status,
		IdempotencyKey: req.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      quote.ExpiresAt,
	}

	if amlResult.ReviewRequired {
		intent.Status = PayoutStatusOnHold
		intent.AMLStatus = AMLStatusFlagged
	}

	intent.AddAuditEntry("payout_created", "service", fmt.Sprintf("quote_id=%s", quote.QuoteID))

	// Save intent
	if err := s.payoutStore.Save(ctx, intent); err != nil {
		return nil, fmt.Errorf("failed to save payout: %w", err)
	}

	// If approved, submit to provider
	if intent.Status == PayoutStatusApproved {
		if err := s.submitToProvider(ctx, intent); err != nil {
			intent.Status = PayoutStatusFailed
			intent.FailureCode = "PROVIDER_ERROR"
			intent.FailureMessage = err.Error()
			s.payoutStore.Save(ctx, intent)
			return intent, err
		}
	}

	// Update limits
	if err := s.limitsStore.UpdateUsage(ctx, req.AccountAddress, intent.FiatAmount.Value); err != nil {
		// Log but don't fail
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.totalPayouts++
	s.metrics.totalAmount += intent.FiatAmount.Value
	s.metrics.mu.Unlock()

	// Clean up quote
	s.quoteStore.Delete(ctx, quote.QuoteID)

	return intent, nil
}

// submitToProvider submits a payout to the provider with retry logic.
func (s *offRampService) submitToProvider(ctx context.Context, intent *PayoutIntent) error {
	provider, ok := s.providers[intent.Provider]
	if !ok {
		return ErrProviderNotConfigured
	}

	var lastErr error
	for attempt := 0; attempt < s.retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay
			delay := s.retryConfig.InitialDelay * time.Duration(1<<uint(attempt-1))
			if delay > s.retryConfig.MaxDelay {
				delay = s.retryConfig.MaxDelay
			}
			time.Sleep(delay)
		}

		intent.Status = PayoutStatusProcessing
		intent.AddAuditEntry("submit_attempt", "service", fmt.Sprintf("attempt=%d", attempt+1))
		s.payoutStore.Save(ctx, intent)

		err := provider.CreatePayout(ctx, intent)
		if err == nil {
			s.payoutStore.Save(ctx, intent)
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !s.isRetryableError(err) {
			break
		}
	}

	return lastErr
}

// isRetryableError checks if an error is retryable.
func (s *offRampService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, retryable := range s.retryConfig.RetryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// GetPayout retrieves a payout by ID.
func (s *offRampService) GetPayout(ctx context.Context, payoutID string) (*PayoutIntent, error) {
	return s.payoutStore.GetByID(ctx, payoutID)
}

// CancelPayout cancels a pending payout.
func (s *offRampService) CancelPayout(ctx context.Context, payoutID string, reason string) error {
	payout, err := s.payoutStore.GetByID(ctx, payoutID)
	if err != nil {
		return err
	}

	if !payout.Status.CanCancel() {
		return ErrPayoutAlreadyProcessed
	}

	// If already submitted to provider, try to cancel there
	if payout.ProviderPayoutID != "" {
		provider, ok := s.providers[payout.Provider]
		if ok {
			if err := provider.CancelPayout(ctx, payout.ProviderPayoutID); err != nil {
				// Log but continue - we'll mark as canceled locally
			}
		}
	}

	payout.Status = PayoutStatusCanceled
	payout.UpdatedAt = time.Now()
	payout.AddAuditEntry("payout_canceled", "user", reason)

	return s.payoutStore.Save(ctx, payout)
}

// ListPayouts lists payouts for an account.
func (s *offRampService) ListPayouts(ctx context.Context, accountAddress string, limit int) ([]*PayoutIntent, error) {
	return s.payoutStore.ListByAccount(ctx, accountAddress, limit)
}

// ============================================================================
// KYC/AML
// ============================================================================

// CheckPayoutEligibility checks if an account can make a payout.
func (s *offRampService) CheckPayoutEligibility(ctx context.Context, accountAddress string, amount int64) (*EligibilityResult, error) {
	result := &EligibilityResult{
		Eligible:        true,
		RequiredActions: make([]string, 0),
	}

	// Check KYC
	kycResult, err := s.kycGate.CheckKYCStatus(ctx, accountAddress, "")
	if err != nil {
		result.Eligible = false
		result.Reason = "Failed to check KYC status"
		return result, nil
	}

	result.KYCStatus = kycResult.Status
	if !kycResult.Status.IsVerified() {
		result.Eligible = false
		result.Reason = "KYC verification required"
		result.RequiredActions = append(result.RequiredActions, "Complete identity verification")
		return result, nil
	}

	// Check limits
	limits, err := s.limitsStore.GetLimits(ctx, accountAddress)
	if err != nil {
		result.Eligible = false
		result.Reason = "Failed to check limits"
		return result, nil
	}

	result.Limits = limits
	if ok, limitErr := limits.CanPayout(amount); !ok {
		result.Eligible = false
		result.Reason = limitErr.Error()
		return result, nil
	}

	// AML status would be checked during actual payout
	result.AMLStatus = AMLStatusPending

	return result, nil
}

// ============================================================================
// Limits
// ============================================================================

// GetPayoutLimits retrieves payout limits for an account.
func (s *offRampService) GetPayoutLimits(ctx context.Context, accountAddress string) (*PayoutLimits, error) {
	return s.limitsStore.GetLimits(ctx, accountAddress)
}

// ============================================================================
// Webhooks
// ============================================================================

// HandleWebhook handles an incoming webhook.
func (s *offRampService) HandleWebhook(ctx context.Context, providerType ProviderType, payload []byte, signature string) error {
	provider, ok := s.providers[providerType]
	if !ok {
		return ErrProviderNotConfigured
	}

	// Validate signature
	if s.config.WebhookConfig.SignatureVerification {
		if err := provider.ValidateWebhook(payload, signature); err != nil {
			return err
		}
	}

	// Parse event
	event, err := provider.ParseWebhookEvent(payload)
	if err != nil {
		return err
	}

	// Process event
	if err := s.webhookHandler.HandleEvent(ctx, event); err != nil {
		return err
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.webhooksProcessed++
	if event.Status == PayoutStatusSucceeded {
		s.metrics.successfulPayouts++
	} else if event.Status == PayoutStatusFailed {
		s.metrics.failedPayouts++
	}
	s.metrics.mu.Unlock()

	return nil
}

// ============================================================================
// Reconciliation
// ============================================================================

// RunReconciliation runs the reconciliation job.
func (s *offRampService) RunReconciliation(ctx context.Context) (*ReconciliationResult, error) {
	s.metrics.mu.Lock()
	s.metrics.reconciliationsRun++
	s.metrics.mu.Unlock()

	return s.reconciliation.Run(ctx)
}

// GetReconciliationRecord retrieves a reconciliation record.
func (s *offRampService) GetReconciliationRecord(ctx context.Context, payoutID string) (*ReconciliationRecord, error) {
	return s.reconcileStore.GetByPayoutID(ctx, payoutID)
}

// ============================================================================
// Health
// ============================================================================

// HealthCheck returns the health status.
func (s *offRampService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Providers: make(map[ProviderType]bool),
		Warnings:  make([]string, 0),
	}

	// Check providers
	for providerType, provider := range s.providers {
		healthy := provider.IsHealthy(ctx)
		status.Providers[providerType] = healthy
		if !healthy {
			status.Warnings = append(status.Warnings, fmt.Sprintf("Provider %s is unhealthy", providerType))
		}
	}

	// Check pending payouts
	pendingPayouts, err := s.payoutStore.ListByStatus(ctx, PayoutStatusProcessing, 100)
	if err == nil {
		status.PendingPayouts = len(pendingPayouts)
	}

	// Check reconciliation job
	if s.reconcileJob != nil {
		lastRun := s.reconcileJob.LastRun()
		if !lastRun.IsZero() {
			status.LastReconciliation = lastRun.Format(time.RFC3339)
		}
		if s.reconcileJob.LastError() != nil {
			status.Warnings = append(status.Warnings, "Last reconciliation had errors")
		}
	}

	// If no providers are healthy, mark service as unhealthy
	allUnhealthy := true
	for _, healthy := range status.Providers {
		if healthy {
			allUnhealthy = false
			break
		}
	}
	if allUnhealthy && len(status.Providers) > 0 {
		status.Healthy = false
		status.Status = "all providers unhealthy"
	}

	return status, nil
}

// Close closes the service.
func (s *offRampService) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	// Stop reconciliation job
	if s.reconcileJob != nil {
		s.reconcileJob.Stop()
	}

	// Close providers
	for _, provider := range s.providers {
		provider.Close()
	}

	return nil
}

// Ensure implementation satisfies interface
var _ Service = (*offRampService)(nil)
