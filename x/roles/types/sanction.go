package types

import (
	"fmt"
	"strings"
)

// Sanction record constraints.
//
// A sanction is a scoped, time-limited, appealable action taken against an
// account. Account state is a *projection* of the active sanctions for that
// account rather than a bare enum that an administrator flips.
const (
	// MinJustificationLength is the minimum free-text justification length.
	MinJustificationLength = 20

	// MaxJustificationLength is the maximum free-text justification length.
	MaxJustificationLength = 4000

	// MaxNoticeLength is the maximum length of the notice served on the
	// sanctioned party.
	MaxNoticeLength = 2000

	// MaxEmergencyHoldSeconds is the maximum duration of an emergency hold
	// imposed single-handedly without a second reviewer (72h).
	MaxEmergencyHoldSeconds int64 = 72 * 60 * 60

	// DefaultEmergencyHoldSeconds is the default emergency hold duration (24h).
	DefaultEmergencyHoldSeconds int64 = 24 * 60 * 60
)

// SanctionScope is the breadth of a sanction's effects.
//
// Scope is recorded, not enforced here: limiting a dispute to an order rather
// than the whole account is a separate workstream. This type exists so the
// recorded breadth is an explicit, auditable field rather than an assumption.
type SanctionScope uint8

const (
	// SanctionScopeUnspecified is the default/invalid scope.
	SanctionScopeUnspecified SanctionScope = iota

	// SanctionScopeAccount applies to the account as a whole.
	SanctionScopeAccount

	// SanctionScopeOrder applies to a single order (ScopeRef is the order ID).
	SanctionScopeOrder

	// SanctionScopeListing applies to a single listing (ScopeRef is the listing ID).
	SanctionScopeListing

	// SanctionScopeProviderOfferings applies to a provider's offerings.
	SanctionScopeProviderOfferings
)

// SanctionScopeNames maps sanction scopes to human-readable names.
var SanctionScopeNames = map[SanctionScope]string{
	SanctionScopeUnspecified:       "unspecified",
	SanctionScopeAccount:           "account",
	SanctionScopeOrder:             "order",
	SanctionScopeListing:           "listing",
	SanctionScopeProviderOfferings: "provider_offerings",
}

