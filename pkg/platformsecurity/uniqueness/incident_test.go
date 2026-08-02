package uniqueness

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

var incidentKinds = []BiometricIncidentKind{
	IncidentTemplateIndexCompromise, IncidentTransformKeyCompromise, IncidentMassFalseMatches,
	IncidentUnlawfulProcessing, IncidentModelPoisoning, IncidentDedupEnumeration,
}

func incidentFixture(t *testing.T, kind BiometricIncidentKind) (BiometricIncident, CustodyNodeSet, map[string]ed25519.PrivateKey) {
	t.Helper()
	nodes, keys := custodyFixture(t)
	setDigest, err := nodes.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return BiometricIncident{
		Version: Version1, IncidentCommitment: digestFixture("incident-1"), Kind: kind, ScopeCommitment: digestFixture("scope"),
		EvidenceDigest: digestFixture("incident evidence"), PolicyDigest: digestFixture("incident policy"), ParticipantSetDigest: setDigest,
		Threshold: nodes.Threshold, AuthorityKeyEpoch: 7, AffectedKeyEpoch: 7, AffectedTransformEpoch: 9,
		AffectedModelCommitment: digestFixture("affected model"), DetectedCoordinate: 100, FreezeCoordinate: 101,
		State: IncidentNotificationPending,
	}, nodes, keys
}

func completeRecoveryActions(kind BiometricIncidentKind) RecoveryActions {
	switch kind {
	case IncidentTemplateIndexCompromise:
		return RecoveryActions{RotateKey: true, RotateTransform: true, RebuildIndex: true, ReenrollmentRequired: true}
	case IncidentTransformKeyCompromise:
		return RecoveryActions{RotateKey: true, RotateTransform: true, ReenrollmentRequired: true}
	case IncidentMassFalseMatches, IncidentModelPoisoning:
		return RecoveryActions{RebuildIndex: true, SuspendMatching: true, RevokeModel: true}
	case IncidentUnlawfulProcessing:
		return RecoveryActions{CeaseProcessing: true, DeleteAffectedData: true}
	case IncidentDedupEnumeration:
		return RecoveryActions{RotateKey: true, RotateTransform: true, EnumerationMitigation: true, ReenrollmentRequired: true}
	default:
		return RecoveryActions{}
	}
}

func actionEvidenceFixture(actions RecoveryActions) []RecoveryActionEvidence {
	names := actions.enabledNames()
	evidence := make([]RecoveryActionEvidence, 0, len(names))
	for _, name := range names {
		evidence = append(evidence, RecoveryActionEvidence{Action: name, EvidenceDigest: digestFixture(name + " completion")})
	}
	return evidence
}

func signIncidentValue(t *testing.T, value []byte, epoch uint64, nodes CustodyNodeSet, keys map[string]ed25519.PrivateKey) []NodeSignature {
	t.Helper()
	signatures := make([]NodeSignature, 0, nodes.Threshold)
	for _, node := range nodes.Nodes[:nodes.Threshold] {
		signatures = append(signatures, NodeSignature{NodeID: node.NodeID, SigningKeyID: node.SigningKeyID, SigningKeyEpoch: epoch, Signature: ed25519.Sign(keys[node.SigningKeyID], value)})
	}
	return signatures
}

func planFixture(t *testing.T, incident BiometricIncident, nodes CustodyNodeSet, keys map[string]ed25519.PrivateKey) (BiometricRecoveryPlan, *VerifiedBiometricRecoveryPlan) {
	t.Helper()
	contained := incident
	contained.State = IncidentContained
	incidentDigest, err := contained.Digest()
	if err != nil {
		t.Fatal(err)
	}
	plan := BiometricRecoveryPlan{
		Version: Version1, IncidentDigest: incidentDigest, Kind: incident.Kind, Actions: completeRecoveryActions(incident.Kind),
		NewKeyEpoch: 8, NewTransformEpoch: 10, ReplacementModelDigest: digestFixture("replacement model"),
		ParticipantSetDigest: incident.ParticipantSetDigest, Threshold: incident.Threshold, AuthorityKeyEpoch: incident.AuthorityKeyEpoch,
		FixtureState: FixtureOnlyState, ExternalReviewBlock: ExternalReviewRequired,
	}
	value, err := plan.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	plan.Approvals = signIncidentValue(t, value, incident.AuthorityKeyEpoch, nodes, keys)
	verified, err := VerifyBiometricRecoveryPlan(contained, plan, nodes)
	if err != nil {
		t.Fatal(err)
	}
	return plan, verified
}

