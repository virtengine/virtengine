// Package keeper implements workload template management.
//
// VE-5F: Workload template keeper methods
package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/mod/semver"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Workload Template Management
// ============================================================================

// CreateWorkloadTemplate creates a new workload template
func (k Keeper) CreateWorkloadTemplate(ctx sdk.Context, template *types.WorkloadTemplate) error {
	if err := template.Validate(); err != nil {
		return err
	}

	// Check for duplicate
	if _, exists := k.GetWorkloadTemplateByVersion(ctx, template.TemplateID, template.Version); exists {
		return types.ErrInvalidWorkloadTemplate.Wrap("template version already exists")
	}

	template.ApprovalStatus = types.WorkloadApprovalPending
	template.CreatedAt = ctx.BlockTime()
	template.UpdatedAt = ctx.BlockTime()
	template.BlockHeight = ctx.BlockHeight()

	if err := k.SetWorkloadTemplate(ctx, *template); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_created",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("type", string(template.Type)),
			sdk.NewAttribute("publisher", template.Publisher),
		),
	)

	return nil
}

// UpdateWorkloadTemplate updates an existing workload template
func (k Keeper) UpdateWorkloadTemplate(ctx sdk.Context, template *types.WorkloadTemplate) error {
	existing, exists := k.GetWorkloadTemplateByVersion(ctx, template.TemplateID, template.Version)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	// Only publisher can update
	if existing.Publisher != template.Publisher {
		return types.ErrUnauthorized.Wrap("only publisher can update template")
	}

	// Cannot update approved templates directly - need new version
	if existing.ApprovalStatus == types.WorkloadApprovalApproved {
		return types.ErrInvalidWorkloadTemplate.Wrap("create new version for approved templates")
	}

	template.CreatedAt = existing.CreatedAt
	template.UpdatedAt = ctx.BlockTime()
	template.ApprovalStatus = existing.ApprovalStatus

	if err := template.Validate(); err != nil {
		return err
	}

	return k.SetWorkloadTemplate(ctx, *template)
}

// GetWorkloadTemplate retrieves a workload template by ID
func (k Keeper) GetWorkloadTemplate(ctx sdk.Context, templateID string) (types.WorkloadTemplate, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetWorkloadTemplateKey(templateID))
	if bz == nil {
		return types.WorkloadTemplate{}, false
	}

	var template types.WorkloadTemplate
	if err := json.Unmarshal(bz, &template); err != nil {
		return types.WorkloadTemplate{}, false
	}
	return template, true
}

// GetWorkloadTemplateByVersion retrieves a specific version of a template
func (k Keeper) GetWorkloadTemplateByVersion(ctx sdk.Context, templateID, version string) (types.WorkloadTemplate, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetWorkloadTemplateVersionKey(templateID, version))
	if bz == nil {
		return types.WorkloadTemplate{}, false
	}

	var template types.WorkloadTemplate
	if err := json.Unmarshal(bz, &template); err != nil {
		return types.WorkloadTemplate{}, false
	}
	return template, true
}

// SetWorkloadTemplate stores a workload template
func (k Keeper) SetWorkloadTemplate(ctx sdk.Context, template types.WorkloadTemplate) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(template)
	if err != nil {
		return err
	}

	// Store by ID and version
	store.Set(types.GetWorkloadTemplateVersionKey(template.TemplateID, template.Version), bz)

	// Store by ID (latest version) if newer or same version
	if shouldUpdateLatestTemplate(ctx, k, template) {
		store.Set(types.GetWorkloadTemplateKey(template.TemplateID), bz)
		// Store by type index for latest version
		store.Set(types.GetWorkloadTemplateByTypeKey(template.Type, template.TemplateID), []byte(template.TemplateID))
	}

	return nil
}

