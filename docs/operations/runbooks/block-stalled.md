# Runbook: Block Production Stalled

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | BlockProductionStalled |
| Severity | Critical |
| Service | virtengine-chain |
| Tier | Tier 0 |
| SLO Impact | SLO-CHAIN-001 (Block time <7s P95) |

## Summary

This alert fires when no new blocks have been produced for more than 60 seconds. This indicates a critical consensus failure that halts the entire chain.

## Impact

- **Critical**: Chain is completely halted
- **Critical**: No transactions can be processed
- **Critical**: All on-chain operations blocked
- **High**: Provider deployments cannot be created/modified
- **High**: VEID verifications stuck in pending state

## Prerequisites

- SSH access to validator nodes
- Access to validator private channels (Slack/Discord)
- Understanding of CometBFT consensus

## Diagnostic Steps

### 1. Confirm Chain Halt

```bash
# Check multiple nodes
for node in node1 node2 node3; do
  echo "=== $node ==="
  curl -s http://$node:26657/status | jq '.result.sync_info.latest_block_height, .result.sync_info.latest_block_time'
done

# Check if block time is stale
date -d "$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_time')"
```

### 2. Check Validator Status

```bash
# Get number of validators
curl -s http://localhost:26657/validators | jq '.result.total'

# Check which validators have signed recent blocks
virtengined q slashing signing-infos --limit 100 | grep -A2 "jailed_until"
```

### 3. Check Consensus State

```bash
# Get consensus state
curl -s http://localhost:26657/consensus_state | jq '.result.round_state'

# Check dump consensus state (detailed)
curl -s http://localhost:26657/dump_consensus_state | jq '.result.peers' | head -50
```

### 4. Check Network Connectivity

```bash
# Check peer count on each validator
for node in validator1 validator2 validator3; do
  echo "=== $node peers ==="
  ssh user@$node "virtengined net info | jq '.n_peers'"
done

# Check if validators can see each other
virtengined net info | jq '.peers[].node_info.moniker'
```

## Root Cause Analysis

### Scenario 1: Insufficient Validators Online

**Symptom**: Less than 2/3 of voting power online

```bash
# Check voting power
curl -s http://localhost:26657/validators | jq '.result.validators[] | {address: .address, voting_power: .voting_power}'

# Sum online voting power vs total
virtengined q staking validators | grep -E "(operator_address|tokens):"
```

**Resolution**:
1. Contact offline validators immediately
2. Check if validators are jailed
3. Coordinate validator restart

### Scenario 2: Network Partition

**Symptom**: Validators cannot communicate

```bash
# Check if validators are reachable
for validator in $VALIDATOR_IPS; do
  echo "=== $validator ==="
  nc -zv $validator 26656 && echo "REACHABLE" || echo "UNREACHABLE"
done

# Check DNS resolution
dig +short validator1.virtengine.io
```

**Resolution**:
1. Check cloud provider status (AWS/GCP/Azure)
2. Check firewall rules
3. Use alternative network paths
4. Consider emergency P2P bootstrap

### Scenario 3: State Divergence

**Symptom**: Validators disagree on state

```bash
# Check app hash on multiple validators
for node in validator1 validator2 validator3; do
  echo "=== $node ==="
  ssh user@$node "virtengined status | jq '.SyncInfo.latest_block_hash'"
done
```

**Resolution**:
1. Identify which validator(s) have wrong state
2. Stop divergent validators
3. Restore from consensus snapshot
4. Coordinate restart

### Scenario 4: CometBFT Bug/Deadlock

**Symptom**: Process running but consensus stuck

```bash
# Check goroutine stack
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | head -200

# Check for mutex contention
curl -s http://localhost:6060/debug/pprof/mutex?debug=1
```

**Resolution**:
1. Capture debug info (pprof, logs)
2. Coordinate rolling restart of validators
3. If persists, escalate to Cosmos SDK team

## Resolution Steps

### Emergency: Coordinate Validator Restart

This should be done in coordination with the validator set:

```bash
# 1. Announce in validator channel
echo "EMERGENCY: Block production stalled. Coordinating restart at $TIME."

# 2. Stop all validators simultaneously
# Each validator runs:
sudo systemctl stop virtengined

# 3. Verify all stopped
for validator in $VALIDATOR_IPS; do
  ssh user@$validator "systemctl is-active virtengined"
done

# 4. Clear WAL (if consensus corruption suspected)
# Each validator runs:
rm -rf /var/lib/virtengine/data/cs.wal

# 5. Start validators in sequence
# Start with highest voting power first
ssh user@validator1 "sudo systemctl start virtengined"
sleep 10
ssh user@validator2 "sudo systemctl start virtengined"
sleep 10
# Continue...

# 6. Monitor consensus recovery
watch -n 2 'curl -s http://localhost:26657/consensus_state | jq ".result.round_state.height_vote_set[0]"'
```

### If Validator Set Split (Fork Recovery)

```bash
# 1. Identify canonical chain (highest voting power)
# 2. Export state at last common block
virtengined export --height <last-common-height> > genesis_export.json

# 3. Coordinate genesis restart with updated validators
# This is a major incident - escalate immediately
```

## Recovery Verification

```bash
# 1. Verify blocks are being produced
watch -n 5 'curl -s http://localhost:26657/status | jq ".result.sync_info.latest_block_height"'

# 2. Check block time is reasonable (<7s)
curl -s http://localhost:26657/consensus_params | jq '.result.consensus_params.block.time_iota_ms'

# 3. Submit test transaction
virtengined tx bank send test-wallet test-recipient 1uve --yes

# 4. Verify all validators signing
virtengined q slashing signing-infos --limit 100 | grep -c "false"
```

## Escalation

**Immediate Escalation Required**:
- This is always a Tier 0 incident
- Notify all stakeholders immediately
- Activate incident bridge/war room

**Escalate to Security Team if**:
- Signs of malicious activity
- Unexpected state changes
- Double-signing detected

## Communication Template

```
🚨 INCIDENT: VirtEngine Chain Halted

Status: Block production has stopped
Time: [UTC timestamp]
Last Block: [block height] at [block time]
Impact: All on-chain operations unavailable

Current Actions:
- Investigating root cause
- Coordinating with validator set
- ETA: Investigating

Updates: #incident-[number]
```

## Post-Incident

1. **Mandatory postmortem** within 48 hours
2. Collect from all validators:
   - Logs from 30 minutes before incident
   - Consensus state dumps
   - Network metrics
3. Review and update:
   - Validator communication procedures
   - Monitoring thresholds
   - Recovery procedures

## Related Alerts

- `NodeDown` - Individual node failures
- `LowValidatorCount` - Precursor to this alert
- `ConsensusRoundTimeout` - Early warning
- `NetworkPartitionDetected` - Network split

## References

- [CometBFT Consensus](https://docs.cometbft.com/v0.38/spec/consensus/)
- [Validator Operations](../../validator-operations.md)
- [Emergency Procedures](../../emergency-procedures.md)
