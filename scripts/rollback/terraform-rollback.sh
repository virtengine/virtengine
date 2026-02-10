#!/bin/bash
# VirtEngine Terraform Rollback Script
# Usage: ./terraform-rollback.sh <environment> [steps-back]
# Example: ./terraform-rollback.sh prod 1

set -euo pipefail

ENVIRONMENT="${1:-}"
STEPS_BACK="${2:-1}"
TERRAFORM_DIR="${TERRAFORM_DIR:-infra/terraform/environments}"
S3_BUCKET="${S3_BUCKET:-virtengine-terraform-state}"
DYNAMODB_TABLE="${DYNAMODB_TABLE:-virtengine-terraform-locks}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    cat << EOF
VirtEngine Terraform State Rollback Script

Usage: $0 <environment> [steps-back]

Arguments:
  environment   Environment to rollback (dev, staging, prod)
  steps-back    Number of versions to rollback (default: 1)

Environment Variables:
  TERRAFORM_DIR    Path to terraform environments (default: infra/terraform/environments)
  S3_BUCKET        State bucket name
  DYNAMODB_TABLE   Lock table name

Examples:
  $0 prod           # Rollback prod to previous state
  $0 staging 2      # Rollback staging by 2 versions
EOF
    exit 1
}

check_prerequisites() {
    for cmd in aws terraform jq; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "$cmd not found. Please install it first."
            exit 1
        fi
    done
}

list_state_versions() {
    local env="$1"
    local key="$env/terraform.tfstate"
    
    log_info "Fetching state versions for $env..."
    aws s3api list-object-versions \
        --bucket "$S3_BUCKET" \
        --prefix "$key" \
        --query 'Versions[*].{VersionId:VersionId,LastModified:LastModified,Size:Size}' \
        --output table
}

get_version_id() {
    local env="$1"
    local steps="$2"
    local key="$env/terraform.tfstate"
    
    aws s3api list-object-versions \
        --bucket "$S3_BUCKET" \
        --prefix "$key" \
        --query "Versions[$steps].VersionId" \
        --output text
}

backup_current_state() {
    local env="$1"
    local backup_dir="/tmp/terraform-backup-$(date +%Y%m%d-%H%M%S)"
    
    mkdir -p "$backup_dir"
    log_info "Backing up current state to $backup_dir"
    
    cd "$TERRAFORM_DIR/$env"
    terraform state pull > "$backup_dir/${env}-current.tfstate"
    
    log_info "Backup created: $backup_dir/${env}-current.tfstate"
    echo "$backup_dir"
}

restore_state_version() {
    local env="$1"
    local version_id="$2"
    local key="$env/terraform.tfstate"
    
    log_info "Restoring state version: $version_id"
    
    # Download the specific version
    local temp_state="/tmp/terraform-restore-${env}-${version_id}.tfstate"
    aws s3api get-object \
        --bucket "$S3_BUCKET" \
        --key "$key" \
        --version-id "$version_id" \
        "$temp_state"
    
    # Push the restored state
    cd "$TERRAFORM_DIR/$env"
    terraform state push "$temp_state"
    
    log_info "State restored successfully"
    rm -f "$temp_state"
}

plan_after_rollback() {
    local env="$1"
    
    log_info "Running terraform plan to verify state..."
    cd "$TERRAFORM_DIR/$env"
    
    if terraform plan -detailed-exitcode; then
        log_info "No changes detected - rollback successful"
    else
        log_warn "Terraform detected changes. Review the plan above."
        log_warn "You may need to run 'terraform apply' to reconcile."
    fi
}

confirm_rollback() {
    local env="$1"
    local steps="$2"
    local version_id="$3"
    
    echo ""
    log_warn "You are about to rollback Terraform state:"
    echo "  Environment: $env"
    echo "  Steps back: $steps"
    echo "  Target version: $version_id"
    echo ""
    log_warn "This will restore the infrastructure state to a previous point."
    log_warn "Running 'terraform apply' afterwards may destroy resources!"
    echo ""
    
    read -p "Are you sure you want to proceed? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        log_info "Rollback cancelled"
        exit 0
    fi
}

main() {
    if [ -z "$ENVIRONMENT" ]; then
        usage
    fi
    
    if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
        log_error "Invalid environment: $ENVIRONMENT"
        usage
    fi
    
    check_prerequisites
    
    # Verify environment directory exists
    if [ ! -d "$TERRAFORM_DIR/$ENVIRONMENT" ]; then
        log_error "Environment directory not found: $TERRAFORM_DIR/$ENVIRONMENT"
        exit 1
    fi
    
    # List available versions
    list_state_versions "$ENVIRONMENT"
    
    # Get target version
    VERSION_ID=$(get_version_id "$ENVIRONMENT" "$STEPS_BACK")
    if [ -z "$VERSION_ID" ] || [ "$VERSION_ID" = "None" ]; then
        log_error "No version found at $STEPS_BACK steps back"
        exit 1
    fi
    
    # Confirm and perform rollback
    confirm_rollback "$ENVIRONMENT" "$STEPS_BACK" "$VERSION_ID"
    
    # Backup current state
    BACKUP_DIR=$(backup_current_state "$ENVIRONMENT")
    
    # Restore previous version
    restore_state_version "$ENVIRONMENT" "$VERSION_ID"
    
    # Run plan to verify
    plan_after_rollback "$ENVIRONMENT"
    
    echo ""
    log_info "Rollback complete!"
    log_info "Backup saved at: $BACKUP_DIR"
    log_warn "Review the terraform plan output above before applying any changes."
}

main "$@"
