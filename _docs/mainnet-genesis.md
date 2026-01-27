# VirtEngine Mainnet Genesis Guide

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-802

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Genesis Generation](#genesis-generation)
4. [Genesis Parameters](#genesis-parameters)
5. [Network Parameters](#network-parameters)
6. [Dry-Run Ceremony](#dry-run-ceremony)
7. [Mainnet Launch Checklist](#mainnet-launch-checklist)

---

## Overview

This document describes the process for generating a deterministic `genesis.json` file for VirtEngine mainnet launch. The genesis file initializes the blockchain state with:

- Initial validator set
- Token distribution
- Module parameters
- Role assignments
- Approved client allowlist

## Prerequisites

### Required Tools

```bash
# VirtEngine binary
virtengine version
# Expected: virtengine v1.0.0

# JSON processor
jq --version

# Validator key generation
virtengine keys add genesis-admin --keyring-backend file
```

### Required Information

| Item | Description |
|------|-------------|
| Chain ID | Unique identifier for the network (e.g., `virtengine-1`) |
| Genesis Time | Coordinated UTC timestamp for network start |
| Initial Validators | List of genesis validators with staking amounts |
| Token Distribution | Allocation for community, team, treasury |
| Approved Clients | Initial approved client public keys |

## Genesis Generation

### Step 1: Initialize Chain

```bash
# Initialize with chain ID
virtengine init mainnet-genesis --chain-id virtengine-1

# This creates:
# - ~/.virtengine/config/genesis.json (template)
# - ~/.virtengine/config/node_key.json
# - ~/.virtengine/config/priv_validator_key.json
```

### Step 2: Configure Accounts

```bash
# Add genesis accounts
virtengine genesis add-genesis-account \
    virtengine1treasury... \
    1000000000000uve \
    --keyring-backend file

# Add team allocation
virtengine genesis add-genesis-account \
    virtengine1team... \
    500000000000uve \
    --vesting-amount 500000000000uve \
    --vesting-start-time $(date -d "+1 year" +%s) \
    --vesting-end-time $(date -d "+4 years" +%s)

# Add community pool seed
virtengine genesis add-genesis-account \
    virtengine1community... \
    200000000000uve
```

### Step 3: Add Genesis Validators

```bash
# Collect gentxs from validators
# Each validator runs:
virtengine genesis gentx \
    validator-key \
    100000000000uve \
    --chain-id virtengine-1 \
    --commission-rate 0.10 \
    --commission-max-rate 0.20 \
    --commission-max-change-rate 0.01 \
    --min-self-delegation 1000000000uve \
    --moniker "Validator Name" \
    --identity "keybase-id" \
    --website "https://validator.example.com" \
    --details "Genesis validator for VirtEngine"

# Collect all gentxs and merge
virtengine genesis collect-gentxs
```

### Step 4: Set Module Parameters

```bash
# Edit genesis.json with custom parameters
jq '.app_state.veid.params = {
    "scoring_delay_blocks": 10,
    "min_identity_score": 0,
    "max_identity_score": 100,
    "approved_clients": []
}' ~/.virtengine/config/genesis.json > genesis_tmp.json

mv genesis_tmp.json ~/.virtengine/config/genesis.json
```

### Step 5: Validate Genesis

```bash
# Validate the genesis file
virtengine genesis validate-genesis

# Check deterministic hash
sha256sum ~/.virtengine/config/genesis.json
# Must match expected: <genesis_hash>
```

## Genesis Parameters

### Staking Parameters

```json
{
    "app_state": {
        "staking": {
            "params": {
                "unbonding_time": "1814400s",
                "max_validators": 100,
                "max_entries": 7,
                "historical_entries": 10000,
                "bond_denom": "uve",
                "min_commission_rate": "0.050000000000000000"
            }
        }
    }
}
```

### VEID Module Parameters

```json
{
    "app_state": {
        "veid": {
            "params": {
                "scoring_delay_blocks": 10,
                "min_identity_score": 0,
                "max_identity_score": 100,
                "score_validity_blocks": 100000,
                "min_validator_identity_score": 50,
                "approved_clients": [
                    {
                        "public_key": "<base64_pubkey>",
                        "label": "VirtEngine Mobile App v1.0",
                        "added_at": "2026-01-24T00:00:00Z"
                    }
                ]
            }
        }
    }
}
```

### MFA Module Parameters

```json
{
    "app_state": {
        "mfa": {
            "params": {
                "challenge_timeout": "300s",
                "session_timeout": "900s",
                "max_factors_per_account": 10,
                "sensitive_tx_types": [
                    "ACCOUNT_RECOVERY",
                    "KEY_ROTATION",
                    "HIGH_VALUE_TRANSFER",
                    "PROVIDER_REGISTRATION",
                    "DELEGATION_CHANGE"
                ],
                "high_value_threshold": "1000000000uve"
            }
        }
    }
}
```

### Encryption Module Parameters

```json
{
    "app_state": {
        "encryption": {
            "params": {
                "max_recipients_per_envelope": 100,
                "max_keys_per_account": 10,
                "allowed_algorithms": [
                    "X25519-XSalsa20-Poly1305"
                ],
                "require_signature": true
            }
        }
    }
}
```

### Market Module Parameters

```json
{
    "app_state": {
        "market": {
            "params": {
                "order_max_bids": 20,
                "bid_timeout": "3600s",
                "min_offering_deposit": "10000000uve",
                "marketplace_fee_rate": "0.025000000000000000",
                "min_identity_score_customer": 30,
                "min_identity_score_provider": 50
            }
        }
    }
}
```

### HPC Module Parameters

```json
{
    "app_state": {
        "hpc": {
            "params": {
                "max_job_duration": "604800s",
                "job_submission_deposit": "1000000uve",
                "accounting_interval": "3600s",
                "reward_distribution_delay": 100
            }
        }
    }
}
```

## Network Parameters

### Governance-Controlled Parameters

These parameters can be changed via governance proposals after launch:

| Parameter | Initial Value | Description |
|-----------|--------------|-------------|
| `veid.approved_clients` | See above | Approved client allowlist |
| `mfa.sensitive_tx_types` | See above | Transactions requiring MFA |
| `market.marketplace_fee_rate` | 2.5% | Platform fee rate |
| `encryption.allowed_algorithms` | X25519-XSalsa20-Poly1305 | Allowed encryption algorithms |
| `hpc.reward_distribution_delay` | 100 blocks | Delay before HPC rewards |

### Admin-Controlled Parameters

These require admin multisig:

| Parameter | Description |
|-----------|-------------|
| Emergency halt | Pause chain operations |
| Approved client revocation | Immediate client key revocation |
| Validator slashing override | Emergency slashing adjustment |

## Dry-Run Ceremony

### Testnet Rehearsal

Before mainnet, conduct a dry-run on testnet:

1. **Genesis Generation**
   - Generate genesis with production parameters
   - Verify deterministic hash across all participants

2. **Validator Coordination**
   - All genesis validators submit gentxs
   - Verify signatures and staking amounts
   - Confirm validator set matches expectations

3. **Network Start**
   - Coordinate genesis time
   - Start nodes simultaneously
   - Verify consensus achieves within expected blocks

4. **Smoke Tests**
   - Submit test transactions
   - Verify all modules functional
   - Test identity verification flow
   - Test marketplace order flow

### Ceremony Checklist

- [ ] All validators have submitted valid gentxs
- [ ] Genesis hash matches across all validators
- [ ] Genesis time is coordinated (UTC)
- [ ] All validators have synced configs
- [ ] Communication channel established
- [ ] Backup genesis file stored securely
- [ ] Rollback plan documented

## Mainnet Launch Checklist

### Pre-Launch (T-7 days)

- [ ] Final genesis parameters approved
- [ ] All genesis validators confirmed
- [ ] Testnet rehearsal completed successfully
- [ ] Security audit findings addressed
- [ ] Documentation published
- [ ] SDKs released
- [ ] Block explorers configured

### Launch Day (T-0)

- [ ] Genesis file distributed to all validators
- [ ] Genesis hash verified by all parties
- [ ] Nodes configured and ready
- [ ] Monitoring dashboards live
- [ ] On-call team assembled
- [ ] Communication channels active

### Post-Launch (T+1)

- [ ] Block production verified
- [ ] Consensus healthy
- [ ] Explorer synced
- [ ] API endpoints responsive
- [ ] First transactions confirmed
- [ ] Incident retrospective scheduled

---

## Security Warnings

> ⚠️ **NEVER** share validator private keys  
> ⚠️ **NEVER** modify genesis.json after distribution without coordination  
> ⚠️ **ALWAYS** verify genesis hash before starting  
> ⚠️ **ALWAYS** use secure channels for coordination

---

## Appendix: Genesis Generation Script

```bash
#!/bin/bash
# genesis_generate.sh - Deterministic genesis generation
# Usage: ./genesis_generate.sh <chain_id> <genesis_time>

set -euo pipefail

CHAIN_ID="${1:-virtengine-1}"
GENESIS_TIME="${2:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

echo "Generating genesis for chain: ${CHAIN_ID}"
echo "Genesis time: ${GENESIS_TIME}"

# Initialize
virtengine init genesis-generator --chain-id "${CHAIN_ID}" --home /tmp/genesis

# Set genesis time
jq --arg time "${GENESIS_TIME}" '.genesis_time = $time' \
    /tmp/genesis/config/genesis.json > /tmp/genesis.tmp
mv /tmp/genesis.tmp /tmp/genesis/config/genesis.json

# Add accounts (example)
# virtengine genesis add-genesis-account ...

# Collect gentxs
# virtengine genesis collect-gentxs --home /tmp/genesis

# Validate
virtengine genesis validate-genesis --home /tmp/genesis

# Output hash
echo "Genesis hash:"
sha256sum /tmp/genesis/config/genesis.json

# Copy to output
cp /tmp/genesis/config/genesis.json ./genesis.json

echo "Genesis generation complete"
```
