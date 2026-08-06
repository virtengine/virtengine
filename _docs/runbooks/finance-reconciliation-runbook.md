# Finance Reconciliation Runbook

This runbook defines the exact evidence finance signs when a treasury-funded provider payout is executed through the off-ramp bridge. It is intentionally limited to commands and artifacts that exist in the current codebase.

## Objective

Finance sign-off is complete only when the same settlement can be traced across:

1. The bridge contract tests in `pkg/payments/offramp`
2. The settlement integration tests in `tests/integration/settlement`
3. Optional chain spot-checks using supported settlement queries

The authoritative evidence payload is the `finance-evidence=` JSON line emitted by `TestFiatConversionPipelineSuccess` and `TestFiatConversionReconciliation`.

## Evidence Sources

The sign-off packet must include these artifacts for every reconciliation run:

| Artifact | Source | Purpose |
|----------|--------|---------|
| `bridge-unit.log` | `go test ./pkg/payments/offramp` | Proves bridge quote, metadata, status, and retry behavior |
| `bridge-e2e.log` | `go test -tags "e2e.integration" ./pkg/payments/offramp` | Proves the bridge lifecycle remains valid under the package e2e contract |
| `settlement-offramp.log` | `go test -tags "e2e.integration" ./tests/integration/settlement -run 'TestFiatConversionPipelineSuccess|TestFiatConversionReconciliation' -v` | Emits the finance sign-off records |
| `finance-evidence.jsonl` | extracted from `settlement-offramp.log` | Machine-readable approval record |
| `payouts-by-provider.json` | `virtengined query settlement payouts --provider ...` | Optional chain spot-check for the provider payout record |
| `fiat-conversion-<id>.json` | `virtengined query settlement fiat-conversion <conversion-id>` | Optional chain spot-check for conversion state and off-ramp fields |

## Operator Sequence

### 1. Run the bridge validation suite

```bash
go test ./pkg/payments/offramp -count=1 | tee bridge-unit.log
go test -tags "e2e.integration" ./pkg/payments/offramp -count=1 | tee bridge-e2e.log
```

Use PowerShell if you are on Windows:

```powershell
go test ./pkg/payments/offramp -count=1 | Tee-Object -FilePath bridge-unit.log
go test -tags "e2e.integration" ./pkg/payments/offramp -count=1 | Tee-Object -FilePath bridge-e2e.log
```

### 2. Generate the reconciliation evidence

```bash
go test -tags "e2e.integration" ./tests/integration/settlement \
  -run 'TestFiatConversionPipelineSuccess|TestFiatConversionReconciliation' \
  -count=1 -v | tee settlement-offramp.log
```

PowerShell equivalent:

```powershell
go test -tags "e2e.integration" ./tests/integration/settlement `
  -run 'TestFiatConversionPipelineSuccess|TestFiatConversionReconciliation' `
  -count=1 -v | Tee-Object -FilePath settlement-offramp.log
```

### 3. Extract the sign-off packet

```bash
rg 'finance-evidence=' settlement-offramp.log | sed 's/^.*finance-evidence=//' > finance-evidence.jsonl
```

PowerShell equivalent:

```powershell
Select-String 'finance-evidence=' settlement-offramp.log |
  ForEach-Object { $_.Line.Substring($_.Line.IndexOf('finance-evidence=') + 17) } |
  Set-Content finance-evidence.jsonl
```

The extraction must produce one line for `TestFiatConversionPipelineSuccess` and one line for `TestFiatConversionReconciliation`.

### 4. Optional chain spot-checks

Use the provider address from the evidence record.

```bash
virtengined query settlement payouts --provider <provider-address> --output json > payouts-by-provider.json
virtengined query settlement fiat-conversion <conversion-id> --output json > fiat-conversion-<conversion-id>.json
```

These queries are a spot-check only. Finance sign-off still depends on the test-backed evidence packet above.

## Required Evidence Fields

Every `finance-evidence` record must contain these fields:

| Field | Meaning |
|-------|---------|
| `provider` | Provider receiving the payout |
| `invoice_id` | Invoice linked to the payout |
| `settlement_id` | Settlement being reconciled |
| `payout_id` | Settlement payout record identifier |
| `payout_state` | Expected terminal payout state |
| `payout_tx_hash` | Deterministic payout execution reference |
| `payout_idempotency_key` | Duplicate-prevention key for payout execution |
| `payout_ledger_entry_types` | Ledger entry sequence observed during payout handling |
| `treasury_balance` | Treasury balance after payout processing |
| `expected_treasury_balance` | Expected treasury balance from platform plus validator fees |
| `conversion_id` | Fiat conversion record identifier |
| `conversion_state` | Fiat conversion terminal state |
| `conversion_idempotency_key` | Duplicate-prevention key for off-ramp submission |
| `off_ramp_provider` | Selected bridge adapter |
| `off_ramp_quote_id` | Quote accepted by the bridge |
| `off_ramp_id` | Provider payout identifier |
| `off_ramp_status` | Provider-reported payout status |
| `off_ramp_reference` | Provider payout reference |
| `bridge_status` | Status returned by the bridge |
| `bridge_reference` | Bridge-visible provider reference |
| `bridge_quote_id` | Bridge-visible quote identifier |
| `conversion_audit_actions` | Ordered conversion audit trail |
| `transition_count` | Number of recorded conversion transitions |

## Finance Sign-Off Rules

Finance signs the packet only when all of these are true:

- `payout_state` is `completed`
- `conversion_state` is `payout_completed`
- `off_ramp_status` is `completed`
- `bridge_status` is `completed`
- `payout_tx_hash` equals the deterministic `fiat-<conversion_id>` reference
- `treasury_balance` matches `expected_treasury_balance`
- `payout_ledger_entry_types` contains `completed`
- `conversion_idempotency_key` and `payout_idempotency_key` are both present
- `off_ramp_reference` matches `bridge_reference`
- `off_ramp_quote_id` matches `bridge_quote_id`
- both evidence lines are present: one immediate-completion flow and one reconciliation-after-processing flow

## Stop Conditions

Do not sign and escalate to engineering if any of these occur:

- `finance-evidence.jsonl` is missing or contains fewer than two records
- any record contains `processing`, `failed`, `cancelled`, or `refunded` in the final bridge or conversion status fields
- `treasury_balance` differs from `expected_treasury_balance`
- the bridge quote ID, provider reference, or provider payout ID does not match the chain spot-check
- an idempotency key is blank
- the bridge suite or bridge e2e suite fails

## Audit Retention

Archive these files together for each reconciliation window:

- `bridge-unit.log`
- `bridge-e2e.log`
- `settlement-offramp.log`
- `finance-evidence.jsonl`
- any optional `payouts-by-provider.json`
- any optional `fiat-conversion-<id>.json`

The archive is the finance approval packet for treasury payout release and month-end reconciliation.
