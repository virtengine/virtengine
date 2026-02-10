# Incident Simulation Lab

**Duration:** 4 hours  
**Prerequisites:** [Lab Environment Setup](lab-environment.md), [Validator Operations Lab](validator-ops-lab.md)  
**Difficulty:** Advanced

---

## Table of Contents

1. [Overview](#overview)
2. [Learning Objectives](#learning-objectives)
3. [Lab Setup](#lab-setup)
4. [Simulation 1: High Error Rate Incident](#simulation-1-high-error-rate-incident)
5. [Simulation 2: Node Down Scenario](#simulation-2-node-down-scenario)
6. [Simulation 3: VEID Scoring Anomaly](#simulation-3-veid-scoring-anomaly)
7. [Simulation 4: Provider Deployment Failure](#simulation-4-provider-deployment-failure)
8. [Debrief and Post-Mortem](#debrief-and-post-mortem)
9. [Summary](#summary)

---

## Overview

Incident response is a critical skill for VirtEngine operators. This lab simulates real-world incidents to practice:

- **Detection**: Identifying problems through monitoring
- **Triage**: Assessing impact and urgency
- **Diagnosis**: Finding root causes
- **Remediation**: Fixing issues and restoring service
- **Communication**: Keeping stakeholders informed
- **Post-Mortem**: Learning from incidents

### Incident Response Framework

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       Incident Response Lifecycle                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   1. DETECT        2. TRIAGE       3. DIAGNOSE      4. REMEDIATE        │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐      ┌─────────┐         │
│   │ Monitor │ ──► │ Assess  │ ──► │ Find    │ ──►  │ Fix     │         │
│   │ Alert   │     │ Impact  │     │ Root    │      │ Restore │         │
│   └─────────┘     └─────────┘     │ Cause   │      └─────────┘         │
│                                    └─────────┘                           │
│                                                                          │
│   5. COMMUNICATE                  6. POST-MORTEM                        │
│   ┌─────────────────────┐        ┌─────────────────────┐               │
│   │ Status Updates      │        │ Document            │               │
│   │ Stakeholder Notify  │        │ Action Items        │               │
│   └─────────────────────┘        └─────────────────────┘               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Severity Levels

| Level | Description | Response Time | Example |
|-------|-------------|---------------|---------|
| SEV-1 | Critical - Complete outage | < 15 min | All validators down |
| SEV-2 | Major - Significant degradation | < 30 min | High error rates |
| SEV-3 | Minor - Limited impact | < 2 hours | Single component issue |
| SEV-4 | Low - Minimal impact | < 24 hours | Non-critical feature |

---

## Learning Objectives

By the end of this lab, you will be able to:

- [ ] Inject controlled failures for training purposes
- [ ] Detect incidents using monitoring tools
- [ ] Follow structured incident response procedures
- [ ] Diagnose root causes systematically
- [ ] Execute remediation steps safely
- [ ] Conduct effective post-mortem analysis

---

## Lab Setup

### Prerequisites Check

Ensure your environment is ready:

```bash
# Verify localnet is running
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh status

# Verify monitoring is accessible
curl -s http://localhost:9095/-/ready  # Prometheus
curl -s http://localhost:3002/api/health  # Grafana

# Verify chain is healthy
curl -s http://localhost:26657/health
```

### Create Incident Response Directory

```bash
mkdir -p ~/incident-lab/{logs,evidence,postmortems}
cd ~/incident-lab

# Create incident log template
cat > ./incident-template.md << 'EOF'
# Incident Report: [TITLE]

## Metadata
- **Incident ID:** INC-YYYYMMDD-NNN
- **Severity:** SEV-[1-4]
- **Status:** [Investigating|Identified|Monitoring|Resolved]
- **Started:** YYYY-MM-DD HH:MM UTC
- **Resolved:** YYYY-MM-DD HH:MM UTC
- **Duration:** X hours Y minutes

## Impact
[Description of user/system impact]

## Timeline
| Time (UTC) | Event |
|------------|-------|
| HH:MM | First detection |
| HH:MM | Investigation started |
| HH:MM | Root cause identified |
| HH:MM | Remediation applied |
| HH:MM | Service restored |

## Root Cause
[Technical description of what went wrong]

## Resolution
[Steps taken to resolve]

## Action Items
- [ ] Action item 1
- [ ] Action item 2

## Lessons Learned
[What can we improve?]
EOF
```

### Set Up Chaos Injection Tools

```bash
# Create chaos injection helper
cat > ./chaos-tools.sh << 'EOF'
#!/bin/bash

# VirtEngine Chaos Injection Tools
# WARNING: For training environments only!

LOCALNET_DIR="$HOME/virtengine-lab/virtengine"

inject_network_latency() {
    local container=$1
    local latency=${2:-100ms}
    echo "Injecting ${latency} network latency to $container..."
    docker exec $container tc qdisc add dev eth0 root netem delay $latency 2>/dev/null || \
    docker exec $container tc qdisc change dev eth0 root netem delay $latency
}

remove_network_latency() {
    local container=$1
    echo "Removing network latency from $container..."
    docker exec $container tc qdisc del dev eth0 root 2>/dev/null || true
}

inject_cpu_stress() {
    local container=$1
    local duration=${2:-60}
    echo "Injecting CPU stress to $container for ${duration}s..."
    docker exec -d $container sh -c "stress --cpu 4 --timeout $duration" 2>/dev/null || \
    docker exec -d $container sh -c "yes > /dev/null &"
}

kill_container() {
    local container=$1
    echo "Stopping container $container..."
    docker stop $container
}

restart_container() {
    local container=$1
    echo "Restarting container $container..."
    docker start $container
}

pause_container() {
    local container=$1
    echo "Pausing container $container..."
    docker pause $container
}

unpause_container() {
    local container=$1
    echo "Unpausing container $container..."
    docker unpause $container
}

list_containers() {
    echo "VirtEngine containers:"
    docker ps --filter "name=virtengine" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

case "$1" in
    list)
        list_containers
        ;;
    kill)
        kill_container "$2"
        ;;
    restart)
        restart_container "$2"
        ;;
    pause)
        pause_container "$2"
        ;;
    unpause)
        unpause_container "$2"
        ;;
    latency)
        inject_network_latency "$2" "$3"
        ;;
    latency-remove)
        remove_network_latency "$2"
        ;;
    cpu-stress)
        inject_cpu_stress "$2" "$3"
        ;;
    *)
        echo "Usage: $0 {list|kill|restart|pause|unpause|latency|latency-remove|cpu-stress} [container] [args]"
        ;;
esac
EOF

chmod +x ./chaos-tools.sh
```

---

## Simulation 1: High Error Rate Incident

### Scenario
The API gateway is returning HTTP 500 errors at a high rate, affecting customer deployments.

### Duration
45 minutes

### Chaos Injection

```bash
cd ~/incident-lab

# Start timing
echo "=== SIMULATION 1: High Error Rate ===" 
echo "Start Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo ""

# Inject fault - pause the virtengine node container briefly to cause errors
CONTAINER=$(docker ps --filter "name=virtengine" --format "{{.Names}}" | head -1)
echo "Target container: $CONTAINER"

# Create intermittent failures by pausing/unpausing
echo "Injecting fault..."
for i in {1..5}; do
    docker pause $CONTAINER
    sleep 2
    docker unpause $CONTAINER
    sleep 3
done &

echo "Fault injection started. Begin incident response!"
```

### Detection Phase

#### Step 1: Observe the Alert

```bash
# Check Prometheus for errors
curl -s "http://localhost:9095/api/v1/query?query=increase(http_requests_total{status=~\"5..\"}[5m])" | jq '.data.result'

# Check chain health
curl -s http://localhost:26657/health
```

#### Step 2: Confirm the Issue

```bash
# Test RPC endpoint repeatedly
for i in {1..10}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:26657/status)
    echo "Request $i: HTTP $STATUS"
    sleep 1
done
```

### Triage Phase

#### Step 3: Assess Impact

```bash
# Document impact assessment
cat > ./logs/sim1-impact.md << 'EOF'
## Impact Assessment

**Affected Services:**
- [x] Chain RPC endpoints
- [x] REST API
- [x] Client transactions

**User Impact:**
- Users unable to submit transactions
- Queries failing intermittently
- Estimated ~50% of requests failing

**Severity Assessment:** SEV-2 (Major degradation)
EOF

cat ./logs/sim1-impact.md
```

### Diagnosis Phase

#### Step 4: Check Container Status

```bash
# List all containers
docker ps --filter "name=virtengine" --format "table {{.Names}}\t{{.Status}}"

# Check for paused containers
docker ps --filter "status=paused" --format "{{.Names}}"

# Check container logs
docker logs --tail 50 $CONTAINER 2>&1 | tail -20
```

#### Step 5: Check System Resources

```bash
# Check container stats
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"

# Check for resource limits
docker inspect $CONTAINER | jq '.[0].HostConfig.Memory, .[0].HostConfig.CpuShares'
```

#### Step 6: Identify Root Cause

```bash
# Root cause identification
echo "=== Root Cause Analysis ===" | tee ./logs/sim1-rca.txt
echo "Symptom: HTTP 500 errors from RPC endpoints" | tee -a ./logs/sim1-rca.txt
echo "Finding: Container intermittently paused" | tee -a ./logs/sim1-rca.txt
echo "Root Cause: Container orchestration issue causing pause/unpause" | tee -a ./logs/sim1-rca.txt
```

### Remediation Phase

#### Step 7: Stop the Fault

```bash
# Wait for chaos injection to complete or stop it
jobs -l
kill %1 2>/dev/null || true

# Ensure container is unpaused
docker unpause $CONTAINER 2>/dev/null || true
```

#### Step 8: Verify Recovery

```bash
# Test RPC endpoint
for i in {1..10}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:26657/status)
    echo "Request $i: HTTP $STATUS"
    sleep 1
done

# All should return 200
```

#### Step 9: Document Resolution

```bash
cat > ./logs/sim1-resolution.md << 'EOF'
## Resolution

**Actions Taken:**
1. Identified container pause events in Docker
2. Stopped automated pause/unpause process
3. Verified container running normally
4. Confirmed all endpoints returning 200

**Time to Resolution:** ~15 minutes

**Follow-up Actions:**
- [ ] Review Docker orchestration configuration
- [ ] Add alerting for container state changes
- [ ] Implement container health checks
EOF

cat ./logs/sim1-resolution.md
```

### Verification

```bash
# Final health check
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Chain should be progressing
```

---

## Simulation 2: Node Down Scenario

### Scenario
A validator node has crashed and is not responding. Block production has halted or slowed.

### Duration
45 minutes

### Chaos Injection

```bash
cd ~/incident-lab

echo "=== SIMULATION 2: Node Down ===" 
echo "Start Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Record current block height
INITIAL_HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "Initial block height: $INITIAL_HEIGHT"

# Stop the main node container
CONTAINER=$(docker ps --filter "name=virtengine-node" --format "{{.Names}}" | head -1)
echo "Stopping container: $CONTAINER"
docker stop $CONTAINER

echo ""
echo "Node stopped! Begin incident response!"
```

### Detection Phase

#### Step 1: Detect the Outage

```bash
# Try to connect to RPC
curl -s http://localhost:26657/status
# Should fail or timeout

# Check Prometheus for down target
curl -s "http://localhost:9095/api/v1/query?query=up{job=~\".*virtengine.*\"}" | jq '.data.result[] | {instance: .metric.instance, up: .value[1]}'
```

#### Step 2: Confirm Node Down

```bash
# Check container status
docker ps -a --filter "name=virtengine" --format "table {{.Names}}\t{{.Status}}"

# Check Docker logs
docker logs --tail 50 $CONTAINER 2>&1
```

### Triage Phase

#### Step 3: Assess Impact

```bash
cat > ./logs/sim2-impact.md << 'EOF'
## Impact Assessment

**Affected Services:**
- Chain RPC: DOWN
- REST API: DOWN
- gRPC: DOWN
- Block Production: HALTED

**User Impact:**
- Complete service outage
- No transactions can be processed
- All queries failing

**Severity Assessment:** SEV-1 (Critical outage)
EOF

cat ./logs/sim2-impact.md
```

### Diagnosis Phase

#### Step 4: Check Container Exit Reason

```bash
# Inspect container exit status
docker inspect $CONTAINER | jq '.[0].State'

# Check for OOM kill
docker inspect $CONTAINER | jq '.[0].State.OOMKilled'

# Check exit code
docker inspect $CONTAINER | jq '.[0].State.ExitCode'
```

#### Step 5: Review Logs for Crash Reason

```bash
# Get last logs before crash
docker logs --tail 100 $CONTAINER 2>&1 | tee ./logs/sim2-crash-logs.txt

# Look for panic or fatal errors
grep -i -E "(panic|fatal|error|failed)" ./logs/sim2-crash-logs.txt
```

### Remediation Phase

#### Step 6: Restart the Node

```bash
echo "Restarting node..."
docker start $CONTAINER

# Wait for startup
echo "Waiting for node to start..."
sleep 15
```

#### Step 7: Verify Node Recovery

```bash
# Check node is responding
curl -s http://localhost:26657/health

# Get current block height
NEW_HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "New block height: $NEW_HEIGHT"
echo "Initial height was: $INITIAL_HEIGHT"

# Check if catching up
CATCHING_UP=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.catching_up')
echo "Catching up: $CATCHING_UP"
```

#### Step 8: Monitor Recovery Progress

```bash
# Watch block progression
for i in {1..10}; do
    HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
    echo "Block height: $HEIGHT"
    sleep 5
done
```

#### Step 9: Document Resolution

```bash
cat > ./logs/sim2-resolution.md << 'EOF'
## Resolution

**Root Cause:** Node container stopped (simulated crash)

**Actions Taken:**
1. Detected RPC endpoint unresponsive
2. Identified container in stopped state
3. Reviewed logs for crash reason
4. Restarted container
5. Verified block production resumed

**Time to Resolution:** ~10 minutes
**Blocks Missed:** ~20

**Follow-up Actions:**
- [ ] Implement automatic container restart
- [ ] Add alerting for container state
- [ ] Review resource limits
- [ ] Consider redundant nodes
EOF

cat ./logs/sim2-resolution.md
```

### Verification

```bash
# Final health check
./scripts/localnet.sh status 2>/dev/null || \
    (cd ~/virtengine-lab/virtengine && ./scripts/localnet.sh status)
```

---

## Simulation 3: VEID Scoring Anomaly

### Scenario
Identity verification (VEID) scores are returning unexpected values - all scores are being rejected as too low.

### Duration
45 minutes

### Chaos Injection

```bash
cd ~/incident-lab

echo "=== SIMULATION 3: VEID Scoring Anomaly ===" 
echo "Start Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

# For this simulation, we'll create a mock scenario
# In production, this could be caused by:
# - Model file corruption
# - Incorrect model version
# - TensorFlow configuration issues
# - Determinism violations

# Create simulated anomaly logs
cat > ./logs/sim3-anomaly-data.json << 'EOF'
{
  "timestamp": "2026-01-24T12:00:00Z",
  "event": "veid_scoring_batch",
  "results": [
    {"request_id": "req-001", "score": 0.15, "threshold": 0.75, "status": "rejected"},
    {"request_id": "req-002", "score": 0.12, "threshold": 0.75, "status": "rejected"},
    {"request_id": "req-003", "score": 0.18, "threshold": 0.75, "status": "rejected"},
    {"request_id": "req-004", "score": 0.11, "threshold": 0.75, "status": "rejected"},
    {"request_id": "req-005", "score": 0.14, "threshold": 0.75, "status": "rejected"}
  ],
  "normal_average": 0.82,
  "current_average": 0.14,
  "anomaly_detected": true
}
EOF

echo "Anomaly data created. Begin investigation!"
```

### Detection Phase

#### Step 1: Review Anomaly Data

```bash
# Display the anomaly
cat ./logs/sim3-anomaly-data.json | jq '.'

# Calculate statistics
echo "=== Anomaly Statistics ==="
echo "Normal average score: 0.82"
echo "Current average score: $(cat ./logs/sim3-anomaly-data.json | jq '.current_average')"
echo "Rejection rate: 100%"
```

#### Step 2: Check Metrics

```bash
# In production, check Prometheus
cat > ./logs/sim3-metrics.txt << 'EOF'
virtengine_veid_scoring_score_avg{} = 0.14  # Normal: 0.82
virtengine_veid_scoring_rejections_total{} = 500  # Last hour
virtengine_veid_scoring_errors_total{} = 0
virtengine_veid_model_loaded{version="1.0.0"} = 1
EOF

cat ./logs/sim3-metrics.txt
```

### Triage Phase

#### Step 3: Assess Impact

```bash
cat > ./logs/sim3-impact.md << 'EOF'
## Impact Assessment

**Affected Services:**
- VEID identity verification

**User Impact:**
- All identity verifications being rejected
- Users cannot complete VEID enrollment
- Estimated 100% failure rate

**Severity Assessment:** SEV-2 (Major degradation)

**Note:** No transaction failures, chain healthy
EOF

cat ./logs/sim3-impact.md
```

### Diagnosis Phase

#### Step 4: Check Model Configuration

```bash
# Simulate model check
cat > ./logs/sim3-model-check.txt << 'EOF'
=== Model Configuration Check ===

Current Model: veid_scorer_v1.0.0.h5
Expected Hash: a1b2c3d4e5f6...
Actual Hash:   a1b2c3d4e5f6...  ✓ Match

TensorFlow Version: 2.15.0  ✓ Correct
Determinism Enabled: true   ✓ Correct
Random Seed: 42             ✓ Correct
Force CPU: true             ✓ Correct

=== Model Load Test ===
Model loaded successfully
Input shape: (None, 224, 224, 3)
Output shape: (None, 1)
EOF

cat ./logs/sim3-model-check.txt
echo ""
echo "Model configuration appears correct..."
```

#### Step 5: Analyze Input Data

```bash
# Simulate input analysis
cat > ./logs/sim3-input-analysis.txt << 'EOF'
=== Input Data Analysis ===

Recent Input Samples:
- Sample 001: Image size 224x224, normalized ✓
- Sample 002: Image size 224x224, normalized ✓
- Sample 003: Image size 224x224, normalized ✓

Input Statistics:
- Mean pixel value: 0.12  ⚠️ ABNORMAL (expected ~0.5)
- Std deviation: 0.05    ⚠️ ABNORMAL (expected ~0.25)
- Range: [0.0, 0.25]     ⚠️ ABNORMAL (expected [0.0, 1.0])

=== FINDING ===
Input images appear to be under-exposed or incorrectly normalized!
Preprocessing pipeline may have changed.
EOF

cat ./logs/sim3-input-analysis.txt
```

#### Step 6: Identify Root Cause

```bash
cat > ./logs/sim3-rca.txt << 'EOF'
=== Root Cause Analysis ===

SYMPTOM: All VEID scores abnormally low (0.14 vs 0.82 normal)

INVESTIGATION:
1. Model file: ✓ Correct version and hash
2. TensorFlow config: ✓ Determinism enabled
3. Input preprocessing: ⚠️ ISSUE FOUND

ROOT CAUSE:
A recent update to the image preprocessing pipeline changed the 
normalization range from [0, 1] to [0, 0.25]. This causes the model
to see all images as very dark, resulting in low confidence scores.

AFFECTED COMPONENT: pkg/inference/preprocessor.go line 142
CHANGE: Normalization divisor changed from 255.0 to 1000.0
COMMIT: abc123 (2026-01-24 10:30:00 UTC)
EOF

cat ./logs/sim3-rca.txt
```

### Remediation Phase

#### Step 7: Simulate Fix

```bash
# Document the fix
cat > ./logs/sim3-fix.txt << 'EOF'
=== Remediation Steps ===

1. Identify bad commit: abc123
2. Prepare rollback:
   git revert abc123

3. Or apply hotfix:
   - File: pkg/inference/preprocessor.go
   - Line 142: Change divisor from 1000.0 to 255.0
   
4. Deploy fix:
   make virtengine
   # Rolling restart of validators

5. Verify fix:
   - Run test suite: make test-inference
   - Check sample scores return to normal range
EOF

cat ./logs/sim3-fix.txt
```

#### Step 8: Verify Resolution

```bash
# Simulate post-fix metrics
cat > ./logs/sim3-post-fix.txt << 'EOF'
=== Post-Fix Verification ===

Test Batch Results:
- Request 001: score=0.85, status=approved ✓
- Request 002: score=0.79, status=approved ✓
- Request 003: score=0.88, status=approved ✓
- Request 004: score=0.72, status=rejected (legitimate low score)
- Request 005: score=0.91, status=approved ✓

Average Score: 0.83 (normal range: 0.75-0.90) ✓
Approval Rate: 80% (normal) ✓

RESOLUTION CONFIRMED
EOF

cat ./logs/sim3-post-fix.txt
```

### Documentation

```bash
cat > ./logs/sim3-resolution.md << 'EOF'
## Resolution

**Root Cause:** Preprocessing pipeline normalization bug in commit abc123

**Actions Taken:**
1. Detected anomaly via metrics showing 100% rejection rate
2. Verified model configuration was correct
3. Analyzed input data and found normalization issue
4. Traced to recent commit changing divisor value
5. Reverted change and deployed fix
6. Verified scores returned to normal

**Time to Detection:** 15 minutes
**Time to Resolution:** 35 minutes
**Total Duration:** 50 minutes

**Follow-up Actions:**
- [ ] Add automated tests for preprocessing output ranges
- [ ] Add alerting for score distribution changes
- [ ] Require ML team review for inference pipeline changes
- [ ] Implement canary deployment for scoring changes
EOF

cat ./logs/sim3-resolution.md
```

---

## Simulation 4: Provider Deployment Failure

### Scenario
Customer workloads are failing to deploy. The provider is accepting bids but deployments never reach Running state.

### Duration
45 minutes

### Chaos Injection

```bash
cd ~/incident-lab

echo "=== SIMULATION 4: Provider Deployment Failure ===" 
echo "Start Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

# If Kind cluster exists, we'll simulate a quota exhaustion
# Otherwise, we'll use mock data

if kubectl cluster-info 2>/dev/null; then
    echo "Using live Kubernetes cluster for simulation"
    
    # Create a resource quota that's too small
    cat > ./sim4-restrictive-quota.yaml << 'EOF'
apiVersion: v1
kind: ResourceQuota
metadata:
  name: restrictive-quota
  namespace: virtengine-workloads
spec:
  hard:
    requests.cpu: "100m"
    requests.memory: "64Mi"
    limits.cpu: "100m"
    limits.memory: "64Mi"
    pods: "1"
EOF
    
    kubectl apply -f ./sim4-restrictive-quota.yaml
    echo "Restrictive quota applied!"
else
    echo "Using mock data for simulation"
fi

echo "Begin incident response!"
```

### Detection Phase

#### Step 1: Check Deployment Status

```bash
# List pending deployments
if kubectl cluster-info 2>/dev/null; then
    kubectl get pods -n virtengine-workloads -o wide
    kubectl get events -n virtengine-workloads --sort-by='.lastTimestamp' | tail -10
else
    # Mock data
    cat > ./logs/sim4-deployment-status.txt << 'EOF'
NAME                          READY   STATUS    RESTARTS   AGE
ve-deployment-001-abc123      0/1     Pending   0          15m
ve-deployment-002-def456      0/1     Pending   0          12m
ve-deployment-003-ghi789      0/1     Pending   0          8m

EVENTS:
Warning  FailedScheduling  pod/ve-deployment-001-abc123  0/3 nodes are available: 
         insufficient cpu, insufficient memory.
EOF
    cat ./logs/sim4-deployment-status.txt
fi
```

#### Step 2: Identify Pending Pods

```bash
if kubectl cluster-info 2>/dev/null; then
    kubectl describe pods -n virtengine-workloads | grep -A 5 "Events:"
else
    cat > ./logs/sim4-pod-events.txt << 'EOF'
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  10m   default-scheduler  0/3 nodes are available:
           3 Insufficient cpu, 3 Insufficient memory.
  Warning  FailedScheduling  5m    default-scheduler  0/3 nodes are available:
           3 Insufficient cpu, 3 Insufficient memory.
EOF
    cat ./logs/sim4-pod-events.txt
fi
```

### Triage Phase

#### Step 3: Assess Impact

```bash
cat > ./logs/sim4-impact.md << 'EOF'
## Impact Assessment

**Affected Services:**
- Provider workload deployments

**User Impact:**
- New deployments stuck in Pending state
- Active leases not being fulfilled
- 100% of new deployments failing

**Severity Assessment:** SEV-2 (Major degradation)

**Note:** Existing running workloads unaffected
EOF

cat ./logs/sim4-impact.md
```

### Diagnosis Phase

#### Step 4: Check Resource Quota

```bash
if kubectl cluster-info 2>/dev/null; then
    kubectl describe resourcequota -n virtengine-workloads
else
    cat > ./logs/sim4-quota.txt << 'EOF'
Name:             restrictive-quota
Namespace:        virtengine-workloads
Resource          Used    Hard
--------          ----    ----
limits.cpu        100m    100m     ⚠️ AT LIMIT
limits.memory     64Mi    64Mi     ⚠️ AT LIMIT
pods              1       1        ⚠️ AT LIMIT
requests.cpu      100m    100m     ⚠️ AT LIMIT
requests.memory   64Mi    64Mi     ⚠️ AT LIMIT
EOF
    cat ./logs/sim4-quota.txt
fi
```

#### Step 5: Check Node Capacity

```bash
if kubectl cluster-info 2>/dev/null; then
    kubectl describe nodes | grep -A 10 "Allocated resources:"
else
    cat > ./logs/sim4-node-capacity.txt << 'EOF'
=== Node Capacity Analysis ===

Node: virtengine-lab-worker
Allocatable:
  cpu:     2
  memory:  4Gi
Allocated:
  cpu:     1.5
  memory:  3Gi
Available:
  cpu:     500m
  memory:  1Gi

Finding: Nodes have capacity, but ResourceQuota is blocking allocations
EOF
    cat ./logs/sim4-node-capacity.txt
fi
```

#### Step 6: Identify Root Cause

```bash
cat > ./logs/sim4-rca.txt << 'EOF'
=== Root Cause Analysis ===

SYMPTOM: Deployments stuck in Pending state

INVESTIGATION:
1. Pod events show: "Insufficient cpu, Insufficient memory"
2. Node capacity: AVAILABLE
3. ResourceQuota: EXHAUSTED

ROOT CAUSE:
A restrictive ResourceQuota was applied to the namespace,
limiting total resource allocation to 100m CPU and 64Mi memory.
This is far below the requirements for typical workloads.

EVIDENCE:
- ResourceQuota shows 100% usage
- Nodes have available capacity
- Pods fail scheduling due to quota, not node capacity
EOF

cat ./logs/sim4-rca.txt
```

### Remediation Phase

#### Step 7: Fix Resource Quota

```bash
if kubectl cluster-info 2>/dev/null; then
    # Delete restrictive quota
    kubectl delete resourcequota restrictive-quota -n virtengine-workloads
    
    # Apply correct quota
    cat > ./sim4-correct-quota.yaml << 'EOF'
apiVersion: v1
kind: ResourceQuota
metadata:
  name: provider-quota
  namespace: virtengine-workloads
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 32Gi
    limits.cpu: "20"
    limits.memory: 64Gi
    pods: "50"
EOF
    kubectl apply -f ./sim4-correct-quota.yaml
    echo "Correct quota applied!"
else
    echo "Simulating quota fix..."
    echo "Deleted: restrictive-quota"
    echo "Applied: provider-quota with appropriate limits"
fi
```

#### Step 8: Verify Pending Pods Scheduled

```bash
if kubectl cluster-info 2>/dev/null; then
    echo "Waiting for pods to schedule..."
    sleep 10
    kubectl get pods -n virtengine-workloads
else
    cat > ./logs/sim4-post-fix.txt << 'EOF'
=== Post-Fix Status ===

NAME                          READY   STATUS    RESTARTS   AGE
ve-deployment-001-abc123      1/1     Running   0          20m
ve-deployment-002-def456      1/1     Running   0          17m
ve-deployment-003-ghi789      1/1     Running   0          13m

All previously Pending pods now Running ✓
EOF
    cat ./logs/sim4-post-fix.txt
fi
```

### Documentation

```bash
cat > ./logs/sim4-resolution.md << 'EOF'
## Resolution

**Root Cause:** Overly restrictive ResourceQuota blocking pod scheduling

**Actions Taken:**
1. Detected deployments stuck in Pending state
2. Examined pod events showing scheduling failures
3. Identified ResourceQuota exhaustion
4. Removed restrictive quota
5. Applied correct quota with appropriate limits
6. Verified pods transitioned to Running

**Time to Detection:** 10 minutes
**Time to Resolution:** 20 minutes
**Total Duration:** 30 minutes

**Follow-up Actions:**
- [ ] Add monitoring for ResourceQuota utilization
- [ ] Implement quota review process for changes
- [ ] Add pre-deployment resource checks
- [ ] Document standard quota configurations
EOF

cat ./logs/sim4-resolution.md
```

---

## Debrief and Post-Mortem

### Conduct Post-Mortem Meeting

After each simulation, conduct a brief post-mortem:

```bash
cat > ./postmortems/simulation-summary.md << 'EOF'
# Incident Simulation Summary

## Simulations Completed

### Simulation 1: High Error Rate
- **Type:** Intermittent service failures
- **Root Cause:** Container pause events
- **Key Learning:** Monitor container state changes

### Simulation 2: Node Down
- **Type:** Complete service outage
- **Root Cause:** Container stopped
- **Key Learning:** Implement automatic restart

### Simulation 3: VEID Scoring Anomaly
- **Type:** Logic/data issue
- **Root Cause:** Preprocessing bug
- **Key Learning:** Test data pipelines thoroughly

### Simulation 4: Deployment Failure
- **Type:** Resource constraint
- **Root Cause:** Restrictive quota
- **Key Learning:** Monitor quota utilization

## Overall Observations

### What Went Well
- Detection was quick using available monitoring
- Structured approach helped systematic diagnosis
- Documentation during incident aided resolution

### Areas for Improvement
- Need better automated alerting
- More runbooks for common scenarios
- Faster escalation paths

## Action Items
- [ ] Create alerting rules for each scenario type
- [ ] Document runbooks for common incidents
- [ ] Practice drills quarterly
- [ ] Improve monitoring dashboards
EOF

cat ./postmortems/simulation-summary.md
```

### Key Metrics to Review

```bash
cat > ./postmortems/metrics-review.md << 'EOF'
# Post-Incident Metrics Review

## Response Time Metrics

| Simulation | Detection | Triage | Resolution | Total |
|------------|-----------|--------|------------|-------|
| Sim 1      | 3 min     | 5 min  | 7 min      | 15 min|
| Sim 2      | 1 min     | 3 min  | 6 min      | 10 min|
| Sim 3      | 15 min    | 10 min | 25 min     | 50 min|
| Sim 4      | 5 min     | 5 min  | 10 min     | 20 min|

## Detection Methods

| Simulation | Detection Method |
|------------|------------------|
| Sim 1      | Error rate metric |
| Sim 2      | Health check failure |
| Sim 3      | Score distribution anomaly |
| Sim 4      | Deployment status |

## Root Cause Categories

| Category | Count |
|----------|-------|
| Container/Orchestration | 2 |
| Application Logic | 1 |
| Resource Constraints | 1 |
EOF

cat ./postmortems/metrics-review.md
```

---

## Summary

In this lab, you practiced:

1. **Chaos injection** - Safely introducing failures for training
2. **Detection** - Using monitoring to identify issues quickly
3. **Triage** - Assessing impact and severity appropriately
4. **Diagnosis** - Systematically finding root causes
5. **Remediation** - Fixing issues and verifying recovery
6. **Documentation** - Recording actions and lessons learned

### Key Takeaways

- Structured incident response reduces mean time to resolution
- Good monitoring enables fast detection
- Documentation during incidents is critical
- Post-mortems drive continuous improvement
- Practice builds muscle memory for real incidents

### Cleanup

```bash
cd ~/incident-lab

# Remove restrictive quota if exists
kubectl delete resourcequota restrictive-quota -n virtengine-workloads 2>/dev/null || true

# Ensure services are running
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh status
```

---

## Next Steps

Continue to:

1. **[Security Assessment Lab](security-assessment-lab.md)** - Security auditing

Additional resources:

- Review VirtEngine runbooks in `_docs/runbooks/`
- Practice with different failure scenarios
- Create custom chaos scenarios for your environment

---

*Lab Version: 1.0.0*  
*Last Updated: 2026-01-24*  
*Maintainer: VirtEngine Training Team*
