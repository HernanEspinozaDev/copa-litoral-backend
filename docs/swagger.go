package docs

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// SwaggerSpec contiene la especificación OpenAPI 3.0 completa
var SwaggerSpec = map[string]interface{}{
	"openapi": "3.0.3",
	"info": map[string]interface{}{
		"title":       "Copa Litoral API",
		"description": "API REST para el sistema de gestión del torneo de tenis Copa Litoral",
		"version":     "1.0.0",
		"contact": map[string]interface{}{
			"name":  "Copa Litoral Team",
			"email": "admin@copalitoral.com",
		},
		"license": map[string]interface{}{
			"name": "MIT",
		},
	},
	"servers": []map[string]interface{}{
		{
			"url":         "http://localhost:8089/api/v1",
			"description": "Servidor de desarrollo",
		},
		{
			"url":         "https://api.copalitoral.com/v1",
			"description": "Servidor de producción",
		},
	},
	"paths": map[string]interface{}{
		// Auth endpoints
		"/auth/register": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Authentication"},
				"summary":     "Registrar nuevo usuario",
				"description": "Crea un nuevo usuario en el sistema",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/RegisterRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Usuario creado exitosamente",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/AuthResponse",
								},
							},
						},
					},
					"400": map[string]interface{}{
						"description": "Datos de entrada inválidos",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ErrorResponse",
								},
							},
						},
					},
					"409": map[string]interface{}{
						"description": "Usuario ya existe",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ErrorResponse",
								},
							},
						},
					},
				},
			},
		},
		"/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Authentication"},
				"summary":     "Iniciar sesión",
				"description": "Autentica un usuario y devuelve un token JWT",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/LoginRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Login exitoso",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/AuthResponse",
								},
							},
						},
					},
					"401": map[string]interface{}{
						"description": "Credenciales inválidas",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ErrorResponse",
								},
							},
						},
					},
				},
			},
		},
		// Jugadores endpoints
		"/jugadores": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Jugadores"},
				"summary":     "Listar jugadores",
				"description": "Obtiene una lista paginada de jugadores con filtros opcionales",
				"parameters": []map[string]interface{}{
					{
						"name":        "page",
						"in":          "query",
						"description": "Número de página (empezando en 1)",
						"schema": map[string]interface{}{
							"type":    "integer",
							"minimum": 1,
							"default": 1,
						},
					},
					{
						"name":        "limit",
						"in":          "query",
						"description": "Número de elementos por página",
						"schema": map[string]interface{}{
							"type":    "integer",
							"minimum": 1,
							"maximum": 100,
							"default": 20,
						},
					},
					{
						"name":        "categoria_id",
						"in":          "query",
						"description": "Filtrar por ID de categoría",
						"schema": map[string]interface{}{
							"type": "integer",
						},
					},
					{
						"name":        "estado",
						"in":          "query",
						"description": "Filtrar por estado de participación",
						"schema": map[string]interface{}{
							"type": "string",
							"enum": []string{"Activo", "Eliminado", "Inactivo"},
						},
					},
					{
						"name":        "search",
						"in":          "query",
						"description": "Buscar por nombre o apellido",
						"schema": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Lista de jugadores",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/JugadoresPaginatedResponse",
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Jugadores"},
				"summary":     "Crear jugador",
				"description": "Crea un nuevo jugador (requiere autenticación de admin)",
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/JugadorRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Jugador creado exitosamente",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/JugadorResponse",
								},
							},
						},
					},
					"401": map[string]interface{}{
						"description": "No autorizado",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ErrorResponse",
								},
							},
						},
					},
					"403": map[string]interface{}{
						"description": "Permisos insuficientes",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ErrorResponse",
								},
							},
						},
					},
				},
			},
		},
		// Health endpoints
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Health"},
				"summary":     "Health check completo",
				"description": "Verifica el estado de salud de todos los componentes del sistema",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Sistema saludable",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/HealthResponse",
								},
							},
						},
					},
					"503": map[string]interface{}{
						"description": "Sistema no saludable",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/HealthResponse",
								},
							},
						},
					},
				},
			},
		},
		"/metrics": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Monitoring"},
				"summary":     "Métricas de Prometheus",
				"description": "Endpoint para scraping de métricas de Prometheus",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Métricas en formato Prometheus",
						"content": map[string]interface{}{
							"text/plain": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
	},
	"components": map[string]interface{}{
		"securitySchemes": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
			},
		},
		"schemas": map[string]interface{}{
			// Auth schemas
			"RegisterRequest": map[string]interface{}{
				"type": "object",
				"required": []string{"nombre_usuario", "password", "email"},
				"properties": map[string]interface{}{
					"nombre_usuario": map[string]interface{}{
						"type":        "string",
						"minLength":   3,
						"maxLength":   50,
						"description": "Nombre de usuario único",
					},
					"password": map[string]interface{}{
						"type":        "string",
						"minLength":   6,
						"description": "Contraseña del usuario",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"format":      "email",
						"description": "Email del usuario",
					},
					"rol": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"jugador", "administrador"},
						"default":     "jugador",
						"description": "Rol del usuario en el sistema",
					},
				},
			},
			"LoginRequest": map[string]interface{}{
				"type": "object",
				"required": []string{"nombre_usuario", "password"},
				"properties": map[string]interface{}{
					"nombre_usuario": map[string]interface{}{
						"type":        "string",
						"description": "Nombre de usuario",
					},
					"password": map[string]interface{}{
						"type":        "string",
						"description": "Contraseña del usuario",
					},
				},
			},
			"AuthResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"success": map[string]interface{}{
						"type":        "boolean",
						"description": "Indica si la operación fue exitosa",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Mensaje descriptivo",
					},
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"token": map[string]interface{}{
								"type":        "string",
								"description": "Token JWT para autenticación",
							},
							"user": map[string]interface{}{
								"$ref": "#/components/schemas/Usuario",
							},
						},
					},
				},
			},
			// Jugador schemas
			"JugadorRequest": map[string]interface{}{
				"type": "object",
				"required": []string{"nombre", "apellido"},
				"properties": map[string]interface{}{
					"nombre": map[string]interface{}{
						"type":        "string",
						"maxLength":   255,
						"description": "Nombre del jugador",
					},
					"apellido": map[string]interface{}{
						"type":        "string",
						"maxLength":   255,
						"description": "Apellido del jugador",
					},
					"telefono_wsp": map[string]interface{}{
						"type":        "string",
						"maxLength":   50,
						"description": "Número de WhatsApp",
					},
					"contacto_visible_en_web": map[string]interface{}{
						"type":        "boolean",
						"default":     false,
						"description": "Si el contacto es visible en la web",
					},
					"categoria_id": map[string]interface{}{
						"type":        "integer",
						"description": "ID de la categoría del jugador",
					},
					"club": map[string]interface{}{
						"type":        "string",
						"maxLength":   255,
						"description": "Club del jugador",
					},
				},
			},
			"JugadorResponse": map[string]interface{}{
				"allOf": []map[string]interface{}{
					{"$ref": "#/components/schemas/BaseResponse"},
					{
						"type": "object",
						"properties": map[string]interface{}{
							"data": map[string]interface{}{
								"$ref": "#/components/schemas/Jugador",
							},
						},
					},
				},
			},
			"JugadoresPaginatedResponse": map[string]interface{}{
				"allOf": []map[string]interface{}{
					{"$ref": "#/components/schemas/BaseResponse"},
					{
						"type": "object",
						"properties": map[string]interface{}{
							"data": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/Jugador",
								},
							},
							"pagination": map[string]interface{}{
								"$ref": "#/components/schemas/PaginationInfo",
							},
						},
					},
				},
			},
			// Base schemas
			"BaseResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"success": map[string]interface{}{
						"type":        "boolean",
						"description": "Indica si la operación fue exitosa",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Mensaje descriptivo de la operación",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Timestamp de la respuesta",
					},
				},
			},
			"ErrorResponse": map[string]interface{}{
				"allOf": []map[string]interface{}{
					{"$ref": "#/components/schemas/BaseResponse"},
					{
						"type": "object",
						"properties": map[string]interface{}{
							"error": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"code": map[string]interface{}{
										"type":        "string",
										"description": "Código de error específico",
									},
									"details": map[string]interface{}{
										"type":        "object",
										"description": "Detalles adicionales del error",
									},
								},
							},
						},
					},
				},
			},
			"PaginationInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page": map[string]interface{}{
						"type":        "integer",
						"description": "Página actual",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Elementos por página",
					},
					"total": map[string]interface{}{
						"type":        "integer",
						"description": "Total de elementos",
					},
					"total_pages": map[string]interface{}{
						"type":        "integer",
						"description": "Total de páginas",
					},
					"has_next": map[string]interface{}{
						"type":        "boolean",
						"description": "Si hay página siguiente",
					},
					"has_prev": map[string]interface{}{
						"type":        "boolean",
						"description": "Si hay página anterior",
					},
				},
			},
			// Entity schemas
			"Usuario": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "integer",
					},
					"nombre_usuario": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type": "string",
					},
					"rol": map[string]interface{}{
						"type": "string",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"Jugador": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "integer",
					},
					"nombre": map[string]interface{}{
						"type": "string",
					},
					"apellido": map[string]interface{}{
						"type": "string",
					},
					"telefono_wsp": map[string]interface{}{
						"type": "string",
					},
					"contacto_visible_en_web": map[string]interface{}{
						"type": "boolean",
					},
					"categoria_id": map[string]interface{}{
						"type": "integer",
					},
					"club": map[string]interface{}{
						"type": "string",
					},
					"estado_participacion": map[string]interface{}{
						"type": "string",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"updated_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"HealthResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"healthy", "degraded", "unhealthy"},
					},
					"timestamp": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"uptime": map[string]interface{}{
						"type": "string",
					},
					"version": map[string]interface{}{
						"type": "string",
					},
					"components": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type": "string",
								},
								"component": map[string]interface{}{
									"type": "string",
								},
								"message": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
	},
}

// SwaggerHandler sirve la especificación OpenAPI en formato JSON
func SwaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SwaggerSpec)
}

// SwaggerUIHandler sirve la interfaz de Swagger UI
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Copa Litoral API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/v1/docs/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// RegisterSwaggerRoutes registra las rutas de Swagger en el router
func RegisterSwaggerRoutes(router *mux.Router) {
	router.HandleFunc("/docs", SwaggerUIHandler).Methods("GET")
	router.HandleFunc("/docs/", SwaggerUIHandler).Methods("GET")
	router.HandleFunc("/docs/swagger.json", SwaggerHandler).Methods("GET")
}
