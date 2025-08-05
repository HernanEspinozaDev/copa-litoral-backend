# 🐳 Documentación de Despliegue Docker - Copa Litoral Backend

## 📋 Resumen del Despliegue

El backend de Copa Litoral ha sido exitosamente desplegado usando Docker con integración a Nginx Proxy Manager para el dominio `https://apicopalitoral.hotusoft.com/`.

## 🔧 Problemas Resueltos y Cambios Finales

### 1. **Corrección de Puerto (8089 → 8080)**

**Problema:** El backend estaba configurado para usar puerto 8089 pero Nginx Proxy Manager esperaba puerto 8080.

**Solución:**
- Agregado `API_PORT=8080` en `.env.docker`
- Agregado `API_PORT=8080` en `docker-compose.yml`
- Cambiado `env_file` de `.env` a `.env.docker` en docker-compose

**Archivos modificados:**
```bash
# .env.docker
API_PORT=8080

# docker-compose.yml
env_file:
  - .env.docker
environment:
  - API_PORT=8080
```

### 2. **Eliminación de Redirección HTTPS Forzada**

**Problema:** El middleware `HTTPSRedirectMiddleware` causaba bucle de redirección HTTP→HTTPS (Error 502 Bad Gateway).

**Diagnóstico:** 
```bash
curl -v http://copa-litoral-backend:8080/health
# Retornaba: HTTP/1.1 301 Moved Permanently
# Location: https://copa-litoral-backend:8080/health
```

**Solución:** Comentado el middleware en `routes/routes.go`:
```go
// Antes:
r.Use(middlewares.HTTPSRedirectMiddleware(cfg))

// Después:
// r.Use(middlewares.HTTPSRedirectMiddleware(cfg)) // Comentado: Nginx Proxy Manager maneja HTTPS
```

**Resultado:** Ahora retorna `HTTP/1.1 200 OK` con JSON válido.

### 3. **Configuración Final de Docker**

**Dockerfile actualizado:**
```dockerfile
FROM golang:1.24-alpine AS builder
# ... resto de la configuración
```

**docker-compose.yml final:**
```yaml
version: '3.8'

services:
  backend:
    image: copa-litoral-backend:latest
    container_name: copa-litoral-backend
    restart: unless-stopped
    env_file:
      - .env.docker
    environment:
      - PORT=8080
      - API_PORT=8080
      - GIN_MODE=release
      - CORS_ALLOWED_ORIGINS=https://apicopalitoral.hotusoft.com,https://www.apicopalitoral.hotusoft.com
      - ENVIRONMENT=production
    networks:
      - mi-red-proxy
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      timeout: 5s
      retries: 5
      start_period: 30s

  postgres:
    image: postgres:15-alpine
    container_name: copa-litoral-postgres
    restart: unless-stopped
    env_file:
      - .env.docker
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - mi-red-proxy
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:

networks:
  mi-red-proxy:
    external: true
```

## 🚀 Comandos de Despliegue

### Construcción y Despliegue:
```bash
# 1. Construir imagen
docker build -t copa-litoral-backend .

# 2. Levantar servicios
docker compose up -d

# 3. Verificar estado
docker ps

# 4. Verificar logs
docker logs copa-litoral-backend
```

### Pruebas de Funcionamiento:
```bash
# Prueba interna (desde contenedor NPM)
docker exec -it [NPM_CONTAINER_ID] /bin/sh
curl -v http://copa-litoral-backend:8080/health
# Debe retornar: HTTP/1.1 200 OK {"status":"healthy"}

# Prueba externa
curl https://apicopalitoral.hotusoft.com/health
```

## 🎯 Estado Final

✅ **Backend funcionando:** Puerto 8080, HTTP 200 OK
✅ **Base de datos conectada:** PostgreSQL 15 Alpine
✅ **Proxy inverso:** Nginx Proxy Manager maneja HTTPS
✅ **CORS configurado:** Para apicopalitoral.hotusoft.com
✅ **Health checks:** Contenedores monitoreados
✅ **Logs estructurados:** JSON con nivel INFO

## 🧪 Pruebas con API

### 1. Health Check
```bash
GET https://apicopalitoral.hotusoft.com/health
```

### 2. Registro de Usuario Administrador
```bash
POST https://apicopalitoral.hotusoft.com/api/v1/auth/register
Content-Type: application/json

{
  "nombre": "Admin",
  "apellido": "Sistema", 
  "email": "admin@copalitoral.com",
  "password": "Admin123!",
  "telefono": "+54911234567",
  "rol": "admin"
}
```

### 3. Login
```bash
POST https://apicopalitoral.hotusoft.com/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@copalitoral.com",
  "password": "Admin123!"
}
```

### 4. Endpoint Protegido
```bash
GET https://apicopalitoral.hotusoft.com/api/v1/protected/profile
Authorization: Bearer [JWT_TOKEN]
```

## 🔐 Arquitectura de Seguridad

- **HTTPS:** Manejado por Nginx Proxy Manager
- **HTTP interno:** Backend responde por HTTP en red Docker
- **CORS:** Restringido a dominio específico
- **JWT:** Autenticación con tokens seguros
- **Rate limiting:** 100 requests por IP
- **Headers de seguridad:** Configurados en middlewares

## 📝 Variables de Entorno Requeridas

Archivo `.env.docker`:
```bash
# Servidor
PORT=8080
API_PORT=8080
GIN_MODE=release
ENVIRONMENT=production

# Base de datos
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu_password_seguro_aqui
DB_NAME=copa_litoral
DB_SSLMODE=disable

# Seguridad
JWT_SECRET=tu_jwt_secret_muy_seguro_aqui
CORS_ALLOWED_ORIGINS=https://apicopalitoral.hotusoft.com,https://www.apicopalitoral.hotusoft.com

# Logging y métricas
LOG_LEVEL=info
METRICS_ENABLED=true

# Backup
BACKUP_ENABLED=true
BACKUP_SCHEDULE=0 2 * * *
```

## 🎉 Conclusión

El despliegue Docker del backend Copa Litoral está completamente funcional y listo para producción. La API responde correctamente en `https://apicopalitoral.hotusoft.com/` con todas las funcionalidades implementadas y documentadas.
