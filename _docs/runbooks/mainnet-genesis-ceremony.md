# Mainnet Genesis Ceremony Runbook

**Version:** 2.1.0
**Last updated:** 2026-04-11
**Owner:** Release Management (Ops)

## Purpose
Provide the exact operator sequence for assembling, validating, and publishing
the VirtEngine mainnet genesis bundle. The ceremony is complete only when the
published artifacts are byte-stable, every validator can reproduce the same
hashes, and the launch packet evidence set is updated with the final values.

## Roles
- **Genesis Coordinator**: freezes inputs, validates submissions, builds the
  final bundle, publishes artifacts, and records launch evidence.
- **Validator Operator**: submits a valid gentx, verifies the published bundle,
  and refuses to start if any hash differs.
- **Release Manager**: confirms the approved input set and signs off on final
  publication.

## Canonical Inputs
- `config/mainnet/genesis-params.json`
- `config/mainnet/genesis-allocations.json`
- `config/mainnet/genesis-checks.json`
- `config/mainnet/gentx-constraints.json`
- `gentx/` directory containing all accepted validator gentx JSON files

## Publication Bundle
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`
- `artifacts/mainnet/gentx.sha256`
- `artifacts/mainnet/ceremony-manifest.json`
- `artifacts/mainnet/ceremony-manifest.sha256`

## Hard Stops
The ceremony must stop immediately if any of the following occur:
- `config/mainnet/genesis-allocations.json` is not `READY`, has an empty
  `accounts[]`, omits the required `community_pool` block, or references
  missing approval evidence.
- Any allocation contains a placeholder or duplicate address, an invalid vesting
  window, a malformed coin amount, or insufficient funds for a submitted
  validator self-delegation.
- Any gentx uses placeholder metadata, omits `security_contact`, uses a
  non-HTTPS website, advertises a private or unroutable P2P endpoint, or
  duplicates another validator address or consensus key.
- Any deterministic genesis check fails, including placeholder strings in the
  assembled payload or bank supply totals that do not match allocations.
- `virtengine genesis validate` fails on the assembled home directory.
- The launch packet evidence manifest contains a placeholder path, a missing
  file, or a mismatched SHA-256 hash.

## Prerequisites
Run the ceremony from the repository root with:
- `virtengine` available on `PATH`, or pass `--binary`.
- `jq`, `sha256sum`, and a working Python 3 launcher available.
- A clean staging location for `gentx/`, `.cache/mainnet-genesis`, and
  `artifacts/mainnet/`.

## Coordinator Sequence

### 1. Freeze the input set
1. Confirm the final reviewed copies of:
   - `config/mainnet/genesis-params.json`
   - `config/mainnet/genesis-allocations.json`
   - `config/mainnet/genesis-checks.json`
   - `config/mainnet/gentx-constraints.json`
2. Confirm the allocation approval record referenced by
   `config/mainnet/genesis-allocations.json` exists and that the
   `community_pool.address` matches
   `virtengine debug module-address distribution`.
3. Collect every accepted validator submission into `./gentx/`.
4. Remove any prior ceremony scratch state:

```bash
rm -rf ./.cache/mainnet-genesis ./artifacts/mainnet
mkdir -p ./gentx ./artifacts/mainnet
```

### 2. Validate the submitted gentxs
Run the validator gate before any genesis assembly:

```bash
scripts/mainnet/validate-gentx.sh \
  --gentx-dir ./gentx \
  --constraints ./config/mainnet/gentx-constraints.json
```

This command is expected to fail if a submission contains:
- a placeholder website, identity, or details string,
- a missing or invalid `security_contact`,
- a private `memo` host such as `192.168.x.x`, `localhost`, or `example.*`,
- duplicate validator addresses, duplicate consensus keys, or duplicate funding
  accounts across gentxs,
- self-delegation or commission values outside the enforced mainnet bounds.

### 3. Assemble the final genesis bundle
Run the full ceremony exactly once against the frozen inputs:

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

The ceremony script will:
1. initialize a fresh genesis home with `virtengine genesis init`,
2. apply the approved parameter overrides while preserving required module
   defaults,
3. add every allocation account, vesting schedule, and the distribution module
   account funding required for the initial community pool,
4. reject blocked or incomplete allocation evidence,
5. validate gentxs and verify each validator self-delegation is funded by the
   approved allocations,
6. collect gentxs into the final `genesis.json`,
7. canonicalize the final JSON before hashing,
8. run deterministic checks plus `virtengine genesis validate`,
9. emit the full publication bundle in `artifacts/mainnet/`.

### 4. Verify the emitted hashes
The coordinator must verify the generated hash files before publication:

```bash
sha256sum -c ./artifacts/mainnet/genesis.sha256
sha256sum -c ./artifacts/mainnet/ceremony-manifest.sha256
```

Recompute the canonical genesis hash and confirm it matches
`artifacts/mainnet/genesis.sha256`:

```bash
scripts/mainnet/genesis-hash.sh --genesis ./artifacts/mainnet/genesis.json
```

`gentx.sha256` is the sorted canonical JSON hash list for every published gentx.
Do not edit or regenerate any gentx after this point.

### 5. Update launch evidence and run the readiness gate
1. Record the final artifact paths and SHA-256 values in
   `_docs/operations/mainnet-launch-packet.md`.
2. Run the readiness/evidence gate:

```bash
scripts/mainnet/prelaunch-checklist.sh
```

Do not use `--allow-pending` or `--allow-unchecked` for the final publication
run. Those flags are only for rehearsal or draft evidence reviews.

### 6. Publish the bundle
Publish exactly these files to the validator coordination channel and the
release archive:
- `genesis.json`
- `genesis.sha256`
- `gentx.sha256`
- `ceremony-manifest.json`
- `ceremony-manifest.sha256`

The publication message must include:
- the chain ID,
- the genesis time,
- the SHA-256 of `genesis.json`,
- the path or bucket location of the full artifact bundle,
- a request for validator ACK before node start.

### 7. Collect validator ACKs
Every validator must confirm all of the following before go-live:
- `sha256sum -c genesis.sha256` succeeds on the downloaded `genesis.json`,
- `sha256sum -c ceremony-manifest.sha256` succeeds,
- the operator has matched their submitted gentx against the published
  `gentx.sha256`,
- there is no divergence between the locally downloaded bundle and the
  coordinator announcement.

## Rollback Criteria
Abort publication and return to the input review stage if:
- any validator reports a hash mismatch,
- any evidence row cannot be backed by an actual file and matching hash,
- a corrected gentx or allocation is submitted after the bundle is built,
- a rerun of `genesis-ceremony.sh` against the same frozen inputs produces a
  different `genesis.sha256`.

## Post-Ceremony Records
Archive:
- the published artifact bundle,
- the final `gentx/` directory used for the build,
- coordinator terminal logs,
- launch packet updates and release manager approval evidence.
