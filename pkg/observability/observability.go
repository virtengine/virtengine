// Package observability provides structured logging, metrics, and tracing
// for the VirtEngine platform.
//
// VE-708: Observability Infrastructure
//
// CRITICAL: Never log sensitive data (private keys, passwords, tokens, encrypted payloads)
package observability

import (
	"context"
	"time"
)

// Level represents log severity level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Config holds observability configuration
type Config struct {
	// ServiceName is the name of the service
	ServiceName string

	// ServiceVersion is the version of the service
	ServiceVersion string

	// Environment (development, staging, production)
	Environment string

	// LogLevel is the minimum log level
	LogLevel Level

	// LogFormat is the log output format (json, text)
	LogFormat string

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool

	// MetricsPort is the port for metrics endpoint
	MetricsPort int

	// TracingEnabled enables distributed tracing
	TracingEnabled bool

	// TracingEndpoint is the OTLP endpoint for traces
	TracingEndpoint string

	// TracingSampleRate is the sampling rate (0.0 to 1.0)
	TracingSampleRate float64

	// SensitiveFields are field names that should be redacted
	SensitiveFields []string
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		ServiceName:       "virtengine",
		ServiceVersion:    "unknown",
		Environment:       "development",
		LogLevel:          LevelInfo,
		LogFormat:         "json",
		MetricsEnabled:    true,
		MetricsPort:       9090,
		TracingEnabled:    false,
		TracingSampleRate: 0.1,
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"mnemonic",
			"private",
			"credential",
			"auth",
			"bearer",
			"signature",
			"encrypted",
		},
	}
}

// Observer is the main observability interface
type Observer interface {
	// Logger returns the structured logger
	Logger() Logger

	// Metrics returns the metrics registry
	Metrics() Metrics

	// Tracer returns the tracer
	Tracer() Tracer

	// Shutdown gracefully shuts down the observer
	Shutdown(ctx context.Context) error
}

// Logger is the structured logging interface
type Logger interface {
	// Debug logs at debug level
	Debug(msg string, fields ...Field)

	// Info logs at info level
	Info(msg string, fields ...Field)

	// Warn logs at warn level
	Warn(msg string, fields ...Field)

	// Error logs at error level
	Error(msg string, fields ...Field)

	// Fatal logs at fatal level and exits
	Fatal(msg string, fields ...Field)

	// With returns a logger with additional fields
	With(fields ...Field) Logger

	// WithContext returns a logger with trace context
	WithContext(ctx context.Context) Logger
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Metrics is the metrics interface
type Metrics interface {
	// Counter returns a counter metric
	Counter(name, help string, labels ...string) Counter

	// Gauge returns a gauge metric
	Gauge(name, help string, labels ...string) Gauge

	// Histogram returns a histogram metric
	Histogram(name, help string, buckets []float64, labels ...string) Histogram

	// Summary returns a summary metric
	Summary(name, help string, objectives map[float64]float64, labels ...string) Summary
}

// Counter is a monotonically increasing counter
type Counter interface {
	// Inc increments the counter by 1
	Inc(labels ...string)

	// Add adds the given value to the counter
	Add(value float64, labels ...string)
}

// Gauge is a metric that can go up and down
type Gauge interface {
	// Set sets the gauge to the given value
	Set(value float64, labels ...string)

	// Inc increments the gauge by 1
	Inc(labels ...string)

	// Dec decrements the gauge by 1
	Dec(labels ...string)

	// Add adds the given value to the gauge
	Add(value float64, labels ...string)

	// Sub subtracts the given value from the gauge
	Sub(value float64, labels ...string)
}

// Histogram is a metric that samples observations
type Histogram interface {
	// Observe records a value
	Observe(value float64, labels ...string)
}

// Summary is a metric that samples observations with quantiles
type Summary interface {
	// Observe records a value
	Observe(value float64, labels ...string)
}

// Tracer is the distributed tracing interface
type Tracer interface {
	// Start starts a new span
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// Extract extracts span context from carrier
	Extract(ctx context.Context, carrier Carrier) context.Context

	// Inject injects span context into carrier
	Inject(ctx context.Context, carrier Carrier) error
}

// Span represents a trace span
type Span interface {
	// End ends the span
	End()

	// SetStatus sets the span status
	SetStatus(code StatusCode, description string)

	// SetAttribute sets a span attribute
	SetAttribute(key string, value interface{})

	// RecordError records an error
	RecordError(err error)

	// AddEvent adds an event to the span
	AddEvent(name string, attrs ...Field)

	// SpanContext returns the span context
	SpanContext() SpanContext
}

// SpanContext contains span identifying information
type SpanContext struct {
	TraceID    string
	SpanID     string
	TraceFlags byte
}

// StatusCode represents span status
type StatusCode int

const (
	StatusUnset StatusCode = iota
	StatusOK
	StatusError
)

// SpanOption configures a span
type SpanOption func(*spanConfig)

type spanConfig struct {
	kind       SpanKind
	attributes []Field
}

// SpanKind represents the type of span
type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption {
	return func(cfg *spanConfig) {
		cfg.kind = kind
	}
}

// WithAttributes sets initial span attributes
func WithAttributes(attrs ...Field) SpanOption {
	return func(cfg *spanConfig) {
		cfg.attributes = append(cfg.attributes, attrs...)
	}
}

// Carrier is used to propagate trace context
type Carrier interface {
	Get(key string) string
	Set(key, value string)
}

// MapCarrier is a simple map-based carrier
type MapCarrier map[string]string

func (c MapCarrier) Get(key string) string {
	return c[key]
}

func (c MapCarrier) Set(key, value string) {
	c[key] = value
}
