# Security Best Practices Training

**Module Duration:** 8 hours  
**Target Audience:** VirtEngine Operators, System Administrators, DevOps Engineers  
**Prerequisites:** Basic understanding of Linux systems, networking fundamentals, VirtEngine architecture  
**Last Updated:** 2024

---

## Table of Contents

1. [Module Overview](#module-overview)
2. [Learning Objectives](#learning-objectives)
3. [Part 1: Operational Security Fundamentals (2 hours)](#part-1-operational-security-fundamentals)
4. [Part 2: Network Security (2 hours)](#part-2-network-security)
5. [Part 3: Key Security and Secrets Management (2 hours)](#part-3-key-security-and-secrets-management)
6. [Part 4: Logging, Auditing, and Compliance (2 hours)](#part-4-logging-auditing-and-compliance)
7. [Security Checklists](#security-checklists)
8. [Hardening Guides](#hardening-guides)
9. [Assessment Exercises](#assessment-exercises)
10. [Additional Resources](#additional-resources)

---

## Module Overview

This comprehensive security training module covers essential security practices for operating VirtEngine infrastructure. Participants will learn defense-in-depth strategies, network hardening techniques, cryptographic key management, and compliance requirements specific to blockchain-based cloud computing environments.

VirtEngine's security model is built on several critical foundations:

- **X25519-XSalsa20-Poly1305** encryption for all sensitive data
- **Three-signature validation** for VEID scopes (client, user, salt binding)
- **Hardware/ledger support** for provider keys
- **Zero-knowledge principles** - secrets never logged or stored in plaintext
- **Memory safety** - vault passwords cleared from memory after use

---

## Learning Objectives

Upon completion of this module, participants will be able to:

1. Implement defense-in-depth security strategies for VirtEngine nodes
2. Configure network security controls including firewalls and DDoS protection
3. Properly manage cryptographic keys and secrets
4. Establish secure deployment practices
5. Configure comprehensive logging and auditing
6. Understand and implement compliance requirements
7. Conduct security assessments and hardening procedures

---

## Part 1: Operational Security Fundamentals

### 1.1 Defense in Depth Strategy

Defense in depth is a security strategy that employs multiple layers of security controls. For VirtEngine operators, this means implementing security at every level of the infrastructure stack.

#### The Seven Layers of Defense

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 7: Application Security                              │
│  - Input validation, secure coding, VEID verification       │
├─────────────────────────────────────────────────────────────┤
│  Layer 6: Data Security                                     │
│  - X25519-XSalsa20-Poly1305 encryption, key management      │
├─────────────────────────────────────────────────────────────┤
│  Layer 5: Host Security                                     │
│  - OS hardening, patch management, endpoint protection      │
├─────────────────────────────────────────────────────────────┤
│  Layer 4: Network Security                                  │
│  - Firewalls, IDS/IPS, network segmentation                 │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Perimeter Security                                │
│  - DDoS protection, sentry nodes, rate limiting             │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: Physical Security                                 │
│  - Data center access controls, hardware security           │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: Policies and Procedures                           │
│  - Security policies, training, incident response           │
└─────────────────────────────────────────────────────────────┘
```

#### Implementation Checklist

- [ ] Document all security layers and controls
- [ ] Identify gaps between layers
- [ ] Implement compensating controls where gaps exist
- [ ] Regularly test each layer independently
- [ ] Conduct integrated security testing

### 1.2 Principle of Least Privilege

The principle of least privilege dictates that users, processes, and systems should only have the minimum permissions necessary to perform their functions.

#### User Access Management

```bash
# Create dedicated service accounts
useradd -r -s /bin/false virtengine-validator
useradd -r -s /bin/false virtengine-provider

# Configure sudo with minimal permissions
cat > /etc/sudoers.d/virtengine << 'EOF'
# VirtEngine operator permissions
virtengine-ops ALL=(virtengine-validator) NOPASSWD: /usr/local/bin/virtengine
virtengine-ops ALL=(virtengine-provider) NOPASSWD: /usr/local/bin/provider-daemon
EOF
```

#### File System Permissions

| Path                        | Owner                 | Permissions | Purpose             |
| --------------------------- | --------------------- | ----------- | ------------------- |
| `/etc/virtengine/`          | root:virtengine       | 750         | Configuration files |
| `/var/lib/virtengine/`      | virtengine:virtengine | 700         | Chain data          |
| `/var/lib/virtengine/keys/` | virtengine:virtengine | 700         | Key storage         |
| `/var/log/virtengine/`      | virtengine:adm        | 750         | Log files           |
| `/run/virtengine/`          | virtengine:virtengine | 755         | Runtime files       |

#### Process Isolation

```yaml
# systemd service configuration for validator
[Service]
User=virtengine-validator
Group=virtengine
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes
LockPersonality=yes
ReadWritePaths=/var/lib/virtengine /var/log/virtengine
```

### 1.3 Secure Configuration Management

#### Configuration File Security

```bash
# Secure configuration directory
chmod 750 /etc/virtengine
chown root:virtengine /etc/virtengine

# Protect sensitive configuration files
chmod 640 /etc/virtengine/config.toml
chmod 600 /etc/virtengine/priv_validator_key.json

# Validate configuration checksums
sha256sum /etc/virtengine/*.toml > /etc/virtengine/.config.sha256
```

#### Configuration Backup and Recovery

```bash
#!/bin/bash
# config-backup.sh - Secure configuration backup

BACKUP_DIR="/var/backups/virtengine"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/config-${DATE}.tar.gz.enc"

# Create encrypted backup
tar -czf - /etc/virtengine/ | \
  openssl enc -aes-256-cbc -salt -pbkdf2 -out "${BACKUP_FILE}"

# Verify backup integrity
openssl enc -d -aes-256-cbc -pbkdf2 -in "${BACKUP_FILE}" | \
  tar -tzf - > /dev/null && echo "Backup verified successfully"

# Rotate old backups (keep last 30)
find "${BACKUP_DIR}" -name "config-*.tar.gz.enc" -mtime +30 -delete
```

### 1.4 Secure Deployment Practices

#### Pre-Deployment Security Checklist

- [ ] Verify binary checksums against official releases
- [ ] Review release notes for security patches
- [ ] Test deployment in staging environment
- [ ] Backup current configuration and state
- [ ] Document rollback procedures
- [ ] Schedule maintenance window
- [ ] Notify relevant stakeholders

#### Binary Verification

```bash
#!/bin/bash
# verify-binary.sh - Verify VirtEngine binary integrity

RELEASE_URL="https://github.com/virtengine/virtengine/releases"
BINARY_PATH="/usr/local/bin/virtengine"

# Download checksums and signature
wget "${RELEASE_URL}/latest/download/SHA256SUMS"
wget "${RELEASE_URL}/latest/download/SHA256SUMS.sig"

# Verify GPG signature
gpg --verify SHA256SUMS.sig SHA256SUMS

# Verify binary checksum
sha256sum -c SHA256SUMS --ignore-missing

# Verify binary was compiled with expected settings
virtengine version --long | grep -E "(build_tags|commit)"
```

#### Secure Upgrade Procedure

```bash
#!/bin/bash
# secure-upgrade.sh - Secure VirtEngine upgrade procedure

set -euo pipefail

# 1. Create state backup
virtengine export --for-zero-height > /var/backups/virtengine/state-$(date +%s).json

# 2. Stop validator gracefully
systemctl stop virtengine-validator

# 3. Backup current binary
cp /usr/local/bin/virtengine /var/backups/virtengine/virtengine-$(virtengine version)

# 4. Install new binary (after verification)
mv virtengine-new /usr/local/bin/virtengine
chmod 755 /usr/local/bin/virtengine

# 5. Verify installation
virtengine version --long

# 6. Start validator
systemctl start virtengine-validator

# 7. Monitor for issues
journalctl -u virtengine-validator -f --since "1 minute ago"
```

---

## Part 2: Network Security

### 2.1 Firewall Configuration

#### Recommended Firewall Rules

```bash
#!/bin/bash
# firewall-setup.sh - VirtEngine firewall configuration

# Reset firewall rules
iptables -F
iptables -X
iptables -Z

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# SSH (restricted to management network)
iptables -A INPUT -p tcp --dport 22 -s 10.0.0.0/8 -j ACCEPT

# P2P networking (Tendermint)
iptables -A INPUT -p tcp --dport 26656 -j ACCEPT

# RPC (restricted to internal network)
iptables -A INPUT -p tcp --dport 26657 -s 10.0.0.0/8 -j ACCEPT

# gRPC (restricted to internal network)
iptables -A INPUT -p tcp --dport 9090 -s 10.0.0.0/8 -j ACCEPT

# REST API (restricted to internal network)
iptables -A INPUT -p tcp --dport 1317 -s 10.0.0.0/8 -j ACCEPT

# Prometheus metrics (restricted)
iptables -A INPUT -p tcp --dport 26660 -s 10.0.0.0/8 -j ACCEPT

# Provider daemon ports (if applicable)
iptables -A INPUT -p tcp --dport 8443 -j ACCEPT

# Rate limiting for P2P port
iptables -A INPUT -p tcp --dport 26656 -m connlimit --connlimit-above 50 -j DROP
iptables -A INPUT -p tcp --dport 26656 -m hashlimit \
  --hashlimit-upto 20/sec --hashlimit-burst 40 \
  --hashlimit-mode srcip --hashlimit-name p2p -j ACCEPT

# Log dropped packets
iptables -A INPUT -j LOG --log-prefix "DROPPED: " --log-level 4

# Drop everything else
iptables -A INPUT -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
```

#### Network Segmentation

```
┌─────────────────────────────────────────────────────────────────────┐
│                         PUBLIC INTERNET                              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    DMZ (DDoS Protection Layer)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │ Sentry Node │  │ Sentry Node │  │ Sentry Node │                 │
│  │   (P2P)     │  │   (P2P)     │  │   (P2P)     │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘                 │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
┌──────────────────────────────┐   ┌──────────────────────────────────┐
│      VALIDATOR NETWORK       │   │       PROVIDER NETWORK           │
│  ┌────────────────────────┐  │   │  ┌────────────────────────────┐  │
│  │    Validator Node      │  │   │  │   Provider Daemon          │  │
│  │  (private_peer_ids)    │  │   │  │   (K8s/OpenStack/VMware)   │  │
│  └────────────────────────┘  │   │  └────────────────────────────┘  │
│  ┌────────────────────────┐  │   │  ┌────────────────────────────┐  │
│  │    State Sync Node     │  │   │  │   Compute Resources        │  │
│  └────────────────────────┘  │   │  └────────────────────────────┘  │
└──────────────────────────────┘   └──────────────────────────────────┘
                    │                               │
                    └───────────────┬───────────────┘
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      MANAGEMENT NETWORK                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │
│  │  Bastion    │  │  Monitoring │  │  Backup/Key Management      │ │
│  │  Host       │  │  (Grafana)  │  │  (Vault/HSM)                │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 DDoS Protection

#### Layer 3/4 DDoS Mitigation

```bash
# Kernel parameters for DDoS mitigation
cat >> /etc/sysctl.d/99-virtengine-ddos.conf << 'EOF'
# Enable SYN cookies
net.ipv4.tcp_syncookies = 1

# Increase SYN backlog
net.ipv4.tcp_max_syn_backlog = 65535

# Decrease SYN-ACK retries
net.ipv4.tcp_synack_retries = 2

# Enable reverse path filtering
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP broadcasts
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Increase conntrack table size
net.netfilter.nf_conntrack_max = 1048576

# Reduce conntrack timeouts
net.netfilter.nf_conntrack_tcp_timeout_established = 600
net.netfilter.nf_conntrack_tcp_timeout_time_wait = 30
EOF

sysctl --system
```

#### Rate Limiting with nftables

```bash
#!/usr/sbin/nft -f
# nftables DDoS protection rules

table inet virtengine_ddos {
    set blacklist {
        type ipv4_addr
        flags timeout
        timeout 1h
    }

    set ratelimit {
        type ipv4_addr
        flags dynamic,timeout
        timeout 60s
    }

    chain input {
        type filter hook input priority -10; policy accept;

        # Drop blacklisted IPs
        ip saddr @blacklist drop

        # Rate limit new connections per source IP
        tcp dport 26656 ct state new \
            add @ratelimit { ip saddr limit rate over 20/second burst 50 packets } \
            add @blacklist { ip saddr timeout 1h } drop

        # Limit concurrent connections per IP
        tcp dport 26656 ct count over 50 drop

        # Limit packet rate
        tcp dport 26656 limit rate 10000/second burst 20000 packets accept
        tcp dport 26656 drop
    }
}
```

### 2.3 Sentry Node Architecture

Sentry nodes protect validators from direct exposure to the public network.

#### Sentry Node Configuration

```toml
# sentry-node/config.toml

[p2p]
# Listen on all interfaces for public peers
laddr = "tcp://0.0.0.0:26656"

# Connect to validator via private network
persistent_peers = "validator_node_id@10.0.1.10:26656"

# Protect validator node ID from being gossiped
private_peer_ids = "validator_node_id"

# Enable peer exchange
pex = true

# Maximum number of inbound/outbound peers
max_num_inbound_peers = 100
max_num_outbound_peers = 50

# External address for peer advertising
external_address = "sentry1.example.com:26656"
```

#### Validator Node Configuration

```toml
# validator-node/config.toml

[p2p]
# Only listen on private network
laddr = "tcp://10.0.1.10:26656"

# Connect only to sentry nodes
persistent_peers = "sentry1_id@10.0.1.1:26656,sentry2_id@10.0.1.2:26656"

# Disable peer exchange (only communicate with sentries)
pex = false

# Limit connections
max_num_inbound_peers = 5
max_num_outbound_peers = 5

# No external address
external_address = ""
```

### 2.4 TLS/mTLS Configuration

#### Generating TLS Certificates

```bash
#!/bin/bash
# generate-certs.sh - Generate TLS certificates for VirtEngine

CERT_DIR="/etc/virtengine/certs"
DAYS_VALID=365

mkdir -p "${CERT_DIR}"
chmod 700 "${CERT_DIR}"

# Generate CA key and certificate
openssl genrsa -out "${CERT_DIR}/ca.key" 4096
openssl req -new -x509 -days 3650 -key "${CERT_DIR}/ca.key" \
  -out "${CERT_DIR}/ca.crt" \
  -subj "/C=US/ST=CA/O=VirtEngine/CN=VirtEngine CA"

# Generate server certificate
openssl genrsa -out "${CERT_DIR}/server.key" 4096
openssl req -new -key "${CERT_DIR}/server.key" \
  -out "${CERT_DIR}/server.csr" \
  -subj "/C=US/ST=CA/O=VirtEngine/CN=validator.example.com"

# Create extensions file
cat > "${CERT_DIR}/server.ext" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = validator.example.com
DNS.2 = localhost
IP.1 = 10.0.1.10
IP.2 = 127.0.0.1
EOF

# Sign server certificate
openssl x509 -req -in "${CERT_DIR}/server.csr" \
  -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" \
  -CAcreateserial -out "${CERT_DIR}/server.crt" \
  -days ${DAYS_VALID} -sha256 -extfile "${CERT_DIR}/server.ext"

# Set permissions
chmod 600 "${CERT_DIR}"/*.key
chmod 644 "${CERT_DIR}"/*.crt
chown -R virtengine:virtengine "${CERT_DIR}"
```

---

## Part 3: Key Security and Secrets Management

### 3.1 VirtEngine Cryptographic Architecture

VirtEngine uses a layered cryptographic approach:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Cryptographic Architecture                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  VEID Identity Layer                                         │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │   │
│  │  │ Client Signature │  │ User Signature  │  │ Salt Binding│  │   │
│  │  │ (Capture App)    │  │ (Wallet)        │  │ (Replay     │  │   │
│  │  │                  │  │                 │  │  Prevention)│  │   │
│  │  └─────────────────┘  └─────────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Encryption Layer: X25519-XSalsa20-Poly1305                  │   │
│  │  ┌─────────────────────────────────────────────────────────┐│   │
│  │  │ EncryptionEnvelope {                                    ││   │
│  │  │   RecipientFingerprint: string  // Validator's key ID   ││   │
│  │  │   Algorithm: "X25519-XSalsa20-Poly1305"                 ││   │
│  │  │   Ciphertext: []byte                                    ││   │
│  │  │   Nonce: []byte                                         ││   │
│  │  │ }                                                       ││   │
│  │  └─────────────────────────────────────────────────────────┘│   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Key Storage Layer                                           │   │
│  │  ┌───────────────┐  ┌───────────────┐  ┌─────────────────┐  │   │
│  │  │ Hardware/HSM  │  │ Ledger Device │  │ Encrypted File  │  │   │
│  │  │ (Recommended) │  │ (Supported)   │  │ (Last Resort)   │  │   │
│  │  └───────────────┘  └───────────────┘  └─────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Key Types and Protection Requirements

| Key Type              | Storage        | Protection Level | Rotation Frequency       |
| --------------------- | -------------- | ---------------- | ------------------------ |
| Validator Private Key | HSM/Ledger     | Critical         | Never (or with ceremony) |
| Node Key              | Encrypted File | High             | Annually                 |
| Provider Key          | HSM/Ledger     | Critical         | Never (or with ceremony) |
| VEID Encryption Key   | HSM            | Critical         | Per-session              |
| TLS Private Key       | Encrypted File | Medium           | Annually                 |
| API Keys              | Vault          | Medium           | Quarterly                |

### 3.3 Hardware Security Module (HSM) Integration

#### YubiHSM2 Configuration

```bash
#!/bin/bash
# setup-yubihsm.sh - Configure YubiHSM2 for VirtEngine

# Install YubiHSM connector
apt-get install -y yubihsm-connector

# Configure connector
cat > /etc/yubihsm-connector.yaml << 'EOF'
listen: 127.0.0.1:12345
timeout: 5m
log:
  level: info
  path: /var/log/yubihsm-connector.log
EOF

# Start connector service
systemctl enable yubihsm-connector
systemctl start yubihsm-connector

# Generate authentication key (replace default)
yubihsm-shell -a generate-asymmetric \
  -A rsa2048 \
  -c sign-pkcs \
  -l "VirtEngine Validator Key"
```

#### Ledger Integration

```go
// Example: Using Ledger for validator signing
package main

import (
    "github.com/cosmos/cosmos-sdk/crypto/ledger"
    "github.com/cosmos/cosmos-sdk/crypto/hd"
)

func initLedgerSigner() error {
    // Connect to Ledger device
    device, err := ledger.NewPrivKeyLedgerSecp256k1(
        hd.CreateHDPath(118, 0, 0), // Cosmos HD path
    )
    if err != nil {
        return fmt.Errorf("failed to connect to Ledger: %w", err)
    }

    // Verify device connection
    pubKey, err := device.PubKey()
    if err != nil {
        return fmt.Errorf("failed to get public key: %w", err)
    }

    log.Printf("Connected to Ledger. Validator address: %s",
        sdk.AccAddress(pubKey.Address()))

    return nil
}
```

### 3.4 Secrets Management with HashiCorp Vault

#### Vault Configuration for VirtEngine

```hcl
# vault-config.hcl - VirtEngine secrets management

# Enable KV secrets engine for VirtEngine
path "secret/data/virtengine/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}

# Enable transit engine for encryption
path "transit/encrypt/virtengine-*" {
  capabilities = ["update"]
}

path "transit/decrypt/virtengine-*" {
  capabilities = ["update"]
}

# Policy for validator nodes
path "secret/data/virtengine/validator/*" {
  capabilities = ["read"]
}

# Policy for provider daemons
path "secret/data/virtengine/provider/*" {
  capabilities = ["read"]
}
```

#### Vault Agent Configuration

```hcl
# vault-agent.hcl - Auto-auth and secret retrieval

auto_auth {
  method "kubernetes" {
    mount_path = "auth/kubernetes"
    config = {
      role = "virtengine-validator"
    }
  }

  sink "file" {
    config = {
      path = "/var/run/vault/.token"
      mode = 0600
    }
  }
}

template {
  source      = "/etc/vault.d/templates/config.toml.ctmpl"
  destination = "/etc/virtengine/config.toml"
  perms       = 0640
  command     = "systemctl reload virtengine-validator"
}
```

### 3.5 Memory Security for Secrets

VirtEngine enforces strict memory security for sensitive data:

```go
// Example: Secure memory handling for vault passwords
package security

import (
    "crypto/subtle"
    "runtime"
    "unsafe"
)

// SecureBytes is a wrapper that ensures bytes are cleared from memory
type SecureBytes struct {
    data []byte
}

// NewSecureBytes creates a new secure byte slice
func NewSecureBytes(size int) *SecureBytes {
    return &SecureBytes{data: make([]byte, size)}
}

// Clear wipes the data from memory
func (s *SecureBytes) Clear() {
    if s.data != nil {
        // Use subtle.ConstantTimeCopy to avoid optimization
        zeros := make([]byte, len(s.data))
        subtle.ConstantTimeCopy(1, s.data, zeros)

        // Prevent garbage collector from optimizing away
        runtime.KeepAlive(s.data)
        s.data = nil
    }
}

// Use finalizer to ensure cleanup
func init() {
    runtime.SetFinalizer(&SecureBytes{}, func(s *SecureBytes) {
        s.Clear()
    })
}
```

### 3.6 Key Ceremony Procedures

#### Validator Key Generation Ceremony

```markdown
## Validator Key Generation Ceremony

### Participants Required

- Key Ceremony Lead (1)
- Key Custodians (3, for split key recovery)
- Witness (1, independent security officer)
- Scribe (1, documentation)

### Pre-Ceremony Checklist

- [ ] Air-gapped computer prepared and verified
- [ ] Tamper-evident bags for key shards
- [ ] Video recording equipment ready
- [ ] Hardware wallets/HSMs initialized and verified
- [ ] All participants' identities verified

### Ceremony Steps

1. **Environment Verification** (15 minutes)
   - Verify air-gapped computer has no network connectivity
   - Boot from verified live USB
   - Verify no recording devices except official camera

2. **Key Generation** (30 minutes)
   - Generate entropy using hardware RNG + dice rolls
   - Execute key generation script
   - Verify key can produce valid signatures
   - Record public key hash

3. **Key Distribution** (30 minutes)
   - Split key using Shamir's Secret Sharing (3-of-5)
   - Place each shard in tamper-evident bag
   - Record bag serial numbers
   - Distribute to key custodians

4. **Verification** (15 minutes)
   - Verify key can be reconstructed from shards
   - Verify signature capability
   - Securely destroy reconstruction

5. **Documentation** (15 minutes)
   - Record all serial numbers
   - All participants sign ceremony log
   - Seal and store documentation
```

---

## Part 4: Logging, Auditing, and Compliance

### 4.1 Comprehensive Logging Configuration

#### VirtEngine Logging Configuration

```toml
# config.toml - Logging configuration

[log]
# Log level: debug, info, warn, error
level = "info"

# Log format: plain, json
format = "json"

# Output destination
output = "file"
output_file = "/var/log/virtengine/virtengine.log"

# Enable module-specific logging
[log.modules]
consensus = "info"
p2p = "warn"
mempool = "warn"
state = "info"
veid = "info"
market = "info"
escrow = "info"
provider = "info"
```

#### Security Event Logging

```yaml
# fluent-bit.conf - Security event collection

[SERVICE]
    Flush         5
    Daemon        Off
    Log_Level     info
    Parsers_File  parsers.conf

[INPUT]
    Name          tail
    Path          /var/log/virtengine/*.log
    Parser        json
    Tag           virtengine.*
    Refresh_Interval 5

[FILTER]
    Name          grep
    Match         virtengine.*
    Regex         level (error|warn)

[FILTER]
    Name          grep
    Match         virtengine.*
    Regex         module (auth|veid|encryption)

[OUTPUT]
    Name          forward
    Match         *
    Host          security-siem.internal
    Port          24224
    tls           on
    tls.verify    on
    tls.ca_file   /etc/ssl/certs/ca.pem
```

### 4.2 Audit Trail Requirements

#### Events Requiring Audit Logging

| Event Category  | Events                                                | Retention |
| --------------- | ----------------------------------------------------- | --------- |
| Authentication  | Login success/failure, key usage, session creation    | 2 years   |
| Authorization   | Permission checks, role changes, access denials       | 2 years   |
| VEID Operations | Scope uploads, verifications, encryption              | 7 years   |
| Financial       | Escrow creation, fund release, market transactions    | 7 years   |
| Configuration   | Parameter changes, governance proposals               | 7 years   |
| Security        | Key generation, certificate issuance, security alerts | 7 years   |

#### Audit Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00.000Z",
  "event_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "event_type": "VEID_SCOPE_UPLOAD",
  "severity": "INFO",
  "actor": {
    "type": "user",
    "id": "virtengine1abc...",
    "ip_address": "10.0.1.100"
  },
  "resource": {
    "type": "veid_scope",
    "id": "scope-12345"
  },
  "action": "CREATE",
  "outcome": "SUCCESS",
  "details": {
    "scope_type": "facial_verification",
    "encryption_algorithm": "X25519-XSalsa20-Poly1305",
    "signatures_verified": ["client", "user", "salt"]
  },
  "metadata": {
    "node_id": "validator-1",
    "chain_id": "virtengine-1",
    "block_height": 12345678
  }
}
```

### 4.3 Compliance Requirements

#### GDPR Compliance Checklist

- [ ] Data Processing Agreement (DPA) in place with all processors
- [ ] Privacy Impact Assessment completed for VEID
- [ ] Consent mechanisms implemented for biometric data
- [ ] Data subject access request procedures documented
- [ ] Right to erasure procedures implemented (where applicable)
- [ ] Data breach notification procedures in place
- [ ] Cross-border transfer safeguards implemented

#### SOC 2 Type II Controls

| Trust Service Criteria    | VirtEngine Controls                           |
| ------------------------- | --------------------------------------------- |
| CC1.1 - COSO Principle 1  | Security policies documented and communicated |
| CC2.1 - COSO Principle 13 | Change management process for deployments     |
| CC5.2 - COSO Principle 12 | System monitoring and alerting                |
| CC6.1 - Logical Access    | Role-based access with MFA                    |
| CC6.7 - Data Transmission | TLS 1.3 for all communications                |
| CC7.1 - System Operations | Automated deployment pipelines                |
| CC7.2 - Change Management | Version control and code review               |

### 4.4 Security Monitoring and Alerting

#### Critical Security Alerts

```yaml
# prometheus-alerts.yml - Security alerting rules

groups:
  - name: virtengine-security
    rules:
      - alert: ValidatorDoubleSign
        expr: tendermint_consensus_byzantine_validators > 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Byzantine behavior detected"
          description: "Validator may have double-signed"

      - alert: UnauthorizedAccessAttempt
        expr: rate(virtengine_auth_failures_total[5m]) > 10
        for: 2m
        labels:
          severity: high
        annotations:
          summary: "High rate of authentication failures"

      - alert: EncryptionKeyUsageAnomaly
        expr: rate(virtengine_encryption_operations_total[5m]) > 1000
        for: 5m
        labels:
          severity: medium
        annotations:
          summary: "Unusual encryption key usage pattern"

      - alert: P2PConnectionAnomaly
        expr: tendermint_p2p_peers < 5 or tendermint_p2p_peers > 100
        for: 10m
        labels:
          severity: medium
        annotations:
          summary: "Abnormal peer count detected"

      - alert: ConsensusHaltDetected
        expr: increase(tendermint_consensus_height[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Blockchain consensus has halted"
```

---

## Security Checklists

### Pre-Production Security Checklist

```markdown
## Pre-Production Security Review

### Infrastructure Security

- [ ] All servers hardened according to CIS benchmarks
- [ ] Firewalls configured and tested
- [ ] DDoS protection in place
- [ ] Sentry node architecture implemented
- [ ] Network segmentation verified
- [ ] TLS/mTLS configured for all services

### Key Management

- [ ] Validator keys stored in HSM/Ledger
- [ ] Key ceremony completed and documented
- [ ] Key backup procedures tested
- [ ] Key recovery procedures documented and tested
- [ ] No keys stored in plaintext anywhere

### Access Control

- [ ] All access follows least privilege principle
- [ ] MFA enabled for all operator accounts
- [ ] Service accounts use minimal permissions
- [ ] Access review process established
- [ ] Privileged access logging enabled

### Monitoring and Logging

- [ ] Centralized logging configured
- [ ] Security alerts configured and tested
- [ ] Audit logging enabled for all security events
- [ ] Log retention policies implemented
- [ ] SIEM integration complete

### Incident Response

- [ ] Incident response plan documented
- [ ] Contact lists current and accessible
- [ ] Runbooks for common scenarios created
- [ ] Tabletop exercises completed
- [ ] Post-incident review process defined

### Compliance

- [ ] All required compliance controls implemented
- [ ] Documentation complete and current
- [ ] Third-party security audit scheduled
- [ ] Penetration test completed
- [ ] Vulnerability assessment completed
```

### Daily Security Operations Checklist

```markdown
## Daily Security Operations

### Morning Review (Start of Day)

- [ ] Review overnight security alerts
- [ ] Check validator uptime and status
- [ ] Verify backup completion
- [ ] Review authentication logs for anomalies
- [ ] Check system resource utilization
- [ ] Verify all services running normally

### Continuous Monitoring

- [ ] Monitor Grafana dashboards
- [ ] Respond to alerts within SLA
- [ ] Document any security incidents
- [ ] Track remediation of open issues

### End of Day

- [ ] Review day's security events
- [ ] Update incident tracker
- [ ] Handoff to next shift (if applicable)
- [ ] Verify overnight monitoring active
```

### Monthly Security Review Checklist

```markdown
## Monthly Security Review

### Access Review

- [ ] Review all user accounts
- [ ] Disable/remove inactive accounts
- [ ] Verify permissions are appropriate
- [ ] Review API key usage

### Vulnerability Management

- [ ] Run vulnerability scans
- [ ] Review and prioritize findings
- [ ] Patch critical vulnerabilities
- [ ] Update dependency versions

### Configuration Review

- [ ] Review firewall rules
- [ ] Verify logging configuration
- [ ] Check backup integrity
- [ ] Review security monitoring rules

### Documentation

- [ ] Update runbooks as needed
- [ ] Review and update incident response plan
- [ ] Update contact lists
- [ ] Document any security changes

### Compliance

- [ ] Review compliance status
- [ ] Address any compliance gaps
- [ ] Update compliance documentation
```

---

## Hardening Guides

### Operating System Hardening

```bash
#!/bin/bash
# os-hardening.sh - Linux system hardening for VirtEngine nodes

set -euo pipefail

echo "=== VirtEngine Node Hardening Script ==="

# 1. Update system
apt-get update && apt-get upgrade -y

# 2. Configure automatic security updates
apt-get install -y unattended-upgrades
cat > /etc/apt/apt.conf.d/50unattended-upgrades << 'EOF'
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF

# 3. Disable unnecessary services
systemctl disable --now cups bluetooth avahi-daemon

# 4. Configure SSH hardening
cat > /etc/ssh/sshd_config.d/hardening.conf << 'EOF'
Protocol 2
PermitRootLogin no
PasswordAuthentication no
ChallengeResponseAuthentication no
UsePAM yes
X11Forwarding no
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server
MaxAuthTries 3
MaxSessions 5
ClientAliveInterval 300
ClientAliveCountMax 2
LoginGraceTime 60
AllowGroups ssh-users virtengine-ops
EOF

systemctl restart sshd

# 5. Configure system auditing
apt-get install -y auditd audispd-plugins
cat >> /etc/audit/rules.d/virtengine.rules << 'EOF'
# Monitor VirtEngine configuration changes
-w /etc/virtengine/ -p wa -k virtengine_config

# Monitor key operations
-w /var/lib/virtengine/keys/ -p rwa -k virtengine_keys

# Monitor authentication events
-w /var/log/auth.log -p wa -k auth_log

# Monitor sudoers
-w /etc/sudoers -p wa -k sudoers
-w /etc/sudoers.d/ -p wa -k sudoers

# Monitor user/group modifications
-w /etc/passwd -p wa -k identity
-w /etc/shadow -p wa -k identity
-w /etc/group -p wa -k identity
-w /etc/gshadow -p wa -k identity

# Capture all failed access attempts
-a always,exit -F arch=b64 -S open -F success=0 -k failed_access

# Monitor network configuration
-w /etc/hosts -p wa -k hosts
-w /etc/network/ -p wa -k network
EOF

systemctl restart auditd

# 6. Configure kernel security parameters
cat > /etc/sysctl.d/99-virtengine-security.conf << 'EOF'
# Kernel hardening
kernel.dmesg_restrict = 1
kernel.kptr_restrict = 2
kernel.perf_event_paranoid = 3
kernel.yama.ptrace_scope = 2
kernel.sysrq = 0

# Disable core dumps
fs.suid_dumpable = 0

# Address space layout randomization
kernel.randomize_va_space = 2

# Network hardening
net.ipv4.tcp_timestamps = 0
net.ipv4.icmp_ignore_bogus_error_responses = 1
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0
net.ipv6.conf.default.accept_source_route = 0
EOF

sysctl --system

# 7. Install and configure fail2ban
apt-get install -y fail2ban
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
ignoreip = 127.0.0.1/8 ::1 10.0.0.0/8

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
EOF

systemctl enable --now fail2ban

# 8. Configure file integrity monitoring
apt-get install -y aide
aideinit
mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db

echo "=== Hardening complete ==="
```

### Container Security Hardening

```yaml
# pod-security-policy.yaml - Kubernetes security for VirtEngine workloads

apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: virtengine-restricted
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfiles: "runtime/default"
    seccomp.security.alpha.kubernetes.io/defaultProfileName: "runtime/default"
spec:
  privileged: false
  allowPrivilegeEscalation: false
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: MustRunAsNonRoot
  runAsGroup:
    rule: MustRunAs
    ranges:
      - min: 1000
        max: 65534
  fsGroup:
    rule: MustRunAs
    ranges:
      - min: 1000
        max: 65534
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: MustRunAs
    ranges:
      - min: 1000
        max: 65534
  volumes:
    - "configMap"
    - "emptyDir"
    - "projected"
    - "secret"
    - "downwardAPI"
    - "persistentVolumeClaim"
  readOnlyRootFilesystem: true
  requiredDropCapabilities:
    - ALL
```

---

## Assessment Exercises

### Exercise 1: Security Configuration Review (1 hour)

**Objective:** Identify security misconfigurations in a VirtEngine node setup.

**Scenario:** You are provided with the following configuration files. Identify at least 5 security issues and propose fixes.

```toml
# EXERCISE: Review this config.toml for security issues

[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]
unsafe = true

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""
seeds = ""
persistent_peers = ""
pex = true
seed_mode = false

[mempool]
size = 5000
max_txs_bytes = 1073741824

[statesync]
enable = true
rpc_servers = "http://public-rpc1.example.com:26657,http://public-rpc2.example.com:26657"
trust_height = 0
trust_hash = ""

[logging]
level = "debug"
format = "plain"
```

**Expected Security Issues:**

1. RPC bound to all interfaces (0.0.0.0)
2. CORS allows all origins (\*)
3. Unsafe RPC mode enabled
4. No persistent peers configured (vulnerable to eclipse attacks)
5. Debug logging level (potential information disclosure)
6. State sync using HTTP instead of HTTPS
7. No trust_hash configured for state sync

---

### Exercise 2: Incident Response Simulation (1 hour)

**Objective:** Practice responding to a security incident.

**Scenario:** At 14:30 UTC, your monitoring system alerts you to the following:

- 500% increase in failed authentication attempts
- Unusual outbound network traffic from validator node
- New process `systemd-networkd-helper` consuming 80% CPU

**Tasks:**

1. Document your initial assessment
2. List your first 5 response actions
3. Identify evidence to preserve
4. Determine escalation criteria
5. Draft initial communication to stakeholders

---

### Exercise 3: Key Ceremony Tabletop (30 minutes)

**Objective:** Walk through a key generation ceremony scenario.

**Scenario:** Your organization is setting up a new validator. Walk through the key ceremony process and identify:

1. Required participants and their roles
2. Equipment needed
3. Security controls at each step
4. Potential failure points
5. Recovery procedures if something goes wrong

---

### Exercise 4: Penetration Test Preparation (30 minutes)

**Objective:** Prepare for a third-party security assessment.

**Tasks:**

1. List all entry points to your VirtEngine infrastructure
2. Document expected attack vectors
3. Identify your most critical assets
4. Prepare scoping document for penetration testers
5. Define rules of engagement

---

## Additional Resources

### External References

- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks)
- [OWASP Security Guidelines](https://owasp.org/www-project-web-security-testing-guide/)
- [Cosmos SDK Security Best Practices](https://docs.cosmos.network/main/learn/advanced/security)

### VirtEngine-Specific Documentation

- [VirtEngine Security Architecture](./../security/architecture.md)
- [VEID Encryption Specification](./../security/veid-encryption.md)
- [Provider Security Guide](./../security/provider-security.md)
- [Incident Response Playbook](./../security/incident-response.md)

### Training Completion

Upon completing this module, operators should:

1. Take the security assessment exam (passing score: 85%)
2. Complete at least 2 hands-on exercises
3. Participate in one tabletop exercise
4. Document their security implementation plan

---

**Module Version:** 1.0  
**Last Review:** 2024-01  
**Next Review:** 2024-07  
**Owner:** VirtEngine Security Team
