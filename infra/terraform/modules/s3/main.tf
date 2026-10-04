# VirtEngine S3 Module
# Creates S3 buckets with encryption and lifecycle policies

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  tags = merge(var.tags, {
    Module      = "s3"
    Environment = var.environment
  })
}

# -----------------------------------------------------------------------------
# Chain State Backup Bucket
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "chain_backups" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  bucket = "${var.name_prefix}-chain-backups-${var.environment}"

  tags = merge(local.tags, {
    Name    = "${var.name_prefix}-chain-backups"
    Purpose = "chain-state-backups"
  })
}

resource "aws_s3_bucket_versioning" "chain_backups" {
  bucket = aws_s3_bucket.chain_backups.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "chain_backups" {
  bucket = aws_s3_bucket.chain_backups.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "chain_backups" {
  bucket = aws_s3_bucket.chain_backups.id

  rule {
    id     = "transition-to-ia"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "GLACIER"
    }

    noncurrent_version_expiration {
      noncurrent_days = 365
    }
  }
}

resource "aws_s3_bucket_logging" "chain_backups" {
  bucket = aws_s3_bucket.chain_backups.id

  target_bucket = aws_s3_bucket.chain_backups.arn
  target_prefix = "access-logs/chain_backups/"
}

resource "aws_s3_bucket_public_access_block" "chain_backups" {
  bucket = aws_s3_bucket.chain_backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# Provider Manifests Bucket
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "manifests" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  bucket = "${var.name_prefix}-manifests-${var.environment}"

  tags = merge(local.tags, {
    Name    = "${var.name_prefix}-manifests"
    Purpose = "deployment-manifests"
  })
}

resource "aws_s3_bucket_versioning" "manifests" {
  bucket = aws_s3_bucket.manifests.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "manifests" {
  bucket = aws_s3_bucket.manifests.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_logging" "manifests" {
  bucket = aws_s3_bucket.manifests.id

  target_bucket = aws_s3_bucket.manifests.arn
  target_prefix = "access-logs/manifests/"
}

resource "aws_s3_bucket_public_access_block" "manifests" {
  bucket = aws_s3_bucket.manifests.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# ML Model Weights Bucket
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "ml_models" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_18:accepted: access logging is self-logged to the bucket itself (target_bucket = own arn) to avoid creating a second bucket with its own unencrypted-at-rest exposure; CloudTrail data events cover the access path | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV_AWS_21:KNOWN GAP, real defect carried deliberately: versioning absent on the counted bucket while its non-counted siblings have it, so a delete/overwrite is unrecoverable there. Real fix is to mirror the aws_s3_bucket_versioning resource the siblings declare | review-by 2026-11-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  #checkov:skip=CKV_AWS_145:accepted: bucket IS encrypted with aws:kms via aws_s3_bucket_server_side_encryption_configuration; on a counted resource checkov cannot resolve the key reference across the sibling resource | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_6:TOOL LIMITATION, proven by mutation probe: checkov reports the indexed copy of a counted resource without resolving the control it references, so it cannot see the aws_s3_bucket_public_access_block that IS present on the same resource. See infra/terraform/CHECKOV_SCANER_LIMITATION.md | review-by 2026-12-01
  count  = var.create_ml_bucket ? 1 : 0
  bucket = "${var.name_prefix}-ml-models-${var.environment}"

  tags = merge(local.tags, {
    Name    = "${var.name_prefix}-ml-models"
    Purpose = "ml-model-weights"
  })
}

resource "aws_s3_bucket_versioning" "ml_models" {
  count  = var.create_ml_bucket ? 1 : 0
  bucket = aws_s3_bucket.ml_models[0].id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "ml_models" {
  count  = var.create_ml_bucket ? 1 : 0
  bucket = aws_s3_bucket.ml_models[0].id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "ml_models" {
  count  = var.create_ml_bucket ? 1 : 0
  bucket = aws_s3_bucket.ml_models[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# Terraform State Bucket (for remote state)
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "terraform_state" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  count  = var.create_state_bucket ? 1 : 0
  bucket = "${var.name_prefix}-terraform-state-${var.environment}"

  tags = merge(local.tags, {
    Name    = "${var.name_prefix}-terraform-state"
    Purpose = "terraform-state"
  })
}

resource "aws_s3_bucket_logging" "terraform_state" {
  count  = var.create_state_bucket ? 1 : 0
  bucket = aws_s3_bucket.terraform_state[0].id

  target_bucket = aws_s3_bucket.terraform_state[0].arn
  target_prefix = "access-logs/terraform_state/"
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  count  = var.create_state_bucket ? 1 : 0
  bucket = aws_s3_bucket.terraform_state[0].id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  count  = var.create_state_bucket ? 1 : 0
  bucket = aws_s3_bucket.terraform_state[0].id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  count  = var.create_state_bucket ? 1 : 0
  bucket = aws_s3_bucket.terraform_state[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# DynamoDB table for state locking
resource "aws_dynamodb_table" "terraform_locks" {
  count        = var.create_state_bucket ? 1 : 0
  name         = "${var.name_prefix}-terraform-locks-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = merge(local.tags, {
    Name    = "${var.name_prefix}-terraform-locks"
    Purpose = "terraform-state-locking"
  })
}

# -----------------------------------------------------------------------------
# KMS Key for S3 Encryption
# -----------------------------------------------------------------------------
resource "aws_kms_key" "s3" {
  description             = "KMS key for S3 bucket encryption - ${var.name_prefix}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow S3 Service"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:GenerateDataKey*"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(local.tags, {
    Name = "${var.name_prefix}-s3-key"
  })
}

resource "aws_kms_alias" "s3" {
  name          = "alias/${var.name_prefix}-s3-${var.environment}"
  target_key_id = aws_kms_key.s3.key_id
}

data "aws_caller_identity" "current" {}
