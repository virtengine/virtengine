// Package keeper implements the delegation module keeper.
//
// VE-922: Delegation operations (delegate, undelegate, redelegate)
package keeper

import (
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

// Delegate delegates tokens from a delegator to a validator
func (k Keeper) Delegate(ctx sdk.Context, delegatorAddr, validatorAddr string, amount sdk.Coin) error {
	params := k.GetParams(ctx)

	// Validate minimum delegation amount
	if amount.Amount.Int64() < params.MinDelegationAmount {
		return types.ErrMinDelegationAmount.Wrapf(
			"minimum delegation is %d, got %s",
			params.MinDelegationAmount,
			amount.String(),
		)
	}

	// Check if delegator has max validators
	existingDelegations := k.GetDelegatorDelegations(ctx, delegatorAddr)
	_, existingDelegation := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !existingDelegation && int64(len(existingDelegations)) >= params.MaxValidatorsPerDelegator {
		return types.ErrMaxValidators.Wrapf(
			"max validators per delegator is %d",
			params.MaxValidatorsPerDelegator,
		)
	}

	// Get or create validator shares
	valShares := k.GetOrCreateValidatorShares(ctx, validatorAddr)

	// Calculate shares for this delegation
	amountStr := amount.Amount.String()
	newShares, err := valShares.CalculateSharesForAmount(amountStr)
	if err != nil {
		return types.ErrInvalidShares.Wrapf("failed to calculate shares: %v", err)
	}

	// Transfer tokens from delegator to module
	delegatorAccAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return types.ErrInvalidDelegator.Wrapf("invalid delegator address: %v", err)
	}

	if k.bankKeeper != nil {
		coins := sdk.NewCoins(amount)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAccAddr, types.ModuleName, coins); err != nil {
			return types.ErrInsufficientBalance.Wrapf("failed to transfer tokens: %v", err)
		}
	}

	// Update or create delegation
	del, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if found {
		// Add shares to existing delegation
		if err := del.AddShares(newShares, ctx.BlockTime()); err != nil {
			return err
		}
	} else {
		// Create new delegation
		del = *types.NewDelegation(
			delegatorAddr,
			validatorAddr,
			newShares,
			amountStr,
			ctx.BlockTime(),
			ctx.BlockHeight(),
		)
	}

	if err := k.SetDelegation(ctx, del); err != nil {
		return err
	}

	// Update validator shares
	if err := valShares.AddShares(newShares, amountStr, ctx.BlockTime()); err != nil {
		return err
	}
	if err := k.SetValidatorShares(ctx, valShares); err != nil {
		return err
	}

	// Emit delegate event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyShares, newShares),
		),
	)

	k.Logger(ctx).Info("delegation created",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"amount", amount.String(),
		"shares", newShares,
	)

	return nil
}