// DeleteWorkloadTemplate removes a workload template
func (k Keeper) DeleteWorkloadTemplate(ctx sdk.Context, templateID string) error {
	template, exists := k.GetWorkloadTemplate(ctx, templateID)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	store := ctx.KVStore(k.skey)
	store.Delete(types.GetWorkloadTemplateKey(templateID))
	store.Delete(types.GetWorkloadTemplateVersionKey(templateID, template.Version))
	store.Delete(types.GetWorkloadTemplateByTypeKey(template.Type, templateID))

	return nil
}

// ApproveWorkloadTemplate approves a workload template
func (k Keeper) ApproveWorkloadTemplate(ctx sdk.Context, templateID, version string, approver sdk.AccAddress) error {
	template, exists := k.GetWorkloadTemplateByVersion(ctx, templateID, version)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	// Check authorization (must be governance module or authority)
	if k.authority != approver.String() {
		return types.ErrUnauthorized.Wrap("only authority can approve templates")
	}

	prevStatus := template.ApprovalStatus
	template.ApprovalStatus = types.WorkloadApprovalApproved
	now := ctx.BlockTime()
	template.ApprovedAt = &now
	template.UpdatedAt = ctx.BlockTime()

	if err := k.SetWorkloadTemplate(ctx, template); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_approved",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("previous_status", string(prevStatus)),
			sdk.NewAttribute("approver", approver.String()),
		),
	)

	return nil
}

// RejectWorkloadTemplate rejects a workload template
func (k Keeper) RejectWorkloadTemplate(ctx sdk.Context, templateID, version string, reason string, rejector sdk.AccAddress) error {
	template, exists := k.GetWorkloadTemplateByVersion(ctx, templateID, version)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	if k.authority != rejector.String() {
		return types.ErrUnauthorized.Wrap("only authority can reject templates")
	}

	prevStatus := template.ApprovalStatus
	template.ApprovalStatus = types.WorkloadApprovalRejected
	template.UpdatedAt = ctx.BlockTime()

	if err := k.SetWorkloadTemplate(ctx, template); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_rejected",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("previous_status", string(prevStatus)),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("rejector", rejector.String()),
		),
	)

	return nil
}

// DeprecateWorkloadTemplate deprecates a workload template
func (k Keeper) DeprecateWorkloadTemplate(ctx sdk.Context, templateID, version string, deprecator sdk.AccAddress) error {
	template, exists := k.GetWorkloadTemplateByVersion(ctx, templateID, version)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	// Publisher or authority can deprecate
	if template.Publisher != deprecator.String() && k.authority != deprecator.String() {
		return types.ErrUnauthorized.Wrap("only publisher or authority can deprecate templates")
	}

	prevStatus := template.ApprovalStatus
	template.ApprovalStatus = types.WorkloadApprovalDeprecated
	template.UpdatedAt = ctx.BlockTime()

	if err := k.SetWorkloadTemplate(ctx, template); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_deprecated",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("previous_status", string(prevStatus)),
			sdk.NewAttribute("deprecator", deprecator.String()),
		),
	)

	return nil
}

// RevokeWorkloadTemplate revokes a workload template for security reasons
func (k Keeper) RevokeWorkloadTemplate(ctx sdk.Context, templateID, version string, securityReason string, revoker sdk.AccAddress) error {
	template, exists := k.GetWorkloadTemplateByVersion(ctx, templateID, version)
	if !exists {
		return types.ErrWorkloadTemplateNotFound
	}

	// Only authority can revoke
	if k.authority != revoker.String() {
		return types.ErrUnauthorized.Wrap("only authority can revoke templates")
	}

	prevStatus := template.ApprovalStatus
	template.ApprovalStatus = types.WorkloadApprovalRevoked
	template.UpdatedAt = ctx.BlockTime()

	if err := k.SetWorkloadTemplate(ctx, template); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_template_revoked",
			sdk.NewAttribute("template_id", template.TemplateID),
			sdk.NewAttribute("version", template.Version),
			sdk.NewAttribute("previous_status", string(prevStatus)),
			sdk.NewAttribute("security_reason", securityReason),
			sdk.NewAttribute("revoker", revoker.String()),
		),
	)

	return nil
}

