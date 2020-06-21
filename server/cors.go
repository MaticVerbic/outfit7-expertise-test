package server

import (
	"net/http"

	"github.com/go-chi/cors"
)

// NewCORS returns Cors struct
func NewCORS() func(next http.Handler) http.Handler {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Core-Api-Key", "X-Core-Accept-Currency", "X-Core-Accept-Language", "X-Accept-Locale", "X-Active-Role"},
		AllowCredentials: true,
	})
	return cors.Handler
}
