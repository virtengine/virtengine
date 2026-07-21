//go:build e2e.integration

package settlement_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/virtengine/virtengine/app"
	providerdaemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	providerv1 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/testutil"
	networktest "github.com/virtengine/virtengine/testutil/network"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const task84BGas = uint64(2_000_000)

type task84BCommittedClient struct {
	rpc networkRPCClient
}

type networkRPCClient interface {
	BroadcastTxCommit(context.Context, tmtypes.Tx) (*coretypes.ResultBroadcastTxCommit, error)
}

func (c task84BCommittedClient) EstimateGas(_ context.Context, _ []byte) (uint64, error) {
	return task84BGas, nil
}

func (c task84BCommittedClient) BroadcastTx(ctx context.Context, txBytes []byte) (string, error) {
	result, err := c.rpc.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return "", err
	}
	if result.CheckTx.Code != 0 {
		return "", errors.New(result.CheckTx.Log)
	}
	if result.TxResult.Code != 0 {
		return "", errors.New(result.TxResult.Log)
	}
	return result.Hash.String(), nil
}

type task84BSigningResolver struct {
	grpcAddress string
	rpc         networkRPCClient
	provider    string
	account     uint64
	sequence    uint64
}

func (r *task84BSigningResolver) ResolveProviderSigningState(ctx context.Context, provider string) (providerdaemon.ActiveProviderKeyBinding, error) {
	connection, err := grpc.NewClient(r.grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return providerdaemon.ActiveProviderKeyBinding{}, err
	}
	defer connection.Close()
	response, err := providerv1.NewQueryClient(connection).ProviderSigningKeyEpochs(ctx, &providerv1.QueryProviderSigningKeyEpochsRequest{Owner: provider})
	if err != nil {
		return providerdaemon.ActiveProviderKeyBinding{}, err
	}
	statusClient, ok := r.rpc.(interface {
		Status(context.Context) (*coretypes.ResultStatus, error)
	})
	if !ok {
		return providerdaemon.ActiveProviderKeyBinding{}, errors.New("status client unavailable")
	}
	status, err := statusClient.Status(ctx)
	if err != nil {
		return providerdaemon.ActiveProviderKeyBinding{}, err
	}
	if len(response.Keys) == 0 {
		return providerdaemon.ActiveProviderKeyBinding{}, errors.New("provider signing key unavailable")
	}
	key := response.Keys[len(response.Keys)-1]
	return providerdaemon.ActiveProviderKeyBinding{
		ProviderAddress: provider,
		KeyID:           key.KeyId,
		Epoch:           key.Epoch,
		PublicKey:       append([]byte(nil), key.PublicKey...),
		Algorithm:       key.KeyType,
		BlockHeight:     status.SyncInfo.LatestBlockHeight,
		BlockTime:       status.SyncInfo.LatestBlockTime,
	}, nil
}

func (r *task84BSigningResolver) ResolveUsageStreamState(ctx context.Context, provider, allocationID, orderID, leaseID string) (providerdaemon.OnChainUsageStreamState, error) {
	connection, err := grpc.NewClient(r.grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return providerdaemon.OnChainUsageStreamState{}, err
	}
	defer connection.Close()
	response, err := settlementv1.NewQueryClient(connection).UsageStreamState(ctx, &settlementv1.QueryUsageStreamStateRequest{
		Provider: provider, AllocationId: allocationID, OrderId: orderID, LeaseId: leaseID,
	})
	if err != nil {
		return providerdaemon.OnChainUsageStreamState{}, err
	}
	return providerdaemon.OnChainUsageStreamState{LastSequence: response.LastSequence}, nil
}

func (r *task84BSigningResolver) ResolveAccountSequence(context.Context, string) (uint64, uint64, error) {
	return r.account, r.sequence, nil
}

