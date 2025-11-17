# --- Builder stage ---------------------------------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# install required tools (git if you fetch private modules)
RUN apk add --no-cache git ca-certificates

# pre-fetch dependencies
ADD go.mod go.sum ./
RUN go mod download

# add the rest of the source
ADD . .

# build static binary for linux
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/fip-controller \
    ./cmd/fip-controller

# --- Runtime stage (hardened, Chainguard static) ---------------------------------------------------------------------
FROM cgr.dev/chainguard/static:latest
WORKDIR /app
# Chainguard static already runs as non-root (uid 65532),
USER 65532:65532
# Copy the statically built binary from the builder stage
COPY --from=builder /app/fip-controller /app/fip-controller
ENTRYPOINT ["/app/fip-controller"]
