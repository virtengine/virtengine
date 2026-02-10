package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/virtengine/virtengine/pkg/security"
)

// HTTPTracingHandler wraps an HTTP handler with OpenTelemetry instrumentation.
func HTTPTracingHandler(handler http.Handler, name string) http.Handler {
	if handler == nil {
		return nil
	}
	if name == "" {
		name = "http.server"
	}
	return otelhttp.NewHandler(handler, name)
}

// TracedHTTPClient returns an HTTP client instrumented with OpenTelemetry.
func TracedHTTPClient() *http.Client {
	client := security.NewSecureHTTPClient()
	client.Transport = otelhttp.NewTransport(client.Transport)
	return client
}
