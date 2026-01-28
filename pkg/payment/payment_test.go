// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Type Tests
// ============================================================================

func TestGatewayType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		gateway  GatewayType
		expected bool
	}{
		{"stripe valid", GatewayStripe, true},
		{"adyen valid", GatewayAdyen, true},
		{"invalid gateway", GatewayType("invalid"), false},
		{"empty gateway", GatewayType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.gateway.IsValid())
		})
	}
}

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		expected bool
	}{
		{"USD valid", CurrencyUSD, true},
		{"EUR valid", CurrencyEUR, true},
		{"GBP valid", CurrencyGBP, true},
		{"JPY valid", CurrencyJPY, true},
		{"invalid currency", Currency("XYZ"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.currency.IsValid())
		})
	}
}

func TestCurrency_MinorUnitFactor(t *testing.T) {
	assert.Equal(t, int64(100), CurrencyUSD.MinorUnitFactor())
	assert.Equal(t, int64(100), CurrencyEUR.MinorUnitFactor())
	assert.Equal(t, int64(1), CurrencyJPY.MinorUnitFactor())
}

func TestAmount_Operations(t *testing.T) {
	t.Run("NewAmount", func(t *testing.T) {
		amount := NewAmount(1000, CurrencyUSD)
		assert.Equal(t, int64(1000), amount.Value)
		assert.Equal(t, CurrencyUSD, amount.Currency)
	})

	t.Run("NewAmountFromMajor", func(t *testing.T) {
		amount := NewAmountFromMajor(10.50, CurrencyUSD)
		assert.Equal(t, int64(1050), amount.Value)
	})

	t.Run("Major", func(t *testing.T) {
		amount := NewAmount(1050, CurrencyUSD)
		assert.Equal(t, 10.50, amount.Major())
	})

	t.Run("IsZero", func(t *testing.T) {
		assert.True(t, NewAmount(0, CurrencyUSD).IsZero())
		assert.False(t, NewAmount(100, CurrencyUSD).IsZero())
	})

	t.Run("IsPositive", func(t *testing.T) {
		assert.True(t, NewAmount(100, CurrencyUSD).IsPositive())
		assert.False(t, NewAmount(0, CurrencyUSD).IsPositive())
		assert.False(t, NewAmount(-100, CurrencyUSD).IsPositive())
	})

	t.Run("Add", func(t *testing.T) {
		a := NewAmount(1000, CurrencyUSD)
		b := NewAmount(500, CurrencyUSD)
		sum, err := a.Add(b)
		require.NoError(t, err)
		assert.Equal(t, int64(1500), sum.Value)
	})

	t.Run("Add different currencies", func(t *testing.T) {
		a := NewAmount(1000, CurrencyUSD)
		b := NewAmount(500, CurrencyEUR)
		_, err := a.Add(b)
		assert.ErrorIs(t, err, ErrInvalidCurrency)
	})

	t.Run("Sub", func(t *testing.T) {
		a := NewAmount(1000, CurrencyUSD)
		b := NewAmount(300, CurrencyUSD)
		diff, err := a.Sub(b)
		require.NoError(t, err)
		assert.Equal(t, int64(700), diff.Value)
	})
}

func TestCardBrand_IsSupported(t *testing.T) {
	assert.True(t, CardBrandVisa.IsSupported())
	assert.True(t, CardBrandMastercard.IsSupported())
	assert.False(t, CardBrandAmex.IsSupported())
	assert.False(t, CardBrandDiscover.IsSupported())
	assert.False(t, CardBrandUnknown.IsSupported())
}

func TestCardToken_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		token := CardToken{
			ExpiryMonth: int(time.Now().Month()) + 1,
			ExpiryYear:  time.Now().Year() + 1,
		}
		assert.False(t, token.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		token := CardToken{
			ExpiryMonth: 1,
			ExpiryYear:  2020,
		}
		assert.True(t, token.IsExpired())
	})
}

func TestCardToken_IsTokenExpired(t *testing.T) {
	t.Run("no expiry set", func(t *testing.T) {
		token := CardToken{}
		assert.False(t, token.IsTokenExpired())
	})

	t.Run("not expired", func(t *testing.T) {
		future := time.Now().Add(time.Hour)
		token := CardToken{ExpiresAt: &future}
		assert.False(t, token.IsTokenExpired())
	})

	t.Run("expired", func(t *testing.T) {
		past := time.Now().Add(-time.Hour)
		token := CardToken{ExpiresAt: &past}
		assert.True(t, token.IsTokenExpired())
	})
}

