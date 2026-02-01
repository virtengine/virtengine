// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Reconciliation Service
// ============================================================================

// ReconciliationService handles reconciliation between on-chain payouts and provider records.
type ReconciliationService struct {
	config    ReconciliationConfig
	payouts   PayoutStore
	reconcile ReconciliationStore
	providers map[ProviderType]Provider
}

// NewReconciliationService creates a new reconciliation service.
func NewReconciliationService(
	config ReconciliationConfig,
	payouts PayoutStore,
	reconcile ReconciliationStore,
	providers map[ProviderType]Provider,
) *ReconciliationService {
	return &ReconciliationService{
		config:    config,
		payouts:   payouts,
		reconcile: reconcile,
		providers: providers,
	}
}

// Run executes the reconciliation job.
func (s *ReconciliationService) Run(ctx context.Context) (*ReconciliationResult, error) {
	if !s.config.Enabled {
		return &ReconciliationResult{}, nil
	}

	startTime := time.Now()
	result := &ReconciliationResult{
		Records: make([]*ReconciliationRecord, 0),
	}

	// Get pending payouts for reconciliation
	pendingPayouts, err := s.payouts.ListPendingReconciliation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending payouts: %w", err)
	}

	// Group payouts by provider
	byProvider := make(map[ProviderType][]*PayoutIntent)
	for _, payout := range pendingPayouts {
		byProvider[payout.Provider] = append(byProvider[payout.Provider], payout)
	}

	// Reconcile each provider
	for providerType, payouts := range byProvider {
		provider, ok := s.providers[providerType]
		if !ok {
			result.Errors++
			continue
		}

		records, err := s.reconcileProvider(ctx, provider, payouts)
		if err != nil {
			result.Errors++
			continue
		}

		for _, record := range records {
			result.PayoutsProcessed++

			switch record.Status {
			case ReconciliationMatched:
				result.Matched++
			case ReconciliationMismatch:
				result.Mismatched++
			case ReconciliationMissing:
				result.Missing++
			}

			result.Records = append(result.Records, record)
		}
	}

	result.Duration = time.Since(startTime).String()

	return result, nil
}

// reconcileProvider reconciles payouts for a single provider.
func (s *ReconciliationService) reconcileProvider(
	ctx context.Context,
	provider Provider,
	payouts []*PayoutIntent,
) ([]*ReconciliationRecord, error) {
	if len(payouts) == 0 {
		return nil, nil
	}

	// Determine date range
	var earliestDate, latestDate time.Time
	for _, payout := range payouts {
		if earliestDate.IsZero() || payout.CreatedAt.Before(earliestDate) {
			earliestDate = payout.CreatedAt
		}
		if latestDate.IsZero() || payout.CreatedAt.After(latestDate) {
			latestDate = payout.CreatedAt
		}
	}

	// Add buffer days
	startDate := earliestDate.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := latestDate.AddDate(0, 0, 1).Format("2006-01-02")

	// Get settlement report from provider
	report, err := provider.GetSettlementReport(ctx, SettlementReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get settlement report: %w", err)
	}

	// Build transaction lookup map
	txByID := make(map[string]SettlementTransaction)
	for _, tx := range report.Transactions {
		if tx.PayoutID != "" {
			txByID[tx.PayoutID] = tx
		}
		if tx.TransactionID != "" {
			txByID[tx.TransactionID] = tx
		}
	}

	// Reconcile each payout
	records := make([]*ReconciliationRecord, 0, len(payouts))
	for _, payout := range payouts {
		record := s.reconcilePayout(ctx, payout, txByID)
		records = append(records, record)

		// Save record
		if err := s.reconcile.Save(ctx, record); err != nil {
			// Log but continue
		}
	}

	return records, nil
}

// reconcilePayout reconciles a single payout.
//
//nolint:unparam // ctx kept for future async payout verification
func (s *ReconciliationService) reconcilePayout(
	_ context.Context,
	payout *PayoutIntent,
	txByID map[string]SettlementTransaction,
) *ReconciliationRecord {
	now := time.Now()

	record := &ReconciliationRecord{
		ID:            fmt.Sprintf("rec_%s", payout.ID),
		PayoutID:      payout.ID,
		OnChainAmount: payout.FiatAmount.Value,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Try to find matching transaction
	var tx SettlementTransaction
	var found bool

	// Try by payout ID
	if tx, found = txByID[payout.ID]; !found {
		// Try by provider payout ID
		if tx, found = txByID[payout.ProviderPayoutID]; !found {
			// Transaction not found
			record.Status = ReconciliationMissing
			record.Notes = "Transaction not found in provider settlement report"
			return record
		}
	}

	record.ProviderAmount = tx.Amount

	// Check for discrepancy
	discrepancy := payout.FiatAmount.Value - tx.Amount
	if discrepancy < 0 {
		discrepancy = -discrepancy
	}
	record.Discrepancy = discrepancy

	if discrepancy <= s.config.DiscrepancyThreshold {
		record.Status = ReconciliationMatched
		if s.config.AutoResolveMatches {
			record.ReconciledAt = &now
			record.ReconciledBy = "auto"
		}
	} else {
		record.Status = ReconciliationMismatch
		record.Notes = fmt.Sprintf("Discrepancy of %d exceeds threshold of %d", discrepancy, s.config.DiscrepancyThreshold)
	}

	return record
}

// ResolveRecord manually resolves a reconciliation record.
func (s *ReconciliationService) ResolveRecord(
	ctx context.Context,
	recordID string,
	resolvedBy string,
	notes string,
) error {
	record, err := s.reconcile.GetByPayoutID(ctx, recordID)
	if err != nil {
		return err
	}

	now := time.Now()
	record.Status = ReconciliationMatched
	record.ReconciledAt = &now
	record.ReconciledBy = resolvedBy
	record.Notes = notes
	record.UpdatedAt = now

	return s.reconcile.Save(ctx, record)
}

// GetMismatches returns all mismatched records.
func (s *ReconciliationService) GetMismatches(ctx context.Context) ([]*ReconciliationRecord, error) {
	return s.reconcile.ListMismatches(ctx)
}

// ============================================================================
// Scheduled Reconciliation Job
// ============================================================================

// ReconciliationJob is a scheduled reconciliation job.
type ReconciliationJob struct {
	service  *ReconciliationService
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}

	// Last run info
	lastRun    time.Time
	lastResult *ReconciliationResult
	lastError  error
}

