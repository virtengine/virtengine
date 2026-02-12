#!/bin/bash
#
# VirtEngine Local Development Network Startup Script
# One-command startup for the complete local development environment
#
# Usage:
#   ./scripts/localnet.sh [command]
#
# Commands:
#   start    - Start the localnet (default)
#   stop     - Stop all services
#   restart  - Restart all services
#   update   - Smart rebuild: only rebuild changed services, preserve data
#   reset    - Stop, clean data, and restart
#   status   - Show service status
#   logs     - Tail logs from all services
#   test     - Run integration tests
#   shell    - Open shell in test-runner container
#
# Environment Variables:
#   CHAIN_ID          - Chain ID (default: virtengine-localnet-1)
#   LOG_LEVEL         - Log level (default: info)
#   DETACH            - Run in background (default: true)
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yaml"
ENV_FILE="${PROJECT_ROOT}/.env.localnet"

CHAIN_ID="${CHAIN_ID:-virtengine-localnet-1}"
LOG_LEVEL="${LOG_LEVEL:-info}"
DETACH="${DETACH:-true}"

# Helper functions
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

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running. Please start Docker."
        exit 1
    fi

    if ! docker compose version &> /dev/null 2>&1; then
        log_error "Docker Compose V2 is required. Please update Docker."
        exit 1
    fi
}

compose_cmd() {
    local env_args=""
    if [ -f "${ENV_FILE}" ]; then
        env_args="--env-file ${ENV_FILE}"
    fi
    if [ -f "${ENV_FILE}.local" ]; then
        env_args="${env_args} --env-file ${ENV_FILE}.local"
    fi
    docker compose -f "${COMPOSE_FILE}" ${env_args} "$@"
}

wait_for_chain() {
    log_info "Waiting for VirtEngine chain to be ready..."
    local max_attempts=60
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -sf http://localhost:26657/status > /dev/null 2>&1; then
            log_success "VirtEngine chain is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        echo -n "."
        sleep 2
    done

    echo ""
    log_error "Timeout waiting for chain to start"
    return 1
}

wait_for_waldur() {
    log_info "Waiting for Waldur API to be ready..."
    local max_attempts=60
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -sf -k https://localhost/health-check/ > /dev/null 2>&1; then
            log_success "Waldur API is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        echo -n "."
        sleep 5
    done

    echo ""
    log_warn "Waldur API not responding yet (may still be starting)"
    return 1
}

print_status() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "                 VirtEngine Local Development Network          "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    echo "  VirtEngine Core:"
    echo "    • Chain RPC:        http://localhost:26657"
    echo "    • Chain REST:       http://localhost:1317"
    echo "    • Chain gRPC:       localhost:9090"
    echo "    • Provider API:     https://localhost:8443"
    echo "    • Explorer:         http://localhost:8088  (blockchain explorer)"
    echo ""
    echo "  Waldur (Provider Backend):"
    echo "    • Waldur UI:        https://localhost"
    echo "    • Waldur API:       https://localhost/api/"
    echo "    • Waldur Admin:     https://localhost/admin/"
    echo "    • Keycloak Admin:   https://localhost/auth/admin"
    echo "    • Health Check:     https://localhost/health-check/"
    echo ""
    echo "  Infrastructure:"
    echo "    • API Gateway:      http://localhost:8000"
    echo "    • Local Portal UI:  http://localhost:3000"
    echo "    • Dev Portal:       http://localhost:3001/portal"
    echo "    • Prometheus:       http://localhost:9095"
    echo "    • Grafana:          http://localhost:3002  (admin/admin)"
    echo ""
    echo "  Useful Commands:"
    echo "    • Create Waldur admin:  ./scripts/localnet.sh create-admin"
    echo "    • View logs:            ./scripts/localnet.sh logs"
    echo "    • Run tests:            ./scripts/localnet.sh test"
    echo "    • Stop network:         ./scripts/localnet.sh stop"
    echo "    • Reset network:        ./scripts/localnet.sh reset"
    echo ""
    echo "  Chain Info:"
    echo "    • Chain ID:       ${CHAIN_ID}"
    echo "    • Keyring:        test"
    echo "    • Offering Sync:  enabled (Waldur ↔ Chain)"
    echo ""
    echo "  Waldur Credentials (after create-admin):"
    echo "    • Admin user:     (set via create-admin command)"
    echo "    • Keycloak admin: admin / admin"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
}

