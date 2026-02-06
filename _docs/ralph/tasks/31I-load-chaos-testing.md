# Task 31I: Load & Chaos Testing Framework

**vibe-kanban ID:** `2b34be86-26e8-48d9-9bb7-2a73dcdae4de`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31I |
| **Title** | test(infra): Load and chaos testing framework |
| **Priority** | P1 |
| **Wave** | 3 |
| **Estimated LOC** | 3000 |
| **Duration** | 3-4 weeks |
| **Dependencies** | None |
| **Blocking** | 31J (Metrics Dashboards) |

---

## Problem Statement

Production readiness requires understanding system behavior under:
- High traffic load (transaction throughput limits)
- Component failures (network partitions, service crashes)
- Resource exhaustion (memory, CPU, disk)
- Cascading failures (provider daemon → escrow → settlements)

No load testing or chaos engineering framework exists today.

### Current State Analysis

```
tests/load/                     ❌ Does not exist
tests/chaos/                    ❌ Does not exist
infra/chaos-mesh/               ❌ No chaos tooling
Benchmark tests:                ⚠️  Limited keeper benchmarks only
```

---

## Acceptance Criteria

### AC-1: Load Testing Framework
- [ ] Transaction load generator (gRPC/REST)
- [ ] Configurable scenarios (VEID submit, orders, bids, settlements)
- [ ] Ramping load profiles (step, linear, spike)
- [ ] Result collection and reporting
- [ ] Integration with CI/CD (nightly runs)

### AC-2: Chaos Testing with Chaos Mesh
- [ ] Chaos Mesh deployment configuration
- [ ] Network chaos (partition, latency, packet loss)
- [ ] Pod chaos (kill, failure injection)
- [ ] Stress chaos (CPU, memory pressure)
- [ ] Scheduled chaos experiments

### AC-3: Resilience Scenarios
- [ ] Validator node failure (⅓ down)
- [ ] Provider daemon crash recovery
- [ ] Database connection failures
- [ ] External API timeouts (ML inference, Waldur)
- [ ] Cascading failure tests

### AC-4: Reporting & CI Integration
- [ ] Test result dashboards
- [ ] Regression detection
- [ ] Alerting on performance degradation
- [ ] GitHub Actions integration
- [ ] Nightly test execution

---

## Technical Requirements

### Load Test Framework

```go
// tests/load/framework/framework.go

package framework

import (
    "context"
    "sync"
    "time"
)

type LoadProfile struct {
    Type       ProfileType
    Duration   time.Duration
    StartRate  float64  // requests per second
    EndRate    float64
    StepSize   float64
    StepTime   time.Duration
}

type ProfileType string

const (
    ProfileConstant ProfileType = "constant"
    ProfileStep     ProfileType = "step"
    ProfileLinear   ProfileType = "linear"
    ProfileSpike    ProfileType = "spike"
)

type Scenario interface {
    Name() string
    Setup(ctx context.Context) error
    Execute(ctx context.Context) (*ExecutionResult, error)
    Teardown(ctx context.Context) error
}

type ExecutionResult struct {
    Success    bool
    Duration   time.Duration
    StatusCode int
    Error      error
    Metadata   map[string]interface{}
}

type LoadTest struct {
    name      string
    scenario  Scenario
    profile   LoadProfile
    metrics   *MetricsCollector
    workers   int
}

func NewLoadTest(name string, scenario Scenario, profile LoadProfile) *LoadTest {
    return &LoadTest{
        name:     name,
        scenario: scenario,
        profile:  profile,
        metrics:  NewMetricsCollector(),
        workers:  100,
    }
}

func (lt *LoadTest) Run(ctx context.Context) (*TestReport, error) {
    // Setup
    if err := lt.scenario.Setup(ctx); err != nil {
        return nil, fmt.Errorf("setup failed: %w", err)
    }
    defer lt.scenario.Teardown(ctx)

    // Create worker pool
    jobs := make(chan struct{}, lt.workers*10)
    results := make(chan *ExecutionResult, lt.workers*10)

    var wg sync.WaitGroup
    for i := 0; i < lt.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for range jobs {
                result, err := lt.scenario.Execute(ctx)
                if err != nil {
                    result = &ExecutionResult{Success: false, Error: err}
                }
                results <- result
            }
        }()
    }

    // Result collector
    go func() {
        for result := range results {
            lt.metrics.Record(result)
        }
    }()

    // Load driver
    startTime := time.Now()
    ticker := lt.createTicker()
    
    for {
        select {
        case <-ctx.Done():
            return lt.generateReport(startTime), ctx.Err()
        case <-ticker.C:
            if time.Since(startTime) > lt.profile.Duration {
                close(jobs)
                wg.Wait()
                close(results)
                return lt.generateReport(startTime), nil
            }
            jobs <- struct{}{}
        }
    }
}

type TestReport struct {
    Name           string
    Duration       time.Duration
    TotalRequests  int64
    SuccessCount   int64
    FailureCount   int64
    AvgLatency     time.Duration
    P50Latency     time.Duration
    P95Latency     time.Duration
    P99Latency     time.Duration
    MaxLatency     time.Duration
    RequestsPerSec float64
    ErrorRate      float64
    Errors         map[string]int
}
```

