# Task 85C Completion Report — 2026-07-23

## Status

`engineering_complete_external_blocked`

Task 85C is complete at the deterministic local engineering boundary. No live
TMKMS/HSM, multi-zone Kubernetes storage, cloud snapshot, regional failover, or
real validator double-sign drill is claimed.

## Implemented

- `deploy/kubernetes` is the canonical application source. The
  `infra/kubernetes` application base and environment overlays import canonical
  resources byte-for-byte; chaos and DR assets remain documented consumers.
  Direct production application uses
  `kubectl apply -k deploy/kubernetes/overlays/prod`, and GitOps targets the same
  Kustomize root.
- Production renders one explicit remote-signer validator and four scalable
  sentry/full nodes. Validator pods do not mount a consensus private key and
  require chain ID, validator address, key fingerprint, signer endpoint, signer
  epoch, fencing token, TMKMS config digest, signer certificate digest, remote
  signer connectivity, and chain-status identity before readiness.
- Provider production renders as a three-replica StatefulSet with shared
  encrypted identity/HA/backup PVCs, PDB, anti-affinity, topology spread, restricted
  pod security, immutable image digests, canonical metrics port 9090, and
  identity/fencing-aware `/ready`.
- Provider mutation state is safe for multiple processes: shared file-store
  operations reload under cross-process locking and atomically replace state.
  Kubernetes `coordination.k8s.io/v1` Leases provide production ownership;
  resource-version updates are the atomic boundary and `leaseTransitions` is a
  monotonic fencing token. Standbys stay unready until failover.
- A chain-client mutation guard rejects all provider write producers while a
  replica is standby, before bid/resource/domain/HPC workers can reach the
  durable mutation submitter. Ownership loss also stops authenticated metering;
  takeover restarts it only after the higher fencing token is active.
- The file key manager persists an Argon2id/AES-256-GCM authenticated encrypted
  keystore with atomic writes, metadata/private-key integrity validation,
  restart/rotation continuity, non-empty secret-file passphrases, and public key
  fingerprint binding. Production never generates a missing identity.
- Hardware, Ledger, and non-custodial key profiles fail closed rather than
  falling back to software. The in-repository SoftHSM provider remains explicitly
  development/test software and is not hardware certification.
- Provider DR scripts and the local smoke drill preserve encrypted identity,
  queue/idempotency records, usage sequence/proof allocation, mutation sequence
  and reconciliation state, fiat reconciliation state, and fencing epochs.
- Kubernetes DR backs up validator chain state only. Remote-signer private
  material, mTLS identity, anti-double-sign state and fencing state remain in
  the independent TMKMS/HSM custody boundary and must be restored and advanced
  there before its metadata projection is rehydrated.
- Semantic render policy rejects a parallel infra topology, unsafe validator
  scaling, signer leakage into sentries, ephemeral provider state, missing
  hardening, mutable images, missing Kubernetes Lease RBAC, and metrics drift.
- Infrastructure CI now enforces the semantic render policy before downstream
  security, plan, apply, or drift jobs. The workflow triggers on `infra/**`,
  `deploy/**`, the Task 85C validator, the INFRA-001 implementation summary,
  focused Task 85C ADR/completion/recovery docs, and the workflow file. The
  validate job installs Node 20 plus checksum-pinned kubectl 1.29.0 and runs
  `node scripts/task85c-validate-kubernetes.mjs` with read-only permissions and
  no deployment credentials.
- `infra/rollouts` no longer carries application topology. The deleted
  rollout workload manifests cannot be applied as an alternate deployment
  source; the retained rollback configuration is optional AnalysisTemplate
  policy only and the directory must never be applied wholesale.
- Release promotion replaces canonical image references with immutable
  `name@sha256:<64-hex-digest>` values before applying the canonical overlay.

## Local Acceptance Evidence

The full `pwsh scripts/task85c-preflight.ps1` gate passed on 2026-07-23:

- `gofmt` clean for Task 85C Go files.
- Complete `go test ./pkg/provider_daemon ./cmd/provider-daemon -count=1` passed.
- Focused settlement durable mutation continuity passed.
- `go vet ./pkg/provider_daemon ./cmd/provider-daemon` passed.
- `golangci-lint` reported `0 issues` for provider-daemon packages.
- `go build ./cmd/provider-daemon` passed.
- Canonical/base/dev/staging/prod plus infra compatibility and chaos renders
  passed semantic policy. Every infra application render matched canonical.
- Policy rejected stale deployment-guide instructions for applying
  `infra/rollouts` as a directory, stale external-secrets paths, workload kinds
  under `infra/rollouts`, and mutable image references in production-consumable
  infrastructure manifests. It scans actual rollout files and documentation;
  provider metrics validation is scoped to the canonical provider metrics role
  and does not reserve port 8445 globally.
- AGENTS documentation validation passed for 9 files.
- Chain and provider backup/restore smoke passed, including provider encrypted
  identity, queue, sequence, fencing, and reconciliation continuity.
- WSL race tests passed for keystore, file/Kubernetes lease, split-brain,
  standby takeover, shared store, readiness, passphrase, and restart tests.
- Repository-wide `go test ./... -run '^$' -count=1` compilation passed.
- Whitespace validation passed.

The final CI enforcement gap validation also passed on 2026-07-23:

- The configured kubectl 1.29.0 Linux amd64 SHA-256 matched the official
  `dl.k8s.io` checksum.
- `.github/workflows/infrastructure.yaml` parsed locally with PyYAML and passed
  `actionlint` 1.7.7, including action input validation.
- `node scripts/task85c-validate-kubernetes.mjs` passed.
- `node scripts/validate-agents-docs.mjs` passed.
- `git diff --check` passed.

## Release-Only / External Blockers

The following require external infrastructure and retained evidence before any
production certification:

1. Real TMKMS or vendor HSM/PKCS#11 integration with mTLS, independently checked
   chain ID/address/key fingerprint/epoch/fencing continuity, outage/failover,
   and double-sign protection evidence.
2. Real validator active/passive failover and adversarial partition drill. Local
   render and metadata checks do not prove a remote signer refuses stale votes.
3. Named Kubernetes distribution/version, External Secrets backend, encrypted
   RWO/RWX storage classes, VolumeSnapshot/backup controller, immutable image
   provenance, and Pod Security admission conformance.
4. Multi-zone pod eviction, detached-volume, rolling-upgrade, storage failure,
   stale Kubernetes Lease holder, and regional restore/failover drills with
   measured RTO/RPO and no duplicate on-chain writes.
5. Backup key custody, off-cluster retention/restore, cloud IAM, and independent
   reconciliation against chain and billing data.

Until those records exist, the deployment is locally engineered and fail-closed,
not a certified production validator, HSM, storage, or DR profile.
