# Mainnet Genesis Ceremony Runbook

**Version:** 1.0.0  
**Last updated:** 2026-02-06  
**Owner:** Release Management (Ops)

## Purpose
Provide a coordinated, deterministic process for generating the final
`genesis.json` for mainnet launch, including validation, signing, and
distribution steps.

## Roles
- **Genesis Coordinator** — collects gentxs, runs ceremony tooling, publishes
  final genesis + hash.
- **Validator Operator** — generates gentx, verifies hash, starts nodes.
- **Release Manager** — approves final parameters + signs off on evidence.

## Inputs
- `config/mainnet/genesis-params.json`
- `config/mainnet/genesis-allocations.json`
- `config/mainnet/gentx-constraints.json`
- `gentx/` directory of validator gentx files

## Output
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`

## Ceremony Steps

### 1) Pre-ceremony Validation (Coordinator)
- [ ] Parameters reviewed and approved
- [ ] Allocation addresses verified
- [ ] gentx submissions complete
- [ ] Validator list confirmed

### 2) Validate gentx submissions
```bash
scripts/mainnet/validate-gentx.sh --gentx-dir ./gentx
```

### 3) Build genesis
```bash
scripts/mainnet/genesis-ceremony.sh \
  --gentx-dir ./gentx \
  --output ./artifacts/mainnet
```

### 4) Verify deterministic hash
```bash
sha256sum artifacts/mainnet/genesis.json
```

### 5) Publish
- Publish `genesis.json` + hash in the validator coordination channel
- Archive the evidence in the launch packet

### 6) Validator confirmation
- Each validator confirms the hash matches before start

## Rollback Criteria
- Any validator reports hash mismatch
- gentx fails validation checks
- genesis validation fails

## Post-ceremony
- Update `_docs/operations/mainnet-launch-packet.md` with final hash
- Notify release manager to proceed with go/no-go meeting