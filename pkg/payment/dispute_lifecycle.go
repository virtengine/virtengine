// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// ErrDisputeInvalidTransition is returned when a status transition is not allowed
var ErrDisputeInvalidTransition = errors.New("invalid dispute status transition")

// DisputeLifecycle defines the allowed dispute status transitions.
//
// Resolution workflows (high-level):
// - Defend (submit evidence) -> needs_response -> under_review -> won/lost
// - Accept (concede) -> accepted (terminal)
// - Expire (missed deadline) -> expired (terminal)
//
// These workflows are codified by the transition map below and are exposed via
// DefaultDisputeResolutionWorkflows for finance and ops documentation.
type DisputeLifecycle struct {
	allowed map[DisputeStatus]map[DisputeStatus]bool
}

// DefaultDisputeLifecycle returns the default dispute state machine.
func DefaultDisputeLifecycle() *DisputeLifecycle {
	allowed := map[DisputeStatus]map[DisputeStatus]bool{
		DisputeStatusOpen: {
			DisputeStatusOpen:          true,
			DisputeStatusNeedsResponse: true,
			DisputeStatusUnderReview:   true,
			DisputeStatusWon:           true,
			DisputeStatusLost:          true,
			DisputeStatusAccepted:      true,
			DisputeStatusExpired:       true,
		},
		DisputeStatusNeedsResponse: {
			DisputeStatusNeedsResponse: true,
			DisputeStatusUnderReview:   true,
			DisputeStatusWon:           true,
			DisputeStatusLost:          true,
			DisputeStatusAccepted:      true,
			DisputeStatusExpired:       true,
		},
		DisputeStatusUnderReview: {
			DisputeStatusUnderReview: true,
			DisputeStatusWon:         true,
			DisputeStatusLost:        true,
			DisputeStatusAccepted:    true,
			DisputeStatusExpired:     true,
		},
		DisputeStatusWon: {
			DisputeStatusWon: true,
		},
		DisputeStatusLost: {
			DisputeStatusLost: true,
		},
		DisputeStatusAccepted: {
			DisputeStatusAccepted: true,
		},
		DisputeStatusExpired: {
			DisputeStatusExpired: true,
		},
	}

	return &DisputeLifecycle{allowed: allowed}
}

// CanTransition returns true if a transition is allowed.
func (l *DisputeLifecycle) CanTransition(from, to DisputeStatus) bool {
	if l == nil {
		return false
	}
	if to == "" {
		return false
	}
	if from == "" {
		return true
	}
	if from == to {
		return true
	}
	if next, ok := l.allowed[from]; ok {
		return next[to]
	}
	return false
}

// Transition updates the record status if the transition is valid.
func (l *DisputeLifecycle) Transition(record *DisputeRecord, newStatus DisputeStatus, actor string, details map[string]string) error {
	if record == nil {
		return fmt.Errorf("nil dispute record")
	}
	if !l.CanTransition(record.Status, newStatus) {
		return ErrDisputeInvalidTransition
	}

	if record.Status != newStatus {
		record.UpdateStatus(newStatus, actor, details)
	}

	return nil
}

// DisputeResolutionAction represents the primary action taken to resolve a dispute.
type DisputeResolutionAction string

const (
	DisputeResolutionSubmitEvidence DisputeResolutionAction = "submit_evidence"
	DisputeResolutionAccept         DisputeResolutionAction = "accept"
	DisputeResolutionMonitor        DisputeResolutionAction = "monitor"
)

// DisputeResolutionRequest contains information for a resolution workflow.
type DisputeResolutionRequest struct {
	DisputeID string
	Action    DisputeResolutionAction
	Evidence  DisputeEvidence
	Actor     string
	Notes     string
}

// DisputeResolutionResult summarizes a resolution workflow execution.
type DisputeResolutionResult struct {
	Record  *DisputeRecord
	Message string
}

// ResolveDispute executes the requested workflow and updates the dispute store.
func ResolveDispute(ctx context.Context, handler DisputeHandler, store DisputeStore, lifecycle *DisputeLifecycle, req DisputeResolutionRequest) (*DisputeResolutionResult, error) {
	if handler == nil || store == nil {
		return nil, fmt.Errorf("dispute handler and store are required")
	}
	if req.DisputeID == "" {
		return nil, fmt.Errorf("dispute ID is required")
	}
	if lifecycle == nil {
		lifecycle = DefaultDisputeLifecycle()
	}

	record, found := store.Get(req.DisputeID)
	if !found {
		disp, err := handler.GetDispute(ctx, req.DisputeID)
		if err != nil {
			return nil, err
		}
		record = NewDisputeRecord(disp, disp.Gateway, "system:resolution")
		if err := store.Save(record); err != nil {
			return nil, err
		}
	}

	actor := req.Actor
	if actor == "" {
		actor = "system:resolution"
	}

	switch req.Action {
	case DisputeResolutionSubmitEvidence:
		if err := handler.SubmitEvidence(ctx, req.DisputeID, req.Evidence); err != nil {
			return nil, err
		}
		record.MarkEvidenceSubmitted(actor, "resolution")
		if err := lifecycle.Transition(record, record.Status, actor, map[string]string{"notes": req.Notes}); err != nil {
			return nil, err
		}
	case DisputeResolutionAccept:
		if err := handler.AcceptDispute(ctx, req.DisputeID); err != nil {
			return nil, err
		}
		record.MarkAccepted(actor, req.Notes)
	case DisputeResolutionMonitor:
		record.AddAuditEntry(DisputeAuditEntry{
			Timestamp: time.Now(),
			Action:    "monitor",
			Actor:     actor,
			Details: map[string]string{
				"notes": req.Notes,
			},
		})
	default:
		return nil, fmt.Errorf("unsupported resolution action: %s", req.Action)
	}

	if err := store.Save(record); err != nil {
		return nil, err
	}

	return &DisputeResolutionResult{
		Record:  record,
		Message: "resolution workflow applied",
	}, nil
}

