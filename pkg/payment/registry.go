// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-49C: Payment processor registry and routing.
package payment

import (
	"context"
	"sort"
	"sync"
)

// FeeSchedule represents gateway fee configuration.
type FeeSchedule struct {
	FixedFee int64   // minor units
	Percent  float64 // percent of amount (e.g. 2.9 for 2.9%)
}

// ProcessorPreference defines fallback ordering.
type ProcessorPreference struct {
	Primary   GatewayType
	Secondary GatewayType
	Tertiary  GatewayType
}

// ProcessorSelectionRequest defines adapter selection parameters.
type ProcessorSelectionRequest struct {
	ProviderID     string
	Region         string
	Amount         Amount
	RequireHealthy bool
}

// PaymentProcessorRegistry provides adapter selection and routing.
type PaymentProcessorRegistry struct {
	mu            sync.RWMutex
	adapters      map[GatewayType]Gateway
	fees          map[GatewayType]FeeSchedule
	regionPrefs   map[string]ProcessorPreference
	providerPrefs map[string]ProcessorPreference
}

// NewPaymentProcessorRegistry creates a new registry.
func NewPaymentProcessorRegistry() *PaymentProcessorRegistry {
	return &PaymentProcessorRegistry{
		adapters:      make(map[GatewayType]Gateway),
		fees:          make(map[GatewayType]FeeSchedule),
		regionPrefs:   make(map[string]ProcessorPreference),
		providerPrefs: make(map[string]ProcessorPreference),
	}
}

// RegisterAdapter registers a gateway adapter.
func (r *PaymentProcessorRegistry) RegisterAdapter(gateway GatewayType, adapter Gateway) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[gateway] = adapter
}

// RegisterFees registers a fee schedule for a gateway.
func (r *PaymentProcessorRegistry) RegisterFees(gateway GatewayType, fee FeeSchedule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fees[gateway] = fee
}

// SetRegionPreference sets regional routing preference.
func (r *PaymentProcessorRegistry) SetRegionPreference(region string, pref ProcessorPreference) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.regionPrefs[region] = pref
}

// SetProviderPreference sets provider-specific routing preference.
func (r *PaymentProcessorRegistry) SetProviderPreference(providerID string, pref ProcessorPreference) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providerPrefs[providerID] = pref
}

// SelectAdapter selects the best available adapter with fallback.
func (r *PaymentProcessorRegistry) SelectAdapter(ctx context.Context, req ProcessorSelectionRequest) (Gateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	preference := ProcessorPreference{}
	if req.ProviderID != "" {
		if pref, ok := r.providerPrefs[req.ProviderID]; ok {
			preference = pref
		}
	}
	if preference.Primary == "" && req.Region != "" {
		if pref, ok := r.regionPrefs[req.Region]; ok {
			preference = pref
		}
	}

	candidates := r.preferenceChain(preference)
	if len(candidates) == 0 {
		candidates = r.bestGatewaysByFeeLocked(req.Amount)
	}

	for _, gateway := range candidates {
		adapter, ok := r.adapters[gateway]
		if !ok {
			continue
		}
		if req.RequireHealthy && !adapter.IsHealthy(ctx) {
			continue
		}
		return adapter, nil
	}

	return nil, ErrNoAvailableGateway
}

// EstimateFee returns the fee amount for a gateway.
func (r *PaymentProcessorRegistry) EstimateFee(amount Amount, gateway GatewayType) Amount {
	r.mu.RLock()
	defer r.mu.RUnlock()
	schedule := r.fees[gateway]
	feeValue := schedule.FixedFee + int64(float64(amount.Value)*schedule.Percent/100.0)
	return Amount{Value: feeValue, Currency: amount.Currency}
}

func (r *PaymentProcessorRegistry) preferenceChain(pref ProcessorPreference) []GatewayType {
	seen := make(map[GatewayType]struct{})
	var chain []GatewayType
	for _, gateway := range []GatewayType{pref.Primary, pref.Secondary, pref.Tertiary} {
		if gateway == "" {
			continue
		}
		if _, ok := seen[gateway]; ok {
			continue
		}
		seen[gateway] = struct{}{}
		chain = append(chain, gateway)
	}
	return chain
}

func (r *PaymentProcessorRegistry) bestGatewaysByFeeLocked(amount Amount) []GatewayType {
	type feeCandidate struct {
		gateway GatewayType
		fee     int64
	}

	candidates := make([]feeCandidate, 0, len(r.adapters))
	for gateway := range r.adapters {
		fee := r.fees[gateway]
		feeValue := fee.FixedFee + int64(float64(amount.Value)*fee.Percent/100.0)
		candidates = append(candidates, feeCandidate{gateway: gateway, fee: feeValue})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].fee < candidates[j].fee
	})

	result := make([]GatewayType, 0, len(candidates))
	for _, cand := range candidates {
		result = append(result, cand.gateway)
	}
	return result
}
