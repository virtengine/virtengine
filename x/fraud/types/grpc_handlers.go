// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - gRPC handlers and registration
package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpc "google.golang.org/grpc"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

// ============================================================================
// Message Server Registration
// ============================================================================

// MsgServer defines the fraud module's Msg service interface
type MsgServer interface {
	SubmitFraudReport(context.Context, *MsgSubmitFraudReport) (*MsgSubmitFraudReportResponse, error)
	AssignModerator(context.Context, *MsgAssignModerator) (*MsgAssignModeratorResponse, error)
	UpdateReportStatus(context.Context, *MsgUpdateReportStatus) (*MsgUpdateReportStatusResponse, error)
	ResolveFraudReport(context.Context, *MsgResolveFraudReport) (*MsgResolveFraudReportResponse, error)
	RejectFraudReport(context.Context, *MsgRejectFraudReport) (*MsgRejectFraudReportResponse, error)
	EscalateFraudReport(context.Context, *MsgEscalateFraudReport) (*MsgEscalateFraudReportResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// RegisterMsgServer registers the MsgServer with gRPC using direct protobuf types
func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	fraudv1.RegisterMsgServer(s, &msgServerAdapter{srv: srv})
}

// msgServerAdapter adapts the local MsgServer to the protobuf-generated interface
type msgServerAdapter struct {
	fraudv1.UnimplementedMsgServer
	srv MsgServer
}

func (a *msgServerAdapter) SubmitFraudReport(ctx context.Context, req *fraudv1.MsgSubmitFraudReport) (*fraudv1.MsgSubmitFraudReportResponse, error) {
	localReq := &MsgSubmitFraudReport{
		Reporter:        req.Reporter,
		ReportedParty:   req.ReportedParty,
		Category:        FraudCategory(req.Category),
		Description:     req.Description,
		Evidence:        convertEvidenceFromProto(req.Evidence),
		RelatedOrderIDs: req.RelatedOrderIds,
	}
	resp, err := a.srv.SubmitFraudReport(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgSubmitFraudReportResponse{
		ReportId: resp.ReportID,
	}, nil
}

func (a *msgServerAdapter) AssignModerator(ctx context.Context, req *fraudv1.MsgAssignModerator) (*fraudv1.MsgAssignModeratorResponse, error) {
	localReq := &MsgAssignModerator{
		Moderator: req.Moderator,
		ReportID:  req.ReportId,
		AssignTo:  req.AssignTo,
	}
	_, err := a.srv.AssignModerator(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgAssignModeratorResponse{}, nil
}

func (a *msgServerAdapter) UpdateReportStatus(ctx context.Context, req *fraudv1.MsgUpdateReportStatus) (*fraudv1.MsgUpdateReportStatusResponse, error) {
	localReq := &MsgUpdateReportStatus{
		Moderator: req.Moderator,
		ReportID:  req.ReportId,
		NewStatus: FraudReportStatus(req.NewStatus),
		Notes:     req.Notes,
	}
	_, err := a.srv.UpdateReportStatus(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgUpdateReportStatusResponse{}, nil
}

func (a *msgServerAdapter) ResolveFraudReport(ctx context.Context, req *fraudv1.MsgResolveFraudReport) (*fraudv1.MsgResolveFraudReportResponse, error) {
	localReq := &MsgResolveFraudReport{
		Moderator:  req.Moderator,
		ReportID:   req.ReportId,
		Resolution: ResolutionType(req.Resolution),
		Notes:      req.Notes,
	}
	_, err := a.srv.ResolveFraudReport(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgResolveFraudReportResponse{}, nil
}

func (a *msgServerAdapter) RejectFraudReport(ctx context.Context, req *fraudv1.MsgRejectFraudReport) (*fraudv1.MsgRejectFraudReportResponse, error) {
	localReq := &MsgRejectFraudReport{
		Moderator: req.Moderator,
		ReportID:  req.ReportId,
		Notes:     req.Notes,
	}
	_, err := a.srv.RejectFraudReport(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgRejectFraudReportResponse{}, nil
}

func (a *msgServerAdapter) EscalateFraudReport(ctx context.Context, req *fraudv1.MsgEscalateFraudReport) (*fraudv1.MsgEscalateFraudReportResponse, error) {
	localReq := &MsgEscalateFraudReport{
		Moderator: req.Moderator,
		ReportID:  req.ReportId,
		Reason:    req.Reason,
	}
	_, err := a.srv.EscalateFraudReport(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgEscalateFraudReportResponse{}, nil
}

func (a *msgServerAdapter) UpdateParams(ctx context.Context, req *fraudv1.MsgUpdateParams) (*fraudv1.MsgUpdateParamsResponse, error) {
	localReq := &MsgUpdateParams{
		Authority: req.Authority,
		Params:    convertParamsFromProtoValue(req.Params),
	}
	_, err := a.srv.UpdateParams(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.MsgUpdateParamsResponse{}, nil
}

// ============================================================================
// Query Server Registration
// ============================================================================

// QueryServer defines the fraud module's Query service interface
type QueryServer interface {
	FraudReport(context.Context, *QueryFraudReportRequest) (*QueryFraudReportResponse, error)
	FraudReports(context.Context, *QueryFraudReportsRequest) (*QueryFraudReportsResponse, error)
	FraudReportsByReporter(context.Context, *QueryFraudReportsByReporterRequest) (*QueryFraudReportsByReporterResponse, error)
	FraudReportsByReportedParty(context.Context, *QueryFraudReportsByReportedPartyRequest) (*QueryFraudReportsByReportedPartyResponse, error)
	ModeratorQueue(context.Context, *QueryModeratorQueueRequest) (*QueryModeratorQueueResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

// RegisterQueryServer registers the QueryServer with gRPC
func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	fraudv1.RegisterQueryServer(s, &queryServerAdapter{srv: srv})
}

// queryServerAdapter adapts the local QueryServer to the protobuf-generated interface
type queryServerAdapter struct {
	fraudv1.UnimplementedQueryServer
	srv QueryServer
}

func (a *queryServerAdapter) FraudReport(ctx context.Context, req *fraudv1.QueryFraudReportRequest) (*fraudv1.QueryFraudReportResponse, error) {
	localReq := &QueryFraudReportRequest{
		ReportID: req.ReportId,
	}
	resp, err := a.srv.FraudReport(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.QueryFraudReportResponse{
		Report: convertReportToProtoValue(resp.Report),
	}, nil
}

func (a *queryServerAdapter) FraudReports(ctx context.Context, req *fraudv1.QueryFraudReportsRequest) (*fraudv1.QueryFraudReportsResponse, error) {
	localReq := &QueryFraudReportsRequest{
		Status: FraudReportStatus(req.Status),
	}
	resp, err := a.srv.FraudReports(ctx, localReq)
	if err != nil {
		return nil, err
	}
	reports := make([]fraudv1.FraudReport, len(resp.Reports))
	for i, r := range resp.Reports {
		reports[i] = convertReportToProtoValue(r)
	}
	return &fraudv1.QueryFraudReportsResponse{
		Reports: reports,
	}, nil
}

func (a *queryServerAdapter) FraudReportsByReporter(ctx context.Context, req *fraudv1.QueryFraudReportsByReporterRequest) (*fraudv1.QueryFraudReportsByReporterResponse, error) {
	localReq := &QueryFraudReportsByReporterRequest{
		Reporter: req.Reporter,
	}
	resp, err := a.srv.FraudReportsByReporter(ctx, localReq)
	if err != nil {
		return nil, err
	}
	reports := make([]fraudv1.FraudReport, len(resp.Reports))
	for i, r := range resp.Reports {
		reports[i] = convertReportToProtoValue(r)
	}
	return &fraudv1.QueryFraudReportsByReporterResponse{
		Reports: reports,
	}, nil
}

func (a *queryServerAdapter) FraudReportsByReportedParty(ctx context.Context, req *fraudv1.QueryFraudReportsByReportedPartyRequest) (*fraudv1.QueryFraudReportsByReportedPartyResponse, error) {
	localReq := &QueryFraudReportsByReportedPartyRequest{
		ReportedParty: req.ReportedParty,
	}
	resp, err := a.srv.FraudReportsByReportedParty(ctx, localReq)
	if err != nil {
		return nil, err
	}
	reports := make([]fraudv1.FraudReport, len(resp.Reports))
	for i, r := range resp.Reports {
		reports[i] = convertReportToProtoValue(r)
	}
	return &fraudv1.QueryFraudReportsByReportedPartyResponse{
		Reports: reports,
	}, nil
}

func (a *queryServerAdapter) ModeratorQueue(ctx context.Context, req *fraudv1.QueryModeratorQueueRequest) (*fraudv1.QueryModeratorQueueResponse, error) {
	localReq := &QueryModeratorQueueRequest{
		Moderator: req.AssignedTo,
	}
	resp, err := a.srv.ModeratorQueue(ctx, localReq)
	if err != nil {
		return nil, err
	}
	entries := make([]fraudv1.ModeratorQueueEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = convertQueueEntryToProtoValue(e)
	}
	return &fraudv1.QueryModeratorQueueResponse{
		QueueEntries: entries,
	}, nil
}

func (a *queryServerAdapter) Params(ctx context.Context, req *fraudv1.QueryParamsRequest) (*fraudv1.QueryParamsResponse, error) {
	localReq := &QueryParamsRequest{}
	resp, err := a.srv.Params(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return &fraudv1.QueryParamsResponse{
		Params: convertParamsToProtoValue(resp.Params),
	}, nil
}

// ============================================================================
// Message Constructors
// ============================================================================

// MsgSubmitFraudReport is the local message type for submitting a fraud report
type MsgSubmitFraudReport struct {
	Reporter        string
	ReportedParty   string
	Category        FraudCategory
	Description     string
	Evidence        []EncryptedEvidence
	RelatedOrderIDs []string
}

// MsgSubmitFraudReportResponse is the response for MsgSubmitFraudReport
type MsgSubmitFraudReportResponse struct {
	ReportID string
}

// NewMsgSubmitFraudReport creates a new MsgSubmitFraudReport
func NewMsgSubmitFraudReport(reporter, reportedParty string, category FraudCategory, description string, evidence []EncryptedEvidence, relatedOrderIDs []string) *MsgSubmitFraudReport {
	return &MsgSubmitFraudReport{
		Reporter:        reporter,
		ReportedParty:   reportedParty,
		Category:        category,
		Description:     description,
		Evidence:        evidence,
		RelatedOrderIDs: relatedOrderIDs,
	}
}

// MsgAssignModerator is the local message type for assigning a moderator
type MsgAssignModerator struct {
	Moderator string
	ReportID  string
	AssignTo  string
}

// MsgAssignModeratorResponse is the response for MsgAssignModerator
type MsgAssignModeratorResponse struct{}

// NewMsgAssignModerator creates a new MsgAssignModerator
func NewMsgAssignModerator(moderator, reportID, assignTo string) *MsgAssignModerator {
	return &MsgAssignModerator{
		Moderator: moderator,
		ReportID:  reportID,
		AssignTo:  assignTo,
	}
}

// MsgUpdateReportStatus is the local message type for updating report status
type MsgUpdateReportStatus struct {
	Moderator string
	ReportID  string
	NewStatus FraudReportStatus
	Notes     string
}

// MsgUpdateReportStatusResponse is the response for MsgUpdateReportStatus
type MsgUpdateReportStatusResponse struct{}

// NewMsgUpdateReportStatus creates a new MsgUpdateReportStatus
func NewMsgUpdateReportStatus(moderator, reportID string, status FraudReportStatus, notes string) *MsgUpdateReportStatus {
	return &MsgUpdateReportStatus{
		Moderator: moderator,
		ReportID:  reportID,
		NewStatus: status,
		Notes:     notes,
	}
}

// MsgResolveFraudReport is the local message type for resolving a fraud report
type MsgResolveFraudReport struct {
	Moderator  string
	ReportID   string
	Resolution ResolutionType
	Notes      string
}

// MsgResolveFraudReportResponse is the response for MsgResolveFraudReport
type MsgResolveFraudReportResponse struct{}

// NewMsgResolveFraudReport creates a new MsgResolveFraudReport
func NewMsgResolveFraudReport(moderator, reportID string, resolution ResolutionType, notes string) *MsgResolveFraudReport {
	return &MsgResolveFraudReport{
		Moderator:  moderator,
		ReportID:   reportID,
		Resolution: resolution,
		Notes:      notes,
	}
}

// MsgRejectFraudReport is the local message type for rejecting a fraud report
type MsgRejectFraudReport struct {
	Moderator string
	ReportID  string
	Notes     string
}

// MsgRejectFraudReportResponse is the response for MsgRejectFraudReport
type MsgRejectFraudReportResponse struct{}

// NewMsgRejectFraudReport creates a new MsgRejectFraudReport
func NewMsgRejectFraudReport(moderator, reportID, reason string) *MsgRejectFraudReport {
	return &MsgRejectFraudReport{
		Moderator: moderator,
		ReportID:  reportID,
		Notes:     reason,
	}
}

// MsgEscalateFraudReport is the local message type for escalating a fraud report
type MsgEscalateFraudReport struct {
	Moderator string
	ReportID  string
	Reason    string
}

// MsgEscalateFraudReportResponse is the response for MsgEscalateFraudReport
type MsgEscalateFraudReportResponse struct{}

// NewMsgEscalateFraudReport creates a new MsgEscalateFraudReport
func NewMsgEscalateFraudReport(moderator, reportID, reason string) *MsgEscalateFraudReport {
	return &MsgEscalateFraudReport{
		Moderator: moderator,
		ReportID:  reportID,
		Reason:    reason,
	}
}

// MsgUpdateParams is the local message type for updating module params
type MsgUpdateParams struct {
	Authority string
	Params    Params
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ============================================================================
// SDK Msg Interface Implementation
// ============================================================================

// MsgSubmitFraudReport implements sdk.Msg
func (*MsgSubmitFraudReport) ProtoMessage() {}
func (m *MsgSubmitFraudReport) Reset()      { *m = MsgSubmitFraudReport{} }
func (m *MsgSubmitFraudReport) String() string {
	return "MsgSubmitFraudReport{Reporter: " + m.Reporter + ", ReportedParty: " + m.ReportedParty + "}"
}

func (m *MsgSubmitFraudReport) Route() string { return RouterKey }
func (m *MsgSubmitFraudReport) Type() string  { return TypeMsgSubmitFraudReport }
func (m *MsgSubmitFraudReport) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Reporter)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgSubmitFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Reporter); err != nil {
		return ErrInvalidReporter.Wrapf("invalid reporter address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ReportedParty); err != nil {
		return ErrInvalidReportedParty.Wrapf("invalid reported party address: %s", err)
	}
	if m.Reporter == m.ReportedParty {
		return ErrInvalidReportedParty.Wrap("cannot report yourself")
	}
	if m.Category == FraudCategoryUnspecified {
		return ErrInvalidCategory.Wrap("category cannot be unspecified")
	}
	if len(m.Description) < MinDescriptionLength {
		return ErrInvalidDescription.Wrapf("description must be at least %d characters", MinDescriptionLength)
	}
	if len(m.Description) > MaxDescriptionLength {
		return ErrInvalidDescription.Wrapf("description cannot exceed %d characters", MaxDescriptionLength)
	}
	if len(m.Evidence) == 0 {
		return ErrInvalidEvidence.Wrap("at least one evidence item is required")
	}
	return nil
}

// MsgAssignModerator implements sdk.Msg
func (*MsgAssignModerator) ProtoMessage() {}
func (m *MsgAssignModerator) Reset()      { *m = MsgAssignModerator{} }
func (m *MsgAssignModerator) String() string {
	return "MsgAssignModerator{Moderator: " + m.Moderator + ", ReportID: " + m.ReportID + "}"
}

func (m *MsgAssignModerator) Route() string { return RouterKey }
func (m *MsgAssignModerator) Type() string  { return TypeMsgAssignModerator }
func (m *MsgAssignModerator) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgAssignModerator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid moderator address: %s", err)
	}
	if m.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.AssignTo); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid assign-to address: %s", err)
	}
	return nil
}

// MsgUpdateReportStatus implements sdk.Msg
func (*MsgUpdateReportStatus) ProtoMessage() {}
func (m *MsgUpdateReportStatus) Reset()      { *m = MsgUpdateReportStatus{} }
func (m *MsgUpdateReportStatus) String() string {
	return "MsgUpdateReportStatus{Moderator: " + m.Moderator + ", ReportID: " + m.ReportID + "}"
}

func (m *MsgUpdateReportStatus) Route() string { return RouterKey }
func (m *MsgUpdateReportStatus) Type() string  { return TypeMsgUpdateReportStatus }
func (m *MsgUpdateReportStatus) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgUpdateReportStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid moderator address: %s", err)
	}
	if m.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID cannot be empty")
	}
	if m.NewStatus == FraudReportStatusUnspecified {
		return ErrInvalidStatus.Wrap("status cannot be unspecified")
	}
	return nil
}

// MsgResolveFraudReport implements sdk.Msg
func (*MsgResolveFraudReport) ProtoMessage() {}
func (m *MsgResolveFraudReport) Reset()      { *m = MsgResolveFraudReport{} }
func (m *MsgResolveFraudReport) String() string {
	return "MsgResolveFraudReport{Moderator: " + m.Moderator + ", ReportID: " + m.ReportID + "}"
}

func (m *MsgResolveFraudReport) Route() string { return RouterKey }
func (m *MsgResolveFraudReport) Type() string  { return TypeMsgResolveFraudReport }
func (m *MsgResolveFraudReport) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgResolveFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid moderator address: %s", err)
	}
	if m.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID cannot be empty")
	}
	if m.Resolution == ResolutionTypeUnspecified {
		return ErrInvalidResolution.Wrap("resolution cannot be unspecified")
	}
	if len(m.Notes) > MaxResolutionNotesLength {
		return ErrInvalidDescription.Wrapf("notes cannot exceed %d characters", MaxResolutionNotesLength)
	}
	return nil
}

