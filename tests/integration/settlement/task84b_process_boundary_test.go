//go:build e2e.integration

package settlement_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cosmoscrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/virtengine/virtengine/app"
	providerdaemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	providerv1 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/testutil"
	networktest "github.com/virtengine/virtengine/testutil/network"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const (
	task84BGas          = uint64(2_000_000)
	task84BOrderID      = "task84b-order"
	task84BLeaseID      = "task84b-lease"
	task84BAllocationID = "task84b-allocation"
	task84BHelperEnv    = "VE_TASK84B_HELPER"
	task84BChainRetry   = "provider-chain-retry"
)

type task84BProviderHelperInput struct {
	RPCAddress     string                          `json:"rpc_address"`
	GRPCAddress    string                          `json:"grpc_address"`
	ChainID        string                          `json:"chain_id"`
	Provider       string                          `json:"provider"`
	ProviderKeyHex string                          `json:"provider_key_hex"`
	QueuePath      string                          `json:"queue_path"`
	Report         providerdaemon.ChainUsageReport `json:"report"`
}

type task84BProviderHelperOutput struct {
	Report             providerdaemon.ChainUsageReport `json:"report"`
	LocalDuplicate     bool                            `json:"local_duplicate"`
	ChainRetryAccepted bool                            `json:"chain_retry_accepted"`
	TxHeight           int64                           `json:"tx_height,omitempty"`
}

type task84BCustomerHelperInput struct {
	RPCAddress     string `json:"rpc_address"`
	GRPCAddress    string `json:"grpc_address"`
	ChainID        string `json:"chain_id"`
	Customer       string `json:"customer"`
	CustomerKeyHex string `json:"customer_key_hex"`
	UsageID        string `json:"usage_id"`
	UsageDigest    []byte `json:"usage_digest"`
	ReplayKey      []byte `json:"replay_key"`
	IssuedHeight   int64  `json:"issued_height"`
	ExpiresHeight  int64  `json:"expires_height"`
	IssuedUnix     int64  `json:"issued_unix"`
	ExpiresUnix    int64  `json:"expires_unix"`
}

type task84BCustomerHelperOutput struct {
	TxHash    string `json:"tx_hash"`
	Signature []byte `json:"signature"`
}