func auditFixture(t *testing.T, incidentDigest string) []IncidentAuditEntry {
	t.Helper()
	states := []BiometricIncidentState{IncidentContained, IncidentRecoveryApproved, IncidentRecovering, IncidentNotificationPending, IncidentClosed}
	actions := []string{"incident_contained", "recovery_approved", "recovery_started", "notification_pending", "incident_closed"}
	coordinates := []uint64{102, 103, 104, 105, 110}
	entries := make([]IncidentAuditEntry, 0, len(states))
	previous := digest(nil)
	for index := range states {
		entry := IncidentAuditEntry{Version: Version1, IncidentDigest: incidentDigest, Sequence: uint64(index + 1), PreviousDigest: previous,
			State: states[index], Action: actions[index], ActorCommitment: digestFixture("incident actor"),
			EvidenceDigest: digestFixture(actions[index] + " evidence"), Coordinate: coordinates[index]}
		var err error
		previous, err = entry.Digest()
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func resolutionFixture(t *testing.T, incident BiometricIncident, nodes CustodyNodeSet, keys map[string]ed25519.PrivateKey) (BiometricIncidentResolution, *VerifiedBiometricRecoveryPlan, []IncidentAuditEntry, []IncidentNotification) {
	t.Helper()
	plan, verifiedPlan := planFixture(t, incident, nodes, keys)
	incidentDigest, _ := incident.Digest()
	planDigest, _ := plan.Digest()
	audit := auditFixture(t, incidentDigest)
	auditTail, _ := audit[len(audit)-1].Digest()
	notifications := []IncidentNotification{
		{Version: Version1, IncidentDigest: incidentDigest, Audience: NotificationUsers, RecipientCommitment: digestFixture("affected users"), PolicyDigest: incident.PolicyDigest, ContentDigest: digestFixture("user notice"), DeadlineCoordinate: 115, CompletedCoordinate: 108, State: NotificationDispatched},
		{Version: Version1, IncidentDigest: incidentDigest, Audience: NotificationRegulators, RecipientCommitment: digestFixture("regulator"), PolicyDigest: incident.PolicyDigest, ContentDigest: digestFixture("regulator notice"), DeadlineCoordinate: 112, CompletedCoordinate: 109, State: NotificationAcknowledged},
	}
	notificationDigest, err := incidentNotificationDigest(notifications, incidentDigest, audit[3].Coordinate, 110)
	if err != nil {
		t.Fatal(err)
	}
	reenrollmentEvidence, deletionReceipt := digest(nil), digest(nil)
	if plan.Actions.ReenrollmentRequired {
		reenrollmentEvidence = digestFixture("reenrollment evidence")
	}
	if plan.Actions.DeleteAffectedData {
		deletionReceipt = digestFixture("deletion receipt")
	}
	resolution := BiometricIncidentResolution{
		Version: Version1, IncidentDigest: incidentDigest, PlanDigest: planDigest, Actions: plan.Actions, NewKeyEpoch: plan.NewKeyEpoch,
		NewTransformEpoch: plan.NewTransformEpoch, ReplacementModelDigest: plan.ReplacementModelDigest,
		ActionEvidenceDigests: actionEvidenceFixture(plan.Actions), ReenrollmentEvidenceDigest: reenrollmentEvidence,
		DeletionReceiptDigest: deletionReceipt, AuditTailDigest: auditTail, NotificationDigest: notificationDigest,
		ClosedCoordinate: 110, ParticipantSetDigest: incident.ParticipantSetDigest, Threshold: incident.Threshold,
		FixtureState: FixtureOnlyState, ExternalReviewBlock: ExternalReviewRequired,
	}
	value, err := resolution.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	resolution.Approvals = signIncidentValue(t, value, incident.AuthorityKeyEpoch, nodes, keys)
	return resolution, verifiedPlan, audit, notifications
}

func cloneResolution(resolution BiometricIncidentResolution) BiometricIncidentResolution {
	clone := resolution
	clone.ActionEvidenceDigests = append([]RecoveryActionEvidence(nil), resolution.ActionEvidenceDigests...)
	clone.Approvals = append([]NodeSignature(nil), resolution.Approvals...)
	for index := range clone.Approvals {
		clone.Approvals[index].Signature = append([]byte(nil), resolution.Approvals[index].Signature...)
	}
	return clone
}

func TestBiometricRecoveryPlanRequiredForApprovalAllKinds(t *testing.T) {
	for _, kind := range incidentKinds {
		t.Run(string(kind), func(t *testing.T) {
			incident, nodes, keys := incidentFixture(t, kind)
			_, verified := planFixture(t, incident, nodes, keys)
			contained, approved := incident, incident
			contained.State, approved.State = IncidentContained, IncidentRecoveryApproved
			if ValidateBiometricIncidentTransition(contained, approved) == nil {
				t.Fatal("generic transition bypassed verified recovery plan")
			}
			if err := ValidateBiometricIncidentRecoveryApproval(contained, approved, verified); err != nil {
				t.Fatal(err)
			}
			if err := ValidateBiometricIncidentRecoveryApproval(contained, approved, nil); err == nil {
				t.Fatal("nil recovery plan authorized approval")
			}
		})
	}
}

func TestBiometricRecoveryPlanRejectsAttacks(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentMassFalseMatches)
	plan, _ := planFixture(t, incident, nodes, keys)
	contained := incident
	contained.State = IncidentContained
	tests := []struct {
		name   string
		mutate func(*BiometricIncident, *BiometricRecoveryPlan)
	}{
		{"not contained", func(value *BiometricIncident, _ *BiometricRecoveryPlan) { value.State = IncidentDetected }},
		{"missing mandatory action", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.Actions.RevokeModel = false }},
		{"same revoked model", func(value *BiometricIncident, plan *BiometricRecoveryPlan) {
			plan.ReplacementModelDigest = value.AffectedModelCommitment
		}},
		{"authority epoch substitution", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.AuthorityKeyEpoch++ }},
		{"participant substitution", func(_ *BiometricIncident, value *BiometricRecoveryPlan) {
			value.ParticipantSetDigest = digestFixture("other set")
		}},
		{"production marker", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.FixtureState = "production" }},
		{"review bypass", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.ExternalReviewBlock = "approved" }},
		{"insufficient quorum", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.Approvals = value.Approvals[:1] }},
		{"forged quorum", func(_ *BiometricIncident, value *BiometricRecoveryPlan) { value.Approvals[0].Signature[0] ^= 1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidateIncident, candidatePlan := contained, plan
			candidatePlan.Approvals = append([]NodeSignature(nil), plan.Approvals...)
			for index := range candidatePlan.Approvals {
				candidatePlan.Approvals[index].Signature = append([]byte(nil), plan.Approvals[index].Signature...)
			}
			test.mutate(&candidateIncident, &candidatePlan)
			if _, err := VerifyBiometricRecoveryPlan(candidateIncident, candidatePlan, nodes); err == nil {
				t.Fatal("invalid recovery plan verified")
			}
		})
	}
}