// MsgRejectFraudReport implements sdk.Msg
func (*MsgRejectFraudReport) ProtoMessage() {}
func (m *MsgRejectFraudReport) Reset()      { *m = MsgRejectFraudReport{} }
func (m *MsgRejectFraudReport) String() string {
	return "MsgRejectFraudReport{Moderator: " + m.Moderator + ", ReportID: " + m.ReportID + "}"
}

func (m *MsgRejectFraudReport) Route() string { return RouterKey }
func (m *MsgRejectFraudReport) Type() string  { return TypeMsgRejectFraudReport }
func (m *MsgRejectFraudReport) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgRejectFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid moderator address: %s", err)
	}
	if m.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID cannot be empty")
	}
	if len(m.Notes) > MaxResolutionNotesLength {
		return ErrInvalidDescription.Wrapf("notes cannot exceed %d characters", MaxResolutionNotesLength)
	}
	return nil
}

// MsgEscalateFraudReport implements sdk.Msg
func (*MsgEscalateFraudReport) ProtoMessage() {}
func (m *MsgEscalateFraudReport) Reset()      { *m = MsgEscalateFraudReport{} }
func (m *MsgEscalateFraudReport) String() string {
	return "MsgEscalateFraudReport{Moderator: " + m.Moderator + ", ReportID: " + m.ReportID + "}"
}

