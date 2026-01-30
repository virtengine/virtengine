# Root Module Variables for VirtEngine Infrastructure

# -----------------------------------------------------------------------------
# General Configuration
# -----------------------------------------------------------------------------
variable "project" {
  description = "Project name for resource naming"
  type        = string
  default     = "virtengine"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod"
  }
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Networking Configuration
# -----------------------------------------------------------------------------
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single NAT Gateway (cost optimization for non-prod)"
  type        = bool
  default     = false
}

variable "enable_flow_logs" {
  description = "Enable VPC Flow Logs"
  type        = bool
  default     = true
}

variable "flow_logs_retention_days" {
  description = "Number of days to retain VPC flow logs"
  type        = number
  default     = 30
}

variable "enable_bastion" {
  description = "Enable bastion host security group"
  type        = bool
  default     = false
}

variable "bastion_allowed_cidrs" {
  description = "List of CIDRs allowed to SSH to bastion"
  type        = list(string)
  default     = []
}

# -----------------------------------------------------------------------------
# EKS Configuration
# -----------------------------------------------------------------------------
variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version for the EKS cluster"
  type        = string
  default     = "1.29"
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

# System Node Group
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

# Application Node Group
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

# Chain Node Group
variable "chain_node_instance_types" {
  description = "Instance types for chain node group"
  type        = list(string)
  default     = ["m5.2xlarge"]
}

variable "chain_node_disk_size" {
  description = "Disk size in GB for chain nodes"
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

# -----------------------------------------------------------------------------
# RDS Configuration
# -----------------------------------------------------------------------------
variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.5"
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "rds_database_name" {
  description = "Name of the database to create"
  type        = string
  default     = "virtengine"
}

variable "rds_allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 100
}

variable "rds_max_allocated_storage" {
  description = "Maximum allocated storage in GB for autoscaling"
  type        = number
  default     = 500
}

variable "rds_storage_type" {
  description = "Storage type (gp3, io1)"
  type        = string
  default     = "gp3"
}

variable "rds_iops" {
  description = "IOPS for io1 storage type"
  type        = number
  default     = 3000
}

variable "rds_multi_az" {
  description = "Enable Multi-AZ deployment"
  type        = bool
  default     = true
}

variable "rds_backup_retention" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "rds_deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "rds_skip_final_snapshot" {
  description = "Skip final snapshot on deletion"
  type        = bool
  default     = false
}

variable "rds_monitoring_interval" {
  description = "Enhanced monitoring interval (0 to disable)"
  type        = number
  default     = 60
}

variable "rds_performance_insights" {
  description = "Enable Performance Insights"
  type        = bool
  default     = true
}

variable "rds_performance_insights_retention" {
  description = "Performance Insights retention period in days"
  type        = number
  default     = 7
}

variable "create_read_replica" {
  description = "Create a read replica"
  type        = bool
  default     = false
}

variable "replica_instance_class" {
  description = "Instance class for read replica"
  type        = string
  default     = ""
}

# -----------------------------------------------------------------------------
# Vault Configuration
# -----------------------------------------------------------------------------
variable "enable_vault" {
  description = "Enable Vault deployment"
  type        = bool
  default     = true
}

variable "vault_namespace" {
  description = "Kubernetes namespace for Vault"
  type        = string
  default     = "vault"
}

variable "vault_replicas" {
  description = "Number of Vault replicas"
  type        = number
  default     = 3
}

variable "vault_memory_request" {
  description = "Memory request for Vault pods"
  type        = string
  default     = "256Mi"
}

variable "vault_memory_limit" {
  description = "Memory limit for Vault pods"
  type        = string
  default     = "512Mi"
}

variable "vault_cpu_request" {
  description = "CPU request for Vault pods"
  type        = string
  default     = "250m"
}

variable "vault_cpu_limit" {
  description = "CPU limit for Vault pods"
  type        = string
  default     = "500m"
}

variable "enable_vault_injector" {
  description = "Enable Vault Agent Injector"
  type        = bool
  default     = true
}

variable "enable_vault_csi" {
  description = "Enable Vault CSI Provider"
  type        = bool
  default     = false
}

variable "enable_external_secrets" {
  description = "Enable External Secrets Operator"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# Monitoring Configuration
# -----------------------------------------------------------------------------
variable "enable_monitoring" {
  description = "Enable monitoring stack"
  type        = bool
  default     = true
}

variable "alert_emails" {
  description = "List of email addresses for alert notifications"
  type        = list(string)
  default     = []
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
}

variable "prometheus_retention" {
  description = "Prometheus data retention period"
  type        = string
  default     = "15d"
}

variable "prometheus_retention_size" {
  description = "Maximum size of Prometheus data"
  type        = string
  default     = "50GB"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "100Gi"
}

variable "grafana_admin_password" {
  description = "Admin password for Grafana"
  type        = string
  sensitive   = true
  default     = ""
}

variable "enable_grafana_ingress" {
  description = "Enable ingress for Grafana"
  type        = bool
  default     = false
}

variable "grafana_hostname" {
  description = "Hostname for Grafana ingress"
  type        = string
  default     = "grafana.internal"
}
