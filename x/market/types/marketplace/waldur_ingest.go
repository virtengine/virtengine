// Package marketplace provides types for the marketplace on-chain module.
//
// VE-3D: Waldur ingestion types for Waldur-to-chain synchronization.
// This file defines the canonical mapping from Waldur offerings to on-chain offerings.
package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// WaldurOfferingImport represents a Waldur offering to be ingested on-chain.
type WaldurOfferingImport struct {
	// UUID is the Waldur offering UUID.
	UUID string `json:"uuid"`

	// Name is the offering name.
	Name string `json:"name"`

	// Description is the offering description.
	Description string `json:"description"`

	// Type is the Waldur offering type (e.g., "VirtEngine.Compute").
	Type string `json:"type"`

	// State is the Waldur state (Active, Paused, Archived).
	State string `json:"state"`

	// CategoryUUID is the Waldur category UUID.
	CategoryUUID string `json:"category_uuid"`

	// CustomerUUID is the Waldur customer UUID (provider organization).
	CustomerUUID string `json:"customer_uuid"`

	// Shared indicates if the offering is publicly visible.
	Shared bool `json:"shared"`

	// Billable indicates if the offering is billable.
	Billable bool `json:"billable"`

	// BackendID is the on-chain offering ID if already synced (optional).
	BackendID string `json:"backend_id,omitempty"`

	// Attributes contains offering attributes from Waldur.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Components contains pricing components.
	Components []WaldurPricingComponent `json:"components,omitempty"`

	// Created is the Waldur creation timestamp.
	Created time.Time `json:"created"`

	// Modified is the Waldur modification timestamp.
	Modified time.Time `json:"modified"`
}

// IngestConfig configures the ingestion mapping behavior.
type IngestConfig struct {
	// CategoryMap maps Waldur category UUIDs to VirtEngine categories.
	CategoryMap map[string]OfferingCategory `json:"category_map"`

	// TypeMap maps Waldur offering types to VirtEngine categories.
	TypeMap map[string]OfferingCategory `json:"type_map"`

	// RegionMap maps Waldur location UUIDs to VirtEngine region codes.
	RegionMap map[string]string `json:"region_map"`

	// CustomerProviderMap maps Waldur customer UUIDs to provider addresses.
	CustomerProviderMap map[string]string `json:"customer_provider_map"`

	// CurrencyDenominator is the denominator for price conversion (e.g., 1000000 for utoken).
	CurrencyDenominator uint64 `json:"currency_denominator"`

	// DefaultCurrency is the default currency if not specified.
	DefaultCurrency string `json:"default_currency"`

	// MinIdentityScore is the minimum identity score required for imported offerings.
	MinIdentityScore uint32 `json:"min_identity_score"`

	// RequireProviderRegistration requires the provider to be registered on-chain.
	RequireProviderRegistration bool `json:"require_provider_registration"`
}

// DefaultIngestConfig returns sensible defaults for ingestion config.
func DefaultIngestConfig() IngestConfig {
	return IngestConfig{
		CategoryMap: map[string]OfferingCategory{},
		TypeMap: map[string]OfferingCategory{
			"VirtEngine.Compute": OfferingCategoryCompute,
			"VirtEngine.Storage": OfferingCategoryStorage,
			"VirtEngine.Network": OfferingCategoryNetwork,
			"VirtEngine.HPC":     OfferingCategoryHPC,
			"VirtEngine.GPU":     OfferingCategoryGPU,
			"VirtEngine.ML":      OfferingCategoryML,
			"VirtEngine.Generic": OfferingCategoryOther,
		},
		RegionMap:                   map[string]string{},
		CustomerProviderMap:         map[string]string{},
		CurrencyDenominator:         1000000,
		DefaultCurrency:             "uvirt",
		MinIdentityScore:            0,
		RequireProviderRegistration: true,
	}
}

// IngestValidationResult contains the result of validating a Waldur offering for ingestion.
type IngestValidationResult struct {
	// Valid indicates if the offering can be ingested.
	Valid bool `json:"valid"`

	// Errors contains validation errors.
	Errors []string `json:"errors,omitempty"`

	// Warnings contains non-fatal warnings.
	Warnings []string `json:"warnings,omitempty"`

	// ProviderAddress is the resolved provider address.
	ProviderAddress string `json:"provider_address,omitempty"`

	// Category is the resolved category.
	Category OfferingCategory `json:"category,omitempty"`

	// NeedsProviderRegistration indicates if provider needs to register first.
	NeedsProviderRegistration bool `json:"needs_provider_registration,omitempty"`

	// NeedsVEIDVerification indicates if provider needs VEID verification.
	NeedsVEIDVerification bool `json:"needs_veid_verification,omitempty"`
}