// Undelegate undelegates tokens from a validator (starts unbonding period)
func (k Keeper) Undelegate(ctx sdk.Context, delegatorAddr, validatorAddr string, amount sdk.Coin) (time.Time, error) {
	params := k.GetParams(ctx)

	// Get delegation
	del, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		return time.Time{}, types.ErrDelegationNotFound.Wrapf(
			"delegation from %s to %s not found",
			delegatorAddr,
			validatorAddr,
		)
	}

	// Get validator shares to calculate shares to unbond
	valShares, found := k.GetValidatorShares(ctx, validatorAddr)
	if !found {
		return time.Time{}, types.ErrValidatorNotFound.Wrapf("validator %s not found", validatorAddr)
	}

	// Calculate shares for the amount to undelegate
	amountStr := amount.Amount.String()
	sharesToUnbond, err := valShares.CalculateSharesForAmount(amountStr)
	if err != nil {
		return time.Time{}, types.ErrInvalidShares.Wrapf("failed to calculate shares: %v", err)
	}

	// Check if delegation has enough shares
	delegatorShares := del.GetSharesBigInt()
	unbondShares, ok := new(big.Int).SetString(sharesToUnbond, 10)
	if !ok {
		return time.Time{}, types.ErrInvalidShares.Wrapf("invalid unbond shares: %s", sharesToUnbond)
	}

	if delegatorShares.Cmp(unbondShares) < 0 {
		return time.Time{}, types.ErrInsufficientShares.Wrapf(
			"delegation has %s shares, need %s",
			del.Shares,
			sharesToUnbond,
		)
	}

	// Calculate completion time
	completionTime := ctx.BlockTime().Add(time.Duration(params.UnbondingPeriod) * time.Second)

	// Subtract shares from delegation
	if err := del.SubtractShares(sharesToUnbond, ctx.BlockTime()); err != nil {
		return time.Time{}, err
	}

	// If all shares are undelegated, delete the delegation
	if del.GetSharesBigInt().Sign() == 0 {
		k.DeleteDelegation(ctx, delegatorAddr, validatorAddr)
	} else {
		if err := k.SetDelegation(ctx, del); err != nil {
			return time.Time{}, err
		}
	}

	// Subtract shares from validator
	if err := valShares.SubtractShares(sharesToUnbond, amountStr, ctx.BlockTime()); err != nil {
		return time.Time{}, err
	}
	if err := k.SetValidatorShares(ctx, valShares); err != nil {
		return time.Time{}, err
	}

	// Create unbonding delegation
	unbondingID := k.generateUnbondingID(ctx, delegatorAddr, validatorAddr)
	ubd := types.NewUnbondingDelegation(
		unbondingID,
		delegatorAddr,
		validatorAddr,
		ctx.BlockHeight(),
		completionTime,
		ctx.BlockTime(),
		amountStr,
		sharesToUnbond,
	)

	if err := k.SetUnbondingDelegation(ctx, *ubd); err != nil {
		return time.Time{}, err
	}

	// Add to unbonding queue
	k.addToUnbondingQueue(ctx, completionTime, unbondingID)

	// Emit undelegate event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUndelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyShares, sharesToUnbond),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyUnbondingID, unbondingID),
		),
	)

	k.Logger(ctx).Info("undelegation initiated",
		"delegator", delegatorAddr,
		"validator", validatorAddr,
		"amount", amount.String(),
		"completion_time", completionTime.Format(time.RFC3339),
	)

	return completionTime, nil
}

