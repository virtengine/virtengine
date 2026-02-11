package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// schemaVersion10 is the schema version for GDPR export data
const schemaVersion10 = "1.0"

// ============================================================================
// GDPR Data Portability Keeper Methods
// ============================================================================
// Implements GDPR Article 20 - Right to Data Portability
// Reference: https://gdpr-info.eu/art-20-gdpr/

// exportRequestStore is the storage format for export requests
type exportRequestStore struct {
	Version          uint32   `json:"version"`
	RequestID        string   `json:"request_id"`
	RequesterAddress string   `json:"requester_address"`
	Categories       []string `json:"categories"`
	Format           string   `json:"format"`
	Status           string   `json:"status"`
	RequestedAt      int64    `json:"requested_at"`
	RequestedAtBlock int64    `json:"requested_at_block"`
	CompletedAt      *int64   `json:"completed_at,omitempty"`
	CompletedAtBlock *int64   `json:"completed_at_block,omitempty"`
	DeadlineAt       int64    `json:"deadline_at"`
	ExpiresAt        *int64   `json:"expires_at,omitempty"`
	ExportDataHash   []byte   `json:"export_data_hash,omitempty"`
	ExportSize       uint64   `json:"export_size,omitempty"`
	ErrorDetails     string   `json:"error_details,omitempty"`
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

	if err := k.RecordAuditEvent(ctx, types.AuditEventTypeGDPRPortability, requesterAddress.String(), map[string]interface{}{
		"request_id": request.RequestID,
		"categories": categories,
		"format":     format,
	}); err != nil {
		return nil, err
	}

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
		_ = k.SetExportRequest(ctx, request)
		return nil, err
	}

	// Generate the export package
	dataPackage := k.generateDataPackage(ctx, requesterAddr, &request)

	// Calculate checksum
	dataBytes, err := json.Marshal(dataPackage)
	if err != nil {
		request.MarkFailed(err.Error())
		_ = k.SetExportRequest(ctx, request)
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
) *types.PortableDataPackage {
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
				_ = k.exportCategory(ctx, address, cat, pkg)
			}
		}
	}

	return pkg
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
	case types.ExportCategoryEscrow:
		return k.exportEscrowData(ctx, address, pkg)
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

	pkg.Metadata.SchemaVersions[types.ExportCategoryIdentity] = schemaVersion10
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
		portableConsent := types.PortableScopeConsent(sc)
		pkg.Consent.ScopeConsents = append(pkg.Consent.ScopeConsents, portableConsent)
	}

	events := k.GetConsentEventsBySubject(ctx, address)
	for _, event := range events {
		pkg.Consent.ConsentHistory = append(pkg.Consent.ConsentHistory, types.PortableConsentEvent{
			EventType:   string(event.EventType),
			ScopeID:     event.ScopeID,
			Timestamp:   event.OccurredAt,
			BlockHeight: event.BlockHeight,
			Details:     event.Details,
		})
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryConsent] = schemaVersion10
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

	pkg.Metadata.SchemaVersions[types.ExportCategoryVerificationHistory] = schemaVersion10
	return nil
}

// exportTransactionData exports transaction history
// Note: This requires access to the bank/auth modules which may not be available here
func (k Keeper) exportTransactionData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	// Use activity records as a deterministic on-chain transaction ledger surrogate.
	records, err := k.GetRecentActivity(ctx, address.String(), 1000)
	if err != nil {
		return err
	}

	transactions := make([]types.PortableTransaction, 0)
	for _, record := range records {
		if record.ActivityType != types.ActivityTypeTransaction {
			continue
		}
		txHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d", record.Address, record.Timestamp.UnixNano(), record.BlockHeight)))
		transactions = append(transactions, types.PortableTransaction{
			TxHash:      hex.EncodeToString(txHash[:]),
			TxType:      string(record.ActivityType),
			Timestamp:   record.Timestamp,
			BlockHeight: record.BlockHeight,
			Status:      "recorded",
		})
	}

	pkg.Transactions = &types.PortableTransactionData{
		TotalTransactions: len(transactions),
		Transactions:      transactions,
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryTransactions] = schemaVersion10
	return nil
}

