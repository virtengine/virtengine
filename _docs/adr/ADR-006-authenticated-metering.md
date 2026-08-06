# ADR-006: Authenticated, Replay-Safe Metering

**Status:** Accepted for local implementation
**Date:** 2026-07-21
**Upgrade:** `v1.5.0`
**Settlement module:** consensus version `2`
**Provider module:** consensus version `4`

## Decision

VirtEngine usage settlement uses version-1 detached signatures over an explicitly length-prefixed binary payload. JSON, maps, protobuf unknown-field behavior, transaction sender authentication, wall time, and external services are not sign-byte inputs or substitutes for the detached proof.

Provider proofs use domain `virtengine.settlement.usage.provider.v1`; customer acknowledgments use `virtengine.settlement.usage.customer.v1`. Every byte string is prefixed by a big-endian `uint32` length. Integers are fixed-width big-endian. Signed decimals must equal the canonical Cosmos SDK 18-decimal string representation. The provider payload begins with `VEUSAGE\x01`; acknowledgment begins with `VEUSACK\x01`.

The provider payload covers:

1. signature version, chain ID, domain, signer role;
2. provider and customer addresses;
3. order, lease, and optional allocation lineage;
4. period start/end;
5. CPU millisecond-seconds, memory/storage byte-seconds, network in/out bytes, and GPU seconds;
6. pricing version, usage units/type, unit-price denom and exact canonical amount;
7. formula/model versions;
8. stream sequence, 32-byte nonce, and 32-byte idempotency key;
9. provider key epoch and deterministic key ID;
10. issued/expiry height and block-time bounds.

The customer payload covers the exact stored authenticated usage digest, usage ID, customer address, chain ID, independent replay key, signature version, and issued/expiry bounds.

## Trust roots and algorithms

Provider verification resolves the exact immutable key epoch from `x/provider`; daemon-local or validator-local keys are never accepted as validators' trust root. X25519 is encryption-only and cannot authenticate usage or key rotation. Ed25519 uses raw 64-byte signatures. secp256k1 uses raw 64-byte `R || S`, SHA-256 as implemented by the Cosmos key type, and mandatory low-S canonicality. DER, recoverable 65-byte, high-S, multisig, ethsecp256k1, and unknown account key forms are rejected.

Provider key rotation is authorized by a version-1 binary proof signed by the retiring key. Epochs increase by one. The old epoch has a bounded overlap of 100 blocks and 24 hours; either bound ending makes it invalid. Revocation has no grace period. Height and block time come from `sdk.Context`.

## Replay and period policy

A stream is SHA-256 over length-prefixed provider, lineage kind, allocation, order, and lease. Allocation is primary when present, but order and lease remain bound. Sequence starts at one and permits no gaps or regressions. Sequence, nonce, and idempotency-key indexes all point to the original usage ID and digest.

An exact retry returns the original usage ID and emits no usage event, billing record, reward, escrow, or reconciliation side effect. Any partial collision or different digest fails. State mutation executes through a Cosmos cached context.

Periods for the same stream and metric type may be contiguous or have a bounded gap up to 24 hours, but may not overlap. Provider proofs are accepted within 1,000 past blocks, two future blocks, two past hours, and two future minutes, and have maximum lifetime 200 blocks/one hour. A usage period is at most 24 hours. Raw dimensions and units are nonnegative bounded integers. Consensus code uses no floats.

Version 1 is the only available pricing/formula/model version. Maximum raw dimension is `10^18`; maximum usage units is `10^15`; integer total cost must fit signed 64-bit. New versions require additive protocol work and fixtures.

## Migration and rollback

`v1.5.0` runs provider `3 -> 4` and settlement `1 -> 2` migrations. Provider migration backfills current keys into immutable epoch history. Settlement migration marks all existing usage as `legacy_unverified`, clears any inferred digest/verification state, then writes the authentication activation marker.

Legacy records remain queryable but cannot newly settle, bill, reward, release escrow, or satisfy usage conditions after activation. Fresh genesis enables authentication by default and accepts only authenticated usage fixtures. Rolling back below `v1.5.0` after migration is unsupported because old binaries do not enforce the replay indexes or legacy financial gate. Before activation, existing transaction behavior remains available for historical compatibility.

## HPC and provider producer decisions

HPC keepers never create signatures and never hold private keys. An off-chain provider/node agent must submit canonical authenticated usage first and attach its IDs to the HPC accounting record. Unsigned HPC accounting remains pending and unbillable.

The provider submitter resolves the active governed key from real gRPC/Comet chain state, checks the local KeyManager key ID/public key/provider/algorithm, persists stream sequence and proof allocation in the durable transaction queue, signs the canonical payload, then sends the generated `MsgRecordUsage` through the durable signed Cosmos transaction path. Restart reuses identical proof allocation and bytes.

## Golden vectors

The canonical usage vector is checked in at `x/settlement/types/metering_auth_test.go`. Legacy protobuf vectors and field-number guards are checked in at `tests/compatibility/task84b_wire_compatibility_test.go`.
