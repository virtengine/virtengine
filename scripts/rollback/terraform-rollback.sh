#!/usr/bin/env bash
# VirtEngine Terraform state rollback helper.
# Restores a previous backend state version and emits a reviewed follow-up plan.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVIRONMENT=""
STEPS_BACK="1"
DRY_RUN="false"
AUTO_APPROVE="false"
ARTIFACT_DIR=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

usage() {
  cat <<'EOF'
VirtEngine Terraform State Rollback Script

Usage:
  terraform-rollback.sh <environment> [--steps-back N] [--artifact-dir DIR] [--dry-run] [--yes]

Arguments:
  environment         Environment to roll back (dev, staging, prod)

Options:
  --steps-back N      State-object versions back from latest non-current version (default: 1)
  --artifact-dir DIR  Directory for evidence and downloaded state artifacts
  --dry-run           Resolve and download the target state version without pushing it
  --yes               Skip the interactive confirmation prompt
EOF
}

fail() {
  log_error "$*"
  exit 1
}

check_prerequisites() {
  local commands=(aws terraform jq awk sed grep)
  for cmd in "${commands[@]}"; do
    command -v "$cmd" >/dev/null 2>&1 || fail "$cmd not found in PATH"
  done
}

extract_backend_value() {
  local file="$1"
  local key="$2"
  awk -F'"' -v wanted="$key" '$1 ~ ("^[[:space:]]*" wanted "[[:space:]]*=") { print $2; exit }' "$file"
}

terraform_target_dir() {
  echo "$ROOT_DIR/infra/terraform/environments/$ENVIRONMENT"
}

state_backend_bucket() {
  extract_backend_value "$(terraform_target_dir)/main.tf" "bucket"
}

state_backend_key() {
  extract_backend_value "$(terraform_target_dir)/main.tf" "key"
}

run_init() {
  local target_dir
  target_dir="$(terraform_target_dir)"
  if [[ -f "$target_dir/terragrunt.hcl" ]]; then
    (cd "$target_dir" && terragrunt init -input=false -no-color)
  else
    (cd "$target_dir" && terraform init -input=false -no-color)
  fi
}

pull_current_state() {
  local destination="$1"
  local target_dir
  target_dir="$(terraform_target_dir)"
  if [[ -f "$target_dir/terragrunt.hcl" ]]; then
    (cd "$target_dir" && terragrunt state pull > "$destination")
  else
    terraform -chdir="$target_dir" state pull > "$destination"
  fi
}

push_state() {
  local source_file="$1"
  local target_dir
  target_dir="$(terraform_target_dir)"
  if [[ -f "$target_dir/terragrunt.hcl" ]]; then
    (cd "$target_dir" && terragrunt state push "$source_file")
  else
    terraform -chdir="$target_dir" state push "$source_file"
  fi
}

list_state_versions() {
  local bucket="$1"
  local key="$2"
  aws s3api list-object-versions \
    --bucket "$bucket" \
    --prefix "$key" \
    --query 'Versions[?IsLatest==`false`].{VersionId:VersionId,LastModified:LastModified,Size:Size}' \
    --output table
}

get_target_version_id() {
  local bucket="$1"
  local key="$2"
  local index="$3"
  aws s3api list-object-versions \
    --bucket "$bucket" \
    --prefix "$key" \
    --query "Versions[?IsLatest==\`false\`][$((index - 1))].VersionId" \
    --output text
}

download_state_version() {
  local bucket="$1"
  local key="$2"
  local version_id="$3"
  local output_file="$4"
  aws s3api get-object \
    --bucket "$bucket" \
    --key "$key" \
    --version-id "$version_id" \
    "$output_file" >/dev/null
}

confirm_rollback() {
  local version_id="$1"
  if [[ "$AUTO_APPROVE" == "true" ]]; then
    return 0
  fi

  echo
  log_warn "You are about to restore a prior Terraform state version."
  echo "  Environment: $ENVIRONMENT"
  echo "  Backend bucket: $STATE_BUCKET"
  echo "  Backend key: $STATE_KEY"
  echo "  Target version: $version_id"
  echo "  Dry run: $DRY_RUN"
  echo
  read -r -p "Type 'yes' to continue: " confirm
  [[ "$confirm" == "yes" ]] || fail "rollback cancelled"
}

