package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/waldur"
)

type mockUsageChainClient struct {
	mu         sync.Mutex
	gasLimit   uint64
	escrow     map[string]int64
	records    []MsgRecordUsageWrapper
	recordIDs  []string
	settles    []MsgSettleOrderWrapper
	broadcasts int
}

func newMockUsageChainClient(gas uint64) *mockUsageChainClient {
	return &mockUsageChainClient{
		gasLimit: gas,
		escrow:   make(map[string]int64),
	}
}

func (m *mockUsageChainClient) EstimateGas(_ context.Context, _ []byte) (uint64, error) {
	return m.gasLimit, nil
}

func (m *mockUsageChainClient) BroadcastTx(_ context.Context, tx []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcasts++

	var env txEnvelope
	if err := json.Unmarshal(tx, &env); err != nil {
		return err
	}

	if m.tryHandleBatch(env.Msg) {
		return nil
	}

	if m.tryHandleUsage(env.Msg) {
		return nil
	}

	if m.tryHandleSettlement(env.Msg) {
		return nil
	}

	return nil
}

func (m *mockUsageChainClient) tryHandleBatch(msg json.RawMessage) bool {
	var batch struct {
		Msgs []MsgRecordUsageWrapper `json:"msgs"`
	}
	if err := json.Unmarshal(msg, &batch); err != nil || len(batch.Msgs) == 0 {
		return false
	}
	for _, usage := range batch.Msgs {
		m.applyUsage(usage)
	}
	return true
}

func (m *mockUsageChainClient) tryHandleUsage(msg json.RawMessage) bool {
	var usage MsgRecordUsageWrapper
	if err := json.Unmarshal(msg, &usage); err != nil || usage.OrderID == "" {
		return false
	}
	if usage.LeaseID == "" || usage.UsageType == "" {
		return false
	}
	m.applyUsage(usage)
	return true
}

func (m *mockUsageChainClient) tryHandleSettlement(msg json.RawMessage) bool {
	var settle MsgSettleOrderWrapper
	if err := json.Unmarshal(msg, &settle); err != nil || settle.OrderID == "" {
		return false
	}
	m.settles = append(m.settles, settle)
	return true
}

func (m *mockUsageChainClient) applyUsage(usage MsgRecordUsageWrapper) {
	m.records = append(m.records, usage)
	hash := sha256.Sum256([]byte(usage.OrderID + usage.LeaseID + usage.UsageType))
	m.recordIDs = append(m.recordIDs, hex.EncodeToString(hash[:8]))
	if usage.UnitPrice.Denom == "" || usage.UsageUnits == 0 {
		return
	}
	if usage.UsageUnits > math.MaxInt64 {
		return
	}
	cost := usage.UnitPrice.Amount.MulInt64(int64(usage.UsageUnits)).TruncateInt64()
	m.escrow[usage.OrderID] -= cost
}

func TestChainSubmitterIntegrationUsageLifecycle(t *testing.T) {
	client := newMockUsageChainClient(250000)
	client.escrow["order-1"] = 1000

	submitter := newSubmitterWithClient(t, client, func(cfg *ChainSubmitterConfig) {
		cfg.EnableIdempotency = true
	})

	usageRecord := &UsageRecord{
		ID:           "usage-1",
		LeaseID:      "lease-1",
		DeploymentID: "order-1",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		Type:         UsageRecordTypePeriodic,
		Metrics:      ResourceMetrics{CPUMilliSeconds: 36000000},
		PricingInputs: PricingInputs{
			AgreedCPURate: "1",
		},
	}

	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), nil)
	reports := pipeline.buildUsageReports(usageRecord)
	require.NotEmpty(t, reports)

	err := submitter.submitSingleReport(context.Background(), reports[0])
	require.NoError(t, err)

	client.mu.Lock()
	escrowRemaining := client.escrow["order-1"]
	client.mu.Unlock()
	require.Equal(t, int64(990), escrowRemaining)

	err = submitter.SubmitSettlementRequest(context.Background(), "order-1", client.recordIDs, true)
	require.NoError(t, err)
	require.Len(t, client.settles, 1)
}

func TestChainSubmitterIntegrationWaldurUsageEvent(t *testing.T) {
	client := newMockUsageChainClient(200000)
	client.escrow["order-2"] = 500

	submitter := newSubmitterWithClient(t, client, nil)
	waldurReport := &waldur.ResourceUsageReport{
		ResourceUUID: "resource-1",
		PeriodStart:  time.Now().Add(-time.Hour),
		PeriodEnd:    time.Now(),
		Components: []waldur.ComponentUsage{
			{Type: "cpu", Amount: 5},
		},
		BackendID: "order-2",
	}

	report := &ChainUsageReport{
		OrderID:     waldurReport.BackendID,
		LeaseID:     "lease-2",
		UsageUnits:  uint64(waldurReport.Components[0].Amount),
		UsageType:   waldurReport.Components[0].Type,
		PeriodStart: waldurReport.PeriodStart,
		PeriodEnd:   waldurReport.PeriodEnd,
		UnitPrice:   sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(1)),
		Signature:   []byte("waldur"),
	}

	err := submitter.submitSingleReport(context.Background(), report)
	require.NoError(t, err)

	client.mu.Lock()
	defer client.mu.Unlock()
	require.Equal(t, int64(495), client.escrow["order-2"])
	require.Len(t, client.records, 1)
}
