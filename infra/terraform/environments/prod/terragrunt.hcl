# Production Environment - Main Configuration
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
  vpc_cidr           = "10.2.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  enable_nat_gateway = true
  single_nat_gateway = false  # HA NAT gateways for production
  enable_flow_logs   = true
  flow_logs_retention_days = 90
  enable_bastion     = true
  bastion_allowed_cidrs = ["10.0.0.0/8"]  # VPN only
  
  # EKS
  cluster_name           = "virtengine-prod"
  kubernetes_version     = "1.29"
  endpoint_private_access = true
  endpoint_public_access  = false  # Private only in production
  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
  cluster_log_retention_days = 90
  
  # System nodes
  system_node_instance_types = ["m5.large"]
  system_node_desired_size   = 3
  system_node_max_size       = 6
  system_node_min_size       = 3
  system_node_disk_size      = 100
  
  # Application nodes
  app_node_instance_types = ["m5.2xlarge", "m5.xlarge"]
  app_node_capacity_type  = "ON_DEMAND"
  app_node_desired_size   = 5
  app_node_max_size       = 20
  app_node_min_size       = 3
  app_node_disk_size      = 200
  
  # Chain nodes (dedicated for blockchain)
  chain_node_instance_types = ["m5.2xlarge"]
  chain_node_disk_size      = 500
  chain_node_desired_size   = 4  # Minimum for consensus
  chain_node_max_size       = 7
  chain_node_min_size       = 4
  
  # RDS
  rds_instance_class        = "db.r5.xlarge"
  rds_allocated_storage     = 500
  rds_max_allocated_storage = 2000
  rds_storage_type          = "io1"
  rds_iops                  = 10000
  rds_multi_az              = true
  rds_deletion_protection   = true
  rds_skip_final_snapshot   = false
  rds_backup_retention      = 35
  rds_monitoring_interval   = 60
  rds_performance_insights  = true
  rds_performance_insights_retention = 7
  create_read_replica       = true
  
  # Vault
  vault_replicas          = 5
  vault_memory_request    = "512Mi"
  vault_memory_limit      = "1Gi"
  vault_cpu_request       = "500m"
  vault_cpu_limit         = "1"
  enable_external_secrets = true
  enable_csi              = true
  
  # Monitoring
  log_retention_days        = 90
  prometheus_retention      = "30d"
  prometheus_retention_size = "100GB"
  prometheus_storage_size   = "200Gi"
  enable_grafana_ingress    = true
  grafana_hostname          = "grafana.virtengine.internal"
  
  # Alerts
  alert_emails = [
    "prod-alerts@virtengine.io",
    "oncall@virtengine.io"
  ]
}
