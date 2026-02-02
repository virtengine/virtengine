// Package keeper provides the escrow module keeper with dispute workflow management capabilities.
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// DisputeKeeper defines the interface for dispute workflow management
type DisputeKeeper interface {
	// CreateDisputeWorkflow creates a new dispute workflow
	CreateDisputeWorkflow(ctx sdk.Context, workflow *billing.DisputeWorkflow) error

	// GetDisputeWorkflow retrieves a dispute workflow by ID
	GetDisputeWorkflow(ctx sdk.Context, disputeID string) (*billing.DisputeWorkflow, error)

	// UpdateDisputeWorkflow updates an existing dispute workflow
	UpdateDisputeWorkflow(ctx sdk.Context, workflow *billing.DisputeWorkflow) error

	// GetDisputesByInvoice retrieves all disputes for an invoice
	GetDisputesByInvoice(ctx sdk.Context, invoiceID string) ([]*billing.DisputeWorkflow, error)

	// GetDisputesByStatus retrieves disputes by status with pagination
	GetDisputesByStatus(ctx sdk.Context, status billing.DisputeStatus, pagination *query.PageRequest) ([]*billing.DisputeWorkflow, *query.PageResponse, error)

	// InitiateDispute initiates a new dispute for an invoice
	InitiateDispute(
		ctx sdk.Context,
		invoiceID string,
		category billing.DisputeCategory,
		subject string,
		description string,
		disputedAmount sdk.Coins,
		initiator string,
	) (*billing.DisputeWorkflow, error)

	// UploadEvidence uploads evidence for a dispute
	UploadEvidence(ctx sdk.Context, disputeID string, evidence *billing.DisputeEvidence) error

	// GetEvidence retrieves evidence by ID
	GetEvidence(ctx sdk.Context, evidenceID string) (*billing.DisputeEvidence, error)

	// GetEvidenceByDispute retrieves all evidence for a dispute
	GetEvidenceByDispute(ctx sdk.Context, disputeID string) ([]*billing.DisputeEvidence, error)

	// SubmitForReview submits a dispute for review
	SubmitForReview(ctx sdk.Context, disputeID string, submitter string) error

	// ResolveDispute resolves a dispute
	ResolveDispute(
		ctx sdk.Context,
		disputeID string,
		resolution billing.DisputeResolutionType,
		details string,
		resolver string,
		refundAmount sdk.Coins,
	) error

	// EscalateDispute escalates a dispute
	EscalateDispute(ctx sdk.Context, disputeID string, escalateTo string, reason string) error

	// CheckDisputeWindowOpen checks if the dispute window is open for an invoice
	CheckDisputeWindowOpen(ctx sdk.Context, invoiceID string) (bool, time.Duration, error)

	// CreateCorrection creates a new billing correction
	CreateCorrection(ctx sdk.Context, correction *billing.Correction) error

	// GetCorrection retrieves a correction by ID
	GetCorrection(ctx sdk.Context, correctionID string) (*billing.Correction, error)

	// ApplyCorrection applies a correction
	ApplyCorrection(ctx sdk.Context, correctionID string, appliedBy string) error

	// GetCorrectionsByInvoice retrieves corrections for an invoice
	GetCorrectionsByInvoice(ctx sdk.Context, invoiceID string) ([]*billing.Correction, error)

	// GetPendingCorrections retrieves pending corrections with pagination
	GetPendingCorrections(ctx sdk.Context, pagination *query.PageRequest) ([]*billing.Correction, *query.PageResponse, error)

	// GetDisputeRules retrieves the dispute rules configuration
	GetDisputeRules(ctx sdk.Context) (*billing.DisputeRules, error)

	// SaveDisputeRules saves the dispute rules configuration
	SaveDisputeRules(ctx sdk.Context, rules *billing.DisputeRules) error

	// WithDisputes iterates over all disputes
	WithDisputes(ctx sdk.Context, fn func(*billing.DisputeWorkflow) bool)
}

// disputeKeeper implements DisputeKeeper
type disputeKeeper struct {
	k *keeper
}

// NewDisputeKeeper creates a new dispute keeper from the base keeper
func (k *keeper) NewDisputeKeeper() DisputeKeeper {
	return &disputeKeeper{k: k}
}

