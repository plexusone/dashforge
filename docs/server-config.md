# Server Configuration

The Dashforge server provides a full-featured backend for dashboard hosting, data queries, and user management.

## Starting the Server

```bash
./dashforge-server serve [flags]
```

## Configuration Methods

Configuration can be provided via:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**

### Command-Line Flags

```bash
./dashforge-server serve \
  --port 8080 \
  --database-url "postgres://user:pass@localhost:5432/dashforge" \
  --jwt-secret "your-secret-key" \
  --auto-migrate \
  --enable-rls
```

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--port` | `PORT` | 8080 | HTTP server port |
| `--database-url` | `DATABASE_URL` | - | PostgreSQL connection string |
| `--dashboard-dir` | `DASHBOARD_DIR` | - | Directory for static dashboard files |
| `--jwt-secret` | `JWT_SECRET` | - | Secret for signing JWT tokens |
| `--base-url` | `BASE_URL` | - | Public URL (for OAuth callbacks) |
| `--auto-migrate` | `AUTO_MIGRATE` | false | Run database migrations on startup |
| `--enable-rls` | `ENABLE_RLS` | false | Enable Row Level Security |
| `--disable-auth` | `DISABLE_AUTH` | false | Disable authentication (dev only) |
| `--config` | - | - | Path to YAML config file |

### Environment Variables

```bash
export PORT=8080
export DATABASE_URL="postgres://user:pass@localhost:5432/dashforge?sslmode=require"
export JWT_SECRET="your-secret-key-at-least-32-characters"
export BASE_URL="https://dashforge.example.com"

# OAuth providers
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"

# Run server
./dashforge-server serve --auto-migrate
```

### Configuration File

Create `config.yaml`:

```yaml
port: 8080
database_url: postgres://user:pass@localhost:5432/dashforge
jwt_secret: your-secret-key
base_url: https://dashforge.example.com
auto_migrate: true
enable_rls: true

oauth:
  github:
    client_id: your-github-client-id
    client_secret: your-github-client-secret
  google:
    client_id: your-google-client-id
    client_secret: your-google-client-secret
```

Run with config file:

```bash
./dashforge-server serve --config config.yaml
```

## Database Configuration

### PostgreSQL Connection String

```
postgres://user:password@host:port/database?sslmode=mode
```

| Parameter | Description |
|-----------|-------------|
| user | Database username |
| password | Database password |
| host | Database host |
| port | Database port (default: 5432) |
| database | Database name |
| sslmode | SSL mode: disable, require, verify-ca, verify-full |

### Connection Pool Settings

The server uses sensible defaults:

- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

### Database Migrations

With `--auto-migrate`, the server automatically:

1. Creates tables using Ent schema
2. Applies RLS policies (if `--enable-rls`)

For production, consider running migrations separately:

```bash
# Run migrations only
./dashforge-server migrate --database-url "$DATABASE_URL"
```

## JWT Configuration

### Secret Requirements

- Minimum 32 characters recommended
- Use cryptographically random string
- Keep secret secure (environment variable or secrets manager)

Generate a secure secret:

```bash
openssl rand -base64 32
```

### Token Lifetimes

Default token lifetimes:

| Token | Lifetime |
|-------|----------|
| Access Token | 15 minutes |
| Refresh Token | 7 days |

## Production Deployment

### Docker

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dashforge-server ./cmd/dashforge-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/dashforge-server .
EXPOSE 8080
CMD ["./dashforge-server", "serve"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  dashforge:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://dashforge:password@db:5432/dashforge?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
      - AUTO_MIGRATE=true
    depends_on:
      - db

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=dashforge
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=dashforge
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### Health Check

The server exposes a health endpoint:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Use this for load balancer health checks and container orchestration.

### Reverse Proxy (nginx)

```nginx
upstream dashforge {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name dashforge.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://dashforge;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Logging

The server uses structured logging (slog). Log output includes:

- Request method and path
- Response status and duration
- Database query timing
- Authentication events

Set log level via environment:

```bash
export LOG_LEVEL=debug  # debug, info, warn, error
```
