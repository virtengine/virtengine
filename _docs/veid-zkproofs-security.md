# VirtEngine VEID ZK Parameter Security

This document defines the production contract for the VEID Groth16 verification parameters under `x/veid/zk/params/`. It replaces the old single-party `dev` assumptions with ceremony-backed artifacts, checksum verification, and fail-closed startup behavior.

## Scope

The VEID module verifies three Groth16 circuits on BN254:

- `age` -> `age-range`
- `residency` -> `residency`
- `score` -> `score-range`

The application embeds only verification material:

- `age_vk.bin`
- `residency_vk.bin`
- `score_vk.bin`
- `params_metadata.json`
- `*.sha256` sidecars for every file above

Proving keys remain off-chain. Optional proving-key loading via `VEID_ZK_PK_DIR` is unchanged and is not used for validator startup.

## Current Embedded Parameter Set

The current embedded parameter set is `veid-zkparams-mpc-20260410`, generated on `2026-04-10T15:00:00Z` from deterministic two-participant trusted-setup exports. The embedded metadata schema is `virtengine.veid.zkparams/v1`, and each circuit record is bound to a verified trusted-setup artifact manifest.

| Circuit | Ceremony ID | VK SHA-256 | Artifact Manifest SHA-256 | Transcript SHA-256 |
| ------- | ----------- | ---------- | ------------------------- | ------------------ |
| `age` | `veid-1775822400` | `e428dc43068a3a4a8d1f3dc9f980e5f8ed0b83e255835852af28fdac810b9350` | `d3b215a136bb43137d98b361ee4add7ea0463ddafe2565242591c0bb819aefb3` | `30e24d82d75687c88478de93e110ccc53b87b5ec3e75955b60bf68a35d5d1fd4` |
| `residency` | `veid-1775826000` | `a3ac0c8d730df659d1d490c936ec56dc75c06bc27eeb3985bc0b08aef3f07b1e` | `c9a5fa60f6c72f69bc9447f8f29aefb78d78608678baf3e077ec09321cf1a360` | `8c413e01ce04ed5cf9773705da4c2f5fd9c12be9526ae00c0eb9e09ec93d59b0` |
| `score` | `veid-1775829600` | `bfa56bc83902d9cefb82207dca93578638f78adc98c80e8b3063254b299cf34f` | `5a892d181bf2b76bc42867734182095ab954f4e5a006a741204535027ca995ff` | `34407c910530dcf1925f2dfbef4add1876441fbf03d54541b0fbea83c1e7299d` |

All three circuit records are signed by coordinator ID `veid-coordinator` in the source trusted-setup export bundle, and the embedded metadata records the contributor IDs, transcript hash, circuit hash, and verification report hash for each circuit.

## Startup Verification Contract

Startup now fails closed in two places:

1. `x/veid/zk/params` validates the parameter bundle before returning a verifying key.
2. `x/veid/keeper.NewKeeper` panics if `NewZKProofSystem` cannot load verified parameters.

The loader rejects startup if any of the following is true:

- `params_metadata.json` or any `*.sha256` sidecar is missing.
- A sidecar hash does not match the file bytes it is supposed to protect.
- Metadata contains non-production marker content such as `dev`, `single-party`, or `dev-setup`.
- The metadata schema or artifact format does not match the expected production values.
- A circuit record is missing required fields, contributor counts, or artifact hashes.
- A verification key hash or size differs from the metadata record.
- The verification key bytes cannot be parsed as a BN254 Groth16 verifying key.
- The compiled circuit hash from the running code does not match the `circuit_hash` recorded in metadata.

There is no fallback path. If verification fails, the application does not continue with a nil ZK system.

## Staged Override Directory

For controlled rotation and recovery, the loader supports `VEID_ZK_PARAMS_DIR`.

If `VEID_ZK_PARAMS_DIR` is set, startup reads the parameter bundle from that directory instead of the embedded files. The directory must contain exactly the same runtime files:

