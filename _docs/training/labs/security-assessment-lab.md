# Security Assessment Lab

**Lab Duration:** 4 hours  
**Prerequisites:** Security Fundamentals, Lab Environment Setup  
**Track:** Security

---

## Overview

This lab provides hands-on experience with security assessment procedures for VirtEngine infrastructure. You will perform security audits, verify configurations, and practice incident response.

---

## Prerequisites

Before starting this lab, ensure you have:
- [ ] Completed [Lab Environment Setup](lab-environment.md)
- [ ] Local VirtEngine network running
- [ ] Access to monitoring dashboards
- [ ] Security checklist template

---

## Exercise 1: Security Configuration Audit

**Duration:** 45 minutes  
**Objective:** Perform a comprehensive security audit of a validator node

### Setup

Start the local environment:
```bash
./scripts/localnet.sh start
```

### Task 1.1: File Permission Audit

Check file permissions on critical files:

```bash
# Check key file permissions
ls -la ~/.virtengine/config/

# Expected: 600 for sensitive files
# priv_validator_key.json should be 600
# node_key.json should be 600

# Check actual permissions
stat -c "%a %n" ~/.virtengine/config/priv_validator_key.json
stat -c "%a %n" ~/.virtengine/config/node_key.json
```

**Expected Output:**
```
600 /home/user/.virtengine/config/priv_validator_key.json
600 /home/user/.virtengine/config/node_key.json
```

**Document findings:**
- [ ] priv_validator_key.json has 600 permissions
- [ ] node_key.json has 600 permissions
- [ ] config.toml has 644 permissions (readable)
- [ ] Directory has 700 permissions

### Task 1.2: Configuration Security Review

Review `config.toml` for security issues:

```bash
# Check RPC binding
grep -E "^laddr.*26657" ~/.virtengine/config/config.toml

# Expected: tcp://127.0.0.1:26657 (localhost only)
# NOT: tcp://0.0.0.0:26657 (exposed to internet!)

# Check P2P configuration
grep -E "^pex|^persistent_peers" ~/.virtengine/config/config.toml

# For validators behind sentries:
# pex = false
# persistent_peers = "sentry_nodes_only"
```

**Security Checklist:**
- [ ] RPC bound to localhost only
- [ ] API bound to localhost only
- [ ] gRPC bound to localhost only
- [ ] P2P correctly configured
- [ ] Prometheus metrics restricted

### Task 1.3: Network Exposure Scan

Identify exposed ports:

```bash
# List listening ports
netstat -tlnp | grep virtengine

# Or using ss
ss -tlnp | grep virtengine

# Check from external perspective (if on different machine)
# nmap -sT -p 26656,26657,1317,9090,26660 <target_ip>
```

**Expected for validator:**
| Port | Binding | Status |
|------|---------|--------|
| 26656 | 0.0.0.0 | ✓ Required for P2P |
| 26657 | 127.0.0.1 | ✓ RPC localhost only |
| 1317 | 127.0.0.1 | ✓ API localhost only |
| 9090 | 127.0.0.1 | ✓ gRPC localhost only |
| 26660 | 127.0.0.1 or specific IP | ✓ Metrics restricted |

### Deliverable

Complete the Security Audit Report:

```markdown
## Security Audit Report

**Date**: YYYY-MM-DD
**Auditor**: [Your Name]
**Node**: [Node Identifier]

### File Permissions
| File | Expected | Actual | Status |
|------|----------|--------|--------|
| priv_validator_key.json | 600 | | |
| node_key.json | 600 | | |
| config.toml | 644 | | |

### Network Configuration
| Port | Expected Binding | Actual | Status |
|------|-----------------|--------|--------|
| 26656 | 0.0.0.0 | | |
| 26657 | 127.0.0.1 | | |

### Findings
1. [Finding description]

### Recommendations
1. [Recommendation]
```

---

## Exercise 2: Key Management Verification

**Duration:** 45 minutes  
**Objective:** Verify key management practices and backup procedures

### Task 2.1: Key Inventory

List all keys and their purposes:

```bash
# List all keys in keyring
virtengine keys list --keyring-backend test

# For each key, verify:
# - Purpose (validator, operator, identity)
# - Key type (secp256k1, ed25519)
# - Address
```

