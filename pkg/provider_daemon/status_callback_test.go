package provider_daemon

import (
	"testing"
)

func TestParseOrderStatusPayload(t *testing.T) {
	body := []byte(`{
		"order_uuid": "ord-1",
		"state": "executing",
		"attributes": {
			"order_id": "cust/1",
			"backend_id": "cust/1"
		}
	}`)

	payload, err := parseOrderStatusPayload(body)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if payload.OrderUUID != "ord-1" {
		t.Fatalf("unexpected order uuid: %s", payload.OrderUUID)
	}
	if payload.OrderID != "cust/1" {
		t.Fatalf("unexpected order id: %s", payload.OrderID)
	}
	if payload.State != "executing" {
		t.Fatalf("unexpected state: %s", payload.State)
	}
}

func TestMapWaldurOrderState(t *testing.T) {
	tests := map[string]string{
		"pending-consumer": "pending_payment",
		"pending-provider": "open",
		"executing":        "provisioning",
		"done":             "active",
		"terminating":      "pending_termination",
		"terminated":       "terminated",
		"erred":            "failed",
		"canceled":         "cancelled",
	}
	for input, expected := range tests {
		if got := mapWaldurOrderState(input); got != expected {
			t.Fatalf("state %s -> %s, want %s", input, got, expected)
		}
	}
}
