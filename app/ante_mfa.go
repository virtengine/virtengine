package app

import (
	"errors"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	dv1beta3 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta3"
	dv1beta4 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"

	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// MFAGatingDecorator enforces MFA for sensitive transactions.
type MFAGatingDecorator struct {
	mfaKeeper mfakeeper.Keeper
}

// NewMFAGatingDecorator creates a new MFA gating decorator.
func NewMFAGatingDecorator(mfaKeeper mfakeeper.Keeper) MFAGatingDecorator {
	return MFAGatingDecorator{mfaKeeper: mfaKeeper}
}

// AnteHandle enforces MFA gating for configured sensitive transactions.
func (d MFAGatingDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return next(ctx, tx, simulate)
	}

	// Cast to SigVerifiableTx to get signers using the new Cosmos SDK v0.50+ interface
	sigTx, ok := tx.(signing.SigVerifiableTx)
	if !ok {
		return ctx, mfatypes.ErrUnauthorized.Wrap("transaction is not signable")
	}

	hooks := mfakeeper.NewMFAGatingHooks(d.mfaKeeper)

	for _, msg := range msgs {
		if err := d.checkMFAGating(ctx, hooks, msg, sigTx); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (d MFAGatingDecorator) checkMFAGating(ctx sdk.Context, hooks mfakeeper.MFAGatingHooks, msg sdk.Msg, sigTx signing.SigVerifiableTx) error {
	transactionType, ok := resolveSensitiveTxType(msg)
	if !ok || !isMFAEnforcedTx(transactionType) {
		return nil
	}

	signer, err := firstSigner(sigTx)
	if err != nil {
		return err
	}

	_, required, requiredCombinations := hooks.RequiresMFA(ctx, signer, transactionType)
	if !required {
		return nil
	}

	if !d.shouldEnforceValueThreshold(ctx, transactionType, msg) {
		return nil
	}

	proofProvider, ok := getMFAProofProvider(msg)
	if !ok {
		emitMFARequiredEvent(ctx, signer, transactionType, requiredCombinations, mfatypes.AttributeValueFailure, "mfa_proof_missing", "")
		return requiredFactorsError(transactionType, requiredCombinations)
	}

	return validateMFAForSigner(ctx, hooks, signer, transactionType, requiredCombinations, proofProvider)
}

func isMFAEnforcedTx(transactionType mfatypes.SensitiveTransactionType) bool {
	switch transactionType {
	case mfatypes.SensitiveTxAccountRecovery,
		mfatypes.SensitiveTxKeyRotation,
		mfatypes.SensitiveTxProviderRegistration,
		mfatypes.SensitiveTxLargeWithdrawal,
		mfatypes.SensitiveTxValidatorRegistration,
		mfatypes.SensitiveTxHighValueOrder:
		return true
	default:
		return false
	}
}

func firstSigner(sigTx signing.SigVerifiableTx) (sdk.AccAddress, error) {
	signers, err := sigTx.GetSigners()
	if err != nil {
		return nil, mfatypes.ErrUnauthorized.Wrapf("failed to get signers: %v", err)
	}
	if len(signers) == 0 {
		return nil, mfatypes.ErrUnauthorized.Wrap("no signers for sensitive transaction")
	}
	return signers[0], nil
}

func getMFAProofProvider(msg sdk.Msg) (mfatypes.MFAProofProvider, bool) {
	proofProvider, ok := msg.(mfatypes.MFAProofProvider)
	return proofProvider, ok
}

func validateMFAForSigner(
	ctx sdk.Context,
	hooks mfakeeper.MFAGatingHooks,
	signer sdk.AccAddress,
	transactionType mfatypes.SensitiveTransactionType,
	requiredCombinations []mfatypes.FactorCombination,
	proofProvider mfatypes.MFAProofProvider,
) error {
	deviceFingerprint := proofProvider.GetDeviceFingerprint()
	proof := proofProvider.GetMFAProof()

	canBypass, reducedFactors := hooks.CanBypassMFA(ctx, signer, transactionType, deviceFingerprint)
	if canBypass && reducedFactors == nil {
		emitMFARequiredEvent(ctx, signer, transactionType, requiredCombinations, mfatypes.AttributeValueSuccess, "trusted_device_bypass", deviceFingerprint)
		return nil
	}

	emitMFARequiredEvent(ctx, signer, transactionType, requiredCombinations, mfatypes.AttributeValuePending, "mfa_required", deviceFingerprint)

	if reducedFactors != nil {
		requiredCombinations = []mfatypes.FactorCombination{*reducedFactors}
	}

	if err := hooks.ValidateMFAProof(ctx, signer, transactionType, proof, deviceFingerprint); err != nil {
		if errors.Is(err, mfatypes.ErrInsufficientFactors) || errors.Is(err, mfatypes.ErrMFARequired) {
			return requiredFactorsError(transactionType, requiredCombinations)
		}
		return err
	}

	return nil
}

func resolveSensitiveTxType(msg sdk.Msg) (mfatypes.SensitiveTransactionType, bool) {
	switch msg.(type) {
	case *rolestypes.MsgSetAccountState:
		return mfatypes.SensitiveTxAccountRecovery, true
	case *veidtypes.MsgRebindWallet:
		return mfatypes.SensitiveTxKeyRotation, true
	case *banktypes.MsgSend:
		return mfatypes.SensitiveTxLargeWithdrawal, true
	case *dv1beta3.MsgCreateDeployment, *dv1beta4.MsgCreateDeployment:
		return mfatypes.SensitiveTxHighValueOrder, true
	default:
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == "" {
			return mfatypes.SensitiveTxUnspecified, false
		}
		return mfatypes.GetSensitiveTransactionType(typeURL)
	}
}

func (d MFAGatingDecorator) shouldEnforceValueThreshold(
	ctx sdk.Context,
	transactionType mfatypes.SensitiveTransactionType,
	msg sdk.Msg,
) bool {
	config, found := d.mfaKeeper.GetSensitiveTxConfig(ctx, transactionType)
	if !found || config == nil || config.ValueThreshold == "" {
		return true
	}

	threshold, ok := sdkmath.NewIntFromString(config.ValueThreshold)
	if !ok {
		return true
	}

	amount, _, ok := extractTransactionAmount(msg)
	if !ok {
		return true
	}

	return amount.GTE(threshold)
}

func extractTransactionAmount(msg sdk.Msg) (sdkmath.Int, string, bool) {
	switch m := msg.(type) {
	case *banktypes.MsgSend:
		return selectCoinAmount(m.Amount)
	case *dv1beta3.MsgCreateDeployment:
		return m.Deposit.Amount, m.Deposit.Denom, true
	case *dv1beta4.MsgCreateDeployment:
		return depositAmount(m.Deposit)
	default:
		return sdkmath.Int{}, "", false
	}
}

func depositAmount(deposit depositv1.Deposit) (sdkmath.Int, string, bool) {
	return deposit.Amount.Amount, deposit.Amount.Denom, true
}

func selectCoinAmount(coins sdk.Coins) (sdkmath.Int, string, bool) {
	if len(coins) == 0 {
		return sdkmath.Int{}, "", false
	}

	for _, coin := range coins {
		if coin.Denom == "uve" {
			return coin.Amount, coin.Denom, true
		}
	}

	coin := coins[0]
	return coin.Amount, coin.Denom, true
}

func requiredFactorsError(
	transactionType mfatypes.SensitiveTransactionType,
	requiredCombinations []mfatypes.FactorCombination,
) error {
	required := formatFactorCombinations(requiredCombinations)
	if required == "" {
		return mfatypes.ErrMFARequired.Wrapf("MFA required for %s", transactionType.String())
	}
	return mfatypes.ErrMFARequired.Wrapf("MFA required for %s. Required factors: %s", transactionType.String(), required)
}

func formatFactorCombinations(combinations []mfatypes.FactorCombination) string {
	if len(combinations) == 0 {
		return ""
	}

	formatted := make([]string, 0, len(combinations))
	for _, combo := range combinations {
		if len(combo.Factors) == 0 {
			continue
		}
		names := make([]string, 0, len(combo.Factors))
		for _, factor := range combo.Factors {
			names = append(names, factor.String())
		}
		formatted = append(formatted, strings.Join(names, "+"))
	}

	return strings.Join(formatted, " or ")
}

func emitMFARequiredEvent(
	ctx sdk.Context,
	signer sdk.AccAddress,
	transactionType mfatypes.SensitiveTransactionType,
	requiredCombinations []mfatypes.FactorCombination,
	status string,
	reason string,
	deviceFingerprint string,
) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(mfatypes.AttributeKeyAccountAddress, signer.String()),
		sdk.NewAttribute(mfatypes.AttributeKeyTransactionType, transactionType.String()),
	}
	if status != "" {
		attrs = append(attrs, sdk.NewAttribute(mfatypes.AttributeKeyStatus, status))
	}
	if reason != "" {
		attrs = append(attrs, sdk.NewAttribute(mfatypes.AttributeKeyReason, reason))
	}
	if deviceFingerprint != "" {
		attrs = append(attrs, sdk.NewAttribute(mfatypes.AttributeKeyDeviceFingerprint, deviceFingerprint))
	}
	if required := formatFactorCombinations(requiredCombinations); required != "" {
		attrs = append(attrs, sdk.NewAttribute(mfatypes.AttributeKeyVerifiedFactors, required))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(mfatypes.EventTypeMFARequired, attrs...),
	)
}
