// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-49C: ACH direct debit adapter implementation.
package payment

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

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// ACH Adapter
// ============================================================================

// OFACScreener provides an integration point for OFAC screening.
type OFACScreener interface {
	Screen(ctx context.Context, account BankAccountDetails) error
}

// ACHAdapterOption configures the ACH adapter.
type ACHAdapterOption func(*ACHAdapter)

// WithOFACScreener sets the OFAC screener.
func WithOFACScreener(screener OFACScreener) ACHAdapterOption {
	return func(a *ACHAdapter) {
		a.ofacScreener = screener
	}
}

// ACHAdapter implements the Gateway interface for ACH debits.
type ACHAdapter struct {
	config        ACHConfig
	httpClient    *http.Client
	baseURL       string
	ofacScreener  OFACScreener
	retryMax      int
	retryDelay    time.Duration
	retryMaxDelay time.Duration
	retryFactor   float64
}

// NewACHAdapter creates a new ACH adapter.
func NewACHAdapter(config ACHConfig, opts ...ACHAdapterOption) (Gateway, error) {
	if config.SecretKey == "" {
		return nil, ErrGatewayNotConfigured
	}

	adapter := &ACHAdapter{
		config:        config,
		httpClient:    security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
		baseURL:       config.GetBaseURL(),
		retryMax:      config.RetryMaxAttempts,
		retryDelay:    config.RetryInitialDelay,
		retryMaxDelay: config.RetryMaxDelay,
		retryFactor:   config.RetryBackoffFactor,
	}

	if adapter.retryMax == 0 {
		adapter.retryMax = 3
	}
	if adapter.retryDelay == 0 {
		adapter.retryDelay = 200 * time.Millisecond
	}
	if adapter.retryMaxDelay == 0 {
		adapter.retryMaxDelay = 2 * time.Second
	}
	if adapter.retryFactor == 0 {
		adapter.retryFactor = 2.0
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter, nil
}

func (a *ACHAdapter) Name() string {
	return "ACH"
}

func (a *ACHAdapter) Type() GatewayType {
	return GatewayACH
}

func (a *ACHAdapter) IsHealthy(ctx context.Context) bool {
	_, err := a.doRequest(ctx, http.MethodGet, "/ach/health", nil, "")
	return err == nil
}

func (a *ACHAdapter) Close() error {
	return nil
}

// ============================================================================
// Customer Management
// ============================================================================

func (a *ACHAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	resp, err := a.doRequest(ctx, http.MethodPost, "/ach/customers", req, "")
	if err != nil {
		return Customer{}, err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return Customer{}, err
	}

	return Customer{
		ID:          result.ID,
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		VEIDAddress: req.VEIDAddress,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}, nil
}

func (a *ACHAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	resp, err := a.doRequest(ctx, http.MethodGet, "/ach/customers/"+customerID, nil, "")
	if err != nil {
		return Customer{}, err
	}

	var result Customer
	if err := json.Unmarshal(resp, &result); err != nil {
		return Customer{}, err
	}
	return result, nil
}

func (a *ACHAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	resp, err := a.doRequest(ctx, http.MethodPatch, "/ach/customers/"+customerID, req, "")
	if err != nil {
		return Customer{}, err
	}

	var result Customer
	if err := json.Unmarshal(resp, &result); err != nil {
		return Customer{}, err
	}
	return result, nil
}

func (a *ACHAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := a.doRequest(ctx, http.MethodDelete, "/ach/customers/"+customerID, nil, "")
	return err
}

// ============================================================================
// Payment Methods
// ============================================================================

func (a *ACHAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}
	return token.Token, nil
}

func (a *ACHAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := a.doRequest(ctx, http.MethodDelete, "/ach/payment_methods/"+paymentMethodID, nil, "")
	return err
}

