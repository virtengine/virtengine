package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

// SDKQueryServer wraps the Keeper to implement the SDK's QueryServer interface.
// It embeds UnimplementedQueryServer to provide forward compatibility and stub
// implementations for methods not yet implemented.
type SDKQueryServer struct {
	veidv1.UnimplementedQueryServer
	keeper Keeper
}

// NewSDKQueryServer creates a new SDKQueryServer
func NewSDKQueryServer(k Keeper) *SDKQueryServer {
	return &SDKQueryServer{keeper: k}
}

var _ veidv1.QueryServer = (*SDKQueryServer)(nil)

// ModelVersion returns the active model info for a given model type.
func (s *SDKQueryServer) ModelVersion(goCtx context.Context, req *veidv1.QueryModelVersionRequest) (*veidv1.QueryModelVersionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	localReq := &types.QueryModelVersionRequest{ModelType: req.GetModelType()}
	resp, err := s.keeper.QueryModelVersion(ctx, localReq)
	if err != nil {
		return nil, err
	}

	var protoModel *veidv1.MLModelInfo
	if resp.ModelInfo != nil {
		protoModel = localModelInfoToProto(resp.ModelInfo)
	}

	return &veidv1.QueryModelVersionResponse{
		ModelInfo: protoModel,
		Found:     protoModel != nil,
	}, nil
}

// ActiveModels returns all currently active model versions.
func (s *SDKQueryServer) ActiveModels(goCtx context.Context, req *veidv1.QueryActiveModelsRequest) (*veidv1.QueryActiveModelsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	resp, err := s.keeper.QueryActiveModels(ctx, &types.QueryActiveModelsRequest{})
	if err != nil {
		return nil, err
	}

	protoModels := make([]veidv1.MLModelInfo, 0, len(resp.Models))
	for _, m := range resp.Models {
		if m != nil {
			protoModels = append(protoModels, *localModelInfoToProto(m))
		}
	}

	return &veidv1.QueryActiveModelsResponse{
		State:  localVersionStateToProto(&resp.State),
		Models: protoModels,
	}, nil
}

// ModelHistory returns the version change history for a model type.
func (s *SDKQueryServer) ModelHistory(goCtx context.Context, req *veidv1.QueryModelHistoryRequest) (*veidv1.QueryModelHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	localReq := &types.QueryModelHistoryRequest{ModelType: req.GetModelType()}
	if req.GetPagination() != nil && req.GetPagination().Limit > 0 {
		limit := req.GetPagination().Limit
		if limit > uint64(^uint32(0)) {
			limit = uint64(^uint32(0))
		}
		l := uint32(limit) //nolint:gosec // bounded above
		localReq.Pagination = &l
	}

	resp, err := s.keeper.QueryModelHistory(ctx, localReq)
	if err != nil {
		return nil, err
	}

	protoHistory := make([]veidv1.ModelVersionHistory, 0, len(resp.History))
	for _, h := range resp.History {
		if h != nil {
			protoHistory = append(protoHistory, localHistoryToProto(h))
		}
	}

	return &veidv1.QueryModelHistoryResponse{
		History: protoHistory,
	}, nil
}

// ValidatorModelSync returns the model sync status for a validator.
func (s *SDKQueryServer) ValidatorModelSync(goCtx context.Context, req *veidv1.QueryValidatorModelSyncRequest) (*veidv1.QueryValidatorModelSyncResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	localReq := &types.QueryValidatorModelSyncRequest{ValidatorAddress: req.GetValidatorAddress()}
	resp, err := s.keeper.QueryValidatorModelSync(ctx, localReq)
	if err != nil {
		return nil, err
	}

	var protoReport *veidv1.ValidatorModelReport
	if resp.Report != nil {
		r := localValidatorReportToProto(resp.Report)
		protoReport = &r
	}

	return &veidv1.QueryValidatorModelSyncResponse{
		Report:   protoReport,
		IsSynced: resp.IsSynced,
	}, nil
}

// ModelParams returns model management parameters.
func (s *SDKQueryServer) ModelParams(goCtx context.Context, req *veidv1.QueryModelParamsRequest) (*veidv1.QueryModelParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := s.keeper.GetModelParams(ctx)
	if err != nil {
		return nil, err
	}

	return &veidv1.QueryModelParamsResponse{
		Params: localModelParamsToProto(params),
	}, nil
}

// ============================================================================
// Conversion helpers: local types → proto types
// ============================================================================

// AssuranceVector returns an account's per-factor assurance vector.
func (s *SDKQueryServer) AssuranceVector(goCtx context.Context, req *veidv1.QueryAssuranceVectorRequest) (*veidv1.QueryAssuranceVectorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	resp, err := NewGRPCQuerier(s.keeper).QueryAssuranceVector(ctx, &types.QueryAssuranceVectorRequest{
		AccountAddress: req.GetAccountAddress(),
	})
	if err != nil {
		return nil, err
	}

	out := &veidv1.QueryAssuranceVectorResponse{
		Found:        resp.Found,
		CurrentEpoch: resp.CurrentEpoch,
	}
	if resp.Vector != nil {
		out.Vector = localAssuranceVectorToProto(resp.Vector)
	}
	if resp.Recency != nil {
		out.Recency = localAssuranceRecencyToProto(resp.Recency)
	}

	return out, nil
}

