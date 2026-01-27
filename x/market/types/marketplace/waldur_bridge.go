// Package marketplace provides types for the marketplace on-chain module.
//
// VE-303: Waldur bridge module: synchronize public ledger data into Waldur
// This file defines types for the Waldur bridge that synchronizes public ledger data
// and routes Waldur actions back on-chain.
package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// WaldurSyncType represents types of data that can be synced to Waldur
type WaldurSyncType string

const (
	// SyncTypeOffering syncs offering data
	SyncTypeOffering WaldurSyncType = "offering"

	// SyncTypeOrder syncs order data
	SyncTypeOrder WaldurSyncType = "order"

	// SyncTypeProvider syncs provider profile data
	SyncTypeProvider WaldurSyncType = "provider"

	// SyncTypeAllocation syncs allocation data
	SyncTypeAllocation WaldurSyncType = "allocation"

	// SyncTypeUsage syncs usage data
	SyncTypeUsage WaldurSyncType = "usage"

	// SyncTypeBid syncs bid data
	SyncTypeBid WaldurSyncType = "bid"
)

// WaldurActionType represents types of actions from Waldur
type WaldurActionType string

const (
	// ActionTypeProvision requests provisioning of an allocation
	ActionTypeProvision WaldurActionType = "provision"

	// ActionTypeTerminate requests termination of an allocation
	ActionTypeTerminate WaldurActionType = "terminate"

	// ActionTypeServiceDesk creates a service desk ticket
	ActionTypeServiceDesk WaldurActionType = "service_desk"

	// ActionTypeUsageReport submits usage report
	ActionTypeUsageReport WaldurActionType = "usage_report"

	// ActionTypeStatusUpdate updates status
	ActionTypeStatusUpdate WaldurActionType = "status_update"

	// ActionTypeApproval requests approval
	ActionTypeApproval WaldurActionType = "approval"
)

// WaldurSyncState represents the sync state of an entity
type WaldurSyncState uint8

const (
	// SyncStatePending indicates sync is pending
	SyncStatePending WaldurSyncState = 0

	// SyncStateSynced indicates entity is synced
	SyncStateSynced WaldurSyncState = 1

	// SyncStateFailed indicates sync failed
	SyncStateFailed WaldurSyncState = 2

	// SyncStateOutOfSync indicates entity is out of sync
	SyncStateOutOfSync WaldurSyncState = 3
)

// WaldurSyncStateNames maps sync states to names
var WaldurSyncStateNames = map[WaldurSyncState]string{
	SyncStatePending:   "pending",
	SyncStateSynced:    "synced",
	SyncStateFailed:    "failed",
	SyncStateOutOfSync: "out_of_sync",
}

