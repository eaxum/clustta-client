package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// jsonResponse writes a JSON response with the given status code.
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// jsonError writes a JSON error response.
func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func refreshRequested(r *http.Request) bool {
	value := strings.ToLower(r.URL.Query().Get("refresh"))
	return value == "1" || value == "true"
}

// decodeBody decodes a JSON request body into the target struct.
func decodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
