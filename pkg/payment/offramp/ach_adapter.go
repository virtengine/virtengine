// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// ACH Adapter (via Stripe Treasury)
// ============================================================================

// ACHAdapter implements the Provider interface for ACH bank transfers.
// This implementation uses Stripe Treasury for ACH payouts.
type ACHAdapter struct {
	config     ACHConfig
	httpClient *http.Client
	baseURL    string
}

// NewACHAdapter creates a new ACH adapter.
func NewACHAdapter(config ACHConfig) (*ACHAdapter, error) {
	if config.SecretKey == "" {
		return nil, ErrProviderNotConfigured
	}

	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.stripe.com/v1"
	}

	return &ACHAdapter{
		config:     config,
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
		baseURL:    baseURL,
	}, nil
}

// Name returns the provider name.
func (a *ACHAdapter) Name() string {
	return "ACH"
}

// Type returns the provider type.
func (a *ACHAdapter) Type() ProviderType {
	return ProviderACH
}

// IsHealthy checks if the ACH provider is operational.
func (a *ACHAdapter) IsHealthy(ctx context.Context) bool {
	// Verify API connectivity by checking account
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/account", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Close releases resources.
func (a *ACHAdapter) Close() error {
	return nil
}

// ============================================================================
// Payout Operations
// ============================================================================

// CreatePayout creates an ACH payout via Stripe Treasury.
func (a *ACHAdapter) CreatePayout(ctx context.Context, intent *PayoutIntent) error {
	if intent.Destination.BankAccount == nil {
		return ErrInvalidPayoutDestination
	}

	// Create an OutboundPayment using Stripe Treasury
	// First, we need to create a PaymentMethod for the bank account
	paymentMethodID, err := a.createBankAccountPaymentMethod(ctx, intent.Destination.BankAccount)
	if err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}

	// Create the outbound payment
	reqBody := fmt.Sprintf(
		"financial_account=%s&amount=%d&currency=%s&destination_payment_method=%s&description=%s&metadata[payout_id]=%s&statement_descriptor=%s",
		a.config.SourceAccountID,
		intent.FiatAmount.Value,
		strings.ToLower(string(intent.FiatAmount.Currency)),
		paymentMethodID,
		intent.Description,
		intent.ID,
		"VirtEngine Payout",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/treasury/outbound_payments", strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", intent.IdempotencyKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp stripeErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			intent.FailureCode = errResp.Error.Code
			intent.FailureMessage = errResp.Error.Message
		}
		return fmt.Errorf("payout failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var paymentResp stripeOutboundPayment
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	intent.ProviderPayoutID = paymentResp.ID
	intent.Status = mapStripeOutboundStatus(paymentResp.Status)
	intent.UpdatedAt = time.Now()

	intent.AddAuditEntry("payout_created", "ach_adapter", fmt.Sprintf("outbound_payment_id=%s", paymentResp.ID))

	return nil
}

// createBankAccountPaymentMethod creates a payment method for the bank account.
func (a *ACHAdapter) createBankAccountPaymentMethod(ctx context.Context, bank *BankAccountDetails) (string, error) {
	reqBody := fmt.Sprintf(
		"type=us_bank_account&us_bank_account[account_holder_type]=%s&us_bank_account[account_type]=%s&us_bank_account[routing_number]=%s&us_bank_account[account_number]=%s",
		bank.AccountHolderType,
		bank.AccountType,
		bank.RoutingNumber,
		bank.AccountNumber,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/payment_methods", strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create payment method: %s", string(body))
	}

	var pmResp struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pmResp); err != nil {
		return "", err
	}

	return pmResp.ID, nil
}

// GetPayoutStatus retrieves the status of an ACH payout.
func (a *ACHAdapter) GetPayoutStatus(ctx context.Context, providerPayoutID string) (PayoutStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/treasury/outbound_payments/"+providerPayoutID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrPayoutNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var paymentResp stripeOutboundPayment
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return mapStripeOutboundStatus(paymentResp.Status), nil
}

// CancelPayout cancels a pending ACH payout.
func (a *ACHAdapter) CancelPayout(ctx context.Context, providerPayoutID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/treasury/outbound_payments/"+providerPayoutID+"/cancel", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrPayoutNotFound
	}

	// Stripe returns 400 if the payment cannot be canceled
	if resp.StatusCode == http.StatusBadRequest {
		return ErrPayoutAlreadyProcessed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cancel failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ============================================================================
// Webhook Handling
// ============================================================================

// ValidateWebhook verifies a Stripe webhook signature.
func (a *ACHAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return nil // Webhook verification disabled
	}

	// Parse Stripe signature format: t=timestamp,v1=signature
	parts := strings.Split(signature, ",")
	var timestamp, sig string
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			sig = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || sig == "" {
		return ErrWebhookSignatureInvalid
	}

	// Compute expected signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(a.config.WebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

// ParseWebhookEvent parses a Stripe webhook event.
func (a *ACHAdapter) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event stripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	webhookEvent := &WebhookEvent{
		ID:         event.ID,
		Provider:   ProviderACH,
		Payload:    payload,
		Timestamp:  time.Unix(event.Created, 0),
		ReceivedAt: time.Now(),
	}

	// Map event type
	switch event.Type {
	case "treasury.outbound_payment.posted":
		webhookEvent.Type = WebhookPayoutCompleted
		webhookEvent.Status = PayoutStatusSucceeded
	case "treasury.outbound_payment.failed":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusFailed
	case "treasury.outbound_payment.returned":
		webhookEvent.Type = WebhookPayoutReturned
		webhookEvent.Status = PayoutStatusReversed
	case "treasury.outbound_payment.canceled":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusCanceled
	case "treasury.outbound_payment.processing":
		webhookEvent.Type = WebhookPayoutPending
		webhookEvent.Status = PayoutStatusProcessing
	default:
		webhookEvent.Type = WebhookEventType(event.Type)
	}

	// Extract payout ID from event data
	if data, ok := event.Data.Object.(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			webhookEvent.ProviderPayoutID = id
		}
		if metadata, ok := data["metadata"].(map[string]interface{}); ok {
			if payoutID, ok := metadata["payout_id"].(string); ok {
				webhookEvent.PayoutID = payoutID
			}
		}
		if returnedDetails, ok := data["returned_details"].(map[string]interface{}); ok {
			if code, ok := returnedDetails["code"].(string); ok {
				webhookEvent.FailureCode = code
			}
			if msg, ok := returnedDetails["message"].(string); ok {
				webhookEvent.FailureMessage = msg
			}
		}
	}

	return webhookEvent, nil
}

// ============================================================================
// Reconciliation
// ============================================================================

// GetSettlementReport retrieves a settlement report from Stripe.
func (a *ACHAdapter) GetSettlementReport(ctx context.Context, req SettlementReportRequest) (*SettlementReport, error) {
	// Parse dates to Unix timestamps
	startTime, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endTime, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// List outbound payments for the period
	listURL := fmt.Sprintf("%s/treasury/outbound_payments?financial_account=%s&created[gte]=%d&created[lte]=%d&limit=100",
		a.baseURL, a.config.SourceAccountID, startTime.Unix(), endTime.Unix())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.config.SecretKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResp struct {
		Data []stripeOutboundPayment `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to settlement report
	report := &SettlementReport{
		ReportID:     fmt.Sprintf("ach_%s_%s", req.StartDate, req.EndDate),
		Provider:     ProviderACH,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		TotalPayouts: len(listResp.Data),
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Transactions: make([]SettlementTransaction, 0, len(listResp.Data)),
	}

	for _, payment := range listResp.Data {
		tx := SettlementTransaction{
			TransactionID: payment.ID,
			Amount:        payment.Amount,
			Status:        payment.Status,
			ProcessedAt:   time.Unix(payment.Created, 0).Format(time.RFC3339),
		}

		// Get payout ID from metadata
		if payoutID, ok := payment.Metadata["payout_id"]; ok {
			tx.PayoutID = payoutID
		}

		report.TotalAmount += payment.Amount
		report.Transactions = append(report.Transactions, tx)
	}

	return report, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func mapStripeOutboundStatus(status string) PayoutStatus {
	switch status {
	case "processing":
		return PayoutStatusProcessing
	case "posted":
		return PayoutStatusSucceeded
	case "failed":
		return PayoutStatusFailed
	case "canceled":
		return PayoutStatusCanceled
	case "returned":
		return PayoutStatusReversed
	default:
		return PayoutStatusPending
	}
}

// ============================================================================
// Stripe API Types
// ============================================================================

type stripeOutboundPayment struct {
	ID              string            `json:"id"`
	Object          string            `json:"object"`
	Amount          int64             `json:"amount"`
	Currency        string            `json:"currency"`
	Created         int64             `json:"created"`
	Status          string            `json:"status"`
	Metadata        map[string]string `json:"metadata"`
	ReturnedDetails *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"returned_details,omitempty"`
}

type stripeWebhookEvent struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Created int64  `json:"created"`
	Data    struct {
		Object interface{} `json:"object"`
	} `json:"data"`
}

type stripeErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ============================================================================
// Alternative: Direct ACH Adapter (for Plaid or other providers)
// ============================================================================

// DirectACHAdapter provides generic direct ACH integration for non-Stripe providers.
type DirectACHAdapter struct {
	config     ACHConfig
	httpClient *http.Client
	baseURL    string
}

// NewDirectACHAdapter creates a new direct ACH adapter.
func NewDirectACHAdapter(config ACHConfig) (*DirectACHAdapter, error) {
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if strings.TrimSpace(config.SecretKey) == "" || baseURL == "" {
		return nil, ErrProviderNotConfigured
	}

	return &DirectACHAdapter{
		config:     config,
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
		baseURL:    baseURL,
	}, nil
}

// Name returns the provider name.
func (a *DirectACHAdapter) Name() string {
	if provider := normalizedACHProvider(a.config.Provider); provider != "" {
		return "Direct ACH (" + provider + ")"
	}
	return "Direct ACH"
}

// Type returns the provider type.
func (a *DirectACHAdapter) Type() ProviderType {
	return ProviderACH
}

// IsHealthy checks if the provider is operational.
func (a *DirectACHAdapter) IsHealthy(ctx context.Context) bool {
	_, statusCode, err := a.doRequest(ctx, http.MethodGet, "/health", nil, "", nil)
	return err == nil && (statusCode == http.StatusOK || statusCode == http.StatusNoContent)
}

// Close releases resources.
func (a *DirectACHAdapter) Close() error {
	return nil
}

// CreatePayout creates an ACH payout.
func (a *DirectACHAdapter) CreatePayout(ctx context.Context, intent *PayoutIntent) error {
	if intent.Destination.BankAccount == nil {
		return ErrInvalidPayoutDestination
	}

	reqBody := directACHPayoutRequest{
		PayoutID:        intent.ID,
		IdempotencyKey:  intent.IdempotencyKey,
		SourceAccountID: strings.TrimSpace(a.config.SourceAccountID),
		Amount:          intent.FiatAmount.Value,
		Currency:        strings.ToLower(string(intent.FiatAmount.Currency)),
		Description:     intent.Description,
		SameDay:         a.config.EnableSameDayACH,
		Destination: directACHBankAccount{
			AccountHolderName: intent.Destination.BankAccount.AccountHolderName,
			AccountHolderType: intent.Destination.BankAccount.AccountHolderType,
			RoutingNumber:     intent.Destination.BankAccount.RoutingNumber,
			AccountNumber:     intent.Destination.BankAccount.AccountNumber,
			AccountType:       intent.Destination.BankAccount.AccountType,
			BankName:          intent.Destination.BankAccount.BankName,
			Country:           intent.Destination.BankAccount.Country,
		},
		Metadata: intent.Metadata,
	}

	respBody, _, err := a.doRequest(ctx, http.MethodPost, "/payouts", reqBody, intent.IdempotencyKey, nil)
	if err != nil {
		return err
	}

	var payoutResp directACHPayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return fmt.Errorf("failed to decode direct ACH payout response: %w", err)
	}
	if strings.TrimSpace(payoutResp.ID) == "" {
		return fmt.Errorf("direct ACH payout response missing provider payout ID")
	}

	intent.ProviderPayoutID = payoutResp.ID
	intent.ProviderBatchID = payoutResp.BatchID
	intent.Status = mapDirectACHStatus(payoutResp.Status)
	intent.FailureCode = payoutResp.FailureCode
	intent.FailureMessage = payoutResp.FailureMessage
	intent.UpdatedAt = time.Now()
	intent.AddAuditEntry("payout_created", "direct_ach_adapter", fmt.Sprintf("provider_payout_id=%s", payoutResp.ID))
	return nil
}

