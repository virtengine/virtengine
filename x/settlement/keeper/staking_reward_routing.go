package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	"github.com/virtengine/virtengine/x/settlement/types"
)

const (
	stakeRoutingBasisPointsMax int64 = 10000
	stakeRoutingWeightScale    int64 = 1000000
)

type stakeRewardRoute struct {
	validator   sdk.AccAddress
	stake       int64
	totalShares *big.Int
	delegations []delegationtypes.Delegation
}

type stakeRewardRemainder struct {
	index     int
	validator string
	remainder *big.Int
}

func (k Keeper) availableStakingRewardPool(ctx sdk.Context) (sdk.Coins, error) {
	accrued := k.totalTreasuryAmountByType(ctx, types.TreasuryRecordValidatorFee)
	if accrued.IsZero() {
		return nil, types.ErrInvalidReward.Wrap("no validator fee rewards accrued")
	}

	distributed := k.totalDistributedStakingRewards(ctx)
	if distributed.IsZero() {
		return accrued, nil
	}
	if !accrued.IsAllGTE(distributed) {
		return nil, types.ErrInvalidReward.Wrap("staking rewards exceed accrued validator fee pool")
	}

	return accrued.Sub(distributed...), nil
}

func (k Keeper) totalTreasuryAmountByType(ctx sdk.Context, recordType types.TreasuryRecordType) sdk.Coins {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixTreasuryRecord)
	defer iter.Close()

	total := sdk.NewCoins()
	for ; iter.Valid(); iter.Next() {
		var record types.TreasuryRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if record.RecordType != recordType || record.Amount.IsZero() {
			continue
		}
		total = total.Add(record.Amount...)
	}

	return total
}

func (k Keeper) totalDistributedStakingRewards(ctx sdk.Context) sdk.Coins {
	total := sdk.NewCoins()
	k.WithRewardDistributions(ctx, func(dist types.RewardDistribution) bool {
		if dist.Source == types.RewardSourceStaking && !dist.TotalRewards.IsZero() {
			total = total.Add(dist.TotalRewards...)
		}
		return false
	})
	return total
}

func (k Keeper) buildStakeRewardRecipients(
	ctx sdk.Context,
	epoch uint64,
	rewardPool sdk.Coins,
) ([]types.RewardRecipient, error) {
	if k.stakeRoutingKeeper == nil {
		return nil, types.ErrInvalidReward.Wrap("stake routing keeper not configured")
	}

	routes, totalStake, err := k.snapshotStakeRewardRoutes(ctx)
	if err != nil {
		return nil, err
	}

	commissionRate := k.stakeRoutingKeeper.GetParams(ctx).ValidatorCommissionRate
	if commissionRate < 0 || commissionRate > stakeRoutingBasisPointsMax {
		commissionRate = delegationtypes.DefaultValidatorCommissionRate
	}

	allocations := allocateStakeRewardPool(routes, totalStake, rewardPool)
	recipients := make([]types.RewardRecipient, 0, len(routes))
	for idx, route := range routes {
		if allocations[idx].IsZero() {
			continue
		}

		splitRecipients, splitErr := buildStakeRecipientsForValidator(route, totalStake, epoch, commissionRate, allocations[idx])
		if splitErr != nil {
			return nil, splitErr
		}
		recipients = append(recipients, splitRecipients...)
	}

	if len(recipients) == 0 {
		return nil, types.ErrInvalidReward.Wrap("no staking rewards to distribute")
	}

	return recipients, nil
}

