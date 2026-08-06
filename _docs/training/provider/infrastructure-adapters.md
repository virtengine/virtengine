# Infrastructure Adapters: Provider Daemon Training

## Purpose
This document explains which provider-daemon adapter paths are launch-critical, what each path must preserve for Waldur reconciliation, and what operators are expected to verify during live fulfillment.

## Supported adapter paths
The provider-daemon operational surface currently treats these as the supported launch paths:

- Kubernetes container workloads through `pkg/provider_daemon/kubernetes_adapter.go` and `pkg/provider_daemon/provisioning_kubernetes.go`
- SLURM-on-Kubernetes cluster bootstrap and lifecycle through `pkg/provider_daemon/slurm_k8s/**`

Legacy cloud-specific adapters remain code in the repo, but they are not the launch-critical provider-daemon operations path and are not covered by the cluster-backed provider daemon suites in `tests/integration/provider/**` and `tests/e2e/provider_daemon/**`.

## Adapter contract
Every supported adapter path must preserve these fields across provisioning, lifecycle, and settlement:

- Waldur offering UUID
- Waldur order UUID
- Waldur resource UUID
- VirtEngine backend ID
- provider-owned operational attributes such as service endpoints, cluster names, queue names, or placement metadata

## Common adapter rules
- the Waldur `backend_id` remains the reconciliation key across all supported adapters
- resource UUID is the key for lifecycle and usage actions
- usage component names must stay canonical:
  - `cpu_hours`
  - `gpu_hours`
  - `ram_gb_hours`
  - `storage_gb_hours`
  - `network_gb`
- retries must be idempotent at the provider boundary; a repeated bridge action must reconcile existing provider state rather than silently duplicating it

## Kubernetes container workloads

### Verified runtime behavior
- the provider daemon creates one namespace per workload and applies deployments, services, PVCs, and network policies inside that namespace
- readiness is derived from real pod and container status, not from a cached deployment record
- a repeated provision request reconciles the existing workload status
- failed, stopped, or terminated workloads are re-applied on the next provisioner pass
- namespace deletion is the cleanup boundary for termination

### Operator checks
- workload namespace exists and is labeled for the expected deployment and lease IDs
- service endpoints resolve from the created Kubernetes Service objects
- pod readiness transitions the workload to `running`
- terminated containers or failed pods transition the workload to `failed`
- if a workload is in `failed`, fix the root cause and rerun provisioning or redeploy before requesting cleanup

### Failure boundary
- the Kubernetes adapter intentionally rejects direct termination from the `failed` state
- operator recovery is:
  1. inspect pod and container failure messages
  2. correct the workload or cluster issue
  3. rerun provisioning so the workload returns to `running`
  4. terminate only after the workload is back on a valid lifecycle path

## SLURM-on-Kubernetes

### Verified runtime behavior
- bootstrap creates controller, database, and compute StatefulSets in the target namespace
- readiness depends on controller readiness, database readiness, compute replica readiness, and a successful `scontrol ping`
- cluster reconcile can resume nodes stuck in `down` when the controller is healthy and compute pods are already present
- cleanup is a Helm uninstall of the SLURM release

### Operator checks
- the Helm release name and namespace match the provider-daemon configuration
- `slurmctld`, `slurmdbd`, and compute StatefulSets report the expected ready replicas
- node state reported by `sinfo` lines up with the compute pod readiness the cluster currently has
- capacity evidence can be regenerated from the live `sinfo` output used by the adapter

### Failure boundary
- if the controller is not ready, reconcile degrades the cluster and waits rather than forcing node state changes
- if bootstrap readiness fails with rollback enabled, the release is expected to be removed before retry
- operator recovery is:
  1. restore controller and database readiness
  2. confirm compute StatefulSets are recreated and ready
  3. verify `scontrol ping` succeeds
  4. rerun bootstrap or reconcile only after the failing component is healthy

## Adapter acceptance checklist
- unique provider `backend_id`
- resource UUID persisted after provisioning
- lifecycle actions work through Waldur or the provider-daemon adapter, not manual side channels
- canonical usage submission succeeds
- partial usage failures are visible and actionable
- cleanup removes the exact workload or release that was provisioned
- operator docs and drills match the passing provider-daemon suites
