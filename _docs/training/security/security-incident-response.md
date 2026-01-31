# Security Incident Response Training

**Module Duration:** 4 hours  
**Target Audience:** VirtEngine Operators, Security Engineers, On-Call Personnel  
**Prerequisites:** Completion of Security Best Practices module  
**Last Updated:** 2024

---

## Table of Contents

1. [Module Overview](#module-overview)
2. [Learning Objectives](#learning-objectives)
3. [Part 1: Incident Classification and Triage (1 hour)](#part-1-incident-classification-and-triage)
4. [Part 2: Response Procedures (1 hour)](#part-2-response-procedures)
5. [Part 3: Evidence and Communication (1 hour)](#part-3-evidence-and-communication)
6. [Part 4: Recovery and Post-Incident (1 hour)](#part-4-recovery-and-post-incident)
7. [Tabletop Exercises](#tabletop-exercises)
8. [Incident Scenarios](#incident-scenarios)
9. [Recovery Procedures](#recovery-procedures)
10. [Quick Reference Cards](#quick-reference-cards)

---

## Module Overview

Security incident response is a critical capability for VirtEngine operators. This module provides comprehensive training on detecting, responding to, and recovering from security incidents affecting blockchain infrastructure, provider systems, and the VEID identity platform.

VirtEngine's security architecture requires specialized incident response procedures:
- **Validator incidents** affecting consensus and chain integrity
- **VEID incidents** involving biometric data and encryption
- **Provider incidents** affecting tenant workloads and infrastructure
- **Economic incidents** affecting escrow and market operations

---

## Learning Objectives

Upon completion of this module, participants will be able to:

1. Classify and prioritize security incidents using standardized criteria
2. Execute incident response procedures for common VirtEngine scenarios
3. Handle key compromise incidents including evidence preservation
4. Detect and respond to malware in provider infrastructure
5. Manage data breach incidents including regulatory requirements
6. Communicate effectively during incidents
7. Conduct post-incident reviews and implement improvements

---

## Part 1: Incident Classification and Triage

### 1.1 Security Incident Definition

A security incident is any event that:
- Compromises the confidentiality, integrity, or availability of VirtEngine systems
- Violates security policies or procedures
- Indicates unauthorized access or malicious activity
- Threatens the security of user data or funds

### 1.2 Incident Classification Framework

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Incident Severity Levels                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  CRITICAL (SEV-1)                                            │   │
│  │  Response Time: IMMEDIATE (< 15 minutes)                     │   │
│  │  ─────────────────────────────────────────────────────────── │   │
│  │  • Active exploitation of validator/provider systems         │   │
│  │  • Confirmed key compromise (validator or provider)          │   │
│  │  • Consensus failure or chain halt                           │   │
│  │  • Large-scale VEID data breach                              │   │
│  │  • Active double-signing detected                            │   │
│  │  • Ransomware or destructive malware                         │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  HIGH (SEV-2)                                                │   │
│  │  Response Time: < 1 hour                                     │   │
│  │  ─────────────────────────────────────────────────────────── │   │
│  │  • Suspected key compromise under investigation              │   │
│  │  • Unauthorized access to production systems                 │   │
│  │  • Significant DDoS attack affecting operations              │   │
│  │  • Provider container escape detected                        │   │
│  │  • Targeted phishing against operators                       │   │
│  │  • Suspicious network activity from validators               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  MEDIUM (SEV-3)                                              │   │
│  │  Response Time: < 4 hours                                    │   │
│  │  ─────────────────────────────────────────────────────────── │   │
│  │  • Failed authentication attempts exceeding threshold        │   │
│  │  • Vulnerability discovered in production system             │   │
│  │  • Suspicious but unconfirmed activity                       │   │
│  │  • Minor data exposure (non-sensitive)                       │   │
│  │  • Security policy violation                                 │   │
│  │  • Failed intrusion attempt                                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  LOW (SEV-4)                                                 │   │
│  │  Response Time: < 24 hours                                   │   │
│  │  ─────────────────────────────────────────────────────────── │   │
│  │  • Reconnaissance activity detected                          │   │
│  │  • Non-critical system compromise                            │   │
│  │  • Security tool alert requiring investigation               │   │
│  │  • Policy deviation (non-critical)                           │   │
│  │  • Informational security events                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.3 Incident Categories

| Category | Description | Examples |
|----------|-------------|----------|
| **KEY-COMP** | Key Compromise | Validator key theft, provider key exposure |
| **AUTH-FAIL** | Authentication Failure | Brute force, credential stuffing |
| **DATA-BREACH** | Data Breach | VEID data exposure, configuration leak |
| **MALWARE** | Malware Infection | Ransomware, cryptominers, backdoors |
| **DOS** | Denial of Service | DDoS, resource exhaustion |
| **INSIDER** | Insider Threat | Unauthorized access by personnel |
| **CONSENSUS** | Consensus Attack | Double-signing, eclipse attack |
| **ESCAPE** | Container Escape | Tenant breakout, privilege escalation |

### 1.4 Triage Decision Tree

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Incident Triage Process                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│                         Incident Detected                            │
│                              │                                       │
│                              ▼                                       │
│                    ┌─────────────────┐                              │
│                    │ Is it actually  │                              │
│                    │ an incident?    │                              │
│                    └────────┬────────┘                              │
│                             │                                        │
│              ┌──────────────┴──────────────┐                        │
│              │ NO                          │ YES                     │
│              ▼                             ▼                         │
│     Document and close          ┌─────────────────┐                 │
│     as false positive           │ Active attack   │                 │
│                                 │ in progress?    │                 │
│                                 └────────┬────────┘                 │
│                                          │                          │
│                           ┌──────────────┴──────────────┐           │
│                           │ YES                         │ NO        │
│                           ▼                             ▼           │
│                    SEV-1 or SEV-2              ┌─────────────────┐  │
│                    Immediate response          │ Data at risk?   │  │
│                                                └────────┬────────┘  │
│                                                         │           │
│                                          ┌──────────────┴──────┐    │
│                                          │ YES                 │ NO │
│                                          ▼                     ▼    │
│                                    SEV-2 or SEV-3        SEV-3/4   │
│                                    Urgent response       Scheduled │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.5 VirtEngine-Specific Triage Criteria

| Indicator | Severity | Immediate Actions |
|-----------|----------|-------------------|
| Double-sign evidence detected | CRITICAL | Stop validator, investigate |
| Validator key accessed from unknown IP | CRITICAL | Rotate key, isolate node |
| VEID encryption key in logs | CRITICAL | Rotate key, assess exposure |
| Failed VEID signature validation spike | HIGH | Block source, investigate |
| Container with host network access | HIGH | Terminate, audit provider |
| Unusual escrow fund movement | HIGH | Freeze transactions, audit |
| P2P peer count anomaly | MEDIUM | Review peer list, monitor |
| High error rate in ML inference | MEDIUM | Check for poisoning, review inputs |

---

## Part 2: Response Procedures

### 2.1 Incident Response Framework

```
┌─────────────────────────────────────────────────────────────────────┐
│                  Incident Response Lifecycle                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐        │
│   │  DETECT  │──▶│ CONTAIN  │──▶│ERADICATE │──▶│ RECOVER  │        │
│   └──────────┘   └──────────┘   └──────────┘   └──────────┘        │
│        │              │              │              │               │
│        │              │              │              │               │
│        │              ▼              ▼              ▼               │
│        │         Limit damage   Remove threat   Restore            │
│        │         Preserve       Clean systems   operations         │
│        │         evidence       Patch vulns     Verify             │
│        │                                                            │
│        │                                                            │
│        │         ┌────────────────────────────────────┐            │
│        └────────▶│         DOCUMENT & LEARN           │            │
│                  │  Throughout and after incident     │            │
│                  └────────────────────────────────────┘            │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Key Compromise Response

Key compromise is one of the most critical incidents for VirtEngine operators. This procedure covers validator, provider, and VEID encryption key compromises.

#### Validator Key Compromise Procedure

```
┌─────────────────────────────────────────────────────────────────────┐
│            Validator Key Compromise Response Runbook                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  PHASE 1: IMMEDIATE (0-5 minutes)                                   │
│  ───────────────────────────────                                    │
│  □ 1. STOP the validator process immediately                       │
│       $ systemctl stop virtengine-validator                         │
│                                                                      │
│  □ 2. DISCONNECT from network (if remote compromise suspected)     │
│       $ iptables -A INPUT -j DROP                                   │
│       $ iptables -A OUTPUT -j DROP                                  │
│                                                                      │
│  □ 3. NOTIFY security team via emergency channel                   │
│       - Use out-of-band communication (phone, Signal)              │
│       - Do NOT use potentially compromised systems                 │
│                                                                      │
│  □ 4. PRESERVE evidence (do not modify files)                      │
│       $ mkdir /evidence/$(date +%s)                                 │
│       $ cp -r /var/lib/virtengine/data/priv_validator_state.json \ │
│              /evidence/$(date +%s)/                                 │
│                                                                      │
│  PHASE 2: ASSESSMENT (5-30 minutes)                                 │
│  ──────────────────────────────────                                 │
│  □ 5. CHECK for double-signing evidence                            │
│       - Query block explorers for conflicting signatures           │
│       - Review local validator state file                          │
│                                                                      │
│  □ 6. DETERMINE compromise vector                                  │
│       - Review SSH access logs                                     │
│       - Check for unauthorized processes                           │
│       - Review recent configuration changes                        │
│                                                                      │
│  □ 7. ASSESS blast radius                                          │
│       - Were other keys on same system?                            │
│       - Is HSM/Ledger intact?                                      │
│       - Are backup keys secured?                                   │
│                                                                      │
│  PHASE 3: CONTAINMENT (30 minutes - 2 hours)                        │
│  ───────────────────────────────────────────                        │
│  □ 8. INFORM network (if validator set affected)                   │
│       - Notify other validators via secure channel                 │
│       - Consider emergency governance proposal                     │
│                                                                      │
│  □ 9. PREPARE new infrastructure                                   │
│       - Provision clean server                                     │
│       - Verify no shared credentials                               │
│                                                                      │
│  □ 10. EXECUTE key rotation                                        │
│       - Generate new key (HSM if possible)                         │
│       - Submit unjail transaction from new key                     │
│       - Update configuration                                       │
│                                                                      │
│  PHASE 4: RECOVERY (2-24 hours)                                     │
│  ──────────────────────────────                                     │
│  □ 11. RESTORE validator operations                                │
│       - Start validator with new key                               │
│       - Monitor for proper signing                                 │
│       - Verify consensus participation                             │
│                                                                      │
│  □ 12. DOCUMENT incident                                           │
│       - Complete incident report                                   │
│       - Preserve all evidence                                      │
│       - Schedule post-mortem                                       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### VEID Encryption Key Compromise

```bash
#!/bin/bash
# veid-key-compromise-response.sh

echo "=== VEID Encryption Key Compromise Response ==="
echo "Time: $(date -u)"

# Phase 1: Immediate containment
echo "[PHASE 1] Immediate containment..."

# Stop VEID-related services
systemctl stop virtengine-validator

# Preserve current key fingerprint
COMPROMISED_FP=$(cat /etc/virtengine/veid-key-fingerprint 2>/dev/null)
echo "Compromised key fingerprint: ${COMPROMISED_FP}"

# Log all recent VEID operations with this key
grep "${COMPROMISED_FP}" /var/log/virtengine/*.log > /evidence/veid-operations.log

# Phase 2: Assessment
echo "[PHASE 2] Assessing impact..."

# Count affected scopes
AFFECTED_SCOPES=$(grep -c "${COMPROMISED_FP}" /var/lib/virtengine/veid/scopes/*)
echo "Potentially affected scopes: ${AFFECTED_SCOPES}"

# Identify time window of exposure
FIRST_USE=$(grep "${COMPROMISED_FP}" /var/log/virtengine/*.log | head -1 | cut -d' ' -f1)
LAST_USE=$(grep "${COMPROMISED_FP}" /var/log/virtengine/*.log | tail -1 | cut -d' ' -f1)
echo "Exposure window: ${FIRST_USE} to ${LAST_USE}"

# Phase 3: Key rotation
echo "[PHASE 3] Rotating encryption key..."

# Generate new X25519 key pair (in HSM if available)
# This is a placeholder - actual implementation depends on key management system
echo "Generate new key using HSM or secure key generation process"
echo "Update key fingerprint in configuration"
echo "Notify affected users if required by policy"

# Phase 4: Notification
echo "[PHASE 4] Determining notification requirements..."
echo "Review regulatory obligations (GDPR, etc.)"
echo "Prepare user notification if required"
echo "Document incident for compliance"
```

### 2.3 Malware Detection and Response

#### Malware Indicators for VirtEngine Systems

| Indicator | Detection Method | Response |
|-----------|-----------------|----------|
| Unknown processes | Process monitoring | Isolate, analyze |
| Unusual CPU usage | Metrics (Prometheus) | Check for cryptominer |
| Unexpected network connections | Network monitoring | Block, investigate |
| Modified binaries | File integrity (AIDE) | Isolate, restore |
| Unusual cron jobs | Configuration audit | Remove, investigate |
| Kernel module changes | System monitoring | Immediate isolation |

#### Malware Response Procedure

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Malware Response Procedure                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  STEP 1: ISOLATE                                                    │
│  ───────────────                                                    │
│  • Disconnect system from network                                   │
│  • Do NOT power off (preserve memory)                               │
│  • Block at network level (switch/firewall)                         │
│                                                                      │
│  Commands:                                                          │
│  $ iptables -I INPUT -j DROP                                        │
│  $ iptables -I OUTPUT -j DROP                                       │
│                                                                      │
│  STEP 2: CAPTURE EVIDENCE                                           │
│  ────────────────────────                                           │
│  • Capture memory dump                                              │
│  • Capture process list                                             │
│  • Capture network connections                                      │
│  • Capture file hashes                                              │
│                                                                      │
│  Commands:                                                          │
│  $ dd if=/dev/mem of=/evidence/memory.dump bs=1M                   │
│  $ ps auxwww > /evidence/processes.txt                             │
│  $ netstat -tulpn > /evidence/network.txt                          │
│  $ ss -tulpn >> /evidence/network.txt                              │
│  $ find / -type f -mtime -1 -exec sha256sum {} \; > /evidence/files.txt │
│                                                                      │
│  STEP 3: ANALYZE                                                    │
│  ────────────────                                                   │
│  • Identify malware type                                            │
│  • Determine initial access vector                                  │
│  • Identify persistence mechanisms                                  │
│  • Assess data exfiltration                                         │
│                                                                      │
│  Analysis checklist:                                                │
│  □ Check /tmp, /var/tmp, /dev/shm for suspicious files             │
│  □ Review crontabs for all users                                   │
│  □ Check systemd services and timers                               │
│  □ Review SSH authorized_keys                                      │
│  □ Check for modified system binaries                              │
│  □ Review /etc/ld.so.preload                                       │
│  □ Check for kernel modules                                        │
│                                                                      │
│  STEP 4: ERADICATE                                                  │
│  ─────────────────                                                  │
│  • For production systems: REBUILD from clean image                │
│  • Do NOT attempt to "clean" compromised systems                   │
│  • Restore data from verified clean backups                        │
│                                                                      │
│  STEP 5: RECOVER                                                    │
│  ──────────────                                                     │
│  • Deploy clean system                                              │
│  • Restore configuration from backup                               │
│  • Apply all security patches                                       │
│  • Implement additional security controls                          │
│  • Monitor closely for re-infection                                │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.4 Data Breach Handling

#### Data Breach Response Timeline

```
┌─────────────────────────────────────────────────────────────────────┐
│                   Data Breach Response Timeline                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  HOUR 0-1: INITIAL RESPONSE                                         │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ • Confirm breach occurred                                      │ │
│  │ • Activate incident response team                              │ │
│  │ • Contain the breach (stop ongoing exfiltration)              │ │
│  │ • Preserve evidence                                            │ │
│  │ • Begin preliminary impact assessment                         │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  HOURS 1-4: ASSESSMENT                                               │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ • Determine what data was accessed/exfiltrated                │ │
│  │ • Identify affected individuals                                │ │
│  │ • Assess risk to affected individuals                         │ │
│  │ • Determine regulatory notification requirements              │ │
│  │ • Engage legal counsel                                         │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  HOURS 4-24: CONTAINMENT & DOCUMENTATION                             │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ • Complete containment measures                                │ │
│  │ • Document all findings                                        │ │
│  │ • Prepare notification drafts                                  │ │
│  │ • Begin remediation planning                                   │ │
│  │ • Continue forensic investigation                             │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  HOURS 24-72: NOTIFICATION                                           │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ • Notify regulators (GDPR: 72 hours)                          │ │
│  │ • Notify affected individuals                                  │ │
│  │ • Prepare public statement if needed                          │ │
│  │ • Implement monitoring for affected individuals               │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  DAYS 3-30: REMEDIATION                                              │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ • Complete system remediation                                  │ │
│  │ • Implement additional security controls                      │ │
│  │ • Conduct post-incident review                                │ │
│  │ • Update incident response procedures                         │ │
│  │ • Submit final regulatory reports                             │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### VEID Biometric Data Breach

When VEID biometric data may be compromised:

```markdown
## VEID Biometric Data Breach Checklist

### Immediate Actions (0-1 hour)
- [ ] Stop affected VEID services
- [ ] Preserve encryption envelope logs
- [ ] Verify X25519 key integrity
- [ ] Assess if three-signature validation was bypassed

### Assessment (1-4 hours)
- [ ] Determine scope of affected identities
- [ ] Check if data was decrypted or only ciphertext accessed
- [ ] Review ML inference logs for anomalies
- [ ] Audit signature validation records

### Classification
- [ ] Encrypted data only (lower severity)
  - X25519-XSalsa20-Poly1305 encryption intact
  - Keys not compromised
- [ ] Decrypted data accessed (CRITICAL)
  - Biometric templates exposed
  - Facial verification data accessible
  - Full incident response required

### Regulatory Considerations
- [ ] GDPR Article 33 notification (72 hours)
- [ ] Biometric data special category considerations
- [ ] Cross-border data transfer implications
- [ ] User notification requirements
```

---

## Part 3: Evidence and Communication

### 3.1 Evidence Preservation

#### Evidence Collection Procedure

```bash
#!/bin/bash
# evidence-collection.sh - Forensic evidence collection for VirtEngine

set -euo pipefail

# Configuration
EVIDENCE_DIR="/evidence/$(date +%Y%m%d-%H%M%S)"
INCIDENT_ID="${1:-unknown}"

echo "=== VirtEngine Forensic Evidence Collection ==="
echo "Incident ID: ${INCIDENT_ID}"
echo "Evidence Directory: ${EVIDENCE_DIR}"
echo "Start Time: $(date -u)"

# Create evidence directory
mkdir -p "${EVIDENCE_DIR}"
cd "${EVIDENCE_DIR}"

# System information
echo "[1/10] Collecting system information..."
{
    echo "=== System Information ==="
    date -u
    uname -a
    cat /etc/os-release
    uptime
    who
    last -20
} > system_info.txt

# Process information
echo "[2/10] Collecting process information..."
{
    echo "=== Process List ==="
    ps auxwww
    echo ""
    echo "=== Process Tree ==="
    pstree -p
    echo ""
    echo "=== Open Files ==="
    lsof +L1 2>/dev/null || true
} > processes.txt

# Network information
echo "[3/10] Collecting network information..."
{
    echo "=== Network Connections ==="
    ss -tulpn
    echo ""
    netstat -an
    echo ""
    echo "=== Routing Table ==="
    ip route
    echo ""
    echo "=== Firewall Rules ==="
    iptables -L -n -v
} > network.txt

# VirtEngine-specific logs
echo "[4/10] Collecting VirtEngine logs..."
mkdir -p logs
cp -r /var/log/virtengine/* logs/ 2>/dev/null || true
journalctl -u virtengine-validator --since "24 hours ago" > logs/validator-journal.log 2>/dev/null || true
journalctl -u provider-daemon --since "24 hours ago" > logs/provider-journal.log 2>/dev/null || true

# Configuration files
echo "[5/10] Collecting configuration..."
mkdir -p config
cp /etc/virtengine/*.toml config/ 2>/dev/null || true
cp /etc/virtengine/*.json config/ 2>/dev/null || true
# Do NOT copy private keys - just note their presence
ls -la /var/lib/virtengine/keys/ > config/key_listing.txt 2>/dev/null || true

# Blockchain state
echo "[6/10] Collecting blockchain state..."
mkdir -p chain_state
cp /var/lib/virtengine/data/priv_validator_state.json chain_state/ 2>/dev/null || true
# Capture last 100 blocks of local chain data
virtengine query block --height $(virtengine status 2>/dev/null | jq -r '.sync_info.latest_block_height') > chain_state/latest_block.json 2>/dev/null || true

# File integrity
echo "[7/10] Calculating file hashes..."
{
    find /usr/local/bin -type f -name "virtengine*" -exec sha256sum {} \;
    find /etc/virtengine -type f -exec sha256sum {} \;
} > file_hashes.txt

# User accounts and authentication
echo "[8/10] Collecting authentication data..."
{
    echo "=== User Accounts ==="
    cat /etc/passwd
    echo ""
    echo "=== Group Memberships ==="
    cat /etc/group
    echo ""
    echo "=== Sudo Configuration ==="
    cat /etc/sudoers 2>/dev/null || true
    cat /etc/sudoers.d/* 2>/dev/null || true
    echo ""
    echo "=== SSH Authorized Keys ==="
    find /home -name "authorized_keys" -exec cat {} \; 2>/dev/null || true
    cat /root/.ssh/authorized_keys 2>/dev/null || true
    echo ""
    echo "=== Recent Auth Logs ==="
    tail -500 /var/log/auth.log
} > authentication.txt

# Cron and scheduled tasks
echo "[9/10] Collecting scheduled tasks..."
{
    echo "=== System Crontabs ==="
    cat /etc/crontab
    echo ""
    ls -la /etc/cron.*
    echo ""
    echo "=== User Crontabs ==="
    for user in $(cut -f1 -d: /etc/passwd); do
        echo "User: ${user}"
        crontab -u "${user}" -l 2>/dev/null || echo "No crontab"
    done
    echo ""
    echo "=== Systemd Timers ==="
    systemctl list-timers --all
} > scheduled_tasks.txt

# Create manifest and checksums
echo "[10/10] Creating evidence manifest..."
{
    echo "Evidence Collection Manifest"
    echo "============================"
    echo "Incident ID: ${INCIDENT_ID}"
    echo "Collection Time: $(date -u)"
    echo "Collector: $(whoami)@$(hostname)"
    echo ""
    echo "Files Collected:"
    find . -type f | sort
} > MANIFEST.txt

sha256sum * > SHA256SUMS.txt

echo ""
echo "=== Evidence collection complete ==="
echo "Evidence location: ${EVIDENCE_DIR}"
echo "Remember to:"
echo "  1. Transfer evidence to secure storage"
echo "  2. Maintain chain of custody documentation"
echo "  3. Do not modify original systems"
```

#### Chain of Custody

```markdown
## Chain of Custody Log

### Evidence Item
- **Item ID:** [Auto-generated]
- **Description:** [What was collected]
- **Incident ID:** [Related incident]
- **Source System:** [Hostname/IP]

### Collection Record
| Date/Time | Action | Person | Signature |
|-----------|--------|--------|-----------|
| | Collected | | |
| | Transferred to | | |
| | Analyzed by | | |
| | Stored at | | |

### Storage Information
- **Location:** [Secure storage location]
- **Access Control:** [Who can access]
- **Encryption:** [Encryption method]
- **Retention:** [Retention period]

### Integrity Verification
- **Original Hash:** [SHA256]
- **Verification Date:** 
- **Verified By:**
- **Hash Match:** [ ] Yes [ ] No
```

### 3.2 Communication Protocols

#### Internal Communication Matrix

| Severity | Notify | Method | Timeframe |
|----------|--------|--------|-----------|
| CRITICAL | Security Lead, CTO, Legal | Phone + Signal | Immediate |
| HIGH | Security Team, Ops Lead | Slack (encrypted) | < 30 min |
| MEDIUM | Security Team | Slack | < 2 hours |
| LOW | Security Team | Email/Ticket | < 24 hours |

#### Communication Templates

**Initial Alert (Internal):**
```
SECURITY INCIDENT ALERT
=======================
Severity: [CRITICAL/HIGH/MEDIUM/LOW]
Category: [KEY-COMP/DATA-BREACH/MALWARE/etc.]
Time Detected: [UTC timestamp]
Affected Systems: [List]

Summary:
[Brief description of what was detected]

Initial Actions Taken:
[List actions already taken]

Response Team:
- Incident Commander: [Name]
- Technical Lead: [Name]
- Communications: [Name]

Next Update: [Time]
```

**Status Update (Internal):**
```
INCIDENT STATUS UPDATE #[N]
===========================
Incident ID: [ID]
Severity: [Current severity]
Status: [ACTIVE/CONTAINED/RESOLVED]

Updates Since Last Report:
[List of findings and actions]

Current Focus:
[What team is working on]

Outstanding Issues:
[List any blockers or concerns]

Next Update: [Time]
```

**External Notification (User):**
```
Subject: Important Security Notice

Dear [User],

We are writing to inform you of a security incident that may 
affect your VirtEngine account.

What Happened:
[Clear, non-technical explanation]

What Information Was Involved:
[List specific data types]

What We Are Doing:
[Actions taken to protect users]

What You Can Do:
[Specific recommended actions]

For More Information:
[Contact details and resources]

We sincerely apologize for any concern this may cause.

The VirtEngine Security Team
```

---

## Part 4: Recovery and Post-Incident

### 4.1 Recovery Procedures

#### Validator Recovery

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Validator Recovery Procedure                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  PRE-RECOVERY CHECKLIST                                             │
│  ─────────────────────                                              │
│  □ Incident fully contained                                         │
│  □ Root cause identified                                            │
│  □ Clean infrastructure prepared                                    │
│  □ New keys generated (if required)                                 │
│  □ Backups verified as clean                                        │
│  □ Security improvements identified                                 │
│                                                                      │
│  RECOVERY STEPS                                                      │
│  ──────────────                                                     │
│  1. Deploy clean server                                             │
│     □ Fresh OS installation                                         │
│     □ Apply all security patches                                    │
│     □ Implement hardening (CIS benchmarks)                          │
│                                                                      │
│  2. Install VirtEngine                                               │
│     □ Verify binary checksums                                       │
│     □ Configure with secure settings                                │
│     □ Set up HSM/Ledger connection                                  │
│                                                                      │
│  3. Restore state                                                    │
│     □ Sync from trusted peers                                       │
│     □ Verify chain state integrity                                  │
│     □ Confirm no fork/double-sign                                   │
│                                                                      │
│  4. Resume operations                                                │
│     □ Start validator in safe mode                                  │
│     □ Monitor for anomalies                                         │
│     □ Gradually increase participation                              │
│                                                                      │
│  5. Verify recovery                                                  │
│     □ Confirm signing correctly                                     │
│     □ Verify peer connectivity                                      │
│     □ Check monitoring/alerting                                     │
│                                                                      │
│  POST-RECOVERY                                                       │
│  ─────────────                                                      │
│  □ Document recovery process                                         │
│  □ Update runbooks with lessons learned                             │
│  □ Schedule post-incident review                                    │
│  □ Implement additional security controls                           │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### Provider Recovery

```bash
#!/bin/bash
# provider-recovery.sh - Provider daemon recovery procedure

echo "=== Provider Daemon Recovery ==="

# Step 1: Verify clean infrastructure
echo "[1] Verifying infrastructure..."
# Ensure Kubernetes cluster is clean
kubectl get pods -A | grep -v "Running\|Completed" && {
    echo "ERROR: Unhealthy pods detected"
    exit 1
}

# Step 2: Verify provider daemon binary
echo "[2] Verifying binary..."
EXPECTED_HASH="<official_hash>"
ACTUAL_HASH=$(sha256sum /usr/local/bin/provider-daemon | cut -d' ' -f1)
if [ "${EXPECTED_HASH}" != "${ACTUAL_HASH}" ]; then
    echo "ERROR: Binary hash mismatch"
    exit 1
fi

# Step 3: Generate new provider key (if compromised)
echo "[3] Key management..."
if [ "${KEY_COMPROMISED}" = "true" ]; then
    echo "Generating new provider key..."
    # This should be done via secure key ceremony
    # provider-daemon keys add provider --ledger
    echo "New key must be registered on-chain"
fi

# Step 4: Verify configuration
echo "[4] Verifying configuration..."
provider-daemon config validate

# Step 5: Start daemon
echo "[5] Starting provider daemon..."
systemctl start provider-daemon

# Step 6: Verify operation
echo "[6] Verifying operation..."
sleep 30
provider-daemon status

# Step 7: Verify connectivity
echo "[7] Verifying chain connectivity..."
provider-daemon query provider $(provider-daemon keys show provider -a)

echo "=== Recovery complete ==="
```

### 4.2 Post-Incident Review

#### Post-Incident Review Template

```markdown
# Post-Incident Review

## Incident Summary
- **Incident ID:** 
- **Date/Time Detected:**
- **Date/Time Resolved:**
- **Duration:**
- **Severity Level:**
- **Category:**

## Executive Summary
[2-3 paragraph summary for leadership]

## Timeline
| Time (UTC) | Event | Actor |
|------------|-------|-------|
| | Incident began | |
| | Incident detected | |
| | Response initiated | |
| | Containment achieved | |
| | Eradication completed | |
| | Recovery completed | |

## Root Cause Analysis

### What Happened
[Detailed technical description]

### Why It Happened
[Root cause(s)]

### Contributing Factors
- 
- 
- 

## Impact Assessment

### Systems Affected
- 

### Data Affected
- 

### Business Impact
- 

### User Impact
- 

## Response Effectiveness

### What Worked Well
1. 
2. 
3. 

### What Could Be Improved
1. 
2. 
3. 

### Detection Time Analysis
- Expected detection time: 
- Actual detection time: 
- Gap analysis: 

## Action Items

| Action | Owner | Due Date | Priority |
|--------|-------|----------|----------|
| | | | |
| | | | |
| | | | |

## Lessons Learned
1. 
2. 
3. 

## Metrics

| Metric | Value |
|--------|-------|
| Time to Detect (TTD) | |
| Time to Respond (TTR) | |
| Time to Contain (TTC) | |
| Time to Recover | |
| Total Downtime | |

## Review Participants
- 
- 
- 

## Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Incident Commander | | | |
| Security Lead | | | |
| Operations Lead | | | |
```

### 4.3 Improvement Implementation

```
┌─────────────────────────────────────────────────────────────────────┐
│              Post-Incident Improvement Cycle                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│                        ┌─────────────────┐                          │
│                        │  POST-INCIDENT  │                          │
│                        │     REVIEW      │                          │
│                        └────────┬────────┘                          │
│                                 │                                    │
│                                 ▼                                    │
│        ┌────────────────────────────────────────────┐              │
│        │           Identify Action Items            │              │
│        │  • Process improvements                    │              │
│        │  • Technical controls                      │              │
│        │  • Training needs                          │              │
│        │  • Documentation updates                   │              │
│        └────────────────────────┬───────────────────┘              │
│                                 │                                    │
│           ┌─────────────────────┼─────────────────────┐            │
│           ▼                     ▼                     ▼            │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐      │
│   │   Quick Wins │     │   Projects   │     │  Long-term   │      │
│   │   (< 1 week) │     │  (1-4 weeks) │     │  (> 1 month) │      │
│   └──────┬───────┘     └──────┬───────┘     └──────┬───────┘      │
│          │                    │                    │                │
│          ▼                    ▼                    ▼                │
│   ┌──────────────────────────────────────────────────────────┐    │
│   │                     Track Progress                        │    │
│   │  • Weekly security meeting review                        │    │
│   │  • Ticket/issue tracking                                 │    │
│   │  • Metrics dashboard                                     │    │
│   └──────────────────────────────────────────────────────────┘    │
│                                 │                                    │
│                                 ▼                                    │
│   ┌──────────────────────────────────────────────────────────┐    │
│   │                   Validate & Close                        │    │
│   │  • Test improvements                                     │    │
│   │  • Update documentation                                  │    │
│   │  • Conduct tabletop exercise                            │    │
│   └──────────────────────────────────────────────────────────┘    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Tabletop Exercises

### Exercise 1: Validator Key Compromise (45 minutes)

**Scenario:**
At 02:00 UTC on a Saturday, your monitoring system alerts that your validator has signed two different blocks at the same height. You have 10 minutes before the network may initiate slashing.

**Participants:**
- Incident Commander
- Validator Operator
- Security Engineer
- Communications Lead

**Inject Schedule:**

| Time | Inject |
|------|--------|
| T+0 | Alert received: Double-signing detected |
| T+5 | HSM logs show access from unknown IP |
| T+10 | Other validators asking questions on Discord |
| T+15 | Team discovers backup key file was accessed |
| T+25 | Media outlet requests comment |
| T+35 | Forensics reveals attacker had access for 48 hours |

**Discussion Questions:**
1. What are your first three actions?
2. How do you communicate with the validator community?
3. When do you involve law enforcement?
4. What evidence do you need to preserve?
5. How do you recover operations?

---

### Exercise 2: VEID Data Breach (45 minutes)

**Scenario:**
A security researcher contacts you claiming they have downloaded encrypted VEID biometric scopes from your infrastructure. They provide proof by sharing metadata from several scopes.

**Participants:**
- Incident Commander
- Security Engineer
- Legal Counsel
- Privacy Officer
- Communications Lead

**Inject Schedule:**

| Time | Inject |
|------|--------|
| T+0 | Initial researcher contact |
| T+5 | Researcher provides sample scope metadata |
| T+10 | Log review shows unauthorized API access |
| T+15 | 10,000+ scopes potentially accessed |
| T+20 | Researcher threatens public disclosure in 24 hours |
| T+30 | Investigation reveals encryption keys are safe |
| T+40 | Regulatory notification deadline approaching |

**Discussion Questions:**
1. Is this a breach if data was encrypted?
2. What are your GDPR notification obligations?
3. How do you handle the researcher's disclosure threat?
4. Which users need to be notified?
5. What's your public statement?

---

### Exercise 3: Provider Infrastructure Attack (45 minutes)

**Scenario:**
Multiple tenants report their workloads are running cryptocurrency miners instead of their applications. Investigation suggests a container escape and lateral movement.

**Participants:**
- Incident Commander
- Provider Operator
- Security Engineer
- Customer Success
- Legal Counsel

**Inject Schedule:**

| Time | Inject |
|------|--------|
| T+0 | First tenant complaint received |
| T+5 | Five more tenants report issues |
| T+10 | Host process list shows unauthorized binaries |
| T+15 | Provider key file was accessed |
| T+20 | Attacker pivoted to three other provider hosts |
| T+30 | Customer data exfiltration detected |
| T+40 | Kubernetes cluster credentials compromised |

**Discussion Questions:**
1. How do you isolate the threat while maintaining service?
2. What's your communication to affected tenants?
3. How do you handle escrow for compromised workloads?
4. When do you shut down the entire provider?
5. What's your legal liability?

---

## Incident Scenarios

### Scenario 1: Ransomware Attack

**Background:**
Your validator server displays a ransomware message demanding 100 ETH for the decryption key. The server is inaccessible.

**Response Playbook:**

```
1. IMMEDIATE ACTIONS
   □ Do NOT pay ransom
   □ Isolate affected systems
   □ Activate backup validator (if available)
   □ Notify security team and legal

2. ASSESSMENT
   □ Identify ransomware variant
   □ Determine encryption status
   □ Check backup integrity
   □ Assess spread to other systems

3. CONTAINMENT
   □ Block C2 domains/IPs
   □ Disable affected accounts
   □ Scan all connected systems
   □ Preserve evidence

4. RECOVERY
   □ Rebuild from clean backup
   □ Restore from verified backups
   □ Apply missing patches
   □ Implement additional controls

5. POST-INCIDENT
   □ Report to law enforcement
   □ Document lessons learned
   □ Update security controls
```

### Scenario 2: Insider Threat

**Background:**
A recently terminated employee still has active access credentials and has been observed accessing production systems.

**Response Playbook:**

```
1. IMMEDIATE ACTIONS
   □ Revoke ALL access immediately
   □ Change shared credentials
   □ Monitor for continued access attempts
   □ Preserve access logs

2. ASSESSMENT
   □ Review all access during employment
   □ Identify accessed systems and data
   □ Check for data exfiltration
   □ Review any changes made

3. CONTAINMENT
   □ Reset affected system credentials
   □ Review for backdoors
   □ Check for scheduled tasks
   □ Audit code changes

4. LEGAL/HR
   □ Engage legal counsel
   □ Document evidence
   □ Coordinate with HR
   □ Consider law enforcement

5. PREVENTION
   □ Review offboarding process
   □ Implement access reviews
   □ Enhance monitoring
```

### Scenario 3: DDoS Attack on Validator

**Background:**
Your validator node is experiencing a sustained DDoS attack, causing missed blocks and potential slashing risk.

**Response Playbook:**

```
1. IMMEDIATE ACTIONS
   □ Activate DDoS protection
   □ Enable rate limiting
   □ Consider temporary failover
   □ Monitor block signing

2. TECHNICAL RESPONSE
   □ Analyze attack vectors
   □ Block attacking IPs (if feasible)
   □ Scale sentry nodes
   □ Enable geographic distribution

3. NETWORK COORDINATION
   □ Alert other validators
   □ Request peer priority
   □ Coordinate defense
   □ Share attacker IPs

4. ESCALATION
   □ Engage DDoS protection provider
   □ Consider ISP-level mitigation
   □ Law enforcement if appropriate

5. POST-ATTACK
   □ Analyze attack patterns
   □ Improve defenses
   □ Update runbooks
   □ Strengthen network resilience
```

---

## Recovery Procedures

### Full Validator Recovery

```bash
#!/bin/bash
# full-validator-recovery.sh

set -euo pipefail

echo "=== Full Validator Recovery Procedure ==="
echo "This script guides through complete validator rebuild"
echo ""

# Pre-checks
echo "[PRE-CHECK] Verifying prerequisites..."
read -p "1. Is the incident fully contained? (y/n): " CONTAINED
[ "${CONTAINED}" != "y" ] && { echo "Complete containment first"; exit 1; }

read -p "2. Is root cause identified? (y/n): " ROOT_CAUSE
[ "${ROOT_CAUSE}" != "y" ] && { echo "Identify root cause first"; exit 1; }

read -p "3. Are clean backups available? (y/n): " BACKUPS
[ "${BACKUPS}" != "y" ] && { echo "Verify backups first"; exit 1; }

# Step 1: Provision clean server
echo ""
echo "[STEP 1] Provision clean server"
echo "  - Fresh OS installation"
echo "  - Apply all security patches"
echo "  - Configure firewall"
echo "  - Set up monitoring"
read -p "Press Enter when server is ready..."

# Step 2: Install VirtEngine
echo ""
echo "[STEP 2] Install VirtEngine"
echo "  Downloading official release..."
RELEASE_URL="https://github.com/virtengine/virtengine/releases/latest"
echo "  Verify checksum before installation"
read -p "Press Enter when VirtEngine is installed..."

# Step 3: Configure HSM/Ledger
echo ""
echo "[STEP 3] Configure key management"
echo "  Options:"
echo "    1. HSM (recommended)"
echo "    2. Ledger"
echo "    3. Encrypted file (not recommended)"
read -p "Select option (1/2/3): " KEY_OPTION
echo "  Configure key according to selection..."
read -p "Press Enter when key is configured..."

# Step 4: Restore state
echo ""
echo "[STEP 4] Restore blockchain state"
echo "  Options:"
echo "    1. State sync from trusted peers"
echo "    2. Restore from backup"
echo "    3. Full sync from genesis"
read -p "Select option (1/2/3): " SYNC_OPTION
echo "  Syncing chain state..."
read -p "Press Enter when sync is complete..."

# Step 5: Verify configuration
echo ""
echo "[STEP 5] Verify configuration"
virtengine validate-genesis
virtengine config validate

# Step 6: Start validator
echo ""
echo "[STEP 6] Start validator"
echo "  Starting in safe mode..."
systemctl start virtengine-validator

# Step 7: Verify operation
echo ""
echo "[STEP 7] Verify operation"
sleep 60
virtengine status
echo ""
echo "Verify:"
echo "  - Blocks are being signed"
echo "  - Peer connections are healthy"
echo "  - Monitoring is receiving data"

echo ""
echo "=== Recovery Complete ==="
echo "Continue to monitor closely for 24 hours"
```

### VEID System Recovery

```bash
#!/bin/bash
# veid-recovery.sh

echo "=== VEID System Recovery ==="

# Step 1: Key rotation
echo "[1] Rotating VEID encryption keys..."
echo "  This requires secure key generation ceremony"
echo "  New keys must be registered on-chain"

# Step 2: Re-encrypt affected scopes
echo "[2] Re-encrypting affected scopes..."
echo "  Scopes encrypted with compromised key:"
echo "  - Must be re-encrypted with new key"
echo "  - Users may need to re-submit data"

# Step 3: Verify three-signature validation
echo "[3] Verifying signature validation..."
echo "  Ensure all three signatures are being validated:"
echo "  - Client signature (capture app)"
echo "  - User signature (wallet)"
echo "  - Salt binding (replay prevention)"

# Step 4: ML model verification
echo "[4] Verifying ML models..."
echo "  Check for model poisoning or tampering"
echo "  Verify model hashes match expected values"

# Step 5: Audit trail review
echo "[5] Reviewing audit trail..."
echo "  Verify all VEID operations are being logged"
echo "  Check for gaps during incident period"

echo "=== VEID Recovery Complete ==="
```

---

## Quick Reference Cards

### Incident Response Quick Card

```
┌─────────────────────────────────────────────────────────────────────┐
│            INCIDENT RESPONSE QUICK REFERENCE                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. STAY CALM - Follow the process                                  │
│                                                                      │
│  2. ASSESS                                                          │
│     • What happened?                                                │
│     • Is it still happening?                                        │
│     • What's affected?                                              │
│                                                                      │
│  3. CONTAIN                                                         │
│     • Stop the bleeding                                             │
│     • Isolate affected systems                                      │
│     • Preserve evidence                                             │
│                                                                      │
│  4. ESCALATE                                                        │
│     CRITICAL: Security Lead + CTO immediately                       │
│     HIGH: Security Team within 30 min                               │
│     MEDIUM: Security Team within 4 hours                            │
│                                                                      │
│  5. DOCUMENT                                                        │
│     • Time of all actions                                           │
│     • What you observed                                             │
│     • What you did                                                  │
│                                                                      │
│  EMERGENCY CONTACTS:                                                │
│  Security Lead: [Phone]                                             │
│  On-Call: [Phone]                                                   │
│  Legal: [Phone]                                                     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Compromise Quick Card

```
┌─────────────────────────────────────────────────────────────────────┐
│               KEY COMPROMISE QUICK REFERENCE                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  VALIDATOR KEY COMPROMISED:                                         │
│  1. STOP validator: systemctl stop virtengine-validator             │
│  2. DISCONNECT network: iptables -A INPUT/OUTPUT -j DROP            │
│  3. NOTIFY: Security Lead (phone)                                   │
│  4. PRESERVE: Copy validator state before any changes               │
│                                                                      │
│  PROVIDER KEY COMPROMISED:                                          │
│  1. REVOKE: Remove provider from active set                         │
│  2. STOP: Stop provider daemon                                      │
│  3. NOTIFY: Affected tenants                                        │
│  4. PRESERVE: Copy all logs and state                               │
│                                                                      │
│  VEID KEY COMPROMISED:                                              │
│  1. STOP: VEID-related services                                     │
│  2. ASSESS: Which scopes were affected?                             │
│  3. ROTATE: Generate new encryption keys                            │
│  4. NOTIFY: Assess regulatory requirements                          │
│                                                                      │
│  DO NOT:                                                            │
│  ✗ Attempt to "clean" the key                                      │
│  ✗ Use the key for any purpose                                     │
│  ✗ Delay notification                                               │
│  ✗ Destroy evidence                                                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Evidence Preservation Quick Card

```
┌─────────────────────────────────────────────────────────────────────┐
│             EVIDENCE PRESERVATION QUICK REFERENCE                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  RULE #1: DO NOT MODIFY ORIGINAL SYSTEMS                            │
│                                                                      │
│  COLLECT (in order of volatility):                                  │
│  1. Memory contents (if possible)                                   │
│  2. Running processes                                               │
│  3. Network connections                                             │
│  4. System logs                                                     │
│  5. Application logs                                                │
│  6. Configuration files                                             │
│  7. File system metadata                                            │
│                                                                      │
│  COMMANDS:                                                          │
│  ps auxwww > processes.txt                                          │
│  netstat -tulpn > network.txt                                       │
│  ss -tulpn >> network.txt                                           │
│  cp -r /var/log/virtengine logs/                                    │
│  journalctl --since "1 day ago" > journal.log                       │
│                                                                      │
│  HASH EVERYTHING:                                                   │
│  sha256sum * > SHA256SUMS.txt                                       │
│                                                                      │
│  CHAIN OF CUSTODY:                                                  │
│  • Document who collected what                                      │
│  • Document where evidence is stored                                │
│  • Document all access to evidence                                  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Additional Resources

### VirtEngine Security Documentation

- [Security Best Practices](./security-best-practices.md)
- [Threat Modeling](./threat-modeling.md)
- [Security Architecture](./../security/architecture.md)
- [VEID Encryption Specification](./../security/veid-encryption.md)

### External Resources

- [NIST Computer Security Incident Handling Guide](https://csrc.nist.gov/publications/detail/sp/800-61/rev-2/final)
- [SANS Incident Handler's Handbook](https://www.sans.org/white-papers/33901/)
- [FIRST CSIRT Services Framework](https://www.first.org/standards/frameworks/csirts/)

### Contact Information

```
SECURITY TEAM CONTACTS
======================
Security Lead:      [Name] - [Phone] - [Email]
Security Engineer:  [Name] - [Phone] - [Email]
On-Call Rotation:   [PagerDuty/OpsGenie link]

ESCALATION CONTACTS
===================
CTO:                [Name] - [Phone]
Legal Counsel:      [Name] - [Phone]
PR/Communications:  [Name] - [Phone]

EXTERNAL CONTACTS
=================
Forensics Provider: [Company] - [Phone]
Legal Firm:         [Company] - [Phone]
Cyber Insurance:    [Company] - [Policy #]
Law Enforcement:    [Local FBI/Cybercrime]
```

---

## Module Completion

### Assessment Criteria

To complete this module, participants must:

1. **Written Assessment** (30% of grade)
   - Complete incident response quiz (passing score: 85%)
   - Submit incident classification exercise

2. **Practical Exercises** (40% of grade)
   - Participate in at least one tabletop exercise
   - Complete evidence collection exercise
   - Draft incident communication

3. **Scenario Analysis** (30% of grade)
   - Lead or participate in full incident simulation
   - Complete post-incident review document

### Certification

Upon successful completion:
- Incident Response Certification issued
- Added to on-call rotation eligibility
- Access to incident response tools and resources
- Invitation to incident response community

---

**Module Version:** 1.0  
**Last Review:** 2024-01  
**Next Review:** 2024-07  
**Owner:** VirtEngine Security Team