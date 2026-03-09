# Cube.js Integration

Dashforge integrates with [Cube.js](https://cube.dev/) to provide a semantic data layer. This enables business-friendly queries and provides rich context for AI-powered dashboard generation.

## Why Cube.js?

### For Users

- **Business terms**: Query "revenue" and "customers" instead of `SUM(o.amount)` and `COUNT(DISTINCT u.id)`
- **Pre-built relationships**: Joins are defined in the model, no need to remember table relationships
- **Consistent metrics**: Everyone uses the same definition of "revenue"

### For AI Agents

- **Semantic context**: LLMs understand business concepts better than raw SQL
- **Guardrails**: Access control prevents unauthorized data access
- **Validated queries**: Cube validates query structure before execution

## Setup

### 1. Install Cube.js

```bash
cd cube
npm install
```

### 2. Configure Database Connection

Create `.env` in the `cube/` directory:

```bash
# PostgreSQL
CUBEJS_DB_TYPE=postgres
CUBEJS_DB_HOST=localhost
CUBEJS_DB_PORT=5432
CUBEJS_DB_NAME=your_database
CUBEJS_DB_USER=your_user
CUBEJS_DB_PASS=your_password

# Or MySQL
# CUBEJS_DB_TYPE=mysql
# CUBEJS_DB_HOST=localhost
# ...

# API settings
CUBEJS_API_SECRET=your-api-secret
```

### 3. Start Cube.js

```bash
cd cube
npm run dev
# Cube.js API available at http://localhost:4000
```

## Data Modeling

Cube.js data models are defined in YAML files in `cube/model/cubes/`.

### Example Cube

```yaml
# cube/model/cubes/orders.yml
cubes:
  - name: Orders
    title: Orders
    description: Customer orders with revenue metrics

    sql: >
      SELECT * FROM orders

    joins:
      - name: Customers
        relationship: many_to_one
        sql: "{CUBE}.customer_id = {Customers}.id"

    measures:
      - name: count
        title: Total Orders
        type: count

      - name: totalRevenue
        title: Total Revenue
        type: sum
        sql: total_amount
        format: currency

      - name: averageOrderValue
        title: Average Order Value
        type: avg
        sql: total_amount
        format: currency

    dimensions:
      - name: id
        title: Order ID
        type: string
        sql: id
        primaryKey: true

      - name: status
        title: Status
        type: string
        sql: status

      - name: createdAt
        title: Created At
        type: time
        sql: created_at
```

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Cube** | A logical entity representing a table or view |
| **Measure** | Aggregated values (SUM, COUNT, AVG, etc.) |
| **Dimension** | Attributes to group or filter by |
| **Join** | Relationship to another cube |
| **Segment** | Pre-defined filter conditions |

## Query Builder

The dashboard builder includes a visual query builder for Cube.js.

### Using the Query Builder

1. Select a cube from the dropdown
2. Choose measures (numeric values to display)
3. Choose dimensions (categories to group by)
4. Add filters (optional)
5. Click "Run Query" to preview results

### Query Format

Cube.js queries use JSON:

```json
{
  "measures": ["Orders.totalRevenue", "Orders.count"],
  "dimensions": ["Orders.status"],
  "timeDimensions": [{
    "dimension": "Orders.createdAt",
    "granularity": "month",
    "dateRange": "last 6 months"
  }],
  "filters": [{
    "member": "Orders.status",
    "operator": "equals",
    "values": ["completed"]
  }],
  "order": {
    "Orders.totalRevenue": "desc"
  },
  "limit": 100
}
```

## Schema Browser

The schema browser shows all available cubes, measures, and dimensions with their descriptions.

```
Orders
├── Measures
│   ├── count (Total Orders) - count
│   ├── totalRevenue (Total Revenue) - sum, currency
│   └── averageOrderValue (Average Order Value) - avg, currency
├── Dimensions
│   ├── id (Order ID) - string 🔑
│   ├── status (Status) - string
│   └── createdAt (Created At) - time
└── Joins
    └── Customers (many_to_one)
```

## Dashboard Data Source

To use Cube.js data in a widget, create a data source with type `cube`:

```json
{
  "dataSources": [{
    "id": "monthly-revenue",
    "type": "cube",
    "query": {
      "measures": ["Orders.totalRevenue"],
      "timeDimensions": [{
        "dimension": "Orders.createdAt",
        "granularity": "month",
        "dateRange": "last 12 months"
      }]
    }
  }],
  "widgets": [{
    "id": "revenue-chart",
    "type": "chart",
    "datasourceId": "monthly-revenue",
    "config": {
      "geometry": "line",
      "encodings": {
        "x": "Orders.createdAt",
        "y": "Orders.totalRevenue"
      }
    }
  }]
}
```

## Pre-aggregations

Cube.js can pre-aggregate data for faster queries:

```yaml
cubes:
  - name: Orders
    # ... measures and dimensions ...

    preAggregations:
      - name: monthlyRevenue
        measures:
          - totalRevenue
          - count
        timeDimension: createdAt
        granularity: month
        refreshKey:
          every: 1 hour
```

## Multi-tenancy

Cube.js supports multi-tenant data isolation:

```javascript
// cube/cube.js
module.exports = {
  contextToAppId: ({ securityContext }) => {
    return securityContext?.tenant_id || 'default';
  },

  queryRewrite: (query, { securityContext }) => {
    if (securityContext?.tenant_id) {
      query.filters = query.filters || [];
      query.filters.push({
        member: 'TenantBase.tenantId',
        operator: 'equals',
        values: [securityContext.tenant_id]
      });
    }
    return query;
  }
};
```

## AI Integration

Cube.js provides semantic context for AI-powered queries:

```
User: "Show me revenue by region for Q4"

AI receives schema context:
- Orders.totalRevenue (Total Revenue, currency)
- Orders.region (Region, string)
- Orders.createdAt (time dimension)

AI generates:
{
  "measures": ["Orders.totalRevenue"],
  "dimensions": ["Orders.region"],
  "timeDimensions": [{
    "dimension": "Orders.createdAt",
    "dateRange": ["2024-10-01", "2024-12-31"]
  }]
}
```

See [AI Features](ai-features.md) for more details.

## Troubleshooting

### Connection Issues

```bash
# Check if Cube.js is running
curl http://localhost:4000/cubejs-api/v1/meta

# Check database connection
cd cube && npm run dev -- --debug
```

### Query Errors

- Ensure measure/dimension names match exactly (case-sensitive)
- Check that joins are defined for cross-cube queries
- Verify time dimension format for date ranges

### Performance

- Use pre-aggregations for frequently-accessed queries
- Add indexes to database columns used in dimensions
- Consider query caching in Cube.js config
