# AI Features

Dashforge is designed from the ground up to work with LLM agents. The non-polymorphic JSON format, semantic data layer, and validation system make it easy for AI to generate valid dashboard configurations.

## Why AI-First?

### Non-polymorphic JSON

Traditional dashboard formats use polymorphic schemas:

```json
// Polymorphic (hard for AI)
{
  "type": "chart",
  "config": { ... }  // Structure depends on chart type
}
```

Dashforge uses flat, explicit structures:

```json
// Non-polymorphic (easy for AI)
{
  "type": "chart",
  "config": {
    "geometry": "line",
    "encodings": { "x": "date", "y": "revenue" },
    "style": { "smooth": true }
  }
}
```

### Semantic Context

Cube.js provides business context that helps AI understand the data:

```
Raw SQL (confusing):
SELECT DATE_TRUNC('month', o.created_at), SUM(o.amount)
FROM orders o GROUP BY 1

Semantic (clear):
Orders.totalRevenue by Orders.createdAt (monthly)
```

### Validated Generation

All AI-generated configurations are validated before use:

```
User prompt → AI generates JSON → Validation → Dashboard/Widget
                                      ↓
                              Error feedback → AI retry
```

## AI Capabilities

### 1. Generate Dashboard from Description

Create a complete dashboard from a natural language description:

```
User: "Create a sales dashboard with revenue by region and monthly trends"

Generated:
- Title: "Sales Dashboard"
- Widgets:
  - Revenue by Region (bar chart)
  - Monthly Revenue Trend (line chart)
  - Total Revenue (metric)
  - Top Products (table)
```

### 2. Add Widget from Description

Add a single widget to an existing dashboard:

```
User: "Add a pie chart showing sales by category"

Generated widget:
{
  "type": "chart",
  "title": "Sales by Category",
  "position": { "x": 0, "y": 6, "w": 4, "h": 3 },
  "config": {
    "geometry": "pie",
    "encodings": { "category": "category", "value": "sales" }
  }
}
```

### 3. Modify Existing Widget

Update a widget based on instructions:

```
User: "Change the bar chart to show percentages"

Before:
{ "geometry": "bar", "encodings": { "y": "revenue" } }

After:
{ "geometry": "bar", "encodings": { "y": "revenue_pct" }, "style": { "format": "percent" } }
```

### 4. Natural Language Queries

Generate Cube.js queries from questions:

```
User: "Which products had the highest profit margin last quarter?"

Generated query:
{
  "measures": ["Products.profitMargin"],
  "dimensions": ["Products.name"],
  "timeDimensions": [{
    "dimension": "Orders.createdAt",
    "dateRange": "last quarter"
  }],
  "order": { "Products.profitMargin": "desc" },
  "limit": 10
}
```

## JSON Schemas

The AI uses JSON Schema definitions to ensure valid output.

### Dashboard Schema

```json
{
  "type": "object",
  "required": ["title", "widgets"],
  "properties": {
    "title": { "type": "string" },
    "layout": {
      "type": "object",
      "properties": {
        "type": { "enum": ["grid"] },
        "columns": { "type": "integer", "default": 12 }
      }
    },
    "widgets": {
      "type": "array",
      "items": { "$ref": "#/$defs/Widget" }
    }
  }
}
```

### Widget Schema

```json
{
  "type": "object",
  "required": ["type", "position"],
  "properties": {
    "type": { "enum": ["chart", "metric", "table", "text"] },
    "title": { "type": "string" },
    "position": {
      "type": "object",
      "properties": {
        "x": { "type": "integer", "minimum": 0 },
        "y": { "type": "integer", "minimum": 0 },
        "w": { "type": "integer", "minimum": 1, "maximum": 12 },
        "h": { "type": "integer", "minimum": 1 }
      }
    }
  }
}
```

### ChartConfig Schema

