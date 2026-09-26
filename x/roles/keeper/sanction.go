package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	"github.com/virtengine/virtengine/x/roles/types"
)

// This file implements the sanction state machine for the roles module.
//
// Account state is a *projection* of the in-force sanctions for an account
// rather than a bare enum an administrator flips. Every transition here is a
// total, side-effect-free function of its inputs plus block time: no wall-clock
// reads, no host/env input, no map-iteration-order dependence.
//
// Two-regime design:
//
//   - Warning: active immediately, no account-state effect.
//   - Emergency hold: active immediately (that is its purpose) so a single
//     moderator can stop active abuse, but it is created PendingReview and
//     auto-expires at its window unless a second, distinct reviewer confirms it.
//   - Suspension / Termination: created PendingReview and *not* in force, so a
//     single moderator cannot impose one at all. A second, distinct
//     moderator-or-above identity must confirm it before it takes effect.

// ============================================================================
// Sequences
// ============================================================================

// GetNextSanctionSequence returns the next sanction sequence number.
func (k Keeper) GetNextSanctionSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.PrefixSanctionSequence)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextSanctionSequence sets the next sanction sequence number.
func (k Keeper) SetNextSanctionSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.PrefixSanctionSequence, bz)
}

// ============================================================================
// Storage
// ============================================================================

// SetSanction stores a sanction record and its subject index entry.
func (k Keeper) SetSanction(ctx sdk.Context, sanction types.Sanction) error {
	if err := sanction.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(sanction)
	if err != nil {
		return err
	}
	store.Set(types.SanctionKey(sanction.ID), bz)
	store.Set(types.SubjectSanctionKey(sanction.Subject, sanction.ID), []byte(sanction.ID))
	return nil
}

// GetSanction returns a sanction record by ID.
func (k Keeper) GetSanction(ctx sdk.Context, sanctionID string) (types.Sanction, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SanctionKey(sanctionID))
	if bz == nil {
		return types.Sanction{}, false
	}
	var sanction types.Sanction
	if err := json.Unmarshal(bz, &sanction); err != nil {
		return types.Sanction{}, false
	}
	return sanction, true
}

// GetSanctionsForSubject returns every sanction record for a subject, in
// deterministic store-key order.
func (k Keeper) GetSanctionsForSubject(ctx sdk.Context, subject string) []types.Sanction {
	store := ctx.KVStore(k.skey)

	var out []types.Sanction
	iter := storetypes.KVStorePrefixIterator(store, types.SubjectSanctionPrefixKey(subject))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if sanction, found := k.GetSanction(ctx, string(iter.Value())); found {
			out = append(out, sanction)
		}
	}
	return out
}

// WithSanctions iterates every sanction record in deterministic store-key order.
func (k Keeper) WithSanctions(ctx sdk.Context, fn func(types.Sanction) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixSanction)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var sanction types.Sanction
		if err := json.Unmarshal(iter.Value(), &sanction); err != nil {
			continue
		}
		if fn(sanction) {
			break
		}
	}
}

// GetAllSanctions returns every sanction record.
func (k Keeper) GetAllSanctions(ctx sdk.Context) []types.Sanction {
	var out []types.Sanction
	k.WithSanctions(ctx, func(s types.Sanction) bool {
		out = append(out, s)
		return false
	})
	return out
}

// GetOpenAppealForSubject returns the open appeal for a subject, if any.
func (k Keeper) GetOpenAppealForSubject(ctx sdk.Context, subject string) (types.Sanction, bool) {
	for _, sanction := range k.GetSanctionsForSubject(ctx, subject) {
		if sanction.Status == types.SanctionStatusAppealPending {
			return sanction, true
		}
	}
	return types.Sanction{}, false
}

// ============================================================================
// Transitions
// ============================================================================

