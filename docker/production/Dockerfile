# Build stage
FROM node:20-alpine AS frontend-builder

# Install git for npm packages that require it
RUN apk add --no-cache git

WORKDIR /app

# Copy package files
COPY package*.json ./

# Create public/fonts directory before npm install to avoid postinstall script issues
RUN mkdir -p public/fonts

# Install all dependencies, skipping problematic postinstall scripts
RUN npm ci --ignore-scripts

# Manually copy fonts (equivalent to postinstall script)
RUN cp -r node_modules/@formio/js/dist/fonts/* public/fonts/

# Copy frontend source
COPY src/ ./src/
COPY vite.config.ts tsconfig.json postcss.config.cjs ./

# Build frontend assets
RUN npm run build

# Go build stage
FROM golang:1.24-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend assets
COPY --from=frontend-builder /app/dist ./dist

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/goformx/goforms/internal/infrastructure/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X github.com/goformx/goforms/internal/infrastructure/version.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -X github.com/goformx/goforms/internal/infrastructure/version.GitCommit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown') -X github.com/goformx/goforms/internal/infrastructure/version.GoVersion=$(go version | cut -d ' ' -f 3)" \
    -o bin/goforms main.go

# Production stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S goforms && \
    adduser -u 1001 -S goforms -G goforms

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/bin/goforms ./bin/goforms

# Copy built frontend assets
COPY --from=go-builder /app/dist ./dist

# Create necessary directories
RUN mkdir -p logs && \
    chown -R goforms:goforms /app

# Switch to non-root user
USER goforms

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run the application
CMD ["./bin/goforms"] 