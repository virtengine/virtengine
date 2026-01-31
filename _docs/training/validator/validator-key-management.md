# Validator Key Management

> **Duration:** 4 hours (1 day)  
> **Prerequisites:** Basic cryptography knowledge, Linux administration  
> **Security Classification:** CONFIDENTIAL - Handle with care

---

## Table of Contents

1. [Module Overview](#module-overview)
2. [Key Types and Purposes](#key-types-and-purposes)
3. [Key Generation and Storage](#key-generation-and-storage)
4. [HSM Integration](#hsm-integration)
5. [Key Rotation Procedures](#key-rotation-procedures)
6. [Backup and Recovery](#backup-and-recovery)
7. [Security Best Practices](#security-best-practices)
8. [Key Compromise Response](#key-compromise-response)
9. [Emergency Runbooks](#emergency-runbooks)
10. [Assessment](#assessment)

---

## Module Overview

### Why Key Management Matters

Validator keys are the most critical security assets in blockchain operations:

| Key Compromise Impact | Consequence |
|----------------------|-------------|
| **Validator Signing Key** | Double-signing → 5% stake slashed, permanent jail |
| **Operator Account Key** | Fund theft, unauthorized staking operations |
| **VEID Identity Key** | Decrypt all identity data assigned to validator |
| **Node P2P Key** | Network impersonation, eclipse attacks |

### Learning Objectives

By the end of this module, operators will be able to:

- [ ] Identify all key types used in VirtEngine validator operations
- [ ] Generate keys securely using proper entropy sources
- [ ] Configure HSM integration for production validators
- [ ] Execute key rotation procedures without downtime
- [ ] Implement comprehensive backup and recovery procedures
- [ ] Respond effectively to key compromise incidents

---

## Key Types and Purposes

### Overview of Validator Keys

```
┌─────────────────────────────────────────────────────────────────────┐
│                    VirtEngine Validator Key Hierarchy               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    VALIDATOR SIGNING KEY                      │   │
│  │  Algorithm: Ed25519                                          │   │
│  │  Purpose: Sign consensus votes (prevote, precommit, propose) │   │
│  │  Location: priv_validator_key.json or HSM                    │   │
│  │  Risk Level: CRITICAL (double-sign = tombstone)              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│         ┌────────────────────┼────────────────────┐                │
│         │                    │                    │                │
│         v                    v                    v                │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────────┐      │
│  │  OPERATOR   │     │   NODE P2P   │     │ VEID IDENTITY   │      │
│  │  ACCOUNT    │     │     KEY      │     │      KEY        │      │
│  │   KEY       │     │              │     │                 │      │
│  ├─────────────┤     ├─────────────┤     ├─────────────────┤      │
│  │ Algo: secp  │     │ Algo: Ed25519│     │ Algo: X25519    │      │
│  │ 256k1       │     │              │     │                 │      │
│  │             │     │              │     │                 │      │
│  │ Purpose:    │     │ Purpose:     │     │ Purpose:        │      │
│  │ - Tx signing│     │ - P2P ident  │     │ - Decrypt VEID  │      │
│  │ - Governance│     │ - Peer auth  │     │   envelopes     │      │
│  │ - Staking   │     │              │     │                 │      │
│  │             │     │              │     │                 │      │
│  │ Risk: HIGH  │     │ Risk: MEDIUM │     │ Risk: HIGH      │      │
│  └─────────────┘     └─────────────┘     └─────────────────┘      │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Details Table

| Key Type | Algorithm | Location | Purpose | Rotation | Backup Priority |
|----------|-----------|----------|---------|----------|-----------------|
| Validator Signing | Ed25519 | `priv_validator_key.json` | Consensus signing | Never (complex) | CRITICAL |
| Operator Account | secp256k1 | Keyring | Transactions, governance | Annually | CRITICAL |
| Node P2P | Ed25519 | `node_key.json` | Peer identification | As needed | HIGH |
| VEID Identity | X25519 | `keyring-file/` | Decrypt identity data | Quarterly | CRITICAL |

---

## Key Generation and Storage

### Session 1: Secure Key Generation (1 hour)

#### 1.1 Entropy Requirements

**Minimum entropy sources for production keys:**

```bash
# Check available entropy
cat /proc/sys/kernel/random/entropy_avail
# Should be > 256 for key generation

# Install hardware random number generator tools (if available)
sudo apt-get install rng-tools

# Enable CPU-based entropy (Intel/AMD)
sudo modprobe tpm_rng  # TPM
sudo modprobe rdrand   # Intel RDRAND

# Verify entropy source
rngtest -c 1000 < /dev/random
# Should show very low FIPS failures
```

#### 1.2 Validator Signing Key Generation

```bash
# Method 1: Standard generation (suitable for testnet)
.cache/bin/virtengine init my-validator --chain-id=virtengine-1

# Generated files:
# ~/.virtengine/config/priv_validator_key.json
# ~/.virtengine/config/node_key.json

# View validator public key
cat ~/.virtengine/config/priv_validator_key.json | jq -r '.pub_key'

# Method 2: Generate with specific key type (Ed25519)
.cache/bin/virtengine init my-validator --chain-id=virtengine-1 \
    --key-algo=ed25519

# Method 3: Import existing key (migration scenario)
.cache/bin/virtengine keys import validator-key backup.json \
    --keyring-backend=file
```

**priv_validator_key.json structure:**

```json
{
  "address": "ABCD1234567890ABCD1234567890ABCD12345678",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "base64_encoded_public_key=="
  },
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "base64_encoded_private_key=="
  }
}
```

#### 1.3 Operator Account Key Generation

```bash
# Generate new operator key with strong encryption
.cache/bin/virtengine keys add operator \
    --keyring-backend=file \
    --algo=secp256k1

# IMPORTANT: Record the mnemonic securely!
# Output:
# - name: operator
#   address: virt1abc123...
#   pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey",...}'
#   mnemonic: "word1 word2 word3 ... word24"

# Verify key was created
.cache/bin/virtengine keys list --keyring-backend=file

# Export public key (safe to share)
.cache/bin/virtengine keys show operator --pubkey --keyring-backend=file
```

#### 1.4 VEID Identity Key Generation

```bash
# Generate X25519 key for VEID operations
.cache/bin/virtengine keys add veid-identity \
    --keyring-backend=file \
    --algo=x25519

# View key fingerprint
.cache/bin/virtengine keys show veid-identity --keyring-backend=file | \
    grep fingerprint

# Register key on chain
.cache/bin/virtengine tx veid register-identity-key \
    --pubkey=$(.cache/bin/virtengine keys show veid-identity --pubkey) \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto \
    --gas-prices=0.025uvirt
```

#### 1.5 Node P2P Key Generation

```bash
# Node key is generated during init
cat ~/.virtengine/config/node_key.json | jq '.'

# Structure:
# {
#   "priv_key": {
#     "type": "tendermint/PrivKeyEd25519",
#     "value": "base64_encoded_key"
#   }
# }

# Get node ID from key
.cache/bin/virtengine tendermint show-node-id
# Output: abc123def456...
```

---

## HSM Integration

### Session 2: Hardware Security Module Setup (1 hour)

#### 2.1 HSM Options for Validators

| HSM Type | Cost | Security Level | Use Case |
|----------|------|----------------|----------|
| **YubiHSM 2** | $650 | FIPS 140-2 L3 | Small validators |
| **Thales Luna** | $10k+ | FIPS 140-2 L3 | Enterprise |
| **AWS CloudHSM** | $1.60/hr | FIPS 140-2 L3 | Cloud validators |
| **SoftHSM** | Free | Software-only | Development/testing |
| **Ledger Nano** | $150 | Secure Element | Operator keys only |

#### 2.2 YubiHSM 2 Integration

```bash
# Step 1: Install YubiHSM SDK
wget https://developers.yubico.com/YubiHSM2/Releases/yubihsm2-sdk-2023.04-ubuntu2204-amd64.tar.gz
tar xzf yubihsm2-sdk-*.tar.gz
sudo dpkg -i yubihsm2-sdk/*.deb

# Step 2: Configure connector
sudo tee /etc/yubihsm-connector.yaml << EOF
listen: 127.0.0.1:12345
device: ""
serial: ""
EOF

sudo systemctl enable yubihsm-connector
sudo systemctl start yubihsm-connector

# Step 3: Initialize HSM
yubihsm-shell
> connect
> session open 1 password
# Change default password immediately!
> put authkey 0 0 admin 1 generate-asymmetric-key:sign-eddsa all:all \
    password123
> session close 0
> session open 0x0002 password123
> generate asymmetric 0x0002 0 validator-key 1 sign-eddsa ed25519
> exit

# Step 4: Configure TMKMS for YubiHSM
# See TMKMS section below
```

#### 2.3 TMKMS (Tendermint Key Management System)

TMKMS provides secure validator signing via HSM:

```bash
# Install TMKMS
cargo install tmkms --features=yubihsm

# Initialize configuration
tmkms init /opt/tmkms

# Configure for YubiHSM
cat > /opt/tmkms/tmkms.toml << EOF
[[chain]]
id = "virtengine-1"
key_format = { type = "cosmos-json", account_key_prefix = "virtpub", consensus_key_prefix = "virtvalconspub" }
state_file = "/opt/tmkms/state/virtengine-1-consensus.json"

[[validator]]
chain_id = "virtengine-1"
addr = "unix:///opt/tmkms/sockets/validator.sock"
secret_key = "/opt/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"

[[providers.yubihsm]]
adapter = { type = "usb" }
auth = { key = 1, password = "password" }
keys = [
    { chain_ids = ["virtengine-1"], key = 1, type = "consensus" }
]
serial_number = "0123456789"
EOF

# Generate KMS identity key
tmkms keygen /opt/tmkms/secrets/kms-identity.key

# Start TMKMS
tmkms start -c /opt/tmkms/tmkms.toml
```

**Configure validator to use TMKMS:**

```toml
# ~/.virtengine/config/config.toml

[priv_validator]
# Use socket connection to TMKMS instead of local key file
laddr = "unix:///opt/tmkms/sockets/validator.sock"
```

#### 2.4 SoftHSM Setup (Development/Testing)

```bash
# Install SoftHSM
sudo apt-get install softhsm2

# Initialize token
softhsm2-util --init-token --slot 0 \
    --label "validator-keys" \
    --pin 12345678 \
    --so-pin 87654321

# Generate key in HSM
pkcs11-tool --module /usr/lib/softhsm/libsofthsm2.so \
    --login --pin 12345678 \
    --keypairgen \
    --key-type EC:ed25519 \
    --id 01 \
    --label "validator-consensus"

# List keys
pkcs11-tool --module /usr/lib/softhsm/libsofthsm2.so \
    --login --pin 12345678 \
    --list-objects

# Configure TMKMS for SoftHSM
cat > /opt/tmkms/tmkms.toml << EOF
[[chain]]
id = "virtengine-1"
key_format = { type = "cosmos-json", account_key_prefix = "virtpub", consensus_key_prefix = "virtvalconspub" }
state_file = "/opt/tmkms/state/virtengine-1-consensus.json"

[[providers.softsign]]
chain_ids = ["virtengine-1"]
key_type = "consensus"
path = "/opt/tmkms/secrets/consensus.key"
EOF
```

#### 2.5 Ledger Device for Operator Keys

```bash
# Install Ledger app support
sudo apt-get install libudev-dev libusb-1.0-0-dev

# Add udev rules for Ledger
sudo tee /etc/udev/rules.d/20-hw1.rules << EOF
SUBSYSTEMS=="usb", ATTRS{idVendor}=="2c97", MODE="0660", GROUP="plugdev"
EOF
sudo udevadm control --reload-rules

# Add operator key from Ledger
.cache/bin/virtengine keys add operator-ledger \
    --ledger \
    --keyring-backend=file

# Sign transactions with Ledger
.cache/bin/virtengine tx staking delegate ... \
    --from=operator-ledger \
    --ledger
```

---

## Key Rotation Procedures

### Session 3: Key Rotation (1 hour)

#### 3.1 Rotation Decision Matrix

| Key Type | Rotation Frequency | Rotation Difficulty | Downtime |
|----------|-------------------|---------------------|----------|
| Validator Signing | Never (avoid) | Very High | Risk of double-sign |
| Operator Account | Annually | Medium | None |
| VEID Identity | Quarterly | Low | None |
| Node P2P | As needed | Low | Brief reconnection |

#### 3.2 Operator Account Key Rotation

**Step-by-Step Procedure:**

```bash
# Phase 1: Preparation (Day -7)

# Generate new operator key
.cache/bin/virtengine keys add operator-new \
    --keyring-backend=file \
    --algo=secp256k1

# Record mnemonic securely
# Store in vault, safety deposit box, etc.

# Fund new account with small amount for testing
.cache/bin/virtengine tx bank send operator operator-new 1000000uvirt \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto

# Phase 2: Authority Transfer (Day 0)

# Transfer validator operator address
.cache/bin/virtengine tx staking edit-validator \
    --operator-address=$(virtengine keys show operator-new -a) \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto

# Transfer remaining funds
BALANCE=$(.cache/bin/virtengine query bank balances $(virtengine keys show operator -a) -o json | jq -r '.balances[0].amount')
.cache/bin/virtengine tx bank send operator operator-new ${BALANCE}uvirt \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto

# Phase 3: Verification (Day +1)

# Verify new key is active
.cache/bin/virtengine query staking validator \
    $(virtengine keys show operator-new --bech val -a)

# Test transaction with new key
.cache/bin/virtengine tx bank send operator-new test-account 1uvirt \
    --from=operator-new \
    --chain-id=virtengine-1 \
    --gas=auto

# Phase 4: Cleanup (Day +7)

# Archive old key
.cache/bin/virtengine keys export operator --keyring-backend=file > /secure/backup/operator-old-$(date +%Y%m%d).key

# Delete old key after confirmation period
.cache/bin/virtengine keys delete operator --keyring-backend=file --yes

# Rename new key
.cache/bin/virtengine keys rename operator-new operator --keyring-backend=file
```

#### 3.3 VEID Identity Key Rotation

```bash
# Phase 1: Generate new key
.cache/bin/virtengine keys add veid-identity-new \
    --keyring-backend=file \
    --algo=x25519

# Phase 2: Register new key (both keys active during transition)
.cache/bin/virtengine tx veid add-identity-key \
    --pubkey=$(virtengine keys show veid-identity-new --pubkey) \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto

# Phase 3: Wait for transition period (24 hours recommended)
# New verifications will use new key
# Old key still valid for in-flight requests

# Phase 4: Verify new key is receiving requests
.cache/bin/virtengine query veid pending-verifications \
    --validator=$(virtengine keys show operator --bech val -a) \
    --output json | jq '.verifications | group_by(.key_fingerprint) | map({key: .[0].key_fingerprint, count: length})'

# Phase 5: Revoke old key
OLD_FINGERPRINT=$(.cache/bin/virtengine keys show veid-identity --fingerprint)
.cache/bin/virtengine tx veid remove-identity-key \
    --key-fingerprint=$OLD_FINGERPRINT \
    --from=operator \
    --chain-id=virtengine-1

# Phase 6: Archive and delete old key
.cache/bin/virtengine keys export veid-identity --keyring-backend=file > /secure/backup/veid-identity-old-$(date +%Y%m%d).key
shred -u ~/.virtengine/keyring-file/veid-identity.info

# Phase 7: Rename new key
.cache/bin/virtengine keys rename veid-identity-new veid-identity --keyring-backend=file
```

#### 3.4 Node P2P Key Rotation

```bash
# This is straightforward but causes temporary peer disconnection

# Phase 1: Stop node
sudo systemctl stop virtengine

# Phase 2: Backup old key
cp ~/.virtengine/config/node_key.json /secure/backup/node_key-$(date +%Y%m%d).json

# Phase 3: Generate new key
rm ~/.virtengine/config/node_key.json
.cache/bin/virtengine tendermint show-node-id 2>/dev/null || \
    .cache/bin/virtengine init --chain-id=virtengine-1 --recover 2>/dev/null

# Alternative: Generate directly
.cache/bin/virtengine tendermint gen-node-key > ~/.virtengine/config/node_key.json

# Phase 4: Update persistent peer configurations
NEW_NODE_ID=$(.cache/bin/virtengine tendermint show-node-id)
echo "New Node ID: $NEW_NODE_ID"
# Notify other validators to update their peer lists

# Phase 5: Restart node
sudo systemctl start virtengine

# Phase 6: Verify peer connections
.cache/bin/virtengine status | jq '.NodeInfo.id'
.cache/bin/virtengine net_info | jq '.n_peers'
```

---

## Backup and Recovery

### Session 4: Backup Procedures (30 minutes)

#### 4.1 Backup Strategy Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Key Backup Strategy                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│  │   PRIMARY    │   │   SECONDARY  │   │   TERTIARY   │        │
│  │   BACKUP     │   │    BACKUP    │   │   BACKUP     │        │
│  ├──────────────┤   ├──────────────┤   ├──────────────┤        │
│  │ Location:    │   │ Location:    │   │ Location:    │        │
│  │ Encrypted    │   │ Hardware     │   │ Paper/Metal  │        │
│  │ Cloud Vault  │   │ Security Box │   │ Safe Deposit │        │
│  │              │   │              │   │              │        │
│  │ Access:      │   │ Access:      │   │ Access:      │        │
│  │ 2 of 3       │   │ Physical +   │   │ Physical +   │        │
│  │ Operators    │   │ Key Card     │   │ 2 Witnesses  │        │
│  │              │   │              │   │              │        │
│  │ Recovery:    │   │ Recovery:    │   │ Recovery:    │        │
│  │ < 1 hour     │   │ < 4 hours    │   │ < 24 hours   │        │
│  └──────────────┘   └──────────────┘   └──────────────┘        │
│                                                                  │
│  Contents at each location:                                      │
│  ✓ Validator signing key (encrypted)                            │
│  ✓ Operator mnemonic (24 words)                                 │
│  ✓ VEID identity key (encrypted)                                │
│  ✓ Recovery passwords (sealed envelope)                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2 Encrypted Backup Script

```bash
#!/bin/bash
# /opt/scripts/backup_validator_keys.sh

set -euo pipefail

BACKUP_DIR="/secure/backups/keys"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/validator_keys_${DATE}.tar.gz.gpg"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Copy keys to temp directory
cp ~/.virtengine/config/priv_validator_key.json "$TEMP_DIR/"
cp ~/.virtengine/config/node_key.json "$TEMP_DIR/"
cp -r ~/.virtengine/keyring-file/ "$TEMP_DIR/keyring/"

# Create checksums
sha256sum "$TEMP_DIR"/* > "$TEMP_DIR/checksums.sha256"

# Create encrypted archive
tar -czf - -C "$TEMP_DIR" . | \
    gpg --symmetric \
        --cipher-algo AES256 \
        --output "$BACKUP_FILE"

# Verify backup can be decrypted
gpg --decrypt "$BACKUP_FILE" | tar -tzf - > /dev/null 2>&1

echo "Backup created: $BACKUP_FILE"
echo "Checksum: $(sha256sum $BACKUP_FILE)"

# Upload to remote storage (optional)
# rclone copy "$BACKUP_FILE" remote:validator-backups/

# Cleanup old backups (keep last 30 days)
find "$BACKUP_DIR" -name "*.gpg" -mtime +30 -delete
```

#### 4.3 Mnemonic Backup Procedure

```markdown
## Mnemonic Backup Checklist

### Materials Needed
- [ ] Metal backup plate (e.g., Cryptosteel, Billfodl)
- [ ] Fireproof safe or safety deposit box
- [ ] Tamper-evident envelope
- [ ] Witness (for multi-signature setups)

### Procedure

1. [ ] Generate mnemonic in secure, air-gapped environment
2. [ ] Write mnemonic on paper (temporary)
3. [ ] Stamp mnemonic on metal backup plate
4. [ ] Verify metal backup by reading back all words
5. [ ] Place metal backup in tamper-evident envelope
6. [ ] Store in primary secure location
7. [ ] Create secondary backup using same procedure
8. [ ] Store secondary backup in geographically separate location
9. [ ] Securely destroy paper copy (shred + burn)
10. [ ] Document storage locations (without revealing mnemonic)
11. [ ] Test recovery procedure quarterly
```

#### 4.4 Recovery Procedures

**Scenario 1: Server Failure (Keys Intact)**

```bash
# New server setup

# Step 1: Provision new server
# (Follow standard server hardening procedures)

# Step 2: Install VirtEngine
git clone https://github.com/virtengine/virtengine.git
cd virtengine && make virtengine

# Step 3: Restore configuration
tar -xzf /backup/config_backup.tar.gz -C ~/.virtengine/

# Step 4: Restore keys from encrypted backup
gpg --decrypt /backup/validator_keys_latest.tar.gz.gpg | \
    tar -xzf - -C ~/.virtengine/

# Step 5: Verify key checksums
sha256sum -c ~/.virtengine/checksums.sha256

# Step 6: Start node with state sync
# (Configure state sync in config.toml)
sudo systemctl start virtengine
```

**Scenario 2: Key Loss (Mnemonic Available)**

```bash
# Step 1: Recover operator key from mnemonic
.cache/bin/virtengine keys add operator \
    --keyring-backend=file \
    --recover
# Enter mnemonic when prompted

# Step 2: Verify recovered address matches
.cache/bin/virtengine keys show operator -a --keyring-backend=file

# Step 3: For validator signing key, contact security team
# Validator key recovery from mnemonic requires special handling
# to prevent double-signing

# Step 4: Generate new VEID identity key
.cache/bin/virtengine keys add veid-identity \
    --keyring-backend=file \
    --algo=x25519

# Step 5: Register new identity key
.cache/bin/virtengine tx veid register-identity-key \
    --pubkey=$(virtengine keys show veid-identity --pubkey) \
    --from=operator
```

---

## Security Best Practices

### Checklist: Key Security

```markdown
## Daily Security Checklist

- [ ] Verify key file permissions (600)
- [ ] Check for unauthorized access attempts in logs
- [ ] Verify HSM connectivity (if applicable)
- [ ] Monitor signing activity for anomalies

## Weekly Security Checklist

- [ ] Review access logs for key storage systems
- [ ] Verify backup integrity (decrypt test)
- [ ] Check for pending security updates
- [ ] Review and rotate access credentials

## Monthly Security Checklist

- [ ] Audit key access permissions
- [ ] Test disaster recovery procedures
- [ ] Review key rotation schedule
- [ ] Update emergency contacts

## Quarterly Security Checklist

- [ ] Rotate VEID identity keys
- [ ] Full backup verification and restore test
- [ ] Security training refresher for operators
- [ ] Review and update key management documentation
```

### File Permissions

```bash
# Set correct permissions for all key files
chmod 600 ~/.virtengine/config/priv_validator_key.json
chmod 600 ~/.virtengine/config/node_key.json
chmod 700 ~/.virtengine/keyring-file/
chmod 600 ~/.virtengine/keyring-file/*

# Verify permissions
ls -la ~/.virtengine/config/*.json
ls -la ~/.virtengine/keyring-file/

# Expected output:
# -rw------- 1 validator validator  priv_validator_key.json
# -rw------- 1 validator validator  node_key.json
```

### Network Security

```bash
# Never expose key files via network

# Firewall rules to block key file access
sudo iptables -A OUTPUT -p tcp -m string \
    --string "priv_validator_key" --algo bm -j DROP

# Disable core dumps (could contain key material)
ulimit -c 0
echo "* hard core 0" >> /etc/security/limits.conf

# Secure swap (keys could be swapped to disk)
sudo swapoff -a  # Disable swap completely
# OR encrypt swap
# cryptsetup -d /dev/urandom create cryptoswap /dev/sdX
```

### Access Control

```bash
# Create dedicated validator user
sudo useradd -r -s /sbin/nologin validator

# Limit sudo access
echo "validator ALL=(ALL) NOPASSWD: /bin/systemctl restart virtengine" | \
    sudo tee /etc/sudoers.d/validator

# Enable audit logging
sudo auditctl -w ~/.virtengine/config/priv_validator_key.json -p rwxa

# Review audit logs
sudo ausearch -f priv_validator_key.json
```

---

## Key Compromise Response

### Incident Response Flowchart

```
┌─────────────────────────────────────────────────────────────────┐
│                KEY COMPROMISE RESPONSE FLOW                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Suspected Compromise Detected                                  │
│            │                                                     │
│            v                                                     │
│   ┌─────────────────┐                                           │
│   │ IMMEDIATE STOP  │ ──> Stop validator node                   │
│   │   (< 5 min)     │ ──> Revoke compromised key                │
│   └────────┬────────┘ ──> Alert security team                   │
│            │                                                     │
│            v                                                     │
│   ┌─────────────────┐                                           │
│   │    ASSESS       │ ──> Identify compromised keys             │
│   │   (< 30 min)    │ ──> Determine scope of exposure           │
│   └────────┬────────┘ ──> Preserve evidence                     │
│            │                                                     │
│            v                                                     │
│   ┌─────────────────┐                                           │
│   │   REMEDIATE     │ ──> Generate new keys                     │
│   │   (< 4 hours)   │ ──> Restore from clean backup             │
│   └────────┬────────┘ ──> Re-register with network              │
│            │                                                     │
│            v                                                     │
│   ┌─────────────────┐                                           │
│   │    RECOVER      │ ──> Verify new key operation              │
│   │   (< 24 hours)  │ ──> Resume validation                     │
│   └────────┬────────┘ ──> Monitor for further issues            │
│            │                                                     │
│            v                                                     │
│   ┌─────────────────┐                                           │
│   │  POST-INCIDENT  │ ──> Root cause analysis                   │
│   │   (< 1 week)    │ ──> Update procedures                     │
│   └─────────────────┘ ──> Stakeholder communication             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Response by Key Type

#### Validator Signing Key Compromise

```bash
# CRITICAL: Act immediately to prevent double-signing

# Step 1: STOP THE VALIDATOR IMMEDIATELY
sudo systemctl stop virtengine

# Step 2: Isolate the server
sudo iptables -A INPUT -j DROP
sudo iptables -A OUTPUT -j DROP
# Keep SSH access only for incident response

# Step 3: Preserve evidence
tar -czf /tmp/evidence_$(date +%Y%m%d_%H%M%S).tar.gz \
    ~/.virtengine/config/ \
    /var/log/virtengine/ \
    ~/.bash_history

# Step 4: Alert security team
# security@virtengine.io
# Include: timestamp, nature of compromise, actions taken

# Step 5: DO NOT RESTART until cleared by security team
# Double-signing risk is severe (5% slash, permanent jail)

# Step 6: Recovery requires coordination with network
# May need emergency governance proposal if validator tombstoned
```

#### Operator Account Key Compromise

```bash
# Step 1: Transfer assets immediately
# Use backup key or contact support for emergency measures

# Step 2: Revoke operator permissions if possible
.cache/bin/virtengine tx authz revoke \
    --grantee=COMPROMISED_ADDRESS \
    --from=backup-operator \
    --chain-id=virtengine-1

# Step 3: Generate new operator key
.cache/bin/virtengine keys add operator-emergency \
    --keyring-backend=file

# Step 4: Update validator operator address
.cache/bin/virtengine tx staking edit-validator \
    --operator-address=$(virtengine keys show operator-emergency -a) \
    --from=backup-operator \
    --chain-id=virtengine-1

# Step 5: Monitor for unauthorized transactions
.cache/bin/virtengine query txs \
    --events="message.sender=COMPROMISED_ADDRESS" \
    --limit=100
```

#### VEID Identity Key Compromise

```bash
# Step 1: Stop VEID processing
sudo systemctl stop virtengine

# Step 2: Revoke compromised key immediately
.cache/bin/virtengine tx veid emergency-revoke-key \
    --key-fingerprint=COMPROMISED_FINGERPRINT \
    --from=operator \
    --chain-id=virtengine-1

# Step 3: Generate new identity key
.cache/bin/virtengine keys add veid-identity-emergency \
    --keyring-backend=file \
    --algo=x25519

# Step 4: Register new key
.cache/bin/virtengine tx veid register-identity-key \
    --pubkey=$(virtengine keys show veid-identity-emergency --pubkey) \
    --from=operator \
    --chain-id=virtengine-1

# Step 5: Audit all decryptions with compromised key
.cache/bin/virtengine query veid key-usage-log \
    --key-fingerprint=COMPROMISED_FINGERPRINT \
    --output json > /evidence/veid_usage_audit.json

# Step 6: Notify affected users (if identity data was exposed)
# Follow data breach notification procedures
```

---

## Emergency Runbooks

### Runbook: Lost Validator Signing Key

```markdown
## Scenario: Validator signing key is lost or corrupted

### Severity: CRITICAL
### Time to Resolution: 4-24 hours

### Prerequisites
- Mnemonic backup available OR
- Encrypted key backup available OR  
- HSM with recoverable key

### Procedure

1. **STOP** - Do not attempt to regenerate key without coordination
   - Risk: Double-signing if old key still exists elsewhere
   
2. **Assess**
   - Is the old key definitely destroyed?
   - Do any backups exist?
   - Was the key in HSM?

3. **If backup exists:**
   ```bash
   # Restore from encrypted backup
   gpg --decrypt backup.gpg | tar -xzf - -C ~/.virtengine/
   
   # Verify key matches registered validator
   .cache/bin/virtengine tendermint show-validator
   # Compare with on-chain registration
   ```

4. **If mnemonic recovery needed:**
   ```bash
   # Contact VirtEngine security team FIRST
   # They will coordinate safe recovery
   # security@virtengine.io
   ```

5. **If key is unrecoverable:**
   - Validator must be re-created with new key
   - Old validator will be eventually jailed for downtime
   - Consider: Was validator tombstoned for double-signing?

### Post-Incident
- [ ] Update backup procedures
- [ ] Document root cause
- [ ] Test restored key operation
```

### Runbook: HSM Failure

```markdown
## Scenario: Hardware Security Module is unresponsive

### Severity: HIGH
### Time to Resolution: 1-4 hours

### Immediate Actions

1. **Check HSM status**
   ```bash
   # YubiHSM status
   yubihsm-connector status
   pkcs11-tool --module /usr/lib/yubihsm.so --show-info
   
   # Check USB connection
   lsusb | grep -i yubico
   dmesg | tail -20
   ```

2. **If HSM unresponsive:**
   ```bash
   # Restart connector service
   sudo systemctl restart yubihsm-connector
   
   # Wait 30 seconds
   sleep 30
   
   # Test connection
   yubihsm-shell -C connect
   ```

3. **If still failing:**
   ```bash
   # Stop validator to prevent missed blocks
   sudo systemctl stop virtengine
   
   # Try physical reset of HSM
   # (Power cycle USB, if applicable)
   ```

4. **If HSM hardware failure:**
   - Switch to backup HSM if available
   - OR restore from software backup (temporary)
   - Contact HSM vendor for replacement

### Failover to Backup

```bash
# Switch to software key (emergency only)
# This reduces security but maintains uptime

# Restore software key from backup
gpg --decrypt /backup/priv_validator_key.json.gpg > \
    ~/.virtengine/config/priv_validator_key.json
chmod 600 ~/.virtengine/config/priv_validator_key.json

# Modify config to use local key instead of HSM
sed -i 's|laddr = "unix:.*"|laddr = ""|' ~/.virtengine/config/config.toml

# Restart validator
sudo systemctl start virtengine

# IMPORTANT: Switch back to HSM as soon as possible
```
```

### Runbook: Suspected Unauthorized Access

```markdown
## Scenario: Evidence of unauthorized access to key systems

### Severity: CRITICAL
### Time to Resolution: 24-72 hours

### Indicators of Compromise
- Unexpected transactions from operator account
- Unfamiliar SSH sessions in logs
- Modified key files or timestamps
- Alerts from intrusion detection

### Immediate Actions

1. **Isolate the system** (within 5 minutes)
   ```bash
   # Block network except SSH from known IP
   sudo iptables -A INPUT -p tcp --dport 22 -s YOUR_IP -j ACCEPT
   sudo iptables -A INPUT -j DROP
   sudo iptables -A OUTPUT -j DROP
   ```

2. **Stop validator** (within 5 minutes)
   ```bash
   sudo systemctl stop virtengine
   ```

3. **Preserve evidence** (within 15 minutes)
   ```bash
   # Memory dump (if forensics needed)
   sudo dd if=/dev/mem of=/evidence/memory_dump.bin bs=1M
   
   # Copy logs
   cp -r /var/log/ /evidence/logs/
   cp ~/.bash_history /evidence/
   
   # Get process list
   ps auxf > /evidence/processes.txt
   
   # Network connections
   netstat -tulpan > /evidence/network.txt
   ss -tulpan >> /evidence/network.txt
   
   # Check for rootkits
   rkhunter --check --skip-keypress > /evidence/rootkit_check.txt
   ```

4. **Revoke potentially compromised keys** (within 30 minutes)
   - Follow key-specific compromise procedures above

5. **Notify stakeholders**
   - Security team: security@virtengine.io
   - Delegators (if funds at risk)
   - Other validators (if network-wide threat)

### Investigation Checklist
- [ ] Review SSH access logs
- [ ] Check for unauthorized cron jobs
- [ ] Verify binary integrity
- [ ] Review firewall logs
- [ ] Check for data exfiltration
- [ ] Analyze memory dump
- [ ] Review all recent transactions
```

---

## Assessment

### Knowledge Check

1. What algorithm is used for validator signing keys in VirtEngine?
2. Why should validator signing keys never be rotated routinely?
3. What are the three backup locations recommended for critical keys?
4. What is the maximum time window for responding to a key compromise?
5. How does TMKMS prevent double-signing?

### Practical Exercises

Complete the following exercises in a test environment:

- [ ] Generate and securely store a complete set of validator keys
- [ ] Configure TMKMS with SoftHSM
- [ ] Perform a VEID identity key rotation
- [ ] Execute a backup and restore test
- [ ] Respond to a simulated key compromise scenario

### Certification Requirements

| Requirement | Minimum |
|-------------|---------|
| Knowledge Check | 90% correct |
| Practical Exercises | All completed |
| Incident Response Time | < 30 minutes |
| Backup Verification | Successful restore |

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│              VALIDATOR KEY MANAGEMENT QUICK REFERENCE            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  KEY FILES                                                       │
│  ~/.virtengine/config/priv_validator_key.json  [CRITICAL]       │
│  ~/.virtengine/config/node_key.json            [HIGH]           │
│  ~/.virtengine/keyring-file/                   [HIGH]           │
│                                                                  │
│  PERMISSIONS                                                     │
│  chmod 600 priv_validator_key.json                              │
│  chmod 600 node_key.json                                        │
│  chmod 700 keyring-file/                                        │
│                                                                  │
│  EMERGENCY COMMANDS                                              │
│  Stop validator:     sudo systemctl stop virtengine             │
│  Check key:          virtengine tendermint show-validator       │
│  Revoke VEID key:    virtengine tx veid emergency-revoke-key    │
│                                                                  │
│  CONTACTS                                                        │
│  Security:           security@virtengine.io                     │
│  Operations:         ops@virtengine.io                          │
│  Discord:            discord.gg/virtengine #validators          │
│                                                                  │
│  BACKUP CHECKLIST                                                │
│  [ ] Validator key encrypted and stored offsite                 │
│  [ ] Operator mnemonic on metal backup                          │
│  [ ] VEID key encrypted and stored offsite                      │
│  [ ] Recovery tested within last 90 days                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

*Document Version: 1.0.0*  
*Last Updated: 2024-01-15*  
*Maintainer: VirtEngine Security Team*  
*Classification: CONFIDENTIAL*