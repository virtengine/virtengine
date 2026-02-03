# VirtEngine Auto-Scaling Optimization Module - Outputs

output "app_node_cpu_policy_arn" {
  description = "ARN of the application node CPU scaling policy"
  value       = aws_autoscaling_policy.app_node_cpu.arn
}

output "app_node_predictive_policy_arn" {
  description = "ARN of the application node predictive scaling policy"
  value       = var.enable_predictive_scaling ? aws_autoscaling_policy.app_node_predictive[0].arn : null
}

output "scheduled_scaling_enabled" {
  description = "Whether scheduled scaling is enabled for this environment"
  value       = local.scaling_config.enable_scheduled_scaling
}

output "scaling_schedule" {
  description = "Scheduled scaling configuration"
  value = local.scaling_config.enable_scheduled_scaling ? {
    scale_down_time  = local.scaling_config.scale_down_time
    scale_up_time    = local.scaling_config.scale_up_time
    min_off_hours    = local.scaling_config.min_capacity_off_hours
    weekend_shutdown = var.enable_weekend_shutdown
  } : null
}

output "cluster_autoscaler_config" {
  description = "Cluster Autoscaler configuration values"
  value = var.deploy_cluster_autoscaler_config ? {
    scale_down_utilization_threshold = var.environment == "prod" ? "0.5" : "0.6"
    scale_down_unneeded_time         = var.environment == "prod" ? "10m" : "5m"
    balance_similar_node_groups      = true
    expander                         = "priority"
  } : null
}

output "ecs_scaling_targets" {
  description = "ECS scaling target information"
  value = var.enable_ecs_scaling ? {
    provider_daemon = {
      resource_id  = aws_appautoscaling_target.provider_daemon[0].resource_id
      min_capacity = aws_appautoscaling_target.provider_daemon[0].min_capacity
      max_capacity = aws_appautoscaling_target.provider_daemon[0].max_capacity
    }
  } : null
}

output "high_cpu_alarm_arn" {
  description = "ARN of the high CPU alarm"
  value       = aws_cloudwatch_metric_alarm.app_node_high_cpu.arn
}

output "scale_out_events_alarm_arn" {
  description = "ARN of the scale out events alarm"
  value       = aws_cloudwatch_metric_alarm.scale_out_events.arn
}
