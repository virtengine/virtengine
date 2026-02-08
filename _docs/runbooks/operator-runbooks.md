# Operator Runbooks (Production)

**Version:** 1.0.0  
**Last updated:** 2026-02-08  
**Owner:** SRE + Protocol Ops

## Purpose
This document provides production-ready operational runbooks for VirtEngine operators. It focuses on day-2 operations, incident response, and recovery workflows for validators and provider infrastructure.

## Scope
- Validators (sentry + validator topology)
- Provider daemon and orchestration adapters
- Core chain services (RPC, gRPC, API)
- Monitoring, alerting, and incident response

## References
- Validator onboarding: `_docs/runbooks/validator-onboarding.md`
- Provider guide: `_docs/provider-guide.md`
- Monitoring: `_docs/operations/monitoring.md`
- Incident response lifecycle: `_docs/operations/incident-response.md`
- SLOs and playbooks: `_docs/slos-and-playbooks.md`
- Disaster recovery plan: `_docs/disaster-recovery.md`
- Upgrade procedures: `docs/operations/runbooks/UPGRADE_PROCEDURES.md`

---

## 1) Validator Setup and Maintenance Guide

### 1.1 Topology (Production)
- **Validator node**: private, no public RPC.
- **Sentry nodes (2+)**: public P2P + limited RPC for ops.
- **Monitoring host**: centralized metrics + logs.

### 1.2 Baseline System Setup
```bash
sudo apt update
sudo apt install -y jq curl ufw chrony
```

#### Time sync
```bash
timedatectl status
chronyc tracking
```

#### Firewall (sentry)
```bash
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to ops IPs)
sudo ufw allow 9100/tcp   # Node exporter (internal)
sudo ufw enable
```

### 1.3 Install and Configure VirtEngine
```bash
curl -L https://releases.virtengine.com/virtengine/v1.0.0/virtengine-linux-amd64 -o virtengine
chmod +x virtengine
sudo mv virtengine /usr/local/bin/virtengine
virtengine version
```

```bash
virtengine init <moniker> --chain-id virtengine-1
```

**Configuration checklist**
- `~/.virtengine/config/config.toml`
  - `external_address`, `persistent_peers` set to sentries
  - `addr_book_strict = true`
  - `pex = true`
- `~/.virtengine/config/app.toml`
  - `minimum-gas-prices = "0.025uve"`
  - disable unsafe APIs (avoid RPC on validator)

### 1.4 Key Management
- Store `priv_validator_key.json` on an offline signing host or HSM.
- Never expose validator signing keys on internet-facing hosts.
- Keep operator keys separate from consensus keys.

### 1.5 Run as systemd
```bash
sudo tee /etc/systemd/system/virtengine.service > /dev/null <<'UNIT'
[Unit]
Description=VirtEngine Node
After=network-online.target

[Service]
User=virtengine
ExecStart=/usr/local/bin/virtengine start --home /home/virtengine/.virtengine
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable --now virtengine
```

### 1.6 Routine Maintenance

**Daily (15 min)**
- Check block height advancing and peer count.
- Review missed blocks/jailing risk.
- Confirm disk usage < 80%.

**Weekly (30 min)**
- Patch OS packages during maintenance window.
- Verify backups completed.
- Prune old logs and check snapshot rotation.

**Monthly (60 min)**
- Restore from backup in staging.
- Rotate non-consensus keys (operator/API keys).
- Validate alerting and pager escalation paths.

### 1.7 Jailed Validator Recovery
```bash
virtengine status | jq -r '.SyncInfo.latest_block_height'
virtengine query slashing signing-info <validator-consensus-address>
```

If jailed due to downtime and safe to unjail:
```bash
virtengine tx slashing unjail --from validator-operator --keyring-backend file
```

### 1.8 Validator Maintenance Checklist (Pre-Upgrade)
- [ ] Sentry health OK, validator peers stable
- [ ] Snapshot/backup completed
- [ ] New binary verified (checksum)
- [ ] Rollback binary ready
- [ ] Comms channel active

---

## 2) Provider Daemon Operational Procedures

### 2.1 Service Health Checks
```bash
# Process
systemctl status provider-daemon

# Metrics endpoint
curl -sf http://localhost:9091/metrics | head
```

