package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// GDPR Data Portability Keeper Methods
// ============================================================================
// Implements GDPR Article 20 - Right to Data Portability
// Reference: https://gdpr-info.eu/art-20-gdpr/

// exportRequestStore is the storage format for export requests
type exportRequestStore struct {
	Version          uint32  `json:"version"`
	RequestID        string  `json:"request_id"`
	RequesterAddress string  `json:"requester_address"`
	Categories       []string `json:"categories"`
	Format           string  `json:"format"`
	Status           string  `json:"status"`
	RequestedAt      int64   `json:"requested_at"`
	RequestedAtBlock int64   `json:"requested_at_block"`
	CompletedAt      *int64  `json:"completed_at,omitempty"`
	CompletedAtBlock *int64  `json:"completed_at_block,omitempty"`
	DeadlineAt       int64   `json:"deadline_at"`
	ExpiresAt        *int64  `json:"expires_at,omitempty"`
	ExportDataHash   []byte  `json:"export_data_hash,omitempty"`
	ExportSize       uint64  `json:"export_size,omitempty"`
	ErrorDetails     string  `json:"error_details,omitempty"`
}

// SubmitExportRequest submits a new GDPR data portability export request
func (k Keeper) SubmitExportRequest(
	ctx sdk.Context,
	requesterAddress sdk.AccAddress,
	categories []types.ExportCategory,
	format types.ExportFormat,
) (*types.PortabilityExportRequest, error) {
	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Generate unique request ID
	requestID := generateExportRequestID(requesterAddress.String(), now, blockHeight)

	// Create the request
	request := types.NewPortabilityExportRequest(
		requestID,
		requesterAddress.String(),
		categories,
		format,
		now,
		blockHeight,
	)

	// Validate
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Store the request
	if err := k.SetExportRequest(ctx, *request); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExportRequested,
			sdk.NewAttribute(types.AttributeKeyExportRequestID, request.RequestID),
			sdk.NewAttribute(types.AttributeKeyRequesterAddress, requesterAddress.String()),
			sdk.NewAttribute(types.AttributeKeyExportFormat, string(format)),
		),
	)

	k.Logger(ctx).Info("GDPR data export request submitted",
		"request_id", request.RequestID,
		"requester", requesterAddress.String(),
		"categories", categories,
		"format", format)

	return request, nil
}

// ProcessExportRequest processes a pending export request and generates the export
func (k Keeper) ProcessExportRequest(ctx sdk.Context, requestID string) (*types.PortableDataPackage, error) {
	request, found := k.GetExportRequest(ctx, requestID)
	if !found {
		return nil, types.ErrInvalidParams.Wrapf("export request not found: %s", requestID)
	}

	if request.Status != types.ExportStatusPending {
		return nil, types.ErrInvalidParams.Wrapf("request is not pending: %s", request.Status)
	}

	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Mark as processing
	request.MarkProcessing()
	if err := k.SetExportRequest(ctx, request); err != nil {
		return nil, err
	}

	// Get requester address
	requesterAddr, err := sdk.AccAddressFromBech32(request.RequesterAddress)
	if err != nil {
		request.MarkFailed(err.Error())
		k.SetExportRequest(ctx, request)
		return nil, err
	}

	// Generate the export package
	dataPackage, err := k.generateDataPackage(ctx, requesterAddr, &request)
	if err != nil {
		request.MarkFailed(err.Error())
		k.SetExportRequest(ctx, request)
		return nil, err
	}

	// Calculate checksum
	dataBytes, err := json.Marshal(dataPackage)
	if err != nil {
		request.MarkFailed(err.Error())
		k.SetExportRequest(ctx, request)
		return nil, err
	}

	checksum := sha256.Sum256(dataBytes)
	dataPackage.Metadata.ChecksumSHA256 = hex.EncodeToString(checksum[:])

	// Mark as completed
	request.MarkCompleted(now, blockHeight, checksum[:], uint64(len(dataBytes)))
	if err := k.SetExportRequest(ctx, request); err != nil {
		return nil, err
	}

	// Emit completion event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExportCompleted,
			sdk.NewAttribute(types.AttributeKeyExportRequestID, request.RequestID),
			sdk.NewAttribute(types.AttributeKeyExportSize, fmt.Sprintf("%d", request.ExportSize)),
			sdk.NewAttribute(types.AttributeKeyExportChecksum, dataPackage.Metadata.ChecksumSHA256),
		),
	)

	k.Logger(ctx).Info("GDPR data export completed",
		"request_id", request.RequestID,
		"size", request.ExportSize,
		"checksum", dataPackage.Metadata.ChecksumSHA256)

	return dataPackage, nil
}

