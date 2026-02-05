package keeper

import (
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestParseOrderState(t *testing.T) {
	state := marketplace.ParseOrderState("active")
	if state != marketplace.OrderStateActive {
		t.Fatalf("unexpected state: %v", state)
	}
}

func TestApplyOrderStateTransition(t *testing.T) {
	now := time.Now().UTC()
	order := marketplace.NewOrder(
		marketplace.OrderID{CustomerAddress: "cust1", Sequence: 1},
		marketplace.OfferingID{ProviderAddress: "prov1", Sequence: 1},
		100,
		1,
	)
	if err := order.SetStateAt(marketplace.OrderStateOpen, "open", now); err != nil {
		t.Fatalf("set open: %v", err)
	}

	if err := applyOrderStateTransition(order, marketplace.OrderStateProvisioning, "provisioning", now); err != nil {
		t.Fatalf("apply transition: %v", err)
	}
	if order.State != marketplace.OrderStateProvisioning {
		t.Fatalf("state = %s, want provisioning", order.State.String())
	}
}