// GetWorkloadTemplatesByType retrieves templates by type
func (k Keeper) GetWorkloadTemplatesByType(ctx sdk.Context, workloadType types.WorkloadType) []types.WorkloadTemplate {
	var templates []types.WorkloadTemplate
	k.WithWorkloadTemplates(ctx, func(template types.WorkloadTemplate) bool {
		if template.Type == workloadType {
			templates = append(templates, template)
		}
		return false
	})
	return templates
}

// GetWorkloadTemplatesByPublisher retrieves templates by publisher
func (k Keeper) GetWorkloadTemplatesByPublisher(ctx sdk.Context, publisher string) []types.WorkloadTemplate {
	var templates []types.WorkloadTemplate
	k.WithWorkloadTemplates(ctx, func(template types.WorkloadTemplate) bool {
		if template.Publisher == publisher {
			templates = append(templates, template)
		}
		return false
	})
	return templates
}

// GetApprovedWorkloadTemplates retrieves all approved templates
func (k Keeper) GetApprovedWorkloadTemplates(ctx sdk.Context) []types.WorkloadTemplate {
	var templates []types.WorkloadTemplate
	k.WithWorkloadTemplates(ctx, func(template types.WorkloadTemplate) bool {
		if template.ApprovalStatus.CanBeUsed() {
			templates = append(templates, template)
		}
		return false
	})
	return templates
}

// WithWorkloadTemplates iterates over all workload templates
func (k Keeper) WithWorkloadTemplates(ctx sdk.Context, fn func(types.WorkloadTemplate) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.WorkloadTemplatePrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var template types.WorkloadTemplate
		if err := json.Unmarshal(iter.Value(), &template); err != nil {
			continue
		}
		if fn(template) {
			break
		}
	}
}

// ============================================================================
// Workload Governance Management
// ============================================================================

// CreateWorkloadProposal creates a new governance proposal
func (k Keeper) CreateWorkloadProposal(ctx sdk.Context, proposal *types.WorkloadGovernanceProposal) error {
	if err := proposal.Validate(); err != nil {
		return err
	}

	// Generate proposal ID if not set
	if proposal.ProposalID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyWorkloadProposal)
		proposal.ProposalID = fmt.Sprintf("wl-proposal-%d", seq)
	}

	params := k.GetWorkloadGovernanceParams(ctx)

	proposal.Status = types.WorkloadProposalStatusPending
	proposal.VotingStartTime = ctx.BlockTime()
	proposal.VotingEndTime = ctx.BlockTime().Add(params.VotingPeriod)
	proposal.CreatedAt = ctx.BlockTime()
	proposal.BlockHeight = ctx.BlockHeight()

	if err := k.SetWorkloadProposal(ctx, *proposal); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"workload_proposal_created",
			sdk.NewAttribute("proposal_id", proposal.ProposalID),
			sdk.NewAttribute("type", string(proposal.Type)),
			sdk.NewAttribute("template_id", proposal.TemplateID),
			sdk.NewAttribute("proposer", proposal.Proposer),
		),
	)

	return nil
}

