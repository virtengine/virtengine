#!/usr/bin/env bash
# Update Global Load Balancer
# Manages Route53 weighted records for multi-region traffic distribution.
set -euo pipefail

DOMAIN="${VE_DOMAIN:-virtengine.io}"
REGIONS=("us-east-1" "eu-west-1" "ap-southeast-1")

usage() {
  cat <<EOF
Usage: $(basename "$0") [COMMAND] [OPTIONS]

Commands:
  status          Show current DNS routing status (default)
  enable REGION   Enable traffic to a region
  disable REGION  Disable traffic to a region
  failover REGION Route all traffic to specified region
  restore         Restore normal multi-region routing

Options:
  --domain DOMAIN   Override domain (default: $DOMAIN)
  --dry-run         Show what would change without applying

Examples:
  $(basename "$0") status
  $(basename "$0") disable us-east-1
  $(basename "$0") failover eu-west-1
  $(basename "$0") restore
EOF
  exit 1
}

DRY_RUN=false
COMMAND="status"
TARGET_REGION=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    status|enable|disable|failover|restore)
      COMMAND="$1"
      shift
      if [[ "$COMMAND" =~ ^(enable|disable|failover)$ ]] && [[ $# -gt 0 ]] && [[ ! "$1" =~ ^-- ]]; then
        TARGET_REGION="$1"
        shift
      fi
      ;;
    --domain) DOMAIN="$2"; shift 2 ;;
    --dry-run) DRY_RUN=true; shift ;;
    *) usage ;;
  esac
done

# Get hosted zone ID
get_zone_id() {
  aws route53 list-hosted-zones-by-name \
    --dns-name "$DOMAIN" \
    --query "HostedZones[0].Id" \
    --output text 2>/dev/null | sed 's|/hostedzone/||'
}

# Show current status
show_status() {
  echo "=== Global LB Status for ${DOMAIN} ==="
  echo ""

  zone_id=$(get_zone_id)
  if [ -z "$zone_id" ] || [ "$zone_id" = "None" ]; then
    echo "  ✗ Hosted zone not found for $DOMAIN"
    exit 1
  fi

  echo "Zone ID: $zone_id"
  echo ""

  for record_prefix in "api" "rpc"; do
    echo "--- ${record_prefix}.${DOMAIN} ---"
    aws route53 list-resource-record-sets \
      --hosted-zone-id "$zone_id" \
      --query "ResourceRecordSets[?Name=='${record_prefix}.${DOMAIN}.'].[SetIdentifier,Weight,Region,AliasTarget.DNSName]" \
      --output table 2>/dev/null || echo "  (no records found)"
    echo ""
  done

  echo "--- Health Checks ---"
  aws route53 list-health-checks \
    --query "HealthChecks[?HealthCheckConfig.FullyQualifiedDomainName!=null].[Id,HealthCheckConfig.FullyQualifiedDomainName,HealthCheckConfig.Port]" \
    --output table 2>/dev/null || echo "  (no health checks found)"
}

# Enable a region
enable_region() {
  local region="$1"
  echo "Enabling region: $region"

  zone_id=$(get_zone_id)
  # Re-enable health check
  health_check_id=$(aws route53 list-health-checks \
    --query "HealthChecks[?HealthCheckConfig.FullyQualifiedDomainName=='rpc-${region}.${DOMAIN}'].Id" \
    --output text 2>/dev/null || echo "")

  if [ -n "$health_check_id" ] && [ "$health_check_id" != "None" ]; then
    if [ "$DRY_RUN" = true ]; then
      echo "  [dry-run] Would enable health check $health_check_id"
    else
      aws route53 update-health-check \
        --health-check-id "$health_check_id" \
        --disabled=false 2>/dev/null
      echo "  ✓ Health check $health_check_id enabled"
    fi
  else
    echo "  ○ No health check found for $region"
  fi
}

# Disable a region
disable_region() {
  local region="$1"
  echo "Disabling region: $region"

  health_check_id=$(aws route53 list-health-checks \
    --query "HealthChecks[?HealthCheckConfig.FullyQualifiedDomainName=='rpc-${region}.${DOMAIN}'].Id" \
    --output text 2>/dev/null || echo "")

  if [ -n "$health_check_id" ] && [ "$health_check_id" != "None" ]; then
    if [ "$DRY_RUN" = true ]; then
      echo "  [dry-run] Would disable health check $health_check_id"
    else
      aws route53 update-health-check \
        --health-check-id "$health_check_id" \
        --disabled=true 2>/dev/null
      echo "  ✓ Health check $health_check_id disabled (region removed from rotation)"
    fi
  else
    echo "  ○ No health check found for $region"
  fi
}

# Failover: route all traffic to one region
do_failover() {
  local target="$1"
  echo "Failing over all traffic to: $target"
  echo ""

  for region in "${REGIONS[@]}"; do
    if [ "$region" = "$target" ]; then
      enable_region "$region"
    else
      disable_region "$region"
    fi
  done

  echo ""
  echo "  ⚠  Failover to $target complete. Only $target is receiving traffic."
  echo "  Run '$(basename "$0") restore' when ready to return to normal routing."
}

# Restore normal routing
do_restore() {
  echo "Restoring normal multi-region routing..."
  echo ""

  for region in "${REGIONS[@]}"; do
    enable_region "$region"
  done

  echo ""
  echo "  ✓ All regions re-enabled for traffic."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
case "$COMMAND" in
  status)
    show_status
    ;;
  enable)
    [ -z "$TARGET_REGION" ] && { echo "Error: region required"; usage; }
    enable_region "$TARGET_REGION"
    ;;
  disable)
    [ -z "$TARGET_REGION" ] && { echo "Error: region required"; usage; }
    disable_region "$TARGET_REGION"
    ;;
  failover)
    [ -z "$TARGET_REGION" ] && { echo "Error: region required"; usage; }
    do_failover "$TARGET_REGION"
    ;;
  restore)
    do_restore
    ;;
  *)
    usage
    ;;
esac
