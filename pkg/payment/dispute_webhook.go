// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import (
	"encoding/json"
	"time"
)

// extractDisputeFromWebhookEvent returns dispute data from event payloads.
func extractDisputeFromWebhookEvent(event WebhookEvent) *Dispute {
	if dispute := extractDisputeFromEventData(event); dispute != nil {
		return dispute
	}

	if len(event.Payload) == 0 {
		return nil
	}

	switch event.Gateway {
	case GatewayStripe:
		dispute, _ := parseStripeDisputePayload(event.Payload, event)
		return dispute
	case GatewayAdyen:
		dispute, _ := parseAdyenDisputePayload(event.Payload, event)
		return dispute
	default:
		return nil
	}
}

func extractDisputeFromEventData(event WebhookEvent) *Dispute {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	obj := data
	if nested, ok := data["object"].(map[string]interface{}); ok {
		obj = nested
	}

	dispute := disputeFromMap(obj, event)
	if dispute == nil {
		return nil
	}

	return dispute
}

func parseStripeDisputePayload(payload []byte, event WebhookEvent) (*Dispute, error) {
	var envelope struct {
		Created int64 `json:"created"`
		Data    struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, err
	}

	if len(envelope.Data.Object) == 0 {
		return nil, nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(envelope.Data.Object, &obj); err != nil {
		return nil, err
	}

	dispute := disputeFromMap(obj, event)
	if dispute == nil {
		return nil, nil
	}

	if dispute.CreatedAt.IsZero() && envelope.Created > 0 {
		dispute.CreatedAt = time.Unix(envelope.Created, 0)
	}

	return dispute, nil
}

func parseAdyenDisputePayload(payload []byte, event WebhookEvent) (*Dispute, error) {
	var notification struct {
		NotificationItems []struct {
			NotificationRequestItem struct {
				EventCode       string `json:"eventCode"`
				EventDate       string `json:"eventDate"`
				PSPReference    string `json:"pspReference"`
				MerchantAccount string `json:"merchantAccount"`
				Amount          struct {
					Value    int64  `json:"value"`
					Currency string `json:"currency"`
				} `json:"amount"`
				AdditionalData map[string]interface{} `json:"additionalData"`
			} `json:"NotificationRequestItem"`
		} `json:"notificationItems"`
	}

	if err := json.Unmarshal(payload, &notification); err != nil {
		return nil, err
	}

	if len(notification.NotificationItems) == 0 {
		return nil, nil
	}

	item := notification.NotificationItems[0].NotificationRequestItem
	dispute := &Dispute{
		ID:        item.PSPReference,
		Gateway:   event.Gateway,
		CreatedAt: event.Timestamp,
		UpdatedAt: time.Now(),
		Amount: Amount{
			Value:    item.Amount.Value,
			Currency: Currency(item.Amount.Currency),
		},
	}

	if item.EventDate != "" {
		if parsed, err := time.Parse(time.RFC3339, item.EventDate); err == nil {
			dispute.CreatedAt = parsed
		}
	}

	if ref, ok := readString(item.AdditionalData, "paymentPspReference", "payment_psp_reference", "originalReference", "original_reference"); ok {
		dispute.PaymentIntentID = ref
	}

	if reason, ok := readString(item.AdditionalData, "disputeReason", "reason", "chargebackReason"); ok {
		dispute.Reason = mapAdyenDisputeReason(reason)
	}

	if dueBy, ok := readString(item.AdditionalData, "defenseDeadline", "defenseDeadlineDate"); ok {
		if parsed, err := time.Parse(time.RFC3339, dueBy); err == nil {
			dispute.EvidenceDueBy = parsed
		}
	}

	if dispute.Status == "" {
		dispute.Status = statusFromWebhookEvent(event.Type)
	}

	if dispute.ID == "" {
		return nil, nil
	}

	return dispute, nil
}

func disputeFromMap(obj map[string]interface{}, event WebhookEvent) *Dispute {
	if obj == nil {
		return nil
	}

	dispute := &Dispute{
		Gateway:   event.Gateway,
		CreatedAt: event.Timestamp,
		UpdatedAt: time.Now(),
	}

	if id, ok := readString(obj, "id", "pspReference"); ok {
		dispute.ID = id
	}
	if status, ok := readString(obj, "status"); ok {
		dispute.Status = normalizeDisputeStatus(status)
	}
	if reason, ok := readString(obj, "reason"); ok {
		dispute.Reason = normalizeDisputeReason(reason)
	}

	if piID, ok := readString(obj, "payment_intent", "paymentIntentId", "paymentIntentID"); ok {
		dispute.PaymentIntentID = piID
	}
	if chargeID, ok := readString(obj, "charge", "charge_id"); ok {
		dispute.ChargeID = chargeID
	}

	if amountVal, ok := readInt64(obj, "amount", "amount.value"); ok {
		dispute.Amount.Value = amountVal
	}
	if currency, ok := readString(obj, "currency", "amount.currency"); ok {
		dispute.Amount.Currency = Currency(currency)
	}

	if dueBy, ok := readInt64(obj, "evidence_due_by", "evidence_details.due_by"); ok {
		dispute.EvidenceDueBy = time.Unix(dueBy, 0)
	}
	if created, ok := readInt64(obj, "created"); ok {
		dispute.CreatedAt = time.Unix(created, 0)
	}

	if dispute.Status == "" {
		dispute.Status = statusFromWebhookEvent(event.Type)
	}

	if dispute.ID == "" {
		return nil
	}

	return dispute
}

func normalizeDisputeStatus(status string) DisputeStatus {
	switch DisputeStatus(status) {
	case DisputeStatusOpen, DisputeStatusNeedsResponse, DisputeStatusUnderReview, DisputeStatusWon,
		DisputeStatusLost, DisputeStatusAccepted, DisputeStatusExpired:
		return DisputeStatus(status)
	default:
		return DisputeStatusOpen
	}
}

func normalizeDisputeReason(reason string) DisputeReason {
	switch DisputeReason(reason) {
	case DisputeReasonFraudulent, DisputeReasonDuplicate, DisputeReasonProductNotReceived,
		DisputeReasonUnrecognized:
		return DisputeReason(reason)
	default:
		return DisputeReasonGeneral
	}
}

func statusFromWebhookEvent(eventType WebhookEventType) DisputeStatus {
	switch eventType {
	case WebhookEventChargeDisputeCreated:
		return DisputeStatusNeedsResponse
	case WebhookEventChargeDisputeUpdated:
		return DisputeStatusUnderReview
	case WebhookEventChargeDisputeClosed:
		return DisputeStatusWon
	case WebhookEventChargeDisputeFundsWithdrawn:
		return DisputeStatusLost
	case WebhookEventChargeDisputeFundsReinstated:
		return DisputeStatusWon
	default:
		return DisputeStatusOpen
	}
}

func readString(obj map[string]interface{}, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := readValue(obj, key); ok {
			if str, ok := value.(string); ok {
				return str, true
			}
		}
	}
	return "", false
}

func readInt64(obj map[string]interface{}, keys ...string) (int64, bool) {
	for _, key := range keys {
		if value, ok := readValue(obj, key); ok {
			switch typed := value.(type) {
			case float64:
				return int64(typed), true
			case int64:
				return typed, true
			case json.Number:
				if parsed, err := typed.Int64(); err == nil {
					return parsed, true
				}
			}
		}
	}
	return 0, false
}

func readValue(obj map[string]interface{}, key string) (interface{}, bool) {
	if obj == nil {
		return nil, false
	}
	if value, ok := obj[key]; ok {
		return value, true
	}

	// Support dotted paths (e.g., evidence_details.due_by)
	if path := splitKey(key); len(path) > 1 {
		current := obj
		for i, part := range path {
			value, ok := current[part]
			if !ok {
				return nil, false
			}
			if i == len(path)-1 {
				return value, true
			}
			next, ok := value.(map[string]interface{})
			if !ok {
				return nil, false
			}
			current = next
		}
	}

	return nil, false
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	if start == 0 {
		return []string{key}
	}
	parts = append(parts, key[start:])
	return parts
}
