package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// GetRewardHistory returns reward history entries for an address.
func (k Keeper) GetRewardHistory(
	ctx sdk.Context,
	address string,
	source string,
	limit uint32,
	offset uint32,
) ([]types.RewardHistoryEntry, error) {
	if address == "" {
		return nil, types.ErrInvalidAddress.Wrap("address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid address")
	}
	if source != "" && !types.IsValidRewardSource(types.RewardSource(source)) {
		return nil, types.ErrInvalidReward.Wrap("invalid reward source")
	}

	entries := make([]types.RewardHistoryEntry, 0)

	k.WithRewardDistributions(ctx, func(dist types.RewardDistribution) bool {
		if source != "" && string(dist.Source) != source {
			return false
		}

		for _, recipient := range dist.Recipients {
			if recipient.Address != address {
				continue
			}

			entries = append(entries, types.RewardHistoryEntry{
				DistributionID: dist.DistributionID,
				EpochNumber:    dist.EpochNumber,
				Source:         dist.Source,
				Amount:         recipient.Amount,
				Reason:         recipient.Reason,
				UsageUnits:     recipient.UsageUnits,
				ReferenceID:    recipient.ReferenceID,
				DistributedAt:  dist.DistributedAt,
			})
		}

		return false
	})

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].DistributedAt.Equal(entries[j].DistributedAt) {
			return entries[i].DistributionID < entries[j].DistributionID
		}
		return entries[i].DistributedAt.After(entries[j].DistributedAt)
	})

	offsetInt := int(offset)
	if offsetInt < 0 || offsetInt >= len(entries) {
		return []types.RewardHistoryEntry{}, nil
	}

	entries = entries[offsetInt:]
	if limit > 0 {
		limitInt := int(limit)
		if limitInt < len(entries) {
			entries = entries[:limitInt]
		}
	}

	return entries, nil
}
