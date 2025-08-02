package middlewares

import (
	"context"
	"net/http"
	"strings"

	"copa-litoral-backend/config"
	"copa-litoral-backend/utils"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RolKey    contextKey = "rol"
)

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el token del header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Token de autorización requerido")
			return
		}

		// Verificar que el header tenga el formato "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		tokenString := tokenParts[1]

		// Parsear y validar el token JWT
		claims, err := utils.ParseJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Token inválido")
			return
		}

		// Agregar la información del usuario al contexto
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RolKey, claims.Rol)

		// Llamar al siguiente handler con el contexto actualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
	}
}

// Helper functions para obtener datos del contexto
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

func GetRolFromContext(ctx context.Context) (string, bool) {
	rol, ok := ctx.Value(RolKey).(string)
	return rol, ok
} 