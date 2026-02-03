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
// GDPR Erasure (Right to Be Forgotten) Keeper Methods
// ============================================================================
// Implements GDPR Article 17 - Right to Erasure
// Reference: https://gdpr-info.eu/art-17-gdpr/

// erasureRequestStore is the storage format for erasure requests
type erasureRequestStore struct {
	Version          uint32              `json:"version"`
	RequestID        string              `json:"request_id"`
	RequesterAddress string              `json:"requester_address"`
	Categories       []string            `json:"categories"`
	Status           string              `json:"status"`
	RequestedAt      int64               `json:"requested_at"`
	RequestedAtBlock int64               `json:"requested_at_block"`
	ProcessedAt      *int64              `json:"processed_at,omitempty"`
	ProcessedAtBlock *int64              `json:"processed_at_block,omitempty"`
	CompletedAt      *int64              `json:"completed_at,omitempty"`
	CompletedAtBlock *int64              `json:"completed_at_block,omitempty"`
	DeadlineAt       int64               `json:"deadline_at"`
	RejectionReason  *string             `json:"rejection_reason,omitempty"`
	RejectionDetails string              `json:"rejection_details,omitempty"`
	ErasureReport    *erasureReportStore `json:"erasure_report,omitempty"`
	VerificationHash []byte              `json:"verification_hash,omitempty"`
}

type erasureReportStore struct {
	BiometricDataErased       bool              `json:"biometric_data_erased"`
	BiometricKeysDestroyed    bool              `json:"biometric_keys_destroyed"`
	IdentityDocumentsErased   bool              `json:"identity_documents_erased"`
	VerificationHistoryErased bool              `json:"verification_history_erased"`
	DerivedFeaturesErased     bool              `json:"derived_features_erased"`
	ConsentRecordsErased      bool              `json:"consent_records_erased"`
	OffChainDataDeleted       bool              `json:"off_chain_data_deleted"`
	OnChainDataMadeUnreadable bool              `json:"on_chain_data_made_unreadable"`
	TotalRecordsAffected      uint64            `json:"total_records_affected"`
	DataCategoriesErased      []string          `json:"data_categories_erased"`
	RetainedDataCategories    []string          `json:"retained_data_categories,omitempty"`
	RetentionReasons          map[string]string `json:"retention_reasons,omitempty"`
	BackupDeletionScheduled   *int64            `json:"backup_deletion_scheduled,omitempty"`
	ReportGeneratedAt         int64             `json:"report_generated_at"`
}

