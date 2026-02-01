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
    echo ""
    echo "  Waldur:"
    echo "    • Waldur UI:        https://localhost"
    echo "    • Waldur API:       https://localhost/api/"
    echo "    • Waldur Admin:     https://localhost/admin/"
    echo "    • Keycloak Admin:   https://localhost/auth/admin"
    echo "    • Health Check:     https://localhost/health-check/"
    echo ""
    echo "  Infrastructure:"
    echo "    • API Gateway:      http://localhost:8000"
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
    echo ""
    echo "  Waldur Credentials (after create-admin):"
    echo "    • Admin user:     admin"
    echo "    • Admin password: admin"
    echo "    • Keycloak admin: admin / admin"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
}

cmd_start() {
    log_info "Starting VirtEngine localnet..."

    check_docker

    # Build images that need building
    log_info "Building Docker images (virtengine-node, provider-daemon)..."
    compose_cmd build virtengine-node provider-daemon

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
    wait_for_waldur || true

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
    echo "  start         Start the localnet (default)"
    echo "  stop          Stop all services"
    echo "  restart       Restart all services"
    echo "  reset         Stop, clean data, and restart"
    echo "  status        Show service status"
    echo "  logs          Tail logs from all services"
    echo "  test          Run integration tests"
    echo "  shell         Open shell in test-runner container"
    echo "  create-admin  Create Waldur admin user (interactive or with flags)"
    echo "  help          Show this help message"
    echo ""
    echo "create-admin Options:"
    echo "  -u, --username  Admin username (default: admin, or prompted)"
    echo "  -p, --password  Admin password (prompted if not provided)"
    echo "  -e, --email     Admin email (default: <username>@localhost)"
    echo ""
    echo "Examples:"
    echo "  $0 create-admin                          # Interactive"
    echo "  $0 create-admin -u myuser -p mypassword  # Non-interactive"
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
