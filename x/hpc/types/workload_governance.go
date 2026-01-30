// Package types contains types for the HPC module.
//
// VE-5F: Workload governance types for admin approval flow
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WorkloadProposalType represents the type of workload governance proposal
type WorkloadProposalType string

const (
	// WorkloadProposalTypeAdd proposes adding a new template
	WorkloadProposalTypeAdd WorkloadProposalType = "add"

	// WorkloadProposalTypeApprove proposes approving a pending template
	WorkloadProposalTypeApprove WorkloadProposalType = "approve"

	// WorkloadProposalTypeReject proposes rejecting a pending template
	WorkloadProposalTypeReject WorkloadProposalType = "reject"

	// WorkloadProposalTypeDeprecate proposes deprecating a template
	WorkloadProposalTypeDeprecate WorkloadProposalType = "deprecate"

	// WorkloadProposalTypeRevoke proposes revoking a template due to security issues
	WorkloadProposalTypeRevoke WorkloadProposalType = "revoke"

	// WorkloadProposalTypeUpdate proposes updating template metadata
	WorkloadProposalTypeUpdate WorkloadProposalType = "update"
)

// IsValid checks if the proposal type is valid
func (t WorkloadProposalType) IsValid() bool {
	switch t {
	case WorkloadProposalTypeAdd, WorkloadProposalTypeApprove, WorkloadProposalTypeReject,
		WorkloadProposalTypeDeprecate, WorkloadProposalTypeRevoke, WorkloadProposalTypeUpdate:
		return true
	default:
		return false
	}
}

// WorkloadProposalStatus represents the status of a workload proposal
type WorkloadProposalStatus string

const (
	// WorkloadProposalStatusPending indicates proposal is pending voting
	WorkloadProposalStatusPending WorkloadProposalStatus = "pending"

	// WorkloadProposalStatusPassed indicates proposal passed
	WorkloadProposalStatusPassed WorkloadProposalStatus = "passed"

	// WorkloadProposalStatusRejected indicates proposal was rejected
	WorkloadProposalStatusRejected WorkloadProposalStatus = "rejected"

	// WorkloadProposalStatusExpired indicates proposal expired
	WorkloadProposalStatusExpired WorkloadProposalStatus = "expired"

	// WorkloadProposalStatusExecuted indicates proposal was executed
	WorkloadProposalStatusExecuted WorkloadProposalStatus = "executed"

	// WorkloadProposalStatusFailed indicates proposal execution failed
	WorkloadProposalStatusFailed WorkloadProposalStatus = "failed"
)

// IsValid checks if the proposal status is valid
func (s WorkloadProposalStatus) IsValid() bool {
	switch s {
	case WorkloadProposalStatusPending, WorkloadProposalStatusPassed, WorkloadProposalStatusRejected,
		WorkloadProposalStatusExpired, WorkloadProposalStatusExecuted, WorkloadProposalStatusFailed:
		return true
	default:
		return false
	}
}

// IsFinal checks if the status is final (no more changes)
func (s WorkloadProposalStatus) IsFinal() bool {
	switch s {
	case WorkloadProposalStatusRejected, WorkloadProposalStatusExpired, WorkloadProposalStatusExecuted, WorkloadProposalStatusFailed:
		return true
	default:
		return false
	}
}

