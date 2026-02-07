package agents

import (
	"context"

	"github.com/virtengine/virtengine/sim/model"
)

// Provider represents a resource provider.
type Provider struct {
	BaseAgent
	capacity float64
}

// NewProvider creates a provider agent.
func NewProvider(id string, capacity float64, base BaseAgent) *Provider {
	return &Provider{BaseAgent: base, capacity: capacity}
}

// Step supplies resources and may exit under poor economics.
func (p *Provider) Step(ctx context.Context, state *model.State) ([]model.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	exitProbability := 0.0
	if state.Market.ComputePrice < 0.5*state.Market.StoragePrice {
		exitProbability = 0.02
	}
	if p.rng.Float64() < exitProbability {
		return []model.Event{{
			Type:      model.EventProviderExit,
			Timestamp: state.Time,
			Data:      model.ProviderExitEvent{ProviderID: p.id, Capacity: p.capacity},
		}}, nil
	}

	supply := model.MarketAction{
		ComputeSupply: p.capacity,
		StorageSupply: p.capacity * 0.8,
		GPUSupply:     p.capacity * 0.4,
	}

	return []model.Event{{
		Type:      model.EventSupply,
		Timestamp: state.Time,
		Data:      model.SupplyEvent{Action: supply},
	}}, nil
}
