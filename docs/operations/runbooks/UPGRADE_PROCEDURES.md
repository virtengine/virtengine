# Upgrade Procedures

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Overview](#overview)
2. [Upgrade Types](#upgrade-types)
3. [Pre-Upgrade Checklist](#pre-upgrade-checklist)
4. [Chain Upgrades (Governance)](#chain-upgrades-governance)
5. [Non-Breaking Upgrades](#non-breaking-upgrades)
6. [Emergency Hotfixes](#emergency-hotfixes)
7. [Provider Daemon Upgrades](#provider-daemon-upgrades)
8. [Model Upgrades (VEID)](#model-upgrades-veid)
9. [Rollback Procedures](#rollback-procedures)
10. [Post-Upgrade Verification](#post-upgrade-verification)

---

## Overview

VirtEngine upgrades are carefully coordinated to maintain network consensus and minimize disruption. This document covers all upgrade scenarios from routine updates to emergency hotfixes.

### Upgrade Principles

1. **Test thoroughly** - All upgrades tested on devnet/testnet first
2. **Coordinate timing** - Major upgrades at low-traffic periods
3. **Have rollback ready** - Always prepare rollback procedure
4. **Communicate clearly** - Notify all stakeholders in advance
5. **Monitor closely** - Enhanced monitoring during and after upgrades

---

## Upgrade Types

| Type | Breaking | Governance | Coordination | Example |
|------|----------|------------|--------------|---------|
| Chain upgrade | Yes | Required | Network-wide halt | v1.0 → v2.0 |
| Software patch | No | Not required | Rolling | v1.0.0 → v1.0.1 |
| Emergency hotfix | Maybe | Emergency | ASAP | Security fix |
| ML model update | No | Required | Validator coordination | Model v1 → v2 |
| Config change | No | Not required | Rolling | Parameter tuning |

---

## Pre-Upgrade Checklist

### 24 Hours Before

- [ ] Review upgrade documentation and changelog
- [ ] Download new binary and verify checksum
- [ ] Test upgrade on staging/testnet
- [ ] Prepare rollback binary
- [ ] Backup current configuration
- [ ] Notify team and stakeholders
- [ ] Schedule maintenance window (if needed)

### 1 Hour Before

- [ ] Verify backup completed
- [ ] Check current chain/service health
- [ ] Confirm team availability
- [ ] Verify monitoring alerts configured
- [ ] Prepare communication channels

### Immediately Before

- [ ] Final health check
- [ ] Announce start of maintenance
- [ ] Have rollback procedure ready

---

## Chain Upgrades (Governance)

Chain upgrades require governance proposal and network-wide coordination.

### Phase 1: Preparation

```bash
# 1. Check for pending upgrade
virtengine query upgrade plan

# Sample output:
# name: "v2.0.0"
# height: 5000000
# info: "https://github.com/virtengine/virtengine/releases/tag/v2.0.0"

# 2. Download new binary
UPGRADE_VERSION="v2.0.0"
wget "https://github.com/virtengine/virtengine/releases/download/${UPGRADE_VERSION}/virtengine_linux_amd64.tar.gz"

# 3. Verify checksum
sha256sum virtengine_linux_amd64.tar.gz
# Compare with published checksum

# 4. Extract to staging location
tar -xzf virtengine_linux_amd64.tar.gz
mv virtengine /usr/local/bin/virtengine-v2.0.0

# 5. Verify binary
/usr/local/bin/virtengine-v2.0.0 version
```

### Phase 2: Automatic Upgrade (Using Cosmovisor)

Cosmovisor handles automatic binary switching at upgrade height.

```bash
# Setup Cosmovisor (if not already configured)
export DAEMON_NAME=virtengine
export DAEMON_HOME=$HOME/.virtengine

# Create upgrade directory
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin

# Copy new binary
cp /usr/local/bin/virtengine-v2.0.0 $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin/virtengine

# Cosmovisor will automatically switch at upgrade height
```

### Phase 3: Manual Upgrade (Without Cosmovisor)

```bash
# 1. Monitor block height
watch -n 5 'virtengine status 2>&1 | jq -r ".SyncInfo.latest_block_height"'

# 2. When chain halts at upgrade height, stop node
sudo systemctl stop virtengine

# 3. Backup current binary
sudo cp /usr/local/bin/virtengine /usr/local/bin/virtengine-v1.0.0

# 4. Install new binary
sudo cp /usr/local/bin/virtengine-v2.0.0 /usr/local/bin/virtengine

# 5. Verify version
virtengine version

# 6. Start node
sudo systemctl start virtengine

# 7. Monitor logs
sudo journalctl -u virtengine -f
```

### Phase 4: Verification

```bash
# Check node is syncing with new version
virtengine status | jq '.NodeInfo.version'

# Check blocks producing
watch -n 5 'virtengine status 2>&1 | jq ".SyncInfo.latest_block_height"'

# Verify upgrade applied
virtengine query upgrade applied
```

### Upgrade Coordination Timeline

```
T-7 days:    Governance proposal submitted
T-3 days:    Voting ends, upgrade confirmed
T-24 hours:  Final preparation, binary ready
T-0:         Chain halts at upgrade height
T+5 min:     Majority validators upgraded
T+15 min:    Chain resumes
T+1 hour:    Full verification complete
```

---

## Non-Breaking Upgrades

Non-breaking upgrades (patches, minor versions) can be done as rolling upgrades without network coordination.

### Rolling Upgrade Procedure

```bash
#!/bin/bash
# rolling-upgrade.sh

NEW_VERSION="v1.0.1"
BINARY_URL="https://github.com/virtengine/virtengine/releases/download/${NEW_VERSION}/virtengine_linux_amd64.tar.gz"

# 1. Download and verify
wget "$BINARY_URL" -O /tmp/virtengine-${NEW_VERSION}.tar.gz
sha256sum /tmp/virtengine-${NEW_VERSION}.tar.gz

read -p "Checksum verified? (yes/no): " confirmed
if [ "$confirmed" != "yes" ]; then
    exit 1
fi

tar -xzf /tmp/virtengine-${NEW_VERSION}.tar.gz -C /tmp/
chmod +x /tmp/virtengine

# 2. Backup current binary
sudo cp /usr/local/bin/virtengine /usr/local/bin/virtengine.bak

# 3. Stop service
sudo systemctl stop virtengine

# 4. Install new binary
sudo mv /tmp/virtengine /usr/local/bin/virtengine

# 5. Start service
sudo systemctl start virtengine

# 6. Verify
sleep 10
virtengine version
virtengine status | jq '.SyncInfo.catching_up'
```

### Validator Rolling Upgrade Order

For validators, upgrade in order of voting power (smallest first):

```bash
# Get validator order
virtengine query staking validators --status bonded | jq -r '.validators | sort_by(.tokens | tonumber) | .[].description.moniker'

# Upgrade each validator one at a time
# Wait for each to rejoin consensus before proceeding to next
```

---

## Emergency Hotfixes

Emergency hotfixes address critical bugs or security vulnerabilities.

### Severity Assessment

| Severity | Response Time | Coordination |
|----------|---------------|--------------|
| Critical (security) | Immediate | Emergency call |
| High (consensus bug) | < 4 hours | Validator chat |
| Medium (functionality) | < 24 hours | Normal process |

### Emergency Hotfix Procedure

```bash
#!/bin/bash
# emergency-hotfix.sh

echo "=== EMERGENCY HOTFIX PROCEDURE ==="
echo "This procedure bypasses normal upgrade process"
echo ""

HOTFIX_VERSION=$1
if [ -z "$HOTFIX_VERSION" ]; then
    echo "Usage: $0 <hotfix-version>"
    exit 1
fi

# 1. Download hotfix (from trusted source only)
BINARY_URL="https://github.com/virtengine/virtengine/releases/download/${HOTFIX_VERSION}/virtengine_linux_amd64.tar.gz"

echo "Downloading $BINARY_URL..."
wget "$BINARY_URL" -O /tmp/hotfix.tar.gz

# 2. Verify checksum (CRITICAL - verify from multiple sources)
echo "CRITICAL: Verify checksum from official announcement"
sha256sum /tmp/hotfix.tar.gz

read -p "Checksum verified from official source? (yes/no): " confirmed
if [ "$confirmed" != "yes" ]; then
    echo "Aborting: Checksum not verified"
    exit 1
fi

# 3. Extract
tar -xzf /tmp/hotfix.tar.gz -C /tmp/

# 4. Backup current
sudo cp /usr/local/bin/virtengine /usr/local/bin/virtengine.pre-hotfix

# 5. Stop service
sudo systemctl stop virtengine

# 6. Apply hotfix
sudo mv /tmp/virtengine /usr/local/bin/virtengine

# 7. Restart
sudo systemctl start virtengine

# 8. Verify
sleep 5
echo "New version: $(virtengine version)"
echo "Sync status: $(virtengine status 2>&1 | jq '.SyncInfo.catching_up')"
```

### Emergency Communication Template

```
Subject: [URGENT] VirtEngine Emergency Hotfix Required

Severity: CRITICAL
Affected Versions: v1.0.0 - v1.0.2
Fixed Version: v1.0.3

IMMEDIATE ACTION REQUIRED

All validators and node operators must upgrade to v1.0.3 immediately.

Issue: [Brief description without exposing vulnerability details]
Impact: [Potential impact if not patched]

Upgrade Steps:
1. Download v1.0.3 from [official release URL]
2. Verify checksum: [SHA256SUM]
3. Stop node: sudo systemctl stop virtengine
4. Replace binary
5. Start node: sudo systemctl start virtengine

Verification:
virtengine version
# Should show: v1.0.3

Support:
- Discord: #validator-emergency
- Email: security@virtengine.com

Timeline:
- [TIME] Issue identified
- [TIME] Hotfix developed
- [TIME] This announcement sent
- [TIME+2H] Expected majority upgraded
```

---

## Provider Daemon Upgrades

Provider daemon upgrades are independent of chain upgrades.

### Standard Upgrade

```bash
#!/bin/bash
# upgrade-provider-daemon.sh

NEW_VERSION="v1.1.0"

# 1. Pause provider operations
echo "Pausing provider..."
virtengine tx provider set-status \
    --status PAUSED \
    --reason "Upgrade to ${NEW_VERSION}" \
    --from provider \
    --keyring-backend file \
    --yes

# Wait for in-flight operations
sleep 30

# 2. Download new binary
wget "https://github.com/virtengine/virtengine/releases/download/${NEW_VERSION}/provider-daemon_linux_amd64.tar.gz"
tar -xzf provider-daemon_linux_amd64.tar.gz

# 3. Stop daemon
sudo systemctl stop provider-daemon

# 4. Backup and replace
sudo cp /usr/local/bin/provider-daemon /usr/local/bin/provider-daemon.bak
sudo mv provider-daemon /usr/local/bin/

# 5. Apply any config migrations
provider-daemon migrate-config --config ~/.provider-daemon/config.yaml

# 6. Start daemon
sudo systemctl start provider-daemon

# 7. Verify health
sleep 10
curl -sk https://localhost:8443/health | jq

# 8. Resume operations
virtengine tx provider set-status \
    --status ACTIVE \
    --from provider \
    --keyring-backend file \
    --yes

echo "Upgrade complete"
```

### Zero-Downtime Upgrade (Kubernetes)

```bash
# Update image version
kubectl set image deployment/provider-daemon \
    provider-daemon=virtengine/provider-daemon:v1.1.0 \
    -n virtengine-system

# Monitor rollout
kubectl rollout status deployment/provider-daemon -n virtengine-system

# Verify
kubectl get pods -n virtengine-system -l app=provider-daemon
```

---

## Model Upgrades (VEID)

ML model upgrades require validator coordination to maintain determinism.

### Model Upgrade Governance

```bash
# 1. Check for model upgrade proposal
virtengine query gov proposals --status voting_period

# Sample proposal:
# {
#   "proposal_id": "42",
#   "content": {
#     "@type": "/virtengine.veid.v1.MsgUpdateModel",
#     "model_version": "v1.1.0",
#     "model_hash": "sha256:abc123...",
#     "effective_height": 6000000
#   }
# }
```

### Model Download and Verification

```bash
#!/bin/bash
# download-model.sh

MODEL_VERSION="v1.1.0"
MODEL_HASH="sha256:abc123..."
MODEL_URL="https://models.virtengine.com/veid_scorer_${MODEL_VERSION}.h5"

# 1. Download model
wget "$MODEL_URL" -O ~/.virtengine/models/veid_scorer_${MODEL_VERSION}.h5

# 2. Verify hash
ACTUAL_HASH="sha256:$(sha256sum ~/.virtengine/models/veid_scorer_${MODEL_VERSION}.h5 | cut -d' ' -f1)"
if [ "$ACTUAL_HASH" != "$MODEL_HASH" ]; then
    echo "ERROR: Model hash mismatch!"
    echo "Expected: $MODEL_HASH"
    echo "Actual: $ACTUAL_HASH"
    exit 1
fi
echo "Model hash verified"

# 3. Test model loads
python3 -c "import tensorflow as tf; tf.keras.models.load_model('$HOME/.virtengine/models/veid_scorer_${MODEL_VERSION}.h5')"
echo "Model loads successfully"
```

### Model Activation

```bash
# Model activates at specified block height
# Monitor current height
watch -n 10 'virtengine status 2>&1 | jq -r ".SyncInfo.latest_block_height"'

# At activation height, update config
sed -i "s/model_path = .*/model_path = \"\/home\/validator\/.virtengine\/models\/veid_scorer_v1.1.0.h5\"/" \
    ~/.virtengine/config/identity.toml

# Restart validator
sudo systemctl restart virtengine

# Verify model active
virtengine query veid model-status
```

---

## Rollback Procedures

### Chain Upgrade Rollback

If upgrade fails and chain cannot produce blocks:

```bash
#!/bin/bash
# rollback-upgrade.sh

echo "=== UPGRADE ROLLBACK ==="
echo "WARNING: This will revert to previous version"
echo ""

# 1. Stop node
sudo systemctl stop virtengine

# 2. Restore previous binary
sudo cp /usr/local/bin/virtengine.bak /usr/local/bin/virtengine
virtengine version

# 3. If state corrupted, rollback state
virtengine rollback --hard

# 4. Start node
sudo systemctl start virtengine

# 5. Monitor
sudo journalctl -u virtengine -f
```

### Provider Daemon Rollback

```bash
# 1. Stop daemon
sudo systemctl stop provider-daemon

# 2. Restore binary
sudo cp /usr/local/bin/provider-daemon.bak /usr/local/bin/provider-daemon

# 3. Revert config if needed
cp ~/.provider-daemon/config.yaml.bak ~/.provider-daemon/config.yaml

# 4. Start daemon
sudo systemctl start provider-daemon

# 5. Verify
curl -sk https://localhost:8443/health
```

### Model Rollback

```bash
# Revert to previous model version
sed -i "s/model_path = .*/model_path = \"\/home\/validator\/.virtengine\/models\/veid_scorer_v1.0.0.h5\"/" \
    ~/.virtengine/config/identity.toml

sudo systemctl restart virtengine

# Note: Model rollbacks may require governance proposal for network-wide coordination
```

---

## Post-Upgrade Verification

### Validator Verification Checklist

```bash
#!/bin/bash
# verify-validator-upgrade.sh

echo "=== Post-Upgrade Verification ==="

# 1. Version check
echo "1. Version"
virtengine version

# 2. Sync status
echo "2. Sync Status"
virtengine status | jq '.SyncInfo'

# 3. Validator signing
echo "3. Validator Signing"
virtengine query slashing signing-info $(virtengine tendermint show-validator)

# 4. Peer connectivity
echo "4. Peer Count"
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# 5. Consensus participation
echo "5. Consensus"
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state.height'

# 6. VEID scoring (if applicable)
echo "6. VEID Scoring"
virtengine query veid model-status

# 7. Memory usage
echo "7. Memory"
ps aux --sort=-%mem | grep virtengine | head -1

# 8. Recent errors
echo "8. Recent Errors"
journalctl -u virtengine --since "10 minutes ago" | grep -i error | tail -5 || echo "No errors"
```

### Provider Verification Checklist

```bash
#!/bin/bash
# verify-provider-upgrade.sh

echo "=== Provider Post-Upgrade Verification ==="

# 1. Daemon health
echo "1. Health"
curl -sk https://localhost:8443/health | jq

# 2. Version
echo "2. Version"
provider-daemon version

# 3. Bid engine
echo "3. Bid Engine"
provider-daemon bids status

# 4. Workloads
echo "4. Active Workloads"
provider-daemon workloads list --state running | wc -l

# 5. Chain connectivity
echo "5. Chain Connectivity"
virtengine query provider info $(virtengine keys show provider -a) | head -5

# 6. Recent errors
echo "6. Recent Errors"
journalctl -u provider-daemon --since "10 minutes ago" | grep -i error | tail -5 || echo "No errors"
```

### Network-Wide Verification

```bash
#!/bin/bash
# verify-network-upgrade.sh

echo "=== Network-Wide Upgrade Verification ==="

# 1. Check all validators on new version
echo "1. Validator Versions"
for v in validator-{1..10}; do
    version=$(curl -s http://$v:26657/status 2>/dev/null | jq -r '.result.node_info.version')
    echo "   $v: $version"
done

# 2. Check block production
echo "2. Block Production"
HEIGHT1=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
sleep 10
HEIGHT2=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "   Blocks in 10s: $((HEIGHT2 - HEIGHT1))"

# 3. Check active validators
echo "3. Active Validators"
virtengine query staking validators --status bonded | jq '.validators | length'

# 4. Check no jailed validators
echo "4. Jailed Validators"
virtengine query staking validators --status unbonding | jq '.validators | length'
```

---

## Appendix: Upgrade Automation

### Cosmovisor Setup

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Set environment
export DAEMON_NAME=virtengine
export DAEMON_HOME=$HOME/.virtengine
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
export DAEMON_RESTART_AFTER_UPGRADE=true
export UNSAFE_SKIP_BACKUP=false

# Initialize
cosmovisor init $(which virtengine)

# Directory structure:
# $DAEMON_HOME/cosmovisor/
# ├── current/ -> genesis/
# ├── genesis/
# │   └── bin/
# │       └── virtengine
# └── upgrades/
#     └── v2.0.0/
#         └── bin/
#             └── virtengine
```

### Systemd with Cosmovisor

```ini
# /etc/systemd/system/virtengine.service
[Unit]
Description=VirtEngine Node (Cosmovisor)
After=network.target

[Service]
Type=simple
User=validator
Environment="DAEMON_NAME=virtengine"
Environment="DAEMON_HOME=/home/validator/.virtengine"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
ExecStart=/home/validator/go/bin/cosmovisor run start
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
