// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3062: PayPal Commerce adapter for payment intents and refunds.
package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// PayPal Adapter
// ============================================================================

// PayPalAdapter implements the Gateway interface for PayPal Commerce Platform.
type PayPalAdapter struct {
	config     PayPalConfig
	httpClient *http.Client

	// OAuth token management
	tokenMu     sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

const paypalCaptureMethodManual = captureMethodManual

// NewPayPalAdapter creates a new PayPal adapter.
func NewPayPalAdapter(config PayPalConfig) (Gateway, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, ErrGatewayNotConfigured
	}

	return &PayPalAdapter{
		config:     config,
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
	}, nil
}

func (a *PayPalAdapter) Name() string {
	return "PayPal"
}

func (a *PayPalAdapter) Type() GatewayType {
	return GatewayPayPal
}

func (a *PayPalAdapter) IsHealthy(ctx context.Context) bool {
	_, err := a.getAccessToken(ctx)
	return err == nil
}

func (a *PayPalAdapter) Close() error {
	return nil
}

// ============================================================================
// Customer Management
// ============================================================================

func (a *PayPalAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	customerID := req.VEIDAddress
	if customerID == "" {
		customerID = fmt.Sprintf("cust_%d", time.Now().UnixNano())
	}

	return Customer{
		ID:          customerID,
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		VEIDAddress: req.VEIDAddress,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}, nil
}

func (a *PayPalAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	return Customer{ID: customerID}, nil
}

func (a *PayPalAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	customer := Customer{ID: customerID}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	return customer, nil
}

func (a *PayPalAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	return nil
}

// ============================================================================
// Payment Methods
// ============================================================================

func (a *PayPalAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}
	return token.Token, nil
}

func (a *PayPalAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return nil
}

func (a *PayPalAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	return nil, nil
}

// ============================================================================
// Payment Intents (PayPal Orders)
// ============================================================================

func (a *PayPalAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	intent := "CAPTURE"
	if req.CaptureMethod == paypalCaptureMethodManual {
		intent = "AUTHORIZE"
	}

	amountValue := formatPayPalAmount(req.Amount)
	orderReq := paypalOrderRequest{
		Intent: intent,
		PurchaseUnits: []paypalPurchaseUnit{
			{
				CustomID:    req.CustomerID,
				Description: req.Description,
				Amount: paypalAmount{
					CurrencyCode: string(req.Amount.Currency),
					Value:        amountValue,
				},
			},
		},
		ApplicationContext: paypalApplicationContext{
			ReturnURL: req.ReturnURL,
			CancelURL: req.ReturnURL,
			BrandName: req.StatementDescriptor,
		},
	}

	if req.ReceiptEmail != "" {
		orderReq.Payer = &paypalPayer{EmailAddress: req.ReceiptEmail}
	}

	respBody, err := a.doRequest(ctx, http.MethodPost, "/v2/checkout/orders", orderReq, req.IdempotencyKey)
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(respBody, &order); err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to parse PayPal order response: %w", err)
	}

	return a.mapPayPalOrder(order, req.Amount), nil
}

func (a *PayPalAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	respBody, err := a.doRequest(ctx, http.MethodGet, "/v2/checkout/orders/"+paymentIntentID, nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(respBody, &order); err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to parse PayPal order response: %w", err)
	}

	amount, err := parsePayPalAmount(order.PurchaseUnits)
	if err != nil {
		return PaymentIntent{}, err
	}

	return a.mapPayPalOrder(order, amount), nil
}

func (a *PayPalAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	respBody, err := a.doRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/capture", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var capture paypalOrderCaptureResponse
	if err := json.Unmarshal(respBody, &capture); err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to parse PayPal capture response: %w", err)
	}

	return a.mapPayPalCapture(capture), nil
}

func (a *PayPalAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	_, err := a.doRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/cancel", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	return PaymentIntent{
		ID:        paymentIntentID,
		Gateway:   GatewayPayPal,
		Status:    PaymentIntentStatusCanceled,
		UpdatedAt: time.Now(),
	}, nil
}

func (a *PayPalAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	payload := map[string]interface{}{}
	if amount != nil {
		payload["amount"] = paypalAmount{
			CurrencyCode: string(amount.Currency),
			Value:        formatPayPalAmount(*amount),
		}
	}

	respBody, err := a.doRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/capture", payload, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var capture paypalOrderCaptureResponse
	if err := json.Unmarshal(respBody, &capture); err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to parse PayPal capture response: %w", err)
	}

	return a.mapPayPalCapture(capture), nil
}

