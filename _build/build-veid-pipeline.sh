#!/bin/bash
# VirtEngine VEID Pipeline - Build Script
# VE-219: Deterministic identity verification runtime
#
# This script builds the deterministic VEID pipeline container and computes
# all necessary hashes for on-chain registration.
#
# Usage: ./build-veid-pipeline.sh [VERSION]
# Example: ./build-veid-pipeline.sh 1.0.0

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${REPO_ROOT}/_build"
ML_DIR="${REPO_ROOT}/ml"
OUTPUT_DIR="${REPO_ROOT}/_build/veid-pipeline-output"

# Image configuration
REGISTRY="${VEID_REGISTRY:-ghcr.io/virtengine}"
IMAGE_NAME="veid-pipeline"
VERSION="${1:-1.0.0}"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:v${VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Ensure required tools are available
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command -v sha256sum &> /dev/null; then
        # macOS uses shasum
        if ! command -v shasum &> /dev/null; then
            log_error "sha256sum or shasum is required"
            exit 1
        fi
        SHA256_CMD="shasum -a 256"
    else
        SHA256_CMD="sha256sum"
    fi

    if ! command -v python3 &> /dev/null; then
        log_error "python3 is required to compute deterministic manifest hashes"
        exit 1
    fi
    
    log_success "All dependencies available"
}

