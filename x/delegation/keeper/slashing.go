// Package keeper implements the delegation module keeper.
//
// VE-922: Delegator slashing logic for validator misbehavior
package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

type delegatorSlashTotals struct {
	amount *big.Int
	shares *big.Int
}

// SlashDelegations slashes delegations, unbondings, and redelegations for a validator.
func (k Keeper) SlashDelegations(ctx sdk.Context, validatorAddr string, fraction sdkmath.LegacyDec, infractionHeight int64) error {
	if fraction.IsNegative() || fraction.IsZero() {
		return nil
	}

	oneDec := sdkmath.LegacyOneDec()
	if fraction.GT(oneDec) {
		fraction = oneDec
	}

	slashTotals := make(map[string]*delegatorSlashTotals)

	if err := k.slashActiveDelegations(ctx, validatorAddr, fraction, infractionHeight, slashTotals); err != nil {
		return err
	}

	if err := k.slashUnbondingDelegations(ctx, validatorAddr, fraction, infractionHeight, slashTotals); err != nil {
		return err
	}

	if err := k.slashRedelegations(ctx, validatorAddr, fraction, infractionHeight, slashTotals); err != nil {
		return err
	}

	for delegatorAddr, totals := range slashTotals {
		if totals.amount.Sign() <= 0 && totals.shares.Sign() <= 0 {
			continue
		}

		event := types.NewDelegatorSlashingEvent(
			"",
			delegatorAddr,
			validatorAddr,
			fraction.String(),
			totals.amount.String(),
			totals.shares.String(),
			infractionHeight,
			ctx.BlockHeight(),
			ctx.BlockTime(),
		)

		if err := k.SetDelegatorSlashingEvent(ctx, *event); err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDelegatorSlashed,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
				sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
				sdk.NewAttribute(types.AttributeKeySlashAmount, totals.amount.String()),
				sdk.NewAttribute(types.AttributeKeySlashFraction, fraction.String()),
				sdk.NewAttribute(types.AttributeKeySlashShares, totals.shares.String()),
				sdk.NewAttribute(types.AttributeKeyInfractionHeight, fmt.Sprintf("%d", infractionHeight)),
			),
		)
	}

	return nil
}

func (k Keeper) slashActiveDelegations(
	ctx sdk.Context,
	validatorAddr string,
	fraction sdkmath.LegacyDec,
	infractionHeight int64,
	slashTotals map[string]*delegatorSlashTotals,
) error {
	valShares, found := k.GetValidatorShares(ctx, validatorAddr)
	if !found {
		return nil
	}

	totalSharesBefore := valShares.GetTotalSharesBigInt()
	totalStakeBefore := valShares.GetTotalStakeBigInt()
	if totalSharesBefore.Sign() == 0 || totalStakeBefore.Sign() == 0 {
		return nil
	}

	delegations := k.GetValidatorDelegations(ctx, validatorAddr)
	totalSharesSlashed := big.NewInt(0)
	totalStakeSlashed := big.NewInt(0)

	for _, del := range delegations {
		if del.Height > infractionHeight {
			continue
		}

		delShares := del.GetSharesBigInt()
		if delShares.Sign() == 0 {
			continue
		}

		slashShares := slashBigIntWithFraction(delShares, fraction)
		if slashShares.Sign() == 0 {
			continue
		}

		slashAmount := calculateAmountForShares(slashShares, totalSharesBefore, totalStakeBefore)
		if slashAmount.Sign() == 0 && slashShares.Sign() == 0 {
			continue
		}

		newShares := new(big.Int).Sub(delShares, slashShares)
		if newShares.Sign() <= 0 {
			k.DeleteDelegation(ctx, del.DelegatorAddress, del.ValidatorAddress)
		} else {
			del.Shares = newShares.String()
			del.UpdatedAt = ctx.BlockTime()
			if err := k.SetDelegation(ctx, del); err != nil {
				return err
			}
		}

		totalSharesSlashed.Add(totalSharesSlashed, slashShares)
		totalStakeSlashed.Add(totalStakeSlashed, slashAmount)

		addDelegatorSlash(slashTotals, del.DelegatorAddress, slashAmount, slashShares)
	}

	if totalSharesSlashed.Sign() == 0 && totalStakeSlashed.Sign() == 0 {
		return nil
	}

	newTotalShares := new(big.Int).Sub(totalSharesBefore, totalSharesSlashed)
	newTotalStake := new(big.Int).Sub(totalStakeBefore, totalStakeSlashed)
	if newTotalShares.Sign() < 0 {
		newTotalShares = big.NewInt(0)
	}
	if newTotalStake.Sign() < 0 {
		newTotalStake = big.NewInt(0)
	}

	valShares.TotalShares = newTotalShares.String()
	valShares.TotalStake = newTotalStake.String()
	valShares.UpdatedAt = ctx.BlockTime()
	return k.SetValidatorShares(ctx, valShares)
}

