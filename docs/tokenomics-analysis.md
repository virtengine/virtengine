# VirtEngine Tokenomics Model Validation

## Current supply policy

Initial supply is **0 tokens**. New tokens are issued only when unique human identities are verified, under the identity-allocation rules. There is **no fixed maximum supply and no hard cap**: issuance follows verified human participation and can continue as new identities are verified and the human population grows. Human population is the basis for issuance, not a fixed numerical token ceiling.

The initial zero balance also applies to the Foundation-controlled genesis account; any Foundation allocation is issued through eligible identity-linked issuance, not a premine. Staking does not independently authorize new token issuance.

## ECON-001 Implementation Summary

This document describes the tokenomics validation framework implemented in `pkg/economics/`.

## Overview

The framework analyzes economic scenarios. Its existing inflation-based simulations use legacy assumptions and do not validate the current zero-initial-supply, uncapped identity-linked policy. This package provides:

1. **Economic Simulation Models** - Monte Carlo and deterministic simulations
2. **Inflation/Deflation Dynamics** - Supply and monetary policy analysis
3. **Staking Reward Optimization** - Validator incentive modeling
4. **Fee Market Analysis** - Transaction fee dynamics and spam resistance
5. **Token Distribution Fairness** - Gini coefficient, Lorenz curves, Nakamoto coefficient
6. **Attack Cost Analysis** - 51% attacks, spam, long-range attacks, cartels
7. **Game Theory Analysis** - Incentive alignment for validators, delegators, verifiers
8. **Economic Security Audit** - Comprehensive audit framework

## Package Structure

```
pkg/economics/
├── doc.go              # Package documentation
├── types.go            # Core types and parameters
├── simulation/
│   ├── doc.go          # Simulation package docs
│   ├── inflation.go    # Inflation dynamics simulation
│   ├── staking.go      # Staking reward simulation
│   ├── fee_market.go   # Fee market simulation
│   └── *_test.go       # Unit tests
├── analysis/
│   ├── doc.go          # Analysis package docs
│   ├── distribution.go # Token distribution analysis
│   ├── attack_cost.go  # Attack cost analysis
│   ├── game_theory.go  # Game theory analysis
│   └── *_test.go       # Unit tests
└── audit/
    ├── doc.go          # Audit package docs
    ├── auditor.go      # Economic security auditor
    └── *_test.go       # Unit tests
```

## Economic Parameters

### Supply policy

| Parameter | Policy | Description |
|-----------|--------|-------------|
| Initial Supply | 0 tokens | Tokens are issued only when unique human identities are verified |
| Maximum Supply | No cap | Supply follows verified human participation and can grow with new verified identities and the human population |

### Legacy simulation parameters

The following values describe existing simulation inputs, not current issuance policy. `economics.DefaultTokenomicsParams()` still uses legacy nonzero initial supply and a finite maximum; code and simulations require separate alignment and validation before they can substantiate the current policy. Inflation targets and block rewards below do not authorize minting independently of identity verification.

| Parameter | Default Value | Description |
|-----------|---------------|-------------|
| Target Inflation | 7% | Annual target inflation rate |
| Min Inflation | 1% | Minimum inflation floor |
| Max Inflation | 20% | Maximum inflation ceiling |
| Target Staking Ratio | 67% | Target percentage staked |
| Unbonding Period | 21 days | Stake unbonding duration |
| Base Reward Per Block | 1 token | Base block producer reward |
| Default Take Rate | 4% | Protocol fee on payments |
| VEID Reward Pool | 10,000 tokens | Identity verification rewards per epoch |

### Legacy Inflation Adjustment Mechanism

The inflation rate adjusts dynamically based on staking ratio:

```
If StakingRatio < TargetStakingRatio:
    Inflation increases → Higher APR → Attracts more staking

If StakingRatio > TargetStakingRatio:
    Inflation decreases → Lower APR → Reduces staking pressure
```

This describes the legacy staking-driven simulation, not the current issuance rule. New token issuance must remain linked to verified human identities; staking demand alone does not trigger minting.

## Simulation Capabilities

### 1. Inflation Simulation

```go
sim := simulation.NewInflationSimulator(params)
result := sim.SimulateYear(networkState)

// Result includes:
// - Monthly snapshots of supply, staking ratio, inflation
// - Projected APR for stakers
// - Sustainability score (0-100)
// - Recommendations for parameter adjustment
```

### 2. Staking Reward Simulation

```go
sim := simulation.NewStakingSimulator(params)
rewards := sim.SimulateRewardDistribution(validators, epochBlocks)

// Result includes:
// - Per-validator reward breakdown
// - Block proposal, VEID verification, uptime rewards
// - Commission splits
// - Effective APR per validator
```

### 3. Fee Market Simulation

```go
sim := simulation.NewFeeMarketSimulator(params)
result := sim.SimulateFeeMarket(state, transactions, takeRate)

// Result includes:
// - Protocol revenue
// - Validator revenue
// - Fee volatility metrics
// - Spam resistance score
```

