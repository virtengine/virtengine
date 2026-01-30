# Staging Environment - Main Configuration
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
  vpc_cidr           = "10.1.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  enable_nat_gateway = true
  single_nat_gateway = true  # Cost optimization
  enable_flow_logs   = true
  enable_bastion     = true
  
  # EKS
  cluster_name           = "virtengine-staging"
  kubernetes_version     = "1.29"
  endpoint_public_access = true
  public_access_cidrs    = ["0.0.0.0/0"]  # Restrict based on VPN/bastion
  
  # System nodes
  system_node_instance_types = ["t3.large"]
  system_node_desired_size   = 2
  system_node_max_size       = 4
  system_node_min_size       = 2
  
  # Application nodes
  app_node_instance_types = ["m5.xlarge"]
  app_node_capacity_type  = "ON_DEMAND"
  app_node_desired_size   = 3
  app_node_max_size       = 6
  app_node_min_size       = 2
  
  # Chain nodes
  chain_node_instance_types = ["m5.xlarge"]
  chain_node_disk_size      = 200
  chain_node_desired_size   = 3
  chain_node_max_size       = 5
  chain_node_min_size       = 2
  
  # RDS
  rds_instance_class        = "db.r5.large"
  rds_allocated_storage     = 100
  rds_max_allocated_storage = 500
  rds_multi_az              = true
  rds_deletion_protection   = true
  rds_skip_final_snapshot   = false
  rds_backup_retention      = 14
  
  # Vault
  vault_replicas         = 3
  enable_external_secrets = true
  
  # Monitoring
  log_retention_days       = 30
  prometheus_retention     = "15d"
  prometheus_storage_size  = "100Gi"
  enable_grafana_ingress   = true
  grafana_hostname         = "grafana.staging.virtengine.internal"
  
  # Alerts
  alert_emails = ["staging-alerts@virtengine.io"]
}