func (a *ACHAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	resp, err := a.doRequest(ctx, http.MethodGet, "/ach/customers/"+customerID+"/payment_methods", nil, "")
	if err != nil {
		return nil, err
	}

	var result []CardToken
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ============================================================================
// Payment Intents
// ============================================================================

func (a *ACHAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	if req.PaymentMethodID == "" && req.BankAccount == nil {
		return PaymentIntent{}, ErrBankAccountRequired
	}

	if req.BankAccount != nil {
		if a.ofacScreener != nil {
			if err := a.ofacScreener.Screen(ctx, *req.BankAccount); err != nil {
				return PaymentIntent{}, err
			}
		}
		if err := a.verifyBankAccount(ctx, *req.BankAccount, req.BankVerificationMethod); err != nil {
			return PaymentIntent{}, err
		}
	}

	debitReq := achDebitRequest{
		Amount:          req.Amount.Value,
		Currency:        string(req.Amount.Currency),
		CustomerID:      req.CustomerID,
		PaymentMethodID: req.PaymentMethodID,
		Metadata:        req.Metadata,
	}
	if req.BankAccount != nil {
		debitReq.BankAccount = req.BankAccount
	}

	resp, err := a.doRequestWithRetry(ctx, http.MethodPost, "/ach/debits", debitReq, req.IdempotencyKey)
	if err != nil {
		return PaymentIntent{}, err
	}

	var result achDebitResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapAchDebitToIntent(result, req.Amount, req.CustomerID)
	return intent, nil
}

func (a *ACHAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	resp, err := a.doRequest(ctx, http.MethodGet, "/ach/debits/"+paymentIntentID, nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var result achDebitResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return PaymentIntent{}, err
	}

	return mapAchDebitToIntent(result, Amount{}, ""), nil
}

func (a *ACHAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	resp, err := a.doRequest(ctx, http.MethodPost, "/ach/debits/"+paymentIntentID+"/confirm", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var result achDebitResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapAchDebitToIntent(result, Amount{}, "")
	intent.PaymentMethodID = paymentMethodID
	return intent, nil
}

func (a *ACHAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	resp, err := a.doRequest(ctx, http.MethodPost, "/ach/debits/"+paymentIntentID+"/cancel", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var result achDebitResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return PaymentIntent{}, err
	}

	return mapAchDebitToIntent(result, Amount{}, ""), nil
}

func (a *ACHAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	req := achCaptureRequest{}
	if amount != nil {
		req.Amount = amount.Value
	}

	resp, err := a.doRequest(ctx, http.MethodPost, "/ach/debits/"+paymentIntentID+"/capture", req, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var result achDebitResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapAchDebitToIntent(result, Amount{}, "")
	if amount != nil {
		intent.CapturedAmount = *amount
	}
	return intent, nil
}

// ============================================================================
// Refunds
// ============================================================================

func (a *ACHAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	refundReq := achRefundRequest{
		PaymentIntentID: req.PaymentIntentID,
		Reason:          string(req.Reason),
		Metadata:        req.Metadata,
	}
	if req.Amount != nil {
		refundReq.Amount = req.Amount.Value
		refundReq.Currency = string(req.Amount.Currency)
	}

	resp, err := a.doRequestWithRetry(ctx, http.MethodPost, "/ach/refunds", refundReq, req.IdempotencyKey)
	if err != nil {
		return Refund{}, err
	}

	var result achRefundResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return Refund{}, err
	}

	refundAmount := Amount{}
	if req.Amount != nil {
		refundAmount = *req.Amount
	} else if result.Amount > 0 {
		refundAmount = NewAmount(result.Amount, Currency(result.Currency))
	}

	return Refund{
		ID:              result.ID,
		PaymentIntentID: result.PaymentIntentID,
		Amount:          refundAmount,
		Status:          mapAchRefundStatus(result.Status),
		Reason:          req.Reason,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
	}, nil
}

func (a *ACHAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	resp, err := a.doRequest(ctx, http.MethodGet, "/ach/refunds/"+refundID, nil, "")
	if err != nil {
		return Refund{}, err
	}

	var result achRefundResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return Refund{}, err
	}

	refundAmount := NewAmount(result.Amount, Currency(result.Currency))
	return Refund{
		ID:              result.ID,
		PaymentIntentID: result.PaymentIntentID,
		Amount:          refundAmount,
		Status:          mapAchRefundStatus(result.Status),
		CreatedAt:       time.Now(),
	}, nil
}

// ============================================================================
// Webhooks
// ============================================================================

func (a *ACHAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return nil
	}

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

	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(a.config.WebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func (a *ACHAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event achWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, err
	}

	webhookEvent := WebhookEvent{
		ID:        event.ID,
		Type:      mapAchWebhookType(event.Type),
		Gateway:   GatewayACH,
		Payload:   payload,
		Timestamp: time.Unix(event.Created, 0),
		Data:      event.Data,
	}
	return webhookEvent, nil
}

// ============================================================================
// NACHA File Generation
// ============================================================================

// ACHEntry represents an entry in a NACHA file.
type ACHEntry struct {
	TraceNumber       string
	TransactionCode   string
	RoutingNumber     string
	AccountNumber     string
	Amount            int64
	AccountHolderName string
	IndividualID      string
}

// BuildNACHAFile generates a NACHA formatted file.
func (a *ACHAdapter) BuildNACHAFile(entries []ACHEntry, effectiveDate time.Time) (string, error) {
	if a.config.NACHAOriginID == "" || a.config.NACHACompanyName == "" {
		return "", ErrGatewayNotConfigured
	}
	return generateNACHAFile(entries, a.config.NACHAOriginID, a.config.NACHACompanyName, effectiveDate), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (a *ACHAdapter) verifyBankAccount(ctx context.Context, account BankAccountDetails, method string) error {
	verifyMethod := strings.ToLower(method)
	if verifyMethod == "" {
		verifyMethod = "micro_deposit"
	}

	path := "/ach/verifications/micro_deposits"
	if verifyMethod == "instant" {
		path = "/ach/verifications/instant"
	}

	resp, err := a.doRequest(ctx, http.MethodPost, path, account, "")
	if err != nil {
		return err
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if !strings.EqualFold(result.Status, "verified") {
		return ErrBankAccountVerificationFailed
	}
	return nil
}

func (a *ACHAdapter) doRequestWithRetry(ctx context.Context, method, path string, body interface{}, idempotencyKey string) ([]byte, error) {
	attempt := 0
	delay := a.retryDelay
	for {
		attempt++
		resp, err := a.doRequest(ctx, method, path, body, idempotencyKey)
		if err == nil {
			return resp, nil
		}
		if attempt >= a.retryMax {
			return nil, err
		}
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * a.retryFactor)
		if delay > a.retryMaxDelay {
			delay = a.retryMaxDelay
		}
	}
}

func (a *ACHAdapter) doRequest(ctx context.Context, method, path string, body interface{}, idempotencyKey string) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, ErrGatewayUnavailable
	}

	return respBody, nil
}

