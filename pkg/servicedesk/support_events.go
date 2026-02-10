package servicedesk

import (
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
)

// SupportEventEnvelope is a parsed support event from ABCI attributes.
type SupportEventEnvelope struct {
	EventType   string
	EventID     string
	BlockHeight int64
	Sequence    uint64
	PayloadJSON string
}

// ExtractSupportEvents parses ABCI events and returns support_event envelopes.
func ExtractSupportEvents(events []abci.Event) ([]SupportEventEnvelope, error) {
	envelopes := make([]SupportEventEnvelope, 0)
	for _, event := range events {
		if event.Type != "support_event" {
			continue
		}

		env := SupportEventEnvelope{}
		for _, attr := range event.Attributes {
			key := string(attr.Key)
			value := string(attr.Value)
			switch key {
			case "event_type":
				env.EventType = value
			case "event_id":
				env.EventID = value
			case "block_height":
				height, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid block_height: %w", err)
				}
				env.BlockHeight = height
			case "sequence":
				seq, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid sequence: %w", err)
				}
				env.Sequence = seq
			case "payload_json":
				env.PayloadJSON = value
			}
		}

		if env.EventType == "" || env.EventID == "" || env.Sequence == 0 {
			continue
		}

		envelopes = append(envelopes, env)
	}

	return envelopes, nil
}
