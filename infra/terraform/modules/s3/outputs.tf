# S3 Module Outputs

output "chain_backups_bucket_id" {
  description = "ID of the chain backups bucket"
  value       = aws_s3_bucket.chain_backups.id
}

output "chain_backups_bucket_arn" {
  description = "ARN of the chain backups bucket"
  value       = aws_s3_bucket.chain_backups.arn
}

output "manifests_bucket_id" {
  description = "ID of the manifests bucket"
  value       = aws_s3_bucket.manifests.id
}

output "manifests_bucket_arn" {
  description = "ARN of the manifests bucket"
  value       = aws_s3_bucket.manifests.arn
}

output "ml_models_bucket_id" {
  description = "ID of the ML models bucket"
  value       = var.create_ml_bucket ? aws_s3_bucket.ml_models[0].id : null
}

output "ml_models_bucket_arn" {
  description = "ARN of the ML models bucket"
  value       = var.create_ml_bucket ? aws_s3_bucket.ml_models[0].arn : null
}

output "terraform_state_bucket_id" {
  description = "ID of the Terraform state bucket"
  value       = var.create_state_bucket ? aws_s3_bucket.terraform_state[0].id : null
}

output "terraform_state_bucket_arn" {
  description = "ARN of the Terraform state bucket"
  value       = var.create_state_bucket ? aws_s3_bucket.terraform_state[0].arn : null
}

output "terraform_locks_table_name" {
  description = "Name of the DynamoDB table for Terraform state locking"
  value       = var.create_state_bucket ? aws_dynamodb_table.terraform_locks[0].name : null
}

output "kms_key_arn" {
  description = "ARN of the KMS key for S3 encryption"
  value       = aws_kms_key.s3.arn
}

output "kms_key_id" {
  description = "ID of the KMS key for S3 encryption"
  value       = aws_kms_key.s3.key_id
}
