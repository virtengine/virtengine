package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgCreateEscrow        = "create_escrow"
	TypeMsgActivateEscrow      = "activate_escrow"
	TypeMsgReleaseEscrow       = "release_escrow"
	TypeMsgRefundEscrow        = "refund_escrow"
	TypeMsgDisputeEscrow       = "dispute_escrow"
	TypeMsgSettleOrder         = "settle_order"
	TypeMsgRecordUsage         = "record_usage"
	TypeMsgAcknowledgeUsage    = "acknowledge_usage"
	TypeMsgClaimRewards        = "claim_rewards"
	TypeMsgDistributeRewards   = "distribute_rewards"
)

var (
	_ sdk.Msg = &MsgCreateEscrow{}
	_ sdk.Msg = &MsgActivateEscrow{}
	_ sdk.Msg = &MsgReleaseEscrow{}
	_ sdk.Msg = &MsgRefundEscrow{}
	_ sdk.Msg = &MsgDisputeEscrow{}
	_ sdk.Msg = &MsgSettleOrder{}
	_ sdk.Msg = &MsgRecordUsage{}
	_ sdk.Msg = &MsgAcknowledgeUsage{}
	_ sdk.Msg = &MsgClaimRewards{}
)

// MsgCreateEscrow creates a new escrow account
type MsgCreateEscrow struct {
	// Sender is the depositor creating the escrow
	Sender string `json:"sender"`

	// OrderID is the linked marketplace order
	OrderID string `json:"order_id"`

	// Amount is the amount to lock in escrow
	Amount sdk.Coins `json:"amount"`

	// ExpiresIn is the duration until expiry (in seconds)
	ExpiresIn uint64 `json:"expires_in"`

	// Conditions are the release conditions
	Conditions []ReleaseCondition `json:"conditions,omitempty"`
}

// NewMsgCreateEscrow creates a new MsgCreateEscrow
func NewMsgCreateEscrow(sender, orderID string, amount sdk.Coins, expiresIn uint64, conditions []ReleaseCondition) *MsgCreateEscrow {
	return &MsgCreateEscrow{
		Sender:     sender,
		OrderID:    orderID,
		Amount:     amount,
		ExpiresIn:  expiresIn,
		Conditions: conditions,
	}
}

func (msg MsgCreateEscrow) Route() string { return RouterKey }
func (msg MsgCreateEscrow) Type() string  { return TypeMsgCreateEscrow }

func (msg MsgCreateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderID == "" {
		return ErrInvalidEscrow.Wrap("order_id cannot be empty")
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidAmount.Wrap("amount must be valid and non-zero")
	}

	if msg.ExpiresIn == 0 {
		return ErrInvalidEscrow.Wrap("expires_in must be greater than zero")
	}

	for i, cond := range msg.Conditions {
		if err := cond.Validate(); err != nil {
			return ErrInvalidCondition.Wrapf("invalid condition %d: %s", i, err.Error())
		}
	}

	return nil
}

