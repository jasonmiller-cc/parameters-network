# syntax=docker/dockerfile:1
FROM golang:1.22-bookworm AS builder

WORKDIR /workspace

# Copy go module files first for layer caching.
COPY go.mod go.sum ./

# parameters-core is a local replace; copy it alongside.
COPY ../parameters-core /parameters-core

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /bin/parameters-network ./cmd/server

# --- runtime image ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        iproute2 \
        nftables \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/parameters-network /usr/local/bin/parameters-network

EXPOSE 8086

ENTRYPOINT ["/usr/local/bin/parameters-network"]
