#!/bin/bash
# Harmony - Container Management Script
# Supports both Docker and Podman

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Detect container runtime
detect_runtime() {
    if command -v podman &> /dev/null; then
        CONTAINER_RUNTIME="podman"
        if command -v podman-compose &> /dev/null; then
            COMPOSE_RUNTIME="podman-compose"
        else
            COMPOSE_RUNTIME="podman compose"
        fi
    elif command -v docker &> /dev/null; then
        CONTAINER_RUNTIME="docker"
        if command -v docker-compose &> /dev/null; then
            COMPOSE_RUNTIME="docker-compose"
        else
            COMPOSE_RUNTIME="docker compose"
        fi
    else
        echo -e "${RED}Error: Neither Docker nor Podman found.${NC}"
        echo "Please install Docker or Podman first."
        exit 1
    fi
}

# Show runtime info
show_info() {
    echo -e "${GREEN}Container Runtime:${NC} $CONTAINER_RUNTIME"
    echo -e "${GREEN}Compose Command:${NC} $COMPOSE_RUNTIME"
    echo ""
    $CONTAINER_RUNTIME --version
}

# Start services
start_services() {
    echo -e "${GREEN}Starting Harmony...${NC}"
    $COMPOSE_RUNTIME up -d
    echo ""
    echo -e "${GREEN}Harmony is running!${NC}"
    echo "  Frontend: http://localhost:${FRONTEND_PORT:-3000}"
    echo "  Backend:  http://localhost:${API_PORT:-8080}"
}

# Stop services
stop_services() {
    echo -e "${YELLOW}Stopping Harmony...${NC}"
    $COMPOSE_RUNTIME down
}

# Build images
build_images() {
    echo -e "${GREEN}Building images...${NC}"
    $COMPOSE_RUNTIME build "$@"
}

# Show logs
show_logs() {
    $COMPOSE_RUNTIME logs -f "$@"
}

# Show status
show_status() {
    $COMPOSE_RUNTIME ps
}

# Trigger scan
trigger_scan() {
    echo -e "${GREEN}Triggering library scan...${NC}"
    curl -X POST "http://localhost:${API_PORT:-8080}/api/v1/library/scan"
    echo ""
}

# Shell into container
shell_into() {
    local container="$1"
    if [ -z "$container" ]; then
        echo "Usage: $0 shell <backend|frontend|redis>"
        exit 1
    fi
    $CONTAINER_RUNTIME exec -it "harmony-$container" /bin/sh
}

# Production mode with Traefik
start_prod() {
    echo -e "${GREEN}Starting Harmony with Traefik (production)...${NC}"
    $COMPOSE_RUNTIME -f docker-compose.yml -f docker-compose.traefik.yml up -d
}

stop_prod() {
    $COMPOSE_RUNTIME -f docker-compose.yml -f docker-compose.traefik.yml down
}

# Show help
show_help() {
    echo "Harmony - Self-Hosted Music Streaming"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  up, start      Start all services"
    echo "  down, stop     Stop all services"
    echo "  restart        Restart all services"
    echo "  build          Build container images"
    echo "  logs [service] Show logs (optionally for specific service)"
    echo "  status, ps     Show service status"
    echo "  shell <name>   Open shell in container (backend/frontend/redis)"
    echo "  scan           Trigger library scan"
    echo "  info           Show runtime information"
    echo ""
    echo "Production:"
    echo "  up-prod        Start with Traefik reverse proxy"
    echo "  down-prod      Stop production deployment"
    echo ""
    echo "Examples:"
    echo "  $0 up                    # Start Harmony"
    echo "  $0 logs backend          # Show backend logs"
    echo "  $0 shell backend         # Shell into backend container"
    echo "  $0 scan                  # Scan music library"
}

# Main
detect_runtime

case "${1:-help}" in
    up|start)
        start_services
        ;;
    down|stop)
        stop_services
        ;;
    restart)
        stop_services
        start_services
        ;;
    build)
        shift
        build_images "$@"
        ;;
    logs)
        shift
        show_logs "$@"
        ;;
    status|ps)
        show_status
        ;;
    shell)
        shift
        shell_into "$@"
        ;;
    scan)
        trigger_scan
        ;;
    info)
        show_info
        ;;
    up-prod)
        start_prod
        ;;
    down-prod)
        stop_prod
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac
