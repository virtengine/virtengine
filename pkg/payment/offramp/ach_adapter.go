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
	"strings"
	"time"
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

	baseURL := "https://api.stripe.com/v1"
	if config.Environment == "sandbox" {
		// Stripe uses the same URL for test mode, just with test API keys
		baseURL = "https://api.stripe.com/v1"
	}

	return &ACHAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
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

// DirectACHAdapter provides direct ACH integration without Stripe.
// This is a placeholder for future implementation with providers like Plaid or Dwolla.
type DirectACHAdapter struct {
	config     ACHConfig
	httpClient *http.Client
}

// NewDirectACHAdapter creates a new direct ACH adapter.
func NewDirectACHAdapter(config ACHConfig) (*DirectACHAdapter, error) {
	return &DirectACHAdapter{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Name returns the provider name.
func (a *DirectACHAdapter) Name() string {
	return "Direct ACH"
}

// Type returns the provider type.
func (a *DirectACHAdapter) Type() ProviderType {
	return ProviderACH
}

// IsHealthy checks if the provider is operational.
func (a *DirectACHAdapter) IsHealthy(ctx context.Context) bool {
	return true // Placeholder
}

// Close releases resources.
func (a *DirectACHAdapter) Close() error {
	return nil
}

// CreatePayout creates an ACH payout.
func (a *DirectACHAdapter) CreatePayout(ctx context.Context, intent *PayoutIntent) error {
	// Placeholder for direct ACH implementation
	return fmt.Errorf("direct ACH not implemented - use Stripe Treasury adapter")
}

// GetPayoutStatus retrieves the status of an ACH payout.
func (a *DirectACHAdapter) GetPayoutStatus(ctx context.Context, providerPayoutID string) (PayoutStatus, error) {
	return "", fmt.Errorf("direct ACH not implemented")
}

// CancelPayout cancels a pending ACH payout.
func (a *DirectACHAdapter) CancelPayout(ctx context.Context, providerPayoutID string) error {
	return fmt.Errorf("direct ACH not implemented")
}

// ValidateWebhook verifies a webhook signature.
func (a *DirectACHAdapter) ValidateWebhook(payload []byte, signature string) error {
	return nil
}

// ParseWebhookEvent parses a webhook event.
func (a *DirectACHAdapter) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	return nil, fmt.Errorf("direct ACH not implemented")
}

// GetSettlementReport retrieves a settlement report.
func (a *DirectACHAdapter) GetSettlementReport(ctx context.Context, req SettlementReportRequest) (*SettlementReport, error) {
	return nil, fmt.Errorf("direct ACH not implemented")
}

// Ensure adapters implement Provider interface
var (
	_ Provider = (*ACHAdapter)(nil)
	_ Provider = (*DirectACHAdapter)(nil)
)

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

// ============================================================================
// Unused import prevention
// ============================================================================

var _ = bytes.Buffer{} // Keep bytes import for potential use

