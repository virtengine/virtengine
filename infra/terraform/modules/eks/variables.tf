# Variables for VirtEngine EKS Module

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version for the EKS cluster"
  type        = string
  default     = "1.29"
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for the EKS cluster"
  type        = list(string)
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the EKS cluster"
  type        = list(string)
}

variable "cluster_security_group_id" {
  description = "Security group ID for the EKS cluster"
  type        = string
}

variable "endpoint_private_access" {
  description = "Enable private API server endpoint"
  type        = bool
  default     = true
}

variable "endpoint_public_access" {
  description = "Enable public API server endpoint"
  type        = bool
  default     = true
}

variable "public_access_cidrs" {
  description = "List of CIDRs allowed to access public API endpoint"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "enabled_cluster_log_types" {
  description = "List of control plane log types to enable"
  type        = list(string)
  default     = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
}

variable "cluster_log_retention_days" {
  description = "Number of days to retain EKS control plane logs"
  type        = number
  default     = 30
}

variable "kms_key_arn" {
  description = "ARN of KMS key for secrets encryption (if not provided, a new key is created)"
  type        = string
  default     = ""
}

# System Node Group Configuration
variable "system_node_instance_types" {
  description = "Instance types for system node group"
  type        = list(string)
  default     = ["t3.large"]
}

variable "system_node_disk_size" {
  description = "Disk size in GB for system nodes"
  type        = number
  default     = 50
}

variable "system_node_desired_size" {
  description = "Desired number of system nodes"
  type        = number
  default     = 2
}

variable "system_node_max_size" {
  description = "Maximum number of system nodes"
  type        = number
  default     = 4
}

variable "system_node_min_size" {
  description = "Minimum number of system nodes"
  type        = number
  default     = 2
}

# Application Node Group Configuration
variable "app_node_instance_types" {
  description = "Instance types for application node group"
  type        = list(string)
  default     = ["m5.xlarge", "m5.2xlarge"]
}

variable "app_node_capacity_type" {
  description = "Capacity type for application nodes (ON_DEMAND or SPOT)"
  type        = string
  default     = "ON_DEMAND"
}

variable "app_node_disk_size" {
  description = "Disk size in GB for application nodes"
  type        = number
  default     = 100
}

variable "app_node_desired_size" {
  description = "Desired number of application nodes"
  type        = number
  default     = 3
}

variable "app_node_max_size" {
  description = "Maximum number of application nodes"
  type        = number
  default     = 10
}

variable "app_node_min_size" {
  description = "Minimum number of application nodes"
  type        = number
  default     = 2
}

# Chain Node Group Configuration
variable "chain_node_instance_types" {
  description = "Instance types for chain node group"
  type        = list(string)
  default     = ["m5.2xlarge"]
}

variable "chain_node_disk_size" {
  description = "Disk size in GB for chain nodes (needs space for blockchain data)"
  type        = number
  default     = 500
}

variable "chain_node_desired_size" {
  description = "Desired number of chain nodes"
  type        = number
  default     = 3
}

variable "chain_node_max_size" {
  description = "Maximum number of chain nodes"
  type        = number
  default     = 5
}

variable "chain_node_min_size" {
  description = "Minimum number of chain nodes"
  type        = number
  default     = 3
}

# EKS Addon Versions
variable "vpc_cni_addon_version" {
  description = "Version of the VPC CNI addon"
  type        = string
  default     = "v1.16.0-eksbuild.1"
}

variable "coredns_addon_version" {
  description = "Version of the CoreDNS addon"
  type        = string
  default     = "v1.11.1-eksbuild.4"
}

variable "kube_proxy_addon_version" {
  description = "Version of the kube-proxy addon"
  type        = string
  default     = "v1.29.0-eksbuild.1"
}

variable "ebs_csi_addon_version" {
  description = "Version of the EBS CSI driver addon"
  type        = string
  default     = "v1.28.0-eksbuild.1"
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
