# VirtEngine Auto-Scaling Optimization Module - Variables

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

variable "timezone" {
  description = "Timezone for scheduled scaling actions"
  type        = string
  default     = "UTC"
}

# -----------------------------------------------------------------------------
# ASG Configuration
# -----------------------------------------------------------------------------

variable "app_node_asg_name" {
  description = "Name of the application node Auto Scaling Group"
  type        = string
}

variable "system_node_asg_name" {
  description = "Name of the system node Auto Scaling Group"
  type        = string
}

variable "app_node_min_size" {
  description = "Minimum size of the application node ASG"
  type        = number
  default     = 2
}

variable "app_node_max_size" {
  description = "Maximum size of the application node ASG"
  type        = number
  default     = 10
}

variable "app_node_desired_size" {
  description = "Desired size of the application node ASG"
  type        = number
  default     = 3
}

variable "system_node_min_size" {
  description = "Minimum size of the system node ASG"
  type        = number
  default     = 2
}

variable "system_node_max_size" {
  description = "Maximum size of the system node ASG"
  type        = number
  default     = 4
}

variable "system_node_desired_size" {
  description = "Desired size of the system node ASG"
  type        = number
  default     = 2
}

# -----------------------------------------------------------------------------
# Scaling Features
# -----------------------------------------------------------------------------

variable "enable_memory_scaling" {
  description = "Enable memory-based scaling (requires CloudWatch Agent)"
  type        = bool
  default     = false
}

variable "enable_predictive_scaling" {
  description = "Enable predictive scaling based on historical patterns"
  type        = bool
  default     = false
}

variable "enable_weekend_shutdown" {
  description = "Enable complete shutdown on weekends for dev environments"
  type        = bool
  default     = false
}

variable "deploy_cluster_autoscaler_config" {
  description = "Deploy Cluster Autoscaler ConfigMaps"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# ECS Configuration (if using ECS for provider daemon)
# -----------------------------------------------------------------------------

variable "enable_ecs_scaling" {
  description = "Enable ECS service auto-scaling"
  type        = bool
  default     = false
}

variable "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
  default     = ""
}

variable "provider_daemon_service_name" {
  description = "Name of the provider daemon ECS service"
  type        = string
  default     = ""
}

variable "provider_daemon_min_count" {
  description = "Minimum count of provider daemon tasks"
  type        = number
  default     = 2
}

variable "provider_daemon_max_count" {
  description = "Maximum count of provider daemon tasks"
  type        = number
  default     = 10
}

# -----------------------------------------------------------------------------
# Alarm Configuration
# -----------------------------------------------------------------------------

variable "scaling_alarm_actions" {
  description = "List of ARNs to notify when scaling alarms trigger"
  type        = list(string)
  default     = []
}
