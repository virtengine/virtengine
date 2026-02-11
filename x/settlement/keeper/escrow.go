package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	escrowmodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// CreateEscrow creates a new escrow account for an order
func (k Keeper) CreateEscrow(
	ctx sdk.Context,
	orderID string,
	depositor sdk.AccAddress,
	amount sdk.Coins,
	expiresIn time.Duration,
	conditions []types.ReleaseCondition,
) (string, error) {
	// Check if escrow already exists for this order
	if _, found := k.GetEscrowByOrder(ctx, orderID); found {
		return "", types.ErrEscrowExists.Wrapf("escrow already exists for order %s", orderID)
	}

	// Validate expiration
	params := k.GetParams(ctx)
	expiresInSeconds := uint64(expiresIn.Seconds())
	if expiresInSeconds < params.MinEscrowDuration {
		return "", types.ErrInvalidEscrow.Wrapf("expires_in must be at least %d seconds", params.MinEscrowDuration)
	}
	if expiresInSeconds > params.MaxEscrowDuration {
		return "", types.ErrInvalidEscrow.Wrapf("expires_in cannot exceed %d seconds", params.MaxEscrowDuration)
	}

	// Generate escrow ID
	seq := k.incrementEscrowSequence(ctx)
	escrowID := generateID("escrow", seq)

	// Calculate expiration time
	expiresAt := ctx.BlockTime().Add(expiresIn)

	// Create escrow account
	escrow := types.NewEscrowAccount(
		escrowID,
		orderID,
		depositor.String(),
		amount,
		expiresAt,
		conditions,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	accountID, err := escrowAccountIDFromInputs(depositor.String(), escrowID)
	if err != nil {
		return "", err
	}

	deposits := buildEscrowDepositors(ctx, depositor, amount)
	if err := k.escrowKeeper.AccountCreate(ctx, accountID, depositor, deposits); err != nil {
		return "", types.ErrInsufficientFunds.Wrap(err.Error())
	}

	// Save escrow
	if err := k.SetEscrow(ctx, *escrow); err != nil {
		return "", err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventEscrowCreated{
		EscrowID:    escrowID,
		OrderID:     orderID,
		Depositor:   depositor.String(),
		Amount:      amount.String(),
		ExpiresAt:   expiresAt.Unix(),
		BlockHeight: ctx.BlockHeight(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit escrow created event", "error", err)
	}

	k.Logger(ctx).Info("escrow created",
		"escrow_id", escrowID,
		"order_id", orderID,
		"depositor", depositor.String(),
		"amount", amount.String(),
	)

	return escrowID, nil
}

// ActivateEscrow activates an escrow when a lease is created
func (k Keeper) ActivateEscrow(ctx sdk.Context, escrowID, leaseID string, recipient sdk.AccAddress) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	// Check if expired
	if escrow.CheckExpiry(ctx.BlockTime()) {
		if err := k.SetEscrow(ctx, escrow); err != nil {
			return err
		}
		return types.ErrEscrowExpired.Wrapf("escrow %s has expired", escrowID)
	}

	oldState := escrow.State

	// Activate the escrow
	if err := escrow.Activate(recipient.String(), ctx.BlockTime()); err != nil {
		return err
	}

	escrow.LeaseID = leaseID

	// Update escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Update state index
	k.updateEscrowState(ctx, escrow, oldState)

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventEscrowActivated{
		EscrowID:    escrowID,
		OrderID:     escrow.OrderID,
		LeaseID:     leaseID,
		Recipient:   recipient.String(),
		ActivatedAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit escrow activated event", "error", err)
	}

	k.Logger(ctx).Info("escrow activated",
		"escrow_id", escrowID,
		"lease_id", leaseID,
		"recipient", recipient.String(),
	)

	return nil
}

// ReleaseEscrow releases escrow funds to the recipient
func (k Keeper) ReleaseEscrow(ctx sdk.Context, escrowID string, reason string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	// Check if expired
	if escrow.CheckExpiry(ctx.BlockTime()) {
		if err := k.SetEscrow(ctx, escrow); err != nil {
			return err
		}
		return types.ErrEscrowExpired.Wrapf("escrow %s has expired", escrowID)
	}

	// Check if active
	if escrow.State != types.EscrowStateActive && escrow.State != types.EscrowStateDisputed {
		return types.ErrEscrowNotActive.Wrapf("escrow %s is not active", escrowID)
	}

	// Check if conditions are met (unless disputed and being resolved)
	if escrow.State != types.EscrowStateDisputed && !escrow.AllConditionsSatisfied() {
		return types.ErrConditionsNotMet.Wrapf("release conditions not met for escrow %s", escrowID)
	}

	// Get recipient
	if escrow.Recipient == "" {
		return types.ErrInvalidEscrow.Wrap("escrow has no recipient")
	}
	recipient, err := sdk.AccAddressFromBech32(escrow.Recipient)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid recipient address")
	}

	// Calculate release amount (remaining balance)
	releaseAmount := escrow.Balance

	oldState := escrow.State

	// Release the escrow
	if err := escrow.Release(ctx.BlockTime(), reason); err != nil {
		return err
	}

	if !releaseAmount.IsZero() {
		if err := escrow.DeductBalance(releaseAmount); err != nil {
			return err
		}
	}

	// Transfer funds to recipient
	if !releaseAmount.IsZero() {
		if err := k.transferEscrowFundsToAccount(ctx, escrow, recipient, releaseAmount, true); err != nil {
			return err
		}
	} else {
		if err := k.closeEscrowAccount(ctx, escrow); err != nil {
			return err
		}
	}

	// Update escrow
	escrow.Balance = sdk.NewCoins()
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	if !releaseAmount.IsZero() {
		if _, err := k.recordEscrowDisbursement(ctx, escrow, releaseAmount, recipient.String(), types.SettlementTypeFinal, true); err != nil {
			return err
		}
	}

	// Update state index
	k.updateEscrowState(ctx, escrow, oldState)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventEscrowReleased{
		EscrowID:   escrowID,
		OrderID:    escrow.OrderID,
		Recipient:  escrow.Recipient,
		Amount:     releaseAmount.String(),
		Reason:     reason,
		ReleasedAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit escrow released event", "error", err)
	}

	k.Logger(ctx).Info("escrow released",
		"escrow_id", escrowID,
		"recipient", escrow.Recipient,
		"amount", releaseAmount.String(),
	)

	return nil
}

// RefundEscrow refunds escrow funds to the depositor
func (k Keeper) RefundEscrow(ctx sdk.Context, escrowID string, reason string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	// Get depositor
	depositor, err := sdk.AccAddressFromBech32(escrow.Depositor)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid depositor address")
	}

	// Calculate refund amount (remaining balance)
	refundAmount := escrow.Balance

	oldState := escrow.State

	// Refund the escrow
	if err := escrow.Refund(ctx.BlockTime(), reason); err != nil {
		return err
	}

	if !refundAmount.IsZero() {
		if err := escrow.DeductBalance(refundAmount); err != nil {
			return err
		}
	}

	// Transfer funds back to depositor
	if !refundAmount.IsZero() {
		if err := k.transferEscrowFundsToAccount(ctx, escrow, depositor, refundAmount, true); err != nil {
			return err
		}
	} else {
		if err := k.closeEscrowAccount(ctx, escrow); err != nil {
			return err
		}
	}

	// Update escrow
	escrow.Balance = sdk.NewCoins()
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	if !refundAmount.IsZero() {
		if _, err := k.recordEscrowDisbursement(ctx, escrow, refundAmount, depositor.String(), types.SettlementTypeRefund, true); err != nil {
			return err
		}
	}

	// Update state index
	k.updateEscrowState(ctx, escrow, oldState)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventEscrowRefunded{
		EscrowID:   escrowID,
		OrderID:    escrow.OrderID,
		Depositor:  escrow.Depositor,
		Amount:     refundAmount.String(),
		Reason:     reason,
		RefundedAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit escrow refunded event", "error", err)
	}

	k.Logger(ctx).Info("escrow refunded",
		"escrow_id", escrowID,
		"depositor", escrow.Depositor,
		"amount", refundAmount.String(),
		"reason", reason,
	)

	return nil
}

