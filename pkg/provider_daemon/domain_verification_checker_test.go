package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/provider/keeper"
)

// mockDNSResolver implements DNSResolver for testing.
type mockDNSResolver struct {
	txtRecords   map[string][]string
	cnameRecords map[string]string
	err          error
}

func (m *mockDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	if records, ok := m.txtRecords[name]; ok {
		return records, nil
	}
	return []string{}, nil
}

func (m *mockDNSResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if cname, ok := m.cnameRecords[host]; ok {
		return cname, nil
	}
	return "", fmt.Errorf("no such host")
}

// mockChainClient implements ChainClient for testing.
type mockChainClient struct {
	providerConfig *ProviderConfig
	err            error
}

func (m *mockChainClient) GetProviderConfig(ctx context.Context, address string) (*ProviderConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.providerConfig, nil
}

func (m *mockChainClient) GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]Order, error) {
	return nil, m.err
}

func (m *mockChainClient) PlaceBid(ctx context.Context, bid *Bid, signature *Signature) error {
	return m.err
}

func (m *mockChainClient) GetProviderBids(ctx context.Context, address string) ([]Bid, error) {
	return nil, m.err
}

func TestNewDomainVerificationChecker(t *testing.T) {
	tests := []struct {
		name        string
		config      DomainVerificationCheckerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: DomainVerificationCheckerConfig{
				Enabled:         true,
				ProviderAddress: "virtengine1abcd1234",
				CometRPC:        "http://localhost:26657",
				GRPCEndpoint:    "localhost:9090",
			},
			expectError: false,
		},
		{
			name: "disabled",
			config: DomainVerificationCheckerConfig{
				Enabled: false,
			},
			expectError: true,
			errorMsg:    "disabled",
		},
		{
			name: "missing provider address",
			config: DomainVerificationCheckerConfig{
				Enabled:  true,
				CometRPC: "http://localhost:26657",
			},
			expectError: true,
			errorMsg:    "provider address is required",
		},
		{
			name: "missing rpc endpoint",
			config: DomainVerificationCheckerConfig{
				Enabled:         true,
				ProviderAddress: "virtengine1abcd1234",
			},
			expectError: true,
			errorMsg:    "comet RPC endpoint or chain client is required",
		},
		{
			name: "with chain client",
			config: DomainVerificationCheckerConfig{
				Enabled:         true,
				ProviderAddress: "virtengine1abcd1234",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var chainClient ChainClient
			if tt.name == "with chain client" {
				chainClient = &mockChainClient{}
			}

			checker, err := NewDomainVerificationChecker(tt.config, nil, chainClient)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, checker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checker)
			}
		})
	}
}

func TestVerifyDNSTXT(t *testing.T) {
	providerAddr := "virtengine1abcd1234"
	domain := "example.com"
	token := "abc123def456"
	verificationName := fmt.Sprintf("%s.%s", keeper.DNSVerificationPrefix, domain)

	tests := []struct {
		name        string
		txtRecords  map[string][]string
		dnsError    error
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful verification",
			txtRecords: map[string][]string{
				verificationName: {token},
			},
			expectError: false,
		},
		{
			name: "token with whitespace",
			txtRecords: map[string][]string{
				verificationName: {" " + token + " "},
			},
			expectError: false,
		},
		{
			name: "token not found",
			txtRecords: map[string][]string{
				verificationName: {"wrong-token"},
			},
			expectError: true,
			errorMsg:    "verification token not found",
		},
		{
			name:        "dns lookup fails",
			dnsError:    fmt.Errorf("dns lookup failed"),
			expectError: true,
			errorMsg:    "dns txt lookup failed",
		},
		{
			name:        "no txt records",
			txtRecords:  map[string][]string{},
			expectError: true,
			errorMsg:    "verification token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &mockDNSResolver{
				txtRecords: tt.txtRecords,
				err:        tt.dnsError,
			}

			cfg := DefaultDomainVerificationCheckerConfig()
			cfg.Enabled = true
			cfg.ProviderAddress = providerAddr
			cfg.DNSResolver = resolver

			checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
			require.NoError(t, err)

			record := &keeper.DomainVerificationRecord{
				ProviderAddress: providerAddr,
				Domain:          domain,
				Token:           token,
				Method:          keeper.VerificationMethodDNSTXT,
				Status:          keeper.DomainVerificationPending,
			}

			ctx := context.Background()
			proof, err := checker.verifyDNSTXT(ctx, record)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Contains(t, proof, "dns_txt:")
				assert.Contains(t, proof, verificationName)
			}
		})
	}
}