// Validate validates a Waldur offering for ingestion.
func (w *WaldurOfferingImport) Validate(cfg IngestConfig) IngestValidationResult {
	result := IngestValidationResult{Valid: true}

	// Validate required fields
	if w.UUID == "" {
		result.Errors = append(result.Errors, "UUID is required")
		result.Valid = false
	}
	if w.Name == "" {
		result.Errors = append(result.Errors, "name is required")
		result.Valid = false
	}
	if w.CustomerUUID == "" {
		result.Errors = append(result.Errors, "customer UUID is required")
		result.Valid = false
	}

	// Validate state
	if w.State != "Active" && w.State != "Paused" && w.State != "Archived" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("unknown state: %s", w.State))
	}

	// Resolve provider address
	if providerAddr, ok := cfg.CustomerProviderMap[w.CustomerUUID]; ok {
		result.ProviderAddress = providerAddr
	} else if cfg.RequireProviderRegistration {
		result.Errors = append(result.Errors, fmt.Sprintf("customer UUID %s not mapped to provider address", w.CustomerUUID))
		result.Valid = false
		result.NeedsProviderRegistration = true
	}

	// Resolve category
	result.Category = w.ResolveCategory(cfg)
	if result.Category == "" {
		result.Category = OfferingCategoryOther
		result.Warnings = append(result.Warnings, "category defaulted to 'other'")
	}

	// Check for VEID requirements
	if cfg.MinIdentityScore > 0 {
		result.NeedsVEIDVerification = true
	}

	return result
}

// ResolveCategory resolves the VirtEngine category from Waldur type/category.
func (w *WaldurOfferingImport) ResolveCategory(cfg IngestConfig) OfferingCategory {
	// First try type mapping
	if cat, ok := cfg.TypeMap[w.Type]; ok {
		return cat
	}

	// Then try category UUID mapping
	if cat, ok := cfg.CategoryMap[w.CategoryUUID]; ok {
		return cat
	}

	// Try to infer from type string
	typeLower := strings.ToLower(w.Type)
	switch {
	case strings.Contains(typeLower, "compute") || strings.Contains(typeLower, "vm"):
		return OfferingCategoryCompute
	case strings.Contains(typeLower, "storage") || strings.Contains(typeLower, "volume"):
		return OfferingCategoryStorage
	case strings.Contains(typeLower, "network"):
		return OfferingCategoryNetwork
	case strings.Contains(typeLower, "hpc") || strings.Contains(typeLower, "slurm"):
		return OfferingCategoryHPC
	case strings.Contains(typeLower, "gpu"):
		return OfferingCategoryGPU
	case strings.Contains(typeLower, "ml") || strings.Contains(typeLower, "machine"):
		return OfferingCategoryML
	}

	return OfferingCategoryOther
}

// ResolveState resolves the VirtEngine offering state from Waldur state.
func (w *WaldurOfferingImport) ResolveState() OfferingState {
	switch w.State {
	case "Active":
		return OfferingStateActive
	case "Paused":
		return OfferingStatePaused
	case "Archived":
		return OfferingStateTerminated
	default:
		return OfferingStatePaused
	}
}

// ResolvePricing extracts pricing information from Waldur components.
func (w *WaldurOfferingImport) ResolvePricing(cfg IngestConfig) PricingInfo {
	pricing := PricingInfo{
		Model:      PricingModelHourly,
		Currency:   cfg.DefaultCurrency,
		UsageRates: make(map[string]uint64),
	}

	for _, comp := range w.Components {
		price := denormalizePrice(comp.Price, cfg.CurrencyDenominator)

		if comp.Name == "base" || comp.Type == "fixed" {
			pricing.BasePrice = price
			pricing.Model = resolvePricingModel(comp.BillingType)
		} else {
			pricing.UsageRates[comp.Name] = price
			if pricing.Model == PricingModelHourly && comp.BillingType == "usage" {
				pricing.Model = PricingModelUsageBased
			}
		}
	}

	return pricing
}

