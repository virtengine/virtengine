package agents

import (
	"context"
	"math"
	"math/big"

	"github.com/virtengine/virtengine/sim/model"
)

// Attacker represents an adversarial agent.
type Attacker struct {
	BaseAgent
	budgetUSD float64
}

// NewAttacker creates an attacker agent.
func NewAttacker(id string, budgetUSD float64, base BaseAgent) *Attacker {
	return &Attacker{BaseAgent: base, budgetUSD: budgetUSD}
}

// Step generates adversarial events.
func (a *Attacker) Step(ctx context.Context, state *model.State) ([]model.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	events := make([]model.Event, 0, 2)

	if a.rng.Float64() < 0.04 {
		cost := math.Min(a.budgetUSD*0.05, 25000)
		impact := cost * (1.2 + a.rng.Float64())
		events = append(events, model.Event{
			Type:      model.EventSybil,
			Timestamp: state.Time,
			Data: model.AttackEvent{
				AttackType: "sybil",
				CostUSD:    cost,
				ImpactUSD:  impact,
			},
		})
	}

	if a.rng.Float64() < 0.03 {
		impact := a.budgetUSD * 0.03
		events = append(events, model.Event{
			Type:      model.EventPriceManipulate,
			Timestamp: state.Time,
			Data: model.AttackEvent{
				AttackType: "manipulation",
				CostUSD:    impact * 0.7,
				ImpactUSD:  impact,
			},
		})
	}

	if a.rng.Float64() < 0.05 {
		fees := big.NewInt(int64(500 + a.rng.Intn(1500)))
		impact := float64(fees.Int64()) * state.TokenPriceUSD
		events = append(events, model.Event{
			Type:      model.EventMEV,
			Timestamp: state.Time,
			Data: model.MEVEvent{
				ExtractedFees: fees,
				ImpactUSD:     impact,
			},
		})
	}

	return events, nil
}
