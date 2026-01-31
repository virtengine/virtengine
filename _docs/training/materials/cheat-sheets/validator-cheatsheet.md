# Validator Operator Cheat Sheet

**Quick Reference for VirtEngine Validators**

---

## Essential Commands

### Node Status
```bash
# Check sync status
virtengine status | jq '.SyncInfo'

# Check node info
virtengine status | jq '.NodeInfo'

# Check validator info
virtengine status | jq '.ValidatorInfo'

# Check connected peers
virtengine status | jq '.NodeInfo.listen_addr'
curl -s localhost:26657/net_info | jq '.result.n_peers'
```

### Key Management
```bash
# List keys
virtengine keys list --keyring-backend file

# Show key address
virtengine keys show validator -a --keyring-backend file

# Export key (backup)
virtengine keys export validator --keyring-backend file > validator.armor

# Import key (restore)
virtengine keys import validator validator.armor --keyring-backend file
```

### Validator Operations
```bash
# Query validator
virtengine query staking validator $(virtengine keys show validator --bech val -a)

# Check signing info
virtengine query slashing signing-info $(virtengine tendermint show-validator)

# Unjail validator
virtengine tx slashing unjail --from validator --keyring-backend file
```

---

## Critical Paths

### Configuration Files
```
~/.virtengine/
├── config/
│   ├── config.toml      # Node configuration
│   ├── app.toml         # Application configuration
│   ├── genesis.json     # Network genesis
│   ├── priv_validator_key.json  # ⚠️ CRITICAL - Validator key
│   └── node_key.json    # Node identity key
└── data/                # Blockchain data
```

### Key Ports
| Port | Service | Access |
|------|---------|--------|
| 26656 | P2P | Public |
| 26657 | RPC | Localhost |
| 26660 | Prometheus | Monitoring |
| 1317 | REST API | Localhost |
| 9090 | gRPC | Localhost |

---

## Monitoring

### Key Metrics
```promql
# Block height
tendermint_consensus_height

# Missed blocks
tendermint_consensus_validators_power{status="missing"}

# Peer count
tendermint_p2p_peers

# Memory usage
process_resident_memory_bytes
```

### Health Check
```bash
# Quick health check
curl -s localhost:26657/health | jq .

# Detailed status
curl -s localhost:26657/status | jq .
```

---

## Emergency Procedures

### Node Not Syncing
```bash
# Check peers
virtengine status | jq '.SyncInfo.catching_up'

# Add peers manually
# Edit config.toml: persistent_peers = "..."

# Restart
sudo systemctl restart virtengine
```

### Validator Jailed
```bash
# Check jail status
virtengine query staking validator $(virtengine keys show validator --bech val -a) | jq '.jailed'

# Wait for downtime period (varies)
# Then unjail
virtengine tx slashing unjail --from validator --keyring-backend file
```

### Key Compromise Suspected
```bash
# 1. STOP IMMEDIATELY
sudo systemctl stop virtengine

# 2. Secure the key file
chmod 000 ~/.virtengine/config/priv_validator_key.json

# 3. Contact security team
# 4. Do NOT restart until investigation complete
```

---

## Common Issues

| Symptom | Likely Cause | Quick Fix |
|---------|--------------|-----------|
| Not syncing | No peers | Add persistent_peers |
| Missed blocks | Clock drift | Sync NTP |
| Out of memory | Memory leak | Restart, check logs |
| Disk full | Old data | Prune, add storage |
| Jailed | Downtime/double sign | Unjail (if downtime) |

---

## Contact

- **Emergency**: security@virtengine.com
- **Support**: #validators channel
- **Docs**: _docs/validator-onboarding.md

---

*Keep this handy during on-call shifts!*