// ExtractRegions extracts region codes from Waldur attributes.
func (w *WaldurOfferingImport) ExtractRegions(cfg IngestConfig) []string {
	var regions []string

	if regionsAttr, ok := w.Attributes["regions"]; ok {
		if regArr, ok := regionsAttr.([]interface{}); ok {
			for _, r := range regArr {
				if rStr, ok := r.(string); ok {
					regions = append(regions, rStr)
				}
			}
		}
	}

	// Map Waldur location UUIDs to chain region codes
	if locations, ok := w.Attributes["locations"]; ok {
		if locArr, ok := locations.([]interface{}); ok {
			for _, loc := range locArr {
				if locStr, ok := loc.(string); ok {
					if region, ok := cfg.RegionMap[locStr]; ok {
						regions = append(regions, region)
					}
				}
			}
		}
	}

	return regions
}

// ExtractTags extracts tags from Waldur attributes.
func (w *WaldurOfferingImport) ExtractTags() []string {
	var tags []string

	if tagsAttr, ok := w.Attributes["tags"]; ok {
		if tagArr, ok := tagsAttr.([]interface{}); ok {
			for _, t := range tagArr {
				if tStr, ok := t.(string); ok {
					tags = append(tags, tStr)
				}
			}
		}
	}

	return tags
}

// ExtractSpecifications extracts specifications from Waldur attributes.
func (w *WaldurOfferingImport) ExtractSpecifications() map[string]string {
	specs := make(map[string]string)

	for key, value := range w.Attributes {
		if strings.HasPrefix(key, "spec_") {
			specKey := strings.TrimPrefix(key, "spec_")
			if strVal, ok := value.(string); ok {
				specs[specKey] = strVal
			} else {
				specs[specKey] = fmt.Sprintf("%v", value)
			}
		}
	}

	return specs
}

// ExtractPublicMetadata extracts public metadata from Waldur attributes.
func (w *WaldurOfferingImport) ExtractPublicMetadata() map[string]string {
	meta := make(map[string]string)

	for key, value := range w.Attributes {
		if strings.HasPrefix(key, "ve_") && !isReservedAttribute(key) {
			metaKey := strings.TrimPrefix(key, "ve_")
			if strVal, ok := value.(string); ok {
				meta[metaKey] = strVal
			} else {
				meta[metaKey] = fmt.Sprintf("%v", value)
			}
		}
	}

	return meta
}

// ExtractIdentityRequirements extracts identity requirements from Waldur attributes.
func (w *WaldurOfferingImport) ExtractIdentityRequirements() IdentityRequirement {
	req := DefaultIdentityRequirement()

	if minScore, ok := w.Attributes["ve_min_identity_score"]; ok {
		switch v := minScore.(type) {
		case float64:
			req.MinScore = uint32(v)
		case int:
			req.MinScore = uint32(v)
		}
	}

	if requireMFA, ok := w.Attributes["ve_require_mfa"]; ok {
		if mfa, ok := requireMFA.(bool); ok {
			req.RequireMFA = mfa
		}
	}

	return req
}

// ToOffering converts a Waldur offering import to an on-chain Offering.
func (w *WaldurOfferingImport) ToOffering(providerAddr string, sequence uint64, cfg IngestConfig) *Offering {
	return w.ToOfferingAt(providerAddr, sequence, cfg, time.Now())
}

// ToOfferingAt converts a Waldur offering import to an on-chain Offering with a specific timestamp.
func (w *WaldurOfferingImport) ToOfferingAt(providerAddr string, sequence uint64, cfg IngestConfig, now time.Time) *Offering {
	id := OfferingID{
		ProviderAddress: providerAddr,
		Sequence:        sequence,
	}

	category := w.ResolveCategory(cfg)
	pricing := w.ResolvePricing(cfg)
	createdAt := now.UTC()

	offering := &Offering{
		ID:                  id,
		State:               w.ResolveState(),
		Category:            category,
		Name:                w.Name,
		Description:         w.Description,
		Version:             "1.0.0",
		Pricing:             pricing,
		IdentityRequirement: w.ExtractIdentityRequirements(),
		PublicMetadata:      w.ExtractPublicMetadata(),
		Specifications:      w.ExtractSpecifications(),
		Tags:                w.ExtractTags(),
		Regions:             w.ExtractRegions(cfg),
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}

	// Extract max concurrent orders
	if maxOrders, ok := w.Attributes["ve_max_concurrent_orders"]; ok {
		switch v := maxOrders.(type) {
		case float64:
			offering.MaxConcurrentOrders = uint32(v)
		case int:
			offering.MaxConcurrentOrders = uint32(v)
		}
	}

	// Extract MFA requirement
	if requireMFA, ok := w.Attributes["ve_require_mfa"]; ok {
		if mfa, ok := requireMFA.(bool); ok {
			offering.RequireMFAForOrders = mfa
		}
	}

	// Set activated timestamp if active
	if offering.State == OfferingStateActive {
		offering.ActivatedAt = &createdAt
	}

	return offering
}

