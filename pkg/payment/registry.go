// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3064: Payment processor registry for routing and fallback.
package payment

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// FeeSchedule represents fee structure for a gateway.
type FeeSchedule struct {
	// FixedFee is a fixed fee in minor units.
	FixedFee int64

	// VariableBps is the variable fee in basis points.
	VariableBps int64
}

// ProcessorRoute configures gateway routing rules.
type ProcessorRoute struct {
	Gateway    GatewayType
	Regions    []string
	Currencies []Currency
	Fee        FeeSchedule
	Enabled    bool
}

// ProviderPreferences defines the ordered fallback chain.
type ProviderPreferences struct {
	Primary   GatewayType
	Secondary GatewayType
	Tertiary  GatewayType
}

// PaymentProcessorRegistry selects adapters by region, currency, and preferences.
type PaymentProcessorRegistry struct {
	mu                  sync.RWMutex
	adapters            map[GatewayType]Gateway
	routes              map[GatewayType]ProcessorRoute
	providerPreferences map[string]ProviderPreferences
	defaultPreferences  ProviderPreferences
}

// NewPaymentProcessorRegistry creates a registry.
func NewPaymentProcessorRegistry() *PaymentProcessorRegistry {
	return &PaymentProcessorRegistry{
		adapters:            make(map[GatewayType]Gateway),
		routes:              make(map[GatewayType]ProcessorRoute),
		providerPreferences: make(map[string]ProviderPreferences),
	}
}

// RegisterAdapter registers an adapter and its routing config.
func (r *PaymentProcessorRegistry) RegisterAdapter(adapter Gateway, route ProcessorRoute) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route.Gateway = adapter.Type()
	if !route.Enabled {
		route.Enabled = true
	}

	r.adapters[adapter.Type()] = adapter
	r.routes[adapter.Type()] = route
}

// SetProviderPreferences sets the fallback order for a provider ID.
func (r *PaymentProcessorRegistry) SetProviderPreferences(providerID string, prefs ProviderPreferences) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providerPreferences[strings.ToLower(providerID)] = prefs
}

// SetDefaultPreferences sets default fallback order when provider-specific preferences are absent.
func (r *PaymentProcessorRegistry) SetDefaultPreferences(prefs ProviderPreferences) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultPreferences = prefs
}

// SelectAdapter selects an adapter based on provider preferences and fallback.
func (r *PaymentProcessorRegistry) SelectAdapter(ctx context.Context, providerID, region string, amount Amount) (Gateway, error) {
	r.mu.RLock()
	prefs, ok := r.providerPreferences[strings.ToLower(providerID)]
	if !ok {
		prefs = r.defaultPreferences
	}
	r.mu.RUnlock()

	candidates := []GatewayType{prefs.Primary, prefs.Secondary, prefs.Tertiary}
	for _, gatewayType := range candidates {
		if gatewayType == "" {
			continue
		}
		adapter, ok := r.getAdapter(ctx, gatewayType, region, amount)
		if ok {
			return adapter, nil
		}
	}

	return r.SelectOptimalAdapter(ctx, region, amount)
}

// SelectOptimalAdapter selects the adapter with the lowest fee for a region/currency.
func (r *PaymentProcessorRegistry) SelectOptimalAdapter(ctx context.Context, region string, amount Amount) (Gateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		bestAdapter Gateway
		bestFee     int64
		found       bool
	)

	for gatewayType, adapter := range r.adapters {
		route, ok := r.routes[gatewayType]
		if !ok || !route.Enabled {
			continue
		}
		if !routeSupports(route, region, amount.Currency) {
			continue
		}
		if !adapter.IsHealthy(ctx) {
			continue
		}

		fee := calculateFee(amount, route.Fee)
		if !found || fee < bestFee {
			bestAdapter = adapter
			bestFee = fee
			found = true
		}
	}

	if !found {
		return nil, ErrGatewayUnavailable
	}

	return bestAdapter, nil
}

func (r *PaymentProcessorRegistry) getAdapter(ctx context.Context, gatewayType GatewayType, region string, amount Amount) (Gateway, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[gatewayType]
	if !ok {
		return nil, false
	}

	route, ok := r.routes[gatewayType]
	if !ok || !route.Enabled {
		return nil, false
	}

	if !routeSupports(route, region, amount.Currency) {
		return nil, false
	}

	if !adapter.IsHealthy(ctx) {
		return nil, false
	}

	return adapter, true
}

func routeSupports(route ProcessorRoute, region string, currency Currency) bool {
	if len(route.Regions) > 0 {
		match := false
		for _, r := range route.Regions {
			if strings.EqualFold(strings.TrimSpace(r), strings.TrimSpace(region)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	if len(route.Currencies) > 0 {
		match := false
		for _, c := range route.Currencies {
			if c == currency {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

func calculateFee(amount Amount, fee FeeSchedule) int64 {
	variable := (amount.Value * fee.VariableBps) / 10000
	return fee.FixedFee + variable
}

// ErrRegistryNotConfigured indicates registry is missing configuration.
var ErrRegistryNotConfigured = errors.New("payment registry not configured")
