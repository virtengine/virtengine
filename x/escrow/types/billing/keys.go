// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"encoding/binary"
	"fmt"
)

// Store key prefixes for billing types
var (
	// InvoicePrefix is the prefix for invoice storage
	InvoicePrefix = []byte{0x01}

	// InvoiceByProviderPrefix indexes invoices by provider
	InvoiceByProviderPrefix = []byte{0x02}

	// InvoiceByCustomerPrefix indexes invoices by customer
	InvoiceByCustomerPrefix = []byte{0x03}

	// InvoiceByStatusPrefix indexes invoices by status
	InvoiceByStatusPrefix = []byte{0x04}

	// InvoiceByEscrowPrefix indexes invoices by escrow ID
	InvoiceByEscrowPrefix = []byte{0x05}

	// DiscountPolicyPrefix is the prefix for discount policies
	DiscountPolicyPrefix = []byte{0x10}

	// CouponCodePrefix is the prefix for coupon codes
	CouponCodePrefix = []byte{0x11}

	// CouponByPolicyPrefix indexes coupons by policy
	CouponByPolicyPrefix = []byte{0x12}

	// LoyaltyProgramPrefix is the prefix for loyalty programs
	LoyaltyProgramPrefix = []byte{0x13}

	// CustomerLoyaltyPrefix is the prefix for customer loyalty
	CustomerLoyaltyPrefix = []byte{0x14}

	// TaxJurisdictionPrefix is the prefix for tax jurisdictions
	TaxJurisdictionPrefix = []byte{0x20}

	// CustomerTaxProfilePrefix is the prefix for customer tax profiles
	CustomerTaxProfilePrefix = []byte{0x21}

	// ProviderTaxProfilePrefix is the prefix for provider tax profiles
	ProviderTaxProfilePrefix = []byte{0x22}

	// PricingPolicyPrefix is the prefix for pricing policies
	PricingPolicyPrefix = []byte{0x30}

	// PricingPolicyByProviderPrefix indexes policies by provider
	PricingPolicyByProviderPrefix = []byte{0x31}

	// DisputeWindowPrefix is the prefix for dispute windows
	DisputeWindowPrefix = []byte{0x40}

	// DisputeByInvoicePrefix indexes disputes by invoice
	DisputeByInvoicePrefix = []byte{0x41}

	// DisputeByStatusPrefix indexes disputes by status
	DisputeByStatusPrefix = []byte{0x42}

	// SettlementConfigPrefix is the prefix for settlement config
	SettlementConfigPrefix = []byte{0x50}

	// SettlementHookResultPrefix is the prefix for hook results
	SettlementHookResultPrefix = []byte{0x51}
)

// BuildInvoiceKey builds the key for an invoice
func BuildInvoiceKey(invoiceID string) []byte {
	key := make([]byte, 0, len(InvoicePrefix)+len(invoiceID))
	key = append(key, InvoicePrefix...)
	return append(key, []byte(invoiceID)...)
}

// ParseInvoiceKey parses an invoice key
func ParseInvoiceKey(key []byte) (string, error) {
	if len(key) <= len(InvoicePrefix) {
		return "", fmt.Errorf("invalid invoice key length")
	}
	return string(key[len(InvoicePrefix):]), nil
}

