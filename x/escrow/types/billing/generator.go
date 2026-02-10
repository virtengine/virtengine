// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InvoiceGeneratorConfig configures the invoice generator
type InvoiceGeneratorConfig struct {
	// InvoiceNumberPrefix is the prefix for invoice numbers
	InvoiceNumberPrefix string `json:"invoice_number_prefix"`

	// DefaultPaymentTermDays is the default days until payment is due
	DefaultPaymentTermDays int32 `json:"default_payment_term_days"`

	// RoundingMode is the rounding mode for calculations
	RoundingMode RoundingMode `json:"rounding_mode"`

	// DefaultCurrency is the default currency
	DefaultCurrency string `json:"default_currency"`

	// ApplyTax indicates if tax should be calculated
	ApplyTax bool `json:"apply_tax"`

	// DefaultTaxJurisdiction is the default tax jurisdiction
	DefaultTaxJurisdiction string `json:"default_tax_jurisdiction"`
}

// DefaultInvoiceGeneratorConfig returns default configuration
func DefaultInvoiceGeneratorConfig() InvoiceGeneratorConfig {
	return InvoiceGeneratorConfig{
		InvoiceNumberPrefix:    "VE-INV",
		DefaultPaymentTermDays: 7,
		RoundingMode:           RoundingModeHalfEven,
		DefaultCurrency:        DefaultCurrency,
		ApplyTax:               false,
		DefaultTaxJurisdiction: "US",
	}
}

