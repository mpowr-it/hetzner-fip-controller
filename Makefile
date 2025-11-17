# ------------------------------------------------------------------------------
# Hetzner FIP Controller – Makefile (Devbox/Nix only)
# ------------------------------------------------------------------------------

APP_NAME      := hetzner-fip-controller
PKG_MAIN      := ./cmd/fip-controller
PKG_ALL       := ./...

GO            := go
GOLANGCI_LINT := golangci-lint
GOFUMPT       := gofumpt

# Security tooling (must be available in PATH, e.g. via Devbox/Nix)
TRIVY         ?= trivy
GRYPE         ?= grype

# Build output directory
BIN_DIR       := bin
BINARY        := $(BIN_DIR)/$(APP_NAME)

# Docker settings (local build only, no push)
DOCKER_IMAGE  ?= hetzner-fip-controller
DOCKER_TAG    ?= dev
DOCKER_FILE   ?= ./Dockerfile
DOCKER_CTX    ?= .

# ------------------------------------------------------------------------------
# Meta targets
# ------------------------------------------------------------------------------

.PHONY: all build test run lint fmt tidy vet ci clean tree \
        docker-build docker-run docker-shell docker-shell-builder \
        docker-scan-trivy docker-scan-grype docker-security

all: fmt lint test build

ci: fmt lint test

# ------------------------------------------------------------------------------
# Build targets
# ------------------------------------------------------------------------------

build:
	@mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$(BINARY)" $(PKG_MAIN)
	@echo "✔ Build complete → $(BINARY)"

run:
	$(GO) run $(PKG_MAIN)

# ------------------------------------------------------------------------------
# Quality tools
# ------------------------------------------------------------------------------

fmt:
	$(GOFUMPT) -w .

lint:
	$(GOLANGCI_LINT) run

test:
	$(GO) test -v ./...

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet $(PKG_ALL)

# ------------------------------------------------------------------------------
# Docker (local-only, no push)
# ------------------------------------------------------------------------------

docker-build:
	docker build \
	  --file $(DOCKER_FILE) \
	  --tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
	  $(DOCKER_CTX)
	@echo "✔ Docker image built → $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run:
	docker run --rm -it \
	  $(DOCKER_IMAGE):$(DOCKER_TAG)

# Chainguard static runtime images contain no shell.
# Use docker-shell-builder if you need an interactive shell.
docker-shell:
	@echo "✖ The runtime image '$(DOCKER_IMAGE):$(DOCKER_TAG)' contains no shell (Chainguard static)."
	@echo "→ Use 'make docker-shell-builder' for a debug shell based on golang:1.24-alpine."
	@false

# Debug shell using the builder base image
docker-shell-builder:
	docker run --rm -it \
	  --entrypoint /bin/sh \
	  golang:1.24-alpine

# ------------------------------------------------------------------------------
# Docker Security Scans (local Trivy / Grype)
# ------------------------------------------------------------------------------

# Run a Trivy vulnerability scan on the locally built image.
# Fails on HIGH/CRITICAL vulnerabilities. Unfixed vulnerabilities are ignored.
docker-scan-trivy: docker-build
	$(TRIVY) image \
	  --exit-code 1 \
	  --severity HIGH,CRITICAL \
	  --ignore-unfixed \
	  $(DOCKER_IMAGE):$(DOCKER_TAG)

# Run a Grype vulnerability scan on the locally built image.
# Fails if at least one critical vulnerability is found.
docker-scan-grype: docker-build
	$(GRYPE) $(DOCKER_IMAGE):$(DOCKER_TAG) \
	  --fail-on critical

# Combined security target
docker-security: docker-scan-trivy docker-scan-grype

# ------------------------------------------------------------------------------
# Dev helpers
# ------------------------------------------------------------------------------

tree:
	@echo "Project tree:"
	@find . -maxdepth 4 -not -path "./vendor/*" -not -path "./.devbox/*" -print

# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

clean:
	@echo "Cleaning build output..."
	rm -rf "$(BIN_DIR)"
	@echo "✔ Clean complete"
