package types

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardSource indicates the source of rewards
type RewardSource string

const (
	RewardSourceStaking      RewardSource = "staking"
	RewardSourceUsage        RewardSource = "usage"
	RewardSourceVerification RewardSource = "verification"
	RewardSourceProvider     RewardSource = "provider"
)

// IsValidRewardSource checks if the source is valid
func IsValidRewardSource(source RewardSource) bool {
	switch source {
	case RewardSourceStaking, RewardSourceUsage, RewardSourceVerification, RewardSourceProvider:
		return true
	default:
		return false
	}
}

// RewardRecipient represents a single reward recipient
type RewardRecipient struct {
	// Address is the recipient's address
	Address string `json:"address"`

	// Amount is the reward amount
	Amount sdk.Coins `json:"amount"`

	// Reason describes why this reward was given
	Reason string `json:"reason"`

	// UsageUnits is the usage units for usage-based rewards
	UsageUnits uint64 `json:"usage_units,omitempty"`

	// VerificationScore is the verification score for verification rewards
	VerificationScore uint32 `json:"verification_score,omitempty"`

	// StakingWeight is the staking weight used for calculation
	StakingWeight string `json:"staking_weight,omitempty"`

	// ReferenceID is a reference to the source record (order, lease, verification, etc.)
	ReferenceID string `json:"reference_id,omitempty"`
}

// Validate validates a reward recipient
func (r *RewardRecipient) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.Address); err != nil {
		return ErrInvalidReward.Wrap("invalid recipient address")
	}

	if !r.Amount.IsValid() || r.Amount.IsZero() {
		return ErrInvalidReward.Wrap("reward amount must be valid and non-zero")
	}

	if r.Reason == "" {
		return ErrInvalidReward.Wrap("reason cannot be empty")
	}

	return nil
}

