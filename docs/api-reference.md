# API Reference

The Dashforge REST API provides programmatic access to dashboards, data sources, and queries.

## Base URL

```
https://your-domain.com/api/v1
```

## Authentication

Most endpoints require authentication via JWT Bearer token:

```bash
curl -H "Authorization: Bearer <access_token>" \
  https://dashforge.example.com/api/v1/dashboards
```

## Response Format

All responses use JSON:

```json
{
  "data": { ... },
  "meta": {
    "total": 100,
    "page": 1,
    "perPage": 20
  }
}
```

Error responses:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Dashboard not found"
  }
}
```

## Endpoints

### Health

#### Check Health

```
GET /health
```

No authentication required.

**Response:**

```json
{
  "status": "ok"
}
```

---

### Authentication

See [Authentication](authentication.md) for detailed OAuth flow documentation.

#### Get Current User

```
GET /api/v1/auth/me
```

**Response:**

```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "role": "admin",
  "active": true,
  "lastLoginAt": "2024-01-15T10:30:00Z",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

#### Refresh Token

```
POST /api/v1/auth/refresh
```

**Request:**

```json
{
  "refreshToken": "your-refresh-token"
}
```

**Response:**

```json
{
  "accessToken": "new-access-token",
  "refreshToken": "new-refresh-token",
  "expiresIn": 900,
  "tokenType": "Bearer"
}
```

---

### Dashboards

#### List Dashboards

