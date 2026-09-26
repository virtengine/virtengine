package types

// Event types for the roles module
const (
	EventTypeRoleAssigned        = "role_assigned"
	EventTypeRoleRevoked         = "role_revoked"
	EventTypeAccountStateChanged = "account_state_changed"
	EventTypeAdminNominated      = "admin_nominated"
	EventTypeSanctionImposed     = "sanction_imposed"
	EventTypeSanctionConfirmed   = "sanction_confirmed"
	EventTypeSanctionEscalated   = "sanction_escalated"
	EventTypeSanctionExpired     = "sanction_expired"
	EventTypeSanctionRevoked     = "sanction_revoked"
	EventTypeAppealOpened        = "appeal_opened"
	EventTypeAppealResolved      = "appeal_resolved"
)

// Event attribute keys
const (
	AttributeKeyAddress        = "address"
	AttributeKeyRole           = "role"
	AttributeKeyAssignedBy     = "assigned_by"
	AttributeKeyRevokedBy      = "revoked_by"
	AttributeKeyPreviousState  = "previous_state"
	AttributeKeyNewState       = "new_state"
	AttributeKeyModifiedBy     = "modified_by"
	AttributeKeyReason         = "reason"
	AttributeKeyNominatedBy    = "nominated_by"
	AttributeKeySanctionID     = "sanction_id"
	AttributeKeySanctionKind   = "sanction_kind"
	AttributeKeySanctionScope  = "sanction_scope"
	AttributeKeySanctionStatus = "sanction_status"
	AttributeKeyImposedBy      = "imposed_by"
	AttributeKeySecondReviewer = "second_reviewer"
	AttributeKeyAppealOf       = "appeal_of"
	AttributeKeyAppealID       = "appeal_id"
	AttributeKeyExpiresAt      = "expires_at"
)

// EventRoleAssigned is emitted when a role is assigned to an account
type EventRoleAssigned struct {
	Address    string `json:"address"`
	Role       string `json:"role"`
	AssignedBy string `json:"assigned_by"`
}

// EventRoleRevoked is emitted when a role is revoked from an account
type EventRoleRevoked struct {
	Address   string `json:"address"`
	Role      string `json:"role"`
	RevokedBy string `json:"revoked_by"`
}

// EventAccountStateChanged is emitted when an account state changes
type EventAccountStateChanged struct {
	Address       string `json:"address"`
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
	ModifiedBy    string `json:"modified_by"`
	Reason        string `json:"reason"`
}

// EventAdminNominated is emitted when a new administrator is nominated
type EventAdminNominated struct {
	Address     string `json:"address"`
	NominatedBy string `json:"nominated_by"`
}
