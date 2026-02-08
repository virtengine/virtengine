# VirtEngine Validator Onboarding Guide

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-802

---

## Table of Contents

1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Setup Guide](#setup-guide)
4. [Identity Verification Responsibilities](#identity-verification-responsibilities)
5. [Model and Version Pinning](#model-and-version-pinning)
6. [Monitoring](#monitoring)
7. [Maintenance](#maintenance)
8. [Security](#security)

---

## Overview

VirtEngine validators have dual responsibilities:
1. **Consensus Participation**: Propose and validate blocks using Tendermint BFT
2. **Identity Network Operation**: Decrypt and score identity verification requests

This guide covers setup, operation, and ongoing maintenance for validator nodes.

## Requirements

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 8 cores | 16+ cores |
| RAM | 32 GB | 64 GB |
| Storage | 1 TB NVMe SSD | 2 TB NVMe SSD |
| Network | 100 Mbps | 1 Gbps |

### Software Requirements

- Ubuntu 22.04 LTS or equivalent
- Go 1.25.5+
- TensorFlow 2.x runtime (for identity scoring)
- Docker (optional, for containerized deployment)

### Network Requirements

- Static IP or reliable dynamic DNS
- Ports: 26656 (P2P), 26657 (RPC), 9090 (gRPC), 1317 (REST)
- DDoS protection recommended
- Sentry node architecture recommended for production

## Setup Guide

### Step 1: Install VirtEngine

```bash
# Download latest release
wget https://github.com/virtengine/virtengine/releases/download/v1.0.0/virtengine_linux_amd64.tar.gz
tar -xzf virtengine_linux_amd64.tar.gz
sudo mv virtengine /usr/local/bin/

# Verify installation
virtengine version
```

### Step 2: Initialize Node

```bash
# Initialize with moniker
virtengine init "MyValidator" --chain-id virtengine-1

# Download genesis
wget -O ~/.virtengine/config/genesis.json \
    https://raw.githubusercontent.com/virtengine/networks/main/virtengine-1/genesis.json

# Verify genesis hash
sha256sum ~/.virtengine/config/genesis.json
# Expected: <official_genesis_hash>
```

### Step 3: Configure Node

Edit `~/.virtengine/config/config.toml`:

```toml
# Moniker (your validator name)
moniker = "MyValidator"

# Seeds (official seed nodes)
seeds = "seed1@ip:26656,seed2@ip:26656"

# Persistent peers
persistent_peers = "peer1@ip:26656,peer2@ip:26656"

# Enable Prometheus metrics
prometheus = true

# RPC configuration
[rpc]
laddr = "tcp://127.0.0.1:26657"

# P2P configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = "tcp://YOUR_PUBLIC_IP:26656"
max_num_inbound_peers = 100
max_num_outbound_peers = 50
```

Edit `~/.virtengine/config/app.toml`:

```toml
# Minimum gas prices
minimum-gas-prices = "0.025uve"

# Enable API
[api]
enable = true
address = "tcp://127.0.0.1:1317"

# Enable gRPC
[grpc]
enable = true
address = "0.0.0.0:9090"

# State sync (for faster initial sync)
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2
```

### Step 4: Create Validator Keys

```bash
# Create operator key
virtengine keys add validator --keyring-backend file

# IMPORTANT: Backup your mnemonic securely!
# Never share this with anyone

# Display address
virtengine keys show validator -a
```

### Step 5: Sync Node

```bash
# Start node (foreground for initial sync)
virtengine start

# Or use state sync for faster sync
# Configure state-sync in config.toml first

# Check sync status
virtengine status | jq '.SyncInfo'
```

### Step 6: Create Validator Transaction

Once synced, create your validator:

```bash
virtengine tx staking create-validator \
    --amount=100000000000uve \
    --pubkey=$(virtengine tendermint show-validator) \
    --moniker="MyValidator" \
    --chain-id=virtengine-1 \
    --commission-rate="0.10" \
    --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="1000000000" \
    --gas="auto" \
    --gas-adjustment="1.5" \
    --gas-prices="0.025uve" \
    --from=validator \
    --keyring-backend=file \
    --identity="<keybase-id>" \
    --website="https://myvalidator.com" \
    --details="Description of your validator"
```

## Identity Verification Responsibilities

Validators participate in the identity verification network by:

1. **Decrypting Identity Scopes**: Using validator keys to decrypt encrypted identity data
2. **Running ML Scoring**: Executing pinned TensorFlow models to compute identity scores
3. **Voting on Scores**: Participating in consensus to finalize scores

### Identity Scoring Setup

```bash
# Install TensorFlow runtime
pip install tensorflow==2.15.0

# Download pinned model
wget -O ~/.virtengine/models/veid_scorer_v1.0.0.h5 \
    https://models.virtengine.com/veid_scorer_v1.0.0.h5

# Verify model hash
sha256sum ~/.virtengine/models/veid_scorer_v1.0.0.h5
# Expected: <model_hash>
```

### Configure Identity Module

Edit `~/.virtengine/config/identity.toml`:

```toml
[identity]
# Enable identity verification
enabled = true

# Model path
model_path = "/home/validator/.virtengine/models/veid_scorer_v1.0.0.h5"

# Scoring timeout
scoring_timeout = "5s"

# Maximum concurrent scoring
max_concurrent_scoring = 4

# Log redaction (CRITICAL: never log plaintext identity data)
log_redaction = true
```

### Decryption Key Management

```bash
# Generate validator identity key (for encrypting identity scopes)
virtengine keys add validator-identity --keyring-backend file

# Register key on-chain
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(virtengine keys show validator-identity --pubkey) \
    --label "Validator Identity Key v1" \
    --from validator \
    --keyring-backend file
```

## Model and Version Pinning

### Current Pinned Versions

| Component | Version | Hash |
|-----------|---------|------|
| VirtEngine Binary | v1.0.0 | `<binary_hash>` |
| VEID Scorer Model | v1.0.0 | `<model_hash>` |
| TensorFlow Runtime | 2.15.0 | N/A |

### Upgrade Process

1. **Governance Proposal**: Model upgrades are proposed via governance
2. **Download New Model**: After proposal passes, download new model
3. **Verify Hash**: Confirm model hash matches governance-approved hash
4. **Coordinate Upgrade**: Switch at specified block height

```bash
# Check for pending upgrades
virtengine query upgrade plan

# Download new model when approved
wget -O ~/.virtengine/models/veid_scorer_v1.1.0.h5 \
    https://models.virtengine.com/veid_scorer_v1.1.0.h5

# Verify hash matches governance proposal
sha256sum ~/.virtengine/models/veid_scorer_v1.1.0.h5
```

## Monitoring

### Prometheus Metrics

Expose Prometheus metrics on port 26660:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'virtengine-validator'
    static_configs:
      - targets: ['localhost:26660']
```

### Key Metrics to Monitor

| Metric | Warning | Critical |
|--------|---------|----------|
| `tendermint_consensus_height` | Lagging > 10 blocks | Lagging > 100 blocks |
| `tendermint_consensus_validators` | Missing | 0 |
| `virtengine_veid_scoring_duration_seconds` | p95 > 3s | p95 > 5s |
| `virtengine_veid_scoring_errors_total` | Rate > 0.01/s | Rate > 0.1/s |
| `process_resident_memory_bytes` | > 80% limit | > 95% limit |
| `process_cpu_seconds_total` | > 80% | > 95% |

### Alerting Rules

```yaml
# alerting_rules.yml
groups:
  - name: validator-alerts
    rules:
      - alert: ValidatorDown
        expr: up{job="virtengine-validator"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Validator node is down"
          
      - alert: ValidatorMissedBlocks
        expr: increase(tendermint_consensus_missed_blocks_total[1h]) > 10
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Validator missing blocks"
          
      - alert: IdentityScoringSlowdown
        expr: histogram_quantile(0.95, rate(virtengine_veid_scoring_duration_seconds_bucket[5m])) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Identity scoring latency elevated"
```

### Grafana Dashboard

Import the VirtEngine Validator Dashboard:
- Dashboard ID: `virtengine-validator-1`
- Panels: Block height, voting power, identity scoring metrics, system resources

## Maintenance

### Regular Tasks

| Task | Frequency |
|------|-----------|
| Check disk space | Daily |
| Review logs for errors | Daily |
| Verify peer connections | Daily |
| Check for upgrades | Weekly |
| Backup validator keys | Monthly |
| Test restore procedure | Quarterly |

### Upgrading VirtEngine

```bash
# Check for upgrade signal
virtengine query upgrade plan

# When upgrade is scheduled:
# 1. Wait for halt height
# 2. Stop node
# 3. Replace binary
# 4. Restart node

# Upgrade binary
wget https://github.com/virtengine/virtengine/releases/download/v1.1.0/virtengine_linux_amd64.tar.gz
tar -xzf virtengine_linux_amd64.tar.gz
sudo mv virtengine /usr/local/bin/virtengine

# Restart
sudo systemctl restart virtengine
```

### Key Rotation

Rotate validator identity keys periodically:

```bash
# Generate new key
virtengine keys add validator-identity-v2 --keyring-backend file

# Register new key
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(virtengine keys show validator-identity-v2 --pubkey) \
    --label "Validator Identity Key v2" \
    --from validator

# After grace period, revoke old key
virtengine tx encryption revoke-recipient-key \
    --fingerprint <old_key_fingerprint> \
    --from validator
```

## Security

### Key Security

> ⚠️ **CRITICAL SECURITY REQUIREMENTS**

1. **Never share validator keys**: Compromise leads to slashing
2. **Use hardware security modules (HSM)**: For production validators
3. **Enable key encryption**: Always use encrypted keyring
4. **Backup securely**: Store mnemonic in secure offline location
5. **Separate operational keys**: Use different keys for different purposes

### Firewall Configuration

```bash
# Allow P2P
ufw allow 26656/tcp

# Restrict RPC to localhost
ufw allow from 127.0.0.1 to any port 26657

# Restrict API to localhost
ufw allow from 127.0.0.1 to any port 1317

# Allow Prometheus (from monitoring server only)
ufw allow from MONITORING_IP to any port 26660
```

### Sentry Node Architecture

For production validators, use sentry nodes:

```
Internet <-> Sentry Nodes <-> Validator Node (private network)
```

Validator `config.toml`:
```toml
pex = false
persistent_peers = "sentry1@private_ip:26656,sentry2@private_ip:26656"
```

### Incident Response

If you suspect key compromise:

1. **Immediately**: Unjail validator if jailed
2. **Rotate Keys**: Generate new validator keys
3. **Notify**: Contact VirtEngine security team
4. **Document**: Log all actions taken
5. **Review**: Conduct post-incident analysis

---

## Support

- Discord: [VirtEngine Validators](https://discord.gg/virtengine)
- Forum: [forum.virtengine.com](https://forum.virtengine.com)
- Security: security@virtengine.com

---

## Appendix: Systemd Service

```ini
# /etc/systemd/system/virtengine.service
[Unit]
Description=VirtEngine Validator Node
After=network.target

[Service]
Type=simple
User=validator
ExecStart=/usr/local/bin/virtengine start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/home/validator/.virtengine

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable virtengine
sudo systemctl start virtengine
sudo journalctl -u virtengine -f
```
