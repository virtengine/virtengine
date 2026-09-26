// Package keeper implements the Fraud module keeper.
//
// VE-912: Fraud reporting flow - Reporter authorization, spam control and the
// response/rebuttal path.
//
// This file implements the additive "open reporting" surface:
//
//  1. Reporter standing - a provider may always report (unchanged), and any
//     other party to a transaction may report when the report is anchored to an
//     order they are a party to, or carries an explicit no-order-available
//     basis whose justification lives in encrypted evidence.
//  2. Spam control - identical resubmissions are de-duplicated on a content
//     fingerprint, and a per-reporter submission rate is enforced over a
//     block-height window (never wall clock, so the check is deterministic).
//  3. Response path - the reported party (or the reporter) may file a response
//     record that lands in the same moderator queue entry as the report.
package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// MarketKeeper is the subset of the market keeper needed to verify that a
// reporter is actually a party to the order they cite as standing.
type MarketKeeper interface {
	GetOrderByID(ctx sdk.Context, orderID string) (interface{}, bool)
	GetOrderCustomer(ctx sdk.Context, orderID string) string
	GetOrderProvider(ctx sdk.Context, orderID string) string
}

// SetMarketKeeper wires the market keeper used for order-standing checks.
//
// It is required for order-linked reporting: a non-provider reporter's standing
// is established by being a party to a real order, and that can only be checked
// against the market module. If this is never called, order-linked reports from
// non-providers are refused with ErrOrderVerificationUnavailable rather than
// being taken on trust (see IsPartyToOrder).
func (k *Keeper) SetMarketKeeper(marketKeeper MarketKeeper) {
	k.marketKeeper = marketKeeper
}

// IsPartyToOrder reports whether addr is the customer or the provider of the
// given order. Only orders that actually exist can confer standing.
//
// It fails closed. If no market keeper is wired, standing cannot be verified, so
// the order reference is refused outright rather than being accepted as an
// unverified claim: an unverifiable claim must never confer standing, because
// that would let any address cite any order ID and be treated as a party to it.
// Callers see the failure as found=false and should surface
// ErrOrderVerificationUnavailable rather than a "no such order" message.
func (k Keeper) IsPartyToOrder(ctx sdk.Context, orderID, addr string) (found bool, isParty bool) {
	if orderID == "" || addr == "" {
		return false, false
	}
	if k.marketKeeper == nil {
		// Unwired market keeper: we cannot verify standing. Refuse.
		return false, false
	}
	if _, ok := k.marketKeeper.GetOrderByID(ctx, orderID); !ok {
		return false, false
	}
	customer := k.marketKeeper.GetOrderCustomer(ctx, orderID)
	provider := k.marketKeeper.GetOrderProvider(ctx, orderID)
	return true, (customer != "" && customer == addr) || (provider != "" && provider == addr)
}

// authorizeReportReporter enforces reporter standing.
//
// Providers keep their existing, unconditional reporting right (this change is
// additive and must not weaken provider reporting). Every other reporter must
// anchor the report to a transaction: either an order they are a party to, or an
// explicit no-order-available basis, whose justification is carried in the
// mandatory encrypted evidence rather than in a public free-text field.
func (k Keeper) authorizeReportReporter(ctx sdk.Context, report types.FraudReport) error {
	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return types.ErrInvalidReporter.Wrap(err.Error())
	}
	if k.IsProvider(ctx, reporterAddr) {
		return nil
	}

	// A provider-only report is fine; a non-provider needs an order anchor.
	if len(report.RelatedOrderIDs) == 0 {
		if !report.NoOrderAvailable {
			return types.ErrMissingOrderReference.Wrap(
				"non-provider reports require a related order/resource reference, or an explicit no-order-available basis with justification in encrypted evidence")
		}
		// The "no order available" basis is only credible with a justification,
		// which by design lives in encrypted evidence for the reviewers.
		if len(report.Evidence) == 0 {
			return types.ErrMissingOrderReference.Wrap(
				"no-order-available basis requires a justification in encrypted evidence")
		}
		return nil
	}

	// With at least one order reference the reporter must be a party to it.
	//
	// If the market keeper is unwired, order standing cannot be verified at all.
	// Report that as its own configuration fault rather than as a per-order
	// "does not exist", which would send operators hunting a reporter problem
	// that does not exist.
	if k.marketKeeper == nil {
		return types.ErrOrderVerificationUnavailable.Wrap(
			"cannot verify order standing: fraud keeper has no market keeper wired")
	}
	for _, orderID := range report.RelatedOrderIDs {
		orderID = strings.TrimSpace(orderID)
		if orderID == "" {
			continue
		}
		found, isParty := k.IsPartyToOrder(ctx, orderID, report.Reporter)
		if !found {
			return types.ErrMissingOrderReference.Wrapf(
				"referenced order %q does not exist", orderID)
		}
		if isParty {
			return nil
		}
	}
	return types.ErrUnauthorizedReporter.Wrap(
		"reporter is not a party to any referenced order")
}

