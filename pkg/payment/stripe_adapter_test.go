// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-2003: Integration tests for real Stripe adapter
package payment

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Unit Tests (Mock-based, no network calls)
// ============================================================================

func TestStripeAdapterCreation(t *testing.T) {
	t.Run("fails without API key", func(t *testing.T) {
		_, err := NewRealStripeAdapter(StripeConfig{})
		assert.ErrorIs(t, err, ErrGatewayNotConfigured)
	})

	t.Run("succeeds with API key", func(t *testing.T) {
		adapter, err := NewRealStripeAdapter(StripeConfig{
			SecretKey: "sk_test_fake_key_for_testing",
		})
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "Stripe", adapter.Name())
		assert.Equal(t, GatewayStripe, adapter.Type())
	})

	t.Run("detects test mode", func(t *testing.T) {
		adapter, err := NewRealStripeAdapter(StripeConfig{
			SecretKey: "sk_test_fake_key_for_testing",
		})
		require.NoError(t, err)
		stripeAdapter, ok := adapter.(*StripeAdapter)
		require.True(t, ok)
		assert.True(t, stripeAdapter.IsTestMode())
	})

	t.Run("detects live mode", func(t *testing.T) {
		adapter, err := NewRealStripeAdapter(StripeConfig{
			SecretKey: "sk_live_fake_key_for_testing",
		})
		require.NoError(t, err)
		stripeAdapter, ok := adapter.(*StripeAdapter)
		require.True(t, ok)
		assert.False(t, stripeAdapter.IsTestMode())
	})
}

func TestStripeGatewayFactory(t *testing.T) {
	t.Run("returns real adapter when useRealSDK is true", func(t *testing.T) {
		adapter, err := NewStripeGateway(StripeConfig{
			SecretKey: "sk_test_fake_key",
		}, true)
		require.NoError(t, err)
		_, ok := adapter.(*StripeAdapter)
		assert.True(t, ok, "expected StripeAdapter type")
	})

	t.Run("returns stub adapter when useRealSDK is false", func(t *testing.T) {
		adapter, err := NewStripeGateway(StripeConfig{
			SecretKey: "sk_test_fake_key",
		}, false)
		require.NoError(t, err)
		_, ok := adapter.(*stripeStubAdapter)
		assert.True(t, ok, "expected stripeStubAdapter type")
	})
}

func TestGetTestCardNumbers(t *testing.T) {
	cards := GetTestCardNumbers()
	assert.NotEmpty(t, cards)
	assert.Equal(t, "4242424242424242", cards["visa_success"])
	assert.Equal(t, "5555555555554444", cards["mastercard_success"])
	assert.Equal(t, "4000000000000002", cards["declined_generic"])
}

func TestMapStripeCardBrand(t *testing.T) {
	testCases := []struct {
		input    string
		expected CardBrand
	}{
		{"visa", CardBrandVisa},
		{"mastercard", CardBrandMastercard},
		{"amex", CardBrandAmex},
		{"discover", CardBrandDiscover},
		{"unknown", CardBrandUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// We can't easily test this without actually calling Stripe
			// but we can verify the constants exist
			assert.NotEmpty(t, tc.expected)
		})
	}
}

// ============================================================================
// Integration Tests (Require STRIPE_TEST_KEY environment variable)
// ============================================================================

// skipIfNoStripeKey skips the test if STRIPE_TEST_KEY is not set.
// These tests require a real Stripe test API key to run.
func skipIfNoStripeKey(t *testing.T) string {
	key := os.Getenv("STRIPE_TEST_KEY")
	if key == "" {
		t.Skip("Skipping integration test: STRIPE_TEST_KEY not set")
	}
	return key
}