func TestVerifyCNAME(t *testing.T) {
	providerAddr := "virtengine1abcd1234"
	domain := "example.com"
	token := "abc123def456"
	verificationName := fmt.Sprintf("%s.%s", keeper.DNSVerificationPrefix, domain)
	expectedTarget := fmt.Sprintf("%s.virtengine.network", token)

	tests := []struct {
		name         string
		cnameRecords map[string]string
		dnsError     error
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful verification",
			cnameRecords: map[string]string{
				verificationName: expectedTarget,
			},
			expectError: false,
		},
		{
			name: "cname with trailing dot",
			cnameRecords: map[string]string{
				verificationName: expectedTarget + ".",
			},
			expectError: false,
		},
		{
			name: "wrong cname target",
			cnameRecords: map[string]string{
				verificationName: "wrong-target.example.com",
			},
			expectError: true,
			errorMsg:    "does not match expected",
		},
		{
			name:        "dns lookup fails",
			dnsError:    fmt.Errorf("dns lookup failed"),
			expectError: true,
			errorMsg:    "dns cname lookup failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &mockDNSResolver{
				cnameRecords: tt.cnameRecords,
				err:          tt.dnsError,
			}

			cfg := DefaultDomainVerificationCheckerConfig()
			cfg.Enabled = true
			cfg.ProviderAddress = providerAddr
			cfg.DNSResolver = resolver

			checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
			require.NoError(t, err)

			record := &keeper.DomainVerificationRecord{
				ProviderAddress: providerAddr,
				Domain:          domain,
				Token:           token,
				Method:          keeper.VerificationMethodDNSCNAME,
				Status:          keeper.DomainVerificationPending,
			}

			ctx := context.Background()
			proof, err := checker.verifyCNAME(ctx, record)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Contains(t, proof, "dns_cname:")
				assert.Contains(t, proof, verificationName)
			}
		})
	}
}

func TestVerifyHTTPWellKnown(t *testing.T) {
	providerAddr := "virtengine1abcd1234"
	token := "abc123def456"

	tests := []struct {
		name         string
		responseBody string
		statusCode   int
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "successful verification - plain text",
			responseBody: token,
			statusCode:   http.StatusOK,
			expectError:  false,
		},
		{
			name:         "successful verification - with whitespace",
			responseBody: " " + token + " \n",
			statusCode:   http.StatusOK,
			expectError:  false,
		},
		{
			name:         "successful verification - json",
			responseBody: fmt.Sprintf(`{"token": "%s"}`, token),
			statusCode:   http.StatusOK,
			expectError:  false,
		},
		{
			name:         "wrong token",
			responseBody: "wrong-token",
			statusCode:   http.StatusOK,
			expectError:  true,
			errorMsg:     "verification token not found",
		},
		{
			name:         "wrong json token",
			responseBody: `{"token": "wrong-token"}`,
			statusCode:   http.StatusOK,
			expectError:  true,
			errorMsg:     "verification token not found",
		},
		{
			name:         "http error",
			responseBody: "",
			statusCode:   http.StatusNotFound,
			expectError:  true,
			errorMsg:     "http status 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test HTTP server
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, keeper.HTTPWellKnownPath, r.URL.Path)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Extract domain from server URL
			domain := strings.TrimPrefix(server.URL, "https://")

			cfg := DefaultDomainVerificationCheckerConfig()
			cfg.Enabled = true
			cfg.ProviderAddress = providerAddr
			cfg.HTTPClient = server.Client()

			checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
			require.NoError(t, err)

			record := &keeper.DomainVerificationRecord{
				ProviderAddress: providerAddr,
				Domain:          domain,
				Token:           token,
				Method:          keeper.VerificationMethodHTTPWellKnown,
				Status:          keeper.DomainVerificationPending,
			}

			ctx := context.Background()
			proof, err := checker.verifyHTTPWellKnown(ctx, record)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Contains(t, proof, "http_well_known:")
			}
		})
	}
}

