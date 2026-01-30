// Package marketplace provides types for the marketplace on-chain module.
//
// VE-2D: Offering sync types for automatic chain-to-Waldur synchronization.
// This file defines the canonical mapping from on-chain offerings to Waldur offerings.
package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// WaldurOfferingType maps VirtEngine offering categories to Waldur offering types.
var WaldurOfferingType = map[OfferingCategory]string{
	OfferingCategoryCompute: "VirtEngine.Compute",
	OfferingCategoryStorage: "VirtEngine.Storage",
	OfferingCategoryNetwork: "VirtEngine.Network",
	OfferingCategoryHPC:     "VirtEngine.HPC",
	OfferingCategoryGPU:     "VirtEngine.GPU",
	OfferingCategoryML:      "VirtEngine.ML",
	OfferingCategoryOther:   "VirtEngine.Generic",
}

// WaldurOfferingState maps VirtEngine offering states to Waldur states.
var WaldurOfferingState = map[OfferingState]string{
	OfferingStateActive:     "Active",
	OfferingStatePaused:     "Paused",
	OfferingStateSuspended:  "Archived",
	OfferingStateDeprecated: "Paused",
	OfferingStateTerminated: "Archived",
}

// WaldurOfferingCreate contains parameters for creating a Waldur offering.
type WaldurOfferingCreate struct {
	// Name is the offering name (max 255 chars).
	Name string `json:"name"`

	// Description is the offering description (markdown supported).
	Description string `json:"description,omitempty"`

	// Type is the Waldur offering type (e.g., "VirtEngine.Compute").
	Type string `json:"type"`

	// State is the initial state (Active, Paused, Archived).
	State string `json:"state"`

	// CategoryUUID is the Waldur category UUID.
	CategoryUUID string `json:"category_uuid,omitempty"`

	// CustomerUUID is the Waldur customer/organization UUID (provider).
	CustomerUUID string `json:"customer_uuid"`

	// Shared indicates if the offering is publicly visible.
	Shared bool `json:"shared"`

	// Billable indicates if the offering is billable.
	Billable bool `json:"billable"`

	// BackendID is the on-chain offering ID for cross-reference.
	BackendID string `json:"backend_id"`

	// Attributes contains additional offering attributes.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Components contains pricing components.
	Components []WaldurPricingComponent `json:"components,omitempty"`
}

// WaldurOfferingUpdate contains parameters for updating a Waldur offering.
type WaldurOfferingUpdate struct {
	// Name is the updated offering name.
	Name string `json:"name,omitempty"`

	// Description is the updated description.
	Description string `json:"description,omitempty"`

	// State is the updated state.
	State string `json:"state,omitempty"`

	// Attributes contains updated attributes.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Components contains updated pricing components.
	Components []WaldurPricingComponent `json:"components,omitempty"`
}

// WaldurPricingComponent represents a Waldur pricing component.
type WaldurPricingComponent struct {
	// Type is the component type (usage, fixed, one_time).
	Type string `json:"type"`

	// Name is the component name.
	Name string `json:"name"`

	// MeasuredUnit is the unit of measurement.
	MeasuredUnit string `json:"measured_unit,omitempty"`

	// BillingType is how billing is calculated.
	BillingType string `json:"billing_type,omitempty"`

	// Price is the price per unit.
	Price string `json:"price"`

	// MinValue is the minimum value.
	MinValue int64 `json:"min_value,omitempty"`

	// MaxValue is the maximum value.
	MaxValue int64 `json:"max_value,omitempty"`
}

