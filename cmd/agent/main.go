package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clustta/cmd/agent/handlers"

	_ "github.com/mattn/go-sqlite3"
)

const agentPort = "1173"

func main() {
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

	server := &http.Server{
		Addr:    "127.0.0.1:" + agentPort,
		Handler: handler,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Clustta Agent listening on 127.0.0.1:%s\n", agentPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Agent server error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down agent...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Agent shutdown error: %v", err)
	}
	fmt.Println("Agent stopped.")
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
