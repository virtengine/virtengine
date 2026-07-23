package offramp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

type staticSecretResolver struct {
	api     []byte
	webhook map[string][]byte
}

func (r *staticSecretResolver) ResolveSecret(_ context.Context, _ SecretReference) ([]byte, error) {
	if len(r.api) == 0 {
		return nil, errors.New("missing secret")
	}
	return append([]byte(nil), r.api...), nil
}

func (r *staticSecretResolver) ResolveWebhookSecret(_ context.Context, ref string) ([]byte, error) {
	secret := r.webhook[ref]
	if len(secret) == 0 {
		return nil, errors.New("missing secret")
	}
	return append([]byte(nil), secret...), nil
}

type partnerFixture struct {
	t             *testing.T
	now           func() time.Time
	provider      string
	apiVersion    string
	quoteTTL      time.Duration
	fiatAmount    string
	quoteID       string
	payoutID      string
	quoteBinding  string
	lastQuote     partnerQuoteRequest
	lastPayout    partnerPayoutRequest
	payoutStatus  Status
	unknownQuote  bool
	oversized     bool
	badPayoutBind bool
	lookupStatus  int
	timeoutPayout bool
	mu            sync.Mutex
}

func (f *partnerFixture) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("X-API-Version") != f.apiVersion {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/partner/v1/health":
		require.NoError(f.t, json.NewEncoder(w).Encode(map[string]string{"status": "ok", "provider": f.provider, "api_version": f.apiVersion}))
	case r.Method == http.MethodPost && r.URL.Path == "/partner/v1/quotes":
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		require.NoError(f.t, decoder.Decode(&f.lastQuote))
		raw, err := json.Marshal(f.lastQuote)
		require.NoError(f.t, err)
		digest := sha256.Sum256(raw)
		f.quoteBinding = hex.EncodeToString(digest[:])
		response := map[string]any{
			"id": f.quoteID, "provider": f.provider, "api_version": f.apiVersion,
			"request_binding": f.quoteBinding, "fiat_amount": f.fiatAmount,
			"exchange_rate": f.fiatAmount, "fee": "10",
			"created_at": f.now(), "expires_at": f.now().Add(f.quoteTTL),
		}
		if f.unknownQuote {
			response["unexpected"] = true
		}
		if f.oversized {
			response["id"] = strings.Repeat("x", 2048)
		}
		require.NoError(f.t, json.NewEncoder(w).Encode(response))
	case r.Method == http.MethodPost && r.URL.Path == "/partner/v1/payouts":
		if r.Header.Get(defaultIdempotencyHeader) == "" || r.Header.Get(defaultCorrelationHeader) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		require.NoError(f.t, decoder.Decode(&f.lastPayout))
		if f.timeoutPayout {
			panic(http.ErrAbortHandler)
		}
		f.writePayout(w, f.payoutStatus)
	case r.Method == http.MethodPost && r.URL.Path == "/partner/v1/payouts/lookup":
		if f.lookupStatus != 0 {
			w.WriteHeader(f.lookupStatus)
			_ = json.NewEncoder(w).Encode(partnerErrorResponse{Code: "ambiguous_lookup", Retryable: true})
			return
		}
		f.writePayout(w, f.payoutStatus)
	case r.Method == http.MethodGet && r.URL.Path == "/partner/v1/payouts/"+f.payoutID:
		f.writePayout(w, f.payoutStatus)
	case r.Method == http.MethodPost && r.URL.Path == "/partner/v1/payouts/"+f.payoutID+"/cancel":
		f.payoutStatus = StatusCancelled
		f.writePayout(w, StatusCancelled)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *partnerFixture) writePayout(w http.ResponseWriter, status Status) {
	now := f.now()
	quoteID := f.quoteID
	correlationID := "corr-http-1"
	fiatAmount := f.fiatAmount
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": correlationID}
	if f.lastPayout.QuoteID != "" {
		quoteID = f.lastPayout.QuoteID
		correlationID = f.lastPayout.CorrelationID
		fiatAmount = f.lastPayout.FiatAmount
		metadata = f.lastPayout.Metadata
	}
	if f.badPayoutBind {
		quoteID = "wrong-quote"
	}
	response := partnerPayoutResponse{
		ID: f.payoutID, QuoteID: quoteID, Provider: f.provider, APIVersion: f.apiVersion,
		CorrelationID: correlationID, Status: status, FiatAmount: fiatAmount,
		CryptoAmount: "1000000", Fee: "10", Reference: "partner-reference", Metadata: metadata,
		InitiatedAt: now, StatusUpdatedAt: now,
	}
	if status == StatusCompleted {
		response.CompletedAt = &now
	}
	if status == StatusFailed {
		response.FailureCode = "provider_failed"
		response.FailureReason = "provider reported terminal failure"
	}
	require.NoError(f.t, json.NewEncoder(w).Encode(response))
}

