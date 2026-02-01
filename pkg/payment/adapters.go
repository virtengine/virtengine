// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
// VE-2003: Real Stripe SDK integration (see stripe_adapter.go)
package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// environmentLive is the Adyen live environment identifier
const environmentLive = "live"

// ============================================================================
// Stripe Adapter Factory
// ============================================================================

// NewStripeGateway creates a Stripe gateway adapter.
// If useRealSDK is true, it returns the real Stripe SDK adapter (stripe_adapter.go).
// If false, it returns the stub adapter for testing without real API calls.
//
// For production use, always set useRealSDK to true.
// The stub adapter is ONLY for unit tests where you don't want network calls.
func NewStripeGateway(config StripeConfig, useRealSDK bool) (Gateway, error) {
	if useRealSDK {
		return NewRealStripeAdapter(config)
	}
	return NewStripeStubAdapter(config)
}

// ============================================================================
// Stripe Stub Adapter (DEPRECATED - Use NewRealStripeAdapter for production)
// ============================================================================

// stripeStubAdapter is a STUB implementation for testing only.
// Deprecated: Use StripeAdapter (stripe_adapter.go) for production.
//
// WARNING: This adapter returns FAKE customer IDs (cus_xxx) and payment intents.
// NO REAL PAYMENTS ARE PROCESSED. Do NOT use in production!
type stripeStubAdapter struct {
	config     StripeConfig
	httpClient *http.Client
	baseURL    string
}

// NewStripeStubAdapter creates a STUB Stripe adapter for testing.
// Deprecated: Use NewRealStripeAdapter for production.
//
// WARNING: This returns fake payment data. For production, use NewStripeGateway(config, true)
// or NewRealStripeAdapter(config).
func NewStripeStubAdapter(config StripeConfig) (Gateway, error) {
	if config.SecretKey == "" {
		return nil, ErrGatewayNotConfigured
	}

	baseURL := "https://api.stripe.com/v1"

	return &stripeStubAdapter{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}, nil
}

// NewStripeAdapter creates a new Stripe gateway adapter.
// Deprecated: This now returns the REAL Stripe SDK adapter.
// For explicit control, use NewStripeGateway(config, useRealSDK) instead.
func NewStripeAdapter(config StripeConfig) (Gateway, error) {
	// VE-2003: Now returns the real Stripe SDK adapter by default
	return NewRealStripeAdapter(config)
}

func (a *stripeStubAdapter) Name() string {
	return "Stripe"
}

func (a *stripeStubAdapter) Type() GatewayType {
	return GatewayStripe
}

func (a *stripeStubAdapter) IsHealthy(ctx context.Context) bool {
	// Check API connectivity by listing balance
	// In production, make actual API call
	return a.config.SecretKey != ""
}

func (a *stripeStubAdapter) Close() error {
	return nil
}

// ---- Customer Management ----

func (a *stripeStubAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	// In production, this would make actual Stripe API call
	// POST https://api.stripe.com/v1/customers
	customer := Customer{
		ID:          fmt.Sprintf("cus_%d", time.Now().UnixNano()),
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		VEIDAddress: req.VEIDAddress,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}
	return customer, nil
}

func (a *stripeStubAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	// GET https://api.stripe.com/v1/customers/{id}
	if !strings.HasPrefix(customerID, "cus_") {
		return Customer{}, ErrPaymentIntentNotFound
	}
	return Customer{ID: customerID}, nil
}

func (a *stripeStubAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	// POST https://api.stripe.com/v1/customers/{id}
	customer, err := a.GetCustomer(ctx, customerID)
	if err != nil {
		return Customer{}, err
	}

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
	for k, v := range req.Metadata {
		if customer.Metadata == nil {
			customer.Metadata = make(map[string]string)
		}
		customer.Metadata[k] = v
	}

	return customer, nil
}

func (a *stripeStubAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	// DELETE https://api.stripe.com/v1/customers/{id}
	return nil
}

// ---- Payment Methods ----

func (a *stripeStubAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	// POST https://api.stripe.com/v1/payment_methods/{id}/attach
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}
	return token.Token, nil
}

