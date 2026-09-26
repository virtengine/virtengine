// Package marketplace provides types for the marketplace on-chain module.
//
// MARKET-SCOPED-ELIGIBILITY: This file implements the scope discipline that keeps a
// marketplace eligibility decision from reaching further than the thing it is actually
// about.
//
// Design rule (VirtEngine constitution 6.1.3/6.2.3, 39.1-39.2, 40.2 — proportionate,
// scoped controls):
//
//	Declining or failing verification must not by itself suspend a general marketplace
//	account. A dispute about one order should normally affect that order, its escrow or
//	the relevant listing — not every service the person uses. Broad account suspension is
//	reserved for independently established serious abuse, with clear criteria and review.
//
// Concretely:
//   - A dispute (open or resolved) may affect the order, its escrow/invoice and the
//     listing it concerns. It may never, by itself, change global account state.
//   - A failed, absent or locked VEID check is an INPUT to a listing's explicit identity
//     requirement (and to fraud scoring). It is never a stand-alone predicate for the
//     whole marketplace, and it never writes global account state.
//   - Only independently established serious abuse — meeting the documented criteria in
//     SeriousAbuseCriteria AND carrying a reviewed AccountWideSanction — may produce an
//     account-wide effect.
package marketplace

import (
	"fmt"
	"sort"
)

// EffectScope classifies how far a single marketplace decision is permitted to reach.
// Every eligibility decision must declare one, and the scope is the contract: a
// decision that may only touch one order must not be able to reach the whole account.
type EffectScope uint8

const (
	// EffectScopeUnspecified is the zero value and is never a valid decision outcome.
	// Treating "unspecified" as benign would silently widen the blast radius, so callers
	// must reject it rather than default to it.
	EffectScopeUnspecified EffectScope = 0

	// EffectScopeOrder affects a single order and its bidding/settlement lifecycle.
	EffectScopeOrder EffectScope = 1

	// EffectScopeEscrow affects the escrow account, payment and invoice attached to an
	// order (including refunds and billing corrections).
	EffectScopeEscrow EffectScope = 2

	// EffectScopeListing affects one offering/listing — for example deactivating the
	// offer whose explicit identity requirement is no longer met.
	EffectScopeListing EffectScope = 3

	// EffectScopeAccount affects the whole account across every service the person
	// uses. Reserved for independently established serious abuse with review.
	EffectScopeAccount EffectScope = 4
)

// EffectScopeNames maps scopes to human-readable names.
var EffectScopeNames = map[EffectScope]string{
	EffectScopeUnspecified: "unspecified",
	EffectScopeOrder:       "order",
	EffectScopeEscrow:      "escrow",
	EffectScopeListing:     "listing",
	EffectScopeAccount:     "account",
}

