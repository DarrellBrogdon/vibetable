# VibeTable Installation Guide

This guide walks you through deploying VibeTable on a VPS such as AWS EC2, DigitalOcean Droplet, or any Linux server.

---

## Prerequisites

- A VPS running Ubuntu 22.04+ (or similar Linux distribution)
- A domain name pointed to your server's IP address
- SSH access to your server
- At least 1GB RAM and 10GB storage (2GB+ RAM recommended)

---

## 1. Server Initial Setup

### Connect to your server

```bash
ssh root@your-server-ip
```

### Update system packages

```bash
apt update && apt upgrade -y
```

### Create a non-root user (recommended)

```bash
adduser vibetable
usermod -aG sudo vibetable
su - vibetable
```

---

## 2. Install Docker and Docker Compose

### Install Docker

```bash
# Install prerequisites
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# Add your user to docker group (avoids needing sudo for docker commands)
sudo usermod -aG docker $USER

# Apply group changes (or log out and back in)
newgrp docker
```

### Install Docker Compose

```bash
# Download Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Make it executable
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
```

---

## 3. Install Nginx (Reverse Proxy)

```bash
sudo apt install -y nginx
sudo systemctl enable nginx
sudo systemctl start nginx
```

---

## 4. Clone the Repository

```bash
cd ~
git clone https://github.com/your-username/vibetable.git
cd vibetable
```

---

## 5. Configure Environment Variables

### Create production environment file

```bash
cp .env.example .env
nano .env
```

### Update the following values for production:

```bash
# Database Configuration
POSTGRES_USER=vibetable
POSTGRES_PASSWORD=<generate-a-strong-password>
POSTGRES_DB=vibetable

# Backend Configuration
DATABASE_URL=postgres://vibetable:<your-password>@db:5432/vibetable?sslmode=disable
PORT=8080
JWT_SECRET=<generate-a-32+-character-secret>
SESSION_SECRET=<generate-another-32+-character-secret>

# Frontend Configuration - Use your domain
PUBLIC_API_URL=https://api.yourdomain.com

# Email Configuration (get from Resend.com)
RESEND_API_KEY=re_xxxxxxxxxxxx
EMAIL_FROM=noreply@yourdomain.com

# Security Configuration - Use your domain
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
CSRF_SECRET=<generate-another-32+-character-secret>

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE=10
```

### Generate secure secrets

```bash
# Generate random secrets (run these and copy the output)
openssl rand -base64 32  # For POSTGRES_PASSWORD
openssl rand -base64 32  # For JWT_SECRET
openssl rand -base64 32  # For SESSION_SECRET
openssl rand -base64 32  # For CSRF_SECRET
```

---

## 6. Create Production Docker Compose Configuration

Create a production-optimized `docker-compose.prod.yml`:

```bash
nano docker-compose.prod.yml
```

```yaml
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    container_name: vibetable-db
    restart: always
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - vibetable-network

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.prod
    container_name: vibetable-backend
    restart: always
    environment:
      DATABASE_URL: ${DATABASE_URL}
      PORT: ${PORT}
      JWT_SECRET: ${JWT_SECRET}
      SESSION_SECRET: ${SESSION_SECRET}
      RESEND_API_KEY: ${RESEND_API_KEY}
      EMAIL_FROM: ${EMAIL_FROM}
      ALLOWED_ORIGINS: ${ALLOWED_ORIGINS}
      CSRF_SECRET: ${CSRF_SECRET}
      RATE_LIMIT_REQUESTS_PER_MINUTE: ${RATE_LIMIT_REQUESTS_PER_MINUTE}
      RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE: ${RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE}
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      db:
        condition: service_healthy
    networks:
      - vibetable-network

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
      args:
        PUBLIC_API_URL: ${PUBLIC_API_URL}
    container_name: vibetable-frontend
    restart: always
    ports:
      - "127.0.0.1:3000:3000"
    depends_on:
      - backend
    networks:
      - vibetable-network

volumes:
  postgres_data:

networks:
  vibetable-network:
    driver: bridge
```

---

## 7. Create Production Dockerfiles

### Backend Production Dockerfile

```bash
nano backend/Dockerfile.prod
```

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
```

### Frontend Production Dockerfile

```bash
nano frontend/Dockerfile.prod
```

```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build argument for API URL
ARG PUBLIC_API_URL
ENV PUBLIC_API_URL=$PUBLIC_API_URL

# Build the application
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Copy package files and install production dependencies only
COPY package*.json ./
RUN npm ci --only=production

# Copy built application
COPY --from=builder /app/build ./build

# Expose port
EXPOSE 3000