func TestCardToken_MaskedNumber(t *testing.T) {
	token := CardToken{Last4: "4242"}
	assert.Equal(t, "•••• •••• •••• 4242", token.MaskedNumber())
}

func TestPaymentIntentStatus_IsFinal(t *testing.T) {
	assert.True(t, PaymentIntentStatusSucceeded.IsFinal())
	assert.True(t, PaymentIntentStatusCanceled.IsFinal())
	assert.True(t, PaymentIntentStatusFailed.IsFinal())
	assert.False(t, PaymentIntentStatusProcessing.IsFinal())
	assert.False(t, PaymentIntentStatusRequiresAction.IsFinal())
}

func TestPaymentIntentStatus_IsSuccessful(t *testing.T) {
	assert.True(t, PaymentIntentStatusSucceeded.IsSuccessful())
	assert.False(t, PaymentIntentStatusFailed.IsSuccessful())
	assert.False(t, PaymentIntentStatusCanceled.IsSuccessful())
}

func TestPaymentIntent_CanRefund(t *testing.T) {
	t.Run("can refund", func(t *testing.T) {
		intent := PaymentIntent{
			Status:         PaymentIntentStatusSucceeded,
			CapturedAmount: NewAmount(1000, CurrencyUSD),
			RefundedAmount: NewAmount(0, CurrencyUSD),
		}
		assert.True(t, intent.CanRefund())
	})

	t.Run("cannot refund - not succeeded", func(t *testing.T) {
		intent := PaymentIntent{
			Status:         PaymentIntentStatusProcessing,
			CapturedAmount: NewAmount(1000, CurrencyUSD),
		}
		assert.False(t, intent.CanRefund())
	})

	t.Run("cannot refund - fully refunded", func(t *testing.T) {
		intent := PaymentIntent{
			Status:         PaymentIntentStatusSucceeded,
			CapturedAmount: NewAmount(1000, CurrencyUSD),
			RefundedAmount: NewAmount(1000, CurrencyUSD),
		}
		assert.False(t, intent.CanRefund())
	})
}

func TestPaymentIntent_RefundableAmount(t *testing.T) {
	intent := PaymentIntent{
		Amount:         NewAmount(1000, CurrencyUSD),
		CapturedAmount: NewAmount(1000, CurrencyUSD),
		RefundedAmount: NewAmount(300, CurrencyUSD),
	}
	refundable := intent.RefundableAmount()
	assert.Equal(t, int64(700), refundable.Value)
	assert.Equal(t, CurrencyUSD, refundable.Currency)
}

func TestConversionQuote_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		quote := ConversionQuote{
			ExpiresAt: time.Now().Add(time.Hour),
		}
		assert.False(t, quote.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		quote := ConversionQuote{
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		assert.True(t, quote.IsExpired())
	})
}

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, GatewayStripe, cfg.Gateway)
	assert.True(t, cfg.WebhookConfig.Enabled)
	assert.True(t, cfg.RateLimitConfig.Enabled)
	assert.True(t, cfg.ConversionConfig.Enabled)
	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
	assert.Len(t, cfg.SupportedCurrencies, 3)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid stripe config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.StripeConfig.SecretKey = "sk_test_xxx"
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid adyen config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Gateway = GatewayAdyen
		cfg.AdyenConfig.APIKey = "test_api_key"
		cfg.AdyenConfig.MerchantAccount = "TestMerchant"
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid gateway", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Gateway = GatewayType("invalid")
		assert.ErrorIs(t, cfg.Validate(), ErrGatewayNotConfigured)
	})

	t.Run("missing stripe key", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.StripeConfig.SecretKey = ""
		assert.ErrorIs(t, cfg.Validate(), ErrGatewayNotConfigured)
	})

	t.Run("missing adyen config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Gateway = GatewayAdyen
		assert.ErrorIs(t, cfg.Validate(), ErrGatewayNotConfigured)
	})

	t.Run("no supported currencies", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.StripeConfig.SecretKey = "sk_test_xxx"
		cfg.SupportedCurrencies = nil
		assert.ErrorIs(t, cfg.Validate(), ErrInvalidCurrency)
	})
}

