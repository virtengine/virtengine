# Production Environment Outputs

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
  sensitive   = true
}

output "kubeconfig_command" {
  description = "Command to update kubeconfig"
  value       = module.eks.kubeconfig_command
}

output "backup_bucket" {
  description = "Chain backups S3 bucket"
  value       = module.s3.chain_backups_bucket_id
}

output "ml_models_bucket" {
  description = "ML models S3 bucket"
  value       = module.s3.ml_models_bucket_id
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
}

output "github_actions_role_arn" {
  description = "GitHub Actions IAM role ARN"
  value       = module.iam.github_actions_role_arn
}

output "waf_acl_arn" {
  description = "WAF ACL ARN for API protection"
  value       = aws_wafv2_web_acl.api.arn
}

output "external_secrets_role_arn" {
  description = "External Secrets Operator IAM role ARN"
  value       = module.iam.external_secrets_role_arn
}

output "load_balancer_controller_role_arn" {
  description = "AWS Load Balancer Controller IAM role ARN"
  value       = module.iam.load_balancer_controller_role_arn
}

# =============================================================================
# TEE Hardware Outputs (TEE-HW-001)
# =============================================================================

output "tee_enabled_platforms" {
  description = "List of enabled TEE platforms"
  value       = module.tee_hardware.enabled_platforms
}

output "tee_nitro_node_group_arn" {
  description = "ARN of the Nitro Enclave node group"
  value       = module.tee_hardware.nitro_node_group_arn
}

output "tee_attestation_security_group_id" {
  description = "Security group ID for TEE attestation traffic"
  value       = module.tee_hardware.tee_attestation_security_group_id
}

output "tee_config_secret_arn" {
  description = "ARN of the TEE configuration secret"
  value       = module.tee_hardware.tee_config_secret_arn
}

output "tee_node_labels" {
  description = "Kubernetes labels for TEE node selection"
  value       = module.tee_hardware.node_labels
}
