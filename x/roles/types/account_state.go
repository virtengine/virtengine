package types

import (
	"fmt"
)

// AccountState represents the state of an account
type AccountState uint8

const (
	// AccountStateUnspecified is the default/invalid state
	AccountStateUnspecified AccountState = iota

	// AccountStateActive represents an active account
	AccountStateActive

	// AccountStateSuspended represents a temporarily suspended account
	AccountStateSuspended

	// AccountStateTerminated represents a permanently terminated account
	AccountStateTerminated
)

// String returns the string representation of an account state
func (s AccountState) String() string {
	switch s {
	case AccountStateUnspecified:
		return "unspecified"
	case AccountStateActive:
		return "active"
	case AccountStateSuspended:
		return "suspended"
	case AccountStateTerminated:
		return "terminated"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// AccountStateFromString converts a string to an AccountState
func AccountStateFromString(str string) (AccountState, error) {
	switch str {
	case "active":
		return AccountStateActive, nil
	case "suspended":
		return AccountStateSuspended, nil
	case "terminated":
		return AccountStateTerminated, nil
	default:
		return AccountStateUnspecified, fmt.Errorf("unknown account state: %s", str)
	}
}

// IsValid checks if the account state is valid
func (s AccountState) IsValid() bool {
	return s >= AccountStateActive && s <= AccountStateTerminated
}

// CanTransitionTo checks if the current state can transition to the target state
func (s AccountState) CanTransitionTo(target AccountState) bool {
	switch s {
	case AccountStateActive:
		// Active accounts can be suspended or terminated
		return target == AccountStateSuspended || target == AccountStateTerminated
	case AccountStateSuspended:
		// Suspended accounts can be reactivated or terminated
		return target == AccountStateActive || target == AccountStateTerminated
	case AccountStateTerminated:
		// Terminated accounts cannot transition (permanent)
		return false
	default:
		return false
	}
}

// IsOperational returns true if the account can perform normal operations
func (s AccountState) IsOperational() bool {
	return s == AccountStateActive
}

// AccountStateRecord represents the stored state of an account
type AccountStateRecord struct {
	Address       string       `json:"address"`
	State         AccountState `json:"state"`
	Reason        string       `json:"reason"`
	ModifiedBy    string       `json:"modified_by"`
	ModifiedAt    int64        `json:"modified_at"`
	PreviousState AccountState `json:"previous_state"`
}

// Validate validates the account state record
func (r AccountStateRecord) Validate() error {
	if r.Address == "" {
		return ErrInvalidAddress
	}
	if !r.State.IsValid() {
		return ErrInvalidAccountState
	}
	return nil
}

// DefaultAccountStateRecord returns the default account state record for a new account
func DefaultAccountStateRecord(address string) AccountStateRecord {
	return AccountStateRecord{
		Address:       address,
		State:         AccountStateActive,
		Reason:        "account created",
		ModifiedBy:    address,
		ModifiedAt:    0,
		PreviousState: AccountStateUnspecified,
	}
}

// AllAccountStates returns all valid account states
func AllAccountStates() []AccountState {
	return []AccountState{
		AccountStateActive,
		AccountStateSuspended,
		AccountStateTerminated,
	}
}
