// Package types provides VEID module types.
//
// This file defines appeal-related event types for the VEID module.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package types

// Appeal event types
const (
	// EventTypeAppealSubmitted is emitted when an appeal is submitted
	EventTypeAppealSubmitted = "appeal_submitted"

	// EventTypeAppealClaimed is emitted when an arbitrator claims an appeal for review
	EventTypeAppealClaimed = "appeal_claimed"

	// EventTypeAppealResolved is emitted when an appeal is resolved (approved/rejected)
	EventTypeAppealResolved = "appeal_resolved"

	// EventTypeAppealWithdrawn is emitted when an appeal is withdrawn by the submitter
	EventTypeAppealWithdrawn = "appeal_withdrawn"

	// EventTypeAppealExpired is emitted when an appeal expires without resolution
	EventTypeAppealExpired = "appeal_expired"

	// EventTypeScoreAdjusted is emitted when a score is adjusted due to an approved appeal
	EventTypeScoreAdjusted = "score_adjusted"
)

// Appeal event attribute keys
const (
	// AttributeKeyAppealID is the appeal ID attribute key
	AttributeKeyAppealID = "appeal_id"

	// AttributeKeyAppealStatus is the appeal status attribute key
	AttributeKeyAppealStatus = "appeal_status"

	// AttributeKeyResolution is the resolution status attribute key
	AttributeKeyResolution = "resolution"

	// AttributeKeyScoreAdjustment is the score adjustment attribute key
	AttributeKeyScoreAdjustment = "score_adjustment"

	// AttributeKeyOriginalScore is the original score before adjustment
	AttributeKeyOriginalScore = "original_score"

	// AttributeKeyNewScore is the new score after adjustment
	AttributeKeyNewScore = "new_score"

	// AttributeKeyReviewer is the reviewer address attribute key
	AttributeKeyReviewer = "reviewer"

	// AttributeKeyAppealNumber is the appeal number for the scope
	AttributeKeyAppealNumber = "appeal_number"

	// AttributeKeyEvidenceCount is the count of evidence hashes submitted
	AttributeKeyEvidenceCount = "evidence_count"

	// AttributeKeyOriginalStatus is the original verification status being appealed
	AttributeKeyOriginalStatus = "original_status"

	// AttributeKeyResolutionReason is the reason for the resolution
	AttributeKeyResolutionReason = "resolution_reason"
)

// EventAppealSubmitted is emitted when an appeal is submitted
type EventAppealSubmitted struct {
	AppealID       string `json:"appeal_id"`
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
	OriginalStatus string `json:"original_status"`
	AppealNumber   uint32 `json:"appeal_number"`
	EvidenceCount  int    `json:"evidence_count"`
	SubmittedAt    int64  `json:"submitted_at"`
}

func (*EventAppealSubmitted) ProtoMessage() {}
func (m *EventAppealSubmitted) Reset()      { *m = EventAppealSubmitted{} }
func (m *EventAppealSubmitted) String() string {
	return "EventAppealSubmitted{AppealID: " + m.AppealID + "}"
}

// EventAppealClaimed is emitted when an arbitrator claims an appeal
type EventAppealClaimed struct {
	AppealID        string `json:"appeal_id"`
	ReviewerAddress string `json:"reviewer_address"`
	ClaimedAt       int64  `json:"claimed_at"`
}

func (*EventAppealClaimed) ProtoMessage() {}
func (m *EventAppealClaimed) Reset()      { *m = EventAppealClaimed{} }
func (m *EventAppealClaimed) String() string {
	return "EventAppealClaimed{AppealID: " + m.AppealID + "}"
}

// EventAppealResolved is emitted when an appeal is resolved
type EventAppealResolved struct {
	AppealID         string       `json:"appeal_id"`
	AccountAddress   string       `json:"account_address"`
	ScopeID          string       `json:"scope_id"`
	Resolution       AppealStatus `json:"resolution"`
	ResolutionReason string       `json:"resolution_reason"`
	ScoreAdjustment  int32        `json:"score_adjustment"`
	ReviewerAddress  string       `json:"reviewer_address"`
	ResolvedAt       int64        `json:"resolved_at"`
}

func (*EventAppealResolved) ProtoMessage() {}
func (m *EventAppealResolved) Reset()      { *m = EventAppealResolved{} }
func (m *EventAppealResolved) String() string {
	return "EventAppealResolved{AppealID: " + m.AppealID + "}"
}

// EventAppealWithdrawn is emitted when an appeal is withdrawn
type EventAppealWithdrawn struct {
	AppealID       string `json:"appeal_id"`
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
	WithdrawnAt    int64  `json:"withdrawn_at"`
}

func (*EventAppealWithdrawn) ProtoMessage() {}
func (m *EventAppealWithdrawn) Reset()      { *m = EventAppealWithdrawn{} }
func (m *EventAppealWithdrawn) String() string {
	return "EventAppealWithdrawn{AppealID: " + m.AppealID + "}"
}

// EventAppealExpired is emitted when an appeal expires
type EventAppealExpired struct {
	AppealID       string `json:"appeal_id"`
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
	ExpiredAt      int64  `json:"expired_at"`
}

func (*EventAppealExpired) ProtoMessage() {}
func (m *EventAppealExpired) Reset()      { *m = EventAppealExpired{} }
func (m *EventAppealExpired) String() string {
	return "EventAppealExpired{AppealID: " + m.AppealID + "}"
}

// EventAppealScoreAdjusted is emitted when a score is adjusted due to an approved appeal
type EventAppealScoreAdjusted struct {
	AppealID        string `json:"appeal_id"`
	AccountAddress  string `json:"account_address"`
	OriginalScore   uint32 `json:"original_score"`
	NewScore        uint32 `json:"new_score"`
	ScoreAdjustment int32  `json:"score_adjustment"`
	AdjustedAt      int64  `json:"adjusted_at"`
}

func (*EventAppealScoreAdjusted) ProtoMessage() {}
func (m *EventAppealScoreAdjusted) Reset()      { *m = EventAppealScoreAdjusted{} }
func (m *EventAppealScoreAdjusted) String() string {
	return "EventAppealScoreAdjusted{AppealID: " + m.AppealID + "}"
}