func mapAchDebitToIntent(result achDebitResponse, amount Amount, customerID string) PaymentIntent {
	intent := PaymentIntent{
		ID:              result.ID,
		Gateway:         GatewayACH,
		Amount:          amount,
		Status:          mapAchStatus(result.Status),
		CustomerID:      customerID,
		PaymentMethodID: result.PaymentMethodID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if result.Amount > 0 {
		intent.Amount = NewAmount(result.Amount, Currency(result.Currency))
	}
	if intent.Status == PaymentIntentStatusSucceeded {
		intent.CapturedAmount = intent.Amount
	}
	if result.FailureCode != "" {
		intent.FailureCode = result.FailureCode
		intent.FailureMessage = result.FailureMessage
	}

	return intent
}

func mapAchStatus(status string) PaymentIntentStatus {
	switch strings.ToLower(status) {
	case "succeeded", "settled":
		return PaymentIntentStatusSucceeded
	case "failed":
		return PaymentIntentStatusFailed
	case "canceled":
		return PaymentIntentStatusCanceled
	case "requires_confirmation":
		return PaymentIntentStatusRequiresConfirmation
	default:
		return PaymentIntentStatusProcessing
	}
}

func mapAchRefundStatus(status string) RefundStatus {
	switch strings.ToLower(status) {
	case "succeeded":
		return RefundStatusSucceeded
	case "failed":
		return RefundStatusFailed
	case "canceled":
		return RefundStatusCanceled
	default:
		return RefundStatusPending
	}
}

func generateNACHAFile(entries []ACHEntry, originID, companyName string, effectiveDate time.Time) string {
	batchDate := effectiveDate.Format("060102")
	fileDate := time.Now().Format("060102")

	lines := make([]string, 0, len(entries)+4)
	lines = append(lines, fmt.Sprintf("1%09s%s%-23s%-6s", "094101", fileDate, padRight("VIRTENGINE", 23), "094101"))

	batchHeader := fmt.Sprintf("5%1s%-16s%-10s%-6s%-6s%-3s%-10s%-8s%-6s",
		"1", padRight(companyName, 16), padRight(originID, 10), "PPD", "SETTLE", batchDate, padRight(originID, 10), "1", "000000")
	lines = append(lines, batchHeader)

	entryHash := int64(0)
	totalDebit := int64(0)
	traceSeq := 1

	for _, entry := range entries {
		routing := padLeft(entry.RoutingNumber, 9, '0')
		account := padRight(entry.AccountNumber, 17)
		amount := padLeft(fmt.Sprintf("%d", entry.Amount), 10, '0')
		name := padRight(entry.AccountHolderName, 22)
		individual := padRight(entry.IndividualID, 15)
		trace := padLeft(fmt.Sprintf("%s%07d", originID, traceSeq), 15, '0')

		lines = append(lines, fmt.Sprintf("6%-2s%s%s%s%s%s%s",
			entry.TransactionCode, routing, account, amount, individual, name, trace))
		traceSeq++
		entryHash += parseRoutingForHash(entry.RoutingNumber)
		totalDebit += entry.Amount
	}

	control := fmt.Sprintf("8%06d%010d%012d%012d%-10s%25s",
		len(entries), entryHash%10000000000, totalDebit, 0, originID, "")
	lines = append(lines, control)

	fileControl := fmt.Sprintf("9%06d%06d%08d%012d%012d%39s",
		1, len(entries), entryHash%100000000, totalDebit, 0, "")
	lines = append(lines, fileControl)

	return strings.Join(lines, "\n")
}

func padRight(value string, length int) string {
	if len(value) >= length {
		return value[:length]
	}
	return value + strings.Repeat(" ", length-len(value))
}

func padLeft(value string, length int, pad rune) string {
	if len(value) >= length {
		return value[len(value)-length:]
	}
	return strings.Repeat(string(pad), length-len(value)) + value
}

func parseRoutingForHash(routing string) int64 {
	if len(routing) > 8 {
		routing = routing[:8]
	}
	var hash int64
	for _, r := range routing {
		if r >= '0' && r <= '9' {
			hash = hash*10 + int64(r-'0')
		}
	}
	return hash
}

// ============================================================================
// ACH API Types
// ============================================================================

type achDebitRequest struct {
	Amount          int64               `json:"amount"`
	Currency        string              `json:"currency"`
	CustomerID      string              `json:"customer_id,omitempty"`
	PaymentMethodID string              `json:"payment_method_id,omitempty"`
	BankAccount     *BankAccountDetails `json:"bank_account,omitempty"`
	Metadata        map[string]string   `json:"metadata,omitempty"`
}

type achDebitResponse struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	PaymentMethodID string `json:"payment_method_id"`
	FailureCode     string `json:"failure_code,omitempty"`
	FailureMessage  string `json:"failure_message,omitempty"`
}

type achCaptureRequest struct {
	Amount int64 `json:"amount,omitempty"`
}

type achRefundRequest struct {
	PaymentIntentID string            `json:"payment_intent_id"`
	Amount          int64             `json:"amount,omitempty"`
	Currency        string            `json:"currency,omitempty"`
	Reason          string            `json:"reason,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type achRefundResponse struct {
	ID              string `json:"id"`
	PaymentIntentID string `json:"payment_intent_id"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
}

type achWebhookEvent struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Created int64       `json:"created"`
	Data    interface{} `json:"data"`
}

func mapAchWebhookType(eventType string) WebhookEventType {
	switch eventType {
	case "ach.debit.succeeded":
		return WebhookEventPaymentIntentSucceeded
	case "ach.debit.failed":
		return WebhookEventPaymentIntentFailed
	case "ach.debit.canceled":
		return WebhookEventPaymentIntentCanceled
	case "ach.debit.processing":
		return WebhookEventPaymentIntentProcessing
	case "ach.refund.succeeded":
		return WebhookEventChargeRefunded
	default:
		return WebhookEventType(eventType)
	}
}

var _ = bytes.Buffer{}