// GetPayoutStatus retrieves the status of an ACH payout.
func (a *DirectACHAdapter) GetPayoutStatus(ctx context.Context, providerPayoutID string) (PayoutStatus, error) {
	respBody, statusCode, err := a.doRequest(ctx, http.MethodGet, "/payouts/"+providerPayoutID, nil, "", nil)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return "", ErrPayoutNotFound
		}
		return "", err
	}

	var payoutResp directACHPayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return "", fmt.Errorf("failed to decode direct ACH payout status response: %w", err)
	}
	return mapDirectACHStatus(payoutResp.Status), nil
}

// CancelPayout cancels a pending ACH payout.
func (a *DirectACHAdapter) CancelPayout(ctx context.Context, providerPayoutID string) error {
	_, statusCode, err := a.doRequest(ctx, http.MethodPost, "/payouts/"+providerPayoutID+"/cancel", nil, "", nil)
	if err != nil {
		switch statusCode {
		case http.StatusNotFound:
			return ErrPayoutNotFound
		case http.StatusBadRequest, http.StatusConflict:
			return ErrPayoutAlreadyProcessed
		default:
			return err
		}
	}
	return nil
}

// ValidateWebhook verifies a webhook signature.
func (a *DirectACHAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return nil
	}

	normalized := strings.TrimSpace(signature)
	normalized = strings.TrimPrefix(normalized, "sha256=")
	if normalized == "" {
		return ErrWebhookSignatureInvalid
	}

	mac := hmac.New(sha256.New, []byte(a.config.WebhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(normalized)), []byte(expectedSig)) {
		return ErrWebhookSignatureInvalid
	}
	return nil
}

// ParseWebhookEvent parses a webhook event.
func (a *DirectACHAdapter) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event directACHWebhookEnvelope
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse direct ACH webhook event: %w", err)
	}

	timestamp := directACHTimestamp(event.Timestamp, event.CreatedAt)
	if event.Created > 0 {
		timestamp = time.Unix(event.Created, 0)
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	webhookEvent := &WebhookEvent{
		ID:               event.ID,
		Provider:         ProviderACH,
		PayoutID:         event.Data.PayoutID,
		ProviderPayoutID: firstNonEmpty(event.Data.ProviderPayoutID, event.Data.ID),
		Status:           mapDirectACHStatus(event.Data.Status),
		FailureCode:      event.Data.FailureCode,
		FailureMessage:   event.Data.FailureMessage,
		Payload:          payload,
		Timestamp:        timestamp,
		ReceivedAt:       time.Now(),
	}

	if webhookEvent.PayoutID == "" && event.Data.Metadata != nil {
		webhookEvent.PayoutID = event.Data.Metadata["payout_id"]
	}

	eventType := strings.ToLower(strings.TrimSpace(event.Type))
	switch eventType {
	case "payout.completed", "payout.succeeded":
		webhookEvent.Type = WebhookPayoutCompleted
		webhookEvent.Status = PayoutStatusSucceeded
	case "payout.failed":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusFailed
	case "payout.returned", "payout.reversed":
		webhookEvent.Type = WebhookPayoutReturned
		webhookEvent.Status = PayoutStatusReversed
	case "payout.processing", "payout.pending":
		webhookEvent.Type = WebhookPayoutPending
		if webhookEvent.Status == "" {
			webhookEvent.Status = PayoutStatusProcessing
		}
	case "payout.canceled", "payout.cancelled":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusCanceled
	default:
		webhookEvent.Type = WebhookEventType(event.Type)
		if webhookEvent.Status == "" {
			webhookEvent.Status = PayoutStatusPending
		}
	}

	return webhookEvent, nil
}

