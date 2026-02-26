package handlers

import "net/http"

// Health returns the agent status.
func Health(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
