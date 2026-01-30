# Production Environment Variables

variable "aws_region" {
  description = "Primary AWS region"
  type        = string
  default     = "us-west-2"
}

variable "dr_region" {
  description = "Disaster recovery AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.2.0.0/16"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for CloudWatch alarms"
  type        = list(string)
  default     = []
}

# =============================================================================
# TEE Hardware Configuration (TEE-HW-001)
# =============================================================================

variable "enable_tee_nitro" {
  description = "Enable AWS Nitro Enclave node group"
  type        = bool
  default     = true
}

variable "enable_tee_sev_snp" {
  description = "Enable AMD SEV-SNP node group"
  type        = bool
  default     = false
}

variable "enable_tee_sgx" {
  description = "Enable Intel SGX node group"
  type        = bool
  default     = false
}

variable "tee_nitro_desired_size" {
  description = "Desired number of Nitro Enclave nodes"
  type        = number
  default     = 2
}

variable "tee_measurement_allowlist" {
  description = "List of allowed enclave measurements"
  type        = list(string)
  default     = []
}

variable "nitro_enclave_image_sha384" {
  description = "SHA384 hash of the approved Nitro Enclave image"
  type        = string
  default     = ""
}
