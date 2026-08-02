package uniqueness

import (
	"errors"
	"slices"
	"sync"
)

type BiometricIncidentKind string

const (
	IncidentTemplateIndexCompromise BiometricIncidentKind = "template_index_compromise"
	IncidentTransformKeyCompromise  BiometricIncidentKind = "transform_key_compromise"
	IncidentMassFalseMatches        BiometricIncidentKind = "mass_false_matches"
	IncidentUnlawfulProcessing      BiometricIncidentKind = "unlawful_processing"
	IncidentModelPoisoning          BiometricIncidentKind = "model_poisoning"
	IncidentDedupEnumeration        BiometricIncidentKind = "dedup_enumeration"
)

type BiometricIncidentState string

const (
	IncidentDetected            BiometricIncidentState = "detected"
	IncidentContained           BiometricIncidentState = "contained"
	IncidentRecoveryApproved    BiometricIncidentState = "recovery_approved"
	IncidentRecovering          BiometricIncidentState = "recovering"
	IncidentNotificationPending BiometricIncidentState = "notification_pending"
	IncidentClosed              BiometricIncidentState = "closed"
	IncidentRevoked             BiometricIncidentState = "revoked"
)

type BiometricIncident struct {
	Version                 uint32                 `json:"version"`
	IncidentCommitment      string                 `json:"incident_commitment"`
	Kind                    BiometricIncidentKind  `json:"kind"`
	ScopeCommitment         string                 `json:"scope_commitment"`
	EvidenceDigest          string                 `json:"evidence_digest"`
	PolicyDigest            string                 `json:"policy_digest"`
	ParticipantSetDigest    string                 `json:"participant_set_digest"`
	Threshold               uint32                 `json:"threshold"`
	AuthorityKeyEpoch       uint64                 `json:"authority_key_epoch"`
	AffectedKeyEpoch        uint64                 `json:"affected_key_epoch"`
	AffectedTransformEpoch  uint64                 `json:"affected_transform_epoch"`
	AffectedModelCommitment string                 `json:"affected_model_commitment"`
	DetectedCoordinate      uint64                 `json:"detected_coordinate"`
	FreezeCoordinate        uint64                 `json:"freeze_coordinate"`
	State                   BiometricIncidentState `json:"state"`
}

func (incident BiometricIncident) CanonicalBytes() ([]byte, error) {
	if incident.Version != Version1 || !validIncidentKind(incident.Kind) ||
		!validDigests(incident.IncidentCommitment, incident.ScopeCommitment, incident.EvidenceDigest, incident.PolicyDigest, incident.ParticipantSetDigest, incident.AffectedModelCommitment) ||
		incident.Threshold < 2 || incident.AuthorityKeyEpoch == 0 || incident.AffectedKeyEpoch == 0 || incident.AffectedTransformEpoch == 0 ||
		incident.DetectedCoordinate == 0 || incident.FreezeCoordinate < incident.DetectedCoordinate || !validIncidentState(incident.State) {
		return nil, errors.New("complete biometric incident is required")
	}
	encoder := newCanonicalEncoder("virtengine.uniqueness.biometric-incident/v1")
	encoder.u32(incident.Version)
	encoder.text(incident.IncidentCommitment)
	encoder.text(string(incident.Kind))
	encoder.text(incident.ScopeCommitment)
	encoder.text(incident.EvidenceDigest)
	encoder.text(incident.PolicyDigest)
	encoder.text(incident.ParticipantSetDigest)
	encoder.u32(incident.Threshold)
	encoder.u64(incident.AuthorityKeyEpoch)
	encoder.u64(incident.AffectedKeyEpoch)
	encoder.u64(incident.AffectedTransformEpoch)
	encoder.text(incident.AffectedModelCommitment)
	encoder.u64(incident.DetectedCoordinate)
	encoder.u64(incident.FreezeCoordinate)
	encoder.text(string(incident.State))
	return encoder.result(), nil
}

func (incident BiometricIncident) Digest() (string, error) {
	value, err := incident.CanonicalBytesWithoutState()
	if err != nil {
		return "", err
	}
	return digest(value), nil
}

