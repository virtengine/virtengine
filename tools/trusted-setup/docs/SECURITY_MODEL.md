# Trusted Setup Security Model

This document summarizes the security assumptions and mitigations for the VEID Groth16 trusted setup ceremony.

## Threat Model

- Attackers can observe all public transcripts and key material.
- Attackers may control some contributors.
- At least one honest contributor must destroy their secret randomness to retain soundness.
- Coordinator may be partially trusted but must be verifiable by transcript checks.

## Mitigations

- Multi-party contributions across diverse organizations and geographies.
- Verifiable transcript hashes for every contribution.
- Final randomness beacon applied during Phase 1 and Phase 2 sealing.
- Independent verification tools to re-derive keys from transcripts.
- Air-gapped contribution option to protect contributor entropy.

## Verification Steps

1. Verify all Phase 1 contributions via transcript hashes and proofs.
2. Verify Phase 2 contributions against the circuit R1CS.
3. Recompute proving/verifying key hashes and match published values.
4. Confirm beacon values and timestamps.

## Operational Guidance

- Require participant attestations (signed statements) that match contribution hashes.
- Publish all transcripts and verification commands.
- Retain audit logs for coordinator actions and file hashes.

## Limitations

- Groth16 requires per-circuit trusted setup; any circuit change requires a new ceremony.
- This model does not remove the trusted setup assumption; it reduces risk via MPC.