package keeper

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// SettleOrder settles an order based on usage records
func (k Keeper) SettleOrder(ctx sdk.Context, orderID string, usageRecordIDs []string, isFinal bool) (*types.SettlementRecord, error) {
	cacheCtx, write := ctx.CacheContext()
	settlement, err := k.settleOrder(cacheCtx, orderID, usageRecordIDs, isFinal)
	if err != nil {
		return nil, err
	}
	write()
	return settlement, nil
}

func (k Keeper) settleOrder(ctx sdk.Context, orderID string, usageRecordIDs []string, isFinal bool) (*types.SettlementRecord, error) {
	// Get the escrow for this order
	escrow, found := k.GetEscrowByOrder(ctx, orderID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", orderID)
	}

	// Check escrow is active
	if escrow.State != types.EscrowStateActive {
		return nil, types.ErrEscrowNotActive.Wrapf("escrow for order %s is not active", orderID)
	}

	// Check if expired
	if escrow.CheckExpiry(ctx.BlockTime()) {
		if err := k.SetEscrow(ctx, escrow); err != nil {
			return nil, err
		}
		return nil, types.ErrEscrowExpired.Wrapf("escrow for order %s has expired", orderID)
	}

	// Get usage records to settle
	var usageRecords []types.UsageRecord
	if len(usageRecordIDs) > 0 {
		// Use specified usage records
		seenUsageIDs := make(map[string]struct{}, len(usageRecordIDs))
		for _, id := range usageRecordIDs {
			if _, duplicate := seenUsageIDs[id]; duplicate {
				return nil, types.ErrUsageReplayConflict.Wrapf("duplicate usage record %s in settlement request", id)
			}
			seenUsageIDs[id] = struct{}{}
			usage, found := k.GetUsageRecord(ctx, id)
			if !found {
				return nil, types.ErrUsageRecordNotFound.Wrapf("usage record %s not found", id)
			}
			if usage.Settled {
				return nil, types.ErrUsageAlreadySettled.Wrapf("usage record %s already settled", id)
			}
			if usage.OrderID != orderID {
				return nil, types.ErrInvalidUsageRecord.Wrapf("usage record %s belongs to different order", id)
			}
			if k.IsUsageAuthenticationActive(ctx) && !usage.IsAuthenticated() {
				return nil, types.ErrUsageAuthenticationRequired.Wrapf("usage record %s is legacy or unverified", id)
			}
			usageRecords = append(usageRecords, usage)
		}
	} else {
		// Get all unsettled usage records
		usageRecords = k.GetUnsettledUsageRecords(ctx, orderID)
	}
	if k.IsUsageAuthenticationActive(ctx) && len(usageRecords) == 0 {
		return nil, types.ErrUsageAuthenticationRequired.Wrap("at least one authenticated usage record is required for settlement")
	}

	// Calculate total cost from usage records
	totalCost := sdk.NewCoins()
	var totalUsageUnits uint64
	var periodStart, periodEnd time.Time

	for i, usage := range usageRecords {
		totalCost = totalCost.Add(usage.TotalCost...)
		if usage.UsageUnits > math.MaxUint64-totalUsageUnits {
			return nil, types.ErrInvalidSettlement.Wrap("total usage units overflow")
		}
		totalUsageUnits += usage.UsageUnits

		if i == 0 {
			periodStart = usage.PeriodStart
			periodEnd = usage.PeriodEnd
		} else {
			if usage.PeriodStart.Before(periodStart) {
				periodStart = usage.PeriodStart
			}
			if usage.PeriodEnd.After(periodEnd) {
				periodEnd = usage.PeriodEnd
			}
		}
	}

	// If no usage and not final, nothing to settle
	if totalCost.IsZero() && !isFinal {
		return nil, types.ErrInvalidSettlement.Wrap("no usage to settle")
	}

	// Determine settlement amount
	var settlementAmount sdk.Coins
	if isFinal {
		// Final settlement uses remaining balance
		settlementAmount = escrow.Balance
	} else {
		// Regular settlement uses calculated cost
		settlementAmount = totalCost
		// Cap at escrow balance
		for _, coin := range settlementAmount {
			balance := escrow.Balance.AmountOf(coin.Denom)
			if coin.Amount.GT(balance) {
				settlementAmount = sdk.NewCoins(sdk.NewCoin(coin.Denom, balance))
			}
		}
	}

	if settlementAmount.IsZero() {
		return nil, types.ErrInvalidSettlement.Wrap("no funds available for settlement")
	}

	// Calculate fee splits
	platformFeeRate := k.getPlatformFeeRate(ctx)
	validatorFeeRate := k.getValidatorFeeRate(ctx)

	platformFee := sdk.NewCoins()
	validatorFee := sdk.NewCoins()
	providerShare := sdk.NewCoins()

	for _, coin := range settlementAmount {
		// Platform fee
		pFee := sdkmath.LegacyNewDecFromInt(coin.Amount).Mul(platformFeeRate).TruncateInt()
		platformFee = platformFee.Add(sdk.NewCoin(coin.Denom, pFee))

		// Validator fee
		vFee := sdkmath.LegacyNewDecFromInt(coin.Amount).Mul(validatorFeeRate).TruncateInt()
		validatorFee = validatorFee.Add(sdk.NewCoin(coin.Denom, vFee))

		// Provider gets the rest
		pShare := coin.Amount.Sub(pFee).Sub(vFee)
		providerShare = providerShare.Add(sdk.NewCoin(coin.Denom, pShare))
	}

	// Get provider and customer addresses
	provider, err := sdk.AccAddressFromBech32(escrow.Recipient)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid provider address")
	}

	if err := k.transferEscrowFundsToModule(ctx, escrow, types.ModuleAccountName, settlementAmount, false); err != nil {
		return nil, err
	}

	// Transfer provider share when billing pipeline is disabled (legacy path).
	if !k.billingEnabled() && !providerShare.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, provider, providerShare); err != nil {
			return nil, types.ErrInsufficientFunds.Wrap(err.Error())
		}
	}

	// Platform and validator fees stay in module account for later distribution
	// In production, you would transfer to appropriate fee collectors

	// Generate settlement ID
	seq := k.incrementSettlementSequence(ctx)
	settlementID := generateIDWithTimestamp("settle", seq, ctx.BlockTime().Unix())

	// Determine settlement type
	settlementType := types.SettlementTypeUsageBased
	if isFinal {
		settlementType = types.SettlementTypeFinal
	} else if len(usageRecords) == 0 {
		settlementType = types.SettlementTypePeriodic
	}

	// Collect usage record IDs
	settledUsageIDs := make([]string, len(usageRecords))
	for i, usage := range usageRecords {
		settledUsageIDs[i] = usage.UsageID
	}

	// Create settlement record
	settlement := types.NewSettlementRecord(
		settlementID,
		escrow.EscrowID,
		orderID,
		escrow.LeaseID,
		escrow.Recipient,
		escrow.Depositor,
		settlementAmount,
		providerShare,
		platformFee,
		validatorFee,
		settledUsageIDs,
		totalUsageUnits,
		periodStart,
		periodEnd,
		settlementType,
		isFinal,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// Save settlement
	if err := k.SetSettlement(ctx, *settlement); err != nil {
		return nil, err
	}

	// Update escrow balance
	if err := escrow.DeductBalance(settlementAmount); err != nil {
		return nil, err
	}

	// Mark usage records as settled
	for _, usage := range usageRecords {
		usage.MarkSettled(settlementID)
		if err := k.SetUsageRecord(ctx, usage); err != nil {
			k.Logger(ctx).Error("failed to mark usage as settled", "error", err, "usage_id", usage.UsageID)
		}
	}

	// If final, release remaining balance to provider and close escrow
	if isFinal && escrow.Balance.IsZero() {
		oldState := escrow.State
		if err := escrow.Release(ctx.BlockTime(), "final settlement"); err != nil {
			return nil, err
		}
		k.updateEscrowState(ctx, escrow, oldState)
		if err := k.closeEscrowAccount(ctx, escrow); err != nil {
			return nil, err
		}
	}

	// Save updated escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return nil, err
	}

	if k.billingEnabled() {
		invoiceID, err := k.generateInvoiceForSettlement(ctx, settlement, usageRecords)
		if err != nil {
			return nil, err
		}

		if err := k.updateBillingUsageRecordsForSettlement(ctx, settlement, usageRecords, invoiceID); err != nil {
			return nil, err
		}

		if pref, ok := k.GetFiatPayoutPreference(ctx, settlement.Provider); ok && pref.Enabled {
			if _, err := k.createConversionFromPreference(ctx, *settlement, invoiceID, pref); err != nil {
				return nil, err
			}
		}

		if _, err := k.ExecutePayout(ctx, invoiceID, settlement.SettlementID); err != nil {
			return nil, err
		}
	}

	if len(usageRecords) > 0 {
		if _, err := k.DistributeUsageRewardsForSettlement(ctx, settlement.SettlementID, usageRecords); err != nil {
			k.Logger(ctx).Debug("usage reward distribution skipped", "error", err, "settlement_id", settlement.SettlementID)
		}
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventOrderSettled{
		SettlementID:   settlementID,
		OrderID:        orderID,
		EscrowID:       escrow.EscrowID,
		Provider:       escrow.Recipient,
		Customer:       escrow.Depositor,
		TotalAmount:    settlementAmount.String(),
		ProviderShare:  providerShare.String(),
		PlatformFee:    platformFee.String(),
		SettlementType: string(settlementType),
		IsFinal:        isFinal,
		SettledAt:      ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit order settled event", "error", err)
	}

	k.Logger(ctx).Info("order settled",
		"settlement_id", settlementID,
		"order_id", orderID,
		"total_amount", settlementAmount.String(),
		"provider_share", providerShare.String(),
		"is_final", isFinal,
	)

	return settlement, nil
}

