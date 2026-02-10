# Provider Daemon Operator Training

**Module Duration:** 16 hours (4 weeks, 4 hours/week)  
**Prerequisites:** Basic understanding of VirtEngine blockchain, Linux administration, container orchestration  
**Certification:** VirtEngine Certified Provider Operator (VCPO)

---

## Table of Contents

1. [Course Overview](#course-overview)
2. [Week 1: Provider Architecture & Daemon Components](#week-1-provider-architecture--daemon-components)
3. [Week 2: Infrastructure Adapters](#week-2-infrastructure-adapters)
4. [Week 3: Workload Lifecycle Management](#week-3-workload-lifecycle-management)
5. [Week 4: Monitoring, Troubleshooting & Maintenance](#week-4-monitoring-troubleshooting--maintenance)
6. [Assessments & Certification](#assessments--certification)

---

## Course Overview

### Learning Objectives

Upon completion of this training, operators will be able to:

- [ ] Configure and deploy the VirtEngine provider daemon
- [ ] Manage provider keys using hardware, ledger, or file-based storage
- [ ] Configure and optimize the bid engine for marketplace participation
- [ ] Deploy and manage workloads across multiple infrastructure backends
- [ ] Implement usage metering and on-chain settlement
- [ ] Troubleshoot common issues and perform maintenance operations

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        VirtEngine Provider Daemon                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │   Bid Engine    │  │   Key Manager   │  │   Usage Meter   │         │
│  │                 │  │                 │  │                 │         │
│  │ • Order Watch   │  │ • File Storage  │  │ • Metrics       │         │
│  │ • Price Calc    │  │ • HSM Support   │  │ • Chain Submit  │         │
│  │ • Rate Limit    │  │ • Ledger        │  │ • Fraud Check   │         │
│  │ • Streaming     │  │ • Rotation      │  │                 │         │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘         │
│           │                    │                    │                   │
│  ┌────────┴────────────────────┴────────────────────┴────────┐         │
│  │                    Chain Client Interface                   │         │
│  └─────────────────────────────┬───────────────────────────────┘         │
│                                │                                         │
├────────────────────────────────┼─────────────────────────────────────────┤
│                                │                                         │
│  ┌─────────────────────────────┴───────────────────────────────┐         │
│  │                  Infrastructure Adapters                     │         │
│  ├─────────┬─────────┬─────────┬─────────┬─────────┬───────────┤         │
│  │   K8s   │OpenStack│   AWS   │  Azure  │ VMware  │  Ansible  │         │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴───────────┘         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Week 1: Provider Architecture & Daemon Components

**Duration:** 4 hours  
**Focus:** Understanding core components and configuration

### Day 1: Provider Daemon Overview (2 hours)

#### Learning Objectives
- Understand the role of the provider daemon in VirtEngine
- Learn the relationship between on-chain and off-chain components
- Configure basic provider daemon settings

#### Core Concepts

The provider daemon (`pkg/provider_daemon/`) is the off-chain service that bridges on-chain marketplace orders to infrastructure provisioning. It consists of three core components:

| Component | File | Primary Function |
|-----------|------|------------------|
| Bid Engine | `bid_engine.go` | Watches orders, computes pricing, submits bids |
| Key Manager | `key_manager.go` | Provider key storage with hardware support |
| Usage Meter | `usage_meter.go` | Collects metrics, submits on-chain usage records |

#### Bid Engine Configuration

```go
// BidEngineConfig configures the bid engine
type BidEngineConfig struct {
    ProviderAddress    string        // Provider's blockchain address
    MaxBidsPerMinute   int           // Rate limit per minute (default: 10)
    MaxBidsPerHour     int           // Rate limit per hour (default: 100)
    MaxConcurrentBids  int           // Concurrent bid operations (default: 5)
    BidRetryDelay      time.Duration // Retry delay (default: 5s)
    MaxBidRetries      int           // Max retries (default: 3)
    ConfigPollInterval time.Duration // Config poll interval (default: 30s)
    OrderPollInterval  time.Duration // Order poll interval (default: 5s)
}
```

**Configuration File Example (`provider-config.yaml`):**

```yaml
provider:
  address: "virtengine1provider..."
  
bid_engine:
  max_bids_per_minute: 10
  max_bids_per_hour: 100
  max_concurrent_bids: 5
  bid_retry_delay: "5s"
  max_bid_retries: 3
  config_poll_interval: "30s"
  order_poll_interval: "5s"

pricing:
  cpu_price_per_core: "0.001"
  memory_price_per_gb: "0.0005"
  storage_price_per_gb: "0.0001"
  gpu_price_per_hour: "0.05"
  min_bid_price: "0.0001"
  bid_markup_percent: 10.0
  currency: "VIRT"

capacity:
  total_cpu_cores: 128
  total_memory_gb: 512
  total_storage_gb: 10000
  total_gpus: 8
  reserved_cpu_cores: 4
  reserved_memory_gb: 16
  reserved_storage_gb: 100
```

#### Hands-On Exercise 1.1: Basic Configuration

**Objective:** Configure a basic provider daemon instance

**Steps:**
1. Create configuration directory:
   ```bash
   mkdir -p /etc/virtengine/provider
   ```

2. Create the configuration file with your provider address
3. Validate configuration syntax
4. Start the daemon in dry-run mode

**Verification Checklist:**
- [ ] Configuration file created and validated
- [ ] Provider address matches on-chain registration
- [ ] Capacity settings reflect actual infrastructure
- [ ] Pricing configuration is competitive

---

### Day 2: Key Manager & Security (2 hours)

#### Learning Objectives
- Configure key storage (file, HSM, Ledger)
- Implement key rotation policies
- Understand key lifecycle management

#### Key Storage Types

```go
const (
    KeyStorageTypeFile        = "file"         // Encrypted file storage
    KeyStorageTypeHardware    = "hardware"     // HSM (PKCS#11)
    KeyStorageTypeLedger      = "ledger"       // Ledger hardware wallet
    KeyStorageTypeNonCustodial = "non_custodial" // External signing
    KeyStorageTypeMemory      = "memory"       // Testing only
)
```

#### Key Manager Configuration

```yaml
key_manager:
  storage_type: "file"  # file, hardware, ledger, non_custodial
  key_dir: "/var/lib/virtengine/keys"
  default_algorithm: "ed25519"
  key_rotation_days: 90
  grace_period_hours: 24
  
  # HSM Configuration (optional)
  hsm:
    library_path: "/usr/lib/softhsm/libsofthsm2.so"
    slot_id: 0
    token_label: "virtengine-provider"
  
  # Ledger Configuration (optional)
  ledger:
    derivation_path: "m/44'/118'/0'/0/0"
    require_confirmation: true
```

#### Key Lifecycle States

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌───────────┐
│ Created  │────▶│ Pending  │────▶│  Active  │────▶│ Rotating  │
└──────────┘     └──────────┘     └──────────┘     └───────────┘
                                        │                │
                                        │                │
                                        ▼                ▼
                                  ┌──────────┐     ┌──────────────┐
                                  │Suspended │     │ Deactivated  │
                                  └──────────┘     └──────────────┘
                                        │                │
                                        │                ▼
                                        │          ┌──────────┐
                                        └─────────▶│ Archived │
                                                   └──────────┘
                                                         │
                                                         ▼
                                  ┌────────────┐   ┌───────────┐
                                  │Compromised │──▶│ Destroyed │
                                  └────────────┘   └───────────┘
```

#### Key Rotation Policy

```go
// Default key lifecycle policy
type KeyLifecyclePolicy struct {
    Name                         string   // Policy name
    MaxActiveAgeDays             int      // 90 days default
    RotationGracePeriodDays      int      // 7 days
    ExpirationDays               int      // 365 days
    ArchiveAfterDeactivationDays int      // 30 days
    DestroyAfterArchiveDays      int      // 365 days
    RequireApprovalForActivation bool     // false
    RequireApprovalForDestruction bool    // true
    AutoRotate                   bool     // true
    NotifyBeforeExpirationDays   int      // 30 days
    AllowedKeyTypes              []string // ed25519, secp256k1, p256
    MinimumKeyStrength           int      // 256 bits
}
```

#### Hands-On Exercise 1.2: Key Management Setup

**Objective:** Set up secure key management with rotation

**Steps:**
1. Initialize key storage:
   ```bash
   virtengine-provider keys init --storage-type file --key-dir /var/lib/virtengine/keys
   ```

2. Generate provider signing key:
   ```bash
   virtengine-provider keys generate --algorithm ed25519
   ```

3. Configure rotation policy:
   ```bash
   virtengine-provider keys set-policy --max-age-days 90 --auto-rotate
   ```

4. Test key signing:
   ```bash
   virtengine-provider keys test-sign
   ```

**Security Checklist:**
- [ ] Key directory has restricted permissions (700)
- [ ] Key files are encrypted at rest
- [ ] Backup of encryption passphrase is secured
- [ ] Key rotation schedule is configured
- [ ] Monitoring for key expiration is enabled

---

## Week 2: Infrastructure Adapters

**Duration:** 4 hours  
**Focus:** Configuring and managing infrastructure backends

*See [infrastructure-adapters.md](infrastructure-adapters.md) for detailed adapter training*

### Day 3: Kubernetes & Container Adapters (2 hours)

#### Kubernetes Adapter Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 Kubernetes Adapter                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Workload Request                                        │
│       │                                                  │
│       ▼                                                  │
│  ┌─────────────────┐                                    │
│  │ Namespace       │ ◀─── Isolated per deployment       │
│  │ Creation        │                                    │
│  └────────┬────────┘                                    │
│           │                                              │
│           ▼                                              │
│  ┌─────────────────┐     ┌─────────────────┐           │
│  │ Pod Deployment  │────▶│ Service/Ingress │           │
│  └─────────────────┘     └─────────────────┘           │
│           │                      │                      │
│           ▼                      ▼                      │
│  ┌─────────────────┐     ┌─────────────────┐           │
│  │ Resource Quota  │     │ Network Policy  │           │
│  └─────────────────┘     └─────────────────┘           │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

#### Configuration

```yaml
adapters:
  kubernetes:
    enabled: true
    kubeconfig: "/etc/virtengine/kubeconfig"
    namespace_prefix: "virt-"
    resource_limits:
      max_pods_per_deployment: 10
      max_cpu_per_pod: "8"
      max_memory_per_pod: "32Gi"
    network_policies:
      enabled: true
      default_deny_ingress: true
    storage_classes:
      - name: "standard"
        default: true
      - name: "fast-ssd"
```

#### Hands-On Exercise 2.1: Kubernetes Deployment

**Objective:** Deploy a workload through the Kubernetes adapter

1. Verify cluster connectivity
2. Deploy test workload
3. Monitor pod status
4. Verify resource allocation
5. Test network isolation

---

### Day 4: Cloud & VM Adapters (2 hours)

#### Waldur Bridge Architecture

The Waldur bridge provides a unified interface for cloud providers:

```
┌─────────────────────────────────────────────────────────────┐
│                     Waldur Bridge                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐                                           │
│  │ VirtEngine   │                                           │
│  │ Provider     │                                           │
│  └──────┬───────┘                                           │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────────┐                                       │
│  │  Waldur Bridge   │                                       │
│  │  (waldur_bridge) │                                       │
│  └──────┬───────────┘                                       │
│         │                                                    │
│    ┌────┴────┬────────┬────────┬────────┐                  │
│    │         │        │        │        │                   │
│    ▼         ▼        ▼        ▼        ▼                   │
│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐              │
│ │OpenSt│ │ AWS  │ │Azure │ │VMware│ │GCP   │              │
│ │ack   │ │      │ │      │ │      │ │      │              │
│ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### Multi-Cloud Configuration

```yaml
adapters:
  waldur:
    api_url: "https://waldur.example.com/api/"
    token: "${WALDUR_TOKEN}"
    
  openstack:
    enabled: true
    project_uuid: "project-uuid"
    default_flavor: "m1.medium"
    default_image: "ubuntu-22.04"
    networks:
      - name: "public"
        external: true
      - name: "private"
        
  aws:
    enabled: true
    region: "us-east-1"
    vpc_id: "vpc-12345"
    default_instance_type: "t3.medium"
    default_ami: "ami-12345"
    
  azure:
    enabled: true
    subscription_id: "subscription-uuid"
    resource_group: "virtengine-rg"
    location: "eastus"
    default_vm_size: "Standard_B2s"
```

---

## Week 3: Workload Lifecycle Management

**Duration:** 4 hours  
**Focus:** Managing workloads through their complete lifecycle

### Day 5: Lifecycle States & Transitions (2 hours)

#### Workload State Machine

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Workload Lifecycle States                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────┐                                                       │
│   │ PENDING │ ◀─── Initial state after order match                  │
│   └────┬────┘                                                       │
│        │                                                             │
│        ▼                                                             │
│   ┌───────────┐                                                     │
│   │ DEPLOYING │ ◀─── Infrastructure provisioning                    │
│   └─────┬─────┘                                                     │
│         │                                                            │
│    ┌────┴────┐                                                      │
│    │         │                                                       │
│    ▼         ▼                                                       │
│ ┌───────┐ ┌────────┐                                                │
│ │RUNNING│ │ FAILED │────────────────────────────────┐               │
│ └───┬───┘ └────────┘                                │               │
│     │                                               │               │
│     ├──────────────────┐                            │               │
│     │                  │                            │               │
│     ▼                  ▼                            │               │
│ ┌────────┐        ┌──────────┐                      │               │
│ │ PAUSED │◀──────▶│ STOPPING │                      │               │
│ └────────┘        └────┬─────┘                      │               │
│                        │                            │               │
│                        ▼                            │               │
│                   ┌─────────┐                       │               │
│                   │ STOPPED │◀──────────────────────┘               │
│                   └────┬────┘                                       │
│                        │                                             │
│                        ▼                                             │
│                  ┌────────────┐                                     │
│                  │ TERMINATED │ ◀─── Final state, resources freed   │
│                  └────────────┘                                     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### Valid State Transitions

```go
var validTransitions = map[WorkloadState][]WorkloadState{
    WorkloadStatePending:    {WorkloadStateDeploying, WorkloadStateFailed},
    WorkloadStateDeploying:  {WorkloadStateRunning, WorkloadStateFailed, WorkloadStateStopped},
    WorkloadStateRunning:    {WorkloadStatePaused, WorkloadStateStopping, WorkloadStateFailed},
    WorkloadStatePaused:     {WorkloadStateRunning, WorkloadStateStopping, WorkloadStateFailed},
    WorkloadStateStopping:   {WorkloadStateStopped, WorkloadStateFailed},
    WorkloadStateStopped:    {WorkloadStateTerminated, WorkloadStateDeploying},
    WorkloadStateFailed:     {WorkloadStateTerminated, WorkloadStateDeploying},
    WorkloadStateTerminated: {}, // Terminal state - no transitions
}
```

#### Deployed Workload Structure

```go
type DeployedWorkload struct {
    ID            string           // Unique workload identifier
    DeploymentID  string           // On-chain deployment ID
    LeaseID       string           // On-chain lease ID
    Namespace     string           // Infrastructure namespace
    State         WorkloadState    // Current lifecycle state
    Manifest      *Manifest        // Deployment manifest
    CreatedAt     time.Time        // Creation timestamp
    UpdatedAt     time.Time        // Last update timestamp
    StatusMessage string           // Status details
    Resources     []DeployedResource // Deployed resources
    Endpoints     []WorkloadEndpoint // Exposed endpoints
}
```

#### Hands-On Exercise 3.1: Lifecycle Management

**Objective:** Practice workload lifecycle operations

**Scenario:** Deploy, pause, resume, and terminate a workload

```bash
# 1. Deploy workload
virtengine-provider workload deploy --manifest workload.yaml

# 2. Monitor deployment
virtengine-provider workload status <workload-id>

# 3. Pause workload
virtengine-provider workload pause <workload-id>

# 4. Resume workload
virtengine-provider workload resume <workload-id>

# 5. Stop workload
virtengine-provider workload stop <workload-id>

# 6. Terminate workload
virtengine-provider workload terminate <workload-id>
```

**Lifecycle Verification Checklist:**
- [ ] Workload deploys successfully to RUNNING state
- [ ] Pause operation suspends resource consumption
- [ ] Resume operation restores workload functionality
- [ ] Stop operation gracefully shuts down workload
- [ ] Terminate operation releases all resources

---

### Day 6: Manifest Validation & Deployment (2 hours)

#### Manifest Structure

The provider daemon supports v1 and v2beta1 manifests:

```yaml
# v2beta1 Manifest Example
version: "v2beta1"
services:
  web:
    image: nginx:latest
    count: 3
    resources:
      cpu:
        units: 1.0
      memory:
        size: 512Mi
      storage:
        size: 1Gi
    expose:
      - port: 80
        as: 80
        proto: tcp
        to:
          - global: true

profiles:
  compute:
    web:
      resources:
        cpu:
          units: 1.0
        memory:
          size: 512Mi
        storage:
          size: 1Gi
  placement:
    default:
      attributes:
        region: us-east
      pricing:
        web:
          denom: uvirt
          amount: 1000

deployment:
  web:
    default:
      profile: web
      count: 3
```

#### Manifest Validation

```go
// Validate manifest before deployment
func (m *Manifest) Validate() error {
    // 1. Version check
    if m.Version != "v1" && m.Version != "v2beta1" {
        return ErrInvalidManifestVersion
    }
    
    // 2. Service validation
    for name, service := range m.Services {
        if err := validateService(name, service); err != nil {
            return err
        }
    }
    
    // 3. Resource validation
    if err := m.validateResources(); err != nil {
        return err
    }
    
    // 4. Exposure rules
    if err := m.validateExposure(); err != nil {
        return err
    }
    
    return nil
}
```

#### Hands-On Exercise 3.2: Manifest Validation

**Objective:** Create and validate deployment manifests

1. Create a multi-service manifest
2. Validate using the provider daemon CLI
3. Identify and fix validation errors
4. Deploy validated manifest

---

## Week 4: Monitoring, Troubleshooting & Maintenance

**Duration:** 4 hours  
**Focus:** Operational excellence and incident response

### Day 7: Monitoring & Usage Metering (2 hours)

#### Usage Meter Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Usage Meter                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                   Metrics Collector                            │  │
│  │                                                                │  │
│  │   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐      │  │
│  │   │   CPU   │   │ Memory  │   │ Storage │   │   GPU   │      │  │
│  │   └────┬────┘   └────┬────┘   └────┬────┘   └────┬────┘      │  │
│  │        └─────────────┴─────────────┴─────────────┘            │  │
│  └───────────────────────────────┬───────────────────────────────┘  │
│                                  │                                   │
│                                  ▼                                   │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    ResourceMetrics                             │  │
│  │                                                                │  │
│  │   CPUMilliSeconds    : int64  // CPU usage in milliseconds    │  │
│  │   MemoryByteSeconds  : int64  // Memory in byte-seconds       │  │
│  │   StorageByteSeconds : int64  // Storage in byte-seconds      │  │
│  │   NetworkBytesIn     : int64  // Inbound network bytes        │  │
│  │   NetworkBytesOut    : int64  // Outbound network bytes       │  │
│  │   GPUSeconds         : int64  // GPU usage in seconds         │  │
│  │                                                                │  │
│  └───────────────────────────────┬───────────────────────────────┘  │
│                                  │                                   │
│                                  ▼                                   │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    Chain Submission                            │  │
│  │                                                                │  │
│  │   • Sign usage record with provider key                       │  │
│  │   • Submit to blockchain for settlement                       │  │
│  │   • Trigger escrow release                                    │  │
│  │                                                                │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### Usage Record Structure

```go
type UsageRecord struct {
    ID            string          // Record ID
    WorkloadID    string          // Workload ID
    DeploymentID  string          // On-chain deployment ID
    LeaseID       string          // On-chain lease ID
    ProviderID    string          // Provider ID
    Type          UsageRecordType // periodic or final
    StartTime     time.Time       // Metering period start
    EndTime       time.Time       // Metering period end
    Metrics       ResourceMetrics // Resource usage
    PricingInputs PricingInputs   // Pricing configuration
    Signature     string          // Provider signature
    CreatedAt     time.Time       // Creation timestamp
}
```

#### Metering Configuration

```yaml
metering:
  interval: "1h"           # Metering interval (1m, 1h, 24h)
  provider_id: "provider1"
  
  alerts:
    high_cpu_threshold: 90     # Alert at 90% CPU
    high_memory_threshold: 85  # Alert at 85% memory
    low_storage_threshold: 10  # Alert at 10% storage remaining
```

#### Fraud Detection

```go
type FraudChecker struct {
    maxCPUUsageRatio       float64       // Max 2.0x allocated
    maxMemoryUsageRatio    float64       // Max 1.5x allocated
    maxNetworkAnomalyRatio float64       // Max 10.0x normal
    minRecordDuration      time.Duration // Min 1 minute
    maxRecordDuration      time.Duration // Max 25 hours
}

// Fraud detection flags
const (
    "DURATION_TOO_SHORT"      // < 1 minute
    "DURATION_TOO_LONG"       // > 25 hours
    "FUTURE_TIMESTAMP"        // End time in future
    "EXCESSIVE_CPU_USAGE"     // > 2x allocated
    "EXCESSIVE_MEMORY_USAGE"  // > 1.5x allocated
    "NEGATIVE_METRICS"        // Any negative value
    "ZERO_DURATION_WITH_USAGE"// Zero duration but usage > 0
)
```

#### Hands-On Exercise 4.1: Monitoring Setup

**Objective:** Configure comprehensive monitoring

1. Enable Prometheus metrics endpoint
2. Configure Grafana dashboards
3. Set up alerting rules
4. Test alert notifications

**Monitoring Dashboard Example:**

```
┌─────────────────────────────────────────────────────────────────┐
│                   Provider Daemon Dashboard                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Active Workloads│  │ Bids Submitted  │  │ Revenue (24h)   │ │
│  │      127        │  │      1,234      │  │   5,420 VIRT    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                  │
│  CPU Utilization    [████████████░░░░░░░░] 62%                  │
│  Memory Usage       [█████████░░░░░░░░░░░] 45%                  │
│  Storage Usage      [██████████████░░░░░░] 70%                  │
│  GPU Utilization    [████████████████░░░░] 80%                  │
│                                                                  │
│  Recent Alerts:                                                  │
│  ⚠️  High CPU usage on workload-abc123 (92%)                    │
│  ✓  Key rotation completed successfully                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### Day 8: Troubleshooting & Maintenance (2 hours)

#### Common Issues & Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Bid Engine Not Starting | No bids submitted | Check chain connectivity, verify provider config |
| Key Storage Locked | Sign operations fail | Unlock key manager with passphrase |
| Workload Stuck in Deploying | Timeout errors | Check infrastructure adapter, review resource limits |
| Usage Records Not Submitting | Missing settlements | Verify chain connectivity, check key status |
| High Latency on Bids | Slow order response | Enable streaming mode, optimize polling |
| Memory Leaks | Increasing memory usage | Review workload cleanup, check for orphaned resources |

#### Troubleshooting Commands

```bash
# Check daemon status
virtengine-provider status

# View active workloads
virtengine-provider workload list --state running

# Check bid engine status
virtengine-provider bid-engine status

# View recent bids
virtengine-provider bids list --last 100

# Check key manager status
virtengine-provider keys status

# View usage records
virtengine-provider usage list --pending

# Check chain connectivity
virtengine-provider chain ping

# View daemon logs
virtengine-provider logs --tail 100 --level error
```

#### Health Checks

```go
// Health check endpoints
type HealthStatus struct {
    Status      string                 // healthy, degraded, unhealthy
    Components  map[string]ComponentHealth
    LastChecked time.Time
}

type ComponentHealth struct {
    Name    string
    Status  string  // up, down, degraded
    Latency time.Duration
    Message string
}
```

**Health Check Endpoints:**

| Endpoint | Description |
|----------|-------------|
| `/health` | Overall health status |
| `/health/bid-engine` | Bid engine status |
| `/health/key-manager` | Key manager status |
| `/health/usage-meter` | Usage metering status |
| `/health/chain` | Chain connectivity |
| `/health/adapters` | Infrastructure adapters |

#### Maintenance Procedures

**Daily Maintenance Checklist:**
- [ ] Review error logs for anomalies
- [ ] Verify all workloads are in expected states
- [ ] Check usage records are being submitted
- [ ] Monitor resource utilization trends
- [ ] Verify bid success rate

**Weekly Maintenance Checklist:**
- [ ] Review key rotation status
- [ ] Analyze bid performance metrics
- [ ] Check for pending software updates
- [ ] Review security alerts
- [ ] Backup key material

**Monthly Maintenance Checklist:**
- [ ] Rotate provider signing keys
- [ ] Review and update pricing strategy
- [ ] Capacity planning review
- [ ] Security audit of configurations
- [ ] Update disaster recovery procedures

#### Hands-On Exercise 4.2: Incident Response

**Scenario:** Provider daemon is not submitting bids

**Steps:**
1. Check daemon status
2. Review error logs
3. Verify chain connectivity
4. Check key manager status
5. Validate provider configuration
6. Test bid submission manually
7. Document root cause and resolution

---

## Assessments & Certification

### Knowledge Check Questions

**Week 1: Architecture & Components**
1. What are the three core components of the provider daemon?
2. Describe the key storage types and when to use each.
3. What is the default maximum bids per minute?
4. Explain the key lifecycle states.

**Week 2: Infrastructure Adapters**
1. How does namespace isolation work in Kubernetes?
2. What is the role of the Waldur bridge?
3. How are VM states mapped to workload states?
4. Explain the security considerations for each adapter.

**Week 3: Workload Lifecycle**
1. Draw the complete workload state machine.
2. What triggers a transition from RUNNING to FAILED?
3. How is manifest validation performed?
4. What happens during workload termination?

**Week 4: Operations**
1. What metrics are included in a usage record?
2. List three fraud detection flags.
3. Describe the daily maintenance checklist.
4. What are the health check endpoints?

### Practical Assessment

**Scenario:** Complete provider daemon deployment

**Requirements:**
1. Configure provider daemon with file-based key storage
2. Enable Kubernetes and OpenStack adapters
3. Set up competitive pricing configuration
4. Deploy three test workloads
5. Configure monitoring and alerting
6. Perform key rotation
7. Submit usage records for all workloads
8. Document the complete setup

**Grading Criteria:**
- Configuration completeness (25%)
- Security implementation (25%)
- Operational procedures (25%)
- Documentation quality (25%)

### Certification

Upon successful completion:
- Pass knowledge check (80% minimum)
- Complete practical assessment
- Demonstrate troubleshooting skills
- Submit signed operator agreement

**Certificate:** VirtEngine Certified Provider Operator (VCPO)

---

## Additional Resources

### Documentation Links
- [Infrastructure Adapters Training](infrastructure-adapters.md)
- [Marketplace Operations Training](marketplace-operations.md)
- [Provider Guide](/docs/provider-guide.md)
- [API Reference](/docs/api-reference.md)

### Support Channels
- Discord: #provider-operators
- GitHub Issues: virtengine/virtengine
- Email: provider-support@virtengine.io

### Reference Materials
- Provider Daemon Source: `pkg/provider_daemon/`
- Configuration Examples: `deploy/examples/provider/`
- Test Manifests: `tests/fixtures/manifests/`

---

*Last Updated: 2025*  
*Version: 1.0.0*
