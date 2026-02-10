# Outputs for VirtEngine Vault Module

output "vault_namespace" {
  description = "Kubernetes namespace where Vault is deployed"
  value       = kubernetes_namespace.vault.metadata[0].name
}

output "vault_iam_role_arn" {
  description = "ARN of the IAM role for Vault"
  value       = aws_iam_role.vault.arn
}

output "kms_key_arn" {
  description = "ARN of the KMS key for auto-unseal"
  value       = aws_kms_key.vault.arn
}

output "kms_key_id" {
  description = "ID of the KMS key for auto-unseal"
  value       = aws_kms_key.vault.key_id
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for HA storage"
  value       = aws_dynamodb_table.vault.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table for HA storage"
  value       = aws_dynamodb_table.vault.arn
}

output "vault_endpoint" {
  description = "Internal endpoint for Vault"
  value       = "http://vault.${var.vault_namespace}.svc.cluster.local:8200"
}

output "vault_ui_endpoint" {
  description = "Internal endpoint for Vault UI"
  value       = "http://vault-ui.${var.vault_namespace}.svc.cluster.local:8200"
}
