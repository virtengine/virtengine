# VirtEngine Service Mesh (Istio)

This directory provides a baseline Istio service mesh setup for VirtEngine microservices, including mTLS, traffic management, circuit breaking, retries, tracing, and service-to-service authorization.

## What is included

- Istio install manifest (Operator) with access logging and OpenTelemetry tracing.
- Namespace onboarding with sidecar injection.
- Strict mTLS policy for in-namespace workloads.
- Service-to-service AuthorizationPolicy (mesh identities only).
- Traffic policies (retries, timeouts) and circuit breakers (outlier detection).
- OpenTelemetry Collector deployment for trace ingestion.

## Install Istio

1) Install the Istio operator or use istioctl (preferred). Example:

```
istioctl operator init
kubectl apply -f deploy/istio/install/istio-operator.yaml
```

2) Wait for Istio control plane to be ready in `istio-system`.

## Apply mesh policies and observability

```
kubectl apply -k deploy/istio
```

## Customize for your cluster

- Service names: update the `host` fields in `deploy/istio/traffic/*` to match your Kubernetes Services.
- AuthorizationPolicy: if your ingress gateway runs outside `istio-system`, add its namespace to `deploy/istio/base/authorizationpolicy.yaml`.
- Tracing backend: `deploy/istio/observability/otel-collector.yaml` exports to logging only; wire exporters to Tempo/Jaeger/OTLP as needed.
- mTLS exceptions: if any non-mesh workloads must talk to these services, add explicit policies or create an exception PeerAuthentication.

## Best practices baked in

- Enforced mTLS with mesh identity checks.
- Explicit traffic policies to bound retries and timeouts.
- Outlier detection for circuit breaking to reduce cascading failures.
- Centralized tracing via OpenTelemetry collector.
- Access logs enabled at the mesh level (see IstioOperator meshConfig).