// WorkloadGovernanceProposal represents a governance proposal for workload templates
type WorkloadGovernanceProposal struct {
	// ProposalID is the unique identifier for the proposal
	ProposalID string `json:"proposal_id"`

	// Type is the type of proposal
	Type WorkloadProposalType `json:"type"`

	// TemplateID is the target template ID
	TemplateID string `json:"template_id"`

	// TemplateVersion is the target template version (if applicable)
	TemplateVersion string `json:"template_version,omitempty"`

	// Template is the full template for add/update proposals
	Template *WorkloadTemplate `json:"template,omitempty"`

	// Title is the proposal title
	Title string `json:"title"`

	// Description is the proposal description
	Description string `json:"description"`

	// Proposer is the address that created the proposal
	Proposer string `json:"proposer"`

	// Status is the current proposal status
	Status WorkloadProposalStatus `json:"status"`

	// VotingStartTime is when voting starts
	VotingStartTime time.Time `json:"voting_start_time"`

	// VotingEndTime is when voting ends
	VotingEndTime time.Time `json:"voting_end_time"`

	// VoteYes is the count of yes votes
	VoteYes int64 `json:"vote_yes"`

	// VoteNo is the count of no votes
	VoteNo int64 `json:"vote_no"`

	// VoteAbstain is the count of abstain votes
	VoteAbstain int64 `json:"vote_abstain"`

	// Deposit is the proposal deposit
	Deposit sdk.Coins `json:"deposit"`

	// SecurityReason is the reason for revocation (for revoke proposals)
	SecurityReason string `json:"security_reason,omitempty"`

	// CreatedAt is when the proposal was created
	CreatedAt time.Time `json:"created_at"`

	// ExecutedAt is when the proposal was executed
	ExecutedAt *time.Time `json:"executed_at,omitempty"`

	// BlockHeight is the block when proposal was created
	BlockHeight int64 `json:"block_height"`
}

// Validate validates a governance proposal
func (p *WorkloadGovernanceProposal) Validate() error {
	if p.ProposalID == "" {
		return ErrWorkloadGovernanceFailed.Wrap("proposal_id required")
	}

	if !p.Type.IsValid() {
		return ErrWorkloadGovernanceFailed.Wrapf("invalid proposal type: %s", p.Type)
	}

	if p.TemplateID == "" {
		return ErrWorkloadGovernanceFailed.Wrap("template_id required")
	}

	if p.Title == "" || len(p.Title) > 256 {
		return ErrWorkloadGovernanceFailed.Wrap("title required and must be <= 256 chars")
	}

	if len(p.Description) > 4096 {
		return ErrWorkloadGovernanceFailed.Wrap("description must be <= 4096 chars")
	}

	if _, err := sdk.AccAddressFromBech32(p.Proposer); err != nil {
		return ErrWorkloadGovernanceFailed.Wrap("invalid proposer address")
	}

	if !p.Status.IsValid() {
		return ErrWorkloadGovernanceFailed.Wrapf("invalid status: %s", p.Status)
	}

	// For add proposals, template is required
	if p.Type == WorkloadProposalTypeAdd && p.Template == nil {
		return ErrWorkloadGovernanceFailed.Wrap("template required for add proposals")
	}

	// For revoke proposals, security reason is required
	if p.Type == WorkloadProposalTypeRevoke && p.SecurityReason == "" {
		return ErrWorkloadGovernanceFailed.Wrap("security_reason required for revoke proposals")
	}

	return nil
}

// WorkloadVote represents a vote on a workload proposal
type WorkloadVote struct {
	// ProposalID is the proposal being voted on
	ProposalID string `json:"proposal_id"`

	// Voter is the address that cast the vote
	Voter string `json:"voter"`

	// Option is the vote option (yes|no|abstain)
	Option string `json:"option"`

	// Weight is the voting weight
	Weight int64 `json:"weight"`

	// Timestamp is when the vote was cast
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block when vote was cast
	BlockHeight int64 `json:"block_height"`
}

// Validate validates a vote
func (v *WorkloadVote) Validate() error {
	if v.ProposalID == "" {
		return ErrWorkloadGovernanceFailed.Wrap("proposal_id required")
	}

	if _, err := sdk.AccAddressFromBech32(v.Voter); err != nil {
		return ErrWorkloadGovernanceFailed.Wrap("invalid voter address")
	}

	validOptions := map[string]bool{"yes": true, "no": true, "abstain": true}
	if !validOptions[v.Option] {
		return ErrWorkloadGovernanceFailed.Wrapf("invalid vote option: %s", v.Option)
	}

	if v.Weight < 0 {
		return ErrWorkloadGovernanceFailed.Wrap("weight cannot be negative")
	}

	return nil
}

