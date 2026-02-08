# Multi-Region Deployment Module
# This module deploys VirtEngine infrastructure across multiple AWS regions
# with automatic failover, cross-region replication, and DR capabilities.

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
      configuration_aliases = [
        aws.primary,
        aws.secondary,
        aws.tertiary,
      ]
    }
  }
}

# -----------------------------------------------------------------------------
# Input Variables
# -----------------------------------------------------------------------------

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "virtengine"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "primary_region" {
  description = "Primary AWS region"
  type        = string
  default     = "us-east-1"
}

variable "secondary_region" {
  description = "Secondary AWS region for failover"
  type        = string
  default     = "eu-west-1"
}

variable "tertiary_region" {
  description = "Tertiary AWS region for geographic redundancy"
  type        = string
  default     = "ap-southeast-1"
}

variable "validator_count_primary" {
  description = "Number of validators in primary region"
  type        = number
  default     = 3
}

variable "validator_count_secondary" {
  description = "Number of validators in secondary region"
  type        = number
  default     = 2
}

variable "validator_count_tertiary" {
  description = "Number of validators in tertiary region"
  type        = number
  default     = 2
}

variable "enable_cross_region_replication" {
  description = "Enable S3 cross-region replication for backups"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 90
}

variable "rto_target_seconds" {
  description = "Recovery Time Objective in seconds"
  type        = number
  default     = 900 # 15 minutes
}

variable "rpo_target_seconds" {
  description = "Recovery Point Objective in seconds"
  type        = number
  default     = 300 # 5 minutes
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Local Variables
# -----------------------------------------------------------------------------

locals {
  common_tags = merge(
    var.tags,
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
      DR          = "enabled"
    }
  )

  regions = {
    primary   = var.primary_region
    secondary = var.secondary_region
    tertiary  = var.tertiary_region
  }

  validator_distribution = {
    primary   = var.validator_count_primary
    secondary = var.validator_count_secondary
    tertiary  = var.validator_count_tertiary
  }
}

# -----------------------------------------------------------------------------
# Backup Buckets with Cross-Region Replication
# -----------------------------------------------------------------------------

# Primary region backup bucket
resource "aws_s3_bucket" "backup_primary" {
  provider = aws.primary
  bucket   = "${var.project_name}-dr-backups-${var.primary_region}"

  tags = merge(local.common_tags, {
    Region = var.primary_region
    Role   = "primary"
  })
}

resource "aws_s3_bucket_versioning" "backup_primary" {
  provider = aws.primary
  bucket   = aws_s3_bucket.backup_primary.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backup_primary" {
  provider = aws.primary
  bucket   = aws_s3_bucket.backup_primary.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.backup_primary.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "backup_primary" {
  provider = aws.primary
  bucket   = aws_s3_bucket.backup_primary.id

  rule {
    id     = "archive-old-backups"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = var.backup_retention_days
      storage_class = "GLACIER"
    }

    expiration {
      days = var.backup_retention_days + 365
    }
  }
}

# Secondary region backup bucket
resource "aws_s3_bucket" "backup_secondary" {
  provider = aws.secondary
  bucket   = "${var.project_name}-dr-backups-${var.secondary_region}"

  tags = merge(local.common_tags, {
    Region = var.secondary_region
    Role   = "secondary"
  })
}

resource "aws_s3_bucket_versioning" "backup_secondary" {
  provider = aws.secondary
  bucket   = aws_s3_bucket.backup_secondary.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backup_secondary" {
  provider = aws.secondary
  bucket   = aws_s3_bucket.backup_secondary.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.backup_secondary.arn
    }
    bucket_key_enabled = true
  }
}

# Cross-region replication (primary -> secondary)
resource "aws_s3_bucket_replication_configuration" "backup_replication" {
  count    = var.enable_cross_region_replication ? 1 : 0
  provider = aws.primary

  bucket = aws_s3_bucket.backup_primary.id
  role   = aws_iam_role.replication[0].arn

  rule {
    id     = "replicate-all"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.backup_secondary.arn
      storage_class = "STANDARD_IA"

      encryption_configuration {
        replica_kms_key_id = aws_kms_key.backup_secondary.arn
      }

      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }

      metrics {
        status = "Enabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
}

# IAM role for S3 replication
resource "aws_iam_role" "replication" {
  count    = var.enable_cross_region_replication ? 1 : 0
  provider = aws.primary
  name     = "${var.project_name}-s3-replication-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "s3.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy" "replication" {
  count    = var.enable_cross_region_replication ? 1 : 0
  provider = aws.primary
  role     = aws_iam_role.replication[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetReplicationConfiguration",
          "s3:ListBucket"
        ]
        Resource = aws_s3_bucket.backup_primary.arn
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObjectVersionForReplication",
          "s3:GetObjectVersionAcl",
          "s3:GetObjectVersionTagging"
        ]
        Resource = "${aws_s3_bucket.backup_primary.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ReplicateObject",
          "s3:ReplicateDelete",
          "s3:ReplicateTags"
        ]
        Resource = "${aws_s3_bucket.backup_secondary.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt"
        ]
        Resource = aws_kms_key.backup_primary.arn
        Condition = {
          StringLike = {
            "kms:ViaService" = "s3.${var.primary_region}.amazonaws.com"
          }
        }
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Encrypt"
        ]
        Resource = aws_kms_key.backup_secondary.arn
        Condition = {
          StringLike = {
            "kms:ViaService" = "s3.${var.secondary_region}.amazonaws.com"
          }
        }
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# KMS Keys for Encryption
# -----------------------------------------------------------------------------

