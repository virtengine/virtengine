// Package keeper implements the Fraud module keeper.
//
// VE-912: Fraud reporting flow - Keeper implementation
// This keeper manages fraud reports with encrypted evidence,
// moderator queue routing, and comprehensive audit trail logging.
package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
)

// IKeeper defines the interface for the Fraud keeper
type IKeeper interface {
	// Fraud Reports
	SubmitFraudReport(ctx sdk.Context, report *types.FraudReport) error
	GetFraudReport(ctx sdk.Context, reportID string) (types.FraudReport, bool)
	SetFraudReport(ctx sdk.Context, report types.FraudReport) error
	GetFraudReportsByReporter(ctx sdk.Context, reporterAddr string) []types.FraudReport
	GetFraudReportsByReportedParty(ctx sdk.Context, reportedAddr string) []types.FraudReport
	GetFraudReportsByStatus(ctx sdk.Context, status types.FraudReportStatus) []types.FraudReport

	// Moderator Queue
	AddToModeratorQueue(ctx sdk.Context, entry types.ModeratorQueueEntry) error
	RemoveFromModeratorQueue(ctx sdk.Context, reportID string) error
	GetModeratorQueue(ctx sdk.Context) []types.ModeratorQueueEntry
	GetModeratorQueueEntry(ctx sdk.Context, reportID string) (types.ModeratorQueueEntry, bool)
	AssignModerator(ctx sdk.Context, reportID, moderatorAddr string) error

	// Status Management
	UpdateReportStatus(ctx sdk.Context, reportID string, newStatus types.FraudReportStatus, actorAddr string, notes string) error
	ResolveFraudReport(ctx sdk.Context, reportID string, resolution types.ResolutionType, notes string, moderatorAddr string) error
	RejectFraudReport(ctx sdk.Context, reportID, notes, moderatorAddr string) error
	EscalateFraudReport(ctx sdk.Context, reportID, reason, moderatorAddr string) error

	// Audit Logging
	CreateAuditLog(ctx sdk.Context, log *types.FraudAuditLog) error
	GetAuditLog(ctx sdk.Context, logID string) (types.FraudAuditLog, bool)
	GetAuditLogsForReport(ctx sdk.Context, reportID string) []types.FraudAuditLog
	GetAllAuditLogs(ctx sdk.Context) []types.FraudAuditLog

	// Authorization
	IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool
	IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool

	// Sequences
	GetNextFraudReportSequence(ctx sdk.Context) uint64
	SetNextFraudReportSequence(ctx sdk.Context, seq uint64)
	GetNextAuditLogSequence(ctx sdk.Context) uint64
	SetNextAuditLogSequence(ctx sdk.Context, seq uint64)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithFraudReports(ctx sdk.Context, fn func(types.FraudReport) bool)
	WithAuditLogs(ctx sdk.Context, fn func(types.FraudAuditLog) bool)
	WithModeratorQueue(ctx sdk.Context, fn func(types.ModeratorQueueEntry) bool)

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	Logger(ctx sdk.Context) log.Logger
}

// RolesKeeper defines the expected roles keeper interface
type RolesKeeper interface {
	HasRole(ctx sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool
	IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool
}

// ProviderKeeper defines the expected provider keeper interface
type ProviderKeeper interface {
	IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool
}

// Keeper implements the Fraud module keeper
type Keeper struct {
	skey           storetypes.StoreKey
	cdc            codec.BinaryCodec
	rolesKeeper    RolesKeeper
	providerKeeper ProviderKeeper
	authority      string
}

// NewKeeper creates and returns an instance for Fraud keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	rolesKeeper RolesKeeper,
	providerKeeper ProviderKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		skey:           skey,
		rolesKeeper:    rolesKeeper,
		providerKeeper: providerKeeper,
		authority:      authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================================================================
// Parameters
// ============================================================================

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(ParamsKey, bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// ============================================================================
// Sequences
// ============================================================================

// GetNextFraudReportSequence returns the next fraud report sequence number
func (k Keeper) GetNextFraudReportSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(SequenceKeyFraudReport)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextFraudReportSequence sets the next fraud report sequence number
func (k Keeper) SetNextFraudReportSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(SequenceKeyFraudReport, bz)
}

// GetNextAuditLogSequence returns the next audit log sequence number
func (k Keeper) GetNextAuditLogSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(SequenceKeyAuditLog)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextAuditLogSequence sets the next audit log sequence number
func (k Keeper) SetNextAuditLogSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(SequenceKeyAuditLog, bz)
}

