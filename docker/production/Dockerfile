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
FROM golang:1.25-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Define build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT=unknown
ARG GO_VERSION

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install build tools
RUN go install go.uber.org/mock/mockgen@latest

# Copy built frontend assets first
COPY --from=frontend-builder /app/dist ./dist

# Copy source code
COPY . .

# Generate code artifacts (mocks)
RUN go generate ./...

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/goformx/goforms/internal/infrastructure/version.Version=${VERSION} -X github.com/goformx/goforms/internal/infrastructure/version.BuildTime=${BUILD_TIME} -X github.com/goformx/goforms/internal/infrastructure/version.GitCommit=${GIT_COMMIT} -X github.com/goformx/goforms/internal/infrastructure/version.GoVersion=${GO_VERSION}" \
    -o bin/goforms main.go

# Production stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

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
