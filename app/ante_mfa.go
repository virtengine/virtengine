package app

import (
	"errors"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/cosmos/gogoproto/proto"

	dv1beta3 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta3"
	dv1beta4 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	mfapb "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"

	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
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

	hooks := mfakeeper.NewMFAGatingHooks(d.mfaKeeper)
	sigTx, _ := tx.(signing.SigVerifiableTx)

	for _, msg := range msgs {
		if err := d.checkMFAGating(ctx, hooks, msg, sigTx); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (d MFAGatingDecorator) checkMFAGating(ctx sdk.Context, hooks mfakeeper.MFAGatingHooks, msg sdk.Msg, sigTx signing.SigVerifiableTx) error {
	transactionType, ok := d.resolveSensitiveTxType(ctx, msg)
	if !ok || !isMFAEnforcedTx(transactionType) {
		return nil
	}

	signers, err := signersForMsg(msg, sigTx)
	if err != nil {
		return err
	}

	if !d.shouldEnforceValueThreshold(ctx, transactionType, msg) {
		return nil
	}

	proofProvider, ok, err := getMFAProofProvider(msg)
	if err != nil {
		return err
	}

	for _, signer := range signers {
		_, required, requiredCombinations := hooks.RequiresMFA(ctx, signer, transactionType)
		if !required {
			continue
		}
		if !ok {
			emitMFARequiredEvent(ctx, signer, transactionType, requiredCombinations, mfatypes.AttributeValueFailure, "mfa_proof_missing", "")
			return requiredFactorsError(transactionType, requiredCombinations)
		}
		if err := validateMFAForSigner(ctx, hooks, signer, transactionType, requiredCombinations, proofProvider); err != nil {
			return err
		}
	}

	return nil
}

func isMFAEnforcedTx(transactionType mfatypes.SensitiveTransactionType) bool {
	return transactionType.IsValid()
}

func signersForMsg(msg sdk.Msg, sigTx signing.SigVerifiableTx) ([]sdk.AccAddress, error) {
	type signerProvider interface {
		GetSigners() []sdk.AccAddress
	}

	msgWithSigners, ok := msg.(signerProvider)
	if ok {
		signers := msgWithSigners.GetSigners()
		if len(signers) == 0 {
			return nil, mfatypes.ErrUnauthorized.Wrap("no signers for sensitive transaction")
		}
		return signers, nil
	}

	if sigTx == nil {
		return nil, mfatypes.ErrUnauthorized.Wrap("sensitive transaction does not expose account signers")
	}

	signerBytes, err := sigTx.GetSigners()
	if err != nil {
		return nil, mfatypes.ErrUnauthorized.Wrapf("failed to resolve transaction signers: %v", err)
	}
	if len(signerBytes) == 0 {
		return nil, mfatypes.ErrUnauthorized.Wrap("no signers for sensitive transaction")
	}

	signers := make([]sdk.AccAddress, 0, len(signerBytes))
	for _, signer := range signerBytes {
		if len(signer) == 0 {
			return nil, mfatypes.ErrUnauthorized.Wrap("encountered empty signer for sensitive transaction")
		}
		signers = append(signers, sdk.AccAddress(signer))
	}

	return signers, nil
}

func getMFAProofProvider(msg sdk.Msg) (mfatypes.MFAProofProvider, bool, error) {
	proofProvider, ok := msg.(mfatypes.MFAProofProvider)
	if ok {
		return proofProvider, true, nil
	}

	switch typed := msg.(type) {
	case *veidv1.MsgRebindWallet:
		proof, err := decodeSerializedMFAProof(typed.GetMfaProof())
		if err != nil {
			return nil, true, mfatypes.ErrInvalidProof.Wrapf("invalid serialized MFA proof: %v", err)
		}
		return newAdaptedMFAProofProvider(proof, typed.GetDeviceFingerprint(), ""), true, nil
	case interface {
		GetMfaProof() *mfapb.MFAProof
		GetDeviceFingerprint() string
	}:
		return newAdaptedMFAProofProvider(convertProtoMFAProof(typed.GetMfaProof()), typed.GetDeviceFingerprint(), ""), true, nil
	case interface {
		GetMfaProof() []byte
		GetDeviceFingerprint() string
	}:
		proof, err := decodeSerializedMFAProof(typed.GetMfaProof())
		if err != nil {
			return nil, true, mfatypes.ErrInvalidProof.Wrapf("invalid serialized MFA proof: %v", err)
		}
		return newAdaptedMFAProofProvider(proof, typed.GetDeviceFingerprint(), ""), true, nil
	default:
		return nil, false, nil
	}
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
	trustToken := ""
	if proof != nil {
		trustToken = proof.TrustToken
	}

	canBypass, reducedFactors := hooks.CanBypassMFA(ctx, signer, transactionType, deviceFingerprint, trustToken)
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

func (d MFAGatingDecorator) resolveSensitiveTxType(ctx sdk.Context, msg sdk.Msg) (mfatypes.SensitiveTransactionType, bool) {
	switch msg.(type) {
	case *veidv1.MsgRebindWallet:
		return mfatypes.SensitiveTxKeyRotation, true
	case *dv1beta3.MsgCreateDeployment, *dv1beta4.MsgCreateDeployment:
		return mfatypes.SensitiveTxHighValueOrder, true
	case *banktypes.MsgSend:
		return d.resolveBankSendSensitiveTxType(ctx, msg)
	default:
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == "" {
			return mfatypes.SensitiveTxUnspecified, false
		}
		return mfatypes.GetSensitiveTransactionType(typeURL)
	}
}

func (d MFAGatingDecorator) resolveBankSendSensitiveTxType(ctx sdk.Context, msg sdk.Msg) (mfatypes.SensitiveTransactionType, bool) {
	amount, ok := extractTransactionAmount(msg)
	if !ok {
		return mfatypes.SensitiveTxUnspecified, false
	}

	for _, txType := range []mfatypes.SensitiveTransactionType{
		mfatypes.SensitiveTxLargeWithdrawal,
		mfatypes.SensitiveTxMediumWithdrawal,
	} {
		config := d.sensitiveTxConfig(ctx, txType)
		if config == nil || !config.Enabled {
			continue
		}
		if config.ValueThreshold == "" {
			return txType, true
		}
		threshold, ok := sdkmath.NewIntFromString(config.ValueThreshold)
		if !ok {
			return txType, true
		}
		if amount.GTE(threshold) {
			return txType, true
		}
	}

	return mfatypes.SensitiveTxUnspecified, false
}

func (d MFAGatingDecorator) shouldEnforceValueThreshold(
	ctx sdk.Context,
	transactionType mfatypes.SensitiveTransactionType,
	msg sdk.Msg,
) bool {
	config, found := d.mfaKeeper.GetSensitiveTxConfig(ctx, transactionType)
	if !found || config == nil {
		config = defaultSensitiveTxConfig(transactionType)
	}
	if config == nil || config.ValueThreshold == "" {
		return true
	}

	threshold, ok := sdkmath.NewIntFromString(config.ValueThreshold)
	if !ok {
		return true
	}

	amount, ok := extractTransactionAmount(msg)
	if !ok {
		return true
	}

	return amount.GTE(threshold)
}

func (d MFAGatingDecorator) sensitiveTxConfig(ctx sdk.Context, txType mfatypes.SensitiveTransactionType) *mfatypes.SensitiveTxConfig {
	config, found := d.mfaKeeper.GetSensitiveTxConfig(ctx, txType)
	if found && config != nil {
		return config
	}
	return defaultSensitiveTxConfig(txType)
}

func extractTransactionAmount(msg sdk.Msg) (sdkmath.Int, bool) {
	switch m := msg.(type) {
	case *banktypes.MsgSend:
		return selectCoinAmount(m.Amount)
	case *dv1beta3.MsgCreateDeployment:
		return m.Deposit.Amount, true
	case *dv1beta4.MsgCreateDeployment:
		return depositAmount(m.Deposit)
	default:
		return sdkmath.Int{}, false
	}
}

func depositAmount(deposit depositv1.Deposit) (sdkmath.Int, bool) {
	return deposit.Amount.Amount, true
}

func selectCoinAmount(coins sdk.Coins) (sdkmath.Int, bool) {
	if len(coins) == 0 {
		return sdkmath.Int{}, false
	}

	for _, coin := range coins {
		if coin.Denom == "uve" {
			return coin.Amount, true
		}
	}

	coin := coins[0]
	return coin.Amount, true
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

type adaptedMFAProofProvider struct {
	proof             *mfatypes.MFAProof
	deviceFingerprint string
	trustToken        string
}

func newAdaptedMFAProofProvider(proof *mfatypes.MFAProof, deviceFingerprint, trustToken string) adaptedMFAProofProvider {
	if proof != nil {
		if proof.DeviceFingerprint == "" {
			proof.DeviceFingerprint = deviceFingerprint
		}
		if proof.TrustToken == "" {
			proof.TrustToken = trustToken
		}
	}
	return adaptedMFAProofProvider{
		proof:             proof,
		deviceFingerprint: deviceFingerprint,
		trustToken:        trustToken,
	}
}

func (a adaptedMFAProofProvider) GetMFAProof() *mfatypes.MFAProof {
	return a.proof
}

func (a adaptedMFAProofProvider) GetDeviceFingerprint() string {
	return a.deviceFingerprint
}

func (a adaptedMFAProofProvider) GetTrustToken() string {
	return a.trustToken
}

func decodeSerializedMFAProof(raw []byte) (*mfatypes.MFAProof, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var proof mfapb.MFAProof
	if err := proto.Unmarshal(raw, &proof); err != nil {
		return nil, err
	}
	return convertProtoMFAProof(&proof), nil
}

func convertProtoMFAProof(proof *mfapb.MFAProof) *mfatypes.MFAProof {
	if proof == nil {
		return nil
	}

	verifiedFactors := make([]mfatypes.FactorType, len(proof.VerifiedFactors))
	for i, factor := range proof.VerifiedFactors {
		switch factor {
		case mfapb.FactorTypeUnspecified:
			verifiedFactors[i] = mfatypes.FactorTypeUnspecified
		case mfapb.FactorTypeTOTP:
			verifiedFactors[i] = mfatypes.FactorTypeTOTP
		case mfapb.FactorTypeFIDO2:
			verifiedFactors[i] = mfatypes.FactorTypeFIDO2
		case mfapb.FactorTypeSMS:
			verifiedFactors[i] = mfatypes.FactorTypeSMS
		case mfapb.FactorTypeEmail:
			verifiedFactors[i] = mfatypes.FactorTypeEmail
		case mfapb.FactorTypeVEID:
			verifiedFactors[i] = mfatypes.FactorTypeVEID
		case mfapb.FactorTypeTrustedDevice:
			verifiedFactors[i] = mfatypes.FactorTypeTrustedDevice
		case mfapb.FactorTypeHardwareKey:
			verifiedFactors[i] = mfatypes.FactorTypeHardwareKey
		default:
			verifiedFactors[i] = mfatypes.FactorTypeUnspecified
		}
	}

	return &mfatypes.MFAProof{
		SessionID:         proof.SessionId,
		VerifiedFactors:   verifiedFactors,
		Timestamp:         proof.Timestamp,
		Signature:         proof.Signature,
		DeviceFingerprint: proof.DeviceFingerprint,
		TrustToken:        proof.TrustToken,
	}
}