// ImposeSanction creates a sanction record from a proposal.
//
// The imposing actor must be moderator-or-above. Suspensions and terminations
// are created PendingReview and therefore have no effect until confirmed by a
// second, distinct reviewer (see ConfirmSanction). Emergency holds bind
// immediately but carry a bounded expiry.
func (k Keeper) ImposeSanction(
	ctx sdk.Context,
	proposal types.Sanction,
	imposedBy sdk.AccAddress,
) (types.Sanction, error) {
	if !k.IsModerator(ctx, imposedBy) {
		return types.Sanction{}, types.ErrSecondReviewerUnauthorized.Wrap(
			"imposing actor is not a moderator")
	}

	proposal.ImposedBy = imposedBy.String()
	if err := proposal.ValidateProposal(); err != nil {
		return types.Sanction{}, err
	}
	if _, err := sdk.AccAddressFromBech32(proposal.Subject); err != nil {
		return types.Sanction{}, types.ErrInvalidAddress.Wrap(err.Error())
	}

	// An open appeal blocks harsher escalation until it is reviewed.
	if appeal, open := k.GetOpenAppealForSubject(ctx, proposal.Subject); open {
		challenged, found := k.GetSanction(ctx, appeal.AppealOf)
		if found && proposal.Kind.SeverityRank() > challenged.Kind.SeverityRank() {
			return types.Sanction{}, types.ErrAppealAlreadyOpen.Wrapf(
				"escalation to %s is blocked until appeal %s is reviewed",
				proposal.Kind, appeal.ID)
		}
	}

	now := ctx.BlockTime().Unix()
	status := types.SanctionStatusActive

	switch proposal.Kind {
	case types.SanctionKindWarning:
		// Advisory only: active immediately, no account-state effect.

	case types.SanctionKindEmergencyHold:
		// Binds immediately, but only for a bounded window and only until a
		// distinct reviewer confirms it.
		status = types.SanctionStatusPendingReview
		if proposal.ExpiresAt == 0 {
			proposal.ExpiresAt = now + types.DefaultEmergencyHoldSeconds
		}
		if proposal.ExpiresAt <= now {
			return types.Sanction{}, types.ErrInvalidSanction.Wrap(
				"an emergency hold must expire in the future")
		}
		if proposal.ExpiresAt > now+types.MaxEmergencyHoldSeconds {
			return types.Sanction{}, types.ErrEmergencyHoldTooLong.Wrapf(
				"maximum unreviewed hold is %d seconds", types.MaxEmergencyHoldSeconds)
		}

	case types.SanctionKindSuspension:
		// Requires a duration: a suspension is time-limited by definition.
		status = types.SanctionStatusPendingReview
		if proposal.ExpiresAt == 0 {
			return types.Sanction{}, types.ErrInvalidSanction.Wrap(
				"a suspension must carry an expiry")
		}
		if proposal.ExpiresAt <= now {
			return types.Sanction{}, types.ErrInvalidSanction.Wrap(
				"a suspension must expire in the future")
		}

	case types.SanctionKindTermination:
		// Indefinite, but reversible through a reviewed path.
		status = types.SanctionStatusPendingReview

	default:
		return types.Sanction{}, types.ErrInvalidSanctionKind.Wrapf(
			"unknown kind %d", proposal.Kind)
	}

	seq := k.GetNextSanctionSequence(ctx)
	k.SetNextSanctionSequence(ctx, seq+1)

	record := types.NewSanctionRecord(
		types.SanctionID(seq), proposal, status, imposedBy.String(), now, ctx.BlockHeight())
	if err := record.Validate(); err != nil {
		return types.Sanction{}, err
	}
	if err := k.SetSanction(ctx, record); err != nil {
		return types.Sanction{}, err
	}

	subject, _ := sdk.AccAddressFromBech32(record.Subject)
	if _, err := k.ProjectAccountState(ctx, subject); err != nil {
		return types.Sanction{}, err
	}

	emitSanctionEvent(ctx, types.EventTypeSanctionImposed, record)
	k.Logger(ctx).Info("sanction imposed",
		"sanction_id", record.ID,
		"subject", record.Subject,
		"kind", record.Kind.String(),
		"scope", record.Scope.String(),
		"status", record.Status.String(),
	)
	return record, nil
}

