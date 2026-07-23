package offramp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

const (
	defaultPartnerTimeout       = 10 * time.Second
	defaultPartnerResponseLimit = int64(1 << 20)
	defaultIdempotencyHeader    = "Idempotency-Key"
	defaultCorrelationHeader    = "X-Correlation-ID"
)

// PartnerEndpoints pins the exact API paths used by a contracted provider.
// The payout identifier placeholder is required in status and cancel paths.
type PartnerEndpoints struct {
	Health         string `json:"health"`
	Quote          string `json:"quote"`
	Payout         string `json:"payout"`
	Status         string `json:"status"`
	Cancel         string `json:"cancel"`
	MetadataLookup string `json:"metadata_lookup"`
}

// SecretResolver resolves secret-store references at request time. Returned
// bytes are never retained in adapter state.
type SecretResolver interface {
	ResolveSecret(ctx context.Context, ref SecretReference) ([]byte, error)
}

// HTTPPartnerConfig configures one exact provider API profile.
type HTTPPartnerConfig struct {
	Profile                     PayoutProfile
	BaseURL                     string
	AllowedPathPrefixes         []string
	Endpoints                   PartnerEndpoints
	Client                      *http.Client
	RequestTimeout              time.Duration
	MaxResponseBytes            int64
	Clock                       func() time.Time
	Secrets                     SecretResolver
	DailyLimits                 DailyLimitRepository
	ProfileAuthorizer           ProfileAuthorizer
	ComplianceAuthorizer        ComplianceAuthorizer
	AllowExternalBlockedSandbox bool
	AuthorizationHeader         string
	IdempotencyHeader           string
	CorrelationHeader           string
}

// HTTPPartnerAdapter is a strict generic adapter for a contracted, versioned
// JSON payout API. It stores only references and reconciliation metadata.
type HTTPPartnerAdapter struct {
	profile             PayoutProfile
	baseURL             *url.URL
	allowedPathPrefixes []string
	endpoints           PartnerEndpoints
	client              *http.Client
	requestTimeout      time.Duration
	maxResponseBytes    int64
	now                 func() time.Time
	secrets             SecretResolver
	dailyLimits         DailyLimitRepository
	profileAuthorizer   ProfileAuthorizer
	compliance          ComplianceAuthorizer
	authorizationHeader string
	idempotencyHeader   string
	correlationHeader   string

	mu             sync.RWMutex
	quotes         map[string]quoteBinding
	payouts        map[string]PayoutResult
	payoutBindings map[string]WebhookBinding
	lookupKeys     map[string]string
	idempotency    map[string]string
}

type quoteBinding struct {
	quote       Quote
	requestHash string
	corridor    PayoutCorridor
}

type partnerQuoteRequest struct {
	APIVersion           string `json:"api_version"`
	CryptoSymbol         string `json:"crypto_symbol"`
	CryptoDenom          string `json:"crypto_denom"`
	CryptoDecimals       uint8  `json:"crypto_decimals"`
	CryptoAmount         string `json:"crypto_amount"`
	FiatCurrency         string `json:"fiat_currency"`
	Rail                 string `json:"rail"`
	Jurisdiction         string `json:"jurisdiction"`
	BeneficiaryReference string `json:"beneficiary_reference"`
	CorrelationID        string `json:"correlation_id"`
	ComplianceReference  string `json:"compliance_reference"`
	KYCDecision          string `json:"kyc_decision"`
	SanctionsDecision    string `json:"sanctions_decision"`
}