// generateDataPackage generates the portable data package
func (k Keeper) generateDataPackage(
	ctx sdk.Context,
	address sdk.AccAddress,
	request *types.PortabilityExportRequest,
) (*types.PortableDataPackage, error) {
	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Initialize the package
	pkg := &types.PortableDataPackage{
		Metadata: types.ExportMetadata{
			ExportVersion:      types.PortabilityExportVersion,
			ExportRequestID:    request.RequestID,
			DataSubjectAddress: address.String(),
			ExportedAt:         now,
			ExportedAtBlock:    blockHeight,
			CategoriesIncluded: request.Categories,
			Format:             request.Format,
			DataController: types.DataControllerInfo{
				Name:    "DET-IO Pty. Ltd. (VirtEngine)",
				Contact: "dpo@virtengine.com",
				Website: "https://virtengine.com",
			},
			SchemaVersions: make(map[types.ExportCategory]string),
		},
	}

	// Process each requested category
	for _, cat := range request.Categories {
		if err := k.exportCategory(ctx, address, cat, pkg); err != nil {
			k.Logger(ctx).Error("failed to export category",
				"category", cat,
				"error", err)
			// Continue with other categories
		}
	}

	// If "all" category, export everything
	if request.HasCategory(types.ExportCategoryAll) {
		for _, cat := range types.AllExportCategories() {
			if cat != types.ExportCategoryAll {
				k.exportCategory(ctx, address, cat, pkg)
			}
		}
	}

	return pkg, nil
}

// exportCategory exports a specific data category
func (k Keeper) exportCategory(
	ctx sdk.Context,
	address sdk.AccAddress,
	category types.ExportCategory,
	pkg *types.PortableDataPackage,
) error {
	switch category {
	case types.ExportCategoryIdentity:
		return k.exportIdentityData(ctx, address, pkg)
	case types.ExportCategoryConsent:
		return k.exportConsentData(ctx, address, pkg)
	case types.ExportCategoryVerificationHistory:
		return k.exportVerificationHistory(ctx, address, pkg)
	case types.ExportCategoryTransactions:
		return k.exportTransactionData(ctx, address, pkg)
	case types.ExportCategoryMarketplace:
		return k.exportMarketplaceData(ctx, address, pkg)
	case types.ExportCategoryDelegations:
		return k.exportDelegationData(ctx, address, pkg)
	}
	return nil
}

// exportIdentityData exports identity and wallet data
func (k Keeper) exportIdentityData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	wallet, found := k.GetWallet(ctx, address)
	if !found {
		return nil // No wallet data
	}

	pkg.Identity = &types.PortableIdentityData{
		WalletAddress:     wallet.AccountAddress,
		WalletID:          wallet.WalletID,
		WalletStatus:      string(wallet.Status),
		CreatedAt:         wallet.CreatedAt,
		CreatedAtBlock:    0, // Block height not stored in IdentityWallet
		VerificationLevel: string(wallet.Tier),
		ActiveScopes:      make([]types.PortableScopeInfo, 0),
	}

	// Export trust score from current wallet score
	pkg.Identity.TrustScore = &types.PortableTrustScore{
		Score:       float64(wallet.CurrentScore),
		Confidence:  1.0, // Score is deterministic
		LastUpdated: wallet.UpdatedAt,
	}

	// Export active scopes using WithScopes
	k.WithScopes(ctx, address, func(scope types.IdentityScope) bool {
		pkg.Identity.ActiveScopes = append(pkg.Identity.ActiveScopes, types.PortableScopeInfo{
			ScopeID:   scope.ScopeID,
			ScopeType: string(scope.ScopeType),
			Status:    string(scope.Status),
			CreatedAt: scope.UploadedAt,
			ExpiresAt: scope.ExpiresAt,
		})
		return false // continue iteration
	})

	pkg.Metadata.SchemaVersions[types.ExportCategoryIdentity] = "1.0"
	return nil
}

// exportConsentData exports consent records
func (k Keeper) exportConsentData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	wallet, found := k.GetWallet(ctx, address)
	if !found {
		return nil
	}

	pkg.Consent = &types.PortableConsentData{
		GlobalSettings: types.PortableGlobalConsent{
			ShareWithProviders:         wallet.ConsentSettings.ShareWithProviders,
			ShareForVerification:       wallet.ConsentSettings.ShareForVerification,
			AllowReVerification:        wallet.ConsentSettings.AllowReVerification,
			AllowDerivedFeatureSharing: wallet.ConsentSettings.AllowDerivedFeatureSharing,
			GlobalExpiresAt:            wallet.ConsentSettings.GlobalExpiresAt,
			LastUpdatedAt:              wallet.ConsentSettings.LastUpdatedAt,
			ConsentVersion:             wallet.ConsentSettings.ConsentVersion,
		},
		ScopeConsents:  make([]types.PortableScopeConsent, 0),
		ConsentHistory: make([]types.PortableConsentEvent, 0),
	}

	// Export scope consents
	for _, sc := range wallet.ConsentSettings.ScopeConsents {
		portableConsent := types.PortableScopeConsent{
			ScopeID:            sc.ScopeID,
			Granted:            sc.Granted,
			GrantedAt:          sc.GrantedAt,
			RevokedAt:          sc.RevokedAt,
			ExpiresAt:          sc.ExpiresAt,
			Purpose:            sc.Purpose,
			GrantedToProviders: sc.GrantedToProviders,
			Restrictions:       sc.Restrictions,
		}
		pkg.Consent.ScopeConsents = append(pkg.Consent.ScopeConsents, portableConsent)
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryConsent] = "1.0"
	return nil
}

