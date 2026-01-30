# SCALE-002: Multi-Region Scaling Module
# Terraform module for horizontal scaling infrastructure
#
# This module provisions:
# - Global load balancers (AWS Global Accelerator)
# - Regional application load balancers
# - Auto Scaling groups for provider daemons
# - Cross-region VPC peering
# - Route53 health checks and failover routing

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "regions" {
  description = "Map of regions to deploy with their configurations"
  type = map(object({
    role               = string       # primary, secondary, tertiary
    priority           = number       # 1 = highest priority
    full_nodes         = number       # Number of full nodes
    provider_daemons   = number       # Number of provider daemons
    validators         = number       # Number of validators
    enable_state_sync  = bool         # Enable state sync provider
    vpc_cidr           = string       # VPC CIDR block
  }))

  default = {
    "us-east-1" = {
      role               = "primary"
      priority           = 1
      full_nodes         = 4
      provider_daemons   = 4
      validators         = 2
      enable_state_sync  = true
      vpc_cidr           = "10.0.0.0/16"
    }
    "eu-west-1" = {
      role               = "secondary"
      priority           = 2
      full_nodes         = 3
      provider_daemons   = 2
      validators         = 2
      enable_state_sync  = true
      vpc_cidr           = "10.1.0.0/16"
    }
    "ap-south-1" = {
      role               = "tertiary"
      priority           = 3
      full_nodes         = 2
      provider_daemons   = 1
      validators         = 1
      enable_state_sync  = true
      vpc_cidr           = "10.2.0.0/16"
    }
  }
}

variable "domain_name" {
  description = "Domain name for VirtEngine (e.g., virtengine.network)"
  type        = string
  default     = "virtengine.network"
}

variable "certificate_arn" {
  description = "ACM certificate ARN for TLS"
  type        = string
}

variable "enable_global_accelerator" {
  description = "Enable AWS Global Accelerator for low-latency routing"
  type        = bool
  default     = true
}

variable "enable_waf" {
  description = "Enable WAF protection for load balancers"
  type        = bool
  default     = true
}

variable "waf_rate_limit" {
  description = "WAF rate limit (requests per 5 minutes per IP)"
  type        = number
  default     = 2000
}

variable "scaling_config" {
  description = "Auto scaling configuration"
  type = object({
    provider_daemon_min     = number
    provider_daemon_max     = number
    provider_daemon_cpu_target = number
    full_node_min          = number
    full_node_max          = number
    full_node_cpu_target   = number
  })
  default = {
    provider_daemon_min        = 2
    provider_daemon_max        = 20
    provider_daemon_cpu_target = 70
    full_node_min              = 3
    full_node_max              = 12
    full_node_cpu_target       = 65
  }
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Local Values
# -----------------------------------------------------------------------------

locals {
  common_tags = merge(var.tags, {
    Project     = "virtengine"
    Environment = var.environment
    ManagedBy   = "terraform"
    Module      = "scaling"
  })

  primary_region = [for region, config in var.regions : region if config.role == "primary"][0]
}

# -----------------------------------------------------------------------------
# Global Accelerator
# -----------------------------------------------------------------------------

resource "aws_globalaccelerator_accelerator" "main" {
  count = var.enable_global_accelerator ? 1 : 0

  name            = "virtengine-${var.environment}"
  ip_address_type = "IPV4"
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = aws_s3_bucket.logs[0].id
    flow_logs_s3_prefix = "global-accelerator/"
  }

  tags = merge(local.common_tags, {
    Name = "virtengine-global-accelerator-${var.environment}"
  })
}

# RPC Listener (HTTPS)
resource "aws_globalaccelerator_listener" "rpc" {
  count = var.enable_global_accelerator ? 1 : 0

  accelerator_arn = aws_globalaccelerator_accelerator.main[0].id
  client_affinity = "NONE"
  protocol        = "TCP"

  port_range {
    from_port = 443
    to_port   = 443
  }
}

