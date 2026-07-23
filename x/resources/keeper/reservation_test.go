package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	"github.com/virtengine/virtengine/x/resources/keeper"
	"github.com/virtengine/virtengine/x/resources/types"
)

func TestReservationTransitionTable(t *testing.T) {
	tests := []struct {
		name string
		from types.ReservationState
		to   types.ReservationState
		ok   bool
	}{
		{"pending active", types.ReservationStatePending, types.ReservationStateActive, true},
		{"active consumed", types.ReservationStateActive, types.ReservationStateConsumed, true},
		{"pending released", types.ReservationStatePending, types.ReservationStateReleased, true},
		{"active released", types.ReservationStateActive, types.ReservationStateReleased, true},
		{"consumed released", types.ReservationStateConsumed, types.ReservationStateReleased, true},
		{"pending expired", types.ReservationStatePending, types.ReservationStateExpired, true},
		{"active quarantined", types.ReservationStateActive, types.ReservationStateQuarantined, true},
		{"quarantined slashed", types.ReservationStateQuarantined, types.ReservationStateSlashed, true},
		{"released active", types.ReservationStateReleased, types.ReservationStateActive, false},
		{"expired active", types.ReservationStateExpired, types.ReservationStateActive, false},
		{"slashed active", types.ReservationStateSlashed, types.ReservationStateActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.ok, types.CanTransitionReservation(tt.from, tt.to))
		})
	}
}

func TestReserveIsIdempotentAndRejectsConflictingPayload(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-idempotent", "virtengine1provideridempotent", 8)

	req := reservationRequest("idem-1", "job", "job-1", 2)
	first, err := k.Reserve(ctx, req)
	require.NoError(t, err)
	require.NotEqual(t, req.IdempotencyKey, first.ReservationId)
	require.Contains(t, first.ReservationId, "reservation/")

	second, err := k.Reserve(ctx, req)
	require.NoError(t, err)
	require.Equal(t, first.ReservationId, second.ReservationId)

	conflict := req
	conflict.Capacity.CpuCores = 3
	_, err = k.Reserve(ctx, conflict)
	require.ErrorIs(t, err, types.ErrReservationConflict)

	inv, found := k.GetInventory(ctx, first.ProviderAddress, first.ResourceClass, first.InventoryId)
	require.True(t, found)
	require.Equal(t, int64(6), inv.Available.CpuCores)
}

func TestLegacyAllocationWritesRemainReplayableUntilActivation(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-legacy-window", "virtengine1providerlegacywindow", 2)

	allocation, err := k.AllocateResources(ctx, types.ResourceRequest{
		RequestId: "legacy-window", RequesterAddress: "virtengine1requesterlegacywindow",
		ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, allocation)

	k.ActivateCanonicalReservations(ctx)
	_, err = k.AllocateResources(ctx, types.ResourceRequest{
		RequestId: "post-activation", RequesterAddress: "virtengine1requesterlegacywindow",
		ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1},
	})
	require.ErrorIs(t, err, types.ErrLegacyAllocationDeprecated)
}

