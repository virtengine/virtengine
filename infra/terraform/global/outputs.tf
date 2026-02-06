# Global Resources Outputs

output "terraform_state_bucket" {
  description = "Terraform state S3 bucket name"
  value       = aws_s3_bucket.terraform_state.id
}

output "terraform_locks_table" {
  description = "DynamoDB table for Terraform state locking"
  value       = aws_dynamodb_table.terraform_locks.name
}

output "cross_region_admin_role_arn" {
  description = "ARN of the cross-region admin IAM role"
  value       = aws_iam_role.cross_region_admin.arn
}

output "github_actions_deploy_role_arn" {
  description = "ARN of the GitHub Actions deployment IAM role"
  value       = aws_iam_role.github_actions_deploy.arn
}

output "dns_zone_id" {
  description = "Route53 hosted zone ID"
  value       = module.dns.zone_id
}

output "api_fqdn" {
  description = "Global API endpoint FQDN"
  value       = module.dns.api_fqdn
}

output "rpc_fqdn" {
  description = "Global RPC endpoint FQDN"
  value       = module.dns.rpc_fqdn
}

output "health_check_ids" {
  description = "Route53 health check IDs per region"
  value       = module.dns.health_check_ids
}
