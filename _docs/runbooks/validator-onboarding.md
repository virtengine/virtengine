# Validator Onboarding Runbook (Mainnet)

**Version:** 2.0.0
**Last updated:** 2026-04-10
**Owner:** Ops + Validator Relations

## Purpose
Provide the exact onboarding flow for mainnet validators, including hardware and
security requirements, gentx creation, coordinator submission, bundle
verification, and first start.

## References
- Hardware requirements: `_docs/validators/hardware-requirements.md`
- Genesis ceremony runbook: `_docs/runbooks/mainnet-genesis-ceremony.md`
- Genesis overview: `_docs/mainnet-genesis.md`
- Gentx policy: `config/mainnet/gentx-constraints.json`

## 1. Required Submission Package
Every validator must provide all of the following to the genesis coordinator:
- the signed gentx JSON file,
- the funding account address (`ve...`) used for self-delegation,
- the validator operator address (`vevaloper...`),
- the public P2P endpoint that appears in the gentx memo,
- a non-placeholder HTTPS website,
- a monitored `security_contact` email address,
- hardware and security evidence:
  - hardware profile that meets the recommended validator baseline,
  - sentry or network topology summary,
  - signing-key custody model (HSM, remote signer, or equivalent),
  - 24x7 operational escalation contact.

Incomplete packages are rejected before the ceremony begins.

## 2. Pre-Onboarding Checklist
- [ ] Hardware meets the recommended validator profile
- [ ] Dedicated validator plus sentry topology is planned
- [ ] NTP or chrony time sync is operational
- [ ] Signing keys are protected by HSM, remote signer, or offline custody
- [ ] Security contact mailbox is monitored
- [ ] 24x7 incident escalation path is documented
- [ ] Coordinator channel access is confirmed

## 3. Base System Preparation

### OS packages
```bash
sudo apt update
sudo apt install -y jq curl ufw chrony
```

### Time sync
```bash
timedatectl status
chronyc tracking
```

### Firewall
Open only the required ports. Restrict RPC access to operations networks.

```bash
sudo ufw allow 26656/tcp
sudo ufw allow from <ops-cidr> to any port 26657 proto tcp
sudo ufw allow from <monitoring-cidr> to any port 9100 proto tcp
sudo ufw enable
```

## 4. Install the VirtEngine Binary
Install the release-approved binary and verify the version before generating any
mainnet files.

```bash
curl -L https://releases.virtengine.com/virtengine/v1.0.0/virtengine-linux-amd64 -o virtengine
chmod +x virtengine
sudo mv virtengine /usr/local/bin/virtengine
virtengine version
```

## 5. Initialize the Home Directory
Create the validator home with the mainnet chain ID:

```bash
virtengine genesis init <moniker> \
  --chain-id virtengine-1 \
  --home ~/.virtengine \
  --overwrite
```

## 6. Create and Secure the Operator Key

```bash
virtengine keys add validator-operator \
  --keyring-backend file \
  --home ~/.virtengine
```

Record and back up:
- the funding account address printed by the key command,
- the keyring credentials,
- the operator key recovery material,
- the validator operator address used in the submitted gentx.

## 7. Build the Gentx
Create the gentx from the same home directory that holds the validator node
identity. Do not rely on CLI defaults for the public endpoint or metadata.

```bash
mkdir -p ./gentx

MONIKER="validator-01"
IDENTITY="validator-keybase-01"
PUBLIC_HOST="validator-01.ops.virtengine.net"
WEBSITE="https://validator-01.ops.virtengine.net"
SECURITY_CONTACT="noc@validator-01.ops.virtengine.net"
DETAILS="VirtEngine mainnet validator with sentry-backed public endpoint"

virtengine genesis gentx validator-operator 1000000000uve \
  --home ~/.virtengine \
  --chain-id virtengine-1 \
  --keyring-backend file \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1000000000 \
  --moniker "$MONIKER" \
  --identity "$IDENTITY" \
  --website "$WEBSITE" \
  --security-contact "$SECURITY_CONTACT" \
  --details "$DETAILS" \
  --ip "$PUBLIC_HOST" \
  --p2p-port 26656 \
  --output-document ./gentx/gentx-validator-operator.json
```

Important:
- `--ip` is mandatory in practice. If omitted, the CLI default may embed a
  private address such as `192.168.0.183`, and the ceremony will reject it.
- `--website` must be a real HTTPS endpoint. `example.*`, placeholder strings,
  or empty values are rejected.
- `--security-contact` is required and must be a valid email address.
- The self-delegation amount must already be funded in the approved mainnet
  allocations for the same `ve...` account.

## 8. Self-Validate Before Submission
Run the same coordinator validation locally before submitting the gentx:

```bash
scripts/mainnet/validate-gentx.sh \
  --gentx-dir ./gentx \
  --constraints ./config/mainnet/gentx-constraints.json
```

Do not submit if this command fails. Fix the gentx and regenerate it instead.

## 9. Submit to the Coordinator
Send the coordinator:
- `gentx/gentx-validator-operator.json`,
- the funding account address,
- the validator operator address,
- the public P2P endpoint,
- the security contact,
- the hardware and signing evidence package.

The coordinator will reject submissions that cannot be mapped to a funded
allocation or that do not satisfy the hardware/security policy.

## 10. Verify the Published Bundle
After the coordinator publishes the ceremony bundle, download:
- `genesis.json`
- `genesis.sha256`
- `gentx.sha256`
- `ceremony-manifest.json`
- `ceremony-manifest.sha256`

Verify the published hashes:

```bash
sha256sum -c genesis.sha256
sha256sum -c ceremony-manifest.sha256
```

Verify your submitted gentx is in the published gentx hash list:

```bash
jq -S . ./gentx/gentx-validator-operator.json | sha256sum
grep 'gentx-validator-operator.json' gentx.sha256
```

Do not start the node until both hash checks succeed and the coordinator calls
for validator ACK.

## 11. Install the Final Genesis and Start
Reinitialize the final home if needed, then install the published genesis:

```bash
virtengine genesis init <moniker> \
  --chain-id virtengine-1 \
  --home ~/.virtengine \
  --overwrite

cp ./genesis.json ~/.virtengine/config/genesis.json
sha256sum -c ./genesis.sha256
```

Apply the required networking configuration in:
- `~/.virtengine/config/config.toml`
- `~/.virtengine/config/app.toml`

Then start:

```bash
virtengine start --home ~/.virtengine
```

Verify node health:

```bash
virtengine status | jq
```

## 12. Post-Launch Monitoring
Minimum required alerts:
- block production halted,
- missed blocks above the operations threshold,
- disk usage above 80 percent,
- memory usage above 90 percent,
- low peer count,
- signer or sentry connectivity failures.

## 13. Incident and Exit Handling
- Use the validator coordination channel for any launch-blocking issue.
- Coordinate unbonding, key rotation, or maintenance windows with the release
  team before acting.
- Preserve incident timelines, hashes, and node logs for any genesis-related
  dispute.