// ConfirmSanction records the second, distinct reviewer's approval of a
// pending sanction and puts it into force.
//
// This is the only path by which a suspension or termination takes effect, and
// the reviewer must differ from the imposing actor. It is also how an emergency
// hold is reviewed before its window lapses.
//
// reviewedUntil is the block time (Unix seconds) the confirmed sanction should
// run to, and is required when confirming an emergency hold: a hold is an
// interim measure, so the reviewer must decide the duration it is being
// converted into. It is ignored for suspensions and terminations, which already
// carry their own duration (or are indefinite).
func (k Keeper) ConfirmSanction(
	ctx sdk.Context,
	sanctionID string,
	reviewer sdk.AccAddress,
	reviewedUntil int64,
) (types.Sanction, error) {
	sanction, found := k.GetSanction(ctx, sanctionID)
	if !found {
		return types.Sanction{}, types.ErrSanctionNotFound
	}
	if sanction.IsAppeal() {
		return types.Sanction{}, types.ErrSanctionNotAppealable.Wrap(
			"appeal records are resolved, not confirmed")
	}
	if !k.IsModerator(ctx, reviewer) {
		return types.Sanction{}, types.ErrSecondReviewerUnauthorized
	}
	if reviewer.String() == sanction.ImposedBy {
		return types.Sanction{}, types.ErrSecondReviewerMustDiffer
	}
	if sanction.Status != types.SanctionStatusPendingReview {
		return types.Sanction{}, types.ErrInvalidSanctionTransition.Wrapf(
			"cannot confirm a sanction in status %s", sanction.Status)
	}

	now := ctx.BlockTime().Unix()
	if sanction.ExpiresAt > 0 && sanction.ExpiresAt <= now {
		// The window closed before review: the hold lapses instead of being
		// confirmable, which is the enforcement for "must expire unless reviewed".
		sanction.Status = types.SanctionStatusExpired
		if err := k.SetSanction(ctx, sanction); err != nil {
			return types.Sanction{}, err
		}
		subject, _ := sdk.AccAddressFromBech32(sanction.Subject)
		if _, err := k.ProjectAccountState(ctx, subject); err != nil {
			return types.Sanction{}, err
		}
		emitSanctionEvent(ctx, types.EventTypeSanctionExpired, sanction)
		return sanction, types.ErrEmergencyHoldExpired
	}

	// A confirmed emergency hold is converted into a reviewed, time-limited
	// sanction whose duration the reviewer sets. Without this the hold would be
	// reaped by its own review window the moment that window elapsed.
	if sanction.Kind == types.SanctionKindEmergencyHold {
		if reviewedUntil <= now {
			return types.Sanction{}, types.ErrInvalidSanction.Wrap(
				"confirming an emergency hold requires a duration extending past the review time")
		}
		sanction.ExpiresAt = reviewedUntil
	}

	sanction.Status = types.SanctionStatusActive
	sanction.SecondReviewer = reviewer.String()
	sanction.SecondReviewedAt = now
	if err := k.SetSanction(ctx, sanction); err != nil {
		return types.Sanction{}, err
	}

	subject, _ := sdk.AccAddressFromBech32(sanction.Subject)
	if _, err := k.ProjectAccountState(ctx, subject); err != nil {
		return types.Sanction{}, err
	}

	emitSanctionEvent(ctx, types.EventTypeSanctionConfirmed, sanction)
	k.Logger(ctx).Info("sanction confirmed",
		"sanction_id", sanction.ID,
		"subject", sanction.Subject,
		"kind", sanction.Kind.String(),
		"second_reviewer", sanction.SecondReviewer,
	)
	return sanction, nil
}

// RevokeSanction clears an in-force or pending sanction. Revocation is
// de-escalation, so it needs only one moderator-or-above actor.
func (k Keeper) RevokeSanction(
	ctx sdk.Context,
	sanctionID string,
	actor sdk.AccAddress,
	reason string,
) (types.Sanction, error) {
	sanction, found := k.GetSanction(ctx, sanctionID)
	if !found {
		return types.Sanction{}, types.ErrSanctionNotFound
	}
	if sanction.IsAppeal() {
		return types.Sanction{}, types.ErrSanctionNotAppealable.Wrap(
			"appeal records are resolved, not revoked")
	}
	if !k.IsModerator(ctx, actor) {
		return types.Sanction{}, types.ErrUnauthorized
	}
	if !sanction.Status.InForce() && sanction.Status != types.SanctionStatusActive {
		return types.Sanction{}, types.ErrInvalidSanctionTransition.Wrapf(
			"cannot revoke a sanction in status %s", sanction.Status)
	}

	sanction.Status = types.SanctionStatusRevoked
	if err := k.SetSanction(ctx, sanction); err != nil {
		return types.Sanction{}, err
	}

	subject, _ := sdk.AccAddressFromBech32(sanction.Subject)
	if _, err := k.ProjectAccountState(ctx, subject); err != nil {
		return types.Sanction{}, err
	}

	emitSanctionEvent(ctx, types.EventTypeSanctionRevoked, sanction,
		sdk.NewAttribute(types.AttributeKeyModifiedBy, actor.String()),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	)
	return sanction, nil
}