type partnerQuoteResponse struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	APIVersion     string    `json:"api_version"`
	RequestBinding string    `json:"request_binding"`
	FiatAmount     string    `json:"fiat_amount"`
	ExchangeRate   string    `json:"exchange_rate"`
	Fee            string    `json:"fee"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type partnerPayoutRequest struct {
	APIVersion           string            `json:"api_version"`
	QuoteID              string            `json:"quote_id"`
	RequestBinding       string            `json:"request_binding"`
	CryptoTxReference    string            `json:"crypto_tx_reference"`
	CryptoAmount         string            `json:"crypto_amount"`
	FiatAmount           string            `json:"fiat_amount"`
	FiatCurrency         string            `json:"fiat_currency"`
	Rail                 string            `json:"rail"`
	Jurisdiction         string            `json:"jurisdiction"`
	BeneficiaryReference string            `json:"beneficiary_reference"`
	CorrelationID        string            `json:"correlation_id"`
	ComplianceReference  string            `json:"compliance_reference"`
	Metadata             map[string]string `json:"metadata"`
}

type partnerPayoutResponse struct {
	ID              string            `json:"id"`
	QuoteID         string            `json:"quote_id"`
	Provider        string            `json:"provider"`
	APIVersion      string            `json:"api_version"`
	CorrelationID   string            `json:"correlation_id"`
	Status          Status            `json:"status"`
	FiatAmount      string            `json:"fiat_amount"`
	CryptoAmount    string            `json:"crypto_amount"`
	Fee             string            `json:"fee"`
	Reference       string            `json:"reference"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	InitiatedAt     time.Time         `json:"initiated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
	StatusUpdatedAt time.Time         `json:"status_updated_at"`
	FailureReason   string            `json:"failure_reason,omitempty"`
	FailureCode     string            `json:"failure_code,omitempty"`
	Retryable       bool              `json:"retryable,omitempty"`
}

type partnerMetadataRequest struct {
	APIVersion string            `json:"api_version"`
	Metadata   map[string]string `json:"metadata"`
}

type partnerErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// NewHTTPPartnerAdapter validates and constructs a partner adapter. A
// production profile always requires HTTPS, including before certification is
// complete.
func NewHTTPPartnerAdapter(cfg HTTPPartnerConfig) (*HTTPPartnerAdapter, error) {
	if err := cfg.Profile.Validate(); err != nil {
		return nil, err
	}
	switch cfg.Profile.Environment {
	case EnvironmentProduction:
		if err := cfg.Profile.ValidateForExecution(ExecutionModeProduction, false); err != nil {
			return nil, err
		}
	case EnvironmentSandbox:
		if err := cfg.Profile.ValidateForExecution(ExecutionModeSandbox, cfg.AllowExternalBlockedSandbox); err != nil {
			return nil, err
		}
	default:
		return nil, ErrProfileNotExecutable
	}
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil || baseURL == nil {
		return nil, fmt.Errorf("%w: invalid partner base URL", ErrInvalidRequest)
	}
	cleanBasePath := path.Clean(baseURL.Path)
	if baseURL.Path == "" {
		cleanBasePath = ""
	}
	if !baseURL.IsAbs() || baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" ||
		baseURL.RawPath != "" || cleanBasePath != baseURL.Path || strings.Contains(baseURL.Path, "..") {
		return nil, fmt.Errorf("%w: invalid partner base URL", ErrInvalidRequest)
	}
	if cfg.Profile.Environment == EnvironmentProduction && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("%w: production partner URL must use HTTPS", ErrInvalidRequest)
	}
	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, fmt.Errorf("%w: unsupported partner URL scheme", ErrInvalidRequest)
	}
	if cfg.Profile.Environment == EnvironmentProduction && cfg.Secrets == nil {
		return nil, fmt.Errorf("%w: secret resolver is required", ErrProfileNotExecutable)
	}
	if cfg.Profile.Environment == EnvironmentProduction {
		if cfg.ProfileAuthorizer == nil {
			return nil, fmt.Errorf("%w: trusted profile authorizer is required", ErrProfileNotExecutable)
		}
		if err := cfg.ProfileAuthorizer.AuthorizePayoutProfile(cfg.Profile); err != nil {
			return nil, fmt.Errorf("%w: payout profile is not authorized", ErrProfileNotExecutable)
		}
		if cfg.ComplianceAuthorizer == nil {
			return nil, fmt.Errorf("%w: independent compliance authorizer is required", ErrProfileNotExecutable)
		}
	}
	dailyLimits := cfg.DailyLimits
	if dailyLimits == nil {
		dailyLimits = NewMemoryDailyLimitRepository()
	}
	if cfg.Profile.Environment == EnvironmentProduction && !dailyLimits.Durable() {
		return nil, fmt.Errorf("%w: durable daily-limit repository is required", ErrProfileNotExecutable)
	}
	if len(cfg.AllowedPathPrefixes) == 0 {
		return nil, fmt.Errorf("%w: path allowlist is required", ErrInvalidRequest)
	}
	for _, prefix := range cfg.AllowedPathPrefixes {
		if !validPartnerPath(prefix, false) {
			return nil, fmt.Errorf("%w: invalid path allowlist", ErrInvalidRequest)
		}
	}
	for _, endpoint := range []struct {
		value       string
		placeholder bool
	}{
		{cfg.Endpoints.Health, false}, {cfg.Endpoints.Quote, false}, {cfg.Endpoints.Payout, false},
		{cfg.Endpoints.Status, true}, {cfg.Endpoints.Cancel, true}, {cfg.Endpoints.MetadataLookup, false},
	} {
		if !validPartnerPath(endpoint.value, endpoint.placeholder) || !pathAllowed(endpoint.value, cfg.AllowedPathPrefixes) {
			return nil, fmt.Errorf("%w: endpoint is not allowlisted", ErrInvalidRequest)
		}
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = defaultPartnerTimeout
	}
	if timeout > time.Minute {
		return nil, fmt.Errorf("%w: request timeout exceeds safety bound", ErrInvalidRequest)
	}
	responseLimit := cfg.MaxResponseBytes
	if responseLimit <= 0 {
		responseLimit = defaultPartnerResponseLimit
	}
	if responseLimit > 8<<20 {
		return nil, fmt.Errorf("%w: response limit exceeds safety bound", ErrInvalidRequest)
	}
	now := cfg.Clock
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	client := &http.Client{}
	if cfg.Client != nil {
		*client = *cfg.Client
		switch transport := cfg.Client.Transport.(type) {
		case nil:
		case *http.Transport:
			client.Transport = transport.Clone()
			if cfg.Profile.Environment == EnvironmentProduction && (transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify || transport.Proxy != nil) {
				return nil, fmt.Errorf("%w: unsafe production transport", ErrInvalidRequest)
			}
		default:
			if cfg.Profile.Environment == EnvironmentProduction {
				return nil, fmt.Errorf("%w: opaque custom transport is not allowed in production", ErrInvalidRequest)
			}
		}
	} else if cfg.Profile.Environment == EnvironmentProduction {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = nil
		client.Transport = transport
	}
	if client.Timeout <= 0 || client.Timeout > time.Minute {
		client.Timeout = timeout
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	for _, header := range []string{cfg.AuthorizationHeader, cfg.IdempotencyHeader, cfg.CorrelationHeader} {
		if header != "" && !validHeaderName(header) {
			return nil, fmt.Errorf("%w: invalid partner header name", ErrInvalidRequest)
		}
	}
	adapter := &HTTPPartnerAdapter{
		profile:             clonePayoutProfile(cfg.Profile),
		baseURL:             baseURL,
		allowedPathPrefixes: slices.Clone(cfg.AllowedPathPrefixes),
		endpoints:           cfg.Endpoints,
		client:              client,
		requestTimeout:      timeout,
		maxResponseBytes:    responseLimit,
		now:                 now,
		secrets:             cfg.Secrets,
		dailyLimits:         dailyLimits,
		profileAuthorizer:   cfg.ProfileAuthorizer,
		compliance:          cfg.ComplianceAuthorizer,
		authorizationHeader: cfg.AuthorizationHeader,
		idempotencyHeader:   cfg.IdempotencyHeader,
		correlationHeader:   cfg.CorrelationHeader,
		quotes:              make(map[string]quoteBinding),
		payouts:             make(map[string]PayoutResult),
		payoutBindings:      make(map[string]WebhookBinding),
		lookupKeys:          make(map[string]string),
		idempotency:         make(map[string]string),
	}
	if adapter.authorizationHeader == "" {
		adapter.authorizationHeader = "Authorization"
	}
	if adapter.idempotencyHeader == "" {
		adapter.idempotencyHeader = defaultIdempotencyHeader
	}
	if adapter.correlationHeader == "" {
		adapter.correlationHeader = defaultCorrelationHeader
	}
	return adapter, nil
}

func (a *HTTPPartnerAdapter) Name() string { return a.profile.Provider }

func (*HTTPPartnerAdapter) productionPayoutAdapter() {}

func (a *HTTPPartnerAdapter) Profile() PayoutProfile { return clonePayoutProfile(a.profile) }

func (a *HTTPPartnerAdapter) NormalizeError(operation string, err error) error {
	return NormalizeError(a.Name(), operation, err)
}

func (a *HTTPPartnerAdapter) SupportsCurrency(currency string) bool {
	return slices.ContainsFunc(a.profile.Corridors, func(c PayoutCorridor) bool { return c.Currency == currency })
}

func (a *HTTPPartnerAdapter) SupportsMethod(method string) bool {
	return slices.ContainsFunc(a.profile.Corridors, func(c PayoutCorridor) bool { return c.Rail == method })
}

func (a *HTTPPartnerAdapter) IsHealthy(ctx context.Context) bool {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return false
	}
	mode := ExecutionModeSandbox
	allowExternalBlocked := true
	if a.profile.Environment == EnvironmentProduction {
		mode = ExecutionModeProduction
		allowExternalBlocked = false
	}
	if err := a.profile.ValidateForExecution(mode, allowExternalBlocked); err != nil {
		return false
	}
	var response struct {
		Status     string `json:"status"`
		Provider   string `json:"provider"`
		APIVersion string `json:"api_version"`
	}
	if err := a.doJSON(ctx, http.MethodGet, a.endpoints.Health, nil, "", "", &response); err != nil {
		return false
	}
	return response.Status == "ok" && response.Provider == a.profile.Provider && response.APIVersion == a.profile.APIVersion
}

func (a *HTTPPartnerAdapter) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return Quote{}, err
	}
	now := a.now().UTC()
	if err := validateQuoteRequest(req); err != nil {
		return Quote{}, err
	}
	if strings.TrimSpace(req.BeneficiaryReference) == "" || req.Destination != req.BeneficiaryReference || strings.TrimSpace(req.CorrelationID) == "" {
		return Quote{}, ErrInvalidRequest
	}
	if !externalIDPattern.MatchString(req.BeneficiaryReference) ||
		!strings.HasPrefix(req.BeneficiaryReference, a.profile.BeneficiaryRequirements.ReferencePrefix) ||
		!externalIDPattern.MatchString(req.CorrelationID) {
		return Quote{}, ErrInvalidRequest
	}
	corridor, err := a.profile.Corridor(req.Jurisdiction, req.FiatCurrency, req.PaymentMethod)
	if err != nil {
		return Quote{}, err
	}
	if err := validateCompliance(req.Compliance, a.profile.DecisionRequirements, now); err != nil {
		return Quote{}, err
	}
	if a.compliance != nil {
		if err := a.compliance.AuthorizePayout(ctx, req.Compliance, req.Sender, req.BeneficiaryReference, corridor.ID, now); err != nil {
			return Quote{}, ErrComplianceRequired
		}
	}
	wireRequest := partnerQuoteRequest{
		APIVersion:           a.profile.APIVersion,
		CryptoSymbol:         req.CryptoSymbol,
		CryptoDenom:          req.CryptoDenom,
		CryptoDecimals:       req.CryptoDecimals,
		CryptoAmount:         req.CryptoAmount.String(),
		FiatCurrency:         req.FiatCurrency,
		Rail:                 req.PaymentMethod,
		Jurisdiction:         req.Jurisdiction,
		BeneficiaryReference: req.BeneficiaryReference,
		CorrelationID:        req.CorrelationID,
		ComplianceReference:  req.Compliance.Reference,
		KYCDecision:          req.Compliance.KYCDecision,
		SanctionsDecision:    req.Compliance.SanctionsDecision,
	}
	requestHash, err := hashStrictJSON(wireRequest)
	if err != nil {
		return Quote{}, err
	}
	var response partnerQuoteResponse
	if err := a.doJSON(ctx, http.MethodPost, a.endpoints.Quote, wireRequest, "", req.CorrelationID, &response); err != nil {
		return Quote{}, err
	}
	if response.Provider != a.profile.Provider || response.APIVersion != a.profile.APIVersion || response.RequestBinding != requestHash || strings.TrimSpace(response.ID) == "" {
		return Quote{}, fmt.Errorf("%w: quote binding mismatch", ErrProviderRejected)
	}
	fiatAmount, err := parsePositiveDecimal(response.FiatAmount)
	if err != nil {
		return Quote{}, err
	}
	if err := validateAmount(corridor, fiatAmount); err != nil {
		return Quote{}, err
	}
	exchangeRate, err := parsePositiveDecimal(response.ExchangeRate)
	if err != nil {
		return Quote{}, err
	}
	fee, err := parseNonNegativeInt(response.Fee)
	if err != nil {
		return Quote{}, err
	}
	if response.CreatedAt.IsZero() || response.ExpiresAt.IsZero() || response.CreatedAt.After(now) || response.CreatedAt.Before(now.Add(-corridor.QuoteTTL)) ||
		!response.ExpiresAt.After(now) || response.ExpiresAt.After(response.CreatedAt.Add(corridor.QuoteTTL)) {
		return Quote{}, ErrQuoteExpired
	}
	divisor := sdkmath.LegacyNewDec(10).Power(uint64(req.CryptoDecimals))
	expectedFiat := exchangeRate.MulInt(req.CryptoAmount).Quo(divisor)
	if !expectedFiat.Equal(fiatAmount) {
		return Quote{}, fmt.Errorf("%w: quote amount does not match exchange rate", ErrProviderRejected)
	}
	quote := Quote{
		ID: response.ID, Request: req, FiatAmount: fiatAmount, ExchangeRate: exchangeRate, Fee: fee,
		Provider: response.Provider, CreatedAt: response.CreatedAt.UTC(), ExpiresAt: response.ExpiresAt.UTC(),
		AuditFields: map[string]string{
			"profile_id": a.profile.ID, "profile_state": string(a.profile.State), "api_version": a.profile.APIVersion,
			"corridor_id": corridor.ID, "request_binding": requestHash, "execution_environment": string(a.profile.Environment),
			"production_floor_eligible": strconv.FormatBool(a.profile.State == ProfileCertifiedEnabled && a.profile.Environment == EnvironmentProduction),
		},
	}
	a.mu.Lock()
	if existing, exists := a.quotes[quote.ID]; exists && (!quotesEqual(existing.quote, quote) || existing.requestHash != requestHash || existing.corridor.ID != corridor.ID) {
		a.mu.Unlock()
		return Quote{}, fmt.Errorf("%w: provider reused quote ID for different terms", ErrProviderRejected)
	}
	a.quotes[quote.ID] = quoteBinding{quote: cloneQuote(quote), requestHash: requestHash, corridor: corridor}
	a.mu.Unlock()
	return quote, nil
}

func (a *HTTPPartnerAdapter) InitiatePayout(ctx context.Context, req PayoutRequest) (PayoutResult, error) {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return PayoutResult{}, err
	}
	now := a.now().UTC()
	if err := validatePayoutInputs(req.Quote, req.CryptoTxRef, req.Destination, now); err != nil {
		return PayoutResult{}, err
	}
	idempotencyKey, correlationID, err := requiredOperationKeys(req.Quote, req.Metadata)
	if err != nil {
		return PayoutResult{}, err
	}
	if err := validateOperationalMetadata(req.Metadata); err != nil {
		return PayoutResult{}, err
	}
	operationBinding, err := payoutOperationBinding(req, idempotencyKey, correlationID)
	if err != nil {
		return PayoutResult{}, err
	}
	a.mu.RLock()
	if existingBinding, exists := a.idempotency[idempotencyKey]; exists && existingBinding != operationBinding {
		a.mu.RUnlock()
		return PayoutResult{}, fmt.Errorf("%w: idempotency key reused for different payout", ErrInvalidRequest)
	}
	if payoutID, ok := a.lookupKeys[metadataLookupKey(a.profile.Provider, req.Metadata)]; ok {
		result := clonePayoutResult(a.payouts[payoutID])
		a.mu.RUnlock()
		return result, nil
	}
	binding, ok := a.quotes[req.Quote.ID]
	a.mu.RUnlock()
	if !ok || !quotesEqual(binding.quote, req.Quote) || req.Destination != req.Quote.Request.BeneficiaryReference {
		return PayoutResult{}, fmt.Errorf("%w: quote is unknown or modified", ErrInvalidRequest)
	}
	if binding.quote.IsExpired(now) {
		return PayoutResult{}, ErrQuoteExpired
	}
	if err := validateCompliance(req.Quote.Request.Compliance, a.profile.DecisionRequirements, now); err != nil {
		return PayoutResult{}, err
	}
	if a.compliance != nil {
		if err := a.compliance.AuthorizePayout(ctx, req.Quote.Request.Compliance, req.Quote.Request.Sender, req.Destination, binding.corridor.ID, now); err != nil {
			return PayoutResult{}, ErrComplianceRequired
		}
	}
	if err := validateAmount(binding.corridor, req.Quote.FiatAmount); err != nil {
		return PayoutResult{}, err
	}
	reservationKey := now.Format("2006-01-02") + "|" + binding.corridor.ID
	reserved, err := a.dailyLimits.ReserveDailyAmount(ctx, reservationKey, idempotencyKey, req.Quote.FiatAmount, binding.corridor.DailyLimit)
	if err != nil {
		return PayoutResult{}, fmt.Errorf("%w: daily-limit repository unavailable", ErrAdapterUnavailable)
	}
	if !reserved {
		return PayoutResult{}, ErrLimitExceeded
	}
	a.mu.Lock()
	if existingBinding, exists := a.idempotency[idempotencyKey]; exists && existingBinding != operationBinding {
		a.mu.Unlock()
		_ = a.dailyLimits.ReleaseDailyAmount(context.Background(), reservationKey, idempotencyKey)
		return PayoutResult{}, fmt.Errorf("%w: idempotency key reused for different payout", ErrInvalidRequest)
	}
	a.idempotency[idempotencyKey] = operationBinding
	a.mu.Unlock()
	releaseReservation := true
	defer func() {
		if releaseReservation {
			if releaseErr := a.dailyLimits.ReleaseDailyAmount(context.Background(), reservationKey, idempotencyKey); releaseErr == nil {
				a.mu.Lock()
				delete(a.idempotency, idempotencyKey)
				a.mu.Unlock()
			}
		}
	}()
	wireRequest := partnerPayoutRequest{
		APIVersion: a.profile.APIVersion, QuoteID: req.Quote.ID, RequestBinding: binding.requestHash,
		CryptoTxReference: req.CryptoTxRef, CryptoAmount: req.Quote.Request.CryptoAmount.String(),
		FiatAmount: req.Quote.FiatAmount.String(), FiatCurrency: req.Quote.Request.FiatCurrency,
		Rail: req.Quote.Request.PaymentMethod, Jurisdiction: req.Quote.Request.Jurisdiction,
		BeneficiaryReference: req.Destination, CorrelationID: correlationID,
		ComplianceReference: req.Quote.Request.Compliance.Reference, Metadata: cloneStringMap(req.Metadata),
	}
	var response partnerPayoutResponse
	err = a.doJSON(ctx, http.MethodPost, a.endpoints.Payout, wireRequest, idempotencyKey, correlationID, &response)
	if err != nil {
		if IsAmbiguous(err) {
			releaseReservation = false
		}
		return PayoutResult{}, err
	}
	result, err := a.convertPayoutResponse(response, &binding, correlationID, req.Metadata)
	if err != nil {
		return PayoutResult{}, err
	}
	a.storePayout(result, WebhookBinding{
		Provider: a.profile.Provider, PayoutID: result.ID, QuoteID: req.Quote.ID, CorrelationID: correlationID,
		ReservationDay: now.Format("2006-01-02"),
	})
	if result.Status == StatusFailed || result.Status == StatusCancelled {
		if err := a.releasePayoutReservation(ctx, result); err != nil {
			return PayoutResult{}, err
		}
	}
	releaseReservation = false
	return result, nil
}

func (a *HTTPPartnerAdapter) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return PayoutResult{}, err
	}
	if strings.TrimSpace(payoutID) == "" {
		return PayoutResult{}, ErrInvalidRequest
	}
	a.mu.RLock()
	binding, known := a.payoutBindings[payoutID]
	a.mu.RUnlock()
	if !known {
		return PayoutResult{}, ErrPayoutNotFound
	}
	var response partnerPayoutResponse
	if err := a.doJSON(ctx, http.MethodGet, expandPayoutPath(a.endpoints.Status, payoutID), nil, "", binding.CorrelationID, &response); err != nil {
		return PayoutResult{}, err
	}
	result, err := a.convertPayoutResponse(response, nil, binding.CorrelationID, nil)
	if err != nil || result.ID != payoutID || result.QuoteID != binding.QuoteID {
		return PayoutResult{}, fmt.Errorf("%w: status reconciliation mismatch", ErrProviderRejected)
	}
	a.mu.RLock()
	current := a.payouts[payoutID]
	a.mu.RUnlock()
	if err := reconcilePayoutStatus(current, result); err != nil {
		return PayoutResult{}, err
	}
	a.storePayout(result, binding)
	if result.Status == StatusFailed || result.Status == StatusCancelled {
		a.mu.RLock()
		_, locallyQuoted := a.quotes[result.QuoteID]
		a.mu.RUnlock()
		if locallyQuoted {
			if err := a.releasePayoutReservation(ctx, result); err != nil {
				return PayoutResult{}, err
			}
		}
	}
	return result, nil
}

func (a *HTTPPartnerAdapter) Cancel(ctx context.Context, payoutID string) error {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return err
	}
	if strings.TrimSpace(payoutID) == "" {
		return ErrInvalidRequest
	}
	a.mu.RLock()
	binding, known := a.payoutBindings[payoutID]
	current := a.payouts[payoutID]
	a.mu.RUnlock()
	if !known {
		return ErrPayoutNotFound
	}
	if current.Status == StatusCancelled {
		return nil
	}
	if current.Status.IsTerminal() {
		return ErrPayoutNotCancellable
	}
	var response partnerPayoutResponse
	if err := a.doJSON(ctx, http.MethodPost, expandPayoutPath(a.endpoints.Cancel, payoutID), struct {
		APIVersion string `json:"api_version"`
	}{APIVersion: a.profile.APIVersion}, "cancel-"+payoutID, binding.CorrelationID, &response); err != nil {
		return err
	}
	result, err := a.convertPayoutResponse(response, nil, binding.CorrelationID, current.Metadata)
	if err != nil || result.ID != payoutID || result.Status != StatusCancelled {
		return ErrPayoutNotCancellable
	}
	if err := reconcilePayoutStatus(current, result); err != nil {
		return err
	}
	a.storePayout(result, binding)
	return a.releasePayoutReservation(ctx, result)
}

func (a *HTTPPartnerAdapter) FindPayoutByMetadata(ctx context.Context, metadata map[string]string) (PayoutResult, error) {
	if err := a.authorizeRuntimeProfile(); err != nil {
		return PayoutResult{}, err
	}
	if len(metadata) == 0 {
		return PayoutResult{}, ErrInvalidRequest
	}
	if err := validateOperationalMetadata(metadata); err != nil {
		return PayoutResult{}, err
	}
	key := metadataLookupKey(a.profile.Provider, metadata)
	a.mu.RLock()
	if payoutID, ok := a.lookupKeys[key]; ok {
		result := clonePayoutResult(a.payouts[payoutID])
		a.mu.RUnlock()
		return result, nil
	}
	a.mu.RUnlock()
	correlationID := metadata["correlation_id"]
	if correlationID == "" {
		return PayoutResult{}, ErrInvalidRequest
	}
	var response partnerPayoutResponse
	err := a.doJSON(ctx, http.MethodPost, a.endpoints.MetadataLookup, partnerMetadataRequest{
		APIVersion: a.profile.APIVersion, Metadata: cloneStringMap(metadata),
	}, "lookup-"+metadata["idempotency_key"], correlationID, &response)
	if err != nil {
		return PayoutResult{}, err
	}
	result, err := a.convertPayoutResponse(response, nil, correlationID, metadata)
	if err != nil {
		return PayoutResult{}, err
	}
	reservationDay := result.InitiatedAt.UTC().Format("2006-01-02")
	a.mu.RLock()
	if existing, ok := a.payoutBindings[result.ID]; ok && existing.ReservationDay != "" {
		reservationDay = existing.ReservationDay
	}
	a.mu.RUnlock()
	binding := WebhookBinding{Provider: a.profile.Provider, PayoutID: result.ID, QuoteID: result.QuoteID, CorrelationID: correlationID, ReservationDay: reservationDay}
	a.storePayout(result, binding)
	return result, nil
}

// LookupWebhookBinding implements WebhookBindingRepository.
func (a *HTTPPartnerAdapter) LookupWebhookBinding(_ context.Context, provider string, payoutID string) (WebhookBinding, error) {
	if provider != a.profile.Provider {
		return WebhookBinding{}, ErrPayoutNotFound
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	binding, ok := a.payoutBindings[payoutID]
	if !ok {
		return WebhookBinding{}, ErrPayoutNotFound
	}
	return binding, nil
}

// Durable is false because this adapter's internal binding cache is an
// engineering convenience. Production webhook verification requires a
// separately persisted WebhookBindingRepository.
func (*HTTPPartnerAdapter) Durable() bool { return false }

func (a *HTTPPartnerAdapter) authorizeRuntimeProfile() error {
	mode := ExecutionModeSandbox
	allowExternalBlocked := true
	if a.profile.Environment == EnvironmentProduction {
		mode = ExecutionModeProduction
		allowExternalBlocked = false
	}
	if err := a.profile.ValidateForExecution(mode, allowExternalBlocked); err != nil {
		return err
	}
	if mode == ExecutionModeProduction && (a.profileAuthorizer == nil || a.profileAuthorizer.AuthorizePayoutProfile(a.profile) != nil) {
		return ErrProfileNotExecutable
	}
	return nil
}

func (a *HTTPPartnerAdapter) doJSON(ctx context.Context, method string, endpoint string, requestBody any, idempotencyKey string, correlationID string, responseBody any) error {
	endpointURL, err := a.endpointURL(endpoint)
	if err != nil {
		return err
	}
	var body io.Reader
	if requestBody != nil {
		encoded, encodeErr := json.Marshal(requestBody)
		if encodeErr != nil {
			return fmt.Errorf("%w: encode partner request", ErrInvalidRequest)
		}
		body = bytes.NewReader(encoded)
	}
	requestCtx, cancel := context.WithTimeout(ctx, a.requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, method, endpointURL.String(), body)
	if err != nil {
		return fmt.Errorf("%w: create partner request", ErrInvalidRequest)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Version", a.profile.APIVersion)
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set(a.idempotencyHeader, idempotencyKey)
	}
	if correlationID != "" {
		req.Header.Set(a.correlationHeader, correlationID)
	}
	if a.secrets != nil {
		credentialRef, refErr := findCredentialRef(a.profile, "api", a.profile.Environment)
		if refErr != nil {
			return fmt.Errorf("%w: API credential reference unavailable", ErrProfileNotExecutable)
		}
		secret, resolveErr := a.secrets.ResolveSecret(requestCtx, credentialRef)
		if resolveErr != nil || len(secret) == 0 {
			return fmt.Errorf("%w: API credential unavailable", ErrAdapterUnavailable)
		}
		req.Header.Set(a.authorizationHeader, "Bearer "+string(secret))
		for i := range secret {
			secret[i] = 0
		}
	}
	response, err := a.client.Do(req)
	if err != nil {
		operation := operationForEndpoint(endpoint, a.endpoints)
		if method == http.MethodPost && (operation == operationInitiatePayout || operation == operationCancel) {
			return &ProviderError{Provider: a.Name(), Operation: operation, Kind: ErrorKindAmbiguous, Retryable: true, Ambiguous: true, err: ErrProviderAmbiguous}
		}
		return NormalizeError(a.Name(), operation, err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, a.maxResponseBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return NormalizeError(a.Name(), operationForEndpoint(endpoint, a.endpoints), err)
	}
	if int64(len(raw)) > a.maxResponseBytes {
		return ErrResponseTooLarge
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return a.normalizeHTTPError(response.StatusCode, raw, operationForEndpoint(endpoint, a.endpoints))
	}
	if responseBody == nil {
		if len(bytes.TrimSpace(raw)) != 0 {
			return fmt.Errorf("%w: unexpected response body", ErrProviderRejected)
		}
		return nil
	}
	contentType := response.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return fmt.Errorf("%w: partner response is not JSON", ErrProviderRejected)
	}
	if err := decodeStrictJSON(raw, responseBody); err != nil {
		return fmt.Errorf("%w: invalid partner response", ErrProviderRejected)
	}
	return nil
}

func (a *HTTPPartnerAdapter) normalizeHTTPError(statusCode int, raw []byte, operation string) error {
	partnerErr := partnerErrorResponse{}
	_ = decodeStrictJSON(raw, &partnerErr)
	kind := ErrorKindUnknown
	retryable := partnerErr.Retryable
	ambiguous := false
	base := ErrAdapterUnavailable
	switch statusCode {
	case http.StatusBadRequest:
		kind, base = ErrorKindInvalidRequest, ErrInvalidRequest
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity:
		kind, base = ErrorKindRejected, ErrProviderRejected
	case http.StatusNotFound:
		kind, base = ErrorKindNotFound, ErrPayoutNotFound
	case http.StatusConflict:
		kind, base, ambiguous = ErrorKindConflict, ErrProviderAmbiguous, operation == operationInitiatePayout
	case http.StatusTooManyRequests:
		kind, base, retryable = ErrorKindRateLimited, ErrProviderTemporary, true
	default:
		if statusCode >= http.StatusInternalServerError {
			kind, base, retryable = ErrorKindTemporary, ErrProviderTemporary, true
		}
	}
	if (operation == operationInitiatePayout || operation == operationCancel) && retryable {
		ambiguous = true
		base = ErrProviderAmbiguous
	}
	return &ProviderError{Provider: a.Name(), Operation: operation, Code: partnerErr.Code, Kind: kind, Retryable: retryable, Ambiguous: ambiguous, err: base}
}

func (a *HTTPPartnerAdapter) convertPayoutResponse(response partnerPayoutResponse, quote *quoteBinding, correlationID string, metadata map[string]string) (PayoutResult, error) {
	if response.Provider != a.profile.Provider || response.APIVersion != a.profile.APIVersion || response.CorrelationID != correlationID ||
		strings.TrimSpace(response.ID) == "" || strings.TrimSpace(response.QuoteID) == "" || !validStatus(response.Status) {
		return PayoutResult{}, fmt.Errorf("%w: payout binding mismatch", ErrProviderRejected)
	}
	fiatAmount, err := parsePositiveDecimal(response.FiatAmount)
	if err != nil {
		return PayoutResult{}, err
	}
	cryptoAmount, err := parsePositiveInt(response.CryptoAmount)
	if err != nil {
		return PayoutResult{}, err
	}
	fee, err := parseNonNegativeInt(response.Fee)
	if err != nil {
		return PayoutResult{}, err
	}
	if quote != nil && (response.QuoteID != quote.quote.ID || !fiatAmount.Equal(quote.quote.FiatAmount) || !cryptoAmount.Equal(quote.quote.Request.CryptoAmount) || !fee.Equal(quote.quote.Fee)) {
		return PayoutResult{}, fmt.Errorf("%w: payout amount does not match quote", ErrProviderRejected)
	}
	if response.InitiatedAt.IsZero() || response.StatusUpdatedAt.IsZero() || response.StatusUpdatedAt.Before(response.InitiatedAt) {
		return PayoutResult{}, fmt.Errorf("%w: invalid payout timestamps", ErrProviderRejected)
	}
	if strings.TrimSpace(response.Reference) == "" || !externalIDPattern.MatchString(response.Reference) {
		return PayoutResult{}, fmt.Errorf("%w: invalid payout reference", ErrProviderRejected)
	}
	if response.Status == StatusCompleted && response.CompletedAt == nil {
		return PayoutResult{}, fmt.Errorf("%w: completed payout lacks finality timestamp", ErrProviderRejected)
	}
	if response.Status != StatusCompleted && response.CompletedAt != nil {
		return PayoutResult{}, fmt.Errorf("%w: non-completed payout has completion timestamp", ErrProviderRejected)
	}
	if response.Status == StatusFailed && strings.TrimSpace(response.FailureCode) == "" {
		return PayoutResult{}, fmt.Errorf("%w: failed payout lacks failure code", ErrProviderRejected)
	}
	if response.Status != StatusFailed && (response.FailureCode != "" || response.FailureReason != "" || response.Retryable) {
		return PayoutResult{}, fmt.Errorf("%w: non-failed payout contains failure fields", ErrProviderRejected)
	}
	result := PayoutResult{
		ID: response.ID, QuoteID: response.QuoteID, Status: response.Status, Provider: response.Provider,
		FiatAmount: fiatAmount, CryptoAmount: cryptoAmount, Fee: fee, Reference: response.Reference,
		Metadata: mergeStringMaps(response.Metadata, metadata), InitiatedAt: response.InitiatedAt.UTC(),
		CompletedAt: response.CompletedAt, StatusUpdatedAt: response.StatusUpdatedAt.UTC(),
		FailureReason: response.FailureReason, FailureCode: response.FailureCode, Retryable: response.Retryable,
		AuditFields: map[string]string{
			"profile_id": a.profile.ID, "profile_state": string(a.profile.State), "api_version": a.profile.APIVersion,
			"execution_environment": string(a.profile.Environment), "correlation_id": correlationID,
			"production_floor_eligible": strconv.FormatBool(a.profile.State == ProfileCertifiedEnabled && a.profile.Environment == EnvironmentProduction),
		},
	}
	return result, nil
}

func (a *HTTPPartnerAdapter) storePayout(result PayoutResult, binding WebhookBinding) {
	cloned := clonePayoutResult(result)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.payouts[result.ID] = cloned
	a.payoutBindings[result.ID] = binding
	cacheMetadataKeys(a.lookupKeys, cloned)
}

func (a *HTTPPartnerAdapter) releasePayoutReservation(ctx context.Context, result PayoutResult) error {
	idempotencyKey := result.Metadata["idempotency_key"]
	if idempotencyKey == "" {
		return fmt.Errorf("%w: payout idempotency key missing", ErrProviderRejected)
	}
	a.mu.RLock()
	quote, ok := a.quotes[result.QuoteID]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: payout quote binding missing", ErrProviderRejected)
	}
	a.mu.RLock()
	binding := a.payoutBindings[result.ID]
	a.mu.RUnlock()
	if binding.ReservationDay == "" {
		return fmt.Errorf("%w: payout reservation day missing", ErrProviderRejected)
	}
	reservationKey := binding.ReservationDay + "|" + quote.corridor.ID
	if err := a.dailyLimits.ReleaseDailyAmount(ctx, reservationKey, idempotencyKey); err != nil {
		return fmt.Errorf("%w: release daily-limit reservation", ErrAdapterUnavailable)
	}
	return nil
}

func (a *HTTPPartnerAdapter) endpointURL(endpoint string) (*url.URL, error) {
	if !validPartnerPath(endpoint, false) || !pathAllowed(endpoint, a.allowedPathPrefixes) {
		return nil, fmt.Errorf("%w: endpoint path is not allowlisted", ErrInvalidRequest)
	}
	resolved := *a.baseURL
	basePath := strings.TrimSuffix(a.baseURL.EscapedPath(), "/")
	resolved.RawPath = ""
	resolved.Path = basePath + endpoint
	resolved.RawQuery = ""
	resolved.Fragment = ""
	if resolved.Scheme != a.baseURL.Scheme || resolved.Hostname() != a.baseURL.Hostname() || resolved.Port() != a.baseURL.Port() {
		return nil, fmt.Errorf("%w: endpoint changed partner origin", ErrInvalidRequest)
	}
	return &resolved, nil
}

func validPartnerPath(value string, allowPlaceholder bool) bool {
	if value == "" || !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#\\") || strings.Contains(value, "//") {
		return false
	}
	cleanedValue := value
	if allowPlaceholder {
		if strings.Count(value, "{payout_id}") != 1 {
			return false
		}
		cleanedValue = strings.ReplaceAll(value, "{payout_id}", "payout-id")
	}
	return path.Clean(cleanedValue) == cleanedValue && !strings.Contains(cleanedValue, "..")
}

func pathAllowed(endpoint string, prefixes []string) bool {
	endpoint = strings.ReplaceAll(endpoint, "{payout_id}", "payout-id")
	for _, prefix := range prefixes {
		if endpoint == prefix || strings.HasPrefix(endpoint, strings.TrimSuffix(prefix, "/")+"/") {
			return true
		}
	}
	return false
}

func expandPayoutPath(template string, payoutID string) string {
	return strings.ReplaceAll(template, "{payout_id}", url.PathEscape(payoutID))
}

func decodeStrictJSON(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func hashStrictJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

func parsePositiveDecimal(raw string) (sdkmath.LegacyDec, error) {
	if raw == "" || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "eE+") {
		return sdkmath.LegacyDec{}, ErrInvalidRequest
	}
	value, err := sdkmath.LegacyNewDecFromStr(raw)
	if err != nil || !value.IsPositive() || value.String() != raw {
		return sdkmath.LegacyDec{}, ErrInvalidRequest
	}
	return value, nil
}

func parsePositiveInt(raw string) (sdkmath.Int, error) {
	value, ok := sdkmath.NewIntFromString(raw)
	if !ok || !value.IsPositive() || value.String() != raw {
		return sdkmath.Int{}, ErrInvalidRequest
	}
	return value, nil
}

func parseNonNegativeInt(raw string) (sdkmath.Int, error) {
	value, ok := sdkmath.NewIntFromString(raw)
	if !ok || value.IsNegative() || value.String() != raw {
		return sdkmath.Int{}, ErrInvalidRequest
	}
	return value, nil
}

func requiredOperationKeys(quote Quote, metadata map[string]string) (string, string, error) {
	idempotencyKey := strings.TrimSpace(metadata["idempotency_key"])
	correlationID := strings.TrimSpace(metadata["correlation_id"])
	if correlationID == "" {
		correlationID = strings.TrimSpace(quote.Request.CorrelationID)
	}
	if idempotencyKey == "" || correlationID == "" || !externalIDPattern.MatchString(idempotencyKey) || !externalIDPattern.MatchString(correlationID) {
		return "", "", ErrInvalidRequest
	}
	if quote.Request.CorrelationID != "" && quote.Request.CorrelationID != correlationID {
		return "", "", ErrInvalidRequest
	}
	return idempotencyKey, correlationID, nil
}

func payoutOperationBinding(req PayoutRequest, idempotencyKey string, correlationID string) (string, error) {
	type binding struct {
		IdempotencyKey string `json:"idempotency_key"`
		CorrelationID  string `json:"correlation_id"`
		QuoteID        string `json:"quote_id"`
		CryptoTxRef    string `json:"crypto_tx_ref"`
		Destination    string `json:"destination"`
	}
	return hashStrictJSON(binding{IdempotencyKey: idempotencyKey, CorrelationID: correlationID, QuoteID: req.Quote.ID, CryptoTxRef: req.CryptoTxRef, Destination: req.Destination})
}

func validHeaderName(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') && !strings.ContainsRune("!#$%&'*+-.^_`|~", character) {
			return false
		}
	}
	return true
}