// exportMarketplaceData exports marketplace activity
// Note: This requires access to the market module
func (k Keeper) exportMarketplaceData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	data := &types.PortableMarketplaceData{
		Orders: make([]types.PortableOrder, 0),
		Bids:   make([]types.PortableBid, 0),
		Leases: make([]types.PortableLease, 0),
	}

	if k.marketKeeper != nil {
		k.marketKeeper.WithOrders(ctx, func(order markettypes.Order) bool {
			if order.ID.Owner != address.String() {
				return false
			}
			specs := make(map[string]interface{})
			if bz, err := json.Marshal(order.Spec); err == nil {
				_ = json.Unmarshal(bz, &specs)
			}
			data.Orders = append(data.Orders, types.PortableOrder{
				OrderID:        order.ID.String(),
				OrderType:      "order",
				Status:         order.State.String(),
				CreatedAt:      time.Unix(order.CreatedAt, 0),
				BlockHeight:    order.CreatedAt,
				Specifications: specs,
			})
			return false
		})

		k.marketKeeper.WithBids(ctx, func(bid markettypes.Bid) bool {
			if bid.ID.Provider != address.String() {
				return false
			}
			data.Bids = append(data.Bids, types.PortableBid{
				BidID:       bid.ID.String(),
				OrderID:     bid.ID.OrderID().String(),
				Provider:    bid.ID.Provider,
				Status:      bid.State.String(),
				CreatedAt:   time.Unix(bid.CreatedAt, 0),
				BlockHeight: bid.CreatedAt,
				Price:       bid.Price.String(),
			})
			return false
		})

		k.marketKeeper.WithLeases(ctx, func(lease mv1.Lease) bool {
			if lease.ID.Owner != address.String() && lease.ID.Provider != address.String() {
				return false
			}
			var closedAt *time.Time
			if lease.ClosedOn > 0 {
				t := time.Unix(lease.ClosedOn, 0)
				closedAt = &t
			}
			data.Leases = append(data.Leases, types.PortableLease{
				LeaseID:   lease.ID.String(),
				Provider:  lease.ID.Provider,
				Status:    lease.State.String(),
				CreatedAt: time.Unix(lease.CreatedAt, 0),
				ClosedAt:  closedAt,
				Price:     lease.Price.String(),
			})
			return false
		})
	}

	data.TotalOrders = len(data.Orders)
	data.TotalBids = len(data.Bids)
	data.TotalLeases = len(data.Leases)

	pkg.Marketplace = data
	pkg.Metadata.SchemaVersions[types.ExportCategoryMarketplace] = schemaVersion10
	return nil
}