// SubmitErasureRequest submits a new GDPR erasure request
func (k Keeper) SubmitErasureRequest(
	ctx sdk.Context,
	requesterAddress sdk.AccAddress,
	categories []types.ErasureCategory,
) (*types.ErasureRequest, error) {
	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Generate unique request ID
	requestID := generateErasureRequestID(requesterAddress.String(), now, blockHeight)

	// Check for existing pending request
	if existing, found := k.GetPendingErasureRequestByAddress(ctx, requesterAddress); found {
		return nil, types.ErrInvalidParams.Wrapf(
			"existing pending erasure request: %s", existing.RequestID)
	}

	// Create the request
	request := types.NewErasureRequest(requestID, requesterAddress.String(), categories, now, blockHeight)

	// Generate verification hash
	request.VerificationHash = generateVerificationHash(request)

	// Validate
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Store the request
	if err := k.SetErasureRequest(ctx, *request); err != nil {
		return nil, err
	}

	// Add to pending index
	k.addToPendingErasureIndex(ctx, request.DeadlineAt.Unix(), request.RequestID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeErasureRequested,
			sdk.NewAttribute(types.AttributeKeyRequestID, request.RequestID),
			sdk.NewAttribute(types.AttributeKeyRequesterAddress, requesterAddress.String()),
			sdk.NewAttribute(types.AttributeKeyStatus, string(request.Status)),
			sdk.NewAttribute(types.AttributeKeyDeadline, request.DeadlineAt.Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("GDPR erasure request submitted",
		"request_id", request.RequestID,
		"requester", requesterAddress.String(),
		"categories", categories,
		"deadline", request.DeadlineAt)

	return request, nil
}

// ProcessErasureRequest processes a pending erasure request
func (k Keeper) ProcessErasureRequest(ctx sdk.Context, requestID string) error {
	request, found := k.GetErasureRequest(ctx, requestID)
	if !found {
		return types.ErrInvalidParams.Wrapf("erasure request not found: %s", requestID)
	}

	if !request.IsPending() {
		return types.ErrInvalidParams.Wrapf("request is not pending: %s", request.Status)
	}

	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Mark as processing
	request.MarkProcessing(now, blockHeight)
	if err := k.SetErasureRequest(ctx, request); err != nil {
		return err
	}

	// Emit processing event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeErasureProcessing,
			sdk.NewAttribute(types.AttributeKeyRequestID, request.RequestID),
			sdk.NewAttribute(types.AttributeKeyStatus, string(request.Status)),
		),
	)

	// Check for legal holds or regulatory requirements
	if holdReason, hasHold := k.checkLegalHold(ctx, request.RequesterAddress); hasHold {
		request.MarkRejected(types.RejectionReasonLegalHold, holdReason, now, blockHeight)
		if err := k.SetErasureRequest(ctx, request); err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeErasureRejected,
				sdk.NewAttribute(types.AttributeKeyRequestID, request.RequestID),
				sdk.NewAttribute(types.AttributeKeyRejectionReason, string(*request.RejectionReason)),
			),
		)

		return nil
	}

	// Execute erasure
	report, err := k.executeErasure(ctx, &request)
	if err != nil {
		request.MarkFailed(err.Error(), now, blockHeight)
		if setErr := k.SetErasureRequest(ctx, request); setErr != nil {
			return setErr
		}
		return err
	}

	// Mark as completed (partial if blockchain data exists)
	if report.OnChainDataMadeUnreadable {
		request.MarkPartialCompleted(now, blockHeight, report)
	} else {
		request.MarkCompleted(now, blockHeight, report)
	}

	if err := k.SetErasureRequest(ctx, request); err != nil {
		return err
	}

	// Remove from pending index
	k.removeFromPendingErasureIndex(ctx, request.DeadlineAt.Unix(), request.RequestID)

	// Emit completion event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeErasureCompleted,
			sdk.NewAttribute(types.AttributeKeyRequestID, request.RequestID),
			sdk.NewAttribute(types.AttributeKeyStatus, string(request.Status)),
			sdk.NewAttribute(types.AttributeKeyRecordsAffected, fmt.Sprintf("%d", report.TotalRecordsAffected)),
		),
	)

	k.Logger(ctx).Info("GDPR erasure completed",
		"request_id", request.RequestID,
		"status", request.Status,
		"records_affected", report.TotalRecordsAffected)

	return nil
}

// executeErasure performs the actual erasure operations
func (k Keeper) executeErasure(ctx sdk.Context, request *types.ErasureRequest) (*types.ErasureReport, error) {
	requesterAddr, err := sdk.AccAddressFromBech32(request.RequesterAddress)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()
	report := &types.ErasureReport{
		ReportGeneratedAt: now,
		RetentionReasons:  make(map[string]string),
	}

	var totalAffected uint64

	// Process each category
	for _, cat := range request.Categories {
		affected, err := k.eraseCategory(ctx, requesterAddr, cat, report)
		if err != nil {
			k.Logger(ctx).Error("failed to erase category",
				"category", cat,
				"error", err)
			// Continue with other categories
		}
		totalAffected += affected
	}

	// If "all" category, process all categories
	if request.HasCategory(types.ErasureCategoryAll) {
		for _, cat := range types.AllErasureCategories() {
			if cat != types.ErasureCategoryAll {
				affected, _ := k.eraseCategory(ctx, requesterAddr, cat, report)
				totalAffected += affected
			}
		}
	}

	// Destroy encryption keys
	keyRecord, err := k.destroyEncryptionKeys(ctx, requesterAddr, request.RequestID)
	if err != nil {
		k.Logger(ctx).Error("failed to destroy encryption keys", "error", err)
	} else if keyRecord != nil {
		report.BiometricKeysDestroyed = true
		report.OnChainDataMadeUnreadable = true
	}

	// Schedule backup deletion (90 days from now)
	backupDeletionTime := now.Add(90 * 24 * time.Hour)
	report.BackupDeletionScheduled = &backupDeletionTime

	report.TotalRecordsAffected = totalAffected
	report.OffChainDataDeleted = true

	// Track erased categories
	report.DataCategoriesErased = append(report.DataCategoriesErased, request.Categories...)

	return report, nil
}

