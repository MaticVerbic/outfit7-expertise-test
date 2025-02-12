package middlewares

import (
	"context"
	"expertisetest/config"
	"net/http"
)

// AuthenticationMiddleware sends user/password to be validated down the context.
func AuthenticationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simply check if BasicAuth exists and is okay.
		// Auth and authorization happens on endpoint level.
		// In case of fail return 401.
		user, pass, ok := r.BasicAuth()

		if !ok {
			w.WriteHeader(401)
		}

		r = r.WithContext(context.WithValue(r.Context(), config.UserKey, user))
		r = r.WithContext(context.WithValue(r.Context(), config.PassKey, pass))

		h.ServeHTTP(w, r)
	})
}
