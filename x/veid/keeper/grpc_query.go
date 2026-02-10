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
	errMsgEmptyRequest          = "empty request"
	errMsgAccountAddressEmpty   = "account address cannot be empty"
	errMsgInvalidAccountAddress = "invalid account address"
	errMsgScopeIDEmpty          = "scope_id cannot be empty"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
type GRPCQuerier struct {
	Keeper
}

// NewGRPCQuerier returns a new GRPCQuerier instance
func NewGRPCQuerier(k Keeper) GRPCQuerier {
	return GRPCQuerier{Keeper: k}
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

	record, found := q.GetIdentityRecord(ctx, address)
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

	scope, found := q.GetScope(ctx, address, req.ScopeID)
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

	scopes := q.GetScopesByType(ctx, address, req.ScopeType)

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

	wallet, found := q.GetWallet(ctx, address)
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

// ApprovedClients returns all approved clients with pagination support
func (q GRPCQuerier) ApprovedClients(goCtx context.Context, req *types.QueryApprovedClientsRequest) (*types.QueryApprovedClientsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	// Default limit if not specified
	limit := req.Limit
	if limit == 0 {
		limit = 100 // Default page size
	}
	if limit > 1000 {
		limit = 1000 // Max page size
	}

	var allClients []types.ApprovedClient
	totalCount := uint32(0)

	q.WithApprovedClients(ctx, func(client types.ApprovedClient) bool {
		totalCount++
		// Apply offset and limit
		//nolint:gosec // slice length is non-negative
		if totalCount > req.Offset && uint32(len(allClients)) < limit {
			allClients = append(allClients, client)
		}
		return false
	})

	return &types.QueryApprovedClientsResponse{
		Clients:    allClients,
		TotalCount: totalCount,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// SSOLinkage returns the SSO linkage metadata for an account or linkage ID.
func (q GRPCQuerier) SSOLinkage(goCtx context.Context, req *types.QuerySSOLinkageRequest) (*types.QuerySSOLinkageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var linkage *types.SSOLinkageMetadata
	var found bool

	if req.LinkageId != "" {
		linkage, found = q.GetSSOLinkage(ctx, req.LinkageId)
	} else {
		if req.AccountAddress == "" {
			return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
		}
		if req.Provider == "" {
			return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
		}
		if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
			return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
		}
		linkageID := q.GetSSOLinkageByAccountAndProvider(ctx, req.AccountAddress, types.SSOProviderType(req.Provider))
		if linkageID != "" {
			linkage, found = q.GetSSOLinkage(ctx, linkageID)
		}
	}

	if !found || linkage == nil {
		return &types.QuerySSOLinkageResponse{Linkage: nil}, nil
	}

	return &types.QuerySSOLinkageResponse{
		Linkage: types.SSOLinkageToProto(linkage),
	}, nil
}

// EmailVerification returns the email verification record by ID or account/email hash.
func (q GRPCQuerier) EmailVerification(goCtx context.Context, req *types.QueryEmailVerificationRequest) (*types.QueryEmailVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var record *types.EmailVerificationRecord
	var found bool

	if req.VerificationId != "" {
		record, found = q.GetEmailVerificationRecord(ctx, req.VerificationId)
	} else {
		if req.AccountAddress == "" {
			return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
		}
		if req.EmailHash == "" {
			return nil, status.Error(codes.InvalidArgument, "email_hash cannot be empty")
		}
		if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
			return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
		}
		verificationID, ok := q.GetEmailVerificationByAccountAndHash(ctx, req.AccountAddress, req.EmailHash)
		if ok {
			record, found = q.GetEmailVerificationRecord(ctx, verificationID)
		}
	}

	if !found || record == nil {
		return &types.QueryEmailVerificationResponse{Record: nil}, nil
	}

	return &types.QueryEmailVerificationResponse{
		Record: types.EmailVerificationRecordToProto(record),
	}, nil
}

// SMSVerification returns the SMS verification record by ID or account/phone hash.
func (q GRPCQuerier) SMSVerification(goCtx context.Context, req *types.QuerySMSVerificationRequest) (*types.QuerySMSVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var record *types.SMSVerificationRecord
	var found bool

	if req.VerificationId != "" {
		record, found = q.GetSMSVerificationRecord(ctx, req.VerificationId)
	} else {
		if req.AccountAddress == "" {
			return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
		}
		if req.PhoneHash == "" {
			return nil, status.Error(codes.InvalidArgument, "phone_hash cannot be empty")
		}
		if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
			return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
		}
		verificationID, ok := q.GetSMSVerificationByAccountAndHash(ctx, req.AccountAddress, req.PhoneHash)
		if ok {
			record, found = q.GetSMSVerificationRecord(ctx, verificationID)
		}
	}

	if !found || record == nil {
		return &types.QuerySMSVerificationResponse{Record: nil}, nil
	}

	return &types.QuerySMSVerificationResponse{
		Record: types.SMSVerificationRecordToProto(record),
	}, nil
}

