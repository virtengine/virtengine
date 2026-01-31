# VirtEngine Cost Optimization Module - Variables

# -----------------------------------------------------------------------------
# General Configuration
# -----------------------------------------------------------------------------

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "cost_center" {
  description = "Cost center for cost allocation"
  type        = string
  default     = "infrastructure"
}

variable "owner" {
  description = "Owner email or team name for cost allocation"
  type        = string
  default     = "platform-team"
}

variable "cluster_name" {
  description = "Name of the EKS cluster for monitoring"
  type        = string
}

# -----------------------------------------------------------------------------
# Budget Configuration
# -----------------------------------------------------------------------------

variable "monthly_budget_limit" {
  description = "Monthly budget limit in USD"
  type        = string
  default     = "15000"
}

variable "compute_budget_limit" {
  description = "Monthly compute budget limit in USD"
  type        = string
  default     = "10000"
}

variable "storage_budget_limit" {
  description = "Monthly storage budget limit in USD"
  type        = string
  default     = "2000"
}

variable "networking_budget_limit" {
  description = "Monthly networking budget limit in USD"
  type        = string
  default     = "2000"
}

variable "data_transfer_budget_limit" {
  description = "Monthly data transfer budget limit in USD"
  type        = string
  default     = "1000"
}

variable "budget_notification_emails" {
  description = "List of email addresses for budget notifications"
  type        = list(string)
  default     = []
}

variable "budget_notification_sns_arns" {
  description = "List of SNS topic ARNs for budget notifications"
  type        = list(string)
  default     = []
}

# -----------------------------------------------------------------------------
# Cost Anomaly Detection Configuration
# -----------------------------------------------------------------------------

variable "enable_organization_monitor" {
  description = "Enable organization-level cost anomaly monitor (requires AWS Organizations)"
  type        = bool
  default     = false
}

variable "anomaly_detection_frequency" {
  description = "Frequency of anomaly detection alerts (DAILY, IMMEDIATE, WEEKLY)"
  type        = string
  default     = "DAILY"

  validation {
    condition     = contains(["DAILY", "IMMEDIATE", "WEEKLY"], var.anomaly_detection_frequency)
    error_message = "Frequency must be DAILY, IMMEDIATE, or WEEKLY."
  }
}

variable "anomaly_notification_email" {
  description = "Email address for anomaly detection alerts"
  type        = string
}

variable "anomaly_notification_sns_arn" {
  description = "SNS topic ARN for anomaly detection alerts (optional)"
  type        = string
  default     = ""
}

variable "anomaly_threshold_absolute" {
  description = "Absolute threshold for anomaly alerts in USD"
  type        = number
  default     = 100
}

# -----------------------------------------------------------------------------
# Resource Cleanup Configuration
# -----------------------------------------------------------------------------

variable "enable_resource_cleanup" {
  description = "Enable automated unused resource cleanup"
  type        = bool
  default     = true
}

variable "cleanup_dry_run" {
  description = "Run cleanup in dry-run mode (no actual deletions)"
  type        = bool
  default     = true
}

variable "cleanup_schedule" {
  description = "CloudWatch Events schedule expression for cleanup (cron or rate)"
  type        = string
  default     = "cron(0 2 ? * SUN *)"  # Every Sunday at 2 AM
}

# -----------------------------------------------------------------------------
# Cost Recommendations Configuration
# -----------------------------------------------------------------------------

variable "enable_cost_recommendations" {
  description = "Enable automated cost recommendation generation"
  type        = bool
  default     = true
}
