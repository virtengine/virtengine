package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	"github.com/virtengine/virtengine/sdk/go/testutil"

	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/market/keeper"
)

func setupBenchKeeper(b *testing.B) (sdk.Context, keeper.IKeeper) {
	b.Helper()

	suite := state.SetupTestSuite(b)
	suite.PrepareMocks(func(ts *state.TestSuite) {
		bkeeper := ts.BankKeeper()

		bkeeper.
			On("SendCoinsFromAccountToModule", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		bkeeper.
			On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		bkeeper.
			On("SendCoinsFromModuleToModule", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
	})

	return suite.Context(), suite.MarketKeeper()
}

// BenchmarkProviderHasActiveLeases benchmarks the provider active lease check
// using the secondary index vs full scan.
func BenchmarkProviderHasActiveLeases(b *testing.B) {
	ctx, k := setupBenchKeeper(b)

	// Create test provider address
	provider := testutil.AccAddress(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = k.ProviderHasActiveLeases(ctx, provider)
	}
}

// BenchmarkGetOrder benchmarks single order retrieval
func BenchmarkGetOrder(b *testing.B) {
	ctx, k := setupBenchKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.GroupSpec(b)

	order, err := k.CreateOrder(ctx, gid, spec)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = k.GetOrder(ctx, order.ID)
	}
}

// BenchmarkWithOrders benchmarks full order iteration
func BenchmarkWithOrders(b *testing.B) {
	ctx, k := setupBenchKeeper(b)

	// Create multiple test orders
	for i := 0; i < 100; i++ {
		gid := dtypes.GroupID{
			Owner: testutil.AccAddress(b).String(),
			DSeq:  uint64(i + 1),
			GSeq:  1,
		}
		spec := testutil.GroupSpec(b)
		_, _ = k.CreateOrder(ctx, gid, spec)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		k.WithOrders(ctx, func(order types.Order) bool {
			count++
			return false
		})
	}
}

// BenchmarkWithBidsForOrder benchmarks bid iteration for a specific order
func BenchmarkWithBidsForOrder(b *testing.B) {
	ctx, k := setupBenchKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.GroupSpec(b)
	order, err := k.CreateOrder(ctx, gid, spec)
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
		price := sdk.NewDecCoin("uakt", sdkmath.NewInt(1000))
		_, _ = k.CreateBid(ctx, bidID, price, types.ResourcesOffer{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		k.WithBidsForOrder(ctx, order.ID, types.BidOpen, func(bid types.Bid) bool {
			count++
			return false
		})
	}
}

// BenchmarkBidCountForOrder benchmarks counting bids for an order
func BenchmarkBidCountForOrder(b *testing.B) {
	ctx, k := setupBenchKeeper(b)

	// Create a test order
	gid := dtypes.GroupID{
		Owner: testutil.AccAddress(b).String(),
		DSeq:  1,
		GSeq:  1,
	}
	spec := testutil.GroupSpec(b)
	order, err := k.CreateOrder(ctx, gid, spec)
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
		price := sdk.NewDecCoin("uakt", sdkmath.NewInt(1000))
		_, _ = k.CreateBid(ctx, bidID, price, types.ResourcesOffer{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = k.BidCountForOrder(ctx, order.ID)
	}
}
