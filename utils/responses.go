package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorCode representa códigos de error estándar
type ErrorCode string

const (
	// Errores de validación
	ErrValidationFailed    ErrorCode = "VALIDATION_FAILED"
	ErrInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrMissingField        ErrorCode = "MISSING_FIELD"
	ErrInvalidFormat       ErrorCode = "INVALID_FORMAT"
	ErrValueOutOfRange     ErrorCode = "VALUE_OUT_OF_RANGE"

	// Errores de autenticación y autorización
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	ErrInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Errores de recursos
	ErrResourceNotFound   ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceExists     ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrResourceConflict   ErrorCode = "RESOURCE_CONFLICT"
	ErrResourceLocked     ErrorCode = "RESOURCE_LOCKED"

	// Errores de base de datos
	ErrDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrDatabaseQuery      ErrorCode = "DATABASE_QUERY_ERROR"
	ErrDatabaseConstraint ErrorCode = "DATABASE_CONSTRAINT_ERROR"
	ErrDatabaseTimeout    ErrorCode = "DATABASE_TIMEOUT"

	// Errores de red y externos
	ErrExternalService    ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	ErrRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Errores del servidor
	ErrInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrNotImplemented     ErrorCode = "NOT_IMPLEMENTED"

	// Errores de API
	ErrUnsupportedVersion ErrorCode = "UNSUPPORTED_VERSION"
	ErrNotAcceptable      ErrorCode = "NOT_ACCEPTABLE"
	ErrMethodNotAllowed   ErrorCode = "METHOD_NOT_ALLOWED"
)

// ErrorDetail contiene detalles específicos del error
type ErrorDetail struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// APIError representa un error de la API
type APIError struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Error     ErrorDetail   `json:"error"`
	Timestamp string        `json:"timestamp"`
	RequestID string        `json:"request_id,omitempty"`
	Path      string        `json:"path,omitempty"`
	Method    string        `json:"method,omitempty"`
	TraceID   string        `json:"trace_id,omitempty"`
}

// ValidationError representa errores de validación múltiples
type ValidationError struct {
	Success   bool                     `json:"success"`
	Message   string                   `json:"message"`
	Errors    []ErrorDetail            `json:"errors"`
	Timestamp string                   `json:"timestamp"`
	RequestID string                   `json:"request_id,omitempty"`
	Path      string                   `json:"path,omitempty"`
	Method    string                   `json:"method,omitempty"`
	Summary   map[string]interface{}   `json:"summary,omitempty"`
}

// SuccessResponse representa una respuesta exitosa estándar
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ResponseManager maneja las respuestas de la API
type ResponseManager struct {
	includeStackTrace bool
	includeRequestID  bool
	defaultMessages   map[ErrorCode]string
}

// NewResponseManager crea un nuevo manager de respuestas
func NewResponseManager(includeStackTrace, includeRequestID bool) *ResponseManager {
	rm := &ResponseManager{
		includeStackTrace: includeStackTrace,
		includeRequestID:  includeRequestID,
		defaultMessages:   make(map[string]string),
	}
	
	rm.setupDefaultMessages()
	return rm
}

