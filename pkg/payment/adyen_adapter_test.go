// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3059: Tests for real Adyen adapter
package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Real Adyen Adapter Unit Tests
// ============================================================================

func TestRealAdyenAdapter_Creation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		adapter, err := NewRealAdyenAdapter(AdyenConfig{
			APIKey:          "test_api_key",
			MerchantAccount: "TestMerchant",
			Environment:     "test",
		})
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "Adyen", adapter.Name())
		assert.Equal(t, GatewayAdyen, adapter.Type())
	})

	t.Run("missing api key", func(t *testing.T) {
		_, err := NewRealAdyenAdapter(AdyenConfig{
			MerchantAccount: "TestMerchant",
		})
		assert.ErrorIs(t, err, ErrGatewayNotConfigured)
	})

	t.Run("missing merchant account", func(t *testing.T) {
		_, err := NewRealAdyenAdapter(AdyenConfig{
			APIKey: "test_api_key",
		})
		assert.ErrorIs(t, err, ErrGatewayNotConfigured)
	})

	t.Run("live environment requires prefix", func(t *testing.T) {
		_, err := NewRealAdyenAdapter(AdyenConfig{
			APIKey:          "test_api_key",
			MerchantAccount: "TestMerchant",
			Environment:     "live",
			// Missing LiveEndpointURLPrefix
		})
		assert.Error(t, err)
	})

	t.Run("live environment with prefix", func(t *testing.T) {
		adapter, err := NewRealAdyenAdapter(AdyenConfig{
			APIKey:                "live_api_key",
			MerchantAccount:       "LiveMerchant",
			Environment:           "live",
			LiveEndpointURLPrefix: "prefix123",
		})
		require.NoError(t, err)
		assert.NotNil(t, adapter)
	})
}