func ValidateBiometricIncidentTransition(previous, next BiometricIncident) error {
	if _, err := previous.CanonicalBytes(); err != nil {
		return err
	}
	if _, err := next.CanonicalBytes(); err != nil {
		return err
	}
	previousState, nextState := previous.State, next.State
	previousBytes, _ := previous.CanonicalBytesWithoutState()
	nextBytes, _ := next.CanonicalBytesWithoutState()
	if !slices.Equal(previousBytes, nextBytes) {
		return errors.New("biometric incident binding changed")
	}
	allowed := map[BiometricIncidentState]BiometricIncidentState{
		IncidentDetected:         IncidentContained,
		IncidentRecoveryApproved: IncidentRecovering, IncidentRecovering: IncidentNotificationPending,
		IncidentNotificationPending: IncidentClosed,
	}
	if nextState == IncidentRevoked && previousState != IncidentClosed && previousState != IncidentRevoked {
		return nil
	}
	if allowed[previousState] != nextState {
		return errors.New("non-monotonic biometric incident transition")
	}
	return nil
}

func ValidateBiometricIncidentRecoveryApproval(previous, next BiometricIncident, plan *VerifiedBiometricRecoveryPlan) error {
	if plan == nil || previous.State != IncidentContained || next.State != IncidentRecoveryApproved {
		return errors.New("verified recovery plan is required for recovery approval")
	}
	previousDigest, err := previous.Digest()
	if err != nil {
		return err
	}
	if plan.incidentDigest != previousDigest {
		return errors.New("recovery plan does not bind the contained incident")
	}
	previousBytes, _ := previous.CanonicalBytesWithoutState()
	nextBytes, err := next.CanonicalBytesWithoutState()
	if err != nil {
		return err
	}
	if !slices.Equal(previousBytes, nextBytes) {
		return errors.New("biometric incident binding changed")
	}
	return nil
}

func (incident BiometricIncident) CanonicalBytesWithoutState() ([]byte, error) {
	state := incident.State
	incident.State = IncidentDetected
	value, err := incident.CanonicalBytes()
	incident.State = state
	return value, err
}

type RecoveryActions struct {
	RotateKey             bool `json:"rotate_key"`
	RotateTransform       bool `json:"rotate_transform"`
	RebuildIndex          bool `json:"rebuild_index"`
	SuspendMatching       bool `json:"suspend_matching"`
	RevokeModel           bool `json:"revoke_model"`
	CeaseProcessing       bool `json:"cease_processing"`
	DeleteAffectedData    bool `json:"delete_affected_data"`
	EnumerationMitigation bool `json:"enumeration_mitigation"`
	ReenrollmentRequired  bool `json:"reenrollment_required"`
}

func (actions RecoveryActions) encode(encoder *canonicalEncoder) {
	encoder.boolean(actions.RotateKey)
	encoder.boolean(actions.RotateTransform)
	encoder.boolean(actions.RebuildIndex)
	encoder.boolean(actions.SuspendMatching)
	encoder.boolean(actions.RevokeModel)
	encoder.boolean(actions.CeaseProcessing)
	encoder.boolean(actions.DeleteAffectedData)
	encoder.boolean(actions.EnumerationMitigation)
	encoder.boolean(actions.ReenrollmentRequired)
}

func (actions RecoveryActions) enabledNames() []string {
	values := []struct {
		name    string
		enabled bool
	}{
		{"rotate_key", actions.RotateKey}, {"rotate_transform", actions.RotateTransform}, {"rebuild_index", actions.RebuildIndex},
		{"suspend_matching", actions.SuspendMatching}, {"revoke_model", actions.RevokeModel}, {"cease_processing", actions.CeaseProcessing},
		{"delete_affected_data", actions.DeleteAffectedData}, {"enumeration_mitigation", actions.EnumerationMitigation},
		{"reenrollment_required", actions.ReenrollmentRequired},
	}
	names := make([]string, 0, len(values))
	for _, value := range values {
		if value.enabled {
			names = append(names, value.name)
		}
	}
	return names
}

type RecoveryActionEvidence struct {
	Action         string `json:"action"`
	EvidenceDigest string `json:"evidence_digest"`
}

