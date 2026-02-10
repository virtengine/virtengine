# Validator Operations Lab

**Duration:** 4 hours  
**Prerequisites:** [Lab Environment Setup](lab-environment.md) completed  
**Difficulty:** Intermediate

---

## Table of Contents

1. [Overview](#overview)
2. [Learning Objectives](#learning-objectives)
3. [Exercise 1: Set Up a Local Validator Node](#exercise-1-set-up-a-local-validator-node)
4. [Exercise 2: Create and Fund Validator Account](#exercise-2-create-and-fund-validator-account)
5. [Exercise 3: Join Consensus and Produce Blocks](#exercise-3-join-consensus-and-produce-blocks)
6. [Exercise 4: Monitor Validator Metrics](#exercise-4-monitor-validator-metrics)
7. [Exercise 5: Perform Key Rotation](#exercise-5-perform-key-rotation)
8. [Troubleshooting](#troubleshooting)
9. [Summary](#summary)

---

## Overview

VirtEngine validators have dual responsibilities:
1. **Consensus Participation**: Propose and validate blocks using Tendermint BFT
2. **Identity Network Operation**: Decrypt and score identity verification requests (VEID)

This lab walks through the complete validator lifecycle from node setup to key rotation.

### Lab Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Validator Setup                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐                     ┌─────────────────┐            │
│  │   Your New      │    Joins Network    │    Existing     │            │
│  │   Validator     │  ─────────────────> │    Localnet     │            │
│  │   Node          │                     │    Validators   │            │
│  └────────┬────────┘                     └─────────────────┘            │
│           │                                                              │
│           │  Exposes                                                     │
│           v                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │  RPC :26657     │  │  Prometheus     │  │  Identity       │         │
│  │  P2P :26656     │  │  Metrics :26660 │  │  Scoring        │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Learning Objectives

By the end of this lab, you will be able to:

- [ ] Initialize and configure a validator node
- [ ] Create validator keys and secure them properly
- [ ] Submit a create-validator transaction
- [ ] Monitor validator performance using Prometheus and Grafana
- [ ] Perform secure key rotation
- [ ] Troubleshoot common validator issues

---

## Exercise 1: Set Up a Local Validator Node

### Objective
Initialize a new validator node and configure it to connect to the localnet.

### Duration
45 minutes

### Prerequisites
- Localnet running (`./scripts/localnet.sh status` shows healthy)
- VirtEngine binary built (`.cache/bin/virtengine`)

### Instructions

#### Step 1: Create Validator Directory

```bash
# Create separate directory for new validator
mkdir -p ~/validator-lab
cd ~/validator-lab

# Set binary path
export VIRTENGINE_BIN="$HOME/virtengine-lab/virtengine/.cache/bin/virtengine"

# Verify binary
$VIRTENGINE_BIN version
```

#### Step 2: Initialize Validator

```bash
# Initialize with a unique moniker
$VIRTENGINE_BIN init "Lab-Validator-1" --chain-id virtengine-localnet-1 --home ./validator-home

# Verify initialization
ls -la ./validator-home/config/
```

**Expected Output:**
```
drwx------  3 user user 4096 Jan 24 10:00 .
drwxrwxr-x  4 user user 4096 Jan 24 10:00 ..
-rw-------  1 user user  386 Jan 24 10:00 app.toml
-rw-------  1 user user 6186 Jan 24 10:00 config.toml
-rw-r--r--  1 user user  292 Jan 24 10:00 genesis.json
-rw-------  1 user user  141 Jan 24 10:00 node_key.json
-rw-------  1 user user  341 Jan 24 10:00 priv_validator_key.json
```

#### Step 3: Copy Genesis from Localnet

```bash
# Get genesis from running localnet
curl -s http://localhost:26657/genesis | jq '.result.genesis' > ./validator-home/config/genesis.json

# Verify genesis
$VIRTENGINE_BIN validate-genesis --home ./validator-home
```

**Expected Output:**
```
File at ./validator-home/config/genesis.json is a valid genesis file
```

#### Step 4: Configure Node Settings

Edit `./validator-home/config/config.toml`:

```bash
# Use sed to configure settings
cat > ./update-config.sh << 'EOF'
#!/bin/bash
CONFIG="./validator-home/config/config.toml"

# Get localnet node ID
NODE_ID=$(curl -s http://localhost:26657/status | jq -r '.result.node_info.id')

# Configure persistent peers (connect to localnet)
sed -i.bak "s/^persistent_peers =.*/persistent_peers = \"${NODE_ID}@127.0.0.1:26656\"/" $CONFIG

# Configure different ports (to avoid conflict with localnet)
sed -i.bak 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/127.0.0.1:36657"/' $CONFIG
sed -i.bak 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:36656"/' $CONFIG
sed -i.bak 's/prometheus_listen_addr = ":26660"/prometheus_listen_addr = ":36660"/' $CONFIG
sed -i.bak 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:7060"/' $CONFIG

# Enable Prometheus metrics
sed -i.bak 's/prometheus = false/prometheus = true/' $CONFIG

# Set moniker
sed -i.bak 's/moniker = ".*"/moniker = "Lab-Validator-1"/' $CONFIG

echo "Configuration updated!"
grep -E "^(persistent_peers|laddr|prometheus)" $CONFIG
EOF

chmod +x ./update-config.sh
./update-config.sh
```

Edit `./validator-home/config/app.toml`:

```bash
APP_CONFIG="./validator-home/config/app.toml"

# Configure different ports for REST and gRPC
sed -i.bak 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/localhost:2317"/' $APP_CONFIG
sed -i.bak 's/address = "0.0.0.0:9090"/address = "0.0.0.0:10090"/' $APP_CONFIG
sed -i.bak 's/address = "0.0.0.0:9091"/address = "0.0.0.0:10091"/' $APP_CONFIG

# Set minimum gas prices
sed -i.bak 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025uve"/' $APP_CONFIG
```

#### Step 5: Start Node in Sync Mode

```bash
# Start the node (foreground for now to watch sync)
$VIRTENGINE_BIN start --home ./validator-home 2>&1 | tee validator.log &

# Wait for it to start
sleep 10

# Check sync status
curl -s http://localhost:36657/status | jq '.result.sync_info'
```

**Expected Output:**
```json
{
  "latest_block_hash": "...",
  "latest_app_hash": "...",
  "latest_block_height": "150",
  "latest_block_time": "2026-01-24T12:00:00.000Z",
  "catching_up": false
}
```

### Verification Checklist

| Check | Command | Expected |
|-------|---------|----------|
| Node running | `ps aux \| grep virtengine` | Process visible |
| RPC responding | `curl http://localhost:36657/health` | Returns empty JSON |
| Syncing | `curl http://localhost:36657/status \| jq '.result.sync_info.catching_up'` | `false` |
| Peers connected | `curl http://localhost:36657/net_info \| jq '.result.n_peers'` | `>= 1` |

---

## Exercise 2: Create and Fund Validator Account

### Objective
Create validator operator keys and fund the account with tokens.

### Duration
45 minutes

### Instructions

#### Step 1: Create Validator Operator Key

```bash
# Create operator key with file-based keyring
$VIRTENGINE_BIN keys add validator-operator \
    --keyring-backend file \
    --home ./validator-home

# IMPORTANT: Save the mnemonic phrase securely!
# You'll see output like:
# - name: validator-operator
#   type: local
#   address: virtengine1abc123...
#   pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}'
#   mnemonic: "word1 word2 word3 ... word24"
```

> ⚠️ **IMPORTANT**: Write down the mnemonic phrase! You'll need it for recovery.

```bash
# Save the address
VALIDATOR_ADDR=$($VIRTENGINE_BIN keys show validator-operator -a --keyring-backend file --home ./validator-home)
echo "Validator address: $VALIDATOR_ADDR"
echo $VALIDATOR_ADDR > ./validator-address.txt
```

#### Step 2: Fund Validator Account

Use the localnet's pre-funded accounts to send tokens:

```bash
# Enter localnet shell
cd ~/virtengine-lab/virtengine
./scripts/localnet.sh shell

# Inside container - send tokens to new validator
VALIDATOR_ADDR="<paste-your-validator-address>"
virtengine tx bank send \
    validator $VALIDATOR_ADDR 100000000000uve \
    --keyring-backend test \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

# Exit container
exit
```

#### Step 3: Verify Balance

```bash
cd ~/validator-lab

# Check balance
$VIRTENGINE_BIN query bank balances $VALIDATOR_ADDR \
    --node http://localhost:26657

# Expected: 100000000000uve (100,000 VE with 6 decimals)
```

**Expected Output:**
```yaml
balances:
- amount: "100000000000"
  denom: uve
pagination:
  next_key: null
  total: "0"
```

#### Step 4: Create Validator Identity Key (for VEID)

```bash
# Create identity key for encrypting VEID data
$VIRTENGINE_BIN keys add validator-identity \
    --keyring-backend file \
    --home ./validator-home

# Save identity address
IDENTITY_ADDR=$($VIRTENGINE_BIN keys show validator-identity -a --keyring-backend file --home ./validator-home)
echo "Identity address: $IDENTITY_ADDR"
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Operator key created | `keys show validator-operator` works |
| Balance funded | Balance shows 100000000000uve |
| Identity key created | `keys show validator-identity` works |

---

## Exercise 3: Join Consensus and Produce Blocks

### Objective
Create the validator on-chain and begin participating in consensus.

### Duration
45 minutes

### Instructions

#### Step 1: Get Validator Public Key

```bash
# Get the tendermint validator public key
VALIDATOR_PUBKEY=$($VIRTENGINE_BIN tendermint show-validator --home ./validator-home)
echo "Validator pubkey: $VALIDATOR_PUBKEY"
```

#### Step 2: Create Validator Transaction

```bash
# Ensure node is synced
curl -s http://localhost:36657/status | jq '.result.sync_info.catching_up'
# Must return: false

# Create validator
$VIRTENGINE_BIN tx staking create-validator \
    --amount=50000000000uve \
    --pubkey=$VALIDATOR_PUBKEY \
    --moniker="Lab-Validator-1" \
    --chain-id=virtengine-localnet-1 \
    --commission-rate="0.10" \
    --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="1000000000" \
    --gas="auto" \
    --gas-adjustment="1.5" \
    --gas-prices="0.025uve" \
    --from=validator-operator \
    --keyring-backend=file \
    --home=./validator-home \
    --node=http://localhost:26657 \
    --identity="lab-validator-keybase-id" \
    --website="https://lab-validator.example.com" \
    --details="Lab validator for training purposes" \
    --yes

# Note the transaction hash
```

**Expected Output:**
```yaml
txhash: ABC123DEF456...
code: 0
```

#### Step 3: Verify Validator Creation

```bash
# Wait for transaction to be included
sleep 10

# Query validator
$VIRTENGINE_BIN query staking validators \
    --node http://localhost:26657 \
    --output json | jq '.validators[] | select(.description.moniker == "Lab-Validator-1")'
```

**Expected Output:**
```json
{
  "operator_address": "virtenginevaloper1...",
  "consensus_pubkey": {...},
  "jailed": false,
  "status": "BOND_STATUS_BONDED",
  "tokens": "50000000000",
  "delegator_shares": "50000000000.000000000000000000",
  "description": {
    "moniker": "Lab-Validator-1",
    "identity": "lab-validator-keybase-id",
    "website": "https://lab-validator.example.com",
    "details": "Lab validator for training purposes"
  },
  "commission": {
    "commission_rates": {
      "rate": "0.100000000000000000",
      "max_rate": "0.200000000000000000",
      "max_change_rate": "0.010000000000000000"
    }
  }
}
```

#### Step 4: Monitor Block Production

```bash
# Watch your validator signing blocks
while true; do
    BLOCK_HEIGHT=$(curl -s http://localhost:36657/status | jq -r '.result.sync_info.latest_block_height')
    BLOCK_INFO=$(curl -s http://localhost:26657/block?height=$BLOCK_HEIGHT)
    PROPOSER=$(echo $BLOCK_INFO | jq -r '.result.block.header.proposer_address')
    
    # Check if our validator proposed
    OUR_ADDR=$($VIRTENGINE_BIN tendermint show-address --home ./validator-home)
    
    echo "Block $BLOCK_HEIGHT - Proposer: $PROPOSER"
    if [ "$PROPOSER" == "$OUR_ADDR" ]; then
        echo "🎉 We proposed this block!"
    fi
    
    sleep 5
done
```

#### Step 5: Check Signing Status

```bash
# Get validator signing info
VALCONS_ADDR=$($VIRTENGINE_BIN tendermint show-address --home ./validator-home)
$VIRTENGINE_BIN query slashing signing-info $VALCONS_ADDR \
    --node http://localhost:26657
```

**Expected Output:**
```yaml
address: virtenginevalcons1...
start_height: "100"
index_offset: "50"
jailed_until: "1970-01-01T00:00:00Z"
tombstoned: false
missed_blocks_counter: "0"
```

### Verification Checklist

| Check | Command | Expected |
|-------|---------|----------|
| Validator exists | `query staking validators` | Shows Lab-Validator-1 |
| Status bonded | Check `status` field | `BOND_STATUS_BONDED` |
| Not jailed | Check `jailed` field | `false` |
| Signing blocks | `query slashing signing-info` | `missed_blocks_counter: "0"` |

---

## Exercise 4: Monitor Validator Metrics

### Objective
Set up comprehensive monitoring for your validator using Prometheus and Grafana.

### Duration
45 minutes

### Instructions

#### Step 1: Verify Prometheus Metrics Endpoint

```bash
# Check metrics are exposed
curl -s http://localhost:36660/metrics | head -20

# Key metrics to look for
curl -s http://localhost:36660/metrics | grep -E "(tendermint_consensus|tendermint_p2p)"
```

**Key Metrics:**
```
tendermint_consensus_height
tendermint_consensus_validators
tendermint_consensus_validator_power
tendermint_consensus_missing_validators
tendermint_p2p_peers
tendermint_p2p_peer_receive_bytes_total
```

#### Step 2: Add Validator to Prometheus

Create a Prometheus configuration file:

```bash
cat > prometheus-validator.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'lab-validator'
    static_configs:
      - targets: ['localhost:36660']
        labels:
          instance: 'lab-validator-1'
          network: 'virtengine-localnet-1'

  - job_name: 'localnet-validator'
    static_configs:
      - targets: ['localhost:26660']
        labels:
          instance: 'localnet-genesis'
          network: 'virtengine-localnet-1'
EOF
```

#### Step 3: Create Monitoring Dashboard Script

```bash
cat > monitor-validator.sh << 'EOF'
#!/bin/bash

# Validator monitoring script
VALIDATOR_RPC="http://localhost:36657"
VALIDATOR_METRICS="http://localhost:36660"

while true; do
    clear
    echo "═══════════════════════════════════════════════════════════════"
    echo "                  Lab Validator Monitor                         "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    # Get status
    STATUS=$(curl -s $VALIDATOR_RPC/status)
    
    # Parse info
    BLOCK_HEIGHT=$(echo $STATUS | jq -r '.result.sync_info.latest_block_height')
    CATCHING_UP=$(echo $STATUS | jq -r '.result.sync_info.catching_up')
    VOTING_POWER=$(echo $STATUS | jq -r '.result.validator_info.voting_power')
    PEERS=$(curl -s $VALIDATOR_RPC/net_info | jq -r '.result.n_peers')
    
    # Get metrics
    CONSENSUS_HEIGHT=$(curl -s $VALIDATOR_METRICS/metrics 2>/dev/null | grep 'tendermint_consensus_height{' | awk '{print $2}')
    
    echo "  Block Height:    $BLOCK_HEIGHT"
    echo "  Catching Up:     $CATCHING_UP"
    echo "  Voting Power:    $VOTING_POWER"
    echo "  Connected Peers: $PEERS"
    echo ""
    
    # Memory usage
    echo "  Process Stats:"
    ps aux | grep '[v]irtengine start' | awk '{printf "    CPU: %s%%  Memory: %s%%\n", $3, $4}'
    echo ""
    
    # Recent blocks
    echo "  Recent Blocks:"
    for i in $(seq 0 4); do
        HEIGHT=$((BLOCK_HEIGHT - i))
        BLOCK=$(curl -s "$VALIDATOR_RPC/block?height=$HEIGHT" 2>/dev/null)
        TIME=$(echo $BLOCK | jq -r '.result.block.header.time' | cut -d'T' -f2 | cut -d'.' -f1)
        TXS=$(echo $BLOCK | jq -r '.result.block.data.txs | length')
        echo "    Block $HEIGHT - Time: $TIME - Txs: $TXS"
    done
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  Press Ctrl+C to exit"
    
    sleep 5
done
EOF

chmod +x monitor-validator.sh
```

#### Step 4: Run Monitor

```bash
./monitor-validator.sh
```

**Expected Output:**
```
═══════════════════════════════════════════════════════════════
                  Lab Validator Monitor                         
═══════════════════════════════════════════════════════════════

  Block Height:    250
  Catching Up:     false
  Voting Power:    50000000000
  Connected Peers: 1

  Process Stats:
    CPU: 2.5%  Memory: 1.2%

  Recent Blocks:
    Block 250 - Time: 12:05:30 - Txs: 0
    Block 249 - Time: 12:05:25 - Txs: 1
    Block 248 - Time: 12:05:20 - Txs: 0
    Block 247 - Time: 12:05:15 - Txs: 0
    Block 246 - Time: 12:05:10 - Txs: 2

═══════════════════════════════════════════════════════════════
```

#### Step 5: Set Up Alerting Rules

Create alert rules file:

```bash
cat > validator-alerts.yml << 'EOF'
groups:
  - name: validator-alerts
    rules:
      # Validator down
      - alert: ValidatorDown
        expr: up{job="lab-validator"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Lab validator is down"
          description: "The lab validator has been unreachable for more than 1 minute."

      # Validator not synced
      - alert: ValidatorNotSynced
        expr: tendermint_consensus_height{job="lab-validator"} < tendermint_consensus_height{job="localnet-validator"} - 5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Lab validator is falling behind"
          description: "The lab validator is more than 5 blocks behind the network."

      # No peers
      - alert: ValidatorNoPeers
        expr: tendermint_p2p_peers{job="lab-validator"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Lab validator has no peers"
          description: "The lab validator is isolated with no peer connections."

      # High memory usage
      - alert: ValidatorHighMemory
        expr: process_resident_memory_bytes{job="lab-validator"} / 1024 / 1024 / 1024 > 4
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Lab validator high memory usage"
          description: "Memory usage exceeds 4GB."
EOF
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Metrics endpoint accessible | `curl localhost:36660/metrics` returns data |
| Monitor script runs | Shows block height and peers |
| Voting power visible | Non-zero voting power |

---

## Exercise 5: Perform Key Rotation

### Objective
Safely rotate validator keys following security best practices.

### Duration
45 minutes

### Instructions

#### Step 1: Backup Current Keys

```bash
# Create backup directory
mkdir -p ./key-backups/$(date +%Y%m%d)

# Backup validator keys
cp ./validator-home/config/priv_validator_key.json \
   ./key-backups/$(date +%Y%m%d)/priv_validator_key.json.bak

cp ./validator-home/config/node_key.json \
   ./key-backups/$(date +%Y%m%d)/node_key.json.bak

# Backup keyring
cp -r ./validator-home/keyring-file \
   ./key-backups/$(date +%Y%m%d)/keyring-file.bak

# Encrypt backup (production: use strong password)
tar -czf ./key-backups/$(date +%Y%m%d)/keys-backup.tar.gz \
    ./key-backups/$(date +%Y%m%d)/*.bak \
    ./key-backups/$(date +%Y%m%d)/keyring-file.bak

echo "Backup created at ./key-backups/$(date +%Y%m%d)/"
ls -la ./key-backups/$(date +%Y%m%d)/
```

#### Step 2: Create New Identity Key

```bash
# Generate new identity key
$VIRTENGINE_BIN keys add validator-identity-v2 \
    --keyring-backend file \
    --home ./validator-home

# Get new key details
NEW_IDENTITY_PUBKEY=$($VIRTENGINE_BIN keys show validator-identity-v2 --pubkey --keyring-backend file --home ./validator-home)
echo "New identity pubkey: $NEW_IDENTITY_PUBKEY"
```

#### Step 3: Register New Key On-Chain

```bash
# Register new encryption key
$VIRTENGINE_BIN tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key "$NEW_IDENTITY_PUBKEY" \
    --label "Validator Identity Key v2" \
    --from validator-operator \
    --keyring-backend file \
    --home ./validator-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-adjustment 1.5 \
    --gas-prices 0.025uve \
    --yes

# Wait for confirmation
sleep 10
```

#### Step 4: Verify New Key Registration

```bash
# Query registered keys
$VIRTENGINE_BIN query encryption recipient-keys \
    --node http://localhost:26657 \
    --output json | jq '.keys'
```

**Expected Output:**
```json
[
  {
    "fingerprint": "abc123...",
    "label": "Validator Identity Key v1",
    "algorithm": "X25519-XSalsa20-Poly1305",
    "status": "ACTIVE"
  },
  {
    "fingerprint": "def456...",
    "label": "Validator Identity Key v2",
    "algorithm": "X25519-XSalsa20-Poly1305",
    "status": "ACTIVE"
  }
]
```

#### Step 5: Update Configuration for New Key

```bash
# Update daemon config to use new key
# In production, this would be in identity.toml or provider config

echo "KEY_ROTATION_COMPLETE=true" >> ./validator-home/key-rotation.log
echo "OLD_KEY=validator-identity" >> ./validator-home/key-rotation.log
echo "NEW_KEY=validator-identity-v2" >> ./validator-home/key-rotation.log
echo "ROTATION_DATE=$(date -Iseconds)" >> ./validator-home/key-rotation.log
```

#### Step 6: Revoke Old Key (After Grace Period)

> ⚠️ **In production**: Wait 24-48 hours before revoking old key to ensure all systems have switched.

```bash
# Get old key fingerprint
OLD_FINGERPRINT="<fingerprint-from-step-4>"

# Revoke old key (demo - in production wait for grace period)
$VIRTENGINE_BIN tx encryption revoke-recipient-key \
    --fingerprint $OLD_FINGERPRINT \
    --from validator-operator \
    --keyring-backend file \
    --home ./validator-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-prices 0.025uve \
    --yes

echo "Old key revoked. Key rotation complete!"
```

#### Step 7: Document Rotation

```bash
cat > ./key-backups/$(date +%Y%m%d)/rotation-record.md << EOF
# Key Rotation Record

**Date:** $(date -Iseconds)
**Validator:** Lab-Validator-1
**Operator:** $(cat ./validator-address.txt)

## Actions Taken

1. Backed up existing keys
2. Generated new identity key (v2)
3. Registered new key on-chain
4. Updated configuration
5. Revoked old key after grace period

## Key Details

- **Old Key Label:** Validator Identity Key v1
- **New Key Label:** Validator Identity Key v2
- **Algorithm:** X25519-XSalsa20-Poly1305

## Verification

- [ ] New key appears in on-chain registry
- [ ] Old key shows REVOKED status
- [ ] Validator continues producing blocks
- [ ] Identity scoring works with new key

## Next Rotation

Scheduled for: $(date -d "+365 days" +%Y-%m-%d) (1 year)
EOF

cat ./key-backups/$(date +%Y%m%d)/rotation-record.md
```

### Verification Checklist

| Check | Expected |
|-------|----------|
| Backup created | Files exist in `./key-backups/` |
| New key registered | Shows in `query encryption recipient-keys` |
| Validator still signing | `missed_blocks_counter` unchanged |
| Rotation documented | `rotation-record.md` created |

---

## Troubleshooting

### Validator Not Joining Consensus

**Symptoms:** Validator created but voting power is 0

```bash
# Check if jailed
$VIRTENGINE_BIN query staking validator $(VALOPER_ADDR) --node http://localhost:26657 | grep jailed

# If jailed, unjail
$VIRTENGINE_BIN tx slashing unjail \
    --from validator-operator \
    --keyring-backend file \
    --home ./validator-home \
    --node http://localhost:26657 \
    --chain-id virtengine-localnet-1 \
    --gas auto \
    --gas-prices 0.025uve \
    --yes
```

### Node Not Syncing

**Symptoms:** `catching_up: true` for extended period

```bash
# Check peer count
curl -s http://localhost:36657/net_info | jq '.result.n_peers'

# If 0 peers, verify config
grep persistent_peers ./validator-home/config/config.toml

# Try adding more peers manually
# Edit config.toml with additional peers
```

### Transaction Failures

**Symptoms:** `out of gas` or `insufficient funds`

```bash
# Check balance
$VIRTENGINE_BIN query bank balances $VALIDATOR_ADDR --node http://localhost:26657

# Increase gas adjustment
$VIRTENGINE_BIN tx ... --gas-adjustment 2.0 --gas-prices 0.05uve
```

### Key Errors

**Symptoms:** `keyring-backend` errors

```bash
# List available keys
$VIRTENGINE_BIN keys list --keyring-backend file --home ./validator-home

# If key missing, restore from backup
cp ./key-backups/*/keyring-file.bak/* ./validator-home/keyring-file/
```

### High Missed Blocks

**Symptoms:** `missed_blocks_counter` increasing

```bash
# Check signing info
$VIRTENGINE_BIN query slashing signing-info $(VALCONS_ADDR) --node http://localhost:26657

# Common causes:
# 1. Node not synced - check sync status
# 2. Network issues - check peer connections
# 3. Resource constraints - check CPU/memory
# 4. Clock skew - sync system time with NTP
```

---

## Summary

In this lab, you learned how to:

1. **Initialize a validator node** - Set up configuration and connect to network
2. **Create and fund accounts** - Generate secure keys and acquire tokens
3. **Join consensus** - Submit create-validator transaction and produce blocks
4. **Monitor operations** - Use Prometheus metrics and custom scripts
5. **Rotate keys securely** - Follow key rotation best practices

### Key Takeaways

- Always backup keys before any operation
- Monitor `missed_blocks_counter` to catch issues early
- Use file-based keyring with strong passwords (or HSM in production)
- Document all key rotations for audit trail
- Test recovery procedures regularly

### Cleanup

```bash
# Stop validator
pkill -f "virtengine start --home ./validator-home"

# Optionally remove data (keeps backups)
rm -rf ./validator-home/data
```

---

## Next Steps

Continue to the following labs:

1. **[Provider Operations Lab](provider-ops-lab.md)** - Set up provider daemon
2. **[Incident Simulation Lab](incident-simulation-lab.md)** - Practice incident response
3. **[Security Assessment Lab](security-assessment-lab.md)** - Security auditing

---

*Lab Version: 1.0.0*  
*Last Updated: 2026-01-24*  
*Maintainer: VirtEngine Training Team*
