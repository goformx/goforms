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
