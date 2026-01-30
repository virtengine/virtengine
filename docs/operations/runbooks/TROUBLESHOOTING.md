# Common Troubleshooting Scenarios

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Quick Diagnostic Commands](#quick-diagnostic-commands)
2. [Chain and Consensus Issues](#chain-and-consensus-issues)
3. [Validator Issues](#validator-issues)
4. [Provider Daemon Issues](#provider-daemon-issues)
5. [Identity Scoring Issues](#identity-scoring-issues)
6. [Network and Connectivity Issues](#network-and-connectivity-issues)
7. [Performance Issues](#performance-issues)
8. [Database Issues](#database-issues)
9. [Key and Authentication Issues](#key-and-authentication-issues)
10. [Log Analysis](#log-analysis)

---

## Quick Diagnostic Commands

### Chain Status

```bash
# Overall status
virtengine status | jq

# Block height and sync status
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Consensus state
curl -s http://localhost:26657/consensus_state | jq

# Latest block info
curl -s http://localhost:26657/block | jq '.result.block.header'
```

### Node Health

```bash
# Service status
sudo systemctl status virtengine

# Recent logs
journalctl -u virtengine -n 50 --no-pager

# Resource usage
ps aux | grep virtengine
free -h
df -h
```

### Network Status

```bash
# Peer count and info
curl -s http://localhost:26657/net_info | jq '.result | {n_peers, peers: [.peers[].node_info.moniker]}'

# Mempool status
curl -s http://localhost:26657/num_unconfirmed_txs | jq
```

---

## Chain and Consensus Issues

### Issue: Node Not Syncing

**Symptoms:**
- `catching_up: true` for extended period
- Block height not increasing
- "Falling behind" warnings in logs

**Diagnosis:**

```bash
# Check sync status
virtengine status | jq '.SyncInfo'

# Check peer count
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Check disk space
df -h ~/.virtengine/data

# Check network connectivity
ping -c 3 seed.virtengine.com
```

**Solutions:**

1. **Low peer count:**
   ```bash
   # Add more seeds
   sed -i 's/seeds = ""/seeds = "seed1@ip:26656,seed2@ip:26656"/' ~/.virtengine/config/config.toml
   sudo systemctl restart virtengine
   ```

2. **Disk space issue:**
   ```bash
   # Clean up old data
   virtengine prune all
   
   # Or expand disk
   ```

3. **Significantly behind (use state sync):**
   ```bash
   # Get trust height and hash
   TRUST_HEIGHT=$(curl -s https://rpc.virtengine.com/status | jq -r '.result.sync_info.latest_block_height')
   TRUST_HASH=$(curl -s "https://rpc.virtengine.com/block?height=$TRUST_HEIGHT" | jq -r '.result.block_id.hash')
   
   # Configure state sync in config.toml
   [statesync]
   enable = true
   rpc_servers = "rpc1.virtengine.com:26657,rpc2.virtengine.com:26657"
   trust_height = TRUST_HEIGHT
   trust_hash = "TRUST_HASH"
   
   # Reset and restart
   virtengine tendermint unsafe-reset-all
   sudo systemctl start virtengine
   ```

---

### Issue: Blocks Not Producing

**Symptoms:**
- Block height stuck
- Consensus timeouts in logs
- "PreCommit" or "Prevote" timeouts

**Diagnosis:**

```bash
# Check consensus state
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state'

# Check validator participation
curl -s http://localhost:26657/validators | jq '.result.validators | length'

# Check for panics
journalctl -u virtengine | grep -i panic
```

**Solutions:**

1. **Single validator issue:**
   - Check that validator is running
   - Check that validator key is correct
   - Restart affected validator

2. **Network partition:**
   - Check network connectivity between validators
   - Verify firewall rules allow P2P traffic

3. **Consensus bug:**
   - Check for common error patterns
   - Apply hotfix if available
   - Coordinate validator restart

---

### Issue: High Mempool Backlog

**Symptoms:**
- Transaction submission slow
- High `num_unconfirmed_txs`
- Users complaining about delays

**Diagnosis:**

```bash
# Check mempool size
curl -s http://localhost:26657/num_unconfirmed_txs | jq

# Check mempool transactions
curl -s http://localhost:26657/unconfirmed_txs?limit=10 | jq
```

**Solutions:**

1. **Normal congestion:**
   ```toml
   # Increase mempool size in config.toml
   [mempool]
   size = 10000
   max_txs_bytes = 2147483648
   ```

2. **Spam attack:**
   - Implement rate limiting
   - Increase minimum gas price
   - Contact security team

---

## Validator Issues

### Issue: Validator Missing Blocks

**Symptoms:**
- Increasing `missed_blocks_counter`
- Validator jailed after too many misses
- Warning alerts firing

**Diagnosis:**

```bash
# Check signing info
virtengine query slashing signing-info $(virtengine tendermint show-validator)

# Check if jailed
virtengine query staking validator $(virtengine keys show operator --bech val -a) | grep jailed

# Check validator logs
journalctl -u virtengine | grep -i "prevote\|precommit" | tail -20
```

**Solutions:**

1. **Performance issue:**
   - Upgrade hardware (more CPU, faster disk)
   - Optimize network latency
   - Use sentry nodes

2. **Network latency:**
   ```bash
   # Check latency to other validators
   for peer in $(curl -s http://localhost:26657/net_info | jq -r '.result.peers[].remote_ip'); do
     echo "$peer: $(ping -c 3 $peer | tail -1 | cut -d'/' -f5)ms"
   done
   ```

3. **If jailed:**
   ```bash
   # Wait for jail period to end, then unjail
   virtengine tx slashing unjail \
       --from operator \
       --chain-id virtengine-1 \
       --gas auto \
       --gas-adjustment 1.5
   ```

---

### Issue: Validator Jailed

**Symptoms:**
- Validator status shows jailed
- Not receiving block rewards
- Alert: `ValidatorJailed`

**Diagnosis:**

```bash
# Check jail reason
virtengine query slashing signing-info $(virtengine tendermint show-validator)

# Possible reasons:
# - missed_blocks_counter > threshold (downtime)
# - tombstoned = true (double sign)
```

**Solutions:**

1. **Jailed for downtime:**
   ```bash
   # Fix the underlying issue first
   # Then unjail
   virtengine tx slashing unjail --from operator
   ```

2. **Tombstoned (double sign):**
   - This is permanent and cannot be recovered
   - Must create new validator with new keys
   - **Prevention**: Never run two validators with same keys

---

### Issue: Validator Key Not Found

**Symptoms:**
- Node starts but not signing
- "validator key not found" in logs
- Can't create validator transaction

**Diagnosis:**

```bash
# Check if key file exists
ls -la ~/.virtengine/config/priv_validator_key.json

# Verify key matches on-chain
virtengine tendermint show-validator
virtengine query staking validator $(virtengine keys show operator --bech val -a) | grep consensus_pubkey
```

**Solutions:**

1. **Key file missing:**
   ```bash
   # Restore from backup
   gpg -d validator-key-backup.gpg > ~/.virtengine/config/priv_validator_key.json
   chmod 600 ~/.virtengine/config/priv_validator_key.json
   ```

2. **Key mismatch:**
   - Ensure using correct key file for this validator
   - Check if key was regenerated accidentally

---

## Provider Daemon Issues

### Issue: Provider Not Receiving Orders

**Symptoms:**
- No bids being submitted
- Provider shows active but no workloads
- Empty bid history

**Diagnosis:**

```bash
# Check provider status
virtengine query provider info $(virtengine keys show provider -a)

# Check bid engine
provider-daemon bids status

# Check order stream
provider-daemon debug orders --limit 10
```

**Solutions:**

1. **Provider paused:**
   ```bash
   virtengine tx provider set-status --status ACTIVE --from provider
   ```

2. **Bid engine disabled:**
   ```bash
   # Check config
   grep -A5 "bidding:" ~/.provider-daemon/config.yaml
   
   # Enable bidding
   sed -i 's/enabled: false/enabled: true/' ~/.provider-daemon/config.yaml
   sudo systemctl restart provider-daemon
   ```

3. **Chain connectivity issue:**
   ```bash
   # Check RPC connection
   curl -s http://localhost:26657/status
   
   # Verify config
   grep -A5 "chain:" ~/.provider-daemon/config.yaml
   ```

---

### Issue: Workloads Failing to Deploy

**Symptoms:**
- Orders accepted but workloads stuck in "Deploying"
- Kubernetes/SLURM errors in logs
- Customer complaints about failed deployments

**Diagnosis:**

```bash
# Check workload status
provider-daemon workloads list --state failed

# Check workload logs
provider-daemon workloads logs <workload-id>

# Check orchestrator status
kubectl get pods -n virtengine-workloads
```

**Solutions:**

1. **Kubernetes resource issues:**
   ```bash
   # Check node status
   kubectl get nodes
   kubectl describe node <node-name>
   
   # Check resource quotas
   kubectl describe quota -n virtengine-workloads
   ```

2. **Image pull failures:**
   ```bash
   # Check image pull secret
   kubectl get secret virtengine-registry -n virtengine-workloads
   
   # Test manual pull
   docker pull <image-name>
   ```

3. **SLURM submission failures:**
   ```bash
   # Check SLURM status
   sinfo
   squeue -a
   
   # Check SLURM logs
   cat /var/log/slurm/slurmd.log | tail -50
   ```

---

### Issue: Usage Submissions Failing

**Symptoms:**
- Pending usage reports piling up
- "transaction failed" errors in logs
- Revenue not being collected

**Diagnosis:**

```bash
# Check pending reports
provider-daemon usage pending

# Check submission errors
journalctl -u provider-daemon | grep -i "usage.*error" | tail -20

# Check provider balance
virtengine query bank balances $(virtengine keys show provider -a)
```

**Solutions:**

1. **Insufficient gas:**
   ```bash
   # Top up provider account
   virtengine tx bank send <source> $(virtengine keys show provider -a) 10000000uve
   ```

2. **Chain congestion:**
   ```bash
   # Increase gas price
   sed -i 's/gas_prices: .*/gas_prices: "0.05uve"/' ~/.provider-daemon/config.yaml
   sudo systemctl restart provider-daemon
   ```

3. **Manual submission:**
   ```bash
   provider-daemon usage submit --force
   ```

---

## Identity Scoring Issues

### Issue: Scoring Queue Backlog

**Symptoms:**
- High number of pending scores
- Scoring latency > 5 minutes
- Users waiting for identity verification

**Diagnosis:**

```bash
# Check queue depth
virtengine query veid pending-scores --count

# Check scoring metrics
curl -s http://localhost:26660/metrics | grep veid_scoring

# Check ML inference logs
journalctl -u ml-inference | tail -50
```

**Solutions:**

1. **Increase concurrency:**
   ```toml
   # In identity.toml
   max_concurrent_scoring = 8
   ```

2. **Scale inference resources:**
   - Add more CPUs
   - If using GPU, verify GPU is available

3. **Temporary rate limiting:**
   - Communicate to users about delays
   - Consider pausing new submissions temporarily

---

### Issue: Model Version Mismatch

**Symptoms:**
- Scoring failures
- "model hash mismatch" errors
- Validators disagreeing on scores

**Diagnosis:**

```bash
# Check model version across validators
for v in validator-{1..5}; do
  echo "$v: $(ssh $v 'sha256sum ~/.virtengine/models/veid_scorer_*.h5')"
done

# Check expected version
virtengine query veid model-status
```

**Solutions:**

```bash
# Download correct model
wget -O ~/.virtengine/models/veid_scorer_v1.0.0.h5 \
    https://models.virtengine.com/veid_scorer_v1.0.0.h5

# Verify hash
sha256sum ~/.virtengine/models/veid_scorer_v1.0.0.h5

# Restart validator
sudo systemctl restart virtengine
```

---

## Network and Connectivity Issues

### Issue: Low Peer Count

**Symptoms:**
- `n_peers` < 10
- Slow block propagation
- Intermittent sync issues

**Diagnosis:**

```bash
# Check current peers
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Check peer details
curl -s http://localhost:26657/net_info | jq '.result.peers[].node_info.moniker'

# Check for connection errors
journalctl -u virtengine | grep -i "dial\|connection" | tail -20
```

**Solutions:**

1. **Add more seeds/peers:**
   ```toml
   # config.toml
   seeds = "seed1@ip:26656,seed2@ip:26656"
   persistent_peers = "peer1@ip:26656,peer2@ip:26656"
   ```

2. **Firewall issues:**
   ```bash
   # Ensure P2P port is open
   sudo ufw allow 26656/tcp
   
   # Check external connectivity
   nc -zv <your-external-ip> 26656
   ```

3. **NAT traversal:**
   ```toml
   # config.toml
   external_address = "tcp://YOUR_PUBLIC_IP:26656"
   ```

---

### Issue: Connection Timeouts

**Symptoms:**
- Frequent peer disconnections
- RPC timeouts
- "context deadline exceeded" errors

**Diagnosis:**

```bash
# Check network latency
for peer in $(curl -s http://localhost:26657/net_info | jq -r '.result.peers[].remote_ip'); do
  ping -c 3 $peer
done

# Check for packet loss
mtr -r -c 10 seed.virtengine.com
```

**Solutions:**

1. **Increase timeouts:**
   ```toml
   # config.toml
   [p2p]
   dial_timeout = "10s"
   handshake_timeout = "30s"
   
   [consensus]
   timeout_propose = "5s"
   timeout_prevote = "2s"
   ```

2. **Network optimization:**
   - Use geographically closer peers
   - Upgrade network bandwidth
   - Check for ISP issues

---

## Performance Issues

### Issue: High Memory Usage

**Symptoms:**
- Memory usage > 80%
- OOM kills
- System slowdown

**Diagnosis:**

```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check virtengine memory
cat /proc/$(pidof virtengine)/status | grep -E "VmRSS|VmSize"

# Go memory stats (if pprof enabled)
curl http://localhost:6060/debug/pprof/heap > heap.pprof
```

**Solutions:**

1. **Adjust Go GC:**
   ```bash
   # In systemd service
   Environment="GOGC=50"
   ```

2. **Reduce state cache:**
   ```toml
   # app.toml
   [state-cache]
   cache-size = 100000  # Reduce from default
   ```

3. **Add more RAM** if consistently high

---

### Issue: High CPU Usage

**Symptoms:**
- CPU consistently > 80%
- Slow block processing
- Lag in responding to queries

**Diagnosis:**

```bash
# Check CPU usage
top -bn1 | grep virtengine
mpstat -P ALL 1 5

# Profile CPU (if pprof enabled)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.pprof
```

**Solutions:**

1. **Identify hotspots:**
   ```bash
   go tool pprof cpu.pprof
   # Use 'top' command to see hotspots
   ```

2. **Reduce concurrent operations:**
   ```toml
   # identity.toml
   max_concurrent_scoring = 2  # Reduce
   ```

3. **Upgrade CPU** if needed

---

### Issue: Slow Disk I/O

**Symptoms:**
- High disk wait times
- Slow sync speed
- Database query timeouts

**Diagnosis:**

```bash
# Check disk I/O
iostat -x 1 5

# Check disk latency
ioping -c 10 ~/.virtengine/data

# Check disk space
df -h
```

**Solutions:**

1. **Use faster storage:**
   - Upgrade to NVMe SSD
   - Use RAID configuration

2. **Optimize pruning:**
   ```toml
   # app.toml
   pruning = "custom"
   pruning-keep-recent = "100"
   pruning-keep-every = "0"
   pruning-interval = "10"
   ```

3. **Move data directory:**
   ```bash
   # Move to faster disk
   sudo systemctl stop virtengine
   mv ~/.virtengine/data /fast-disk/virtengine-data
   ln -s /fast-disk/virtengine-data ~/.virtengine/data
   sudo systemctl start virtengine
   ```

---

## Database Issues

### Issue: PostgreSQL Connection Pool Exhausted

**Symptoms:**
- "too many connections" errors
- Application timeouts
- Connection refused

**Diagnosis:**

```bash
# Check connection count
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check connections by state
psql -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# Find long-running queries
psql -c "SELECT pid, now() - query_start AS duration, query 
         FROM pg_stat_activity 
         WHERE state = 'active' 
         ORDER BY duration DESC LIMIT 10;"
```

**Solutions:**

1. **Kill idle connections:**
   ```sql
   SELECT pg_terminate_backend(pid) 
   FROM pg_stat_activity 
   WHERE state = 'idle' 
   AND query_start < now() - interval '30 minutes';
   ```

2. **Increase max connections:**
   ```ini
   # postgresql.conf
   max_connections = 500
   ```

3. **Use connection pooler (PgBouncer):**
   ```ini
   # pgbouncer.ini
   [databases]
   virtengine = host=localhost port=5432 dbname=virtengine_db
   
   [pgbouncer]
   pool_mode = transaction
   max_client_conn = 1000
   default_pool_size = 50
   ```

---

## Key and Authentication Issues

### Issue: Keyring Locked

**Symptoms:**
- "keyring is locked" errors
- Can't sign transactions
- Service won't start

**Diagnosis:**

```bash
# Check keyring backend
virtengine config keyring-backend

# Test key access
virtengine keys list --keyring-backend file
```

**Solutions:**

1. **Unlock keyring:**
   ```bash
   # Interactive unlock
   virtengine keys list --keyring-backend file
   # Enter password when prompted
   ```

2. **Use environment variable:**
   ```bash
   # In systemd service
   Environment="VIRTENGINE_KEYRING_PASSPHRASE=your-passphrase"
   ```

3. **Reset keyring:**
   ```bash
   # If password lost, restore from backup
   rm -rf ~/.virtengine/keyring-file
   # Restore from backup
   ```

---

### Issue: Transaction Signing Failed

**Symptoms:**
- "signature verification failed" errors
- Transactions rejected
- Key mismatch errors

**Diagnosis:**

```bash
# Check account sequence
virtengine query account $(virtengine keys show operator -a)

# Check nonce
virtengine query auth account $(virtengine keys show operator -a)
```

**Solutions:**

1. **Sequence mismatch:**
   ```bash
   # Use auto sequence
   virtengine tx bank send ... --sequence auto
   ```

2. **Wrong chain ID:**
   ```bash
   # Verify chain ID
   virtengine query node-info | jq '.default_node_info.network'
   ```

---

## Log Analysis

### Common Error Patterns

```bash
# Consensus issues
journalctl -u virtengine | grep -E "CONSENSUS|prevote|precommit" | tail -50

# P2P issues
journalctl -u virtengine | grep -E "dial|peer|connection" | tail -50

# State sync issues
journalctl -u virtengine | grep -i "statesync\|snapshot" | tail -50

# ABCI errors
journalctl -u virtengine | grep -i "ABCI\|app_hash" | tail -50

# Panic/crash
journalctl -u virtengine | grep -E "panic|fatal|SIGSEGV" | tail -50
```

### Log Level Adjustment

```toml
# config.toml - for debugging
log_level = "debug"
log_format = "json"

# For specific modules
log_level = "consensus:debug,state:info,*:warn"
```

### Structured Log Queries (Loki)

```logql
# All errors
{job="virtengine-node"} |= "error" | json

# Consensus issues
{job="virtengine-node"} |= "consensus" |= "timeout" | json

# Specific module errors
{job="virtengine-node"} | json | module="veid" | level="error"

# Panic events
{job="virtengine-node"} |~ "panic|fatal"
```

---

## Appendix: Diagnostic Scripts

### Full Diagnostic Script

```bash
#!/bin/bash
# full-diagnostic.sh

echo "=== VirtEngine Full Diagnostic ==="
echo "Timestamp: $(date)"
echo ""

echo "=== System Info ==="
uname -a
echo ""

echo "=== Service Status ==="
systemctl status virtengine --no-pager
echo ""

echo "=== Node Status ==="
virtengine status 2>&1 | jq
echo ""

echo "=== Sync Info ==="
curl -s http://localhost:26657/status | jq '.result.sync_info'
echo ""

echo "=== Peer Info ==="
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
echo ""

echo "=== Mempool ==="
curl -s http://localhost:26657/num_unconfirmed_txs | jq
echo ""

echo "=== Resource Usage ==="
free -h
df -h ~/.virtengine
ps aux --sort=-%mem | grep virtengine | head -5
echo ""

echo "=== Recent Errors ==="
journalctl -u virtengine --since "30 minutes ago" | grep -i error | tail -20
echo ""

echo "=== Validator Info ==="
virtengine query slashing signing-info $(virtengine tendermint show-validator) 2>/dev/null || echo "Not a validator"
echo ""

echo "=== End Diagnostic ==="
```

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
