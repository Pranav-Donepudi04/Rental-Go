# Deployment Guide

This guide covers deploying the Rental Management System to production using Docker.

> **Looking for FREE deployment without credit card?**  
> See **[FREE_DEPLOYMENT.md](./FREE_DEPLOYMENT.md)** for Render, Fly.io, and other free options!

## 🐳 Docker Deployment

### Prerequisites

- Docker installed ([Install Docker](https://docs.docker.com/get-docker/))
- Docker Compose installed (usually comes with Docker Desktop)
- Neon DB connection string (`DATABASE_URL`)

### Quick Start (Local Testing)

1. **Create `.env` file** in the project root:
```env
SERVER_PORT=8080
LOG_LEVEL=info
DATABASE_URL=postgres://user:password@host/database?sslmode=require
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=25
DB_CONNECTION_TIMEOUT=30
```

2. **Build and run with Docker:**
```bash
# Build the image
docker build -t rental-app .

# Run the container
docker run -p 8080:8080 --env-file .env rental-app
```

**Note:** For local testing only. For cloud deployment (Render, etc.), you don't need docker-compose.

3. **Access the application:**
- Open http://localhost:8080 in your browser

### Production Deployment

#### Option 1: Docker on VPS (DigitalOcean, AWS EC2, etc.)

1. **SSH into your server:**
```bash
ssh user@your-server-ip
```

2. **Clone your repository:**
```bash
git clone <your-repo-url>
cd pythonProject2
```

3. **Create `.env` file** with production values:
```env
SERVER_PORT=8080
LOG_LEVEL=info
DATABASE_URL=postgres://neon-db-connection-string
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=25
DB_CONNECTION_TIMEOUT=30
```

4. **Build and run:**
```bash
docker-compose up -d --build
```

5. **View logs:**
```bash
docker-compose logs -f
```

6. **Set up reverse proxy (Nginx):**
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Option 2: Cloud Platforms

##### Railway

1. **Connect your GitHub repository**
2. **Add environment variables** in Railway dashboard:
   - `DATABASE_URL` (from Neon)
   - `SERVER_PORT=8080`
   - `LOG_LEVEL=info`
3. **Deploy** - Railway auto-detects Dockerfile

##### Render

1. **Create new Web Service**
2. **Connect GitHub repository**
3. **Configure:**
   - **Build Command:** `docker build -t rental-app .`
   - **Start Command:** `docker run -p $PORT:8080 --env-file .env rental-app`
   - **Environment:** Add all `.env` variables

##### Fly.io

1. **Install Fly CLI:**
```bash
curl -L https://fly.io/install.sh | sh
```

2. **Login:**
```bash
fly auth login
```

3. **Create `fly.toml`:**
```toml
app = "rental-management-app"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    port = 80
    handlers = ["http"]
    force_https = true

  [[services.env]]
    SERVER_PORT = "8080"
    DATABASE_URL = "from-secrets"
```

4. **Deploy:**
```bash
fly secrets set DATABASE_URL="your-neon-url"
fly deploy
```

##### AWS App Runner / ECS

See cloud-specific documentation for containerized Go apps.

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_PORT` | HTTP server port | `8080` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `DATABASE_URL` | Full PostgreSQL connection string | - | **Yes** |
| `DB_HOST` | Database host (if not using DATABASE_URL) | `localhost` | No |
| `DB_PORT` | Database port | `5432` | No |
| `DB_USER` | Database user | `postgres` | No |
| `DB_PASSWORD` | Database password | - | No |
| `DB_NAME` | Database name | `formdb` | No |
| `DB_SSL_MODE` | SSL mode for Postgres | `require` | No |
| `DB_MAX_CONNECTIONS` | Max DB connection pool size | `25` | No |
| `DB_CONNECTION_TIMEOUT` | Connection timeout (seconds) | `30` | No |

### Docker Commands (Local Development)

```bash
# Build image
docker build -t rental-management-app .

# Run container
docker run -p 8080:8080 --env-file .env rental-management-app

# Run in background
docker run -d -p 8080:8080 --env-file .env --name rental-app rental-management-app

# View logs
docker logs -f rental-app

# Stop container
docker stop rental-app

# Remove container
docker rm rental-app
```

**Note:** For cloud platforms like Render, Fly.io, etc., you only need the `Dockerfile`. No docker-compose required!

### Health Checks

Add to your deployment:

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/login"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 🔍 Monitoring & Logs

### View Application Logs

```bash
# Docker Compose
docker-compose logs -f app

# Docker
docker logs -f rental-app
```

### Database Connection Monitoring

The application logs database connection details on startup. Monitor for:
- Connection pool exhaustion
- Timeout errors
- SSL connection issues

## 🔒 Security Best Practices

1. **Never commit `.env` files** - Use secrets management in your deployment platform
2. **Use HTTPS** - Set up SSL/TLS via reverse proxy (Nginx/Caddy) or platform (Railway/Render provide it)
3. **Restrict database access** - Neon DB allows IP whitelisting
4. **Rotate credentials** - Regularly update database passwords
5. **Enable firewall** - Only expose necessary ports (80/443)

## 📊 Neon DB Optimization

### Connection Pooling

Your app is configured for connection pooling. The current settings:
- `DB_MAX_CONNECTIONS=25` - Maximum concurrent connections
- `DB_CONNECTION_TIMEOUT=30` - Timeout in seconds

**Note:** Neon DB free tier has connection limits. Adjust `DB_MAX_CONNECTIONS` based on your Neon plan:
- **Free tier:** ~10-15 connections max
- **Pro tier:** 50+ connections

### Neon DB Advantages You're Using

✅ **Serverless scaling** - Automatically scales with traffic
✅ **Auto-pausing** - Free tier pauses inactive databases (saves costs)
✅ **Branching** - Create database branches for testing (not currently used, but available)
✅ **Point-in-time recovery** - Automatic backups (Pro tier)

### Optimization Recommendations

1. **Reduce connection pool** if on free tier:
   ```env
   DB_MAX_CONNECTIONS=10
   ```

2. **Use connection string** - Prefer `DATABASE_URL` over individual DB_* vars (already doing this ✅)

3. **Enable Neon connection pooling** (if using Pro tier):
   - Use Neon's connection pooler endpoint
   - Update `DATABASE_URL` to use pooler URL

4. **Monitor connection usage** in Neon dashboard

## 🚀 Production Checklist

- [ ] Environment variables configured
- [ ] Database migrations run
- [ ] SSL/TLS certificate configured
- [ ] Domain name pointed to server
- [ ] Health checks configured
- [ ] Logging/monitoring set up
- [ ] Backup strategy in place
- [ ] Connection pool size adjusted for Neon tier
- [ ] Security headers configured (via reverse proxy)

## 🔧 Troubleshooting

### Container won't start
```bash
docker logs rental-app
```

### Database connection issues
- Verify `DATABASE_URL` is correct
- Check Neon DB dashboard for connection status
- Ensure IP is whitelisted in Neon (if required)

### Port already in use
```bash
# Change SERVER_PORT in .env or docker-compose.yml
SERVER_PORT=8081
```

### Build failures
```bash
# Clean build
docker-compose down
docker-compose build --no-cache
```

