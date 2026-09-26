// Package keeper implements the Fraud module keeper.
//
// A suspension or termination resolution removes an account's access
// network-wide. Neither may be driven by a single moderator, so those
// resolutions are recorded as a PendingResolution and only take effect once a
// second, distinct moderator-or-above identity confirms them within a bounded
// window. Unreviewed proposals lapse: the report is left as it was and the
// resolution can be re-proposed, which is the fail-safe direction.
package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// ============================================================================
// Storage
// ============================================================================

// SetPendingResolution stores a pending resolution for a report.
func (k Keeper) SetPendingResolution(ctx sdk.Context, pending types.PendingResolution) error {
	if err := pending.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	store.Set(PendingResolutionKey(pending.ReportID), bz)
	return nil
}

// GetPendingResolution returns the pending resolution for a report, if any.
func (k Keeper) GetPendingResolution(ctx sdk.Context, reportID string) (types.PendingResolution, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(PendingResolutionKey(reportID))
	if bz == nil {
		return types.PendingResolution{}, false
	}
	var pending types.PendingResolution
	if err := json.Unmarshal(bz, &pending); err != nil {
		return types.PendingResolution{}, false
	}
	return pending, true
}

// DeletePendingResolution removes a report's pending resolution.
func (k Keeper) DeletePendingResolution(ctx sdk.Context, reportID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(PendingResolutionKey(reportID))
}

// WithPendingResolutions iterates every pending resolution in deterministic
// store-key order.
func (k Keeper) WithPendingResolutions(ctx sdk.Context, fn func(types.PendingResolution) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, PendingResolutionPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var pending types.PendingResolution
		if err := json.Unmarshal(iter.Value(), &pending); err != nil {
			continue
		}
		if fn(pending) {
			break
		}
	}
}

// ============================================================================
// Proposal
// ============================================================================

// ProposeResolution records a suspension or termination resolution for review
// instead of applying it. The report's status and the subject's account are
// left untouched: that is the safeguard.
//
// Calling this for a resolution that does not require a second reviewer is an
// error; those are applied directly by ResolveFraudReport.
func (k Keeper) ProposeResolution(
	ctx sdk.Context,
	reportID string,
	resolution types.ResolutionType,
	notes string,
	moderatorAddr string,
) (types.PendingResolution, error) {
	if !resolution.RequiresSecondReviewer() {
		return types.PendingResolution{}, types.ErrInvalidResolution.Wrapf(
			"resolution %s does not require a second reviewer", resolution)
	}

	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.PendingResolution{}, types.ErrReportNotFound
	}
	if report.Status.IsTerminal() {
		return types.PendingResolution{}, types.ErrReportAlreadyResolved
	}
	if k.hasActiveCanonicalFinancialCase(ctx, report) {
		return types.PendingResolution{}, settlementtypes.ErrLegacyFinancialMutationFenced.Wrap(
			"canonical financial case remains active")
	}

	modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
	if err != nil {
		return types.PendingResolution{}, types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, modAddr) {
		return types.PendingResolution{}, types.ErrUnauthorizedModerator
	}
	if len(notes) > types.MaxResolutionNotesLength {
		return types.PendingResolution{}, types.ErrInvalidResolutionNotes.Wrapf(
			"maximum %d characters allowed", types.MaxResolutionNotesLength)
	}

	// A proposal for the same report replaces the previous one, so a moderator
	// can correct a lapsed or superseded proposal without operator action.
	pending := types.NewPendingResolution(
		reportID, moderatorAddr, resolution, notes,
		ctx.BlockTime().Unix(), ctx.BlockHeight())
	if err := k.SetPendingResolution(ctx, pending); err != nil {
		return types.PendingResolution{}, err
	}

	auditLog := k.createAuditLogEntry(
		ctx, reportID, types.AuditActionResolutionProposed, moderatorAddr,
		report.Status, report.Status,
		fmt.Sprintf("Resolution %s proposed pending second review by %s (lapses at %d)",
			resolution, moderatorAddr, pending.ExpiresAt),
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return types.PendingResolution{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResolutionProposed,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyResolution, resolution.String()),
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
			sdk.NewAttribute(types.AttributeKeyExpiresAt,
				strconv.FormatInt(pending.ExpiresAt, 10)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight,
				strconv.FormatInt(ctx.BlockHeight(), 10)),
		),
	)

	k.Logger(ctx).Info("resolution proposed pending second review",
		"report_id", reportID,
		"resolution", resolution.String(),
		"proposed_by", moderatorAddr,
		"expires_at", pending.ExpiresAt,
	)

	return pending, nil
}

// ============================================================================
// Confirmation
// ============================================================================