func (f *partnerFixture) update(update func(*partnerFixture)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	update(f)
}

func (f *partnerFixture) snapshotPayout() partnerPayoutRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := f.lastPayout
	result.Metadata = cloneStringMap(result.Metadata)
	return result
}

func blockedSandboxProfile() PayoutProfile {
	return PayoutProfile{
		ID: "partner-us-usd-ach-sandbox", State: ProfileEngineeringCompleteExternalBlocked,
		Provider: "contract-partner", APIVersion: "2026-07-01", Environment: EnvironmentSandbox,
		Corridors: []PayoutCorridor{{
			ID: "US-USD-ach", Jurisdiction: "US", Currency: "USD", Rail: "ach",
			MinimumAmount: sdkmath.LegacyNewDec(10), MaximumAmount: sdkmath.LegacyNewDec(1000),
			DailyLimit: sdkmath.LegacyNewDec(1500), QuoteTTL: 2 * time.Minute, Finality: "provider_terminal_webhook",
		}},
		BeneficiaryRequirements: BeneficiaryRequirements{
			TokenizedReferenceRequired: true, ReferencePrefix: "beneficiary-token-",
			RequiredFields: []string{"beneficiary_reference"}, ProhibitedRawFields: []string{"account_number", "routing_number", "iban"},
		},
		DecisionRequirements: DecisionRequirements{KYCRequired: true, SanctionsRequired: true},
		CredentialSecretRefs: []SecretReference{{Purpose: "api", Ref: "vault://offramp/sandbox/api", Version: "1", Scope: string(EnvironmentSandbox)}},
		Webhook: WebhookProfile{
			Version: "2026-07-01", Algorithm: WebhookAlgorithmHMACSHA256,
			Keys: []WebhookKeyReference{
				{KeyID: "partner-key", Version: "1", SecretRef: "vault://offramp/sandbox/webhook/1"},
				{KeyID: "partner-key", Version: "2", SecretRef: "vault://offramp/sandbox/webhook/2"},
			},
		},
	}
}

func certifiedProductionProfile() PayoutProfile {
	profile := blockedSandboxProfile()
	profile.ID = "partner-us-usd-ach-production"
	profile.State = ProfileCertifiedEnabled
	profile.Environment = EnvironmentProduction
	profile.CredentialSecretRefs[0] = SecretReference{Purpose: "api", Ref: "vault://offramp/production/api", Version: "1", Scope: string(EnvironmentProduction)}
	profile.Webhook.Keys[0].SecretRef = "vault://offramp/production/webhook/1"
	profile.Webhook.Keys[1].SecretRef = "vault://offramp/production/webhook/2"
	approval := ApprovalEvidence{Reference: "evidence://approved", Owner: "named-owner"}
	profile.Evidence = ProfileEvidence{
		Contract: approval, Legal: approval, DPA: approval, Compliance: approval,
		Custody: approval, Banking: approval, WebhookRegistration: approval, Corridor: approval,
	}
	profile.Owners = ProfileOwners{Engineering: "engineering-owner", Operations: "operations-owner", Compliance: "compliance-owner", Security: "security-owner"}
	return profile
}

