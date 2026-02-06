# Task 31B: Distributed Tracing Instrumentation

**vibe-kanban ID:** `6386f6fb-dbc0-4919-8b3c-9d70b5305a8c`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31B |
| **Title** | feat(observability): Distributed tracing instrumentation |
| **Priority** | P1 |
| **Wave** | 2 |
| **Estimated LOC** | 2000 |
| **Duration** | 2 weeks |
| **Dependencies** | OpenTelemetry deps exist but unwired |
| **Blocking** | None |

---

## Problem Statement

Production observability requires end-to-end distributed tracing across all services. While OpenTelemetry dependencies exist in go.mod, they are not wired into the application. Without distributed tracing:

1. Debugging production issues requires log correlation by hand
2. Latency bottlenecks are difficult to identify
3. Cross-service request flows are opaque
4. Performance regressions go undetected

### Current State Analysis

```
go.mod:
  go.opentelemetry.io/otel      ✅ Dependency exists

x/*/keeper/*.go:                ❌ No span creation
pkg/provider_daemon/:           ❌ No trace context propagation
app/ante.go:                    ❌ No transaction tracing
```

---

## Acceptance Criteria

### AC-1: Keeper Instrumentation
- [ ] Add trace context to all x/ module keepers
- [ ] Create spans for all MsgServer handlers
- [ ] Create spans for all QueryServer handlers
- [ ] Propagate trace context in cross-module calls
- [ ] Add attributes for operation metadata

### AC-2: Provider Daemon Tracing
- [ ] HTTP server request tracing with spans
- [ ] WebSocket event tracing
- [ ] Waldur API call instrumentation
- [ ] Adapter operation spans (K8s, SLURM, etc.)
- [ ] Async job tracing

### AC-3: Collector Configuration
- [ ] Jaeger or Tempo deployment configuration
- [ ] Sampling rate configuration (adaptive sampling)
- [ ] Trace aggregation rules
- [ ] Retention policies (30 days default)
- [ ] Production vs development settings

### AC-4: Service Correlation
- [ ] Trace ID propagation across services
- [ ] Portal → Chain request correlation
- [ ] Provider daemon → Waldur tracing
- [ ] Async event correlation
- [ ] Error tracking with trace context

---

## Technical Requirements

### OpenTelemetry Setup

```go
// pkg/observability/tracing/tracer.go

package tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Config struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    OTLPEndpoint   string
    SampleRate     float64
}

func InitTracer(cfg Config) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptrace.New(
        context.Background(),
        otlptracegrpc.NewClient(
            otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
            otlptracegrpc.WithInsecure(),
        ),
    )
    if err != nil {
        return nil, err
    }

    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceName(cfg.ServiceName),
        semconv.ServiceVersion(cfg.ServiceVersion),
        semconv.DeploymentEnvironment(cfg.Environment),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(
            sdktrace.ParentBased(
                sdktrace.TraceIDRatioBased(cfg.SampleRate),
            ),
        ),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(
        propagation.NewCompositeTextMapPropagator(
            propagation.TraceContext{},
            propagation.Baggage{},
        ),
    )

    return tp, nil
}
```

### Keeper Instrumentation

```go
// x/veid/keeper/tracing.go

package keeper

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("x/veid/keeper")

func (k Keeper) SubmitScope(ctx context.Context, msg *types.MsgSubmitScope) (*types.MsgSubmitScopeResponse, error) {
    ctx, span := tracer.Start(ctx, "Keeper.SubmitScope",
        trace.WithAttributes(
            attribute.String("account", msg.Account),
            attribute.String("scope_type", msg.ScopeType.String()),
        ),
    )
    defer span.End()

    // ... existing logic ...

    span.SetAttributes(
        attribute.Bool("success", true),
        attribute.String("scope_id", scopeID),
    )
    return &types.MsgSubmitScopeResponse{ScopeId: scopeID}, nil
}

// TracingMiddleware wraps keeper with automatic span creation
func TracingMiddleware(k Keeper) Keeper {
    return &tracedKeeper{inner: k}
}

type tracedKeeper struct {
    inner Keeper
}

func (tk *tracedKeeper) GetIdentityRecord(ctx context.Context, addr string) (*types.IdentityRecord, error) {
    ctx, span := tracer.Start(ctx, "Keeper.GetIdentityRecord",
        trace.WithAttributes(attribute.String("address", addr)),
    )
    defer span.End()
    
    rec, err := tk.inner.GetIdentityRecord(ctx, addr)
    if err != nil {
        span.RecordError(err)
    }
    return rec, err
}
```

