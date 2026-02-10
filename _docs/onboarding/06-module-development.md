# Module Development Guide

This guide explains how to build Cosmos SDK modules for VirtEngine.

## Table of Contents

1. [Module Structure](#module-structure)
2. [Creating a New Module](#creating-a-new-module)
3. [Keeper Pattern](#keeper-pattern)
4. [Messages and Handlers](#messages-and-handlers)
5. [Queries](#queries)
6. [Genesis](#genesis)
7. [Events](#events)
8. [Integration with Other Modules](#integration-with-other-modules)
9. [Module Registration](#module-registration)

---

## Module Structure

Every VirtEngine module follows this directory structure:

```
x/mymodule/
├── keeper/
│   ├── keeper.go          # Keeper struct and core logic
│   ├── msg_server.go      # Message handlers
│   ├── grpc_query.go      # Query handlers
│   ├── genesis.go         # Genesis import/export
│   └── *_test.go          # Tests
├── types/
│   ├── keys.go            # Store keys and prefixes
│   ├── genesis.go         # Genesis state type
│   ├── genesis.pb.go      # Generated from proto
│   ├── params.go          # Module parameters
│   ├── msgs.go            # Message types
│   ├── query.pb.go        # Generated from proto
│   ├── tx.pb.go           # Generated from proto
│   ├── errors.go          # Sentinel errors
│   └── *_test.go          # Type tests
├── client/
│   └── cli/
│       ├── query.go       # CLI query commands
│       └── tx.go          # CLI transaction commands
├── module.go              # Module registration
└── README.md              # Module documentation
```

---

## Creating a New Module

### Step 1: Create Directory Structure

```bash
mkdir -p x/mymodule/{keeper,types,client/cli}
```

### Step 2: Define Store Keys (`types/keys.go`)

```go
package types

const (
    // ModuleName defines the module name
    ModuleName = "mymodule"

    // StoreKey defines the primary module store key
    StoreKey = ModuleName

    // RouterKey defines the module's message routing key
    RouterKey = ModuleName

    // QuerierRoute defines the module's query routing key
    QuerierRoute = ModuleName
)

// Store key prefixes
var (
    // ItemKey is the prefix for item storage
    ItemKeyPrefix = []byte{0x01}
    
    // ParamsKey is the prefix for module parameters
    ParamsKey = []byte{0x02}
)

// ItemKey returns the store key for an item by ID
func ItemKey(id string) []byte {
    return append(ItemKeyPrefix, []byte(id)...)
}
```

### Step 3: Define Types (`types/types.go`)

```go
package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// Item represents an item in the module
type Item struct {
    ID        string    `json:"id"`
    Owner     string    `json:"owner"`
    Data      string    `json:"data"`
    CreatedAt int64     `json:"created_at"`
}

// Validate performs basic validation
func (i Item) Validate() error {
    if i.ID == "" {
        return ErrInvalidItem.Wrap("id cannot be empty")
    }
    if _, err := sdk.AccAddressFromBech32(i.Owner); err != nil {
        return ErrInvalidItem.Wrap("invalid owner address")
    }
    return nil
}
```

### Step 4: Define Errors (`types/errors.go`)

```go
package types

import (
    sdkerrors "cosmossdk.io/errors"
)

// Module sentinel errors
var (
    ErrInvalidItem    = sdkerrors.Register(ModuleName, 1, "invalid item")
    ErrItemNotFound   = sdkerrors.Register(ModuleName, 2, "item not found")
    ErrUnauthorized   = sdkerrors.Register(ModuleName, 3, "unauthorized")
    ErrItemExists     = sdkerrors.Register(ModuleName, 4, "item already exists")
)
```

---

## Keeper Pattern

### IKeeper Interface

Always define an interface before the concrete Keeper:

```go
package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/virtengine/virtengine/x/mymodule/types"
)

// IKeeper defines the expected interface for the keeper
type IKeeper interface {
    // Items
    CreateItem(ctx sdk.Context, msg *types.MsgCreateItem) (*types.Item, error)
    GetItem(ctx sdk.Context, id string) (types.Item, bool)
    SetItem(ctx sdk.Context, item types.Item) error
    DeleteItem(ctx sdk.Context, id string) error
    WithItems(ctx sdk.Context, fn func(types.Item) bool)
    
    // Params
    GetParams(ctx sdk.Context) types.Params
    SetParams(ctx sdk.Context, params types.Params) error
}
```

### Keeper Implementation

```go
package keeper

import (
    "github.com/cosmos/cosmos-sdk/codec"
    storetypes "cosmossdk.io/store/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "github.com/virtengine/virtengine/x/mymodule/types"
)

// Keeper implements the IKeeper interface
type Keeper struct {
    cdc       codec.BinaryCodec
    storeKey  storetypes.StoreKey
    authority string  // x/gov module account for MsgUpdateParams
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
    authority string,
) Keeper {
    return Keeper{
        cdc:       cdc,
        storeKey:  storeKey,
        authority: authority,
    }
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
    return k.authority
}

// CreateItem creates a new item
func (k Keeper) CreateItem(ctx sdk.Context, msg *types.MsgCreateItem) (*types.Item, error) {
    // Check if item already exists
    if _, found := k.GetItem(ctx, msg.Id); found {
        return nil, types.ErrItemExists
    }
    
    item := types.Item{
        ID:        msg.Id,
        Owner:     msg.Owner,
        Data:      msg.Data,
        CreatedAt: ctx.BlockTime().Unix(),
    }
    
    if err := item.Validate(); err != nil {
        return nil, err
    }
    
    if err := k.SetItem(ctx, item); err != nil {
        return nil, err
    }
    
    return &item, nil
}

// GetItem returns an item by ID
func (k Keeper) GetItem(ctx sdk.Context, id string) (types.Item, bool) {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.ItemKey(id))
    if bz == nil {
        return types.Item{}, false
    }
    
    var item types.Item
    k.cdc.MustUnmarshal(bz, &item)
    return item, true
}

// SetItem stores an item
func (k Keeper) SetItem(ctx sdk.Context, item types.Item) error {
    store := ctx.KVStore(k.storeKey)
    bz, err := k.cdc.Marshal(&item)
    if err != nil {
        return err
    }
    store.Set(types.ItemKey(item.ID), bz)
    return nil
}

// DeleteItem removes an item
func (k Keeper) DeleteItem(ctx sdk.Context, id string) error {
    store := ctx.KVStore(k.storeKey)
    store.Delete(types.ItemKey(id))
    return nil
}

// WithItems iterates over all items
func (k Keeper) WithItems(ctx sdk.Context, fn func(types.Item) bool) {
    store := ctx.KVStore(k.storeKey)
    iter := storetypes.KVStorePrefixIterator(store, types.ItemKeyPrefix)
    defer iter.Close()
    
    for ; iter.Valid(); iter.Next() {
        var item types.Item
        k.cdc.MustUnmarshal(iter.Value(), &item)
        if fn(item) {
            break
        }
    }
}
```

---

## Messages and Handlers

### Define Messages (`types/msgs.go`)

```go
package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
    TypeMsgCreateItem = "create_item"
    TypeMsgUpdateItem = "update_item"
    TypeMsgDeleteItem = "delete_item"
)

var (
    _ sdk.Msg = &MsgCreateItem{}
    _ sdk.Msg = &MsgUpdateItem{}
    _ sdk.Msg = &MsgDeleteItem{}
)

// MsgCreateItem creates a new item
type MsgCreateItem struct {
    Owner string `json:"owner"`
    Id    string `json:"id"`
    Data  string `json:"data"`
}

func NewMsgCreateItem(owner, id, data string) *MsgCreateItem {
    return &MsgCreateItem{
        Owner: owner,
        Id:    id,
        Data:  data,
    }
}

func (msg MsgCreateItem) Route() string { return RouterKey }
func (msg MsgCreateItem) Type() string  { return TypeMsgCreateItem }

func (msg MsgCreateItem) GetSigners() []sdk.AccAddress {
    owner, _ := sdk.AccAddressFromBech32(msg.Owner)
    return []sdk.AccAddress{owner}
}

func (msg MsgCreateItem) ValidateBasic() error {
    if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
        return ErrInvalidItem.Wrap("invalid owner address")
    }
    if msg.Id == "" {
        return ErrInvalidItem.Wrap("id cannot be empty")
    }
    return nil
}

func (msg MsgCreateItem) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
```

### Message Server (`keeper/msg_server.go`)

```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/virtengine/virtengine/x/mymodule/types"
)

type msgServer struct {
    Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
    return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateItem handles MsgCreateItem
func (m msgServer) CreateItem(goCtx context.Context, msg *types.MsgCreateItem) (*types.MsgCreateItemResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    item, err := m.Keeper.CreateItem(ctx, msg)
    if err != nil {
        return nil, err
    }
    
    // Emit event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeItemCreated,
            sdk.NewAttribute(types.AttributeKeyItemID, item.ID),
            sdk.NewAttribute(types.AttributeKeyOwner, item.Owner),
        ),
    )
    
    return &types.MsgCreateItemResponse{
        Id: item.ID,
    }, nil
}

// UpdateParams handles MsgUpdateParams (governance only)
func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Only governance can update params
    if m.GetAuthority() != msg.Authority {
        return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
    }
    
    if err := m.SetParams(ctx, msg.Params); err != nil {
        return nil, err
    }
    
    return &types.MsgUpdateParamsResponse{}, nil
}
```

---

## Queries

### Query Server (`keeper/grpc_query.go`)

```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/query"
    "github.com/virtengine/virtengine/x/mymodule/types"
)

type queryServer struct {
    Keeper
}

// NewQueryServerImpl returns an implementation of QueryServer
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
    return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Item returns a single item by ID
func (q queryServer) Item(goCtx context.Context, req *types.QueryItemRequest) (*types.QueryItemResponse, error) {
    if req == nil {
        return nil, types.ErrInvalidItem.Wrap("empty request")
    }
    
    ctx := sdk.UnwrapSDKContext(goCtx)
    item, found := q.Keeper.GetItem(ctx, req.Id)
    if !found {
        return nil, types.ErrItemNotFound
    }
    
    return &types.QueryItemResponse{Item: item}, nil
}

// Items returns paginated items
func (q queryServer) Items(goCtx context.Context, req *types.QueryItemsRequest) (*types.QueryItemsResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    var items []types.Item
    store := ctx.KVStore(q.storeKey)
    itemStore := prefix.NewStore(store, types.ItemKeyPrefix)
    
    pageRes, err := query.Paginate(itemStore, req.Pagination, func(key []byte, value []byte) error {
        var item types.Item
        if err := q.cdc.Unmarshal(value, &item); err != nil {
            return err
        }
        items = append(items, item)
        return nil
    })
    if err != nil {
        return nil, err
    }
    
    return &types.QueryItemsResponse{
        Items:      items,
        Pagination: pageRes,
    }, nil
}
```

---

## Genesis

### Genesis Types (`types/genesis.go`)

```go
package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
    return &GenesisState{
        Params: DefaultParams(),
        Items:  []Item{},
    }
}

// Validate performs basic validation of genesis state
func (gs GenesisState) Validate() error {
    // Validate params
    if err := gs.Params.Validate(); err != nil {
        return err
    }
    
    // Check for duplicate items
    itemIDs := make(map[string]bool)
    for _, item := range gs.Items {
        if err := item.Validate(); err != nil {
            return err
        }
        if itemIDs[item.ID] {
            return ErrInvalidItem.Wrapf("duplicate item ID: %s", item.ID)
        }
        itemIDs[item.ID] = true
    }
    
    return nil
}
```

### Genesis Keeper (`keeper/genesis.go`)

```go
package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/virtengine/virtengine/x/mymodule/types"
)

// InitGenesis initializes the module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) {
    // Set params
    if err := k.SetParams(ctx, gs.Params); err != nil {
        panic(err)
    }
    
    // Set items
    for _, item := range gs.Items {
        if err := k.SetItem(ctx, item); err != nil {
            panic(err)
        }
    }
}

// ExportGenesis exports the module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
    var items []types.Item
    k.WithItems(ctx, func(item types.Item) bool {
        items = append(items, item)
        return false
    })
    
    return &types.GenesisState{
        Params: k.GetParams(ctx),
        Items:  items,
    }
}
```

---

## Events

### Define Events (`types/events.go`)

```go
package types

const (
    EventTypeItemCreated = "item_created"
    EventTypeItemUpdated = "item_updated"
    EventTypeItemDeleted = "item_deleted"
    
    AttributeKeyItemID   = "item_id"
    AttributeKeyOwner    = "owner"
)
```

### Emit Events

```go
// In message handler
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeItemCreated,
        sdk.NewAttribute(types.AttributeKeyItemID, item.ID),
        sdk.NewAttribute(types.AttributeKeyOwner, item.Owner),
    ),
)
```

---

## Integration with Other Modules

### Expected Keepers

Define interfaces for modules you depend on:

```go
package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank module interface
type BankKeeper interface {
    SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
    GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// VEIDKeeper defines the expected VEID module interface
type VEIDKeeper interface {
    GetScore(ctx sdk.Context, address string) (uint8, bool)
}
```

### Inject Dependencies

```go
type Keeper struct {
    cdc        codec.BinaryCodec
    storeKey   storetypes.StoreKey
    authority  string
    bankKeeper types.BankKeeper
    veidKeeper types.VEIDKeeper
}

func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
    authority string,
    bankKeeper types.BankKeeper,
    veidKeeper types.VEIDKeeper,
) Keeper {
    return Keeper{
        cdc:        cdc,
        storeKey:   storeKey,
        authority:  authority,
        bankKeeper: bankKeeper,
        veidKeeper: veidKeeper,
    }
}
```

---

## Module Registration

### Module Definition (`module.go`)

```go
package mymodule

import (
    "context"
    "encoding/json"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/spf13/cobra"
    
    "github.com/virtengine/virtengine/x/mymodule/keeper"
    "github.com/virtengine/virtengine/x/mymodule/types"
)

var (
    _ module.AppModule      = AppModule{}
    _ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements module.AppModuleBasic
type AppModuleBasic struct {
    cdc codec.Codec
}

func (AppModuleBasic) Name() string {
    return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
    types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
    types.RegisterInterfaces(registry)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    return cdc.MustMarshalJSON(types.DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
    var gs types.GenesisState
    if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
        return err
    }
    return gs.Validate()
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
    types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// VirtEngine modules do not export CLI commands via Cosmos interface
func (AppModuleBasic) GetTxCmd() *cobra.Command {
    panic("virtengine modules do not export cli commands via cosmos interface")
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
    panic("virtengine modules do not export cli commands via cosmos interface")
}

// AppModule implements module.AppModule
type AppModule struct {
    AppModuleBasic
    keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
    return AppModule{
        AppModuleBasic: AppModuleBasic{cdc: cdc},
        keeper:         keeper,
    }
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
    types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
    types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
    var gs types.GenesisState
    cdc.MustUnmarshalJSON(data, &gs)
    am.keeper.InitGenesis(ctx, &gs)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    gs := am.keeper.ExportGenesis(ctx)
    return cdc.MustMarshalJSON(gs)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }
```

---

## Best Practices Summary

| Practice | Description |
|----------|-------------|
| Define IKeeper interface | All public keeper methods in interface |
| Use authority for params | `x/gov` module account for `MsgUpdateParams` |
| Use storetypes.StoreKey | Not deprecated `sdk.StoreKey` |
| Emit events | For all significant state changes |
| Validate input | In `ValidateBasic()` for messages |
| Test genesis | Import/export round-trip tests |
| Bounded iterations | Use pagination for large queries |

---

## Related Documentation

- [Architecture Overview](./02-architecture-overview.md) - System design
- [Patterns & Anti-patterns](./07-patterns-antipatterns.md) - Best practices
- [Code Review Checklist](./05-code-review-checklist.md) - Review standards
- [Cosmos SDK Modules](https://docs.cosmos.network/v0.50/build/building-modules/intro) - Upstream documentation
