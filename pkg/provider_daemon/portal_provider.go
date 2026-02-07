package provider_daemon

import (
	"context"
	"net/http"
)

// ProviderInfoProvider supplies provider info for portal endpoints.
type ProviderInfoProvider interface {
	Info(ctx context.Context) ProviderInfo
	Pricing(ctx context.Context) ProviderPricing
	Capacity(ctx context.Context) ProviderCapacity
	Attributes(ctx context.Context) ProviderAttributes
}

// StaticProviderInfoProvider returns static values.
type StaticProviderInfoProvider struct {
	info       ProviderInfo
	pricing    ProviderPricing
	capacity   ProviderCapacity
	attributes ProviderAttributes
}

// NewStaticProviderInfoProvider creates a static provider info provider.
func NewStaticProviderInfoProvider(info ProviderInfo, pricing ProviderPricing, capacity ProviderCapacity, attributes ProviderAttributes) *StaticProviderInfoProvider {
	return &StaticProviderInfoProvider{
		info:       info,
		pricing:    pricing,
		capacity:   capacity,
		attributes: attributes,
	}
}

// Info returns provider info.
func (p *StaticProviderInfoProvider) Info(_ context.Context) ProviderInfo {
	return p.info
}

// Pricing returns provider pricing.
func (p *StaticProviderInfoProvider) Pricing(_ context.Context) ProviderPricing {
	return p.pricing
}

// Capacity returns provider capacity.
func (p *StaticProviderInfoProvider) Capacity(_ context.Context) ProviderCapacity {
	return p.capacity
}

// Attributes returns provider attributes.
func (p *StaticProviderInfoProvider) Attributes(_ context.Context) ProviderAttributes {
	return p.attributes
}

func (s *PortalAPIServer) handleProviderInfo(w http.ResponseWriter, r *http.Request) {
	info := s.providerInfo.Info(r.Context())
	writeJSON(w, http.StatusOK, info)
}

func (s *PortalAPIServer) handleProviderPricing(w http.ResponseWriter, r *http.Request) {
	pricing := s.providerInfo.Pricing(r.Context())
	writeJSON(w, http.StatusOK, pricing)
}

func (s *PortalAPIServer) handleProviderCapacity(w http.ResponseWriter, r *http.Request) {
	capacity := s.providerInfo.Capacity(r.Context())
	writeJSON(w, http.StatusOK, capacity)
}

func (s *PortalAPIServer) handleProviderAttributes(w http.ResponseWriter, r *http.Request) {
	attrs := s.providerInfo.Attributes(r.Context())
	writeJSON(w, http.StatusOK, attrs)
}
