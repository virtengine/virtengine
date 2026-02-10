// Package marketplace provides types for the marketplace on-chain module.
//
// VE-300: Marketplace on-chain data model: offerings, orders, allocations, and states
// This file defines the Offering type with provider metadata, pricing, identity requirements,
// and encrypted provider secrets.
package marketplace

import (
	"crypto/sha256"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// OfferingState represents the lifecycle state of an offering
type OfferingState uint8

const (
	// OfferingStateUnspecified represents an unspecified offering state
	OfferingStateUnspecified OfferingState = 0

	// OfferingStateActive indicates the offering is active and available for orders
	OfferingStateActive OfferingState = 1

	// OfferingStatePaused indicates the offering is temporarily paused
	OfferingStatePaused OfferingState = 2

	// OfferingStateSuspended indicates the offering is suspended by admin/moderator
	OfferingStateSuspended OfferingState = 3

	// OfferingStateDeprecated indicates the offering is deprecated (no new orders)
	OfferingStateDeprecated OfferingState = 4

	// OfferingStateTerminated indicates the offering is permanently terminated
	OfferingStateTerminated OfferingState = 5
)

// OfferingStateNames maps offering states to human-readable names
var OfferingStateNames = map[OfferingState]string{
	OfferingStateUnspecified: "unspecified",
	OfferingStateActive:      "active",
	OfferingStatePaused:      "paused",
	OfferingStateSuspended:   "suspended",
	OfferingStateDeprecated:  "deprecated",
	OfferingStateTerminated:  "terminated",
}

// String returns the string representation of an OfferingState
func (s OfferingState) String() string {
	if name, ok := OfferingStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the offering state is valid
func (s OfferingState) IsValid() bool {
	return s >= OfferingStateActive && s <= OfferingStateTerminated
}

// IsAcceptingOrders returns true if the offering can accept new orders
func (s OfferingState) IsAcceptingOrders() bool {
	return s == OfferingStateActive
}

// OfferingCategory represents the category of an offering
type OfferingCategory string

const (
	// OfferingCategoryCompute represents compute/VM offerings
	OfferingCategoryCompute OfferingCategory = "compute"

	// OfferingCategoryStorage represents storage offerings
	OfferingCategoryStorage OfferingCategory = "storage"

	// OfferingCategoryNetwork represents network offerings
	OfferingCategoryNetwork OfferingCategory = "network"

	// OfferingCategoryHPC represents high-performance computing offerings
	OfferingCategoryHPC OfferingCategory = "hpc"

	// OfferingCategoryGPU represents GPU compute offerings
	OfferingCategoryGPU OfferingCategory = "gpu"

	// OfferingCategoryML represents machine learning offerings
	OfferingCategoryML OfferingCategory = "ml"

	// OfferingCategoryOther represents other/custom offerings
	OfferingCategoryOther OfferingCategory = "other"
)

// PricingModel represents how an offering is priced
type PricingModel string

const (
	// PricingModelHourly represents hourly pricing
	PricingModelHourly PricingModel = "hourly"

	// PricingModelDaily represents daily pricing
	PricingModelDaily PricingModel = "daily"

	// PricingModelMonthly represents monthly pricing
	PricingModelMonthly PricingModel = "monthly"

	// PricingModelUsageBased represents usage-based pricing
	PricingModelUsageBased PricingModel = "usage_based"

	// PricingModelFixed represents fixed/one-time pricing
	PricingModelFixed PricingModel = "fixed"
)

// PriceComponentResourceType represents known resource types for component pricing.
type PriceComponentResourceType string

const (
	// PriceComponentCPU represents CPU pricing.
	PriceComponentCPU PriceComponentResourceType = "cpu"

	// PriceComponentRAM represents memory pricing.
	PriceComponentRAM PriceComponentResourceType = "ram"

	// PriceComponentStorage represents storage pricing.
	PriceComponentStorage PriceComponentResourceType = "storage"

	// PriceComponentGPU represents GPU pricing.
	PriceComponentGPU PriceComponentResourceType = "gpu"

	// PriceComponentNetwork represents network pricing.
	PriceComponentNetwork PriceComponentResourceType = "network"
)

// IdentityRequirement defines the identity verification requirements for an offering
type IdentityRequirement struct {
	// MinScore is the minimum VEID identity score required (0-100)
	MinScore uint32 `json:"min_score"`

	// RequiredStatus is the minimum identity status required
	RequiredStatus string `json:"required_status"`

	// RequireVerifiedEmail indicates if email verification is required
	RequireVerifiedEmail bool `json:"require_verified_email"`

	// RequireVerifiedDomain indicates if domain verification is required (for providers)
	RequireVerifiedDomain bool `json:"require_verified_domain"`

	// RequireMFA indicates if MFA must be enabled for orders
	RequireMFA bool `json:"require_mfa"`
}

// DefaultIdentityRequirement returns the default identity requirement
func DefaultIdentityRequirement() IdentityRequirement {
	return IdentityRequirement{
		MinScore:              0,
		RequiredStatus:        "",
		RequireVerifiedEmail:  false,
		RequireVerifiedDomain: false,
		RequireMFA:            false,
	}
}

// Validate validates the identity requirement
func (r *IdentityRequirement) Validate() error {
	if r.MinScore > 100 {
		return fmt.Errorf("min_score cannot exceed 100: got %d", r.MinScore)
	}
	return nil
}

// IsSatisfiedBy checks if an account meets the identity requirements
func (r *IdentityRequirement) IsSatisfiedBy(score uint32, status string, emailVerified, domainVerified, mfaEnabled bool) bool {
	if score < r.MinScore {
		return false
	}
	if r.RequiredStatus != "" && status != r.RequiredStatus {
		return false
	}
	if r.RequireVerifiedEmail && !emailVerified {
		return false
	}
	if r.RequireVerifiedDomain && !domainVerified {
		return false
	}
	if r.RequireMFA && !mfaEnabled {
		return false
	}
	return true
}

// PricingInfo defines the pricing structure for an offering
type PricingInfo struct {
	// Model is the pricing model
	Model PricingModel `json:"model"`

	// BasePrice is the base price in the smallest token denomination
	BasePrice uint64 `json:"base_price"`

	// Currency is the token denomination
	Currency string `json:"currency"`

	// UsageRates contains usage-based pricing rates (for usage_based model)
	UsageRates map[string]uint64 `json:"usage_rates,omitempty"`

	// MinimumCommitment is the minimum commitment period in seconds (0 = none)
	MinimumCommitment int64 `json:"minimum_commitment,omitempty"`
}

// Validate validates the pricing info
func (p *PricingInfo) Validate() error {
	if p.Model == "" {
		return fmt.Errorf("pricing model is required")
	}
	if p.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}

// PriceComponent represents component-based pricing for an offering.
type PriceComponent struct {
	// ResourceType is the resource type (cpu, ram, storage, gpu, network).
	ResourceType PriceComponentResourceType `json:"resource_type"`

	// Unit is the unit of the component (vcpu, gb, hour, month).
	Unit string `json:"unit"`

	// Price is the unit price in chain currency.
	Price sdk.Coin `json:"price"`

	// USDReference is the USD price at time of creation (display only).
	USDReference string `json:"usd_reference,omitempty"`
}

// Validate validates the price component.
func (p *PriceComponent) Validate() error {
	if p.ResourceType == "" {
		return fmt.Errorf("price component resource_type is required")
	}
	if p.Unit == "" {
		return fmt.Errorf("price component unit is required")
	}
	if !p.Price.IsValid() {
		return fmt.Errorf("price component price is invalid")
	}
	if !p.Price.Amount.IsPositive() {
		return fmt.Errorf("price component price must be positive")
	}
	if !p.Price.Amount.IsUint64() {
		return fmt.Errorf("price component price exceeds uint64")
	}
	return nil
}

// OfferingID is the unique identifier for an offering
type OfferingID struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// Sequence is the provider-scoped sequential offering number
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the offering ID
func (id OfferingID) String() string {
	return fmt.Sprintf("%s/%d", id.ProviderAddress, id.Sequence)
}

// Validate validates the offering ID
func (id OfferingID) Validate() error {
	if id.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// Hash returns a unique hash of the offering ID
func (id OfferingID) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(id.String()))
	return h.Sum(nil)
}

// EncryptedProviderSecrets holds encrypted provider secrets
// These are stored on-chain but only decryptable by intended recipients
type EncryptedProviderSecrets struct {
	// Envelope contains the encrypted provider secrets
	Envelope encryptiontypes.EncryptedPayloadEnvelope `json:"envelope"`

	// EnvelopeRef optionally points to the stored envelope in off-chain storage
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// RecipientKeyIDs are the key IDs that can decrypt this (optional redundancy)
	RecipientKeyIDs []string `json:"recipient_key_ids,omitempty"`
}

// Validate validates the encrypted provider secrets
func (s *EncryptedProviderSecrets) Validate() error {
	if err := s.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}

	for _, keyID := range s.RecipientKeyIDs {
		if !s.Envelope.IsRecipient(keyID) {
			return fmt.Errorf("recipient key id not present in envelope recipients: %s", keyID)
		}
	}

	return nil
}

