# Validator Setup and Operations Guide

**Owner:** Validator Relations + SRE
**Last Updated:** 2026-04-11

This guide is the operator-facing validator procedure for staging, rehearsal,
and mainnet launch preparation. It intentionally fails closed on missing network
bundle data instead of inventing public launch values.

## Before You Start

Public mainnet start-up is allowed only when all of the following exist:

- the final published genesis bundle,
- the checksum files for that bundle,
- an approved release tag,
- an explicit `GO` decision in the launch package.

If any of those are missing, operators may prepare hosts and rehearse the flow,
but they must not join a public mainnet.

## Hardware and OS Baseline

| Component | Minimum | Recommended |
| --- | --- | --- |
| CPU | `8` cores | `16+` cores |
| RAM | `32 GiB` | `64 GiB` |
| Storage | `1 TiB` NVMe | `2 TiB` mirrored NVMe |
| Network | `100 Mbps` | `1 Gbps` dedicated |
| OS | Ubuntu `22.04` or equivalent | Ubuntu `22.04` LTS |

Install base packages:

```bash
sudo apt update
sudo apt install -y build-essential curl git jq lz4 ufw chrony
timedatectl status
chronyc tracking
```

## Source of Truth for Network Values

Read canonical launch-prep values from the repository:

```bash
CHAIN_ID="$(jq -r '.chain_id' config/mainnet/genesis-params.json)"
GENESIS_TIME="$(jq -r '.genesis_time' config/mainnet/genesis-params.json)"
echo "$CHAIN_ID $GENESIS_TIME"
```

For a public launch, also require the published artifact bundle:

- `${MAINNET_BUNDLE_DIR}/genesis.json`
- `${MAINNET_BUNDLE_DIR}/genesis.sha256`
- `${MAINNET_BUNDLE_DIR}/gentx.sha256`
- `${MAINNET_BUNDLE_DIR}/ceremony-manifest.json`
- `${MAINNET_BUNDLE_DIR}/ceremony-manifest.sha256`

Set `MAINNET_BUNDLE_DIR` to the downloaded publication bundle location before
continuing. The repository does not check in final publication artifacts before
launch approval.

```bash
MAINNET_BUNDLE_DIR="${MAINNET_BUNDLE_DIR:?download the published mainnet artifact bundle first}"
test -f "${MAINNET_BUNDLE_DIR}/genesis.json"
test -f "${MAINNET_BUNDLE_DIR}/genesis.sha256"
test -f "${MAINNET_BUNDLE_DIR}/gentx.sha256"
test -f "${MAINNET_BUNDLE_DIR}/ceremony-manifest.json"
test -f "${MAINNET_BUNDLE_DIR}/ceremony-manifest.sha256"
```

If any of those files are missing, stop the validator-join flow.

## Install the Binary

Use either a reviewed release tag or a source build from the exact commit under
validation.

### From source

```bash
make virtengine
sudo install -m 0755 .cache/bin/virtengine /usr/local/bin/virtengine
virtengine version
```

### From an approved release tag

```bash
RELEASE_TAG="${RELEASE_TAG:?set RELEASE_TAG to the approved release tag}"
curl -L "https://github.com/virtengine/virtengine/releases/download/${RELEASE_TAG}/virtengine_linux_amd64.tar.gz" -o /tmp/virtengine.tar.gz
tar -xzf /tmp/virtengine.tar.gz -C /tmp
sudo install -m 0755 /tmp/virtengine /usr/local/bin/virtengine
virtengine version
```

## Initialize the Home Directory

```bash
MONIKER="${MONIKER:?set validator moniker}"
virtengine genesis init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME/.virtengine" --overwrite
```

## Firewall and Network Layout

Use sentry topology for production validators.

```bash
OPS_CIDR="${OPS_CIDR:?set operator management CIDR}"
MONITORING_CIDR="${MONITORING_CIDR:?set monitoring CIDR}"
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 26656/tcp
sudo ufw allow from "$OPS_CIDR" to any port 26657 proto tcp
sudo ufw allow from "$MONITORING_CIDR" to any port 26660 proto tcp
sudo ufw enable
```

## Key Creation and Custody

```bash
virtengine keys add validator-operator --keyring-backend file --home "$HOME/.virtengine"
virtengine tendermint show-validator
```

