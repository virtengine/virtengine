# Load Testing Framework

## Overview

This load testing framework provides comprehensive tools for testing VirtEngine blockchain performance under various load conditions.

## Features

- **Configurable Load Profiles**: Constant, step, linear, and spike patterns
- **Multiple Scenarios**: VEID submission, order creation, bid submission, settlement
- **Detailed Metrics**: Latency percentiles (P50, P95, P99), throughput, error rates
- **Flexible Output**: JSON, CSV, and console reporting

## Usage

### Build

```bash
go build -o loadtest ./tests/load/cmd/loadtest
```

### Run a Load Test

```bash
./loadtest \
  --scenario veid_submit \
  --duration 5m \
  --target-rps 1000 \
  --endpoint localhost:9090 \
  --workers 100 \
  --output results.json
```

### Available Scenarios

- `veid_submit`: VEID scope submission
- `order_create`: Marketplace order creation (TODO)
- `bid_submit`: Provider bid submission (TODO)
- `settlement`: Settlement flow execution (TODO)

### Load Profile Options

- `--duration`: Test duration (e.g., 5m, 1h)
- `--target-rps`: Target requests per second
- `--workers`: Number of concurrent workers

### Output Formats

- `results.json`: JSON format
- `results.csv`: CSV format
- Console: Human-readable summary (always shown)

## Architecture

### Framework Components

1. **framework.go**: Core load test engine with worker pool
2. **metrics.go**: Metrics collection and aggregation
3. **report.go**: Report generation and formatting

### Adding New Scenarios

Implement the `Scenario` interface:

```go
type Scenario interface {
    Name() string
    Setup(ctx context.Context) error
    Execute(ctx context.Context) (*ExecutionResult, error)
    Teardown(ctx context.Context) error
}
```

## Performance Targets

| Metric | Target |
|--------|--------|
| VEID Submit | > 1000 TPS |
| Order Create | > 500 TPS |
| P99 Latency | < 500ms |

## CI/CD Integration

Load tests run nightly via GitHub Actions. See `.github/workflows/load-test.yaml`.
