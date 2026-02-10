package agents

import (
	"context"
	"math/big"

	"github.com/virtengine/virtengine/sim/model"
)

// Validator represents a staking validator.
type Validator struct {
	BaseAgent
	stake        *big.Int
	uptimeTarget int64
}

// NewValidator creates a validator agent.
func NewValidator(id string, stake *big.Int, base BaseAgent) *Validator {
	return &Validator{
		BaseAgent:    base,
		stake:        new(big.Int).Set(stake),
		uptimeTarget: 9500,
	}
}

// Step adjusts stake and emits slashing events when uptime falls.
func (v *Validator) Step(ctx context.Context, state *model.State) ([]model.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	events := make([]model.Event, 0, 2)

	// Random staking adjustments.
	if v.rng.Float64() < 0.05 {
		amount := new(big.Int).Div(v.stake, big.NewInt(100))
		if amount.Sign() > 0 {
			events = append(events, model.Event{
				Type:      model.EventStake,
				Timestamp: state.Time,
				Data:      model.StakeEvent{Amount: amount},
			})
		}
	}

	// Simulate uptime slashing risk.
	uptime := int64(v.rng.Intn(10000))
	if uptime < v.uptimeTarget && v.rng.Float64() < 0.02 {
		slashAmount := new(big.Int).Div(v.stake, big.NewInt(200))
		events = append(events, model.Event{
			Type:      model.EventSlash,
			Timestamp: state.Time,
			Data:      model.SlashEvent{Amount: slashAmount, Reason: "downtime"},
		})
	}

	return events, nil
}
