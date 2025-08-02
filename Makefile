# Copa Litoral Backend - Makefile
# Comandos para desarrollo, testing y deployment

.PHONY: help build run test test-unit test-integration test-api test-coverage clean lint format deps

# Variables
BINARY_NAME=copa-litoral-backend
COVERAGE_DIR=coverage
TEST_DB_NAME=copa_litoral_test

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Mostrar ayuda
	@echo "Copa Litoral Backend - Comandos disponibles:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Comandos de desarrollo
build: ## Compilar la aplicación
	@echo "$(GREEN)Compilando aplicación...$(NC)"
	go build -o bin/$(BINARY_NAME) .

run: ## Ejecutar la aplicación
	@echo "$(GREEN)Ejecutando aplicación...$(NC)"
	go run .

deps: ## Instalar dependencias
	@echo "$(GREEN)Instalando dependencias...$(NC)"
	go mod download
	go mod tidy

clean: ## Limpiar archivos generados
	@echo "$(YELLOW)Limpiando archivos...$(NC)"
	rm -rf bin/
	rm -rf $(COVERAGE_DIR)/
	go clean

# Comandos de testing
test: ## Ejecutar todas las pruebas con coverage
	@echo "$(GREEN)Ejecutando suite completa de pruebas...$(NC)"
	./scripts/test.sh

test-unit: ## Ejecutar solo pruebas unitarias
	@echo "$(GREEN)Ejecutando pruebas unitarias...$(NC)"
	./scripts/test.sh --unit-only

test-integration: ## Ejecutar solo pruebas de integración
	@echo "$(GREEN)Ejecutando pruebas de integración...$(NC)"
	./scripts/test.sh --integration-only

test-api: ## Ejecutar solo pruebas de API
	@echo "$(GREEN)Ejecutando pruebas de API...$(NC)"
	./scripts/test.sh --api-only

test-bench: ## Ejecutar pruebas con benchmarks
	@echo "$(GREEN)Ejecutando pruebas con benchmarks...$(NC)"
	./scripts/test.sh --with-benchmarks

test-coverage: ## Generar reporte de coverage y abrirlo
	@echo "$(GREEN)Generando reporte de coverage...$(NC)"
	./scripts/test.sh
	@if [ -f "$(COVERAGE_DIR)/coverage.html" ]; then \
		echo "$(GREEN)Abriendo reporte de coverage...$(NC)"; \
		xdg-open $(COVERAGE_DIR)/coverage.html 2>/dev/null || open $(COVERAGE_DIR)/coverage.html 2>/dev/null || echo "$(YELLOW)Reporte generado en: $(COVERAGE_DIR)/coverage.html$(NC)"; \
	fi

test-quick: ## Ejecutar pruebas rápidas (sin integración)
	@echo "$(GREEN)Ejecutando pruebas rápidas...$(NC)"
	go test -short ./tests/unit/...

# Comandos de calidad de código
lint: ## Ejecutar análisis de código
	@echo "$(GREEN)Ejecutando análisis de código...$(NC)"
	go vet ./...
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "$(YELLOW)golint no está instalado$(NC)"; \
	fi

format: ## Formatear código
	@echo "$(GREEN)Formateando código...$(NC)"
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)goimports no está instalado$(NC)"; \
	fi

# Comandos de base de datos
db-setup: ## Configurar base de datos de pruebas
	@echo "$(GREEN)Configurando base de datos de pruebas...$(NC)"
	@if command -v psql >/dev/null 2>&1; then \
		createdb $(TEST_DB_NAME) 2>/dev/null || echo "$(YELLOW)Base de datos $(TEST_DB_NAME) ya existe$(NC)"; \
		echo "$(GREEN)Base de datos de pruebas lista$(NC)"; \
	else \
		echo "$(RED)PostgreSQL no está instalado$(NC)"; \
	fi

db-reset: ## Resetear base de datos de pruebas
	@echo "$(YELLOW)Reseteando base de datos de pruebas...$(NC)"
	@if command -v psql >/dev/null 2>&1; then \
		dropdb $(TEST_DB_NAME) 2>/dev/null || true; \
		createdb $(TEST_DB_NAME); \
		echo "$(GREEN)Base de datos de pruebas reseteada$(NC)"; \
	else \
		echo "$(RED)PostgreSQL no está instalado$(NC)"; \
	fi

# Comandos de desarrollo
dev: ## Ejecutar en modo desarrollo con recarga automática
	@echo "$(GREEN)Iniciando modo desarrollo...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)air no está instalado, ejecutando normalmente...$(NC)"; \
		go run .; \
	fi

install-tools: ## Instalar herramientas de desarrollo
	@echo "$(GREEN)Instalando herramientas de desarrollo...$(NC)"
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Herramientas instaladas$(NC)"

# Comandos de documentación
docs: ## Generar documentación
	@echo "$(GREEN)Generando documentación...$(NC)"
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Documentación disponible en: http://localhost:6060/pkg/copa-litoral-backend/"; \
		godoc -http=:6060; \
	else \
		echo "$(YELLOW)godoc no está instalado$(NC)"; \
		echo "Instalar con: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Comandos de deployment
docker-build: ## Construir imagen Docker
	@echo "$(GREEN)Construyendo imagen Docker...$(NC)"
	docker build -t copa-litoral-backend .

docker-run: ## Ejecutar contenedor Docker
	@echo "$(GREEN)Ejecutando contenedor Docker...$(NC)"
	docker run -p 8080:8080 copa-litoral-backend

# Comandos de git
git-hooks: ## Configurar git hooks
	@echo "$(GREEN)Configurando git hooks...$(NC)"
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake test-quick' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)Git hooks configurados$(NC)"

# Comandos de CI/CD
ci: ## Ejecutar pipeline de CI (como en GitHub Actions)
	@echo "$(GREEN)Ejecutando pipeline de CI...$(NC)"
	make deps
	make lint
	make format
	make test
	make build

# Información del sistema
info: ## Mostrar información del sistema
	@echo "$(GREEN)Información del sistema:$(NC)"
	@echo "Go version: $$(go version)"
	@echo "OS: $$(uname -s)"
	@echo "Architecture: $$(uname -m)"
	@echo "Working directory: $$(pwd)"
	@echo "Git branch: $$(git branch --show-current 2>/dev/null || echo 'N/A')"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"

# Comando por defecto
.DEFAULT_GOAL := help