// AssuranceVectorHistory returns an account's assurance vectors, newest first.
func (s *SDKQueryServer) AssuranceVectorHistory(goCtx context.Context, req *veidv1.QueryAssuranceVectorHistoryRequest) (*veidv1.QueryAssuranceVectorHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	resp, err := NewGRPCQuerier(s.keeper).QueryAssuranceVectorHistory(ctx, &types.QueryAssuranceVectorHistoryRequest{
		AccountAddress: req.GetAccountAddress(),
	})
	if err != nil {
		return nil, err
	}

	out := &veidv1.QueryAssuranceVectorHistoryResponse{
		Vectors: make([]veidv1.AssuranceVector, 0, len(resp.Vectors)),
	}
	for _, vector := range resp.Vectors {
		if vector == nil {
			continue
		}
		out.Vectors = append(out.Vectors, *localAssuranceVectorToProto(vector))
	}

	return out, nil
}

func localAssuranceVectorToProto(v *types.AssuranceVector) *veidv1.AssuranceVector {
	if v == nil {
		return nil
	}

	factors := make([]veidv1.AssuranceFactorEntry, 0, len(v.Factors))
	for _, entry := range v.Factors {
		factors = append(factors, veidv1.AssuranceFactorEntry{
			Factor:           string(entry.Factor),
			ScoreBps:         entry.ScoreBps,
			WeightBps:        entry.WeightBps,
			WeightedScoreBps: entry.WeightedScoreBps,
			Measured:         entry.Measured,
			PassedThreshold:  entry.PassedThreshold,
		})
	}

	return &veidv1.AssuranceVector{
		Version:        v.Version,
		Account:        v.Account,
		Epoch:          v.Epoch,
		Factors:        factors,
		OverallBps:     v.OverallBps,
		ScalarScore:    v.ScalarScore,
		Passed:         v.Passed,
		ScoreVersion:   v.ScoreVersion,
		ModelVersion:   v.ModelVersion,
		AccountAgeBps:  v.AccountAgeBps,
		VerifiedHeight: v.VerifiedHeight,
		VerifiedAtUnix: v.VerifiedAtUnix,
		InputHash:      append([]byte(nil), v.InputHash...),
		Commitment:     append([]byte(nil), v.Commitment...),
	}
}

func localAssuranceRecencyToProto(r *types.AssuranceVectorRecency) *veidv1.AssuranceVectorRecency {
	if r == nil {
		return nil
	}

	return &veidv1.AssuranceVectorRecency{
		Epoch:          r.Epoch,
		AgeSeconds:     r.AgeSeconds,
		VerifiedHeight: r.VerifiedHeight,
		VerifiedAtUnix: r.VerifiedAtUnix,
		Stale:          r.Stale,
		MaxAgeSeconds:  r.MaxAgeSeconds,
	}
}

func localModelStatusToProto(s types.ModelStatus) veidv1.ModelStatus {
	switch s {
	case types.ModelStatusPending:
		return veidv1.ModelStatusPending
	case types.ModelStatusActive:
		return veidv1.ModelStatusActive
	case types.ModelStatusDeprecated:
		return veidv1.ModelStatusDeprecated
	case types.ModelStatusRevoked:
		return veidv1.ModelStatusRevoked
	default:
		return veidv1.ModelStatusUnspecified
	}
}

func localModelInfoToProto(m *types.MLModelInfo) *veidv1.MLModelInfo {
	return &veidv1.MLModelInfo{
		ModelId:      m.ModelID,
		Name:         m.Name,
		Version:      m.Version,
		ModelType:    m.ModelType,
		Sha256Hash:   m.SHA256Hash,
		Description:  m.Description,
		ActivatedAt:  m.ActivatedAt,
		RegisteredAt: m.RegisteredAt,
		RegisteredBy: m.RegisteredBy,
		GovernanceId: m.GovernanceID,
		Status:       localModelStatusToProto(m.Status),
	}
}

func localVersionStateToProto(s *types.ModelVersionState) veidv1.ModelVersionState {
	return veidv1.ModelVersionState{
		TrustScoreModel:       s.TrustScoreModel,
		FaceVerificationModel: s.FaceVerificationModel,
		LivenessModel:         s.LivenessModel,
		GanDetectionModel:     s.GANDetectionModel,
		OcrModel:              s.OCRModel,
		LastUpdated:           s.LastUpdated,
	}
}

func localHistoryToProto(h *types.ModelVersionHistory) veidv1.ModelVersionHistory {
	return veidv1.ModelVersionHistory{
		HistoryId:       h.HistoryID,
		ModelType:       h.ModelType,
		OldModelId:      h.OldModelID,
		NewModelId:      h.NewModelID,
		OldModelHash:    h.OldModelHash,
		NewModelHash:    h.NewModelHash,
		ChangedAt:       h.ChangedAt,
		GovernanceId:    h.GovernanceID,
		ProposerAddress: h.ProposerAddress,
		Reason:          h.Reason,
	}
}

func localValidatorReportToProto(r *types.ValidatorModelReport) veidv1.ValidatorModelReport {
	return veidv1.ValidatorModelReport{
		ValidatorAddress: r.ValidatorAddress,
		ModelVersions:    r.ModelVersions,
		ReportedAt:       r.ReportedAt,
		LastVerified:     r.LastVerified,
		IsSynced:         r.IsSynced,
		MismatchedModels: r.MismatchedModels,
	}
}

func localModelParamsToProto(p *types.ModelParams) veidv1.ModelParams {
	return veidv1.ModelParams{
		RequiredModelTypes:       p.RequiredModelTypes,
		ActivationDelayBlocks:    p.ActivationDelayBlocks,
		MaxModelAgeDays:          p.MaxModelAgeDays,
		AllowedRegistrars:        p.AllowedRegistrars,
		ValidatorSyncGracePeriod: p.ValidatorSyncGracePeriod,
		ModelUpdateQuorum:        p.ModelUpdateQuorum,
		EnableGovernanceUpdates:  p.EnableGovernanceUpdates,
	}
}
