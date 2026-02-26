package bridge

import (
	"context"
	"log"
	"net/http"
	"time"

	"clustta/internal/bridge/handlers"
)

const bridgePort = "1173"

var server *http.Server

// Start launches the bridge HTTP server in the background (non-blocking).
// Safe to call multiple times; does nothing if the server is already running.
func Start() {
	if server != nil {
		return
	}
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", handlers.Health)

	// Accounts
	mux.HandleFunc("GET /accounts", handlers.ListAccounts)
	mux.HandleFunc("GET /accounts/active", handlers.GetActiveAccount)
	mux.HandleFunc("POST /accounts/switch", handlers.SwitchAccount)

	// Studios
	mux.HandleFunc("GET /studios", handlers.ListStudios)
	mux.HandleFunc("GET /studios/active", handlers.GetActiveStudio)
	mux.HandleFunc("POST /studios/switch", handlers.SwitchStudio)

	// Projects
	mux.HandleFunc("GET /projects", handlers.ListProjects)
	mux.HandleFunc("GET /projects/active", handlers.GetActiveProject)
	mux.HandleFunc("POST /projects/switch", handlers.SwitchProject)

	// Assets
	mux.HandleFunc("GET /assets", handlers.ListAssets)

	// Checkpoints
	mux.HandleFunc("GET /assets/{assetId}/checkpoints", handlers.ListCheckpoints)

	handler := corsMiddleware(mux)

	server = &http.Server{
		Addr:         "127.0.0.1:" + bridgePort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Clustta Bridge listening on 127.0.0.1:%s", bridgePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Bridge server error: %v", err)
		}
	}()
}

// Stop gracefully shuts down the bridge server.
func Stop() {
	if server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Bridge shutdown error: %v", err)
	}
	server = nil
	log.Println("Clustta Bridge stopped.")
}

// corsMiddleware allows localhost origins for DCC addon requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
