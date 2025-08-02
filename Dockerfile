FROM golang:1.24-alpine AS builder

# Instalar dependencias del sistema
RUN apk add --no-cache git ca-certificates tzdata

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación con optimizaciones
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o copa-litoral-backend .

# Etapa final - imagen mínima
FROM alpine:3.19

# Instalar certificados SSL y timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Crear usuario no-root para seguridad
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Establecer directorio de trabajo
WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /app/copa-litoral-backend .

# Copiar archivos de configuración si existen
COPY --from=builder /app/.env* ./

# Cambiar ownership al usuario no-root
RUN chown -R appuser:appgroup /app

# Cambiar al usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Configurar health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando para ejecutar la aplicación
CMD ["./copa-litoral-backend"]
