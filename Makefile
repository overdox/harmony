# Harmony - Container Management
# Supports both Docker and Podman

# Detect container runtime (prefer podman if available)
CONTAINER_RUNTIME := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
COMPOSE_RUNTIME := $(shell command -v podman-compose 2>/dev/null || command -v docker-compose 2>/dev/null || echo "$(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null) compose")

# Check if using podman
IS_PODMAN := $(shell echo $(CONTAINER_RUNTIME) | grep -q podman && echo true || echo false)

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

.PHONY: help info up down build start stop restart logs shell-backend shell-frontend clean prune scan

# Default target
help:
	@echo "Harmony - Self-Hosted Music Streaming"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Container Management:"
	@echo "  up          Start all services (detached)"
	@echo "  down        Stop and remove all services"
	@echo "  build       Build all container images"
	@echo "  start       Start existing containers"
	@echo "  stop        Stop running containers"
	@echo "  restart     Restart all services"
	@echo "  logs        Follow container logs"
	@echo ""
	@echo "Development:"
	@echo "  shell-backend   Open shell in backend container"
	@echo "  shell-frontend  Open shell in frontend container"
	@echo "  scan            Trigger library scan"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean       Remove containers and images"
	@echo "  prune       Remove unused containers, images, and volumes"
	@echo ""
	@echo "Information:"
	@echo "  info        Show container runtime information"
	@echo ""

# Show runtime info
info:
	@echo "$(GREEN)Container Runtime:$(NC) $(CONTAINER_RUNTIME)"
	@echo "$(GREEN)Compose Command:$(NC) $(COMPOSE_RUNTIME)"
	@echo "$(GREEN)Using Podman:$(NC) $(IS_PODMAN)"
	@echo ""
	@$(CONTAINER_RUNTIME) --version
	@echo ""
	@if [ "$(IS_PODMAN)" = "true" ]; then \
		podman-compose --version 2>/dev/null || echo "podman-compose not installed (using podman compose)"; \
	else \
		docker compose version 2>/dev/null || docker-compose --version; \
	fi

# Start all services
up:
	@echo "$(GREEN)Starting Harmony...$(NC)"
	@$(COMPOSE_RUNTIME) up -d
	@echo ""
	@echo "$(GREEN)Harmony is running!$(NC)"
	@echo "  Frontend: http://localhost:$${FRONTEND_PORT:-3000}"
	@echo "  Backend:  http://localhost:$${API_PORT:-8080}"

# Stop and remove services
down:
	@echo "$(YELLOW)Stopping Harmony...$(NC)"
	@$(COMPOSE_RUNTIME) down

# Build images
build:
	@echo "$(GREEN)Building images...$(NC)"
	@$(COMPOSE_RUNTIME) build

# Build with no cache
build-fresh:
	@echo "$(GREEN)Building images (no cache)...$(NC)"
	@$(COMPOSE_RUNTIME) build --no-cache

# Start existing containers
start:
	@$(COMPOSE_RUNTIME) start

# Stop containers
stop:
	@$(COMPOSE_RUNTIME) stop

# Restart services
restart:
	@echo "$(YELLOW)Restarting Harmony...$(NC)"
	@$(COMPOSE_RUNTIME) restart

# Follow logs
logs:
	@$(COMPOSE_RUNTIME) logs -f

# Backend logs only
logs-backend:
	@$(COMPOSE_RUNTIME) logs -f backend

# Frontend logs only
logs-frontend:
	@$(COMPOSE_RUNTIME) logs -f frontend

# Shell into backend
shell-backend:
	@$(CONTAINER_RUNTIME) exec -it harmony-backend /bin/sh

# Shell into frontend
shell-frontend:
	@$(CONTAINER_RUNTIME) exec -it harmony-frontend /bin/sh

# Trigger library scan
scan:
	@echo "$(GREEN)Triggering library scan...$(NC)"
	@curl -X POST http://localhost:$${API_PORT:-8080}/api/v1/library/scan
	@echo ""

# Show service status
status:
	@$(COMPOSE_RUNTIME) ps

# Clean up containers and images
clean:
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@$(COMPOSE_RUNTIME) down --rmi local

# Prune unused resources
prune:
	@echo "$(YELLOW)Pruning unused resources...$(NC)"
	@$(CONTAINER_RUNTIME) system prune -f

# Production deployment with Traefik
up-prod:
	@echo "$(GREEN)Starting Harmony with Traefik...$(NC)"
	@$(COMPOSE_RUNTIME) -f docker-compose.yml -f docker-compose.traefik.yml up -d

down-prod:
	@$(COMPOSE_RUNTIME) -f docker-compose.yml -f docker-compose.traefik.yml down

# Development mode (with override)
up-dev:
	@echo "$(GREEN)Starting Harmony in development mode...$(NC)"
	@$(COMPOSE_RUNTIME) up -d
