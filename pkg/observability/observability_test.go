// Package observability provides observability testing
package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	t.Run("logs at correct level", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogLevel = LevelInfo
		cfg.LogFormat = "json"

		logger := newDefaultLogger(cfg)
		var buf bytes.Buffer
		logger.output = &buf

		logger.Debug("debug message") // Should be filtered
		logger.Info("info message")
		logger.Warn("warn message")

		output := buf.String()
		if strings.Contains(output, "debug message") {
			t.Error("debug should be filtered when level is info")
		}
		if !strings.Contains(output, "info message") {
			t.Error("info should be logged")
		}
		if !strings.Contains(output, "warn message") {
			t.Error("warn should be logged")
		}
	})

	t.Run("redacts sensitive fields", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogFormat = "json"

		logger := newDefaultLogger(cfg)
		var buf bytes.Buffer
		logger.output = &buf

		logger.Info("test",
			String("password", "secret123"),
			String("token", "abc123"),
			String("api_key", "xyz789"),
			String("safe_field", "visible"),
		)

		output := buf.String()
		if strings.Contains(output, "secret123") {
			t.Error("password should be redacted")
		}
		if strings.Contains(output, "abc123") {
			t.Error("token should be redacted")
		}
		if strings.Contains(output, "xyz789") {
			t.Error("api_key should be redacted")
		}
		if !strings.Contains(output, "visible") {
			t.Error("safe fields should not be redacted")
		}
		if !strings.Contains(output, "[REDACTED]") {
			t.Error("should contain REDACTED marker")
		}
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogFormat = "json"

		logger := newDefaultLogger(cfg)
		var buf bytes.Buffer
		logger.output = &buf

		logger.Info("test message", String("key", "value"))

		var parsed map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &parsed)
		if err != nil {
			t.Errorf("output should be valid JSON: %v", err)
		}

		if parsed["msg"] != "test message" {
			t.Error("message not found in output")
		}
		if parsed["key"] != "value" {
			t.Error("field not found in output")
		}
	})

	t.Run("outputs text format", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogFormat = "text"

		logger := newDefaultLogger(cfg)
		var buf bytes.Buffer
		logger.output = &buf

		logger.Info("test message", String("key", "value"))

		output := buf.String()
		if !strings.Contains(output, "INFO") {
			t.Error("level not in output")
		}
		if !strings.Contains(output, "test message") {
			t.Error("message not in output")
		}
		if !strings.Contains(output, "key=value") {
			t.Error("field not in output")
		}
	})

	t.Run("with adds fields", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.LogFormat = "json"

		logger := newDefaultLogger(cfg)
		var buf bytes.Buffer
		logger.output = &buf

		childLogger := logger.With(String("request_id", "abc123"))
		childLogger.(*defaultLogger).output = &buf

		childLogger.Info("test")

		output := buf.String()
		if !strings.Contains(output, "abc123") {
			t.Error("with field should be present")
		}
	})
}

func TestFields(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		f := String("key", "value")
		if f.Key != "key" || f.Value != "value" {
			t.Error("String field incorrect")
		}
	})

	t.Run("Int field", func(t *testing.T) {
		f := Int("key", 42)
		if f.Key != "key" || f.Value != 42 {
			t.Error("Int field incorrect")
		}
	})

	t.Run("Int64 field", func(t *testing.T) {
		f := Int64("key", 9223372036854775807)
		if f.Key != "key" {
			t.Error("Int64 field incorrect")
		}
	})

	t.Run("Float64 field", func(t *testing.T) {
		f := Float64("key", 3.14)
		if f.Key != "key" || f.Value != 3.14 {
			t.Error("Float64 field incorrect")
		}
	})

	t.Run("Bool field", func(t *testing.T) {
		f := Bool("key", true)
		if f.Key != "key" || f.Value != true {
			t.Error("Bool field incorrect")
		}
	})

	t.Run("Error field", func(t *testing.T) {
		f := Error(nil)
		if f.Key != "error" || f.Value != nil {
			t.Error("nil Error field incorrect")
		}
	})
}

func TestMetrics(t *testing.T) {
	t.Run("Counter increments", func(t *testing.T) {
		cfg := DefaultConfig()
		metrics := newDefaultMetrics(cfg)

		counter := metrics.Counter("test_counter", "A test counter")
		counter.Inc()
		counter.Inc()
		counter.Add(5)

		// Internal check
		c := metrics.counters["test_counter"]
		if c.values[""] != 7 {
			t.Errorf("expected 7, got %v", c.values[""])
		}
	})

	t.Run("Counter with labels", func(t *testing.T) {
		cfg := DefaultConfig()
		metrics := newDefaultMetrics(cfg)

		counter := metrics.Counter("requests_total", "Total requests", "method", "status")
		counter.Inc("GET", "200")
		counter.Inc("GET", "200")
		counter.Inc("POST", "201")

		c := metrics.counters["requests_total"]
		if c.values["GET,200"] != 2 {
			t.Error("GET,200 count incorrect")
		}
		if c.values["POST,201"] != 1 {
			t.Error("POST,201 count incorrect")
		}
	})

	t.Run("Gauge operations", func(t *testing.T) {
		cfg := DefaultConfig()
		metrics := newDefaultMetrics(cfg)

		gauge := metrics.Gauge("temperature", "Temperature")
		gauge.Set(20)
		gauge.Inc()
		gauge.Add(5)
		gauge.Dec()
		gauge.Sub(2)

		g := metrics.gauges["temperature"]
		if g.values[""] != 23 { // 20 + 1 + 5 - 1 - 2 = 23
			t.Errorf("expected 23, got %v", g.values[""])
		}
	})

	t.Run("Histogram observes", func(t *testing.T) {
		cfg := DefaultConfig()
		metrics := newDefaultMetrics(cfg)

		histogram := metrics.Histogram("latency", "Latency", []float64{0.1, 0.5, 1, 5})
		histogram.Observe(0.3)
		histogram.Observe(0.7)
		histogram.Observe(2.5)

		h := metrics.histograms["latency"]
		if len(h.values[""]) != 3 {
			t.Error("histogram should have 3 observations")
		}
	})
}

