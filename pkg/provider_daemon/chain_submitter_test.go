package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/x/settlement"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const (
	testSubmitterProviderAddress = "provider-1"
	testSubmitterChainID         = "virtengine-test"
)

type testSigningStateResolver struct{ binding ActiveProviderKeyBinding }

func (r testSigningStateResolver) ResolveProviderSigningState(context.Context, string) (ActiveProviderKeyBinding, error) {
	return r.binding, nil
}

type testUsageStateResolver struct {
	states map[string]OnChainUsageStreamState
}

func (r testUsageStateResolver) ResolveUsageStreamState(_ context.Context, provider, allocationID, orderID, leaseID string) (OnChainUsageStreamState, error) {
	streamID, err := settlementtypes.UsageStreamID(provider, allocationID, orderID, leaseID)
	if err != nil {
		return OnChainUsageStreamState{}, err
	}
	return r.states[hex.EncodeToString(streamID)], nil
}

func testSigningResolver(t *testing.T, km *KeyManager) ProviderSigningStateResolver {
	t.Helper()
	key, err := km.GetActiveKey()
	require.NoError(t, err)
	publicKey, err := hex.DecodeString(key.PublicKey)
	require.NoError(t, err)
	return testSigningStateResolver{binding: ActiveProviderKeyBinding{
		ProviderAddress: key.ProviderAddress,
		KeyID:           key.KeyID,
		Epoch:           1,
		PublicKey:       publicKey,
		Algorithm:       key.Algorithm,
		BlockHeight:     100,
		BlockTime:       time.Now().UTC(),
	}}
}

type mockSubmitterClient struct {
	mu             sync.Mutex
	gasLimit       uint64
	broadcastCalls int
	estimateCalls  int
	broadcastErrs  []error
	txs            [][]byte
	sequences      []uint64
}

func newMockSubmitterClient(gas uint64, errs ...error) *mockSubmitterClient {
	return &mockSubmitterClient{
		gasLimit:      gas,
		broadcastErrs: append([]error(nil), errs...),
	}
}

func (m *mockSubmitterClient) EstimateGas(_ context.Context, tx []byte) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.estimateCalls++
	m.txs = append(m.txs, tx)
	return m.gasLimit, nil
}

func (m *mockSubmitterClient) BroadcastTx(_ context.Context, tx []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcastCalls++
	m.txs = append(m.txs, tx)
	var env txEnvelope
	if err := json.Unmarshal(tx, &env); err == nil {
		m.sequences = append(m.sequences, env.Sequence)
	}
	if len(m.broadcastErrs) > 0 {
		err := m.broadcastErrs[0]
		m.broadcastErrs = m.broadcastErrs[1:]
		return "", err
	}
	return "tx-hash-mock", nil
}

func (m *mockSubmitterClient) Calls() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.broadcastCalls, m.estimateCalls
}

func (m *mockSubmitterClient) LastTx() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.txs) == 0 {
		return nil
	}
	return m.txs[len(m.txs)-1]
}

func (m *mockSubmitterClient) Sequences() []uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]uint64(nil), m.sequences...)
}

func newTestKeyManager(t *testing.T) *KeyManager {
	t.Helper()
	km, err := NewKeyManager(KeyManagerConfig{
		StorageType:      KeyStorageTypeMemory,
		DefaultAlgorithm: string(HSMKeyTypeEd25519),
		KeyRotationDays:  1,
	})
	require.NoError(t, err)
	require.NoError(t, km.Unlock(""))
	_, err = km.GenerateKey(testSubmitterProviderAddress)
	require.NoError(t, err)
	return km
}

func newTestReport() *ChainUsageReport {
	now := time.Now()
	return &ChainUsageReport{
		OrderID:         "order-1",
		LeaseID:         "lease-1",
		CustomerAddress: "customer-1",
		UsageUnits:      10,
		UsageType:       "cpu",
		PeriodStart:     now.Add(-time.Hour),
		PeriodEnd:       now,
		UnitPrice:       sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(1)),
		RawMetrics:      ResourceMetrics{CPUMilliSeconds: 36_000_000},
	}
}