func (m *MsgEscalateFraudReport) Route() string { return RouterKey }
func (m *MsgEscalateFraudReport) Type() string  { return TypeMsgEscalateFraudReport }
func (m *MsgEscalateFraudReport) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgEscalateFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid moderator address: %s", err)
	}
	if m.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID cannot be empty")
	}
	if m.Reason == "" {
		return ErrInvalidDescription.Wrap("escalation reason cannot be empty")
	}
	return nil
}

// MsgUpdateParams implements sdk.Msg
func (*MsgUpdateParams) ProtoMessage() {}
func (m *MsgUpdateParams) Reset()      { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string {
	return "MsgUpdateParams{Authority: " + m.Authority + "}"
}

func (m *MsgUpdateParams) Route() string { return RouterKey }
func (m *MsgUpdateParams) Type() string  { return TypeMsgUpdateParams }
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorizedModerator.Wrapf("invalid authority address: %s", err)
	}
	return m.Params.Validate()
}

// ============================================================================
// Query Request/Response Types
// ============================================================================

// QueryFraudReportRequest is the request for querying a fraud report
type QueryFraudReportRequest struct {
	ReportID string
}

// QueryFraudReportResponse is the response for querying a fraud report
type QueryFraudReportResponse struct {
	Report *FraudReport
}

