# DNS Module Outputs

output "zone_id" {
  description = "Route53 hosted zone ID"
  value       = local.zone_id
}

output "api_fqdn" {
  description = "FQDN for the global API endpoint"
  value       = "api.${var.domain_name}"
}

output "rpc_fqdn" {
  description = "FQDN for the global RPC endpoint"
  value       = "rpc.${var.domain_name}"
}

output "regional_rpc_fqdns" {
  description = "Map of region to direct RPC FQDN"
  value = {
    for k, v in var.regional_endpoints : k => "rpc-${k}.${var.domain_name}"
  }
}

output "health_check_ids" {
  description = "Map of region to Route53 health check ID"
  value = {
    for k, v in aws_route53_health_check.regional : k => v.id
  }
}
