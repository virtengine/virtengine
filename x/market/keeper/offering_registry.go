package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ===============================================================================
// VE-25H: Offering Registry for Waldur-synced offerings
// This file implements on-chain storage and management of marketplace offerings
// that are synchronized from Waldur.
// ===============================================================================

// Offering registry storage prefixes.
var (
	OfferingRegistryPrefix   = []byte{0x14, 0x00}
	OfferingByProviderPrefix = []byte{0x14, 0x01}
	OfferingByCategoryPrefix = []byte{0x14, 0x02}
	OfferingByStatePrefix    = []byte{0x14, 0x03}
)

// marshalOffering serializes an offering to bytes.
func marshalOffering(offering *marketplace.Offering) ([]byte, error) {
	return json.Marshal(offering)
}

// unmarshalOffering deserializes an offering from bytes.
func unmarshalOffering(data []byte) (*marketplace.Offering, error) {
	var offering marketplace.Offering
	if err := json.Unmarshal(data, &offering); err != nil {
		return nil, err
	}
	return &offering, nil
}

// offeringByProviderKey returns the key for offerings by provider.
func offeringByProviderKey(provider sdk.AccAddress, offeringHash []byte) []byte {
	key := make([]byte, 0, len(OfferingByProviderPrefix)+len(provider)+len(offeringHash))
	key = append(key, OfferingByProviderPrefix...)
	key = append(key, provider.Bytes()...)
	key = append(key, offeringHash...)
	return key
}

// offeringByCategoryKey returns the key for category index.
func offeringByCategoryKey(category string, offeringHash []byte) []byte {
	key := make([]byte, 0, len(OfferingByCategoryPrefix)+len(category)+len(offeringHash))
	key = append(key, OfferingByCategoryPrefix...)
	key = append(key, []byte(category)...)
	key = append(key, offeringHash...)
	return key
}

// offeringByStateKey returns the key for state index.
func offeringByStateKey(state marketplace.OfferingState, offeringHash []byte) []byte {
	key := make([]byte, 0, len(OfferingByStatePrefix)+1+len(offeringHash))
	key = append(key, OfferingByStatePrefix...)
	key = append(key, byte(state))
	key = append(key, offeringHash...)
	return key
}

// offeringKey returns the primary key for an offering.
func offeringKey(id marketplace.OfferingID) []byte {
	hash := id.Hash()
	key := make([]byte, 0, len(OfferingRegistryPrefix)+len(hash))
	key = append(key, OfferingRegistryPrefix...)
	key = append(key, hash...)
	return key
}

// CreateOfferingFromWaldur creates a new on-chain offering from Waldur sync.
func (k Keeper) CreateOfferingFromWaldur(ctx sdk.Context, offering *marketplace.Offering) error {
	if offering == nil {
		return fmt.Errorf("offering is nil")
	}

	if err := offering.Validate(); err != nil {
		return fmt.Errorf("invalid offering: %w", err)
	}

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)

	// Check if offering already exists
	if store.Has(key) {
		return fmt.Errorf("offering already exists: %s", offering.ID.String())
	}

	// Set timestamps
	now := ctx.BlockTime()
	offering.CreatedAt = now
	offering.UpdatedAt = now
	if offering.State == marketplace.OfferingStateActive {
		offering.ActivatedAt = &now
	}

	// Store the offering
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	// Create indexes
	k.indexOffering(ctx, offering)

	// Log event (use attribute-based events for Cosmos SDK compatibility)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"offering_created",
			sdk.NewAttribute("offering_id", offering.ID.String()),
			sdk.NewAttribute("provider", offering.ID.ProviderAddress),
			sdk.NewAttribute("name", offering.Name),
			sdk.NewAttribute("category", string(offering.Category)),
			sdk.NewAttribute("state", offering.State.String()),
		),
	)

	ctx.Logger().Info("created offering from Waldur",
		"offering_id", offering.ID.String(),
		"name", offering.Name,
		"category", offering.Category,
	)

	return nil
}

