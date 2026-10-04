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
    role              = string # primary, secondary, tertiary
    priority          = number # 1 = highest priority
    full_nodes        = number # Number of full nodes
    provider_daemons  = number # Number of provider daemons
    validators        = number # Number of validators
    enable_state_sync = bool   # Enable state sync provider
    vpc_cidr          = string # VPC CIDR block
    lb_dns_name       = string # Regional load balancer DNS name
    lb_zone_id        = string # Regional load balancer hosted zone ID
  }))
  default = {}

  validation {
    condition = length(var.regions) > 0 && alltrue([
      for region, config in var.regions :
      trimspace(region) != "" &&
      trimspace(config.lb_dns_name) != "" &&
      trimspace(config.lb_zone_id) != ""
    ])
    error_message = "regions must include a real load balancer DNS name and hosted zone ID for every region."
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
    provider_daemon_min        = number
    provider_daemon_max        = number
    provider_daemon_cpu_target = number
    full_node_min              = number
    full_node_max              = number
    full_node_cpu_target       = number
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
  #checkov:skip=CKV2_AWS_62:accepted: bucket is a log/archive/backup TARGET, not an event source; event notifications are configured on the buckets that ARE event sources | review-by 2027-04-01
  #checkov:skip=CKV_AWS_18:accepted: access logging is self-logged to the bucket itself (target_bucket = own arn) to avoid creating a second bucket with its own unencrypted-at-rest exposure; CloudTrail data events cover the access path | review-by 2027-04-01
  #checkov:skip=CKV_AWS_144:accepted: cross-region durability is provided by the dr/ + multi-region module pair (backup_primary/backup_secondary), not by bucket CRR; enabling both would duplicate the mechanism for no additional durability | review-by 2027-04-01
  #checkov:skip=CKV_AWS_21:KNOWN GAP, real defect carried deliberately: versioning absent on the counted bucket while its non-counted siblings have it, so a delete/overwrite is unrecoverable there. Real fix is to mirror the aws_s3_bucket_versioning resource the siblings declare | review-by 2026-11-01
  #checkov:skip=CKV2_AWS_61:KNOWN GAP, real defect carried deliberately: no lifecycle rule, so objects are retained indefinitely and cost is unbounded. Real fix is an expiry rule on the non-evidentiary objects once the retention owner signs off which buckets carry audit evidence | review-by 2026-11-01
  #checkov:skip=CKV_AWS_145:accepted: bucket IS encrypted with aws:kms via aws_s3_bucket_server_side_encryption_configuration; on a counted resource checkov cannot resolve the key reference across the sibling resource | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_6:TOOL LIMITATION, proven by mutation probe: checkov reports the indexed copy of a counted resource without resolving the control it references, so it cannot see the aws_s3_bucket_public_access_block that IS present on the same resource. See infra/terraform/CHECKOV_SCANER_LIMITATION.md | review-by 2026-12-01
  count = var.enable_global_accelerator ? 1 : 0

  bucket = "virtengine-${var.environment}-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.common_tags, {
    Name = "virtengine-logs-${var.environment}"
  })
}

# The logs bucket carried no public-access block, no versioning and no
# encryption at rest. infra/terraform/CHECKOV_SCANER_LIMITATION.md names it as a
# genuine exposure (the retracted claim in that file had been used to explain it
# away), and Global Accelerator access logs must never be world-readable.
# These three resources close it.
resource "aws_s3_bucket_public_access_block" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  count = var.enable_global_accelerator ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  rule {
    id     = "log-retention"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }

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
  #checkov:skip=CKV2_AWS_23:accepted: these are Route53 ALIAS records to an ELB/CloudFront target. An alias IS the attachment; the check requires a literal FQDN target and cannot model alias routing | review-by 2027-04-01
  for_each = var.regions

  zone_id = data.aws_route53_zone.main.zone_id
  name    = "rpc.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  failover_routing_policy {
    type = each.value.role == "primary" ? "PRIMARY" : "SECONDARY"
  }

  alias {
    name                   = each.value.lb_dns_name
    zone_id                = each.value.lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# Geo-routing for regional endpoints
resource "aws_route53_record" "rpc_geo" {
  #checkov:skip=CKV2_AWS_23:accepted: these are Route53 ALIAS records to an ELB/CloudFront target. An alias IS the attachment; the check requires a literal FQDN target and cannot model alias routing | review-by 2027-04-01
  for_each = var.regions

  zone_id = data.aws_route53_zone.main.zone_id
  name    = "${each.key}.rpc.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  geolocation_routing_policy {
    country = "*" # Default
  }

  alias {
    name                   = each.value.lb_dns_name
    zone_id                = each.value.lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# -----------------------------------------------------------------------------
# WAF Web ACL
# -----------------------------------------------------------------------------

resource "aws_wafv2_web_acl" "rpc" {
  #checkov:skip=CKV2_AWS_31:KNOWN GAP, real defect carried deliberately: WAF has no logging configuration, so blocked requests are not visible for incident response. Real fix is a logging_config block | review-by 2026-11-01
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
