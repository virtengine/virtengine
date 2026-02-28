//go:build e2e.integration

package waldur

import (
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

func TestWaldurCustomerOnboardingAndOrderFlow(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.AutoApproveOrders = true
		cfg.AutoProvision = true
		cfg.ProvisionDelay = 25 * time.Millisecond
	})

	compute := h.createOffering("Standard Compute", h.providerAddr+"/compute-standard", "VirtEngine.Compute", nil)
	h.createOffering("GPU A100", h.providerAddr+"/gpu-a100", "VirtEngine.GPU", nil)
	h.createOffering("Block Storage", h.providerAddr+"/storage-block", "VirtEngine.Storage", nil)

	ctx, cancel := h.context()
	offerings, err := h.marketplace.ListOfferings(ctx, waldur.ListOfferingsParams{CustomerUUID: h.customerUUID})
	cancel()
	if err != nil {
		t.Fatalf("list offerings: %v", err)
	}
	if len(offerings) < 3 {
		t.Fatalf("expected at least 3 offerings, got %d", len(offerings))
	}

	order := h.createOrder(compute.UUID, "customer-auto-provision", map[string]interface{}{
		"allocation_id":    "alloc-customer-001",
		"customer_address": "ve1customerflow",
	})

	h.waitFor("auto-provisioned resource creation", 2*time.Second, func() bool {
		stored := h.mock.GetOrder(order.UUID)
		return stored != nil && stored.State == "done" && stored.ResourceUUID != ""
	})

	stored := h.mock.GetOrder(order.UUID)
	if stored == nil || stored.ResourceUUID == "" {
		t.Fatalf("expected auto-provisioned order resource UUID")
	}

	ctx, cancel = h.context()
	resource, err := h.marketplace.GetResource(ctx, stored.ResourceUUID)
	cancel()
	if err != nil {
		t.Fatalf("get provisioned resource: %v", err)
	}
	if resource.State == "" {
		t.Fatalf("expected provisioned resource state")
	}
	if resource.OfferingUUID != compute.UUID {
		t.Fatalf("resource offering UUID = %s, want %s", resource.OfferingUUID, compute.UUID)
	}
}
