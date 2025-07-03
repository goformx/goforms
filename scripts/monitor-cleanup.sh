#!/bin/bash

# Middleware Cleanup Monitoring Script
# This script helps monitor system health during legacy code cleanup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Middleware Cleanup Monitor${NC}"
echo "=================================="

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

# Check if application is running
print_status "info" "Checking application status..."

APP_PID=$(pgrep goforms 2>/dev/null || echo "")
if [ -z "$APP_PID" ]; then
    print_status "error" "Application is not running"
    echo "Start the application first: ./scripts/deploy-middleware.sh"
    exit 1
else
    print_status "success" "Application is running (PID: $APP_PID)"
fi

# Check middleware system status
print_status "info" "Checking middleware system status..."

MIDDLEWARE_STATUS=$(curl -s http://localhost:8090/api/v1/middleware/status 2>/dev/null || echo "{}")

if echo "$MIDDLEWARE_STATUS" | grep -q "new_system_enabled"; then
    NEW_SYSTEM_ENABLED=$(echo "$MIDDLEWARE_STATUS" | jq -r '.new_system_enabled' 2>/dev/null || echo "false")

    if [ "$NEW_SYSTEM_ENABLED" = "true" ]; then
        print_status "success" "New middleware system is enabled"
    else
        print_status "warning" "Legacy middleware system is still active"
    fi

    REGISTERED_COUNT=$(echo "$MIDDLEWARE_STATUS" | jq -r '.registered_middleware | length' 2>/dev/null || echo "0")
    CHAIN_COUNT=$(echo "$MIDDLEWARE_STATUS" | jq -r '.available_chains | length' 2>/dev/null || echo "0")

    print_status "info" "Registered middleware: $REGISTERED_COUNT"
    print_status "info" "Available chains: $CHAIN_COUNT"
else
    print_status "warning" "Middleware status endpoint not available"
fi

# Check performance metrics
print_status "info" "Checking performance metrics..."

PERF_METRICS=$(curl -s http://localhost:8090/api/v1/middleware/performance 2>/dev/null || echo "{}")

if echo "$PERF_METRICS" | grep -q "chain_build_times"; then
    echo ""
    echo "📊 Chain Building Performance:"
    echo "$PERF_METRICS" | jq -r '.chain_build_times | to_entries[] | "  \(.key): \(.value)ms"' 2>/dev/null || echo "$PERF_METRICS"
else
    print_status "warning" "Performance metrics not available"
fi

# Check system metrics
print_status "info" "Checking system metrics..."

if [ ! -z "$APP_PID" ]; then
    MEMORY_USAGE=$(ps -o rss= -p $APP_PID | awk '{print $1/1024 " MB"}' 2>/dev/null || echo "Unknown")
    CPU_USAGE=$(ps -o %cpu= -p $APP_PID 2>/dev/null || echo "Unknown")
    UPTIME=$(ps -o etime= -p $APP_PID 2>/dev/null || echo "Unknown")

    echo ""
    echo "📈 System Metrics:"
    echo "  Memory Usage: $MEMORY_USAGE"
    echo "  CPU Usage: ${CPU_USAGE}%"
    echo "  Uptime: $UPTIME"
fi

# Check for legacy code usage
print_status "info" "Checking legacy code usage..."

LEGACY_MANAGER_USAGE=$(grep -r "Manager" internal/application/middleware/ --include="*.go" | grep -v "//" | wc -l 2>/dev/null || echo "0")
NEW_ARCHITECTURE_USAGE=$(grep -r "Orchestrator\|Registry\|Chain" internal/application/middleware/ --include="*.go" | grep -v "//" | wc -l 2>/dev/null || echo "0")

echo ""
echo "🏗️ Architecture Usage:"
echo "  Legacy Manager references: $LEGACY_MANAGER_USAGE"
echo "  New architecture references: $NEW_ARCHITECTURE_USAGE"

# Calculate migration progress
if [ "$LEGACY_MANAGER_USAGE" -gt 0 ] && [ "$NEW_ARCHITECTURE_USAGE" -gt 0 ]; then
    TOTAL_REFERENCES=$((LEGACY_MANAGER_USAGE + NEW_ARCHITECTURE_USAGE))
    MIGRATION_PERCENTAGE=$((NEW_ARCHITECTURE_USAGE * 100 / TOTAL_REFERENCES))

    echo "  Migration progress: ${MIGRATION_PERCENTAGE}%"

    if [ "$MIGRATION_PERCENTAGE" -ge 80 ]; then
        print_status "success" "Migration is nearly complete"
    elif [ "$MIGRATION_PERCENTAGE" -ge 50 ]; then
        print_status "warning" "Migration is in progress"
    else
        print_status "warning" "Migration is in early stages"
    fi
fi

# Check for deprecated code
print_status "info" "Checking for deprecated code..."

DEPRECATED_COUNT=$(grep -r "@deprecated" internal/application/middleware/ --include="*.go" | wc -l 2>/dev/null || echo "0")

if [ "$DEPRECATED_COUNT" -gt 0 ]; then
    print_status "warning" "Found $DEPRECATED_COUNT deprecated functions"
    echo "  These should be removed in the next cleanup phase"
else
    print_status "success" "No deprecated code found"
fi

# Check for unused imports
print_status "info" "Checking for unused imports..."

UNUSED_IMPORTS=$(goimports -l internal/application/middleware/ 2>/dev/null | wc -l || echo "0")

if [ "$UNUSED_IMPORTS" -gt 0 ]; then
    print_status "warning" "Found $UNUSED_IMPORTS files with unused imports"
    echo "  Run 'task cleanup:safe' to fix these"
else
    print_status "success" "No unused imports found"
fi

# Check recent logs for errors
print_status "info" "Checking recent logs..."

if [ -f "logs/app.log" ]; then
    RECENT_ERRORS=$(tail -n 100 logs/app.log | grep -i "error\|panic\|fatal" | wc -l 2>/dev/null || echo "0")

    if [ "$RECENT_ERRORS" -gt 0 ]; then
        print_status "warning" "Found $RECENT_ERRORS errors in recent logs"
        echo "  Check logs/app.log for details"
    else
        print_status "success" "No recent errors found in logs"
    fi
else
    print_status "warning" "Log file not found"
fi

# Cleanup readiness assessment
echo ""
echo "🧹 Cleanup Readiness Assessment:"
echo "================================"

# Check if new system is stable
if [ "$NEW_SYSTEM_ENABLED" = "true" ]; then
    print_status "success" "New system is enabled and running"
else
    print_status "warning" "New system not yet enabled - not ready for cleanup"
fi

# Check performance
if echo "$PERF_METRICS" | grep -q "chain_build_times"; then
    BUILD_TIMES=$(echo "$PERF_METRICS" | jq -r '.chain_build_times | to_entries[] | .value' 2>/dev/null)
    MAX_BUILD_TIME=0

    for time in $BUILD_TIMES; do
        if (( $(echo "$time > $MAX_BUILD_TIME" | bc -l) )); then
            MAX_BUILD_TIME=$time
        fi
    done

    if (( $(echo "$MAX_BUILD_TIME < 10" | bc -l) )); then
        print_status "success" "Performance is good (max build time: ${MAX_BUILD_TIME}ms)"
    else
        print_status "warning" "Performance needs monitoring (max build time: ${MAX_BUILD_TIME}ms)"
    fi
else
    print_status "warning" "Performance data not available"
fi

# Check error rates
if [ "$RECENT_ERRORS" -eq 0 ]; then
    print_status "success" "No recent errors - system is stable"
else
    print_status "warning" "Recent errors detected - not ready for cleanup"
fi

# Overall recommendation
echo ""
echo "🎯 Cleanup Recommendation:"
echo "=========================="

if [ "$NEW_SYSTEM_ENABLED" = "true" ] && [ "$RECENT_ERRORS" -eq 0 ] && [ "$MIGRATION_PERCENTAGE" -ge 80 ]; then
    print_status "success" "System appears ready for Phase 2 cleanup (deprecation)"
    echo "  Run: task cleanup:deprecate"
elif [ "$NEW_SYSTEM_ENABLED" = "true" ] && [ "$RECENT_ERRORS" -eq 0 ]; then
    print_status "warning" "System ready for Phase 1 cleanup (safe cleanup only)"
    echo "  Run: task cleanup:safe"
else
    print_status "error" "System not ready for cleanup"
    echo "  Focus on stabilizing the new system first"
fi

echo ""
print_status "info" "Monitoring completed"
