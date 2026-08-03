package privileged

import "fmt"

const ContractVersion = "virtengine.roles.privileged/v1"

// FeatureState is ordered so dependency-incomplete builds can enforce a hard cap.
type FeatureState uint8

const (
	FeatureDisabled FeatureState = iota
	FeatureFixtureOnly
	FeatureSandbox
	FeatureProduction
)

func (s FeatureState) String() string {
	switch s {
	case FeatureDisabled:
		return "disabled"
	case FeatureFixtureOnly:
		return "fixture_only"
	case FeatureSandbox:
		return "sandbox"
	case FeatureProduction:
		return "production"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// FeatureGate makes every externally observable activation decision explicit.
type FeatureGate struct {
	State         FeatureState `json:"state"`
	Registration  bool         `json:"registration"`
	Advertisement bool         `json:"advertisement"`
	Readiness     bool         `json:"readiness"`
	Mutation      bool         `json:"mutation"`
	Blockers      []string     `json:"blockers"`
}

func (g FeatureGate) Validate() error {
	if g.State > FeatureFixtureOnly {
		return fmt.Errorf("feature state %s exceeds fixture-only cap", g.State)
	}
	if g.Registration || g.Advertisement || g.Readiness || g.Mutation {
		return fmt.Errorf("fixture-only privileged governance cannot register, advertise, become ready, or mutate")
	}
	if len(g.Blockers) == 0 {
		return fmt.Errorf("at least one activation blocker is required")
	}
	for _, blocker := range g.Blockers {
		if blocker == "" {
			return fmt.Errorf("activation blockers must be non-empty")
		}
	}
	return nil
}
