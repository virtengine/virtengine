package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFiatConversionStateMachineHappyPath(t *testing.T) {
	now := time.Now().UTC()
	record := NewFiatConversionRecord(
		"conv-1",
		FiatConversionRequest{
			InvoiceID:       "inv-1",
			SettlementID:    "set-1",
			PayoutID:        "pay-1",
			Provider:        sdk.AccAddress("provider-1").String(),
			Customer:        sdk.AccAddress("customer-1").String(),
			RequestedBy:     "provider-1",
			CryptoAmount:    sdk.NewCoin("uve", sdkmath.NewInt(100)),
			FiatCurrency:    "USD",
			PaymentMethod:   "bank_transfer",
			DestinationHash: "dest",
			CryptoToken:     TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
			StableToken:     TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
			EncryptedPayload: &EncryptedSettlementPayload{
				EnvelopeRef: "enc-ref",
			},
		},
		sdk.NewCoin("uve", sdkmath.NewInt(100)),
		now,
	)

	require.NoError(t, record.MarkSwapPending(now))
	require.NoError(t, record.MarkSwapSubmitted("quote-1", now))
	require.NoError(t, record.MarkSwapSettled("swap-tx", sdk.NewCoin("uusdc", sdkmath.NewInt(90)), now))
	require.NoError(t, record.MarkOffRampPending(now))
	require.NoError(t, record.MarkPayoutPending(now))
	require.NoError(t, record.MarkPayoutSubmitted("off-1", "processing", "ref-1", now))
	require.NoError(t, record.MarkPayoutCompleted(now))

	require.Equal(t, FiatConversionStatePayoutCompleted, record.State)
	require.True(t, record.State.IsTerminal())
	require.NotEmpty(t, record.IdempotencyKey)
	require.GreaterOrEqual(t, len(record.TransitionHistory), 7)
}

func TestFiatConversionStateMachineRejectsInvalidTransition(t *testing.T) {
	now := time.Now().UTC()
	record := &FiatConversionRecord{State: FiatConversionStateCreated}
	err := record.MarkPayoutCompleted(now)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid fiat conversion transition")
}

func TestFiatConversionStateMachineLegacyNormalization(t *testing.T) {
	now := time.Now().UTC()
	record := &FiatConversionRecord{
		ConversionID:    "legacy-1",
		Provider:        sdk.AccAddress("provider-1").String(),
		Customer:        sdk.AccAddress("customer-1").String(),
		State:           FiatConversionState("completed"),
		CryptoAmount:    sdk.NewCoin("uve", sdkmath.NewInt(1)),
		StableAmount:    sdk.NewCoin("uusdc", sdkmath.NewInt(1)),
		FiatCurrency:    "USD",
		PaymentMethod:   "bank_transfer",
		DestinationHash: "dest",
		CryptoToken:     TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:     TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		RequestedAt:     now,
		UpdatedAt:       now,
	}
	require.NoError(t, record.Validate())
	require.Equal(t, FiatConversionStatePayoutCompleted, record.State)
}