func (k Keeper) recordEscrowDisbursement(
	ctx sdk.Context,
	escrow types.EscrowAccount,
	amount sdk.Coins,
	recipient string,
	settlementType types.SettlementType,
	isFinal bool,
) (*types.SettlementRecord, error) {
	if amount.IsZero() {
		return nil, nil
	}

	seq := k.incrementSettlementSequence(ctx)
	settlementID := generateIDWithTimestamp("settle", seq, ctx.BlockTime().Unix())

	periodStart := escrow.CreatedAt
	if escrow.ActivatedAt != nil {
		periodStart = *escrow.ActivatedAt
	}

	settlement := types.NewSettlementRecord(
		settlementID,
		escrow.EscrowID,
		escrow.OrderID,
		escrow.LeaseID,
		recipient,
		escrow.Depositor,
		amount,
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		nil,
		0,
		periodStart,
		ctx.BlockTime(),
		settlementType,
		isFinal,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	if err := k.SetSettlement(ctx, *settlement); err != nil {
		return nil, err
	}

	return settlement, nil
}

// DisputeEscrow marks an escrow as disputed
func (k Keeper) DisputeEscrow(ctx sdk.Context, escrowID string, reason string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	// Check if expired
	if escrow.CheckExpiry(ctx.BlockTime()) {
		if err := k.SetEscrow(ctx, escrow); err != nil {
			return err
		}
		return types.ErrEscrowExpired.Wrapf("escrow %s has expired", escrowID)
	}

	oldState := escrow.State

	// Dispute the escrow
	if err := escrow.Dispute(ctx.BlockTime(), reason); err != nil {
		return err
	}

	// Update escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Update state index
	k.updateEscrowState(ctx, escrow, oldState)

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventEscrowDisputed{
		EscrowID:   escrowID,
		OrderID:    escrow.OrderID,
		Reason:     reason,
		DisputedAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit escrow disputed event", "error", err)
	}

	k.Logger(ctx).Info("escrow disputed",
		"escrow_id", escrowID,
		"reason", reason,
	)

	return nil
}

// ProcessExpiredEscrows processes all expired escrows
func (k Keeper) ProcessExpiredEscrows(ctx sdk.Context) error {
	// Check pending and active escrows for expiry
	states := []types.EscrowState{types.EscrowStatePending, types.EscrowStateActive}

	for _, state := range states {
		k.WithEscrowsByState(ctx, state, func(escrow types.EscrowAccount) bool {
			if escrow.CheckExpiry(ctx.BlockTime()) {
				oldState := escrow.State

				// Refund remaining balance to depositor
				if !escrow.Balance.IsZero() {
					depositor, err := sdk.AccAddressFromBech32(escrow.Depositor)
					if err == nil {
						if err := k.transferEscrowFundsToAccount(ctx, escrow, depositor, escrow.Balance, true); err != nil {
							k.Logger(ctx).Error("failed to refund expired escrow", "error", err, "escrow_id", escrow.EscrowID)
						}
					}
				} else if err := k.closeEscrowAccount(ctx, escrow); err != nil {
					k.Logger(ctx).Error("failed to close expired escrow account", "error", err, "escrow_id", escrow.EscrowID)
				}

				escrow.Balance = sdk.NewCoins()
				if err := k.SetEscrow(ctx, escrow); err != nil {
					k.Logger(ctx).Error("failed to save expired escrow", "error", err, "escrow_id", escrow.EscrowID)
				}

				k.updateEscrowState(ctx, escrow, oldState)

				// Emit event
				_ = ctx.EventManager().EmitTypedEvent(&types.EventEscrowExpired{
					EscrowID:  escrow.EscrowID,
					OrderID:   escrow.OrderID,
					Balance:   escrow.Balance.String(),
					ExpiredAt: ctx.BlockTime().Unix(),
				})

				k.Logger(ctx).Info("escrow expired",
					"escrow_id", escrow.EscrowID,
				)
			}
			return false
		})
	}
	return nil
}

// SatisfyTimelockConditions satisfies timelock conditions (interface method - iterates all escrows)
func (k Keeper) SatisfyTimelockConditions(ctx sdk.Context) error {
	k.WithEscrowsByState(ctx, types.EscrowStatePending, func(escrow types.EscrowAccount) bool {
		_ = k.satisfyTimelockForEscrow(ctx, escrow.EscrowID)
		return false
	})
	k.WithEscrowsByState(ctx, types.EscrowStateActive, func(escrow types.EscrowAccount) bool {
		_ = k.satisfyTimelockForEscrow(ctx, escrow.EscrowID)
		return false
	})
	return nil
}

// satisfyTimelockForEscrow checks and satisfies timelock conditions for a specific escrow
func (k Keeper) satisfyTimelockForEscrow(ctx sdk.Context, escrowID string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	modified := false
	now := ctx.BlockTime()

	for i := range escrow.Conditions {
		cond := &escrow.Conditions[i]
		if cond.Type == types.ConditionTypeTimelock && !cond.Satisfied {
			if cond.UnlockAfter != nil && now.After(*cond.UnlockAfter) {
				cond.Satisfied = true
				cond.SatisfiedAt = &now
				modified = true
			}
		}
	}

	if modified {
		return k.SetEscrow(ctx, escrow)
	}

	return nil
}

// SatisfyUsageCondition checks and satisfies usage threshold conditions
func (k Keeper) SatisfyUsageCondition(ctx sdk.Context, escrowID string, usageUnits uint64) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("escrow %s not found", escrowID)
	}

	modified := false
	now := ctx.BlockTime()

	for i := range escrow.Conditions {
		cond := &escrow.Conditions[i]
		if cond.Type == types.ConditionTypeUsageThreshold && !cond.Satisfied {
			if usageUnits >= cond.MinUsageUnits {
				cond.Satisfied = true
				cond.SatisfiedAt = &now
				modified = true
			}
		}
	}

	if modified {
		return k.SetEscrow(ctx, escrow)
	}

	return nil
}