func reportToUsageMsg(report *ChainUsageReport) *MsgRecordUsageWrapper {
	return &MsgRecordUsageWrapper{
		Sender:           testSubmitterProviderAddress,
		OrderID:          report.OrderID,
		LeaseID:          report.LeaseID,
		UsageUnits:       report.UsageUnits,
		UsageType:        report.UsageType,
		PeriodStart:      report.PeriodStart.Unix(),
		PeriodEnd:        report.PeriodEnd.Unix(),
		UnitPrice:        report.UnitPrice,
		Signature:        report.Signature,
		AllocationID:     report.AllocationID,
		ChainID:          report.ChainID,
		RawMetrics:       report.RawMetrics,
		PricingVersion:   report.PricingVersion,
		FormulaVersion:   report.FormulaVersion,
		ModelVersion:     report.ModelVersion,
		StreamSequence:   report.StreamSequence,
		Nonce:            report.Nonce,
		IdempotencyKey:   report.IdempotencyKey,
		ProviderKeyEpoch: report.ProviderKeyEpoch,
		ProviderKeyID:    report.ProviderKeyID,
		IssuedAtHeight:   report.IssuedAtHeight,
		ExpiresAtHeight:  report.ExpiresAtHeight,
		IssuedAtUnix:     report.IssuedAtUnix,
		ExpiresAtUnix:    report.ExpiresAtUnix,
		SignatureVersion: report.SignatureVersion,
	}
}

func newSubmitterWithClient(t *testing.T, client ChainSubmitterClient, cfgOverrides func(*ChainSubmitterConfig)) *ChainUsageSubmitterImpl {
	t.Helper()
	cfg := DefaultChainSubmitterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testSubmitterProviderAddress
	cfg.ChainID = testSubmitterChainID
	cfg.ChainClient = client
	cfg.AllowTestLegacyChainClient = true
	cfg.CometRPC = ""
	cfg.RetryBackoff = 0
	cfg.QueueStatePath = filepath.Join(t.TempDir(), "queue-state.json")
	if cfgOverrides != nil {
		cfgOverrides(&cfg)
	}
	km := newTestKeyManager(t)
	if cfg.ProviderSigningState == nil {
		cfg.ProviderSigningState = testSigningResolver(t, km)
	}
	submitter, err := NewChainUsageSubmitter(cfg, km, nil)
	require.NoError(t, err)
	t.Cleanup(submitter.Stop)
	return submitter
}

func TestChainSubmitterInitialization(t *testing.T) {
	cfg := DefaultChainSubmitterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = ""
	cfg.CometRPC = "http://localhost:26657"
	_, err := NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.Error(t, err)

	mockClient := newMockSubmitterClient(0)
	cfg.ProviderAddress = testSubmitterProviderAddress
	cfg.CometRPC = ""
	cfg.ChainClient = mockClient
	_, err = NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.ErrorIs(t, err, ErrProviderMutationUnavailable)
	cfg.AllowTestLegacyChainClient = true
	submitter, err := NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.NoError(t, err)
	require.NotNil(t, submitter.chainClient)
}

func TestChainSubmitterSignsAndBroadcasts(t *testing.T) {
	mockClient := newMockSubmitterClient(150000)
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.GasLimit = 200000
	})

	err := submitter.submitSingleReport(context.Background(), newTestReport())
	require.NoError(t, err)

	broadcastCalls, estimateCalls := mockClient.Calls()
	assert.Equal(t, 1, broadcastCalls)
	assert.Equal(t, 1, estimateCalls)

	var env txEnvelope
	require.NoError(t, json.Unmarshal(mockClient.LastTx(), &env))
	assert.Equal(t, uint64(150000), env.GasLimit)
	assert.Equal(t, testSubmitterChainID, env.ChainID)
}