Back up:

- operator mnemonic or hardware-backed recovery material,
- `priv_validator_key.json`,
- `priv_validator_state.json`,
- keyring credentials,
- off-host encrypted backup evidence.

Use `scripts/dr/backup-keys.sh --type all` after the node is prepared.

## Genesis and Gentx Flow

### Launch-prep or ceremony mode

Use the mainnet coordinator flow and validate locally first:

```bash
scripts/mainnet/validate-gentx.sh \
  --gentx-dir ./gentx \
  --constraints ./config/mainnet/gentx-constraints.json
```

Generate a gentx only after the self-delegation account is funded in the
approved allocations.

### Published bundle mode

Verify the published bundle before first start:

```bash
cd "$MAINNET_BUNDLE_DIR"
sha256sum -c genesis.sha256
sha256sum -c ceremony-manifest.sha256
cp "${MAINNET_BUNDLE_DIR}/genesis.json" "$HOME/.virtengine/config/genesis.json"
```

If any checksum fails, abort startup.

## Recommended Config

Minimum validator settings:

```toml
# ~/.virtengine/config/config.toml
moniker = "validator"
db_backend = "goleveldb"
prometheus = true
prometheus_listen_addr = ":26660"

[rpc]
laddr = "tcp://127.0.0.1:26657"

[p2p]
laddr = "tcp://0.0.0.0:26656"
pex = false
```

```toml
# ~/.virtengine/config/app.toml
minimum-gas-prices = "0.025uve"
pruning = "custom"
pruning-keep-recent = "100"
pruning-interval = "10"
```

## Systemd Service

```ini
[Unit]
Description=VirtEngine Validator
After=network.target

[Service]
User=validator
ExecStart=/usr/local/bin/virtengine start --home /home/validator/.virtengine
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
Environment="HOME=/home/validator"

[Install]
WantedBy=multi-user.target
```

## Validator Monitoring and Response

This section is the primary runbook for:

- `ValidatorDown`
- `ValidatorMissingBlocks`
- `ValidatorMissedBlocks`
- `ValidatorLowUptime`
- `MissedBlocksHigh`
- `SLONodeUptimeBudgetBurning`
- `SLONodeUptimeBudgetCritical`
- `NodeDown`
- `LowPeerCount`
- `NoPeers`
- `NodeOutOfSync`
- `NodeBehind`

Daily checks:

```bash
virtengine status | jq '.SyncInfo'
virtengine query slashing signing-info "$(virtengine tendermint show-validator)"
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
df -h "$HOME/.virtengine"
journalctl -u virtengine --since "24 hours ago" | tail -50
```

Response rules:

1. If the validator is jailed, fix the underlying cause before unjailing.
2. If peer count is low, repair connectivity before changing validator keys or
   state.
3. If the node is out of sync after a restart, use a verified state-sync or
   backup restore path.
4. If more than one third of validators are affected, escalate to consensus
   incident handling instead of working node-by-node.

## VEID-Participating Validators

Use this section when these alerts fire:

- `VEIDModelVersionMismatch`
- `VEIDModelNotLoaded`
- `VEIDNonDeterministicInference`
- `VEIDInferenceNonDeterministic`
- `VEIDQueueBacklog`
- `VEIDQueueCritical`

Required checks:

```bash
virtengine query veid model-status
curl -s http://localhost:26660/metrics | grep -E "veid_|inference_"
```

Rules:

1. Do not run a model bundle whose hash is not in the release manifest.
2. Do not continue serving verification traffic if the local manifest check
   fails.
3. Treat non-deterministic inference as a consensus-safety incident.

## Backup and Recovery Readiness

Minimum validator recovery commands:

```bash
./scripts/dr/backup-chain-state.sh
./scripts/dr/backup-keys.sh --type all
./scripts/dr/dr-test.sh --environment staging --report
```

If those commands cannot succeed for the validator environment, the validator is
not operationally ready.

## Do Not Do

- Do not start a public validator from a bundle whose hashes were not verified.
- Do not run two validators with the same signing key.
- Do not bypass the launch `GO` gate for public mainnet participation.
- Do not rotate validator keys during an unresolved consensus or attestation
  incident unless incident command approves it.
