package bridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"clustta/internal/bridge/handlers"
)

const bridgePort = "1173"

var server *http.Server
var bridgeToken string

// Start launches the bridge HTTP server in the background (non-blocking).
// Safe to call multiple times; does nothing if the server is already running.
func Start() {
	if server != nil {
		return
	}

	// Generate a random auth token and write it to a known file
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Failed to generate bridge token: %v", err)
		return
	}
	bridgeToken = hex.EncodeToString(tokenBytes)

	tokenPath, err := getBridgeTokenPath()
	if err != nil {
		log.Printf("Failed to get bridge token path: %v", err)
		return
	}
	os.MkdirAll(filepath.Dir(tokenPath), 0700)
	if err := os.WriteFile(tokenPath, []byte(bridgeToken), 0600); err != nil {
		log.Printf("Failed to write bridge token: %v", err)
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

	// Versioned DCC API
	mux.HandleFunc("GET /v1/capabilities", handlers.V1Capabilities)
	mux.HandleFunc("GET /v1/bootstrap", handlers.V1Bootstrap)
	mux.HandleFunc("GET /v1/context", handlers.V1ResolveContext)
	mux.HandleFunc("GET /v1/projects", handlers.V1ListProjects)
	mux.HandleFunc("GET /v1/projects/{projectId}/workspace", handlers.V1ProjectWorkspace)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets", handlers.V1ListAssignedAssets)
	mux.HandleFunc("GET /v1/projects/{projectId}/statuses", handlers.V1ListStatuses)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets/{assetId}/dependencies", handlers.V1ListDependencies)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/dependencies", handlers.V1CreateDependency)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/dependencies/{edgeId}/selector", handlers.V1UpdateDependencySelector)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets/{assetId}/dependency-options/{dependencyId}", handlers.V1DependencySelectorOptions)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets/{assetId}/checkpoint-tags", handlers.V1ListCheckpointTags)
	mux.HandleFunc("POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags", handlers.V1SetCheckpointTag)
	mux.HandleFunc("POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags/{tagId}", handlers.V1SetCheckpointTag)
	mux.HandleFunc("POST /v1/projects/{projectId}/checkpoint-groups/{groupId}/tags", handlers.V1SetCheckpointTagsForGroup)
	mux.HandleFunc("DELETE /v1/projects/{projectId}/checkpoint-tags/{tagId}", handlers.V1DeleteCheckpointTag)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets/{assetId}/checkpoints", handlers.V1ListCheckpoints)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/checkpoints", handlers.V1CreateCheckpoint)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/status", handlers.V1ChangeStatus)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/open", handlers.V1OpenAsset)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/reveal", handlers.V1RevealAsset)
	mux.HandleFunc("GET /v1/projects/{projectId}/assets/{assetId}/build-plan", handlers.V1DependencyBuildPlan)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/build", handlers.V1BuildAsset)
	mux.HandleFunc("POST /v1/projects/{projectId}/assets/{assetId}/revert", handlers.V1RevertAsset)
	mux.HandleFunc("GET /v1/jobs/{jobId}", handlers.V1GetJob)
	mux.HandleFunc("POST /v1/jobs/{jobId}/cancel", handlers.V1CancelJob)

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
	go handlers.WarmDCCCache()
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

	// Clean up token file
	if tokenPath, err := getBridgeTokenPath(); err == nil {
		os.Remove(tokenPath)
	}
	bridgeToken = ""
	handlers.ResetDCCCache()

	log.Println("Clustta Bridge stopped.")
}

// getBridgeTokenPath returns the path for the bridge auth token file.
func getBridgeTokenPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default:
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, ".config")
		}
		base = appData
	}
	return filepath.Join(base, "Clustta", ".bridge-token"), nil
}

// corsMiddleware validates the bridge auth token and sets CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Clustta-Studio",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Validate auth token (skip for health check)
		if r.URL.Path != "/health" {
			auth := r.Header.Get("Authorization")
			token := strings.TrimPrefix(auth, "Bearer ")
			if token == "" || token != bridgeToken {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