func TestChainSubmitterBatchQueuesPerReport(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, nil)

	reports := []*ChainUsageReport{newTestReport(), newTestReport()}
	reports[1].LeaseID = "lease-2"
	err := submitter.submitBatch(context.Background(), reports)
	require.NoError(t, err)

	broadcastCalls, _ := mockClient.Calls()
	assert.Equal(t, 2, broadcastCalls)
	assert.Equal(t, 2, len(submitter.queueState.Items))
}

func TestChainSubmitterBuildsRealSDKTxWhenNotLegacy(t *testing.T) {
	submitter := newSubmitterWithClient(t, newMockSubmitterClient(100000), nil)
	submitter.useLegacyEnvelope = false
	submitter.encCfg = sdkutil.MakeEncodingConfig(settlement.AppModuleBasic{})

	bz, err := submitter.buildSignedTx(reportToUsageMsg(newTestReport()), 150000)
	require.NoError(t, err)

	tx, err := submitter.encCfg.TxConfig.TxDecoder()(bz)
	require.NoError(t, err)
	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1)
	usageMsg, ok := msgs[0].(*settlementv1.MsgRecordUsage)
	require.True(t, ok)
	assert.Equal(t, "order-1", usageMsg.OrderId)
	assert.Equal(t, uint64(10), usageMsg.UsageUnits)
}

func TestChainSubmitterRetryOnBroadcastFailure(t *testing.T) {
	mockClient := newMockSubmitterClient(100000, errors.New("network"), ErrSequenceMismatch)
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.RetryAttempts = 3
		cfg.Sequence = 10
	})

	err := submitter.submitSingleReport(context.Background(), newTestReport())
	require.NoError(t, err)

	seqs := mockClient.Sequences()
	require.Len(t, seqs, 3)
	assert.Equal(t, uint64(10), seqs[0])
	assert.Equal(t, uint64(10), seqs[1])
	assert.Equal(t, uint64(11), seqs[2])
}

func TestChainSubmitterIdempotency(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.EnableIdempotency = true
	})

	report := newTestReport()
	require.NoError(t, submitter.submitSingleReport(context.Background(), report))
	err := submitter.submitSingleReport(context.Background(), report)
	assert.ErrorIs(t, err, ErrDuplicateReport)
}

func TestChainSubmitterValidationErrors(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, nil)

	err := submitter.submitSingleReport(context.Background(), nil)
	assert.ErrorIs(t, err, ErrInvalidReport)

	report := newTestReport()
	report.LeaseID = ""
	err = submitter.submitSingleReport(context.Background(), report)
	assert.ErrorIs(t, err, ErrInvalidReport)

	expiredSubmitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.ReportValidator = func(_ *ChainUsageReport) error {
			return ErrLeaseExpired
		}
	})
	err = expiredSubmitter.submitSingleReport(context.Background(), newTestReport())
	assert.ErrorIs(t, err, ErrLeaseExpired)
}

func TestChainSubmitterGasEstimation(t *testing.T) {
	mockClient := newMockSubmitterClient(424242)
	submitter := newSubmitterWithClient(t, mockClient, nil)

	err := submitter.submitSingleReport(context.Background(), newTestReport())
	require.NoError(t, err)

	var env txEnvelope
	require.NoError(t, json.Unmarshal(mockClient.LastTx(), &env))
	assert.Equal(t, uint64(424242), env.GasLimit)
}

func TestChainSubmitterConcurrentSubmissionSafety(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			report := newTestReport()
			report.OrderID = report.OrderID + "-" + sdkmath.NewInt(int64(idx)).String()
			_ = submitter.submitSingleReport(context.Background(), report)
		}(i)
	}
	wg.Wait()

	broadcastCalls, _ := mockClient.Calls()
	assert.Equal(t, 10, broadcastCalls)
}

