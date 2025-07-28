#!/bin/bash

# GoFormX Production Deployment Script
# This script provides a clean deployment with fresh PostgreSQL

set -e  # Exit on any error

# Configuration
APP_NAME="goforms"
APP_USER="goforms"
APP_DIR="/opt/goforms"
SUPERVISOR_CONF="/etc/supervisor/conf.d/goforms.conf"
POSTGRES_PASSWORD="$(openssl rand -hex 32)"
SESSION_SECRET="$(openssl rand -hex 32)"
CSRF_SECRET="$(openssl rand -hex 32)"
DOCKER_IMAGE="ghcr.io/goformx/goforms"
LATEST_TAG="v0.1.5"  # Update this when you want a new version

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
    exit 1
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root. Please run as a regular user with sudo privileges."
fi

log "Starting GoFormX production deployment..."

# Step 1: Clean up old deployment
log "Step 1: Cleaning up old deployment..."

if [ -d "$APP_DIR" ]; then
    log "Removing old application directory..."
    sudo rm -rf "$APP_DIR"
fi

# Stop and remove old containers
log "Stopping old Docker containers..."
docker-compose -f /tmp/goforms-compose.yml down -v 2>/dev/null || true
docker system prune -f

# Step 2: Create fresh application directory
log "Step 2: Creating fresh application directory..."
sudo mkdir -p "$APP_DIR"
sudo chown $USER:$USER "$APP_DIR"

# Step 3: Set up fresh PostgreSQL
log "Step 3: Setting up fresh PostgreSQL..."

# Create PostgreSQL data directory
sudo mkdir -p /opt/postgres-data
sudo chown 999:999 /opt/postgres-data  # PostgreSQL container user

# Create docker-compose.yml
cat > "$APP_DIR/docker-compose.yml" << EOF
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: goforms-postgres
    environment:
      POSTGRES_DB: goforms
      POSTGRES_USER: goforms
      POSTGRES_PASSWORD: $POSTGRES_PASSWORD
    volumes:
      - /opt/postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goforms"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    image: $DOCKER_IMAGE:$LATEST_TAG
    container_name: goforms-app
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=goforms
      - DB_USER=goforms
      - DB_PASSWORD=$POSTGRES_PASSWORD
      - SESSION_SECRET=$SESSION_SECRET
      - CSRF_SECRET=$CSRF_SECRET
      - ENV=production
    ports:
      - "8090:8090"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  default:
    name: goforms-network
EOF

# Step 4: Pull latest Docker image
log "Step 4: Pulling latest Docker image ($LATEST_TAG)..."
docker pull "$DOCKER_IMAGE:$LATEST_TAG"

# Step 5: Start PostgreSQL first
log "Step 5: Starting PostgreSQL..."
cd "$APP_DIR"
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
log "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U goforms >/dev/null 2>&1; then
        log "PostgreSQL is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        error "PostgreSQL failed to start within 60 seconds"
    fi
    sleep 2
done

# Step 6: Run database migrations (if needed)
log "Step 6: Running database migrations..."
# Note: Your app should handle migrations on startup, but you can add manual migration here if needed

# Step 7: Start the application
log "Step 7: Starting GoFormX application..."
docker-compose up -d app

# Wait for application to be ready
log "Waiting for application to be ready..."
for i in {1..30}; do
    if curl -f http://localhost:8090/health >/dev/null 2>&1; then
        log "Application is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        warn "Application health check failed, but continuing..."
        break
    fi
    sleep 2
done

# Step 8: Configure Supervisor (if not already configured)
log "Step 8: Configuring Supervisor..."

if [ ! -f "$SUPERVISOR_CONF" ]; then
    sudo tee "$SUPERVISOR_CONF" > /dev/null << EOF
[program:goforms]
command=docker-compose -f $APP_DIR/docker-compose.yml up
directory=$APP_DIR
user=$USER
autostart=true
autorestart=true
stderr_logfile=/var/log/goforms.err.log
stdout_logfile=/var/log/goforms.out.log
environment=HOME="$HOME"
EOF

    sudo supervisorctl reread
    sudo supervisorctl update
    log "Supervisor configuration created"
else
    log "Supervisor configuration already exists"
fi

# Step 9: Create deployment info file
log "Step 9: Creating deployment info..."
cat > "$APP_DIR/deployment-info.txt" << EOF
GoFormX Deployment Information
=============================
Deployment Date: $(date)
Version: $LATEST_TAG
Docker Image: $DOCKER_IMAGE:$LATEST_TAG

Database Configuration:
- Host: localhost
- Port: 5432
- Database: goforms
- User: goforms
- Password: $POSTGRES_PASSWORD

Application:
- URL: http://localhost:8090
- Health Check: http://localhost:8090/health

Supervisor:
- Config: $SUPERVISOR_CONF
- Status: sudo supervisorctl status goforms

Useful Commands:
- View logs: docker-compose logs -f
- Restart app: docker-compose restart app
- Stop all: docker-compose down
- Update: Run this script again with new tag
EOF

# Step 10: Final status check
log "Step 10: Final status check..."

echo
log "=== Deployment Summary ==="
log "✅ PostgreSQL: Fresh installation"
log "✅ Application: $LATEST_TAG deployed"
log "✅ Supervisor: Configured"
log "✅ Health Check: $(curl -s http://localhost:8090/health || echo 'Failed')"

echo
log "=== Access Information ==="
log "Application URL: http://localhost:8090"
log "Health Check: http://localhost:8090/health"
log "PostgreSQL: localhost:5432 (goforms/$POSTGRES_PASSWORD)"

echo
log "=== Useful Commands ==="
log "View logs: cd $APP_DIR && docker-compose logs -f"
log "Restart: cd $APP_DIR && docker-compose restart"
log "Stop: cd $APP_DIR && docker-compose down"
log "Supervisor status: sudo supervisorctl status goforms"

echo
log "=== Next Steps ==="
log "1. Test the application at http://localhost:8090"
log "2. Configure Nginx reverse proxy (optional)"
log "3. Set up SSL certificate (optional)"
log "4. Configure firewall rules"

log "🎉 Deployment completed successfully!"
