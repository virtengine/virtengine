package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pkg.akt.dev/node/x/veid/types"
)

// ============================================================================
// Score Query Endpoints
// ============================================================================

// QueryIdentityScore returns the identity score for an account
func (q GRPCQuerier) QueryIdentityScore(ctx sdk.Context, req *types.QueryIdentityScoreRequest) (*types.QueryIdentityScoreResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	identityScore, found := q.Keeper.GetIdentityScore(ctx, req.AccountAddress)
	if !found {
		return &types.QueryIdentityScoreResponse{
			Score: nil,
			Found: false,
		}, nil
	}

	return &types.QueryIdentityScoreResponse{
		Score: identityScore,
		Found: true,
	}, nil
}

// QueryIdentityStatus returns the identity status for an account
func (q GRPCQuerier) QueryIdentityStatus(ctx sdk.Context, req *types.QueryIdentityStatusRequest) (*types.QueryIdentityStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	score, accountStatus, found := q.Keeper.GetScore(ctx, req.AccountAddress)
	if !found {
		return &types.QueryIdentityStatusResponse{
			AccountAddress: req.AccountAddress,
			Status:         types.AccountStatusUnknown,
			Tier:           types.TierUnverified,
			TierName:       types.TierToString(types.TierUnverified),
			Score:          0,
			Found:          false,
		}, nil
	}

	tier := types.ComputeTierFromScoreValue(score, accountStatus)
	
	// Get last updated time from identity score if available
	var lastUpdatedAt *types.IdentityScore
	identityScore, scoreFound := q.Keeper.GetIdentityScore(ctx, req.AccountAddress)
	if scoreFound {
		lastUpdatedAt = identityScore
	}

	response := &types.QueryIdentityStatusResponse{
		AccountAddress: req.AccountAddress,
		Status:         accountStatus,
		Tier:           tier,
		TierName:       types.TierToString(tier),
		Score:          score,
		Found:          true,
	}

	if lastUpdatedAt != nil {
		response.ModelVersion = lastUpdatedAt.ModelVersion
		computedAt := lastUpdatedAt.ComputedAt
		response.LastUpdatedAt = &computedAt
	}

	return response, nil
}

// QueryScoreHistory returns the score history for an account
func (q GRPCQuerier) QueryScoreHistory(ctx sdk.Context, req *types.QueryScoreHistoryRequest) (*types.QueryScoreHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	// Validate address format
	_, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	// Apply pagination
	var limit uint32 = 100
	var offset uint32 = 0

	if req.Pagination != nil {
		if req.Pagination.Limit > 0 && req.Pagination.Limit <= 1000 {
			limit = uint32(req.Pagination.Limit)
		}
		if req.Pagination.Offset > 0 {
			offset = uint32(req.Pagination.Offset)
		}
	}

	entries := q.Keeper.GetScoreHistoryPaginated(ctx, req.AccountAddress, limit, offset)

	return &types.QueryScoreHistoryResponse{
		AccountAddress: req.AccountAddress,
		Entries:        entries,
	}, nil
}

// QueryRequiredScopes returns the required scopes for an offering type
func (q GRPCQuerier) QueryRequiredScopes(ctx sdk.Context, req *types.QueryRequiredScopesRequest) (*types.QueryRequiredScopesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if !types.IsValidOfferingType(req.OfferingType) {
		return nil, status.Error(codes.InvalidArgument, "invalid offering type")
	}

	requirements := types.GetRequiredScopesForOffering(req.OfferingType)

	return &types.QueryRequiredScopesResponse{
		Requirements: requirements,
	}, nil
}

// QueryAllRequiredScopes returns requirements for all offering types
func (q GRPCQuerier) QueryAllRequiredScopes(ctx sdk.Context, req *types.QueryAllRequiredScopesRequest) (*types.QueryAllRequiredScopesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var allRequirements []types.RequiredScopes
	for _, offeringType := range types.AllOfferingTypes() {
		allRequirements = append(allRequirements, types.GetRequiredScopesForOffering(offeringType))
	}

	return &types.QueryAllRequiredScopesResponse{
		AllRequirements: allRequirements,
	}, nil
}

