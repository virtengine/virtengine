# VirtEngine Auto-Scaling Optimization Module
#
# This module implements optimized auto-scaling policies for cost optimization:
# - Tuned EKS node group scaling policies
# - Scheduled scaling for dev/staging environments
# - Predictive scaling based on usage patterns
# - Scale-to-zero for non-production during off-hours

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.25"
    }
  }
}

# -----------------------------------------------------------------------------
# Local Values
# -----------------------------------------------------------------------------

locals {
  common_tags = merge(var.tags, {
    Project     = "virtengine"
    Environment = var.environment
    ManagedBy   = "terraform"
    Module      = "autoscaling-optimization"
  })

  # Environment-specific scaling configurations
  env_scaling_configs = {
    dev = {
      enable_scheduled_scaling = true
      scale_down_time          = "0 20 * * MON-FRI"  # 8 PM weekdays
      scale_up_time            = "0 8 * * MON-FRI"   # 8 AM weekdays
      weekend_scale_down       = "0 20 * * FRI"      # Friday 8 PM
      weekend_scale_up         = "0 8 * * MON"       # Monday 8 AM
      min_capacity_off_hours   = 0
      target_cpu_utilization   = 70
      scale_in_cooldown        = 300
      scale_out_cooldown       = 60
    }
    staging = {
      enable_scheduled_scaling = true
      scale_down_time          = "0 22 * * *"        # 10 PM daily
      scale_up_time            = "0 6 * * *"         # 6 AM daily
      weekend_scale_down       = "0 22 * * FRI"
      weekend_scale_up         = "0 6 * * MON"
      min_capacity_off_hours   = 1
      target_cpu_utilization   = 65
      scale_in_cooldown        = 300
      scale_out_cooldown       = 120
    }
    prod = {
      enable_scheduled_scaling = false
      scale_down_time          = ""
      scale_up_time            = ""
      weekend_scale_down       = ""
      weekend_scale_up         = ""
      min_capacity_off_hours   = 2
      target_cpu_utilization   = 60
      scale_in_cooldown        = 600
      scale_out_cooldown       = 180
    }
  }

  scaling_config = local.env_scaling_configs[var.environment]
}

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# -----------------------------------------------------------------------------
# EKS Node Group Scaling Policies
# -----------------------------------------------------------------------------

# Target Tracking Scaling Policy for Application Nodes (CPU)
resource "aws_autoscaling_policy" "app_node_cpu" {
  name                   = "virtengine-${var.environment}-app-cpu-target"
  autoscaling_group_name = var.app_node_asg_name
  policy_type            = "TargetTrackingScaling"

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value     = local.scaling_config.target_cpu_utilization
    disable_scale_in = false
  }
}

# Target Tracking Scaling Policy for Application Nodes (Memory)
resource "aws_autoscaling_policy" "app_node_memory" {
  count = var.enable_memory_scaling ? 1 : 0

  name                   = "virtengine-${var.environment}-app-memory-target"
  autoscaling_group_name = var.app_node_asg_name
  policy_type            = "TargetTrackingScaling"

  target_tracking_configuration {
    customized_metric_specification {
      metric_dimension {
        name  = "AutoScalingGroupName"
        value = var.app_node_asg_name
      }
      metric_name = "MemoryUtilization"
      namespace   = "CWAgent"
      statistic   = "Average"
    }
    target_value = 75
  }
}

# Predictive Scaling Policy for Application Nodes
resource "aws_autoscaling_policy" "app_node_predictive" {
  count = var.enable_predictive_scaling ? 1 : 0

  name                   = "virtengine-${var.environment}-app-predictive"
  autoscaling_group_name = var.app_node_asg_name
  policy_type            = "PredictiveScaling"

  predictive_scaling_configuration {
    metric_specification {
      target_value = local.scaling_config.target_cpu_utilization

      predefined_load_metric_specification {
        predefined_metric_type = "ASGTotalCPUUtilization"
      }

      predefined_scaling_metric_specification {
        predefined_metric_type = "ASGAverageCPUUtilization"
      }
    }

    mode                         = "ForecastAndScale"
    scheduling_buffer_time       = 300
    max_capacity_breach_behavior = "IncreaseMaxCapacity"
    max_capacity_buffer          = 10
  }
}

# Step Scaling Policy for rapid scale-out on high load
resource "aws_autoscaling_policy" "app_node_step_scale_out" {
  name                   = "virtengine-${var.environment}-app-step-scale-out"
  autoscaling_group_name = var.app_node_asg_name
  policy_type            = "StepScaling"
  adjustment_type        = "ChangeInCapacity"

  step_adjustment {
    scaling_adjustment          = 1
    metric_interval_lower_bound = 0
    metric_interval_upper_bound = 10
  }

  step_adjustment {
    scaling_adjustment          = 2
    metric_interval_lower_bound = 10
    metric_interval_upper_bound = 20
  }

  step_adjustment {
    scaling_adjustment          = 4
    metric_interval_lower_bound = 20
  }
}

