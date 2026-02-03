// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-2003: Real Stripe payment adapter implementation
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/customer"
	"github.com/stripe/stripe-go/v80/dispute"
	"github.com/stripe/stripe-go/v80/paymentintent"
	"github.com/stripe/stripe-go/v80/paymentmethod"
	"github.com/stripe/stripe-go/v80/refund"
	"github.com/stripe/stripe-go/v80/webhook"
)

// ============================================================================
// Real Stripe Adapter - VE-2003
// ============================================================================

// StripeAdapter implements the Gateway interface using the real Stripe SDK.
// This replaces the stub stripeAdapter with actual Stripe API calls.
//
// Security Notes:
// - NEVER log API keys
// - NEVER store raw card data (use Stripe tokenization)
// - All sensitive data is handled through Stripe's PCI-compliant APIs
type StripeAdapter struct {
	config     StripeConfig
	httpClient *http.Client
	testMode   bool
}

// NewRealStripeAdapter creates a new Stripe gateway adapter with real SDK integration.
// It validates the configuration and sets up the Stripe SDK with the provided API key.
//
// Parameters:
//   - config: StripeConfig containing API keys and settings
//
// Returns:
//   - Gateway: The configured Stripe adapter
//   - error: ErrGatewayNotConfigured if API key is missing
func NewRealStripeAdapter(config StripeConfig) (Gateway, error) {
	if config.SecretKey == "" {
		return nil, ErrGatewayNotConfigured
	}

	// Set the global Stripe API key
	// The SDK uses this for all subsequent API calls
	stripe.Key = config.SecretKey

	// Note: API version is managed by the SDK and cannot be overridden directly
	// The SDK uses the API version it was built against

	// Determine if we're in test mode based on the key prefix
	testMode := len(config.SecretKey) > 3 && config.SecretKey[:7] == "sk_test"

	return &StripeAdapter{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		testMode:   testMode,
	}, nil
}

func (a *StripeAdapter) Name() string {
	return "Stripe"
}

func (a *StripeAdapter) Type() GatewayType {
	return GatewayStripe
}

func (a *StripeAdapter) IsHealthy(ctx context.Context) bool {
	// Verify API connectivity by attempting to list customers with limit 1
	// This is a lightweight operation that validates API key correctness
	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(1)
	params.Context = ctx

	iter := customer.List(params)
	// If we can iterate without error, the API is healthy
	// Note: Even empty results indicate a working connection
	if iter.Err() != nil {
		return false
	}
	return true
}

func (a *StripeAdapter) Close() error {
	// The Stripe SDK doesn't require explicit cleanup
	return nil
}

// ============================================================================
// Customer Management
// ============================================================================

func (a *StripeAdapter) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(req.Email),
	}
	params.Context = ctx

	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}
	if req.Phone != "" {
		params.Phone = stripe.String(req.Phone)
	}

	// Add metadata including VEID address for blockchain correlation
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	if req.VEIDAddress != "" {
		req.Metadata["veid_address"] = req.VEIDAddress
	}
	for k, v := range req.Metadata {
		params.AddMetadata(k, v)
	}

	result, err := customer.New(params)
	if err != nil {
		return Customer{}, convertStripeError(err)
	}

	return Customer{
		ID:          result.ID,
		Email:       result.Email,
		Name:        result.Name,
		Phone:       result.Phone,
		VEIDAddress: req.VEIDAddress,
		Metadata:    result.Metadata,
		CreatedAt:   time.Unix(result.Created, 0),
	}, nil
}

func (a *StripeAdapter) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	params := &stripe.CustomerParams{}
	params.Context = ctx

	result, err := customer.Get(customerID, params)
	if err != nil {
		return Customer{}, convertStripeError(err)
	}

	return Customer{
		ID:                     result.ID,
		Email:                  result.Email,
		Name:                   result.Name,
		Phone:                  result.Phone,
		DefaultPaymentMethodID: getDefaultPaymentMethodID(result),
		Metadata:               result.Metadata,
		CreatedAt:              time.Unix(result.Created, 0),
	}, nil
}