// UpdateOfferingFromWaldur updates an existing on-chain offering.
func (k Keeper) UpdateOfferingFromWaldur(ctx sdk.Context, offering *marketplace.Offering) error {
	if offering == nil {
		return fmt.Errorf("offering is nil")
	}

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)

	// Check if offering exists
	if !store.Has(key) {
		return fmt.Errorf("offering not found: %s", offering.ID.String())
	}

	// Load existing offering to preserve some fields
	existing := k.GetOfferingByID(ctx, offering.ID)
	if existing == nil {
		return fmt.Errorf("failed to load existing offering: %s", offering.ID.String())
	}

	// Update timestamps
	offering.CreatedAt = existing.CreatedAt
	offering.UpdatedAt = ctx.BlockTime()

	// Handle state transitions
	if existing.State != offering.State {
		now := ctx.BlockTime()
		if offering.State == marketplace.OfferingStateActive && existing.ActivatedAt == nil {
			offering.ActivatedAt = &now
		}
		if offering.State == marketplace.OfferingStateTerminated {
			offering.TerminatedAt = &now
		}

		// Remove old state index
		k.removeOfferingStateIndex(ctx, existing)
	}

	// Store the updated offering
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	// Update indexes
	k.indexOffering(ctx, offering)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"offering_updated",
			sdk.NewAttribute("offering_id", offering.ID.String()),
			sdk.NewAttribute("provider", offering.ID.ProviderAddress),
			sdk.NewAttribute("name", offering.Name),
			sdk.NewAttribute("state", offering.State.String()),
		),
	)

	ctx.Logger().Info("updated offering from Waldur",
		"offering_id", offering.ID.String(),
		"name", offering.Name,
		"state", offering.State.String(),
	)

	return nil
}

// GetOfferingByID retrieves an offering by its ID.
func (k Keeper) GetOfferingByID(ctx sdk.Context, id marketplace.OfferingID) *marketplace.Offering {
	store := ctx.KVStore(k.skey)
	key := offeringKey(id)

	data := store.Get(key)
	if data == nil {
		return nil
	}

	offering, err := unmarshalOffering(data)
	if err != nil {
		ctx.Logger().Error("failed to unmarshal offering", "error", err)
		return nil
	}
	return offering
}

// GetOfferingByStringID retrieves an offering by its string ID.
func (k Keeper) GetOfferingByStringID(ctx sdk.Context, offeringID string) (*marketplace.Offering, error) {
	id, err := marketplace.ParseOfferingID(offeringID)
	if err != nil {
		return nil, fmt.Errorf("invalid offering ID: %w", err)
	}
	offering := k.GetOfferingByID(ctx, id)
	if offering == nil {
		return nil, fmt.Errorf("offering not found: %s", offeringID)
	}
	return offering, nil
}

// SetOfferingState updates the state of an offering.
func (k Keeper) SetOfferingState(ctx sdk.Context, offeringID string, state marketplace.OfferingState) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}

	oldState := offering.State
	offering.State = state
	now := ctx.BlockTime()
	offering.UpdatedAt = now

	if state == marketplace.OfferingStateActive && offering.ActivatedAt == nil {
		offering.ActivatedAt = &now
	}
	if state == marketplace.OfferingStateTerminated {
		offering.TerminatedAt = &now
	}

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	// Update state index
	if oldState != state {
		k.removeOfferingStateIndex(ctx, offering)
		k.indexOfferingByState(ctx, offering)
	}

	return nil
}

// DeprecateOffering marks an offering as deprecated.
func (k Keeper) DeprecateOffering(ctx sdk.Context, offeringID string) error {
	return k.SetOfferingState(ctx, offeringID, marketplace.OfferingStateDeprecated)
}

// TerminateOffering marks an offering as terminated.
func (k Keeper) TerminateOffering(ctx sdk.Context, offeringID string) error {
	return k.SetOfferingState(ctx, offeringID, marketplace.OfferingStateTerminated)
}

