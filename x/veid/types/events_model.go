// Package types provides VEID module types.
//
// This file defines model versioning events for the VEID module.
//
// Task Reference: VE-3007 - Model Versioning and Governance
package types

// Model versioning event types
const (
	// EventTypeModelRegistered is emitted when a new model is registered
	EventTypeModelRegistered = "model_registered"

	// EventTypeModelUpdateProposed is emitted when a model update is proposed
	EventTypeModelUpdateProposed = "model_update_proposed"

	// EventTypeModelActivated is emitted when a model is activated
	EventTypeModelActivated = "model_activated"

	// EventTypeModelVersionMismatch is emitted when a validator has version mismatch
	EventTypeModelVersionMismatch = "model_version_mismatch"

	// EventTypeModelDeprecated is emitted when a model is deprecated
	EventTypeModelDeprecated = "model_deprecated"

	// EventTypeModelRevoked is emitted when a model is revoked
	EventTypeModelRevoked = "model_revoked"

	// EventTypeValidatorModelReport is emitted when a validator reports model versions
	EventTypeValidatorModelReport = "validator_model_report"

	// EventTypeModelProposalApproved is emitted when a model proposal is approved
	EventTypeModelProposalApproved = "model_proposal_approved"

	// EventTypeModelProposalRejected is emitted when a model proposal is rejected
	EventTypeModelProposalRejected = "model_proposal_rejected"
)

// Model versioning event attribute keys
const (
	// AttributeKeyModelID is the model ID attribute
	AttributeKeyModelID = "model_id"

	// AttributeKeyModelName is the model name attribute
	AttributeKeyModelName = "model_name"

	// AttributeKeyModelType is the model type attribute
	AttributeKeyModelType = "model_type"

	// AttributeKeyModelHash is the model SHA256 hash attribute
	AttributeKeyModelHash = "model_hash"

	// AttributeKeyOldModelID is the previous model ID attribute
	AttributeKeyOldModelID = "old_model_id"

	// AttributeKeyNewModelID is the new model ID attribute
	AttributeKeyNewModelID = "new_model_id"

	// AttributeKeyRegistrar is the registrar address attribute
	AttributeKeyRegistrar = "registrar"

	// AttributeKeyGovernanceID is the governance proposal ID attribute
	AttributeKeyGovernanceID = "governance_id"

	// AttributeKeyActivationHeight is the activation height attribute
	AttributeKeyActivationHeight = "activation_height"

	// AttributeKeyProposalTitle is the proposal title attribute
	AttributeKeyProposalTitle = "proposal_title"

	// AttributeKeyExpectedHash is the expected hash attribute
	AttributeKeyExpectedHash = "expected_hash"

	// AttributeKeyReportedHash is the reported hash attribute
	AttributeKeyReportedHash = "reported_hash"

	// AttributeKeyIsSynced is the sync status attribute
	AttributeKeyIsSynced = "is_synced"
)

// ============================================================================
// Event Types for Model Versioning
// ============================================================================

// EventModelRegistered is emitted when a new model is registered
type EventModelRegistered struct {
	ModelID      string `json:"model_id"`
	ModelType    string `json:"model_type"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	SHA256Hash   string `json:"sha256_hash"`
	RegisteredBy string `json:"registered_by"`
	RegisteredAt int64  `json:"registered_at"`
}

// EventModelUpdateProposed is emitted when a model update is proposed
type EventModelUpdateProposed struct {
	ProposalID       string `json:"proposal_id,omitempty"`
	ModelType        string `json:"model_type"`
	NewModelID       string `json:"new_model_id"`
	NewModelHash     string `json:"new_model_hash"`
	ProposerAddress  string `json:"proposer_address"`
	ActivationDelay  int64  `json:"activation_delay"`
	ActivationHeight int64  `json:"activation_height"`
	ProposedAt       int64  `json:"proposed_at"`
}

// EventModelActivated is emitted when a model is activated
type EventModelActivated struct {
	ModelType    string `json:"model_type"`
	OldModelID   string `json:"old_model_id,omitempty"`
	NewModelID   string `json:"new_model_id"`
	NewModelHash string `json:"new_model_hash"`
	GovernanceID uint64 `json:"governance_id,omitempty"`
	ActivatedAt  int64  `json:"activated_at"`
}

// EventModelVersionMismatch is emitted when a validator has a version mismatch
type EventModelVersionMismatch struct {
	ValidatorAddress string `json:"validator_address"`
	ModelType        string `json:"model_type"`
	ExpectedHash     string `json:"expected_hash"`
	ReportedHash     string `json:"reported_hash"`
	BlockHeight      int64  `json:"block_height"`
}

// EventModelDeprecated is emitted when a model is deprecated
type EventModelDeprecated struct {
	ModelID      string `json:"model_id"`
	ModelType    string `json:"model_type"`
	DeprecatedBy string `json:"deprecated_by"`
	DeprecatedAt int64  `json:"deprecated_at"`
	Reason       string `json:"reason,omitempty"`
}

// EventModelRevoked is emitted when a model is revoked
type EventModelRevoked struct {
	ModelID   string `json:"model_id"`
	ModelType string `json:"model_type"`
	RevokedBy string `json:"revoked_by"`
	RevokedAt int64  `json:"revoked_at"`
	Reason    string `json:"reason,omitempty"`
}

// EventValidatorModelReport is emitted when a validator reports model versions
type EventValidatorModelReport struct {
	ValidatorAddress string   `json:"validator_address"`
	IsSynced         bool     `json:"is_synced"`
	MismatchedModels []string `json:"mismatched_models,omitempty"`
	ReportedAt       int64    `json:"reported_at"`
}

// EventModelProposalApproved is emitted when a model proposal is approved
type EventModelProposalApproved struct {
	GovernanceID     uint64 `json:"governance_id"`
	ModelType        string `json:"model_type"`
	NewModelID       string `json:"new_model_id"`
	ActivationHeight int64  `json:"activation_height"`
	ApprovedAt       int64  `json:"approved_at"`
}

// EventModelProposalRejected is emitted when a model proposal is rejected
type EventModelProposalRejected struct {
	GovernanceID uint64 `json:"governance_id"`
	ModelType    string `json:"model_type"`
	NewModelID   string `json:"new_model_id"`
	Reason       string `json:"reason,omitempty"`
	RejectedAt   int64  `json:"rejected_at"`
}