// String returns the string representation of a SanctionScope.
func (s SanctionScope) String() string {
	if name, ok := SanctionScopeNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true when the scope is a known value.
func (s SanctionScope) IsValid() bool {
	return s >= SanctionScopeAccount && s <= SanctionScopeProviderOfferings
}

// RequiresRef reports whether this scope must carry a scope reference.
func (s SanctionScope) RequiresRef() bool {
	return s == SanctionScopeOrder || s == SanctionScopeListing
}

// AllSanctionScopes returns every valid sanction scope.
func AllSanctionScopes() []SanctionScope {
	return []SanctionScope{
		SanctionScopeAccount,
		SanctionScopeOrder,
		SanctionScopeListing,
		SanctionScopeProviderOfferings,
	}
}

// sanctionScopeNamesOrdered is the canonical, order-stable list of scope names.
// Reverse lookups iterate this slice rather than the map so the result never
// depends on Go's map iteration order.
var sanctionScopeNamesOrdered = []struct {
	scope SanctionScope
	name  string
}{
	{SanctionScopeUnspecified, "unspecified"},
	{SanctionScopeAccount, "account"},
	{SanctionScopeOrder, "order"},
	{SanctionScopeListing, "listing"},
	{SanctionScopeProviderOfferings, "provider_offerings"},
}

// SanctionScopeFromString converts a string to a SanctionScope.
func SanctionScopeFromString(str string) (SanctionScope, error) {
	for _, entry := range sanctionScopeNamesOrdered {
		if entry.name == str {
			return entry.scope, nil
		}
	}
	return SanctionScopeUnspecified, fmt.Errorf("unknown sanction scope: %s", str)
}

// SanctionKind is the severity class of a sanction.
type SanctionKind uint8

const (
	// SanctionKindUnspecified is the default/invalid kind.
	SanctionKindUnspecified SanctionKind = iota

	// SanctionKindWarning is an advisory action with no state effect.
	SanctionKindWarning

	// SanctionKindEmergencyHold is a single-handed interim hold. It MUST
	// expire unless a second, distinct reviewer confirms it within the window.
	SanctionKindEmergencyHold

	// SanctionKindSuspension is a time-limited suspension. It requires a
	// second, distinct reviewer.
	SanctionKindSuspension

	// SanctionKindTermination is an indefinite removal of access. It requires
	// a second, distinct reviewer and remains reversible through review.
	SanctionKindTermination
)

// SanctionKindNames maps sanction kinds to human-readable names.
var SanctionKindNames = map[SanctionKind]string{
	SanctionKindUnspecified:   "unspecified",
	SanctionKindWarning:       "warning",
	SanctionKindEmergencyHold: "emergency_hold",
	SanctionKindSuspension:    "suspension",
	SanctionKindTermination:   "termination",
}

// String returns the string representation of a SanctionKind.
func (k SanctionKind) String() string {
	if name, ok := SanctionKindNames[k]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", k)
}

// IsValid returns true when the kind is a known value.
func (k SanctionKind) IsValid() bool {
	return k >= SanctionKindWarning && k <= SanctionKindTermination
}

// SeverityRank returns the monotonic severity rank of a sanction kind.
// Higher means harsher. Used to decide whether a new sanction escalates, and
// is deliberately a total order so comparisons are deterministic.
func (k SanctionKind) SeverityRank() int {
	switch k {
	case SanctionKindWarning:
		return 10
	case SanctionKindEmergencyHold:
		return 20
	case SanctionKindSuspension:
		return 30
	case SanctionKindTermination:
		return 40
	default:
		return 0
	}
}

// RequiresSecondReviewer reports whether this kind may never be imposed by a
// single actor.
func (k SanctionKind) RequiresSecondReviewer() bool {
	return k == SanctionKindSuspension || k == SanctionKindTermination
}

// RequiresNotice reports whether the sanctioned party must be served a notice.
func (k SanctionKind) RequiresNotice() bool {
	return k == SanctionKindEmergencyHold || k == SanctionKindSuspension ||
		k == SanctionKindTermination
}

// AllSanctionKinds returns every valid sanction kind.
func AllSanctionKinds() []SanctionKind {
	return []SanctionKind{
		SanctionKindWarning,
		SanctionKindEmergencyHold,
		SanctionKindSuspension,
		SanctionKindTermination,
	}
}

// sanctionKindNamesOrdered is the canonical, order-stable list of kind names.
var sanctionKindNamesOrdered = []struct {
	kind SanctionKind
	name string
}{
	{SanctionKindUnspecified, "unspecified"},
	{SanctionKindWarning, "warning"},
	{SanctionKindEmergencyHold, "emergency_hold"},
	{SanctionKindSuspension, "suspension"},
	{SanctionKindTermination, "termination"},
}

// SanctionKindFromString converts a string to a SanctionKind.
func SanctionKindFromString(str string) (SanctionKind, error) {
	for _, entry := range sanctionKindNamesOrdered {
		if entry.name == str {
			return entry.kind, nil
		}
	}
	return SanctionKindUnspecified, fmt.Errorf("unknown sanction kind: %s", str)
}

// SanctionReasonCode is a machine-classifiable reason for a sanction.
type SanctionReasonCode uint8

const (
	// SanctionReasonUnspecified is the default/invalid reason code.
	SanctionReasonUnspecified SanctionReasonCode = iota

	// SanctionReasonFraudConfirmed is confirmed fraudulent activity.
	SanctionReasonFraudConfirmed

	// SanctionReasonFraudSuspected is suspected fraud (the usual basis for an
	// emergency hold).
	SanctionReasonFraudSuspected

	// SanctionReasonPaymentFraud is payment-related fraud.
	SanctionReasonPaymentFraud

	// SanctionReasonIdentityFraud is fake or stolen identity.
	SanctionReasonIdentityFraud

	// SanctionReasonResourceAbuse is abuse of allocated resources.
	SanctionReasonResourceAbuse

	// SanctionReasonTermsViolation is a terms-of-service violation.
	SanctionReasonTermsViolation

	// SanctionReasonSecurityIncident is an active security incident.
	SanctionReasonSecurityIncident

	// SanctionReasonLegalOrder is a binding legal order.
	SanctionReasonLegalOrder
)

// SanctionReasonNames maps reason codes to human-readable names.
var SanctionReasonNames = map[SanctionReasonCode]string{
	SanctionReasonUnspecified:      "unspecified",
	SanctionReasonFraudConfirmed:   "fraud_confirmed",
	SanctionReasonFraudSuspected:   "fraud_suspected",
	SanctionReasonPaymentFraud:     "payment_fraud",
	SanctionReasonIdentityFraud:    "identity_fraud",
	SanctionReasonResourceAbuse:    "resource_abuse",
	SanctionReasonTermsViolation:   "terms_violation",
	SanctionReasonSecurityIncident: "security_incident",
	SanctionReasonLegalOrder:       "legal_order",
}

// String returns the string representation of a SanctionReasonCode.
func (r SanctionReasonCode) String() string {
	if name, ok := SanctionReasonNames[r]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", r)
}

// IsValid returns true when the reason code is a known value.
func (r SanctionReasonCode) IsValid() bool {
	return r >= SanctionReasonFraudConfirmed && r <= SanctionReasonLegalOrder
}

// AllSanctionReasonCodes returns every valid reason code.
func AllSanctionReasonCodes() []SanctionReasonCode {
	return []SanctionReasonCode{
		SanctionReasonFraudConfirmed,
		SanctionReasonFraudSuspected,
		SanctionReasonPaymentFraud,
		SanctionReasonIdentityFraud,
		SanctionReasonResourceAbuse,
		SanctionReasonTermsViolation,
		SanctionReasonSecurityIncident,
		SanctionReasonLegalOrder,
	}
}

// SanctionReasonCodeFromString converts a string to a SanctionReasonCode.
func SanctionReasonCodeFromString(str string) (SanctionReasonCode, error) {
	for code, name := range SanctionReasonNames {
		if name == str {
			return code, nil
		}
	}
	return SanctionReasonUnspecified, fmt.Errorf("unknown sanction reason code: %s", str)
}

// SanctionStatus is the lifecycle state of a sanction record.
type SanctionStatus uint8

const (
	// SanctionStatusUnspecified is the default/invalid status.
	SanctionStatusUnspecified SanctionStatus = iota

	// SanctionStatusActive is an imposed, in-force sanction.
	SanctionStatusActive

	// SanctionStatusPendingReview is an emergency hold that has not yet been
	// confirmed by a second, distinct reviewer. It is in force but expires.
	SanctionStatusPendingReview

	// SanctionStatusExpired is a sanction whose expiry time has passed.
	SanctionStatusExpired

	// SanctionStatusRevoked is a sanction cleared by review or a granted appeal.
	SanctionStatusRevoked

	// SanctionStatusAppealPending marks an open appeal against a sanction.
	SanctionStatusAppealPending

	// SanctionStatusAppealDenied marks an appeal that was reviewed and refused.
	SanctionStatusAppealDenied

	// SanctionStatusAppealGranted marks an appeal that was reviewed and allowed.
	SanctionStatusAppealGranted
)

// SanctionStatusNames maps sanction statuses to human-readable names.
var SanctionStatusNames = map[SanctionStatus]string{
	SanctionStatusUnspecified:   "unspecified",
	SanctionStatusActive:        "active",
	SanctionStatusPendingReview: "pending_review",
	SanctionStatusExpired:       "expired",
	SanctionStatusRevoked:       "revoked",
	SanctionStatusAppealPending: "appeal_pending",
	SanctionStatusAppealDenied:  "appeal_denied",
	SanctionStatusAppealGranted: "appeal_granted",
}

// String returns the string representation of a SanctionStatus.
func (s SanctionStatus) String() string {
	if name, ok := SanctionStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true when the status is a known value.
func (s SanctionStatus) IsValid() bool {
	return s >= SanctionStatusActive && s <= SanctionStatusAppealGranted
}

// InForce reports whether the status, on its own, restricts the subject.
// A pending-review record is only in force for emergency holds, so the
// kind-aware check lives on Sanction.InForce rather than here.
func (s SanctionStatus) InForce() bool {
	return s == SanctionStatusActive
}

// IsAppeal reports whether this status denotes an appeal record.
func (s SanctionStatus) IsAppeal() bool {
	return s == SanctionStatusAppealPending || s == SanctionStatusAppealDenied ||
		s == SanctionStatusAppealGranted
}

// IsTerminal reports whether no further transition is permitted.
func (s SanctionStatus) IsTerminal() bool {
	return s == SanctionStatusExpired || s == SanctionStatusRevoked ||
		s == SanctionStatusAppealDenied || s == SanctionStatusAppealGranted
}

// CanTransitionTo reports whether the status may move to the target status.
// This is a total, side-effect-free function of its inputs only.
func (s SanctionStatus) CanTransitionTo(target SanctionStatus) bool {
	switch s {
	case SanctionStatusActive:
		return target == SanctionStatusExpired || target == SanctionStatusRevoked
	case SanctionStatusPendingReview:
		return target == SanctionStatusActive || target == SanctionStatusExpired ||
			target == SanctionStatusRevoked
	case SanctionStatusAppealPending:
		return target == SanctionStatusAppealGranted || target == SanctionStatusAppealDenied
	case SanctionStatusExpired, SanctionStatusRevoked,
		SanctionStatusAppealDenied, SanctionStatusAppealGranted:
		return false
	default:
		return false
	}
}

// AllSanctionStatuses returns every valid sanction status.
func AllSanctionStatuses() []SanctionStatus {
	return []SanctionStatus{
		SanctionStatusActive,
		SanctionStatusPendingReview,
		SanctionStatusExpired,
		SanctionStatusRevoked,
		SanctionStatusAppealPending,
		SanctionStatusAppealDenied,
		SanctionStatusAppealGranted,
	}
}

// Sanction is a scoped, time-limited, appealable action against an account.
//
// It is stored as JSON in the roles store, mirroring the fraud module's
// FraudReport. There is deliberately no proto twin yet: adding one requires the
// pinned generation image, so exposing sanctions over gRPC is tracked as
// follow-up work rather than guessed at here.
type Sanction struct {
	// ID is the unique identifier for this sanction (or appeal) record.
	ID string `json:"id"`

	// Subject is the account address the sanction applies to.
	Subject string `json:"subject"`

	// Scope is the breadth of the sanction's effects.
	Scope SanctionScope `json:"scope"`

	// ScopeRef is the order/listing ID for ref-bearing scopes.
	ScopeRef string `json:"scope_ref,omitempty"`

	// Kind is the severity class of the sanction.
	Kind SanctionKind `json:"kind"`

	// ReasonCode is the machine-classifiable reason.
	ReasonCode SanctionReasonCode `json:"reason_code"`

	// Justification is the free-text basis for the action.
	Justification string `json:"justification"`

	// Notice is the text served on the sanctioned party, including how to
	// appeal. Required for any sanction with a state effect.
	Notice string `json:"notice,omitempty"`

	// ImposedBy is the account that imposed (or, for appeals, opened) it.
	ImposedBy string `json:"imposed_by"`

	// ImposedAt is the block time (Unix seconds) the record was created.
	ImposedAt int64 `json:"imposed_at"`

	// ExpiresAt is the block time (Unix seconds) at which the sanction lapses.
	// Zero means indefinite (termination only).
	ExpiresAt int64 `json:"expires_at,omitempty"`

	// Status is the lifecycle state of the record.
	Status SanctionStatus `json:"status"`

	// SecondReviewer is the distinct moderator-or-above identity that
	// independently reviewed the action. Required for suspension/termination.
	SecondReviewer string `json:"second_reviewer,omitempty"`

	// SecondReviewedAt is the block time (Unix seconds) of that review.
	SecondReviewedAt int64 `json:"second_reviewed_at,omitempty"`

	// AppealOf is the sanction ID this record appeals, when it is an appeal.
	AppealOf string `json:"appeal_of,omitempty"`

	// BlockHeight is the height at which the record was created.
	BlockHeight int64 `json:"block_height"`
}

// Validate performs internal consistency checks on a sanction record. Address
// validity is enforced by the keeper, which has the SDK available.
func (s Sanction) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return ErrInvalidSanction.Wrap("sanction ID is required")
	}
	if s.IsAppeal() {
		// An appeal record carries the appellant's justification but none of
		// the notice or second-reviewer duties of the sanction it challenges.
		if strings.TrimSpace(s.AppealOf) == "" {
			return ErrInvalidSanction.Wrap("appeal_of is required on an appeal record")
		}
		if !s.Status.IsAppeal() {
			return ErrInvalidSanctionStatus.Wrapf("appeal record has non-appeal status %s", s.Status)
		}
		if s.ImposedAt <= 0 {
			return ErrInvalidSanction.Wrap("imposed_at is required")
		}
		return s.ValidateCommon()
	}
	if err := s.ValidateProposal(); err != nil {
		return err
	}
	if !s.Status.IsValid() {
		return ErrInvalidSanctionStatus.Wrapf("unknown status %d", s.Status)
	}
	if s.ImposedAt <= 0 {
		return ErrInvalidSanction.Wrap("imposed_at is required")
	}
	if s.ExpiresAt < 0 {
		return ErrInvalidSanction.Wrap("expires_at must not be negative")
	}
	if s.ExpiresAt > 0 && s.ExpiresAt <= s.ImposedAt &&
		s.Status != SanctionStatusExpired {
		return ErrInvalidSanction.Wrap("expires_at must be after imposed_at")
	}
	if s.Kind.RequiresSecondReviewer() && s.Status == SanctionStatusActive &&
		strings.TrimSpace(s.SecondReviewer) == "" {
		return ErrSecondReviewerRequired
	}
	if s.Kind == SanctionKindEmergencyHold && s.ExpiresAt == 0 {
		return ErrInvalidSanction.Wrap("an emergency hold must carry an expiry")
	}
	return nil
}