// ============================================================================
// Authorization
// ============================================================================

// IsProvider checks if an address is a registered provider
func (k Keeper) IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.providerKeeper != nil {
		return k.providerKeeper.IsProvider(ctx, addr)
	}
	// Fallback: check roles
	return k.rolesKeeper != nil && k.rolesKeeper.HasRole(ctx, addr, rolestypes.RoleServiceProvider)
}

// IsModerator checks if an address has moderator role
func (k Keeper) IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.rolesKeeper != nil {
		return k.rolesKeeper.IsModerator(ctx, addr)
	}
	return false
}

// IsAdmin checks if an address has admin role
func (k Keeper) IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.rolesKeeper != nil {
		return k.rolesKeeper.IsAdmin(ctx, addr)
	}
	return false
}

// ============================================================================
// Fraud Reports
// ============================================================================

// SubmitFraudReport submits a new fraud report
func (k Keeper) SubmitFraudReport(ctx sdk.Context, report *types.FraudReport) error {
	if err := report.Validate(); err != nil {
		return err
	}

	// Check reporter is a provider
	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return types.ErrInvalidReporter.Wrap(err.Error())
	}
	if !k.IsProvider(ctx, reporterAddr) {
		return types.ErrUnauthorizedReporter
	}

	// Assign sequence ID if not set
	if report.ID == "" {
		seq := k.GetNextFraudReportSequence(ctx)
		report.ID = fmt.Sprintf("fraud-report-%d", seq)
		k.SetNextFraudReportSequence(ctx, seq+1)
	}

	// Store the report
	if err := k.SetFraudReport(ctx, *report); err != nil {
		return err
	}

	// Add to moderator queue
	priority := k.calculatePriority(report.Category)
	queueEntry := types.NewModeratorQueueEntry(
		report.ID,
		report.Category,
		priority,
		report.SubmittedAt,
	)
	if err := k.AddToModeratorQueue(ctx, *queueEntry); err != nil {
		return err
	}

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		report.ID,
		types.AuditActionSubmitted,
		report.Reporter,
		types.FraudReportStatusUnspecified,
		types.FraudReportStatusSubmitted,
		"Fraud report submitted with encrypted evidence",
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportSubmitted,
			sdk.NewAttribute(types.AttributeKeyReportID, report.ID),
			sdk.NewAttribute(types.AttributeKeyReporter, report.Reporter),
			sdk.NewAttribute(types.AttributeKeyReportedParty, report.ReportedParty),
			sdk.NewAttribute(types.AttributeKeyCategory, report.Category.String()),
			sdk.NewAttribute(types.AttributeKeyStatus, report.Status.String()),
			sdk.NewAttribute(types.AttributeKeyHasEvidence, strconv.FormatBool(len(report.Evidence) > 0)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, strconv.FormatInt(report.BlockHeight, 10)),
		),
	)

	k.Logger(ctx).Info("fraud report submitted",
		"report_id", report.ID,
		"reporter", report.Reporter,
		"reported_party", report.ReportedParty,
		"category", report.Category.String(),
	)

	return nil
}

// calculatePriority calculates the queue priority based on category
func (k Keeper) calculatePriority(category types.FraudCategory) uint8 {
	switch category {
	case types.FraudCategoryFakeIdentity, types.FraudCategorySybilAttack:
		return 10 // Highest priority
	case types.FraudCategoryPaymentFraud, types.FraudCategoryMaliciousContent:
		return 8
	case types.FraudCategoryServiceMisrepresentation, types.FraudCategoryResourceAbuse:
		return 6
	case types.FraudCategoryTermsViolation:
		return 4
	default:
		return 2
	}
}

// GetFraudReport returns a fraud report by ID
func (k Keeper) GetFraudReport(ctx sdk.Context, reportID string) (types.FraudReport, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(FraudReportKey(reportID))
	if bz == nil {
		return types.FraudReport{}, false
	}

	var report types.FraudReport
	if err := json.Unmarshal(bz, &report); err != nil {
		return types.FraudReport{}, false
	}
	return report, true
}

// SetFraudReport stores a fraud report
func (k Keeper) SetFraudReport(ctx sdk.Context, report types.FraudReport) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(report)
	if err != nil {
		return err
	}

	store.Set(FraudReportKey(report.ID), bz)

	// Update indexes
	k.updateReportIndexes(ctx, report)

	return nil
}