func (msg MsgCreateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgActivateEscrow activates an escrow when a lease is created
type MsgActivateEscrow struct {
	// Sender is the account activating the escrow (usually the system/market module)
	Sender string `json:"sender"`

	// EscrowID is the escrow to activate
	EscrowID string `json:"escrow_id"`

	// LeaseID is the linked lease
	LeaseID string `json:"lease_id"`

	// Recipient is the provider who will receive funds
	Recipient string `json:"recipient"`
}

// NewMsgActivateEscrow creates a new MsgActivateEscrow
func NewMsgActivateEscrow(sender, escrowID, leaseID, recipient string) *MsgActivateEscrow {
	return &MsgActivateEscrow{
		Sender:    sender,
		EscrowID:  escrowID,
		LeaseID:   leaseID,
		Recipient: recipient,
	}
}

func (msg MsgActivateEscrow) Route() string { return RouterKey }
func (msg MsgActivateEscrow) Type() string  { return TypeMsgActivateEscrow }

func (msg MsgActivateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowID == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.LeaseID == "" {
		return ErrInvalidEscrow.Wrap("lease_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrInvalidAddress.Wrap("invalid recipient address")
	}

	return nil
}

func (msg MsgActivateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgReleaseEscrow releases escrow funds to the recipient
type MsgReleaseEscrow struct {
	// Sender is the account releasing the escrow
	Sender string `json:"sender"`

	// EscrowID is the escrow to release
	EscrowID string `json:"escrow_id"`

	// Amount is the amount to release (if partial release, otherwise full balance)
	Amount sdk.Coins `json:"amount,omitempty"`

	// Reason for release
	Reason string `json:"reason,omitempty"`
}

// NewMsgReleaseEscrow creates a new MsgReleaseEscrow
func NewMsgReleaseEscrow(sender, escrowID string, amount sdk.Coins, reason string) *MsgReleaseEscrow {
	return &MsgReleaseEscrow{
		Sender:   sender,
		EscrowID: escrowID,
		Amount:   amount,
		Reason:   reason,
	}
}

func (msg MsgReleaseEscrow) Route() string { return RouterKey }
func (msg MsgReleaseEscrow) Type() string  { return TypeMsgReleaseEscrow }

func (msg MsgReleaseEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowID == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Amount != nil && !msg.Amount.IsValid() {
		return ErrInvalidAmount.Wrap("amount must be valid")
	}

	return nil
}

func (msg MsgReleaseEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgRefundEscrow refunds escrow funds to the depositor
type MsgRefundEscrow struct {
	// Sender is the account requesting the refund
	Sender string `json:"sender"`

	// EscrowID is the escrow to refund
	EscrowID string `json:"escrow_id"`

	// Reason for refund
	Reason string `json:"reason"`
}

// NewMsgRefundEscrow creates a new MsgRefundEscrow
func NewMsgRefundEscrow(sender, escrowID, reason string) *MsgRefundEscrow {
	return &MsgRefundEscrow{
		Sender:   sender,
		EscrowID: escrowID,
		Reason:   reason,
	}
}

func (msg MsgRefundEscrow) Route() string { return RouterKey }
func (msg MsgRefundEscrow) Type() string  { return TypeMsgRefundEscrow }

func (msg MsgRefundEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowID == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg MsgRefundEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgDisputeEscrow marks an escrow as disputed
type MsgDisputeEscrow struct {
	// Sender is the account filing the dispute
	Sender string `json:"sender"`

	// EscrowID is the escrow to dispute
	EscrowID string `json:"escrow_id"`

	// Reason for dispute
	Reason string `json:"reason"`

	// Evidence is optional evidence for the dispute
	Evidence string `json:"evidence,omitempty"`
}

// NewMsgDisputeEscrow creates a new MsgDisputeEscrow
func NewMsgDisputeEscrow(sender, escrowID, reason, evidence string) *MsgDisputeEscrow {
	return &MsgDisputeEscrow{
		Sender:   sender,
		EscrowID: escrowID,
		Reason:   reason,
		Evidence: evidence,
	}
}

func (msg MsgDisputeEscrow) Route() string { return RouterKey }
func (msg MsgDisputeEscrow) Type() string  { return TypeMsgDisputeEscrow }

func (msg MsgDisputeEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowID == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg MsgDisputeEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgSettleOrder settles an order based on usage records
type MsgSettleOrder struct {
	// Sender is the account initiating the settlement
	Sender string `json:"sender"`

	// OrderID is the order to settle
	OrderID string `json:"order_id"`

	// UsageRecordIDs are the usage records to include in this settlement
	UsageRecordIDs []string `json:"usage_record_ids,omitempty"`

	// IsFinal indicates if this is the final settlement
	IsFinal bool `json:"is_final"`
}

// NewMsgSettleOrder creates a new MsgSettleOrder
func NewMsgSettleOrder(sender, orderID string, usageRecordIDs []string, isFinal bool) *MsgSettleOrder {
	return &MsgSettleOrder{
		Sender:         sender,
		OrderID:        orderID,
		UsageRecordIDs: usageRecordIDs,
		IsFinal:        isFinal,
	}
}

func (msg MsgSettleOrder) Route() string { return RouterKey }
func (msg MsgSettleOrder) Type() string  { return TypeMsgSettleOrder }

func (msg MsgSettleOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderID == "" {
		return ErrInvalidSettlement.Wrap("order_id cannot be empty")
	}

	return nil
}

func (msg MsgSettleOrder) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgRecordUsage records usage from a provider
type MsgRecordUsage struct {
	// Sender is the provider recording usage
	Sender string `json:"sender"`

	// OrderID is the linked order
	OrderID string `json:"order_id"`

	// LeaseID is the linked lease
	LeaseID string `json:"lease_id"`

	// UsageUnits is the number of usage units
	UsageUnits uint64 `json:"usage_units"`

	// UsageType is the type of usage
	UsageType string `json:"usage_type"`

	// PeriodStart is the start of the usage period
	PeriodStart int64 `json:"period_start"`

	// PeriodEnd is the end of the usage period
	PeriodEnd int64 `json:"period_end"`

	// UnitPrice is the price per usage unit
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// Signature is the provider's signature
	Signature []byte `json:"signature"`

	// Metadata contains additional usage details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewMsgRecordUsage creates a new MsgRecordUsage
func NewMsgRecordUsage(
	sender, orderID, leaseID string,
	usageUnits uint64,
	usageType string,
	periodStart, periodEnd int64,
	unitPrice sdk.DecCoin,
	signature []byte,
) *MsgRecordUsage {
	return &MsgRecordUsage{
		Sender:      sender,
		OrderID:     orderID,
		LeaseID:     leaseID,
		UsageUnits:  usageUnits,
		UsageType:   usageType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		UnitPrice:   unitPrice,
		Signature:   signature,
	}
}

func (msg MsgRecordUsage) Route() string { return RouterKey }
func (msg MsgRecordUsage) Type() string  { return TypeMsgRecordUsage }

func (msg MsgRecordUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderID == "" {
		return ErrInvalidUsageRecord.Wrap("order_id cannot be empty")
	}

	if msg.LeaseID == "" {
		return ErrInvalidUsageRecord.Wrap("lease_id cannot be empty")
	}

	if msg.UsageUnits == 0 {
		return ErrInvalidUsageRecord.Wrap("usage_units must be greater than zero")
	}

	if msg.UsageType == "" {
		return ErrInvalidUsageRecord.Wrap("usage_type cannot be empty")
	}

	if msg.PeriodEnd <= msg.PeriodStart {
		return ErrInvalidUsageRecord.Wrap("period_end must be after period_start")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	return nil
}

func (msg MsgRecordUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgAcknowledgeUsage acknowledges a usage record
type MsgAcknowledgeUsage struct {
	// Sender is the customer acknowledging usage
	Sender string `json:"sender"`

	// UsageID is the usage record to acknowledge
	UsageID string `json:"usage_id"`

	// Signature is the customer's signature
	Signature []byte `json:"signature"`
}

// NewMsgAcknowledgeUsage creates a new MsgAcknowledgeUsage
func NewMsgAcknowledgeUsage(sender, usageID string, signature []byte) *MsgAcknowledgeUsage {
	return &MsgAcknowledgeUsage{
		Sender:    sender,
		UsageID:   usageID,
		Signature: signature,
	}
}

func (msg MsgAcknowledgeUsage) Route() string { return RouterKey }
func (msg MsgAcknowledgeUsage) Type() string  { return TypeMsgAcknowledgeUsage }

func (msg MsgAcknowledgeUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.UsageID == "" {
		return ErrInvalidUsageRecord.Wrap("usage_id cannot be empty")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	return nil
}

func (msg MsgAcknowledgeUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgClaimRewards claims accumulated rewards
type MsgClaimRewards struct {
	// Sender is the account claiming rewards
	Sender string `json:"sender"`

	// Source is the optional source to claim from (if empty, claims all)
	Source string `json:"source,omitempty"`
}

// NewMsgClaimRewards creates a new MsgClaimRewards
func NewMsgClaimRewards(sender, source string) *MsgClaimRewards {
	return &MsgClaimRewards{
		Sender: sender,
		Source: source,
	}
}

func (msg MsgClaimRewards) Route() string { return RouterKey }
func (msg MsgClaimRewards) Type() string  { return TypeMsgClaimRewards }

func (msg MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.Source != "" && !IsValidRewardSource(RewardSource(msg.Source)) {
		return ErrInvalidReward.Wrapf("invalid reward source: %s", msg.Source)
	}

	return nil
}

func (msg MsgClaimRewards) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// Message response types

// MsgCreateEscrowResponse is the response for MsgCreateEscrow
type MsgCreateEscrowResponse struct {
	EscrowID  string `json:"escrow_id"`
	CreatedAt int64  `json:"created_at"`
}

// MsgActivateEscrowResponse is the response for MsgActivateEscrow
type MsgActivateEscrowResponse struct {
	ActivatedAt int64 `json:"activated_at"`
}

// MsgReleaseEscrowResponse is the response for MsgReleaseEscrow
type MsgReleaseEscrowResponse struct {
	ReleasedAmount string `json:"released_amount"`
	ReleasedAt     int64  `json:"released_at"`
}

// MsgRefundEscrowResponse is the response for MsgRefundEscrow
type MsgRefundEscrowResponse struct {
	RefundedAmount string `json:"refunded_amount"`
	RefundedAt     int64  `json:"refunded_at"`
}

// MsgDisputeEscrowResponse is the response for MsgDisputeEscrow
type MsgDisputeEscrowResponse struct {
	DisputedAt int64 `json:"disputed_at"`
}

// MsgSettleOrderResponse is the response for MsgSettleOrder
type MsgSettleOrderResponse struct {
	SettlementID  string `json:"settlement_id"`
	TotalAmount   string `json:"total_amount"`
	ProviderShare string `json:"provider_share"`
	PlatformFee   string `json:"platform_fee"`
	SettledAt     int64  `json:"settled_at"`
}

// MsgRecordUsageResponse is the response for MsgRecordUsage
type MsgRecordUsageResponse struct {
	UsageID    string `json:"usage_id"`
	TotalCost  string `json:"total_cost"`
	RecordedAt int64  `json:"recorded_at"`
}

// MsgAcknowledgeUsageResponse is the response for MsgAcknowledgeUsage
type MsgAcknowledgeUsageResponse struct {
	AcknowledgedAt int64 `json:"acknowledged_at"`
}

// MsgClaimRewardsResponse is the response for MsgClaimRewards
type MsgClaimRewardsResponse struct {
	ClaimedAmount string `json:"claimed_amount"`
	ClaimedAt     int64  `json:"claimed_at"`
}