# Create output directory
setup_output_dir() {
    log_info "Setting up output directory..."
    mkdir -p "${OUTPUT_DIR}"
    rm -f "${OUTPUT_DIR}"/*
    log_success "Output directory ready: ${OUTPUT_DIR}"
}

# Compute hash of a file
compute_file_hash() {
    local file="$1"
    ${SHA256_CMD} "${file}" | awk '{print $1}'
}

# Compute model weight hashes
compute_model_hashes() {
    log_info "Computing model weight hashes..."
    
    local models_dir="${ML_DIR}/models"
    local manifest_file="${OUTPUT_DIR}/model_manifest.json"
    local allow_placeholders="${VEID_ALLOW_PLACEHOLDER_HASHES:-false}"
    
    # Check if models directory exists (may be placeholder in development)
    if [ ! -d "${models_dir}" ]; then
        if [ "${allow_placeholders}" != "true" ]; then
            log_error "Models directory not found. Set VEID_ALLOW_PLACEHOLDER_HASHES=true to allow placeholders."
            exit 1
        fi

        log_warn "Models directory not found, creating placeholder manifest"
        cat > "${manifest_file}" << EOF
{
    "version": "1.0.0",
    "models": [
        {
            "name": "deepface_facenet512",
            "version": "1.0.0",
            "weights_hash": "sha256:placeholder_hash_for_deepface_facenet512_model_weights",
            "framework": "tensorflow",
            "purpose": "face_recognition"
        },
        {
            "name": "craft_text_detection",
            "version": "1.0.0",
            "weights_hash": "sha256:placeholder_hash_for_craft_text_detection_model_weights",
            "framework": "pytorch",
            "purpose": "text_detection"
        },
        {
            "name": "unet_face_extraction",
            "version": "1.0.0",
            "weights_hash": "sha256:placeholder_hash_for_unet_face_extraction_model_weights",
            "framework": "tensorflow",
            "purpose": "face_extraction"
        },
        {
            "name": "identity_scorer_v1",
            "version": "1.0.0",
            "weights_hash": "sha256:placeholder_hash_for_identity_scorer_model_weights",
            "framework": "tensorflow",
            "purpose": "identity_scoring"
        }
    ],
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    else
        log_info "Computing actual model hashes..."
        
        local deepface_hash=""
        local craft_hash=""
        local unet_hash=""
        local scorer_hash=""
        
        if [ -f "${models_dir}/deepface_facenet512.h5" ]; then
            deepface_hash=$(compute_file_hash "${models_dir}/deepface_facenet512.h5")
        fi
        
        if [ -f "${models_dir}/craft_mlt_25k.pth" ]; then
            craft_hash=$(compute_file_hash "${models_dir}/craft_mlt_25k.pth")
        fi
        
        if [ -f "${models_dir}/unet_face_extraction.h5" ]; then
            unet_hash=$(compute_file_hash "${models_dir}/unet_face_extraction.h5")
        fi
        
        if [ -f "${models_dir}/identity_scorer_v1.h5" ]; then
            scorer_hash=$(compute_file_hash "${models_dir}/identity_scorer_v1.h5")
        fi

        if [ "${allow_placeholders}" != "true" ]; then
            if [ -z "${deepface_hash}" ] || [ -z "${craft_hash}" ] || [ -z "${unet_hash}" ] || [ -z "${scorer_hash}" ]; then
                log_error "Missing model weights. Set VEID_ALLOW_PLACEHOLDER_HASHES=true to allow placeholders."
                exit 1
            fi
        fi

        cat > "${manifest_file}" << EOF
{
    "version": "1.0.0",
    "models": [
        {
            "name": "deepface_facenet512",
            "version": "1.0.0",
            "weights_hash": "sha256:${deepface_hash:-placeholder}",
            "framework": "tensorflow",
            "purpose": "face_recognition"
        },
        {
            "name": "craft_text_detection",
            "version": "1.0.0",
            "weights_hash": "sha256:${craft_hash:-placeholder}",
            "framework": "pytorch",
            "purpose": "text_detection"
        },
        {
            "name": "unet_face_extraction",
            "version": "1.0.0",
            "weights_hash": "sha256:${unet_hash:-placeholder}",
            "framework": "tensorflow",
            "purpose": "face_extraction"
        },
        {
            "name": "identity_scorer_v1",
            "version": "1.0.0",
            "weights_hash": "sha256:${scorer_hash:-placeholder}",
            "framework": "tensorflow",
            "purpose": "identity_scoring"
        }
    ],
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    fi
    
    # Compute manifest hash
    local manifest_hash
    manifest_hash=$(python3 - << 'PY' "${manifest_file}"
import hashlib
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as fh:
    data = json.load(fh)

version = data.get("version", "")
models = data.get("models", [])

model_by_name = {m["name"]: m for m in models}
names = sorted(model_by_name.keys())

h = hashlib.sha256()
h.update(version.encode())
for name in names:
    model = model_by_name[name]
    h.update(name.encode())
    h.update(model.get("version", "").encode())
    h.update(model.get("weights_hash", "").encode())
    config_hash = model.get("config_hash", "")
    if config_hash:
        h.update(config_hash.encode())
    h.update(model.get("framework", "").encode())

print(h.hexdigest())
PY
)
    echo "${manifest_hash}" > "${OUTPUT_DIR}/manifest_hash.txt"

    python3 - << 'PY' "${manifest_file}" "${manifest_hash}"
import json
import sys

path = sys.argv[1]
manifest_hash = sys.argv[2]
with open(path, "r", encoding="utf-8") as fh:
    data = json.load(fh)

data["manifest_hash"] = manifest_hash

with open(path, "w", encoding="utf-8") as fh:
    json.dump(data, fh, indent=4, sort_keys=False)
    fh.write("\n")
PY
    
    log_success "Model manifest created: ${manifest_file}"
    log_info "Manifest hash: ${manifest_hash}"
}

# Build the Docker image
build_image() {
    log_info "Building Docker image: ${FULL_IMAGE}"
    
    cd "${REPO_ROOT}"
    
    # Build with BuildKit for reproducibility
    DOCKER_BUILDKIT=1 docker build \
        --no-cache \
        --file "${BUILD_DIR}/Dockerfile.veid-pipeline" \
        --tag "${FULL_IMAGE}" \
        --label "org.virtengine.pipeline.version=${VERSION}" \
        --label "org.virtengine.build.timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        . 2>&1 | tee "${OUTPUT_DIR}/build.log"
    
    log_success "Image built successfully"
}

# Get image digest/hash
get_image_hash() {
    log_info "Computing image hash..."
    
    # Get the image digest
    local image_id
    image_id=$(docker inspect --format='{{.Id}}' "${FULL_IMAGE}")
    
    # Get the repo digest (if pushed)
    local repo_digest
    repo_digest=$(docker inspect --format='{{index .RepoDigests 0}}' "${FULL_IMAGE}" 2>/dev/null || echo "not-pushed")
    
    # Save hashes
    echo "${image_id}" > "${OUTPUT_DIR}/image_id.txt"
    echo "${repo_digest}" > "${OUTPUT_DIR}/repo_digest.txt"
    
    log_success "Image ID: ${image_id}"
    log_info "Repo digest: ${repo_digest}"
}

# Generate pipeline version registration info
generate_registration_info() {
    log_info "Generating pipeline version registration info..."
    
    local image_hash
    image_hash=$(cat "${OUTPUT_DIR}/image_id.txt")
    
    local manifest_hash
    manifest_hash=$(cat "${OUTPUT_DIR}/manifest_hash.txt")
    
    cat > "${OUTPUT_DIR}/pipeline_version.json" << EOF
{
    "version": "${VERSION}",
    "image_hash": "${image_hash}",
    "image_ref": "${FULL_IMAGE}",
    "model_manifest_hash": "${manifest_hash}",
    "status": "pending",
    "determinism_config": {
        "random_seed": 42,
        "force_cpu": true,
        "single_thread": true,
        "float_precision": 6,
        "tensorflow_deterministic": true,
        "disable_cudnn": true,
        "onnx_deterministic": true
    },
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    log_success "Pipeline version info saved to: ${OUTPUT_DIR}/pipeline_version.json"
}

# Run basic verification
verify_image() {
    log_info "Running basic image verification..."
    
    # Verify image can start
    docker run --rm "${FULL_IMAGE}" python3 -c "print('Image verification passed')" || {
        log_error "Image verification failed"
        exit 1
    }
    
    # Verify determinism settings
    local tf_deterministic
    tf_deterministic=$(docker run --rm "${FULL_IMAGE}" printenv TF_DETERMINISTIC_OPS)
    if [ "${tf_deterministic}" != "1" ]; then
        log_error "TF_DETERMINISTIC_OPS is not set to 1"
        exit 1
    fi
    
    local cuda_visible
    cuda_visible=$(docker run --rm "${FULL_IMAGE}" printenv CUDA_VISIBLE_DEVICES)
    if [ "${cuda_visible}" != "-1" ]; then
        log_error "CUDA_VISIBLE_DEVICES is not set to -1"
        exit 1
    fi
    
    log_success "Image verification passed"
}

# Print summary
print_summary() {
    echo ""
    echo "=========================================="
    echo "VirtEngine VEID Pipeline Build Complete"
    echo "=========================================="
    echo ""
    echo "Version:        ${VERSION}"
    echo "Image:          ${FULL_IMAGE}"
    echo "Image Hash:     $(cat "${OUTPUT_DIR}/image_id.txt")"
    echo "Manifest Hash:  $(cat "${OUTPUT_DIR}/manifest_hash.txt")"
    echo ""
    echo "Output files:"
    echo "  - ${OUTPUT_DIR}/pipeline_version.json"
    echo "  - ${OUTPUT_DIR}/model_manifest.json"
    echo "  - ${OUTPUT_DIR}/build.log"
    echo ""
    echo "Next steps:"
    echo "  1. Push image: docker push ${FULL_IMAGE}"
    echo "  2. Register on chain:"
    echo "     virtengine tx veid register-pipeline-version \\"
    echo "       --version ${VERSION} \\"
    echo "       --image-hash \$(cat ${OUTPUT_DIR}/image_id.txt) \\"
    echo "       --manifest-hash \$(cat ${OUTPUT_DIR}/manifest_hash.txt)"
    echo ""
    echo "=========================================="
}

# Main execution
main() {
    log_info "Building VirtEngine VEID Pipeline v${VERSION}"
    
    check_dependencies
    setup_output_dir
    compute_model_hashes
    build_image
    get_image_hash
    generate_registration_info
    verify_image
    print_summary
}

main "$@"
