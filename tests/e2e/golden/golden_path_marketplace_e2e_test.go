//go:build e2e.integration

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
)

type goldenPathArtifacts struct {
	t       *testing.T
	logBuf  bytes.Buffer
	states  []artifactState
	baseDir string
}

type artifactState struct {
	Name  string      `json:"name"`
	State interface{} `json:"state"`
}

func newGoldenPathArtifacts(t *testing.T) *goldenPathArtifacts {
	t.Helper()

	baseDir := os.Getenv("VE_E2E_ARTIFACTS_DIR")
	if baseDir == "" {
		baseDir = filepath.Join("_build", "artifacts", "e2e")
	}

	a := &goldenPathArtifacts{t: t, baseDir: baseDir}
	t.Cleanup(func() {
		if t.Failed() {
			_ = a.write()
		}
	})

	return a
}

func (a *goldenPathArtifacts) logf(format string, args ...interface{}) {
	a.t.Helper()
	_, _ = fmt.Fprintf(&a.logBuf, format+"\n", args...)
}

func (a *goldenPathArtifacts) snapshot(name string, state interface{}) {
	a.t.Helper()
	a.states = append(a.states, artifactState{Name: name, State: state})
}

func (a *goldenPathArtifacts) write() error {
	a.t.Helper()

	artifactDir := filepath.Join(a.baseDir, sanitizeArtifactName(a.t.Name()))
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(artifactDir, "steps.log"), a.logBuf.Bytes(), 0o600); err != nil {
		return err
	}

	payload := struct {
		TestName string          `json:"test_name"`
		States   []artifactState `json:"states"`
	}{
		TestName: a.t.Name(),
		States:   a.states,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(artifactDir, "state.json"), data, 0o600)
}

func sanitizeArtifactName(name string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		default:
			return '_'
		}
	}, name)
}

//nolint:unparam
func retryWithBackoff(ctx context.Context, initial, max time.Duration, fn func() (bool, error)) error {
	delay := initial
	for {
		ok, err := fn()
		if ok {
			return nil
		}
		if err != nil && ctx.Err() != nil {
			return err
		}

		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		case <-time.After(delay):
		}

		delay *= 2
		if delay > max {
			delay = max
		}
	}
}