// OpenAppeal opens an appeal record against an in-force sanction.
//
// While the appeal is pending it blocks any harsher escalation against the same
// subject. Only the sanctioned party may appeal, and only one appeal may be open
// at a time.
func (k Keeper) OpenAppeal(
	ctx sdk.Context,
	sanctionID string,
	appellant sdk.AccAddress,
	justification string,
) (types.Sanction, error) {
	target, found := k.GetSanction(ctx, sanctionID)
	if !found {
		return types.Sanction{}, types.ErrSanctionNotFound
	}
	if target.IsAppeal() {
		return types.Sanction{}, types.ErrSanctionNotAppealable.Wrap(
			"an appeal record cannot itself be appealed")
	}
	if !target.InForce() {
		return types.Sanction{}, types.ErrSanctionNotAppealable.Wrapf(
			"sanction %s is not in force", target.ID)
	}
	if appellant.String() != target.Subject {
		return types.Sanction{}, types.ErrSanctionSubjectMismatch.Wrap(
			"only the sanctioned party may appeal")
	}
	if existing, open := k.GetOpenAppealForSubject(ctx, target.Subject); open {
		return types.Sanction{}, types.ErrAppealAlreadyOpen.Wrapf(
			"appeal %s is already open", existing.ID)
	}

	now := ctx.BlockTime().Unix()
	appeal := types.Sanction{
		Subject:       target.Subject,
		Scope:         target.Scope,
		ScopeRef:      target.ScopeRef,
		Kind:          target.Kind,
		ReasonCode:    target.ReasonCode,
		Justification: justification,
		ImposedBy:     appellant.String(),
		AppealOf:      target.ID,
		Status:        types.SanctionStatusAppealPending,
		ExpiresAt:     0,
	}

	if err := appeal.ValidateAppealProposal(); err != nil {
		return types.Sanction{}, err
	}

	seq := k.GetNextSanctionSequence(ctx)
	k.SetNextSanctionSequence(ctx, seq+1)

	record := types.NewSanctionRecord(
		types.SanctionID(seq), appeal, types.SanctionStatusAppealPending,
		appellant.String(), now, ctx.BlockHeight())
	if err := record.Validate(); err != nil {
		return types.Sanction{}, err
	}
	if err := k.SetSanction(ctx, record); err != nil {
		return types.Sanction{}, err
	}

	emitSanctionEvent(ctx, types.EventTypeAppealOpened, record,
		sdk.NewAttribute(types.AttributeKeyAppealID, record.ID),
	)
	return record, nil
}

// ResolveAppeal resolves an open appeal.
//
// Granting the appeal revokes the challenged sanction; the account-state
// projection then restores the previous state without manual intervention.
// Denying it leaves the challenged sanction in force.
func (k Keeper) ResolveAppeal(
	ctx sdk.Context,
	appealID string,
	reviewer sdk.AccAddress,
	grant bool,
	notes string,
) (types.Sanction, error) {
	appeal, found := k.GetSanction(ctx, appealID)
	if !found {
		return types.Sanction{}, types.ErrSanctionNotFound
	}
	if appeal.Status != types.SanctionStatusAppealPending {
		return types.Sanction{}, types.ErrAppealNotOpen.Wrapf(
			"appeal %s is in status %s", appeal.ID, appeal.Status)
	}
	if !k.IsModerator(ctx, reviewer) {
		return types.Sanction{}, types.ErrSecondReviewerUnauthorized
	}
	// The appellant may not judge their own appeal...
	if reviewer.String() == appeal.ImposedBy {
		return types.Sanction{}, types.ErrSecondReviewerMustDiffer
	}

	target, targetFound := k.GetSanction(ctx, appeal.AppealOf)
	if targetFound {
		// ...nor may anyone who took part in the decision under appeal: neither
		// the actor who imposed the challenged sanction nor the second reviewer
		// who put it into force.
		if reviewer.String() == target.ImposedBy ||
			(reviewer.String() != "" && reviewer.String() == target.SecondReviewer) {
			return types.Sanction{}, types.ErrSecondReviewerMustDiffer.Wrap(
				"a participant in the challenged decision may not decide the appeal")
		}
	}

	now := ctx.BlockTime().Unix()
	if grant {
		appeal.Status = types.SanctionStatusAppealGranted
		if targetFound {
			target.Status = types.SanctionStatusRevoked
		}
	} else {
		appeal.Status = types.SanctionStatusAppealDenied
	}
	appeal.SecondReviewer = reviewer.String()
	appeal.SecondReviewedAt = now

	if err := k.SetSanction(ctx, appeal); err != nil {
		return types.Sanction{}, err
	}
	if targetFound {
		if err := k.SetSanction(ctx, target); err != nil {
			return types.Sanction{}, err
		}
	}

	subject, _ := sdk.AccAddressFromBech32(appeal.Subject)
	if _, err := k.ProjectAccountState(ctx, subject); err != nil {
		return types.Sanction{}, err
	}

	emitSanctionEvent(ctx, types.EventTypeAppealResolved, appeal,
		sdk.NewAttribute(types.AttributeKeyAppealID, appeal.ID),
		sdk.NewAttribute(types.AttributeKeyReason, notes),
	)
	k.Logger(ctx).Info("appeal resolved",
		"appeal_id", appeal.ID,
		"sanction_id", appeal.AppealOf,
		"granted", grant,
		"reviewer", reviewer.String(),
	)
	return appeal, nil
}

