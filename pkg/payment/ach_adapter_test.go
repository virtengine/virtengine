package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v80"
)

func TestACHAdapter_PaymentIntentAndRefund(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/v1/payment_intents", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, _ := ioReadAll(r)
		values, _ := url.ParseQuery(string(body))
		require.Equal(t, "1000", values.Get("amount"))
		require.Equal(t, "usd", values.Get("currency"))

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "pi_123",
			"amount":          1000,
			"currency":        "usd",
			"status":          "processing",
			"customer":        "cust_1",
			"payment_method":  "pm_1",
			"amount_received": 0,
			"amount_refunded": 0,
			"created":         time.Now().Unix(),
		})
	})
	handler.HandleFunc("/v1/payment_intents/pi_123/confirm", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "pi_123",
			"amount":          1000,
			"currency":        "usd",
			"status":          "succeeded",
			"customer":        "cust_1",
			"payment_method":  "pm_1",
			"amount_received": 1000,
			"amount_refunded": 0,
			"created":         time.Now().Unix(),
		})
	})
	handler.HandleFunc("/v1/refunds", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "re_123",
			"amount":   500,
			"currency": "usd",
			"status":   "succeeded",
			"created":  time.Now().Unix(),
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	gw, err := NewACHAdapter(ACHConfig{
		SecretKey: "sk_test",
		BaseURL:   server.URL + "/v1",
	})
	require.NoError(t, err)

	adapter := gw.(*ACHAdapter)
	adapter.httpClient = server.Client()

	ctx := context.Background()
	intent, err := adapter.CreatePaymentIntent(ctx, PaymentIntentRequest{
		Amount:          NewAmount(1000, CurrencyUSD),
		CustomerID:      "cust_1",
		PaymentMethodID: "pm_1",
	})
	require.NoError(t, err)
	require.Equal(t, PaymentIntentStatusProcessing, intent.Status)

	intent, err = adapter.ConfirmPaymentIntent(ctx, "pi_123", "")
	require.NoError(t, err)
	require.Equal(t, PaymentIntentStatusSucceeded, intent.Status)

	refund, err := adapter.CreateRefund(ctx, RefundRequest{
		PaymentIntentID: "pi_123",
		Amount:          ptrAmount(NewAmount(500, CurrencyUSD)),
	})
	require.NoError(t, err)
	require.Equal(t, RefundStatusSucceeded, refund.Status)
}

func TestACHAdapter_ValidateWebhook(t *testing.T) {
	gw, err := NewACHAdapter(ACHConfig{
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
		BaseURL:       "https://example.test",
	})
	require.NoError(t, err)

	adapter := gw.(*ACHAdapter)
	payload := []byte(fmt.Sprintf(`{"id":"evt_1","type":"payment_intent.succeeded","created":1700000000,"api_version":"%s"}`, stripe.APIVersion))
	timestamp := time.Now().Unix()
	sig := buildStripeSignature(adapter.config.WebhookSecret, payload, timestamp)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	require.NoError(t, adapter.ValidateWebhook(payload, header))
}

func buildStripeSignature(secret string, payload []byte, timestamp int64) string {
	signed := fmt.Sprintf("%d.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	return hex.EncodeToString(mac.Sum(nil))
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
