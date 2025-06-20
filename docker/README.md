
---

## Quick Start

### Deploy GoFormX Application & Database (Docker Compose)

**Production:**

```bash
# Set up environment variables (NO .env files in production!)
export POSTGRES_DB="goforms"
export POSTGRES_USER="goforms"
export POSTGRES_PASSWORD="your-secure-password"
export SESSION_SECRET="$(openssl rand -hex 32)"
export CSRF_SECRET="$(openssl rand -hex 32)"
export CORS_ORIGINS="https://goforms.example.com"
export DOCKER_REGISTRY="ghcr.io"
export GITHUB_REPOSITORY="goformx/goforms"

# Deploy
cd docker/production
docker compose up -d
```

**Development:**

```bash
cd docker/development
docker compose up -d
```

---

## Nginx Reverse Proxy (Host)

### 1. Install nginx

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install nginx

# CentOS/RHEL
sudo yum install nginx
# or
sudo dnf install nginx
```

### 2. Configure nginx

```bash
# Copy the example configuration
sudo cp docker/host/nginx/host-nginx.conf.example /etc/nginx/sites-available/goforms

# Edit the configuration as needed
sudo nano /etc/nginx/sites-available/goforms

# Enable the site
sudo ln -s /etc/nginx/sites-available/goforms /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### 3. SSL Certificate

#### Let's Encrypt (Recommended)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d goforms.example.com -d www.goforms.example.com
# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Self-signed (Development)

```bash
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/goforms.key \
  -out /etc/ssl/certs/goforms.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=goforms.example.com"
```

---

## Nginx Configuration Details

- **SSL/TLS:** Modern protocols and ciphers.
- **Security Headers:** X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, Referrer-Policy, Content-Security-Policy.
- **Rate Limiting:** Example config included.
- **Caching:** Static assets are cached for performance.
- **Health Checks:** `/health` endpoint is proxied and protected.
- **Customizations:** Update `server_name`, SSL paths, and upstream as needed.

---

## Environment Configuration

**Production:**

- Do NOT use `.env` files in production. Use environment variables directly.
- Required variables:
  - `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
  - `SESSION_SECRET`, `CSRF_SECRET`
  - `CORS_ORIGINS`
  - `DOCKER_REGISTRY`, `GITHUB_REPOSITORY`, `IMAGE_TAG` (for pulling images)

**Development:**

- You may use `.env` files in `docker/development/`.

---

## Management Commands

```bash
# View logs
docker compose logs -f

# Stop services
docker compose down

# Restart services
docker compose restart

# Update to latest image
docker compose pull
docker compose up -d
```

---

## Security Considerations

- **Never use .env files in production.**
- **Use strong secrets** (generate with `openssl rand -hex 32`).
- **Rotate secrets regularly.**
- **Bind app to localhost** and use nginx as the only public entrypoint.
- **Enable HTTPS** and set security headers in nginx.
- **Restrict direct access to app/database ports.**
- **Monitor logs and health endpoints.**

---

## Troubleshooting

- **502 Bad Gateway:** Check if GoFormX is running (`docker ps`), check logs, test direct connection (`curl http://127.0.0.1:8090/health`).
- **SSL Issues:** Check certificate validity, permissions, and nginx config.
- **Missing Environment Variables:** Ensure all required variables are exported before running Docker Compose.
- **Database Issues:** Test connection with `psql`, check Docker Compose logs.

---

## Backup and Recovery

- **Database:** Use `pg_dump` and `psql` via Docker Compose for backup/restore.
- **Nginx Config:** Backup `/etc/nginx/` and SSL certs.
- **App Logs:** Backup Docker volumes if needed.

---

## Monitoring

- **Nginx logs:** `/var/log/nginx/goforms_access.log`, `/var/log/nginx/goforms_error.log`
- **App health:** `curl http://localhost:8090/health`
- **Metrics:** `/metrics` endpoint if enabled

---

## Support

- Check logs: `docker compose logs`
- Verify configuration: `docker compose config`
- Test connectivity: Use health check endpoints
- Review this documentation and the main project README

---

## References

- See `docker/host/nginx/host-nginx.conf.example` for a full nginx config template.
- See `docker/production/docker-compose.yml` for the production stack.
- See `docker/development/docker-compose.yml` for the development stack.

---

**Deploy GoFormX securely and efficiently with Docker and host-based nginx!**
