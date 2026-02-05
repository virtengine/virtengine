package provider_daemon

import "testing"

func TestParseOrderRoutingRequest(t *testing.T) {
	data := map[string]interface{}{
		"order_id":         "customer1/1",
		"customer_address": "customer1",
		"offering_id":      "provider1/42",
		"provider_address": "provider1",
		"region":           "us-east-1",
		"quantity":         float64(2),
		"max_bid_price":    float64(50000),
	}

	req, ok := parseOrderRoutingRequest(data, "evt-1", 10)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if req.OrderID != "customer1/1" {
		t.Fatalf("order id = %s", req.OrderID)
	}
	if req.OfferingID != "provider1/42" {
		t.Fatalf("offering id = %s", req.OfferingID)
	}
	if req.ProviderAddress != "provider1" {
		t.Fatalf("provider = %s", req.ProviderAddress)
	}
	if req.Region != "us-east-1" {
		t.Fatalf("region = %s", req.Region)
	}
	if req.Quantity != 2 {
		t.Fatalf("quantity = %d", req.Quantity)
	}
	if req.MaxBidPrice != 50000 {
		t.Fatalf("max bid price = %d", req.MaxBidPrice)
	}
}
