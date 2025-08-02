package utils

import (
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Registrar validaciones personalizadas
	validate.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	validate.RegisterValidation("safe_string", validateSafeString)
	validate.RegisterValidation("phone", validatePhone)
}

// ValidateStruct valida una estructura usando tags de validación
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// SanitizeString sanitiza una cadena de texto
func SanitizeString(input string) string {
	// Escapar HTML
	sanitized := html.EscapeString(input)
	
	// Remover caracteres de control
	sanitized = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, sanitized)
	
	// Trim espacios
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// SanitizeJSON sanitiza todos los strings en un JSON
func SanitizeJSON(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[key] = SanitizeString(v)
		case map[string]interface{}:
			sanitized[key] = SanitizeJSON(v)
		default:
			sanitized[key] = value
		}
	}
	
	return sanitized
}

// ParseAndValidateJSON parsea y valida JSON de una request
func ParseAndValidateJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Rechazar campos desconocidos
	
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	
	return ValidateStruct(dst)
}

// Validaciones personalizadas

// validateNoSQLInjection verifica que no haya patrones de inyección SQL
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Patrones comunes de inyección SQL
	sqlPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`(?i)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)(or|and)\s+['"].*['"]`,
		`(?i)(\-\-|\#|\/\*)`,
		`(?i)(script|javascript|vbscript)`,
	}
	
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			return false
		}
	}
	
	return true
}

// validateSafeString verifica que sea una cadena segura
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Verificar caracteres peligrosos
	dangerousChars := []string{"<", ">", "&", "\"", "'", "/", "\\"}
	
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return false
		}
	}
	
	return true
}

// validatePhone valida formato de teléfono
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	
	// Patrón para teléfonos (formato flexible)
	phonePattern := `^[\+]?[1-9][\d]{0,15}$`
	matched, _ := regexp.MatchString(phonePattern, phone)
	
	return matched
}

// ValidateAndSanitizeInput valida y sanitiza input de usuario
func ValidateAndSanitizeInput(input string, maxLength int) (string, bool) {
	// Sanitizar primero
	sanitized := SanitizeString(input)
	
	// Verificar longitud
	if len(sanitized) > maxLength {
		return "", false
	}
	
	// Verificar que no esté vacío después de sanitizar
	if strings.TrimSpace(sanitized) == "" {
		return "", false
	}
	
	return sanitized, true
}