// ListOfferingsByProvider returns all offerings for a provider.
func (k Keeper) ListOfferingsByProvider(ctx sdk.Context, provider sdk.AccAddress) []*marketplace.Offering {
	store := ctx.KVStore(k.skey)
	providerBytes := provider.Bytes()
	prefix := make([]byte, len(OfferingByProviderPrefix)+len(providerBytes))
	copy(prefix, OfferingByProviderPrefix)
	copy(prefix[len(OfferingByProviderPrefix):], providerBytes)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var offerings []*marketplace.Offering
	for ; iter.Valid(); iter.Next() {
		// The value is the offering ID bytes (hash)
		idBytes := iter.Value()
		key := make([]byte, 0, len(OfferingRegistryPrefix)+len(idBytes))
		key = append(key, OfferingRegistryPrefix...)
		key = append(key, idBytes...)
		data := store.Get(key)
		if data != nil {
			if offering, err := unmarshalOffering(data); err == nil {
				offerings = append(offerings, offering)
			}
		}
	}

	return offerings
}

// ListOfferingsByCategory returns all offerings in a category.
func (k Keeper) ListOfferingsByCategory(ctx sdk.Context, category string) []*marketplace.Offering {
	store := ctx.KVStore(k.skey)
	catBytes := []byte(category)
	prefix := make([]byte, len(OfferingByCategoryPrefix)+len(catBytes))
	copy(prefix, OfferingByCategoryPrefix)
	copy(prefix[len(OfferingByCategoryPrefix):], catBytes)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var offerings []*marketplace.Offering
	for ; iter.Valid(); iter.Next() {
		idBytes := iter.Value()
		key := make([]byte, 0, len(OfferingRegistryPrefix)+len(idBytes))
		key = append(key, OfferingRegistryPrefix...)
		key = append(key, idBytes...)
		data := store.Get(key)
		if data != nil {
			if offering, err := unmarshalOffering(data); err == nil {
				offerings = append(offerings, offering)
			}
		}
	}

	return offerings
}

// ListOfferingsByState returns all offerings in a given state.
func (k Keeper) ListOfferingsByState(ctx sdk.Context, state marketplace.OfferingState) []*marketplace.Offering {
	store := ctx.KVStore(k.skey)
	prefix := make([]byte, len(OfferingByStatePrefix)+1)
	copy(prefix, OfferingByStatePrefix)
	prefix[len(OfferingByStatePrefix)] = byte(state)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var offerings []*marketplace.Offering
	for ; iter.Valid(); iter.Next() {
		idBytes := iter.Value()
		key := make([]byte, 0, len(OfferingRegistryPrefix)+len(idBytes))
		key = append(key, OfferingRegistryPrefix...)
		key = append(key, idBytes...)
		data := store.Get(key)
		if data != nil {
			if offering, err := unmarshalOffering(data); err == nil {
				offerings = append(offerings, offering)
			}
		}
	}

	return offerings
}

// ListActiveOfferings returns all active offerings.
func (k Keeper) ListActiveOfferings(ctx sdk.Context) []*marketplace.Offering {
	return k.ListOfferingsByState(ctx, marketplace.OfferingStateActive)
}

// WithOfferings iterates over all offerings.
func (k Keeper) WithOfferings(ctx sdk.Context, fn func(*marketplace.Offering) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, OfferingRegistryPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		offering, err := unmarshalOffering(iter.Value())
		if err != nil {
			continue
		}
		if stop := fn(offering); stop {
			break
		}
	}
}

// GetOfferingCount returns the total number of offerings.
func (k Keeper) GetOfferingCount(ctx sdk.Context) int {
	count := 0
	k.WithOfferings(ctx, func(_ *marketplace.Offering) bool {
		count++
		return false
	})
	return count
}

// GetActiveOfferingCount returns the number of active offerings.
func (k Keeper) GetActiveOfferingCount(ctx sdk.Context) int {
	return len(k.ListActiveOfferings(ctx))
}