// ValidateProposal validates the fields of a sanction before it is imposed;
// it does not require an ID, status or timestamps, which the keeper assigns.
func (s Sanction) ValidateProposal() error {
	if err := s.ValidateCommon(); err != nil {
		return err
	}
	if !s.Kind.IsValid() {
		return ErrInvalidSanctionKind.Wrapf("unknown kind %d", s.Kind)
	}
	if !s.ReasonCode.IsValid() {
		return ErrInvalidSanctionReason.Wrapf("unknown reason code %d", s.ReasonCode)
	}
	if s.Kind.RequiresNotice() && strings.TrimSpace(s.Notice) == "" {
		return ErrSanctionNoticeRequired.Wrapf("kind %s requires a notice", s.Kind)
	}
	return nil
}

// ValidateAppealProposal validates an appeal record's fields.
//
// An appeal carries the challenged sanction's kind and reason code for context,
// but none of that sanction's duties: it must not be required to serve a notice,
// and it is not the appellant's place to supply a second reviewer.
func (s Sanction) ValidateAppealProposal() error {
	if err := s.ValidateCommon(); err != nil {
		return err
	}
	if !s.Kind.IsValid() {
		return ErrInvalidSanctionKind.Wrapf("unknown kind %d", s.Kind)
	}
	if !s.ReasonCode.IsValid() {
		return ErrInvalidSanctionReason.Wrapf("unknown reason code %d", s.ReasonCode)
	}
	if strings.TrimSpace(s.AppealOf) == "" {
		return ErrInvalidSanction.Wrap("appeal_of is required on an appeal record")
	}
	return nil
}