// RecordUsage records a usage record from a provider
func (k Keeper) RecordUsage(ctx sdk.Context, record *types.UsageRecord) error {
	cacheCtx, write := ctx.CacheContext()
	if err := k.recordUsage(cacheCtx, record); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) recordUsage(ctx sdk.Context, record *types.UsageRecord) error {
	if record == nil {
		return types.ErrInvalidUsageRecord.Wrap("usage record cannot be nil")
	}

	// Calculate total cost exactly from signed unit price and units. A caller-
	// supplied TotalCost is never an independent financial input.
	if record.UnitPrice.Denom != "" && !record.UnitPrice.IsNegative() {
		quantity := sdkmath.LegacyNewDecFromInt(sdkmath.NewIntFromUint64(record.UsageUnits))
		total := record.UnitPrice.Amount.Mul(quantity).TruncateInt()
		expected := sdk.NewCoins()
		if total.IsPositive() {
			expected = sdk.NewCoins(sdk.NewCoin(record.UnitPrice.Denom, total))
		}
		if record.SignatureVersion == types.SignatureVersionV1 && !record.TotalCost.IsZero() && !record.TotalCost.Equal(expected) {
			return types.ErrInvalidUsageRecord.Wrap("total_cost does not match signed units and unit price")
		}
		record.TotalCost = expected
	}

	authenticationActive := k.IsUsageAuthenticationActive(ctx)
	var streamID, usageDigest []byte
	if record.SignatureVersion == types.SignatureVersionV1 {
		digest, stream, duplicateID, duplicate, err := k.verifyAuthenticatedUsage(ctx, record)
		if err != nil {
			return err
		}
		if duplicate {
			existing, found := k.GetUsageRecord(ctx, duplicateID)
			if !found {
				return types.ErrUsageReplayConflict.Wrap("replay index references missing usage record")
			}
			*record = existing
			record.ExactDuplicate = true
			return nil
		}
		usageDigest = digest
		streamID = stream
		record.UsageDigest = append([]byte(nil), digest...)
		record.SignatureVerified = true
		record.AuthenticationStatus = types.UsageAuthenticationStatusVerified
		record.LegacyUnverified = false
	} else {
		if authenticationActive {
			return types.ErrUsageAuthenticationRequired.Wrap("version-1 provider proof required")
		}
		record.SignatureVerified = false
		record.AuthenticationStatus = types.UsageAuthenticationStatusLegacy
		record.LegacyUnverified = true
	}

	// Generate the usage ID only after proof and replay checks. Exact retries do
	// not consume the module sequence.
	if record.UsageID == "" {
		seq := k.incrementUsageSequence(ctx)
		record.UsageID = generateIDWithTimestamp("usage", seq, ctx.BlockTime().Unix())
	}
	record.SubmittedAt = ctx.BlockTime()
	record.BlockHeight = ctx.BlockHeight()

	// Validate the record
	if err := record.Validate(); err != nil {
		return err
	}

	// Check if escrow exists and is active
	escrow, found := k.GetEscrowByOrder(ctx, record.OrderID)
	if !found {
		return types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", record.OrderID)
	}

	if escrow.State != types.EscrowStateActive {
		return types.ErrEscrowNotActive.Wrapf("escrow for order %s is not active", record.OrderID)
	}

	// Verify provider matches escrow recipient
	if record.Provider != escrow.Recipient {
		return types.ErrUnauthorized.Wrap("provider does not match escrow recipient")
	}

	// Save usage record
	if err := k.SetUsageRecord(ctx, *record); err != nil {
		return err
	}

	if err := k.recordBillingUsage(ctx, *record); err != nil {
		return err
	}

	if record.IsAuthenticated() {
		if err := k.commitUsageReplay(ctx, *record, streamID, usageDigest); err != nil {
			return err
		}
	}

	// Check if usage condition should be satisfied
	totalUsage := k.calculateTotalUsage(ctx, record.OrderID)
	if err := k.SatisfyUsageCondition(ctx, escrow.EscrowID, totalUsage); err != nil {
		k.Logger(ctx).Debug("failed to check usage condition", "error", err)
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventUsageRecorded{
		UsageID:    record.UsageID,
		OrderID:    record.OrderID,
		LeaseID:    record.LeaseID,
		Provider:   record.Provider,
		UsageUnits: record.UsageUnits,
		UsageType:  record.UsageType,
		TotalCost:  record.TotalCost.String(),
		RecordedAt: ctx.BlockTime().Unix(),
	})
	if err != nil {
		k.Logger(ctx).Error("failed to emit usage recorded event", "error", err)
	}

	k.Logger(ctx).Info("usage recorded",
		"usage_id", record.UsageID,
		"order_id", record.OrderID,
		"usage_units", record.UsageUnits,
	)

	return nil
}