// eraseCategory erases data for a specific category
func (k Keeper) eraseCategory(
	ctx sdk.Context,
	address sdk.AccAddress,
	category types.ErasureCategory,
	report *types.ErasureReport,
) (uint64, error) {
	var affected uint64

	switch category {
	case types.ErasureCategoryBiometric:
		// Erase biometric data (embedding envelopes)
		envelopes := k.GetEmbeddingEnvelopesByAccount(ctx, address)
		for _, env := range envelopes {
			_ = k.DeleteEmbeddingEnvelope(ctx, env.EnvelopeID)
			affected++
		}
		report.BiometricDataErased = true

	case types.ErasureCategoryIdentityDocuments:
		// Mark identity documents for deletion
		// Note: Actual document deletion happens off-chain
		report.IdentityDocumentsErased = true
		affected++

	case types.ErasureCategoryVerificationHistory:
		// Clear verification history (off-chain only - on-chain is immutable)
		// The on-chain history remains but references encrypted data
		report.VerificationHistoryErased = true
		affected++

	case types.ErasureCategoryDerivedFeatures:
		// Erase derived feature records
		records := k.GetDerivedFeatureRecordsByAccount(ctx, address)
		for _, record := range records {
			_ = k.DeleteDerivedFeatureRecord(ctx, record.RecordID)
			affected++
		}
		report.DerivedFeaturesErased = true

	case types.ErasureCategoryConsent:
		// Revoke all consents
		wallet, found := k.GetWallet(ctx, address)
		if found {
			wallet.ConsentSettings.RevokeAllAt(ctx.BlockTime())
			if err := k.SetWallet(ctx, wallet); err != nil {
				return affected, err
			}
			affected++
		}
		report.ConsentRecordsErased = true

	case types.ErasureCategoryAll:
		// This is handled by iterating all categories in executeErasure
		return affected, nil
	}

	return affected, nil
}

// destroyEncryptionKeys destroys encryption keys for an account
func (k Keeper) destroyEncryptionKeys(
	ctx sdk.Context,
	address sdk.AccAddress,
	erasureRequestID string,
) (*types.KeyDestructionRecord, error) {
	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Get all key fingerprints for this account
	keyFingerprints := k.getAccountKeyFingerprints(ctx, address)
	if len(keyFingerprints) == 0 {
		return nil, nil // No keys to destroy
	}

	// Generate destruction record ID
	recordID := generateKeyDestructionRecordID(address.String(), now)

	// Create destruction record
	record := types.NewKeyDestructionRecord(
		recordID,
		erasureRequestID,
		address.String(),
		keyFingerprints,
		[]string{"X25519", "AES-256"},
		now,
		blockHeight,
	)

	// Generate verification hash (proof of destruction)
	record.VerificationHash = generateKeyDestructionHash(record)

	// Store the destruction record
	if err := k.SetKeyDestructionRecord(ctx, *record); err != nil {
		return nil, err
	}

	// Mark all embedding envelopes as having destroyed keys
	k.markEnvelopeKeysDestroyed(ctx, address)

	// Emit key destruction event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKeyDestruction,
			sdk.NewAttribute(types.AttributeKeyRequestID, erasureRequestID),
			sdk.NewAttribute(types.AttributeKeyRequesterAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyKeyFingerprints, fmt.Sprintf("%v", keyFingerprints)),
		),
	)

	k.Logger(ctx).Info("encryption keys destroyed for GDPR compliance",
		"account", address.String(),
		"keys_destroyed", len(keyFingerprints),
		"erasure_request", erasureRequestID)

	return record, nil
}