// GetSettlementReport retrieves a settlement report.
func (a *DirectACHAdapter) GetSettlementReport(ctx context.Context, req SettlementReportRequest) (*SettlementReport, error) {
	startTime, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endTime, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end date must not be before start date")
	}

	query := url.Values{}
	query.Set("start_date", req.StartDate)
	query.Set("end_date", req.EndDate)
	if strings.TrimSpace(req.Format) != "" {
		query.Set("format", req.Format)
	}

	respBody, _, err := a.doRequest(ctx, http.MethodGet, "/settlements", nil, "", query)
	if err != nil {
		return nil, err
	}

	var settlementResp directACHSettlementReportResponse
	if err := json.Unmarshal(respBody, &settlementResp); err != nil {
		return nil, fmt.Errorf("failed to decode direct ACH settlement report: %w", err)
	}

	report := &SettlementReport{
		ReportID:     firstNonEmpty(settlementResp.ReportID, settlementResp.ID, fmt.Sprintf("direct_ach_%s_%s", req.StartDate, req.EndDate)),
		Provider:     ProviderACH,
		StartDate:    firstNonEmpty(settlementResp.StartDate, req.StartDate),
		EndDate:      firstNonEmpty(settlementResp.EndDate, req.EndDate),
		TotalPayouts: settlementResp.TotalPayouts,
		TotalAmount:  settlementResp.TotalAmount,
		TotalFees:    settlementResp.TotalFees,
		GeneratedAt:  firstNonEmpty(settlementResp.GeneratedAt, time.Now().Format(time.RFC3339)),
		Transactions: make([]SettlementTransaction, 0, len(settlementResp.Transactions)),
	}

	for _, tx := range settlementResp.Transactions {
		report.Transactions = append(report.Transactions, SettlementTransaction{
			TransactionID: tx.TransactionID,
			PayoutID:      tx.PayoutID,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
			Status:        string(mapDirectACHStatus(tx.Status)),
			ProcessedAt:   firstNonEmpty(tx.ProcessedAt, time.Now().Format(time.RFC3339)),
		})
		if report.TotalAmount == 0 {
			report.TotalAmount += tx.Amount
		}
		if report.TotalFees == 0 {
			report.TotalFees += tx.Fee
		}
	}

	if report.TotalPayouts == 0 {
		report.TotalPayouts = len(report.Transactions)
	}

	return report, nil
}

// Ensure adapters implement Provider interface
var (
	_ Provider = (*ACHAdapter)(nil)
	_ Provider = (*DirectACHAdapter)(nil)
)

type directACHPayoutRequest struct {
	PayoutID        string               `json:"payout_id"`
	IdempotencyKey  string               `json:"idempotency_key,omitempty"`
	SourceAccountID string               `json:"source_account_id,omitempty"`
	Amount          int64                `json:"amount"`
	Currency        string               `json:"currency"`
	Description     string               `json:"description,omitempty"`
	SameDay         bool                 `json:"same_day"`
	Destination     directACHBankAccount `json:"destination"`
	Metadata        map[string]string    `json:"metadata,omitempty"`
}

type directACHBankAccount struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountHolderType string `json:"account_holder_type"`
	RoutingNumber     string `json:"routing_number"`
	AccountNumber     string `json:"account_number"`
	AccountType       string `json:"account_type"`
	BankName          string `json:"bank_name,omitempty"`
	Country           string `json:"country"`
}

type directACHPayoutResponse struct {
	ID             string `json:"id"`
	BatchID        string `json:"batch_id,omitempty"`
	Status         string `json:"status"`
	FailureCode    string `json:"failure_code,omitempty"`
	FailureMessage string `json:"failure_message,omitempty"`
}

type directACHWebhookEnvelope struct {
	ID        string               `json:"id"`
	Type      string               `json:"type"`
	Timestamp string               `json:"timestamp,omitempty"`
	CreatedAt string               `json:"created_at,omitempty"`
	Created   int64                `json:"created,omitempty"`
	Data      directACHWebhookData `json:"data"`
}

