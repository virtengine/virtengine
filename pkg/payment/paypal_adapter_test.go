package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayPalAdapter_CreateCaptureRefund(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token","expires_in":3600}`))
	})
	mux.HandleFunc("/v2/checkout/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"ORDER123",
			"status":"CREATED",
			"links":[{"href":"https://paypal.test/approve","rel":"approve"}],
			"purchase_units":[{"amount":{"currency_code":"USD","value":"10.00"}}]
		}`))
	})
	mux.HandleFunc("/v2/checkout/orders/ORDER123/capture", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"ORDER123",
			"status":"COMPLETED",
			"purchase_units":[{"payments":{"captures":[{"id":"CAPTURE123","status":"COMPLETED","amount":{"currency_code":"USD","value":"10.00"}}]}}]
		}`))
	})
	mux.HandleFunc("/v2/payments/captures/CAPTURE123/refund", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"REFUND123","status":"COMPLETED","amount":{"currency_code":"USD","value":"10.00"}}`))
	})
	mux.HandleFunc("/v1/notifications/verify-webhook-signature", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"verification_status":"SUCCESS"}`))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	cfg := PayPalConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		Environment:  "sandbox",
		BaseURL:      server.URL,
		WebhookID:    "wh_123",
	}

	adapter, err := NewPayPalAdapter(cfg)
	require.NoError(t, err)

	intent, err := adapter.CreatePaymentIntent(context.Background(), PaymentIntentRequest{
		Amount:    NewAmount(1000, CurrencyUSD),
		ReturnURL: "https://return",
	})
	require.NoError(t, err)
	assert.Equal(t, PaymentIntentStatusRequiresAction, intent.Status)
	assert.Equal(t, "https://paypal.test/approve", intent.SCARedirectURL)

	captured, err := adapter.ConfirmPaymentIntent(context.Background(), "ORDER123", "")
	require.NoError(t, err)
	assert.Equal(t, PaymentIntentStatusSucceeded, captured.Status)
	assert.Equal(t, int64(1000), captured.CapturedAmount.Value)

	refund, err := adapter.CreateRefund(context.Background(), RefundRequest{
		PaymentIntentID: "CAPTURE123",
		Amount:          ptrAmount(NewAmount(1000, CurrencyUSD)),
	})
	require.NoError(t, err)
	assert.Equal(t, RefundStatusSucceeded, refund.Status)

	signature, err := json.Marshal(paypalWebhookHeaders{
		TransmissionID:   "tid",
		TransmissionSig:  "sig",
		TransmissionTime: "2024-01-01T00:00:00Z",
		CertURL:          "https://cert",
		AuthAlgo:         "SHA256",
	})
	require.NoError(t, err)
	err = adapter.ValidateWebhook([]byte(`{"id":"evt_1"}`), string(signature))
	require.NoError(t, err)
}

func ptrAmount(amount Amount) *Amount {
	return &amount
}
