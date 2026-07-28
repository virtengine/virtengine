# ADR-009: Canonical Kubernetes Rendering and Identity Separation

## Status

Accepted

## Date

2026-07-23

## Context

Task 85C found two deployment risks in the Kubernetes manifests:

- `deploy/kubernetes` and `infra/kubernetes` described separate application
  topologies, which allowed production, chaos, and DR manifests to drift.
- The chain node could be scaled as a validator while consuming one shared
  validator identity secret, creating a double-signing risk when replicas were
  greater than one.
- The provider daemon used ephemeral state for key material, WALDUR bridge
  state, order checkpoints, chain usage queues, provisioning state, audit logs,
  and fiat conversion state.

This ADR records the manifest-level decision. It does not certify a live
cluster, remote signer, TMKMS deployment, or HSM.

## Decision

`deploy/kubernetes` is the canonical application manifest tree. The
`infra/kubernetes/base` and `infra/kubernetes/overlays/*` entry points are
compatibility shims that import the matching canonical deploy root and must not
define application Deployments, StatefulSets, Services, ExternalSecrets, PVCs,
PDBs, HPAs, KEDA resources, or NetworkPolicies.

`infra/rollouts` is not an application topology source. It may retain optional
Argo Rollouts `AnalysisTemplate` policy, such as rollback health checks, but it
must not define `Rollout`, `Deployment`, `StatefulSet`, or `DaemonSet`
workloads and must never be applied wholesale as a production deployment tree.
The only supported direct production application path is
`kubectl apply -k deploy/kubernetes/overlays/prod`; ArgoCD must target that same
Kustomize root.

Production promotion changes each canonical image reference to an immutable
`name@sha256:<64-hex-digest>` value before the canonical overlay is applied.
Mutable tags and alternate rollout workload manifests are rejected by policy.

The chain topology is split into two workloads:

- `virtengine-validator`: one StatefulSet replica, one validator identity, and
  remote-signer/TMKMS metadata from `validator-signer-identity`.
- `virtengine-node`: horizontally scalable sentry/full nodes that expose RPC,
  REST, gRPC, and P2P services without validator signer metadata.

The validator manifest must fail closed when signer metadata is absent. The
Kubernetes pod receives validator address, key fingerprint, signer endpoint,
signer epoch, fencing token, TMKMS config digest, and signer certificate digest.
Raw validator private key material is not mounted into the pod.

The provider daemon runs as a StatefulSet. Every replica mounts the same
encrypted `provider-data` identity PVC and all shared HA files are routed through durable PVCs:
`provider-ha-state` and `provider-backups`. Provider manifests must include
anti-affinity, topology spread, PDBs, pod security context, readiness/liveness
checks, immutable image digests, and metadata for distributed lease/fencing.
Production uses `coordination.k8s.io/v1` Lease objects with optimistic
resource-version updates. `LeaseTransitions` is the monotonic fencing token;
standbys remain unready until ownership, and stale tokens cannot renew or
submit. The shared-file lease is retained only as an explicit compatible
profile for filesystems with verified shared locking; process-local leases are
forbidden in production.

Provider signing keys use an Argon2id-derived AES-256-GCM encrypted file
keystore with authenticated metadata and atomic replacement. The passphrase is
mounted as a secret file and never logged. Missing production keystores fail
closed instead of generating a replacement identity. Hardware, Ledger, and
non-custodial profiles fail closed until a real provider is integrated; the
in-repository SoftHSM implementation is development/test software and is not
hardware certification.

Chaos and DR resources are consumers of canonical resources. They target
canonical labels, services, and PVC names and do not define an alternate
application topology. Kubernetes DR backs up validator chain state but never
exports the remote signer private key, mTLS identity, anti-double-sign state or
fencing state; those remain the responsibility of the independent TMKMS/HSM
custody system. Provider snapshots use the supported backup script contract and
must be signed through a separately mounted DR signing key.

## Consequences

- Scaling sentry/full nodes no longer scales validator identity.
- Production render policy can reject mutable images, shared validator secrets,
  missing provider persistence, and divergent infra topology before cluster
  application.
- Stale blue/green Rollout manifests cannot reintroduce a second production
  topology beside the canonical StatefulSet rendering.
- DR jobs reference canonical PVC names while independent signer custody owns
  signer metadata, anti-double-sign state, fencing continuity and key recovery.
- Existing cluster-specific overlays that need IRSA, signer network CIDRs, or
  real storage classes must patch the canonical deploy tree instead of creating
  an infra-local application base.

## Validation

The infrastructure CI workflow enforces this ADR for the paths that can change
the canonical render contract: `infra/**`, `deploy/**`, the Task 85C validator,
`docs/documentation/INFRA-001-IMPLEMENTATION-SUMMARY.md`, this ADR, the Task
85C completion report, the Kubernetes identity recovery runbook, and the
workflow file. The workflow `validate` job installs Node 20 and checksum-pinned
kubectl 1.29.0, runs the Task 85C validator, and must pass before security,
plan, apply, or drift jobs. That CI gate uses read-only repository permissions
and no deployment secrets.

Run:

```bash
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/prod
kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/prod
node scripts/task85c-validate-kubernetes.mjs
node scripts/validate-agents-docs.mjs
```

The Task 85C validator renders every canonical and compatibility entry point,
compares infra render output to deploy render output, and enforces production
identity, persistence, HA, immutable-image invariants, infra shim boundaries,
and the policy-only scope of `infra/rollouts`. It scans the rollout files
themselves for forbidden workload kinds and mutable image references, and scans
deployment documentation for deleted manifest paths, wholesale rollout apply
commands, and obsolete validator deployment instructions. Provider metrics
port enforcement is scoped to the canonical provider metrics role rather than
prohibiting otherwise legitimate use of a port by unrelated workloads.