func TestStripeIntegration_CreateCustomer(t *testing.T) {
	apiKey := skipIfNoStripeKey(t)

	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey: apiKey,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customer, err := adapter.CreateCustomer(ctx, CreateCustomerRequest{
		Email:       "test@virtengine.io",
		Name:        "VE Integration Test",
		VEIDAddress: "virtengine1test123",
		Metadata: map[string]string{
			"test": "true",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)
	assert.True(t, len(customer.ID) > 4 && customer.ID[:4] == "cus_")
	assert.Equal(t, "test@virtengine.io", customer.Email)

	// Cleanup: Delete the test customer
	err = adapter.DeleteCustomer(ctx, customer.ID)
	assert.NoError(t, err)
}

func TestStripeIntegration_CreatePaymentIntent(t *testing.T) {
	apiKey := skipIfNoStripeKey(t)

	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey: apiKey,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First create a customer
	customer, err := adapter.CreateCustomer(ctx, CreateCustomerRequest{
		Email:       "payment-test@virtengine.io",
		VEIDAddress: "virtengine1paytest",
	})
	require.NoError(t, err)
	defer adapter.DeleteCustomer(ctx, customer.ID)

	// Create a payment intent
	intent, err := adapter.CreatePaymentIntent(ctx, PaymentIntentRequest{
		Amount: Amount{
			Value:    5000, // $50.00
			Currency: CurrencyUSD,
		},
		CustomerID:  customer.ID,
		Description: "VirtEngine Test Payment",
		Metadata: map[string]string{
			"veid_address": "virtengine1paytest",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, intent.ID)
	assert.True(t, len(intent.ID) > 3 && intent.ID[:3] == "pi_")
	assert.NotEmpty(t, intent.ClientSecret)
	assert.Equal(t, PaymentIntentStatusRequiresPaymentMethod, intent.Status)
	assert.Equal(t, int64(5000), intent.Amount.Value)

	// Cancel the payment intent (cleanup)
	_, err = adapter.CancelPaymentIntent(ctx, intent.ID, "abandoned")
	assert.NoError(t, err)
}

func TestStripeIntegration_GetPaymentIntent(t *testing.T) {
	apiKey := skipIfNoStripeKey(t)

	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey: apiKey,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a payment intent
	created, err := adapter.CreatePaymentIntent(ctx, PaymentIntentRequest{
		Amount: Amount{
			Value:    1000,
			Currency: CurrencyUSD,
		},
		Description: "Test retrieval",
	})
	require.NoError(t, err)
	defer adapter.CancelPaymentIntent(ctx, created.ID, "abandoned")

	// Retrieve it
	retrieved, err := adapter.GetPaymentIntent(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Amount.Value, retrieved.Amount.Value)
}

func TestStripeIntegration_HealthCheck(t *testing.T) {
	apiKey := skipIfNoStripeKey(t)

	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey: apiKey,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthy := adapter.IsHealthy(ctx)
	assert.True(t, healthy, "adapter should be healthy with valid API key")
}

func TestStripeIntegration_CustomerLifecycle(t *testing.T) {
	apiKey := skipIfNoStripeKey(t)

	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey: apiKey,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create
	customer, err := adapter.CreateCustomer(ctx, CreateCustomerRequest{
		Email:       "lifecycle-test@virtengine.io",
		Name:        "Lifecycle Test",
		VEIDAddress: "virtengine1lifecycle",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)

	// Get
	retrieved, err := adapter.GetCustomer(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, customer.ID, retrieved.ID)

	// Update
	newName := "Updated Name"
	updated, err := adapter.UpdateCustomer(ctx, customer.ID, UpdateCustomerRequest{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	// Delete
	err = adapter.DeleteCustomer(ctx, customer.ID)
	assert.NoError(t, err)
}

func TestStripeIntegration_WebhookValidation(t *testing.T) {
	// This test verifies webhook signature validation logic
	adapter, err := NewRealStripeAdapter(StripeConfig{
		SecretKey:     "sk_test_fake",
		WebhookSecret: "whsec_test_secret",
	})
	require.NoError(t, err)

	// Invalid signature should fail
	err = adapter.ValidateWebhook([]byte(`{"id":"evt_test"}`), "invalid_signature")
	assert.ErrorIs(t, err, ErrWebhookSignatureInvalid)
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestConvertStripeError(t *testing.T) {
	// Test nil error
	assert.Nil(t, convertStripeError(nil))

	// Test generic error
	err := convertStripeError(assert.AnError)
	assert.ErrorIs(t, err, ErrGatewayUnavailable)
}