// AcknowledgeUsage records a customer's acknowledgment of a usage record
func (k Keeper) AcknowledgeUsage(ctx sdk.Context, usageID string, customerSignature []byte) error {
	cacheCtx, write := ctx.CacheContext()
	if err := k.acknowledgeUsageLegacy(cacheCtx, usageID, customerSignature); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) acknowledgeUsageLegacy(ctx sdk.Context, usageID string, customerSignature []byte) error {
	if k.IsUsageAuthenticationActive(ctx) {
		return types.ErrUsageAuthenticationRequired.Wrap("versioned customer acknowledgment proof required")
	}
	usage, found := k.GetUsageRecord(ctx, usageID)
	if !found {
		return types.ErrUsageRecordNotFound.Wrapf("usage record %s not found", usageID)
	}

	if usage.Settled {
		return types.ErrUsageAlreadySettled.Wrapf("usage record %s already settled", usageID)
	}

	usage.CustomerAcknowledged = true
	usage.CustomerSignature = customerSignature

	if err := k.SetUsageRecord(ctx, usage); err != nil {
		return err
	}

	k.Logger(ctx).Info("usage acknowledged",
		"usage_id", usageID,
		"customer", usage.Customer,
	)

	return nil
}

// calculateTotalUsage calculates total usage units for an order
func (k Keeper) calculateTotalUsage(ctx sdk.Context, orderID string) uint64 {
	usages := k.GetUsageRecordsByOrder(ctx, orderID)
	var total uint64
	for _, usage := range usages {
		if k.IsUsageAuthenticationActive(ctx) && !usage.IsAuthenticated() {
			continue
		}
		if usage.UsageUnits > math.MaxUint64-total {
			return math.MaxUint64
		}
		total += usage.UsageUnits
	}
	return total
}