func encodeRecoveryActionEvidence(encoder *canonicalEncoder, actions RecoveryActions, evidence []RecoveryActionEvidence) error {
	names := actions.enabledNames()
	if len(evidence) != len(names) {
		return errors.New("every recovery action requires completion evidence")
	}
	encoder.u32(uint32(len(evidence)))
	for index, value := range evidence {
		if value.Action != names[index] || !validDigest(value.EvidenceDigest) {
			return errors.New("recovery action evidence is incomplete or noncanonical")
		}
		encoder.text(value.Action)
		encoder.text(value.EvidenceDigest)
	}
	return nil
}

func validateRecoveryActions(kind BiometricIncidentKind, actions RecoveryActions) error {
	required := RecoveryActions{}
	switch kind {
	case IncidentTemplateIndexCompromise:
		required = RecoveryActions{RotateKey: true, RotateTransform: true, RebuildIndex: true, ReenrollmentRequired: true}
	case IncidentTransformKeyCompromise:
		required = RecoveryActions{RotateKey: true, RotateTransform: true, ReenrollmentRequired: true}
	case IncidentMassFalseMatches:
		required = RecoveryActions{RebuildIndex: true, SuspendMatching: true, RevokeModel: true}
	case IncidentUnlawfulProcessing:
		required = RecoveryActions{CeaseProcessing: true, DeleteAffectedData: true}
	case IncidentModelPoisoning:
		required = RecoveryActions{RebuildIndex: true, SuspendMatching: true, RevokeModel: true}
	case IncidentDedupEnumeration:
		required = RecoveryActions{RotateKey: true, RotateTransform: true, EnumerationMitigation: true, ReenrollmentRequired: true}
	default:
		return errors.New("unknown biometric incident kind")
	}
	if required.RotateKey && !actions.RotateKey || required.RotateTransform && !actions.RotateTransform ||
		required.RebuildIndex && !actions.RebuildIndex || required.SuspendMatching && !actions.SuspendMatching ||
		required.RevokeModel && !actions.RevokeModel || required.CeaseProcessing && !actions.CeaseProcessing ||
		required.DeleteAffectedData && !actions.DeleteAffectedData || required.EnumerationMitigation && !actions.EnumerationMitigation ||
		required.ReenrollmentRequired && !actions.ReenrollmentRequired {
		return errors.New("incident recovery plan omits a mandatory action")
	}
	return nil
}

type BiometricRecoveryPlan struct {
	Version                uint32                `json:"version"`
	IncidentDigest         string                `json:"incident_digest"`
	Kind                   BiometricIncidentKind `json:"kind"`
	Actions                RecoveryActions       `json:"actions"`
	NewKeyEpoch            uint64                `json:"new_key_epoch"`
	NewTransformEpoch      uint64                `json:"new_transform_epoch"`
	ReplacementModelDigest string                `json:"replacement_model_digest"`
	ParticipantSetDigest   string                `json:"participant_set_digest"`
	Threshold              uint32                `json:"threshold"`
	AuthorityKeyEpoch      uint64                `json:"authority_key_epoch"`
	FixtureState           string                `json:"fixture_state"`
	ExternalReviewBlock    string                `json:"external_review_block"`
	Approvals              []NodeSignature       `json:"approvals"`
}

func (plan BiometricRecoveryPlan) CanonicalBytes() ([]byte, error) {
	if plan.Version != Version1 || !validIncidentKind(plan.Kind) ||
		!validDigests(plan.IncidentDigest, plan.ReplacementModelDigest, plan.ParticipantSetDigest) ||
		plan.NewKeyEpoch == 0 || plan.NewTransformEpoch == 0 || plan.Threshold < 2 || plan.AuthorityKeyEpoch == 0 ||
		plan.FixtureState != FixtureOnlyState || plan.ExternalReviewBlock != ExternalReviewRequired {
		return nil, errors.New("complete fixture-only biometric recovery plan is required")
	}
	encoder := newCanonicalEncoder("virtengine.uniqueness.biometric-recovery-plan/v1")
	encoder.u32(plan.Version)
	encoder.text(plan.IncidentDigest)
	encoder.text(string(plan.Kind))
	plan.Actions.encode(encoder)
	encoder.u64(plan.NewKeyEpoch)
	encoder.u64(plan.NewTransformEpoch)
	encoder.text(plan.ReplacementModelDigest)
	encoder.text(plan.ParticipantSetDigest)
	encoder.u32(plan.Threshold)
	encoder.u64(plan.AuthorityKeyEpoch)
	encoder.text(plan.FixtureState)
	encoder.text(plan.ExternalReviewBlock)
	return encoder.result(), nil
}

