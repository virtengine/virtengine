// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package market

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
)

func TestValidateGenesisRejectsActiveLeaseWithoutReservation(t *testing.T) {
	owner := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	provider := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()
	lease := marketv1.Lease{ID: marketv1.LeaseID{Owner: owner, DSeq: 1, GSeq: 1, OSeq: 1, Provider: provider, BSeq: 1}, State: marketv1.LeaseActive}
	genesis := &marketv1beta5.GenesisState{Params: marketv1beta5.DefaultParams(), Leases: marketv1.Leases{lease}}

	require.ErrorContains(t, ValidateGenesis(genesis), "has no authoritative reservation")
	genesis.Leases[0].ReservationId = "reservation/lease"
	require.NoError(t, ValidateGenesis(genesis))
}