func (a *stripeStubAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	// POST https://api.stripe.com/v1/payment_methods/{id}/detach
	return nil
}

func (a *stripeStubAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	// GET https://api.stripe.com/v1/payment_methods?customer={id}&type=card
	return nil, nil
}

// ---- Payment Intents ----

func (a *stripeStubAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	// POST https://api.stripe.com/v1/payment_intents
	intent := PaymentIntent{
		ID:                  fmt.Sprintf("pi_%d", time.Now().UnixNano()),
		Gateway:             GatewayStripe,
		Amount:              req.Amount,
		Status:              PaymentIntentStatusRequiresPaymentMethod,
		CustomerID:          req.CustomerID,
		PaymentMethodID:     req.PaymentMethodID,
		Description:         req.Description,
		Metadata:            req.Metadata,
		ClientSecret:        fmt.Sprintf("pi_%d_secret_%d", time.Now().UnixNano(), time.Now().UnixNano()),
		ReceiptEmail:        req.ReceiptEmail,
		StatementDescriptor: req.StatementDescriptor,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// If payment method is provided, update status
	if req.PaymentMethodID != "" {
		intent.Status = PaymentIntentStatusRequiresConfirmation
	}

	return intent, nil
}

func (a *stripeStubAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	// GET https://api.stripe.com/v1/payment_intents/{id}
	if !strings.HasPrefix(paymentIntentID, "pi_") {
		return PaymentIntent{}, ErrPaymentIntentNotFound
	}
	return PaymentIntent{
		ID:      paymentIntentID,
		Gateway: GatewayStripe,
		Status:  PaymentIntentStatusSucceeded,
	}, nil
}

func (a *stripeStubAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	// POST https://api.stripe.com/v1/payment_intents/{id}/confirm
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	intent.PaymentMethodID = paymentMethodID
	intent.Status = PaymentIntentStatusSucceeded
	intent.CapturedAmount = intent.Amount
	intent.UpdatedAt = time.Now()

	return intent, nil
}

func (a *stripeStubAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	// POST https://api.stripe.com/v1/payment_intents/{id}/cancel
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	intent.Status = PaymentIntentStatusCanceled
	intent.UpdatedAt = time.Now()

	return intent, nil
}

func (a *stripeStubAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	// POST https://api.stripe.com/v1/payment_intents/{id}/capture
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	if amount != nil {
		intent.CapturedAmount = *amount
	} else {
		intent.CapturedAmount = intent.Amount
	}
	intent.Status = PaymentIntentStatusSucceeded
	intent.UpdatedAt = time.Now()

	return intent, nil
}

// ---- Refunds ----

func (a *stripeStubAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	// POST https://api.stripe.com/v1/refunds
	intent, err := a.GetPaymentIntent(ctx, req.PaymentIntentID)
	if err != nil {
		return Refund{}, err
	}

	refundAmount := intent.Amount
	if req.Amount != nil {
		refundAmount = *req.Amount
	}

	refund := Refund{
		ID:              fmt.Sprintf("re_%d", time.Now().UnixNano()),
		PaymentIntentID: req.PaymentIntentID,
		Amount:          refundAmount,
		Status:          RefundStatusSucceeded,
		Reason:          req.Reason,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
	}

	return refund, nil
}

func (a *stripeStubAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	// GET https://api.stripe.com/v1/refunds/{id}
	if !strings.HasPrefix(refundID, "re_") {
		return Refund{}, ErrRefundNotAllowed
	}
	return Refund{
		ID:     refundID,
		Status: RefundStatusSucceeded,
	}, nil
}

// ---- Webhooks ----

func (a *stripeStubAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return ErrWebhookSignatureInvalid
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

func (a *stripeStubAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var event struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Created int64           `json:"created"`
		Data    json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, err
	}

	return WebhookEvent{
		ID:        event.ID,
		Type:      WebhookEventType(event.Type),
		Gateway:   GatewayStripe,
		Payload:   payload,
		Timestamp: time.Unix(event.Created, 0),
	}, nil
}

// ============================================================================
// Adyen Adapter Factory
// ============================================================================

// NewAdyenGateway creates an Adyen gateway adapter.
// If useRealAPI is true, it returns the real Adyen API adapter (adyen_adapter.go).
// If false, it returns the stub adapter for testing without real API calls.
//
// For production use, always set useRealAPI to true.
// The stub adapter is ONLY for unit tests where you don't want network calls.
func NewAdyenGateway(config AdyenConfig, useRealAPI bool) (Gateway, error) {
	if useRealAPI {
		return NewRealAdyenAdapter(config)
	}
	return NewAdyenStubAdapter(config)
}

// ============================================================================
// Adyen Stub Adapter (DEPRECATED - Use NewRealAdyenAdapter for production)
// ============================================================================

// adyenStubAdapter is a STUB implementation for testing only.
// DEPRECATED: Use RealAdyenAdapter (adyen_adapter.go) for production.
//
// WARNING: This adapter returns FAKE payment IDs and responses.
// NO REAL PAYMENTS ARE PROCESSED. Do NOT use in production!
type adyenStubAdapter struct {
	config     AdyenConfig
	httpClient *http.Client
	baseURL    string
}

// NewAdyenStubAdapter creates a STUB Adyen adapter for testing.
// DEPRECATED: Use NewRealAdyenAdapter for production.
//
// WARNING: This returns fake payment data. For production, use NewAdyenGateway(config, true)
// or NewRealAdyenAdapter(config).
func NewAdyenStubAdapter(config AdyenConfig) (Gateway, error) {
	if config.APIKey == "" || config.MerchantAccount == "" {
		return nil, ErrGatewayNotConfigured
	}

	baseURL := "https://checkout-test.adyen.com/v71"
	if config.Environment == environmentLive {
		baseURL = fmt.Sprintf("https://%s-checkout-live.adyenpayments.com/checkout/v71", config.LiveEndpointURLPrefix)
	}

	return &adyenStubAdapter{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}, nil
}

// NewAdyenAdapter creates a new Adyen gateway adapter.
// DEPRECATED: This now returns the REAL Adyen API adapter.
// For explicit control, use NewAdyenGateway(config, useRealAPI) instead.
func NewAdyenAdapter(config AdyenConfig) (Gateway, error) {
	// VE-3059: Now returns the real Adyen API adapter by default
	return NewRealAdyenAdapter(config)
}

func (a *adyenStubAdapter) Name() string {
	return "Adyen"
}

func (a *adyenStubAdapter) Type() GatewayType {
	return GatewayAdyen
}

func (a *adyenStubAdapter) IsHealthy(ctx context.Context) bool {
	return a.config.APIKey != ""
}

func (a *adyenStubAdapter) Close() error {
	return nil
}

// ---- Customer Management (Adyen uses shopperReference) ----

func (a *adyenStubAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	// Adyen doesn't have explicit customer creation
	// Customers are created implicitly via shopperReference
	customer := Customer{
		ID:          req.VEIDAddress, // Use VEID address as shopperReference
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		VEIDAddress: req.VEIDAddress,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}
	return customer, nil
}

func (a *adyenStubAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	// Adyen doesn't have customer retrieval API
	return Customer{ID: customerID}, nil
}

func (a *adyenStubAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	customer := Customer{ID: customerID}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Name != nil {
		customer.Name = *req.Name
	}
	return customer, nil
}

func (a *adyenStubAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	// Adyen doesn't have customer deletion
	return nil
}

// ---- Payment Methods ----

func (a *adyenStubAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	// POST /paymentMethods/storedPaymentMethods
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}
	return token.Token, nil
}