// Redelegate redelegates tokens from one validator to another
func (k Keeper) Redelegate(ctx sdk.Context, delegatorAddr, srcValidator, dstValidator string, amount sdk.Coin) (time.Time, error) {
	params := k.GetParams(ctx)

	// Cannot redelegate to the same validator
	if srcValidator == dstValidator {
		return time.Time{}, types.ErrSelfRedelegation.Wrap("source and destination validators cannot be the same")
	}

	// Check for transitive redelegation (redelegating from a validator that received a redelegation)
	if k.HasRedelegation(ctx, delegatorAddr, srcValidator) {
		return time.Time{}, types.ErrTransitiveRedelegation.Wrap("transitive redelegation not allowed")
	}

	// Check max redelegations
	redelegationCount := k.CountDelegatorRedelegations(ctx, delegatorAddr)
	if int64(redelegationCount) >= params.MaxRedelegations {
		return time.Time{}, types.ErrMaxRedelegations.Wrapf(
			"max redelegations is %d",
			params.MaxRedelegations,
		)
	}

	// Get source delegation
	srcDel, found := k.GetDelegation(ctx, delegatorAddr, srcValidator)
	if !found {
		return time.Time{}, types.ErrDelegationNotFound.Wrapf(
			"delegation from %s to %s not found",
			delegatorAddr,
			srcValidator,
		)
	}

	// Get source validator shares
	srcValShares, found := k.GetValidatorShares(ctx, srcValidator)
	if !found {
		return time.Time{}, types.ErrValidatorNotFound.Wrapf("source validator %s not found", srcValidator)
	}

	// Calculate shares to redelegate
	amountStr := amount.Amount.String()
	sharesToRedelegate, err := srcValShares.CalculateSharesForAmount(amountStr)
	if err != nil {
		return time.Time{}, types.ErrInvalidShares.Wrapf("failed to calculate shares: %v", err)
	}

	// Check if source delegation has enough shares
	srcShares := srcDel.GetSharesBigInt()
	redShares, ok := new(big.Int).SetString(sharesToRedelegate, 10)
	if !ok {
		return time.Time{}, types.ErrInvalidShares.Wrapf("invalid redelegate shares: %s", sharesToRedelegate)
	}

	if srcShares.Cmp(redShares) < 0 {
		return time.Time{}, types.ErrInsufficientShares.Wrapf(
			"source delegation has %s shares, need %s",
			srcDel.Shares,
			sharesToRedelegate,
		)
	}

	// Get or create destination validator shares
	dstValShares := k.GetOrCreateValidatorShares(ctx, dstValidator)

	// Calculate new shares at destination validator
	newDstShares, err := dstValShares.CalculateSharesForAmount(amountStr)
	if err != nil {
		return time.Time{}, types.ErrInvalidShares.Wrapf("failed to calculate destination shares: %v", err)
	}

	// Calculate completion time
	completionTime := ctx.BlockTime().Add(time.Duration(params.UnbondingPeriod) * time.Second)

	// Subtract shares from source delegation
	if err := srcDel.SubtractShares(sharesToRedelegate, ctx.BlockTime()); err != nil {
		return time.Time{}, err
	}

	// If all shares are redelegated, delete the source delegation
	if srcDel.GetSharesBigInt().Sign() == 0 {
		k.DeleteDelegation(ctx, delegatorAddr, srcValidator)
	} else {
		if err := k.SetDelegation(ctx, srcDel); err != nil {
			return time.Time{}, err
		}
	}

	// Subtract shares from source validator
	if err := srcValShares.SubtractShares(sharesToRedelegate, amountStr, ctx.BlockTime()); err != nil {
		return time.Time{}, err
	}
	if err := k.SetValidatorShares(ctx, srcValShares); err != nil {
		return time.Time{}, err
	}

	// Add shares to destination delegation
	dstDel, found := k.GetDelegation(ctx, delegatorAddr, dstValidator)
	if found {
		if err := dstDel.AddShares(newDstShares, ctx.BlockTime()); err != nil {
			return time.Time{}, err
		}
	} else {
		dstDel = *types.NewDelegation(
			delegatorAddr,
			dstValidator,
			newDstShares,
			amountStr,
			ctx.BlockTime(),
			ctx.BlockHeight(),
		)
	}

	if err := k.SetDelegation(ctx, dstDel); err != nil {
		return time.Time{}, err
	}

	// Add shares to destination validator
	if err := dstValShares.AddShares(newDstShares, amountStr, ctx.BlockTime()); err != nil {
		return time.Time{}, err
	}
	if err := k.SetValidatorShares(ctx, dstValShares); err != nil {
		return time.Time{}, err
	}

	// Create redelegation record
	redelegationID := k.generateRedelegationID(ctx, delegatorAddr, srcValidator, dstValidator)
	red := types.NewRedelegation(
		redelegationID,
		delegatorAddr,
		srcValidator,
		dstValidator,
		ctx.BlockHeight(),
		completionTime,
		ctx.BlockTime(),
		amountStr,
		newDstShares,
	)

	if err := k.SetRedelegation(ctx, *red); err != nil {
		return time.Time{}, err
	}

	// Add to redelegation queue
	k.addToRedelegationQueue(ctx, completionTime, redelegationID)

	// Emit redelegate event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr),
			sdk.NewAttribute(types.AttributeKeySrcValidator, srcValidator),
			sdk.NewAttribute(types.AttributeKeyDstValidator, dstValidator),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyShares, newDstShares),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyRedelegationID, redelegationID),
		),
	)

	k.Logger(ctx).Info("redelegation initiated",
		"delegator", delegatorAddr,
		"src_validator", srcValidator,
		"dst_validator", dstValidator,
		"amount", amount.String(),
		"completion_time", completionTime.Format(time.RFC3339),
	)

	return completionTime, nil
}

