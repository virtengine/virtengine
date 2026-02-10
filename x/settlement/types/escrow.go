package types

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EscrowState represents the state of an escrow account
type EscrowState string

const (
	EscrowStatePending  EscrowState = "pending"
	EscrowStateActive   EscrowState = "active"
	EscrowStateReleased EscrowState = "released"
	EscrowStateRefunded EscrowState = "refunded"
	EscrowStateDisputed EscrowState = "disputed"
	EscrowStateExpired  EscrowState = "expired"
)

// IsValidEscrowState checks if the state is valid
func IsValidEscrowState(state EscrowState) bool {
	switch state {
	case EscrowStatePending, EscrowStateActive, EscrowStateReleased,
		EscrowStateRefunded, EscrowStateDisputed, EscrowStateExpired:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the escrow is in a terminal state
func (s EscrowState) IsTerminal() bool {
	switch s {
	case EscrowStateReleased, EscrowStateRefunded, EscrowStateExpired:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if transition to the new state is valid
func (s EscrowState) CanTransitionTo(newState EscrowState) bool {
	switch s {
	case EscrowStatePending:
		return newState == EscrowStateActive || newState == EscrowStateRefunded || newState == EscrowStateExpired
	case EscrowStateActive:
		return newState == EscrowStateReleased || newState == EscrowStateRefunded ||
			newState == EscrowStateDisputed || newState == EscrowStateExpired
	case EscrowStateDisputed:
		return newState == EscrowStateReleased || newState == EscrowStateRefunded
	default:
		return false // Terminal states cannot transition
	}
}

// ReleaseConditionType defines types of release conditions
type ReleaseConditionType string

const (
	ConditionTypeTimelock       ReleaseConditionType = "timelock"
	ConditionTypeSignature      ReleaseConditionType = "signature"
	ConditionTypeUsageThreshold ReleaseConditionType = "usage_threshold"
	ConditionTypeVerification   ReleaseConditionType = "verification"
	ConditionTypeMultisig       ReleaseConditionType = "multisig"
)

// ReleaseCondition defines a condition that must be met for escrow release
type ReleaseCondition struct {
	// Type of the condition
	Type ReleaseConditionType `json:"type"`

	// Timelock conditions
	UnlockAfter *time.Time `json:"unlock_after,omitempty"`

	// Signature conditions
	RequiredSigners    []string `json:"required_signers,omitempty"`
	SignatureThreshold uint32   `json:"signature_threshold,omitempty"`

	// Usage threshold conditions
	MinUsageUnits uint64 `json:"min_usage_units,omitempty"`

	// Verification conditions
	RequiredVerificationScore uint32 `json:"required_verification_score,omitempty"`

	// Whether this condition has been satisfied
	Satisfied bool `json:"satisfied"`

	// Time when condition was satisfied
	SatisfiedAt *time.Time `json:"satisfied_at,omitempty"`
}

// Validate validates a release condition
func (c *ReleaseCondition) Validate() error {
	if c.Type == "" {
		return ErrInvalidCondition.Wrap("condition type cannot be empty")
	}

	switch c.Type {
	case ConditionTypeTimelock:
		if c.UnlockAfter == nil {
			return ErrInvalidCondition.Wrap("timelock condition requires unlock_after")
		}
	case ConditionTypeSignature, ConditionTypeMultisig:
		if len(c.RequiredSigners) == 0 {
			return ErrInvalidCondition.Wrap("signature condition requires signers")
		}
		if c.Type == ConditionTypeMultisig && c.SignatureThreshold == 0 {
			return ErrInvalidCondition.Wrap("multisig condition requires threshold")
		}
	case ConditionTypeUsageThreshold:
		if c.MinUsageUnits == 0 {
			return ErrInvalidCondition.Wrap("usage threshold condition requires min_usage_units")
		}
	case ConditionTypeVerification:
		// verification score can be 0, so no validation needed
	default:
		return ErrInvalidCondition.Wrapf("unknown condition type: %s", c.Type)
	}

	return nil
}

// EscrowAccount represents funds held in escrow for a marketplace order
type EscrowAccount struct {
	// EscrowID is the unique identifier for this escrow
	EscrowID string `json:"escrow_id"`

	// OrderID is the linked marketplace order
	OrderID string `json:"order_id"`

	// LeaseID is the linked marketplace lease (optional, set when lease is created)
	LeaseID string `json:"lease_id,omitempty"`

	// Depositor is the account that deposited funds
	Depositor string `json:"depositor"`

	// Recipient is the intended recipient (usually the provider)
	Recipient string `json:"recipient,omitempty"`

	// Amount is the total locked amount
	Amount sdk.Coins `json:"amount"`

	// Balance is the remaining balance (after partial settlements)
	Balance sdk.Coins `json:"balance"`

	// State is the current escrow state
	State EscrowState `json:"state"`

	// Conditions are the conditions that must be met for release
	Conditions []ReleaseCondition `json:"conditions,omitempty"`

	// CreatedAt is when the escrow was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the escrow expires
	ExpiresAt time.Time `json:"expires_at"`

	// ActivatedAt is when the escrow became active (optional)
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// ClosedAt is when the escrow was closed (released/refunded/expired)
	ClosedAt *time.Time `json:"closed_at,omitempty"`

	// Reason for closure (for refunds/disputes)
	ClosureReason string `json:"closure_reason,omitempty"`

	// TotalSettled is the total amount settled from this escrow
	TotalSettled sdk.Coins `json:"total_settled"`

	// SettlementCount is the number of settlements made
	SettlementCount uint32 `json:"settlement_count"`

	// BlockHeight is when the escrow was created
	BlockHeight int64 `json:"block_height"`
}

// NewEscrowAccount creates a new escrow account
func NewEscrowAccount(
	escrowID string,
	orderID string,
	depositor string,
	amount sdk.Coins,
	expiresAt time.Time,
	conditions []ReleaseCondition,
	blockTime time.Time,
	blockHeight int64,
) *EscrowAccount {
	return &EscrowAccount{
		EscrowID:        escrowID,
		OrderID:         orderID,
		Depositor:       depositor,
		Amount:          amount,
		Balance:         amount,
		State:           EscrowStatePending,
		Conditions:      conditions,
		CreatedAt:       blockTime,
		ExpiresAt:       expiresAt,
		TotalSettled:    sdk.NewCoins(),
		SettlementCount: 0,
		BlockHeight:     blockHeight,
	}
}

// Validate validates an escrow account
func (e *EscrowAccount) Validate() error {
	if e.EscrowID == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if len(e.EscrowID) > 64 {
		return ErrInvalidEscrow.Wrap("escrow_id exceeds maximum length")
	}

	if e.OrderID == "" {
		return ErrInvalidEscrow.Wrap("order_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(e.Depositor); err != nil {
		return ErrInvalidEscrow.Wrap("invalid depositor address")
	}

	if e.Recipient != "" {
		if _, err := sdk.AccAddressFromBech32(e.Recipient); err != nil {
			return ErrInvalidEscrow.Wrap("invalid recipient address")
		}
	}

	if !e.Amount.IsValid() || e.Amount.IsZero() {
		return ErrInvalidEscrow.Wrap("amount must be valid and non-zero")
	}

	if !e.Balance.IsValid() {
		return ErrInvalidEscrow.Wrap("balance must be valid")
	}

	if !IsValidEscrowState(e.State) {
		return ErrInvalidEscrow.Wrapf("invalid state: %s", e.State)
	}

	for i, cond := range e.Conditions {
		if err := cond.Validate(); err != nil {
			return ErrInvalidEscrow.Wrapf("invalid condition %d: %s", i, err.Error())
		}
	}

	if e.ExpiresAt.Before(e.CreatedAt) {
		return ErrInvalidEscrow.Wrap("expires_at must be after created_at")
	}

	return nil
}

// Activate transitions the escrow to active state
func (e *EscrowAccount) Activate(recipient string, blockTime time.Time) error {
	if !e.State.CanTransitionTo(EscrowStateActive) {
		return ErrInvalidStateTransition.Wrapf("cannot activate escrow in state %s", e.State)
	}

	if _, err := sdk.AccAddressFromBech32(recipient); err != nil {
		return ErrInvalidEscrow.Wrap("invalid recipient address")
	}

	e.State = EscrowStateActive
	e.Recipient = recipient
	e.ActivatedAt = &blockTime
	return nil
}

// Release transitions the escrow to released state
func (e *EscrowAccount) Release(blockTime time.Time, reason string) error {
	if !e.State.CanTransitionTo(EscrowStateReleased) {
		return ErrInvalidStateTransition.Wrapf("cannot release escrow in state %s", e.State)
	}

	e.State = EscrowStateReleased
	e.ClosedAt = &blockTime
	e.ClosureReason = reason
	return nil
}

// Refund transitions the escrow to refunded state
func (e *EscrowAccount) Refund(blockTime time.Time, reason string) error {
	if !e.State.CanTransitionTo(EscrowStateRefunded) {
		return ErrInvalidStateTransition.Wrapf("cannot refund escrow in state %s", e.State)
	}

	e.State = EscrowStateRefunded
	e.ClosedAt = &blockTime
	e.ClosureReason = reason
	return nil
}

// Dispute transitions the escrow to disputed state
func (e *EscrowAccount) Dispute(blockTime time.Time, reason string) error {
	if !e.State.CanTransitionTo(EscrowStateDisputed) {
		return ErrInvalidStateTransition.Wrapf("cannot dispute escrow in state %s", e.State)
	}

	e.State = EscrowStateDisputed
	e.ClosureReason = reason
	return nil
}

// CheckExpiry checks if escrow has expired and transitions state if so
func (e *EscrowAccount) CheckExpiry(blockTime time.Time) bool {
	if blockTime.After(e.ExpiresAt) && !e.State.IsTerminal() {
		e.State = EscrowStateExpired
		e.ClosedAt = &blockTime
		e.ClosureReason = "expired"
		return true
	}
	return false
}

// AllConditionsSatisfied checks if all conditions are satisfied
func (e *EscrowAccount) AllConditionsSatisfied() bool {
	for _, cond := range e.Conditions {
		if !cond.Satisfied {
			return false
		}
	}
	return true
}

// IsReleasable checks if the escrow can be released
func (e *EscrowAccount) IsReleasable() bool {
	return e.State == EscrowStateActive && e.AllConditionsSatisfied()
}

// DeductBalance deducts an amount from the balance
func (e *EscrowAccount) DeductBalance(amount sdk.Coins) error {
	if !e.Balance.IsAllGTE(amount) {
		return ErrInsufficientFunds.Wrap("insufficient escrow balance")
	}

	newBalance, hasNeg := e.Balance.SafeSub(amount...)
	if hasNeg {
		return ErrInsufficientFunds.Wrap("insufficient escrow balance")
	}

	e.Balance = newBalance
	e.TotalSettled = e.TotalSettled.Add(amount...)
	e.SettlementCount++
	return nil
}

// MarshalJSON implements json.Marshaler
func (e EscrowAccount) MarshalJSON() ([]byte, error) {
	type Alias EscrowAccount
	return json.Marshal(&struct {
		Alias
		Amount       []sdk.Coin `json:"amount"`
		Balance      []sdk.Coin `json:"balance"`
		TotalSettled []sdk.Coin `json:"total_settled"`
	}{
		Alias:        (Alias)(e),
		Amount:       e.Amount,
		Balance:      e.Balance,
		TotalSettled: e.TotalSettled,
	})
}