### VEID Load Scenario

```go
// tests/load/scenarios/veid_submit.go

package scenarios

import (
    "context"
    "crypto/rand"
    "time"
    
    "github.com/virtengine/virtengine/tests/load/framework"
    veidv1 "github.com/virtengine/virtengine/x/veid/types/v1"
)

type VEIDSubmitScenario struct {
    client    veidv1.MsgClient
    accounts  []string
    scopeType veidv1.ScopeType
}

func NewVEIDSubmitScenario(grpcEndpoint string, accounts []string) *VEIDSubmitScenario {
    return &VEIDSubmitScenario{
        accounts:  accounts,
        scopeType: veidv1.ScopeTypeFace,
    }
}

func (s *VEIDSubmitScenario) Name() string {
    return "veid_submit"
}

func (s *VEIDSubmitScenario) Setup(ctx context.Context) error {
    conn, err := grpc.DialContext(ctx, s.grpcEndpoint, grpc.WithInsecure())
    if err != nil {
        return err
    }
    s.client = veidv1.NewMsgClient(conn)
    return nil
}

func (s *VEIDSubmitScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
    start := time.Now()
    
    // Generate random encrypted scope data
    scopeData := make([]byte, 32*1024) // 32KB typical scope
    rand.Read(scopeData)
    
    // Pick random account
    account := s.accounts[rand.Intn(len(s.accounts))]
    
    msg := &veidv1.MsgSubmitScope{
        Signer:        account,
        ScopeType:     s.scopeType,
        EncryptedData: scopeData,
        ClientSig:     generateSignature(scopeData),
    }
    
    resp, err := s.client.SubmitScope(ctx, msg)
    
    result := &framework.ExecutionResult{
        Duration: time.Since(start),
        Metadata: map[string]interface{}{
            "account": account,
        },
    }
    
    if err != nil {
        result.Success = false
        result.Error = err
        return result, nil
    }
    
    result.Success = true
    result.Metadata["scope_id"] = resp.ScopeId
    return result, nil
}

func (s *VEIDSubmitScenario) Teardown(ctx context.Context) error {
    return nil
}
```

### Chaos Mesh Configuration

```yaml
# infra/chaos-mesh/network-partition.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: validator-partition
  namespace: chaos-testing
spec:
  action: partition
  mode: all
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: virtengine-validator
  direction: both
  target:
    mode: fixed
    value: "1"
  duration: "60s"
  scheduler:
    cron: "@every 6h"

---
# infra/chaos-mesh/pod-failure.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: provider-daemon-kill
  namespace: chaos-testing
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: provider-daemon
  scheduler:
    cron: "@every 4h"

---
# infra/chaos-mesh/stress-test.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: memory-pressure
  namespace: chaos-testing
spec:
  mode: one
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: virtengine-node
  stressors:
    memory:
      workers: 4
      size: '256MB'
  duration: "5m"
  scheduler:
    cron: "@every 8h"

---
# infra/chaos-mesh/latency-injection.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: ml-inference-latency
  namespace: chaos-testing
spec:
  action: delay
  mode: all
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: ml-scorer
  delay:
    latency: "500ms"
    jitter: "100ms"
    correlation: "25"
  duration: "30m"
```