**Document:**

| Key Name | Purpose | Type | Address | Notes |
|----------|---------|------|---------|-------|
| validator | Block signing | ed25519 | ... | |
| operator | Tx signing | secp256k1 | ... | |

### Task 2.2: Backup Verification

Verify backup procedures work:

```bash
# Export a test key
virtengine keys export test-key --keyring-backend test > test-key.armor

# Verify export contents (encrypted)
cat test-key.armor
# Should start with: -----BEGIN TENDERMINT PRIVATE KEY-----

# Test import to verify backup works
virtengine keys delete test-key --keyring-backend test -y
virtengine keys import test-key test-key.armor --keyring-backend test

# Verify restored
virtengine keys show test-key --keyring-backend test

# Clean up
rm test-key.armor
```

### Task 2.3: Key Rotation Simulation

Practice the key rotation procedure:

```bash
# Generate new key
virtengine keys add new-operator-key --keyring-backend test

# In production, you would:
# 1. Register new key on-chain
# 2. Wait for confirmation
# 3. Update references to use new key
# 4. After grace period, revoke old key

# Verify both keys exist
virtengine keys list --keyring-backend test
```

### Deliverable

Complete Key Management Verification:

```markdown
## Key Management Verification

**Date**: YYYY-MM-DD
**Verifier**: [Your Name]

### Key Inventory
- [ ] All keys documented
- [ ] Purposes assigned
- [ ] No unused keys present

### Backup Verification
- [ ] Export procedure works
- [ ] Import procedure works
- [ ] Backup location documented
- [ ] Backup encrypted

### Rotation Readiness
- [ ] Rotation procedure documented
- [ ] Practice rotation completed
- [ ] Recovery procedure tested
```

---

## Exercise 3: Network Security Testing

**Duration:** 45 minutes  
**Objective:** Verify network security controls

### Task 3.1: Firewall Rule Verification

Test firewall rules are enforced:

```bash
# Test RPC access from localhost (should work)
curl -s http://127.0.0.1:26657/status | jq .result.node_info.moniker

# Test RPC access from external (should fail in production)
# From another machine:
# curl -s http://<validator_ip>:26657/status
# Should timeout or be refused

# Test P2P port (should be open)
# nc -zv <validator_ip> 26656
```

### Task 3.2: TLS/Encryption Verification

Verify encrypted communications:

```bash
# Check if gRPC uses TLS (in production)
# openssl s_client -connect localhost:9090 </dev/null 2>/dev/null | openssl x509 -noout -subject

# Check peer connections use encryption
virtengine query ibc client connections
```

### Task 3.3: DDoS Resilience Check

Verify rate limiting and connection limits:

```bash
# Check connection limits in config
grep -E "max_num_inbound_peers|max_num_outbound_peers" ~/.virtengine/config/config.toml

# Expected:
# max_num_inbound_peers = 100
# max_num_outbound_peers = 50

# Check current connections
curl -s localhost:26657/net_info | jq '.result.n_peers'
```

### Deliverable

Complete Network Security Checklist:

```markdown
## Network Security Verification

**Date**: YYYY-MM-DD

### Firewall Rules
- [ ] RPC restricted to localhost
- [ ] API restricted to localhost
- [ ] P2P port accessible
- [ ] Prometheus metrics restricted

### Encryption
- [ ] P2P connections encrypted
- [ ] API uses TLS (if exposed)

### DDoS Protection
- [ ] Connection limits configured
- [ ] Rate limiting enabled
- [ ] Sentry architecture (production)
```

---

## Exercise 4: Log Analysis for Security Events

**Duration:** 45 minutes  
**Objective:** Identify security-relevant events in logs

### Task 4.1: Authentication Events

Search for authentication-related events:

```bash
# Search for failed authentication attempts
journalctl -u virtengine | grep -iE "auth|failed|denied|rejected" | tail -50

# Look for patterns:
# - Repeated failures from same source
# - Unusual times of activity
# - Unknown accounts
```

### Task 4.2: Consensus Anomalies

Check for consensus-related security events:

