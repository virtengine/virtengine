package payment

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ErrInvalidGatewayBaseURL is returned when an operator-supplied payment gateway
// base URL is not a safe destination for outbound API calls.
var ErrInvalidGatewayBaseURL = errors.New("invalid payment gateway base URL")

// ValidateGatewayBaseURL validates and normalises a payment gateway base URL.
//
// The base URL decides where payment API calls — and the OAuth credentials
// attached to them — are sent, so a tampered configuration value is a
// server-side request forgery primitive and a credential-exfiltration path
// (gosec G704). Only absolute https URLs are accepted. Plain http is allowed
// solely for loopback hosts so local test servers keep working, and embedded
// credentials, query strings and fragments are rejected outright.
func ValidateGatewayBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%w: url is empty", ErrInvalidGatewayBaseURL)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidGatewayBaseURL, err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "https":
		// Always acceptable.
	case "http":
		if !isLoopbackHost(parsed.Hostname()) {
			return "", fmt.Errorf("%w: plain http is only allowed for loopback hosts, got %q",
				ErrInvalidGatewayBaseURL, parsed.Hostname())
		}
	default:
		return "", fmt.Errorf("%w: unsupported scheme %q", ErrInvalidGatewayBaseURL, parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("%w: missing host", ErrInvalidGatewayBaseURL)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("%w: credentials must not be embedded in the URL", ErrInvalidGatewayBaseURL)
	}
	if parsed.RawQuery != "" {
		return "", fmt.Errorf("%w: query strings are not allowed", ErrInvalidGatewayBaseURL)
	}
	if parsed.Fragment != "" {
		return "", fmt.Errorf("%w: fragments are not allowed", ErrInvalidGatewayBaseURL)
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

// isLoopbackHost reports whether host names the local machine (or is empty,
// which callers treat as a missing host separately).
func isLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