func TestBiometricIncidentResolutionAllKindsAndResumePolicy(t *testing.T) {
	for _, kind := range incidentKinds {
		t.Run(string(kind), func(t *testing.T) {
			incident, nodes, keys := incidentFixture(t, kind)
			resolution, plan, audit, notifications := resolutionFixture(t, incident, nodes, keys)
			applied := false
			verified, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, NewMemoryIncidentResolutionConsumerFixture(func() error { applied = true; return nil }))
			if err != nil || !applied {
				t.Fatalf("resolution was not atomically applied: %v", err)
			}
			if !errors.Is(verified.CanResumeMatching(), ErrProductionUnavailable) {
				t.Fatal("production matching resume was not blocked")
			}
			fixtureErr := verified.CanResumeMatchingInFixture()
			if kind == IncidentUnlawfulProcessing && fixtureErr == nil {
				t.Fatal("unlawful processing received fixture resume capability")
			}
			if kind != IncidentUnlawfulProcessing && fixtureErr != nil {
				t.Fatal(fixtureErr)
			}
		})
	}
}

func TestBiometricIncidentResolutionRejectsAttacks(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentTemplateIndexCompromise)
	resolution, plan, audit, notifications := resolutionFixture(t, incident, nodes, keys)
	tests := []struct {
		name   string
		mutate func(*BiometricIncident, *BiometricIncidentResolution, *[]IncidentAuditEntry, *[]IncidentNotification)
	}{
		{"plan substitution", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.PlanDigest = digestFixture("other plan")
		}},
		{"action substitution", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.Actions.RotateKey = false
		}},
		{"missing action evidence", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.ActionEvidenceDigests = value.ActionEvidenceDigests[:len(value.ActionEvidenceDigests)-1]
		}},
		{"substituted action evidence", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.ActionEvidenceDigests[0].Action = "cease_processing"
		}},
		{"missing reenrollment evidence", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.ReenrollmentEvidenceDigest = ""
		}},
		{"missing deletion receipt", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.DeletionReceiptDigest = ""
		}},
		{"insufficient quorum", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.Approvals = value.Approvals[:1]
		}},
		{"forged approval", func(_ *BiometricIncident, value *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.Approvals[0].Signature[0] ^= 1
		}},
		{"missing audit state", func(_ *BiometricIncident, _ *BiometricIncidentResolution, values *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			*values = append((*values)[:2], (*values)[3:]...)
		}},
		{"nonincreasing audit coordinate", func(_ *BiometricIncident, _ *BiometricIncidentResolution, values *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			(*values)[2].Coordinate = (*values)[1].Coordinate
		}},
		{"wrong audit action", func(_ *BiometricIncident, _ *BiometricIncidentResolution, values *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			(*values)[2].Action = "incident_reopened"
		}},
		{"missing regulator notice", func(_ *BiometricIncident, _ *BiometricIncidentResolution, _ *[]IncidentAuditEntry, values *[]IncidentNotification) {
			*values = (*values)[:1]
		}},
		{"notice before freeze", func(value *BiometricIncident, _ *BiometricIncidentResolution, _ *[]IncidentAuditEntry, values *[]IncidentNotification) {
			(*values)[0].CompletedCoordinate = value.FreezeCoordinate
		}},
		{"notice after closure", func(_ *BiometricIncident, _ *BiometricIncidentResolution, _ *[]IncidentAuditEntry, values *[]IncidentNotification) {
			(*values)[1].CompletedCoordinate = 111
		}},
		{"notice after deadline", func(_ *BiometricIncident, _ *BiometricIncidentResolution, _ *[]IncidentAuditEntry, values *[]IncidentNotification) {
			(*values)[1].CompletedCoordinate = 113
		}},
		{"wrong incident state", func(value *BiometricIncident, _ *BiometricIncidentResolution, _ *[]IncidentAuditEntry, _ *[]IncidentNotification) {
			value.State = IncidentRecovering
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidateIncident, candidateResolution := incident, cloneResolution(resolution)
			candidateAudit := append([]IncidentAuditEntry(nil), audit...)
			candidateNotifications := append([]IncidentNotification(nil), notifications...)
			test.mutate(&candidateIncident, &candidateResolution, &candidateAudit, &candidateNotifications)
			if _, err := VerifyBiometricIncidentResolution(candidateIncident, candidateResolution, plan, nodes, candidateAudit, candidateNotifications, NewMemoryIncidentResolutionConsumerFixture(func() error { return nil })); err == nil {
				t.Fatal("invalid incident resolution verified")
			}
		})
	}
}

