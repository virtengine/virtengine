// Package observability provides structured logging, metrics, and tracing
// for the VirtEngine platform.
//
// MONITOR-001: Module-specific tracers and tracing configuration
//
// This file provides configuration and module-specific tracers.
// The core tracing interfaces (Tracer, Span, SpanContext) are defined
// in observability.go. The no-op implementations are in logger.go.
//
// For full OTLP tracing integration, see deploy/monitoring/otel/otel-collector.yaml
package observability

import (
	"context"
	"sync"
	"time"
)

// TracingConfig holds tracing configuration
type TracingConfig struct {
	// ServiceName is the name of the service
	ServiceName string

	// ServiceVersion is the version of the service
	ServiceVersion string

	// Environment (development, staging, production)
	Environment string

	// Enabled enables distributed tracing
	Enabled bool

	// Endpoint is the OTLP endpoint for traces (e.g., "localhost:4317")
	Endpoint string

	// Insecure disables TLS for the exporter connection
	Insecure bool

	// SampleRate is the sampling rate (0.0 to 1.0)
	SampleRate float64

	// BatchTimeout is the maximum time before a batch is exported
	BatchTimeout time.Duration

	// MaxExportBatchSize is the maximum number of spans in a batch
	MaxExportBatchSize int

	// MaxQueueSize is the maximum number of spans queued for export
	MaxQueueSize int
}

// DefaultTracingConfig returns sensible defaults for tracing
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		ServiceName:        "virtengine",
		ServiceVersion:     "unknown",
		Environment:        "development",
		Enabled:            false,
		Endpoint:           "localhost:4317",
		Insecure:           true,
		SampleRate:         0.1,
		BatchTimeout:       5 * time.Second,
		MaxExportBatchSize: 512,
		MaxQueueSize:       2048,
	}
}

// ============================================================================
// TracerProvider
// ============================================================================

// TracerProvider provides tracers
type TracerProvider struct {
	config  TracingConfig
	tracers map[string]Tracer
	mu      sync.RWMutex
}

var (
	globalTracerProvider *TracerProvider
	tracerProviderOnce   sync.Once
)

// InitTracing initializes the global tracer provider
func InitTracing(cfg TracingConfig) (*TracerProvider, error) {
	tracerProviderOnce.Do(func() {
		globalTracerProvider = &TracerProvider{
			config:  cfg,
			tracers: make(map[string]Tracer),
		}
	})
	return globalTracerProvider, nil
}

// GetTracerProvider returns the global tracer provider
func GetTracerProvider() *TracerProvider {
	if globalTracerProvider == nil {
		// Return default if not initialized
		globalTracerProvider = &TracerProvider{
			config:  DefaultTracingConfig(),
			tracers: make(map[string]Tracer),
		}
	}
	return globalTracerProvider
}

// Tracer returns a tracer for the given name
func (tp *TracerProvider) Tracer(name string) Tracer {
	tp.mu.RLock()
	if t, ok := tp.tracers[name]; ok {
		tp.mu.RUnlock()
		return t
	}
	tp.mu.RUnlock()

	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Double-check after acquiring write lock
	if t, ok := tp.tracers[name]; ok {
		return t
	}

	t := &noopTracer{}
	tp.tracers[name] = t
	return t
}

// Shutdown shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	return nil
}

// ============================================================================
// Span Helpers
// ============================================================================

// SpanAttributes contains common span attributes
type SpanAttributes struct {
	// User attributes
	UserID    string
	AccountID string

	// Request attributes
	RequestID   string
	TraceParent string

	// Module attributes
	Module    string
	Operation string
	Component string

	// Error attributes
	ErrorCode    string
	ErrorMessage string

	// Custom attributes
	Custom map[string]interface{}
}

// ToFields converts SpanAttributes to Field slice
func (a SpanAttributes) ToFields() []Field {
	var fields []Field

	if a.UserID != "" {
		fields = append(fields, String("user.id", a.UserID))
	}
	if a.AccountID != "" {
		fields = append(fields, String("account.id", a.AccountID))
	}
	if a.RequestID != "" {
		fields = append(fields, String("request.id", a.RequestID))
	}
	if a.Module != "" {
		fields = append(fields, String("virtengine.module", a.Module))
	}
	if a.Operation != "" {
		fields = append(fields, String("virtengine.operation", a.Operation))
	}
	if a.Component != "" {
		fields = append(fields, String("virtengine.component", a.Component))
	}
	if a.ErrorCode != "" {
		fields = append(fields, String("error.code", a.ErrorCode))
	}
	if a.ErrorMessage != "" {
		fields = append(fields, String("error.message", a.ErrorMessage))
	}

	for k, v := range a.Custom {
		switch val := v.(type) {
		case string:
			fields = append(fields, String(k, val))
		case int:
			fields = append(fields, Int(k, val))
		case int64:
			fields = append(fields, Int64(k, val))
		case float64:
			fields = append(fields, Float64(k, val))
		case bool:
			fields = append(fields, Bool(k, val))
		}
	}

	return fields
}

// StartSpan starts a new span with common attributes
func StartSpan(ctx context.Context, name string, attrs SpanAttributes) (context.Context, Span) {
	tracer := GetTracerProvider().Tracer("virtengine")
	return tracer.Start(ctx, name, WithAttributes(attrs.ToFields()...))
}

// StartSpanWithKind starts a new span with a specific kind
func StartSpanWithKind(ctx context.Context, name string, kind SpanKind, attrs SpanAttributes) (context.Context, Span) {
	tracer := GetTracerProvider().Tracer("virtengine")
	return tracer.Start(ctx, name, WithSpanKind(kind), WithAttributes(attrs.ToFields()...))
}

