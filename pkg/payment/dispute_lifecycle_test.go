// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Tests for dispute lifecycle, webhooks, and reporting
package payment

import (
	"testing"
	"time"
)

func TestDisputeLifecycleTransitions(t *testing.T) {
	lifecycle := DefaultDisputeLifecycle()
	record := NewDisputeRecord(Dispute{
		ID:     "dp_lifecycle",
		Status: DisputeStatusNeedsResponse,
	}, GatewayStripe, "test")

	if err := lifecycle.Transition(record, DisputeStatusUnderReview, "test", nil); err != nil {
		t.Fatalf("expected valid transition: %v", err)
	}
	if record.Status != DisputeStatusUnderReview {
		t.Fatalf("expected status under_review, got %s", record.Status)
	}

	if err := lifecycle.Transition(record, DisputeStatusLost, "test", nil); err != nil {
		t.Fatalf("expected valid transition to lost: %v", err)
	}

	if err := lifecycle.Transition(record, DisputeStatusNeedsResponse, "test", nil); err == nil {
		t.Fatalf("expected invalid transition to be rejected")
	}
}

func TestGenerateDisputeReport(t *testing.T) {
	store := NewInMemoryDisputeStore()
	now := time.Now()

	openDispute := NewDisputeRecord(Dispute{
		ID:            "dp_open",
		Status:        DisputeStatusNeedsResponse,
		Amount:        NewAmount(10000, CurrencyUSD),
		EvidenceDueBy: now.Add(48 * time.Hour),
		CreatedAt:     now.Add(-48 * time.Hour),
	}, GatewayStripe, "test")
	openDispute.CreatedAt = now.Add(-48 * time.Hour)
	openDispute.EvidenceSubmitted = true
	_ = store.Save(openDispute)

	closedDispute := NewDisputeRecord(Dispute{
		ID:        "dp_closed",
		Status:    DisputeStatusWon,
		Amount:    NewAmount(20000, CurrencyUSD),
		CreatedAt: now.Add(-10 * 24 * time.Hour),
	}, GatewayStripe, "test")
	closedDispute.CreatedAt = now.Add(-10 * 24 * time.Hour)
	closedDispute.ClosedAt = ptrTime(now.Add(-24 * time.Hour))
	closedDispute.Status = DisputeStatusWon
	_ = store.Save(closedDispute)

	report, err := GenerateDisputeReport(store, DisputeReportOptions{})
	if err != nil {
		t.Fatalf("report generation failed: %v", err)
	}
	if report.TotalDisputes != 2 {
		t.Fatalf("expected 2 disputes, got %d", report.TotalDisputes)
	}
	if report.OpenDisputes != 1 {
		t.Fatalf("expected 1 open dispute, got %d", report.OpenDisputes)
	}
	if report.ClosedDisputes != 1 {
		t.Fatalf("expected 1 closed dispute, got %d", report.ClosedDisputes)
	}
	if report.TotalAmountByCurrency[CurrencyUSD] != 30000 {
		t.Fatalf("expected total amount 30000, got %d", report.TotalAmountByCurrency[CurrencyUSD])
	}
	if report.EvidenceSubmittedCount != 1 {
		t.Fatalf("expected 1 evidence submitted, got %d", report.EvidenceSubmittedCount)
	}
}

func TestExtractDisputeFromWebhookEventStripe(t *testing.T) {
	payload := []byte(`{
		"id":"evt_123",
		"type":"charge.dispute.created",
		"created":1700000000,
		"data":{
			"object":{
				"id":"dp_123",
				"status":"needs_response",
				"reason":"fraudulent",
				"payment_intent":"pi_123",
				"charge":"ch_123",
				"amount":1200,
				"currency":"usd",
				"evidence_details":{"due_by":1700003600}
			}
		}
	}`)

	event := WebhookEvent{
		ID:        "evt_123",
		Type:      WebhookEventChargeDisputeCreated,
		Gateway:   GatewayStripe,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		t.Fatalf("expected dispute to be parsed")
	}
	if dispute.ID != "dp_123" {
		t.Fatalf("expected dispute ID dp_123, got %s", dispute.ID)
	}
	if dispute.Status != DisputeStatusNeedsResponse {
		t.Fatalf("expected status needs_response, got %s", dispute.Status)
	}
	if dispute.Reason != DisputeReasonFraudulent {
		t.Fatalf("expected reason fraudulent, got %s", dispute.Reason)
	}
	if dispute.Amount.Value != 1200 {
		t.Fatalf("expected amount 1200, got %d", dispute.Amount.Value)
	}
}

func TestExtractDisputeFromWebhookEventAdyen(t *testing.T) {
	payload := []byte(`{
		"notificationItems": [
			{
				"NotificationRequestItem": {
					"eventCode": "CHARGEBACK",
					"eventDate": "2024-01-02T15:04:05Z",
					"pspReference": "psp_123",
					"amount": {"value": 5000, "currency": "EUR"},
					"additionalData": {
						"paymentPspReference": "pi_adyen",
						"disputeReason": "Fraud",
						"defenseDeadline": "2024-01-10T00:00:00Z"
					}
				}
			}
		]
	}`)

	event := WebhookEvent{
		ID:        "evt_adyen",
		Type:      WebhookEventChargeDisputeCreated,
		Gateway:   GatewayAdyen,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	dispute := extractDisputeFromWebhookEvent(event)
	if dispute == nil {
		t.Fatalf("expected dispute to be parsed")
	}
	if dispute.ID != "psp_123" {
		t.Fatalf("expected dispute ID psp_123, got %s", dispute.ID)
	}
	if dispute.PaymentIntentID != "pi_adyen" {
		t.Fatalf("expected payment intent pi_adyen, got %s", dispute.PaymentIntentID)
	}
	if dispute.Reason != DisputeReasonFraudulent {
		t.Fatalf("expected reason fraudulent, got %s", dispute.Reason)
	}
	if dispute.Status != DisputeStatusNeedsResponse {
		t.Fatalf("expected status needs_response, got %s", dispute.Status)
	}
	if dispute.Amount.Value != 5000 {
		t.Fatalf("expected amount 5000, got %d", dispute.Amount.Value)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