func TestReservationsByProviderPaginationCursor(t *testing.T) {
	k, ctx := setupKeeper(t)
	provider := seedInventory(t, k, ctx, "inv-pagination", "virtengine1providerpagination", 3)
	for i := 0; i < 3; i++ {
		request := reservationRequest(fmt.Sprintf("page-%d", i), "hpc_job", fmt.Sprintf("job-page-%d", i), 1)
		request.ProviderAddress = provider
		_, err := k.Reserve(ctx, request)
		require.NoError(t, err)
	}

	querier := keeper.NewQuerier(k)
	first, err := querier.ReservationsByProvider(ctx, &resourcesv1.QueryReservationsByProviderRequest{
		ProviderAddress: provider, Pagination: &sdkquery.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, first.Reservations, 2)
	require.NotEmpty(t, first.Pagination.NextKey)

	second, err := querier.ReservationsByProvider(ctx, &resourcesv1.QueryReservationsByProviderRequest{
		ProviderAddress: provider, Pagination: &sdkquery.PageRequest{Key: first.Pagination.NextKey, Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, second.Reservations, 1)
	require.Empty(t, second.Pagination.NextKey)
}

func TestReservationLineagePaginationCursor(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-lineage-pagination", "ignored", 1)
	reservation, err := k.Reserve(ctx, reservationRequest("lineage-pagination", "hpc_job", "job-lineage-pagination", 1))
	require.NoError(t, err)
	_, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-lineage-pagination", HpcJobId: "job-lineage-pagination"})
	require.NoError(t, err)
	_, err = k.ReleaseReservation(ctx, reservation.ReservationId, "lineage_pagination_release")
	require.NoError(t, err)

	querier := keeper.NewQuerier(k)
	first, err := querier.ReservationLineage(ctx, &resourcesv1.QueryReservationLineageRequest{
		ReservationId: reservation.ReservationId, Pagination: &sdkquery.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, first.Events, 2)
	require.NotEmpty(t, first.Pagination.NextKey)
	second, err := querier.ReservationLineage(ctx, &resourcesv1.QueryReservationLineageRequest{
		ReservationId: reservation.ReservationId, Pagination: &sdkquery.PageRequest{Key: first.Pagination.NextKey, Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, second.Events, 1)
	require.Empty(t, second.Pagination.NextKey)
}

func TestReservationLifecycleConservesCapacityAndTerminalIsFinal(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-life", "virtengine1providerlife", 8)

	reservation, err := k.Reserve(ctx, reservationRequest("life-1", "market_lease", "lease-1", 3))
	require.NoError(t, err)

	reservation, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{
		ConsumerType:  "market_lease",
		ConsumerId:    "lease-1",
		MarketOrderId: "order-1",
		MarketBidId:   "bid-1",
		MarketLeaseId: "lease-1",
	})
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateActive, reservation.State)

	reservation, err = k.ConsumeReservation(ctx, reservation.ReservationId, reservation.Capacity, "workload_started")
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateConsumed, reservation.State)

	reservation, err = k.ReleaseReservation(ctx, reservation.ReservationId, "lease_closed")
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateReleased, reservation.State)

	duplicate, err := k.ReleaseReservation(ctx, reservation.ReservationId, "lease_closed")
	require.NoError(t, err)
	require.Equal(t, reservation.ReleasedAt, duplicate.ReleasedAt)

	_, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "market_lease", ConsumerId: "lease-1"})
	require.ErrorIs(t, err, types.ErrInvalidReservationTransition)
	require.NoError(t, k.ValidateCapacityConservation(ctx))

	inv, found := k.GetInventory(ctx, reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
	require.True(t, found)
	require.Equal(t, inv.Total.CpuCores, inv.Available.CpuCores)
}

func TestReservationQuarantineRetainsCapacityUntilSlash(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-hold", "virtengine1providerhold", 4)

	reservation, err := k.Reserve(ctx, reservationRequest("hold-1", "hpc_job", "job-hold", 4))
	require.NoError(t, err)
	reservation, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-hold", HpcJobId: "job-hold"})
	require.NoError(t, err)

	reservation, err = k.QuarantineReservation(ctx, reservation.ReservationId, "dispute_open")
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateQuarantined, reservation.State)
	inv, found := k.GetInventory(ctx, reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
	require.True(t, found)
	require.Zero(t, inv.Available.CpuCores)

	reservation, err = k.SlashReservation(ctx, reservation.ReservationId, "provider_non_fulfillment")
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateSlashed, reservation.State)
	inv, found = k.GetInventory(ctx, reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
	require.True(t, found)
	require.Equal(t, int64(4), inv.Available.CpuCores)
}

func TestReservationFinancialCaseFinalizationClearsBinding(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-financial-case", "ignored", 2)
	reservation, err := k.Reserve(ctx, reservationRequest("financial-case", "hpc_job", "job-financial-case", 2))
	require.NoError(t, err)
	reservation, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-financial-case", HpcJobId: "job-financial-case"})
	require.NoError(t, err)
	reservation, err = k.HoldReservationForFinancialCase(ctx, reservation.ReservationId, "financial-case/84d")
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateDisputed, reservation.State)

	reservation, err = k.FinalizeReservationFinancialCase(ctx, reservation.ReservationId, "financial-case/84d", false)
	require.NoError(t, err)
	require.Equal(t, types.ReservationStateReleased, reservation.State)
	require.Empty(t, reservation.FinancialCaseId)
	require.Zero(t, reservation.PreDisputeState)
	require.Nil(t, reservation.DisputedAt)
	require.Zero(t, reservation.DisputedHeight)
	require.NoError(t, k.ValidateCapacityConservation(ctx))
}