func httpQuoteRequest(now time.Time) QuoteRequest {
	return QuoteRequest{
		CryptoSymbol: "USDC", CryptoDenom: "uusdc", CryptoDecimals: 6, CryptoAmount: sdkmath.NewInt(1_000_000),
		FiatCurrency: "USD", PaymentMethod: "ach", Sender: "provider-ref",
		Destination: "beneficiary-token-1", BeneficiaryReference: "beneficiary-token-1",
		Jurisdiction: "US", CorrelationID: "corr-http-1",
		Compliance: ComplianceDecision{
			Reference: "decision-ref-1", KYCDecision: "approved", SanctionsDecision: "approved", ValidUntil: now.Add(time.Hour),
		},
	}
}

func newHTTPTestAdapter(t *testing.T, clock *testClock, profile PayoutProfile, fixtureOptions func(*partnerFixture)) (*HTTPPartnerAdapter, *partnerFixture) {
	t.Helper()
	fixture := &partnerFixture{
		t: t, now: clock.Now, provider: profile.Provider, apiVersion: profile.APIVersion,
		quoteTTL: time.Minute, fiatAmount: "100.000000000000000000", quoteID: "quote-http-1",
		payoutID: "payout-http-1", payoutStatus: StatusProcessing,
	}
	if fixtureOptions != nil {
		fixtureOptions(fixture)
	}
	server := httptest.NewServer(http.HandlerFunc(fixture.handler))
	t.Cleanup(server.Close)
	adapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: profile, BaseURL: server.URL, AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: PartnerEndpoints{
			Health: "/partner/v1/health", Quote: "/partner/v1/quotes", Payout: "/partner/v1/payouts",
			Status: "/partner/v1/payouts/{payout_id}", Cancel: "/partner/v1/payouts/{payout_id}/cancel",
			MetadataLookup: "/partner/v1/payouts/lookup",
		},
		Client: server.Client(), Clock: clock.Now,
		Secrets:                     &staticSecretResolver{api: []byte("test-api-secret-not-production")},
		AllowExternalBlockedSandbox: true,
	})
	require.NoError(t, err)
	return adapter, fixture
}

func TestHTTPPartnerAdapterLifecycle(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, fixture := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	ctx := context.Background()
	require.True(t, adapter.IsHealthy(ctx))
	quote, err := adapter.GetQuote(ctx, httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	require.Equal(t, "US-USD-ach", quote.AuditFields["corridor_id"])
	require.Equal(t, sdkmath.LegacyNewDec(100), quote.FiatAmount)
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"}
	started, err := adapter.InitiatePayout(ctx, PayoutRequest{Quote: quote, CryptoTxRef: "swap-tx-http-1", Destination: "beneficiary-token-1", Metadata: metadata})
	require.NoError(t, err)
	require.Equal(t, StatusProcessing, started.Status)
	require.Equal(t, "idem-http-1", metadata["idempotency_key"])
	status, err := adapter.GetStatus(ctx, started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusProcessing, status.Status)
	require.NoError(t, adapter.Cancel(ctx, started.ID))
	cancelled, err := adapter.GetStatus(ctx, started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCancelled, cancelled.Status)
	found, err := adapter.FindPayoutByMetadata(ctx, metadata)
	require.NoError(t, err)
	require.Equal(t, started.ID, found.ID)
	lastPayout := fixture.snapshotPayout()
	require.Equal(t, quote.ID, lastPayout.QuoteID)
	require.Equal(t, quote.AuditFields["request_binding"], lastPayout.RequestBinding)
	require.Equal(t, "100.000000000000000000", lastPayout.FiatAmount)
}

func TestHTTPPartnerAdapterRejectsExpiredOrModifiedQuote(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	modified := cloneQuote(quote)
	modified.FiatAmount = sdkmath.LegacyNewDec(101)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: modified, CryptoTxRef: "tx", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"},
	})
	require.ErrorIs(t, err, ErrInvalidRequest)
	clock.Advance(2 * time.Minute)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"},
	})
	require.ErrorIs(t, err, ErrQuoteExpired)
}

