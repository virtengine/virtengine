# S3 Module Variables

variable "name_prefix" {
  description = "Prefix for all bucket names"
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

variable "create_ml_bucket" {
  description = "Create ML model weights bucket"
  type        = bool
  default     = true
}

variable "create_state_bucket" {
  description = "Create Terraform state bucket"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
