package security

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	gproto "google.golang.org/protobuf/proto"

	"github.com/virtengine/virtengine/app"
	mfapb "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	rolespb "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

type mfaTestVEIDKeeper struct{}

func (mfaTestVEIDKeeper) GetVEIDScore(_ sdk.Context, _ sdk.AccAddress) (uint32, bool) {
	return 90, true
}

type mfaTestRolesKeeper struct{}

func (mfaTestRolesKeeper) IsAccountOperational(_ sdk.Context, _ sdk.AccAddress) bool {
	return true
}

type mfaTestTx struct {
	msgs    []sdk.Msg
	signers []sdk.AccAddress
}

func (tx mfaTestTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (mfaTestTx) GetMsgsV2() ([]gproto.Message, error) {
	return nil, nil
}

func (tx mfaTestTx) GetSigners() ([][]byte, error) {
	signers := make([][]byte, 0, len(tx.signers))
	for _, signer := range tx.signers {
		signers = append(signers, signer.Bytes())
	}
	return signers, nil
}

func (mfaTestTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	return nil, nil
}

func (mfaTestTx) GetSignaturesV2() ([]txsigning.SignatureV2, error) {
	return nil, nil
}

type mfaTestEnv struct {
	ctx       sdk.Context
	keeper    mfakeeper.Keeper
	decorator app.MFAGatingDecorator
}

func newMFATestEnv(t *testing.T) mfaTestEnv {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	mfatypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(mfatypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Unix(1_700_000_000, 0).UTC(),
		Height: 100,
	}, false, log.NewNopLogger()).WithEventManager(sdk.NewEventManager())

	keeper := mfakeeper.NewKeeper(cdc, storeKey, "authority", mfaTestVEIDKeeper{}, mfaTestRolesKeeper{})
	require.NoError(t, keeper.SetParams(ctx, mfatypes.DefaultParams()))

	return mfaTestEnv{
		ctx:       ctx,
		keeper:    keeper,
		decorator: app.NewMFAGatingDecorator(keeper),
	}
}

func enrollActiveFactor(t *testing.T, keeper mfakeeper.Keeper, ctx sdk.Context, address sdk.AccAddress, factorType mfatypes.FactorType, factorID string) {
	t.Helper()

	enrollment := &mfatypes.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       factorType,
		FactorID:         factorID,
		PublicIdentifier: []byte("factor-key"),
		Status:           mfatypes.EnrollmentStatusActive,
		EnrolledAt:       ctx.BlockTime().Unix(),
	}
	if factorType == mfatypes.FactorTypeFIDO2 {
		enrollment.Metadata = &mfatypes.FactorMetadata{
			FIDO2Info: &mfatypes.FIDO2CredentialInfo{
				CredentialID: []byte(factorID),
				PublicKey:    []byte("public-key"),
			},
		}
	}

	require.NoError(t, keeper.EnrollFactor(ctx, enrollment))
}

func setMFAPolicy(t *testing.T, keeper mfakeeper.Keeper, ctx sdk.Context, address sdk.AccAddress, factors ...mfatypes.FactorType) {
	t.Helper()

	require.NoError(t, keeper.SetMFAPolicy(ctx, &mfatypes.MFAPolicy{
		AccountAddress: address.String(),
		Enabled:        true,
		RequiredFactors: []mfatypes.FactorCombination{
			{Factors: factors},
		},
		CreatedAt: ctx.BlockTime().Unix(),
		UpdatedAt: ctx.BlockTime().Unix(),
	}))
}

func setSensitiveTxConfig(
	t *testing.T,
	keeper mfakeeper.Keeper,
	ctx sdk.Context,
	txType mfatypes.SensitiveTransactionType,
	threshold string,
	factors ...mfatypes.FactorType,
) {
	t.Helper()

	require.NoError(t, keeper.SetSensitiveTxConfig(ctx, &mfatypes.SensitiveTxConfig{
		TransactionType: txType,
		Enabled:         true,
		RequiredFactorCombinations: []mfatypes.FactorCombination{
			{Factors: factors},
		},
		ValueThreshold: threshold,
		Description:    txType.String() + " requires MFA",
	}))
}

func createSession(
	t *testing.T,
	keeper mfakeeper.Keeper,
	ctx sdk.Context,
	address sdk.AccAddress,
	txType mfatypes.SensitiveTransactionType,
	sessionID string,
	factors ...mfatypes.FactorType,
) {
	t.Helper()

	require.NoError(t, keeper.CreateAuthorizationSession(ctx, &mfatypes.AuthorizationSession{
		SessionID:       sessionID,
		AccountAddress:  address.String(),
		TransactionType: txType,
		CreatedAt:       ctx.BlockTime().Unix(),
		ExpiresAt:       ctx.BlockTime().Add(time.Hour).Unix(),
		VerifiedFactors: factors,
	}))
}