write_summary() {
  local target_version="$1"
  local post_plan_dir="$2"
  local summary_file="$ARTIFACT_DIR/summary.md"
  cat > "$summary_file" <<EOF
# Terraform Rollback Summary

- Environment: $ENVIRONMENT
- Dry Run: $DRY_RUN
- Backend Bucket: $STATE_BUCKET
- Backend Key: $STATE_KEY
- Target Version: $target_version
- Artifacts: $ARTIFACT_DIR
- Post-Rollback Plan Dir: $post_plan_dir
EOF
}

if [[ $# -lt 1 ]]; then
  usage >&2
  exit 1
fi

ENVIRONMENT="$1"
shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --steps-back)
      STEPS_BACK="$2"
      shift 2
      ;;
    --artifact-dir)
      ARTIFACT_DIR="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN="true"
      shift
      ;;
    --yes)
      AUTO_APPROVE="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown option: $1"
      ;;
  esac
done

[[ "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]] || fail "environment must be one of: dev, staging, prod"
[[ "$STEPS_BACK" =~ ^[0-9]+$ ]] || fail "--steps-back must be a positive integer"
(( STEPS_BACK >= 1 )) || fail "--steps-back must be >= 1"

check_prerequisites

TARGET_DIR="$(terraform_target_dir)"
[[ -d "$TARGET_DIR" ]] || fail "missing Terraform environment directory: $TARGET_DIR"

STATE_BUCKET="$(state_backend_bucket)"
STATE_KEY="$(state_backend_key)"
[[ -n "$STATE_BUCKET" ]] || fail "could not determine backend bucket from $TARGET_DIR/main.tf"
[[ -n "$STATE_KEY" ]] || fail "could not determine backend key from $TARGET_DIR/main.tf"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT_DIR="${ARTIFACT_DIR:-$ROOT_DIR/output/rollback/terraform/$ENVIRONMENT/$TIMESTAMP}"
mkdir -p "$ARTIFACT_DIR"

TARGET_STATE_FILE="$ARTIFACT_DIR/target.tfstate"
CURRENT_STATE_FILE="$ARTIFACT_DIR/current.tfstate"
POST_PLAN_DIR="$ARTIFACT_DIR/post-rollback-plan"

log_info "Available rollback candidates for $ENVIRONMENT:"
list_state_versions "$STATE_BUCKET" "$STATE_KEY"

TARGET_VERSION_ID="$(get_target_version_id "$STATE_BUCKET" "$STATE_KEY" "$STEPS_BACK")"
[[ -n "$TARGET_VERSION_ID" && "$TARGET_VERSION_ID" != "None" ]] || fail "no backend object version found $STEPS_BACK step(s) back"

download_state_version "$STATE_BUCKET" "$STATE_KEY" "$TARGET_VERSION_ID" "$TARGET_STATE_FILE"
sha256sum "$TARGET_STATE_FILE" > "$ARTIFACT_DIR/target.tfstate.sha256"
confirm_rollback "$TARGET_VERSION_ID"

if [[ "$DRY_RUN" == "true" ]]; then
  write_summary "$TARGET_VERSION_ID" "not-run"
  log_info "Dry run complete. Downloaded target state to $TARGET_STATE_FILE"
  exit 0
fi

run_init
pull_current_state "$CURRENT_STATE_FILE"
sha256sum "$CURRENT_STATE_FILE" > "$ARTIFACT_DIR/current.tfstate.sha256"

log_warn "Pushing previous state version into the backend"
push_state "$TARGET_STATE_FILE"

log_info "Generating reviewed follow-up plan via infra/scripts/terraform-run.sh"
"$ROOT_DIR/infra/scripts/terraform-run.sh" plan "$TARGET_DIR" "$POST_PLAN_DIR"
write_summary "$TARGET_VERSION_ID" "$POST_PLAN_DIR"

log_info "Rollback state push completed."
log_warn "Review $POST_PLAN_DIR/plan.txt before approving any corrective apply."
