// Package security provides security utilities for HTTP client configuration.
package security

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// HTTPClientConfig contains configuration options for secure HTTP clients.
type HTTPClientConfig struct {
	// Timeout is the total timeout for a request including connection, headers, and body.
	// Default: 30 seconds
	Timeout time.Duration

	// ConnectTimeout is the maximum time to wait for a connection to be established.
	// Default: 10 seconds
	ConnectTimeout time.Duration

	// TLSHandshakeTimeout is the maximum time to wait for TLS handshake.
	// Default: 10 seconds
	TLSHandshakeTimeout time.Duration

	// ResponseHeaderTimeout is the time to wait for response headers.
	// Default: 10 seconds
	ResponseHeaderTimeout time.Duration

	// IdleConnTimeout is how long idle connections are kept in the pool.
	// Default: 90 seconds
	IdleConnTimeout time.Duration

	// MaxIdleConns is the maximum number of idle connections across all hosts.
	// Default: 100
	MaxIdleConns int

	// MaxIdleConnsPerHost is the maximum number of idle connections per host.
	// Default: 10
	MaxIdleConnsPerHost int

	// MaxConnsPerHost is the maximum number of connections per host.
	// Default: 100
	MaxConnsPerHost int

	// MinTLSVersion is the minimum TLS version to use.
	// Default: tls.VersionTLS12
	MinTLSVersion uint16

	// InsecureSkipVerify disables TLS certificate verification.
	// DANGER: Only use for local development or testing. Never in production.
	InsecureSkipVerify bool

	// DisableKeepAlives disables HTTP keep-alives and only uses connections once.
	// Default: false
	DisableKeepAlives bool

	// ExpectContinueTimeout is the time to wait for server's 100-continue response.
	// Default: 1 second
	ExpectContinueTimeout time.Duration
}

// HTTPClientOption is a functional option for configuring HTTPClientConfig.
type HTTPClientOption func(*HTTPClientConfig)

// WithTimeout sets the total request timeout.
func WithTimeout(d time.Duration) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.Timeout = d
	}
}

// WithConnectTimeout sets the connection establishment timeout.
func WithConnectTimeout(d time.Duration) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.ConnectTimeout = d
	}
}

// WithTLSHandshakeTimeout sets the TLS handshake timeout.
func WithTLSHandshakeTimeout(d time.Duration) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.TLSHandshakeTimeout = d
	}
}

// WithMinTLSVersion sets the minimum TLS version.
func WithMinTLSVersion(version uint16) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.MinTLSVersion = version
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// DANGER: Only use for local development or testing. Never in production.
func WithInsecureSkipVerify(skip bool) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.InsecureSkipVerify = skip
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.MaxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost sets the maximum number of idle connections per host.
func WithMaxIdleConnsPerHost(n int) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.MaxIdleConnsPerHost = n
	}
}

// WithDisableKeepAlives disables HTTP keep-alives.
func WithDisableKeepAlives(disable bool) HTTPClientOption {
	return func(c *HTTPClientConfig) {
		c.DisableKeepAlives = disable
	}
}

// DefaultHTTPClientConfig returns the default secure HTTP client configuration.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:               30 * time.Second,
		ConnectTimeout:        10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       100,
		MinTLSVersion:         tls.VersionTLS12,
		InsecureSkipVerify:    false,
		DisableKeepAlives:     false,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewSecureHTTPClient creates a new HTTP client with secure defaults.
// This client enforces TLS 1.2 minimum, proper timeouts, and certificate verification.
//
// Use this instead of http.DefaultClient or &http.Client{} to ensure security.
//
// Example:
//
//	client := security.NewSecureHTTPClient()
//	resp, err := client.Get("https://api.example.com/data")
func NewSecureHTTPClient(opts ...HTTPClientOption) *http.Client {
	config := DefaultHTTPClientConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return NewHTTPClientFromConfig(config)
}

// NewHTTPClientFromConfig creates an HTTP client from a configuration struct.
func NewHTTPClientFromConfig(config HTTPClientConfig) *http.Client {
	// Create TLS config with secure defaults
	tlsConfig := &tls.Config{
		MinVersion:         config.MinTLSVersion,
		InsecureSkipVerify: config.InsecureSkipVerify, //nolint:gosec // G402: Configurable for dev/test
	}

	// Create transport with timeouts
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.ConnectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       tlsConfig,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		IdleConnTimeout:       config.IdleConnTimeout,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		DisableKeepAlives:     config.DisableKeepAlives,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// NewSecureHTTPClientTLS13 creates an HTTP client that requires TLS 1.3 minimum.
// Use this for connections to services that support TLS 1.3.
func NewSecureHTTPClientTLS13(opts ...HTTPClientOption) *http.Client {
	allOpts := append([]HTTPClientOption{WithMinTLSVersion(tls.VersionTLS13)}, opts...)
	return NewSecureHTTPClient(allOpts...)
}

// NewDevHTTPClient creates an HTTP client for development/testing that may skip TLS verification.
// DANGER: This should NEVER be used in production code.
//
// Example:
//
//	client := security.NewDevHTTPClient(true) // Skip TLS verification
func NewDevHTTPClient(skipTLSVerify bool) *http.Client {
	return NewSecureHTTPClient(WithInsecureSkipVerify(skipTLSVerify))
}

// SecureTLSConfig returns a TLS configuration with secure defaults.
// Use this when creating custom transports or TLS connections.
//
// Defaults:
//   - MinVersion: TLS 1.2
//   - Secure cipher suites (when using TLS 1.2)
//   - Certificate verification enabled
func SecureTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		// Prefer secure cipher suites for TLS 1.2
		// TLS 1.3 cipher suites are automatically selected by Go
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		// Prefer server cipher suites
		PreferServerCipherSuites: true,
	}
}

// SecureTLSConfigTLS13 returns a TLS configuration requiring TLS 1.3 minimum.
func SecureTLSConfigTLS13() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		// TLS 1.3 cipher suites are automatically selected by Go
		// No need to specify CipherSuites for TLS 1.3 only
	}
}

// MustSecureTLSConfig is like SecureTLSConfig but also validates the config.
// Use this when you want to ensure the configuration is valid at initialization.
func MustSecureTLSConfig() *tls.Config {
	config := SecureTLSConfig()
	// Validate minimum version is at least TLS 1.2
	if config.MinVersion < tls.VersionTLS12 {
		panic("TLS configuration has insecure minimum version")
	}
	return config
}
