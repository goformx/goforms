#!/bin/bash

# Middleware System Deployment Script
# This script helps deploy and monitor the new middleware architecture

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-"staging"}
ENABLE_NEW_SYSTEM=${2:-"false"}

echo -e "${BLUE}🚀 Middleware System Deployment${NC}"
echo -e "${BLUE}Environment: ${ENVIRONMENT}${NC}"
echo -e "${BLUE}Enable New System: ${ENABLE_NEW_SYSTEM}${NC}"
echo ""

# Function to print status
print_status() {
    local status=$1
    local message=$2
    case $status in
        "success")
            echo -e "${GREEN}✅ ${message}${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  ${message}${NC}"
            ;;
        "error")
            echo -e "${RED}❌ ${message}${NC}"
            ;;
        "info")
            echo -e "${BLUE}ℹ️  ${message}${NC}"
            ;;
    esac
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Pre-deployment checks
print_status "info" "Running pre-deployment checks..."

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_status "error" "Not in GoForms project directory"
    exit 1
fi

# Check if task is available
if ! command_exists task; then
    print_status "warning" "Task runner not found. Install with: go install github.com/go/task/cmd/task@latest"
fi

# Check if Go is available
if ! command_exists go; then
    print_status "error" "Go is not installed"
    exit 1
fi

print_status "success" "Pre-deployment checks passed"

# Set environment variables
print_status "info" "Setting environment variables..."

if [ "$ENABLE_NEW_SYSTEM" = "true" ]; then
    export MIDDLEWARE_USE_NEW_SYSTEM=true
    print_status "success" "New middleware system enabled"
else
    export MIDDLEWARE_USE_NEW_SYSTEM=false
    print_status "info" "Using legacy middleware system"
fi

# Set environment-specific variables
case $ENVIRONMENT in
    "development")
        export GO_ENV=development
        export MIDDLEWARE_ENABLE_LOGGING=true
        export MIDDLEWARE_ENABLE_DEBUG=true
        ;;
    "staging")
        export GO_ENV=staging
        export MIDDLEWARE_ENABLE_LOGGING=true
        export MIDDLEWARE_ENABLE_DEBUG=false
        ;;
    "production")
        export GO_ENV=production
        export MIDDLEWARE_ENABLE_LOGGING=true
        export MIDDLEWARE_ENABLE_DEBUG=false
        ;;
    *)
        print_status "error" "Invalid environment: $ENVIRONMENT"
        exit 1
        ;;
esac

print_status "success" "Environment variables set for $ENVIRONMENT"

# Run tests
print_status "info" "Running middleware tests..."

if command_exists task; then
    if task test:middleware 2>/dev/null; then
        print_status "success" "Middleware tests passed"
    else
        print_status "warning" "Middleware tests failed or not configured"
    fi
else
    # Fallback to go test
    if go test ./internal/application/middleware/... -v; then
        print_status "success" "Middleware tests passed"
    else
        print_status "warning" "Some middleware tests failed"
    fi
fi

# Build the application
print_status "info" "Building application..."

if command_exists task; then
    if task build 2>/dev/null; then
        print_status "success" "Application built successfully"
    else
        print_status "warning" "Task build failed, trying go build"
        if go build -o bin/goforms .; then
            print_status "success" "Application built successfully"
        else
            print_status "error" "Build failed"
            exit 1
        fi
    fi
else
    if go build -o bin/goforms .; then
        print_status "success" "Application built successfully"
    else
        print_status "error" "Build failed"
        exit 1
    fi
fi

# Run database migrations
print_status "info" "Running database migrations..."

if command_exists task; then
    if task migrate:up 2>/dev/null; then
        print_status "success" "Database migrations completed"
    else
        print_status "warning" "Database migrations failed or not configured"
    fi
fi

# Start the application
print_status "info" "Starting application..."

# Create logs directory if it doesn't exist
mkdir -p logs

# Start the application in the background
if [ -f "bin/goforms" ]; then
    ./bin/goforms > logs/app.log 2>&1 &
    APP_PID=$!
    print_status "success" "Application started with PID: $APP_PID"
else
    print_status "error" "Application binary not found"
    exit 1
fi

# Wait a moment for the application to start
sleep 3

# Check if the application is running
if kill -0 $APP_PID 2>/dev/null; then
    print_status "success" "Application is running"
else
    print_status "error" "Application failed to start"
    echo "Check logs/logs/app.log for details"
    exit 1
