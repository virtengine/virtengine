# Root Module Outputs for VirtEngine Infrastructure

# -----------------------------------------------------------------------------
# Networking Outputs
# -----------------------------------------------------------------------------
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.networking.vpc_id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = module.networking.vpc_cidr_block
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = module.networking.private_subnet_ids
}

output "database_subnet_ids" {
  description = "List of database subnet IDs"
  value       = module.networking.database_subnet_ids
}

output "nat_gateway_public_ips" {
  description = "List of NAT Gateway public IPs"
  value       = module.networking.nat_gateway_public_ips
}

# -----------------------------------------------------------------------------
# EKS Outputs
# -----------------------------------------------------------------------------
output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS cluster API server"
  value       = module.eks.cluster_endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for the cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "cluster_oidc_issuer_url" {
  description = "OIDC issuer URL for the cluster"
  value       = module.eks.cluster_oidc_issuer_url
}

output "oidc_provider_arn" {
  description = "ARN of the OIDC provider for IRSA"
  value       = module.eks.oidc_provider_arn
}

output "kubeconfig_command" {
  description = "Command to update kubeconfig"
  value       = module.eks.kubeconfig_command
}

# -----------------------------------------------------------------------------
# RDS Outputs
# -----------------------------------------------------------------------------
output "db_instance_endpoint" {
  description = "Connection endpoint for the RDS instance"
  value       = module.rds.db_instance_endpoint
}

output "db_instance_address" {
  description = "Hostname of the RDS instance"
  value       = module.rds.db_instance_address
}

output "db_credentials_secret_arn" {
  description = "ARN of the Secrets Manager secret containing DB credentials"
  value       = module.rds.db_credentials_secret_arn
}

output "db_replica_endpoint" {
  description = "Connection endpoint for the read replica"
  value       = module.rds.db_replica_endpoint
}

# -----------------------------------------------------------------------------
# Vault Outputs
# -----------------------------------------------------------------------------
output "vault_endpoint" {
  description = "Internal endpoint for Vault"
  value       = var.enable_vault ? module.vault[0].vault_endpoint : null
}

output "vault_iam_role_arn" {
  description = "ARN of the IAM role for Vault"
  value       = var.enable_vault ? module.vault[0].vault_iam_role_arn : null
}

# -----------------------------------------------------------------------------
# Monitoring Outputs
# -----------------------------------------------------------------------------
output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = var.enable_monitoring ? module.monitoring[0].sns_topic_arn : null
}

output "cloudwatch_dashboard_url" {
  description = "URL for the CloudWatch dashboard"
  value       = var.enable_monitoring ? module.monitoring[0].dashboard_url : null
}

output "prometheus_endpoint" {
  description = "Internal endpoint for Prometheus"
  value       = var.enable_monitoring ? module.monitoring[0].prometheus_endpoint : null
}

output "grafana_endpoint" {
  description = "Internal endpoint for Grafana"
  value       = var.enable_monitoring ? module.monitoring[0].grafana_endpoint : null
}
