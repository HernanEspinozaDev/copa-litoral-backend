package middlewares

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)
} 