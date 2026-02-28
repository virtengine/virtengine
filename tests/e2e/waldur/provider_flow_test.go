//go:build e2e.integration

package waldur

import (
	"testing"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

func TestWaldurProviderOfferingLifecycle(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.AutoApproveOrders = false
		cfg.AutoProvision = false
	})

	offering := h.createOffering("Provider Managed Compute", h.providerAddr+"/provider-compute", "VirtEngine.Compute", []waldur.PricingComponent{
		{Type: "usage", Name: "cpu_hours", MeasuredUnit: "hour", BillingType: "usage", Price: "0.15"},
	})

	ctx, cancel := h.context()
	updated, err := h.marketplace.UpdateOffering(ctx, offering.UUID, waldur.UpdateOfferingRequest{
		Description: "Updated provider lifecycle description",
		Attributes: map[string]interface{}{
			"region": "ap-southeast-2",
		},
	})
	cancel()
	if err != nil {
		t.Fatalf("update offering: %v", err)
	}
	if updated.Description != "Updated provider lifecycle description" {
		t.Fatalf("updated description = %q", updated.Description)
	}

	ctx, cancel = h.context()
	found, err := h.marketplace.GetOfferingByBackendID(ctx, h.providerAddr+"/provider-compute")
	cancel()
	if err != nil {
		t.Fatalf("find updated offering by backend ID: %v", err)
	}
	if found.UUID != offering.UUID {
		t.Fatalf("found offering UUID = %s, want %s", found.UUID, offering.UUID)
	}
}

func TestWaldurProviderOrderFulfillment(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.AutoApproveOrders = false
		cfg.AutoProvision = false
	})

	backendID := h.providerAddr + "/provider-order-compute"
	offering := h.createOffering("Provider Order Compute", backendID, "VirtEngine.Compute", nil)

	order := h.createOrder(offering.UUID, "chain-order-001", map[string]interface{}{
		"ve_order_id":         "chain-order-001",
		"ve_customer_address": "ve1customerrouter",
		"ve_provider_address": h.providerAddr,
		"region":              "ap-southeast-2",
	})

	ctx, cancel := h.context()
	if err := h.marketplace.ApproveOrderByProvider(ctx, order.UUID); err != nil {
		cancel()
		t.Fatalf("approve provider order: %v", err)
	}
	if err := h.marketplace.SetOrderBackendID(ctx, order.UUID, "chain-order-001"); err != nil {
		cancel()
		t.Fatalf("set provider backend ID: %v", err)
	}
	cancel()

	ctx, cancel = h.context()
	storedOrder, err := h.marketplace.GetOrder(ctx, order.UUID)
	cancel()
	if err != nil {
		t.Fatalf("get fulfilled order: %v", err)
	}
	if storedOrder.State != "approved" {
		t.Fatalf("fulfilled order state = %s, want approved", storedOrder.State)
	}
	if h.mock.GetOrder(order.UUID).BackendID != "chain-order-001" {
		t.Fatalf("backend ID not persisted on fulfilled order")
	}

	resourceUUID := h.registerResource(
		offering.UUID,
		order.UUID,
		"chain-order-001",
		string(waldur.ResourceStateOK),
		map[string]interface{}{"external_ip": "203.0.113.30"},
	)
	ctx, cancel = h.context()
	resource, err := h.marketplace.GetResource(ctx, resourceUUID)
	cancel()
	if err != nil {
		t.Fatalf("get fulfilled resource: %v", err)
	}
	if resource.State != string(waldur.ResourceStateOK) {
		t.Fatalf("fulfilled resource state = %s, want %s", resource.State, waldur.ResourceStateOK)
	}
}