// VoteOnWorkloadProposal records a vote on a proposal
func (k Keeper) VoteOnWorkloadProposal(ctx sdk.Context, vote *types.WorkloadVote) error {
	if err := vote.Validate(); err != nil {
		return err
	}

	proposal, exists := k.GetWorkloadProposal(ctx, vote.ProposalID)
	if !exists {
		return types.ErrWorkloadGovernanceFailed.Wrap("proposal not found")
	}

	// Check if voting is open
	if proposal.Status != types.WorkloadProposalStatusPending {
		return types.ErrWorkloadGovernanceFailed.Wrap("voting is not open")
	}

	if ctx.BlockTime().After(proposal.VotingEndTime) {
		return types.ErrWorkloadGovernanceFailed.Wrap("voting period has ended")
	}

	vote.Timestamp = ctx.BlockTime()
	vote.BlockHeight = ctx.BlockHeight()

	// Update vote counts
	switch vote.Option {
	case "yes":
		proposal.VoteYes += vote.Weight
	case "no":
		proposal.VoteNo += vote.Weight
	case "abstain":
		proposal.VoteAbstain += vote.Weight
	}

	if err := k.SetWorkloadProposal(ctx, proposal); err != nil {
		return err
	}

	if err := k.SetWorkloadVote(ctx, *vote); err != nil {
		return err
	}

	return nil
}

// TallyWorkloadProposal tallies votes and updates proposal status
func (k Keeper) TallyWorkloadProposal(ctx sdk.Context, proposalID string) error {
	proposal, exists := k.GetWorkloadProposal(ctx, proposalID)
	if !exists {
		return types.ErrWorkloadGovernanceFailed.Wrap("proposal not found")
	}

	if proposal.Status.IsFinal() {
		return nil // Already finalized
	}

	// Check if voting period has ended
	if ctx.BlockTime().Before(proposal.VotingEndTime) {
		return types.ErrWorkloadGovernanceFailed.Wrap("voting period not ended")
	}

	totalVotes := proposal.VoteYes + proposal.VoteNo + proposal.VoteAbstain

	// Simple majority for now
	if proposal.VoteYes > proposal.VoteNo {
		proposal.Status = types.WorkloadProposalStatusPassed
	} else if totalVotes == 0 {
		proposal.Status = types.WorkloadProposalStatusExpired
	} else {
		proposal.Status = types.WorkloadProposalStatusRejected
	}

	if err := k.SetWorkloadProposal(ctx, proposal); err != nil {
		return err
	}

	// Execute if passed
	if proposal.Status == types.WorkloadProposalStatusPassed {
		if err := k.ExecuteWorkloadProposal(ctx, &proposal); err != nil {
			proposal.Status = types.WorkloadProposalStatusFailed
			_ = k.SetWorkloadProposal(ctx, proposal)
			return err
		}
	}

	return nil
}

// ExecuteWorkloadProposal executes a passed proposal
func (k Keeper) ExecuteWorkloadProposal(ctx sdk.Context, proposal *types.WorkloadGovernanceProposal) error {
	authority, err := sdk.AccAddressFromBech32(k.authority)
	if err != nil {
		return err
	}

	switch proposal.Type {
	case types.WorkloadProposalTypeAdd:
		if proposal.Template == nil {
			return types.ErrWorkloadGovernanceFailed.Wrap("template required for add proposals")
		}
		if err := k.CreateWorkloadTemplate(ctx, proposal.Template); err != nil {
			return err
		}
		// Auto-approve since it passed governance
		return k.ApproveWorkloadTemplate(ctx, proposal.Template.TemplateID, proposal.Template.Version, authority)

	case types.WorkloadProposalTypeApprove:
		return k.ApproveWorkloadTemplate(ctx, proposal.TemplateID, proposal.TemplateVersion, authority)

	case types.WorkloadProposalTypeReject:
		return k.RejectWorkloadTemplate(ctx, proposal.TemplateID, proposal.TemplateVersion, "Rejected by governance", authority)

	case types.WorkloadProposalTypeDeprecate:
		return k.DeprecateWorkloadTemplate(ctx, proposal.TemplateID, proposal.TemplateVersion, authority)

	case types.WorkloadProposalTypeRevoke:
		return k.RevokeWorkloadTemplate(ctx, proposal.TemplateID, proposal.TemplateVersion, proposal.SecurityReason, authority)

	default:
		return types.ErrWorkloadGovernanceFailed.Wrapf("unknown proposal type: %s", proposal.Type)
	}
}