func (k Keeper) slashUnbondingDelegations(
	ctx sdk.Context,
	validatorAddr string,
	fraction sdkmath.LegacyDec,
	infractionHeight int64,
	slashTotals map[string]*delegatorSlashTotals,
) error {
	unbondings := k.GetValidatorUnbondingDelegations(ctx, validatorAddr)
	now := ctx.BlockTime()

	for _, ubd := range unbondings {
		changed := false

		for idx, entry := range ubd.Entries {
			if entry.CreationHeight > infractionHeight || !entry.CompletionTime.After(now) {
				continue
			}

			balance := parseBigInt(entry.Balance)
			if balance.Sign() == 0 {
				continue
			}

			slashAmount := slashBigIntWithFraction(balance, fraction)
			if slashAmount.Sign() == 0 {
				continue
			}

			unbondingShares := parseBigInt(entry.UnbondingShares)
			slashShares := slashBigIntWithFraction(unbondingShares, fraction)

			newBalance := new(big.Int).Sub(balance, slashAmount)
			if newBalance.Sign() < 0 {
				newBalance = big.NewInt(0)
			}

			newInitial := new(big.Int).Sub(parseBigInt(entry.InitialBalance), slashAmount)
			if newInitial.Sign() < 0 {
				newInitial = big.NewInt(0)
			}

			newUnbondingShares := new(big.Int).Sub(unbondingShares, slashShares)
			if newUnbondingShares.Sign() < 0 {
				newUnbondingShares = big.NewInt(0)
			}

			entry.Balance = newBalance.String()
			entry.InitialBalance = newInitial.String()
			entry.UnbondingShares = newUnbondingShares.String()
			ubd.Entries[idx] = entry
			changed = true

			addDelegatorSlash(slashTotals, ubd.DelegatorAddress, slashAmount, slashShares)
		}

		if changed {
			if err := k.SetUnbondingDelegation(ctx, ubd); err != nil {
				return err
			}
		}
	}

	return nil
}

func (k Keeper) slashRedelegations(
	ctx sdk.Context,
	validatorAddr string,
	fraction sdkmath.LegacyDec,
	infractionHeight int64,
	slashTotals map[string]*delegatorSlashTotals,
) error {
	now := ctx.BlockTime()
	var outErr error

	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		if red.ValidatorSrcAddress != validatorAddr {
			return false
		}

		changed := false
		for idx, entry := range red.Entries {
			if entry.CreationHeight > infractionHeight || !entry.CompletionTime.After(now) {
				continue
			}

			sharesDst := parseBigInt(entry.SharesDst)
			if sharesDst.Sign() == 0 {
				continue
			}

			slashShares := slashBigIntWithFraction(sharesDst, fraction)
			if slashShares.Sign() == 0 {
				continue
			}

			dstValShares, found := k.GetValidatorShares(ctx, red.ValidatorDstAddress)
			if !found {
				continue
			}

			totalSharesBefore := dstValShares.GetTotalSharesBigInt()
			totalStakeBefore := dstValShares.GetTotalStakeBigInt()
			if totalSharesBefore.Sign() == 0 || totalStakeBefore.Sign() == 0 {
				continue
			}

			slashAmount := calculateAmountForShares(slashShares, totalSharesBefore, totalStakeBefore)

			dstDel, found := k.GetDelegation(ctx, red.DelegatorAddress, red.ValidatorDstAddress)
			if found {
				dstShares := dstDel.GetSharesBigInt()
				newDstShares := new(big.Int).Sub(dstShares, slashShares)
				if newDstShares.Sign() <= 0 {
					k.DeleteDelegation(ctx, red.DelegatorAddress, red.ValidatorDstAddress)
				} else {
					dstDel.Shares = newDstShares.String()
					dstDel.UpdatedAt = ctx.BlockTime()
					if err := k.SetDelegation(ctx, dstDel); err != nil {
						outErr = err
						return true
					}
				}
			}

			newTotalShares := new(big.Int).Sub(totalSharesBefore, slashShares)
			newTotalStake := new(big.Int).Sub(totalStakeBefore, slashAmount)
			if newTotalShares.Sign() < 0 {
				newTotalShares = big.NewInt(0)
			}
			if newTotalStake.Sign() < 0 {
				newTotalStake = big.NewInt(0)
			}

			dstValShares.TotalShares = newTotalShares.String()
			dstValShares.TotalStake = newTotalStake.String()
			dstValShares.UpdatedAt = ctx.BlockTime()
			if err := k.SetValidatorShares(ctx, dstValShares); err != nil {
				outErr = err
				return true
			}

			entry.SharesDst = new(big.Int).Sub(sharesDst, slashShares).String()
			newInitial := new(big.Int).Sub(parseBigInt(entry.InitialBalance), slashAmount)
			if newInitial.Sign() < 0 {
				newInitial = big.NewInt(0)
			}
			entry.InitialBalance = newInitial.String()
			red.Entries[idx] = entry
			changed = true

			addDelegatorSlash(slashTotals, red.DelegatorAddress, slashAmount, slashShares)
		}

		if changed {
			if err := k.SetRedelegation(ctx, red); err != nil {
				outErr = err
				return true
			}
		}

		return false
	})

	return outErr
}

func slashBigIntWithFraction(value *big.Int, fraction sdkmath.LegacyDec) *big.Int {
	if value.Sign() <= 0 {
		return big.NewInt(0)
	}
	dec := fraction.MulInt(sdkmath.NewIntFromBigInt(value))
	return dec.TruncateInt().BigInt()
}

func calculateAmountForShares(shares, totalShares, totalStake *big.Int) *big.Int {
	if shares.Sign() <= 0 || totalShares.Sign() == 0 || totalStake.Sign() == 0 {
		return big.NewInt(0)
	}
	amount := new(big.Int).Mul(shares, totalStake)
	amount.Div(amount, totalShares)
	return amount
}

func parseBigInt(value string) *big.Int {
	out, ok := new(big.Int).SetString(value, 10)
	if !ok || out == nil {
		return big.NewInt(0)
	}
	return out
}

func addDelegatorSlash(totals map[string]*delegatorSlashTotals, delegator string, amount, shares *big.Int) {
	if amount.Sign() == 0 && shares.Sign() == 0 {
		return
	}

	existing, ok := totals[delegator]
	if !ok {
		existing = &delegatorSlashTotals{amount: big.NewInt(0), shares: big.NewInt(0)}
		totals[delegator] = existing
	}

	existing.amount.Add(existing.amount, amount)
	existing.shares.Add(existing.shares, shares)
}
