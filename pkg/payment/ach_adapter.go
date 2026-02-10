// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3063: ACH direct debit adapter for settlement.
package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v80/webhook"
	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// ACH Adapter
// ============================================================================

// OFACScreeningHook is an integration point for OFAC checks.
// Return an error to block a payment.
type OFACScreeningHook func(ctx context.Context, customerID, paymentMethodID string) error

// ACHAdapter implements the Gateway interface for ACH direct debit.
// This uses Stripe's ACH APIs for verification and collection by default.
type ACHAdapter struct {
	config     ACHConfig
	httpClient *http.Client
	baseURL    string

	ofacHook OFACScreeningHook
}

// NewACHAdapter creates a new ACH adapter.
func NewACHAdapter(config ACHConfig) (Gateway, error) {
	if config.SecretKey == "" {
		return nil, ErrGatewayNotConfigured
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

// SetOFACScreeningHook sets the OFAC screening hook.
func (a *ACHAdapter) SetOFACScreeningHook(hook OFACScreeningHook) {
	a.ofacHook = hook
}

func (a *ACHAdapter) Name() string {
	return "ACH"
}

func (a *ACHAdapter) Type() GatewayType {
	return GatewayACH
}

func (a *ACHAdapter) IsHealthy(ctx context.Context) bool {
	resp, err := a.doRequest(ctx, http.MethodGet, "/account", "", "")
	return err == nil && len(resp) > 0
}

func (a *ACHAdapter) Close() error {
	return nil
}

// ============================================================================
// Customer Management
// ============================================================================

func (a *ACHAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	values := url.Values{}
	if req.Email != "" {
		values.Set("email", req.Email)
	}
	if req.Name != "" {
		values.Set("name", req.Name)
	}
	if req.Phone != "" {
		values.Set("phone", req.Phone)
	}
	if req.VEIDAddress != "" {
		values.Set("metadata[veid_address]", req.VEIDAddress)
	}
	for k, v := range req.Metadata {
		values.Set(fmt.Sprintf("metadata[%s]", k), v)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/customers", values.Encode(), "")
	if err != nil {
		return Customer{}, err
	}

	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return Customer{}, fmt.Errorf("failed to parse customer response: %w", err)
	}

	return Customer{
		ID:        resp.ID,
		Email:     resp.Email,
		Name:      resp.Name,
		Phone:     resp.Phone,
		CreatedAt: time.Now(),
	}, nil
}

func (a *ACHAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	body, err := a.doRequest(ctx, http.MethodGet, "/customers/"+customerID, "", "")
	if err != nil {
		return Customer{}, err
	}

	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return Customer{}, fmt.Errorf("failed to parse customer response: %w", err)
	}

	return Customer{
		ID:        resp.ID,
		Email:     resp.Email,
		Name:      resp.Name,
		Phone:     resp.Phone,
		CreatedAt: time.Now(),
	}, nil
}

func (a *ACHAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	values := url.Values{}
	if req.Email != nil {
		values.Set("email", *req.Email)
	}
	if req.Name != nil {
		values.Set("name", *req.Name)
	}
	if req.Phone != nil {
		values.Set("phone", *req.Phone)
	}
	for k, v := range req.Metadata {
		values.Set(fmt.Sprintf("metadata[%s]", k), v)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/customers/"+customerID, values.Encode(), "")
	if err != nil {
		return Customer{}, err
	}

	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return Customer{}, fmt.Errorf("failed to parse customer response: %w", err)
	}

	return Customer{
		ID:        resp.ID,
		Email:     resp.Email,
		Name:      resp.Name,
		Phone:     resp.Phone,
		CreatedAt: time.Now(),
	}, nil
}

func (a *ACHAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := a.doRequest(ctx, http.MethodDelete, "/customers/"+customerID, "", "")
	return err
}

// ============================================================================
// Payment Methods
// ============================================================================

func (a *ACHAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}

	values := url.Values{}
	values.Set("customer", customerID)

	_, err := a.doRequest(ctx, http.MethodPost, "/payment_methods/"+token.Token+"/attach", values.Encode(), "")
	if err != nil {
		return "", err
	}
	return token.Token, nil
}

