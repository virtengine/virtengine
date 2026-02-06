# VirtEngine DNS and Global Load Balancing Module
# Manages Route53 zones, health checks, and geo-routing

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  tags = merge(var.tags, {
    Module      = "dns"
    Environment = var.environment
  })
}

# -----------------------------------------------------------------------------
# Route53 Hosted Zone
# -----------------------------------------------------------------------------
data "aws_route53_zone" "main" {
  count = var.create_hosted_zone ? 0 : 1
  name  = var.domain_name
}

resource "aws_route53_zone" "main" {
  count = var.create_hosted_zone ? 1 : 0
  name  = var.domain_name

  tags = merge(local.tags, {
    Name = var.domain_name
  })
}

locals {
  zone_id = var.create_hosted_zone ? aws_route53_zone.main[0].zone_id : data.aws_route53_zone.main[0].zone_id
}

# -----------------------------------------------------------------------------
# Health Checks per Region
# -----------------------------------------------------------------------------
resource "aws_route53_health_check" "regional" {
  for_each = var.regional_endpoints

  fqdn              = each.value.health_check_fqdn
  port               = each.value.health_check_port
  type               = "HTTPS"
  resource_path      = each.value.health_check_path
  failure_threshold  = 3
  request_interval   = 10
  measure_latency    = true

  tags = merge(local.tags, {
    Name   = "virtengine-${each.key}-health"
    Region = each.key
  })
}

# -----------------------------------------------------------------------------
# Global API Endpoint (Latency-Based Routing)
# -----------------------------------------------------------------------------
resource "aws_route53_record" "api_regional" {
  for_each = var.regional_endpoints

  zone_id = local.zone_id
  name    = "api.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  latency_routing_policy {
    region = each.key
  }

  alias {
    name                   = each.value.lb_dns_name
    zone_id                = each.value.lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# -----------------------------------------------------------------------------
# Global RPC Endpoint (Latency-Based Routing)
# -----------------------------------------------------------------------------
resource "aws_route53_record" "rpc_regional" {
  for_each = var.regional_endpoints

  zone_id = local.zone_id
  name    = "rpc.${var.domain_name}"
  type    = "A"

  set_identifier = each.key

  latency_routing_policy {
    region = each.key
  }

  alias {
    name                   = each.value.lb_dns_name
    zone_id                = each.value.lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[each.key].id
}

# -----------------------------------------------------------------------------
# Regional RPC Endpoints (Direct)
# -----------------------------------------------------------------------------
resource "aws_route53_record" "rpc_direct" {
  for_each = var.regional_endpoints

  zone_id = local.zone_id
  name    = "rpc-${each.key}.${var.domain_name}"
  type    = "A"

  alias {
    name                   = each.value.lb_dns_name
    zone_id                = each.value.lb_zone_id
    evaluate_target_health = true
  }
}

# -----------------------------------------------------------------------------
# Failover Records (Primary/Secondary)
# -----------------------------------------------------------------------------
resource "aws_route53_record" "api_failover_primary" {
  count = var.enable_failover ? 1 : 0

  zone_id = local.zone_id
  name    = "api-failover.${var.domain_name}"
  type    = "A"

  set_identifier = "primary"

  failover_routing_policy {
    type = "PRIMARY"
  }

  alias {
    name                   = var.regional_endpoints[var.primary_region].lb_dns_name
    zone_id                = var.regional_endpoints[var.primary_region].lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[var.primary_region].id
}

resource "aws_route53_record" "api_failover_secondary" {
  count = var.enable_failover ? 1 : 0

  zone_id = local.zone_id
  name    = "api-failover.${var.domain_name}"
  type    = "A"

  set_identifier = "secondary"

  failover_routing_policy {
    type = "SECONDARY"
  }

  alias {
    name                   = var.regional_endpoints[var.secondary_region].lb_dns_name
    zone_id                = var.regional_endpoints[var.secondary_region].lb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.regional[var.secondary_region].id
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms for Health Checks
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "health_check" {
  for_each = var.regional_endpoints

  alarm_name          = "virtengine-${each.key}-health-check"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = 60
  statistic           = "Minimum"
  threshold           = 1
  alarm_description   = "Route53 health check failed for ${each.key}"
  alarm_actions       = var.alarm_sns_topic_arns

  dimensions = {
    HealthCheckId = aws_route53_health_check.regional[each.key].id
  }

  tags = local.tags
}