// DisputeResolutionWorkflow documents the recommended steps for each outcome.
type DisputeResolutionWorkflow struct {
	Outcome      DisputeStatus
	Description  string
	Steps        []string
	OwnerPrimary string
	SLATarget    string
}

// DefaultDisputeResolutionWorkflows returns workflow documentation for finance teams.
func DefaultDisputeResolutionWorkflows() []DisputeResolutionWorkflow {
	return []DisputeResolutionWorkflow{
		{
			Outcome:      DisputeStatusNeedsResponse,
			Description:  "Active disputes requiring evidence submission",
			OwnerPrimary: "finance-ops",
			SLATarget:    "Submit evidence within gateway deadline",
			Steps: []string{
				"Collect order metadata and delivery confirmation",
				"Compile customer communication records",
				"Submit evidence through gateway within deadline",
				"Monitor webhook updates for under_review status",
			},
		},
		{
			Outcome:      DisputeStatusUnderReview,
			Description:  "Evidence submitted and awaiting card network decision",
			OwnerPrimary: "finance-ops",
			SLATarget:    "Weekly follow-up until decision",
			Steps: []string{
				"Verify evidence submission confirmation",
				"Track expected resolution date",
				"Communicate status to accounting",
			},
		},
		{
			Outcome:      DisputeStatusWon,
			Description:  "Chargeback reversed in merchant favor",
			OwnerPrimary: "finance",
			SLATarget:    "Reconcile funds within 2 business days",
			Steps: []string{
				"Confirm funds reinstatement in gateway",
				"Update internal ledger and revenue reporting",
				"Close dispute record and archive evidence",
			},
		},
		{
			Outcome:      DisputeStatusLost,
			Description:  "Chargeback upheld in cardholder favor",
			OwnerPrimary: "finance",
			SLATarget:    "Record loss within 1 business day",
			Steps: []string{
				"Confirm loss and fee impact",
				"Update risk controls for root cause",
				"Notify revenue recognition team",
			},
		},
		{
			Outcome:      DisputeStatusAccepted,
			Description:  "Merchant conceded dispute",
			OwnerPrimary: "finance",
			SLATarget:    "Record concession immediately",
			Steps: []string{
				"Confirm acceptance action in gateway",
				"Update dispute ledger and write-off entry",
			},
		},
		{
			Outcome:      DisputeStatusExpired,
			Description:  "Response window missed",
			OwnerPrimary: "finance-ops",
			SLATarget:    "Escalate within 1 business day",
			Steps: []string{
				"Log missed deadline in incident tracker",
				"Update dispute ledger to expired",
				"Review process gaps and remediate",
			},
		},
	}
}

// DisputeReportOptions configures dispute reporting.
type DisputeReportOptions struct {
	ListOptions    DisputeListOptions
	IncludeRecords bool
	AsOf           time.Time
	AgingBuckets   []DisputeAgingBucketDefinition
}

// DisputeAgingBucketDefinition defines an aging bucket for open disputes.
type DisputeAgingBucketDefinition struct {
	Label string
	Min   time.Duration
	Max   time.Duration
}

// DisputeAgingBucket summarizes open disputes by age.
type DisputeAgingBucket struct {
	Label            string
	Count            int
	AmountByCurrency map[Currency]int64
}

// DisputeReport summarizes dispute activity for finance teams.
type DisputeReport struct {
	GeneratedAt            time.Time
	WindowStart            *time.Time
	WindowEnd              *time.Time
	TotalDisputes          int
	OpenDisputes           int
	ClosedDisputes         int
	PastDueDisputes        int
	EvidenceSubmittedCount int
	EvidenceSubmissionRate float64
	AverageResolutionHours float64
	TotalAmountByCurrency  map[Currency]int64
	OpenAmountByCurrency   map[Currency]int64
	ClosedAmountByCurrency map[Currency]int64
	StatusCounts           map[DisputeStatus]int
	GatewayCounts          map[GatewayType]int
	ReasonCounts           map[DisputeReason]int
	AgingBuckets           []DisputeAgingBucket
	Records                []*DisputeRecord
}