type networkRPCClient interface {
	Status(context.Context) (*coretypes.ResultStatus, error)
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
	status, err := r.rpc.Status(ctx)
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

func TestTask84BProcessHelper(t *testing.T) {
	mode := os.Getenv(task84BHelperEnv)
	if mode == "" {
		t.Skip("Task 84B subprocess helper")
	}
	inputPath := os.Getenv("VE_TASK84B_HELPER_INPUT")
	outputPath := os.Getenv("VE_TASK84B_HELPER_OUTPUT")
	require.NotEmpty(t, inputPath)
	require.NotEmpty(t, outputPath)

	switch mode {
	case "provider-submit", "provider-restart", task84BChainRetry:
		var input task84BProviderHelperInput
		readTask84BJSON(t, inputPath, &input)
		privateKeyBytes, err := hex.DecodeString(input.ProviderKeyHex)
		require.NoError(t, err)
		keyManager, err := providerdaemon.NewKeyManager(providerdaemon.KeyManagerConfig{
			StorageType: providerdaemon.KeyStorageTypeMemory, DefaultAlgorithm: providerv1.PublicKeyTypeEd25519,
		})
		require.NoError(t, err)
		require.NoError(t, keyManager.Unlock(""))
		_, err = keyManager.ImportKey(input.Provider, privateKeyBytes, providerv1.PublicKeyTypeEd25519)
		require.NoError(t, err)
		connection, err := grpc.NewClient(input.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer connection.Close()
		rpc, err := rpchttp.New(input.RPCAddress, "/websocket")
		require.NoError(t, err)
		require.NoError(t, rpc.Start())
		account, err := authtypes.NewQueryClient(connection).AccountInfo(context.Background(), &authtypes.QueryAccountInfoRequest{Address: input.Provider})
		require.NoError(t, err)
		require.NotNil(t, account.Info)
		resolver := &task84BSigningResolver{
			grpcAddress: input.GRPCAddress,
			rpc:         rpc,
			provider:    input.Provider,
			account:     account.Info.AccountNumber,
			sequence:    account.Info.Sequence,
		}
		config := providerdaemon.DefaultChainSubmitterConfig()
		config.ProviderAddress = input.Provider
		config.ChainID = input.ChainID
		config.RPCClient = rpc
		config.CometRPC = input.RPCAddress
		config.QueueStatePath = input.QueuePath
		if mode == task84BChainRetry {
			config.QueueStatePath += ".fresh"
			prepareTask84BChainRetryQueue(t, input.QueuePath, config.QueueStatePath)
		}
		config.ProviderSigningState = resolver
		config.UsageStreamState = resolver
		config.AccountNumber = account.Info.AccountNumber
		config.Sequence = account.Info.Sequence
		config.RetryBackoff = 0
		submitter, err := providerdaemon.NewChainUsageSubmitter(config, keyManager, nil)
		require.NoError(t, err)
		require.NoError(t, submitter.Start(context.Background()))
		defer submitter.Stop()
		report := input.Report
		err = submitter.SubmitUsageReport(context.Background(), &report)
		localDuplicate := errors.Is(err, providerdaemon.ErrDuplicateReport)
		if mode == "provider-restart" {
			require.True(t, localDuplicate, "durable queue restart must suppress the committed report")
		} else {
			require.NoError(t, err)
		}
		var txHeight int64
		if mode == task84BChainRetry {
			status, statusErr := rpc.Status(context.Background())
			require.NoError(t, statusErr)
			txHeight = status.SyncInfo.LatestBlockHeight
		}
		writeTask84BJSON(t, outputPath, task84BProviderHelperOutput{
			Report: report, LocalDuplicate: localDuplicate,
			ChainRetryAccepted: mode == task84BChainRetry && err == nil,
			TxHeight:           txHeight,
		})
	case "customer-ack":
		var input task84BCustomerHelperInput
		readTask84BJSON(t, inputPath, &input)
		privateKeyBytes, err := hex.DecodeString(input.CustomerKeyHex)
		require.NoError(t, err)
		privateKey := &cosmossecp256k1.PrivKey{Key: privateKeyBytes}
		payload := settlementtypes.CanonicalAcknowledgmentPayload{
			SignatureVersion: settlementtypes.SignatureVersionV1,
			ChainID:          input.ChainID,
			Domain:           settlementtypes.UsageCustomerDomainV1,
			SignerRole:       settlementtypes.SignerRoleCustomer,
			Customer:         input.Customer,
			UsageID:          input.UsageID,
			UsageDigest:      input.UsageDigest,
			ReplayKey:        input.ReplayKey,
			IssuedAtHeight:   input.IssuedHeight,
			ExpiresAtHeight:  input.ExpiresHeight,
			IssuedAtUnix:     input.IssuedUnix,
			ExpiresAtUnix:    input.ExpiresUnix,
		}
		signBytes, err := settlementtypes.CanonicalAcknowledgmentSignBytes(payload)
		require.NoError(t, err)
		signature, err := privateKey.Sign(signBytes)
		require.NoError(t, err)
		connection, err := grpc.NewClient(input.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer connection.Close()
		account, err := authtypes.NewQueryClient(connection).AccountInfo(context.Background(), &authtypes.QueryAccountInfoRequest{Address: input.Customer})
		require.NoError(t, err)
		require.NotNil(t, account.Info)
		message := &settlementv1.MsgAcknowledgeUsage{
			Sender: input.Customer, UsageId: input.UsageID, Signature: signature,
			UsageDigest: input.UsageDigest, ReplayKey: input.ReplayKey,
			IssuedAtHeight: input.IssuedHeight, ExpiresAtHeight: input.ExpiresHeight,
			IssuedAtUnix: input.IssuedUnix, ExpiresAtUnix: input.ExpiresUnix,
			SignatureVersion: settlementtypes.SignatureVersionV1,
		}
		txBytes, err := buildTask84BSignedTx(input.ChainID, account.Info.AccountNumber, account.Info.Sequence, privateKey, message)
		require.NoError(t, err)
		rpc, err := rpchttp.New(input.RPCAddress, "/websocket")
		require.NoError(t, err)
		require.NoError(t, rpc.Start())
		defer func() { require.NoError(t, rpc.Stop()) }()
		result, err := rpc.BroadcastTxCommit(context.Background(), txBytes)
		require.NoError(t, err)
		require.Zero(t, result.CheckTx.Code, result.CheckTx.Log)
		require.Zero(t, result.TxResult.Code, result.TxResult.Log)
		writeTask84BJSON(t, outputPath, task84BCustomerHelperOutput{TxHash: result.Hash.String(), Signature: signature})
	default:
		t.Fatalf("unknown Task 84B helper mode %q", mode)
	}
}

func TestTask84BFourValidatorProviderRestartCustomerAckExactlyOnce(t *testing.T) {
	if os.Getenv("VE_RUN_TASK84B_PROCESS") != "1" {
		t.Skip("set VE_RUN_TASK84B_PROCESS=1 to run the four-validator process-boundary test")
	}

	cfg := networktest.DefaultConfig(testutil.NewTestNetworkFixture)
	cfg.ChainID = "task-84b-process-boundary"
	cfg.NumValidators = 4
	cfg.AdditionalAccounts = 1
	cfg.TimeoutCommit = 100 * time.Millisecond
	// Windows CometBFT peer routines can outlive node shutdown briefly and read
	// the block store after directory removal. Retain the isolated temp network
	// on Windows to avoid turning a passing protocol test into a teardown panic.
	cfg.CleanupDir = false
	providerCosmosKey := cosmosed25519.GenPrivKey()
	providerAddress := sdk.AccAddress(providerCosmosKey.PubKey().Address())
	var providerSequence uint64
	seedProviderGenesis(t, &cfg, providerAddress)
	net := networktest.New(t, cfg)
	t.Cleanup(net.Cleanup)
	require.NoError(t, net.WaitForNextBlock())
	assertTask84BAuthenticationActive(t, net)

	validator := net.Validators[0]
	customerAccount := net.AdditionalAccounts[0]
	broadcastWithPrivateKey(t, cfg, validator, providerAddress, providerCosmosKey, 0,
		&providerv1.MsgSetProviderSigningKey{Owner: providerAddress.String(), PublicKey: providerCosmosKey.PubKey().Bytes(), KeyType: providerv1.PublicKeyTypeEd25519},
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
		&settlementv1.MsgActivateEscrow{Sender: customerAccount.Address.String(), EscrowId: escrows.Escrows[0].EscrowId, LeaseId: "task84b-lease", Recipient: providerAddress.String()},
	)

	status, err := validator.RPCClient.Status(context.Background())
	require.NoError(t, err)
	periodEnd := status.SyncInfo.LatestBlockTime.UTC()
	periodStart := periodEnd.Add(-time.Hour)
	providerKeyID := providerv1.ComputeProviderKeyID(providerv1.PublicKeyTypeEd25519, providerCosmosKey.PubKey().Bytes())
	arbitrary := task84BUsageMessage(cfg.ChainID, providerAddress.String(), providerKeyID, periodStart, periodEnd, status.SyncInfo.LatestBlockHeight)
	arbitrary.Signature = bytes.Repeat([]byte{0xA5}, 64)
	arbitraryResult := broadcastWithPrivateKeyExpectDeliverFailure(t, cfg, validator, providerAddress, providerCosmosKey, 1, arbitrary)
	require.Contains(t, strings.ToLower(arbitraryResult.TxResult.Log), "signature")
	require.Zero(t, countTask84BEvents(arbitraryResult.TxResult.Events, settlementtypes.EventTypeUsageRecorded))
	assertTask84BUsageCount(t, settlementQuery, 0)

	queuePath := filepath.Join(t.TempDir(), "provider-queue.json")
	providerInput := task84BProviderHelperInput{
		RPCAddress:     strings.Replace(validator.RPCAddress, "tcp://", "http://", 1),
		GRPCAddress:    validator.AppConfig.GRPC.Address,
		ChainID:        cfg.ChainID,
		Provider:       providerAddress.String(),
		ProviderKeyHex: hex.EncodeToString(providerCosmosKey.Bytes()),
		QueuePath:      queuePath,
		Report: providerdaemon.ChainUsageReport{
			OrderID: "task84b-order", LeaseID: "task84b-lease", CustomerAddress: customerAccount.Address.String(), AllocationID: "task84b-allocation",
			UsageUnits: 1, UsageType: "cpu", PeriodStart: periodStart, PeriodEnd: periodEnd,
			UnitPrice: sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(100)), RawMetrics: providerdaemon.ResourceMetrics{CPUMilliSeconds: 3_600_000},
		},
	}
	providerFirst := runTask84BSubprocess[task84BProviderHelperOutput](t, "provider-submit", providerInput)
	report := &providerFirst.Report
	require.NotEmpty(t, report.Signature)
	require.Equal(t, uint64(1), report.StreamSequence)
	queueStateBeforeRestart, err := os.ReadFile(queuePath)
	require.NoError(t, err)
	require.Contains(t, string(queueStateBeforeRestart), "broadcasted")
	providerRestart := runTask84BSubprocess[task84BProviderHelperOutput](t, "provider-restart", providerInput)
	require.True(t, providerRestart.LocalDuplicate)

	usageResponse, err := settlementQuery.UsageRecordsByOrder(context.Background(), &settlementv1.QueryUsageRecordsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, usageResponse.UsageRecords, 1)
	usage := usageResponse.UsageRecords[0]
	require.True(t, usage.SignatureVerified)
	require.Equal(t, settlementtypes.UsageAuthenticationStatusVerified, usage.AuthenticationStatus)
	require.Equal(t, report.StreamSequence, usage.StreamSequence)
	require.Equal(t, report.ProviderKeyID, usage.ProviderKeyId)
	require.True(t, usage.SignatureMaterialRedacted)
	require.Empty(t, usage.Nonce)
	require.Empty(t, usage.IdempotencyKey)
	require.Empty(t, usage.ProviderSignature)
	require.Len(t, usage.UsageDigest, settlementtypes.DigestSize)
	usageBlockResults := waitForTask84BBlockResults(t, validator, usage.BlockHeight)
	require.Equal(t, 1, countTask84BEventsInBlock(usageBlockResults, settlementtypes.EventTypeUsageRecorded))

	// The producer API suppresses an identical report from the durable queue and
	// returns ErrDuplicateReport. A third provider process with a fresh local
	// queue submits the identical proof, proving the chain replay index returns
	// idempotent success without a duplicate event or state transition.
	chainRetry := runTask84BSubprocess[task84BProviderHelperOutput](t, task84BChainRetry, providerInput)
	require.True(t, chainRetry.ChainRetryAccepted)
	require.Equal(t, report.Signature, chainRetry.Report.Signature)
	require.Equal(t, report.IdempotencyKey, chainRetry.Report.IdempotencyKey)
	require.Zero(t, countTask84BEventsInBlock(waitForTask84BBlockResults(t, validator, chainRetry.TxHeight), settlementtypes.EventTypeUsageRecorded))
	assertTask84BUsageCount(t, settlementQuery, 1)

	// A different signed payload under the accepted replay key must fail. The
	// changed raw CPU delta preserves the one-unit pricing formula, ensuring the
	// rejection comes from replay conflict rather than malformed accounting.
	conflict := task84BUsageMessageFromReport(providerAddress.String(), report)
	conflict.RawMetrics.CpuMilliSeconds--
	conflict.Signature = signTask84BUsage(t, customerAccount.Address.String(), conflict, providerCosmosKey)
	providerSequence = task84BAccountSequence(t, cfg, validator, providerAddress)
	conflictResult := broadcastWithPrivateKeyExpectDeliverFailure(t, cfg, validator, providerAddress, providerCosmosKey, providerSequence, conflict)
	require.Contains(t, strings.ToLower(conflictResult.TxResult.Log), "replay")
	require.Zero(t, countTask84BEvents(conflictResult.TxResult.Events, settlementtypes.EventTypeUsageRecorded))
	assertTask84BUsageCount(t, settlementQuery, 1)

	ackReplay, err := settlementtypes.DeriveReplayKey("task84b-process-ack", usage.UsageId)
	require.NoError(t, err)
	status, err = validator.RPCClient.Status(context.Background())
	require.NoError(t, err)
	ackPayload := settlementtypes.CanonicalAcknowledgmentPayload{
		SignatureVersion: settlementtypes.SignatureVersionV1, ChainID: cfg.ChainID, Domain: settlementtypes.UsageCustomerDomainV1, SignerRole: settlementtypes.SignerRoleCustomer,
		Customer: customerAccount.Address.String(), UsageID: usage.UsageId, UsageDigest: usage.UsageDigest, ReplayKey: ackReplay,
		IssuedAtHeight: status.SyncInfo.LatestBlockHeight, ExpiresAtHeight: status.SyncInfo.LatestBlockHeight + 20,
		IssuedAtUnix: status.SyncInfo.LatestBlockTime.Unix(), ExpiresAtUnix: status.SyncInfo.LatestBlockTime.Add(10 * time.Minute).Unix(),
	}
	customerPriv := exportTestPrivateKey(t, validator.ClientCtx.Keyring, customerAccount.Name)
	customerResult := runTask84BSubprocess[task84BCustomerHelperOutput](t, "customer-ack", task84BCustomerHelperInput{
		RPCAddress:     strings.Replace(validator.RPCAddress, "tcp://", "http://", 1),
		GRPCAddress:    validator.AppConfig.GRPC.Address,
		ChainID:        cfg.ChainID,
		Customer:       customerAccount.Address.String(),
		CustomerKeyHex: hex.EncodeToString(customerPriv.Bytes()),
		UsageID:        usage.UsageId,
		UsageDigest:    append([]byte(nil), usage.UsageDigest...),
		ReplayKey:      ackReplay,
		IssuedHeight:   ackPayload.IssuedAtHeight,
		ExpiresHeight:  ackPayload.ExpiresAtHeight,
		IssuedUnix:     ackPayload.IssuedAtUnix,
		ExpiresUnix:    ackPayload.ExpiresAtUnix,
	})
	require.NotEmpty(t, customerResult.TxHash)
	require.NotEmpty(t, customerResult.Signature)
	ackBytes, err := settlementtypes.CanonicalAcknowledgmentSignBytes(ackPayload)
	require.NoError(t, err)
	require.True(t, customerPriv.PubKey().VerifySignature(ackBytes, customerResult.Signature))
	providerSequence = task84BAccountSequence(t, cfg, validator, providerAddress)
	broadcastWithPrivateKey(t, cfg, validator, providerAddress, providerCosmosKey, providerSequence, &settlementv1.MsgSettleOrder{Sender: providerAddress.String(), OrderId: "task84b-order", UsageRecordIds: []string{usage.UsageId}})

	settlements, err := settlementQuery.SettlementsByOrder(context.Background(), &settlementv1.QuerySettlementsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, settlements.Settlements, 1)
	require.Equal(t, "100uve", settlements.Settlements[0].TotalAmount.String())
	require.GreaterOrEqual(t, settlements.Settlements[0].BlockHeight, int64(0))
	rewards, err := settlementQuery.RewardsByEpoch(context.Background(), &settlementv1.QueryRewardsByEpochRequest{EpochNumber: uint64(settlements.Settlements[0].BlockHeight / 100)}) //nolint:gosec // non-negative height asserted above
	require.NoError(t, err)
	usageRewards := 0
	for _, reward := range rewards.Distributions {
		if reward.Source == "usage" {
			usageRewards++
		}
	}
	require.Equal(t, 1, usageRewards)
	settlementHeight := settlements.Settlements[0].BlockHeight
	settlementResults := waitForTask84BBlockResults(t, validator, settlementHeight)
	require.Equal(t, 1, countTask84BEventsInBlock(settlementResults, settlementtypes.EventTypeOrderSettled))
	require.Equal(t, 1, countTask84BEventsInBlock(settlementResults, settlementtypes.EventTypeRewardsDistributed))
	require.Equal(t, 1, totalTask84BUsageEvents(t, validator, usage.BlockHeight, settlementHeight))

	providerBalance := task84BBalance(t, connection, providerAddress.String(), "uve")
	escrowBalance := task84BEscrowBalance(t, settlementQuery)
	providerSequence = task84BAccountSequence(t, cfg, validator, providerAddress)
	duplicateSettlement := broadcastWithPrivateKeyExpectDeliverFailure(t, cfg, validator, providerAddress, providerCosmosKey, providerSequence,
		&settlementv1.MsgSettleOrder{Sender: providerAddress.String(), OrderId: "task84b-order", UsageRecordIds: []string{usage.UsageId}},
	)
	require.Contains(t, strings.ToLower(duplicateSettlement.TxResult.Log), "settled")
	settlements, err = settlementQuery.SettlementsByOrder(context.Background(), &settlementv1.QuerySettlementsByOrderRequest{OrderId: "task84b-order"})
	require.NoError(t, err)
	require.Len(t, settlements.Settlements, 1)
	require.Equal(t, providerBalance, task84BBalance(t, connection, providerAddress.String(), "uve"))
	require.Equal(t, escrowBalance, task84BEscrowBalance(t, settlementQuery))
	require.Equal(t, 1, totalTask84BUsageEvents(t, validator, usage.BlockHeight, latestTask84BHeight(t, validator)))

	finalHeight := latestTask84BHeight(t, validator)
	heights, appHashes := task84BValidatorAppHashes(t, net, finalHeight)
	for index := 1; index < len(appHashes); index++ {
		require.Equal(t, appHashes[0], appHashes[index], "validator %d app hash differs at height %d", index, finalHeight)
	}
	t.Logf("TASK84B_RESULT validators=4 activation=fresh_genesis height=%d app_hash=%X validator_heights=%v arbitrary_signature=rejected queue_restart=verified local_duplicate=ErrDuplicateReport chain_exact_retry=success conflict_replay=rejected usage=%s usage_events=1 acknowledgment=1 settlements=1 usage_rewards=1 provider_balance=%s escrow_balance=%s queue=%s",
		finalHeight, appHashes[0], heights, usage.UsageId, providerBalance, escrowBalance, queuePath)
}

func seedProviderGenesis(t *testing.T, cfg *networktest.Config, provider sdk.AccAddress) {
	t.Helper()
	var authGenesis authtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGenesis)
	accounts, err := authtypes.PackAccounts([]authtypes.GenesisAccount{authtypes.NewBaseAccount(provider, nil, 0, 0)})
	require.NoError(t, err)
	authGenesis.Accounts = append(authGenesis.Accounts, accounts...)
	cfg.GenesisState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authGenesis)

	var bankGenesis banktypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[banktypes.ModuleName], &bankGenesis)
	bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
		Address: provider.String(), Coins: sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, cfg.AccountTokens)),
	})
	cfg.GenesisState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankGenesis)

	var providerGenesis providerv1.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[providerv1.ModuleName], &providerGenesis)
	providerGenesis.Providers = append(providerGenesis.Providers, providerv1.Provider{
		Owner: provider.String(), HostURI: "https://provider.task84b.example",
	})
	cfg.GenesisState[providerv1.ModuleName] = cfg.Codec.MustMarshalJSON(&providerGenesis)
}