// ============================================================================
// Refunds
// ============================================================================

func (a *PayPalAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	captureID := req.PaymentIntentID
	if captureID == "" {
		return Refund{}, ErrPaymentIntentNotFound
	}

	payload := paypalRefundRequest{}
	if req.Amount != nil {
		payload.Amount = &paypalAmount{
			CurrencyCode: string(req.Amount.Currency),
			Value:        formatPayPalAmount(*req.Amount),
		}
	}

	respBody, err := a.doRequest(ctx, http.MethodPost, "/v2/payments/captures/"+captureID+"/refund", payload, req.IdempotencyKey)
	if err != nil {
		return Refund{}, err
	}

	var refundResp paypalRefundResponse
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return Refund{}, fmt.Errorf("failed to parse PayPal refund response: %w", err)
	}

	return mapPayPalRefund(refundResp), nil
}

func (a *PayPalAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	respBody, err := a.doRequest(ctx, http.MethodGet, "/v2/payments/refunds/"+refundID, nil, "")
	if err != nil {
		return Refund{}, err
	}

	var refundResp paypalRefundResponse
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return Refund{}, fmt.Errorf("failed to parse PayPal refund response: %w", err)
	}

	return mapPayPalRefund(refundResp), nil
}

// ============================================================================
// Webhooks
// ============================================================================