// String returns the string representation of a sync state
func (s WaldurSyncState) String() string {
	if name, ok := WaldurSyncStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// WaldurSyncRecord tracks sync state for on-chain entities
type WaldurSyncRecord struct {
	// EntityType is the type of entity being synced
	EntityType WaldurSyncType `json:"entity_type"`

	// EntityID is the on-chain entity ID
	EntityID string `json:"entity_id"`

	// WaldurID is the corresponding Waldur entity ID
	WaldurID string `json:"waldur_id,omitempty"`

	// State is the current sync state
	State WaldurSyncState `json:"state"`

	// LastSyncedAt is when the entity was last synced
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`

	// LastSyncAttemptAt is when sync was last attempted
	LastSyncAttemptAt *time.Time `json:"last_sync_attempt_at,omitempty"`

	// SyncVersion is the version of data that was synced
	SyncVersion uint64 `json:"sync_version"`

	// ChainVersion is the current chain version of the data
	ChainVersion uint64 `json:"chain_version"`

	// FailureCount is the number of consecutive sync failures
	FailureCount uint32 `json:"failure_count"`

	// LastError is the last sync error
	LastError string `json:"last_error,omitempty"`

	// Checksum is a checksum of the synced data
	Checksum string `json:"checksum"`
}

// NewWaldurSyncRecord creates a new sync record
func NewWaldurSyncRecord(entityType WaldurSyncType, entityID string) *WaldurSyncRecord {
	return &WaldurSyncRecord{
		EntityType:   entityType,
		EntityID:     entityID,
		State:        SyncStatePending,
		SyncVersion:  0,
		ChainVersion: 1,
	}
}

// NeedsSync returns true if the entity needs syncing
func (r *WaldurSyncRecord) NeedsSync() bool {
	return r.State == SyncStatePending ||
		r.State == SyncStateOutOfSync ||
		r.SyncVersion < r.ChainVersion
}

// MarkSynced marks the entity as synced
func (r *WaldurSyncRecord) MarkSynced(waldurID string, checksum string) {
	now := time.Now().UTC()
	r.WaldurID = waldurID
	r.State = SyncStateSynced
	r.LastSyncedAt = &now
	r.SyncVersion = r.ChainVersion
	r.FailureCount = 0
	r.LastError = ""
	r.Checksum = checksum
}

// MarkFailed marks sync as failed
func (r *WaldurSyncRecord) MarkFailed(err string) {
	now := time.Now().UTC()
	r.State = SyncStateFailed
	r.LastSyncAttemptAt = &now
	r.FailureCount++
	r.LastError = err
}

// IncrementChainVersion increments the chain version
func (r *WaldurSyncRecord) IncrementChainVersion() {
	r.ChainVersion++
	if r.State == SyncStateSynced {
		r.State = SyncStateOutOfSync
	}
}

// WaldurCallback represents a callback from Waldur to the chain
type WaldurCallback struct {
	// ID is the unique callback ID
	ID string `json:"id"`

	// ActionType is the type of action requested
	ActionType WaldurActionType `json:"action_type"`

	// WaldurID is the Waldur entity ID
	WaldurID string `json:"waldur_id"`

	// ChainEntityType is the chain entity type
	ChainEntityType WaldurSyncType `json:"chain_entity_type"`

	// ChainEntityID is the chain entity ID
	ChainEntityID string `json:"chain_entity_id"`

	// Payload contains action-specific payload (public data only)
	Payload map[string]string `json:"payload,omitempty"`

	// Signature is the signature over the callback
	Signature []byte `json:"signature"`

	// SignerID identifies the signer
	SignerID string `json:"signer_id"`

	// Nonce is a unique nonce for replay protection
	Nonce string `json:"nonce"`

	// Timestamp is when the callback was created
	Timestamp time.Time `json:"timestamp"`

	// ExpiresAt is when the callback expires
	ExpiresAt time.Time `json:"expires_at"`
}

// NewWaldurCallback creates a new Waldur callback
func NewWaldurCallback(actionType WaldurActionType, waldurID string, chainEntityType WaldurSyncType, chainEntityID string) *WaldurCallback {
	now := time.Now().UTC()
	nonce := generateNonce()

	return &WaldurCallback{
		ID:              fmt.Sprintf("wcb_%s_%s", chainEntityID, nonce[:8]),
		ActionType:      actionType,
		WaldurID:        waldurID,
		ChainEntityType: chainEntityType,
		ChainEntityID:   chainEntityID,
		Payload:         make(map[string]string),
		Nonce:           nonce,
		Timestamp:       now,
		ExpiresAt:       now.Add(time.Hour), // 1 hour expiry
	}
}

// generateNonce generates a random nonce
func generateNonce() string {
	now := time.Now().UnixNano()
	h := sha256.Sum256([]byte(fmt.Sprintf("%d", now)))
	return hex.EncodeToString(h[:16])
}

// SigningPayload returns the payload to be signed
func (c *WaldurCallback) SigningPayload() []byte {
	h := sha256.New()
	h.Write([]byte(c.ID))
	h.Write([]byte(c.ActionType))
	h.Write([]byte(c.WaldurID))
	h.Write([]byte(c.ChainEntityType))
	h.Write([]byte(c.ChainEntityID))
	h.Write([]byte(c.Nonce))
	h.Write([]byte(fmt.Sprintf("%d", c.Timestamp.Unix())))
	return h.Sum(nil)
}

// IsExpired returns true if the callback has expired
func (c *WaldurCallback) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// Validate validates the callback
func (c *WaldurCallback) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("callback ID is required")
	}
	if c.ActionType == "" {
		return fmt.Errorf("action type is required")
	}
	if c.ChainEntityID == "" {
		return fmt.Errorf("chain entity ID is required")
	}
	if c.Nonce == "" {
		return fmt.Errorf("nonce is required")
	}
	if len(c.Signature) == 0 {
		return fmt.Errorf("signature is required")
	}
	if c.IsExpired() {
		return fmt.Errorf("callback has expired")
	}
	return nil
}

// WaldurCallbackState represents the processing state of a callback
type WaldurCallbackState uint8

const (
	// CallbackStatePending indicates callback is pending processing
	CallbackStatePending WaldurCallbackState = 0

	// CallbackStateProcessing indicates callback is being processed
	CallbackStateProcessing WaldurCallbackState = 1

	// CallbackStateCompleted indicates callback was processed successfully
	CallbackStateCompleted WaldurCallbackState = 2

	// CallbackStateFailed indicates callback processing failed
	CallbackStateFailed WaldurCallbackState = 3

	// CallbackStateRejected indicates callback was rejected (invalid/expired)
	CallbackStateRejected WaldurCallbackState = 4
)

// WaldurCallbackRecord tracks callback processing
type WaldurCallbackRecord struct {
	// CallbackID is the callback ID
	CallbackID string `json:"callback_id"`

	// State is the processing state
	State WaldurCallbackState `json:"state"`

	// ReceivedAt is when the callback was received
	ReceivedAt time.Time `json:"received_at"`

	// ProcessedAt is when the callback was processed
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// TransactionHash is the resulting transaction hash
	TransactionHash string `json:"transaction_hash,omitempty"`

	// Error is any error message
	Error string `json:"error,omitempty"`

	// RetryCount is the number of processing retries
	RetryCount uint32 `json:"retry_count"`
}

// NewWaldurCallbackRecord creates a new callback record
func NewWaldurCallbackRecord(callbackID string) *WaldurCallbackRecord {
	return &WaldurCallbackRecord{
		CallbackID: callbackID,
		State:      CallbackStatePending,
		ReceivedAt: time.Now().UTC(),
	}
}

// ProcessedNonces tracks processed nonces for replay protection
type ProcessedNonces struct {
	// Nonces is a map of nonce -> expiry time
	Nonces map[string]time.Time `json:"nonces"`

	// MaxAge is the maximum age of nonces to track
	MaxAge time.Duration `json:"max_age"`
}

// NewProcessedNonces creates a new processed nonces tracker
func NewProcessedNonces(maxAge time.Duration) *ProcessedNonces {
	return &ProcessedNonces{
		Nonces: make(map[string]time.Time),
		MaxAge: maxAge,
	}
}

// IsProcessed checks if a nonce has been processed
func (p *ProcessedNonces) IsProcessed(nonce string) bool {
	expiry, exists := p.Nonces[nonce]
	if !exists {
		return false
	}
	// Check if nonce record has expired
	if time.Now().After(expiry) {
		delete(p.Nonces, nonce)
		return false
	}
	return true
}

// MarkProcessed marks a nonce as processed
func (p *ProcessedNonces) MarkProcessed(nonce string) {
	p.Nonces[nonce] = time.Now().Add(p.MaxAge)
}

// Cleanup removes expired nonces
func (p *ProcessedNonces) Cleanup() {
	now := time.Now()
	for nonce, expiry := range p.Nonces {
		if now.After(expiry) {
			delete(p.Nonces, nonce)
		}
	}
}

// WaldurOfferingExport represents an offering exported to Waldur
type WaldurOfferingExport struct {
	// ChainOfferingID is the on-chain offering ID
	ChainOfferingID string `json:"chain_offering_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// Name is the offering name
	Name string `json:"name"`

	// Description is the offering description
	Description string `json:"description"`

	// Category is the offering category
	Category string `json:"category"`

	// State is the offering state
	State string `json:"state"`

	// PricingModel is the pricing model
	PricingModel string `json:"pricing_model"`

	// BasePrice is the base price
	BasePrice uint64 `json:"base_price"`

	// Currency is the currency
	Currency string `json:"currency"`

	// Regions are supported regions
	Regions []string `json:"regions,omitempty"`

	// Tags are searchable tags
	Tags []string `json:"tags,omitempty"`

	// Specifications are technical specifications
	Specifications map[string]string `json:"specifications,omitempty"`

	// IdentityScoreRequired is the minimum identity score
	IdentityScoreRequired uint32 `json:"identity_score_required"`

	// MFARequired indicates if MFA is required
	MFARequired bool `json:"mfa_required"`

	// Version is the data version for sync
	Version uint64 `json:"version"`

	// UpdatedAt is when the data was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// FromOffering creates an export from an offering
func (e *WaldurOfferingExport) FromOffering(o *Offering) {
	e.ChainOfferingID = o.ID.String()
	e.ProviderAddress = o.ID.ProviderAddress
	e.Name = o.Name
	e.Description = o.Description
	e.Category = string(o.Category)
	e.State = o.State.String()
	e.PricingModel = string(o.Pricing.Model)
	e.BasePrice = o.Pricing.BasePrice
	e.Currency = o.Pricing.Currency
	e.Regions = o.Regions
	e.Tags = o.Tags
	e.Specifications = o.Specifications
	e.IdentityScoreRequired = o.IdentityRequirement.MinScore
	e.MFARequired = o.RequireMFAForOrders
	e.UpdatedAt = o.UpdatedAt
}

// WaldurOrderExport represents an order exported to Waldur
type WaldurOrderExport struct {
	// ChainOrderID is the on-chain order ID
	ChainOrderID string `json:"chain_order_id"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`

	// OfferingID is the offering ID
	OfferingID string `json:"offering_id"`

	// State is the order state
	State string `json:"state"`

	// Region is the requested region
	Region string `json:"region,omitempty"`

	// Quantity is the requested quantity
	Quantity uint32 `json:"quantity"`

	// PublicMetadata is publicly visible metadata
	PublicMetadata map[string]string `json:"public_metadata,omitempty"`

	// AllocatedProvider is the allocated provider (if any)
	AllocatedProvider string `json:"allocated_provider,omitempty"`

	// Version is the data version for sync
	Version uint64 `json:"version"`

	// CreatedAt is when the order was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the order was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// FromOrder creates an export from an order
func (e *WaldurOrderExport) FromOrder(o *Order) {
	e.ChainOrderID = o.ID.String()
	e.CustomerAddress = o.ID.CustomerAddress
	e.OfferingID = o.OfferingID.String()
	e.State = o.State.String()
	e.Region = o.Region
	e.Quantity = o.RequestedQuantity
	e.PublicMetadata = o.PublicMetadata
	e.AllocatedProvider = o.AllocatedProviderAddress
	e.CreatedAt = o.CreatedAt
	e.UpdatedAt = o.UpdatedAt
}

// WaldurProviderExport represents a provider profile exported to Waldur
type WaldurProviderExport struct {
	// Address is the provider's address
	Address string `json:"address"`

	// Name is the provider name
	Name string `json:"name"`

	// Description is the provider description
	Description string `json:"description"`

	// IdentityVerified indicates if identity is verified
	IdentityVerified bool `json:"identity_verified"`

	// IdentityScore is the provider's identity score
	IdentityScore uint32 `json:"identity_score"`

	// DomainVerified indicates if domain is verified
	DomainVerified bool `json:"domain_verified"`

	// Domain is the verified domain
	Domain string `json:"domain,omitempty"`

	// ActiveOfferingCount is the number of active offerings
	ActiveOfferingCount uint32 `json:"active_offering_count"`

	// TotalOrderCount is the total orders fulfilled
	TotalOrderCount uint64 `json:"total_order_count"`

	// Regions are regions where provider operates
	Regions []string `json:"regions,omitempty"`

	// Capabilities are provider capabilities
	Capabilities []string `json:"capabilities,omitempty"`

	// Version is the data version for sync
	Version uint64 `json:"version"`

	// UpdatedAt is when the profile was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// WaldurBridgeConfig holds configuration for the Waldur bridge
type WaldurBridgeConfig struct {
	// Enabled indicates if the bridge is enabled
	Enabled bool `json:"enabled"`

	// WaldurEndpoint is the Waldur API endpoint
	WaldurEndpoint string `json:"waldur_endpoint"`

	// SyncIntervalSeconds is the sync interval
	SyncIntervalSeconds int64 `json:"sync_interval_seconds"`

	// CallbackExpirySeconds is how long callbacks are valid
	CallbackExpirySeconds int64 `json:"callback_expiry_seconds"`

	// MaxRetries is the maximum sync retries
	MaxRetries uint32 `json:"max_retries"`

	// RetryBackoffSeconds is the retry backoff
	RetryBackoffSeconds int64 `json:"retry_backoff_seconds"`

	// SignerPubKeys are the authorized signer public keys
	SignerPubKeys []string `json:"signer_pub_keys"`

	// NonceWindowSeconds is the nonce validity window
	NonceWindowSeconds int64 `json:"nonce_window_seconds"`
}

// DefaultWaldurBridgeConfig returns default bridge configuration
func DefaultWaldurBridgeConfig() WaldurBridgeConfig {
	return WaldurBridgeConfig{
		Enabled:               false,
		SyncIntervalSeconds:   60,
		CallbackExpirySeconds: 3600,
		MaxRetries:            3,
		RetryBackoffSeconds:   30,
		NonceWindowSeconds:    7200, // 2 hours
	}
}
