# Build stage
FROM golang:1.22-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

# Labels for container metadata
LABEL org.opencontainers.image.title="Harmony Backend"
LABEL org.opencontainers.image.description="Self-hosted music streaming backend API"
LABEL org.opencontainers.image.vendor="Harmony"
LABEL org.opencontainers.image.source="https://github.com/yourusername/harmony"

# Install runtime dependencies
# - ffmpeg: audio transcoding
# - ca-certificates: HTTPS support
# - sqlite-libs: database
# - tzdata: timezone support
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    sqlite-libs \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 harmony && \
    adduser -D -u 1000 -G harmony harmony

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /usr/local/bin/server

# Create required directories with proper permissions
RUN mkdir -p /data /media /app/cache /app/artwork && \
    chown -R harmony:harmony /data /media /app

# Switch to non-root user
USER harmony

# Expose API port
EXPOSE 8080

# Environment defaults
ENV GIN_MODE=release
ENV DB_PATH=/data/harmony.db
ENV MEDIA_PATH=/media
ENV ARTWORK_PATH=/app/artwork
ENV CACHE_PATH=/app/cache

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["server"]