// checkDuplicateReport rejects a byte-identical resubmission from the same
// reporter so a repeat report never creates a second moderator-queue entry.
func (k Keeper) checkDuplicateReport(ctx sdk.Context, report types.FraudReport, fingerprint string) error {
	store := ctx.KVStore(k.skey)
	key := types.GetDedupIndexKey(report.Reporter, fingerprint)
	if existing := store.Get(key); existing != nil {
		return types.ErrDuplicateReport.Wrapf(
			"identical report already submitted as %s", string(existing))
	}
	return nil
}

// recordReportFingerprint stores the dedup fingerprint and the reporter's
// submission height for rate limiting.
func (k Keeper) recordReportFingerprint(ctx sdk.Context, report types.FraudReport, fingerprint string) {
	store := ctx.KVStore(k.skey)
	store.Set(types.GetDedupIndexKey(report.Reporter, fingerprint), []byte(report.ID))
	store.Set(types.GetReporterActivityKey(report.Reporter, ctx.BlockHeight(), report.ID), []byte(report.ID))
}

// checkReporterRateLimit enforces the per-reporter submission limit over a
// block-height window. Iteration is over a byte-ordered prefix, so the result is
// deterministic across nodes.
func (k Keeper) checkReporterRateLimit(ctx sdk.Context, reporter string) error {
	params := k.GetParams(ctx)
	maxPerWindow := params.MaxReportsPerWindow
	window := params.ReportWindowBlocks
	if maxPerWindow <= 0 || window <= 0 {
		return nil
	}

	height := ctx.BlockHeight()
	lowerBound := height - int64(window)
	store := ctx.KVStore(k.skey)

	iter := storetypes.KVStorePrefixIterator(store, types.GetReporterActivityPrefix(reporter))
	defer iter.Close()

	used := 0
	// Collect expired entries and delete them after iteration: mutating the store
	// while an iterator is open is unsafe, and an unbounded per-reporter prefix is
	// a gas/DoS smell that a growing chain would otherwise never reclaim.
	var stale [][]byte
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		prefix := types.GetReporterActivityPrefix(reporter)
		rest := key[len(prefix):]
		if len(rest) < 8 {
			continue
		}
		rawHeight := binary.BigEndian.Uint64(rest[:8])
		if rawHeight > math.MaxInt64 {
			// This keeper only writes non-negative block heights, so a key
			// with the high bit set did not come from here; reclaim it
			// instead of wrapping it into a negative height.
			stale = append(stale, append([]byte(nil), key...))
			continue
		}
		submittedAt := int64(rawHeight) /* #nosec G115 -- range-checked against math.MaxInt64 above */ //nolint:gosec
		switch {
		case submittedAt > height:
			// Defensive: a future height can never come from this keeper, but if it
			// does, it must not be counted as usage in the current window.
			stale = append(stale, append([]byte(nil), key...))
		case submittedAt > lowerBound:
			used++
		default:
			stale = append(stale, append([]byte(nil), key...))
		}
	}

	for _, key := range stale {
		store.Delete(key)
	}

	if used >= maxPerWindow {
		return types.ErrReporterRateLimited.Wrapf(
			"reporter %s already submitted %d reports within the last %d blocks",
			reporter, used, window)
	}
	return nil
}

// ============================================================================
// Responses / rebuttals
// ============================================================================

// GetNextFraudResponseSequence returns the next fraud response sequence number
func (k Keeper) GetNextFraudResponseSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyFraudResponse)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextFraudResponseSequence sets the next fraud response sequence number
func (k Keeper) SetNextFraudResponseSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.SequenceKeyFraudResponse, bz)
}