// updateReportIndexes updates all indexes for a report
func (k Keeper) updateReportIndexes(ctx sdk.Context, report types.FraudReport) {
	store := ctx.KVStore(k.skey)

	// Reporter index
	reporterIndexKey := append(ReporterIndexKey(report.Reporter), []byte("/"+report.ID)...)
	store.Set(reporterIndexKey, []byte(report.ID))

	// Reported party index
	reportedIndexKey := append(ReportedPartyIndexKey(report.ReportedParty), []byte("/"+report.ID)...)
	store.Set(reportedIndexKey, []byte(report.ID))

	// Status index
	statusIndexKey := append(StatusIndexKey(report.Status), []byte("/"+report.ID)...)
	store.Set(statusIndexKey, []byte(report.ID))
}

// GetFraudReportsByReporter returns all reports by a specific reporter
func (k Keeper) GetFraudReportsByReporter(ctx sdk.Context, reporterAddr string) []types.FraudReport {
	store := ctx.KVStore(k.skey)
	var reports []types.FraudReport

	indexPrefix := ReporterIndexKey(reporterAddr)
	iter := storetypes.KVStorePrefixIterator(store, indexPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		reportID := string(iter.Value())
		if report, found := k.GetFraudReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetFraudReportsByReportedParty returns all reports against a specific party
func (k Keeper) GetFraudReportsByReportedParty(ctx sdk.Context, reportedAddr string) []types.FraudReport {
	store := ctx.KVStore(k.skey)
	var reports []types.FraudReport

	indexPrefix := ReportedPartyIndexKey(reportedAddr)
	iter := storetypes.KVStorePrefixIterator(store, indexPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		reportID := string(iter.Value())
		if report, found := k.GetFraudReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetFraudReportsByStatus returns all reports with a specific status
func (k Keeper) GetFraudReportsByStatus(ctx sdk.Context, status types.FraudReportStatus) []types.FraudReport {
	store := ctx.KVStore(k.skey)
	var reports []types.FraudReport

	indexPrefix := StatusIndexKey(status)
	iter := storetypes.KVStorePrefixIterator(store, indexPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		reportID := string(iter.Value())
		if report, found := k.GetFraudReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// WithFraudReports iterates over all fraud reports
func (k Keeper) WithFraudReports(ctx sdk.Context, fn func(types.FraudReport) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, FraudReportPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report types.FraudReport
		if err := json.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}
		if fn(report) {
			break
		}
	}
}

// ============================================================================
// Moderator Queue
// ============================================================================

// AddToModeratorQueue adds a report to the moderator queue
func (k Keeper) AddToModeratorQueue(ctx sdk.Context, entry types.ModeratorQueueEntry) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	store.Set(ModeratorQueueKey(entry.ReportID), bz)

	// Emit queue updated event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorQueueUpdated,
			sdk.NewAttribute(types.AttributeKeyReportID, entry.ReportID),
			sdk.NewAttribute(types.AttributeKeyCategory, entry.Category.String()),
			sdk.NewAttribute(types.AttributeKeyAction, "added"),
		),
	)

	return nil
}

// RemoveFromModeratorQueue removes a report from the moderator queue
func (k Keeper) RemoveFromModeratorQueue(ctx sdk.Context, reportID string) error {
	store := ctx.KVStore(k.skey)

	if !store.Has(ModeratorQueueKey(reportID)) {
		return types.ErrReportNotInQueue
	}

	store.Delete(ModeratorQueueKey(reportID))

	// Emit queue updated event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorQueueUpdated,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyAction, "removed"),
		),
	)

	return nil
}

// GetModeratorQueueEntry returns a queue entry by report ID
func (k Keeper) GetModeratorQueueEntry(ctx sdk.Context, reportID string) (types.ModeratorQueueEntry, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(ModeratorQueueKey(reportID))
	if bz == nil {
		return types.ModeratorQueueEntry{}, false
	}

	var entry types.ModeratorQueueEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return types.ModeratorQueueEntry{}, false
	}
	return entry, true
}

// GetModeratorQueue returns all pending queue entries
func (k Keeper) GetModeratorQueue(ctx sdk.Context) []types.ModeratorQueueEntry {
	var entries []types.ModeratorQueueEntry
	k.WithModeratorQueue(ctx, func(entry types.ModeratorQueueEntry) bool {
		entries = append(entries, entry)
		return false
	})
	return entries
}

