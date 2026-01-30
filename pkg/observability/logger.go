// Package observability provides structured logging, metrics, and tracing
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// defaultObserver is the default observer implementation
type defaultObserver struct {
	config  Config
	logger  *defaultLogger
	metrics *defaultMetrics
	tracer  *defaultTracer
	closers []func() error
}

// New creates a new Observer with the given configuration
func New(cfg Config) (Observer, error) {
	obs := &defaultObserver{
		config: cfg,
	}

	// Initialize logger
	obs.logger = newDefaultLogger(cfg)

	// Initialize metrics
	if cfg.MetricsEnabled {
		obs.metrics = newDefaultMetrics(cfg)
	}

	// Initialize tracer
	if cfg.TracingEnabled {
		obs.tracer = newDefaultTracer(cfg)
	}

	return obs, nil
}

// Logger returns the structured logger
func (o *defaultObserver) Logger() Logger {
	return o.logger
}

// Metrics returns the metrics registry
func (o *defaultObserver) Metrics() Metrics {
	if o.metrics == nil {
		return &noopMetrics{}
	}
	return o.metrics
}

// Tracer returns the tracer
func (o *defaultObserver) Tracer() Tracer {
	if o.tracer == nil {
		return &noopTracer{}
	}
	return o.tracer
}

// Shutdown gracefully shuts down the observer
func (o *defaultObserver) Shutdown(ctx context.Context) error {
	var errs []error
	for _, closer := range o.closers {
		if err := closer(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// defaultLogger implements Logger
type defaultLogger struct {
	config    Config
	output    io.Writer
	fields    []Field
	mu        sync.Mutex
	sensitive map[string]bool
}

func newDefaultLogger(cfg Config) *defaultLogger {
	sensitive := make(map[string]bool)
	for _, field := range cfg.SensitiveFields {
		sensitive[strings.ToLower(field)] = true
	}

	return &defaultLogger{
		config:    cfg,
		output:    os.Stdout,
		sensitive: sensitive,
	}
}

func (l *defaultLogger) log(level Level, msg string, fields ...Field) {
	if level < l.config.LogLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Combine base fields with provided fields
	allFields := make([]Field, 0, len(l.fields)+len(fields)+4)
	allFields = append(allFields,
		String("level", level.String()),
		String("msg", msg),
		Time("time", time.Now().UTC()),
		String("service", l.config.ServiceName),
	)
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	// Redact sensitive fields
	for i := range allFields {
		if l.isSensitive(allFields[i].Key) {
			allFields[i].Value = "[REDACTED]"
		}
	}

	if l.config.LogFormat == "json" {
		l.outputJSON(allFields)
	} else {
		l.outputText(level, msg, allFields)
	}
}

func (l *defaultLogger) isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for sensitive := range l.sensitive {
		if strings.Contains(lower, sensitive) {
			return true
		}
	}
	return false
}

func (l *defaultLogger) outputJSON(fields []Field) {
	data := make(map[string]interface{})
	for _, f := range fields {
		data[f.Key] = f.Value
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		_, _ = fmt.Fprintf(l.output, `{"error":"failed to marshal log: %v"}`+"\n", err)
		return
	}

	_, _ = fmt.Fprintln(l.output, string(bytes))
}

func (l *defaultLogger) outputText(level Level, msg string, fields []Field) {
	var b strings.Builder
	b.WriteString(time.Now().UTC().Format(time.RFC3339))
	b.WriteString(" ")
	b.WriteString(level.String())
	b.WriteString(" ")
	b.WriteString(msg)

	for _, f := range fields {
		if f.Key == "level" || f.Key == "msg" || f.Key == "time" {
			continue
		}
		b.WriteString(" ")
		b.WriteString(f.Key)
		b.WriteString("=")
		_, _ = fmt.Fprintf(&b, "%v", f.Value)
	}

	_, _ = fmt.Fprintln(l.output, b.String())
}

func (l *defaultLogger) Debug(msg string, fields ...Field) { l.log(LevelDebug, msg, fields...) }
func (l *defaultLogger) Info(msg string, fields ...Field)  { l.log(LevelInfo, msg, fields...) }
func (l *defaultLogger) Warn(msg string, fields ...Field)  { l.log(LevelWarn, msg, fields...) }
func (l *defaultLogger) Error(msg string, fields ...Field) { l.log(LevelError, msg, fields...) }

func (l *defaultLogger) Fatal(msg string, fields ...Field) {
	l.log(LevelFatal, msg, fields...)
	os.Exit(1)
}

func (l *defaultLogger) With(fields ...Field) Logger {
	newLogger := &defaultLogger{
		config:    l.config,
		output:    l.output,
		sensitive: l.sensitive,
		fields:    make([]Field, len(l.fields)+len(fields)),
	}
	copy(newLogger.fields, l.fields)
	copy(newLogger.fields[len(l.fields):], fields)
	return newLogger
}

func (l *defaultLogger) WithContext(ctx context.Context) Logger {
	// Extract trace context if available
	if span := SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		return l.With(
			String("trace_id", sc.TraceID),
			String("span_id", sc.SpanID),
		)
	}
	return l
}

// SpanFromContext extracts span from context
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanContextKey{}).(Span); ok {
		return span
	}
	return nil
}

