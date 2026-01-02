# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ffmpeg ca-certificates sqlite-libs

# Create non-root user
RUN adduser -D -u 1000 harmony

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /usr/local/bin/server

# Create data directory
RUN mkdir -p /data /media && chown -R harmony:harmony /data /media

USER harmony

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["server"]
