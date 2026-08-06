# VirtEngine Disaster Recovery Automation

This directory contains the repo-owned backup, restore, and drill commands that
match the A18 infrastructure model.

## Contracts

- Primary environments are `dev`, `staging`, and `prod`.
- Multi-region DR validation targets `us-east-1`, `eu-west-1`, and `ap-southeast-1`.
- Public endpoint checks default to `virtengine.com`.
- Evidence-producing failover drills run through [`infra/dr/run-failover-drill.sh`](../../infra/dr/run-failover-drill.sh).

## Operational Scripts

| Script | Purpose | Output |
| --- | --- | --- |
| `backup-chain-state.sh` | Snapshot, verify, upload, and restore validator state | snapshot archive, checksum, metadata |
| `backup-provider-state.sh` | Snapshot, verify, upload, and restore provider state | provider backup archive, checksum, metadata |
| `backup-keys.sh` | Export encrypted validator/provider key bundles | encrypted key bundle, metadata |
| `dr-test.sh` | Validate backup freshness, remote access, DNS, RPC, and restore readiness | console summary, optional JSON report |
| `failover-test.sh` | Local/docker failover rehearsal for chain recovery flow | docker-based failover validation |

## Standard Operator Flow

1. Run the backup scripts on their normal cadence or manually for an incident:
   ```bash
   ./scripts/dr/backup-chain-state.sh
   ./scripts/dr/backup-provider-state.sh
   ./scripts/dr/backup-keys.sh --type all
   ```
2. Validate the environment contract:
   ```bash
   ./scripts/dr/dr-test.sh --environment staging --report
   ```
3. Run a rehearsal or live failover drill and capture evidence:
   ```bash
   ./infra/dr/run-failover-drill.sh --mode rehearsal --output-dir output/drill/rehearsal
   ./infra/dr/run-failover-drill.sh --mode live --live-validation --output-dir output/drill/live
   ```

## Environment Variables

| Variable | Purpose | Default |
| --- | --- | --- |
| `DR_BUCKET` | S3 bucket URL for backup upload and drill evidence | unset |
| `BASE_DOMAIN` | DNS suffix used by `dr-test.sh` endpoint checks | `virtengine.com` |
| `REGION_LIST` | Comma-separated regions used by `dr-test.sh` | `us-east-1,eu-west-1,ap-southeast-1` |
| `RESULTS_DIR` | Report output directory for `dr-test.sh` | `output/dr-tests` |
| `SLACK_WEBHOOK` | Slack notification endpoint for `dr-test.sh --notify` | unset |
| `SNAPSHOT_SIGNING_KEY` | Private key used to sign snapshot manifests | unset |
| `SNAPSHOT_VERIFY_PUBKEY` | Public key used to verify snapshot manifests | unset |

Backup-script-specific settings remain documented inline in each script header.

## Recovery Notes

- Restore validator state with `backup-chain-state.sh --restore <height>`.
- Restore provider state with `backup-provider-state.sh --restore <backup-id>`.
- Recover encrypted key bundles with `backup-keys.sh --restore <bundle>` plus the required passphrase or Shamir shares.
- For cluster failover, use `infra/dr/run-failover-drill.sh` as the canonical evidence-producing wrapper instead of invoking individual drill helpers ad hoc.