// exportVerificationHistory exports verification history
func (k Keeper) exportVerificationHistory(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	// Get derived feature records as verification history
	records := k.GetDerivedFeatureRecordsByAccount(ctx, address)

	pkg.VerificationHistory = &types.PortableVerificationData{
		TotalVerifications:      len(records),
		SuccessfulVerifications: 0,
		FailedVerifications:     0,
		Verifications:           make([]types.PortableVerificationRecord, 0),
	}

	for _, record := range records {
		// Check status to determine if verification passed
		passed := record.Status == types.VerificationResultStatusSuccess
		status := string(record.Status)
		if passed {
			pkg.VerificationHistory.SuccessfulVerifications++
		} else {
			pkg.VerificationHistory.FailedVerifications++
		}

		portableRecord := types.PortableVerificationRecord{
			VerificationID:   record.RecordID,
			VerificationType: "derived_feature",
			Status:           status,
			Score:            float64(record.Score),
			Confidence:       float64(record.Confidence),
			Timestamp:        record.ComputedAt,
			BlockHeight:      record.BlockHeight,
			MLModelVersion:   record.ModelVersion,
			PipelineVersion:  "", // Not available in this record type
		}
		pkg.VerificationHistory.Verifications = append(pkg.VerificationHistory.Verifications, portableRecord)
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryVerificationHistory] = "1.0"
	return nil
}

// exportTransactionData exports transaction history
// Note: This requires access to the bank/auth modules which may not be available here
func (k Keeper) exportTransactionData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	// Transaction data would be exported from chain history
	// This is a placeholder - actual implementation would query the chain
	pkg.Transactions = &types.PortableTransactionData{
		TotalTransactions: 0,
		Transactions:      make([]types.PortableTransaction, 0),
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryTransactions] = "1.0"
	return nil
}

// exportMarketplaceData exports marketplace activity
// Note: This requires access to the market module
func (k Keeper) exportMarketplaceData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	// Marketplace data would be exported from the market module
	// This is a placeholder - actual implementation would query the market keeper
	pkg.Marketplace = &types.PortableMarketplaceData{
		TotalOrders: 0,
		TotalLeases: 0,
		Orders:      make([]types.PortableOrder, 0),
		Leases:      make([]types.PortableLease, 0),
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryMarketplace] = "1.0"
	return nil
}

// exportDelegationData exports delegation relationships
func (k Keeper) exportDelegationData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	// Get delegations
	delegations, err := k.ListDelegationsForDelegator(ctx, address, false)
	if err != nil {
		return err
	}

	pkg.Delegations = &types.PortableDelegationData{
		TotalDelegations: len(delegations),
		Delegations:      make([]types.PortableDelegation, 0),
	}

	for _, del := range delegations {
		// Convert DelegationPermission slice to string slice
		permissions := make([]string, len(del.Permissions))
		for i, p := range del.Permissions {
			permissions[i] = string(p)
		}

		// Convert ExpiresAt value to pointer
		expiresAt := del.ExpiresAt

		portableDel := types.PortableDelegation{
			DelegationID:     del.DelegationID,
			DelegatorAddress: del.DelegatorAddress,
			DelegateAddress:  del.DelegateAddress,
			Permissions:      permissions,
			Status:           string(del.Status),
			CreatedAt:        del.CreatedAt,
			ExpiresAt:        &expiresAt,
			RevokedAt:        del.RevokedAt,
		}
		pkg.Delegations.Delegations = append(pkg.Delegations.Delegations, portableDel)
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryDelegations] = "1.0"
	return nil
}

// ============================================================================
// Storage Methods
// ============================================================================

// SetExportRequest stores an export request
func (k Keeper) SetExportRequest(ctx sdk.Context, request types.PortabilityExportRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	rs := exportRequestToStore(&request)
	bz, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	store.Set(exportRequestKey(request.RequestID), bz)

	// Index by requester address
	store.Set(exportRequestByAddressKey(request.RequesterAddress, request.RequestID), []byte{1})

	return nil
}

