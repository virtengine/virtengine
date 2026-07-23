package offramp

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultQuoteValidity    = 5 * time.Minute
	operationQuote          = "quote"
	operationInitiatePayout = "initiate_payout"
	operationGetStatus      = "get_status"
	operationCancel         = "cancel"
	operationMetadataLookup = "metadata_lookup"
)

type quotePlan struct {
	request    QuoteRequest
	candidates []Quote
}

// bridgeImpl aggregates multiple off-ramp adapters.
type bridgeImpl struct {
	adapters             map[string]Adapter
	adapterMu            sync.RWMutex
	operations           map[string]PayoutResult
	lookupKeys           map[string]string
	quotePlans           map[string]quotePlan
	opMu                 sync.RWMutex
	now                  func() time.Time
	mode                 ExecutionMode
	allowExternalBlocked bool
	repository           PayoutRepository
	profileAuthorizer    ProfileAuthorizer
}

// NewBridge creates a new off-ramp bridge.
func NewBridge() *bridgeImpl {
	return newBridgeWithDependencies(ExecutionModeLegacy, false, nil, nil, func() time.Time {
		return time.Now().UTC()
	})
}

// ProductionBridgeConfig supplies the durable state and trusted support-matrix
// authority required by a production bridge.
type ProductionBridgeConfig struct {
	Repository PayoutRepository
	Authorizer ProfileAuthorizer
}

// NewProductionBridge creates a bridge that accepts only trusted certified
// production profiles, rejects test adapters, and requires durable payout
// state. Omitting config intentionally produces a fail-closed bridge while
// preserving compatibility for callers that only validate rejection paths.
func NewProductionBridge(config ...ProductionBridgeConfig) *bridgeImpl {
	var repository PayoutRepository
	var authorizer ProfileAuthorizer
	if len(config) == 1 {
		repository = config[0].Repository
		authorizer = config[0].Authorizer
	}
	return newBridgeWithDependencies(ExecutionModeProduction, false, repository, authorizer, func() time.Time {
		return time.Now().UTC()
	})
}

// NewEngineeringSandboxBridge creates an engineering bridge. Externally
// blocked profiles require explicit opt-in and remain visible in audit fields.
func NewEngineeringSandboxBridge(allowExternalBlocked bool) *bridgeImpl {
	return newBridgeWithDependencies(ExecutionModeSandbox, allowExternalBlocked, nil, nil, func() time.Time {
		return time.Now().UTC()
	})
}

func newBridgeWithClock(now func() time.Time) *bridgeImpl {
	return newBridgeWithDependencies(ExecutionModeLegacy, false, nil, nil, now)
}

func newBridgeWithOptions(mode ExecutionMode, allowExternalBlocked bool, now func() time.Time) *bridgeImpl {
	return newBridgeWithDependencies(mode, allowExternalBlocked, nil, nil, now)
}

func newBridgeWithDependencies(mode ExecutionMode, allowExternalBlocked bool, repository PayoutRepository, authorizer ProfileAuthorizer, now func() time.Time) *bridgeImpl {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &bridgeImpl{
		adapters:             make(map[string]Adapter),
		operations:           make(map[string]PayoutResult),
		lookupKeys:           make(map[string]string),
		quotePlans:           make(map[string]quotePlan),
		now:                  now,
		mode:                 mode,
		allowExternalBlocked: allowExternalBlocked,
		repository:           repository,
		profileAuthorizer:    authorizer,
	}
}

