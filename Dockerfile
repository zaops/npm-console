# Build stage
FROM golang:1.21-alpine AS builder

# Install git for version information
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o npm-console .

# Runtime stage
FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache \
    ca-certificates \
    nodejs \
    npm \
    yarn \
    && rm -rf /var/cache/apk/*

# Install pnpm and bun
RUN npm install -g pnpm @antfu/ni
RUN npm install -g bun || echo "Bun installation failed, continuing without it"

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/npm-console /usr/local/bin/npm-console

# Copy web assets
COPY --from=builder /app/web/dist ./web/dist

# Create directories for cache and config
RUN mkdir -p /home/appuser/.cache /home/appuser/.config && \
    chown -R appuser:appgroup /home/appuser /app

# Switch to non-root user
USER appuser

# Expose port for web interface
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD npm-console version || exit 1

# Default command
CMD ["npm-console", "web", "--host", "0.0.0.0", "--port", "8080"]
