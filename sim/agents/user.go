package agents

import (
	"context"
	"math"

	"github.com/virtengine/virtengine/sim/model"
)

// User represents a network user generating demand.
type User struct {
	BaseAgent
	mean float64
	std  float64
}

// NewUser creates a new user agent.
func NewUser(id string, mean, std float64, base BaseAgent) *User {
	return &User{BaseAgent: base, mean: mean, std: std}
}

// Step produces demand events.
func (u *User) Step(ctx context.Context, state *model.State) ([]model.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	demand := u.mean + u.rng.NormFloat64()*u.std
	if demand < 0 {
		demand = math.Abs(demand)
	}

	action := model.MarketAction{
		ComputeDemand: demand,
		StorageDemand: demand * 0.6,
		GPUDemand:     demand * 0.25,
		GasDemand:     demand * 0.15,
	}

	return []model.Event{{
		Type:      model.EventDemand,
		Timestamp: state.Time,
		Data:      model.DemandEvent{Action: action},
	}}, nil
}