func task84BUsageMessage(chainID, provider, providerKeyID string, periodStart, periodEnd time.Time, issuedHeight int64) *settlementv1.MsgRecordUsage {
	nonce, _ := settlementtypes.DeriveReplayKey("task84b-negative-nonce", provider, task84BOrderID)
	idempotencyKey, _ := settlementtypes.DeriveReplayKey("task84b-negative-idempotency", provider, task84BOrderID)
	return &settlementv1.MsgRecordUsage{
		Sender:           provider,
		OrderId:          task84BOrderID,
		LeaseId:          task84BLeaseID,
		UsageUnits:       1,
		UsageType:        "cpu",
		PeriodStart:      periodStart.Unix(),
		PeriodEnd:        periodEnd.Unix(),
		UnitPrice:        sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(100)),
		AllocationId:     task84BAllocationID,
		ChainId:          chainID,
		RawMetrics:       &settlementv1.RawUsageMetrics{CpuMilliSeconds: 3_600_000},
		PricingVersion:   1,
		FormulaVersion:   1,
		ModelVersion:     1,
		StreamSequence:   1,
		Nonce:            nonce,
		IdempotencyKey:   idempotencyKey,
		ProviderKeyEpoch: 1,
		ProviderKeyId:    providerKeyID,
		IssuedAtHeight:   issuedHeight,
		ExpiresAtHeight:  issuedHeight + 20,
		IssuedAtUnix:     periodEnd.Unix(),
		ExpiresAtUnix:    periodEnd.Add(10 * time.Minute).Unix(),
		SignatureVersion: settlementtypes.SignatureVersionV1,
	}
}