func TestTask84BFourValidatorProviderRestartCustomerAckExactlyOnce(t *testing.T) {
	if os.Getenv("VE_RUN_TASK84B_PROCESS") != "1" {
		t.Skip("set VE_RUN_TASK84B_PROCESS=1 to run the four-validator process-boundary test")
	}

	cfg := networktest.DefaultConfig(testutil.NewTestNetworkFixture)
	cfg.ChainID = "task-84b-process-boundary"
	cfg.NumValidators = 4
	cfg.AdditionalAccounts = 2
	cfg.TimeoutCommit = 100 * time.Millisecond
	net := networktest.New(t, cfg)
	t.Cleanup(net.Cleanup)
	require.NoError(t, net.WaitForNextBlock())

	validator := net.Validators[0]
	providerAccount := net.AdditionalAccounts[0]
	customerAccount := net.AdditionalAccounts[1]
	providerPriv, err := validator.ClientCtx.Keyring.ExportPrivateKeyObject(providerAccount.Name)
	require.NoError(t, err)
	providerCosmosKey, ok := providerPriv.(*cosmosed25519.PrivKey)
	if !ok {
		t.Skip("test network signer is not ed25519")
	}

	broadcastMsgs(t, cfg, validator, providerAccount, 0,
		&providerv1.MsgCreateProvider{Owner: providerAccount.Address.String(), HostURI: "https://provider.task84b.example"},
	)
	broadcastMsgs(t, cfg, validator, providerAccount, 1,
		&providerv1.MsgSetProviderSigningKey{Owner: providerAccount.Address.String(), PublicKey: providerCosmosKey.PubKey().Bytes(), KeyType: providerv1.PublicKeyTypeEd25519},
	)
	broadcastMsgs(t, cfg, validator, customerAccount, 0,
		&settlementv1.MsgCreateEscrow{Sender: customerAccount.Address.String(), OrderId: "task84b-order", Amount: sdk.NewCoins(sdk.NewInt64Coin("uve", 10_000)), ExpiresIn: 3_600},
	)

	connection, err := grpc.NewClient(validator.AppConfig.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	settlementQuery := settlementv1.NewQueryClient(connection)
	escrows, err := settlementQuery.EscrowsByOrder(context.Background(), &settlementv1.QueryEscrowsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, escrows.Escrows, 1)
	broadcastMsgs(t, cfg, validator, customerAccount, 1,
		&settlementv1.MsgActivateEscrow{Sender: customerAccount.Address.String(), EscrowId: escrows.Escrows[0].EscrowId, LeaseId: "task84b-lease", Recipient: providerAccount.Address.String()},
	)

	providerKM, err := providerdaemon.NewKeyManager(providerdaemon.KeyManagerConfig{StorageType: providerdaemon.KeyStorageTypeMemory, DefaultAlgorithm: providerv1.PublicKeyTypeEd25519})
	require.NoError(t, err)
	require.NoError(t, providerKM.Unlock(""))
	_, err = providerKM.ImportKey(providerAccount.Address.String(), providerCosmosKey.Bytes(), providerv1.PublicKeyTypeEd25519)
	require.NoError(t, err)
	providerNumber, providerSequence, err := cfg.AccountRetriever.GetAccountNumberSequence(validator.ClientCtx.WithClient(validator.RPCClient), providerAccount.Address)
	require.NoError(t, err)
	resolver := &task84BSigningResolver{
		grpcAddress: validator.AppConfig.GRPC.Address,
		rpc:         validator.RPCClient,
		provider:    providerAccount.Address.String(),
		account:     providerNumber,
		sequence:    providerSequence,
	}
	queuePath := filepath.Join(t.TempDir(), "provider-queue.json")
	submitterCfg := providerdaemon.DefaultChainSubmitterConfig()
	submitterCfg.ProviderAddress = providerAccount.Address.String()
	submitterCfg.ChainID = cfg.ChainID
	submitterCfg.ChainClient = task84BCommittedClient{rpc: validator.RPCClient}
	submitterCfg.QueueStatePath = queuePath
	submitterCfg.ProviderSigningState = resolver
	submitterCfg.UsageStreamState = resolver
	submitterCfg.AccountNumber = providerNumber
	submitterCfg.Sequence = providerSequence
	submitterCfg.RetryBackoff = 0

	report := &providerdaemon.ChainUsageReport{
		OrderID: "task84b-order", LeaseID: "task84b-lease", CustomerAddress: customerAccount.Address.String(), AllocationID: "task84b-allocation",
		UsageUnits: 1, UsageType: "cpu", PeriodStart: time.Now().UTC().Add(-time.Hour), PeriodEnd: time.Now().UTC(),
		UnitPrice: sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(100)), RawMetrics: providerdaemon.ResourceMetrics{CPUMilliSeconds: 3_600_000},
	}
	producer1, err := providerdaemon.NewChainUsageSubmitter(submitterCfg, providerKM, nil)
	require.NoError(t, err)
	require.NoError(t, producer1.SubmitUsageReport(context.Background(), report))
	producer1.Stop()
	resolver.sequence++

	producer2, err := providerdaemon.NewChainUsageSubmitter(submitterCfg, providerKM, nil)
	require.NoError(t, err)
	t.Cleanup(producer2.Stop)
	retry := *report
	require.ErrorIs(t, producer2.SubmitUsageReport(context.Background(), &retry), providerdaemon.ErrDuplicateReport)

	usageResponse, err := settlementQuery.UsageRecordsByOrder(context.Background(), &settlementv1.QueryUsageRecordsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, usageResponse.UsageRecords, 1)
	usage := usageResponse.UsageRecords[0]
	require.True(t, usage.SignatureVerified)

	ackReplay, err := settlementtypes.DeriveReplayKey("task84b-process-ack", usage.UsageId)
	require.NoError(t, err)
	status, err := validator.RPCClient.Status(context.Background())
	require.NoError(t, err)
	ackPayload := settlementtypes.CanonicalAcknowledgmentPayload{
		SignatureVersion: settlementtypes.SignatureVersionV1, ChainID: cfg.ChainID, Domain: settlementtypes.UsageCustomerDomainV1, SignerRole: settlementtypes.SignerRoleCustomer,
		Customer: customerAccount.Address.String(), UsageID: usage.UsageId, UsageDigest: usage.UsageDigest, ReplayKey: ackReplay,
		IssuedAtHeight: status.SyncInfo.LatestBlockHeight, ExpiresAtHeight: status.SyncInfo.LatestBlockHeight + 20,
		IssuedAtUnix: status.SyncInfo.LatestBlockTime.Unix(), ExpiresAtUnix: status.SyncInfo.LatestBlockTime.Add(10 * time.Minute).Unix(),
	}
	ackBytes, err := settlementtypes.CanonicalAcknowledgmentSignBytes(ackPayload)
	require.NoError(t, err)
	customerPriv, err := validator.ClientCtx.Keyring.ExportPrivateKeyObject(customerAccount.Name)
	require.NoError(t, err)
	ackSignature, err := customerPriv.Sign(ackBytes)
	require.NoError(t, err)
	broadcastMsgs(t, cfg, validator, customerAccount, 2, &settlementv1.MsgAcknowledgeUsage{
		Sender: customerAccount.Address.String(), UsageId: usage.UsageId, Signature: ackSignature, UsageDigest: usage.UsageDigest, ReplayKey: ackReplay,
		IssuedAtHeight: ackPayload.IssuedAtHeight, ExpiresAtHeight: ackPayload.ExpiresAtHeight, IssuedAtUnix: ackPayload.IssuedAtUnix, ExpiresAtUnix: ackPayload.ExpiresAtUnix,
		SignatureVersion: settlementtypes.SignatureVersionV1,
	})
	broadcastMsgs(t, cfg, validator, providerAccount, resolver.sequence, &settlementv1.MsgSettleOrder{Sender: providerAccount.Address.String(), OrderId: "task84b-order", UsageRecordIds: []string{usage.UsageId}})

	settlements, err := settlementQuery.SettlementsByOrder(context.Background(), &settlementv1.QuerySettlementsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, settlements.Settlements, 1)
	rewards, err := settlementQuery.RewardsByEpoch(context.Background(), &settlementv1.QueryRewardsByEpochRequest{EpochNumber: uint64(settlements.Settlements[0].BlockHeight / 100)})
	require.NoError(t, err)
	usageRewards := 0
	for _, reward := range rewards.Distributions {
		if reward.Source == "usage" {
			usageRewards++
		}
	}
	require.Equal(t, 1, usageRewards)
	for index := range net.Validators {
		appAtValidator, err := net.ValidatorApp(index)
		require.NoError(t, err)
		veApp := appAtValidator.(*app.VirtEngineApp)
		ctx := veApp.NewUncachedContext(false, cmtproto.Header{ChainID: cfg.ChainID, Height: settlements.Settlements[0].BlockHeight})
		stored := veApp.Keepers.VirtEngine.Settlement.GetUsageRecordsByOrder(ctx, "task84b-order")
		require.Len(t, stored, 1)
		require.True(t, stored[0].CustomerAcknowledged)
		require.True(t, stored[0].Settled)
	}

	t.Logf("TASK84B_RESULT validators=4 usage=%s acknowledgment=1 settlements=1 usage_rewards=1 queue=%s", usage.UsageId, queuePath)
}

