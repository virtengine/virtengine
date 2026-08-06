#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  terraform-run.sh plan <target-dir> <artifact-dir>
  terraform-run.sh apply <target-dir> <artifact-dir>
  terraform-run.sh drift <target-dir> <artifact-dir>
EOF
}

fail() {
  echo "[terraform-run] ERROR: $*" >&2
  exit 1
}

if [[ $# -ne 3 ]]; then
  usage >&2
  exit 1
fi

MODE="$1"
TARGET_DIR="$2"
ARTIFACT_DIR="$3"

[[ -d "$TARGET_DIR" ]] || fail "target directory does not exist: $TARGET_DIR"

mkdir -p "$ARTIFACT_DIR"
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"
ARTIFACT_DIR="$(cd "$ARTIFACT_DIR" && pwd)"

PLAN_FILE="$ARTIFACT_DIR/tfplan.binary"
PLAN_JSON="$ARTIFACT_DIR/plan.json"
PLAN_TEXT="$ARTIFACT_DIR/plan.txt"
PLAN_SHA="$ARTIFACT_DIR/tfplan.sha256"
MANIFEST_FILE="$ARTIFACT_DIR/manifest.env"

run_init() {
  if [[ -f "$TARGET_DIR/terragrunt.hcl" ]]; then
    (cd "$TARGET_DIR" && terragrunt init -input=false -no-color)
  else
    (cd "$TARGET_DIR" && terraform init -input=false -no-color)
  fi
}

run_plan() {
  if [[ -f "$TARGET_DIR/terragrunt.hcl" ]]; then
    (cd "$TARGET_DIR" && terragrunt plan -out="$PLAN_FILE" -input=false -no-color) | tee "$PLAN_TEXT"
  else
    (cd "$TARGET_DIR" && terraform plan -out="$PLAN_FILE" -input=false -no-color) | tee "$PLAN_TEXT"
  fi
}

run_drift() {
  local rc
  set +e
  if [[ -f "$TARGET_DIR/terragrunt.hcl" ]]; then
    (cd "$TARGET_DIR" && terragrunt plan -detailed-exitcode -out="$PLAN_FILE" -input=false -no-color) | tee "$PLAN_TEXT"
    rc=${PIPESTATUS[0]}
  else
    (cd "$TARGET_DIR" && terraform plan -detailed-exitcode -out="$PLAN_FILE" -input=false -no-color) | tee "$PLAN_TEXT"
    rc=${PIPESTATUS[0]}
  fi
  set -e

  case "$rc" in
    0)
      echo "status=clean" > "$MANIFEST_FILE"
      ;;
    2)
      terraform -chdir="$TARGET_DIR" show -json "$PLAN_FILE" > "$PLAN_JSON"
      sha256sum "$PLAN_FILE" | tee "$PLAN_SHA"
      {
        echo "status=drift"
        echo "plan_file=$(basename "$PLAN_FILE")"
      } > "$MANIFEST_FILE"
      return 2
      ;;
    *)
      return "$rc"
      ;;
  esac
}

write_plan_artifacts() {
  terraform -chdir="$TARGET_DIR" show -json "$PLAN_FILE" > "$PLAN_JSON"
  sha256sum "$PLAN_FILE" | tee "$PLAN_SHA"
  {
    echo "status=planned"
    echo "plan_file=$(basename "$PLAN_FILE")"
    echo "plan_sha256=$(cut -d' ' -f1 "$PLAN_SHA")"
    echo "target_dir=$TARGET_DIR"
  } > "$MANIFEST_FILE"
}

verify_plan() {
  [[ -f "$PLAN_FILE" ]] || fail "missing plan artifact: $PLAN_FILE"
  [[ -f "$PLAN_SHA" ]] || fail "missing plan checksum: $PLAN_SHA"
  (cd "$ARTIFACT_DIR" && sha256sum -c "$(basename "$PLAN_SHA")")
}

run_apply() {
  if [[ -f "$TARGET_DIR/terragrunt.hcl" ]]; then
    (cd "$TARGET_DIR" && terragrunt apply -input=false -no-color "$PLAN_FILE")
  else
    (cd "$TARGET_DIR" && terraform apply -input=false -no-color "$PLAN_FILE")
  fi
}

run_init

case "$MODE" in
  plan)
    run_plan
    write_plan_artifacts
    ;;
  apply)
    verify_plan
    run_apply
    ;;
  drift)
    run_drift
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