// BuildInvoiceByProviderKey builds the index key for invoices by provider
func BuildInvoiceByProviderKey(provider string, invoiceID string) []byte {
	key := make([]byte, 0, len(InvoiceByProviderPrefix)+len(provider)+1+len(invoiceID))
	key = append(key, InvoiceByProviderPrefix...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return append(key, []byte(invoiceID)...)
}

// BuildInvoiceByProviderPrefix builds the prefix for provider's invoices
func BuildInvoiceByProviderPrefix(provider string) []byte {
	key := make([]byte, 0, len(InvoiceByProviderPrefix)+len(provider)+1)
	key = append(key, InvoiceByProviderPrefix...)
	key = append(key, []byte(provider)...)
	return append(key, byte('/'))
}

// BuildInvoiceByCustomerKey builds the index key for invoices by customer
func BuildInvoiceByCustomerKey(customer string, invoiceID string) []byte {
	key := make([]byte, 0, len(InvoiceByCustomerPrefix)+len(customer)+1+len(invoiceID))
	key = append(key, InvoiceByCustomerPrefix...)
	key = append(key, []byte(customer)...)
	key = append(key, byte('/'))
	return append(key, []byte(invoiceID)...)
}

// BuildInvoiceByCustomerPrefix builds the prefix for customer's invoices
func BuildInvoiceByCustomerPrefix(customer string) []byte {
	key := make([]byte, 0, len(InvoiceByCustomerPrefix)+len(customer)+1)
	key = append(key, InvoiceByCustomerPrefix...)
	key = append(key, []byte(customer)...)
	return append(key, byte('/'))
}

// BuildInvoiceByStatusKey builds the index key for invoices by status
func BuildInvoiceByStatusKey(status InvoiceStatus, invoiceID string) []byte {
	key := make([]byte, 0, len(InvoiceByStatusPrefix)+1+1+len(invoiceID))
	key = append(key, InvoiceByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(invoiceID)...)
}

// BuildInvoiceByStatusPrefix builds the prefix for invoices by status
func BuildInvoiceByStatusPrefix(status InvoiceStatus) []byte {
	key := make([]byte, 0, len(InvoiceByStatusPrefix)+1+1)
	key = append(key, InvoiceByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildInvoiceByEscrowKey builds the index key for invoices by escrow
func BuildInvoiceByEscrowKey(escrowID string, invoiceID string) []byte {
	key := make([]byte, 0, len(InvoiceByEscrowPrefix)+len(escrowID)+1+len(invoiceID))
	key = append(key, InvoiceByEscrowPrefix...)
	key = append(key, []byte(escrowID)...)
	key = append(key, byte('/'))
	return append(key, []byte(invoiceID)...)
}

// BuildInvoiceByEscrowPrefix builds the prefix for escrow's invoices
func BuildInvoiceByEscrowPrefix(escrowID string) []byte {
	key := make([]byte, 0, len(InvoiceByEscrowPrefix)+len(escrowID)+1)
	key = append(key, InvoiceByEscrowPrefix...)
	key = append(key, []byte(escrowID)...)
	return append(key, byte('/'))
}

// BuildDiscountPolicyKey builds the key for a discount policy
func BuildDiscountPolicyKey(policyID string) []byte {
	key := make([]byte, 0, len(DiscountPolicyPrefix)+len(policyID))
	key = append(key, DiscountPolicyPrefix...)
	return append(key, []byte(policyID)...)
}

// BuildCouponCodeKey builds the key for a coupon code
func BuildCouponCodeKey(code string) []byte {
	key := make([]byte, 0, len(CouponCodePrefix)+len(code))
	key = append(key, CouponCodePrefix...)
	return append(key, []byte(code)...)
}

// BuildCouponByPolicyKey builds the index key for coupons by policy
func BuildCouponByPolicyKey(policyID string, code string) []byte {
	key := make([]byte, 0, len(CouponByPolicyPrefix)+len(policyID)+1+len(code))
	key = append(key, CouponByPolicyPrefix...)
	key = append(key, []byte(policyID)...)
	key = append(key, byte('/'))
	return append(key, []byte(code)...)
}

// BuildLoyaltyProgramKey builds the key for a loyalty program
func BuildLoyaltyProgramKey(programID string) []byte {
	key := make([]byte, 0, len(LoyaltyProgramPrefix)+len(programID))
	key = append(key, LoyaltyProgramPrefix...)
	return append(key, []byte(programID)...)
}

// BuildCustomerLoyaltyKey builds the key for customer loyalty
func BuildCustomerLoyaltyKey(customer string, programID string) []byte {
	key := make([]byte, 0, len(CustomerLoyaltyPrefix)+len(customer)+1+len(programID))
	key = append(key, CustomerLoyaltyPrefix...)
	key = append(key, []byte(customer)...)
	key = append(key, byte('/'))
	return append(key, []byte(programID)...)
}

// BuildTaxJurisdictionKey builds the key for a tax jurisdiction
func BuildTaxJurisdictionKey(countryCode string) []byte {
	key := make([]byte, 0, len(TaxJurisdictionPrefix)+len(countryCode))
	key = append(key, TaxJurisdictionPrefix...)
	return append(key, []byte(countryCode)...)
}

// BuildTaxJurisdictionRegionKey builds the key for a regional tax jurisdiction
func BuildTaxJurisdictionRegionKey(countryCode string, regionCode string) []byte {
	key := make([]byte, 0, len(TaxJurisdictionPrefix)+len(countryCode)+1+len(regionCode))
	key = append(key, TaxJurisdictionPrefix...)
	key = append(key, []byte(countryCode)...)
	key = append(key, byte('/'))
	return append(key, []byte(regionCode)...)
}

// BuildCustomerTaxProfileKey builds the key for a customer tax profile
func BuildCustomerTaxProfileKey(customer string) []byte {
	key := make([]byte, 0, len(CustomerTaxProfilePrefix)+len(customer))
	key = append(key, CustomerTaxProfilePrefix...)
	return append(key, []byte(customer)...)
}

// BuildProviderTaxProfileKey builds the key for a provider tax profile
func BuildProviderTaxProfileKey(provider string) []byte {
	key := make([]byte, 0, len(ProviderTaxProfilePrefix)+len(provider))
	key = append(key, ProviderTaxProfilePrefix...)
	return append(key, []byte(provider)...)
}

// BuildPricingPolicyKey builds the key for a pricing policy
func BuildPricingPolicyKey(policyID string) []byte {
	key := make([]byte, 0, len(PricingPolicyPrefix)+len(policyID))
	key = append(key, PricingPolicyPrefix...)
	return append(key, []byte(policyID)...)
}

// BuildPricingPolicyByProviderKey builds the index key for policies by provider
func BuildPricingPolicyByProviderKey(provider string, policyID string) []byte {
	key := make([]byte, 0, len(PricingPolicyByProviderPrefix)+len(provider)+1+len(policyID))
	key = append(key, PricingPolicyByProviderPrefix...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return append(key, []byte(policyID)...)
}

// BuildPricingPolicyByProviderPrefix builds the prefix for provider's policies
func BuildPricingPolicyByProviderPrefix(provider string) []byte {
	key := make([]byte, 0, len(PricingPolicyByProviderPrefix)+len(provider)+1)
	key = append(key, PricingPolicyByProviderPrefix...)
	key = append(key, []byte(provider)...)
	return append(key, byte('/'))
}

// BuildDisputeWindowKey builds the key for a dispute window
func BuildDisputeWindowKey(windowID string) []byte {
	key := make([]byte, 0, len(DisputeWindowPrefix)+len(windowID))
	key = append(key, DisputeWindowPrefix...)
	return append(key, []byte(windowID)...)
}

// BuildDisputeByInvoiceKey builds the index key for disputes by invoice
func BuildDisputeByInvoiceKey(invoiceID string, windowID string) []byte {
	key := make([]byte, 0, len(DisputeByInvoicePrefix)+len(invoiceID)+1+len(windowID))
	key = append(key, DisputeByInvoicePrefix...)
	key = append(key, []byte(invoiceID)...)
	key = append(key, byte('/'))
	return append(key, []byte(windowID)...)
}

// BuildDisputeByStatusKey builds the index key for disputes by status
func BuildDisputeByStatusKey(status DisputeStatus, windowID string) []byte {
	key := make([]byte, 0, len(DisputeByStatusPrefix)+1+1+len(windowID))
	key = append(key, DisputeByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(windowID)...)
}

// BuildDisputeByStatusPrefix builds the prefix for disputes by status
func BuildDisputeByStatusPrefix(status DisputeStatus) []byte {
	key := make([]byte, 0, len(DisputeByStatusPrefix)+1+1)
	key = append(key, DisputeByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildSettlementHookResultKey builds the key for a hook result
func BuildSettlementHookResultKey(settlementID string, hookID string, timestamp int64) []byte {
	key := make([]byte, 0, len(SettlementHookResultPrefix)+len(settlementID)+1+len(hookID)+1+8)
	key = append(key, SettlementHookResultPrefix...)
	key = append(key, []byte(settlementID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(hookID)...)
	key = append(key, byte('/'))

	// Append timestamp as big-endian uint64
	tsBytes := make([]byte, 8)
	if timestamp < 0 {
		timestamp = 0
	}
	//nolint:gosec // timestamp checked for negativity
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	return append(key, tsBytes...)
}

// InvoiceSequenceKey is the key for invoice number sequence
var InvoiceSequenceKey = []byte("invoice_sequence")

// NextInvoiceNumber generates the next invoice number
func NextInvoiceNumber(currentSequence uint64, prefix string) string {
	return fmt.Sprintf("%s-%08d", prefix, currentSequence+1)
}

// InvoiceNumberFromID generates a deterministic invoice number from an invoice ID.
func InvoiceNumberFromID(invoiceID string, prefix string) string {
	if invoiceID == "" {
		return NextInvoiceNumber(0, prefix)
	}

	if len(invoiceID) > 12 {
		invoiceID = invoiceID[:12]
	}

	return fmt.Sprintf("%s-%s", prefix, invoiceID)
}

// Settlement and treasury key prefixes
var (
	// SettlementRecordPrefix is the prefix for settlement records
	SettlementRecordPrefix = []byte{0x60}

	// SettlementByInvoicePrefix indexes settlements by invoice
	SettlementByInvoicePrefix = []byte{0x61}

	// SettlementByProviderPrefix indexes settlements by provider
	SettlementByProviderPrefix = []byte{0x62}

	// SettlementByStatusPrefix indexes settlements by status
	SettlementByStatusPrefix = []byte{0x63}

	// SettlementByEscrowPrefix indexes settlements by escrow
	SettlementByEscrowPrefix = []byte{0x64}

	// TreasuryAllocationPrefix is the prefix for treasury allocations
	TreasuryAllocationPrefix = []byte{0x70}

	// TreasuryAllocationBySettlementPrefix indexes allocations by settlement
	TreasuryAllocationBySettlementPrefix = []byte{0x71}

	// TreasuryAllocationByFeeTypePrefix indexes allocations by fee type
	TreasuryAllocationByFeeTypePrefix = []byte{0x72}

	// FeeConfigKey is the key for fee configuration
	FeeConfigKey = []byte("fee_config")

	// SettlementSequenceKey is the key for settlement sequence
	SettlementSequenceKey = []byte("settlement_sequence")
)

// BuildSettlementRecordKey builds the key for a settlement record
func BuildSettlementRecordKey(settlementID string) []byte {
	key := make([]byte, 0, len(SettlementRecordPrefix)+len(settlementID))
	key = append(key, SettlementRecordPrefix...)
	return append(key, []byte(settlementID)...)
}

// BuildSettlementByInvoiceKey builds the index key for settlements by invoice
func BuildSettlementByInvoiceKey(invoiceID string, settlementID string) []byte {
	key := make([]byte, 0, len(SettlementByInvoicePrefix)+len(invoiceID)+1+len(settlementID))
	key = append(key, SettlementByInvoicePrefix...)
	key = append(key, []byte(invoiceID)...)
	key = append(key, byte('/'))
	return append(key, []byte(settlementID)...)
}

// BuildSettlementByInvoicePrefix builds the prefix for invoice's settlements
func BuildSettlementByInvoicePrefix(invoiceID string) []byte {
	key := make([]byte, 0, len(SettlementByInvoicePrefix)+len(invoiceID)+1)
	key = append(key, SettlementByInvoicePrefix...)
	key = append(key, []byte(invoiceID)...)
	return append(key, byte('/'))
}

// BuildSettlementByProviderKey builds the index key for settlements by provider
func BuildSettlementByProviderKey(provider string, settlementID string) []byte {
	key := make([]byte, 0, len(SettlementByProviderPrefix)+len(provider)+1+len(settlementID))
	key = append(key, SettlementByProviderPrefix...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return append(key, []byte(settlementID)...)
}

// BuildSettlementByProviderPrefix builds the prefix for provider's settlements
func BuildSettlementByProviderPrefix(provider string) []byte {
	key := make([]byte, 0, len(SettlementByProviderPrefix)+len(provider)+1)
	key = append(key, SettlementByProviderPrefix...)
	key = append(key, []byte(provider)...)
	return append(key, byte('/'))
}

// BuildSettlementByStatusKey builds the index key for settlements by status
func BuildSettlementByStatusKey(status SettlementStatus, settlementID string) []byte {
	key := make([]byte, 0, len(SettlementByStatusPrefix)+1+1+len(settlementID))
	key = append(key, SettlementByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(settlementID)...)
}

// BuildSettlementByStatusPrefix builds the prefix for settlements by status
func BuildSettlementByStatusPrefix(status SettlementStatus) []byte {
	key := make([]byte, 0, len(SettlementByStatusPrefix)+1+1)
	key = append(key, SettlementByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildSettlementByEscrowKey builds the index key for settlements by escrow
func BuildSettlementByEscrowKey(escrowID string, settlementID string) []byte {
	key := make([]byte, 0, len(SettlementByEscrowPrefix)+len(escrowID)+1+len(settlementID))
	key = append(key, SettlementByEscrowPrefix...)
	key = append(key, []byte(escrowID)...)
	key = append(key, byte('/'))
	return append(key, []byte(settlementID)...)
}

// BuildSettlementByEscrowPrefix builds the prefix for escrow's settlements
func BuildSettlementByEscrowPrefix(escrowID string) []byte {
	key := make([]byte, 0, len(SettlementByEscrowPrefix)+len(escrowID)+1)
	key = append(key, SettlementByEscrowPrefix...)
	key = append(key, []byte(escrowID)...)
	return append(key, byte('/'))
}

// BuildTreasuryAllocationKey builds the key for a treasury allocation
func BuildTreasuryAllocationKey(allocationID string) []byte {
	key := make([]byte, 0, len(TreasuryAllocationPrefix)+len(allocationID))
	key = append(key, TreasuryAllocationPrefix...)
	return append(key, []byte(allocationID)...)
}

// BuildTreasuryAllocationBySettlementKey builds the index key for allocations by settlement
func BuildTreasuryAllocationBySettlementKey(settlementID string, allocationID string) []byte {
	key := make([]byte, 0, len(TreasuryAllocationBySettlementPrefix)+len(settlementID)+1+len(allocationID))
	key = append(key, TreasuryAllocationBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	key = append(key, byte('/'))
	return append(key, []byte(allocationID)...)
}

// BuildTreasuryAllocationBySettlementPrefix builds the prefix for settlement's allocations
func BuildTreasuryAllocationBySettlementPrefix(settlementID string) []byte {
	key := make([]byte, 0, len(TreasuryAllocationBySettlementPrefix)+len(settlementID)+1)
	key = append(key, TreasuryAllocationBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	return append(key, byte('/'))
}

// BuildTreasuryAllocationByFeeTypeKey builds the index key for allocations by fee type
func BuildTreasuryAllocationByFeeTypeKey(feeType FeeType, allocationID string) []byte {
	key := make([]byte, 0, len(TreasuryAllocationByFeeTypePrefix)+1+1+len(allocationID))
	key = append(key, TreasuryAllocationByFeeTypePrefix...)
	key = append(key, byte(feeType))
	key = append(key, byte('/'))
	return append(key, []byte(allocationID)...)
}

// BuildTreasuryAllocationByFeeTypePrefix builds the prefix for allocations by fee type
func BuildTreasuryAllocationByFeeTypePrefix(feeType FeeType) []byte {
	key := make([]byte, 0, len(TreasuryAllocationByFeeTypePrefix)+1+1)
	key = append(key, TreasuryAllocationByFeeTypePrefix...)
	key = append(key, byte(feeType))
	return append(key, byte('/'))
}

// NextSettlementID generates the next settlement ID
func NextSettlementID(currentSequence uint64, prefix string) string {
	return fmt.Sprintf("%s-STL-%08d", prefix, currentSequence+1)
}

// NextTreasuryAllocationID generates the next treasury allocation ID
func NextTreasuryAllocationID(settlementID string, feeType FeeType) string {
	return fmt.Sprintf("%s-%s", settlementID, feeType.String())
}
