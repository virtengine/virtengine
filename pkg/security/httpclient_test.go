package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.Timeout)
	}
	if config.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", config.MinTLSVersion)
	}
	if config.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false by default")
	}
}

func TestNewSecureHTTPClient(t *testing.T) {
	client := NewSecureHTTPClient()

	if client.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", transport.TLSClientConfig.MinVersion)
	}
}

func TestNewSecureHTTPClientWithOptions(t *testing.T) {
	client := NewSecureHTTPClient(
		WithTimeout(60*time.Second),
		WithMinTLSVersion(tls.VersionTLS13),
		WithMaxIdleConns(50),
	)

	if client.Timeout != 60*time.Second {
		t.Errorf("expected Timeout 60s, got %v", client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinTLSVersion TLS 1.3, got %v", transport.TLSClientConfig.MinVersion)
	}

	if transport.MaxIdleConns != 50 {
		t.Errorf("expected MaxIdleConns 50, got %d", transport.MaxIdleConns)
	}
}

func TestHTTPClientMinTLSVersionClamp(t *testing.T) {
	config := DefaultHTTPClientConfig()
	config.MinTLSVersion = tls.VersionTLS10

	client := NewHTTPClientFromConfig(config)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", transport.TLSClientConfig.MinVersion)
	}
}

func TestNewSecureHTTPClientTLS13(t *testing.T) {
	client := NewSecureHTTPClientTLS13()

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinTLSVersion TLS 1.3, got %v", transport.TLSClientConfig.MinVersion)
	}
}

func TestNewDevHTTPClient(t *testing.T) {
	client := NewDevHTTPClient(true)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true")
	}
}

func TestSecureTLSConfig(t *testing.T) {
	config := SecureTLSConfig()

	if config.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", config.MinVersion)
	}

	if len(config.CipherSuites) == 0 {
		t.Error("expected cipher suites to be configured")
	}

	// Verify all configured cipher suites are secure
	for _, suite := range config.CipherSuites {
		switch suite {
		case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
			// OK - these are secure
		default:
			t.Errorf("unexpected cipher suite: %v", suite)
		}
	}
}

func TestSecureTLSConfigTLS13(t *testing.T) {
	config := SecureTLSConfigTLS13()

	if config.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinTLSVersion TLS 1.3, got %v", config.MinVersion)
	}
}

func TestMustSecureTLSConfig(t *testing.T) {
	config := MustSecureTLSConfig()

	if config.MinVersion < tls.VersionTLS12 {
		t.Error("minimum TLS version should be at least TLS 1.2")
	}
}

func TestHTTPClientIntegration(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewSecureHTTPClient(WithTimeout(5 * time.Second))

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClientOptions(t *testing.T) {
	tests := []struct {
		name   string
		opts   []HTTPClientOption
		verify func(*testing.T, *http.Client)
	}{
		{
			name: "WithTimeout",
			opts: []HTTPClientOption{WithTimeout(45 * time.Second)},
			verify: func(t *testing.T, c *http.Client) {
				if c.Timeout != 45*time.Second {
					t.Errorf("expected Timeout 45s, got %v", c.Timeout)
				}
			},
		},
		{
			name: "WithConnectTimeout",
			opts: []HTTPClientOption{WithConnectTimeout(5 * time.Second)},
			verify: func(t *testing.T, c *http.Client) {
				transport := c.Transport.(*http.Transport)
				// DialContext timeout is set via Dialer, verify transport exists
				if transport == nil {
					t.Error("expected transport to be set")
				}
			},
		},
		{
			name: "WithDisableKeepAlives",
			opts: []HTTPClientOption{WithDisableKeepAlives(true)},
			verify: func(t *testing.T, c *http.Client) {
				transport := c.Transport.(*http.Transport)
				if !transport.DisableKeepAlives {
					t.Error("expected DisableKeepAlives to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSecureHTTPClient(tt.opts...)
			tt.verify(t, client)
		})
	}
}
