package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultTraceEndpoint = "localhost:4317"

type otelTracer struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	provider   *sdktrace.TracerProvider
}

type otelSpanWrapper struct {
	span trace.Span
}

func newOTelTracer(cfg Config) (*otelTracer, func(context.Context) error, error) {
	endpoint := strings.TrimSpace(cfg.TracingEndpoint)
	if endpoint == "" {
		endpoint = defaultTraceEndpoint
	}

	clientOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if cfg.TracingEndpoint == "" || !strings.HasPrefix(strings.ToLower(cfg.TracingEndpoint), "https://") {
		clientOpts = append(clientOpts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(clientOpts...),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("init otlp exporter: %w", err)
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("init otel resource: %w", err)
	}

	sampleRate := cfg.TracingSampleRate
	if sampleRate <= 0 || sampleRate > 1 {
		sampleRate = DefaultConfig().TracingSampleRate
	}
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	tracer := &otelTracer{
		tracer:     provider.Tracer(cfg.ServiceName),
		propagator: otel.GetTextMapPropagator(),
		provider:   provider,
	}

	return tracer, provider.Shutdown, nil
}

func newOTelTracerFromTracingConfig(cfg TracingConfig) (*otelTracer, func(context.Context) error, error) {
	obsCfg := Config{
		ServiceName:       cfg.ServiceName,
		ServiceVersion:    cfg.ServiceVersion,
		Environment:       cfg.Environment,
		TracingEnabled:    cfg.Enabled,
		TracingEndpoint:   cfg.Endpoint,
		TracingSampleRate: cfg.SampleRate,
	}
	return newOTelTracer(obsCfg)
}

func (t *otelTracer) withName(name string) *otelTracer {
	if name == "" {
		return t
	}
	return &otelTracer{
		tracer:     t.provider.Tracer(name),
		propagator: t.propagator,
		provider:   t.provider,
	}
}

func (t *otelTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	cfg := &spanConfig{
		kind: SpanKindInternal,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	spanKind := trace.SpanKindInternal
	switch cfg.kind {
	case SpanKindServer:
		spanKind = trace.SpanKindServer
	case SpanKindClient:
		spanKind = trace.SpanKindClient
	case SpanKindProducer:
		spanKind = trace.SpanKindProducer
	case SpanKindConsumer:
		spanKind = trace.SpanKindConsumer
	}

	attrs := make([]attribute.KeyValue, 0, len(cfg.attributes))
	for _, attr := range cfg.attributes {
		if kv, ok := fieldToAttribute(attr); ok {
			attrs = append(attrs, kv)
		}
	}

	ctx, span := t.tracer.Start(
		ctx,
		name,
		trace.WithSpanKind(spanKind),
		trace.WithAttributes(attrs...),
	)

	wrapped := &otelSpanWrapper{span: span}
	ctx = context.WithValue(ctx, spanContextKey{}, wrapped)
	return ctx, wrapped
}

func (t *otelTracer) Extract(ctx context.Context, carrier Carrier) context.Context {
	ctx = t.propagator.Extract(ctx, carrierAdapter{carrier: carrier})
	if span := trace.SpanFromContext(ctx); span != nil && span.SpanContext().IsValid() {
		ctx = context.WithValue(ctx, spanContextKey{}, &otelSpanWrapper{span: span})
	}
	return ctx
}

func (t *otelTracer) Inject(ctx context.Context, carrier Carrier) error {
	t.propagator.Inject(ctx, carrierAdapter{carrier: carrier})
	return nil
}

func (s *otelSpanWrapper) End() {
	s.span.End()
}

func (s *otelSpanWrapper) SetStatus(code StatusCode, description string) {
	switch code {
	case StatusError:
		s.span.SetStatus(codes.Error, description)
	case StatusOK:
		s.span.SetStatus(codes.Ok, description)
	default:
		s.span.SetStatus(codes.Unset, description)
	}
}

func (s *otelSpanWrapper) SetAttribute(key string, value interface{}) {
	if kv, ok := fieldToAttribute(Field{Key: key, Value: value}); ok {
		s.span.SetAttributes(kv)
	}
}

func (s *otelSpanWrapper) RecordError(err error) {
	if err == nil {
		return
	}
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

func (s *otelSpanWrapper) AddEvent(name string, attrs ...Field) {
	otelAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if kv, ok := fieldToAttribute(attr); ok {
			otelAttrs = append(otelAttrs, kv)
		}
	}
	s.span.AddEvent(name, trace.WithAttributes(otelAttrs...))
}

func (s *otelSpanWrapper) SpanContext() SpanContext {
	sc := s.span.SpanContext()
	return SpanContext{
		TraceID:    sc.TraceID().String(),
		SpanID:     sc.SpanID().String(),
		TraceFlags: byte(sc.TraceFlags()),
	}
}

func fieldToAttribute(field Field) (attribute.KeyValue, bool) {
	if field.Key == "" {
		return attribute.KeyValue{}, false
	}
	switch v := field.Value.(type) {
	case string:
		return attribute.String(field.Key, v), true
	case int:
		return attribute.Int(field.Key, v), true
	case int64:
		return attribute.Int64(field.Key, v), true
	case float64:
		return attribute.Float64(field.Key, v), true
	case bool:
		return attribute.Bool(field.Key, v), true
	case time.Duration:
		return attribute.Int64(field.Key, int64(v)), true
	case time.Time:
		return attribute.String(field.Key, v.UTC().Format(time.RFC3339Nano)), true
	default:
		return attribute.String(field.Key, fmt.Sprintf("%v", field.Value)), true
	}
}

type carrierAdapter struct {
	carrier Carrier
}

func (c carrierAdapter) Get(key string) string {
	return c.carrier.Get(key)
}

func (c carrierAdapter) Set(key, value string) {
	c.carrier.Set(key, value)
}

func (c carrierAdapter) Keys() []string {
	type keyer interface {
		Keys() []string
	}
	if k, ok := c.carrier.(keyer); ok {
		return k.Keys()
	}
	return nil
}
