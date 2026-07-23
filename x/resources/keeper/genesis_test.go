// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	resources "github.com/virtengine/virtengine/x/resources"
	"github.com/virtengine/virtengine/x/resources/types"
)

func TestGenesisRoundTripRebuildsTerminalLineageAndIdempotencyIndexes(t *testing.T) {
	k, ctx := setupKeeper(t)
	provider := sdk.AccAddress(bytes.Repeat([]byte{31}, 20)).String()
	requester := sdk.AccAddress(bytes.Repeat([]byte{32}, 20)).String()
	require.NoError(t, k.SetInventory(ctx, types.ResourceInventory{
		InventoryId: "genesis-inventory", ProviderAddress: provider, ResourceClass: types.ResourceClassCompute,
		Total: types.ResourceCapacity{CpuCores: 1}, Available: types.ResourceCapacity{CpuCores: 1},
		Active: true, HeartbeatSequence: 1, LastHeartbeat: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(),
	}))
	request := types.ReservationRequest{
		IdempotencyKey: "genesis-idempotency", RequestId: "genesis-request", RequesterAddress: requester,
		ProviderAddress: provider, InventoryId: "genesis-inventory", ResourceClass: types.ResourceClassCompute,
		Capacity: types.ResourceCapacity{CpuCores: 1}, ConsumerType: "hpc_job", ConsumerId: "genesis-job",
		HpcJobId: "genesis-job", Version: 1,
	}
	reservation, err := k.Reserve(ctx, request)
	require.NoError(t, err)
	_, err = k.ReleaseReservation(ctx, reservation.ReservationId, "genesis_release")
	require.NoError(t, err)
	k.ActivateCanonicalReservations(ctx)

	exported := resources.ExportGenesis(ctx, k)
	require.True(t, exported.CanonicalReservationsActive)
	require.Len(t, exported.Reservations, 1)
	require.Len(t, exported.ReservationEvents, 2)

	restoredKeeper, restoredCtx := setupKeeper(t)
	resources.InitGenesis(restoredCtx, restoredKeeper, exported)
	byJob, found := restoredKeeper.GetReservationByLineage(restoredCtx, "job", "genesis-job")
	require.True(t, found)
	require.Equal(t, reservation.ReservationId, byJob.ReservationId)
	replayed, err := restoredKeeper.Reserve(restoredCtx, request)
	require.NoError(t, err)
	require.Equal(t, reservation.ReservationId, replayed.ReservationId)
	require.Equal(t, types.ReservationStateReleased, replayed.State)
	events, err := restoredKeeper.ReservationEvents(restoredCtx, reservation.ReservationId)
	require.NoError(t, err)
	require.Len(t, events, 2)
}
