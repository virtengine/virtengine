// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InvoiceStatusTransition defines a valid status transition
type InvoiceStatusTransition struct {
	From        InvoiceStatus
	To          InvoiceStatus
	RequiresPay bool   // true if this transition requires payment
	Description string // description of the transition
}

// ValidInvoiceTransitions defines all valid invoice status transitions
var ValidInvoiceTransitions = []InvoiceStatusTransition{
	{From: InvoiceStatusDraft, To: InvoiceStatusPending, Description: "issue invoice"},
	{From: InvoiceStatusDraft, To: InvoiceStatusCancelled, Description: "cancel draft"},
	{From: InvoiceStatusPending, To: InvoiceStatusPaid, RequiresPay: true, Description: "full payment"},
	{From: InvoiceStatusPending, To: InvoiceStatusPartiallyPaid, RequiresPay: true, Description: "partial payment"},
	{From: InvoiceStatusPending, To: InvoiceStatusOverdue, Description: "mark overdue"},
	{From: InvoiceStatusPending, To: InvoiceStatusDisputed, Description: "dispute invoice"},
	{From: InvoiceStatusPending, To: InvoiceStatusCancelled, Description: "cancel pending"},
	{From: InvoiceStatusPartiallyPaid, To: InvoiceStatusPaid, RequiresPay: true, Description: "complete payment"},
	{From: InvoiceStatusPartiallyPaid, To: InvoiceStatusOverdue, Description: "mark overdue"},
	{From: InvoiceStatusPartiallyPaid, To: InvoiceStatusDisputed, Description: "dispute invoice"},
	{From: InvoiceStatusOverdue, To: InvoiceStatusPaid, RequiresPay: true, Description: "pay overdue"},
	{From: InvoiceStatusOverdue, To: InvoiceStatusPartiallyPaid, RequiresPay: true, Description: "partial pay overdue"},
	{From: InvoiceStatusOverdue, To: InvoiceStatusDisputed, Description: "dispute overdue"},
	{From: InvoiceStatusOverdue, To: InvoiceStatusCancelled, Description: "write off"},
	{From: InvoiceStatusDisputed, To: InvoiceStatusPending, Description: "resolve dispute - pending"},
	{From: InvoiceStatusDisputed, To: InvoiceStatusPaid, Description: "resolve dispute - paid"},
	{From: InvoiceStatusDisputed, To: InvoiceStatusCancelled, Description: "resolve dispute - cancel"},
	{From: InvoiceStatusDisputed, To: InvoiceStatusRefunded, Description: "resolve dispute - refund"},
	{From: InvoiceStatusPaid, To: InvoiceStatusRefunded, Description: "refund paid"},
}

// validTransitionsMap is a pre-computed map for fast lookup
var validTransitionsMap map[InvoiceStatus]map[InvoiceStatus]InvoiceStatusTransition

func init() {
	validTransitionsMap = make(map[InvoiceStatus]map[InvoiceStatus]InvoiceStatusTransition)
	for _, t := range ValidInvoiceTransitions {
		if validTransitionsMap[t.From] == nil {
			validTransitionsMap[t.From] = make(map[InvoiceStatus]InvoiceStatusTransition)
		}
		validTransitionsMap[t.From][t.To] = t
	}
}

// IsValidTransition checks if a status transition is valid
func IsValidTransition(from, to InvoiceStatus) bool {
	if fromMap, ok := validTransitionsMap[from]; ok {
		_, valid := fromMap[to]
		return valid
	}
	return false
}

// GetTransition returns the transition details if valid
func GetTransition(from, to InvoiceStatus) (InvoiceStatusTransition, bool) {
	if fromMap, ok := validTransitionsMap[from]; ok {
		t, valid := fromMap[to]
		return t, valid
	}
	return InvoiceStatusTransition{}, false
}

// GetValidNextStates returns all valid next states from current state
func GetValidNextStates(current InvoiceStatus) []InvoiceStatus {
	var states []InvoiceStatus
	if fromMap, ok := validTransitionsMap[current]; ok {
		for to := range fromMap {
			states = append(states, to)
		}
	}
	return states
}

// InvoiceStatusMachine manages invoice status transitions
type InvoiceStatusMachine struct {
	invoice     *Invoice
	ledgerEntry *InvoiceLedgerEntry
	now         time.Time
	blockHeight int64
}

// NewInvoiceStatusMachine creates a new status machine for an invoice
func NewInvoiceStatusMachine(inv *Invoice, blockHeight int64, now time.Time) *InvoiceStatusMachine {
	return &InvoiceStatusMachine{
		invoice:     inv,
		now:         now,
		blockHeight: blockHeight,
	}
}

