// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Exercise descriptor-backed SDK registration, not only the Go type spelling.
func TestComputeCreditContracts(t *testing.T) {
	registry := cdctypes.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	for _, msg := range []sdk.Msg{
		&MsgMintVCC{Owner: "owner", To: "owner", CoinsToBurn: sdk.NewCoin(NativeDenom, math.NewInt(100))},
		&MsgBurnVCC{Owner: "owner", To: "owner", CoinsToBurn: sdk.NewCoin(ComputeCreditDenom, math.NewInt(100))},
	} {
		t.Run(sdk.MsgTypeURL(msg), func(t *testing.T) {
			packed, err := cdctypes.NewAnyWithValue(msg)
			require.NoError(t, err)
			require.Contains(t, packed.TypeUrl, "VCC")
			var decoded sdk.Msg
			wire := &cdctypes.Any{TypeUrl: packed.TypeUrl, Value: packed.Value}
			require.NoError(t, cdc.UnpackAny(wire, &decoded))
			require.Equal(t, msg, decoded)
			encoded, err := cdc.MarshalJSON(msg)
			require.NoError(t, err)
			require.Contains(t, string(encoded), "100")
		})
	}
}
