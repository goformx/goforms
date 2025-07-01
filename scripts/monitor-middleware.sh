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
