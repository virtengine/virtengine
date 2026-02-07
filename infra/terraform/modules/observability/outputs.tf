# Observability Module Outputs

output "monitoring_namespace" {
  description = "Kubernetes namespace for monitoring"
  value       = kubernetes_namespace.monitoring.metadata[0].name
}

output "prometheus_endpoint" {
  description = "Internal endpoint for Prometheus"
  value       = "http://kube-prometheus-stack-prometheus.${local.namespace}.svc.cluster.local:9090"
}

output "alertmanager_endpoint" {
  description = "Internal endpoint for Alertmanager"
  value       = "http://kube-prometheus-stack-alertmanager.${local.namespace}.svc.cluster.local:9093"
}

output "grafana_endpoint" {
  description = "Internal endpoint for Grafana (primary region only)"
  value       = var.is_primary_region ? "http://kube-prometheus-stack-grafana.${local.namespace}.svc.cluster.local:80" : ""
}

output "loki_endpoint" {
  description = "Internal endpoint for Loki"
  value       = "http://loki.${local.namespace}.svc.cluster.local:3100"
}
