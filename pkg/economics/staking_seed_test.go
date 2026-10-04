package economics

import (
	"math/big"
	"testing"
)

// TestDefaultTokenomicsStakingSeedIsSelfConsistent pins the invariant that the
// staking seed encodes the ratio the same struct literal declares.
//
// TotalStaked, StakingRatioBPS and TargetStakingRatioBPS were three independent
// literals that disagreed: the ratio fields said 6700 (67%) while TotalStaked
// encoded 33.5% of the supply. Consumers that derive the ratio from the supply
// rather than reading StakingRatioBPS -- which is what the simulation engine
// does, via Staked*10000/TokenSupply -- saw 33.5% regardless of what the ratio
// field declared. Nothing caught it because StakingRatioBPS is never read
// outside pkg/economics/analysis and pkg/economics/simulation.
//
// This test derives the ratio the same way that consumer does, so it fails if
// the two ever drift apart again, whatever values they hold.
func TestDefaultTokenomicsStakingSeedIsSelfConsistent(t *testing.T) {
	p := DefaultTokenomicsParams()

	if p.CurrentSupply == nil || p.CurrentSupply.Sign() == 0 {
		t.Fatalf("CurrentSupply must be set, got %v", p.CurrentSupply)
	}
	if p.TotalStaked == nil {
		t.Fatal("TotalStaked must be set")
	}

	// Derive the ratio the way sim/core.Engine.calculateStakingRatio does.
	derived := new(big.Int).Mul(p.TotalStaked, big.NewInt(10000))
	derived.Div(derived, p.CurrentSupply)
	got := derived.Int64()

	if got != p.StakingRatioBPS {
		t.Errorf("staking seed disagrees with the declared ratio: "+
			"TotalStaked*10000/CurrentSupply = %d bps, but StakingRatioBPS = %d bps "+
			"(off by %d bps). Fix the seed, not the ratio.",
			got, p.StakingRatioBPS, p.StakingRatioBPS-got)
	}
	if got != p.TargetStakingRatioBPS {
		t.Errorf("staking seed disagrees with the target ratio: derived %d bps, "+
			"TargetStakingRatioBPS = %d bps", got, p.TargetStakingRatioBPS)
	}
}

// TestDefaultTokenomicsStakingSeedIsNonZero guards the degenerate case: a zero
// seed makes Engine.calculateStakingRatio return 0, which then makes
// calculateAPR return 0, which silently disables the staking market.
func TestDefaultTokenomicsStakingSeedIsNonZero(t *testing.T) {
	p := DefaultTokenomicsParams()
	if p.TotalStaked.Sign() == 0 {
		t.Fatal("TotalStaked must be non-zero, else the staking ratio and APR both collapse to 0")
	}
}