// ValidateCommon validates the fields shared by sanctions and appeals.
func (s Sanction) ValidateCommon() error {
	if strings.TrimSpace(s.Subject) == "" {
		return ErrInvalidSanction.Wrap("subject is required")
	}
	if !s.Scope.IsValid() {
		return ErrInvalidSanctionScope.Wrapf("unknown scope %d", s.Scope)
	}
	if s.Scope.RequiresRef() && strings.TrimSpace(s.ScopeRef) == "" {
		return ErrInvalidSanction.Wrapf("scope %s requires a scope reference", s.Scope)
	}
	if s.Scope == SanctionScopeAccount && strings.TrimSpace(s.ScopeRef) != "" {
		return ErrInvalidSanction.Wrap("account scope must not carry a scope reference")
	}
	if len(strings.TrimSpace(s.Justification)) < MinJustificationLength {
		return ErrInvalidSanctionJustification.Wrapf(
			"minimum %d characters required", MinJustificationLength)
	}
	if len(s.Justification) > MaxJustificationLength {
		return ErrInvalidSanctionJustification.Wrapf(
			"maximum %d characters allowed", MaxJustificationLength)
	}
	if len(s.Notice) > MaxNoticeLength {
		return ErrInvalidSanction.Wrapf("notice exceeds %d characters", MaxNoticeLength)
	}
	if strings.TrimSpace(s.ImposedBy) == "" {
		return ErrInvalidSanction.Wrap("imposed_by is required")
	}
	return nil
}