func (a *PayPalAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookID == "" {
		return nil
	}

	var sig paypalWebhookSignature
	if err := json.Unmarshal([]byte(signature), &sig); err != nil {
		return ErrWebhookSignatureInvalid
	}

	if sig.TransmissionID == "" || sig.TransmissionSig == "" || sig.TransmissionTime == "" {
		return ErrWebhookSignatureInvalid
	}

	accessToken, err := a.getAccessToken(context.Background())
	if err != nil {
		return err
	}

	verifyReq := paypalWebhookVerifyRequest{
		AuthAlgo:         sig.AuthAlgo,
		CertURL:          sig.CertURL,
		TransmissionID:   sig.TransmissionID,
		TransmissionSig:  sig.TransmissionSig,
		TransmissionTime: sig.TransmissionTime,
		WebhookID:        a.config.WebhookID,
		WebhookEvent:     json.RawMessage(payload),
	}

	body, err := json.Marshal(verifyReq)
	if err != nil {
		return fmt.Errorf("failed to encode PayPal webhook verification: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, a.config.GetBaseURL()+"/v1/notifications/verify-webhook-signature", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create PayPal webhook verification request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook verification failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrWebhookSignatureInvalid
	}

	var verifyResp struct {
		VerificationStatus string `json:"verification_status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return ErrWebhookSignatureInvalid
	}

	if strings.ToUpper(verifyResp.VerificationStatus) != "SUCCESS" {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func (a *PayPalAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event paypalWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("failed to parse PayPal webhook event: %w", err)
	}

	webhookEvent := WebhookEvent{
		ID:        event.ID,
		Gateway:   GatewayPayPal,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if t, err := time.Parse(time.RFC3339, event.CreateTime); err == nil {
		webhookEvent.Timestamp = t
	}

	switch event.EventType {
	case "CHECKOUT.ORDER.APPROVED", "CHECKOUT.ORDER.COMPLETED", "PAYMENT.CAPTURE.COMPLETED":
		webhookEvent.Type = WebhookEventPaymentIntentSucceeded
	case "CHECKOUT.ORDER.CANCELLED", "PAYMENT.CAPTURE.DENIED":
		webhookEvent.Type = WebhookEventPaymentIntentFailed
	case "PAYMENT.CAPTURE.REFUNDED", "PAYMENT.CAPTURE.REVERSED":
		webhookEvent.Type = WebhookEventChargeRefunded
	case "CUSTOMER.DISPUTE.CREATED":
		webhookEvent.Type = WebhookEventChargeDisputeCreated
	case "CUSTOMER.DISPUTE.UPDATED":
		webhookEvent.Type = WebhookEventChargeDisputeUpdated
	case "CUSTOMER.DISPUTE.RESOLVED":
		webhookEvent.Type = WebhookEventChargeDisputeClosed
	default:
		return WebhookEvent{}, ErrWebhookEventUnknown
	}

	return webhookEvent, nil
}

// ============================================================================
// PayPal OAuth and HTTP helpers
// ============================================================================

func (a *PayPalAdapter) getAccessToken(ctx context.Context) (string, error) {
	a.tokenMu.RLock()
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		token := a.accessToken
		a.tokenMu.RUnlock()
		return token, nil
	}
	a.tokenMu.RUnlock()

	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.accessToken, nil
	}

	data := "grant_type=client_credentials"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.config.GetBaseURL()+"/v1/oauth2/token", strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create PayPal token request: %w", err)
	}

	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("PayPal token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("PayPal token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse PayPal token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", ErrGatewayUnavailable
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return a.accessToken, nil
}

func (a *PayPalAdapter) doRequest(ctx context.Context, method, path string, body interface{}, idempotencyKey string) ([]byte, error) {
	accessToken, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode PayPal request: %w", err)
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.config.GetBaseURL()+path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create PayPal request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("PayPal-Request-Id", idempotencyKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PayPal request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PayPal response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("PayPal request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ============================================================================
// PayPal mapping helpers
// ============================================================================

type paypalOrderRequest struct {
	Intent             string                   `json:"intent"`
	PurchaseUnits      []paypalPurchaseUnit     `json:"purchase_units"`
	ApplicationContext paypalApplicationContext `json:"application_context,omitempty"`
	Payer              *paypalPayer             `json:"payer,omitempty"`
}

type paypalApplicationContext struct {
	ReturnURL string `json:"return_url,omitempty"`
	CancelURL string `json:"cancel_url,omitempty"`
	BrandName string `json:"brand_name,omitempty"`
}

type paypalPayer struct {
	EmailAddress string `json:"email_address,omitempty"`
}

type paypalPurchaseUnit struct {
	Amount      paypalAmount `json:"amount"`
	CustomID    string       `json:"custom_id,omitempty"`
	Description string       `json:"description,omitempty"`
}

type paypalAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type paypalOrderResponse struct {
	ID            string               `json:"id"`
	Status        string               `json:"status"`
	Links         []paypalLink         `json:"links,omitempty"`
	PurchaseUnits []paypalPurchaseUnit `json:"purchase_units,omitempty"`
}

type paypalOrderCaptureResponse struct {
	ID            string                   `json:"id"`
	Status        string                   `json:"status"`
	PurchaseUnits []paypalPurchaseUnitItem `json:"purchase_units,omitempty"`
}

type paypalPurchaseUnitItem struct {
	Payments paypalPayments `json:"payments"`
}

type paypalPayments struct {
	Captures []paypalCapture `json:"captures,omitempty"`
}

type paypalCapture struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Amount paypalAmount `json:"amount"`
}

type paypalLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type paypalRefundRequest struct {
	Amount *paypalAmount `json:"amount,omitempty"`
}

type paypalRefundResponse struct {
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Amount        paypalAmount `json:"amount"`
	CreateTime    string       `json:"create_time"`
	UpdateTime    string       `json:"update_time"`
	FailureReason string       `json:"reason_code"`
}

type paypalWebhookEvent struct {
	ID         string `json:"id"`
	EventType  string `json:"event_type"`
	CreateTime string `json:"create_time"`
}

type paypalWebhookSignature struct {
	TransmissionID   string `json:"transmission_id"`
	TransmissionTime string `json:"transmission_time"`
	CertURL          string `json:"cert_url"`
	AuthAlgo         string `json:"auth_algo"`
	TransmissionSig  string `json:"transmission_sig"`
	WebhookID        string `json:"webhook_id"`
}

type paypalWebhookVerifyRequest struct {
	AuthAlgo         string          `json:"auth_algo"`
	CertURL          string          `json:"cert_url"`
	TransmissionID   string          `json:"transmission_id"`
	TransmissionSig  string          `json:"transmission_sig"`
	TransmissionTime string          `json:"transmission_time"`
	WebhookID        string          `json:"webhook_id"`
	WebhookEvent     json.RawMessage `json:"webhook_event"`
}

func (a *PayPalAdapter) mapPayPalOrder(order paypalOrderResponse, amount Amount) PaymentIntent {
	status := mapPayPalOrderStatus(order.Status)
	redirectURL := ""
	for _, link := range order.Links {
		if link.Rel == "approve" {
			redirectURL = link.Href
			break
		}
	}

	intent := PaymentIntent{
		ID:             order.ID,
		Gateway:        GatewayPayPal,
		Amount:         amount,
		Status:         status,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		RequiresSCA:    redirectURL != "",
		SCARedirectURL: redirectURL,
	}

	if status == PaymentIntentStatusSucceeded {
		intent.CapturedAmount = amount
	}

	return intent
}

func (a *PayPalAdapter) mapPayPalCapture(capture paypalOrderCaptureResponse) PaymentIntent {
	amount := Amount{}
	captureID := ""
	if len(capture.PurchaseUnits) > 0 && len(capture.PurchaseUnits[0].Payments.Captures) > 0 {
		cap := capture.PurchaseUnits[0].Payments.Captures[0]
		captureID = cap.ID
		parsed, err := parsePayPalAmount([]paypalPurchaseUnit{{Amount: cap.Amount}})
		if err == nil {
			amount = parsed
		}
	}

	status := mapPayPalOrderStatus(capture.Status)
	intent := PaymentIntent{
		ID:              capture.ID,
		Gateway:         GatewayPayPal,
		Amount:          amount,
		Status:          status,
		CapturedAmount:  amount,
		PaymentMethodID: captureID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return intent
}

func mapPayPalRefund(refund paypalRefundResponse) Refund {
	amount := Amount{}
	if refund.Amount.CurrencyCode != "" {
		parsed, err := parsePayPalAmount([]paypalPurchaseUnit{{Amount: refund.Amount}})
		if err == nil {
			amount = parsed
		}
	}

	status := RefundStatusPending
	switch strings.ToUpper(refund.Status) {
	case "COMPLETED":
		status = RefundStatusSucceeded
	case "FAILED", "CANCELLED":
		status = RefundStatusFailed
	}

	createdAt := time.Now()
	if refund.CreateTime != "" {
		if t, err := time.Parse(time.RFC3339, refund.CreateTime); err == nil {
			createdAt = t
		}
	}

	return Refund{
		ID:            refund.ID,
		Amount:        amount,
		Status:        status,
		FailureReason: refund.FailureReason,
		CreatedAt:     createdAt,
	}
}

func mapPayPalOrderStatus(status string) PaymentIntentStatus {
	switch strings.ToUpper(status) {
	case "CREATED", "SAVED", "APPROVED", "PAYER_ACTION_REQUIRED":
		return PaymentIntentStatusRequiresAction
	case "COMPLETED":
		return PaymentIntentStatusSucceeded
	case "VOIDED", "CANCELLED":
		return PaymentIntentStatusCanceled
	case "FAILED", "DENIED":
		return PaymentIntentStatusFailed
	default:
		return PaymentIntentStatusProcessing
	}
}

func formatPayPalAmount(amount Amount) string {
	factor := amount.Currency.MinorUnitFactor()
	if factor <= 1 {
		return strconv.FormatInt(amount.Value, 10)
	}
	decimals := 0
	for factor > 1 {
		decimals++
		factor /= 10
	}
	value := float64(amount.Value) / float64(amount.Currency.MinorUnitFactor())
	return strconv.FormatFloat(value, 'f', decimals, 64)
}

func parsePayPalAmount(units []paypalPurchaseUnit) (Amount, error) {
	if len(units) == 0 {
		return Amount{}, ErrInvalidCurrency
	}

	currency := Currency(strings.ToUpper(units[0].Amount.CurrencyCode))
	if !currency.IsValid() {
		return Amount{}, ErrInvalidCurrency
	}

	minor, err := parsePayPalValueToMinor(units[0].Amount.Value, currency)
	if err != nil {
		return Amount{}, err
	}

	return Amount{Value: minor, Currency: currency}, nil
}

func parsePayPalValueToMinor(value string, currency Currency) (int64, error) {
	if value == "" {
		return 0, ErrInvalidCurrency
	}

	factor := currency.MinorUnitFactor()
	if factor == 1 {
		return strconv.ParseInt(value, 10, 64)
	}

	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	fractional := int64(0)
	decimals := int64(1)
	if len(parts) == 2 {
		fractionalStr := parts[1]
		for len(fractionalStr) < int64Digits(factor) {
			fractionalStr += "0"
		}
		if len(fractionalStr) > int64Digits(factor) {
			fractionalStr = fractionalStr[:int64Digits(factor)]
		}
		fractional, err = strconv.ParseInt(fractionalStr, 10, 64)
		if err != nil {
			return 0, err
		}
		for i := 0; i < int64Digits(factor); i++ {
			decimals *= 10
		}
	}

	return whole*factor + fractional*(factor/decimals), nil
}

func int64Digits(value int64) int {
	digits := 0
	for value > 1 {
		digits++
		value /= 10
	}
	if digits == 0 {
		return 1
	}
	return digits
}