# Run the application
CMD ["node", "build"]
```

---

## 8. Configure Nginx Reverse Proxy

### Create Nginx configuration

```bash
sudo nano /etc/nginx/sites-available/vibetable
```

```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# Frontend
server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL certificates (will be added by Certbot)
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}

# API Backend
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL certificates (will be added by Certbot)
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Increase max body size for file uploads
    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Enable the site

```bash
sudo ln -s /etc/nginx/sites-available/vibetable /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default  # Remove default site
sudo nginx -t  # Test configuration
```

---

## 9. Install SSL Certificates (Let's Encrypt)

### Install Certbot

```bash
sudo apt install -y certbot python3-certbot-nginx
```

### Obtain SSL certificates

First, temporarily comment out the SSL lines in your Nginx config, then:

```bash
# Get certificates for all domains
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com -d api.yourdomain.com
```

Follow the prompts to complete the certificate installation. Certbot will automatically update your Nginx configuration.

### Auto-renewal

Certbot automatically sets up a cron job for renewal. Test it with:

```bash
sudo certbot renew --dry-run
```

---

## 10. Configure Firewall

```bash
# Allow SSH, HTTP, and HTTPS
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
sudo ufw status
```

---

## 11. Start the Application

### Build and start containers

```bash
cd ~/vibetable
docker-compose -f docker-compose.prod.yml up -d --build
```

### Restart Nginx

```bash
sudo systemctl restart nginx
```

### Verify everything is running

```bash
# Check Docker containers
docker-compose -f docker-compose.prod.yml ps

# Check container logs
docker-compose -f docker-compose.prod.yml logs -f

# Test the API
curl https://api.yourdomain.com/health
```

---

## 12. DNS Configuration

Create the following DNS records pointing to your server's IP address:

| Type | Name | Value |
|------|------|-------|
| A | @ | your-server-ip |
| A | www | your-server-ip |
| A | api | your-server-ip |

---

## Maintenance Commands

### View logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f backend
docker-compose -f docker-compose.prod.yml logs -f frontend
docker-compose -f docker-compose.prod.yml logs -f db
```

### Restart services

```bash
# Restart all
docker-compose -f docker-compose.prod.yml restart

# Restart specific service
docker-compose -f docker-compose.prod.yml restart backend
```

### Update application

```bash
cd ~/vibetable
git pull origin main
docker-compose -f docker-compose.prod.yml up -d --build
```

### Database backup

```bash
# Create backup
docker-compose -f docker-compose.prod.yml exec db pg_dump -U vibetable vibetable > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
cat backup.sql | docker-compose -f docker-compose.prod.yml exec -T db psql -U vibetable vibetable
```

### Stop application

```bash
docker-compose -f docker-compose.prod.yml down
```

### Reset database (WARNING: destroys all data)

```bash
docker-compose -f docker-compose.prod.yml down -v
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## Troubleshooting

### Container won't start

```bash
# Check logs for errors
docker-compose -f docker-compose.prod.yml logs backend

# Rebuild from scratch
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d
```

### Database connection issues

```bash
# Check if database is healthy
docker-compose -f docker-compose.prod.yml ps

# Connect to database directly
docker-compose -f docker-compose.prod.yml exec db psql -U vibetable vibetable
```

### 502 Bad Gateway

- Check if containers are running: `docker-compose -f docker-compose.prod.yml ps`
- Check backend logs: `docker-compose -f docker-compose.prod.yml logs backend`
- Verify Nginx config: `sudo nginx -t`

### SSL certificate issues

```bash
# Check certificate status
sudo certbot certificates

# Renew certificates
sudo certbot renew
```

---

## Email Configuration (Resend)

1. Sign up at [resend.com](https://resend.com)
2. Add your domain and verify DNS records
3. Create an API key
4. Update `.env` with your API key and verified email address

---

## Security Recommendations

1. **Keep system updated**: Run `sudo apt update && sudo apt upgrade` regularly
2. **Use strong passwords**: All secrets should be randomly generated
3. **Enable automatic security updates**:
   ```bash
   sudo apt install unattended-upgrades
   sudo dpkg-reconfigure -plow unattended-upgrades
   ```
4. **Monitor logs**: Check application and system logs regularly
5. **Backup regularly**: Set up automated database backups
6. **Consider fail2ban**: Protect against brute force attacks
   ```bash
   sudo apt install fail2ban
   sudo systemctl enable fail2ban
   ```

---

## Resource Requirements

| Users | RAM | CPU | Storage |
|-------|-----|-----|---------|
| 1-10 | 1GB | 1 vCPU | 10GB |
| 10-100 | 2GB | 2 vCPU | 20GB |
| 100-500 | 4GB | 2 vCPU | 50GB |

---

## Support

For issues and feature requests, please open an issue on the GitHub repository.
