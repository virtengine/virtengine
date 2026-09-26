package marketplace

import (
	"testing"
	"time"
)

// TestMaxScopeFor_DecisionMap pins the eligibility map: which signals may reach how far.
// A dispute and a failed identity check must be capped at the order/escrow/listing they
// are actually about; only detected manipulation may reach account scope.
func TestMaxScopeFor_DecisionMap(t *testing.T) {
	tests := []struct {
		name    string
		trigger EligibilityTrigger
		want    EffectScope
	}{
		{"dispute opened affects its escrow/order", TriggerOrderDisputeOpened, EffectScopeEscrow},
		{"dispute resolved affects its escrow/order", TriggerOrderDisputeResolved, EffectScopeEscrow},
		{"failed veid check affects the order", TriggerVEIDCheckFailed, EffectScopeOrder},
		{"absent veid record affects the order", TriggerVEIDRecordAbsent, EffectScopeOrder},
		{"locked identity affects the order", TriggerVEIDIdentityLocked, EffectScopeOrder},
		{"unmet listing requirement affects the listing", TriggerListingRequirementUnmet, EffectScopeListing},
		{"provider compliance affects the listings", TriggerProviderComplianceIncomplete, EffectScopeListing},
		{"manipulation may reach account scope", TriggerManipulationDetected, EffectScopeAccount},
		{"unspecified trigger has no scope", TriggerUnspecified, EffectScopeUnspecified},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MaxScopeFor(tc.trigger); got != tc.want {
				t.Errorf("MaxScopeFor(%s) = %s, want %s", tc.trigger, got, tc.want)
			}
		})
	}
}

// TestNonAbuseTriggersNeverReachAccountScope is the direct assertion of the operator's
// design rule: neither a dispute nor a VEID signal may, on its own, affect global account
// state.
func TestNonAbuseTriggersNeverReachAccountScope(t *testing.T) {
	for _, trigger := range []EligibilityTrigger{
		TriggerOrderDisputeOpened,
		TriggerOrderDisputeResolved,
		TriggerVEIDCheckFailed,
		TriggerVEIDRecordAbsent,
		TriggerVEIDIdentityLocked,
		TriggerListingRequirementUnmet,
		TriggerProviderComplianceIncomplete,
	} {
		scope := MaxScopeFor(trigger)
		if scope.IsAccountWide() {
			t.Errorf("trigger %s must not be account-wide, got %s", trigger, scope)
		}

		// Even a fully valid, reviewed sanction may not widen these triggers: you cannot
		// buy a wider blast radius for a dispute by attaching paperwork to it.
		sanction := validSanction()
		resolved, err := ApplyEligibilityEffect(trigger, &sanction)
		if err == nil {
			t.Errorf("trigger %s must reject an account-wide sanction, got scope %s", trigger, resolved)
		}
		if resolved.IsAccountWide() {
			t.Errorf("trigger %s resolved to account-wide scope %s", trigger, resolved)
		}
	}
}

// TestApplyEligibilityEffect_RejectsUnknownTrigger ensures an unset/unknown trigger is
// rejected rather than silently treated as harmless or as maximally broad.
func TestApplyEligibilityEffect_RejectsUnknownTrigger(t *testing.T) {
	for _, trigger := range []EligibilityTrigger{TriggerUnspecified, EligibilityTrigger(200)} {
		scope, err := ApplyEligibilityEffect(trigger, nil)
		if err == nil {
			t.Errorf("trigger %d should be rejected", uint8(trigger))
		}
		if scope != EffectScopeUnspecified {
			t.Errorf("trigger %d should resolve to unspecified, got %s", uint8(trigger), scope)
		}
	}
}