func TestTracer(t *testing.T) {
	t.Run("creates span", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TracingEnabled = true
		tracer := newDefaultTracer(cfg)

		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		if span == nil {
			t.Error("span should not be nil")
		}

		sc := span.SpanContext()
		if sc.TraceID == "" {
			t.Error("trace ID should be set")
		}
		if sc.SpanID == "" {
			t.Error("span ID should be set")
		}

		// Verify span can be retrieved from context
		retrieved := SpanFromContext(ctx)
		if retrieved == nil {
			t.Error("span should be in context")
		}
	})

	t.Run("span attributes", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TracingEnabled = true
		tracer := newDefaultTracer(cfg)

		_, span := tracer.Start(context.Background(), "test-span",
			WithAttributes(String("key", "value")),
		)
		defer span.End()

		span.SetAttribute("another_key", "another_value")

		ds := span.(*defaultSpan)
		if ds.attrs["key"] != "value" {
			t.Error("initial attribute not set")
		}
		if ds.attrs["another_key"] != "another_value" {
			t.Error("added attribute not set")
		}
	})

	t.Run("span events", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TracingEnabled = true
		tracer := newDefaultTracer(cfg)

		_, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		span.AddEvent("test-event", String("event_key", "event_value"))

		ds := span.(*defaultSpan)
		if len(ds.events) != 1 {
			t.Error("event not recorded")
		}
		if ds.events[0].name != "test-event" {
			t.Error("event name incorrect")
		}
	})

	t.Run("carrier injection and extraction", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TracingEnabled = true
		tracer := newDefaultTracer(cfg)

		ctx, span := tracer.Start(context.Background(), "test-span")
		sc := span.SpanContext()

		carrier := MapCarrier{}
		err := tracer.Inject(ctx, carrier)
		if err != nil {
			t.Errorf("inject error: %v", err)
		}

		if carrier.Get("traceparent") != sc.TraceID {
			t.Error("trace ID not injected")
		}

		newCtx := tracer.Extract(context.Background(), carrier)
		extractedSpan := SpanFromContext(newCtx)
		if extractedSpan == nil {
			t.Error("span should be extracted from context")
		}
	})
}

func TestObserver(t *testing.T) {
	t.Run("creates observer with config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.ServiceName = "test-service"
		cfg.MetricsEnabled = true
		cfg.TracingEnabled = true

		obs, err := New(cfg)
		if err != nil {
			t.Fatalf("failed to create observer: %v", err)
		}

		if obs.Logger() == nil {
			t.Error("logger should not be nil")
		}
		if obs.Metrics() == nil {
			t.Error("metrics should not be nil")
		}
		if obs.Tracer() == nil {
			t.Error("tracer should not be nil")
		}
	})

	t.Run("shutdown completes", func(t *testing.T) {
		cfg := DefaultConfig()
		obs, _ := New(cfg)

		err := obs.Shutdown(context.Background())
		if err != nil {
			t.Errorf("shutdown error: %v", err)
		}
	})
}

func TestNoopImplementations(t *testing.T) {
	t.Run("noop metrics don't panic", func(t *testing.T) {
		m := &noopMetrics{}
		counter := m.Counter("test", "help")
		counter.Inc()
		counter.Add(5)

		gauge := m.Gauge("test", "help")
		gauge.Set(10)
		gauge.Inc()
		gauge.Dec()
		gauge.Add(5)
		gauge.Sub(2)

		histogram := m.Histogram("test", "help", []float64{1, 5, 10})
		histogram.Observe(3)

		summary := m.Summary("test", "help", map[float64]float64{0.5: 0.05})
		summary.Observe(3)
	})

	t.Run("noop tracer doesn't panic", func(t *testing.T) {
		tracer := &noopTracer{}
		ctx, span := tracer.Start(context.Background(), "test")
		span.SetAttribute("key", "value")
		span.AddEvent("event")
		span.SetStatus(StatusOK, "ok")
		span.RecordError(nil)
		span.End()

		carrier := MapCarrier{}
		_ = tracer.Inject(ctx, carrier)
		_ = tracer.Extract(ctx, carrier)
	})
}
