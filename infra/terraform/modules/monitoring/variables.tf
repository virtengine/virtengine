# Variables for VirtEngine Monitoring Module

variable "project" {
  description = "Project name for resource naming"
  type        = string
  default     = "virtengine"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "rds_instance_id" {
  description = "ID of the RDS instance to monitor"
  type        = string
  default     = ""
}

variable "nat_gateway_ids" {
  description = "List of NAT Gateway IDs to monitor"
  type        = list(string)
  default     = []
}

variable "alb_arn_suffix" {
  description = "ARN suffix of the Application Load Balancer"
  type        = string
  default     = ""
}

variable "alert_emails" {
  description = "List of email addresses for alert notifications"
  type        = list(string)
  default     = []
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
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

variable "grafana_admin_password" {
  description = "Admin password for Grafana"
  type        = string
  sensitive   = true
  default     = ""
}

variable "enable_grafana_ingress" {
  description = "Enable ingress for Grafana"
  type        = bool
  default     = false
}

variable "grafana_hostname" {
  description = "Hostname for Grafana ingress"
  type        = string
  default     = "grafana.internal"
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