// QueryFraudReportsRequest is the request for querying fraud reports
type QueryFraudReportsRequest struct {
	Status FraudReportStatus
}

// QueryFraudReportsResponse is the response for querying fraud reports
type QueryFraudReportsResponse struct {
	Reports []*FraudReport
}

// QueryFraudReportsByReporterRequest is the request for querying reports by reporter
type QueryFraudReportsByReporterRequest struct {
	Reporter string
}

// QueryFraudReportsByReporterResponse is the response for querying reports by reporter
type QueryFraudReportsByReporterResponse struct {
	Reports []*FraudReport
}

// QueryFraudReportsByReportedPartyRequest is the request for querying reports by reported party
type QueryFraudReportsByReportedPartyRequest struct {
	ReportedParty string
}

// QueryFraudReportsByReportedPartyResponse is the response for querying reports by reported party
type QueryFraudReportsByReportedPartyResponse struct {
	Reports []*FraudReport
}

// QueryModeratorQueueRequest is the request for querying a moderator's queue
type QueryModeratorQueueRequest struct {
	Moderator string
}

// QueryModeratorQueueResponse is the response for querying a moderator's queue
type QueryModeratorQueueResponse struct {
	Entries []*ModeratorQueueEntry
}

// QueryParamsRequest is the request for querying module params
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for querying module params
type QueryParamsResponse struct {
	Params Params
}

