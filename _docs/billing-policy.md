# VirtEngine Billing Policy

This document defines the billing schema, pricing rules, tax handling, and settlement hooks for VirtEngine marketplace services.

## Table of Contents

1. [Invoice Schema](#invoice-schema)
2. [Line Items](#line-items)
3. [Currency and Rounding](#currency-and-rounding)
4. [Pricing Rules](#pricing-rules)
5. [Discount Policies](#discount-policies)
6. [Tax Handling](#tax-handling)
7. [Settlement and Dispute Windows](#settlement-and-dispute-windows)
8. [For Providers](#for-providers)
9. [For Customers](#for-customers)
10. [Export Formats](#export-formats)
11. [Immutable Ledger](#immutable-ledger)
12. [Usage→Invoice→Settlement Pipeline](#usageinvoicesettlement-pipeline)

---

## Invoice Schema

### Invoice Lifecycle

```
Draft → Pending → [Paid | Partially Paid | Overdue | Disputed] → [Paid | Cancelled | Refunded]
```

### Invoice Fields

| Field | Type | Description |
|-------|------|-------------|
| `invoice_id` | string | Unique identifier (max 64 chars) |
| `invoice_number` | string | Human-readable number (e.g., `VE-00001234`) |
| `escrow_id` | string | Linked escrow account |
| `order_id` | string | Linked marketplace order |
| `lease_id` | string | Linked marketplace lease |
| `provider` | string | Provider's bech32 address |
| `customer` | string | Customer's bech32 address |
| `status` | enum | Current invoice status |
| `billing_period` | object | Billing period details |
| `line_items` | array | Individual charges |
| `subtotal` | Coins | Sum before adjustments |
| `discounts` | array | Applied discounts |
| `tax_details` | object | Tax calculations |
| `total` | Coins | Final amount due |
| `due_date` | timestamp | Payment deadline |

### Invoice Statuses

| Status | Description |
|--------|-------------|
| `draft` | Invoice being prepared, not yet issued |
| `pending` | Issued, awaiting payment |
| `paid` | Fully paid |
| `partially_paid` | Partial payment received |
| `overdue` | Past due date without full payment |
| `disputed` | Under dispute |
| `cancelled` | Voided invoice |
| `refunded` | Refunded after payment |

---

## Line Items

### Usage Types

| Type | Unit | Description |
|------|------|-------------|
| `cpu` | core-hour | CPU compute usage |
| `memory` | gb-hour | Memory usage |
| `storage` | gb-month | Persistent storage |
| `network` | gb | Network bandwidth |
| `gpu` | gpu-hour | GPU compute usage |
| `fixed` | unit | Fixed/flat charges |
| `setup` | unit | One-time setup fees |
| `other` | unit | Miscellaneous charges |

### Line Item Structure

```json
{
  "line_item_id": "line-001",
  "description": "CPU Usage - 4 cores for 720 hours",
  "usage_type": "cpu",
  "quantity": "2880",
  "unit": "core-hour",
  "unit_price": {
    "denom": "uvirt",
    "amount": "10000"
  },
  "amount": [{"denom": "uvirt", "amount": "28800000"}],
  "usage_record_ids": ["usage-001", "usage-002"],
  "pricing_tier": "standard"
}
```

---

## Currency and Rounding

### Supported Denominations

| Denomination | Precision | Description |
|--------------|-----------|-------------|
| `uvirt` | 6 | Micro-virt (1 VIRT = 1,000,000 uvirt) |
| `nvirt` | 9 | Nano-virt |
| `avirt` | 18 | Atto-virt |
| `uusd` | 6 | Stablecoin (if supported) |
| `ibc/*` | 6 | IBC tokens (default precision) |

### Rounding Rules

VirtEngine uses **Banker's Rounding** (half-to-even) by default:

| Value | Rounded |
|-------|---------|
| 1.5 | 2 |
| 2.5 | 2 |
| 3.5 | 4 |
| 4.5 | 4 |

**Rationale**: Banker's rounding eliminates systematic bias that occurs with traditional "round half up" methods, ensuring fair billing over many transactions.

### Rounding Configuration

Providers can configure rounding per pricing policy:

| Mode | Description |
|------|-------------|
| `half_even` | Banker's rounding (default) |
| `half_up` | Traditional rounding |
| `down` | Always truncate (floor) |
| `up` | Always round up (ceiling) |

### Minimum Charges

- **Default minimum charge**: 1000 uvirt (0.001 VIRT)
- **Per-resource minimums**: Configurable per usage type

---

## Pricing Rules

### Pricing Policy Structure

```json
{
  "policy_id": "provider-standard",
  "provider": "cosmos1provider...",
  "resource_pricing": {
    "cpu": {
      "base_rate": {"denom": "uvirt", "amount": "0.01"},
      "unit": "core-hour",
      "min_quantity": "0.1",
      "granularity_seconds": 3600
    },
    "memory": {
      "base_rate": {"denom": "uvirt", "amount": "0.005"},
      "unit": "gb-hour",
      "min_quantity": "0.1",
      "granularity_seconds": 3600
    }
  },
  "pricing_config": {
    "rounding_mode": "half_even",
    "minimum_charge": {"denom": "uvirt", "amount": "1000"},
    "default_payment_term_days": 7
  }
}
```

### Tiered Pricing

Volume-based pricing tiers allow providers to offer discounts at higher usage levels:

```json
{
  "tier_pricing": [
    {
      "tier_id": "tier-1",
      "tier_name": "Standard",
      "min_quantity": "0",
      "max_quantity": "1000",
      "discount_bps": 0
    },
    {
      "tier_id": "tier-2",
      "tier_name": "Volume",
      "min_quantity": "1000",
      "max_quantity": "10000",
      "discount_bps": 1000
    },
    {
      "tier_id": "tier-3",
      "tier_name": "Enterprise",
      "min_quantity": "10000",
      "max_quantity": "0",
      "discount_bps": 2000
    }
  ]
}
```

---

## Discount Policies

### Discount Types

| Type | Description |
|------|-------------|
| `percentage` | Percentage off (e.g., 10% off) |
| `fixed` | Fixed amount off (e.g., $10 off) |
| `volume` | Tiered by usage volume |
| `coupon` | Redeemable coupon code |
| `referral` | Referral program discount |
| `loyalty` | Loyalty points redemption |

### Coupon Codes

```json
{
  "code": "SAVE20",
  "discount_policy_id": "disc-20percent",
  "max_redemptions": 1000,
  "per_customer_limit": 1,
  "valid_from": "2026-01-01T00:00:00Z",
  "valid_until": "2026-12-31T23:59:59Z",
  "is_active": true
}
```

### Stacking Rules

- By default, discounts do not stack
- `stackable_with` field specifies compatible discount policies
- Maximum combined discount: 50% of subtotal (configurable)

### Loyalty Program

```json
{
  "program_id": "virt-rewards",
  "points_per_unit": "1",
  "redemption_rate": "0.001",
  "min_redemption_points": 1000,
  "max_redemption_percent_bps": 2500,
  "tiers": [
    {"tier_id": "bronze", "min_points": 0, "bonus_multiplier_bps": 10000},
    {"tier_id": "silver", "min_points": 10000, "bonus_multiplier_bps": 11000},
    {"tier_id": "gold", "min_points": 100000, "bonus_multiplier_bps": 12500}
  ]
}
```

---

## Tax Handling

### Jurisdiction Mapping

Tax jurisdictions are identified by ISO 3166-1 alpha-2 country codes:

| Country | Tax Type | Standard Rate |
|---------|----------|---------------|
| GB | VAT | 20% |
| DE | VAT | 19% |
| FR | VAT | 20% |
| SG | GST | 9% |
| AU | GST | 10% |
| US | None | 0% (federal) |

### Tax Exemption Categories

| Category | Description |
|----------|-------------|
| `none` | No exemption (default) |
| `b2b` | B2B reverse charge |
| `non_profit` | Non-profit organization |
| `government` | Government entity |
| `education` | Educational institution |
| `export` | Export exemption |
| `small_business` | Small business threshold |

### B2B Reverse Charge

For cross-border B2B transactions:
1. Customer provides valid Tax ID
2. Tax ID is verified
3. Reverse charge applies (0% tax on invoice)
4. Customer accounts for tax in their jurisdiction

### Tax Calculation

```
Subtotal: 100,000 uvirt
Discount: -10,000 uvirt
Taxable:   90,000 uvirt
VAT (20%): 18,000 uvirt
Total:    108,000 uvirt
```

### Customer Tax Profile

Customers can set their tax profile:

```json
{
  "customer": "cosmos1customer...",
  "country_code": "GB",
  "tax_id": "GB123456789",
  "tax_id_verified": true,
  "exemption_category": "b2b",
  "business_name": "Example Ltd",
  "is_b2b": true
}
```

---

## Settlement and Dispute Windows

### Default Configuration

| Parameter | Default | Range |
|-----------|---------|-------|
| Dispute Window | 7 days | 1-30 days |
| Escalation Timeout | 2 days | 1-7 days |
| Max Escalation Steps | 3 | 1-5 |
| Auto-settle after window | Yes | Yes/No |

### Dispute Lifecycle

```
Open → Under Review → [Resolved | Escalated] → Closed
                           ↓
                       Arbitration
```

### Dispute Resolutions

| Resolution | Description |
|------------|-------------|
| `provider_win` | Provider's charges upheld |
| `customer_win` | Full refund to customer |
| `partial_refund` | Partial refund agreed |
| `mutual_agreement` | Custom settlement |
| `arbitration` | Third-party decision |

### Settlement Hooks

| Hook | When | Purpose |
|------|------|---------|
| `pre_validation` | Before invoice validation | Verify usage records |
| `post_validation` | After validation | Log validation results |
| `pre_settlement` | Before payment transfer | Check escrow balance |
| `post_settlement` | After payment transfer | Emit events, notifications |
| `on_dispute` | Dispute initiated | Notify parties |
| `on_resolution` | Dispute resolved | Process refunds |

---

## For Providers

### Setting Up Pricing

1. **Create Pricing Policy**
   ```bash
   virtengine tx escrow create-pricing-policy \
     --currency uvirt \
     --cpu-rate 10000 \
     --memory-rate 5000 \
     --storage-rate 1000 \
     --from provider
   ```

2. **Configure Tax Profile**
   ```bash
   virtengine tx escrow set-provider-tax-profile \
     --country-code GB \
     --tax-id GB123456789 \
     --business-name "Provider Ltd" \
     --from provider
   ```

3. **Create Discount Policy** (optional)
   ```bash
   virtengine tx escrow create-discount-policy \
     --name "Volume Discount" \
     --type volume \
     --thresholds "1000:500,10000:1000,100000:2000" \
     --from provider
   ```

### Best Practices

- Set competitive but sustainable pricing
- Use tiered pricing for volume incentives
- Configure appropriate minimum charges
- Keep tax profiles up to date
- Monitor dispute rates

---

## For Customers

### Understanding Your Invoice

1. **Line Items**: Each resource type charged separately
2. **Subtotal**: Sum of all line items
3. **Discounts**: Applied promotions/coupons
4. **Taxes**: VAT/GST based on jurisdiction
5. **Total**: Final amount due
6. **Due Date**: Payment deadline

### Disputing an Invoice

1. Initiate dispute within the dispute window (typically 7 days)
2. Provide reason and supporting evidence
3. Wait for provider response
4. Escalate if not resolved
5. Request arbitration as last resort

### Tax Exemptions

If eligible for tax exemption:

1. Set up customer tax profile with valid Tax ID
2. Wait for verification (usually automated)
3. Future invoices will reflect exemption
4. B2B customers: reverse charge applies automatically

### Payment Terms

- Default: 7 days from invoice issue
- Late fees may apply after grace period
- Overdue accounts may have services suspended

---

## API Reference

### Query Invoice

```bash
virtengine query escrow invoice [invoice-id]
```

### List Customer Invoices

```bash
virtengine query escrow invoices-by-customer [customer-address]
```

### Dispute Invoice

```bash
virtengine tx escrow dispute-invoice [invoice-id] \
  --reason "Incorrect usage charges" \
  --from customer
```

### Resolve Dispute

```bash
virtengine tx escrow resolve-dispute [dispute-id] \
  --resolution partial_refund \
  --refund-amount 50000uvirt \
  --from provider
```

---

## Usage→Invoice→Settlement Pipeline

### Overview

The usage→invoice→settlement pipeline automates the flow from provider usage reports to customer invoices and final escrow settlement. This end-to-end pipeline ensures providers are paid for consumed resources and customers are billed accurately.

```
Provider Daemon → UsageReport → Invoice Generation → Approval/Dispute → Settlement → Payout
```

### Pipeline Stages

#### 1. Usage Report Submission

Providers submit usage reports via the `UsagePipelineKeeper.SubmitUsageReport` interface. Each report contains:

| Field | Type | Description |
|-------|------|-------------|
| `provider` | bech32 address | Provider submitting the report |
| `customer` | bech32 address | Customer being billed |
| `lease_id` | string | Active lease identifier |
| `escrow_id` | string | Linked escrow account |
| `resources` | array | Resource usage breakdown |
| `period_start` | timestamp | Usage period start |
| `period_end` | timestamp | Usage period end |

**Validation rules:**
- Provider and customer must be valid bech32 addresses
- Period end must be strictly after period start
- At least one resource must be reported
- Resource quantities must be non-negative
- Resource units must be non-empty

#### 2. Invoice Generation

`GenerateInvoiceFromUsage` converts validated usage reports into invoices:

1. Each `ResourceUsage` entry becomes a `LineItem`
2. Line item amounts are calculated as `quantity × unit_price`
3. All line items are summed into the invoice subtotal
4. The invoice is created in `Pending` status with a 30-day due date
5. A resource breakdown is stored on-chain for audit

**Resource types supported:**
- CPU (milli-seconds)
- Memory (byte-seconds)
- Storage (byte-seconds)
- Network (bytes)
- GPU (seconds)

#### 3. Invoice Approval and Dispute

Invoices follow the standard lifecycle:

- **Approve**: Customer or automated process moves invoice to `Paid` status
- **Dispute**: Customer can dispute within the dispute window, which places a holdback on escrow funds

#### 4. Settlement

`ProcessUsageSettlement` handles the full settlement flow:

1. Generates invoice from usage report
2. Settles the invoice through escrow (fees deducted)
3. Calculates holdback amounts for disputed invoices
4. Tracks payout records for provider reconciliation

**Fee breakdown (default):**
- Platform fee: 2.5%
- Network fee: 0.5%
- Community fee: 1.0%
- Take fee: 4.0%
- **Total fees: 8.0%**

#### 5. Off-Chain Reconciliation

The `ReconciliationService` in `pkg/billing/` provides usage↔invoice↔payout comparison:

- Detects missing invoices for submitted usage
- Identifies missing payouts for settled invoices
- Flags amount mismatches and overpayments
- Reports provider address mismatches
- Configurable variance threshold for rounding tolerance

**Discrepancy severity levels:**
- `low`: Within acceptable variance
- `medium`: Requires review
- `high`: Likely error, needs immediate attention
- `critical`: System integrity concern

### Reconciliation Reporting

Reconciliation reports include:
- Total usage records processed
- Total invoices matched
- Total payouts verified
- Discrepancy count by type and severity
- Summary totals (usage amount, invoiced amount, paid out amount)

### Export Formats

All billing data supports export in CSV, JSON, and PDF formats via the `ExportService`:
- **Invoice CSV**: Per-line-item export with all pricing details
- **Invoice JSON**: Full invoice document with schema versioning
- **Invoice PDF**: Formatted document with company branding
- **Settlement CSV**: Settlement summary with fee breakdowns

---

## Appendix: JSON Schema

Full JSON schema for invoices is available at:
- `x/escrow/types/billing/schema/invoice.schema.json`

Example invoices:
- `x/escrow/types/billing/schema/example-invoice-compute.json`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.1.0 | 2026-01-31 | Added PDF export, hash chain ledger |
| 1.0.0 | 2026-01-30 | Initial billing policy |

---

## Export Formats

### Supported Formats

VirtEngine supports multiple export formats for invoices and billing data:

| Format | Extension | MIME Type | Description |
|--------|-----------|-----------|-------------|
| PDF | `.pdf` | `application/pdf` | Print-ready invoice documents |
| JSON | `.json` | `application/json` | Machine-readable structured data |
| CSV | `.csv` | `text/csv` | Spreadsheet-compatible format |
| Excel | `.xlsx` | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` | Microsoft Excel format |

### PDF Export

PDF exports are generated using the `pkg/billing` export service and include:

- **Header**: Company branding, contact information
- **Invoice Information**: Number, dates, status
- **Party Details**: Provider and customer addresses
- **Line Items**: Detailed usage breakdown
- **Summary**: Subtotal, discounts, tax, total
- **Payment Instructions**: For pending invoices
- **Terms and Conditions**: Configurable footer

#### PDF Configuration

```go
PDFConfig{
    CompanyName:    "VirtEngine",
    CompanyAddress: []string{"VirtEngine Inc."},
    CompanyEmail:   "billing@virtengine.io",
    PrimaryColor:   "#2563eb",
    PageSize:       "A4",
    IncludeTerms:   true,
    ShowTaxBreakdown: true,
}
```

#### Generating PDF Export

```bash
# Via CLI (planned)
virtengine query escrow invoice [invoice-id] --output pdf > invoice.pdf

# Via gRPC
QueryInvoiceExport(invoice_id, format="PDF")
```

### JSON Export

JSON exports follow the `virtengine/invoice/v1` schema:

```json
{
  "version": "1.0",
  "schema_version": "virtengine/invoice/v1",
  "exported_at": "2026-01-31T12:00:00Z",
  "invoice": { ... }
}
```

### CSV Export

CSV exports include:
- **Invoice Summary CSV**: One row per invoice with key fields
- **Line Items CSV**: Detailed line item breakdown

---

## Immutable Ledger

### Hash Chain Design

Invoice ledger entries form an immutable hash chain for audit integrity:

```
[Genesis Entry] → [Status Change] → [Payment] → [Resolution]
     ↓                  ↓               ↓            ↓
   Hash₀              Hash₁           Hash₂        Hash₃
   (Zero)            (Hash₀)         (Hash₁)      (Hash₂)
```

### Ledger Entry Structure

| Field | Description |
|-------|-------------|
| `entry_id` | Unique entry identifier |
| `invoice_id` | Associated invoice |
| `entry_type` | created, issued, payment, disputed, resolved, etc. |
| `previous_entry_hash` | SHA-256 hash of previous entry |
| `entry_hash` | SHA-256 hash of this entry |
| `sequence_number` | Position in chain (1-indexed) |
| `block_height` | Blockchain height at creation |
| `timestamp` | Entry creation time |

### Genesis Entry

The first entry in any invoice's ledger chain uses a zero hash:

```
PreviousEntryHash: "0000000000000000000000000000000000000000000000000000000000000000"
SequenceNumber: 1
```

### Chain Verification

The ledger chain can be verified at any time:

```bash
virtengine query escrow verify-ledger-chain [invoice-id]
```

This verifies:
1. All entries have valid hash computations
2. Each entry links correctly to its predecessor
3. Sequence numbers are consecutive
4. Genesis entry has zero hash

### Audit Properties

- **Immutability**: Once recorded, entries cannot be modified
- **Ordering**: Cryptographic chain ensures entry ordering
- **Integrity**: Hash verification detects tampering
- **Traceability**: Complete audit trail for disputes
