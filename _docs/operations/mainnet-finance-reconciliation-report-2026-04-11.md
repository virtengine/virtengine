# Mainnet Finance Reconciliation Report - 2026-04-11

Last updated: 2026-04-11
Owner: Finance Lead

## Scope
Repository-backed launch rehearsal evidence for checklist item `6E`: billing
reconciliation, payout execution, and dispute-path coverage.

## Commands

```bash
go test ./pkg/payments/offramp -count=1 -v
go test -tags="e2e.integration" ./pkg/payments/offramp -count=1 -v
go test -tags="e2e.integration" ./tests/integration/settlement \
  -run "TestFiatConversionPipelineSuccess|TestFiatConversionReconciliation|TestSettlementPayoutOffRampFlow|TestDisputeArbitrationRefundFlow|TestFullPipelineTestSuite" \
  -count=1 -v
```

## Evidence artifacts

| Artifact | Location | Completed (UTC) | SHA-256 |
| --- | --- | --- | --- |
| Off-ramp bridge unit log | `output/mainnet-launch/2026-04-11/finance-bridge-unit.log` | 2026-04-11 05:48:16 | `7466be2ca692d2a7f62c19ea42a1170464039cfae75239538959a7882f80e766` |
| Off-ramp bridge e2e log | `output/mainnet-launch/2026-04-11/finance-bridge-e2e.log` | 2026-04-11 05:48:16 | `9b4284411a46252ce3d2cc5c859ff1874d22ca24b78edce1b3687d86cdd17b83` |
| Settlement integration log | `output/mainnet-launch/2026-04-11/finance-settlement.log` | 2026-04-11 05:49:00 | `52204a1cf3310a5635ca0fe6895c31883380194755d162ee9fd7881d16f57add` |
| Extracted finance evidence rows | `output/mainnet-launch/2026-04-11/finance-evidence.jsonl` | 2026-04-11 05:55:40 | `e77e10735cbcf3dea9cb9c89a00511a8cc54ff6520431e2dbee58e1823dd3a33` |

## Result
- Status: `PASS`
- Launch relevance: satisfies the finance reconciliation prerequisite for the
  rehearsal bundle and launch packet.

## Reconciliation summary
- Two finance evidence rows were extracted from the settlement integration log.
- Both rows recorded:
  - `payout_state = completed`
  - `conversion_state = payout_completed`
  - `bridge_status = completed`
  - `treasury_balance = expected_treasury_balance = 60uve`
- The settlement rehearsal also passed:
  - dispute-resolution adjusted payout coverage
  - escrow expiry auto-settlement
  - multi-order settlement batch
  - partial refund
  - escrow invariant validation
  - usage-to-invoice-to-payout flow

## Source references
- `pkg/payments/offramp`
- `tests/integration/settlement`
- `output/mainnet-launch/2026-04-11/finance-settlement.log`
- `output/mainnet-launch/2026-04-11/finance-evidence.jsonl`
