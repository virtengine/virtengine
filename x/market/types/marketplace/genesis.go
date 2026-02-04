// Package marketplace provides types for the marketplace on-chain module.
//
// VE-300 to VE-304: Marketplace on-chain module
// This file defines genesis state and keys for the marketplace module.
package marketplace

import (
	"fmt"
)

// ModuleName is the module name
const ModuleName = "mktplace"

// StoreKey is the store key
const StoreKey = ModuleName

// RouterKey is the router key
const RouterKey = ModuleName

// Key prefixes for state storage
var (
	// OfferingKeyPrefix is the prefix for offering storage
	OfferingKeyPrefix = []byte{0x01}

	// OrderKeyPrefix is the prefix for order storage
	OrderKeyPrefix = []byte{0x02}

	// AllocationKeyPrefix is the prefix for allocation storage
	AllocationKeyPrefix = []byte{0x03}

	// AllocationByCustomerPrefix indexes allocations by customer address
	AllocationByCustomerPrefix = []byte{0x0F}

	// AllocationByProviderPrefix indexes allocations by provider address
	AllocationByProviderPrefix = []byte{0x10}

	// BidKeyPrefix is the prefix for bid storage
	BidKeyPrefix = []byte{0x04}

	// ParamsKeyPrefix is the prefix for params storage
	ParamsKeyPrefix = []byte{0x05}

	// SyncRecordKeyPrefix is the prefix for Waldur sync records
	SyncRecordKeyPrefix = []byte{0x06}

	// CallbackRecordKeyPrefix is the prefix for callback records
	CallbackRecordKeyPrefix = []byte{0x07}

	// ProcessedNonceKeyPrefix is the prefix for processed nonces
	ProcessedNonceKeyPrefix = []byte{0x08}

	// EventSequenceKeyPrefix is the prefix for event sequence
	EventSequenceKeyPrefix = []byte{0x09}

	// EventCheckpointKeyPrefix is the prefix for event checkpoints
	EventCheckpointKeyPrefix = []byte{0x0A}

	// SubscriptionKeyPrefix is the prefix for event subscriptions
	SubscriptionKeyPrefix = []byte{0x0B}

	// ProviderSettingsKeyPrefix is the prefix for provider identity settings
	ProviderSettingsKeyPrefix = []byte{0x0C}

	// MFAConfigKeyPrefix is the prefix for MFA action configs
	MFAConfigKeyPrefix = []byte{0x0D}

	// MFAAuditKeyPrefix is the prefix for MFA audit records
	MFAAuditKeyPrefix = []byte{0x0E}
)

// Key construction functions

// OfferingKey returns the key for an offering
func OfferingKey(id OfferingID) []byte {
	return append(OfferingKeyPrefix, []byte(id.String())...)
}

// OrderKey returns the key for an order
func OrderKey(id OrderID) []byte {
	return append(OrderKeyPrefix, []byte(id.String())...)
}

// AllocationKey returns the key for an allocation
func AllocationKey(id AllocationID) []byte {
	return append(AllocationKeyPrefix, []byte(id.String())...)
}

// AllocationByCustomerKey returns the key for allocation indexed by customer
func AllocationByCustomerKey(customerAddress, allocationID string) []byte {
	return append(AllocationByCustomerPrefix, []byte(fmt.Sprintf("%s/%s", customerAddress, allocationID))...)
}

// AllocationByCustomerPrefixKey returns the prefix for allocations by customer
func AllocationByCustomerPrefixKey(customerAddress string) []byte {
	return append(AllocationByCustomerPrefix, []byte(customerAddress+"/")...)
}

// AllocationByProviderKey returns the key for allocation indexed by provider
func AllocationByProviderKey(providerAddress, allocationID string) []byte {
	return append(AllocationByProviderPrefix, []byte(fmt.Sprintf("%s/%s", providerAddress, allocationID))...)
}

// AllocationByProviderPrefixKey returns the prefix for allocations by provider
func AllocationByProviderPrefixKey(providerAddress string) []byte {
	return append(AllocationByProviderPrefix, []byte(providerAddress+"/")...)
}

// BidKey returns the key for a bid
func BidKey(id BidID) []byte {
	return append(BidKeyPrefix, []byte(id.String())...)
}

// ParamsKey returns the params key
func ParamsKey() []byte {
	return ParamsKeyPrefix
}