// ============================================================================
// Conversion Helpers
// ============================================================================

func convertEvidenceFromProto(evidence []fraudv1.EncryptedEvidence) []EncryptedEvidence {
	result := make([]EncryptedEvidence, len(evidence))
	for i, e := range evidence {
		result[i] = EncryptedEvidence{
			AlgorithmID:  e.AlgorithmId,
			Ciphertext:   e.Ciphertext,
			EvidenceHash: e.EvidenceHash,
			ContentType:  e.ContentType,
			SenderPubKey: e.SenderPubKey,
		}
	}
	return result
}

func convertParamsFromProto(params *fraudv1.Params) Params {
	if params == nil {
		return DefaultParams()
	}
	return Params{
		MinDescriptionLength:    int(params.MinDescriptionLength),
		MaxDescriptionLength:    int(params.MaxDescriptionLength),
		MaxEvidenceCount:        int(params.MaxEvidenceCount),
		ReportRetentionDays:     int(params.ReportRetentionDays),
		AutoAssignEnabled:       params.AutoAssignEnabled,
		EscalationThresholdDays: int(params.EscalationThresholdDays),
	}
}

func convertParamsFromProtoValue(params fraudv1.Params) Params {
	return Params{
		MinDescriptionLength:    int(params.MinDescriptionLength),
		MaxDescriptionLength:    int(params.MaxDescriptionLength),
		MaxEvidenceCount:        int(params.MaxEvidenceCount),
		ReportRetentionDays:     int(params.ReportRetentionDays),
		AutoAssignEnabled:       params.AutoAssignEnabled,
		EscalationThresholdDays: int(params.EscalationThresholdDays),
	}
}

