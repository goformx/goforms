#!/bin/bash

# GoFormX Production Deployment Script
# This script generates environment variables and uses goforms-compose CLI for orchestration

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
LATEST_TAG="${IMAGE_TAG:-v0.1.5}"  # Use IMAGE_TAG env var if set

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

# Check if goforms-compose CLI is available
if ! command -v goforms-compose &> /dev/null; then
    # Try to find it in common locations
    if [ -f "./bin/goforms-compose" ]; then
        GOFORMS_COMPOSE="./bin/goforms-compose"
    elif [ -f "/usr/local/bin/goforms-compose" ]; then
        GOFORMS_COMPOSE="/usr/local/bin/goforms-compose"
    else
        error "goforms-compose CLI not found. Please build it first: go build -o bin/goforms-compose ./cmd/goforms-compose"
    fi
else
    GOFORMS_COMPOSE="goforms-compose"
fi

log "Starting GoFormX production deployment..."

# Step 1: Clean up old deployment
log "Step 1: Cleaning up old deployment..."

if [ -d "$APP_DIR" ]; then
    log "Removing old application directory..."
    sudo rm -rf "$APP_DIR"
fi

# Stop and remove old containers using CLI
log "Stopping old Docker containers..."
if [ -f "$APP_DIR/docker-compose.prod.yml" ]; then
    cd "$APP_DIR" || true
    $GOFORMS_COMPOSE prod down --project-dir "$APP_DIR" --compose-file docker-compose.prod.yml || true
fi
docker system prune -f

# Step 2: Create fresh application directory
log "Step 2: Creating fresh application directory..."
sudo mkdir -p "$APP_DIR"
sudo chown $USER:$USER "$APP_DIR"

# Step 3: Copy compose file to deployment directory
log "Step 3: Setting up Compose configuration..."
# Copy the production compose file from repo to deployment directory
# In a real scenario, this would come from the repo or be templated
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$REPO_ROOT/docker-compose.prod.yml" ]; then
    cp "$REPO_ROOT/docker-compose.prod.yml" "$APP_DIR/docker-compose.prod.yml"
    log "Copied docker-compose.prod.yml to $APP_DIR"
else
    error "docker-compose.prod.yml not found in repository root"
fi

# Step 4: Create environment file
log "Step 4: Creating environment file..."
cat > "$APP_DIR/.env" << EOF
# GoFormX Production Environment
APP_NAME=${APP_NAME}
APP_ENV=production
APP_DEBUG=false
APP_LOG_LEVEL=info

# Database Configuration
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_NAME=goforms
DB_USERNAME=goforms
DB_PASSWORD=${POSTGRES_PASSWORD}
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# PostgreSQL Configuration (for postgres service)
POSTGRES_DB=goforms
POSTGRES_USER=goforms
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

# Security
SESSION_SECRET=${SESSION_SECRET}
SECURITY_CSRF_SECRET=${CSRF_SECRET}
SECURE_COOKIES=true

# Docker Image
DOCKER_REGISTRY=${DOCKER_REGISTRY:-ghcr.io}
GITHUB_REPOSITORY=${GITHUB_REPOSITORY:-goformx/goforms}
IMAGE_TAG=${LATEST_TAG}

# CORS (adjust as needed)
CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS:-https://goforms.example.com}
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With,X-API-Key
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=3600
EOF

log "Environment file created at $APP_DIR/.env"

# Step 5: Deploy using goforms-compose CLI
log "Step 5: Deploying with goforms-compose..."
cd "$APP_DIR"

# Deploy with pull and wait for health
log "Deploying services..."
$GOFORMS_COMPOSE prod deploy \
    --project-name "$APP_NAME" \
    --compose-file docker-compose.prod.yml \
    --env-file .env \
    --project-dir "$APP_DIR" \
    --tag "$LATEST_TAG" \
    --pull

if [ $? -ne 0 ]; then
    error "Deployment failed"
fi

# Step 6: Configure Supervisor (if not already configured)
log "Step 6: Configuring Supervisor..."

if [ ! -f "$SUPERVISOR_CONF" ]; then
    sudo tee "$SUPERVISOR_CONF" > /dev/null << EOF
[program:goforms]
command=$GOFORMS_COMPOSE prod deploy --project-dir $APP_DIR --compose-file docker-compose.prod.yml --env-file .env --tag ${LATEST_TAG}
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

# Step 7: Create deployment info file
log "Step 7: Creating deployment info..."
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
- View logs: $GOFORMS_COMPOSE prod logs --project-dir $APP_DIR
- Status: $GOFORMS_COMPOSE prod status --project-dir $APP_DIR
- Restart: $GOFORMS_COMPOSE prod deploy --tag $LATEST_TAG --project-dir $APP_DIR
- Stop: $GOFORMS_COMPOSE prod down --project-dir $APP_DIR
- Rollback: $GOFORMS_COMPOSE prod rollback --project-dir $APP_DIR
EOF

# Step 8: Final status check
log "Step 8: Final status check..."

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
log "View logs: cd $APP_DIR && $GOFORMS_COMPOSE prod logs"
log "Status: cd $APP_DIR && $GOFORMS_COMPOSE prod status"
log "Restart: cd $APP_DIR && $GOFORMS_COMPOSE prod deploy --tag $LATEST_TAG"
log "Stop: cd $APP_DIR && $GOFORMS_COMPOSE prod down"
log "Rollback: cd $APP_DIR && $GOFORMS_COMPOSE prod rollback"
log "Supervisor status: sudo supervisorctl status goforms"

echo
log "=== Next Steps ==="
log "1. Test the application at http://localhost:8090"
log "2. Configure Nginx reverse proxy (optional)"
log "3. Set up SSL certificate (optional)"
log "4. Configure firewall rules"

log "🎉 Deployment completed successfully!"