// Offering represents a marketplace offering from a provider
type Offering struct {
	// ID is the unique offering identifier
	ID OfferingID `json:"id"`

	// State is the current offering state
	State OfferingState `json:"state"`

	// Category is the offering category
	Category OfferingCategory `json:"category"`

	// Name is the public name of the offering
	Name string `json:"name"`

	// Description is the public description
	Description string `json:"description"`

	// Version is the offering version (semantic versioning)
	Version string `json:"version"`

	// Pricing contains the pricing information
	Pricing PricingInfo `json:"pricing"`

	// Prices contains component-based pricing information (preferred).
	Prices []PriceComponent `json:"prices,omitempty"`

	// AllowBidding indicates if bidding is allowed for this offering.
	AllowBidding bool `json:"allow_bidding"`

	// MinBid is the minimum bid price when bidding is enabled.
	MinBid sdk.Coin `json:"min_bid,omitempty"`

	// IdentityRequirement defines identity verification requirements
	IdentityRequirement IdentityRequirement `json:"identity_requirement"`

	// RequireMFAForOrders indicates if MFA is required for placing orders
	RequireMFAForOrders bool `json:"require_mfa_for_orders"`

	// PublicMetadata contains publicly visible metadata
	PublicMetadata map[string]string `json:"public_metadata,omitempty"`

	// EncryptedSecrets contains encrypted provider secrets (credentials, endpoints, etc.)
	EncryptedSecrets *EncryptedProviderSecrets `json:"encrypted_secrets,omitempty"`

	// Specifications contains technical specifications
	Specifications map[string]string `json:"specifications,omitempty"`

	// Tags are searchable tags
	Tags []string `json:"tags,omitempty"`

	// Regions are supported regions
	Regions []string `json:"regions,omitempty"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`

	// ActivatedAt is when the offering became active
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// TerminatedAt is when the offering was terminated
	TerminatedAt *time.Time `json:"terminated_at,omitempty"`

	// MaxConcurrentOrders is the maximum concurrent orders (0 = unlimited)
	MaxConcurrentOrders uint32 `json:"max_concurrent_orders,omitempty"`

	// TotalOrderCount is the total number of orders placed
	TotalOrderCount uint64 `json:"total_order_count"`

	// ActiveOrderCount is the current number of active orders
	ActiveOrderCount uint64 `json:"active_order_count"`
}

