# EKS Module Outputs

output "cluster_id" {
  description = "ID of the EKS cluster"
  value       = aws_eks_cluster.main.id
}

output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = aws_eks_cluster.main.name
}

output "cluster_arn" {
  description = "ARN of the EKS cluster"
  value       = aws_eks_cluster.main.arn
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS cluster API server"
  value       = aws_eks_cluster.main.endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for the cluster"
  value       = aws_eks_cluster.main.certificate_authority[0].data
}

output "cluster_version" {
  description = "Kubernetes version of the cluster"
  value       = aws_eks_cluster.main.version
}

output "cluster_security_group_id" {
  description = "Security group ID for the cluster"
  value       = aws_security_group.cluster.id
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = aws_iam_role.cluster.arn
}

output "node_iam_role_arn" {
  description = "IAM role ARN for the EKS node groups"
  value       = aws_iam_role.node.arn
}

output "oidc_provider_arn" {
  description = "ARN of the OIDC provider for IRSA"
  value       = aws_iam_openid_connect_provider.eks.arn
}

output "oidc_provider_url" {
  description = "URL of the OIDC provider for IRSA"
  value       = aws_iam_openid_connect_provider.eks.url
}

output "kms_key_arn" {
  description = "ARN of the KMS key used for secrets encryption"
  value       = aws_kms_key.eks.arn
}

output "node_groups" {
  description = "Map of node group names to their resources"
  value = {
    for k, v in aws_eks_node_group.main : k => {
      arn    = v.arn
      status = v.status
    }
  }
}

# Kubeconfig helper output
output "kubeconfig_command" {
  description = "AWS CLI command to update kubeconfig"
  value       = "aws eks update-kubeconfig --region ${data.aws_region.current.name} --name ${aws_eks_cluster.main.name}"
}

data "aws_region" "current" {}