func (k Keeper) snapshotStakeRewardRoutes(ctx sdk.Context) ([]stakeRewardRoute, int64, error) {
	validators := append([]sdk.AccAddress(nil), k.stakeRoutingKeeper.GetAllValidators(ctx)...)
	if len(validators) == 0 {
		return nil, 0, types.ErrInvalidReward.Wrap("no validators available for stake routing")
	}

	sort.Slice(validators, func(i, j int) bool {
		return validators[i].String() < validators[j].String()
	})

	visibleStake := int64(0)
	routes := make([]stakeRewardRoute, 0, len(validators))
	for _, validator := range validators {
		stake := k.stakeRoutingKeeper.GetValidatorStake(ctx, validator)
		if stake <= 0 {
			continue
		}

		shares, found := k.stakeRoutingKeeper.GetValidatorShares(ctx, validator.String())
		totalShares := big.NewInt(0)
		if found {
			totalShares = shares.GetTotalSharesBigInt()
		}

		delegations := append([]delegationtypes.Delegation(nil), k.stakeRoutingKeeper.GetValidatorDelegations(ctx, validator.String())...)
		sort.Slice(delegations, func(i, j int) bool {
			return delegations[i].DelegatorAddress < delegations[j].DelegatorAddress
		})

		routes = append(routes, stakeRewardRoute{
			validator:   validator,
			stake:       stake,
			totalShares: totalShares,
			delegations: delegations,
		})
		visibleStake += stake
	}

	if visibleStake <= 0 {
		return nil, 0, types.ErrInvalidReward.Wrap("no positive validator stake available for routing")
	}

	totalStake := k.stakeRoutingKeeper.GetTotalStake(ctx)
	if totalStake < visibleStake {
		totalStake = visibleStake
	}
	if totalStake <= 0 {
		return nil, 0, types.ErrInvalidReward.Wrap("total stake is zero")
	}

	return routes, totalStake, nil
}

func allocateStakeRewardPool(routes []stakeRewardRoute, totalStake int64, rewardPool sdk.Coins) []sdk.Coins {
	allocations := make([]sdk.Coins, len(routes))
	for _, coin := range rewardPool {
		if !coin.Amount.IsPositive() {
			continue
		}

		distribution := allocateStakeRewardCoin(routes, totalStake, coin.Amount.BigInt())
		for idx, amount := range distribution {
			if amount == nil || amount.Sign() <= 0 {
				continue
			}
			allocations[idx] = allocations[idx].Add(sdk.NewCoin(coin.Denom, sdkmath.NewIntFromBigInt(amount)))
		}
	}

	return allocations
}

func allocateStakeRewardCoin(routes []stakeRewardRoute, totalStake int64, totalAmount *big.Int) []*big.Int {
	results := make([]*big.Int, len(routes))
	if totalStake <= 0 || totalAmount == nil || totalAmount.Sign() <= 0 {
		for idx := range results {
			results[idx] = big.NewInt(0)
		}
		return results
	}

	denominator := big.NewInt(totalStake)
	remainders := make([]stakeRewardRemainder, 0, len(routes))
	totalNumerator := big.NewInt(0)
	allocated := big.NewInt(0)

	for idx, route := range routes {
		results[idx] = big.NewInt(0)
		if route.stake <= 0 {
			remainders = append(remainders, stakeRewardRemainder{
				index:     idx,
				validator: route.validator.String(),
				remainder: big.NewInt(0),
			})
			continue
		}

		product := new(big.Int).Mul(new(big.Int).Set(totalAmount), big.NewInt(route.stake))
		totalNumerator.Add(totalNumerator, product)

		quotient := new(big.Int)
		remainder := new(big.Int)
		quotient.QuoRem(product, denominator, remainder)
		results[idx] = quotient
		allocated.Add(allocated, quotient)

		remainders = append(remainders, stakeRewardRemainder{
			index:     idx,
			validator: route.validator.String(),
			remainder: remainder,
		})
	}

	totalEarned := new(big.Int).Quo(totalNumerator, denominator)
	leftover := new(big.Int).Sub(totalEarned, allocated)
	if leftover.Sign() <= 0 {
		return results
	}

	sort.SliceStable(remainders, func(i, j int) bool {
		cmp := remainders[i].remainder.Cmp(remainders[j].remainder)
		if cmp != 0 {
			return cmp > 0
		}
		return remainders[i].validator < remainders[j].validator
	})

	for idx := 0; idx < len(remainders) && leftover.Sign() > 0; idx++ {
		results[remainders[idx].index].Add(results[remainders[idx].index], big.NewInt(1))
		leftover.Sub(leftover, big.NewInt(1))
	}

	return results
}

