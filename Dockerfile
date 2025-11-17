# --- Builder stage ---------------------------------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install required tools and update ca-certificates
RUN apk add --no-cache git ca-certificates tzdata file && \
    apk upgrade --no-cache

# Pre-fetch dependencies (this layer is cached if go.mod/go.sum don't change)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source
COPY . .

# Build static binary for linux with additional hardening flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -tags netgo,osusergo \
    -o /app/fip-controller \
    ./cmd/fip-controller

# Verify the binary is actually static
RUN file /app/fip-controller | grep -q "statically linked"

# --- Runtime stage (hardened, Chainguard static) ---------------------------------------------------------------------
FROM cgr.dev/chainguard/static:latest

# Set metadata labels
LABEL org.opencontainers.image.title="hetzner-fip-controller" \
      org.opencontainers.image.description="Kubernetes controller for Hetzner Cloud Floating IPs" \
      org.opencontainers.image.source="https://github.com/mpowr-it/hetzner-fip-controller" \
      org.opencontainers.image.url="https://github.com/mpowr-it/hetzner-fip-controller" \
      org.opencontainers.image.authors="Patrick Paechnatz <patrick@mpowr.it>" \
      org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

# Copy the statically built binary from the builder stage
COPY --from=builder --chown=65532:65532 /app/fip-controller /app/fip-controller

# Chainguard static already runs as non-root (uid 65532 = nonroot user)
USER 65532:65532

ENTRYPOINT ["/app/fip-controller"]