# gRPC Listener
resource "aws_globalaccelerator_listener" "grpc" {
  count = var.enable_global_accelerator ? 1 : 0

  accelerator_arn = aws_globalaccelerator_accelerator.main[0].id
  client_affinity = "NONE"
  protocol        = "TCP"

  port_range {
    from_port = 9090
    to_port   = 9090
  }
}

# -----------------------------------------------------------------------------
# S3 Bucket for Logs
# -----------------------------------------------------------------------------

resource "aws_s3_bucket" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = "virtengine-${var.environment}-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.common_tags, {
    Name = "virtengine-logs-${var.environment}"
  })
}

resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  rule {
    id     = "log-retention"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }
}

# -----------------------------------------------------------------------------
# Route53 Health Checks
# -----------------------------------------------------------------------------

resource "aws_route53_health_check" "regional" {
  for_each = var.regions

  fqdn              = "${each.key}.rpc.${var.domain_name}"
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 10

  tags = merge(local.common_tags, {
    Name   = "virtengine-${each.key}-health"
    Region = each.key
  })
}

# -----------------------------------------------------------------------------
# Route53 Failover Routing
# -----------------------------------------------------------------------------

resource "aws_route53_record" "rpc_regional" {
  for_each = var.regions

  zone_id = data.aws_route53_zone.main.zone_id
  name    = "rpc.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  failover_routing_policy {
    type = each.value.role == "primary" ? "PRIMARY" : "SECONDARY"
  }

  alias {
    name                   = "placeholder.elb.${each.key}.amazonaws.com"  # Replace with actual ALB
    zone_id                = "Z35SXDOTRQ7X7K"  # Placeholder - use actual zone ID
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# Geo-routing for regional endpoints
resource "aws_route53_record" "rpc_geo" {
  for_each = var.regions

  zone_id = data.aws_route53_zone.main.zone_id
  name    = "${each.key}.rpc.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  geolocation_routing_policy {
    country = "*"  # Default
  }

  alias {
    name                   = "placeholder.elb.${each.key}.amazonaws.com"
    zone_id                = "Z35SXDOTRQ7X7K"
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# -----------------------------------------------------------------------------
# WAF Web ACL
# -----------------------------------------------------------------------------

resource "aws_wafv2_web_acl" "rpc" {
  count = var.enable_waf ? 1 : 0

  name  = "virtengine-rpc-waf-${var.environment}"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  # Rate limiting rule
  rule {
    name     = "rate-limit"
    priority = 1

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = var.waf_rate_limit
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
    name     = "aws-managed-common"
    priority = 2

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        vendor_name = "AWS"
        name        = "AWSManagedRulesCommonRuleSet"
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
    name     = "aws-managed-bad-inputs"
    priority = 3

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        vendor_name = "AWS"
        name        = "AWSManagedRulesKnownBadInputsRuleSet"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesKnownBadInputs"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "VirtEngineRPCWAF"
    sampled_requests_enabled   = true
  }

  tags = merge(local.common_tags, {
    Name = "virtengine-waf-${var.environment}"
  })
}

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------

data "aws_caller_identity" "current" {}

data "aws_route53_zone" "main" {
  name         = var.domain_name
  private_zone = false
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "global_accelerator_dns" {
  description = "Global Accelerator DNS name"
  value       = var.enable_global_accelerator ? aws_globalaccelerator_accelerator.main[0].dns_name : null
}

output "global_accelerator_ips" {
  description = "Global Accelerator IP addresses"
  value       = var.enable_global_accelerator ? aws_globalaccelerator_accelerator.main[0].ip_sets : null
}

output "health_check_ids" {
  description = "Route53 health check IDs by region"
  value       = { for k, v in aws_route53_health_check.regional : k => v.id }
}

output "waf_web_acl_arn" {
  description = "WAF Web ACL ARN"
  value       = var.enable_waf ? aws_wafv2_web_acl.rpc[0].arn : null
}

output "regions_config" {
  description = "Configured regions"
  value       = var.regions
}

output "scaling_config" {
  description = "Auto scaling configuration"
  value       = var.scaling_config
}
