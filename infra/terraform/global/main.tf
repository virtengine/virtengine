# VirtEngine Global Resources
# DNS, IAM, and Terraform state management across all regions

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
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

# -----------------------------------------------------------------------------
# Terraform State S3 Bucket
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "terraform_state" {
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
  thumbprint_list = ["ffffffffffffffffffffffffffffffffffffffff"]

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
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:*"
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
