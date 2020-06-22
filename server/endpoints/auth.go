package endpoints

import (
	"context"
	"expertisetest/config"
	"fmt"
	"net/http"
)

// This function handle the authorization of the clients
func authorize(ctx context.Context, w http.ResponseWriter, entity string) bool {
	// Fetch credentials from authentication middleware.
	user, ok := ctx.Value(config.UserKey).(string)
	if !ok {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("internal system error"), nil)
		return false
	}

	pass, ok := ctx.Value(config.PassKey).(string)
	if !ok {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("internal system error"), nil)
		return false
	}

	// Validate the client.
	c := config.GetInstance()
	switch entity {
	case "admin":
		if !(c.AdminUser == user && c.AdminPass == pass) {
			writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid username or password"), nil)
			return false
		}

		return true
	case "client":
		if !(c.ClientUser == user && c.ClientPass == pass) {
			writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid username or password"), nil)
			return false
		}

		return true
	case "any":
		if !(c.AdminUser == user && c.AdminPass == pass) && !(c.ClientUser == user && c.ClientPass == pass) {
			writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid username or password"), nil)
			return false
		}

		return true
	}

	writeResponse(w, http.StatusUnauthorized, fmt.Sprintf("invalid username or password"), nil)
	return false
}