func (plan BiometricRecoveryPlan) Digest() (string, error) {
	value, err := plan.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return digest(value), nil
}

type VerifiedBiometricRecoveryPlan struct {
	incidentDigest         string
	planDigest             string
	actions                RecoveryActions
	newKeyEpoch            uint64
	newTransformEpoch      uint64
	replacementModelDigest string
}

func VerifyBiometricRecoveryPlan(incident BiometricIncident, plan BiometricRecoveryPlan, nodes CustodyNodeSet) (*VerifiedBiometricRecoveryPlan, error) {
	if incident.State != IncidentContained {
		return nil, errors.New("incident must be contained before recovery approval")
	}
	incidentDigest, err := incident.Digest()
	if err != nil {
		return nil, err
	}
	if plan.IncidentDigest != incidentDigest || plan.Kind != incident.Kind || plan.AuthorityKeyEpoch != incident.AuthorityKeyEpoch {
		return nil, errors.New("recovery plan does not bind the contained incident and authority epoch")
	}
	if err := validateRecoveryActions(incident.Kind, plan.Actions); err != nil {
		return nil, err
	}
	if plan.Actions.RotateKey && plan.NewKeyEpoch <= incident.AffectedKeyEpoch ||
		plan.Actions.RotateTransform && plan.NewTransformEpoch <= incident.AffectedTransformEpoch {
		return nil, errors.New("recovery plan epochs must advance compromised epochs")
	}
	if plan.Actions.RevokeModel && plan.ReplacementModelDigest == incident.AffectedModelCommitment {
		return nil, errors.New("revoked model cannot be its own replacement")
	}
	setDigest, err := nodes.Digest()
	if err != nil {
		return nil, err
	}
	if plan.ParticipantSetDigest != incident.ParticipantSetDigest || plan.ParticipantSetDigest != setDigest || plan.Threshold != incident.Threshold || plan.Threshold != nodes.Threshold {
		return nil, errors.New("recovery plan does not bind frozen incident authority")
	}
	value, err := plan.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	if err := verifyNodeSignatures(value, plan.AuthorityKeyEpoch, plan.Threshold, plan.Approvals, nodes); err != nil {
		return nil, err
	}
	return &VerifiedBiometricRecoveryPlan{
		incidentDigest: incidentDigest, planDigest: digest(value), actions: plan.Actions,
		newKeyEpoch: plan.NewKeyEpoch, newTransformEpoch: plan.NewTransformEpoch, replacementModelDigest: plan.ReplacementModelDigest,
	}, nil
}

type IncidentAuditEntry struct {
	Version         uint32                 `json:"version"`
	IncidentDigest  string                 `json:"incident_digest"`
	Sequence        uint64                 `json:"sequence"`
	PreviousDigest  string                 `json:"previous_digest"`
	State           BiometricIncidentState `json:"state"`
	Action          string                 `json:"action"`
	ActorCommitment string                 `json:"actor_commitment"`
	EvidenceDigest  string                 `json:"evidence_digest"`
	Coordinate      uint64                 `json:"coordinate"`
}

func (entry IncidentAuditEntry) CanonicalBytes() ([]byte, error) {
	if entry.Version != Version1 || !validDigests(entry.IncidentDigest, entry.PreviousDigest, entry.ActorCommitment, entry.EvidenceDigest) ||
		entry.Sequence == 0 || !validIncidentState(entry.State) || !validOpaqueID(entry.Action) || entry.Coordinate == 0 {
		return nil, errors.New("complete incident audit entry is required")
	}
	encoder := newCanonicalEncoder("virtengine.uniqueness.incident-audit/v1")
	encoder.u32(entry.Version)
	encoder.text(entry.IncidentDigest)
	encoder.u64(entry.Sequence)
	encoder.text(entry.PreviousDigest)
	encoder.text(string(entry.State))
	encoder.text(entry.Action)
	encoder.text(entry.ActorCommitment)
	encoder.text(entry.EvidenceDigest)
	encoder.u64(entry.Coordinate)
	return encoder.result(), nil
}