type spanContextKey struct{}

// defaultMetrics implements Metrics
type defaultMetrics struct {
	config Config
	mu     sync.RWMutex
	// In-memory metrics for demonstration
	// In production, use prometheus client
	counters   map[string]*defaultCounter
	gauges     map[string]*defaultGauge
	histograms map[string]*defaultHistogram
	summaries  map[string]*defaultSummary
}

func newDefaultMetrics(cfg Config) *defaultMetrics {
	return &defaultMetrics{
		config:     cfg,
		counters:   make(map[string]*defaultCounter),
		gauges:     make(map[string]*defaultGauge),
		histograms: make(map[string]*defaultHistogram),
		summaries:  make(map[string]*defaultSummary),
	}
}

func (m *defaultMetrics) Counter(name, help string, labels ...string) Counter {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if c, ok := m.counters[key]; ok {
		return c
	}

	c := &defaultCounter{
		name:   name,
		help:   help,
		labels: labels,
		values: make(map[string]float64),
	}
	m.counters[key] = c
	return c
}

func (m *defaultMetrics) Gauge(name, help string, labels ...string) Gauge {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if g, ok := m.gauges[key]; ok {
		return g
	}

	g := &defaultGauge{
		name:   name,
		help:   help,
		labels: labels,
		values: make(map[string]float64),
	}
	m.gauges[key] = g
	return g
}

func (m *defaultMetrics) Histogram(name, help string, buckets []float64, labels ...string) Histogram {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if h, ok := m.histograms[key]; ok {
		return h
	}

	h := &defaultHistogram{
		name:    name,
		help:    help,
		buckets: buckets,
		labels:  labels,
		values:  make(map[string][]float64),
	}
	m.histograms[key] = h
	return h
}

func (m *defaultMetrics) Summary(name, help string, objectives map[float64]float64, labels ...string) Summary {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := name
	if s, ok := m.summaries[key]; ok {
		return s
	}

	s := &defaultSummary{
		name:       name,
		help:       help,
		objectives: objectives,
		labels:     labels,
		values:     make(map[string][]float64),
	}
	m.summaries[key] = s
	return s
}

type defaultCounter struct {
	name   string
	help   string
	labels []string
	mu     sync.Mutex
	values map[string]float64
}

func (c *defaultCounter) Inc(labels ...string) { c.Add(1, labels...) }

func (c *defaultCounter) Add(value float64, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := strings.Join(labels, ",")
	c.values[key] += value
}

type defaultGauge struct {
	name   string
	help   string
	labels []string
	mu     sync.Mutex
	values map[string]float64
}

func (g *defaultGauge) Set(value float64, labels ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := strings.Join(labels, ",")
	g.values[key] = value
}

func (g *defaultGauge) Inc(labels ...string) { g.Add(1, labels...) }
func (g *defaultGauge) Dec(labels ...string) { g.Sub(1, labels...) }

func (g *defaultGauge) Add(value float64, labels ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := strings.Join(labels, ",")
	g.values[key] += value
}

func (g *defaultGauge) Sub(value float64, labels ...string) {
	g.Add(-value, labels...)
}

type defaultHistogram struct {
	name    string
	help    string
	buckets []float64
	labels  []string
	mu      sync.Mutex
	values  map[string][]float64
}

func (h *defaultHistogram) Observe(value float64, labels ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := strings.Join(labels, ",")
	h.values[key] = append(h.values[key], value)
}

type defaultSummary struct {
	name       string
	help       string
	objectives map[float64]float64
	labels     []string
	mu         sync.Mutex
	values     map[string][]float64
}

func (s *defaultSummary) Observe(value float64, labels ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.Join(labels, ",")
	s.values[key] = append(s.values[key], value)
}

// defaultTracer implements Tracer
type defaultTracer struct {
	config Config
}

