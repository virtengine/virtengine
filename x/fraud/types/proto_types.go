// Package types provides type aliases and adapters for the generated proto types.
//
// VE-3053: This file bridges the local fraud types in x/fraud/types with the generated
// protobuf types in sdk/go/node/fraud/v1. It provides type aliases for proto message
// types and adapter implementations to translate between local and generated interfaces.
package types

import (
	"context"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

// =============================================================================
// Proto Type Aliases - Message Types
// =============================================================================

type (
	// MsgSubmitFraudReport is the generated proto type for submitting fraud reports
	MsgSubmitFraudReport = fraudv1.MsgSubmitFraudReport
	// MsgSubmitFraudReportResponse is the generated proto response type
	MsgSubmitFraudReportResponse = fraudv1.MsgSubmitFraudReportResponse
	// MsgAssignModerator is the generated proto type for assigning moderators
	MsgAssignModerator = fraudv1.MsgAssignModerator
	// MsgAssignModeratorResponse is the generated proto response type
	MsgAssignModeratorResponse = fraudv1.MsgAssignModeratorResponse
	// MsgUpdateReportStatus is the generated proto type for updating report status
	MsgUpdateReportStatus = fraudv1.MsgUpdateReportStatus
	// MsgUpdateReportStatusResponse is the generated proto response type
	MsgUpdateReportStatusResponse = fraudv1.MsgUpdateReportStatusResponse
	// MsgResolveFraudReport is the generated proto type for resolving fraud reports
	MsgResolveFraudReport = fraudv1.MsgResolveFraudReport
	// MsgResolveFraudReportResponse is the generated proto response type
	MsgResolveFraudReportResponse = fraudv1.MsgResolveFraudReportResponse
	// MsgRejectFraudReport is the generated proto type for rejecting fraud reports
	MsgRejectFraudReport = fraudv1.MsgRejectFraudReport
	// MsgRejectFraudReportResponse is the generated proto response type
	MsgRejectFraudReportResponse = fraudv1.MsgRejectFraudReportResponse
	// MsgEscalateFraudReport is the generated proto type for escalating fraud reports
	MsgEscalateFraudReport = fraudv1.MsgEscalateFraudReport
	// MsgEscalateFraudReportResponse is the generated proto response type
	MsgEscalateFraudReportResponse = fraudv1.MsgEscalateFraudReportResponse
	// MsgUpdateParams is the generated proto type for updating module params
	MsgUpdateParams = fraudv1.MsgUpdateParams
	// MsgUpdateParamsResponse is the generated proto response type
	MsgUpdateParamsResponse = fraudv1.MsgUpdateParamsResponse
)

// =============================================================================
// Proto Type Aliases - Query Types
// =============================================================================

type (
	// QueryParamsRequest is the generated proto type for params query
	QueryParamsRequest = fraudv1.QueryParamsRequest
	// QueryParamsResponse is the generated proto response type
	QueryParamsResponse = fraudv1.QueryParamsResponse
	// QueryFraudReportRequest is the generated proto type for fraud report query
	QueryFraudReportRequest = fraudv1.QueryFraudReportRequest
	// QueryFraudReportResponse is the generated proto response type
	QueryFraudReportResponse = fraudv1.QueryFraudReportResponse
	// QueryFraudReportsRequest is the generated proto type for fraud reports query
	QueryFraudReportsRequest = fraudv1.QueryFraudReportsRequest
	// QueryFraudReportsResponse is the generated proto response type
	QueryFraudReportsResponse = fraudv1.QueryFraudReportsResponse
	// QueryFraudReportsByReporterRequest is the generated proto type
	QueryFraudReportsByReporterRequest = fraudv1.QueryFraudReportsByReporterRequest
	// QueryFraudReportsByReporterResponse is the generated proto response type
	QueryFraudReportsByReporterResponse = fraudv1.QueryFraudReportsByReporterResponse
	// QueryFraudReportsByReportedPartyRequest is the generated proto type
	QueryFraudReportsByReportedPartyRequest = fraudv1.QueryFraudReportsByReportedPartyRequest
	// QueryFraudReportsByReportedPartyResponse is the generated proto response type
	QueryFraudReportsByReportedPartyResponse = fraudv1.QueryFraudReportsByReportedPartyResponse
	// QueryAuditLogRequest is the generated proto type for audit log query
	QueryAuditLogRequest = fraudv1.QueryAuditLogRequest
	// QueryAuditLogResponse is the generated proto response type
	QueryAuditLogResponse = fraudv1.QueryAuditLogResponse
	// QueryModeratorQueueRequest is the generated proto type for moderator queue query
	QueryModeratorQueueRequest = fraudv1.QueryModeratorQueueRequest
	// QueryModeratorQueueResponse is the generated proto response type
	QueryModeratorQueueResponse = fraudv1.QueryModeratorQueueResponse
)

// =============================================================================
// Proto Type Aliases - Data Types
// =============================================================================

type (
	// FraudReportPB is the generated proto type for fraud reports (PB suffix for proto)
	FraudReportPB = fraudv1.FraudReport
	// FraudAuditLogPB is the generated proto type for audit logs
	FraudAuditLogPB = fraudv1.FraudAuditLog
	// ModeratorQueueEntryPB is the generated proto type for queue entries
	ModeratorQueueEntryPB = fraudv1.ModeratorQueueEntry
	// EncryptedEvidencePB is the generated proto type for encrypted evidence
	EncryptedEvidencePB = fraudv1.EncryptedEvidence
	// ParamsPB is the generated proto type for module params
	ParamsPB = fraudv1.Params
	// GenesisStatePB is the generated proto type for genesis state
	GenesisStatePB = fraudv1.GenesisState
)

// =============================================================================
// Proto Enum Aliases
// =============================================================================

type (
	// FraudReportStatusPB is the generated proto enum for report status
	FraudReportStatusPB = fraudv1.FraudReportStatus
	// FraudCategoryPB is the generated proto enum for fraud category
	FraudCategoryPB = fraudv1.FraudCategory
	// ResolutionTypePB is the generated proto enum for resolution type
	ResolutionTypePB = fraudv1.ResolutionType
	// AuditActionPB is the generated proto enum for audit actions
	AuditActionPB = fraudv1.AuditAction
)

// Proto enum value constants - FraudReportStatus
const (
	FraudReportStatusPBUnspecified = fraudv1.FraudReportStatusUnspecified
	FraudReportStatusPBSubmitted   = fraudv1.FraudReportStatusSubmitted
	FraudReportStatusPBReviewing   = fraudv1.FraudReportStatusReviewing
	FraudReportStatusPBResolved    = fraudv1.FraudReportStatusResolved
	FraudReportStatusPBRejected    = fraudv1.FraudReportStatusRejected
	FraudReportStatusPBEscalated   = fraudv1.FraudReportStatusEscalated
)

// Proto enum value constants - FraudCategory
const (
	FraudCategoryPBUnspecified              = fraudv1.FraudCategoryUnspecified
	FraudCategoryPBFakeIdentity             = fraudv1.FraudCategoryFakeIdentity
	FraudCategoryPBPaymentFraud             = fraudv1.FraudCategoryPaymentFraud
	FraudCategoryPBServiceMisrepresentation = fraudv1.FraudCategoryServiceMisrepresentation
	FraudCategoryPBResourceAbuse            = fraudv1.FraudCategoryResourceAbuse
	FraudCategoryPBSybilAttack              = fraudv1.FraudCategorySybilAttack
	FraudCategoryPBMaliciousContent         = fraudv1.FraudCategoryMaliciousContent
	FraudCategoryPBTermsViolation           = fraudv1.FraudCategoryTermsViolation
	FraudCategoryPBOther                    = fraudv1.FraudCategoryOther
)

// Proto enum value constants - ResolutionType
const (
	ResolutionTypePBUnspecified = fraudv1.ResolutionTypeUnspecified
	ResolutionTypePBWarning     = fraudv1.ResolutionTypeWarning
	ResolutionTypePBSuspension  = fraudv1.ResolutionTypeSuspension
	ResolutionTypePBTermination = fraudv1.ResolutionTypeTermination
	ResolutionTypePBRefund      = fraudv1.ResolutionTypeRefund
	ResolutionTypePBNoAction    = fraudv1.ResolutionTypeNoAction
)

// Proto enum value constants - AuditAction
const (
	AuditActionPBUnspecified    = fraudv1.AuditActionUnspecified
	AuditActionPBSubmitted      = fraudv1.AuditActionSubmitted
	AuditActionPBAssigned       = fraudv1.AuditActionAssigned
	AuditActionPBStatusChanged  = fraudv1.AuditActionStatusChanged
	AuditActionPBEvidenceViewed = fraudv1.AuditActionEvidenceViewed
	AuditActionPBResolved       = fraudv1.AuditActionResolved
	AuditActionPBRejected       = fraudv1.AuditActionRejected
	AuditActionPBEscalated      = fraudv1.AuditActionEscalated
	AuditActionPBCommentAdded   = fraudv1.AuditActionCommentAdded
)

// =============================================================================
// MsgServer Interface
// =============================================================================

// MsgServer defines the fraud module's message server interface.
// This is a type alias to the proto-generated MsgServer.
type MsgServer = fraudv1.MsgServer

// UnimplementedMsgServer provides default unimplemented methods.
type UnimplementedMsgServer = fraudv1.UnimplementedMsgServer

// =============================================================================
// QueryServer Interface
// =============================================================================

// QueryServer defines the fraud module's query server interface.
// This is a type alias to the proto-generated QueryServer.
type QueryServer = fraudv1.QueryServer

// UnimplementedQueryServer provides default unimplemented methods.
type UnimplementedQueryServer = fraudv1.UnimplementedQueryServer

// =============================================================================
// Type Conversion Functions - FraudReportStatus
// =============================================================================

// FraudReportStatusToProto converts local FraudReportStatus to proto enum
func FraudReportStatusToProto(s FraudReportStatus) FraudReportStatusPB {
	return FraudReportStatusPB(s)
}

// FraudReportStatusFromProto converts proto enum to local FraudReportStatus
func FraudReportStatusFromProto(s FraudReportStatusPB) FraudReportStatus {
	if s < FraudReportStatusPBUnspecified || s > FraudReportStatusPBEscalated {
		return FraudReportStatusUnspecified
	}
	if s > FraudReportStatusPB(^uint8(0)) {
		return FraudReportStatusUnspecified
	}
	//nolint:gosec // range checked above
	return FraudReportStatus(uint8(s))
}

// =============================================================================
// Type Conversion Functions - FraudCategory
// =============================================================================

// FraudCategoryToProto converts local FraudCategory to proto enum
func FraudCategoryToProto(c FraudCategory) FraudCategoryPB {
	return FraudCategoryPB(c)
}

// FraudCategoryFromProto converts proto enum to local FraudCategory
func FraudCategoryFromProto(c FraudCategoryPB) FraudCategory {
	if c < FraudCategoryPBUnspecified || c > FraudCategoryPBOther {
		return FraudCategoryUnspecified
	}
	if c > FraudCategoryPB(^uint8(0)) {
		return FraudCategoryUnspecified
	}
	//nolint:gosec // range checked above
	return FraudCategory(uint8(c))
}

// =============================================================================
// Type Conversion Functions - ResolutionType
// =============================================================================

// ResolutionTypeToProto converts local ResolutionType to proto enum
func ResolutionTypeToProto(r ResolutionType) ResolutionTypePB {
	return ResolutionTypePB(r)
}

// ResolutionTypeFromProto converts proto enum to local ResolutionType
func ResolutionTypeFromProto(r ResolutionTypePB) ResolutionType {
	if r < ResolutionTypePBUnspecified || r > ResolutionTypePBNoAction {
		return ResolutionTypeUnspecified
	}
	if r > ResolutionTypePB(^uint8(0)) {
		return ResolutionTypeUnspecified
	}
	//nolint:gosec // range checked above
	return ResolutionType(uint8(r))
}

// =============================================================================
// Type Conversion Functions - AuditAction
// =============================================================================

// AuditActionToProto converts local AuditAction to proto enum
func AuditActionToProto(a AuditAction) AuditActionPB {
	return AuditActionPB(a)
}

// AuditActionFromProto converts proto enum to local AuditAction
func AuditActionFromProto(a AuditActionPB) AuditAction {
	if a < AuditActionPBUnspecified || a > AuditActionPBCommentAdded {
		return AuditActionUnspecified
	}
	if a > AuditActionPB(^uint8(0)) {
		return AuditActionUnspecified
	}
	//nolint:gosec // range checked above
	return AuditAction(uint8(a))
}

// =============================================================================
// Type Conversion Functions - FraudReport
// =============================================================================

// FraudReportToProto converts local FraudReport to proto FraudReport
func FraudReportToProto(r *FraudReport) *FraudReportPB {
	if r == nil {
		return nil
	}

	evidence := make([]EncryptedEvidencePB, len(r.Evidence))
	for i, e := range r.Evidence {
		evidence[i] = EncryptedEvidenceToProto(&e)
	}

	pb := &FraudReportPB{
		Id:                r.ID,
		Reporter:          r.Reporter,
		ReportedParty:     r.ReportedParty,
		Category:          FraudCategoryToProto(r.Category),
		Description:       r.Description,
		Evidence:          evidence,
		Status:            FraudReportStatusToProto(r.Status),
		AssignedModerator: r.AssignedModerator,
		Resolution:        ResolutionTypeToProto(r.Resolution),
		ResolutionNotes:   r.ResolutionNotes,
		SubmittedAt:       r.SubmittedAt,
		UpdatedAt:         r.UpdatedAt,
		BlockHeight:       r.BlockHeight,
		ContentHash:       r.ContentHash,
		RelatedOrderIds:   r.RelatedOrderIDs,
	}

	if r.ResolvedAt != nil {
		pb.ResolvedAt = r.ResolvedAt
	}

	return pb
}

// FraudReportFromProto converts proto FraudReport to local FraudReport
func FraudReportFromProto(pb *FraudReportPB) *FraudReport {
	if pb == nil {
		return nil
	}

	evidence := make([]EncryptedEvidence, len(pb.Evidence))
	for i, e := range pb.Evidence {
		evidence[i] = EncryptedEvidenceFromProto(&e)
	}

	r := &FraudReport{
		ID:                pb.Id,
		Reporter:          pb.Reporter,
		ReportedParty:     pb.ReportedParty,
		Category:          FraudCategoryFromProto(pb.Category),
		Description:       pb.Description,
		Evidence:          evidence,
		Status:            FraudReportStatusFromProto(pb.Status),
		AssignedModerator: pb.AssignedModerator,
		Resolution:        ResolutionTypeFromProto(pb.Resolution),
		ResolutionNotes:   pb.ResolutionNotes,
		SubmittedAt:       pb.SubmittedAt,
		UpdatedAt:         pb.UpdatedAt,
		BlockHeight:       pb.BlockHeight,
		ContentHash:       pb.ContentHash,
		RelatedOrderIDs:   pb.RelatedOrderIds,
	}

	if pb.ResolvedAt != nil {
		r.ResolvedAt = pb.ResolvedAt
	}

	return r
}

// =============================================================================
// Type Conversion Functions - EncryptedEvidence
// =============================================================================

// EncryptedEvidenceToProto converts local EncryptedEvidence to proto
func EncryptedEvidenceToProto(e *EncryptedEvidence) EncryptedEvidencePB {
	if e == nil {
		return EncryptedEvidencePB{}
	}
	return EncryptedEvidencePB{
		AlgorithmId:     e.AlgorithmID,
		RecipientKeyIds: e.RecipientKeyIDs,
		EncryptedKeys:   e.EncryptedKeys,
		Nonce:           e.Nonce,
		Ciphertext:      e.Ciphertext,
		SenderSignature: e.SenderSignature,
		SenderPubKey:    e.SenderPubKey,
		ContentType:     e.ContentType,
		EvidenceHash:    e.EvidenceHash,
	}
}

// EncryptedEvidenceFromProto converts proto EncryptedEvidence to local
func EncryptedEvidenceFromProto(pb *EncryptedEvidencePB) EncryptedEvidence {
	if pb == nil {
		return EncryptedEvidence{}
	}
	return EncryptedEvidence{
		AlgorithmID:     pb.AlgorithmId,
		RecipientKeyIDs: pb.RecipientKeyIds,
		EncryptedKeys:   pb.EncryptedKeys,
		Nonce:           pb.Nonce,
		Ciphertext:      pb.Ciphertext,
		SenderSignature: pb.SenderSignature,
		SenderPubKey:    pb.SenderPubKey,
		ContentType:     pb.ContentType,
		EvidenceHash:    pb.EvidenceHash,
	}
}

// =============================================================================
// Type Conversion Functions - FraudAuditLog
// =============================================================================

// FraudAuditLogToProto converts local FraudAuditLog to proto
func FraudAuditLogToProto(l *FraudAuditLog) *FraudAuditLogPB {
	if l == nil {
		return nil
	}
	return &FraudAuditLogPB{
		Id:             l.ID,
		ReportId:       l.ReportID,
		Action:         AuditActionToProto(l.Action),
		Actor:          l.Actor,
		PreviousStatus: FraudReportStatusToProto(l.PreviousStatus),
		NewStatus:      FraudReportStatusToProto(l.NewStatus),
		Details:        l.Details,
		Timestamp:      l.Timestamp,
		BlockHeight:    l.BlockHeight,
		TxHash:         l.TxHash,
	}
}

// FraudAuditLogFromProto converts proto FraudAuditLog to local
func FraudAuditLogFromProto(pb *FraudAuditLogPB) *FraudAuditLog {
	if pb == nil {
		return nil
	}
	return &FraudAuditLog{
		ID:             pb.Id,
		ReportID:       pb.ReportId,
		Action:         AuditActionFromProto(pb.Action),
		Actor:          pb.Actor,
		PreviousStatus: FraudReportStatusFromProto(pb.PreviousStatus),
		NewStatus:      FraudReportStatusFromProto(pb.NewStatus),
		Details:        pb.Details,
		Timestamp:      pb.Timestamp,
		BlockHeight:    pb.BlockHeight,
		TxHash:         pb.TxHash,
	}
}

// =============================================================================
// Type Conversion Functions - ModeratorQueueEntry
// =============================================================================

// ModeratorQueueEntryToProto converts local ModeratorQueueEntry to proto
func ModeratorQueueEntryToProto(e *ModeratorQueueEntry) *ModeratorQueueEntryPB {
	if e == nil {
		return nil
	}
	return &ModeratorQueueEntryPB{
		ReportId:   e.ReportID,
		Priority:   uint32(e.Priority),
		QueuedAt:   e.QueuedAt,
		Category:   FraudCategoryToProto(e.Category),
		AssignedTo: e.AssignedTo,
	}
}

// ModeratorQueueEntryFromProto converts proto ModeratorQueueEntry to local
func ModeratorQueueEntryFromProto(pb *ModeratorQueueEntryPB) *ModeratorQueueEntry {
	if pb == nil {
		return nil
	}
	priority := uint8(0)
	if pb.Priority > uint32(^uint8(0)) {
		priority = ^uint8(0)
	} else {
		priority = uint8(pb.Priority)
	}
	return &ModeratorQueueEntry{
		ReportID:   pb.ReportId,
		Priority:   priority,
		QueuedAt:   pb.QueuedAt,
		Category:   FraudCategoryFromProto(pb.Category),
		AssignedTo: pb.AssignedTo,
	}
}

// =============================================================================
// Type Conversion Functions - Params
// =============================================================================

// ParamsToProto converts local Params to proto Params
func ParamsToProto(p *Params) *ParamsPB {
	if p == nil {
		return nil
	}
	maxInt32 := int(^uint32(0) >> 1)
	minInt32 := -maxInt32 - 1
	minDesc := p.MinDescriptionLength
	if minDesc > maxInt32 {
		minDesc = maxInt32
	}
	if minDesc < minInt32 {
		minDesc = minInt32
	}
	maxDesc := p.MaxDescriptionLength
	if maxDesc > maxInt32 {
		maxDesc = maxInt32
	}
	if maxDesc < minInt32 {
		maxDesc = minInt32
	}
	maxEvidenceCount := p.MaxEvidenceCount
	if maxEvidenceCount > maxInt32 {
		maxEvidenceCount = maxInt32
	}
	if maxEvidenceCount < minInt32 {
		maxEvidenceCount = minInt32
	}
	escalation := p.EscalationThresholdDays
	if escalation > maxInt32 {
		escalation = maxInt32
	}
	if escalation < minInt32 {
		escalation = minInt32
	}
	reportRetention := p.ReportRetentionDays
	if reportRetention > maxInt32 {
		reportRetention = maxInt32
	}
	if reportRetention < minInt32 {
		reportRetention = minInt32
	}
	auditRetention := p.AuditLogRetentionDays
	if auditRetention > maxInt32 {
		auditRetention = maxInt32
	}
	if auditRetention < minInt32 {
		auditRetention = minInt32
	}
	return &ParamsPB{
		MinDescriptionLength:    safeInt32FromInt(minDesc),
		MaxDescriptionLength:    safeInt32FromInt(maxDesc),
		MaxEvidenceCount:        safeInt32FromInt(maxEvidenceCount),
		MaxEvidenceSizeBytes:    p.MaxEvidenceSizeBytes,
		AutoAssignEnabled:       p.AutoAssignEnabled,
		EscalationThresholdDays: safeInt32FromInt(escalation),
		ReportRetentionDays:     safeInt32FromInt(reportRetention),
		AuditLogRetentionDays:   safeInt32FromInt(auditRetention),
	}
}

// ParamsFromProto converts proto Params to local Params
func ParamsFromProto(pb *ParamsPB) *Params {
	if pb == nil {
		return nil
	}
	return &Params{
		MinDescriptionLength:    int(pb.MinDescriptionLength),
		MaxDescriptionLength:    int(pb.MaxDescriptionLength),
		MaxEvidenceCount:        int(pb.MaxEvidenceCount),
		MaxEvidenceSizeBytes:    pb.MaxEvidenceSizeBytes,
		AutoAssignEnabled:       pb.AutoAssignEnabled,
		EscalationThresholdDays: int(pb.EscalationThresholdDays),
		ReportRetentionDays:     int(pb.ReportRetentionDays),
		AuditLogRetentionDays:   int(pb.AuditLogRetentionDays),
	}
}

func safeInt32FromInt(value int) int32 {
	maxInt32 := int(^uint32(0) >> 1)
	minInt32 := -maxInt32 - 1
	if value > maxInt32 {
		//nolint:gosec // range checked above
		return int32(maxInt32)
	}
	if value < minInt32 {
		//nolint:gosec // range checked above
		return int32(minInt32)
	}
	//nolint:gosec // range checked above
	return int32(value)
}

// =============================================================================
// MsgServer Adapter
// =============================================================================

// msgServerAdapter adapts a local MsgServerImpl to the generated proto MsgServer.
type msgServerAdapter struct {
	impl MsgServerImpl
}

// MsgServerImpl is the interface that keeper implementations must satisfy.
// This mirrors the proto MsgServer but uses local types for convenience.
type MsgServerImpl interface {
	SubmitFraudReport(ctx context.Context, msg *MsgSubmitFraudReport) (*MsgSubmitFraudReportResponse, error)
	AssignModerator(ctx context.Context, msg *MsgAssignModerator) (*MsgAssignModeratorResponse, error)
	UpdateReportStatus(ctx context.Context, msg *MsgUpdateReportStatus) (*MsgUpdateReportStatusResponse, error)
	ResolveFraudReport(ctx context.Context, msg *MsgResolveFraudReport) (*MsgResolveFraudReportResponse, error)
	RejectFraudReport(ctx context.Context, msg *MsgRejectFraudReport) (*MsgRejectFraudReportResponse, error)
	EscalateFraudReport(ctx context.Context, msg *MsgEscalateFraudReport) (*MsgEscalateFraudReportResponse, error)
	UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// NewMsgServerAdapter creates a new adapter that wraps a MsgServerImpl
// to implement the generated proto MsgServer interface.
func NewMsgServerAdapter(impl MsgServerImpl) fraudv1.MsgServer {
	return &msgServerAdapter{impl: impl}
}

func (a *msgServerAdapter) SubmitFraudReport(ctx context.Context, req *fraudv1.MsgSubmitFraudReport) (*fraudv1.MsgSubmitFraudReportResponse, error) {
	return a.impl.SubmitFraudReport(ctx, req)
}

func (a *msgServerAdapter) AssignModerator(ctx context.Context, req *fraudv1.MsgAssignModerator) (*fraudv1.MsgAssignModeratorResponse, error) {
	return a.impl.AssignModerator(ctx, req)
}

func (a *msgServerAdapter) UpdateReportStatus(ctx context.Context, req *fraudv1.MsgUpdateReportStatus) (*fraudv1.MsgUpdateReportStatusResponse, error) {
	return a.impl.UpdateReportStatus(ctx, req)
}

func (a *msgServerAdapter) ResolveFraudReport(ctx context.Context, req *fraudv1.MsgResolveFraudReport) (*fraudv1.MsgResolveFraudReportResponse, error) {
	return a.impl.ResolveFraudReport(ctx, req)
}

func (a *msgServerAdapter) RejectFraudReport(ctx context.Context, req *fraudv1.MsgRejectFraudReport) (*fraudv1.MsgRejectFraudReportResponse, error) {
	return a.impl.RejectFraudReport(ctx, req)
}

func (a *msgServerAdapter) EscalateFraudReport(ctx context.Context, req *fraudv1.MsgEscalateFraudReport) (*fraudv1.MsgEscalateFraudReportResponse, error) {
	return a.impl.EscalateFraudReport(ctx, req)
}

func (a *msgServerAdapter) UpdateParams(ctx context.Context, req *fraudv1.MsgUpdateParams) (*fraudv1.MsgUpdateParamsResponse, error) {
	return a.impl.UpdateParams(ctx, req)
}

// =============================================================================
// QueryServer Adapter
// =============================================================================

// queryServerAdapter adapts a local QueryServerImpl to the generated proto QueryServer.
type queryServerAdapter struct {
	impl QueryServerImpl
}

// QueryServerImpl is the interface that keeper implementations must satisfy.
type QueryServerImpl interface {
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	FraudReport(ctx context.Context, req *QueryFraudReportRequest) (*QueryFraudReportResponse, error)
	FraudReports(ctx context.Context, req *QueryFraudReportsRequest) (*QueryFraudReportsResponse, error)
	FraudReportsByReporter(ctx context.Context, req *QueryFraudReportsByReporterRequest) (*QueryFraudReportsByReporterResponse, error)
	FraudReportsByReportedParty(ctx context.Context, req *QueryFraudReportsByReportedPartyRequest) (*QueryFraudReportsByReportedPartyResponse, error)
	AuditLog(ctx context.Context, req *QueryAuditLogRequest) (*QueryAuditLogResponse, error)
	ModeratorQueue(ctx context.Context, req *QueryModeratorQueueRequest) (*QueryModeratorQueueResponse, error)
}

// NewQueryServerAdapter creates a new adapter that wraps a QueryServerImpl
// to implement the generated proto QueryServer interface.
func NewQueryServerAdapter(impl QueryServerImpl) fraudv1.QueryServer {
	return &queryServerAdapter{impl: impl}
}

func (a *queryServerAdapter) Params(ctx context.Context, req *fraudv1.QueryParamsRequest) (*fraudv1.QueryParamsResponse, error) {
	return a.impl.Params(ctx, req)
}

func (a *queryServerAdapter) FraudReport(ctx context.Context, req *fraudv1.QueryFraudReportRequest) (*fraudv1.QueryFraudReportResponse, error) {
	return a.impl.FraudReport(ctx, req)
}

func (a *queryServerAdapter) FraudReports(ctx context.Context, req *fraudv1.QueryFraudReportsRequest) (*fraudv1.QueryFraudReportsResponse, error) {
	return a.impl.FraudReports(ctx, req)
}

func (a *queryServerAdapter) FraudReportsByReporter(ctx context.Context, req *fraudv1.QueryFraudReportsByReporterRequest) (*fraudv1.QueryFraudReportsByReporterResponse, error) {
	return a.impl.FraudReportsByReporter(ctx, req)
}

func (a *queryServerAdapter) FraudReportsByReportedParty(ctx context.Context, req *fraudv1.QueryFraudReportsByReportedPartyRequest) (*fraudv1.QueryFraudReportsByReportedPartyResponse, error) {
	return a.impl.FraudReportsByReportedParty(ctx, req)
}

func (a *queryServerAdapter) AuditLog(ctx context.Context, req *fraudv1.QueryAuditLogRequest) (*fraudv1.QueryAuditLogResponse, error) {
	return a.impl.AuditLog(ctx, req)
}

func (a *queryServerAdapter) ModeratorQueue(ctx context.Context, req *fraudv1.QueryModeratorQueueRequest) (*fraudv1.QueryModeratorQueueResponse, error) {
	return a.impl.ModeratorQueue(ctx, req)
}

// =============================================================================
// Type Conversion Functions - GenesisState
// =============================================================================

// GenesisStateToProto converts local GenesisState to proto GenesisState
func GenesisStateToProto(gs *GenesisState) *GenesisStatePB {
	if gs == nil {
		return nil
	}

	// Convert fraud reports
	reports := make([]FraudReportPB, len(gs.FraudReports))
	for i, r := range gs.FraudReports {
		reports[i] = *FraudReportToProto(&r)
	}

	// Convert audit logs
	auditLogs := make([]FraudAuditLogPB, len(gs.AuditLogs))
	for i, l := range gs.AuditLogs {
		auditLogs[i] = *FraudAuditLogToProto(&l)
	}

	// Convert moderator queue
	queue := make([]ModeratorQueueEntryPB, len(gs.ModeratorQueue))
	for i, e := range gs.ModeratorQueue {
		queue[i] = *ModeratorQueueEntryToProto(&e)
	}

	return &GenesisStatePB{
		Params:                  *ParamsToProto(&gs.Params),
		Reports:                 reports,
		AuditLogs:               auditLogs,
		ModeratorQueue:          queue,
		NextFraudReportSequence: gs.NextFraudReportSequence,
		NextAuditLogSequence:    gs.NextAuditLogSequence,
	}
}

// GenesisStateFromProto converts proto GenesisState to local GenesisState
func GenesisStateFromProto(pb *GenesisStatePB) *GenesisState {
	if pb == nil {
		return nil
	}

	// Convert fraud reports
	reports := make([]FraudReport, len(pb.Reports))
	for i, r := range pb.Reports {
		rCopy := r
		reports[i] = *FraudReportFromProto(&rCopy)
	}

	// Convert audit logs
	auditLogs := make([]FraudAuditLog, len(pb.AuditLogs))
	for i, l := range pb.AuditLogs {
		lCopy := l
		auditLogs[i] = *FraudAuditLogFromProto(&lCopy)
	}

	// Convert moderator queue
	queue := make([]ModeratorQueueEntry, len(pb.ModeratorQueue))
	for i, e := range pb.ModeratorQueue {
		eCopy := e
		queue[i] = *ModeratorQueueEntryFromProto(&eCopy)
	}

	return &GenesisState{
		Params:                  *ParamsFromProto(&pb.Params),
		FraudReports:            reports,
		AuditLogs:               auditLogs,
		ModeratorQueue:          queue,
		NextFraudReportSequence: pb.NextFraudReportSequence,
		NextAuditLogSequence:    pb.NextAuditLogSequence,
	}
}

// DefaultGenesisStatePB returns the default proto genesis state
func DefaultGenesisStatePB() *GenesisStatePB {
	return GenesisStateToProto(DefaultGenesisState())
}