// NewOffering creates a new offering with required fields
func NewOffering(id OfferingID, name string, category OfferingCategory, pricing PricingInfo) *Offering {
	return NewOfferingAt(id, name, category, pricing, time.Unix(0, 0))
}

// NewOfferingAt creates a new offering with a caller-provided timestamp
func NewOfferingAt(id OfferingID, name string, category OfferingCategory, pricing PricingInfo, now time.Time) *Offering {
	createdAt := now.UTC()
	return &Offering{
		ID:                  id,
		State:               OfferingStateActive,
		Category:            category,
		Name:                name,
		Pricing:             pricing,
		IdentityRequirement: DefaultIdentityRequirement(),
		PublicMetadata:      make(map[string]string),
		Specifications:      make(map[string]string),
		Tags:                make([]string, 0),
		Regions:             make([]string, 0),
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}
}

// Validate validates the offering
func (o *Offering) Validate() error {
	if err := o.ID.Validate(); err != nil {
		return fmt.Errorf("invalid offering ID: %w", err)
	}

	if !o.State.IsValid() {
		return fmt.Errorf("invalid offering state: %s", o.State)
	}

	if o.Name == "" {
		return fmt.Errorf("offering name is required")
	}

	if o.Category == "" {
		return fmt.Errorf("offering category is required")
	}

	if len(o.Prices) > 0 {
		if err := validatePriceComponents(o.Prices); err != nil {
			return fmt.Errorf("invalid price components: %w", err)
		}
	} else if err := o.Pricing.Validate(); err != nil {
		return fmt.Errorf("invalid pricing: %w", err)
	}

	if err := o.IdentityRequirement.Validate(); err != nil {
		return fmt.Errorf("invalid identity requirement: %w", err)
	}

	if o.EncryptedSecrets != nil {
		if err := o.EncryptedSecrets.Validate(); err != nil {
			return fmt.Errorf("invalid encrypted secrets: %w", err)
		}
	}

	if o.AllowBidding {
		if !o.MinBid.IsValid() {
			return fmt.Errorf("min bid is invalid")
		}
		if !o.MinBid.Amount.IsPositive() {
			return fmt.Errorf("min bid must be positive")
		}
	}

	return nil
}