func (a *StripeAdapter) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	params := &stripe.CustomerParams{}
	params.Context = ctx

	if req.Email != nil {
		params.Email = stripe.String(*req.Email)
	}
	if req.Name != nil {
		params.Name = stripe.String(*req.Name)
	}
	if req.Phone != nil {
		params.Phone = stripe.String(*req.Phone)
	}
	if req.DefaultPaymentMethodID != nil {
		params.InvoiceSettings = &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(*req.DefaultPaymentMethodID),
		}
	}
	for k, v := range req.Metadata {
		params.AddMetadata(k, v)
	}

	result, err := customer.Update(customerID, params)
	if err != nil {
		return Customer{}, convertStripeError(err)
	}

	return Customer{
		ID:                     result.ID,
		Email:                  result.Email,
		Name:                   result.Name,
		Phone:                  result.Phone,
		DefaultPaymentMethodID: getDefaultPaymentMethodID(result),
		Metadata:               result.Metadata,
		CreatedAt:              time.Unix(result.Created, 0),
	}, nil
}

func (a *StripeAdapter) DeleteCustomer(ctx context.Context, customerID string) error {
	params := &stripe.CustomerParams{}
	params.Context = ctx

	_, err := customer.Del(customerID, params)
	if err != nil {
		return convertStripeError(err)
	}
	return nil
}

// ============================================================================
// Payment Methods
// ============================================================================

func (a *StripeAdapter) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	if token.Token == "" {
		return "", ErrInvalidCardToken
	}

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	params.Context = ctx

	result, err := paymentmethod.Attach(token.Token, params)
	if err != nil {
		return "", convertStripeError(err)
	}

	return result.ID, nil
}

func (a *StripeAdapter) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	params := &stripe.PaymentMethodDetachParams{}
	params.Context = ctx

	_, err := paymentmethod.Detach(paymentMethodID, params)
	if err != nil {
		return convertStripeError(err)
	}
	return nil
}

func (a *StripeAdapter) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}
	params.Context = ctx

	iter := paymentmethod.List(params)
	var methods []CardToken

	for iter.Next() {
		pm := iter.PaymentMethod()
		if pm.Card != nil {
			methods = append(methods, CardToken{
				Token:       pm.ID,
				Last4:       pm.Card.Last4,
				ExpiryMonth: int(pm.Card.ExpMonth),
				ExpiryYear:  int(pm.Card.ExpYear),
				Brand:       mapStripeCardBrand(pm.Card.Brand),
			})
		}
	}

	if err := iter.Err(); err != nil {
		return nil, convertStripeError(err)
	}

	return methods, nil
}

// ============================================================================
// Payment Intents
// ============================================================================

func (a *StripeAdapter) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount.Value),
		Currency: stripe.String(string(req.Amount.Currency)),
	}
	params.Context = ctx

	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}
	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
	}
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if req.ReceiptEmail != "" {
		params.ReceiptEmail = stripe.String(req.ReceiptEmail)
	}
	if req.StatementDescriptor != "" {
		params.StatementDescriptorSuffix = stripe.String(req.StatementDescriptor)
	}

	// Set capture method based on request
	if req.CaptureMethod == "manual" {
		params.CaptureMethod = stripe.String(string(stripe.PaymentIntentCaptureMethodManual))
	} else {
		params.CaptureMethod = stripe.String(string(stripe.PaymentIntentCaptureMethodAutomatic))
	}

	// Add metadata
	for k, v := range req.Metadata {
		params.AddMetadata(k, v)
	}

	// Add idempotency key if provided
	if req.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(req.IdempotencyKey)
	}

	result, err := paymentintent.New(params)
	if err != nil {
		return PaymentIntent{}, convertStripeError(err)
	}

	return mapStripePaymentIntent(result), nil
}