// InForce reports whether the sanction currently restricts the subject.
//
// A pending-review record is in force only when it is an emergency hold: an
// interim hold must bind immediately (that is its purpose) but expires unless
// a second, distinct reviewer confirms it. A pending suspension or termination
// deliberately does not bind at all, which is what stops one moderator from
// imposing a suspension or termination alone.
func (s Sanction) InForce() bool {
	switch s.Status {
	case SanctionStatusActive:
		return true
	case SanctionStatusPendingReview:
		return s.Kind == SanctionKindEmergencyHold
	default:
		return false
	}
}

// IsAppeal reports whether this record is an appeal.
func (s Sanction) IsAppeal() bool {
	return strings.TrimSpace(s.AppealOf) != "" || s.Status.IsAppeal()
}

// EffectedAccountState projects the account state implied by this sanction.
// Only in-force sanctions project a state; warnings project nothing.
func (s Sanction) EffectedAccountState() (AccountState, bool) {
	if !s.InForce() {
		return AccountStateUnspecified, false
	}
	switch s.Kind {
	case SanctionKindTermination:
		return AccountStateTerminated, true
	case SanctionKindSuspension, SanctionKindEmergencyHold:
		return AccountStateSuspended, true
	default:
		return AccountStateUnspecified, false
	}
}

// SanctionID renders a sanction ID from a sequence number.
func SanctionID(seq uint64) string {
	return fmt.Sprintf("sanction-%d", seq)
}

// NewSanctionRecord builds a sanction record from a proposal, assigning the
// caller-supplied identity, status and block context.
func NewSanctionRecord(
	id string,
	proposal Sanction,
	status SanctionStatus,
	imposedBy string,
	blockTime int64,
	blockHeight int64,
) Sanction {
	record := proposal
	record.ID = id
	record.Status = status
	record.ImposedBy = imposedBy
	record.ImposedAt = blockTime
	record.BlockHeight = blockHeight
	return record
}
