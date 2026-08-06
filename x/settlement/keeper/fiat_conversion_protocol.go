package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	fiatConversionProtocolVersion uint32 = 1
	terminalPolicyCompleted              = "external_fiat_completed_custody_sink"
	terminalPolicyManualReview           = "manual_review_value_held"
	terminalPolicyCancelNoSwap           = "cancel_before_swap_value_held"
	fiatIndexProvider                    = "provider"
	fiatIndexState                       = "state"
)

var fiatAmountPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)

type FiatConversionObservationResult struct {
	Conversion        types.FiatConversionRecord
	ExactDuplicate    bool
	ObservationDigest []byte
}

// RecordFiatConversionObservation atomically validates and applies one signed
// off-chain observation. SDK transaction authentication supplies the signer;
// the keeper additionally binds that signer to the conversion provider.
func (k Keeper) RecordFiatConversionObservation(ctx sdk.Context, msg *types.MsgRecordFiatConversionObservation) (*FiatConversionObservationResult, error) {
	cacheCtx, write := ctx.CacheContext()
	result, err := k.recordFiatConversionObservation(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return result, nil
}

func (k Keeper) recordFiatConversionObservation(ctx sdk.Context, msg *types.MsgRecordFiatConversionObservation) (*FiatConversionObservationResult, error) {
	if msg == nil {
		return nil, types.ErrFiatObservationEvidence.Wrap("observation required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	conversion, found := k.GetFiatConversion(ctx, msg.ConversionId)
	if !found {
		return nil, types.ErrFiatConversionNotFound.Wrapf("conversion %s not found", msg.ConversionId)
	}
	if conversion.LegacyQuarantined {
		return nil, types.ErrFiatConversionQuarantined.Wrap(conversion.QuarantineReason)
	}
	if msg.Sender != conversion.Provider {
		return nil, types.ErrUnauthorized.Wrap("observation signer is not conversion provider")
	}
	digest, err := fiatObservationDigest(msg)
	if err != nil {
		return nil, types.ErrFiatObservationEvidence.Wrap(err.Error())
	}
	store := ctx.KVStore(k.skey)
	replayKey := types.FiatObservationReplayKey(conversion.ConversionID, msg.IdempotencyKey)
	if existing := store.Get(replayKey); existing != nil {
		if bytes.Equal(existing, digest) {
			return &FiatConversionObservationResult{Conversion: conversion, ExactDuplicate: true, ObservationDigest: digest}, nil
		}
		return nil, types.ErrFiatObservationReplayConflict
	}
	sequenceKey := types.FiatObservationSequenceKey(conversion.ConversionID, msg.ObservationSequence)
	if existing := store.Get(sequenceKey); existing != nil {
		if bytes.Equal(existing, digest) {
			return &FiatConversionObservationResult{Conversion: conversion, ExactDuplicate: true, ObservationDigest: digest}, nil
		}
		return nil, types.ErrFiatObservationReplayConflict.Wrap("sequence already bound to different payload")
	}
	if conversion.State.IsTerminal() {
		return nil, types.ErrInvalidStateTransition.Wrap("terminal conversion rejects new observations")
	}
	if msg.ObservationSequence != conversion.ObservationSequence+1 {
		return nil, types.ErrFiatObservationSequence.Wrapf("expected %d got %d", conversion.ObservationSequence+1, msg.ObservationSequence)
	}

	params := k.GetParams(ctx)
	if err := validateImmutableFiatProfileCommitments(conversion, msg); err != nil {
		return nil, err
	}
	newSideEffect := fiatObservationAuthorizesNewSideEffect(msg.Stage)
	if newSideEffect {
		if err := validateCurrentFiatProfileCommitments(params, conversion, msg); err != nil {
			return nil, err
		}
	}
	if len(conversion.Observations) >= int(params.FiatConversionMaxObservations) {
		return nil, types.ErrFiatObservationEvidence.Wrap("observation history limit reached")
	}
	now := ctx.BlockTime().Unix()
	if params.FiatConversionObservationMaxPastSeconds > math.MaxInt64 || params.FiatConversionObservationMaxFutureSeconds > math.MaxInt64 {
		return nil, types.ErrFiatObservationEvidence.Wrap("observation time bounds overflow")
	}
	pastSeconds := int64(params.FiatConversionObservationMaxPastSeconds)     //nolint:gosec // checked against MaxInt64 above
	futureSeconds := int64(params.FiatConversionObservationMaxFutureSeconds) //nolint:gosec // checked against MaxInt64 above
	if msg.ObservedAt < now-pastSeconds || msg.ObservedAt > now+futureSeconds {
		return nil, types.ErrFiatObservationEvidence.Wrap("observed_at outside block-time bounds")
	}
	if err := k.ensureFiatConversionNoActiveCase(ctx, conversion); err != nil {
		if newSideEffect {
			return nil, err
		}
	}
	if err := validateCommittedFiatCompliance(conversion, msg.ComplianceDecisionHash); err != nil {
		return nil, err
	}
	if newSideEffect {
		if err := k.validateCurrentFiatCompliance(ctx, conversion, msg.ComplianceDecisionHash); err != nil {
			return nil, err
		}
	}
	if err := validateFiatObservationStage(params, &conversion, msg, now); err != nil {
		return nil, err
	}
	if err := k.applyFiatObservation(ctx, &conversion, msg); err != nil {
		return nil, err
	}

	lineageInput := append(append([]byte(nil), conversion.LastObservationDigest...), digest...)
	lineage := sha256.Sum256(lineageInput)
	conversion.ObservationSequence = msg.ObservationSequence
	conversion.LastObservationDigest = append([]byte(nil), digest...)
	conversion.Observations = append(conversion.Observations, types.FiatConversionObservation{
		Sequence: msg.ObservationSequence, IdempotencyKey: append([]byte(nil), msg.IdempotencyKey...),
		Stage: msg.Stage, Status: msg.Status, ObservedAt: msg.ObservedAt, RecordedAt: now,
		RecordedHeight: ctx.BlockHeight(), EvidenceHash: append([]byte(nil), msg.EvidenceHash...),
		ObservationDigest: append([]byte(nil), digest...), LineageDigest: lineage[:], FailureCode: msg.FailureCode,
	})
	conversion.EvidenceHash = append([]byte(nil), msg.EvidenceHash...)
	if err := k.SetFiatConversion(ctx, conversion); err != nil {
		return nil, err
	}
	store.Set(replayKey, digest)
	store.Set(sequenceKey, digest)
	if err := ctx.EventManager().EmitTypedEvent(&settlementv1.EventFiatConversionObservationRecorded{
		ConversionId: conversion.ConversionID, Provider: conversion.Provider,
		ObservationSequence: conversion.ObservationSequence, Stage: msg.Stage, State: string(conversion.State),
		ObservationDigest: digest, RecordedHeight: ctx.BlockHeight(),
	}); err != nil {
		return nil, err
	}
	if conversion.State.IsTerminal() {
		if err := ctx.EventManager().EmitTypedEvent(&settlementv1.EventFiatConversionTerminal{
			ConversionId: conversion.ConversionID, PayoutId: conversion.PayoutID,
			Stage: msg.Stage, TerminalPolicy: conversion.TerminalPolicy, EvidenceHash: append([]byte(nil), msg.EvidenceHash...),
		}); err != nil {
			return nil, err
		}
	}
	return &FiatConversionObservationResult{Conversion: conversion, ObservationDigest: digest}, nil
}

func validateImmutableFiatProfileCommitments(conversion types.FiatConversionRecord, msg *types.MsgRecordFiatConversionObservation) error {
	if msg.DexProfileId != conversion.DEXProfileID || msg.PayoutProfileId != conversion.PayoutProfileID ||
		!bytes.Equal(msg.DexProfileDigest, conversion.DEXProfileDigest) || !bytes.Equal(msg.PayoutProfileDigest, conversion.PayoutProfileDigest) {
		return types.ErrFiatProfileCommitment.Wrap("observation differs from accepted profile commitments")
	}
	return nil
}

func validateCurrentFiatProfileCommitments(params types.Params, conversion types.FiatConversionRecord, msg *types.MsgRecordFiatConversionObservation) error {
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	if !params.FiatConversionEnabled || params.FiatConversionDEXProfileState != certified || params.FiatConversionPayoutProfileState != certified {
		return types.ErrFiatProfileCommitment.Wrap("fiat conversion profiles are not currently certified_enabled")
	}
	if msg.DexProfileId != conversion.DEXProfileID || msg.PayoutProfileId != conversion.PayoutProfileID ||
		msg.DexProfileId != params.FiatConversionDEXProfileID || msg.PayoutProfileId != params.FiatConversionPayoutProfileID ||
		!bytes.Equal(msg.DexProfileDigest, conversion.DEXProfileDigest) || !bytes.Equal(msg.PayoutProfileDigest, conversion.PayoutProfileDigest) ||
		!bytes.Equal(msg.DexProfileDigest, params.FiatConversionDEXProfileDigest) || !bytes.Equal(msg.PayoutProfileDigest, params.FiatConversionPayoutProfileDigest) {
		return types.ErrFiatProfileCommitment
	}
	return nil
}

func validateCommittedFiatCompliance(conversion types.FiatConversionRecord, reported []byte) error {
	if len(reported) != sha256.Size || !bytes.Equal(reported, conversion.ComplianceDecisionHash) {
		return types.ErrComplianceRequired.Wrap("compliance decision digest mismatch")
	}
	return nil
}

func (k Keeper) validateCurrentFiatCompliance(ctx sdk.Context, conversion types.FiatConversionRecord, reported []byte) error {
	if err := validateCommittedFiatCompliance(conversion, reported); err != nil {
		return err
	}
	if k.complianceKeeper == nil {
		return types.ErrComplianceRequired.Wrap("compliance keeper not configured")
	}
	record, found := k.complianceKeeper.GetComplianceRecord(ctx, conversion.Provider)
	if !found || record == nil || record.IsExpired(ctx.BlockTime().Unix()) || !strings.EqualFold(record.Status.String(), conversion.ComplianceStatus) {
		return types.ErrComplianceRequired.Wrap("compliance decision revoked or expired")
	}
	digest, err := complianceDecisionDigest(record)
	if err != nil || !bytes.Equal(digest, reported) {
		return types.ErrComplianceRequired.Wrap("current compliance decision changed")
	}
	return nil
}

// fiatObservationAuthorizesNewSideEffect distinguishes current authorization
// from immutable evidence. Quote observations can cause the worker to sign or
// initiate a new external operation and therefore require current governance,
// profile, compliance, and hold authorization. Submission/finality/failure
// observations only reconcile an already-crossed external boundary against the
// commitments accepted when the conversion was created.
func fiatObservationAuthorizesNewSideEffect(stage settlementv1.FiatConversionObservationStage) bool {
	switch stage {
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED,
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED:
		return true
	default:
		return false
	}
}

func fiatConversionCrossedIrreversibleBoundary(conversion types.FiatConversionRecord) bool {
	if conversion.SwapTxHash != "" || conversion.OffRampID != "" || conversion.ValueMovementApplied {
		return true
	}
	switch conversion.State {
	case types.FiatConversionStateSwapSubmitted,
		types.FiatConversionStateSwapSettled,
		types.FiatConversionStateOffRampPending,
		types.FiatConversionStatePayoutPending,
		types.FiatConversionStatePayoutSubmitted,
		types.FiatConversionStatePayoutCompleted:
		return true
	default:
		return false
	}
}

func validateFiatObservationStage(params types.Params, conversion *types.FiatConversionRecord, msg *types.MsgRecordFiatConversionObservation, blockTime int64) error {
	if conversion == nil {
		return types.ErrFiatConversionNotFound
	}
	requireHash := func(value []byte, name string) error {
		if len(value) != 32 {
			return types.ErrFiatObservationEvidence.Wrapf("%s must be SHA-256", name)
		}
		return nil
	}
	requireCoin := func(coin sdk.Coin, denom string, positive bool, name string) error {
		if !coin.IsValid() || coin.Denom != denom || (positive && !coin.IsPositive()) {
			return types.ErrFiatObservationEvidence.Wrapf("invalid %s", name)
		}
		return nil
	}
	status := strings.ToLower(strings.TrimSpace(msg.Status))
	switch msg.Stage {
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED:
		replacement := conversion.State == types.FiatConversionStateSwapPending
		if conversion.State != types.FiatConversionStateCreated && !replacement || requireHash(msg.QuoteDigest, "quote_digest") != nil || msg.QuoteExpiry <= blockTime {
			return types.ErrInvalidStateTransition.Wrap("invalid quote acceptance")
		}
		if replacement && (conversion.SwapTxHash != "" || len(conversion.QuoteDigest) != sha256.Size || conversion.QuoteExpiry <= 0 || blockTime < conversion.QuoteExpiry || bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest)) {
			return types.ErrInvalidStateTransition.Wrap("replacement quote requires an expired, unsubmitted prior quote")
		}
		if err := requireCoin(msg.MinimumStableOutput, conversion.StableToken.Denom, true, "minimum stable output"); err != nil {
			return err
		}
		if status != "accepted" {
			return types.ErrFiatObservationEvidence.Wrap("quote status must be accepted")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED:
		if conversion.State != types.FiatConversionStateSwapPending || msg.SwapTxHash == "" || msg.ObservedAt >= conversion.QuoteExpiry ||
			!bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest) || !msg.MinimumStableOutput.IsEqual(conversion.MinimumStableOutput) {
			return types.ErrInvalidStateTransition.Wrap("swap submission detached from accepted quote")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED:
		if conversion.State != types.FiatConversionStateSwapSubmitted || msg.SwapTxHash != conversion.SwapTxHash || msg.SwapHeight <= 0 ||
			requireHash(msg.SwapBlockHash, "swap_block_hash") != nil || requireHash(msg.SwapFinalityHash, "swap_finality_hash") != nil ||
			msg.SwapFinalityConfirmations < params.FiatConversionMinSwapFinalityConfirmations || !bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest) ||
			!msg.MinimumStableOutput.IsEqual(conversion.MinimumStableOutput) {
			return types.ErrFiatObservationEvidence.Wrap("swap finality evidence insufficient")
		}
		if err := requireCoin(msg.StableAmount, conversion.StableToken.Denom, true, "stable amount"); err != nil {
			return err
		}
		if msg.StableAmount.Amount.LT(conversion.MinimumStableOutput.Amount) {
			return types.ErrFiatObservationEvidence.Wrap("stable amount below minimum output")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED:
		replacement := conversion.State == types.FiatConversionStatePayoutPending
		if conversion.State != types.FiatConversionStateSwapSettled && !replacement || msg.OffRampQuoteId == "" || msg.QuoteExpiry <= blockTime || requireHash(msg.QuoteDigest, "payout quote digest") != nil {
			return types.ErrInvalidStateTransition.Wrap("invalid payout quote")
		}
		if replacement && (conversion.OffRampID != "" || conversion.OffRampReference != "" || len(conversion.PrivacySafeReferenceHash) != 0 ||
			conversion.OffRampQuoteID == "" || conversion.QuoteExpiry <= 0 || blockTime < conversion.QuoteExpiry ||
			msg.OffRampQuoteId == conversion.OffRampQuoteID || bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest)) {
			return types.ErrInvalidStateTransition.Wrap("replacement payout quote requires an expired, unsubmitted, distinct prior quote")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED:
		if conversion.State != types.FiatConversionStatePayoutPending || msg.OffRampQuoteId != conversion.OffRampQuoteID || msg.OffRampPayoutId == "" || msg.ObservedAt >= conversion.QuoteExpiry || !bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest) || requireHash(msg.PrivacySafeReferenceHash, "privacy_safe_reference_hash") != nil {
			return types.ErrInvalidStateTransition.Wrap("payout submission detached from quote")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED:
		if conversion.State != types.FiatConversionStatePayoutSubmitted || msg.OffRampPayoutId != conversion.OffRampID || status != "completed" ||
			msg.OffRampQuoteId != conversion.OffRampQuoteID || !bytes.Equal(msg.QuoteDigest, conversion.QuoteDigest) ||
			!validPositiveFiatAmount(msg.FiatAmount) || requireHash(msg.PrivacySafeReferenceHash, "privacy_safe_reference_hash") != nil || !bytes.Equal(msg.PrivacySafeReferenceHash, conversion.PrivacySafeReferenceHash) || requireHash(msg.PayoutFinalityHash, "payout_finality_hash") != nil {
			return types.ErrFiatObservationEvidence.Wrap("invalid terminal payout evidence")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED:
		if msg.FailureCode == "" || len(msg.FailureCode) > 64 || !boundedCode(msg.FailureCode) {
			return types.ErrFiatObservationEvidence.Wrap("bounded failure code required")
		}
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED:
		if conversion.State != types.FiatConversionStateCreated && conversion.State != types.FiatConversionStateSwapPending {
			return types.ErrInvalidStateTransition.Wrap("cancellation allowed only before swap submission")
		}
		if msg.FailureCode == "" || len(msg.FailureCode) > 64 || !boundedCode(msg.FailureCode) {
			return types.ErrFiatObservationEvidence.Wrap("bounded cancellation code required")
		}
	default:
		return types.ErrInvalidStateTransition.Wrap("unsupported fiat observation stage")
	}
	return nil
}

func (k Keeper) applyFiatObservation(ctx sdk.Context, conversion *types.FiatConversionRecord, msg *types.MsgRecordFiatConversionObservation) error {
	observed := time.Unix(msg.ObservedAt, 0).UTC()
	switch msg.Stage {
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED:
		conversion.QuoteDigest = append([]byte(nil), msg.QuoteDigest...)
		conversion.QuoteExpiry = msg.QuoteExpiry
		conversion.MinimumStableOutput = msg.MinimumStableOutput
		conversion.SwapStatus = "quote_accepted"
		return conversion.MarkSwapPending(observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED:
		if err := k.releaseCanonicalPayoutHoldAtIrreversibleBoundary(ctx, *conversion); err != nil {
			return err
		}
		conversion.SwapTxHash = msg.SwapTxHash
		return conversion.MarkSwapSubmitted(hex.EncodeToString(conversion.QuoteDigest), observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED:
		conversion.SwapHeight = msg.SwapHeight
		conversion.SwapBlockHash = append([]byte(nil), msg.SwapBlockHash...)
		conversion.SwapFinalityConfirmations = msg.SwapFinalityConfirmations
		conversion.SwapFinalityHash = append([]byte(nil), msg.SwapFinalityHash...)
		return conversion.MarkSwapSettled(msg.SwapTxHash, msg.StableAmount, observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED:
		if _, err := k.requireLinkedPendingFiatPayout(ctx, *conversion); err != nil {
			return err
		}
		replacement := conversion.State == types.FiatConversionStatePayoutPending
		conversion.OffRampQuoteID = msg.OffRampQuoteId
		conversion.QuoteDigest = append([]byte(nil), msg.QuoteDigest...)
		conversion.QuoteExpiry = msg.QuoteExpiry
		if replacement {
			conversion.OffRampStatus = "quote_replaced"
			conversion.AddAuditEntry("expired_payout_quote_replaced", conversion.Provider, "", map[string]string{
				"offramp_quote_id": msg.OffRampQuoteId,
			}, observed)
			return nil
		}
		if err := conversion.MarkOffRampPending(observed); err != nil {
			return err
		}
		return conversion.MarkPayoutPending(observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED:
		if _, err := k.requireLinkedPendingFiatPayout(ctx, *conversion); err != nil {
			return err
		}
		conversion.PrivacySafeReferenceHash = append([]byte(nil), msg.PrivacySafeReferenceHash...)
		return conversion.MarkPayoutSubmitted(msg.OffRampPayoutId, msg.Status, "", observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED:
		payout, err := k.requireLinkedPendingFiatPayout(ctx, *conversion)
		if err != nil {
			return err
		}
		if err := payout.MarkProcessing(observed); err != nil {
			return err
		}
		if err := payout.MarkCompleted("", observed); err != nil {
			return err
		}
		effectHash := fiatCustodyEffectHash(*conversion, *payout, msg.PayoutFinalityHash)
		if existing := ctx.KVStore(k.skey).Get(types.FiatCustodyEffectKey(conversion.ConversionID)); existing != nil {
			if bytes.Equal(existing, effectHash) {
				return types.ErrInvalidStateTransition.Wrap("fiat custody effect already applied")
			}
			return types.ErrFiatObservationReplayConflict.Wrap("fiat custody effect bound to different payload")
		}
		if err := k.validateRetainedTreasuryEntries(ctx, payout); err != nil {
			return err
		}
		amount := sdk.NewCoins(conversion.CryptoAmount)
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, types.FiatConversionCustodyAccountName, amount); err != nil {
			return types.ErrPayoutExecutionFailed.Wrapf("move fiat payout value to custody sink: %v", err)
		}
		payout.ExternalFinalityHash = append([]byte(nil), msg.PayoutFinalityHash...)
		payout.ValueMovementApplied = true
		payout.ValueMovementEffectHash = append([]byte(nil), effectHash...)
		if err := k.SetPayout(ctx, *payout); err != nil {
			return err
		}
		if err := k.recordPayoutRetainedTreasuryEntries(ctx, payout); err != nil {
			return err
		}
		k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryCompleted,
			types.PayoutStatePending, types.PayoutStateCompleted, amount,
			"external fiat payout authenticated; native value moved to governed custody sink", "fiat_conversion_observation")
		conversion.FiatAmount = msg.FiatAmount
		conversion.PayoutFinalityHash = append([]byte(nil), msg.PayoutFinalityHash...)
		conversion.TerminalPolicy = terminalPolicyCompleted
		conversion.ValueMovementApplied = true
		conversion.CustodySinkAmount = conversion.CryptoAmount
		conversion.CustodySinkEffectHash = append([]byte(nil), effectHash...)
		ctx.KVStore(k.skey).Set(types.FiatCustodyEffectKey(conversion.ConversionID), effectHash)
		return conversion.MarkPayoutCompleted(observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED:
		conversion.TerminalPolicy = terminalPolicyManualReview
		return conversion.MarkFailed(msg.FailureCode, observed)
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED:
		if err := k.releaseFiatDailyQuota(ctx, conversion); err != nil {
			return err
		}
		conversion.TerminalPolicy = terminalPolicyCancelNoSwap
		return conversion.MarkCancelled(msg.FailureCode, observed)
	}
	return types.ErrInvalidStateTransition
}

func (k Keeper) releaseCanonicalPayoutHoldAtIrreversibleBoundary(ctx sdk.Context, conversion types.FiatConversionRecord) error {
	payout, err := k.requireLinkedFiatPayout(ctx, conversion)
	if err != nil {
		return err
	}
	if payout.State != types.PayoutStateHeld {
		return nil
	}
	if !strings.HasPrefix(payout.DisputeID, "financial-case/") {
		return types.ErrFinancialCaseHold.Wrap("fiat payout is held by a noncanonical owner")
	}
	financialCase, found := k.GetFinancialCase(ctx, payout.DisputeID)
	if !found || !types.IsActiveFinancialCaseStatus(financialCase.Status) || financialCase.Exposure.PayoutId != payout.PayoutID {
		return types.ErrFinancialCaseHold.Wrap("fiat payout canonical hold is malformed")
	}
	if financialCase.ActiveHoldCount <= 1 {
		return types.ErrFinancialCaseHold.Wrap("irreversible boundary would leave canonical incident without a local hold")
	}
	payout.State = types.PayoutStatePending
	payout.DisputeID = ""
	payout.HoldReason = ""
	if err := k.SetPayout(ctx, *payout); err != nil {
		return err
	}
	financialCase.ActiveHoldCount--
	return k.SetFinancialCase(ctx, financialCase)
}

func fiatCustodyEffectHash(conversion types.FiatConversionRecord, payout types.PayoutRecord, finalityHash []byte) []byte {
	hash := sha256.New()
	for _, value := range []string{
		"virtengine/settlement/fiat-custody-sink/v1", conversion.ConversionID, payout.PayoutID,
		conversion.CryptoAmount.Denom, conversion.CryptoAmount.Amount.String(), types.ModuleAccountName,
		types.FiatConversionCustodyAccountName,
	} {
		writeCanonicalFiatBytes(hash, []byte(value))
	}
	writeCanonicalFiatBytes(hash, finalityHash)
	return hash.Sum(nil)
}

func (k Keeper) requireLinkedPendingFiatPayout(ctx sdk.Context, conversion types.FiatConversionRecord) (*types.PayoutRecord, error) {
	payout, err := k.requireLinkedFiatPayout(ctx, conversion)
	if err != nil {
		return nil, err
	}
	if payout.State != types.PayoutStatePending {
		return nil, types.ErrInvalidStateTransition.Wrapf("linked payout state is %s", payout.State)
	}
	return payout, nil
}

func (k Keeper) requireLinkedFiatPayout(ctx sdk.Context, conversion types.FiatConversionRecord) (*types.PayoutRecord, error) {
	if conversion.PayoutID == "" {
		return nil, types.ErrPayoutNotFound.Wrap("fiat conversion is not linked to a payout")
	}
	payout, found := k.GetPayout(ctx, conversion.PayoutID)
	if !found || payout.FiatConversionID != conversion.ConversionID || payout.Provider != conversion.Provider || payout.Customer != conversion.Customer || payout.SettlementID != conversion.SettlementID ||
		payout.InvoiceID != conversion.InvoiceID || len(payout.NetAmount) != 1 || !payout.NetAmount[0].IsEqual(conversion.CryptoAmount) {
		return nil, types.ErrInvalidPayout.Wrap("fiat conversion payout lineage mismatch")
	}
	return &payout, nil
}

func (k Keeper) ensureFiatConversionNoActiveCase(ctx sdk.Context, conversion types.FiatConversionRecord) error {
	for _, entry := range []struct{ kind, value string }{
		{"invoice", conversion.InvoiceID}, {"settlement", conversion.SettlementID}, {"order", conversion.OrderID},
		{"escrow", conversion.EscrowID}, {"lease", conversion.LeaseID},
	} {
		if caseID, held := k.HasActiveFinancialCase(ctx, entry.kind, entry.value); held {
			return types.ErrDisputeActive.Wrapf("%s held by canonical case %s", entry.kind, caseID)
		}
	}
	return nil
}

func fiatObservationDigest(msg *types.MsgRecordFiatConversionObservation) ([]byte, error) {
	clone := *msg
	bz, err := proto.Marshal(&clone)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(bz)
	return digest[:], nil
}

func canonicalFiatRequestDigest(request types.FiatConversionRequest, params types.Params, complianceDigest []byte) ([]byte, error) {
	if request.EncryptedPayload == nil || len(request.EncryptedPayload.EnvelopeHash) != sha256.Size {
		return nil, types.ErrInvalidParams.Wrap("canonical request requires encrypted envelope digest")
	}
	slippage, err := types.ValidateExactSlippage(request.SlippageToleranceExact, request.SlippageTolerance, false)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("virtengine/settlement/fiat-request/v2"))
	for _, value := range []string{
		request.InvoiceID, request.SettlementID, request.PayoutID, request.Provider, request.Customer, request.RequestedBy,
		request.CryptoAmount.Denom, request.CryptoAmount.Amount.String(), request.FiatCurrency, request.PaymentMethod,
		request.DestinationHash, request.DestinationRegion, params.FiatConversionDEXProfileID, params.FiatConversionPayoutProfileID,
		slippage, request.CryptoToken.Symbol, request.CryptoToken.Denom, request.CryptoToken.ChainID,
		request.StableToken.Symbol, request.StableToken.Denom, request.StableToken.ChainID,
	} {
		writeCanonicalFiatBytes(hash, []byte(value))
	}
	var decimals [8]byte
	binary.BigEndian.PutUint32(decimals[:4], uint32(request.CryptoToken.Decimals))
	binary.BigEndian.PutUint32(decimals[4:], uint32(request.StableToken.Decimals))
	writeCanonicalFiatBytes(hash, decimals[:])
	writeCanonicalFiatBytes(hash, request.EncryptedPayload.EnvelopeHash)
	writeCanonicalFiatBytes(hash, params.FiatConversionDEXProfileDigest)
	writeCanonicalFiatBytes(hash, params.FiatConversionPayoutProfileDigest)
	writeCanonicalFiatBytes(hash, complianceDigest)
	return hash.Sum(nil), nil
}

func writeCanonicalFiatBytes(hash interface{ Write([]byte) (int, error) }, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) //nolint:gosec // all fields are protocol-bounded well below uint32
	_, _ = hash.Write(length[:])
	_, _ = hash.Write(value)
}

func dailyFiatBucket(now time.Time) string { return now.UTC().Format("20060102") }

func complianceDecisionDigest(record *veidtypes.ComplianceRecord) ([]byte, error) {
	if record == nil {
		return nil, fmt.Errorf("compliance record required")
	}
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(bz)
	return digest[:], nil
}

func boundedCode(value string) bool {
	for _, char := range value {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}

func validPositiveFiatAmount(value string) bool {
	if !fiatAmountPattern.MatchString(value) || len(value) > 64 {
		return false
	}
	amount, err := sdkmath.LegacyNewDecFromStr(value)
	return err == nil && amount.IsPositive()
}

func encodeDailyTotal(value sdkmath.Int) []byte { return []byte(value.String()) }

func decodeDailyTotal(value []byte) (sdkmath.Int, error) {
	if len(value) == 0 {
		return sdkmath.ZeroInt(), nil
	}
	result, ok := sdkmath.NewIntFromString(string(value))
	if !ok || result.IsNegative() {
		return sdkmath.Int{}, types.ErrInvalidSettlement.Wrap("malformed fiat daily accounting total")
	}
	return result, nil
}

func (k Keeper) releaseFiatDailyQuota(ctx sdk.Context, conversion *types.FiatConversionRecord) error {
	if conversion == nil || !conversion.DailyQuotaReserved {
		return nil
	}
	store := ctx.KVStore(k.skey)
	key := types.FiatDailyTotalKey(conversion.Provider, conversion.DailyBucket)
	total, err := decodeDailyTotal(store.Get(key))
	if err != nil || total.LT(conversion.CryptoAmount.Amount) {
		return types.ErrInvalidSettlement.Wrap("fiat daily accounting underflow")
	}
	remaining := total.Sub(conversion.CryptoAmount.Amount)
	if remaining.IsZero() {
		store.Delete(key)
	} else {
		store.Set(key, encodeDailyTotal(remaining))
	}
	conversion.DailyQuotaReserved = false
	return nil
}