func task84BUsageMessageFromReport(provider string, report *providerdaemon.ChainUsageReport) *settlementv1.MsgRecordUsage {
	return &settlementv1.MsgRecordUsage{
		Sender:       provider,
		OrderId:      report.OrderID,
		LeaseId:      report.LeaseID,
		UsageUnits:   report.UsageUnits,
		UsageType:    report.UsageType,
		PeriodStart:  report.PeriodStart.Unix(),
		PeriodEnd:    report.PeriodEnd.Unix(),
		UnitPrice:    report.UnitPrice,
		Signature:    append([]byte(nil), report.Signature...),
		AllocationId: report.AllocationID,
		ChainId:      report.ChainID,
		RawMetrics: &settlementv1.RawUsageMetrics{
			CpuMilliSeconds:    report.RawMetrics.CPUMilliSeconds,
			MemoryByteSeconds:  report.RawMetrics.MemoryByteSeconds,
			StorageByteSeconds: report.RawMetrics.StorageByteSeconds,
			NetworkBytesIn:     report.RawMetrics.NetworkBytesIn,
			NetworkBytesOut:    report.RawMetrics.NetworkBytesOut,
			GpuSeconds:         report.RawMetrics.GPUSeconds,
		},
		PricingVersion:   report.PricingVersion,
		FormulaVersion:   report.FormulaVersion,
		ModelVersion:     report.ModelVersion,
		StreamSequence:   report.StreamSequence,
		Nonce:            append([]byte(nil), report.Nonce...),
		IdempotencyKey:   append([]byte(nil), report.IdempotencyKey...),
		ProviderKeyEpoch: report.ProviderKeyEpoch,
		ProviderKeyId:    report.ProviderKeyID,
		IssuedAtHeight:   report.IssuedAtHeight,
		ExpiresAtHeight:  report.ExpiresAtHeight,
		IssuedAtUnix:     report.IssuedAtUnix,
		ExpiresAtUnix:    report.ExpiresAtUnix,
		SignatureVersion: report.SignatureVersion,
	}
}

