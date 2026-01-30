# IAM Module Variables

variable "name_prefix" {
  description = "Prefix for all IAM resource names"
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

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the EKS OIDC provider"
  type        = string
}

variable "oidc_provider_url" {
  description = "URL of the EKS OIDC provider"
  type        = string
}

variable "kms_key_arns" {
  description = "List of KMS key ARNs for decryption"
  type        = list(string)
  default     = []
}

variable "backup_bucket_arn" {
  description = "ARN of the backup S3 bucket"
  type        = string
}

variable "manifests_bucket_arn" {
  description = "ARN of the manifests S3 bucket"
  type        = string
}

variable "enable_github_actions_role" {
  description = "Create IAM role for GitHub Actions OIDC"
  type        = bool
  default     = false
}

variable "github_org" {
  description = "GitHub organization name"
  type        = string
  default     = "virtengine-network"
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "virtengine"
}

variable "state_bucket_arn" {
  description = "ARN of the Terraform state bucket (for GitHub Actions)"
  type        = string
  default     = null
}

variable "locks_table_arn" {
  description = "ARN of the DynamoDB locks table (for GitHub Actions)"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
