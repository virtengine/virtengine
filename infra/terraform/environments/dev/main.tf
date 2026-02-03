# VirtEngine Development Environment
# Terraform configuration for dev environment

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Uncomment and configure for remote state
  # backend "s3" {
  #   bucket         = "virtengine-terraform-state-dev"
  #   key            = "dev/terraform.tfstate"
  #   region         = "us-west-2"
  #   encrypt        = true
  #   dynamodb_table = "virtengine-terraform-locks-dev"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = "dev"
      ManagedBy   = "terraform"
      Repository  = "virtengine-network/virtengine"
    }
  }
}

locals {
  name_prefix  = "virtengine"
  environment  = "dev"
  cluster_name = "${local.name_prefix}-${local.environment}"

  common_tags = {
    Project     = "virtengine"
    Environment = local.environment
    ManagedBy   = "terraform"
  }
}

# -----------------------------------------------------------------------------
# VPC Module
# -----------------------------------------------------------------------------
module "vpc" {
  source = "../../modules/vpc"

  name                    = local.name_prefix
  environment             = local.environment
  vpc_cidr                = var.vpc_cidr
  az_count                = 2 # Reduced for dev cost savings
  cluster_name            = local.cluster_name
  enable_nat_gateway      = true
  create_database_subnets = false # Not needed for dev
  enable_flow_logs        = false # Disabled for dev cost savings
  enable_vpc_endpoints    = true

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# EKS Module
# -----------------------------------------------------------------------------
module "eks" {
  source = "../../modules/eks"

  cluster_name           = local.cluster_name
  environment            = local.environment
  kubernetes_version     = var.kubernetes_version
  vpc_id                 = module.vpc.vpc_id
  subnet_ids             = module.vpc.private_subnet_ids
  enable_public_endpoint = true
  public_access_cidrs    = var.allowed_cidr_blocks
  log_retention_days     = 7    # Short retention for dev
  enable_ssm_access      = true # Enable for debugging

  node_groups = {
    system = {
      instance_types = ["t3.medium"]
      capacity_type  = "SPOT" # Use spot for dev cost savings
      disk_size      = 30
      desired_size   = 2
      max_size       = 4
      min_size       = 1
      labels = {
        role = "system"
      }
      taints = []
    }
    workload = {
      instance_types = ["t3.large", "t3a.large"]
      capacity_type  = "SPOT"
      disk_size      = 50
      desired_size   = 1
      max_size       = 3
      min_size       = 0
      labels = {
        role = "workload"
      }
      taints = []
    }
  }

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# S3 Module
# -----------------------------------------------------------------------------
module "s3" {
  source = "../../modules/s3"

  name_prefix         = local.name_prefix
  environment         = local.environment
  create_ml_bucket    = false # Not needed for dev
  create_state_bucket = false # Use local state for dev

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# IAM Module
# -----------------------------------------------------------------------------
module "iam" {
  source = "../../modules/iam"

  name_prefix                = local.name_prefix
  environment                = local.environment
  cluster_name               = local.cluster_name
  oidc_provider_arn          = module.eks.oidc_provider_arn
  oidc_provider_url          = module.eks.oidc_provider_url
  kms_key_arns               = [module.s3.kms_key_arn, module.eks.kms_key_arn]
  backup_bucket_arn          = module.s3.chain_backups_bucket_arn
  manifests_bucket_arn       = module.s3.manifests_bucket_arn
  enable_github_actions_role = false # Not needed for dev

  tags = local.common_tags
}