func signTask84BUsage(t *testing.T, customer string, message *settlementv1.MsgRecordUsage, privateKey cryptotypes.PrivKey) []byte {
	t.Helper()
	record := settlementtypes.NewUsageRecord(
		"", message.OrderId, message.LeaseId, message.Sender, customer,
		message.UsageUnits, message.UsageType, time.Unix(message.PeriodStart, 0), time.Unix(message.PeriodEnd, 0),
		message.UnitPrice, nil, time.Time{}, 0,
	)
	record.AllocationID = message.AllocationId
	record.ChainID = message.ChainId
	record.Metrics = settlementtypes.RawUsageMetrics{
		CPUMilliSeconds:    message.RawMetrics.CpuMilliSeconds,
		MemoryByteSeconds:  message.RawMetrics.MemoryByteSeconds,
		StorageByteSeconds: message.RawMetrics.StorageByteSeconds,
		NetworkBytesIn:     message.RawMetrics.NetworkBytesIn,
		NetworkBytesOut:    message.RawMetrics.NetworkBytesOut,
		GPUSeconds:         message.RawMetrics.GpuSeconds,
	}
	record.PricingVersion = message.PricingVersion
	record.FormulaVersion = message.FormulaVersion
	record.ModelVersion = message.ModelVersion
	record.Sequence = message.StreamSequence
	record.Nonce = message.Nonce
	record.IdempotencyKey = message.IdempotencyKey
	record.ProviderKeyEpoch = message.ProviderKeyEpoch
	record.ProviderKeyID = message.ProviderKeyId
	record.IssuedAtHeight = message.IssuedAtHeight
	record.ExpiresAtHeight = message.ExpiresAtHeight
	record.IssuedAtUnix = message.IssuedAtUnix
	record.ExpiresAtUnix = message.ExpiresAtUnix
	record.SignatureVersion = message.SignatureVersion
	signBytes, err := settlementtypes.CanonicalUsageSignBytes(record.CanonicalUsagePayload(message.ChainId))
	require.NoError(t, err)
	signature, err := privateKey.Sign(signBytes)
	require.NoError(t, err)
	return signature
}

