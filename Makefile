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

PYTHON_SCRIPT := populate.py

# Default target
.PHONY: all
all: build

# Run development environment
.PHONY: dev
dev:
	wails3 dev

# Run development environment
.PHONY: dev-server
dev-server:
ifeq ($(DETECTED_OS),Windows)
	air -c .\.server_air.toml
else
	air -c ./.server_air_darwin.toml
endif

# Run development environment for studio
.PHONY: dev-studio
dev-studio:
ifeq ($(DETECTED_OS),Windows)
	air -c .\.studio_server_air.toml
else
	air -c ./.studio_server_air_darwin.toml
endif

# Build the Clustta Engine project
.PHONY: build
build:

ifeq ($(DETECTED_OS),Windows)
	wails3 package
	powershell -Command "Start-Process 'MsixPackagingTool.exe' -ArgumentList 'create-package','--template','.\Clustta_template.xml','-v' -Verb RunAs"
else ifeq ($(DETECTED_OS),Darwin)
	wails3 package
	bash ./macappstore-build.sh
endif

# Build the server
.PHONY: build-server
build-server:
	@echo "Building Server"
ifeq ($(DETECTED_OS),Windows)
	go build -ldflags "-s -w" -o "..\clustta\$(SERVER_BINARY)" ./cmd/studio_server
else
	go build -ldflags "-s -w" -o "../clustta/$(SERVER_BINARY)" ./cmd/studio_server
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