func TestTransactionBuilderBuildUsageReportTx(t *testing.T) {
	km := newTestKeyManager(t)
	builder := NewTransactionBuilder(ChainSubmitterConfig{
		ProviderAddress: testSubmitterProviderAddress,
	}, km)
	report := newTestReport()
	txBytes, err := builder.BuildUsageReportTx(report, SigningData{ChainID: testSubmitterChainID, Sequence: 3})
	require.NoError(t, err)
	var tx map[string]interface{}
	require.NoError(t, json.Unmarshal(txBytes, &tx))
	assert.Equal(t, testSubmitterChainID, tx["chain_id"])
}

func TestSignatureVerifierAndHash(t *testing.T) {
	submitter := newSubmitterWithClient(t, newMockSubmitterClient(100000), nil)
	verifier := NewSignatureVerifier()
	report := newTestReport()
	require.NoError(t, submitter.prepareAuthenticatedUsageReport(context.Background(), report))
	key, err := submitter.keyManager.GetActiveKey()
	require.NoError(t, err)
	publicKey, err := hex.DecodeString(key.PublicKey)
	require.NoError(t, err)

	verifier.AddTrustedProvider(testSubmitterProviderAddress, publicKey)
	ok, err := verifier.VerifyUsageReport(report, testSubmitterProviderAddress)
	require.NoError(t, err)
	assert.True(t, ok)
	report.RawMetrics.CPUMilliSeconds += 3_600_000
	report.UsageUnits++
	ok, err = verifier.VerifyUsageReport(report, testSubmitterProviderAddress)
	require.NoError(t, err)
	assert.False(t, ok)

	hash := UsageReportHashHex(report)
	assert.NotEmpty(t, hash)
}

func TestChainSubmitterQueuePersistenceAcrossRestart(t *testing.T) {
	report := newTestReport()
	queuePath := filepath.Join(t.TempDir(), "queue-state.json")
	cfg := DefaultChainSubmitterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testSubmitterProviderAddress
	cfg.ChainID = testSubmitterChainID
	cfg.CometRPC = ""
	cfg.QueueStatePath = queuePath
	cfg.ChainClient = newMockSubmitterClient(100000)
	cfg.AllowTestLegacyChainClient = true

	submitter1, err := NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.NoError(t, err)
	item, existed, err := submitter1.enqueueMessage(queueItemKindUsage, reportToUsageMsg(report))
	require.NoError(t, err)
	require.False(t, existed)
	submitter1.Stop()

	mockClient := newMockSubmitterClient(100000)
	cfg.ChainClient = mockClient
	cfg.AllowTestLegacyChainClient = true
	submitter2, err := NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.NoError(t, err)
	t.Cleanup(submitter2.Stop)
	require.NoError(t, submitter2.processQueueItem(context.Background(), item.IdempotencyKey, false))

	broadcastCalls, _ := mockClient.Calls()
	assert.Equal(t, 1, broadcastCalls)
	require.Contains(t, submitter2.queueState.Items, item.IdempotencyKey)
	assert.Equal(t, queueItemStatusBroadcasted, submitter2.queueState.Items[item.IdempotencyKey].Status)
	assert.NotEmpty(t, submitter2.queueState.Items[item.IdempotencyKey].BroadcastTxHash)
}

