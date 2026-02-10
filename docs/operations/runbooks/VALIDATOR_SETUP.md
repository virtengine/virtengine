# Validator Setup and Maintenance Guide

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Validator Creation](#validator-creation)
4. [Identity Verification Setup](#identity-verification-setup)
5. [Monitoring Configuration](#monitoring-configuration)
6. [Daily Operations](#daily-operations)
7. [Maintenance Procedures](#maintenance-procedures)
8. [Security Hardening](#security-hardening)
9. [Sentry Node Architecture](#sentry-node-architecture)

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| CPU | 8 cores | 16+ cores | AMD EPYC or Intel Xeon recommended |
| RAM | 32 GB | 64 GB | ECC memory preferred |
| Storage | 1 TB NVMe SSD | 2 TB NVMe SSD | RAID-1 for redundancy |
| Network | 100 Mbps | 1 Gbps | Dedicated bandwidth |

### Software Requirements

```bash
# Required packages
- Ubuntu 22.04 LTS (or equivalent)
- Go 1.21+
- TensorFlow 2.15.0 (for VEID scoring)
- Docker (optional)
- Prometheus node_exporter
```

### Network Requirements

| Port | Protocol | Purpose | Exposure |
|------|----------|---------|----------|
| 26656 | TCP | P2P | Public |
| 26657 | TCP | RPC | Internal only |
| 9090 | TCP | gRPC | Internal only |
| 1317 | TCP | REST API | Internal only |
| 26660 | TCP | Prometheus metrics | Monitoring only |

---

## Initial Setup

### Step 1: Prepare the System

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y build-essential git curl wget jq lz4

# Create validator user
sudo useradd -m -s /bin/bash validator
sudo su - validator

# Set up Go (if building from source)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Step 2: Install VirtEngine Binary

```bash
# Option A: Download pre-built binary
RELEASE_VERSION="v1.0.0"
wget https://github.com/virtengine/virtengine/releases/download/${RELEASE_VERSION}/virtengine_linux_amd64.tar.gz
tar -xzf virtengine_linux_amd64.tar.gz
sudo mv virtengine /usr/local/bin/

# Verify installation
virtengine version
# Expected: virtengine version v1.0.0

# Verify binary checksum
sha256sum /usr/local/bin/virtengine
# Compare with official checksum from releases page
```

```bash
# Option B: Build from source
git clone https://github.com/virtengine/virtengine.git
cd virtengine
git checkout ${RELEASE_VERSION}
make virtengine
sudo mv .cache/bin/virtengine /usr/local/bin/
```

### Step 3: Initialize Node

```bash
# Set chain ID
CHAIN_ID="virtengine-1"
MONIKER="your-validator-name"

# Initialize node
virtengine init "${MONIKER}" --chain-id ${CHAIN_ID}

# Download genesis file
wget -O ~/.virtengine/config/genesis.json \
    https://raw.githubusercontent.com/virtengine/networks/main/${CHAIN_ID}/genesis.json

# Verify genesis hash
EXPECTED_GENESIS_HASH="<official_genesis_hash>"
ACTUAL_HASH=$(sha256sum ~/.virtengine/config/genesis.json | cut -d' ' -f1)
if [ "$EXPECTED_GENESIS_HASH" != "$ACTUAL_HASH" ]; then
    echo "ERROR: Genesis hash mismatch!"
    exit 1
fi
echo "Genesis verified successfully"
```

### Step 4: Configure Node

**config.toml** - `~/.virtengine/config/config.toml`:

```toml
# Node identification
moniker = "your-validator-name"

# Database backend (goleveldb recommended for validators)
db_backend = "goleveldb"

# Seed nodes (official)
seeds = "seed1@seed1.virtengine.com:26656,seed2@seed2.virtengine.com:26656"

# Persistent peers (get from network registry)
persistent_peers = ""

# Enable Prometheus metrics
prometheus = true
prometheus_listen_addr = ":26660"

[rpc]
# Bind to localhost only
laddr = "tcp://127.0.0.1:26657"
# Max connections
max_open_connections = 900

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = "tcp://YOUR_PUBLIC_IP:26656"
max_num_inbound_peers = 100
max_num_outbound_peers = 50
# Seed mode (false for validators)
seed_mode = false
# Peer exchange
pex = true

[mempool]
# Mempool size
size = 5000
max_txs_bytes = 1073741824
cache_size = 10000

[consensus]
# Timeout settings
timeout_propose = "3s"
timeout_prevote = "1s"
timeout_precommit = "1s"
timeout_commit = "5s"
# Double sign check (CRITICAL for validators)
double_sign_check_height = 10

[statesync]
# Enable for faster initial sync
enable = false
```

**app.toml** - `~/.virtengine/config/app.toml`:

```toml
# Minimum gas prices
minimum-gas-prices = "0.025uve"

# Pruning strategy
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "1000"
pruning-interval = "10"

[api]
enable = true
address = "tcp://127.0.0.1:1317"
max-open-connections = 1000

[grpc]
enable = true
address = "0.0.0.0:9090"

[state-sync]
# Enable snapshots for other nodes
snapshot-interval = 1000
snapshot-keep-recent = 2

[telemetry]
enabled = true
prometheus-retention-time = 3600
```

### Step 5: Create Systemd Service

```ini
# /etc/systemd/system/virtengine.service
[Unit]
Description=VirtEngine Validator Node
After=network.target

[Service]
Type=simple
User=validator
Group=validator
ExecStart=/usr/local/bin/virtengine start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
LimitNPROC=65535

# Environment
Environment="HOME=/home/validator"

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/validator/.virtengine
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable virtengine
sudo systemctl start virtengine

# Check status
sudo systemctl status virtengine

# View logs
sudo journalctl -u virtengine -f
```

### Step 6: Sync Node

```bash
# Check sync status
virtengine status | jq '.SyncInfo'

# Expected output when synced:
# "catching_up": false
# "latest_block_height": <current_network_height>
```

**Fast Sync Options:**

```bash
# Option A: State sync (fastest)
# Configure in config.toml:
# [statesync]
# enable = true
# rpc_servers = "rpc1.virtengine.com:26657,rpc2.virtengine.com:26657"
# trust_height = <recent_block_height>
# trust_hash = "<recent_block_hash>"

# Option B: Snapshot restore
# Stop node first
sudo systemctl stop virtengine

# Download and extract snapshot
wget https://snapshots.virtengine.com/virtengine-1/latest.tar.lz4
lz4 -d latest.tar.lz4 | tar -xf - -C ~/.virtengine/data/

# Restart
sudo systemctl start virtengine
```

---

## Validator Creation

### Step 1: Create Operator Keys

```bash
# Create operator key (SAVE THE MNEMONIC!)
virtengine keys add operator --keyring-backend file

# CRITICAL: Backup the mnemonic phrase securely
# Never store it digitally or share with anyone

# Display address
virtengine keys show operator -a
# virtengine1xxxxx...

# Get funds (testnet faucet or transfer)
# Verify balance
virtengine query bank balances $(virtengine keys show operator -a)
```

### Step 2: Create Validator Transaction

```bash
# Get validator consensus pubkey
VALIDATOR_PUBKEY=$(virtengine tendermint show-validator)

# Create validator
virtengine tx staking create-validator \
    --amount=100000000000uve \
    --pubkey=${VALIDATOR_PUBKEY} \
    --moniker="your-validator-name" \
    --chain-id=virtengine-1 \
    --commission-rate="0.10" \
    --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="1000000000" \
    --gas="auto" \
    --gas-adjustment="1.5" \
    --gas-prices="0.025uve" \
    --from=operator \
    --keyring-backend=file \
    --identity="<keybase-identity>" \
    --website="https://your-website.com" \
    --details="Description of your validator" \
    --security-contact="security@your-domain.com"
```

### Step 3: Verify Validator Status

```bash
# Check validator status
virtengine query staking validator $(virtengine keys show operator --bech val -a)

# Check if in active set
virtengine query staking validators --status bonded | grep -A5 "your-validator-name"

# Check signing info (after some blocks)
virtengine query slashing signing-info $(virtengine tendermint show-validator)
```

---

## Identity Verification Setup

Validators participate in VEID identity scoring by running ML inference.

### Step 1: Install TensorFlow Runtime

```bash
# Create Python virtual environment
python3 -m venv ~/.virtengine/venv
source ~/.virtengine/venv/bin/activate

# Install TensorFlow (pinned version for determinism)
pip install tensorflow==2.15.0

# Install additional dependencies
pip install numpy==1.24.3 pillow==10.0.0
```

### Step 2: Download Pinned Model

```bash
# Create models directory
mkdir -p ~/.virtengine/models

# Download model
wget -O ~/.virtengine/models/veid_scorer_v1.0.0.h5 \
    https://models.virtengine.com/veid_scorer_v1.0.0.h5

# Verify model hash
EXPECTED_MODEL_HASH="<official_model_hash>"
ACTUAL_HASH=$(sha256sum ~/.virtengine/models/veid_scorer_v1.0.0.h5 | cut -d' ' -f1)
if [ "$EXPECTED_MODEL_HASH" != "$ACTUAL_HASH" ]; then
    echo "ERROR: Model hash mismatch!"
    exit 1
fi
echo "Model verified successfully"
```

### Step 3: Configure Identity Module

Create `~/.virtengine/config/identity.toml`:

```toml
[identity]
# Enable identity verification
enabled = true

# Model path
model_path = "/home/validator/.virtengine/models/veid_scorer_v1.0.0.h5"
model_version = "v1.0.0"

# Scoring configuration
scoring_timeout = "5s"
max_concurrent_scoring = 4

# Determinism settings (CRITICAL for consensus)
force_cpu = true
random_seed = 42
deterministic_ops = true

# Log redaction (NEVER log plaintext identity data)
log_redaction = true
```

### Step 4: Create Validator Identity Key

```bash
# Generate identity key for decryption
virtengine keys add validator-identity --keyring-backend file

# Register key on-chain
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(virtengine keys show validator-identity --pubkey) \
    --label "Validator Identity Key v1" \
    --from operator \
    --keyring-backend file

# Verify registration
virtengine query encryption recipient-keys $(virtengine keys show operator -a)
```

---

## Monitoring Configuration

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'virtengine-validator'
    static_configs:
      - targets: ['localhost:26660']
    metrics_path: /metrics
    
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
```

### Key Metrics to Monitor

| Metric | Warning | Critical | Description |
|--------|---------|----------|-------------|
| `tendermint_consensus_height` | Behind > 10 | Behind > 100 | Block height |
| `tendermint_consensus_validators` | Changed | 0 | Validator count |
| `tendermint_consensus_missing_validators` | > 1 | > 10 | Missing validators |
| `tendermint_consensus_byzantine_validators` | > 0 | > 0 | Byzantine validators |
| `virtengine_veid_scoring_duration_seconds` | p95 > 3s | p95 > 5s | Scoring latency |
| `virtengine_veid_scoring_errors_total` | Rate > 0.01/s | Rate > 0.1/s | Scoring errors |
| `node_filesystem_avail_bytes` | < 20% | < 10% | Disk space |
| `node_memory_MemAvailable_bytes` | < 20% | < 10% | Memory |

### Alert Rules

```yaml
# alerting_rules.yml
groups:
  - name: validator-critical
    rules:
      - alert: ValidatorDown
        expr: up{job="virtengine-validator"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Validator node is down"
          
      - alert: ValidatorJailed
        expr: virtengine_staking_jailed == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Validator has been jailed"

      - alert: MissedBlocksHigh
        expr: increase(tendermint_consensus_validator_missed_blocks[1h]) > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Validator missing blocks"

      - alert: IdentityScoringDown
        expr: rate(virtengine_veid_scoring_total[5m]) == 0
        for: 10m
        labels:
          severity: high
        annotations:
          summary: "Identity scoring stopped"
```

### Grafana Dashboard

Import dashboard ID: `virtengine-validator-1`

Key panels:
- Block height (current vs network)
- Voting power
- Missed blocks
- VEID scoring latency and throughput
- System resources (CPU, memory, disk, network)

---

## Daily Operations

### Daily Checklist

- [ ] Check validator is signing blocks
- [ ] Verify no missed blocks in last 24h
- [ ] Check disk space (> 30% free)
- [ ] Review error logs
- [ ] Verify peer connections (> 20)
- [ ] Check VEID scoring metrics

### Daily Commands

```bash
# Check validator signing status
virtengine query slashing signing-info $(virtengine tendermint show-validator)

# Check missed blocks
curl -s http://localhost:26660/metrics | grep missed_blocks

# Check peer count
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Check disk usage
df -h ~/.virtengine/data

# Check recent errors
journalctl -u virtengine --since "24 hours ago" | grep -i error | tail -20

# Check VEID scoring
curl -s http://localhost:26660/metrics | grep veid_scoring
```

### Weekly Checklist

- [ ] Review all alerts from the past week
- [ ] Check for available upgrades
- [ ] Verify backup completed successfully
- [ ] Review resource utilization trends
- [ ] Test restore procedure (monthly)

### Weekly Commands

```bash
# Check for pending upgrades
virtengine query upgrade plan

# Check governance proposals
virtengine query gov proposals --status voting_period

# Verify backup exists
ls -la /backups/virtengine/

# Check upgrade schedule
virtengine query upgrade applied
```

---

## Maintenance Procedures

### Restarting the Validator

```bash
# Graceful restart
sudo systemctl restart virtengine

# Wait for catch-up
watch -n 5 'virtengine status 2>&1 | jq ".SyncInfo.catching_up"'

# Verify signing resumed
virtengine query slashing signing-info $(virtengine tendermint show-validator)
```

### Pruning Old Data

```bash
# Stop validator
sudo systemctl stop virtengine

# Run pruning
virtengine prune all

# Restart
sudo systemctl start virtengine
```

### Rotating Logs

```bash
# Configure logrotate
cat > /etc/logrotate.d/virtengine << EOF
/var/log/virtengine/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 validator validator
}
EOF
```

### Key Rotation

```bash
# 1. Generate new key
virtengine keys add validator-identity-v2 --keyring-backend file

# 2. Register new key on-chain
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(virtengine keys show validator-identity-v2 --pubkey) \
    --label "Validator Identity Key v2" \
    --from operator

# 3. Wait for grace period (48 hours)

# 4. Update config to use new key
# Edit ~/.virtengine/config/identity.toml

# 5. Restart validator
sudo systemctl restart virtengine

# 6. Revoke old key
virtengine tx encryption revoke-recipient-key \
    --fingerprint <old_key_fingerprint> \
    --from operator
```

---

## Security Hardening

### Firewall Configuration

```bash
# Install UFW
sudo apt install ufw

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow P2P (public)
sudo ufw allow 26656/tcp comment "Tendermint P2P"

# Allow SSH (restricted)
sudo ufw allow from YOUR_IP to any port 22

# Allow Prometheus (from monitoring only)
sudo ufw allow from MONITORING_IP to any port 26660

# Enable firewall
sudo ufw enable

# Verify rules
sudo ufw status verbose
```

### SSH Hardening

```bash
# /etc/ssh/sshd_config
Port 22022
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AllowUsers validator
MaxAuthTries 3
LoginGraceTime 30
```

### Key Security

> ⚠️ **CRITICAL SECURITY REQUIREMENTS**

1. **Mnemonic Security**
   - Never store digitally
   - Use metal backup plates
   - Store in secure physical location
   - Consider multi-location storage

2. **HSM Integration** (Production)
   ```bash
   # Configure HSM backend
   virtengine config keyring-backend ledger
   ```

3. **Key Encryption**
   ```bash
   # Always use encrypted keyring
   virtengine keys add operator --keyring-backend file
   # Password will be required for each operation
   ```

### Fail2Ban Configuration

```bash
# Install fail2ban
sudo apt install fail2ban

# Configure for SSH
cat > /etc/fail2ban/jail.local << EOF
[sshd]
enabled = true
port = 22022
maxretry = 3
bantime = 86400
findtime = 600
EOF

sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

---

## Sentry Node Architecture

For production validators, use sentry nodes to protect the validator.

### Architecture Overview

```
                    Internet
                       │
         ┌─────────────┴─────────────┐
         │                           │
    ┌────▼────┐                ┌────▼────┐
    │ Sentry  │                │ Sentry  │
    │  Node 1 │                │  Node 2 │
    └────┬────┘                └────┬────┘
         │     Private Network      │
         └───────────┬──────────────┘
                     │
              ┌──────▼──────┐
              │  Validator  │
              │    Node     │
              └─────────────┘
```

### Sentry Node Configuration

```toml
# Sentry config.toml
[p2p]
pex = true
persistent_peers = "<validator_node_id>@<validator_private_ip>:26656"
private_peer_ids = "<validator_node_id>"
unconditional_peer_ids = "<validator_node_id>"
addr_book_strict = false
```

### Validator Configuration with Sentries

```toml
# Validator config.toml
[p2p]
pex = false
persistent_peers = "<sentry1_id>@<sentry1_ip>:26656,<sentry2_id>@<sentry2_ip>:26656"
addr_book_strict = false
```

### Network Configuration

```bash
# On validator (private network only)
sudo ufw allow from SENTRY1_IP to any port 26656
sudo ufw allow from SENTRY2_IP to any port 26656
sudo ufw deny 26656/tcp  # Block public P2P

# On sentries (public-facing)
sudo ufw allow 26656/tcp  # Allow public P2P
```

---

## Troubleshooting

### Validator Not Signing

```bash
# 1. Check if synced
virtengine status | jq '.SyncInfo.catching_up'

# 2. Check if in active set
virtengine query staking validators --status bonded | grep $(virtengine keys show operator -a)

# 3. Check if jailed
virtengine query slashing signing-info $(virtengine tendermint show-validator) | grep jailed

# 4. Unjail if needed
virtengine tx slashing unjail --from operator
```

### Node Out of Sync

```bash
# 1. Check block height vs network
NETWORK_HEIGHT=$(curl -s https://rpc.virtengine.com/status | jq -r '.result.sync_info.latest_block_height')
LOCAL_HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "Network: $NETWORK_HEIGHT, Local: $LOCAL_HEIGHT, Behind: $((NETWORK_HEIGHT - LOCAL_HEIGHT))"

# 2. If significantly behind, use state sync
# See Initial Setup > Step 6 for state sync configuration
```

### Memory Issues

```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -5

# Check Go memory stats
curl http://localhost:6060/debug/pprof/heap > heap.pprof
go tool pprof heap.pprof

# Adjust Go garbage collection
echo 'Environment="GOGC=50"' | sudo tee -a /etc/systemd/system/virtengine.service.d/override.conf
sudo systemctl daemon-reload
sudo systemctl restart virtengine
```

---

## Appendix: Command Reference

### Status Commands

```bash
virtengine status                                    # Node status
virtengine query staking validator <valoper>         # Validator info
virtengine query slashing signing-info <pubkey>      # Signing info
virtengine query staking validators --status bonded  # Active validators
```

### Key Commands

```bash
virtengine keys list                                 # List keys
virtengine keys show <name> -a                       # Show address
virtengine keys show <name> --bech val               # Show valoper address
virtengine tendermint show-validator                 # Show consensus pubkey
```

### Transaction Commands

```bash
virtengine tx staking create-validator ...           # Create validator
virtengine tx staking edit-validator ...             # Edit validator
virtengine tx slashing unjail ...                    # Unjail validator
virtengine tx staking delegate ...                   # Self-delegate
```

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