func (entry IncidentAuditEntry) Digest() (string, error) {
	value, err := entry.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return digest(value), nil
}

func ValidateIncidentAudit(entries []IncidentAuditEntry, incidentDigest string) (string, error) {
	states := []BiometricIncidentState{IncidentContained, IncidentRecoveryApproved, IncidentRecovering, IncidentNotificationPending, IncidentClosed}
	actions := []string{"incident_contained", "recovery_approved", "recovery_started", "notification_pending", "incident_closed"}
	if len(entries) != len(states) || !validDigest(incidentDigest) {
		return "", errors.New("incident audit chain is required")
	}
	previous := digest(nil)
	for index, entry := range entries {
		if entry.Sequence != uint64(index+1) || entry.IncidentDigest != incidentDigest || entry.PreviousDigest != previous ||
			entry.State != states[index] || entry.Action != actions[index] || index > 0 && entry.Coordinate <= entries[index-1].Coordinate {
			return "", errors.New("incident audit chain is discontinuous")
		}
		var err error
		previous, err = entry.Digest()
		if err != nil {
			return "", err
		}
	}
	return previous, nil
}

type NotificationAudience string

const (
	NotificationUsers      NotificationAudience = "users"
	NotificationRegulators NotificationAudience = "regulators"
)

type NotificationState string

const (
	NotificationPending      NotificationState = "pending"
	NotificationDispatched   NotificationState = "dispatched"
	NotificationAcknowledged NotificationState = "acknowledged"
)

type IncidentNotification struct {
	Version             uint32               `json:"version"`
	IncidentDigest      string               `json:"incident_digest"`
	Audience            NotificationAudience `json:"audience"`
	RecipientCommitment string               `json:"recipient_commitment"`
	PolicyDigest        string               `json:"policy_digest"`
	ContentDigest       string               `json:"content_digest"`
	DeadlineCoordinate  uint64               `json:"deadline_coordinate"`
	CompletedCoordinate uint64               `json:"completed_coordinate"`
	State               NotificationState    `json:"state"`
}

func (notification IncidentNotification) CanonicalBytes() ([]byte, error) {
	if notification.Version != Version1 || !validDigests(notification.IncidentDigest, notification.RecipientCommitment, notification.PolicyDigest, notification.ContentDigest) ||
		(notification.Audience != NotificationUsers && notification.Audience != NotificationRegulators) || notification.DeadlineCoordinate == 0 {
		return nil, errors.New("complete incident notification is required")
	}
	if notification.State != NotificationPending && notification.State != NotificationDispatched && notification.State != NotificationAcknowledged {
		return nil, errors.New("unknown incident notification state")
	}
	if notification.State == NotificationPending && notification.CompletedCoordinate != 0 ||
		notification.State != NotificationPending && (notification.CompletedCoordinate == 0 || notification.CompletedCoordinate > notification.DeadlineCoordinate) {
		return nil, errors.New("incident notification completion violates its deadline")
	}
	encoder := newCanonicalEncoder("virtengine.uniqueness.incident-notification/v1")
	encoder.u32(notification.Version)
	encoder.text(notification.IncidentDigest)
	encoder.text(string(notification.Audience))
	encoder.text(notification.RecipientCommitment)
	encoder.text(notification.PolicyDigest)
	encoder.text(notification.ContentDigest)
	encoder.u64(notification.DeadlineCoordinate)
	encoder.u64(notification.CompletedCoordinate)
	encoder.text(string(notification.State))
	return encoder.result(), nil
}