// IngestChecksum computes a deterministic checksum for drift detection.
func (w *WaldurOfferingImport) IngestChecksum() string {
	h := sha256.New()

	h.Write([]byte(w.UUID))
	h.Write([]byte(w.Name))
	h.Write([]byte(w.Description))
	h.Write([]byte(w.Type))
	h.Write([]byte(w.State))
	h.Write([]byte(w.CategoryUUID))
	h.Write([]byte(w.CustomerUUID))
	h.Write([]byte(fmt.Sprintf("%t", w.Shared)))
	h.Write([]byte(fmt.Sprintf("%t", w.Billable)))
	h.Write([]byte(fmt.Sprintf("%d", w.Modified.Unix())))

	for _, comp := range w.Components {
		h.Write([]byte(comp.Name))
		h.Write([]byte(comp.Price))
		h.Write([]byte(comp.Type))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// IngestSyncRecord tracks the ingestion state of a Waldur offering.
type IngestSyncRecord struct {
	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid"`

	// ChainOfferingID is the on-chain offering ID (if created).
	ChainOfferingID string `json:"chain_offering_id,omitempty"`

	// State is the current ingestion state.
	State IngestState `json:"state"`

	// WaldurVersion is a hash of the Waldur data at last ingest.
	WaldurVersion string `json:"waldur_version"`

	// ChainVersion is the on-chain offering version.
	ChainVersion uint64 `json:"chain_version"`

	// LastIngestedAt is when the offering was last ingested.
	LastIngestedAt *time.Time `json:"last_ingested_at,omitempty"`

	// LastAttemptAt is when ingestion was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastError is the most recent error message.
	LastError string `json:"last_error,omitempty"`

	// RetryCount is the number of consecutive retry attempts.
	RetryCount int `json:"retry_count"`

	// NextRetryAt is when the next retry should be attempted.
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// ProviderAddress is the resolved provider address.
	ProviderAddress string `json:"provider_address"`

	// CreatedAt is when this record was first created.
	CreatedAt time.Time `json:"created_at"`
}

// IngestState represents the ingestion state of a Waldur offering.
type IngestState string

const (
	// IngestStatePending indicates ingestion is pending.
	IngestStatePending IngestState = "pending"

	// IngestStateIngested indicates offering is ingested on-chain.
	IngestStateIngested IngestState = "ingested"

	// IngestStateFailed indicates last ingestion attempt failed.
	IngestStateFailed IngestState = "failed"

	// IngestStateRetrying indicates ingestion is being retried.
	IngestStateRetrying IngestState = "retrying"

	// IngestStateDeadLettered indicates ingestion failed permanently.
	IngestStateDeadLettered IngestState = "dead_lettered"

	// IngestStateOutOfSync indicates Waldur data changed since last ingest.
	IngestStateOutOfSync IngestState = "out_of_sync"

	// IngestStateSkipped indicates offering was intentionally skipped.
	IngestStateSkipped IngestState = "skipped"

	// IngestStateDeprecated indicates the Waldur offering was archived.
	IngestStateDeprecated IngestState = "deprecated"
)

// IngestResult represents the result of an ingestion operation.
type IngestResult struct {
	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid"`

	// ChainOfferingID is the on-chain offering ID (on success).
	ChainOfferingID string `json:"chain_offering_id,omitempty"`

	// Action is the ingestion action performed (create, update, deprecate).
	Action IngestAction `json:"action"`

	// Success indicates if the ingestion succeeded.
	Success bool `json:"success"`

	// Error is the error message on failure.
	Error string `json:"error,omitempty"`

	// Checksum is the data checksum after ingestion.
	Checksum string `json:"checksum,omitempty"`

	// Timestamp is when the ingestion completed.
	Timestamp time.Time `json:"timestamp"`

	// RetryCount is the number of retries attempted.
	RetryCount int `json:"retry_count"`

	// Duration is how long the ingestion took.
	Duration time.Duration `json:"duration_ns"`
}

// IngestAction represents the ingestion action to perform.
type IngestAction string

const (
	// IngestActionCreate creates a new on-chain offering.
	IngestActionCreate IngestAction = "create"

	// IngestActionUpdate updates an existing on-chain offering.
	IngestActionUpdate IngestAction = "update"

	// IngestActionDeprecate deprecates an on-chain offering.
	IngestActionDeprecate IngestAction = "deprecate"

	// IngestActionSkip skips ingestion (validation failed).
	IngestActionSkip IngestAction = "skip"
)

// IngestMetrics tracks ingestion operation metrics.
type IngestMetrics struct {
	// TotalIngests is the total ingestion attempts.
	TotalIngests int64 `json:"total_ingests"`

	// SuccessfulIngests is the count of successful ingestions.
	SuccessfulIngests int64 `json:"successful_ingests"`

	// FailedIngests is the count of failed ingestion attempts.
	FailedIngests int64 `json:"failed_ingests"`

	// DeadLetteredIngests is the count of dead-lettered ingestions.
	DeadLetteredIngests int64 `json:"dead_lettered_ingests"`

	// SkippedIngests is the count of skipped ingestions.
	SkippedIngests int64 `json:"skipped_ingests"`

	// DriftDetections is the count of drift detections.
	DriftDetections int64 `json:"drift_detections"`

	// ReconciliationsRun is the count of reconciliation runs.
	ReconciliationsRun int64 `json:"reconciliations_run"`

	// LastIngestTime is the timestamp of the last ingestion attempt.
	LastIngestTime time.Time `json:"last_ingest_time"`

	// LastSuccessTime is the timestamp of the last successful ingestion.
	LastSuccessTime time.Time `json:"last_success_time"`

	// LastReconcileTime is the timestamp of the last reconciliation.
	LastReconcileTime time.Time `json:"last_reconcile_time"`

	// AverageIngestDurationMs is the average ingestion duration in milliseconds.
	AverageIngestDurationMs float64 `json:"average_ingest_duration_ms"`

	// OfferingsIngested is the count of offerings successfully ingested.
	OfferingsIngested int64 `json:"offerings_ingested"`

	// OfferingsUpdated is the count of offerings updated.
	OfferingsUpdated int64 `json:"offerings_updated"`

	// OfferingsDeprecated is the count of offerings deprecated.
	OfferingsDeprecated int64 `json:"offerings_deprecated"`
}

// Helper functions

// denormalizePrice converts a decimal string price to smallest token denomination.
func denormalizePrice(priceStr string, denominator uint64) uint64 {
	if denominator == 0 {
		denominator = 1000000
	}

	// Parse the price string
	priceStr = strings.TrimSpace(priceStr)
	if priceStr == "" {
		return 0
	}

	// Handle integer prices
	if !strings.Contains(priceStr, ".") {
		intVal, err := strconv.ParseUint(priceStr, 10, 64)
		if err != nil {
			return 0
		}
		return intVal * denominator
	}

	// Split on decimal point
	parts := strings.Split(priceStr, ".")
	if len(parts) != 2 {
		return 0
	}

	intPart, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0
	}

	// Pad or truncate fractional part to 6 digits
	fracStr := parts[1]
	for len(fracStr) < 6 {
		fracStr += "0"
	}
	if len(fracStr) > 6 {
		fracStr = fracStr[:6]
	}

	fracPart, err := strconv.ParseUint(fracStr, 10, 64)
	if err != nil {
		return 0
	}

	return intPart*denominator + fracPart
}

// resolvePricingModel maps Waldur billing type to VirtEngine pricing model.
func resolvePricingModel(billingType string) PricingModel {
	switch strings.ToLower(billingType) {
	case "usage":
		return PricingModelHourly
	case "monthly":
		return PricingModelMonthly
	case "fixed", "one":
		return PricingModelFixed
	default:
		return PricingModelHourly
	}
}

// isReservedAttribute returns true if the attribute key is reserved.
func isReservedAttribute(key string) bool {
	reserved := map[string]bool{
		"ve_offering_id":          true,
		"ve_provider":             true,
		"ve_category":             true,
		"ve_version":              true,
		"ve_min_identity_score":   true,
		"ve_require_mfa":          true,
		"ve_max_concurrent_orders": true,
	}
	return reserved[key]
}

// Ingestion event types (extend the events in events.go)
const (
	// EventOfferingIngested is emitted when an offering is ingested from Waldur.
	EventOfferingIngested MarketplaceEventType = "offering_ingested"

	// EventOfferingIngestFailed is emitted when an offering ingestion fails.
	EventOfferingIngestFailed MarketplaceEventType = "offering_ingest_failed"
)