// OfferingSyncConfig configures offering synchronization behavior.
type OfferingSyncConfig struct {
	// Enabled toggles automatic offering sync.
	Enabled bool `json:"enabled"`

	// SyncInterval is how often to check for drift (seconds).
	SyncIntervalSeconds int64 `json:"sync_interval_seconds"`

	// ReconcileOnStartup triggers full reconciliation on worker start.
	ReconcileOnStartup bool `json:"reconcile_on_startup"`

	// MaxRetries is the maximum sync retry attempts before dead-lettering.
	MaxRetries int `json:"max_retries"`

	// RetryBackoffSeconds is the base backoff between retries.
	RetryBackoffSeconds int64 `json:"retry_backoff_seconds"`

	// MaxBackoffSeconds is the maximum backoff time.
	MaxBackoffSeconds int64 `json:"max_backoff_seconds"`

	// WaldurCustomerUUID is the Waldur customer UUID for offerings.
	WaldurCustomerUUID string `json:"waldur_customer_uuid"`

	// WaldurCategoryMap maps offering categories to Waldur category UUIDs.
	WaldurCategoryMap map[string]string `json:"waldur_category_map,omitempty"`

	// DefaultRegionMap maps VirtEngine region codes to Waldur location UUIDs.
	DefaultRegionMap map[string]string `json:"default_region_map,omitempty"`

	// CurrencyDenominator is the denominator for price conversion (e.g., 1000000 for utoken).
	CurrencyDenominator uint64 `json:"currency_denominator"`
}

// DefaultOfferingSyncConfig returns sensible defaults for offering sync.
func DefaultOfferingSyncConfig() OfferingSyncConfig {
	return OfferingSyncConfig{
		Enabled:             false,
		SyncIntervalSeconds: 300, // 5 minutes
		ReconcileOnStartup:  true,
		MaxRetries:          5,
		RetryBackoffSeconds: 30,
		MaxBackoffSeconds:   3600, // 1 hour
		CurrencyDenominator: 1000000,
		WaldurCategoryMap:   make(map[string]string),
		DefaultRegionMap:    make(map[string]string),
	}
}

// OfferingSyncResult represents the result of an offering sync operation.
type OfferingSyncResult struct {
	// OfferingID is the on-chain offering ID.
	OfferingID string `json:"offering_id"`

	// WaldurUUID is the Waldur offering UUID (on success).
	WaldurUUID string `json:"waldur_uuid,omitempty"`

	// Action is the sync action performed (create, update, disable).
	Action string `json:"action"`

	// Success indicates if the sync succeeded.
	Success bool `json:"success"`

	// Error is the error message on failure.
	Error string `json:"error,omitempty"`

	// Checksum is the data checksum after sync.
	Checksum string `json:"checksum,omitempty"`

	// Version is the on-chain version that was synced.
	Version uint64 `json:"version"`

	// Timestamp is when the sync completed.
	Timestamp time.Time `json:"timestamp"`

	// RetryCount is the number of retries attempted.
	RetryCount int `json:"retry_count"`
}

// OfferingSyncMetrics tracks sync operation metrics.
type OfferingSyncMetrics struct {
	// TotalSyncs is the total sync attempts.
	TotalSyncs int64 `json:"total_syncs"`

	// SuccessfulSyncs is the count of successful syncs.
	SuccessfulSyncs int64 `json:"successful_syncs"`

	// FailedSyncs is the count of failed syncs.
	FailedSyncs int64 `json:"failed_syncs"`

	// DeadLetteredSyncs is the count of dead-lettered syncs.
	DeadLetteredSyncs int64 `json:"dead_lettered_syncs"`

	// DriftDetections is the count of drift detections.
	DriftDetections int64 `json:"drift_detections"`

	// ReconciliationsRun is the count of reconciliation runs.
	ReconciliationsRun int64 `json:"reconciliations_run"`

	// LastSyncTime is the timestamp of the last sync attempt.
	LastSyncTime time.Time `json:"last_sync_time"`

	// LastSuccessTime is the timestamp of the last successful sync.
	LastSuccessTime time.Time `json:"last_success_time"`

	// LastReconcileTime is the timestamp of the last reconciliation.
	LastReconcileTime time.Time `json:"last_reconcile_time"`

	// AverageSyncDurationMs is the average sync duration in milliseconds.
	AverageSyncDurationMs float64 `json:"average_sync_duration_ms"`
}