func ValidateIncidentNotificationTransition(previous, next IncidentNotification) error {
	previousBytes, err := previous.CanonicalBytes()
	if err != nil {
		return err
	}
	nextBytes, err := next.CanonicalBytes()
	if err != nil {
		return err
	}
	_ = previousBytes
	_ = nextBytes
	if previous.Version != next.Version || previous.IncidentDigest != next.IncidentDigest || previous.Audience != next.Audience ||
		previous.RecipientCommitment != next.RecipientCommitment || previous.PolicyDigest != next.PolicyDigest || previous.ContentDigest != next.ContentDigest ||
		previous.DeadlineCoordinate != next.DeadlineCoordinate {
		return errors.New("incident notification binding changed")
	}
	allowed := previous.State == NotificationPending && next.State == NotificationDispatched ||
		previous.State == NotificationDispatched && next.State == NotificationAcknowledged
	if !allowed || next.CompletedCoordinate < previous.CompletedCoordinate {
		return errors.New("non-monotonic incident notification transition")
	}
	return nil
}

func incidentNotificationDigest(notifications []IncidentNotification, incidentDigest string, notificationPendingCoordinate, closedCoordinate uint64) (string, error) {
	if len(notifications) != 2 {
		return "", errors.New("user and regulator notification states are required")
	}
	values := slices.Clone(notifications)
	slices.SortFunc(values, func(left, right IncidentNotification) int {
		return compareString(string(left.Audience), string(right.Audience))
	})
	encoder := newCanonicalEncoder("virtengine.uniqueness.incident-notification-set/v1")
	seen := map[NotificationAudience]bool{}
	for _, notification := range values {
		value, err := notification.CanonicalBytes()
		if err != nil {
			return "", err
		}
		if notification.IncidentDigest != incidentDigest || seen[notification.Audience] || notification.State == NotificationPending ||
			notification.CompletedCoordinate <= notificationPendingCoordinate || notification.CompletedCoordinate > closedCoordinate {
			return "", errors.New("incident notification obligation is incomplete or duplicated")
		}
		seen[notification.Audience] = true
		encoder.bytes(value)
	}
	if !seen[NotificationUsers] || !seen[NotificationRegulators] {
		return "", errors.New("user and regulator notification states are required")
	}
	return digest(encoder.result()), nil
}

type BiometricIncidentResolution struct {
	Version                    uint32                   `json:"version"`
	IncidentDigest             string                   `json:"incident_digest"`
	PlanDigest                 string                   `json:"plan_digest"`
	Actions                    RecoveryActions          `json:"actions"`
	NewKeyEpoch                uint64                   `json:"new_key_epoch"`
	NewTransformEpoch          uint64                   `json:"new_transform_epoch"`
	ReplacementModelDigest     string                   `json:"replacement_model_digest"`
	ActionEvidenceDigests      []RecoveryActionEvidence `json:"action_evidence_digests"`
	ReenrollmentEvidenceDigest string                   `json:"reenrollment_evidence_digest"`
	DeletionReceiptDigest      string                   `json:"deletion_receipt_digest"`
	AuditTailDigest            string                   `json:"audit_tail_digest"`
	NotificationDigest         string                   `json:"notification_digest"`
	ClosedCoordinate           uint64                   `json:"closed_coordinate"`
	ParticipantSetDigest       string                   `json:"participant_set_digest"`
	Threshold                  uint32                   `json:"threshold"`
	FixtureState               string                   `json:"fixture_state"`
	ExternalReviewBlock        string                   `json:"external_review_block"`
	Approvals                  []NodeSignature          `json:"approvals"`
}