// ============================================================================
// Expiry
// ============================================================================

// ExpireSanctions lapses every sanction whose expiry has passed, returning the
// records that were expired in deterministic order.
//
// Both confirmed sanctions and unconfirmed pending ones are covered: an
// emergency hold that is never reviewed lapses at its window, and so does a
// suspension that was proposed but never confirmed. Neither has any further
// effect once lapsed.
func (k Keeper) ExpireSanctions(ctx sdk.Context) ([]types.Sanction, error) {
	now := ctx.BlockTime().Unix()

	// WithSanctions iterates in store-key order, so `due` is deterministic.
	var due []types.Sanction
	k.WithSanctions(ctx, func(s types.Sanction) bool {
		if s.IsAppeal() {
			return false
		}
		lapsable := s.Status == types.SanctionStatusActive ||
			s.Status == types.SanctionStatusPendingReview
		if lapsable && s.ExpiresAt > 0 && s.ExpiresAt <= now {
			due = append(due, s)
		}
		return false
	})

	for i := range due {
		sanction := due[i]
		sanction.Status = types.SanctionStatusExpired
		if err := k.SetSanction(ctx, sanction); err != nil {
			return nil, err
		}
		emitSanctionEvent(ctx, types.EventTypeSanctionExpired, sanction)
	}
	return due, nil
}