func TestReservationExactRetryRejectsConflictingPayload(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-exact-retry", "virtengine1providerexactretry", 2)

	reservation, err := k.Reserve(ctx, reservationRequest("exact-retry", "hpc_job", "job-exact", 1))
	require.NoError(t, err)
	link := types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-exact", HpcJobId: "job-exact"}
	_, err = k.ActivateReservation(ctx, reservation.ReservationId, link)
	require.NoError(t, err)
	_, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-exact", HpcJobId: "different"})
	require.ErrorIs(t, err, types.ErrLineageConflict)

	_, err = k.ConsumeReservation(ctx, reservation.ReservationId, reservation.Capacity, "started")
	require.NoError(t, err)
	_, err = k.ConsumeReservation(ctx, reservation.ReservationId, reservation.Capacity, "different")
	require.ErrorIs(t, err, types.ErrReservationConflict)
	_, err = k.ReleaseReservation(ctx, reservation.ReservationId, "finished")
	require.NoError(t, err)
	_, err = k.ReleaseReservation(ctx, reservation.ReservationId, "different")
	require.ErrorIs(t, err, types.ErrReservationConflict)
}

func TestReservationDuplicateConsumerAndLineageRollBackCapacity(t *testing.T) {
	k, ctx := setupKeeper(t)
	provider := seedInventory(t, k, ctx, "inv-index-owner", "virtengine1providerindexowner", 2)

	firstRequest := reservationRequest("index-first", "hpc_job", "job-shared", 1)
	firstRequest.HpcJobId = "job-shared"
	_, err := k.Reserve(ctx, firstRequest)
	require.NoError(t, err)

	secondRequest := reservationRequest("index-second", "hpc_job", "job-shared", 1)
	secondRequest.HpcJobId = "job-shared"
	_, err = k.Reserve(ctx, secondRequest)
	require.ErrorIs(t, err, types.ErrLineageConflict)
	inventory, found := k.GetInventory(ctx, provider, types.ResourceClassCompute, "inv-index-owner")
	require.True(t, found)
	require.Equal(t, int64(1), inventory.Available.CpuCores)
	require.NoError(t, k.ValidateCapacityConservation(ctx))
}

func TestSlashReservationExactRetryDoesNotDuplicateEvidence(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-slash-retry", "virtengine1providerslashretry", 1)
	reservation, err := k.Reserve(ctx, reservationRequest("slash-retry", "hpc_job", "job-slash", 1))
	require.NoError(t, err)
	_, err = k.ActivateReservation(ctx, reservation.ReservationId, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: "job-slash", HpcJobId: "job-slash"})
	require.NoError(t, err)
	_, err = k.QuarantineReservation(ctx, reservation.ReservationId, "disputed")
	require.NoError(t, err)
	_, err = k.SlashReservation(ctx, reservation.ReservationId, "failed")
	require.NoError(t, err)
	_, err = k.SlashReservation(ctx, reservation.ReservationId, "failed")
	require.NoError(t, err)
	prefix := append(append([]byte{}, types.SlashingEventKeyPrefix...), []byte(reservation.ReservationId+"\x00")...)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.StoreKey()), prefix)
	defer iterator.Close()
	count := 0
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	require.Equal(t, 1, count)
}

func TestReservationExpirationUsesIndex(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-expiry-index", "virtengine1providerexpiryindex", 2)

	request := reservationRequest("expiry-1", "hpc_job", "job-expiry", 2)
	expires := ctx.BlockTime().Add(time.Second)
	request.ExpiresAt = &expires
	reservation, err := k.Reserve(ctx, request)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second)).WithBlockHeight(ctx.BlockHeight() + 1)
	processed, err := k.ExpireReservations(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(1), processed)

	stored, found := k.GetReservation(ctx, reservation.ReservationId)
	require.True(t, found)
	require.Equal(t, types.ReservationStateExpired, stored.State)
	require.Zero(t, stored.ReleasedHeight)
}

