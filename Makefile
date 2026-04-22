# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY_EXT := .exe
    PATH_SEP := \\
else
    DETECTED_OS := $(shell uname -s)
    BINARY_EXT :=
    PATH_SEP := /
endif

# Set binary names based on OS
ifeq ($(DETECTED_OS),Windows)
    SIDECAR_BINARY := src-tauri$(PATH_SEP)clustta_cli-x86_64-pc-windows-msvc.exe
    SERVER_BINARY := src-tauri$(PATH_SEP)clustta_server-x86_64-pc-windows-msvc.exe
else ifeq ($(DETECTED_OS),Darwin)
    SIDECAR_BINARY := src-tauri/clustta_cli-aarch64-apple-darwin
    SERVER_BINARY := src-tauri/clustta_server-aarch64-apple-darwin
endif

# Default target
.PHONY: all
all: build

# Run development environment
.PHONY: client
client:
	wails3 dev

# Run the dashboard (web) dev server from this repo's frontend/ in web mode
.PHONY: dashboard
dashboard:
	cd frontend && yarn dev:dashboard

# Run the dashboard against the live (production) API.
.PHONY: dashboard-live
dashboard-live:
	cd frontend && yarn live:dashboard

# Sync this client's frontend into the sibling clustta-app repo
# Pass ARGS="-DryRun" or ARGS="-UseGit" to forward flags.
.PHONY: sync
sync:
ifeq ($(DETECTED_OS),Windows)
	powershell -ExecutionPolicy Bypass -File ../clustta-app/scripts/sync-repos.ps1 $(ARGS)
else
	pwsh -File ../clustta-app/scripts/sync-repos.ps1 $(ARGS)
endif

# Regenerate services-contract.json from the Go service layer.
# Commit the result. CI fails if this file is out of date.
.PHONY: contract
contract:
	go run ./cmd/extract-services

# Verify the contract is up to date (CI gate). Fails if `make contract` would
# produce a different file than what's committed.
.PHONY: contract-check
contract-check:
	go run ./cmd/extract-services
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "if ((git status --porcelain services-contract.json) -ne '') { Write-Error 'services-contract.json is out of date. Run: make contract'; exit 1 }"
else
	@if [ -n "$$(git status --porcelain services-contract.json)" ]; then echo 'services-contract.json is out of date. Run: make contract'; exit 1; fi
endif


# Build the Clustta Engine project
.PHONY: build
build:

ifeq ($(DETECTED_OS),Windows)
	wails3 package
	powershell -ExecutionPolicy Bypass -File ./windows-sign.ps1
	powershell -Command "Start-Process 'MsixPackagingTool.exe' -ArgumentList 'create-package','--template','.\Clustta_template.xml','-v' -Verb RunAs"
else ifeq ($(DETECTED_OS),Darwin)
	wails3 package
	bash ./macappstore-build.sh
	bash ./website-build.sh
endif

# Build for development
.PHONY: build-dev
build-dev:
	@echo "Building Dev Engine"
ifeq ($(DETECTED_OS),Windows)
	go build -ldflags "-X 'clustta/internal/constants.host=http://127.0.0.1:5000'" -o "..\clustta\$(SIDECAR_BINARY)" ./cmd/cli
else
	go build -ldflags "-X 'clustta/internal/constants.host=http://127.0.0.1:5000'" -o "../clustta/$(SIDECAR_BINARY)" ./cmd/cli
endif

# install dev dependencies
.PHONY: install-dev
install-dev:
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest

# Build the Clustta Bridge (HTTP server for addon integrations)
.PHONY: bridge
bridge:
	@echo "Building Clustta Bridge"
ifeq ($(DETECTED_OS),Windows)
	wails3 generate syso -arch amd64 -icon build/windows/icon.ico -manifest build/windows/wails.exe.manifest -info build/windows/info.json -out bridge_windows_amd64.syso
	go build -trimpath -ldflags="-w -s" -o bin$(PATH_SEP)clustta-bridge$(BINARY_EXT) ./cmd/bridge
	-powershell -Command "Remove-Item bridge_windows_amd64.syso -ErrorAction SilentlyContinue"
else ifeq ($(DETECTED_OS),Darwin)
	go build -trimpath -ldflags="-w -s" -o bin/clustta-bridge ./cmd/bridge
else
	go build -trimpath -ldflags="-w -s" -o bin/clustta-bridge ./cmd/bridge
endif

# Build the bridge for development (no icon, no trimming)
.PHONY: bridge-dev
bridge-dev:
	@echo "Building Clustta Bridge (dev)"
	go build -o bin$(PATH_SEP)clustta-bridge$(BINARY_EXT) ./cmd/bridge