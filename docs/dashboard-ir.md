# Dashboard IR Reference

The Dashboard Intermediate Representation (IR) is the JSON format that defines dashboards in Dashforge.

## Design Principles

1. **Non-polymorphic** - Each field has a single type, making it easy for code generation
2. **AI-friendly** - Structured for LLM generation and manipulation
3. **Progressive** - Simple dashboards require minimal configuration
4. **Composable** - Data sources, transforms, and widgets can be combined flexibly

## Top-Level Structure

```json
{
  "id": "dashboard-id",
  "title": "Dashboard Title",
  "description": "Optional description",
  "version": "1.0.0",
  "layout": { ... },
  "variables": [ ... ],
  "dataSources": [ ... ],
  "widgets": [ ... ],
  "theme": { ... }
}
```

## Fields

### id (required)

Unique identifier for the dashboard. Used in URLs and API calls.

```json
"id": "sales-q4-2024"
```

### title (required)

Display title shown in the dashboard header.

```json
"title": "Q4 2024 Sales Dashboard"
```

### description (optional)

Longer description for documentation.

```json
"description": "Tracks sales metrics and team performance for Q4 2024"
```

### version (optional)

Semantic version for tracking dashboard changes.

```json
"version": "2.1.0"
```

### layout (required)

Defines how widgets are arranged.

```json
"layout": {
  "type": "grid",
  "columns": 12,
  "rowHeight": 80,
  "gap": 16
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| type | string | "grid" | Layout type (currently only "grid") |
| columns | number | 12 | Number of grid columns |
| rowHeight | number | 80 | Height of each grid row in pixels |
| gap | number | 16 | Gap between widgets in pixels |

### variables (optional)

Dashboard-level variables for filtering and parameterization.

```json
"variables": [
  {
    "id": "date_range",
    "type": "dateRange",
    "label": "Date Range",
    "default": { "start": "2024-01-01", "end": "2024-12-31" }
  },
  {
    "id": "region",
    "type": "select",
    "label": "Region",
    "options": ["North", "South", "East", "West"],
    "default": "North"
  }
]
```

See [Variables](#variables-1) for full reference.

### dataSources (required)

Array of data source definitions. See [Data Sources](data-sources.md).

### widgets (required)

Array of widget definitions. See [Widgets](widgets.md).

### theme (optional)

Custom theme overrides.

```json
"theme": {
  "colors": {
    "primary": "#4F46E5",
    "success": "#22C55E",
    "warning": "#F59E0B",
    "error": "#EF4444"
  },
  "fontFamily": "Inter, sans-serif"
}
```

## Variables

Variables allow users to filter and parameterize dashboard data.

### Variable Types

#### text

Free-form text input.

```json
{
  "id": "search",
  "type": "text",
  "label": "Search",
  "placeholder": "Enter search term...",
  "default": ""
}
```

#### select

Single selection from options.

```json
{
  "id": "status",
  "type": "select",
  "label": "Status",
  "options": ["active", "pending", "closed"],
  "default": "active"
}
```

#### multiSelect

Multiple selection from options.

```json
{
  "id": "categories",
  "type": "multiSelect",
  "label": "Categories",
  "options": ["Electronics", "Clothing", "Food"],
  "default": ["Electronics"]
}
```

#### dateRange

Date range picker.

```json
{
  "id": "period",
  "type": "dateRange",
  "label": "Time Period",
  "default": {
    "start": "2024-01-01",
    "end": "2024-12-31"
  }
}
```

#### number

Numeric input with optional min/max.

```json
{
  "id": "limit",
  "type": "number",
  "label": "Row Limit",
  "min": 1,
  "max": 1000,
  "default": 100
}
```

### Using Variables

Reference variables in queries and transforms using `${variableName}` syntax:

```json
{
  "id": "filtered-data",
  "type": "postgres",
  "query": "SELECT * FROM orders WHERE status = '${status}' AND created_at BETWEEN '${period.start}' AND '${period.end}'"
}
```

## Complete Example

```json
{
  "id": "sales-dashboard",
  "title": "Sales Dashboard",
  "description": "Real-time sales metrics and trends",
  "version": "1.0.0",
  "layout": {
    "type": "grid",
    "columns": 12,
    "rowHeight": 80
  },
  "variables": [
    {
      "id": "region",
      "type": "select",
      "label": "Region",
      "options": ["All", "North", "South", "East", "West"],
      "default": "All"
    }
  ],
  "dataSources": [
    {
      "id": "sales-summary",
      "type": "postgres",
      "query": "SELECT SUM(amount) as total, COUNT(*) as orders FROM sales WHERE region = '${region}' OR '${region}' = 'All'"
    }
  ],
  "widgets": [
    {
      "id": "total-sales",
      "type": "metric",
      "title": "Total Sales",
      "position": { "x": 0, "y": 0, "w": 4, "h": 2 },
      "dataSourceId": "sales-summary",
      "config": {
        "valueField": "total",
        "format": "currency"
      }
    }
  ]
}
```
