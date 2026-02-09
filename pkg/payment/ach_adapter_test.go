package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestACHAdapter_CreatePaymentIntentWithRetry(t *testing.T) {
	var debitCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/ach/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ach/verifications/micro_deposits", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"verified"}`))
	})
	mux.HandleFunc("/ach/debits", func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&debitCalls, 1) == 1 {
			http.Error(w, "temporary error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"debit_123","status":"succeeded","amount":1000,"currency":"USD","payment_method_id":"pm_123"}`))
	})
	mux.HandleFunc("/ach/refunds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"refund_123","payment_intent_id":"debit_123","status":"succeeded","amount":1000,"currency":"USD"}`))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	cfg := ACHConfig{
		SecretKey:          "ach_secret",
		BaseURL:            server.URL,
		RetryMaxAttempts:   2,
		RetryInitialDelay:  10 * time.Millisecond,
		RetryMaxDelay:      20 * time.Millisecond,
		RetryBackoffFactor: 2.0,
	}

	adapter, err := NewACHAdapter(cfg)
	require.NoError(t, err)

	intent, err := adapter.CreatePaymentIntent(context.Background(), PaymentIntentRequest{
		Amount:                 NewAmount(1000, CurrencyUSD),
		CustomerID:             "cust_1",
		BankVerificationMethod: "micro_deposit",
		BankAccount: &BankAccountDetails{
			AccountNumber: "000111222333",
			RoutingNumber: "110000000",
			AccountType:   "checking",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, PaymentIntentStatusSucceeded, intent.Status)
	assert.Equal(t, int64(2), int64(atomic.LoadInt32(&debitCalls)))

	refund, err := adapter.CreateRefund(context.Background(), RefundRequest{
		PaymentIntentID: intent.ID,
		Amount:          ptrAmount(NewAmount(1000, CurrencyUSD)),
	})
	require.NoError(t, err)
	assert.Equal(t, RefundStatusSucceeded, refund.Status)
}

func TestACHAdapter_WebhookValidationAndParsing(t *testing.T) {
	cfg := ACHConfig{
		SecretKey:     "ach_secret",
		WebhookSecret: "wh_secret",
		BaseURL:       "https://ach.test",
	}
	adapter, err := NewACHAdapter(cfg)
	require.NoError(t, err)

	payload := []byte(`{"id":"evt_1","type":"ach.debit.succeeded","created":1700000000,"data":{"object":{"id":"debit_123"}}}`)
	timestamp := "1700000000"
	mac := hmac.New(sha256.New, []byte(cfg.WebhookSecret))
	mac.Write([]byte(timestamp + "." + string(payload)))
	signature := "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))

	require.NoError(t, adapter.ValidateWebhook(payload, signature))

	event, err := adapter.ParseWebhookEvent(payload)
	require.NoError(t, err)
	assert.Equal(t, WebhookEventPaymentIntentSucceeded, event.Type)
	assert.Equal(t, GatewayACH, event.Gateway)
}

func TestACHAdapter_BuildNACHAFile(t *testing.T) {
	cfg := ACHConfig{
		SecretKey:        "ach_secret",
		NACHAOriginID:    "123456789",
		NACHACompanyName: "VirtEngine",
	}
	gateway, err := NewACHAdapter(cfg)
	require.NoError(t, err)
	adapter := gateway.(*ACHAdapter)

	file, err := adapter.BuildNACHAFile([]ACHEntry{
		{
			TransactionCode:   "27",
			RoutingNumber:     "110000000",
			AccountNumber:     "000111222333",
			Amount:            1000,
			AccountHolderName: "Alice",
			IndividualID:      "INV001",
		},
		{
			TransactionCode:   "27",
			RoutingNumber:     "110000000",
			AccountNumber:     "000111222444",
			Amount:            2000,
			AccountHolderName: "Bob",
			IndividualID:      "INV002",
		},
	}, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	lines := strings.Split(file, "\n")
	assert.GreaterOrEqual(t, len(lines), 4)
	assert.Equal(t, byte('1'), lines[0][0])
	assert.Equal(t, byte('5'), lines[1][0])
}
