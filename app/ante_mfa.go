package app

import (
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	proofProvider, err := getMFAProofProvider(msg)
	if err != nil {
		return err
	}

	return validateMFAForSigner(ctx, hooks, signer, transactionType, proofProvider)
}

func isMFAEnforcedTx(transactionType mfatypes.SensitiveTransactionType) bool {
	switch transactionType {
	case mfatypes.SensitiveTxAccountRecovery, mfatypes.SensitiveTxKeyRotation:
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

func getMFAProofProvider(msg sdk.Msg) (mfatypes.MFAProofProvider, error) {
	proofProvider, ok := msg.(mfatypes.MFAProofProvider)
	if !ok {
		return nil, mfatypes.ErrMFARequired.Wrap("MFA proof missing for sensitive transaction")
	}
	return proofProvider, nil
}

func validateMFAForSigner(
	ctx sdk.Context,
	hooks mfakeeper.MFAGatingHooks,
	signer sdk.AccAddress,
	transactionType mfatypes.SensitiveTransactionType,
	proofProvider mfatypes.MFAProofProvider,
) error {
	_, required, _ := hooks.RequiresMFA(ctx, signer, transactionType)
	if !required {
		return nil
	}

	deviceFingerprint := proofProvider.GetDeviceFingerprint()
	proof := proofProvider.GetMFAProof()

	canBypass, reducedFactors := hooks.CanBypassMFA(ctx, signer, transactionType, deviceFingerprint)
	if canBypass && reducedFactors == nil {
		return nil
	}

	return hooks.ValidateMFAProof(ctx, signer, transactionType, proof, deviceFingerprint)
}

func resolveSensitiveTxType(msg sdk.Msg) (mfatypes.SensitiveTransactionType, bool) {
	switch msg.(type) {
	case *rolestypes.MsgSetAccountState:
		return mfatypes.SensitiveTxAccountRecovery, true
	case *veidtypes.MsgRebindWallet:
		return mfatypes.SensitiveTxKeyRotation, true
	default:
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == "" {
			return mfatypes.SensitiveTxUnspecified, false
		}
		return mfatypes.GetSensitiveTransactionType(typeURL)
	}
}