func (a *adyenStubAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	// DELETE /paymentMethods/{storedPaymentMethodId}
	return nil
}

func (a *adyenStubAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	// POST /paymentMethods with shopperReference
	return nil, nil
}

// ---- Payment Intents (Adyen uses /payments) ----

func (a *adyenStubAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	// POST /payments
	intent := PaymentIntent{
		ID:                  fmt.Sprintf("ADYEN_%d", time.Now().UnixNano()),
		Gateway:             GatewayAdyen,
		Amount:              req.Amount,
		Status:              PaymentIntentStatusRequiresPaymentMethod,
		CustomerID:          req.CustomerID,
		PaymentMethodID:     req.PaymentMethodID,
		Description:         req.Description,
		Metadata:            req.Metadata,
		StatementDescriptor: req.StatementDescriptor,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	return intent, nil
}

func (a *adyenStubAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	// GET payment details via webhook or stored state
	if !strings.HasPrefix(paymentIntentID, "ADYEN_") {
		return PaymentIntent{}, ErrPaymentIntentNotFound
	}
	return PaymentIntent{
		ID:      paymentIntentID,
		Gateway: GatewayAdyen,
		Status:  PaymentIntentStatusSucceeded,
	}, nil
}

func (a *adyenStubAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	// POST /payments/details for 3DS completion
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	intent.PaymentMethodID = paymentMethodID
	intent.Status = PaymentIntentStatusSucceeded
	intent.CapturedAmount = intent.Amount
	intent.UpdatedAt = time.Now()

	return intent, nil
}

func (a *adyenStubAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	// POST /payments/{paymentPspReference}/cancels
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	intent.Status = PaymentIntentStatusCanceled
	intent.UpdatedAt = time.Now()

	return intent, nil
}

func (a *adyenStubAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	// POST /payments/{paymentPspReference}/captures
	intent, err := a.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return PaymentIntent{}, err
	}

	if amount != nil {
		intent.CapturedAmount = *amount
	} else {
		intent.CapturedAmount = intent.Amount
	}
	intent.Status = PaymentIntentStatusSucceeded
	intent.UpdatedAt = time.Now()

	return intent, nil
}

