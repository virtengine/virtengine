#!/bin/bash
# VirtEngine VEID Pipeline - Build Script
#
# Builds the VEID pipeline container, validates a production model bundle,
# and publishes deterministic release-manifest evidence.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${REPO_ROOT}/_build"

RAW_VERSION="${1:-1.0.0}"
MODEL_VERSION="${RAW_VERSION#v}"
RELEASE_VERSION="v${MODEL_VERSION}"
BUILD_PROFILE="${VEID_BUILD_PROFILE:-production}"
OUTPUT_DIR="${VEID_OUTPUT_DIR:-${REPO_ROOT}/_build/veid-pipeline-output}"
MODEL_BUNDLE_DIR="${VEID_MODEL_BUNDLE_DIR:-${REPO_ROOT}/models/trust_score/${RELEASE_VERSION}}"
MODEL_NAME="${VEID_MODEL_NAME:-trust_score}"
MODEL_CONFIG="${VEID_MODEL_CONFIG:-${REPO_ROOT}/ml/training/configs/trust_score_v1.yaml}"
MODEL_CARD="${VEID_MODEL_CARD:-${REPO_ROOT}/models/trust_score/MODEL_CARD.md}"
MODEL_PROVENANCE="${VEID_MODEL_PROVENANCE:-${MODEL_BUNDLE_DIR}/model_provenance.json}"
SKIP_DOCKER_BUILD="${VEID_SKIP_DOCKER_BUILD:-false}"
PYTHON_BIN="${VEID_PYTHON:-}"
SOURCE_REVISION=""
SOURCE_DESCRIBE=""
BUNDLE_DISPLAY_PATH=""
CONFIG_DISPLAY_PATH=""
MODEL_CARD_DISPLAY_PATH=""

REGISTRY="${VEID_REGISTRY:-ghcr.io/virtengine}"
IMAGE_NAME="veid-pipeline"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${RELEASE_VERSION}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."

    if [ -z "${PYTHON_BIN}" ]; then
        if [ -n "${MSYSTEM:-}" ] && command -v python >/dev/null 2>&1; then
            PYTHON_BIN="python"
        else
            PYTHON_BIN="python3"
        fi
    fi
    if ! command -v "${PYTHON_BIN}" >/dev/null 2>&1; then
        log_error "Python is required: ${PYTHON_BIN}"
        exit 1
    fi

    if ! command -v git >/dev/null 2>&1; then
        log_error "git is required"
        exit 1
    fi

    if ! command -v sha256sum >/dev/null 2>&1; then
        if ! command -v shasum >/dev/null 2>&1; then
            log_error "sha256sum or shasum is required"
            exit 1
        fi
        SHA256_CMD=(shasum -a 256)
    else
        SHA256_CMD=(sha256sum)
    fi

    if [ "${SKIP_DOCKER_BUILD}" != "true" ] && ! command -v docker >/dev/null 2>&1; then
        log_error "docker is required unless VEID_SKIP_DOCKER_BUILD=true"
        exit 1
    fi

    log_success "Dependencies available"
}

