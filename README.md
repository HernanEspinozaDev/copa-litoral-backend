# Copa Litoral Backend

Backend en Go para el sistema de gestión de torneos de tenis Copa Litoral.

## Características

- API RESTful con autenticación JWT
- Conexión a PostgreSQL
- Gestión de jugadores, partidos, torneos y categorías
- Sistema de roles (administrador, jugador)
- CORS configurado para frontend
- Manejo de propuestas de horarios entre jugadores
- Reporte y aprobación de resultados

## Requisitos

- Go 1.19 o superior
- PostgreSQL 12 o superior
- Base de datos `copa_litoral` creada

## Instalación

1. Clonar el repositorio:
```bash
git clone <url-del-repositorio>
cd copa-litoral-backend
```

2. Instalar dependencias:
```bash
go mod tidy
```

3. Crear archivo `.env` en la raíz del proyecto:
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

4. Compilar y ejecutar:
```bash
go build -o copa-litoral-backend
./copa-litoral-backend
```

O ejecutar directamente:
```bash
go run main.go
```

## Estructura del Proyecto

```
copa-litoral-backend/
├── config/          # Configuración y variables de entorno
├── database/        # Conexión a la base de datos
├── handlers/        # Controladores HTTP
├── middlewares/     # Middlewares (auth, CORS, roles)
├── models/          # Modelos de datos (structs)
├── routes/          # Definición de rutas
├── services/        # Lógica de negocio
├── utils/           # Utilidades (auth, helpers)
├── main.go          # Punto de entrada
└── go.mod           # Dependencias de Go
```

## Endpoints de la API

### Autenticación (Públicos)
- `POST /api/v1/register` - Registrar nuevo usuario
- `POST /api/v1/login` - Iniciar sesión

### Jugadores (Públicos)
- `GET /api/v1/jugadores` - Obtener todos los jugadores
- `GET /api/v1/jugadores/{id}` - Obtener jugador por ID

### Jugadores (Protegidos - Admin)
- `POST /api/v1/admin/jugadores` - Crear jugador
- `PUT /api/v1/admin/jugadores/{id}` - Actualizar jugador
- `DELETE /api/v1/admin/jugadores/{id}` - Eliminar jugador

### Partidos (Públicos)
- `GET /api/v1/partidos` - Obtener todos los partidos
- `GET /api/v1/partidos/{id}` - Obtener partido por ID

### Partidos (Protegidos - Admin)
- `POST /api/v1/admin/partidos` - Crear partido
- `PUT /api/v1/admin/partidos/{id}` - Actualizar partido
- `DELETE /api/v1/admin/partidos/{id}` - Eliminar partido
- `PUT /api/v1/admin/partidos/{id}/approve-result` - Aprobar resultado

### Partidos (Protegidos - Jugadores)
- `POST /api/v1/player/partidos/{id}/propose-time` - Proponer horario
- `POST /api/v1/player/partidos/{id}/accept-time` - Aceptar horario
- `POST /api/v1/player/partidos/{id}/report-result` - Reportar resultado

### Torneos (Públicos)
- `GET /api/v1/torneos` - Obtener todos los torneos
- `GET /api/v1/torneos/{id}` - Obtener torneo por ID

### Torneos (Protegidos - Admin)
- `POST /api/v1/admin/torneos` - Crear torneo
- `PUT /api/v1/admin/torneos/{id}` - Actualizar torneo
- `DELETE /api/v1/admin/torneos/{id}` - Eliminar torneo

### Categorías (Públicos)
- `GET /api/v1/categorias` - Obtener todas las categorías
- `GET /api/v1/categorias/{id}` - Obtener categoría por ID

### Categorías (Protegidos - Admin)
- `POST /api/v1/admin/categorias` - Crear categoría
- `PUT /api/v1/admin/categorias/{id}` - Actualizar categoría
- `DELETE /api/v1/admin/categorias/{id}` - Eliminar categoría

## Autenticación

Para endpoints protegidos, incluir el header:
```
Authorization: Bearer <token_jwt>
```

## Ejemplos de Uso

### Registrar un usuario administrador
```bash
curl -X POST http://localhost:8089/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "nombre_usuario": "admin",
    "email": "admin@example.com",
    "password": "password123",
    "rol": "administrador"
  }'
```

### Iniciar sesión
```bash
curl -X POST http://localhost:8089/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "nombre_usuario": "admin",
    "password": "password123"
  }'
```

### Crear un jugador (requiere token de admin)
```bash
curl -X POST http://localhost:8089/api/v1/admin/jugadores \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token_jwt>" \
  -d '{
    "nombre": "Juan",
    "apellido": "Perez",
    "categoria_id": 1,
    "club": "Mi Club"
  }'
```

## Desarrollo

Para desarrollo, puedes usar:
```bash
go run main.go
```

El servidor se iniciará en el puerto configurado (por defecto 8089).

## Notas de Producción

- Cambiar `JWT_SECRET` por una clave segura y larga
- Configurar CORS para los dominios de producción
- Usar variables de entorno para todas las configuraciones sensibles
- Implementar logging más robusto
- Agregar tests unitarios e integración 