resource "aws_kms_key" "backup_primary" {
  provider                = aws.primary
  description             = "KMS key for backup encryption in ${var.primary_region}"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  multi_region            = false

  tags = merge(local.common_tags, {
    Region = var.primary_region
  })
}

resource "aws_kms_alias" "backup_primary" {
  provider      = aws.primary
  name          = "alias/${var.project_name}-dr-backups-${var.primary_region}"
  target_key_id = aws_kms_key.backup_primary.key_id
}

resource "aws_kms_key" "backup_secondary" {
  provider                = aws.secondary
  description             = "KMS key for backup encryption in ${var.secondary_region}"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  multi_region            = false

  tags = merge(local.common_tags, {
    Region = var.secondary_region
  })
}

resource "aws_kms_alias" "backup_secondary" {
  provider      = aws.secondary
  name          = "alias/${var.project_name}-dr-backups-${var.secondary_region}"
  target_key_id = aws_kms_key.backup_secondary.key_id
}

# -----------------------------------------------------------------------------
# Route53 Health Checks and Failover DNS
# -----------------------------------------------------------------------------

resource "aws_route53_health_check" "primary" {
  provider          = aws.primary
  fqdn              = "rpc-${var.primary_region}.${var.project_name}.io"
  port              = 443
  type              = "HTTPS"
  resource_path     = "/status"
  failure_threshold = 3
  request_interval  = 30
  measure_latency   = true

  tags = merge(local.common_tags, {
    Region = var.primary_region
    Role   = "primary"
  })
}

resource "aws_route53_health_check" "secondary" {
  provider          = aws.secondary
  fqdn              = "rpc-${var.secondary_region}.${var.project_name}.io"
  port              = 443
  type              = "HTTPS"
  resource_path     = "/status"
  failure_threshold = 3
  request_interval  = 30
  measure_latency   = true

  tags = merge(local.common_tags, {
    Region = var.secondary_region
    Role   = "secondary"
  })
}

# CloudWatch alarms for health checks
resource "aws_cloudwatch_metric_alarm" "primary_health" {
  provider            = aws.primary
  alarm_name          = "${var.project_name}-dr-primary-region-health"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 2
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = 60
  statistic           = "Minimum"
  threshold           = 1
  alarm_description   = "Primary region health check failing"
  alarm_actions       = [aws_sns_topic.dr_alerts_primary.arn]

  dimensions = {
    HealthCheckId = aws_route53_health_check.primary.id
  }

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# SNS Topics for DR Notifications
# -----------------------------------------------------------------------------

resource "aws_sns_topic" "dr_alerts_primary" {
  provider = aws.primary
  name     = "${var.project_name}-dr-alerts-${var.primary_region}"

  tags = local.common_tags
}

resource "aws_sns_topic" "dr_alerts_secondary" {
  provider = aws.secondary
  name     = "${var.project_name}-dr-alerts-${var.secondary_region}"

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "backup_buckets" {
  description = "Backup S3 bucket names by region"
  value = {
    primary   = aws_s3_bucket.backup_primary.id
    secondary = aws_s3_bucket.backup_secondary.id
  }
}

output "kms_key_arns" {
  description = "KMS key ARNs for backup encryption"
  value = {
    primary   = aws_kms_key.backup_primary.arn
    secondary = aws_kms_key.backup_secondary.arn
  }
}

output "health_check_ids" {
  description = "Route53 health check IDs"
  value = {
    primary   = aws_route53_health_check.primary.id
    secondary = aws_route53_health_check.secondary.id
  }
}

output "sns_topic_arns" {
  description = "SNS topic ARNs for DR alerts"
  value = {
    primary   = aws_sns_topic.dr_alerts_primary.arn
    secondary = aws_sns_topic.dr_alerts_secondary.arn
  }
}

output "rto_target" {
  description = "Recovery Time Objective (seconds)"
  value       = var.rto_target_seconds
}

output "rpo_target" {
  description = "Recovery Point Objective (seconds)"
  value       = var.rpo_target_seconds
}

output "replication_status" {
  description = "Cross-region replication configuration status"
  value       = var.enable_cross_region_replication ? "enabled" : "disabled"
}