### Chaos Test Runner

```go
// tests/chaos/runner.go

package chaos

import (
    "context"
    "fmt"
    "time"
    
    "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosRunner struct {
    kubeClient client.Client
    namespace  string
    validators []HealthChecker
}

type HealthChecker interface {
    Name() string
    Check(ctx context.Context) error
}

type ChaosExperiment struct {
    Name        string
    ChaosObject client.Object
    Duration    time.Duration
    
    // Expected behavior
    MaxDowntime       time.Duration
    ExpectedRecovery  time.Duration
    AllowDataLoss     bool
    
    // Validation
    PreConditions  []Condition
    PostConditions []Condition
}

type Condition func(ctx context.Context) error

func (r *ChaosRunner) RunExperiment(ctx context.Context, exp ChaosExperiment) (*ExperimentResult, error) {
    result := &ExperimentResult{
        Name:      exp.Name,
        StartTime: time.Now(),
    }
    
    // Validate pre-conditions
    for _, cond := range exp.PreConditions {
        if err := cond(ctx); err != nil {
            return nil, fmt.Errorf("pre-condition failed: %w", err)
        }
    }
    
    // Create chaos object
    if err := r.kubeClient.Create(ctx, exp.ChaosObject); err != nil {
        return nil, fmt.Errorf("create chaos: %w", err)
    }
    defer r.cleanup(ctx, exp.ChaosObject)
    
    // Monitor health during experiment
    healthTicker := time.NewTicker(5 * time.Second)
    defer healthTicker.Stop()
    
    var downtimeStart time.Time
    var totalDowntime time.Duration
    
    timeout := time.After(exp.Duration + exp.MaxDowntime + exp.ExpectedRecovery)
    
    for {
        select {
        case <-timeout:
            result.EndTime = time.Now()
            result.TotalDowntime = totalDowntime
            
            // Validate post-conditions
            for _, cond := range exp.PostConditions {
                if err := cond(ctx); err != nil {
                    result.PostConditionsFailed = append(result.PostConditionsFailed, err.Error())
                }
            }
            
            result.Success = len(result.PostConditionsFailed) == 0 &&
                totalDowntime <= exp.MaxDowntime
            
            return result, nil
            
        case <-healthTicker.C:
            healthy := true
            for _, checker := range r.validators {
                if err := checker.Check(ctx); err != nil {
                    healthy = false
                    result.HealthErrors = append(result.HealthErrors, HealthError{
                        Time:    time.Now(),
                        Checker: checker.Name(),
                        Error:   err.Error(),
                    })
                    break
                }
            }
            
            if !healthy && downtimeStart.IsZero() {
                downtimeStart = time.Now()
            } else if healthy && !downtimeStart.IsZero() {
                totalDowntime += time.Since(downtimeStart)
                result.RecoveryTime = time.Since(downtimeStart)
                downtimeStart = time.Time{}
            }
        }
    }
}

func (r *ChaosRunner) cleanup(ctx context.Context, obj client.Object) {
    if err := r.kubeClient.Delete(ctx, obj); err != nil {
        fmt.Printf("Failed to cleanup chaos object: %v\n", err)
    }
}

type ExperimentResult struct {
    Name                  string
    StartTime             time.Time
    EndTime               time.Time
    Success               bool
    TotalDowntime         time.Duration
    RecoveryTime          time.Duration
    HealthErrors          []HealthError
    PostConditionsFailed  []string
}

type HealthError struct {
    Time    time.Time
    Checker string
    Error   string
}
```

### CI Integration