// Transition attempts to transition to a new status
func (sm *InvoiceStatusMachine) Transition(
	to InvoiceStatus,
	initiator string,
	description string,
) error {
	from := sm.invoice.Status

	transition, valid := GetTransition(from, to)
	if !valid {
		return fmt.Errorf("invalid transition from %s to %s", from, to)
	}

	// Create ledger entry
	entryType := sm.getEntryTypeForTransition(to)
	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		entryType,
		from,
		to,
		sdk.NewCoins(),
		fmt.Sprintf("%s: %s", transition.Description, description),
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	// Update invoice status
	sm.invoice.Status = to
	sm.invoice.UpdatedAt = sm.now

	return nil
}

// TransitionWithPayment attempts a payment-related transition
func (sm *InvoiceStatusMachine) TransitionWithPayment(
	amount sdk.Coins,
	initiator string,
) error {
	from := sm.invoice.Status

	// Record payment on invoice
	if err := sm.invoice.RecordPayment(amount, sm.now); err != nil {
		return err
	}

	to := sm.invoice.Status

	transition, valid := GetTransition(from, to)
	if !valid {
		return fmt.Errorf("invalid transition from %s to %s after payment", from, to)
	}

	if !transition.RequiresPay {
		return fmt.Errorf("transition from %s to %s does not require payment", from, to)
	}

	// Create ledger entry
	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-pay-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		LedgerEntryTypePayment,
		from,
		to,
		amount,
		fmt.Sprintf("payment recorded: %s", amount.String()),
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	return nil
}

// MarkIssued transitions from draft to issued/pending
func (sm *InvoiceStatusMachine) MarkIssued(initiator string) error {
	if sm.invoice.Status != InvoiceStatusDraft {
		return fmt.Errorf("can only issue draft invoices, current status: %s", sm.invoice.Status)
	}

	if err := sm.invoice.Finalize(sm.now); err != nil {
		return err
	}

	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-issued-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		LedgerEntryTypeIssued,
		InvoiceStatusDraft,
		InvoiceStatusPending,
		sdk.NewCoins(),
		"invoice issued",
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	return nil
}

// MarkDisputed transitions to disputed status
func (sm *InvoiceStatusMachine) MarkDisputed(
	disputeWindow *DisputeWindow,
	initiator string,
	reason string,
) error {
	from := sm.invoice.Status

	if err := sm.invoice.MarkDisputed(disputeWindow, sm.now); err != nil {
		return err
	}

	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-disputed-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		LedgerEntryTypeDisputed,
		from,
		InvoiceStatusDisputed,
		sdk.NewCoins(),
		fmt.Sprintf("disputed: %s", reason),
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	return nil
}

// MarkOverdue transitions to overdue status
func (sm *InvoiceStatusMachine) MarkOverdue(initiator string) error {
	from := sm.invoice.Status

	if !IsValidTransition(from, InvoiceStatusOverdue) {
		return fmt.Errorf("cannot mark as overdue from status: %s", from)
	}

	// Check if actually past due date
	if sm.now.Before(sm.invoice.DueDate) {
		return fmt.Errorf("invoice is not past due date")
	}

	sm.invoice.Status = InvoiceStatusOverdue
	sm.invoice.UpdatedAt = sm.now

	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-overdue-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		LedgerEntryTypeOverdue,
		from,
		InvoiceStatusOverdue,
		sdk.NewCoins(),
		"marked as overdue",
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	return nil
}

// Cancel cancels the invoice
func (sm *InvoiceStatusMachine) Cancel(initiator string, reason string) error {
	from := sm.invoice.Status

	if err := sm.invoice.Cancel(sm.now); err != nil {
		return err
	}

	sm.ledgerEntry = NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-cancelled-%d", sm.invoice.InvoiceID, sm.blockHeight),
		sm.invoice.InvoiceID,
		LedgerEntryTypeCancelled,
		from,
		InvoiceStatusCancelled,
		sdk.NewCoins(),
		fmt.Sprintf("cancelled: %s", reason),
		initiator,
		"",
		sm.blockHeight,
		sm.now,
	)

	return nil
}

// GetLedgerEntry returns the ledger entry created by the last operation
func (sm *InvoiceStatusMachine) GetLedgerEntry() *InvoiceLedgerEntry {
	return sm.ledgerEntry
}

// GetInvoice returns the modified invoice
func (sm *InvoiceStatusMachine) GetInvoice() *Invoice {
	return sm.invoice
}

// getEntryTypeForTransition maps status changes to entry types
func (sm *InvoiceStatusMachine) getEntryTypeForTransition(to InvoiceStatus) InvoiceLedgerEntryType {
	switch to {
	case InvoiceStatusPending:
		return LedgerEntryTypeIssued
	case InvoiceStatusPaid, InvoiceStatusPartiallyPaid:
		return LedgerEntryTypePayment
	case InvoiceStatusDisputed:
		return LedgerEntryTypeDisputed
	case InvoiceStatusCancelled:
		return LedgerEntryTypeCancelled
	case InvoiceStatusRefunded:
		return LedgerEntryTypeRefunded
	case InvoiceStatusOverdue:
		return LedgerEntryTypeOverdue
	default:
		return LedgerEntryTypeCreated
	}
}