// ---- Refunds ----

func (a *adyenStubAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	// POST /payments/{paymentPspReference}/refunds
	intent, err := a.GetPaymentIntent(ctx, req.PaymentIntentID)
	if err != nil {
		return Refund{}, err
	}

	refundAmount := intent.Amount
	if req.Amount != nil {
		refundAmount = *req.Amount
	}

	refund := Refund{
		ID:              fmt.Sprintf("ADYEN_RF_%d", time.Now().UnixNano()),
		PaymentIntentID: req.PaymentIntentID,
		Amount:          refundAmount,
		Status:          RefundStatusPending, // Adyen refunds are async
		Reason:          req.Reason,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
	}

	return refund, nil
}

func (a *adyenStubAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	if !strings.HasPrefix(refundID, "ADYEN_RF_") {
		return Refund{}, ErrRefundNotAllowed
	}
	return Refund{
		ID:     refundID,
		Status: RefundStatusSucceeded,
	}, nil
}

// ---- Webhooks ----

func (a *adyenStubAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.HMACKey == "" {
		return nil // HMAC validation disabled
	}

	// Validate Adyen HMAC signature
	mac := hmac.New(sha256.New, []byte(a.config.HMACKey))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func (a *adyenStubAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	var notification struct {
		NotificationItems []struct {
			NotificationRequestItem struct {
				EventCode   string          `json:"eventCode"`
				EventDate   string          `json:"eventDate"`
				PspReference string         `json:"pspReference"`
				AdditionalData json.RawMessage `json:"additionalData"`
			} `json:"NotificationRequestItem"`
		} `json:"notificationItems"`
	}

	if err := json.Unmarshal(payload, &notification); err != nil {
		return WebhookEvent{}, err
	}

	if len(notification.NotificationItems) == 0 {
		return WebhookEvent{}, ErrWebhookEventUnknown
	}

	item := notification.NotificationItems[0].NotificationRequestItem

	// Map Adyen event codes to our event types
	var eventType WebhookEventType
	switch item.EventCode {
	case "AUTHORISATION":
		eventType = WebhookEventPaymentIntentSucceeded
	case "CANCELLATION":
		eventType = WebhookEventPaymentIntentCanceled
	case "REFUND":
		eventType = WebhookEventChargeRefunded
	case "CHARGEBACK":
		eventType = WebhookEventChargeDisputeCreated
	default:
		eventType = WebhookEventType(item.EventCode)
	}

	return WebhookEvent{
		ID:        item.PspReference,
		Type:      eventType,
		Gateway:   GatewayAdyen,
		Payload:   payload,
		Timestamp: time.Now(),
	}, nil
}

