package server

import (
	"net/http"

	"github.com/go-chi/cors"
)

// NewCORS returns Cors struct
func NewCORS() func(next http.Handler) http.Handler {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	return cors.Handler
}