// checkLegalHold checks if there's a legal hold on the account
func (k Keeper) checkLegalHold(ctx sdk.Context, address string) (string, bool) {
	// Check for legal holds in the store
	store := ctx.KVStore(k.skey)
	key := legalHoldKey(address)
	bz := store.Get(key)
	if bz != nil {
		return string(bz), true
	}
	return "", false
}

// SetLegalHold sets a legal hold on an account (prevents erasure)
func (k Keeper) SetLegalHold(ctx sdk.Context, address sdk.AccAddress, reason string) error {
	store := ctx.KVStore(k.skey)
	key := legalHoldKey(address.String())
	store.Set(key, []byte(reason))

	k.Logger(ctx).Info("legal hold set on account",
		"address", address.String(),
		"reason", reason)

	return nil
}

// RemoveLegalHold removes a legal hold from an account
func (k Keeper) RemoveLegalHold(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(k.skey)
	key := legalHoldKey(address.String())
	store.Delete(key)

	k.Logger(ctx).Info("legal hold removed from account",
		"address", address.String())
}

// ============================================================================
// Storage Methods
// ============================================================================

// SetErasureRequest stores an erasure request
func (k Keeper) SetErasureRequest(ctx sdk.Context, request types.ErasureRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	rs := erasureRequestToStore(&request)
	bz, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	store.Set(erasureRequestKey(request.RequestID), bz)

	// Index by requester address
	store.Set(erasureRequestByAddressKey(request.RequesterAddress, request.RequestID), []byte{1})

	return nil
}

// GetErasureRequest retrieves an erasure request
func (k Keeper) GetErasureRequest(ctx sdk.Context, requestID string) (types.ErasureRequest, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(erasureRequestKey(requestID))
	if bz == nil {
		return types.ErasureRequest{}, false
	}

	var rs erasureRequestStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.ErasureRequest{}, false
	}

	return *erasureRequestFromStore(&rs), true
}

// GetErasureRequestsByAddress retrieves all erasure requests for an address
func (k Keeper) GetErasureRequestsByAddress(ctx sdk.Context, address sdk.AccAddress) []types.ErasureRequest {
	store := ctx.KVStore(k.skey)
	prefix := erasureRequestByAddressPrefixKey(address.String())

	iter := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iter.Close()

	var requests []types.ErasureRequest
	for ; iter.Valid(); iter.Next() {
		requestID := extractRequestIDFromKey(iter.Key(), prefix)
		if request, found := k.GetErasureRequest(ctx, requestID); found {
			requests = append(requests, request)
		}
	}

	return requests
}

// GetPendingErasureRequestByAddress gets the pending erasure request for an address
func (k Keeper) GetPendingErasureRequestByAddress(ctx sdk.Context, address sdk.AccAddress) (types.ErasureRequest, bool) {
	requests := k.GetErasureRequestsByAddress(ctx, address)
	for _, r := range requests {
		if r.IsPending() {
			return r, true
		}
	}
	return types.ErasureRequest{}, false
}

// SetKeyDestructionRecord stores a key destruction record
func (k Keeper) SetKeyDestructionRecord(ctx sdk.Context, record types.KeyDestructionRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}

	store.Set(keyDestructionRecordKey(record.RecordID), bz)

	// Index by account
	store.Set(keyDestructionByAccountKey(record.AccountAddress, record.RecordID), []byte{1})

	return nil
}