// GetProviderOfferingCount returns the number of offerings for a provider.
func (k Keeper) GetProviderOfferingCount(ctx sdk.Context, provider sdk.AccAddress) int {
	return len(k.ListOfferingsByProvider(ctx, provider))
}

// IncrementOfferingOrderCount increments the order count for an offering.
func (k Keeper) IncrementOfferingOrderCount(ctx sdk.Context, offeringID string) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}

	offering.TotalOrderCount++
	offering.ActiveOrderCount++
	offering.UpdatedAt = ctx.BlockTime()

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	return nil
}

// DecrementOfferingActiveOrders decrements the active order count for an offering.
func (k Keeper) DecrementOfferingActiveOrders(ctx sdk.Context, offeringID string) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}

	if offering.ActiveOrderCount > 0 {
		offering.ActiveOrderCount--
		offering.UpdatedAt = ctx.BlockTime()

		store := ctx.KVStore(k.skey)
		key := offeringKey(offering.ID)
		data, err := marshalOffering(offering)
		if err != nil {
			return fmt.Errorf("failed to marshal offering: %w", err)
		}
		store.Set(key, data)
	}

	return nil
}

// indexOffering creates all indexes for an offering.
func (k Keeper) indexOffering(ctx sdk.Context, offering *marketplace.Offering) {
	k.indexOfferingByProvider(ctx, offering)
	k.indexOfferingByCategory(ctx, offering)
	k.indexOfferingByState(ctx, offering)
}

// indexOfferingByProvider creates provider index.
func (k Keeper) indexOfferingByProvider(ctx sdk.Context, offering *marketplace.Offering) {
	store := ctx.KVStore(k.skey)
	provider, err := sdk.AccAddressFromBech32(offering.ID.ProviderAddress)
	if err != nil {
		return
	}

	indexKey := offeringByProviderKey(provider, offering.ID.Hash())
	store.Set(indexKey, offering.ID.Hash())
}

// indexOfferingByCategory creates category index.
func (k Keeper) indexOfferingByCategory(ctx sdk.Context, offering *marketplace.Offering) {
	store := ctx.KVStore(k.skey)
	indexKey := offeringByCategoryKey(string(offering.Category), offering.ID.Hash())
	store.Set(indexKey, offering.ID.Hash())
}

// indexOfferingByState creates state index.
func (k Keeper) indexOfferingByState(ctx sdk.Context, offering *marketplace.Offering) {
	store := ctx.KVStore(k.skey)
	indexKey := offeringByStateKey(offering.State, offering.ID.Hash())
	store.Set(indexKey, offering.ID.Hash())
}

// removeOfferingStateIndex removes state index for an offering.
func (k Keeper) removeOfferingStateIndex(ctx sdk.Context, offering *marketplace.Offering) {
	store := ctx.KVStore(k.skey)
	indexKey := offeringByStateKey(offering.State, offering.ID.Hash())
	store.Delete(indexKey)
}

// OfferingRegistryStats contains statistics about the offering registry.
type OfferingRegistryStats struct {
	TotalOfferings  int            `json:"total_offerings"`
	ActiveOfferings int            `json:"active_offerings"`
	ByCategory      map[string]int `json:"by_category"`
	ByState         map[string]int `json:"by_state"`
}

// GetOfferingRegistryStats returns statistics about the offering registry.
func (k Keeper) GetOfferingRegistryStats(ctx sdk.Context) OfferingRegistryStats {
	stats := OfferingRegistryStats{
		ByCategory: make(map[string]int),
		ByState:    make(map[string]int),
	}

	k.WithOfferings(ctx, func(offering *marketplace.Offering) bool {
		stats.TotalOfferings++
		if offering.State == marketplace.OfferingStateActive {
			stats.ActiveOfferings++
		}
		stats.ByCategory[string(offering.Category)]++
		stats.ByState[offering.State.String()]++
		return false
	})

	return stats
}