// QueryAccountsByScoreTier returns accounts filtered by tier
func (q GRPCQuerier) QueryAccountsByScoreTier(ctx sdk.Context, req *types.QueryAccountsByScoreTierRequest) (*types.QueryAccountsByScoreTierResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	// Validate tier
	if req.Tier < types.TierUnverified || req.Tier > types.TierPremium {
		return nil, status.Error(codes.InvalidArgument, "invalid tier value")
	}

	// Get accounts by tier
	var accounts []types.AccountScoreSummary

	if req.MinScore != nil || req.MaxScore != nil {
		// Use score range filtering
		minScore := uint32(0)
		maxScore := types.MaxScore
		if req.MinScore != nil {
			minScore = *req.MinScore
		}
		if req.MaxScore != nil {
			maxScore = *req.MaxScore
		}
		accounts = q.Keeper.GetAccountsByScoreRange(ctx, minScore, maxScore)
	} else {
		// Use tier filtering
		accounts = q.Keeper.GetAccountsByTier(ctx, req.Tier)
	}

	// Apply status filter if provided
	if req.Status != nil {
		filteredAccounts := make([]types.AccountScoreSummary, 0)
		for _, acc := range accounts {
			if acc.Status == *req.Status {
				filteredAccounts = append(filteredAccounts, acc)
			}
		}
		accounts = filteredAccounts
	}

	// Apply pagination
	total := uint64(len(accounts))
	var offset uint64 = 0
	var limit uint64 = 100

	if req.Pagination != nil {
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		if req.Pagination.Limit > 0 && req.Pagination.Limit <= 1000 {
			limit = req.Pagination.Limit
		}
	}

	// Paginate results
	if offset >= uint64(len(accounts)) {
		accounts = []types.AccountScoreSummary{}
	} else {
		end := offset + limit
		if end > uint64(len(accounts)) {
			end = uint64(len(accounts))
		}
		accounts = accounts[offset:end]
	}

	return &types.QueryAccountsByScoreTierResponse{
		Accounts: accounts,
		Total:    total,
	}, nil
}

// QueryScoreThresholdCheck checks if an account meets a score threshold
func (q GRPCQuerier) QueryScoreThresholdCheck(ctx sdk.Context, req *types.QueryScoreThresholdCheckRequest) (*types.QueryScoreThresholdCheckResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	// Validate address format
	_, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	score, accountStatus, found := q.Keeper.GetScore(ctx, req.AccountAddress)
	if !found {
		return &types.QueryScoreThresholdCheckResponse{
			AccountAddress: req.AccountAddress,
			MeetsThreshold: false,
			CurrentScore:   0,
			CurrentStatus:  types.AccountStatusUnknown,
			Threshold:      req.Threshold,
			Found:          false,
		}, nil
	}

	meetsThreshold := score >= req.Threshold
	if req.RequireVerified && accountStatus != types.AccountStatusVerified {
		meetsThreshold = false
	}

	return &types.QueryScoreThresholdCheckResponse{
		AccountAddress: req.AccountAddress,
		MeetsThreshold: meetsThreshold,
		CurrentScore:   score,
		CurrentStatus:  accountStatus,
		Threshold:      req.Threshold,
		Found:          true,
	}, nil
}

// QueryEligibility checks eligibility for an offering type
func (q GRPCQuerier) QueryEligibility(ctx sdk.Context, req *types.QueryEligibilityRequest) (*types.QueryEligibilityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if !types.IsValidOfferingType(req.OfferingType) {
		return nil, status.Error(codes.InvalidArgument, "invalid offering type")
	}

	// Validate address format
	_, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	result := q.Keeper.CheckEligibility(ctx, req.AccountAddress, req.OfferingType)

	return &types.QueryEligibilityResponse{
		Result: result,
	}, nil
}

// QueryScoreStatistics returns aggregate score statistics
func (q GRPCQuerier) QueryScoreStatistics(ctx sdk.Context, req *types.QueryScoreStatisticsRequest) (*types.QueryScoreStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	stats := q.Keeper.GetScoreStatistics(ctx)

	return &types.QueryScoreStatisticsResponse{
		Statistics: stats,
	}, nil
}