# CloudWatch Alarm for step scaling
resource "aws_cloudwatch_metric_alarm" "app_node_high_cpu" {
  alarm_name          = "virtengine-${var.environment}-app-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Triggers step scaling when CPU exceeds 80%"

  dimensions = {
    AutoScalingGroupName = var.app_node_asg_name
  }

  alarm_actions = [aws_autoscaling_policy.app_node_step_scale_out.arn]

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Scheduled Scaling for Non-Production Environments
# -----------------------------------------------------------------------------

# Scale down during off-hours (weekdays)
resource "aws_autoscaling_schedule" "app_scale_down_evening" {
  count = local.scaling_config.enable_scheduled_scaling ? 1 : 0

  scheduled_action_name  = "scale-down-evening"
  autoscaling_group_name = var.app_node_asg_name
  recurrence             = local.scaling_config.scale_down_time
  min_size               = local.scaling_config.min_capacity_off_hours
  max_size               = var.app_node_max_size
  desired_capacity       = local.scaling_config.min_capacity_off_hours
  time_zone              = var.timezone
}

# Scale up in morning (weekdays)
resource "aws_autoscaling_schedule" "app_scale_up_morning" {
  count = local.scaling_config.enable_scheduled_scaling ? 1 : 0

  scheduled_action_name  = "scale-up-morning"
  autoscaling_group_name = var.app_node_asg_name
  recurrence             = local.scaling_config.scale_up_time
  min_size               = var.app_node_min_size
  max_size               = var.app_node_max_size
  desired_capacity       = var.app_node_desired_size
  time_zone              = var.timezone
}

# Weekend scale down (Friday evening)
resource "aws_autoscaling_schedule" "app_scale_down_weekend" {
  count = local.scaling_config.enable_scheduled_scaling && var.enable_weekend_shutdown ? 1 : 0

  scheduled_action_name  = "scale-down-weekend"
  autoscaling_group_name = var.app_node_asg_name
  recurrence             = local.scaling_config.weekend_scale_down
  min_size               = 0
  max_size               = var.app_node_max_size
  desired_capacity       = 0
  time_zone              = var.timezone
}

# Weekend scale up (Monday morning)
resource "aws_autoscaling_schedule" "app_scale_up_weekend" {
  count = local.scaling_config.enable_scheduled_scaling && var.enable_weekend_shutdown ? 1 : 0

  scheduled_action_name  = "scale-up-weekend"
  autoscaling_group_name = var.app_node_asg_name
  recurrence             = local.scaling_config.weekend_scale_up
  min_size               = var.app_node_min_size
  max_size               = var.app_node_max_size
  desired_capacity       = var.app_node_desired_size
  time_zone              = var.timezone
}

# System node scheduled scaling
resource "aws_autoscaling_schedule" "system_scale_down_evening" {
  count = local.scaling_config.enable_scheduled_scaling && var.environment == "dev" ? 1 : 0

  scheduled_action_name  = "system-scale-down-evening"
  autoscaling_group_name = var.system_node_asg_name
  recurrence             = local.scaling_config.scale_down_time
  min_size               = 1
  max_size               = var.system_node_max_size
  desired_capacity       = 1
  time_zone              = var.timezone
}

resource "aws_autoscaling_schedule" "system_scale_up_morning" {
  count = local.scaling_config.enable_scheduled_scaling && var.environment == "dev" ? 1 : 0

  scheduled_action_name  = "system-scale-up-morning"
  autoscaling_group_name = var.system_node_asg_name
  recurrence             = local.scaling_config.scale_up_time
  min_size               = var.system_node_min_size
  max_size               = var.system_node_max_size
  desired_capacity       = var.system_node_desired_size
  time_zone              = var.timezone
}

# -----------------------------------------------------------------------------
# Cluster Autoscaler Configuration (Kubernetes)
# -----------------------------------------------------------------------------

resource "kubernetes_config_map" "cluster_autoscaler" {
  count = var.deploy_cluster_autoscaler_config ? 1 : 0

  metadata {
    name      = "cluster-autoscaler-priority-expander"
    namespace = "kube-system"
  }

  data = {
    "priorities" = yamlencode({
      # Prioritize Spot instances over On-Demand for cost optimization
      priorities = [
        {
          id    = 10
          regex = ".*spot.*"
        },
        {
          id    = 50
          regex = ".*"
        }
      ]
    })
  }
}

# Cluster Autoscaler ConfigMap for optimization
resource "kubernetes_config_map" "cluster_autoscaler_config" {
  count = var.deploy_cluster_autoscaler_config ? 1 : 0

  metadata {
    name      = "cluster-autoscaler-config"
    namespace = "kube-system"
  }

  data = {
    # Scale-down configurations for cost optimization
    "scale-down-enabled"              = "true"
    "scale-down-delay-after-add"      = var.environment == "prod" ? "10m" : "5m"
    "scale-down-delay-after-delete"   = "0s"
    "scale-down-delay-after-failure"  = "3m"
    "scale-down-unneeded-time"        = var.environment == "prod" ? "10m" : "5m"
    "scale-down-unready-time"         = "20m"
    "scale-down-utilization-threshold" = var.environment == "prod" ? "0.5" : "0.6"
    
    # Balance similar node groups for even distribution
    "balance-similar-node-groups"     = "true"
    
    # Skip nodes with local storage for safer scaling
    "skip-nodes-with-local-storage"   = "false"
    "skip-nodes-with-system-pods"     = "true"
    
    # Expander strategy
    "expander"                        = "priority"
    
    # Node readiness and resource estimation
    "max-node-provision-time"         = "15m"
    "max-graceful-termination-sec"    = "600"
    
    # New pod handling
    "new-pod-scale-up-delay"          = "0s"
  }
}

# -----------------------------------------------------------------------------
# Application Auto-scaling for ECS/Fargate (if used)
# -----------------------------------------------------------------------------

resource "aws_appautoscaling_target" "provider_daemon" {
  count = var.enable_ecs_scaling ? 1 : 0

  max_capacity       = var.provider_daemon_max_count
  min_capacity       = var.provider_daemon_min_count
  resource_id        = "service/${var.ecs_cluster_name}/${var.provider_daemon_service_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "provider_daemon_cpu" {
  count = var.enable_ecs_scaling ? 1 : 0

  name               = "virtengine-${var.environment}-provider-cpu"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.provider_daemon[0].resource_id
  scalable_dimension = aws_appautoscaling_target.provider_daemon[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.provider_daemon[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70
    scale_in_cooldown  = local.scaling_config.scale_in_cooldown
    scale_out_cooldown = local.scaling_config.scale_out_cooldown
  }
}

resource "aws_appautoscaling_policy" "provider_daemon_memory" {
  count = var.enable_ecs_scaling ? 1 : 0

  name               = "virtengine-${var.environment}-provider-memory"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.provider_daemon[0].resource_id
  scalable_dimension = aws_appautoscaling_target.provider_daemon[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.provider_daemon[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 80
    scale_in_cooldown  = local.scaling_config.scale_in_cooldown
    scale_out_cooldown = local.scaling_config.scale_out_cooldown
  }
}

# Scheduled scaling for ECS services in non-prod
resource "aws_appautoscaling_scheduled_action" "provider_daemon_scale_down" {
  count = var.enable_ecs_scaling && local.scaling_config.enable_scheduled_scaling ? 1 : 0

  name               = "scale-down-evening"
  service_namespace  = aws_appautoscaling_target.provider_daemon[0].service_namespace
  resource_id        = aws_appautoscaling_target.provider_daemon[0].resource_id
  scalable_dimension = aws_appautoscaling_target.provider_daemon[0].scalable_dimension
  schedule           = "cron(${local.scaling_config.scale_down_time})"
  timezone           = var.timezone

  scalable_target_action {
    min_capacity = local.scaling_config.min_capacity_off_hours
    max_capacity = var.provider_daemon_max_count
  }
}

resource "aws_appautoscaling_scheduled_action" "provider_daemon_scale_up" {
  count = var.enable_ecs_scaling && local.scaling_config.enable_scheduled_scaling ? 1 : 0

  name               = "scale-up-morning"
  service_namespace  = aws_appautoscaling_target.provider_daemon[0].service_namespace
  resource_id        = aws_appautoscaling_target.provider_daemon[0].resource_id
  scalable_dimension = aws_appautoscaling_target.provider_daemon[0].scalable_dimension
  schedule           = "cron(${local.scaling_config.scale_up_time})"
  timezone           = var.timezone

  scalable_target_action {
    min_capacity = var.provider_daemon_min_count
    max_capacity = var.provider_daemon_max_count
  }
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms for Scaling Events
# -----------------------------------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "scale_out_events" {
  alarm_name          = "virtengine-${var.environment}-scaling-events"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "GroupDesiredCapacity"
  namespace           = "AWS/AutoScaling"
  period              = 300
  statistic           = "Average"
  threshold           = var.app_node_max_size * 0.8
  alarm_description   = "Alert when ASG reaches 80% of max capacity"
  
  dimensions = {
    AutoScalingGroupName = var.app_node_asg_name
  }

  alarm_actions = var.scaling_alarm_actions

  tags = local.common_tags
}