func convertParamsToProto(params Params) *fraudv1.Params {
	return &fraudv1.Params{
		MinDescriptionLength:    int32(params.MinDescriptionLength),
		MaxDescriptionLength:    int32(params.MaxDescriptionLength),
		MaxEvidenceCount:        int32(params.MaxEvidenceCount),
		ReportRetentionDays:     int32(params.ReportRetentionDays),
		AutoAssignEnabled:       params.AutoAssignEnabled,
		EscalationThresholdDays: int32(params.EscalationThresholdDays),
	}
}

func convertParamsToProtoValue(params Params) fraudv1.Params {
	return fraudv1.Params{
		MinDescriptionLength:    int32(params.MinDescriptionLength),
		MaxDescriptionLength:    int32(params.MaxDescriptionLength),
		MaxEvidenceCount:        int32(params.MaxEvidenceCount),
		ReportRetentionDays:     int32(params.ReportRetentionDays),
		AutoAssignEnabled:       params.AutoAssignEnabled,
		EscalationThresholdDays: int32(params.EscalationThresholdDays),
	}
}

func convertReportToProto(report *FraudReport) *fraudv1.FraudReport {
	if report == nil {
		return nil
	}
	evidence := make([]fraudv1.EncryptedEvidence, len(report.Evidence))
	for i, e := range report.Evidence {
		evidence[i] = fraudv1.EncryptedEvidence{
			AlgorithmId:  e.AlgorithmID,
			Ciphertext:   e.Ciphertext,
			EvidenceHash: e.EvidenceHash,
			ContentType:  e.ContentType,
			SenderPubKey: e.SenderPubKey,
		}
	}
	return &fraudv1.FraudReport{
		Id:                report.ID,
		Reporter:          report.Reporter,
		ReportedParty:     report.ReportedParty,
		Category:          fraudv1.FraudCategory(report.Category),
		Status:            fraudv1.FraudReportStatus(report.Status),
		Description:       report.Description,
		Evidence:          evidence,
		RelatedOrderIds:   report.RelatedOrderIDs,
		AssignedModerator: report.AssignedModerator,
		Resolution:        fraudv1.ResolutionType(report.Resolution),
		ResolutionNotes:   report.ResolutionNotes,
		ContentHash:       report.ContentHash,
		BlockHeight:       report.BlockHeight,
	}
}