setup_output_dir() {
    log_info "Preparing output directory..."
    mkdir -p "${OUTPUT_DIR}"
    rm -f "${OUTPUT_DIR}"/*
    log_success "Output directory ready: ${OUTPUT_DIR}"
}

compute_file_hash() {
    local file="$1"
    "${SHA256_CMD[@]}" "${file}" | awk '{print $1}'
}

write_sha_file() {
    local source_file="$1"
    local label="$2"
    local output_file="$3"
    printf '%s  %s\n' "$(compute_file_hash "${source_file}")" "${label}" > "${output_file}"
}

portable_path() {
    local raw_path="$1"
    "${PYTHON_BIN}" - "${REPO_ROOT}" "${raw_path}" <<'PY'
from pathlib import Path
import sys

repo_root = Path(sys.argv[1]).resolve()
raw_path = Path(sys.argv[2])
candidate = raw_path if raw_path.is_absolute() else (repo_root / raw_path)
candidate = candidate.resolve()

try:
    print(candidate.relative_to(repo_root).as_posix())
except ValueError:
    print(candidate.as_posix())
PY
}

shell_path() {
    local raw_path="$1"
    if [[ "${raw_path}" =~ ^([A-Za-z]):[\\/](.*)$ ]]; then
        if command -v cygpath >/dev/null 2>&1; then
            cygpath -u "${raw_path}"
            return
        fi
        local drive_letter="${BASH_REMATCH[1],,}"
        local remainder="${BASH_REMATCH[2]//\\//}"
        if [ -d "/${drive_letter}" ]; then
            printf '/%s/%s\n' "${drive_letter}" "${remainder}"
            return
        fi
        printf '%s\n' "${raw_path}"
        return
    fi
    printf '%s\n' "${raw_path}"
}

normalize_input_paths() {
    MODEL_BUNDLE_DIR="$(shell_path "${MODEL_BUNDLE_DIR}")"
    OUTPUT_DIR="$(shell_path "${OUTPUT_DIR}")"
    MODEL_CONFIG="$(shell_path "${MODEL_CONFIG}")"
    MODEL_CARD="$(shell_path "${MODEL_CARD}")"
    MODEL_PROVENANCE="$(shell_path "${MODEL_PROVENANCE}")"
}

validate_build_inputs() {
    case "${BUILD_PROFILE}" in
        production|fixture_only) ;;
        *)
            log_error "Build profile must be production or fixture_only: ${BUILD_PROFILE}"
            exit 1
            ;;
    esac

    "${PYTHON_BIN}" - "${REPO_ROOT}" "${OUTPUT_DIR}" "${MODEL_BUNDLE_DIR}" "${MODEL_CONFIG}" "${MODEL_CARD}" "${MODEL_PROVENANCE}" <<'PY'
from pathlib import Path
import sys

repo_root = Path(sys.argv[1]).resolve()

def resolve(raw: str) -> Path:
    path = Path(raw)
    return (path if path.is_absolute() else repo_root / path).resolve()

output_dir = resolve(sys.argv[2])
bundle_dir = resolve(sys.argv[3])
protected = {
    "training config": resolve(sys.argv[4]),
    "model card": resolve(sys.argv[5]),
    "model provenance": resolve(sys.argv[6]),
}

if output_dir == bundle_dir or output_dir in bundle_dir.parents or bundle_dir in output_dir.parents:
    raise SystemExit("output directory and model bundle directory must be disjoint")

for label, path in protected.items():
    if path == output_dir or output_dir in path.parents:
        raise SystemExit(f"{label} must not be beneath output directory")
PY
}

require_bundle() {
    if [ ! -d "${MODEL_BUNDLE_DIR}" ]; then
        log_error "Model bundle directory not found: ${MODEL_BUNDLE_DIR}"
        exit 1
    fi
    if [ ! -d "${MODEL_BUNDLE_DIR}/model" ]; then
        log_error "Model bundle is missing model/: ${MODEL_BUNDLE_DIR}/model"
        exit 1
    fi

    for artifact in manifest.json export_metadata.json MODEL_HASH.txt model_frozen.pb; do
        if [ ! -f "${MODEL_BUNDLE_DIR}/${artifact}" ]; then
            log_error "Model bundle is missing ${artifact}: ${MODEL_BUNDLE_DIR}/${artifact}"
            exit 1
        fi
    done

    if [ ! -f "${MODEL_CONFIG}" ]; then
        log_error "Training config not found: ${MODEL_CONFIG}"
        exit 1
    fi

    if [ ! -f "${MODEL_CARD}" ]; then
        log_error "Model card not found: ${MODEL_CARD}"
        exit 1
    fi

    if [ ! -f "${MODEL_PROVENANCE}" ]; then
        log_error "Model provenance not found: ${MODEL_PROVENANCE}"
        exit 1
    fi

    if [ "${BUILD_PROFILE}" = "production" ] && [ "$(basename "${MODEL_BUNDLE_DIR}")" != "${RELEASE_VERSION}" ]; then
        log_error "Production bundle directory must end with ${RELEASE_VERSION}: ${MODEL_BUNDLE_DIR}"
        exit 1
    fi

    if grep -R -n -i -E 'placeholder|pending|tbd|not published yet|<path>|sha256:placeholder' \
        "${MODEL_BUNDLE_DIR}" "${MODEL_CARD}" "${MODEL_CONFIG}" "${MODEL_PROVENANCE}" >/dev/null 2>&1; then
        log_error "Placeholder content detected in model bundle, config, model card, or provenance"
        exit 1
    fi
}

prepare_release_context() {
    SOURCE_REVISION="$(git -C "${REPO_ROOT}" rev-parse HEAD)"
    SOURCE_DESCRIBE="$(git -C "${REPO_ROOT}" describe --tags --always --dirty --match 'v*' 2>/dev/null || git -C "${REPO_ROOT}" rev-parse --short=12 HEAD)"
    BUNDLE_DISPLAY_PATH="$(portable_path "${MODEL_BUNDLE_DIR}")"
    CONFIG_DISPLAY_PATH="$(portable_path "${MODEL_CONFIG}")"
    MODEL_CARD_DISPLAY_PATH="$(portable_path "${MODEL_CARD}")"
}

preflight_release_manifest() {
    log_info "Validating release manifest inputs..."

    "${PYTHON_BIN}" "${REPO_ROOT}/ml/training/model/release_manifest.py" \
        --bundle-dir "${MODEL_BUNDLE_DIR}" \
        --output "${OUTPUT_DIR}/release_manifest.json" \
        --model-name "${MODEL_NAME}" \
        --version "${RELEASE_VERSION}" \
        --profile "${BUILD_PROFILE}" \
        --config-path "${MODEL_CONFIG}" \
        --model-card-path "${MODEL_CARD}" \
        --provenance-path "${MODEL_PROVENANCE}" \
        --source-revision "${SOURCE_REVISION}" \
        --source-describe "${SOURCE_DESCRIBE}" \
        --bundle-display-path "${BUNDLE_DISPLAY_PATH}" \
        --config-display-path "${CONFIG_DISPLAY_PATH}" \
        --model-card-display-path "${MODEL_CARD_DISPLAY_PATH}" \
        --validate-only

    log_success "Release manifest inputs validated"
}

generate_release_manifest() {
    log_info "Generating deterministic release manifest..."

    OUTPUT_MANIFEST_PATH="${OUTPUT_DIR}/release_manifest.json"

    "${PYTHON_BIN}" "${REPO_ROOT}/ml/training/model/release_manifest.py" \
        --bundle-dir "${MODEL_BUNDLE_DIR}" \
        --output "${OUTPUT_MANIFEST_PATH}" \
        --model-name "${MODEL_NAME}" \
        --version "${RELEASE_VERSION}" \
        --profile "${BUILD_PROFILE}" \
        --config-path "${MODEL_CONFIG}" \
        --model-card-path "${MODEL_CARD}" \
        --provenance-path "${MODEL_PROVENANCE}" \
        --source-revision "${SOURCE_REVISION}" \
        --source-describe "${SOURCE_DESCRIBE}" \
        --bundle-display-path "${BUNDLE_DISPLAY_PATH}" \
        --config-display-path "${CONFIG_DISPLAY_PATH}" \
        --model-card-display-path "${MODEL_CARD_DISPLAY_PATH}"

    write_sha_file "${OUTPUT_MANIFEST_PATH}" "release_manifest.json" "${OUTPUT_DIR}/release_manifest.json.sha256"

    log_success "Release manifest written: ${OUTPUT_MANIFEST_PATH}"
}

generate_source_provenance() {
    log_info "Publishing source provenance..."

    "${PYTHON_BIN}" - "${OUTPUT_DIR}" "${BUILD_PROFILE}" "${RELEASE_VERSION}" "${MODEL_NAME}" "${SOURCE_REVISION}" "${SOURCE_DESCRIBE}" "${BUNDLE_DISPLAY_PATH}" "${CONFIG_DISPLAY_PATH}" "${MODEL_CARD_DISPLAY_PATH}" <<'PY'
import json
import sys
from pathlib import Path

output_dir = Path(sys.argv[1])
build_profile = sys.argv[2]
release_version = sys.argv[3]
model_name = sys.argv[4]
source_revision = sys.argv[5]
source_describe = sys.argv[6]
bundle_path = sys.argv[7]
config_path = sys.argv[8]
model_card_path = sys.argv[9]

release_manifest = json.loads((output_dir / "release_manifest.json").read_text(encoding="utf-8"))
release_manifest_sha = (output_dir / "release_manifest.json.sha256").read_text(encoding="utf-8").split()[0]
payload = {
    "schema_version": "veid.source.provenance/v1",
    "profile": build_profile,
    "release_version": release_version,
    "model_name": model_name,
    "source_revision": source_revision,
    "source_revision_short": source_revision[:12],
    "source_describe": source_describe,
    "bundle_path": bundle_path,
    "config_path": config_path,
    "model_card_path": model_card_path,
    "build_script_path": "_build/build-veid-pipeline.sh",
    "release_manifest_path": "release_manifest.json",
    "release_manifest_sha256": release_manifest_sha,
    "bundle_artifact_index_sha256": release_manifest["source"]["bundle_artifact_index_sha256"],
}

(output_dir / "source_provenance.json").write_text(
    json.dumps(payload, indent=2, sort_keys=True) + "\n",
    encoding="utf-8",
)
PY

    write_sha_file "${OUTPUT_DIR}/source_provenance.json" "source_provenance.json" "${OUTPUT_DIR}/source_provenance.json.sha256"
    log_success "Source provenance written"
}

generate_model_signature() {
    log_info "Publishing model signature contract..."

    "${PYTHON_BIN}" - "${OUTPUT_DIR}" "${BUILD_PROFILE}" <<'PY'
import json
import sys
from pathlib import Path

output_dir = Path(sys.argv[1])
build_profile = sys.argv[2]
release_manifest = json.loads((output_dir / "release_manifest.json").read_text(encoding="utf-8"))
release_manifest_sha = (output_dir / "release_manifest.json.sha256").read_text(encoding="utf-8").split()[0]
model = release_manifest["model"]

payload = {
    "schema_version": "veid.model.signature/v1",
    "profile": build_profile,
    "model_name": model["name"],
    "model_version": model["version"],
    "runtime_hash": model["runtime_hash"],
    "signature_name": model["signature_name"],
    "input_signature": model["input_signature"],
    "output_signature": model["output_signature"],
    "release_manifest_path": "release_manifest.json",
    "release_manifest_sha256": release_manifest_sha,
}

(output_dir / "model_signature.json").write_text(
    json.dumps(payload, indent=2, sort_keys=True) + "\n",
    encoding="utf-8",
)
PY

    write_sha_file "${OUTPUT_DIR}/model_signature.json" "model_signature.json" "${OUTPUT_DIR}/model_signature.json.sha256"
    log_success "Model signature contract written"
}

build_image() {
    if [ "${SKIP_DOCKER_BUILD}" = "true" ]; then
        log_warn "Skipping Docker build; manifest publication only"
        IMAGE_ID=""
        REPO_DIGEST=""
        printf '%s\n' "${IMAGE_ID}" > "${OUTPUT_DIR}/image_id.txt"
        printf '%s\n' "${REPO_DIGEST}" > "${OUTPUT_DIR}/repo_digest.txt"
        return
    fi

    log_info "Building Docker image: ${FULL_IMAGE}"
    (
        cd "${REPO_ROOT}"
        DOCKER_BUILDKIT=1 docker build \
            --no-cache \
            --file "${BUILD_DIR}/Dockerfile.veid-pipeline" \
            --tag "${FULL_IMAGE}" \
            --label "org.virtengine.pipeline.version=${RELEASE_VERSION}" \
            . 2>&1 | tee "${OUTPUT_DIR}/build.log"
    )

    IMAGE_ID="$(docker inspect --format='{{.Id}}' "${FULL_IMAGE}")"
    REPO_DIGEST="$(docker inspect --format='{{index .RepoDigests 0}}' "${FULL_IMAGE}" 2>/dev/null || true)"

    printf '%s\n' "${IMAGE_ID}" > "${OUTPUT_DIR}/image_id.txt"
    printf '%s\n' "${REPO_DIGEST}" > "${OUTPUT_DIR}/repo_digest.txt"

    log_success "Image built successfully"
    log_info "Image ID: ${IMAGE_ID:-<none>}"
    if [ -n "${REPO_DIGEST}" ]; then
        log_info "Repo digest: ${REPO_DIGEST}"
    fi
}

generate_pipeline_version() {
    log_info "Generating pipeline metadata..."

    "${PYTHON_BIN}" - "${OUTPUT_DIR}" "${FULL_IMAGE}" "${RELEASE_VERSION}" "${BUILD_PROFILE}" "${BUNDLE_DISPLAY_PATH}" "${SOURCE_REVISION}" "${SOURCE_DESCRIBE}" <<'PY'
import json
import sys
from pathlib import Path

output_dir = Path(sys.argv[1])
full_image = sys.argv[2]
release_version = sys.argv[3]
build_profile = sys.argv[4]
bundle_dir = sys.argv[5]
source_revision = sys.argv[6]
source_describe = sys.argv[7]

image_id = (output_dir / "image_id.txt").read_text(encoding="utf-8").strip()
repo_digest = (output_dir / "repo_digest.txt").read_text(encoding="utf-8").strip()
release_manifest_path = "release_manifest.json"
release_manifest_sha = (output_dir / "release_manifest.json.sha256").read_text(encoding="utf-8").split()[0]

payload = {
    "schema_version": "veid.pipeline.version/v1",
    "profile": build_profile,
    "version": release_version,
    "image_ref": full_image,
    "image_id": image_id,
    "repo_digest": repo_digest,
    "model_bundle_dir": bundle_dir,
    "release_manifest_path": release_manifest_path,
    "release_manifest_sha256": release_manifest_sha,
    "source_provenance_path": "source_provenance.json",
    "model_signature_path": "model_signature.json",
    "bundle_checksums_path": "bundle_checksums.txt",
    "source_revision": source_revision,
    "source_describe": source_describe,
}

(output_dir / "pipeline_version.json").write_text(
    json.dumps(payload, indent=2, sort_keys=True) + "\n",
    encoding="utf-8",
)
PY

    write_sha_file "${OUTPUT_DIR}/pipeline_version.json" "pipeline_version.json" "${OUTPUT_DIR}/pipeline_version.json.sha256"
    log_success "Pipeline metadata written"
}

publish_checksums() {
    log_info "Publishing checksum index..."

    local checksum_file="${OUTPUT_DIR}/bundle_checksums.txt"
    : > "${checksum_file}"

    while IFS= read -r path; do
        local rel
        rel="${path#${MODEL_BUNDLE_DIR}/}"
        printf '%s  %s\n' "$(compute_file_hash "${path}")" "${rel}" >> "${checksum_file}"
    done < <(LC_ALL=C find "${MODEL_BUNDLE_DIR}" -type f | LC_ALL=C sort)

    for published_artifact in \
        release_manifest.json \
        release_manifest.json.sha256 \
        source_provenance.json \
        source_provenance.json.sha256 \
        model_signature.json \
        model_signature.json.sha256 \
        pipeline_version.json \
        pipeline_version.json.sha256 \
        image_id.txt \
        repo_digest.txt; do
        if [ -f "${OUTPUT_DIR}/${published_artifact}" ]; then
            printf '%s  %s\n' "$(compute_file_hash "${OUTPUT_DIR}/${published_artifact}")" "${published_artifact}" >> "${checksum_file}"
        fi
    done

    write_sha_file "${checksum_file}" "bundle_checksums.txt" "${OUTPUT_DIR}/bundle_checksums.txt.sha256"

    log_success "Checksum index written: ${checksum_file}"
}

verify_release_bundle() {
    log_info "Verifying release bundle consistency..."

    "${PYTHON_BIN}" - "${MODEL_BUNDLE_DIR}" "${OUTPUT_DIR}" <<'PY'
import hashlib
import json
import sys
from pathlib import Path

bundle_dir = Path(sys.argv[1])
output_dir = Path(sys.argv[2])

release_manifest = json.loads((output_dir / "release_manifest.json").read_text(encoding="utf-8"))
source_provenance = json.loads((output_dir / "source_provenance.json").read_text(encoding="utf-8"))
model_signature = json.loads((output_dir / "model_signature.json").read_text(encoding="utf-8"))
pipeline_version = json.loads((output_dir / "pipeline_version.json").read_text(encoding="utf-8"))
model = release_manifest["model"]
artifact_paths = {artifact["path"] for artifact in release_manifest["artifacts"]}

required = {
    "MODEL_HASH.txt",
    "export_metadata.json",
    "manifest.json",
    "model_provenance.json",
    "model_frozen.pb",
    "model/saved_model.pb",
}
missing = sorted(required - artifact_paths)
if missing:
    raise SystemExit(f"release manifest missing required artifacts: {missing}")

if model["version"] == "" or model["runtime_hash"] == "" or model["frozen_graph_hash"] == "":
    raise SystemExit("release manifest contains empty model identity fields")

provenance = release_manifest.get("provenance", {})
if provenance.get("path") != "model_provenance.json":
    raise SystemExit("release manifest provenance path must be model_provenance.json")

if release_manifest["source"]["bundle_path"] != pipeline_version["model_bundle_dir"]:
    raise SystemExit("release manifest bundle path does not match pipeline metadata")

if source_provenance["bundle_path"] != release_manifest["source"]["bundle_path"]:
    raise SystemExit("source provenance bundle path does not match release manifest")

if source_provenance["config_path"] != release_manifest["source"]["config_path"]:
    raise SystemExit("source provenance config path does not match release manifest")

if source_provenance["model_card_path"] != release_manifest["source"]["model_card_path"]:
    raise SystemExit("source provenance model card path does not match release manifest")

if source_provenance["source_revision"] != pipeline_version["source_revision"]:
    raise SystemExit("source revision mismatch between provenance and pipeline metadata")

if source_provenance["source_describe"] != pipeline_version["source_describe"]:
    raise SystemExit("source describe mismatch between provenance and pipeline metadata")

if source_provenance["source_describe"] != release_manifest["source"]["describe"]:
    raise SystemExit("source describe mismatch between provenance and release manifest")

if model_signature["model_version"] != model["version"]:
    raise SystemExit("model signature metadata version mismatch")

if model_signature["runtime_hash"] != model["runtime_hash"]:
    raise SystemExit("model signature metadata runtime hash mismatch")

if model_signature["signature_name"] != model["signature_name"]:
    raise SystemExit("model signature metadata signature_name mismatch")

def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()

def parse_sha_file(path: Path, expected_label: str) -> str:
    parts = path.read_text(encoding="utf-8").strip().split()
    if len(parts) != 2:
        raise SystemExit(f"{path.name} must contain '<sha256>  <file>'")
    digest, label = parts
    if label != expected_label:
        raise SystemExit(f"{path.name} points at {label}, expected {expected_label}")
    if len(digest) != 64:
        raise SystemExit(f"{path.name} contains an invalid digest")
    return digest

release_manifest_sha = parse_sha_file(output_dir / "release_manifest.json.sha256", "release_manifest.json")
if source_provenance["release_manifest_sha256"] != release_manifest_sha:
    raise SystemExit("source provenance release manifest sha mismatch")

if source_provenance["bundle_artifact_index_sha256"] != release_manifest["source"]["bundle_artifact_index_sha256"]:
    raise SystemExit("source provenance bundle artifact index mismatch")

required_output = {
    "release_manifest.json",
    "release_manifest.json.sha256",
    "source_provenance.json",
    "source_provenance.json.sha256",
    "model_signature.json",
    "model_signature.json.sha256",
    "pipeline_version.json",
    "pipeline_version.json.sha256",
    "bundle_checksums.txt",
    "bundle_checksums.txt.sha256",
    "image_id.txt",
    "repo_digest.txt",
}

for output_name in required_output:
    if not (output_dir / output_name).exists():
        raise SystemExit(f"missing published output artifact: {output_name}")

for sha_name, target_name in (
    ("release_manifest.json.sha256", "release_manifest.json"),
    ("source_provenance.json.sha256", "source_provenance.json"),
    ("model_signature.json.sha256", "model_signature.json"),
    ("pipeline_version.json.sha256", "pipeline_version.json"),
    ("bundle_checksums.txt.sha256", "bundle_checksums.txt"),
):
    digest = parse_sha_file(output_dir / sha_name, target_name)
    actual = sha256_file(output_dir / target_name)
    if digest != actual:
        raise SystemExit(f"{sha_name} does not match {target_name}")

checksum_entries = {}
for line in (output_dir / "bundle_checksums.txt").read_text(encoding="utf-8").splitlines():
    if not line.strip():
        continue
    digest, label = line.split(None, 1)
    checksum_entries[label.strip()] = digest.strip()

required_checksum_entries = {
    "release_manifest.json",
    "release_manifest.json.sha256",
    "source_provenance.json",
    "source_provenance.json.sha256",
    "model_signature.json",
    "model_signature.json.sha256",
    "pipeline_version.json",
    "pipeline_version.json.sha256",
    "image_id.txt",
    "repo_digest.txt",
}

missing_checksum_entries = sorted(required_checksum_entries - set(checksum_entries))
if missing_checksum_entries:
    raise SystemExit(f"bundle_checksums.txt is missing published entries: {missing_checksum_entries}")

for label in required_checksum_entries:
    actual = sha256_file(output_dir / label)
    if checksum_entries[label] != actual:
        raise SystemExit(f"bundle_checksums.txt entry mismatch for {label}")

for artifact in release_manifest["artifacts"]:
    label = artifact["path"]
    actual = sha256_file(bundle_dir / label)
    if checksum_entries.get(label) != actual:
        raise SystemExit(f"bundle_checksums.txt entry mismatch for bundle artifact {label}")
    if label == "model_provenance.json" and provenance.get("sha256") != actual:
        raise SystemExit("release manifest provenance digest does not match its artifact")
PY

    log_success "Release bundle verification passed"
}

verify_image() {
    if [ "${SKIP_DOCKER_BUILD}" = "true" ]; then
        return
    fi

    log_info "Running basic image verification..."

    docker run --rm "${FULL_IMAGE}" python3 -c "print('Image verification passed')" >/dev/null

    local tf_deterministic
    tf_deterministic="$(docker run --rm "${FULL_IMAGE}" printenv TF_DETERMINISTIC_OPS)"
    if [ "${tf_deterministic}" != "1" ]; then
        log_error "TF_DETERMINISTIC_OPS is not set to 1"
        exit 1
    fi

    local cuda_visible
    cuda_visible="$(docker run --rm "${FULL_IMAGE}" printenv CUDA_VISIBLE_DEVICES)"
    if [ "${cuda_visible}" != "-1" ]; then
        log_error "CUDA_VISIBLE_DEVICES is not set to -1"
        exit 1
    fi

    log_success "Image verification passed"
}

print_summary() {
    echo
    echo "=========================================="
    echo "VirtEngine VEID Pipeline Build Complete"
    echo "=========================================="
    echo
    echo "Version:             ${RELEASE_VERSION}"
    echo "Profile:             ${BUILD_PROFILE}"
    echo "Bundle:              ${MODEL_BUNDLE_DIR}"
    echo "Image:               ${FULL_IMAGE}"
    echo "Image Hash:          $(cat "${OUTPUT_DIR}/image_id.txt")"
    echo "Release Manifest:    ${OUTPUT_DIR}/release_manifest.json"
    echo "Pipeline Metadata:   ${OUTPUT_DIR}/pipeline_version.json"
    echo "Checksum Index:      ${OUTPUT_DIR}/bundle_checksums.txt"
    echo
}

main() {
    log_info "Building VirtEngine VEID Pipeline ${RELEASE_VERSION}"

    check_dependencies
    normalize_input_paths
    validate_build_inputs
    require_bundle
    prepare_release_context
    preflight_release_manifest
    setup_output_dir
    generate_release_manifest
    generate_source_provenance
    generate_model_signature
    build_image
    generate_pipeline_version
    publish_checksums
    verify_release_bundle
    verify_image
    print_summary
}

main "$@"