### 2.2 Log and Error Review
```bash
journalctl -u provider-daemon -n 200 --no-pager
```

**Common log signals**
- `bid_rejected`: check pricing thresholds
- `usage_submit_failed`: verify chain RPC and keyring
- `workload_health_failed`: check orchestration adapter

### 2.3 Configuration Drift
- Store `~/.provider-daemon/config.yaml` in a secured config repo.
- Use immutable config changes (update, restart, verify).

### 2.4 Key Rotation
- Rotate encryption key quarterly.
- Register new recipient key on-chain before switching.

```bash
provider-daemon keys rotate provider-encryption
virtengine tx encryption register-recipient-key \
  --algorithm X25519-XSalsa20-Poly1305 \
  --public-key $(provider-daemon keys show provider-encryption --pubkey) \
  --label "Provider Order Decryption Key" \
  --from provider
```

### 2.5 Usage Reporting
- Ensure reports submitted hourly (default). 
- Monitor backlog size and retry counts.

### 2.6 Rolling Restart
```bash
sudo systemctl restart provider-daemon
sleep 5
systemctl is-active provider-daemon
```

### 2.7 Scaling and Failover
- Active/standby per region.
- Stateful data (keyring, cache) replicated to standby.
- Ensure standby uses a distinct node ID and metrics labels.

---

## 3) Incident Response Playbooks

### 3.1 Severity Definitions
- **P0**: Chain halted, funds at risk, major security incident.
- **P1**: Partial outage, validator set instability, major service degradation.
- **P2**: Degraded performance, localized outages.
- **P3**: Minor issue, no impact to users.

### 3.2 Standard Incident Flow
1. Triage and declare incident
2. Mitigate blast radius
3. Restore service
4. Post-incident review

### 3.3 Playbook: Chain Halt
**Signals**: No new blocks, consensus timeout alerts.

**Immediate Actions**
1. Verify local node height vs reference RPC.
2. Check validator logs for panic or upgrade halt.
3. Confirm governance upgrade height (if planned).

**Mitigation**
- If upgrade planned, proceed with upgrade steps.
- If unplanned, coordinate with core team to identify failing validators.

**Verification**
- Block height advances on 2+ independent nodes.
- No consensus errors in logs.

### 3.4 Playbook: Validator Missing Blocks
**Signals**: Missed block alerts, slashing risk.

**Immediate Actions**
1. Check time sync (chrony).
2. Validate peer count and p2p connectivity.
3. Check disk IO and CPU saturation.

**Mitigation**
- Restart validator (if safe).
- Increase sentry capacity or replace failing sentry.

### 3.5 Playbook: RPC/GRPC Outage
**Signals**: API gateway 5xx, RPC probe failures.

**Immediate Actions**
1. Confirm node is healthy (blocks advancing).
2. Restart RPC service or fullnode.
3. Fail over to backup RPC endpoint.

### 3.6 Playbook: Provider Daemon Down
**Signals**: Provider health checks failing, no bids.

**Immediate Actions**
1. Restart daemon.
2. Verify chain connectivity.
3. Validate keyring access permissions.

**Mitigation**
- Fail over to standby daemon.
- Suspend bids if infrastructure unstable.

### 3.7 Playbook: Key Compromise
**Signals**: Unauthorized signatures, unexpected transfers.

**Immediate Actions**
1. Revoke compromised keys.
2. Rotate encryption and operator keys.
3. Notify security and follow incident process.

**Recovery**
- Migrate operations to new keys.
- Update allowlists and monitoring.

---

## 4) Disaster Recovery Procedures

### 4.1 DR Categories
Follow `_docs/disaster-recovery.md` for RTO/RPO and region strategy.

### 4.2 Validator Restore (From Snapshot)
1. Provision fresh host.
2. Restore `config/` and `data/` from latest snapshot.
3. Verify checksums match snapshot manifest.
4. Start node and compare height to reference RPC.

### 4.3 Provider Daemon Restore
1. Restore `~/.provider-daemon/` (config, keyring). 
2. Verify encryption keys registered on-chain.
3. Start service and validate bids + health checks.