func broadcastMsgs(t *testing.T, cfg networktest.Config, validator *networktest.Validator, account networktest.TestAccount, sequence uint64, messages ...sdk.Msg) {
	t.Helper()
	accountNumber, _, err := cfg.AccountRetriever.GetAccountNumberSequence(validator.ClientCtx.WithClient(validator.RPCClient), account.Address)
	require.NoError(t, err)
	privateKey, err := validator.ClientCtx.Keyring.ExportPrivateKeyObject(account.Name)
	require.NoError(t, err)
	builder := cfg.TxConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(messages...))
	builder.SetGasLimit(task84BGas)
	placeholder := signing.SignatureV2{PubKey: privateKey.PubKey(), Data: &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}, Sequence: sequence}
	require.NoError(t, builder.SetSignatures(placeholder))
	signature, err := clienttx.SignWithPrivKey(context.Background(), signing.SignMode_SIGN_MODE_DIRECT, authsigning.SignerData{
		ChainID: cfg.ChainID, AccountNumber: accountNumber, Sequence: sequence,
	}, builder, privateKey, cfg.TxConfig, sequence)
	require.NoError(t, err)
	require.NoError(t, builder.SetSignatures(signature))
	txBytes, err := cfg.TxConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)
	result, err := validator.RPCClient.BroadcastTxCommit(context.Background(), txBytes)
	require.NoError(t, err)
	require.Zero(t, result.CheckTx.Code, result.CheckTx.Log)
	require.Zero(t, result.TxResult.Code, result.TxResult.Log)
}