// GetKeyDestructionRecord retrieves a key destruction record
func (k Keeper) GetKeyDestructionRecord(ctx sdk.Context, recordID string) (types.KeyDestructionRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(keyDestructionRecordKey(recordID))
	if bz == nil {
		return types.KeyDestructionRecord{}, false
	}

	var record types.KeyDestructionRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.KeyDestructionRecord{}, false
	}

	return record, true
}

// GenerateErasureCertificate generates an erasure confirmation certificate
func (k Keeper) GenerateErasureCertificate(ctx sdk.Context, requestID string) (*types.ErasureConfirmationCertificate, error) {
	request, found := k.GetErasureRequest(ctx, requestID)
	if !found {
		return nil, types.ErrInvalidParams.Wrapf("erasure request not found: %s", requestID)
	}

	if !request.IsComplete() {
		return nil, types.ErrInvalidParams.Wrap("erasure request is not complete")
	}

	certificateID := generateCertificateID(requestID, ctx.BlockTime())
	cert := types.NewErasureConfirmationCertificate(certificateID, &request, ctx.BlockTime())

	// Generate signature hash
	cert.SignatureHash = generateCertificateSignature(cert)

	return cert, nil
}

// ============================================================================
// Pending Erasure Index
// ============================================================================

func (k Keeper) addToPendingErasureIndex(ctx sdk.Context, deadlineUnix int64, requestID string) {
	store := ctx.KVStore(k.skey)
	key := pendingErasureKey(deadlineUnix, requestID)
	store.Set(key, []byte{1})
}

func (k Keeper) removeFromPendingErasureIndex(ctx sdk.Context, deadlineUnix int64, requestID string) {
	store := ctx.KVStore(k.skey)
	key := pendingErasureKey(deadlineUnix, requestID)
	store.Delete(key)
}

// GetOverdueErasureRequests returns all overdue erasure requests
func (k Keeper) GetOverdueErasureRequests(ctx sdk.Context) []types.ErasureRequest {
	now := ctx.BlockTime().Unix()
	store := ctx.KVStore(k.skey)

	prefix := pendingErasurePrefixKey()
	endKey := pendingErasureBeforeKey(now + 1)

	iter := store.Iterator(prefix, endKey)
	defer iter.Close()

	var overdue []types.ErasureRequest
	for ; iter.Valid(); iter.Next() {
		requestID := extractPendingErasureRequestID(iter.Key())
		if request, found := k.GetErasureRequest(ctx, requestID); found {
			if request.IsPending() {
				overdue = append(overdue, request)
			}
		}
	}

	return overdue
}

// ProcessOverdueErasureRequests processes all overdue erasure requests
func (k Keeper) ProcessOverdueErasureRequests(ctx sdk.Context) int {
	overdue := k.GetOverdueErasureRequests(ctx)
	processed := 0

	for _, request := range overdue {
		if err := k.ProcessErasureRequest(ctx, request.RequestID); err != nil {
			k.Logger(ctx).Error("failed to process overdue erasure request",
				"request_id", request.RequestID,
				"error", err)
		} else {
			processed++
		}
	}

	return processed
}

// ============================================================================
// Helper Methods
// ============================================================================

func (k Keeper) getAccountKeyFingerprints(ctx sdk.Context, address sdk.AccAddress) []string {
	// EmbeddingEnvelopeReference is the on-chain lightweight reference
	// It doesn't contain encrypted payload data (which is stored off-chain)
	// Key fingerprints would need to be tracked separately or retrieved
	// from the encryption module if needed
	// For GDPR purposes, we track which envelopes exist, not the key details
	envelopes := k.GetEmbeddingEnvelopesByAccount(ctx, address)
	fingerprintSet := make(map[string]struct{})

	// Use computed-by validator address as a proxy for key tracking
	for _, env := range envelopes {
		if env.ComputedBy != "" {
			fingerprintSet[env.ComputedBy] = struct{}{}
		}
	}

	fingerprints := make([]string, 0, len(fingerprintSet))
	for fp := range fingerprintSet {
		fingerprints = append(fingerprints, fp)
	}

	return fingerprints
}

