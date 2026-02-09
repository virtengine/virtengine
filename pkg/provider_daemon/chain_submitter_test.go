package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func (m *mockSubmitterClient) BroadcastTx(_ context.Context, tx []byte) error {
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
		return err
	}
	return nil
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
	_, err = km.GenerateKey("provider-1")
	require.NoError(t, err)
	return km
}

func newTestReport() *ChainUsageReport {
	now := time.Now()
	return &ChainUsageReport{
		OrderID:     "order-1",
		LeaseID:     "lease-1",
		UsageUnits:  10,
		UsageType:   "cpu",
		PeriodStart: now.Add(-time.Hour),
		PeriodEnd:   now,
		UnitPrice:   sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(1)),
		Signature:   []byte("sig"),
	}
}

func newSubmitterWithClient(t *testing.T, client ChainSubmitterClient, cfgOverrides func(*ChainSubmitterConfig)) *ChainUsageSubmitterImpl {
	t.Helper()
	cfg := DefaultChainSubmitterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "provider-1"
	cfg.ChainID = "virtengine-test"
	cfg.ChainClient = client
	cfg.CometRPC = ""
	cfg.RetryBackoff = 0
	if cfgOverrides != nil {
		cfgOverrides(&cfg)
	}
	submitter, err := NewChainUsageSubmitter(cfg, newTestKeyManager(t), nil)
	require.NoError(t, err)
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
	cfg.ProviderAddress = "provider-1"
	cfg.CometRPC = ""
	cfg.ChainClient = mockClient
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
	assert.Equal(t, "virtengine-test", env.ChainID)
}

func TestChainSubmitterBatchingSingleTx(t *testing.T) {
	mockClient := newMockSubmitterClient(100000)
	submitter := newSubmitterWithClient(t, mockClient, nil)

	reports := []*ChainUsageReport{newTestReport(), newTestReport()}
	err := submitter.submitBatch(context.Background(), reports)
	require.NoError(t, err)

	broadcastCalls, _ := mockClient.Calls()
	assert.Equal(t, 1, broadcastCalls)

	var env txEnvelope
	require.NoError(t, json.Unmarshal(mockClient.LastTx(), &env))
	var batch struct {
		Msgs []MsgRecordUsageWrapper `json:"msgs"`
	}
	require.NoError(t, json.Unmarshal(env.Msg, &batch))
	assert.Len(t, batch.Msgs, 2)
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
		ProviderAddress: "provider-1",
	}, km)
	report := newTestReport()
	txBytes, err := builder.BuildUsageReportTx(report, SigningData{ChainID: "virtengine-test", Sequence: 3})
	require.NoError(t, err)
	var tx map[string]interface{}
	require.NoError(t, json.Unmarshal(txBytes, &tx))
	assert.Equal(t, "virtengine-test", tx["chain_id"])
}

func TestSignatureVerifierAndHash(t *testing.T) {
	verifier := NewSignatureVerifier()
	report := newTestReport()
	report.Signature = []byte("sig")

	verifier.AddTrustedProvider("provider-1", []byte("pub"))
	ok, err := verifier.VerifyUsageReport(report, "provider-1")
	require.NoError(t, err)
	assert.True(t, ok)

	hash := UsageReportHashHex(report)
	assert.NotEmpty(t, hash)
}