cmd_start() {
    log_info "Starting VirtEngine localnet..."

    check_docker

    # Build images that need building
    log_info "Building Docker images (virtengine-node, provider-daemon, portal)..."
    compose_cmd build virtengine-node provider-daemon portal

    # Start services
    if [ "${DETACH}" = "true" ]; then
        log_info "Starting services in background..."
        compose_cmd up -d
    else
        log_info "Starting services in foreground..."
        compose_cmd up
        return
    fi

    # Wait for chain
    wait_for_chain

    # Wait for Waldur (non-fatal if slow)
    if wait_for_waldur; then
        # Initialize marketplace categories after Waldur is ready
        log_info "Initializing Waldur marketplace categories..."
        cmd_init_categories_internal || log_warn "Category initialization failed (can retry with 'init-categories')"
    fi

    # Print status
    print_status

    log_success "VirtEngine localnet is running!"
}

cmd_stop() {
    log_info "Stopping VirtEngine localnet..."
    check_docker
    compose_cmd down
    log_success "VirtEngine localnet stopped."
}

cmd_restart() {
    log_info "Restarting VirtEngine localnet..."
    cmd_stop
    cmd_start
}

# Track last update timestamp for smart rebuilds
LAST_UPDATE_FILE="${PROJECT_ROOT}/.localnet-last-update"

get_service_source_hash() {
    local service="$1"
    case "$service" in
        virtengine-node)
            # Hash key source files for the chain node
            find "${PROJECT_ROOT}/cmd/virtengine" "${PROJECT_ROOT}/x" "${PROJECT_ROOT}/app" \
                "${PROJECT_ROOT}/_build/Dockerfile.virtengine" "${PROJECT_ROOT}/scripts/init-chain.sh" \
                -type f \( -name "*.go" -o -name "Dockerfile*" -o -name "*.sh" \) 2>/dev/null | \
                xargs cat 2>/dev/null | sha256sum | cut -d' ' -f1
            ;;
        provider-daemon)
            # Hash key source files for provider daemon
            find "${PROJECT_ROOT}/cmd/provider-daemon" "${PROJECT_ROOT}/pkg/provider_daemon" \
                "${PROJECT_ROOT}/_build/Dockerfile.provider-daemon" \
                -type f \( -name "*.go" -o -name "Dockerfile*" \) 2>/dev/null | \
                xargs cat 2>/dev/null | sha256sum | cut -d' ' -f1
            ;;
        portal)
            # Hash key source files for the local portal UI
            find "${PROJECT_ROOT}/portal" "${PROJECT_ROOT}/lib/portal" "${PROJECT_ROOT}/lib/capture" \
                "${PROJECT_ROOT}/lib/admin" "${PROJECT_ROOT}/_build/Dockerfile.portal" \
                "${PROJECT_ROOT}/pnpm-workspace.yaml" "${PROJECT_ROOT}/pnpm-lock.yaml" \
                -type f \( -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" -o -name "*.json" -o -name "*.css" -o -name "*.md" -o -name "Dockerfile*" \) 2>/dev/null | \
                xargs cat 2>/dev/null | sha256sum | cut -d' ' -f1
            ;;
        *)
            echo "upstream"  # Upstream images don't need rebuilding
            ;;
    esac
}