// ToWaldurCreate converts an on-chain offering to Waldur creation parameters.
func (o *Offering) ToWaldurCreate(cfg OfferingSyncConfig) WaldurOfferingCreate {
	// Truncate name to Waldur's 255 char limit
	name := o.Name
	if len(name) > 255 {
		name = name[:252] + "..."
	}

	// Map category to Waldur type
	offeringType := WaldurOfferingType[o.Category]
	if offeringType == "" {
		offeringType = "VirtEngine.Generic"
	}

	// Map state
	state := WaldurOfferingState[o.State]
	if state == "" {
		state = "Paused"
	}

	// Build attributes
	attrs := make(map[string]interface{})

	// Core VirtEngine attributes
	attrs["ve_offering_id"] = o.ID.String()
	attrs["ve_provider"] = o.ID.ProviderAddress
	attrs["ve_category"] = string(o.Category)
	attrs["ve_version"] = o.Version
	attrs["ve_min_identity_score"] = o.IdentityRequirement.MinScore
	attrs["ve_require_mfa"] = o.RequireMFAForOrders
	attrs["ve_max_concurrent_orders"] = o.MaxConcurrentOrders

	// Tags
	if len(o.Tags) > 0 {
		attrs["tags"] = o.Tags
	}

	// Regions
	if len(o.Regions) > 0 {
		attrs["regions"] = o.Regions
	}

	// Specifications
	for k, v := range o.Specifications {
		attrs["spec_"+k] = v
	}

	// Public metadata with ve_ prefix
	for k, v := range o.PublicMetadata {
		attrs["ve_"+k] = v
	}

	// Build pricing components
	var components []WaldurPricingComponent
	components = append(components, WaldurPricingComponent{
		Type:         "usage",
		Name:         "base",
		MeasuredUnit: string(o.Pricing.Model),
		BillingType:  mapPricingModel(o.Pricing.Model),
		Price:        normalizePrice(o.Pricing.BasePrice, cfg.CurrencyDenominator),
	})

	// Add usage rate components
	for name, rate := range o.Pricing.UsageRates {
		components = append(components, WaldurPricingComponent{
			Type:         "usage",
			Name:         name,
			MeasuredUnit: name,
			BillingType:  "usage",
			Price:        normalizePrice(rate, cfg.CurrencyDenominator),
		})
	}

	// Resolve category UUID
	categoryUUID := cfg.WaldurCategoryMap[string(o.Category)]

	return WaldurOfferingCreate{
		Name:         name,
		Description:  o.Description,
		Type:         offeringType,
		State:        state,
		CategoryUUID: categoryUUID,
		CustomerUUID: cfg.WaldurCustomerUUID,
		Shared:       true, // VirtEngine offerings are public marketplace offerings
		Billable:     true,
		BackendID:    o.ID.String(),
		Attributes:   attrs,
		Components:   components,
	}
}

// ToWaldurUpdate converts offering changes to Waldur update parameters.
func (o *Offering) ToWaldurUpdate(cfg OfferingSyncConfig) WaldurOfferingUpdate {
	// Truncate name
	name := o.Name
	if len(name) > 255 {
		name = name[:252] + "..."
	}

	// Map state
	state := WaldurOfferingState[o.State]
	if state == "" {
		state = "Paused"
	}

	// Build attributes (same as create)
	attrs := make(map[string]interface{})
	attrs["ve_offering_id"] = o.ID.String()
	attrs["ve_provider"] = o.ID.ProviderAddress
	attrs["ve_category"] = string(o.Category)
	attrs["ve_version"] = o.Version
	attrs["ve_min_identity_score"] = o.IdentityRequirement.MinScore
	attrs["ve_require_mfa"] = o.RequireMFAForOrders
	attrs["ve_max_concurrent_orders"] = o.MaxConcurrentOrders

	if len(o.Tags) > 0 {
		attrs["tags"] = o.Tags
	}
	if len(o.Regions) > 0 {
		attrs["regions"] = o.Regions
	}
	for k, v := range o.Specifications {
		attrs["spec_"+k] = v
	}
	for k, v := range o.PublicMetadata {
		attrs["ve_"+k] = v
	}

	// Build pricing components
	var components []WaldurPricingComponent
	components = append(components, WaldurPricingComponent{
		Type:         "usage",
		Name:         "base",
		MeasuredUnit: string(o.Pricing.Model),
		BillingType:  mapPricingModel(o.Pricing.Model),
		Price:        normalizePrice(o.Pricing.BasePrice, cfg.CurrencyDenominator),
	})
	for name, rate := range o.Pricing.UsageRates {
		components = append(components, WaldurPricingComponent{
			Type:         "usage",
			Name:         name,
			MeasuredUnit: name,
			BillingType:  "usage",
			Price:        normalizePrice(rate, cfg.CurrencyDenominator),
		})
	}

	return WaldurOfferingUpdate{
		Name:        name,
		Description: o.Description,
		State:       state,
		Attributes:  attrs,
		Components:  components,
	}
}