// String returns the string representation of an effect scope.
func (s EffectScope) String() string {
	if name, ok := EffectScopeNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid reports whether the scope is one of the defined, actionable scopes.
func (s EffectScope) IsValid() bool {
	return s >= EffectScopeOrder && s <= EffectScopeAccount
}

// IsAccountWide reports whether the scope reaches beyond a single order/listing into
// global account state. This is the predicate every enforcement path must consult before
// writing global state, so that "is this allowed to switch off every service?" is a
// question with one answer in one place.
func (s EffectScope) IsAccountWide() bool {
	return s == EffectScopeAccount
}

// EligibilityTrigger identifies the class of signal being evaluated. Naming the trigger
// explicitly is what allows the policy table below to be audited against the
// constitution, instead of each call site guessing a scope for itself.
type EligibilityTrigger uint8

const (
	// TriggerUnspecified is the zero value and must not be used for a real decision.
	TriggerUnspecified EligibilityTrigger = 0

	// TriggerOrderDisputeOpened is a dispute opened against one order/invoice.
	TriggerOrderDisputeOpened EligibilityTrigger = 1

	// TriggerOrderDisputeResolved is the resolution of such a dispute.
	TriggerOrderDisputeResolved EligibilityTrigger = 2

	// TriggerVEIDCheckFailed is a failed VEID gating check for one order or listing.
	TriggerVEIDCheckFailed EligibilityTrigger = 3

	// TriggerVEIDRecordAbsent is a missing identity record where an offering required one.
	TriggerVEIDRecordAbsent EligibilityTrigger = 4

	// TriggerVEIDIdentityLocked is a locked identity record.
	TriggerVEIDIdentityLocked EligibilityTrigger = 5

	// TriggerListingRequirementUnmet is an unmet listing-scoped requirement
	// (identity score/tier, email/domain verification, MFA) on one offering.
	TriggerListingRequirementUnmet EligibilityTrigger = 6

	// TriggerProviderComplianceIncomplete is incomplete provider compliance evidence.
	TriggerProviderComplianceIncomplete EligibilityTrigger = 7

	// TriggerManipulationDetected is a detected market-manipulation violation. This is
	// the only trigger class that can reach account-wide scope, and only through the
	// criteria + review gate.
	TriggerManipulationDetected EligibilityTrigger = 8
)

// EligibilityTriggerNames maps triggers to human-readable names.
var EligibilityTriggerNames = map[EligibilityTrigger]string{
	TriggerUnspecified:                  "unspecified",
	TriggerOrderDisputeOpened:           "order_dispute_opened",
	TriggerOrderDisputeResolved:         "order_dispute_resolved",
	TriggerVEIDCheckFailed:              "veid_check_failed",
	TriggerVEIDRecordAbsent:             "veid_record_absent",
	TriggerVEIDIdentityLocked:           "veid_identity_locked",
	TriggerListingRequirementUnmet:      "listing_requirement_unmet",
	TriggerProviderComplianceIncomplete: "provider_compliance_incomplete",
	TriggerManipulationDetected:         "manipulation_detected",
}

// String returns the string representation of an eligibility trigger.
func (t EligibilityTrigger) String() string {
	if name, ok := EligibilityTriggerNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid reports whether the trigger is a defined trigger.
func (t EligibilityTrigger) IsValid() bool {
	return t >= TriggerOrderDisputeOpened && t <= TriggerManipulationDetected
}

// MaxScopeFor returns the most severe effect a trigger may cause on its own, with no
// further authorization.
//
// This table is the enforcement of the operator's design rule and is deliberately narrow:
// only TriggerManipulationDetected returns EffectScopeAccount, and even then
// ApplyEligibilityEffect still requires a reviewed AccountWideSanction before that scope
// is actually granted. Disputes and VEID signals are capped at the order/escrow/listing
// they are genuinely about.
func MaxScopeFor(trigger EligibilityTrigger) EffectScope {
	switch trigger {
	case TriggerOrderDisputeOpened, TriggerOrderDisputeResolved:
		// A dispute is about one order: its escrow/invoice and the listing it concerns.
		// Never the whole account.
		return EffectScopeEscrow
	case TriggerVEIDCheckFailed, TriggerVEIDRecordAbsent, TriggerVEIDIdentityLocked:
		// A failed/absent/locked identity check blocks the order being placed (or the
		// listing that declared the requirement). It is not an account verdict.
		return EffectScopeOrder
	case TriggerListingRequirementUnmet:
		// The explicit requirement belongs to the listing, so the listing is the ceiling.
		return EffectScopeListing
	case TriggerProviderComplianceIncomplete:
		// Provider compliance gates the provider's listings, not the account globally.
		return EffectScopeListing
	case TriggerManipulationDetected:
		return EffectScopeAccount
	default:
		return EffectScopeUnspecified
	}
}

// SeriousAbuseCriteria defines the independently established conditions that permit
// account-wide enforcement. Every field is explicit so a reviewer can see exactly what
// justified switching off an account.
type SeriousAbuseCriteria struct {
	// MinViolationSeverity is the per-violation severity floor (0-10) for a violation to
	// count toward serious abuse.
	MinViolationSeverity uint8 `json:"min_violation_severity"`

	// MinQualifyingViolations is the number of qualifying violations required.
	MinQualifyingViolations uint32 `json:"min_qualifying_violations"`

	// MinDistinctViolationTypes is the number of distinct manipulation types required.
	MinDistinctViolationTypes uint32 `json:"min_distinct_violation_types"`

	// AccountWideTypes restricts which manipulation types may ever scale to account-wide
	// enforcement. A violation of a type outside this list is handled at order/listing
	// scope and can never accumulate into an account suspension.
	AccountWideTypes []ManipulationType `json:"account_wide_types"`
}

// DefaultSeriousAbuseCriteria returns the criteria used when a module has not been
// configured otherwise.
//
// The defaults require genuine, repeated, high-severity abuse: at least three qualifying
// violations whose individual severity is >= 7 (wash trading, price manipulation, front
// running, sybil attack). Low-severity noise (order spamming at severity 4, cancellation
// behaviour at 6) can never accumulate into an account suspension under these defaults.
func DefaultSeriousAbuseCriteria() SeriousAbuseCriteria {
	types := make([]ManipulationType, 0, len(ManipulationTypeSeverity))
	for t, severity := range ManipulationTypeSeverity {
		if t == ManipulationTypeNone {
			continue
		}
		if severity >= 7 {
			types = append(types, t)
		}
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })

	return SeriousAbuseCriteria{
		MinViolationSeverity:      7,
		MinQualifyingViolations:   3,
		MinDistinctViolationTypes: 1,
		AccountWideTypes:          types,
	}
}

// Validate validates the criteria so a misconfiguration cannot silently widen enforcement.
func (c SeriousAbuseCriteria) Validate() error {
	if c.MinViolationSeverity > 10 {
		return fmt.Errorf("min_violation_severity cannot exceed 10")
	}
	if c.MinQualifyingViolations == 0 {
		return fmt.Errorf("min_qualifying_violations must be positive")
	}
	if c.MinDistinctViolationTypes == 0 {
		return fmt.Errorf("min_distinct_violation_types must be positive")
	}
	if len(c.AccountWideTypes) == 0 {
		return fmt.Errorf("account_wide_types must name at least one manipulation type")
	}
	for _, t := range c.AccountWideTypes {
		if t == ManipulationTypeNone {
			return fmt.Errorf("account_wide_types must not include the none type")
		}
		if !t.isDefined() {
			return fmt.Errorf("account_wide_types contains undefined type %d", uint8(t))
		}
	}
	return nil
}

// isDefined reports whether the manipulation type is a known type.
func (t ManipulationType) isDefined() bool {
	_, ok := ManipulationTypeSeverity[t]
	return ok
}

// AbuseAssessment is the auditable verdict of evaluating recorded violations against the
// serious-abuse criteria.
type AbuseAssessment struct {
	// CriteriaMet reports whether the criteria were satisfied.
	CriteriaMet bool `json:"criteria_met"`

	// QualifyingViolations is the count of violations that met the severity floor and the
	// account-wide type allowlist.
	QualifyingViolations uint32 `json:"qualifying_violations"`

	// DistinctTypes is the number of distinct manipulation types among qualifying
	// violations.
	DistinctTypes uint32 `json:"distinct_types"`

	// HighestSeverity is the highest severity observed among qualifying violations.
	HighestSeverity uint8 `json:"highest_severity"`

	// Reasons explains, in review-auditable terms, why the criteria were or were not met.
	Reasons []string `json:"reasons"`
}

// AssessSeriousAbuse evaluates recorded violations against the criteria.
//
// This function is the single definition of "serious abuse" in the marketplace module.
// It is pure: it reads violations and returns a verdict, so it can be asserted on
// directly in tests and cited verbatim in a review.
func AssessSeriousAbuse(violations []ViolationRecord, criteria SeriousAbuseCriteria) AbuseAssessment {
	assessment := AbuseAssessment{
		// Reasons is always non-empty so the caller always has something to record.
		Reasons: make([]string, 0, 3),
	}

	if err := criteria.Validate(); err != nil {
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("criteria invalid: %s", err))
		return assessment
	}

	allowed := make(map[ManipulationType]bool, len(criteria.AccountWideTypes))
	for _, t := range criteria.AccountWideTypes {
		allowed[t] = true
	}

	distinct := make(map[ManipulationType]bool)
	for _, v := range violations {
		// A violation whose type is outside the allowlist is handled at order/listing
		// scope and can never accumulate into account-wide enforcement.
		if !allowed[v.Type] {
			continue
		}
		severity := v.Severity
		if severity == 0 {
			// Fall back to the type's canonical severity when the record did not set one,
			// rather than treating an unset field as "not serious".
			severity = v.Type.Severity()
		}
		if severity < criteria.MinViolationSeverity {
			continue
		}

		assessment.QualifyingViolations++
		distinct[v.Type] = true
		if severity > assessment.HighestSeverity {
			assessment.HighestSeverity = severity
		}
	}

	assessment.DistinctTypes = uint32(len(distinct))

	if assessment.QualifyingViolations < criteria.MinQualifyingViolations {
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf(
			"qualifying violations %d below required %d",
			assessment.QualifyingViolations, criteria.MinQualifyingViolations))
	}
	if assessment.DistinctTypes < criteria.MinDistinctViolationTypes {
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf(
			"distinct violation types %d below required %d",
			assessment.DistinctTypes, criteria.MinDistinctViolationTypes))
	}

	assessment.CriteriaMet = assessment.QualifyingViolations >= criteria.MinQualifyingViolations &&
		assessment.DistinctTypes >= criteria.MinDistinctViolationTypes

	if assessment.CriteriaMet {
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf(
			"serious abuse established: %d qualifying violations across %d type(s), highest severity %d",
			assessment.QualifyingViolations, assessment.DistinctTypes, assessment.HighestSeverity))
	}

	return assessment
}

