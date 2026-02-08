package keeper

import (
	"encoding/binary"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

// Keeper maintains the resources module state.
type Keeper struct {
	cdc        codec.BinaryCodec
	skey       storetypes.StoreKey
	paramSpace paramtypes.Subspace
	authority  string
}

// NewKeeper creates a new resources keeper.
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, paramSpace paramtypes.Subspace, authority string) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		skey:       skey,
		paramSpace: paramSpace,
		authority:  authority,
	}
}

// Codec returns the codec.
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns the store key.
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// Logger returns module logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// GetAuthority returns the module authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetParams gets module params.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets module params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	k.paramSpace.SetParamSet(ctx, &params)
	return nil
}

func (k Keeper) nextSequence(ctx sdk.Context, key []byte) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(key)
	if len(bz) == 0 {
		bz = make([]byte, 8)
	}
	seq := binary.BigEndian.Uint64(bz)
	seq++
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(key, bz)
	return seq
}
