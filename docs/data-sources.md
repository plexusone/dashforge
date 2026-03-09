# Data Sources

Data sources define where dashboard data comes from. Dashforge supports multiple data source types for different deployment scenarios.

## Overview

Dashforge has two levels of data sources:

1. **Dashboard Data Sources** - Defined in the dashboard JSON, these specify how widgets get their data (inline, URL, database query, etc.)
2. **External Data Sources** - Server-side database connections (PostgreSQL, MySQL, etc.) managed via API that dashboard queries can reference

This page covers both. For the external data source API, see [API Reference](api-reference.md#data-sources).

## Common Properties

All data sources share these properties:

```json
{
  "id": "source-id",
  "type": "url",
  "refreshInterval": 60,
  "transform": [ ... ]
}
```

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| id | string | Yes | Unique identifier for referencing |
| type | string | Yes | Data source type |
| refreshInterval | number | No | Auto-refresh interval in seconds |
| transform | array | No | Data transformations to apply |

## Data Source Types

### Inline

Embed data directly in the dashboard JSON. Best for small, static datasets.

```json
{
  "id": "categories",
  "type": "inline",
  "data": [
    { "name": "Electronics", "value": 450 },
    { "name": "Clothing", "value": 320 },
    { "name": "Food", "value": 280 },
    { "name": "Books", "value": 150 }
  ]
}
```

### URL

Fetch data from a URL (JSON or CSV). Works in static mode.

```json
{
  "id": "sales-data",
  "type": "url",
  "url": "./data/sales.json",
  "format": "json",
  "refreshInterval": 300
}
```

#### With Authentication

```json
{
  "id": "api-data",
  "type": "url",
  "url": "https://api.example.com/data",
  "format": "json",
  "headers": {
    "Authorization": "Bearer ${API_TOKEN}"
  }
}
```

#### CSV Format

```json
{
  "id": "csv-data",
  "type": "url",
  "url": "./data/report.csv",
  "format": "csv",
  "csvOptions": {
    "delimiter": ",",
    "header": true
  }
}
```

### PostgreSQL

Query a PostgreSQL database. Requires server mode.

```json
{
  "id": "orders",
  "type": "postgres",
  "query": "SELECT date, SUM(amount) as total, COUNT(*) as orders FROM orders WHERE status = 'completed' GROUP BY date ORDER BY date"
}
```

#### With Parameters

Use dashboard variables in queries:

```json
{
  "id": "filtered-orders",
  "type": "postgres",
  "query": "SELECT * FROM orders WHERE region = '${region}' AND created_at >= '${dateRange.start}' AND created_at <= '${dateRange.end}' ORDER BY created_at DESC LIMIT ${limit}"
}
```

!!! warning "SQL Injection"
    Variables are parameterized by the server to prevent SQL injection. Never concatenate user input directly into queries.

#### Connection Reference

For multi-database setups:

```json
{
  "id": "analytics",
  "type": "postgres",
  "connectionId": "analytics-db",
  "query": "SELECT * FROM events"
}
```

### Saved Query

Reference a saved query by ID (server mode only).

```json
{
  "id": "monthly-report",
  "type": "savedQuery",
  "queryId": "q_abc123"
}
```

## Transforms

Apply transformations to data after fetching. See [Transforms](transforms.md) for full reference.

```json
{
  "id": "top-products",
  "type": "postgres",
  "query": "SELECT * FROM products",
  "transform": [
    { "type": "sort", "config": { "field": "sales", "direction": "desc" } },
    { "type": "limit", "config": { "count": 10 } }
  ]
}
```

## Caching

Data sources support caching to reduce database load:

```json
{
  "id": "expensive-query",
  "type": "postgres",
  "query": "SELECT ... complex aggregation ...",
  "cache": {
    "enabled": true,
    "ttl": 300,
    "key": "expensive-query-${region}"
  }
}
```

## Combining Data Sources

Widgets can reference multiple data sources:

```json
{
  "dataSources": [
    {
      "id": "current-sales",
      "type": "postgres",
      "query": "SELECT SUM(amount) as current FROM sales WHERE year = 2024"
    },
    {
      "id": "previous-sales",
      "type": "postgres",
      "query": "SELECT SUM(amount) as previous FROM sales WHERE year = 2023"
    }
  ],
  "widgets": [
    {
      "id": "yoy-comparison",
      "type": "metric",
      "dataSourceId": "current-sales",
      "config": {
        "valueField": "current",
        "comparison": {
          "dataSourceId": "previous-sales",
          "field": "previous",
          "type": "percentage"
        }
      }
    }
  ]
}
```

## Error Handling

Configure error behavior:

```json
{
  "id": "data",
  "type": "url",
  "url": "https://api.example.com/data",
  "onError": {
    "behavior": "fallback",
    "fallbackData": [],
    "message": "Unable to load data"
  }
}
```

| Behavior | Description |
|----------|-------------|
| fail | Show error message (default) |
| fallback | Use fallback data |
| retry | Retry with backoff |
| ignore | Show empty state |

## External Data Sources (Server Mode)

In server mode, Dashforge supports connecting to external databases through a plugin-style provider system. This allows dashboards to query multiple databases like Metabase.

### Supported Providers

| Provider | Status | Description |
|----------|--------|-------------|
| postgres | Stable | PostgreSQL 12+ |
| mysql | Stable | MySQL 5.7+ / MariaDB 10.3+ |
| clickhouse | Planned | ClickHouse analytics database |
| duckdb | Planned | DuckDB embedded analytics |

### Creating a Data Source

Use the API to register external database connections:

```bash
curl -X POST http://localhost:8080/api/v1/datasources \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Analytics",
    "slug": "prod-analytics",
    "type": "postgres",
    "connectionUrl": "postgres://user:pass@db.example.com:5432/analytics?sslmode=require",
    "maxConnections": 10,
    "queryTimeoutSeconds": 30,
    "readOnly": true,
    "sslEnabled": true,
    "active": true
  }'
```

#### Using Environment Variables

For security, connection URLs can reference environment variables:

```json
{
  "name": "Production DB",
  "slug": "prod-db",
  "type": "postgres",
  "connectionUrlEnv": "PROD_DATABASE_URL",
  "readOnly": true
}
```

The server reads the URL from `PROD_DATABASE_URL` at connection time.

### Testing Connections

Before using a data source, test the connection:

```bash
curl -X POST http://localhost:8080/api/v1/datasources/1/test
```

Response:

```json
{
  "success": true,
  "durationMs": 45
}
```

### Executing Queries

Query an external data source:

```bash
curl -X POST http://localhost:8080/api/v1/datasources/1/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT date, SUM(amount) as total FROM sales GROUP BY date ORDER BY date",
    "parameters": {},
    "maxRows": 1000
  }'
```

Response:

```json
{
  "columns": [
    {"name": "date", "type": "DATE", "nullable": false},
    {"name": "total", "type": "NUMERIC", "nullable": true}
  ],
  "rows": [
    {"date": "2024-01-01", "total": 15000},
    {"date": "2024-01-02", "total": 18500}
  ],
  "rowCount": 2,
  "executionTimeMs": 125
}
```

### Schema Introspection

Get the database schema for a data source:

```bash
curl "http://localhost:8080/api/v1/datasources/1/schema?columns=true"
```

Response:

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
        {"name": "amount", "type": "NUMERIC", "nullable": true}
      ]
    }
  ]
}
```

### Read-Only Mode

When `readOnly: true`, the data source only allows SELECT queries:

```json
{
  "readOnly": true
}
```

Attempting INSERT/UPDATE/DELETE returns an error:

```json
{
  "error": "write operation not allowed on read-only connection"
}
```

!!! tip "Best Practice"
    Always set `readOnly: true` for dashboard data sources. Use a database user with read-only permissions as an additional safeguard.

### Connection Pooling

Each data source maintains its own connection pool:

| Setting | Default | Description |
|---------|---------|-------------|
| maxConnections | 10 | Maximum open connections |
| queryTimeoutSeconds | 30 | Query execution timeout |

Connection pools are managed automatically. Connections are reused across queries and closed when the data source is deleted or the server shuts down.

### Provider Capabilities

Query available providers and their capabilities:

```bash
curl http://localhost:8080/api/v1/datasources/providers
```

Response:

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

### Parameter Styles

Different databases use different parameter placeholders:

| Provider | Style | Example |
|----------|-------|---------|
| postgres | Positional dollar | `$1`, `$2`, `$3` |
| mysql | Positional question | `?`, `?`, `?` |

When using named parameters (`:name` or `@name`), Dashforge automatically converts them to the provider's native style.

### Custom Providers

Developers can register custom providers programmatically:

```go
import "github.com/grokify/dashforge/datasource"

// Implement the Provider interface
type MyProvider struct{}

func (p *MyProvider) Name() string { return "mydb" }
func (p *MyProvider) Connect(ctx context.Context, cfg datasource.ConnectionConfig) (datasource.Connection, error) { ... }
func (p *MyProvider) ValidateConfig(cfg datasource.ConnectionConfig) error { ... }
func (p *MyProvider) Capabilities() datasource.Capabilities { ... }

// Register at startup
datasource.Register(&MyProvider{})

// Or inject into manager for runtime registration
manager.RegisterCustomProvider(&MyProvider{})
```

See the [Development](development.md) guide for more details on extending Dashforge.