func TestConfig_IsCurrencySupported(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.IsCurrencySupported(CurrencyUSD))
	assert.True(t, cfg.IsCurrencySupported(CurrencyEUR))
	assert.False(t, cfg.IsCurrencySupported(CurrencyJPY))
}

func TestConfig_ValidateAmount(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"

	t.Run("valid amount", func(t *testing.T) {
		amount := NewAmount(5000, CurrencyUSD)
		assert.NoError(t, cfg.ValidateAmount(amount))
	})

	t.Run("unsupported currency", func(t *testing.T) {
		amount := NewAmount(5000, CurrencyJPY)
		assert.ErrorIs(t, cfg.ValidateAmount(amount), ErrInvalidCurrency)
	})

	t.Run("below minimum", func(t *testing.T) {
		amount := NewAmount(50, CurrencyUSD) // $0.50, min is $1.00
		assert.ErrorIs(t, cfg.ValidateAmount(amount), ErrAmountBelowMinimum)
	})

	t.Run("above maximum", func(t *testing.T) {
		amount := NewAmount(20000000, CurrencyUSD) // $200,000, max is $100,000
		assert.ErrorIs(t, cfg.ValidateAmount(amount), ErrAmountAboveMaximum)
	})
}

// ============================================================================
// Service Tests
// ============================================================================

func TestNewService(t *testing.T) {
	t.Run("stripe service", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.StripeConfig.SecretKey = "sk_test_xxx"

		svc, err := NewService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, GatewayStripe, svc.Type())
		assert.Equal(t, "Stripe", svc.Name())
	})

	t.Run("adyen service", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Gateway = GatewayAdyen
		cfg.AdyenConfig.APIKey = "test_key"
		cfg.AdyenConfig.MerchantAccount = "TestMerchant"

		svc, err := NewService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, GatewayAdyen, svc.Type())
		assert.Equal(t, "Adyen", svc.Name())
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := Config{}
		_, err := NewService(cfg)
		assert.Error(t, err)
	})
}

