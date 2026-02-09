package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubGateway struct {
	name    string
	gateway GatewayType
	healthy bool
}

func (s *stubGateway) Name() string                       { return s.name }
func (s *stubGateway) Type() GatewayType                  { return s.gateway }
func (s *stubGateway) IsHealthy(ctx context.Context) bool { return s.healthy }
func (s *stubGateway) Close() error                       { return nil }
func (s *stubGateway) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	return Customer{ID: "cust"}, nil
}
func (s *stubGateway) GetCustomer(ctx context.Context, customerID string) (Customer, error) {
	return Customer{ID: customerID}, nil
}
func (s *stubGateway) UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (Customer, error) {
	return Customer{ID: customerID}, nil
}
func (s *stubGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	return nil
}
func (s *stubGateway) AttachPaymentMethod(ctx context.Context, customerID string, token CardToken) (string, error) {
	return token.Token, nil
}
func (s *stubGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return nil
}
func (s *stubGateway) ListPaymentMethods(ctx context.Context, customerID string) ([]CardToken, error) {
	return []CardToken{}, nil
}
func (s *stubGateway) CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (PaymentIntent, error) {
	return PaymentIntent{ID: "pi_1", Status: PaymentIntentStatusSucceeded}, nil
}
func (s *stubGateway) GetPaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	return PaymentIntent{ID: paymentIntentID, Status: PaymentIntentStatusSucceeded}, nil
}
func (s *stubGateway) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (PaymentIntent, error) {
	return PaymentIntent{ID: paymentIntentID, Status: PaymentIntentStatusSucceeded}, nil
}
func (s *stubGateway) CancelPaymentIntent(ctx context.Context, paymentIntentID string, reason string) (PaymentIntent, error) {
	return PaymentIntent{ID: paymentIntentID, Status: PaymentIntentStatusCanceled}, nil
}
func (s *stubGateway) CapturePaymentIntent(ctx context.Context, paymentIntentID string, amount *Amount) (PaymentIntent, error) {
	return PaymentIntent{ID: paymentIntentID, Status: PaymentIntentStatusSucceeded}, nil
}
func (s *stubGateway) CreateRefund(ctx context.Context, req RefundRequest) (Refund, error) {
	return Refund{ID: "re_1", Status: RefundStatusSucceeded}, nil
}
func (s *stubGateway) GetRefund(ctx context.Context, refundID string) (Refund, error) {
	return Refund{ID: refundID, Status: RefundStatusSucceeded}, nil
}
func (s *stubGateway) ValidateWebhook(payload []byte, signature string) error { return nil }
func (s *stubGateway) ParseWebhookEvent(payload []byte) (WebhookEvent, error) {
	return WebhookEvent{ID: "evt_1", Gateway: s.gateway}, nil
}

func TestRegistry_SelectAdapterFallback(t *testing.T) {
	registry := NewPaymentProcessorRegistry()
	registry.RegisterAdapter(GatewayStripe, &stubGateway{name: "stripe", gateway: GatewayStripe, healthy: false})
	registry.RegisterAdapter(GatewayPayPal, &stubGateway{name: "paypal", gateway: GatewayPayPal, healthy: true})
	registry.SetProviderPreference("provider-1", ProcessorPreference{
		Primary:   GatewayStripe,
		Secondary: GatewayPayPal,
	})

	adapter, err := registry.SelectAdapter(context.Background(), ProcessorSelectionRequest{
		ProviderID:     "provider-1",
		RequireHealthy: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, GatewayPayPal, adapter.Type())
}

func TestRegistry_SelectAdapterRegion(t *testing.T) {
	registry := NewPaymentProcessorRegistry()
	registry.RegisterAdapter(GatewayStripe, &stubGateway{name: "stripe", gateway: GatewayStripe, healthy: true})
	registry.SetRegionPreference("EU", ProcessorPreference{Primary: GatewayStripe})

	adapter, err := registry.SelectAdapter(context.Background(), ProcessorSelectionRequest{
		Region:         "EU",
		RequireHealthy: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, GatewayStripe, adapter.Type())
}

func TestRegistry_FeeRouting(t *testing.T) {
	registry := NewPaymentProcessorRegistry()
	registry.RegisterAdapter(GatewayStripe, &stubGateway{name: "stripe", gateway: GatewayStripe, healthy: true})
	registry.RegisterAdapter(GatewayAdyen, &stubGateway{name: "adyen", gateway: GatewayAdyen, healthy: true})
	registry.RegisterFees(GatewayStripe, FeeSchedule{FixedFee: 30, Percent: 2.9})
	registry.RegisterFees(GatewayAdyen, FeeSchedule{FixedFee: 10, Percent: 1.5})

	adapter, err := registry.SelectAdapter(context.Background(), ProcessorSelectionRequest{
		Amount:         NewAmount(10000, CurrencyUSD),
		RequireHealthy: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, GatewayAdyen, adapter.Type())
}