// RegisterAdapter registers an adapter with the bridge.
func (b *bridgeImpl) RegisterAdapter(adapter Adapter) error {
	if adapter == nil {
		return ErrAdapterUnavailable
	}
	name := strings.TrimSpace(adapter.Name())
	if name == "" {
		return ErrInvalidRequest
	}
	if b.mode == ExecutionModeProduction {
		if testAdapter, ok := adapter.(TestOnlyAdapter); ok && testAdapter.IsTestOnly() {
			return ErrTestAdapter
		}
		if b.repository == nil || !b.repository.Durable() || b.profileAuthorizer == nil {
			return ErrProfileNotExecutable
		}
		if _, ok := adapter.(interface{ productionPayoutAdapter() }); !ok {
			return ErrProfileNotExecutable
		}
	}
	if b.mode != ExecutionModeLegacy {
		profiled, ok := adapter.(ProfiledAdapter)
		if !ok {
			return ErrProfileNotExecutable
		}
		profile := profiled.Profile()
		if profile.Provider != name {
			return fmt.Errorf("%w: adapter name does not match profile provider", ErrInvalidRequest)
		}
		if err := profile.ValidateForExecution(b.mode, b.allowExternalBlocked); err != nil {
			return err
		}
		if b.mode == ExecutionModeProduction {
			if err := b.profileAuthorizer.AuthorizePayoutProfile(profile); err != nil {
				return ErrProfileNotExecutable
			}
		}
	}
	b.adapterMu.Lock()
	defer b.adapterMu.Unlock()
	if _, exists := b.adapters[name]; exists {
		return ErrInvalidRequest
	}
	b.adapters[name] = adapter
	return nil
}

// GetQuote returns the best quote across adapters.
func (b *bridgeImpl) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	if err := validateQuoteRequest(req); err != nil {
		return Quote{}, err
	}

	evaluatedAt := b.now()
	candidates, err := b.collectQuoteCandidates(ctx, req, evaluatedAt)
	if err != nil {
		return Quote{}, err
	}

	selected := cloneQuote(candidates[0])
	selected.AuditFields = mergeStringMaps(selected.AuditFields, map[string]string{
		"bridge_selected_provider": selected.Provider,
		"bridge_candidate_count":   strconv.Itoa(len(candidates)),
		"bridge_selected_rank":     "1",
	})

	b.opMu.Lock()
	b.quotePlans[selected.ID] = quotePlan{
		request:    req,
		candidates: cloneQuotes(candidates),
	}
	b.opMu.Unlock()

	return selected, nil
}

// InitiatePayout executes a payout via the selected adapter.
func (b *bridgeImpl) InitiatePayout(ctx context.Context, quote Quote, cryptoTxRef string, destination string, metadata map[string]string) (PayoutResult, error) {
	now := b.now()
	if err := validatePayoutInputs(quote, cryptoTxRef, destination, now); err != nil {
		return PayoutResult{}, err
	}
	b.opMu.RLock()
	plan, knownQuote := b.quotePlans[quote.ID]
	b.opMu.RUnlock()
	if b.mode != ExecutionModeLegacy {
		if !knownQuote || len(plan.candidates) == 0 || !quotesEqual(plan.candidates[0], quote) {
			return PayoutResult{}, ErrInvalidRequest
		}
	}

	metadata = cloneStringMap(metadata)
	if existing, ok := b.lookupExistingPayout(ctx, "", metadata); ok {
		return existing, nil
	}

	candidates := b.payoutCandidatesForQuote(quote)
	var lastErr error

	for idx, candidate := range candidates {
		if candidate.IsExpired(now) {
			lastErr = ErrQuoteExpired
			continue
		}

		adapter, err := b.adapterByName(candidate.Provider)
		if err != nil {
			lastErr = err
			continue
		}
		if err := b.validateAdapterRequest(adapter, candidate.Request, now); err != nil {
			lastErr = err
			continue
		}

		result, err := adapter.InitiatePayout(ctx, PayoutRequest{
			Quote:       candidate,
			CryptoTxRef: cryptoTxRef,
			Destination: destination,
			Metadata:    metadata,
		})
		if err == nil {
			normalized, normErr := b.normalizePayoutResult(candidate, result, metadata, idx+1, false)
			if normErr != nil {
				return PayoutResult{}, normErr
			}
			if err := b.cacheLookupResult(ctx, normalized); err != nil {
				return PayoutResult{}, err
			}
			return clonePayoutResult(normalized), nil
		}

		normalizedErr := normalizeAdapterError(adapter, operationInitiatePayout, err)
		if existing, ok := b.lookupExistingPayout(ctx, candidate.Provider, metadata); ok {
			existing.AuditFields = mergeStringMaps(existing.AuditFields, map[string]string{
				"bridge_recovered_provider": candidate.Provider,
				"bridge_recovery_reason":    "metadata_lookup",
				"bridge_attempt":            strconv.Itoa(idx + 1),
			})
			if err := b.cacheLookupResult(ctx, existing); err != nil {
				return PayoutResult{}, err
			}
			return existing, nil
		}

		lastErr = normalizedErr
		if !CanFailover(normalizedErr) {
			break
		}
	}

	if lastErr != nil {
		return PayoutResult{}, lastErr
	}
	return PayoutResult{}, ErrAdapterUnavailable
}

