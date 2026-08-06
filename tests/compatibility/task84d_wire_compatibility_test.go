// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package compatibility

import (
	"reflect"
	"regexp"
	"strconv"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

func TestTask84DLegacySettlementMessageWireAndTypeURL(t *testing.T) {
	// Frozen pre-84D wire: sender="s", escrow_id="e", reason="r", evidence="h".
	fixture := []byte{0x0a, 0x01, 0x73, 0x12, 0x01, 0x65, 0x1a, 0x01, 0x72, 0x22, 0x01, 0x68}
	var msg settlementv1.MsgDisputeEscrow
	require.NoError(t, proto.Unmarshal(fixture, &msg))
	require.Equal(t, "s", msg.Sender)
	require.Equal(t, "e", msg.EscrowId)
	require.Equal(t, "r", msg.Reason)
	require.Equal(t, "h", msg.Evidence)

	roundTrip, err := proto.Marshal(&msg)
	require.NoError(t, err)
	require.Equal(t, fixture, roundTrip)
	packed, err := codectypes.NewAnyWithValue(&msg)
	require.NoError(t, err)
	require.Equal(t, "/virtengine.settlement.v1.MsgDisputeEscrow", packed.TypeUrl)
}

func TestTask84DFinancialCaseAdditivePartyFieldsPreserveEarlierWire(t *testing.T) {
	// Frozen earlier Task 84D aggregate prefix: version=1, case_id="x". The
	// additive provider/customer fields 36/37 must decode absent and must not
	// perturb the old bytes when unset.
	fixture := []byte{0x08, 0x01, 0x12, 0x01, 0x78}
	var financialCase settlementv1.FinancialCase
	require.NoError(t, proto.Unmarshal(fixture, &financialCase))
	require.Equal(t, uint32(1), financialCase.Version)
	require.Equal(t, "x", financialCase.CaseId)
	require.Empty(t, financialCase.Provider)
	require.Empty(t, financialCase.Customer)

	roundTrip, err := proto.Marshal(&financialCase)
	require.NoError(t, err)
	var decodedAgain settlementv1.FinancialCase
	require.NoError(t, proto.Unmarshal(roundTrip, &decodedAgain))
	require.Empty(t, decodedAgain.Provider)
	require.Empty(t, decodedAgain.Customer)
}

func TestTask84DCanonicalMessageTypeURL(t *testing.T) {
	tests := []struct {
		message proto.Message
		typeURL string
	}{
		{&settlementv1.MsgOpenFinancialCase{}, "/virtengine.settlement.v1.MsgOpenFinancialCase"},
		{&settlementv1.MsgAddFinancialClaim{}, "/virtengine.settlement.v1.MsgAddFinancialClaim"},
		{&settlementv1.MsgSubmitFinancialCaseForReview{}, "/virtengine.settlement.v1.MsgSubmitFinancialCaseForReview"},
		{&settlementv1.MsgEscalateFinancialCase{}, "/virtengine.settlement.v1.MsgEscalateFinancialCase"},
		{&settlementv1.MsgResolveFinancialCase{}, "/virtengine.settlement.v1.MsgResolveFinancialCase"},
		{&settlementv1.MsgAppealFinancialCase{}, "/virtengine.settlement.v1.MsgAppealFinancialCase"},
		{&settlementv1.MsgCancelFinancialCase{}, "/virtengine.settlement.v1.MsgCancelFinancialCase"},
		{&settlementv1.MsgFinalizeFinancialCase{}, "/virtengine.settlement.v1.MsgFinalizeFinancialCase"},
	}
	for _, test := range tests {
		packed, err := codectypes.NewAnyWithValue(test.message)
		require.NoError(t, err)
		require.Equal(t, test.typeURL, packed.TypeUrl)
	}
}

func TestTask84DCanonicalFieldNumbers(t *testing.T) {
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.FinancialSubject{}), map[string]int{
		"type": 1, "primary_id": 2, "order_id": 3, "invoice_id": 4, "usage_id": 5,
		"hpc_job_id": 6, "settlement_id": 7, "escrow_id": 8, "reservation_id": 9, "lease_id": 10,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.FinancialCase{}), map[string]int{
		"version": 1, "case_id": 2, "subject": 3, "claims": 4, "claimant": 5,
		"respondent": 6, "exposure": 7, "status": 8, "provider": 36, "customer": 37,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.MsgOpenFinancialCase{}), map[string]int{
		"sender": 1, "subject": 2, "claim_type": 3, "respondent": 4, "evidence_hash": 5,
		"encrypted_reference": 6, "idempotency_key": 7, "source_module": 8, "source_reference": 9,
	})
	assertFieldNumbers(t, reflect.TypeOf(settlementv1.FinancialAppeal{}), map[string]int{
		"appeal_id": 1, "appellant": 2, "evidence_hash": 3, "encrypted_reference": 4,
		"created_height": 5, "created_at": 6, "idempotency_key": 7,
	})
}

func assertFieldNumbers(t *testing.T, message reflect.Type, expected map[string]int) {
	t.Helper()
	for name, number := range expected {
		found := false
		for i := 0; i < message.NumField(); i++ {
			field := message.Field(i)
			match := regexp.MustCompile(`^[^,]+,(\d+),`).FindStringSubmatch(field.Tag.Get("protobuf"))
			if len(match) != 2 || field.Tag.Get("json") == "" {
				continue
			}
			jsonName := regexp.MustCompile(`^([^,]+)`).FindStringSubmatch(field.Tag.Get("json"))[1]
			if jsonName != name {
				continue
			}
			actual, err := strconv.Atoi(match[1])
			require.NoError(t, err)
			require.Equal(t, number, actual, name)
			found = true
			break
		}
		require.True(t, found, name)
	}
}
