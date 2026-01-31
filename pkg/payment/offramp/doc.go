// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
//
// This package implements:
//   - PayPal Payouts API integration for instant payouts
//   - ACH bank transfer integration via Stripe Treasury
//   - KYC/AML gating for compliance
//   - Payout status tracking and reconciliation
//   - Webhook handling for payout notifications
//
// # Architecture
//
// The off-ramp service follows the same adapter pattern as the payment package:
//
//	                    ┌─────────────────┐
//	                    │  OffRampService │
//	                    └────────┬────────┘
//	                             │
//	              ┌──────────────┼──────────────┐
//	              │              │              │
//	     ┌────────▼────────┐ ┌───▼───┐ ┌───────▼───────┐
//	     │  PayPalAdapter  │ │ ACH   │ │  KYC/AML Gate │
//	     └────────┬────────┘ └───┬───┘ └───────────────┘
//	              │              │
//	     ┌────────▼────────┐ ┌───▼───────────────┐
//	     │  PayPal Payouts │ │  Stripe Treasury  │
//	     │      API        │ │       API         │
//	     └─────────────────┘ └───────────────────┘
//
// # Compliance
//
// All payouts require:
//   - Verified VEID identity with minimum verification level
//   - Passed KYC checks (identity verification)
//   - Passed AML screening (sanctions, PEP lists)
//   - Rate limiting to prevent abuse
//
// # Reconciliation
//
// The reconciliation job runs periodically to:
//   - Match on-chain settlement records with provider reports
//   - Flag discrepancies for manual review
//   - Update payout status based on provider confirmations
//
// Task Reference: VE-5E - Fiat off-ramp integration
package offramp