// FindPayoutByMetadata resolves an off-ramp payout by idempotent metadata.
func (b *bridgeImpl) FindPayoutByMetadata(ctx context.Context, provider string, metadata map[string]string) (PayoutResult, error) {
	if len(metadata) == 0 {
		return PayoutResult{}, ErrInvalidRequest
	}

	if result, ok := b.lookupExistingPayout(ctx, provider, metadata); ok {
		return result, nil
	}

	return PayoutResult{}, NormalizeError(provider, operationMetadataLookup, ErrPayoutNotFound)
}

// GetStatus retrieves payout status and refreshes from adapter when available.
func (b *bridgeImpl) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	b.opMu.RLock()
	result, ok := b.operations[payoutID]
	b.opMu.RUnlock()
	if !ok && b.repository != nil {
		persisted, err := b.repository.GetPayout(ctx, payoutID)
		if err == nil {
			result, ok = clonePayoutResult(persisted), true
		}
	}
	if !ok {
		return PayoutResult{}, ErrPayoutNotFound
	}

	result = clonePayoutResult(result)
	if result.IsTerminal() {
		return result, nil
	}

	adapter, err := b.adapterByName(result.Provider)
	if err != nil {
		return PayoutResult{}, err
	}

	updated, err := adapter.GetStatus(ctx, payoutID)
	if err != nil {
		normalized := normalizeAdapterError(adapter, operationGetStatus, err)
		if IsNotFound(normalized) {
			recovered, recoveryErr := b.restoreAdapterPayoutBinding(ctx, adapter, result)
			if recoveryErr != nil {
				return PayoutResult{}, recoveryErr
			}
			if err := b.cacheLookupResult(ctx, recovered); err != nil {
				return PayoutResult{}, err
			}
			updated, err = adapter.GetStatus(ctx, payoutID)
			if err != nil {
				return PayoutResult{}, normalizeAdapterError(adapter, operationGetStatus, err)
			}
			updated.AuditFields = mergeStringMaps(updated.AuditFields, recovered.AuditFields)
		} else {
			result.AuditFields = mergeStringMaps(result.AuditFields, map[string]string{
				"bridge_status_source":     "cache",
				"bridge_status_refresh":    "stale",
				"bridge_status_last_error": normalized.Error(),
			})
			if err := b.cacheLookupResult(ctx, result); err != nil {
				return PayoutResult{}, err
			}
			return result, nil
		}
	}

	normalized, normErr := b.normalizePayoutResult(resultingQuote(result), updated, result.Metadata, intValue(result.AuditFields["bridge_attempt"], 1), true)
	if normErr != nil {
		return PayoutResult{}, normErr
	}
	normalized.AuditFields = mergeStringMaps(normalized.AuditFields, map[string]string{
		"bridge_status_source": "provider",
	})
	if err := b.cacheLookupResult(ctx, normalized); err != nil {
		return PayoutResult{}, err
	}
	return clonePayoutResult(normalized), nil
}

