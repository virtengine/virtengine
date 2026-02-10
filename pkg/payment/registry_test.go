package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockGateway struct {
	name    string
	gwType  GatewayType
	healthy bool
}

func (m mockGateway) Name() string                       { return m.name }
func (m mockGateway) Type() GatewayType                  { return m.gwType }
func (m mockGateway) IsHealthy(ctx context.Context) bool { return m.healthy }
func (m mockGateway) Close() error                       { return nil }
func (m mockGateway) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	return Customer{}, nil
}
func (m mockGateway) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	return Customer{}, nil
}
func (m mockGateway) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	return Customer{}, nil
}
func (m mockGateway) DeleteCustomer(ctx context.Context, customerID string) error { return nil }
func (m mockGateway) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	return "", nil
}
func (m mockGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return nil
}
func (m mockGateway) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	return nil, nil
}
func (m mockGateway) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	return PaymentIntent{}, nil
}
func (m mockGateway) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	return PaymentIntent{}, nil
}
func (m mockGateway) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	return PaymentIntent{}, nil
}
func (m mockGateway) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	return PaymentIntent{}, nil
}
func (m mockGateway) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	return PaymentIntent{}, nil
}
func (m mockGateway) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	return Refund{}, nil
}
func (m mockGateway) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	return Refund{}, nil
}
func (m mockGateway) ValidateWebhook(payload []byte, signature string) error { return nil }
func (m mockGateway) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	return WebhookEvent{}, nil
}

func TestPaymentProcessorRegistry_Fallback(t *testing.T) {
	registry := NewPaymentProcessorRegistry()

	registry.RegisterAdapter(mockGateway{name: "Stripe", gwType: GatewayStripe, healthy: false}, ProcessorRoute{
		Regions:    []string{"US"},
		Currencies: []Currency{CurrencyUSD},
		Fee:        FeeSchedule{FixedFee: 30, VariableBps: 290},
	})
	registry.RegisterAdapter(mockGateway{name: "Adyen", gwType: GatewayAdyen, healthy: true}, ProcessorRoute{
		Regions:    []string{"US"},
		Currencies: []Currency{CurrencyUSD},
		Fee:        FeeSchedule{FixedFee: 25, VariableBps: 300},
	})
	registry.RegisterAdapter(mockGateway{name: "PayPal", gwType: GatewayPayPal, healthy: true}, ProcessorRoute{
		Regions:    []string{"US"},
		Currencies: []Currency{CurrencyUSD},
		Fee:        FeeSchedule{FixedFee: 40, VariableBps: 350},
	})

	registry.SetProviderPreferences("provider-1", ProviderPreferences{
		Primary:   GatewayStripe,
		Secondary: GatewayAdyen,
		Tertiary:  GatewayPayPal,
	})

	adapter, err := registry.SelectAdapter(context.Background(), "provider-1", "US", NewAmount(1000, CurrencyUSD))
	require.NoError(t, err)
	require.Equal(t, GatewayAdyen, adapter.Type())
}

func TestPaymentProcessorRegistry_Optimal(t *testing.T) {
	registry := NewPaymentProcessorRegistry()

	registry.RegisterAdapter(mockGateway{name: "Stripe", gwType: GatewayStripe, healthy: true}, ProcessorRoute{
		Regions:    []string{"EU"},
		Currencies: []Currency{CurrencyEUR},
		Fee:        FeeSchedule{FixedFee: 30, VariableBps: 290},
	})
	registry.RegisterAdapter(mockGateway{name: "Adyen", gwType: GatewayAdyen, healthy: true}, ProcessorRoute{
		Regions:    []string{"EU"},
		Currencies: []Currency{CurrencyEUR},
		Fee:        FeeSchedule{FixedFee: 20, VariableBps: 250},
	})

	adapter, err := registry.SelectOptimalAdapter(context.Background(), "EU", NewAmount(1000, CurrencyEUR))
	require.NoError(t, err)
	require.Equal(t, GatewayAdyen, adapter.Type())
}
