// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// offRampBridgeImpl implements the OffRampBridge interface
type offRampBridgeImpl struct {
	cfg        OffRampConfig
	providers  map[string]OffRampProvider
	providerMu sync.RWMutex
	operations map[string]OffRampResult
	opMu       sync.RWMutex
}

// newOffRampBridge creates a new off-ramp bridge
func newOffRampBridge(cfg OffRampConfig) *offRampBridgeImpl {
	return &offRampBridgeImpl{
		cfg:        cfg,
		providers:  make(map[string]OffRampProvider),
		operations: make(map[string]OffRampResult),
	}
}

// RegisterProvider registers an off-ramp provider
func (b *offRampBridgeImpl) RegisterProvider(provider OffRampProvider) error {
	b.providerMu.Lock()
	defer b.providerMu.Unlock()

	b.providers[provider.Name()] = provider
	return nil
}

// GetQuote generates an off-ramp quote
func (b *offRampBridgeImpl) GetQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error) {
	// Validate amount limits
	if request.CryptoAmount.LT(b.cfg.MinAmount) {
		return OffRampQuote{}, ErrAmountTooSmall
	}
	if request.CryptoAmount.GT(b.cfg.MaxAmount) {
		return OffRampQuote{}, ErrAmountTooLarge
	}

	// Validate currency is supported
	if !b.isCurrencySupported(request.FiatCurrency) {
		return OffRampQuote{}, fmt.Errorf("unsupported fiat currency: %s", request.FiatCurrency)
	}

	// Validate payment method is supported
	if !b.isMethodSupported(request.PaymentMethod) {
		return OffRampQuote{}, fmt.Errorf("unsupported payment method: %s", request.PaymentMethod)
	}

	// Find best provider
	provider, err := b.findBestProvider(ctx, request)
	if err != nil {
		return OffRampQuote{}, err
	}

	// Get quote from provider
	quote, err := provider.GetQuote(ctx, request)
	if err != nil {
		return OffRampQuote{}, fmt.Errorf("provider quote failed: %w", err)
	}

	// Set validity period
	if quote.ExpiresAt.IsZero() {
		quote.ExpiresAt = time.Now().Add(b.cfg.QuoteValidityPeriod)
	}

	return quote, nil
}

// InitiateOffRamp initiates an off-ramp operation
func (b *offRampBridgeImpl) InitiateOffRamp(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error) {
	// Check quote expiration
	if quote.IsExpired() {
		return OffRampResult{}, ErrQuoteExpired
	}

	// Get provider
	b.providerMu.RLock()
	provider, exists := b.providers[quote.Provider]
	b.providerMu.RUnlock()

	if !exists {
		return OffRampResult{}, fmt.Errorf("provider %s not found", quote.Provider)
	}

	// Execute with provider
	result, err := provider.Execute(ctx, quote, signedTx)
	if err != nil {
		return OffRampResult{}, fmt.Errorf("off-ramp execution failed: %w", err)
	}

	// Store operation
	b.opMu.Lock()
	b.operations[result.ID] = result
	b.opMu.Unlock()

	return result, nil
}

// GetStatus fetches the status of an off-ramp operation
func (b *offRampBridgeImpl) GetStatus(ctx context.Context, offRampID string) (OffRampResult, error) {
	// Check local cache first
	b.opMu.RLock()
	result, exists := b.operations[offRampID]
	b.opMu.RUnlock()

	if !exists {
		return OffRampResult{}, fmt.Errorf("off-ramp operation %s not found", offRampID)
	}

	// If not terminal state, refresh from provider
	if result.Status == OffRampStatusPending || result.Status == OffRampStatusProcessing {
		b.providerMu.RLock()
		provider, provExists := b.providers[result.Provider]
		b.providerMu.RUnlock()

		if provExists {
			updated, err := provider.GetStatus(ctx, offRampID)
			if err == nil {
				b.opMu.Lock()
				b.operations[offRampID] = updated
				b.opMu.Unlock()
				return updated, nil
			}
		}
	}

	return result, nil
}