func (b *bridgeImpl) restoreAdapterPayoutBinding(ctx context.Context, adapter Adapter, expected PayoutResult) (PayoutResult, error) {
	if expected.ID == "" || expected.Provider == "" || expected.Provider != adapter.Name() || expected.IsTerminal() || len(expected.Metadata) == 0 {
		return PayoutResult{}, ErrPayoutNotFound
	}
	restorer, ok := adapter.(PayoutBindingRecoveryAdapter)
	if !ok {
		return PayoutResult{}, ErrPayoutNotFound
	}
	recovered, err := restorer.RestorePayoutBinding(ctx, clonePayoutResult(expected))
	if err != nil {
		return PayoutResult{}, normalizeAdapterError(adapter, operationMetadataLookup, err)
	}
	if err := validateRecoveredPayoutBinding(expected, recovered); err != nil {
		return PayoutResult{}, err
	}
	recovered.AuditFields = mergeStringMaps(recovered.AuditFields, map[string]string{
		"bridge_recovered_provider": adapter.Name(),
		"bridge_recovery_reason":    "durable_binding_restore",
	})
	return recovered, nil
}

func validateRecoveredPayoutBinding(expected, recovered PayoutResult) error {
	if expected.ID != recovered.ID || expected.Provider != recovered.Provider || expected.QuoteID != recovered.QuoteID ||
		expected.Reference != recovered.Reference || !expected.FiatAmount.Equal(recovered.FiatAmount) ||
		!expected.CryptoAmount.Equal(recovered.CryptoAmount) || !expected.Fee.Equal(recovered.Fee) ||
		expected.DailyReservationKey != recovered.DailyReservationKey ||
		expected.DailyReservationOperationID != recovered.DailyReservationOperationID ||
		!maps.Equal(expected.Metadata, recovered.Metadata) || !expected.InitiatedAt.Equal(recovered.InitiatedAt) ||
		reconcilePayoutStatus(expected, recovered) != nil {
		return ErrProviderRejected
	}
	return nil
}

// Cancel attempts to cancel a payout.
func (b *bridgeImpl) Cancel(ctx context.Context, payoutID string) error {
	b.opMu.RLock()
	result, ok := b.operations[payoutID]
	b.opMu.RUnlock()
	if !ok && b.repository != nil {
		persisted, err := b.repository.GetPayout(ctx, payoutID)
		if err == nil {
			result, ok = clonePayoutResult(persisted), true
		}
	}
	if !ok {
		return ErrPayoutNotFound
	}

	if result.Status == StatusCancelled {
		return nil
	}
	if result.Status.IsTerminal() {
		return ErrPayoutNotCancellable
	}

	adapter, err := b.adapterByName(result.Provider)
	if err != nil {
		return err
	}

	if err := adapter.Cancel(ctx, payoutID); err != nil {
		return normalizeAdapterError(adapter, operationCancel, err)
	}

	cancelled := clonePayoutResult(result)
	cancelled.Status = StatusCancelled
	cancelled.StatusUpdatedAt = b.now()
	cancelled.Retryable = false
	cancelled.AuditFields = mergeStringMaps(cancelled.AuditFields, map[string]string{
		"bridge_cancelled":           "true",
		"bridge_cancel_confirmation": "provider_ack",
		"bridge_status_source":       "bridge",
	})

	return b.cacheLookupResult(ctx, cancelled)
}

