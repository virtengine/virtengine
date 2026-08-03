# T4-07F SLURM Chart Retirement Evidence

Status: partial, dependency-blocked

The competing `_build/helm/slurm-cluster` chart was removed. Repository build,
workflow, deployment, and runtime entry points did not reference it. The only
remaining canonical authoring source is `deploy/slurm/slurm-cluster`.

The discovery validator rejects reintroduction of an unknown or retired SLURM
chart. This retirement does not complete Task 88C: semantic Helm render,
capacity equality, least-privilege and tenant isolation, live durable-state
recovery, and dependency evidence remain unavailable.