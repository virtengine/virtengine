# VirtEngine Mainnet Genesis Guide

**Version:** 1.1.0  
**Date:** 2026-02-06  
**Task Reference:** 22B

---

## Table of Contents

1. [Overview](#overview)
2. [Inputs and Artifacts](#inputs-and-artifacts)
3. [Prerequisites](#prerequisites)
4. [Mainnet Parameter Configuration](#mainnet-parameter-configuration)
5. [Genesis Ceremony (Production)](#genesis-ceremony-production)
6. [Genesis Validation Checks](#genesis-validation-checks)
7. [Validator Onboarding (Summary)](#validator-onboarding-summary)
8. [Genesis Distribution + Hash Verification](#genesis-distribution--hash-verification)
9. [Pre-Launch Checklist Automation](#pre-launch-checklist-automation)

---

## Overview

This guide defines the deterministic process for producing the VirtEngine
mainnet `genesis.json` using audited configuration inputs and ceremony
tooling. Production parameters live in `config/mainnet/*` and are applied via
`scripts/mainnet/*` so the final genesis is reproducible across all
participants.

## Inputs and Artifacts

### Canonical Inputs
- `config/mainnet/genesis-params.json` — chain + module parameters
- `config/mainnet/genesis-allocations.json` — initial account allocations
- `config/mainnet/gentx-constraints.json` — gentx validation rules
- `config/mainnet/genesis-checks.json` — deterministic validation assertions

### Ceremony Inputs
- `gentx/` — directory containing validator gentx JSON files

### Outputs
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`

## Prerequisites

```bash
virtengine version
jq --version
sha256sum --version
```

## Mainnet Parameter Configuration

All mainnet parameters are captured in `config/mainnet/genesis-params.json`.
This file is treated as the canonical source of production values.

### Chain Parameters (Core)
- **Staking**: 2-week unbonding, 100 validators, 5% minimum commission
- **Mint**: inflation bounds 7–20%, goal bonded 67%
- **Gov**: 2-week max deposit, 3-day voting, 20% quorum
- **Slashing**: 30,000 block window, 5% min signed, 5% double-sign slash

### VEID / MFA / Encryption / HPC
- **VEID**: deterministic score tiers, signatures required
- **MFA**: 15-minute session, 5-minute challenge TTL
- **Encryption**: X25519-XSalsa20-Poly1305 only, signatures required
- **HPC**: fee distribution and routing enforcement enabled

### Marketplace Module (mktplace)
Marketplace params use chain defaults unless explicitly overridden. The module
name in genesis is `mktplace` (not `market`). If overrides are required, add
explicit checks to `config/mainnet/genesis-checks.json` and document in the
launch packet.

## Genesis Ceremony (Production)

All ceremony steps are automated using `scripts/mainnet/genesis-ceremony.sh`.
This script performs:
1. `virtengine init` with the target chain ID
2. Applies `genesis-params.json`
3. Adds all allocation accounts
4. Validates gentxs against constraints
5. Collects gentxs and builds the final genesis
6. Runs deterministic validation checks
7. Outputs genesis + hash to `artifacts/mainnet/`

### Example
```bash
scripts/mainnet/genesis-ceremony.sh \
  --gentx-dir ./gentx \
  --output ./artifacts/mainnet \
  --chain-id virtengine-1 \
  --genesis-time 2026-06-01T00:00:00Z
```

## Genesis Validation Checks

Deterministic checks live in `config/mainnet/genesis-checks.json` and are
executed via:

```bash
scripts/mainnet/genesis-validate.sh \
  --genesis artifacts/mainnet/genesis.json \
  --checks config/mainnet/genesis-checks.json
```

This verifies core parameter values, VEID/MFA/Encryption/HPC defaults, and
ensures the chain ID + genesis time match the approved values.

## Validator Onboarding (Summary)

Validator onboarding is documented end-to-end in:
`_docs/runbooks/validator-onboarding.md`

Key onboarding dependencies:
- Hardware profile: `_docs/validators/hardware-requirements.md`
- Gentx requirements: `config/mainnet/gentx-constraints.json`

## Genesis Distribution + Hash Verification

Distribute the finalized `genesis.json` and hash to all validators:

```bash
sha256sum artifacts/mainnet/genesis.json
```

All validators MUST confirm the hash matches before starting nodes.

## Pre-Launch Checklist Automation

The pre-launch automation checks readiness checklists + evidence hashes:

```bash
scripts/mainnet/prelaunch-checklist.sh
```

Use `--allow-pending` or `--allow-unchecked` only during dry runs.