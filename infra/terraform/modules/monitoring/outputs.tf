# Outputs for VirtEngine Monitoring Module

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = aws_sns_topic.alerts.arn
}

output "dashboard_url" {
  description = "URL for the CloudWatch dashboard"
  value       = "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}

output "monitoring_namespace" {
  description = "Kubernetes namespace for monitoring"
  value       = kubernetes_namespace.monitoring.metadata[0].name
}

output "prometheus_endpoint" {
  description = "Internal endpoint for Prometheus"
  value       = "http://kube-prometheus-stack-prometheus.monitoring.svc.cluster.local:9090"
}

output "grafana_endpoint" {
  description = "Internal endpoint for Grafana"
  value       = "http://kube-prometheus-stack-grafana.monitoring.svc.cluster.local:80"
}

output "alertmanager_endpoint" {
  description = "Internal endpoint for Alertmanager"
  value       = "http://kube-prometheus-stack-alertmanager.monitoring.svc.cluster.local:9093"
}