// ListProviders lists registered adapters.
func (b *bridgeImpl) ListProviders() []string {
	b.adapterMu.RLock()
	defer b.adapterMu.RUnlock()

	providers := make([]string, 0, len(b.adapters))
	for name := range b.adapters {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

func (b *bridgeImpl) collectQuoteCandidates(ctx context.Context, req QuoteRequest, evaluatedAt time.Time) ([]Quote, error) {
	providerNames := b.ListProviders()
	candidates := make([]Quote, 0, len(providerNames))
	var lastErr error

	for _, name := range providerNames {
		adapter, err := b.adapterByName(name)
		if err != nil {
			lastErr = err
			continue
		}
		if err := b.validateAdapterRequest(adapter, req, evaluatedAt); err != nil {
			lastErr = err
			continue
		}
		if !adapter.IsHealthy(ctx) {
			lastErr = ErrAdapterUnavailable
			continue
		}
		if !adapter.SupportsCurrency(req.FiatCurrency) || !adapter.SupportsMethod(req.PaymentMethod) {
			continue
		}

		quote, err := adapter.GetQuote(ctx, req)
		if err != nil {
			lastErr = normalizeAdapterError(adapter, operationQuote, err)
			continue
		}

		normalized, normErr := normalizeQuote(quote, req, name, evaluatedAt)
		if normErr != nil {
			lastErr = normErr
			continue
		}
		candidates = append(candidates, normalized)
	}

	if len(candidates) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, ErrAdapterUnavailable
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return quoteLess(candidates[i], candidates[j])
	})

	for idx := range candidates {
		candidates[idx].AuditFields = mergeStringMaps(candidates[idx].AuditFields, map[string]string{
			"bridge_candidate_rank":  strconv.Itoa(idx + 1),
			"bridge_candidate_count": strconv.Itoa(len(candidates)),
		})
	}

	return candidates, nil
}

func (b *bridgeImpl) payoutCandidatesForQuote(quote Quote) []Quote {
	b.opMu.RLock()
	plan, ok := b.quotePlans[quote.ID]
	b.opMu.RUnlock()
	if ok && len(plan.candidates) > 0 {
		return cloneQuotes(plan.candidates)
	}
	return []Quote{cloneQuote(quote)}
}

func (b *bridgeImpl) lookupExistingPayout(ctx context.Context, provider string, metadata map[string]string) (PayoutResult, bool) {
	if len(metadata) == 0 {
		return PayoutResult{}, false
	}

	keys := []string{metadataLookupKey(provider, metadata)}
	if provider != "" {
		keys = append(keys, metadataLookupKey("", metadata))
	}

	b.opMu.RLock()
	for _, key := range keys {
		if payoutID, ok := b.lookupKeys[key]; ok {
			if result, exists := b.operations[payoutID]; exists {
				b.opMu.RUnlock()
				return clonePayoutResult(result), true
			}
		}
	}
	for _, result := range b.operations {
		if provider != "" && result.Provider != provider {
			continue
		}
		if metadataMatches(result.Metadata, metadata) {
			b.opMu.RUnlock()
			return clonePayoutResult(result), true
		}
	}
	b.opMu.RUnlock()
	if b.repository != nil {
		persisted, err := b.repository.FindPayout(ctx, provider, metadata)
		if err == nil {
			if b.cacheLookupResult(ctx, persisted) != nil {
				return PayoutResult{}, false
			}
			return clonePayoutResult(persisted), true
		}
	}

	providers := b.providersForLookup(provider)
	for _, adapter := range providers {
		lookup, ok := adapter.(MetadataLookupAdapter)
		if !ok {
			continue
		}
		result, err := lookup.FindPayoutByMetadata(ctx, metadata)
		if err != nil {
			continue
		}
		normalized, normErr := b.normalizePayoutResult(resultingQuote(result), result, metadata, 1, true)
		if normErr != nil {
			continue
		}
		normalized.AuditFields = mergeStringMaps(normalized.AuditFields, map[string]string{
			"bridge_recovered_provider": normalized.Provider,
			"bridge_recovery_reason":    "metadata_lookup",
		})
		if b.cacheLookupResult(ctx, normalized) != nil {
			continue
		}
		return clonePayoutResult(normalized), true
	}

	return PayoutResult{}, false
}

func (b *bridgeImpl) providersForLookup(provider string) []Adapter {
	b.adapterMu.RLock()
	defer b.adapterMu.RUnlock()

	if provider != "" {
		adapter, ok := b.adapters[provider]
		if !ok {
			return nil
		}
		return []Adapter{adapter}
	}

	names := make([]string, 0, len(b.adapters))
	for name := range b.adapters {
		names = append(names, name)
	}
	sort.Strings(names)

	adapters := make([]Adapter, 0, len(names))
	for _, name := range names {
		adapters = append(adapters, b.adapters[name])
	}
	return adapters
}