func (a *StripeAdapter) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{}
	params.Context = ctx

	result, err := paymentintent.Get(paymentIntentID, params)
	if err != nil {
		return PaymentIntent{}, convertStripeError(err)
	}

	return mapStripePaymentIntent(result), nil
}

func (a *StripeAdapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{}
	params.Context = ctx

	if paymentMethodID != "" {
		params.PaymentMethod = stripe.String(paymentMethodID)
	}

	result, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return PaymentIntent{}, convertStripeError(err)
	}

	return mapStripePaymentIntent(result), nil
}

func (a *StripeAdapter) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	params := &stripe.PaymentIntentCancelParams{}
	params.Context = ctx

	// Map cancellation reason to Stripe's enum
	if reason != "" {
		switch reason {
		case "duplicate":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonDuplicate))
		case "fraudulent":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonFraudulent))
		case "requested_by_customer":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonRequestedByCustomer))
		case "abandoned":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonAbandoned))
		}
	}

	result, err := paymentintent.Cancel(paymentIntentID, params)
	if err != nil {
		return PaymentIntent{}, convertStripeError(err)
	}

	return mapStripePaymentIntent(result), nil
}

func (a *StripeAdapter) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	params := &stripe.PaymentIntentCaptureParams{}
	params.Context = ctx

	// Optional: Capture a partial amount
	if amount != nil {
		params.AmountToCapture = stripe.Int64(amount.Value)
	}

	result, err := paymentintent.Capture(paymentIntentID, params)
	if err != nil {
		return PaymentIntent{}, convertStripeError(err)
	}

	return mapStripePaymentIntent(result), nil
}

// ============================================================================
// Refunds
// ============================================================================

func (a *StripeAdapter) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
	}
	params.Context = ctx

	// Optional: Partial refund
	if req.Amount != nil {
		params.Amount = stripe.Int64(req.Amount.Value)
	}

	// Map refund reason
	if req.Reason != "" {
		switch req.Reason {
		case "duplicate":
			params.Reason = stripe.String(string(stripe.RefundReasonDuplicate))
		case "fraudulent":
			params.Reason = stripe.String(string(stripe.RefundReasonFraudulent))
		case "requested_by_customer":
			params.Reason = stripe.String(string(stripe.RefundReasonRequestedByCustomer))
		}
	}

	// Add metadata
	for k, v := range req.Metadata {
		params.AddMetadata(k, v)
	}

	// Add idempotency key if provided
	if req.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(req.IdempotencyKey)
	}

	result, err := refund.New(params)
	if err != nil {
		return Refund{}, convertStripeError(err)
	}

	return mapStripeRefund(result), nil
}

func (a *StripeAdapter) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	params := &stripe.RefundParams{}
	params.Context = ctx

	result, err := refund.Get(refundID, params)
	if err != nil {
		return Refund{}, convertStripeError(err)
	}

	return mapStripeRefund(result), nil
}

// ============================================================================
// Webhooks
// ============================================================================