func TestAuthenticatedUsageProofAllocationPersistsAcrossRestart(t *testing.T) {
	queuePath := filepath.Join(t.TempDir(), "queue-state.json")
	original := newTestReport()
	km1 := newTestKeyManager(t)
	active1, err := km1.GetActiveKey()
	require.NoError(t, err)
	privateKey := append([]byte(nil), active1.privateKey...)

	cfg := DefaultChainSubmitterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testSubmitterProviderAddress
	cfg.ChainID = testSubmitterChainID
	cfg.ChainClient = newMockSubmitterClient(100000)
	cfg.AllowTestLegacyChainClient = true
	cfg.QueueStatePath = queuePath
	cfg.ProviderSigningState = testSigningResolver(t, km1)
	submitter1, err := NewChainUsageSubmitter(cfg, km1, nil)
	require.NoError(t, err)
	first := *original
	require.NoError(t, submitter1.prepareAuthenticatedUsageReport(context.Background(), &first))
	submitter1.Stop()

	km2, err := NewKeyManager(KeyManagerConfig{StorageType: KeyStorageTypeMemory, DefaultAlgorithm: string(HSMKeyTypeEd25519)})
	require.NoError(t, err)
	require.NoError(t, km2.Unlock(""))
	_, err = km2.ImportKey(testSubmitterProviderAddress, privateKey, string(HSMKeyTypeEd25519))
	require.NoError(t, err)
	cfg.ProviderSigningState = testSigningResolver(t, km2)
	submitter2, err := NewChainUsageSubmitter(cfg, km2, nil)
	require.NoError(t, err)
	t.Cleanup(submitter2.Stop)
	second := *original
	require.NoError(t, submitter2.prepareAuthenticatedUsageReport(context.Background(), &second))
	require.Equal(t, first.StreamSequence, second.StreamSequence)
	require.Equal(t, first.Nonce, second.Nonce)
	require.Equal(t, first.IdempotencyKey, second.IdempotencyKey)
	require.Equal(t, first.Signature, second.Signature)
}

func TestAuthenticatedUsageFailedPreparationDoesNotConsumeSequence(t *testing.T) {
	submitter := newSubmitterWithClient(t, newMockSubmitterClient(100000), nil)
	first := newTestReport()
	first.UnitPrice = sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyZeroDec())
	require.Error(t, submitter.prepareAuthenticatedUsageReport(context.Background(), first))

	second := newTestReport()
	require.NoError(t, submitter.prepareAuthenticatedUsageReport(context.Background(), second))
	require.Equal(t, uint64(1), second.StreamSequence)
}

func TestAuthenticatedUsageReconcilesSequenceFromChain(t *testing.T) {
	submitter := newSubmitterWithClient(t, newMockSubmitterClient(100000), nil)
	report := newTestReport()
	streamID, err := settlementtypes.UsageStreamID(testSubmitterProviderAddress, report.AllocationID, report.OrderID, report.LeaseID)
	require.NoError(t, err)
	submitter.cfg.UsageStreamState = testUsageStateResolver{states: map[string]OnChainUsageStreamState{
		hex.EncodeToString(streamID): {LastSequence: 7},
	}}

	require.NoError(t, submitter.prepareAuthenticatedUsageReport(context.Background(), report))
	require.Equal(t, uint64(8), report.StreamSequence)
}

func TestAuthenticatedUsageRefreshesExpiredUnbroadcastProof(t *testing.T) {
	queuePath := filepath.Join(t.TempDir(), "queue-state.json")
	km := newTestKeyManager(t)
	resolver := testSigningResolver(t, km).(testSigningStateResolver)
	resolver.binding.BlockTime = time.Now().UTC()
	cfg := DefaultChainSubmitterConfig()
	cfg.ProviderAddress = testSubmitterProviderAddress
	cfg.ChainID = testSubmitterChainID
	cfg.ChainClient = newMockSubmitterClient(100000)
	cfg.AllowTestLegacyChainClient = true
	cfg.QueueStatePath = queuePath
	cfg.ProviderSigningState = resolver
	submitter, err := NewChainUsageSubmitter(cfg, km, nil)
	require.NoError(t, err)
	t.Cleanup(submitter.Stop)

	report := newTestReport()
	require.NoError(t, submitter.prepareAuthenticatedUsageReport(context.Background(), report))
	oldExpiry := report.ExpiresAtHeight

	resolver.binding.BlockHeight = oldExpiry + 1
	resolver.binding.BlockTime = resolver.binding.BlockTime.Add(time.Minute)
	submitter.cfg.ProviderSigningState = resolver
	refreshed := newTestReport()
	refreshed.PeriodStart = report.PeriodStart
	refreshed.PeriodEnd = report.PeriodEnd
	require.NoError(t, submitter.prepareAuthenticatedUsageReport(context.Background(), refreshed))
	require.Equal(t, report.StreamSequence, refreshed.StreamSequence)
	require.Greater(t, refreshed.ExpiresAtHeight, oldExpiry)
	require.NotEqual(t, report.Signature, refreshed.Signature)
}