func TestGoldenPathMarketplaceProvisionUsageInvoicePayout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	artifacts := newGoldenPathArtifacts(t)

	config := fixtures.DefaultFixtureConfig()
	config.NumOfferings = 1
	config.NumOrders = 1
	config.NumJobs = 1

	fixture := fixtures.NewHPCProviderFixture(t, config)
	require.NoError(t, fixture.Setup())
	t.Cleanup(fixture.Teardown)

	artifacts.logf("fixture ready with provider=%s customer=%s", fixture.ProviderAddress, fixture.CustomerAddress)

	require.NotNil(t, fixture.Waldur)
	offerings := fixture.Waldur.GetOfferings()
	require.NotEmpty(t, offerings)
	selectedOffering := offerings[0]
	artifacts.logf("selected offering=%s", selectedOffering.OfferingID)

	order := fixture.CreateOrder(selectedOffering.OfferingID, "10.0")
	require.Equal(t, "open", order.Status)
	artifacts.logf("order created id=%s", order.OrderID)

	bid, err := fixture.PlaceBid(order.OrderID, "8.0")
	require.NoError(t, err)
	artifacts.logf("bid placed id=%s", bid.BidID)

	allocation, err := fixture.AcceptBid(order.OrderID, bid.BidID)
	require.NoError(t, err)
	artifacts.logf("allocation created id=%s", allocation.AllocationID)

	require.NoError(t, retryWithBackoff(ctx, 50*time.Millisecond, 500*time.Millisecond, func() (bool, error) {
		current := fixture.Waldur.GetOrder(order.OrderID)
		if current == nil {
			return false, fmt.Errorf("order %s missing", order.OrderID)
		}
		return current.Status == "matched", nil
	}))

	require.Equal(t, "provisioned", allocation.Status)

	job := fixture.CreateJob("golden", 4, 8192, 1)
	schedulerJob, err := fixture.SubmitJob(job)
	require.NoError(t, err)
	require.Equal(t, pd.HPCJobStatePending, schedulerJob.State)
	artifacts.logf("job submitted id=%s", job.JobID)

	fixture.Scheduler.SetJobState(job.JobID, pd.HPCJobStateRunning)
	require.NoError(t, retryWithBackoff(ctx, 50*time.Millisecond, 500*time.Millisecond, func() (bool, error) {
		status, err := fixture.Scheduler.GetJobStatus(ctx, job.JobID)
		if err != nil {
			return false, err
		}
		return status.State == pd.HPCJobStateRunning, nil
	}))

	metrics := &pd.HPCSchedulerMetrics{
		WallClockSeconds: 3600,
		CPUCoreSeconds:   14400,
		MemoryGBSeconds:  28800,
		GPUSeconds:       3600,
		NodesUsed:        1,
		NodeHours:        1.0,
	}
	fixture.Scheduler.SetJobMetrics(job.JobID, metrics)
	fixture.Scheduler.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	require.NoError(t, retryWithBackoff(ctx, 50*time.Millisecond, 500*time.Millisecond, func() (bool, error) {
		status, err := fixture.Scheduler.GetJobStatus(ctx, job.JobID)
		if err != nil {
			return false, err
		}
		return status.State == pd.HPCJobStateCompleted, nil
	}))

	usageRecord := fixture.GenerateUsageRecord(job.JobID, metrics)
	artifacts.logf("usage record created id=%s", usageRecord.RecordID)

	require.NoError(t, retryWithBackoff(ctx, 50*time.Millisecond, 500*time.Millisecond, func() (bool, error) {
		records := fixture.UsageReporter.GetRecordsForJob(job.JobID)
		return len(records) == 1 && records[0].RecordID == usageRecord.RecordID, nil
	}))

	cpuCost, err := sdkmath.LegacyNewDecFromStr("1.44")
	require.NoError(t, err)
	memCost, err := sdkmath.LegacyNewDecFromStr("0.288")
	require.NoError(t, err)
	gpuCost, err := sdkmath.LegacyNewDecFromStr("3.6")
	require.NoError(t, err)

	totalAmount := cpuCost.Add(memCost).Add(gpuCost)

	lineItems := []fixtures.MockLineItem{
		{
			ResourceType: "cpu",
			Quantity:     sdkmath.LegacyNewDec(14400),
			UnitPrice:    "0.0001",
			TotalCost:    cpuCost.String(),
		},
		{
			ResourceType: "memory",
			Quantity:     sdkmath.LegacyNewDec(28800),
			UnitPrice:    "0.00001",
			TotalCost:    memCost.String(),
		},
		{
			ResourceType: "gpu",
			Quantity:     sdkmath.LegacyNewDec(3600),
			UnitPrice:    "0.001",
			TotalCost:    gpuCost.String(),
		},
	}

	invoice := fixture.CreateInvoice(order.OrderID, lineItems, totalAmount.String())
	require.Equal(t, "pending", invoice.Status)
	artifacts.logf("invoice created id=%s total=%s", invoice.InvoiceID, invoice.TotalAmount)

	require.NoError(t, fixture.Settlement.SettleInvoice(invoice.InvoiceID))

	settled := fixture.Settlement.GetInvoice(invoice.InvoiceID)
	require.Equal(t, "settled", settled.Status)

	payout := fixture.Settlement.GetPayout(invoice.InvoiceID)
	require.NotNil(t, payout)
	require.Equal(t, "completed", payout.Status)

	fee := fixture.Settlement.GetFee(invoice.InvoiceID)
	require.NotNil(t, fee)

	feeRate, err := sdkmath.LegacyNewDecFromStr("0.025")
	require.NoError(t, err)
	expectedFee := totalAmount.Mul(feeRate)
	expectedPayout := totalAmount.Sub(expectedFee)

	feeDec, err := sdkmath.LegacyNewDecFromStr(fee.Amount)
	require.NoError(t, err)
	payoutDec, err := sdkmath.LegacyNewDecFromStr(payout.Amount)
	require.NoError(t, err)

	require.True(t, feeDec.Equal(expectedFee), "expected fee %s got %s", expectedFee, feeDec)
	require.True(t, payoutDec.Equal(expectedPayout), "expected payout %s got %s", expectedPayout, payoutDec)

	artifacts.snapshot("order", order)
	artifacts.snapshot("bid", bid)
	artifacts.snapshot("allocation", allocation)
	artifacts.snapshot("job", job)
	artifacts.snapshot("usage", usageRecord)
	artifacts.snapshot("invoice", settled)
	artifacts.snapshot("payout", payout)
	artifacts.snapshot("fee", fee)
}