// GetNextOfferingSequence returns the next sequence number for a provider's offerings.
func (k Keeper) GetNextOfferingSequence(ctx sdk.Context, provider sdk.AccAddress) uint64 {
	offerings := k.ListOfferingsByProvider(ctx, provider)
	var maxSeq uint64
	for _, o := range offerings {
		if o.ID.Sequence > maxSeq {
			maxSeq = o.ID.Sequence
		}
	}
	return maxSeq + 1
}

// CanAcceptOfferingOrder checks if an offering can accept new orders.
func (k Keeper) CanAcceptOfferingOrder(ctx sdk.Context, offeringID string) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}
	return offering.CanAcceptOrder()
}

// UpdateOfferingPricing updates the pricing for an offering.
func (k Keeper) UpdateOfferingPricing(ctx sdk.Context, offeringID string, pricing marketplace.PricingInfo, prices []marketplace.PriceComponent) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}

	offering.Pricing = pricing
	offering.Prices = prices
	offering.UpdatedAt = ctx.BlockTime()

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	return nil
}

// GetOfferingsByIDs retrieves multiple offerings by their IDs.
func (k Keeper) GetOfferingsByIDs(ctx sdk.Context, ids []string) ([]*marketplace.Offering, error) {
	offerings := make([]*marketplace.Offering, 0, len(ids))
	for _, id := range ids {
		offering, err := k.GetOfferingByStringID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get offering %s: %w", id, err)
		}
		offerings = append(offerings, offering)
	}
	return offerings, nil
}

// OfferingExistsForProvider checks if an offering exists for a provider.
func (k Keeper) OfferingExistsForProvider(ctx sdk.Context, provider sdk.AccAddress, sequence uint64) bool {
	id := marketplace.OfferingID{
		ProviderAddress: provider.String(),
		Sequence:        sequence,
	}
	return k.GetOfferingByID(ctx, id) != nil
}

// OfferingSyncMetadata contains metadata for synced offerings.
type OfferingSyncMetadata struct {
	WaldurUUID     string    `json:"waldur_uuid"`
	SyncChecksum   string    `json:"sync_checksum"`
	LastSyncedAt   time.Time `json:"last_synced_at"`
	SyncVersion    uint64    `json:"sync_version"`
	ProviderDaemon string    `json:"provider_daemon,omitempty"`
}

// SetOfferingSyncMetadata stores sync metadata for an offering.
func (k Keeper) SetOfferingSyncMetadata(ctx sdk.Context, offeringID string, metadata OfferingSyncMetadata) error {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return err
	}

	// Store metadata in public metadata
	if offering.PublicMetadata == nil {
		offering.PublicMetadata = make(map[string]string)
	}
	offering.PublicMetadata["waldur_uuid"] = metadata.WaldurUUID
	offering.PublicMetadata["sync_checksum"] = metadata.SyncChecksum
	offering.PublicMetadata["last_synced_at"] = metadata.LastSyncedAt.UTC().Format(time.RFC3339)

	store := ctx.KVStore(k.skey)
	key := offeringKey(offering.ID)
	data, err := marshalOffering(offering)
	if err != nil {
		return fmt.Errorf("failed to marshal offering: %w", err)
	}
	store.Set(key, data)

	return nil
}

// GetOfferingSyncMetadata retrieves sync metadata for an offering.
func (k Keeper) GetOfferingSyncMetadata(ctx sdk.Context, offeringID string) (*OfferingSyncMetadata, error) {
	offering, err := k.GetOfferingByStringID(ctx, offeringID)
	if err != nil {
		return nil, err
	}

	if offering.PublicMetadata == nil {
		return nil, fmt.Errorf("no sync metadata found")
	}

	waldurUUID := offering.PublicMetadata["waldur_uuid"]
	if waldurUUID == "" {
		return nil, fmt.Errorf("offering not synced from Waldur")
	}

	metadata := &OfferingSyncMetadata{
		WaldurUUID:   waldurUUID,
		SyncChecksum: offering.PublicMetadata["sync_checksum"],
	}

	if ts := offering.PublicMetadata["last_synced_at"]; ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			metadata.LastSyncedAt = t
		}
	}

	return metadata, nil
}