func TestReservationExpiryRequiresFutureTimeAndHeight(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-expiry-bounds", "virtengine1providerexpirybounds", 2)

	request := reservationRequest("expired-time", "hpc_job", "job-expired-time", 1)
	expires := ctx.BlockTime()
	request.ExpiresAt = &expires
	_, err := k.Reserve(ctx, request)
	require.ErrorIs(t, err, types.ErrInvalidRequest)

	request = reservationRequest("expired-height", "hpc_job", "job-expired-height", 1)
	request.ExpiresHeight = ctx.BlockHeight()
	_, err = k.Reserve(ctx, request)
	require.ErrorIs(t, err, types.ErrInvalidRequest)
}

func TestReservationLineageIndexes(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-lineage", "virtengine1providerlineage", 4)

	request := reservationRequest("lineage-1", "market_lease", "lease-lineage", 1)
	request.MarketOrderId = "order-lineage"
	request.MarketBidId = "bid-lineage"
	request.MarketLeaseId = "lease-lineage"
	request.HpcJobId = "job-lineage"
	reservation, err := k.Reserve(ctx, request)
	require.NoError(t, err)

	for kind, id := range map[string]string{
		"order": "order-lineage", "bid": "bid-lineage", "lease": "lease-lineage", "job": "job-lineage", "consumer": "lease-lineage",
	} {
		found, ok := k.GetReservationByLineage(ctx, kind, id)
		require.Truef(t, ok, "%s lineage missing", kind)
		require.Equal(t, reservation.ReservationId, found.ReservationId)
	}
}

func TestReservationByBidQuery(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-bid-query", "virtengine1providerbidquery", 2)
	request := reservationRequest("bid-query", "market_lease", "lease-bid-query", 1)
	request.MarketBidId = "bid-query-id"
	reservation, err := k.Reserve(ctx, request)
	require.NoError(t, err)

	querier := keeper.NewQuerier(k)
	response, err := querier.ReservationByBid(ctx, &resourcesv1.QueryReservationByBidRequest{BidId: "bid-query-id"})
	require.NoError(t, err)
	require.Equal(t, reservation.ReservationId, response.Reservation.ReservationId)
	_, err = querier.ReservationByBid(context.Background(), &resourcesv1.QueryReservationByBidRequest{})
	require.Error(t, err)
}

func TestReservationLineageRejectsMalformedEvent(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-malformed-event", "ignored", 1)
	reservation, err := k.Reserve(ctx, reservationRequest("malformed-event", "hpc_job", "job-malformed-event", 1))
	require.NoError(t, err)
	ctx.KVStore(k.StoreKey()).Set(types.ReservationEventKey(reservation.ReservationId, 2), []byte("not-json"))

	_, err = keeper.NewQuerier(k).ReservationLineage(ctx, &resourcesv1.QueryReservationLineageRequest{ReservationId: reservation.ReservationId})
	require.Error(t, err)
}

func TestCachedContextRollbackDoesNotConsumeLastUnit(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-cache", "virtengine1providercache", 1)

	cacheCtx, _ := ctx.CacheContext()
	_, err := k.Reserve(cacheCtx, reservationRequest("cache-abort", "market_lease", "lease-abort", 1))
	require.NoError(t, err)

	winner, err := k.Reserve(ctx, reservationRequest("cache-win", "hpc_job", "job-win", 1))
	require.NoError(t, err)
	require.NotEmpty(t, winner.ReservationId)
	_, found := k.GetReservation(ctx, "cache-abort")
	require.False(t, found)
}

func TestMarketAndHPCCompeteForLastUnit(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-last", "virtengine1providerlast", 1)

	market, marketErr := k.Reserve(ctx, reservationRequest("market-last", "market_lease", "lease-last", 1))
	hpc, hpcErr := k.Reserve(ctx, reservationRequest("hpc-last", "hpc_job", "job-last", 1))
	require.NoError(t, marketErr)
	require.ErrorIs(t, hpcErr, types.ErrNoEligibleInventory)
	require.NotNil(t, market)
	require.Nil(t, hpc)
	require.NoError(t, k.ValidateCapacityConservation(ctx))
}

