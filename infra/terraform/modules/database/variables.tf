# CockroachDB Database Module Variables

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "region" {
  description = "AWS region for this database instance"
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the OIDC provider for IRSA"
  type        = string
}

variable "oidc_provider_url" {
  description = "URL of the OIDC provider for IRSA"
  type        = string
}

variable "cockroachdb_chart_version" {
  description = "Version of the CockroachDB Helm chart"
  type        = string
  default     = "13.0.1"
}

variable "cockroachdb_version" {
  description = "Version of the CockroachDB image"
  type        = string
  default     = "v24.1.4"
}

variable "replicas" {
  description = "Number of CockroachDB replicas per region"
  type        = number
  default     = 3
}

variable "join_addresses" {
  description = "List of CockroachDB join addresses for multi-region clustering"
  type        = list(string)
  default     = []
}

variable "storage_size" {
  description = "Persistent volume size for CockroachDB"
  type        = string
  default     = "500Gi"
}

variable "storage_class" {
  description = "Storage class for CockroachDB persistent volumes"
  type        = string
  default     = "gp3-encrypted"
}

variable "cache_size" {
  description = "CockroachDB cache size"
  type        = string
  default     = "25%"
}

variable "max_sql_memory" {
  description = "CockroachDB maximum SQL memory"
  type        = string
  default     = "25%"
}

variable "cpu_request" {
  description = "CPU request for CockroachDB pods"
  type        = string
  default     = "2"
}

variable "cpu_limit" {
  description = "CPU limit for CockroachDB pods"
  type        = string
  default     = "4"
}

variable "memory_request" {
  description = "Memory request for CockroachDB pods"
  type        = string
  default     = "8Gi"
}

variable "memory_limit" {
  description = "Memory limit for CockroachDB pods"
  type        = string
  default     = "16Gi"
}

variable "tolerations" {
  description = "Tolerations for CockroachDB pods"
  type = list(object({
    key      = string
    operator = string
    value    = string
    effect   = string
  }))
  default = []
}

variable "backup_schedule" {
  description = "Cron schedule for CockroachDB backups"
  type        = string
  default     = "0 */6 * * *"
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 365
}

variable "backup_max_age_seconds" {
  description = "Maximum acceptable backup age in seconds before alarm"
  type        = number
  default     = 25200
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for CloudWatch alarms"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
