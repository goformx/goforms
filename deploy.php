<?php

namespace Deployer;

require 'recipe/common.php';

// Configuration
set('application', 'goformx');
set('repository', 'https://github.com/goformx/goforms.git');
set('git_tty', true);
set('keep_releases', 5);
set('shared_files', []);
set('shared_dirs', []);
set('writable_dirs', []);

// Hosts
host('production')
    ->setHostname('goformx.com')
    ->set('deploy_path', '/opt/goformx')
    ->set('branch', 'main')
    ->set('docker_image', 'ghcr.io/goformx/goforms:{{branch}}');

host('staging')
    ->setHostname('staging.goformx.com')
    ->set('deploy_path', '/opt/goformx-staging')
    ->set('branch', 'main')
    ->set('docker_image', 'ghcr.io/goformx/goforms:staging-{{branch}}');

// Tasks
task('deploy:docker', function () {
    $image = get('docker_image');
    $deployPath = get('deploy_path');

    // Create docker-compose.yml
    $composeContent = <<<YAML
version: '3.8'

services:
  goforms:
    image: {$image}
    restart: unless-stopped
    ports:
      - "127.0.0.1:8090:8090"
    environment:
      - GOFORMS_APP_NAME=GoFormX
      - GOFORMS_APP_ENV={{host}}
      - GOFORMS_APP_DEBUG=false
      - GOFORMS_APP_LOGLEVEL=info
      - GOFORMS_APP_SCHEME=https
      - GOFORMS_APP_PORT=8090
      - GOFORMS_APP_HOST=0.0.0.0
      - GOFORMS_DB_CONNECTION=postgres
      - GOFORMS_DB_HOST=postgres
      - GOFORMS_DB_PORT=5432
      - GOFORMS_DB_NAME=goforms
      - GOFORMS_DB_USER=goforms
      - GOFORMS_DB_PASSWORD={{env.POSTGRES_PASSWORD}}
      - GOFORMS_SESSION_SECRET={{env.SESSION_SECRET}}
      - GOFORMS_SECURITY_CSRF_SECRET={{env.CSRF_SECRET}}
      - GOFORMS_SECURE_COOKIES=true
      - GOFORMS_CORS_ALLOWED_ORIGINS=https://{{host}}
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - goforms-network
    volumes:
      - goforms-logs:/app/logs
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_DB=goforms
      - POSTGRES_USER=goforms
      - POSTGRES_PASSWORD={{env.POSTGRES_PASSWORD}}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - goforms-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goforms -d goforms"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

networks:
  goforms-network:
    driver: bridge

volumes:
  postgres-data:
    driver: local
  goforms-logs:
    driver: local
YAML;

    // Write docker-compose.yml
    run("mkdir -p {$deployPath}");
    uploadContent($composeContent, "{$deployPath}/docker-compose.yml");

    // Pull and restart
    run("cd {$deployPath} && docker-compose pull");
    run("cd {$deployPath} && docker-compose down");
    run("cd {$deployPath} && docker-compose up -d");

    // Wait for health check
    run("sleep 30");

    // Check if services are running
    $status = run("cd {$deployPath} && docker-compose ps --format json");
    if (strpos($status, '"State":"Up"') === false) {
        throw new Exception('Services failed to start');
    }
});

task('deploy:health_check', function () {
    $host = get('hostname');
    $maxAttempts = 30;
    $delay = 10;

    for ($i = 1; $i <= $maxAttempts; $i++) {
        try {
            $response = run("curl -f -s https://{$host}/health");
            if ($response !== false) {
                writeln("✅ Health check passed");
                return;
            }
        } catch (Exception $e) {
            writeln("⚠️ Health check attempt {$i}/{$maxAttempts} failed, retrying in {$delay}s...");
            sleep($delay);
        }
    }

    throw new Exception("Health check failed after {$maxAttempts} attempts");
});

// Main deployment task
task('deploy', [
    'deploy:docker',
    'deploy:health_check',
]);

// Rollback task
task('rollback', function () {
    $deployPath = get('deploy_path');
    run("cd {$deployPath} && docker-compose down");
    run("cd {$deployPath} && docker-compose up -d");
});

after('deploy:failed', 'deploy:unlock');
