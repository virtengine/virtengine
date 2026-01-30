# VirtEngine TEE Hardware Module - Variables
# TEE-HW-001: Deploy TEE hardware & attestation in production

# =============================================================================
# General Configuration
# =============================================================================

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID for security groups"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs for node groups"
  type        = list(string)
}

variable "node_role_arn" {
  description = "ARN of the IAM role for EKS nodes"
  type        = string
}

variable "node_role_name" {
  description = "Name of the IAM role for EKS nodes"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of KMS key for encryption"
  type        = string
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for CloudWatch alarms"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# =============================================================================
# Platform Enablement
# =============================================================================

variable "enable_nitro" {
  description = "Enable AWS Nitro Enclave node group"
  type        = bool
  default     = true
}

variable "enable_sev_snp" {
  description = "Enable AMD SEV-SNP node group"
  type        = bool
  default     = false
}

variable "enable_sgx" {
  description = "Enable Intel SGX node group"
  type        = bool
  default     = false
}

# =============================================================================
# Nitro Enclave Configuration
# =============================================================================

variable "nitro_instance_types" {
  description = "Instance types for Nitro Enclave nodes"
  type        = list(string)
  default     = null
}

variable "nitro_desired_size" {
  description = "Desired number of Nitro Enclave nodes"
  type        = number
  default     = 2
}

variable "nitro_min_size" {
  description = "Minimum number of Nitro Enclave nodes"
  type        = number
  default     = 2
}

variable "nitro_max_size" {
  description = "Maximum number of Nitro Enclave nodes"
  type        = number
  default     = 6
}

variable "nitro_disk_size" {
  description = "EBS volume size for Nitro nodes in GB"
  type        = number
  default     = 100
}

variable "nitro_enclave_memory_mb" {
  description = "Memory to allocate for Nitro Enclaves in MB"
  type        = number
  default     = 2048
}

variable "nitro_enclave_cpu_count" {
  description = "CPU cores to allocate for Nitro Enclaves"
  type        = number
  default     = 2
}

variable "nitro_enclave_image_sha384" {
  description = "SHA384 hash of the approved Nitro Enclave image (for KMS attestation)"
  type        = string
  default     = ""
}

# =============================================================================
# SEV-SNP Configuration
# =============================================================================

variable "sev_snp_instance_types" {
  description = "Instance types for SEV-SNP nodes"
  type        = list(string)
  default     = null
}

variable "sev_snp_desired_size" {
  description = "Desired number of SEV-SNP nodes"
  type        = number
  default     = 2
}

variable "sev_snp_min_size" {
  description = "Minimum number of SEV-SNP nodes"
  type        = number
  default     = 2
}

variable "sev_snp_max_size" {
  description = "Maximum number of SEV-SNP nodes"
  type        = number
  default     = 6
}

variable "sev_snp_disk_size" {
  description = "EBS volume size for SEV-SNP nodes in GB"
  type        = number
  default     = 100
}

# =============================================================================
# SGX Configuration
# =============================================================================

variable "sgx_instance_types" {
  description = "Instance types for SGX nodes"
  type        = list(string)
  default     = null
}

variable "sgx_desired_size" {
  description = "Desired number of SGX nodes"
  type        = number
  default     = 2
}

variable "sgx_min_size" {
  description = "Minimum number of SGX nodes"
  type        = number
  default     = 2
}

variable "sgx_max_size" {
  description = "Maximum number of SGX nodes"
  type        = number
  default     = 6
}

variable "sgx_disk_size" {
  description = "EBS volume size for SGX nodes in GB"
  type        = number
  default     = 100
}

variable "sgx_pccs_endpoint" {
  description = "Intel PCCS (Provisioning Certificate Caching Service) endpoint"
  type        = string
  default     = "https://localhost:8081/sgx/certification/v4/"
}

# =============================================================================
# Attestation Configuration
# =============================================================================

variable "measurement_allowlist" {
  description = "JSON array of allowed enclave measurements (MRENCLAVE/MRSIGNER values)"
  type        = list(string)
  default     = []
}

variable "min_tcb_version" {
  description = "Minimum TCB version to accept (format: BL.TEE.SNP.UC for SEV-SNP)"
  type        = string
  default     = "2.0.8.115"
}
