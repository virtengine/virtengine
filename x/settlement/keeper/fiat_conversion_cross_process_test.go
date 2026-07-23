package keeper_test

import (
	"context"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"

	txsigning "cosmossdk.io/x/tx/signing"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/provider_daemon"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	hpcmodule "github.com/virtengine/virtengine/x/hpc"
	marketmodule "github.com/virtengine/virtengine/x/market"
	marketplacemodule "github.com/virtengine/virtengine/x/marketplace"
	providermodule "github.com/virtengine/virtengine/x/provider"
	resourcesmodule "github.com/virtengine/virtengine/x/resources"
	settlementmodule "github.com/virtengine/virtengine/x/settlement"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
	supportmodule "github.com/virtengine/virtengine/x/support"
)

type settlementMutationProcessBridge struct {
	chainID       string
	msgServer     settlementv1.MsgServer
	sdkCtx        sdk.Context
	accountNumber uint64
	sequence      uint64
	height        int64
	blockHash     string
	latestHeight  int64
	decodeTx      func([]byte) (sdk.Tx, error)
	signMode      *txsigning.HandlerMap
	confirmed     map[string]provider_daemon.ProviderTxConfirmation
	broadcasts    int
}

func (b *settlementMutationProcessBridge) ResolveAccountSequence(context.Context, string) (uint64, uint64, error) {
	return b.accountNumber, b.sequence, nil
}

func (*settlementMutationProcessBridge) EstimateGas(context.Context, []byte) (uint64, error) {
	return 100_000, nil
}

func (b *settlementMutationProcessBridge) BroadcastTx(ctx context.Context, txBytes []byte) (string, error) {
	tx, err := b.decodeTx(txBytes)
	if err != nil {
		return "", err
	}
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return "", types.ErrFiatObservationEvidence.Wrap("exactly one observation message required")
	}
	msg, ok := msgs[0].(*settlementv1.MsgRecordFiatConversionObservation)
	if !ok {
		return "", types.ErrFiatObservationEvidence.Wrap("unexpected mutation message type")
	}
	signatureTx, ok := tx.(interface {
		GetSignaturesV2() ([]signing.SignatureV2, error)
	})
	if !ok {
		return "", types.ErrUnauthorized.Wrap("decoded transaction has no SDK signatures")
	}
	signatures, err := signatureTx.GetSignaturesV2()
	if err != nil || len(signatures) != 1 || signatures[0].PubKey == nil {
		return "", types.ErrUnauthorized.Wrap("one provider SDK signature required")
	}
	single, ok := signatures[0].Data.(*signing.SingleSignatureData)
	if !ok || single.SignMode != signing.SignMode_SIGN_MODE_DIRECT || signatures[0].Sequence != b.sequence {
		return "", types.ErrUnauthorized.Wrap("unexpected provider SDK signature data")
	}
	signBytes, err := authsigning.GetSignBytesAdapter(ctx, b.signMode, single.SignMode, authsigning.SignerData{
		ChainID: b.chainID, AccountNumber: b.accountNumber, Sequence: b.sequence,
		Address: msg.Sender, PubKey: signatures[0].PubKey,
	}, tx)
	if err != nil || !signatures[0].PubKey.VerifySignature(signBytes, single.Signature) || sdk.AccAddress(signatures[0].PubKey.Address()).String() != msg.Sender {
		return "", types.ErrUnauthorized.Wrap("provider SDK signature verification failed")
	}
	if _, err := b.msgServer.RecordFiatConversionObservation(b.sdkCtx, msg); err != nil {
		return "", err
	}
	b.broadcasts++
	txHash := "TASK85B-MUTATION-" + time.Now().UTC().Format("150405.000000000")
	b.confirmed[txHash] = provider_daemon.ProviderTxConfirmation{Found: true, TxHash: txHash, Height: b.height, BlockHash: b.blockHash}
	b.sequence++
	return txHash, nil
}

func (b *settlementMutationProcessBridge) ConfirmTx(_ context.Context, txHash string) (provider_daemon.ProviderTxConfirmation, error) {
	return b.confirmed[txHash], nil
}

func (b *settlementMutationProcessBridge) LatestHeight(context.Context) (int64, error) {
	return b.latestHeight, nil
}

