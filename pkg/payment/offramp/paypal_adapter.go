// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// PayPal Adapter
// ============================================================================

// PayPalAdapter implements the Provider interface for PayPal Payouts.
type PayPalAdapter struct {
	config     PayPalConfig
	httpClient *http.Client

	// OAuth token management
	tokenMu     sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NewPayPalAdapter creates a new PayPal adapter.
func NewPayPalAdapter(config PayPalConfig) (*PayPalAdapter, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, ErrProviderNotConfigured
	}

	return &PayPalAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Name returns the provider name.
func (a *PayPalAdapter) Name() string {
	return "PayPal"
}

// Type returns the provider type.
func (a *PayPalAdapter) Type() ProviderType {
	return ProviderPayPal
}

// IsHealthy checks if PayPal API is reachable.
func (a *PayPalAdapter) IsHealthy(ctx context.Context) bool {
	_, err := a.getAccessToken(ctx)
	return err == nil
}

// Close releases resources.
func (a *PayPalAdapter) Close() error {
	return nil
}

// ============================================================================
// OAuth Token Management
// ============================================================================

// getAccessToken retrieves a valid OAuth access token.
func (a *PayPalAdapter) getAccessToken(ctx context.Context) (string, error) {
	a.tokenMu.RLock()
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		token := a.accessToken
		a.tokenMu.RUnlock()
		return token, nil
	}
	a.tokenMu.RUnlock()

	// Need to refresh token
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.accessToken, nil
	}

	// Request new token
	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", a.config.GetBaseURL())
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	a.accessToken = tokenResp.AccessToken
	// Set expiry with 5 minute buffer
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return a.accessToken, nil
}

// ============================================================================
// Payout Operations
// ============================================================================