func (a *StripeAdapter) ValidateWebhook(payload []byte, signature string) error {
	if a.config.WebhookSecret == "" {
		return ErrWebhookSignatureInvalid
	}

	_, err := webhook.ConstructEvent(payload, signature, a.config.WebhookSecret)
	if err != nil {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func (a *StripeAdapter) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	event, err := webhook.ConstructEventWithOptions(payload, "", a.config.WebhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		// If webhook secret is empty, try parsing without validation
		if a.config.WebhookSecret == "" {
			var stripeEvent stripe.Event
			if jsonErr := json.Unmarshal(payload, &stripeEvent); jsonErr != nil {
				return WebhookEvent{}, jsonErr
			}
			event = stripeEvent
		} else {
			return WebhookEvent{}, err
		}
	}

	return WebhookEvent{
		ID:        event.ID,
		Type:      mapStripeEventType(string(event.Type)),
		Gateway:   GatewayStripe,
		Payload:   payload,
		Timestamp: time.Unix(event.Created, 0),
	}, nil
}

// ============================================================================
// Dispute Methods - PAY-003
// ============================================================================

// GetDispute retrieves a dispute by ID from Stripe
func (a *StripeAdapter) GetDispute(ctx context.Context, disputeID string) (Dispute, error) {
	params := &stripe.DisputeParams{}
	params.Context = ctx

	result, err := dispute.Get(disputeID, params)
	if err != nil {
		return Dispute{}, convertStripeError(err)
	}

	return mapStripeDispute(result), nil
}

// ListDisputes retrieves disputes for a payment intent
func (a *StripeAdapter) ListDisputes(ctx context.Context, paymentIntentID string) ([]Dispute, error) {
	params := &stripe.DisputeListParams{}
	params.Context = ctx
	params.PaymentIntent = stripe.String(paymentIntentID)

	iter := dispute.List(params)
	var disputes []Dispute

	for iter.Next() {
		d := iter.Dispute()
		disputes = append(disputes, mapStripeDispute(d))
	}

	if err := iter.Err(); err != nil {
		return nil, convertStripeError(err)
	}

	return disputes, nil
}

// SubmitDisputeEvidence submits evidence for a dispute
func (a *StripeAdapter) SubmitDisputeEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error {
	params := &stripe.DisputeParams{}
	params.Context = ctx

	// Map evidence fields to Stripe's evidence structure
	if evidence.ProductDescription != "" {
		params.Evidence = &stripe.DisputeEvidenceParams{
			ProductDescription: stripe.String(evidence.ProductDescription),
		}
	}
	if evidence.CustomerEmail != "" {
		if params.Evidence == nil {
			params.Evidence = &stripe.DisputeEvidenceParams{}
		}
		params.Evidence.CustomerEmailAddress = stripe.String(evidence.CustomerEmail)
	}
	if evidence.CustomerPurchaseIP != "" {
		if params.Evidence == nil {
			params.Evidence = &stripe.DisputeEvidenceParams{}
		}
		params.Evidence.CustomerPurchaseIP = stripe.String(evidence.CustomerPurchaseIP)
	}
	if evidence.BillingAddress != "" {
		if params.Evidence == nil {
			params.Evidence = &stripe.DisputeEvidenceParams{}
		}
		params.Evidence.BillingAddress = stripe.String(evidence.BillingAddress)
	}
	if evidence.UncategorizedText != "" {
		if params.Evidence == nil {
			params.Evidence = &stripe.DisputeEvidenceParams{}
		}
		params.Evidence.UncategorizedText = stripe.String(evidence.UncategorizedText)
	}

	// Submit the evidence (doesn't mark as "submit" yet - needs explicit Submit=true)
	params.Submit = stripe.Bool(true)

	_, err := dispute.Update(disputeID, params)
	if err != nil {
		return convertStripeError(err)
	}

	return nil
}

// AcceptDispute accepts (concedes) a dispute in Stripe
// This closes the dispute in favor of the cardholder
func (a *StripeAdapter) AcceptDispute(ctx context.Context, disputeID string) error {
	params := &stripe.DisputeParams{}
	params.Context = ctx

	_, err := dispute.Close(disputeID, params)
	if err != nil {
		return convertStripeError(err)
	}

	return nil
}

// mapStripeDispute converts a Stripe Dispute to our Dispute type
func mapStripeDispute(d *stripe.Dispute) Dispute {
	disp := Dispute{
		ID:        d.ID,
		Gateway:   GatewayStripe,
		Status:    mapStripeDisputeStatus(d.Status),
		Reason:    mapStripeDisputeReason(d.Reason),
		CreatedAt: time.Unix(d.Created, 0),
		UpdatedAt: time.Now(),
	}

	// Map amount
	disp.Amount = Amount{
		Value:    d.Amount,
		Currency: Currency(d.Currency),
	}

	// Map charge -> payment intent relationship
	if d.Charge != nil {
		disp.ChargeID = d.Charge.ID
	}
	if d.PaymentIntent != nil {
		disp.PaymentIntentID = d.PaymentIntent.ID
	}

	// Map evidence due date
	if d.EvidenceDetails != nil && d.EvidenceDetails.DueBy > 0 {
		disp.EvidenceDueBy = time.Unix(d.EvidenceDetails.DueBy, 0)
	}

	// Map network reason code
	disp.NetworkReasonCode = d.NetworkReasonCode

	// Check if refundable (only certain statuses allow refund)
	disp.IsRefundable = d.Status == stripe.DisputeStatusNeedsResponse ||
		d.Status == stripe.DisputeStatusWarningNeedsResponse

	// Add metadata
	if d.Metadata != nil {
		disp.Metadata = d.Metadata
	}

	return disp
}

// mapStripeDisputeStatus converts Stripe dispute status to our DisputeStatus
func mapStripeDisputeStatus(status stripe.DisputeStatus) DisputeStatus {
	switch status {
	case stripe.DisputeStatusNeedsResponse, stripe.DisputeStatusWarningNeedsResponse:
		return DisputeStatusNeedsResponse
	case stripe.DisputeStatusUnderReview, stripe.DisputeStatusWarningUnderReview:
		return DisputeStatusUnderReview
	case stripe.DisputeStatusWon:
		return DisputeStatusWon
	case stripe.DisputeStatusLost:
		return DisputeStatusLost
	case stripe.DisputeStatusWarningClosed:
		return DisputeStatusAccepted
	default:
		return DisputeStatusOpen
	}
}

// mapStripeDisputeReason converts Stripe dispute reason to our DisputeReason
func mapStripeDisputeReason(reason stripe.DisputeReason) DisputeReason {
	switch reason {
	case stripe.DisputeReasonFraudulent:
		return DisputeReasonFraudulent
	case stripe.DisputeReasonDuplicate:
		return DisputeReasonDuplicate
	case stripe.DisputeReasonProductNotReceived:
		return DisputeReasonProductNotReceived
	case stripe.DisputeReasonUnrecognized:
		return DisputeReasonUnrecognized
	default:
		return DisputeReasonGeneral
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// getDefaultPaymentMethodID extracts the default payment method from a customer.
func getDefaultPaymentMethodID(c *stripe.Customer) string {
	if c.InvoiceSettings != nil && c.InvoiceSettings.DefaultPaymentMethod != nil {
		return c.InvoiceSettings.DefaultPaymentMethod.ID
	}
	return ""
}

// mapStripeCardBrand converts Stripe card brand to our CardBrand type.
func mapStripeCardBrand(brand stripe.PaymentMethodCardBrand) CardBrand {
	switch brand {
	case stripe.PaymentMethodCardBrandVisa:
		return CardBrandVisa
	case stripe.PaymentMethodCardBrandMastercard:
		return CardBrandMastercard
	case stripe.PaymentMethodCardBrandAmex:
		return CardBrandAmex
	case stripe.PaymentMethodCardBrandDiscover:
		return CardBrandDiscover
	case stripe.PaymentMethodCardBrandJCB:
		return CardBrand("jcb")
	case stripe.PaymentMethodCardBrandDiners:
		return CardBrand("diners")
	case stripe.PaymentMethodCardBrandUnionpay:
		return CardBrand("unionpay")
	default:
		return CardBrandUnknown
	}
}

// mapStripePaymentIntentStatus converts Stripe status to our PaymentIntentStatus.
func mapStripePaymentIntentStatus(status stripe.PaymentIntentStatus) PaymentIntentStatus {
	switch status {
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return PaymentIntentStatusRequiresPaymentMethod
	case stripe.PaymentIntentStatusRequiresConfirmation:
		return PaymentIntentStatusRequiresConfirmation
	case stripe.PaymentIntentStatusRequiresAction:
		return PaymentIntentStatusRequiresAction
	case stripe.PaymentIntentStatusProcessing:
		return PaymentIntentStatusProcessing
	case stripe.PaymentIntentStatusSucceeded:
		return PaymentIntentStatusSucceeded
	case stripe.PaymentIntentStatusCanceled:
		return PaymentIntentStatusCanceled
	case stripe.PaymentIntentStatusRequiresCapture:
		// RequiresCapture maps to processing since we don't have a specific status for it
		return PaymentIntentStatusProcessing
	default:
		return PaymentIntentStatusProcessing
	}
}

// mapStripePaymentIntent converts a Stripe PaymentIntent to our PaymentIntent type.
func mapStripePaymentIntent(pi *stripe.PaymentIntent) PaymentIntent {
	intent := PaymentIntent{
		ID:      pi.ID,
		Gateway: GatewayStripe,
		Amount: Amount{
			Value:    pi.Amount,
			Currency: Currency(pi.Currency),
		},
		Status:       mapStripePaymentIntentStatus(pi.Status),
		ClientSecret: pi.ClientSecret,
		Metadata:     pi.Metadata,
		CreatedAt:    time.Unix(pi.Created, 0),
		UpdatedAt:    time.Now(),
	}

	if pi.Customer != nil {
		intent.CustomerID = pi.Customer.ID
	}
	if pi.PaymentMethod != nil {
		intent.PaymentMethodID = pi.PaymentMethod.ID
	}
	if pi.Description != "" {
		intent.Description = pi.Description
	}
	if pi.ReceiptEmail != "" {
		intent.ReceiptEmail = pi.ReceiptEmail
	}
	if pi.StatementDescriptorSuffix != "" {
		intent.StatementDescriptor = pi.StatementDescriptorSuffix
	}

	// Set captured amount if payment succeeded
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		intent.CapturedAmount = Amount{
			Value:    pi.AmountReceived,
			Currency: Currency(pi.Currency),
		}
	}

	// Calculate refunded amount if any charges exist
	if pi.LatestCharge != nil {
		intent.RefundedAmount = Amount{
			Value:    pi.LatestCharge.AmountRefunded,
			Currency: Currency(pi.Currency),
		}
	}

	return intent
}

// mapStripeRefundStatus converts Stripe refund status to our RefundStatus.
func mapStripeRefundStatus(status stripe.RefundStatus) RefundStatus {
	switch status {
	case stripe.RefundStatusPending:
		return RefundStatusPending
	case stripe.RefundStatusSucceeded:
		return RefundStatusSucceeded
	case stripe.RefundStatusFailed:
		return RefundStatusFailed
	case stripe.RefundStatusCanceled:
		return RefundStatusCanceled
	default:
		return RefundStatusPending
	}
}

// mapStripeRefund converts a Stripe Refund to our Refund type.
func mapStripeRefund(r *stripe.Refund) Refund {
	ref := Refund{
		ID:     r.ID,
		Status: mapStripeRefundStatus(r.Status),
		Amount: Amount{
			Value:    r.Amount,
			Currency: Currency(r.Currency),
		},
		Metadata:  r.Metadata,
		CreatedAt: time.Unix(r.Created, 0),
	}

	if r.PaymentIntent != nil {
		ref.PaymentIntentID = r.PaymentIntent.ID
	}
	if r.Reason != "" {
		ref.Reason = RefundReason(r.Reason)
	}
	if r.FailureReason != "" {
		ref.FailureReason = string(r.FailureReason)
	}

	return ref
}

// mapStripeEventType converts Stripe event types to our WebhookEventType.
func mapStripeEventType(eventType string) WebhookEventType {
	switch eventType {
	case "payment_intent.succeeded":
		return WebhookEventPaymentIntentSucceeded
	case "payment_intent.payment_failed":
		return WebhookEventPaymentIntentFailed
	case "payment_intent.canceled":
		return WebhookEventPaymentIntentCanceled
	case "payment_intent.requires_action":
		// Map to processing since we don't have requires_action event type
		return WebhookEventPaymentIntentProcessing
	case "payment_method.attached":
		return WebhookEventPaymentMethodAttached
	case "payment_method.detached":
		return WebhookEventPaymentMethodDetached
	case "charge.refunded":
		return WebhookEventChargeRefunded
	case "charge.dispute.created":
		return WebhookEventChargeDisputeCreated
	case "charge.dispute.updated":
		return WebhookEventChargeDisputeUpdated
	case "charge.dispute.closed":
		return WebhookEventChargeDisputeClosed
	case "charge.dispute.funds_withdrawn":
		return WebhookEventChargeDisputeFundsWithdrawn
	case "charge.dispute.funds_reinstated":
		return WebhookEventChargeDisputeFundsReinstated
	case "customer.created":
		return WebhookEventCustomerCreated
	case "customer.updated":
		// Map to created since we don't have updated event type
		return WebhookEventCustomerCreated
	case "customer.deleted":
		return WebhookEventCustomerDeleted
	default:
		return WebhookEventType(eventType)
	}
}

// convertStripeError converts Stripe API errors to our domain errors.
// SECURITY: This function redacts sensitive error details while preserving
// useful debugging information (transaction IDs, error codes).
func convertStripeError(err error) error {
	if err == nil {
		return nil
	}

	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.Code {
		case stripe.ErrorCodeCardDeclined:
			return ErrPaymentDeclined
		case stripe.ErrorCodeExpiredCard:
			return ErrCardExpired
		case stripe.ErrorCodeInsufficientFunds:
			return ErrInsufficientFunds
		case stripe.ErrorCodeInvalidCardType:
			return ErrInvalidCardToken
		case stripe.ErrorCodeResourceMissing:
			return ErrPaymentIntentNotFound
		case stripe.ErrorCodeIdempotencyKeyInUse:
			return ErrDuplicateIdempotencyKey
		}

		// For other errors, wrap with context but don't expose raw Stripe messages
		// as they may contain sensitive information
		switch stripeErr.Type {
		case stripe.ErrorTypeCard:
			return fmt.Errorf("card error: %w", ErrPaymentDeclined)
		case stripe.ErrorTypeInvalidRequest:
			return fmt.Errorf("invalid request: %s", stripeErr.Code)
		case stripe.ErrorTypeAPI:
			return ErrGatewayUnavailable
		}
	}

	// Generic error - don't expose details
	return fmt.Errorf("payment processing failed: %w", ErrGatewayUnavailable)
}

// ============================================================================
// Test Mode Utilities
// ============================================================================

// IsTestMode returns true if the adapter is configured with test API keys.
func (a *StripeAdapter) IsTestMode() bool {
	return a.testMode
}

// GetTestCardNumbers returns a map of test card scenarios for integration testing.
// These are Stripe's official test card numbers.
func GetTestCardNumbers() map[string]string {
	return map[string]string{
		"visa_success":           "4242424242424242",
		"visa_debit":             "4000056655665556",
		"mastercard_success":     "5555555555554444",
		"mastercard_debit":       "5200828282828210",
		"amex_success":           "378282246310005",
		"discover_success":       "6011111111111117",
		"declined_generic":       "4000000000000002",
		"declined_insufficient":  "4000000000009995",
		"declined_expired":       "4000000000000069",
		"declined_cvc":           "4000000000000127",
		"declined_processing":    "4000000000000119",
		"3ds_required":           "4000002500003155",
		"3ds_required_challenge": "4000002760003184",
	}
}