// ============================================================================
// Module-Specific Tracers
// ============================================================================

// VEIDTracer provides tracing for VEID operations
type VEIDTracer struct {
	tracer Tracer
}

// NewVEIDTracer creates a new VEID tracer
func NewVEIDTracer() *VEIDTracer {
	return &VEIDTracer{
		tracer: GetTracerProvider().Tracer("virtengine/veid"),
	}
}

// StartVerification starts a verification span
func (t *VEIDTracer) StartVerification(ctx context.Context, requestID, accountAddress string) (context.Context, Span) {
	return t.tracer.Start(ctx, "veid.verification",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("veid.request_id", requestID),
			String("veid.account_address", accountAddress),
			String("virtengine.module", "veid"),
			String("virtengine.operation", "verification"),
		),
	)
}

// StartInference starts an ML inference span
func (t *VEIDTracer) StartInference(ctx context.Context, modelVersion string) (context.Context, Span) {
	return t.tracer.Start(ctx, "veid.inference",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("veid.model_version", modelVersion),
			String("virtengine.module", "veid"),
			String("virtengine.operation", "inference"),
		),
	)
}

// StartScopeUpload starts a scope upload span
func (t *VEIDTracer) StartScopeUpload(ctx context.Context, scopeType string) (context.Context, Span) {
	return t.tracer.Start(ctx, "veid.scope_upload",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("veid.scope_type", scopeType),
			String("virtengine.module", "veid"),
			String("virtengine.operation", "scope_upload"),
		),
	)
}

// MarketTracer provides tracing for marketplace operations
type MarketTracer struct {
	tracer Tracer
}

// NewMarketTracer creates a new market tracer
func NewMarketTracer() *MarketTracer {
	return &MarketTracer{
		tracer: GetTracerProvider().Tracer("virtengine/market"),
	}
}

// StartOrderCreate starts an order creation span
func (t *MarketTracer) StartOrderCreate(ctx context.Context, orderID string) (context.Context, Span) {
	return t.tracer.Start(ctx, "market.order_create",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("market.order_id", orderID),
			String("virtengine.module", "market"),
			String("virtengine.operation", "order_create"),
		),
	)
}

// StartBidPlace starts a bid placement span
func (t *MarketTracer) StartBidPlace(ctx context.Context, orderID, providerID string) (context.Context, Span) {
	return t.tracer.Start(ctx, "market.bid_place",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("market.order_id", orderID),
			String("market.provider_id", providerID),
			String("virtengine.module", "market"),
			String("virtengine.operation", "bid_place"),
		),
	)
}

// StartLeaseCreate starts a lease creation span
func (t *MarketTracer) StartLeaseCreate(ctx context.Context, bidID string) (context.Context, Span) {
	return t.tracer.Start(ctx, "market.lease_create",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("market.bid_id", bidID),
			String("virtengine.module", "market"),
			String("virtengine.operation", "lease_create"),
		),
	)
}

// ProviderTracer provides tracing for provider daemon operations
type ProviderTracer struct {
	tracer Tracer
}

// NewProviderTracer creates a new provider tracer
func NewProviderTracer() *ProviderTracer {
	return &ProviderTracer{
		tracer: GetTracerProvider().Tracer("virtengine/provider"),
	}
}

// StartDeployment starts a deployment span
func (t *ProviderTracer) StartDeployment(ctx context.Context, leaseID, adapter string) (context.Context, Span) {
	return t.tracer.Start(ctx, "provider.deployment",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("provider.lease_id", leaseID),
			String("provider.adapter", adapter),
			String("virtengine.module", "provider"),
			String("virtengine.operation", "deployment"),
		),
	)
}

// StartBidEngine starts a bid engine span
func (t *ProviderTracer) StartBidEngine(ctx context.Context, orderID string) (context.Context, Span) {
	return t.tracer.Start(ctx, "provider.bid_engine",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("provider.order_id", orderID),
			String("virtengine.module", "provider"),
			String("virtengine.operation", "bid_engine"),
		),
	)
}

// StartUsageMeter starts a usage metering span
func (t *ProviderTracer) StartUsageMeter(ctx context.Context, deploymentID string) (context.Context, Span) {
	return t.tracer.Start(ctx, "provider.usage_meter",
		WithSpanKind(SpanKindInternal),
		WithAttributes(
			String("provider.deployment_id", deploymentID),
			String("virtengine.module", "provider"),
			String("virtengine.operation", "usage_meter"),
		),
	)
}

// ============================================================================
// Context Propagation
// ============================================================================

// InjectTraceContext injects trace context into a map
func InjectTraceContext(ctx context.Context, headers map[string]string) {
	tracer := GetTracerProvider().Tracer("virtengine")
	carrier := MapCarrier(headers)
	_ = tracer.Inject(ctx, carrier)
}

// ExtractTraceContext extracts trace context from a map
func ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	tracer := GetTracerProvider().Tracer("virtengine")
	carrier := MapCarrier(headers)
	return tracer.Extract(ctx, carrier)
}

// TraceIDFromContext returns the trace ID from context
// This is a placeholder - full implementation requires real tracing provider
func TraceIDFromContext(ctx context.Context) string {
	return ""
}

// SpanIDFromContext returns the span ID from context
// This is a placeholder - full implementation requires real tracing provider
func SpanIDFromContext(ctx context.Context) string {
	return ""
}
