package privileged

import "fmt"

type RoleGrantState uint8

const (
	RoleGrantPending RoleGrantState = iota + 1
	RoleGrantActive
	RoleGrantSuspended
	RoleGrantRevoked
	RoleGrantExpired
)

type RoleGrant struct {
	GrantID         string         `json:"grant_id"`
	AccountID       string         `json:"account_id"`
	RoleID          string         `json:"role_id"`
	State           RoleGrantState `json:"state"`
	Revision        uint64         `json:"revision"`
	GrantedBy       string         `json:"granted_by"`
	CreatedAt       int64          `json:"created_at"`
	ActivatedAt     int64          `json:"activated_at,omitempty"`
	SuspendedAt     int64          `json:"suspended_at,omitempty"`
	RevokedAt       int64          `json:"revoked_at,omitempty"`
	ExpiresAt       int64          `json:"expires_at"`
	ApprovalDigest  [32]byte       `json:"approval_digest"`
	LastEventDigest [32]byte       `json:"last_event_digest"`
}

func (g RoleGrant) Validate() error {
	if invalidExactValue(g.GrantID) || invalidExactValue(g.AccountID) || invalidExactValue(g.RoleID) || invalidExactValue(g.GrantedBy) {
		return fmt.Errorf("role grant identity is incomplete")
	}
	if g.Revision == 0 || g.CreatedAt <= 0 || g.ExpiresAt <= g.CreatedAt || g.ApprovalDigest == ([32]byte{}) || g.LastEventDigest == ([32]byte{}) {
		return fmt.Errorf("role grant revision, validity, approval, and event lineage are required")
	}
	if g.State < RoleGrantPending || g.State > RoleGrantExpired {
		return fmt.Errorf("invalid role grant state")
	}
	return nil
}

func (g RoleGrant) Transition(next RoleGrantState, at int64, eventDigest [32]byte) (RoleGrant, error) {
	if err := g.Validate(); err != nil {
		return RoleGrant{}, err
	}
	if at <= g.CreatedAt || eventDigest == ([32]byte{}) {
		return RoleGrant{}, fmt.Errorf("transition time and event digest are required")
	}
	allowed := false
	switch g.State {
	case RoleGrantPending:
		allowed = next == RoleGrantActive || next == RoleGrantRevoked || next == RoleGrantExpired
	case RoleGrantActive:
		allowed = next == RoleGrantSuspended || next == RoleGrantRevoked || next == RoleGrantExpired
	case RoleGrantSuspended:
		allowed = next == RoleGrantActive || next == RoleGrantRevoked || next == RoleGrantExpired
	}
	if !allowed {
		return RoleGrant{}, fmt.Errorf("role grant transition %d -> %d is forbidden", g.State, next)
	}
	if next != RoleGrantExpired && at >= g.ExpiresAt {
		return RoleGrant{}, fmt.Errorf("expired grant cannot transition to a live state")
	}
	g.State = next
	g.Revision++
	g.LastEventDigest = eventDigest
	switch next {
	case RoleGrantActive:
		g.ActivatedAt = at
	case RoleGrantSuspended:
		g.SuspendedAt = at
	case RoleGrantRevoked:
		g.RevokedAt = at
	}
	return g, nil
}

type AccountLifecycleState uint8

const (
	AccountPending AccountLifecycleState = iota + 1
	AccountActive
	AccountSuspended
	AccountRecoveryPending
	AccountHeld
	AccountClosed
)

type AccountLifecycle struct {
	AccountID       string                `json:"account_id"`
	State           AccountLifecycleState `json:"state"`
	Revision        uint64                `json:"revision"`
	UpdatedAt       int64                 `json:"updated_at"`
	AuthorityID     string                `json:"authority_id"`
	Reason          string                `json:"reason"`
	LastEventDigest [32]byte              `json:"last_event_digest"`
}

func (a AccountLifecycle) Validate() error {
	if invalidExactValue(a.AccountID) || invalidExactValue(a.AuthorityID) || a.Reason == "" || a.Revision == 0 || a.UpdatedAt <= 0 || a.LastEventDigest == ([32]byte{}) {
		return fmt.Errorf("account lifecycle identity, authority, revision, reason, and lineage are required")
	}
	if a.State < AccountPending || a.State > AccountClosed {
		return fmt.Errorf("invalid account lifecycle state")
	}
	return nil
}

func (a AccountLifecycle) Transition(next AccountLifecycleState, at int64, authorityID, reason string, eventDigest [32]byte) (AccountLifecycle, error) {
	if err := a.Validate(); err != nil {
		return AccountLifecycle{}, err
	}
	if at <= a.UpdatedAt || invalidExactValue(authorityID) || reason == "" || eventDigest == ([32]byte{}) {
		return AccountLifecycle{}, fmt.Errorf("account transition metadata is incomplete")
	}
	allowed := false
	switch a.State {
	case AccountPending:
		allowed = next == AccountActive || next == AccountClosed
	case AccountActive:
		allowed = next == AccountSuspended || next == AccountRecoveryPending || next == AccountHeld || next == AccountClosed
	case AccountSuspended:
		allowed = next == AccountActive || next == AccountRecoveryPending || next == AccountHeld || next == AccountClosed
	case AccountRecoveryPending:
		allowed = next == AccountActive || next == AccountSuspended || next == AccountHeld || next == AccountClosed
	case AccountHeld:
		allowed = next == AccountActive || next == AccountSuspended || next == AccountClosed
	}
	if !allowed {
		return AccountLifecycle{}, fmt.Errorf("account transition %d -> %d is forbidden", a.State, next)
	}
	a.State = next
	a.Revision++
	a.UpdatedAt = at
	a.AuthorityID = authorityID
	a.Reason = reason
	a.LastEventDigest = eventDigest
	return a, nil
}