func newDefaultTracer(cfg Config) *defaultTracer {
	return &defaultTracer{config: cfg}
}

func (t *defaultTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	cfg := &spanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	span := &defaultSpan{
		name:      name,
		startTime: time.Now(),
		attrs:     make(map[string]interface{}),
		context: SpanContext{
			TraceID: generateID(),
			SpanID:  generateID(),
		},
	}

	for _, attr := range cfg.attributes {
		span.attrs[attr.Key] = attr.Value
	}

	return context.WithValue(ctx, spanContextKey{}, span), span
}

func (t *defaultTracer) Extract(ctx context.Context, carrier Carrier) context.Context {
	traceID := carrier.Get("traceparent")
	if traceID == "" {
		return ctx
	}
	// Simplified: just store trace ID
	span := &defaultSpan{
		context: SpanContext{
			TraceID: traceID,
		},
	}
	return context.WithValue(ctx, spanContextKey{}, span)
}

func (t *defaultTracer) Inject(ctx context.Context, carrier Carrier) error {
	span := SpanFromContext(ctx)
	if span == nil {
		return nil
	}
	sc := span.SpanContext()
	carrier.Set("traceparent", sc.TraceID)
	return nil
}

type defaultSpan struct {
	name      string
	startTime time.Time
	endTime   time.Time
	status    StatusCode
	statusMsg string
	attrs     map[string]interface{}
	events    []spanEvent
	context   SpanContext
	mu        sync.Mutex
}

type spanEvent struct {
	name  string
	time  time.Time
	attrs []Field
}

func (s *defaultSpan) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endTime = time.Now()
}

func (s *defaultSpan) SetStatus(code StatusCode, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = code
	s.statusMsg = description
}

func (s *defaultSpan) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attrs[key] = value
}

func (s *defaultSpan) RecordError(err error) {
	if err == nil {
		return
	}
	s.AddEvent("exception", String("exception.message", err.Error()))
	s.SetStatus(StatusError, err.Error())
}

func (s *defaultSpan) AddEvent(name string, attrs ...Field) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, spanEvent{
		name:  name,
		time:  time.Now(),
		attrs: attrs,
	})
}

func (s *defaultSpan) SpanContext() SpanContext {
	return s.context
}

// generateID generates a simple random ID
func generateID() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

// noopMetrics is a no-op metrics implementation
type noopMetrics struct{}

func (m *noopMetrics) Counter(name, help string, labels ...string) Counter { return &noopCounter{} }
func (m *noopMetrics) Gauge(name, help string, labels ...string) Gauge     { return &noopGauge{} }
func (m *noopMetrics) Histogram(name, help string, buckets []float64, labels ...string) Histogram {
	return &noopHistogram{}
}
func (m *noopMetrics) Summary(name, help string, objectives map[float64]float64, labels ...string) Summary {
	return &noopSummary{}
}

type noopCounter struct{}

func (c *noopCounter) Inc(labels ...string)                {}
func (c *noopCounter) Add(value float64, labels ...string) {}

type noopGauge struct{}

func (g *noopGauge) Set(value float64, labels ...string) {}
func (g *noopGauge) Inc(labels ...string)                {}
func (g *noopGauge) Dec(labels ...string)                {}
func (g *noopGauge) Add(value float64, labels ...string) {}
func (g *noopGauge) Sub(value float64, labels ...string) {}

type noopHistogram struct{}

func (h *noopHistogram) Observe(value float64, labels ...string) {}

type noopSummary struct{}

func (s *noopSummary) Observe(value float64, labels ...string) {}

// noopTracer is a no-op tracer implementation
type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}
func (t *noopTracer) Extract(ctx context.Context, carrier Carrier) context.Context { return ctx }
func (t *noopTracer) Inject(ctx context.Context, carrier Carrier) error            { return nil }

type noopSpan struct{}

func (s *noopSpan) End()                                          {}
func (s *noopSpan) SetStatus(code StatusCode, description string) {}
func (s *noopSpan) SetAttribute(key string, value interface{})    {}
func (s *noopSpan) RecordError(err error)                         {}
func (s *noopSpan) AddEvent(name string, attrs ...Field)          {}
func (s *noopSpan) SpanContext() SpanContext                      { return SpanContext{} }

// ============================================================================
// Observability Convenience Wrapper
// ============================================================================

// defaultObservability implements the Observability interface
type defaultObservability struct {
	observer Observer
	gauges   map[string]Gauge
	counters map[string]Counter
	histos   map[string]Histogram
	mu       sync.RWMutex
}