func TestChainSubmitterSharedQueueInstancesDoNotOverwrite(t *testing.T) {
	queuePath := filepath.Join(t.TempDir(), "queue-state.json")
	newSharedSubmitter := func() (*ChainUsageSubmitterImpl, error) {
		cfg := DefaultChainSubmitterConfig()
		cfg.ProviderAddress = testSubmitterProviderAddress
		cfg.ChainID = testSubmitterChainID
		cfg.ChainClient = newMockSubmitterClient(100000)
		cfg.AllowTestLegacyChainClient = true
		cfg.QueueStatePath = queuePath
		km := newTestKeyManager(t)
		cfg.ProviderSigningState = testSigningResolver(t, km)
		return NewChainUsageSubmitter(cfg, km, nil)
	}

	first, err := newSharedSubmitter()
	require.NoError(t, err)
	t.Cleanup(first.Stop)
	second, err := newSharedSubmitter()
	require.Nil(t, second)
	require.ErrorContains(t, err, "already owned")
}

func TestAuthenticatedUsageRefusesLocalKeyMismatch(t *testing.T) {
	submitter := newSubmitterWithClient(t, newMockSubmitterClient(100000), nil)
	resolver := submitter.cfg.ProviderSigningState.(testSigningStateResolver)
	resolver.binding.KeyID = "ed25519:wrong"
	submitter.cfg.ProviderSigningState = resolver
	err := submitter.submitSingleReport(context.Background(), newTestReport())
	require.ErrorIs(t, err, ErrProviderKeyMismatch)
}

func TestChainSubmitterDuplicateDeliveryProtection(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.EnableIdempotency = true
	})
	report := newTestReport()
	require.NoError(t, submitter.submitSingleReport(context.Background(), report))
	err := submitter.submitSingleReport(context.Background(), report)
	require.ErrorIs(t, err, ErrDuplicateReport)
	broadcastCalls, _ := mockClient.Calls()
	assert.Equal(t, 1, broadcastCalls)
}

func TestChainSubmitterTerminalNonRetryableFailure(t *testing.T) {
	mockClient := newMockSubmitterClient(100000, &classifiedBroadcastError{
		Message:   "insufficient funds",
		Retryable: false,
	})
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.RetryBackoff = 0
		cfg.MaxAttempts = 4
	})
	report := newTestReport()
	err := submitter.submitSingleReport(context.Background(), report)
	require.Error(t, err)

	key := submitter.computeIdempotencyKey(queueItemKindUsage, mustMarshal(t, reportToUsageMsg(report)))
	item, ok := submitter.queueState.Items[key]
	require.True(t, ok)
	assert.Equal(t, queueItemStatusFailed, item.Status)
	assert.Equal(t, 1, item.AttemptCount)
}

func TestChainSubmitterRetryableFailureSchedulesRetry(t *testing.T) {
	mockClient := newMockSubmitterClient(100000, errors.New("timeout"), nil)
	submitter := newSubmitterWithClient(t, mockClient, func(cfg *ChainSubmitterConfig) {
		cfg.RetryBackoff = 0
		cfg.MaxAttempts = 3
	})
	report := newTestReport()
	item, existed, err := submitter.enqueueMessage(queueItemKindUsage, reportToUsageMsg(report))
	require.NoError(t, err)
	require.False(t, existed)

	require.NoError(t, submitter.processQueueItem(context.Background(), item.IdempotencyKey, false))
	stored := submitter.queueState.Items[item.IdempotencyKey]
	require.Equal(t, queueItemStatusRetryableFailed, stored.Status)

	require.NoError(t, submitter.processQueueItem(context.Background(), item.IdempotencyKey, false))
	stored = submitter.queueState.Items[item.IdempotencyKey]
	require.Equal(t, queueItemStatusBroadcasted, stored.Status)
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return out
}
