# AP-Southeast-1 Region Outputs

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

output "database_endpoint" {
  description = "CockroachDB service endpoint"
  value       = module.database.service_endpoint
}

output "prometheus_endpoint" {
  description = "Prometheus endpoint"
  value       = module.observability.prometheus_endpoint
}

output "kubeconfig_command" {
  description = "Command to update kubeconfig"
  value       = module.eks.kubeconfig_command
}