```bash
# Look for double-signing warnings
journalctl -u virtengine | grep -iE "double|sign|slash" | tail -20

# Look for peer misbehavior
journalctl -u virtengine | grep -iE "bad|malicious|banned" | tail -20
```

### Task 4.3: Resource Anomalies

Check for resource-based attacks:

```bash
# Memory usage spikes
journalctl -u virtengine | grep -iE "out of memory|oom|memory" | tail -20

# Disk space issues
journalctl -u virtengine | grep -iE "disk|space|full" | tail -20

# CPU issues
journalctl -u virtengine | grep -iE "timeout|slow|deadline" | tail -20
```

### Task 4.4: Create Security Event Summary

Review logs and summarize:

```markdown
## Security Log Analysis

**Time Period**: [Start] to [End]
**Analyst**: [Your Name]

### Authentication Events
| Time | Event | Source | Action Taken |
|------|-------|--------|--------------|
| | | | |

### Consensus Events
| Time | Event | Severity | Action Taken |
|------|-------|----------|--------------|
| | | | |

### Anomalies Detected
1. [Description]

### Recommendations
1. [Recommendation]
```

---

## Exercise 5: Incident Response Drill

**Duration:** 60 minutes  
**Objective:** Practice responding to a security incident

### Scenario: Key Compromise Suspected

**Setup**: You receive an alert that unauthorized transactions may have been signed with your validator key.

### Task 5.1: Immediate Response (15 minutes)

Follow the incident response procedure:

```bash
# Step 1: Verify the alert
# Check recent signing activity
virtengine query slashing signing-infos

# Step 2: Check for unauthorized transactions
virtengine query txs --events 'message.sender=<your_validator_address>' --limit 10

# Step 3: If compromise confirmed, stop signing immediately
# DO NOT ACTUALLY DO THIS IN PRODUCTION WITHOUT CONFIRMATION
# sudo systemctl stop virtengine

# Step 4: Document timeline
# Note all actions with timestamps
```

### Task 5.2: Investigation (20 minutes)

Investigate the potential compromise:

```bash
# Check access logs
sudo journalctl -u virtengine --since "24 hours ago" | grep -iE "key|sign|access"

# Check SSH access logs
sudo journalctl -u sshd --since "24 hours ago"

# Check file modification times
stat ~/.virtengine/config/priv_validator_key.json

# Look for unauthorized processes
ps aux | grep -v grep | grep -E "virtengine|keyring"
```

### Task 5.3: Containment (15 minutes)

If compromise is confirmed:

```bash
# 1. Rotate keys (in test environment)
# Generate new validator key
virtengine init temp-recovery --recover

# 2. Update firewall to block suspicious sources
# ufw deny from <suspicious_ip>

# 3. Notify stakeholders
# Send notification to #security channel
```

### Task 5.4: Documentation (10 minutes)

Create incident report:

```markdown
## Security Incident Report

**Incident ID**: SEC-YYYY-MM-DD-001
**Severity**: [Critical/High/Medium/Low]
**Status**: [Investigating/Contained/Resolved]

### Timeline
| Time | Action |
|------|--------|
| | Alert received |
| | Investigation started |
| | [Compromise confirmed/false positive] |
| | Containment actions taken |

### Impact
- [Description of impact]

### Root Cause
- [If determined]

### Actions Taken
1. [Action 1]
2. [Action 2]

### Follow-up Required
1. [ ] [Action item]
```

---

## Lab Completion Checklist

- [ ] Exercise 1: Security Configuration Audit completed
- [ ] Exercise 2: Key Management Verification completed
- [ ] Exercise 3: Network Security Testing completed
- [ ] Exercise 4: Log Analysis completed
- [ ] Exercise 5: Incident Response Drill completed
- [ ] All deliverables documented

## Cleanup

```bash
# Stop local environment
./scripts/localnet.sh stop

# Clean up test artifacts
rm -f *.armor *.backup

# Review and securely delete any sensitive test data
```

---

## Key Takeaways

1. **Regular audits** catch configuration drift
2. **Key management** requires documented procedures
3. **Network security** must be verified, not assumed
4. **Log analysis** reveals security events
5. **Practice incidents** prepare you for real ones

---

**Document Owner**: Training Team  
**Last Updated**: 2026-01-31  
**Version**: 1.0.0
