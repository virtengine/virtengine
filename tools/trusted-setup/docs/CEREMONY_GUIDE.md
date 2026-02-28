# VirtEngine Trusted Setup Ceremony Guide

This guide documents the exact coordinator and participant flow supported by `tools/trusted-setup`. The ceremony is restart-safe, contribution signatures are required, air-gapped participation is supported through signed request and response bundles, and final publication artifacts are exported as a signed manifest plus checksums.

## Directory Layout

The coordinator keeps all mutable state under a single working directory:

- `config.json`: ceremony configuration
- `transcript.json`: canonical transcript and contribution audit trail
- `phase1/commons.bin`: phase 1 public parameters
- `phase1/contrib-*.bin`: accepted phase 1 contributions
- `phase2/r1cs.bin`: compiled circuit
- `phase2/contrib-*.bin`: accepted phase 2 contributions
- `phase2/proving_key.bin`: finalized proving key
- `phase2/verifying_key.bin`: finalized verifying key

The coordinator can stop and restart at any point. Re-running `init`, `start-phase2`, or `finalize` against the same state directory is safe as long as the ceremony inputs have not changed.

## Coordinator Flow

1. Initialize the ceremony state:

```bash
ceremony init \
  --dir ceremony-data \
  --circuit score-range \
  --min-contributors 2 \
  --beacon "public-beacon-2026-04-10" \
  --note "mainnet veid circuit"
```

2. Inspect current progress at any time:

```bash
ceremony status --dir ceremony-data
ceremony verify-transcript --dir ceremony-data
```

3. If participants will contribute online, run the coordinator HTTP server:

```bash
ceremony server --dir ceremony-data --addr :8080
```

4. After the minimum number of valid phase 1 contributions is accepted, start phase 2:

```bash
ceremony start-phase2 --dir ceremony-data
```

5. After the minimum number of valid phase 2 contributions is accepted, finalize:

```bash
ceremony finalize --dir ceremony-data --version veid-score-range-v1
```

6. Export the publication bundle with a coordinator signing identity:

```bash
ceremony export-artifacts \
  --dir ceremony-data \
  --out ceremony-publication \
  --identity coordinator_identity.json \
  --signer coordinator
```

7. Verify the publication bundle before distribution:

```bash
ceremony verify-export --dir ceremony-publication
```

## Participant Flow

Participants need an identity file and a stable participant identifier. The identity is used to sign contribution metadata; the coordinator rejects missing or invalid signatures.

### Online Contribution

Phase 1:

```bash
ceremony participate \
  --url http://coordinator.example.com:8080 \
  --phase phase1 \
  --identity participant_identity.json \
  --participant alice \
  --attestation "offline entropy mixed with hardware rng"
```

Phase 2:

```bash
ceremony participate \
  --url http://coordinator.example.com:8080 \
  --phase phase2 \
  --identity participant_identity.json \
  --participant alice \
  --attestation "offline entropy mixed with hardware rng"
```

### Air-Gapped Contribution

The air-gapped flow uses signed JSON manifests plus SHA-256 sidecars so the coordinator and participant can verify what moved across removable media.

Coordinator exports the current phase bundle:

```bash
ceremony export-phase-bundle \
  --dir ceremony-data \
  --phase phase1 \
  --out transfer/phase1-request
```

The exported directory contains:

- `current.bin`
- `current.bin.sha256`
- `request.json`
- `request.json.sha256`

Participant responds on the offline system:

```bash
ceremony contribute-bundle \
  --bundle transfer/phase1-request \
  --out transfer/phase1-response \
  --identity participant_identity.json \
  --participant alice \
  --attestation "air-gapped workstation with local entropy"
```

The response directory contains:

- `contribution.bin`
- `contribution.bin.sha256`
- `response.json`
- `response.json.sha256`

Coordinator imports the response:

```bash
ceremony accept-bundle \
  --dir ceremony-data \
  --phase phase1 \
  --bundle transfer/phase1-response
```

Repeat the same sequence with `--phase phase2` after the coordinator starts phase 2.

## Validation Rules

The coordinator fails closed when any of the following is true:

- The participant signature does not match the claimed public key, input hash, output hash, or attestation.
- The request or response bundle hash does not match the payload on disk.
- A contribution does not chain from the current accepted phase state.
- A duplicate submission changes the payload or signature for an already accepted step.
- Phase 2 starts before enough valid phase 1 contributions exist.
- Finalization runs before enough valid phase 2 contributions exist.

## Publication Bundle

`ceremony export-artifacts` writes a signed, verifiable output set outside the coordinator state directory:

- `artifact-manifest.json`
- `artifact-manifest.json.sha256`
- `artifact-manifest.sig`
- `artifact-manifest.sig.sha256`
- `verification.json`
- `verification.json.sha256`
- `config.json`
- `transcript.json`
- accepted `phase1/contrib-*.bin` files
- accepted `phase2/contrib-*.bin` files
- `phase1/commons.bin`
- `phase2/r1cs.bin`
- `phase2/proving_key.bin`
- `phase2/verifying_key.bin`

Anyone validating the release should run:

```bash
ceremony verify-export --dir ceremony-publication
```

This checks the signed manifest, transcript hash, verification report, file sizes, and file SHA-256 digests.

## Operational Guidance

- Keep mutable ceremony state private to the coordinator environment until export time.
- Use a fresh removable medium per air-gapped transfer when practical.
- Record participant attestations exactly as provided; they are part of the signed contribution envelope.
- Publish the signed artifact bundle, not ad hoc copied files.
- Treat any verification failure as a hard stop and restart from the last verified coordinator state rather than manually editing transcript files.
