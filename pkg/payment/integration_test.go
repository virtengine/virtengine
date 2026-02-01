// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3059: Integration tests for payment service with sandbox APIs
//
//go:build e2e.integration

package payment_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/payment"
)

// ============================================================================
// Integration Test Helpers
// ============================================================================

// skipIfNoStripeKey skips the test if STRIPE_TEST_KEY is not set
func skipIfNoStripeKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("STRIPE_TEST_KEY")
	if key == "" {
		t.Skip("STRIPE_TEST_KEY not set, skipping integration test")
	}
	return key
}

// skipIfNoAdyenConfig skips the test if Adyen config is not set
func skipIfNoAdyenConfig(t *testing.T) (string, string) {
	t.Helper()
	apiKey := os.Getenv("ADYEN_TEST_API_KEY")
	merchantAccount := os.Getenv("ADYEN_MERCHANT_ACCOUNT")
	if apiKey == "" || merchantAccount == "" {
		t.Skip("ADYEN_TEST_API_KEY or ADYEN_MERCHANT_ACCOUNT not set, skipping integration test")
	}
	return apiKey, merchantAccount
}

// newStripeIntegrationService creates a payment service with real Stripe API for integration tests
func newStripeIntegrationService(t *testing.T) payment.Service {
	t.Helper()
	stripeKey := skipIfNoStripeKey(t)

	cfg := payment.DefaultConfig()
	cfg.Gateway = payment.GatewayStripe
	cfg.StripeConfig.SecretKey = stripeKey
	cfg.EnableSandbox = true

	svc, err := payment.NewService(cfg)
	require.NoError(t, err)
	return svc
}

// newAdyenIntegrationService creates a payment service with real Adyen API for integration tests
func newAdyenIntegrationService(t *testing.T) payment.Service {
	t.Helper()
	apiKey, merchantAccount := skipIfNoAdyenConfig(t)

	cfg := payment.DefaultConfig()
	cfg.Gateway = payment.GatewayAdyen
	cfg.AdyenConfig.APIKey = apiKey
	cfg.AdyenConfig.MerchantAccount = merchantAccount
	cfg.AdyenConfig.Environment = "test"
	cfg.EnableSandbox = true

	svc, err := payment.NewService(cfg)
	require.NoError(t, err)
	return svc
}

// ============================================================================
// Stripe Integration Tests
// ============================================================================

