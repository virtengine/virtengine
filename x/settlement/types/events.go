package types

import "fmt"

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
	EventTypeOrderSettled      = "order_settled"
	EventTypeSettlementCreated = "settlement_created"
	EventTypeUsageRecorded     = "usage_recorded"
	EventTypeUsageAcknowledged = "usage_acknowledged"

	// Reward events
	EventTypeRewardsDistributed             = "rewards_distributed"
	EventTypeRewardsClaimed                 = "rewards_claimed"
	EventTypeStakingRewardsDistributed      = "staking_rewards_distributed"
	EventTypeProviderRewardsDistributed     = "provider_rewards_distributed"
	EventTypeVerificationRewardsDistributed = "verification_rewards_distributed"

	// Fiat conversion events
	EventTypeFiatConversionRequested  = "fiat_conversion_requested"
	EventTypeFiatConversionCompleted  = "fiat_conversion_completed"
	EventTypeFiatConversionFailed     = "fiat_conversion_failed"
	EventTypeFiatConversionReconciled = "fiat_conversion_reconciled"
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
	AttributeKeyReason         = "reason"
	AttributeKeyEpochNumber    = "epoch_number"
	AttributeKeyRewardSource   = "reward_source"
	AttributeKeyRecipientCount = "recipient_count"
	AttributeKeyUsageUnits     = "usage_units"
	AttributeKeyUsageType      = "usage_type"
	AttributeKeySettlementType = "settlement_type"
	AttributeKeyIsFinal        = "is_final"
	AttributeKeyBlockHeight    = "block_height"

	// Fiat conversion attributes
	AttributeKeyConversionID  = "conversion_id"
	AttributeKeyFiatCurrency  = "fiat_currency"
	AttributeKeyFiatAmount    = "fiat_amount"
	AttributeKeyPaymentMethod = "payment_method"
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
	EscrowID   string `json:"escrow_id"`
	OrderID    string `json:"order_id"`
	Recipient  string `json:"recipient"`
	Amount     string `json:"amount"`
	Reason     string `json:"reason,omitempty"`
	ReleasedAt int64  `json:"released_at"`
}

// EventEscrowRefunded is emitted when an escrow is refunded
type EventEscrowRefunded struct {
	EscrowID   string `json:"escrow_id"`
	OrderID    string `json:"order_id"`
	Depositor  string `json:"depositor"`
	Amount     string `json:"amount"`
	Reason     string `json:"reason"`
	RefundedAt int64  `json:"refunded_at"`
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
	EscrowID  string `json:"escrow_id"`
	OrderID   string `json:"order_id"`
	Balance   string `json:"balance"`
	ExpiredAt int64  `json:"expired_at"`
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
	UsageID    string `json:"usage_id"`
	OrderID    string `json:"order_id"`
	LeaseID    string `json:"lease_id"`
	Provider   string `json:"provider"`
	UsageUnits uint64 `json:"usage_units"`
	UsageType  string `json:"usage_type"`
	TotalCost  string `json:"total_cost"`
	RecordedAt int64  `json:"recorded_at"`
}

// EventFiatConversionRequested is emitted when a fiat conversion is requested
type EventFiatConversionRequested struct {
	ConversionID  string `json:"conversion_id"`
	InvoiceID     string `json:"invoice_id,omitempty"`
	SettlementID  string `json:"settlement_id,omitempty"`
	Provider      string `json:"provider"`
	FiatCurrency  string `json:"fiat_currency"`
	PaymentMethod string `json:"payment_method"`
	RequestedAt   int64  `json:"requested_at"`
}

// EventFiatConversionCompleted is emitted when conversion completes
type EventFiatConversionCompleted struct {
	ConversionID string `json:"conversion_id"`
	Provider     string `json:"provider"`
	FiatCurrency string `json:"fiat_currency"`
	FiatAmount   string `json:"fiat_amount"`
	CompletedAt  int64  `json:"completed_at"`
}

// EventFiatConversionFailed is emitted when conversion fails
type EventFiatConversionFailed struct {
	ConversionID string `json:"conversion_id"`
	Provider     string `json:"provider"`
	Reason       string `json:"reason"`
	FailedAt     int64  `json:"failed_at"`
}

