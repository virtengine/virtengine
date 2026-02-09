// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-49C: PayPal Commerce adapter implementation.
package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// PayPal Commerce Adapter
// ============================================================================

// PayPalAdapter implements the Gateway interface using PayPal REST APIs.
type PayPalAdapter struct {
	config     PayPalConfig
	httpClient *http.Client

	tokenMu     sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NewPayPalAdapter creates a new PayPal Commerce adapter.
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
// OAuth Token Management
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

	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", a.config.GetBaseURL())
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paypal token request failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return a.accessToken, nil
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
	if req.DefaultPaymentMethodID != nil {
		customer.DefaultPaymentMethodID = *req.DefaultPaymentMethodID
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
	return []CardToken{}, nil
}

// ============================================================================
// Payment Intents (PayPal Orders API)
// ============================================================================

func (a *PayPalAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	purchaseUnits := []paypalPurchaseUnit{
		{
			ReferenceID: req.IdempotencyKey,
			Description: req.Description,
			Amount: paypalAmount{
				CurrencyCode: string(req.Amount.Currency),
				Value:        formatPayPalAmount(req.Amount),
			},
		},
	}

	orderReq := paypalOrderRequest{
		Intent:        "CAPTURE",
		PurchaseUnits: purchaseUnits,
	}

	if req.ReturnURL != "" {
		orderReq.ApplicationContext = &paypalApplicationContext{
			ReturnURL: req.ReturnURL,
			CancelURL: req.ReturnURL + "?cancel=true",
		}
	}

	if req.PaymentMethodType != "" {
		orderReq.PaymentSource = map[string]map[string]string{
			strings.ToLower(req.PaymentMethodType): {},
		}
	}

	body, err := a.doPayPalRequest(ctx, http.MethodPost, "/v2/checkout/orders", orderReq, req.IdempotencyKey)
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapPayPalOrder(order, req.Amount, req.CustomerID)
	return intent, nil
}

func (a *PayPalAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	body, err := a.doPayPalRequest(ctx, http.MethodGet, "/v2/checkout/orders/"+paymentIntentID, nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapPayPalOrder(order, Amount{}, "")
	return intent, nil
}

func (a *PayPalAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	body, err := a.doPayPalRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/capture", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapPayPalOrder(order, Amount{}, "")
	intent.PaymentMethodID = paymentMethodID
	return intent, nil
}

func (a *PayPalAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	_, err := a.doPayPalRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/cancel", nil, "")
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
	body, err := a.doPayPalRequest(ctx, http.MethodPost, "/v2/checkout/orders/"+paymentIntentID+"/capture", nil, "")
	if err != nil {
		return PaymentIntent{}, err
	}

	var order paypalOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return PaymentIntent{}, err
	}

	intent := mapPayPalOrder(order, Amount{}, "")
	if amount != nil {
		intent.CapturedAmount = *amount
	}
	return intent, nil
}

// ============================================================================
// Refunds
// ============================================================================

func (a *PayPalAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	refundReq := paypalRefundRequest{}
	if req.Amount != nil {
		refundReq.Amount = &paypalAmount{
			CurrencyCode: string(req.Amount.Currency),
			Value:        formatPayPalAmount(*req.Amount),
		}
	}

	body, err := a.doPayPalRequest(ctx, http.MethodPost, "/v2/payments/captures/"+req.PaymentIntentID+"/refund", refundReq, req.IdempotencyKey)
	if err != nil {
		return Refund{}, err
	}

	var refundResp paypalRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return Refund{}, err
	}

	refundAmount := Amount{}
	if refundResp.Amount != nil {
		parsed, err := parsePayPalAmount(refundResp.Amount.Value, refundResp.Amount.CurrencyCode)
		if err == nil {
			refundAmount = parsed
		}
	}
	if req.Amount != nil {
		refundAmount = *req.Amount
	}

	return Refund{
		ID:              refundResp.ID,
		PaymentIntentID: req.PaymentIntentID,
		Amount:          refundAmount,
		Status:          mapPayPalRefundStatus(refundResp.Status),
		Reason:          req.Reason,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
	}, nil
}

func (a *PayPalAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	body, err := a.doPayPalRequest(ctx, http.MethodGet, "/v2/payments/refunds/"+refundID, nil, "")
	if err != nil {
		return Refund{}, err
	}

	var refundResp paypalRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return Refund{}, err
	}

	refundAmount := Amount{}
	if refundResp.Amount != nil {
		parsed, err := parsePayPalAmount(refundResp.Amount.Value, refundResp.Amount.CurrencyCode)
		if err == nil {
			refundAmount = parsed
		}
	}

	return Refund{
		ID:        refundResp.ID,
		Status:    mapPayPalRefundStatus(refundResp.Status),
		Amount:    refundAmount,
		CreatedAt: time.Now(),
	}, nil
}