func TestHTTPPartnerAdapterRejectsUnsupportedAndComplianceFailures(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	tests := []struct {
		name string
		edit func(*QuoteRequest)
		want error
	}{
		{"jurisdiction", func(req *QuoteRequest) { req.Jurisdiction = "CA" }, ErrUnsupportedCorridor},
		{"currency", func(req *QuoteRequest) { req.FiatCurrency = "EUR" }, ErrUnsupportedCorridor},
		{"rail", func(req *QuoteRequest) { req.PaymentMethod = "wire" }, ErrUnsupportedCorridor},
		{"missing compliance", func(req *QuoteRequest) { req.Compliance.Reference = "" }, ErrComplianceRequired},
		{"revoked compliance", func(req *QuoteRequest) { req.Compliance.Revoked = true }, ErrComplianceRequired},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httpQuoteRequest(clock.Now())
			tc.edit(&req)
			_, err := adapter.GetQuote(context.Background(), req)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestHTTPPartnerAdapterLimits(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	tests := []struct {
		name   string
		amount string
	}{
		{"minimum", "9.000000000000000000"},
		{"maximum", "1001.000000000000000000"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), func(f *partnerFixture) { f.fiatAmount = tc.amount })
			_, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
			require.ErrorIs(t, err, ErrLimitExceeded)
		})
	}
	profile := blockedSandboxProfile()
	profile.Corridors[0].MaximumAmount = sdkmath.LegacyNewDec(100)
	profile.Corridors[0].DailyLimit = sdkmath.LegacyNewDec(150)
	adapter, fixture := newHTTPTestAdapter(t, clock, profile, nil)
	first, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: first, CryptoTxRef: "tx-1", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"},
	})
	require.NoError(t, err)
	fixture.update(func(f *partnerFixture) { f.quoteID = "quote-http-2" })
	secondReq := httpQuoteRequest(clock.Now())
	secondReq.CorrelationID = "corr-http-2"
	second, err := adapter.GetQuote(context.Background(), secondReq)
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: second, CryptoTxRef: "tx-2", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-2", "correlation_id": "corr-http-2"},
	})
	require.ErrorIs(t, err, ErrLimitExceeded)
}

func TestHTTPPartnerAdapterStrictTransportAndResponseBounds(t *testing.T) {
	t.Parallel()
	profile := certifiedProductionProfile()
	_, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: profile, BaseURL: "http://partner.example", AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: PartnerEndpoints{Health: "/partner/v1/health", Quote: "/partner/v1/quotes", Payout: "/partner/v1/payouts", Status: "/partner/v1/payouts/{payout_id}", Cancel: "/partner/v1/payouts/{payout_id}/cancel", MetadataLookup: "/partner/v1/payouts/lookup"},
		Secrets:   &staticSecretResolver{api: []byte("secret")},
	})
	require.Error(t, err)

	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	for _, tc := range []struct {
		name             string
		configureFixture func(*partnerFixture)
		want             error
	}{
		{"unknown field", func(f *partnerFixture) { f.unknownQuote = true }, ErrProviderRejected},
		{"oversized response", func(f *partnerFixture) { f.oversized = true }, ErrResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), tc.configureFixture)
			if tc.name == "oversized response" {
				adapter.maxResponseBytes = 512
			}
			_, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestHTTPPartnerAdapterAmbiguousInitiationRecoveredByMetadata(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, fixture := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), func(f *partnerFixture) { f.timeoutPayout = true })
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"}
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1", Metadata: metadata})
	require.True(t, IsAmbiguous(err), fmt.Sprintf("expected ambiguous error, got %v", err))
	fixture.update(func(f *partnerFixture) {
		f.timeoutPayout = false
		f.lastPayout = partnerPayoutRequest{
			QuoteID: quote.ID, CorrelationID: "corr-http-1", FiatAmount: quote.FiatAmount.String(), Metadata: metadata,
		}
	})
	recovered, err := adapter.FindPayoutByMetadata(context.Background(), metadata)
	require.NoError(t, err)
	require.Equal(t, fixture.payoutID, recovered.ID)
}

