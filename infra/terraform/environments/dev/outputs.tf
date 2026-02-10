# Development Environment Outputs

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
}

output "kubeconfig_command" {
  description = "Command to update kubeconfig"
  value       = module.eks.kubeconfig_command
}

output "backup_bucket" {
  description = "Chain backups S3 bucket"
  value       = module.s3.chain_backups_bucket_id
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
}
