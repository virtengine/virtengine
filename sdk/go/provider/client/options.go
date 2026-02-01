package rest

import (
	"crypto/tls"
	"fmt"
	"log"

	ajwt "github.com/virtengine/virtengine/sdk/go/util/jwt"
	atls "github.com/virtengine/virtengine/sdk/go/util/tls"
)

type clientOptions struct {
	certs       []tls.Certificate
	signer      ajwt.SignerI
	token       string
	providerURL string
	certQuerier atls.CertificateQuerier
	// insecureSkipVerify disables TLS certificate verification.
	// This should only be used in development/testing environments.
	insecureSkipVerify bool
}

// ClientOption is a function type that modifies a clientOptions struct and returns an error.
// It follows the functional options pattern for configuring the REST client.
//
// ClientOption functions can be passed to client constructors to customize the client configuration.
// Common options include:
// - WithCerts: Configure TLS certificates for secure communication
// - WithJWTSigner: Set a JWT signer for authentication
// - WithToken: Provide an authentication token
//
// If an error occurs while applying the option, it will be returned to the caller.
type ClientOption func(options *clientOptions) error

// WithAuthCerts configures the client with the provided TLS certificates for secure communication.
func WithAuthCerts(certs []tls.Certificate) ClientOption {
	return func(options *clientOptions) error {
		options.certs = certs

		return nil
	}
}

// WithAuthJWTSigner sets the JWT signer for authentication in the client.
func WithAuthJWTSigner(val ajwt.SignerI) ClientOption {
	return func(options *clientOptions) error {
		options.signer = val
		return nil
	}
}

// WithAuthToken provides an authentication token for the client.
func WithAuthToken(val string) ClientOption {
	return func(options *clientOptions) error {
		options.token = val
		return nil
	}
}

// WithProviderURL configures the client to use the specified provider URL directly.
// This option is mutually exclusive with WithQueryClient.
func WithProviderURL(providerURL string) ClientOption {
	return func(options *clientOptions) error {
		options.providerURL = providerURL
		return nil
	}
}

// WithCertQuerier configures the client to use the specified certificate querier for certificate validation.
func WithCertQuerier(certQuerier atls.CertificateQuerier) ClientOption {
	return func(options *clientOptions) error {
		if certQuerier == nil {
			return fmt.Errorf("certQuerier must not be nil")
		}

		options.certQuerier = certQuerier
		return nil
	}
}

// WithInsecureSkipVerify configures the client to skip TLS certificate verification.
// WARNING: This option disables certificate verification, making the connection
// vulnerable to man-in-the-middle attacks. USE WITH CAUTION.
// This should only be used in development/testing environments where the provider
// uses self-signed certificates that cannot be verified through standard PKI or on-chain validation.
func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(options *clientOptions) error {
		if skip {
			log.Println("WARNING: TLS certificate verification disabled - NOT RECOMMENDED FOR PRODUCTION. " +
				"This makes the connection vulnerable to man-in-the-middle attacks.")
		}
		options.insecureSkipVerify = skip
		return nil
	}
}

