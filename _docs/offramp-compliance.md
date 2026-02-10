# Off-Ramp Compliance and Operational Workflows

VE-5E: Fiat Off-Ramp Integration (PayPal/ACH) Compliance Documentation

## Overview

This document describes the compliance requirements, operational workflows, and best practices for the VirtEngine fiat off-ramp integration. The off-ramp enables users to convert tokens to fiat currency (USD) via PayPal or ACH bank transfers.

## Table of Contents

1. [Regulatory Framework](#regulatory-framework)
2. [KYC Requirements](#kyc-requirements)
3. [AML Screening](#aml-screening)
4. [Payout Limits](#payout-limits)
5. [Operational Workflows](#operational-workflows)
6. [Reconciliation Process](#reconciliation-process)
7. [Incident Response](#incident-response)
8. [Audit Trail](#audit-trail)
9. [Provider Configuration](#provider-configuration)

## Regulatory Framework

### Applicable Regulations

The off-ramp integration must comply with:

- **Bank Secrecy Act (BSA)** - Anti-money laundering requirements
- **USA PATRIOT Act** - Customer identification program (CIP) requirements
- **FinCEN Requirements** - Suspicious activity reporting (SAR)
- **OFAC Sanctions** - Prohibited party screening
- **State Money Transmitter Laws** - Licensing requirements by state

### Licensing Considerations

VirtEngine operates as a technology provider. Actual money transmission is handled by licensed providers:

- **PayPal**: Licensed money transmitter in all US states
- **Stripe Treasury (ACH)**: Partner bank-based ACH origination

## KYC Requirements

### Identity Verification Tiers

The off-ramp uses tiered KYC based on transaction volumes:

| Tier | Daily Limit | Monthly Limit | Requirements |
|------|-------------|---------------|--------------|
| Basic | $1,000 | $5,000 | Email, phone verification |
| Standard | $10,000 | $50,000 | Government ID, address |
| Enhanced | $100,000 | $500,000 | Full KYC + enhanced due diligence |

### Required Documents

**Standard Verification:**
- Government-issued photo ID (passport, driver's license, state ID)
- Proof of address (utility bill, bank statement dated within 90 days)
- Selfie with liveness detection

**Enhanced Verification (>$10,000/day):**
- All standard documents
- Source of funds documentation
- Additional identity verification

### VEID Integration

The off-ramp integrates with VirtEngine Identity (VEID) for verification:

```go
// KYC check is required before payout approval
result, err := service.CheckPayoutEligibility(ctx, userID)
if !result.Eligible {
    // User must complete KYC before proceeding
    return ErrKYCNotVerified
}
```

### KYC Status Codes

| Status | Description | Payout Allowed |
|--------|-------------|----------------|
| `verified` | Full verification complete | Yes |
| `pending` | Verification in progress | No |
| `rejected` | Verification failed | No |
| `expired` | Re-verification required | No |
| `review` | Manual review required | No |

## AML Screening

### Screening Requirements

All payouts are screened against:

1. **OFAC SDN List** - Specially Designated Nationals
2. **OFAC Consolidated List** - All OFAC sanctions lists
3. **PEP Lists** - Politically Exposed Persons
4. **Adverse Media** - Negative news screening
5. **Internal Watchlists** - VirtEngine risk indicators

### Risk Scoring

Each payout receives a risk score from 0-100:

| Score Range | Risk Level | Action |
|-------------|------------|--------|
| 0-30 | Low | Auto-approve |
| 31-60 | Medium | Enhanced monitoring |
| 61-80 | High | Manual review required |
| 81-100 | Critical | Block, file SAR if applicable |

### Transaction Patterns

The following patterns trigger enhanced review:

- Rapid successive payouts (structuring detection)
- Amounts just below reporting thresholds ($10,000)
- Inconsistent transaction patterns
- New account with high-value transactions
- Payouts to high-risk jurisdictions

### SAR Filing

Suspicious Activity Reports are filed when:

- AML score exceeds critical threshold (81+)
- Pattern analysis indicates potential structuring
- Transaction involves sanctioned entities
- Manual review identifies suspicious behavior

## Payout Limits

### System Limits

```yaml
limits:
  # Per-transaction limits
  min_payout_usd: 1.00
  max_payout_usd: 100,000.00
  
  # Volume limits
  daily_limit_usd: 1,000,000.00
  monthly_limit_usd: 5,000,000.00
  
  # KYC-based limits (per user)
  basic_daily_limit: 1,000.00
  standard_daily_limit: 10,000.00
  enhanced_daily_limit: 100,000.00
```

### Rate Limiting

- Maximum 10 payout requests per minute per user
- Maximum 100 payout requests per hour per user
- Burst protection for high-volume periods

## Operational Workflows

### Payout Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Quote     │────▶│  KYC/AML    │────▶│  Provider   │
│  Creation   │     │   Check     │     │ Submission  │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                    ┌─────────────┐           │
                    │  Completed  │◀──────────┤
                    │  /Failed    │           ▼
                    └─────────────┘     ┌─────────────┐
                           ▲            │  Webhook    │
                           └────────────│  Update     │
                                        └─────────────┘
```

### Standard Payout Flow

1. **Quote Request**
   - User requests conversion quote
   - System calculates exchange rate and fees
   - Quote valid for 60 seconds

2. **Eligibility Check**
   - Verify KYC status
   - Check daily/monthly limits
   - Run AML pre-screening

3. **Payout Creation**
   - Lock tokens in escrow
   - Create payout intent
   - Full AML screening

4. **Provider Submission**
   - Submit to PayPal/ACH provider
   - Retry with exponential backoff on failure
   - Maximum 3 retry attempts

5. **Status Updates**
   - Receive webhook notifications
   - Update payout status
   - Trigger reconciliation

6. **Completion**
   - Mark payout as completed/failed
   - Release or refund escrowed tokens
   - Update user limits

### Manual Review Process

Payouts flagged for manual review require:

1. Compliance officer assignment
2. Document verification
3. Source of funds check (if applicable)
4. Approval/rejection decision
5. Documentation of decision rationale

### Escalation Procedures

| Trigger | Escalation Level | Response Time |
|---------|------------------|---------------|
| AML high-risk | Compliance Team | 4 hours |
| SAR required | Compliance Officer | 24 hours |
| Sanctions match | Legal + Compliance | Immediate |
| Pattern alert | Risk Analyst | Same day |

## Reconciliation Process

### Daily Reconciliation

The reconciliation job runs automatically and:

1. Fetches settlement reports from providers
2. Matches on-chain payout records
3. Identifies discrepancies
4. Auto-resolves minor differences (<1%)
5. Flags major discrepancies for review

### Reconciliation States

| State | Description |
|-------|-------------|
| `matched` | On-chain and provider records match |
| `mismatch` | Discrepancy detected, requires review |
| `pending` | Awaiting provider settlement |
| `resolved` | Discrepancy resolved manually |

### Discrepancy Handling

```yaml
reconciliation:
  # Auto-resolve discrepancies below this threshold
  auto_resolve_threshold_cents: 100  # $1.00
  
  # Alert threshold for mismatches
  alert_threshold_percent: 1.0  # 1%
  
  # Time to wait before flagging pending
  settlement_timeout_hours: 72
```

## Incident Response

### Alert Categories

| Category | Severity | Response |
|----------|----------|----------|
| Provider Outage | Critical | Failover, pause payouts |
| High Failure Rate | Warning | Investigate, notify ops |
| Reconciliation Mismatch | Warning | Manual review |
| Sanctions Match | Critical | Immediate block, notify legal |
| Unusual Volume | Info | Monitor, investigate if persistent |

### Runbook: Provider Outage

1. Confirm outage via provider status page
2. Pause new payout submissions
3. Enable backup provider if available
4. Communicate to affected users
5. Resume when provider recovers
6. Reconcile any failed transactions

### Runbook: Suspicious Activity

1. Block user payouts immediately
2. Freeze related accounts
3. Gather transaction history
4. Conduct internal investigation
5. File SAR if warranted (within 30 days)
6. Document all actions taken

## Audit Trail

### Logged Events

All payout operations are logged:

- Quote creation and expiration
- KYC/AML check results
- Payout creation and status changes
- Provider API calls and responses
- Webhook receipts
- Reconciliation results
- Manual review decisions

### Retention Periods

| Data Type | Retention Period |
|-----------|------------------|
| Transaction records | 7 years |
| KYC documents | 5 years after relationship ends |
| AML screening results | 5 years |
| Audit logs | 7 years |
| SARs | 5 years |

### Access Controls

- Audit logs are append-only
- Access restricted to compliance role
- All access is logged
- Regular access reviews (quarterly)

## Provider Configuration

### PayPal Setup

```yaml
paypal:
  sandbox_url: "https://api-m.sandbox.paypal.com"
  production_url: "https://api-m.paypal.com"
  
  # OAuth credentials (from environment)
  client_id: ${PAYPAL_CLIENT_ID}
  client_secret: ${PAYPAL_CLIENT_SECRET}
  
  # Webhook configuration
  webhook_id: ${PAYPAL_WEBHOOK_ID}
  
  # Payout settings
  email_subject: "VirtEngine Payout"
  email_message: "Your payout from VirtEngine"
```

### ACH Setup (Stripe Treasury)

```yaml
ach:
  api_key: ${STRIPE_SECRET_KEY}
  
  # Treasury settings
  financial_account: ${STRIPE_FINANCIAL_ACCOUNT}
  
  # Webhook configuration
  webhook_secret: ${STRIPE_WEBHOOK_SECRET}
  
  # ACH settings
  statement_descriptor: "VIRTENGINE"
```

### Sandbox Testing

Before production deployment:

1. Configure sandbox credentials
2. Run integration test suite
3. Test all payout scenarios
4. Verify webhook handling
5. Test reconciliation with mock data
6. Validate error handling

### Production Checklist

- [ ] Production API credentials configured
- [ ] Webhook endpoints registered with providers
- [ ] TLS/SSL certificates valid
- [ ] Monitoring and alerting enabled
- [ ] Reconciliation job scheduled
- [ ] Runbooks reviewed by operations
- [ ] Compliance sign-off obtained
- [ ] Legal review completed

## Appendix

### Error Codes

| Code | Description | User Action |
|------|-------------|-------------|
| `kyc_not_verified` | KYC verification required | Complete identity verification |
| `kyc_expired` | KYC verification expired | Re-verify identity |
| `aml_check_failed` | AML screening failed | Contact support |
| `daily_limit_exceeded` | Daily limit reached | Wait until tomorrow |
| `monthly_limit_exceeded` | Monthly limit reached | Wait until next month |
| `min_amount_not_met` | Below minimum payout | Increase payout amount |
| `max_amount_exceeded` | Above maximum payout | Decrease payout amount |
| `provider_error` | Provider API error | Retry later |
| `quote_expired` | Quote has expired | Request new quote |

### Contact Information

- **Compliance Team**: compliance@virtengine.io
- **Operations**: ops@virtengine.io
- **Security**: security@virtengine.io
- **Legal**: legal@virtengine.io