## Analysis Capabilities

### 1. Distribution Analysis

Calculates key decentralization metrics:

- **Gini Coefficient**: Wealth inequality (0=equal, 1=unequal)
- **Lorenz Curve**: Cumulative wealth distribution
- **Nakamoto Coefficient**: Minimum entities for 51% control
- **Herfindahl-Hirschman Index**: Market concentration

### 2. Attack Cost Analysis

Analyzes costs for various attack vectors:

| Attack Type | Key Metrics |
|-------------|-------------|
| 51% Attack | USD cost, tokens required, time to prepare |
| Spam Attack | Cost per hour, detection difficulty |
| Long-Range Attack | Unbonding period vulnerability |
| Cartel Formation | Minimum validators for collusion |
| Nothing-at-Stake | Slashing effectiveness |

### 3. Game Theory Analysis

Analyzes Nash equilibria for:

- **Validator Incentives**: Honest vs malicious strategies
- **Delegator Incentives**: Delegation strategy optimization
- **VEID Verifier Incentives**: Quality vs speed tradeoffs
- **Provider Incentives**: Service quality incentives

## Economic Security Audit

The `EconomicAuditor` provides comprehensive security assessment:

```go
auditor := audit.NewEconomicAuditor(params)
result := auditor.PerformAudit(input)

// Result includes:
// - Overall security score (0-100)
// - Inflation analysis
// - Staking analysis
// - Fee market analysis
// - Distribution metrics
// - Attack vulnerability assessments
// - Game theory analyses
// - Identified vulnerabilities
// - Prioritized recommendations
```

### Audit Scoring

The overall score is calculated from:

- Inflation sustainability (-20 max penalty)
- Staking ratio health (-25 max penalty)
- Fee market efficiency (-15 max penalty)
- Token distribution fairness (-20 max penalty)
- Attack vulnerability (-20 max penalty)

### Vulnerability Severity Levels

| Level | Description |
|-------|-------------|
| Critical | Immediate action required, network at risk |
| High | Significant risk, should address soon |
| Medium | Moderate concern, plan for mitigation |
| Low | Minor issue, monitor over time |

## Recommendations

Based on analysis, the framework generates recommendations:

### Inflation Recommendations
- Adjust inflation curve if deviating from target
- Balance token value preservation with staking incentives

### Staking Recommendations
- Monitor staking ratio for security thresholds
- Implement validator caps if concentration is high
- Adjust unbonding period based on market conditions

### Fee Market Recommendations
- Set minimum gas price to deter spam
- Enable adaptive minimum gas pricing tied to congestion metrics
- Implement priority fee mechanisms for congestion
- Validate fee market changes with `ve-sim suite` exports
- Balance take rate for protocol sustainability

### Distribution Recommendations
- Encourage delegation to smaller validators
- Implement progressive rewards if Gini is high
- Monitor Nakamoto coefficient for centralization risk

## Usage Example

```go
package main

import (
    "github.com/virtengine/virtengine/pkg/economics"
    "github.com/virtengine/virtengine/pkg/economics/audit"
)

func main() {
    // Legacy simulation defaults; not the current identity-linked supply policy
    params := economics.DefaultTokenomicsParams()
    
    // Create auditor
    auditor := audit.NewEconomicAuditor(params)
    
    // Prepare input data
    input := audit.AuditInput{
        NetworkState: currentNetworkState,
        Validators:   activeValidators,
        Holdings:     tokenHoldings,
        HistoricalFees: recentFees,
        TokenPriceUSD:  1.50,
        SlashingEnabled: true,
        SlashingPenaltyBPS: 500,
    }
    
    // Perform audit
    result := auditor.PerformAudit(input)
    
    // Generate report
    report := auditor.GenerateAuditReport(result)
    
    // Process results...
}
```

## Test Coverage

All components include comprehensive unit tests:

```bash
go test -v ./pkg/economics/...
```

Test coverage includes:
- Inflation simulation scenarios
- Staking reward calculations
- Fee market dynamics
- Distribution metrics accuracy
- Attack cost calculations
- Game theory equilibrium finding
- Audit scoring and reporting

## Future Enhancements

Potential improvements for future iterations:

1. **Monte Carlo Simulations** - Add stochastic variations
2. **Historical Data Integration** - Analyze real network data
3. **Parameter Sensitivity Analysis** - Automated optimization
4. **Visualization Tools** - Charts and dashboards
5. **Real-time Monitoring** - Continuous economic health checks
6. **Machine Learning** - Predictive economic modeling

## Conclusion

The VirtEngine tokenomics framework provides robust tools for:

- Validating economic model assumptions
- Identifying potential vulnerabilities
- Optimizing incentive structures
- Ensuring long-term network sustainability

The modular design allows for easy extension and integration with the broader VirtEngine ecosystem.
