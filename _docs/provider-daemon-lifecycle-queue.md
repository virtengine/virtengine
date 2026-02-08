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
- **Dead-lettering:** Commands exceeding max retries are marked dead-lettered and surfaced via metrics for operator inspection.