// RewardDistribution represents a batch of rewards distributed
type RewardDistribution struct {
	// DistributionID is the unique identifier for this distribution
	DistributionID string `json:"distribution_id"`

	// EpochNumber is the epoch this distribution belongs to
	EpochNumber uint64 `json:"epoch_number"`

	// TotalRewards is the total rewards distributed
	TotalRewards sdk.Coins `json:"total_rewards"`

	// Recipients is the list of reward recipients
	Recipients []RewardRecipient `json:"recipients"`

	// Source indicates the reward source
	Source RewardSource `json:"source"`

	// DistributedAt is when the distribution occurred
	DistributedAt time.Time `json:"distributed_at"`

	// BlockHeight is when the distribution was recorded
	BlockHeight int64 `json:"block_height"`

	// ReferenceTxHashes are the transaction hashes that triggered this distribution
	ReferenceTxHashes []string `json:"reference_tx_hashes,omitempty"`

	// Metadata contains additional distribution details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewRewardDistribution creates a new reward distribution
func NewRewardDistribution(
	distributionID string,
	epochNumber uint64,
	source RewardSource,
	recipients []RewardRecipient,
	blockTime time.Time,
	blockHeight int64,
) *RewardDistribution {
	// Calculate total rewards
	totalRewards := sdk.NewCoins()
	for _, r := range recipients {
		totalRewards = totalRewards.Add(r.Amount...)
	}

	return &RewardDistribution{
		DistributionID: distributionID,
		EpochNumber:    epochNumber,
		TotalRewards:   totalRewards,
		Recipients:     recipients,
		Source:         source,
		DistributedAt:  blockTime,
		BlockHeight:    blockHeight,
		Metadata:       make(map[string]string),
	}
}

// Validate validates a reward distribution
func (r *RewardDistribution) Validate() error {
	if r.DistributionID == "" {
		return ErrInvalidReward.Wrap("distribution_id cannot be empty")
	}

	if len(r.DistributionID) > 64 {
		return ErrInvalidReward.Wrap("distribution_id exceeds maximum length")
	}

	if !IsValidRewardSource(r.Source) {
		return ErrInvalidReward.Wrapf("invalid reward source: %s", r.Source)
	}

	if len(r.Recipients) == 0 {
		return ErrInvalidReward.Wrap("recipients cannot be empty")
	}

	// Validate each recipient
	calculatedTotal := sdk.NewCoins()
	for i, recipient := range r.Recipients {
		if err := recipient.Validate(); err != nil {
			return ErrInvalidReward.Wrapf("invalid recipient %d: %s", i, err.Error())
		}
		calculatedTotal = calculatedTotal.Add(recipient.Amount...)
	}

	// Validate total matches
	if !calculatedTotal.IsEqual(r.TotalRewards) {
		return ErrInvalidReward.Wrap("total_rewards must equal sum of recipient amounts")
	}

	return nil
}

// ClaimableRewards represents rewards that can be claimed by an address
type ClaimableRewards struct {
	// Address is the account address
	Address string `json:"address"`

	// TotalClaimable is the total claimable amount
	TotalClaimable sdk.Coins `json:"total_claimable"`

	// RewardEntries are the individual reward entries
	RewardEntries []RewardEntry `json:"reward_entries"`

	// LastUpdated is when this was last updated
	LastUpdated time.Time `json:"last_updated"`

	// TotalClaimed is the total amount claimed historically
	TotalClaimed sdk.Coins `json:"total_claimed"`
}

// RewardEntry represents a single claimable reward entry
type RewardEntry struct {
	// DistributionID is the source distribution
	DistributionID string `json:"distribution_id"`

	// Source is the reward source
	Source RewardSource `json:"source"`

	// Amount is the claimable amount
	Amount sdk.Coins `json:"amount"`

	// CreatedAt is when this entry was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this entry expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Reason describes the reward
	Reason string `json:"reason"`
}

// NewClaimableRewards creates a new claimable rewards record
func NewClaimableRewards(address string, blockTime time.Time) *ClaimableRewards {
	return &ClaimableRewards{
		Address:        address,
		TotalClaimable: sdk.NewCoins(),
		RewardEntries:  []RewardEntry{},
		LastUpdated:    blockTime,
		TotalClaimed:   sdk.NewCoins(),
	}
}

// AddReward adds a reward entry
func (c *ClaimableRewards) AddReward(entry RewardEntry) {
	c.RewardEntries = append(c.RewardEntries, entry)
	c.TotalClaimable = c.TotalClaimable.Add(entry.Amount...)
}

// ClaimAll claims all available rewards
func (c *ClaimableRewards) ClaimAll(blockTime time.Time) (sdk.Coins, []RewardEntry) {
	claimed := c.TotalClaimable
	claimedEntries := c.RewardEntries

	c.TotalClaimable = sdk.NewCoins()
	c.RewardEntries = []RewardEntry{}
	c.TotalClaimed = c.TotalClaimed.Add(claimed...)
	c.LastUpdated = blockTime

	return claimed, claimedEntries
}

// ClaimBySource claims rewards from a specific source
func (c *ClaimableRewards) ClaimBySource(source RewardSource, blockTime time.Time) (sdk.Coins, []RewardEntry) {
	claimed := sdk.NewCoins()
	var claimedEntries []RewardEntry
	var remainingEntries []RewardEntry

	for _, entry := range c.RewardEntries {
		if entry.Source == source {
			claimed = claimed.Add(entry.Amount...)
			claimedEntries = append(claimedEntries, entry)
		} else {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	c.RewardEntries = remainingEntries
	c.TotalClaimable = sdk.NewCoins()
	for _, entry := range remainingEntries {
		c.TotalClaimable = c.TotalClaimable.Add(entry.Amount...)
	}
	c.TotalClaimed = c.TotalClaimed.Add(claimed...)
	c.LastUpdated = blockTime

	return claimed, claimedEntries
}

// RemoveExpired removes expired reward entries
func (c *ClaimableRewards) RemoveExpired(blockTime time.Time) sdk.Coins {
	expired := sdk.NewCoins()
	var remainingEntries []RewardEntry

	for _, entry := range c.RewardEntries {
		if entry.ExpiresAt != nil && blockTime.After(*entry.ExpiresAt) {
			expired = expired.Add(entry.Amount...)
		} else {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	c.RewardEntries = remainingEntries
	c.TotalClaimable = sdk.NewCoins()
	for _, entry := range remainingEntries {
		c.TotalClaimable = c.TotalClaimable.Add(entry.Amount...)
	}
	c.LastUpdated = blockTime

	return expired
}

// MarshalJSON implements json.Marshaler
func (r RewardDistribution) MarshalJSON() ([]byte, error) {
	type Alias RewardDistribution
	return json.Marshal(&struct {
		Alias
		TotalRewards []sdk.Coin `json:"total_rewards"`
	}{
		Alias:        (Alias)(r),
		TotalRewards: r.TotalRewards,
	})
}
