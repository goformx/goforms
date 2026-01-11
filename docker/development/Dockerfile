# Development stage
FROM golang:1.25-alpine

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    nodejs \
    npm \
    wget \
    go-task-task

WORKDIR /app

# Add Go tools to PATH
ENV PATH=$PATH:/go/bin:/root/go/bin:/usr/local/bin

# Install Go tools
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest \
    && go install github.com/a-h/templ/cmd/templ@latest \
    && go install go.uber.org/mock/mockgen@latest \
    && go install github.com/air-verse/air@latest \
    && go install github.com/bombsimon/wsl/v4/cmd/wsl@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/securego/gosec/v2/cmd/gosec@latest \
    && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy package files for frontend
COPY package*.json ./

# Create public/fonts directory before npm install to avoid postinstall script issues
RUN mkdir -p public/fonts

# Install dependencies, skipping problematic postinstall scripts that require git
RUN npm ci --ignore-scripts

# Manually copy fonts (equivalent to postinstall script)
RUN cp -r node_modules/@formio/js/dist/fonts/* public/fonts/ 2>/dev/null || true

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