// SyncRecordKey returns the key for a sync record
func SyncRecordKey(entityType WaldurSyncType, entityID string) []byte {
	return append(SyncRecordKeyPrefix, []byte(fmt.Sprintf("%s/%s", entityType, entityID))...)
}

// CallbackRecordKey returns the key for a callback record
func CallbackRecordKey(callbackID string) []byte {
	return append(CallbackRecordKeyPrefix, []byte(callbackID)...)
}

// ProcessedNonceKey returns the key for a processed nonce
func ProcessedNonceKey(nonce string) []byte {
	return append(ProcessedNonceKeyPrefix, []byte(nonce)...)
}

// EventSequenceKey returns the key for the event sequence counter
func EventSequenceKey() []byte {
	return EventSequenceKeyPrefix
}

// EventCheckpointKey returns the key for an event checkpoint
func EventCheckpointKey(subscriberID string) []byte {
	return append(EventCheckpointKeyPrefix, []byte(subscriberID)...)
}

// SubscriptionKey returns the key for a subscription
func SubscriptionKey(subscriberID string) []byte {
	return append(SubscriptionKeyPrefix, []byte(subscriberID)...)
}

// ProviderSettingsKey returns the key for provider identity settings
func ProviderSettingsKey(providerAddress string) []byte {
	return append(ProviderSettingsKeyPrefix, []byte(providerAddress)...)
}

// MFAConfigKey returns the key for MFA action config
func MFAConfigKey(actionType MarketplaceActionType) []byte {
	return append(MFAConfigKeyPrefix, []byte(fmt.Sprintf("%d", actionType))...)
}

// MFAAuditKey returns the key for an MFA audit record
func MFAAuditKey(challengeID string) []byte {
	return append(MFAAuditKeyPrefix, []byte(challengeID)...)
}