### 4.4 Regional Failover
- Promote secondary region RPC/GRPC endpoints.
- Shift provider bids to active region.
- Validate monitoring and alerting continuity.

---

## 5) Backup and Restore Procedures

### 5.1 What to Back Up
- `~/.virtengine/config/`
- `~/.virtengine/data/` (or state sync snapshots)
- `~/.virtengine/cosmovisor/` binaries
- `~/.provider-daemon/` (config + keyring)
- Monitoring rules and dashboards

### 5.2 Backup Schedule
- **State snapshots**: every 1,000 blocks
- **Config**: hourly or on change
- **Key material**: after rotation, encrypted and offline

### 5.3 Sample Backup (Node)
```bash
SNAPSHOT_DIR=/data/snapshots
NODE_HOME=$HOME/.virtengine
TIMESTAMP=$(date -u +%Y%m%d_%H%M%S)

mkdir -p "$SNAPSHOT_DIR"

# Config
tar -czf "$SNAPSHOT_DIR/config_${TIMESTAMP}.tar.gz" -C "$NODE_HOME" config

# Data (exclude logs)
tar -czf "$SNAPSHOT_DIR/data_${TIMESTAMP}.tar.gz" -C "$NODE_HOME" \
  --exclude='*.log' data

# Checksums
cd "$SNAPSHOT_DIR"
sha256sum config_${TIMESTAMP}.tar.gz data_${TIMESTAMP}.tar.gz > checksums_${TIMESTAMP}.sha256
```

### 5.4 Restore Validation
- Verify checksums before restore.
- Validate block height within expected delta.
- Confirm node ID and validator keys are correct.

---

## 6) Upgrade Procedures

Use `docs/operations/runbooks/UPGRADE_PROCEDURES.md` for detailed steps.

### 6.1 Chain Upgrade (Cosmovisor)
```bash
export DAEMON_NAME=virtengine
export DAEMON_HOME=$HOME/.virtengine

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/<upgrade-name>/bin
cp /usr/local/bin/virtengine-$UPGRADE_VERSION \
  $DAEMON_HOME/cosmovisor/upgrades/<upgrade-name>/bin/virtengine
```

### 6.2 Emergency Hotfix
- Coordinate with core team before rollout.
- Patch sentries first, then validator.
- Monitor missed blocks and consensus health.

### 6.3 Provider Daemon Upgrade
- Rolling restart per region.
- Validate bids and usage reporting post-upgrade.

---

## 7) Common Troubleshooting Scenarios

### 7.1 Node Stuck or Not Syncing
- Verify peers and `persistent_peers`.
- Check disk IO and file descriptor limits.
- Restart node and monitor logs.

### 7.2 High Missed Blocks
- Confirm NTP sync.
- Reduce CPU contention.
- Ensure sentries are healthy.

### 7.3 RPC 5xx Errors
- Check RPC fullnode health and memory usage.
- Rotate RPC endpoint if needed.

### 7.4 Provider Daemon Not Bidding
- Ensure bidding enabled in config.
- Verify wallet balance and gas prices.
- Check chain connectivity.

### 7.5 Usage Report Submission Failures
- Inspect retry counts.
- Verify keyring access and chain RPC.
- Check mempool congestion.

---

## 8) Performance Tuning Guide

### 8.1 Validator / Full Node
- Increase `LimitNOFILE` (65k+).
- Use NVMe storage and disable swap.
- Consider pruning and state sync snapshots.

**Config tuning**
- `p2p.max_num_inbound_peers`: 40-60
- `p2p.max_num_outbound_peers`: 15-25
- `mempool.size`: adjust to workload

### 8.2 Provider Daemon
- Increase `usage.batch_size` for high volume.
- Tune `workloads.max_concurrent` based on cluster capacity.
- Enable metrics and set alert thresholds.

### 8.3 Network and OS
```bash
# Example sysctl tuning
sudo sysctl -w net.core.somaxconn=1024
sudo sysctl -w net.ipv4.tcp_fin_timeout=15
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
```

---

## Change Log
- 2026-02-08: Initial operator runbooks for validators, providers, incident response, DR, backups, upgrades, troubleshooting, and tuning.