// exportEscrowData exports escrow accounts and payments.
func (k Keeper) exportEscrowData(ctx sdk.Context, address sdk.AccAddress, pkg *types.PortableDataPackage) error {
	data := &types.PortableEscrowData{
		Accounts: make([]types.PortableEscrowAccount, 0),
		Payments: make([]types.PortableEscrowPayment, 0),
	}

	if k.escrowKeeper != nil {
		k.escrowKeeper.WithAccounts(ctx, func(account etypes.Account) bool {
			if account.State.Owner != address.String() {
				return false
			}
			transferred := make([]string, 0)
			if len(account.State.Transferred) > 0 {
				transferred = append(transferred, account.State.Transferred.String())
			}
			funds := make([]string, 0, len(account.State.Funds))
			for _, fund := range account.State.Funds {
				funds = append(funds, fmt.Sprintf("%v", fund))
			}
			deposits := make([]string, 0, len(account.State.Deposits))
			for _, dep := range account.State.Deposits {
				deposits = append(deposits, fmt.Sprintf("%v", dep))
			}

			data.Accounts = append(data.Accounts, types.PortableEscrowAccount{
				AccountID:   fmt.Sprintf("%v", account.ID),
				Owner:       account.State.Owner,
				State:       account.State.State.String(),
				Deposits:    deposits,
				Funds:       funds,
				SettledAt:   account.State.SettledAt,
				Transferred: transferred,
			})
			return false
		})

		k.escrowKeeper.WithPayments(ctx, func(payment etypes.Payment) bool {
			if payment.State.Owner != address.String() {
				return false
			}
			data.Payments = append(data.Payments, types.PortableEscrowPayment{
				PaymentID: fmt.Sprintf("%v", payment.ID),
				Owner:     payment.State.Owner,
				State:     payment.State.State.String(),
				Rate:      payment.State.Rate.String(),
				Balance:   payment.State.Balance.String(),
				Unsettled: payment.State.Unsettled.String(),
				Withdrawn: payment.State.Withdrawn.String(),
			})
			return false
		})
	}

	data.TotalAccounts = len(data.Accounts)
	data.TotalPayments = len(data.Payments)

	pkg.Escrow = data
	pkg.Metadata.SchemaVersions[types.ExportCategoryEscrow] = schemaVersion10
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
		TotalDelegations:     len(delegations),
		Delegations:          make([]types.PortableDelegation, 0),
		StakingDelegations:   make([]types.PortableStakingDelegation, 0),
		UnbondingDelegations: make([]types.PortableUnbondingDelegation, 0),
		Redelegations:        make([]types.PortableRedelegation, 0),
		Rewards:              make([]types.PortableDelegationReward, 0),
		SlashingEvents:       make([]types.PortableDelegationSlashingEvent, 0),
	}

	for _, del := range delegations {
		// Convert DelegationPermission slice to string slice
		permissions := make([]string, len(del.Permissions))
		for i, p := range del.Permissions {
			permissions[i] = p.String()
		}

		// Convert ExpiresAt value to pointer
		expiresAt := del.ExpiresAt

		portableDel := types.PortableDelegation{
			DelegationID:     del.DelegationID,
			DelegatorAddress: del.DelegatorAddress,
			DelegateAddress:  del.DelegateAddress,
			Permissions:      permissions,
			Status:           del.Status.String(),
			CreatedAt:        del.CreatedAt,
			ExpiresAt:        &expiresAt,
			RevokedAt:        del.RevokedAt,
		}
		pkg.Delegations.Delegations = append(pkg.Delegations.Delegations, portableDel)
	}

	if k.delegationKeeper != nil {
		stakingDelegations := k.delegationKeeper.GetDelegatorDelegations(ctx, address.String())
		for _, del := range stakingDelegations {
			pkg.Delegations.StakingDelegations = append(pkg.Delegations.StakingDelegations, types.PortableStakingDelegation{
				DelegatorAddress: del.DelegatorAddress,
				ValidatorAddress: del.ValidatorAddress,
				Shares:           del.Shares,
				CreatedAt:        del.CreatedAt,
				Status:           string(delegationtypes.DelegationStatusActive),
			})
		}

		unbondings := k.delegationKeeper.GetDelegatorUnbondingDelegations(ctx, address.String())
		for _, ub := range unbondings {
			for _, entry := range ub.Entries {
				pkg.Delegations.UnbondingDelegations = append(pkg.Delegations.UnbondingDelegations, types.PortableUnbondingDelegation{
					ID:               ub.ID,
					DelegatorAddress: ub.DelegatorAddress,
					ValidatorAddress: ub.ValidatorAddress,
					CompletionTime:   entry.CompletionTime,
					Balance:          entry.Balance,
				})
			}
		}

		redelegations := k.delegationKeeper.GetDelegatorRedelegations(ctx, address.String())
		for _, redel := range redelegations {
			for _, entry := range redel.Entries {
				pkg.Delegations.Redelegations = append(pkg.Delegations.Redelegations, types.PortableRedelegation{
					ID:               redel.ID,
					DelegatorAddress: redel.DelegatorAddress,
					ValidatorSrc:     redel.ValidatorSrcAddress,
					ValidatorDst:     redel.ValidatorDstAddress,
					CompletionTime:   entry.CompletionTime,
					Balance:          entry.InitialBalance,
				})
			}
		}

		rewards := k.delegationKeeper.GetDelegatorUnclaimedRewards(ctx, address.String())
		for _, reward := range rewards {
			pkg.Delegations.Rewards = append(pkg.Delegations.Rewards, types.PortableDelegationReward{
				DelegatorAddress: reward.DelegatorAddress,
				ValidatorAddress: reward.ValidatorAddress,
				Epoch:            reward.EpochNumber,
				Amount:           reward.Reward,
				Claimed:          reward.Claimed,
			})
		}

		slashingEvents := k.delegationKeeper.GetDelegatorSlashingEvents(ctx, address.String())
		for _, event := range slashingEvents {
			pkg.Delegations.SlashingEvents = append(pkg.Delegations.SlashingEvents, types.PortableDelegationSlashingEvent{
				ID:               event.ID,
				DelegatorAddress: event.DelegatorAddress,
				ValidatorAddress: event.ValidatorAddress,
				BlockHeight:      event.BlockHeight,
				Reason:           event.SlashFraction,
				Penalty:          event.SlashAmount,
			})
		}
	}

	pkg.Metadata.SchemaVersions[types.ExportCategoryDelegations] = schemaVersion10
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
	key := make([]byte, 0, len(prefixExportRequestByAddress)+len(address)+len(requestID)+1)
	key = append(key, prefixExportRequestByAddress...)
	key = append(key, []byte(address)...)
	key = append(key, byte(0x00))
	key = append(key, []byte(requestID)...)
	return key
}

func exportRequestByAddressPrefixKey(address string) []byte {
	key := make([]byte, 0, len(prefixExportRequestByAddress)+len(address)+1)
	key = append(key, prefixExportRequestByAddress...)
	key = append(key, []byte(address)...)
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
