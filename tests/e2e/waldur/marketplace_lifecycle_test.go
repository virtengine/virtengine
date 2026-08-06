//go:build e2e.integration

package waldur

import (
	"testing"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

func TestWaldurMarketplaceLifecycle(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.AutoApproveOrders = false
		cfg.AutoProvision = false
	})

	offering := h.createOffering(
		"VirtEngine Compute",
		h.providerAddr+"/compute-001",
		"VirtEngine.Compute",
		[]waldur.PricingComponent{
			{Type: "usage", Name: "cpu_hours", MeasuredUnit: "hour", BillingType: "usage", Price: "0.10"},
			{Type: "usage", Name: "ram_gb_hours", MeasuredUnit: "gb-hour", BillingType: "usage", Price: "0.05"},
		},
	)

	ctx, cancel := h.context()
	found, err := h.marketplace.GetOfferingByBackendID(ctx, h.providerAddr+"/compute-001")
	cancel()
	if err != nil {
		t.Fatalf("get offering by backend ID: %v", err)
	}
	if found.UUID != offering.UUID {
		t.Fatalf("unexpected offering UUID: got %s want %s", found.UUID, offering.UUID)
	}

	order := h.createOrder(offering.UUID, "allocation-lifecycle", map[string]interface{}{
		"allocation_id":    "alloc-lifecycle-001",
		"customer_address": "ve1customer123",
		"provider_address": h.providerAddr,
	})

	ctx, cancel = h.context()
	if err := h.marketplace.ApproveOrderByProvider(ctx, order.UUID); err != nil {
		cancel()
		t.Fatalf("approve order: %v", err)
	}
	if err := h.marketplace.SetOrderBackendID(ctx, order.UUID, "alloc-lifecycle-001"); err != nil {
		cancel()
		t.Fatalf("set order backend ID: %v", err)
	}
	cancel()

	resourceUUID := h.registerResource(
		offering.UUID,
		order.UUID,
		"alloc-lifecycle-001",
		string(waldur.ResourceStateOK),
		map[string]interface{}{
			"external_ip": "203.0.113.10",
		},
	)

	ctx, cancel = h.context()
	resource, err := h.marketplace.GetResource(ctx, resourceUUID)
	cancel()
	if err != nil {
		t.Fatalf("get resource: %v", err)
	}
	if resource.State != string(waldur.ResourceStateOK) {
		t.Fatalf("unexpected initial resource state: %s", resource.State)
	}

	for _, tc := range []struct {
		name       string
		action     func() error
		wantState  waldur.ResourceState
	}{
		{
			name: "stop",
			action: func() error {
				ctx, cancel := h.context()
				defer cancel()
				_, err := h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resourceUUID, IdempotencyKey: "stop-lifecycle"})
				return err
			},
			wantState: waldur.ResourceStateStopped,
		},
		{
			name: "start",
			action: func() error {
				ctx, cancel := h.context()
				defer cancel()
				_, err := h.lifecycle.Start(ctx, waldur.LifecycleRequest{ResourceUUID: resourceUUID, IdempotencyKey: "start-lifecycle"})
				return err
			},
			wantState: waldur.ResourceStateOK,
		},
		{
			name: "restart",
			action: func() error {
				ctx, cancel := h.context()
				defer cancel()
				_, err := h.lifecycle.Restart(ctx, waldur.LifecycleRequest{ResourceUUID: resourceUUID, IdempotencyKey: "restart-lifecycle"})
				return err
			},
			wantState: waldur.ResourceStateOK,
		},
		{
			name: "resize",
			action: func() error {
				ctx, cancel := h.context()
				defer cancel()
				cpu := 8
				mem := 16384
				_, err := h.lifecycle.Resize(ctx, waldur.ResizeRequest{
					LifecycleRequest: waldur.LifecycleRequest{ResourceUUID: resourceUUID, IdempotencyKey: "resize-lifecycle"},
					CPUCores:         &cpu,
					MemoryMB:         &mem,
				})
				return err
			},
			wantState: waldur.ResourceStateOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.action(); err != nil {
				t.Fatalf("%s resource: %v", tc.name, err)
			}
			ctx, cancel := h.context()
			state, err := h.lifecycle.GetResourceState(ctx, resourceUUID)
			cancel()
			if err != nil {
				t.Fatalf("get resource state after %s: %v", tc.name, err)
			}
			if state != tc.wantState {
				t.Fatalf("resource state after %s = %s, want %s", tc.name, state, tc.wantState)
			}
		})
	}

	ctx, cancel = h.context()
	resp, err := h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{
		ResourceUUID:   resourceUUID,
		IdempotencyKey: "terminate-lifecycle",
		Immediate:      true,
	})
	cancel()
	if err != nil {
		t.Fatalf("terminate resource: %v", err)
	}
	if resp.State == "" {
		t.Fatalf("terminate response missing state")
	}

	ctx, cancel = h.context()
	state, err := h.lifecycle.GetResourceState(ctx, resourceUUID)
	cancel()
	if err != nil {
		t.Fatalf("get final resource state: %v", err)
	}
	if state != waldur.ResourceStateTerminated {
		t.Fatalf("final resource state = %s, want %s", state, waldur.ResourceStateTerminated)
	}

	if err := waldur.ValidateLifecycleAction(waldur.ResourceStateTerminated, waldur.LifecycleActionStart); err == nil {
		t.Fatalf("expected terminated resource to reject start action")
	}
}