// setupDefaultMessages configura mensajes por defecto para códigos de error
func (rm *ResponseManager) setupDefaultMessages() {
	rm.defaultMessages = map[ErrorCode]string{
		ErrValidationFailed:    "Los datos proporcionados no son válidos",
		ErrInvalidInput:        "La entrada proporcionada no es válida",
		ErrMissingField:        "Campo requerido faltante",
		ErrInvalidFormat:       "El formato del campo no es válido",
		ErrValueOutOfRange:     "El valor está fuera del rango permitido",
		
		ErrUnauthorized:        "No autorizado para acceder a este recurso",
		ErrForbidden:          "Acceso prohibido a este recurso",
		ErrInvalidToken:       "Token de autenticación inválido",
		ErrTokenExpired:       "Token de autenticación expirado",
		ErrInvalidCredentials: "Credenciales inválidas",
		
		ErrResourceNotFound:   "Recurso no encontrado",
		ErrResourceExists:     "El recurso ya existe",
		ErrResourceConflict:   "Conflicto con el estado actual del recurso",
		ErrResourceLocked:     "El recurso está bloqueado",
		
		ErrDatabaseConnection: "Error de conexión a la base de datos",
		ErrDatabaseQuery:      "Error en la consulta a la base de datos",
		ErrDatabaseConstraint: "Error de restricción en la base de datos",
		ErrDatabaseTimeout:    "Timeout en la consulta a la base de datos",
		
		ErrExternalService:    "Error en servicio externo",
		ErrNetworkTimeout:     "Timeout de red",
		ErrRateLimitExceeded:  "Límite de velocidad excedido",
		
		ErrInternalServer:     "Error interno del servidor",
		ErrServiceUnavailable: "Servicio no disponible",
		ErrNotImplemented:     "Funcionalidad no implementada",
		
		ErrUnsupportedVersion: "Versión de API no soportada",
		ErrNotAcceptable:      "Tipo de contenido no aceptable",
		ErrMethodNotAllowed:   "Método HTTP no permitido",
	}
}

// WriteErrorResponse escribe una respuesta de error
func (rm *ResponseManager) WriteErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode ErrorCode, message string, details map[string]interface{}) {
	// Usar mensaje por defecto si no se proporciona
	if message == "" {
		if defaultMsg, exists := rm.defaultMessages[errorCode]; exists {
			message = defaultMsg
		} else {
			message = "Ha ocurrido un error"
		}
	}

	errorDetail := ErrorDetail{
		Code:    errorCode,
		Message: message,
		Details: details,
	}

	// Agregar stack trace en desarrollo
	if rm.includeStackTrace && (statusCode >= 500 || errorCode == ErrInternalServer) {
		if details == nil {
			errorDetail.Details = make(map[string]interface{})
		}
		errorDetail.Details["stack_trace"] = rm.getStackTrace()
	}

	apiError := APIError{
		Success:   false,
		Message:   message,
		Error:     errorDetail,
		Timestamp: GetCurrentTimestamp(),
		Path:      r.URL.Path,
		Method:    r.Method,
	}

	// Agregar Request ID si está habilitado
	if rm.includeRequestID {
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			apiError.RequestID = requestID
		}
	}

	rm.writeJSONResponse(w, statusCode, apiError)
}

// WriteValidationErrors escribe errores de validación múltiples
func (rm *ResponseManager) WriteValidationErrors(w http.ResponseWriter, r *http.Request, errors []ErrorDetail) {
	validationError := ValidationError{
		Success:   false,
		Message:   "Errores de validación encontrados",
		Errors:    errors,
		Timestamp: GetCurrentTimestamp(),
		Path:      r.URL.Path,
		Method:    r.Method,
		Summary: map[string]interface{}{
			"total_errors": len(errors),
			"error_fields": rm.getErrorFields(errors),
		},
	}

	// Agregar Request ID si está habilitado
	if rm.includeRequestID {
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			validationError.RequestID = requestID
		}
	}

	rm.writeJSONResponse(w, http.StatusBadRequest, validationError)
}

// WriteSuccessResponse escribe una respuesta exitosa
func (rm *ResponseManager) WriteSuccessResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string, data interface{}, meta interface{}) {
	if message == "" {
		message = "Operación completada exitosamente"
	}

	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: GetCurrentTimestamp(),
	}

	// Agregar Request ID si está habilitado
	if rm.includeRequestID {
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			response.RequestID = requestID
		}
	}

	rm.writeJSONResponse(w, statusCode, response)
}

// writeJSONResponse escribe una respuesta JSON
func (rm *ResponseManager) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Si falla la codificación JSON, enviar error básico
		http.Error(w, `{"success":false,"message":"Error encoding response"}`, http.StatusInternalServerError)
	}
}

