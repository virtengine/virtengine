package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/support/types"
)

// IKeeper defines the interface for the support keeper
type IKeeper interface {
	// External reference management
	RegisterExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error
	GetExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) (types.ExternalTicketRef, bool)
	UpdateExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error
	RemoveExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) error

	// Query methods
	GetExternalRefsByOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) []types.ExternalTicketRef
	WithExternalRefs(ctx sdk.Context, fn func(ref types.ExternalTicketRef) bool)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the support store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// NewKeeper creates and returns an instance for support keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		skey:      skey,
		authority: authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Parameters
// ============================================================================

// paramsStore is the stored format of params
type paramsStore struct {
	AllowedExternalSystems []string `json:"allowed_external_systems"`
	AllowedExternalDomains []string `json:"allowed_external_domains"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		AllowedExternalSystems: params.AllowedExternalSystems,
		AllowedExternalDomains: params.AllowedExternalDomains,
	})
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey(), bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey())
	if bz == nil {
		return types.DefaultParams()
	}

	var ps paramsStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.DefaultParams()
	}

	return types.Params{
		AllowedExternalSystems: ps.AllowedExternalSystems,
		AllowedExternalDomains: ps.AllowedExternalDomains,
	}
}

// ============================================================================
// External Reference Management
// ============================================================================

// externalRefStore is the stored format of an external ticket reference
type externalRefStore struct {
	ResourceID       string `json:"resource_id"`
	ResourceType     string `json:"resource_type"`
	ExternalSystem   string `json:"external_system"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	CreatedBy        string `json:"created_by"`
	UpdatedAt        int64  `json:"updated_at"`
}

// RegisterExternalRef registers a new external ticket reference
func (k Keeper) RegisterExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	if err := ref.Validate(); err != nil {
		return err
	}

	// Check if ref already exists
	if _, found := k.GetExternalRef(ctx, ref.ResourceType, ref.ResourceID); found {
		return types.ErrRefAlreadyExists.Wrapf("ref for %s/%s already exists", ref.ResourceType, ref.ResourceID)
	}

	// Validate external system is allowed
	params := k.GetParams(ctx)
	if !params.IsSystemAllowed(ref.ExternalSystem) {
		return types.ErrInvalidExternalSystem.Wrapf("system %s is not allowed", ref.ExternalSystem)
	}

	// Set timestamps
	now := ctx.BlockTime()
	ref.CreatedAt = now
	ref.UpdatedAt = now

	// Store the reference
	return k.setExternalRef(ctx, ref)
}

// GetExternalRef returns an external ticket reference
func (k Keeper) GetExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) (types.ExternalTicketRef, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ExternalRefKey(resourceType, resourceID)
	bz := store.Get(key)
	if bz == nil {
		return types.ExternalTicketRef{}, false
	}

	var rs externalRefStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.ExternalTicketRef{}, false
	}

	return k.refStoreToRef(rs), true
}

// UpdateExternalRef updates an existing external ticket reference
func (k Keeper) UpdateExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	// Check if ref exists
	existing, found := k.GetExternalRef(ctx, ref.ResourceType, ref.ResourceID)
	if !found {
		return types.ErrRefNotFound.Wrapf("ref for %s/%s not found", ref.ResourceType, ref.ResourceID)
	}

	// Preserve original creation info
	ref.CreatedAt = existing.CreatedAt
	ref.CreatedBy = existing.CreatedBy
	ref.UpdatedAt = ctx.BlockTime()

	return k.setExternalRef(ctx, ref)
}

// RemoveExternalRef removes an external ticket reference
func (k Keeper) RemoveExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) error {
	ref, found := k.GetExternalRef(ctx, resourceType, resourceID)
	if !found {
		return types.ErrRefNotFound.Wrapf("ref for %s/%s not found", resourceType, resourceID)
	}

	store := ctx.KVStore(k.skey)

	// Remove owner index
	ownerAddr, _ := sdk.AccAddressFromBech32(ref.CreatedBy)
	store.Delete(types.ExternalRefByOwnerKey(ownerAddr.Bytes(), resourceType, resourceID))

	// Remove the reference
	store.Delete(types.ExternalRefKey(resourceType, resourceID))

	return nil
}

// setExternalRef stores an external ticket reference
func (k Keeper) setExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	rs := externalRefStore{
		ResourceID:       ref.ResourceID,
		ResourceType:     string(ref.ResourceType),
		ExternalSystem:   string(ref.ExternalSystem),
		ExternalTicketID: ref.ExternalTicketID,
		ExternalURL:      ref.ExternalURL,
		CreatedAt:        ref.CreatedAt.Unix(),
		CreatedBy:        ref.CreatedBy,
		UpdatedAt:        ref.UpdatedAt.Unix(),
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ExternalRefKey(ref.ResourceType, ref.ResourceID), bz)

	// Add owner index
	ownerAddr, _ := sdk.AccAddressFromBech32(ref.CreatedBy)
	store.Set(types.ExternalRefByOwnerKey(ownerAddr.Bytes(), ref.ResourceType, ref.ResourceID), []byte{1})

	return nil
}

// refStoreToRef converts a stored format to ExternalTicketRef
func (k Keeper) refStoreToRef(rs externalRefStore) types.ExternalTicketRef {
	return types.ExternalTicketRef{
		ResourceID:       rs.ResourceID,
		ResourceType:     types.ResourceType(rs.ResourceType),
		ExternalSystem:   types.ExternalSystem(rs.ExternalSystem),
		ExternalTicketID: rs.ExternalTicketID,
		ExternalURL:      rs.ExternalURL,
		CreatedAt:        time.Unix(rs.CreatedAt, 0),
		CreatedBy:        rs.CreatedBy,
		UpdatedAt:        time.Unix(rs.UpdatedAt, 0),
	}
}

// ============================================================================
// Query Methods
// ============================================================================

// GetExternalRefsByOwner returns all external refs created by an owner
func (k Keeper) GetExternalRefsByOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) []types.ExternalTicketRef {
	var refs []types.ExternalTicketRef

	store := ctx.KVStore(k.skey)
	prefix := types.ExternalRefByOwnerPrefixKey(ownerAddr.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Parse key to get resource type and ID
		key := iter.Key()
		remaining := key[len(prefix):]

		// Find separator between resource type and ID
		for i := range remaining {
			if remaining[i] == '/' {
				resourceType := types.ResourceType(remaining[:i])
				resourceID := string(remaining[i+1:])
				if ref, found := k.GetExternalRef(ctx, resourceType, resourceID); found {
					refs = append(refs, ref)
				}
				break
			}
		}
	}

	return refs
}

// WithExternalRefs iterates over all external refs
func (k Keeper) WithExternalRefs(ctx sdk.Context, fn func(ref types.ExternalTicketRef) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixExternalRef)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rs externalRefStore
		if err := json.Unmarshal(iter.Value(), &rs); err != nil {
			continue
		}

		if fn(k.refStoreToRef(rs)) {
			break
		}
	}
}

// NewGRPCQuerier returns a new GRPCQuerier
func (k Keeper) NewGRPCQuerier() GRPCQuerier {
	return GRPCQuerier{Keeper: k}
}
