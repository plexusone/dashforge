# Widgets

Widgets are the visual components that display data in your dashboard.

## Common Widget Properties

All widgets share these properties:

```json
{
  "id": "widget-id",
  "type": "chart",
  "title": "Widget Title",
  "description": "Optional description",
  "position": { "x": 0, "y": 0, "w": 6, "h": 4 },
  "dataSourceId": "data-source-id",
  "transform": [ ... ],
  "config": { ... }
}
```

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| id | string | Yes | Unique widget identifier |
| type | string | Yes | Widget type (chart, metric, table, text) |
| title | string | No | Display title |
| description | string | No | Tooltip description |
| position | object | Yes | Grid position and size |
| dataSourceId | string | No | Reference to data source |
| transform | array | No | Widget-level data transforms |
| config | object | Yes | Type-specific configuration |

### Position

Position uses a 12-column grid system:

```json
"position": {
  "x": 0,      // Column start (0-11)
  "y": 0,      // Row start
  "w": 6,      // Width in columns (1-12)
  "h": 4       // Height in rows
}
```

## Widget Types

### Chart

Displays data visualizations using ChartIR (ECharts-based).

```json
{
  "id": "sales-trend",
  "type": "chart",
  "title": "Sales Trend",
  "position": { "x": 0, "y": 0, "w": 8, "h": 4 },
  "dataSourceId": "monthly-sales",
  "config": {
    "marks": [
      {
        "geometry": "line",
        "encode": {
          "x": "month",
          "y": "revenue"
        },
        "style": {
          "stroke": "#4F46E5",
          "strokeWidth": 2
        }
      }
    ],
    "axes": [
      { "type": "category", "position": "bottom", "label": "Month" },
      { "type": "value", "position": "left", "label": "Revenue ($)" }
    ],
    "legend": {
      "show": true,
      "position": "top"
    }
  }
}
```

#### Chart Geometries

| Geometry | Description | Best For |
|----------|-------------|----------|
| line | Line chart | Trends over time |
| bar | Bar chart | Categorical comparisons |
| area | Area chart | Volume over time |
| pie | Pie/donut chart | Part-to-whole |
| scatter | Scatter plot | Correlations |

#### Bar Chart Example

```json
{
  "config": {
    "marks": [
      {
        "geometry": "bar",
        "encode": { "x": "category", "y": "value" },
        "style": { "fill": "#22C55E" }
      }
    ],
    "axes": [
      { "type": "category", "position": "bottom" },
      { "type": "value", "position": "left" }
    ]
  }
}
```

#### Multi-Series Line Chart

```json
{
  "config": {
    "marks": [
      {
        "geometry": "line",
        "encode": { "x": "date", "y": "sales", "color": "region" }
      }
    ],
    "axes": [
      { "type": "time", "position": "bottom" },
      { "type": "value", "position": "left" }
    ],
    "legend": { "show": true }
  }
}
```

#### Pie Chart Example

```json
{
  "config": {
    "marks": [
      {
        "geometry": "pie",
        "encode": { "angle": "value", "color": "category" },
        "style": { "innerRadius": 0.5 }
      }
    ],
    "legend": { "show": true, "position": "right" }
  }
}
```

### Metric

Displays a single value with optional formatting and thresholds.

```json
{
  "id": "total-revenue",
  "type": "metric",
  "title": "Total Revenue",
  "position": { "x": 0, "y": 0, "w": 3, "h": 2 },
  "dataSourceId": "summary",
  "config": {
    "valueField": "revenue",
    "format": "currency",
    "prefix": "$",
    "decimals": 0,
    "comparison": {
      "field": "previous_revenue",
      "type": "percentage"
    },
    "thresholds": [
      { "value": 0, "color": "#EF4444", "label": "Low" },
      { "value": 50000, "color": "#F59E0B", "label": "Medium" },
      { "value": 100000, "color": "#22C55E", "label": "High" }
    ]
  }
}
```

#### Format Options

| Format | Description | Example |
|--------|-------------|---------|
| number | Plain number | 1,234 |
| currency | Currency format | $1,234.00 |
| percent | Percentage | 85.5% |
| compact | Compact notation | 1.2K |
| duration | Time duration | 2h 30m |

### Table

Displays data in a tabular format.

```json
{
  "id": "top-products",
  "type": "table",
  "title": "Top Products",
  "position": { "x": 0, "y": 4, "w": 12, "h": 6 },
  "dataSourceId": "products",
  "config": {
    "columns": [
      { "field": "name", "header": "Product Name", "width": 200 },
      { "field": "category", "header": "Category" },
      { "field": "sales", "header": "Sales", "format": "currency", "align": "right" },
      {
        "field": "growth",
        "header": "Growth",
        "format": "percent",
        "conditionalStyle": {
          "positive": { "color": "#22C55E" },
          "negative": { "color": "#EF4444" }
        }
      },
      {
        "field": "status",
        "header": "Status",
        "type": "badge",
        "colorMap": {
          "active": "#22C55E",
          "pending": "#F59E0B",
          "inactive": "#6B7280"
        }
      }
    ],
    "sortable": true,
    "filterable": true,
    "pagination": {
      "enabled": true,
      "pageSize": 10
    }
  }
}
```

### Text

Displays static or dynamic text content.

```json
{
  "id": "header",
  "type": "text",
  "position": { "x": 0, "y": 0, "w": 12, "h": 1 },
  "config": {
    "content": "## Welcome to the Dashboard\n\nLast updated: {{lastUpdated}}",
    "markdown": true,
    "align": "center"
  }
}
```

## Widget-Level Transforms

Apply data transformations specific to a widget:

```json
{
  "id": "top-5-chart",
  "type": "chart",
  "dataSourceId": "all-products",
  "transform": [
    { "type": "sort", "config": { "field": "sales", "direction": "desc" } },
    { "type": "limit", "config": { "count": 5 } }
  ],
  "config": { ... }
}
```

See [Transforms](transforms.md) for available transform types.
