//go:build e2e.integration

package waldur

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

func TestWaldurDuplicateBackendConflict(t *testing.T) {
	h := newWaldurHarness(t, nil)

	backendID := h.providerAddr + "/duplicate-offering"
	h.createOffering("Duplicate One", backendID, "VirtEngine.Compute", nil)
	h.createOffering("Duplicate Two", backendID, "VirtEngine.Compute", nil)

	ctx, cancel := h.context()
	_, err := h.marketplace.GetOfferingByBackendID(ctx, backendID)
	cancel()
	if !errors.Is(err, waldur.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestWaldurBulkUsagePartialFailure(t *testing.T) {
	h := newWaldurHarness(t, nil)

	offering := h.createOffering("Usage Error Compute", h.providerAddr+"/usage-error", "VirtEngine.Compute", nil)
	goodResourceUUID := h.registerResource(offering.UUID, "", "alloc-good-001", string(waldur.ResourceStateOK), nil)

	ctx, cancel := h.context()
	responses, err := h.usage.SubmitBulkUsage(ctx, []*waldur.ResourceUsageReport{
		{
			ResourceUUID: goodResourceUUID,
			PeriodStart:  time.Now().Add(-1 * time.Hour).UTC(),
			PeriodEnd:    time.Now().UTC(),
			Components:   []waldur.ComponentUsage{{Type: "cpu_hours", Amount: 1}},
		},
		{
			ResourceUUID: "00000000-0000-0000-0000-000000000099",
			PeriodStart:  time.Now().Add(-1 * time.Hour).UTC(),
			PeriodEnd:    time.Now().UTC(),
			Components:   []waldur.ComponentUsage{{Type: "cpu_hours", Amount: 1}},
		},
	})
	cancel()
	if err == nil {
		t.Fatalf("expected bulk usage partial error")
	}
	if !strings.Contains(err.Error(), "00000000-0000-0000-0000-000000000099") {
		t.Fatalf("partial error did not mention failed resource: %v", err)
	}
	if len(responses) != 2 || responses[1].State != "failed" {
		t.Fatalf("unexpected bulk usage responses: %+v", responses)
	}
}

func TestWaldurCreateOrderFailureThenRecovery(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.AutoApproveOrders = false
		cfg.AutoProvision = false
	})

	offering := h.createOffering("Recovery Compute", h.providerAddr+"/recovery-compute", "VirtEngine.Compute", nil)

	h.mock.SetErrorState(&mocks.WaldurMockErrorState{
		FailCreateOrder: true,
		ErrorMessage:    "provider unavailable",
	})

	ctx, cancel := h.context()
	_, err := h.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID: offering.UUID,
		ProjectUUID:  h.projectUUID,
		Type:         "Create",
		Name:         "should-fail-first",
	})
	cancel()
	if !errors.Is(err, waldur.ErrServerError) {
		t.Fatalf("expected ErrServerError on first create order, got %v", err)
	}

	h.mock.ClearErrorState()

	ctx, cancel = h.context()
	order, err := h.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID: offering.UUID,
		ProjectUUID:  h.projectUUID,
		Type:         "Create",
		Name:         "should-succeed-second",
	})
	cancel()
	if err != nil {
		t.Fatalf("expected recovery create order success, got %v", err)
	}
	if order.UUID == "" {
		t.Fatalf("expected successful recovery order UUID")
	}
}
