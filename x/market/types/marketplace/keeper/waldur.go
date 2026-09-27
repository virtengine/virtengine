package keeper

import (
	"encoding/json"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ============================================================================
// Waldur source registry (ADR-010)
// ============================================================================

// SetWaldurSource registers or updates a trusted Waldur instance.
func (k Keeper) SetWaldurSource(ctx sdk.Context, source *marketplace.WaldurSource) error {
	if err := source.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(source)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(marketplace.WaldurSourceKey(source.InstanceID), bz)
	return nil
}

// GetWaldurSource returns a registered Waldur instance.
func (k Keeper) GetWaldurSource(ctx sdk.Context, instanceID string) (*marketplace.WaldurSource, bool) {
	bz := ctx.KVStore(k.skey).Get(marketplace.WaldurSourceKey(instanceID))
	if bz == nil {
		return nil, false
	}
	var source marketplace.WaldurSource
	if err := json.Unmarshal(bz, &source); err != nil {
		return nil, false
	}
	return &source, true
}

// WithWaldurSources iterates over registered Waldur sources.
func (k Keeper) WithWaldurSources(ctx sdk.Context, fn func(marketplace.WaldurSource) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.WaldurSourceKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var source marketplace.WaldurSource
		if err := json.Unmarshal(iter.Value(), &source); err != nil {
			continue
		}
		if fn(source) {
			break
		}
	}
}

// ============================================================================
// Waldur command queue (ADR-010, chain -> adapter)
// ============================================================================

// EnqueueWaldurCommand persists a durable command for off-chain adapters.
func (k Keeper) EnqueueWaldurCommand(ctx sdk.Context, command *marketplace.WaldurCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(command)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(marketplace.WaldurCommandKey(command.ID), bz)
	return nil
}

// GetWaldurCommand returns a durable command by ID.
func (k Keeper) GetWaldurCommand(ctx sdk.Context, id string) (*marketplace.WaldurCommand, bool) {
	bz := ctx.KVStore(k.skey).Get(marketplace.WaldurCommandKey(id))
	if bz == nil {
		return nil, false
	}
	var command marketplace.WaldurCommand
	if err := json.Unmarshal(bz, &command); err != nil {
		return nil, false
	}
	return &command, true
}

// WithWaldurCommands iterates over durable commands.
func (k Keeper) WithWaldurCommands(ctx sdk.Context, fn func(marketplace.WaldurCommand) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.WaldurCommandKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var command marketplace.WaldurCommand
		if err := json.Unmarshal(iter.Value(), &command); err != nil {
			continue
		}
		if fn(command) {
			break
		}
	}
}

// AckWaldurCommand marks a durable command as acknowledged.
func (k Keeper) AckWaldurCommand(ctx sdk.Context, id string) error {
	command, found := k.GetWaldurCommand(ctx, id)
	if !found {
		return marketplace.ErrWaldurCallbackInvalid.Wrap("waldur command not found")
	}
	command.MarkAcked(ctx.BlockTime())
	bz, err := json.Marshal(command)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(marketplace.WaldurCommandKey(command.ID), bz)
	return nil
}

// ============================================================================
// Waldur offering ingestion (ADR-010, adapter -> chain)
// ============================================================================