func (a *ACHAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := a.doRequest(ctx, http.MethodPost, "/payment_methods/"+paymentMethodID+"/detach", "", "")
	return err
}

func (a *ACHAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	query := "/payment_methods?customer=" + url.QueryEscape(customerID) + "&type=us_bank_account"
	body, err := a.doRequest(ctx, http.MethodGet, query, "", "")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			ID            string `json:"id"`
			USBankAccount struct {
				BankName string `json:"bank_name"`
				Last4    string `json:"last4"`
			} `json:"us_bank_account"`
			Created int64 `json:"created"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse payment methods: %w", err)
	}

	methods := make([]CardToken, 0, len(resp.Data))
	for _, item := range resp.Data {
		methods = append(methods, CardToken{
			Token:     item.ID,
			Gateway:   GatewayACH,
			Last4:     item.USBankAccount.Last4,
			CreatedAt: time.Unix(item.Created, 0),
		})
	}

	return methods, nil
}

// ============================================================================
// Payment Intents (Stripe ACH)
// ============================================================================

func (a *ACHAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	if a.ofacHook != nil && req.PaymentMethodID != "" {
		if err := a.ofacHook(ctx, req.CustomerID, req.PaymentMethodID); err != nil {
			return PaymentIntent{}, err
		}
	}

	values := url.Values{}
	values.Set("amount", strconv.FormatInt(req.Amount.Value, 10))
	values.Set("currency", strings.ToLower(string(req.Amount.Currency)))
	values.Add("payment_method_types[]", "us_bank_account")

	if req.CustomerID != "" {
		values.Set("customer", req.CustomerID)
	}
	if req.PaymentMethodID != "" {
		values.Set("payment_method", req.PaymentMethodID)
	}
	if req.Description != "" {
		values.Set("description", req.Description)
	}
	if req.ReceiptEmail != "" {
		values.Set("receipt_email", req.ReceiptEmail)
	}
	if req.CaptureMethod == "manual" {
		values.Set("capture_method", "manual")
	} else {
		values.Set("capture_method", "automatic")
	}
	for k, v := range req.Metadata {
		values.Set(fmt.Sprintf("metadata[%s]", k), v)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/payment_intents", values.Encode(), req.IdempotencyKey)
	if err != nil {
		return PaymentIntent{}, err
	}

	return mapStripeLikePaymentIntent(body), nil
}

func (a *ACHAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	body, err := a.doRequest(ctx, http.MethodGet, "/payment_intents/"+paymentIntentID, "", "")
	if err != nil {
		return PaymentIntent{}, err
	}

	return mapStripeLikePaymentIntent(body), nil
}

func (a *ACHAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	values := url.Values{}
	if paymentMethodID != "" {
		values.Set("payment_method", paymentMethodID)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/payment_intents/"+paymentIntentID+"/confirm", values.Encode(), "")
	if err != nil {
		return PaymentIntent{}, err
	}

	return mapStripeLikePaymentIntent(body), nil
}

func (a *ACHAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	values := url.Values{}
	if reason != "" {
		values.Set("cancellation_reason", reason)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/payment_intents/"+paymentIntentID+"/cancel", values.Encode(), "")
	if err != nil {
		return PaymentIntent{}, err
	}

	return mapStripeLikePaymentIntent(body), nil
}

func (a *ACHAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	values := url.Values{}
	if amount != nil {
		values.Set("amount_to_capture", strconv.FormatInt(amount.Value, 10))
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/payment_intents/"+paymentIntentID+"/capture", values.Encode(), "")
	if err != nil {
		return PaymentIntent{}, err
	}

	return mapStripeLikePaymentIntent(body), nil
}

// ============================================================================
// Refunds
// ============================================================================

func (a *ACHAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	values := url.Values{}
	values.Set("payment_intent", req.PaymentIntentID)
	if req.Amount != nil {
		values.Set("amount", strconv.FormatInt(req.Amount.Value, 10))
	}
	for k, v := range req.Metadata {
		values.Set(fmt.Sprintf("metadata[%s]", k), v)
	}

	body, err := a.doRequest(ctx, http.MethodPost, "/refunds", values.Encode(), req.IdempotencyKey)
	if err != nil {
		return Refund{}, err
	}

	return mapStripeLikeRefund(body), nil
}

func (a *ACHAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	body, err := a.doRequest(ctx, http.MethodGet, "/refunds/"+refundID, "", "")
	if err != nil {
		return Refund{}, err
	}

	return mapStripeLikeRefund(body), nil
}

// ============================================================================
// Webhooks
// ============================================================================

func (a *ACHAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return nil
	}

	_, err := webhook.ConstructEvent(payload, signature, a.config.WebhookSecret)
	if err != nil {
		return ErrWebhookSignatureInvalid
	}
	return nil
}

func (a *ACHAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Created    int64  `json:"created"`
		APIVersion string `json:"api_version"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("failed to parse ACH webhook: %w", err)
	}

	webhookEvent := WebhookEvent{
		ID:         event.ID,
		Gateway:    GatewayACH,
		Payload:    payload,
		Timestamp:  time.Unix(event.Created, 0),
		APIVersion: event.APIVersion,
	}

	switch event.Type {
	case "payment_intent.succeeded":
		webhookEvent.Type = WebhookEventPaymentIntentSucceeded
	case "payment_intent.payment_failed":
		webhookEvent.Type = WebhookEventPaymentIntentFailed
	case "payment_intent.processing":
		webhookEvent.Type = WebhookEventPaymentIntentProcessing
	case "payment_intent.canceled":
		webhookEvent.Type = WebhookEventPaymentIntentCanceled
	case "charge.refunded":
		webhookEvent.Type = WebhookEventChargeRefunded
	case "charge.dispute.created":
		webhookEvent.Type = WebhookEventChargeDisputeCreated
	case "charge.dispute.updated":
		webhookEvent.Type = WebhookEventChargeDisputeUpdated
	case "charge.dispute.closed":
		webhookEvent.Type = WebhookEventChargeDisputeClosed
	default:
		return WebhookEvent{}, ErrWebhookEventUnknown
	}

	return webhookEvent, nil
}