func escrowAccountIDFromInputs(depositor string, escrowID string) (escrowid.Account, error) {
	if depositor == "" {
		return escrowid.Account{}, types.ErrInvalidEscrow.Wrap("depositor is required")
	}

	if _, err := sdk.AccAddressFromBech32(depositor); err != nil {
		return escrowid.Account{}, types.ErrInvalidEscrow.Wrap("invalid depositor address")
	}

	seq, err := parseEscrowSequence(escrowID)
	if err != nil {
		return escrowid.Account{}, err
	}

	account := escrowid.Account{
		Scope: escrowid.ScopeDeployment,
		XID:   fmt.Sprintf("%s/%d", depositor, seq),
	}
	if err := account.ValidateBasic(); err != nil {
		return escrowid.Account{}, types.ErrInvalidEscrow.Wrap(err.Error())
	}

	return account, nil
}

func escrowAccountIDFromEscrow(escrow types.EscrowAccount) (escrowid.Account, error) {
	return escrowAccountIDFromInputs(escrow.Depositor, escrow.EscrowID)
}

func parseEscrowSequence(escrowID string) (uint64, error) {
	parts := strings.Split(escrowID, "-")
	if len(parts) < 2 {
		return 0, types.ErrInvalidEscrow.Wrap("escrow_id missing sequence")
	}

	seq, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if err != nil {
		return 0, types.ErrInvalidEscrow.Wrap("escrow_id has invalid sequence")
	}

	return seq, nil
}

