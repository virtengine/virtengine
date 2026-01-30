# IAM Module Outputs

output "external_secrets_role_arn" {
  description = "ARN of the External Secrets Operator IAM role"
  value       = aws_iam_role.external_secrets.arn
}

output "load_balancer_controller_role_arn" {
  description = "ARN of the AWS Load Balancer Controller IAM role"
  value       = aws_iam_role.load_balancer_controller.arn
}

output "cluster_autoscaler_role_arn" {
  description = "ARN of the Cluster Autoscaler IAM role"
  value       = aws_iam_role.cluster_autoscaler.arn
}

output "virtengine_node_role_arn" {
  description = "ARN of the VirtEngine node service account IAM role"
  value       = aws_iam_role.virtengine_node.arn
}

output "provider_daemon_role_arn" {
  description = "ARN of the Provider Daemon service account IAM role"
  value       = aws_iam_role.provider_daemon.arn
}

output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions IAM role"
  value       = var.enable_github_actions_role ? aws_iam_role.github_actions[0].arn : null
}