// WorkloadGovernanceParams contains governance parameters for workload templates
type WorkloadGovernanceParams struct {
	// VotingPeriod is the duration of the voting period
	VotingPeriod time.Duration `json:"voting_period"`

	// MinDeposit is the minimum deposit required
	MinDeposit sdk.Coins `json:"min_deposit"`

	// Quorum is the minimum participation required (fixed-point, 6 decimals)
	Quorum string `json:"quorum"`

	// Threshold is the minimum yes votes required (fixed-point, 6 decimals)
	Threshold string `json:"threshold"`

	// VetoThreshold is the veto threshold (fixed-point, 6 decimals)
	VetoThreshold string `json:"veto_threshold"`

	// AllowedProposers lists addresses allowed to create proposals (empty = anyone)
	AllowedProposers []string `json:"allowed_proposers,omitempty"`

	// AutoApproveBuiltinUpdates auto-approves updates from trusted publishers
	AutoApproveBuiltinUpdates bool `json:"auto_approve_builtin_updates"`

	// SecurityRevokeQuorum is lower quorum for security revocations
	SecurityRevokeQuorum string `json:"security_revoke_quorum"`
}

// DefaultWorkloadGovernanceParams returns default governance parameters
func DefaultWorkloadGovernanceParams() WorkloadGovernanceParams {
	return WorkloadGovernanceParams{
		VotingPeriod:              7 * 24 * time.Hour, // 7 days
		MinDeposit:                sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000000000)), // 1000 VIRT
		Quorum:                    "0.334000", // 33.4%
		Threshold:                 "0.500000", // 50%
		VetoThreshold:             "0.334000", // 33.4%
		AllowedProposers:          nil, // Anyone can propose
		AutoApproveBuiltinUpdates: true,
		SecurityRevokeQuorum:      "0.200000", // 20% for security revocations
	}
}

// Validate validates governance parameters
func (p *WorkloadGovernanceParams) Validate() error {
	if p.VotingPeriod < time.Hour {
		return ErrWorkloadGovernanceFailed.Wrap("voting_period must be at least 1 hour")
	}

	if !p.MinDeposit.IsValid() {
		return ErrWorkloadGovernanceFailed.Wrap("min_deposit must be valid coins")
	}

	return nil
}

// WorkloadApprovalEvent represents a workload approval event
type WorkloadApprovalEvent struct {
	// EventType is the type of event
	EventType string `json:"event_type"`

	// TemplateID is the template affected
	TemplateID string `json:"template_id"`

	// TemplateVersion is the template version
	TemplateVersion string `json:"template_version"`

	// PreviousStatus is the previous approval status
	PreviousStatus WorkloadApprovalStatus `json:"previous_status"`

	// NewStatus is the new approval status
	NewStatus WorkloadApprovalStatus `json:"new_status"`

	// ProposalID is the related proposal (if any)
	ProposalID string `json:"proposal_id,omitempty"`

	// Actor is the address that triggered the event
	Actor string `json:"actor"`

	// Reason is the reason for the change
	Reason string `json:"reason,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block when event occurred
	BlockHeight int64 `json:"block_height"`
}

// NewWorkloadApprovalEvent creates a new approval event
func NewWorkloadApprovalEvent(
	eventType string,
	templateID, version string,
	prevStatus, newStatus WorkloadApprovalStatus,
	actor string,
	blockHeight int64,
) WorkloadApprovalEvent {
	return WorkloadApprovalEvent{
		EventType:       eventType,
		TemplateID:      templateID,
		TemplateVersion: version,
		PreviousStatus:  prevStatus,
		NewStatus:       newStatus,
		Actor:           actor,
		Timestamp:       time.Now().UTC(),
		BlockHeight:     blockHeight,
	}
}
