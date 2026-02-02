// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ReconciliationConfig defines configuration for reconciliation operations
type ReconciliationConfig struct {
	// VarianceThreshold is the percentage threshold for flagging discrepancies (e.g., 0.01 for 1%)
	VarianceThreshold sdkmath.LegacyDec `json:"variance_threshold"`

	// AutoResolveThreshold is the maximum variance percentage that can be auto-resolved
	AutoResolveThreshold sdkmath.LegacyDec `json:"auto_resolve_threshold"`

	// MaxDiscrepanciesBeforeFail is the maximum number of discrepancies before the reconciliation fails
	MaxDiscrepanciesBeforeFail uint32 `json:"max_discrepancies_before_fail"`

	// RequireManualReviewAbove is the amount threshold above which manual review is required
	RequireManualReviewAbove sdk.Coins `json:"require_manual_review_above"`

	// IncludeUnmatchedRecords indicates whether to include unmatched records in the report
	IncludeUnmatchedRecords bool `json:"include_unmatched_records"`

	// GenerateAlerts indicates whether to generate alerts for discrepancies
	GenerateAlerts bool `json:"generate_alerts"`
}

// DefaultReconciliationConfig returns the default reconciliation configuration
func DefaultReconciliationConfig() ReconciliationConfig {
	return ReconciliationConfig{
		VarianceThreshold:          sdkmath.LegacyNewDecWithPrec(1, 2), // 0.01 = 1%
		AutoResolveThreshold:       sdkmath.LegacyNewDecWithPrec(1, 3), // 0.001 = 0.1%
		MaxDiscrepanciesBeforeFail: 100,
		RequireManualReviewAbove:   sdk.NewCoins(sdk.NewCoin(DefaultCurrency, sdkmath.NewInt(1000000000))), // 1000 virt
		IncludeUnmatchedRecords:    true,
		GenerateAlerts:             true,
	}
}

// Validate validates the reconciliation configuration
func (c *ReconciliationConfig) Validate() error {
	if c.VarianceThreshold.IsNegative() {
		return fmt.Errorf("variance_threshold cannot be negative")
	}

	if c.VarianceThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("variance_threshold cannot exceed 1.0 (100%%)")
	}

	if c.AutoResolveThreshold.IsNegative() {
		return fmt.Errorf("auto_resolve_threshold cannot be negative")
	}

	if c.AutoResolveThreshold.GT(c.VarianceThreshold) {
		return fmt.Errorf("auto_resolve_threshold cannot exceed variance_threshold")
	}

	if !c.RequireManualReviewAbove.IsValid() {
		return fmt.Errorf("require_manual_review_above must be valid coins")
	}

	return nil
}

// UsageRecordStatus defines the status of a usage record
type UsageRecordStatus uint8

const (
	// UsageRecordStatusPending is a pending usage record
	UsageRecordStatusPending UsageRecordStatus = 0

	// UsageRecordStatusInvoiced is an invoiced usage record
	UsageRecordStatusInvoiced UsageRecordStatus = 1

	// UsageRecordStatusSettled is a settled usage record
	UsageRecordStatusSettled UsageRecordStatus = 2
)

// UsageRecordStatusNames maps status to names
var UsageRecordStatusNames = map[UsageRecordStatus]string{
	UsageRecordStatusPending:  "pending",
	UsageRecordStatusInvoiced: "invoiced",
	UsageRecordStatusSettled:  "settled",
}