func buildEscrowDepositors(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coins) []etypes.Depositor {
	deposits := make([]etypes.Depositor, 0, len(amount))
	for _, coin := range amount {
		deposits = append(deposits, etypes.Depositor{
			Owner:   depositor.String(),
			Height:  ctx.BlockHeight(),
			Source:  deposit.SourceBalance,
			Balance: sdk.NewDecCoinFromCoin(coin),
			Direct:  true,
		})
	}
	return deposits
}

func (k Keeper) transferEscrowFundsToAccount(ctx sdk.Context, escrow types.EscrowAccount, recipient sdk.AccAddress, amount sdk.Coins, closeAccount bool) error {
	if amount.IsZero() {
		return nil
	}

	accountID, err := escrowAccountIDFromEscrow(escrow)
	if err != nil {
		return err
	}

	account, err := k.escrowKeeper.GetAccount(ctx, accountID)
	if err != nil {
		return types.ErrEscrowNotFound.Wrap(err.Error())
	}

	updated := account
	if err := deductEscrowAccountBalance(&updated, amount); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, escrowmodule.ModuleName, recipient, amount); err != nil {
		return types.ErrInsufficientFunds.Wrap(err.Error())
	}

	updated.State.SettledAt = ctx.BlockHeight()
	if closeAccount {
		updated.State.State = etypes.StateClosed
	}

	if err := updated.ValidateBasic(); err != nil {
		return types.ErrInvalidEscrow.Wrap(err.Error())
	}

	return k.escrowKeeper.SaveAccount(ctx, updated)
}

