# Runbook: Node Down

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | NodeDown |
| Severity | Critical |
| Service | virtengine-node |
| Tier | Tier 0 |
| SLO Impact | SLO-NODE-001 (99.95% availability) |

## Summary

This alert fires when a VirtEngine node becomes unresponsive for more than 1 minute. This is a critical alert that directly impacts chain availability.

## Impact

- **High**: Node cannot participate in consensus
- **High**: Transactions may not be processed
- **Medium**: If validator node, missed blocks result in slashing
- **Medium**: Reduced network resilience

## Prerequisites

- SSH access to the affected node
- Access to infrastructure management (K8s/VM console)
- `virtengined` CLI access

## Diagnostic Steps

### 1. Verify the Alert

```bash
# Check if node is reachable
ping <node-ip>

# Check if RPC is responding
curl -s http://<node-ip>:26657/status | jq '.result.sync_info'

# Check metrics endpoint
curl -s http://<node-ip>:26660/metrics | head -20
```

### 2. Check Node Process

```bash
# SSH to node
ssh user@<node-ip>

# Check if process is running
systemctl status virtengined

# Check process resource usage
ps aux | grep virtengined
top -p $(pgrep virtengined)
```

### 3. Check System Resources

```bash
# Memory usage
free -h

# Disk usage (critical for state DB)
df -h /var/lib/virtengine

# CPU load
uptime

# Check for OOM kills
dmesg | grep -i "killed process" | tail -10
```

### 4. Check Logs

```bash
# Recent logs
journalctl -u virtengined -n 200 --no-pager

# Errors only
journalctl -u virtengined -p err -n 100 --no-pager

# Search for panic
journalctl -u virtengined | grep -i panic | tail -20
```

## Resolution Steps

### Scenario 1: Process Crashed

```bash
# Restart the service
sudo systemctl restart virtengined

# Wait for startup (30-60 seconds)
sleep 60

# Verify it's running
systemctl status virtengined
virtengined status | jq '.SyncInfo.catching_up'
```

### Scenario 2: Out of Memory

```bash
# Check current memory usage
free -h

# Check for memory leak patterns in logs
journalctl -u virtengined | grep -i "memory" | tail -20

# Increase memory limit if running in container
# Edit: /etc/systemd/system/virtengined.service
# Add: MemoryLimit=16G

# Restart with new limits
sudo systemctl daemon-reload
sudo systemctl restart virtengined
```

### Scenario 3: Disk Full

```bash
# Check disk usage
df -h

# Find large files
du -sh /var/lib/virtengine/*

# If WAL is large, compact state (with node stopped)
sudo systemctl stop virtengined
virtengined compact-state
sudo systemctl start virtengined

# If data directory is corrupted, restore from snapshot
# See: docs/operations/restore-from-snapshot.md
```

### Scenario 4: Network Issues

```bash
# Check peer connectivity
virtengined net info | jq '.n_peers'

# Check if seeds are reachable
nc -zv <seed-node-ip> 26656

# Check firewall rules
sudo iptables -L -n | grep 26656
sudo ss -tlnp | grep 26656

# Restart network stack if needed
sudo systemctl restart networking
```

### Scenario 5: State Database Corruption

```bash
# Stop the node
sudo systemctl stop virtengined

# Check database integrity
virtengined debug store-integrity /var/lib/virtengine/data

# If corruption detected, restore from snapshot
./restore-snapshot.sh

# Restart
sudo systemctl start virtengined
```

## Recovery Verification

After applying fixes:

```bash
# 1. Verify process is running
systemctl is-active virtengined

# 2. Check sync status
virtengined status | jq '.SyncInfo'

# 3. Verify block height is increasing
watch -n 5 'virtengined status | jq .SyncInfo.latest_block_height'

# 4. Check peer connections
virtengined net info | jq '.n_peers'

# 5. Verify metrics are being scraped
curl -s http://localhost:26660/metrics | grep virtengine_chain_block_height
```

## Escalation

Escalate to L2 if:
- Node fails to restart after 3 attempts
- State database is corrupted
- Multiple nodes are affected
- Issue recurs within 24 hours

Escalate to L3 if:
- More than 1/3 of validators are down
- Chain is halted
- Data loss is suspected

## Post-Incident

1. Create incident ticket
2. Capture logs for analysis
3. Schedule postmortem if:
   - Downtime > 5 minutes
   - Root cause unknown
   - Customer impact occurred

## Related Alerts

- `BlockProductionStalled` - May fire alongside if this is a validator
- `LowPeerCount` - May indicate network issues
- `ValidatorMissedBlocks` - Consequence of node downtime

## References

- [Node Operations Guide](../../node-operations.md)
- [State Restore Procedure](../../state-restore.md)
- [VirtEngine Status Codes](../../status-codes.md)
