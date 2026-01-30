# Runbook: Low Validator Count

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | LowValidatorCount |
| Severity | Critical |
| Service | virtengine-chain |
| Tier | Tier 0 |
| SLO Impact | Chain consensus resilience |

## Summary

This alert fires when the number of active validators falls below the safe threshold. CometBFT requires >2/3 of voting power to be online for consensus. This alert warns before reaching that critical threshold.

- **Warning**: Active validators < 67% of bonded set
- **Critical**: Active validators approaching 2/3 threshold

## Impact

- **Critical**: Risk of chain halt if more validators go offline
- **High**: Reduced security - fewer validators signing blocks
- **High**: Higher centralization risk
- **Medium**: May indicate coordinated attack or infrastructure issue

## Prerequisites

- Access to validator communication channels
- Understanding of validator set composition
- Access to staking module queries

## Diagnostic Steps

### 1. Check Current Validator Status

```bash
# Get active validator count
virtengined q staking validators --status bonded -o json | jq '.validators | length'

# Get validator details
virtengined q staking validators --status bonded -o json | jq '.validators[] | {moniker: .description.moniker, voting_power: .tokens, jailed: .jailed}'

# Check consensus validators
curl -s http://localhost:26657/validators | jq '.result.total, .result.validators | length'
```

### 2. Identify Offline Validators

```bash
# Compare expected vs active
EXPECTED=$(virtengined q staking validators --status bonded -o json | jq '.validators | length')
ACTIVE=$(curl -s http://localhost:26657/validators | jq '.result.validators | length')
echo "Expected: $EXPECTED, Active: $ACTIVE, Missing: $((EXPECTED - ACTIVE))"

# Find validators with missed blocks
virtengined q slashing signing-infos --limit 100 -o json | jq '.info[] | select(.missed_blocks_counter > 10) | {address: .address, missed: .missed_blocks_counter}'

# Check recently jailed validators
virtengined q staking validators --status unbonding -o json | jq '.validators[] | select(.jailed == true) | {moniker: .description.moniker, jailed_until: .unbonding_time}'
```

### 3. Check Block Signatures

```bash
# Get recent block signatures
HEIGHT=$(virtengined status | jq -r '.SyncInfo.latest_block_height')
curl -s "http://localhost:26657/block?height=$HEIGHT" | jq '.result.block.last_commit.signatures | length'

# Check signature percentage
TOTAL=$(virtengined q staking validators --status bonded -o json | jq '.validators | length')
SIGNED=$(curl -s "http://localhost:26657/block?height=$HEIGHT" | jq '.result.block.last_commit.signatures | map(select(.signature != null)) | length')
echo "Signature rate: $SIGNED / $TOTAL"
```

### 4. Check for Patterns

```bash
# Check if validators went offline at similar times
grep -i "validator.*offline\|peer.*disconnected" /var/log/virtengine/node.log | tail -50

# Check for network issues
virtengined net info | jq '.peers | length'

# Check for infrastructure correlation (same cloud provider, region, etc.)
# This requires validator registry/metadata
```

## Resolution Steps

### Scenario 1: Validators Jailed for Downtime

```bash
# 1. Contact validator operators
# Use validator emergency contact channels

# 2. Help validators unjail
# Validator runs:
virtengined tx slashing unjail --from <validator-key> --chain-id <chain-id> --yes

# 3. Monitor for recovery
watch -n 5 'virtengined q staking validators --status bonded -o json | jq ".validators | length"'
```

### Scenario 2: Network Partition

```bash
# 1. Identify isolated validators
for validator in $VALIDATOR_IPS; do
  echo "=== $validator ==="
  ssh user@$validator "virtengined net info | jq '.n_peers'" 2>/dev/null || echo "UNREACHABLE"
done

# 2. Check network connectivity between validators
for src in $VALIDATOR_IPS; do
  for dst in $VALIDATOR_IPS; do
    ssh user@$src "nc -zv $dst 26656 2>&1 | grep -q succeeded && echo '$src -> $dst: OK' || echo '$src -> $dst: FAIL'"
  done
done

# 3. If partition found, check:
# - Cloud provider status
# - Firewall rules
# - BGP/routing issues

# 4. Use alternative network paths if available
# Validators can add additional seeds/persistent peers
```