func TestHandleVerificationFailure(t *testing.T) {
	cfg := DefaultDomainVerificationCheckerConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "virtengine1abcd1234"
	cfg.MaxRetries = 3
	cfg.InitialBackoff = 1 * time.Second
	cfg.MaxBackoff = 10 * time.Second

	checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
	require.NoError(t, err)

	record := &keeper.DomainVerificationRecord{
		ProviderAddress: cfg.ProviderAddress,
		Domain:          "example.com",
		Token:           "test-token",
		Method:          keeper.VerificationMethodDNSTXT,
		Status:          keeper.DomainVerificationPending,
	}

	ctx := context.Background()
	testErr := fmt.Errorf("test error")

	// First failure
	checker.handleVerificationFailure(ctx, record, testErr)
	state := checker.retryState[record.Domain]
	require.NotNil(t, state)
	assert.Equal(t, 1, state.Attempts)
	assert.Equal(t, 2*time.Second, state.Backoff)
	assert.True(t, state.NextAttempt.After(time.Now()))

	// Second failure - backoff doubles
	checker.handleVerificationFailure(ctx, record, testErr)
	state = checker.retryState[record.Domain]
	assert.Equal(t, 2, state.Attempts)
	assert.Equal(t, 4*time.Second, state.Backoff)

	// Third failure - backoff doubles again
	checker.handleVerificationFailure(ctx, record, testErr)
	state = checker.retryState[record.Domain]
	assert.Equal(t, 3, state.Attempts)
	assert.Equal(t, 8*time.Second, state.Backoff)

	// Fourth failure - max retries exceeded
	checker.handleVerificationFailure(ctx, record, testErr)
	state = checker.retryState[record.Domain]
	assert.Equal(t, 4, state.Attempts)
}

func TestShouldRetry(t *testing.T) {
	cfg := DefaultDomainVerificationCheckerConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "virtengine1abcd1234"

	checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
	require.NoError(t, err)

	domain := "example.com"

	// Should retry when no retry state exists
	assert.True(t, checker.shouldRetry(domain))

	// Add retry state with future next attempt
	checker.retryState[domain] = &verificationRetryState{
		Attempts:    1,
		NextAttempt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, checker.shouldRetry(domain))

	// Add retry state with past next attempt
	checker.retryState[domain] = &verificationRetryState{
		Attempts:    1,
		NextAttempt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, checker.shouldRetry(domain))
}

func TestBackoffExponentialGrowth(t *testing.T) {
	cfg := DefaultDomainVerificationCheckerConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "virtengine1abcd1234"
	cfg.InitialBackoff = 1 * time.Second
	cfg.MaxBackoff = 100 * time.Second

	checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
	require.NoError(t, err)

	record := &keeper.DomainVerificationRecord{
		ProviderAddress: cfg.ProviderAddress,
		Domain:          "example.com",
		Token:           "test-token",
		Method:          keeper.VerificationMethodDNSTXT,
		Status:          keeper.DomainVerificationPending,
	}

	ctx := context.Background()
	testErr := fmt.Errorf("test error")

	expectedBackoffs := []time.Duration{
		2 * time.Second,   // 1 * 2
		4 * time.Second,   // 2 * 2
		8 * time.Second,   // 4 * 2
		16 * time.Second,  // 8 * 2
		32 * time.Second,  // 16 * 2
		64 * time.Second,  // 32 * 2
		100 * time.Second, // capped at MaxBackoff
		100 * time.Second, // stays at MaxBackoff
	}

	for i, expected := range expectedBackoffs {
		checker.handleVerificationFailure(ctx, record, testErr)
		state := checker.retryState[record.Domain]
		assert.Equal(t, expected, state.Backoff, "iteration %d", i)
	}
}

