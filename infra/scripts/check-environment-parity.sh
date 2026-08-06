#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXPECTED_TF_VERSION="1.6.6"

log() {
  echo "[infra-parity] $*"
}

fail() {
  echo "[infra-parity] ERROR: $*" >&2
  exit 1
}

require_file() {
  local file="$1"
  [[ -f "$file" ]] || fail "missing required file: ${file#$ROOT_DIR/}"
}

assert_contains() {
  local file="$1"
  local pattern="$2"
  grep -Fq "$pattern" "$file" || fail "${file#$ROOT_DIR/} is missing expected content: $pattern"
}

assert_not_contains() {
  local file="$1"
  local pattern="$2"
  if grep -Fq "$pattern" "$file"; then
    fail "${file#$ROOT_DIR/} still contains forbidden content: $pattern"
  fi
}

for env_name in dev staging prod; do
  env_dir="$ROOT_DIR/infra/terraform/environments/$env_name"
  require_file "$env_dir/main.tf"
  require_file "$env_dir/variables.tf"
  require_file "$env_dir/outputs.tf"
  require_file "$env_dir/terraform.tfvars"
  require_file "$env_dir/terragrunt.hcl"
  require_file "$env_dir/env.hcl"
done

for region in us-east-1 eu-west-1 ap-southeast-1; do
  region_dir="$ROOT_DIR/infra/terraform/regions/$region"
  require_file "$region_dir/main.tf"
  require_file "$region_dir/variables.tf"
  require_file "$region_dir/outputs.tf"
  require_file "$region_dir/terraform.tfvars"
done

for workflow in \
  "$ROOT_DIR/.github/workflows/infrastructure.yaml" \
  "$ROOT_DIR/.github/workflows/multi-region-deploy.yaml"; do
  assert_contains "$workflow" "TF_VERSION: '$EXPECTED_TF_VERSION'"
done

assert_not_contains "$ROOT_DIR/infra/terraform/global/main.tf" "ffffffffffffffffffffffffffffffffffffffff"
assert_not_contains "$ROOT_DIR/infra/terraform/modules/scaling/main.tf" "placeholder.elb."
assert_not_contains "$ROOT_DIR/infra/terraform/environments/staging/terragrunt.hcl" "0.0.0.0/0"
assert_not_contains "$ROOT_DIR/infra/terraform/environments/prod/terragrunt.hcl" "0.0.0.0/0"

assert_contains "$ROOT_DIR/infra/terraform/global/variables.tf" "repo:virtengine/virtengine:environment:infra-prod"
assert_contains "$ROOT_DIR/infra/terraform/global/main.tf" "data \"tls_certificate\" \"github_actions\""

log "environment parity checks passed"