// WithModeratorQueue iterates over all moderator queue entries
func (k Keeper) WithModeratorQueue(ctx sdk.Context, fn func(types.ModeratorQueueEntry) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, ModeratorQueuePrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var entry types.ModeratorQueueEntry
		if err := json.Unmarshal(iter.Value(), &entry); err != nil {
			continue
		}
		if fn(entry) {
			break
		}
	}
}

// AssignModerator assigns a moderator to a report
func (k Keeper) AssignModerator(ctx sdk.Context, reportID, moderatorAddr string) error {
	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Verify moderator role
	modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
	if err != nil {
		return types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, modAddr) {
		return types.ErrUnauthorizedModerator
	}

	// Update report
	previousStatus := report.Status
	report.AssignedModerator = moderatorAddr
	if report.Status == types.FraudReportStatusSubmitted {
		report.Status = types.FraudReportStatusReviewing
	}
	report.UpdatedAt = ctx.BlockTime()

	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Update queue entry
	if entry, found := k.GetModeratorQueueEntry(ctx, reportID); found {
		entry.AssignedTo = moderatorAddr
		if err := k.AddToModeratorQueue(ctx, entry); err != nil {
			return err
		}
	}

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		reportID,
		types.AuditActionAssigned,
		moderatorAddr,
		previousStatus,
		report.Status,
		fmt.Sprintf("Report assigned to moderator %s", moderatorAddr),
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportAssigned,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
		),
	)

	k.Logger(ctx).Info("fraud report assigned",
		"report_id", reportID,
		"moderator", moderatorAddr,
	)

	return nil
}

// ============================================================================
// Status Management
// ============================================================================

// UpdateReportStatus updates the status of a fraud report
func (k Keeper) UpdateReportStatus(ctx sdk.Context, reportID string, newStatus types.FraudReportStatus, actorAddr string, notes string) error {
	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Verify actor is moderator
	actor, err := sdk.AccAddressFromBech32(actorAddr)
	if err != nil {
		return types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, actor) {
		return types.ErrUnauthorizedModerator
	}

	previousStatus := report.Status
	if err := report.UpdateStatus(newStatus, ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		reportID,
		types.AuditActionStatusChanged,
		actorAddr,
		previousStatus,
		newStatus,
		notes,
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportStatusChanged,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyPreviousStatus, previousStatus.String()),
			sdk.NewAttribute(types.AttributeKeyStatus, newStatus.String()),
			sdk.NewAttribute(types.AttributeKeyModerator, actorAddr),
		),
	)

	return nil
}

// ResolveFraudReport resolves a fraud report
func (k Keeper) ResolveFraudReport(ctx sdk.Context, reportID string, resolution types.ResolutionType, notes string, moderatorAddr string) error {
	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Verify moderator role
	modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
	if err != nil {
		return types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, modAddr) {
		return types.ErrUnauthorizedModerator
	}

	previousStatus := report.Status
	if err := report.Resolve(resolution, notes, ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Remove from queue
	_ = k.RemoveFromModeratorQueue(ctx, reportID)

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		reportID,
		types.AuditActionResolved,
		moderatorAddr,
		previousStatus,
		types.FraudReportStatusResolved,
		fmt.Sprintf("Resolution: %s. Notes: %s", resolution.String(), notes),
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportResolved,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyResolution, resolution.String()),
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
		),
	)

	k.Logger(ctx).Info("fraud report resolved",
		"report_id", reportID,
		"resolution", resolution.String(),
		"moderator", moderatorAddr,
	)

	return nil
}

// RejectFraudReport rejects a fraud report
func (k Keeper) RejectFraudReport(ctx sdk.Context, reportID, notes, moderatorAddr string) error {
	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Verify moderator role
	modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
	if err != nil {
		return types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, modAddr) {
		return types.ErrUnauthorizedModerator
	}

	previousStatus := report.Status
	if err := report.Reject(notes, ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Remove from queue
	_ = k.RemoveFromModeratorQueue(ctx, reportID)

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		reportID,
		types.AuditActionRejected,
		moderatorAddr,
		previousStatus,
		types.FraudReportStatusRejected,
		notes,
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportRejected,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
		),
	)

	k.Logger(ctx).Info("fraud report rejected",
		"report_id", reportID,
		"moderator", moderatorAddr,
	)

	return nil
}

