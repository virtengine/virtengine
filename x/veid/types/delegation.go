package types

import (
	"fmt"
	"time"
)

// DelegationPermission represents a permission that can be delegated
type DelegationPermission int

const (
	// PermissionViewIdentity allows viewing identity information
	PermissionViewIdentity DelegationPermission = iota
	// PermissionProveIdentity allows proving identity claims
	PermissionProveIdentity
	// PermissionSignOnBehalf allows signing transactions on behalf of delegator
	PermissionSignOnBehalf
	// PermissionManageScopes allows managing identity scopes
	PermissionManageScopes
)

// String returns the string representation of a DelegationPermission
func (p DelegationPermission) String() string {
	switch p {
	case PermissionViewIdentity:
		return "view_identity"
	case PermissionProveIdentity:
		return "prove_identity"
	case PermissionSignOnBehalf:
		return "sign_on_behalf"
	case PermissionManageScopes:
		return "manage_scopes"
	default:
		return fmt.Sprintf("unknown_permission_%d", p)
	}
}

// IsValid returns true if the permission is valid
func (p DelegationPermission) IsValid() bool {
	return p >= PermissionViewIdentity && p <= PermissionManageScopes
}

// ParseDelegationPermission parses a string into a DelegationPermission
func ParseDelegationPermission(s string) (DelegationPermission, error) {
	switch s {
	case "view_identity":
		return PermissionViewIdentity, nil
	case "prove_identity":
		return PermissionProveIdentity, nil
	case "sign_on_behalf":
		return PermissionSignOnBehalf, nil
	case "manage_scopes":
		return PermissionManageScopes, nil
	default:
		return 0, fmt.Errorf("invalid delegation permission: %s", s)
	}
}

// DelegationStatus represents the status of a delegation
type DelegationStatus int

const (
	// DelegationActive means the delegation is currently valid
	DelegationActive DelegationStatus = iota
	// DelegationExpired means the delegation has passed its expiration time
	DelegationExpired
	// DelegationRevoked means the delegation was revoked by the delegator
	DelegationRevoked
	// DelegationExhausted means the delegation has used all available uses
	DelegationExhausted
)

// String returns the string representation of a DelegationStatus
func (s DelegationStatus) String() string {
	switch s {
	case DelegationActive:
		return "active"
	case DelegationExpired:
		return "expired"
	case DelegationRevoked:
		return "revoked"
	case DelegationExhausted:
		return "exhausted"
	default:
		return fmt.Sprintf("unknown_status_%d", s)
	}
}

// IsValid returns true if the status is valid
func (s DelegationStatus) IsValid() bool {
	return s >= DelegationActive && s <= DelegationExhausted
}

// IsTerminal returns true if the status is a terminal state (cannot be changed back to active)
func (s DelegationStatus) IsTerminal() bool {
	return s == DelegationRevoked || s == DelegationExhausted
}