// ProcessSanctionExpiry is the module EndBlocker: it lapses due sanctions and
// re-projects the account state of every affected subject. An expired
// suspension therefore restores the preceding state with no manual action.
func (k Keeper) ProcessSanctionExpiry(ctx sdk.Context) error {
	expired, err := k.ExpireSanctions(ctx)
	if err != nil {
		return err
	}
	if len(expired) == 0 {
		return nil
	}

	// Collect the distinct subjects; the set is only read by key, and the
	// slice is sorted, so the projection order does not depend on map order.
	subjects := make([]string, 0, len(expired))
	seen := make(map[string]bool, len(expired))
	for i := range expired {
		if !seen[expired[i].Subject] {
			seen[expired[i].Subject] = true
			subjects = append(subjects, expired[i].Subject)
		}
	}
	sort.Strings(subjects)

	for _, subject := range subjects {
		addr, err := sdk.AccAddressFromBech32(subject)
		if err != nil {
			continue
		}
		if _, err := k.ProjectAccountState(ctx, addr); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// Projection
// ============================================================================

// EffectiveAccountState returns the account state implied by the subject's
// in-force sanctions. Accounts with no sanctions are Active.
func (k Keeper) EffectiveAccountState(ctx sdk.Context, address sdk.AccAddress) types.AccountState {
	sanctions := k.GetSanctionsForSubject(ctx, address.String())
	if len(sanctions) == 0 {
		return types.AccountStateActive
	}
	state, _, _ := projectSanctionEffects(sanctions)
	return state
}

// ProjectAccountState recomputes and stores the account state implied by the
// subject's sanctions. It is idempotent: a subject whose projected state is
// unchanged is not rewritten.
//
// Unlike SetAccountState, which is the raw administrative/genesis path guarded
// by AccountState.CanTransitionTo, the projection is authoritative and may move
// a terminated account back to active when the terminating sanction is revoked
// or expires. That reactivation is only reachable through a reviewed path
// (ConfirmSanction / ResolveAppeal), which is what the card requires.
func (k Keeper) ProjectAccountState(
	ctx sdk.Context,
	subject sdk.AccAddress,
) (types.AccountState, error) {
	sanctions := k.GetSanctionsForSubject(ctx, subject.String())
	if len(sanctions) == 0 {
		// Never sanctioned: leave any pre-existing administrative state alone.
		return types.AccountStateActive, nil
	}

	target, reason, modifiedBy := projectSanctionEffects(sanctions)

	current, found := k.GetAccountState(ctx, subject)
	if found && current.State == target {
		return target, nil
	}

	if err := k.writeProjectedAccountState(ctx, subject, target, reason, modifiedBy); err != nil {
		return target, err
	}
	return target, nil
}

// projectSanctionEffects reduces a subject's sanctions to the single account
// state they imply. The reduction uses a total order on (severity, ID), so the
// result is independent of the order in which sanctions are visited.
func projectSanctionEffects(sanctions []types.Sanction) (types.AccountState, string, string) {
	bestRank := -1
	bestID := ""
	target := types.AccountStateActive
	reason := "no in-force sanctions"
	modifiedBy := ""

	for i := range sanctions {
		sanction := sanctions[i]
		if !sanction.InForce() {
			continue
		}
		state, ok := sanction.EffectedAccountState()
		if !ok {
			continue
		}
		rank := sanction.Kind.SeverityRank()
		if rank > bestRank || (rank == bestRank && sanction.ID > bestID) {
			bestRank = rank
			bestID = sanction.ID
			target = state
			reason = fmt.Sprintf("%s sanction %s: %s",
				sanction.Kind, sanction.ID, sanction.ReasonCode)
			modifiedBy = sanction.SecondReviewer
			if modifiedBy == "" {
				modifiedBy = sanction.ImposedBy
			}
		}
	}
	return target, reason, modifiedBy
}

// writeProjectedAccountState writes a projected account state, bypassing the
// CanTransitionTo guard because the projection is authoritative. The write is
// idempotent and emits the same event the raw path emits.
func (k Keeper) writeProjectedAccountState(
	ctx sdk.Context,
	subject sdk.AccAddress,
	state types.AccountState,
	reason string,
	modifiedBy string,
) error {
	store := ctx.KVStore(k.skey)
	key := types.AccountStateKey(subject.Bytes())

	previous := types.AccountStateUnspecified
	if bz := store.Get(key); bz != nil {
		var existing rolesv1.AccountStateStore
		k.cdc.MustUnmarshal(bz, &existing)
		previous = safeAccountStateFromUint32(existing.State)
	}

	if previous == state {
		return nil
	}

	record := rolesv1.AccountStateStore{
		State:         uint32(state),
		Reason:        reason,
		ModifiedBy:    modifiedBy,
		ModifiedAt:    ctx.BlockTime().Unix(),
		PreviousState: uint32(previous),
	}
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return ctx.EventManager().EmitTypedEvent(&types.EventAccountStateChanged{
		Address:       subject.String(),
		PreviousState: previous.String(),
		NewState:      state.String(),
		ModifiedBy:    modifiedBy,
		Reason:        reason,
	})
}

// ============================================================================
// Events
// ============================================================================

// emitSanctionEvent emits one sanction lifecycle event with a stable attribute
// set.
func emitSanctionEvent(ctx sdk.Context, eventType string, sanction types.Sanction, extra ...sdk.Attribute) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeySanctionID, sanction.ID),
		sdk.NewAttribute(types.AttributeKeyAddress, sanction.Subject),
		sdk.NewAttribute(types.AttributeKeySanctionKind, sanction.Kind.String()),
		sdk.NewAttribute(types.AttributeKeySanctionScope, sanction.Scope.String()),
		sdk.NewAttribute(types.AttributeKeySanctionStatus, sanction.Status.String()),
		sdk.NewAttribute(types.AttributeKeyImposedBy, sanction.ImposedBy),
	}
	if sanction.SecondReviewer != "" {
		attrs = append(attrs,
			sdk.NewAttribute(types.AttributeKeySecondReviewer, sanction.SecondReviewer))
	}
	if sanction.AppealOf != "" {
		attrs = append(attrs,
			sdk.NewAttribute(types.AttributeKeyAppealOf, sanction.AppealOf))
	}
	if sanction.ExpiresAt > 0 {
		attrs = append(attrs,
			sdk.NewAttribute(types.AttributeKeyExpiresAt, strconv.FormatInt(sanction.ExpiresAt, 10)))
	}
	attrs = append(attrs, extra...)

	ctx.EventManager().EmitEvent(sdk.NewEvent(eventType, attrs...))
}
