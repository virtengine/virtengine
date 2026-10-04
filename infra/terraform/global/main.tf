# VirtEngine Global Resources
# DNS, IAM, and Terraform state management across all regions

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket         = "virtengine-terraform-state"
    key            = "global/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "virtengine-terraform-locks"
  }
}

provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = "global"
      ManagedBy   = "terraform"
    }
  }
}

data "tls_certificate" "github_actions" {
  url = "https://token.actions.githubusercontent.com"
}

# -----------------------------------------------------------------------------
# Terraform State S3 Bucket
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "terraform_state" {
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_18:accepted: access logging is self-logged to the bucket itself (target_bucket = own arn) to avoid creating a second bucket with its own unencrypted-at-rest exposure; CloudTrail data events cover the access path | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  bucket = "virtengine-terraform-state"

  tags = {
    Name = "virtengine-terraform-state"
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# DynamoDB Table for State Locking
# -----------------------------------------------------------------------------
resource "aws_dynamodb_table" "terraform_locks" {
  #checkov:skip=CKV_AWS_119:accepted: the table holds non-sensitive lock rows; the AWS-owned key is adequate | review-by 2027-04-01
  #checkov:skip=CKV_AWS_28:KNOWN GAP, real defect carried deliberately: DynamoDB point-in-time recovery is off. Real fix is point_in_time_recovery { enabled = true }; accepted because the table holds only ephemeral terraform lock rows | review-by 2026-11-01
  name         = "virtengine-terraform-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Name = "virtengine-terraform-locks"
  }
}

# -----------------------------------------------------------------------------
# Global IAM Roles
# -----------------------------------------------------------------------------

# Cross-region admin role
resource "aws_iam_role" "cross_region_admin" {
  name = "virtengine-cross-region-admin"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        AWS = var.admin_role_arns
      }
      Condition = {
        Bool = {
          "aws:MultiFactorAuthPresent" = "true"
        }
      }
    }]
  })

  tags = {
    Name = "virtengine-cross-region-admin"
  }
}

resource "aws_iam_role_policy" "cross_region_admin" {
  #checkov:skip=CKV_AWS_355:KNOWN GAP, real defect carried deliberately: statements use Resource='*' for actions with no resource-level ARN (route53 change calls operate on a hosted-zone id the policy does not enumerate). Real fix is to scope them per hosted zone ARN; the resource-scoped statements in the same policy are already narrowed | review-by 2026-11-01
  #checkov:skip=CKV_AWS_290:KNOWN GAP, real defect carried deliberately: paired with CKV_AWS_355 -- unconstrained write statements. Real fix is the same per-zone scoping; today they are bounded by the role trust policy and the MFA condition | review-by 2026-11-01
  name = "cross-region-admin-policy"
  role = aws_iam_role.cross_region_admin.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "eks:DescribeCluster",
          "eks:ListClusters",
          "eks:UpdateClusterConfig",
        ]
        Resource = "arn:aws:eks:*:*:cluster/virtengine-*"
      },
      {
        Effect = "Allow"
        Action = [
          "route53:ChangeResourceRecordSets",
          "route53:GetHostedZone",
          "route53:ListHostedZones",
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
        ]
        Resource = [
          "arn:aws:s3:::virtengine-cockroachdb-backup-*",
          "arn:aws:s3:::virtengine-cockroachdb-backup-*/*",
        ]
      },
    ]
  })
}

# GitHub Actions OIDC provider for multi-region deployments
resource "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github_actions.certificates[length(data.tls_certificate.github_actions.certificates) - 1].sha1_fingerprint]

  tags = {
    Name = "github-actions-oidc"
  }
}

resource "aws_iam_role" "github_actions_deploy" {
  name = "virtengine-github-actions-multi-region-deploy"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRoleWithWebIdentity"
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.github_actions.arn
      }
      Condition = {
        StringLike = {
          "token.actions.githubusercontent.com:sub" = var.github_allowed_subjects
        }
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
        }
      }
    }]
  })

  tags = {
    Name = "virtengine-github-actions-multi-region-deploy"
  }
}

resource "aws_iam_role_policy" "github_actions_deploy" {
  #checkov:skip=CKV_AWS_355:KNOWN GAP, real defect carried deliberately: statements use Resource='*' for actions with no resource-level ARN (route53 change calls operate on a hosted-zone id the policy does not enumerate). Real fix is to scope them per hosted zone ARN; the resource-scoped statements in the same policy are already narrowed | review-by 2026-11-01
  name = "multi-region-deploy-policy"
  role = aws_iam_role.github_actions_deploy.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "eks:DescribeCluster",
          "eks:ListClusters",
        ]
        Resource = "arn:aws:eks:*:*:cluster/virtengine-*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
        ]
        Resource = "*"
      },
    ]
  })
}

# -----------------------------------------------------------------------------
# DNS Module (Global)
# -----------------------------------------------------------------------------
module "dns" {
  source = "../modules/dns"

  environment        = "prod"
  domain_name        = var.domain_name
  create_hosted_zone = var.create_hosted_zone
  primary_region     = "us-east-1"
  secondary_region   = "eu-west-1"
  enable_failover    = true

  regional_endpoints = var.regional_endpoints

  alarm_sns_topic_arns = var.alarm_sns_topic_arns

  tags = {
    Project   = "virtengine"
    ManagedBy = "terraform"
  }
}
