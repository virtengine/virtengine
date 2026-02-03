package client

import (
	"context"
	"net"
	"net/http"
	"time"
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

	client := &http.Client{
		Transport: &http.Transport{
			// Connection pooling settings
			MaxIdleConns:          100,              // Maximum number of idle connections across all hosts
			MaxIdleConnsPerHost:   10,               // Maximum number of idle connections per host
			MaxConnsPerHost:       50,               // Maximum number of connections per host
			IdleConnTimeout:       90 * time.Second, // How long idle connections are kept alive
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			// Enable connection reuse
			DisableKeepAlives: false,

			// Set to true to prevent GZIP-bomb DoS attacks
			DisableCompression: true,
			DialContext:        dialFn,

			// Force HTTP/1.1 to ensure better connection pooling behavior
			// Some RPC nodes may not handle HTTP/2 connection pooling optimally
			ForceAttemptHTTP2: false,
		},
	}

	return client, nil
}
