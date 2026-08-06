# Provider Daemon Lifecycle Queue (VE-34E)

This document describes the durable lifecycle command queue used by the provider daemon to execute lifecycle actions (start/stop/resize/terminate) safely across restarts.

## Overview

The lifecycle command queue persists action requests to disk (Badger-backed by default), retries transient failures with exponential backoff, and reconciles desired state drift with Waldur resource state.

Key behaviors:

- **Durable storage:** lifecycle commands are persisted in `data/lifecycle_queue` (configurable).
- **Idempotent execution:** commands include an idempotency key derived from chain operation IDs to prevent duplicates.
- **Retry policy:** exponential backoff with a configurable max retry count before dead-lettering.
- **Reconciliation loop:** periodically checks desired allocation state vs. Waldur resource state and reissues safe commands when drift is detected.
- **Crash recovery:** pending/executing commands are requeued after restart; stale executions are reissued.

## Adapter safety boundaries

The lifecycle queue is intentionally conservative around adapter-specific state machines:

- Kubernetes container workloads can be retried safely for startup and readiness reconciliation because repeated provision attempts reuse the existing workload ID and namespace.
- Kubernetes cleanup is not forced from the `failed` state. If a workload has failed, the operator must repair the root cause and allow the provisioner to re-apply the workload before termination cleanup is retried.
- SLURM-on-Kubernetes reconcile will only resume down nodes when the controller is healthy and compute pods are already present. It will not fabricate a healthy cluster when the control plane is still broken.
- When `rollback_on_failure` is enabled for SLURM bootstrap, a readiness failure should leave no Helm release behind. Queue retries are only appropriate after the readiness blocker has been corrected.

## Configuration flags

All flags are exposed via `provider-daemon`:

- `--waldur-lifecycle-queue-enabled` (default: true)
- `--waldur-lifecycle-queue-backend` (default: badger)
- `--waldur-lifecycle-queue-path` (default: data/lifecycle_queue)
- `--waldur-lifecycle-queue-workers` (default: 2)
- `--waldur-lifecycle-queue-max-retries` (default: 5)
- `--waldur-lifecycle-queue-retry-backoff` (default: 10s)
- `--waldur-lifecycle-queue-max-backoff` (default: 5m)
- `--waldur-lifecycle-queue-poll-interval` (default: 2s)
- `--waldur-lifecycle-queue-reconcile-interval` (default: 5m)
- `--waldur-lifecycle-queue-reconcile-on-start` (default: true)
- `--waldur-lifecycle-queue-stale-after` (default: 20m)

## Metrics

Prometheus metrics emitted by the queue:

- `provider_daemon_lifecycle_queue_depth{status=...}` — queue depth by status
- `provider_daemon_lifecycle_queue_retries_total{action=...}` — retry count by action
- `provider_daemon_lifecycle_queue_commands_total{action=...,outcome=...}` — command outcomes
- `provider_daemon_lifecycle_reconcile_runs_total{outcome=...}` — reconciliation cycles
- `provider_daemon_lifecycle_reconcile_commands_total{action=...,outcome=...}` — reconcile outcomes

## Operational notes

- **Crash/restart safety:** Commands are stored before execution. On restart, pending or stale executing commands are requeued.
- **Drift handling:** If Waldur state diverges from the desired allocation state, the reconciler issues safe lifecycle actions (e.g., start/resume/stop/terminate).
- **Operator intervention point:** Reconcile will surface invalid transitions and degraded adapter states; it does not bypass provider-daemon safety checks to force cleanup or recovery.
- **Dead-lettering:** Commands exceeding max retries are marked dead-lettered and surfaced via metrics for operator inspection.