func TestUnlawfulProcessingRequiresDeletionReceipt(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentUnlawfulProcessing)
	resolution, plan, audit, notifications := resolutionFixture(t, incident, nodes, keys)
	resolution.DeletionReceiptDigest = digest(nil)
	if _, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, NewMemoryIncidentResolutionConsumerFixture(func() error { return nil })); err == nil {
		t.Fatal("unlawful-processing resolution without deletion receipt verified")
	}
}

func TestIncidentResolutionConsumerReplayAndCallbackRollback(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentDedupEnumeration)
	resolution, plan, audit, notifications := resolutionFixture(t, incident, nodes, keys)
	applied := 0
	consumer := NewMemoryIncidentResolutionConsumerFixture(func() error { applied++; return nil })
	if _, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, consumer); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, consumer); err == nil || applied != 1 {
		t.Fatal("resolution replay was not atomically rejected")
	}
	callbackErr := errors.New("apply failed")
	rollbackConsumer := NewMemoryIncidentResolutionConsumerFixture(func() error { return callbackErr })
	if verified, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, rollbackConsumer); verified != nil || !errors.Is(err, callbackErr) {
		t.Fatal("callback failure returned a verified or consumed resolution")
	}
	rollbackConsumer.apply = func() error { return nil }
	if _, err := VerifyBiometricIncidentResolution(incident, resolution, plan, nodes, audit, notifications, rollbackConsumer); err != nil {
		t.Fatalf("callback failure was not rolled back: %v", err)
	}
}