func validatePriceComponents(prices []PriceComponent) error {
	if len(prices) == 0 {
		return fmt.Errorf("price components are required")
	}

	var denom string
	seen := make(map[string]struct{})
	for _, component := range prices {
		if err := component.Validate(); err != nil {
			return err
		}
		if denom == "" {
			denom = component.Price.Denom
		} else if denom != component.Price.Denom {
			return fmt.Errorf("price component denom mismatch: %s vs %s", denom, component.Price.Denom)
		}
		key := fmt.Sprintf("%s:%s", component.ResourceType, component.Unit)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate price component: %s", key)
		}
		seen[key] = struct{}{}
	}

	return nil
}

// CanAcceptOrder checks if the offering can accept new orders
func (o *Offering) CanAcceptOrder() error {
	if !o.State.IsAcceptingOrders() {
		return fmt.Errorf("offering is not accepting orders: state=%s", o.State)
	}

	if o.MaxConcurrentOrders > 0 && o.ActiveOrderCount >= uint64(o.MaxConcurrentOrders) {
		return fmt.Errorf("offering has reached maximum concurrent orders: %d", o.MaxConcurrentOrders)
	}

	return nil
}

// Hash returns a unique hash of the offering
func (o *Offering) Hash() []byte {
	h := sha256.New()
	h.Write(o.ID.Hash())
	h.Write([]byte(o.Name))
	h.Write([]byte(o.Version))
	_, _ = fmt.Fprintf(h, "%d", o.State)
	return h.Sum(nil)
}

// Offerings is a slice of Offering
type Offerings []Offering

// Active returns only active offerings
func (offerings Offerings) Active() Offerings {
	result := make(Offerings, 0)
	for _, o := range offerings {
		if o.State == OfferingStateActive {
			result = append(result, o)
		}
	}
	return result
}

// ByProvider returns offerings for a specific provider
func (offerings Offerings) ByProvider(providerAddress string) Offerings {
	result := make(Offerings, 0)
	for _, o := range offerings {
		if o.ID.ProviderAddress == providerAddress {
			result = append(result, o)
		}
	}
	return result
}
