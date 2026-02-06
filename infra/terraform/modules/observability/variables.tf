# Observability Module Variables

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "is_primary_region" {
  description = "Whether this is the primary region (hosts Grafana, central federation)"
  type        = bool
  default     = false
}

variable "prometheus_stack_version" {
  description = "Version of the kube-prometheus-stack Helm chart"
  type        = string
  default     = "56.6.2"
}

variable "prometheus_retention" {
  description = "Prometheus data retention period"
  type        = string
  default     = "15d"
}

variable "prometheus_retention_size" {
  description = "Maximum size of Prometheus data"
  type        = string
  default     = "50GB"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "100Gi"
}

variable "storage_class" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "gp3"
}

variable "central_prometheus_url" {
  description = "Remote-write URL for central Prometheus (non-primary regions)"
  type        = string
  default     = ""
}

variable "federation_targets" {
  description = "List of federation targets for the primary region"
  type = list(object({
    region              = string
    prometheus_endpoint = string
  }))
  default = []
}

variable "alert_sns_topic_arn" {
  description = "SNS topic ARN for alerts"
  type        = string
}

variable "pagerduty_service_key" {
  description = "PagerDuty service key for critical alerts"
  type        = string
  default     = ""
  sensitive   = true
}

variable "loki_stack_version" {
  description = "Version of the Loki stack Helm chart"
  type        = string
  default     = "2.10.0"
}

variable "loki_storage_size" {
  description = "Storage size for Loki"
  type        = string
  default     = "50Gi"
}

variable "loki_s3_bucket" {
  description = "S3 bucket for Loki log storage"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