func TestService_CustomerOperations(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("create customer", func(t *testing.T) {
		customer, err := svc.CreateCustomer(ctx, CreateCustomerRequest{
			Email:       "test@example.com",
			Name:        "Test User",
			VEIDAddress: "virtengine1xxx",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, customer.ID)
		assert.Equal(t, "test@example.com", customer.Email)
	})

	t.Run("get customer", func(t *testing.T) {
		customer, err := svc.GetCustomer(ctx, "cus_test123")
		require.NoError(t, err)
		assert.Equal(t, "cus_test123", customer.ID)
	})
}

func TestService_PaymentIntentOperations(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("create payment intent", func(t *testing.T) {
		intent, err := svc.CreatePaymentIntent(ctx, PaymentIntentRequest{
			Amount:      NewAmount(10000, CurrencyUSD),
			CustomerID:  "cus_xxx",
			Description: "Test payment",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, intent.ID)
		assert.Equal(t, int64(10000), intent.Amount.Value)
		assert.Equal(t, PaymentIntentStatusRequiresPaymentMethod, intent.Status)
	})

	t.Run("create payment intent with invalid amount", func(t *testing.T) {
		_, err := svc.CreatePaymentIntent(ctx, PaymentIntentRequest{
			Amount:     NewAmount(50, CurrencyUSD), // Below minimum
			CustomerID: "cus_xxx",
		})
		assert.ErrorIs(t, err, ErrAmountBelowMinimum)
	})

	t.Run("confirm payment intent", func(t *testing.T) {
		intent, err := svc.CreatePaymentIntent(ctx, PaymentIntentRequest{
			Amount:     NewAmount(5000, CurrencyUSD),
			CustomerID: "cus_xxx",
		})
		require.NoError(t, err)

		confirmed, err := svc.ConfirmPaymentIntent(ctx, intent.ID, "pm_xxx")
		require.NoError(t, err)
		assert.Equal(t, PaymentIntentStatusSucceeded, confirmed.Status)
	})

	t.Run("cancel payment intent", func(t *testing.T) {
		intent, err := svc.CreatePaymentIntent(ctx, PaymentIntentRequest{
			Amount:     NewAmount(5000, CurrencyUSD),
			CustomerID: "cus_xxx",
		})
		require.NoError(t, err)

		canceled, err := svc.CancelPaymentIntent(ctx, intent.ID, "customer_request")
		require.NoError(t, err)
		assert.Equal(t, PaymentIntentStatusCanceled, canceled.Status)
	})
}

func TestService_RefundOperations(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("create refund", func(t *testing.T) {
		refund, err := svc.CreateRefund(ctx, RefundRequest{
			PaymentIntentID: "pi_test123",
			Reason:          RefundReasonRequestedByCustomer,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, refund.ID)
		assert.Equal(t, RefundStatusSucceeded, refund.Status)
	})

	t.Run("create partial refund", func(t *testing.T) {
		amount := NewAmount(5000, CurrencyUSD)
		refund, err := svc.CreateRefund(ctx, RefundRequest{
			PaymentIntentID: "pi_test123",
			Amount:          &amount,
			Reason:          RefundReasonDuplicate,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(5000), refund.Amount.Value)
	})
}

func TestService_TokenValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		token := CardToken{
			Token:       "pm_xxx",
			Gateway:     GatewayStripe,
			Brand:       CardBrandVisa,
			ExpiryMonth: int(time.Now().Month()),
			ExpiryYear:  time.Now().Year() + 1,
		}
		assert.NoError(t, svc.ValidateToken(ctx, token))
	})

	t.Run("wrong gateway token", func(t *testing.T) {
		token := CardToken{
			Token:   "pm_xxx",
			Gateway: GatewayAdyen,
		}
		assert.ErrorIs(t, svc.ValidateToken(ctx, token), ErrInvalidCardToken)
	})

	t.Run("expired card", func(t *testing.T) {
		token := CardToken{
			Token:       "pm_xxx",
			Gateway:     GatewayStripe,
			Brand:       CardBrandVisa,
			ExpiryMonth: 1,
			ExpiryYear:  2020,
		}
		assert.ErrorIs(t, svc.ValidateToken(ctx, token), ErrCardExpired)
	})

	t.Run("unsupported card brand", func(t *testing.T) {
		token := CardToken{
			Token:       "pm_xxx",
			Gateway:     GatewayStripe,
			Brand:       CardBrandAmex,
			ExpiryMonth: int(time.Now().Month()),
			ExpiryYear:  time.Now().Year() + 1,
		}
		assert.ErrorIs(t, svc.ValidateToken(ctx, token), ErrPaymentDeclined)
	})
}

func TestService_WebhookHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	handlerCalled := false

	// Register handler
	svc.RegisterHandler(WebhookEventPaymentIntentSucceeded, func(ctx context.Context, event WebhookEvent) error {
		handlerCalled = true
		return nil
	})

	// Create and handle event
	event := NewWebhookEventBuilder().
		WithType(WebhookEventPaymentIntentSucceeded).
		WithGateway(GatewayStripe).
		Build()

	err = svc.HandleEvent(ctx, event)
	require.NoError(t, err)
	assert.True(t, handlerCalled)

	// Unregister and verify
	svc.UnregisterHandler(WebhookEventPaymentIntentSucceeded)
	handlerCalled = false
	err = svc.HandleEvent(ctx, event)
	require.NoError(t, err)
	assert.False(t, handlerCalled)
}

func TestService_SCAHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("initiate SCA", func(t *testing.T) {
		intent := PaymentIntent{
			ID:             "pi_test123",
			RequiresSCA:    true,
			SCARedirectURL: "https://stripe.com/3ds/xxx",
		}

		challenge, err := svc.InitiateSCA(ctx, intent)
		require.NoError(t, err)
		assert.Equal(t, "sca_pi_test123", challenge.ID)
		assert.Equal(t, intent.SCARedirectURL, challenge.RedirectURL)
	})

	t.Run("no SCA required", func(t *testing.T) {
		intent := PaymentIntent{
			ID:          "pi_test123",
			RequiresSCA: false,
		}

		challenge, err := svc.InitiateSCA(ctx, intent)
		require.NoError(t, err)
		assert.Empty(t, challenge.ID)
	})
}

