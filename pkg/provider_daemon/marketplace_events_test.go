package provider_daemon

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
)

func TestExtractMarketplaceEvents(t *testing.T) {
	events := []abci.Event{
		{
			Type: "marketplace_event",
			Attributes: []abci.EventAttribute{
				{Key: "event_type", Value: "allocation_created"},
				{Key: "event_id", Value: "evt_alloc_1"},
				{Key: "block_height", Value: "42"},
				{Key: "sequence", Value: "7"},
				{Key: "payload_json", Value: `{"allocation_id":"a1"}`},
			},
		},
	}

	envelopes, err := ExtractMarketplaceEvents(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 envelope, got %d", len(envelopes))
	}

	env := envelopes[0]
	if env.EventType != "allocation_created" {
		t.Fatalf("unexpected event type: %s", env.EventType)
	}
	if env.EventID != "evt_alloc_1" {
		t.Fatalf("unexpected event id: %s", env.EventID)
	}
	if env.BlockHeight != 42 {
		t.Fatalf("unexpected block height: %d", env.BlockHeight)
	}
	if env.Sequence != 7 {
		t.Fatalf("unexpected sequence: %d", env.Sequence)
	}
	if env.PayloadJSON == "" {
		t.Fatalf("expected payload JSON")
	}
}