```json
{
  "type": "object",
  "required": ["geometry"],
  "properties": {
    "geometry": { "enum": ["line", "bar", "pie", "scatter", "area"] },
    "encodings": {
      "type": "object",
      "properties": {
        "x": { "type": "string" },
        "y": { "type": "string" },
        "color": { "type": "string" },
        "value": { "type": "string" },
        "category": { "type": "string" }
      }
    },
    "style": {
      "type": "object",
      "properties": {
        "showLegend": { "type": "boolean" },
        "smooth": { "type": "boolean" },
        "stack": { "type": "boolean" }
      }
    }
  }
}
```

## System Prompts

The builder includes optimized prompts for dashboard generation.

### Dashboard Generation Prompt

```
You are a dashboard design assistant. You help users create data dashboards
by generating JSON configurations.

You output JSON that follows the DashboardIR specification. Key rules:
- Use a 12-column grid layout
- Position widgets using x, y, w, h coordinates
- Common widget sizes: metrics (2x2), charts (4x3 or 6x3), tables (6x4)
- Align widgets to avoid overlap
- Use descriptive titles for widgets

Chart types: line, bar, pie, scatter, area
Widget types: chart, metric, table, text

Always respond with valid JSON only.
```

### Widget Generation Prompt

The prompt includes:

1. Schema definition
2. Existing widget positions (to avoid overlap)
3. Available data fields (from Cube.js schema)
4. User's request

## Validation

All AI-generated output is validated before use.

### Validation Steps

1. **JSON Parse**: Ensure valid JSON syntax
2. **Schema Validation**: Check against JSON Schema
3. **Semantic Validation**: Verify field references exist
4. **Layout Validation**: Check for widget overlaps

### Auto-fix

Common issues are automatically fixed:

- Missing IDs are generated
- Negative positions are set to 0
- Widths exceeding 12 are capped
- Missing required fields use defaults

### Error Feedback

If validation fails, errors are returned to the AI for retry:

```json
{
  "valid": false,
  "errors": [
    "Position x cannot be negative",
    "Chart geometry must be one of: line, bar, pie, scatter, area"
  ],
  "warnings": [
    "Widget has no title"
  ]
}
```

## API Integration

### AI Generation Endpoint

```
POST /api/v1/ai/generate
Content-Type: application/json

{
  "prompt": "Create a sales dashboard",
  "type": "dashboard",
  "schema": { ... },  // Optional: Cube.js schema for context
  "options": {
    "model": "claude-3-sonnet",
    "temperature": 0.7
  }
}
```

### Response

```json
{
  "success": true,
  "data": {
    "title": "Sales Dashboard",
    "widgets": [ ... ]
  },
  "warnings": []
}
```

## Mock Generation

For development without an AI backend, the builder includes mock generation:

```typescript
import { mockGenerateWidget } from './api/ai'

const result = await mockGenerateWidget(
  "Add a line chart showing trends",
  existingWidgets
)
// Returns a basic widget based on keyword matching
```

## Best Practices

### For Prompt Engineering

1. Be specific about data fields: "revenue by region" not "sales data"
2. Specify chart types when possible: "bar chart" not "visualization"
3. Include time ranges: "last 6 months" not "recent"
4. Mention comparisons: "vs previous period"

### For AI Integration

1. Always validate generated output
2. Provide schema context from Cube.js
3. Include existing widget positions to avoid overlap
4. Use temperature 0.7 for creativity, 0.3 for precision
5. Implement retry logic for validation failures

## Example Prompts

### Dashboard Generation

```
"Create an executive dashboard with:
- Revenue and profit KPIs at the top
- Monthly trend chart in the middle
- Top 10 products table on the right
- Revenue by region pie chart"
```

### Widget Addition

```
"Add a metric showing total customers with a sparkline
showing the trend over the last 30 days"
```

### Query Generation

```
"Show me the top 5 sales reps by revenue this quarter,
with their quota attainment percentage"
```