// DelegationRecord represents a delegation of identity permissions from one account to another
type DelegationRecord struct {
	// DelegationID is the unique identifier for this delegation
	DelegationID string `json:"delegation_id"`

	// DelegatorAddress is the identity owner who grants the delegation
	DelegatorAddress string `json:"delegator_address"`

	// DelegateAddress is the account that receives the delegation
	DelegateAddress string `json:"delegate_address"`

	// Permissions is the list of permissions granted
	Permissions []DelegationPermission `json:"permissions"`

	// ExpiresAt is when the delegation expires
	ExpiresAt time.Time `json:"expires_at"`

	// MaxUses is the maximum number of times the delegation can be used (0 = unlimited)
	MaxUses uint64 `json:"max_uses"`

	// UsesRemaining is the number of uses remaining (decremented on each use)
	UsesRemaining uint64 `json:"uses_remaining"`

	// CreatedAt is when the delegation was created
	CreatedAt time.Time `json:"created_at"`

	// RevokedAt is when the delegation was revoked (nil if not revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Status is the current status of the delegation
	Status DelegationStatus `json:"status"`

	// RevocationReason is the reason for revocation (if revoked)
	RevocationReason string `json:"revocation_reason,omitempty"`

	// Metadata contains optional metadata for the delegation
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewDelegationRecord creates a new delegation record
func NewDelegationRecord(
	delegationID string,
	delegatorAddress string,
	delegateAddress string,
	permissions []DelegationPermission,
	expiresAt time.Time,
	maxUses uint64,
	createdAt time.Time,
) *DelegationRecord {
	usesRemaining := maxUses
	if maxUses == 0 {
		// Unlimited uses represented by max uint64
		usesRemaining = ^uint64(0)
	}

	return &DelegationRecord{
		DelegationID:     delegationID,
		DelegatorAddress: delegatorAddress,
		DelegateAddress:  delegateAddress,
		Permissions:      permissions,
		ExpiresAt:        expiresAt,
		MaxUses:          maxUses,
		UsesRemaining:    usesRemaining,
		CreatedAt:        createdAt,
		Status:           DelegationActive,
		Metadata:         make(map[string]string),
	}
}

// Validate validates the delegation record
func (d *DelegationRecord) Validate() error {
	if d.DelegationID == "" {
		return ErrInvalidDelegation.Wrap("delegation_id cannot be empty")
	}

	if d.DelegatorAddress == "" {
		return ErrInvalidDelegation.Wrap("delegator_address cannot be empty")
	}

	if d.DelegateAddress == "" {
		return ErrInvalidDelegation.Wrap("delegate_address cannot be empty")
	}

	if d.DelegatorAddress == d.DelegateAddress {
		return ErrInvalidDelegation.Wrap("cannot delegate to self")
	}

	if len(d.Permissions) == 0 {
		return ErrInvalidDelegation.Wrap("at least one permission is required")
	}

	// Check for duplicate permissions
	seen := make(map[DelegationPermission]bool)
	for _, p := range d.Permissions {
		if !p.IsValid() {
			return ErrInvalidDelegation.Wrapf("invalid permission: %d", p)
		}
		if seen[p] {
			return ErrInvalidDelegation.Wrapf("duplicate permission: %s", p.String())
		}
		seen[p] = true
	}

	if d.CreatedAt.IsZero() {
		return ErrInvalidDelegation.Wrap("created_at cannot be zero")
	}

	if d.ExpiresAt.IsZero() {
		return ErrInvalidDelegation.Wrap("expires_at cannot be zero")
	}

	if d.ExpiresAt.Before(d.CreatedAt) {
		return ErrInvalidDelegation.Wrap("expires_at cannot be before created_at")
	}

	if !d.Status.IsValid() {
		return ErrInvalidDelegation.Wrapf("invalid status: %d", d.Status)
	}

	return nil
}

// IsActive returns true if the delegation is currently active
func (d *DelegationRecord) IsActive(now time.Time) bool {
	if d.Status != DelegationActive {
		return false
	}

	if now.After(d.ExpiresAt) {
		return false
	}

	if d.MaxUses > 0 && d.UsesRemaining == 0 {
		return false
	}

	return true
}

// HasPermission returns true if the delegation grants the specified permission
func (d *DelegationRecord) HasPermission(permission DelegationPermission) bool {
	for _, p := range d.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// UseOnce decrements the remaining uses and returns true if successful
// Returns false if no uses remaining or delegation is not active
func (d *DelegationRecord) UseOnce(now time.Time) bool {
	if !d.IsActive(now) {
		return false
	}

	// If MaxUses is 0 (unlimited), UsesRemaining is max uint64, so we don't decrement
	if d.MaxUses == 0 {
		return true
	}

	if d.UsesRemaining > 0 {
		d.UsesRemaining--
		if d.UsesRemaining == 0 {
			d.Status = DelegationExhausted
		}
		return true
	}

	return false
}

// Revoke marks the delegation as revoked
func (d *DelegationRecord) Revoke(revokedAt time.Time, reason string) error {
	if d.Status.IsTerminal() {
		return ErrDelegationAlreadyRevoked.Wrapf("delegation is already %s", d.Status.String())
	}

	d.Status = DelegationRevoked
	d.RevokedAt = &revokedAt
	d.RevocationReason = reason
	return nil
}

// UpdateStatus updates the delegation status based on current time
func (d *DelegationRecord) UpdateStatus(now time.Time) {
	if d.Status.IsTerminal() {
		return
	}

	if now.After(d.ExpiresAt) {
		d.Status = DelegationExpired
	}
}

// DelegationValidationResult contains the result of validating a delegation
type DelegationValidationResult struct {
	// Valid indicates if the delegation is valid
	Valid bool

	// Reason provides details about why validation failed (if not valid)
	Reason string

	// Delegation is the delegation record (if found)
	Delegation *DelegationRecord
}

// DelegationQuery represents query parameters for listing delegations
type DelegationQuery struct {
	// DelegatorAddress filters by delegator (optional)
	DelegatorAddress string

	// DelegateAddress filters by delegate (optional)
	DelegateAddress string

	// Status filters by status (optional, nil = all)
	Status *DelegationStatus

	// Permission filters by required permission (optional)
	Permission *DelegationPermission

	// ActiveOnly only returns active delegations
	ActiveOnly bool

	// IncludeExpired includes expired delegations
	IncludeExpired bool

	// Limit is the maximum number of results to return
	Limit uint32

	// Offset is the offset for pagination
	Offset uint32
}