func TestVerificationTimeout(t *testing.T) {
	cfg := DefaultDomainVerificationCheckerConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "virtengine1abcd1234"
	cfg.VerificationTimeout = 10 * time.Millisecond

	// Create slow DNS resolver that will exceed timeout
	cfg.DNSResolver = &mockDNSResolver{
		txtRecords: map[string][]string{
			"_virtengine-verification.example.com": {"test-token"},
		},
		err: context.DeadlineExceeded,
	}

	checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
	require.NoError(t, err)

	record := &keeper.DomainVerificationRecord{
		ProviderAddress: cfg.ProviderAddress,
		Domain:          "example.com",
		Token:           "test-token",
		Method:          keeper.VerificationMethodDNSTXT,
		Status:          keeper.DomainVerificationPending,
	}

	ctx := context.Background()
	_, err = checker.verifyDNSTXT(ctx, record)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dns txt lookup failed")
}

func TestHTTPWellKnownJSONFormats(t *testing.T) {
	providerAddr := "virtengine1abcd1234"
	token := "abc123def456"

	tests := []struct {
		name         string
		responseBody string
		expectError  bool
	}{
		{
			name:         "simple json",
			responseBody: fmt.Sprintf(`{"token":"%s"}`, token),
			expectError:  false,
		},
		{
			name:         "json with whitespace",
			responseBody: fmt.Sprintf(`{ "token" : "%s" }`, token),
			expectError:  false,
		},
		{
			name:         "json with extra fields",
			responseBody: fmt.Sprintf(`{"token":"%s","other":"value"}`, token),
			expectError:  false,
		},
		{
			name:         "invalid json",
			responseBody: `{"token": invalid}`,
			expectError:  true,
		},
		{
			name:         "empty json",
			responseBody: `{}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			domain := strings.TrimPrefix(server.URL, "https://")

			cfg := DefaultDomainVerificationCheckerConfig()
			cfg.Enabled = true
			cfg.ProviderAddress = providerAddr
			cfg.HTTPClient = server.Client()

			checker, err := NewDomainVerificationChecker(cfg, nil, &mockChainClient{})
			require.NoError(t, err)

			record := &keeper.DomainVerificationRecord{
				ProviderAddress: providerAddr,
				Domain:          domain,
				Token:           token,
				Method:          keeper.VerificationMethodHTTPWellKnown,
				Status:          keeper.DomainVerificationPending,
			}

			ctx := context.Background()
			_, err = checker.verifyHTTPWellKnown(ctx, record)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDomainVerificationRecordJSON(t *testing.T) {
	record := &keeper.DomainVerificationRecord{
		ProviderAddress: "virtengine1abcd1234",
		Domain:          "example.com",
		Token:           "test-token",
		Method:          keeper.VerificationMethodDNSTXT,
		Status:          keeper.DomainVerificationPending,
		GeneratedAt:     time.Now().Unix(),
		ExpiresAt:       time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(record)
	require.NoError(t, err)

	// Unmarshal from JSON
	var decoded keeper.DomainVerificationRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, record.ProviderAddress, decoded.ProviderAddress)
	assert.Equal(t, record.Domain, decoded.Domain)
	assert.Equal(t, record.Token, decoded.Token)
	assert.Equal(t, record.Method, decoded.Method)
	assert.Equal(t, record.Status, decoded.Status)
}
