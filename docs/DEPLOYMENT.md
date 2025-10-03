# AgentOS Deployment Guide

This guide provides comprehensive instructions for deploying AgentOS in various environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Production Deployment](#production-deployment)
- [Environment Configuration](#environment-configuration)
- [Database Setup](#database-setup)
- [Monitoring & Logging](#monitoring--logging)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Docker and Docker Compose (recommended)
- OR Go 1.21+ (for native deployment)
- OpenAI API key (or other LLM provider)

### Using Docker Compose (Recommended)

1. **Clone the repository**
```bash
git clone https://github.com/rexleimo/agno-go.git
cd agno-Go
```

2. **Configure environment**
```bash
cp .env.example .env
# Edit .env and add your API keys
nano .env  # or vim, code, etc.
```

3. **Start all services**
```bash
# Start core services (AgentOS + PostgreSQL + Redis)
docker-compose up -d

# Or with optional services (Ollama + ChromaDB)
docker-compose --profile with-ollama --profile with-vectordb up -d
```

4. **Verify deployment**
```bash
# Check health
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","service":"agentos","time":1704067200}
```

5. **View logs**
```bash
docker-compose logs -f agentos
```

### Native Deployment (Without Docker)

1. **Build the application**
```bash
cd agno-Go
go build -o agentos cmd/server/main.go
```

2. **Set environment variables**
```bash
export OPENAI_API_KEY=sk-your-key-here
export AGENTOS_ADDRESS=:8080
export AGENTOS_DEBUG=true
```

3. **Run the server**
```bash
./agentos
```

## Docker Deployment

### Building Custom Docker Image

```bash
# Build the image
docker build -t agentos:latest .

# Run the container
docker run -d \
  -p 8080:8080 \
  -e OPENAI_API_KEY=sk-your-key-here \
  --name agentos \
  agentos:latest
```

### Multi-Stage Build Explanation

The Dockerfile uses multi-stage builds for optimal image size:

```dockerfile
# Stage 1: Build (golang:1.21-alpine)
# - Compiles Go application
# - Includes build tools

# Stage 2: Runtime (alpine:latest)
# - Minimal runtime environment (~5MB base)
# - Only includes compiled binary + CA certificates
# - Runs as non-root user (agno:1000)
```

**Image sizes:**
- Builder stage: ~300MB
- Final image: ~15MB

### Docker Compose Profiles

```bash
# Core services only (default)
docker-compose up -d

# With local Ollama
docker-compose --profile with-ollama up -d

# With ChromaDB vector database
docker-compose --profile with-vectordb up -d

# All services
docker-compose --profile with-ollama --profile with-vectordb up -d
```

### Service Architecture

```
┌─────────────────────────────────────────────┐
│            Load Balancer (Optional)          │
└──────────────────┬──────────────────────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
    ▼              ▼              ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│AgentOS  │  │AgentOS  │  │AgentOS  │
│Instance1│  │Instance2│  │Instance3│
└────┬────┘  └────┬────┘  └────┬────┘
     │            │            │
     └────────────┼────────────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌─────────┐  ┌────────┐  ┌──────────┐
│PostgreSQL│ │ Redis  │  │ Ollama   │
└─────────┘  └────────┘  └──────────┘
```

## Kubernetes Deployment

### Namespace Setup

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: agentos
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentos-config
  namespace: agentos
data:
  AGENTOS_ADDRESS: ":8080"
  AGENTOS_DEBUG: "false"
  POSTGRES_HOST: "postgres-service"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "agentos"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
```

### Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: agentos-secrets
  namespace: agentos
type: Opaque
stringData:
  OPENAI_API_KEY: "sk-your-openai-api-key"
  ANTHROPIC_API_KEY: "sk-ant-your-anthropic-key"
  POSTGRES_PASSWORD: "your-secure-password"
```

### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentos
  namespace: agentos
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agentos
  template:
    metadata:
      labels:
        app: agentos
    spec:
      containers:
      - name: agentos
        image: agentos:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: agentos-secrets
              key: OPENAI_API_KEY
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: agentos-secrets
              key: ANTHROPIC_API_KEY
        envFrom:
        - configMapRef:
            name: agentos-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: agentos-service
  namespace: agentos
spec:
  type: LoadBalancer
  selector:
    app: agentos
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

### Apply Kubernetes Resources

```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Verify deployment
kubectl get pods -n agentos
kubectl get svc -n agentos
```

## Production Deployment

### 1. Infrastructure Requirements

**Minimum Requirements:**
- CPU: 2 cores
- RAM: 2GB
- Storage: 10GB
- Network: 100Mbps

**Recommended for Production:**
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 50GB+ SSD
- Network: 1Gbps

### 2. Database Configuration

**PostgreSQL (Production Settings):**

```sql
-- Connection pooling
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
work_mem = 16MB

-- Logging
log_destination = 'stderr'
logging_collector = on
log_directory = 'pg_log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_rotation_age = 1d
log_rotation_size = 100MB

-- Performance
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
```

**Redis (Production Settings):**

```conf
# Memory
maxmemory 1gb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000

# Security
requirepass your-secure-password
```

### 3. Reverse Proxy (Nginx)

```nginx
upstream agentos {
    least_conn;
    server agentos1:8080 max_fails=3 fail_timeout=30s;
    server agentos2:8080 max_fails=3 fail_timeout=30s;
    server agentos3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL certificates
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # Timeouts
    proxy_connect_timeout 30s;
    proxy_send_timeout 30s;
    proxy_read_timeout 60s;

    # Buffer sizes
    client_max_body_size 10m;
    proxy_buffering on;
    proxy_buffer_size 4k;
    proxy_buffers 8 4k;

    location / {
        proxy_pass http://agentos;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check endpoint (no rate limiting)
    location /health {
        proxy_pass http://agentos;
        access_log off;
    }
}
```

### 4. Systemd Service (Native Deployment)

```ini
# /etc/systemd/system/agentos.service
[Unit]
Description=AgentOS API Server
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=agentos
Group=agentos
WorkingDirectory=/opt/agentos
EnvironmentFile=/opt/agentos/.env
ExecStart=/opt/agentos/agentos
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=agentos

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/agentos/logs

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

**Enable and start:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable agentos
sudo systemctl start agentos
sudo systemctl status agentos

# View logs
sudo journalctl -u agentos -f
```

## Environment Configuration

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AGENTOS_ADDRESS` | No | `:8080` | Server listen address |
| `AGENTOS_DEBUG` | No | `false` | Enable debug mode |
| `OPENAI_API_KEY` | Yes* | - | OpenAI API key |
| `ANTHROPIC_API_KEY` | No | - | Anthropic Claude API key |
| `OLLAMA_BASE_URL` | No | `http://localhost:11434` | Ollama server URL |
| `POSTGRES_HOST` | No | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | No | `5432` | PostgreSQL port |
| `POSTGRES_USER` | No | `agno` | PostgreSQL username |
| `POSTGRES_PASSWORD` | No | - | PostgreSQL password |
| `POSTGRES_DB` | No | `agentos` | PostgreSQL database |
| `REDIS_HOST` | No | `localhost` | Redis host |
| `REDIS_PORT` | No | `6379` | Redis port |
| `REDIS_PASSWORD` | No | - | Redis password |
| `REQUEST_TIMEOUT` | No | `30` | Request timeout (seconds) |
| `MAX_REQUEST_SIZE` | No | `10` | Max request size (MB) |
| `ALLOW_ORIGINS` | No | `*` | CORS allowed origins |
| `LOG_LEVEL` | No | `info` | Log level (debug/info/warn/error) |

*At least one LLM provider API key is required

### Configuration Files

**1. Create .env file**
```bash
cp .env.example .env
nano .env
```

**2. Secure the file**
```bash
chmod 600 .env
chown agentos:agentos .env
```

**3. Validate configuration**
```bash
# Check required variables
grep -E "^(OPENAI_API_KEY|ANTHROPIC_API_KEY)" .env

# Test connection
./agentos --validate-config
```

## Database Setup

### PostgreSQL Setup

**1. Install PostgreSQL**
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-15

# macOS
brew install postgresql@15

# Docker
docker run -d \
  --name postgres \
  -e POSTGRES_USER=agno \
  -e POSTGRES_PASSWORD=secure_password \
  -e POSTGRES_DB=agentos \
  -p 5432:5432 \
  postgres:15-alpine
```

**2. Initialize database**
```bash
# Using psql
psql -h localhost -U agno -d agentos -f scripts/init-db.sql

# Or using Docker
docker exec -i postgres psql -U agno -d agentos < scripts/init-db.sql
```

**3. Verify setup**
```bash
psql -h localhost -U agno -d agentos -c "\dt"
```

### Redis Setup

**1. Install Redis**
```bash
# Ubuntu/Debian
sudo apt-get install redis-server

# macOS
brew install redis

# Docker
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine
```

**2. Configure Redis**
```bash
# Edit redis.conf
sudo nano /etc/redis/redis.conf

# Set password
requirepass your-secure-password

# Enable persistence
save 900 1
save 300 10
```

**3. Restart Redis**
```bash
sudo systemctl restart redis
```

## Monitoring & Logging

### Health Checks

**Endpoint:** `GET /health`

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed monitoring script
#!/bin/bash
while true; do
    STATUS=$(curl -s http://localhost:8080/health | jq -r '.status')
    if [ "$STATUS" != "healthy" ]; then
        echo "$(date): AgentOS is unhealthy!"
        # Send alert (email, Slack, PagerDuty, etc.)
    fi
    sleep 30
done
```

### Prometheus Metrics (Future Enhancement)

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'agentos'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Structured Logging

AgentOS uses `log/slog` for structured logging:

```json
{
  "time": "2025-10-02T00:15:15Z",
  "level": "INFO",
  "msg": "request",
  "method": "POST",
  "path": "/api/v1/agents/assistant/run",
  "status": 200,
  "duration": "1.234s",
  "ip": "192.168.1.1"
}
```

**Configure log output:**
```bash
# JSON format (production)
export LOG_FORMAT=json

# Text format (development)
export LOG_FORMAT=text

# Log level
export LOG_LEVEL=info  # debug, info, warn, error
```

### Log Aggregation

**Using Loki + Promtail:**

```yaml
# promtail-config.yaml
server:
  http_listen_port: 9080

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: agentos
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'
```

## Security Best Practices

### 1. API Key Management

**Never commit API keys to version control:**
```bash
# Use .gitignore
echo ".env" >> .gitignore
echo "*.key" >> .gitignore
```

**Use secrets management:**
```bash
# AWS Secrets Manager
aws secretsmanager get-secret-value \
  --secret-id agentos/openai-api-key \
  --query SecretString \
  --output text

# HashiCorp Vault
vault kv get secret/agentos/api-keys
```

### 2. Network Security

**Firewall rules:**
```bash
# Allow only necessary ports
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw deny 8080/tcp  # Block direct access to AgentOS
sudo ufw enable
```

**Use private networking:**
```yaml
# docker-compose.yml
networks:
  internal:
    driver: bridge
    internal: true  # No external access
  external:
    driver: bridge
```

### 3. Authentication & Authorization

**Add authentication middleware:**
```go
// middleware/auth.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        // Validate token (JWT, API key, etc.)
        if !validateToken(token) {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 4. Rate Limiting

**Application-level:**
```go
import "github.com/ulule/limiter/v3"

// 10 requests per second per IP
rate := limiter.Rate{
    Period: 1 * time.Second,
    Limit:  10,
}
```

**Nginx-level:**
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req zone=api burst=20 nodelay;
```

### 5. HTTPS/TLS

**Let's Encrypt (Free SSL):**
```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d api.yourdomain.com

# Auto-renewal
sudo certbot renew --dry-run
```

## Troubleshooting

### Common Issues

**1. Server won't start**
```bash
# Check if port is already in use
lsof -i :8080
netstat -tuln | grep 8080

# Check logs
docker-compose logs agentos
journalctl -u agentos -n 50
```

**2. Database connection failed**
```bash
# Test PostgreSQL connection
psql -h localhost -U agno -d agentos -c "SELECT 1;"

# Check PostgreSQL logs
docker-compose logs postgres
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

**3. Agent execution timeout**
```bash
# Increase timeout in .env
REQUEST_TIMEOUT=60

# Or in code
server, _ := agentos.NewServer(&agentos.Config{
    RequestTimeout: 60 * time.Second,
})
```

**4. Memory issues**
```bash
# Check memory usage
docker stats agentos

# Increase memory limit
docker-compose.yml:
  services:
    agentos:
      deploy:
        resources:
          limits:
            memory: 1G
```

### Debug Mode

**Enable debug logging:**
```bash
export AGENTOS_DEBUG=true
export LOG_LEVEL=debug
```

**Debug output example:**
```
DEBUG model invoked provider=openai model=gpt-4 tokens=150
DEBUG tool called name=calculator function=add args={"a":5,"b":3}
DEBUG session retrieved session_id=abc-123 message_count=5
```

### Performance Tuning

**1. Database optimization**
```sql
-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM sessions WHERE agent_id = 'assistant';

-- Add missing indexes
CREATE INDEX idx_sessions_agent_user ON sessions(agent_id, user_id);

-- Vacuum and analyze
VACUUM ANALYZE sessions;
```

**2. Redis caching**
```go
// Cache agent responses
rdb.Set(ctx, cacheKey, response, 5*time.Minute)
```

**3. Connection pooling**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Backup & Recovery

### Database Backup

**Automated PostgreSQL backup:**
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/agentos_$TIMESTAMP.sql.gz"

pg_dump -h localhost -U agno agentos | gzip > $BACKUP_FILE

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE"
```

**Restore from backup:**
```bash
gunzip -c backup.sql.gz | psql -h localhost -U agno agentos
```

### Disaster Recovery

**1. Document recovery procedures**
**2. Test backups regularly**
**3. Maintain offsite backups**
**4. Implement monitoring and alerting**

## Scaling Strategies

### Horizontal Scaling

**Load balanced setup:**
```
       Internet
           |
     Load Balancer
           |
    ┌──────┼──────┐
    │      │      │
  App1   App2   App3
    │      │      │
    └──────┼──────┘
           |
    ┌──────┼──────┐
    │             │
Database         Redis
```

**Auto-scaling (Kubernetes):**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: agentos-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: agentos
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Support

- 📚 [Full Documentation](https://docs.agno.com)
- 💬 [GitHub Discussions](https://github.com/rexleimo/agno-go/discussions)
- 🐛 [Report Issues](https://github.com/rexleimo/agno-go/issues)

## License

MIT License - See [LICENSE](../LICENSE) for details.
