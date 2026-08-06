package compatibility

import (
	"reflect"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

func TestTask85BLegacyFiatConversionWireRemainsStable(t *testing.T) {
	fixture := []byte{0x0a, 0x01, 0x63, 0x42, 0x01, 0x70, 0x4a, 0x01, 0x75}
	var record settlementv1.FiatConversionRecord
	require.NoError(t, proto.Unmarshal(fixture, &record))
	require.Equal(t, "c", record.ConversionId)
	require.Equal(t, "p", record.Provider)
	require.Equal(t, "u", record.Customer)
	require.Zero(t, record.ProtocolVersion)
	require.Zero(t, record.ObservationSequence)
	require.Empty(t, record.DexProfileId)
	require.Empty(t, record.PayoutProfileId)
	require.Empty(t, record.RequestDigest)
}

func TestTask85BObservationTypeURLs(t *testing.T) {
	for message, expected := range map[proto.Message]string{
		&settlementv1.MsgRecordFiatConversionObservation{}: "/virtengine.settlement.v1.MsgRecordFiatConversionObservation",
		&settlementv1.MsgUpdateParams{}:                    "/virtengine.settlement.v1.MsgUpdateParams",
	} {
		packed, err := codectypes.NewAnyWithValue(message)
		require.NoError(t, err)
		require.Equal(t, expected, packed.TypeUrl)
	}
}

func TestTask85BAdditiveFieldNumbers(t *testing.T) {
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.PayoutRecord{}), map[string]int{
		"payout_id": 1, "block_height": 26, "external_finality_hash": 27,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.FiatConversionRecord{}), map[string]int{
		"conversion_id": 1, "transition_history": 45, "protocol_version": 46,
		"observation_sequence": 47, "last_observation_digest": 48, "observations": 49,
		"dex_profile_id": 50, "payout_profile_id": 52, "request_digest": 65,
		"daily_bucket": 66, "value_movement_applied": 70,
		"slippage_tolerance_exact": 71, "daily_quota_reserved": 72,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.Params{}), map[string]int{
		"financial_case_timeout_batch_limit": 49, "fiat_conversion_dex_profile_id": 50,
		"fiat_conversion_dex_profile_digest": 51, "fiat_conversion_dex_profile_state": 52,
		"fiat_conversion_payout_profile_id": 53, "fiat_conversion_payout_profile_digest": 54,
		"fiat_conversion_payout_profile_state": 55, "fiat_conversion_max_observations": 59,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.MsgRecordFiatConversionObservation{}), map[string]int{
		"sender": 1, "conversion_id": 2, "observation_sequence": 3, "idempotency_key": 4,
		"stage": 5, "dex_profile_id": 6, "payout_profile_id": 8, "evidence_hash": 26,
		"failure_code": 27, "payout_finality_hash": 28,
	})
}