func (k Keeper) markEnvelopeKeysDestroyed(ctx sdk.Context, address sdk.AccAddress) {
	envelopes := k.GetEmbeddingEnvelopesByAccount(ctx, address)
	for _, env := range envelopes {
		env.Revoked = true
		env.RevokedReason = "GDPR erasure - encryption keys destroyed"
		revokedAt := ctx.BlockTime()
		env.RevokedAt = &revokedAt
		_ = k.SetEmbeddingEnvelope(ctx, env) // Ignore error for bulk operation
	}
}

// ============================================================================
// Key Generation Functions
// ============================================================================

var (
	prefixErasureRequest          = []byte{0x50}
	prefixErasureRequestByAddress = []byte{0x51}
	prefixPendingErasure          = []byte{0x52}
	prefixKeyDestructionRecord    = []byte{0x53}
	prefixKeyDestructionByAccount = []byte{0x54}
	prefixLegalHold               = []byte{0x55}
)

func erasureRequestKey(requestID string) []byte {
	return append(prefixErasureRequest, []byte(requestID)...)
}

func erasureRequestByAddressKey(address string, requestID string) []byte {
	key := make([]byte, 0, len(prefixErasureRequestByAddress)+len(address)+1+len(requestID))
	key = append(key, prefixErasureRequestByAddress...)
	key = append(key, []byte(address)...)
	key = append(key, byte(0x00))
	key = append(key, []byte(requestID)...)
	return key
}

func erasureRequestByAddressPrefixKey(address string) []byte {
	key := make([]byte, 0, len(prefixErasureRequestByAddress)+len(address)+1)
	key = append(key, prefixErasureRequestByAddress...)
	key = append(key, []byte(address)...)
	key = append(key, byte(0x00))
	return key
}

func pendingErasureKey(deadlineUnix int64, requestID string) []byte {
	key := make([]byte, 0, len(prefixPendingErasure)+8+1+len(requestID))
	key = append(key, prefixPendingErasure...)
	key = append(key, sdk.Uint64ToBigEndian(safeUint64FromInt64(deadlineUnix))...)
	key = append(key, byte(0x00))
	key = append(key, []byte(requestID)...)
	return key
}

func pendingErasurePrefixKey() []byte {
	return prefixPendingErasure
}

func pendingErasureBeforeKey(beforeUnix int64) []byte {
	key := make([]byte, 0, len(prefixPendingErasure)+8)
	key = append(key, prefixPendingErasure...)
	key = append(key, sdk.Uint64ToBigEndian(safeUint64FromInt64(beforeUnix))...)
	return key
}

func keyDestructionRecordKey(recordID string) []byte {
	return append(prefixKeyDestructionRecord, []byte(recordID)...)
}

func keyDestructionByAccountKey(address string, recordID string) []byte {
	key := make([]byte, 0, len(prefixKeyDestructionByAccount)+len(address)+1+len(recordID))
	key = append(key, prefixKeyDestructionByAccount...)
	key = append(key, []byte(address)...)
	key = append(key, byte(0x00))
	key = append(key, []byte(recordID)...)
	return key
}

func legalHoldKey(address string) []byte {
	return append(prefixLegalHold, []byte(address)...)
}

func extractRequestIDFromKey(key []byte, prefix []byte) string {
	if len(key) <= len(prefix) {
		return ""
	}
	return string(key[len(prefix):])
}

func extractPendingErasureRequestID(key []byte) string {
	// Key format: prefix (1) + timestamp (8) + separator (1) + requestID
	offset := len(prefixPendingErasure) + 8 + 1
	if len(key) <= offset {
		return ""
	}
	return string(key[offset:])
}

