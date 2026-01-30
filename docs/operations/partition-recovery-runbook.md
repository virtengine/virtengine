# Network Partition Recovery Runbook

This document provides operational procedures for detecting, diagnosing, and recovering from network partition events in VirtEngine blockchain networks.

## Table of Contents

1. [Overview](#overview)
2. [Partition Detection](#partition-detection)
3. [Diagnosis Procedures](#diagnosis-procedures)
4. [Recovery Procedures](#recovery-procedures)
5. [Post-Recovery Validation](#post-recovery-validation)
6. [Metrics and Monitoring](#metrics-and-monitoring)
7. [Preventive Measures](#preventive-measures)

---

## Overview

### What is a Network Partition?

A network partition occurs when a subset of validators loses connectivity with another subset, causing the network to split into isolated groups. This can lead to:

- **Consensus stall**: If no group has sufficient quorum (2/3+1 of voting power)
- **Fork risk**: If multiple groups continue producing blocks independently
- **State divergence**: Different groups may have different chain states

### Partition Types

| Type | Description | Consensus Impact |
|------|-------------|------------------|
| **Simple** | Network split into two equal groups | Consensus stalls (no quorum) |
| **Majority/Minority** | One group has >2/3 voting power | Majority continues, minority stalls |
| **Multi-way** | Network split into 3+ groups | Consensus stalls in all groups |
| **Asymmetric** | One-way communication failure | Unpredictable behavior |
| **Byzantine** | Malicious nodes isolated | Network continues if honest nodes have quorum |

---

## Partition Detection

### Symptoms of Network Partition

1. **Block Production Stall**
   - No new blocks for extended period (>30 seconds)
   - Block height not advancing
   - `virtengine status` shows stale block height

2. **Validator Connectivity Issues**
   - Validators reporting peer disconnections
   - Reduced peer count in `virtengine status`
   - P2P layer errors in logs

3. **Consensus Warnings**
   - Timeouts in prevote/precommit rounds
   - Missing votes from subset of validators
   - Round increments without finalization

### Detection Commands

```bash
# Check current block height and time
virtengine status 2>&1 | jq '.sync_info'

# Check connected peers
virtengine status 2>&1 | jq '.node_info.listen_addr'

# Check validator voting power
virtengine query staking validators --output json | jq '.validators[] | {moniker: .description.moniker, power: .tokens}'

# Check consensus state
curl localhost:26657/consensus_state | jq '.result.round_state'
```

### Monitoring Alerts

Configure alerts for:

| Metric | Threshold | Action |
|--------|-----------|--------|
| `blocks_since_last` | >30s | Warning: Potential partition |
| `connected_peers` | <50% of validators | Critical: Likely partition |
| `consensus_rounds_without_commit` | >5 | Warning: Consensus issues |
| `validator_missing_votes` | >1/3 of validators | Critical: Partition detected |

---

## Diagnosis Procedures

### Step 1: Identify Partition Boundaries

1. **Collect peer lists from all validators**

   ```bash
   # On each validator node
   virtengine status 2>&1 | jq '.peers'
   ```

2. **Build connectivity matrix**

   Create a matrix showing which validators can reach each other:

   | | V0 | V1 | V2 | V3 |
   |---|---|---|---|---|
   | **V0** | - | ✓ | ✗ | ✗ |
   | **V1** | ✓ | - | ✗ | ✗ |
   | **V2** | ✗ | ✗ | - | ✓ |
   | **V3** | ✗ | ✗ | ✓ | - |

   This shows a partition: {V0, V1} and {V2, V3}

3. **Check network infrastructure**

   ```bash
   # Test connectivity between validator IPs
   ping -c 3 <validator_ip>
   traceroute <validator_ip>
   
   # Check firewall rules
   iptables -L -n | grep <validator_port>
   ```

### Step 2: Assess Partition Impact

1. **Calculate voting power distribution**

   ```bash
   # Get voting power per validator
   virtengine query staking validators --output json | \
     jq '.validators[] | {moniker: .description.moniker, power: (.tokens | tonumber)}'
   ```

2. **Determine if any group has quorum**

   - Quorum requires >2/3 of total voting power
   - If no group has quorum, consensus has stalled
   - If one group has quorum, that group's chain is canonical

### Step 3: Identify Root Cause

Common causes and their indicators:

| Cause | Indicators | Resolution |
|-------|------------|------------|
| **Cloud provider outage** | Multiple nodes in same region affected | Wait for provider recovery |
| **DNS failure** | Nodes can't resolve peer hostnames | Use IP addresses, fix DNS |
| **Firewall changes** | Sudden connectivity loss | Revert firewall changes |
| **Network congestion** | High latency, packet loss | Increase bandwidth, optimize routing |
| **DDoS attack** | Traffic spikes, resource exhaustion | Enable DDoS protection |

---

## Recovery Procedures

### Scenario A: Simple Partition (No Quorum)

When the network is split and no group has quorum:

1. **Do NOT restart nodes immediately** - This can cause state corruption

2. **Identify and fix the network issue**

   ```bash
   # Common fixes
   # Fix firewall rules
   sudo ufw allow from <validator_ip> to any port 26656
   
   # Restart networking
   sudo systemctl restart networking
   ```

3. **Verify network connectivity is restored**

   ```bash
   # From each validator, confirm can reach others
   nc -zv <peer_ip> 26656
   ```

4. **Wait for automatic recovery**

   - CometBFT will automatically reconnect peers
   - Consensus will resume once quorum is restored
   - Monitor block production resuming

5. **If automatic recovery fails (>5 minutes)**

   ```bash
   # Restart P2P layer without full restart
   virtengine unsafe-reset-all --keep-addr-book
   virtengine start
   ```

### Scenario B: Majority/Minority Partition

When one group has quorum and continued producing blocks:

1. **Identify the canonical chain** (majority group's chain)

   ```bash
   # Check block heights on all validators
   # Majority group will have higher block height
   virtengine status 2>&1 | jq '.sync_info.latest_block_height'
   ```

2. **Restore network connectivity** (same as Scenario A)

3. **Minority nodes will automatically sync**

   - State sync will catch up minority nodes
   - Monitor sync progress:
   
   ```bash
   virtengine status 2>&1 | jq '.sync_info.catching_up'
   ```

4. **Verify state consistency**

   ```bash
   # Compare app hash at same height across nodes
   virtengine query block <height> --output json | jq '.block.header.app_hash'
   ```

### Scenario C: Multi-Way Partition

When network is split into 3+ groups:

1. **Prioritize restoring largest group connectivity first**

2. **Follow Scenario A procedure for full recovery**

3. **All groups will need to sync to the canonical chain**

### Scenario D: Byzantine Node Isolation

When specific nodes are intentionally isolated:

1. **Verify the node is actually Byzantine** (producing invalid blocks, equivocating)

2. **If intentional isolation is correct:**
   - Document the isolation event
   - Consider governance proposal to slash/jail the validator

3. **If isolation is a mistake:**
   - Restore connectivity following Scenario A

---

## Post-Recovery Validation

### Checklist

After network partition recovery, verify:

- [ ] All validators connected and syncing
- [ ] Block production resumed at normal rate
- [ ] No chain forks (all nodes same block hash at same height)
- [ ] Consensus messages flowing normally
- [ ] Pending transactions being processed
- [ ] State hashes consistent across all nodes

### Verification Commands

```bash
# 1. Verify all validators online
virtengine query staking validators --output json | \
  jq '.validators[] | select(.status == "BOND_STATUS_BONDED") | .description.moniker'

# 2. Verify block production (watch for increasing height)
watch -n 1 "virtengine status 2>&1 | jq '.sync_info.latest_block_height'"

# 3. Verify consensus participation
curl localhost:26657/consensus_state | jq '.result.round_state.votes'

# 4. Compare state hashes across validators
# Run on each validator and compare results
virtengine status 2>&1 | jq '.sync_info.latest_app_hash'

# 5. Check for stuck transactions
virtengine query txs --query "tx.height>0" --limit 10
```

### Post-Incident Report Template

```markdown
## Partition Incident Report

**Date/Time:** [Partition start] - [Partition end]
**Duration:** [Total partition duration]
**Affected Validators:** [List of affected validators]

### Timeline
- [Time]: Partition detected
- [Time]: Root cause identified
- [Time]: Network connectivity restored
- [Time]: Consensus resumed
- [Time]: All nodes synced

### Impact
- Blocks missed: [Number]
- Transactions delayed: [Estimate]
- Services affected: [List]

### Root Cause
[Description of what caused the partition]

### Recovery Actions
[Steps taken to recover]

### Lessons Learned
[What can be done to prevent/detect/recover faster]

### Action Items
- [ ] [Preventive measure 1]
- [ ] [Preventive measure 2]
```

---

## Metrics and Monitoring

### Key Metrics to Track

#### Network Health Metrics

| Metric | Description | Collection Method |
|--------|-------------|-------------------|
| `peer_count` | Number of connected peers | Prometheus/Status API |
| `block_height` | Current block height | Status API |
| `time_since_last_block` | Seconds since last block | Calculated |
| `consensus_round` | Current consensus round | Consensus API |

#### Partition-Specific Metrics

| Metric | Description | Purpose |
|--------|-------------|---------|
| `partition_duration_seconds` | How long partition lasted | Recovery SLA |
| `time_to_first_block_after_heal` | Recovery speed | Performance benchmark |
| `time_to_full_consensus` | Full recovery time | SLA measurement |
| `messages_replayed` | Duplicate messages detected | Security monitoring |
| `state_sync_duration` | Time to sync after partition | Capacity planning |

### Prometheus Alert Rules

```yaml
groups:
  - name: partition_alerts
    rules:
      - alert: PotentialNetworkPartition
        expr: increase(blocks_total[1m]) == 0
        for: 30s
        labels:
          severity: warning
        annotations:
          summary: "No blocks produced in 30 seconds"
          
      - alert: NetworkPartitionDetected
        expr: connected_peers < (total_validators * 0.5)
        for: 10s
        labels:
          severity: critical
        annotations:
          summary: "Connected to less than 50% of validators"
          
      - alert: ConsensusMissingVotes
        expr: missing_votes_ratio > 0.33
        for: 5s
        labels:
          severity: critical
        annotations:
          summary: "Missing votes from >33% of validators"
```

### Grafana Dashboard

Create a dashboard with:

1. **Network Connectivity Panel**
   - Peer count over time
   - Peer count per validator

2. **Block Production Panel**
   - Block height over time
   - Time between blocks

3. **Consensus Health Panel**
   - Rounds per block
   - Missing votes heatmap

4. **Partition Events Panel**
   - Partition duration histogram
   - Recovery time percentiles

---

## Preventive Measures

### Infrastructure Recommendations

1. **Geographic Distribution**
   - Spread validators across multiple regions/providers
   - Avoid >1/3 of voting power in any single region

2. **Network Redundancy**
   - Multiple network paths between validators
   - Consider private peering for critical connections

3. **Sentry Node Architecture**
   - Use sentry nodes to protect validators
   - Multiple sentries per validator

### Configuration Best Practices

```toml
# config.toml recommendations

[p2p]
# More persistent peers for resilience
persistent_peers = "<comma-separated peer list with >50% of validators>"

# Allow more peers for redundancy
max_num_inbound_peers = 40
max_num_outbound_peers = 20

# Enable peer exchange
pex = true

# Faster reconnection
dial_timeout = "3s"

[consensus]
# Tune timeouts for your network
timeout_propose = "3s"
timeout_prevote = "1s"
timeout_precommit = "1s"
timeout_commit = "5s"
```

### Regular Testing

1. **Scheduled Partition Drills**
   - Monthly: Test documentation and procedures
   - Quarterly: Simulate partition on testnet
   - Annually: Full partition drill on mainnet (with notice)

2. **Automated Testing**
   - Run partition recovery tests in CI/CD
   - See `tests/integration/partition/` for test suite

### Emergency Contacts

Maintain up-to-date contact list for:

- [ ] All validator operators
- [ ] Network infrastructure providers
- [ ] On-call engineering team
- [ ] Community communication channels

---

## Appendix

### Related Documentation

- [CometBFT Consensus Documentation](https://docs.cometbft.com/v0.38/spec/consensus/)
- [VirtEngine Network Configuration](../configuration/network.md)
- [Validator Operations Guide](./validator-operations.md)

### Glossary

| Term | Definition |
|------|------------|
| **Quorum** | >2/3 of voting power required for consensus |
| **Finality** | Block is irreversibly committed |
| **State Sync** | Fast sync method to catch up on chain state |
| **Byzantine** | Malicious or faulty validator behavior |
| **Split-brain** | Network partition causing competing chains |

### Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-30 | Initial version |
