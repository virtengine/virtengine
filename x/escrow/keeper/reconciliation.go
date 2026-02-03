// Package keeper provides the escrow module keeper with reconciliation management capabilities.
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

// ReconciliationKeeper defines the interface for reconciliation management
type ReconciliationKeeper interface {
	// CreateReconciliationReport creates a new reconciliation report
	CreateReconciliationReport(ctx sdk.Context, report *billing.ReconciliationReport) error

	// GetReconciliationReport retrieves a reconciliation report by ID
	GetReconciliationReport(ctx sdk.Context, reportID string) (*billing.ReconciliationReport, error)

	// GetReconciliationReportsByProvider retrieves reports by provider
	GetReconciliationReportsByProvider(ctx sdk.Context, provider string, pagination *query.PageRequest) ([]*billing.ReconciliationReport, *query.PageResponse, error)

	// GetReconciliationReportsByPeriod retrieves reports by time period
	GetReconciliationReportsByPeriod(ctx sdk.Context, start, end time.Time, pagination *query.PageRequest) ([]*billing.ReconciliationReport, *query.PageResponse, error)

	// SaveUsageRecord saves a usage record
	SaveUsageRecord(ctx sdk.Context, record *billing.UsageRecord) error

	// GetUsageRecord retrieves a usage record by ID
	GetUsageRecord(ctx sdk.Context, recordID string) (*billing.UsageRecord, error)

	// GetUsageRecordsByLease retrieves usage records by lease ID
	GetUsageRecordsByLease(ctx sdk.Context, leaseID string) ([]*billing.UsageRecord, error)

	// GetUnreconciledUsageRecords retrieves all unreconciled (pending) usage records
	GetUnreconciledUsageRecords(ctx sdk.Context) ([]*billing.UsageRecord, error)

	// SavePayoutRecord saves a payout record
	SavePayoutRecord(ctx sdk.Context, record *billing.PayoutRecord) error

	// GetPayoutRecord retrieves a payout record by ID
	GetPayoutRecord(ctx sdk.Context, payoutID string) (*billing.PayoutRecord, error)

	// GetPayoutRecordsByProvider retrieves payout records by provider
	GetPayoutRecordsByProvider(ctx sdk.Context, provider string) ([]*billing.PayoutRecord, error)

	// RunReconciliationJob runs a reconciliation job for the specified period
	RunReconciliationJob(ctx sdk.Context, config billing.ReconciliationConfig, periodStart, periodEnd time.Time) (*billing.ReconciliationReport, error)

	// GetReconciliationJobConfig retrieves a reconciliation job configuration
	GetReconciliationJobConfig(ctx sdk.Context, jobID string) (*billing.ReconciliationJobConfig, error)

	// SaveReconciliationJobConfig saves a reconciliation job configuration
	SaveReconciliationJobConfig(ctx sdk.Context, config *billing.ReconciliationJobConfig) error

	// WithReconciliationReports iterates over all reconciliation reports
	WithReconciliationReports(ctx sdk.Context, fn func(*billing.ReconciliationReport) bool)
}

// reconciliationKeeper implements ReconciliationKeeper
type reconciliationKeeper struct {
	k *keeper
}

// NewReconciliationKeeper creates a new reconciliation keeper from the base keeper
func (k *keeper) NewReconciliationKeeper() ReconciliationKeeper {
	return &reconciliationKeeper{k: k}
}

// CreateReconciliationReport creates a new reconciliation report
func (rk *reconciliationKeeper) CreateReconciliationReport(ctx sdk.Context, report *billing.ReconciliationReport) error {
	store := ctx.KVStore(rk.k.skey)

	// Check if report already exists
	key := billing.BuildReconciliationReportKey(report.ReportID)
	if store.Has(key) {
		return fmt.Errorf("reconciliation report already exists: %s", report.ReportID)
	}

	// Validate report
	if err := report.Validate(); err != nil {
		return fmt.Errorf("invalid reconciliation report: %w", err)
	}

	// Set block height and generated at if not set
	if report.BlockHeight == 0 {
		report.BlockHeight = ctx.BlockHeight()
	}
	if report.GeneratedAt.IsZero() {
		report.GeneratedAt = ctx.BlockTime()
	}

	// Marshal and store
	bz, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal reconciliation report: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	rk.setReconciliationReportIndexes(store, report)

	return nil
}

