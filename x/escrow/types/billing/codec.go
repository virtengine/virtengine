// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers billing types for Amino encoding
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Invoice types
	cdc.RegisterConcrete(&Invoice{}, "virtengine/billing/Invoice", nil)
	cdc.RegisterConcrete(&LineItem{}, "virtengine/billing/LineItem", nil)
	cdc.RegisterConcrete(&InvoiceSummary{}, "virtengine/billing/InvoiceSummary", nil)

	// Ledger types
	cdc.RegisterConcrete(&InvoiceLedgerRecord{}, "virtengine/billing/InvoiceLedgerRecord", nil)
	cdc.RegisterConcrete(&InvoiceLedgerEntry{}, "virtengine/billing/InvoiceLedgerEntry", nil)

	// Artifact types
	cdc.RegisterConcrete(&InvoiceArtifact{}, "virtengine/billing/InvoiceArtifact", nil)

	// Discount types
	cdc.RegisterConcrete(&DiscountPolicy{}, "virtengine/billing/DiscountPolicy", nil)
	cdc.RegisterConcrete(&CouponCode{}, "virtengine/billing/CouponCode", nil)
	cdc.RegisterConcrete(&AppliedDiscount{}, "virtengine/billing/AppliedDiscount", nil)
	cdc.RegisterConcrete(&LoyaltyProgram{}, "virtengine/billing/LoyaltyProgram", nil)
	cdc.RegisterConcrete(&CustomerLoyalty{}, "virtengine/billing/CustomerLoyalty", nil)

	// Tax types
	cdc.RegisterConcrete(&TaxJurisdiction{}, "virtengine/billing/TaxJurisdiction", nil)
	cdc.RegisterConcrete(&TaxDetails{}, "virtengine/billing/TaxDetails", nil)
	cdc.RegisterConcrete(&CustomerTaxProfile{}, "virtengine/billing/CustomerTaxProfile", nil)
	cdc.RegisterConcrete(&ProviderTaxProfile{}, "virtengine/billing/ProviderTaxProfile", nil)

	// Dispute types
	cdc.RegisterConcrete(&DisputeWindow{}, "virtengine/billing/DisputeWindow", nil)

	// Pricing types
	cdc.RegisterConcrete(&PricingPolicy{}, "virtengine/billing/PricingPolicy", nil)
	cdc.RegisterConcrete(&ResourcePricing{}, "virtengine/billing/ResourcePricing", nil)

	// Settlement types
	cdc.RegisterConcrete(&SettlementConfig{}, "virtengine/billing/SettlementConfig", nil)
	cdc.RegisterConcrete(&SettlementHookConfig{}, "virtengine/billing/SettlementHookConfig", nil)
	cdc.RegisterConcrete(&SettlementHookResult{}, "virtengine/billing/SettlementHookResult", nil)

	// Reconciliation types
	cdc.RegisterConcrete(&ReconciliationReport{}, "virtengine/billing/ReconciliationReport", nil)
	cdc.RegisterConcrete(&ReconciliationDiscrepancy{}, "virtengine/billing/ReconciliationDiscrepancy", nil)
	cdc.RegisterConcrete(&ReconciliationHookConfig{}, "virtengine/billing/ReconciliationHookConfig", nil)
}

// RegisterInterfaces registers billing types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Billing types don't implement sdk.Msg, so no message registration needed
	// This is here for future extension if needed
}
