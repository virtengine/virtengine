#!/bin/bash
# VirtEngine Blue/Green Traffic Switch Script
# Usage: ./blue-green-switch.sh <app> <target-version>
# Example: ./blue-green-switch.sh virtengine-node green

set -euo pipefail

APP="${1:-}"
TARGET="${2:-}"
NAMESPACE="${NAMESPACE:-virtengine-prod}"
WEIGHT_STEP="${WEIGHT_STEP:-10}"
WAIT_SECONDS="${WAIT_SECONDS:-30}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    cat << EOF
VirtEngine Blue/Green Traffic Switch Script

Usage: $0 <app> <target-version>

Arguments:
  app              Application name (virtengine-node, provider-daemon)
  target-version   Target version to switch to (blue, green)

Environment Variables:
  NAMESPACE        Kubernetes namespace (default: virtengine-prod)
  WEIGHT_STEP      Traffic weight increment (default: 10)
  WAIT_SECONDS     Wait between increments (default: 30)

Examples:
  $0 virtengine-node green     # Switch to green deployment
  $0 provider-daemon blue      # Switch back to blue
EOF
    exit 1
}

check_prerequisites() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found"
        exit 1
    fi
    
    if ! command -v istioctl &> /dev/null; then
        log_warn "istioctl not found - some checks will be skipped"
    fi
}

get_current_weights() {
    local app="$1"
    kubectl get virtualservice "$app" -n "$NAMESPACE" -o json | \
        jq -r '.spec.http[-1].route[] | "\(.destination.host): \(.weight)"'
}

check_target_health() {
    local app="$1"
    local target="$2"
    
    log_info "Checking health of $app-$target..."
    
    local ready
    ready=$(kubectl get pods -n "$NAMESPACE" \
        -l "app.kubernetes.io/name=$app,version=$target" \
        -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}')
    
    if [[ "$ready" == *"False"* ]] || [ -z "$ready" ]; then
        log_error "Target pods are not healthy"
        return 1
    fi
    
    log_info "Target pods are healthy"
    return 0
}

update_weights() {
    local app="$1"
    local blue_weight="$2"
    local green_weight="$3"
    
    log_info "Setting weights: blue=$blue_weight, green=$green_weight"
    
    kubectl patch virtualservice "$app" -n "$NAMESPACE" --type='json' -p="[
        {\"op\": \"replace\", \"path\": \"/spec/http/-1/route/0/weight\", \"value\": $blue_weight},
        {\"op\": \"replace\", \"path\": \"/spec/http/-1/route/1/weight\", \"value\": $green_weight}
    ]"
}

gradual_switch() {
    local app="$1"
    local target="$2"
    
    local current_blue=100
    local current_green=0
    
    if [ "$target" = "blue" ]; then
        current_blue=0
        current_green=100
    fi
    
    log_info "Starting gradual traffic switch to $target"
    
    while true; do
        if [ "$target" = "green" ]; then
            current_blue=$((current_blue - WEIGHT_STEP))
            current_green=$((current_green + WEIGHT_STEP))
            
            if [ "$current_green" -gt 100 ]; then
                current_green=100
                current_blue=0
            fi
        else
            current_green=$((current_green - WEIGHT_STEP))
            current_blue=$((current_blue + WEIGHT_STEP))
            
            if [ "$current_blue" -gt 100 ]; then
                current_blue=100
                current_green=0
            fi
        fi
        
        update_weights "$app" "$current_blue" "$current_green"
        
        log_info "Current weights: blue=$current_blue%, green=$current_green%"
        
        if [ "$target" = "green" ] && [ "$current_green" -eq 100 ]; then
            break
        fi
        if [ "$target" = "blue" ] && [ "$current_blue" -eq 100 ]; then
            break
        fi
        
        log_info "Waiting ${WAIT_SECONDS}s before next increment..."
        sleep "$WAIT_SECONDS"
        
        # Check for errors
        if ! check_error_rate "$app" "$target"; then
            log_error "Error rate too high, aborting switch"
            instant_rollback "$app" "$target"
            exit 1
        fi
    done
    
    log_info "Traffic switch complete!"
}

instant_switch() {
    local app="$1"
    local target="$2"
    
    log_info "Performing instant switch to $target"
    
    if [ "$target" = "green" ]; then
        update_weights "$app" 0 100
    else
        update_weights "$app" 100 0
    fi
    
    log_info "Traffic switched to $target"
}

instant_rollback() {
    local app="$1"
    local failed_target="$2"
    
    log_warn "Rolling back from $failed_target"
    
    if [ "$failed_target" = "green" ]; then
        update_weights "$app" 100 0
    else
        update_weights "$app" 0 100
    fi
    
    log_info "Rollback complete"
}

check_error_rate() {
    local app="$1"
    local target="$2"
    
    # This is a simplified check - in production, query Prometheus
    # Example: error_rate=$(prometheus_query "rate(http_requests_total{status=~'5..'}[1m])")
    
    log_info "Checking error rates..."
    # Placeholder - implement actual metric check
    return 0
}

confirm_switch() {
    local app="$1"
    local target="$2"
    
    echo ""
    log_warn "You are about to switch traffic:"
    echo "  Application: $app"
    echo "  Namespace: $NAMESPACE"
    echo "  Target version: $target"
    echo ""
    echo "Current weights:"
    get_current_weights "$app"
    echo ""
    
    read -p "Switch type (gradual/instant/cancel): " switch_type
    
    case "$switch_type" in
        gradual)
            gradual_switch "$app" "$target"
            ;;
        instant)
            instant_switch "$app" "$target"
            ;;
        cancel)
            log_info "Switch cancelled"
            exit 0
            ;;
        *)
            log_error "Invalid option"
            exit 1
            ;;
    esac
}

main() {
    if [ -z "$APP" ] || [ -z "$TARGET" ]; then
        usage
    fi
    
    if [[ ! "$TARGET" =~ ^(blue|green)$ ]]; then
        log_error "Target must be 'blue' or 'green'"
        usage
    fi
    
    check_prerequisites
    
    # Verify VirtualService exists
    if ! kubectl get virtualservice "$APP" -n "$NAMESPACE" &> /dev/null; then
        log_error "VirtualService '$APP' not found in namespace '$NAMESPACE'"
        exit 1
    fi
    
    # Check target health
    if ! check_target_health "$APP" "$TARGET"; then
        log_error "Target deployment is not healthy. Fix issues before switching."
        exit 1
    fi
    
    # Confirm and perform switch
    confirm_switch "$APP" "$TARGET"
}

main "$@"
