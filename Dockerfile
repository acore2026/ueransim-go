# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy the Go source code
# We assume the build context is the root of the repository
COPY go/go.mod go/go.sum ./
RUN go mod download

COPY go/ .

# Build the binaries
RUN go build -o nr-gnb ./cmd/nr-gnb
RUN go build -o nr-ue ./cmd/nr-ue

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies (standard for 5G/Networking)
RUN apk add --no-cache iproute2 bash iputils ca-certificates

WORKDIR /ueransim

# Copy binaries from the builder stage
COPY --from=builder /app/nr-gnb .
COPY --from=builder /app/nr-ue .

# Create a default config directory
RUN mkdir -p ./config

# The entrypoint can be overridden in docker-compose.
# By default, we provide a usage hint or run gNB.
ENTRYPOINT ["./nr-gnb"]