// DefaultDisputeAgingBuckets provides standard aging buckets for reporting.
func DefaultDisputeAgingBuckets() []DisputeAgingBucketDefinition {
	return []DisputeAgingBucketDefinition{
		{Label: "0-7d", Min: 0, Max: 7 * 24 * time.Hour},
		{Label: "8-14d", Min: 7 * 24 * time.Hour, Max: 14 * 24 * time.Hour},
		{Label: "15-30d", Min: 14 * 24 * time.Hour, Max: 30 * 24 * time.Hour},
		{Label: "31d+", Min: 30 * 24 * time.Hour, Max: 0},
	}
}

// GenerateDisputeReport aggregates dispute metrics for finance reporting.
func GenerateDisputeReport(store DisputeStore, opts DisputeReportOptions) (*DisputeReport, error) {
	if store == nil {
		return nil, fmt.Errorf("dispute store is required")
	}

	asOf := opts.AsOf
	if asOf.IsZero() {
		asOf = time.Now()
	}

	records := store.List(opts.ListOptions)

	report := &DisputeReport{
		GeneratedAt:            asOf,
		WindowStart:            opts.ListOptions.CreatedAfter,
		WindowEnd:              opts.ListOptions.CreatedBefore,
		TotalAmountByCurrency:  make(map[Currency]int64),
		OpenAmountByCurrency:   make(map[Currency]int64),
		ClosedAmountByCurrency: make(map[Currency]int64),
		StatusCounts:           make(map[DisputeStatus]int),
		GatewayCounts:          make(map[GatewayType]int),
		ReasonCounts:           make(map[DisputeReason]int),
	}

	agingDefs := opts.AgingBuckets
	if agingDefs == nil {
		agingDefs = DefaultDisputeAgingBuckets()
	}
	report.AgingBuckets = buildAgingBuckets(agingDefs)

	var totalResolutionHours float64
	var resolutionCount int

	for _, record := range records {
		report.TotalDisputes++
		report.StatusCounts[record.Status]++
		report.GatewayCounts[record.Gateway]++
		report.ReasonCounts[record.Reason]++

		amount := record.Amount.Value
		currency := record.Amount.Currency

		report.TotalAmountByCurrency[currency] += amount

		if record.Status.IsFinal() {
			report.ClosedDisputes++
			report.ClosedAmountByCurrency[currency] += amount
		} else {
			report.OpenDisputes++
			report.OpenAmountByCurrency[currency] += amount
			if record.Status == DisputeStatusNeedsResponse && !record.EvidenceDueBy.IsZero() && record.EvidenceDueBy.Before(asOf) {
				report.PastDueDisputes++
			}
			updateAgingBuckets(report.AgingBuckets, agingDefs, record.CreatedAt, asOf, record.Amount)
		}

		if record.EvidenceSubmitted {
			report.EvidenceSubmittedCount++
		}

		if record.ClosedAt != nil && !record.CreatedAt.IsZero() {
			resolutionHours := record.ClosedAt.Sub(record.CreatedAt).Hours()
			if resolutionHours >= 0 {
				totalResolutionHours += resolutionHours
				resolutionCount++
			}
		}
	}

	if report.TotalDisputes > 0 {
		report.EvidenceSubmissionRate = float64(report.EvidenceSubmittedCount) / float64(report.TotalDisputes)
	}
	if resolutionCount > 0 {
		report.AverageResolutionHours = totalResolutionHours / float64(resolutionCount)
	}

	if opts.IncludeRecords {
		report.Records = records
	}

	return report, nil
}

func buildAgingBuckets(defs []DisputeAgingBucketDefinition) []DisputeAgingBucket {
	buckets := make([]DisputeAgingBucket, len(defs))
	for i := range defs {
		buckets[i] = DisputeAgingBucket{
			Label:            defs[i].Label,
			AmountByCurrency: make(map[Currency]int64),
		}
	}
	return buckets
}

func updateAgingBuckets(buckets []DisputeAgingBucket, defs []DisputeAgingBucketDefinition, createdAt time.Time, asOf time.Time, amount Amount) {
	if createdAt.IsZero() || len(buckets) == 0 {
		return
	}

	age := asOf.Sub(createdAt)
	if age < 0 {
		age = 0
	}

	for i := range buckets {
		if i >= len(defs) {
			return
		}
		def := defs[i]
		if def.Max == 0 {
			if age >= def.Min {
				buckets[i].Count++
				buckets[i].AmountByCurrency[amount.Currency] += amount.Value
				return
			}
			continue
		}
		if age >= def.Min && age < def.Max {
			buckets[i].Count++
			buckets[i].AmountByCurrency[amount.Currency] += amount.Value
			return
		}
	}
}

// SortedDisputeStatuses returns dispute statuses in stable order for reporting.
func SortedDisputeStatuses() []DisputeStatus {
	statuses := []DisputeStatus{
		DisputeStatusOpen,
		DisputeStatusNeedsResponse,
		DisputeStatusUnderReview,
		DisputeStatusWon,
		DisputeStatusLost,
		DisputeStatusAccepted,
		DisputeStatusExpired,
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i] < statuses[j] })
	return statuses
}
