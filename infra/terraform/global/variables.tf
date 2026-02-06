# Global Resources Variables

variable "domain_name" {
  description = "Primary domain name for VirtEngine"
  type        = string
  default     = "virtengine.io"
}

variable "create_hosted_zone" {
  description = "Whether to create a new Route53 hosted zone"
  type        = bool
  default     = false
}

variable "admin_role_arns" {
  description = "IAM role ARNs allowed to assume cross-region admin role"
  type        = list(string)
  default     = []
}

variable "github_org" {
  description = "GitHub organization for OIDC"
  type        = string
  default     = "virtengine-gh"
}

variable "github_repo" {
  description = "GitHub repository for OIDC"
  type        = string
  default     = "virtengine"
}

variable "regional_endpoints" {
  description = "Map of regional endpoint configurations for DNS"
  type = map(object({
    lb_dns_name       = string
    lb_zone_id        = string
    health_check_fqdn = string
    health_check_port = number
    health_check_path = string
  }))
  default = {}
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for CloudWatch alarms"
  type        = list(string)
  default     = []
}
