package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/veid/types"
)

// Error message constants
const (
	errMsgEmptyRequest           = "empty request"
	errMsgAccountAddressEmpty    = "account address cannot be empty"
	errMsgInvalidAccountAddress  = "invalid account address"
	errMsgScopeIDEmpty           = "scope_id cannot be empty"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// IdentityRecord returns the identity record for an address
func (q GRPCQuerier) IdentityRecord(goCtx context.Context, req *types.QueryIdentityRecordRequest) (*types.QueryIdentityRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	record, found := q.Keeper.GetIdentityRecord(ctx, address)
	if !found {
		return &types.QueryIdentityRecordResponse{
			Record: nil,
		}, nil
	}

	return &types.QueryIdentityRecordResponse{
		Record: &record,
	}, nil
}

// Scope returns a specific scope for an address
func (q GRPCQuerier) Scope(goCtx context.Context, req *types.QueryScopeRequest) (*types.QueryScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if req.ScopeID == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgScopeIDEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	scope, found := q.Keeper.GetScope(ctx, address, req.ScopeID)
	if !found {
		return &types.QueryScopeResponse{
			Scope: nil,
		}, nil
	}

	return &types.QueryScopeResponse{
		Scope: &scope,
	}, nil
}

// ScopesByType returns all scopes of a specific type for an address
func (q GRPCQuerier) ScopesByType(goCtx context.Context, req *types.QueryScopesByTypeRequest) (*types.QueryScopesByTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if !types.IsValidScopeType(req.ScopeType) {
		return nil, status.Error(codes.InvalidArgument, "invalid scope type")
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	scopes := q.Keeper.GetScopesByType(ctx, address, req.ScopeType)

	return &types.QueryScopesByTypeResponse{
		Scopes: scopes,
	}, nil
}

// VerificationHistory returns the verification history for an address
func (q GRPCQuerier) VerificationHistory(goCtx context.Context, req *types.QueryVerificationHistoryRequest) (*types.QueryVerificationHistoryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	wallet, found := q.Keeper.GetWallet(ctx, address)
	if !found {
		return &types.QueryVerificationHistoryResponse{
			Entries:    []types.PublicVerificationHistoryEntry{},
			TotalCount: 0,
		}, nil
	}

	total := len(wallet.VerificationHistory)
	if total == 0 {
		return &types.QueryVerificationHistoryResponse{
			Entries:    []types.PublicVerificationHistoryEntry{},
			TotalCount: 0,
		}, nil
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	limit := int(req.Limit)
	if limit <= 0 || offset+limit > total {
		limit = total - offset
	}

	entries := make([]types.PublicVerificationHistoryEntry, 0, limit)
	for _, entry := range wallet.VerificationHistory[offset : offset+limit] {
		entries = append(entries, buildPublicVerificationEntry(entry))
	}

	return &types.QueryVerificationHistoryResponse{
		Entries:    entries,
		TotalCount: total,
	}, nil
}

// ApprovedClients returns all approved clients
func (q GRPCQuerier) ApprovedClients(goCtx context.Context, req *types.QueryApprovedClientsRequest) (*types.QueryApprovedClientsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var clients []types.ApprovedClient

	q.Keeper.WithApprovedClients(ctx, func(client types.ApprovedClient) bool {
		clients = append(clients, client)
		return false
	})

	return &types.QueryApprovedClientsResponse{
		Clients: clients,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// IdentityWallet returns the identity wallet for an address
func (q GRPCQuerier) IdentityWallet(goCtx context.Context, req *types.QueryIdentityWalletRequest) (*types.QueryIdentityWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	// Get public wallet info
	publicInfo, found := q.Keeper.GetWalletPublicMetadata(ctx, address)
	if !found {
		return &types.QueryIdentityWalletResponse{
			Wallet: nil,
			Found:  false,
		}, nil
	}

	return &types.QueryIdentityWalletResponse{
		Wallet: &publicInfo,
		Found:  true,
	}, nil
}

// WalletScopes returns the scopes for a wallet
func (q GRPCQuerier) WalletScopes(goCtx context.Context, req *types.QueryWalletScopesRequest) (*types.QueryWalletScopesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	wallet, found := q.Keeper.GetWallet(ctx, address)
	if !found {
		return &types.QueryWalletScopesResponse{
			Scopes:      []types.WalletScopeInfo{},
			TotalCount:  0,
			ActiveCount: 0,
		}, nil
	}

	var scopeTypeFilter types.ScopeType
	if req.ScopeType != "" {
		scopeTypeFilter = types.ScopeType(req.ScopeType)
		if !types.IsValidScopeType(scopeTypeFilter) {
			return nil, status.Error(codes.InvalidArgument, "invalid scope_type")
		}
	}

	var statusFilter types.ScopeRefStatus
	if req.StatusFilter != "" {
		statusFilter = types.ScopeRefStatus(req.StatusFilter)
		if !types.IsValidScopeRefStatus(statusFilter) {
			return nil, status.Error(codes.InvalidArgument, "invalid status_filter")
		}
	}

	currentTime := ctx.BlockTime()
	activeCount := 0
	for _, ref := range wallet.ScopeRefs {
		if ref.IsActive(currentTime) {
			activeCount++
		}
	}

	filtered := make([]types.WalletScopeInfo, 0, len(wallet.ScopeRefs))
	for _, ref := range wallet.ScopeRefs {
		if req.ActiveOnly && !ref.IsActive(currentTime) {
			continue
		}
		if req.ScopeType != "" && ref.ScopeType != scopeTypeFilter {
			continue
		}
		if req.StatusFilter != "" && ref.Status != statusFilter {
			continue
		}
		filtered = append(filtered, types.WalletScopeInfo{
			ScopeID:        ref.ScopeID,
			ScopeType:      ref.ScopeType,
			Status:         ref.Status,
			AddedAt:        ref.AddedAt,
			ConsentGranted: ref.ConsentGranted,
			ExpiresAt:      ref.ExpiresAt,
		})
	}

	return &types.QueryWalletScopesResponse{
		Scopes:      filtered,
		TotalCount:  len(wallet.ScopeRefs),
		ActiveCount: activeCount,
	}, nil
}

// ConsentSettings returns the consent settings for an address
func (q GRPCQuerier) ConsentSettings(goCtx context.Context, req *types.QueryConsentSettingsRequest) (*types.QueryConsentSettingsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	wallet, found := q.Keeper.GetWallet(ctx, address)
	if !found {
		return &types.QueryConsentSettingsResponse{
			ScopeConsents:  []types.PublicConsentInfo{},
			ConsentVersion: 0,
			LastUpdatedAt:  0,
		}, nil
	}

	currentTime := ctx.BlockTime()
	consents := make([]types.PublicConsentInfo, 0, len(wallet.ConsentSettings.ScopeConsents))
	if req.ScopeID != "" {
		if consent, ok := wallet.ConsentSettings.GetScopeConsent(req.ScopeID); ok {
			consents = append(consents, buildPublicConsentInfo(consent, currentTime))
		}
	} else {
		for _, consent := range wallet.ConsentSettings.ScopeConsents {
			consents = append(consents, buildPublicConsentInfo(consent, currentTime))
		}
	}

	resp := &types.QueryConsentSettingsResponse{
		ScopeConsents:  consents,
		ConsentVersion: wallet.ConsentSettings.ConsentVersion,
		LastUpdatedAt:  wallet.ConsentSettings.LastUpdatedAt.Unix(),
	}
	resp.GlobalSettings.ShareWithProviders = wallet.ConsentSettings.ShareWithProviders
	resp.GlobalSettings.ShareForVerification = wallet.ConsentSettings.ShareForVerification
	resp.GlobalSettings.AllowReVerification = wallet.ConsentSettings.AllowReVerification
	resp.GlobalSettings.AllowDerivedFeatureSharing = wallet.ConsentSettings.AllowDerivedFeatureSharing

	return resp, nil
}

func buildPublicConsentInfo(consent types.ScopeConsent, now time.Time) types.PublicConsentInfo {
	info := types.PublicConsentInfo{
		ScopeID:   consent.ScopeID,
		Granted:   consent.Granted,
		IsActive:  consent.IsActive(now),
		Purpose:   consent.Purpose,
		ExpiresAt: nil,
	}
	if consent.ExpiresAt != nil {
		exp := consent.ExpiresAt.Unix()
		info.ExpiresAt = &exp
	}
	return info
}

// DerivedFeatures returns the derived features for an address
func (q GRPCQuerier) DerivedFeatures(goCtx context.Context, req *types.QueryDerivedFeaturesRequest) (*types.QueryDerivedFeaturesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	wallet, found := q.Keeper.GetWallet(ctx, address)
	if !found {
		return &types.QueryDerivedFeaturesResponse{Features: nil, Found: false}, nil
	}

	info := wallet.DerivedFeatures.ToPublicInfo()
	return &types.QueryDerivedFeaturesResponse{Features: &info, Found: true}, nil
}

// DerivedFeatureHashes returns the derived feature hashes for an address
func (q GRPCQuerier) DerivedFeatureHashes(goCtx context.Context, req *types.QueryDerivedFeatureHashesRequest) (*types.QueryDerivedFeatureHashesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	wallet, found := q.Keeper.GetWallet(ctx, address)
	if !found {
		return &types.QueryDerivedFeatureHashesResponse{
			Allowed:      false,
			DenialReason: "wallet not found",
		}, nil
	}

	if !wallet.ConsentSettings.AllowDerivedFeatureSharing {
		return &types.QueryDerivedFeatureHashesResponse{
			Allowed:      false,
			DenialReason: "derived feature sharing not allowed",
		}, nil
	}

	return &types.QueryDerivedFeatureHashesResponse{
		Allowed:           true,
		FaceEmbeddingHash: wallet.DerivedFeatures.FaceEmbeddingHash,
		DocFieldHashes:    wallet.DerivedFeatures.DocFieldHashes,
		BiometricHash:     wallet.DerivedFeatures.BiometricHash,
		ModelVersion:      wallet.DerivedFeatures.ModelVersion,
	}, nil
}

func buildPublicVerificationEntry(entry types.VerificationHistoryEntry) types.PublicVerificationHistoryEntry {
	return types.PublicVerificationHistoryEntry{
		EntryID:        entry.EntryID,
		Timestamp:      entry.Timestamp.Unix(),
		BlockHeight:    entry.BlockHeight,
		PreviousScore:  entry.PreviousScore,
		NewScore:       entry.NewScore,
		PreviousStatus: string(entry.PreviousStatus),
		NewStatus:      string(entry.NewStatus),
		ScopeCount:     len(entry.ScopesEvaluated),
		ModelVersion:   entry.ModelVersion,
	}
}
