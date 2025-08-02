# Copa Litoral Backend - Documentación Completa de la API

## 📋 Índice
1. [Resumen del Sistema](#resumen-del-sistema)
2. [Arquitectura](#arquitectura)
3. [Configuración](#configuración)
4. [Modelos de Datos](#modelos-de-datos)
5. [Endpoints de la API](#endpoints-de-la-api)
6. [Autenticación y Autorización](#autenticación-y-autorización)
7. [Middlewares](#middlewares)
8. [Flujo de Trabajo](#flujo-de-trabajo)
9. [Sugerencias de Mejora](#sugerencias-de-mejora)

## 🎯 Resumen del Sistema

Copa Litoral Backend es una API REST desarrollada en Go para la gestión de torneos de tenis. El sistema permite:

- **Gestión de jugadores**: Registro, actualización y consulta de participantes
- **Administración de torneos**: Creación y gestión de competencias
- **Manejo de partidos**: Programación, seguimiento y registro de resultados
- **Sistema de categorías**: Organización por niveles de competencia
- **Autenticación JWT**: Control de acceso basado en roles (administrador/jugador)

### Tecnologías Utilizadas
- **Lenguaje**: Go 1.24.4
- **Router**: Gorilla Mux
- **Base de Datos**: PostgreSQL
- **Autenticación**: JWT (JSON Web Tokens)
- **CORS**: Configurado para múltiples orígenes

## 🏗️ Arquitectura

El proyecto sigue una **arquitectura limpia** con separación clara de responsabilidades:

```
copa-litoral-backend/
├── main.go                 # Punto de entrada de la aplicación
├── config/                 # Configuración de la aplicación
├── database/              # Conexión y configuración de BD
├── models/                # Modelos de datos (structs)
├── handlers/              # Controladores HTTP
├── services/              # Lógica de negocio
├── middlewares/           # Middlewares (auth, CORS, roles)
├── routes/                # Definición de rutas
└── utils/                 # Utilidades y helpers
```

### Flujo de Datos
```
HTTP Request → Router → Middleware → Handler → Service → Database
                ↓
HTTP Response ← JSON Response ← Handler ← Service ← Database
```

## ⚙️ Configuración

### Variables de Entorno (.env)
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=nandev
DB_PASSWORD=Admin1234
DB_NAME=copa_litoral
API_PORT=8089
JWT_SECRET=supersecretkeyforexample
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### Configuración por Defecto
- **Puerto API**: 8089
- **Base de Datos**: PostgreSQL en localhost:5432
- **Timeouts**: Read/Write 15s, Idle 60s
- **CORS**: Configurado para desarrollo local

## 📊 Modelos de Datos

### Usuario
```go
type Usuario struct {
    ID           int       `json:"id"`
    NombreUsuario string   `json:"nombre_usuario"`
    Password     string    `json:"password"`
    Rol          string    `json:"rol"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### Jugador
```go
type Jugador struct {
    ID                    int            `json:"id"`
    Nombre                string         `json:"nombre"`
    Apellido              string         `json:"apellido"`
    TelefonoWSP           sql.NullString `json:"telefono_wsp"`
    ContactoVisibleEnWeb  bool           `json:"contacto_visible_en_web"`
    CategoriaID           sql.NullInt32  `json:"categoria_id"`
    Club                  sql.NullString `json:"club"`
    EstadoParticipacion   string         `json:"estado_participacion"`
    CreatedAt             time.Time      `json:"created_at"`
    UpdatedAt             time.Time      `json:"updated_at"`
}
```

### Partido
```go
type Partido struct {
    ID                  int            `json:"id"`
    TorneoID            int            `json:"torneo_id"`
    CategoriaID         int            `json:"categoria_id"`
    Jugador1ID          int            `json:"jugador1_id"`
    Jugador2ID          int            `json:"jugador2_id"`
    Fase                string         `json:"fase"`
    FechaAgendada       sql.NullTime   `json:"fecha_agendada"`
    HoraAgendada        sql.NullTime   `json:"hora_agendada"`
    PropuestaFechaJ1    sql.NullTime   `json:"propuesta_fecha_j1"`
    PropuestaHoraJ1     sql.NullTime   `json:"propuesta_hora_j1"`
    PropuestaFechaJ2    sql.NullTime   `json:"propuesta_fecha_j2"`
    PropuestaHoraJ2     sql.NullTime   `json:"propuesta_hora_j2"`
    PropuestaAceptadaJ1 bool           `json:"propuesta_aceptada_j1"`
    PropuestaAceptadaJ2 bool           `json:"propuesta_aceptada_j2"`
    Estado              EstadoPartido  `json:"estado"`
    ResultadoSetsJ1     sql.NullInt32  `json:"resultado_sets_j1"`
    ResultadoSetsJ2     sql.NullInt32  `json:"resultado_sets_j2"`
    GanadorID           sql.NullInt32  `json:"ganador_id"`
    PerdedorID          sql.NullInt32  `json:"perdedor_id"`
    ResultadoAprobado   bool           `json:"resultado_aprobado"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`
}
```

### Estados de Partido
```go
const (
    EstadoAgendado    EstadoPartido = "agendado"
    EstadoEnJuego     EstadoPartido = "en_juego"
    EstadoFinalizado  EstadoPartido = "finalizado"
    EstadoCancelado   EstadoPartido = "cancelado"
)
```

### Torneo
```go
type Torneo struct {
    ID          int       `json:"id"`
    Nombre      string    `json:"nombre"`
    Descripcion string    `json:"descripcion"`
    FechaInicio time.Time `json:"fecha_inicio"`
    FechaFin    time.Time `json:"fecha_fin"`
    Estado      string    `json:"estado"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Categoria
```go
type Categoria struct {
    ID     int    `json:"id"`
    Nombre string `json:"nombre"`
}
```

## 🔗 Endpoints de la API

### Rutas Públicas

#### Autenticación
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/register` | Registro de usuario | `{"nombre_usuario": "string", "password": "string"}` | `{"message": "Usuario creado exitosamente"}` |
| POST | `/api/v1/login` | Inicio de sesión | `{"nombre_usuario": "string", "password": "string"}` | `{"token": "jwt_token", "user": {...}}` |

#### Jugadores (Consulta)
| Método | Endpoint | Descripción | Parámetros | Salida |
|--------|----------|-------------|------------|--------|
| GET | `/api/v1/jugadores` | Listar jugadores | - | `[{jugador1}, {jugador2}, ...]` |
| GET | `/api/v1/jugadores/{id}` | Obtener jugador | `id: int` | `{jugador}` |

#### Partidos (Consulta)
| Método | Endpoint | Descripción | Parámetros | Salida |
|--------|----------|-------------|------------|--------|
| GET | `/api/v1/partidos` | Listar partidos | `categoria_id?: int` | `[{partido1}, {partido2}, ...]` |
| GET | `/api/v1/partidos/{id}` | Obtener partido | `id: int` | `{partido}` |

#### Torneos (Consulta)
| Método | Endpoint | Descripción | Parámetros | Salida |
|--------|----------|-------------|------------|--------|
| GET | `/api/v1/torneos` | Listar torneos | - | `[{torneo1}, {torneo2}, ...]` |
| GET | `/api/v1/torneos/{id}` | Obtener torneo | `id: int` | `{torneo}` |

#### Categorías (Consulta)
| Método | Endpoint | Descripción | Parámetros | Salida |
|--------|----------|-------------|------------|--------|
| GET | `/api/v1/categorias` | Listar categorías | - | `[{categoria1}, {categoria2}, ...]` |
| GET | `/api/v1/categorias/{id}` | Obtener categoría | `id: int` | `{categoria}` |

#### Prueba
| Método | Endpoint | Descripción | Salida |
|--------|----------|-------------|--------|
| GET | `/test` | Verificar servidor | `{"message": "Servidor funcionando correctamente"}` |

### Rutas Protegidas (Requieren Autenticación)

#### Funcionalidades de Jugador
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/player/partidos/{id}/propose-time` | Proponer horario | `{"fecha": "2024-01-01", "hora": "14:00"}` | `{"message": "Propuesta enviada"}` |
| POST | `/api/v1/player/partidos/{id}/accept-time` | Aceptar horario | - | `{"message": "Horario aceptado"}` |
| POST | `/api/v1/player/partidos/{id}/report-result` | Reportar resultado | `{"sets_j1": 2, "sets_j2": 1}` | `{"message": "Resultado reportado"}` |

### Rutas de Administrador (Requieren Rol Admin)

#### Gestión de Jugadores
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/admin/jugadores` | Crear jugador | `{jugador_data}` | `{jugador_creado}` |
| PUT | `/api/v1/admin/jugadores/{id}` | Actualizar jugador | `{jugador_data}` | `{jugador_actualizado}` |
| DELETE | `/api/v1/admin/jugadores/{id}` | Eliminar jugador | - | `{"message": "Jugador eliminado"}` |

#### Gestión de Partidos
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/admin/partidos` | Crear partido | `{partido_data}` | `{partido_creado}` |
| PUT | `/api/v1/admin/partidos/{id}` | Actualizar partido | `{partido_data}` | `{partido_actualizado}` |
| DELETE | `/api/v1/admin/partidos/{id}` | Eliminar partido | - | `{"message": "Partido eliminado"}` |
| PUT | `/api/v1/admin/partidos/{id}/approve-result` | Aprobar resultado | - | `{"message": "Resultado aprobado"}` |

#### Gestión de Torneos
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/admin/torneos` | Crear torneo | `{torneo_data}` | `{torneo_creado}` |
| PUT | `/api/v1/admin/torneos/{id}` | Actualizar torneo | `{torneo_data}` | `{torneo_actualizado}` |
| DELETE | `/api/v1/admin/torneos/{id}` | Eliminar torneo | - | `{"message": "Torneo eliminado"}` |

#### Gestión de Categorías
| Método | Endpoint | Descripción | Entrada | Salida |
|--------|----------|-------------|---------|--------|
| POST | `/api/v1/admin/categorias` | Crear categoría | `{categoria_data}` | `{categoria_creada}` |
| PUT | `/api/v1/admin/categorias/{id}` | Actualizar categoría | `{categoria_data}` | `{categoria_actualizada}` |
| DELETE | `/api/v1/admin/categorias/{id}` | Eliminar categoría | - | `{"message": "Categoría eliminada"}` |

## 🔐 Autenticación y Autorización

### Sistema JWT
- **Algoritmo**: HS256
- **Expiración**: Configurable
- **Claims**: user_id, rol, exp, iat

### Roles del Sistema
1. **administrador**: Acceso completo a todas las funcionalidades
2. **jugador**: Acceso limitado a funcionalidades específicas de jugador

### Headers Requeridos
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## 🛡️ Middlewares

### 1. CORS Middleware
- **Función**: Manejo de Cross-Origin Resource Sharing
- **Configuración**: Orígenes permitidos desde variables de entorno
- **Headers**: Permite GET, POST, PUT, DELETE, OPTIONS

### 2. Auth Middleware
- **Función**: Validación de tokens JWT
- **Aplicación**: Rutas protegidas (`/api/v1/admin/*`, `/api/v1/player/*`)
- **Context**: Inyecta `user_id` y `rol` en el contexto de la request

### 3. Role Middleware
- **Función**: Validación de roles específicos
- **Aplicación**: Rutas administrativas
- **Validación**: Verifica que el usuario tenga el rol requerido

## 🔄 Flujo de Trabajo

### 1. Registro y Autenticación
```
1. Usuario se registra → POST /api/v1/register
2. Usuario inicia sesión → POST /api/v1/login
3. Sistema devuelve JWT token
4. Cliente incluye token en headers para requests protegidas
```

### 2. Gestión de Partidos
```
1. Admin crea partido → POST /api/v1/admin/partidos
2. Jugador propone horario → POST /api/v1/player/partidos/{id}/propose-time
3. Otro jugador acepta → POST /api/v1/player/partidos/{id}/accept-time
4. Se juega el partido
5. Jugador reporta resultado → POST /api/v1/player/partidos/{id}/report-result
6. Admin aprueba resultado → PUT /api/v1/admin/partidos/{id}/approve-result
```

### 3. Consultas Públicas
```
- Cualquier usuario puede consultar jugadores, partidos, torneos y categorías
- No requiere autenticación
- Datos públicos del torneo
```

## 🚀 Sugerencias de Mejora

### 1. Seguridad
- **JWT Secret**: Usar variable de entorno en producción (actualmente hardcodeado)
- **Validación de entrada**: Implementar validación más robusta de datos
- **Rate limiting**: Agregar limitación de requests por IP
- **HTTPS**: Forzar conexiones seguras en producción
- **Sanitización**: Validar y sanitizar todos los inputs

### 2. Arquitectura y Código
- **Logging estructurado**: Implementar logging con niveles (logrus/zap)
- **Métricas**: Agregar Prometheus/métricas de performance
- **Health checks**: Endpoint `/health` para monitoreo
- **Graceful shutdown**: Mejorar el cierre controlado del servidor
- **Validación de structs**: Usar tags de validación (go-playground/validator)

### 3. Base de Datos
- **Migraciones**: Sistema de migraciones de BD
- **Connection pooling**: Configurar pool de conexiones
- **Transacciones**: Implementar transacciones para operaciones complejas
- **Índices**: Optimizar consultas con índices apropiados
- **Backup**: Estrategia de respaldo automático

### 4. API y Documentación
- **OpenAPI/Swagger**: Documentación interactiva de la API
- **Versionado**: Estrategia de versionado más robusta
- **Paginación**: Implementar paginación en listados
- **Filtros**: Agregar filtros avanzados en consultas
- **Respuestas consistentes**: Estandarizar formato de respuestas de error

### 5. Testing ✅ IMPLEMENTADO
- **Unit tests**: Pruebas unitarias para servicios y utilidades
- **Integration tests**: Pruebas de integración con BD
- **API tests**: Pruebas end-to-end de endpoints HTTP
- **Mocking**: Sistema completo de mocks para testing
- **Coverage**: Configuración completa de reporte de cobertura
- **Test automation**: Scripts y Makefile para automatización

### 6. DevOps y Deployment
- **Docker**: Containerización de la aplicación
- **CI/CD**: Pipeline de integración continua
- **Environment configs**: Configuraciones por ambiente
- **Monitoring**: Alertas y monitoreo en producción
- **Load balancing**: Preparar para múltiples instancias

### 7. Funcionalidades Adicionales
- **Notificaciones**: Sistema de notificaciones (email/SMS)
- **Estadísticas**: Dashboard con estadísticas del torneo
- **Exportación**: Exportar datos a PDF/Excel
- **Calendario**: Integración con calendarios externos
- **Chat**: Sistema de mensajería entre jugadores

### 8. Performance
- **Caching**: Implementar Redis para cache
- **Compresión**: Gzip para responses
- **Optimización de queries**: Lazy loading, eager loading
- **CDN**: Para archivos estáticos
- **Database sharding**: Para escalabilidad

## 🧪 Mejoras de Testing Implementadas

### Resumen de Implementación

Se ha implementado un sistema completo de testing que incluye pruebas unitarias, de integración, API tests, mocking y coverage reporting. El sistema está diseñado para garantizar la calidad y confiabilidad del código.

### 1. Estructura de Testing

```
tests/
├── config.go              # Configuración y utilidades de testing
├── simple_test.go          # Pruebas básicas de verificación
├── mocks/
│   └── database_mock.go    # Mocks para base de datos y servicios
├── unit/
│   ├── utils_test.go       # Pruebas unitarias de utilidades
│   └── services_test.go    # Pruebas unitarias de servicios
├── integration/
│   └── database_test.go    # Pruebas de integración con BD
└── api/
    └── endpoints_test.go   # Pruebas end-to-end de API
```

### 2. Tipos de Pruebas Implementadas

#### 2.1 Pruebas Unitarias
- **Utilidades**: Paginación, filtros, versionado, respuestas, JWT
- **Servicios**: Mocks de servicios de jugadores, auth, email, HTTP client
- **Benchmarks**: Pruebas de rendimiento para operaciones críticas

#### 2.2 Pruebas de Integración
- **Base de Datos**: Conectividad, CRUD operations, transacciones
- **Migraciones**: Verificación de esquema y datos
- **Connection Pooling**: Pruebas de rendimiento de conexiones
- **Constraints**: Validación de relaciones entre entidades

#### 2.3 Pruebas de API (End-to-End)
- **Health Endpoints**: `/health`, `/health/ready`, `/health/live`
- **Authentication**: Registro, login, validación de tokens
- **CRUD Operations**: Jugadores, partidos, torneos, categorías
- **Pagination & Filtering**: Pruebas de parámetros avanzados
- **Error Handling**: Validación de respuestas de error
- **CORS & Rate Limiting**: Pruebas de middleware

### 3. Sistema de Mocking

#### 3.1 Database Mock
```go
type MockDB struct {
    queries map[string]MockResult
    errors  map[string]error
}
```

#### 3.2 Service Mocks
- **JugadorService**: CRUD operations con validación
- **AuthService**: Login, registro, validación de tokens
- **EmailService**: Envío de notificaciones
- **HTTPClient**: Requests externos

### 4. Coverage y Reporting

#### 4.1 Script de Testing (`scripts/test.sh`)
- Ejecución de todas las pruebas con coverage
- Generación de reportes HTML y texto
- Análisis de código con `go vet` y `golint`
- Benchmarks automáticos
- Reportes consolidados

#### 4.2 Makefile para Automatización
```bash
make test           # Ejecutar todas las pruebas
make test-unit      # Solo pruebas unitarias
make test-integration # Solo pruebas de integración
make test-api       # Solo pruebas de API
make test-coverage  # Generar y abrir reporte de coverage
make test-bench     # Incluir benchmarks
```

### 5. Configuración de Testing

#### 5.1 Variables de Entorno
```bash
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=password
TEST_DB_NAME=copa_litoral_test
```

#### 5.2 Helpers de Testing
- `SetupTestDB()`: Configuración de BD de pruebas
- `CleanupTestData()`: Limpieza entre pruebas
- `CreateTestUser()`: Creación de usuarios de prueba
- `CreateTestJugador()`: Creación de jugadores de prueba
- `AssertEqual()`, `AssertNoError()`: Assertions personalizadas

### 6. Comandos de Testing

#### 6.1 Ejecución Básica
```bash
# Todas las pruebas
go test ./...

# Con coverage
go test -coverprofile=coverage.out ./...

# Solo pruebas rápidas
go test -short ./...

# Con benchmarks
go test -bench=. ./...
```

#### 6.2 Usando Scripts
```bash
# Script completo
./scripts/test.sh

# Solo unitarias
./scripts/test.sh --unit-only

# Con benchmarks
./scripts/test.sh --with-benchmarks

# Sin linting
./scripts/test.sh --no-lint
```

### 7. Reportes Generados

#### 7.1 Coverage Reports
- `coverage/coverage.html`: Reporte visual interactivo
- `coverage/coverage.out`: Datos de coverage para herramientas
- `coverage/coverage_report.txt`: Reporte detallado por función

#### 7.2 Benchmark Reports
- `coverage/benchmarks_unit.txt`: Benchmarks de utilidades
- `coverage/benchmarks_integration.txt`: Benchmarks de BD
- `coverage/benchmarks_api.txt`: Benchmarks de endpoints

#### 7.3 Quality Reports
- `coverage/lint_report.txt`: Sugerencias de golint
- `coverage/test_report.md`: Reporte consolidado en Markdown

### 8. Beneficios Implementados

#### 8.1 Calidad de Código
- **Detección temprana de bugs** mediante pruebas automatizadas
- **Refactoring seguro** con cobertura completa
- **Documentación viva** a través de ejemplos de uso

#### 8.2 Desarrollo Ágil
- **Feedback rápido** con pruebas automatizadas
- **Integración continua** preparada para CI/CD
- **Regression testing** automático

#### 8.3 Mantenibilidad
- **Mocks reutilizables** para diferentes escenarios
- **Helpers de testing** para setup/teardown
- **Benchmarks** para monitoreo de performance

### 9. Próximos Pasos

#### 9.1 Integración CI/CD
- Configurar GitHub Actions para ejecutar pruebas automáticamente
- Reportes de coverage en PRs
- Quality gates basados en coverage mínimo

#### 9.2 Pruebas Adicionales
- Load testing para endpoints críticos
- Security testing para vulnerabilidades
- Contract testing para APIs externas

#### 9.3 Monitoreo
- Métricas de testing en producción
- Alertas basadas en fallos de health checks
- Dashboard de calidad de código

## 📝 Conclusión

El backend de Copa Litoral está bien estructurado y funcional, siguiendo buenas prácticas de Go y arquitectura limpia. Con la implementación completa del sistema de testing, el proyecto ahora cuenta con:

- **Cobertura completa** de pruebas unitarias, integración y API
- **Sistema de mocking** robusto para testing aislado
- **Automatización completa** con scripts y Makefile
- **Reportes detallados** de coverage y calidad
- **Herramientas de CI/CD** listas para implementar

El sistema maneja efectivamente la gestión de torneos de tenis con autenticación JWT, control de roles, y ahora con garantías de calidad a través de testing comprehensivo.

Las mejoras sugeridas se enfocan en seguridad, escalabilidad, mantenibilidad y experiencia del desarrollador. La implementación gradual de estas mejoras permitirá que el sistema evolucione hacia un producto de nivel empresarial.

El código actual es sólido como base y las mejoras propuestas pueden implementarse de forma incremental sin afectar la funcionalidad existente.

## ✅ Mejoras Implementadas

### 🔐 1. Seguridad (COMPLETADO)
Todas las mejoras de seguridad han sido **implementadas exitosamente**:

- ✅ **JWT Secret desde Variables de Entorno**: Eliminado JWT hardcodeado, configuración desde .env
- ✅ **Validación Robusta**: Implementado go-playground/validator con validaciones personalizadas
- ✅ **Rate Limiting**: Sistema inteligente por IP con límites diferenciados
- ✅ **HTTPS en Producción**: Redirección automática y headers de seguridad
- ✅ **Sanitización Completa**: Protección contra XSS e inyección SQL

**Archivos creados**: `utils/validation.go`, `middlewares/rate_limit_middleware.go`, `middlewares/https_middleware.go`
**Documentación completa**: `MEJORAS_SEGURIDAD_IMPLEMENTADAS.md`

### 🏗️ 2. Arquitectura y Código (COMPLETADO)
Todas las mejoras de arquitectura han sido **implementadas exitosamente**:

- ✅ **Logging Estructurado**: Sistema completo con Logrus, niveles configurables
- ✅ **Métricas Prometheus**: Métricas HTTP, BD, auth y rate limiting
- ✅ **Health Checks**: Endpoints `/health`, `/health/ready`, `/health/live`
- ✅ **Graceful Shutdown**: Cierre elegante con timeout configurable
- ✅ **Validación de Structs**: Ya implementada con tags de validación

**Nuevos endpoints de monitoreo**:
- `GET /health` - Estado completo del sistema
- `GET /health/ready` - Readiness check (Kubernetes)
- `GET /health/live` - Liveness check (Kubernetes)
- `GET /metrics` - Métricas para Prometheus

**Archivos creados**: `utils/logger.go`, `utils/metrics.go`, `utils/health.go`
**Variables de entorno nuevas**: `LOG_LEVEL`, `LOG_FILE`, `ENVIRONMENT`

### 📊 Métricas Disponibles
- `http_requests_total` - Total de requests HTTP
- `http_request_duration_seconds` - Duración de requests
- `http_requests_in_flight` - Requests activas
- `db_connections_active/idle` - Conexiones de BD
- `auth_attempts_total` - Intentos de autenticación
- `rate_limit_hits_total` - Hits de rate limiting

### 🔧 Configuración de Producción
```env
# Seguridad
JWT_SECRET=your-super-secret-key-here
ENVIRONMENT=production

# Logging
LOG_LEVEL=warn
LOG_FILE=/var/log/copa-litoral/app.log

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 🗄️ 3. Base de Datos (COMPLETADO)
Todas las mejoras de base de datos han sido **implementadas exitosamente**:

- ✅ **Sistema de Migraciones**: Control de versiones automático con rollback
- ✅ **Connection Pooling**: Pool optimizado por ambiente con métricas
- ✅ **Transacciones Robustas**: Sistema completo con rollback automático
- ✅ **Índices Optimizados**: Consultas mejoradas con índices estratégicos
- ✅ **Backup Automático**: Sistema completo de respaldo y restauración

**Archivos creados**: `database/migrations.go`, `database/pool.go`, `database/transactions.go`, `database/backup.go`
**Migraciones**: Sistema completo en `database/migrations/`
**Health checks**: Monitoreo de BD integrado en `/health`

### 🔧 Configuración de Base de Datos
```env
# Database Pool Configuration
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5    # minutes
DB_CONN_MAX_IDLE_TIME=2   # minutes

# Backup Configuration
BACKUP_ENABLED=true
BACKUP_INTERVAL=24        # hours
BACKUP_RETENTION=7        # backups to keep
BACKUP_DIRECTORY=backups
```

### 📊 Funcionalidades de BD Implementadas
- **Migraciones automáticas** al iniciar la aplicación
- **Pool de conexiones** configurado por ambiente
- **Transacciones con timeout** y rollback automático
- **Backup programado** con compresión y rotación
- **Health checks detallados** con estadísticas del pool
- **Índices optimizados** para consultas frecuentes
