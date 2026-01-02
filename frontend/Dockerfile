# Build stage
FROM node:20-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG PUBLIC_API_URL=/api/v1

WORKDIR /app

# Install build dependencies for native modules (if needed)
RUN apk add --no-cache python3 make g++

# Copy package files for better caching
COPY package*.json ./

# Install all dependencies (including dev)
RUN npm ci

# Copy source code
COPY . .

# Set build-time environment variables
ENV PUBLIC_API_URL=${PUBLIC_API_URL}
ENV NODE_ENV=production

# Build the application
RUN npm run build

# Prune dev dependencies
RUN npm prune --production

# Runtime stage
FROM node:20-alpine AS runtime

# Labels for container metadata
LABEL org.opencontainers.image.title="Harmony Frontend"
LABEL org.opencontainers.image.description="Self-hosted music streaming web interface"
LABEL org.opencontainers.image.vendor="Harmony"
LABEL org.opencontainers.image.source="https://github.com/yourusername/harmony"

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# Copy built application from builder (use existing node user from base image)
COPY --from=builder --chown=node:node /app/build ./build
COPY --from=builder --chown=node:node /app/package.json ./
COPY --from=builder --chown=node:node /app/node_modules ./node_modules

# Switch to non-root user (node user already exists in node:alpine)
USER node

# Expose port
EXPOSE 3000

# Environment variables
ENV NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0
ENV ORIGIN=http://localhost:3000
ENV BODY_SIZE_LIMIT=Infinity

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:3000 || exit 1

# Use dumb-init as PID 1 for proper signal handling
ENTRYPOINT ["dumb-init", "--"]

# Run the server
CMD ["node", "build"]
