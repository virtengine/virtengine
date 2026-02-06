# CockroachDB Database Module Outputs

output "namespace" {
  description = "Kubernetes namespace for CockroachDB"
  value       = kubernetes_namespace.cockroachdb.metadata[0].name
}

output "backup_bucket_arn" {
  description = "ARN of the S3 backup bucket"
  value       = aws_s3_bucket.backups.arn
}

output "backup_bucket_id" {
  description = "ID of the S3 backup bucket"
  value       = aws_s3_bucket.backups.id
}

output "backup_role_arn" {
  description = "IAM role ARN for CockroachDB backups"
  value       = aws_iam_role.cockroachdb_backup.arn
}

output "kms_key_arn" {
  description = "KMS key ARN for database encryption"
  value       = aws_kms_key.cockroachdb.arn
}

output "service_endpoint" {
  description = "Internal service endpoint for CockroachDB"
  value       = "cockroachdb-public.${local.namespace}.svc.cluster.local:26257"
}

output "sql_endpoint" {
  description = "SQL endpoint for CockroachDB"
  value       = "cockroachdb-public.${local.namespace}.svc.cluster.local:26257"
}

output "admin_endpoint" {
  description = "Admin UI endpoint for CockroachDB"
  value       = "cockroachdb-public.${local.namespace}.svc.cluster.local:8080"
}
