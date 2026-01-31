# Validator Operator Training Program

> **Duration:** 16 hours (4 weeks × 4 hours/week)  
> **Prerequisites:** Linux system administration, basic blockchain concepts, networking fundamentals  
> **Certification:** VirtEngine Certified Validator Operator (VE-CVO)

---

## Table of Contents

1. [Program Overview](#program-overview)
2. [Week 1: Validator Fundamentals](#week-1-validator-fundamentals)
3. [Week 2: Consensus Operations](#week-2-consensus-operations)
4. [Week 3: Monitoring and Performance](#week-3-monitoring-and-performance)
5. [Week 4: Upgrades and Maintenance](#week-4-upgrades-and-maintenance)
6. [Assessment and Certification](#assessment-and-certification)
7. [Appendices](#appendices)

---

## Program Overview

### Learning Objectives

Upon completion of this training program, validators will be able to:

- [ ] Deploy and configure a VirtEngine validator node from scratch
- [ ] Understand CometBFT v0.38.x consensus mechanics and block production
- [ ] Implement comprehensive monitoring and alerting systems
- [ ] Execute network upgrades with zero downtime
- [ ] Troubleshoot common validator issues effectively
- [ ] Maintain security best practices for validator operations

### Hardware Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **CPU** | 8 cores (x86_64) | 16+ cores | AMD EPYC or Intel Xeon preferred |
| **RAM** | 32 GB | 64 GB | ECC memory recommended |
| **Storage** | 1 TB NVMe SSD | 2 TB NVMe SSD | RAID 1 for redundancy |
| **Network** | 1 Gbps | 10 Gbps | Low latency (<50ms to peers) |
| **OS** | Ubuntu 22.04 LTS | Ubuntu 24.04 LTS | Debian 12 also supported |

### Software Prerequisites

```bash
# Required software versions
go version        # Go 1.21.0+
make --version    # GNU Make 4.0+
git --version     # Git 2.30+
docker --version  # Docker 24.0+ (for testing)
```

---

## Week 1: Validator Fundamentals

### Day 1: VirtEngine Architecture (2 hours)

#### Learning Objectives
- Understand VirtEngine's modular architecture
- Identify core components and their interactions
- Recognize the role of validators in the ecosystem

#### 1.1 VirtEngine Overview

VirtEngine is a Cosmos SDK v0.53.x-based blockchain for decentralized cloud computing with ML-powered identity verification (VEID).

**Core Architecture:**

```
┌─────────────────────────────────────────────────────────────────┐
│                         VirtEngine Node                          │
├─────────────────────────────────────────────────────────────────┤
│  Application Layer (app/)                                        │
│  ├── Module Wiring (app.go)                                     │
│  ├── Ante Handlers (authentication, fees)                       │
│  └── Genesis Configuration                                       │
├─────────────────────────────────────────────────────────────────┤
│  Blockchain Modules (x/)                                         │
│  ├── x/veid      - Identity verification with ML scoring        │
│  ├── x/mfa       - Multi-factor authentication                  │
│  ├── x/encryption - X25519 encryption envelopes                 │
│  ├── x/market    - Marketplace orders and bids                  │
│  ├── x/escrow    - Payment escrow management                    │
│  ├── x/roles     - Role-based access control                    │
│  └── x/hpc       - High-performance computing                   │
├─────────────────────────────────────────────────────────────────┤
│  Cosmos SDK v0.53.x                                              │
│  ├── Bank, Staking, Slashing, Distribution                      │
│  ├── Governance, Auth, Params                                   │
│  └── IBC-Go v10 (cross-chain)                                   │
├─────────────────────────────────────────────────────────────────┤
│  CometBFT v0.38.x (Consensus Engine)                            │
│  ├── Block Production                                           │
│  ├── Peer-to-Peer Networking                                    │
│  └── Byzantine Fault Tolerance                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Key Validator Responsibilities:**

| Responsibility | Description | Impact |
|---------------|-------------|--------|
| Block Production | Propose new blocks when selected | Network liveness |
| Voting | Sign prevotes/precommits | Consensus finality |
| VEID Processing | Decrypt and score identity data | Identity verification |
| Governance | Vote on proposals | Network evolution |

#### 1.2 Hands-On Exercise: Environment Setup

```bash
# Step 1: Clone the repository
git clone https://github.com/virtengine/virtengine.git
cd virtengine

# Step 2: Install dependencies
make setup

# Step 3: Build the virtengine binary
make virtengine

# Step 4: Verify installation
.cache/bin/virtengine version
# Expected output: virtengine v0.9.x

# Step 5: Initialize the node
.cache/bin/virtengine init my-validator --chain-id virtengine-testnet-1
```

#### 1.3 Knowledge Check

1. What consensus algorithm does VirtEngine use?
2. Name three blockchain modules unique to VirtEngine.
3. What is the role of a validator in VEID processing?

---

### Day 2: Node Configuration (2 hours)

#### Learning Objectives
- Configure validator node settings
- Understand configuration file structure
- Set up secure networking

#### 2.1 Configuration Files Overview

```
~/.virtengine/
├── config/
│   ├── app.toml          # Application configuration
│   ├── client.toml       # CLI client settings
│   ├── config.toml       # CometBFT configuration
│   ├── genesis.json      # Chain genesis state
│   ├── node_key.json     # P2P identity key
│   └── priv_validator_key.json  # Validator signing key
└── data/
    ├── application.db/   # Application state
    ├── blockstore.db/    # Block storage
    ├── cs.wal/           # Consensus write-ahead log
    ├── evidence.db/      # Evidence storage
    ├── snapshots/        # State sync snapshots
    ├── state.db/         # Consensus state
    └── tx_index.db/      # Transaction index
```

#### 2.2 Critical Configuration Settings

**config.toml (CometBFT Configuration):**

```toml
# ~/.virtengine/config/config.toml

[rpc]
# RPC listen address - bind to localhost for security
laddr = "tcp://127.0.0.1:26657"

# Enable CORS for specific origins only
cors_allowed_origins = []

[p2p]
# P2P listen address
laddr = "tcp://0.0.0.0:26656"

# Persistent peers (comma-separated)
persistent_peers = "node_id@ip:26656,node_id@ip:26656"

# Private peer IDs (your sentry nodes)
private_peer_ids = ""

# Address book strict mode
addr_book_strict = true

# Maximum inbound peers
max_num_inbound_peers = 40

# Maximum outbound peers
max_num_outbound_peers = 10

# Seed mode (for seed nodes only)
seed_mode = false

[mempool]
# Maximum transactions in mempool
size = 5000

# Maximum bytes in mempool
max_txs_bytes = 1073741824

# Cache size for rejected transactions
cache_size = 10000

[consensus]
# Timeout for proposal (ms)
timeout_propose = "3s"

# Timeout for prevote (ms)
timeout_prevote = "1s"

# Timeout for precommit (ms)
timeout_precommit = "1s"

# Timeout for commit (ms)
timeout_commit = "5s"

# Skip timeout commit (faster blocks)
skip_timeout_commit = false

# Double sign check height
double_sign_check_height = 10
```

**app.toml (Application Configuration):**

```toml
# ~/.virtengine/config/app.toml

[api]
# Enable API server
enable = true

# Swagger documentation
swagger = false

# API listen address
address = "tcp://127.0.0.1:1317"

[grpc]
# gRPC server
enable = true
address = "0.0.0.0:9090"

[grpc-web]
# gRPC-Web gateway
enable = true
address = "0.0.0.0:9091"

[state-sync]
# State sync snapshot interval
snapshot-interval = 1000

# Keep last N snapshots
snapshot-keep-recent = 2

[pruning]
# Pruning strategy: default, nothing, everything, custom
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "500"
pruning-interval = "10"

[telemetry]
# Enable Prometheus metrics
enabled = true
prometheus-retention-time = 60

[veid]
# VEID-specific settings
max-verification-timeout = "30s"
encryption-algorithm = "X25519-XSalsa20-Poly1305"
deterministic-inference = true
```

#### 2.3 Hands-On Exercise: Configure Your Node

```bash
# Step 1: Backup default configuration
cp -r ~/.virtengine/config ~/.virtengine/config.backup

# Step 2: Configure persistent peers
PEERS="a1b2c3d4e5f6@validator1.virtengine.io:26656,..."
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" \
    ~/.virtengine/config/config.toml

# Step 3: Configure minimum gas prices
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025uvirt"/' \
    ~/.virtengine/config/app.toml

# Step 4: Enable Prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' \
    ~/.virtengine/config/config.toml

# Step 5: Validate configuration
.cache/bin/virtengine validate-genesis
```

#### 2.4 Network Security Configuration

```bash
# Firewall rules (UFW example)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 26656/tcp   # P2P
sudo ufw allow 26657/tcp   # RPC (restrict to trusted IPs)
sudo ufw allow 9090/tcp    # gRPC
sudo ufw enable

# Rate limiting with iptables
sudo iptables -A INPUT -p tcp --dport 26656 -m connlimit \
    --connlimit-above 100 -j REJECT
```

---

### Day 3: Genesis and Chain Initialization (1 hour)

#### Learning Objectives
- Understand genesis file structure
- Join a network with correct genesis
- Verify chain synchronization

#### 3.1 Genesis File Structure

```json
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "virtengine-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000"
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    }
  },
  "app_state": {
    "staking": {
      "params": {
        "unbonding_time": "1209600s",
        "max_validators": 100,
        "bond_denom": "uvirt"
      }
    },
    "veid": {
      "params": {
        "min_score_threshold": "0.85",
        "encryption_algorithm": "X25519-XSalsa20-Poly1305",
        "model_version": "v1.2.0"
      }
    }
  }
}
```

#### 3.2 Joining an Existing Network

```bash
# Step 1: Download genesis file
curl -s https://raw.githubusercontent.com/virtengine/networks/main/mainnet/genesis.json \
    > ~/.virtengine/config/genesis.json

# Step 2: Verify genesis hash
EXPECTED_HASH="abc123..."
ACTUAL_HASH=$(sha256sum ~/.virtengine/config/genesis.json | cut -d' ' -f1)
if [ "$EXPECTED_HASH" == "$ACTUAL_HASH" ]; then
    echo "Genesis verified successfully"
else
    echo "ERROR: Genesis hash mismatch!"
    exit 1
fi

# Step 3: Configure state sync (optional, for faster sync)
SNAP_RPC="https://rpc.virtengine.io:443"
LATEST_HEIGHT=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((LATEST_HEIGHT - 2000))
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$TRUST_HEIGHT" | \
    jq -r .result.block_id.hash)

sed -i \
    -e "s/enable = false/enable = true/" \
    -e "s/rpc_servers = \"\"/rpc_servers = \"$SNAP_RPC,$SNAP_RPC\"/" \
    -e "s/trust_height = 0/trust_height = $TRUST_HEIGHT/" \
    -e "s/trust_hash = \"\"/trust_hash = \"$TRUST_HASH\"/" \
    ~/.virtengine/config/config.toml

# Step 4: Start the node
.cache/bin/virtengine start
```

---

### Day 4: Creating a Validator (1 hour)

#### Learning Objectives
- Create validator transaction
- Understand staking parameters
- Configure commission rates

#### 4.1 Validator Creation Process

```bash
# Step 1: Ensure node is fully synced
.cache/bin/virtengine status | jq '.SyncInfo.catching_up'
# Must return: false

# Step 2: Create or import operator wallet
.cache/bin/virtengine keys add operator --keyring-backend file
# Save mnemonic securely!

# Step 3: Fund the operator wallet
# Transfer tokens from exchange or faucet

# Step 4: Verify balance
.cache/bin/virtengine query bank balances $(virtengine keys show operator -a)

# Step 5: Create validator
.cache/bin/virtengine tx staking create-validator \
    --amount=1000000uvirt \
    --pubkey=$(.cache/bin/virtengine tendermint show-validator) \
    --moniker="MyValidator" \
    --chain-id=virtengine-1 \
    --commission-rate="0.10" \
    --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="1" \
    --gas="auto" \
    --gas-adjustment="1.5" \
    --gas-prices="0.025uvirt" \
    --from=operator \
    --keyring-backend=file
```

#### 4.2 Validator Parameters Explained

| Parameter | Description | Recommendation |
|-----------|-------------|----------------|
| `commission-rate` | Current commission percentage | 5-10% for competitive |
| `commission-max-rate` | Maximum commission (immutable) | Set carefully, 20-25% |
| `commission-max-change-rate` | Max daily change | 1-2% for delegator trust |
| `min-self-delegation` | Minimum self-bond | Higher = more commitment |

#### 4.3 Verification Checklist

- [ ] Validator appears in active set: `virtengine query staking validators`
- [ ] Signing blocks: Check explorer or logs
- [ ] Commission configured correctly
- [ ] Moniker and identity set properly
- [ ] Contact info in keybase identity

---

## Week 2: Consensus Operations

### Day 5: CometBFT Consensus Deep Dive (2 hours)

#### Learning Objectives
- Understand Tendermint BFT consensus
- Trace block production lifecycle
- Identify consensus failure modes

#### 5.1 CometBFT v0.38.x Consensus Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    CometBFT Consensus Round                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────┐    ┌─────────┐    ┌──────────┐    ┌─────────┐      │
│  │ Propose │───>│ Prevote │───>│ Precommit│───>│ Commit  │      │
│  └─────────┘    └─────────┘    └──────────┘    └─────────┘      │
│       │              │               │              │            │
│   timeout_       timeout_        timeout_       timeout_         │
│   propose        prevote        precommit       commit           │
│    (3s)           (1s)           (1s)           (5s)            │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘

Round Success Criteria:
- Propose: Valid block from designated proposer
- Prevote: >2/3 voting power prevotes for same block
- Precommit: >2/3 voting power precommits for same block
- Commit: Block finalized and applied to state
```

#### 5.2 Block Production Mechanics

```go
// Simplified block production flow
type BlockProduction struct {
    Height     int64
    Round      int32
    Proposer   ValidatorAddress
    Block      *Block
    Signatures []CommitSig
}

// Proposer selection (deterministic round-robin weighted by voting power)
func GetProposer(height, round int64, validators []Validator) Validator {
    // Priority based on voting power and previous proposals
    // Ensures fair distribution over time
}
```

**Block Structure:**

```json
{
  "block": {
    "header": {
      "height": "12345678",
      "time": "2024-01-15T10:30:00Z",
      "proposer_address": "ABCD1234...",
      "validators_hash": "...",
      "app_hash": "..."
    },
    "data": {
      "txs": ["base64_encoded_tx_1", "base64_encoded_tx_2"]
    },
    "evidence": {
      "evidence": []
    },
    "last_commit": {
      "signatures": [...]
    }
  }
}
```

#### 5.3 Voting Power and Stake

```bash
# Check your validator's voting power
.cache/bin/virtengine query staking validator \
    $(.cache/bin/virtengine keys show operator --bech val -a) \
    --output json | jq '.tokens, .delegator_shares'

# View all validators by voting power
.cache/bin/virtengine query staking validators \
    --output json | jq '.validators | sort_by(.tokens | tonumber) | reverse'
```

#### 5.4 Hands-On Exercise: Consensus Analysis

```bash
# Exercise: Analyze recent blocks

# 1. Get latest block
.cache/bin/virtengine query block --type=height 12345678

# 2. Identify proposer
.cache/bin/virtengine query tendermint-validator-set 12345678

# 3. Analyze commit signatures
# Count signers vs total validators

# 4. Calculate participation rate
SIGNED=$(echo $COMMIT | jq '.signatures | map(select(.block_id_flag == 2)) | length')
TOTAL=$(echo $VALSET | jq '.validators | length')
echo "Participation: $SIGNED / $TOTAL"
```

---

### Day 6: Slashing and Jail Conditions (1 hour)

#### Learning Objectives
- Understand slashing conditions
- Prevent downtime and double-signing
- Recover from jail

#### 6.1 Slashing Conditions

| Infraction | Penalty | Jail Duration | Recovery |
|------------|---------|---------------|----------|
| **Downtime** | 0.01% stake | 10 minutes | Unjail transaction |
| **Double Sign** | 5% stake | Permanent | Cannot unjail |

**Downtime Detection:**

```
Signed Blocks Window: 10,000 blocks
Minimum Signed: 5,000 blocks (50%)
Downtime Jail: Missed > 5,000 in window
```

#### 6.2 Preventing Slashing

```bash
# Monitor signing status
.cache/bin/virtengine query slashing signing-info \
    $(.cache/bin/virtengine tendermint show-validator)

# Output example:
# {
#   "address": "virtvalcons1...",
#   "start_height": "1000000",
#   "index_offset": "50000",
#   "jailed_until": "1970-01-01T00:00:00Z",
#   "tombstoned": false,
#   "missed_blocks_counter": "5"
# }
```

**Double-Sign Prevention Checklist:**

- [ ] Never run two validators with the same key
- [ ] Use TMKMS (Tendermint Key Management System)
- [ ] Implement proper failover procedures
- [ ] Monitor for duplicate P2P connections
- [ ] Use separate signing infrastructure

#### 6.3 Unjailing Procedure

```bash
# Step 1: Verify jail status
.cache/bin/virtengine query staking validator \
    $(.cache/bin/virtengine keys show operator --bech val -a) | grep jailed

# Step 2: Wait for jail period to expire
# Minimum: 10 minutes

# Step 3: Fix the underlying issue (node sync, connectivity)

# Step 4: Send unjail transaction
.cache/bin/virtengine tx slashing unjail \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas="auto" \
    --gas-prices="0.025uvirt" \
    --keyring-backend=file
```

---

### Day 7: Governance Participation (1 hour)

#### Learning Objectives
- Participate in governance votes
- Submit proposals
- Understand proposal lifecycle

#### 7.1 Proposal Types

| Type | Description | Deposit | Voting Period |
|------|-------------|---------|---------------|
| Text | Signaling proposal | 1000 VIRT | 7 days |
| Parameter Change | Modify chain params | 1000 VIRT | 7 days |
| Software Upgrade | Coordinate upgrades | 5000 VIRT | 7 days |
| Community Pool | Spend community funds | 1000 VIRT | 7 days |

#### 7.2 Voting on Proposals

```bash
# List active proposals
.cache/bin/virtengine query gov proposals --status=voting_period

# View proposal details
.cache/bin/virtengine query gov proposal 42

# Vote on proposal
.cache/bin/virtengine tx gov vote 42 yes \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas="auto" \
    --gas-prices="0.025uvirt" \
    --keyring-backend=file

# Vote options: yes, no, no_with_veto, abstain
```

#### 7.3 Validator Governance Best Practices

1. **Review thoroughly** - Read all proposal details before voting
2. **Communicate** - Share voting rationale with delegators
3. **Vote early** - Allow delegators to override if needed
4. **Track quorum** - Ensure proposals reach quorum
5. **Emergency response** - Have procedures for urgent proposals

---

## Week 3: Monitoring and Performance

### Day 8: Monitoring Infrastructure (2 hours)

#### Learning Objectives
- Deploy Prometheus and Grafana
- Configure alerting rules
- Create operational dashboards

#### 8.1 Prometheus Configuration

```yaml
# /etc/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

rule_files:
  - "virtengine_alerts.yml"

scrape_configs:
  - job_name: 'virtengine'
    static_configs:
      - targets: ['localhost:26660']
    metrics_path: /metrics

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']

  - job_name: 'process-exporter'
    static_configs:
      - targets: ['localhost:9256']
```

#### 8.2 Critical Alerts

```yaml
# /etc/prometheus/virtengine_alerts.yml
groups:
  - name: virtengine_validator
    rules:
      # Block signing alerts
      - alert: ValidatorMissedBlocks
        expr: increase(tendermint_consensus_validator_missed_blocks[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Validator missing blocks"
          description: "Missed {{ $value }} blocks in the last 5 minutes"

      - alert: ValidatorMissedBlocksCritical
        expr: increase(tendermint_consensus_validator_missed_blocks[10m]) > 100
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Critical block signing failure"
          description: "Missed {{ $value }} blocks - jail imminent!"

      # Consensus health
      - alert: ConsensusTimeout
        expr: increase(tendermint_consensus_round[5m]) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Multiple consensus rounds"
          description: "{{ $value }} rounds in 5 minutes indicates consensus issues"

      # Peer connectivity
      - alert: LowPeerCount
        expr: tendermint_p2p_peers < 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low peer count"
          description: "Only {{ $value }} peers connected"

      # Memory and resources
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / 1024 / 1024 / 1024 > 50
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Process using {{ $value }}GB memory"

      # Sync status
      - alert: NodeOutOfSync
        expr: tendermint_consensus_latest_block_height < (max(tendermint_consensus_latest_block_height) - 100)
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Node out of sync"
          description: "Node is {{ $value }} blocks behind"

      # VEID specific
      - alert: VEIDProcessingBacklog
        expr: veid_pending_verifications > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "VEID verification backlog"
          description: "{{ $value }} pending verifications"

      - alert: VEIDMLModelError
        expr: increase(veid_ml_inference_errors_total[5m]) > 5
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "VEID ML model errors"
          description: "{{ $value }} inference errors in 5 minutes"
```

#### 8.3 Grafana Dashboard Setup

```json
{
  "dashboard": {
    "title": "VirtEngine Validator",
    "panels": [
      {
        "title": "Block Height",
        "type": "stat",
        "targets": [
          {
            "expr": "tendermint_consensus_height"
          }
        ]
      },
      {
        "title": "Missed Blocks (24h)",
        "type": "stat",
        "targets": [
          {
            "expr": "increase(tendermint_consensus_validator_missed_blocks[24h])"
          }
        ]
      },
      {
        "title": "Peer Count",
        "type": "gauge",
        "targets": [
          {
            "expr": "tendermint_p2p_peers"
          }
        ]
      },
      {
        "title": "Consensus Rounds",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(tendermint_consensus_rounds_total[5m])"
          }
        ]
      },
      {
        "title": "VEID Verifications",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(veid_verifications_total[5m])"
          }
        ]
      }
    ]
  }
}
```

#### 8.4 Hands-On Exercise: Deploy Monitoring Stack

```bash
# Step 1: Install Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvf prometheus-2.45.0.linux-amd64.tar.gz
sudo mv prometheus-2.45.0.linux-amd64 /opt/prometheus

# Step 2: Configure as systemd service
sudo tee /etc/systemd/system/prometheus.service << EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
ExecStart=/opt/prometheus/prometheus \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/var/lib/prometheus
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus

# Step 3: Install Grafana
sudo apt-get install -y apt-transport-https software-properties-common
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
echo "deb https://packages.grafana.com/oss/deb stable main" | \
    sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get update
sudo apt-get install grafana
sudo systemctl enable grafana-server
sudo systemctl start grafana-server
```

---

### Day 9: Performance Tuning (1 hour)

#### Learning Objectives
- Optimize node performance
- Tune database settings
- Manage disk and memory

#### 9.1 System-Level Optimization

```bash
# /etc/sysctl.d/99-virtengine.conf

# Network optimization
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.core.netdev_max_backlog = 30000
net.ipv4.tcp_congestion_control = bbr
net.core.default_qdisc = fq

# File descriptors
fs.file-max = 2097152
fs.nr_open = 2097152

# Virtual memory
vm.swappiness = 1
vm.dirty_ratio = 40
vm.dirty_background_ratio = 10

# Apply changes
sudo sysctl -p /etc/sysctl.d/99-virtengine.conf
```

#### 9.2 LevelDB Tuning

```toml
# ~/.virtengine/config/app.toml

[db]
# Use goleveldb for best performance
backend = "goleveldb"

# Increase block cache (default 8MB)
block-cache-capacity = 134217728  # 128MB

# Write buffer size
write-buffer-size = 67108864  # 64MB

# Compaction settings
compaction-table-size = 8388608  # 8MB
compaction-total-size = 134217728  # 128MB
```

#### 9.3 Pruning Strategy

| Strategy | Description | Disk Usage | Use Case |
|----------|-------------|------------|----------|
| `default` | Keep 100 + every 500th | Medium | Most validators |
| `nothing` | Keep all states | Very High | Archive nodes |
| `everything` | Keep only latest | Lowest | Minimal requirements |
| `custom` | Configure manually | Variable | Specific needs |

```toml
# Custom pruning for validators
[pruning]
pruning = "custom"
pruning-keep-recent = "100"     # Last 100 heights
pruning-keep-every = "1000"     # Every 1000th state
pruning-interval = "10"         # Prune every 10 blocks
```

#### 9.4 Performance Monitoring Commands

```bash
# Check sync speed
watch -n 10 '.cache/bin/virtengine status | jq ".SyncInfo | {height: .latest_block_height, time: .latest_block_time}"'

# Monitor disk I/O
iostat -x 5

# Check memory usage
ps aux --sort=-%mem | head -20

# Analyze block times
.cache/bin/virtengine query block | jq '.block.header.time' 
```

---

### Day 10: Log Analysis and Debugging (1 hour)

#### Learning Objectives
- Configure structured logging
- Analyze consensus logs
- Debug common issues

#### 10.1 Log Configuration

```toml
# ~/.virtengine/config/config.toml

[log]
# Log level: trace, debug, info, warn, error
log_level = "info"

# Log format: plain or json
log_format = "json"

# Module-specific log levels
log_level = "main:info,state:info,statesync:info,*:error"
```

**Log Aggregation with journald:**

```bash
# View real-time logs
journalctl -u virtengine -f

# Filter by severity
journalctl -u virtengine -p err

# Export to file
journalctl -u virtengine --since "1 hour ago" > validator_logs.txt
```

#### 10.2 Critical Log Patterns

```bash
# Consensus issues
grep -E "(round|votes|timeout)" /var/log/virtengine.log

# Peer connectivity
grep -E "(peer|connection|dial)" /var/log/virtengine.log

# Signing activity
grep -E "(signed|vote|precommit|prevote)" /var/log/virtengine.log

# VEID processing
grep -E "(veid|verification|decryption)" /var/log/virtengine.log

# Errors and warnings
grep -E "(ERR|error|WARN|warning)" /var/log/virtengine.log
```

#### 10.3 Common Issues and Solutions

| Issue | Log Pattern | Solution |
|-------|-------------|----------|
| Block production failure | `"Failed to create proposal block"` | Check disk space, memory |
| Peer disconnection | `"Stopping peer for error"` | Review firewall, network |
| Consensus timeout | `"Timed out waiting for votes"` | Check network latency |
| DB corruption | `"leveldb: manifest corrupted"` | Restore from snapshot |
| Out of memory | `"runtime: out of memory"` | Increase RAM or optimize |

---

## Week 4: Upgrades and Maintenance

### Day 11: Network Upgrades (2 hours)

#### Learning Objectives
- Understand upgrade proposals
- Execute coordinated upgrades
- Handle upgrade failures

#### 11.1 Upgrade Types

| Type | Description | Downtime | Coordination |
|------|-------------|----------|--------------|
| **Soft Fork** | Backward compatible | None | Optional |
| **Hard Fork** | Breaking changes | Brief | Required |
| **Emergency** | Critical fixes | Variable | As needed |

#### 11.2 Upgrade Proposal Workflow

```bash
# Step 1: Monitor for upgrade proposals
.cache/bin/virtengine query gov proposals --status=voting_period

# Step 2: Review upgrade details
.cache/bin/virtengine query upgrade plan

# Output example:
# {
#   "name": "v0.10.0",
#   "height": "5000000",
#   "info": "Upgrade to v0.10.0 with new VEID features"
# }

# Step 3: Vote on proposal
.cache/bin/virtengine tx gov vote PROPOSAL_ID yes \
    --from=operator \
    --chain-id=virtengine-1

# Step 4: Prepare binary
git fetch --tags
git checkout v0.10.0
make virtengine
cp .cache/bin/virtengine ~/.virtengine/cosmovisor/upgrades/v0.10.0/bin/
```

#### 11.3 Cosmovisor Setup

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Directory structure
mkdir -p ~/.virtengine/cosmovisor/genesis/bin
mkdir -p ~/.virtengine/cosmovisor/upgrades

# Copy current binary
cp .cache/bin/virtengine ~/.virtengine/cosmovisor/genesis/bin/

# Environment variables
export DAEMON_NAME=virtengine
export DAEMON_HOME=$HOME/.virtengine
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
export DAEMON_RESTART_AFTER_UPGRADE=true
export UNSAFE_SKIP_BACKUP=false

# Create systemd service
sudo tee /etc/systemd/system/virtengine.service << EOF
[Unit]
Description=VirtEngine Validator
After=network.target

[Service]
Type=simple
User=validator
ExecStart=$(which cosmovisor) run start
Restart=always
RestartSec=3
Environment="DAEMON_NAME=virtengine"
Environment="DAEMON_HOME=/home/validator/.virtengine"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable virtengine
```

#### 11.4 Manual Upgrade Procedure

```bash
# At upgrade height, node will halt automatically

# Step 1: Stop the node (if not already stopped)
sudo systemctl stop virtengine

# Step 2: Backup data
tar -czvf ~/.virtengine/backup_$(date +%Y%m%d).tar.gz \
    ~/.virtengine/config \
    ~/.virtengine/data

# Step 3: Replace binary
git checkout v0.10.0
make virtengine
sudo cp .cache/bin/virtengine /usr/local/bin/

# Step 4: Verify version
virtengine version
# Expected: v0.10.0

# Step 5: Restart
sudo systemctl start virtengine

# Step 6: Verify operation
journalctl -u virtengine -f
.cache/bin/virtengine status
```

#### 11.5 Rollback Procedure

```bash
# If upgrade fails, rollback immediately

# Step 1: Stop node
sudo systemctl stop virtengine

# Step 2: Restore backup
tar -xzvf ~/.virtengine/backup_YYYYMMDD.tar.gz -C /

# Step 3: Restore old binary
git checkout v0.9.x
make virtengine
sudo cp .cache/bin/virtengine /usr/local/bin/

# Step 4: Reset state if needed
.cache/bin/virtengine tendermint unsafe-reset-all --keep-addr-book

# Step 5: Restart with old version
sudo systemctl start virtengine
```

---

### Day 12: Backup and Recovery (1 hour)

#### Learning Objectives
- Implement backup strategies
- Perform state restoration
- Test disaster recovery

#### 12.1 Backup Strategy

| Component | Frequency | Retention | Method |
|-----------|-----------|-----------|--------|
| `config/` | On change | Indefinite | Git + encrypted storage |
| `priv_validator_key.json` | Once | Forever | HSM + encrypted backup |
| `data/` | Daily | 7 days | Snapshot + rsync |
| `node_key.json` | Once | Forever | Encrypted backup |

#### 12.2 Automated Backup Script

```bash
#!/bin/bash
# /opt/scripts/validator_backup.sh

set -e

BACKUP_DIR="/backup/virtengine"
DATE=$(date +%Y%m%d_%H%M%S)
VIRTENGINE_HOME="$HOME/.virtengine"

# Create backup directory
mkdir -p "$BACKUP_DIR/$DATE"

# Stop node for consistent backup
sudo systemctl stop virtengine

# Backup configuration (encrypted)
tar -czf - "$VIRTENGINE_HOME/config" | \
    gpg --symmetric --cipher-algo AES256 \
    > "$BACKUP_DIR/$DATE/config.tar.gz.gpg"

# Backup data directory
tar -czf "$BACKUP_DIR/$DATE/data.tar.gz" \
    "$VIRTENGINE_HOME/data"

# Restart node
sudo systemctl start virtengine

# Cleanup old backups (keep 7 days)
find "$BACKUP_DIR" -type d -mtime +7 -exec rm -rf {} +

# Upload to remote storage (example with rclone)
rclone sync "$BACKUP_DIR" remote:virtengine-backups

echo "Backup completed: $BACKUP_DIR/$DATE"
```

#### 12.3 State Sync Recovery

```bash
# Fast recovery using state sync

# Step 1: Stop node and reset
sudo systemctl stop virtengine
.cache/bin/virtengine tendermint unsafe-reset-all --keep-addr-book

# Step 2: Configure state sync
SNAP_RPC="https://rpc.virtengine.io:443"
LATEST=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((LATEST - 2000))
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$TRUST_HEIGHT" | \
    jq -r .result.block_id.hash)

cat >> ~/.virtengine/config/config.toml << EOF

[statesync]
enable = true
rpc_servers = "$SNAP_RPC,$SNAP_RPC"
trust_height = $TRUST_HEIGHT
trust_hash = "$TRUST_HASH"
trust_period = "168h"
EOF

# Step 3: Start and sync
sudo systemctl start virtengine
journalctl -u virtengine -f
```

---

### Day 13: Troubleshooting (1 hour)

#### Learning Objectives
- Diagnose common issues
- Apply systematic debugging
- Escalate appropriately

#### 13.1 Troubleshooting Decision Tree

```
Issue Reported
     │
     ├──> Node not starting?
     │         ├──> Check logs: journalctl -u virtengine
     │         ├──> Verify disk space: df -h
     │         ├──> Check memory: free -h
     │         └──> Validate config: virtengine validate-genesis
     │
     ├──> Missing blocks?
     │         ├──> Check sync: virtengine status | jq '.SyncInfo'
     │         ├──> Verify connectivity: virtengine tendermint show-node-id
     │         ├──> Check signing: query slashing signing-info
     │         └──> Review peers: net_info
     │
     ├──> Performance issues?
     │         ├──> Monitor CPU: top, htop
     │         ├──> Check I/O: iostat -x 5
     │         ├──> Analyze DB: du -sh ~/.virtengine/data/*
     │         └──> Review logs for slow queries
     │
     └──> Consensus failures?
               ├──> Check proposer selection
               ├──> Verify network latency
               ├──> Review timeout settings
               └──> Analyze peer connectivity
```

#### 13.2 Emergency Runbooks

**Runbook: Node Stuck at Height**

```bash
# Diagnosis
.cache/bin/virtengine status | jq '.SyncInfo'
journalctl -u virtengine --since "10 minutes ago" | grep -E "(error|panic)"

# Solution 1: Restart
sudo systemctl restart virtengine

# Solution 2: Clear mempool
.cache/bin/virtengine tendermint unsafe-reset-all --keep-addr-book

# Solution 3: State sync from fresh
# (See State Sync Recovery above)
```

**Runbook: Jailed for Downtime**

```bash
# Verify jail status
.cache/bin/virtengine query staking validator $(virtengine keys show operator --bech val -a)

# Check jail expiry time
.cache/bin/virtengine query slashing signing-info $(virtengine tendermint show-validator)

# Ensure node is synced
.cache/bin/virtengine status | jq '.SyncInfo.catching_up'

# Unjail when ready
.cache/bin/virtengine tx slashing unjail \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas="auto" \
    --gas-prices="0.025uvirt"
```

**Runbook: Database Corruption**

```bash
# Symptoms: Node crashes with leveldb errors

# Step 1: Stop node
sudo systemctl stop virtengine

# Step 2: Backup corrupted data
mv ~/.virtengine/data ~/.virtengine/data.corrupted

# Step 3: Try repair
.cache/bin/virtengine tendermint repair-wal

# Step 4: If repair fails, state sync
.cache/bin/virtengine tendermint unsafe-reset-all --keep-addr-book
# Configure state sync and restart
```

---

### Day 14: Review and Assessment (1 hour)

#### 14.1 Knowledge Assessment

**Written Exam (30 minutes)**

1. Describe the CometBFT consensus process and explain each phase.
2. What are the slashing conditions and their penalties?
3. How does Cosmovisor facilitate network upgrades?
4. Explain the VEID verification pipeline and validator's role.
5. What monitoring metrics are critical for validator operations?

**Practical Assessment (30 minutes)**

- [ ] Deploy a validator node from scratch
- [ ] Configure monitoring and alerting
- [ ] Simulate and recover from a jailed state
- [ ] Execute a mock upgrade procedure
- [ ] Debug a provided misconfiguration

#### 14.2 Certification Requirements

| Requirement | Minimum Score |
|-------------|---------------|
| Written Exam | 80% |
| Practical Assessment | Pass all tasks |
| Attendance | 90% of sessions |
| Lab Exercises | All completed |

---

## Assessment and Certification

### Final Exam Structure

| Section | Duration | Weight |
|---------|----------|--------|
| Multiple Choice | 30 min | 25% |
| Short Answer | 30 min | 25% |
| Practical Lab | 60 min | 50% |

### Certification Levels

| Level | Requirements | Validity |
|-------|--------------|----------|
| **VE-CVO** (Certified Validator Operator) | Pass this course | 2 years |
| **VE-SVO** (Senior Validator Operator) | VE-CVO + 1 year experience | 2 years |
| **VE-EVO** (Expert Validator Operator) | VE-SVO + advanced training | 3 years |

---

## Appendices

### Appendix A: Quick Reference Commands

```bash
# Node Operations
virtengine start                          # Start node
virtengine status                         # Check status
virtengine tendermint show-validator      # Show validator pubkey
virtengine tendermint show-node-id        # Show P2P node ID

# Staking
virtengine query staking validators       # List validators
virtengine query staking validator VAL    # Validator info
virtengine tx staking delegate ...        # Delegate tokens
virtengine tx staking unbond ...          # Unbond tokens

# Slashing
virtengine query slashing signing-info    # Signing status
virtengine tx slashing unjail             # Unjail validator

# Governance
virtengine query gov proposals            # List proposals
virtengine tx gov vote ...                # Vote on proposal

# Diagnostics
virtengine query tendermint-validator-set # Current validator set
virtengine query block HEIGHT             # Get block info
```

### Appendix B: Network Endpoints

| Network | RPC | gRPC | REST |
|---------|-----|------|------|
| Mainnet | rpc.virtengine.io:443 | grpc.virtengine.io:443 | api.virtengine.io:443 |
| Testnet | rpc-testnet.virtengine.io:443 | grpc-testnet.virtengine.io:443 | api-testnet.virtengine.io:443 |

### Appendix C: Emergency Contacts

| Issue | Contact | Response Time |
|-------|---------|---------------|
| Critical Security | security@virtengine.io | 1 hour |
| Network Issues | ops@virtengine.io | 4 hours |
| Validator Support | Discord #validators | Community |
| Documentation | docs@virtengine.io | 24 hours |

### Appendix D: Additional Resources

- [VirtEngine Documentation](https://docs.virtengine.io)
- [Cosmos SDK Documentation](https://docs.cosmos.network)
- [CometBFT Documentation](https://docs.cometbft.com)
- [Validator Community Discord](https://discord.gg/virtengine)
- [GitHub Repository](https://github.com/virtengine/virtengine)

---

*Document Version: 1.0.0*  
*Last Updated: 2024-01-15*  
*Maintainer: VirtEngine Validator Operations Team*
