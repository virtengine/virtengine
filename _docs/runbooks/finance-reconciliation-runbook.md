# Finance Reconciliation Runbook

This runbook describes the procedures for reconciling payouts, invoices, and escrow settlements in VirtEngine.

## Overview

The settlement module handles three key financial flows:
1. **Invoice Processing** - Converting usage into billable invoices
2. **Escrow Settlement** - Releasing funds from escrow upon completion
3. **Provider Payouts** - Transferring funds to providers after deducting fees

## Key Concepts

### Payout States

| State | Description |
|-------|-------------|
| `pending` | Payout created, awaiting execution |
| `processing` | Payout execution in progress |
| `completed` | Funds successfully transferred |
| `failed` | Transfer failed (will be retried) |
| `held` | Payout frozen due to dispute |
| `refunded` | Funds returned to customer |
| `cancelled` | Payout cancelled before execution |

### Fee Structure

- **Platform Fee**: Percentage of gross amount (default 5%)
- **Validator Fee**: Percentage of gross amount (default 1%)
- **Holdback**: Optional percentage reserved for disputes (default 0%)
- **Net Amount**: Gross - Platform Fee - Validator Fee - Holdback

## Daily Reconciliation Procedure

### Step 1: Generate Reconciliation Report

```bash
# Query all payouts for a specific date
virtengined query settlement reconciliation-report \
  --start-time 2024-01-15T00:00:00Z \
  --end-time 2024-01-16T00:00:00Z \
  --output json > daily_report.json
```

### Step 2: Verify Invoice-to-Payout Matching

For each paid invoice, verify a corresponding payout record exists:

```bash
# List all paid invoices
virtengined query escrow invoices --status paid --output json

# Verify payout exists for invoice
virtengined query settlement payout-by-invoice <invoice_id>
```

### Step 3: Check Failed Payouts

```bash
# List failed payouts requiring attention
virtengined query settlement payouts --state failed --output json

# Retry failed payout (if within retry limit)
virtengined tx settlement retry-payout <payout_id> \
  --from <operator_key>
```

### Step 4: Verify Treasury Balances

```bash
# Check module account balance
virtengined query bank balance settlement

# Verify treasury records match
virtengined query settlement treasury-summary \
  --start-time 2024-01-15T00:00:00Z \
  --end-time 2024-01-16T00:00:00Z
```

## Dispute Handling

### When a Dispute is Opened

1. The payout is automatically placed in `held` state
2. Funds remain in the settlement module account
3. No further processing occurs until resolution

### Dispute Resolution Outcomes

| Resolution | Action |
|------------|--------|
| Provider Win | Payout released, funds transferred to provider |
| Customer Win | Payout refunded, funds returned to customer |
| Partial Settlement | Proportional payout/refund based on ruling |

### Processing Dispute Resolution

```bash
# After dispute resolved in provider's favor
virtengined tx settlement release-payout-hold <payout_id> \
  --from <operator_key>

# After dispute resolved in customer's favor
virtengined tx settlement refund-payout <payout_id> \
  --from <operator_key>
```

## Idempotency and Duplicate Prevention

Each payout has a unique idempotency key: `payout-{invoiceID}-{settlementID}`

If a duplicate payout attempt is made:
- The existing payout record is returned
- No duplicate transfer occurs
- Event `payout_idempotent_skip` is emitted

### Checking for Duplicates

```bash
# Query by idempotency key
virtengined query settlement payout-by-key \
  "payout-inv-12345-settle-67890"
```

## Monthly Reconciliation Checklist

- [ ] Verify total platform fees collected match treasury records
- [ ] Verify total validator fees distributed match staking records
- [ ] Ensure all completed payouts have valid transaction hashes
- [ ] Verify no payouts stuck in `processing` state > 24 hours
- [ ] Review and resolve all `held` payouts > 30 days
- [ ] Generate and archive monthly reconciliation report

## Troubleshooting

### Payout Stuck in Processing

If a payout remains in `processing` for extended time:

1. Check transaction status on-chain
2. If transaction failed but payout not updated:
   ```bash
   virtengined tx settlement mark-payout-failed <payout_id> \
     --error "transaction timeout" \
     --from <operator_key>
   ```

### Missing Payout for Paid Invoice

1. Verify invoice status is actually `paid`
2. Check for invoice hooks execution in logs
3. Manually trigger payout if needed:
   ```bash
   virtengined tx settlement execute-payout \
     --invoice-id <invoice_id> \
     --from <operator_key>
   ```

### Balance Mismatch

If settlement module balance doesn't match expected:

1. Export all ledger entries for the period
2. Calculate expected balance from entries
3. Identify missing or duplicate entries
4. Report discrepancy to development team

## Events for Monitoring

Configure alerting on these events:

| Event | Alert Level | Description |
|-------|-------------|-------------|
| `payout_failed` | Warning | Payout execution failed |
| `payout_held` | Info | Dispute halted a payout |
| `payout_completed` | Debug | Successful payout |
| `treasury_withdrawal` | Critical | Manual treasury withdrawal |

## Contact

- **Finance Operations**: finance-ops@virtengine.io
- **On-call Engineering**: oncall@virtengine.io
- **Escalation**: settlement-escalation@virtengine.io