// SyncChecksum computes a deterministic checksum of the offering for drift detection.
func (o *Offering) SyncChecksum() string {
	h := sha256.New()

	// Include all sync-relevant fields in a deterministic order
	h.Write([]byte(o.ID.String()))
	h.Write([]byte(o.Name))
	h.Write([]byte(o.Description))
	h.Write([]byte(o.Category))
	h.Write([]byte(o.Version))
	h.Write([]byte(fmt.Sprintf("%d", o.State)))
	h.Write([]byte(o.Pricing.Model))
	h.Write([]byte(fmt.Sprintf("%d", o.Pricing.BasePrice)))
	h.Write([]byte(o.Pricing.Currency))
	h.Write([]byte(fmt.Sprintf("%d", o.IdentityRequirement.MinScore)))
	h.Write([]byte(fmt.Sprintf("%t", o.RequireMFAForOrders)))
	h.Write([]byte(fmt.Sprintf("%d", o.MaxConcurrentOrders)))
	h.Write([]byte(fmt.Sprintf("%d", o.UpdatedAt.Unix())))

	// Include tags in sorted order
	for _, tag := range o.Tags {
		h.Write([]byte(tag))
	}

	// Include regions in sorted order
	for _, region := range o.Regions {
		h.Write([]byte(region))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// mapPricingModel converts VirtEngine pricing model to Waldur billing type.
func mapPricingModel(model PricingModel) string {
	switch model {
	case PricingModelHourly:
		return "usage"
	case PricingModelDaily:
		return "usage"
	case PricingModelMonthly:
		return "monthly"
	case PricingModelUsageBased:
		return "usage"
	case PricingModelFixed:
		return "fixed"
	default:
		return "usage"
	}
}

// normalizePrice converts smallest token denomination to decimal string.
func normalizePrice(amount uint64, denominator uint64) string {
	if denominator == 0 {
		denominator = 1000000 // default: micro-units
	}
	intPart := amount / denominator
	fracPart := amount % denominator
	return fmt.Sprintf("%d.%06d", intPart, fracPart)
}

// OfferingEventData contains data for offering sync events.
type OfferingEventData struct {
	// OfferingID is the on-chain offering ID.
	OfferingID string `json:"offering_id"`

	// ProviderAddress is the provider's address.
	ProviderAddress string `json:"provider_address"`

	// EventType is the event type (created, updated, terminated).
	EventType string `json:"event_type"`

	// Version is the current on-chain version.
	Version uint64 `json:"version"`

	// Checksum is the offering data checksum.
	Checksum string `json:"checksum"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// DeadLetterEntry represents a failed sync that exceeded retry limits.
type DeadLetterEntry struct {
	// OfferingID is the offering that failed to sync.
	OfferingID string `json:"offering_id"`

	// Action is the action that failed (create, update, disable).
	Action string `json:"action"`

	// LastError is the last error encountered.
	LastError string `json:"last_error"`

	// RetryCount is the number of retries attempted.
	RetryCount int `json:"retry_count"`

	// FirstAttempt is when the first sync was attempted.
	FirstAttempt time.Time `json:"first_attempt"`

	// LastAttempt is when the last sync was attempted.
	LastAttempt time.Time `json:"last_attempt"`

	// OfferingChecksum is the checksum at time of failure.
	OfferingChecksum string `json:"offering_checksum"`
}