func (b *bridgeImpl) normalizePayoutResult(quote Quote, result PayoutResult, metadata map[string]string, attempt int, fromLookup bool) (PayoutResult, error) {
	now := b.now()
	if result.Provider != "" && quote.Provider != "" && result.Provider != quote.Provider {
		return PayoutResult{}, ErrInvalidRequest
	}
	if result.QuoteID != "" && quote.ID != "" && result.QuoteID != quote.ID {
		return PayoutResult{}, ErrInvalidRequest
	}
	if !result.FiatAmount.IsNil() && !quote.FiatAmount.IsNil() && !result.FiatAmount.Equal(quote.FiatAmount) {
		return PayoutResult{}, ErrInvalidRequest
	}
	if !result.CryptoAmount.IsNil() && !quote.Request.CryptoAmount.IsNil() && !result.CryptoAmount.Equal(quote.Request.CryptoAmount) {
		return PayoutResult{}, ErrInvalidRequest
	}
	if result.Provider == "" {
		result.Provider = quote.Provider
	}
	if result.QuoteID == "" {
		result.QuoteID = quote.ID
	}
	if result.ID == "" {
		return PayoutResult{}, ErrInvalidRequest
	}
	if result.Status == "" {
		result.Status = StatusProcessing
	}
	if result.FiatAmount.IsNil() || result.FiatAmount.IsZero() {
		result.FiatAmount = quote.FiatAmount
	}
	if result.CryptoAmount.IsNil() || !result.CryptoAmount.IsPositive() {
		result.CryptoAmount = quote.Request.CryptoAmount
	}
	if result.Fee.IsNil() {
		result.Fee = quote.Fee
	}
	if result.InitiatedAt.IsZero() {
		result.InitiatedAt = now
	}
	if result.StatusUpdatedAt.IsZero() {
		if result.CompletedAt != nil {
			result.StatusUpdatedAt = result.CompletedAt.UTC()
		} else {
			result.StatusUpdatedAt = now
		}
	}
	if result.Metadata == nil && len(metadata) > 0 {
		result.Metadata = cloneStringMap(metadata)
	}
	result.Metadata = mergeStringMaps(result.Metadata, metadata)
	bridgeAudit := map[string]string{
		"bridge_attempt":           strconv.Itoa(attempt),
		"bridge_status_source":     ternaryString(fromLookup, "provider_lookup", "provider"),
		"bridge_quote_provider":    quote.Provider,
		"bridge_selected_provider": result.Provider,
	}
	if profiled, ok := b.profileForProvider(result.Provider); ok {
		bridgeAudit["bridge_profile_id"] = profiled.ID
		bridgeAudit["bridge_profile_state"] = string(profiled.State)
		bridgeAudit["bridge_execution_environment"] = string(profiled.Environment)
		bridgeAudit["bridge_production_floor_eligible"] = strconv.FormatBool(profiled.State == ProfileCertifiedEnabled && profiled.Environment == EnvironmentProduction)
	}
	result.AuditFields = mergeStringMaps(result.AuditFields, bridgeAudit)

	if result.Status == StatusCompleted && result.CompletedAt == nil {
		completedAt := now
		result.CompletedAt = &completedAt
	}

	if result.Status == StatusFailed && result.FailureReason == "" {
		result.FailureReason = "provider reported failed payout"
	}

	return clonePayoutResult(result), nil
}