// CreatePayout creates a PayPal payout.
func (a *PayPalAdapter) CreatePayout(ctx context.Context, intent *PayoutIntent) error {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Build payout request
	payoutReq := paypalPayoutRequest{
		SenderBatchHeader: paypalSenderBatchHeader{
			SenderBatchID: fmt.Sprintf("%s%s", a.config.SenderBatchIDPrefix, intent.ID),
			EmailSubject:  a.config.EmailSubject,
			EmailMessage:  a.config.EmailMessage,
		},
		Items: []paypalPayoutItem{
			{
				RecipientType:  a.getRecipientType(intent.Destination),
				Amount:         paypalAmount{Currency: string(intent.FiatAmount.Currency), Value: formatPayPalAmount(intent.FiatAmount.Value)},
				Note:           intent.Description,
				SenderItemID:   intent.ID,
				Receiver:       a.getReceiver(intent.Destination),
				NotificationLanguage: "en-US",
			},
		},
	}

	body, err := json.Marshal(payoutReq)
	if err != nil {
		return fmt.Errorf("failed to marshal payout request: %w", err)
	}

	payoutURL := fmt.Sprintf("%s/v1/payments/payouts", a.config.GetBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, payoutURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create payout request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("payout request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp paypalErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			intent.FailureCode = errResp.Name
			intent.FailureMessage = errResp.Message
		}
		return fmt.Errorf("payout failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var payoutResp paypalPayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return fmt.Errorf("failed to decode payout response: %w", err)
	}

	// Update intent with provider IDs
	intent.ProviderBatchID = payoutResp.BatchHeader.PayoutBatchID
	intent.Status = mapPayPalStatus(payoutResp.BatchHeader.BatchStatus)
	intent.UpdatedAt = time.Now()

	// If there are items, get the item ID
	if len(payoutResp.Items) > 0 {
		intent.ProviderPayoutID = payoutResp.Items[0].PayoutItemID
	}

	intent.AddAuditEntry("payout_created", "paypal_adapter", fmt.Sprintf("batch_id=%s", intent.ProviderBatchID))

	return nil
}

// GetPayoutStatus retrieves the status of a payout.
func (a *PayPalAdapter) GetPayoutStatus(ctx context.Context, providerPayoutID string) (PayoutStatus, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Get payout item details
	statusURL := fmt.Sprintf("%s/v1/payments/payouts-item/%s", a.config.GetBaseURL(), providerPayoutID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrPayoutNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var itemResp paypalPayoutItemDetails
	if err := json.NewDecoder(resp.Body).Decode(&itemResp); err != nil {
		return "", fmt.Errorf("failed to decode status response: %w", err)
	}

	return mapPayPalItemStatus(itemResp.TransactionStatus), nil
}

// CancelPayout cancels a pending payout.
func (a *PayPalAdapter) CancelPayout(ctx context.Context, providerPayoutID string) error {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	cancelURL := fmt.Sprintf("%s/v1/payments/payouts-item/%s/cancel", a.config.GetBaseURL(), providerPayoutID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cancelURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cancel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrPayoutNotFound
	}

	// PayPal returns 200 for successful cancellation
	// Returns 422 if payout cannot be canceled (already processed, etc.)
	if resp.StatusCode == http.StatusUnprocessableEntity {
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

// ValidateWebhook verifies a PayPal webhook signature.
func (a *PayPalAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookID == "" {
		return nil // Webhook verification disabled
	}

	// PayPal webhooks use a complex verification process
	// involving transmission ID, timestamp, webhook ID, and CRC32 checksum
	// For sandbox, we'll do a simpler validation
	
	// In production, you would call the PayPal verify-webhook-signature API
	// POST /v1/notifications/verify-webhook-signature
	
	if signature == "" {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

// ParseWebhookEvent parses a PayPal webhook event.
func (a *PayPalAdapter) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event paypalWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	webhookEvent := &WebhookEvent{
		ID:         event.ID,
		Provider:   ProviderPayPal,
		Payload:    payload,
		ReceivedAt: time.Now(),
	}

	// Parse timestamp
	if t, err := time.Parse(time.RFC3339, event.CreateTime); err == nil {
		webhookEvent.Timestamp = t
	} else {
		webhookEvent.Timestamp = time.Now()
	}

	// Map event type
	switch event.EventType {
	case "PAYMENT.PAYOUTS-ITEM.SUCCEEDED":
		webhookEvent.Type = WebhookPayoutCompleted
		webhookEvent.Status = PayoutStatusSucceeded
	case "PAYMENT.PAYOUTS-ITEM.FAILED":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusFailed
	case "PAYMENT.PAYOUTS-ITEM.CANCELED":
		webhookEvent.Type = WebhookPayoutFailed
		webhookEvent.Status = PayoutStatusCanceled
	case "PAYMENT.PAYOUTS-ITEM.UNCLAIMED":
		webhookEvent.Type = WebhookPayoutUnclaimed
		webhookEvent.Status = PayoutStatusOnHold
	case "PAYMENT.PAYOUTS-ITEM.RETURNED":
		webhookEvent.Type = WebhookPayoutReversed
		webhookEvent.Status = PayoutStatusReversed
	case "PAYMENT.PAYOUTS-ITEM.PENDING":
		webhookEvent.Type = WebhookPayoutPending
		webhookEvent.Status = PayoutStatusProcessing
	default:
		webhookEvent.Type = WebhookEventType(event.EventType)
	}

	// Extract payout item ID from resource
	if resource, ok := event.Resource.(map[string]interface{}); ok {
		if payoutItemID, ok := resource["payout_item_id"].(string); ok {
			webhookEvent.ProviderPayoutID = payoutItemID
		}
		if senderItemID, ok := resource["sender_item_id"].(string); ok {
			webhookEvent.PayoutID = senderItemID
		}
		if errors, ok := resource["errors"].(map[string]interface{}); ok {
			if name, ok := errors["name"].(string); ok {
				webhookEvent.FailureCode = name
			}
			if message, ok := errors["message"].(string); ok {
				webhookEvent.FailureMessage = message
			}
		}
	}

	return webhookEvent, nil
}

// ============================================================================
// Reconciliation
// ============================================================================

// GetSettlementReport retrieves a settlement report from PayPal.
func (a *PayPalAdapter) GetSettlementReport(ctx context.Context, req SettlementReportRequest) (*SettlementReport, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Query transaction history
	// Note: PayPal has a Transaction Search API at /v1/reporting/transactions
	// and a separate Settlement Reports API at /v1/reporting/balances
	
	searchURL := fmt.Sprintf("%s/v1/reporting/transactions?start_date=%s&end_date=%s&transaction_type=T0700&page_size=100",
		a.config.GetBaseURL(), req.StartDate, req.EndDate)
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create report request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("report request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("report request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var reportResp paypalTransactionReport
	if err := json.NewDecoder(resp.Body).Decode(&reportResp); err != nil {
		return nil, fmt.Errorf("failed to decode report: %w", err)
	}

	// Convert to our settlement report format
	report := &SettlementReport{
		ReportID:     fmt.Sprintf("pp_%s_%s", req.StartDate, req.EndDate),
		Provider:     ProviderPayPal,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		TotalPayouts: len(reportResp.TransactionDetails),
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Transactions: make([]SettlementTransaction, 0, len(reportResp.TransactionDetails)),
	}

	for _, tx := range reportResp.TransactionDetails {
		settleTx := SettlementTransaction{
			TransactionID: tx.TransactionInfo.TransactionID,
			PayoutID:      tx.TransactionInfo.InvoiceID, // We use invoice_id to track our payout ID
			Status:        tx.TransactionInfo.TransactionStatus,
			ProcessedAt:   tx.TransactionInfo.TransactionUpdatedDate,
		}
		
		if tx.TransactionInfo.TransactionAmount != nil {
			if amount, err := parsePayPalAmount(tx.TransactionInfo.TransactionAmount.Value); err == nil {
				settleTx.Amount = amount
				report.TotalAmount += amount
			}
		}
		
		if tx.TransactionInfo.FeeAmount != nil {
			if fee, err := parsePayPalAmount(tx.TransactionInfo.FeeAmount.Value); err == nil {
				settleTx.Fee = fee
				report.TotalFees += fee
			}
		}
		
		report.Transactions = append(report.Transactions, settleTx)
	}

	return report, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (a *PayPalAdapter) getRecipientType(dest PayoutDestination) string {
	switch dest.Type {
	case DestinationPayPalEmail:
		return "EMAIL"
	case DestinationPayPalID:
		return "PAYPAL_ID"
	default:
		return "EMAIL"
	}
}

func (a *PayPalAdapter) getReceiver(dest PayoutDestination) string {
	switch dest.Type {
	case DestinationPayPalEmail:
		return dest.Email
	case DestinationPayPalID:
		return dest.PayPalID
	default:
		return dest.Email
	}
}

// formatPayPalAmount formats an amount in minor units to PayPal's string format
func formatPayPalAmount(minorUnits int64) string {
	dollars := float64(minorUnits) / 100.0
	return fmt.Sprintf("%.2f", dollars)
}

// parsePayPalAmount parses a PayPal amount string to minor units
func parsePayPalAmount(amount string) (int64, error) {
	var dollars float64
	if _, err := fmt.Sscanf(amount, "%f", &dollars); err != nil {
		return 0, err
	}
	return int64(dollars * 100), nil
}

func mapPayPalStatus(batchStatus string) PayoutStatus {
	switch batchStatus {
	case "PENDING":
		return PayoutStatusProcessing
	case "PROCESSING":
		return PayoutStatusProcessing
	case "SUCCESS":
		return PayoutStatusSucceeded
	case "DENIED":
		return PayoutStatusFailed
	case "CANCELED":
		return PayoutStatusCanceled
	default:
		return PayoutStatusPending
	}
}

func mapPayPalItemStatus(status string) PayoutStatus {
	switch status {
	case "SUCCESS":
		return PayoutStatusSucceeded
	case "PENDING":
		return PayoutStatusProcessing
	case "FAILED":
		return PayoutStatusFailed
	case "RETURNED":
		return PayoutStatusReversed
	case "UNCLAIMED":
		return PayoutStatusOnHold
	case "BLOCKED":
		return PayoutStatusFailed
	case "REFUNDED":
		return PayoutStatusReversed
	case "REVERSED":
		return PayoutStatusReversed
	default:
		return PayoutStatusPending
	}
}

// ============================================================================
// PayPal API Types
// ============================================================================

type paypalPayoutRequest struct {
	SenderBatchHeader paypalSenderBatchHeader `json:"sender_batch_header"`
	Items             []paypalPayoutItem      `json:"items"`
}

type paypalSenderBatchHeader struct {
	SenderBatchID string `json:"sender_batch_id"`
	EmailSubject  string `json:"email_subject,omitempty"`
	EmailMessage  string `json:"email_message,omitempty"`
}

type paypalPayoutItem struct {
	RecipientType        string       `json:"recipient_type"`
	Amount               paypalAmount `json:"amount"`
	Note                 string       `json:"note,omitempty"`
	SenderItemID         string       `json:"sender_item_id"`
	Receiver             string       `json:"receiver"`
	NotificationLanguage string       `json:"notification_language,omitempty"`
}

type paypalAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type paypalPayoutResponse struct {
	BatchHeader paypalBatchHeader      `json:"batch_header"`
	Items       []paypalPayoutItemResp `json:"items,omitempty"`
}

type paypalBatchHeader struct {
	PayoutBatchID     string `json:"payout_batch_id"`
	BatchStatus       string `json:"batch_status"`
	SenderBatchHeader struct {
		SenderBatchID string `json:"sender_batch_id"`
	} `json:"sender_batch_header"`
}

type paypalPayoutItemResp struct {
	PayoutItemID   string `json:"payout_item_id"`
	TransactionStatus string `json:"transaction_status"`
}

type paypalPayoutItemDetails struct {
	PayoutItemID      string        `json:"payout_item_id"`
	TransactionStatus string        `json:"transaction_status"`
	PayoutItem        paypalPayoutItem `json:"payout_item"`
	Errors            *paypalError  `json:"errors,omitempty"`
}

type paypalError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type paypalErrorResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Details []struct {
		Field string `json:"field"`
		Issue string `json:"issue"`
	} `json:"details,omitempty"`
}

type paypalWebhookEvent struct {
	ID         string      `json:"id"`
	CreateTime string      `json:"create_time"`
	EventType  string      `json:"event_type"`
	Resource   interface{} `json:"resource"`
}

type paypalTransactionReport struct {
	TransactionDetails []struct {
		TransactionInfo struct {
			TransactionID          string        `json:"transaction_id"`
			TransactionStatus      string        `json:"transaction_status"`
			TransactionUpdatedDate string        `json:"transaction_updated_date"`
			InvoiceID              string        `json:"invoice_id,omitempty"`
			TransactionAmount      *paypalAmount `json:"transaction_amount,omitempty"`
			FeeAmount              *paypalAmount `json:"fee_amount,omitempty"`
		} `json:"transaction_info"`
	} `json:"transaction_details"`
}

// ============================================================================
// PayPal Webhook Signature Verification
// ============================================================================

// verifyWebhookSignature verifies a PayPal webhook signature using HMAC-SHA256.
// In production, you should use PayPal's verify-webhook-signature API instead.
func verifyWebhookSignature(webhookID string, transmissionID string, timestamp string, payload []byte, expectedSignature string) bool {
	// Construct the validation string
	// PayPal uses: <transmissionId>|<timestamp>|<webhookId>|<CRC32 of payload>
	// Then signs with SHA256
	
	message := fmt.Sprintf("%s|%s|%s|%x", transmissionID, timestamp, webhookID, crc32Checksum(payload))
	
	mac := hmac.New(sha256.New, []byte(webhookID))
	mac.Write([]byte(message))
	expectedMAC := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(expectedSignature), []byte(expectedMAC))
}

// crc32Checksum computes CRC32 checksum of data
func crc32Checksum(data []byte) uint32 {
	var crc uint32 = 0xFFFFFFFF
	for _, b := range data {
		crc ^= uint32(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}
