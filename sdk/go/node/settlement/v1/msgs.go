// Package v1 provides additional methods for generated settlement types.
package v1

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgCreateEscrow

func (msg *MsgCreateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidEscrow.Wrap("order_id cannot be empty")
	}

	// Convert to sdk.Coins for validation
	amount := sdk.NewCoins(msg.Amount...)
	if !amount.IsValid() || amount.IsZero() {
		return ErrInvalidAmount.Wrap("amount must be valid and non-zero")
	}

	if msg.ExpiresIn == 0 {
		return ErrInvalidEscrow.Wrap("expires_in must be greater than zero")
	}

	return nil
}

func (msg *MsgCreateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgActivateEscrow

func (msg *MsgActivateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.LeaseId == "" {
		return ErrInvalidEscrow.Wrap("lease_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrInvalidAddress.Wrap("invalid recipient address")
	}

	return nil
}

func (msg *MsgActivateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgReleaseEscrow

func (msg *MsgReleaseEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if len(msg.Amount) > 0 {
		amount := sdk.NewCoins(msg.Amount...)
		if !amount.IsValid() {
			return ErrInvalidAmount.Wrap("amount must be valid")
		}
	}

	return nil
}

func (msg *MsgReleaseEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgRefundEscrow

func (msg *MsgRefundEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg *MsgRefundEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgDisputeEscrow

func (msg *MsgDisputeEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg *MsgDisputeEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgSettleOrder

func (msg *MsgSettleOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidSettlement.Wrap("order_id cannot be empty")
	}

	return nil
}

func (msg *MsgSettleOrder) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgRecordUsage

func (msg *MsgRecordUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidUsageRecord.Wrap("order_id cannot be empty")
	}

	if msg.LeaseId == "" {
		return ErrInvalidUsageRecord.Wrap("lease_id cannot be empty")
	}

	if msg.UsageUnits == 0 {
		return ErrInvalidUsageRecord.Wrap("usage_units must be greater than zero")
	}

	if msg.UsageType == "" {
		return ErrInvalidUsageRecord.Wrap("usage_type cannot be empty")
	}

	if msg.PeriodEnd <= msg.PeriodStart {
		return ErrInvalidUsageRecord.Wrap("period_end must be after period_start")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}
	if msg.SignatureVersion != 0 {
		if msg.SignatureVersion != 1 || msg.ChainId == "" || msg.PricingVersion == 0 || msg.FormulaVersion == 0 || msg.ModelVersion == 0 ||
			msg.StreamSequence == 0 || len(msg.Nonce) != 32 || len(msg.IdempotencyKey) != 32 ||
			msg.ProviderKeyEpoch == 0 || msg.ProviderKeyId == "" {
			return ErrInvalidUsageRecord.Wrap("incomplete versioned usage proof")
		}
	}

	return nil
}

func (msg *MsgRecordUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgAcknowledgeUsage

func (msg *MsgAcknowledgeUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.UsageId == "" {
		return ErrInvalidUsageRecord.Wrap("usage_id cannot be empty")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}
	if msg.SignatureVersion != 0 {
		if msg.SignatureVersion != 1 || len(msg.UsageDigest) != 32 || len(msg.ReplayKey) != 32 {
			return ErrInvalidSignature.Wrap("incomplete versioned acknowledgment proof")
		}
	}

	return nil
}

func (msg *MsgAcknowledgeUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgClaimRewards

func (msg *MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Source is optional, no validation needed if empty
	return nil
}

func (msg *MsgClaimRewards) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic validates the bounded, privacy-safe observation envelope.
func (msg *MsgRecordFiatConversionObservation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}
	if msg.ConversionId == "" || len(msg.ConversionId) > 128 || msg.ObservationSequence == 0 {
		return ErrInvalidSettlement.Wrap("conversion id and positive observation sequence required")
	}
	if len(msg.IdempotencyKey) != 32 || msg.Stage == FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_UNSPECIFIED {
		return ErrInvalidSettlement.Wrap("versioned observation stage and 32-byte idempotency key required")
	}
	for _, value := range []string{msg.DexProfileId, msg.PayoutProfileId} {
		if value == "" || len(value) > 128 {
			return ErrInvalidSettlement.Wrap("bounded profile ids required")
		}
	}
	if len(msg.DexProfileDigest) != 32 || len(msg.PayoutProfileDigest) != 32 || len(msg.EvidenceHash) != 32 {
		return ErrInvalidSettlement.Wrap("profile and evidence digests must be SHA-256")
	}
	if len(msg.ComplianceDecisionHash) != 32 {
		return ErrInvalidSettlement.Wrap("compliance decision digest must be SHA-256")
	}
	for _, value := range []string{msg.SwapTxHash, msg.OffRampQuoteId, msg.OffRampPayoutId, msg.Status, msg.FiatAmount, msg.FailureCode} {
		if len(value) > 128 || strings.ContainsAny(value, "\r\n\x00") {
			return ErrInvalidSettlement.Wrap("observation identifier is not bounded")
		}
	}
	if msg.ObservedAt <= 0 {
		return ErrInvalidSettlement.Wrap("observed_at must be positive")
	}
	switch msg.Stage {
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED:
		if len(msg.QuoteDigest) != 32 || msg.QuoteExpiry <= 0 || !msg.MinimumStableOutput.IsValid() || !msg.MinimumStableOutput.IsPositive() {
			return ErrInvalidSettlement.Wrap("quote acceptance evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED:
		if msg.SwapTxHash == "" || len(msg.QuoteDigest) != 32 {
			return ErrInvalidSettlement.Wrap("swap submission evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED:
		if msg.SwapTxHash == "" || len(msg.QuoteDigest) != 32 || !msg.MinimumStableOutput.IsValid() || !msg.MinimumStableOutput.IsPositive() ||
			msg.SwapHeight <= 0 || len(msg.SwapBlockHash) != 32 || len(msg.SwapFinalityHash) != 32 || !msg.StableAmount.IsValid() || !msg.StableAmount.IsPositive() {
			return ErrInvalidSettlement.Wrap("swap finality evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED:
		if msg.OffRampQuoteId == "" || len(msg.QuoteDigest) != 32 || msg.QuoteExpiry <= 0 {
			return ErrInvalidSettlement.Wrap("payout quote evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED:
		if msg.OffRampQuoteId == "" || msg.OffRampPayoutId == "" || len(msg.PrivacySafeReferenceHash) != 32 {
			return ErrInvalidSettlement.Wrap("payout submission evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED:
		if msg.OffRampQuoteId == "" || msg.OffRampPayoutId == "" || len(msg.QuoteDigest) != 32 || len(msg.PrivacySafeReferenceHash) != 32 || len(msg.PayoutFinalityHash) != 32 || msg.FiatAmount == "" {
			return ErrInvalidSettlement.Wrap("payout completion evidence incomplete")
		}
	case FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED,
		FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED:
		if msg.FailureCode == "" {
			return ErrInvalidSettlement.Wrap("failure code required")
		}
	}
	return nil
}

func (msg *MsgRecordFiatConversionObservation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}
	return nil
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
