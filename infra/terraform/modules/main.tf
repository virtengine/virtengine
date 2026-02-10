# VirtEngine Production Infrastructure
# Root module that composes all infrastructure components

terraform {
  required_version = ">= 1.5.0"
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
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Networking Module
# -----------------------------------------------------------------------------
module "networking" {
  source = "./networking"

  project                  = var.project
  environment              = var.environment
  cluster_name             = var.cluster_name
  vpc_cidr                 = var.vpc_cidr
  availability_zones       = var.availability_zones
  enable_nat_gateway       = var.enable_nat_gateway
  single_nat_gateway       = var.single_nat_gateway
  enable_flow_logs         = var.enable_flow_logs
  flow_logs_retention_days = var.flow_logs_retention_days
  enable_bastion           = var.enable_bastion
  bastion_allowed_cidrs    = var.bastion_allowed_cidrs
  tags                     = var.tags
}

# -----------------------------------------------------------------------------
# EKS Module
# -----------------------------------------------------------------------------
module "eks" {
  source = "./eks"

  cluster_name              = var.cluster_name
  kubernetes_version        = var.kubernetes_version
  private_subnet_ids        = module.networking.private_subnet_ids
  public_subnet_ids         = module.networking.public_subnet_ids
  cluster_security_group_id = module.networking.eks_cluster_security_group_id

  endpoint_private_access = var.endpoint_private_access
  endpoint_public_access  = var.endpoint_public_access
  public_access_cidrs     = var.public_access_cidrs

  enabled_cluster_log_types  = var.enabled_cluster_log_types
  cluster_log_retention_days = var.cluster_log_retention_days

  # System nodes
  system_node_instance_types = var.system_node_instance_types
  system_node_disk_size      = var.system_node_disk_size
  system_node_desired_size   = var.system_node_desired_size
  system_node_max_size       = var.system_node_max_size
  system_node_min_size       = var.system_node_min_size

  # Application nodes
  app_node_instance_types = var.app_node_instance_types
  app_node_capacity_type  = var.app_node_capacity_type
  app_node_disk_size      = var.app_node_disk_size
  app_node_desired_size   = var.app_node_desired_size
  app_node_max_size       = var.app_node_max_size
  app_node_min_size       = var.app_node_min_size

  # Chain nodes
  chain_node_instance_types = var.chain_node_instance_types
  chain_node_disk_size      = var.chain_node_disk_size
  chain_node_desired_size   = var.chain_node_desired_size
  chain_node_max_size       = var.chain_node_max_size
  chain_node_min_size       = var.chain_node_min_size

  tags = var.tags

  depends_on = [module.networking]
}

# -----------------------------------------------------------------------------
# RDS Module
# -----------------------------------------------------------------------------
module "rds" {
  source = "./rds"

  project              = var.project
  environment          = var.environment
  db_subnet_group_name = module.networking.database_subnet_group_name
  security_group_id    = module.networking.database_security_group_id

  engine_version        = var.rds_engine_version
  instance_class        = var.rds_instance_class
  database_name         = var.rds_database_name
  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_type          = var.rds_storage_type
  iops                  = var.rds_iops
  multi_az              = var.rds_multi_az

  backup_retention_period = var.rds_backup_retention
  deletion_protection     = var.rds_deletion_protection
  skip_final_snapshot     = var.rds_skip_final_snapshot

  monitoring_interval                   = var.rds_monitoring_interval
  performance_insights_enabled          = var.rds_performance_insights
  performance_insights_retention_period = var.rds_performance_insights_retention

  create_read_replica    = var.create_read_replica
  replica_instance_class = var.replica_instance_class

  alarm_actions = var.enable_monitoring ? [module.monitoring[0].sns_topic_arn] : []

  tags = var.tags

  depends_on = [module.networking]
}

# -----------------------------------------------------------------------------
# Vault Module
# -----------------------------------------------------------------------------
module "vault" {
  count  = var.enable_vault ? 1 : 0
  source = "./vault"

  project           = var.project
  environment       = var.environment
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.cluster_oidc_issuer_url

  vault_namespace      = var.vault_namespace
  vault_replicas       = var.vault_replicas
  vault_memory_request = var.vault_memory_request
  vault_memory_limit   = var.vault_memory_limit
  vault_cpu_request    = var.vault_cpu_request
  vault_cpu_limit      = var.vault_cpu_limit

  enable_injector         = var.enable_vault_injector
  enable_csi              = var.enable_vault_csi
  enable_external_secrets = var.enable_external_secrets

  tags = var.tags

  depends_on = [module.eks]
}

# -----------------------------------------------------------------------------
# Monitoring Module
# -----------------------------------------------------------------------------
module "monitoring" {
  count  = var.enable_monitoring ? 1 : 0
  source = "./monitoring"

  project         = var.project
  environment     = var.environment
  cluster_name    = module.eks.cluster_name
  rds_instance_id = module.rds.db_instance_id
  nat_gateway_ids = module.networking.nat_gateway_ids

  alert_emails       = var.alert_emails
  log_retention_days = var.log_retention_days

  prometheus_retention      = var.prometheus_retention
  prometheus_retention_size = var.prometheus_retention_size
  prometheus_storage_size   = var.prometheus_storage_size

  grafana_admin_password = var.grafana_admin_password
  enable_grafana_ingress = var.enable_grafana_ingress
  grafana_hostname       = var.grafana_hostname

  tags = var.tags

  depends_on = [module.eks, module.rds]
}
