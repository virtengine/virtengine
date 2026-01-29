# Reliability Testing Framework

## Overview

Reliability testing ensures VirtEngine services meet SLO targets under real-world conditions. This framework defines testing strategies, tools, and practices for validating reliability.

## Table of Contents
1. [Testing Philosophy](#testing-philosophy)
2. [Testing Types](#testing-types)
3. [Chaos Engineering](#chaos-engineering)
4. [Load and Stress Testing](#load-and-stress-testing)
5. [Failure Mode Testing](#failure-mode-testing)
6. [Disaster Recovery Testing](#disaster-recovery-testing)
7. [Testing Schedule](#testing-schedule)

---

## Testing Philosophy

### Principles

1. **Test in Production**: Staging can't replicate all production conditions
2. **Test Continuously**: Reliability testing is not a one-time event
3. **Test Realistically**: Simulate real failure scenarios
4. **Measure Impact**: Quantify reliability improvements
5. **Learn from Failures**: Every test is a learning opportunity

### Goals

- **Validate SLOs**: Confirm services meet reliability targets
- **Find Weaknesses**: Discover failure modes before users do
- **Build Confidence**: Know the system can handle failures
- **Improve Response**: Practice incident response
- **Prevent Regressions**: Catch reliability degradations early

---

## Testing Types

### 1. Synthetic Monitoring

**Purpose**: Continuously validate critical user journeys

**Implementation**:

```go
// pkg/sre/synthetic/monitor.go

package synthetic

import (
	"context"
	"time"
)

type Monitor struct {
	name     string
	interval time.Duration
	check    CheckFunc
}

type CheckFunc func(ctx context.Context) error

type Result struct {
	Timestamp time.Time
	Success   bool
	Latency   time.Duration
	Error     error
}

func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runCheck(ctx)
		}
	}
}

func (m *Monitor) runCheck(ctx context.Context) Result {
	start := time.Now()
	err := m.check(ctx)
	latency := time.Since(start)

	return Result{
		Timestamp: start,
		Success:   err == nil,
		Latency:   latency,
		Error:     err,
	}
}
```

**Checks**:

```go
// Transaction submission check
func checkTransactionSubmission(ctx context.Context) error {
	client := getClient()

	tx := createTestTransaction()
	_, err := client.SubmitTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("tx submission failed: %w", err)
	}

	// Wait for confirmation
	confirmed, err := waitForConfirmation(ctx, tx.Hash(), 30*time.Second)
	if err != nil {
		return fmt.Errorf("tx confirmation failed: %w", err)
	}

	if !confirmed {
		return errors.New("tx not confirmed within timeout")
	}

	return nil
}

// Deployment provisioning check
func checkDeploymentProvisioning(ctx context.Context) error {
	client := getProviderClient()

	manifest := createTestManifest()
	deployment, err := client.CreateDeployment(ctx, manifest)
	if err != nil {
		return fmt.Errorf("deployment creation failed: %w", err)
	}

	// Wait for running state
	err = waitForRunning(ctx, deployment.ID, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("deployment failed to start: %w", err)
	}

	// Cleanup
	defer client.CloseDeployment(ctx, deployment.ID)

	return nil
}

// API query check
func checkAPIQuery(ctx context.Context) error {
	client := getAPIClient()

	start := time.Now()
	_, err := client.QueryMarkets(ctx)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	latency := time.Since(start)
	if latency > 2*time.Second {
		return fmt.Errorf("query too slow: %v", latency)
	}

	return nil
}
```

**Schedule**: Every 1-5 minutes depending on criticality

---

### 2. End-to-End Testing

**Purpose**: Validate complete user workflows

**Example Test**:

```go
// tests/e2e/deployment_lifecycle_test.go

func TestDeploymentLifecycle(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create deployment
	t.Log("Creating deployment...")
	deployment := createDeployment(t, ctx, testManifest)
	require.NotNil(t, deployment)

	// Step 2: Place bid
	t.Log("Waiting for bids...")
	bids := waitForBids(t, ctx, deployment.ID, 30*time.Second)
	require.NotEmpty(t, bids, "no bids received")

	// Step 3: Accept bid
	t.Log("Accepting bid...")
	err := acceptBid(ctx, deployment.ID, bids[0].ID)
	require.NoError(t, err)

	// Step 4: Wait for provisioning
	t.Log("Waiting for provisioning...")
	err = waitForStatus(ctx, deployment.ID, StatusRunning, 5*time.Minute)
	require.NoError(t, err)

	// Step 5: Verify deployment health
	t.Log("Checking deployment health...")
	healthy, err := checkDeploymentHealth(ctx, deployment.ID)
	require.NoError(t, err)
	require.True(t, healthy)

	// Step 6: Close deployment
	t.Log("Closing deployment...")
	err = closeDeployment(ctx, deployment.ID)
	require.NoError(t, err)

	// Step 7: Verify settlement
	t.Log("Verifying settlement...")
	settled, err := checkSettlement(ctx, deployment.ID)
	require.NoError(t, err)
	require.True(t, settled)
}
```

**Schedule**: Pre-release, nightly in staging/production

---

### 3. Performance Testing

**Purpose**: Validate performance under load

**Existing Framework**: `tests/load/scenarios_test.go`

**Enhanced Tests**:

```go
// tests/load/reliability_scenarios_test.go

func TestSustainedLoadReliability(t *testing.T) {
	scenario := LoadScenario{
		Name:          "Sustained 50 TPS for 1 hour",
		Duration:      1 * time.Hour,
		TargetTPS:     50,
		RampUpPeriod:  5 * time.Minute,
		Workers:       10,
		ErrorBudget:   0.01, // 1% error rate allowed
	}

	results := runLoadTest(t, scenario)

	// Validate against performance budgets
	assert.Less(t, results.ErrorRate, 0.01,
		"Error rate exceeded budget")
	assert.Less(t, results.LatencyP95, 10*time.Second,
		"P95 latency exceeded budget")
	assert.Greater(t, results.ThroughputAvg, 50.0,
		"Average throughput below target")
}

func TestBurstLoadReliability(t *testing.T) {
	scenario := LoadScenario{
		Name:         "Burst to 100 TPS for 5 minutes",
		Duration:     5 * time.Minute,
		TargetTPS:    100,
		RampUpPeriod: 30 * time.Second,
		Workers:      20,
		ErrorBudget:  0.02, // 2% error rate allowed during burst
	}

	results := runLoadTest(t, scenario)

	// System should handle burst without major degradation
	assert.Less(t, results.ErrorRate, 0.02)
	assert.Less(t, results.LatencyP95, 15*time.Second,
		"P95 latency during burst exceeded threshold")
}
```

**Schedule**: Weekly, pre-release

---

## Chaos Engineering

### Philosophy

> "The best way to avoid failure is to fail constantly." - Netflix Chaos Monkey

**Goals**:
1. Find weaknesses before they cause outages
2. Build confidence in system resilience
3. Practice incident response
4. Validate monitoring and alerting

### Chaos Experiments

#### Experiment 1: Random Pod Termination

**Hypothesis**: System handles pod failures gracefully

**Implementation**:

```yaml
# chaos/experiments/pod-kill.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: virtengine-pod-kill
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: virtengine-node
  scheduler:
    cron: '@every 6h'
```

**Success Criteria**:
- [ ] Service maintains availability (no SLO violation)
- [ ] Kubernetes reschedules pod within 30 seconds
- [ ] Consensus continues (validator set remains functional)
- [ ] Alerts fire correctly
- [ ] On-call responds appropriately

**Schedule**: Every 6 hours in production

---

#### Experiment 2: Network Latency Injection

**Hypothesis**: System tolerates increased network latency

**Implementation**:

```yaml
# chaos/experiments/network-latency.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-latency
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: provider-daemon
  delay:
    latency: "100ms"
    correlation: "25"
    jitter: "10ms"
  duration: "10m"
  scheduler:
    cron: '@daily'
```

**Success Criteria**:
- [ ] Deployment provisioning stays within 350s (300s + 50s tolerance)
- [ ] Bid placement latency increases but stays < 10s
- [ ] No cascading failures
- [ ] Metrics show latency increase

**Schedule**: Daily in staging

---

#### Experiment 3: Disk Pressure

**Hypothesis**: System handles disk pressure gracefully

**Implementation**:

```yaml
# chaos/experiments/disk-pressure.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: disk-pressure
spec:
  mode: one
  selector:
    namespaces:
      - virtengine
    labelSelectors:
      app: virtengine-node
  stressors:
    ioChaos:
      workers: 4
      size: '1GB'
  duration: '30m'
```

**Success Criteria**:
- [ ] System detects disk pressure
- [ ] Alerts fire for high disk usage
- [ ] Performance degrades gracefully
- [ ] System recovers after experiment ends

**Schedule**: Weekly in staging

---

#### Experiment 4: Database Failure

**Hypothesis**: System handles database outage with degraded mode

**Implementation**:

```bash
#!/bin/bash
# chaos/experiments/db-failure.sh

# Simulate database outage by blocking database port
kubectl exec -it postgres-0 -- iptables -A INPUT -p tcp --dport 5432 -j DROP

# Wait 5 minutes
sleep 300

# Restore
kubectl exec -it postgres-0 -- iptables -D INPUT -p tcp --dport 5432 -j DROP
```

**Success Criteria**:
- [ ] API returns cached responses
- [ ] Writes queue for retry
- [ ] Error rate stays < 5%
- [ ] System recovers fully when DB restored

**Schedule**: Monthly in staging, quarterly in production (off-hours)

---

### Chaos Testing Tools

**Recommended Tools**:
1. **Chaos Mesh**: Kubernetes-native chaos engineering
2. **Litmus**: Cloud-native chaos engineering framework
3. **Gremlin**: Enterprise chaos engineering platform
4. **Pumba**: Docker chaos testing

**Integration**:

```bash
# Install Chaos Mesh
kubectl apply -f https://mirrors.chaos-mesh.org/latest/chaos-mesh.yaml

# Verify installation
kubectl get pods -n chaos-mesh

# Run experiment
kubectl apply -f chaos/experiments/pod-kill.yaml

# Observe results
kubectl logs -n chaos-mesh -l app.kubernetes.io/component=controller-manager
```

---

## Load and Stress Testing

### Load Testing Strategy

**Levels**:
1. **Smoke Test**: 10% of production load (validate basic functionality)
2. **Load Test**: 100% of production load (validate SLOs at scale)
3. **Stress Test**: 150-200% of production load (find breaking point)
4. **Soak Test**: 100% load for extended period (find memory leaks)

### Load Test Scenarios

#### Scenario 1: Transaction Load

```go
// tests/load/transaction_load_test.go

func TestTransactionLoad(t *testing.T) {
	tests := []struct {
		name      string
		tps       int
		duration  time.Duration
		maxErrors float64
	}{
		{"Smoke", 5, 5 * time.Minute, 0.01},
		{"Load", 50, 30 * time.Minute, 0.01},
		{"Stress", 100, 10 * time.Minute, 0.05},
		{"Soak", 50, 2 * time.Hour, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := LoadScenario{
				Name:         tt.name,
				TargetTPS:    tt.tps,
				Duration:     tt.duration,
				RampUpPeriod: 1 * time.Minute,
				Workers:      tt.tps / 5,
				ErrorBudget:  tt.maxErrors,
			}

			results := runLoadTest(t, scenario)

			// Validate results
			assert.Less(t, results.ErrorRate, tt.maxErrors)
			t.Logf("Results: TPS=%.2f, P95=%.2fs, Errors=%.2f%%",
				results.ThroughputAvg,
				results.LatencyP95.Seconds(),
				results.ErrorRate*100)
		})
	}
}
```

#### Scenario 2: Deployment Load

```go
func TestDeploymentLoad(t *testing.T) {
	scenario := LoadScenario{
		Name:          "Concurrent Deployments",
		Concurrency:   50,
		Duration:      30 * time.Minute,
		ErrorBudget:   0.01,
	}

	results := runDeploymentLoadTest(t, scenario)

	// Validate provider capacity
	assert.Less(t, results.ProvisioningP95, 300*time.Second)
	assert.Greater(t, results.SuccessRate, 0.99)
}
```

**Schedule**:
- **Smoke**: Pre-deployment (CI/CD)
- **Load**: Weekly
- **Stress**: Monthly
- **Soak**: Quarterly

---

## Failure Mode Testing

### Common Failure Modes

#### 1. Cascading Failures

**Test**: Overload one component, observe propagation

```go
func TestCascadingFailureProtection(t *testing.T) {
	// Overload provider bid engine
	overloadBidEngine(t, 1000 /* requests/sec */)

	// Verify circuit breaker trips
	time.Sleep(30 * time.Second)
	stats := getBidEngineStats()
	assert.True(t, stats.CircuitBreakerOpen,
		"Circuit breaker should open under overload")

	// Verify other components continue functioning
	apiHealth := checkAPIHealth()
	assert.True(t, apiHealth, "API should remain healthy")
}
```

**Expected Behavior**:
- Circuit breaker opens
- Requests fail fast
- Other components unaffected

---

#### 2. Split Brain

**Test**: Partition validator network

```bash
#!/bin/bash
# chaos/experiments/network-partition.sh

# Create network partition (2 validators isolated)
kubectl exec validator-0 -- iptables -A INPUT -s validator-2-ip -j DROP
kubectl exec validator-1 -- iptables -A INPUT -s validator-2-ip -j DROP

# Wait for detection (should fail to reach consensus)
sleep 60

# Heal partition
kubectl exec validator-0 -- iptables -D INPUT -s validator-2-ip -j DROP
kubectl exec validator-1 -- iptables -D INPUT -s validator-2-ip -j DROP
```

**Success Criteria**:
- [ ] Consensus stops (no blocks produced)
- [ ] Alerts fire immediately
- [ ] System recovers when partition healed
- [ ] No data loss

---

#### 3. Resource Exhaustion

**Test**: Gradually exhaust resources

```go
func TestMemoryExhaustion(t *testing.T) {
	// Monitor memory usage
	initialMem := getMemoryUsage()

	// Create memory pressure
	allocateMemory(t, 8*GB)

	// Verify graceful degradation
	time.Sleep(1 * time.Minute)

	// Check alerts
	alerts := getActiveAlerts()
	assert.Contains(t, alerts, "HighMemoryUsage")

	// Verify no OOM kill
	podStatus := getPodStatus("virtengine-node-0")
	assert.NotEqual(t, "OOMKilled", podStatus.LastTerminationReason)
}
```

---

## Disaster Recovery Testing

### DR Scenarios

#### Scenario 1: Complete Data Center Loss

**Test Procedure**:
1. Simulate data center outage (shut down all nodes in region)
2. Failover to backup region
3. Restore services
4. Verify data integrity

**Success Criteria**:
- [ ] Failover completes within 15 minutes
- [ ] Data loss < 1 minute of transactions
- [ ] Services resume normal operation
- [ ] RTO: < 15 minutes
- [ ] RPO: < 1 minute

#### Scenario 2: Database Corruption

**Test Procedure**:
1. Corrupt database (simulate hardware failure)
2. Restore from backup
3. Replay transaction log
4. Verify state consistency

**Success Criteria**:
- [ ] Restore completes within 2 hours
- [ ] No data loss (WAL replay successful)
- [ ] State matches blockchain consensus
- [ ] RTO: < 2 hours
- [ ] RPO: 0 (no acceptable data loss)

#### Scenario 3: Complete System Rebuild

**Test Procedure**:
1. Start with blank infrastructure
2. Run Infrastructure as Code
3. Restore from backups
4. Rejoin network

**Success Criteria**:
- [ ] Rebuild completes within 4 hours
- [ ] All services operational
- [ ] Consensus participation resumed
- [ ] Monitoring fully functional

**Schedule**: Quarterly DR drill

---

## Testing Schedule

### Continuous (Automated)

| Test Type | Frequency | Environment |
|-----------|-----------|-------------|
| Synthetic Monitoring | Every 1-5 min | Production |
| Unit Tests | Every commit | CI |
| Integration Tests | Every PR | CI |
| E2E Tests | Nightly | Staging |

### Regular (Scheduled)

| Test Type | Frequency | Environment |
|-----------|-----------|-------------|
| Load Tests | Weekly | Staging |
| Stress Tests | Monthly | Staging |
| Soak Tests | Quarterly | Staging |
| Chaos: Pod Kill | Every 6 hours | Production |
| Chaos: Network Latency | Daily | Staging |
| Chaos: Disk Pressure | Weekly | Staging |
| Chaos: Database Failure | Monthly | Staging |

### Periodic (Planned)

| Test Type | Frequency | Environment |
|-----------|-----------|-------------|
| DR Drill | Quarterly | Production (off-hours) |
| Gameday Exercise | Quarterly | Production (controlled) |
| Performance Regression | Pre-release | Staging |

---

## Gameday Exercises

### What is a Gameday?

A **gameday** is a practice incident where the team simulates a real outage to:
- Test incident response procedures
- Practice communication
- Identify gaps in runbooks
- Build muscle memory
- Reduce stress during real incidents

### Gameday Schedule

**Quarterly Gameday**: Last Friday of quarter, 2-hour window

**Example Scenarios**:
1. **Q1**: Database outage during peak traffic
2. **Q2**: Multi-region network partition
3. **Q3**: Critical security vulnerability disclosure
4. **Q4**: Complete provider network failure

### Gameday Template

```markdown
## Gameday: Database Outage

**Date**: 2026-03-28, 14:00-16:00 UTC
**Facilitator**: SRE Lead
**Participants**: On-call rotation, Engineering leads

### Scenario
At 14:00, the primary database becomes unavailable.
Symptoms:
- API errors increase to 50%
- Query latency spikes to 30s
- Alerts firing

### Timeline
14:00 - Incident injected
14:05 - On-call paged
14:10 - Incident declared
14:15 - War room established
14:20 - Root cause identified
14:30 - Failover initiated
14:45 - Services restored
15:00 - Incident closed
15:00-16:00 - Hot wash / Retrospective

### Success Criteria
- [ ] Incident detected within 5 minutes
- [ ] On-call paged within 5 minutes
- [ ] Services restored within 30 minutes
- [ ] Communication followed process
- [ ] Runbooks used effectively

### Learnings
[To be filled during retrospective]
```

---

## References

- [Google SRE Book - Testing for Reliability](https://sre.google/sre-book/testing-reliability/)
- [Chaos Engineering Principles](https://principlesofchaos.org/)
- [SLI/SLO/SLA Framework](SLI_SLO_SLA.md)
- [Incident Response](INCIDENT_RESPONSE.md)
- [Load Testing Framework](../tests/load/README.md)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
