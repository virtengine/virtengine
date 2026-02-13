package client

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

func makeHTTPDialer(remoteAddr string) (func(context.Context, string, string) (net.Conn, error), error) {
	u, err := newParsedURL(remoteAddr)
	if err != nil {
		return nil, err
	}

	protocol := u.Scheme

	// accept http(s) as an alias for tcp
	switch protocol {
	case protoHTTP, protoHTTPS:
		protocol = protoTCP
	}

	dialFn := func(ctx context.Context, proto, addr string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   10 * time.Second, // Connection timeout
			KeepAlive: 30 * time.Second, // Keep-alive period
		}).DialContext(ctx, protocol, u.GetDialAddress())
	}

	return dialFn, nil
}

// NewHTTPClient is used to create an http client with some default parameters.
// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://).
// An error will be returned in case of invalid remoteAddr.
func NewHTTPClient(ctx context.Context, remoteAddr string) (*http.Client, error) {
	dialFn, err := makeHTTPDialer(remoteAddr)
	if err != nil {
		return nil, err
	}

	config := security.DefaultHTTPClientConfig()
	client := &http.Client{
		Transport: &http.Transport{
			// Connection pooling settings
			MaxIdleConns:          config.MaxIdleConns,        // Maximum number of idle connections across all hosts
			MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost, // Maximum number of idle connections per host
			MaxConnsPerHost:       config.MaxConnsPerHost,     // Maximum number of connections per host
			IdleConnTimeout:       config.IdleConnTimeout,     // How long idle connections are kept alive
			TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
			ResponseHeaderTimeout: config.ResponseHeaderTimeout,
			ExpectContinueTimeout: config.ExpectContinueTimeout,

			// Enable connection reuse
			DisableKeepAlives: config.DisableKeepAlives,

			// Set to true to prevent GZIP-bomb DoS attacks
			DisableCompression: true,
			DialContext:        dialFn,
			TLSClientConfig:    security.SecureTLSConfig(),

			// Force HTTP/1.1 to ensure better connection pooling behavior
			// Some RPC nodes may not handle HTTP/2 connection pooling optimally
			ForceAttemptHTTP2: false,
		},
		Timeout: config.Timeout,
	}

	return client, nil
}
