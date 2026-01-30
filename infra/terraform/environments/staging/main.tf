# VirtEngine Staging Environment
# Terraform configuration for staging environment

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "virtengine-terraform-state-staging"
    key            = "staging/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "virtengine-terraform-locks-staging"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = "staging"
      ManagedBy   = "terraform"
      Repository  = "virtengine-network/virtengine"
    }
  }
}

locals {
  name_prefix  = "virtengine"
  environment  = "staging"
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
  az_count                = 3
  cluster_name            = local.cluster_name
  enable_nat_gateway      = true
  create_database_subnets = true
  enable_flow_logs        = true
  flow_logs_retention_days = 14
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
  log_retention_days     = 14
  enable_ssm_access      = true

  node_groups = {
    system = {
      instance_types = ["t3.large"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 50
      desired_size   = 2
      max_size       = 4
      min_size       = 2
      labels = {
        role = "system"
      }
      taints = []
    }
    workload = {
      instance_types = ["t3.xlarge", "t3a.xlarge"]
      capacity_type  = "SPOT"
      disk_size      = 100
      desired_size   = 2
      max_size       = 6
      min_size       = 1
      labels = {
        role = "workload"
      }
      taints = []
    }
    validators = {
      instance_types = ["m5.large"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 200
      desired_size   = 3
      max_size       = 5
      min_size       = 3
      labels = {
        role                  = "validator"
        "virtengine.io/chain" = "true"
      }
      taints = [{
        key    = "dedicated"
        value  = "validator"
        effect = "NO_SCHEDULE"
      }]
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
  create_ml_bucket    = true
  create_state_bucket = false  # Managed separately

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
  enable_github_actions_role = true
  github_org                 = "virtengine-network"
  github_repo                = "virtengine"

  tags = local.common_tags
}