// ============================================================================
// ACH Verification & Compliance Helpers
// ============================================================================

// VerifyBankAccountMicroDeposits verifies a bank account using micro-deposit amounts.
func (a *ACHAdapter) VerifyBankAccountMicroDeposits(ctx context.Context, paymentMethodID string, amounts []int64) error {
	if len(amounts) == 0 {
		return ErrInvalidAmount
	}

	values := url.Values{}
	for _, amount := range amounts {
		values.Add("amounts[]", strconv.FormatInt(amount, 10))
	}

	_, err := a.doRequest(ctx, http.MethodPost, "/payment_methods/"+paymentMethodID+"/verify", values.Encode(), "")
	return err
}

// RetryPaymentIntent retries confirmation with exponential backoff.
func (a *ACHAdapter) RetryPaymentIntent(ctx context.Context, paymentIntentID string, maxAttempts int, initialDelay time.Duration) (PaymentIntent, error) {
	delay := initialDelay
	if delay <= 0 {
		delay = 2 * time.Second
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		intent, err := a.ConfirmPaymentIntent(ctx, paymentIntentID, "")
		if err == nil {
			return intent, nil
		}
		lastErr = err
		time.Sleep(delay)
		delay *= 2
	}

	if lastErr == nil {
		lastErr = ErrGatewayUnavailable
	}
	return PaymentIntent{}, lastErr
}

// ============================================================================
// NACHA File Generation (bulk settlement)
// ============================================================================

// ACHEntry represents a single NACHA entry detail.
type ACHEntry struct {
	TraceNumber     string
	RoutingNumber   string
	AccountNumber   string
	AccountName     string
	Amount          Amount
	TransactionCode string
}