func validateOperationalMetadata(metadata map[string]string) error {
	if len(metadata) == 0 || len(metadata) > 32 {
		return ErrInvalidRequest
	}
	for key, value := range metadata {
		lowerKey := strings.ToLower(key)
		if !externalIDPattern.MatchString(key) || !externalIDPattern.MatchString(value) ||
			strings.Contains(lowerKey, "account") || strings.Contains(lowerKey, "routing") ||
			strings.Contains(lowerKey, "iban") || strings.Contains(lowerKey, "swift") ||
			strings.Contains(lowerKey, "beneficiary") || strings.Contains(lowerKey, "bank") ||
			strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "credential") {
			return ErrInvalidRequest
		}
	}
	return nil
}

func reconcilePayoutStatus(current PayoutResult, updated PayoutResult) error {
	if current.ID == "" || current.Provider != updated.Provider || current.QuoteID != updated.QuoteID ||
		!current.FiatAmount.Equal(updated.FiatAmount) || !current.CryptoAmount.Equal(updated.CryptoAmount) ||
		!current.Fee.Equal(updated.Fee) || current.Reference != updated.Reference || updated.StatusUpdatedAt.Before(current.StatusUpdatedAt) {
		return fmt.Errorf("%w: payout reconciliation mismatch", ErrProviderRejected)
	}
	if current.Status.IsTerminal() && updated.Status != current.Status {
		return fmt.Errorf("%w: terminal payout status changed", ErrProviderRejected)
	}
	if !validStatusTransition(current.Status, updated.Status) {
		return fmt.Errorf("%w: payout status regressed", ErrProviderRejected)
	}
	return nil
}

