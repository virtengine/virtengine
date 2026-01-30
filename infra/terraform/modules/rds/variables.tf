# Variables for VirtEngine RDS Module

variable "project" {
  description = "Project name for resource naming"
  type        = string
  default     = "virtengine"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "db_subnet_group_name" {
  description = "Name of the DB subnet group"
  type        = string
}

variable "security_group_id" {
  description = "Security group ID for the RDS instance"
  type        = string
}

variable "engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.5"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "database_name" {
  description = "Name of the database to create"
  type        = string
  default     = "virtengine"
}

variable "master_username" {
  description = "Master username for the database"
  type        = string
  default     = "virtengine_admin"
}

variable "allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 100
}

variable "max_allocated_storage" {
  description = "Maximum allocated storage in GB for autoscaling"
  type        = number
  default     = 500
}

variable "storage_type" {
  description = "Storage type (gp3, io1)"
  type        = string
  default     = "gp3"
}

variable "iops" {
  description = "IOPS for io1 storage type"
  type        = number
  default     = 3000
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment"
  type        = bool
  default     = true
}

variable "backup_retention_period" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "backup_window" {
  description = "Preferred backup window"
  type        = string
  default     = "03:00-04:00"
}

variable "maintenance_window" {
  description = "Preferred maintenance window"
  type        = string
  default     = "sun:04:00-sun:05:00"
}

variable "auto_minor_version_upgrade" {
  description = "Enable automatic minor version upgrades"
  type        = bool
  default     = true
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "skip_final_snapshot" {
  description = "Skip final snapshot on deletion"
  type        = bool
  default     = false
}

variable "monitoring_interval" {
  description = "Enhanced monitoring interval (0 to disable)"
  type        = number
  default     = 60
}

variable "performance_insights_enabled" {
  description = "Enable Performance Insights"
  type        = bool
  default     = true
}

variable "performance_insights_retention_period" {
  description = "Performance Insights retention period in days"
  type        = number
  default     = 7
}

variable "log_min_duration_statement" {
  description = "Log queries longer than this (ms)"
  type        = string
  default     = "1000"
}

variable "additional_parameters" {
  description = "Additional DB parameters"
  type = list(object({
    name         = string
    value        = string
    apply_method = optional(string)
  }))
  default = []
}

variable "kms_key_arn" {
  description = "ARN of KMS key for encryption (if not provided, a new key is created)"
  type        = string
  default     = ""
}

variable "alarm_actions" {
  description = "List of ARNs for alarm actions (SNS topics)"
  type        = list(string)
  default     = []
}

variable "max_connections_threshold" {
  description = "Threshold for connections alarm"
  type        = number
  default     = 100
}

variable "create_read_replica" {
  description = "Create a read replica"
  type        = bool
  default     = false
}

variable "replica_instance_class" {
  description = "Instance class for read replica (defaults to same as primary)"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