// ConfirmResolution applies a pending suspension or termination once a second,
// distinct moderator-or-above identity confirms it.
//
// This is the only path by which a suspension or termination takes effect.
func (k Keeper) ConfirmResolution(
	ctx sdk.Context,
	reportID string,
	reviewerAddr string,
) (types.ResolutionType, error) {
	pending, found := k.GetPendingResolution(ctx, reportID)
	if !found {
		return types.ResolutionTypeUnspecified, types.ErrResolutionNotPending
	}

	reviewer, err := sdk.AccAddressFromBech32(reviewerAddr)
	if err != nil {
		return types.ResolutionTypeUnspecified, types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, reviewer) {
		return types.ResolutionTypeUnspecified, types.ErrUnauthorizedModerator
	}
	if !pending.IsConfirmedBy(reviewerAddr) {
		return types.ResolutionTypeUnspecified, types.ErrSecondReviewerMustDiffer
	}

	// A proposal reviewed after its window has closed lapses rather than taking
	// effect. This is the enforcement for "bounded window": an unreviewed
	// suspension/termination simply never happens.
	if pending.IsExpired(ctx.BlockTime().Unix()) {
		k.DeletePendingResolution(ctx, reportID)
		k.recordResolutionLapse(ctx, reportID, pending)
		return types.ResolutionTypeUnspecified, types.ErrPendingResolutionExpired
	}

	// Apply the resolution through the ordinary path, now that two distinct
	// moderators have agreed to it.
	if err := k.applyResolution(ctx, reportID, pending.Resolution, pending.Notes, reviewerAddr); err != nil {
		return types.ResolutionTypeUnspecified, err
	}

	k.DeletePendingResolution(ctx, reportID)

	report, found := k.GetFraudReport(ctx, reportID)
	previous := types.FraudReportStatusReviewing
	if found {
		previous = report.Status
	}

	auditLog := k.createAuditLogEntry(
		ctx, reportID, types.AuditActionResolutionConfirmed, reviewerAddr,
		previous, types.FraudReportStatusResolved,
		fmt.Sprintf("Resolution %s confirmed by second reviewer %s (proposed by %s)",
			pending.Resolution, reviewerAddr, pending.ProposedBy),
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return types.ResolutionTypeUnspecified, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResolutionConfirmed,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyResolution, pending.Resolution.String()),
			sdk.NewAttribute(types.AttributeKeyModerator, pending.ProposedBy),
			sdk.NewAttribute(types.AttributeKeySecondReviewer, reviewerAddr),
			sdk.NewAttribute(types.AttributeKeyBlockHeight,
				strconv.FormatInt(ctx.BlockHeight(), 10)),
		),
	)

	k.Logger(ctx).Info("resolution confirmed by second reviewer",
		"report_id", reportID,
		"resolution", pending.Resolution.String(),
		"proposed_by", pending.ProposedBy,
		"second_reviewer", reviewerAddr,
	)

	return pending.Resolution, nil
}

// ============================================================================
// Expiry
// ============================================================================

// ExpirePendingResolutions removes every pending resolution whose window has
// closed, returning the proposals that lapsed in deterministic order.
func (k Keeper) ExpirePendingResolutions(ctx sdk.Context) ([]types.PendingResolution, error) {
	now := ctx.BlockTime().Unix()

	// WithPendingResolutions iterates in store-key order, so this is deterministic.
	var due []types.PendingResolution
	k.WithPendingResolutions(ctx, func(p types.PendingResolution) bool {
		if p.IsExpired(now) {
			due = append(due, p)
		}
		return false
	})

	for i := range due {
		k.DeletePendingResolution(ctx, due[i].ReportID)
		k.recordResolutionLapse(ctx, due[i].ReportID, due[i])
	}
	return due, nil
}

// ProcessPendingResolutionExpiry is the module EndBlocker for the co-signature
// window: an unreviewed suspension or termination lapses here, with no effect
// on the report or the subject, and no operator action required.
func (k Keeper) ProcessPendingResolutionExpiry(ctx sdk.Context) error {
	_, err := k.ExpirePendingResolutions(ctx)
	return err
}

// recordResolutionLapse writes the audit trail and event for a lapsed proposal.
func (k Keeper) recordResolutionLapse(ctx sdk.Context, reportID string, pending types.PendingResolution) {
	report, found := k.GetFraudReport(ctx, reportID)
	status := types.FraudReportStatusReviewing
	if found {
		status = report.Status
	}

	auditLog := k.createAuditLogEntry(
		ctx, reportID, types.AuditActionResolutionLapsed, pending.ProposedBy,
		status, status,
		fmt.Sprintf("Resolution %s lapsed unreviewed at %d",
			pending.Resolution, pending.ExpiresAt),
	)
	// Best-effort: a lapse must never fail the block, since the only effect is
	// that nothing happens.
	_ = k.CreateAuditLog(ctx, auditLog)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResolutionLapsed,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyResolution, pending.Resolution.String()),
			sdk.NewAttribute(types.AttributeKeyModerator, pending.ProposedBy),
			sdk.NewAttribute(types.AttributeKeyExpiresAt,
				strconv.FormatInt(pending.ExpiresAt, 10)),
		),
	)

	k.Logger(ctx).Info("pending resolution lapsed unreviewed",
		"report_id", reportID,
		"resolution", pending.Resolution.String(),
		"proposed_by", pending.ProposedBy,
	)
}