func validStatusTransition(current Status, updated Status) bool {
	if current == updated {
		return true
	}
	switch current {
	case StatusPending:
		return updated == StatusProcessing || updated == StatusCompleted || updated == StatusFailed || updated == StatusCancelled
	case StatusProcessing:
		return updated == StatusCompleted || updated == StatusFailed || updated == StatusCancelled
	case StatusCompleted, StatusFailed, StatusCancelled:
		return false
	default:
		return false
	}
}

func quotesEqual(left Quote, right Quote) bool {
	left.AuditFields = nil
	right.AuditFields = nil
	return left.ID == right.ID && left.Provider == right.Provider && quoteRequestsEqual(left.Request, right.Request) &&
		left.FiatAmount.Equal(right.FiatAmount) && left.ExchangeRate.Equal(right.ExchangeRate) && left.Fee.Equal(right.Fee) &&
		left.CreatedAt.Equal(right.CreatedAt) && left.ExpiresAt.Equal(right.ExpiresAt)
}

func quoteRequestsEqual(left QuoteRequest, right QuoteRequest) bool {
	return left.CryptoSymbol == right.CryptoSymbol && left.CryptoDenom == right.CryptoDenom &&
		left.CryptoDecimals == right.CryptoDecimals && left.CryptoAmount.Equal(right.CryptoAmount) && left.FiatCurrency == right.FiatCurrency &&
		left.PaymentMethod == right.PaymentMethod && left.Sender == right.Sender && left.Destination == right.Destination &&
		left.Jurisdiction == right.Jurisdiction && left.BeneficiaryReference == right.BeneficiaryReference &&
		left.CorrelationID == right.CorrelationID && left.Compliance == right.Compliance
}

func validStatus(status Status) bool {
	switch status {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

func operationForEndpoint(endpoint string, endpoints PartnerEndpoints) string {
	switch {
	case endpoint == endpoints.Quote:
		return operationQuote
	case endpoint == endpoints.Payout:
		return operationInitiatePayout
	case endpoint == endpoints.MetadataLookup:
		return operationMetadataLookup
	case matchesEndpointTemplate(endpoint, endpoints.Cancel):
		return operationCancel
	case matchesEndpointTemplate(endpoint, endpoints.Status):
		return operationGetStatus
	default:
		return "health"
	}
}

func matchesEndpointTemplate(endpoint string, template string) bool {
	parts := strings.Split(template, "{payout_id}")
	return len(parts) == 2 && strings.HasPrefix(endpoint, parts[0]) && strings.HasSuffix(endpoint, parts[1]) &&
		len(endpoint) > len(parts[0])+len(parts[1])
}

var _ Adapter = (*HTTPPartnerAdapter)(nil)
var _ MetadataLookupAdapter = (*HTTPPartnerAdapter)(nil)
var _ ProfiledAdapter = (*HTTPPartnerAdapter)(nil)
var _ WebhookBindingRepository = (*HTTPPartnerAdapter)(nil)
