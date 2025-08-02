package middlewares

import (
	"net/http"

	"copa-litoral-backend/utils"
)

func RoleMiddleware(requiredRoles []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el rol del usuario del contexto
		userRol, ok := GetRolFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Información de usuario no disponible")
			return
		}

		// Verificar si el rol del usuario está en la lista de roles requeridos
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRol == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			utils.RespondWithError(w, http.StatusForbidden, "Acceso denegado: permisos insuficientes")
			return
		}

		// Si el rol es correcto, continuar con el siguiente handler
		next.ServeHTTP(w, r)
	})
} 