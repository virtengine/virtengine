# Provider Operations Lab

**Duration:** 4 hours  
**Prerequisites:** [Lab Environment Setup](lab-environment.md) completed  
**Difficulty:** Intermediate

---

## Table of Contents

1. [Overview](#overview)
2. [Learning Objectives](#learning-objectives)
3. [Exercise 1: Set Up Provider Daemon Locally](#exercise-1-set-up-provider-daemon-locally)
4. [Exercise 2: Configure Infrastructure Adapter (Kubernetes)](#exercise-2-configure-infrastructure-adapter-kubernetes)
5. [Exercise 3: Submit Bids and Create Leases](#exercise-3-submit-bids-and-create-leases)
6. [Exercise 4: Deploy a Workload and Monitor Usage](#exercise-4-deploy-a-workload-and-monitor-usage)
7. [Exercise 5: Handle Workload Lifecycle Transitions](#exercise-5-handle-workload-lifecycle-transitions)
8. [Troubleshooting](#troubleshooting)
9. [Summary](#summary)

---

## Overview

VirtEngine providers offer compute resources through the decentralized marketplace. Providers:

1. **Watch marketplace orders** - Monitor the chain for new deployment requests
2. **Submit competitive bids** - Offer resources at competitive prices
3. **Provision workloads** - Deploy customer containers/VMs on infrastructure
4. **Report usage** - Submit signed usage records for billing
5. **Handle lifecycle** - Manage workload transitions (start, pause, stop, terminate)

This lab covers the complete provider workflow from setup to workload management.

### Lab Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       Provider Architecture                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐      Watches       ┌─────────────────┐            │
│  │   Provider      │ ◄────────────────► │   VirtEngine    │            │
│  │   Daemon        │      Submits Bids  │   Chain         │            │
│  └────────┬────────┘      Reports Usage └─────────────────┘            │
│           │                                                              │
│           │  Provisions                                                  │
│           v                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Infrastructure Adapter                        │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │   │
│  │  │Kubernetes│  │  SLURM   │  │ OpenStack│  │  VMware  │        │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Workload Lifecycle States

```
Pending ──► Deploying ──► Running ──► Stopping ──► Stopped ──► Terminated
    │           │            │            │
    │           │            ▼            │
    │           │         Paused ─────────┘
    │           │
    └───────────┴───────────► Failed
```

---

## Learning Objectives

By the end of this lab, you will be able to:

- [ ] Set up and configure the provider daemon
- [ ] Configure the Kubernetes infrastructure adapter
- [ ] Participate in the bidding process
- [ ] Deploy and monitor workloads
- [ ] Handle workload lifecycle transitions
- [ ] Troubleshoot common provider issues

---

## Exercise 1: Set Up Provider Daemon Locally

### Objective
Install and configure the provider daemon to connect to the localnet.

### Duration
45 minutes

### Prerequisites
- Localnet running (`./scripts/localnet.sh status` shows healthy)
- VirtEngine binary built

### Instructions

#### Step 1: Create Provider Directory

```bash
# Create provider workspace
mkdir -p ~/provider-lab
cd ~/provider-lab

# Set binary paths
export VIRTENGINE_BIN="$HOME/virtengine-lab/virtengine/.cache/bin/virtengine"
export PROVIDER_DAEMON="$HOME/virtengine-lab/virtengine/.cache/bin/provider-daemon"

# Create provider home directory
mkdir -p ./provider-home/{keyring-file,config,logs}
```

#### Step 2: Create Provider Keys

```bash
# Create provider signing key
$VIRTENGINE_BIN keys add provider-operator \
    --keyring-backend file \
    --home ./provider-home

# Save the mnemonic securely!
PROVIDER_ADDR=$($VIRTENGINE_BIN keys show provider-operator -a --keyring-backend file --home ./provider-home)
echo "Provider address: $PROVIDER_ADDR"
echo $PROVIDER_ADDR > ./provider-address.txt

# Create encryption key for receiving order details
$VIRTENGINE_BIN keys add provider-encryption \
    --keyring-backend file \
    --home ./provider-home

ENCRYPTION_PUBKEY=$($VIRTENGINE_BIN keys show provider-encryption --pubkey --keyring-backend file --home ./provider-home)
echo "Encryption pubkey: $ENCRYPTION_PUBKEY"
```

#### Step 3: Fund Provider Account

```bash
# Enter localnet shell to fund the provider
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh shell

# Inside container
PROVIDER_ADDR="<paste-your-provider-address>"
virtengine tx bank send \
    validator $PROVIDER_ADDR 200000000000uve \
    --keyring-backend test \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

exit
```

#### Step 4: Verify Balance

```bash
cd ~/provider-lab

$VIRTENGINE_BIN query bank balances $PROVIDER_ADDR \
    --node http://localhost:26657
```

**Expected Output:**
```yaml
balances:
- amount: "200000000000"
  denom: uve
```

#### Step 5: Create Provider Configuration

```bash
cat > ./provider-home/config/config.yaml << 'EOF'
# Provider Daemon Configuration

# Provider identity
provider:
  address: ""  # Will be set from keyring
  name: "Lab-Provider-1"
  contact_email: "provider-lab@example.com"

# Chain connection
chain:
  node_url: "http://localhost:26657"
  grpc_url: "localhost:9090"
  chain_id: "virtengine-localnet-1"
  gas_prices: "0.025uve"
  gas_adjustment: 1.5

# Key management
keys:
  keyring_path: "./provider-home/keyring-file"
  keyring_backend: "file"
  signing_key: "provider-operator"
  encryption_key: "provider-encryption"

# Orchestration (Kubernetes for this lab)
orchestration:
  adapter: "kubernetes"
  kubernetes:
    kubeconfig: "${HOME}/.kube/config"
    namespace: "virtengine-workloads"
    resource_quota:
      cpu: "10"
      memory: "32Gi"
      gpu: "0"

# Bid engine
bidding:
  enabled: true
  strategy: "fixed"
  base_prices:
    cpu_core: 10       # uve per CPU core per hour
    memory_gb: 5       # uve per GB RAM per hour
    storage_gb: 1      # uve per GB storage per hour
    gpu_unit: 500      # uve per GPU per hour

# Workload management
workloads:
  max_concurrent: 10
  health_check_interval: "30s"
  eviction_timeout: "10m"

# Usage reporting
usage:
  report_interval: "5m"  # More frequent for lab
  batch_size: 10
  max_retries: 3

# Logging
logging:
  level: "debug"  # Verbose for lab
  format: "text"
  redact_sensitive: true

# Metrics
metrics:
  enabled: true
  port: 9091
  path: "/metrics"
EOF

echo "Configuration created at ./provider-home/config/config.yaml"
```

#### Step 6: Register Provider On-Chain

```bash
# Register as a provider
$VIRTENGINE_BIN tx provider create \
    --name "Lab-Provider-1" \
    --contact "provider-lab@example.com" \
    --website "https://lab-provider.example.com" \
    --deposit 100000000000uve \
    --from provider-operator \
    --keyring-backend file \
    --home ./provider-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

# Wait for confirmation
sleep 10

# Verify registration
$VIRTENGINE_BIN query provider list \
    --node http://localhost:26657 \
    --output json | jq '.providers'
```

**Expected Output:**
```json
[
  {
    "owner": "virtengine1...",
    "name": "Lab-Provider-1",
    "contact": "provider-lab@example.com",
    "website": "https://lab-provider.example.com",
    "status": "ACTIVE"
  }
]
```

#### Step 7: Register Encryption Key

```bash
# Register encryption key for receiving orders
$VIRTENGINE_BIN tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key "$ENCRYPTION_PUBKEY" \
    --label "Lab Provider Order Decryption Key" \
    --from provider-operator \
    --keyring-backend file \
    --home ./provider-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-prices 0.025uve \
    --yes
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Provider keys created | `keys list` shows provider-operator and provider-encryption |
| Provider funded | Balance shows 200000000000uve (or remaining after deposit) |
| Provider registered | Shows in `query provider list` |
| Encryption key registered | Shows in `query encryption recipient-keys` |

---

## Exercise 2: Configure Infrastructure Adapter (Kubernetes)

### Objective
Set up a local Kubernetes cluster and configure the provider to use it.

### Duration
45 minutes

### Instructions

#### Step 1: Install Kind (Kubernetes in Docker)

```bash
# Install kind
go install sigs.k8s.io/kind@latest

# Verify installation
kind version
```

#### Step 2: Create Kubernetes Cluster

```bash
# Create cluster configuration
cat > ./kind-config.yaml << 'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: virtengine-lab
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30000
        hostPort: 30000
        protocol: TCP
      - containerPort: 30001
        hostPort: 30001
        protocol: TCP
  - role: worker
  - role: worker
EOF

# Create cluster
kind create cluster --config ./kind-config.yaml

# Verify cluster
kubectl cluster-info
kubectl get nodes
```

**Expected Output:**
```
NAME                          STATUS   ROLES           AGE   VERSION
virtengine-lab-control-plane  Ready    control-plane   1m    v1.29.0
virtengine-lab-worker         Ready    <none>          1m    v1.29.0
virtengine-lab-worker2        Ready    <none>          1m    v1.29.0
```

#### Step 3: Create Namespace and RBAC

```bash
# Create namespace for workloads
kubectl create namespace virtengine-workloads

# Create service account for provider
cat > ./provider-rbac.yaml << 'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: provider-daemon
  namespace: virtengine-workloads
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: provider-daemon
rules:
  - apiGroups: [""]
    resources: ["namespaces", "pods", "services", "configmaps", "secrets", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["networkpolicies", "ingresses"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["pods/log", "pods/exec"]
    verbs: ["get", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: provider-daemon
subjects:
  - kind: ServiceAccount
    name: provider-daemon
    namespace: virtengine-workloads
roleRef:
  kind: ClusterRole
  name: provider-daemon
  apiGroup: rbac.authorization.k8s.io
EOF

kubectl apply -f ./provider-rbac.yaml
```

#### Step 4: Create Resource Quota

```bash
# Set resource limits for the namespace
cat > ./resource-quota.yaml << 'EOF'
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
    persistentvolumeclaims: "10"
    pods: "50"
EOF

kubectl apply -f ./resource-quota.yaml

# Verify quota
kubectl describe resourcequota -n virtengine-workloads
```

#### Step 5: Test Kubernetes Connectivity

```bash
# Create test deployment
cat > ./test-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-nginx
  namespace: virtengine-workloads
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-nginx
  template:
    metadata:
      labels:
        app: test-nginx
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: 80
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "200m"
              memory: "256Mi"
EOF

kubectl apply -f ./test-deployment.yaml

# Wait for deployment
kubectl wait --for=condition=available deployment/test-nginx -n virtengine-workloads --timeout=60s

# Verify
kubectl get pods -n virtengine-workloads
```

**Expected Output:**
```
NAME                          READY   STATUS    RESTARTS   AGE
test-nginx-abc123-xyz         1/1     Running   0          30s
```

#### Step 6: Cleanup Test Deployment

```bash
kubectl delete deployment test-nginx -n virtengine-workloads
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Kind cluster running | `kubectl cluster-info` returns API server URL |
| 3 nodes available | `kubectl get nodes` shows 3 Ready nodes |
| Namespace created | `kubectl get ns virtengine-workloads` exists |
| RBAC configured | Service account and roles created |
| Test deployment works | Pod runs successfully |

---

## Exercise 3: Submit Bids and Create Leases

### Objective
Understand and simulate the bidding process for marketplace orders.

### Duration
45 minutes

### Instructions

#### Step 1: Create a Test Deployment Order (as Customer)

First, create a deployment order from a customer perspective:

```bash
# Create deployment manifest
cat > ./deployment-manifest.yaml << 'EOF'
version: "2.0"
services:
  web:
    image: nginx:alpine
    expose:
      - port: 80
        to:
          - global: true
profiles:
  compute:
    web:
      resources:
        cpu:
          units: 0.5
        memory:
          size: 512Mi
        storage:
          size: 1Gi
  placement:
    dcloud:
      attributes:
        region: local
      pricing:
        web:
          denom: uve
          amount: 100
deployment:
  web:
    dcloud:
      profile: web
      count: 1
EOF
```

#### Step 2: Create Deployment Group

```bash
# Using localnet alice account to create deployment
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh shell

# Inside container - create deployment
virtengine tx deployment create /deployment-manifest.yaml \
    --from alice \
    --keyring-backend test \
    --chain-id virtengine-localnet-1 \
    --deposit 10000000uve \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

# Get deployment ID
DSEQ=$(virtengine query deployment list --owner $(virtengine keys show alice -a --keyring-backend test) --output json | jq -r '.deployments[0].deployment.deployment_id.dseq')
echo "Deployment sequence: $DSEQ"

exit
```

#### Step 3: Query Open Orders

```bash
cd ~/provider-lab

# Query orders awaiting bids
$VIRTENGINE_BIN query market order list \
    --state open \
    --node http://localhost:26657 \
    --output json | jq '.orders'
```

**Expected Output:**
```json
[
  {
    "order_id": {
      "owner": "virtengine1alice...",
      "dseq": "1",
      "gseq": 1,
      "oseq": 1
    },
    "state": "open",
    "spec": {
      "resources": [...]
    }
  }
]
```

#### Step 4: Submit a Bid

```bash
# Get order details
ORDER_OWNER="<alice-address-from-above>"
DSEQ="<dseq-from-above>"

# Submit bid
$VIRTENGINE_BIN tx market bid create \
    --owner $ORDER_OWNER \
    --dseq $DSEQ \
    --gseq 1 \
    --oseq 1 \
    --price 50uve \
    --deposit 5000000uve \
    --from provider-operator \
    --keyring-backend file \
    --home ./provider-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

echo "Bid submitted!"
```

#### Step 5: Check Bid Status

```bash
# Query bids for the order
$VIRTENGINE_BIN query market bid list \
    --owner $ORDER_OWNER \
    --dseq $DSEQ \
    --node http://localhost:26657 \
    --output json | jq '.bids'
```

**Expected Output:**
```json
[
  {
    "bid_id": {
      "owner": "virtengine1alice...",
      "dseq": "1",
      "gseq": 1,
      "oseq": 1,
      "provider": "virtengine1provider..."
    },
    "state": "open",
    "price": {
      "denom": "uve",
      "amount": "50"
    }
  }
]
```

#### Step 6: Accept Bid (as Customer)

```bash
# Back to localnet shell as alice
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh shell

# Accept the bid (creates lease)
PROVIDER_ADDR="<provider-address>"
virtengine tx market lease create \
    --owner $(virtengine keys show alice -a --keyring-backend test) \
    --dseq $DSEQ \
    --gseq 1 \
    --oseq 1 \
    --provider $PROVIDER_ADDR \
    --from alice \
    --keyring-backend test \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-prices 0.025uve \
    --yes

exit
```

#### Step 7: Verify Lease Created

```bash
cd ~/provider-lab

# Query leases
$VIRTENGINE_BIN query market lease list \
    --provider $PROVIDER_ADDR \
    --node http://localhost:26657 \
    --output json | jq '.leases'
```

**Expected Output:**
```json
[
  {
    "lease_id": {
      "owner": "virtengine1alice...",
      "dseq": "1",
      "gseq": 1,
      "oseq": 1,
      "provider": "virtengine1provider..."
    },
    "state": "active",
    "price": {
      "denom": "uve",
      "amount": "50"
    }
  }
]
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Deployment created | Shows in `query deployment list` |
| Order open | Shows in `query market order list --state open` |
| Bid submitted | Shows in `query market bid list` |
| Lease active | Shows in `query market lease list` with state "active" |

---

## Exercise 4: Deploy a Workload and Monitor Usage

### Objective
Deploy the customer's workload and set up usage monitoring.

### Duration
45 minutes

### Instructions

#### Step 1: Process Lease and Deploy Workload

The provider daemon would normally do this automatically. For learning, we'll do it manually:

```bash
# Get lease details
LEASE_OWNER=$ORDER_OWNER
LEASE_DSEQ=$DSEQ
LEASE_GSEQ=1
LEASE_OSEQ=1

# Create Kubernetes deployment for the lease
cat > ./workload-deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ}
  namespace: virtengine-workloads
  labels:
    virtengine.com/owner: "$LEASE_OWNER"
    virtengine.com/dseq: "$LEASE_DSEQ"
    virtengine.com/gseq: "$LEASE_GSEQ"
    virtengine.com/oseq: "$LEASE_OSEQ"
spec:
  replicas: 1
  selector:
    matchLabels:
      virtengine.com/dseq: "$LEASE_DSEQ"
  template:
    metadata:
      labels:
        virtengine.com/owner: "$LEASE_OWNER"
        virtengine.com/dseq: "$LEASE_DSEQ"
        virtengine.com/gseq: "$LEASE_GSEQ"
        virtengine.com/oseq: "$LEASE_OSEQ"
    spec:
      containers:
        - name: web
          image: nginx:alpine
          ports:
            - containerPort: 80
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
---
apiVersion: v1
kind: Service
metadata:
  name: ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ}
  namespace: virtengine-workloads
spec:
  type: NodePort
  selector:
    virtengine.com/dseq: "$LEASE_DSEQ"
  ports:
    - port: 80
      targetPort: 80
      nodePort: 30000
EOF

# Apply deployment
kubectl apply -f ./workload-deployment.yaml

# Wait for deployment
kubectl wait --for=condition=available deployment/ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    -n virtengine-workloads --timeout=120s
```

#### Step 2: Verify Workload Running

```bash
# Check pod status
kubectl get pods -n virtengine-workloads -l virtengine.com/dseq=$LEASE_DSEQ

# Check service
kubectl get svc -n virtengine-workloads

# Test the workload
curl http://localhost:30000
```

**Expected Output:**
```html
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
...
```

#### Step 3: Set Up Usage Monitoring Script

```bash
cat > ./usage-monitor.sh << 'EOF'
#!/bin/bash

# Usage monitoring script for provider workloads
NAMESPACE="virtengine-workloads"

echo "═══════════════════════════════════════════════════════════════"
echo "              Provider Workload Usage Monitor                   "
echo "═══════════════════════════════════════════════════════════════"

while true; do
    clear
    echo "═══════════════════════════════════════════════════════════════"
    echo "              Provider Workload Usage Monitor                   "
    echo "              $(date '+%Y-%m-%d %H:%M:%S')                      "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    # Get all VirtEngine workloads
    PODS=$(kubectl get pods -n $NAMESPACE -l virtengine.com/dseq -o json)
    
    POD_COUNT=$(echo $PODS | jq '.items | length')
    echo "Active Workloads: $POD_COUNT"
    echo ""
    
    # Per-pod metrics
    echo "┌─────────────────────────────────────────────────────────────┐"
    printf "│ %-20s │ %-8s │ %-8s │ %-10s │\n" "POD" "CPU" "MEMORY" "STATUS"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    for row in $(echo $PODS | jq -r '.items[] | @base64'); do
        _jq() {
            echo ${row} | base64 --decode | jq -r ${1}
        }
        
        POD_NAME=$(_jq '.metadata.name' | cut -c1-20)
        STATUS=$(_jq '.status.phase')
        
        # Get metrics (if metrics-server available)
        METRICS=$(kubectl top pod $(_jq '.metadata.name') -n $NAMESPACE 2>/dev/null || echo "N/A N/A")
        CPU=$(echo $METRICS | awk '{print $2}')
        MEM=$(echo $METRICS | awk '{print $3}')
        
        printf "│ %-20s │ %-8s │ %-8s │ %-10s │\n" "$POD_NAME" "$CPU" "$MEM" "$STATUS"
    done
    
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Resource quota usage
    echo "Resource Quota Usage:"
    kubectl describe resourcequota provider-quota -n $NAMESPACE 2>/dev/null | grep -E "(Used|Hard)" | head -10
    
    echo ""
    echo "Press Ctrl+C to exit"
    
    sleep 10
done
EOF

chmod +x ./usage-monitor.sh
```

#### Step 4: Run Usage Monitor

```bash
./usage-monitor.sh
```

#### Step 5: Simulate Usage Record Submission

```bash
# Create usage record structure
cat > ./usage-record.json << EOF
{
  "lease_id": {
    "owner": "$LEASE_OWNER",
    "dseq": "$LEASE_DSEQ",
    "gseq": $LEASE_GSEQ,
    "oseq": $LEASE_OSEQ,
    "provider": "$PROVIDER_ADDR"
  },
  "start_time": "$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)",
  "end_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "resources": {
    "cpu_milli_seconds": 1800000,
    "memory_byte_seconds": 1932735283200,
    "storage_byte_seconds": 3865470566400
  }
}
EOF

cat ./usage-record.json

# In production, provider daemon would submit this on-chain
# $VIRTENGINE_BIN tx escrow submit-usage ./usage-record.json ...
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Deployment created in K8s | `kubectl get deployments` shows workload |
| Pod running | `kubectl get pods` shows Running status |
| Service accessible | `curl localhost:30000` returns nginx page |
| Usage script works | Shows pod metrics |

---

## Exercise 5: Handle Workload Lifecycle Transitions

### Objective
Practice managing workloads through their lifecycle states.

### Duration
45 minutes

### Instructions

#### Step 1: Understand Lifecycle States

Valid state transitions:

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Workload Lifecycle States                          │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  PENDING ───────► DEPLOYING ───────► RUNNING ───────► STOPPING       │
│     │                 │                 │                  │          │
│     │                 │                 ▼                  ▼          │
│     │                 │              PAUSED ◄──────►   STOPPED        │
│     │                 │                                    │          │
│     ▼                 ▼                                    ▼          │
│   FAILED ◄───────────────────────────────────────────  TERMINATED    │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

#### Step 2: Pause a Workload

```bash
# Scale to 0 (pause)
kubectl scale deployment ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    --replicas=0 -n virtengine-workloads

# Verify paused
kubectl get deployment ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    -n virtengine-workloads

# Check no pods running
kubectl get pods -n virtengine-workloads -l virtengine.com/dseq=$LEASE_DSEQ
```

**Expected Output:**
```
NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
ve-1-1-1                          0/0     0            0           10m

No resources found in virtengine-workloads namespace.
```

#### Step 3: Resume a Workload

```bash
# Scale back up (resume)
kubectl scale deployment ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    --replicas=1 -n virtengine-workloads

# Wait for pod
kubectl wait --for=condition=available \
    deployment/ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    -n virtengine-workloads --timeout=60s

# Verify running
kubectl get pods -n virtengine-workloads -l virtengine.com/dseq=$LEASE_DSEQ
```

#### Step 4: Handle Workload Failure

```bash
# Simulate a failing deployment
cat > ./failing-workload.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: failing-workload
  namespace: virtengine-workloads
  labels:
    virtengine.com/test: "true"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: failing
  template:
    metadata:
      labels:
        app: failing
    spec:
      containers:
        - name: failing
          image: busybox
          command: ["sh", "-c", "exit 1"]
          resources:
            requests:
              cpu: "100m"
              memory: "64Mi"
EOF

kubectl apply -f ./failing-workload.yaml

# Watch failure
kubectl get pods -n virtengine-workloads -l app=failing -w
# Press Ctrl+C after seeing CrashLoopBackOff
```

#### Step 5: Diagnose Failed Workload

```bash
# Get pod details
kubectl describe pod -n virtengine-workloads -l app=failing

# Check logs
kubectl logs -n virtengine-workloads -l app=failing --previous

# Get events
kubectl get events -n virtengine-workloads --sort-by='.lastTimestamp' | tail -10
```

#### Step 6: Cleanup Failed Workload

```bash
# Delete failed deployment
kubectl delete deployment failing-workload -n virtengine-workloads

# Verify cleanup
kubectl get deployments -n virtengine-workloads
```

#### Step 7: Terminate Workload (Lease End)

```bash
# Terminate the main workload
kubectl delete deployment ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    -n virtengine-workloads
kubectl delete service ve-${LEASE_DSEQ}-${LEASE_GSEQ}-${LEASE_OSEQ} \
    -n virtengine-workloads

# Verify termination
kubectl get all -n virtengine-workloads

# Close the lease on-chain
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh shell

# Inside container
virtengine tx market lease close \
    --owner $(virtengine keys show alice -a --keyring-backend test) \
    --dseq $DSEQ \
    --gseq 1 \
    --oseq 1 \
    --provider $PROVIDER_ADDR \
    --from alice \
    --keyring-backend test \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-prices 0.025uve \
    --yes

exit
```

#### Step 8: Create Lifecycle Transition Script

```bash
cat > ./workload-lifecycle.sh << 'EOF'
#!/bin/bash

# Workload lifecycle management script
NAMESPACE="virtengine-workloads"

usage() {
    echo "Usage: $0 <command> <deployment-name>"
    echo ""
    echo "Commands:"
    echo "  status    - Show workload status"
    echo "  pause     - Pause workload (scale to 0)"
    echo "  resume    - Resume workload (scale to 1)"
    echo "  stop      - Gracefully stop workload"
    echo "  terminate - Delete workload completely"
    echo "  logs      - Show workload logs"
}

case "$1" in
    status)
        echo "=== Deployment Status ==="
        kubectl get deployment "$2" -n $NAMESPACE
        echo ""
        echo "=== Pod Status ==="
        kubectl get pods -n $NAMESPACE -l app="$2"
        ;;
    pause)
        echo "Pausing $2..."
        kubectl scale deployment "$2" --replicas=0 -n $NAMESPACE
        echo "Workload paused."
        ;;
    resume)
        echo "Resuming $2..."
        kubectl scale deployment "$2" --replicas=1 -n $NAMESPACE
        kubectl wait --for=condition=available deployment/"$2" -n $NAMESPACE --timeout=60s
        echo "Workload resumed."
        ;;
    stop)
        echo "Stopping $2..."
        kubectl scale deployment "$2" --replicas=0 -n $NAMESPACE
        echo "Workload stopped."
        ;;
    terminate)
        echo "Terminating $2..."
        kubectl delete deployment "$2" -n $NAMESPACE
        kubectl delete service "$2" -n $NAMESPACE 2>/dev/null || true
        echo "Workload terminated."
        ;;
    logs)
        kubectl logs -n $NAMESPACE -l app="$2" --tail=100 -f
        ;;
    *)
        usage
        exit 1
        ;;
esac
EOF

chmod +x ./workload-lifecycle.sh
echo "Lifecycle script created: ./workload-lifecycle.sh"
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Pause works | Deployment shows 0/0 replicas |
| Resume works | Pod returns to Running state |
| Failure diagnosis | Events and logs retrieved |
| Termination complete | Resources deleted from K8s |
| Lease closed | Lease state shows "closed" |

---

## Troubleshooting

### Provider Not Receiving Orders

**Symptoms:** No bids being submitted automatically

```bash
# Check provider daemon logs
tail -f ./provider-home/logs/provider.log

# Verify chain connectivity
curl http://localhost:26657/status

# Check provider status
$VIRTENGINE_BIN query provider list --node http://localhost:26657
```

### Kubernetes Deployment Failures

**Symptoms:** Pods stuck in Pending or CrashLoopBackOff

```bash
# Check pod events
kubectl describe pod <pod-name> -n virtengine-workloads

# Check resource availability
kubectl describe nodes

# Check namespace quota
kubectl describe resourcequota -n virtengine-workloads
```

### Bid Not Winning

**Symptoms:** Bids submitted but lease created with different provider

```bash
# Check bid price vs competition
$VIRTENGINE_BIN query market bid list --dseq $DSEQ --node http://localhost:26657

# Verify bid is in open state
# Lower price may be needed
```

### Usage Submission Failures

**Symptoms:** Usage records rejected

```bash
# Check escrow balance
$VIRTENGINE_BIN query escrow account --owner $LEASE_OWNER --provider $PROVIDER_ADDR

# Verify lease is still active
$VIRTENGINE_BIN query market lease list --provider $PROVIDER_ADDR
```

### Kind Cluster Issues

**Symptoms:** kubectl commands fail

```bash
# Check cluster status
kind get clusters

# Recreate if needed
kind delete cluster --name virtengine-lab
kind create cluster --config ./kind-config.yaml

# Update kubeconfig
kind get kubeconfig --name virtengine-lab > ~/.kube/config
```

---

## Summary

In this lab, you learned how to:

1. **Set up provider daemon** - Configure keys, chain connection, and registration
2. **Configure Kubernetes adapter** - Create cluster, namespace, RBAC, and quotas
3. **Participate in bidding** - Watch orders, submit bids, create leases
4. **Deploy workloads** - Provision customer containers and monitor usage
5. **Manage lifecycle** - Handle pause, resume, stop, and terminate operations

### Key Takeaways

- Providers must register on-chain before submitting bids
- Encryption keys enable secure receipt of order details
- The Kubernetes adapter requires proper RBAC configuration
- Usage monitoring is critical for accurate billing
- Lifecycle management ensures clean resource handling

### Cleanup

```bash
# Delete Kind cluster
kind delete cluster --name virtengine-lab

# Stop provider (if running as daemon)
pkill -f provider-daemon

# Optional: Remove provider directory
# rm -rf ~/provider-lab
```

---

## Next Steps

Continue to the following labs:

1. **[Incident Simulation Lab](incident-simulation-lab.md)** - Practice incident response
2. **[Security Assessment Lab](security-assessment-lab.md)** - Security auditing

---

*Lab Version: 1.0.0*  
*Last Updated: 2026-01-24*  
*Maintainer: VirtEngine Training Team*
