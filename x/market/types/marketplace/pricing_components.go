// Package marketplace provides types for the marketplace on-chain module.
//
// VE-23Z: Component-based pricing calculations for marketplace offerings.
package marketplace

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PricingQuote represents a computed price for an offering.
type PricingQuote struct {
	// Total is the total computed price.
	Total sdk.Coin `json:"total"`
}

// CalculateOfferingPrice computes the total price for an offering given resource units.
// If component prices are present, they are used. Otherwise, legacy PricingInfo is used.
func CalculateOfferingPrice(offering *Offering, resourceUnits map[string]uint64, quantity uint32) (PricingQuote, error) {
	if offering == nil {
		return PricingQuote{}, fmt.Errorf("offering is required")
	}
	if quantity == 0 {
		return PricingQuote{}, fmt.Errorf("quantity must be positive")
	}

	if len(offering.Prices) > 0 {
		return calculateComponentPrice(offering.Prices, resourceUnits, quantity)
	}

	return calculateLegacyPrice(offering.Pricing, resourceUnits, quantity)
}

func calculateComponentPrice(components []PriceComponent, resourceUnits map[string]uint64, quantity uint32) (PricingQuote, error) {
	if err := validatePriceComponents(components); err != nil {
		return PricingQuote{}, err
	}

	denom := components[0].Price.Denom
	total := sdk.NewCoin(denom, sdkmath.NewInt(0))

	for _, component := range components {
		units := resourceUnits[string(component.ResourceType)]
		if units == 0 {
			continue
		}
		lineAmount := component.Price.Amount.Mul(sdkmath.NewIntFromUint64(units))
		total = total.Add(sdk.NewCoin(denom, lineAmount))
	}

	if quantity > 1 {
		total = sdk.NewCoin(denom, total.Amount.Mul(sdkmath.NewIntFromUint64(uint64(quantity))))
	}

	if !total.Amount.IsPositive() {
		return PricingQuote{}, fmt.Errorf("calculated price is zero")
	}

	return PricingQuote{Total: total}, nil
}

func calculateLegacyPrice(pricing PricingInfo, resourceUnits map[string]uint64, quantity uint32) (PricingQuote, error) {
	if err := pricing.Validate(); err != nil {
		return PricingQuote{}, err
	}
	if pricing.Currency == "" {
		return PricingQuote{}, fmt.Errorf("pricing currency is required")
	}

	total := sdk.NewCoin(pricing.Currency, sdkmath.NewInt(0))
	if pricing.BasePrice > 0 {
		total = total.Add(sdk.NewCoin(pricing.Currency, sdkmath.NewIntFromUint64(pricing.BasePrice)))
	}

	for resourceType, units := range resourceUnits {
		if units == 0 {
			continue
		}
		rate, ok := pricing.UsageRates[resourceType]
		if !ok {
			continue
		}
		lineAmount := sdkmath.NewIntFromUint64(rate).Mul(sdkmath.NewIntFromUint64(units))
		total = total.Add(sdk.NewCoin(pricing.Currency, lineAmount))
	}

	if quantity > 1 {
		total = sdk.NewCoin(pricing.Currency, total.Amount.Mul(sdkmath.NewIntFromUint64(uint64(quantity))))
	}

	if !total.Amount.IsPositive() {
		return PricingQuote{}, fmt.Errorf("calculated price is zero")
	}

	return PricingQuote{Total: total}, nil
}