### HTTP Server Middleware

```go
// pkg/provider_daemon/middleware/tracing.go

package middleware

import (
    "net/http"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func TracingMiddleware(next http.Handler) http.Handler {
    return otelhttp.NewHandler(next, "provider-api",
        otelhttp.WithPropagators(
            propagation.NewCompositeTextMapPropagator(
                propagation.TraceContext{},
                propagation.Baggage{},
            ),
        ),
    )
}

// Wrap HTTP client for outbound calls
func TracedHTTPClient() *http.Client {
    return &http.Client{
        Transport: otelhttp.NewTransport(http.DefaultTransport),
    }
}
```

### gRPC Interceptors

```go
// pkg/observability/tracing/grpc.go

package tracing

import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return otelgrpc.UnaryServerInterceptor()
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
    return otelgrpc.StreamServerInterceptor()
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return otelgrpc.UnaryClientInterceptor()
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
    return otelgrpc.StreamClientInterceptor()
}
```

---

## Collector Configuration

### Jaeger Deployment

```yaml
# deploy/observability/jaeger.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: virtengine-observability
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:1.50
        ports:
        - containerPort: 6831  # UDP - Thrift compact
        - containerPort: 6832  # UDP - Thrift binary
        - containerPort: 16686 # UI
        - containerPort: 14268 # HTTP - collector
        - containerPort: 4317  # gRPC - OTLP
        - containerPort: 4318  # HTTP - OTLP
        env:
        - name: COLLECTOR_OTLP_ENABLED
          value: "true"
        - name: SPAN_STORAGE_TYPE
          value: "badger"
        - name: BADGER_DIRECTORY_VALUE
          value: "/badger/data"
        - name: BADGER_DIRECTORY_KEY
          value: "/badger/key"
        volumeMounts:
        - name: badger-storage
          mountPath: /badger
      volumes:
      - name: badger-storage
        persistentVolumeClaim:
          claimName: jaeger-badger-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger
  namespace: virtengine-observability
spec:
  ports:
  - name: otlp-grpc
    port: 4317
  - name: otlp-http
    port: 4318
  - name: ui
    port: 16686
  selector:
    app: jaeger
```

### Docker Compose (Development)

```yaml
# docker-compose.observability.yaml (update)
services:
  jaeger:
    image: jaegertracing/all-in-one:1.50
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
```

---

## Directory Structure

```
pkg/observability/tracing/
├── tracer.go             # Tracer initialization
├── config.go             # Configuration types
├── grpc.go               # gRPC interceptors
├── http.go               # HTTP middleware
└── context.go            # Context utilities

x/veid/keeper/
├── tracing.go            # VEID keeper tracing
├── (existing files updated with spans)

x/market/keeper/
├── tracing.go            # Market keeper tracing

x/escrow/keeper/
├── tracing.go            # Escrow keeper tracing

pkg/provider_daemon/
├── middleware/
│   └── tracing.go        # HTTP tracing middleware
├── (existing files updated)
```

---

## Testing Requirements

### Unit Tests
- Span creation verification
- Attribute propagation
- Error recording

### Integration Tests
- End-to-end trace correlation
- Multi-service trace linking
- Sampling behavior verification

### Performance Tests
- Measure overhead (< 5% latency impact)
- Memory usage under high trace volume
- Exporter backpressure handling

---

## Security Considerations

1. **Sensitive Data**: Never include PII in span attributes
2. **Sampling**: Use adaptive sampling to prevent data explosion
3. **Access Control**: Restrict Jaeger UI access to operators only
4. **Retention**: Auto-delete traces older than retention policy