### Scenario 3: Coordinated Infrastructure Issue

```bash
# 1. Check cloud provider status pages
# AWS: https://status.aws.amazon.com
# GCP: https://status.cloud.google.com
# Azure: https://status.azure.com

# 2. If cloud outage, contact affected validator operators
# Discuss migration to backup infrastructure

# 3. For critical situations, consider emergency validator set changes
# This requires governance and is a last resort
```

### Scenario 4: Double-Signing / Slashing Event

```bash
# 1. Check for slashing events
virtengined q slashing signing-infos --limit 100 -o json | jq '.info[] | select(.tombstoned == true)'

# 2. Check recent slashing transactions
virtengined q txs --events 'message.action=/cosmos.slashing.v1beta1.MsgUnjail' --limit 10

# 3. If double-signing detected:
# - Validator is tombstoned permanently
# - Investigate cause (key compromise, software bug, operator error)
# - May need to coordinate new validator entry
```

## Emergency Procedures

### If Approaching 1/3 Threshold

```bash
# 1. Immediate escalation to all stakeholders
echo "EMERGENCY: Validator count critical - risk of chain halt"

# 2. Contact ALL validator operators immediately
# Use all available channels: Slack, Discord, phone, SMS

# 3. Identify validators that can come online quickly
# Prioritize by:
# - Voting power
# - Response time
# - Technical readiness

# 4. Prepare for potential chain halt
# See: runbooks/block-stalled.md

# 5. Consider emergency governance to:
# - Reduce unbonding period temporarily
# - Fast-track new validators
```

### If Chain Halts Due to Validator Loss

```bash
# 1. Follow block-stalled.md runbook

# 2. Coordinate validator restart
# All remaining validators must coordinate restart

# 3. If >1/3 of voting power permanently lost:
# - This is a catastrophic scenario
# - May require hard fork with new validator set
# - Escalate to executive level immediately
```

## Recovery Verification

```bash
# 1. Verify validator count is recovering
watch -n 10 'virtengined q staking validators --status bonded -o json | jq ".validators | length"'

# 2. Check all validators are signing blocks
HEIGHT=$(virtengined status | jq -r '.SyncInfo.latest_block_height')
curl -s "http://localhost:26657/block?height=$HEIGHT" | jq '.result.block.last_commit.signatures | map(select(.signature != null)) | length'

# 3. Verify no validators are jailed
virtengined q staking validators --status bonded -o json | jq '[.validators[] | select(.jailed == true)] | length'

# 4. Check consensus health
curl -s http://localhost:26657/consensus_state | jq '.result.round_state.height_vote_set[0]'
```

## Prevention

### Validator Monitoring

All validators should have:
- Uptime monitoring
- Missed block alerts
- Peer connectivity monitoring
- Hardware health monitoring

### Validator Diversity

Ensure validator set diversity:
- Multiple cloud providers
- Multiple geographic regions
- Multiple client implementations (if available)
- Independent infrastructure operators

### Communication

Maintain:
- Validator emergency contact list
- Multiple communication channels
- Regular validator coordination calls
- Incident response procedures

## Escalation

**Immediate L2 escalation**:
- Always for this alert type

**Immediate L3 escalation if**:
- Active validators < 75% of bonded set
- Multiple validators unreachable
- Suspected coordinated attack

**Executive escalation if**:
- Risk of chain halt (approaching 2/3 threshold)
- Suspected security incident
- Potential need for emergency governance

## Post-Incident

1. **Mandatory postmortem** for any validator count drop
2. Review:
   - Validator infrastructure resilience
   - Communication procedures
   - Monitoring coverage
3. Update:
   - Validator requirements/SLAs
   - Emergency procedures
   - Escalation contacts

## Related Alerts

- `BlockProductionStalled` - Consequence if too few validators
- `NodeDown` - Individual validator failure
- `ValidatorMissedBlocks` - Early warning
- `NetworkPartitionDetected` - Network issues

## References

- [Validator Operations Guide](../../validator-operations.md)
- [CometBFT Consensus](https://docs.cometbft.com/v0.38/spec/consensus/)
- [Staking Module](../../x/staking/README.md)