// CreateDisputeWorkflow creates a new dispute workflow
func (dk *disputeKeeper) CreateDisputeWorkflow(ctx sdk.Context, workflow *billing.DisputeWorkflow) error {
	store := ctx.KVStore(dk.k.skey)

	// Check if dispute already exists
	key := billing.BuildDisputeWorkflowKey(workflow.DisputeID)
	if store.Has(key) {
		return fmt.Errorf("dispute already exists: %s", workflow.DisputeID)
	}

	// Validate workflow
	if err := workflow.Validate(); err != nil {
		return fmt.Errorf("invalid dispute workflow: %w", err)
	}

	// Marshal and store
	bz, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal dispute workflow: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	dk.setDisputeIndexes(store, workflow)

	return nil
}

// GetDisputeWorkflow retrieves a dispute workflow by ID
func (dk *disputeKeeper) GetDisputeWorkflow(ctx sdk.Context, disputeID string) (*billing.DisputeWorkflow, error) {
	store := ctx.KVStore(dk.k.skey)
	key := billing.BuildDisputeWorkflowKey(disputeID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("dispute not found: %s", disputeID)
	}

	var workflow billing.DisputeWorkflow
	if err := json.Unmarshal(bz, &workflow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dispute workflow: %w", err)
	}

	return &workflow, nil
}

// UpdateDisputeWorkflow updates an existing dispute workflow
func (dk *disputeKeeper) UpdateDisputeWorkflow(ctx sdk.Context, workflow *billing.DisputeWorkflow) error {
	store := ctx.KVStore(dk.k.skey)

	// Check if dispute exists
	key := billing.BuildDisputeWorkflowKey(workflow.DisputeID)
	if !store.Has(key) {
		return fmt.Errorf("dispute not found: %s", workflow.DisputeID)
	}

	// Get old workflow to update status index if needed
	oldWorkflow, err := dk.GetDisputeWorkflow(ctx, workflow.DisputeID)
	if err != nil {
		return err
	}

	// Remove old status index if status changed
	if oldWorkflow.Status != workflow.Status {
		oldStatusKey := billing.BuildDisputeWorkflowByStatusKey(oldWorkflow.Status, workflow.DisputeID)
		store.Delete(oldStatusKey)

		// Add new status index
		newStatusKey := billing.BuildDisputeWorkflowByStatusKey(workflow.Status, workflow.DisputeID)
		store.Set(newStatusKey, []byte(workflow.DisputeID))
	}

	// Update timestamp
	workflow.UpdatedAt = ctx.BlockTime()

	// Validate workflow
	if err := workflow.Validate(); err != nil {
		return fmt.Errorf("invalid dispute workflow: %w", err)
	}

	// Marshal and store
	bz, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal dispute workflow: %w", err)
	}
	store.Set(key, bz)

	return nil
}