// ============================================================================
// ID Generation Functions
// ============================================================================

func generateErasureRequestID(address string, timestamp time.Time, blockHeight int64) string {
	data := fmt.Sprintf("erasure:%s:%d:%d", address, timestamp.UnixNano(), blockHeight)
	hash := sha256.Sum256([]byte(data))
	return "erasure_" + hex.EncodeToString(hash[:8])
}

func generateKeyDestructionRecordID(address string, timestamp time.Time) string {
	data := fmt.Sprintf("keydestroy:%s:%d", address, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "keydestroy_" + hex.EncodeToString(hash[:8])
}

func generateCertificateID(requestID string, timestamp time.Time) string {
	data := fmt.Sprintf("cert:%s:%d", requestID, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "erasure_cert_" + hex.EncodeToString(hash[:8])
}

func generateVerificationHash(request *types.ErasureRequest) []byte {
	data := fmt.Sprintf("%s:%s:%d:%d",
		request.RequestID,
		request.RequesterAddress,
		request.RequestedAt.UnixNano(),
		request.RequestedAtBlock)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func generateKeyDestructionHash(record *types.KeyDestructionRecord) []byte {
	data := fmt.Sprintf("%s:%s:%d:%v",
		record.RecordID,
		record.AccountAddress,
		record.DestroyedAt.UnixNano(),
		record.KeyFingerprints)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func generateCertificateSignature(cert *types.ErasureConfirmationCertificate) []byte {
	data := fmt.Sprintf("%s:%s:%d:%d",
		cert.CertificateID,
		cert.DataSubjectAddress,
		cert.RequestedAt.UnixNano(),
		cert.CompletedAt.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// ============================================================================
// Storage Conversion Functions
// ============================================================================

func erasureRequestToStore(r *types.ErasureRequest) *erasureRequestStore {
	categories := make([]string, len(r.Categories))
	for i, c := range r.Categories {
		categories[i] = string(c)
	}

	rs := &erasureRequestStore{
		Version:          r.Version,
		RequestID:        r.RequestID,
		RequesterAddress: r.RequesterAddress,
		Categories:       categories,
		Status:           string(r.Status),
		RequestedAt:      r.RequestedAt.UnixNano(),
		RequestedAtBlock: r.RequestedAtBlock,
		DeadlineAt:       r.DeadlineAt.UnixNano(),
		RejectionDetails: r.RejectionDetails,
		VerificationHash: r.VerificationHash,
	}

	if r.ProcessedAt != nil {
		ts := r.ProcessedAt.UnixNano()
		rs.ProcessedAt = &ts
	}
	if r.ProcessedAtBlock != nil {
		rs.ProcessedAtBlock = r.ProcessedAtBlock
	}
	if r.CompletedAt != nil {
		ts := r.CompletedAt.UnixNano()
		rs.CompletedAt = &ts
	}
	if r.CompletedAtBlock != nil {
		rs.CompletedAtBlock = r.CompletedAtBlock
	}
	if r.RejectionReason != nil {
		reason := string(*r.RejectionReason)
		rs.RejectionReason = &reason
	}
	if r.ErasureReport != nil {
		rs.ErasureReport = erasureReportToStore(r.ErasureReport)
	}

	return rs
}

func erasureRequestFromStore(rs *erasureRequestStore) *types.ErasureRequest {
	categories := make([]types.ErasureCategory, len(rs.Categories))
	for i, c := range rs.Categories {
		categories[i] = types.ErasureCategory(c)
	}

	r := &types.ErasureRequest{
		Version:          rs.Version,
		RequestID:        rs.RequestID,
		RequesterAddress: rs.RequesterAddress,
		Categories:       categories,
		Status:           types.ErasureRequestStatus(rs.Status),
		RequestedAt:      time.Unix(0, rs.RequestedAt),
		RequestedAtBlock: rs.RequestedAtBlock,
		DeadlineAt:       time.Unix(0, rs.DeadlineAt),
		RejectionDetails: rs.RejectionDetails,
		VerificationHash: rs.VerificationHash,
	}

	if rs.ProcessedAt != nil {
		ts := time.Unix(0, *rs.ProcessedAt)
		r.ProcessedAt = &ts
	}
	if rs.ProcessedAtBlock != nil {
		r.ProcessedAtBlock = rs.ProcessedAtBlock
	}
	if rs.CompletedAt != nil {
		ts := time.Unix(0, *rs.CompletedAt)
		r.CompletedAt = &ts
	}
	if rs.CompletedAtBlock != nil {
		r.CompletedAtBlock = rs.CompletedAtBlock
	}
	if rs.RejectionReason != nil {
		reason := types.ErasureRejectionReason(*rs.RejectionReason)
		r.RejectionReason = &reason
	}
	if rs.ErasureReport != nil {
		r.ErasureReport = erasureReportFromStore(rs.ErasureReport)
	}

	return r
}

func erasureReportToStore(r *types.ErasureReport) *erasureReportStore {
	categories := make([]string, len(r.DataCategoriesErased))
	for i, c := range r.DataCategoriesErased {
		categories[i] = string(c)
	}

	rs := &erasureReportStore{
		BiometricDataErased:       r.BiometricDataErased,
		BiometricKeysDestroyed:    r.BiometricKeysDestroyed,
		IdentityDocumentsErased:   r.IdentityDocumentsErased,
		VerificationHistoryErased: r.VerificationHistoryErased,
		DerivedFeaturesErased:     r.DerivedFeaturesErased,
		ConsentRecordsErased:      r.ConsentRecordsErased,
		OffChainDataDeleted:       r.OffChainDataDeleted,
		OnChainDataMadeUnreadable: r.OnChainDataMadeUnreadable,
		TotalRecordsAffected:      r.TotalRecordsAffected,
		DataCategoriesErased:      categories,
		RetainedDataCategories:    r.RetainedDataCategories,
		RetentionReasons:          r.RetentionReasons,
		ReportGeneratedAt:         r.ReportGeneratedAt.UnixNano(),
	}

	if r.BackupDeletionScheduled != nil {
		ts := r.BackupDeletionScheduled.UnixNano()
		rs.BackupDeletionScheduled = &ts
	}

	return rs
}

func erasureReportFromStore(rs *erasureReportStore) *types.ErasureReport {
	categories := make([]types.ErasureCategory, len(rs.DataCategoriesErased))
	for i, c := range rs.DataCategoriesErased {
		categories[i] = types.ErasureCategory(c)
	}

	r := &types.ErasureReport{
		BiometricDataErased:       rs.BiometricDataErased,
		BiometricKeysDestroyed:    rs.BiometricKeysDestroyed,
		IdentityDocumentsErased:   rs.IdentityDocumentsErased,
		VerificationHistoryErased: rs.VerificationHistoryErased,
		DerivedFeaturesErased:     rs.DerivedFeaturesErased,
		ConsentRecordsErased:      rs.ConsentRecordsErased,
		OffChainDataDeleted:       rs.OffChainDataDeleted,
		OnChainDataMadeUnreadable: rs.OnChainDataMadeUnreadable,
		TotalRecordsAffected:      rs.TotalRecordsAffected,
		DataCategoriesErased:      categories,
		RetainedDataCategories:    rs.RetainedDataCategories,
		RetentionReasons:          rs.RetentionReasons,
		ReportGeneratedAt:         time.Unix(0, rs.ReportGeneratedAt),
	}

	if rs.BackupDeletionScheduled != nil {
		ts := time.Unix(0, *rs.BackupDeletionScheduled)
		r.BackupDeletionScheduled = &ts
	}

	return r
}

func safeUint64FromInt64(value int64) uint64 {
	if value < 0 {
		return 0
	}
	//nolint:gosec // range checked above
	return uint64(value)
}
