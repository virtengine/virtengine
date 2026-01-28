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

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose."
        exit 1
    fi
}

compose_cmd() {
    if docker compose version &> /dev/null 2>&1; then
        docker compose -f "${COMPOSE_FILE}" "$@"
    else
        docker-compose -f "${COMPOSE_FILE}" "$@"
    fi
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

print_status() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "                 VirtEngine Local Development Network          "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    echo "  Services:"
    echo "    • Chain RPC:      http://localhost:26657"
    echo "    • Chain REST:     http://localhost:1317"
    echo "    • Chain gRPC:     localhost:9090"
    echo "    • Waldur API:     http://localhost:8080"
    echo "    • Portal:         http://localhost:3000"
    echo "    • Provider API:   https://localhost:8443"
    echo ""
    echo "  Useful Commands:"
    echo "    • View logs:      ./scripts/localnet.sh logs"
    echo "    • Run tests:      ./scripts/localnet.sh test"
    echo "    • Stop network:   ./scripts/localnet.sh stop"
    echo "    • Reset network:  ./scripts/localnet.sh reset"
    echo ""
    echo "  Chain Info:"
    echo "    • Chain ID:       ${CHAIN_ID}"
    echo "    • Keyring:        test"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
}

cmd_start() {
    log_info "Starting VirtEngine localnet..."

    check_docker

    # Build images if needed
    log_info "Building Docker images..."
    compose_cmd build

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
    echo "  start     Start the localnet (default)"
    echo "  stop      Stop all services"
    echo "  restart   Restart all services"
    echo "  reset     Stop, clean data, and restart"
    echo "  status    Show service status"
    echo "  logs      Tail logs from all services"
    echo "  test      Run integration tests"
    echo "  shell     Open shell in test-runner container"
    echo "  help      Show this help message"
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
