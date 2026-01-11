# Development stage
FROM golang:1.25-alpine

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    nodejs \
    npm \
    wget

WORKDIR /app

# Install Go tools
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest \
    && go install github.com/a-h/templ/cmd/templ@latest \
    && go install go.uber.org/mock/mockgen@latest \
    && go install github.com/air-verse/air@latest \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Copy package files for frontend
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Build frontend assets for development
RUN npm run build

# Create necessary directories
RUN mkdir -p logs

# Expose ports
EXPOSE 8090 5173

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run with Air for hot reload
CMD ["air", "-c", ".air.toml"]