// GetReconciliationReport retrieves a reconciliation report by ID
func (rk *reconciliationKeeper) GetReconciliationReport(ctx sdk.Context, reportID string) (*billing.ReconciliationReport, error) {
	store := ctx.KVStore(rk.k.skey)
	key := billing.BuildReconciliationReportKey(reportID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("reconciliation report not found: %s", reportID)
	}

	var report billing.ReconciliationReport
	if err := json.Unmarshal(bz, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reconciliation report: %w", err)
	}

	return &report, nil
}

// GetReconciliationReportsByProvider retrieves reports by provider
func (rk *reconciliationKeeper) GetReconciliationReportsByProvider(
	ctx sdk.Context,
	provider string,
	pagination *query.PageRequest,
) ([]*billing.ReconciliationReport, *query.PageResponse, error) {
	store := ctx.KVStore(rk.k.skey)
	prefix := rk.buildReconciliationReportByProviderPrefix(provider)

	return rk.paginateReconciliationReportIndex(store, prefix, pagination)
}

// GetReconciliationReportsByPeriod retrieves reports by time period
func (rk *reconciliationKeeper) GetReconciliationReportsByPeriod(
	ctx sdk.Context,
	start, end time.Time,
	pagination *query.PageRequest,
) ([]*billing.ReconciliationReport, *query.PageResponse, error) {
	store := ctx.KVStore(rk.k.skey)

	// Iterate over all reports and filter by period
	var reports []*billing.ReconciliationReport
	iter := storetypes.KVStorePrefixIterator(store, billing.ReconciliationReportPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report billing.ReconciliationReport
		if err := json.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}

		// Check if report period overlaps with requested period
		if rk.periodsOverlap(report.PeriodStart, report.PeriodEnd, start, end) {
			reports = append(reports, &report)
		}
	}

	// Apply pagination manually
	return rk.applyPagination(reports, pagination)
}

// SaveUsageRecord saves a usage record
func (rk *reconciliationKeeper) SaveUsageRecord(ctx sdk.Context, record *billing.UsageRecord) error {
	store := ctx.KVStore(rk.k.skey)

	// Validate record
	if err := record.Validate(); err != nil {
		return fmt.Errorf("invalid usage record: %w", err)
	}

	// Set timestamps if not set
	if record.CreatedAt.IsZero() {
		record.CreatedAt = ctx.BlockTime()
	}
	record.UpdatedAt = ctx.BlockTime()

	if record.BlockHeight == 0 {
		record.BlockHeight = ctx.BlockHeight()
	}

	// Marshal and store
	key := billing.BuildUsageRecordKey(record.RecordID)
	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal usage record: %w", err)
	}
	store.Set(key, bz)

	// Create lease index
	leaseKey := billing.BuildUsageRecordByLeaseKey(record.LeaseID, record.RecordID)
	store.Set(leaseKey, []byte(record.RecordID))

	return nil
}

// GetUsageRecord retrieves a usage record by ID
func (rk *reconciliationKeeper) GetUsageRecord(ctx sdk.Context, recordID string) (*billing.UsageRecord, error) {
	store := ctx.KVStore(rk.k.skey)
	key := billing.BuildUsageRecordKey(recordID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("usage record not found: %s", recordID)
	}

	var record billing.UsageRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage record: %w", err)
	}

	return &record, nil
}

