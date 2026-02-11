# Economics Simulation Suite

## Purpose
Run deterministic economics simulations, export dashboards, and enforce regression gates for gas pricing and congestion policy changes.

## When to Use
- After changing gas pricing, congestion logic, or fee market parameters.
- Before releases that touch economics, mempool, or simulation code.
- During incident retrospectives to validate assumptions.

## Commands

### Run the full suite
```bash
go run ./cmd/ve-sim suite --scenario baseline --output-dir ./sim-output
```

Outputs:
- `simulation.json` (full run state)
- `metrics.json` (summary metrics for CI/regression checks)
- `monte-carlo.json` + `monte-carlo.csv`
- `sensitivity.json`
- `dashboard.html` (static dashboard)

### Validate metrics thresholds
```bash
go run ./cmd/ve-sim check --metrics ./sim-output/metrics.json
```

### Serve interactive dashboard
```bash
go run ./cmd/ve-sim dashboard \
  --monte-carlo ./sim-output/monte-carlo.json \
  --sensitivity ./sim-output/sensitivity.json
```

## Operational Guidance
- Always run `suite` with the same scenario before/after changes to ensure deterministic comparisons.
- If metrics regress, identify the parameter causing the shift (use `sensitivity`).
- Archive `metrics.json` and `dashboard.html` with the incident report or PR.

## Notes
- The suite uses deterministic seeds and fixed dates to keep outputs stable across runs.
- CI uses `metrics.json` to enforce regression gates.