type directACHWebhookData struct {
	ID               string            `json:"id,omitempty"`
	PayoutID         string            `json:"payout_id,omitempty"`
	ProviderPayoutID string            `json:"provider_payout_id,omitempty"`
	Status           string            `json:"status,omitempty"`
	FailureCode      string            `json:"failure_code,omitempty"`
	FailureMessage   string            `json:"failure_message,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

type directACHSettlementReportResponse struct {
	ID           string                           `json:"id,omitempty"`
	ReportID     string                           `json:"report_id,omitempty"`
	StartDate    string                           `json:"start_date,omitempty"`
	EndDate      string                           `json:"end_date,omitempty"`
	TotalPayouts int                              `json:"total_payouts,omitempty"`
	TotalAmount  int64                            `json:"total_amount,omitempty"`
	TotalFees    int64                            `json:"total_fees,omitempty"`
	GeneratedAt  string                           `json:"generated_at,omitempty"`
	Transactions []directACHSettlementTransaction `json:"transactions,omitempty"`
}

type directACHSettlementTransaction struct {
	TransactionID string `json:"transaction_id"`
	PayoutID      string `json:"payout_id,omitempty"`
	Amount        int64  `json:"amount"`
	Fee           int64  `json:"fee,omitempty"`
	Status        string `json:"status"`
	ProcessedAt   string `json:"processed_at,omitempty"`
}

func (a *DirectACHAdapter) doRequest(ctx context.Context, method string, path string, payload any, idempotencyKey string, query url.Values) ([]byte, int, error) {
	requestURL := a.baseURL + path
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal direct ACH payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create direct ACH request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(idempotencyKey) != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("direct ACH request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read direct ACH response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, resp.StatusCode, fmt.Errorf("direct ACH request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

func mapDirectACHStatus(status string) PayoutStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "processing", "submitted", "queued":
		return PayoutStatusProcessing
	case "succeeded", "completed", "posted", "settled":
		return PayoutStatusSucceeded
	case "failed", "rejected":
		return PayoutStatusFailed
	case "canceled", "cancelled":
		return PayoutStatusCanceled
	case "returned", "reversed":
		return PayoutStatusReversed
	case "approved":
		return PayoutStatusApproved
	default:
		return PayoutStatusPending
	}
}

func directACHTimestamp(candidates ...string) time.Time {
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			parsed, err := time.Parse(layout, candidate)
			if err == nil {
				return parsed
			}
		}
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// ============================================================================
// ACH Return Codes
// ============================================================================

// ACHReturnCode represents an ACH return reason code.
type ACHReturnCode string

const (
	// R01 - Insufficient Funds
	ACHReturnR01 ACHReturnCode = "R01"
	// R02 - Account Closed
	ACHReturnR02 ACHReturnCode = "R02"
	// R03 - No Account/Unable to Locate Account
	ACHReturnR03 ACHReturnCode = "R03"
	// R04 - Invalid Account Number
	ACHReturnR04 ACHReturnCode = "R04"
	// R05 - Unauthorized Debit to Consumer Account
	ACHReturnR05 ACHReturnCode = "R05"
	// R06 - Returned per ODFI's Request
	ACHReturnR06 ACHReturnCode = "R06"
	// R07 - Authorization Revoked by Customer
	ACHReturnR07 ACHReturnCode = "R07"
	// R08 - Payment Stopped
	ACHReturnR08 ACHReturnCode = "R08"
	// R09 - Uncollected Funds
	ACHReturnR09 ACHReturnCode = "R09"
	// R10 - Customer Advises Not Authorized
	ACHReturnR10 ACHReturnCode = "R10"
)

// IsRetryable returns true if the return code indicates a retryable error.
func (c ACHReturnCode) IsRetryable() bool {
	switch c {
	case ACHReturnR01, ACHReturnR09: // Insufficient funds, uncollected funds
		return true
	default:
		return false
	}
}

// Description returns a human-readable description of the return code.
func (c ACHReturnCode) Description() string {
	switch c {
	case ACHReturnR01:
		return "Insufficient Funds"
	case ACHReturnR02:
		return "Account Closed"
	case ACHReturnR03:
		return "No Account/Unable to Locate Account"
	case ACHReturnR04:
		return "Invalid Account Number"
	case ACHReturnR05:
		return "Unauthorized Debit"
	case ACHReturnR06:
		return "Returned per ODFI's Request"
	case ACHReturnR07:
		return "Authorization Revoked"
	case ACHReturnR08:
		return "Payment Stopped"
	case ACHReturnR09:
		return "Uncollected Funds"
	case ACHReturnR10:
		return "Customer Advises Not Authorized"
	default:
		return "Unknown Return Code"
	}
}