func TestService_ConversionService(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StripeConfig.SecretKey = "sk_test_xxx"
	svc, err := NewService(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("get conversion rate", func(t *testing.T) {
		rate, err := svc.GetConversionRate(ctx, CurrencyUSD, "uve")
		require.NoError(t, err)
		assert.Equal(t, CurrencyUSD, rate.FromCurrency)
		assert.Equal(t, "uve", rate.ToCrypto)
	})

	t.Run("create conversion quote", func(t *testing.T) {
		quote, err := svc.CreateConversionQuote(ctx, ConversionQuoteRequest{
			FiatAmount:         NewAmount(10000, CurrencyUSD),
			CryptoDenom:        "uve",
			DestinationAddress: "virtengine1xxx",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, quote.ID)
		assert.Equal(t, "uve", quote.CryptoDenom)
		assert.Equal(t, "virtengine1xxx", quote.DestinationAddress)
		assert.False(t, quote.IsExpired())
	})
}

// ============================================================================
// Adapter Tests
// ============================================================================

func TestStripeAdapter(t *testing.T) {
	adapter, err := NewStripeAdapter(StripeConfig{
		SecretKey: "sk_test_xxx",
	})
	require.NoError(t, err)

	assert.Equal(t, "Stripe", adapter.Name())
	assert.Equal(t, GatewayStripe, adapter.Type())
	assert.True(t, adapter.IsHealthy(context.Background()))
}

func TestStripeAdapter_InvalidConfig(t *testing.T) {
	_, err := NewStripeAdapter(StripeConfig{})
	assert.ErrorIs(t, err, ErrGatewayNotConfigured)
}

func TestAdyenAdapter(t *testing.T) {
	adapter, err := NewAdyenAdapter(AdyenConfig{
		APIKey:          "test_key",
		MerchantAccount: "TestMerchant",
	})
	require.NoError(t, err)

	assert.Equal(t, "Adyen", adapter.Name())
	assert.Equal(t, GatewayAdyen, adapter.Type())
	assert.True(t, adapter.IsHealthy(context.Background()))
}

func TestAdyenAdapter_InvalidConfig(t *testing.T) {
	_, err := NewAdyenAdapter(AdyenConfig{})
	assert.ErrorIs(t, err, ErrGatewayNotConfigured)
}

// ============================================================================
// Webhook Server Tests
// ============================================================================

func TestWebhookEventBuilder(t *testing.T) {
	event := NewWebhookEventBuilder().
		WithType(WebhookEventPaymentIntentSucceeded).
		WithGateway(GatewayStripe).
		WithPaymentIntent(PaymentIntent{
			ID:     "pi_test",
			Status: PaymentIntentStatusSucceeded,
		}).
		Build()

	assert.Equal(t, WebhookEventPaymentIntentSucceeded, event.Type)
	assert.Equal(t, GatewayStripe, event.Gateway)
	assert.NotNil(t, event.Data)
}

// ============================================================================
// Rate Limiter Tests
// ============================================================================

func TestRateLimiter(t *testing.T) {
	cfg := RateLimitConfig{
		Enabled:              true,
		MaxPaymentsPerHour:   2,
		MaxRefundsPerDay:     2,
	}
	limiter := newRateLimiter(cfg)

	t.Run("payment limit", func(t *testing.T) {
		assert.NoError(t, limiter.checkPaymentLimit("cus_1"))
		assert.NoError(t, limiter.checkPaymentLimit("cus_1"))
		assert.ErrorIs(t, limiter.checkPaymentLimit("cus_1"), ErrRateLimitExceeded)

		// Different customer should work
		assert.NoError(t, limiter.checkPaymentLimit("cus_2"))
	})

	t.Run("refund limit", func(t *testing.T) {
		assert.NoError(t, limiter.checkRefundLimit())
		assert.NoError(t, limiter.checkRefundLimit())
		assert.ErrorIs(t, limiter.checkRefundLimit(), ErrRateLimitExceeded)
	})
}

func TestRateLimiter_Disabled(t *testing.T) {
	cfg := RateLimitConfig{Enabled: false}
	limiter := newRateLimiter(cfg)

	// Should never error when disabled
	for i := 0; i < 100; i++ {
		assert.NoError(t, limiter.checkPaymentLimit("cus_1"))
		assert.NoError(t, limiter.checkRefundLimit())
	}
}