// String returns string representation
func (s UsageRecordStatus) String() string {
	if name, ok := UsageRecordStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// UsageRecord represents a usage record for reconciliation comparison
type UsageRecord struct {
	// RecordID is the unique identifier for this usage record
	RecordID string `json:"record_id"`

	// LeaseID links to the marketplace lease
	LeaseID string `json:"lease_id"`

	// Provider is the service provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// StartTime is when the usage period started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the usage period ended
	EndTime time.Time `json:"end_time"`

	// ResourceType is the type of resource (CPU, Memory, Storage, etc.)
	ResourceType UsageType `json:"resource_type"`

	// UsageAmount is the quantity of usage
	UsageAmount sdkmath.LegacyDec `json:"usage_amount"`

	// UnitPrice is the price per unit
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// TotalAmount is the calculated total amount
	TotalAmount sdk.Coins `json:"total_amount"`

	// InvoiceID links to the invoice (may be empty if not yet invoiced)
	InvoiceID string `json:"invoice_id,omitempty"`

	// Status is the current status of this usage record
	Status UsageRecordStatus `json:"status"`

	// BlockHeight is when this record was created on-chain
	BlockHeight int64 `json:"block_height"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the usage record
func (r *UsageRecord) Validate() error {
	if r.RecordID == "" {
		return fmt.Errorf("record_id is required")
	}

	if len(r.RecordID) > 64 {
		return fmt.Errorf("record_id exceeds maximum length of 64")
	}

	if r.LeaseID == "" {
		return fmt.Errorf("lease_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if r.EndTime.Before(r.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	if r.UsageAmount.IsNegative() {
		return fmt.Errorf("usage_amount cannot be negative")
	}

	if !r.TotalAmount.IsValid() {
		return fmt.Errorf("total_amount must be valid coins")
	}

	return nil
}

// PayoutStatus defines the status of a payout
type PayoutStatus uint8

const (
	// PayoutStatusPending is a pending payout
	PayoutStatusPending PayoutStatus = 0

	// PayoutStatusProcessing is a payout being processed
	PayoutStatusProcessing PayoutStatus = 1

	// PayoutStatusCompleted is a completed payout
	PayoutStatusCompleted PayoutStatus = 2

	// PayoutStatusFailed is a failed payout
	PayoutStatusFailed PayoutStatus = 3

	// PayoutStatusCancelled is a cancelled payout
	PayoutStatusCancelled PayoutStatus = 4

	// PayoutStatusRefunded is a refunded payout
	PayoutStatusRefunded PayoutStatus = 5
)

// PayoutStatusNames maps status to names
var PayoutStatusNames = map[PayoutStatus]string{
	PayoutStatusPending:    "pending",
	PayoutStatusProcessing: "processing",
	PayoutStatusCompleted:  "completed",
	PayoutStatusFailed:     "failed",
	PayoutStatusCancelled:  "cancelled",
	PayoutStatusRefunded:   "refunded",
}

// String returns string representation
func (s PayoutStatus) String() string {
	if name, ok := PayoutStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// PayoutMethod defines the method of payout
type PayoutMethod uint8

const (
	// PayoutMethodCrypto is a crypto payout
	PayoutMethodCrypto PayoutMethod = 0

	// PayoutMethodFiatPayPal is a fiat payout via PayPal
	PayoutMethodFiatPayPal PayoutMethod = 1

	// PayoutMethodFiatACH is a fiat payout via ACH
	PayoutMethodFiatACH PayoutMethod = 2

	// PayoutMethodFiatWire is a fiat payout via wire transfer
	PayoutMethodFiatWire PayoutMethod = 3
)

// PayoutMethodNames maps methods to names
var PayoutMethodNames = map[PayoutMethod]string{
	PayoutMethodCrypto:     "crypto",
	PayoutMethodFiatPayPal: "fiat_paypal",
	PayoutMethodFiatACH:    "fiat_ach",
	PayoutMethodFiatWire:   "fiat_wire",
}

// String returns string representation
func (m PayoutMethod) String() string {
	if name, ok := PayoutMethodNames[m]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", m)
}

// PayoutRecord represents a payout record for reconciliation
type PayoutRecord struct {
	// PayoutID is the unique identifier for this payout
	PayoutID string `json:"payout_id"`

	// SettlementID links to the settlement (optional, for settlement-based payouts)
	SettlementID string `json:"settlement_id,omitempty"`

	// Provider is the provider address receiving the payout
	Provider string `json:"provider"`

	// GrossAmount is the total amount before fees (same as PayoutAmount for backwards compatibility)
	GrossAmount sdk.Coins `json:"gross_amount,omitempty"`

	// PayoutAmount is the gross payout amount
	PayoutAmount sdk.Coins `json:"payout_amount"`

	// FeeAmount is the fee deducted from the payout (alias: TotalFees)
	FeeAmount sdk.Coins `json:"fee_amount"`

	// TotalFees is the total fees deducted (same as FeeAmount)
	TotalFees sdk.Coins `json:"total_fees,omitempty"`

	// NetAmount is the net amount after fees
	NetAmount sdk.Coins `json:"net_amount"`

	// HoldbackAmount is any amount held back for disputes
	HoldbackAmount sdk.Coins `json:"holdback_amount,omitempty"`

	// InvoiceIDs are the invoices covered by this payout
	InvoiceIDs []string `json:"invoice_ids"`

	// Status is the current status of this payout
	Status PayoutStatus `json:"status"`

	// Method is the payout method
	Method PayoutMethod `json:"method"`

	// PayoutDate is when the payout was/will be made
	PayoutDate time.Time `json:"payout_date"`

	// TransactionRef is the external transaction reference
	TransactionRef string `json:"transaction_ref,omitempty"`

	// TransactionHash is the on-chain transaction hash
	TransactionHash string `json:"transaction_hash,omitempty"`

	// FailureReason is the reason for failure (if failed)
	FailureReason string `json:"failure_reason,omitempty"`

	// RefundAmount is the refund amount (if refunded)
	RefundAmount sdk.Coins `json:"refund_amount,omitempty"`

	// RefundReason is the reason for refund
	RefundReason string `json:"refund_reason,omitempty"`

	// RefundedAt is when the refund occurred
	RefundedAt *time.Time `json:"refunded_at,omitempty"`

	// CompletedAt is when the payout was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// BlockHeight is when this record was created on-chain
	BlockHeight int64 `json:"block_height"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewPayoutRecord creates a new payout record for settlement-based payouts
func NewPayoutRecord(
	payoutID string,
	settlementID string,
	provider string,
	grossAmount sdk.Coins,
	totalFees sdk.Coins,
	netAmount sdk.Coins,
	payoutAmount sdk.Coins,
	blockHeight int64,
	now time.Time,
) *PayoutRecord {
	return &PayoutRecord{
		PayoutID:     payoutID,
		SettlementID: settlementID,
		Provider:     provider,
		GrossAmount:  grossAmount,
		PayoutAmount: grossAmount, // For compatibility
		FeeAmount:    totalFees,
		TotalFees:    totalFees,
		NetAmount:    netAmount,
		Status:       PayoutStatusPending,
		Method:       PayoutMethodCrypto,
		PayoutDate:   now,
		BlockHeight:  blockHeight,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate validates the payout record
func (r *PayoutRecord) Validate() error {
	if r.PayoutID == "" {
		return fmt.Errorf("payout_id is required")
	}

	if len(r.PayoutID) > 64 {
		return fmt.Errorf("payout_id exceeds maximum length of 64")
	}

	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if !r.PayoutAmount.IsValid() {
		return fmt.Errorf("payout_amount must be valid coins")
	}

	if !r.FeeAmount.IsValid() {
		return fmt.Errorf("fee_amount must be valid coins")
	}

	if !r.NetAmount.IsValid() {
		return fmt.Errorf("net_amount must be valid coins")
	}

	// Either InvoiceIDs or SettlementID is required
	if len(r.InvoiceIDs) == 0 && r.SettlementID == "" {
		return fmt.Errorf("at least one invoice_id or settlement_id is required")
	}

	if r.PayoutDate.IsZero() {
		return fmt.Errorf("payout_date is required")
	}

	return nil
}

// PayoutCalculation represents the calculation for a payout
type PayoutCalculation struct {
	// SettlementID is the settlement being calculated
	SettlementID string `json:"settlement_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// GrossAmount is the total amount before fees
	GrossAmount sdk.Coins `json:"gross_amount"`

	// TotalFees is the total fees to deduct
	TotalFees sdk.Coins `json:"total_fees"`

	// NetAmount is the net amount after fees
	NetAmount sdk.Coins `json:"net_amount"`

	// HoldbackAmount is any amount to hold back
	HoldbackAmount sdk.Coins `json:"holdback_amount"`

	// PayableAmount is the amount to pay out
	PayableAmount sdk.Coins `json:"payable_amount"`

	// CalculatedAt is when the calculation was performed
	CalculatedAt time.Time `json:"calculated_at"`

	// BlockHeight is when the calculation was performed
	BlockHeight int64 `json:"block_height"`
}

// PayoutSummary provides a summary of payouts for a provider
type PayoutSummary struct {
	// Provider is the provider address
	Provider string `json:"provider"`

	// PeriodStart is the start of the summary period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the summary period
	PeriodEnd time.Time `json:"period_end"`

	// TotalPayouts is the number of payouts
	TotalPayouts uint32 `json:"total_payouts"`

	// TotalGrossAmount is the total gross amount
	TotalGrossAmount sdk.Coins `json:"total_gross_amount"`

	// TotalFees is the total fees deducted
	TotalFees sdk.Coins `json:"total_fees"`

	// TotalNetAmount is the total net amount
	TotalNetAmount sdk.Coins `json:"total_net_amount"`

	// TotalPayoutAmount is the total payout amount
	TotalPayoutAmount sdk.Coins `json:"total_payout_amount"`

	// PendingPayouts is the number of pending payouts
	PendingPayouts uint32 `json:"pending_payouts"`

	// PendingAmount is the total pending amount
	PendingAmount sdk.Coins `json:"pending_amount"`

	// CompletedPayouts is the number of completed payouts
	CompletedPayouts uint32 `json:"completed_payouts"`

	// CompletedAmount is the total completed amount
	CompletedAmount sdk.Coins `json:"completed_amount"`

	// FailedPayouts is the number of failed payouts
	FailedPayouts uint32 `json:"failed_payouts"`

	// FailedAmount is the total failed amount
	FailedAmount sdk.Coins `json:"failed_amount"`

	// GeneratedAt is when the summary was generated
	GeneratedAt time.Time `json:"generated_at"`

	// BlockHeight is when the summary was generated
	BlockHeight int64 `json:"block_height"`
}

// NewPayoutSummary creates a new payout summary
func NewPayoutSummary(provider string, periodStart, periodEnd time.Time, blockHeight int64, now time.Time) *PayoutSummary {
	return &PayoutSummary{
		Provider:          provider,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		TotalPayouts:      0,
		TotalGrossAmount:  sdk.NewCoins(),
		TotalFees:         sdk.NewCoins(),
		TotalNetAmount:    sdk.NewCoins(),
		TotalPayoutAmount: sdk.NewCoins(),
		PendingPayouts:    0,
		PendingAmount:     sdk.NewCoins(),
		CompletedPayouts:  0,
		CompletedAmount:   sdk.NewCoins(),
		FailedPayouts:     0,
		FailedAmount:      sdk.NewCoins(),
		GeneratedAt:       now,
		BlockHeight:       blockHeight,
	}
}

// AddPayout adds a payout to the summary
func (ps *PayoutSummary) AddPayout(payout *PayoutRecord) {
	ps.TotalPayouts++
	ps.TotalGrossAmount = ps.TotalGrossAmount.Add(payout.GrossAmount...)
	ps.TotalFees = ps.TotalFees.Add(payout.FeeAmount...)
	ps.TotalNetAmount = ps.TotalNetAmount.Add(payout.NetAmount...)
	ps.TotalPayoutAmount = ps.TotalPayoutAmount.Add(payout.PayoutAmount...)

	switch payout.Status {
	case PayoutStatusPending, PayoutStatusProcessing:
		ps.PendingPayouts++
		ps.PendingAmount = ps.PendingAmount.Add(payout.PayoutAmount...)
	case PayoutStatusCompleted:
		ps.CompletedPayouts++
		ps.CompletedAmount = ps.CompletedAmount.Add(payout.PayoutAmount...)
	case PayoutStatusFailed:
		ps.FailedPayouts++
		ps.FailedAmount = ps.FailedAmount.Add(payout.PayoutAmount...)
	}
}

// ReconciliationOption defines options for reconciliation operations
type ReconciliationOption func(*reconciliationOptions)

// reconciliationOptions holds internal reconciliation options
type reconciliationOptions struct {
	provider        string
	customer        string
	includeVoided   bool
	skipValidation  bool //nolint:unused // Reserved for optional validation bypass
	maxRecords      uint32
	notifyOnSuccess bool
}

// WithProvider filters reconciliation by provider
func WithProvider(provider string) ReconciliationOption {
	return func(opts *reconciliationOptions) {
		opts.provider = provider
	}
}

// WithCustomer filters reconciliation by customer
func WithCustomer(customer string) ReconciliationOption {
	return func(opts *reconciliationOptions) {
		opts.customer = customer
	}
}

// WithIncludeVoided includes voided records in reconciliation
func WithIncludeVoided(include bool) ReconciliationOption {
	return func(opts *reconciliationOptions) {
		opts.includeVoided = include
	}
}

// WithMaxRecords limits the number of records processed
func WithMaxRecords(max uint32) ReconciliationOption {
	return func(opts *reconciliationOptions) {
		opts.maxRecords = max
	}
}

// WithNotifyOnSuccess enables notifications on successful reconciliation
func WithNotifyOnSuccess(notify bool) ReconciliationOption {
	return func(opts *reconciliationOptions) {
		opts.notifyOnSuccess = notify
	}
}

// ReconciliationService defines the interface for reconciliation operations
type ReconciliationService interface {
	// GenerateReport generates a reconciliation report for the specified period
	GenerateReport(
		ctx context.Context,
		config ReconciliationConfig,
		periodStart time.Time,
		periodEnd time.Time,
		opts ...ReconciliationOption,
	) (*ReconciliationReport, error)

	// ReconcileUsageToInvoices reconciles usage records against invoices
	ReconcileUsageToInvoices(
		ctx context.Context,
		usageRecords []UsageRecord,
		invoices []*InvoiceLedgerRecord,
	) ([]ReconciliationDiscrepancy, error)

	// ReconcileInvoicesToPayouts reconciles invoices against payouts
	ReconcileInvoicesToPayouts(
		ctx context.Context,
		invoices []*InvoiceLedgerRecord,
		payouts []PayoutRecord,
	) ([]ReconciliationDiscrepancy, error)

	// ComputeVariance computes the variance between expected and actual amounts
	// Returns: difference coins, variance percentage, error
	ComputeVariance(expected, actual sdk.Coins) (sdk.Coins, sdkmath.LegacyDec, error)

	// AutoResolveDiscrepancy attempts to automatically resolve a discrepancy
	// Returns: resolved (bool), error
	AutoResolveDiscrepancy(
		ctx context.Context,
		discrepancy ReconciliationDiscrepancy,
	) (bool, error)

	// ScheduleNightlyReconciliation schedules a nightly reconciliation job
	ScheduleNightlyReconciliation(ctx context.Context, hour int) error
}

// ReconciliationJobConfig defines configuration for scheduled reconciliation jobs
type ReconciliationJobConfig struct {
	// JobID is the unique identifier for this job
	JobID string `json:"job_id"`

	// CronSchedule is the cron expression for scheduling (e.g., "0 2 * * *" for 2am daily)
	CronSchedule string `json:"cron_schedule"`

	// ReportType is the type of reconciliation report to generate
	ReportType ReconciliationReportType `json:"report_type"`

	// Config is the reconciliation configuration to use
	Config ReconciliationConfig `json:"config"`

	// NotifyOnFailure indicates whether to notify on failure
	NotifyOnFailure bool `json:"notify_on_failure"`

	// NotifyOnDiscrepancy indicates whether to notify when discrepancies are found
	NotifyOnDiscrepancy bool `json:"notify_on_discrepancy"`

	// NotificationRecipients are the addresses to notify
	NotificationRecipients []string `json:"notification_recipients"`

	// IsEnabled indicates whether this job is enabled
	IsEnabled bool `json:"is_enabled"`

	// LastRunAt is when the job was last run
	LastRunAt *time.Time `json:"last_run_at,omitempty"`

	// LastRunStatus is the status of the last run
	LastRunStatus ReconciliationStatus `json:"last_run_status,omitempty"`

	// CreatedAt is when the job was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the job was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the reconciliation job configuration
func (c *ReconciliationJobConfig) Validate() error {
	if c.JobID == "" {
		return fmt.Errorf("job_id is required")
	}

	if len(c.JobID) > 64 {
		return fmt.Errorf("job_id exceeds maximum length of 64")
	}

	if c.CronSchedule == "" {
		return fmt.Errorf("cron_schedule is required")
	}

	if err := c.Config.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// Validate notification recipients are valid addresses
	for i, recipient := range c.NotificationRecipients {
		if _, err := sdk.AccAddressFromBech32(recipient); err != nil {
			return fmt.Errorf("invalid notification_recipient[%d]: %w", i, err)
		}
	}

	return nil
}

// NewReconciliationJob creates a new reconciliation job configuration
func NewReconciliationJob(
	jobID string,
	cronSchedule string,
	reportType ReconciliationReportType,
	config ReconciliationConfig,
	notifyOnFailure bool,
	notifyOnDiscrepancy bool,
	recipients []string,
) *ReconciliationJobConfig {
	now := time.Now().UTC()
	return &ReconciliationJobConfig{
		JobID:                  jobID,
		CronSchedule:           cronSchedule,
		ReportType:             reportType,
		Config:                 config,
		NotifyOnFailure:        notifyOnFailure,
		NotifyOnDiscrepancy:    notifyOnDiscrepancy,
		NotificationRecipients: recipients,
		IsEnabled:              true,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// NewNightlyReconciliationJob creates a new nightly reconciliation job
func NewNightlyReconciliationJob(jobID string, hour int, recipients []string) *ReconciliationJobConfig {
	cronSchedule := fmt.Sprintf("0 %d * * *", hour)
	return NewReconciliationJob(
		jobID,
		cronSchedule,
		ReconciliationReportTypeDaily,
		DefaultReconciliationConfig(),
		true,
		true,
		recipients,
	)
}

// CalculateVariancePercentage calculates the variance percentage between expected and actual coin
// Returns the absolute variance as a percentage (0.01 = 1%)
func CalculateVariancePercentage(expected, actual sdk.Coin) sdkmath.LegacyDec {
	if expected.IsZero() && actual.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	if expected.IsZero() {
		// If expected is zero but actual is not, variance is 100%
		return sdkmath.LegacyOneDec()
	}

	if expected.Denom != actual.Denom {
		// Different denominations cannot be compared
		return sdkmath.LegacyOneDec()
	}

	expectedDec := sdkmath.LegacyNewDecFromInt(expected.Amount)
	actualDec := sdkmath.LegacyNewDecFromInt(actual.Amount)

	// Calculate absolute difference
	diff := expectedDec.Sub(actualDec)
	if diff.IsNegative() {
		diff = diff.Neg()
	}

	// Calculate variance percentage
	variance := diff.Quo(expectedDec)

	return variance
}

// CalculateTotalVariance calculates the total variance percentage for multiple coins
func CalculateTotalVariance(expected, actual sdk.Coins) sdkmath.LegacyDec {
	if expected.IsZero() && actual.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	if expected.IsZero() {
		return sdkmath.LegacyOneDec()
	}

	totalExpected := sdkmath.LegacyZeroDec()
	totalDiff := sdkmath.LegacyZeroDec()

	// Process all expected coins
	for _, exp := range expected {
		expDec := sdkmath.LegacyNewDecFromInt(exp.Amount)
		totalExpected = totalExpected.Add(expDec)

		act := actual.AmountOf(exp.Denom)
		actDec := sdkmath.LegacyNewDecFromInt(act)

		diff := expDec.Sub(actDec)
		if diff.IsNegative() {
			diff = diff.Neg()
		}
		totalDiff = totalDiff.Add(diff)
	}

	// Process any actual coins not in expected
	for _, act := range actual {
		if expected.AmountOf(act.Denom).IsZero() {
			actDec := sdkmath.LegacyNewDecFromInt(act.Amount)
			totalDiff = totalDiff.Add(actDec)
		}
	}

	if totalExpected.IsZero() {
		return sdkmath.LegacyOneDec()
	}

	return totalDiff.Quo(totalExpected)
}

// ComputeCoinDifference computes the difference between expected and actual coins
func ComputeCoinDifference(expected, actual sdk.Coins) sdk.Coins {
	diff := sdk.NewCoins()

	// Calculate difference for each expected coin
	for _, exp := range expected {
		act := actual.AmountOf(exp.Denom)
		coinDiff := exp.Amount.Sub(act)
		if !coinDiff.IsZero() {
			if coinDiff.IsNegative() {
				diff = diff.Add(sdk.NewCoin(exp.Denom, coinDiff.Neg()))
			} else {
				diff = diff.Add(sdk.NewCoin(exp.Denom, coinDiff))
			}
		}
	}

	// Add any actual coins not in expected
	for _, act := range actual {
		if expected.AmountOf(act.Denom).IsZero() {
			diff = diff.Add(act)
		}
	}

	return diff
}

// Store key prefixes for reconciliation service types
var (
	// UsageRecordPrefix is the prefix for usage records
	UsageRecordPrefix = []byte{0x75}

	// UsageRecordByLeasePrefix indexes usage records by lease
	UsageRecordByLeasePrefix = []byte{0x76}

	// PayoutRecordPrefix is the prefix for payout records
	PayoutRecordPrefix = []byte{0x77}

	// PayoutRecordByProviderPrefix indexes payout records by provider
	PayoutRecordByProviderPrefix = []byte{0x78}

	// ReconciliationJobPrefix is the prefix for reconciliation jobs
	ReconciliationJobPrefix = []byte{0x79}

	// PayoutRecordByStatusPrefix indexes payout records by status
	PayoutRecordByStatusPrefix = []byte{0x7a}

	// PayoutRecordBySettlementPrefix indexes payout records by settlement
	PayoutRecordBySettlementPrefix = []byte{0x7b}

	// PayoutSequenceKey is the key for payout sequence
	PayoutSequenceKey = []byte("payout_sequence")
)

// BuildUsageRecordKey builds the key for a usage record
func BuildUsageRecordKey(recordID string) []byte {
	return append(UsageRecordPrefix, []byte(recordID)...)
}

// ParseUsageRecordKey parses a usage record key
func ParseUsageRecordKey(key []byte) (string, error) {
	if len(key) <= len(UsageRecordPrefix) {
		return "", fmt.Errorf("invalid usage record key length")
	}
	return string(key[len(UsageRecordPrefix):]), nil
}

// BuildUsageRecordByLeaseKey builds the index key for usage records by lease
func BuildUsageRecordByLeaseKey(leaseID string, recordID string) []byte {
	key := make([]byte, 0, len(UsageRecordByLeasePrefix)+len(leaseID)+1+len(recordID))
	key = append(key, UsageRecordByLeasePrefix...)
	key = append(key, []byte(leaseID)...)
	key = append(key, byte('/'))
	return append(key, []byte(recordID)...)
}

// BuildUsageRecordByLeasePrefix builds the prefix for lease's usage records
func BuildUsageRecordByLeasePrefix(leaseID string) []byte {
	key := make([]byte, 0, len(UsageRecordByLeasePrefix)+len(leaseID)+1)
	key = append(key, UsageRecordByLeasePrefix...)
	key = append(key, []byte(leaseID)...)
	return append(key, byte('/'))
}

// BuildPayoutRecordKey builds the key for a payout record
func BuildPayoutRecordKey(payoutID string) []byte {
	key := make([]byte, 0, len(PayoutRecordPrefix)+len(payoutID))
	key = append(key, PayoutRecordPrefix...)
	return append(key, []byte(payoutID)...)
}

// ParsePayoutRecordKey parses a payout record key
func ParsePayoutRecordKey(key []byte) (string, error) {
	if len(key) <= len(PayoutRecordPrefix) {
		return "", fmt.Errorf("invalid payout record key length")
	}
	return string(key[len(PayoutRecordPrefix):]), nil
}

// BuildPayoutRecordByProviderKey builds the index key for payout records by provider
func BuildPayoutRecordByProviderKey(provider string, payoutID string) []byte {
	key := make([]byte, 0, len(PayoutRecordByProviderPrefix)+len(provider)+1+len(payoutID))
	key = append(key, PayoutRecordByProviderPrefix...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return append(key, []byte(payoutID)...)
}

// BuildPayoutRecordByProviderPrefix builds the prefix for provider's payout records
func BuildPayoutRecordByProviderPrefix(provider string) []byte {
	key := make([]byte, 0, len(PayoutRecordByProviderPrefix)+len(provider)+1)
	key = append(key, PayoutRecordByProviderPrefix...)
	key = append(key, []byte(provider)...)
	return append(key, byte('/'))
}

// BuildPayoutRecordByDateKey builds the index key for payout records by date
func BuildPayoutRecordByDateKey(timestamp int64, payoutID string) []byte {
	key := make([]byte, 0, len(PayoutRecordPrefix)+1+8+1+len(payoutID))
	key = append(key, PayoutRecordPrefix...)
	key = append(key, byte('/'))
	// Append timestamp as big-endian uint64 for proper ordering
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	key = append(key, tsBytes...)
	key = append(key, byte('/'))
	return append(key, []byte(payoutID)...)
}

// BuildReconciliationJobKey builds the key for a reconciliation job
func BuildReconciliationJobKey(jobID string) []byte {
	key := make([]byte, 0, len(ReconciliationJobPrefix)+len(jobID))
	key = append(key, ReconciliationJobPrefix...)
	return append(key, []byte(jobID)...)
}

// ParseReconciliationJobKey parses a reconciliation job key
func ParseReconciliationJobKey(key []byte) (string, error) {
	if len(key) <= len(ReconciliationJobPrefix) {
		return "", fmt.Errorf("invalid reconciliation job key length")
	}
	return string(key[len(ReconciliationJobPrefix):]), nil
}

// BuildPayoutRecordByStatusKey builds the index key for payout records by status
func BuildPayoutRecordByStatusKey(status PayoutStatus, payoutID string) []byte {
	key := make([]byte, 0, len(PayoutRecordByStatusPrefix)+2+len(payoutID))
	key = append(key, PayoutRecordByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(payoutID)...)
}

// BuildPayoutRecordByStatusPrefix builds the prefix for payouts by status
func BuildPayoutRecordByStatusPrefix(status PayoutStatus) []byte {
	key := make([]byte, 0, len(PayoutRecordByStatusPrefix)+2)
	key = append(key, PayoutRecordByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildPayoutRecordBySettlementKey builds the index key for payout records by settlement
func BuildPayoutRecordBySettlementKey(settlementID string, payoutID string) []byte {
	key := make([]byte, 0, len(PayoutRecordBySettlementPrefix)+len(settlementID)+1+len(payoutID))
	key = append(key, PayoutRecordBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	key = append(key, byte('/'))
	return append(key, []byte(payoutID)...)
}

// BuildPayoutRecordBySettlementPrefix builds the prefix for settlement's payout records
func BuildPayoutRecordBySettlementPrefix(settlementID string) []byte {
	key := make([]byte, 0, len(PayoutRecordBySettlementPrefix)+len(settlementID)+1)
	key = append(key, PayoutRecordBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	return append(key, byte('/'))
}

// NextPayoutID generates the next payout ID
func NextPayoutID(currentSequence uint64, prefix string) string {
	return fmt.Sprintf("%s-PAY-%08d", prefix, currentSequence+1)
}

// IsVarianceWithinThreshold checks if the variance is within the acceptable threshold
func IsVarianceWithinThreshold(variance sdkmath.LegacyDec, threshold sdkmath.LegacyDec) bool {
	return variance.LTE(threshold)
}

// IsAutoResolvable checks if a discrepancy can be automatically resolved
func IsAutoResolvable(discrepancy ReconciliationDiscrepancy, config ReconciliationConfig) bool {
	// Cannot auto-resolve critical discrepancies
	if discrepancy.Severity == DiscrepancySeverityCritical {
		return false
	}

	// Cannot auto-resolve if amount exceeds manual review threshold
	if discrepancy.Difference.IsAllGTE(config.RequireManualReviewAbove) {
		return false
	}

	// Check variance against auto-resolve threshold
	variance := CalculateTotalVariance(discrepancy.ExpectedAmount, discrepancy.ActualAmount)
	return variance.LTE(config.AutoResolveThreshold)
}

// CreateDiscrepancyFromComparison creates a discrepancy from comparing expected and actual amounts
func CreateDiscrepancyFromComparison(
	discrepancyID string,
	discrepancyType DiscrepancyType,
	invoiceID string,
	usageRecordID string,
	expected sdk.Coins,
	actual sdk.Coins,
	description string,
) ReconciliationDiscrepancy {
	diff := ComputeCoinDifference(expected, actual)
	variance := CalculateTotalVariance(expected, actual)

	// Determine severity based on variance
	var severity DiscrepancySeverity
	switch {
	case variance.LTE(sdkmath.LegacyNewDecWithPrec(1, 3)): // <= 0.1%
		severity = DiscrepancySeverityLow
	case variance.LTE(sdkmath.LegacyNewDecWithPrec(1, 2)): // <= 1%
		severity = DiscrepancySeverityMedium
	case variance.LTE(sdkmath.LegacyNewDecWithPrec(5, 2)): // <= 5%
		severity = DiscrepancySeverityHigh
	default:
		severity = DiscrepancySeverityCritical
	}

	return ReconciliationDiscrepancy{
		DiscrepancyID:  discrepancyID,
		Type:           discrepancyType,
		InvoiceID:      invoiceID,
		UsageRecordID:  usageRecordID,
		ExpectedAmount: expected,
		ActualAmount:   actual,
		Difference:     diff,
		Description:    description,
		Severity:       severity,
	}
}