func assertTask84BAuthenticationActive(t *testing.T, net *networktest.Network) {
	t.Helper()
	for index := range net.Validators {
		status, err := net.Validators[index].RPCClient.Status(context.Background())
		require.NoError(t, err)
		appAtValidator, err := net.ValidatorApp(index)
		require.NoError(t, err)
		veApp := appAtValidator.(*app.VirtEngineApp)
		ctx := veApp.NewUncachedContext(false, cmtproto.Header{
			ChainID: net.Config.ChainID,
			Height:  status.SyncInfo.LatestBlockHeight,
			Time:    status.SyncInfo.LatestBlockTime,
		})
		require.True(t, veApp.Keepers.VirtEngine.Settlement.IsUsageAuthenticationActive(ctx),
			"validator %d fresh genesis did not activate Task 84B", index)
	}
}

func task84BValidatorAppHashes(t *testing.T, net *networktest.Network, height int64) ([]int64, [][]byte) {
	t.Helper()
	heights := make([]int64, len(net.Validators))
	hashes := make([][]byte, len(net.Validators))
	for index := range net.Validators {
		results := waitForTask84BBlockResults(t, net.Validators[index], height)
		require.Equal(t, height, results.Height)
		require.NotEmpty(t, results.AppHash)
		heights[index] = results.Height
		hashes[index] = append([]byte(nil), results.AppHash...)
	}
	return heights, hashes
}

