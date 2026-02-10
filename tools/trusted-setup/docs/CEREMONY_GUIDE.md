# VirtEngine ZK Trusted Setup Ceremony Guide

This guide describes the recommended multi-party trusted setup workflow for the VEID Groth16 circuits (BN254). It is intended for coordinators and participants.

## Quick Start (Coordinator)

1. Create a working directory and initialize Phase 1:

```
ceremony init-phase1 --dir ceremony-data --domain 65536 --min-contributors 20
```

2. Initialize Phase 2 for each circuit after collecting Phase 1 contributions:

```
ceremony init-phase2 --dir ceremony-data --circuit age --beacon <beacon>
ceremony init-phase2 --dir ceremony-data --circuit residency --beacon <beacon>
ceremony init-phase2 --dir ceremony-data --circuit score --beacon <beacon>
```

3. After collecting Phase 2 contributions, finalize each circuit:

```
ceremony finalize --dir ceremony-data --circuit age --beacon <beacon>
ceremony finalize --dir ceremony-data --circuit residency --beacon <beacon>
ceremony finalize --dir ceremony-data --circuit score --beacon <beacon>
```

4. Copy the verifying keys into `x/veid/zk/params/` and update `params_metadata.json`.

## Quick Start (Participant)

### Online Participation

```
ceremony participate --url https://coordinator.example.com --participant <id> --circuit phase1
ceremony participate --url https://coordinator.example.com --participant <id> --circuit age
```

### Air-Gapped Participation

1. Download the latest transcript file from the coordinator.
2. Contribute offline:

```
ceremony contribute-phase1 --in phase1_latest.bin --out phase1_contrib.bin
```

3. Transfer the output back to the coordinator for validation.

## Recommended Operational Checklist

- Require participants to sign attestations for their contribution hashes.
- Ensure phase1 and phase2 contributions are sequential and verified.
- Use a public randomness beacon for final sealing (Phase 1 and Phase 2).
- Publish transcripts, contribution hashes, and final key hashes for audit.

## Air-Gapped Guidance

- Use a dedicated offline machine.
- Download the latest transcript via trusted media.
- Provide entropy locally (keyboard timing, hardware RNG, etc.).
- Securely delete intermediate files after generating the contribution.

## Output Artifacts

- Phase 1 transcript files: `phase1/phase1_*.bin`
- Phase 2 transcript files: `phase2/<circuit>/phase2_*.bin`
- Final proving key: `<circuit>_pk.bin`
- Final verifying key: `<circuit>_vk.bin`
- Transcript: `transcripts/<circuit>.json`

## Notes

- Phase 1 domain size must be a power of two and larger than circuit constraints.
- You must run a new ceremony for every circuit change.