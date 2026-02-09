package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "settlement"

	// StoreKey is the store key string for settlement module
	StoreKey = ModuleName

	// RouterKey is the message route for settlement module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for settlement module
	QuerierRoute = ModuleName

	// ModuleAccountName is the name for the module's account
	ModuleAccountName = ModuleName
)

// Store key prefixes
var (
	// PrefixEscrow is the prefix for escrow account storage
	// Key: PrefixEscrow | escrow_id -> EscrowAccount
	PrefixEscrow = []byte{0x01}

	// PrefixEscrowByOrder is the prefix for escrow lookup by order
	// Key: PrefixEscrowByOrder | order_id -> escrow_id
	PrefixEscrowByOrder = []byte{0x02}

	// PrefixEscrowByState is the prefix for escrow lookup by state
	// Key: PrefixEscrowByState | state | escrow_id -> nil
	PrefixEscrowByState = []byte{0x03}

	// PrefixSettlement is the prefix for settlement record storage
	// Key: PrefixSettlement | settlement_id -> SettlementRecord
	PrefixSettlement = []byte{0x04}

	// PrefixSettlementByEscrow is the prefix for settlement lookup by escrow
	// Key: PrefixSettlementByEscrow | escrow_id -> []settlement_id
	PrefixSettlementByEscrow = []byte{0x05}

	// PrefixSettlementByOrder is the prefix for settlement lookup by order
	// Key: PrefixSettlementByOrder | order_id -> settlement_id
	PrefixSettlementByOrder = []byte{0x06}

	// PrefixRewardDistribution is the prefix for reward distribution storage
	// Key: PrefixRewardDistribution | distribution_id -> RewardDistribution
	PrefixRewardDistribution = []byte{0x07}

	// PrefixRewardByEpoch is the prefix for reward lookup by epoch
	// Key: PrefixRewardByEpoch | epoch_number -> []distribution_id
	PrefixRewardByEpoch = []byte{0x08}

	// PrefixRewardByRecipient is the prefix for reward lookup by recipient
	// Key: PrefixRewardByRecipient | address -> []distribution_id
	PrefixRewardByRecipient = []byte{0x09}

	// PrefixUsageRecord is the prefix for usage record storage
	// Key: PrefixUsageRecord | usage_id -> UsageRecord
	PrefixUsageRecord = []byte{0x0A}

	// PrefixUsageByOrder is the prefix for usage lookup by order
	// Key: PrefixUsageByOrder | order_id -> []usage_id
	PrefixUsageByOrder = []byte{0x0B}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x0C}

	// PrefixEscrowSequence is the prefix for escrow sequence counter
	PrefixEscrowSequence = []byte{0x0D}

	// PrefixSettlementSequence is the prefix for settlement sequence counter
	PrefixSettlementSequence = []byte{0x0E}

	// PrefixDistributionSequence is the prefix for distribution sequence counter
	PrefixDistributionSequence = []byte{0x0F}

	// PrefixUsageSequence is the prefix for usage sequence counter
	PrefixUsageSequence = []byte{0x10}

	// PrefixClaimableRewards is the prefix for claimable rewards by address
	// Key: PrefixClaimableRewards | address -> ClaimableRewards
	PrefixClaimableRewards = []byte{0x11}

	// PrefixFiatConversion is the prefix for fiat conversion storage
	// Key: PrefixFiatConversion | conversion_id -> FiatConversionRecord
	PrefixFiatConversion = []byte{0x12}

	// PrefixFiatConversionByInvoice is the prefix for conversion lookup by invoice
	// Key: PrefixFiatConversionByInvoice | invoice_id -> conversion_id
	PrefixFiatConversionByInvoice = []byte{0x13}

	// PrefixFiatConversionBySettlement is the prefix for conversion lookup by settlement
	// Key: PrefixFiatConversionBySettlement | settlement_id -> conversion_id
	PrefixFiatConversionBySettlement = []byte{0x14}

	// PrefixFiatConversionByPayout is the prefix for conversion lookup by payout
	// Key: PrefixFiatConversionByPayout | payout_id -> conversion_id
	PrefixFiatConversionByPayout = []byte{0x15}

	// PrefixFiatConversionByProvider is the prefix for conversion lookup by provider
	// Key: PrefixFiatConversionByProvider | provider | conversion_id -> nil
	PrefixFiatConversionByProvider = []byte{0x16}

	// PrefixFiatConversionByState is the prefix for conversion lookup by state
	// Key: PrefixFiatConversionByState | state | conversion_id -> nil
	PrefixFiatConversionByState = []byte{0x17}

	// PrefixFiatConversionSequence is the prefix for conversion sequence counter
	PrefixFiatConversionSequence = []byte{0x18}

	// PrefixFiatPayoutPreference stores provider fiat payout preferences
	// Key: PrefixFiatPayoutPreference | provider -> FiatPayoutPreference
	PrefixFiatPayoutPreference = []byte{0x19}

	// PrefixFiatDailyTotals stores daily fiat conversion totals
	// Key: PrefixFiatDailyTotals | provider | yyyymmdd -> sdkmath.Int (bytes)
	PrefixFiatDailyTotals = []byte{0x1A}
)

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// EscrowKey returns the store key for an escrow account
func EscrowKey(escrowID string) []byte {
	key := make([]byte, 0, len(PrefixEscrow)+len(escrowID))
	key = append(key, PrefixEscrow...)
	key = append(key, []byte(escrowID)...)
	return key
}