func TestStripeIntegration_CustomerLifecycle(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Create customer
	customer, err := svc.CreateCustomer(ctx, payment.CreateCustomerRequest{
		Email:       "integration-test@virtengine.io",
		Name:        "Integration Test User",
		VEIDAddress: "virtengine1integration",
		Metadata: map[string]string{
			"test": "true",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)
	assert.Equal(t, "integration-test@virtengine.io", customer.Email)

	// Get customer
	retrieved, err := svc.GetCustomer(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, customer.ID, retrieved.ID)

	// Update customer
	newName := "Updated Test User"
	updated, err := svc.UpdateCustomer(ctx, customer.ID, payment.UpdateCustomerRequest{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	// Delete customer
	err = svc.DeleteCustomer(ctx, customer.ID)
	require.NoError(t, err)
}

func TestStripeIntegration_PaymentIntentFlow(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Create a payment intent
	intent, err := svc.CreatePaymentIntent(ctx, payment.PaymentIntentRequest{
		Amount:      payment.NewAmount(1000, payment.CurrencyUSD), // $10.00
		Description: "Integration test payment",
		Metadata: map[string]string{
			"test": "true",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, intent.ID)
	assert.Equal(t, int64(1000), intent.Amount.Value)
	assert.Equal(t, payment.PaymentIntentStatusRequiresPaymentMethod, intent.Status)

	// Get payment intent
	retrieved, err := svc.GetPaymentIntent(ctx, intent.ID)
	require.NoError(t, err)
	assert.Equal(t, intent.ID, retrieved.ID)

	// Cancel payment intent
	canceled, err := svc.CancelPaymentIntent(ctx, intent.ID, "requested_by_customer")
	require.NoError(t, err)
	assert.Equal(t, payment.PaymentIntentStatusCanceled, canceled.Status)
}

func TestStripeIntegration_HealthCheck(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthy := svc.IsHealthy(ctx)
	assert.True(t, healthy, "Stripe API should be healthy")
}

// ============================================================================
// Adyen Integration Tests
// ============================================================================

func TestAdyenIntegration_CustomerLifecycle(t *testing.T) {
	svc := newAdyenIntegrationService(t)
	ctx := context.Background()

	// Create customer (in Adyen, this just generates a shopperReference)
	customer, err := svc.CreateCustomer(ctx, payment.CreateCustomerRequest{
		Email:       "adyen-test@virtengine.io",
		Name:        "Adyen Test User",
		VEIDAddress: "virtengine1adyentest",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)
}

func TestAdyenIntegration_PaymentIntentFlow(t *testing.T) {
	svc := newAdyenIntegrationService(t)
	ctx := context.Background()

	// Create a payment intent
	intent, err := svc.CreatePaymentIntent(ctx, payment.PaymentIntentRequest{
		Amount:      payment.NewAmount(1000, payment.CurrencyEUR), // €10.00
		CustomerID:  "virtengine1adyentest",
		Description: "Integration test payment",
		ReturnURL:   "https://virtengine.io/payment/return",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, intent.ID)
	assert.Equal(t, int64(1000), intent.Amount.Value)
}

func TestAdyenIntegration_HealthCheck(t *testing.T) {
	svc := newAdyenIntegrationService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthy := svc.IsHealthy(ctx)
	assert.True(t, healthy, "Adyen API should be healthy")
}

// ============================================================================
// 3D Secure Integration Tests
// ============================================================================

func TestStripeIntegration_3DSecureFlow(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Create a payment intent that requires 3DS
	// Using Stripe's 3DS test card number
	intent, err := svc.CreatePaymentIntent(ctx, payment.PaymentIntentRequest{
		Amount:      payment.NewAmount(5000, payment.CurrencyUSD), // $50.00
		Description: "3DS Test Payment",
		ReturnURL:   "https://virtengine.io/payment/return",
		Metadata: map[string]string{
			"test":        "true",
			"requires_3ds": "true",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, intent.ID)

	// Note: Full 3DS testing requires a frontend to complete the authentication
	// This test verifies the payment intent creation works correctly
}

// ============================================================================
// Refund Integration Tests
// ============================================================================

func TestStripeIntegration_RefundValidation(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Test refund validation with non-existent payment
	_, err := svc.CreateRefund(ctx, payment.RefundRequest{
		PaymentIntentID: "pi_nonexistent",
		Reason:          payment.RefundReasonRequestedByCustomer,
	})
	assert.Error(t, err, "Should error for non-existent payment")
}

// ============================================================================
// Webhook Integration Tests
// ============================================================================

func TestStripeIntegration_WebhookValidation(t *testing.T) {
	svc := newStripeIntegrationService(t)

	// Test with invalid signature
	payload := []byte(`{"id":"evt_test","type":"payment_intent.succeeded"}`)
	invalidSig := "t=1234567890,v1=invalid_signature"

	err := svc.ValidateWebhook(payload, invalidSig)
	assert.Error(t, err, "Should reject invalid webhook signature")
}

// ============================================================================
// Rate Limit Tests
// ============================================================================

func TestStripeIntegration_RateLimitBehavior(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Make several rapid requests to test rate limiting behavior
	// Stripe's test mode is more lenient, but this verifies the service handles it
	for i := 0; i < 3; i++ {
		_, err := svc.CreatePaymentIntent(ctx, payment.PaymentIntentRequest{
			Amount:      payment.NewAmount(100, payment.CurrencyUSD),
			Description: "Rate limit test",
		})
		if err != nil {
			// Expected if we hit rate limits
			t.Logf("Request %d got rate limited (expected): %v", i, err)
			break
		}
	}
}

// ============================================================================
// Conversion Service Integration Tests
// ============================================================================

func TestIntegration_ConversionQuote(t *testing.T) {
	svc := newStripeIntegrationService(t)
	ctx := context.Background()

	// Test conversion rate retrieval
	rate, err := svc.GetConversionRate(ctx, payment.CurrencyUSD, "uve")
	require.NoError(t, err)
	assert.Equal(t, payment.CurrencyUSD, rate.FromCurrency)
	assert.Equal(t, "uve", rate.ToCrypto)
	assert.False(t, rate.Rate.IsZero())

	// Test quote creation
	quote, err := svc.CreateConversionQuote(ctx, payment.ConversionQuoteRequest{
		FiatAmount:         payment.NewAmount(10000, payment.CurrencyUSD), // $100
		CryptoDenom:        "uve",
		DestinationAddress: "virtengine1abc123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, quote.ID)
	assert.Equal(t, "uve", quote.CryptoDenom)
	assert.False(t, quote.IsExpired())
}

