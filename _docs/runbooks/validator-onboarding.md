# Validator Onboarding Runbook (Mainnet)

**Version:** 1.0.0  
**Last updated:** 2026-02-06  
**Owner:** Ops + Validator Relations

## Purpose
This runbook provides end-to-end steps for onboarding mainnet validators,
including hardware validation, key management, genesis participation, and
post-launch monitoring.

## References
- Hardware requirements: `_docs/validators/hardware-requirements.md`
- Genesis ceremony tooling: `scripts/mainnet/genesis-ceremony.sh`
- Genesis guide: `_docs/mainnet-genesis.md`
- Launch readiness checklist: `_docs/operations/mainnet-launch-readiness-checklist.md`

## 1) Pre-onboarding Checklist
- [ ] Hardware meets the Recommended profile
- [ ] Dedicated validator + sentry topology planned
- [ ] Time sync (NTP) verified
- [ ] Secure key storage plan (HSM or offline signing host)
- [ ] Ops contact + 24/7 pager escalation path provided
- [ ] Discord/Slack validator coordination channel access confirmed

## 2) Base System Setup

### OS + Packages
```bash
sudo apt update
sudo apt install -y jq curl ufw chrony
```

### Time Sync
```bash
timedatectl status
chronyc tracking
```

### Firewall (sentry nodes)
```bash
sudo ufw allow 26656/tcp  # P2P
sudo ufw allow 26657/tcp  # RPC (restrict to ops IPs)
sudo ufw allow 9100/tcp   # Node exporter (internal)
sudo ufw enable
```

## 3) Install VirtEngine Binary

```bash
curl -L https://releases.virtengine.com/virtengine/v1.0.0/virtengine-linux-amd64 -o virtengine
chmod +x virtengine
sudo mv virtengine /usr/local/bin/virtengine
virtengine version
```

## 4) Key Management

### Consensus Key (HSM recommended)
- Generate consensus keys in a secured environment.
- Never store `priv_validator_key.json` in plain text on internet-facing hosts.
- Use signing service or HSM integration where possible.

### Operator Key
```bash
virtengine keys add validator-operator --keyring-backend file
```

Record the operator address and secure backup.

## 5) Configure the Node

```bash
virtengine init <moniker> --chain-id virtengine-1
```

Edit configs:
- `~/.virtengine/config/config.toml`
  - `external_address` and `persistent_peers`
  - `addr_book_strict = true`
- `~/.virtengine/config/app.toml`
  - `minimum-gas-prices = "0.025uve"`
  - disable unsafe API endpoints on validator nodes

## 6) Genesis Participation

### Create a gentx
```bash
virtengine genesis gentx validator-operator 1000000000uve \
  --chain-id virtengine-1 \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1000000000 \
  --moniker "<Your Moniker>" \
  --identity "<keybase-id>" \
  --website "https://validator.example.com" \
  --details "VirtEngine mainnet validator"
```

### Submit gentx
- Send the resulting gentx file to the genesis coordinator.
- Include operator address + contact info.

### Verify inclusion
The coordinator will publish a final `genesis.json` and SHA-256 hash.
Verify:
```bash
sha256sum genesis.json
```

## 7) Start Validator

```bash
virtengine start \
  --home ~/.virtengine \
  --p2p.persistent_peers "<peer_list>"
```

Check status:
```bash
virtengine status | jq
```

## 8) Post-Launch Monitoring

### Required Alerts
- Block production halted
- Missed blocks > 5% in 1 hour
- Disk usage > 80%
- Memory usage > 90%
- P2P peers < 5 for > 10 min

### Dashboard
- Prometheus + Grafana with exporter on each node
- Logs forwarded to central SIEM

## 9) Upgrade + Incident Response
- Follow `_docs/runbooks/mainnet-launch-runbook.md` for coordinated upgrades.
- Escalate incidents via validator coordination channel.

## 10) Validator Exit (If Needed)
- Coordinate with release team before jailing/unbonding.
- Ensure slashing risks are mitigated.