func TestHTTPPartnerAdapterRestartRestoresDurableBindingAndPolls(t *testing.T) {
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	profile := blockedSandboxProfile()
	fixture := &partnerFixture{t: t, now: clock.Now, provider: profile.Provider, apiVersion: profile.APIVersion, quoteTTL: time.Minute, fiatAmount: "100.000000000000000000", quoteID: "quote-http-1", payoutID: "payout-http-1", payoutStatus: StatusProcessing}
	server := httptest.NewServer(http.HandlerFunc(fixture.handler))
	t.Cleanup(server.Close)
	newAdapter := func() *HTTPPartnerAdapter {
		adapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{Profile: profile, BaseURL: server.URL, AllowedPathPrefixes: []string{"/partner/v1"}, Endpoints: testPartnerEndpoints(), Client: server.Client(), Clock: clock.Now, Secrets: &staticSecretResolver{api: []byte("secret")}, AllowExternalBlockedSandbox: true})
		require.NoError(t, err)
		return adapter
	}
	repository := newDurableTestPayoutRepository()
	firstBridge := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
	require.NoError(t, firstBridge.RegisterAdapter(newAdapter()))
	quote, err := firstBridge.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1", "conversion_id": "conversion-http-1"}
	started, err := firstBridge.InitiatePayout(context.Background(), quote, "tx", "beneficiary-token-1", metadata)
	require.NoError(t, err)
	require.NotEmpty(t, started.DailyReservationKey)
	fixture.update(func(f *partnerFixture) { f.payoutStatus = StatusCompleted })
	restartedBridge := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
	require.NoError(t, restartedBridge.RegisterAdapter(newAdapter()))
	completed, err := restartedBridge.GetStatus(context.Background(), started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, completed.Status)
	require.Equal(t, "durable_binding_restore", completed.AuditFields["bridge_recovery_reason"])
	require.Equal(t, started.DailyReservationKey, completed.DailyReservationKey)
}

func TestHTTPPartnerAdapterRestartFailureReleasesExactDailyReservation(t *testing.T) {
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	profile := blockedSandboxProfile()
	profile.Corridors[0].MaximumAmount = sdkmath.LegacyNewDec(100)
	profile.Corridors[0].DailyLimit = sdkmath.LegacyNewDec(100)
	limits := NewMemoryDailyLimitRepository()
	fixture := &partnerFixture{t: t, now: clock.Now, provider: profile.Provider, apiVersion: profile.APIVersion, quoteTTL: time.Minute, fiatAmount: "100.000000000000000000", quoteID: "quote-http-1", payoutID: "payout-http-1", payoutStatus: StatusProcessing}
	server := httptest.NewServer(http.HandlerFunc(fixture.handler))
	t.Cleanup(server.Close)
	newAdapter := func() *HTTPPartnerAdapter {
		adapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{Profile: profile, BaseURL: server.URL, AllowedPathPrefixes: []string{"/partner/v1"}, Endpoints: testPartnerEndpoints(), Client: server.Client(), Clock: clock.Now, Secrets: &staticSecretResolver{api: []byte("secret")}, DailyLimits: limits, AllowExternalBlockedSandbox: true})
		require.NoError(t, err)
		return adapter
	}
	repository := newDurableTestPayoutRepository()
	first := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
	require.NoError(t, first.RegisterAdapter(newAdapter()))
	quote, err := first.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	started, err := first.InitiatePayout(context.Background(), quote, "tx", "beneficiary-token-1", map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1", "conversion_id": "conversion-http-1"})
	require.NoError(t, err)
	fixture.update(func(f *partnerFixture) { f.payoutStatus = StatusFailed })
	restarted := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
	require.NoError(t, restarted.RegisterAdapter(newAdapter()))
	failed, err := restarted.GetStatus(context.Background(), started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusFailed, failed.Status)
	reserved, err := limits.ReserveDailyAmount(context.Background(), started.DailyReservationKey, "replacement-operation", sdkmath.LegacyNewDec(100), sdkmath.LegacyNewDec(100))
	require.NoError(t, err)
	require.True(t, reserved, "terminal restart reconciliation must release the original exact reservation")
}