// GetExportRequest retrieves an export request
func (k Keeper) GetExportRequest(ctx sdk.Context, requestID string) (types.PortabilityExportRequest, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(exportRequestKey(requestID))
	if bz == nil {
		return types.PortabilityExportRequest{}, false
	}

	var rs exportRequestStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.PortabilityExportRequest{}, false
	}

	return *exportRequestFromStore(&rs), true
}

// GetExportRequestsByAddress retrieves all export requests for an address
func (k Keeper) GetExportRequestsByAddress(ctx sdk.Context, address sdk.AccAddress) []types.PortabilityExportRequest {
	store := ctx.KVStore(k.skey)
	prefix := exportRequestByAddressPrefixKey(address.String())

	iter := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iter.Close()

	var requests []types.PortabilityExportRequest
	for ; iter.Valid(); iter.Next() {
		requestID := extractExportRequestIDFromKey(iter.Key(), prefix)
		if request, found := k.GetExportRequest(ctx, requestID); found {
			requests = append(requests, request)
		}
	}

	return requests
}

// ============================================================================
// Key Generation Functions
// ============================================================================

var (
	prefixExportRequest          = []byte{0x60}
	prefixExportRequestByAddress = []byte{0x61}
)

func exportRequestKey(requestID string) []byte {
	return append(prefixExportRequest, []byte(requestID)...)
}

func exportRequestByAddressKey(address string, requestID string) []byte {
	key := append(prefixExportRequestByAddress, []byte(address)...)
	key = append(key, byte(0x00))
	key = append(key, []byte(requestID)...)
	return key
}

func exportRequestByAddressPrefixKey(address string) []byte {
	key := append(prefixExportRequestByAddress, []byte(address)...)
	key = append(key, byte(0x00))
	return key
}

func extractExportRequestIDFromKey(key []byte, prefix []byte) string {
	if len(key) <= len(prefix) {
		return ""
	}
	return string(key[len(prefix):])
}

func generateExportRequestID(address string, timestamp time.Time, blockHeight int64) string {
	data := fmt.Sprintf("export:%s:%d:%d", address, timestamp.UnixNano(), blockHeight)
	hash := sha256.Sum256([]byte(data))
	return "export_" + hex.EncodeToString(hash[:8])
}

// ============================================================================
// Storage Conversion Functions
// ============================================================================

func exportRequestToStore(r *types.PortabilityExportRequest) *exportRequestStore {
	categories := make([]string, len(r.Categories))
	for i, c := range r.Categories {
		categories[i] = string(c)
	}

	rs := &exportRequestStore{
		Version:          r.Version,
		RequestID:        r.RequestID,
		RequesterAddress: r.RequesterAddress,
		Categories:       categories,
		Format:           string(r.Format),
		Status:           string(r.Status),
		RequestedAt:      r.RequestedAt.UnixNano(),
		RequestedAtBlock: r.RequestedAtBlock,
		DeadlineAt:       r.DeadlineAt.UnixNano(),
		ExportDataHash:   r.ExportDataHash,
		ExportSize:       r.ExportSize,
		ErrorDetails:     r.ErrorDetails,
	}

	if r.CompletedAt != nil {
		ts := r.CompletedAt.UnixNano()
		rs.CompletedAt = &ts
	}
	if r.CompletedAtBlock != nil {
		rs.CompletedAtBlock = r.CompletedAtBlock
	}
	if r.ExpiresAt != nil {
		ts := r.ExpiresAt.UnixNano()
		rs.ExpiresAt = &ts
	}

	return rs
}

func exportRequestFromStore(rs *exportRequestStore) *types.PortabilityExportRequest {
	categories := make([]types.ExportCategory, len(rs.Categories))
	for i, c := range rs.Categories {
		categories[i] = types.ExportCategory(c)
	}

	r := &types.PortabilityExportRequest{
		Version:          rs.Version,
		RequestID:        rs.RequestID,
		RequesterAddress: rs.RequesterAddress,
		Categories:       categories,
		Format:           types.ExportFormat(rs.Format),
		Status:           types.ExportRequestStatus(rs.Status),
		RequestedAt:      time.Unix(0, rs.RequestedAt),
		RequestedAtBlock: rs.RequestedAtBlock,
		DeadlineAt:       time.Unix(0, rs.DeadlineAt),
		ExportDataHash:   rs.ExportDataHash,
		ExportSize:       rs.ExportSize,
		ErrorDetails:     rs.ErrorDetails,
	}

	if rs.CompletedAt != nil {
		ts := time.Unix(0, *rs.CompletedAt)
		r.CompletedAt = &ts
	}
	if rs.CompletedAtBlock != nil {
		r.CompletedAtBlock = rs.CompletedAtBlock
	}
	if rs.ExpiresAt != nil {
		ts := time.Unix(0, *rs.ExpiresAt)
		r.ExpiresAt = &ts
	}

	return r
}
