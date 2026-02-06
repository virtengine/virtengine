# DNS Module Variables

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "domain_name" {
  description = "Primary domain name"
  type        = string
  default     = "virtengine.io"
}

variable "create_hosted_zone" {
  description = "Whether to create a new hosted zone or use existing"
  type        = bool
  default     = false
}

variable "regional_endpoints" {
  description = "Map of regional endpoint configurations"
  type = map(object({
    lb_dns_name       = string
    lb_zone_id        = string
    health_check_fqdn = string
    health_check_port = number
    health_check_path = string
  }))
}

variable "primary_region" {
  description = "Primary region for failover routing"
  type        = string
  default     = "us-east-1"
}

variable "secondary_region" {
  description = "Secondary region for failover routing"
  type        = string
  default     = "eu-west-1"
}

variable "enable_failover" {
  description = "Enable failover routing records"
  type        = bool
  default     = true
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