func TestHTTPPartnerAdapterRestartRecoveryRejectsMismatchAndAmbiguousLookup(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*partnerFixture)
	}{
		{name: "mismatched payout", mutate: func(f *partnerFixture) { f.payoutID = "different-payout" }},
		{name: "mismatched quote", mutate: func(f *partnerFixture) { f.badPayoutBind = true }},
		{name: "ambiguous duplicate lookup", mutate: func(f *partnerFixture) { f.lookupStatus = http.StatusConflict }},
	} {
		t.Run(test.name, func(t *testing.T) {
			clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
			profile := blockedSandboxProfile()
			fixture := &partnerFixture{t: t, now: clock.Now, provider: profile.Provider, apiVersion: profile.APIVersion, quoteTTL: time.Minute, fiatAmount: "100.000000000000000000", quoteID: "quote-http-1", payoutID: "payout-http-1", payoutStatus: StatusProcessing}
			server := httptest.NewServer(http.HandlerFunc(fixture.handler))
			defer server.Close()
			newAdapter := func() *HTTPPartnerAdapter {
				adapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{Profile: profile, BaseURL: server.URL, AllowedPathPrefixes: []string{"/partner/v1"}, Endpoints: testPartnerEndpoints(), Client: server.Client(), Clock: clock.Now, Secrets: &staticSecretResolver{api: []byte("secret")}, AllowExternalBlockedSandbox: true})
				require.NoError(t, err)
				return adapter
			}
			repository := newDurableTestPayoutRepository()
			first := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
			require.NoError(t, first.RegisterAdapter(newAdapter()))
			quote, err := first.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
			require.NoError(t, err)
			started, err := first.InitiatePayout(context.Background(), quote, "tx", "beneficiary-token-1", map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1", "conversion_id": "conversion-http-1"})
			require.NoError(t, err)
			fixture.update(test.mutate)
			restarted := newBridgeWithDependencies(ExecutionModeSandbox, true, repository, nil, clock.Now)
			require.NoError(t, restarted.RegisterAdapter(newAdapter()))
			_, err = restarted.GetStatus(context.Background(), started.ID)
			require.Error(t, err)
			if test.name == "ambiguous duplicate lookup" {
				require.True(t, IsAmbiguous(err))
			} else {
				require.ErrorIs(t, err, ErrProviderRejected)
			}
		})
	}
}

func TestHTTPPartnerAdapterRejectsProviderBindingMismatch(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), func(f *partnerFixture) { f.badPayoutBind = true })
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"},
	})
	require.ErrorIs(t, err, ErrProviderRejected)
}

func TestHTTPPartnerAdapterRequiresExplicitSandboxOptInAndDurableProductionLimits(t *testing.T) {
	t.Parallel()
	profile := blockedSandboxProfile()
	_, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: profile, BaseURL: "https://partner.example", AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: testPartnerEndpoints(), Secrets: &staticSecretResolver{api: []byte("secret")},
	})
	require.ErrorIs(t, err, ErrProfileNotExecutable)

	production := certifiedProductionProfile()
	_, err = NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: production, BaseURL: "https://partner.example", AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: testPartnerEndpoints(), Secrets: &staticSecretResolver{api: []byte("secret")},
		DailyLimits: NewMemoryDailyLimitRepository(),
	})
	require.ErrorIs(t, err, ErrProfileNotExecutable)
}

func TestHTTPPartnerAdapterRejectsRawBeneficiaryMetadataAndRedirects(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1", "bank_account": "raw-account"},
	})
	require.ErrorIs(t, err, ErrInvalidRequest)

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://untrusted.example", http.StatusFound)
	}))
	t.Cleanup(redirectServer.Close)
	redirectAdapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: blockedSandboxProfile(), BaseURL: redirectServer.URL, AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: testPartnerEndpoints(), Client: redirectServer.Client(),
		Secrets: &staticSecretResolver{api: []byte("secret")}, AllowExternalBlockedSandbox: true,
	})
	require.NoError(t, err)
	_, err = redirectAdapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.Error(t, err)
}

