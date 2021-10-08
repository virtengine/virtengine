package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepers "github.com/virtengine/virtengine/x/market/handler"
	"github.com/virtengine/virtengine/x/market/types"
	ptypes "github.com/virtengine/virtengine/x/provider/types"
)

func getOrdersWithState(ctx sdk.Context, ks keepers.Keepers, state types.Order_State) []types.Order {
	var orders []types.Order

	ks.Market.WithOrders(ctx, func(order types.Order) bool {
		if order.State == state {
			orders = append(orders, order)
		}

		return false
	})

	return orders
}

func getProviders(ctx sdk.Context, ks keepers.Keepers) []ptypes.Provider {
	var providers []ptypes.Provider

	ks.Provider.WithProviders(ctx, func(provider ptypes.Provider) bool {
		providers = append(providers, provider)

		return false
	})

	return providers
}