func waitForTask84BBlockResults(t *testing.T, validator *networktest.Validator, height int64) *coretypes.ResultBlockResults {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		results, err := validator.RPCClient.BlockResults(ctx, &height)
		if err == nil {
			return results
		}
		select {
		case <-ctx.Done():
			t.Fatalf("validator %d results for height %d unavailable: %v", validator.Index, height, err)
		case <-ticker.C:
		}
	}
}

func countTask84BEventsInBlock(results *coretypes.ResultBlockResults, eventType string) int {
	count := countTask84BEvents(results.FinalizeBlockEvents, eventType)
	for _, result := range results.TxsResults {
		if result != nil {
			count += countTask84BEvents(result.Events, eventType)
		}
	}
	return count
}

func countTask84BEvents(events []abci.Event, eventType string) int {
	count := 0
	for _, event := range events {
		if event.Type == eventType || strings.HasSuffix(event.Type, "."+eventType) {
			count++
		}
	}
	return count
}

func totalTask84BUsageEvents(t *testing.T, validator *networktest.Validator, firstHeight, lastHeight int64) int {
	t.Helper()
	total := 0
	for height := firstHeight; height <= lastHeight; height++ {
		total += countTask84BEventsInBlock(waitForTask84BBlockResults(t, validator, height), settlementtypes.EventTypeUsageRecorded)
	}
	return total
}

func assertTask84BUsageCount(t *testing.T, query settlementv1.QueryClient, expected int) {
	t.Helper()
	response, err := query.UsageRecordsByOrder(context.Background(), &settlementv1.QueryUsageRecordsByOrderRequest{OrderId: task84BOrderID})
	require.NoError(t, err)
	require.Len(t, response.UsageRecords, expected)
}

func latestTask84BHeight(t *testing.T, validator *networktest.Validator) int64 {
	t.Helper()
	status, err := validator.RPCClient.Status(context.Background())
	require.NoError(t, err)
	return status.SyncInfo.LatestBlockHeight
}

func task84BAccountSequence(t *testing.T, cfg networktest.Config, validator *networktest.Validator, address sdk.AccAddress) uint64 {
	t.Helper()
	_, sequence, err := cfg.AccountRetriever.GetAccountNumberSequence(validator.ClientCtx.WithClient(validator.RPCClient), address)
	require.NoError(t, err)
	return sequence
}

func task84BBalance(t *testing.T, connection *grpc.ClientConn, address, denom string) string {
	t.Helper()
	response, err := banktypes.NewQueryClient(connection).Balance(context.Background(), &banktypes.QueryBalanceRequest{Address: address, Denom: denom})
	require.NoError(t, err)
	require.NotNil(t, response.Balance)
	return response.Balance.String()
}

func task84BEscrowBalance(t *testing.T, query settlementv1.QueryClient) string {
	t.Helper()
	response, err := query.EscrowsByOrder(context.Background(), &settlementv1.QueryEscrowsByOrderRequest{OrderId: task84BOrderID})
	require.NoError(t, err)
	require.Len(t, response.Escrows, 1)
	return response.Escrows[0].Balance.String()
}