// GetDisputesByInvoice retrieves all disputes for an invoice
func (dk *disputeKeeper) GetDisputesByInvoice(ctx sdk.Context, invoiceID string) ([]*billing.DisputeWorkflow, error) {
	store := ctx.KVStore(dk.k.skey)
	prefix := billing.BuildDisputeWorkflowByInvoicePrefix(invoiceID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var disputes []*billing.DisputeWorkflow
	for ; iter.Valid(); iter.Next() {
		disputeID := string(iter.Value())
		workflow, err := dk.GetDisputeWorkflow(ctx, disputeID)
		if err != nil {
			continue
		}
		disputes = append(disputes, workflow)
	}

	return disputes, nil
}

// GetDisputesByStatus retrieves disputes by status with pagination
func (dk *disputeKeeper) GetDisputesByStatus(
	ctx sdk.Context,
	status billing.DisputeStatus,
	pagination *query.PageRequest,
) ([]*billing.DisputeWorkflow, *query.PageResponse, error) {
	store := ctx.KVStore(dk.k.skey)
	prefix := billing.BuildDisputeWorkflowByStatusPrefix(status)

	return dk.paginateDisputeIndex(store, prefix, pagination)
}

// InitiateDispute initiates a new dispute for an invoice
func (dk *disputeKeeper) InitiateDispute(
	ctx sdk.Context,
	invoiceID string,
	category billing.DisputeCategory,
	subject string,
	description string,
	disputedAmount sdk.Coins,
	initiator string,
) (*billing.DisputeWorkflow, error) {
	// Get invoice keeper to validate invoice exists and can be disputed
	invoiceKeeper := dk.k.NewInvoiceKeeper()

	// Validate the invoice exists
	invoice, err := invoiceKeeper.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be disputed
	if invoice.Status.IsTerminal() {
		return nil, fmt.Errorf("cannot dispute invoice with terminal status: %s", invoice.Status)
	}

	if invoice.Status == billing.InvoiceStatusDisputed {
		return nil, fmt.Errorf("invoice is already disputed: %s", invoiceID)
	}

	// Check dispute window is open
	isOpen, remaining, err := dk.CheckDisputeWindowOpen(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check dispute window: %w", err)
	}

	if !isOpen {
		return nil, fmt.Errorf("dispute window is closed for invoice: %s", invoiceID)
	}

	// Get dispute rules and validate
	rules, err := dk.GetDisputeRules(ctx)
	if err != nil {
		// Use default rules if none set
		defaultRules := billing.DefaultDisputeRules()
		rules = &defaultRules
	}

	// Check minimum dispute amount
	if disputedAmount.IsAllLT(rules.MinDisputeAmount) {
		return nil, fmt.Errorf("disputed amount %s is below minimum %s", disputedAmount.String(), rules.MinDisputeAmount.String())
	}

	// Check disputed amount doesn't exceed invoice total
	if disputedAmount.IsAllGT(invoice.Total) {
		return nil, fmt.Errorf("disputed amount %s exceeds invoice total %s", disputedAmount.String(), invoice.Total.String())
	}

	// Generate dispute ID
	seq := dk.GetDisputeSequence(ctx)
	disputeID := billing.NextDisputeWorkflowID(seq, "VE")

	// Create dispute window configuration
	disputeWindow := billing.NewDisputeWindow(
		fmt.Sprintf("%s-window", disputeID),
		invoiceID,
		invoice.IssuedAt,
		int64(remaining.Seconds()),
	)

	// Create correction limit from default
	correctionLimit := billing.DefaultCorrectionLimit()

	// Create the dispute workflow
	workflow := billing.NewDisputeWorkflow(
		disputeID,
		invoiceID,
		initiator,
		category,
		subject,
		description,
		disputedAmount,
		disputeWindow,
		&correctionLimit,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	// Save the dispute workflow
	if err := dk.CreateDisputeWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to create dispute workflow: %w", err)
	}

	// Update invoice status to Disputed
	_, err = invoiceKeeper.UpdateInvoiceStatus(ctx, invoiceID, billing.InvoiceStatusDisputed, initiator)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Increment sequence
	dk.SetDisputeSequence(ctx, seq+1)

	return workflow, nil
}

// UploadEvidence uploads evidence for a dispute
func (dk *disputeKeeper) UploadEvidence(ctx sdk.Context, disputeID string, evidence *billing.DisputeEvidence) error {
	store := ctx.KVStore(dk.k.skey)

	// Get the dispute workflow
	workflow, err := dk.GetDisputeWorkflow(ctx, disputeID)
	if err != nil {
		return err
	}

	// Validate evidence
	if err := evidence.Validate(); err != nil {
		return fmt.Errorf("invalid evidence: %w", err)
	}

	// Check dispute rules for evidence limits
	rules, err := dk.GetDisputeRules(ctx)
	if err != nil {
		defaultRules := billing.DefaultDisputeRules()
		rules = &defaultRules
	}

	// Check evidence count limit
	if uint32(len(workflow.Evidence)) >= rules.MaxEvidenceCount {
		return fmt.Errorf("evidence limit reached: max %d evidence items allowed", rules.MaxEvidenceCount)
	}

	// Check evidence size limit
	if evidence.FileSize > rules.MaxEvidenceSize {
		return fmt.Errorf("evidence file size %d exceeds maximum %d", evidence.FileSize, rules.MaxEvidenceSize)
	}

	// Check evidence type is allowed
	if !rules.IsEvidenceTypeAllowed(evidence.Type) {
		return fmt.Errorf("evidence type %s is not allowed", evidence.Type.String())
	}

	// Add evidence to workflow
	if err := workflow.AddEvidence(*evidence, ctx.BlockTime()); err != nil {
		return err
	}

	// Save updated workflow
	if err := dk.UpdateDisputeWorkflow(ctx, workflow); err != nil {
		return err
	}

	// Store evidence separately for direct lookup
	evidenceKey := billing.BuildDisputeEvidenceKey(evidence.EvidenceID)
	bz, err := json.Marshal(evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}
	store.Set(evidenceKey, bz)

	// Create evidence by dispute index
	indexKey := billing.BuildDisputeEvidenceByDisputeKey(disputeID, evidence.EvidenceID)
	store.Set(indexKey, []byte(evidence.EvidenceID))

	return nil
}

// GetEvidence retrieves evidence by ID
func (dk *disputeKeeper) GetEvidence(ctx sdk.Context, evidenceID string) (*billing.DisputeEvidence, error) {
	store := ctx.KVStore(dk.k.skey)
	key := billing.BuildDisputeEvidenceKey(evidenceID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("evidence not found: %s", evidenceID)
	}

	var evidence billing.DisputeEvidence
	if err := json.Unmarshal(bz, &evidence); err != nil {
		return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
	}

	return &evidence, nil
}

// GetEvidenceByDispute retrieves all evidence for a dispute
func (dk *disputeKeeper) GetEvidenceByDispute(ctx sdk.Context, disputeID string) ([]*billing.DisputeEvidence, error) {
	store := ctx.KVStore(dk.k.skey)
	prefix := billing.BuildDisputeEvidenceByDisputePrefix(disputeID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var evidenceList []*billing.DisputeEvidence
	for ; iter.Valid(); iter.Next() {
		evidenceID := string(iter.Value())
		evidence, err := dk.GetEvidence(ctx, evidenceID)
		if err != nil {
			continue
		}
		evidenceList = append(evidenceList, evidence)
	}

	return evidenceList, nil
}

// SubmitForReview submits a dispute for review
func (dk *disputeKeeper) SubmitForReview(ctx sdk.Context, disputeID string, submitter string) error {
	workflow, err := dk.GetDisputeWorkflow(ctx, disputeID)
	if err != nil {
		return err
	}

	if err := workflow.SubmitForReview(submitter, ctx.BlockTime()); err != nil {
		return err
	}

	return dk.UpdateDisputeWorkflow(ctx, workflow)
}

// ResolveDispute resolves a dispute
func (dk *disputeKeeper) ResolveDispute(
	ctx sdk.Context,
	disputeID string,
	resolution billing.DisputeResolutionType,
	details string,
	resolver string,
	refundAmount sdk.Coins,
) error {
	workflow, err := dk.GetDisputeWorkflow(ctx, disputeID)
	if err != nil {
		return err
	}

	// Resolve the dispute
	if err := workflow.Resolve(resolution, details, resolver, refundAmount, ctx.BlockTime()); err != nil {
		return err
	}

	// Save updated workflow
	if err := dk.UpdateDisputeWorkflow(ctx, workflow); err != nil {
		return err
	}

	// Get invoice keeper to update invoice status
	invoiceKeeper := dk.k.NewInvoiceKeeper()

	// Determine new invoice status based on resolution
	var newStatus billing.InvoiceStatus
	switch resolution {
	case billing.DisputeResolutionProviderWin:
		// Provider won, invoice remains payable
		newStatus = billing.InvoiceStatusPending
	case billing.DisputeResolutionCustomerWin:
		// Customer won, invoice may be cancelled or refunded
		if refundAmount.IsZero() {
			newStatus = billing.InvoiceStatusCancelled
		} else {
			newStatus = billing.InvoiceStatusRefunded
		}
	case billing.DisputeResolutionPartialRefund, billing.DisputeResolutionMutualAgreement:
		// Partial resolution, create a correction
		if !refundAmount.IsZero() {
			invoice, err := invoiceKeeper.GetInvoice(ctx, workflow.InvoiceID)
			if err != nil {
				return fmt.Errorf("failed to get invoice for correction: %w", err)
			}

			// Create correction
			correctionSeq := dk.GetCorrectionSequence(ctx)
			correctionID := billing.NextCorrectionID(correctionSeq, "VE")
			correction := billing.NewCorrection(
				correctionID,
				workflow.InvoiceID,
				"", // No settlement ID yet
				billing.CorrectionTypeRefundAdjustment,
				invoice.Total,
				invoice.Total.Sub(refundAmount...),
				fmt.Sprintf("dispute_resolution_%s", disputeID),
				details,
				resolver,
				ctx.BlockHeight(),
				ctx.BlockTime(),
			)

			if err := dk.CreateCorrection(ctx, correction); err != nil {
				return fmt.Errorf("failed to create correction: %w", err)
			}

			dk.SetCorrectionSequence(ctx, correctionSeq+1)
		}
		newStatus = billing.InvoiceStatusPending
	case billing.DisputeResolutionArbitration:
		// Arbitration decided, apply the decision
		if refundAmount.IsZero() {
			newStatus = billing.InvoiceStatusPending
		} else if refundAmount.Equal(workflow.DisputedAmount) {
			newStatus = billing.InvoiceStatusRefunded
		} else {
			newStatus = billing.InvoiceStatusPending
		}
	default:
		newStatus = billing.InvoiceStatusPending
	}

	// Update invoice status
	_, err = invoiceKeeper.UpdateInvoiceStatus(ctx, workflow.InvoiceID, newStatus, resolver)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	return nil
}

// EscalateDispute escalates a dispute
func (dk *disputeKeeper) EscalateDispute(ctx sdk.Context, disputeID string, escalateTo string, reason string) error {
	workflow, err := dk.GetDisputeWorkflow(ctx, disputeID)
	if err != nil {
		return err
	}

	// Get initiator from workflow for audit
	escalatedBy := workflow.InitiatedBy

	if err := workflow.Escalate(escalateTo, reason, escalatedBy, ctx.BlockTime()); err != nil {
		return err
	}

	return dk.UpdateDisputeWorkflow(ctx, workflow)
}

// CheckDisputeWindowOpen checks if the dispute window is open for an invoice
func (dk *disputeKeeper) CheckDisputeWindowOpen(ctx sdk.Context, invoiceID string) (bool, time.Duration, error) {
	invoiceKeeper := dk.k.NewInvoiceKeeper()

	invoice, err := invoiceKeeper.GetInvoice(ctx, invoiceID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get settlement config for dispute window settings
	config := billing.DefaultSettlementConfig()
	windowDuration := time.Duration(config.DefaultDisputeWindowSeconds) * time.Second

	// Calculate window end time from invoice issued date
	windowEndTime := invoice.IssuedAt.Add(windowDuration)
	now := ctx.BlockTime()

	if now.After(windowEndTime) {
		return false, 0, nil
	}

	remaining := windowEndTime.Sub(now)
	return true, remaining, nil
}

// CreateCorrection creates a new billing correction
func (dk *disputeKeeper) CreateCorrection(ctx sdk.Context, correction *billing.Correction) error {
	store := ctx.KVStore(dk.k.skey)

	// Check if correction already exists
	key := billing.BuildCorrectionKey(correction.CorrectionID)
	if store.Has(key) {
		return fmt.Errorf("correction already exists: %s", correction.CorrectionID)
	}

	// Validate correction
	if err := correction.Validate(); err != nil {
		return fmt.Errorf("invalid correction: %w", err)
	}

	// Marshal and store
	bz, err := json.Marshal(correction)
	if err != nil {
		return fmt.Errorf("failed to marshal correction: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	dk.setCorrectionIndexes(store, correction)

	return nil
}

// GetCorrection retrieves a correction by ID
func (dk *disputeKeeper) GetCorrection(ctx sdk.Context, correctionID string) (*billing.Correction, error) {
	store := ctx.KVStore(dk.k.skey)
	key := billing.BuildCorrectionKey(correctionID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("correction not found: %s", correctionID)
	}

	var correction billing.Correction
	if err := json.Unmarshal(bz, &correction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal correction: %w", err)
	}

	return &correction, nil
}

// ApplyCorrection applies a correction
func (dk *disputeKeeper) ApplyCorrection(ctx sdk.Context, correctionID string, appliedBy string) error {
	store := ctx.KVStore(dk.k.skey)

	correction, err := dk.GetCorrection(ctx, correctionID)
	if err != nil {
		return err
	}

	oldStatus := correction.Status

	// Apply the correction
	if err := correction.Apply(ctx.BlockTime()); err != nil {
		return err
	}

	// Update status index
	if oldStatus != correction.Status {
		oldStatusKey := billing.BuildCorrectionByStatusKey(oldStatus, correctionID)
		store.Delete(oldStatusKey)

		newStatusKey := billing.BuildCorrectionByStatusKey(correction.Status, correctionID)
		store.Set(newStatusKey, []byte(correctionID))
	}

	// Update approval information
	if correction.ApprovedBy == "" {
		correction.ApprovedBy = appliedBy
	}

	// Save updated correction
	key := billing.BuildCorrectionKey(correctionID)
	bz, err := json.Marshal(correction)
	if err != nil {
		return fmt.Errorf("failed to marshal correction: %w", err)
	}
	store.Set(key, bz)

	// Create ledger entry for the correction
	entryID := fmt.Sprintf("%s-applied-%d", correctionID, ctx.BlockHeight())
	entry := billing.NewCorrectionLedgerEntry(
		entryID,
		correctionID,
		correction.InvoiceID,
		billing.CorrectionLedgerEntryTypeApplied,
		oldStatus,
		correction.Status,
		correction.Difference,
		"correction applied",
		appliedBy,
		"",
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	dk.saveCorrectionLedgerEntry(store, entry)

	return nil
}

// GetCorrectionsByInvoice retrieves corrections for an invoice
func (dk *disputeKeeper) GetCorrectionsByInvoice(ctx sdk.Context, invoiceID string) ([]*billing.Correction, error) {
	store := ctx.KVStore(dk.k.skey)
	prefix := billing.BuildCorrectionByInvoicePrefix(invoiceID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var corrections []*billing.Correction
	for ; iter.Valid(); iter.Next() {
		correctionID := string(iter.Value())
		correction, err := dk.GetCorrection(ctx, correctionID)
		if err != nil {
			continue
		}
		corrections = append(corrections, correction)
	}

	return corrections, nil
}

// GetPendingCorrections retrieves pending corrections with pagination
func (dk *disputeKeeper) GetPendingCorrections(ctx sdk.Context, pagination *query.PageRequest) ([]*billing.Correction, *query.PageResponse, error) {
	store := ctx.KVStore(dk.k.skey)
	prefix := billing.BuildCorrectionByStatusPrefix(billing.CorrectionStatusPending)

	return dk.paginateCorrectionIndex(store, prefix, pagination)
}

// GetDisputeRules retrieves the dispute rules configuration
func (dk *disputeKeeper) GetDisputeRules(ctx sdk.Context) (*billing.DisputeRules, error) {
	store := ctx.KVStore(dk.k.skey)
	key := billing.BuildDisputeRulesKey("default")

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("dispute rules not found")
	}

	var rules billing.DisputeRules
	if err := json.Unmarshal(bz, &rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dispute rules: %w", err)
	}

	return &rules, nil
}

// SaveDisputeRules saves the dispute rules configuration
func (dk *disputeKeeper) SaveDisputeRules(ctx sdk.Context, rules *billing.DisputeRules) error {
	store := ctx.KVStore(dk.k.skey)

	if err := rules.Validate(); err != nil {
		return fmt.Errorf("invalid dispute rules: %w", err)
	}

	key := billing.BuildDisputeRulesKey("default")
	bz, err := json.Marshal(rules)
	if err != nil {
		return fmt.Errorf("failed to marshal dispute rules: %w", err)
	}
	store.Set(key, bz)

	return nil
}

// WithDisputes iterates over all disputes
func (dk *disputeKeeper) WithDisputes(ctx sdk.Context, fn func(*billing.DisputeWorkflow) bool) {
	store := ctx.KVStore(dk.k.skey)
	iter := storetypes.KVStorePrefixIterator(store, billing.DisputeWorkflowPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var workflow billing.DisputeWorkflow
		if err := json.Unmarshal(iter.Value(), &workflow); err != nil {
			continue
		}

		if stop := fn(&workflow); stop {
			break
		}
	}
}

// Helper methods

func (dk *disputeKeeper) setDisputeIndexes(store storetypes.KVStore, workflow *billing.DisputeWorkflow) {
	// Invoice index
	invoiceKey := billing.BuildDisputeWorkflowByInvoiceKey(workflow.InvoiceID, workflow.DisputeID)
	store.Set(invoiceKey, []byte(workflow.DisputeID))

	// Status index
	statusKey := billing.BuildDisputeWorkflowByStatusKey(workflow.Status, workflow.DisputeID)
	store.Set(statusKey, []byte(workflow.DisputeID))
}

func (dk *disputeKeeper) setCorrectionIndexes(store storetypes.KVStore, correction *billing.Correction) {
	// Invoice index
	invoiceKey := billing.BuildCorrectionByInvoiceKey(correction.InvoiceID, correction.CorrectionID)
	store.Set(invoiceKey, []byte(correction.CorrectionID))

	// Status index
	statusKey := billing.BuildCorrectionByStatusKey(correction.Status, correction.CorrectionID)
	store.Set(statusKey, []byte(correction.CorrectionID))

	// Requester index
	requesterKey := billing.BuildCorrectionByRequesterKey(correction.RequestedBy, correction.CorrectionID)
	store.Set(requesterKey, []byte(correction.CorrectionID))

	// Settlement index if present
	if correction.SettlementID != "" {
		settlementKey := billing.BuildCorrectionBySettlementKey(correction.SettlementID, correction.CorrectionID)
		store.Set(settlementKey, []byte(correction.CorrectionID))
	}
}

func (dk *disputeKeeper) saveCorrectionLedgerEntry(store storetypes.KVStore, entry *billing.CorrectionLedgerEntry) {
	entryKey := billing.BuildCorrectionLedgerEntryKey(entry.EntryID)
	//nolint:errchkjson // entry contains sdk.Coins which is safe for Marshal
	bz, _ := json.Marshal(entry)
	store.Set(entryKey, bz)

	// Create correction index
	indexKey := billing.BuildCorrectionLedgerEntryByCorrectionKey(entry.CorrectionID, entry.Timestamp.UnixNano())
	store.Set(indexKey, []byte(entry.EntryID))
}

//nolint:unparam // prefix kept for future index-specific pagination
func (dk *disputeKeeper) paginateDisputeIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.DisputeWorkflow, *query.PageResponse, error) {
	var workflows []*billing.DisputeWorkflow

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		disputeID := string(value)
		workflowKey := billing.BuildDisputeWorkflowKey(disputeID)
		bz := store.Get(workflowKey)
		if bz == nil {
			return nil
		}

		var workflow billing.DisputeWorkflow
		if err := json.Unmarshal(bz, &workflow); err != nil {
			return nil
		}

		workflows = append(workflows, &workflow)
		return nil
	})

	return workflows, pageRes, err
}

//nolint:unparam // prefix kept for future index-specific pagination
func (dk *disputeKeeper) paginateCorrectionIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.Correction, *query.PageResponse, error) {
	var corrections []*billing.Correction

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		correctionID := string(value)
		correctionKey := billing.BuildCorrectionKey(correctionID)
		bz := store.Get(correctionKey)
		if bz == nil {
			return nil
		}

		var correction billing.Correction
		if err := json.Unmarshal(bz, &correction); err != nil {
			return nil
		}

		corrections = append(corrections, &correction)
		return nil
	})

	return corrections, pageRes, err
}

// Sequence management for dispute IDs

// GetDisputeSequence gets the current dispute sequence number
func (dk *disputeKeeper) GetDisputeSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(dk.k.skey)
	bz := store.Get(billing.DisputeWorkflowSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetDisputeSequence sets the dispute sequence number
func (dk *disputeKeeper) SetDisputeSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(dk.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.DisputeWorkflowSequenceKey, bz)
}

// GetCorrectionSequence gets the current correction sequence number
func (dk *disputeKeeper) GetCorrectionSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(dk.k.skey)
	bz := store.Get(billing.CorrectionSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetCorrectionSequence sets the correction sequence number
func (dk *disputeKeeper) SetCorrectionSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(dk.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.CorrectionSequenceKey, bz)
}

// GetEvidenceSequence gets the current evidence sequence number
func (dk *disputeKeeper) GetEvidenceSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(dk.k.skey)
	bz := store.Get(billing.DisputeEvidenceSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetEvidenceSequence sets the evidence sequence number
func (dk *disputeKeeper) SetEvidenceSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(dk.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.DisputeEvidenceSequenceKey, bz)
}