```
GET /api/v1/dashboards
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Page number (default: 1) |
| perPage | int | Items per page (default: 20, max: 100) |
| search | string | Search title/description |
| sort | string | Sort field (created_at, updated_at, title) |
| order | string | Sort order (asc, desc) |

**Response:**

```json
{
  "data": [
    {
      "id": 1,
      "slug": "sales-dashboard",
      "title": "Sales Dashboard",
      "description": "Monthly sales metrics",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "perPage": 20
  }
}
```

#### Get Dashboard

```
GET /api/v1/dashboards/:id
```

**Response:**

```json
{
  "id": 1,
  "slug": "sales-dashboard",
  "title": "Sales Dashboard",
  "description": "Monthly sales metrics",
  "definition": {
    "layout": { ... },
    "dataSources": [ ... ],
    "widgets": [ ... ]
  },
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

#### Create Dashboard

```
POST /api/v1/dashboards
```

**Request:**

```json
{
  "slug": "new-dashboard",
  "title": "New Dashboard",
  "description": "Optional description",
  "definition": {
    "layout": {
      "type": "grid",
      "columns": 12,
      "rowHeight": 80
    },
    "dataSources": [],
    "widgets": []
  }
}
```

**Response:** `201 Created`

```json
{
  "id": 2,
  "slug": "new-dashboard",
  "title": "New Dashboard",
  "createdAt": "2024-01-15T12:00:00Z",
  "updatedAt": "2024-01-15T12:00:00Z"
}
```

#### Update Dashboard

```
PUT /api/v1/dashboards/:id
```

**Request:**

```json
{
  "title": "Updated Title",
  "definition": { ... }
}
```

**Response:** `200 OK`

#### Delete Dashboard

```
DELETE /api/v1/dashboards/:id
```

**Response:** `204 No Content`

---

### Data Sources

External database connections for querying. See [Data Sources](data-sources.md#external-data-sources-server-mode) for concepts.

#### List Providers

```
GET /api/v1/datasources/providers
```

No authentication required.

**Response:**

```json
{
  "providers": [
    {
      "name": "postgres",
      "capabilities": {
        "supportsTransactions": true,
        "supportsStreaming": true,
        "supportsPreparedStmts": true,
        "supportsNamedParams": false,
        "parameterStyle": "positional_dollar"
      }
    },
    {
      "name": "mysql",
      "capabilities": {
        "supportsTransactions": true,
        "supportsStreaming": true,
        "supportsPreparedStmts": true,
        "supportsNamedParams": false,
        "parameterStyle": "positional_question"
      }
    }
  ]
}
```

#### List Data Sources

```
GET /api/v1/datasources
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| active | bool | Filter by active status |
| type | string | Filter by provider type (postgres, mysql) |

**Response:**

```json
{
  "dataSources": [
    {
      "id": 1,
      "name": "Production DB",
      "slug": "prod-db",
      "type": "postgres",
      "maxConnections": 10,
      "queryTimeoutSeconds": 30,
      "readOnly": true,
      "sslEnabled": true,
      "active": true,
      "lastConnectedAt": "2024-01-15T10:30:00Z",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

!!! note "Secrets Excluded"
    Connection URLs are never returned in list/get responses for security.

#### Get Data Source

```
GET /api/v1/datasources/:id
```

**Response:**

```json
{
  "id": 1,
  "name": "Production DB",
  "slug": "prod-db",
  "type": "postgres",
  "maxConnections": 10,
  "queryTimeoutSeconds": 30,
  "readOnly": true,
  "sslEnabled": true,
  "active": true,
  "connectionUrlEnv": "PROD_DATABASE_URL",
  "lastConnectedAt": "2024-01-15T10:30:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

#### Create Data Source

```
POST /api/v1/datasources
```

**Request:**

```json
{
  "name": "Analytics DB",
  "slug": "analytics-db",
  "type": "postgres",
  "connectionUrl": "postgres://user:pass@db.example.com:5432/analytics?sslmode=require",
  "maxConnections": 10,
  "queryTimeoutSeconds": 30,
  "readOnly": true,
  "sslEnabled": true,
  "active": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Display name |
| slug | string | Yes | URL-safe identifier (unique) |
| type | string | Yes | Provider type (postgres, mysql) |
| connectionUrl | string | * | Connection string |
| connectionUrlEnv | string | * | Env var containing connection URL |
| maxConnections | int | No | Max pool size (default: 10) |
| queryTimeoutSeconds | int | No | Query timeout (default: 30) |
| readOnly | bool | No | Restrict to SELECT queries |
| sslEnabled | bool | No | Enable SSL/TLS |
| active | bool | No | Enable data source |

*Either `connectionUrl` or `connectionUrlEnv` is required.

**Response:** `201 Created`

```json
{
  "id": 2,
  "name": "Analytics DB",
  "slug": "analytics-db",
  "type": "postgres",
  "active": true,
  "createdAt": "2024-01-15T12:00:00Z"
}
```

#### Update Data Source

```
PUT /api/v1/datasources/:id
```

**Request:**

```json
{
  "name": "Updated Name",
  "connectionUrl": "postgres://newurl...",
  "maxConnections": 20,
  "active": false
}
```

All fields are optional. Only provided fields are updated.

**Response:** `200 OK`

#### Delete Data Source

```
DELETE /api/v1/datasources/:id
```

Closes any active connections before deleting.

**Response:** `204 No Content`

#### Test Connection

```
POST /api/v1/datasources/:id/test
```

Tests the connection without caching. Updates `lastConnectedAt` on success.

**Response:**

```json
{
  "success": true,
  "durationMs": 45
}
```

On failure:

```json
{
  "success": false,
  "error": "connection refused",
  "durationMs": 5023
}
```

#### Execute Query

```
POST /api/v1/datasources/:id/query
```

**Request:**

```json
{
  "query": "SELECT date, SUM(amount) as total FROM sales WHERE region = :region GROUP BY date",
  "parameters": {
    "region": "US"
  },
  "maxRows": 1000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| query | string | Yes | SQL query |
| parameters | object | No | Named parameters (`:name` or `@name`) |
| maxRows | int | No | Max rows to return (default: 1000, max: 10000) |

**Response:**

```json
{
  "columns": [
    {"name": "date", "type": "DATE", "nullable": false},
    {"name": "total", "type": "NUMERIC", "nullable": true, "precision": 10, "scale": 2}
  ],
  "rows": [
    {"date": "2024-01-01", "total": 15000.50},
    {"date": "2024-01-02", "total": 18500.75}
  ],
  "rowCount": 2,
  "executionTimeMs": 125
}
```

!!! warning "Read-Only Data Sources"
    If the data source has `readOnly: true`, only SELECT queries are allowed. INSERT/UPDATE/DELETE will return a 400 error.

#### Get Schema

```
GET /api/v1/datasources/:id/schema
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| schema | string | Filter by schema name (default: public/current db) |
| columns | bool | Include column details |
| filter | string | Filter tables by name pattern |

**Response:**

```json
{
  "tables": [
    {
      "schema": "public",
      "name": "sales",
      "type": "table",
      "columns": [
        {"name": "id", "type": "INTEGER", "nullable": false},
        {"name": "date", "type": "DATE", "nullable": false},
        {"name": "amount", "type": "NUMERIC", "nullable": true, "precision": 10, "scale": 2},
        {"name": "region", "type": "VARCHAR", "nullable": true, "length": 50}
      ]
    },
    {
      "schema": "public",
      "name": "sales_summary",
      "type": "view"
    }
  ]
}
```

---

### Queries

#### Execute Query

```
POST /api/v1/query
```

**Request:**

```json
{
  "dataSourceId": 1,
  "query": "SELECT date, SUM(amount) as total FROM sales GROUP BY date",
  "params": {
    "limit": 100
  }
}
```

**Response:**

```json
{
  "columns": ["date", "total"],
  "rows": [
    { "date": "2024-01-01", "total": 15000 },
    { "date": "2024-01-02", "total": 18500 }
  ],
  "rowCount": 2,
  "executionTimeMs": 125
}
```

#### List Saved Queries

```
GET /api/v1/queries
```

#### Save Query

```
POST /api/v1/queries
```

**Request:**

```json
{
  "name": "Monthly Revenue",
  "description": "Total revenue by month",
  "dataSourceId": 1,
  "query": "SELECT DATE_TRUNC('month', created_at) as month, SUM(amount) as revenue FROM orders GROUP BY 1"
}
```

#### Run Saved Query

```
POST /api/v1/queries/:id/run
```

**Request:**

```json
{
  "params": {
    "year": 2024
  }
}
```

---

### Users (Admin)

Requires `admin` or `owner` role.

#### List Users

```
GET /api/v1/admin/users
```

#### Create User

```
POST /api/v1/admin/users
```

**Request:**

```json
{
  "email": "newuser@example.com",
  "name": "New User",
  "role": "editor"
}
```

#### Update User Role

```
PATCH /api/v1/admin/users/:id
```

**Request:**

```json
{
  "role": "admin"
}
```

#### Deactivate User

```
DELETE /api/v1/admin/users/:id
```

---

### Tenants (Admin)

Requires `owner` role.

#### List Tenants

```
GET /api/v1/admin/tenants
```

#### Create Tenant

```
POST /api/v1/admin/tenants
```

**Request:**

```json
{
  "slug": "new-tenant",
  "name": "New Tenant Inc",
  "plan": "pro"
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| VALIDATION_ERROR | 400 | Invalid request data |
| CONFLICT | 409 | Resource already exists |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |

## Rate Limiting

API requests are rate limited:

| Tier | Requests/minute |
|------|-----------------|
| Free | 60 |
| Pro | 300 |
| Enterprise | 1000 |

Rate limit headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1705311360
```

## Pagination

List endpoints support pagination:

```
GET /api/v1/dashboards?page=2&perPage=50
```

Response includes meta:

```json
{
  "meta": {
    "total": 150,
    "page": 2,
    "perPage": 50,
    "totalPages": 3
  }
}
```
