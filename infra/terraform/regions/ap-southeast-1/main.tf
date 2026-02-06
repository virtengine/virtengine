# VirtEngine AP-Southeast-1 Regional Configuration (Tertiary)

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }

  backend "s3" {
    bucket         = "virtengine-terraform-state"
    key            = "regions/ap-southeast-1/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "virtengine-terraform-locks"
  }
}

provider "aws" {
  region = "ap-southeast-1"

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = var.environment
      Region      = "ap-southeast-1"
      ManagedBy   = "terraform"
    }
  }
}

locals {
  region       = "ap-southeast-1"
  cluster_name = "virtengine-${var.environment}-${local.region}"
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
module "vpc" {
  source = "../../modules/vpc"

  name                     = "virtengine-${local.region}"
  environment              = var.environment
  vpc_cidr                 = var.vpc_cidr
  az_count                 = 3
  cluster_name             = local.cluster_name
  enable_nat_gateway       = true
  create_database_subnets  = true
  enable_flow_logs         = true
  flow_logs_retention_days = 90
  enable_vpc_endpoints     = true

  tags = var.tags
}

# -----------------------------------------------------------------------------
# EKS Cluster
# -----------------------------------------------------------------------------
module "eks" {
  source = "../../modules/eks"

  cluster_name           = local.cluster_name
  environment            = var.environment
  kubernetes_version     = var.kubernetes_version
  vpc_id                 = module.vpc.vpc_id
  subnet_ids             = module.vpc.private_subnet_ids
  enable_public_endpoint = false
  log_retention_days     = 90

  node_groups = {
    system = {
      instance_types = ["m5.xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 100
      desired_size   = 3
      max_size       = 6
      min_size       = 3
      labels = {
        role                   = "system"
        "virtengine.io/region" = local.region
      }
      taints = []
    }
    validators = {
      instance_types = ["m5.2xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 500
      desired_size   = var.validator_count
      max_size       = var.validator_count + 2
      min_size       = var.validator_count
      labels = {
        role                   = "validator"
        "virtengine.io/chain"  = "true"
        "virtengine.io/region" = local.region
      }
      taints = [{
        key    = "dedicated"
        value  = "validator"
        effect = "NO_SCHEDULE"
      }]
    }
    workload = {
      instance_types = ["m5.2xlarge", "m5a.2xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 200
      desired_size   = 3
      max_size       = 10
      min_size       = 2
      labels = {
        role                   = "workload"
        "virtengine.io/region" = local.region
      }
      taints = []
    }
    archive = {
      instance_types = ["r5.xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 1000
      desired_size   = 1
      max_size       = 2
      min_size       = 1
      labels = {
        role                   = "archive"
        "virtengine.io/chain"  = "true"
        "virtengine.io/region" = local.region
      }
      taints = [{
        key    = "dedicated"
        value  = "archive"
        effect = "NO_SCHEDULE"
      }]
    }
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# CockroachDB
# -----------------------------------------------------------------------------
module "database" {
  source = "../../modules/database"

  cluster_name      = local.cluster_name
  environment       = var.environment
  region            = local.region
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
  replicas          = 3
  join_addresses    = var.cockroachdb_join_addresses
  storage_size      = "500Gi"
  storage_class     = "gp3-encrypted"

  alarm_sns_topic_arns = var.alarm_sns_topic_arns

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Observability
# -----------------------------------------------------------------------------
module "observability" {
  source = "../../modules/observability"

  cluster_name           = local.cluster_name
  environment            = var.environment
  region                 = local.region
  is_primary_region      = false
  alert_sns_topic_arn    = var.alert_sns_topic_arn
  central_prometheus_url = var.central_prometheus_url

  tags = var.tags
}