// NewObservability creates a new Observability instance wrapping an Observer
func NewObservability(obs Observer) Observability {
	return &defaultObservability{
		observer: obs,
		gauges:   make(map[string]Gauge),
		counters: make(map[string]Counter),
		histos:   make(map[string]Histogram),
	}
}

// NewNoopObservability creates a no-op Observability for testing
func NewNoopObservability() Observability {
	return &noopObservability{}
}

// LogInfo logs at info level with key-value pairs
func (o *defaultObservability) LogInfo(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := kvToFields(keysAndValues...)
	o.observer.Logger().WithContext(ctx).Info(msg, fields...)
}

// LogWarn logs at warn level with key-value pairs
func (o *defaultObservability) LogWarn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := kvToFields(keysAndValues...)
	o.observer.Logger().WithContext(ctx).Warn(msg, fields...)
}

// LogError logs at error level with key-value pairs
func (o *defaultObservability) LogError(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := kvToFields(keysAndValues...)
	o.observer.Logger().WithContext(ctx).Error(msg, fields...)
}

// LogDebug logs at debug level with key-value pairs
func (o *defaultObservability) LogDebug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := kvToFields(keysAndValues...)
	o.observer.Logger().WithContext(ctx).Debug(msg, fields...)
}

// RecordGauge records a gauge metric
func (o *defaultObservability) RecordGauge(ctx context.Context, name string, value float64, labels map[string]string) {
	if o.observer.Metrics() == nil {
		return
	}
	o.mu.Lock()
	gauge, exists := o.gauges[name]
	if !exists {
		labelKeys := make([]string, 0, len(labels))
		for k := range labels {
			labelKeys = append(labelKeys, k)
		}
		gauge = o.observer.Metrics().Gauge(name, name, labelKeys...)
		o.gauges[name] = gauge
	}
	o.mu.Unlock()

	labelValues := make([]string, 0, len(labels))
	for _, v := range labels {
		labelValues = append(labelValues, v)
	}
	gauge.Set(value, labelValues...)
}

// RecordHistogram records a histogram metric
func (o *defaultObservability) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) {
	if o.observer.Metrics() == nil {
		return
	}
	o.mu.Lock()
	histo, exists := o.histos[name]
	if !exists {
		labelKeys := make([]string, 0, len(labels))
		for k := range labels {
			labelKeys = append(labelKeys, k)
		}
		histo = o.observer.Metrics().Histogram(name, name, nil, labelKeys...)
		o.histos[name] = histo
	}
	o.mu.Unlock()

	labelValues := make([]string, 0, len(labels))
	for _, v := range labels {
		labelValues = append(labelValues, v)
	}
	histo.Observe(value, labelValues...)
}

// RecordCounter increments a counter metric
func (o *defaultObservability) RecordCounter(ctx context.Context, name string, value float64, labels map[string]string) {
	if o.observer.Metrics() == nil {
		return
	}
	o.mu.Lock()
	counter, exists := o.counters[name]
	if !exists {
		labelKeys := make([]string, 0, len(labels))
		for k := range labels {
			labelKeys = append(labelKeys, k)
		}
		counter = o.observer.Metrics().Counter(name, name, labelKeys...)
		o.counters[name] = counter
	}
	o.mu.Unlock()

	labelValues := make([]string, 0, len(labels))
	for _, v := range labels {
		labelValues = append(labelValues, v)
	}
	counter.Add(value, labelValues...)
}

// kvToFields converts key-value pairs to Fields
func kvToFields(keysAndValues ...interface{}) []Field {
	fields := make([]Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keysAndValues[i])
		}
		fields = append(fields, Any(key, keysAndValues[i+1]))
	}
	return fields
}

// noopObservability is a no-op implementation of Observability
type noopObservability struct{}

func (o *noopObservability) LogInfo(ctx context.Context, msg string, keysAndValues ...interface{})  {}
func (o *noopObservability) LogWarn(ctx context.Context, msg string, keysAndValues ...interface{})  {}
func (o *noopObservability) LogError(ctx context.Context, msg string, keysAndValues ...interface{}) {}
func (o *noopObservability) LogDebug(ctx context.Context, msg string, keysAndValues ...interface{}) {}
func (o *noopObservability) RecordGauge(ctx context.Context, name string, value float64, labels map[string]string) {
}
func (o *noopObservability) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) {
}
func (o *noopObservability) RecordCounter(ctx context.Context, name string, value float64, labels map[string]string) {
}