// getStackTrace obtiene el stack trace actual
func (rm *ResponseManager) getStackTrace() []string {
	var stackTrace []string
	for i := 2; i < 10; i++ { // Saltar las primeras 2 llamadas (esta función y la que la llamó)
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// Simplificar la ruta del archivo
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}
		
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stackTrace
}

// getErrorFields extrae los campos con errores
func (rm *ResponseManager) getErrorFields(errors []ErrorDetail) []string {
	fields := make([]string, 0, len(errors))
	fieldSet := make(map[string]bool)
	
	for _, err := range errors {
		if err.Field != "" && !fieldSet[err.Field] {
			fields = append(fields, err.Field)
			fieldSet[err.Field] = true
		}
	}
	
	return fields
}

// Helper functions para respuestas comunes

// BadRequest respuesta para errores 400
func BadRequest(w http.ResponseWriter, r *http.Request, message string, details map[string]interface{}) {
	rm := NewResponseManager(false, true)
	rm.WriteErrorResponse(w, r, http.StatusBadRequest, ErrInvalidInput, message, details)
}

// Unauthorized respuesta para errores 401
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	rm := NewResponseManager(false, true)
	rm.WriteErrorResponse(w, r, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

// Forbidden respuesta para errores 403
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	rm := NewResponseManager(false, true)
	rm.WriteErrorResponse(w, r, http.StatusForbidden, ErrForbidden, message, nil)
}

// NotFound respuesta para errores 404
func NotFound(w http.ResponseWriter, r *http.Request, resource string) {
	rm := NewResponseManager(false, true)
	message := fmt.Sprintf("%s no encontrado", resource)
	rm.WriteErrorResponse(w, r, http.StatusNotFound, ErrResourceNotFound, message, nil)
}

// Conflict respuesta para errores 409
func Conflict(w http.ResponseWriter, r *http.Request, message string, details map[string]interface{}) {
	rm := NewResponseManager(false, true)
	rm.WriteErrorResponse(w, r, http.StatusConflict, ErrResourceConflict, message, details)
}

// InternalServerError respuesta para errores 500
func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	rm := NewResponseManager(true, true) // Incluir stack trace en errores 500
	details := map[string]interface{}{
		"error": err.Error(),
	}
	rm.WriteErrorResponse(w, r, http.StatusInternalServerError, ErrInternalServer, "", details)
}

// Success respuesta exitosa genérica
func Success(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	rm := NewResponseManager(false, true)
	rm.WriteSuccessResponse(w, r, http.StatusOK, message, data, nil)
}

// Created respuesta para recursos creados (201)
func Created(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	rm := NewResponseManager(false, true)
	rm.WriteSuccessResponse(w, r, http.StatusCreated, message, data, nil)
}

// NoContent respuesta sin contenido (204)
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// ValidationErrors respuesta para errores de validación
func ValidationErrors(w http.ResponseWriter, r *http.Request, errors []ErrorDetail) {
	rm := NewResponseManager(false, true)
	rm.WriteValidationErrors(w, r, errors)
}

// CreateValidationError crea un error de validación
func CreateValidationError(field, message string, value interface{}, constraints map[string]interface{}) ErrorDetail {
	return ErrorDetail{
		Code:        ErrValidationFailed,
		Message:     message,
		Field:       field,
		Value:       value,
		Constraints: constraints,
	}
}

// CreateFieldError crea un error de campo específico
func CreateFieldError(field string, code ErrorCode, message string, value interface{}) ErrorDetail {
	return ErrorDetail{
		Code:    code,
		Message: message,
		Field:   field,
		Value:   value,
	}
}

// GetCurrentTimestamp devuelve el timestamp actual en formato ISO
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// WriteJSONResponse función auxiliar para escribir respuestas JSON
func WriteJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