// Params defines the module parameters
type Params struct {
	// MaxOfferingsPerProvider is the maximum offerings per provider
	MaxOfferingsPerProvider uint32 `json:"max_offerings_per_provider"`

	// MaxOrdersPerCustomer is the maximum concurrent orders per customer
	MaxOrdersPerCustomer uint32 `json:"max_orders_per_customer"`

	// MaxBidsPerOrder is the maximum bids per order
	MaxBidsPerOrder uint32 `json:"max_bids_per_order"`

	// OrderExpiryBlocks is the default order expiry in blocks
	OrderExpiryBlocks int64 `json:"order_expiry_blocks"`

	// DefaultIdentityScoreRequired is the default identity score required
	DefaultIdentityScoreRequired uint32 `json:"default_identity_score_required"`

	// EnableIdentityGating enables identity gating
	EnableIdentityGating bool `json:"enable_identity_gating"`

	// EnableMFAGating enables MFA gating
	EnableMFAGating bool `json:"enable_mfa_gating"`

	// EnableWaldurBridge enables Waldur bridge
	EnableWaldurBridge bool `json:"enable_waldur_bridge"`

	// WaldurConfig is the Waldur bridge configuration
	WaldurConfig WaldurBridgeConfig `json:"waldur_config"`

	// MFAConfigs are the MFA action configurations
	MFAConfigs []MFAActionConfig `json:"mfa_configs"`

	// ECON-002: Marketplace Economics Parameters

	// EconomicsParams are the marketplace economics parameters
	EconomicsParams EconomicsParams `json:"economics_params"`

	// ProviderIncentiveConfig is the provider incentive configuration
	ProviderIncentiveConfig ProviderIncentiveConfig `json:"provider_incentive_config"`

	// LiquidityIncentiveParams are the liquidity incentive parameters
	LiquidityIncentiveParams LiquidityIncentiveParams `json:"liquidity_incentive_params"`

	// PriceDiscoveryParams are the price discovery parameters
	PriceDiscoveryParams PriceDiscoveryParams `json:"price_discovery_params"`

	// SafeguardParams are the anti-manipulation safeguard parameters
	SafeguardParams SafeguardParams `json:"safeguard_params"`

	// MarketMetricsParams are the market metrics parameters
	MarketMetricsParams MarketMetricsParams `json:"market_metrics_params"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxOfferingsPerProvider:      100,
		MaxOrdersPerCustomer:         50,
		MaxBidsPerOrder:              100,
		OrderExpiryBlocks:            1000,
		DefaultIdentityScoreRequired: 0,
		EnableIdentityGating:         true,
		EnableMFAGating:              true,
		EnableWaldurBridge:           false,
		WaldurConfig:                 DefaultWaldurBridgeConfig(),
		MFAConfigs:                   DefaultMFAActionConfigs(),
		// ECON-002: Marketplace Economics defaults
		EconomicsParams:          DefaultEconomicsParams(),
		ProviderIncentiveConfig:  DefaultProviderIncentiveConfig(),
		LiquidityIncentiveParams: DefaultLiquidityIncentiveParams(),
		PriceDiscoveryParams:     DefaultPriceDiscoveryParams(),
		SafeguardParams:          DefaultSafeguardParams(),
		MarketMetricsParams:      DefaultMarketMetricsParams(),
	}
}

// Validate validates parameters
func (p Params) Validate() error {
	if p.MaxOfferingsPerProvider == 0 {
		return fmt.Errorf("max_offerings_per_provider must be positive")
	}
	if p.MaxOrdersPerCustomer == 0 {
		return fmt.Errorf("max_orders_per_customer must be positive")
	}
	if p.MaxBidsPerOrder == 0 {
		return fmt.Errorf("max_bids_per_order must be positive")
	}
	if p.DefaultIdentityScoreRequired > 100 {
		return fmt.Errorf("default_identity_score_required cannot exceed 100")
	}
	// ECON-002: Validate economics parameters
	if err := p.EconomicsParams.Validate(); err != nil {
		return fmt.Errorf("invalid economics_params: %w", err)
	}
	if err := p.ProviderIncentiveConfig.Validate(); err != nil {
		return fmt.Errorf("invalid provider_incentive_config: %w", err)
	}
	if err := p.LiquidityIncentiveParams.Validate(); err != nil {
		return fmt.Errorf("invalid liquidity_incentive_params: %w", err)
	}
	if err := p.PriceDiscoveryParams.Validate(); err != nil {
		return fmt.Errorf("invalid price_discovery_params: %w", err)
	}
	if err := p.SafeguardParams.Validate(); err != nil {
		return fmt.Errorf("invalid safeguard_params: %w", err)
	}
	if err := p.MarketMetricsParams.Validate(); err != nil {
		return fmt.Errorf("invalid market_metrics_params: %w", err)
	}
	return nil
}

// GenesisState defines the genesis state
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// Offerings are the offerings
	Offerings []Offering `json:"offerings"`

	// Orders are the orders
	Orders []Order `json:"orders"`

	// Allocations are the allocations
	Allocations []Allocation `json:"allocations"`

	// Bids are the marketplace bids
	Bids []MarketplaceBid `json:"bids"`

	// ProviderSettings are the provider identity settings
	ProviderSettings map[string]ProviderIdentitySettings `json:"provider_settings"`

	// MFAConfigs are the MFA action configurations
	MFAConfigs []MFAActionConfig `json:"mfa_configs"`

	// EventSequence is the current event sequence number
	EventSequence uint64 `json:"event_sequence"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:           DefaultParams(),
		Offerings:        make([]Offering, 0),
		Orders:           make([]Order, 0),
		Allocations:      make([]Allocation, 0),
		Bids:             make([]MarketplaceBid, 0),
		ProviderSettings: make(map[string]ProviderIdentitySettings),
		MFAConfigs:       DefaultMFAActionConfigs(),
		EventSequence:    0,
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	offeringIDs := make(map[string]bool)
	for _, offering := range gs.Offerings {
		if err := offering.Validate(); err != nil {
			return fmt.Errorf("invalid offering %s: %w", offering.ID.String(), err)
		}
		if offeringIDs[offering.ID.String()] {
			return fmt.Errorf("duplicate offering ID: %s", offering.ID.String())
		}
		offeringIDs[offering.ID.String()] = true
	}

	orderIDs := make(map[string]bool)
	for _, order := range gs.Orders {
		if err := order.Validate(); err != nil {
			return fmt.Errorf("invalid order %s: %w", order.ID.String(), err)
		}
		if orderIDs[order.ID.String()] {
			return fmt.Errorf("duplicate order ID: %s", order.ID.String())
		}
		orderIDs[order.ID.String()] = true
	}

	allocationIDs := make(map[string]bool)
	for _, allocation := range gs.Allocations {
		if err := allocation.Validate(); err != nil {
			return fmt.Errorf("invalid allocation %s: %w", allocation.ID.String(), err)
		}
		if allocationIDs[allocation.ID.String()] {
			return fmt.Errorf("duplicate allocation ID: %s", allocation.ID.String())
		}
		allocationIDs[allocation.ID.String()] = true
	}

	return nil
}