// EscrowByOrderKey returns the store key for escrow lookup by order
func EscrowByOrderKey(orderID string) []byte {
	key := make([]byte, 0, len(PrefixEscrowByOrder)+len(orderID))
	key = append(key, PrefixEscrowByOrder...)
	key = append(key, []byte(orderID)...)
	return key
}

// EscrowByStateKey returns the store key for escrow lookup by state
func EscrowByStateKey(state EscrowState, escrowID string) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixEscrowByState)+len(stateBytes)+1+len(escrowID))
	key = append(key, PrefixEscrowByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	key = append(key, []byte(escrowID)...)
	return key
}

// EscrowByStatePrefixKey returns the prefix key for escrow lookup by state
func EscrowByStatePrefixKey(state EscrowState) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixEscrowByState)+len(stateBytes)+1)
	key = append(key, PrefixEscrowByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	return key
}

// SettlementKey returns the store key for a settlement record
func SettlementKey(settlementID string) []byte {
	key := make([]byte, 0, len(PrefixSettlement)+len(settlementID))
	key = append(key, PrefixSettlement...)
	key = append(key, []byte(settlementID)...)
	return key
}

// SettlementByEscrowKey returns the store key for settlement lookup by escrow
func SettlementByEscrowKey(escrowID string) []byte {
	key := make([]byte, 0, len(PrefixSettlementByEscrow)+len(escrowID))
	key = append(key, PrefixSettlementByEscrow...)
	key = append(key, []byte(escrowID)...)
	return key
}

// SettlementByOrderKey returns the store key for settlement lookup by order
func SettlementByOrderKey(orderID string) []byte {
	key := make([]byte, 0, len(PrefixSettlementByOrder)+len(orderID))
	key = append(key, PrefixSettlementByOrder...)
	key = append(key, []byte(orderID)...)
	return key
}

// RewardDistributionKey returns the store key for a reward distribution
func RewardDistributionKey(distributionID string) []byte {
	key := make([]byte, 0, len(PrefixRewardDistribution)+len(distributionID))
	key = append(key, PrefixRewardDistribution...)
	key = append(key, []byte(distributionID)...)
	return key
}

// RewardByEpochKey returns the store key for reward lookup by epoch
func RewardByEpochKey(epochNumber uint64) []byte {
	key := make([]byte, 0, len(PrefixRewardByEpoch)+8)
	key = append(key, PrefixRewardByEpoch...)
	key = appendUint64(key, epochNumber)
	return key
}

// RewardByRecipientKey returns the store key for reward lookup by recipient
func RewardByRecipientKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixRewardByRecipient)+len(address))
	key = append(key, PrefixRewardByRecipient...)
	key = append(key, address...)
	return key
}

// UsageRecordKey returns the store key for a usage record
func UsageRecordKey(usageID string) []byte {
	key := make([]byte, 0, len(PrefixUsageRecord)+len(usageID))
	key = append(key, PrefixUsageRecord...)
	key = append(key, []byte(usageID)...)
	return key
}

// UsageByOrderKey returns the store key for usage lookup by order
func UsageByOrderKey(orderID string) []byte {
	key := make([]byte, 0, len(PrefixUsageByOrder)+len(orderID))
	key = append(key, PrefixUsageByOrder...)
	key = append(key, []byte(orderID)...)
	return key
}