// AccountWideSanction is the reviewed authorization required before any account-wide
// marketplace effect may be applied.
//
// It exists so that "we suspended this account everywhere" is always backed by an
// assessment that met the criteria and a named reviewer — never by a bare counter.
type AccountWideSanction struct {
	// Assessment is the criteria verdict that justified the scope.
	Assessment AbuseAssessment `json:"assessment"`

	// ReviewedBy is the address of the reviewer who authorized the account-wide effect.
	ReviewedBy string `json:"reviewed_by"`

	// ReviewReason is the reviewer's justification, recorded for audit.
	ReviewReason string `json:"review_reason"`

	// CaseID references the reviewed case this sanction belongs to.
	CaseID string `json:"case_id"`
}

// Validate validates that the sanction genuinely authorizes account-wide action.
func (s *AccountWideSanction) Validate() error {
	if s == nil {
		return fmt.Errorf("account-wide sanction is required for an account-wide effect")
	}
	if !s.Assessment.CriteriaMet {
		return fmt.Errorf("account-wide sanction rejected: serious-abuse criteria were not met")
	}
	if s.ReviewedBy == "" {
		return fmt.Errorf("account-wide sanction requires a reviewer")
	}
	if s.ReviewReason == "" {
		return fmt.Errorf("account-wide sanction requires a review reason")
	}
	if s.CaseID == "" {
		return fmt.Errorf("account-wide sanction requires a case id")
	}
	return nil
}