func normalizeQuote(quote Quote, req QuoteRequest, provider string, evaluatedAt time.Time) (Quote, error) {
	quote = cloneQuote(quote)
	quote.Request = req
	if quote.Provider == "" {
		quote.Provider = provider
	}
	if quote.Provider == "" || quote.ID == "" {
		return Quote{}, ErrInvalidRequest
	}
	if quote.CreatedAt.IsZero() {
		quote.CreatedAt = evaluatedAt
	}
	if quote.ExpiresAt.IsZero() {
		quote.ExpiresAt = quote.CreatedAt.Add(defaultQuoteValidity)
	}
	if !quote.FiatAmount.IsPositive() || !quote.ExchangeRate.IsPositive() {
		return Quote{}, ErrInvalidRequest
	}
	if !quote.ExpiresAt.After(quote.CreatedAt) {
		return Quote{}, ErrInvalidRequest
	}
	return quote, nil
}

func validateQuoteRequest(req QuoteRequest) error {
	if strings.TrimSpace(req.CryptoSymbol) == "" || strings.TrimSpace(req.CryptoDenom) == "" {
		return ErrInvalidRequest
	}
	if req.CryptoAmount.IsNil() || !req.CryptoAmount.IsPositive() {
		return ErrInvalidRequest
	}
	if req.CryptoDecimals > 18 {
		return ErrInvalidRequest
	}
	if strings.TrimSpace(req.FiatCurrency) == "" || strings.TrimSpace(req.PaymentMethod) == "" {
		return ErrInvalidRequest
	}
	if strings.TrimSpace(req.Sender) == "" || strings.TrimSpace(req.Destination) == "" {
		return ErrInvalidRequest
	}
	if strings.TrimSpace(req.BeneficiaryReference) != "" && strings.TrimSpace(req.Destination) != strings.TrimSpace(req.BeneficiaryReference) {
		return ErrInvalidRequest
	}
	return nil
}

func (b *bridgeImpl) validateAdapterRequest(adapter Adapter, req QuoteRequest, now time.Time) error {
	profiled, ok := adapter.(ProfiledAdapter)
	if !ok {
		if b.mode == ExecutionModeLegacy {
			return nil
		}
		return ErrProfileNotExecutable
	}
	profile := profiled.Profile()
	if b.mode != ExecutionModeLegacy {
		if err := profile.ValidateForExecution(b.mode, b.allowExternalBlocked); err != nil {
			return err
		}
		if b.mode == ExecutionModeProduction {
			if b.profileAuthorizer == nil || b.profileAuthorizer.AuthorizePayoutProfile(profile) != nil {
				return ErrProfileNotExecutable
			}
		}
	}
	if req.Jurisdiction == "" {
		return ErrUnsupportedCorridor
	}
	if _, err := profile.Corridor(req.Jurisdiction, req.FiatCurrency, req.PaymentMethod); err != nil {
		return err
	}
	if strings.TrimSpace(req.BeneficiaryReference) == "" {
		return ErrInvalidRequest
	}
	return validateCompliance(req.Compliance, profile.DecisionRequirements, now)
}

func validatePayoutInputs(quote Quote, cryptoTxRef string, destination string, now time.Time) error {
	if strings.TrimSpace(cryptoTxRef) == "" || strings.TrimSpace(destination) == "" {
		return ErrInvalidRequest
	}
	if err := validateQuoteRequest(quote.Request); err != nil {
		return err
	}
	if _, err := normalizeQuote(quote, quote.Request, quote.Provider, quote.CreatedAt); err != nil {
		return err
	}
	if quote.IsExpired(now) {
		return ErrQuoteExpired
	}
	if quote.Request.Destination != "" && quote.Request.Destination != destination {
		return ErrInvalidRequest
	}
	return nil
}

func quoteLess(left Quote, right Quote) bool {
	switch {
	case left.FiatAmount.GT(right.FiatAmount):
		return true
	case left.FiatAmount.LT(right.FiatAmount):
		return false
	case left.ExchangeRate.GT(right.ExchangeRate):
		return true
	case left.ExchangeRate.LT(right.ExchangeRate):
		return false
	case left.Fee.LT(right.Fee):
		return true
	case left.Fee.GT(right.Fee):
		return false
	case left.ExpiresAt.After(right.ExpiresAt):
		return true
	case left.ExpiresAt.Before(right.ExpiresAt):
		return false
	default:
		return left.Provider < right.Provider
	}
}

