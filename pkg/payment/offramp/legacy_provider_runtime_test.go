package offramp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/payment"
)

func TestConfigValidateRequiresBaseURLForDirectACH(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultProvider = ProviderACH
	cfg.ACHConfig.Provider = "dwolla"
	cfg.ACHConfig.SecretKey = "ach_secret"
	cfg.ACHConfig.BaseURL = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected direct ACH config validation to fail without base_url")
	}
}

func TestNewServiceUsesHTTPAMLClientWhenConfigured(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultProvider = ProviderACH
	cfg.ACHConfig.SecretKey = "ach_secret"
	cfg.AMLConfig.Provider = "sumsub"
	cfg.AMLConfig.APIURL = "https://aml.example.test"
	cfg.AMLConfig.APIKey = "aml_secret"

	svcRaw, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer svcRaw.Close()

	svc := svcRaw.(*offRampService)
	screener, ok := svc.amlScreener.(*DefaultAMLScreener)
	if !ok {
		t.Fatalf("expected DefaultAMLScreener, got %T", svc.amlScreener)
	}
	if _, ok := screener.client.(*HTTPAMLClient); !ok {
		t.Fatalf("expected HTTPAMLClient, got %T", screener.client)
	}
}

func TestNewServiceRejectsPartialAMLProviderConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultProvider = ProviderACH
	cfg.ACHConfig.SecretKey = "ach_secret"
	cfg.AMLConfig.Provider = "sumsub"
	cfg.AMLConfig.APIURL = "https://aml.example.test"
	cfg.AMLConfig.APIKey = ""

	if _, err := NewService(cfg); err == nil {
		t.Fatal("expected NewService() to fail when AML provider config is incomplete")
	}
}

func TestNewServiceUsesDirectACHProviderWhenConfigured(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultProvider = ProviderACH
	cfg.ACHConfig.Provider = "dwolla"
	cfg.ACHConfig.SecretKey = "ach_secret"
	cfg.ACHConfig.BaseURL = "https://direct-ach.example.test"
	cfg.AMLConfig.Enabled = false

	svcRaw, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer svcRaw.Close()

	svc := svcRaw.(*offRampService)
	provider := svc.providers[ProviderACH]
	if _, ok := provider.(*DirectACHAdapter); !ok {
		t.Fatalf("expected DirectACHAdapter, got %T", provider)
	}
}

func TestHTTPAMLClientScreenAndStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/screenings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /screenings, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer aml_secret" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		var req AMLClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode AML request: %v", err)
		}
		if req.FullName != "Alice Example" {
			t.Fatalf("unexpected AML full name: %s", req.FullName)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "scr_123",
			"status":      "processing",
			"risk_score":  12,
			"matches":     []any{},
			"screened_at": time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/screenings/scr_123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET /screenings/scr_123, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"screening_id": "scr_123",
			"status":       "completed",
			"risk_score":   0,
			"matches":      []any{},
			"screened_at":  time.Date(2026, 4, 12, 9, 5, 0, 0, time.UTC).Format(time.RFC3339),
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewHTTPAMLClient(AMLConfig{
		APIURL: server.URL,
		APIKey: "aml_secret",
	})
	if err != nil {
		t.Fatalf("NewHTTPAMLClient() error = %v", err)
	}
	client.httpClient = server.Client()

	resp, err := client.Screen(context.Background(), &AMLClientRequest{
		FullName:   "Alice Example",
		Country:    "US",
		EntityType: "individual",
	})
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}
	if resp.ScreeningID != "scr_123" {
		t.Fatalf("expected screening ID scr_123, got %s", resp.ScreeningID)
	}

	statusResp, err := client.GetScreeningStatus(context.Background(), "scr_123")
	if err != nil {
		t.Fatalf("GetScreeningStatus() error = %v", err)
	}
	if statusResp.Status != "completed" {
		t.Fatalf("expected completed status, got %s", statusResp.Status)
	}
}

