# Development Environment - Main Configuration
# VirtEngine Infrastructure

include "root" {
  path = find_in_parent_folders("terragrunt.hcl")
}

locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
}

terraform {
  source = "../../modules//."
}

inputs = {
  # Networking
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b"]
  enable_nat_gateway = true
  single_nat_gateway = true  # Cost optimization for dev
  enable_flow_logs   = false # Disabled for dev
  enable_bastion     = true
  
  # EKS
  cluster_name           = "virtengine-dev"
  kubernetes_version     = "1.29"
  endpoint_public_access = true
  public_access_cidrs    = ["0.0.0.0/0"]  # Restrict in production
  
  # System nodes
  system_node_instance_types = ["t3.medium"]
  system_node_desired_size   = 2
  system_node_max_size       = 3
  system_node_min_size       = 1
  
  # Application nodes
  app_node_instance_types = ["t3.large"]
  app_node_capacity_type  = "SPOT"  # Use spot for dev
  app_node_desired_size   = 2
  app_node_max_size       = 5
  app_node_min_size       = 1
  
  # Chain nodes
  chain_node_instance_types = ["t3.xlarge"]
  chain_node_disk_size      = 100
  chain_node_desired_size   = 1
  chain_node_max_size       = 2
  chain_node_min_size       = 1
  
  # RDS
  rds_instance_class        = "db.t3.medium"
  rds_allocated_storage     = 50
  rds_max_allocated_storage = 100
  rds_multi_az              = false
  rds_deletion_protection   = false
  rds_skip_final_snapshot   = true
  rds_backup_retention      = 7
  
  # Vault
  vault_replicas         = 1
  enable_external_secrets = true
  
  # Monitoring
  log_retention_days       = 7
  prometheus_retention     = "7d"
  prometheus_storage_size  = "50Gi"
  enable_grafana_ingress   = false
  
  # Alerts
  alert_emails = []
}
