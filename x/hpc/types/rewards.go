// Package types contains types for the HPC module.
//
// VE-504: Rewards distribution for HPC contributors
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HPCRewardSource indicates the source of HPC rewards
type HPCRewardSource string

const (
	// HPCRewardSourceJobCompletion is for completed job rewards
	HPCRewardSourceJobCompletion HPCRewardSource = "job_completion"

	// HPCRewardSourceUsage is for usage-based rewards
	HPCRewardSourceUsage HPCRewardSource = "usage"

	// HPCRewardSourceBonus is for bonus rewards
	HPCRewardSourceBonus HPCRewardSource = "bonus"
)

// IsValidHPCRewardSource checks if the source is valid
func IsValidHPCRewardSource(source HPCRewardSource) bool {
	switch source {
	case HPCRewardSourceJobCompletion, HPCRewardSourceUsage, HPCRewardSourceBonus:
		return true
	default:
		return false
	}
}

// HPCRewardRecord represents a reward distribution for HPC contribution
type HPCRewardRecord struct {
	// RewardID is the unique identifier
	RewardID string `json:"reward_id"`

	// JobID is the job this reward is for
	JobID string `json:"job_id"`

	// ClusterID is the cluster that processed the job
	ClusterID string `json:"cluster_id"`

	// Source is the reward source
	Source HPCRewardSource `json:"source"`

	// TotalReward is the total reward amount
	TotalReward sdk.Coins `json:"total_reward"`

	// Recipients are the reward recipients
	Recipients []HPCRewardRecipient `json:"recipients"`

	// ReferencedUsageRecords are the usage record IDs used for calculation
	ReferencedUsageRecords []string `json:"referenced_usage_records"`

	// JobCompletionStatus is the job's completion status
	JobCompletionStatus JobState `json:"job_completion_status"`

	// FormulaVersion is the reward formula version used
	FormulaVersion string `json:"formula_version"`

	// CalculationDetails contains calculation transparency data
	CalculationDetails RewardCalculationDetails `json:"calculation_details"`

	// Disputed indicates if this reward has been disputed
	Disputed bool `json:"disputed"`

	// DisputeID links to the dispute if any
	DisputeID string `json:"dispute_id,omitempty"`

	// IssuedAt is when the reward was issued
	IssuedAt time.Time `json:"issued_at"`

	// BlockHeight is when the reward was recorded
	BlockHeight int64 `json:"block_height"`
}

// HPCRewardRecipient represents a recipient of HPC rewards
type HPCRewardRecipient struct {
	// Address is the recipient address
	Address string `json:"address"`

	// Amount is the reward amount
	Amount sdk.Coins `json:"amount"`

	// RecipientType is the type of recipient (provider, node_operator, platform)
	RecipientType string `json:"recipient_type"`

	// NodeID is the node ID if this is a node operator reward
	NodeID string `json:"node_id,omitempty"`

	// ContributionWeight is the contribution weight (fixed-point, 6 decimals)
	ContributionWeight string `json:"contribution_weight"`

	// Reason describes the reward
	Reason string `json:"reason"`
}

// RewardCalculationDetails contains transparency data for reward calculation
type RewardCalculationDetails struct {
	// TotalUsageValue is the total usage value in base units
	TotalUsageValue string `json:"total_usage_value"`

	// RewardPoolContribution is the contribution to reward pool
	RewardPoolContribution string `json:"reward_pool_contribution"`

	// PlatformFeeRate is the platform fee rate used (fixed-point, 6 decimals)
	PlatformFeeRate string `json:"platform_fee_rate"`

	// NodeContributionFormula documents the formula used
	NodeContributionFormula string `json:"node_contribution_formula"`

	// InputMetrics are the input metrics used
	InputMetrics map[string]string `json:"input_metrics"`
}