func (b *settlementMutationProcessBridge) BlockHash(context.Context, int64) (string, error) {
	return b.blockHash, nil
}

func (*settlementMutationProcessBridge) ReconcileMutation(context.Context, *provider_daemon.ProviderMutationEnvelope, sdk.Msg) (provider_daemon.ProviderMutationReconciliation, error) {
	return provider_daemon.ProviderMutationReconciliation{}, nil
}

func TestFiatObservationDurableMutationToAuthenticatedMsgServerProgression(t *testing.T) {
	keyConfig := provider_daemon.DefaultKeyManagerConfig()
	keyConfig.StorageType = provider_daemon.KeyStorageTypeMemory
	keyManager, err := provider_daemon.NewKeyManager(keyConfig)
	require.NoError(t, err)
	require.NoError(t, keyManager.Unlock(""))
	generated, err := keyManager.GenerateKey(sdk.AccAddress(make([]byte, 20)).String())
	require.NoError(t, err)
	publicKey, err := hex.DecodeString(generated.PublicKey)
	require.NoError(t, err)
	providerAddress := sdk.AccAddress((&cosmosed25519.PubKey{Key: publicKey}).Address()).String()
	generated.ProviderAddress = providerAddress

	provider, err := sdk.AccAddressFromBech32(providerAddress)
	require.NoError(t, err)
	s, conversion, _, compliance := setupAuthenticatedFiatConversionForProvider(t, false, provider)
	msgServer := settlementkeeper.NewMsgServerImpl(s.keeper)
	queuePath := filepath.Join(t.TempDir(), "task85b-mutations.json")

	config := provider_daemon.DefaultProviderMutationSubmitterConfig()
	config.ChainID = "virtengine-task85b-process"
	config.ProviderAddress = providerAddress
	config.QueueStatePath = queuePath
	config.PollInterval = time.Millisecond
	config.ConfirmationTimeout = time.Second
	config.FinalityBlocks = 0
	config.RetryBackoff = time.Millisecond
	config.MaxRetryBackoff = 2 * time.Millisecond
	encoding := sdkutil.MakeEncodingConfig(
		marketmodule.AppModuleBasic{}, hpcmodule.AppModuleBasic{}, resourcesmodule.AppModuleBasic{},
		providermodule.AppModuleBasic{}, settlementmodule.AppModuleBasic{}, marketplacemodule.AppModuleBasic{}, supportmodule.AppModuleBasic{},
	)
	bridge := &settlementMutationProcessBridge{
		chainID: config.ChainID, msgServer: msgServer, sdkCtx: s.ctx, accountNumber: 7, sequence: 11,
		height: 100, blockHash: "AABBCC", latestHeight: 102,
		decodeTx: encoding.TxConfig.TxDecoder(), signMode: encoding.TxConfig.SignModeHandler(),
		confirmed: make(map[string]provider_daemon.ProviderTxConfirmation),
	}

	config.Chain = bridge
	submitter, err := provider_daemon.NewProviderMutationSubmitter(config, keyManager)
	require.NoError(t, err)
	require.NoError(t, submitter.Start(context.Background()))

	quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	quote.QuoteDigest = bytesOfTest(31, 32)
	quote.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	result, err := submitter.SubmitFiatConversionObservation(context.Background(), quote)
	require.NoError(t, err)
	require.True(t, result.Final)
	require.Equal(t, 1, bridge.broadcasts)
	stored, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateSwapPending, stored.State)
	require.Equal(t, uint64(1), stored.ObservationSequence)

	require.NoError(t, submitter.Stop(context.Background()))
	restarted, err := provider_daemon.NewProviderMutationSubmitter(config, keyManager)
	require.NoError(t, err)
	require.NoError(t, restarted.Start(context.Background()))
	t.Cleanup(func() { require.NoError(t, restarted.Stop(context.Background())) })
	replay, err := restarted.SubmitFiatConversionObservation(context.Background(), quote)
	require.NoError(t, err)
	require.True(t, replay.Final)
	require.True(t, replay.Existed)
	require.Equal(t, 1, bridge.broadcasts, "durable terminal mutation must not rebroadcast after restart")
	stored, found = s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, uint64(1), stored.ObservationSequence)
}