func (resolution BiometricIncidentResolution) CanonicalBytes() ([]byte, error) {
	if resolution.Version != Version1 || !validDigests(resolution.IncidentDigest, resolution.PlanDigest, resolution.ReplacementModelDigest,
		resolution.ReenrollmentEvidenceDigest, resolution.DeletionReceiptDigest,
		resolution.AuditTailDigest, resolution.NotificationDigest, resolution.ParticipantSetDigest) || resolution.NewKeyEpoch == 0 ||
		resolution.NewTransformEpoch == 0 || resolution.ClosedCoordinate == 0 || resolution.Threshold < 2 ||
		resolution.FixtureState != FixtureOnlyState || resolution.ExternalReviewBlock != ExternalReviewRequired {
		return nil, errors.New("complete biometric incident resolution is required")
	}
	if resolution.Actions.ReenrollmentRequired != (resolution.ReenrollmentEvidenceDigest != digest(nil)) ||
		resolution.Actions.DeleteAffectedData != (resolution.DeletionReceiptDigest != digest(nil)) {
		return nil, errors.New("specialized recovery evidence does not match required actions")
	}
	encoder := newCanonicalEncoder("virtengine.uniqueness.biometric-incident-resolution/v1")
	encoder.u32(resolution.Version)
	encoder.text(resolution.IncidentDigest)
	encoder.text(resolution.PlanDigest)
	resolution.Actions.encode(encoder)
	encoder.u64(resolution.NewKeyEpoch)
	encoder.u64(resolution.NewTransformEpoch)
	encoder.text(resolution.ReplacementModelDigest)
	if err := encodeRecoveryActionEvidence(encoder, resolution.Actions, resolution.ActionEvidenceDigests); err != nil {
		return nil, err
	}
	encoder.text(resolution.ReenrollmentEvidenceDigest)
	encoder.text(resolution.DeletionReceiptDigest)
	encoder.text(resolution.AuditTailDigest)
	encoder.text(resolution.NotificationDigest)
	encoder.u64(resolution.ClosedCoordinate)
	encoder.text(resolution.ParticipantSetDigest)
	encoder.u32(resolution.Threshold)
	encoder.text(resolution.FixtureState)
	encoder.text(resolution.ExternalReviewBlock)
	return encoder.result(), nil
}

type VerifiedBiometricIncidentResolution struct {
	incidentDigest string
	kind           BiometricIncidentKind
	fixtureState   string
	externalReview string
}

type IncidentResolutionTransaction interface {
	ApplyIncidentClosed() error
}

type IncidentResolutionConsumer interface {
	ConsumeIncidentResolution(string, func(IncidentResolutionTransaction) error) error
}

type MemoryIncidentResolutionConsumerFixture struct {
	mu       sync.Mutex
	consumed map[string]bool
	apply    func() error
}

func NewMemoryIncidentResolutionConsumerFixture(apply func() error) *MemoryIncidentResolutionConsumerFixture {
	return &MemoryIncidentResolutionConsumerFixture{consumed: make(map[string]bool), apply: apply}
}

func (consumer *MemoryIncidentResolutionConsumerFixture) FixtureState() string {
	return FixtureOnlyState
}
func (consumer *MemoryIncidentResolutionConsumerFixture) ProductionReady() bool { return false }
func (consumer *MemoryIncidentResolutionConsumerFixture) ReviewStatus() string {
	return ExternalReviewRequired
}

func (consumer *MemoryIncidentResolutionConsumerFixture) ConsumeIncidentResolution(resolutionDigest string, apply func(IncidentResolutionTransaction) error) error {
	if consumer == nil || !validDigest(resolutionDigest) || apply == nil || consumer.apply == nil {
		return errors.New("valid incident resolution and fixture transaction are required")
	}
	consumer.mu.Lock()
	defer consumer.mu.Unlock()
	if consumer.consumed[resolutionDigest] {
		return errors.New("incident resolution replay rejected")
	}
	if err := apply(memoryIncidentResolutionTransaction{apply: consumer.apply}); err != nil {
		return err
	}
	consumer.consumed[resolutionDigest] = true
	return nil
}

type memoryIncidentResolutionTransaction struct{ apply func() error }

func (transaction memoryIncidentResolutionTransaction) ApplyIncidentClosed() error {
	return transaction.apply()
}

