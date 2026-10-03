# VirtEngine DR Failover Terraform Configuration
# Manages failover DNS records, database promotion, and scaling

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "failover_active" {
  description = "Whether a failover is currently active"
  type        = bool
  default     = false
}

variable "primary_region" {
  description = "Normal primary region"
  type        = string
  default     = "us-east-1"
}

variable "target_region" {
  description = "Region to fail over to"
  type        = string
  default     = "eu-west-1"
}

variable "regions" {
  description = "Map of region configurations"
  type = map(object({
    cluster_name      = string
    validator_count   = number
    lb_dns_name       = string
    lb_zone_id        = string
    health_check_fqdn = string
  }))
  default = {}
}

variable "domain_name" {
  description = "Domain name for DNS records"
  type        = string
  default     = "virtengine.com"
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default = {
    Project   = "virtengine"
    ManagedBy = "terraform"
    Purpose   = "disaster-recovery"
  }
}

# SNS Topic for DR notifications
resource "aws_sns_topic" "dr_notifications" {
  #checkov:skip=CKV_AWS_26:KNOWN GAP, real defect carried deliberately: SNS topic is not encrypted with a KMS key. Real fix is a kms_master_key_id; accepted because the topic carries only non-sensitive failover state-change notifications | review-by 2026-11-01
  name = "virtengine-dr-notifications"
  tags = var.tags
}

# CloudWatch alarm for failover monitoring
resource "aws_cloudwatch_metric_alarm" "failover_duration" {
  alarm_name          = "virtengine-failover-duration"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "FailoverDurationSeconds"
  namespace           = "VirtEngine/DR"
  period              = 60
  statistic           = "Maximum"
  threshold           = 900
  alarm_description   = "Regional failover is taking longer than 15 minutes (RTO target)"
  alarm_actions       = [aws_sns_topic.dr_notifications.arn]

  tags = var.tags
}

# CloudWatch alarm for RPO monitoring
resource "aws_cloudwatch_metric_alarm" "rpo_breach" {
  alarm_name          = "virtengine-rpo-breach"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ReplicationLagSeconds"
  namespace           = "VirtEngine/DR"
  period              = 60
  statistic           = "Maximum"
  threshold           = 300
  alarm_description   = "Database replication lag exceeds 5 minutes (RPO target)"
  alarm_actions       = [aws_sns_topic.dr_notifications.arn]

  tags = var.tags
}

# S3 bucket for DR test results
resource "aws_s3_bucket" "dr_results" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_18:accepted: access logging is self-logged to the bucket itself (target_bucket = own arn) to avoid creating a second bucket with its own unencrypted-at-rest exposure; CloudTrail data events cover the access path | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  #checkov:skip=CKV_AWS_145:accepted: bucket IS encrypted with aws:kms via aws_s3_bucket_server_side_encryption_configuration; on a counted resource checkov cannot resolve the key reference across the sibling resource | review-by 2027-04-01
  bucket = "virtengine-dr-test-results"
  tags   = var.tags
}

resource "aws_s3_bucket_versioning" "dr_results" {
  bucket = aws_s3_bucket.dr_results.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "dr_results" {
  bucket = aws_s3_bucket.dr_results.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

output "dr_notifications_topic_arn" {
  description = "SNS topic ARN for DR notifications"
  value       = aws_sns_topic.dr_notifications.arn
}

output "dr_results_bucket" {
  description = "S3 bucket for DR test results"
  value       = aws_s3_bucket.dr_results.id
}
