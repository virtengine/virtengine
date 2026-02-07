# Treasury Operations Runbook

## Purpose
This runbook documents treasury-grade exchange routing, custody controls, FX snapshotting, and reconciliation for VE token conversions and multi-currency holdings.

## Scope
- DEX + CEX conversion routing with best-execution rules
- Hot/cold wallet custody controls and rotation cadence
- Withdrawal policy enforcement and multi-sig approvals
- FX rate snapshots for invoicing and settlements
- Ledger reconciliation and drift detection

## Roles & Responsibilities
- **Treasury Ops:** Executes conversions, monitors balances, initiates withdrawals
- **Finance Lead:** Approves large withdrawals, reviews reconciliation reports
- **Security Lead:** Maintains custody policies, rotation schedules, key ceremonies

## Custody Controls
### Wallet Separation
- **Hot wallet:** Operational liquidity only (daily withdrawal cap enforced)
- **Cold wallet:** Long-term custody, requires multi-sig approvals

### Rotation Schedule
- Hot wallet: rotate keys every 30 days or immediately after incident
- Cold wallet: rotate keys quarterly or after custody ceremony
- Log every rotation with reason and sign-off

### Withdrawal Policy
- Enforce per-transaction and daily hot limits
- Blocklisted destinations are denied automatically
- Allowlist enforcement for high-risk corridors
- Multi-sig required above threshold and for all cold wallet withdrawals

## Conversion Execution
### Best Execution Rules
- Collect quotes from DEX + CEX adapters
- Filter on max slippage and fee caps
- Select highest net output (after fees)
- Execute and record conversion ledger entry

### Failure Handling
- If no valid quote, halt execution and alert Treasury Ops
- Retry with alternate adapter when possible
- Escalate to manual execution for critical settlements

## FX Rate Snapshots
- Capture FX snapshots at conversion time and end-of-day
- Store historical snapshots for invoicing and settlements
- Use historical snapshot at settlement timestamp for audit accuracy

## Reconciliation
1. Export ledger balances for all treasury assets
2. Compare with on-chain/custody balances
3. Investigate drift; open incident if > 0.5% variance
4. Record corrective adjustment entry

## Monitoring & Alerts
- Balance thresholds (hot liquidity minimum)
- Conversion failure rates
- Unauthorized withdrawal attempts
- Rotation overdue alerts

## Incident Response
- Freeze withdrawals on custody anomaly
- Rotate hot wallets immediately on suspected compromise
- Notify Security Lead and Finance Lead within 1 hour
- Document incident in audit log

## References
- _docs/runbooks/finance-reconciliation-runbook.md
- _docs/key-management.md
