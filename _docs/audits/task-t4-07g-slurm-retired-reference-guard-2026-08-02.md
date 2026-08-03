# T4-07G Retired SLURM Reference Guard Evidence

Status: partial, dependency-blocked

The aggregate prototype integration validator now scans tracked operational,
build, release, and runtime surfaces for references to every retired SLURM
chart path. It fails closed when a reference is found. Historical documentation
and test fixtures are excluded so they can retain retirement evidence and
negative cases without becoming consumers.

The focused validator includes a negative runtime-reference case and the real
repository scan reports no operational reference to `_build/helm/slurm-cluster`.
This closes the source-drift guard only. Task 88C remains blocked on semantic
render, live capacity, tenant isolation, failover/restore, dependency, and
signed-accounting evidence.