// AutoSettle performs automatic settlement for orders due for settlement
func (k Keeper) AutoSettle(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	settlementPeriod, err := durationFromSeconds(params.SettlementPeriod)
	if err != nil {
		return err
	}

	k.WithEscrowsByState(ctx, types.EscrowStateActive, func(escrow types.EscrowAccount) bool {
		// Check if due for settlement
		var lastSettlement time.Time
		settlements := k.GetSettlementsByOrder(ctx, escrow.OrderID)
		if len(settlements) > 0 {
			lastSettlement = settlements[len(settlements)-1].SettledAt
		} else if escrow.ActivatedAt != nil {
			lastSettlement = *escrow.ActivatedAt
		} else {
			lastSettlement = escrow.CreatedAt
		}

		if ctx.BlockTime().Sub(lastSettlement) >= settlementPeriod {
			// Get unsettled usage
			unsettled := k.GetUnsettledUsageRecords(ctx, escrow.OrderID)
			if len(unsettled) > 0 {
				usageIDs := make([]string, len(unsettled))
				for i, u := range unsettled {
					usageIDs[i] = u.UsageID
				}

				if _, err := k.SettleOrder(ctx, escrow.OrderID, usageIDs, false); err != nil {
					k.Logger(ctx).Error("auto-settlement failed",
						"order_id", escrow.OrderID,
						"error", err,
					)
				}
			}
		}

		return false
	})

	return nil
}