// SocialMediaScope returns a social media scope by ID.
func (q GRPCQuerier) SocialMediaScope(goCtx context.Context, req *types.QuerySocialMediaScopeRequest) (*types.QuerySocialMediaScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}
	if req.ScopeID == "" {
		return nil, status.Error(codes.InvalidArgument, "scope_id cannot be empty")
	}

	scope, found := q.GetSocialMediaScope(ctx, req.ScopeID)
	if !found || scope == nil {
		return &types.QuerySocialMediaScopeResponse{Scope: nil}, nil
	}

	return &types.QuerySocialMediaScopeResponse{
		Scope: types.SocialMediaScopeToProto(scope),
	}, nil
}

// SocialMediaScopes returns social media scopes for an account.
func (q GRPCQuerier) SocialMediaScopes(goCtx context.Context, req *types.QuerySocialMediaScopesRequest) (*types.QuerySocialMediaScopesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}
	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}
	if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	var providerPtr *types.SocialMediaProviderType
	if req.Provider != "" {
		provider := types.SocialMediaProviderType(req.Provider)
		if !types.IsValidSocialMediaProvider(provider) {
			return nil, status.Error(codes.InvalidArgument, "invalid provider")
		}
		providerPtr = &provider
	}

	records := q.GetSocialMediaScopesByAccount(ctx, req.AccountAddress, providerPtr)
	resp := make([]types.SocialMediaScopePB, 0, len(records))
	for i := range records {
		resp = append(resp, *types.SocialMediaScopeToProto(&records[i]))
	}

	return &types.QuerySocialMediaScopesResponse{
		Scopes: resp,
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
	publicInfo, found := q.GetWalletPublicMetadata(ctx, address)
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

	wallet, found := q.GetWallet(ctx, address)
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

	wallet, found := q.GetWallet(ctx, address)
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

	wallet, found := q.GetWallet(ctx, address)
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

	wallet, found := q.GetWallet(ctx, address)
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

// ============================================================================
// Appeal Queries (VE-3020)
// ============================================================================

// Appeal returns a specific appeal by ID
func (q GRPCQuerier) Appeal(goCtx context.Context, req *types.QueryAppealRequest) (*types.QueryAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AppealID == "" {
		return nil, status.Error(codes.InvalidArgument, "appeal_id cannot be empty")
	}

	appeal, found := q.GetAppeal(ctx, req.AppealID)
	if !found {
		return &types.QueryAppealResponse{Appeal: nil}, nil
	}

	return &types.QueryAppealResponse{Appeal: appeal}, nil
}

// AppealsByAccount returns all appeals for an account
func (q GRPCQuerier) AppealsByAccount(goCtx context.Context, req *types.QueryAppealsByAccountRequest) (*types.QueryAppealsByAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	_, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	appeals := q.GetAppealsByAccount(ctx, req.AccountAddress)
	summary := q.GetAppealSummary(ctx, req.AccountAddress)

	return &types.QueryAppealsByAccountResponse{
		Appeals: appeals,
		Summary: summary,
	}, nil
}

// PendingAppeals returns all pending appeals (for reviewers) with pagination support
func (q GRPCQuerier) PendingAppeals(goCtx context.Context, req *types.QueryPendingAppealsRequest) (*types.QueryPendingAppealsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	allAppeals := q.GetPendingAppeals(ctx)
	total := safeUint32FromIntBiometric(len(allAppeals))

	// Apply offset
	offset := int(req.Offset)
	if offset > len(allAppeals) {
		offset = len(allAppeals)
	}
	appeals := allAppeals[offset:]

	// Apply limit
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100 // Default page size
	}
	if limit > 1000 {
		limit = 1000 // Max page size
	}
	if len(appeals) > limit {
		appeals = appeals[:limit]
	}

	return &types.QueryPendingAppealsResponse{
		Appeals: appeals,
		Total:   total,
	}, nil
}

// AppealParams returns the appeal system parameters
func (q GRPCQuerier) AppealParams(goCtx context.Context, req *types.QueryAppealParamsRequest) (*types.QueryAppealParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	params := q.GetAppealParams(ctx)

	return &types.QueryAppealParamsResponse{
		Params: params,
	}, nil
}
