# GoFormX Deployment Guide

## Quick Setup with Deployer

This project uses [Deployer](https://deployer.org/) for simple, reliable deployments to DigitalOcean droplets.

### Prerequisites

1. **DigitalOcean Droplet** with Docker installed
2. **SSH Key Pair** for secure access
3. **GitHub Repository** with secrets configured

### Step 1: Create a New DigitalOcean Droplet

```bash
# Create a new droplet (recommended specs)
- Ubuntu 22.04 LTS
- 2GB RAM / 1 vCPU (minimum)
- 50GB SSD
- Enable backups
```

### Step 2: Install Docker on the Droplet

```bash
# SSH into your droplet
ssh root@your-droplet-ip

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add your user to docker group
sudo usermod -aG docker $USER
```

### Step 3: Generate SSH Key for Deployment

```bash
# Generate a new SSH key pair
ssh-keygen -t ed25519 -C "deploy@goformx" -f ~/.ssh/deploy_key

# Add public key to droplet
ssh-copy-id -i ~/.ssh/deploy_key.pub root@your-droplet-ip

# Copy private key content (you'll need this for GitHub secrets)
cat ~/.ssh/deploy_key
```

### Step 4: Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add these secrets:

| Secret Name         | Value                            |
| ------------------- | -------------------------------- |
| `SSH_PRIVATE_KEY`   | Content of your private key file |
| `DROPLET_IP`        | Your DigitalOcean droplet IP     |
| `POSTGRES_PASSWORD` | Strong database password         |
| `SESSION_SECRET`    | Random 32-character string       |
| `CSRF_SECRET`       | Random 32-character string       |

Generate secrets:

```bash
# Generate session and CSRF secrets
openssl rand -hex 32
```

### Step 5: Update Deployer Configuration

Edit `deploy.php` and update the hostnames:

```php
host('production')
    ->setHostname('goformx.com')  // Your domain
    ->set('deploy_path', '/opt/goformx');

host('staging')
    ->setHostname('staging.goformx.com')  // Your staging domain
    ->set('deploy_path', '/opt/goformx-staging');
```

### Step 6: Deploy!

#### Option A: Manual Deployment

```bash
# Install Deployer locally
curl -LO https://deployer.org/deployer.phar
chmod +x deployer.phar
sudo mv deployer.phar /usr/local/bin/dep

# Deploy to staging
dep deploy staging

# Deploy to production
dep deploy production
```

#### Option B: GitHub Actions (Recommended)

1. Push to `main` branch → Automatic deployment to production
2. Create a tag `v1.0.0` → Automatic deployment to production
3. Use GitHub Actions UI → Manual deployment to staging/production

### Step 7: Configure Nginx (Optional)

If you want to use a domain name, set up Nginx as a reverse proxy:

```bash
# Install Nginx
sudo apt update
sudo apt install nginx

# Create Nginx config
sudo nano /etc/nginx/sites-available/goformx
```

```nginx
server {
    listen 80;
    server_name goformx.com;

    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/goformx /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Step 8: SSL Certificate (Optional)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d goformx.com
```

## Deployment Commands

```bash
# Deploy to staging
dep deploy staging

# Deploy to production
dep deploy production

# Rollback last deployment
dep rollback production

# Check deployment status
dep status production
```

## Troubleshooting

### Check Docker Services

```bash
# SSH into droplet
ssh root@your-droplet-ip

# Check running containers
docker ps

# Check logs
docker-compose logs -f

# Restart services
docker-compose restart
```

### Check Application Health

```bash
# Test health endpoint
curl http://localhost:8090/health

# Check database connection
docker-compose exec postgres psql -U goforms -d goforms -c "SELECT 1;"
```

## Benefits of This Approach

✅ **Simple**: Uses established deployment patterns
✅ **Reliable**: Deployer is battle-tested
✅ **Rollback**: Easy to rollback deployments
✅ **Secure**: SSH key-based authentication
✅ **Flexible**: Works with any Docker setup
✅ **CI/CD Ready**: Integrates with GitHub Actions

## Alternative: CapRover (Even Simpler)

If you want something even simpler, consider [CapRover](https://caprover.com/):

1. Install CapRover on your droplet
2. Connect your GitHub repo
3. Deploy with one click

CapRover handles SSL, domains, and scaling automatically.
