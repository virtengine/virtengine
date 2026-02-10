package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// IKeeper defines the expected interface for the BME module keeper.
type IKeeper interface {
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	GetAuthority() string

	// Params
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// State
	GetState(ctx sdk.Context) types.State
	SetState(ctx sdk.Context, state types.State) error

	// Escrow Operations
	HoldEscrow(ctx sdk.Context, orderID string, depositor sdk.AccAddress, amount sdk.Coins) error
	ReleaseEscrow(ctx sdk.Context, orderID string, recipient sdk.AccAddress) error
	GetEscrowBalance(ctx sdk.Context, orderID string) sdk.Coins

	// Settlement Operations
	SettleBilling(ctx sdk.Context, leaseID string, provider sdk.AccAddress, usageAmount sdk.Coins) error

	// Fee Collection
	CollectFees(ctx sdk.Context, payer sdk.AccAddress, amount sdk.Coins) error

	// Token Operations
	MintTokens(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coins) error
	BurnTokens(ctx sdk.Context, from sdk.AccAddress, amount sdk.Coins) error

	// Query server
	NewQuerier() Querier
}

// Keeper implements the BME module keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string

	// External keepers
	bankKeeper   BankKeeper
	oracleKeeper OracleKeeper
}

// NewKeeper creates a new BME Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	bankKeeper BankKeeper,
	oracleKeeper OracleKeeper,
) IKeeper {
	return &Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		authority:    authority,
		bankKeeper:   bankKeeper,
		oracleKeeper: oracleKeeper,
	}
}

// BankKeeper returns the bank keeper (for use in msg_server).
func (k *Keeper) BankKeeper() BankKeeper {
	return k.bankKeeper
}

// OracleKeeper returns the oracle keeper (for use in msg_server).
func (k *Keeper) OracleKeeper() OracleKeeper {
	return k.oracleKeeper
}

var (
	// ErrEscrowNotFound is returned when an escrow record is not found.
	ErrEscrowNotFound = errors.New("escrow not found")
	// ErrInsufficientFunds is returned when there are not enough funds.
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrInvalidAmount is returned when an amount is invalid.
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrBankKeeperNotSet is returned when bank keeper is not configured.
	ErrBankKeeperNotSet = errors.New("bank keeper not configured")
)

// Codec returns the keeper's codec.
func (k *Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns the keeper's store key.
func (k *Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetAuthority returns the x/bme module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current BME module parameters.
func (k *Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsPrefix())
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the BME module parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsPrefix(), bz)

	return nil
}

// GetState returns the current vault state.
func (k *Keeper) GetState(ctx sdk.Context) types.State {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.StatePrefix())
	if bz == nil {
		return types.State{
			Balances:      sdk.Coins{},
			TotalBurned:   sdk.Coins{},
			TotalMinted:   sdk.Coins{},
			RemintCredits: sdk.Coins{},
		}
	}

	var state types.State
	k.cdc.MustUnmarshal(bz, &state)
	return state
}

// SetState sets the vault state.
func (k *Keeper) SetState(ctx sdk.Context, state types.State) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.StatePrefix(), bz)
	return nil
}

// NewQuerier creates a new Querier instance.
func (k *Keeper) NewQuerier() Querier {
	return Querier{keeper: k}
}

// Querier implements the grpc query service for the BME module.
type Querier struct {
	keeper IKeeper
}

// Params implements the Query/Params gRPC method.
func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.keeper.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// VaultState implements the Query/VaultState gRPC method.
func (q Querier) VaultState(ctx context.Context, req *types.QueryVaultStateRequest) (*types.QueryVaultStateResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	state := q.keeper.GetState(sdkCtx)
	return &types.QueryVaultStateResponse{VaultState: state}, nil
}

// Status implements the Query/Status gRPC method.
func (q Querier) Status(ctx context.Context, req *types.QueryStatusRequest) (*types.QueryStatusResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.keeper.GetParams(sdkCtx)

	// Calculate status based on params and current state
	// For now, return healthy status as default
	warnThreshold := math.LegacyNewDecFromInt(math.NewInt(int64(params.CircuitBreakerWarnThreshold))).Quo(math.LegacyNewDec(10000))
	haltThreshold := math.LegacyNewDecFromInt(math.NewInt(int64(params.CircuitBreakerHaltThreshold))).Quo(math.LegacyNewDec(10000))

	return &types.QueryStatusResponse{
		Status:          types.MintStatusHealthy,
		CollateralRatio: math.LegacyOneDec(),
		WarnThreshold:   warnThreshold,
		HaltThreshold:   haltThreshold,
		MintsAllowed:    true,
		RefundsAllowed:  true,
	}, nil
}

// Ensure Querier implements the QueryServer interface
var _ types.QueryServer = Querier{}
