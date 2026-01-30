// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TaxType defines the type of tax
type TaxType uint8

const (
	// TaxTypeVAT is Value Added Tax
	TaxTypeVAT TaxType = 0

	// TaxTypeGST is Goods and Services Tax
	TaxTypeGST TaxType = 1

	// TaxTypeSalesTax is Sales Tax
	TaxTypeSalesTax TaxType = 2

	// TaxTypeWithholding is Withholding Tax
	TaxTypeWithholding TaxType = 3

	// TaxTypeServiceTax is Service Tax
	TaxTypeServiceTax TaxType = 4

	// TaxTypeDigitalServices is Digital Services Tax
	TaxTypeDigitalServices TaxType = 5

	// TaxTypeOther is other tax types
	TaxTypeOther TaxType = 6
)

// TaxTypeNames maps types to names
var TaxTypeNames = map[TaxType]string{
	TaxTypeVAT:             "vat",
	TaxTypeGST:             "gst",
	TaxTypeSalesTax:        "sales_tax",
	TaxTypeWithholding:     "withholding",
	TaxTypeServiceTax:      "service_tax",
	TaxTypeDigitalServices: "digital_services",
	TaxTypeOther:           "other",
}

// String returns string representation
func (t TaxType) String() string {
	if name, ok := TaxTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// TaxExemptionCategory defines exemption categories
type TaxExemptionCategory uint8

const (
	// TaxExemptionNone means no exemption
	TaxExemptionNone TaxExemptionCategory = 0

	// TaxExemptionB2B is B2B reverse charge
	TaxExemptionB2B TaxExemptionCategory = 1

	// TaxExemptionNonProfit is non-profit exemption
	TaxExemptionNonProfit TaxExemptionCategory = 2

	// TaxExemptionGovernment is government exemption
	TaxExemptionGovernment TaxExemptionCategory = 3

	// TaxExemptionEducation is educational institution exemption
	TaxExemptionEducation TaxExemptionCategory = 4

	// TaxExemptionExport is export exemption
	TaxExemptionExport TaxExemptionCategory = 5

	// TaxExemptionSmallBusiness is small business exemption
	TaxExemptionSmallBusiness TaxExemptionCategory = 6
)

// TaxExemptionCategoryNames maps categories to names
var TaxExemptionCategoryNames = map[TaxExemptionCategory]string{
	TaxExemptionNone:          "none",
	TaxExemptionB2B:           "b2b",
	TaxExemptionNonProfit:     "non_profit",
	TaxExemptionGovernment:    "government",
	TaxExemptionEducation:     "education",
	TaxExemptionExport:        "export",
	TaxExemptionSmallBusiness: "small_business",
}

// String returns string representation
func (c TaxExemptionCategory) String() string {
	if name, ok := TaxExemptionCategoryNames[c]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", c)
}

// TaxJurisdiction defines a tax jurisdiction
type TaxJurisdiction struct {
	// JurisdictionID is the unique identifier
	JurisdictionID string `json:"jurisdiction_id"`

	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"country_code"`

	// CountryName is the country name
	CountryName string `json:"country_name"`

	// RegionCode is the region/state code (if applicable)
	RegionCode string `json:"region_code,omitempty"`

	// RegionName is the region/state name
	RegionName string `json:"region_name,omitempty"`

	// TaxRates are the tax rates for this jurisdiction
	TaxRates []TaxRate `json:"tax_rates"`

	// RequiredTaxID indicates if tax ID is required for B2B
	RequiredTaxID bool `json:"required_tax_id"`

	// TaxIDFormat is the expected format for tax IDs
	TaxIDFormat string `json:"tax_id_format,omitempty"`

	// DigitalServicesApplicable indicates if DST applies
	DigitalServicesApplicable bool `json:"digital_services_applicable"`

	// EffectiveFrom is when this jurisdiction config takes effect
	EffectiveFrom time.Time `json:"effective_from"`

	// EffectiveUntil is when this config expires
	EffectiveUntil time.Time `json:"effective_until,omitempty"`
}

// TaxRate defines a tax rate
type TaxRate struct {
	// RateID is the unique identifier
	RateID string `json:"rate_id"`

	// TaxType is the type of tax
	TaxType TaxType `json:"tax_type"`

	// Name is the tax name
	Name string `json:"name"`

	// RateBps is the tax rate in basis points (e.g., 2000 = 20%)
	RateBps uint32 `json:"rate_bps"`

	// IsCompound indicates if this is compounded on other taxes
	IsCompound bool `json:"is_compound"`

	// AppliesToUsageTypes limits which usage types are taxed
	AppliesToUsageTypes []UsageType `json:"applies_to_usage_types,omitempty"`

	// ExemptCategories are categories exempt from this tax
	ExemptCategories []TaxExemptionCategory `json:"exempt_categories,omitempty"`

	// IncludedInPrice indicates if tax is included in listed prices
	IncludedInPrice bool `json:"included_in_price"`
}

// Validate validates the tax rate
func (r *TaxRate) Validate() error {
	if r.RateID == "" {
		return fmt.Errorf("rate_id is required")
	}

	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.RateBps > 10000 {
		return fmt.Errorf("rate_bps cannot exceed 10000 (100%%)")
	}

	return nil
}

// GetRateDecimal returns the rate as a decimal (e.g., 0.20 for 20%)
func (r *TaxRate) GetRateDecimal() sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(r.RateBps)).Quo(sdkmath.LegacyNewDec(10000))
}

// IsExempt checks if a category is exempt from this tax
func (r *TaxRate) IsExempt(category TaxExemptionCategory) bool {
	for _, exempt := range r.ExemptCategories {
		if exempt == category {
			return true
		}
	}
	return false
}

// Validate validates the tax jurisdiction
func (j *TaxJurisdiction) Validate() error {
	if j.JurisdictionID == "" {
		return fmt.Errorf("jurisdiction_id is required")
	}

	if j.CountryCode == "" {
		return fmt.Errorf("country_code is required")
	}

	if len(j.CountryCode) != 2 {
		return fmt.Errorf("country_code must be ISO 3166-1 alpha-2 (2 characters)")
	}

	for i, rate := range j.TaxRates {
		if err := rate.Validate(); err != nil {
			return fmt.Errorf("tax_rates[%d]: %w", i, err)
		}
	}

	return nil
}

// GetApplicableRates returns applicable tax rates for a given exemption category
func (j *TaxJurisdiction) GetApplicableRates(category TaxExemptionCategory) []TaxRate {
	var rates []TaxRate
	for _, rate := range j.TaxRates {
		if !rate.IsExempt(category) {
			rates = append(rates, rate)
		}
	}
	return rates
}

// TaxDetails contains tax calculation details for an invoice
type TaxDetails struct {
	// Jurisdiction is the tax jurisdiction
	Jurisdiction TaxJurisdiction `json:"jurisdiction"`

	// CustomerTaxID is the customer's tax ID
	CustomerTaxID string `json:"customer_tax_id,omitempty"`

	// ProviderTaxID is the provider's tax ID
	ProviderTaxID string `json:"provider_tax_id,omitempty"`

	// ExemptionCategory is the customer's exemption category
	ExemptionCategory TaxExemptionCategory `json:"exemption_category"`

	// TaxLineItems are the individual tax calculations
	TaxLineItems []TaxLineItem `json:"tax_line_items"`

	// TotalTax is the total tax amount
	TotalTax sdk.Coins `json:"total_tax"`

	// IsReverseCharge indicates if reverse charge applies
	IsReverseCharge bool `json:"is_reverse_charge"`

	// ReverseChargeNote is the note for reverse charge
	ReverseChargeNote string `json:"reverse_charge_note,omitempty"`

	// TaxPointDate is the tax point date
	TaxPointDate time.Time `json:"tax_point_date"`

	// CalculatedAt is when tax was calculated
	CalculatedAt time.Time `json:"calculated_at"`
}

// TaxLineItem represents a single tax calculation
type TaxLineItem struct {
	// TaxRateID is the tax rate applied
	TaxRateID string `json:"tax_rate_id"`

	// TaxType is the type of tax
	TaxType TaxType `json:"tax_type"`

	// Name is the tax name
	Name string `json:"name"`

	// RateBps is the rate in basis points
	RateBps uint32 `json:"rate_bps"`

	// TaxableAmount is the amount subject to tax
	TaxableAmount sdk.Coins `json:"taxable_amount"`

	// TaxAmount is the calculated tax
	TaxAmount sdk.Coins `json:"tax_amount"`

	// IsExempt indicates if this line is exempt
	IsExempt bool `json:"is_exempt"`

	// ExemptionReason is the reason for exemption
	ExemptionReason string `json:"exemption_reason,omitempty"`
}

// Validate validates the tax details
func (t *TaxDetails) Validate() error {
	if err := t.Jurisdiction.Validate(); err != nil {
		return fmt.Errorf("jurisdiction: %w", err)
	}

	if !t.TotalTax.IsValid() {
		return fmt.Errorf("total_tax must be valid")
	}

	return nil
}

// TaxCalculator calculates taxes for an invoice
type TaxCalculator struct {
	// Jurisdictions maps country codes to jurisdictions
	Jurisdictions map[string]TaxJurisdiction `json:"jurisdictions"`

	// DefaultJurisdiction is the default jurisdiction code
	DefaultJurisdiction string `json:"default_jurisdiction"`

	// RoundingMode is the rounding mode for tax calculations
	RoundingMode RoundingMode `json:"rounding_mode"`
}

// NewTaxCalculator creates a new tax calculator
func NewTaxCalculator(defaultJurisdiction string) *TaxCalculator {
	return &TaxCalculator{
		Jurisdictions:       make(map[string]TaxJurisdiction),
		DefaultJurisdiction: defaultJurisdiction,
		RoundingMode:        RoundingModeHalfUp,
	}
}

// AddJurisdiction adds a tax jurisdiction
func (tc *TaxCalculator) AddJurisdiction(j TaxJurisdiction) error {
	if err := j.Validate(); err != nil {
		return err
	}
	tc.Jurisdictions[j.CountryCode] = j
	return nil
}

// CalculateTax calculates tax for an invoice
func (tc *TaxCalculator) CalculateTax(
	subtotal sdk.Coins,
	customerCountry string,
	exemptionCategory TaxExemptionCategory,
	providerCountry string,
	now time.Time,
) (*TaxDetails, error) {
	jurisdiction, ok := tc.Jurisdictions[customerCountry]
	if !ok {
		jurisdiction, ok = tc.Jurisdictions[tc.DefaultJurisdiction]
		if !ok {
			// No jurisdiction found - no tax
			return &TaxDetails{
				ExemptionCategory: exemptionCategory,
				TaxLineItems:      []TaxLineItem{},
				TotalTax:          sdk.NewCoins(),
				TaxPointDate:      now,
				CalculatedAt:      now,
			}, nil
		}
	}

	details := &TaxDetails{
		Jurisdiction:      jurisdiction,
		ExemptionCategory: exemptionCategory,
		TaxLineItems:      make([]TaxLineItem, 0),
		TotalTax:          sdk.NewCoins(),
		TaxPointDate:      now,
		CalculatedAt:      now,
	}

	// Check for B2B reverse charge (provider and customer in different countries)
	if exemptionCategory == TaxExemptionB2B && providerCountry != customerCountry {
		details.IsReverseCharge = true
		details.ReverseChargeNote = "Reverse charge: VAT to be accounted for by the recipient"
		return details, nil
	}

	// Get applicable rates
	applicableRates := jurisdiction.GetApplicableRates(exemptionCategory)

	totalTax := sdk.NewCoins()
	taxableBase := subtotal

	for _, rate := range applicableRates {
		// Calculate tax for this rate
		taxLineItem := TaxLineItem{
			TaxRateID:     rate.RateID,
			TaxType:       rate.TaxType,
			Name:          rate.Name,
			RateBps:       rate.RateBps,
			TaxableAmount: taxableBase,
		}

		if rate.IsExempt(exemptionCategory) {
			taxLineItem.IsExempt = true
			taxLineItem.ExemptionReason = exemptionCategory.String()
			taxLineItem.TaxAmount = sdk.NewCoins()
		} else {
			// Calculate tax amount
			rateDecimal := rate.GetRateDecimal()
			taxAmount := sdk.NewCoins()

			for _, coin := range taxableBase {
				taxAmt := sdkmath.LegacyNewDecFromInt(coin.Amount).Mul(rateDecimal)
				roundedTax := applyRounding(taxAmt, tc.RoundingMode)
				taxAmount = taxAmount.Add(sdk.NewCoin(coin.Denom, roundedTax))
			}

			taxLineItem.TaxAmount = taxAmount
			totalTax = totalTax.Add(taxAmount...)

			// If compound, add to taxable base for next rate
			if rate.IsCompound {
				taxableBase = taxableBase.Add(taxAmount...)
			}
		}

		details.TaxLineItems = append(details.TaxLineItems, taxLineItem)
	}

	details.TotalTax = totalTax
	return details, nil
}

// CustomerTaxProfile stores a customer's tax information
type CustomerTaxProfile struct {
	// Customer is the customer address
	Customer string `json:"customer"`

	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"country_code"`

	// RegionCode is the region/state code
	RegionCode string `json:"region_code,omitempty"`

	// TaxID is the customer's tax identification number
	TaxID string `json:"tax_id,omitempty"`

	// TaxIDVerified indicates if tax ID was verified
	TaxIDVerified bool `json:"tax_id_verified"`

	// TaxIDVerifiedAt is when tax ID was verified
	TaxIDVerifiedAt *time.Time `json:"tax_id_verified_at,omitempty"`

	// ExemptionCategory is the customer's exemption category
	ExemptionCategory TaxExemptionCategory `json:"exemption_category"`

	// ExemptionCertificate is reference to exemption certificate
	ExemptionCertificate string `json:"exemption_certificate,omitempty"`

	// BusinessName is the business name
	BusinessName string `json:"business_name,omitempty"`

	// IsB2B indicates if this is a B2B customer
	IsB2B bool `json:"is_b2b"`

	// UpdatedAt is when the profile was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the customer tax profile
func (p *CustomerTaxProfile) Validate() error {
	if p.Customer == "" {
		return fmt.Errorf("customer is required")
	}

	if _, err := sdk.AccAddressFromBech32(p.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if p.CountryCode == "" {
		return fmt.Errorf("country_code is required")
	}

	if len(p.CountryCode) != 2 {
		return fmt.Errorf("country_code must be ISO 3166-1 alpha-2 (2 characters)")
	}

	return nil
}

// ProviderTaxProfile stores a provider's tax information
type ProviderTaxProfile struct {
	// Provider is the provider address
	Provider string `json:"provider"`

	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"country_code"`

	// RegionCode is the region/state code
	RegionCode string `json:"region_code,omitempty"`

	// TaxID is the provider's tax identification number
	TaxID string `json:"tax_id"`

	// TaxIDVerified indicates if tax ID was verified
	TaxIDVerified bool `json:"tax_id_verified"`

	// BusinessName is the registered business name
	BusinessName string `json:"business_name"`

	// TaxRegistrationNumber is the tax registration number
	TaxRegistrationNumber string `json:"tax_registration_number,omitempty"`

	// UpdatedAt is when the profile was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the provider tax profile
func (p *ProviderTaxProfile) Validate() error {
	if p.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if _, err := sdk.AccAddressFromBech32(p.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if p.CountryCode == "" {
		return fmt.Errorf("country_code is required")
	}

	if p.TaxID == "" {
		return fmt.Errorf("tax_id is required for providers")
	}

	return nil
}

// DefaultTaxJurisdictions returns common tax jurisdictions
func DefaultTaxJurisdictions() map[string]TaxJurisdiction {
	return map[string]TaxJurisdiction{
		"US": {
			JurisdictionID:            "US",
			CountryCode:               "US",
			CountryName:               "United States",
			RequiredTaxID:             false,
			DigitalServicesApplicable: false,
			TaxRates:                  []TaxRate{}, // US has no federal digital services tax
		},
		"GB": {
			JurisdictionID:            "GB",
			CountryCode:               "GB",
			CountryName:               "United Kingdom",
			RequiredTaxID:             true,
			TaxIDFormat:               "GB[0-9]{9}",
			DigitalServicesApplicable: true,
			TaxRates: []TaxRate{
				{
					RateID:   "GB-VAT-20",
					TaxType:  TaxTypeVAT,
					Name:     "VAT Standard Rate",
					RateBps:  2000, // 20%
					ExemptCategories: []TaxExemptionCategory{
						TaxExemptionB2B,
						TaxExemptionExport,
					},
				},
			},
		},
		"DE": {
			JurisdictionID:            "DE",
			CountryCode:               "DE",
			CountryName:               "Germany",
			RequiredTaxID:             true,
			TaxIDFormat:               "DE[0-9]{9}",
			DigitalServicesApplicable: true,
			TaxRates: []TaxRate{
				{
					RateID:   "DE-VAT-19",
					TaxType:  TaxTypeVAT,
					Name:     "Mehrwertsteuer",
					RateBps:  1900, // 19%
					ExemptCategories: []TaxExemptionCategory{
						TaxExemptionB2B,
						TaxExemptionExport,
					},
				},
			},
		},
		"SG": {
			JurisdictionID:            "SG",
			CountryCode:               "SG",
			CountryName:               "Singapore",
			RequiredTaxID:             true,
			TaxIDFormat:               "[0-9]{9}[A-Z]",
			DigitalServicesApplicable: true,
			TaxRates: []TaxRate{
				{
					RateID:   "SG-GST-9",
					TaxType:  TaxTypeGST,
					Name:     "Goods and Services Tax",
					RateBps:  900, // 9%
					ExemptCategories: []TaxExemptionCategory{
						TaxExemptionExport,
					},
				},
			},
		},
		"AU": {
			JurisdictionID:            "AU",
			CountryCode:               "AU",
			CountryName:               "Australia",
			RequiredTaxID:             true,
			TaxIDFormat:               "[0-9]{11}",
			DigitalServicesApplicable: true,
			TaxRates: []TaxRate{
				{
					RateID:   "AU-GST-10",
					TaxType:  TaxTypeGST,
					Name:     "Goods and Services Tax",
					RateBps:  1000, // 10%
					ExemptCategories: []TaxExemptionCategory{
						TaxExemptionExport,
					},
				},
			},
		},
	}
}
