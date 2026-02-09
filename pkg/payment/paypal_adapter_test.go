package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPayPalAdapter_CreateConfirmRefund(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	})
	handler.HandleFunc("/v2/checkout/orders", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var req paypalOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "CAPTURE", req.Intent)

		_ = json.NewEncoder(w).Encode(paypalOrderResponse{
			ID:     "ORDER123",
			Status: "CREATED",
			Links: []paypalLink{
				{Rel: "approve", Href: "https://example.test/approve"},
			},
			PurchaseUnits: req.PurchaseUnits,
		})
	})
	handler.HandleFunc("/v2/checkout/orders/ORDER123/capture", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		_ = json.NewEncoder(w).Encode(paypalOrderCaptureResponse{
			ID:     "ORDER123",
			Status: "COMPLETED",
			PurchaseUnits: []paypalPurchaseUnitItem{
				{
					Payments: paypalPayments{
						Captures: []paypalCapture{
							{
								ID:     "CAPTURE123",
								Status: "COMPLETED",
								Amount: paypalAmount{CurrencyCode: "USD", Value: "10.00"},
							},
						},
					},
				},
			},
		})
	})
	handler.HandleFunc("/v2/payments/captures/CAPTURE123/refund", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		_ = json.NewEncoder(w).Encode(paypalRefundResponse{
			ID:         "REFUND123",
			Status:     "COMPLETED",
			Amount:     paypalAmount{CurrencyCode: "USD", Value: "5.00"},
			CreateTime: time.Now().UTC().Format(time.RFC3339),
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	gw, err := NewPayPalAdapter(PayPalConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      server.URL,
	})
	require.NoError(t, err)

	adapter := gw.(*PayPalAdapter)
	adapter.httpClient = server.Client()

	ctx := context.Background()
	intent, err := adapter.CreatePaymentIntent(ctx, PaymentIntentRequest{
		Amount:        NewAmount(1000, CurrencyUSD),
		CustomerID:    "cust",
		Description:   "test",
		ReceiptEmail:  "test@example.com",
		CaptureMethod: "automatic",
	})
	require.NoError(t, err)
	require.Equal(t, PaymentIntentStatusRequiresAction, intent.Status)
	require.True(t, strings.HasPrefix(intent.SCARedirectURL, "https://"))

	intent, err = adapter.ConfirmPaymentIntent(ctx, "ORDER123", "")
	require.NoError(t, err)
	require.Equal(t, PaymentIntentStatusSucceeded, intent.Status)
	require.Equal(t, "CAPTURE123", intent.PaymentMethodID)

	refund, err := adapter.CreateRefund(ctx, RefundRequest{
		PaymentIntentID: "CAPTURE123",
		Amount:          ptrAmount(NewAmount(500, CurrencyUSD)),
	})
	require.NoError(t, err)
	require.Equal(t, RefundStatusSucceeded, refund.Status)
}

func TestPayPalAdapter_WebhookVerification(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	})
	handler.HandleFunc("/v1/notifications/verify-webhook-signature", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"verification_status": "SUCCESS",
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	gw, err := NewPayPalAdapter(PayPalConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		WebhookID:    "wh_123",
		BaseURL:      server.URL,
	})
	require.NoError(t, err)

	adapter := gw.(*PayPalAdapter)
	adapter.httpClient = server.Client()

	payload := []byte(`{"id":"evt_1","event_type":"CHECKOUT.ORDER.COMPLETED","create_time":"2024-01-01T00:00:00Z"}`)
	signature := `{"transmission_id":"tx","transmission_time":"2024-01-01T00:00:00Z","cert_url":"https://example.test","auth_algo":"SHA256withRSA","transmission_sig":"sig","webhook_id":"wh_123"}`

	require.NoError(t, adapter.ValidateWebhook(payload, signature))

	event, err := adapter.ParseWebhookEvent(payload)
	require.NoError(t, err)
	require.Equal(t, WebhookEventPaymentIntentSucceeded, event.Type)
}

func ptrAmount(amount Amount) *Amount {
	return &amount
}
