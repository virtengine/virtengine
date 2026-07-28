# Kubernetes Manifests

`deploy/kubernetes` is the canonical application manifest tree for VirtEngine.
The `infra/kubernetes/base` and `infra/kubernetes/overlays/*` entries are
compatibility shims that import the canonical deploy tree. They must not define
their own application Deployments, StatefulSets, Services, ExternalSecrets, PVCs,
PDBs, HPAs, KEDA objects, or NetworkPolicies.

Chaos and DR resources under this directory are consumers of the canonical
application labels, services, and PVC names. They should be updated only when the
canonical render changes, and `node scripts/task85c-validate-kubernetes.mjs`
must pass before those changes are merged.