cmd_update() {
    log_info "Smart update: checking for changed services..."
    check_docker

    local services_to_rebuild=()
    local services_to_restart=()

    # Check if localnet is running
    if ! docker ps --format '{{.Names}}' | grep -q virtengine-node; then
        log_warn "Localnet is not running. Starting fresh..."
        cmd_start
        return
    fi

    # Services we build from source
    local build_services=("virtengine-node" "provider-daemon" "portal")

    # Check each build service for changes
    for service in "${build_services[@]}"; do
        local current_hash
        current_hash=$(get_service_source_hash "$service")
        local stored_hash=""
        
        if [ -f "${LAST_UPDATE_FILE}.${service}" ]; then
            stored_hash=$(cat "${LAST_UPDATE_FILE}.${service}")
        fi

        if [ "$current_hash" != "$stored_hash" ]; then
            log_info "Changes detected in ${service}"
            services_to_rebuild+=("$service")
        else
            log_info "No changes in ${service}"
        fi
    done

    if [ ${#services_to_rebuild[@]} -eq 0 ]; then
        log_success "No changes detected. Environment is up to date."
        return
    fi

    # Rebuild changed services
    log_info "Rebuilding ${#services_to_rebuild[@]} service(s): ${services_to_rebuild[*]}"
    
    for service in "${services_to_rebuild[@]}"; do
        log_info "Building ${service}..."
        if ! compose_cmd build --no-cache "$service"; then
            log_error "Failed to build ${service}"
            exit 1
        fi
        
        # Store the new hash
        get_service_source_hash "$service" > "${LAST_UPDATE_FILE}.${service}"
    done

    # Restart only the rebuilt services (preserves data volumes)
    log_info "Restarting rebuilt services..."
    for service in "${services_to_rebuild[@]}"; do
        log_info "Restarting ${service}..."
        compose_cmd up -d --no-deps "$service"
    done

    # Wait for chain if it was rebuilt
    if [[ " ${services_to_rebuild[*]} " =~ " virtengine-node " ]]; then
        wait_for_chain
    fi

    # Wait for Waldur if provider-daemon was rebuilt (it depends on Waldur)
    if [[ " ${services_to_rebuild[*]} " =~ " provider-daemon " ]]; then
        sleep 5  # Give provider-daemon time to connect
    fi

    log_success "Update complete! Rebuilt: ${services_to_rebuild[*]}"
    echo ""
    log_info "Data volumes preserved. Use 'reset' to start fresh."
}

cmd_reset() {
    log_warn "This will delete all localnet data. Continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Aborted."
        exit 0
    fi

    log_info "Resetting VirtEngine localnet..."
    check_docker

    # Stop and remove containers, volumes
    compose_cmd down -v --remove-orphans

    # Clean up any remaining data
    rm -rf "${PROJECT_ROOT}/.localnet" 2>/dev/null || true

    log_success "Localnet data cleaned."

    # Restart
    cmd_start
}

cmd_status() {
    check_docker
    echo ""
    log_info "Service Status:"
    echo ""
    compose_cmd ps
    echo ""

    # Check if chain is responding
    if curl -sf http://localhost:26657/status > /dev/null 2>&1; then
        log_success "Chain RPC: Healthy"
        
        # Get chain info
        local chain_info
        chain_info=$(curl -sf http://localhost:26657/status 2>/dev/null)
        if [ -n "$chain_info" ]; then
            local latest_height
            latest_height=$(echo "$chain_info" | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' | head -1)
            if [ -n "$latest_height" ]; then
                echo "    Latest Block Height: ${latest_height}"
            fi
        fi
    else
        log_warn "Chain RPC: Not responding"
    fi

    # Check Waldur
    if curl -sf -k https://localhost/health-check/ > /dev/null 2>&1; then
        log_success "Waldur API: Healthy"
    else
        log_warn "Waldur API: Not responding"
    fi
}

cmd_create_admin() {
    local username=""
    local password=""
    local email=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -u|--username)
                username="$2"
                shift 2
                ;;
            -p|--password)
                password="$2"
                shift 2
                ;;
            -e|--email)
                email="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    # Interactive prompts if not provided
    if [ -z "$username" ]; then
        read -rp "Enter admin username [admin]: " username
        username="${username:-admin}"
    fi

    if [ -z "$email" ]; then
        read -rp "Enter admin email [${username}@localhost]: " email
        email="${email:-${username}@localhost}"
    fi

    if [ -z "$password" ]; then
        while true; do
            read -rsp "Enter admin password: " password
            echo ""
            if [ -z "$password" ]; then
                log_warn "Password cannot be empty. Please try again."
                continue
            fi
            read -rsp "Confirm password: " password_confirm
            echo ""
            if [ "$password" != "$password_confirm" ]; then
                log_warn "Passwords do not match. Please try again."
                continue
            fi
            break
        done
    fi

    log_info "Creating Waldur admin user '${username}'..."
    check_docker

    if ! docker ps --format '{{.Names}}' | grep -q waldur-mastermind-api; then
        log_error "Waldur API is not running. Start the localnet first."
        exit 1
    fi

    # Create superuser (--noinput skips interactive prompts)
    if ! docker exec waldur-mastermind-api waldur createsuperuser \
        --username "${username}" \
        --email "${email}" \
        --noinput 2>/dev/null; then
        log_warn "User '${username}' may already exist, attempting password update..."
    fi

    # Set password using Django shell (more reliable than changepassword)
    local python_cmd="from django.contrib.auth import get_user_model; User = get_user_model(); u = User.objects.get(username='${username}'); u.set_password('${password}'); u.save(); print('Password set successfully')"
    
    if docker exec waldur-mastermind-api waldur shell -c "${python_cmd}" 2>/dev/null; then
        log_success "Waldur admin user '${username}' created/updated successfully!"
    else
        log_error "Failed to set password for '${username}'"
        exit 1
    fi

    log_info "Access Waldur UI at: https://localhost"
    log_info "Access Waldur API at: https://localhost/api/"
    echo ""
    log_info "To get an API token, run:"
    echo "  curl -k -X POST https://localhost/api-auth/password/ -d 'username=${username}&password=<your-password>'"
}

# VE-25A: Internal function to initialize categories (called from cmd_start)
cmd_init_categories_internal() {
    local max_attempts=3
    local attempt=0

    # Get admin token for API calls
    local token=""
    
    # Try to get token from existing admin user
    # First check if we can authenticate
    token=$(curl -sf -k -X POST https://localhost/api-auth/password/ \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=admin&password=admin" 2>/dev/null | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -z "$token" ]; then
        log_warn "Could not get admin token. Create admin user first with 'create-admin' command."
        return 1
    fi

    log_info "Creating VirtEngine marketplace categories..."

    # Define categories
    local categories=(
        "Compute|Virtual machines, containers, and general-purpose computing resources."
        "HPC|High-performance computing resources including MPI clusters and batch processing."
        "GPU|GPU-accelerated computing instances for machine learning and deep learning."
        "Storage|Object storage, block storage, and file storage solutions."
        "Network|Networking services including VPNs, load balancers, and firewalls."
        "TEE|Trusted Execution Environment resources for confidential computing."
        "AI/ML|Machine learning platforms, model training, and inference services."
    )

    local created=0
    local existing=0
    local failed=0

    for cat_def in "${categories[@]}"; do
        local title="${cat_def%%|*}"
        local description="${cat_def#*|}"

        # Check if category already exists
        local exists
        exists=$(curl -sf -k "https://localhost/api/marketplace-categories/?title=${title}" \
            -H "Authorization: Token ${token}" 2>/dev/null | grep -c "\"title\":\"${title}\"" || echo "0")

        if [ "$exists" -gt 0 ]; then
            log_info "  Category '${title}' already exists"
            existing=$((existing + 1))
            continue
        fi

        # Create category
        local result
        result=$(curl -sf -k -X POST "https://localhost/api/marketplace-categories/" \
            -H "Authorization: Token ${token}" \
            -H "Content-Type: application/json" \
            -d "{\"title\":\"${title}\",\"description\":\"${description}\"}" 2>/dev/null)

        if echo "$result" | grep -q "\"uuid\""; then
            log_success "  Created category: ${title}"
            created=$((created + 1))
        else
            log_warn "  Failed to create category: ${title}"
            failed=$((failed + 1))
        fi
    done

    echo ""
    log_info "Category initialization complete:"
    log_info "  Created: ${created}"
    log_info "  Existing: ${existing}"
    [ $failed -gt 0 ] && log_warn "  Failed: ${failed}"

    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# VE-25A: Public command to initialize categories
cmd_init_categories() {
    log_info "Initializing Waldur marketplace categories..."
    check_docker

    # Check if Waldur is running
    if ! curl -sf -k https://localhost/health-check/ > /dev/null 2>&1; then
        log_error "Waldur API is not responding. Start localnet first."
        exit 1
    fi

    cmd_init_categories_internal
    local result=$?

    if [ $result -eq 0 ]; then
        log_success "All categories initialized successfully!"
    else
        log_error "Some categories failed to initialize. Check logs above."
        exit 1
    fi
}

cmd_logs() {
    check_docker
    local service="${1:-}"

    if [ -n "$service" ]; then
        compose_cmd logs -f "$service"
    else
        compose_cmd logs -f
    fi
}

cmd_test() {
    log_info "Running integration tests..."
    check_docker

    # Ensure services are running
    if ! curl -sf http://localhost:26657/status > /dev/null 2>&1; then
        log_warn "Chain is not running. Starting localnet first..."
        cmd_start
    fi

    # Run tests
    compose_cmd run --rm test-runner go test -v ./tests/integration/...
}

cmd_shell() {
    log_info "Opening shell in test-runner container..."
    check_docker
    compose_cmd run --rm test-runner /bin/sh
}

cmd_help() {
    echo "VirtEngine Local Development Network"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  start            Start the localnet (default)"
    echo "  stop             Stop all services"
    echo "  restart          Restart all services (full restart)"
    echo "  update           Smart rebuild: only rebuild changed services, preserve data"
    echo "  reset            Stop, clean data, and restart (destructive)"
    echo "  status           Show service status"
    echo "  logs             Tail logs from all services"
    echo "  test             Run integration tests"
    echo "  shell            Open shell in test-runner container"
    echo "  create-admin     Create Waldur admin user (interactive or with flags)"
    echo "  init-categories  Initialize Waldur marketplace categories"
    echo "  help             Show this help message"
    echo ""
    echo "Workflow:"
    echo "  First time:   $0 start                # Build and start everything"
    echo "  After edits:  $0 update               # Smart rebuild changed services only"
    echo "  Full reset:   $0 reset                # Wipe data and start fresh"
    echo ""
    echo "Waldur Setup:"
    echo "  1. $0 start                           # Start localnet"
    echo "  2. $0 create-admin -u admin -p admin  # Create admin user"
    echo "  3. $0 init-categories                 # Create marketplace categories"
    echo ""
    echo "create-admin Options:"
    echo "  -u, --username  Admin username (default: admin, or prompted)"
    echo "  -p, --password  Admin password (prompted if not provided)"
    echo "  -e, --email     Admin email (default: <username>@localhost)"
    echo ""
    echo "Examples:"
    echo "  $0 update                              # Rebuild only changed services"
    echo "  $0 create-admin                        # Interactive admin creation"
    echo "  $0 create-admin -u myuser -p mypassword # Non-interactive"
    echo "  $0 init-categories                     # Create VirtEngine categories"
    echo ""
    echo "Environment Variables:"
    echo "  CHAIN_ID      Chain ID (default: virtengine-localnet-1)"
    echo "  LOG_LEVEL     Log level (default: info)"
    echo "  DETACH        Run in background (default: true)"
}

# Main
main() {
    local command="${1:-start}"

    case "$command" in
        start)
            cmd_start
            ;;
        stop)
            cmd_stop
            ;;
        restart)
            cmd_restart
            ;;
        update)
            cmd_update
            ;;
        reset)
            cmd_reset
            ;;
        status)
            cmd_status
            ;;
        logs)
            shift
            cmd_logs "$@"
            ;;
        test)
            cmd_test
            ;;
        shell)
            cmd_shell
            ;;
        create-admin)
            shift
            cmd_create_admin "$@"
            ;;
        init-categories)
            cmd_init_categories
            ;;
        help|--help|-h)
            cmd_help
            ;;
        *)
            log_error "Unknown command: $command"
            cmd_help
            exit 1
            ;;
    esac
}

main "$@"
