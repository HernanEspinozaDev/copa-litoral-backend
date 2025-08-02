#!/bin/bash

# Script para ejecutar todas las pruebas con coverage
# Copa Litoral Backend - Test Runner

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para imprimir con colores
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuración
PROJECT_ROOT=$(pwd)
COVERAGE_DIR="coverage"
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

# Crear directorio de coverage
mkdir -p $COVERAGE_DIR

print_status "Iniciando suite de pruebas para Copa Litoral Backend..."

# Verificar que Go esté instalado
if ! command -v go &> /dev/null; then
    print_error "Go no está instalado o no está en el PATH"
    exit 1
fi

# Verificar variables de entorno para testing
if [ -z "$TEST_DB_HOST" ]; then
    print_warning "TEST_DB_HOST no está configurado, usando 'localhost'"
    export TEST_DB_HOST="localhost"
fi

if [ -z "$TEST_DB_NAME" ]; then
    print_warning "TEST_DB_NAME no está configurado, usando 'copa_litoral_test'"
    export TEST_DB_NAME="copa_litoral_test"
fi

# Función para ejecutar pruebas unitarias
run_unit_tests() {
    print_status "Ejecutando pruebas unitarias..."
    
    go test -v -race -coverprofile=$COVERAGE_DIR/unit.out ./tests/unit/... || {
        print_error "Las pruebas unitarias fallaron"
        return 1
    }
    
    print_success "Pruebas unitarias completadas"
}

# Función para ejecutar pruebas de integración
run_integration_tests() {
    print_status "Ejecutando pruebas de integración..."
    
    # Verificar conexión a la base de datos de pruebas
    if ! pg_isready -h ${TEST_DB_HOST:-localhost} -p ${TEST_DB_PORT:-5432} -U ${TEST_DB_USER:-postgres} &> /dev/null; then
        print_warning "No se puede conectar a la base de datos de pruebas. Saltando pruebas de integración."
        return 0
    fi
    
    go test -v -race -coverprofile=$COVERAGE_DIR/integration.out ./tests/integration/... || {
        print_error "Las pruebas de integración fallaron"
        return 1
    }
    
    print_success "Pruebas de integración completadas"
}

# Función para ejecutar pruebas de API
run_api_tests() {
    print_status "Ejecutando pruebas de API (end-to-end)..."
    
    go test -v -race -coverprofile=$COVERAGE_DIR/api.out ./tests/api/... || {
        print_error "Las pruebas de API fallaron"
        return 1
    }
    
    print_success "Pruebas de API completadas"
}

# Función para ejecutar benchmarks
run_benchmarks() {
    print_status "Ejecutando benchmarks..."
    
    go test -bench=. -benchmem ./tests/unit/... > $COVERAGE_DIR/benchmarks_unit.txt
    go test -bench=. -benchmem ./tests/integration/... > $COVERAGE_DIR/benchmarks_integration.txt
    go test -bench=. -benchmem ./tests/api/... > $COVERAGE_DIR/benchmarks_api.txt
    
    print_success "Benchmarks completados"
}