// ============================================================================
// Webhooks
// ============================================================================

func (a *PayPalAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookID == "" {
		return nil
	}

	var headers paypalWebhookHeaders
	if err := json.Unmarshal([]byte(signature), &headers); err != nil {
		return ErrWebhookSignatureInvalid
	}
	if headers.TransmissionID == "" || headers.TransmissionSig == "" {
		return ErrWebhookSignatureInvalid
	}

	verifyReq := paypalWebhookVerificationRequest{
		TransmissionID:   headers.TransmissionID,
		TransmissionTime: headers.TransmissionTime,
		TransmissionSig:  headers.TransmissionSig,
		CertURL:          headers.CertURL,
		AuthAlgo:         headers.AuthAlgo,
		WebhookID:        a.config.WebhookID,
		WebhookEvent:     json.RawMessage(payload),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body, err := a.doPayPalRequest(ctx, http.MethodPost, "/v1/notifications/verify-webhook-signature", verifyReq, "")
	if err != nil {
		return ErrWebhookSignatureInvalid
	}

	var verifyResp struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return ErrWebhookSignatureInvalid
	}
	if verifyResp.VerificationStatus != "SUCCESS" {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func (a *PayPalAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event paypalWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, err
	}

	webhookEvent := WebhookEvent{
		ID:      event.ID,
		Type:    mapPayPalEventType(event.EventType),
		Gateway: GatewayPayPal,
		Payload: payload,
	}

	if t, err := time.Parse(time.RFC3339, event.CreateTime); err == nil {
		webhookEvent.Timestamp = t
	} else {
		webhookEvent.Timestamp = time.Now()
	}

	if resource, ok := event.Resource.(map[string]interface{}); ok {
		webhookEvent.Data = map[string]interface{}{
			"object": resource,
		}
	}

	return webhookEvent, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (a *PayPalAdapter) doPayPalRequest(ctx context.Context, method, path string, body interface{}, idempotencyKey string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, ErrGatewayUnavailable
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.config.GetBaseURL()+path, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("PayPal-Request-Id", idempotencyKey)
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

func mapPayPalOrder(order paypalOrderResponse, amount Amount, customerID string) PaymentIntent {
	intent := PaymentIntent{
		ID:         order.ID,
		Gateway:    GatewayPayPal,
		Amount:     amount,
		Status:     mapPayPalOrderStatus(order.Status),
		CustomerID: customerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if len(order.PurchaseUnits) > 0 {
		unit := order.PurchaseUnits[0]
		if unit.Amount != nil {
			if parsed, err := parsePayPalAmount(unit.Amount.Value, unit.Amount.CurrencyCode); err == nil {
				intent.Amount = parsed
			}
		}
	}

	if order.Status == "CREATED" || order.Status == "PAYER_ACTION_REQUIRED" {
		intent.RequiresSCA = true
		if link := findPayPalLink(order.Links, "approve"); link != "" {
			intent.SCARedirectURL = link
		}
	}

	if capture := findPayPalCapture(order.PurchaseUnits); capture != nil {
		if parsed, err := parsePayPalAmount(capture.Amount.Value, capture.Amount.CurrencyCode); err == nil {
			intent.CapturedAmount = parsed
		}
	}

	return intent
}

func mapPayPalOrderStatus(status string) PaymentIntentStatus {
	switch status {
	case "CREATED", "PAYER_ACTION_REQUIRED":
		return PaymentIntentStatusRequiresAction
	case "APPROVED":
		return PaymentIntentStatusRequiresConfirmation
	case "COMPLETED":
		return PaymentIntentStatusSucceeded
	case "CANCELLED", "VOIDED":
		return PaymentIntentStatusCanceled
	case "FAILED", "DECLINED":
		return PaymentIntentStatusFailed
	default:
		return PaymentIntentStatusProcessing
	}
}

func mapPayPalRefundStatus(status string) RefundStatus {
	switch status {
	case "COMPLETED":
		return RefundStatusSucceeded
	case "PENDING":
		return RefundStatusPending
	case "FAILED", "CANCELLED":
		return RefundStatusFailed
	default:
		return RefundStatusPending
	}
}

func formatPayPalAmount(amount Amount) string {
	major := amount.Major()
	if amount.Currency.MinorUnitFactor() == 1 {
		return fmt.Sprintf("%.0f", major)
	}
	return fmt.Sprintf("%.2f", major)
}

func parsePayPalAmount(value string, currency string) (Amount, error) {
	var major float64
	if _, err := fmt.Sscanf(value, "%f", &major); err != nil {
		return Amount{}, err
	}
	cur := Currency(strings.ToUpper(currency))
	factor := cur.MinorUnitFactor()
	minor := int64(major * float64(factor))
	return NewAmount(minor, cur), nil
}

func mapPayPalEventType(eventType string) WebhookEventType {
	switch eventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		return WebhookEventPaymentIntentSucceeded
	case "PAYMENT.CAPTURE.DENIED", "PAYMENT.CAPTURE.FAILED":
		return WebhookEventPaymentIntentFailed
	case "PAYMENT.CAPTURE.REFUNDED":
		return WebhookEventChargeRefunded
	case "CHECKOUT.ORDER.CANCELLED":
		return WebhookEventPaymentIntentCanceled
	case "CHECKOUT.ORDER.APPROVED":
		return WebhookEventPaymentIntentProcessing
	default:
		return WebhookEventType(eventType)
	}
}

func findPayPalLink(links []paypalLink, rel string) string {
	for _, link := range links {
		if strings.EqualFold(link.Rel, rel) {
			return link.Href
		}
	}
	return ""
}

func findPayPalCapture(units []paypalPurchaseUnitResponse) *paypalCapture {
	if len(units) == 0 {
		return nil
	}
	unit := units[0]
	if unit.Payments == nil || len(unit.Payments.Captures) == 0 {
		return nil
	}
	return &unit.Payments.Captures[0]
}

// ============================================================================
// PayPal API Types
// ============================================================================

type paypalOrderRequest struct {
	Intent             string                       `json:"intent"`
	PurchaseUnits      []paypalPurchaseUnit         `json:"purchase_units"`
	ApplicationContext *paypalApplicationContext    `json:"application_context,omitempty"`
	PaymentSource      map[string]map[string]string `json:"payment_source,omitempty"`
}

type paypalPurchaseUnit struct {
	ReferenceID string       `json:"reference_id,omitempty"`
	Description string       `json:"description,omitempty"`
	Amount      paypalAmount `json:"amount"`
}

type paypalApplicationContext struct {
	ReturnURL string `json:"return_url"`
	CancelURL string `json:"cancel_url"`
}

type paypalAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type paypalLink struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type paypalOrderResponse struct {
	ID            string                       `json:"id"`
	Status        string                       `json:"status"`
	Links         []paypalLink                 `json:"links,omitempty"`
	PurchaseUnits []paypalPurchaseUnitResponse `json:"purchase_units,omitempty"`
}

type paypalPurchaseUnitResponse struct {
	Amount   *paypalAmount               `json:"amount,omitempty"`
	Payments *paypalPurchaseUnitPayments `json:"payments,omitempty"`
}

type paypalPurchaseUnitPayments struct {
	Captures []paypalCapture `json:"captures,omitempty"`
}

type paypalCapture struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Amount paypalAmount `json:"amount"`
}

type paypalRefundRequest struct {
	Amount *paypalAmount `json:"amount,omitempty"`
}

type paypalRefundResponse struct {
	ID     string        `json:"id"`
	Status string        `json:"status"`
	Amount *paypalAmount `json:"amount,omitempty"`
}

type paypalWebhookEvent struct {
	ID         string      `json:"id"`
	EventType  string      `json:"event_type"`
	CreateTime string      `json:"create_time"`
	Resource   interface{} `json:"resource"`
}

type paypalWebhookHeaders struct {
	TransmissionID   string `json:"transmission_id"`
	TransmissionSig  string `json:"transmission_sig"`
	TransmissionTime string `json:"transmission_time"`
	CertURL          string `json:"cert_url"`
	AuthAlgo         string `json:"auth_algo"`
}

type paypalWebhookVerificationRequest struct {
	TransmissionID   string          `json:"transmission_id"`
	TransmissionTime string          `json:"transmission_time"`
	TransmissionSig  string          `json:"transmission_sig"`
	CertURL          string          `json:"cert_url"`
	AuthAlgo         string          `json:"auth_algo"`
	WebhookID        string          `json:"webhook_id"`
	WebhookEvent     json.RawMessage `json:"webhook_event"`
}