func runTask84BSubprocess[T any](t *testing.T, mode string, input any) T {
	t.Helper()
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")
	outputPath := filepath.Join(dir, "output.json")
	writeTask84BJSON(t, inputPath, input)
	command := exec.Command(os.Args[0], "-test.run=^TestTask84BProcessHelper$") //nolint:gosec // fixed current test binary and selector
	command.Env = append(os.Environ(),
		task84BHelperEnv+"="+mode,
		"VE_TASK84B_HELPER_INPUT="+inputPath,
		"VE_TASK84B_HELPER_OUTPUT="+outputPath,
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, "Task 84B %s helper failed: %s", mode, output)
	var result T
	readTask84BJSON(t, outputPath, &result)
	return result
}

func readTask84BJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target))
}

func writeTask84BJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func prepareTask84BChainRetryQueue(t *testing.T, source, target string) {
	t.Helper()
	var state map[string]json.RawMessage
	readTask84BJSON(t, source, &state)
	// Keep usage_sequences and usage_proofs so the canonical nonce, sequence,
	// idempotency key, key epoch, expiry, and signature are reconstructed
	// exactly. Clear only producer-local queue suppression to exercise the
	// independent on-chain replay index.
	state["items"] = json.RawMessage(`{}`)
	writeTask84BJSON(t, target, state)
}

func buildTask84BSignedTx(chainID string, accountNumber, sequence uint64, privateKey cryptotypes.PrivKey, messages ...sdk.Msg) ([]byte, error) {
	encoding := sdkutil.MakeEncodingConfig()
	app.ModuleBasics().RegisterInterfaces(encoding.InterfaceRegistry)
	builder := encoding.TxConfig.NewTxBuilder()
	if err := builder.SetMsgs(messages...); err != nil {
		return nil, err
	}
	builder.SetGasLimit(task84BGas)
	placeholder := signing.SignatureV2{PubKey: privateKey.PubKey(), Data: &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}, Sequence: sequence}
	if err := builder.SetSignatures(placeholder); err != nil {
		return nil, err
	}
	signature, err := clienttx.SignWithPrivKey(context.Background(), signing.SignMode_SIGN_MODE_DIRECT, authsigning.SignerData{
		ChainID: chainID, AccountNumber: accountNumber, Sequence: sequence,
	}, builder, privateKey, encoding.TxConfig, sequence)
	if err != nil {
		return nil, err
	}
	if err := builder.SetSignatures(signature); err != nil {
		return nil, err
	}
	return encoding.TxConfig.TxEncoder()(builder.GetTx())
}

func broadcastMsgs(t *testing.T, cfg networktest.Config, validator *networktest.Validator, account networktest.TestAccount, sequence uint64, messages ...sdk.Msg) {
	t.Helper()
	privateKey := exportTestPrivateKey(t, validator.ClientCtx.Keyring, account.Name)
	broadcastWithPrivateKey(t, cfg, validator, account.Address, privateKey, sequence, messages...)
}

func broadcastWithPrivateKey(t *testing.T, cfg networktest.Config, validator *networktest.Validator, address sdk.AccAddress, privateKey cryptotypes.PrivKey, sequence uint64, messages ...sdk.Msg) {
	t.Helper()
	result := broadcastWithPrivateKeyRaw(t, cfg, validator, address, privateKey, sequence, messages...)
	require.Zero(t, result.CheckTx.Code, result.CheckTx.Log)
	require.Zero(t, result.TxResult.Code, result.TxResult.Log)
}

func broadcastWithPrivateKeyExpectDeliverFailure(t *testing.T, cfg networktest.Config, validator *networktest.Validator, address sdk.AccAddress, privateKey cryptotypes.PrivKey, sequence uint64, messages ...sdk.Msg) *coretypes.ResultBroadcastTxCommit {
	t.Helper()
	result := broadcastWithPrivateKeyRaw(t, cfg, validator, address, privateKey, sequence, messages...)
	require.Zero(t, result.CheckTx.Code, result.CheckTx.Log)
	require.NotZero(t, result.TxResult.Code, "transaction unexpectedly succeeded")
	return result
}

func broadcastWithPrivateKeyRaw(t *testing.T, cfg networktest.Config, validator *networktest.Validator, address sdk.AccAddress, privateKey cryptotypes.PrivKey, sequence uint64, messages ...sdk.Msg) *coretypes.ResultBroadcastTxCommit {
	t.Helper()
	accountNumber, _, err := cfg.AccountRetriever.GetAccountNumberSequence(validator.ClientCtx.WithClient(validator.RPCClient), address)
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
	return result
}

func exportTestPrivateKey(t *testing.T, keyringBackend keyring.Keyring, name string) cryptotypes.PrivKey {
	t.Helper()
	armor, err := keyringBackend.ExportPrivKeyArmor(name, "")
	require.NoError(t, err)
	privateKey, _, err := cosmoscrypto.UnarmorDecryptPrivKey(armor, "")
	require.NoError(t, err)
	return privateKey
}