// GetUsageRecordsByLease retrieves usage records by lease ID
func (rk *reconciliationKeeper) GetUsageRecordsByLease(ctx sdk.Context, leaseID string) ([]*billing.UsageRecord, error) {
	store := ctx.KVStore(rk.k.skey)
	prefix := billing.BuildUsageRecordByLeasePrefix(leaseID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var records []*billing.UsageRecord
	for ; iter.Valid(); iter.Next() {
		recordID := string(iter.Value())
		record, err := rk.GetUsageRecord(ctx, recordID)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// GetUnreconciledUsageRecords retrieves all unreconciled (pending) usage records
func (rk *reconciliationKeeper) GetUnreconciledUsageRecords(ctx sdk.Context) ([]*billing.UsageRecord, error) {
	store := ctx.KVStore(rk.k.skey)

	iter := storetypes.KVStorePrefixIterator(store, billing.UsageRecordPrefix)
	defer iter.Close()

	var records []*billing.UsageRecord
	for ; iter.Valid(); iter.Next() {
		var record billing.UsageRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		// Only include pending records
		if record.Status == billing.UsageRecordStatusPending {
			records = append(records, &record)
		}
	}

	return records, nil
}

// SavePayoutRecord saves a payout record
func (rk *reconciliationKeeper) SavePayoutRecord(ctx sdk.Context, record *billing.PayoutRecord) error {
	store := ctx.KVStore(rk.k.skey)

	// Validate record
	if err := record.Validate(); err != nil {
		return fmt.Errorf("invalid payout record: %w", err)
	}

	// Set timestamps if not set
	if record.CreatedAt.IsZero() {
		record.CreatedAt = ctx.BlockTime()
	}
	record.UpdatedAt = ctx.BlockTime()

	if record.BlockHeight == 0 {
		record.BlockHeight = ctx.BlockHeight()
	}

	// Marshal and store
	key := billing.BuildPayoutRecordKey(record.PayoutID)
	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal payout record: %w", err)
	}
	store.Set(key, bz)

	// Create provider index
	providerKey := billing.BuildPayoutRecordByProviderKey(record.Provider, record.PayoutID)
	store.Set(providerKey, []byte(record.PayoutID))

	return nil
}

// GetPayoutRecord retrieves a payout record by ID
func (rk *reconciliationKeeper) GetPayoutRecord(ctx sdk.Context, payoutID string) (*billing.PayoutRecord, error) {
	store := ctx.KVStore(rk.k.skey)
	key := billing.BuildPayoutRecordKey(payoutID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("payout record not found: %s", payoutID)
	}

	var record billing.PayoutRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payout record: %w", err)
	}

	return &record, nil
}

// GetPayoutRecordsByProvider retrieves payout records by provider
func (rk *reconciliationKeeper) GetPayoutRecordsByProvider(ctx sdk.Context, provider string) ([]*billing.PayoutRecord, error) {
	store := ctx.KVStore(rk.k.skey)
	prefix := billing.BuildPayoutRecordByProviderPrefix(provider)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var records []*billing.PayoutRecord
	for ; iter.Valid(); iter.Next() {
		payoutID := string(iter.Value())
		record, err := rk.GetPayoutRecord(ctx, payoutID)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// RunReconciliationJob runs a reconciliation job for the specified period
func (rk *reconciliationKeeper) RunReconciliationJob(
	ctx sdk.Context,
	config billing.ReconciliationConfig,
	periodStart, periodEnd time.Time,
) (*billing.ReconciliationReport, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid reconciliation config: %w", err)
	}

	// Get all usage records for the period
	usageRecords := rk.getUsageRecordsForPeriod(ctx, periodStart, periodEnd)

	// Get all invoices for the period using the invoice keeper
	invoices := rk.getInvoicesForPeriod(ctx, periodStart, periodEnd)

	// Get all payout records for the period
	payoutRecords := rk.getPayoutRecordsForPeriod(ctx, periodStart, periodEnd)

	// Compare and find discrepancies
	discrepancies := rk.findDiscrepancies(ctx, config, usageRecords, invoices, payoutRecords)

	// Calculate summary statistics
	summary := rk.calculateSummary(usageRecords, invoices, payoutRecords, discrepancies)

	// Determine status based on discrepancies
	status := billing.ReconciliationStatusComplete
	if len(discrepancies) > 0 {
		if len(discrepancies) > int(config.MaxDiscrepanciesBeforeFail) {
			status = billing.ReconciliationStatusFailed
		} else {
			status = billing.ReconciliationStatusPartial
		}
	}

	// Collect IDs
	invoiceIDs := make([]string, 0, len(invoices))
	for _, inv := range invoices {
		invoiceIDs = append(invoiceIDs, inv.InvoiceID)
	}

	usageRecordIDs := make([]string, 0, len(usageRecords))
	for _, rec := range usageRecords {
		usageRecordIDs = append(usageRecordIDs, rec.RecordID)
	}

	// Generate report ID
	reportID := fmt.Sprintf("reconciliation-%d-%d", periodStart.Unix(), ctx.BlockHeight())

	// Create and save report
	report := &billing.ReconciliationReport{
		ReportID:       reportID,
		ReportType:     billing.ReconciliationReportTypeOnDemand,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Status:         status,
		Summary:        summary,
		Discrepancies:  discrepancies,
		InvoiceIDs:     invoiceIDs,
		UsageRecordIDs: usageRecordIDs,
		GeneratedAt:    ctx.BlockTime(),
		GeneratedBy:    "system",
		BlockHeight:    ctx.BlockHeight(),
	}

	if err := rk.CreateReconciliationReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save reconciliation report: %w", err)
	}

	return report, nil
}

// GetReconciliationJobConfig retrieves a reconciliation job configuration
func (rk *reconciliationKeeper) GetReconciliationJobConfig(ctx sdk.Context, jobID string) (*billing.ReconciliationJobConfig, error) {
	store := ctx.KVStore(rk.k.skey)
	key := billing.BuildReconciliationJobKey(jobID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("reconciliation job config not found: %s", jobID)
	}

	var config billing.ReconciliationJobConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reconciliation job config: %w", err)
	}

	return &config, nil
}