// CompleteUnbonding completes mature unbonding delegations and returns tokens
func (k Keeper) CompleteUnbonding(ctx sdk.Context, unbondingID string) error {
	ubd, found := k.GetUnbondingDelegation(ctx, unbondingID)
	if !found {
		return types.ErrUnbondingNotFound.Wrapf("unbonding %s not found", unbondingID)
	}

	now := ctx.BlockTime()
	var completedAmount *big.Int = big.NewInt(0)
	var remainingEntries []types.UnbondingDelegationEntry

	for _, entry := range ubd.Entries {
		if !entry.CompletionTime.After(now) {
			// Entry is mature, add to completed amount
			balance, ok := new(big.Int).SetString(entry.Balance, 10)
			if ok {
				completedAmount.Add(completedAmount, balance)
			}
		} else {
			// Entry not yet mature
			remainingEntries = append(remainingEntries, entry)
		}
	}

	if completedAmount.Sign() > 0 {
		// Transfer tokens back to delegator
		delegatorAccAddr, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
		if err != nil {
			return types.ErrInvalidDelegator.Wrapf("invalid delegator address: %v", err)
		}

		if k.bankKeeper != nil {
			params := k.GetParams(ctx)
			coins := sdk.NewCoins(sdk.NewCoin(params.StakeDenom, math.NewIntFromBigInt(completedAmount)))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAccAddr, coins); err != nil {
				return fmt.Errorf("failed to return tokens: %w", err)
			}
		}

		// Emit complete unbonding event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(types.AttributeKeyDelegator, ubd.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeyValidator, ubd.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyAmount, completedAmount.String()),
				sdk.NewAttribute(types.AttributeKeyUnbondingID, unbondingID),
			),
		)

		k.Logger(ctx).Info("unbonding completed",
			"delegator", ubd.DelegatorAddress,
			"validator", ubd.ValidatorAddress,
			"amount", completedAmount.String(),
		)
	}

	// Update or delete unbonding delegation
	if len(remainingEntries) == 0 {
		k.DeleteUnbondingDelegation(ctx, unbondingID)
	} else {
		ubd.Entries = remainingEntries
		if err := k.SetUnbondingDelegation(ctx, ubd); err != nil {
			return err
		}
	}

	return nil
}

// CompleteRedelegation completes a mature redelegation
func (k Keeper) CompleteRedelegation(ctx sdk.Context, redelegationID string) error {
	red, found := k.GetRedelegation(ctx, redelegationID)
	if !found {
		return types.ErrRedelegationNotFound.Wrapf("redelegation %s not found", redelegationID)
	}

	now := ctx.BlockTime()
	var remainingEntries []types.RedelegationEntry

	for _, entry := range red.Entries {
		if !entry.CompletionTime.After(now) {
			// Entry is mature
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCompleteRedelegation,
					sdk.NewAttribute(types.AttributeKeyDelegator, red.DelegatorAddress),
					sdk.NewAttribute(types.AttributeKeySrcValidator, red.ValidatorSrcAddress),
					sdk.NewAttribute(types.AttributeKeyDstValidator, red.ValidatorDstAddress),
					sdk.NewAttribute(types.AttributeKeyAmount, entry.InitialBalance),
					sdk.NewAttribute(types.AttributeKeyRedelegationID, redelegationID),
				),
			)

			k.Logger(ctx).Info("redelegation completed",
				"delegator", red.DelegatorAddress,
				"src_validator", red.ValidatorSrcAddress,
				"dst_validator", red.ValidatorDstAddress,
				"amount", entry.InitialBalance,
			)
		} else {
			// Entry not yet mature
			remainingEntries = append(remainingEntries, entry)
		}
	}

	// Update or delete redelegation
	if len(remainingEntries) == 0 {
		k.DeleteRedelegation(ctx, redelegationID)
	} else {
		red.Entries = remainingEntries
		if err := k.SetRedelegation(ctx, red); err != nil {
			return err
		}
	}

	return nil
}
