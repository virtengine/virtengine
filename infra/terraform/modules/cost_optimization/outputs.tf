# VirtEngine Cost Optimization Module - Outputs

output "cost_alerts_sns_topic_arn" {
  description = "ARN of the SNS topic for cost alerts"
  value       = aws_sns_topic.cost_alerts.arn
}

output "cost_alerts_sns_topic_name" {
  description = "Name of the SNS topic for cost alerts"
  value       = aws_sns_topic.cost_alerts.name
}

output "anomaly_monitor_arn" {
  description = "ARN of the cost anomaly monitor"
  value       = aws_ce_anomaly_monitor.project.arn
}

output "anomaly_subscription_arn" {
  description = "ARN of the cost anomaly subscription"
  value       = aws_ce_anomaly_subscription.main.arn
}

output "budget_names" {
  description = "Names of all created budgets"
  value = {
    monthly_total = aws_budgets_budget.monthly_total.name
    compute       = aws_budgets_budget.compute.name
    storage       = aws_budgets_budget.storage.name
    networking    = aws_budgets_budget.networking.name
    data_transfer = aws_budgets_budget.data_transfer.name
  }
}

output "cleanup_lambda_arn" {
  description = "ARN of the resource cleanup Lambda function"
  value       = var.enable_resource_cleanup ? aws_lambda_function.resource_cleanup[0].arn : null
}

output "cleanup_lambda_function_name" {
  description = "Name of the resource cleanup Lambda function"
  value       = var.enable_resource_cleanup ? aws_lambda_function.resource_cleanup[0].function_name : null
}

output "recommendations_lambda_arn" {
  description = "ARN of the cost recommendations Lambda function"
  value       = var.enable_cost_recommendations ? aws_lambda_function.cost_recommendations[0].arn : null
}

output "cost_dashboard_name" {
  description = "Name of the cost monitoring CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.cost.dashboard_name
}

output "cost_allocation_tags" {
  description = "List of enabled cost allocation tags"
  value = [
    aws_ce_cost_allocation_tag.project.tag_key,
    aws_ce_cost_allocation_tag.environment.tag_key,
    aws_ce_cost_allocation_tag.cost_center.tag_key,
    aws_ce_cost_allocation_tag.owner.tag_key,
    aws_ce_cost_allocation_tag.service.tag_key,
    aws_ce_cost_allocation_tag.team.tag_key,
  ]
}
