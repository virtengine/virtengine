#!/bin/bash
# VirtEngine ArgoCD Rollback Script
# Usage: ./argocd-rollback.sh <app-name> [revision]
# Example: ./argocd-rollback.sh virtengine-prod 5

set -euo pipefail

APP_NAME="${1:-}"
REVISION="${2:-}"
ARGOCD_SERVER="${ARGOCD_SERVER:-argocd.virtengine.internal}"
TIMEOUT="${TIMEOUT:-600}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    cat << EOF
VirtEngine ArgoCD Rollback Script

Usage: $0 <app-name> [revision]

Arguments:
  app-name    Name of the ArgoCD application to rollback
  revision    (Optional) Specific revision to rollback to. If not provided,
              will rollback to the previous revision.

Environment Variables:
  ARGOCD_SERVER   ArgoCD server address (default: argocd.virtengine.internal)
  TIMEOUT         Sync timeout in seconds (default: 600)

Examples:
  $0 virtengine-prod           # Rollback to previous revision
  $0 virtengine-prod 5         # Rollback to revision 5
  $0 provider-daemon-prod      # Rollback provider daemon
EOF
    exit 1
}

check_prerequisites() {
    if ! command -v argocd &> /dev/null; then
        log_error "argocd CLI not found. Please install it first."
        exit 1
    fi

    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install it first."
        exit 1
    fi
}

get_current_revision() {
    local app="$1"
    argocd app get "$app" -o json | jq -r '.status.history[-1].id'
}

get_previous_revision() {
    local app="$1"
    local history
    history=$(argocd app get "$app" -o json | jq -r '.status.history')
    local count
    count=$(echo "$history" | jq 'length')
    
    if [ "$count" -lt 2 ]; then
        log_error "No previous revision available for rollback"
        exit 1
    fi
    
    echo "$history" | jq -r '.[-2].id'
}

list_revisions() {
    local app="$1"
    log_info "Available revisions for $app:"
    argocd app history "$app" --output wide
}

perform_rollback() {
    local app="$1"
    local revision="$2"
    
    log_info "Starting rollback of $app to revision $revision"
    
    # Get current state for comparison
    local current_revision
    current_revision=$(get_current_revision "$app")
    log_info "Current revision: $current_revision"
    
    # Create a snapshot before rollback
    local snapshot_time
    snapshot_time=$(date +%Y%m%d-%H%M%S)
    log_info "Creating pre-rollback snapshot: $snapshot_time"
    kubectl get all -n "virtengine-${app##*-}" -o yaml > "/tmp/pre-rollback-${app}-${snapshot_time}.yaml" 2>/dev/null || true
    
    # Perform the rollback
    log_info "Executing rollback..."
    argocd app rollback "$app" "$revision" --prune
    
    # Wait for sync
    log_info "Waiting for sync to complete (timeout: ${TIMEOUT}s)..."
    if argocd app wait "$app" --timeout "$TIMEOUT" --health --sync; then
        log_info "Rollback completed successfully!"
        
        # Verify health
        local health
        health=$(argocd app get "$app" -o json | jq -r '.status.health.status')
        if [ "$health" = "Healthy" ]; then
            log_info "Application is healthy"
        else
            log_warn "Application health status: $health"
        fi
        
        return 0
    else
        log_error "Rollback failed or timed out"
        return 1
    fi
}

confirm_rollback() {
    local app="$1"
    local revision="$2"
    
    echo ""
    log_warn "You are about to rollback:"
    echo "  Application: $app"
    echo "  Target Revision: $revision"
    echo ""
    
    read -p "Are you sure you want to proceed? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        log_info "Rollback cancelled"
        exit 0
    fi
}

main() {
    if [ -z "$APP_NAME" ]; then
        usage
    fi
    
    check_prerequisites
    
    # Login check
    if ! argocd app list &> /dev/null; then
        log_error "Not logged into ArgoCD. Please run: argocd login $ARGOCD_SERVER"
        exit 1
    fi
    
    # Verify app exists
    if ! argocd app get "$APP_NAME" &> /dev/null; then
        log_error "Application '$APP_NAME' not found"
        exit 1
    fi
    
    # List available revisions
    list_revisions "$APP_NAME"
    echo ""
    
    # Determine target revision
    if [ -z "$REVISION" ]; then
        REVISION=$(get_previous_revision "$APP_NAME")
        log_info "No revision specified, using previous revision: $REVISION"
    fi
    
    # Confirm and perform rollback
    confirm_rollback "$APP_NAME" "$REVISION"
    perform_rollback "$APP_NAME" "$REVISION"
}

main "$@"