// ClaimableRewardsKey returns the store key for claimable rewards by address
func ClaimableRewardsKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixClaimableRewards)+len(address))
	key = append(key, PrefixClaimableRewards...)
	key = append(key, address...)
	return key
}

// FiatConversionKey returns the store key for a fiat conversion record
func FiatConversionKey(conversionID string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversion)+len(conversionID))
	key = append(key, PrefixFiatConversion...)
	key = append(key, []byte(conversionID)...)
	return key
}

// FiatConversionByInvoiceKey returns the store key for conversion lookup by invoice
func FiatConversionByInvoiceKey(invoiceID string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversionByInvoice)+len(invoiceID))
	key = append(key, PrefixFiatConversionByInvoice...)
	key = append(key, []byte(invoiceID)...)
	return key
}

// FiatConversionBySettlementKey returns the store key for conversion lookup by settlement
func FiatConversionBySettlementKey(settlementID string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversionBySettlement)+len(settlementID))
	key = append(key, PrefixFiatConversionBySettlement...)
	key = append(key, []byte(settlementID)...)
	return key
}

// FiatConversionByPayoutKey returns the store key for conversion lookup by payout
func FiatConversionByPayoutKey(payoutID string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversionByPayout)+len(payoutID))
	key = append(key, PrefixFiatConversionByPayout...)
	key = append(key, []byte(payoutID)...)
	return key
}

// FiatConversionByProviderKey returns the store key for conversion lookup by provider
func FiatConversionByProviderKey(provider string, conversionID string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversionByProvider)+len(provider)+1+len(conversionID))
	key = append(key, PrefixFiatConversionByProvider...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	key = append(key, []byte(conversionID)...)
	return key
}

// FiatConversionByProviderPrefixKey returns the prefix for conversion lookup by provider
func FiatConversionByProviderPrefixKey(provider string) []byte {
	key := make([]byte, 0, len(PrefixFiatConversionByProvider)+len(provider)+1)
	key = append(key, PrefixFiatConversionByProvider...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return key
}

// FiatConversionByStateKey returns the store key for conversion lookup by state
func FiatConversionByStateKey(state FiatConversionState, conversionID string) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixFiatConversionByState)+len(stateBytes)+1+len(conversionID))
	key = append(key, PrefixFiatConversionByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	key = append(key, []byte(conversionID)...)
	return key
}

// FiatConversionByStatePrefixKey returns the prefix for conversion lookup by state
func FiatConversionByStatePrefixKey(state FiatConversionState) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixFiatConversionByState)+len(stateBytes)+1)
	key = append(key, PrefixFiatConversionByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	return key
}

// FiatConversionSequenceKey returns the store key for conversion sequence
func FiatConversionSequenceKey() []byte {
	return PrefixFiatConversionSequence
}

// FiatPayoutPreferenceKey returns the store key for provider payout preferences
func FiatPayoutPreferenceKey(provider string) []byte {
	key := make([]byte, 0, len(PrefixFiatPayoutPreference)+len(provider))
	key = append(key, PrefixFiatPayoutPreference...)
	key = append(key, []byte(provider)...)
	return key
}

// FiatDailyTotalKey returns the store key for daily totals
func FiatDailyTotalKey(provider string, day string) []byte {
	key := make([]byte, 0, len(PrefixFiatDailyTotals)+len(provider)+1+len(day))
	key = append(key, PrefixFiatDailyTotals...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	key = append(key, []byte(day)...)
	return key
}

// EscrowSequenceKey returns the store key for escrow sequence
func EscrowSequenceKey() []byte {
	return PrefixEscrowSequence
}

// SettlementSequenceKey returns the store key for settlement sequence
func SettlementSequenceKey() []byte {
	return PrefixSettlementSequence
}

// DistributionSequenceKey returns the store key for distribution sequence
func DistributionSequenceKey() []byte {
	return PrefixDistributionSequence
}

// UsageSequenceKey returns the store key for usage sequence
func UsageSequenceKey() []byte {
	return PrefixUsageSequence
}

// appendUint64 appends a uint64 to a byte slice in big-endian order
func appendUint64(bz []byte, n uint64) []byte {
	for i := 7; i >= 0; i-- {
		bz = append(bz, byte(n>>(i*8)))
	}
	return bz
}