// EventFiatConversionReconciled is emitted when conversion status is reconciled
type EventFiatConversionReconciled struct {
	ConversionID string `json:"conversion_id"`
	Provider     string `json:"provider"`
	State        string `json:"state"`
	ReconciledAt int64  `json:"reconciled_at"`
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

// ProtoMessage stubs for Event types
func (*EventEscrowCreated) ProtoMessage()            {}
func (*EventEscrowActivated) ProtoMessage()          {}
func (*EventEscrowReleased) ProtoMessage()           {}
func (*EventEscrowRefunded) ProtoMessage()           {}
func (*EventEscrowDisputed) ProtoMessage()           {}
func (*EventEscrowExpired) ProtoMessage()            {}
func (*EventRewardsDistributed) ProtoMessage()       {}
func (*EventRewardsClaimed) ProtoMessage()           {}
func (*EventOrderSettled) ProtoMessage()             {}
func (*EventUsageRecorded) ProtoMessage()            {}
func (*EventFiatConversionRequested) ProtoMessage()  {}
func (*EventFiatConversionCompleted) ProtoMessage()  {}
func (*EventFiatConversionFailed) ProtoMessage()     {}
func (*EventFiatConversionReconciled) ProtoMessage() {}

// Reset stubs for Event types
func (e *EventEscrowCreated) Reset()            { *e = EventEscrowCreated{} }
func (e *EventEscrowActivated) Reset()          { *e = EventEscrowActivated{} }
func (e *EventEscrowReleased) Reset()           { *e = EventEscrowReleased{} }
func (e *EventEscrowRefunded) Reset()           { *e = EventEscrowRefunded{} }
func (e *EventEscrowDisputed) Reset()           { *e = EventEscrowDisputed{} }
func (e *EventEscrowExpired) Reset()            { *e = EventEscrowExpired{} }
func (e *EventRewardsDistributed) Reset()       { *e = EventRewardsDistributed{} }
func (e *EventRewardsClaimed) Reset()           { *e = EventRewardsClaimed{} }
func (e *EventOrderSettled) Reset()             { *e = EventOrderSettled{} }
func (e *EventUsageRecorded) Reset()            { *e = EventUsageRecorded{} }
func (e *EventFiatConversionRequested) Reset()  { *e = EventFiatConversionRequested{} }
func (e *EventFiatConversionCompleted) Reset()  { *e = EventFiatConversionCompleted{} }
func (e *EventFiatConversionFailed) Reset()     { *e = EventFiatConversionFailed{} }
func (e *EventFiatConversionReconciled) Reset() { *e = EventFiatConversionReconciled{} }

// String stubs for Fiat Conversion Event types
func (e *EventFiatConversionRequested) String() string  { return fmt.Sprintf("%+v", *e) }
func (e *EventFiatConversionCompleted) String() string  { return fmt.Sprintf("%+v", *e) }
func (e *EventFiatConversionFailed) String() string     { return fmt.Sprintf("%+v", *e) }
func (e *EventFiatConversionReconciled) String() string { return fmt.Sprintf("%+v", *e) }

// String stubs for Event types
func (e *EventEscrowCreated) String() string      { return fmt.Sprintf("%+v", *e) }
func (e *EventEscrowActivated) String() string    { return fmt.Sprintf("%+v", *e) }
func (e *EventEscrowReleased) String() string     { return fmt.Sprintf("%+v", *e) }
func (e *EventEscrowRefunded) String() string     { return fmt.Sprintf("%+v", *e) }
func (e *EventEscrowDisputed) String() string     { return fmt.Sprintf("%+v", *e) }
func (e *EventEscrowExpired) String() string      { return fmt.Sprintf("%+v", *e) }
func (e *EventRewardsDistributed) String() string { return fmt.Sprintf("%+v", *e) }
func (e *EventRewardsClaimed) String() string     { return fmt.Sprintf("%+v", *e) }
func (e *EventOrderSettled) String() string       { return fmt.Sprintf("%+v", *e) }
func (e *EventUsageRecorded) String() string      { return fmt.Sprintf("%+v", *e) }

// Payout events

// EventPayoutCompleted is emitted when a payout is completed
type EventPayoutCompleted struct {
	PayoutID     string `json:"payout_id"`
	SettlementID string `json:"settlement_id"`
	InvoiceID    string `json:"invoice_id,omitempty"`
	Provider     string `json:"provider"`
	NetAmount    string `json:"net_amount"`
	PlatformFee  string `json:"platform_fee"`
	CompletedAt  int64  `json:"completed_at"`
}

// EventPayoutHeld is emitted when a payout is held
type EventPayoutHeld struct {
	PayoutID  string `json:"payout_id"`
	DisputeID string `json:"dispute_id"`
	Reason    string `json:"reason"`
	HeldAt    int64  `json:"held_at"`
}

// EventPayoutReleased is emitted when a payout hold is released
type EventPayoutReleased struct {
	PayoutID   string `json:"payout_id"`
	ReleasedAt int64  `json:"released_at"`
}

// EventPayoutRefunded is emitted when a payout is refunded
type EventPayoutRefunded struct {
	PayoutID   string `json:"payout_id"`
	Customer   string `json:"customer"`
	Amount     string `json:"amount"`
	Reason     string `json:"reason"`
	RefundedAt int64  `json:"refunded_at"`
}

// EventPayoutFailed is emitted when a payout fails
type EventPayoutFailed struct {
	PayoutID string `json:"payout_id"`
	Error    string `json:"error"`
	FailedAt int64  `json:"failed_at"`
}

// ProtoMessage stubs for Payout Event types
func (*EventPayoutCompleted) ProtoMessage() {}
func (*EventPayoutHeld) ProtoMessage()      {}
func (*EventPayoutReleased) ProtoMessage()  {}
func (*EventPayoutRefunded) ProtoMessage()  {}
func (*EventPayoutFailed) ProtoMessage()    {}

// Reset stubs for Payout Event types
func (e *EventPayoutCompleted) Reset() { *e = EventPayoutCompleted{} }
func (e *EventPayoutHeld) Reset()      { *e = EventPayoutHeld{} }
func (e *EventPayoutReleased) Reset()  { *e = EventPayoutReleased{} }
func (e *EventPayoutRefunded) Reset()  { *e = EventPayoutRefunded{} }
func (e *EventPayoutFailed) Reset()    { *e = EventPayoutFailed{} }

// String stubs for Payout Event types
func (e *EventPayoutCompleted) String() string { return fmt.Sprintf("%+v", *e) }
func (e *EventPayoutHeld) String() string      { return fmt.Sprintf("%+v", *e) }
func (e *EventPayoutReleased) String() string  { return fmt.Sprintf("%+v", *e) }
func (e *EventPayoutRefunded) String() string  { return fmt.Sprintf("%+v", *e) }
func (e *EventPayoutFailed) String() string    { return fmt.Sprintf("%+v", *e) }