// ApplyEligibilityEffect decides the strongest effect scope permitted for a trigger.
//
// Rules:
//   - An unknown trigger is rejected rather than defaulted.
//   - For every trigger except manipulation detection, the scope is capped at
//     MaxScopeFor(trigger) and supplying a sanction is an error (you may not buy a wider
//     scope for a dispute by attaching paperwork to it).
//   - For manipulation detection, account-wide scope is granted only when the sanction is
//     present and validates; otherwise the effect is capped at order scope, where the
//     manipulation was actually observed.
func ApplyEligibilityEffect(trigger EligibilityTrigger, sanction *AccountWideSanction) (EffectScope, error) {
	if !trigger.IsValid() {
		return EffectScopeUnspecified, fmt.Errorf("unspecified or unknown eligibility trigger %d", uint8(trigger))
	}

	maxScope := MaxScopeFor(trigger)

	if !maxScope.IsAccountWide() {
		if sanction != nil {
			return EffectScopeUnspecified, fmt.Errorf(
				"trigger %s may affect at most %s scope; an account-wide sanction is not applicable",
				trigger, maxScope)
		}
		return maxScope, nil
	}

	// Manipulation detection: the only trigger that may reach account scope.
	if err := sanction.Validate(); err != nil {
		// Not an error: an unreviewed manipulation report is still a legitimate
		// order/listing-scoped action. It simply may not reach account scope.
		return EffectScopeOrder, nil
	}

	return EffectScopeAccount, nil
}
