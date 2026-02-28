# VirtEngine Mainnet Genesis Guide

**Version:** 2.1.0
**Date:** 2026-04-11
**Task Reference:** B1

## Overview
This guide describes the reproducible process for producing the VirtEngine
mainnet `genesis.json` from the canonical repository inputs in `config/mainnet`
using the automation in `scripts/mainnet`.

The process is considered valid only when:
- the assembled genesis passes deterministic checks,
- `virtengine genesis validate` succeeds,
- the emitted SHA-256 files match the published artifacts,
- the launch packet evidence is updated with the final publication bundle.

## Canonical Inputs
- `config/mainnet/genesis-params.json`
- `config/mainnet/genesis-allocations.json`
- `config/mainnet/genesis-checks.json`
- `config/mainnet/gentx-constraints.json`
- `gentx/` directory containing the accepted validator submissions

## Publication Artifacts
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`
- `artifacts/mainnet/gentx.sha256`
- `artifacts/mainnet/ceremony-manifest.json`
- `artifacts/mainnet/ceremony-manifest.sha256`

As of `2026-04-11`, the final checked-in genesis bundle is published with
SHA-256 `a8d8a4a4f19882503265482c9433a6646d8dbbfe62f81c5945e81c32da9e6032`.

## Prerequisites

```bash
virtengine version
jq --version
sha256sum --version
python --version
```

The scripts will fail fast if `jq`, `sha256sum`, or a working Python 3 launcher
are unavailable.

## Mainnet Parameter Configuration
`config/mainnet/genesis-params.json` contains the approved chain and module
overrides. The parameter applicator merges those overrides onto the chain's
required default params so partial module overrides do not erase mandatory
fields from the genesis template.

Areas currently enforced by deterministic checks include:
- staking, slashing, gov, mint, and crisis settings,
- VEID and MFA defaults,
- encryption module settings,
- HPC, marketplace, and settlement parameters,
- chain ID and genesis time.

## Ceremony Flow
The coordinator runbook is in `_docs/runbooks/mainnet-genesis-ceremony.md`.
At a high level the ceremony is:

1. Validate every gentx against `config/mainnet/gentx-constraints.json`.
2. Build genesis from the frozen mainnet inputs with
   `scripts/mainnet/genesis-ceremony.sh`.
3. Confirm the emitted hashes with `sha256sum -c`.
4. Update launch evidence and run `scripts/mainnet/prelaunch-checklist.sh`.
5. Publish the full artifact bundle and collect validator ACKs.

### Example Ceremony Invocation

```bash
scripts/mainnet/genesis-ceremony.sh \
  --gentx-dir ./gentx \
  --home ./.cache/mainnet-genesis \
  --output ./artifacts/mainnet \
  --params ./config/mainnet/genesis-params.json \
  --allocations ./config/mainnet/genesis-allocations.json \
  --checks ./config/mainnet/genesis-checks.json \
  --constraints ./config/mainnet/gentx-constraints.json
```

## Failure Conditions Enforced by the Tooling
The mainnet scripts are intentionally fail-closed. The ceremony will stop if it
detects any of the following:
- blocked or evidence-incomplete allocations,
- placeholder or duplicate allocation addresses,
- invalid vesting timestamps or malformed allocation totals,
- gentxs with placeholder metadata, missing `security_contact`, or private P2P
  endpoints,
- gentxs whose self-delegation is not funded by the approved allocations,
- placeholder strings inside the assembled genesis payload,
- bank supply totals that do not equal the approved allocations,
- launch packet evidence rows with placeholder paths, missing files, or
  mismatched hashes,
- native `virtengine genesis validate` failures.

## Deterministic Hash Verification
The ceremony canonicalizes JSON artifacts before publication. Verify the
published bundle with:

```bash
sha256sum -c ./artifacts/mainnet/genesis.sha256
sha256sum -c ./artifacts/mainnet/ceremony-manifest.sha256
scripts/mainnet/genesis-hash.sh --genesis ./artifacts/mainnet/genesis.json
```

The output of `genesis-hash.sh` must match the hash recorded in
`artifacts/mainnet/genesis.sha256`.

The approved allocation set is archived in
`_docs/operations/mainnet-allocation-control-record-2026-04-11.md`.

## Validator Onboarding Summary
Validator-specific instructions live in `_docs/runbooks/validator-onboarding.md`.
Every validator must:
- generate a gentx with a real public P2P endpoint,
- provide non-placeholder website and security contact metadata,
- confirm their self-delegation account is funded in the approved allocations,
- verify the published `genesis.json`, `gentx.sha256`, and
  `ceremony-manifest.json` before starting.

## Launch Packet and Checklist Gate
Before publication, update the launch evidence entries in
`_docs/operations/mainnet-launch-packet.md` and run:

```bash
scripts/mainnet/prelaunch-checklist.sh
```

Use `--allow-pending` or `--allow-unchecked` only for rehearsal or draft review
runs. Final publication must pass without either flag.