func runMFADecorator(t *testing.T, decorator app.MFAGatingDecorator, ctx sdk.Context, signers []sdk.AccAddress, msg sdk.Msg) error {
	t.Helper()

	nextCalled := false
	_, err := decorator.AnteHandle(ctx, mfaTestTx{msgs: []sdk.Msg{msg}, signers: signers}, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	})
	if err == nil {
		require.True(t, nextCalled, "ante handler should reach next decorator on success")
	}
	return err
}

func TestMFAGatingDecorator_BankSendThresholds(t *testing.T) {
	env := newMFATestEnv(t)

	sender := sdk.AccAddress([]byte("security-sender-addr"))
	recipient := sdk.AccAddress([]byte("security-recipient-"))

	enrollActiveFactor(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP, "totp-send")
	setMFAPolicy(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP)
	setSensitiveTxConfig(t, env.keeper, env.ctx, mfatypes.SensitiveTxMediumWithdrawal, "1000", mfatypes.FactorTypeTOTP)
	setSensitiveTxConfig(t, env.keeper, env.ctx, mfatypes.SensitiveTxLargeWithdrawal, "10000", mfatypes.FactorTypeTOTP)

	lowValue := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("uve", 999)),
	}
	require.NoError(t, runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, lowValue))

	mediumValue := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
	}
	err := runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, mediumValue)
	require.Error(t, err)
	require.ErrorContains(t, err, mfatypes.SensitiveTxMediumWithdrawal.String())

	largeValue := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("uve", 10000)),
	}
	err = runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, largeValue)
	require.Error(t, err)
	require.ErrorContains(t, err, mfatypes.SensitiveTxLargeWithdrawal.String())
}

func TestMFAGatingDecorator_AccountRecoveryRequiresValidProof(t *testing.T) {
	env := newMFATestEnv(t)

	sender := sdk.AccAddress([]byte("security-accountsend"))
	target := sdk.AccAddress([]byte("security-accounttgt"))

	enrollActiveFactor(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP, "totp-recovery")
	setMFAPolicy(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP)
	setSensitiveTxConfig(t, env.keeper, env.ctx, mfatypes.SensitiveTxAccountRecovery, "", mfatypes.FactorTypeTOTP)

	msg := &rolespb.MsgSetAccountState{
		Sender:  sender.String(),
		Address: target.String(),
		State:   "suspended",
		Reason:  "security test without proof",
	}

	err := runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg)
	require.Error(t, err)
	require.ErrorContains(t, err, mfatypes.SensitiveTxAccountRecovery.String())

	createSession(t, env.keeper, env.ctx, sender, mfatypes.SensitiveTxAccountRecovery, "acct-recovery-session", mfatypes.FactorTypeTOTP)

	msg.MfaProof = &mfapb.MFAProof{
		SessionId:       "acct-recovery-session",
		VerifiedFactors: []mfapb.FactorType{mfapb.FactorTypeTOTP},
		Timestamp:       env.ctx.BlockTime().Unix(),
	}
	msg.DeviceFingerprint = "security-account-device"

	require.NoError(t, runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg))
}

func TestMFAGatingDecorator_RebindWalletAcceptsSerializedProof(t *testing.T) {
	env := newMFATestEnv(t)

	sender := sdk.AccAddress([]byte("security-walletsender"))

	enrollActiveFactor(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP, "totp-rebind")
	setMFAPolicy(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP)
	setSensitiveTxConfig(t, env.keeper, env.ctx, mfatypes.SensitiveTxKeyRotation, "", mfatypes.FactorTypeTOTP)

	msg := &veidv1.MsgRebindWallet{
		Sender:              sender.String(),
		NewBindingSignature: []byte("new-binding-signature"),
		NewBindingPubKey:    []byte("new-binding-pubkey"),
		OldSignature:        []byte("old-wallet-signature"),
		DeviceFingerprint:   "security-wallet-device",
	}

	err := runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg)
	require.Error(t, err)
	require.ErrorContains(t, err, mfatypes.SensitiveTxKeyRotation.String())

	createSession(t, env.keeper, env.ctx, sender, mfatypes.SensitiveTxKeyRotation, "wallet-rebind-session", mfatypes.FactorTypeTOTP)

	rawProof, marshalErr := proto.Marshal(&mfapb.MFAProof{
		SessionId:       "wallet-rebind-session",
		VerifiedFactors: []mfapb.FactorType{mfapb.FactorTypeTOTP},
		Timestamp:       env.ctx.BlockTime().Unix(),
	})
	require.NoError(t, marshalErr)

	msg.MfaProof = rawProof

	require.NoError(t, runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg))
}