// SaveReconciliationJobConfig saves a reconciliation job configuration
func (rk *reconciliationKeeper) SaveReconciliationJobConfig(ctx sdk.Context, config *billing.ReconciliationJobConfig) error {
	store := ctx.KVStore(rk.k.skey)

	// Validate config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid reconciliation job config: %w", err)
	}

	// Set timestamps if not set
	if config.CreatedAt.IsZero() {
		config.CreatedAt = ctx.BlockTime()
	}
	config.UpdatedAt = ctx.BlockTime()

	// Marshal and store
	key := billing.BuildReconciliationJobKey(config.JobID)
	bz, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal reconciliation job config: %w", err)
	}
	store.Set(key, bz)

	return nil
}

// WithReconciliationReports iterates over all reconciliation reports
func (rk *reconciliationKeeper) WithReconciliationReports(ctx sdk.Context, fn func(*billing.ReconciliationReport) bool) {
	store := ctx.KVStore(rk.k.skey)
	iter := storetypes.KVStorePrefixIterator(store, billing.ReconciliationReportPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report billing.ReconciliationReport
		if err := json.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}

		if stop := fn(&report); stop {
			break
		}
	}
}

// Helper methods

func (rk *reconciliationKeeper) setReconciliationReportIndexes(store storetypes.KVStore, report *billing.ReconciliationReport) {
	// Provider index
	if report.Provider != "" {
		providerKey := billing.BuildReconciliationReportByProviderKey(report.Provider, report.ReportID)
		store.Set(providerKey, []byte(report.ReportID))
	}

	// Customer index
	if report.Customer != "" {
		customerKey := billing.BuildReconciliationReportByCustomerKey(report.Customer, report.ReportID)
		store.Set(customerKey, []byte(report.ReportID))
	}
}

func (rk *reconciliationKeeper) buildReconciliationReportByProviderPrefix(provider string) []byte {
	key := make([]byte, 0, len(billing.ReconciliationReportByProviderPrefix)+len(provider)+1)
	key = append(key, billing.ReconciliationReportByProviderPrefix...)
	key = append(key, []byte(provider)...)
	return append(key, byte('/'))
}