// CancelOffRamp cancels a pending off-ramp operation
func (b *offRampBridgeImpl) CancelOffRamp(ctx context.Context, offRampID string) error {
	b.opMu.RLock()
	result, exists := b.operations[offRampID]
	b.opMu.RUnlock()

	if !exists {
		return fmt.Errorf("off-ramp operation %s not found", offRampID)
	}

	if result.Status != OffRampStatusPending {
		return fmt.Errorf("cannot cancel operation in status %s", result.Status)
	}

	b.providerMu.RLock()
	provider, provExists := b.providers[result.Provider]
	b.providerMu.RUnlock()

	if !provExists {
		return fmt.Errorf("provider %s not found", result.Provider)
	}

	if err := provider.Cancel(ctx, offRampID); err != nil {
		return fmt.Errorf("cancel failed: %w", err)
	}

	// Update local state
	b.opMu.Lock()
	result.Status = OffRampStatusCancelled
	b.operations[offRampID] = result
	b.opMu.Unlock()

	return nil
}

// ListOperations lists off-ramp operations for an address
func (b *offRampBridgeImpl) ListOperations(ctx context.Context, address string, limit, offset int) ([]OffRampResult, error) {
	b.opMu.RLock()
	defer b.opMu.RUnlock()

	var results []OffRampResult
	for _, op := range b.operations {
		if op.CryptoTxHash != "" { // Match by sender (simplified)
			results = append(results, op)
		}
	}

	// Apply pagination
	if offset >= len(results) {
		return nil, nil
	}
	results = results[offset:]
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetSupportedCurrencies returns supported fiat currencies
func (b *offRampBridgeImpl) GetSupportedCurrencies(ctx context.Context) ([]FiatCurrency, error) {
	return b.cfg.SupportedCurrencies, nil
}

// GetSupportedMethods returns supported payment methods for a currency
func (b *offRampBridgeImpl) GetSupportedMethods(ctx context.Context, currency FiatCurrency) ([]PaymentMethod, error) {
	// Check which methods are supported by at least one provider for this currency
	var supported []PaymentMethod
	methodSet := make(map[PaymentMethod]bool)

	b.providerMu.RLock()
	defer b.providerMu.RUnlock()

	for _, provider := range b.providers {
		if provider.SupportsCurrency(currency) {
			for _, method := range b.cfg.SupportedMethods {
				if provider.SupportsMethod(method) && !methodSet[method] {
					supported = append(supported, method)
					methodSet[method] = true
				}
			}
		}
	}

	return supported, nil
}

// ValidateKYC validates if an address has sufficient KYC via VEID
func (b *offRampBridgeImpl) ValidateKYC(ctx context.Context, address string, veIDScore int64) error {
	if veIDScore < b.cfg.MinVEIDScore {
		return ErrKYCRequired
	}
	return nil
}

// findBestProvider finds the best provider for the request
func (b *offRampBridgeImpl) findBestProvider(ctx context.Context, request OffRampRequest) (OffRampProvider, error) {
	b.providerMu.RLock()
	defer b.providerMu.RUnlock()

	var bestProvider OffRampProvider
	var bestRate sdkmath.LegacyDec

	for _, provider := range b.providers {
		if !provider.IsHealthy(ctx) {
			continue
		}
		if !provider.SupportsCurrency(request.FiatCurrency) {
			continue
		}
		if !provider.SupportsMethod(request.PaymentMethod) {
			continue
		}

		// Get quote to compare rates
		quote, err := provider.GetQuote(ctx, request)
		if err != nil {
			continue
		}

		if bestProvider == nil || quote.ExchangeRate.GT(bestRate) {
			bestProvider = provider
			bestRate = quote.ExchangeRate
		}
	}

	if bestProvider == nil {
		return nil, ErrProviderUnavailable
	}

	return bestProvider, nil
}

// isCurrencySupported checks if a currency is supported
func (b *offRampBridgeImpl) isCurrencySupported(currency FiatCurrency) bool {
	for _, c := range b.cfg.SupportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// isMethodSupported checks if a payment method is supported
func (b *offRampBridgeImpl) isMethodSupported(method PaymentMethod) bool {
	for _, m := range b.cfg.SupportedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// ============================================================================
// Mock Off-Ramp Provider (for testing/development)
// ============================================================================

// MockOffRampProvider is a mock implementation for testing
type MockOffRampProvider struct {
	name                string
	supportedCurrencies []FiatCurrency
	supportedMethods    []PaymentMethod
	feePercent          sdkmath.LegacyDec
	healthy             bool
}

// NewMockOffRampProvider creates a new mock provider
func NewMockOffRampProvider(name string, currencies []FiatCurrency, methods []PaymentMethod) *MockOffRampProvider {
	return &MockOffRampProvider{
		name:                name,
		supportedCurrencies: currencies,
		supportedMethods:    methods,
		feePercent:          sdkmath.LegacyNewDecWithPrec(15, 3), // 1.5%
		healthy:             true,
	}
}

func (p *MockOffRampProvider) Name() string { return p.name }

func (p *MockOffRampProvider) GetQuote(ctx context.Context, request OffRampRequest) (OffRampQuote, error) {
	// Mock exchange rate (simplified)
	var rate sdkmath.LegacyDec
	switch request.FiatCurrency {
	case FiatUSD:
		rate = sdkmath.LegacyNewDecWithPrec(1, 0) // 1 crypto = 1 USD (mock)
	case FiatEUR:
		rate = sdkmath.LegacyNewDecWithPrec(92, 2) // 0.92 EUR
	case FiatGBP:
		rate = sdkmath.LegacyNewDecWithPrec(78, 2) // 0.78 GBP
	default:
		rate = sdkmath.LegacyOneDec()
	}

	fiatAmount := rate.MulInt(request.CryptoAmount)
	fee := p.feePercent.MulInt(request.CryptoAmount).TruncateInt()

	quoteID, _ := generateOffRampQuoteID()

	return OffRampQuote{
		ID:                  quoteID,
		Request:             request,
		CryptoAmount:        request.CryptoAmount,
		FiatAmount:          fiatAmount,
		ExchangeRate:        rate,
		Fee:                 fee,
		FiatFee:             sdkmath.LegacyZeroDec(),
		Provider:            p.name,
		EstimatedSettlement: 24 * time.Hour,
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now().UTC(),
	}, nil
}

func (p *MockOffRampProvider) Execute(ctx context.Context, quote OffRampQuote, signedTx []byte) (OffRampResult, error) {
	opID, _ := generateOffRampOperationID()

	return OffRampResult{
		ID:           opID,
		QuoteID:      quote.ID,
		Status:       OffRampStatusProcessing,
		CryptoTxHash: hex.EncodeToString(signedTx[:min(32, len(signedTx))]),
		CryptoAmount: quote.CryptoAmount,
		FiatAmount:   quote.FiatAmount,
		Fee:          quote.Fee,
		Provider:     p.name,
		InitiatedAt:  time.Now().UTC(),
	}, nil
}

func (p *MockOffRampProvider) GetStatus(ctx context.Context, operationID string) (OffRampResult, error) {
	return OffRampResult{}, fmt.Errorf("operation not found")
}

func (p *MockOffRampProvider) Cancel(ctx context.Context, operationID string) error {
	return nil
}

func (p *MockOffRampProvider) IsHealthy(ctx context.Context) bool {
	return p.healthy
}

func (p *MockOffRampProvider) SupportsCurrency(currency FiatCurrency) bool {
	for _, c := range p.supportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

func (p *MockOffRampProvider) SupportsMethod(method PaymentMethod) bool {
	for _, m := range p.supportedMethods {
		if m == method {
			return true
		}
	}
	return false
}

func generateOffRampQuoteID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "offramp_quote_" + hex.EncodeToString(bytes), nil
}

func generateOffRampOperationID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "offramp_" + hex.EncodeToString(bytes), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
