# US-East-1 Region Variables

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.10.0.0/16"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "validator_count" {
  description = "Number of validators in this region"
  type        = number
  default     = 4
}

variable "cockroachdb_join_addresses" {
  description = "CockroachDB join addresses for multi-region cluster"
  type        = list(string)
  default     = []
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for CloudWatch alarms"
  type        = list(string)
  default     = []
}

variable "alert_sns_topic_arn" {
  description = "SNS topic ARN for Prometheus alerts"
  type        = string
  default     = ""
}

variable "federation_targets" {
  description = "Prometheus federation targets from other regions"
  type = list(object({
    region              = string
    prometheus_endpoint = string
  }))
  default = []
}

variable "enable_cross_region_peering" {
  description = "Enable VPC peering to other regions"
  type        = bool
  default     = false
}

variable "eu_west_vpc_id" {
  description = "VPC ID for EU-West-1 region (for peering)"
  type        = string
  default     = ""
}

variable "ap_southeast_vpc_id" {
  description = "VPC ID for AP-Southeast-1 region (for peering)"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default = {
    Project   = "virtengine"
    ManagedBy = "terraform"
  }
}