- `age_vk.bin`
- `age_vk.bin.sha256`
- `residency_vk.bin`
- `residency_vk.bin.sha256`
- `score_vk.bin`
- `score_vk.bin.sha256`
- `params_metadata.json`
- `params_metadata.json.sha256`

This override exists to let operators stage a new verified ceremony bundle, prove startup behavior in staging, and roll back quickly without rebuilding binaries.

## Rotation Procedure

Use this sequence for every VEID parameter rotation.

1. Run or receive a completed trusted-setup export bundle for each VEID circuit.
   Required artifact: a verified `ceremony export-artifacts` output for `age-range`, `residency`, and `score-range`.

2. Verify each export bundle before touching runtime params.

```bash
go run ./tools/trusted-setup/cmd/ceremony verify-export --dir /path/to/age-export
go run ./tools/trusted-setup/cmd/ceremony verify-export --dir /path/to/residency-export
go run ./tools/trusted-setup/cmd/ceremony verify-export --dir /path/to/score-export
```

3. Build a staged runtime directory.

   Copy `phase2/verifying_key.bin` from each verified export as:

- `age_vk.bin`
- `residency_vk.bin`
- `score_vk.bin`

4. Generate SHA-256 sidecars for every staged file.

   Each sidecar must use the canonical format:

```text
<sha256>  <filename>
```

5. Update `params_metadata.json`.

   Every circuit record must be copied from the verified ceremony outputs:

- `ceremony_id`
- `artifact_manifest_hash`
- `verification_report_hash`
- `transcript_hash`
- `circuit_hash`
- `verification_key_hash`
- `verification_key_size_bytes`
- `contributor_count`
- `contributors`
- `coordinator_id`
- `coordinator_public_key`
- `beacon`
- `parameters_version`

6. Generate `params_metadata.json.sha256`.

7. Stage the directory in a startup environment.

```bash
export VEID_ZK_PARAMS_DIR=/srv/virtengine/veid-zkparams-next
```

8. Restart the application. Startup must succeed with the staged directory and must use the new metadata values.

9. Re-run the scoped validation gate:

```bash
go test ./x/veid/zk/... -count=1
go test ./x/veid/keeper/... -count=1
go test -tags=integration ./x/veid/zk/... -count=1
go test -tags "e2e.integration" ./x/veid/zk/... ./x/veid/keeper/... -count=1
```

10. After staging is accepted, either:

- keep the verified directory and corresponding deployment configuration as the active source of truth, or
- replace the embedded files under `x/veid/zk/params/` with the same verified bundle for the next released binary.

## Recovery And Rollback

If a rotation attempt fails, do not bypass the loader and do not remove checksums.

1. Restore the previous known-good parameter directory.
2. Repoint `VEID_ZK_PARAMS_DIR` to that directory, or restore the previous embedded bundle in the release artifact.
3. Restart the application.
4. Confirm that startup succeeds and the VEID keeper initializes normally.
5. Investigate the failed bundle using the exact loader error:

- checksum mismatch
- placeholder metadata
- verification key hash drift
- compiled circuit hash mismatch
- invalid verifying key parse failure

Rollback is complete only after the application starts with a fully verified bundle. A node that cannot verify VEID parameters must stay down rather than run with degraded proof verification.

## Operational Notes

- `params_metadata.json` is a signed-off publication record, not a scratch file.
- Contributor IDs and coordinator public keys are part of the runtime trust chain and must not be replaced with editorial summaries.
- Circuit updates require a new ceremony and a new `circuit_hash`; reusing old metadata against a new compiled circuit is rejected at startup.
- Recomputing only the `*.sha256` files is not enough. The metadata hashes must still match the verified ceremony artifacts.

## Test Coverage

This track ships three layers of coverage:

- Unit: `x/veid/zk/params` rejects placeholder metadata, checksum mismatches, and circuit drift.
- Integration: staged parameter directories load successfully only when hashes and compiled circuit bindings match.
- E2E: keeper startup accepts verified staged params and panics on placeholder or corrupted bundles.