// TestApplyEligibilityEffect_ManipulationRequiresReview is DONE-WHEN criterion 3: an
// account-wide effect is granted only when the criteria are met AND a named reviewer
// authorized it.
func TestApplyEligibilityEffect_ManipulationRequiresReview(t *testing.T) {
	t.Run("no sanction at all stays order-scoped", func(t *testing.T) {
		scope, err := ApplyEligibilityEffect(TriggerManipulationDetected, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsAccountWide() {
			t.Errorf("unreviewed manipulation must not be account-wide, got %s", scope)
		}
		if scope != EffectScopeOrder {
			t.Errorf("unreviewed manipulation should stay at order scope, got %s", scope)
		}
	})

	t.Run("criteria not met stays order-scoped", func(t *testing.T) {
		s := validSanction()
		s.Assessment.CriteriaMet = false
		scope, err := ApplyEligibilityEffect(TriggerManipulationDetected, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsAccountWide() {
			t.Errorf("criteria-failing sanction must not be account-wide, got %s", scope)
		}
	})

	t.Run("missing reviewer stays order-scoped", func(t *testing.T) {
		s := validSanction()
		s.ReviewedBy = ""
		scope, err := ApplyEligibilityEffect(TriggerManipulationDetected, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsAccountWide() {
			t.Errorf("unreviewed sanction must not be account-wide, got %s", scope)
		}
	})

	t.Run("valid reviewed sanction grants account scope", func(t *testing.T) {
		s := validSanction()
		scope, err := ApplyEligibilityEffect(TriggerManipulationDetected, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope != EffectScopeAccount {
			t.Errorf("reviewed sanction meeting criteria should grant account scope, got %s", scope)
		}
	})
}

// TestAssessSeriousAbuse exercises the documented criteria that permit account-wide
// action. Low-severity or non-allowlisted noise must never accumulate into a suspension.
func TestAssessSeriousAbuse(t *testing.T) {
	criteria := DefaultSeriousAbuseCriteria()
	now := time.Unix(1700000000, 0).UTC()

	mk := func(violations int, mt ManipulationType, severity uint8) []ViolationRecord {
		out := make([]ViolationRecord, 0, violations)
		for i := 0; i < violations; i++ {
			out = append(out, ViolationRecord{
				Address:     "acct",
				Type:        mt,
				Severity:    severity,
				DetectedAt:  now,
				BlockHeight: int64(i),
			})
		}
		return out
	}

	t.Run("no violations is not serious abuse", func(t *testing.T) {
		a := AssessSeriousAbuse(nil, criteria)
		if a.CriteriaMet {
			t.Error("no violations must not establish serious abuse")
		}
		if len(a.Reasons) == 0 {
			t.Error("assessment must always explain itself")
		}
	})

	t.Run("low severity of an allowlisted type is not enough", func(t *testing.T) {
		// SybilAttack is allowlisted but severity 3 is below the floor of 7.
		a := AssessSeriousAbuse(mk(100, ManipulationTypeSybilAttack, 3), criteria)
		if a.CriteriaMet {
			t.Errorf("sub-threshold severity must not establish serious abuse: %v", a.Reasons)
		}
		if a.QualifyingViolations != 0 {
			t.Errorf("expected 0 qualifying violations, got %d", a.QualifyingViolations)
		}
	})

	t.Run("high severity but too few violations is not enough", func(t *testing.T) {
		a := AssessSeriousAbuse(mk(2, ManipulationTypeSybilAttack, 10), criteria)
		if a.CriteriaMet {
			t.Errorf("two violations must not establish serious abuse: %v", a.Reasons)
		}
	})

	t.Run("repeated high severity abuse is established", func(t *testing.T) {
		a := AssessSeriousAbuse(mk(3, ManipulationTypeSybilAttack, 10), criteria)
		if !a.CriteriaMet {
			t.Fatalf("expected criteria met, got %v", a.Reasons)
		}
		if a.QualifyingViolations != 3 {
			t.Errorf("expected 3 qualifying violations, got %d", a.QualifyingViolations)
		}
		if a.HighestSeverity != 10 {
			t.Errorf("expected highest severity 10, got %d", a.HighestSeverity)
		}
	})

	t.Run("non-allowlisted type never accumulates toward account scope", func(t *testing.T) {
		// OrderSpamming is well outside the account-wide allowlist, so 500 of them still
		// cannot switch off the account.
		a := AssessSeriousAbuse(mk(500, ManipulationTypeOrderSpamming, 10), criteria)
		if a.CriteriaMet {
			t.Errorf("non-allowlisted type must not establish serious abuse: %v", a.Reasons)
		}
	})

	t.Run("severity falls back to the canonical type severity", func(t *testing.T) {
		// A record that did not set Severity must not be read as "not serious" when its
		// type is by definition severe.
		recs := mk(3, ManipulationTypeSybilAttack, 0)
		a := AssessSeriousAbuse(recs, criteria)
		if !a.CriteriaMet {
			t.Errorf("expected canonical severity fallback to establish abuse: %v", a.Reasons)
		}
	})

	t.Run("invalid criteria is reported, never silently satisfied", func(t *testing.T) {
		bad := criteria
		bad.AccountWideTypes = nil
		a := AssessSeriousAbuse(mk(50, ManipulationTypeSybilAttack, 10), bad)
		if a.CriteriaMet {
			t.Error("invalid criteria must not be treated as satisfied")
		}
	})
}

// TestApplyAccountWideSanction_Gate proves the single enforcement gate: account-wide
// state is mutated only for a criteria-meeting, reviewed sanction.
func TestApplyAccountWideSanction_Gate(t *testing.T) {
	config := DefaultPenaltyConfig()
	now := time.Unix(1700000000, 0).UTC()

	// Build an account that genuinely meets the serious-abuse criteria.
	abusive := NewAccountSafeguardState("acct")
	for i := 0; i < 3; i++ {
		abusive.RecordViolation(ViolationRecord{
			Address:    "acct",
			Type:       ManipulationTypeSybilAttack,
			Severity:   10,
			DetectedAt: now,
		}, config)
	}

	t.Run("nil sanction is refused and mutates nothing", func(t *testing.T) {
		s := NewAccountSafeguardState("acct")
		_, err := s.ApplyAccountWideSanction(nil, config, now)
		if err == nil {
			t.Error("expected nil sanction to be refused")
		}
		if s.IsSuspended || s.IsBanned {
			t.Error("refused sanction must not mutate account state")
		}
	})

	t.Run("criteria-failing sanction is refused and mutates nothing", func(t *testing.T) {
		s := NewAccountSafeguardState("acct")
		sanction := validSanction()
		sanction.Assessment.CriteriaMet = false
		_, err := s.ApplyAccountWideSanction(&sanction, config, now)
		if err == nil {
			t.Error("expected criteria-failing sanction to be refused")
		}
		if s.IsSuspended || s.IsBanned {
			t.Error("refused sanction must not mutate account state")
		}
	})

	t.Run("unreviewed sanction is refused and mutates nothing", func(t *testing.T) {
		s := NewAccountSafeguardState("acct")
		sanction := validSanction()
		sanction.ReviewedBy = ""
		_, err := s.ApplyAccountWideSanction(&sanction, config, now)
		if err == nil {
			t.Error("expected unreviewed sanction to be refused")
		}
		if s.IsSuspended || s.IsBanned {
			t.Error("refused sanction must not mutate account state")
		}
	})

	t.Run("valid reviewed sanction applies", func(t *testing.T) {
		assessment := abusive.AssessAbuse(DefaultSeriousAbuseCriteria())
		if !assessment.CriteriaMet {
			t.Fatalf("fixture should meet criteria: %v", assessment.Reasons)
		}

		sanction := AccountWideSanction{
			Assessment:   assessment,
			ReviewedBy:   "moderator-1",
			ReviewReason: "sustained sybil activity across distinct orders",
			CaseID:       "case-1",
		}

		// Drive the ban branch explicitly: the default ban threshold is 20, so a 3-violation
		// fixture would otherwise (correctly) land on the suspension branch below.
		banConfig := config
		banConfig.BanThreshold = 3

		action, err := abusive.ApplyAccountWideSanction(&sanction, banConfig, now)
		if err != nil {
			t.Fatalf("expected sanction to apply: %v", err)
		}
		if action != PenaltyActionBan {
			t.Errorf("expected ban recommendation, got %s", action)
		}
		if !abusive.IsBanned {
			t.Error("account meeting criteria with review should be banned")
		}
		if abusive.BannedAt == nil {
			t.Error("banned account must record when it was banned")
		}
	})

	t.Run("suspension path is time-bounded", func(t *testing.T) {
		// With the default ban threshold (20) the same reviewed sanction produces a
		// time-bounded suspension rather than a permanent ban.
		s := NewAccountSafeguardState("acct")
		for i := 0; i < 3; i++ {
			s.RecordViolation(ViolationRecord{
				Address: "acct", Type: ManipulationTypeSybilAttack, Severity: 10, DetectedAt: now,
			}, config)
		}

		sanction := validSanction()
		action, err := s.ApplyAccountWideSanction(&sanction, config, now)
		if err != nil {
			t.Fatalf("expected suspension to apply: %v", err)
		}
		if action != PenaltyActionSuspension {
			t.Errorf("expected suspension action, got %s", action)
		}
		if !s.IsSuspended {
			t.Error("expected account to be suspended")
		}
		if s.SuspendedUntil == nil || !s.SuspendedUntil.After(now) {
			t.Error("suspension must be time-bounded in the future")
		}
	})
}

// TestRecordViolationDoesNotSuspendAccount is DONE-WHEN criterion 2's second half: a
// recorded violation (including identity-derived ones) never sets global account state by
// itself, no matter how many are recorded.
func TestRecordViolationDoesNotSuspendAccount(t *testing.T) {
	config := DefaultPenaltyConfig()
	now := time.Unix(1700000000, 0).UTC()
	s := NewAccountSafeguardState("acct")

	var lastAction PenaltyAction
	for i := 0; i < 50; i++ {
		lastAction = s.RecordViolation(ViolationRecord{
			Address:    "acct",
			Type:       ManipulationTypeSybilAttack,
			Severity:   10,
			DetectedAt: now,
		}, config)

		if s.IsSuspended {
			t.Fatalf("violation %d suspended the account without criteria + review", i)
		}
		if s.IsBanned {
			t.Fatalf("violation %d banned the account without criteria + review", i)
		}
	}

	// The recommendation still escalates so the reviewed path has a signal to act on.
	if lastAction != PenaltyActionBan {
		t.Errorf("expected ban to be recommended, got %s", lastAction)
	}
}

// TestSeriousAbuseCriteriaValidate guards against a misconfiguration silently widening or
// disabling account-wide enforcement.
func TestSeriousAbuseCriteriaValidate(t *testing.T) {
	valid := DefaultSeriousAbuseCriteria()
	if err := valid.Validate(); err != nil {
		t.Fatalf("default criteria must be valid: %v", err)
	}

	mods := map[string]func(c *SeriousAbuseCriteria){
		"severity above 10":   func(c *SeriousAbuseCriteria) { c.MinViolationSeverity = 11 },
		"zero violations":     func(c *SeriousAbuseCriteria) { c.MinQualifyingViolations = 0 },
		"zero distinct types": func(c *SeriousAbuseCriteria) { c.MinDistinctViolationTypes = 0 },
		"empty allowlist":     func(c *SeriousAbuseCriteria) { c.AccountWideTypes = nil },
		"allowlist with none": func(c *SeriousAbuseCriteria) { c.AccountWideTypes = []ManipulationType{ManipulationTypeNone} },
		"allowlist undefined": func(c *SeriousAbuseCriteria) { c.AccountWideTypes = []ManipulationType{ManipulationType(99)} },
	}

	for name, mutate := range mods {
		t.Run(name, func(t *testing.T) {
			c := DefaultSeriousAbuseCriteria()
			mutate(&c)
			if err := c.Validate(); err == nil {
				t.Error("expected criteria to be rejected")
			}
		})
	}
}

// TestEffectScopeHelpers covers the small predicates the enforcement paths rely on.
func TestEffectScopeHelpers(t *testing.T) {
	for scope, name := range EffectScopeNames {
		if scope.String() != name {
			t.Errorf("String() mismatch for %d: %s != %s", scope, scope.String(), name)
		}
	}
	if EffectScope(200).String() != "unknown(200)" {
		t.Errorf("unexpected unknown scope string: %s", EffectScope(200).String())
	}
	if !EffectScopeAccount.IsAccountWide() {
		t.Error("account scope must report account-wide")
	}
	for _, s := range []EffectScope{EffectScopeOrder, EffectScopeEscrow, EffectScopeListing} {
		if s.IsAccountWide() {
			t.Errorf("%s must not report account-wide", s)
		}
		if !s.IsValid() {
			t.Errorf("%s must be valid", s)
		}
	}
	if EffectScopeUnspecified.IsValid() {
		t.Error("unspecified scope must not be valid")
	}
}

// validSanction returns a sanction that passes Validate, for use in negative tests where
// a different field is the one under test.
func validSanction() AccountWideSanction {
	return AccountWideSanction{
		Assessment: AbuseAssessment{
			CriteriaMet:          true,
			QualifyingViolations: 3,
			DistinctTypes:        1,
			HighestSeverity:      10,
			Reasons:              []string{"test fixture"},
		},
		ReviewedBy:   "reviewer",
		ReviewReason: "test fixture reason",
		CaseID:       "case-fixture",
	}
}