func TestDirectACHAdapterOperations(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET /health, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/payouts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /payouts, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ach_secret" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "idem_123" {
			t.Fatalf("unexpected idempotency header: %s", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read direct ACH payout request: %v", err)
		}
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to decode direct ACH payout request: %v", err)
		}
		if req["amount"].(float64) != 12500 {
			t.Fatalf("unexpected payout amount: %v", req["amount"])
		}
		if req["currency"].(string) != "usd" {
			t.Fatalf("unexpected payout currency: %v", req["currency"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       "provider_payout_123",
			"batch_id": "batch_123",
			"status":   "processing",
		})
	})
	mux.HandleFunc("/payouts/provider_payout_123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET /payouts/provider_payout_123, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "provider_payout_123",
			"status": "settled",
		})
	})
	mux.HandleFunc("/payouts/provider_payout_123/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /payouts/provider_payout_123/cancel, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"canceled"}`))
	})
	mux.HandleFunc("/settlements", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("start_date"); got != "2026-04-01" {
			t.Fatalf("unexpected start_date query: %s", got)
		}
		if got := r.URL.Query().Get("end_date"); got != "2026-04-02" {
			t.Fatalf("unexpected end_date query: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"report_id":     "settlement_123",
			"start_date":    "2026-04-01",
			"end_date":      "2026-04-02",
			"generated_at":  "2026-04-03T00:00:00Z",
			"total_payouts": 1,
			"total_amount":  12500,
			"total_fees":    45,
			"transactions": []map[string]any{
				{
					"transaction_id": "txn_123",
					"payout_id":      "payout_local_123",
					"amount":         12500,
					"fee":            45,
					"status":         "settled",
					"processed_at":   "2026-04-02T10:00:00Z",
				},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	adapter, err := NewDirectACHAdapter(ACHConfig{
		Provider:        "dwolla",
		BaseURL:         server.URL,
		SecretKey:       "ach_secret",
		WebhookSecret:   "whsec_direct",
		SourceAccountID: "treasury_123",
	})
	if err != nil {
		t.Fatalf("NewDirectACHAdapter() error = %v", err)
	}
	adapter.httpClient = server.Client()

	if !adapter.IsHealthy(context.Background()) {
		t.Fatal("expected direct ACH adapter to be healthy")
	}

	intent := &PayoutIntent{
		ID:             "payout_local_123",
		Status:         PayoutStatusApproved,
		FiatAmount:     payment.Amount{Value: 12500, Currency: payment.CurrencyUSD},
		IdempotencyKey: "idem_123",
		Destination: PayoutDestination{
			Type: DestinationBankAccount,
			BankAccount: &BankAccountDetails{
				AccountHolderName: "Alice Example",
				AccountHolderType: "individual",
				RoutingNumber:     "110000000",
				AccountNumber:     "000123456789",
				AccountType:       "checking",
				Country:           "US",
			},
		},
	}
	if err := adapter.CreatePayout(context.Background(), intent); err != nil {
		t.Fatalf("CreatePayout() error = %v", err)
	}
	if intent.ProviderPayoutID != "provider_payout_123" {
		t.Fatalf("expected provider payout ID provider_payout_123, got %s", intent.ProviderPayoutID)
	}
	if intent.Status != PayoutStatusProcessing {
		t.Fatalf("expected processing status, got %s", intent.Status)
	}

	status, err := adapter.GetPayoutStatus(context.Background(), "provider_payout_123")
	if err != nil {
		t.Fatalf("GetPayoutStatus() error = %v", err)
	}
	if status != PayoutStatusSucceeded {
		t.Fatalf("expected succeeded status, got %s", status)
	}

	if err := adapter.CancelPayout(context.Background(), "provider_payout_123"); err != nil {
		t.Fatalf("CancelPayout() error = %v", err)
	}

	report, err := adapter.GetSettlementReport(context.Background(), SettlementReportRequest{
		StartDate: "2026-04-01",
		EndDate:   "2026-04-02",
	})
	if err != nil {
		t.Fatalf("GetSettlementReport() error = %v", err)
	}
	if report.ReportID != "settlement_123" {
		t.Fatalf("expected settlement report settlement_123, got %s", report.ReportID)
	}
	if report.TotalAmount != 12500 {
		t.Fatalf("expected settlement total amount 12500, got %d", report.TotalAmount)
	}
}

func TestDirectACHAdapterWebhookValidationAndParsing(t *testing.T) {
	adapter, err := NewDirectACHAdapter(ACHConfig{
		Provider:      "dwolla",
		BaseURL:       "https://direct-ach.example.test",
		SecretKey:     "ach_secret",
		WebhookSecret: "whsec_direct",
	})
	if err != nil {
		t.Fatalf("NewDirectACHAdapter() error = %v", err)
	}

	payload := []byte(`{"id":"evt_123","type":"payout.completed","timestamp":"2026-04-12T10:00:00Z","data":{"id":"provider_payout_123","payout_id":"payout_local_123","status":"succeeded","metadata":{"payout_id":"payout_local_123"}}}`)
	signature := buildDirectACHSignature(adapter.config.WebhookSecret, payload)

	if err := adapter.ValidateWebhook(payload, "sha256="+signature); err != nil {
		t.Fatalf("ValidateWebhook() error = %v", err)
	}

	event, err := adapter.ParseWebhookEvent(payload)
	if err != nil {
		t.Fatalf("ParseWebhookEvent() error = %v", err)
	}
	if event.Type != WebhookPayoutCompleted {
		t.Fatalf("expected payout completed event, got %s", event.Type)
	}
	if event.Status != PayoutStatusSucceeded {
		t.Fatalf("expected succeeded status, got %s", event.Status)
	}
	if event.PayoutID != "payout_local_123" {
		t.Fatalf("expected payout ID payout_local_123, got %s", event.PayoutID)
	}
}

func buildDirectACHSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestNewDirectACHAdapterRequiresSecretAndBaseURL(t *testing.T) {
	_, err := NewDirectACHAdapter(ACHConfig{
		Provider:  "dwolla",
		SecretKey: "",
		BaseURL:   "https://direct-ach.example.test",
	})
	if err == nil {
		t.Fatal("expected missing secret key to fail")
	}

	_, err = NewDirectACHAdapter(ACHConfig{
		Provider:  "dwolla",
		SecretKey: "ach_secret",
		BaseURL:   "",
	})
	if err == nil {
		t.Fatal("expected missing base URL to fail")
	}
}

func TestNewServiceUsesMockAMLClientWithoutProviderConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultProvider = ProviderACH
	cfg.ACHConfig.SecretKey = "ach_secret"
	cfg.AMLConfig.Enabled = true
	cfg.AMLConfig.Provider = ""
	cfg.AMLConfig.APIURL = ""
	cfg.AMLConfig.APIKey = ""

	svcRaw, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer svcRaw.Close()

	svc := svcRaw.(*offRampService)
	screener, ok := svc.amlScreener.(*DefaultAMLScreener)
	if !ok {
		t.Fatalf("expected DefaultAMLScreener, got %T", svc.amlScreener)
	}
	if _, ok := screener.client.(*MockAMLClient); !ok {
		t.Fatalf("expected MockAMLClient, got %T", screener.client)
	}
}

func TestDirectACHAdapterCancelMapsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/cancel") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		http.Error(w, "already settled", http.StatusConflict)
	}))
	defer server.Close()

	adapter, err := NewDirectACHAdapter(ACHConfig{
		Provider:  "dwolla",
		BaseURL:   server.URL,
		SecretKey: "ach_secret",
	})
	if err != nil {
		t.Fatalf("NewDirectACHAdapter() error = %v", err)
	}
	adapter.httpClient = server.Client()

	err = adapter.CancelPayout(context.Background(), "provider_payout_123")
	if err != ErrPayoutAlreadyProcessed {
		t.Fatalf("expected ErrPayoutAlreadyProcessed, got %v", err)
	}
}