// GetWorkloadProposal retrieves a proposal by ID
func (k Keeper) GetWorkloadProposal(ctx sdk.Context, proposalID string) (types.WorkloadGovernanceProposal, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetWorkloadProposalKey(proposalID))
	if bz == nil {
		return types.WorkloadGovernanceProposal{}, false
	}

	var proposal types.WorkloadGovernanceProposal
	if err := json.Unmarshal(bz, &proposal); err != nil {
		return types.WorkloadGovernanceProposal{}, false
	}
	return proposal, true
}

// SetWorkloadProposal stores a proposal
func (k Keeper) SetWorkloadProposal(ctx sdk.Context, proposal types.WorkloadGovernanceProposal) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(proposal)
	if err != nil {
		return err
	}
	store.Set(types.GetWorkloadProposalKey(proposal.ProposalID), bz)
	return nil
}

// SetWorkloadVote stores a vote
func (k Keeper) SetWorkloadVote(ctx sdk.Context, vote types.WorkloadVote) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(vote)
	if err != nil {
		return err
	}
	store.Set(types.GetWorkloadVoteKey(vote.ProposalID, vote.Voter), bz)
	return nil
}

// GetWorkloadGovernanceParams returns governance parameters
func (k Keeper) GetWorkloadGovernanceParams(ctx sdk.Context) types.WorkloadGovernanceParams {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.WorkloadGovernanceParamsKey)
	if bz == nil {
		return types.DefaultWorkloadGovernanceParams()
	}
	var params types.WorkloadGovernanceParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultWorkloadGovernanceParams()
	}
	return params
}

// SetWorkloadGovernanceParams sets governance parameters
func (k Keeper) SetWorkloadGovernanceParams(ctx sdk.Context, params types.WorkloadGovernanceParams) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.WorkloadGovernanceParamsKey, bz)
	return nil
}

// ProcessWorkloadProposals processes proposals that need tallying
func (k Keeper) ProcessWorkloadProposals(ctx sdk.Context) error {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.WorkloadProposalPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var proposal types.WorkloadGovernanceProposal
		if err := json.Unmarshal(iter.Value(), &proposal); err != nil {
			continue
		}

		// Check if voting period ended and needs tallying
		if proposal.Status == types.WorkloadProposalStatusPending &&
			ctx.BlockTime().After(proposal.VotingEndTime) {
			if err := k.TallyWorkloadProposal(ctx, proposal.ProposalID); err != nil {
				k.Logger(ctx).Error("failed to tally proposal",
					"proposal_id", proposal.ProposalID, "error", err)
			}
		}
	}

	return nil
}

// GetNextWorkloadProposalSequence gets and increments the next proposal sequence
func (k Keeper) GetNextWorkloadProposalSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyWorkloadProposal)
}

// SetNextWorkloadProposalSequence sets the next proposal sequence
func (k Keeper) SetNextWorkloadProposalSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyWorkloadProposal, seq)
}

// SetNextWorkloadTemplateSequence sets the next template sequence
func (k Keeper) SetNextWorkloadTemplateSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyWorkloadTemplate, seq)
}

// shouldUpdateLatestTemplate determines if the latest template entry should be updated.
func shouldUpdateLatestTemplate(ctx sdk.Context, k Keeper, template types.WorkloadTemplate) bool {
	current, exists := k.GetWorkloadTemplate(ctx, template.TemplateID)
	if !exists {
		return true
	}

	if current.Version == template.Version {
		return true
	}

	return compareSemver(template.Version, current.Version) > 0
}

// compareSemver compares two semver strings (without leading "v").
// Returns 1 if a > b, -1 if a < b, 0 if equal or unparsable.
func compareSemver(a, b string) int {
	na := "v" + strings.TrimPrefix(a, "v")
	nb := "v" + strings.TrimPrefix(b, "v")
	if !semver.IsValid(na) || !semver.IsValid(nb) {
		return 0
	}
	return semver.Compare(na, nb)
}