// BuildNACHAFile creates a basic NACHA file for bulk settlements.
func BuildNACHAFile(companyName, companyID, immediateOrigin, immediateDestination string, entries []ACHEntry, createdAt time.Time) (string, error) {
	if companyName == "" || companyID == "" || immediateOrigin == "" || immediateDestination == "" {
		return "", ErrInvalidAmount
	}

	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	header := fmt.Sprintf("101 %-10s%-10s%s%s094101%s\n",
		immediateDestination,
		immediateOrigin,
		createdAt.Format("060102"),
		createdAt.Format("1504"),
		strings.ToUpper(companyName)[:minLen(len(companyName), 23)],
	)

	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, header)

	for _, entry := range entries {
		amount := fmt.Sprintf("%010d", entry.Amount.Value)
		line := fmt.Sprintf("622%-9s%-17s%s%-22s%s\n",
			entry.RoutingNumber,
			entry.AccountNumber,
			amount,
			padRight(entry.AccountName, 22),
			entry.TraceNumber,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, ""), nil
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value[:width]
	}
	return value + strings.Repeat(" ", width-len(value))
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// Internal helpers
// ============================================================================

func (a *ACHAdapter) doRequest(ctx context.Context, method, path, body, idempotencyKey string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create ACH request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.SecretKey)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ACH request failed: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACH response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ACH request failed with status %d: %s", resp.StatusCode, string(payload))
	}

	return payload, nil
}

func mapStripeLikePaymentIntent(payload []byte) PaymentIntent {
	var resp struct {
		ID             string            `json:"id"`
		Amount         int64             `json:"amount"`
		Currency       string            `json:"currency"`
		Status         string            `json:"status"`
		Customer       string            `json:"customer"`
		PaymentMethod  string            `json:"payment_method"`
		Description    string            `json:"description"`
		ClientSecret   string            `json:"client_secret"`
		AmountReceived int64             `json:"amount_received"`
		AmountRefunded int64             `json:"amount_refunded"`
		Metadata       map[string]string `json:"metadata"`
		Created        int64             `json:"created"`
	}

	_ = json.Unmarshal(payload, &resp)

	currency := Currency(strings.ToUpper(resp.Currency))
	amount := Amount{Value: resp.Amount, Currency: currency}
	captured := Amount{Value: resp.AmountReceived, Currency: currency}
	refunded := Amount{Value: resp.AmountRefunded, Currency: currency}

	if resp.AmountReceived == 0 && mapStripeStatus(resp.Status) == PaymentIntentStatusSucceeded {
		captured = amount
	}

	return PaymentIntent{
		ID:              resp.ID,
		Gateway:         GatewayACH,
		Amount:          amount,
		Status:          mapStripeStatus(resp.Status),
		CustomerID:      resp.Customer,
		PaymentMethodID: resp.PaymentMethod,
		Description:     resp.Description,
		ClientSecret:    resp.ClientSecret,
		CapturedAmount:  captured,
		RefundedAmount:  refunded,
		Metadata:        resp.Metadata,
		CreatedAt:       time.Unix(resp.Created, 0),
		UpdatedAt:       time.Now(),
	}
}

func mapStripeStatus(status string) PaymentIntentStatus {
	switch status {
	case "requires_payment_method":
		return PaymentIntentStatusRequiresPaymentMethod
	case "requires_confirmation":
		return PaymentIntentStatusRequiresConfirmation
	case "requires_action":
		return PaymentIntentStatusRequiresAction
	case "processing":
		return PaymentIntentStatusProcessing
	case "succeeded":
		return PaymentIntentStatusSucceeded
	case "canceled":
		return PaymentIntentStatusCanceled
	default:
		return PaymentIntentStatusFailed
	}
}

func mapStripeLikeRefund(payload []byte) Refund {
	var resp struct {
		ID            string `json:"id"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
		FailureReason string `json:"failure_reason"`
		Created       int64  `json:"created"`
	}

	_ = json.Unmarshal(payload, &resp)

	status := RefundStatusPending
	switch resp.Status {
	case "succeeded":
		status = RefundStatusSucceeded
	case "failed", "canceled":
		status = RefundStatusFailed
	}

	return Refund{
		ID:            resp.ID,
		Amount:        Amount{Value: resp.Amount, Currency: Currency(strings.ToUpper(resp.Currency))},
		Status:        status,
		FailureReason: resp.FailureReason,
		CreatedAt:     time.Unix(resp.Created, 0),
	}
}
