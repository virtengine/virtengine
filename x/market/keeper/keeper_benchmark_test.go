package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	testutil "github.com/virtengine/virtengine/testutil/keeper"
)

// BenchmarkProviderHasActiveLeases benchmarks the provider active lease check
// using the secondary index vs full scan.
func BenchmarkProviderHasActiveLeases(b *testing.B) {
	ctx, keeper := testutil.MarketKeeper(b)

	// Create test provider address
	provider := testutil.AccAddress(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = keeper.ProviderHasActiveLeases(ctx, provider)
	}
}

// BenchmarkGetOrder benchmarks single order retrieval
func BenchmarkGetOrder(b *testing.B) {
	ctx, keeper := testutil.MarketKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.DefaultGroupSpec()

	order, err := keeper.CreateOrder(ctx, gid, spec)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = keeper.GetOrder(ctx, order.ID)
	}
}

// BenchmarkWithOrders benchmarks full order iteration
func BenchmarkWithOrders(b *testing.B) {
	ctx, keeper := testutil.MarketKeeper(b)

	// Create multiple test orders
	for i := 0; i < 100; i++ {
		gid := dtypes.GroupID{
			Owner: testutil.AccAddress(b).String(),
			DSeq:  uint64(i + 1),
			GSeq:  1,
		}
		spec := testutil.DefaultGroupSpec()
		_, _ = keeper.CreateOrder(ctx, gid, spec)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		keeper.WithOrders(ctx, func(order types.Order) bool {
			count++
			return false
		})
	}
}

// BenchmarkWithBidsForOrder benchmarks bid iteration for a specific order
func BenchmarkWithBidsForOrder(b *testing.B) {
	ctx, keeper := testutil.MarketKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.DefaultGroupSpec()
	order, err := keeper.CreateOrder(ctx, gid, spec)
	require.NoError(b, err)

	// Create multiple bids for the order
	for i := 0; i < 50; i++ {
		bidID := mv1.BidID{
			Owner:    order.ID.Owner,
			DSeq:     order.ID.DSeq,
			GSeq:     order.ID.GSeq,
			OSeq:     order.ID.OSeq,
			Provider: testutil.AccAddress(b).String(),
			BSeq:     uint32(i + 1),
		}
		price := sdk.NewDecCoin("uakt", sdk.NewInt(1000))
		_, _ = keeper.CreateBid(ctx, bidID, price, types.ResourcesOffer{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		keeper.WithBidsForOrder(ctx, order.ID, types.BidOpen, func(bid types.Bid) bool {
			count++
			return false
		})
	}
}

// BenchmarkBidCountForOrder benchmarks counting bids for an order
func BenchmarkBidCountForOrder(b *testing.B) {
	ctx, keeper := testutil.MarketKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.DefaultGroupSpec()
	order, err := keeper.CreateOrder(ctx, gid, spec)
	require.NoError(b, err)

	// Create multiple bids
	for i := 0; i < 20; i++ {
		bidID := mv1.BidID{
			Owner:    order.ID.Owner,
			DSeq:     order.ID.DSeq,
			GSeq:     order.ID.GSeq,
			OSeq:     order.ID.OSeq,
			Provider: testutil.AccAddress(b).String(),
			BSeq:     uint32(i + 1),
		}
		price := sdk.NewDecCoin("uakt", sdk.NewInt(1000))
		_, _ = keeper.CreateBid(ctx, bidID, price, types.ResourcesOffer{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = keeper.BidCountForOrder(ctx, order.ID)
	}
}