// IngestWaldurOffering verifies and applies a signed Waldur offering snapshot.
//
// The HTTP fetch happens off-chain; only the signed snapshot enters consensus.
// Ingestion is idempotent and replay-protected via monotonic snapshot heights.
func (k Keeper) IngestWaldurOffering(
	ctx sdk.Context,
	imp *marketplace.WaldurOfferingImport,
	attestation *marketplace.WaldurOfferingAttestation,
) (marketplace.IngestResult, error) {
	result := marketplace.IngestResult{Timestamp: ctx.BlockTime().UTC()}
	if imp == nil {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap("offering import is required")
	}
	result.WaldurUUID = imp.UUID

	source, found := k.GetWaldurSource(ctx, imp.InstanceID)
	if !found {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrapf("unregistered waldur instance: %s", imp.InstanceID)
	}
	if !source.Active {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrapf("waldur instance is inactive: %s", imp.InstanceID)
	}
	if attestation == nil {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap("attestation is required")
	}
	if attestation.InstanceID != imp.InstanceID || attestation.OfferingUUID != imp.UUID {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap("attestation does not match offering")
	}
	if err := attestation.Verify(source.PublicKey); err != nil {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
	}

	// Replay/idempotency protection: ignore stale snapshots.
	if existing, ok := k.GetWaldurSyncRecord(ctx, marketplace.SyncTypeOffering, imp.UUID); ok {
		if existing.SyncVersion >= attestation.SnapshotHeight {
			result.Success = true
			result.Action = marketplace.IngestActionSkip
			result.ChainOfferingID = existing.WaldurID
			result.Checksum = existing.Checksum
			return result, nil
		}
	}

	cfg := marketplace.DefaultIngestConfig()
	// Provider resolution is handled explicitly below so ingestion does not
	// depend on an off-chain customer->provider map being configured. The
	// on-chain parameter map is merged in when operators configure it.
	cfg.RequireProviderRegistration = false
	for customerUUID, providerAddress := range k.GetParams(ctx).WaldurIngestCustomerProviders {
		cfg.CustomerProviderMap[customerUUID] = providerAddress
	}
	validation := imp.Validate(cfg)
	if !validation.Valid {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap(strings.Join(validation.Errors, "; "))
	}

	providerAddress := validation.ProviderAddress
	if providerAddress == "" {
		if raw, ok := imp.Attributes["ve_provider"]; ok {
			if s, ok := raw.(string); ok {
				providerAddress = strings.TrimSpace(s)
			}
		}
	}
	if providerAddress == "" {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap("provider address could not be resolved")
	}
	providerAddr, err := sdk.AccAddressFromBech32(providerAddress)
	if err != nil {
		return result, marketplace.ErrWaldurCallbackInvalid.Wrap("invalid provider address in offering")
	}
	if !k.IsProvider(ctx, providerAddr) {
		return result, marketplace.ErrNotProvider
	}

	now := ctx.BlockTime().UTC()
	sequence := nextOfferingSequence(ctx, &k, providerAddress)
	offering := imp.ToOfferingAt(providerAddress, sequence, cfg, now)

	if _, exists := k.GetOffering(ctx, offering.ID); exists {
		if err := k.UpdateOffering(ctx, offering); err != nil {
			return result, err
		}
		result.Action = marketplace.IngestActionUpdate
	} else {
		if err := k.CreateOffering(ctx, offering); err != nil {
			return result, err
		}
		result.Action = marketplace.IngestActionCreate
	}

	checksum := imp.IngestChecksum()
	record := marketplace.NewWaldurSyncRecord(marketplace.SyncTypeOffering, imp.UUID)
	record.ChainVersion = attestation.SnapshotHeight
	record.MarkSyncedAt(offering.ID.String(), checksum, now)
	if err := k.SetWaldurSyncRecord(ctx, record); err != nil {
		return result, err
	}

	result.Success = true
	result.ChainOfferingID = offering.ID.String()
	result.Checksum = checksum
	return result, nil
}

// ============================================================================
// Unified catalog (ADR-010, browse path)
// ============================================================================

// CatalogFilter narrows the unified catalog query.
type CatalogFilter struct {
	Category        marketplace.OfferingCategory
	Regions         []string
	Backends        []string
	IncludeUnlisted bool
	Source          marketplace.OfferingSource
}

// UnifiedCatalog returns active, browsable offerings from all sources (native
// and Waldur) in deterministic order.
func (k Keeper) UnifiedCatalog(ctx sdk.Context, filter CatalogFilter) []marketplace.Offering {
	result := make([]marketplace.Offering, 0)
	k.WithOfferings(ctx, func(offering marketplace.Offering) bool {
		if !offering.State.IsAcceptingOrders() {
			return false
		}
		if !filter.IncludeUnlisted && !offering.Visibility.IsPublic() {
			return false
		}
		if filter.Category != "" && offering.Category != filter.Category {
			return false
		}
		if filter.Source != "" && offering.Source.Effective() != filter.Source.Effective() {
			return false
		}
		if len(filter.Regions) > 0 && !intersectsStrings(filter.Regions, offering.Regions) {
			return false
		}
		if len(filter.Backends) > 0 && !containsStr(filter.Backends, offering.EffectiveBackendType()) {
			return false
		}
		result = append(result, offering)
		return false
	})

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Source.Effective() != result[j].Source.Effective() {
			return result[i].Source.Effective() < result[j].Source.Effective()
		}
		if result[i].ID.ProviderAddress != result[j].ID.ProviderAddress {
			return result[i].ID.ProviderAddress < result[j].ID.ProviderAddress
		}
		return result[i].ID.Sequence < result[j].ID.Sequence
	})
	return result
}

func intersectsStrings(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

func containsStr(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