func TestQuoteExpiresAtExactBoundary(t *testing.T) {
	t.Parallel()
	expiry := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	require.True(t, (Quote{ExpiresAt: expiry}).IsExpired(expiry))
}

func TestHTTPPartnerAdapterRejectsNonCanonicalDecimal(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, _ := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), func(f *partnerFixture) {
		f.fiatAmount = "1e2"
	})
	_, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func TestHTTPPartnerAdapterRetryableInitiationResponseIsAmbiguous(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	profile := blockedSandboxProfile()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/partner/v1/health" {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "provider": profile.Provider, "api_version": profile.APIVersion})
			return
		}
		if r.URL.Path == "/partner/v1/quotes" {
			var request partnerQuoteRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			binding, err := hashStrictJSON(request)
			require.NoError(t, err)
			_ = json.NewEncoder(w).Encode(partnerQuoteResponse{
				ID: "quote-http-1", Provider: profile.Provider, APIVersion: profile.APIVersion, RequestBinding: binding,
				FiatAmount: "100.000000000000000000", ExchangeRate: "100.000000000000000000", Fee: "10",
				CreatedAt: clock.Now(), ExpiresAt: clock.Now().Add(time.Minute),
			})
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(partnerErrorResponse{Code: "temporary", Retryable: true})
	}))
	t.Cleanup(server.Close)
	adapter, err := NewHTTPPartnerAdapter(HTTPPartnerConfig{
		Profile: profile, BaseURL: server.URL, AllowedPathPrefixes: []string{"/partner/v1"},
		Endpoints: testPartnerEndpoints(), Client: server.Client(), Clock: clock.Now,
		Secrets: &staticSecretResolver{api: []byte("secret")}, AllowExternalBlockedSandbox: true,
	})
	require.NoError(t, err)
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{
		Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1",
		Metadata: map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"},
	})
	require.True(t, IsAmbiguous(err))
}

func TestHTTPPartnerAdapterRejectsStatusRegression(t *testing.T) {
	t.Parallel()
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, fixture := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"}
	started, err := adapter.InitiatePayout(context.Background(), PayoutRequest{Quote: quote, CryptoTxRef: "tx", Destination: "beneficiary-token-1", Metadata: metadata})
	require.NoError(t, err)
	fixture.update(func(f *partnerFixture) { f.payoutStatus = StatusPending })
	_, err = adapter.GetStatus(context.Background(), started.ID)
	require.ErrorIs(t, err, ErrProviderRejected)
}

func TestHTTPPartnerAdapterRejectsIdempotencyKeyMutation(t *testing.T) {
	clock := newTestClock(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC))
	adapter, fixture := newHTTPTestAdapter(t, clock, blockedSandboxProfile(), nil)
	quote, err := adapter.GetQuote(context.Background(), httpQuoteRequest(clock.Now()))
	require.NoError(t, err)
	metadata := map[string]string{"idempotency_key": "idem-http-1", "correlation_id": "corr-http-1"}
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{Quote: quote, CryptoTxRef: "tx-1", Destination: "beneficiary-token-1", Metadata: metadata})
	require.NoError(t, err)
	fixture.update(func(f *partnerFixture) { f.quoteID = "quote-http-2" })
	request := httpQuoteRequest(clock.Now())
	request.CorrelationID = "corr-http-2"
	second, err := adapter.GetQuote(context.Background(), request)
	require.NoError(t, err)
	_, err = adapter.InitiatePayout(context.Background(), PayoutRequest{Quote: second, CryptoTxRef: "tx-2", Destination: "beneficiary-token-1", Metadata: map[string]string{
		"idempotency_key": "idem-http-1", "correlation_id": "corr-http-2",
	}})
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func testPartnerEndpoints() PartnerEndpoints {
	return PartnerEndpoints{
		Health: "/partner/v1/health", Quote: "/partner/v1/quotes", Payout: "/partner/v1/payouts",
		Status: "/partner/v1/payouts/{payout_id}", Cancel: "/partner/v1/payouts/{payout_id}/cancel",
		MetadataLookup: "/partner/v1/payouts/lookup",
	}
}