// SubmitFraudResponse stores a response/rebuttal against an existing report.
//
// Only the reported party (role reported_party) or the original reporter
// (role reporter) may respond, and only while the report is still pending
// moderator action. The response is linked to its report so it appears in the
// same moderator-queue entry as the report.
func (k Keeper) SubmitFraudResponse(ctx sdk.Context, response *types.FraudResponse) error {
	// All rejection checks run BEFORE the response ID is minted.
	//
	// Minting writes the fraud-response sequence, which is consensus-critical
	// state. If the ID were assigned first, every rejected response would still
	// consume a sequence number. The msgServer currently contains that because it
	// wraps this call in a CacheContext and only commits on success, but nothing
	// in the keeper enforces it: a direct keeper caller would silently burn
	// sequences. Minting only once the response is known-acceptable makes the
	// keeper correct on its own, and makes the write unconditional afterwards.
	//
	// Note the ordering constraint: Validate requires a non-empty ID, and
	// ComputeContentHash includes the ID, so minting must still precede both.
	report, found := k.GetFraudReport(ctx, response.ReportID)
	if !found {
		return types.ErrReportNotFound
	}
	if report.Status.IsTerminal() {
		return types.ErrReportNotPending.Wrapf("report %s is %s", report.ID, report.Status.String())
	}

	if response.Respondent == "" {
		return types.ErrUnauthorizedRespondent.Wrap("respondent address is required")
	}

	// Determine the respondent's role from the report itself: the role is
	// derived, never trusted from the caller alone.
	switch response.Respondent {
	case report.ReportedParty:
		response.Role = types.FraudRespondentRoleReportedParty
	case report.Reporter:
		response.Role = types.FraudRespondentRoleReporter
	default:
		return types.ErrUnauthorizedRespondent.Wrapf(
			"address %s is neither the reported party nor the reporter of %s",
			response.Respondent, report.ID)
	}

	if response.ID == "" {
		seq := k.GetNextFraudResponseSequence(ctx)
		response.ID = fmt.Sprintf("%s/response-%d", response.ReportID, seq)
		k.SetNextFraudResponseSequence(ctx, seq+1)
	}
	response.ContentHash = response.ComputeContentHash()

	if err := response.Validate(); err != nil {
		return err
	}

	if err := k.SetFraudResponse(ctx, *response); err != nil {
		return err
	}

	// Link the response to the report and keep the public counter accurate.
	store := ctx.KVStore(k.skey)
	store.Set(types.GetReportResponseKey(report.ID, response.ID), []byte(response.ID))

	report.ResponseCount++
	report.UpdatedAt = ctx.BlockTime()
	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Surface the response alongside the report in the moderator queue.
	if entry, inQueue := k.GetModeratorQueueEntry(ctx, report.ID); inQueue {
		entry.ResponseCount = report.ResponseCount
		if err := k.AddToModeratorQueue(ctx, entry); err != nil {
			return err
		}
	}

	auditLog := k.createAuditLogEntry(
		ctx,
		report.ID,
		types.AuditActionResponded,
		response.Respondent,
		report.Status,
		report.Status,
		fmt.Sprintf("Response %s filed by %s (%s)", response.ID, response.Respondent, response.Role.String()),
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudResponseSubmitted,
			sdk.NewAttribute(types.AttributeKeyReportID, report.ID),
			sdk.NewAttribute(types.AttributeKeyResponseID, response.ID),
			sdk.NewAttribute(types.AttributeKeyRespondent, response.Respondent),
			sdk.NewAttribute(types.AttributeKeyRespondentRole, response.Role.String()),
			sdk.NewAttribute(types.AttributeKeyResponseCount, fmt.Sprintf("%d", report.ResponseCount)),
		),
	)

	k.Logger(ctx).Info("fraud response submitted",
		"report_id", report.ID,
		"response_id", response.ID,
		"respondent", response.Respondent,
		"role", response.Role.String(),
	)

	return nil
}

// SetFraudResponse stores a response record (and its report index entry).
//
// Responses are stored as JSON like the other local fraud types (FraudReport,
// FraudAuditLog, ModeratorQueueEntry) so the keeper does not need the local type
// registered as a proto message.
func (k Keeper) SetFraudResponse(ctx sdk.Context, response types.FraudResponse) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(response)
	if err != nil {
		return err
	}
	store.Set(types.GetResponseKey(response.ID), bz)
	return nil
}

// GetFraudResponse returns a response record by ID.
func (k Keeper) GetFraudResponse(ctx sdk.Context, responseID string) (types.FraudResponse, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetResponseKey(responseID))
	if bz == nil {
		return types.FraudResponse{}, false
	}
	var response types.FraudResponse
	if err := json.Unmarshal(bz, &response); err != nil {
		return types.FraudResponse{}, false
	}
	return response, true
}

// GetFraudResponses returns every response filed against a report, in insertion
// (byte-order of the sequence-preserving ID) order.
func (k Keeper) GetFraudResponses(ctx sdk.Context, reportID string) []types.FraudResponse {
	var responses []types.FraudResponse
	k.WithFraudResponses(ctx, reportID, func(response types.FraudResponse) bool {
		responses = append(responses, response)
		return false
	})
	return responses
}

// WithFraudResponses iterates over the responses filed against a report.
func (k Keeper) WithFraudResponses(ctx sdk.Context, reportID string, fn func(types.FraudResponse) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.GetReportResponsesKey(reportID))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		responseID := string(iter.Value())
		response, found := k.GetFraudResponse(ctx, responseID)
		if !found {
			continue
		}
		if fn(response) {
			break
		}
	}
}
