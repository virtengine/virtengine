package types

// Event types for the settlement module
const (
	// Escrow events
	EventTypeEscrowCreated   = "escrow_created"
	EventTypeEscrowActivated = "escrow_activated"
	EventTypeEscrowReleased  = "escrow_released"
	EventTypeEscrowRefunded  = "escrow_refunded"
	EventTypeEscrowDisputed  = "escrow_disputed"
	EventTypeEscrowExpired   = "escrow_expired"

	// Settlement events
	EventTypeOrderSettled       = "order_settled"
	EventTypeSettlementCreated  = "settlement_created"
	EventTypeUsageRecorded      = "usage_recorded"
	EventTypeUsageAcknowledged  = "usage_acknowledged"

	// Reward events
	EventTypeRewardsDistributed     = "rewards_distributed"
	EventTypeRewardsClaimed         = "rewards_claimed"
	EventTypeStakingRewardsDistributed = "staking_rewards_distributed"
	EventTypeProviderRewardsDistributed = "provider_rewards_distributed"
	EventTypeVerificationRewardsDistributed = "verification_rewards_distributed"
)

// Event attribute keys
const (
	// Common attributes
	AttributeKeyEscrowID       = "escrow_id"
	AttributeKeyOrderID        = "order_id"
	AttributeKeyLeaseID        = "lease_id"
	AttributeKeySettlementID   = "settlement_id"
	AttributeKeyDistributionID = "distribution_id"
	AttributeKeyUsageID        = "usage_id"

	// Address attributes
	AttributeKeyDepositor = "depositor"
	AttributeKeyRecipient = "recipient"
	AttributeKeyProvider  = "provider"
	AttributeKeyCustomer  = "customer"
	AttributeKeyClaimer   = "claimer"

	// Amount attributes
	AttributeKeyAmount        = "amount"
	AttributeKeyBalance       = "balance"
	AttributeKeyProviderShare = "provider_share"
	AttributeKeyPlatformFee   = "platform_fee"
	AttributeKeyValidatorFee  = "validator_fee"
	AttributeKeyTotalRewards  = "total_rewards"
	AttributeKeyClaimedAmount = "claimed_amount"

	// State attributes
	AttributeKeyState         = "state"
	AttributeKeyPreviousState = "previous_state"
	AttributeKeyNewState      = "new_state"

	// Time attributes
	AttributeKeyTimestamp   = "timestamp"
	AttributeKeyExpiresAt   = "expires_at"
	AttributeKeyPeriodStart = "period_start"
	AttributeKeyPeriodEnd   = "period_end"

	// Other attributes
	AttributeKeyReason           = "reason"
	AttributeKeyEpochNumber      = "epoch_number"
	AttributeKeyRewardSource     = "reward_source"
	AttributeKeyRecipientCount   = "recipient_count"
	AttributeKeyUsageUnits       = "usage_units"
	AttributeKeyUsageType        = "usage_type"
	AttributeKeySettlementType   = "settlement_type"
	AttributeKeyIsFinal          = "is_final"
	AttributeKeyBlockHeight      = "block_height"
)

// EventEscrowCreated is emitted when an escrow is created
type EventEscrowCreated struct {
	EscrowID    string `json:"escrow_id"`
	OrderID     string `json:"order_id"`
	Depositor   string `json:"depositor"`
	Amount      string `json:"amount"`
	ExpiresAt   int64  `json:"expires_at"`
	BlockHeight int64  `json:"block_height"`
}

// EventEscrowActivated is emitted when an escrow becomes active
type EventEscrowActivated struct {
	EscrowID    string `json:"escrow_id"`
	OrderID     string `json:"order_id"`
	LeaseID     string `json:"lease_id"`
	Recipient   string `json:"recipient"`
	ActivatedAt int64  `json:"activated_at"`
}

// EventEscrowReleased is emitted when an escrow is released
type EventEscrowReleased struct {
	EscrowID    string `json:"escrow_id"`
	OrderID     string `json:"order_id"`
	Recipient   string `json:"recipient"`
	Amount      string `json:"amount"`
	Reason      string `json:"reason,omitempty"`
	ReleasedAt  int64  `json:"released_at"`
}

// EventEscrowRefunded is emitted when an escrow is refunded
type EventEscrowRefunded struct {
	EscrowID    string `json:"escrow_id"`
	OrderID     string `json:"order_id"`
	Depositor   string `json:"depositor"`
	Amount      string `json:"amount"`
	Reason      string `json:"reason"`
	RefundedAt  int64  `json:"refunded_at"`
}

// EventEscrowDisputed is emitted when an escrow is disputed
type EventEscrowDisputed struct {
	EscrowID   string `json:"escrow_id"`
	OrderID    string `json:"order_id"`
	Reason     string `json:"reason"`
	DisputedAt int64  `json:"disputed_at"`
}

// EventEscrowExpired is emitted when an escrow expires
type EventEscrowExpired struct {
	EscrowID   string `json:"escrow_id"`
	OrderID    string `json:"order_id"`
	Balance    string `json:"balance"`
	ExpiredAt  int64  `json:"expired_at"`
}

// EventOrderSettled is emitted when an order is settled
type EventOrderSettled struct {
	SettlementID   string `json:"settlement_id"`
	OrderID        string `json:"order_id"`
	EscrowID       string `json:"escrow_id"`
	Provider       string `json:"provider"`
	Customer       string `json:"customer"`
	TotalAmount    string `json:"total_amount"`
	ProviderShare  string `json:"provider_share"`
	PlatformFee    string `json:"platform_fee"`
	SettlementType string `json:"settlement_type"`
	IsFinal        bool   `json:"is_final"`
	SettledAt      int64  `json:"settled_at"`
}

// EventUsageRecorded is emitted when usage is recorded
type EventUsageRecorded struct {
	UsageID     string `json:"usage_id"`
	OrderID     string `json:"order_id"`
	LeaseID     string `json:"lease_id"`
	Provider    string `json:"provider"`
	UsageUnits  uint64 `json:"usage_units"`
	UsageType   string `json:"usage_type"`
	TotalCost   string `json:"total_cost"`
	RecordedAt  int64  `json:"recorded_at"`
}

// EventRewardsDistributed is emitted when rewards are distributed
type EventRewardsDistributed struct {
	DistributionID string `json:"distribution_id"`
	EpochNumber    uint64 `json:"epoch_number"`
	Source         string `json:"source"`
	TotalRewards   string `json:"total_rewards"`
	RecipientCount uint32 `json:"recipient_count"`
	DistributedAt  int64  `json:"distributed_at"`
}

// EventRewardsClaimed is emitted when rewards are claimed
type EventRewardsClaimed struct {
	Claimer       string `json:"claimer"`
	ClaimedAmount string `json:"claimed_amount"`
	Source        string `json:"source,omitempty"`
	ClaimedAt     int64  `json:"claimed_at"`
}