fi

# Health check
print_status "info" "Performing health check..."

# Wait for the application to be ready
for i in {1..30}; do
    if curl -f http://localhost:8090/health 2>/dev/null >/dev/null; then
        print_status "success" "Health check passed"
        break
    fi

    if [ $i -eq 30 ]; then
        print_status "error" "Health check failed after 30 attempts"
        kill $APP_PID 2>/dev/null || true
        exit 1
    fi

    sleep 1
done

# Get middleware status
print_status "info" "Checking middleware system status..."

MIDDLEWARE_STATUS=$(curl -s http://localhost:8090/api/v1/middleware/status 2>/dev/null || echo "{}")

if echo "$MIDDLEWARE_STATUS" | grep -q "new_system_enabled"; then
    print_status "success" "Middleware status endpoint accessible"
    echo "$MIDDLEWARE_STATUS" | jq . 2>/dev/null || echo "$MIDDLEWARE_STATUS"
else
    print_status "warning" "Middleware status endpoint not available"
fi

# Performance monitoring
print_status "info" "Setting up performance monitoring..."

# Create monitoring script
cat > scripts/monitor-middleware.sh << 'EOF'
#!/bin/bash

# Middleware Performance Monitoring Script

echo "🔍 Middleware Performance Monitor"
echo "=================================="

# Get middleware performance metrics
PERF_METRICS=$(curl -s http://localhost:8090/api/v1/middleware/performance 2>/dev/null || echo "{}")

if echo "$PERF_METRICS" | grep -q "chain_build_times"; then
    echo "📊 Chain Building Performance:"
    echo "$PERF_METRICS" | jq -r '.chain_build_times | to_entries[] | "  \(.key): \(.value)ms"' 2>/dev/null || echo "$PERF_METRICS"
else
    echo "⚠️  Performance metrics not available"
fi

# Get system metrics
echo ""
echo "📈 System Metrics:"
echo "  Memory Usage: $(ps -o rss= -p $(pgrep goforms) | awk '{print $1/1024 " MB"}')"
echo "  CPU Usage: $(ps -o %cpu= -p $(pgrep goforms))%"
echo "  Uptime: $(ps -o etime= -p $(pgrep goforms))"

# Get recent logs
echo ""
echo "📝 Recent Logs (last 10 lines):"
tail -n 10 logs/app.log 2>/dev/null || echo "  No logs available"
EOF

chmod +x scripts/monitor-middleware.sh

print_status "success" "Performance monitoring script created: scripts/monitor-middleware.sh"

# Create rollback script
cat > scripts/rollback-middleware.sh << 'EOF'
#!/bin/bash

# Middleware System Rollback Script

echo "🔄 Rolling back middleware system..."

# Stop the application
APP_PID=$(pgrep goforms)
if [ ! -z "$APP_PID" ]; then
    echo "Stopping application (PID: $APP_PID)..."
    kill $APP_PID
    sleep 2
fi

# Disable new system
export MIDDLEWARE_USE_NEW_SYSTEM=false

# Restart with legacy system
echo "Restarting with legacy middleware system..."
./bin/goforms > logs/app.log 2>&1 &
NEW_PID=$!

echo "Application restarted with PID: $NEW_PID"
echo "Legacy middleware system activated"
echo "Check logs/app.log for details"
EOF

chmod +x scripts/rollback-middleware.sh

print_status "success" "Rollback script created: scripts/rollback-middleware.sh"

# Final status
echo ""
echo -e "${GREEN}🎉 Deployment Complete!${NC}"
echo ""
echo -e "${BLUE}📋 Next Steps:${NC}"
echo "1. Monitor the application: ./scripts/monitor-middleware.sh"
echo "2. Test API endpoints: curl http://localhost:8090/api/v1/health"
echo "3. Check middleware status: curl http://localhost:8090/api/v1/middleware/status"
echo "4. View logs: tail -f logs/app.log"
echo ""
echo -e "${BLUE}🔄 Rollback (if needed):${NC}"
echo "./scripts/rollback-middleware.sh"
echo ""
echo -e "${BLUE}📊 Performance Monitoring:${NC}"
echo "./scripts/monitor-middleware.sh"
echo ""

# Save PID for later use
echo $APP_PID > .app.pid
print_status "info" "Application PID saved to .app.pid"

print_status "success" "Deployment completed successfully!"