// UsageInput represents usage data for invoice generation
type UsageInput struct {
	// UsageRecordID is the unique usage record ID
	UsageRecordID string `json:"usage_record_id"`

	// UsageType is the type of usage
	UsageType UsageType `json:"usage_type"`

	// Quantity is the quantity consumed
	Quantity sdkmath.LegacyDec `json:"quantity"`

	// Unit is the measurement unit
	Unit string `json:"unit"`

	// UnitPrice is the price per unit
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// Description is a description of the usage
	Description string `json:"description"`

	// PeriodStart is the usage period start
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the usage period end
	PeriodEnd time.Time `json:"period_end"`

	// Metadata contains additional usage details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate validates the usage input
func (u *UsageInput) Validate() error {
	if u.UsageRecordID == "" {
		return fmt.Errorf("usage_record_id is required")
	}

	if u.Quantity.IsNegative() {
		return fmt.Errorf("quantity cannot be negative")
	}

	if u.UnitPrice.IsNegative() {
		return fmt.Errorf("unit_price cannot be negative")
	}

	if u.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// InvoiceGenerationRequest contains all data needed to generate an invoice
type InvoiceGenerationRequest struct {
	// EscrowID is the escrow account ID
	EscrowID string `json:"escrow_id"`

	// OrderID is the marketplace order ID
	OrderID string `json:"order_id"`

	// LeaseID is the marketplace lease ID
	LeaseID string `json:"lease_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// UsageInputs are the usage records to include
	UsageInputs []UsageInput `json:"usage_inputs"`

	// BillingPeriod is the billing period
	BillingPeriod BillingPeriod `json:"billing_period"`

	// Currency is the invoice currency
	Currency string `json:"currency"`

	// DiscountPolicies are discount policies to apply
	DiscountPolicies []DiscountPolicy `json:"discount_policies,omitempty"`

	// CustomerTaxProfile is the customer's tax profile (optional)
	CustomerTaxProfile *CustomerTaxProfile `json:"customer_tax_profile,omitempty"`

	// ProviderTaxProfile is the provider's tax profile (optional)
	ProviderTaxProfile *ProviderTaxProfile `json:"provider_tax_profile,omitempty"`

	// Metadata contains additional invoice details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate validates the request
func (r *InvoiceGenerationRequest) Validate() error {
	if r.EscrowID == "" {
		return fmt.Errorf("escrow_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if len(r.UsageInputs) == 0 {
		return fmt.Errorf("at least one usage_input is required")
	}

	for i, input := range r.UsageInputs {
		if err := input.Validate(); err != nil {
			return fmt.Errorf("usage_inputs[%d]: %w", i, err)
		}
	}

	if r.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	return nil
}

// InvoiceGenerator generates invoices from usage records
type InvoiceGenerator struct {
	config        InvoiceGeneratorConfig
	taxCalculator *TaxCalculator
	sequence      uint64
}

// NewInvoiceGenerator creates a new invoice generator
func NewInvoiceGenerator(config InvoiceGeneratorConfig) *InvoiceGenerator {
	gen := &InvoiceGenerator{
		config:   config,
		sequence: 0,
	}

	// Initialize tax calculator if tax is enabled
	if config.ApplyTax {
		gen.taxCalculator = NewTaxCalculator(config.DefaultTaxJurisdiction)
		for code, jurisdiction := range DefaultTaxJurisdictions() {
			_ = gen.taxCalculator.AddJurisdiction(jurisdiction)
			_ = code // silence unused variable warning
		}
	}

	return gen
}

// SetSequence sets the invoice sequence number
func (g *InvoiceGenerator) SetSequence(seq uint64) {
	g.sequence = seq
}

// GenerateInvoice generates an invoice from the request
func (g *InvoiceGenerator) GenerateInvoice(
	req InvoiceGenerationRequest,
	blockHeight int64,
	now time.Time,
) (*Invoice, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate invoice ID deterministically
	invoiceID := g.generateInvoiceID(req)

	// Generate invoice number
	g.sequence++
	invoiceNumber := InvoiceNumberFromID(invoiceID, g.config.InvoiceNumberPrefix)

	// Calculate due date
	dueDate := now.Add(time.Duration(g.config.DefaultPaymentTermDays) * 24 * time.Hour)

	// Create invoice
	invoice := NewInvoice(
		invoiceID,
		invoiceNumber,
		req.EscrowID,
		req.OrderID,
		req.LeaseID,
		req.Provider,
		req.Customer,
		req.Currency,
		req.BillingPeriod,
		dueDate,
		blockHeight,
		now,
	)

	// Add line items from usage inputs (deterministic ordering)
	normalizedInputs := normalizeUsageInputs(req.UsageInputs)
	for _, usage := range normalizedInputs {
		lineItem := g.createLineItem(usage, req.Currency)
		invoice.AddLineItem(lineItem)
	}

	// Apply discounts
	if len(req.DiscountPolicies) > 0 {
		g.applyDiscounts(invoice, req.DiscountPolicies, now)
	}

	// Apply tax if configured
	if g.config.ApplyTax && g.taxCalculator != nil {
		g.applyTax(invoice, req.CustomerTaxProfile, req.ProviderTaxProfile, now)
	}

	// Add metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			invoice.Metadata[k] = v
		}
	}

	// Validate final invoice
	if err := invoice.Validate(); err != nil {
		return nil, fmt.Errorf("generated invoice is invalid: %w", err)
	}

	return invoice, nil
}

// generateInvoiceID generates a deterministic invoice ID
func (g *InvoiceGenerator) generateInvoiceID(req InvoiceGenerationRequest) string {
	normalizedInputs := normalizeUsageInputs(req.UsageInputs)
	canonicalInputs := make([]canonicalUsageInput, 0, len(normalizedInputs))
	for _, input := range normalizedInputs {
		canonicalInputs = append(canonicalInputs, canonicalUsageInput{
			UsageRecordID: input.UsageRecordID,
			UsageType:     input.UsageType.String(),
			Quantity:      input.Quantity.String(),
			Unit:          input.Unit,
			UnitPrice:     input.UnitPrice.String(),
			Description:   input.Description,
			PeriodStart:   input.PeriodStart.Unix(),
			PeriodEnd:     input.PeriodEnd.Unix(),
			Metadata:      sortedMetadataPairs(input.Metadata),
		})
	}

	canonical := struct {
		EscrowID     string                `json:"escrow_id"`
		OrderID      string                `json:"order_id"`
		LeaseID      string                `json:"lease_id"`
		Provider     string                `json:"provider"`
		Customer     string                `json:"customer"`
		Currency     string                `json:"currency"`
		BillingStart int64                 `json:"billing_start"`
		BillingEnd   int64                 `json:"billing_end"`
		BillingType  string                `json:"billing_type"`
		UsageInputs  []canonicalUsageInput `json:"usage_inputs"`
		Metadata     []metadataPair        `json:"metadata"`
	}{
		EscrowID:     req.EscrowID,
		OrderID:      req.OrderID,
		LeaseID:      req.LeaseID,
		Provider:     req.Provider,
		Customer:     req.Customer,
		Currency:     req.Currency,
		BillingStart: req.BillingPeriod.StartTime.Unix(),
		BillingEnd:   req.BillingPeriod.EndTime.Unix(),
		BillingType:  req.BillingPeriod.PeriodType.String(),
		UsageInputs:  canonicalInputs,
		Metadata:     sortedMetadataPairs(req.Metadata),
	}

	bytes, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for ID
}

// createLineItem creates a line item from usage input
func (g *InvoiceGenerator) createLineItem(usage UsageInput, currency string) LineItem {
	// Calculate line item amount with deterministic rounding
	rawAmount := usage.Quantity.Mul(usage.UnitPrice.Amount)
	roundedAmount := applyRounding(rawAmount, g.config.RoundingMode)

	return LineItem{
		LineItemID:     lineItemIDFromUsage(usage),
		Description:    usage.Description,
		UsageType:      usage.UsageType,
		Quantity:       usage.Quantity,
		Unit:           usage.Unit,
		UnitPrice:      usage.UnitPrice,
		Amount:         sdk.NewCoins(sdk.NewCoin(currency, roundedAmount)),
		UsageRecordIDs: []string{usage.UsageRecordID},
		Metadata:       usage.Metadata,
	}
}

type canonicalUsageInput struct {
	UsageRecordID string         `json:"usage_record_id"`
	UsageType     string         `json:"usage_type"`
	Quantity      string         `json:"quantity"`
	Unit          string         `json:"unit"`
	UnitPrice     string         `json:"unit_price"`
	Description   string         `json:"description"`
	PeriodStart   int64          `json:"period_start"`
	PeriodEnd     int64          `json:"period_end"`
	Metadata      []metadataPair `json:"metadata"`
}

type metadataPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func normalizeUsageInputs(inputs []UsageInput) []UsageInput {
	if len(inputs) == 0 {
		return inputs
	}

	normalized := make([]UsageInput, 0, len(inputs))
	normalized = append(normalized, inputs...)

	sort.SliceStable(normalized, func(i, j int) bool {
		return usageInputSortKey(normalized[i]) < usageInputSortKey(normalized[j])
	})

	return normalized
}

func usageInputSortKey(input UsageInput) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		input.UsageRecordID,
		input.UsageType.String(),
		input.Unit,
		input.UnitPrice.String(),
		input.Quantity.String(),
		input.PeriodStart.UTC().Format(time.RFC3339Nano),
		input.PeriodEnd.UTC().Format(time.RFC3339Nano),
	)
}

func sortedMetadataPairs(metadata map[string]string) []metadataPair {
	if len(metadata) == 0 {
		return nil
	}

	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]metadataPair, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, metadataPair{Key: key, Value: metadata[key]})
	}

	return pairs
}

func lineItemIDFromUsage(usage UsageInput) string {
	canonical := struct {
		UsageRecordID string         `json:"usage_record_id"`
		UsageType     string         `json:"usage_type"`
		Quantity      string         `json:"quantity"`
		Unit          string         `json:"unit"`
		UnitPrice     string         `json:"unit_price"`
		PeriodStart   int64          `json:"period_start"`
		PeriodEnd     int64          `json:"period_end"`
		Metadata      []metadataPair `json:"metadata"`
	}{
		UsageRecordID: usage.UsageRecordID,
		UsageType:     usage.UsageType.String(),
		Quantity:      usage.Quantity.String(),
		Unit:          usage.Unit,
		UnitPrice:     usage.UnitPrice.String(),
		PeriodStart:   usage.PeriodStart.Unix(),
		PeriodEnd:     usage.PeriodEnd.Unix(),
		Metadata:      sortedMetadataPairs(usage.Metadata),
	}

	bytes, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("li-%s", hex.EncodeToString(hash[:6]))
}

// applyDiscounts applies discount policies to the invoice
func (g *InvoiceGenerator) applyDiscounts(
	invoice *Invoice,
	policies []DiscountPolicy,
	now time.Time,
) {
	// Calculate total volume for volume-based discounts
	totalVolume := sdkmath.LegacyZeroDec()
	for _, item := range invoice.LineItems {
		totalVolume = totalVolume.Add(item.Quantity)
	}

	for _, policy := range policies {
		if !policy.IsValidAt(now) {
			continue
		}

		discountAmount := policy.CalculateDiscount(invoice.Subtotal, totalVolume)
		if discountAmount.IsZero() {
			continue
		}

		applied := AppliedDiscount{
			DiscountID:  fmt.Sprintf("disc-%s", policy.PolicyID),
			PolicyID:    policy.PolicyID,
			Type:        policy.Type,
			Description: policy.Description,
			Amount:      discountAmount,
			AppliedAt:   now,
			AppliedBy:   "system",
		}

		invoice.Discounts = append(invoice.Discounts, applied)
	}

	// Recalculate totals
	invoice.recalculateTotals()
}

// applyTax applies tax to the invoice
func (g *InvoiceGenerator) applyTax(
	invoice *Invoice,
	customerProfile *CustomerTaxProfile,
	providerProfile *ProviderTaxProfile,
	now time.Time,
) {
	customerCountry := g.config.DefaultTaxJurisdiction
	providerCountry := g.config.DefaultTaxJurisdiction
	exemptionCategory := TaxExemptionNone

	if customerProfile != nil {
		customerCountry = customerProfile.CountryCode
		exemptionCategory = customerProfile.ExemptionCategory
	}

	if providerProfile != nil {
		providerCountry = providerProfile.CountryCode
	}

	// Calculate tax on the subtotal after discounts
	taxableAmount := invoice.Subtotal.Sub(invoice.DiscountTotal...)

	taxDetails, err := g.taxCalculator.CalculateTax(
		taxableAmount,
		customerCountry,
		exemptionCategory,
		providerCountry,
		now,
	)
	if err != nil {
		return // Skip tax on error
	}

	if customerProfile != nil {
		taxDetails.CustomerTaxID = customerProfile.TaxID
	}

	if providerProfile != nil {
		taxDetails.ProviderTaxID = providerProfile.TaxID
	}

	invoice.TaxDetails = taxDetails
	invoice.recalculateTotals()
}

// GenerateInvoiceFromSettlement generates an invoice from a settlement record
func (g *InvoiceGenerator) GenerateInvoiceFromSettlement(
	settlementID string,
	escrowID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	totalAmount sdk.Coins,
	usageRecordIDs []string,
	periodStart time.Time,
	periodEnd time.Time,
	blockHeight int64,
	now time.Time,
) (*Invoice, error) {
	// Create usage inputs from the total amount
	currency := DefaultCurrency
	if len(totalAmount) > 0 {
		currency = totalAmount[0].Denom
	}

	// Create a single usage input for the settlement
	usageInputs := []UsageInput{
		{
			UsageRecordID: fmt.Sprintf("settlement-%s", settlementID),
			UsageType:     UsageTypeOther,
			Quantity:      sdkmath.LegacyOneDec(),
			Unit:          "unit",
			UnitPrice:     sdk.NewDecCoinFromCoin(totalAmount[0]),
			Description:   fmt.Sprintf("Settlement for order %s", orderID),
			PeriodStart:   periodStart,
			PeriodEnd:     periodEnd,
		},
	}

	req := InvoiceGenerationRequest{
		EscrowID:    escrowID,
		OrderID:     orderID,
		LeaseID:     leaseID,
		Provider:    provider,
		Customer:    customer,
		UsageInputs: usageInputs,
		BillingPeriod: BillingPeriod{
			StartTime:       periodStart,
			EndTime:         periodEnd,
			DurationSeconds: int64(periodEnd.Sub(periodStart).Seconds()),
			PeriodType:      BillingPeriodTypeUsageBased,
		},
		Currency: currency,
		Metadata: map[string]string{
			"settlement_id":    settlementID,
			"usage_record_ids": fmt.Sprintf("%v", usageRecordIDs),
		},
	}

	invoice, err := g.GenerateInvoice(req, blockHeight, now)
	if err != nil {
		return nil, err
	}

	invoice.SettlementID = settlementID
	return invoice, nil
}

// HPCUsageInput represents HPC-specific usage data
type HPCUsageInput struct {
	// JobID is the HPC job ID
	JobID string `json:"job_id"`

	// CPUHours is the CPU hours consumed
	CPUHours sdkmath.LegacyDec `json:"cpu_hours"`

	// GPUHours is the GPU hours consumed
	GPUHours sdkmath.LegacyDec `json:"gpu_hours"`

	// MemoryGBHours is the memory GB-hours consumed
	MemoryGBHours sdkmath.LegacyDec `json:"memory_gb_hours"`

	// StorageGBMonths is the storage GB-months consumed
	StorageGBMonths sdkmath.LegacyDec `json:"storage_gb_months"`

	// NetworkGB is the network transfer in GB
	NetworkGB sdkmath.LegacyDec `json:"network_gb"`

	// PeriodStart is the usage period start
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the usage period end
	PeriodEnd time.Time `json:"period_end"`
}

// GenerateHPCInvoice generates an invoice from HPC usage data
func (g *InvoiceGenerator) GenerateHPCInvoice(
	escrowID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	hpcUsage HPCUsageInput,
	pricingPolicy *PricingPolicy,
	blockHeight int64,
	now time.Time,
) (*Invoice, error) {
	currency := g.config.DefaultCurrency
	if pricingPolicy != nil {
		currency = pricingPolicy.DefaultCurrency
	}

	var usageInputs []UsageInput

	// Add CPU usage
	if hpcUsage.CPUHours.IsPositive() {
		cpuPricing := DefaultResourcePricing(currency)[UsageTypeCPU]
		if pricingPolicy != nil {
			if rp, ok := pricingPolicy.GetResourcePricing(UsageTypeCPU); ok {
				cpuPricing = rp
			}
		}

		usageInputs = append(usageInputs, UsageInput{
			UsageRecordID: fmt.Sprintf("%s-cpu", hpcUsage.JobID),
			UsageType:     UsageTypeCPU,
			Quantity:      hpcUsage.CPUHours,
			Unit:          cpuPricing.Unit,
			UnitPrice:     cpuPricing.BaseRate,
			Description:   fmt.Sprintf("CPU usage for job %s", hpcUsage.JobID),
			PeriodStart:   hpcUsage.PeriodStart,
			PeriodEnd:     hpcUsage.PeriodEnd,
		})
	}

	// Add GPU usage
	if hpcUsage.GPUHours.IsPositive() {
		gpuPricing := DefaultResourcePricing(currency)[UsageTypeGPU]
		if pricingPolicy != nil {
			if rp, ok := pricingPolicy.GetResourcePricing(UsageTypeGPU); ok {
				gpuPricing = rp
			}
		}

		usageInputs = append(usageInputs, UsageInput{
			UsageRecordID: fmt.Sprintf("%s-gpu", hpcUsage.JobID),
			UsageType:     UsageTypeGPU,
			Quantity:      hpcUsage.GPUHours,
			Unit:          gpuPricing.Unit,
			UnitPrice:     gpuPricing.BaseRate,
			Description:   fmt.Sprintf("GPU usage for job %s", hpcUsage.JobID),
			PeriodStart:   hpcUsage.PeriodStart,
			PeriodEnd:     hpcUsage.PeriodEnd,
		})
	}

	// Add Memory usage
	if hpcUsage.MemoryGBHours.IsPositive() {
		memPricing := DefaultResourcePricing(currency)[UsageTypeMemory]
		if pricingPolicy != nil {
			if rp, ok := pricingPolicy.GetResourcePricing(UsageTypeMemory); ok {
				memPricing = rp
			}
		}

		usageInputs = append(usageInputs, UsageInput{
			UsageRecordID: fmt.Sprintf("%s-mem", hpcUsage.JobID),
			UsageType:     UsageTypeMemory,
			Quantity:      hpcUsage.MemoryGBHours,
			Unit:          memPricing.Unit,
			UnitPrice:     memPricing.BaseRate,
			Description:   fmt.Sprintf("Memory usage for job %s", hpcUsage.JobID),
			PeriodStart:   hpcUsage.PeriodStart,
			PeriodEnd:     hpcUsage.PeriodEnd,
		})
	}

	// Add Storage usage
	if hpcUsage.StorageGBMonths.IsPositive() {
		storagePricing := DefaultResourcePricing(currency)[UsageTypeStorage]
		if pricingPolicy != nil {
			if rp, ok := pricingPolicy.GetResourcePricing(UsageTypeStorage); ok {
				storagePricing = rp
			}
		}

		usageInputs = append(usageInputs, UsageInput{
			UsageRecordID: fmt.Sprintf("%s-storage", hpcUsage.JobID),
			UsageType:     UsageTypeStorage,
			Quantity:      hpcUsage.StorageGBMonths,
			Unit:          storagePricing.Unit,
			UnitPrice:     storagePricing.BaseRate,
			Description:   fmt.Sprintf("Storage usage for job %s", hpcUsage.JobID),
			PeriodStart:   hpcUsage.PeriodStart,
			PeriodEnd:     hpcUsage.PeriodEnd,
		})
	}

	// Add Network usage
	if hpcUsage.NetworkGB.IsPositive() {
		netPricing := DefaultResourcePricing(currency)[UsageTypeNetwork]
		if pricingPolicy != nil {
			if rp, ok := pricingPolicy.GetResourcePricing(UsageTypeNetwork); ok {
				netPricing = rp
			}
		}

		usageInputs = append(usageInputs, UsageInput{
			UsageRecordID: fmt.Sprintf("%s-net", hpcUsage.JobID),
			UsageType:     UsageTypeNetwork,
			Quantity:      hpcUsage.NetworkGB,
			Unit:          netPricing.Unit,
			UnitPrice:     netPricing.BaseRate,
			Description:   fmt.Sprintf("Network usage for job %s", hpcUsage.JobID),
			PeriodStart:   hpcUsage.PeriodStart,
			PeriodEnd:     hpcUsage.PeriodEnd,
		})
	}

	if len(usageInputs) == 0 {
		return nil, fmt.Errorf("no usage data provided")
	}

	req := InvoiceGenerationRequest{
		EscrowID:    escrowID,
		OrderID:     orderID,
		LeaseID:     leaseID,
		Provider:    provider,
		Customer:    customer,
		UsageInputs: usageInputs,
		BillingPeriod: BillingPeriod{
			StartTime:       hpcUsage.PeriodStart,
			EndTime:         hpcUsage.PeriodEnd,
			DurationSeconds: int64(hpcUsage.PeriodEnd.Sub(hpcUsage.PeriodStart).Seconds()),
			PeriodType:      BillingPeriodTypeUsageBased,
		},
		Currency: currency,
		Metadata: map[string]string{
			"hpc_job_id": hpcUsage.JobID,
		},
	}

	return g.GenerateInvoice(req, blockHeight, now)
}