func VerifyBiometricIncidentResolution(incident BiometricIncident, resolution BiometricIncidentResolution, plan *VerifiedBiometricRecoveryPlan, nodes CustodyNodeSet, audit []IncidentAuditEntry, notifications []IncidentNotification, consumer IncidentResolutionConsumer) (*VerifiedBiometricIncidentResolution, error) {
	if incident.State != IncidentNotificationPending {
		return nil, errors.New("incident is not ready for closure")
	}
	incidentDigest, err := incident.Digest()
	if err != nil {
		return nil, err
	}
	if plan == nil || plan.incidentDigest != incidentDigest || resolution.PlanDigest != plan.planDigest ||
		resolution.Actions != plan.actions || resolution.NewKeyEpoch != plan.newKeyEpoch || resolution.NewTransformEpoch != plan.newTransformEpoch ||
		resolution.ReplacementModelDigest != plan.replacementModelDigest || resolution.IncidentDigest != incidentDigest || resolution.ClosedCoordinate < incident.FreezeCoordinate {
		return nil, errors.New("resolution does not bind the contained incident")
	}
	if err := validateRecoveryActions(incident.Kind, resolution.Actions); err != nil {
		return nil, err
	}
	if resolution.Actions.RotateKey && resolution.NewKeyEpoch <= incident.AffectedKeyEpoch ||
		resolution.Actions.RotateTransform && resolution.NewTransformEpoch <= incident.AffectedTransformEpoch {
		return nil, errors.New("recovery epochs must advance compromised epochs")
	}
	if resolution.Actions.RevokeModel && resolution.ReplacementModelDigest == incident.AffectedModelCommitment {
		return nil, errors.New("revoked model cannot be its own replacement")
	}
	setDigest, err := nodes.Digest()
	if err != nil {
		return nil, err
	}
	if resolution.ParticipantSetDigest != incident.ParticipantSetDigest || resolution.ParticipantSetDigest != setDigest || resolution.Threshold != incident.Threshold || resolution.Threshold != nodes.Threshold {
		return nil, errors.New("resolution does not bind frozen incident authority")
	}
	auditTail, err := ValidateIncidentAudit(audit, incidentDigest)
	if err != nil {
		return nil, err
	}
	lastAudit := audit[len(audit)-1]
	if audit[0].Coordinate < incident.FreezeCoordinate || auditTail != resolution.AuditTailDigest || lastAudit.Coordinate != resolution.ClosedCoordinate {
		return nil, errors.New("incident closure audit is incomplete")
	}
	notificationDigest, err := incidentNotificationDigest(notifications, incidentDigest, audit[3].Coordinate, resolution.ClosedCoordinate)
	if err != nil {
		return nil, err
	}
	if notificationDigest != resolution.NotificationDigest {
		return nil, errors.New("incident notification state was substituted")
	}
	value, err := resolution.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	if err := verifyNodeSignatures(value, incident.AuthorityKeyEpoch, resolution.Threshold, resolution.Approvals, nodes); err != nil {
		return nil, err
	}
	if consumer == nil {
		return nil, errors.New("atomic incident resolution consumer is required")
	}
	if err := consumer.ConsumeIncidentResolution(digest(value), func(transaction IncidentResolutionTransaction) error {
		if transaction == nil {
			return errors.New("incident resolution transaction is required")
		}
		return transaction.ApplyIncidentClosed()
	}); err != nil {
		return nil, err
	}
	return &VerifiedBiometricIncidentResolution{incidentDigest: incidentDigest, kind: incident.Kind, fixtureState: resolution.FixtureState, externalReview: resolution.ExternalReviewBlock}, nil
}

func (verified *VerifiedBiometricIncidentResolution) CanResumeMatching() error {
	return ErrProductionUnavailable

}

func (verified *VerifiedBiometricIncidentResolution) CanResumeMatchingInFixture() error {
	if verified == nil || !validDigest(verified.incidentDigest) || verified.fixtureState != FixtureOnlyState || verified.externalReview != ExternalReviewRequired {
		return errors.New("verified fixture-only biometric incident resolution is required")
	}
	if verified.kind == IncidentUnlawfulProcessing {
		return errors.New("unlawful processing can never authorize matching resume")
	}
	return nil
}

func validIncidentKind(kind BiometricIncidentKind) bool {
	switch kind {
	case IncidentTemplateIndexCompromise, IncidentTransformKeyCompromise, IncidentMassFalseMatches,
		IncidentUnlawfulProcessing, IncidentModelPoisoning, IncidentDedupEnumeration:
		return true
	}
	return false
}

func validIncidentState(state BiometricIncidentState) bool {
	switch state {
	case IncidentDetected, IncidentContained, IncidentRecoveryApproved, IncidentRecovering,
		IncidentNotificationPending, IncidentClosed, IncidentRevoked:
		return true
	}
	return false
}
