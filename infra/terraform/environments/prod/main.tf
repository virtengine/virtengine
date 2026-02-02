# VirtEngine Production Environment
# Terraform configuration for production environment
# CRITICAL: Changes require approval via pull request

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "virtengine-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "virtengine-terraform-locks-prod"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = "prod"
      ManagedBy   = "terraform"
      Repository  = "virtengine-network/virtengine"
      CostCenter  = "infrastructure"
    }
  }
}

# Secondary region for DR
provider "aws" {
  alias  = "dr"
  region = var.dr_region

  default_tags {
    tags = {
      Project     = "virtengine"
      Environment = "prod-dr"
      ManagedBy   = "terraform"
      Repository  = "virtengine-network/virtengine"
      CostCenter  = "infrastructure"
    }
  }
}

locals {
  name_prefix  = "virtengine"
  environment  = "prod"
  cluster_name = "${local.name_prefix}-${local.environment}"

  common_tags = {
    Project     = "virtengine"
    Environment = local.environment
    ManagedBy   = "terraform"
    CostCenter  = "infrastructure"
  }
}

# -----------------------------------------------------------------------------
# VPC Module
# -----------------------------------------------------------------------------
module "vpc" {
  source = "../../modules/vpc"

  name                     = local.name_prefix
  environment              = local.environment
  vpc_cidr                 = var.vpc_cidr
  az_count                 = 3 # Multi-AZ for HA
  cluster_name             = local.cluster_name
  enable_nat_gateway       = true
  create_database_subnets  = true
  enable_flow_logs         = true
  flow_logs_retention_days = 90 # Extended retention for compliance
  enable_vpc_endpoints     = true

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
  enable_public_endpoint = false # Private-only for production
  log_retention_days     = 90
  enable_ssm_access      = false # Disabled for security

  enabled_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]

  node_groups = {
    system = {
      instance_types = ["m5.xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 100
      desired_size   = 3
      max_size       = 6
      min_size       = 3
      labels = {
        role = "system"
      }
      taints = []
    }
    workload = {
      instance_types = ["m5.2xlarge", "m5a.2xlarge"]
      capacity_type  = "ON_DEMAND" # On-demand for production stability
      disk_size      = 200
      desired_size   = 4
      max_size       = 12
      min_size       = 3
      labels = {
        role = "workload"
      }
      taints = []
    }
    validators = {
      instance_types = ["m5.2xlarge"]
      capacity_type  = "ON_DEMAND"
      disk_size      = 500 # Large disk for chain data
      desired_size   = 4
      max_size       = 7
      min_size       = 4
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
    inference = {
      instance_types = ["g4dn.xlarge"] # GPU for ML inference
      capacity_type  = "ON_DEMAND"
      disk_size      = 200
      desired_size   = 2
      max_size       = 6
      min_size       = 2
      labels = {
        role                        = "inference"
        "virtengine.io/accelerator" = "nvidia"
      }
      taints = [{
        key    = "dedicated"
        value  = "inference"
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
  create_state_bucket = false # Managed separately

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

# -----------------------------------------------------------------------------
# Additional Production Resources
# -----------------------------------------------------------------------------

# WAF for API Gateway protection
resource "aws_wafv2_web_acl" "api" {
  name        = "${local.name_prefix}-api-waf-${local.environment}"
  description = "WAF for VirtEngine API endpoints"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  # Rate limiting rule
  rule {
    name     = "RateLimitRule"
    priority = 1

    override_action {
      none {}
    }

    statement {
      rate_based_statement {
        limit              = 10000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
  }

  # AWS Managed Rules - Common Rule Set
  rule {
    name     = "AWSManagedRulesCommonRuleSet"
    priority = 2

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesCommonRuleSet"
      sampled_requests_enabled   = true
    }
  }

  # AWS Managed Rules - Known Bad Inputs
  rule {
    name     = "AWSManagedRulesKnownBadInputsRuleSet"
    priority = 3

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesKnownBadInputsRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesKnownBadInputsRuleSet"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${local.name_prefix}-api-waf"
    sampled_requests_enabled   = true
  }

  tags = local.common_tags
}

# CloudWatch Alarms for critical metrics
resource "aws_cloudwatch_metric_alarm" "cluster_cpu" {
  alarm_name          = "${local.cluster_name}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "node_cpu_utilization"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "EKS cluster CPU utilization is high"
  alarm_actions       = var.alarm_sns_topic_arns

  dimensions = {
    ClusterName = local.cluster_name
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_metric_alarm" "cluster_memory" {
  alarm_name          = "${local.cluster_name}-high-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "node_memory_utilization"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "EKS cluster memory utilization is high"
  alarm_actions       = var.alarm_sns_topic_arns

  dimensions = {
    ClusterName = local.cluster_name
  }

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# TEE Hardware Module (TEE-HW-001)
# Provisions SGX/SEV-SNP/Nitro enclave nodes for production
# -----------------------------------------------------------------------------
module "tee_hardware" {
  source = "../../modules/tee-hardware"

  cluster_name   = local.cluster_name
  aws_region     = var.aws_region
  vpc_id         = module.vpc.vpc_id
  subnet_ids     = module.vpc.private_subnet_ids
  node_role_arn  = module.eks.node_role_arn
  node_role_name = module.eks.node_role_name
  kms_key_arn    = module.eks.kms_key_arn

  # Platform enablement
  enable_nitro   = var.enable_tee_nitro
  enable_sev_snp = var.enable_tee_sev_snp
  enable_sgx     = var.enable_tee_sgx

  # Nitro configuration
  nitro_desired_size         = var.tee_nitro_desired_size
  nitro_min_size             = 2
  nitro_max_size             = 6
  nitro_enclave_memory_mb    = 2048
  nitro_enclave_cpu_count    = 2
  nitro_enclave_image_sha384 = var.nitro_enclave_image_sha384

  # SEV-SNP configuration (when enabled)
  sev_snp_desired_size = 2
  sev_snp_min_size     = 2
  sev_snp_max_size     = 6

  # SGX configuration (when enabled)
  sgx_desired_size  = 2
  sgx_min_size      = 2
  sgx_max_size      = 6
  sgx_pccs_endpoint = "https://pccs.virtengine.io/sgx/certification/v4/"

  # Attestation configuration
  measurement_allowlist = var.tee_measurement_allowlist
  min_tcb_version       = "2.0.8.115"

  # Alerting
  alarm_sns_topic_arns = var.alarm_sns_topic_arns

  tags = local.common_tags
}
