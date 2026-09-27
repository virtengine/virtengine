package payment

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateGatewayBaseURLRejectsUnsafeDestinations(t *testing.T) {
	t.Parallel()

	unsafe := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"blank", "   "},
		{"file scheme", "file:///etc/passwd"},
		{"gopher scheme", "gopher://127.0.0.1:70/_"},
		{"ftp scheme", "ftp://api-m.paypal.com"},
		{"scheme relative", "//attacker.example/v1"},
		{"plain http to remote host", "http://attacker.example"},
		{"plain http to remote ip", "http://203.0.113.10"},
		{"missing host", "https://"},
		{"embedded credentials", "https://user:secret@attacker.example"},
		{"query string", "https://api-m.paypal.com?redirect=evil"},
		{"fragment", "https://api-m.paypal.com#evil"},
		{"not a url", "://nope"},
	}

	for _, tc := range unsafe {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateGatewayBaseURL(tc.url)
			if err == nil {
				t.Fatalf("ValidateGatewayBaseURL(%q) = %q, want error", tc.url, got)
			}
			if !errors.Is(err, ErrInvalidGatewayBaseURL) {
				t.Fatalf("ValidateGatewayBaseURL(%q) error = %v, want ErrInvalidGatewayBaseURL", tc.url, err)
			}
		})
	}
}

func TestValidateGatewayBaseURLAcceptsSafeDestinations(t *testing.T) {
	t.Parallel()

	safe := []struct {
		name     string
		url      string
		expected string
	}{
		{"paypal live", "https://api-m.paypal.com", "https://api-m.paypal.com"},
		{"paypal sandbox", "https://api-m.sandbox.paypal.com", "https://api-m.sandbox.paypal.com"},
		{"trailing slash trimmed", "https://api-m.paypal.com/", "https://api-m.paypal.com"},
		{"whitespace trimmed", "  https://api-m.paypal.com  ", "https://api-m.paypal.com"},
		{"loopback http", "http://127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"localhost http", "http://localhost:1234/v1", "http://localhost:1234/v1"},
		{"ipv6 loopback http", "http://[::1]:8080", "http://[::1]:8080"},
	}

	for _, tc := range safe {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateGatewayBaseURL(tc.url)
			if err != nil {
				t.Fatalf("ValidateGatewayBaseURL(%q) returned unexpected error: %v", tc.url, err)
			}
			if got != tc.expected {
				t.Fatalf("ValidateGatewayBaseURL(%q) = %q, want %q", tc.url, got, tc.expected)
			}
		})
	}
}

// TestNewPayPalAdapterRejectsHostileBaseURL proves the constructor refuses to
// build an adapter that would send the OAuth credentials to an attacker-chosen
// destination (gosec G704 SSRF).
func TestNewPayPalAdapterRejectsHostileBaseURL(t *testing.T) {
	t.Parallel()

	hostile := []string{
		"file:///etc/passwd",
		"http://attacker.example",
		"https://user:secret@attacker.example",
		"gopher://127.0.0.1:70/_",
	}

	for _, base := range hostile {
		t.Run(base, func(t *testing.T) {
			t.Parallel()

			adapter, err := NewPayPalAdapter(PayPalConfig{
				ClientID:     "client",
				ClientSecret: "secret",
				BaseURL:      base,
			})
			if err == nil {
				t.Fatalf("NewPayPalAdapter accepted hostile base URL %q", base)
			}
			if adapter != nil {
				t.Fatalf("NewPayPalAdapter returned a non-nil adapter for hostile base URL %q", base)
			}
			if !errors.Is(err, ErrInvalidGatewayBaseURL) {
				t.Fatalf("NewPayPalAdapter(%q) error = %v, want ErrInvalidGatewayBaseURL", base, err)
			}
		})
	}
}

// TestNewPayPalAdapterAcceptsLoopbackTestServer keeps local test servers working
// while still validating the destination.
func TestNewPayPalAdapterAcceptsLoopbackTestServer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter, err := NewPayPalAdapter(PayPalConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		BaseURL:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewPayPalAdapter rejected loopback test server: %v", err)
	}

	pp, ok := adapter.(*PayPalAdapter)
	if !ok {
		t.Fatalf("NewPayPalAdapter returned %T, want *PayPalAdapter", adapter)
	}
	if pp.baseURL != server.URL {
		t.Fatalf("adapter baseURL = %q, want %q", pp.baseURL, server.URL)
	}
}
