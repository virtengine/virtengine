package agents

import (
	"context"
	"math/rand" //nolint:gosec // G404: simulation agents use weak random for reproducibility, not security

	"github.com/virtengine/virtengine/sim/model"
)

// AgentType identifies the agent class.
type AgentType string

const (
	UserAgent      AgentType = "user"
	ProviderAgent  AgentType = "provider"
	ValidatorAgent AgentType = "validator"
	AttackerAgent  AgentType = "attacker"
)

// Agent performs actions during simulation steps.
type Agent interface {
	ID() string
	Type() AgentType
	Step(ctx context.Context, state *model.State) ([]model.Event, error)
}

// BaseAgent provides shared fields.
type BaseAgent struct {
	id   string
	kind AgentType
	rng  *rand.Rand
}

// NewBaseAgent constructs a base agent.
func NewBaseAgent(id string, kind AgentType, rng *rand.Rand) BaseAgent {
	return BaseAgent{id: id, kind: kind, rng: rng}
}

// ID returns agent id.
func (a *BaseAgent) ID() string {
	return a.id
}

// Type returns agent type.
func (a *BaseAgent) Type() AgentType {
	return a.kind
}