// EscalateFraudReport escalates a fraud report to admin
func (k Keeper) EscalateFraudReport(ctx sdk.Context, reportID, reason, moderatorAddr string) error {
	report, found := k.GetFraudReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Verify moderator role
	modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
	if err != nil {
		return types.ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if !k.IsModerator(ctx, modAddr) {
		return types.ErrUnauthorizedModerator
	}

	previousStatus := report.Status
	if err := report.UpdateStatus(types.FraudReportStatusEscalated, ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetFraudReport(ctx, report); err != nil {
		return err
	}

	// Update queue priority
	if entry, found := k.GetModeratorQueueEntry(ctx, reportID); found {
		entry.Priority = 15 // Escalated = highest priority
		entry.AssignedTo = "" // Unassign for admin to pick up
		if err := k.AddToModeratorQueue(ctx, entry); err != nil {
			return err
		}
	}

	// Create audit log
	auditLog := k.createAuditLogEntry(
		ctx,
		reportID,
		types.AuditActionEscalated,
		moderatorAddr,
		previousStatus,
		types.FraudReportStatusEscalated,
		reason,
	)
	if err := k.CreateAuditLog(ctx, auditLog); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraudReportEscalated,
			sdk.NewAttribute(types.AttributeKeyReportID, reportID),
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
		),
	)

	k.Logger(ctx).Info("fraud report escalated",
		"report_id", reportID,
		"reason", reason,
		"moderator", moderatorAddr,
	)

	return nil
}

// ============================================================================
// Audit Logging
// ============================================================================

// createAuditLogEntry creates a new audit log entry with auto-generated ID
func (k Keeper) createAuditLogEntry(
	ctx sdk.Context,
	reportID string,
	action types.AuditAction,
	actor string,
	previousStatus types.FraudReportStatus,
	newStatus types.FraudReportStatus,
	details string,
) *types.FraudAuditLog {
	seq := k.GetNextAuditLogSequence(ctx)
	logID := fmt.Sprintf("%s/audit-%d", reportID, seq)
	k.SetNextAuditLogSequence(ctx, seq+1)

	return types.NewFraudAuditLog(
		logID,
		reportID,
		action,
		actor,
		previousStatus,
		newStatus,
		details,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
}

// CreateAuditLog stores an audit log entry
func (k Keeper) CreateAuditLog(ctx sdk.Context, log *types.FraudAuditLog) error {
	if err := log.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(log)
	if err != nil {
		return err
	}

	store.Set(AuditLogKey(log.ID), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuditLogCreated,
			sdk.NewAttribute(types.AttributeKeyAuditLogID, log.ID),
			sdk.NewAttribute(types.AttributeKeyReportID, log.ReportID),
			sdk.NewAttribute(types.AttributeKeyAction, log.Action.String()),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, strconv.FormatInt(log.BlockHeight, 10)),
		),
	)

	return nil
}

// GetAuditLog returns an audit log entry by ID
func (k Keeper) GetAuditLog(ctx sdk.Context, logID string) (types.FraudAuditLog, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(AuditLogKey(logID))
	if bz == nil {
		return types.FraudAuditLog{}, false
	}

	var log types.FraudAuditLog
	if err := json.Unmarshal(bz, &log); err != nil {
		return types.FraudAuditLog{}, false
	}
	return log, true
}

// GetAuditLogsForReport returns all audit logs for a specific report
func (k Keeper) GetAuditLogsForReport(ctx sdk.Context, reportID string) []types.FraudAuditLog {
	store := ctx.KVStore(k.skey)
	var logs []types.FraudAuditLog

	// Use prefix iterator to find all logs for this report
	prefixStore := prefix.NewStore(store, AuditLogPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var log types.FraudAuditLog
		if err := json.Unmarshal(iter.Value(), &log); err != nil {
			continue
		}
		if log.ReportID == reportID {
			logs = append(logs, log)
		}
	}

	return logs
}

// GetAllAuditLogs returns all audit logs
func (k Keeper) GetAllAuditLogs(ctx sdk.Context) []types.FraudAuditLog {
	var logs []types.FraudAuditLog
	k.WithAuditLogs(ctx, func(log types.FraudAuditLog) bool {
		logs = append(logs, log)
		return false
	})
	return logs
}

// WithAuditLogs iterates over all audit logs
func (k Keeper) WithAuditLogs(ctx sdk.Context, fn func(types.FraudAuditLog) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, AuditLogPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var log types.FraudAuditLog
		if err := json.Unmarshal(iter.Value(), &log); err != nil {
			continue
		}
		if fn(log) {
			break
		}
	}
}