```yaml
# .github/workflows/load-test.yaml
name: Load Test

on:
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM UTC
  workflow_dispatch:
    inputs:
      duration:
        description: 'Test duration in minutes'
        required: false
        default: '30'
      target_rps:
        description: 'Target requests per second'
        required: false
        default: '1000'

jobs:
  load-test:
    runs-on: ubuntu-latest
    environment: load-test
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build load test binary
        run: go build -o loadtest ./tests/load/cmd/loadtest
      
      - name: Deploy test environment
        run: ./scripts/deploy-test-env.sh
        env:
          TESTNET_ENDPOINT: ${{ secrets.TESTNET_ENDPOINT }}
      
      - name: Run VEID load test
        run: |
          ./loadtest \
            --scenario veid_submit \
            --duration ${{ github.event.inputs.duration || '30' }}m \
            --target-rps ${{ github.event.inputs.target_rps || '1000' }} \
            --endpoint ${{ secrets.TESTNET_GRPC_ENDPOINT }} \
            --output results/veid-load.json
      
      - name: Run Order load test
        run: |
          ./loadtest \
            --scenario order_create \
            --duration ${{ github.event.inputs.duration || '30' }}m \
            --target-rps ${{ github.event.inputs.target_rps || '500' }} \
            --endpoint ${{ secrets.TESTNET_GRPC_ENDPOINT }} \
            --output results/order-load.json
      
      - name: Publish results
        run: |
          ./scripts/publish-results.sh results/
        env:
          GRAFANA_API_KEY: ${{ secrets.GRAFANA_API_KEY }}
      
      - name: Check for regressions
        run: |
          ./loadtest analyze \
            --baseline results/baseline.json \
            --current results/veid-load.json \
            --threshold 10

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: load-test-results
          path: results/

---
# .github/workflows/chaos-test.yaml
name: Chaos Test

on:
  schedule:
    - cron: '0 3 * * 0'  # Weekly on Sunday at 3 AM UTC
  workflow_dispatch:

jobs:
  chaos-test:
    runs-on: ubuntu-latest
    environment: chaos-test
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup kubectl
        uses: azure/setup-kubectl@v3
      
      - name: Configure kubeconfig
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" > /tmp/kubeconfig
          export KUBECONFIG=/tmp/kubeconfig
      
      - name: Install Chaos Mesh
        run: |
          helm repo add chaos-mesh https://charts.chaos-mesh.org
          helm install chaos-mesh chaos-mesh/chaos-mesh \
            --namespace chaos-testing --create-namespace
      
      - name: Run validator partition test
        run: |
          kubectl apply -f infra/chaos-mesh/network-partition.yaml
          sleep 120
          go test -v ./tests/chaos/... -run TestValidatorPartition
      
      - name: Run provider daemon kill test
        run: |
          kubectl apply -f infra/chaos-mesh/pod-failure.yaml
          sleep 60
          go test -v ./tests/chaos/... -run TestProviderDaemonRecovery
      
      - name: Cleanup
        if: always()
        run: |
          kubectl delete -f infra/chaos-mesh/ --ignore-not-found
```

---

## Directory Structure

```
tests/
├── load/
│   ├── cmd/
│   │   └── loadtest/
│   │       └── main.go           # CLI tool
│   ├── framework/
│   │   ├── framework.go          # Core framework
│   │   ├── metrics.go            # Metrics collection
│   │   └── report.go             # Report generation
│   └── scenarios/
│       ├── veid_submit.go
│       ├── order_create.go
│       ├── bid_submit.go
│       └── settlement.go
├── chaos/
│   ├── runner.go                 # Chaos experiment runner
│   ├── validators.go             # Health checkers
│   └── experiments/
│       ├── validator_partition_test.go
│       ├── provider_failure_test.go
│       └── ml_timeout_test.go
└── benchmarks/                   # Existing benchmark tests

infra/chaos-mesh/
├── network-partition.yaml
├── pod-failure.yaml
├── stress-test.yaml
└── latency-injection.yaml
```

---

## Testing Requirements

### Unit Tests
- Load profile calculation
- Metrics aggregation
- Report generation

### Integration Tests
- Local load test execution
- Chaos Mesh integration (Kind cluster)

### Baseline Tests
- Establish performance baselines
- Document expected throughput

---

## Success Metrics

| Metric | Target |
|--------|--------|
| VEID Submit throughput | > 1000 TPS |
| Order Create throughput | > 500 TPS |
| P99 latency under load | < 500ms |
| Recovery after pod kill | < 30s |
| Recovery after partition | < 60s after heal |
| Zero data loss in chaos | Yes |
