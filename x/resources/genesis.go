package resources

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/keeper"
	"github.com/virtengine/virtengine/x/resources/types"
)

// InitGenesis initializes the module state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types.GenesisState) {
	if gs == nil {
		gs = types.DefaultGenesisState()
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
	if gs.CanonicalReservationsActive {
		k.ActivateCanonicalReservations(ctx)
	}

	for _, inventory := range gs.Inventories {
		if err := k.SetInventory(ctx, inventory); err != nil {
			panic(err)
		}
	}
	for _, allocation := range gs.Allocations {
		if err := k.SetAllocation(ctx, allocation); err != nil {
			panic(err)
		}
	}
	for _, reservation := range gs.Reservations {
		if err := k.SetReservation(ctx, reservation); err != nil {
			panic(err)
		}
		if err := k.RebuildReservationIndexes(ctx, reservation); err != nil {
			panic(err)
		}
	}
	for _, event := range gs.ReservationEvents {
		if err := k.SetReservationEvent(ctx, event); err != nil {
			panic(err)
		}
	}
	if err := k.ValidateCapacityConservation(ctx); err != nil {
		panic(err)
	}
}

// ExportGenesis exports module state to genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	gs := types.DefaultGenesisState()
	gs.Params = k.GetParams(ctx)

	k.WithInventories(ctx, func(inv types.ResourceInventory) bool {
		gs.Inventories = append(gs.Inventories, inv)
		return false
	})
	k.WithAllocations(ctx, func(allocation types.ResourceAllocation) bool {
		gs.Allocations = append(gs.Allocations, allocation)
		return false
	})
	k.WithReservations(ctx, func(reservation types.Reservation) bool {
		gs.Reservations = append(gs.Reservations, reservation)
		return false
	})
	if err := k.WithReservationEvents(ctx, func(event types.ReservationEvent) bool {
		gs.ReservationEvents = append(gs.ReservationEvents, event)
		return false
	}); err != nil {
		panic(err)
	}
	gs.CanonicalReservationsActive = k.IsCanonicalReservationsActive(ctx)

	return gs
}

// MustMarshalGenesis marshals genesis state.
func MustMarshalGenesis(gs *types.GenesisState) []byte {
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}