func TestBiometricIncidentAndNotificationTransitionsAreMonotonic(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentModelPoisoning)
	_, plan := planFixture(t, incident, nodes, keys)
	incident.State = IncidentDetected
	contained := incident
	contained.State = IncidentContained
	if err := ValidateBiometricIncidentTransition(incident, contained); err != nil {
		t.Fatal(err)
	}
	approved := contained
	approved.State = IncidentRecoveryApproved
	if err := ValidateBiometricIncidentRecoveryApproval(contained, approved, plan); err != nil {
		t.Fatal(err)
	}
	for _, state := range []BiometricIncidentState{IncidentRecovering, IncidentNotificationPending, IncidentClosed} {
		next := approved
		next.State = state
		if err := ValidateBiometricIncidentTransition(approved, next); err != nil {
			t.Fatal(err)
		}
		approved = next
	}
	rollback := approved
	rollback.State = IncidentRecovering
	if ValidateBiometricIncidentTransition(approved, rollback) == nil {
		t.Fatal("incident rollback accepted")
	}
	incidentDigest, _ := incident.Digest()
	pending := IncidentNotification{Version: Version1, IncidentDigest: incidentDigest, Audience: NotificationUsers, RecipientCommitment: digestFixture("users"), PolicyDigest: incident.PolicyDigest, ContentDigest: digestFixture("notice"), DeadlineCoordinate: 110, State: NotificationPending}
	dispatched := pending
	dispatched.State, dispatched.CompletedCoordinate = NotificationDispatched, 105
	if err := ValidateIncidentNotificationTransition(pending, dispatched); err != nil {
		t.Fatal(err)
	}
	if ValidateIncidentNotificationTransition(dispatched, pending) == nil {
		t.Fatal("notification rollback accepted")
	}
}

func TestBiometricIncidentPublicRecordsAreCommitmentOnly(t *testing.T) {
	incident, nodes, keys := incidentFixture(t, IncidentDedupEnumeration)
	plan, _ := planFixture(t, incident, nodes, keys)
	resolution, _, audit, notifications := resolutionFixture(t, incident, nodes, keys)
	for _, value := range []any{incident, plan, resolution, audit, notifications} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(encoded))
		for _, forbidden := range []string{"incident_id", "subject_id", "account_id", "template_bytes", "embedding", "candidate_id", "recipient_id", "raw_biometric"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%T leaks %q", value, forbidden)
			}
		}
	}
}