func convertReportToProtoValue(report *FraudReport) fraudv1.FraudReport {
	if report == nil {
		return fraudv1.FraudReport{}
	}
	evidence := make([]fraudv1.EncryptedEvidence, len(report.Evidence))
	for i, e := range report.Evidence {
		evidence[i] = fraudv1.EncryptedEvidence{
			AlgorithmId:  e.AlgorithmID,
			Ciphertext:   e.Ciphertext,
			EvidenceHash: e.EvidenceHash,
			ContentType:  e.ContentType,
			SenderPubKey: e.SenderPubKey,
		}
	}
	return fraudv1.FraudReport{
		Id:                report.ID,
		Reporter:          report.Reporter,
		ReportedParty:     report.ReportedParty,
		Category:          fraudv1.FraudCategory(report.Category),
		Status:            fraudv1.FraudReportStatus(report.Status),
		Description:       report.Description,
		Evidence:          evidence,
		RelatedOrderIds:   report.RelatedOrderIDs,
		AssignedModerator: report.AssignedModerator,
		Resolution:        fraudv1.ResolutionType(report.Resolution),
		ResolutionNotes:   report.ResolutionNotes,
		ContentHash:       report.ContentHash,
		BlockHeight:       report.BlockHeight,
	}
}

func convertQueueEntryToProto(entry *ModeratorQueueEntry) *fraudv1.ModeratorQueueEntry {
	if entry == nil {
		return nil
	}
	return &fraudv1.ModeratorQueueEntry{
		ReportId:   entry.ReportID,
		AssignedTo: entry.AssignedTo,
		Priority:   uint32(entry.Priority),
	}
}

func convertQueueEntryToProtoValue(entry *ModeratorQueueEntry) fraudv1.ModeratorQueueEntry {
	if entry == nil {
		return fraudv1.ModeratorQueueEntry{}
	}
	return fraudv1.ModeratorQueueEntry{
		ReportId:   entry.ReportID,
		AssignedTo: entry.AssignedTo,
		Priority:   uint32(entry.Priority),
	}
}
