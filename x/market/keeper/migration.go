package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
)

// MigrateReservationLinks returns deterministic canonical market preflight counts.
func (k Keeper) MigrateReservationLinks(ctx sdk.Context) (orders, bids, leases, linked uint64) {
	orderIDs := make([]string, 0)
	bidIDs := make([]string, 0)
	leaseIDs := make([]string, 0)
	k.WithOrders(ctx, func(order markettypes.Order) bool {
		orderIDs = append(orderIDs, order.ID.String())
		if order.ReservationId != "" {
			linked++
		}
		return false
	})
	k.WithBids(ctx, func(bid markettypes.Bid) bool {
		bidIDs = append(bidIDs, bid.ID.String())
		if bid.ReservationId != "" {
			linked++
		}
		return false
	})
	k.WithLeases(ctx, func(lease mv1.Lease) bool {
		leaseIDs = append(leaseIDs, lease.ID.String())
		if lease.ReservationId != "" {
			linked++
		}
		return false
	})
	sort.Strings(orderIDs)
	sort.Strings(bidIDs)
	sort.Strings(leaseIDs)
	return uint64(len(orderIDs)), uint64(len(bidIDs)), uint64(len(leaseIDs)), linked
}