# Función para combinar coverage
combine_coverage() {
    print_status "Combinando reportes de coverage..."
    
    # Combinar todos los archivos de coverage
    echo "mode: atomic" > $COVERAGE_DIR/$COVERAGE_FILE
    
    for file in $COVERAGE_DIR/*.out; do
        if [ -f "$file" ] && [ "$file" != "$COVERAGE_DIR/$COVERAGE_FILE" ]; then
            tail -n +2 "$file" >> $COVERAGE_DIR/$COVERAGE_FILE
        fi
    done
    
    # Generar reporte HTML
    go tool cover -html=$COVERAGE_DIR/$COVERAGE_FILE -o $COVERAGE_DIR/$COVERAGE_HTML
    
    # Mostrar estadísticas de coverage
    COVERAGE_PERCENT=$(go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep total | awk '{print $3}')
    print_success "Coverage total: $COVERAGE_PERCENT"
    
    # Generar reporte detallado
    go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE > $COVERAGE_DIR/coverage_report.txt
}

# Función para ejecutar linting
run_linting() {
    print_status "Ejecutando análisis de código..."
    
    # go vet
    if go vet ./...; then
        print_success "go vet: sin problemas"
    else
        print_warning "go vet: encontró problemas"
    fi
    
    # golint si está disponible
    if command -v golint &> /dev/null; then
        golint ./... > $COVERAGE_DIR/lint_report.txt
        if [ -s $COVERAGE_DIR/lint_report.txt ]; then
            print_warning "golint: encontró sugerencias (ver coverage/lint_report.txt)"
        else
            print_success "golint: sin sugerencias"
        fi
    else
        print_warning "golint no está instalado"
    fi
    
    # gofmt
    UNFORMATTED=$(gofmt -l .)
    if [ -z "$UNFORMATTED" ]; then
        print_success "gofmt: código correctamente formateado"
    else
        print_warning "gofmt: archivos sin formatear:"
        echo "$UNFORMATTED"
    fi
}

# Función para generar reporte final
generate_report() {
    print_status "Generando reporte final..."
    
    REPORT_FILE="$COVERAGE_DIR/test_report.md"
    
    cat > $REPORT_FILE << EOF
# Reporte de Pruebas - Copa Litoral Backend

**Fecha:** $(date)
**Commit:** $(git rev-parse --short HEAD 2>/dev/null || echo "N/A")

## Resumen de Coverage

**Coverage Total:** $COVERAGE_PERCENT

## Archivos Generados

- \`coverage.html\` - Reporte visual de coverage
- \`coverage.out\` - Datos de coverage combinados
- \`coverage_report.txt\` - Reporte detallado por función
- \`benchmarks_*.txt\` - Resultados de benchmarks
- \`lint_report.txt\` - Reporte de linting

## Comandos para Ver Reportes

\`\`\`bash
# Ver coverage en el navegador
open coverage/coverage.html

# Ver reporte de coverage en terminal
go tool cover -func=coverage/coverage.out

# Ver benchmarks
cat coverage/benchmarks_*.txt
\`\`\`

## Estadísticas de Pruebas

EOF

    # Agregar estadísticas si están disponibles
    if [ -f "$COVERAGE_DIR/coverage_report.txt" ]; then
        echo "### Coverage por Archivo" >> $REPORT_FILE
        echo '```' >> $REPORT_FILE
        head -20 $COVERAGE_DIR/coverage_report.txt >> $REPORT_FILE
        echo '```' >> $REPORT_FILE
    fi
    
    print_success "Reporte generado: $REPORT_FILE"
}

# Función principal
main() {
    local run_unit=true
    local run_integration=true
    local run_api=true
    local run_bench=false
    local run_lint=true
    
    # Parsear argumentos
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit-only)
                run_integration=false
                run_api=false
                shift
                ;;
            --integration-only)
                run_unit=false
                run_api=false
                shift
                ;;
            --api-only)
                run_unit=false
                run_integration=false
                shift
                ;;
            --with-benchmarks)
                run_bench=true
                shift
                ;;
            --no-lint)
                run_lint=false
                shift
                ;;
            --help)
                echo "Uso: $0 [opciones]"
                echo "Opciones:"
                echo "  --unit-only        Solo ejecutar pruebas unitarias"
                echo "  --integration-only Solo ejecutar pruebas de integración"
                echo "  --api-only         Solo ejecutar pruebas de API"
                echo "  --with-benchmarks  Incluir benchmarks"
                echo "  --no-lint          Saltar análisis de código"
                echo "  --help             Mostrar esta ayuda"
                exit 0
                ;;
            *)
                print_error "Opción desconocida: $1"
                exit 1
                ;;
        esac
    done
    
    # Limpiar coverage anterior
    rm -f $COVERAGE_DIR/*.out $COVERAGE_DIR/*.txt $COVERAGE_DIR/*.html
    
    # Ejecutar pruebas según configuración
    if [ "$run_unit" = true ]; then
        run_unit_tests || exit 1
    fi
    
    if [ "$run_integration" = true ]; then
        run_integration_tests || exit 1
    fi
    
    if [ "$run_api" = true ]; then
        run_api_tests || exit 1
    fi
    
    if [ "$run_bench" = true ]; then
        run_benchmarks
    fi
    
    if [ "$run_lint" = true ]; then
        run_linting
    fi
    
    # Combinar coverage y generar reportes
    combine_coverage
    generate_report
    
    print_success "¡Suite de pruebas completada exitosamente!"
    print_status "Ver reporte completo en: $COVERAGE_DIR/coverage.html"
    
    # Mostrar resumen final
    echo ""
    echo "=== RESUMEN FINAL ==="
    echo "Coverage: $COVERAGE_PERCENT"
    echo "Reportes: $COVERAGE_DIR/"
    echo "====================="
}

# Ejecutar función principal con todos los argumentos
main "$@"
