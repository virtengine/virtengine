// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package compatibility_test

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	providerv1 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

func TestTask84BLegacyMsgRecordUsageWireFixture(t *testing.T) {
	legacy := &settlementv1.MsgRecordUsage{
		Sender:      "virtengine1provider",
		OrderId:     "order-1",
		LeaseId:     "lease-1",
		UsageUnits:  7,
		UsageType:   "cpu",
		PeriodStart: 100,
		PeriodEnd:   200,
		UnitPrice:   sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(2)),
		Signature:   []byte{1, 2, 3},
	}
	encoded, err := proto.Marshal(legacy)
	require.NoError(t, err)
	const golden = "0a1376697274656e67696e653170726f766964657212076f726465722d311a076c656173652d3120072a03637075306438c801421a0a037576651213323030303030303030303030303030303030304a03010203"
	require.Equal(t, golden, hex.EncodeToString(encoded))

	var decoded settlementv1.MsgRecordUsage
	require.NoError(t, proto.Unmarshal(encoded, &decoded))
	require.Equal(t, uint32(0), decoded.SignatureVersion)
	require.Zero(t, decoded.StreamSequence)
	require.Empty(t, decoded.IdempotencyKey)
}

func TestTask84BLegacyUsageRecordWireFixture(t *testing.T) {
	legacy := &settlementv1.UsageRecord{
		UsageId:           "usage-1",
		OrderId:           "order-1",
		LeaseId:           "lease-1",
		Provider:          "virtengine1provider",
		Customer:          "virtengine1customer",
		UsageUnits:        7,
		UsageType:         "cpu",
		PeriodStart:       100,
		PeriodEnd:         200,
		UnitPrice:         sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(2)),
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(14))),
		ProviderSignature: []byte{1, 2, 3},
	}
	encoded, err := proto.Marshal(legacy)
	require.NoError(t, err)
	const golden = "0a0775736167652d3112076f726465722d311a076c656173652d31221376697274656e67696e653170726f76696465722a1376697274656e67696e6531637573746f6d657230073a03637075406448c801521a0a037576651213323030303030303030303030303030303030305a090a03757665120231346203010203"
	require.Equal(t, golden, hex.EncodeToString(encoded))

	var decoded settlementv1.UsageRecord
	require.NoError(t, proto.Unmarshal(encoded, &decoded))
	require.False(t, decoded.SignatureVerified)
	require.False(t, decoded.LegacyUnverified)
	require.Empty(t, decoded.UsageDigest)
}

func TestTask84BProviderGenesisOldFixtureStillDecodes(t *testing.T) {
	legacy := &providerv1.GenesisState{Providers: providerv1.Providers{}}
	encoded, err := proto.Marshal(legacy)
	require.NoError(t, err)
	var decoded providerv1.GenesisState
	require.NoError(t, proto.Unmarshal(encoded, &decoded))
	require.Empty(t, decoded.SigningKeys)
}

func TestTask84BFieldNumbersAreAdditive(t *testing.T) {
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.MsgRecordUsage{}), "Signature", "9")
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.MsgRecordUsage{}), "AllocationId", "10")
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.MsgRecordUsage{}), "SignatureVersion", "25")
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.UsageRecord{}), "Metadata", "19")
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.UsageRecord{}), "AllocationId", "20")
	assertProtoFieldNumber(t, reflect.TypeOf(settlementv1.UsageRecord{}), "SignatureMaterialRedacted", "46")
	assertProtoFieldNumber(t, reflect.TypeOf(providerv1.GenesisState{}), "Providers", "1")
	assertProtoFieldNumber(t, reflect.TypeOf(providerv1.GenesisState{}), "SigningKeys", "2")
}

func assertProtoFieldNumber(t *testing.T, message reflect.Type, fieldName, expected string) {
	t.Helper()
	field, found := message.FieldByName(fieldName)
	require.True(t, found, "%s.%s missing", message.Name(), fieldName)
	tag := field.Tag.Get("protobuf")
	parts := strings.Split(tag, ",")
	require.GreaterOrEqual(t, len(parts), 2, "invalid protobuf tag %q", tag)
	require.Equal(t, expected, parts[1])
}