// NewReconciliationJob creates a new reconciliation job.
func NewReconciliationJob(service *ReconciliationService, interval time.Duration) *ReconciliationJob {
	return &ReconciliationJob{
		service:  service,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start starts the reconciliation job.
func (j *ReconciliationJob) Start(ctx context.Context) {
	go j.run(ctx)
}

// Stop stops the reconciliation job.
func (j *ReconciliationJob) Stop() {
	close(j.stopCh)
	<-j.doneCh
}

// run is the job loop.
func (j *ReconciliationJob) run(ctx context.Context) {
	defer close(j.doneCh)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Run immediately on start
	j.runOnce(ctx)

	for {
		select {
		case <-ticker.C:
			j.runOnce(ctx)
		case <-j.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runOnce runs a single reconciliation.
func (j *ReconciliationJob) runOnce(ctx context.Context) {
	j.lastRun = time.Now()
	result, err := j.service.Run(ctx)
	j.lastResult = result
	j.lastError = err
}

// LastRun returns the time of the last run.
func (j *ReconciliationJob) LastRun() time.Time {
	return j.lastRun
}

// LastResult returns the result of the last run.
func (j *ReconciliationJob) LastResult() *ReconciliationResult {
	return j.lastResult
}

// LastError returns the error from the last run.
func (j *ReconciliationJob) LastError() error {
	return j.lastError
}

// ============================================================================
// Reconciliation Report Generator
// ============================================================================

// ReconciliationReportGenerator generates reconciliation reports.
type ReconciliationReportGenerator struct {
	store ReconciliationStore
}

// NewReconciliationReportGenerator creates a new report generator.
func NewReconciliationReportGenerator(store ReconciliationStore) *ReconciliationReportGenerator {
	return &ReconciliationReportGenerator{
		store: store,
	}
}

// GenerateReport generates a reconciliation report.
type ReconciliationReport struct {
	// GeneratedAt is when the report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// Period is the reporting period
	Period string `json:"period"`

	// Summary contains summary statistics
	Summary ReconciliationSummary `json:"summary"`

	// Mismatches contains unresolved mismatches
	Mismatches []*ReconciliationRecord `json:"mismatches,omitempty"`

	// Missing contains missing records
	Missing []*ReconciliationRecord `json:"missing,omitempty"`
}

// ReconciliationSummary contains reconciliation summary statistics.
type ReconciliationSummary struct {
	// TotalRecords is the total number of records
	TotalRecords int `json:"total_records"`

	// Matched is the number of matched records
	Matched int `json:"matched"`

	// Mismatched is the number of mismatched records
	Mismatched int `json:"mismatched"`

	// Missing is the number of missing records
	Missing int `json:"missing"`

	// Reviewing is the number of records under review
	Reviewing int `json:"reviewing"`

	// TotalDiscrepancy is the total discrepancy amount
	TotalDiscrepancy int64 `json:"total_discrepancy"`

	// MatchRate is the match rate percentage
	MatchRate float64 `json:"match_rate"`
}

// Generate generates a reconciliation report.
func (g *ReconciliationReportGenerator) Generate(ctx context.Context, startDate, endDate time.Time) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		GeneratedAt: time.Now(),
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}

	// Get mismatches
	mismatches, err := g.store.ListByStatus(ctx, ReconciliationMismatch)
	if err != nil {
		return nil, err
	}
	report.Mismatches = mismatches
	report.Summary.Mismatched = len(mismatches)

	// Get missing
	missing, err := g.store.ListByStatus(ctx, ReconciliationMissing)
	if err != nil {
		return nil, err
	}
	report.Missing = missing
	report.Summary.Missing = len(missing)

	// Get matched
	matched, err := g.store.ListByStatus(ctx, ReconciliationMatched)
	if err != nil {
		return nil, err
	}
	report.Summary.Matched = len(matched)

	// Get reviewing
	reviewing, err := g.store.ListByStatus(ctx, ReconciliationReviewing)
	if err != nil {
		return nil, err
	}
	report.Summary.Reviewing = len(reviewing)

	// Calculate totals
	report.Summary.TotalRecords = report.Summary.Matched + report.Summary.Mismatched + report.Summary.Missing + report.Summary.Reviewing

	// Calculate total discrepancy
	for _, record := range mismatches {
		report.Summary.TotalDiscrepancy += record.Discrepancy
	}

	// Calculate match rate
	if report.Summary.TotalRecords > 0 {
		report.Summary.MatchRate = float64(report.Summary.Matched) / float64(report.Summary.TotalRecords) * 100
	}

	return report, nil
}