func (b *bridgeImpl) cacheLookupResult(ctx context.Context, result PayoutResult) error {
	cloned := clonePayoutResult(result)
	if b.repository != nil {
		if err := b.repository.PutPayout(ctx, cloned); err != nil {
			return fmt.Errorf("%w: durable payout repository unavailable", ErrAdapterUnavailable)
		}
	}
	b.opMu.Lock()
	b.operations[cloned.ID] = cloned
	cacheMetadataKeys(b.lookupKeys, cloned)
	b.opMu.Unlock()
	return nil
}

func cacheMetadataKeys(index map[string]string, result PayoutResult) {
	if key := metadataLookupKey(result.Provider, result.Metadata); key != "" {
		index[key] = result.ID
	}
	if key := metadataLookupKey("", result.Metadata); key != "" {
		index[key] = result.ID
	}
}

func metadataLookupKey(provider string, metadata map[string]string) string {
	if len(metadata) == 0 {
		return ""
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, provider)
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, metadata[key]))
	}
	return strings.Join(parts, "|")
}

func metadataMatches(existing map[string]string, required map[string]string) bool {
	if len(required) == 0 {
		return false
	}
	for key, value := range required {
		if existing[key] != value {
			return false
		}
	}
	return true
}

func normalizeAdapterError(adapter Adapter, operation string, err error) error {
	if err == nil {
		return nil
	}
	if normalizer, ok := adapter.(ErrorNormalizer); ok {
		return NormalizeError(adapter.Name(), operation, normalizer.NormalizeError(operation, err))
	}
	return NormalizeError(adapter.Name(), operation, err)
}

func (b *bridgeImpl) adapterByName(name string) (Adapter, error) {
	b.adapterMu.RLock()
	defer b.adapterMu.RUnlock()
	adapter, ok := b.adapters[name]
	if !ok {
		return nil, ErrAdapterUnavailable
	}
	return adapter, nil
}

func (b *bridgeImpl) profileForProvider(name string) (PayoutProfile, bool) {
	adapter, err := b.adapterByName(name)
	if err != nil {
		return PayoutProfile{}, false
	}
	profiled, ok := adapter.(ProfiledAdapter)
	if !ok {
		return PayoutProfile{}, false
	}
	return profiled.Profile(), true
}

func cloneQuote(quote Quote) Quote {
	quote.AuditFields = cloneStringMap(quote.AuditFields)
	return quote
}

func cloneQuotes(quotes []Quote) []Quote {
	cloned := make([]Quote, 0, len(quotes))
	for _, quote := range quotes {
		cloned = append(cloned, cloneQuote(quote))
	}
	return cloned
}

func clonePayoutResult(result PayoutResult) PayoutResult {
	result.Metadata = cloneStringMap(result.Metadata)
	result.AuditFields = cloneStringMap(result.AuditFields)
	if result.CompletedAt != nil {
		completedAt := result.CompletedAt.UTC()
		result.CompletedAt = &completedAt
	}
	return result
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	return maps.Clone(source)
}

func mergeStringMaps(base map[string]string, extra map[string]string) map[string]string {
	switch {
	case len(base) == 0 && len(extra) == 0:
		return nil
	case len(base) == 0:
		return cloneStringMap(extra)
	case len(extra) == 0:
		return cloneStringMap(base)
	}

	merged := maps.Clone(base)
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func resultingQuote(result PayoutResult) Quote {
	return Quote{
		ID:         result.QuoteID,
		Provider:   result.Provider,
		FiatAmount: result.FiatAmount,
		Fee:        result.Fee,
		Request: QuoteRequest{
			CryptoAmount: result.CryptoAmount,
		},
	}
}

func ternaryString(condition bool, whenTrue string, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func intValue(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