func TestReservationGPUTypeRequirementFailsClosed(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.NoError(t, k.SetInventory(ctx, types.ResourceInventory{
		InventoryId: "inv-gpu", ProviderAddress: "virtengine1providergpu", ResourceClass: types.ResourceClassCompute,
		Total: types.ResourceCapacity{Gpus: 1, GpuType: "nvidia-a100"}, Available: types.ResourceCapacity{Gpus: 1, GpuType: "nvidia-a100"},
		Active: true, HeartbeatSequence: 1, LastHeartbeat: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(),
	}))

	request := reservationRequest("gpu-type", "hpc_job", "job-gpu", 0)
	request.Capacity = types.ResourceCapacity{Gpus: 1, GpuType: "nvidia-h100"}
	reservation, err := k.Reserve(ctx, request)
	require.ErrorIs(t, err, types.ErrNoEligibleInventory)
	require.Nil(t, reservation)
}

func TestRandomizedReservationCapacityInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "inv-property", "virtengine1providerproperty", 32)
	rng := rand.New(rand.NewSource(84)) //nolint:gosec // deterministic property sequence
	ids := make([]string, 0, 64)

	for i := 0; i < 2000; i++ {
		switch rng.Intn(6) {
		case 0:
			key := fmt.Sprintf("property-%d", i)
			reservation, err := k.Reserve(ctx, reservationRequest(key, "hpc_job", key, int64(rng.Intn(4)+1)))
			if err == nil {
				ids = append(ids, reservation.ReservationId)
			}
		default:
			if len(ids) == 0 {
				continue
			}
			id := ids[rng.Intn(len(ids))]
			reservation, found := k.GetReservation(ctx, id)
			if !found {
				continue
			}
			switch reservation.State {
			case types.ReservationStatePending:
				if rng.Intn(2) == 0 {
					_, _ = k.ActivateReservation(ctx, id, types.ReservationLink{ConsumerType: "hpc_job", ConsumerId: reservation.ConsumerId, HpcJobId: reservation.ConsumerId})
				} else {
					_, _ = k.ReleaseReservation(ctx, id, "property_release")
				}
			case types.ReservationStateActive:
				if rng.Intn(3) == 0 {
					_, _ = k.ConsumeReservation(ctx, id, reservation.Capacity, "property_consume")
				} else if rng.Intn(2) == 0 {
					_, _ = k.QuarantineReservation(ctx, id, "property_hold")
				} else {
					_, _ = k.ReleaseReservation(ctx, id, "property_release")
				}
			case types.ReservationStateConsumed:
				_, _ = k.ReleaseReservation(ctx, id, "property_release")
			case types.ReservationStateQuarantined:
				_, _ = k.SlashReservation(ctx, id, "property_slash")
			}
		}
		require.NoErrorf(t, k.ValidateCapacityConservation(ctx), "invariant failed after operation %d", i)
	}
}

func reservationRequest(key, consumerType, consumerID string, cpu int64) types.ReservationRequest {
	return types.ReservationRequest{
		IdempotencyKey:   key,
		RequestId:        key,
		RequesterAddress: "virtengine1requesterreservation",
		ProviderAddress:  "",
		ResourceClass:    types.ResourceClassCompute,
		Capacity:         types.ResourceCapacity{CpuCores: cpu},
		ConsumerType:     consumerType,
		ConsumerId:       consumerID,
		Version:          1,
	}
}

func seedInventory(t *testing.T, k interface {
	SetInventory(ctx sdk.Context, inventory types.ResourceInventory) error
}, ctx sdk.Context, inventoryID, _ string, cpu int64) string {
	t.Helper()
	provider := sdk.AccAddress(bytes.Repeat([]byte{byte(len(inventoryID) + 1)}, 20)).String()
	require.NoError(t, k.SetInventory(ctx, types.ResourceInventory{
		InventoryId: inventoryID, ProviderAddress: provider, ResourceClass: types.ResourceClassCompute,
		Total: types.ResourceCapacity{CpuCores: cpu}, Available: types.ResourceCapacity{CpuCores: cpu},
		Active: true, HeartbeatSequence: 1, LastHeartbeat: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(),
	}))
	return provider
}

var _ = resourcesv1.Reservation{}
