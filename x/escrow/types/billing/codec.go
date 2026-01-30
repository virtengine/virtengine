// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers billing types for Amino encoding
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Invoice{}, "virtengine/billing/Invoice", nil)
	cdc.RegisterConcrete(&LineItem{}, "virtengine/billing/LineItem", nil)
	cdc.RegisterConcrete(&DiscountPolicy{}, "virtengine/billing/DiscountPolicy", nil)
	cdc.RegisterConcrete(&CouponCode{}, "virtengine/billing/CouponCode", nil)
	cdc.RegisterConcrete(&AppliedDiscount{}, "virtengine/billing/AppliedDiscount", nil)
	cdc.RegisterConcrete(&TaxJurisdiction{}, "virtengine/billing/TaxJurisdiction", nil)
	cdc.RegisterConcrete(&TaxDetails{}, "virtengine/billing/TaxDetails", nil)
	cdc.RegisterConcrete(&CustomerTaxProfile{}, "virtengine/billing/CustomerTaxProfile", nil)
	cdc.RegisterConcrete(&ProviderTaxProfile{}, "virtengine/billing/ProviderTaxProfile", nil)
	cdc.RegisterConcrete(&DisputeWindow{}, "virtengine/billing/DisputeWindow", nil)
	cdc.RegisterConcrete(&PricingPolicy{}, "virtengine/billing/PricingPolicy", nil)
	cdc.RegisterConcrete(&SettlementConfig{}, "virtengine/billing/SettlementConfig", nil)
	cdc.RegisterConcrete(&LoyaltyProgram{}, "virtengine/billing/LoyaltyProgram", nil)
	cdc.RegisterConcrete(&CustomerLoyalty{}, "virtengine/billing/CustomerLoyalty", nil)
}

// RegisterInterfaces registers billing types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Billing types don't implement sdk.Msg, so no message registration needed
	// This is here for future extension if needed
}