// Validate validates an HPC reward record
func (r *HPCRewardRecord) Validate() error {
	if r.RewardID == "" {
		return ErrInvalidReward.Wrap("reward_id cannot be empty")
	}

	if r.JobID == "" {
		return ErrInvalidReward.Wrap("job_id cannot be empty")
	}

	if r.ClusterID == "" {
		return ErrInvalidReward.Wrap("cluster_id cannot be empty")
	}

	if !IsValidHPCRewardSource(r.Source) {
		return ErrInvalidReward.Wrapf("invalid reward source: %s", r.Source)
	}

	if !r.TotalReward.IsValid() || r.TotalReward.IsZero() {
		return ErrInvalidReward.Wrap("total_reward must be valid and non-zero")
	}

	if len(r.Recipients) == 0 {
		return ErrInvalidReward.Wrap("recipients cannot be empty")
	}

	// Validate recipients sum to total
	calculatedTotal := sdk.NewCoins()
	for i, recipient := range r.Recipients {
		if _, err := sdk.AccAddressFromBech32(recipient.Address); err != nil {
			return ErrInvalidReward.Wrapf("invalid recipient address at index %d", i)
		}
		if !recipient.Amount.IsValid() || recipient.Amount.IsZero() {
			return ErrInvalidReward.Wrapf("invalid recipient amount at index %d", i)
		}
		calculatedTotal = calculatedTotal.Add(recipient.Amount...)
	}

	if !calculatedTotal.IsEqual(r.TotalReward) {
		return ErrInvalidReward.Wrap("recipient amounts must sum to total_reward")
	}

	return nil
}

// DisputeStatus indicates the status of a dispute
type DisputeStatus string

const (
	// DisputeStatusPending indicates the dispute is pending
	DisputeStatusPending DisputeStatus = "pending"

	// DisputeStatusUnderReview indicates the dispute is under review
	DisputeStatusUnderReview DisputeStatus = "under_review"

	// DisputeStatusResolved indicates the dispute is resolved
	DisputeStatusResolved DisputeStatus = "resolved"

	// DisputeStatusRejected indicates the dispute was rejected
	DisputeStatusRejected DisputeStatus = "rejected"
)

// IsValidDisputeStatus checks if the status is valid
func IsValidDisputeStatus(s DisputeStatus) bool {
	switch s {
	case DisputeStatusPending, DisputeStatusUnderReview, DisputeStatusResolved, DisputeStatusRejected:
		return true
	default:
		return false
	}
}

// HPCDispute represents a dispute for HPC rewards/usage
type HPCDispute struct {
	// DisputeID is the unique identifier
	DisputeID string `json:"dispute_id"`

	// JobID is the job being disputed
	JobID string `json:"job_id"`

	// RewardID is the reward being disputed (if applicable)
	RewardID string `json:"reward_id,omitempty"`

	// DisputerAddress is who filed the dispute
	DisputerAddress string `json:"disputer_address"`

	// DisputeType describes what is being disputed
	DisputeType string `json:"dispute_type"`

	// Reason is the reason for the dispute
	Reason string `json:"reason"`

	// Evidence contains evidence supporting the dispute
	Evidence string `json:"evidence,omitempty"`

	// Status is the dispute status
	Status DisputeStatus `json:"status"`

	// Resolution is the resolution if resolved
	Resolution string `json:"resolution,omitempty"`

	// ResolverAddress is who resolved the dispute
	ResolverAddress string `json:"resolver_address,omitempty"`

	// CreatedAt is when the dispute was created
	CreatedAt time.Time `json:"created_at"`

	// ResolvedAt is when the dispute was resolved
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// BlockHeight is when the dispute was recorded
	BlockHeight int64 `json:"block_height"`
}

// Validate validates a dispute
func (d *HPCDispute) Validate() error {
	if d.DisputeID == "" {
		return ErrInvalidDispute.Wrap("dispute_id cannot be empty")
	}

	if d.JobID == "" {
		return ErrInvalidDispute.Wrap("job_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(d.DisputerAddress); err != nil {
		return ErrInvalidDispute.Wrap("invalid disputer address")
	}

	if d.DisputeType == "" {
		return ErrInvalidDispute.Wrap("dispute_type cannot be empty")
	}

	if d.Reason == "" {
		return ErrInvalidDispute.Wrap("reason cannot be empty")
	}

	if !IsValidDisputeStatus(d.Status) {
		return ErrInvalidDispute.Wrapf("invalid dispute status: %s", d.Status)
	}

	return nil
}
