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

# Stamp release metadata before packaging.
# Usage:
#   make release-version VERSION=0.4.36
#   make release-version VERSION=0.4.36 RELEASE_DATE=2026-07-02
.PHONY: release-version
release-version:
ifeq ($(strip $(VERSION)),)
	$(error VERSION required. Usage: make release-version VERSION=0.4.36)
endif
	go run ./cmd/release-version -version "$(VERSION)" $(if $(RELEASE_DATE),-date "$(RELEASE_DATE)",)


# Build the Clustta Engine project
.PHONY: build
build:
ifneq ($(strip $(VERSION)),)
	$(MAKE) release-version VERSION=$(VERSION) $(if $(RELEASE_DATE),RELEASE_DATE=$(RELEASE_DATE),)
endif

ifeq ($(DETECTED_OS),Windows)
	wails3 task windows:package PRODUCTION=true
	powershell -ExecutionPolicy Bypass -File ../clustta-deployment/windows/windows-sign.ps1
ifneq ($(strip $(VERSION)),)
	powershell -NoProfile -Command "Copy-Item -LiteralPath './bin/clustta-amd64-installer.exe' -Destination './bin/clustta-$(VERSION)-windows-amd64.exe' -Force"
endif
	powershell -Command "Start-Process 'MsixPackagingTool.exe' -ArgumentList 'create-package','--template','../clustta-deployment/windows/Clustta_template.xml','-v' -Verb RunAs"
else ifeq ($(DETECTED_OS),Darwin)
	wails3 task darwin:package PRODUCTION=true
	bash ../clustta-deployment/darwin/website-build.sh
	wails3 task darwin:package PRODUCTION=true
	bash ../clustta-deployment/darwin/macappstore-build.sh
else ifeq ($(DETECTED_OS),Linux)
	wails3 task linux:package PRODUCTION=true
# 	wails3 task linux:create:flatpak
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

# ─── Flatpak / Flathub ───────────────────────────────────────────────────────

DEPLOYMENT_DIR := ../clustta-deployment
FLATHUB_DIR := ../flathub-clustta
FLATPAK_SCRIPTS := $(DEPLOYMENT_DIR)/linux/flatpak

# Check that required sibling repos exist
.PHONY: flatpak-check
flatpak-check:
	@test -d $(DEPLOYMENT_DIR) || (echo "Error: $(DEPLOYMENT_DIR) not found. Clone clustta-deployment alongside this repo." && exit 1)
	@test -d $(FLATHUB_DIR) || (echo "Error: $(FLATHUB_DIR) not found. Clone the flathub fork alongside this repo." && exit 1)

# Generate go-sources.json and generated-sources.json for offline Flatpak builds
.PHONY: flatpak-sources
flatpak-sources: flatpak-check
	@echo "Generating go-sources.json..."
	python3 $(FLATPAK_SCRIPTS)/generate_go_sources.py . $(FLATHUB_DIR)/go-sources.json
	@echo "Generating generated-sources.json..."
	python3 $(FLATPAK_SCRIPTS)/generate_yarn_sources.py . $(FLATHUB_DIR)/generated-sources.json

# Tag and update the Flathub manifest
.PHONY: flatpak-tag
flatpak-tag: flatpak-check
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION required. Usage: make flatpak-tag VERSION=v0.4.39-flatpak.1"; exit 1; fi
	git tag $(VERSION)
	git push origin $(VERSION)
	@COMMIT=$$(git rev-parse $(VERSION)); \
	sed -i "s|tag: v.*|tag: $(VERSION)|" $(FLATHUB_DIR)/com.clustta.clustta.yml; \
	sed -i "s|commit: .*|commit: $$COMMIT|" $(FLATHUB_DIR)/com.clustta.clustta.yml; \
	echo "Updated manifest: tag=$(VERSION) commit=$$COMMIT"

# Push Flathub changes
.PHONY: flatpak-push
flatpak-push: flatpak-check
	cd $(FLATHUB_DIR) && git add -A && git commit -m "release: update to $(VERSION)" && git push
	@echo ""
	@echo "Done. Comment 'bot, build' on the Flathub PR to trigger CI."

# Full flatpak release pipeline
.PHONY: flatpak
flatpak: flatpak-sources flatpak-tag flatpak-push