func (k Keeper) transferEscrowFundsToModule(ctx sdk.Context, escrow types.EscrowAccount, recipientModule string, amount sdk.Coins, closeAccount bool) error {
	if amount.IsZero() {
		return nil
	}

	accountID, err := escrowAccountIDFromEscrow(escrow)
	if err != nil {
		return err
	}

	account, err := k.escrowKeeper.GetAccount(ctx, accountID)
	if err != nil {
		return types.ErrEscrowNotFound.Wrap(err.Error())
	}

	updated := account
	if err := deductEscrowAccountBalance(&updated, amount); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, escrowmodule.ModuleName, recipientModule, amount); err != nil {
		return types.ErrInsufficientFunds.Wrap(err.Error())
	}

	updated.State.SettledAt = ctx.BlockHeight()
	if closeAccount {
		updated.State.State = etypes.StateClosed
	}

	if err := updated.ValidateBasic(); err != nil {
		return types.ErrInvalidEscrow.Wrap(err.Error())
	}

	return k.escrowKeeper.SaveAccount(ctx, updated)
}

func (k Keeper) closeEscrowAccount(ctx sdk.Context, escrow types.EscrowAccount) error {
	accountID, err := escrowAccountIDFromEscrow(escrow)
	if err != nil {
		return err
	}

	account, err := k.escrowKeeper.GetAccount(ctx, accountID)
	if err != nil {
		return types.ErrEscrowNotFound.Wrap(err.Error())
	}

	account.State.State = etypes.StateClosed
	account.State.SettledAt = ctx.BlockHeight()

	if err := account.ValidateBasic(); err != nil {
		return types.ErrInvalidEscrow.Wrap(err.Error())
	}

	return k.escrowKeeper.SaveAccount(ctx, account)
}

func deductEscrowAccountBalance(account *etypes.Account, amount sdk.Coins) error {
	for _, coin := range amount {
		if coin.Amount.IsZero() {
			continue
		}

		var funds *etypes.Balance
		var transferred *sdk.DecCoin

		for i := range account.State.Funds {
			if account.State.Funds[i].Denom == coin.Denom {
				funds = &account.State.Funds[i]
				break
			}
		}

		for i := range account.State.Transferred {
			if account.State.Transferred[i].Denom == coin.Denom {
				transferred = &account.State.Transferred[i]
				break
			}
		}

		if funds == nil || transferred == nil {
			return types.ErrInvalidEscrow.Wrapf("unknown escrow denom %s", coin.Denom)
		}

		remaining := sdkmath.LegacyZeroDec()
		remaining.AddMut(sdkmath.LegacyNewDecFromInt(coin.Amount))
		withdrew := sdkmath.LegacyZeroDec()
		idx := 0

		for i, d := range account.State.Deposits {
			toWithdraw := sdkmath.LegacyZeroDec()
			if d.Balance.Amount.LT(remaining) {
				toWithdraw.AddMut(d.Balance.Amount)
			} else {
				toWithdraw.AddMut(remaining)
			}

			account.State.Deposits[i].Balance.Amount.SubMut(toWithdraw)
			if account.State.Deposits[i].Balance.IsZero() {
				idx++
			}

			remaining.SubMut(toWithdraw)
			withdrew.AddMut(toWithdraw)
			transferred.Amount.AddMut(toWithdraw)

			if remaining.IsZero() {
				break
			}
		}

		if idx > 0 {
			account.State.Deposits = account.State.Deposits[idx:]
		}

		funds.Amount.SubMut(withdrew)
		if !remaining.IsZero() {
			return types.ErrInsufficientFunds.Wrap("insufficient escrow balance")
		}
	}

	return nil
}
