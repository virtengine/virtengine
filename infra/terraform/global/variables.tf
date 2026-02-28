# Global Resources Variables

variable "domain_name" {
  description = "Primary domain name for VirtEngine"
  type        = string
  default     = "virtengine.com"
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
  default     = "virtengine"
}

variable "github_repo" {
  description = "GitHub repository for OIDC"
  type        = string
  default     = "virtengine"
}

variable "github_allowed_subjects" {
  description = "Exact GitHub OIDC subject claims allowed to assume the deployment role"
  type        = list(string)
  default = [
    "repo:virtengine/virtengine:ref:refs/heads/main",
    "repo:virtengine/virtengine:ref:refs/heads/release/*",
    "repo:virtengine/virtengine:environment:infra-dev",
    "repo:virtengine/virtengine:environment:infra-staging",
    "repo:virtengine/virtengine:environment:infra-prod",
    "repo:virtengine/virtengine:environment:production-us-east-1",
    "repo:virtengine/virtengine:environment:production-eu-west-1",
    "repo:virtengine/virtengine:environment:production-ap-southeast-1",
    "repo:virtengine/virtengine:environment:production-global",
    "repo:virtengine/virtengine:environment:dr-drill",
  ]

  validation {
    condition     = length(var.github_allowed_subjects) > 0
    error_message = "github_allowed_subjects must include the GitHub refs or environments permitted to deploy."
  }
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