//nolint:unparam // prefix kept for future index-specific pagination
func (rk *reconciliationKeeper) paginateReconciliationReportIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.ReconciliationReport, *query.PageResponse, error) {
	var reports []*billing.ReconciliationReport

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		reportID := string(value)
		reportKey := billing.BuildReconciliationReportKey(reportID)
		bz := store.Get(reportKey)
		if bz == nil {
			return nil
		}

		var report billing.ReconciliationReport
		if err := json.Unmarshal(bz, &report); err != nil {
			return nil
		}

		reports = append(reports, &report)
		return nil
	})

	return reports, pageRes, err
}

func (rk *reconciliationKeeper) periodsOverlap(start1, end1, start2, end2 time.Time) bool {
	return !start1.After(end2) && !start2.After(end1)
}

//nolint:unparam // result 2 (error) reserved for future pagination failures
func (rk *reconciliationKeeper) applyPagination(
	reports []*billing.ReconciliationReport,
	pagination *query.PageRequest,
) ([]*billing.ReconciliationReport, *query.PageResponse, error) {
	total := uint64(len(reports))

	if pagination == nil {
		return reports, &query.PageResponse{Total: total}, nil
	}

	offset := pagination.Offset
	limit := pagination.Limit
	if limit == 0 {
		limit = 100 // default limit
	}

	if offset >= total {
		return nil, &query.PageResponse{Total: total}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	result := reports[offset:end]

	var nextKey []byte
	if end < total {
		nextKey = []byte(fmt.Sprintf("%d", end))
	}

	return result, &query.PageResponse{
		NextKey: nextKey,
		Total:   total,
	}, nil
}

func (rk *reconciliationKeeper) getUsageRecordsForPeriod(
	ctx sdk.Context,
	periodStart, periodEnd time.Time,
) []*billing.UsageRecord {
	store := ctx.KVStore(rk.k.skey)

	iter := storetypes.KVStorePrefixIterator(store, billing.UsageRecordPrefix)
	defer iter.Close()

	var records []*billing.UsageRecord
	for ; iter.Valid(); iter.Next() {
		var record billing.UsageRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		// Check if record falls within the period
		if rk.periodsOverlap(record.StartTime, record.EndTime, periodStart, periodEnd) {
			records = append(records, &record)
		}
	}

	return records
}

func (rk *reconciliationKeeper) getInvoicesForPeriod(
	ctx sdk.Context,
	periodStart, periodEnd time.Time,
) []*billing.InvoiceLedgerRecord {
	store := ctx.KVStore(rk.k.skey)

	iter := storetypes.KVStorePrefixIterator(store, billing.InvoiceLedgerRecordPrefix)
	defer iter.Close()

	var invoices []*billing.InvoiceLedgerRecord
	for ; iter.Valid(); iter.Next() {
		var invoice billing.InvoiceLedgerRecord
		if err := json.Unmarshal(iter.Value(), &invoice); err != nil {
			continue
		}

		// Check if invoice billing period falls within the requested period
		if rk.periodsOverlap(invoice.BillingPeriodStart, invoice.BillingPeriodEnd, periodStart, periodEnd) {
			invoices = append(invoices, &invoice)
		}
	}

	return invoices
}

func (rk *reconciliationKeeper) getPayoutRecordsForPeriod(
	ctx sdk.Context,
	periodStart, periodEnd time.Time,
) []*billing.PayoutRecord {
	store := ctx.KVStore(rk.k.skey)

	iter := storetypes.KVStorePrefixIterator(store, billing.PayoutRecordPrefix)
	defer iter.Close()

	var records []*billing.PayoutRecord
	for ; iter.Valid(); iter.Next() {
		var record billing.PayoutRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		// Check if payout date falls within the period
		if !record.PayoutDate.Before(periodStart) && !record.PayoutDate.After(periodEnd) {
			records = append(records, &record)
		}
	}

	return records
}

//nolint:unparam // ctx kept for future on-chain state queries
func (rk *reconciliationKeeper) findDiscrepancies(
	ctx sdk.Context,
	config billing.ReconciliationConfig,
	usageRecords []*billing.UsageRecord,
	invoices []*billing.InvoiceLedgerRecord,
	payoutRecords []*billing.PayoutRecord,
) []billing.ReconciliationDiscrepancy {
	var discrepancies []billing.ReconciliationDiscrepancy
	discrepancyCount := 0

	// Create maps for efficient lookup
	usageByLease := make(map[string][]*billing.UsageRecord)
	for _, rec := range usageRecords {
		usageByLease[rec.LeaseID] = append(usageByLease[rec.LeaseID], rec)
	}

	invoicesByLease := make(map[string][]*billing.InvoiceLedgerRecord)
	for _, inv := range invoices {
		invoicesByLease[inv.LeaseID] = append(invoicesByLease[inv.LeaseID], inv)
	}

	// Check for usage records without corresponding invoices
	for leaseID, usageRecs := range usageByLease {
		leaseInvoices, hasInvoices := invoicesByLease[leaseID]

		if !hasInvoices && config.IncludeUnmatchedRecords {
			// Missing invoice for usage
			for _, rec := range usageRecs {
				discrepancyCount++
				discrepancies = append(discrepancies, billing.ReconciliationDiscrepancy{
					DiscrepancyID:  fmt.Sprintf("disc-%d", discrepancyCount),
					Type:           billing.DiscrepancyTypeMissingInvoice,
					UsageRecordID:  rec.RecordID,
					ExpectedAmount: rec.TotalAmount,
					ActualAmount:   sdk.NewCoins(),
					Difference:     rec.TotalAmount,
					Description:    fmt.Sprintf("usage record %s has no corresponding invoice", rec.RecordID),
					Severity:       billing.DiscrepancySeverityHigh,
				})
			}
		} else if hasInvoices {
			// Compare usage amounts with invoice amounts
			totalUsage := sdk.NewCoins()
			for _, rec := range usageRecs {
				totalUsage = totalUsage.Add(rec.TotalAmount...)
			}

			totalInvoiced := sdk.NewCoins()
			for _, inv := range leaseInvoices {
				totalInvoiced = totalInvoiced.Add(inv.Total...)
			}

			// Check for amount mismatch
			variance := billing.CalculateTotalVariance(totalUsage, totalInvoiced)
			if variance.GT(config.VarianceThreshold) {
				discrepancyCount++
				discrepancies = append(discrepancies, billing.CreateDiscrepancyFromComparison(
					fmt.Sprintf("disc-%d", discrepancyCount),
					billing.DiscrepancyTypeAmountMismatch,
					"",
					"",
					totalUsage,
					totalInvoiced,
					fmt.Sprintf("lease %s: usage total does not match invoice total", leaseID),
				))
			}
		}
	}

	// Check for invoices without corresponding usage records
	if config.IncludeUnmatchedRecords {
		for leaseID, leaseInvoices := range invoicesByLease {
			if _, hasUsage := usageByLease[leaseID]; !hasUsage {
				for _, inv := range leaseInvoices {
					discrepancyCount++
					discrepancies = append(discrepancies, billing.ReconciliationDiscrepancy{
						DiscrepancyID:  fmt.Sprintf("disc-%d", discrepancyCount),
						Type:           billing.DiscrepancyTypeMissingUsage,
						InvoiceID:      inv.InvoiceID,
						ExpectedAmount: inv.Total,
						ActualAmount:   sdk.NewCoins(),
						Difference:     inv.Total,
						Description:    fmt.Sprintf("invoice %s has no corresponding usage records", inv.InvoiceID),
						Severity:       billing.DiscrepancySeverityMedium,
					})
				}
			}
		}
	}

	// Check payouts against paid invoices
	paidInvoicesByProvider := make(map[string]sdk.Coins)
	for _, inv := range invoices {
		if inv.Status == billing.InvoiceStatusPaid {
			paidInvoicesByProvider[inv.Provider] = paidInvoicesByProvider[inv.Provider].Add(inv.AmountPaid...)
		}
	}

	payoutsByProvider := make(map[string]sdk.Coins)
	for _, payout := range payoutRecords {
		if payout.Status == billing.PayoutStatusCompleted {
			payoutsByProvider[payout.Provider] = payoutsByProvider[payout.Provider].Add(payout.PayoutAmount...)
		}
	}

	// Compare paid invoices to payouts
	for provider, paidAmount := range paidInvoicesByProvider {
		payoutAmount := payoutsByProvider[provider]

		variance := billing.CalculateTotalVariance(paidAmount, payoutAmount)
		if variance.GT(config.VarianceThreshold) {
			discrepancyCount++
			discrepancies = append(discrepancies, billing.CreateDiscrepancyFromComparison(
				fmt.Sprintf("disc-%d", discrepancyCount),
				billing.DiscrepancyTypeAmountMismatch,
				"",
				"",
				paidAmount,
				payoutAmount,
				fmt.Sprintf("provider %s: paid invoice total does not match payout total", provider),
			))
		}
	}

	return discrepancies
}

func (rk *reconciliationKeeper) calculateSummary(
	usageRecords []*billing.UsageRecord,
	invoices []*billing.InvoiceLedgerRecord,
	payoutRecords []*billing.PayoutRecord,
	discrepancies []billing.ReconciliationDiscrepancy,
) billing.ReconciliationSummary {
	summary := billing.ReconciliationSummary{
		//nolint:gosec // slice lengths are non-negative
		TotalUsageRecords: uint32(len(usageRecords)),
		//nolint:gosec // slice lengths are non-negative
		TotalInvoices: uint32(len(invoices)),
		//nolint:gosec // slice lengths are non-negative
		TotalSettlements:      uint32(len(payoutRecords)),
		TotalInvoiceAmount:    sdk.NewCoins(),
		TotalSettlementAmount: sdk.NewCoins(),
		TotalUsageAmount:      sdk.NewCoins(),
		PaidAmount:            sdk.NewCoins(),
		OutstandingAmount:     sdk.NewCoins(),
		DisputedAmount:        sdk.NewCoins(),
		OverdueAmount:         sdk.NewCoins(),
		DiscrepancyAmount:     sdk.NewCoins(),
		//nolint:gosec // slice length is non-negative
		DiscrepancyCount: uint32(len(discrepancies)),
	}

	// Calculate usage totals
	for _, rec := range usageRecords {
		summary.TotalUsageAmount = summary.TotalUsageAmount.Add(rec.TotalAmount...)
	}

	// Calculate invoice totals
	for _, inv := range invoices {
		summary.TotalInvoiceAmount = summary.TotalInvoiceAmount.Add(inv.Total...)

		switch inv.Status {
		case billing.InvoiceStatusPaid:
			summary.PaidInvoices++
			summary.PaidAmount = summary.PaidAmount.Add(inv.AmountPaid...)
		case billing.InvoiceStatusDisputed:
			summary.DisputedInvoices++
			summary.DisputedAmount = summary.DisputedAmount.Add(inv.Total...)
		case billing.InvoiceStatusOverdue:
			summary.OverdueInvoices++
			summary.OverdueAmount = summary.OverdueAmount.Add(inv.AmountDue...)
		default:
			if !inv.Status.IsTerminal() {
				summary.OutstandingInvoices++
				summary.OutstandingAmount = summary.OutstandingAmount.Add(inv.AmountDue...)
			}
		}
	}

	// Calculate payout totals
	for _, payout := range payoutRecords {
		if payout.Status == billing.PayoutStatusCompleted {
			summary.TotalSettlementAmount = summary.TotalSettlementAmount.Add(payout.NetAmount...)
		}
	}

	// Calculate discrepancy totals
	for _, disc := range discrepancies {
		summary.DiscrepancyAmount = summary.DiscrepancyAmount.Add(disc.Difference...)
	}

	return summary
}