func buildStakeRecipientsForValidator(
	route stakeRewardRoute,
	totalStake int64,
	epoch uint64,
	commissionRate int64,
	reward sdk.Coins,
) ([]types.RewardRecipient, error) {
	validatorReward := sdk.NewCoins()
	delegatorRewards := make(map[string]sdk.Coins)
	delegatorShares := make(map[string]*big.Int)
	totalShares := route.totalShares
	hasDelegations := len(route.delegations) > 0 && totalShares != nil && totalShares.Sign() > 0

	for _, delegation := range route.delegations {
		shares := delegation.GetSharesBigInt()
		if shares.Sign() <= 0 {
			continue
		}
		delegatorShares[delegation.DelegatorAddress] = shares
	}
	if len(delegatorShares) == 0 {
		hasDelegations = false
	}

	for _, coin := range reward {
		if !coin.Amount.IsPositive() {
			continue
		}
		if !hasDelegations {
			validatorReward = validatorReward.Add(coin)
			continue
		}

		commission := new(big.Int).Mul(coin.Amount.BigInt(), big.NewInt(commissionRate))
		commission.Div(commission, big.NewInt(stakeRoutingBasisPointsMax))
		distributable := new(big.Int).Sub(coin.Amount.BigInt(), commission)

		distributedToDelegators := big.NewInt(0)
		for _, delegation := range route.delegations {
			shares := delegatorShares[delegation.DelegatorAddress]
			if shares == nil || shares.Sign() <= 0 {
				continue
			}

			delegatorAmount := new(big.Int).Mul(distributable, shares)
			delegatorAmount.Div(delegatorAmount, totalShares)
			if delegatorAmount.Sign() <= 0 {
				continue
			}

			distributedToDelegators.Add(distributedToDelegators, delegatorAmount)
			delegatorRewards[delegation.DelegatorAddress] = delegatorRewards[delegation.DelegatorAddress].Add(
				sdk.NewCoin(coin.Denom, sdkmath.NewIntFromBigInt(delegatorAmount)),
			)
		}

		validatorAmount := new(big.Int).Add(commission, new(big.Int).Sub(distributable, distributedToDelegators))
		if validatorAmount.Sign() > 0 {
			validatorReward = validatorReward.Add(sdk.NewCoin(coin.Denom, sdkmath.NewIntFromBigInt(validatorAmount)))
		}
	}

	recipients := make([]types.RewardRecipient, 0, 1+len(delegatorRewards))
	if !validatorReward.IsZero() {
		recipients = append(recipients, types.RewardRecipient{
			Address:       route.validator.String(),
			Amount:        validatorReward,
			Reason:        "staking_validator_reward",
			StakingWeight: formatScaledStakeWeight(big.NewInt(route.stake), big.NewInt(totalStake)),
			ReferenceID:   fmt.Sprintf("%s:%d", route.validator.String(), epoch),
		})
	}

	delegatorAddrs := make([]string, 0, len(delegatorRewards))
	for addr := range delegatorRewards {
		delegatorAddrs = append(delegatorAddrs, addr)
	}
	sort.Strings(delegatorAddrs)

	totalStakeBig := big.NewInt(totalStake)
	validatorStakeBig := big.NewInt(route.stake)
	totalWeightDenominator := new(big.Int).Mul(totalStakeBig, totalShares)
	for _, addr := range delegatorAddrs {
		rewardCoins := delegatorRewards[addr]
		if rewardCoins.IsZero() {
			continue
		}
		shareWeight := new(big.Int).Mul(new(big.Int).Set(validatorStakeBig), delegatorShares[addr])
		recipients = append(recipients, types.RewardRecipient{
			Address:       addr,
			Amount:        rewardCoins,
			Reason:        "staking_delegator_reward",
			StakingWeight: formatScaledStakeWeight(shareWeight, totalWeightDenominator),
			ReferenceID:   route.validator.String(),
		})
	}

	if len(recipients) == 0 {
		return nil, types.ErrInvalidReward.Wrap("stake reward split produced no recipients")
	}

	return recipients, nil
}

func formatScaledStakeWeight(numerator, denominator *big.Int) string {
	if numerator == nil || denominator == nil || numerator.Sign() <= 0 || denominator.Sign() <= 0 {
		return "0"
	}

	scaled := new(big.Int).Mul(new(big.Int).Set(numerator), big.NewInt(stakeRoutingWeightScale))
	scaled.Div(scaled, denominator)
	return scaled.String()
}