func TestRealAdyenAdapter_CustomerOperations(t *testing.T) {
	adapter, err := NewRealAdyenAdapter(AdyenConfig{
		APIKey:          "test_api_key",
		MerchantAccount: "TestMerchant",
		Environment:     "test",
	})
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("create customer", func(t *testing.T) {
		customer, err := adapter.CreateCustomer(ctx, CreateCustomerRequest{
			Email:       "test@example.com",
			Name:        "Test User",
			VEIDAddress: "virtengine1test",
		})
		require.NoError(t, err)
		// Adyen uses VEIDAddress as shopperReference
		assert.Equal(t, "virtengine1test", customer.ID)
		assert.Equal(t, "test@example.com", customer.Email)
	})

	t.Run("create customer without veid generates id", func(t *testing.T) {
		customer, err := adapter.CreateCustomer(ctx, CreateCustomerRequest{
			Email: "test2@example.com",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, customer.ID)
	})

	t.Run("get customer", func(t *testing.T) {
		customer, err := adapter.GetCustomer(ctx, "virtengine1test")
		require.NoError(t, err)
		assert.Equal(t, "virtengine1test", customer.ID)
	})

	t.Run("update customer", func(t *testing.T) {
		newEmail := "updated@example.com"
		customer, err := adapter.UpdateCustomer(ctx, "virtengine1test", UpdateCustomerRequest{
			Email: &newEmail,
		})
		require.NoError(t, err)
		assert.Equal(t, newEmail, customer.Email)
	})

	t.Run("delete customer is no-op", func(t *testing.T) {
		err := adapter.DeleteCustomer(ctx, "virtengine1test")
		assert.NoError(t, err)
	})
}

func TestRealAdyenAdapter_PaymentMethodOperations(t *testing.T) {
	adapter, err := NewRealAdyenAdapter(AdyenConfig{
		APIKey:          "test_api_key",
		MerchantAccount: "TestMerchant",
		Environment:     "test",
	})
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("attach payment method", func(t *testing.T) {
		pmID, err := adapter.AttachPaymentMethod(ctx, "cust_123", CardToken{
			Token: "pm_test123",
		})
		require.NoError(t, err)
		assert.Equal(t, "pm_test123", pmID)
	})

	t.Run("attach empty token fails", func(t *testing.T) {
		_, err := adapter.AttachPaymentMethod(ctx, "cust_123", CardToken{})
		assert.ErrorIs(t, err, ErrInvalidCardToken)
	})
}

func TestRealAdyenAdapter_WebhookParsing(t *testing.T) {
	adapter, err := NewRealAdyenAdapter(AdyenConfig{
		APIKey:          "test_api_key",
		MerchantAccount: "TestMerchant",
		Environment:     "test",
	})
	require.NoError(t, err)

	t.Run("parse authorisation success", func(t *testing.T) {
		payload := []byte(`{
			"notificationItems": [{
				"NotificationRequestItem": {
					"eventCode": "AUTHORISATION",
					"pspReference": "psp123",
					"success": "true",
					"merchantAccount": "TestMerchant",
					"amount": {
						"value": 1000,
						"currency": "EUR"
					}
				}
			}]
		}`)

		event, err := adapter.ParseWebhookEvent(payload)
		require.NoError(t, err)
		assert.Equal(t, WebhookEventPaymentIntentSucceeded, event.Type)
		assert.Equal(t, "psp123", event.ID)
		assert.Equal(t, GatewayAdyen, event.Gateway)
	})

	t.Run("parse authorisation failure", func(t *testing.T) {
		payload := []byte(`{
			"notificationItems": [{
				"NotificationRequestItem": {
					"eventCode": "AUTHORISATION",
					"pspReference": "psp456",
					"success": "false"
				}
			}]
		}`)

		event, err := adapter.ParseWebhookEvent(payload)
		require.NoError(t, err)
		assert.Equal(t, WebhookEventPaymentIntentFailed, event.Type)
	})

	t.Run("parse refund event", func(t *testing.T) {
		payload := []byte(`{
			"notificationItems": [{
				"NotificationRequestItem": {
					"eventCode": "REFUND",
					"pspReference": "psp789",
					"success": "true"
				}
			}]
		}`)

		event, err := adapter.ParseWebhookEvent(payload)
		require.NoError(t, err)
		assert.Equal(t, WebhookEventChargeRefunded, event.Type)
	})

	t.Run("parse chargeback event", func(t *testing.T) {
		payload := []byte(`{
			"notificationItems": [{
				"NotificationRequestItem": {
					"eventCode": "CHARGEBACK",
					"pspReference": "psp101"
				}
			}]
		}`)

		event, err := adapter.ParseWebhookEvent(payload)
		require.NoError(t, err)
		assert.Equal(t, WebhookEventChargeDisputeCreated, event.Type)
	})

	t.Run("parse empty notification", func(t *testing.T) {
		payload := []byte(`{"notificationItems": []}`)

		_, err := adapter.ParseWebhookEvent(payload)
		assert.ErrorIs(t, err, ErrWebhookEventUnknown)
	})

	t.Run("parse invalid json", func(t *testing.T) {
		payload := []byte(`{invalid json}`)

		_, err := adapter.ParseWebhookEvent(payload)
		assert.Error(t, err)
	})
}

func TestRealAdyenAdapter_WebhookValidation(t *testing.T) {
	t.Run("validation disabled when no hmac key", func(t *testing.T) {
		adapter, _ := NewRealAdyenAdapter(AdyenConfig{
			APIKey:          "test_api_key",
			MerchantAccount: "TestMerchant",
			Environment:     "test",
			// No HMACKey
		})

		err := adapter.ValidateWebhook([]byte("payload"), "signature")
		assert.NoError(t, err) // Validation is disabled
	})
}

func TestRealAdyenAdapter_TestMode(t *testing.T) {
	t.Run("test environment", func(t *testing.T) {
		adapter, _ := NewRealAdyenAdapter(AdyenConfig{
			APIKey:          "test_api_key",
			MerchantAccount: "TestMerchant",
			Environment:     "test",
		})
		realAdapter := adapter.(*RealAdyenAdapter)
		assert.True(t, realAdapter.IsTestMode())
	})

	t.Run("live environment", func(t *testing.T) {
		adapter, _ := NewRealAdyenAdapter(AdyenConfig{
			APIKey:                "live_api_key",
			MerchantAccount:       "LiveMerchant",
			Environment:           "live",
			LiveEndpointURLPrefix: "prefix",
		})
		realAdapter := adapter.(*RealAdyenAdapter)
		assert.False(t, realAdapter.IsTestMode())
	})
}

func TestGetAdyenTestCardNumbers(t *testing.T) {
	cards := GetAdyenTestCardNumbers()
	assert.NotEmpty(t, cards)
	assert.Contains(t, cards, "visa_success")
	assert.Contains(t, cards, "mastercard_success")
	assert.Contains(t, cards, "3ds_required")
}

func TestMapAdyenCardBrand(t *testing.T) {
	tests := []struct {
		input    string
		expected CardBrand
	}{
		{"visa", CardBrandVisa},
		{"VISA", CardBrandVisa},
		{"mc", CardBrandMastercard},
		{"mastercard", CardBrandMastercard},
		{"amex", CardBrandAmex},
		{"discover", CardBrandDiscover},
		{"unknown", CardBrandUnknown},
		{"", CardBrandUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapAdyenCardBrand(tt.input))
		})
	}
}
